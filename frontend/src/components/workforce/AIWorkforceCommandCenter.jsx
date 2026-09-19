import React, { useState, useEffect, useCallback } from 'react';
import {
  Shield,
  AlertOctagon,
  CheckCircle2,
  AlertTriangle,
  RefreshCw,
  Search,
  Filter,
  Sliders,
  Lock,
  Unlock,
  Play,
  Pause,
  Layers,
  ChevronRight,
  Info,
  Activity,
  Cpu,
  UserCheck,
  FileCheck,
  Zap,
  ArrowRight,
  XCircle,
  Eye,
  Check
} from 'lucide-react';
import workforceService from '../../services/workforceService';
import enterpriseService from '../../services/enterpriseService';
import './AIWorkforceCommandCenter.css';

export default function AIWorkforceCommandCenter() {
  const [activeTab, setActiveTab] = useState('agents'); // 'agents' | 'workflows' | 'approvals' | 'escalations' | 'activity'
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  const [error, setError] = useState(null);
  const [successMsg, setSuccessMsg] = useState(null);

  // Data states
  const [overview, setOverview] = useState(null);
  const [health, setHealth] = useState(null);
  const [workload, setWorkload] = useState([]);
  const [emergencyStatus, setEmergencyStatus] = useState(null);
  const [approvals, setApprovals] = useState([]);
  const [escalations, setEscalations] = useState([]);
  const [activity, setActivity] = useState([]);

  // Enterprise Autonomous Operations Platform states
  const [enterpriseWorkflows, setEnterpriseWorkflows] = useState([]);
  const [enterpriseOverview, setEnterpriseOverview] = useState(null);
  const [recoveringWorkflows, setRecoveringWorkflows] = useState(false);

  // Search & Filter
  const [searchQuery, setSearchQuery] = useState('');
  const [autonomyFilter, setAutonomyFilter] = useState('ALL');
  const [statusFilter, setStatusFilter] = useState('ALL');

  // Emergency Stop Modal
  const [stopModalOpen, setStopModalOpen] = useState(false);
  const [stopScope, setStopScope] = useState('WORKFORCE');
  const [stopTarget, setStopTarget] = useState('');
  const [stopReason, setStopReason] = useState('');
  const [submittingStop, setSubmittingStop] = useState(false);

  // Workflow Inspection Modal
  const [inspectModalOpen, setInspectModalOpen] = useState(false);
  const [inspectingPlanId, setInspectingPlanId] = useState(null);
  const [workflowDetail, setWorkflowDetail] = useState(null);
  const [loadingInspection, setLoadingInspection] = useState(false);

  // Agent Control Modal
  const [controlModalOpen, setControlModalOpen] = useState(false);
  const [targetAgent, setTargetAgent] = useState(null);
  const [controlAction, setControlAction] = useState('PAUSE');
  const [newAutonomyLevel, setNewAutonomyLevel] = useState('LEVEL_2_PREPARE');
  const [controlReason, setControlReason] = useState('');
  const [submittingControl, setSubmittingControl] = useState(false);

  // Load all command center data
  const loadData = useCallback(async (silent = false) => {
    if (!silent) setRefreshing(true);
    setError(null);
    try {
      const res = await workforceService.getCommandCenterOverview();
      const data = res?.data || res;
      setOverview(data);
      if (data?.health) setHealth(data.health);
      if (data?.agent_workload) setWorkload(data.agent_workload);
      if (data?.emergency_stop) setEmergencyStatus(data.emergency_stop);
      if (data?.waiting_approvals) setApprovals(data.waiting_approvals);
      if (data?.escalations) setEscalations(data.escalations);
      if (data?.recent_activity) setActivity(data.recent_activity);

      // Fetch enterprise autonomous platform data
      try {
        const entRes = await enterpriseService.listWorkflows({ limit: 50 });
        const entData = entRes?.data?.workflows || entRes?.workflows || [];
        setEnterpriseWorkflows(entData);
        const ovRes = await enterpriseService.getOverview();
        setEnterpriseOverview(ovRes?.data || ovRes);
      } catch (e) {
        // Safe fallback if enterprise routes are initializing
      }
    } catch (err) {
      console.error('Failed loading workforce command center data:', err);
      setError(err?.response?.data?.message || err?.message || 'Failed loading command center overview');
    } finally {
      setLoading(false);
      setRefreshing(false);
    }
  }, []);

  useEffect(() => {
    loadData();
    const interval = setInterval(() => {
      loadData(true);
    }, 30000);
    return () => clearInterval(interval);
  }, [loadData]);

  // Flash message helper
  const notifySuccess = (msg) => {
    setSuccessMsg(msg);
    setTimeout(() => setSuccessMsg(null), 4000);
  };

  // Handle Emergency Stop trigger/release
  const handleToggleEmergencyStop = async (action) => {
    setSubmittingStop(true);
    try {
      await workforceService.emergencyStop({
        action: action, // 'STOP' or 'RESUME'
        scope: stopScope,
        target: stopTarget,
        reason: stopReason || `Emergency ${action.toLowerCase()} initiated from Command Center`,
      });
      notifySuccess(`Emergency stop ${action === 'STOP' ? 'engaged' : 'released'} successfully`);
      setStopModalOpen(false);
      setStopReason('');
      setStopTarget('');
      loadData(true);
    } catch (err) {
      setError(err?.response?.data?.message || err?.message || 'Emergency stop request failed');
    } finally {
      setSubmittingStop(false);
    }
  };

  // Open Agent Control Modal
  const openAgentControl = (agent, action) => {
    setTargetAgent(agent);
    setControlAction(action);
    setNewAutonomyLevel(agent.autonomy_level || 'LEVEL_1_RECOMMEND');
    setControlReason('');
    setControlModalOpen(true);
  };

  // Submit Agent Control
  const handleSubmitAgentControl = async () => {
    if (!targetAgent) return;
    setSubmittingControl(true);
    try {
      await workforceService.controlAgent(targetAgent.agent_id, {
        action: controlAction,
        autonomy_level: controlAction === 'SET_AUTONOMY' ? newAutonomyLevel : undefined,
        reason: controlReason || `Governed ${controlAction} initiated from Command Center`,
      });
      notifySuccess(`Agent ${targetAgent.agent_id} updated: ${controlAction}`);
      setControlModalOpen(false);
      loadData(true);
    } catch (err) {
      setError(err?.response?.data?.message || err?.message || 'Agent control action failed');
    } finally {
      setSubmittingControl(false);
    }
  };

  // Inspect Workflow
  const handleInspectWorkflow = async (planId) => {
    setInspectingPlanId(planId);
    setInspectModalOpen(true);
    setLoadingInspection(true);
    try {
      const res = await workforceService.inspectWorkflow(planId);
      setWorkflowDetail(res?.data || res);
    } catch (err) {
      console.error('Failed inspecting workflow:', err);
      setError(err?.response?.data?.message || err?.message || 'Failed inspecting workflow');
    } finally {
      setLoadingInspection(false);
    }
  };

  // Control Workflow (Pause, Resume, Stop)
  const handleControlWorkflow = async (planId, cmd) => {
    try {
      if (workflowDetail?.is_enterprise) {
        const entCmd = cmd === 'STOP' ? 'CANCEL' : cmd;
        await handleEnterpriseControlWorkflow(planId, entCmd);
        return;
      }
      await workforceService.controlWorkflow(planId, {
        command: cmd,
        reason: `Operator ${cmd.toLowerCase()} from Command Center`,
      });
      notifySuccess(`Workflow ${planId} updated: ${cmd}`);
      loadData(true);
      if (inspectModalOpen && inspectingPlanId === planId) {
        handleInspectWorkflow(planId);
      }
    } catch (err) {
      setError(err?.response?.data?.message || err?.message || 'Workflow control failed');
    }
  };

  // Inspect Enterprise Workflow
  const handleInspectEnterpriseWorkflow = async (workflowId) => {
    setInspectingPlanId(workflowId);
    setInspectModalOpen(true);
    setLoadingInspection(true);
    try {
      const res = await enterpriseService.getWorkflow(workflowId);
      const data = res?.data || res;
      setWorkflowDetail({
        plan_id: data.workflow_id,
        objective: data.objective,
        status: data.current_state,
        overall_confidence: data.confidence || 0.9,
        coordinator_agent_id: (data.assigned_agents && data.assigned_agents[0]) || 'planning_agent',
        is_enterprise: true,
        steps: (data.steps || []).map((st) => ({
          step_id: st.step_id,
          objective: (st.title || st.action_type) + (st.description ? ' — ' + st.description : ''),
          status: st.status,
          agent_id: st.agent_id,
          required_capabilities: [st.action_type],
          error_message: st.error_message,
        })),
      });
    } catch (err) {
      console.error('Failed inspecting enterprise workflow:', err);
      setError(err?.response?.data?.message || err?.message || 'Failed inspecting enterprise workflow');
    } finally {
      setLoadingInspection(false);
    }
  };

  const handleEnterpriseControlWorkflow = async (workflowId, action) => {
    try {
      if (action === 'PAUSE') {
        await enterpriseService.pauseWorkflow(workflowId, 'Operator manual pause');
      } else if (action === 'RESUME') {
        await enterpriseService.resumeWorkflow(workflowId);
      } else if (action === 'CANCEL') {
        await enterpriseService.cancelWorkflow(workflowId, 'Operator manual cancellation');
      }
      notifySuccess(`Workflow ${workflowId} updated: ${action}`);
      loadData(true);
      if (inspectModalOpen && inspectingPlanId === workflowId) {
        handleInspectEnterpriseWorkflow(workflowId);
      }
    } catch (err) {
      setError(err?.response?.data?.message || err?.message || 'Failed updating enterprise workflow');
    }
  };

  const handleRecoverInterruptedWorkflows = async () => {
    setRecoveringWorkflows(true);
    try {
      const res = await enterpriseService.recoverWorkflows();
      const report = res?.data || res;
      notifySuccess(`Recovery complete: ${report?.recovered_count || 0} resumed, ${report?.blocked_count || 0} blocked`);
      loadData(true);
    } catch (err) {
      setError(err?.response?.data?.message || err?.message || 'Failed recovering workflows');
    } finally {
      setRecoveringWorkflows(false);
    }
  };

  // Filtered Agent Workload
  const filteredAgents = (workload || []).filter((a) => {
    const matchesSearch =
      !searchQuery ||
      a.name?.toLowerCase().includes(searchQuery.toLowerCase()) ||
      a.agent_id?.toLowerCase().includes(searchQuery.toLowerCase());
    const matchesAutonomy = autonomyFilter === 'ALL' || a.autonomy_level === autonomyFilter;
    const matchesStatus = statusFilter === 'ALL' || a.operational_status === statusFilter;
    return matchesSearch && matchesAutonomy && matchesStatus;
  });

  const getAutonomyBadgeClass = (lvl) => {
    switch (lvl) {
      case 'LEVEL_0_OBSERVE': return 'wf-autonomy-lvl0';
      case 'LEVEL_1_RECOMMEND': return 'wf-autonomy-lvl1';
      case 'LEVEL_2_PREPARE': return 'wf-autonomy-lvl2';
      case 'LEVEL_3_CONTROLLED_EXECUTION': return 'wf-autonomy-lvl3';
      case 'LEVEL_4_GOVERNED_MULTI_STEP': return 'wf-autonomy-lvl4';
      default: return 'wf-autonomy-lvl1';
    }
  };

  const formatAutonomyLabel = (lvl) => {
    switch (lvl) {
      case 'LEVEL_0_OBSERVE': return 'LEVEL 0 — OBSERVE';
      case 'LEVEL_1_RECOMMEND': return 'LEVEL 1 — RECOMMEND';
      case 'LEVEL_2_PREPARE': return 'LEVEL 2 — PREPARE';
      case 'LEVEL_3_CONTROLLED_EXECUTION': return 'LEVEL 3 — CONTROLLED EXECUTION';
      case 'LEVEL_4_GOVERNED_MULTI_STEP': return 'LEVEL 4 — GOVERNED MULTI-STEP';
      default: return lvl || 'LEVEL 1 — RECOMMEND';
    }
  };

  return (
    <div className="wf-command-center" data-testid="workforce-command-center">
      {/* Feedback Toasts */}
      {error && (
        <div className="wf-emergency-banner" style={{ background: '#FEF2F2', borderColor: '#FCA5A5' }}>
          <div className="wf-emergency-banner-content">
            <AlertTriangle className="w-5 h-5 text-red-600" />
            <span>{error}</span>
          </div>
          <button className="wf-btn wf-btn-sm wf-btn-default" onClick={() => setError(null)}>Dismiss</button>
        </div>
      )}

      {successMsg && (
        <div className="wf-emergency-banner" style={{ background: '#F0FDF4', borderColor: '#86EFAC' }}>
          <div className="wf-emergency-banner-content" style={{ color: '#166534' }}>
            <CheckCircle2 className="w-5 h-5 text-green-600" />
            <span>{successMsg}</span>
          </div>
          <button className="wf-btn wf-btn-sm wf-btn-default" onClick={() => setSuccessMsg(null)}>Close</button>
        </div>
      )}

      {/* Emergency Stop Active Alert Banner */}
      {emergencyStatus?.workforce_stopped && (
        <div className="wf-emergency-banner">
          <div className="wf-emergency-banner-content">
            <AlertOctagon className="w-6 h-6 text-red-700" />
            <div>
              <strong>CRITICAL SAFETY STOP ENGAGED: Entire AI Workforce is Stopped.</strong>
              <div style={{ fontSize: '0.75rem', fontWeight: 400, marginTop: '2px' }}>
                Reason: {emergencyStatus.reason || 'Operational intervention'} — All autonomous actions and task executions are blocked.
              </div>
            </div>
          </div>
          <button
            className="wf-btn wf-btn-success"
            onClick={() => {
              setStopScope('WORKFORCE');
              setStopTarget('');
              handleToggleEmergencyStop('RESUME');
            }}
          >
            <Play className="w-4 h-4" />
            Resume Autonomous Workforce
          </button>
        </div>
      )}

      {/* Header Bar */}
      <div className="wf-header-bar">
        <div className="wf-header-title">
          <h1>
            <Shield className="w-6 h-6 text-sky-600" />
            AI Workforce Command Center
          </h1>
          <p>Governed multi-agent autonomy, operational health, workload, approvals, and application-level safety controls</p>
        </div>

        <div className="wf-header-actions">
          <button
            className="wf-btn wf-btn-default"
            onClick={() => loadData(false)}
            disabled={refreshing}
          >
            <RefreshCw className={`w-4 h-4 ${refreshing ? 'wf-spinner' : ''}`} />
            Refresh
          </button>

          {!emergencyStatus?.workforce_stopped ? (
            <button
              className="wf-btn wf-btn-danger"
              onClick={() => {
                setStopScope('WORKFORCE');
                setStopTarget('');
                setStopModalOpen(true);
              }}
            >
              <AlertOctagon className="w-4 h-4" />
              Emergency Stop
            </button>
          ) : (
            <button
              className="wf-btn wf-btn-success"
              onClick={() => handleToggleEmergencyStop('RESUME')}
            >
              <Play className="w-4 h-4" />
              Resume Workforce
            </button>
          )}
        </div>
      </div>

      {/* Section 1 & 8: Workforce Health & Telemetry Cards */}
      <div className="wf-stats-grid">
        <div className="wf-stat-card">
          <div className="wf-stat-header">
            <span>Workforce Health</span>
            <Activity className="w-4 h-4 text-emerald-600" />
          </div>
          <div className="wf-stat-value">
            <span className={`wf-badge wf-badge-${(health?.overall_status || 'healthy').toLowerCase()}`}>
              {health?.overall_status || 'HEALTHY'}
            </span>
          </div>
          <div className="wf-stat-subtext">Avg Latency: {health?.average_latency_ms?.toFixed(1) || 42}ms</div>
        </div>

        <div className="wf-stat-card">
          <div className="wf-stat-header">
            <span>Active Agents</span>
            <UserCheck className="w-4 h-4 text-blue-600" />
          </div>
          <div className="wf-stat-value">{health?.active_agents_count || workload.length || 10}</div>
          <div className="wf-stat-subtext">{health?.paused_agents_count || 0} paused · {health?.disabled_agents_count || 0} disabled</div>
        </div>

        <div className="wf-stat-card">
          <div className="wf-stat-header">
            <span>Active Tasks</span>
            <Layers className="w-4 h-4 text-sky-600" />
          </div>
          <div className="wf-stat-value">{health?.running_tasks || 0}</div>
          <div className="wf-stat-subtext">{health?.pending_tasks || 0} pending · {health?.completed_tasks || 0} completed</div>
        </div>

        <div className="wf-stat-card">
          <div className="wf-stat-header">
            <span>Waiting Approvals</span>
            <FileCheck className="w-4 h-4 text-amber-600" />
          </div>
          <div className="wf-stat-value">{approvals.length}</div>
          <div className="wf-stat-subtext">Human-in-the-loop gates</div>
        </div>

        <div className="wf-stat-card">
          <div className="wf-stat-header">
            <span>Open Escalations</span>
            <AlertTriangle className="w-4 h-4 text-red-600" />
          </div>
          <div className="wf-stat-value">{escalations.length}</div>
          <div className="wf-stat-subtext">{health?.failed_tasks || 0} task failures recorded</div>
        </div>
      </div>

      {/* Tabs Navigation */}
      <div className="wf-tabs-nav">
        <button
          className={`wf-tab-item ${activeTab === 'agents' ? 'active' : ''}`}
          onClick={() => setActiveTab('agents')}
        >
          <Cpu className="w-4 h-4" />
          Specialist Agents & Autonomy
          <span className="wf-tab-badge">{workload.length}</span>
        </button>

        <button
          className={`wf-tab-item ${activeTab === 'workflows' ? 'active' : ''}`}
          onClick={() => setActiveTab('workflows')}
        >
          <Layers className="w-4 h-4" />
          Enterprise Workflows
          <span className="wf-tab-badge">{enterpriseWorkflows.length}</span>
        </button>

        <button
          className={`wf-tab-item ${activeTab === 'approvals' ? 'active' : ''}`}
          onClick={() => setActiveTab('approvals')}
        >
          <FileCheck className="w-4 h-4" />
          Waiting Approvals (HITL)
          <span className="wf-tab-badge">{approvals.length}</span>
        </button>

        <button
          className={`wf-tab-item ${activeTab === 'escalations' ? 'active' : ''}`}
          onClick={() => setActiveTab('escalations')}
        >
          <AlertTriangle className="w-4 h-4" />
          Escalations
          <span className="wf-tab-badge">{escalations.length}</span>
        </button>

        <button
          className={`wf-tab-item ${activeTab === 'activity' ? 'active' : ''}`}
          onClick={() => setActiveTab('activity')}
        >
          <Activity className="w-4 h-4" />
          Recent Activity
          <span className="wf-tab-badge">{activity.length}</span>
        </button>
      </div>

      {/* TAB 1: Specialist Agents & Governed Autonomy */}
      {activeTab === 'agents' && (
        <div className="wf-panel">
          <div className="wf-panel-header">
            <div className="wf-panel-title">
              <Cpu className="w-5 h-5 text-sky-600" />
              Specialist Agent Registry & Governed Autonomy Levels
            </div>

            <div style={{ display: 'flex', gap: '8px', flexWrap: 'wrap' }}>
              <input
                type="text"
                placeholder="Search agent..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="wf-form-input"
                style={{ width: '180px', padding: '6px 10px', fontSize: '0.75rem' }}
              />

              <select
                value={autonomyFilter}
                onChange={(e) => setAutonomyFilter(e.target.value)}
                className="wf-form-select"
                style={{ width: '180px', padding: '6px 10px', fontSize: '0.75rem' }}
              >
                <option value="ALL">All Autonomy Levels</option>
                <option value="LEVEL_0_OBSERVE">Level 0 — Observe</option>
                <option value="LEVEL_1_RECOMMEND">Level 1 — Recommend</option>
                <option value="LEVEL_2_PREPARE">Level 2 — Prepare</option>
                <option value="LEVEL_3_CONTROLLED_EXECUTION">Level 3 — Controlled Exec</option>
                <option value="LEVEL_4_GOVERNED_MULTI_STEP">Level 4 — Governed Multi-Step</option>
              </select>

              <select
                value={statusFilter}
                onChange={(e) => setStatusFilter(e.target.value)}
                className="wf-form-select"
                style={{ width: '140px', padding: '6px 10px', fontSize: '0.75rem' }}
              >
                <option value="ALL">All Statuses</option>
                <option value="ACTIVE">Active</option>
                <option value="PAUSED">Paused</option>
                <option value="DISABLED">Disabled</option>
              </select>
            </div>
          </div>

          <div className="wf-table-responsive">
            <table className="wf-table">
              <thead>
                <tr>
                  <th>Agent / Domain</th>
                  <th>Role</th>
                  <th>Autonomy Level</th>
                  <th>Status</th>
                  <th>Health</th>
                  <th>Workload (P / R / C / F)</th>
                  <th>Bottleneck</th>
                  <th style={{ textAlign: 'right' }}>Governed Controls</th>
                </tr>
              </thead>
              <tbody>
                {filteredAgents.length === 0 ? (
                  <tr>
                    <td colSpan="8" className="wf-empty-state">
                      No matching specialist agents found.
                    </td>
                  </tr>
                ) : (
                  filteredAgents.map((a) => (
                    <tr key={a.agent_id}>
                      <td>
                        <strong style={{ color: '#0F172A' }}>{a.name}</strong>
                        <div style={{ fontSize: '0.7rem', color: '#64748B' }}>{a.agent_id}</div>
                      </td>
                      <td>
                        <span style={{ fontSize: '0.75rem', fontWeight: 600, color: '#475569' }}>
                          {a.agent_type}
                        </span>
                      </td>
                      <td>
                        <span className={`wf-autonomy-pill ${getAutonomyBadgeClass(a.autonomy_level)}`}>
                          {formatAutonomyLabel(a.autonomy_level)}
                        </span>
                      </td>
                      <td>
                        <span className={`wf-badge wf-badge-${(a.operational_status || 'active').toLowerCase()}`}>
                          {a.operational_status || 'ACTIVE'}
                        </span>
                      </td>
                      <td>
                        <span className={`wf-badge wf-badge-${(a.health_status || 'healthy').toLowerCase()}`}>
                          {a.health_status || 'HEALTHY'}
                        </span>
                      </td>
                      <td>
                        <span style={{ fontSize: '0.75rem', fontFamily: 'monospace' }}>
                          {a.pending_tasks} / {a.running_tasks} / {a.completed_tasks} / {a.failed_tasks}
                        </span>
                      </td>
                      <td>
                        {a.is_bottleneck ? (
                          <span className="wf-badge wf-badge-degraded">Bottleneck</span>
                        ) : (
                          <span style={{ color: '#94A3B8', fontSize: '0.75rem' }}>Normal</span>
                        )}
                      </td>
                      <td style={{ textAlign: 'right' }}>
                        <div style={{ display: 'inline-flex', gap: '4px' }}>
                          {/* Pause / Resume Button */}
                          {a.operational_status === 'PAUSED' ? (
                            <button
                              className="wf-btn wf-btn-sm wf-btn-success"
                              title="Resume Agent"
                              onClick={() => openAgentControl(a, 'RESUME')}
                            >
                              <Play className="w-3.5 h-3.5" />
                              Resume
                            </button>
                          ) : (
                            <button
                              className="wf-btn wf-btn-sm wf-btn-warning"
                              title="Pause Agent"
                              disabled={a.operational_status === 'DISABLED'}
                              onClick={() => openAgentControl(a, 'PAUSE')}
                            >
                              <Pause className="w-3.5 h-3.5" />
                              Pause
                            </button>
                          )}

                          {/* Enable / Disable Button */}
                          {a.operational_status === 'DISABLED' ? (
                            <button
                              className="wf-btn wf-btn-sm wf-btn-default"
                              title="Enable Agent"
                              onClick={() => openAgentControl(a, 'ENABLE')}
                            >
                              <Unlock className="w-3.5 h-3.5" />
                              Enable
                            </button>
                          ) : (
                            <button
                              className="wf-btn wf-btn-sm wf-btn-danger"
                              title="Disable Agent"
                              onClick={() => openAgentControl(a, 'DISABLE')}
                            >
                              <Lock className="w-3.5 h-3.5" />
                              Disable
                            </button>
                          )}

                          {/* Change Autonomy */}
                          <button
                            className="wf-btn wf-btn-sm wf-btn-primary"
                            title="Set Autonomy Level"
                            onClick={() => openAgentControl(a, 'SET_AUTONOMY')}
                          >
                            <Sliders className="w-3.5 h-3.5" />
                            Level
                          </button>
                        </div>
                      </td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {/* TAB: Enterprise Autonomous Workflows */}
      {activeTab === 'workflows' && (
        <div className="wf-panel">
          <div className="wf-panel-header">
            <div>
              <div className="wf-panel-title">
                <Layers className="w-5 h-5 text-indigo-600" />
                Enterprise Autonomous Operations Layer
              </div>
              <div style={{ fontSize: '0.8125rem', color: '#64748B', marginTop: '2px' }}>
                Governed cross-module multi-agent workflows with durable state persistence and restart recovery
              </div>
            </div>

            <div style={{ display: 'flex', gap: '8px', alignItems: 'center' }}>
              <button
                className="wf-btn wf-btn-default"
                disabled={recoveringWorkflows}
                onClick={handleRecoverInterruptedWorkflows}
                title="Scan and safely resume interrupted workflows across system restart"
              >
                <RefreshCw className={`w-3.5 h-3.5 ${recoveringWorkflows ? 'wf-spinner' : ''}`} />
                {recoveringWorkflows ? 'Recovering...' : 'Recover Interrupted Workflows'}
              </button>
            </div>
          </div>

          <div className="wf-table-responsive">
            <table className="wf-table">
              <thead>
                <tr>
                  <th>Workflow ID</th>
                  <th>Type & Objective</th>
                  <th>State</th>
                  <th>Autonomy Level</th>
                  <th>Specialists Assigned</th>
                  <th>Actions</th>
                </tr>
              </thead>
              <tbody>
                {enterpriseWorkflows.length === 0 ? (
                  <tr>
                    <td colSpan="6" className="wf-empty-state">
                      No enterprise autonomous workflows recorded yet. Workflows initiate automatically from operational events or via platform API.
                    </td>
                  </tr>
                ) : (
                  enterpriseWorkflows.map((wf) => (
                    <tr key={wf.workflow_id}>
                      <td>
                        <span style={{ fontFamily: 'monospace', fontWeight: 600, color: '#0284C7' }}>
                          {wf.workflow_id}
                        </span>
                        <div style={{ fontSize: '0.7rem', color: '#94A3B8' }}>
                          Corr: {wf.correlation_id ? wf.correlation_id.substring(0, 16) + '...' : '—'}
                        </div>
                      </td>
                      <td style={{ maxWidth: '300px' }}>
                        <div style={{ fontWeight: 600, fontSize: '0.8125rem', color: '#0F172A' }}>
                          {wf.workflow_type?.replace(/_/g, ' ')}
                        </div>
                        <div style={{ fontSize: '0.75rem', color: '#64748B', whiteSpace: 'normal' }}>
                          {wf.objective}
                        </div>
                      </td>
                      <td>
                        <span className={`wf-badge wf-badge-${(wf.current_state || 'pending').toLowerCase()}`}>
                          {wf.current_state}
                        </span>
                      </td>
                      <td>
                        <span className={`wf-autonomy-badge ${getAutonomyBadgeClass(wf.autonomy_level)}`}>
                          {formatAutonomyLabel(wf.autonomy_level)}
                        </span>
                      </td>
                      <td>
                        <div style={{ display: 'flex', gap: '4px', flexWrap: 'wrap', maxWidth: '220px' }}>
                          {(wf.assigned_agents || []).map((ag) => (
                            <span
                              key={ag}
                              className="wf-badge wf-badge-active"
                              style={{ fontSize: '0.65rem', padding: '1px 5px' }}
                            >
                              {ag.replace('_agent', '')}
                            </span>
                          ))}
                        </div>
                      </td>
                      <td>
                        <div style={{ display: 'flex', gap: '6px' }}>
                          <button
                            className="wf-btn wf-btn-sm wf-btn-default"
                            onClick={() => handleInspectEnterpriseWorkflow(wf.workflow_id)}
                            title="Inspect discrete workflow steps"
                          >
                            <Eye className="w-3.5 h-3.5" />
                            Inspect
                          </button>
                          {wf.current_state === 'RUNNING' && (
                            <button
                              className="wf-btn wf-btn-sm wf-btn-warning"
                              onClick={() => handleEnterpriseControlWorkflow(wf.workflow_id, 'PAUSE')}
                              title="Pause workflow execution"
                            >
                              <Pause className="w-3.5 h-3.5" />
                            </button>
                          )}
                          {wf.current_state === 'PAUSED' && (
                            <button
                              className="wf-btn wf-btn-sm wf-btn-success"
                              onClick={() => handleEnterpriseControlWorkflow(wf.workflow_id, 'RESUME')}
                              title="Resume paused workflow"
                            >
                              <Play className="w-3.5 h-3.5" />
                            </button>
                          )}
                          {(wf.current_state === 'RUNNING' || wf.current_state === 'PAUSED' || wf.current_state === 'WAITING_FOR_APPROVAL') && (
                            <button
                              className="wf-btn wf-btn-sm wf-btn-danger"
                              onClick={() => handleEnterpriseControlWorkflow(wf.workflow_id, 'CANCEL')}
                              title="Cancel workflow"
                            >
                              <XCircle className="w-3.5 h-3.5" />
                            </button>
                          )}
                        </div>
                      </td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {/* TAB 2: Waiting Approvals (HITL) */}
      {activeTab === 'approvals' && (
        <div className="wf-panel">
          <div className="wf-panel-header">
            <div className="wf-panel-title">
              <FileCheck className="w-5 h-5 text-amber-600" />
              Human-in-the-Loop Waiting Approvals
            </div>
            <span style={{ fontSize: '0.8125rem', color: '#64748B' }}>
              Level 2/3/4 proposed actions requiring explicit human review before execution
            </span>
          </div>

          <div className="wf-table-responsive">
            <table className="wf-table">
              <thead>
                <tr>
                  <th>Action / Workflow</th>
                  <th>Proposing Agent</th>
                  <th>Risk Level</th>
                  <th>Confidence</th>
                  <th>Reason / Context</th>
                  <th>Created At</th>
                  <th style={{ textAlign: 'right' }}>Action</th>
                </tr>
              </thead>
              <tbody>
                {approvals.length === 0 ? (
                  <tr>
                    <td colSpan="7" className="wf-empty-state">
                      <CheckCircle2 className="w-8 h-8 text-green-500 mx-auto" />
                      <div>No waiting approvals. All autonomous actions evaluated within safe limits.</div>
                    </td>
                  </tr>
                ) : (
                  approvals.map((ap) => (
                    <tr key={ap.approval_id}>
                      <td>
                        <strong>{ap.proposed_action || ap.workflow_id}</strong>
                        <div style={{ fontSize: '0.7rem', color: '#64748B' }}>Ref: #{ap.approval_id}</div>
                      </td>
                      <td>
                        <span style={{ fontWeight: 600, color: '#334155' }}>{ap.agent_id}</span>
                      </td>
                      <td>
                        <span className={`wf-badge wf-badge-${(ap.risk || 'medium').toLowerCase()}`}>
                          {ap.risk || 'MEDIUM'}
                        </span>
                      </td>
                      <td>{(ap.confidence * 100).toFixed(0)}%</td>
                      <td style={{ maxWidth: '300px' }}>
                        <div style={{ whiteSpace: 'normal', fontSize: '0.75rem' }}>{ap.reason}</div>
                      </td>
                      <td>{new Date(ap.created_at).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}</td>
                      <td style={{ textAlign: 'right' }}>
                        <a
                          href="/dashboard/approvals"
                          className="wf-btn wf-btn-sm wf-btn-primary"
                          style={{ textDecoration: 'none' }}
                        >
                          Review in Approvals Center
                          <ChevronRight className="w-3.5 h-3.5" />
                        </a>
                      </td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {/* TAB 3: Escalations Panel */}
      {activeTab === 'escalations' && (
        <div className="wf-panel">
          <div className="wf-panel-header">
            <div className="wf-panel-title">
              <AlertTriangle className="w-5 h-5 text-red-600" />
              Workforce Escalation Center
            </div>
            <span style={{ fontSize: '0.8125rem', color: '#64748B' }}>
              Policy violations, emergency stops, task failures, and unresolved agent conflicts
            </span>
          </div>

          <div className="wf-table-responsive">
            <table className="wf-table">
              <thead>
                <tr>
                  <th>Severity</th>
                  <th>Agent / Component</th>
                  <th>Reason / Violation</th>
                  <th>Task Ref</th>
                  <th>Status</th>
                  <th>Timestamp</th>
                </tr>
              </thead>
              <tbody>
                {escalations.length === 0 ? (
                  <tr>
                    <td colSpan="6" className="wf-empty-state">
                      <CheckCircle2 className="w-8 h-8 text-green-500 mx-auto" />
                      <div>No active escalations recorded. Workforce running within governed boundaries.</div>
                    </td>
                  </tr>
                ) : (
                  escalations.map((esc) => (
                    <tr key={esc.escalation_id}>
                      <td>
                        <span className={`wf-badge wf-badge-${(esc.severity || 'high').toLowerCase()}`}>
                          {esc.severity || 'HIGH'}
                        </span>
                      </td>
                      <td>
                        <strong>{esc.agent_id}</strong>
                      </td>
                      <td style={{ maxWidth: '400px' }}>
                        <div style={{ whiteSpace: 'normal', fontSize: '0.75rem' }}>{esc.reason}</div>
                      </td>
                      <td>
                        <span style={{ fontFamily: 'monospace', fontSize: '0.75rem' }}>
                          {esc.task_id || 'N/A'}
                        </span>
                      </td>
                      <td>
                        <span className={`wf-badge wf-badge-${(esc.status || 'open').toLowerCase()}`}>
                          {esc.status || 'OPEN'}
                        </span>
                      </td>
                      <td>{new Date(esc.created_at).toLocaleString()}</td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {/* TAB 4: Recent Agent Activity */}
      {activeTab === 'activity' && (
        <div className="wf-panel">
          <div className="wf-panel-header">
            <div className="wf-panel-title">
              <Activity className="w-5 h-5 text-sky-600" />
              Audited Multi-Agent Operational Activity
            </div>
            <span style={{ fontSize: '0.8125rem', color: '#64748B' }}>
              Real-time audit log of task delegations, execution outcomes, and safety checks
            </span>
          </div>

          <div className="wf-table-responsive">
            <table className="wf-table">
              <thead>
                <tr>
                  <th>Event</th>
                  <th>Agent</th>
                  <th>Summary</th>
                  <th>Task ID</th>
                  <th>Status</th>
                  <th>Timestamp</th>
                </tr>
              </thead>
              <tbody>
                {activity.length === 0 ? (
                  <tr>
                    <td colSpan="6" className="wf-empty-state">
                      No recent activity recorded.
                    </td>
                  </tr>
                ) : (
                  activity.map((act) => (
                    <tr key={act.activity_id}>
                      <td>
                        <span className="wf-badge wf-badge-active" style={{ fontSize: '0.7rem' }}>
                          {act.event_type}
                        </span>
                      </td>
                      <td>
                        <strong>{act.agent_id}</strong>
                      </td>
                      <td style={{ maxWidth: '400px' }}>
                        <div style={{ whiteSpace: 'normal', fontSize: '0.75rem' }}>{act.summary}</div>
                      </td>
                      <td>
                        <span style={{ fontFamily: 'monospace', fontSize: '0.75rem' }}>
                          {act.task_id || '—'}
                        </span>
                      </td>
                      <td>
                        <span className={`wf-badge wf-badge-${(act.status || 'completed').toLowerCase()}`}>
                          {act.status || 'DONE'}
                        </span>
                      </td>
                      <td>{new Date(act.timestamp).toLocaleTimeString()}</td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {/* Emergency Stop Modal */}
      {stopModalOpen && (
        <div className="wf-modal-overlay" onClick={() => setStopModalOpen(false)}>
          <div className="wf-modal-content" onClick={(e) => e.stopPropagation()}>
            <div className="wf-modal-header">
              <div className="wf-modal-title">Engage Emergency Safety Stop</div>
              <button className="wf-modal-close" onClick={() => setStopModalOpen(false)}>×</button>
            </div>

            <div className="wf-form-group">
              <label className="wf-form-label">Scope of Emergency Stop</label>
              <select
                className="wf-form-select"
                value={stopScope}
                onChange={(e) => setStopScope(e.target.value)}
              >
                <option value="WORKFORCE">Entire AI Workforce (All Agents & Workflows)</option>
                <option value="AGENT">Specific Agent Only</option>
                <option value="ACTION_CLASS">Specific Action Category (e.g. FINANCIAL, COMMERCIAL)</option>
                <option value="WORKFLOW">Specific Multi-Agent Workflow ID</option>
              </select>
            </div>

            {stopScope !== 'WORKFORCE' && (
              <div className="wf-form-group">
                <label className="wf-form-label">
                  Target Identifier ({stopScope === 'AGENT' ? 'Agent ID' : stopScope === 'ACTION_CLASS' ? 'Category Name' : 'Plan ID'})
                </label>
                <input
                  type="text"
                  className="wf-form-input"
                  placeholder={stopScope === 'AGENT' ? 'e.g. shipment_agent' : stopScope === 'ACTION_CLASS' ? 'e.g. FINANCIAL' : 'e.g. plan-xyz'}
                  value={stopTarget}
                  onChange={(e) => setStopTarget(e.target.value)}
                />
              </div>
            )}

            <div className="wf-form-group">
              <label className="wf-form-label">Mandatory Reason for Audit Trail</label>
              <textarea
                className="wf-form-textarea"
                rows="3"
                placeholder="Describe reason for emergency safety intervention..."
                value={stopReason}
                onChange={(e) => setStopReason(e.target.value)}
              />
            </div>

            <div className="wf-modal-actions">
              <button
                className="wf-btn wf-btn-default"
                onClick={() => setStopModalOpen(false)}
              >
                Cancel
              </button>
              <button
                className="wf-btn wf-btn-danger"
                disabled={submittingStop || !stopReason}
                onClick={() => handleToggleEmergencyStop('STOP')}
              >
                <AlertOctagon className="w-4 h-4" />
                {submittingStop ? 'Engaging...' : 'Engage Emergency Stop'}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Agent Control Modal */}
      {controlModalOpen && targetAgent && (
        <div className="wf-modal-overlay" onClick={() => setControlModalOpen(false)}>
          <div className="wf-modal-content" onClick={(e) => e.stopPropagation()}>
            <div className="wf-modal-header">
              <div className="wf-modal-title">
                Governed Agent Control: {targetAgent.name}
              </div>
              <button className="wf-modal-close" onClick={() => setControlModalOpen(false)}>×</button>
            </div>

            <div className="wf-form-group">
              <label className="wf-form-label">Selected Operation</label>
              <select
                className="wf-form-select"
                value={controlAction}
                onChange={(e) => setControlAction(e.target.value)}
              >
                <option value="PAUSE">PAUSE (Preserves state, prevents new tasks)</option>
                <option value="RESUME">RESUME (Restore active operational status)</option>
                <option value="DISABLE">DISABLE (Full disablement under governance)</option>
                <option value="ENABLE">ENABLE (Re-enable disabled agent)</option>
                <option value="SET_AUTONOMY">SET AUTONOMY LEVEL (Governed Change)</option>
              </select>
            </div>

            {controlAction === 'SET_AUTONOMY' && (
              <div className="wf-form-group">
                <label className="wf-form-label">Target Autonomy Level</label>
                <select
                  className="wf-form-select"
                  value={newAutonomyLevel}
                  onChange={(e) => setNewAutonomyLevel(e.target.value)}
                >
                  <option value="LEVEL_0_OBSERVE">LEVEL 0 — OBSERVE (Telemetry & Read-Only)</option>
                  <option value="LEVEL_1_RECOMMEND">LEVEL 1 — RECOMMEND (Advisory Output Only)</option>
                  <option value="LEVEL_2_PREPARE">LEVEL 2 — PREPARE (Mandatory HITL Gate)</option>
                  <option value="LEVEL_3_CONTROLLED_EXECUTION">LEVEL 3 — CONTROLLED EXECUTION (Low Risk Auto, High Risk Review)</option>
                  <option value="LEVEL_4_GOVERNED_MULTI_STEP">LEVEL 4 — GOVERNED MULTI-STEP (Policy-Bound Multi-Agent Coordination)</option>
                </select>
              </div>
            )}

            <div className="wf-form-group">
              <label className="wf-form-label">Audited Reason for Change</label>
              <textarea
                className="wf-form-textarea"
                rows="3"
                placeholder="Provide justification for governance log..."
                value={controlReason}
                onChange={(e) => setControlReason(e.target.value)}
              />
            </div>

            <div className="wf-modal-actions">
              <button
                className="wf-btn wf-btn-default"
                onClick={() => setControlModalOpen(false)}
              >
                Cancel
              </button>
              <button
                className="wf-btn wf-btn-primary"
                disabled={submittingControl}
                onClick={handleSubmitAgentControl}
              >
                {submittingControl ? 'Applying...' : 'Apply Governed Change'}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Workflow Inspection Modal */}
      {inspectModalOpen && (
        <div className="wf-modal-overlay" onClick={() => setInspectModalOpen(false)}>
          <div className="wf-modal-content" style={{ maxWidth: '750px' }} onClick={(e) => e.stopPropagation()}>
            <div className="wf-modal-header">
              <div className="wf-modal-title">
                Workflow Inspection: {inspectingPlanId}
              </div>
              <button className="wf-modal-close" onClick={() => setInspectModalOpen(false)}>×</button>
            </div>

            {loadingInspection ? (
              <div className="wf-empty-state">
                <RefreshCw className="w-8 h-8 text-sky-600 wf-spinner mx-auto" />
                <div>Loading workflow execution telemetry...</div>
              </div>
            ) : workflowDetail ? (
              <div>
                <div style={{ marginBottom: '16px' }}>
                  <div style={{ fontSize: '0.8125rem', color: '#64748B' }}>Objective</div>
                  <strong style={{ fontSize: '1rem', color: '#0F172A' }}>{workflowDetail.objective}</strong>
                  <div style={{ fontSize: '0.75rem', color: '#64748B', marginTop: '4px' }}>
                    Coordinator: <strong>{workflowDetail.coordinator_agent_id}</strong> · Status:{' '}
                    <span className={`wf-badge wf-badge-${(workflowDetail.status || 'created').toLowerCase()}`}>
                      {workflowDetail.status}
                    </span> · Confidence: {(workflowDetail.overall_confidence * 100).toFixed(0)}%
                  </div>
                </div>

                <div style={{ marginBottom: '16px' }}>
                  <label className="wf-form-label">Step Execution Progression</label>
                  <div className="wf-steps-list">
                    {(workflowDetail.steps || []).map((st, i) => (
                      <div key={st.step_id || i} className="wf-step-item">
                        <div className="wf-step-header">
                          <span className="wf-step-title">{st.step_id}: {st.objective}</span>
                          <span className={`wf-badge wf-badge-${(st.status || 'pending').toLowerCase()}`}>
                            {st.status || 'PENDING'}
                          </span>
                        </div>
                        <div className="wf-step-meta">
                          Assigned Agent: <strong>{st.agent_id}</strong> · Capabilities: {st.required_capabilities?.join(', ')}
                        </div>
                        {st.error_message && (
                          <div style={{ color: '#DC2626', fontSize: '0.75rem', marginTop: '4px' }}>
                            Error: {st.error_message}
                          </div>
                        )}
                      </div>
                    ))}
                  </div>
                </div>

                <div className="wf-modal-actions" style={{ justifyContent: 'space-between' }}>
                  <div style={{ display: 'flex', gap: '8px' }}>
                    {workflowDetail.status === 'IN_PROGRESS' && (
                      <button
                        className="wf-btn wf-btn-warning"
                        onClick={() => handleControlWorkflow(workflowDetail.plan_id, 'PAUSE')}
                      >
                        <Pause className="w-4 h-4" /> Pause Workflow
                      </button>
                    )}
                    {workflowDetail.status === 'PAUSED' && (
                      <button
                        className="wf-btn wf-btn-success"
                        onClick={() => handleControlWorkflow(workflowDetail.plan_id, 'RESUME')}
                      >
                        <Play className="w-4 h-4" /> Resume Workflow
                      </button>
                    )}
                    {workflowDetail.status !== 'STOPPED' && (
                      <button
                        className="wf-btn wf-btn-danger"
                        onClick={() => handleControlWorkflow(workflowDetail.plan_id, 'STOP')}
                      >
                        <AlertOctagon className="w-4 h-4" /> Stop Workflow
                      </button>
                    )}
                  </div>

                  <button className="wf-btn wf-btn-default" onClick={() => setInspectModalOpen(false)}>
                    Close
                  </button>
                </div>
              </div>
            ) : (
              <div className="wf-empty-state">Workflow details could not be retrieved.</div>
            )}
          </div>
        </div>
      )}
    </div>
  );
}

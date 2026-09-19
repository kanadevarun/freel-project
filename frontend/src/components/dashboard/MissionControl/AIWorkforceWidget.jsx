import React, { useState, useEffect, useCallback, useRef } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  Bot,
  Activity,
  CheckCircle2,
  AlertCircle,
  AlertTriangle,
  Clock,
  RefreshCw,
  ExternalLink,
  Zap,
  RotateCcw,
  X,
  ArrowUpRight,
  Brain,
} from 'lucide-react';
import aiTaskService from '../../../services/aiTaskService';
import memoryService from '../../../services/memoryService';
import monitoringService from '../../../services/monitoringService';
import AgentStatusBadge from '../../agent/AgentStatusBadge';
import { WORKFORCE_STATUS } from '../../../utils/workforceStatus';
import './AIWorkforceWidget.css';

const POLLING_INTERVAL_MS = 30000; // 30 seconds

export default function AIWorkforceWidget({ onOpenRecord }) {
  const navigate = useNavigate();

  // Primary data states
  const [summary, setSummary] = useState(null);
  const [tasks, setTasks] = useState([]);
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  const [error, setError] = useState(null);
  const [isUnauthorized, setIsUnauthorized] = useState(false);
  const [lastUpdated, setLastUpdated] = useState(null);
  const [memoryStats, setMemoryStats] = useState(null);
  const [monitoringHealth, setMonitoringHealth] = useState(null);

  // Filter & interaction states
  const [activeTab, setActiveTab] = useState('ALL'); // 'ALL' | 'ATTENTION' | 'APPROVAL' | 'ACTIVE'
  const [selectedAgentKey, setSelectedAgentKey] = useState(null); // Filter by specific agent
  const [selectedTask, setSelectedTask] = useState(null); // Modal detail
  const [actionLoading, setActionLoading] = useState({});

  const isMountedRef = useRef(true);

  // Fetch summary and recent tasks
  const fetchData = useCallback(async (isManualRefresh = false) => {
    if (isManualRefresh) setRefreshing(true);
    setError(null);
    setIsUnauthorized(false);

    try {
      const [summaryRes, tasksRes, memRes, monitorRes] = await Promise.all([
        aiTaskService.getWorkforceSummary(),
        aiTaskService.getWorkforceTasks({ limit: 8 }),
        memoryService.getStats().catch(() => null),
        monitoringService.getHealthSummary().catch(() => null),
      ]);

      if (!isMountedRef.current) return;

      if (memRes?.data) setMemoryStats(memRes.data);
      if (monitorRes?.data) setMonitoringHealth(monitorRes.data);
      const summaryData = summaryRes?.data || summaryRes || {};
      const tasksData = tasksRes?.data || tasksRes || {};

      const rawTasks = tasksData.tasks || tasksData.items || (Array.isArray(tasksData) ? tasksData : []);
      const normalizedTasks = rawTasks.map((t) => {
        const taskId = t.task_id || t.id;
        const status = t.workforce_status || t.status || 'unknown';
        const module = t.business_module || t.module || 'SYSTEM';
        const requiresApproval = t.requires_approval !== undefined
          ? t.requires_approval
          : (!!t.approval_id || status === 'waiting_for_approval');
        return {
          ...t,
          task_id: taskId,
          id: taskId,
          workforce_status: status,
          status: status,
          business_module: module,
          module: module,
          safe_error_msg: t.safe_error_msg || t.error_message || '',
          safe_error_category: t.safe_error_category || t.error_category || '',
          requires_approval: requiresApproval,
          duration_ms: t.duration_ms ?? 0,
          mock_mode: t.mock_mode || false,
          provider_failover: t.provider_failover || false,
          related_ref: t.related_ref || (taskId ? `#${String(taskId).slice(0, 8)}` : '—'),
        };
      });

      setSummary(summaryData);
      setTasks(normalizedTasks);
      setLastUpdated(new Date());
    } catch (err) {
      if (!isMountedRef.current) return;
      console.error('Failed to fetch AI workforce metrics:', err);
      if (err?.status === 401 || err?.response?.status === 401) {
        setIsUnauthorized(true);
      } else {
        const msg = err?.response?.data?.message || err?.message || 'Unable to connect to AI workforce runtime.';
        setError(msg);
      }
    } finally {
      if (isMountedRef.current) {
        setLoading(false);
        setRefreshing(false);
      }
    }
  }, []);

  // Polling with document visibility detection
  useEffect(() => {
    isMountedRef.current = true;
    fetchData();

    let intervalId = null;

    const startPolling = () => {
      if (!intervalId) {
        intervalId = setInterval(() => {
          if (document.visibilityState === 'visible') {
            fetchData();
          }
        }, POLLING_INTERVAL_MS);
      }
    };

    const stopPolling = () => {
      if (intervalId) {
        clearInterval(intervalId);
        intervalId = null;
      }
    };

    const handleVisibilityChange = () => {
      if (document.visibilityState === 'visible') {
        fetchData();
        startPolling();
      } else {
        stopPolling();
      }
    };

    startPolling();
    document.addEventListener('visibilitychange', handleVisibilityChange);

    return () => {
      isMountedRef.current = false;
      stopPolling();
      document.removeEventListener('visibilitychange', handleVisibilityChange);
    };
  }, [fetchData]);

  // Safe retry handler
  const handleRetryTask = async (taskId, e) => {
    if (e) e.stopPropagation();
    setActionLoading((prev) => ({ ...prev, [taskId]: 'retry' }));
    try {
      await aiTaskService.retryTask(taskId);
      await fetchData();
      if (selectedTask?.task_id === taskId) {
        setSelectedTask(null);
      }
    } catch (err) {
      console.error('Failed to retry task:', err);
      alert('Unable to retry task. Please verify permissions or task status.');
    } finally {
      setActionLoading((prev) => ({ ...prev, [taskId]: null }));
    }
  };

  // Safe cancel handler
  const handleCancelTask = async (taskId, e) => {
    if (e) e.stopPropagation();
    if (!window.confirm('Cancel this in-flight AI task?')) return;

    setActionLoading((prev) => ({ ...prev, [taskId]: 'cancel' }));
    try {
      await aiTaskService.cancelTask(taskId, 'Cancelled by operator via Workforce Monitor');
      await fetchData();
      if (selectedTask?.task_id === taskId) {
        setSelectedTask(null);
      }
    } catch (err) {
      console.error('Failed to cancel task:', err);
      alert('Unable to cancel task. The task may have already completed.');
    } finally {
      setActionLoading((prev) => ({ ...prev, [taskId]: null }));
    }
  };

  // Safe navigation to related record
  const handleNavigateToRecord = (task, e) => {
    if (e) e.stopPropagation();
    if (onOpenRecord) {
      onOpenRecord(task);
      return;
    }

    const mod = String(task.related_module || task.business_module || '').toLowerCase();
    const ref = task.related_id || task.related_ref;

    switch (mod) {
      case 'rfq':
      case 'quotes':
      case 'pricing':
        navigate('/dashboard/rfqs');
        break;
      case 'contracts':
      case 'compliance':
        navigate('/dashboard/contracts');
        break;
      case 'invoices':
      case 'finance':
        navigate('/dashboard/invoices');
        break;
      case 'shipments':
      case 'tracking':
      case 'operations':
        navigate('/dashboard/shipments');
        break;
      case 'leads':
      case 'sales':
      case 'outreach':
        navigate('/dashboard/leads');
        break;
      default:
        break;
    }
  };

  // Format relative or exact time
  const formatTime = (isoString) => {
    if (!isoString) return '—';
    try {
      const d = new Date(isoString);
      return d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' });
    } catch {
      return String(isoString);
    }
  };

  // Format duration in ms/s
  const formatDuration = (ms) => {
    if (!ms && ms !== 0) return '—';
    if (ms < 1000) return `${ms}ms`;
    return `${(ms / 1000).toFixed(1)}s`;
  };

  // 1. Loading State
  if (loading && !summary) {
    return (
      <div className="ai-wf-dashboard-grid" aria-label="Loading AI Workforce">
        <div className="ai-wf-card ai-wf-loading-state">
          <div className="ai-wf-spinner" />
          <p>Connecting to AI workforce runtime and queue diagnostics...</p>
        </div>
      </div>
    );
  }

  // 2. Unauthorized State
  if (isUnauthorized) {
    return (
      <div className="ai-wf-dashboard-grid">
        <div className="ai-wf-card ai-wf-error-state">
          <AlertCircle size={24} />
          <strong>Operational Access Required</strong>
          <p>You do not have permission to view AI workforce operational telemetry.</p>
        </div>
      </div>
    );
  }

  // 3. Error State
  if (error && !summary) {
    return (
      <div className="ai-wf-dashboard-grid">
        <div className="ai-wf-card ai-wf-error-state">
          <AlertTriangle size={24} />
          <strong>Workforce Telemetry Unavailable</strong>
          <p>{error}</p>
          <button className="ai-wf-btn-light" onClick={() => fetchData(true)}>
            <RefreshCw size={13} /> Retry Connection
          </button>
        </div>
      </div>
    );
  }

  const counts = summary?.counts || {
    total_active: summary?.total_active_tasks ?? 0,
    processing: summary?.processing_tasks ?? 0,
    queued: summary?.queued_tasks ?? 0,
    waiting_for_approval: summary?.waiting_for_approval_tasks ?? 0,
    attention_required: (summary?.failed_tasks ?? 0) + (summary?.stale_tasks ?? 0),
    failed: summary?.failed_tasks ?? 0,
    stale: summary?.stale_tasks ?? 0,
    completed_24h: summary?.completed_recent_24h ?? 0,
    avg_duration_ms: summary?.avg_duration_ms ?? 0,
  };
  const health = summary?.health || {};
  const agents = Array.isArray(summary?.agents)
    ? summary.agents
    : summary?.by_agent
    ? Object.entries(summary.by_agent).map(([key, a]) => ({
        agent_key: a.agent_key || key,
        name: a.display_name || a.name || key,
        module: a.module,
        active_tasks: a.active_tasks ?? 0,
        completed_24h: a.completed_24h ?? 0,
        failed_24h: a.failed_24h ?? 0,
        waiting_approvals: a.waiting_approvals ?? 0,
        status: a.status || 'IDLE',
      }))
    : [];
  const systemHealth = String(health.overall_status || 'healthy').toLowerCase();
  const checkpointHealth = health.checkpoint_status || health.checkpoint_storage || 'healthy';
  const workerIsActive = health.worker_status === 'healthy' || health.worker_status === 'degraded' || health.worker_active === true;
  const providerIsReady = health.primary_provider_status === 'healthy' || health.provider_ready === true;
  const failoverCount = summary?.failover_completed_tasks ?? summary?.completed_with_failover_24h ?? 0;
  const mockCount = summary?.mock_completed_tasks ?? summary?.completed_in_mock_mode_24h ?? 0;

  // Filter tasks
  let filteredTasks = tasks;
  if (selectedAgentKey) {
    filteredTasks = filteredTasks.filter((t) => t.agent_key === selectedAgentKey);
  }
  if (activeTab === 'ATTENTION') {
    filteredTasks = filteredTasks.filter(
      (t) => t.workforce_status === WORKFORCE_STATUS.FAILED || t.workforce_status === WORKFORCE_STATUS.STALE
    );
  } else if (activeTab === 'APPROVAL') {
    filteredTasks = filteredTasks.filter(
      (t) => t.workforce_status === WORKFORCE_STATUS.WAITING_FOR_APPROVAL || t.requires_approval
    );
  } else if (activeTab === 'ACTIVE') {
    filteredTasks = filteredTasks.filter(
      (t) =>
        t.workforce_status === WORKFORCE_STATUS.PROCESSING ||
        t.workforce_status === WORKFORCE_STATUS.CLAIMED ||
        t.workforce_status === WORKFORCE_STATUS.QUEUED ||
        t.workforce_status === WORKFORCE_STATUS.RETRYING
    );
  }

  return (
    <div className="ai-wf-dashboard-grid" aria-label="AI Workforce Section">
      {/* ── CARD 1 (LEFT): AI WORKFORCE OVERVIEW & OPERATIONAL STATUS ── */}
      <section className="ai-wf-card ai-wf-overview-card" aria-label="AI Workforce Status">
        {/* Header */}
        <div className="ai-wf-card-header">
          <div className="ai-wf-card-header-left">
            <div className="ai-wf-icon-box">
              <Bot size={20} />
            </div>
            <div className="ai-wf-title-group">
              <div className="ai-wf-title-row">
                <h3 className="ai-wf-card-title">
                  AI Workforce Command & Telemetry
                </h3>
                <span className={`ai-wf-health-pill ${systemHealth}`}>
                  <span className={`ai-wf-health-dot ${systemHealth !== 'unavailable' ? 'pulse' : ''}`} />
                  {systemHealth.charAt(0).toUpperCase() + systemHealth.slice(1)}
                </span>
              </div>
              <p className="ai-wf-card-subtitle">
                Live multi-agent orchestration, LangGraph persistence, and queue health
              </p>
            </div>
          </div>

          <div className="ai-wf-card-header-right">
            {lastUpdated && (
              <span className="ai-wf-timestamp-text" title={lastUpdated.toISOString()}>
                Last updated: {lastUpdated.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
              </span>
            )}
            <button
              className="ai-wf-icon-btn"
              onClick={() => fetchData(true)}
              disabled={refreshing}
              title="Refresh workforce metrics"
            >
              <RefreshCw size={13} className={refreshing ? 'ai-wf-spin' : ''} />
            </button>
          </div>
        </div>

        {/* Mode / Failover Notices if active */}
        {(failoverCount > 0 || mockCount > 0) && (
          <div className="ai-wf-mode-notices">
            {failoverCount > 0 && (
              <div className="ai-wf-notice-pill failover">
                <Zap size={12} />
                <span><strong>{failoverCount} tasks</strong> completed via secondary failover provider</span>
              </div>
            )}
            {mockCount > 0 && (
              <div className="ai-wf-notice-pill mock">
                <Bot size={12} />
                <span><strong>{mockCount} tasks</strong> executed in deterministic mock engine</span>
              </div>
            )}
          </div>
        )}

        {/* AI Memory & Personalization telemetry banner */}
        <div className="ai-wf-telemetry-banner">
          <div className="ai-wf-telemetry-banner-left">
            <Brain size={15} color="#2563eb" />
            <span><strong>AI Memory:</strong> {memoryStats?.personalization_enabled !== false ? 'Active' : 'Disabled'} ({memoryStats?.active_personal_count || 0} Personal, {memoryStats?.active_org_count || 0} Org-Wide)</span>
          </div>
          <button 
            type="button"
            onClick={() => navigate('/dashboard/settings/memory')}
            className="ai-wf-banner-action"
          >
            Manage Preferences →
          </button>
        </div>

        {/* Live AI Performance & Monitoring Strip (Phase 2 Task 2.11) */}
        <div className="ai-wf-telemetry-banner" style={{ marginBottom: '16px' }}>
          <div className="ai-wf-telemetry-banner-left">
            <Activity size={15} color="#16a34a" />
            <span>
              <strong>AI Observability:</strong>{' '}
              <span style={{ 
                color: monitoringHealth?.overall_status === 'DEGRADED' ? '#b45309' : '#15803d',
                fontWeight: 700
              }}>
                {monitoringHealth?.overall_status || 'HEALTHY'}
              </span>
              {' • '}
              {monitoringHealth?.success_rate_24h ? `${monitoringHealth.success_rate_24h.toFixed(0)}% success` : '100% success'}
              {monitoringHealth?.p95_latency_ms_24h ? ` • ${monitoringHealth.p95_latency_ms_24h}ms P95` : ''}
            </span>
          </div>
          <button 
            type="button"
            onClick={() => navigate('/dashboard/settings/monitoring')}
            className="ai-wf-banner-action"
          >
            View Observability →
          </button>
        </div>

        {/* 4 Small Light Metric Cards */}
        <div className="ai-wf-metrics-row">
          {/* Active Workflows */}
          <div className="ai-wf-metric-box">
            <div className="ai-wf-metric-top">
              <div className="ai-wf-metric-icon blue">
                <Activity size={15} />
              </div>
              <span className="ai-wf-metric-number">{counts.total_active ?? 0}</span>
            </div>
            <div className="ai-wf-metric-label">Active Workflows</div>
            <div className="ai-wf-metric-desc">
              {counts.processing ?? 0} running, {counts.queued ?? 0} queued
            </div>
          </div>

          {/* Awaiting Sign-Off */}
          <div className="ai-wf-metric-box">
            <div className="ai-wf-metric-top">
              <div className="ai-wf-metric-icon amber">
                <Clock size={15} />
              </div>
              <span className="ai-wf-metric-number">{counts.waiting_for_approval ?? 0}</span>
            </div>
            <div className="ai-wf-metric-label">Awaiting Sign-Off</div>
            <div className="ai-wf-metric-desc">Human-in-the-loop</div>
          </div>

          {/* Attention Required */}
          <div className="ai-wf-metric-box">
            <div className="ai-wf-metric-top">
              <div className="ai-wf-metric-icon red">
                <AlertTriangle size={15} />
              </div>
              <span className="ai-wf-metric-number">{counts.attention_required ?? 0}</span>
            </div>
            <div className="ai-wf-metric-label">Attention Required</div>
            <div className="ai-wf-metric-desc">
              {counts.failed ?? 0} failed, {counts.stale ?? 0} stale
            </div>
          </div>

          {/* Completed 24h */}
          <div className="ai-wf-metric-box">
            <div className="ai-wf-metric-top">
              <div className="ai-wf-metric-icon green">
                <CheckCircle2 size={15} />
              </div>
              <span className="ai-wf-metric-number">{counts.completed_24h ?? 0}</span>
            </div>
            <div className="ai-wf-metric-label">Completed (24h)</div>
            <div className="ai-wf-metric-desc">
              {counts.avg_duration_ms ? `Avg: ${formatDuration(counts.avg_duration_ms)}` : '—'}
            </div>
          </div>
        </div>

        {/* Autonomous Agents Operational Status */}
        <div className="ai-wf-subhead" style={{ marginTop: '10px' }}>
          <span>Autonomous Agents Operational Status</span>
          {selectedAgentKey && (
            <button
              className="ai-wf-clear-filter-btn"
              onClick={() => setSelectedAgentKey(null)}
            >
              Clear Filter
            </button>
          )}
        </div>

        <div className="ai-wf-agents-list">
          {agents.map((agent) => {
            const isSelected = selectedAgentKey === agent.agent_key;
            const statusClass = String(agent.status || 'idle').toLowerCase();
            return (
              <div
                key={agent.agent_key}
                className={`ai-wf-agent-item ${isSelected ? 'selected' : ''}`}
                onClick={() => setSelectedAgentKey(isSelected ? null : agent.agent_key)}
                title={`Click to filter tasks by ${agent.name}`}
              >
                <div className="ai-wf-agent-text">
                  <span className="ai-wf-agent-name">{agent.name}</span>
                  <span className="ai-wf-agent-sub">
                    {agent.active_tasks > 0
                      ? `${agent.active_tasks} active`
                      : `${agent.completed_24h} done (24h)`}
                  </span>
                </div>
                <span className={`ai-wf-agent-badge ${statusClass}`}>{agent.status}</span>
              </div>
            );
          })}
        </div>

        {/* Infrastructure Diagnostics Strip */}
        <div className="ai-wf-infra-bar" style={{ marginTop: '8px' }}>
          <div className="ai-wf-infra-signals">
            <div className="ai-wf-signal-item">
              <span className={`ai-wf-signal-dot ${health.sidecar_status === 'healthy' ? 'green' : 'amber'}`} />
              <span>Sidecar: {health.sidecar_status || 'healthy'}</span>
            </div>
            <div className="ai-wf-signal-item">
              <span className={`ai-wf-signal-dot ${checkpointHealth === 'healthy' ? 'green' : 'amber'}`} />
              <span>Checkpointer: {checkpointHealth || 'healthy'}</span>
            </div>
            <div className="ai-wf-signal-item">
              <span className={`ai-wf-signal-dot ${workerIsActive ? 'green' : 'amber'}`} />
              <span>Worker: {workerIsActive ? 'Active' : 'Idle / Standby'}</span>
            </div>
            <div className="ai-wf-signal-item">
              <span className={`ai-wf-signal-dot ${providerIsReady ? 'green' : 'amber'}`} />
              <span>Providers: {providerIsReady ? 'Configured' : 'Degraded'}</span>
            </div>
          </div>
          {health.last_worker_heartbeat && (
            <span className="ai-wf-heartbeat-text">
              Heartbeat: {formatTime(health.last_worker_heartbeat)}
            </span>
          )}
        </div>
      </section>

      {/* ── CARD 2 (RIGHT): RECENT AI OPERATIONS ── */}
      <section className="ai-wf-card ai-wf-operations-card" aria-label="Recent AI Operations">
        {/* Header */}
        <div className="ai-wf-card-header">
          <div className="ai-wf-card-header-left">
            <h4 className="ai-wf-card-title">
              Recent AI Operations {selectedAgentKey ? `(${selectedAgentKey})` : ''}
            </h4>
          </div>
          <div className="ai-wf-card-header-right">
            <div className="ai-wf-tab-pills">
              <button
                className={`ai-wf-tab-pill ${activeTab === 'ALL' ? 'active' : ''}`}
                onClick={() => setActiveTab('ALL')}
              >
                All Recent
              </button>
              <button
                className={`ai-wf-tab-pill ${activeTab === 'ATTENTION' ? 'active' : ''}`}
                onClick={() => setActiveTab('ATTENTION')}
              >
                Attention ({counts.attention_required ?? 0})
              </button>
              <button
                className={`ai-wf-tab-pill ${activeTab === 'APPROVAL' ? 'active' : ''}`}
                onClick={() => setActiveTab('APPROVAL')}
              >
                Sign-Off ({counts.waiting_for_approval ?? 0})
              </button>
              <button
                className={`ai-wf-tab-pill ${activeTab === 'ACTIVE' ? 'active' : ''}`}
                onClick={() => setActiveTab('ACTIVE')}
              >
                Active ({counts.total_active ?? 0})
              </button>
            </div>
            <button
              className="ai-wf-link-btn"
              onClick={() => navigate('/dashboard/approvals')}
              title="Open Approvals & Workflow Queue"
            >
              View All <ArrowUpRight size={14} />
            </button>
          </div>
        </div>

        {/* Table Content */}
        {filteredTasks.length === 0 ? (
          <div className="ai-wf-empty-box">
            <Bot size={22} className="text-slate-400" />
            <p>No AI tasks found matching the active criteria.</p>
          </div>
        ) : (
          <div className="ai-wf-table-responsive">
            <table className="ai-wf-light-table">
              <thead>
                <tr>
                  <th>Agent</th>
                  <th>Module / Ref</th>
                  <th>Status</th>
                  <th>Elapsed</th>
                  <th>Updated</th>
                  <th style={{ textAlign: 'right' }}>Actions</th>
                </tr>
              </thead>
              <tbody>
                {filteredTasks.slice(0, 6).map((t) => {
                  const isActionBusy = !!actionLoading[t.task_id];
                  return (
                    <tr
                      key={t.task_id}
                      className="ai-wf-row clickable"
                      onClick={() => setSelectedTask(t)}
                    >
                      <td className="ai-wf-agent-cell">
                        <strong>{t.agent_name || t.agent_key}</strong>
                      </td>
                      <td>
                        <span className="ai-wf-task-module-tag">{t.business_module}</span>{' '}
                        <span className="ai-wf-task-ref">{t.related_ref || `#${t.task_id.substring(0, 8)}`}</span>
                      </td>
                      <td>
                        <AgentStatusBadge
                          status={t.workforce_status}
                          error={t.safe_error_msg}
                          mockMode={t.mock_mode}
                          providerFailover={t.provider_failover}
                        />
                      </td>
                      <td className="ai-wf-time-cell">{formatDuration(t.duration_ms)}</td>
                      <td className="ai-wf-time-cell">{formatTime(t.updated_at)}</td>
                      <td style={{ textAlign: 'right' }}>
                        <div className="ai-wf-actions-inline" onClick={(e) => e.stopPropagation()}>
                          {t.can_retry && (
                            <button
                              className="ai-wf-btn-subtle retry"
                              onClick={(e) => handleRetryTask(t.task_id, e)}
                              disabled={isActionBusy}
                              title="Retry task"
                            >
                              <RotateCcw size={11} />
                              {actionLoading[t.task_id] === 'retry' ? '...' : 'Retry'}
                            </button>
                          )}
                          {t.requires_approval && (
                            <button
                              className="ai-wf-btn-subtle signoff"
                              onClick={() => navigate('/dashboard/approvals')}
                              title="Sign off"
                            >
                              <CheckCircle2 size={11} /> Sign Off
                            </button>
                          )}
                          {t.can_cancel && (
                            <button
                              className="ai-wf-btn-subtle cancel"
                              onClick={(e) => handleCancelTask(t.task_id, e)}
                              disabled={isActionBusy}
                              title="Cancel task"
                            >
                              <X size={11} />
                            </button>
                          )}
                          <button
                            className="ai-wf-btn-subtle icon-only"
                            onClick={(e) => handleNavigateToRecord(t, e)}
                            title="View Record"
                          >
                            <ArrowUpRight size={13} />
                          </button>
                        </div>
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        )}
      </section>

      {/* ── SAFE LIGHT TASK DETAIL MODAL ── */}
      {selectedTask && (
        <div className="ai-wf-modal-backdrop" onClick={() => setSelectedTask(null)}>
          <div className="ai-wf-modal-box" onClick={(e) => e.stopPropagation()}>
            <div className="ai-wf-modal-top">
              <h4>AI Operation Diagnostics</h4>
              <button
                className="ai-wf-modal-x"
                onClick={() => setSelectedTask(null)}
                aria-label="Close dialog"
              >
                <X size={16} />
              </button>
            </div>

            <div className="ai-wf-modal-content">
              <div className="ai-wf-detail-grid">
                <div className="ai-wf-detail-cell">
                  <span className="ai-wf-label">Task ID</span>
                  <span className="ai-wf-val font-mono">{selectedTask.task_id}</span>
                </div>

                <div className="ai-wf-detail-cell">
                  <span className="ai-wf-label">Status</span>
                  <div>
                    <AgentStatusBadge
                      status={selectedTask.workforce_status}
                      mockMode={selectedTask.mock_mode}
                      providerFailover={selectedTask.provider_failover}
                    />
                  </div>
                </div>

                <div className="ai-wf-detail-cell">
                  <span className="ai-wf-label">Agent</span>
                  <span className="ai-wf-val">{selectedTask.agent_name || selectedTask.agent_key}</span>
                </div>

                <div className="ai-wf-detail-cell">
                  <span className="ai-wf-label">Module / Reference</span>
                  <span className="ai-wf-val">
                    {selectedTask.business_module}: {selectedTask.related_ref || 'None'}
                  </span>
                </div>

                <div className="ai-wf-detail-cell">
                  <span className="ai-wf-label">Execution Duration</span>
                  <span className="ai-wf-val">{formatDuration(selectedTask.duration_ms)}</span>
                </div>

                <div className="ai-wf-detail-cell">
                  <span className="ai-wf-label">Retries Attempted</span>
                  <span className="ai-wf-val">{selectedTask.retry_count} / {selectedTask.max_retries}</span>
                </div>

                <div className="ai-wf-detail-cell">
                  <span className="ai-wf-label">Created At</span>
                  <span className="ai-wf-val">{formatTime(selectedTask.created_at)}</span>
                </div>

                <div className="ai-wf-detail-cell">
                  <span className="ai-wf-label">Updated At</span>
                  <span className="ai-wf-val">{formatTime(selectedTask.updated_at)}</span>
                </div>
              </div>

              {selectedTask.safe_error_msg && (
                <div className="ai-wf-light-error-box">
                  <div className="ai-wf-error-box-head">
                    <AlertTriangle size={13} />
                    <span>Safe Failure Category: {selectedTask.safe_error_category || 'General Execution Fault'}</span>
                  </div>
                  <p>{selectedTask.safe_error_msg}</p>
                </div>
              )}

              {selectedTask.provider_failover && (
                <div className="ai-wf-notice-pill failover" style={{ width: 'fit-content' }}>
                  <Zap size={12} /> Execution succeeded via automated failover provider
                </div>
              )}

              {selectedTask.mock_mode && (
                <div className="ai-wf-notice-pill mock" style={{ width: 'fit-content' }}>
                  <Bot size={12} /> Execution completed in mock development mode
                </div>
              )}
            </div>

            <div className="ai-wf-modal-actions">
              {selectedTask.can_retry && (
                <button
                  className="ai-wf-btn-subtle retry"
                  onClick={() => handleRetryTask(selectedTask.task_id)}
                  disabled={!!actionLoading[selectedTask.task_id]}
                >
                  <RotateCcw size={12} /> Retry Task
                </button>
              )}

              {selectedTask.requires_approval && (
                <button
                  className="ai-wf-btn-subtle signoff"
                  onClick={() => {
                    setSelectedTask(null);
                    navigate('/dashboard/approvals');
                  }}
                >
                  <CheckCircle2 size={12} /> Open Approval Center
                </button>
              )}

              {selectedTask.can_cancel && (
                <button
                  className="ai-wf-btn-subtle cancel"
                  onClick={() => handleCancelTask(selectedTask.task_id)}
                  disabled={!!actionLoading[selectedTask.task_id]}
                >
                  <X size={12} /> Cancel Task
                </button>
              )}

              {selectedTask.related_ref && (
                <button
                  className="ai-wf-btn-subtle"
                  onClick={() => {
                    const t = selectedTask;
                    setSelectedTask(null);
                    handleNavigateToRecord(t);
                  }}
                >
                  <ExternalLink size={12} /> Go to Record
                </button>
              )}

              <button className="ai-wf-btn-light" onClick={() => setSelectedTask(null)}>
                Close
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

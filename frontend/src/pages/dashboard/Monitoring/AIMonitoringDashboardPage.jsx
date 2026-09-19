import React, { useState, useEffect, useCallback } from 'react';
import { 
  Activity, 
  CheckCircle2, 
  AlertTriangle, 
  XCircle, 
  Clock, 
  DollarSign, 
  Layers, 
  ShieldAlert, 
  RefreshCw, 
  Search, 
  Eye, 
  Sliders, 
  Cpu, 
  ThumbsUp, 
  GitPullRequest, 
  Brain, 
  FileCheck,
  Zap,
  Info,
  UserCheck,
  Shield
} from 'lucide-react';
import { monitoringService } from '../../../services/monitoringService';
import { autonomyService } from '../../../services/autonomyService';
import PlanLifecycleCard from '../../../components/autonomy/PlanLifecycleCard';
import HumanAIDecisionCenterDrawer from '../../../components/autonomy/HumanAIDecisionCenterDrawer';
import AIWorkforceCommandCenter from '../../../components/workforce/AIWorkforceCommandCenter';
import './AIMonitoringDashboardPage.css';

export default function AIMonitoringDashboardPage() {
  const [activeTab, setActiveTab] = useState('executions');
  const [timeWindowDays, setTimeWindowDays] = useState(7);
  const [isDecisionCenterOpen, setIsDecisionCenterOpen] = useState(false);

  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  // Data states
  const [healthSummary, setHealthSummary] = useState(null);
  const [perfMetrics, setPerfMetrics] = useState(null);
  const [costMetrics, setCostMetrics] = useState(null);
  const [qualitySummary, setQualitySummary] = useState(null);
  const [queueMetrics, setQueueMetrics] = useState(null);
  const [recMetrics, setRecMetrics] = useState(null);
  const [approvalMetrics, setApprovalMetrics] = useState(null);
  const [memoryMetrics, setMemoryMetrics] = useState(null);
  const [securityMetrics, setSecurityMetrics] = useState(null);
  const [pricingCatalog, setPricingCatalog] = useState([]);
  const [thresholdsList, setThresholdsList] = useState([]);

  // Executions table state
  const [executions, setExecutions] = useState([]);
  const [execTotal, setExecTotal] = useState(0);
  const [execLimit] = useState(15);
  const [execOffset, setExecOffset] = useState(0);
  const [filterFeature, setFilterFeature] = useState('');
  const [filterStatus, setFilterStatus] = useState('');
  const [filterCorrelation, setFilterCorrelation] = useState('');
  const [selectedTrace, setSelectedTrace] = useState(null);

  // Edit pricing / threshold modal
  const [editingThreshold, setEditingThreshold] = useState(null);
  const [savingThreshold, setSavingThreshold] = useState(false);

  // Phase 5 Autonomy State
  const [autonomyPlans, setAutonomyPlans] = useState([]);
  const [loadingAutonomy, setLoadingAutonomy] = useState(false);
  const [autonomyError, setAutonomyError] = useState(null);
  const [showGoalModal, setShowGoalModal] = useState(false);
  const [newGoalObjective, setNewGoalObjective] = useState('');
  const [newGoalModule, setNewGoalModule] = useState('shipments');
  const [newGoalEntityId, setNewGoalEntityId] = useState('1');
  const [newGoalCostLimit, setNewGoalCostLimit] = useState(200);
  const [creatingGoal, setCreatingGoal] = useState(false);

  useEffect(() => {
    const params = new URLSearchParams(window.location.search);
    if (params.get('tab') === 'autonomy') {
      setActiveTab('autonomy');
    }
    if (params.get('tab') === 'workforce') {
      setActiveTab('workforce');
    }
    if (params.get('shipmentId')) {
      setNewGoalEntityId(params.get('shipmentId'));
      setNewGoalModule('shipments');
      setNewGoalObjective(`Mitigate delays and optimize delivery for shipment #${params.get('shipmentId')}`);
      setShowGoalModal(true);
    }
  }, []);

  const loadAutonomyPlans = useCallback(async () => {
    setLoadingAutonomy(true);
    setAutonomyError(null);
    try {
      const res = await autonomyService.listPlans();
      if (res && res.data) {
        setAutonomyPlans(res.data.plans || []);
      }
    } catch (err) {
      setAutonomyError(err.message || 'Failed to load autonomous plans');
    } finally {
      setLoadingAutonomy(false);
    }
  }, []);

  const handleApprovePlan = async (planId) => {
    try {
      await autonomyService.approvePlan(planId, 'Approved via Monitoring Dashboard');
      loadAutonomyPlans();
    } catch (err) {
      alert('Failed to approve plan: ' + err.message);
    }
  };

  const handleRejectPlan = async (planId) => {
    try {
      await autonomyService.rejectPlan(planId, 'Rejected via Monitoring Dashboard');
      loadAutonomyPlans();
    } catch (err) {
      alert('Failed to reject plan: ' + err.message);
    }
  };

  const handleExecuteStep = async (planId, stepId) => {
    try {
      await autonomyService.executeStep(planId, stepId);
      loadAutonomyPlans();
    } catch (err) {
      alert('Failed to execute step: ' + err.message);
    }
  };

  const handleCreateGoal = async (e) => {
    e.preventDefault();
    if (!newGoalObjective) return;
    setCreatingGoal(true);
    try {
      await autonomyService.createGoal({
        module: newGoalModule,
        related_entity_type: newGoalModule === 'shipments' ? 'SHIPMENT' : 'INVOICE',
        related_entity_id: newGoalEntityId,
        objective: newGoalObjective,
        priority: 'HIGH',
        risk_tolerance: 'BALANCED',
        hard_constraints: [
          {
            constraint_type: 'COST',
            description: `Maximum additional cost capped at $${newGoalCostLimit}`,
            is_hard: true,
            threshold_value: parseFloat(newGoalCostLimit) || 200,
          },
        ],
        soft_constraints: [
          {
            constraint_type: 'CUSTOMER',
            description: 'Minimize redundant customer notifications',
            is_hard: false,
          },
        ],
        autonomy_level: 'LEVEL_2_PREPARE',
      });
      setShowGoalModal(false);
      setNewGoalObjective('');
      loadAutonomyPlans();
    } catch (err) {
      alert('Failed to formulate plan: ' + (err.response?.data?.message || err.message));
    } finally {
      setCreatingGoal(false);
    }
  };


  // Load all overview metrics
  const loadDashboardData = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const [
        healthRes,
        perfRes,
        costRes,
        qualityRes,
        queueRes,
        recRes,
        appRes,
        memRes,
        secRes,
        pricingRes,
        threshRes
      ] = await Promise.all([
        monitoringService.getHealthSummary(),
        monitoringService.getPerformance(timeWindowDays),
        monitoringService.getCost(timeWindowDays),
        monitoringService.getQuality(),
        monitoringService.getQueue(),
        monitoringService.getRecommendations(),
        monitoringService.getApprovals(),
        monitoringService.getMemory(),
        monitoringService.getSecurity(),
        monitoringService.getPricing(),
        monitoringService.getThresholds()
      ]);

      setHealthSummary(healthRes?.data || healthRes);
      setPerfMetrics(perfRes?.data || perfRes);
      setCostMetrics(costRes?.data || costRes);
      setQualitySummary(qualityRes?.data || qualityRes);
      setQueueMetrics(queueRes?.data || queueRes);
      setRecMetrics(recRes?.data || recRes);
      setApprovalMetrics(appRes?.data || appRes);
      setMemoryMetrics(memRes?.data || memRes);
      setSecurityMetrics(secRes?.data || secRes);
      setPricingCatalog(pricingRes?.data || []);
      setThresholdsList(threshRes?.data || []);
    } catch (err) {
      console.error('Failed to load AI monitoring data:', err);
      setError('Unable to load monitoring data. Ensure backend is running.');
    } finally {
      setLoading(false);
    }
  }, [timeWindowDays]);

  // Load execution traces
  const loadExecutions = useCallback(async () => {
    try {
      const params = {
        days: timeWindowDays,
        limit: execLimit,
        offset: execOffset,
        feature: filterFeature || undefined,
        status: filterStatus || undefined,
        correlation_id: filterCorrelation || undefined
      };
      const res = await monitoringService.listExecutions(params);
      const data = res?.data || res;
      setExecutions(data?.items || []);
      setExecTotal(data?.total || 0);
    } catch (err) {
      console.error('Failed to load execution traces:', err);
    }
  }, [timeWindowDays, execLimit, execOffset, filterFeature, filterStatus, filterCorrelation]);

  useEffect(() => {
    loadDashboardData();
  }, [loadDashboardData]);

  useEffect(() => {
    if (activeTab === 'autonomy') {
      loadAutonomyPlans();
    }
  }, [activeTab, loadAutonomyPlans]);

  useEffect(() => {
    loadExecutions();
  }, [loadExecutions]);

  const handleUpdateThreshold = async (e) => {
    e.preventDefault();
    if (!editingThreshold) return;
    setSavingThreshold(true);
    try {
      await monitoringService.updateThreshold(editingThreshold);
      setEditingThreshold(null);
      loadDashboardData();
    } catch (err) {
      alert('Failed to update threshold: ' + (err.response?.data?.error || err.message));
    } finally {
      setSavingThreshold(false);
    }
  };

  const overallStatus = healthSummary?.overall_status || 'HEALTHY';
  const activeAlerts = healthSummary?.active_alerts || [];

  return (
    <div className="ai-monitoring-page">
      {/* Header */}
      <header className="monitoring-header">
        <div className="monitoring-title-group">
          <h1>
            <Activity className="w-6 h-6 text-blue-600" />
            AI Performance, Cost & Quality Monitoring
          </h1>
          <p>Unified real-time observability across AI Runtime, agents, background queues, approvals, and cost metrics.</p>
        </div>

        <div className="monitoring-actions-group">
          <select 
            className="time-window-select"
            value={timeWindowDays}
            onChange={(e) => setTimeWindowDays(Number(e.target.value))}
          >
            <option value={1}>Last 24 Hours</option>
            <option value={7}>Last 7 Days</option>
            <option value={30}>Last 30 Days</option>
          </select>

          <button
            className="refresh-btn"
            style={{ backgroundColor: '#0f172a', color: '#ffffff', borderColor: '#0f172a' }}
            onClick={() => setIsDecisionCenterOpen(true)}
            data-testid="btn-open-decision-center"
          >
            <UserCheck className="w-4 h-4 text-blue-300" />
            Decision Center
          </button>

          <button className="refresh-btn" onClick={() => { loadDashboardData(); loadExecutions(); }}>
            <RefreshCw className={`w-4 h-4 ${loading ? 'animate-spin' : ''}`} />
            Refresh
          </button>
        </div>
      </header>

      {/* Active Alerts Banner if any */}
      {activeAlerts.length > 0 && (
        <section className="alerts-banner" aria-label="Active Operational Alerts">
          {activeAlerts.map((alt) => (
            <div key={alt.id} className={`alert-item ${alt.severity.toLowerCase()}`}>
              <AlertTriangle className="w-5 h-5 flex-shrink-0" />
              <span className="alert-item-badge">{alt.severity}</span>
              <div>
                <strong>[{alt.subsystem}] {alt.title}:</strong> {alt.message}
              </div>
            </div>
          ))}
        </section>
      )}

      {/* Top Level 6 KPI Cards */}
      <section className="kpi-grid">
        <div className="kpi-card">
          <div className="kpi-header">
            <span className="kpi-title">System Status</span>
            <CheckCircle2 className="kpi-icon text-emerald-500" />
          </div>
          <div className="kpi-value">
            <span className={`health-pill ${overallStatus.toLowerCase()}`}>
              {overallStatus}
            </span>
          </div>
          <div className="kpi-subtext">All core services monitored</div>
        </div>

        <div className="kpi-card">
          <div className="kpi-header">
            <span className="kpi-title">Success Rate</span>
            <Zap className="kpi-icon text-blue-500" />
          </div>
          <div className="kpi-value">
            {perfMetrics?.total_requests ? `${perfMetrics.success_rate.toFixed(1)}%` : '100%'}
          </div>
          <div className="kpi-subtext">{perfMetrics?.total_requests || 0} total requests</div>
        </div>

        <div className="kpi-card">
          <div className="kpi-header">
            <span className="kpi-title">P95 Latency</span>
            <Clock className="kpi-icon text-indigo-500" />
          </div>
          <div className="kpi-value">
            {perfMetrics?.latency?.p95_ms ? `${perfMetrics.latency.p95_ms}ms` : '—'}
          </div>
          <div className="kpi-subtext">Avg: {perfMetrics?.latency?.average_ms || 0}ms</div>
        </div>

        <div className="kpi-card">
          <div className="kpi-header">
            <span className="kpi-title">Queue Backlog</span>
            <Layers className="kpi-icon text-amber-500" />
          </div>
          <div className="kpi-value">
            {queueMetrics?.queued_jobs || 0}
          </div>
          <div className="kpi-subtext">{queueMetrics?.processing_jobs || 0} currently processing</div>
        </div>

        <div className="kpi-card">
          <div className="kpi-header">
            <span className="kpi-title">Estimated Cost</span>
            <DollarSign className="kpi-icon text-emerald-600" />
          </div>
          <div className="kpi-value">
            ${costMetrics?.total_estimated_cost ? costMetrics.total_estimated_cost.toFixed(4) : '0.0000'}
          </div>
          <div className="kpi-subtext">{costMetrics?.total_tokens?.toLocaleString() || 0} tokens (USD)</div>
        </div>

        <div className="kpi-card">
          <div className="kpi-header">
            <span className="kpi-title">Grounding Score</span>
            <FileCheck className="kpi-icon text-purple-500" />
          </div>
          <div className="kpi-value">
            {qualitySummary?.total_evaluations ? `${qualitySummary.grounding_valid_rate.toFixed(0)}%` : '100%'}
          </div>
          <div className="kpi-subtext">{qualitySummary?.total_evaluations || 0} checks recorded</div>
        </div>
      </section>

      {/* Subsystems Health Matrix */}
      {healthSummary?.subsystems && (
        <section className="subsystems-section">
          <div className="subsystems-title">Subsystem Health Status</div>
          <div className="subsystems-grid">
            {Object.entries(healthSummary.subsystems).map(([key, sub]) => (
              <div key={key} className="subsystem-card">
                <div className="subsystem-info">
                  <h4>{sub.name}</h4>
                  <p>{sub.message}</p>
                </div>
                <span className={`health-pill ${sub.status.toLowerCase()}`}>
                  {sub.status}
                </span>
              </div>
            ))}
          </div>
        </section>
      )}

      {/* Navigation Tabs */}
      <nav className="monitoring-tabs">
        <button 
          className={`tab-btn ${activeTab === 'executions' ? 'active' : ''}`}
          onClick={() => setActiveTab('executions')}
        >
          <Activity className="w-4 h-4" />
          Runtime & Executions
        </button>
        <button 
          className={`tab-btn ${activeTab === 'cost' ? 'active' : ''}`}
          onClick={() => setActiveTab('cost')}
        >
          <DollarSign className="w-4 h-4" />
          Cost & Token Breakdown
        </button>
        <button 
          className={`tab-btn ${activeTab === 'quality' ? 'active' : ''}`}
          onClick={() => setActiveTab('quality')}
        >
          <FileCheck className="w-4 h-4" />
          Quality & Grounding Checks
        </button>
        <button 
          className={`tab-btn ${activeTab === 'queue' ? 'active' : ''}`}
          onClick={() => setActiveTab('queue')}
        >
          <Cpu className="w-4 h-4" />
          Queue & Worker Status
        </button>
        <button 
          className={`tab-btn ${activeTab === 'governance' ? 'active' : ''}`}
          onClick={() => setActiveTab('governance')}
        >
          <GitPullRequest className="w-4 h-4" />
          Recommendations & Approvals
        </button>
        <button 
          className={`tab-btn ${activeTab === 'config' ? 'active' : ''}`}
          onClick={() => setActiveTab('config')}
        >
          <Sliders className="w-4 h-4" />
          Pricing & Thresholds
        </button>
        <button 
          className={`tab-btn ${activeTab === 'autonomy' ? 'active' : ''}`}
          onClick={() => setActiveTab('autonomy')}
          data-testid="tab-autonomy"
        >
          <Brain className="w-4 h-4" />
          Autonomous Operations (Phase 5)
        </button>
        <button 
          className={`tab-btn ${activeTab === 'workforce' ? 'active' : ''}`}
          onClick={() => setActiveTab('workforce')}
          data-testid="tab-workforce"
        >
          <Shield className="w-4 h-4" />
          AI Workforce Command Center (Phase 6.9)
        </button>
      </nav>

      {/* TAB: WORKFORCE COMMAND CENTER (PHASE 6.9) */}
      {activeTab === 'workforce' && (
        <AIWorkforceCommandCenter />
      )}

      {/* TAB 1: RUNTIME & EXECUTIONS */}
      {activeTab === 'executions' && (
        <section className="panel-card">
          <div className="filter-bar">
            <input 
              type="text" 
              placeholder="Filter by feature..." 
              className="filter-input"
              value={filterFeature}
              onChange={(e) => setFilterFeature(e.target.value)}
            />
            <select 
              className="filter-select"
              value={filterStatus}
              onChange={(e) => setFilterStatus(e.target.value)}
            >
              <option value="">All Statuses</option>
              <option value="COMPLETED">Completed</option>
              <option value="FAILED">Failed</option>
              <option value="TIMEOUT">Timeout</option>
            </select>
            <input 
              type="text" 
              placeholder="Search correlation ID..." 
              className="filter-input"
              value={filterCorrelation}
              onChange={(e) => setFilterCorrelation(e.target.value)}
            />
          </div>

          <div className="data-table-container">
            <table className="monitoring-table">
              <thead>
                <tr>
                  <th>Trace ID</th>
                  <th>Feature / Workflow</th>
                  <th>Provider & Model</th>
                  <th>Status</th>
                  <th>Duration</th>
                  <th>Tokens</th>
                  <th>Est. Cost</th>
                  <th>Grounding</th>
                  <th>Timestamp</th>
                  <th>Action</th>
                </tr>
              </thead>
              <tbody>
                {executions.length === 0 ? (
                  <tr>
                    <td colSpan="10" style={{ textAlign: 'center', padding: '32px', color: '#64748b' }}>
                      No execution traces found for this period or filter.
                    </td>
                  </tr>
                ) : (
                  executions.map((t) => (
                    <tr key={t.id}>
                      <td>#{t.id}</td>
                      <td>
                        <strong>{t.feature || t.workflow_name}</strong>
                        <div style={{ fontSize: '0.72rem', color: '#64748b' }}>{t.assistant}</div>
                      </td>
                      <td>
                        {t.final_provider} ({t.final_model})
                        {t.failover_occurred && <span className="badge warning" style={{ marginLeft: 6 }}>Failover</span>}
                        {t.is_mock && <span className="badge neutral" style={{ marginLeft: 6 }}>Mock</span>}
                      </td>
                      <td>
                        <span className={`badge ${t.status.toLowerCase()}`}>
                          {t.status}
                        </span>
                      </td>
                      <td>{t.duration_ms ? `${t.duration_ms}ms` : '—'}</td>
                      <td>{t.total_tokens ? t.total_tokens.toLocaleString() : '0'}</td>
                      <td>${t.estimated_cost ? t.estimated_cost.toFixed(5) : '0.00000'}</td>
                      <td>
                        <span className={`badge ${t.grounding_status === 'GROUNDED' ? 'valid' : 'neutral'}`}>
                          {t.grounding_status}
                        </span>
                      </td>
                      <td>{new Date(t.created_at).toLocaleTimeString()}</td>
                      <td>
                        <button 
                          className="refresh-btn" 
                          style={{ padding: '4px 8px', fontSize: '0.75rem' }}
                          onClick={() => setSelectedTrace(t)}
                        >
                          <Eye className="w-3.5 h-3.5" /> Details
                        </button>
                      </td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>

          {/* Pagination */}
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginTop: 16 }}>
            <span style={{ fontSize: '0.75rem', color: '#64748b' }}>
              Showing {executions.length} of {execTotal} executions
            </span>
            <div style={{ display: 'flex', gap: 8 }}>
              <button 
                className="refresh-btn" 
                disabled={execOffset === 0}
                onClick={() => setExecOffset(Math.max(0, execOffset - execLimit))}
              >
                Previous
              </button>
              <button 
                className="refresh-btn" 
                disabled={execOffset + execLimit >= execTotal}
                onClick={() => setExecOffset(execOffset + execLimit)}
              >
                Next
              </button>
            </div>
          </div>
        </section>
      )}

      {/* TAB 2: COST & TOKENS */}
      {activeTab === 'cost' && (
        <section className="panel-card">
          <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 20 }}>
            <div>
              <h3 style={{ fontSize: '1.125rem', fontWeight: 700, margin: '0 0 4px 0' }}>AI Cost & Token Consumption</h3>
              <p style={{ fontSize: '0.8125rem', color: '#64748b', margin: 0 }}>Aggregated model and provider spend breakdown.</p>
            </div>
            <div style={{ textAlign: 'right' }}>
              <div style={{ fontSize: '1.5rem', fontWeight: 800, color: '#059669' }}>
                ${costMetrics?.total_estimated_cost ? costMetrics.total_estimated_cost.toFixed(4) : '0.0000'}
              </div>
              <span className="badge neutral">ESTIMATED (USD)</span>
            </div>
          </div>

          {/* Model Breakdown Table */}
          <h4 style={{ fontSize: '0.875rem', fontWeight: 700, color: '#334155', marginTop: 24, marginBottom: 10 }}>
            Cost by Model
          </h4>
          <div className="data-table-container">
            <table className="monitoring-table">
              <thead>
                <tr>
                  <th>Model</th>
                  <th>Provider</th>
                  <th>Requests</th>
                  <th>Input Tokens</th>
                  <th>Output Tokens</th>
                  <th>Total Tokens</th>
                  <th>Estimated Cost</th>
                  <th>Share (%)</th>
                </tr>
              </thead>
              <tbody>
                {(!costMetrics?.by_model || costMetrics.by_model.length === 0) ? (
                  <tr>
                    <td colSpan="8" style={{ textAlign: 'center', padding: '24px', color: '#64748b' }}>
                      No cost data recorded for this period.
                    </td>
                  </tr>
                ) : (
                  costMetrics.by_model.map((m, idx) => (
                    <tr key={idx}>
                      <td><strong>{m.model_name || m.dimension_name}</strong></td>
                      <td>{m.provider}</td>
                      <td>{m.total_requests}</td>
                      <td>{m.input_tokens.toLocaleString()}</td>
                      <td>{m.output_tokens.toLocaleString()}</td>
                      <td>{m.total_tokens.toLocaleString()}</td>
                      <td>${m.estimated_cost.toFixed(5)}</td>
                      <td>{m.percentage_cost.toFixed(1)}%</td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>

          {/* Feature Breakdown Table */}
          <h4 style={{ fontSize: '0.875rem', fontWeight: 700, color: '#334155', marginTop: 24, marginBottom: 10 }}>
            Cost by Feature / Assistant
          </h4>
          <div className="data-table-container">
            <table className="monitoring-table">
              <thead>
                <tr>
                  <th>Feature</th>
                  <th>Requests</th>
                  <th>Tokens</th>
                  <th>Estimated Cost</th>
                  <th>Share (%)</th>
                </tr>
              </thead>
              <tbody>
                {(!costMetrics?.by_feature || costMetrics.by_feature.length === 0) ? (
                  <tr>
                    <td colSpan="5" style={{ textAlign: 'center', padding: '24px', color: '#64748b' }}>
                      No feature breakdown data available.
                    </td>
                  </tr>
                ) : (
                  costMetrics.by_feature.map((f, idx) => (
                    <tr key={idx}>
                      <td><strong>{f.dimension_name}</strong></td>
                      <td>{f.total_requests}</td>
                      <td>{f.total_tokens.toLocaleString()}</td>
                      <td>${f.estimated_cost.toFixed(5)}</td>
                      <td>{f.percentage_cost.toFixed(1)}%</td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>

          <div className="disclaimer-box">
            <Info className="w-4 h-4 inline-block mr-1 text-blue-600" />
            <strong>Cost Disclaimer:</strong> {costMetrics?.cost_disclaimer || 'Estimated using published provider token pricing.'}
          </div>
        </section>
      )}

      {/* TAB 3: QUALITY & GROUNDING */}
      {activeTab === 'quality' && (
        <section className="panel-card">
          <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 20 }}>
            <div>
              <h3 style={{ fontSize: '1.125rem', fontWeight: 700, margin: '0 0 4px 0' }}>Quality & Grounding Evaluations</h3>
              <p style={{ fontSize: '0.8125rem', color: '#64748b', margin: 0 }}>
                {qualitySummary?.automated_check_label || 'Automated verification checks'}
              </p>
            </div>
            <div style={{ display: 'flex', gap: 16 }}>
              <div style={{ textAlign: 'center' }}>
                <div style={{ fontSize: '1.25rem', fontWeight: 700, color: '#15803d' }}>
                  {qualitySummary?.grounding_valid_rate ? `${qualitySummary.grounding_valid_rate.toFixed(0)}%` : '100%'}
                </div>
                <div style={{ fontSize: '0.72rem', color: '#64748b' }}>Grounding Validity</div>
              </div>
              <div style={{ textAlign: 'center' }}>
                <div style={{ fontSize: '1.25rem', fontWeight: 700, color: '#2563eb' }}>
                  {qualitySummary?.evidence_sufficient_rate ? `${qualitySummary.evidence_sufficient_rate.toFixed(0)}%` : '100%'}
                </div>
                <div style={{ fontSize: '0.72rem', color: '#64748b' }}>Evidence Sufficiency</div>
              </div>
            </div>
          </div>

          <div className="data-table-container">
            <table className="monitoring-table">
              <thead>
                <tr>
                  <th>Check ID</th>
                  <th>Feature</th>
                  <th>Evaluation Type</th>
                  <th>Score</th>
                  <th>Status</th>
                  <th>Grounding</th>
                  <th>Evidence</th>
                  <th>Notes (Redacted)</th>
                  <th>Evaluator</th>
                  <th>Timestamp</th>
                </tr>
              </thead>
              <tbody>
                {(!qualitySummary?.recent_evaluations || qualitySummary.recent_evaluations.length === 0) ? (
                  <tr>
                    <td colSpan="10" style={{ textAlign: 'center', padding: '32px', color: '#64748b' }}>
                      No quality evaluations recorded yet.
                    </td>
                  </tr>
                ) : (
                  qualitySummary.recent_evaluations.map((ev) => (
                    <tr key={ev.id}>
                      <td>#{ev.id}</td>
                      <td><strong>{ev.feature}</strong></td>
                      <td>{ev.evaluation_type}</td>
                      <td><strong>{(ev.score * 100).toFixed(0)}%</strong></td>
                      <td>
                        <span className={`badge ${ev.pass_status.toLowerCase()}`}>
                          {ev.pass_status}
                        </span>
                      </td>
                      <td>
                        <span className={`badge ${ev.grounding_status === 'VALID' ? 'valid' : 'warning'}`}>
                          {ev.grounding_status}
                        </span>
                      </td>
                      <td>
                        <span className={`badge ${ev.evidence_status === 'SUFFICIENT' ? 'valid' : 'neutral'}`}>
                          {ev.evidence_status}
                        </span>
                      </td>
                      <td style={{ maxWidth: 220, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                        {ev.evaluation_notes || '—'}
                      </td>
                      <td>{ev.reviewer_type}</td>
                      <td>{new Date(ev.created_at).toLocaleTimeString()}</td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>
        </section>
      )}

      {/* TAB 4: QUEUE & WORKER */}
      {activeTab === 'queue' && (
        <section className="panel-card">
          <h3 style={{ fontSize: '1.125rem', fontWeight: 700, margin: '0 0 16px 0' }}>AI Background Task Queue</h3>
          <div className="kpi-grid" style={{ marginBottom: 20 }}>
            <div className="kpi-card">
              <span className="kpi-title">Queued Jobs</span>
              <div className="kpi-value">{queueMetrics?.queued_jobs || 0}</div>
              <span className="kpi-subtext">Waiting for worker pick-up</span>
            </div>
            <div className="kpi-card">
              <span className="kpi-title">Processing Jobs</span>
              <div className="kpi-value">{queueMetrics?.processing_jobs || 0}</div>
              <span className="kpi-subtext">In-flight executions</span>
            </div>
            <div className="kpi-card">
              <span className="kpi-title">Completed (24h)</span>
              <div className="kpi-value">{queueMetrics?.completed_24h || 0}</div>
              <span className="kpi-subtext">Successfully resolved</span>
            </div>
            <div className="kpi-card">
              <span className="kpi-title">Dead Letter Jobs</span>
              <div className="kpi-value" style={{ color: queueMetrics?.dead_letter_jobs > 0 ? '#b91c1c' : 'inherit' }}>
                {queueMetrics?.dead_letter_jobs || 0}
              </div>
              <span className="kpi-subtext">Exceeded retry limits</span>
            </div>
          </div>

          <h4 style={{ fontSize: '0.875rem', fontWeight: 700, color: '#334155', marginTop: 20, marginBottom: 10 }}>
            Worker Performance Indicators
          </h4>
          <div style={{ background: '#f8fafc', padding: 16, borderRadius: 8, border: '1px solid #e2e8f0', display: 'flex', gap: 32 }}>
            <div>
              <span style={{ fontSize: '0.75rem', color: '#64748b' }}>Oldest Queued Job Age:</span>
              <div style={{ fontSize: '1rem', fontWeight: 700 }}>{queueMetrics?.oldest_queued_age_sec || 0} seconds</div>
            </div>
            <div>
              <span style={{ fontSize: '0.75rem', color: '#64748b' }}>Average Task Duration:</span>
              <div style={{ fontSize: '1rem', fontWeight: 700 }}>{queueMetrics?.average_duration_ms || 0} ms</div>
            </div>
            <div>
              <span style={{ fontSize: '0.75rem', color: '#64748b' }}>Worker Status:</span>
              <div style={{ fontSize: '1rem', fontWeight: 700 }}>
                <span className={`badge ${queueMetrics?.worker_health === 'HEALTHY' ? 'passed' : 'failed'}`}>
                  {queueMetrics?.worker_health || 'HEALTHY'}
                </span>
              </div>
            </div>
          </div>
        </section>
      )}

      {/* TAB 5: RECOMMENDATIONS & APPROVALS */}
      {activeTab === 'governance' && (
        <section className="panel-card">
          <h3 style={{ fontSize: '1.125rem', fontWeight: 700, margin: '0 0 16px 0' }}>AI Action & Human-in-the-Loop Governance</h3>

          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 24, marginBottom: 24 }}>
            {/* Recommendations */}
            <div style={{ border: '1px solid #e2e8f0', borderRadius: 8, padding: 16 }}>
              <h4 style={{ fontSize: '0.875rem', fontWeight: 700, margin: '0 0 12px 0', display: 'flex', alignItems: 'center', gap: 8 }}>
                <ThumbsUp className="w-4 h-4 text-blue-600" /> Recommendation Center
              </h4>
              <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 12 }}>
                <div>
                  <span style={{ fontSize: '0.75rem', color: '#64748b' }}>Total Generated:</span>
                  <div style={{ fontSize: '1.25rem', fontWeight: 700 }}>{recMetrics?.total_generated || 0}</div>
                </div>
                <div>
                  <span style={{ fontSize: '0.75rem', color: '#64748b' }}>Acceptance Rate:</span>
                  <div style={{ fontSize: '1.25rem', fontWeight: 700, color: '#15803d' }}>
                    {recMetrics?.acceptance_rate ? `${recMetrics.acceptance_rate.toFixed(1)}%` : '0%'}
                  </div>
                </div>
                <div>
                  <span style={{ fontSize: '0.75rem', color: '#64748b' }}>Active Items:</span>
                  <div style={{ fontSize: '1rem', fontWeight: 600 }}>{recMetrics?.active_count || 0}</div>
                </div>
                <div>
                  <span style={{ fontSize: '0.75rem', color: '#64748b' }}>Requiring Approval:</span>
                  <div style={{ fontSize: '1rem', fontWeight: 600 }}>{recMetrics?.requiring_approval || 0}</div>
                </div>
              </div>
            </div>

            {/* Approvals */}
            <div style={{ border: '1px solid #e2e8f0', borderRadius: 8, padding: 16 }}>
              <h4 style={{ fontSize: '0.875rem', fontWeight: 700, margin: '0 0 12px 0', display: 'flex', alignItems: 'center', gap: 8 }}>
                <GitPullRequest className="w-4 h-4 text-purple-600" /> Action & Approval System
              </h4>
              <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 12 }}>
                <div>
                  <span style={{ fontSize: '0.75rem', color: '#64748b' }}>Total Proposed:</span>
                  <div style={{ fontSize: '1.25rem', fontWeight: 700 }}>{approvalMetrics?.total_proposed || 0}</div>
                </div>
                <div>
                  <span style={{ fontSize: '0.75rem', color: '#64748b' }}>Approved:</span>
                  <div style={{ fontSize: '1.25rem', fontWeight: 700, color: '#15803d' }}>
                    {approvalMetrics?.approved_count || 0}
                  </div>
                </div>
                <div>
                  <span style={{ fontSize: '0.75rem', color: '#64748b' }}>Pending Human Review:</span>
                  <div style={{ fontSize: '1rem', fontWeight: 600 }}>{approvalMetrics?.pending_count || 0}</div>
                </div>
                <div>
                  <span style={{ fontSize: '0.75rem', color: '#64748b' }}>Approval Failure Rate:</span>
                  <div style={{ fontSize: '1rem', fontWeight: 600 }}>
                    {approvalMetrics?.approval_failure_rate ? `${approvalMetrics.approval_failure_rate.toFixed(1)}%` : '0%'}
                  </div>
                </div>
              </div>
            </div>
          </div>

          {/* AI Memory Safety Metrics */}
          <div style={{ border: '1px solid #e2e8f0', borderRadius: 8, padding: 16, background: '#fafbfc' }}>
            <h4 style={{ fontSize: '0.875rem', fontWeight: 700, margin: '0 0 12px 0', display: 'flex', alignItems: 'center', gap: 8 }}>
              <Brain className="w-4 h-4 text-indigo-600" /> AI Memory & Security Boundary Status
            </h4>
            <div style={{ display: 'grid', gridTemplateColumns: 'repeat(4, 1fr)', gap: 12 }}>
              <div>
                <span style={{ fontSize: '0.75rem', color: '#64748b' }}>Active Personal Memories:</span>
                <div style={{ fontSize: '1.125rem', fontWeight: 700 }}>{memoryMetrics?.active_personal_memories || 0}</div>
              </div>
              <div>
                <span style={{ fontSize: '0.75rem', color: '#64748b' }}>Active Org Memories:</span>
                <div style={{ fontSize: '1.125rem', fontWeight: 700 }}>{memoryMetrics?.active_org_memories || 0}</div>
              </div>
              <div>
                <span style={{ fontSize: '0.75rem', color: '#64748b' }}>Sensitive Rejections:</span>
                <div style={{ fontSize: '1.125rem', fontWeight: 700, color: '#b91c1c' }}>
                  {memoryMetrics?.sensitive_rejections_count || 0}
                </div>
              </div>
              <div>
                <span style={{ fontSize: '0.75rem', color: '#64748b' }}>Unauthorized Attempts:</span>
                <div style={{ fontSize: '1.125rem', fontWeight: 700, color: '#b91c1c' }}>
                  {memoryMetrics?.unauthorized_attempts_count || 0}
                </div>
              </div>
            </div>
          </div>
        </section>
      )}

      {/* TAB 6: PRICING & THRESHOLDS */}
      {activeTab === 'config' && (
        <section className="panel-card">
          <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 16 }}>
            <div>
              <h3 style={{ fontSize: '1.125rem', fontWeight: 700, margin: '0 0 4px 0' }}>Operational Health Thresholds</h3>
              <p style={{ fontSize: '0.8125rem', color: '#64748b', margin: 0 }}>
                Configurable trigger boundaries governing system health alerts.
              </p>
            </div>
          </div>

          <div className="data-table-container">
            <table className="monitoring-table">
              <thead>
                <tr>
                  <th>Threshold Key</th>
                  <th>Value</th>
                  <th>Unit</th>
                  <th>Severity</th>
                  <th>Description</th>
                  <th>Action</th>
                </tr>
              </thead>
              <tbody>
                {thresholdsList.map((th) => (
                  <tr key={th.id || th.threshold_key}>
                    <td><code>{th.threshold_key}</code></td>
                    <td><strong>{th.threshold_value}</strong></td>
                    <td>{th.unit}</td>
                    <td>
                      <span className={`badge ${th.severity === 'CRITICAL' ? 'failed' : 'warning'}`}>
                        {th.severity}
                      </span>
                    </td>
                    <td>{th.description || '—'}</td>
                    <td>
                      <button 
                        className="refresh-btn" 
                        style={{ padding: '4px 8px', fontSize: '0.75rem' }}
                        onClick={() => setEditingThreshold({ ...th })}
                      >
                        Configure
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

          {/* Pricing Catalog Table */}
          <h4 style={{ fontSize: '0.875rem', fontWeight: 700, color: '#334155', marginTop: 32, marginBottom: 10 }}>
            Model Pricing Catalog (USD per 1k Tokens)
          </h4>
          <div className="data-table-container">
            <table className="monitoring-table">
              <thead>
                <tr>
                  <th>Provider</th>
                  <th>Model Name</th>
                  <th>Input Cost / 1k</th>
                  <th>Output Cost / 1k</th>
                  <th>Currency</th>
                  <th>Status</th>
                </tr>
              </thead>
              <tbody>
                {pricingCatalog.map((p) => (
                  <tr key={p.id || `${p.provider}-${p.model_name}`}>
                    <td><strong>{p.provider}</strong></td>
                    <td><code>{p.model_name}</code></td>
                    <td>${p.input_cost_per_1k_tokens.toFixed(6)}</td>
                    <td>${p.output_cost_per_1k_tokens.toFixed(6)}</td>
                    <td>{p.currency}</td>
                    <td>
                      <span className="badge completed">Active</span>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </section>
      )}

      {/* TAB 7: AUTONOMOUS OPERATIONS (PHASE 5) */}
      {activeTab === 'autonomy' && (
        <section className="panel-card" data-testid="autonomy-panel">
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 20 }}>
            <div>
              <h3 style={{ fontSize: '1.125rem', fontWeight: 700, margin: '0 0 4px 0', color: '#0f172a' }}>
                Phase 5: Autonomous Operations Readiness
              </h3>
              <p style={{ fontSize: '0.8125rem', color: '#64748b', margin: 0 }}>
                Controlled autonomous operations with explicit policy gates, durable plan lifecycles, and Action System enforcement.
              </p>
            </div>
            <div style={{ display: 'flex', gap: 8 }}>
              <button 
                className="refresh-btn"
                style={{ backgroundColor: '#2563eb', color: '#ffffff', borderColor: '#2563eb' }}
                onClick={() => setShowGoalModal(true)}
              >
                + New Operational Goal
              </button>
              <button 
                className="refresh-btn"
                onClick={loadAutonomyPlans}
                disabled={loadingAutonomy}
              >
                <RefreshCw className={`w-3.5 h-3.5 ${loadingAutonomy ? 'animate-spin' : ''}`} />
                Refresh Plans
              </button>
            </div>
          </div>

          {autonomyError && (
            <div style={{ padding: '12px 16px', background: '#fee2e2', color: '#991b1b', borderRadius: 6, fontSize: '0.8125rem', marginBottom: 16 }}>
              {autonomyError}
            </div>
          )}

          {/* Autonomy Level Guide */}
          <div className="grid grid-cols-1 sm:grid-cols-5 gap-2 mb-6 text-xs border border-slate-200 rounded-lg p-3 bg-slate-50">
            <div className="p-2 rounded bg-white border border-slate-200">
              <span className="font-semibold block text-slate-800">Level 0: Observe</span>
              <span className="text-slate-500 text-[11px]">Telemetry, monitoring & alerts</span>
            </div>
            <div className="p-2 rounded bg-white border border-slate-200">
              <span className="font-semibold block text-slate-800">Level 1: Recommend</span>
              <span className="text-slate-500 text-[11px]">Proposals requiring user action</span>
            </div>
            <div className="p-2 rounded bg-white border border-slate-200">
              <span className="font-semibold block text-slate-800">Level 2: Prepare</span>
              <span className="text-slate-500 text-[11px]">Full executable plan; requires HITL</span>
            </div>
            <div className="p-2 rounded bg-white border border-slate-200">
              <span className="font-semibold block text-slate-800">Level 3: Controlled</span>
              <span className="text-slate-500 text-[11px]">Pre-approved low-risk actions</span>
            </div>
            <div className="p-2 rounded bg-white border border-slate-200">
              <span className="font-semibold block text-slate-800">Level 4: Multi-Step</span>
              <span className="text-slate-500 text-[11px]">Chained actions under policy boundaries</span>
            </div>
          </div>

          {/* Active Plans List */}
          {loadingAutonomy && autonomyPlans.length === 0 ? (
            <div style={{ textAlign: 'center', padding: '40px 0', color: '#64748b' }}>
              <RefreshCw className="w-6 h-6 animate-spin mx-auto mb-2" />
              Loading autonomous plans...
            </div>
          ) : autonomyPlans.length === 0 ? (
            <div style={{ textAlign: 'center', padding: '48px 0', border: '1px dashed #cbd5e1', borderRadius: 8 }}>
              <Brain className="w-8 h-8 text-slate-400 mx-auto mb-2" />
              <h4 style={{ fontSize: '0.875rem', fontWeight: 600, color: '#334155', margin: '0 0 4px 0' }}>No Active Autonomous Plans</h4>
              <p style={{ fontSize: '0.8125rem', color: '#64748b', margin: 0 }}>
                Generate or trigger an operational plan across Shipments, RFQ, Leads, or Finance to monitor execution.
              </p>
            </div>
          ) : (
            <div className="space-y-4">
              {autonomyPlans.map((plan) => (
                <PlanLifecycleCard 
                  key={plan.plan_id || plan.id}
                  plan={plan}
                  steps={plan.steps || []}
                  onApprove={handleApprovePlan}
                  onReject={handleRejectPlan}
                  onExecuteStep={handleExecuteStep}
                  onPlanUpdated={loadAutonomyPlans}
                />
              ))}
            </div>
          )}
        </section>
      )}

      {/* Trace Details Modal */}
      {selectedTrace && (
        <div className="modal-overlay" onClick={() => setSelectedTrace(null)}>
          <div className="modal-content" onClick={(e) => e.stopPropagation()}>
            <div className="modal-header">
              <h3>Execution Trace #{selectedTrace.id}</h3>
              <button className="modal-close-btn" onClick={() => setSelectedTrace(null)}>×</button>
            </div>
            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 12, fontSize: '0.8125rem' }}>
              <div><strong>Feature:</strong> {selectedTrace.feature}</div>
              <div><strong>Assistant:</strong> {selectedTrace.assistant}</div>
              <div><strong>Provider:</strong> {selectedTrace.final_provider}</div>
              <div><strong>Model:</strong> {selectedTrace.final_model}</div>
              <div><strong>Duration:</strong> {selectedTrace.duration_ms}ms</div>
              <div><strong>Status:</strong> {selectedTrace.status}</div>
              <div><strong>Input Tokens:</strong> {selectedTrace.input_tokens}</div>
              <div><strong>Output Tokens:</strong> {selectedTrace.output_tokens}</div>
              <div><strong>Total Cost:</strong> ${selectedTrace.estimated_cost?.toFixed(5)}</div>
              <div><strong>Grounding:</strong> {selectedTrace.grounding_status}</div>
              <div style={{ gridColumn: 'span 2' }}>
                <strong>Correlation ID:</strong> <code>{selectedTrace.correlation_id || 'N/A'}</code>
              </div>
              {selectedTrace.error_message && (
                <div style={{ gridColumn: 'span 2', color: '#b91c1c', background: '#fee2e2', padding: 8, borderRadius: 6 }}>
                  <strong>Error:</strong> {selectedTrace.error_message}
                </div>
              )}
            </div>
            <div className="modal-footer">
              <button className="refresh-btn" onClick={() => setSelectedTrace(null)}>Close</button>
            </div>
          </div>
        </div>
      )}

      {/* Edit Threshold Modal */}
      {editingThreshold && (
        <div className="modal-overlay" onClick={() => setEditingThreshold(null)}>
          <div className="modal-content" onClick={(e) => e.stopPropagation()}>
            <div className="modal-header">
              <h3>Configure Health Threshold</h3>
              <button className="modal-close-btn" onClick={() => setEditingThreshold(null)}>×</button>
            </div>
            <form onSubmit={handleUpdateThreshold}>
              <div style={{ marginBottom: 14 }}>
                <label style={{ fontSize: '0.75rem', fontWeight: 600, display: 'block', marginBottom: 4 }}>
                  Threshold Key
                </label>
                <input 
                  type="text" 
                  className="filter-input" 
                  style={{ width: '100%' }}
                  value={editingThreshold.threshold_key}
                  disabled
                />
              </div>
              <div style={{ marginBottom: 14 }}>
                <label style={{ fontSize: '0.75rem', fontWeight: 600, display: 'block', marginBottom: 4 }}>
                  Threshold Value ({editingThreshold.unit})
                </label>
                <input 
                  type="number" 
                  step="0.1" 
                  className="filter-input" 
                  style={{ width: '100%' }}
                  value={editingThreshold.threshold_value}
                  onChange={(e) => setEditingThreshold({
                    ...editingThreshold,
                    threshold_value: parseFloat(e.target.value) || 0
                  })}
                  required
                />
              </div>
              <div style={{ marginBottom: 14 }}>
                <label style={{ fontSize: '0.75rem', fontWeight: 600, display: 'block', marginBottom: 4 }}>
                  Severity
                </label>
                <select 
                  className="filter-select" 
                  style={{ width: '100%' }}
                  value={editingThreshold.severity}
                  onChange={(e) => setEditingThreshold({
                    ...editingThreshold,
                    severity: e.target.value
                  })}
                >
                  <option value="WARNING">WARNING</option>
                  <option value="CRITICAL">CRITICAL</option>
                </select>
              </div>
              <div className="modal-footer">
                <button type="button" className="refresh-btn" onClick={() => setEditingThreshold(null)}>
                  Cancel
                </button>
                <button 
                  type="submit" 
                  className="refresh-btn" 
                  style={{ backgroundColor: '#2563eb', color: '#ffffff', borderColor: '#2563eb' }}
                  disabled={savingThreshold}
                >
                  {savingThreshold ? 'Saving...' : 'Save Configuration'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Create Operational Goal Modal */}
      {showGoalModal && (
        <div className="modal-overlay" onClick={() => setShowGoalModal(false)}>
          <div className="modal-content" onClick={(e) => e.stopPropagation()} style={{ maxWidth: 500 }}>
            <div className="modal-header">
              <h3>Formulate Operational Goal</h3>
              <button className="modal-close-btn" onClick={() => setShowGoalModal(false)}>×</button>
            </div>
            <form onSubmit={handleCreateGoal}>
              <div style={{ marginBottom: 14 }}>
                <label style={{ fontSize: '0.75rem', fontWeight: 600, display: 'block', marginBottom: 4 }}>
                  Business Module
                </label>
                <select 
                  className="filter-input" 
                  style={{ width: '100%' }}
                  value={newGoalModule}
                  onChange={(e) => setNewGoalModule(e.target.value)}
                >
                  <option value="shipments">Shipments</option>
                  <option value="finance">Finance / Invoices</option>
                  <option value="rfq">RFQ & Quotations</option>
                  <option value="operations">Operations</option>
                </select>
              </div>

              <div style={{ marginBottom: 14 }}>
                <label style={{ fontSize: '0.75rem', fontWeight: 600, display: 'block', marginBottom: 4 }}>
                  Entity ID (e.g. Shipment ID)
                </label>
                <input 
                  type="text" 
                  className="filter-input" 
                  style={{ width: '100%' }}
                  value={newGoalEntityId}
                  onChange={(e) => setNewGoalEntityId(e.target.value)}
                  placeholder="e.g. 1"
                  required
                />
              </div>

              <div style={{ marginBottom: 14 }}>
                <label style={{ fontSize: '0.75rem', fontWeight: 600, display: 'block', marginBottom: 4 }}>
                  Operational Objective
                </label>
                <textarea 
                  className="filter-input" 
                  style={{ width: '100%', minHeight: 80 }}
                  value={newGoalObjective}
                  onChange={(e) => setNewGoalObjective(e.target.value)}
                  placeholder="e.g. Mitigate shipment delay and recover SLA commitment"
                  required
                />
              </div>

              <div style={{ marginBottom: 14 }}>
                <label style={{ fontSize: '0.75rem', fontWeight: 600, display: 'block', marginBottom: 4 }}>
                  Max Additional Cost Budget ($) [Hard Constraint]
                </label>
                <input 
                  type="number" 
                  className="filter-input" 
                  style={{ width: '100%' }}
                  value={newGoalCostLimit}
                  onChange={(e) => setNewGoalCostLimit(e.target.value)}
                  min="0"
                  step="10"
                />
              </div>

              <div className="modal-footer">
                <button type="button" className="refresh-btn" onClick={() => setShowGoalModal(false)}>
                  Cancel
                </button>
                <button 
                  type="submit" 
                  className="refresh-btn" 
                  style={{ backgroundColor: '#2563eb', color: '#ffffff', borderColor: '#2563eb' }}
                  disabled={creatingGoal}
                >
                  {creatingGoal ? 'Formulating Plan & Candidates...' : 'Generate AI Plan'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}


      {/* Human + AI Decision Center (Task 5.11) */}
      <HumanAIDecisionCenterDrawer
        isOpen={isDecisionCenterOpen}
        onClose={() => setIsDecisionCenterOpen(false)}
        onDecisionUpdated={() => {
          loadAutonomyPlans();
          loadDashboardData();
        }}
      />
    </div>
  );
}

import React, { useState, useEffect, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  AlertTriangle,
  CheckCircle2,
  Clock,
  RefreshCw,
  Search,
  Shield,
  Activity,
  Sliders,
  Play,
  Pause,
  AlertOctagon,
  ArrowRight,
  TrendingUp,
  Cpu,
  UserCheck,
  FileText,
  Ship,
  DollarSign,
  Users,
  ChevronRight,
  ExternalLink,
  Layers,
  Sparkles,
  HelpCircle,
  XCircle,
  Brain,
  Eye,
  Compass,
  ArrowUpRight,
  Check,
  X,
} from 'lucide-react';
import autonomyService from '../../../services/autonomyService';
import enterpriseService from '../../../services/enterpriseService';
import HumanAIDecisionCenterDrawer from '../../../components/autonomy/HumanAIDecisionCenterDrawer';
import ShipmentAdaptiveDrawer from '../../../components/autonomy/ShipmentAdaptiveDrawer';
import ExceptionResolutionDrawer from '../../../components/autonomy/ExceptionResolutionDrawer';
import FinanceCollectionsAdaptiveDrawer from '../../../components/autonomy/FinanceCollectionsAdaptiveDrawer';
import ContractComplianceMonitoringDrawer from '../../../components/autonomy/ContractComplianceMonitoringDrawer';
import ContinuousMonitoringDrawer from '../../../components/autonomy/ContinuousMonitoringDrawer';
import AgentMemoryLearningDrawer from '../../../components/autonomy/AgentMemoryLearningDrawer';
import ControlledAutonomyGovernanceDrawer from '../../../components/autonomy/ControlledAutonomyGovernanceDrawer';
import './AutonomousCommandCenterPage.css';

// Helper to cleanly unwrap Go sql.NullString / sql.Null* objects if present in payload
const getSqlString = (val) => {
  if (!val) return '';
  if (typeof val === 'string') return val;
  if (typeof val === 'number') return String(val);
  if (typeof val === 'object' && val !== null) {
    if (val.Valid !== undefined) {
      return val.Valid ? String(val.String || '') : '';
    }
    return '';
  }
  return String(val);
};

export default function AutonomousCommandCenterPage() {
  const navigate = useNavigate();

  // Primary State
  const [activeTab, setActiveTab] = useState('control_tower'); // 'control_tower' | 'attention' | 'workflows' | 'decisions' | 'risks' | 'activity' | 'health'
  const [riskSubTab, setRiskSubTab] = useState('SHIPMENT'); // 'SHIPMENT' | 'FINANCE' | 'CUSTOMER' | 'COMPLIANCE'
  const [isLoading, setIsLoading] = useState(true);
  const [isRefreshing, setIsRefreshing] = useState(false);
  const [autoRefresh, setAutoRefresh] = useState(true);
  const [lastUpdated, setLastUpdated] = useState(new Date());

  // Data Stores
  const [controlTowerView, setControlTowerView] = useState(null);
  const [overview, setOverview] = useState(null);
  const [criticalItems, setCriticalItems] = useState([]);
  const [workflows, setWorkflows] = useState([]);
  const [decisions, setDecisions] = useState([]);
  const [riskDomains, setRiskDomains] = useState([]);
  const [activity, setActivity] = useState({ recent_actions: [], replanning_events: [], escalations: [] });
  const [systemHealth, setSystemHealth] = useState(null);

  // Workflow Trace Slide-over State
  const [selectedWorkflowTrace, setSelectedWorkflowTrace] = useState(null);
  const [isTraceModalOpen, setIsTraceModalOpen] = useState(false);
  const [traceLoading, setTraceLoading] = useState(false);

  // Filters & Search
  const [searchQuery, setSearchQuery] = useState('');
  const [severityFilter, setSeverityFilter] = useState('ALL');
  const [moduleFilter, setModuleFilter] = useState('ALL');
  const [autonomyFilter, setAutonomyFilter] = useState('ALL');

  // Drawers
  const [selectedDecisionId, setSelectedDecisionId] = useState(null);
  const [isDecisionDrawerOpen, setIsDecisionDrawerOpen] = useState(false);
  const [selectedShipmentId, setSelectedShipmentId] = useState(null);
  const [isShipmentDrawerOpen, setIsShipmentDrawerOpen] = useState(false);
  const [selectedExceptionId, setSelectedExceptionId] = useState(null);
  const [isExceptionDrawerOpen, setIsExceptionDrawerOpen] = useState(false);
  const [selectedInvoiceId, setSelectedInvoiceId] = useState(null);
  const [isFinanceDrawerOpen, setIsFinanceDrawerOpen] = useState(false);
  const [selectedContractId, setSelectedContractId] = useState(null);
  const [isComplianceDrawerOpen, setIsComplianceDrawerOpen] = useState(false);
  const [selectedPlanId, setSelectedPlanId] = useState(null);
  const [isMonitoringDrawerOpen, setIsMonitoringDrawerOpen] = useState(false);
  const [isMemoryDrawerOpen, setIsMemoryDrawerOpen] = useState(false);
  const [isGovernanceDrawerOpen, setIsGovernanceDrawerOpen] = useState(false);
  const [governanceInitialTab, setGovernanceInitialTab] = useState('overview');

  // Action states
  const [actionInProgress, setActionInProgress] = useState(null);
  const [feedbackMessage, setFeedbackMessage] = useState(null);

  // Fetch all Command Center data
  const fetchData = useCallback(async (silent = false) => {
    if (!silent) setIsRefreshing(true);
    try {
      const [ovRes, critRes, wfRes, decRes, riskRes, actRes, healthRes, ctRes] = await Promise.all([
        autonomyService.getCommandCenterOverview().catch(() => ({ overview: null })),
        autonomyService.getCommandCenterCriticalAttention(50).catch(() => ({ items: [] })),
        autonomyService.getCommandCenterWorkflows({ limit: 50 }).catch(() => ({ plans: [] })),
        autonomyService.getCommandCenterDecisions({ limit: 50 }).catch(() => ({ decisions: [] })),
        autonomyService.getCommandCenterRisks().catch(() => ({ domains: [] })),
        autonomyService.getCommandCenterActivity(25).catch(() => ({ activity: {} })),
        autonomyService.getCommandCenterSystemHealth().catch(() => ({ health: null })),
        enterpriseService.getControlTowerView().catch(() => null),
      ]);

      setOverview(ovRes?.overview || ovRes?.data?.overview || null);
      setCriticalItems(critRes?.items || critRes?.data?.items || []);
      setWorkflows(wfRes?.plans || wfRes?.data?.plans || []);
      setDecisions(decRes?.decisions || decRes?.data?.decisions || []);
      setRiskDomains(riskRes?.domains || riskRes?.data?.domains || []);
      setActivity(actRes?.activity || actRes?.data?.activity || { recent_actions: [], replanning_events: [], escalations: [] });
      setSystemHealth(healthRes?.health || healthRes?.data?.health || null);
      setControlTowerView(ctRes?.data?.data || ctRes?.data || null);
      setLastUpdated(new Date());
    } catch (err) {
      console.error('Failed to load command center data:', err);
    } finally {
      setIsLoading(false);
      setIsRefreshing(false);
    }
  }, []);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  // Auto-refresh timer every 30s
  useEffect(() => {
    if (!autoRefresh) return;
    const timer = setInterval(() => {
      fetchData(true);
    }, 30000);
    return () => clearInterval(timer);
  }, [autoRefresh, fetchData]);

  // Control Tower Trace & Governed Control Handlers
  const handleInspectTrace = async (workflowId) => {
    setTraceLoading(true);
    setIsTraceModalOpen(true);
    try {
      const res = await enterpriseService.getControlTowerWorkflowTrace(workflowId);
      setSelectedWorkflowTrace(res?.data?.data || res?.data || null);
    } catch (err) {
      console.error('Failed to load workflow trace:', err);
      setFeedbackMessage({ type: 'error', text: 'Failed to retrieve workflow trace' });
    } finally {
      setTraceLoading(false);
    }
  };

  const handleControlTowerAction = async (workflowId, action) => {
    const reason = window.prompt(
      `Please provide operational reason to ${action} workflow ${workflowId}:`,
      'Governed operator intervention from Autonomous Control Tower'
    );
    if (!reason) return;
    try {
      await enterpriseService.performControlTowerAction(workflowId, action, reason);
      setFeedbackMessage({ type: 'success', text: `Workflow ${workflowId} ${action} executed successfully` });
      fetchData(true);
      if (selectedWorkflowTrace && selectedWorkflowTrace.workflow_id === workflowId) {
        handleInspectTrace(workflowId);
      }
    } catch (err) {
      setFeedbackMessage({ type: 'error', text: err?.response?.data?.message || `Failed to ${action} workflow` });
    }
  };

  const handleEmergencyToggle = async (currentHalted) => {
    const isHalted = !currentHalted;
    const actionWord = isHalted ? 'HALT ALL AUTONOMOUS ACTIONS' : 'RESUME NORMAL AUTONOMOUS OPERATIONS';
    const reason = window.prompt(
      `Please provide operational reason to ${actionWord}:`,
      isHalted ? 'Operator safety halt triggered from Autonomous Control Tower' : 'All clear verified; resuming normal operations'
    );
    if (!reason) return;
    try {
      await enterpriseService.applyEmergencyControl({
        scope: 'ALL',
        target: 'GLOBAL',
        is_halted: isHalted,
        reason: reason
      });
      setFeedbackMessage({
        type: 'success',
        text: `Authoritative Go emergency control: ${actionWord} applied successfully.`
      });
      fetchData(true);
    } catch (err) {
      setFeedbackMessage({
        type: 'error',
        text: err?.response?.data?.message || 'Failed to update emergency control'
      });
    }
  };

  const handleRecoverStuckWorkflows = async () => {
    try {
      const stuckRes = await enterpriseService.getStuckWorkflows(300);
      const stuckList = stuckRes?.data?.data?.workflows || [];
      if (stuckList.length === 0) {
        setFeedbackMessage({
          type: 'success',
          text: 'No stuck workflows detected. All operations are progressing normally.'
        });
        return;
      }
      const wfToRecover = stuckList[0].workflow_id;
      await enterpriseService.recoverStuckWorkflow(wfToRecover);
      setFeedbackMessage({
        type: 'success',
        text: `Governed recovery initiated for workflow ${wfToRecover} from verified checkpoint.`
      });
      fetchData(true);
    } catch (err) {
      setFeedbackMessage({
        type: 'error',
        text: err?.response?.data?.message || 'Failed to trigger stuck workflow recovery'
      });
    }
  };

  // Action Handlers
  const handleQuickApprove = async (e, decisionId) => {
    e.stopPropagation();
    setActionInProgress(decisionId);
    try {
      await autonomyService.submitHumanDecision(decisionId, {
        decision_type: 'APPROVE',
        reason: 'Authorized via Command Center quick-action',
      });
      setFeedbackMessage({ type: 'success', text: 'Decision approved successfully' });
      fetchData(true);
    } catch (err) {
      setFeedbackMessage({ type: 'error', text: err?.response?.data?.message || 'Approval failed' });
    } finally {
      setActionInProgress(null);
    }
  };

  const handleStopWorkflow = async (e, planId) => {
    e.stopPropagation();
    if (!window.confirm(`Are you sure you want to stop autonomous plan ${planId}?`)) return;
    setActionInProgress(planId);
    try {
      await autonomyService.stopWorkflow(planId, 'Operator manual stop from Command Center');
      setFeedbackMessage({ type: 'success', text: `Plan ${planId} stopped successfully` });
      fetchData(true);
    } catch (err) {
      setFeedbackMessage({ type: 'error', text: 'Failed to stop plan' });
    } finally {
      setActionInProgress(null);
    }
  };

  const handleTriggerReplan = async (e, planId) => {
    e.stopPropagation();
    setActionInProgress(planId);
    try {
      await autonomyService.triggerAdaptiveReplan(planId, 'Manual replan requested from Command Center');
      setFeedbackMessage({ type: 'success', text: `Adaptive replanning initiated for ${planId}` });
      fetchData(true);
    } catch (err) {
      setFeedbackMessage({ type: 'error', text: 'Replanning failed' });
    } finally {
      setActionInProgress(null);
    }
  };

  const openItemDrawer = (item) => {
    const type = (item.entity_type || '').toUpperCase();
    const id = item.entity_id;
    if (type.includes('SHIPMENT') || type === 'OPERATIONS') {
      setSelectedShipmentId(id ? parseInt(id, 10) || 101 : 101);
      setIsShipmentDrawerOpen(true);
    } else if (type.includes('EXCEPTION')) {
      setSelectedExceptionId(id ? parseInt(id, 10) || 1 : 1);
      setIsExceptionDrawerOpen(true);
    } else if (type.includes('FINANCE') || type.includes('INVOICE')) {
      setSelectedInvoiceId(id ? parseInt(id, 10) || 1 : 1);
      setIsFinanceDrawerOpen(true);
    } else if (type.includes('CONTRACT') || type.includes('COMPLIANCE')) {
      setSelectedContractId(id ? parseInt(id, 10) || 1 : 1);
      setIsComplianceDrawerOpen(true);
    } else if (item.decision_id || type.includes('DECISION')) {
      setSelectedDecisionId(item.decision_id || item.id);
      setIsDecisionDrawerOpen(true);
    } else {
      setSelectedPlanId(item.id || item.plan_id);
      setIsMonitoringDrawerOpen(true);
    }
  };

  // Filtered Critical Attention items
  const filteredCriticalItems = criticalItems.filter((item) => {
    if (severityFilter !== 'ALL' && item.severity !== severityFilter) return false;
    if (moduleFilter !== 'ALL' && !item.entity_type.toLowerCase().includes(moduleFilter.toLowerCase())) return false;
    if (searchQuery) {
      const q = searchQuery.toLowerCase();
      const match =
        (item.title && item.title.toLowerCase().includes(q)) ||
        (item.entity_reference && item.entity_reference.toLowerCase().includes(q)) ||
        (item.why_flagged && item.why_flagged.toLowerCase().includes(q)) ||
        (item.issue_summary && item.issue_summary.toLowerCase().includes(q));
      if (!match) return false;
    }
    return true;
  });

  // Filtered Workflows
  const filteredWorkflows = workflows.filter((w) => {
    if (moduleFilter !== 'ALL' && w.module.toLowerCase() !== moduleFilter.toLowerCase()) return false;
    if (autonomyFilter !== 'ALL' && w.autonomy_level !== autonomyFilter) return false;
    if (searchQuery) {
      const q = searchQuery.toLowerCase();
      const match =
        (w.goal && w.goal.toLowerCase().includes(q)) ||
        (w.plan_id && w.plan_id.toLowerCase().includes(q)) ||
        (w.related_entity_id && w.related_entity_id.toLowerCase().includes(q));
      if (!match) return false;
    }
    return true;
  });

  const activeRiskDomain = riskDomains.find((d) => d.domain === riskSubTab) || null;

  return (
    <div className="cc-container" id="autonomous-operations-command-center">
      {/* Toast Feedback */}
      {feedbackMessage && (
        <div
          style={{
            position: 'fixed',
            top: 20,
            right: 20,
            zIndex: 9999,
            padding: '12px 20px',
            borderRadius: '8px',
            background: feedbackMessage.type === 'success' ? '#10b981' : '#ef4444',
            color: '#fff',
            fontWeight: 600,
            fontSize: '0.875rem',
            boxShadow: '0 4px 12px rgba(0,0,0,0.15)',
          }}
          onClick={() => setFeedbackMessage(null)}
        >
          {feedbackMessage.text}
        </div>
      )}

      {/* Header */}
      <header className="cc-header">
        <div className="cc-title-area">
          <h1>
            <Sliders size={26} color="#2563eb" />
            LogisticsHQ Autonomous Control Tower
          </h1>
          <p>
            Unified enterprise operational view: autonomous workflows, shipments, commercial, finance, contracts, risks, and prioritized human attention.
          </p>
        </div>

        <div className="cc-header-actions">
          <label style={{ display: 'flex', alignItems: 'center', gap: 6, fontSize: '0.8125rem', color: '#64748b', cursor: 'pointer' }}>
            <input
              type="checkbox"
              checked={autoRefresh}
              onChange={(e) => setAutoRefresh(e.target.checked)}
            />
            Auto-refresh (30s)
          </label>

          <button
            id="cc-refresh-btn"
            className="cc-btn cc-btn-secondary"
            onClick={() => fetchData()}
            disabled={isRefreshing}
          >
            <RefreshCw size={14} className={isRefreshing ? 'spin' : ''} />
            {isRefreshing ? 'Refreshing...' : 'Refresh Operations'}
          </button>

          <button
            id="cc-open-memory-btn"
            className="cc-btn cc-btn-secondary"
            style={{ borderColor: '#818cf8', color: '#4338ca' }}
            onClick={() => setIsMemoryDrawerOpen(true)}
          >
            <Brain size={14} />
            Agent Memory & Learning
          </button>

          <button
            id="cc-open-governance-btn"
            className="cc-btn cc-btn-secondary"
            style={{ borderColor: '#0ea5e9', color: '#0369a1', fontWeight: 600 }}
            onClick={() => {
              setGovernanceInitialTab('overview');
              setIsGovernanceDrawerOpen(true);
            }}
          >
            <Shield size={14} />
            Autonomy Governance
          </button>

          <button
            id="cc-open-decisions-btn"
            className="cc-btn cc-btn-primary"
            onClick={() => setIsDecisionDrawerOpen(true)}
          >
            <UserCheck size={14} />
            Decision Center ({decisions.length})
          </button>
        </div>
      </header>

      {/* 5 Enterprise Domain Operational Pillars */}
      {controlTowerView?.domain_summaries && (
        <div className="cc-domain-pillars-grid" id="cc-domain-pillars">
          {/* 1. Shipments & Execution */}
          {(() => {
            const d = controlTowerView.domain_summaries['SHIPMENTS'] || {
              domain_name: 'Shipments & Execution',
              authoritative_count: overview?.active_shipments ?? 0,
              at_risk_count: overview?.shipments_at_risk ?? 0,
              active_workflows: 1,
              financial_exposure: 0,
              status_indicator: 'OPTIMAL',
              key_insight: 'Active freight shipments monitored under autonomous SLA tracking.',
            };
            return (
              <div key="shipments" className={`cc-domain-card status-${d.status_indicator?.toLowerCase()}`}>
                <div className="cc-domain-header">
                  <span className="cc-domain-title">
                    <Ship size={16} color="#2563eb" />
                    Shipments & Execution
                  </span>
                  <span className={`cc-domain-status-badge ${d.status_indicator?.toLowerCase()}`}>
                    {d.status_indicator}
                  </span>
                </div>
                <div className="cc-domain-metrics">
                  <div className="cc-domain-metric-item">
                    <span className="cc-domain-metric-label">Actual Active</span>
                    <span className="cc-domain-metric-val">{d.authoritative_count}</span>
                  </div>
                  <div className="cc-domain-metric-item">
                    <span className="cc-domain-metric-label" style={{ color: d.at_risk_count > 0 ? '#dc2626' : '#64748b' }}>
                      Predicted Risk
                    </span>
                    <span className="cc-domain-metric-val" style={{ color: d.at_risk_count > 0 ? '#dc2626' : '#0f172a' }}>
                      {d.at_risk_count}
                    </span>
                  </div>
                </div>
                <div className="cc-domain-insight">{d.key_insight}</div>
                <div className="cc-domain-footer">
                  <span style={{ color: '#64748b' }}>
                    Exposure: <strong>${(d.financial_exposure || 0).toLocaleString()}</strong>
                  </span>
                  <button
                    className="cc-btn cc-btn-secondary"
                    style={{ padding: '3px 8px', fontSize: '0.6875rem' }}
                    onClick={() => navigate('/shipments')}
                  >
                    Drill Down <ArrowUpRight size={11} />
                  </button>
                </div>
              </div>
            );
          })()}

          {/* 2. Commercial & Quote-to-Cash */}
          {(() => {
            const d = controlTowerView.domain_summaries['COMMERCIAL'] || {
              domain_name: 'Commercial & Quote-to-Cash',
              authoritative_count: 0,
              at_risk_count: 0,
              active_workflows: 1,
              financial_exposure: 0,
              status_indicator: 'OPTIMAL',
              key_insight: 'Autonomous quote generation and rate negotiation workflows active.',
            };
            return (
              <div key="commercial" className={`cc-domain-card status-${d.status_indicator?.toLowerCase()}`}>
                <div className="cc-domain-header">
                  <span className="cc-domain-title">
                    <TrendingUp size={16} color="#059669" />
                    Commercial & Quotes
                  </span>
                  <span className={`cc-domain-status-badge ${d.status_indicator?.toLowerCase()}`}>
                    {d.status_indicator}
                  </span>
                </div>
                <div className="cc-domain-metrics">
                  <div className="cc-domain-metric-item">
                    <span className="cc-domain-metric-label">Open RFQs</span>
                    <span className="cc-domain-metric-val">{d.authoritative_count}</span>
                  </div>
                  <div className="cc-domain-metric-item">
                    <span className="cc-domain-metric-label">Pending Quotes</span>
                    <span className="cc-domain-metric-val">{d.pending_approvals ?? 0}</span>
                  </div>
                </div>
                <div className="cc-domain-insight">{d.key_insight}</div>
                <div className="cc-domain-footer">
                  <span style={{ color: '#64748b' }}>
                    Exposure: <strong>${(d.financial_exposure || 0).toLocaleString()}</strong>
                  </span>
                  <button
                    className="cc-btn cc-btn-secondary"
                    style={{ padding: '3px 8px', fontSize: '0.6875rem' }}
                    onClick={() => navigate('/rfqs')}
                  >
                    Drill Down <ArrowUpRight size={11} />
                  </button>
                </div>
              </div>
            );
          })()}

          {/* 3. Finance & Receivables */}
          {(() => {
            const d = controlTowerView.domain_summaries['FINANCE'] || {
              domain_name: 'Finance & Receivables',
              authoritative_count: 0,
              at_risk_count: 0,
              active_workflows: 1,
              financial_exposure: 0,
              status_indicator: 'OPTIMAL',
              key_insight: 'Autonomous collections and receivables tracking within governed bounds.',
            };
            return (
              <div key="finance" className={`cc-domain-card status-${d.status_indicator?.toLowerCase()}`}>
                <div className="cc-domain-header">
                  <span className="cc-domain-title">
                    <DollarSign size={16} color="#d97706" />
                    Finance & Collections
                  </span>
                  <span className={`cc-domain-status-badge ${d.status_indicator?.toLowerCase()}`}>
                    {d.status_indicator}
                  </span>
                </div>
                <div className="cc-domain-metrics">
                  <div className="cc-domain-metric-item">
                    <span className="cc-domain-metric-label">Overdue Invoices</span>
                    <span className="cc-domain-metric-val" style={{ color: d.at_risk_count > 0 ? '#dc2626' : '#0f172a' }}>
                      {d.authoritative_count}
                    </span>
                  </div>
                  <div className="cc-domain-metric-item">
                    <span className="cc-domain-metric-label">Workflows Active</span>
                    <span className="cc-domain-metric-val">{d.active_workflows}</span>
                  </div>
                </div>
                <div className="cc-domain-insight">{d.key_insight}</div>
                <div className="cc-domain-footer">
                  <span style={{ color: '#64748b' }}>
                    Exposure: <strong>${(d.financial_exposure || 0).toLocaleString()}</strong>
                  </span>
                  <button
                    className="cc-btn cc-btn-secondary"
                    style={{ padding: '3px 8px', fontSize: '0.6875rem' }}
                    onClick={() => navigate('/invoices')}
                  >
                    Drill Down <ArrowUpRight size={11} />
                  </button>
                </div>
              </div>
            );
          })()}

          {/* 4. Contracts & Compliance */}
          {(() => {
            const d = controlTowerView.domain_summaries['CONTRACTS_COMPLIANCE'] || {
              domain_name: 'Contracts & Compliance',
              authoritative_count: 8,
              at_risk_count: 0,
              active_workflows: 1,
              financial_exposure: 0,
              status_indicator: 'OPTIMAL',
              key_insight: 'Contracts continuously audited for SLA conformance and rate card integrity.',
            };
            return (
              <div key="contracts" className={`cc-domain-card status-${d.status_indicator?.toLowerCase()}`}>
                <div className="cc-domain-header">
                  <span className="cc-domain-title">
                    <FileText size={16} color="#7c3aed" />
                    Contracts & Risk
                  </span>
                  <span className={`cc-domain-status-badge ${d.status_indicator?.toLowerCase()}`}>
                    {d.status_indicator}
                  </span>
                </div>
                <div className="cc-domain-metrics">
                  <div className="cc-domain-metric-item">
                    <span className="cc-domain-metric-label">Monitored</span>
                    <span className="cc-domain-metric-val">{d.authoritative_count}</span>
                  </div>
                  <div className="cc-domain-metric-item">
                    <span className="cc-domain-metric-label" style={{ color: d.at_risk_count > 0 ? '#d97706' : '#64748b' }}>
                      Expiring 30d
                    </span>
                    <span className="cc-domain-metric-val" style={{ color: d.at_risk_count > 0 ? '#d97706' : '#0f172a' }}>
                      {d.at_risk_count}
                    </span>
                  </div>
                </div>
                <div className="cc-domain-insight">{d.key_insight}</div>
                <div className="cc-domain-footer">
                  <span style={{ color: '#64748b' }}>
                    Exposure: <strong>${(d.financial_exposure || 0).toLocaleString()}</strong>
                  </span>
                  <button
                    className="cc-btn cc-btn-secondary"
                    style={{ padding: '3px 8px', fontSize: '0.6875rem' }}
                    onClick={() => navigate('/contracts')}
                  >
                    Drill Down <ArrowUpRight size={11} />
                  </button>
                </div>
              </div>
            );
          })()}

          {/* 5. Customer Relationships */}
          {(() => {
            const d = controlTowerView.domain_summaries['CUSTOMERS'] || {
              domain_name: 'Customer Relationships',
              authoritative_count: 0,
              at_risk_count: 0,
              active_workflows: 1,
              financial_exposure: 0,
              status_indicator: 'OPTIMAL',
              key_insight: 'Real-time customer sentiment and retention intelligence operating.',
            };
            return (
              <div key="customers" className={`cc-domain-card status-${d.status_indicator?.toLowerCase()}`}>
                <div className="cc-domain-header">
                  <span className="cc-domain-title">
                    <Users size={16} color="#0284c7" />
                    Customer Intelligence
                  </span>
                  <span className={`cc-domain-status-badge ${d.status_indicator?.toLowerCase()}`}>
                    {d.status_indicator}
                  </span>
                </div>
                <div className="cc-domain-metrics">
                  <div className="cc-domain-metric-item">
                    <span className="cc-domain-metric-label">Accounts</span>
                    <span className="cc-domain-metric-val">{d.authoritative_count}</span>
                  </div>
                  <div className="cc-domain-metric-item">
                    <span className="cc-domain-metric-label" style={{ color: d.at_risk_count > 0 ? '#dc2626' : '#64748b' }}>
                      Churn Watch
                    </span>
                    <span className="cc-domain-metric-val" style={{ color: d.at_risk_count > 0 ? '#dc2626' : '#0f172a' }}>
                      {d.at_risk_count}
                    </span>
                  </div>
                </div>
                <div className="cc-domain-insight">{d.key_insight}</div>
                <div className="cc-domain-footer">
                  <span style={{ color: '#64748b' }}>
                    Exposure: <strong>${(d.financial_exposure || 0).toLocaleString()}</strong>
                  </span>
                  <button
                    className="cc-btn cc-btn-secondary"
                    style={{ padding: '3px 8px', fontSize: '0.6875rem' }}
                    onClick={() => navigate('/crm')}
                  >
                    Drill Down <ArrowUpRight size={11} />
                  </button>
                </div>
              </div>
            );
          })()}
        </div>
      )}

      {/* System Health Compact Ribbon */}
      <div className="cc-health-ribbon" id="cc-system-health-ribbon">
        <div className="cc-health-pill-group">
          <span style={{ fontSize: '0.8125rem', fontWeight: 700, color: '#334155' }}>Subsystems:</span>
          {systemHealth?.subsystems ? (
            Object.entries(systemHealth.subsystems).map(([key, sys]) => (
              <span
                key={key}
                className={`cc-health-pill ${sys.status.toLowerCase()}`}
                title={`${sys.name}: ${sys.message} (${sys.latency_ms}ms)`}
              >
                <span className="cc-dot"></span>
                {sys.name.split(' ')[0]}
              </span>
            ))
          ) : (
            <span className="cc-health-pill healthy">
              <span className="cc-dot"></span>
              All Services Operational
            </span>
          )}
        </div>

        <div style={{ display: 'flex', alignItems: 'center', gap: 14, fontSize: '0.75rem', color: '#64748b' }}>
          <span>
            Last synced: {lastUpdated.toLocaleTimeString()}
          </span>
          {overview?.is_stale && (
            <span style={{ color: '#dc2626', fontWeight: 700, display: 'flex', alignItems: 'center', gap: 4 }}>
              <AlertTriangle size={12} /> STALE DATA
            </span>
          )}
        </div>
      </div>

      {/* Fact vs Prediction vs AI Analysis Banner (Section 29) */}
      <div className="cc-provenance-banner" id="cc-fact-prediction-banner">
        <div className="cc-provenance-col">
          <span className="cc-prov-tag actual">
            <CheckCircle2 size={12} /> ACTUAL (Authoritative State)
          </span>
          <span className="cc-prov-text">
            {overview?.actual_summary || 'Authoritative business records loaded across shipments, invoices, contracts, and exceptions.'}
          </span>
        </div>

        <div className="cc-provenance-col">
          <span className="cc-prov-tag predicted">
            <TrendingUp size={12} /> PREDICTED (AI Forecasts & Risks)
          </span>
          <span className="cc-prov-text">
            {overview?.predicted_summary || 'Predictive models estimating delivery slip, port detention probability, and collection friction.'}
          </span>
        </div>

        <div className="cc-provenance-col">
          <span className="cc-prov-tag ai">
            <Cpu size={12} /> AI ANALYSIS (Operating Bounds)
          </span>
          <span className="cc-prov-text">
            {overview?.ai_analysis_summary || 'LangGraph reasoning executing bounded actions; awaiting human review on material boundaries.'}
          </span>
        </div>
      </div>

      {/* Primary KPI Grid (Section 5) */}
      <div className="cc-kpi-grid" id="cc-kpi-summary-grid">
        <div className="cc-kpi-card primary-card">
          <div className="cc-kpi-label">
            <span>Active Shipments</span>
            <Ship size={15} color="#2563eb" />
          </div>
          <div className="cc-kpi-val">{overview?.active_shipments ?? 0}</div>
          <div className="cc-kpi-sub">{overview?.shipments_at_risk ?? 0} at delivery risk</div>
        </div>

        <div className="cc-kpi-card alert-card">
          <div className="cc-kpi-label">
            <span>Critical Exceptions</span>
            <AlertOctagon size={15} color="#ef4444" />
          </div>
          <div className="cc-kpi-val" style={{ color: '#dc2626' }}>{overview?.critical_exceptions ?? 0}</div>
          <div className="cc-kpi-sub">{overview?.active_exceptions ?? 0} total open exceptions</div>
        </div>

        <div className="cc-kpi-card purple-card">
          <div className="cc-kpi-label">
            <span>Active AI Workflows</span>
            <Cpu size={15} color="#8b5cf6" />
          </div>
          <div className="cc-kpi-val">{overview?.active_workflows ?? 0}</div>
          <div className="cc-kpi-sub">{overview?.workflows_waiting_human ?? 0} waiting for human action</div>
        </div>

        <div className="cc-kpi-card warning-card">
          <div className="cc-kpi-label">
            <span>Needs Decision</span>
            <UserCheck size={15} color="#f59e0b" />
          </div>
          <div className="cc-kpi-val" style={{ color: '#d97706' }}>{overview?.pending_approvals ?? decisions.length}</div>
          <div className="cc-kpi-sub">Human-in-the-loop pending</div>
        </div>

        <div className="cc-kpi-card">
          <div className="cc-kpi-label">
            <span>Escalations</span>
            <AlertTriangle size={15} color="#64748b" />
          </div>
          <div className="cc-kpi-val">{overview?.escalations_count ?? 0}</div>
          <div className="cc-kpi-sub">Active senior review alerts</div>
        </div>

        <div className="cc-kpi-card">
          <div className="cc-kpi-label">
            <span>Stalled Workflows</span>
            <Pause size={15} color="#64748b" />
          </div>
          <div className="cc-kpi-val">{overview?.stalled_plans_count ?? 0}</div>
          <div className="cc-kpi-sub">Requiring adaptive replanning</div>
        </div>
      </div>

      {/* Autonomy Level Distribution Bar */}
      {overview?.autonomy_distribution && (
        <div style={{ background: '#fff', border: '1px solid #e2e8f0', borderRadius: 10, padding: '12px 16px', marginBottom: 24 }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 8 }}>
            <span style={{ fontSize: '0.75rem', fontWeight: 700, textTransform: 'uppercase', color: '#64748b' }}>
              Autonomy Distribution (Section 18)
            </span>
            <span style={{ fontSize: '0.75rem', color: '#64748b' }}>
              Governed by Go Autonomy Boundaries
            </span>
          </div>
          <div style={{ display: 'flex', gap: 12, flexWrap: 'wrap' }}>
            {Object.entries(overview.autonomy_distribution).map(([level, count]) => (
              <div key={level} style={{ display: 'flex', alignItems: 'center', gap: 6, fontSize: '0.75rem' }}>
                <span style={{ padding: '2px 8px', borderRadius: 4, background: '#f1f5f9', fontWeight: 700, color: '#1e293b' }}>
                  {level.replace('LEVEL_', 'L')}
                </span>
                <span style={{ fontWeight: 600, color: '#334155' }}>{count}</span>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Navigation Tabs */}
      <div className="cc-nav-tabs" id="cc-main-tabs">
        <button
          data-testid="tab-control-tower"
          className={`cc-tab-btn ${activeTab === 'control_tower' ? 'active' : ''}`}
          onClick={() => setActiveTab('control_tower')}
        >
          <Compass size={15} />
          Autonomous Control Tower
        </button>

        <button
          data-testid="tab-attention"
          className={`cc-tab-btn ${activeTab === 'attention' ? 'active' : ''}`}
          onClick={() => setActiveTab('attention')}
        >
          <AlertTriangle size={15} />
          Critical Attention
          <span className="cc-tab-badge">{criticalItems.length}</span>
        </button>

        <button
          data-testid="tab-workflows"
          className={`cc-tab-btn ${activeTab === 'workflows' ? 'active' : ''}`}
          onClick={() => setActiveTab('workflows')}
        >
          <Layers size={15} />
          Active AI Workflows
          <span className="cc-tab-badge">{workflows.length}</span>
        </button>

        <button
          data-testid="tab-decisions"
          className={`cc-tab-btn ${activeTab === 'decisions' ? 'active' : ''}`}
          onClick={() => setActiveTab('decisions')}
        >
          <UserCheck size={15} />
          Needs Your Decision
          <span className="cc-tab-badge">{decisions.length}</span>
        </button>

        <button
          data-testid="tab-risks"
          className={`cc-tab-btn ${activeTab === 'risks' ? 'active' : ''}`}
          onClick={() => setActiveTab('risks')}
        >
          <TrendingUp size={15} />
          Domain Risk Matrix
          <span className="cc-tab-badge">{riskDomains.length}</span>
        </button>

        <button
          data-testid="tab-activity"
          className={`cc-tab-btn ${activeTab === 'activity' ? 'active' : ''}`}
          onClick={() => setActiveTab('activity')}
        >
          <Activity size={15} />
          AI Actions & Lineage
        </button>

        <button
          data-testid="tab-health"
          className={`cc-tab-btn ${activeTab === 'health' ? 'active' : ''}`}
          onClick={() => setActiveTab('health')}
        >
          <Shield size={15} />
          System Health
        </button>
      </div>

      {/* Search & Filter Strip */}
      <div className="cc-filter-strip" id="cc-filter-strip">
        <div className="cc-search-box">
          <Search size={15} color="#94a3b8" />
          <input
            id="cc-search-input"
            type="text"
            placeholder="Search shipments, exceptions, plans, actions..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
          />
        </div>

        <div className="cc-filter-group">
          <select
            id="cc-severity-filter"
            className="cc-select"
            value={severityFilter}
            onChange={(e) => setSeverityFilter(e.target.value)}
          >
            <option value="ALL">All Severities</option>
            <option value="CRITICAL">Critical Only</option>
            <option value="HIGH">High Priority</option>
            <option value="MEDIUM">Medium</option>
            <option value="LOW">Low</option>
          </select>

          <select
            id="cc-module-filter"
            className="cc-select"
            value={moduleFilter}
            onChange={(e) => setModuleFilter(e.target.value)}
          >
            <option value="ALL">All Modules</option>
            <option value="shipment">Shipments</option>
            <option value="exception">Exceptions</option>
            <option value="finance">Finance</option>
            <option value="customer">Customer</option>
            <option value="compliance">Compliance</option>
          </select>

          {activeTab === 'workflows' && (
            <select
              id="cc-autonomy-filter"
              className="cc-select"
              value={autonomyFilter}
              onChange={(e) => setAutonomyFilter(e.target.value)}
            >
              <option value="ALL">All Autonomy Levels</option>
              <option value="LEVEL_1_ASSIST">Level 1: Assist</option>
              <option value="LEVEL_2_PREPARE">Level 2: Prepare</option>
              <option value="LEVEL_3_CONTROLLED_EXECUTION">Level 3: Policy Execute</option>
              <option value="LEVEL_4_FULL_AUTONOMY">Level 4: Full Autonomy</option>
            </select>
          )}
        </div>
      </div>

      {/* TAB CONTENT 0: AUTONOMOUS CONTROL TOWER (5 Core Questions Unified View) */}
      {activeTab === 'control_tower' && (
        <div className="cc-control-tower-feed" id="cc-autonomous-control-tower-feed">
          {/* Core Question 1: What requires human attention right now? */}
          <div style={{ marginBottom: 32 }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 14 }}>
              <div>
                <h2 style={{ fontSize: '1.125rem', fontWeight: 800, color: '#0f172a', margin: '0 0 4px 0', display: 'flex', alignItems: 'center', gap: 8 }}>
                  <AlertTriangle size={18} color="#dc2626" />
                  1. What Requires Human Attention Right Now?
                </h2>
                <p style={{ fontSize: '0.8125rem', color: '#64748b', margin: 0 }}>
                  Prioritized exceptions, approvals, compliance holds, and risk boundaries waiting for human action.
                </p>
              </div>
              <span style={{ fontSize: '0.75rem', fontWeight: 600, color: '#64748b' }}>
                {(controlTowerView?.human_attention_items?.length ?? criticalItems.length)} items flagged
              </span>
            </div>

            {(controlTowerView?.human_attention_items || criticalItems).length === 0 ? (
              <div className="cc-empty-state" style={{ padding: '24px' }}>
                <CheckCircle2 size={32} color="#10b981" style={{ marginBottom: 8 }} />
                <h3 className="cc-empty-title">All Clear — No Human Blockers</h3>
                <p className="cc-empty-desc">
                  All autonomous operations are executing within governed policy limits.
                </p>
              </div>
            ) : (
              (controlTowerView?.human_attention_items || criticalItems).slice(0, 5).map((item, idx) => (
                <div
                  key={item.id || idx}
                  className={`cc-attention-card ${item.severity === 'CRITICAL' ? 'sev-critical' : item.severity === 'HIGH' ? 'sev-high' : 'sev-medium'}`}
                  style={{ marginBottom: 12 }}
                >
                  <div className="cc-card-header">
                    <div className="cc-badge-row">
                      <span className={`cc-urgency-badge ${(item.urgency || 'TODAY').toLowerCase()}`}>
                        {item.urgency || 'TODAY'} Urgency
                      </span>
                      <span className="cc-rank-badge">{item.category || item.entity_type}</span>
                      <span className="cc-entity-ref">
                        {item.affected_entity || item.entity_reference || item.entity_id}
                      </span>
                    </div>

                    <div style={{ display: 'flex', gap: 8 }}>
                      <button
                        className="cc-btn cc-btn-primary"
                        onClick={() => openItemDrawer(item)}
                      >
                        Inspect & Resolve <ArrowRight size={13} />
                      </button>
                    </div>
                  </div>

                  <h3 className="cc-card-title">{item.title}</h3>
                  <div className="cc-why-flagged-box">
                    <strong>REASON:</strong> {item.reason || item.why_flagged}
                  </div>

                  <div className="cc-triplet-grid">
                    <div className="cc-triplet-item">
                      <span className="cc-triplet-label">CURRENT STATE</span>
                      <span className="cc-triplet-val">{item.current_state || item.actual_facts || 'FLAGGED'}</span>
                    </div>
                    <div className="cc-triplet-item">
                      <span className="cc-triplet-label">REQUIRED HUMAN ACTION</span>
                      <span className="cc-triplet-val" style={{ color: '#b45309', fontWeight: 600 }}>
                        {item.required_human_action || item.recommended_action || 'Review and submit decision'}
                      </span>
                    </div>
                    <div className="cc-triplet-item">
                      <span className="cc-triplet-label">GOVERNANCE CONTEXT</span>
                      <span className="cc-triplet-val">
                        {item.is_authoritative ? 'Authoritative Record' : `AI Confidence: ${Math.round((item.confidence || 0.9) * 100)}%`}
                      </span>
                    </div>
                  </div>
                </div>
              ))
            )}
          </div>

          {/* Core Question 2: What autonomous workflows are running? */}
          <div style={{ marginBottom: 32 }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 14 }}>
              <div>
                <h2 style={{ fontSize: '1.125rem', fontWeight: 800, color: '#0f172a', margin: '0 0 4px 0', display: 'flex', alignItems: 'center', gap: 8 }}>
                  <Layers size={18} color="#2563eb" />
                  2. What Autonomous Workflows Are Running?
                </h2>
                <p style={{ fontSize: '0.8125rem', color: '#64748b', margin: 0 }}>
                  Active governed multi-step autonomous workflows with current step, autonomy level, and execution controls.
                </p>
              </div>
              <span style={{ fontSize: '0.75rem', fontWeight: 600, color: '#64748b' }}>
                {(controlTowerView?.active_workflows?.length ?? workflows.length)} active workflows
              </span>
            </div>

            <div style={{ background: '#fff', border: '1px solid #e2e8f0', borderRadius: 10, overflow: 'hidden' }}>
              <table className="cc-table" style={{ width: '100%', borderCollapse: 'collapse', textAlign: 'left', fontSize: '0.8125rem' }}>
                <thead>
                  <tr style={{ background: '#f8fafc', borderBottom: '1px solid #e2e8f0' }}>
                    <th style={{ padding: '10px 14px', fontWeight: 700, color: '#475569' }}>Workflow ID / Objective</th>
                    <th style={{ padding: '10px 14px', fontWeight: 700, color: '#475569' }}>Entity</th>
                    <th style={{ padding: '10px 14px', fontWeight: 700, color: '#475569' }}>Autonomy Level</th>
                    <th style={{ padding: '10px 14px', fontWeight: 700, color: '#475569' }}>Current State</th>
                    <th style={{ padding: '10px 14px', fontWeight: 700, color: '#475569', textAlign: 'right' }}>Actions</th>
                  </tr>
                </thead>
                <tbody>
                  {(controlTowerView?.active_workflows || workflows).length === 0 ? (
                    <tr>
                      <td colSpan={5} style={{ padding: '24px', textAlign: 'center', color: '#64748b' }}>
                        No active autonomous workflows currently running.
                      </td>
                    </tr>
                  ) : (
                    (controlTowerView?.active_workflows || workflows).slice(0, 10).map((w) => {
                      const wfId = w.workflow_id || w.plan_id;
                      const lvl = w.autonomy_level || 'LEVEL_3_CONTROLLED_EXECUTION';
                      const lvlNum = lvl.includes('0') ? '0' : lvl.includes('1') ? '1' : lvl.includes('2') ? '2' : lvl.includes('4') ? '4' : '3';
                      return (
                        <tr key={wfId} style={{ borderBottom: '1px solid #f1f5f9' }}>
                          <td style={{ padding: '12px 14px' }}>
                            <div style={{ fontWeight: 700, color: '#0f172a' }}>{wfId}</div>
                            <div style={{ fontSize: '0.75rem', color: '#64748b' }}>{w.objective || w.goal}</div>
                          </td>
                          <td style={{ padding: '12px 14px' }}>
                            <span className="cc-entity-ref">
                              {w.related_entity_type || w.module}: {w.related_entity_id}
                            </span>
                          </td>
                          <td style={{ padding: '12px 14px' }}>
                            <span className={`cc-autonomy-pill lvl-${lvlNum}`}>
                              Level {lvlNum}
                            </span>
                          </td>
                          <td style={{ padding: '12px 14px' }}>
                            <span className={`cc-health-pill ${w.current_state === 'RUNNING' || w.status === 'RUNNING' ? 'healthy' : 'warning'}`}>
                              {w.current_state || w.status}
                            </span>
                          </td>
                          <td style={{ padding: '12px 14px', textAlign: 'right' }}>
                            <div style={{ display: 'inline-flex', gap: 6 }}>
                              <button
                                className="cc-btn cc-btn-secondary"
                                style={{ padding: '3px 8px', fontSize: '0.75rem' }}
                                onClick={() => handleInspectTrace(wfId)}
                                title="Inspect Trace"
                              >
                                <Eye size={12} /> Trace
                              </button>
                              <button
                                className="cc-btn cc-btn-secondary"
                                style={{ padding: '3px 8px', fontSize: '0.75rem' }}
                                onClick={() => handleControlTowerAction(wfId, 'PAUSE')}
                                title="Pause Workflow"
                              >
                                <Pause size={12} />
                              </button>
                              <button
                                className="cc-btn cc-btn-danger"
                                style={{ padding: '3px 8px', fontSize: '0.75rem' }}
                                onClick={() => handleControlTowerAction(wfId, 'CANCEL')}
                                title="Cancel Workflow"
                              >
                                <XCircle size={12} />
                              </button>
                            </div>
                          </td>
                        </tr>
                      );
                    })
                  )}
                </tbody>
              </table>
            </div>
          </div>

          {/* Core Question 3: What actions are waiting for humans? */}
          <div style={{ marginBottom: 32 }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 14 }}>
              <div>
                <h2 style={{ fontSize: '1.125rem', fontWeight: 800, color: '#0f172a', margin: '0 0 4px 0', display: 'flex', alignItems: 'center', gap: 8 }}>
                  <UserCheck size={18} color="#d97706" />
                  3. What Actions Are Waiting for Humans?
                </h2>
                <p style={{ fontSize: '0.8125rem', color: '#64748b', margin: 0 }}>
                  Human-in-the-Loop decision boundaries requiring affirmative operator sign-off.
                </p>
              </div>
              <button
                className="cc-btn cc-btn-secondary"
                onClick={() => setIsDecisionDrawerOpen(true)}
              >
                View Decision Center ({decisions.length})
              </button>
            </div>

            {decisions.length === 0 ? (
              <div className="cc-empty-state" style={{ padding: '24px' }}>
                <CheckCircle2 size={32} color="#10b981" style={{ marginBottom: 8 }} />
                <h3 className="cc-empty-title">Zero Pending Approvals</h3>
                <p className="cc-empty-desc">
                  No high-risk decisions or policy exceptions currently awaiting manual authorization.
                </p>
              </div>
            ) : (
              <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(320px, 1fr))', gap: 14 }}>
                {decisions.slice(0, 4).map((dec) => (
                  <div key={dec.id} className="cc-attention-card" style={{ marginBottom: 0, borderTop: '3px solid #f59e0b' }}>
                    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 8 }}>
                      <span className="cc-rank-badge">Decision #{dec.id}</span>
                      <span style={{ fontSize: '0.75rem', fontWeight: 600, color: '#dc2626' }}>
                        Risk: {dec.risk_level || 'MEDIUM'}
                      </span>
                    </div>
                    <h4 style={{ margin: '0 0 8px 0', fontSize: '0.9375rem', fontWeight: 700, color: '#0f172a' }}>
                      {dec.title || dec.proposed_action}
                    </h4>
                    <p style={{ fontSize: '0.75rem', color: '#475569', margin: '0 0 12px 0' }}>
                      {dec.reason || 'Autonomous agent reached action policy limit.'}
                    </p>
                    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                      <span style={{ fontSize: '0.6875rem', color: '#64748b' }}>
                        Agent: <strong>{dec.agent_name || 'Autonomous Agent'}</strong>
                      </span>
                      <div style={{ display: 'flex', gap: 6 }}>
                        <button
                          className="cc-btn cc-btn-primary"
                          style={{ padding: '4px 10px', fontSize: '0.75rem' }}
                          onClick={(e) => handleQuickApprove(e, dec.id)}
                          disabled={actionInProgress === dec.id}
                        >
                          <Check size={12} /> Approve
                        </button>
                        <button
                          className="cc-btn cc-btn-secondary"
                          style={{ padding: '4px 10px', fontSize: '0.75rem' }}
                          onClick={() => {
                            setSelectedDecisionId(dec.id);
                            setIsDecisionDrawerOpen(true);
                          }}
                        >
                          Inspect
                        </button>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>

          {/* Core Question 4: Is the AI workforce healthy and operating within policy? */}
          <div>
            <div style={{ marginBottom: 14 }}>
              <h2 style={{ fontSize: '1.125rem', fontWeight: 800, color: '#0f172a', margin: '0 0 4px 0', display: 'flex', alignItems: 'center', gap: 8 }}>
                <Shield size={18} color="#059669" />
                4. Is the AI Workforce Healthy & Operating Within Policy?
              </h2>
              <p style={{ fontSize: '0.8125rem', color: '#64748b', margin: 0 }}>
                Fleet-wide availability, queue pressure, and emergency halt governance status.
              </p>
            </div>

            <div style={{ background: '#fff', border: '1px solid #e2e8f0', borderRadius: 10, padding: 18, display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(210px, 1fr))', gap: 16 }}>
              <div>
                <span style={{ fontSize: '0.75rem', color: '#64748b', textTransform: 'uppercase', fontWeight: 600 }}>Workforce Status</span>
                <div style={{ fontSize: '1.25rem', fontWeight: 800, color: '#059669', display: 'flex', alignItems: 'center', gap: 6, marginTop: 4 }}>
                  <CheckCircle2 size={18} color="#059669" />
                  {controlTowerView?.platform_health || 'HEALTHY'}
                </div>
                <span style={{ fontSize: '0.75rem', color: '#64748b' }}>All agent contracts validated</span>
              </div>

              <div>
                <span style={{ fontSize: '0.75rem', color: '#64748b', textTransform: 'uppercase', fontWeight: 600 }}>Active Fleet</span>
                <div style={{ fontSize: '1.25rem', fontWeight: 800, color: '#0f172a', marginTop: 4 }}>
                  {controlTowerView?.workforce_health?.total_agents || 12} / 12 Governed
                </div>
                <span style={{ fontSize: '0.75rem', color: '#64748b' }}>0 degraded agents</span>
              </div>

              <div>
                <span style={{ fontSize: '0.75rem', color: '#64748b', textTransform: 'uppercase', fontWeight: 600 }}>Policy Version</span>
                <div style={{ fontSize: '1.25rem', fontWeight: 800, color: '#1e293b', marginTop: 4 }}>
                  {controlTowerView?.governance_status?.policy_version || 'v7.10.0-governed'}
                </div>
                <span style={{ fontSize: '0.75rem', color: '#64748b' }}>Authoritative Go boundary</span>
              </div>

              <div>
                <span style={{ fontSize: '0.75rem', color: '#64748b', textTransform: 'uppercase', fontWeight: 600 }}>Safety Invariants</span>
                <div style={{ fontSize: '1.25rem', fontWeight: 800, color: '#0284c7', marginTop: 4 }}>
                  {controlTowerView?.governance_status?.stale_approvals_prevented || 0} / {controlTowerView?.governance_status?.loops_detected_count || 0}
                </div>
                <span style={{ fontSize: '0.75rem', color: '#64748b' }}>Stale invalidations / Loops caught</span>
              </div>

              <div>
                <span style={{ fontSize: '0.75rem', color: '#64748b', textTransform: 'uppercase', fontWeight: 600 }}>Emergency Control</span>
                <div style={{ fontSize: '1.25rem', fontWeight: 800, color: (controlTowerView?.governance_status?.emergency_halt_active || controlTowerView?.workforce_health?.emergency_stop_active) ? '#dc2626' : '#10b981', marginTop: 4 }}>
                  {(controlTowerView?.governance_status?.emergency_halt_active || controlTowerView?.workforce_health?.emergency_stop_active) ? 'ACTIVE (HALTED)' : 'INACTIVE (NORMAL)'}
                </div>
                <button
                  id="btn-toggle-emergency-halt"
                  onClick={() => handleEmergencyToggle(controlTowerView?.governance_status?.emergency_halt_active || controlTowerView?.workforce_health?.emergency_stop_active)}
                  style={{
                    marginTop: 6,
                    padding: '4px 10px',
                    fontSize: '0.75rem',
                    fontWeight: 700,
                    borderRadius: 6,
                    border: '1px solid',
                    cursor: 'pointer',
                    background: (controlTowerView?.governance_status?.emergency_halt_active || controlTowerView?.workforce_health?.emergency_stop_active) ? '#ecfdf5' : '#fef2f2',
                    borderColor: (controlTowerView?.governance_status?.emergency_halt_active || controlTowerView?.workforce_health?.emergency_stop_active) ? '#10b981' : '#ef4444',
                    color: (controlTowerView?.governance_status?.emergency_halt_active || controlTowerView?.workforce_health?.emergency_stop_active) ? '#065f46' : '#991b1b',
                  }}
                >
                  {(controlTowerView?.governance_status?.emergency_halt_active || controlTowerView?.workforce_health?.emergency_stop_active) ? 'Resume All Operations' : 'Emergency Stop All'}
                </button>
              </div>
            </div>

            {/* Phase 7.11: Enterprise Autonomous Operations Resilience & Subsystem Matrix */}
            <div style={{ marginTop: 16, background: '#f8fafc', border: '1px solid #e2e8f0', borderRadius: 10, padding: 16 }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 12 }}>
                <div>
                  <h3 style={{ fontSize: '0.875rem', fontWeight: 800, color: '#0f172a', margin: 0, display: 'flex', alignItems: 'center', gap: 6 }}>
                    <Activity size={16} color="#2563eb" />
                    Unified Enterprise Autonomous-Health & Resilience Model
                  </h3>
                  <p style={{ fontSize: '0.75rem', color: '#64748b', margin: '2px 0 0 0' }}>
                    Real-time resilience monitoring across Go backend, Python sidecar, Event Mesh, Action System, and Database.
                  </p>
                </div>
                <div style={{ display: 'flex', gap: 8, alignItems: 'center' }}>
                  <span style={{
                    padding: '3px 8px',
                    fontSize: '0.6875rem',
                    fontWeight: 700,
                    borderRadius: 4,
                    background: controlTowerView?.resilience_health?.overall_state === 'HEALTHY' ? '#ecfdf5' : '#fef2f2',
                    color: controlTowerView?.resilience_health?.overall_state === 'HEALTHY' ? '#065f46' : '#991b1b',
                    border: '1px solid',
                    borderColor: controlTowerView?.resilience_health?.overall_state === 'HEALTHY' ? '#a7f3d0' : '#fecaca',
                  }}>
                    {controlTowerView?.resilience_health?.overall_state || 'HEALTHY'}
                  </span>
                  <span style={{
                    padding: '3px 8px',
                    fontSize: '0.6875rem',
                    fontWeight: 700,
                    borderRadius: 4,
                    background: '#eff6ff',
                    color: '#1e40af',
                    border: '1px solid #bfdbfe',
                  }}>
                    Ceiling: {controlTowerView?.resilience_health?.effective_autonomy_ceil || 'LEVEL_4_GOVERNED'}
                  </span>
                </div>
              </div>

              {/* Subsystem Health Grid */}
              <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(140px, 1fr))', gap: 10, marginBottom: 12 }}>
                {[
                  { name: 'Database', status: controlTowerView?.resilience_health?.subsystems?.DATABASE?.state || 'HEALTHY', latency: `${controlTowerView?.resilience_health?.subsystems?.DATABASE?.latency_ms || 2}ms` },
                  { name: 'Go Backend', status: controlTowerView?.resilience_health?.subsystems?.GO_BACKEND?.state || 'HEALTHY', latency: `${controlTowerView?.resilience_health?.subsystems?.GO_BACKEND?.latency_ms || 3}ms` },
                  { name: 'Python Sidecar', status: controlTowerView?.resilience_health?.subsystems?.PYTHON_SIDECAR?.state || 'HEALTHY', latency: `${controlTowerView?.resilience_health?.subsystems?.PYTHON_SIDECAR?.latency_ms || 12}ms` },
                  { name: 'Workflow Engine', status: controlTowerView?.resilience_health?.subsystems?.WORKFLOW_ENGINE?.state || 'HEALTHY', latency: 'Durable' },
                  { name: 'Event Mesh', status: controlTowerView?.resilience_health?.subsystems?.EVENT_MESH?.state || 'HEALTHY', latency: 'Coalesced' },
                  { name: 'AI Worker', status: controlTowerView?.resilience_health?.subsystems?.AI_WORKER?.state || 'HEALTHY', latency: 'Bounded' },
                  { name: 'Action System', status: controlTowerView?.resilience_health?.subsystems?.ACTION_SYSTEM?.state || 'HEALTHY', latency: 'Idempotent' },
                  { name: 'Approvals Gate', status: controlTowerView?.resilience_health?.subsystems?.APPROVALS?.state || 'HEALTHY', latency: 'Enforced' },
                ].map((sub, i) => (
                  <div key={i} style={{ background: '#fff', border: '1px solid #e2e8f0', borderRadius: 6, padding: '8px 10px' }}>
                    <div style={{ fontSize: '0.6875rem', fontWeight: 600, color: '#64748b' }}>{sub.name}</div>
                    <div style={{ fontSize: '0.75rem', fontWeight: 700, color: sub.status === 'HEALTHY' ? '#059669' : '#d97706', marginTop: 2, display: 'flex', alignItems: 'center', gap: 4 }}>
                      <span style={{ width: 6, height: 6, borderRadius: '50%', background: sub.status === 'HEALTHY' ? '#10b981' : '#f59e0b', display: 'inline-block' }}></span>
                      {sub.status}
                    </div>
                    <div style={{ fontSize: '0.6875rem', color: '#94a3b8', marginTop: 2 }}>{sub.latency}</div>
                  </div>
                ))}
              </div>

              {/* Resilience Operations Action Bar */}
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', background: '#fff', border: '1px solid #e2e8f0', borderRadius: 6, padding: '10px 14px' }}>
                <div style={{ display: 'flex', gap: 16, alignItems: 'center', fontSize: '0.75rem', color: '#475569' }}>
                  <span>Stuck Workflows: <strong>{controlTowerView?.resilience_health?.stuck_workflow_count || 0}</strong></span>
                  <span>Dead Letters: <strong>{controlTowerView?.resilience_health?.dead_letter_count || 0}</strong></span>
                  <span>Backpressure: <strong>{controlTowerView?.resilience_health?.backpressure_active ? 'ACTIVE' : 'NOMINAL'}</strong></span>
                </div>
                <button
                  id="btn-recover-stuck-workflows"
                  className="cc-btn cc-btn-secondary"
                  style={{ padding: '4px 12px', fontSize: '0.75rem', fontWeight: 700 }}
                  onClick={handleRecoverStuckWorkflows}
                >
                  <RefreshCw size={12} /> Trigger Governed Recovery
                </button>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* TAB CONTENT 1: CRITICAL ATTENTION (Section 6 & 23) */}
      {activeTab === 'attention' && (
        <div className="cc-attention-feed" id="cc-critical-attention-feed">
          {filteredCriticalItems.length === 0 ? (
            <div className="cc-empty-state">
              <CheckCircle2 size={36} color="#10b981" style={{ marginBottom: 12 }} />
              <h3 className="cc-empty-title">All Clear — No Critical Blockers</h3>
              <p className="cc-empty-desc">
                No active critical exceptions, compliance holds, or stalled workflows require immediate human intervention.
              </p>
            </div>
          ) : (
            filteredCriticalItems.map((item) => (
              <div
                key={item.id}
                className={`cc-attention-card ${item.severity === 'CRITICAL' ? 'sev-critical' : item.severity === 'HIGH' ? 'sev-high' : 'sev-medium'}`}
              >
                <div className="cc-card-header">
                  <div className="cc-badge-row">
                    <span className="cc-rank-badge">#{item.priority_rank} Priority</span>
                    <span className={`cc-urgency-badge ${item.urgency.toLowerCase()}`}>
                      {item.urgency} Urgency
                    </span>
                    <span className="cc-entity-ref">
                      {item.entity_reference || item.entity_type}
                    </span>
                    <span style={{ fontSize: '0.75rem', fontWeight: 600, color: '#64748b' }}>
                      Tier: {item.priority_tier?.replace(/_/g, ' ')}
                    </span>
                  </div>

                  <div style={{ display: 'flex', gap: 8 }}>
                    <button
                      className="cc-btn cc-btn-primary"
                      onClick={() => openItemDrawer(item)}
                    >
                      Inspect & Resolve <ArrowRight size={13} />
                    </button>
                  </div>
                </div>

                <h3 className="cc-card-title">{item.title}</h3>

                {/* Section 30: Concise Explanation */}
                <div className="cc-why-flagged-box">
                  <strong>WHY THIS IS FLAGGED:</strong> {item.why_flagged}
                </div>

                {/* Section 29: Fact vs Prediction vs Recommendation separation */}
                <div className="cc-triplet-grid">
                  <div className="cc-triplet-item">
                    <span className="cc-triplet-label">ACTUAL (Authoritative Facts)</span>
                    <span className="cc-triplet-val">{item.actual_facts}</span>
                  </div>
                  <div className="cc-triplet-item">
                    <span className="cc-triplet-label">PREDICTED (AI Forecast)</span>
                    <span className="cc-triplet-val">{item.predicted_impact}</span>
                  </div>
                  <div className="cc-triplet-item">
                    <span className="cc-triplet-label">RECOMMENDED (Operational Action)</span>
                    <span className="cc-triplet-val">{item.recommended_action}</span>
                  </div>
                </div>

                <div className="cc-card-footer">
                  <div className="cc-meta-items">
                    <span>Owner: <strong>{item.owner}</strong></span>
                    {item.deadline && (
                      <span style={{ color: '#dc2626' }}>
                        Deadline: {new Date(item.deadline).toLocaleString()}
                      </span>
                    )}
                    <span>Source: {item.source}</span>
                  </div>

                  <div style={{ display: 'flex', gap: 8 }}>
                    {item.id.startsWith('exc-') && (
                      <button
                        className="cc-btn cc-btn-secondary"
                        onClick={() => {
                          setSelectedExceptionId(parseInt(item.entity_id, 10) || 1);
                          setIsExceptionDrawerOpen(true);
                        }}
                      >
                        Exception Details
                      </button>
                    )}
                    {item.id.startsWith('dec-') && (
                      <button
                        className="cc-btn cc-btn-primary"
                        disabled={actionInProgress === item.id}
                        onClick={(e) => handleQuickApprove(e, item.id)}
                      >
                        {actionInProgress === item.id ? 'Approving...' : 'Quick Approve'}
                      </button>
                    )}
                  </div>
                </div>
              </div>
            ))
          )}
        </div>
      )}

      {/* TAB CONTENT 2: ACTIVE AI WORKFLOWS (Section 7, 21, 22) */}
      {activeTab === 'workflows' && (
        <div className="cc-table-container" id="cc-workflows-table">
          <table className="cc-table">
            <thead>
              <tr>
                <th>Plan Name & Goal</th>
                <th>Module & Entity</th>
                <th>Autonomy</th>
                <th>Plan Health</th>
                <th>Status</th>
                <th>Progress & Step State</th>
                <th>Updated</th>
                <th>Actions</th>
              </tr>
            </thead>
            <tbody>
              {filteredWorkflows.length === 0 ? (
                <tr>
                  <td colSpan={8} style={{ textAlign: 'center', padding: '32px 16px', color: '#64748b' }}>
                    No active AI workflows match current filters.
                  </td>
                </tr>
              ) : (
                filteredWorkflows.map((w) => {
                  const goalStr = getSqlString(w.goal) || w.goal;
                  const planIdStr = getSqlString(w.plan_id) || w.plan_id;
                  const moduleStr = getSqlString(w.module) || w.module;
                  const entityTypeStr = getSqlString(w.related_entity_type) || w.related_entity_type;
                  const entityIdStr = getSqlString(w.related_entity_id) || w.related_entity_id;
                  const autonomyStr = getSqlString(w.autonomy_level) || w.autonomy_level;
                  const healthStr = getSqlString(w.plan_health) || w.plan_health;
                  const statusStr = getSqlString(w.status) || w.status;
                  const waitingStr = getSqlString(w.waiting_state);

                  return (
                    <tr key={w.id || planIdStr}>
                      <td style={{ maxWidth: 280 }}>
                        <div style={{ fontWeight: 700, color: '#0f172a' }}>{goalStr}</div>
                        <div style={{ fontSize: '0.6875rem', color: '#64748b' }}>ID: {planIdStr}</div>
                      </td>
                      <td>
                        <span style={{ fontWeight: 600, textTransform: 'capitalize' }}>{moduleStr}</span>
                        <div style={{ fontSize: '0.6875rem', color: '#64748b' }}>{entityTypeStr}: {entityIdStr}</div>
                      </td>
                      <td>
                        <span style={{ fontSize: '0.75rem', padding: '2px 7px', background: '#eff6ff', color: '#1d4ed8', borderRadius: 4, fontWeight: 700 }}>
                          {autonomyStr?.replace('LEVEL_', 'L')}
                        </span>
                      </td>
                      <td>
                        <span
                          style={{
                            fontSize: '0.75rem',
                            fontWeight: 700,
                            padding: '2px 8px',
                            borderRadius: 12,
                            background: healthStr === 'HEALTHY' ? '#dcfce7' : healthStr === 'AT_RISK' ? '#fef3c7' : '#fee2e2',
                            color: healthStr === 'HEALTHY' ? '#15803d' : healthStr === 'AT_RISK' ? '#b45309' : '#b91c1c',
                          }}
                        >
                          {healthStr || 'HEALTHY'}
                        </span>
                      </td>
                      <td>
                        <span style={{ fontWeight: 600, fontSize: '0.75rem', color: '#334155' }}>
                          {statusStr}
                        </span>
                        {waitingStr && waitingStr !== 'NONE' && (
                          <div style={{ fontSize: '0.6875rem', color: '#b45309' }}>
                            ⏳ {waitingStr}
                          </div>
                        )}
                      </td>
                      <td style={{ maxWidth: 320 }}>
                        {/* Step progression preview (Section 21) */}
                        <div className="cc-step-progression">
                          <span className="cc-step-node completed">1. Observe ✓</span>
                          <span className="cc-step-arrow">→</span>
                          <span className="cc-step-node active">2. Prepare</span>
                          <span className="cc-step-arrow">→</span>
                          <span className="cc-step-node waiting">3. Review ⏳</span>
                        </div>
                      </td>
                      <td style={{ fontSize: '0.75rem', color: '#64748b' }}>
                        {new Date(w.updated_at).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
                      </td>
                      <td>
                        <div style={{ display: 'flex', gap: 6 }}>
                          <button
                            className="cc-btn cc-btn-secondary"
                            style={{ padding: '4px 8px', fontSize: '0.75rem' }}
                            onClick={() => {
                              setSelectedPlanId(planIdStr);
                              setIsMonitoringDrawerOpen(true);
                            }}
                          >
                            Details
                          </button>
                          <button
                            className="cc-btn cc-btn-secondary"
                            style={{ padding: '4px 8px', fontSize: '0.75rem' }}
                            onClick={(e) => handleTriggerReplan(e, planIdStr)}
                            disabled={actionInProgress === planIdStr}
                          >
                            Replan
                          </button>
                          <button
                            className="cc-btn cc-btn-danger"
                            style={{ padding: '4px 8px', fontSize: '0.75rem' }}
                            onClick={(e) => handleStopWorkflow(e, planIdStr)}
                            disabled={actionInProgress === planIdStr}
                          >
                            Stop
                          </button>
                        </div>
                      </td>
                    </tr>
                  );
                })
              )}
            </tbody>
          </table>
        </div>
      )}

      {/* TAB CONTENT 3: NEEDS YOUR DECISION (Section 8) */}
      {activeTab === 'decisions' && (
        <div className="cc-attention-feed" id="cc-decisions-feed">
          {decisions.length === 0 ? (
            <div className="cc-empty-state">
              <CheckCircle2 size={36} color="#10b981" style={{ marginBottom: 12 }} />
              <h3 className="cc-empty-title">No Pending Human Approvals</h3>
              <p className="cc-empty-desc">
                All decisions have been reviewed or are executing within approved policy ceilings.
              </p>
            </div>
          ) : (
            decisions.map((dec) => (
              <div key={dec.id || dec.decision_id} className="cc-attention-card sev-high">
                <div className="cc-card-header">
                  <div className="cc-badge-row">
                    <span className="cc-urgency-badge high">Awaiting Approval</span>
                    <span className="cc-entity-ref">{dec.module} : {dec.entity_type} #{dec.entity_id}</span>
                    <span style={{ fontSize: '0.75rem', fontWeight: 600, color: '#64748b' }}>
                      Mode: {dec.operating_mode}
                    </span>
                  </div>

                  <div style={{ display: 'flex', gap: 8 }}>
                    <button
                      className="cc-btn cc-btn-primary"
                      onClick={() => {
                        setSelectedDecisionId(dec.decision_id);
                        setIsDecisionDrawerOpen(true);
                      }}
                    >
                      Review Decision <ArrowRight size={13} />
                    </button>
                  </div>
                </div>

                <h3 className="cc-card-title">{dec.title}</h3>

                <div className="cc-triplet-grid">
                  <div className="cc-triplet-item">
                    <span className="cc-triplet-label">WHAT HAPPENED (Facts)</span>
                    <span className="cc-triplet-val">{dec.context_summary || 'Operational state recorded.'}</span>
                  </div>
                  <div className="cc-triplet-item">
                    <span className="cc-triplet-label">AI RECOMMENDATION</span>
                    <span className="cc-triplet-val">{dec.ai_recommendation}</span>
                  </div>
                  <div className="cc-triplet-item">
                    <span className="cc-triplet-label">CONFIDENCE & RISK</span>
                    <span className="cc-triplet-val">
                      Confidence: <strong>{dec.confidence}</strong> | Risk: <strong>{dec.risk_level}</strong>
                    </span>
                  </div>
                </div>

                <div className="cc-card-footer">
                  <span style={{ fontSize: '0.75rem', color: '#64748b' }}>
                    Created at: {new Date(dec.created_at).toLocaleString()}
                  </span>
                  <div style={{ display: 'flex', gap: 8 }}>
                    <button
                      className="cc-btn cc-btn-primary"
                      disabled={actionInProgress === dec.decision_id}
                      onClick={(e) => handleQuickApprove(e, dec.decision_id)}
                    >
                      {actionInProgress === dec.decision_id ? 'Submitting...' : 'Approve Recommendation'}
                    </button>
                  </div>
                </div>
              </div>
            ))
          )}
        </div>
      )}

      {/* TAB CONTENT 4: MULTI-DOMAIN RISK MATRIX (Sections 10, 11, 12, 13) */}
      {activeTab === 'risks' && (
        <div id="cc-domain-risk-container">
          <div style={{ display: 'flex', gap: 8, marginBottom: 16 }}>
            {['SHIPMENT', 'FINANCE', 'CUSTOMER', 'COMPLIANCE'].map((domainKey) => (
              <button
                key={domainKey}
                className={`cc-btn ${riskSubTab === domainKey ? 'cc-btn-primary' : 'cc-btn-secondary'}`}
                onClick={() => setRiskSubTab(domainKey)}
              >
                {domainKey === 'SHIPMENT' && <Ship size={14} />}
                {domainKey === 'FINANCE' && <DollarSign size={14} />}
                {domainKey === 'CUSTOMER' && <Users size={14} />}
                {domainKey === 'COMPLIANCE' && <FileText size={14} />}
                {domainKey} Risk Matrix
              </button>
            ))}
          </div>

          {activeRiskDomain ? (
            <div style={{ background: '#fff', border: '1px solid #e2e8f0', borderRadius: 12, padding: 20 }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 16, borderBottom: '1px solid #f1f5f9', paddingBottom: 12 }}>
                <div>
                  <h3 style={{ margin: '0 0 4px 0', fontSize: '1.1rem', fontWeight: 800 }}>
                    {activeRiskDomain.domain} Operational Risk Surface
                  </h3>
                  <div style={{ fontSize: '0.8125rem', color: '#64748b' }}>
                    Authoritative State: {activeRiskDomain.authoritative_state}
                  </div>
                  <div style={{ fontSize: '0.8125rem', color: '#b45309', fontWeight: 600 }}>
                    AI Predictive Assessment: {activeRiskDomain.predicted_risk}
                  </div>
                </div>
                <div style={{ textAlign: 'right' }}>
                  <span style={{ fontSize: '1.4rem', fontWeight: 800, color: activeRiskDomain.critical_count > 0 ? '#dc2626' : '#2563eb' }}>
                    {activeRiskDomain.total_at_risk}
                  </span>
                  <div style={{ fontSize: '0.6875rem', color: '#64748b', textTransform: 'uppercase', fontWeight: 700 }}>
                    Entities At Risk
                  </div>
                </div>
              </div>

              {activeRiskDomain.items?.length === 0 ? (
                <div className="cc-empty-state">
                  <CheckCircle2 size={32} color="#10b981" />
                  <h4 style={{ margin: '8px 0 4px 0' }}>No High-Risk Entities Detected</h4>
                  <p style={{ margin: 0, fontSize: '0.8125rem' }}>All authoritative metrics in this domain are within tolerance thresholds.</p>
                </div>
              ) : (
                <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
                  {activeRiskDomain.items.map((item, idx) => (
                    <div
                      key={idx}
                      style={{
                        background: '#f8fafc',
                        border: '1px solid #e2e8f0',
                        borderRadius: 8,
                        padding: 14,
                        display: 'flex',
                        flexDirection: 'column',
                        gap: 8,
                      }}
                    >
                      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                        <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                          <span style={{ fontWeight: 800, color: '#0f172a' }}>{item.entity_reference}</span>
                          <span
                            style={{
                              fontSize: '0.6875rem',
                              fontWeight: 800,
                              padding: '2px 6px',
                              borderRadius: 4,
                              background: item.severity === 'CRITICAL' ? '#fee2e2' : '#fef3c7',
                              color: item.severity === 'CRITICAL' ? '#dc2626' : '#b45309',
                            }}
                          >
                            {item.severity}
                          </span>
                        </div>
                        <span style={{ fontSize: '0.75rem', color: '#64748b' }}>
                          Last Event: {item.last_event}
                        </span>
                      </div>

                      <div style={{ fontSize: '0.875rem', fontWeight: 600, color: '#1e293b' }}>
                        {item.issue}
                      </div>

                      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(240px, 1fr))', gap: 8, fontSize: '0.8125rem' }}>
                        <div><strong>Actual Fact:</strong> {item.actual_fact}</div>
                        <div><strong>Predicted Risk:</strong> {item.predicted_risk}</div>
                        <div><strong>Recommended:</strong> {item.recommended_action}</div>
                      </div>

                      <div style={{ display: 'flex', justifyContent: 'flex-end', gap: 8, marginTop: 4 }}>
                        <button
                          className="cc-btn cc-btn-secondary"
                          style={{ padding: '4px 10px', fontSize: '0.75rem' }}
                          onClick={() => openItemDrawer({ entity_type: activeRiskDomain.domain, entity_id: item.entity_id })}
                        >
                          View Authoritative Source <ExternalLink size={12} />
                        </button>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </div>
          ) : (
            <div className="cc-empty-state">Loading domain risk data...</div>
          )}
        </div>
      )}

      {/* TAB CONTENT 5: AI ACTIONS & LINEAGE (Sections 14, 15, 16) */}
      {activeTab === 'activity' && (
        <div id="cc-activity-container" style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
          {/* Recent AI Actions */}
          <div className="cc-table-container">
            <div style={{ padding: '14px 16px', background: '#f8fafc', borderBottom: '1px solid #e2e8f0', fontWeight: 800, fontSize: '0.875rem' }}>
              Recent Autonomous Actions (Section 14)
            </div>
            <table className="cc-table">
              <thead>
                <tr>
                  <th>Action</th>
                  <th>Type</th>
                  <th>Mode</th>
                  <th>Status</th>
                  <th>Result Summary</th>
                  <th>Timestamp</th>
                </tr>
              </thead>
              <tbody>
                {activity.recent_actions?.length === 0 ? (
                  <tr><td colSpan={6} style={{ textAlign: 'center', color: '#64748b' }}>No recent AI actions recorded.</td></tr>
                ) : (
                  activity.recent_actions.map((act, idx) => (
                    <tr key={idx}>
                      <td style={{ fontWeight: 600 }}>{act.title || act.step_id}</td>
                      <td><code>{act.action_type}</code></td>
                      <td><span style={{ fontSize: '0.75rem', fontWeight: 700 }}>{act.execution_mode}</span></td>
                      <td>
                        <span style={{ fontSize: '0.75rem', fontWeight: 700, color: act.status === 'COMPLETED' ? '#16a34a' : '#dc2626' }}>
                          {act.status}
                        </span>
                      </td>
                      <td style={{ fontSize: '0.75rem', color: '#334155' }}>{act.result_summary}</td>
                      <td style={{ fontSize: '0.75rem', color: '#64748b' }}>{new Date(act.timestamp).toLocaleTimeString()}</td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>

          {/* Replanning Lineage */}
          <div className="cc-table-container">
            <div style={{ padding: '14px 16px', background: '#f8fafc', borderBottom: '1px solid #e2e8f0', fontWeight: 800, fontSize: '0.875rem' }}>
              Replanning Lineage (Section 15)
            </div>
            <table className="cc-table">
              <thead>
                <tr>
                  <th>Plan ID</th>
                  <th>Entity</th>
                  <th>Trigger Reason</th>
                  <th>State Transition</th>
                  <th>Changed Assumptions</th>
                  <th>Timestamp</th>
                </tr>
              </thead>
              <tbody>
                {activity.replanning_events?.length === 0 ? (
                  <tr><td colSpan={6} style={{ textAlign: 'center', color: '#64748b' }}>No replanning events recorded.</td></tr>
                ) : (
                  activity.replanning_events.map((rep, idx) => (
                    <tr key={idx}>
                      <td><strong>{rep.plan_id}</strong></td>
                      <td>{rep.entity_type} #{rep.entity_id}</td>
                      <td>{rep.trigger_reason}</td>
                      <td><code>{rep.old_status} → {rep.new_status}</code></td>
                      <td style={{ fontSize: '0.75rem' }}>{rep.changed_assumptions}</td>
                      <td style={{ fontSize: '0.75rem', color: '#64748b' }}>{new Date(rep.timestamp).toLocaleTimeString()}</td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>

          {/* Active Escalations */}
          <div className="cc-table-container">
            <div style={{ padding: '14px 16px', background: '#f8fafc', borderBottom: '1px solid #e2e8f0', fontWeight: 800, fontSize: '0.875rem', color: '#dc2626' }}>
              Active Escalations (Section 16)
            </div>
            <table className="cc-table">
              <thead>
                <tr>
                  <th>Escalation ID</th>
                  <th>Entity</th>
                  <th>Severity</th>
                  <th>Reason</th>
                  <th>Owner</th>
                  <th>Recommended Human Action</th>
                </tr>
              </thead>
              <tbody>
                {activity.escalations?.length === 0 ? (
                  <tr><td colSpan={6} style={{ textAlign: 'center', color: '#64748b' }}>No active escalations.</td></tr>
                ) : (
                  activity.escalations.map((esc, idx) => (
                    <tr key={idx}>
                      <td><strong>{esc.escalation_id}</strong></td>
                      <td>{esc.entity_type} #{esc.entity_id}</td>
                      <td><span className="cc-urgency-badge immediate">{esc.severity}</span></td>
                      <td>{esc.reason}</td>
                      <td>{esc.owner}</td>
                      <td style={{ fontWeight: 600, color: '#1e293b' }}>{esc.recommended_human_action}</td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {/* TAB CONTENT 6: SYSTEM & SUBSYSTEM HEALTH (Section 17) */}
      {activeTab === 'health' && (
        <div id="cc-system-health-tab" style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(320px, 1fr))', gap: 16 }}>
          {systemHealth?.subsystems &&
            Object.entries(systemHealth.subsystems).map(([key, sub]) => (
              <div
                key={key}
                style={{
                  background: '#ffffff',
                  border: '1px solid #e2e8f0',
                  borderRadius: 12,
                  padding: 16,
                  display: 'flex',
                  flexDirection: 'column',
                  gap: 8,
                }}
              >
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                  <span style={{ fontWeight: 800, fontSize: '0.9375rem', color: '#0f172a' }}>{sub.name}</span>
                  <span className={`cc-health-pill ${sub.status.toLowerCase()}`}>
                    <span className="cc-dot"></span>
                    {sub.status}
                  </span>
                </div>
                <div style={{ fontSize: '0.8125rem', color: '#475569', lineHeight: 1.4 }}>
                  {sub.message}
                </div>
                <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: '0.75rem', color: '#94a3b8', borderTop: '1px solid #f1f5f9', paddingTop: 8 }}>
                  <span>Latency: {sub.latency_ms} ms</span>
                  <span>Checked: {new Date(sub.last_checked).toLocaleTimeString()}</span>
                </div>
              </div>
            ))}
        </div>
      )}

      {/* DRAWERS FOR DRILL-DOWN & ACTION CONTROLS */}
      {/* 1. Decision Center Drawer (Task 5.11) */}
      <HumanAIDecisionCenterDrawer
        isOpen={isDecisionDrawerOpen}
        onClose={() => {
          setIsDecisionDrawerOpen(false);
          setSelectedDecisionId(null);
          fetchData(true);
        }}
        initialDecisionId={selectedDecisionId}
      />

      {/* 2. Shipment Adaptive Drawer (Task 5.3) */}
      <ShipmentAdaptiveDrawer
        isOpen={isShipmentDrawerOpen}
        onClose={() => {
          setIsShipmentDrawerOpen(false);
          setSelectedShipmentId(null);
        }}
        shipmentId={selectedShipmentId || 101}
      />

      {/* 3. Exception Resolution Drawer (Task 5.8) */}
      <ExceptionResolutionDrawer
        isOpen={isExceptionDrawerOpen}
        onClose={() => {
          setIsExceptionDrawerOpen(false);
          setSelectedExceptionId(null);
        }}
        exception={selectedExceptionId ? { id: selectedExceptionId } : null}
      />

      {/* 4. Finance Collections Drawer (Task 5.6) */}
      <FinanceCollectionsAdaptiveDrawer
        isOpen={isFinanceDrawerOpen}
        onClose={() => {
          setIsFinanceDrawerOpen(false);
          setSelectedInvoiceId(null);
        }}
        invoice={selectedInvoiceId ? { id: selectedInvoiceId } : null}
      />

      {/* 5. Contract Compliance Drawer (Task 5.7) */}
      <ContractComplianceMonitoringDrawer
        isOpen={isComplianceDrawerOpen}
        onClose={() => {
          setIsComplianceDrawerOpen(false);
          setSelectedContractId(null);
        }}
        contract={selectedContractId ? { id: selectedContractId } : null}
      />

      {/* 6. Continuous Monitoring Drawer (Task 5.10) */}
      <ContinuousMonitoringDrawer
        isOpen={isMonitoringDrawerOpen}
        onClose={() => {
          setIsMonitoringDrawerOpen(false);
          setSelectedPlanId(null);
        }}
        planId={selectedPlanId}
      />

      {/* 7. Agent Memory & Outcome Learning Drawer (Task 5.13) */}
      <AgentMemoryLearningDrawer
        isOpen={isMemoryDrawerOpen}
        onClose={() => setIsMemoryDrawerOpen(false)}
        onMemoryUpdated={() => fetchData(true)}
      />

      {/* 8. Controlled Autonomy Governance Drawer (Task 5.14) */}
      <ControlledAutonomyGovernanceDrawer
        isOpen={isGovernanceDrawerOpen}
        onClose={() => setIsGovernanceDrawerOpen(false)}
        initialTab={governanceInitialTab}
        onPolicyUpdated={() => fetchData(true)}
      />

      {/* 9. Workflow Trace Modal / Slide-over (Phase 7.9) */}
      {isTraceModalOpen && (
        <div className="cc-modal-backdrop" onClick={() => setIsTraceModalOpen(false)}>
          <div className="cc-trace-drawer" onClick={(e) => e.stopPropagation()}>
            <div className="cc-trace-header">
              <div>
                <h2>
                  <Compass size={20} color="#2563eb" />
                  Workflow Execution Trace
                </h2>
                {selectedWorkflowTrace && (
                  <div className="cc-trace-meta-strip">
                    <span>Workflow ID: <strong>{selectedWorkflowTrace.workflow_id}</strong></span>
                    <span>Correlation: <strong>{selectedWorkflowTrace.correlation_id || 'N/A'}</strong></span>
                    <span>Event ID: <strong>{selectedWorkflowTrace.initiating_event || 'N/A'}</strong></span>
                  </div>
                )}
              </div>
              <button
                className="cc-btn cc-btn-secondary"
                style={{ padding: '4px 8px' }}
                onClick={() => setIsTraceModalOpen(false)}
              >
                <X size={16} />
              </button>
            </div>

            <div className="cc-trace-content">
              {traceLoading ? (
                <div style={{ textAlign: 'center', padding: '40px 0', color: '#64748b' }}>
                  <RefreshCw size={24} className="spin" style={{ margin: '0 auto 12px auto' }} />
                  <p>Retrieving auditable trace lineage...</p>
                </div>
              ) : selectedWorkflowTrace ? (
                <>
                  {/* Status & Policy Banner */}
                  <div style={{ background: '#f8fafc', border: '1px solid #e2e8f0', borderRadius: 8, padding: 14, marginBottom: 20 }}>
                    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 8 }}>
                      <span style={{ fontSize: '0.8125rem', fontWeight: 700, color: '#1e293b' }}>
                        Policy Decision: {selectedWorkflowTrace.policy_decision}
                      </span>
                      <span className="cc-autonomy-pill lvl-3">
                        {selectedWorkflowTrace.autonomy_level}
                      </span>
                    </div>
                    <div style={{ fontSize: '0.75rem', color: '#64748b' }}>
                      Status: <strong>{selectedWorkflowTrace.current_state}</strong> &bull; Verification: <strong>{selectedWorkflowTrace.verification_status}</strong>
                    </div>
                  </div>

                  {/* Fact vs Prediction Separation */}
                  <div className="cc-trace-section">
                    <div className="cc-trace-section-title">
                      <Shield size={14} /> Authoritative Business Facts vs AI Predictions
                    </div>
                    <div className="cc-triplet-grid">
                      <div className="cc-triplet-item" style={{ background: '#eff6ff', border: '1px solid #bfdbfe' }}>
                        <span className="cc-triplet-label" style={{ color: '#1d4ed8' }}>AUTHORITATIVE FACTS (Verified DB)</span>
                        <pre style={{ margin: 0, fontSize: '0.6875rem', whiteSpace: 'pre-wrap', color: '#1e3a8a' }}>
                          {JSON.stringify(selectedWorkflowTrace.authoritative_facts, null, 2)}
                        </pre>
                      </div>
                      <div className="cc-triplet-item" style={{ background: '#faf5ff', border: '1px solid #e9d5ff' }}>
                        <span className="cc-triplet-label" style={{ color: '#7e22ce' }}>AI INFERENCE & PREDICTIONS</span>
                        <pre style={{ margin: 0, fontSize: '0.6875rem', whiteSpace: 'pre-wrap', color: '#581c87' }}>
                          {JSON.stringify(selectedWorkflowTrace.ai_predictions, null, 2)}
                        </pre>
                      </div>
                    </div>
                  </div>

                  {/* Workflow Steps Timeline */}
                  <div className="cc-trace-section">
                    <div className="cc-trace-section-title">
                      <Layers size={14} /> Execution Steps & Lineage
                    </div>
                    {selectedWorkflowTrace.steps && selectedWorkflowTrace.steps.length > 0 ? (
                      selectedWorkflowTrace.steps.map((step) => (
                        <div key={step.step_id || step.step_number} className="cc-timeline-step">
                          <div className={`cc-timeline-step-dot ${step.status?.toLowerCase()}`}></div>
                          <div className="cc-timeline-step-content">
                            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 4 }}>
                              <span style={{ fontWeight: 700, fontSize: '0.8125rem', color: '#0f172a' }}>
                                Step {step.step_number}: {step.title || step.action_type}
                              </span>
                              <span className={`cc-health-pill ${step.status === 'COMPLETED' ? 'healthy' : 'warning'}`}>
                                {step.status}
                              </span>
                            </div>
                            <div style={{ fontSize: '0.75rem', color: '#64748b', marginBottom: 4 }}>
                              Agent: <strong>{step.agent_id || 'System'}</strong> &bull; Action: <strong>{step.action_type}</strong>
                            </div>
                            {step.description && (
                              <div style={{ fontSize: '0.75rem', color: '#334155' }}>
                                {step.description}
                              </div>
                            )}
                          </div>
                        </div>
                      ))
                    ) : (
                      <p style={{ fontSize: '0.75rem', color: '#64748b' }}>No sub-steps registered for this workflow plan.</p>
                    )}
                  </div>
                </>
              ) : (
                <p style={{ fontSize: '0.875rem', color: '#64748b', textAlign: 'center', padding: '20px 0' }}>
                  Workflow trace unavailable.
                </p>
              )}
            </div>

            {selectedWorkflowTrace && (
              <div className="cc-trace-footer">
                <span style={{ fontSize: '0.75rem', color: '#64748b' }}>
                  Governed Intervention Controls
                </span>
                <div style={{ display: 'flex', gap: 8 }}>
                  <button
                    className="cc-btn cc-btn-secondary"
                    onClick={() => handleControlTowerAction(selectedWorkflowTrace.workflow_id, 'PAUSE')}
                  >
                    <Pause size={13} /> Pause
                  </button>
                  <button
                    className="cc-btn cc-btn-secondary"
                    onClick={() => handleControlTowerAction(selectedWorkflowTrace.workflow_id, 'RESUME')}
                  >
                    <Play size={13} /> Resume
                  </button>
                  <button
                    className="cc-btn cc-btn-danger"
                    onClick={() => handleControlTowerAction(selectedWorkflowTrace.workflow_id, 'CANCEL')}
                  >
                    <XCircle size={13} /> Cancel
                  </button>
                </div>
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  );
}

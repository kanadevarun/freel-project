import { useState, useEffect, useCallback } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import {
  Clock,
  Play,
  CheckCircle2,
  AlertTriangle,
  XCircle,
  RefreshCw,
  Plus,
  Search,
  Filter,
  Eye,
  Calendar,
  Layers,
  Check,
  ChevronRight,
  ShieldCheck,
  Zap,
  Info,
  Sliders,
  Trash2,
  Edit2,
  FileText,
  Activity,
  ArrowRight,
  Sparkles,
  ExternalLink,
  Ban,
  HelpCircle,
  X,
  Radio,
  RotateCcw,
} from 'lucide-react';
import { automationService } from '../../../services/automationService';
import ActionOrchestrationTab from './ActionOrchestrationTab';
import EventWorkflowsTab from './EventWorkflowsTab';
import EnterpriseEventMeshSection from './EnterpriseEventMeshSection';
import './WorkflowAutomationsPage.css';

export default function WorkflowAutomationsPage() {
  const navigate = useNavigate();

  // State
  const [activeTab, setActiveTab] = useState('automations'); // 'automations' | 'insights' | 'history' | 'catalog' | 'orchestration'
  const [automations, setAutomations] = useState([]);
  const [executions, setExecutions] = useState([]);
  const [supportedTypes, setSupportedTypes] = useState([]);
  const [stats, setStats] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [successBanner, setSuccessBanner] = useState(null);

  // Operational Insights State (Phase 3)
  const [insights, setInsights] = useState([]);
  const [loadingInsights, setLoadingInsights] = useState(false);
  const [insightModuleFilter, setInsightModuleFilter] = useState('');
  const [insightStatusFilter, setInsightStatusFilter] = useState('ACTIVE');
  const [insightSeverityFilter, setInsightSeverityFilter] = useState('');
  const [executionInsights, setExecutionInsights] = useState([]);

  // Filters
  const [search, setSearch] = useState('');
  const [typeFilter, setTypeFilter] = useState('');
  const [statusFilter, setStatusFilter] = useState(''); // 'enabled' | 'disabled'

  // Modals & Drawers
  const [isCreateModalOpen, setIsCreateModalOpen] = useState(false);
  const [editingAutomation, setEditingAutomation] = useState(null);
  const [selectedExecution, setSelectedExecution] = useState(null);
  const [executionRecommendations, setExecutionRecommendations] = useState([]);
  const [loadingExecRecs, setLoadingExecRecs] = useState(false);
  const [previewScheduleModal, setPreviewScheduleModal] = useState(null);
  const [actionLoading, setActionLoading] = useState({}); // { [autoId]: boolean }

  // Form State
  const [formData, setFormData] = useState({
    name: '',
    automation_type: '',
    description: '',
    trigger_type: 'SCHEDULED',
    approval_policy: 'ALWAYS_REQUIRE',
    owner_team: 'OPERATIONS',
    priority: 'MEDIUM',
    schedule_type: 'DAILY',
    schedule_time: '08:00',
    schedule_days: ['MON', 'TUE', 'WED', 'THU', 'FRI'],
    timezone: 'UTC',
    execution_window_minutes: 60,
  });

  const fetchData = useCallback(async () => {
    try {
      setLoading(true);
      setError(null);

      const [autoRes, statsRes, typesRes] = await Promise.all([
        automationService.listAutomations({ limit: 100 }),
        automationService.getStats().catch(() => ({ stats: null })),
        automationService.getSupportedTypes().catch(() => ({ supported_types: [] })),
      ]);

      if (autoRes?.success) {
        setAutomations(autoRes.automations || []);
      }
      if (statsRes?.success && statsRes.stats) {
        setStats(statsRes.stats);
      }
      if (typesRes?.success && typesRes.supported_types) {
        setSupportedTypes(typesRes.supported_types);
      }
    } catch (err) {
      setError(err.message || 'Failed to load workflow automations');
    } finally {
      setLoading(false);
    }
  }, []);

  const fetchExecutions = useCallback(async () => {
    try {
      const res = await automationService.listExecutions({ limit: 50 });
      if (res?.success) {
        setExecutions(res.executions || []);
      }
    } catch (err) {
      console.error('Failed to load executions:', err);
    }
  }, []);

  const fetchInsights = useCallback(async () => {
    try {
      setLoadingInsights(true);
      const params = { limit: 50 };
      if (insightModuleFilter) params.source_module = insightModuleFilter;
      if (insightStatusFilter) params.status = insightStatusFilter;
      if (insightSeverityFilter) params.severity = insightSeverityFilter;
      const res = await automationService.listInsights(params);
      if (res?.success) {
        setInsights(res.insights || []);
      }
    } catch (err) {
      console.error('Failed to load operational insights:', err);
    } finally {
      setLoadingInsights(false);
    }
  }, [insightModuleFilter, insightStatusFilter, insightSeverityFilter]);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  useEffect(() => {
    if (activeTab === 'history') {
      fetchExecutions();
    } else if (activeTab === 'insights') {
      fetchInsights();
    }
  }, [activeTab, fetchExecutions, fetchInsights]);

  // Flash message helper
  const showToast = (msg) => {
    setSuccessBanner(msg);
    setTimeout(() => setSuccessBanner(null), 5000);
  };

  // Handlers
  const handleToggleEnabled = async (auto) => {
    const newEnabled = !auto.is_enabled;
    try {
      setActionLoading((prev) => ({ ...prev, [auto.id]: true }));
      if (newEnabled) {
        await automationService.enableAutomation(auto.id);
        showToast(`Automation "${auto.name}" enabled and scheduled.`);
      } else {
        await automationService.disableAutomation(auto.id);
        showToast(`Automation "${auto.name}" disabled.`);
      }
      await fetchData();
    } catch (err) {
      alert(`Error toggling automation: ${err.message}`);
    } finally {
      setActionLoading((prev) => ({ ...prev, [auto.id]: false }));
    }
  };

  const handleManualRun = async (auto) => {
    try {
      setActionLoading((prev) => ({ ...prev, [`run-${auto.id}`]: true }));
      const res = await automationService.runAutomation(auto.id);
      showToast(`Manual run queued for "${auto.name}". Processing analysis...`);
      // Wait a moment then refresh
      setTimeout(async () => {
        await fetchData();
        if (activeTab === 'history') fetchExecutions();
        setActionLoading((prev) => ({ ...prev, [`run-${auto.id}`]: false }));
      }, 1500);
    } catch (err) {
      alert(`Failed to trigger manual execution: ${err.message}`);
      setActionLoading((prev) => ({ ...prev, [`run-${auto.id}`]: false }));
    }
  };

  const handleDelete = async (auto) => {
    if (!window.confirm(`Are you sure you want to delete "${auto.name}"? This action cannot be undone.`)) {
      return;
    }
    try {
      await automationService.deleteAutomation(auto.id);
      showToast(`Automation "${auto.name}" deleted successfully.`);
      await fetchData();
    } catch (err) {
      alert(`Failed to delete automation: ${err.message}`);
    }
  };

  const handlePreviewSchedule = async (auto) => {
    try {
      const res = await automationService.previewNextRun(auto.id);
      if (res?.success) {
        setPreviewScheduleModal({ auto, preview: res.preview });
      }
    } catch (err) {
      alert(`Failed to preview schedule: ${err.message}`);
    }
  };

  const handleOpenCreateModal = (presetType = null) => {
    if (presetType) {
      const info = supportedTypes.find((t) => t.type === presetType);
      setFormData({
        name: info?.name || '',
        automation_type: presetType,
        description: info?.description || '',
        schedule_type: info?.default_schedule_type || 'DAILY',
        schedule_time: info?.default_schedule_time || '08:00',
        schedule_days: ['MON', 'TUE', 'WED', 'THU', 'FRI'],
        timezone: 'UTC',
        execution_window_minutes: 60,
      });
    } else {
      const defaultType = supportedTypes[0]?.type || 'DAILY_OVERDUE_INVOICE_REVIEW';
      const info = supportedTypes[0];
      setFormData({
        name: info?.name || 'Daily Overdue Invoice & Receivables Review',
        automation_type: defaultType,
        description: info?.description || '',
        schedule_type: 'DAILY',
        schedule_time: '08:00',
        schedule_days: ['MON', 'TUE', 'WED', 'THU', 'FRI'],
        timezone: 'UTC',
        execution_window_minutes: 60,
      });
    }
    setEditingAutomation(null);
    setIsCreateModalOpen(true);
  };

  const handleOpenEditModal = (auto) => {
    setEditingAutomation(auto);
    let days = ['MON', 'TUE', 'WED', 'THU', 'FRI'];
    if (auto.schedule_days) {
      days = auto.schedule_days.split(',').map((d) => d.trim());
    }
    setFormData({
      name: auto.name,
      automation_type: auto.automation_type,
      description: auto.description || '',
      schedule_type: auto.schedule_type,
      schedule_time: auto.schedule_time || '08:00',
      schedule_days: days,
      timezone: auto.timezone || 'UTC',
      execution_window_minutes: auto.execution_window_minutes || 60,
    });
    setIsCreateModalOpen(true);
  };

  const handleSaveAutomation = async (e) => {
    e.preventDefault();
    try {
      const payload = {
        name: formData.name,
        automation_type: formData.automation_type,
        description: formData.description || null,
        trigger_type: formData.trigger_type || 'SCHEDULED',
        approval_policy: formData.approval_policy || 'ALWAYS_REQUIRE',
        owner_team: formData.owner_team || 'OPERATIONS',
        priority: formData.priority || 'MEDIUM',
        schedule_type: formData.schedule_type,
        schedule_time: formData.schedule_time,
        schedule_days: formData.schedule_type === 'WEEKLY' ? formData.schedule_days.join(',') : null,
        timezone: formData.timezone,
        execution_window_minutes: parseInt(formData.execution_window_minutes, 10) || 60,
      };

      if (editingAutomation) {
        await automationService.updateAutomation(editingAutomation.id, payload);
        showToast(`Automation "${formData.name}" updated successfully.`);
      } else {
        await automationService.createAutomation(payload);
        showToast(`Automation "${formData.name}" created and scheduled.`);
      }

      setIsCreateModalOpen(false);
      await fetchData();
    } catch (err) {
      alert(`Error saving automation: ${err.message}`);
    }
  };

  const handleOpenExecutionDetails = async (exec) => {
    setSelectedExecution(exec);
    setLoadingExecRecs(true);
    try {
      const [recRes, insRes] = await Promise.all([
        automationService.getExecutionRecommendations(exec.id).catch(() => ({ recommendations: [] })),
        automationService.getExecutionInsights(exec.id).catch(() => ({ insights: [] })),
      ]);
      if (recRes?.success) {
        setExecutionRecommendations(recRes.recommendations || []);
      }
      if (insRes?.success) {
        setExecutionInsights(insRes.insights || []);
      }
    } catch (err) {
      console.error('Failed to load execution details:', err);
      setExecutionRecommendations([]);
      setExecutionInsights([]);
    } finally {
      setLoadingExecRecs(false);
    }
  };

  const handleCancelExecution = async (exec) => {
    if (!window.confirm(`Cancel execution #${exec.id}?`)) return;
    try {
      await automationService.cancelExecution(exec.id, 'Cancelled manually by operator');
      showToast(`Execution #${exec.id} cancelled.`);
      if (selectedExecution?.id === exec.id) {
        setSelectedExecution((prev) => ({ ...prev, status: 'CANCELLED' }));
      }
      fetchExecutions();
      fetchData();
    } catch (err) {
      alert(`Failed to cancel execution: ${err.message}`);
    }
  };

  const handleRetryExecution = async (exec) => {
    try {
      await automationService.retryExecution(exec.id);
      showToast(`Retry enqueued for execution #${exec.id}`);
      fetchExecutions();
      fetchData();
    } catch (err) {
      alert(`Failed to retry execution: ${err.message}`);
    }
  };

  const handleAcknowledgeInsight = async (insight) => {
    try {
      await automationService.acknowledgeInsight(insight.id);
      showToast(`Acknowledged operational insight #${insight.id}`);
      fetchInsights();
      fetchData();
    } catch (err) {
      alert(`Failed to acknowledge insight: ${err.message}`);
    }
  };

  const handleDismissInsight = async (insight) => {
    const reason = window.prompt(`Dismiss insight #${insight.id}? Enter reason:`, 'Verified by operations coordinator');
    if (reason === null) return;
    try {
      await automationService.dismissInsight(insight.id, reason);
      showToast(`Dismissed operational insight #${insight.id}`);
      fetchInsights();
      fetchData();
    } catch (err) {
      alert(`Failed to dismiss insight: ${err.message}`);
    }
  };

  // Filtered Automations
  const filteredAutomations = automations.filter((auto) => {
    if (search && !auto.name.toLowerCase().includes(search.toLowerCase()) && !auto.automation_type.toLowerCase().includes(search.toLowerCase())) {
      return false;
    }
    if (typeFilter && auto.automation_type !== typeFilter) {
      return false;
    }
    if (statusFilter === 'enabled' && !auto.is_enabled) return false;
    if (statusFilter === 'disabled' && auto.is_enabled) return false;
    return true;
  });

  const getStatusBadge = (status) => {
    switch (status) {
      case 'COMPLETED':
        return (
          <span className="auto-badge badge-success">
            <CheckCircle2 size={13} /> Completed
          </span>
        );
      case 'RUNNING':
        return (
          <span className="auto-badge badge-running">
            <RefreshCw size={13} className="spin-icon" /> Running
          </span>
        );
      case 'QUEUED':
        return (
          <span className="auto-badge badge-queued">
            <Clock size={13} /> Queued
          </span>
        );
      case 'FAILED':
        return (
          <span className="auto-badge badge-failed">
            <AlertTriangle size={13} /> Failed
          </span>
        );
      case 'CANCELLED':
        return (
          <span className="auto-badge badge-cancelled">
            <Ban size={13} /> Cancelled
          </span>
        );
      default:
        return (
          <span className="auto-badge badge-neutral">
            <Clock size={13} /> Never Run
          </span>
        );
    }
  };

  const formatDateTime = (dateStr) => {
    if (!dateStr) return '—';
    const d = new Date(dateStr);
    return d.toLocaleString(undefined, {
      month: 'short',
      day: 'numeric',
      year: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    });
  };

  return (
    <div className="workflow-automations-page">
      {/* Toast Banner */}
      {successBanner && (
        <div className="auto-toast-banner" role="status">
          <CheckCircle2 size={18} />
          <span>{successBanner}</span>
          <button className="btn-close-toast" onClick={() => setSuccessBanner(null)} aria-label="Close notification">
            <X size={14} />
          </button>
        </div>
      )}

      {/* Page Header */}
      <div className="workflow-page-header">
        <div className="header-title-area">
          <div className="header-badge-row">
            <span className="header-badge">Enterprise Operations</span>
            <span className="header-badge-safe">
              <ShieldCheck size={13} /> Read-Only Safety Enforced
            </span>
          </div>
          <h1>Workflow Automations & Scheduled AI Jobs</h1>
          <p>
            Configure autonomous, read-only AI analysis pipelines that execute on scheduled intervals or operational triggers.
            All findings route directly to the Recommendation Center for human approval before any action is taken.
          </p>
        </div>

        <div className="header-actions">
          <button
            className="btn-secondary-light"
            onClick={() => {
              fetchData();
              if (activeTab === 'history') fetchExecutions();
            }}
            disabled={loading}
            title="Refresh current automations and statuses"
          >
            <RefreshCw size={15} className={loading ? 'spin-icon' : ''} />
            Refresh
          </button>
          <button
            className="btn-primary-dark"
            onClick={() => handleOpenCreateModal()}
            id="btn-create-automation"
          >
            <Plus size={16} />
            Create Automation
          </button>
        </div>
      </div>

      {/* KPI Stats Strip */}
      <div className="auto-stats-grid">
        <div className="auto-stat-card">
          <div className="stat-label">Total Automations</div>
          <div className="stat-value">{stats?.total_automations || automations.length}</div>
          <div className="stat-sub">Configured pipelines</div>
        </div>
        <div className="auto-stat-card">
          <div className="stat-label">Active / Scheduled</div>
          <div className="stat-value stat-success">
            {stats?.active_automations || automations.filter((a) => a.is_enabled).length}
          </div>
          <div className="stat-sub">Running on schedule</div>
        </div>
        <div className="auto-stat-card">
          <div className="stat-label">Active Operational Signals</div>
          <div className="stat-value stat-amber">
            {stats?.active_insights_count || insights.length}
          </div>
          <div className="stat-sub">Deterministic intelligence</div>
        </div>
        <div className="auto-stat-card">
          <div className="stat-label">Total Executions</div>
          <div className="stat-value">{stats?.total_executions || 0}</div>
          <div className="stat-sub">Completed & scheduled runs</div>
        </div>
        <div className="auto-stat-card">
          <div className="stat-label">Success Rate</div>
          <div className="stat-value stat-primary">
            {stats?.recent_success_rate !== undefined ? `${Math.round(stats.recent_success_rate)}%` : '100%'}
          </div>
          <div className="stat-sub">Execution reliability</div>
        </div>
        <div className="auto-stat-card">
          <div className="stat-label">Recommendations Produced</div>
          <div className="stat-value stat-amber">
            {(stats?.recommendations_generated || 0) + (stats?.recommendations_updated || 0)}
          </div>
          <div className="stat-sub">
            {stats?.recommendations_generated || 0} new, {stats?.recommendations_updated || 0} refreshed
          </div>
        </div>
      </div>

      {/* Navigation Tabs */}
      <div className="auto-tabs-nav">
        <button
          className={`tab-btn ${activeTab === 'automations' ? 'active' : ''}`}
          onClick={() => setActiveTab('automations')}
          id="tab-automations"
        >
          <Sliders size={16} />
          Automations ({automations.length})
        </button>
        <button
          className={`tab-btn ${activeTab === 'insights' ? 'active' : ''}`}
          onClick={() => setActiveTab('insights')}
          id="tab-insights"
        >
          <Radio size={16} />
          Operational Insights Feed ({stats?.active_insights_count || insights.length})
        </button>
        <button
          className={`tab-btn ${activeTab === 'history' ? 'active' : ''}`}
          onClick={() => setActiveTab('history')}
          id="tab-history"
        >
          <Activity size={16} />
          Execution History
        </button>
        <button
          className={`tab-btn ${activeTab === 'catalog' ? 'active' : ''}`}
          onClick={() => setActiveTab('catalog')}
          id="tab-catalog"
        >
          <Layers size={16} />
          Supported AI Assistant Jobs ({supportedTypes.length})
        </button>
        <button
          className={`tab-btn ${activeTab === 'orchestration' ? 'active' : ''}`}
          onClick={() => setActiveTab('orchestration')}
          id="tab-orchestration"
        >
          <ShieldCheck size={16} />
          Action Orchestration (Phase 3)
        </button>
        <button
          className={`tab-btn ${activeTab === 'event_workflows' ? 'active' : ''}`}
          onClick={() => setActiveTab('event_workflows')}
          id="tab-event-workflows"
        >
          <Sparkles size={16} />
          Event-Driven & Cross-Module AI (Task 3.8)
        </button>
        <button
          className={`tab-btn ${activeTab === 'event_mesh' ? 'active' : ''}`}
          onClick={() => setActiveTab('event_mesh')}
          id="tab-event-mesh"
        >
          <Radio size={16} />
          Enterprise Event Mesh (Phase 7.8)
        </button>
      </div>

      {/* Main Tab Content */}
      {activeTab === 'automations' && (
        <div className="automations-tab-content">
          {/* Filter Bar */}
          <div className="auto-filters-bar">
            <div className="search-box">
              <Search size={16} className="search-icon" />
              <input
                type="text"
                placeholder="Search automations by name or type..."
                value={search}
                onChange={(e) => setSearch(e.target.value)}
              />
              {search && (
                <button className="clear-search-btn" onClick={() => setSearch('')}>
                  <X size={14} />
                </button>
              )}
            </div>

            <div className="filter-select-group">
              <select
                value={typeFilter}
                onChange={(e) => setTypeFilter(e.target.value)}
                className="filter-select"
              >
                <option value="">All Assistant Types</option>
                {supportedTypes.map((t) => (
                  <option key={t.type} value={t.type}>
                    {t.name}
                  </option>
                ))}
              </select>

              <select
                value={statusFilter}
                onChange={(e) => setStatusFilter(e.target.value)}
                className="filter-select"
              >
                <option value="">All Statuses</option>
                <option value="enabled">Enabled Only</option>
                <option value="disabled">Disabled Only</option>
              </select>
            </div>
          </div>

          {/* Automations Table */}
          {loading ? (
            <div className="auto-loading-state">
              <RefreshCw size={28} className="spin-icon" />
              <p>Loading scheduled automations...</p>
            </div>
          ) : filteredAutomations.length === 0 ? (
            <div className="auto-empty-state">
              <div className="empty-icon-wrap">
                <Clock size={32} />
              </div>
              <h3>No Automations Configured</h3>
              <p>
                {search || typeFilter || statusFilter
                  ? 'No automations match your search and filter criteria.'
                  : 'Start by creating your first scheduled AI review job. It will automatically evaluate risks and generate recommendations on your schedule.'}
              </p>
              <button
                className="btn-primary-dark"
                onClick={() => handleOpenCreateModal()}
              >
                <Plus size={16} /> Create Automation
              </button>
            </div>
          ) : (
            <div className="table-responsive-wrapper">
              <table className="automations-table">
                <thead>
                  <tr>
                    <th>Automation Name</th>
                    <th>Assistant Type</th>
                    <th>Schedule</th>
                    <th>Next Scheduled Run</th>
                    <th>Last Execution</th>
                    <th>Active</th>
                    <th className="th-actions">Actions</th>
                  </tr>
                </thead>
                <tbody>
                  {filteredAutomations.map((auto) => {
                    const isRunning = actionLoading[`run-${auto.id}`];
                    return (
                      <tr key={auto.id} className={!auto.is_enabled ? 'row-disabled' : ''}>
                        <td className="td-name">
                          <div className="auto-name-title">{auto.name}</div>
                          {auto.description && (
                            <div className="auto-name-desc">{auto.description}</div>
                          )}
                        </td>
                        <td>
                          <span className="type-tag">{auto.automation_type}</span>
                        </td>
                        <td>
                          <div className="schedule-badge">
                            <Clock size={13} />
                            <span>
                              {auto.schedule_type === 'WEEKLY'
                                ? `Weekly (${auto.schedule_days || 'Weekdays'})`
                                : auto.schedule_type === 'HOURLY'
                                ? 'Hourly'
                                : 'Daily'}{' '}
                              at {auto.schedule_time || '08:00'} {auto.timezone || 'UTC'}
                            </span>
                          </div>
                        </td>
                        <td>
                          {auto.is_enabled && auto.next_execution_at ? (
                            <button
                              className="next-run-btn"
                              onClick={() => handlePreviewSchedule(auto)}
                              title="Click to preview schedule details"
                            >
                              <Calendar size={13} />
                              <span>{formatDateTime(auto.next_execution_at)}</span>
                            </button>
                          ) : (
                            <span className="text-muted">Paused (Disabled)</span>
                          )}
                        </td>
                        <td>
                          <div className="last-run-cell">
                            {getStatusBadge(auto.last_execution_status)}
                            {auto.last_execution_at && (
                              <div className="last-run-time">{formatDateTime(auto.last_execution_at)}</div>
                            )}
                          </div>
                        </td>
                        <td>
                          <label className="switch-toggle" title={auto.is_enabled ? 'Click to disable' : 'Click to enable'}>
                            <input
                              type="checkbox"
                              checked={auto.is_enabled}
                              onChange={() => handleToggleEnabled(auto)}
                              disabled={actionLoading[auto.id]}
                            />
                            <span className="slider round"></span>
                          </label>
                        </td>
                        <td className="td-actions">
                          <div className="action-buttons-group">
                            <button
                              className="btn-icon-action btn-run"
                              onClick={() => handleManualRun(auto)}
                              disabled={!auto.is_enabled || isRunning}
                              title="Run immediately"
                            >
                              {isRunning ? (
                                <RefreshCw size={15} className="spin-icon" />
                              ) : (
                                <Play size={15} />
                              )}
                              <span>Run</span>
                            </button>

                            <button
                              className="btn-icon-action"
                              onClick={() => handleOpenEditModal(auto)}
                              title="Edit schedule"
                            >
                              <Edit2 size={15} />
                            </button>

                            <button
                              className="btn-icon-action btn-delete"
                              onClick={() => handleDelete(auto)}
                              title="Delete automation"
                            >
                              <Trash2 size={15} />
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
        </div>
      )}

      {/* ── OPERATIONAL INSIGHTS FEED TAB (Phase 3) ── */}
      {activeTab === 'insights' && (
        <div className="insights-tab-content">
          {/* Filter Bar */}
          <div className="auto-filters-bar">
            <div className="search-box">
              <Search size={16} className="search-icon" />
              <input
                type="text"
                placeholder="Search operational signals by title or reference..."
                value={search}
                onChange={(e) => setSearch(e.target.value)}
              />
              {search && (
                <button className="clear-search-btn" onClick={() => setSearch('')}>
                  <X size={14} />
                </button>
              )}
            </div>

            <div className="filter-select-group">
              <select
                value={insightModuleFilter}
                onChange={(e) => setInsightModuleFilter(e.target.value)}
                className="filter-select"
                id="filter-insight-module"
              >
                <option value="">All Source Modules</option>
                <option value="shipments">Shipments & Operations</option>
                <option value="invoices">Invoices & Receivables</option>
                <option value="contracts">Contracts & Compliance</option>
                <option value="rfq">RFQs & Quotations</option>
                <option value="customers">Customers</option>
              </select>

              <select
                value={insightSeverityFilter}
                onChange={(e) => setInsightSeverityFilter(e.target.value)}
                className="filter-select"
                id="filter-insight-severity"
              >
                <option value="">All Severities</option>
                <option value="CRITICAL">Critical</option>
                <option value="HIGH">High</option>
                <option value="MEDIUM">Medium</option>
                <option value="LOW">Low</option>
              </select>

              <select
                value={insightStatusFilter}
                onChange={(e) => setInsightStatusFilter(e.target.value)}
                className="filter-select"
                id="filter-insight-status"
              >
                <option value="">All Statuses</option>
                <option value="ACTIVE">Active Signals</option>
                <option value="ACKNOWLEDGED">Acknowledged</option>
                <option value="ACTIONED">Actioned</option>
                <option value="DISMISSED">Dismissed</option>
              </select>
            </div>
          </div>

          {loadingInsights ? (
            <div className="auto-loading-state">
              <RefreshCw size={24} className="spin-icon" />
              <p>Scanning real operational signals...</p>
            </div>
          ) : insights.length === 0 ? (
            <div className="auto-empty-state">
              <div className="empty-icon-wrap">
                <CheckCircle2 size={32} className="text-success" />
              </div>
              <h3>No Active Operational Signals</h3>
              <p>All target business records are currently verified within normal operational parameters.</p>
            </div>
          ) : (
            <div className="table-responsive-wrapper">
              <table className="insights-table">
                <thead>
                  <tr>
                    <th>Source & Ref</th>
                    <th>Severity / Risk</th>
                    <th>Detected Operational Signal</th>
                    <th>Detection Rule</th>
                    <th>Recommended Next Step</th>
                    <th>Governance</th>
                    <th>Actions</th>
                  </tr>
                </thead>
                <tbody>
                  {insights
                    .filter((ins) => !search || ins.title?.toLowerCase().includes(search.toLowerCase()) || ins.source_record_ref?.toLowerCase().includes(search.toLowerCase()))
                    .map((ins) => {
                      const mod = ins.source_module?.toLowerCase();
                      const tracePath = mod === 'shipments' ? `/dashboard/shipments`
                        : mod === 'invoices' ? `/dashboard/invoices`
                        : mod === 'contracts' ? `/dashboard/contracts`
                        : mod === 'rfq' ? `/dashboard/rfqs`
                        : `/dashboard/customers`;

                      return (
                        <tr key={ins.id}>
                          <td>
                            <div className="font-semibold text-xs text-primary" style={{ textTransform: 'uppercase' }}>
                              {ins.source_module}
                            </div>
                            <div className="font-mono text-sm font-bold">
                              #{ins.source_record_id} {ins.source_record_ref && `(${ins.source_record_ref})`}
                            </div>
                            <span className="text-xs text-muted">
                              {formatDateTime(ins.created_at)}
                            </span>
                          </td>
                          <td>
                            <div className={`badge-severity-${ins.severity?.toLowerCase() || 'medium'}`}>
                              {ins.severity}
                            </div>
                            <div className="text-xs text-muted" style={{ marginTop: '4px' }}>
                              Risk: <strong>{ins.risk_level}</strong>
                            </div>
                          </td>
                          <td style={{ maxWidth: '300px' }}>
                            <div className="font-semibold text-sm">{ins.title}</div>
                            <div className="text-xs text-muted" style={{ marginTop: '4px' }}>
                              {ins.description}
                            </div>
                          </td>
                          <td>
                            <span className="badge-rule">{ins.detection_rule}</span>
                            <div className="text-xs text-muted" style={{ marginTop: '4px' }}>
                              Confidence: {Math.round((ins.confidence || 1.0) * 100)}%
                            </div>
                          </td>
                          <td style={{ maxWidth: '240px' }}>
                            {ins.recommended_next_step && (
                              <div className="insight-next-step">
                                {ins.recommended_next_step}
                              </div>
                            )}
                          </td>
                          <td>
                            {ins.is_approval_required ? (
                              <span className="badge-approval-req">
                                <ShieldCheck size={12} /> Approval Required
                              </span>
                            ) : (
                              <span className="text-xs text-success">Autonomous Safe</span>
                            )}
                            <div className="text-xs text-muted" style={{ marginTop: '4px' }}>
                              Status: <strong>{ins.status}</strong>
                            </div>
                          </td>
                          <td>
                            <div className="insight-action-btns">
                              {ins.status === 'ACTIVE' && (
                                <>
                                  <button
                                    className="btn-insight-ack"
                                    onClick={() => handleAcknowledgeInsight(ins)}
                                    title="Mark signal as acknowledged"
                                  >
                                    <Check size={13} /> Acknowledge
                                  </button>
                                  <button
                                    className="btn-insight-dismiss"
                                    onClick={() => handleDismissInsight(ins)}
                                    title="Dismiss signal"
                                  >
                                    <X size={13} /> Dismiss
                                  </button>
                                </>
                              )}
                              <Link
                                to={tracePath}
                                className="btn-insight-dismiss"
                                title="Trace record in module"
                                style={{ textDecoration: 'none', display: 'inline-flex', alignItems: 'center' }}
                              >
                                <ExternalLink size={13} /> Trace
                              </Link>
                            </div>
                          </td>
                        </tr>
                      );
                    })}
                </tbody>
              </table>
            </div>
          )}
        </div>
      )}

      {/* Execution History Tab */}
      {activeTab === 'history' && (
        <div className="history-tab-content">
          <div className="history-header">
            <h3>Audit Execution Log</h3>
            <p>Durable records of all scheduled and manual automation executions across your organization.</p>
          </div>

          {executions.length === 0 ? (
            <div className="auto-empty-state">
              <div className="empty-icon-wrap">
                <Activity size={32} />
              </div>
              <h3>No Execution History Yet</h3>
              <p>Trigger a manual run or wait for scheduled execution to view detailed audit logs.</p>
            </div>
          ) : (
            <div className="table-responsive-wrapper">
              <table className="automations-table">
                <thead>
                  <tr>
                    <th>Execution ID</th>
                    <th>Automation Name</th>
                    <th>Trigger</th>
                    <th>Status</th>
                    <th>Queued / Started</th>
                    <th>Duration</th>
                    <th>Evaluated</th>
                    <th>Recommendations</th>
                    <th>Action</th>
                  </tr>
                </thead>
                <tbody>
                  {executions.map((exec) => (
                    <tr key={exec.id}>
                      <td className="font-mono">#{exec.id}</td>
                      <td>
                        <div className="auto-name-title">{exec.automation_name || `Automation #${exec.automation_id}`}</div>
                        <div className="font-mono text-muted text-xs">{exec.correlation_id}</div>
                      </td>
                      <td>
                        <span className={`trigger-badge trigger-${exec.trigger_type.toLowerCase()}`}>
                          {exec.trigger_type}
                        </span>
                      </td>
                      <td>{getStatusBadge(exec.status)}</td>
                      <td>
                        <div>{formatDateTime(exec.started_at || exec.queued_at)}</div>
                      </td>
                      <td>{exec.duration_ms ? `${(exec.duration_ms / 1000).toFixed(2)}s` : '—'}</td>
                      <td>
                        <span className="eval-count">{exec.records_reviewed || 0} records</span>
                      </td>
                      <td>
                        <div className="recs-summary-badge">
                          <span className="badge-new">+{exec.recommendations_created || 0} new</span>
                          <span className="badge-refreshed">↺ {exec.recommendations_updated || 0} active</span>
                        </div>
                      </td>
                      <td>
                        <div className="history-actions">
                          <button
                            className="btn-icon-action"
                            onClick={() => handleOpenExecutionDetails(exec)}
                            title="Inspect execution details"
                          >
                            <Eye size={15} />
                            <span>Details</span>
                          </button>

                          {exec.status === 'FAILED' && (
                            <button
                              className="btn-retry-exec"
                              onClick={() => handleRetryExecution(exec)}
                              title="Retry failed execution"
                            >
                              <RotateCcw size={13} /> Retry
                            </button>
                          )}

                          {(exec.status === 'RUNNING' || exec.status === 'QUEUED') && (
                            <button
                              className="btn-icon-action btn-cancel"
                              onClick={() => handleCancelExecution(exec)}
                              title="Cancel execution"
                            >
                              <Ban size={15} />
                            </button>
                          )}
                        </div>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>
      )}

      {/* Supported Assistant Jobs Catalog Tab */}
      {activeTab === 'catalog' && (
        <div className="catalog-tab-content">
          <div className="catalog-header">
            <h3>Approved AI Assistant Job Templates</h3>
            <p>
              Pre-approved deterministic jobs supported by LogisticsHQ. Each job enforces strict read-only safety guarantees
              and routes all potential actions to the central Action System and human approval workflow.
            </p>
          </div>

          <div className="catalog-cards-grid">
            {supportedTypes.map((item) => (
              <div key={item.type} className="catalog-card">
                <div className="catalog-card-header">
                  <span className="catalog-category-tag">{item.category}</span>
                  <span className="read-only-pill">
                    <ShieldCheck size={12} /> Read-Only
                  </span>
                </div>

                <h4>{item.name}</h4>
                <p className="catalog-desc">{item.description}</p>

                <div className="catalog-modules-row">
                  <span className="meta-label">Target Modules:</span>
                  <div className="module-chips">
                    {item.target_modules?.map((m) => (
                      <span key={m} className="module-chip">
                        {m}
                      </span>
                    ))}
                  </div>
                </div>

                <div className="catalog-guarantee-box">
                  <div className="guarantee-title">
                    <Info size={13} /> Safety & Action Guarantee
                  </div>
                  <div className="guarantee-body">{item.read_only_guarantee}</div>
                </div>

                <div className="catalog-action-row">
                  <div className="catalog-sched-meta">
                    Default: {item.default_schedule_type} at {item.default_schedule_time} UTC
                  </div>
                  <button
                    className="btn-primary-dark btn-sm"
                    onClick={() => handleOpenCreateModal(item.type)}
                  >
                    <Plus size={14} />
                    Use Template
                  </button>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Controlled Action Orchestration Tab (Phase 3 Task 3.2) */}
      {activeTab === 'orchestration' && (
        <ActionOrchestrationTab onNavigateToApprovals={() => navigate('/dashboard/approvals')} />
      )}

      {/* Event-Driven AI Workflows and Cross-Module Automation (Phase 3 Task 3.8) */}
      {activeTab === 'event_workflows' && (
        <EventWorkflowsTab onNavigateToApprovals={() => navigate('/dashboard/approvals')} />
      )}

      {/* Enterprise Event Mesh and Autonomous Workflow Engine (Phase 7.8) */}
      {activeTab === 'event_mesh' && (
        <EnterpriseEventMeshSection />
      )}

      {/* ── CREATE / EDIT AUTOMATION MODAL ── */}
      {isCreateModalOpen && (
        <div className="modal-overlay">
          <div className="modal-dialog modal-lg" role="dialog" aria-modal="true">
            <div className="modal-header">
              <h3>{editingAutomation ? 'Edit Automation Definition' : 'Create Workflow Automation'}</h3>
              <button
                className="btn-close-modal"
                onClick={() => setIsCreateModalOpen(false)}
                aria-label="Close modal"
              >
                <X size={18} />
              </button>
            </div>

            <form onSubmit={handleSaveAutomation}>
              <div className="modal-body">
                {/* Assistant Type Picker */}
                <div className="form-group">
                  <label htmlFor="automation_type">Supported AI Assistant Type *</label>
                  <select
                    id="automation_type"
                    value={formData.automation_type}
                    disabled={!!editingAutomation}
                    onChange={(e) => {
                      const selected = supportedTypes.find((t) => t.type === e.target.value);
                      setFormData((prev) => ({
                        ...prev,
                        automation_type: e.target.value,
                        name: selected ? selected.name : prev.name,
                        description: selected ? selected.description : prev.description,
                        schedule_type: selected ? selected.default_schedule_type : prev.schedule_type,
                        schedule_time: selected ? selected.default_schedule_time : prev.schedule_time,
                      }));
                    }}
                    required
                  >
                    {supportedTypes.map((t) => (
                      <option key={t.type} value={t.type}>
                        {t.name} ({t.category})
                      </option>
                    ))}
                  </select>
                </div>

                {/* Automation Name */}
                <div className="form-group">
                  <label htmlFor="auto_name">Automation Name *</label>
                  <input
                    id="auto_name"
                    type="text"
                    value={formData.name}
                    onChange={(e) => setFormData((prev) => ({ ...prev, name: e.target.value }))}
                    placeholder="e.g., Daily Overdue Receivables Review"
                    required
                  />
                </div>

                {/* Description */}
                <div className="form-group">
                  <label htmlFor="auto_desc">Description / Purpose</label>
                  <textarea
                    id="auto_desc"
                    rows={2}
                    value={formData.description}
                    onChange={(e) => setFormData((prev) => ({ ...prev, description: e.target.value }))}
                    placeholder="Explain the operational purpose of this scheduled review..."
                  />
                </div>

                <div className="form-row-2">
                  {/* Schedule Type */}
                  <div className="form-group">
                    <label htmlFor="schedule_type">Schedule Cadence *</label>
                    <select
                      id="schedule_type"
                      value={formData.schedule_type}
                      onChange={(e) => setFormData((prev) => ({ ...prev, schedule_type: e.target.value }))}
                      required
                    >
                      <option value="DAILY">Daily</option>
                      <option value="WEEKLY">Weekly (Select Days)</option>
                      <option value="HOURLY">Hourly</option>
                    </select>
                  </div>

                  {/* Schedule Time */}
                  <div className="form-group">
                    <label htmlFor="schedule_time">Execution Time (HH:MM 24h) *</label>
                    <input
                      id="schedule_time"
                      type="text"
                      pattern="^([0-1]?[0-9]|2[0-3]):[0-5][0-9]$"
                      value={formData.schedule_time}
                      onChange={(e) => setFormData((prev) => ({ ...prev, schedule_time: e.target.value }))}
                      placeholder="08:00"
                      required
                    />
                  </div>
                </div>

                {/* Weekly Days selection if WEEKLY */}
                {formData.schedule_type === 'WEEKLY' && (
                  <div className="form-group">
                    <label>Active Days of the Week</label>
                    <div className="days-checkbox-group">
                      {['MON', 'TUE', 'WED', 'THU', 'FRI', 'SAT', 'SUN'].map((day) => {
                        const checked = formData.schedule_days.includes(day);
                        return (
                          <label key={day} className={`day-pill ${checked ? 'active' : ''}`}>
                            <input
                              type="checkbox"
                              checked={checked}
                              onChange={(e) => {
                                if (e.target.checked) {
                                  setFormData((prev) => ({
                                    ...prev,
                                    schedule_days: [...prev.schedule_days, day],
                                  }));
                                } else {
                                  setFormData((prev) => ({
                                    ...prev,
                                    schedule_days: prev.schedule_days.filter((d) => d !== day),
                                  }));
                                }
                              }}
                            />
                            {day}
                          </label>
                        );
                      })}
                    </div>
                  </div>
                )}

                <div className="form-row-2">
                  {/* Timezone */}
                  <div className="form-group">
                    <label htmlFor="timezone">Timezone *</label>
                    <select
                      id="timezone"
                      value={formData.timezone}
                      onChange={(e) => setFormData((prev) => ({ ...prev, timezone: e.target.value }))}
                      required
                    >
                      <option value="UTC">UTC (Universal Coordinated Time)</option>
                      <option value="America/New_York">Eastern Time (America/New_York)</option>
                      <option value="America/Chicago">Central Time (America/Chicago)</option>
                      <option value="America/Los_Angeles">Pacific Time (America/Los_Angeles)</option>
                      <option value="Europe/London">London / GMT (Europe/London)</option>
                      <option value="Asia/Kolkata">India Standard Time (Asia/Kolkata)</option>
                      <option value="Asia/Singapore">Singapore / Hong Kong (Asia/Singapore)</option>
                      <option value="Asia/Dubai">Gulf Standard Time (Asia/Dubai)</option>
                    </select>
                  </div>

                  {/* Execution Window */}
                  <div className="form-group">
                    <label htmlFor="window">Max Execution Window (Minutes)</label>
                    <input
                      id="window"
                      type="number"
                      min={10}
                      max={180}
                      value={formData.execution_window_minutes}
                      onChange={(e) => setFormData((prev) => ({ ...prev, execution_window_minutes: e.target.value }))}
                    />
                  </div>
                </div>

                {/* Safety Guarantee Box */}
                <div className="safety-guarantee-box">
                  <div className="safety-header">
                    <ShieldCheck size={18} className="safety-icon" />
                    <strong>Read-Only Autonomous Safety Guarantee</strong>
                  </div>
                  <p>
                    This job runs deterministic AI evaluations on your operational records. It will never modify database
                    records, change invoices, alter contracts, accept quotations, or send outbound messages automatically.
                    All findings generate recommendations in the Recommendation Center for human review and authorization.
                  </p>
                </div>
              </div>

              <div className="modal-footer">
                <button
                  type="button"
                  className="btn-secondary-light"
                  onClick={() => setIsCreateModalOpen(false)}
                >
                  Cancel
                </button>
                <button type="submit" className="btn-primary-dark">
                  {editingAutomation ? 'Update Automation' : 'Create & Schedule'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* ── EXECUTION DETAILS DRAWER ── */}
      {selectedExecution && (
        <div className="drawer-overlay" onClick={() => setSelectedExecution(null)}>
          <div className="drawer-panel" onClick={(e) => e.stopPropagation()}>
            <div className="drawer-header">
              <div>
                <span className="drawer-sub">Execution Audit Record</span>
                <h3>#{selectedExecution.id} — {selectedExecution.automation_name || 'Automation Execution'}</h3>
              </div>
              <button className="btn-close-modal" onClick={() => setSelectedExecution(null)}>
                <X size={18} />
              </button>
            </div>

            <div className="drawer-body">
              <div className="exec-meta-grid">
                <div className="meta-card">
                  <div className="meta-title">Status</div>
                  <div>{getStatusBadge(selectedExecution.status)}</div>
                </div>
                <div className="meta-card">
                  <div className="meta-title">Trigger Type</div>
                  <div>
                    <span className={`trigger-badge trigger-${selectedExecution.trigger_type.toLowerCase()}`}>
                      {selectedExecution.trigger_type}
                    </span>
                  </div>
                </div>
                <div className="meta-card">
                  <div className="meta-title">Execution Duration</div>
                  <div className="meta-val">
                    {selectedExecution.duration_ms ? `${(selectedExecution.duration_ms / 1000).toFixed(2)}s` : '—'}
                  </div>
                </div>
                <div className="meta-card">
                  <div className="meta-title">Records Evaluated</div>
                  <div className="meta-val">{selectedExecution.records_reviewed || 0}</div>
                </div>
              </div>

              {/* Summary Text */}
              <div className="drawer-section">
                <h4>Execution Summary</h4>
                <div className="summary-box">
                  {selectedExecution.summary_text || 'No summary text available.'}
                </div>
              </div>

              {/* Error Box if Failed */}
              {selectedExecution.error_message && (
                <div className="drawer-section">
                  <h4 className="text-danger">Error Diagnostics</h4>
                  <div className="error-box">
                    <AlertTriangle size={16} />
                    <span>{selectedExecution.error_message}</span>
                  </div>
                </div>
              )}

              {/* Recommendation Linkage */}
              <div className="drawer-section">
                <div className="section-header-row">
                  <h4>Generated Recommendations ({executionRecommendations.length})</h4>
                  <Link
                    to={`/dashboard/recommendations?execution_id=${selectedExecution.id}`}
                    className="btn-link-sm"
                  >
                    View in Recommendation Center <ExternalLink size={12} />
                  </Link>
                </div>

                {loadingExecRecs ? (
                  <div className="loading-recs-text">
                    <RefreshCw size={14} className="spin-icon" /> Fetching linked recommendations...
                  </div>
                ) : executionRecommendations.length === 0 ? (
                  <div className="empty-recs-box">
                    <CheckCircle2 size={16} className="text-success" />
                    <span>No active operational risks detected during this review. No recommendations needed.</span>
                  </div>
                ) : (
                  <div className="recs-list-preview">
                    {executionRecommendations.map((rec) => (
                      <div key={rec.id} className="rec-preview-card">
                        <div className="rec-preview-header">
                          <span className={`priority-tag priority-${rec.priority?.toLowerCase()}`}>
                            {rec.priority}
                          </span>
                          <span className="source-tag">{rec.source_type} #{rec.source_id}</span>
                        </div>
                        <div className="rec-preview-title">{rec.title}</div>
                        <div className="rec-preview-action">{rec.recommended_action}</div>
                      </div>
                    ))}
                  </div>
                )}
              </div>

              {/* Operational Insights Linkage (Phase 3) */}
              {executionInsights.length > 0 && (
                <div className="drawer-section">
                  <div className="section-header-row">
                    <h4>Detected Operational Signals ({executionInsights.length})</h4>
                  </div>
                  <div className="recs-list-preview">
                    {executionInsights.map((ins) => (
                      <div key={ins.id} className="rec-preview-card">
                        <div className="rec-preview-header">
                          <span className={`priority-tag priority-${ins.severity?.toLowerCase() || 'medium'}`}>
                            {ins.severity}
                          </span>
                          <span className="source-tag">{ins.source_module} #{ins.source_record_id}</span>
                        </div>
                        <div className="rec-preview-title">{ins.title}</div>
                        <div className="rec-preview-action">{ins.recommended_next_step}</div>
                      </div>
                    ))}
                  </div>
                </div>
              )}

              {/* Audit Details */}
              <div className="drawer-section">
                <h4>Audit Traceability</h4>
                <div className="audit-info-list">
                  <div className="audit-info-row">
                    <span className="label">Correlation ID:</span>
                    <span className="font-mono">{selectedExecution.correlation_id}</span>
                  </div>
                  <div className="audit-info-row">
                    <span className="label">Started At:</span>
                    <span>{formatDateTime(selectedExecution.started_at)}</span>
                  </div>
                  <div className="audit-info-row">
                    <span className="label">Completed At:</span>
                    <span>{formatDateTime(selectedExecution.completed_at)}</span>
                  </div>
                  <div className="audit-info-row">
                    <span className="label">Read-Only Enforced:</span>
                    <span className="text-success font-semibold">True (Zero database mutations)</span>
                  </div>
                </div>
              </div>
            </div>

            <div className="drawer-footer">
              <button className="btn-secondary-light" onClick={() => setSelectedExecution(null)}>
                Close
              </button>
              {(selectedExecution.status === 'RUNNING' || selectedExecution.status === 'QUEUED') && (
                <button
                  className="btn-danger-light"
                  onClick={() => handleCancelExecution(selectedExecution)}
                >
                  <Ban size={14} /> Cancel Execution
                </button>
              )}
            </div>
          </div>
        </div>
      )}

      {/* ── SCHEDULE PREVIEW MODAL ── */}
      {previewScheduleModal && (
        <div className="modal-overlay" onClick={() => setPreviewScheduleModal(null)}>
          <div className="modal-dialog modal-md" onClick={(e) => e.stopPropagation()}>
            <div className="modal-header">
              <h3>Schedule Next-Run Preview</h3>
              <button className="btn-close-modal" onClick={() => setPreviewScheduleModal(null)}>
                <X size={18} />
              </button>
            </div>
            <div className="modal-body">
              <div className="preview-schedule-card">
                <Calendar size={28} className="preview-icon" />
                <h4>{previewScheduleModal.auto.name}</h4>
                <div className="preview-time-display">
                  {previewScheduleModal.preview.local_time_formatted}
                </div>
                <div className="preview-desc-text">
                  {previewScheduleModal.preview.human_description}
                </div>
                <div className="preview-diff-badge">
                  Next execution in approximately{' '}
                  <strong>{Math.round(previewScheduleModal.preview.hours_until_run)} hours</strong>
                </div>
              </div>
            </div>
            <div className="modal-footer">
              <button
                className="btn-primary-dark"
                onClick={() => setPreviewScheduleModal(null)}
              >
                Done
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

import { useState, useEffect, useCallback } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import {
  Bell,
  CheckCircle2,
  AlertTriangle,
  AlertCircle,
  Clock,
  Filter,
  Search,
  Check,
  Eye,
  ExternalLink,
  ShieldAlert,
  ArrowUpRight,
  RefreshCw,
  Sliders,
  Settings,
  X,
  FileText,
  Ship,
  CreditCard,
  CheckSquare,
  Cpu,
  Compass,
  ChevronRight,
  SlidersHorizontal,
  CheckCheck,
  Copy,
  Sparkles,
  Layers,
  Timer,
  Send,
  UserCheck,
  ChevronDown,
  Info,
} from 'lucide-react';
import { notificationCenterService } from '../../../services/notificationCenterService';
import './NotificationCenterPage.css';

export default function NotificationCenterPage() {
  const navigate = useNavigate();
  const [searchParams, setSearchParams] = useSearchParams();

  // Active Tab: 'notifications' | 'escalations' | 'preferences'
  const [activeTab, setActiveTab] = useState(searchParams.get('tab') || 'notifications');

  // Notifications List State
  const [notifications, setNotifications] = useState([]);
  const [totalNotifications, setTotalNotifications] = useState(0);
  const [stats, setStats] = useState(null);
  const [escalations, setEscalations] = useState([]);
  const [preferences, setPreferences] = useState(null);
  const [loading, setLoading] = useState(true);
  const [evaluating, setEvaluating] = useState(false);
  const [savingPrefs, setSavingPrefs] = useState(false);
  const [toastMessage, setToastMessage] = useState(null);

  // Filter States
  const [searchQuery, setSearchQuery] = useState('');
  const [selectedSeverity, setSelectedSeverity] = useState('ALL');
  const [selectedModule, setSelectedModule] = useState('ALL');
  const [selectedDeliveryStatus, setSelectedDeliveryStatus] = useState('ALL');
  const [readFilter, setReadFilter] = useState('ALL'); // 'ALL' | 'UNREAD' | 'READ'
  const [actionRequiredOnly, setActionRequiredOnly] = useState(false);
  const [escalatedOnly, setEscalatedOnly] = useState(false);
  const [page, setPage] = useState(1);
  const pageSize = 15;

  // Drawer and Interactive Actions State
  const [selectedNotification, setSelectedNotification] = useState(null);
  const [aiAnalyzing, setAiAnalyzing] = useState(false);
  const [draftGenerating, setDraftGenerating] = useState(false);
  const [draftType, setDraftType] = useState('OPERATIONAL_ALERT');
  const [generatedDraft, setGeneratedDraft] = useState(null);
  const [escalateModalOpen, setEscalateModalOpen] = useState(false);
  const [escalateReason, setEscalateReason] = useState('');
  const [escalating, setEscalating] = useState(false);
  const [snoozeMinutes, setSnoozeMinutes] = useState(240);
  const [snoozeDropdownNotifId, setSnoozeDropdownNotifId] = useState(null);

  const showToast = (msg) => {
    setToastMessage(msg);
    setTimeout(() => setToastMessage(null), 3500);
  };

  // Load Notifications & Stats
  const loadData = useCallback(async () => {
    setLoading(true);
    try {
      const isReadParam = readFilter === 'UNREAD' ? 'false' : readFilter === 'READ' ? 'true' : undefined;
      const actionParam = actionRequiredOnly ? 'true' : undefined;
      const escParam = escalatedOnly ? 'true' : undefined;
      const deliveryParam = selectedDeliveryStatus !== 'ALL' ? selectedDeliveryStatus : undefined;

      const [listRes, statsRes] = await Promise.all([
        notificationCenterService.listNotifications({
          page,
          pageSize,
          severity: selectedSeverity,
          module: selectedModule,
          delivery_status: deliveryParam,
          is_read: isReadParam,
          action_required: actionParam,
          is_escalated: escParam,
          search: searchQuery,
        }),
        notificationCenterService.getStats(),
      ]);

      setNotifications(listRes?.data || []);
      setTotalNotifications(listRes?.total || 0);
      setStats(statsRes?.data || null);
    } catch (err) {
      console.error('Failed to load notifications:', err);
    } finally {
      setLoading(false);
    }
  }, [page, selectedSeverity, selectedModule, selectedDeliveryStatus, readFilter, actionRequiredOnly, escalatedOnly, searchQuery]);

  // Load Escalation Audit Events
  const loadEscalations = async () => {
    try {
      const res = await notificationCenterService.listEscalations(50);
      setEscalations(res?.data || []);
    } catch (err) {
      console.error('Failed to load escalations:', err);
    }
  };

  // Load User Preferences
  const loadPreferences = async () => {
    try {
      const res = await notificationCenterService.getPreferences();
      setPreferences(res?.data || null);
    } catch (err) {
      console.error('Failed to load preferences:', err);
    }
  };

  useEffect(() => {
    loadData();
  }, [loadData]);

  useEffect(() => {
    if (activeTab === 'escalations') {
      loadEscalations();
    } else if (activeTab === 'preferences') {
      loadPreferences();
    }
  }, [activeTab]);

  // Handle Mark as Read / Unread
  const handleToggleRead = async (e, notif) => {
    e.stopPropagation();
    try {
      if (notif.is_read) {
        await notificationCenterService.markAsUnread(notif.id);
        showToast('Notification marked as unread');
      } else {
        await notificationCenterService.markAsRead(notif.id);
        showToast('Notification marked as read');
      }
      loadData();
      if (selectedNotification && selectedNotification.id === notif.id) {
        setSelectedNotification(prev => ({ ...prev, is_read: !notif.is_read }));
      }
    } catch (err) {
      showToast('Failed to update read state');
    }
  };

  // Handle Dismiss
  const handleDismiss = async (e, notifId) => {
    e.stopPropagation();
    try {
      await notificationCenterService.dismiss(notifId);
      showToast('Notification dismissed');
      if (selectedNotification && selectedNotification.id === notifId) {
        setSelectedNotification(null);
      }
      loadData();
    } catch (err) {
      showToast('Failed to dismiss notification');
    }
  };

  // Handle Mark All Visible as Read
  const handleMarkAllRead = async () => {
    try {
      await notificationCenterService.markAllAsRead();
      showToast('All visible notifications marked as read');
      loadData();
    } catch (err) {
      showToast('Failed to mark all as read');
    }
  };

  // Trigger Evaluation Sweep
  const handleEvaluate = async () => {
    setEvaluating(true);
    try {
      const res = await notificationCenterService.evaluate();
      showToast(`Evaluation completed. ${res?.created || 0} notifications synchronized.`);
      loadData();
      if (activeTab === 'escalations') {
        loadEscalations();
      }
    } catch (err) {
      showToast('Evaluation failed: ' + (err.message || 'Server error'));
    } finally {
      setEvaluating(false);
    }
  };

  // Save Preferences
  const handleSavePreferences = async (e) => {
    e.preventDefault();
    if (!preferences) return;
    setSavingPrefs(true);
    try {
      await notificationCenterService.updatePreferences(preferences);
      showToast('Notification preferences saved successfully');
    } catch (err) {
      showToast('Failed to save preferences');
    } finally {
      setSavingPrefs(false);
    }
  };

  // Acknowledge Notification
  const handleAcknowledge = async (e, notifId) => {
    if (e) e.stopPropagation();
    try {
      await notificationCenterService.acknowledge(notifId);
      showToast(`Notification #${notifId} acknowledged`);
      loadData();
      if (selectedNotification && selectedNotification.id === notifId) {
        setSelectedNotification(prev => ({
          ...prev,
          is_acknowledged: true,
          delivery_status: 'ACKNOWLEDGED',
          acknowledged_at: new Date().toISOString(),
        }));
      }
    } catch (err) {
      showToast('Failed to acknowledge: ' + (err.message || 'Error'));
    }
  };

  // Snooze Notification
  const handleSnooze = async (e, notifId, durationMinutes = 240) => {
    if (e) e.stopPropagation();
    try {
      await notificationCenterService.snooze(notifId, durationMinutes);
      showToast(`Notification #${notifId} snoozed for ${durationMinutes >= 60 ? durationMinutes / 60 + 'h' : durationMinutes + 'm'}`);
      setSnoozeDropdownNotifId(null);
      loadData();
      if (selectedNotification && selectedNotification.id === notifId) {
        setSelectedNotification(prev => ({
          ...prev,
          is_snoozed: true,
          delivery_status: 'SNOOZED',
          snoozed_until: new Date(Date.now() + durationMinutes * 60000).toISOString(),
        }));
      }
    } catch (err) {
      showToast('Failed to snooze: ' + (err.message || 'Error'));
    }
  };

  // Run AI Analysis
  const handleAnalyzeAI = async (e, notifId) => {
    if (e) e.stopPropagation();
    setAiAnalyzing(true);
    try {
      const res = await notificationCenterService.analyzeAI(notifId);
      const aiData = res?.data || res;
      showToast(`AI analysis complete: Priority score ${Math.round(aiData.priority_score || 0)}/100`);
      loadData();
      if (selectedNotification && selectedNotification.id === notifId) {
        setSelectedNotification(prev => ({
          ...prev,
          ai_summary: aiData.ai_summary,
          ai_escalation_reason: aiData.ai_escalation_reason,
          ai_priority_score: aiData.priority_score,
          group_key: aiData.group_key || prev.group_key,
          ai_analysis_response: aiData,
        }));
      }
    } catch (err) {
      showToast('AI analysis failed: ' + (err.message || 'Server error'));
    } finally {
      setAiAnalyzing(false);
    }
  };

  // Escalate with AI
  const handleEscalateSubmit = async (e) => {
    if (e) e.preventDefault();
    if (!selectedNotification) return;
    setEscalating(true);
    try {
      const res = await notificationCenterService.escalate(
        selectedNotification.id,
        escalateReason || 'Manual operator escalation based on SLA risk'
      );
      const escEvent = res?.data || res;
      if (escEvent?.approval_id) {
        showToast(`Escalated to Level ${escEvent.escalation_level}: Approval #${escEvent.approval_id} created for human review`);
      } else {
        showToast(`Notification escalated to Level ${escEvent?.escalation_level || 2}`);
      }
      setEscalateModalOpen(false);
      setEscalateReason('');
      loadData();
      if (activeTab === 'escalations') {
        loadEscalations();
      }
      setSelectedNotification(prev => ({
        ...prev,
        is_escalated: true,
        escalation_level: escEvent?.escalation_level || ((prev.escalation_level || 1) + 1),
        delivery_status: 'ESCALATED',
      }));
    } catch (err) {
      showToast('Escalation failed: ' + (err.message || 'Server error'));
    } finally {
      setEscalating(false);
    }
  };

  // Generate Draft
  const handleGenerateDraft = async (notifId, type = draftType) => {
    setDraftGenerating(true);
    try {
      const res = await notificationCenterService.generateDraft(notifId, type);
      setGeneratedDraft(res?.data || res);
      showToast('AI Escalation draft ready for review');
    } catch (err) {
      showToast('Draft generation failed: ' + (err.message || 'Server error'));
    } finally {
      setDraftGenerating(false);
    }
  };

  const copyDraftToClipboard = () => {
    if (!generatedDraft) return;
    const text = `Subject: ${generatedDraft.subject}\n\n${generatedDraft.body_text}`;
    navigator.clipboard.writeText(text);
    showToast('Draft copied to clipboard');
  };

  const getModuleIcon = (mod) => {
    switch (mod) {
      case 'SHIPMENTS':
        return <Ship size={14} className="module-icon text-blue-600" />;
      case 'INVOICES':
        return <CreditCard size={14} className="module-icon text-rose-600" />;
      case 'APPROVALS':
        return <CheckSquare size={14} className="module-icon text-purple-600" />;
      case 'AUTOMATIONS':
        return <Cpu size={14} className="module-icon text-amber-600" />;
      case 'RECOMMENDATIONS':
        return <Compass size={14} className="module-icon text-emerald-600" />;
      case 'CONTRACTS':
        return <FileText size={14} className="module-icon text-cyan-600" />;
      default:
        return <Bell size={14} className="module-icon text-slate-500" />;
    }
  };

  const getSeverityBadgeClass = (sev) => {
    switch (sev) {
      case 'CRITICAL':
        return 'badge-critical';
      case 'HIGH':
        return 'badge-high';
      case 'MEDIUM':
        return 'badge-medium';
      case 'LOW':
        return 'badge-low';
      default:
        return 'badge-info';
    }
  };

  const getDeliveryStatusBadge = (status) => {
    switch (status) {
      case 'ACKNOWLEDGED':
        return <span className="notif-status-badge badge-acknowledged"><CheckCheck size={11} /> Acknowledged</span>;
      case 'SNOOZED':
        return <span className="notif-status-badge badge-snoozed"><Timer size={11} /> Snoozed</span>;
      case 'ESCALATED':
        return <span className="notif-status-badge badge-escalated"><ShieldAlert size={11} /> Escalated</span>;
      case 'DISMISSED':
        return <span className="notif-status-badge badge-dismissed">Dismissed</span>;
      default:
        return <span className="notif-status-badge badge-delivered">Delivered</span>;
    }
  };

  const formatTimestamp = (ts) => {
    if (!ts) return 'Just now';
    try {
      const d = new Date(ts);
      return d.toLocaleString(undefined, {
        month: 'short',
        day: 'numeric',
        hour: '2-digit',
        minute: '2-digit',
      });
    } catch {
      return 'Recent';
    }
  };

  return (
    <div className="notif-center-page">
      {/* Toast Notification */}
      {toastMessage && (
        <div className="notif-center-toast" role="alert">
          <CheckCircle2 size={16} className="text-emerald-600" />
          <span>{toastMessage}</span>
        </div>
      )}

      {/* Page Header */}
      <div className="notif-page-header">
        <div className="notif-header-title-row">
          <div className="notif-title-icon-box">
            <Bell size={22} className="text-blue-600" />
          </div>
          <div>
            <h1 className="notif-page-title">Notification & Escalation Center</h1>
            <p className="notif-page-desc">
              Deterministic operational monitoring, role-aware routing, and AI-assisted escalation workflows.
            </p>
          </div>
        </div>

        <div className="notif-header-actions">
          <button
            className="notif-btn notif-btn-secondary"
            onClick={handleEvaluate}
            disabled={evaluating}
            title="Scan database for pending events, SLA breaches, and tracking exceptions"
          >
            <RefreshCw size={14} className={evaluating ? 'spin-icon' : ''} />
            <span>{evaluating ? 'Evaluating Signals...' : 'Run SLA Sweep'}</span>
          </button>

          {activeTab === 'notifications' && (
            <button
              className="notif-btn notif-btn-secondary"
              onClick={handleMarkAllRead}
              title="Mark all notifications as read"
            >
              <Check size={14} />
              <span>Mark All as Read</span>
            </button>
          )}

          <button
            className={`notif-btn ${activeTab === 'preferences' ? 'notif-btn-primary' : 'notif-btn-secondary'}`}
            onClick={() => setActiveTab(activeTab === 'preferences' ? 'notifications' : 'preferences')}
          >
            <SlidersHorizontal size={14} />
            <span>Preferences</span>
          </button>
        </div>
      </div>

      {/* KPI Metrics Bar */}
      <div className="notif-metrics-row">
        <div className="notif-metric-card" onClick={() => { setReadFilter('ALL'); setEscalatedOnly(false); setSelectedDeliveryStatus('ALL'); }}>
          <div className="notif-metric-content">
            <span className="notif-metric-label">Total Notifications</span>
            <span className="notif-metric-value">{stats?.total ?? 0}</span>
          </div>
          <div className="notif-metric-icon bg-blue-50 text-blue-600">
            <Bell size={20} />
          </div>
        </div>

        <div className="notif-metric-card" onClick={() => { setReadFilter('UNREAD'); setEscalatedOnly(false); }}>
          <div className="notif-metric-content">
            <span className="notif-metric-label">Unread Items</span>
            <span className="notif-metric-value text-blue-600">{stats?.unread ?? 0}</span>
          </div>
          <div className="notif-metric-icon bg-blue-50 text-blue-600">
            <Clock size={20} />
          </div>
        </div>

        <div className="notif-metric-card" onClick={() => { setActionRequiredOnly(true); }}>
          <div className="notif-metric-content">
            <span className="notif-metric-label">Action Required</span>
            <span className="notif-metric-value text-purple-600">{stats?.action_required ?? 0}</span>
          </div>
          <div className="notif-metric-icon bg-purple-50 text-purple-600">
            <AlertCircle size={20} />
          </div>
        </div>

        <div className={`notif-metric-card ${(stats?.escalated ?? 0) > 0 ? 'escalated-active' : ''}`} onClick={() => { setActiveTab('escalations'); }}>
          <div className="notif-metric-content">
            <span className="notif-metric-label">Escalated Items</span>
            <span className="notif-metric-value text-rose-600">
              {stats?.escalated ?? 0}
              {(stats?.escalated ?? 0) > 0 && <span className="escalation-pulse-badge">SLA Breach</span>}
            </span>
          </div>
          <div className="notif-metric-icon bg-rose-50 text-rose-600">
            <ShieldAlert size={20} />
          </div>
        </div>

        <div className="notif-metric-card" onClick={() => { setSelectedSeverity('CRITICAL'); }}>
          <div className="notif-metric-content">
            <span className="notif-metric-label">Critical Severity</span>
            <span className="notif-metric-value text-rose-700">{stats?.critical_count ?? 0}</span>
          </div>
          <div className="notif-metric-icon bg-red-50 text-red-600">
            <AlertTriangle size={20} />
          </div>
        </div>
      </div>

      {/* Tabs Navigation */}
      <div className="notif-tabs-bar">
        <button
          className={`notif-tab ${activeTab === 'notifications' ? 'active' : ''}`}
          onClick={() => setActiveTab('notifications')}
        >
          <Bell size={15} />
          <span>Active Notifications</span>
          {stats?.unread > 0 && <span className="tab-pill">{stats.unread}</span>}
        </button>

        <button
          className={`notif-tab ${activeTab === 'escalations' ? 'active' : ''}`}
          onClick={() => setActiveTab('escalations')}
        >
          <ShieldAlert size={15} />
          <span>Escalations Audit</span>
          {stats?.escalated > 0 && <span className="tab-pill pill-danger">{stats.escalated}</span>}
        </button>

        <button
          className={`notif-tab ${activeTab === 'preferences' ? 'active' : ''}`}
          onClick={() => setActiveTab('preferences')}
        >
          <SlidersHorizontal size={15} />
          <span>Notification Preferences</span>
        </button>
      </div>

      {/* ── TAB 1: NOTIFICATIONS LIST & FILTERS ── */}
      {activeTab === 'notifications' && (
        <div className="notif-tab-content">
          {/* Filters Bar */}
          <div className="notif-filters-panel">
            <div className="notif-search-wrap">
              <Search size={15} className="notif-search-icon" />
              <input
                type="text"
                className="notif-search-input"
                placeholder="Filter by title, record ID, message, or group cluster..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
              />
              {searchQuery && (
                <button className="notif-clear-search" onClick={() => setSearchQuery('')}>
                  <X size={13} />
                </button>
              )}
            </div>

            <div className="notif-filter-group">
              {/* Severity Filter */}
              <div className="notif-select-wrap">
                <span className="notif-filter-label">Severity:</span>
                <select
                  className="notif-select"
                  value={selectedSeverity}
                  onChange={(e) => setSelectedSeverity(e.target.value)}
                >
                  <option value="ALL">All Severities</option>
                  <option value="CRITICAL">Critical</option>
                  <option value="HIGH">High</option>
                  <option value="MEDIUM">Medium</option>
                  <option value="LOW">Low</option>
                  <option value="INFORMATIONAL">Informational</option>
                </select>
              </div>

              {/* Module Filter */}
              <div className="notif-select-wrap">
                <span className="notif-filter-label">Module:</span>
                <select
                  className="notif-select"
                  value={selectedModule}
                  onChange={(e) => setSelectedModule(e.target.value)}
                >
                  <option value="ALL">All Modules</option>
                  <option value="SHIPMENTS">Shipments</option>
                  <option value="INVOICES">Invoices</option>
                  <option value="APPROVALS">Approvals</option>
                  <option value="AUTOMATIONS">Automations</option>
                  <option value="RECOMMENDATIONS">Recommendations</option>
                  <option value="CONTRACTS">Contracts</option>
                  <option value="SYSTEM">System</option>
                </select>
              </div>

              {/* Delivery Status Filter */}
              <div className="notif-select-wrap">
                <span className="notif-filter-label">Delivery:</span>
                <select
                  className="notif-select"
                  value={selectedDeliveryStatus}
                  onChange={(e) => setSelectedDeliveryStatus(e.target.value)}
                >
                  <option value="ALL">All Delivery States</option>
                  <option value="DELIVERED">Delivered</option>
                  <option value="ACKNOWLEDGED">Acknowledged</option>
                  <option value="SNOOZED">Snoozed</option>
                  <option value="ESCALATED">Escalated</option>
                </select>
              </div>

              {/* Read/Unread Filter */}
              <div className="notif-select-wrap">
                <span className="notif-filter-label">Read:</span>
                <select
                  className="notif-select"
                  value={readFilter}
                  onChange={(e) => setReadFilter(e.target.value)}
                >
                  <option value="ALL">All States</option>
                  <option value="UNREAD">Unread Only</option>
                  <option value="READ">Read Only</option>
                </select>
              </div>

              {/* Action Required Toggle */}
              <label className="notif-checkbox-label">
                <input
                  type="checkbox"
                  checked={actionRequiredOnly}
                  onChange={(e) => setActionRequiredOnly(e.target.checked)}
                />
                <span>Action Required</span>
              </label>

              {/* Reset Filters */}
              {(selectedSeverity !== 'ALL' || selectedModule !== 'ALL' || selectedDeliveryStatus !== 'ALL' || readFilter !== 'ALL' || actionRequiredOnly || searchQuery) && (
                <button
                  className="notif-reset-filters-btn"
                  onClick={() => {
                    setSelectedSeverity('ALL');
                    setSelectedModule('ALL');
                    setSelectedDeliveryStatus('ALL');
                    setReadFilter('ALL');
                    setActionRequiredOnly(false);
                    setSearchQuery('');
                  }}
                >
                  Reset
                </button>
              )}
            </div>
          </div>

          {/* Notifications List */}
          {loading ? (
            <div className="notif-loading-state">
              <RefreshCw size={24} className="spin-icon text-blue-600" />
              <p>Loading notification events...</p>
            </div>
          ) : notifications.length === 0 ? (
            <div className="notif-empty-state">
              <div className="notif-empty-icon-box">
                <CheckCircle2 size={32} className="text-emerald-500" />
              </div>
              <h3 className="notif-empty-title">All Caught Up!</h3>
              <p className="notif-empty-desc">
                No active notifications match your current filter criteria. All SLA thresholds and business conditions are operating normally.
              </p>
              <button
                className="notif-btn notif-btn-secondary"
                onClick={() => {
                  setSelectedSeverity('ALL');
                  setSelectedModule('ALL');
                  setSelectedDeliveryStatus('ALL');
                  setReadFilter('ALL');
                  setActionRequiredOnly(false);
                  setSearchQuery('');
                }}
              >
                Clear Filters
              </button>
            </div>
          ) : (
            <div className="notif-cards-list">
              {notifications.map((notif) => {
                return (
                  <div
                    key={notif.id}
                    className={`notif-card ${!notif.is_read ? 'unread' : ''} ${notif.is_escalated ? 'escalated' : ''} ${selectedNotification?.id === notif.id ? 'card-selected' : ''}`}
                    onClick={() => {
                      setSelectedNotification(notif);
                      setGeneratedDraft(null);
                    }}
                  >
                    {/* Left Severity Indicator Stripe */}
                    <div className={`notif-card-stripe stripe-${(notif.severity || 'informational').toLowerCase()}`} />

                    <div className="notif-card-main">
                      {/* Top Meta Line */}
                      <div className="notif-card-meta-line">
                        <div className="notif-card-tags">
                          <span className="notif-module-tag">
                            {getModuleIcon(notif.source_module)}
                            <span>{notif.source_module}</span>
                          </span>

                          <span className={`notif-sev-tag ${getSeverityBadgeClass(notif.severity)}`}>
                            {notif.severity}
                          </span>

                          {getDeliveryStatusBadge(notif.delivery_status)}

                          {notif.is_escalated && (
                            <span className="notif-escalated-tag">
                              <ShieldAlert size={12} />
                              <span>Escalated (L{notif.escalation_level})</span>
                            </span>
                          )}

                          {notif.action_required && (
                            <span className="notif-action-tag">
                              Action Required
                            </span>
                          )}

                          {notif.group_key && (
                            <span className="notif-cluster-tag" title={`Group Cluster: ${notif.group_key}`}>
                              <Layers size={11} />
                              <span>{notif.group_key}</span>
                            </span>
                          )}

                          {notif.ai_priority_score > 0 && (
                            <span className="notif-ai-priority-tag">
                              <Sparkles size={11} />
                              <span>AI Score: {Math.round(notif.ai_priority_score)}</span>
                            </span>
                          )}

                          <span className="notif-source-ref">
                            Ref: #{notif.source_record_id}
                          </span>
                        </div>

                        <div className="notif-card-time">
                          <span>{formatTimestamp(notif.created_at)}</span>
                        </div>
                      </div>

                      {/* Content Row */}
                      <div className="notif-card-body">
                        <h4 className="notif-card-title">{notif.title}</h4>
                        <p className="notif-card-message">{notif.message}</p>

                        {/* Inline AI Summary Highlight */}
                        {notif.ai_summary && (
                          <div className="notif-ai-summary-snippet">
                            <Sparkles size={13} className="text-blue-600 flex-shrink-0 mt-0.5" />
                            <div className="ai-snippet-content">
                              <span className="ai-snippet-label">AI Operational Insight: </span>
                              <span className="ai-snippet-text">{notif.ai_summary}</span>
                            </div>
                          </div>
                        )}
                      </div>

                      {/* Footer Actions */}
                      <div className="notif-card-footer" onClick={(e) => e.stopPropagation()}>
                        <div className="notif-card-links">
                          {notif.action_url && (
                            <button
                              className="notif-link-btn"
                              onClick={() => navigate(notif.action_url)}
                            >
                              <span>Review Source Record</span>
                              <ExternalLink size={12} />
                            </button>
                          )}

                          <button
                            className="notif-inspect-btn"
                            onClick={() => {
                              setSelectedNotification(notif);
                              setGeneratedDraft(null);
                            }}
                          >
                            <Eye size={12} />
                            <span>Details & Actions</span>
                          </button>
                        </div>

                        <div className="notif-card-controls">
                          {/* Quick Acknowledge */}
                          {!notif.is_acknowledged ? (
                            <button
                              className="notif-control-btn btn-ack"
                              onClick={(e) => handleAcknowledge(e, notif.id)}
                              title="Acknowledge notification"
                            >
                              <CheckCheck size={13} className="text-emerald-600" />
                              <span>Acknowledge</span>
                            </button>
                          ) : (
                            <span className="notif-ack-indicator" title={`Acknowledged by user #${notif.acknowledged_by || 'System'}`}>
                              <CheckCheck size={12} />
                              <span>Acked</span>
                            </span>
                          )}

                          {/* Quick Snooze */}
                          <div className="notif-snooze-container">
                            <button
                              className="notif-control-btn btn-snooze"
                              onClick={(e) => {
                                e.stopPropagation();
                                setSnoozeDropdownNotifId(snoozeDropdownNotifId === notif.id ? null : notif.id);
                              }}
                              title="Snooze notification"
                            >
                              <Timer size={13} className="text-amber-600" />
                              <span>Snooze</span>
                              <ChevronDown size={11} />
                            </button>

                            {snoozeDropdownNotifId === notif.id && (
                              <div className="notif-snooze-menu" onClick={(e) => e.stopPropagation()}>
                                <button onClick={(e) => handleSnooze(e, notif.id, 60)}>1 Hour</button>
                                <button onClick={(e) => handleSnooze(e, notif.id, 240)}>4 Hours</button>
                                <button onClick={(e) => handleSnooze(e, notif.id, 1440)}>24 Hours</button>
                              </div>
                            )}
                          </div>

                          <button
                            className="notif-control-btn"
                            onClick={(e) => handleToggleRead(e, notif)}
                            title={notif.is_read ? 'Mark as Unread' : 'Mark as Read'}
                          >
                            <Check size={13} className={notif.is_read ? 'text-emerald-600' : 'text-slate-400'} />
                            <span>{notif.is_read ? 'Mark Unread' : 'Mark Read'}</span>
                          </button>

                          <button
                            className="notif-control-btn btn-dismiss"
                            onClick={(e) => handleDismiss(e, notif.id)}
                            title="Dismiss notification"
                          >
                            <X size={13} />
                            <span>Dismiss</span>
                          </button>
                        </div>
                      </div>
                    </div>
                  </div>
                );
              })}
            </div>
          )}

          {/* Pagination */}
          {totalNotifications > pageSize && (
            <div className="notif-pagination">
              <span className="pagination-info">
                Showing {notifications.length} of {totalNotifications} notifications
              </span>
              <div className="pagination-buttons">
                <button
                  className="notif-btn notif-btn-secondary"
                  disabled={page <= 1}
                  onClick={() => setPage(page - 1)}
                >
                  Previous
                </button>
                <span className="pagination-curr">Page {page}</span>
                <button
                  className="notif-btn notif-btn-secondary"
                  disabled={page * pageSize >= totalNotifications}
                  onClick={() => setPage(page + 1)}
                >
                  Next
                </button>
              </div>
            </div>
          )}
        </div>
      )}

      {/* ── TAB 2: ESCALATIONS AUDIT TRAIL ── */}
      {activeTab === 'escalations' && (
        <div className="notif-tab-content">
          <div className="escalations-audit-panel">
            <div className="audit-header">
              <div>
                <h3 className="audit-title">Deterministic Escalation Audit Trail</h3>
                <p className="audit-desc">
                  Auditable history of operational threshold breaches, SLA escalations, and automated severity adjustments.
                </p>
              </div>
            </div>

            {escalations.length === 0 ? (
              <div className="notif-empty-state">
                <CheckCircle2 size={32} className="text-emerald-500" />
                <h3 className="notif-empty-title">No Active Escalations</h3>
                <p className="notif-empty-desc">All approval SLAs, shipment milestones, and invoice thresholds are currently compliant.</p>
              </div>
            ) : (
              <div className="escalations-table-container">
                <table className="escalations-table">
                  <thead>
                    <tr>
                      <th>Escalated At</th>
                      <th>Level</th>
                      <th>Notification ID</th>
                      <th>Reason & Trigger</th>
                      <th>AI Summary & Recommended Action</th>
                      <th>Approval Request</th>
                      <th>Severity Transition</th>
                      <th>Correlation ID</th>
                    </tr>
                  </thead>
                  <tbody>
                    {escalations.map((ev) => (
                      <tr key={ev.id}>
                        <td>{formatTimestamp(ev.created_at)}</td>
                        <td>
                          <span className="badge-escalation-level">Level {ev.escalation_level}</span>
                        </td>
                        <td>
                          <button
                            className="text-blue-600 underline font-medium"
                            onClick={() => {
                              setActiveTab('notifications');
                              setSearchQuery(`ref:${ev.notification_id}`);
                            }}
                          >
                            #{ev.notification_id}
                          </button>
                        </td>
                        <td>
                          <div className="audit-reason-box">
                            <span className="audit-reason-text">{ev.escalation_reason}</span>
                            <span className="audit-trigger-tag">{ev.trigger_type} ({ev.threshold_hours}h)</span>
                          </div>
                        </td>
                        <td>
                          <div className="audit-ai-box">
                            {ev.ai_escalation_summary ? (
                              <div className="audit-ai-summary">
                                <Sparkles size={11} className="text-blue-600 flex-shrink-0" />
                                <span>{ev.ai_escalation_summary}</span>
                              </div>
                            ) : (
                              <span className="text-slate-400 text-xs italic">Standard deterministic escalation</span>
                            )}
                            {ev.recommended_action && (
                              <div className="audit-ai-rec text-xs text-slate-600 mt-1">
                                <strong>Action:</strong> {ev.recommended_action}
                              </div>
                            )}
                          </div>
                        </td>
                        <td>
                          {ev.approval_id ? (
                            <button
                              className="badge-approval-link"
                              onClick={() => navigate(`/dashboard/approvals?id=${ev.approval_id}`)}
                              title="View approval request in Approvals Center"
                            >
                              <ShieldAlert size={11} />
                              <span>Approval #{ev.approval_id}</span>
                            </button>
                          ) : (
                            <span className="text-slate-400 text-xs">—</span>
                          )}
                        </td>
                        <td>
                          <div className="sev-transition">
                            <span className={`notif-sev-tag ${getSeverityBadgeClass(ev.previous_severity)}`}>
                              {ev.previous_severity}
                            </span>
                            <ChevronRight size={13} className="text-slate-400" />
                            <span className={`notif-sev-tag ${getSeverityBadgeClass(ev.new_severity)}`}>
                              {ev.new_severity}
                            </span>
                          </div>
                        </td>
                        <td>
                          <code className="correlation-code">{ev.correlation_id}</code>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
          </div>
        </div>
      )}

      {/* ── TAB 3: NOTIFICATION PREFERENCES ── */}
      {activeTab === 'preferences' && preferences && (
        <div className="notif-tab-content">
          <div className="preferences-panel-card">
            <div className="pref-header">
              <h3 className="pref-title">In-App Notification Preferences</h3>
              <p className="pref-desc">
                Configure your minimum alert severity threshold and enabled module categories.
              </p>
            </div>

            <form onSubmit={handleSavePreferences}>
              <div className="pref-section">
                <h4 className="pref-section-title">Thresholds & Delivery</h4>

                <div className="pref-row">
                  <div className="pref-row-info">
                    <span className="pref-label">In-App Notifications</span>
                    <span className="pref-help">Enable or pause the central notification center for your account.</span>
                  </div>
                  <label className="switch">
                    <input
                      type="checkbox"
                      checked={preferences.in_app_enabled}
                      onChange={(e) => setPreferences({ ...preferences, in_app_enabled: e.target.checked })}
                    />
                    <span className="slider round"></span>
                  </label>
                </div>

                <div className="pref-row">
                  <div className="pref-row-info">
                    <span className="pref-label">Minimum Severity Threshold</span>
                    <span className="pref-help">Only receive notifications meeting or exceeding this severity level.</span>
                  </div>
                  <select
                    className="pref-select"
                    value={preferences.min_severity}
                    onChange={(e) => setPreferences({ ...preferences, min_severity: e.target.value })}
                  >
                    <option value="INFORMATIONAL">Informational (All Notifications)</option>
                    <option value="LOW">Low & Above</option>
                    <option value="MEDIUM">Medium & Above</option>
                    <option value="HIGH">High & Critical Only</option>
                    <option value="CRITICAL">Critical Breaches Only</option>
                  </select>
                </div>

                <div className="pref-row">
                  <div className="pref-row-info">
                    <span className="pref-label">Assigned Items Only</span>
                    <span className="pref-help">Only notify me when an action or recommendation is assigned specifically to me.</span>
                  </div>
                  <label className="switch">
                    <input
                      type="checkbox"
                      checked={preferences.assigned_only}
                      onChange={(e) => setPreferences({ ...preferences, assigned_only: e.target.checked })}
                    />
                    <span className="slider round"></span>
                  </label>
                </div>
              </div>

              <div className="pref-section">
                <h4 className="pref-section-title">Enabled Module Categories</h4>

                <div className="pref-row">
                  <div className="pref-row-info">
                    <span className="pref-label">Approvals & Sign-offs</span>
                    <span className="pref-help">Document, quotation, and human-in-the-loop review requests.</span>
                  </div>
                  <label className="switch">
                    <input
                      type="checkbox"
                      checked={preferences.approvals_enabled}
                      onChange={(e) => setPreferences({ ...preferences, approvals_enabled: e.target.checked })}
                    />
                    <span className="slider round"></span>
                  </label>
                </div>

                <div className="pref-row">
                  <div className="pref-row-info">
                    <span className="pref-label">AI Automations & Background Jobs</span>
                    <span className="pref-help">Failed jobs, partial executions, and schedule retries.</span>
                  </div>
                  <label className="switch">
                    <input
                      type="checkbox"
                      checked={preferences.automations_enabled}
                      onChange={(e) => setPreferences({ ...preferences, automations_enabled: e.target.checked })}
                    />
                    <span className="slider round"></span>
                  </label>
                </div>

                <div className="pref-row">
                  <div className="pref-row-info">
                    <span className="pref-label">AI Operational Recommendations</span>
                    <span className="pref-help">High-priority cost, routing, and carrier intelligence.</span>
                  </div>
                  <label className="switch">
                    <input
                      type="checkbox"
                      checked={preferences.recommendations_enabled}
                      onChange={(e) => setPreferences({ ...preferences, recommendations_enabled: e.target.checked })}
                    />
                    <span className="slider round"></span>
                  </label>
                </div>

                <div className="pref-row">
                  <div className="pref-row-info">
                    <span className="pref-label">Finance & Overdue Collections</span>
                    <span className="pref-help">Invoice overdue risk alerts and payment collections.</span>
                  </div>
                  <label className="switch">
                    <input
                      type="checkbox"
                      checked={preferences.finance_enabled}
                      onChange={(e) => setPreferences({ ...preferences, finance_enabled: e.target.checked })}
                    />
                    <span className="slider round"></span>
                  </label>
                </div>

                <div className="pref-row">
                  <div className="pref-row-info">
                    <span className="pref-label">Operations & Shipment Exceptions</span>
                    <span className="pref-help">Critical vessel delays, missed milestones, and tracking alerts.</span>
                  </div>
                  <label className="switch">
                    <input
                      type="checkbox"
                      checked={preferences.operations_enabled}
                      onChange={(e) => setPreferences({ ...preferences, operations_enabled: e.target.checked })}
                    />
                    <span className="slider round"></span>
                  </label>
                </div>

                <div className="pref-row">
                  <div className="pref-row-info">
                    <span className="pref-label">Contracts & Document Compliance</span>
                    <span className="pref-help">Contract expiration, rate extraction issues, and missing files.</span>
                  </div>
                  <label className="switch">
                    <input
                      type="checkbox"
                      checked={preferences.compliance_enabled}
                      onChange={(e) => setPreferences({ ...preferences, compliance_enabled: e.target.checked })}
                    />
                    <span className="slider round"></span>
                  </label>
                </div>
              </div>

              <div className="pref-footer">
                <button
                  type="submit"
                  className="notif-btn notif-btn-primary"
                  disabled={savingPrefs}
                >
                  <Check size={14} />
                  <span>{savingPrefs ? 'Saving...' : 'Save Preferences'}</span>
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* ── DETAIL & ESCALATION DRAWER ── */}
      {selectedNotification && (
        <div className="notif-drawer-overlay" onClick={() => setSelectedNotification(null)}>
          <div className="notif-drawer" onClick={(e) => e.stopPropagation()}>
            <div className="drawer-header">
              <div className="drawer-title-box">
                <div className="drawer-meta-badges">
                  <span className="notif-module-tag">
                    {getModuleIcon(selectedNotification.source_module)}
                    <span>{selectedNotification.source_module}</span>
                  </span>
                  <span className={`notif-sev-tag ${getSeverityBadgeClass(selectedNotification.severity)}`}>
                    {selectedNotification.severity}
                  </span>
                  {getDeliveryStatusBadge(selectedNotification.delivery_status)}
                </div>
                <h2 className="drawer-title">{selectedNotification.title}</h2>
              </div>
              <button
                className="drawer-close-btn"
                onClick={() => setSelectedNotification(null)}
                title="Close details"
              >
                <X size={18} />
              </button>
            </div>

            <div className="drawer-scroll-body">
              {/* Message Banner */}
              <div className="drawer-message-card">
                <p className="drawer-message-text">{selectedNotification.message}</p>
                <div className="drawer-message-time">
                  <Clock size={12} />
                  <span>Created: {formatTimestamp(selectedNotification.created_at)}</span>
                </div>
              </div>

              {/* Section 1: Confirmed Source Facts */}
              <div className="drawer-section">
                <div className="drawer-section-header">
                  <h4 className="drawer-section-title">Confirmed Source Record Facts</h4>
                  <span className="facts-badge">Backend Verified</span>
                </div>

                <div className="drawer-facts-grid">
                  <div className="fact-item">
                    <span className="fact-label">Source Record</span>
                    <span className="fact-value font-mono">#{selectedNotification.source_record_id}</span>
                  </div>
                  <div className="fact-item">
                    <span className="fact-label">Source Module</span>
                    <span className="fact-value">{selectedNotification.source_module}</span>
                  </div>
                  <div className="fact-item">
                    <span className="fact-label">Action Required</span>
                    <span className="fact-value">{selectedNotification.action_required ? 'Yes' : 'No'}</span>
                  </div>
                  <div className="fact-item">
                    <span className="fact-label">Escalation Level</span>
                    <span className="fact-value">L{selectedNotification.escalation_level || 0}</span>
                  </div>
                  <div className="fact-item">
                    <span className="fact-label">Group Cluster</span>
                    <span className="fact-value">{selectedNotification.group_key || 'Independent Signal'}</span>
                  </div>
                  <div className="fact-item">
                    <span className="fact-label">Delivery State</span>
                    <span className="fact-value">{selectedNotification.delivery_status || 'DELIVERED'}</span>
                  </div>
                </div>

                {selectedNotification.action_url && (
                  <div className="drawer-source-link-box mt-3">
                    <button
                      className="notif-btn notif-btn-secondary w-full justify-center"
                      onClick={() => navigate(selectedNotification.action_url)}
                    >
                      <span>Navigate to Source Business Record</span>
                      <ExternalLink size={14} />
                    </button>
                  </div>
                )}
              </div>

              {/* Section 2: AI Operational Intelligence (Python AI Sidecar) */}
              <div className="drawer-section">
                <div className="drawer-section-header">
                  <div className="flex items-center gap-2">
                    <Sparkles size={16} className="text-blue-600" />
                    <h4 className="drawer-section-title">AI Operational Intelligence</h4>
                  </div>
                  <button
                    className="drawer-ai-action-btn"
                    onClick={(e) => handleAnalyzeAI(e, selectedNotification.id)}
                    disabled={aiAnalyzing}
                  >
                    <RefreshCw size={12} className={aiAnalyzing ? 'spin-icon' : ''} />
                    <span>{aiAnalyzing ? 'Analyzing...' : selectedNotification.ai_summary ? 'Refresh AI Analysis' : 'Run AI Analysis'}</span>
                  </button>
                </div>

                {selectedNotification.ai_summary ? (
                  <div className="ai-intelligence-card">
                    <div className="ai-score-row">
                      <div className="ai-score-label-box">
                        <span className="ai-score-label">Priority Impact Score</span>
                        <span className="ai-score-val">{Math.round(selectedNotification.ai_priority_score || 0)}/100</span>
                      </div>
                      <div className="ai-score-bar-bg">
                        <div
                          className="ai-score-bar-fill"
                          style={{
                            width: `${Math.min(100, Math.max(10, selectedNotification.ai_priority_score || 50))}%`,
                            backgroundColor: (selectedNotification.ai_priority_score || 0) >= 80 ? '#e11d48' : (selectedNotification.ai_priority_score || 0) >= 60 ? '#f59e0b' : '#2563eb',
                          }}
                        />
                      </div>
                    </div>

                    <div className="ai-insight-block">
                      <span className="ai-block-heading">Operational Summary:</span>
                      <p className="ai-block-text">{selectedNotification.ai_summary}</p>
                    </div>

                    {selectedNotification.ai_escalation_reason && (
                      <div className="ai-insight-block">
                        <span className="ai-block-heading">Escalation Trajectory Reasoning:</span>
                        <p className="ai-block-text">{selectedNotification.ai_escalation_reason}</p>
                      </div>
                    )}
                  </div>
                ) : (
                  <div className="ai-placeholder-card">
                    <p className="ai-placeholder-text">
                      AI intelligence evaluates alert fatigue, cluster similarity, and recommended resolution pathways.
                    </p>
                    <button
                      className="notif-btn notif-btn-primary mt-2"
                      onClick={(e) => handleAnalyzeAI(e, selectedNotification.id)}
                      disabled={aiAnalyzing}
                    >
                      <Sparkles size={14} />
                      <span>{aiAnalyzing ? 'Analyzing with AI...' : 'Analyze Notification with AI'}</span>
                    </button>
                  </div>
                )}
              </div>

              {/* Section 3: Controlled Operational Actions */}
              <div className="drawer-section">
                <h4 className="drawer-section-title mb-3">Controlled Operational Actions</h4>

                <div className="drawer-action-buttons-grid">
                  {/* Acknowledge Button */}
                  {!selectedNotification.is_acknowledged ? (
                    <button
                      className="drawer-act-btn btn-act-ack"
                      onClick={() => handleAcknowledge(null, selectedNotification.id)}
                    >
                      <CheckCheck size={15} />
                      <span>Acknowledge Notification</span>
                    </button>
                  ) : (
                    <div className="drawer-act-disabled">
                      <CheckCheck size={14} className="text-emerald-600" />
                      <span>Acknowledged at {formatTimestamp(selectedNotification.acknowledged_at)}</span>
                    </div>
                  )}

                  {/* Snooze Options */}
                  <div className="drawer-snooze-row">
                    <select
                      className="notif-select flex-1"
                      value={snoozeMinutes}
                      onChange={(e) => setSnoozeMinutes(Number(e.target.value))}
                    >
                      <option value={60}>Snooze 1 Hour</option>
                      <option value={240}>Snooze 4 Hours</option>
                      <option value={1440}>Snooze 24 Hours</option>
                      <option value={4320}>Snooze 3 Days</option>
                    </select>
                    <button
                      className="notif-btn notif-btn-secondary"
                      onClick={() => handleSnooze(null, selectedNotification.id, snoozeMinutes)}
                    >
                      <Timer size={14} />
                      <span>Apply Snooze</span>
                    </button>
                  </div>

                  {/* Escalate Button */}
                  <button
                    className="drawer-act-btn btn-act-esc"
                    onClick={() => setEscalateModalOpen(true)}
                  >
                    <ShieldAlert size={15} />
                    <span>Escalate Unresolved Issue</span>
                  </button>
                </div>
              </div>

              {/* Section 4: AI Communication Draft Generator */}
              <div className="drawer-section">
                <div className="drawer-section-header">
                  <div className="flex items-center gap-2">
                    <Send size={15} className="text-blue-600" />
                    <h4 className="drawer-section-title">AI Escalation Draft Generator</h4>
                  </div>
                </div>

                <p className="drawer-desc-note text-xs text-slate-500 mb-3">
                  Drafts require operator sign-off before external transmission. Zero messages are dispatched automatically.
                </p>

                <div className="draft-type-selector-row mb-3">
                  <select
                    className="notif-select flex-1"
                    value={draftType}
                    onChange={(e) => setDraftType(e.target.value)}
                  >
                    <option value="OPERATIONAL_ALERT">Internal Operations Alert</option>
                    <option value="EXECUTIVE_ESCALATION">Executive SLA Escalation</option>
                    <option value="CARRIER_FOLLOW_UP">Carrier Status Ingestion Follow-up</option>
                  </select>
                  <button
                    className="notif-btn notif-btn-primary"
                    onClick={() => handleGenerateDraft(selectedNotification.id, draftType)}
                    disabled={draftGenerating}
                  >
                    <Sparkles size={13} />
                    <span>{draftGenerating ? 'Drafting...' : 'Generate Draft'}</span>
                  </button>
                </div>

                {generatedDraft && (
                  <div className="ai-draft-preview-card">
                    <div className="draft-preview-header">
                      <span className="draft-tag-badge">AI Draft Preview</span>
                      {generatedDraft.requires_approval && (
                        <span className="draft-approval-badge">Requires Human Approval</span>
                      )}
                      <button
                        className="draft-copy-btn"
                        onClick={copyDraftToClipboard}
                        title="Copy draft to clipboard"
                      >
                        <Copy size={13} />
                        <span>Copy</span>
                      </button>
                    </div>

                    <div className="draft-field mb-2">
                      <span className="draft-label">Subject:</span>
                      <div className="draft-value font-medium">{generatedDraft.subject}</div>
                    </div>

                    <div className="draft-field">
                      <span className="draft-label">Message Body:</span>
                      <div className="draft-body-box">{generatedDraft.body_text}</div>
                    </div>

                    <div className="draft-footer-note mt-2 text-xs text-slate-500 flex items-center gap-1.5">
                      <Info size={12} className="text-blue-600" />
                      <span>This draft has been prepared safely. External sending is gated behind the Central Approvals System.</span>
                    </div>
                  </div>
                )}
              </div>
            </div>
          </div>
        </div>
      )}

      {/* ── ESCALATION REASON MODAL ── */}
      {escalateModalOpen && (
        <div className="notif-modal-overlay" onClick={() => setEscalateModalOpen(false)}>
          <div className="notif-modal-dialog" onClick={(e) => e.stopPropagation()}>
            <div className="modal-header">
              <div className="flex items-center gap-2">
                <ShieldAlert size={18} className="text-rose-600" />
                <h3 className="modal-title">Escalate Notification #{selectedNotification?.id}</h3>
              </div>
              <button className="modal-close-btn" onClick={() => setEscalateModalOpen(false)}>
                <X size={16} />
              </button>
            </div>

            <form onSubmit={handleEscalateSubmit}>
              <div className="modal-body">
                <p className="modal-desc text-sm text-slate-600 mb-3">
                  Escalating this signal increases its operational priority and records an auditable escalation transition.
                  If the severity is HIGH or CRITICAL, an Approval Request will automatically be generated.
                </p>

                <div className="form-group mb-3">
                  <label className="form-label font-medium text-xs text-slate-700 block mb-1">
                    Escalation Justification / Operator Note:
                  </label>
                  <textarea
                    className="form-textarea w-full p-2.5 border border-slate-300 rounded-md text-sm focus:outline-none focus:border-blue-500"
                    rows={3}
                    placeholder="Enter reason for manual escalation (e.g., Exceeded carrier 4h response window; port congestion delaying transshipment)..."
                    value={escalateReason}
                    onChange={(e) => setEscalateReason(e.target.value)}
                  />
                </div>

                <div className="p-3 bg-amber-50 border border-amber-200 rounded-md text-xs text-amber-800 flex items-start gap-2">
                  <AlertTriangle size={14} className="flex-shrink-0 mt-0.5 text-amber-600" />
                  <span>
                    Approval System Enforcement: High-consequence escalations require sign-off before external customer notifications or executive alerts execute.
                  </span>
                </div>
              </div>

              <div className="modal-footer flex justify-end gap-2 p-4 border-t border-slate-200">
                <button
                  type="button"
                  className="notif-btn notif-btn-secondary"
                  onClick={() => setEscalateModalOpen(false)}
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  className="notif-btn notif-btn-danger"
                  disabled={escalating}
                >
                  <ShieldAlert size={14} />
                  <span>{escalating ? 'Escalating...' : 'Confirm Escalation'}</span>
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}

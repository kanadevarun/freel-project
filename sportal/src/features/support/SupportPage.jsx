import React, { useState, useEffect, useMemo } from 'react';
import { useSearchParams, Link } from 'react-router-dom';
import {
  Headphones,
  Bell,
  Activity,
  ShieldCheck,
  Search,
  RefreshCw,
  AlertCircle,
  AlertTriangle,
  Clock,
  User,
  CheckCircle2,
  Filter,
  Eye,
  ChevronRight,
  MessageSquare,
  ArrowRight,
  ExternalLink,
  X,
  Send,
  Lock,
  Plus,
  FileText,
  Check,
  Mail,
  Building2,
  Terminal,
  Share2,
} from 'lucide-react';
import PageHeader from '../../components/common/PageHeader';
import KpiCard from '../../components/common/KpiCard';
import StatusBadge from '../../components/common/StatusBadge';
import EmptyState from '../../components/common/EmptyState';
import { sportalService } from '../../services/sportalService';

export function SupportPage() {
  const [searchParams, setSearchParams] = useSearchParams();
  const initialTab = searchParams.get('tab') || 'cases';
  const initialOrgId = searchParams.get('orgId') || 'ALL';

  const [activeTab, setActiveTab] = useState(initialTab);
  const [selectedOrgId, setSelectedOrgId] = useState(initialOrgId);
  const [organizations, setOrganizations] = useState([]);

  // Subsystem 1: Support Cases State
  const [casesOverview, setCasesOverview] = useState(null);
  const [casesLoading, setCasesLoading] = useState(true);
  const [caseStatusFilter, setCaseStatusFilter] = useState('ALL');
  const [caseSeverityFilter, setCaseSeverityFilter] = useState('ALL');
  const [caseSearch, setCaseSearch] = useState('');
  const [selectedCase, setSelectedCase] = useState(null);
  const [caseDetailLoading, setCaseDetailLoading] = useState(false);
  const [caseDetailModalOpen, setCaseDetailModalOpen] = useState(false);
  const [updatingCaseStatus, setUpdatingCaseStatus] = useState(false);
  const [newStatus, setNewStatus] = useState('');
  const [resolutionNotes, setResolutionNotes] = useState('');
  const [newNoteContent, setNewNoteContent] = useState('');
  const [submittingNote, setSubmittingNote] = useState(false);
  const [actionSuccessMsg, setActionSuccessMsg] = useState('');

  // Subsystem 2: Notifications Center State
  const [notificationsOverview, setNotificationsOverview] = useState(null);
  const [notificationsLoading, setNotificationsLoading] = useState(true);
  const [notifReadFilter, setNotifReadFilter] = useState('ALL');
  const [notifSeverityFilter, setNotifSeverityFilter] = useState('ALL');

  // Subsystem 3: Activity Timeline State
  const [activityItems, setActivityItems] = useState([]);
  const [activityLoading, setActivityLoading] = useState(true);
  const [activityCategoryFilter, setActivityCategoryFilter] = useState('ALL');

  // Subsystem 4: Forensic Audit State
  const [auditLogs, setAuditLogs] = useState([]);
  const [auditLoading, setAuditLoading] = useState(true);
  const [auditSearch, setAuditSearch] = useState('');
  const [auditModuleFilter, setAuditModuleFilter] = useState('ALL');
  const [auditResultFilter, setAuditResultFilter] = useState('ALL');
  const [selectedAuditLog, setSelectedAuditLog] = useState(null);

  // Load organizations for customer switcher
  useEffect(() => {
    async function loadOrgs() {
      try {
        const res = await sportalService.getOrganizations({ pageSize: 100 });
        const items = res?.items || res?.data?.items || res?.data?.organizations || (Array.isArray(res?.data) ? res.data : []);
        setOrganizations(items);
      } catch (err) {
        console.error('Failed to load organizations in support page:', err);
      }
    }
    loadOrgs();
  }, []);

  // Sync tab with URL
  useEffect(() => {
    const tab = searchParams.get('tab');
    if (tab && tab !== activeTab) {
      setActiveTab(tab);
    }
    const org = searchParams.get('orgId');
    if (org && org !== selectedOrgId) {
      setSelectedOrgId(org);
    }
    const caseId = searchParams.get('caseId');
    if (caseId) {
      handleInspectCase(parseInt(caseId, 10));
    }
  }, [searchParams]);

  const handleTabChange = (newTab) => {
    setActiveTab(newTab);
    const params = new URLSearchParams(searchParams);
    params.set('tab', newTab);
    setSearchParams(params);
  };

  const handleOrgScopeChange = (orgId) => {
    setSelectedOrgId(orgId);
    const params = new URLSearchParams(searchParams);
    if (orgId === 'ALL') {
      params.delete('orgId');
    } else {
      params.set('orgId', orgId);
    }
    setSearchParams(params);
  };

  // 1. Fetch Support Cases
  const loadCases = async () => {
    setCasesLoading(true);
    try {
      const params = {
        orgId: selectedOrgId !== 'ALL' ? selectedOrgId : undefined,
        status: caseStatusFilter !== 'ALL' ? caseStatusFilter : undefined,
        severity: caseSeverityFilter !== 'ALL' ? caseSeverityFilter : undefined,
        search: caseSearch || undefined,
        limit: 50,
      };
      const res = await sportalService.getSupportCases(params);
      setCasesOverview(res?.data || res);
    } catch (err) {
      console.error('Failed to load support cases:', err);
    } finally {
      setCasesLoading(false);
    }
  };

  // 2. Fetch Notifications
  const loadNotifications = async () => {
    setNotificationsLoading(true);
    try {
      const params = {
        orgId: selectedOrgId !== 'ALL' ? selectedOrgId : undefined,
        isRead: notifReadFilter === 'UNREAD' ? false : notifReadFilter === 'READ' ? true : undefined,
        severity: notifSeverityFilter !== 'ALL' ? notifSeverityFilter : undefined,
        limit: 50,
      };
      const res = await sportalService.getNotifications(params);
      setNotificationsOverview(res?.data || res);
    } catch (err) {
      console.error('Failed to load notifications:', err);
    } finally {
      setNotificationsLoading(false);
    }
  };

  // 3. Fetch Activity Timeline
  const loadActivity = async () => {
    setActivityLoading(true);
    try {
      const params = {
        orgId: selectedOrgId !== 'ALL' ? selectedOrgId : undefined,
        category: activityCategoryFilter !== 'ALL' ? activityCategoryFilter : undefined,
        limit: 40,
      };
      const res = await sportalService.getUnifiedActivityTimeline(params);
      const items = Array.isArray(res) ? res : (res?.items || res?.data || []);
      setActivityItems(items);
    } catch (err) {
      console.error('Failed to load activity timeline:', err);
    } finally {
      setActivityLoading(false);
    }
  };

  // 4. Fetch Forensic Audit
  const loadAuditLogs = async () => {
    setAuditLoading(true);
    try {
      const params = {
        orgId: selectedOrgId !== 'ALL' ? selectedOrgId : undefined,
        module: auditModuleFilter !== 'ALL' ? auditModuleFilter : undefined,
        result: auditResultFilter !== 'ALL' ? auditResultFilter : undefined,
        search: auditSearch || undefined,
        limit: 50,
      };
      const res = await sportalService.searchAuditLogs(params);
      const logs = res?.items || (Array.isArray(res) ? res : (res?.data || []));
      setAuditLogs(logs);
    } catch (err) {
      console.error('Failed to load audit logs:', err);
    } finally {
      setAuditLoading(false);
    }
  };

  // Trigger loads based on active tab and org
  useEffect(() => {
    if (activeTab === 'cases') loadCases();
    if (activeTab === 'notifications') loadNotifications();
    if (activeTab === 'activity') loadActivity();
    if (activeTab === 'audit') loadAuditLogs();
  }, [activeTab, selectedOrgId, caseStatusFilter, caseSeverityFilter, notifReadFilter, notifSeverityFilter, activityCategoryFilter, auditModuleFilter, auditResultFilter]);

  // Open Support Case Detail
  const handleInspectCase = async (caseId) => {
    setCaseDetailLoading(true);
    setCaseDetailModalOpen(true);
    setActionSuccessMsg('');
    try {
      const res = await sportalService.getSupportCaseDetail(caseId);
      const c = res?.data || res;
      setSelectedCase(c);
      setNewStatus(c.status);
      setResolutionNotes(c.resolution_notes || '');
    } catch (err) {
      console.error('Failed to get case detail:', err);
    } finally {
      setCaseDetailLoading(false);
    }
  };

  // Submit Status Change
  const handleUpdateStatus = async (e) => {
    e.preventDefault();
    if (!selectedCase) return;
    setUpdatingCaseStatus(true);
    setActionSuccessMsg('');
    try {
      await sportalService.updateSupportCaseStatus(selectedCase.id, {
        status: newStatus,
        resolution_notes: resolutionNotes,
      });
      setActionSuccessMsg(`Status successfully updated to ${newStatus}`);
      // Refresh case detail and overview
      const updated = await sportalService.getSupportCaseDetail(selectedCase.id);
      setSelectedCase(updated?.data || updated);
      loadCases();
    } catch (err) {
      console.error('Failed to update case status:', err);
      alert('Failed to update case: ' + (err.message || 'Unknown error'));
    } finally {
      setUpdatingCaseStatus(false);
    }
  };

  // Submit Internal Note
  const handleAddNote = async (e) => {
    e.preventDefault();
    if (!selectedCase || !newNoteContent.trim()) return;
    setSubmittingNote(true);
    try {
      await sportalService.addSupportCaseNote(selectedCase.id, {
        content: newNoteContent,
        is_internal_only: true,
      });
      setNewNoteContent('');
      setActionSuccessMsg('Internal support note recorded successfully');
      const updated = await sportalService.getSupportCaseDetail(selectedCase.id);
      setSelectedCase(updated?.data || updated);
    } catch (err) {
      console.error('Failed to add note:', err);
      alert('Failed to add note: ' + (err.message || 'Unknown error'));
    } finally {
      setSubmittingNote(false);
    }
  };

  // Notification Actions
  const handleMarkNotificationRead = async (notifId) => {
    try {
      await sportalService.markNotificationRead(notifId);
      loadNotifications();
    } catch (err) {
      console.error('Failed to mark read:', err);
    }
  };

  const handleAcknowledgeNotification = async (notifId) => {
    try {
      await sportalService.acknowledgeNotification(notifId);
      loadNotifications();
    } catch (err) {
      console.error('Failed to acknowledge notification:', err);
    }
  };

  const handleMarkAllRead = async () => {
    try {
      await sportalService.markAllNotificationsRead(selectedOrgId !== 'ALL' ? selectedOrgId : null);
      loadNotifications();
    } catch (err) {
      console.error('Failed to mark all read:', err);
    }
  };

  // Priority color helper
  const getPriorityBadgeClass = (priority) => {
    const p = String(priority || '').toUpperCase();
    switch (p) {
      case 'CRITICAL':
        return 'bg-rose-50 text-rose-700 border-rose-200';
      case 'HIGH':
        return 'bg-amber-50 text-amber-700 border-amber-200';
      case 'MEDIUM':
        return 'bg-blue-50 text-blue-700 border-blue-200';
      default:
        return 'bg-slate-50 text-slate-700 border-slate-200';
    }
  };

  // Status color helper
  const getStatusBadgeClass = (status) => {
    const s = String(status || '').toUpperCase();
    switch (s) {
      case 'RESOLVED':
        return 'bg-emerald-50 text-emerald-700 border-emerald-200';
      case 'IN_PROGRESS':
        return 'bg-blue-50 text-blue-700 border-blue-200';
      case 'ACKNOWLEDGED':
        return 'bg-purple-50 text-purple-700 border-purple-200';
      case 'DISMISSED':
        return 'bg-slate-100 text-slate-600 border-slate-200';
      default:
        return 'bg-amber-50 text-amber-700 border-amber-200';
    }
  };

  // Delivery status color helper
  const getDeliveryStatusBadgeClass = (status) => {
    const s = String(status || '').toUpperCase();
    switch (s) {
      case 'DELIVERED':
        return 'bg-emerald-50 text-emerald-700 border-emerald-200';
      case 'ESCALATED':
        return 'bg-rose-50 text-rose-700 border-rose-200';
      case 'QUEUED':
      case 'PENDING':
        return 'bg-amber-50 text-amber-700 border-amber-200';
      case 'FAILED':
        return 'bg-red-100 text-red-800 border-red-200';
      default:
        return 'bg-slate-50 text-slate-600 border-slate-200';
    }
  };

  // Format date helper
  const formatDate = (dStr) => {
    if (!dStr) return '—';
    try {
      const d = new Date(dStr);
      return isNaN(d.getTime()) ? '—' : d.toLocaleString('en-US', {
        month: 'short',
        day: 'numeric',
        year: 'numeric',
        hour: '2-digit',
        minute: '2-digit',
      });
    } catch {
      return '—';
    }
  };

  const currentOrg = organizations.find((o) => String(o.id) === String(selectedOrgId));

  return (
    <div className="p-6 sm:p-8 max-w-7xl mx-auto space-y-6 animate-fade-in">
      {/* 1. Standard Page Header with Customer Switcher */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <div className="flex items-center gap-2 text-xs font-semibold text-slate-500 mb-1">
            <Link to="/" className="hover:text-blue-600 transition">SPortal</Link>
            <ChevronRight className="w-3.5 h-3.5" />
            <span>Support & Operations Control</span>
          </div>
          <div className="flex items-center gap-3">
            <h1 className="text-2xl font-bold text-slate-900 tracking-tight">Support, Activity & Governance Center</h1>
            <span className="px-2 py-0.5 text-xs font-semibold rounded-full bg-blue-50 text-blue-700 border border-blue-200">
              Multi-Customer Operations
            </span>
          </div>
          <p className="text-xs text-slate-500 mt-1">
            Unified operational support triage, proactive customer communications, real-time activity stream, and immutable audit trail.
          </p>
        </div>

        {/* Global Customer Scope Selector */}
        <div className="flex items-center gap-3 bg-white p-2 rounded-xl border border-slate-200 shadow-2xs">
          <Building2 className="w-4 h-4 text-slate-400 shrink-0" />
          <div className="flex flex-col">
            <label htmlFor="customer-scope-select" className="text-[10px] uppercase font-bold text-slate-400">Customer Scope</label>
            <select
              id="customer-scope-select"
              value={selectedOrgId}
              onChange={(e) => handleOrgScopeChange(e.target.value)}
              className="text-xs font-bold text-slate-800 bg-transparent border-none p-0 focus:ring-0 cursor-pointer pr-4"
            >
              <option value="ALL">All Customer Organizations (Platform Aggregate)</option>
              {organizations.map((org) => (
                <option key={org.id} value={org.id}>
                  {org.name} (ORG-{String(org.id).padStart(4, '0')})
                </option>
              ))}
            </select>
          </div>
          <button
            onClick={() => {
              if (activeTab === 'cases') loadCases();
              if (activeTab === 'notifications') loadNotifications();
              if (activeTab === 'activity') loadActivity();
              if (activeTab === 'audit') loadAuditLogs();
            }}
            className="p-1.5 text-slate-500 hover:text-slate-900 hover:bg-slate-100 rounded-lg transition"
            title="Refresh current view"
          >
            <RefreshCw className="w-4 h-4" />
          </button>
        </div>
      </div>

      {/* 2. Top Navigation Tabs */}
      <div className="border-b border-slate-200 flex items-center justify-between">
        <nav className="flex space-x-6">
          <button
            id="tab-support-cases"
            onClick={() => handleTabChange('cases')}
            className={`pb-3 text-xs font-bold transition-all border-b-2 flex items-center gap-2 cursor-pointer ${
              activeTab === 'cases'
                ? 'border-blue-600 text-blue-600'
                : 'border-transparent text-slate-500 hover:text-slate-800'
            }`}
          >
            <Headphones className="w-4 h-4" />
            <span>Support Cases</span>
            {casesOverview && (
              <span className={`px-1.5 py-0.2 rounded-full text-[10px] font-bold ${
                casesOverview.open_cases > 0 ? 'bg-amber-100 text-amber-800' : 'bg-slate-100 text-slate-600'
              }`}>
                {casesOverview.open_cases} open
              </span>
            )}
          </button>

          <button
            id="tab-notifications"
            onClick={() => handleTabChange('notifications')}
            className={`pb-3 text-xs font-bold transition-all border-b-2 flex items-center gap-2 cursor-pointer ${
              activeTab === 'notifications'
                ? 'border-blue-600 text-blue-600'
                : 'border-transparent text-slate-500 hover:text-slate-800'
            }`}
          >
            <Bell className="w-4 h-4" />
            <span>Notifications Center</span>
            {notificationsOverview && notificationsOverview.unread_count > 0 && (
              <span className="px-1.5 py-0.2 rounded-full text-[10px] font-bold bg-blue-100 text-blue-800">
                {notificationsOverview.unread_count} unread
              </span>
            )}
          </button>

          <button
            id="tab-activity-timeline"
            onClick={() => handleTabChange('activity')}
            className={`pb-3 text-xs font-bold transition-all border-b-2 flex items-center gap-2 cursor-pointer ${
              activeTab === 'activity'
                ? 'border-blue-600 text-blue-600'
                : 'border-transparent text-slate-500 hover:text-slate-800'
            }`}
          >
            <Activity className="w-4 h-4" />
            <span>Unified Activity Timeline</span>
          </button>

          <button
            id="tab-forensic-audit"
            onClick={() => handleTabChange('audit')}
            className={`pb-3 text-xs font-bold transition-all border-b-2 flex items-center gap-2 cursor-pointer ${
              activeTab === 'audit'
                ? 'border-blue-600 text-blue-600'
                : 'border-transparent text-slate-500 hover:text-slate-800'
            }`}
          >
            <ShieldCheck className="w-4 h-4" />
            <span>Forensic Audit Ledger</span>
            <span className="px-1.5 py-0.2 rounded-full text-[10px] font-mono font-bold bg-emerald-50 text-emerald-700 border border-emerald-200">
              Immutable
            </span>
          </button>
        </nav>

        {currentOrg && (
          <div className="hidden md:flex items-center gap-1 text-xs text-slate-500 pb-2">
            <span>Scoped to:</span>
            <Link to={`/organizations/${currentOrg.id}`} className="font-bold text-blue-600 hover:underline flex items-center gap-1">
              <span>{currentOrg.name}</span>
              <ExternalLink className="w-3 h-3" />
            </Link>
          </div>
        )}
      </div>

      {/* ========================================================================= */}
      {/* TAB 1: SUPPORT CASES */}
      {/* ========================================================================= */}
      {activeTab === 'cases' && (
        <div className="space-y-6">
          {/* KPI Summary Cards */}
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-5 gap-3.5">
            <KpiCard
              title="Total Cases Recorded"
              value={casesOverview?.total_cases ?? 0}
              unit="cases"
              subtitle="All historical exceptions & tickets"
              icon={Headphones}
              variant="default"
            />
            <KpiCard
              title="Open Support Cases"
              value={casesOverview?.open_cases ?? 0}
              unit="active"
              subtitle="Requiring coordinator action"
              icon={AlertCircle}
              variant={casesOverview?.open_cases > 0 ? 'warning' : 'success'}
            />
            <KpiCard
              title="Critical Escalations"
              value={casesOverview?.critical_cases ?? 0}
              unit="urgent"
              subtitle="High impact delays & holds"
              icon={AlertTriangle}
              variant={casesOverview?.critical_cases > 0 ? 'danger' : 'success'}
            />
            <KpiCard
              title="Resolved Cases"
              value={casesOverview?.resolved_cases ?? 0}
              unit="resolved"
              subtitle="Closed with verified resolutions"
              icon={CheckCircle2}
              variant="success"
            />
            <KpiCard
              title="Avg Resolution Time"
              value={casesOverview?.avg_resolution_hours ?? 0}
              unit="hours"
              subtitle="Average turnaround to closure"
              icon={Clock}
              variant="primary"
            />
          </div>

          {/* Filter Bar */}
          <div className="bg-white p-4 rounded-xl border border-slate-200 shadow-xs flex flex-wrap items-center justify-between gap-4">
            <div className="flex flex-wrap items-center gap-3">
              {/* Status Tabs */}
              <div className="flex items-center gap-1 bg-slate-100 p-1 rounded-lg">
                {['ALL', 'OPEN', 'IN_PROGRESS', 'RESOLVED', 'DISMISSED'].map((st) => (
                  <button
                    key={st}
                    onClick={() => setCaseStatusFilter(st)}
                    className={`px-3 py-1 text-xs font-semibold rounded-md transition cursor-pointer ${
                      caseStatusFilter === st
                        ? 'bg-white text-slate-900 shadow-2xs font-bold'
                        : 'text-slate-500 hover:text-slate-800'
                    }`}
                  >
                    {st === 'ALL' ? 'All Statuses' : st.replace('_', ' ')}
                  </button>
                ))}
              </div>

              {/* Priority Select */}
              <select
                value={caseSeverityFilter}
                onChange={(e) => setCaseSeverityFilter(e.target.value)}
                className="text-xs bg-slate-50 border border-slate-200 rounded-lg px-3 py-1.5 text-slate-700 font-medium focus:ring-1 focus:ring-blue-500"
              >
                <option value="ALL">All Priorities</option>
                <option value="CRITICAL">Critical</option>
                <option value="HIGH">High</option>
                <option value="MEDIUM">Medium</option>
                <option value="LOW">Low</option>
              </select>
            </div>

            {/* Search Input */}
            <div className="relative w-full sm:w-64">
              <Search className="w-3.5 h-3.5 text-slate-400 absolute left-3 top-1/2 -translate-y-1/2" />
              <input
                type="text"
                placeholder="Search subject, customer..."
                value={caseSearch}
                onChange={(e) => setCaseSearch(e.target.value)}
                onKeyDown={(e) => e.key === 'Enter' && loadCases()}
                className="w-full pl-8 pr-3 py-1.5 text-xs bg-slate-50 border border-slate-200 rounded-lg focus:outline-none focus:ring-1 focus:ring-blue-500"
              />
            </div>
          </div>

          {/* Cases Table */}
          <div className="bg-white rounded-xl border border-slate-200 shadow-xs overflow-hidden">
            {casesLoading ? (
              <div className="p-12 text-center text-xs text-slate-400">Loading authoritative support records...</div>
            ) : !casesOverview?.items || casesOverview.items.length === 0 ? (
              <EmptyState
                icon={Headphones}
                title="No support cases found"
                description={
                  selectedOrgId !== 'ALL'
                    ? `No support cases or exceptions recorded for ${currentOrg?.name || 'this customer'}.`
                    : 'No support cases match the active filter criteria.'
                }
              />
            ) : (
              <div className="overflow-x-auto">
                <table className="w-full text-left text-xs min-w-[940px]">
                  <thead>
                    <tr className="border-b border-slate-200 bg-slate-50/90 text-slate-500 font-semibold uppercase tracking-wider text-[11px]">
                      <th className="py-3 px-3">Case Code</th>
                      <th className="py-3 px-3">Customer</th>
                      <th className="py-3 px-3">Subject & Cargo Entity</th>
                      <th className="py-3 px-3">Category</th>
                      <th className="py-3 px-3">Priority</th>
                      <th className="py-3 px-3">Status</th>
                      <th className="py-3 px-3">SLA State</th>
                      <th className="py-3 px-3">Owner</th>
                      <th className="py-3 px-3">Created</th>
                      <th className="py-3 px-3 text-right sticky right-0 bg-slate-50 shadow-[-4px_0_6px_-2px_rgba(0,0,0,0.06)] z-10">Action</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-slate-100">
                    {casesOverview.items.map((item) => (
                      <tr
                        key={item.id}
                        onClick={() => handleInspectCase(item.id)}
                        className="group hover:bg-blue-50/50 transition cursor-pointer"
                        title="Click to inspect case details and manage lifecycle"
                      >
                        <td className="py-3 px-3 font-mono font-bold text-blue-600 whitespace-nowrap">
                          {item.case_code}
                        </td>
                        <td className="py-3 px-3 font-semibold text-slate-900 max-w-[150px] truncate">
                          <Link
                            to={`/organizations/${item.org_id}`}
                            onClick={(e) => e.stopPropagation()}
                            className="hover:text-blue-600 transition flex items-center gap-1"
                          >
                            <span className="truncate">{item.org_name}</span>
                          </Link>
                        </td>
                        <td className="py-3 px-3 max-w-[240px]">
                          <div className="font-semibold text-slate-900 truncate" title={item.subject}>{item.subject}</div>
                          <div className="text-[11px] text-slate-400 font-mono flex items-center gap-1">
                            <span>Linked:</span>
                            <span className="text-slate-600 font-medium">{item.shipment_ref}</span>
                          </div>
                        </td>
                        <td className="py-3 px-3 whitespace-nowrap">
                          <span className="px-2 py-0.5 text-[10px] font-mono rounded bg-slate-100 text-slate-700">
                            {item.category}
                          </span>
                        </td>
                        <td className="py-3 px-3 whitespace-nowrap">
                          <span className={`px-2 py-0.5 rounded-full text-[10px] font-bold border ${getPriorityBadgeClass(item.priority)}`}>
                            {item.priority}
                          </span>
                        </td>
                        <td className="py-3 px-3 whitespace-nowrap">
                          <span className={`px-2 py-0.5 rounded-full text-[10px] font-bold border ${getStatusBadgeClass(item.status)}`}>
                            {item.status.replace('_', ' ')}
                          </span>
                        </td>
                        <td className="py-3 px-3 whitespace-nowrap">
                          <span className={`px-2 py-0.5 rounded text-[10px] font-bold ${
                            item.sla_state === 'MET'
                              ? 'bg-emerald-50 text-emerald-700 border border-emerald-200'
                              : item.sla_state === 'BREACHED'
                              ? 'bg-rose-50 text-rose-700 border border-rose-200 font-black animate-pulse'
                              : item.sla_state === 'WARNING'
                              ? 'bg-amber-50 text-amber-700 border border-amber-200'
                              : 'bg-slate-50 text-slate-600 border border-slate-200'
                          }`}>
                            {item.sla_state === 'MET' ? '✓ SLA Met' : item.sla_state === 'BREACHED' ? '⚠️ SLA Breached' : item.sla_state === 'WARNING' ? '⏳ At Risk' : 'Normal SLA'}
                          </span>
                        </td>
                        <td className="py-3 px-3 text-slate-600 truncate max-w-[120px]">
                          {item.assigned_owner || 'Unassigned'}
                        </td>
                        <td className="py-3 px-3 text-slate-500 whitespace-nowrap">
                          {formatDate(item.created_at)}
                        </td>
                        <td className="py-3 px-3 text-right whitespace-nowrap sticky right-0 bg-white group-hover:bg-blue-50/70 transition shadow-[-4px_0_6px_-2px_rgba(0,0,0,0.06)] z-10">
                          <button
                            onClick={(e) => {
                              e.stopPropagation();
                              handleInspectCase(item.id);
                            }}
                            className="inline-flex items-center gap-1.5 px-3 py-1.5 text-xs font-bold rounded-lg bg-blue-50 text-blue-700 hover:bg-blue-600 hover:text-white border border-blue-200 hover:border-blue-600 transition cursor-pointer shadow-2xs"
                          >
                            <Eye className="w-3.5 h-3.5" />
                            <span>Inspect</span>
                          </button>
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

      {/* ========================================================================= */}
      {/* TAB 2: NOTIFICATIONS CENTER */}
      {/* ========================================================================= */}
      {activeTab === 'notifications' && (
        <div className="space-y-6">
          {/* Notifications KPI Cards */}
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
            <KpiCard
              title="Total Notifications"
              value={notificationsOverview?.total_count ?? 0}
              unit="alerts"
              subtitle="All system and event alerts"
              icon={Bell}
              variant="primary"
            />
            <KpiCard
              title="Unread Attention Items"
              value={notificationsOverview?.unread_count ?? 0}
              unit="unread"
              subtitle="Awaiting coordinator review"
              icon={AlertCircle}
              variant={notificationsOverview?.unread_count > 0 ? 'warning' : 'success'}
            />
            <KpiCard
              title="Critical Severity"
              value={notificationsOverview?.critical_count ?? 0}
              unit="critical"
              subtitle="Requires immediate response"
              icon={AlertTriangle}
              variant={notificationsOverview?.critical_count > 0 ? 'danger' : 'success'}
            />
            <KpiCard
              title="Action Required"
              value={notificationsOverview?.action_required_count ?? 0}
              unit="actionable"
              subtitle="Pending human decision or approval"
              icon={CheckCircle2}
              variant="default"
            />
          </div>

          {/* Filter Bar & Batch Actions */}
          <div className="bg-white p-4 rounded-xl border border-slate-200 shadow-xs flex flex-wrap items-center justify-between gap-4">
            <div className="flex flex-wrap items-center gap-3">
              {/* Read Filter */}
              <div className="flex items-center gap-1 bg-slate-100 p-1 rounded-lg">
                {[
                  { id: 'ALL', label: 'All Alerts' },
                  { id: 'UNREAD', label: 'Unread Only' },
                  { id: 'READ', label: 'Read Only' },
                ].map((f) => (
                  <button
                    key={f.id}
                    onClick={() => setNotifReadFilter(f.id)}
                    className={`px-3 py-1 text-xs font-semibold rounded-md transition cursor-pointer ${
                      notifReadFilter === f.id
                        ? 'bg-white text-slate-900 shadow-2xs font-bold'
                        : 'text-slate-500 hover:text-slate-800'
                    }`}
                  >
                    {f.label}
                  </button>
                ))}
              </div>

              {/* Severity Select */}
              <select
                value={notifSeverityFilter}
                onChange={(e) => setNotifSeverityFilter(e.target.value)}
                className="text-xs bg-slate-50 border border-slate-200 rounded-lg px-3 py-1.5 text-slate-700 font-medium focus:ring-1 focus:ring-blue-500"
              >
                <option value="ALL">All Severities</option>
                <option value="CRITICAL">Critical</option>
                <option value="HIGH">High</option>
                <option value="MEDIUM">Medium</option>
                <option value="INFORMATIONAL">Informational</option>
              </select>
            </div>

            <button
              onClick={handleMarkAllRead}
              className="px-3 py-1.5 text-xs font-semibold rounded-lg border border-slate-300 text-slate-700 hover:bg-slate-50 transition flex items-center gap-1.5 cursor-pointer shadow-2xs"
            >
              <Check className="w-3.5 h-3.5 text-slate-500" />
              <span>Mark All as Read</span>
            </button>
          </div>

          {/* Notifications List */}
          <div className="space-y-3">
            {notificationsLoading ? (
              <div className="p-12 text-center text-xs text-slate-400 bg-white rounded-xl border border-slate-200">
                Loading live notification queue...
              </div>
            ) : !notificationsOverview?.items || notificationsOverview.items.length === 0 ? (
              <EmptyState
                icon={Bell}
                title="No notifications to display"
                description={
                  selectedOrgId !== 'ALL'
                    ? `No notification items found for ${currentOrg?.name || 'this customer'}.`
                    : 'All notifications have been cleared or acknowledged.'
                }
              />
            ) : (
              notificationsOverview.items.map((notif) => (
                <div
                  key={notif.id}
                  className={`p-4 rounded-xl border transition flex flex-col sm:flex-row sm:items-center justify-between gap-4 ${
                    !notif.is_read
                      ? 'bg-white border-blue-200 shadow-xs'
                      : 'bg-slate-50/70 border-slate-200 text-slate-600'
                  }`}
                >
                  <div className="flex items-start gap-3.5">
                    <div className={`w-8 h-8 rounded-lg flex items-center justify-center shrink-0 mt-0.5 ${
                      notif.severity === 'CRITICAL'
                        ? 'bg-rose-50 text-rose-600'
                        : notif.severity === 'HIGH'
                        ? 'bg-amber-50 text-amber-600'
                        : 'bg-blue-50 text-blue-600'
                    }`}>
                      <Bell className="w-4 h-4" />
                    </div>

                    <div className="space-y-1">
                      <div className="flex flex-wrap items-center gap-2">
                        <span className={`font-bold text-sm ${!notif.is_read ? 'text-slate-900' : 'text-slate-700'}`}>
                          {notif.title}
                        </span>
                        {!notif.is_read && (
                          <span className="w-2 h-2 rounded-full bg-blue-600 shrink-0" title="Unread notification"></span>
                        )}
                        <span className={`px-2 py-0.2 rounded-full text-[10px] font-bold border ${getPriorityBadgeClass(notif.severity)}`}>
                          {notif.severity}
                        </span>
                        <span className={`px-2 py-0.2 rounded-full text-[10px] font-mono font-semibold border ${getDeliveryStatusBadgeClass(notif.deliveryStatus)}`}>
                          {notif.deliveryStatus}
                        </span>
                        <span className="px-2 py-0.2 rounded text-[10px] font-mono bg-slate-100 text-slate-600">
                          {notif.source_module}
                        </span>
                      </div>

                      <p className="text-xs text-slate-600 leading-relaxed max-w-2xl">
                        {notif.message}
                      </p>

                      <div className="flex flex-wrap items-center gap-3 text-[11px] text-slate-400 pt-0.5">
                        <span className="flex items-center gap-1 text-slate-500 font-medium">
                          <Building2 className="w-3 h-3 text-slate-400" />
                          <span>{notif.org_name}</span>
                        </span>
                        <span>•</span>
                        <span>{formatDate(notif.created_at)}</span>
                        {notif.ai_summary && (
                          <>
                            <span>•</span>
                            <span className="text-purple-600 font-medium truncate max-w-md">
                              AI Insight: {notif.ai_summary}
                            </span>
                          </>
                        )}
                      </div>
                    </div>
                  </div>

                  {/* Actions */}
                  <div className="flex items-center gap-2 shrink-0 self-end sm:self-center">
                    {!notif.is_read && (
                      <button
                        onClick={() => handleMarkNotificationRead(notif.id)}
                        className="px-2.5 py-1 text-xs font-semibold rounded bg-slate-100 text-slate-700 hover:bg-blue-50 hover:text-blue-700 transition cursor-pointer"
                      >
                        Mark Read
                      </button>
                    )}
                    {!notif.is_acknowledged && (
                      <button
                        onClick={() => handleAcknowledgeNotification(notif.id)}
                        className="px-2.5 py-1 text-xs font-semibold rounded bg-blue-50 text-blue-700 hover:bg-blue-100 transition cursor-pointer"
                      >
                        Acknowledge
                      </button>
                    )}
                    {notif.action_url && (
                      <a
                        href={notif.action_url}
                        className="p-1.5 text-slate-400 hover:text-slate-700 rounded hover:bg-slate-100 transition"
                        title="Navigate to related entity"
                      >
                        <ExternalLink className="w-4 h-4" />
                      </a>
                    )}
                  </div>
                </div>
              ))
            )}
          </div>
        </div>
      )}

      {/* ========================================================================= */}
      {/* TAB 3: UNIFIED ACTIVITY TIMELINE */}
      {/* ========================================================================= */}
      {activeTab === 'activity' && (
        <div className="space-y-6">
          {/* Controls */}
          <div className="bg-white p-4 rounded-xl border border-slate-200 shadow-xs flex flex-wrap items-center justify-between gap-4">
            <div className="flex items-center gap-3">
              <span className="text-xs font-bold text-slate-700">Filter Event Domain:</span>
              <div className="flex flex-wrap items-center gap-1 bg-slate-100 p-1 rounded-lg">
                {['ALL', 'SHIPMENTS', 'DOCUMENTS', 'LEADS', 'BILLING', 'SUPPORT', 'AUTHENTICATION', 'EMAIL'].map((cat) => (
                  <button
                    key={cat}
                    onClick={() => setActivityCategoryFilter(cat)}
                    className={`px-3 py-1 text-xs font-semibold rounded-md transition cursor-pointer ${
                      activityCategoryFilter === cat
                        ? 'bg-white text-slate-900 shadow-2xs font-bold'
                        : 'text-slate-500 hover:text-slate-800'
                    }`}
                  >
                    {cat}
                  </button>
                ))}
              </div>
            </div>

            <div className="text-xs text-slate-500">
              Showing chronological multi-tenant event stream
            </div>
          </div>

          {/* Activity Feed */}
          <div className="bg-white rounded-xl border border-slate-200 shadow-xs p-6">
            {activityLoading ? (
              <div className="p-12 text-center text-xs text-slate-400">Compiling unified multi-source timeline...</div>
            ) : activityItems.length === 0 ? (
              <EmptyState
                icon={Activity}
                title="No activity recorded"
                description="No recent operational activities or audit events found matching the criteria."
              />
            ) : (
              <div className="relative pl-6 space-y-6 before:absolute before:left-2 before:top-2 before:bottom-2 before:w-0.5 before:bg-slate-200">
                {activityItems.map((act) => (
                  <div key={act.id} className="relative group">
                    {/* Node Dot */}
                    <div className="absolute -left-6 top-1.5 w-3.5 h-3.5 rounded-full border-2 border-white shadow-xs flex items-center justify-center bg-blue-600"></div>

                    <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-2">
                      <div className="space-y-1">
                        <div className="flex flex-wrap items-center gap-2">
                          <span className="font-bold text-sm text-slate-900">{act.action}</span>
                          <span className="px-2 py-0.2 rounded-full text-[10px] font-mono font-bold bg-slate-100 text-slate-700">
                            {act.category}
                          </span>
                          <span className="px-2 py-0.2 rounded text-[10px] font-mono bg-blue-50 text-blue-700 border border-blue-200">
                            {act.entity}
                          </span>
                          <span className={`px-1.5 py-0.2 rounded text-[10px] font-bold ${
                            act.result === 'SUCCESS' || act.result === 'DELIVERED'
                              ? 'bg-emerald-50 text-emerald-700'
                              : 'bg-rose-50 text-rose-700'
                          }`}>
                            {act.result}
                          </span>
                        </div>

                        <p className="text-xs text-slate-600 max-w-2xl">{act.description}</p>

                        <div className="flex flex-wrap items-center gap-3 text-[11px] text-slate-400 pt-0.5">
                          <span className="font-medium text-slate-700 flex items-center gap-1">
                            <User className="w-3 h-3 text-slate-400" />
                            <span>{act.actor_name}</span>
                          </span>
                          <span>•</span>
                          <span className="flex items-center gap-1 text-slate-600">
                            <Building2 className="w-3 h-3 text-slate-400" />
                            <span>{act.org_name}</span>
                          </span>
                          <span>•</span>
                          <span className="font-mono text-slate-400">Corr: {act.correlation_id}</span>
                        </div>
                      </div>

                      <span className="text-xs text-slate-400 font-mono shrink-0 whitespace-nowrap self-start sm:self-center">
                        {formatDate(act.timestamp)}
                      </span>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>
        </div>
      )}

      {/* ========================================================================= */}
      {/* TAB 4: FORENSIC AUDIT LEDGER */}
      {/* ========================================================================= */}
      {activeTab === 'audit' && (
        <div className="space-y-6">
          {/* Controls */}
          <div className="bg-white p-4 rounded-xl border border-slate-200 shadow-xs flex flex-wrap items-center justify-between gap-4">
            <div className="flex flex-wrap items-center gap-3">
              {/* Module Filter */}
              <select
                value={auditModuleFilter}
                onChange={(e) => setAuditModuleFilter(e.target.value)}
                className="text-xs bg-slate-50 border border-slate-200 rounded-lg px-3 py-1.5 text-slate-700 font-medium focus:ring-1 focus:ring-blue-500"
              >
                <option value="ALL">All Modules</option>
                <option value="AUTHENTICATION">Authentication</option>
                <option value="SUPPORT">Support</option>
                <option value="SHIPMENTS">Shipments</option>
                <option value="DOCUMENTS">Documents</option>
                <option value="LEADS">Leads</option>
                <option value="PRICING">Pricing</option>
                <option value="CONTRACTS">Contracts</option>
                <option value="SETTINGS">Settings</option>
                <option value="AI_GOVERNANCE">AI Governance</option>
              </select>

              {/* Result Filter */}
              <select
                value={auditResultFilter}
                onChange={(e) => setAuditResultFilter(e.target.value)}
                className="text-xs bg-slate-50 border border-slate-200 rounded-lg px-3 py-1.5 text-slate-700 font-medium focus:ring-1 focus:ring-blue-500"
              >
                <option value="ALL">All Results</option>
                <option value="SUCCESS">Success</option>
                <option value="FAILURE">Failure</option>
                <option value="DENIED">Denied</option>
              </select>
            </div>

            {/* Search Input */}
            <div className="relative w-full sm:w-64">
              <Search className="w-3.5 h-3.5 text-slate-400 absolute left-3 top-1/2 -translate-y-1/2" />
              <input
                type="text"
                placeholder="Search action, actor, description..."
                value={auditSearch}
                onChange={(e) => setAuditSearch(e.target.value)}
                onKeyDown={(e) => e.key === 'Enter' && loadAuditLogs()}
                className="w-full pl-8 pr-3 py-1.5 text-xs bg-slate-50 border border-slate-200 rounded-lg focus:outline-none focus:ring-1 focus:ring-blue-500"
              />
            </div>
          </div>

          {/* Forensic Audit Table */}
          <div className="bg-white rounded-xl border border-slate-200 shadow-xs overflow-hidden">
            {auditLoading ? (
              <div className="p-12 text-center text-xs text-slate-400">Loading forensic audit records...</div>
            ) : auditLogs.length === 0 ? (
              <EmptyState
                icon={ShieldCheck}
                title="No audit events found"
                description="No immutable audit entries match the current filter criteria."
              />
            ) : (
              <div className="overflow-x-auto">
                <table className="w-full text-left text-xs min-w-[900px]">
                  <thead>
                    <tr className="border-b border-slate-200 bg-slate-50/90 text-slate-500 font-semibold uppercase tracking-wider text-[11px]">
                      <th className="py-3 px-3">Audit ID</th>
                      <th className="py-3 px-3">Timestamp</th>
                      <th className="py-3 px-3">Actor</th>
                      <th className="py-3 px-3">Module</th>
                      <th className="py-3 px-3">Action</th>
                      <th className="py-3 px-3">Resource Target</th>
                      <th className="py-3 px-3">Description</th>
                      <th className="py-3 px-3">Result</th>
                      <th className="py-3 px-3 text-right sticky right-0 bg-slate-50 shadow-[-4px_0_6px_-2px_rgba(0,0,0,0.06)] z-10">View</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-slate-100">
                    {auditLogs.map((log) => (
                      <tr
                        key={log.id}
                        onClick={() => setSelectedAuditLog(log)}
                        className="hover:bg-blue-50/50 transition cursor-pointer group"
                        title="Click to view raw forensic audit record"
                      >
                        <td className="py-3 px-3 font-mono font-semibold text-slate-500 whitespace-nowrap">
                          #{log.id}
                        </td>
                        <td className="py-3 px-3 text-slate-500 whitespace-nowrap font-mono text-[11px]">
                          {formatDate(log.created_at)}
                        </td>
                        <td className="py-3 px-3 font-semibold text-slate-900">
                          {log.actor_name}
                        </td>
                        <td className="py-3 px-3 whitespace-nowrap">
                          <span className="px-2 py-0.5 rounded text-[10px] font-mono font-bold bg-slate-100 text-slate-700">
                            {log.module}
                          </span>
                        </td>
                        <td className="py-3 px-3 font-semibold text-slate-800 font-mono text-[11px] whitespace-nowrap">
                          {log.action}
                        </td>
                        <td className="py-3 px-3 text-slate-600 font-mono text-[11px] whitespace-nowrap">
                          {log.resource_type} {log.resource_id ? `#${log.resource_id}` : ''}
                        </td>
                        <td className="py-3 px-3 text-slate-600 max-w-xs truncate" title={log.description}>
                          {log.description}
                        </td>
                        <td className="py-3 px-3 whitespace-nowrap">
                          <span className={`px-2 py-0.5 rounded text-[10px] font-bold ${
                            log.result === 'SUCCESS'
                              ? 'bg-emerald-50 text-emerald-700 border border-emerald-200'
                              : 'bg-rose-50 text-rose-700 border border-rose-200'
                          }`}>
                            {log.result}
                          </span>
                        </td>
                        <td className="py-3 px-3 text-right whitespace-nowrap sticky right-0 bg-white group-hover:bg-blue-50/70 transition shadow-[-4px_0_6px_-2px_rgba(0,0,0,0.06)] z-10">
                          <button
                            onClick={(e) => {
                              e.stopPropagation();
                              setSelectedAuditLog(log);
                            }}
                            className="p-1.5 text-slate-500 hover:text-blue-700 rounded-lg hover:bg-blue-50 transition cursor-pointer"
                            title="Inspect audit payload"
                          >
                            <Terminal className="w-3.5 h-3.5" />
                          </button>
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

      {/* ========================================================================= */}
      {/* SUPPORT CASE DETAIL & MANAGEMENT MODAL */}
      {/* ========================================================================= */}
      {caseDetailModalOpen && (
        <div className="fixed inset-0 z-50 overflow-y-auto bg-slate-900/60 backdrop-blur-xs flex items-center justify-center p-4">
          <div className="relative bg-white rounded-2xl shadow-2xl border border-slate-200 w-full max-w-4xl max-h-[90vh] flex flex-col overflow-hidden animate-fade-in">
            {/* Modal Header */}
            <div className="px-6 py-4 border-b border-slate-200 flex items-center justify-between bg-slate-50/50">
              <div className="flex items-center gap-3">
                <div className="w-10 h-10 rounded-xl bg-blue-600 text-white flex items-center justify-center shadow-xs">
                  <Headphones className="w-5 h-5" />
                </div>
                <div>
                  <div className="flex items-center gap-2">
                    <span className="font-mono font-bold text-blue-600">{selectedCase?.case_code || 'Support Case'}</span>
                    <span className={`px-2 py-0.2 text-[10px] font-bold rounded-full border ${getPriorityBadgeClass(selectedCase?.priority)}`}>
                      {selectedCase?.priority} Priority
                    </span>
                    <span className={`px-2 py-0.2 text-[10px] font-bold rounded-full border ${getStatusBadgeClass(selectedCase?.status)}`}>
                      {selectedCase?.status?.replace('_', ' ')}
                    </span>
                  </div>
                  <h2 className="text-base font-bold text-slate-900 mt-0.5">{selectedCase?.subject}</h2>
                </div>
              </div>

              <button
                onClick={() => setCaseDetailModalOpen(false)}
                className="p-2 text-slate-400 hover:text-slate-700 rounded-lg hover:bg-slate-100 transition cursor-pointer"
              >
                <X className="w-5 h-5" />
              </button>
            </div>

            {/* Modal Body */}
            <div className="flex-1 overflow-y-auto p-6 space-y-6">
              {caseDetailLoading ? (
                <div className="p-12 text-center text-xs text-slate-400">Loading case intelligence and internal notes...</div>
              ) : selectedCase ? (
                <>
                  {actionSuccessMsg && (
                    <div className="p-3 bg-emerald-50 border border-emerald-200 text-emerald-800 rounded-xl text-xs font-semibold flex items-center gap-2">
                      <CheckCircle2 className="w-4 h-4 text-emerald-600" />
                      <span>{actionSuccessMsg}</span>
                    </div>
                  )}

                  {/* Top Metadata Grid */}
                  <div className="grid grid-cols-2 sm:grid-cols-4 gap-4 p-4 rounded-xl bg-slate-50 border border-slate-200 text-xs">
                    <div>
                      <span className="text-slate-400 uppercase text-[10px] font-bold">Customer Organization</span>
                      <div className="font-bold text-slate-900 mt-0.5 flex items-center gap-1">
                        <Link to={`/organizations/${selectedCase.org_id}`} className="hover:text-blue-600 hover:underline">
                          {selectedCase.org_name}
                        </Link>
                      </div>
                    </div>

                    <div>
                      <span className="text-slate-400 uppercase text-[10px] font-bold">Linked Shipment Entity</span>
                      <div className="font-mono font-semibold text-slate-900 mt-0.5">
                        {selectedCase.shipment_ref}
                      </div>
                    </div>

                    <div>
                      <span className="text-slate-400 uppercase text-[10px] font-bold">Service Level Agreement</span>
                      <div className="font-bold text-slate-900 mt-0.5">
                        <span className={`px-1.5 py-0.2 rounded text-[10px] ${
                          selectedCase.sla_state === 'MET'
                            ? 'bg-emerald-100 text-emerald-800'
                            : selectedCase.sla_state === 'BREACHED'
                            ? 'bg-rose-100 text-rose-800 font-black'
                            : 'bg-amber-100 text-amber-800'
                        }`}>
                          {selectedCase.sla_state}
                        </span>
                      </div>
                    </div>

                    <div>
                      <span className="text-slate-400 uppercase text-[10px] font-bold">Assigned Owner</span>
                      <div className="font-semibold text-slate-900 mt-0.5">
                        {selectedCase.assigned_owner || 'Internal Support Queue'}
                      </div>
                    </div>
                  </div>

                  {/* Description & AI Diagnostic */}
                  <div className="space-y-2">
                    <h3 className="text-xs font-bold uppercase tracking-wider text-slate-500">Case Description</h3>
                    <div className="p-3.5 bg-slate-50/70 border border-slate-200 rounded-xl text-xs text-slate-800 leading-relaxed">
                      {selectedCase.description || 'No detailed narrative provided for this operational event.'}
                    </div>

                    {selectedCase.ai_summary && (
                      <div className="p-3 bg-purple-50 border border-purple-200 rounded-xl text-xs text-purple-900 leading-relaxed flex items-start gap-2">
                        <span className="font-bold shrink-0">AI Diagnostic:</span>
                        <span>{selectedCase.ai_summary}</span>
                      </div>
                    )}
                  </div>

                  {/* Lifecycle Status & Resolution Update Form */}
                  <form onSubmit={handleUpdateStatus} className="p-4 rounded-xl border border-slate-200 bg-white space-y-3">
                    <h3 className="text-xs font-bold uppercase tracking-wider text-slate-700 flex items-center gap-1.5">
                      <CheckCircle2 className="w-4 h-4 text-blue-600" />
                      <span>Update Case Lifecycle & Resolution</span>
                    </h3>

                    <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
                      <div>
                        <label htmlFor="modal-status-select" className="block text-[11px] font-semibold text-slate-600 mb-1">Target Status</label>
                        <select
                          id="modal-status-select"
                          value={newStatus}
                          onChange={(e) => setNewStatus(e.target.value)}
                          className="w-full text-xs bg-slate-50 border border-slate-300 rounded-lg p-2 focus:ring-1 focus:ring-blue-500"
                        >
                          <option value="OPEN">OPEN (Under review)</option>
                          <option value="ACKNOWLEDGED">ACKNOWLEDGED (Coordinator assigned)</option>
                          <option value="IN_PROGRESS">IN_PROGRESS (Investigation in progress)</option>
                          <option value="RESOLVED">RESOLVED (Resolution applied)</option>
                          <option value="DISMISSED">DISMISSED (Non-actionable)</option>
                        </select>
                      </div>

                      <div>
                        <label htmlFor="modal-resolution-input" className="block text-[11px] font-semibold text-slate-600 mb-1">Resolution Summary</label>
                        <input
                          id="modal-resolution-input"
                          type="text"
                          placeholder="e.g. Customs document uploaded and terminal fee cleared"
                          value={resolutionNotes}
                          onChange={(e) => setResolutionNotes(e.target.value)}
                          className="w-full text-xs bg-slate-50 border border-slate-300 rounded-lg p-2 focus:ring-1 focus:ring-blue-500"
                        />
                      </div>
                    </div>

                    <button
                      type="submit"
                      disabled={updatingCaseStatus}
                      className="px-4 py-2 bg-blue-600 text-white rounded-lg text-xs font-bold hover:bg-blue-700 transition disabled:opacity-50 cursor-pointer shadow-xs"
                    >
                      {updatingCaseStatus ? 'Saving changes...' : 'Save Lifecycle Update'}
                    </button>
                  </form>

                  {/* Internal LogisticsHQ Notes (Protected Domain) */}
                  <div className="space-y-3 pt-2">
                    <div className="flex items-center justify-between">
                      <div className="flex items-center gap-2">
                        <Lock className="w-4 h-4 text-amber-600" />
                        <h3 className="text-xs font-bold uppercase tracking-wider text-slate-900">
                          Internal LogisticsHQ Notes
                        </h3>
                        <span className="px-2 py-0.2 rounded text-[10px] font-bold bg-amber-50 text-amber-800 border border-amber-200">
                          Internal Staff Only — Never Exposed to Customer Portal
                        </span>
                      </div>
                    </div>

                    {/* Notes Journal */}
                    <div className="space-y-2 max-h-48 overflow-y-auto">
                      {!selectedCase.internal_notes || selectedCase.internal_notes.length === 0 ? (
                        <div className="p-4 bg-slate-50 border border-slate-200 rounded-xl text-xs text-slate-500 text-center">
                          No internal notes logged for this customer case yet.
                        </div>
                      ) : (
                        selectedCase.internal_notes.map((note) => (
                          <div key={note.id} className="p-3 bg-slate-50/90 border border-slate-200 rounded-xl text-xs space-y-1">
                            <div className="flex items-center justify-between text-[11px] text-slate-400">
                              <span className="font-bold text-slate-700">{note.author_name}</span>
                              <span>{formatDate(note.created_at)}</span>
                            </div>
                            <p className="text-slate-800 whitespace-pre-wrap">{note.content}</p>
                          </div>
                        ))
                      )}
                    </div>

                    {/* Add Note Form */}
                    <form onSubmit={handleAddNote} className="flex gap-2">
                      <input
                        type="text"
                        placeholder="Add internal escalation note or triage comment..."
                        value={newNoteContent}
                        onChange={(e) => setNewNoteContent(e.target.value)}
                        className="flex-1 text-xs bg-slate-50 border border-slate-300 rounded-lg p-2 focus:ring-1 focus:ring-blue-500"
                      />
                      <button
                        type="submit"
                        disabled={submittingNote || !newNoteContent.trim()}
                        className="px-4 py-2 bg-slate-800 text-white rounded-lg text-xs font-bold hover:bg-slate-900 transition disabled:opacity-50 cursor-pointer flex items-center gap-1"
                      >
                        <Send className="w-3.5 h-3.5" />
                        <span>Add Note</span>
                      </button>
                    </form>
                  </div>
                </>
              ) : null}
            </div>

            {/* Modal Footer */}
            <div className="px-6 py-3 border-t border-slate-200 bg-slate-50 flex items-center justify-between">
              <span className="text-[11px] text-slate-400">
                Audit trail automatically recorded for all case mutations.
              </span>
              <button
                onClick={() => setCaseDetailModalOpen(false)}
                className="px-4 py-1.5 text-xs font-semibold rounded-lg bg-slate-200 text-slate-700 hover:bg-slate-300 transition cursor-pointer"
              >
                Close
              </button>
            </div>
          </div>
        </div>
      )}

      {/* ========================================================================= */}
      {/* AUDIT LOG RAW JSON MODAL */}
      {/* ========================================================================= */}
      {selectedAuditLog && (
        <div className="fixed inset-0 z-50 overflow-y-auto bg-slate-900/60 backdrop-blur-xs flex items-center justify-center p-4">
          <div className="relative bg-white rounded-2xl shadow-2xl border border-slate-200 w-full max-w-2xl max-h-[85vh] flex flex-col overflow-hidden animate-fade-in">
            <div className="px-6 py-4 border-b border-slate-200 flex items-center justify-between bg-slate-50/50">
              <div className="flex items-center gap-2">
                <ShieldCheck className="w-5 h-5 text-emerald-600" />
                <h3 className="font-bold text-sm text-slate-900">
                  Forensic Audit Record #{selectedAuditLog.id}
                </h3>
              </div>
              <button
                onClick={() => setSelectedAuditLog(null)}
                className="p-1.5 text-slate-400 hover:text-slate-700 rounded-lg transition cursor-pointer"
              >
                <X className="w-4 h-4" />
              </button>
            </div>

            <div className="p-6 overflow-y-auto space-y-4">
              <div className="grid grid-cols-2 gap-3 text-xs bg-slate-50 p-3.5 rounded-xl border border-slate-200">
                <div>
                  <span className="text-[10px] uppercase font-bold text-slate-400">Action</span>
                  <div className="font-mono font-bold text-slate-900">{selectedAuditLog.action}</div>
                </div>
                <div>
                  <span className="text-[10px] uppercase font-bold text-slate-400">Module</span>
                  <div className="font-mono font-bold text-slate-900">{selectedAuditLog.module}</div>
                </div>
                <div>
                  <span className="text-[10px] uppercase font-bold text-slate-400">Actor</span>
                  <div className="font-bold text-slate-900">{selectedAuditLog.actor_name}</div>
                </div>
                <div>
                  <span className="text-[10px] uppercase font-bold text-slate-400">Result</span>
                  <div className="font-bold text-emerald-700">{selectedAuditLog.result}</div>
                </div>
              </div>

              <div>
                <span className="text-[10px] uppercase font-bold text-slate-400 block mb-1">Description</span>
                <p className="text-xs text-slate-800 p-3 bg-slate-50 rounded-xl border border-slate-200">
                  {selectedAuditLog.description}
                </p>
              </div>

              <div>
                <span className="text-[10px] uppercase font-bold text-slate-400 block mb-1">Raw Audit Representation</span>
                <pre className="p-3.5 bg-slate-900 text-slate-200 rounded-xl text-[11px] font-mono overflow-x-auto">
                  {JSON.stringify(selectedAuditLog, null, 2)}
                </pre>
              </div>
            </div>

            <div className="px-6 py-3 border-t border-slate-200 bg-slate-50 flex justify-end">
              <button
                onClick={() => setSelectedAuditLog(null)}
                className="px-4 py-1.5 text-xs font-semibold rounded-lg bg-slate-200 text-slate-700 hover:bg-slate-300 transition cursor-pointer"
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

import React, { useState, useEffect, useRef, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  Shield,
  Download,
  Calendar,
  User,
  Layers,
  Zap,
  CheckCircle2,
  XCircle,
  Search,
  X,
  ChevronDown,
  ChevronLeft,
  ChevronRight,
  ExternalLink,
  Clock,
  MapPin,
  Laptop,
  ArrowRight,
  Building2,
  FileText,
  Truck,
  Mail,
  Receipt,
  CreditCard,
  Package,
  FileSpreadsheet,
  BarChart2,
  CheckSquare,
  Folder,
  Send,
  Users,
  ShieldCheck,
  RefreshCw,
  AlertCircle
} from 'lucide-react';
import { auditService } from '../../../services/auditService';
import api from '../../../services/api';
import toast from 'react-hot-toast';
import './AuditLogsPage.css';

// ── Module Metadata & Icon Resolver ──────────────────────────────────────────
const MODULE_CONFIG = {
  AUTHENTICATION: { label: 'Authentication', icon: Shield, color: '#3B82F6' },
  USERS: { label: 'Users & Team', icon: Users, color: '#6366F1' },
  ROLES_PERMISSIONS: { label: 'Roles & Permissions', icon: ShieldCheck, color: '#8B5CF6' },
  LEADS: { label: 'Leads', icon: Building2, color: '#0EA5E9' },
  RFQS: { label: 'RFQs', icon: FileSpreadsheet, color: '#10B981' },
  QUOTATIONS: { label: 'Quotations', icon: FileText, color: '#F59E0B' },
  BOOKINGS: { label: 'Bookings', icon: Package, color: '#6366F1' },
  SHIPMENTS: { label: 'Shipments', icon: Truck, color: '#2563EB' },
  TRACKING: { label: 'Tracking', icon: MapPin, color: '#0D9488' },
  RATE_MANAGEMENT: { label: 'Rate Management', icon: BarChart2, color: '#EC4899' },
  CONTRACTS: { label: 'Contracts', icon: FileText, color: '#8B5CF6' },
  CUSTOMERS: { label: 'Customers', icon: Building2, color: '#3B82F6' },
  OUTREACH: { label: 'Outreach', icon: Send, color: '#F97316' },
  DOCUMENTS: { label: 'Documents', icon: Folder, color: '#64748B' },
  APPROVALS: { label: 'Approvals', icon: CheckSquare, color: '#10B981' },
  INVOICES: { label: 'Invoices', icon: Receipt, color: '#10B981' },
  PAYMENTS: { label: 'Payments', icon: CreditCard, color: '#059669' },
  CARRIER_INTEGRATIONS: { label: 'Carrier Integration', icon: Truck, color: '#2563EB' },
  SETTINGS: { label: 'Settings', icon: Layers, color: '#64748B' },
  GENERAL: { label: 'General', icon: Layers, color: '#64748B' }
};

const ACTION_CONFIG = {
  CREATE: { label: 'CREATED', bg: '#EFF6FF', color: '#1D4ED8', border: '#BFDBFE' },
  CREATED: { label: 'CREATED', bg: '#EFF6FF', color: '#1D4ED8', border: '#BFDBFE' },
  UPDATE: { label: 'UPDATED', bg: '#FFFBEB', color: '#B45309', border: '#FDE68A' },
  UPDATED: { label: 'UPDATED', bg: '#FFFBEB', color: '#B45309', border: '#FDE68A' },
  DELETE: { label: 'DELETED', bg: '#FFF1F2', color: '#BE123C', border: '#FECDD3' },
  DELETED: { label: 'DELETED', bg: '#FFF1F2', color: '#BE123C', border: '#FECDD3' },
  LOGIN: { label: 'LOGIN', bg: '#EFF6FF', color: '#2563EB', border: '#BFDBFE' },
  LOGIN_FAILED: { label: 'LOGIN FAILED', bg: '#FEF2F2', color: '#DC2626', border: '#FECACA' },
  LOGOUT: { label: 'LOGOUT', bg: '#F1F5F9', color: '#475569', border: '#E2E8F0' },
  INVITE: { label: 'INVITED', bg: '#EEF2FF', color: '#4F46E5', border: '#C7D2FE' },
  INVITED: { label: 'INVITED', bg: '#EEF2FF', color: '#4F46E5', border: '#C7D2FE' },
  ROLE_CHANGED: { label: 'ROLE CHANGED', bg: '#F5F3FF', color: '#6D28D9', border: '#DDD6FE' },
  PERMISSION_CHANGED: { label: 'PERM CHANGED', bg: '#F5F3FF', color: '#6D28D9', border: '#DDD6FE' },
  APPROVE: { label: 'APPROVED', bg: '#ECFDF5', color: '#047857', border: '#A7F3D0' },
  APPROVED: { label: 'APPROVED', bg: '#ECFDF5', color: '#047857', border: '#A7F3D0' },
  REJECT: { label: 'REJECTED', bg: '#FFF1F2', color: '#BE123C', border: '#FECDD3' },
  REJECTED: { label: 'REJECTED', bg: '#FFF1F2', color: '#BE123C', border: '#FECDD3' },
  SEND: { label: 'SENT', bg: '#F0F9FF', color: '#0369A1', border: '#BAE6FD' },
  SENT: { label: 'SENT', bg: '#F0F9FF', color: '#0369A1', border: '#BAE6FD' },
  FOLLOW_UP_SENT: { label: 'FOLLOW-UP SENT', bg: '#F0F9FF', color: '#0369A1', border: '#BAE6FD' },
  CONNECT: { label: 'CONNECTED', bg: '#ECFDF5', color: '#047857', border: '#A7F3D0' },
  CONNECTED: { label: 'CONNECTED', bg: '#ECFDF5', color: '#047857', border: '#A7F3D0' },
  CARRIER_CONNECTED: { label: 'CONNECTED', bg: '#ECFDF5', color: '#047857', border: '#A7F3D0' },
  DISCONNECT: { label: 'DISCONNECTED', bg: '#F1F5F9', color: '#475569', border: '#E2E8F0' },
  DISCONNECTED: { label: 'DISCONNECTED', bg: '#F1F5F9', color: '#475569', border: '#E2E8F0' },
  ENABLE: { label: 'ENABLED', bg: '#ECFDF5', color: '#047857', border: '#A7F3D0' },
  ENABLED: { label: 'ENABLED', bg: '#ECFDF5', color: '#047857', border: '#A7F3D0' },
  DISABLE: { label: 'DISABLED', bg: '#FFFBEB', color: '#B45309', border: '#FDE68A' },
  DISABLED: { label: 'DISABLED', bg: '#FFFBEB', color: '#B45309', border: '#FDE68A' },
  SYNC: { label: 'SYNC', bg: '#F5F3FF', color: '#7C3AED', border: '#DDD6FE' },
  SYNCED: { label: 'SYNCED', bg: '#F5F3FF', color: '#7C3AED', border: '#DDD6FE' },
  CARRIER_SYNC_COMPLETED: { label: 'SYNC', bg: '#F5F3FF', color: '#7C3AED', border: '#DDD6FE' },
  CARRIER_GET_TRACKING: { label: 'TRACKING SYNC', bg: '#F0FDFA', color: '#0F766E', border: '#99F6E4' },
  PAYMENT_RECORDED: { label: 'PAYMENT_RECORDED', bg: '#ECFDF5', color: '#047857', border: '#A7F3D0' },
  EXPORT: { label: 'EXPORTED', bg: '#F1F5F9', color: '#475569', border: '#E2E8F0' }
};

const AVATAR_COLORS = [
  { bg: '#2563EB', color: '#FFFFFF' }, // Blue
  { bg: '#7C3AED', color: '#FFFFFF' }, // Purple
  { bg: '#059669', color: '#FFFFFF' }, // Emerald
  { bg: '#D97706', color: '#FFFFFF' }, // Amber
  { bg: '#DC2626', color: '#FFFFFF' }, // Red
  { bg: '#0891B2', color: '#FFFFFF' }, // Cyan
  { bg: '#4F46E5', color: '#FFFFFF' }, // Indigo
];

function getAvatarStyle(name = '') {
  if (!name || name === 'System') {
    return { bg: '#64748B', color: '#FFFFFF', initial: 'S' };
  }
  const clean = name.trim();
  const charCode = clean.charCodeAt(0) || 0;
  const palette = AVATAR_COLORS[charCode % AVATAR_COLORS.length];
  return { ...palette, initial: clean.charAt(0).toUpperCase() };
}

function formatAuditTime(dateStr) {
  if (!dateStr) return { time: '—', date: '—' };
  try {
    const d = new Date(dateStr);
    if (isNaN(d.getTime())) return { time: '—', date: '—' };

    const time = d.toLocaleTimeString('en-US', { hour: 'numeric', minute: '2-digit', hour12: true });
    
    // Check if today / yesterday
    const now = new Date();
    const isToday = d.toDateString() === now.toDateString();
    const yesterday = new Date(now);
    yesterday.setDate(yesterday.getDate() - 1);
    const isYesterday = d.toDateString() === yesterday.toDateString();

    let date;
    if (isToday) {
      date = d.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' });
    } else if (isYesterday) {
      date = 'Yesterday';
    } else {
      date = d.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' });
    }

    return { time, date, full: `${date} at ${time}` };
  } catch {
    return { time: '—', date: '—', full: '—' };
  }
}

function normalizeActionName(action = '') {
  const norm = action.toUpperCase();
  if (ACTION_CONFIG[norm]) {
    return ACTION_CONFIG[norm].label;
  }
  return norm.replace(/_/g, ' ');
}

export default function AuditLogsPage() {
  const navigate = useNavigate();

  // Data states
  const [logs, setLogs] = useState([]);
  const [totalCount, setTotalCount] = useState(0);
  const [totalPages, setTotalPages] = useState(1);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState(null);

  // Selected event for detail panel
  const [selectedLog, setSelectedLog] = useState(null);

  // Organization users for User Filter
  const [usersList, setUsersList] = useState([]);

  // Filter states
  const [search, setSearch] = useState('');
  const [debouncedSearch, setDebouncedSearch] = useState('');
  const [datePreset, setDatePreset] = useState('ALL');
  const [customStartDate, setCustomStartDate] = useState('');
  const [customEndDate, setCustomEndDate] = useState('');
  const [selectedUserId, setSelectedUserId] = useState('ALL');
  const [selectedModule, setSelectedModule] = useState('ALL');
  const [selectedAction, setSelectedAction] = useState('ALL');
  const [selectedResult, setSelectedResult] = useState('ALL');

  // Pagination states
  const [page, setPage] = useState(1);
  const [limit, setLimit] = useState(10);

  // Dropdown open states
  const [openDropdown, setOpenDropdown] = useState(null); // 'date' | 'user' | 'module' | 'action' | 'result' | 'limit'
  const filterBarRef = useRef(null);

  // Close dropdowns on click outside
  useEffect(() => {
    function handleClickOutside(e) {
      if (filterBarRef.current && !filterBarRef.current.contains(e.target)) {
        setOpenDropdown(null);
      }
    }
    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  // Keyboard navigation for closing drawer
  useEffect(() => {
    function handleKeyDown(e) {
      if (e.key === 'Escape' && selectedLog) {
        setSelectedLog(null);
      }
    }
    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [selectedLog]);

  // Load organization users for dropdown
  useEffect(() => {
    async function loadOrgUsers() {
      try {
        const res = await api.get('/api/v1/users');
        if (Array.isArray(res)) {
          setUsersList(res);
        } else if (res?.data && Array.isArray(res.data)) {
          setUsersList(res.data);
        }
      } catch (e) {
        // Fallback gracefully
      }
    }
    loadOrgUsers();
  }, []);

  // Debounce search input
  useEffect(() => {
    const timer = setTimeout(() => {
      setDebouncedSearch(search);
      setPage(1);
    }, 300);
    return () => clearTimeout(timer);
  }, [search]);

  // Compute start/end dates from preset
  const getDateRangeParams = useCallback(() => {
    if (datePreset === 'TODAY') {
      const start = new Date();
      start.setHours(0, 0, 0, 0);
      const end = new Date();
      end.setHours(23, 59, 59, 999);
      return { start_date: start.toISOString(), end_date: end.toISOString() };
    }
    if (datePreset === 'YESTERDAY') {
      const start = new Date();
      start.setDate(start.getDate() - 1);
      start.setHours(0, 0, 0, 0);
      const end = new Date();
      end.setDate(end.getDate() - 1);
      end.setHours(23, 59, 59, 999);
      return { start_date: start.toISOString(), end_date: end.toISOString() };
    }
    if (datePreset === '7D') {
      const start = new Date();
      start.setDate(start.getDate() - 7);
      start.setHours(0, 0, 0, 0);
      return { start_date: start.toISOString() };
    }
    if (datePreset === '30D') {
      const start = new Date();
      start.setDate(start.getDate() - 30);
      start.setHours(0, 0, 0, 0);
      return { start_date: start.toISOString() };
    }
    if (datePreset === 'CUSTOM') {
      const params = {};
      if (customStartDate) {
        params.start_date = new Date(customStartDate).toISOString();
      }
      if (customEndDate) {
        const end = new Date(customEndDate);
        end.setHours(23, 59, 59, 999);
        params.end_date = end.toISOString();
      }
      return params;
    }
    return {};
  }, [datePreset, customStartDate, customEndDate]);

  // Fetch Audit Logs
  const fetchLogs = useCallback(async () => {
    setIsLoading(true);
    setError(null);
    try {
      const dateParams = getDateRangeParams();
      const params = {
        page,
        limit,
        search: debouncedSearch,
        module: selectedModule !== 'ALL' ? selectedModule : undefined,
        action: selectedAction !== 'ALL' ? selectedAction : undefined,
        result: selectedResult !== 'ALL' ? selectedResult : undefined,
        actor_id: selectedUserId !== 'ALL' && selectedUserId !== 'SYSTEM' ? Number(selectedUserId) : undefined,
        actor_type: selectedUserId === 'SYSTEM' ? 'SYSTEM' : undefined,
        ...dateParams
      };

      const res = await auditService.getAuditLogs(params);
      const items = res.items || res.data || [];
      const total = res.total !== undefined ? res.total : (res.total_count || items.length);
      const pages = res.total_pages || Math.ceil(total / limit) || 1;

      setLogs(items);
      setTotalCount(total);
      setTotalPages(pages);

      // If currently selected log is in this batch, update it; otherwise keep or deselect if not on page
      if (selectedLog) {
        const freshSelected = items.find(l => l.id === selectedLog.id);
        if (freshSelected) {
          setSelectedLog(freshSelected);
        }
      }
    } catch (err) {
      console.error('Failed to load audit logs:', err);
      setError(err.message || 'Failed to load audit logs. Please try again.');
    } finally {
      setIsLoading(false);
    }
  }, [page, limit, debouncedSearch, selectedModule, selectedAction, selectedResult, selectedUserId, getDateRangeParams]);

  useEffect(() => {
    fetchLogs();
  }, [fetchLogs]);

  // Clear all filters
  const handleClearAll = () => {
    setSearch('');
    setDebouncedSearch('');
    setDatePreset('ALL');
    setCustomStartDate('');
    setCustomEndDate('');
    setSelectedUserId('ALL');
    setSelectedModule('ALL');
    setSelectedAction('ALL');
    setSelectedResult('ALL');
    setPage(1);
    setOpenDropdown(null);
  };

  const hasActiveFilters = Boolean(
    search ||
    datePreset !== 'ALL' ||
    selectedUserId !== 'ALL' ||
    selectedModule !== 'ALL' ||
    selectedAction !== 'ALL' ||
    selectedResult !== 'ALL'
  );

  // Export CSV
  const handleExportCSV = async () => {
    try {
      toast.loading('Preparing export...', { id: 'csv-export' });
      // Fetch up to 1000 items matching current filters for export
      const dateParams = getDateRangeParams();
      const params = {
        page: 1,
        limit: 1000,
        search: debouncedSearch,
        module: selectedModule !== 'ALL' ? selectedModule : undefined,
        action: selectedAction !== 'ALL' ? selectedAction : undefined,
        result: selectedResult !== 'ALL' ? selectedResult : undefined,
        actor_id: selectedUserId !== 'ALL' && selectedUserId !== 'SYSTEM' ? Number(selectedUserId) : undefined,
        actor_type: selectedUserId === 'SYSTEM' ? 'SYSTEM' : undefined,
        ...dateParams
      };
      const res = await auditService.getAuditLogs(params);
      const items = res.items || res.data || logs;
      auditService.exportToCSV(items);
      toast.success(`Exported ${items.length} audit logs.`, { id: 'csv-export' });
    } catch (err) {
      toast.error('Failed to export audit logs.', { id: 'csv-export' });
    }
  };

  // Resolve Real Route for Resource Navigation
  const getResourceRoute = (log) => {
    if (!log || !log.resource_type) return null;
    const type = log.resource_type.toUpperCase();
    const id = log.resource_id;

    switch (type) {
      case 'LEAD':
        return { label: 'View Lead', path: '/dashboard/leads' };
      case 'RFQ':
        return { label: 'View RFQ', path: id ? `/dashboard/rfqs/${id}` : '/dashboard/rfqs' };
      case 'QUOTE':
      case 'QUOTATION':
        return { label: 'View Quotation', path: '/dashboard/quotations' };
      case 'BOOKING':
        return { label: 'View Booking', path: id ? `/dashboard/bookings/${id}` : '/dashboard/bookings' };
      case 'SHIPMENT':
        return { label: 'View Shipment', path: id ? `/dashboard/shipments/${id}` : '/dashboard/shipments' };
      case 'TRACKING':
        return { label: 'View Tracking', path: id ? `/dashboard/tracking/${id}` : '/dashboard/tracking' };
      case 'CUSTOMER':
        return { label: 'View Customer', path: id ? `/dashboard/customers/${id}` : '/dashboard/customers' };
      case 'RATE':
      case 'RATE_MANAGEMENT':
        return { label: 'View Rate', path: '/dashboard/rate-management' };
      case 'CONTRACT':
        return { label: 'View Contract', path: '/dashboard/contracts' };
      case 'INVOICE':
        return { label: 'View Invoice', path: '/dashboard/invoices' };
      case 'PAYMENT':
        return { label: 'View Invoices & Payments', path: '/dashboard/invoices' };
      case 'CARRIER':
      case 'CARRIER_INTEGRATION':
        return { label: 'View Carrier Integrations', path: '/dashboard/settings/carrier-integrations' };
      case 'USER':
        return { label: 'View Users & Team', path: '/dashboard/settings/users' };
      case 'ROLE':
        return { label: 'View Roles & Permissions', path: '/dashboard/settings/roles' };
      case 'APPROVAL':
        return { label: 'View Approvals', path: '/dashboard/approvals' };
      case 'DOCUMENT':
        return { label: 'View Documents', path: '/dashboard/documents' };
      case 'OUTREACH':
      case 'CAMPAIGN':
        return { label: 'View Outreach', path: '/dashboard/outreach' };
      default:
        return null;
    }
  };

  // Label resolvers for Dropdown triggers
  const getDateFilterLabel = () => {
    if (datePreset === 'ALL') return 'All Time';
    if (datePreset === 'TODAY') return 'Today';
    if (datePreset === 'YESTERDAY') return 'Yesterday';
    if (datePreset === '7D') return 'Last 7 Days';
    if (datePreset === '30D') return 'Last 30 Days';
    if (datePreset === 'CUSTOM') {
      if (customStartDate && customEndDate) {
        return `${customStartDate} - ${customEndDate}`;
      }
      return 'Custom Range';
    }
    return 'Date';
  };

  const getUserFilterLabel = () => {
    if (selectedUserId === 'ALL') return 'All Users';
    if (selectedUserId === 'SYSTEM') return 'System';
    const found = usersList.find(u => (u.user_id || u.id) === Number(selectedUserId));
    if (found) {
      return found.first_name ? `${found.first_name} ${found.last_name || ''}`.trim() : found.email;
    }
    return 'User';
  };

  const getModuleFilterLabel = () => {
    if (selectedModule === 'ALL') return 'All Modules';
    return MODULE_CONFIG[selectedModule]?.label || selectedModule;
  };

  const getActionFilterLabel = () => {
    if (selectedAction === 'ALL') return 'All Actions';
    return normalizeActionName(selectedAction);
  };

  const getResultFilterLabel = () => {
    if (selectedResult === 'ALL') return 'All Results';
    return selectedResult === 'SUCCESS' ? 'Success' : 'Failed';
  };

  return (
    <div className={`audit-logs-page ${selectedLog ? 'with-drawer-open' : ''}`}>
      
      {/* ── Center Content Area: Header, Filter Bar, Table, Pagination ── */}
      <div className="audit-main-content">
        
        {/* Page Header */}
        <div className="audit-page-header">
          <div className="audit-header-left">
            <div className="audit-title-icon">
              <Shield className="w-5 h-5 text-blue-600" />
            </div>
            <div>
              <h1 className="audit-page-title">Audit Logs</h1>
              <p className="audit-page-subtitle">Track security and business activity across your organization.</p>
            </div>
          </div>
          <button
            type="button"
            className="audit-export-btn"
            onClick={handleExportCSV}
            title="Export filtered logs to CSV"
          >
            <Download className="w-4 h-4" />
            Export CSV
          </button>
        </div>

        {/* Filter Controls Row */}
        <div className="audit-filters-bar" ref={filterBarRef}>
          
          {/* 1. Date Filter Dropdown */}
          <div className="audit-filter-dropdown-wrapper">
            <button
              type="button"
              className={`audit-filter-btn ${datePreset !== 'ALL' ? 'active-filter' : ''}`}
              onClick={() => setOpenDropdown(openDropdown === 'date' ? null : 'date')}
            >
              <Calendar className="audit-filter-icon" />
              <span>{getDateFilterLabel()}</span>
              <ChevronDown className="audit-filter-chevron" />
            </button>

            {openDropdown === 'date' && (
              <div className="audit-dropdown-menu audit-date-menu">
                <button
                  type="button"
                  className={`audit-dropdown-item ${datePreset === 'ALL' ? 'selected' : ''}`}
                  onClick={() => { setDatePreset('ALL'); setOpenDropdown(null); setPage(1); }}
                >
                  All Time
                </button>
                <button
                  type="button"
                  className={`audit-dropdown-item ${datePreset === 'TODAY' ? 'selected' : ''}`}
                  onClick={() => { setDatePreset('TODAY'); setOpenDropdown(null); setPage(1); }}
                >
                  Today
                </button>
                <button
                  type="button"
                  className={`audit-dropdown-item ${datePreset === 'YESTERDAY' ? 'selected' : ''}`}
                  onClick={() => { setDatePreset('YESTERDAY'); setOpenDropdown(null); setPage(1); }}
                >
                  Yesterday
                </button>
                <button
                  type="button"
                  className={`audit-dropdown-item ${datePreset === '7D' ? 'selected' : ''}`}
                  onClick={() => { setDatePreset('7D'); setOpenDropdown(null); setPage(1); }}
                >
                  Last 7 Days
                </button>
                <button
                  type="button"
                  className={`audit-dropdown-item ${datePreset === '30D' ? 'selected' : ''}`}
                  onClick={() => { setDatePreset('30D'); setOpenDropdown(null); setPage(1); }}
                >
                  Last 30 Days
                </button>
                
                <div className="audit-dropdown-divider" />
                
                <div className="audit-custom-date-section">
                  <div className="audit-custom-date-title">Custom Range</div>
                  <div className="audit-date-inputs">
                    <input
                      type="date"
                      className="audit-date-input"
                      value={customStartDate}
                      onChange={(e) => { setCustomStartDate(e.target.value); setDatePreset('CUSTOM'); }}
                      placeholder="Start date"
                    />
                    <span className="audit-date-to">to</span>
                    <input
                      type="date"
                      className="audit-date-input"
                      value={customEndDate}
                      onChange={(e) => { setCustomEndDate(e.target.value); setDatePreset('CUSTOM'); }}
                      placeholder="End date"
                    />
                  </div>
                  <button
                    type="button"
                    className="audit-apply-date-btn"
                    onClick={() => { setDatePreset('CUSTOM'); setOpenDropdown(null); setPage(1); }}
                  >
                    Apply Range
                  </button>
                </div>
              </div>
            )}
          </div>

          {/* 2. User Filter Dropdown */}
          <div className="audit-filter-dropdown-wrapper">
            <button
              type="button"
              className={`audit-filter-btn ${selectedUserId !== 'ALL' ? 'active-filter' : ''}`}
              onClick={() => setOpenDropdown(openDropdown === 'user' ? null : 'user')}
            >
              <User className="audit-filter-icon" />
              <span>{getUserFilterLabel()}</span>
              <ChevronDown className="audit-filter-chevron" />
            </button>

            {openDropdown === 'user' && (
              <div className="audit-dropdown-menu">
                <button
                  type="button"
                  className={`audit-dropdown-item ${selectedUserId === 'ALL' ? 'selected' : ''}`}
                  onClick={() => { setSelectedUserId('ALL'); setOpenDropdown(null); setPage(1); }}
                >
                  All Users
                </button>
                <button
                  type="button"
                  className={`audit-dropdown-item ${selectedUserId === 'SYSTEM' ? 'selected' : ''}`}
                  onClick={() => { setSelectedUserId('SYSTEM'); setOpenDropdown(null); setPage(1); }}
                >
                  System
                </button>
                <div className="audit-dropdown-divider" />
                {usersList.map((u) => {
                  const uid = u.user_id || u.id;
                  const name = u.first_name ? `${u.first_name} ${u.last_name || ''}`.trim() : u.email;
                  return (
                    <button
                      key={uid}
                      type="button"
                      className={`audit-dropdown-item ${String(selectedUserId) === String(uid) ? 'selected' : ''}`}
                      onClick={() => { setSelectedUserId(String(uid)); setOpenDropdown(null); setPage(1); }}
                    >
                      <span className="truncate">{name}</span>
                      {u.role_name && <span className="audit-item-subtag">{u.role_name}</span>}
                    </button>
                  );
                })}
              </div>
            )}
          </div>

          {/* 3. Module Filter Dropdown */}
          <div className="audit-filter-dropdown-wrapper">
            <button
              type="button"
              className={`audit-filter-btn ${selectedModule !== 'ALL' ? 'active-filter' : ''}`}
              onClick={() => setOpenDropdown(openDropdown === 'module' ? null : 'module')}
            >
              <Layers className="audit-filter-icon" />
              <span>{getModuleFilterLabel()}</span>
              <ChevronDown className="audit-filter-chevron" />
            </button>

            {openDropdown === 'module' && (
              <div className="audit-dropdown-menu audit-scroll-menu">
                <button
                  type="button"
                  className={`audit-dropdown-item ${selectedModule === 'ALL' ? 'selected' : ''}`}
                  onClick={() => { setSelectedModule('ALL'); setOpenDropdown(null); setPage(1); }}
                >
                  All Modules
                </button>
                <div className="audit-dropdown-divider" />
                {Object.entries(MODULE_CONFIG).filter(([k]) => k !== 'GENERAL').map(([key, config]) => {
                  const Icon = config.icon;
                  return (
                    <button
                      key={key}
                      type="button"
                      className={`audit-dropdown-item ${selectedModule === key ? 'selected' : ''}`}
                      onClick={() => { setSelectedModule(key); setOpenDropdown(null); setPage(1); }}
                    >
                      <Icon className="w-3.5 h-3.5 mr-2 opacity-70" style={{ color: config.color }} />
                      <span>{config.label}</span>
                    </button>
                  );
                })}
              </div>
            )}
          </div>

          {/* 4. Action Filter Dropdown */}
          <div className="audit-filter-dropdown-wrapper">
            <button
              type="button"
              className={`audit-filter-btn ${selectedAction !== 'ALL' ? 'active-filter' : ''}`}
              onClick={() => setOpenDropdown(openDropdown === 'action' ? null : 'action')}
            >
              <Zap className="audit-filter-icon" />
              <span>{getActionFilterLabel()}</span>
              <ChevronDown className="audit-filter-chevron" />
            </button>

            {openDropdown === 'action' && (
              <div className="audit-dropdown-menu audit-scroll-menu">
                <button
                  type="button"
                  className={`audit-dropdown-item ${selectedAction === 'ALL' ? 'selected' : ''}`}
                  onClick={() => { setSelectedAction('ALL'); setOpenDropdown(null); setPage(1); }}
                >
                  All Actions
                </button>
                <div className="audit-dropdown-divider" />
                {['CREATE', 'UPDATE', 'DELETE', 'LOGIN', 'LOGIN_FAILED', 'LOGOUT', 'INVITE', 'ROLE_CHANGED', 'PERMISSION_CHANGED', 'APPROVE', 'REJECT', 'SEND', 'CONNECT', 'DISCONNECT', 'ENABLE', 'DISABLE', 'SYNC', 'PAYMENT_RECORDED'].map((act) => (
                  <button
                    key={act}
                    type="button"
                    className={`audit-dropdown-item ${selectedAction === act ? 'selected' : ''}`}
                    onClick={() => { setSelectedAction(act); setOpenDropdown(null); setPage(1); }}
                  >
                    {normalizeActionName(act)}
                  </button>
                ))}
              </div>
            )}
          </div>

          {/* 5. Result Filter Dropdown */}
          <div className="audit-filter-dropdown-wrapper">
            <button
              type="button"
              className={`audit-filter-btn ${selectedResult !== 'ALL' ? 'active-filter' : ''}`}
              onClick={() => setOpenDropdown(openDropdown === 'result' ? null : 'result')}
            >
              <CheckCircle2 className="audit-filter-icon" />
              <span>{getResultFilterLabel()}</span>
              <ChevronDown className="audit-filter-chevron" />
            </button>

            {openDropdown === 'result' && (
              <div className="audit-dropdown-menu">
                <button
                  type="button"
                  className={`audit-dropdown-item ${selectedResult === 'ALL' ? 'selected' : ''}`}
                  onClick={() => { setSelectedResult('ALL'); setOpenDropdown(null); setPage(1); }}
                >
                  All Results
                </button>
                <button
                  type="button"
                  className={`audit-dropdown-item ${selectedResult === 'SUCCESS' ? 'selected' : ''}`}
                  onClick={() => { setSelectedResult('SUCCESS'); setOpenDropdown(null); setPage(1); }}
                >
                  <span className="w-2 h-2 rounded-full bg-emerald-500 mr-2" />
                  Success
                </button>
                <button
                  type="button"
                  className={`audit-dropdown-item ${selectedResult === 'FAILED' ? 'selected' : ''}`}
                  onClick={() => { setSelectedResult('FAILED'); setOpenDropdown(null); setPage(1); }}
                >
                  <span className="w-2 h-2 rounded-full bg-rose-500 mr-2" />
                  Failed
                </button>
              </div>
            )}
          </div>

        </div>

        {/* Search Input Row with Clear Action */}
        <div className="audit-search-row">
          <div className="audit-search-input-wrapper">
            <Search className="audit-search-icon" />
            <input
              type="text"
              className="audit-search-input"
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              placeholder="Search by user, action, module, or record ID..."
            />
            {search && (
              <button
                type="button"
                className="audit-search-clear-btn"
                onClick={() => setSearch('')}
                title="Clear search text"
              >
                <X className="w-3.5 h-3.5" />
              </button>
            )}
          </div>
          {hasActiveFilters && (
            <button
              type="button"
              className="audit-clear-all-btn"
              onClick={handleClearAll}
            >
              Clear all
            </button>
          )}
        </div>

        {/* ── Table Container Card ── */}
        <div className="audit-table-card">
          
          {/* Loading Skeleton */}
          {isLoading && (
            <div className="audit-table-skeleton-container">
              <div className="audit-table-skeleton-header" />
              {[...Array(6)].map((_, i) => (
                <div key={i} className="audit-table-skeleton-row">
                  <div className="skeleton-cell w-20" />
                  <div className="skeleton-cell w-28" />
                  <div className="skeleton-cell w-24" />
                  <div className="skeleton-cell w-32" />
                  <div className="skeleton-cell w-24" />
                  <div className="skeleton-cell w-16" />
                </div>
              ))}
            </div>
          )}

          {/* Error State */}
          {!isLoading && error && (
            <div className="audit-empty-state">
              <AlertCircle className="w-10 h-10 text-rose-500 mb-3" />
              <h3 className="audit-empty-title">Failed to Load Audit Logs</h3>
              <p className="audit-empty-subtitle">{error}</p>
              <button
                type="button"
                className="audit-retry-btn"
                onClick={fetchLogs}
              >
                <RefreshCw className="w-4 h-4 mr-1.5" />
                Retry
              </button>
            </div>
          )}

          {/* Empty State: No logs matched or empty database */}
          {!isLoading && !error && logs.length === 0 && (
            <div className="audit-empty-state">
              <Shield className="w-12 h-12 text-slate-300 mb-3" />
              {hasActiveFilters ? (
                <>
                  <h3 className="audit-empty-title">No matching audit events</h3>
                  <p className="audit-empty-subtitle">Try adjusting your filters or search terms.</p>
                  <button
                    type="button"
                    className="audit-retry-btn mt-3"
                    onClick={handleClearAll}
                  >
                    Clear filters
                  </button>
                </>
              ) : (
                <>
                  <h3 className="audit-empty-title">No audit activity yet</h3>
                  <p className="audit-empty-subtitle">Important security events and workspace operations will appear here.</p>
                </>
              )}
            </div>
          )}

          {/* Real Table */}
          {!isLoading && !error && logs.length > 0 && (
            <div className="audit-table-responsive-wrapper">
              <table className="audit-table">
                <thead>
                  <tr>
                    <th className="th-time">Time ↓</th>
                    <th className="th-actor">User</th>
                    <th className="th-action">Action</th>
                    <th className="th-module">Module</th>
                    <th className="th-record">Record</th>
                    <th className="th-result">Result</th>
                    <th className="th-action-arrow" />
                  </tr>
                </thead>
                <tbody>
                  {logs.map((log) => {
                    const timeObj = formatAuditTime(log.created_at);
                    const avatar = getAvatarStyle(log.actor_name);
                    const modConfig = MODULE_CONFIG[log.module] || MODULE_CONFIG.GENERAL;
                    const ModIcon = modConfig.icon;
                    const actionCfg = ACTION_CONFIG[log.action] || {
                      label: normalizeActionName(log.action),
                      bg: '#F1F5F9',
                      color: '#475569',
                      border: '#E2E8F0'
                    };
                    const isSelected = selectedLog && selectedLog.id === log.id;
                    const isSuccess = log.result === 'SUCCESS';

                    // Display Record text
                    const recordText = log.resource_name || log.resource_id || '—';

                    return (
                      <tr
                        key={log.id}
                        className={`audit-table-row ${isSelected ? 'row-selected' : ''}`}
                        onClick={() => setSelectedLog(isSelected ? null : log)}
                        tabIndex={0}
                        onKeyDown={(e) => {
                          if (e.key === 'Enter' || e.key === ' ') {
                            e.preventDefault();
                            setSelectedLog(isSelected ? null : log);
                          }
                        }}
                      >
                        {/* Time Column */}
                        <td className="td-time">
                          <div className="time-primary">{timeObj.time}</div>
                          <div className="time-secondary">{timeObj.date}</div>
                        </td>

                        {/* Actor / User Column */}
                        <td className="td-actor">
                          <div className="actor-cell">
                            <div
                              className="actor-avatar"
                              style={{ backgroundColor: avatar.bg, color: avatar.color }}
                            >
                              {avatar.initial}
                            </div>
                            <div className="actor-info">
                              <span className="actor-name">{log.actor_name || (log.actor_type === 'SYSTEM' ? 'System' : 'Unknown User')}</span>
                              <span className="actor-role">{log.actor_role || (log.actor_type === 'SYSTEM' ? 'System Service' : 'Team Member')}</span>
                            </div>
                          </div>
                        </td>

                        {/* Action Column */}
                        <td className="td-action">
                          <span
                            className="action-pill"
                            style={{
                              backgroundColor: actionCfg.bg,
                              color: actionCfg.color,
                              borderColor: actionCfg.border
                            }}
                          >
                            {actionCfg.label}
                          </span>
                        </td>

                        {/* Module Column */}
                        <td className="td-module">
                          <div className="module-cell">
                            <div className="module-icon-box" style={{ color: modConfig.color }}>
                              <ModIcon className="w-4 h-4" />
                            </div>
                            <div className="module-info">
                              <span className="module-primary">{modConfig.label}</span>
                              {log.resource_name && log.resource_name !== log.resource_id && (
                                <span className="module-secondary">{log.resource_name}</span>
                              )}
                            </div>
                          </div>
                        </td>

                        {/* Record Column */}
                        <td className="td-record">
                          <span className="record-label" title={recordText}>
                            {recordText}
                          </span>
                        </td>

                        {/* Result Column */}
                        <td className="td-result">
                          <span className={`result-pill ${isSuccess ? 'result-success' : 'result-failed'}`}>
                            {isSuccess ? 'Success' : 'Failed'}
                          </span>
                        </td>

                        {/* Row Click Indicator Arrow */}
                        <td className="td-arrow">
                          <ChevronRight className="row-chevron" />
                        </td>
                      </tr>
                    );
                  })}
                </tbody>
              </table>
            </div>
          )}

          {/* ── Server Pagination Footer ── */}
          {!isLoading && !error && logs.length > 0 && (
            <div className="audit-pagination-footer">
              <div className="pagination-range-text">
                Showing {((page - 1) * limit) + 1} to {Math.min(page * limit, totalCount)} of {totalCount} results
              </div>

              <div className="pagination-controls">
                <button
                  type="button"
                  className="pagination-arrow-btn"
                  disabled={page <= 1}
                  onClick={() => setPage(p => Math.max(1, p - 1))}
                  title="Previous page"
                >
                  <ChevronLeft className="w-4 h-4" />
                </button>

                {Array.from({ length: totalPages }, (_, i) => i + 1)
                  .filter(p => p === 1 || p === totalPages || Math.abs(p - page) <= 1)
                  .map((p, idx, arr) => {
                    const prev = arr[idx - 1];
                    const showEllipsis = prev && p - prev > 1;
                    return (
                      <React.Fragment key={p}>
                        {showEllipsis && <span className="pagination-ellipsis">...</span>}
                        <button
                          type="button"
                          className={`pagination-page-btn ${page === p ? 'active' : ''}`}
                          onClick={() => setPage(p)}
                        >
                          {p}
                        </button>
                      </React.Fragment>
                    );
                  })}

                <button
                  type="button"
                  className="pagination-arrow-btn"
                  disabled={page >= totalPages}
                  onClick={() => setPage(p => Math.min(totalPages, p + 1))}
                  title="Next page"
                >
                  <ChevronRight className="w-4 h-4" />
                </button>
              </div>

              {/* Limit Selector */}
              <div className="pagination-limit-wrapper">
                <select
                  className="pagination-limit-select"
                  value={limit}
                  onChange={(e) => {
                    setLimit(Number(e.target.value));
                    setPage(1);
                  }}
                >
                  <option value={10}>10 per page</option>
                  <option value={20}>20 per page</option>
                  <option value={50}>50 per page</option>
                  <option value={100}>100 per page</option>
                </select>
              </div>
            </div>
          )}

        </div>

      </div>

      {/* ── Contextual Detail Panel / Drawer (Right side) ── */}
      {selectedLog && (
        <aside className="audit-detail-panel" aria-label="Audit Event Details">
          
          {/* Drawer Header */}
          <div className="detail-panel-header">
            <h2 className="detail-panel-title">Log Details</h2>
            <button
              type="button"
              className="detail-close-btn"
              onClick={() => setSelectedLog(null)}
              aria-label="Close detail panel"
              title="Close panel (Esc)"
            >
              <X className="w-5 h-5" />
            </button>
          </div>

          <div className="detail-panel-body">
            
            {/* Status Result Pill */}
            <div className="detail-status-section">
              <span className={`detail-status-pill ${selectedLog.result === 'SUCCESS' ? 'status-success' : 'status-failed'}`}>
                {selectedLog.result === 'SUCCESS' ? (
                  <>
                    <CheckCircle2 className="w-3.5 h-3.5 mr-1" />
                    Success
                  </>
                ) : (
                  <>
                    <XCircle className="w-3.5 h-3.5 mr-1" />
                    Failed
                  </>
                )}
              </span>
            </div>

            {/* Main Action Title & Timestamp */}
            <div className="detail-event-header">
              <h3 className="detail-event-title">
                {selectedLog.description || `${normalizeActionName(selectedLog.action)} ${selectedLog.resource_type || ''}`.trim()}
              </h3>
              <p className="detail-event-timestamp">
                {formatAuditTime(selectedLog.created_at).full}
              </p>
            </div>

            {/* If Failure: Display Error Message */}
            {selectedLog.result === 'FAILED' && selectedLog.error_message && (
              <div className="detail-error-box">
                <div className="detail-error-label">
                  <AlertCircle className="w-4 h-4 text-rose-500 mr-1.5" />
                  Failure Reason
                </div>
                <div className="detail-error-text">{selectedLog.error_message}</div>
              </div>
            )}

            {/* Core Metadata Attribute Pairs */}
            <div className="detail-attributes-card">
              
              {/* User */}
              <div className="detail-attr-row">
                <div className="detail-attr-label">
                  <User className="w-4 h-4 text-slate-400" />
                  <span>User</span>
                </div>
                <div className="detail-attr-value font-medium text-slate-900">
                  {selectedLog.actor_name || (selectedLog.actor_type === 'SYSTEM' ? 'System' : '—')}
                </div>
              </div>

              {/* Role */}
              <div className="detail-attr-row">
                <div className="detail-attr-label">
                  <ShieldCheck className="w-4 h-4 text-slate-400" />
                  <span>Role</span>
                </div>
                <div className="detail-attr-value">
                  {selectedLog.actor_role || (selectedLog.actor_type === 'SYSTEM' ? 'System Service' : 'Team Member')}
                </div>
              </div>

              {/* Module */}
              <div className="detail-attr-row">
                <div className="detail-attr-label">
                  <Layers className="w-4 h-4 text-slate-400" />
                  <span>Module</span>
                </div>
                <div className="detail-attr-value">
                  {MODULE_CONFIG[selectedLog.module]?.label || selectedLog.module || 'General'}
                </div>
              </div>

              {/* Record ID / Name */}
              {selectedLog.resource_id && (
                <div className="detail-attr-row">
                  <div className="detail-attr-label">
                    <FileText className="w-4 h-4 text-slate-400" />
                    <span>Record ID</span>
                  </div>
                  <div className="detail-attr-value font-mono">
                    {selectedLog.resource_id}
                    {selectedLog.resource_name && selectedLog.resource_name !== selectedLog.resource_id && (
                      <span className="block text-xs text-slate-500 font-sans mt-0.5">{selectedLog.resource_name}</span>
                    )}
                  </div>
                </div>
              )}

              {/* IP Address */}
              {selectedLog.ip_address && (
                <div className="detail-attr-row">
                  <div className="detail-attr-label">
                    <MapPin className="w-4 h-4 text-slate-400" />
                    <span>IP Address</span>
                  </div>
                  <div className="detail-attr-value font-mono text-xs">
                    {selectedLog.ip_address}
                  </div>
                </div>
              )}

              {/* User Agent / Client */}
              {selectedLog.user_agent && (
                <div className="detail-attr-row">
                  <div className="detail-attr-label">
                    <Laptop className="w-4 h-4 text-slate-400" />
                    <span>Client / Browser</span>
                  </div>
                  <div className="detail-attr-value text-xs text-slate-600 truncate" title={selectedLog.user_agent}>
                    {selectedLog.user_agent}
                  </div>
                </div>
              )}

            </div>

            {/* ── Changes Section (Before vs After Diffs) ── */}
            {((selectedLog.changes && selectedLog.changes.length > 0) || (selectedLog.before_data && selectedLog.after_data)) && (
              <div className="detail-changes-section">
                <h4 className="detail-section-heading">Changes</h4>
                <div className="detail-changes-table-wrapper">
                  <table className="detail-changes-table">
                    <thead>
                      <tr>
                        <th>Field</th>
                        <th>Before</th>
                        <th>After</th>
                      </tr>
                    </thead>
                    <tbody>
                      {selectedLog.changes && selectedLog.changes.length > 0 ? (
                        selectedLog.changes.map((ch, idx) => (
                          <tr key={idx}>
                            <td className="font-medium text-slate-700 capitalize">
                              {ch.field?.replace(/_/g, ' ') || 'Field'}
                            </td>
                            <td className="text-slate-500">
                              {formatDiffValue(ch.before)}
                            </td>
                            <td className="text-emerald-700 font-medium">
                              {formatDiffValue(ch.after)}
                            </td>
                          </tr>
                        ))
                      ) : (
                        // Render computed diff from before_data / after_data
                        Object.keys(selectedLog.after_data || {}).map((key) => {
                          const beforeVal = selectedLog.before_data ? selectedLog.before_data[key] : null;
                          const afterVal = selectedLog.after_data[key];
                          if (JSON.stringify(beforeVal) === JSON.stringify(afterVal)) return null;
                          return (
                            <tr key={key}>
                              <td className="font-medium text-slate-700 capitalize">
                                {key.replace(/_/g, ' ')}
                              </td>
                              <td className="text-slate-500">
                                {formatDiffValue(beforeVal)}
                              </td>
                              <td className="text-emerald-700 font-medium">
                                {formatDiffValue(afterVal)}
                              </td>
                            </tr>
                          );
                        })
                      )}
                    </tbody>
                  </table>
                </div>
              </div>
            )}

            {/* ── Metadata & Safe Context ── */}
            {selectedLog.after_data && !selectedLog.before_data && Object.keys(selectedLog.after_data).length > 0 && (
              <div className="detail-changes-section">
                <h4 className="detail-section-heading">Context Metadata</h4>
                <div className="detail-meta-list">
                  {Object.entries(selectedLog.after_data).map(([k, v]) => (
                    <div key={k} className="detail-meta-item">
                      <span className="detail-meta-key capitalize">{k.replace(/_/g, ' ')}:</span>
                      <span className="detail-meta-val">{formatDiffValue(v)}</span>
                    </div>
                  ))}
                </div>
              </div>
            )}

          </div>

          {/* Action CTA: View Real Underlying Resource Route */}
          {(() => {
            const route = getResourceRoute(selectedLog);
            if (!route) return null;
            return (
              <div className="detail-panel-footer">
                <button
                  type="button"
                  className="detail-action-link-btn"
                  onClick={() => navigate(route.path)}
                >
                  <ExternalLink className="w-4 h-4 mr-2" />
                  {route.label}
                </button>
              </div>
            );
          })()}

        </aside>
      )}

    </div>
  );
}

// ── Diff value formatter helper ───────────────────────────────────────────────
function formatDiffValue(val) {
  if (val === null || val === undefined || val === '') {
    return <span className="text-slate-400 italic">—</span>;
  }
  if (typeof val === 'boolean') {
    return val ? 'Yes' : 'No';
  }
  if (typeof val === 'number') {
    return val.toLocaleString();
  }
  if (typeof val === 'object') {
    try {
      return JSON.stringify(val);
    } catch {
      return String(val);
    }
  }
  return String(val);
}

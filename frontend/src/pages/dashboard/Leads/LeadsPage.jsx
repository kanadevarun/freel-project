import { useState, useEffect, useMemo, useCallback, useRef } from 'react';
import { createPortal } from 'react-dom';
import { listLeads, bulkUpdateLeads, getLead } from '../../../services/leadsService';
import api from '../../../services/api';
import PageHeader from '../../../components/dashboard/PageHeader';
import StatusBadge from '../../../components/dashboard/StatusBadge';
import LeadDetailPanel from './LeadDetailPanel';
import AddLeadModal from './AddLeadModal';
import ImportLeadsModal from './ImportLeadsModal';
import RFQBuilder from '../RFQ/RFQBuilder';
import LeadConversionModal from '../Customers/LeadConversionModal';
import { useRBAC } from '../../../context/RBACContext';
import ModuleHeroEmptyState from '../../../components/dashboard/ModuleHeroEmptyState';
import { Users, UploadCloud, Zap, Target, Building2, Plus } from 'lucide-react';
import './LeadsPage.css';

// ── Constants ─────────────────────────────────────────────────────────────────

export const LEAD_STATUS = {
  NEW:         'NEW',
  QUALIFIED:   'QUALIFIED',
  IN_PROGRESS: 'IN_PROGRESS',
  REJECTED:    'REJECTED',
  CONVERTED:   'CONVERTED',
};

const TABS = [
  { id: 'all',         label: 'All Leads',   statusFilter: null },
  { id: 'new',         label: 'New',         statusFilter: LEAD_STATUS.NEW },
  { id: 'in_progress', label: 'In Progress', statusFilter: LEAD_STATUS.IN_PROGRESS },
  { id: 'qualified',   label: 'Qualified',   statusFilter: LEAD_STATUS.QUALIFIED },
  { id: 'converted',   label: 'Converted',   statusFilter: LEAD_STATUS.CONVERTED },
  { id: 'rejected',    label: 'Rejected',    statusFilter: LEAD_STATUS.REJECTED },
];

const STATUS_CFG = {
  [LEAD_STATUS.NEW]:         { label: 'New',         type: 'info' },
  [LEAD_STATUS.QUALIFIED]:   { label: 'Qualified',   type: 'success' },
  [LEAD_STATUS.IN_PROGRESS]: { label: 'In Progress', type: 'warning' },
  [LEAD_STATUS.REJECTED]:    { label: 'Rejected',    type: 'danger' },
  [LEAD_STATUS.CONVERTED]:   { label: 'Converted',   type: 'converted' },
  'ACTIVE':                  { label: 'Active',      type: 'active' },
};

function SkeletonRows({ count = 6 }) {
  return Array.from({ length: count }).map((_, i) => (
    <tr key={i} className="leads-skeleton-row">
      {[150, 130, 80, 100, 70, 80, 50].map((w, j) => (
        <td key={j}><div className="leads-skeleton-cell" style={{ width: w }} /></td>
      ))}
    </tr>
  ));
}

export default function LeadsPage() {
  const { can } = useRBAC();
  const canCreateLead = can('LEADS', 'CREATE');

  // ── State ──────────────────────────────────────────────────────────────────
  const [allLeads, setAllLeads]           = useState([]);   // Paginated leads from API
  const [loading, setLoading]             = useState(true); // API loading state
  const [activeTab, setActiveTab]         = useState('all');
  const [searchQuery, setSearchQuery]     = useState('');
  const [activeSourceFilter, setActiveSourceFilter] = useState('');
  const [selectedLead, setSelectedLead]   = useState(null); // Lead shown in panel
  const [isClosingDrawer, setIsClosingDrawer] = useState(false);
  const [drawerIsDirty, setDrawerIsDirty] = useState(false);
  const [activeLeadForDrawer, setActiveLeadForDrawer] = useState(null);
  const [showAddModal, setShowAddModal]   = useState(false);
  const [showImportModal, setShowImportModal] = useState(false);
  const [convertingLead, setConvertingLead] = useState(null);
  const [convertingCustomerLead, setConvertingCustomerLead] = useState(null);
  const [showFilterPopover, setShowFilterPopover] = useState(false);

  // Pagination states
  const [currentPage, setCurrentPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const [totalCount, setTotalCount] = useState(0);

  // Database-wide counts for stats cards and tabs
  const [dbCounts, setDbCounts] = useState({
    all: 0,
    new: 0,
    qualified: 0,
    in_progress: 0,
    converted: 0,
    rejected: 0,
  });

  // Bulk actions & user list
  const [selectedLeadIDs, setSelectedLeadIDs] = useState([]);
  const [users, setUsers] = useState([]);
  const [bulkError, setBulkError] = useState('');
  const [bulkSuccess, setBulkSuccess] = useState('');
  const [bulkActionLoading, setBulkActionLoading] = useState(false);

  const [drawerInitialTab, setDrawerInitialTab] = useState('overview');

  // ── Handle URL Search Parameters (Part 4) ──────────────────────────────────
  useEffect(() => {
    const searchParams = new URLSearchParams(window.location.search);
    const urlLeadId = searchParams.get('leadId');
    const urlTab = searchParams.get('tab');
    if (urlLeadId) {
      const targetTab = (urlTab === 'email' || urlTab === 'emails') ? 'emails' : 'overview';
      setDrawerInitialTab(targetTab);
      getLead(urlLeadId)
        .then(res => {
          const lData = res?.data || res;
          if (lData && lData.id) {
            setSelectedLead(lData);
            setActiveLeadForDrawer(lData);
          }
        })
        .catch(err => console.error('Failed to load lead from URL param:', err));
    }
  }, []);

  // ── Fetch Active Users ─────────────────────────────────────────────────────
  useEffect(() => {
    const fetchUsers = async () => {
      try {
        const res = await api.get('/api/v1/users');
        const userList = res?.data || res || [];
        setUsers(userList.filter(u => u.status === 'ACTIVE'));
      } catch (err) {
        console.error('Failed to fetch active users:', err);
      }
    };
    fetchUsers();
  }, []);

  // ── Fetch Database-wide Counts ─────────────────────────────────────────────
  const fetchDbCounts = useCallback(async () => {
    try {
      const [allRes, newRes, qualRes, ipRes, convRes, rejRes] = await Promise.all([
        listLeads({ limit: 0 }),
        listLeads({ limit: 0, status: 'NEW' }),
        listLeads({ limit: 0, status: 'QUALIFIED' }),
        listLeads({ limit: 0, status: 'IN_PROGRESS' }),
        listLeads({ limit: 0, status: 'CONVERTED' }),
        listLeads({ limit: 0, status: 'REJECTED' }),
      ]);
      setDbCounts({
        all: allRes?.total_count || 0,
        new: newRes?.total_count || 0,
        qualified: qualRes?.total_count || 0,
        in_progress: ipRes?.total_count || 0,
        converted: convRes?.total_count || 0,
        rejected: rejRes?.total_count || 0,
      });
    } catch (e) {
      console.error('Failed to fetch DB counts:', e);
    }
  }, []);

  // ── Fetch Paginated Leads ──────────────────────────────────────────────────
  const fetchLeads = useCallback(async () => {
    setLoading(true);
    setSelectedLeadIDs([]); // Clear selection when data changes/reloads
    try {
      const statusParam = activeTab !== 'all' ? activeTab.toUpperCase() : undefined;
      const offset = (currentPage - 1) * pageSize;
      const res = await listLeads({
        limit: pageSize,
        offset,
        status: statusParam,
        search: searchQuery || undefined,
        source: activeSourceFilter || undefined,
      });

      const leadsList = Array.isArray(res) ? res : (res?.data || res?.leads || []);
      const count = res?.total_count || leadsList.length;

      setAllLeads(leadsList);
      setTotalCount(count);

      if (currentPage > 1 && leadsList.length === 0 && count > 0) {
        const maxPage = Math.ceil(count / pageSize);
        setCurrentPage(Math.max(1, maxPage));
        return;
      }
    } catch (err) {
      console.error('Failed to load leads:', err);
    } finally {
      setLoading(false);
    }
  }, [currentPage, pageSize, activeTab, searchQuery, activeSourceFilter]);

  // Bulk Actions
  const handleBulkUpdate = async (updateFields) => {
    if (selectedLeadIDs.length === 0) return;
    
    setBulkActionLoading(true);
    setBulkError('');
    setBulkSuccess('');
    
    try {
      const payload = {
        ids: selectedLeadIDs,
        ...updateFields
      };
      
      const res = await bulkUpdateLeads(payload);
      
      const successCount = res?.data?.success_ids?.length || res?.success_ids?.length || 0;
      const failedMap = res?.data?.failed_ids || res?.failed_ids || {};
      const failedCount = Object.keys(failedMap).length;
      
      if (failedCount > 0) {
        const errorList = Object.entries(failedMap).map(([id, reason]) => `Lead ID ${id}: ${reason}`).join('; ');
        setBulkError(`Updated ${successCount} successfully. Failed ${failedCount} updates. Details: ${errorList}`);
      } else {
        setBulkSuccess(`Successfully updated ${successCount} leads!`);
      }
      
      fetchLeads();
      fetchDbCounts();
      setSelectedLeadIDs([]);
    } catch (err) {
      setBulkError(err.message || 'Bulk update failed.');
    } finally {
      setBulkActionLoading(false);
    }
  };

  // Safe wrapper for setting selected lead with dirty check
  const changeSelectedLead = useCallback((nextLead) => {
    if (drawerIsDirty) {
      const confirmDiscard = window.confirm("You have unsaved changes. Do you want to discard them?");
      if (!confirmDiscard) {
        return;
      }
    }
    setDrawerIsDirty(false);
    setSelectedLead(nextLead);
  }, [drawerIsDirty]);

  // Synchronize lead update/delete from details panel
  const handleLeadUpdated = useCallback((updatedLead) => {
    fetchLeads();
    fetchDbCounts();
    if (updatedLead) {
      setSelectedLead(updatedLead);
    } else {
      setSelectedLead(null);
    }
  }, [fetchLeads, fetchDbCounts]);

  // Trigger fetches
  useEffect(() => {
    fetchLeads();
  }, [fetchLeads]);

  useEffect(() => {
    fetchDbCounts();
  }, [fetchDbCounts]);

  // Handle escape key to close details panel if no modal is open
  useEffect(() => {
    const handleKeyDown = (e) => {
      if (e.key === 'Escape') {
        const hasOpenModal = document.querySelector('.leads-modal-overlay');
        if (!hasOpenModal) {
          changeSelectedLead(null);
        }
      }
    };
    if (selectedLead) {
      document.addEventListener('keydown', handleKeyDown);
    }
    return () => {
      document.removeEventListener('keydown', handleKeyDown);
    };
  }, [selectedLead]);

  // Handle drawer opening and closing animations
  useEffect(() => {
    if (selectedLead) {
      setActiveLeadForDrawer(selectedLead);
      setIsClosingDrawer(false);
    } else if (activeLeadForDrawer) {
      setIsClosingDrawer(true);
      const timer = setTimeout(() => {
        setActiveLeadForDrawer(null);
        setIsClosingDrawer(false);
      }, 250);
      return () => clearTimeout(timer);
    }
  }, [selectedLead, activeLeadForDrawer]);

  // Lock body scroll when drawer is open
  useEffect(() => {
    if (activeLeadForDrawer) {
      document.body.style.overflow = 'hidden';
    } else {
      document.body.style.overflow = '';
    }
    return () => {
      document.body.style.overflow = '';
    };
  }, [activeLeadForDrawer]);

  // ── Reset Page on Filter Changes ───────────────────────────────────────────
  const handleTabChange = (tabId) => {
    setActiveTab(tabId);
    setCurrentPage(1);
  };

  const handleSearchChange = (e) => {
    setSearchQuery(e.target.value);
    setCurrentPage(1);
  };

  const handleSourceChange = (e) => {
    setActiveSourceFilter(e.target.value);
    setCurrentPage(1);
  };

  const handlePageSizeChange = (e) => {
    setPageSize(Number(e.target.value));
    setCurrentPage(1);
  };

  const handleResetFilters = () => {
    setSearchQuery('');
    setActiveSourceFilter('');
    setCurrentPage(1);
  };

  // Derive total pages
  const totalPages = Math.max(1, Math.ceil(totalCount / pageSize));

  // Derive union of tags on currently selected leads
  const selectedLeadsTags = useMemo(() => {
    const tagsSet = new Set();
    allLeads.forEach(lead => {
      if (selectedLeadIDs.includes(lead.id) && lead.tags) {
        lead.tags.forEach(tag => tagsSet.add(tag));
      }
    });
    return Array.from(tagsSet);
  }, [selectedLeadIDs, allLeads]);

  // Determine active filter count
  const activeFilterCount = (searchQuery ? 1 : 0) + (activeSourceFilter ? 1 : 0);

  return (
    <div className="leads-page">
      <style>{`
        .app-shell-main {
          overflow: visible !important;
        }
        .app-shell-container {
          height: auto !important;
          min-height: 100% !important;
          overflow: visible !important;
        }
        .leads-page {
          display: block !important;
          height: auto !important;
          min-height: 100% !important;
          overflow: visible !important;
        }

        .leads-empty {
          display: flex;
          flex-direction: column;
          align-items: center;
          justify-content: center;
          padding: 60px 40px;
          text-align: center;
          background: #ffffff;
          border: 1px dashed #CBD5E1;
          border-radius: 16px;
          margin: 20px 0;
          box-shadow: 0 10px 30px rgba(15, 23, 42, 0.02);
        }
        .leads-empty-icon {
          font-size: 44px;
          margin-bottom: 16px;
          display: flex;
          align-items: center;
          justify-content: center;
          width: 80px;
          height: 80px;
          background: #F1F5F9;
          border-radius: 50%;
          color: #3b82f6;
          box-shadow: inset 0 2px 4px rgba(0, 0, 0, 0.06);
          animation: pulseGlow 3s infinite ease-in-out;
        }
        .leads-empty-title {
          font-size: 18px;
          font-weight: 600;
          color: #1E293B;
          margin-bottom: 8px;
          font-family: 'Outfit', sans-serif;
        }
        .leads-empty-sub {
          font-size: 14px;
          color: #64748B;
          max-width: 380px;
          line-height: 1.5;
          margin-bottom: 24px;
        }
        .leads-empty-actions {
          display: flex;
          gap: 12px;
        }
        .leads-empty-btn {
          display: flex;
          align-items: center;
          gap: 8px;
          padding: 10px 20px;
          border-radius: 8px;
          font-size: 13.5px;
          font-weight: 500;
          cursor: pointer;
          transition: all 0.2s ease;
          border: none;
        }
        .leads-empty-btn.primary {
          background: #0F172A;
          color: #ffffff;
        }
        .leads-empty-btn.primary:hover {
          background: #1E293B;
          transform: translateY(-1px);
        }
        .leads-empty-btn.secondary {
          background: #ffffff;
          color: #475569;
          border: 1px solid #E2E8F0;
        }
        .leads-empty-btn.secondary:hover {
          background: #F8FAFC;
          color: #0F172A;
        }
        @keyframes pulseGlow {
          0%, 100% { transform: scale(1); box-shadow: 0 0 0 0 rgba(59, 130, 246, 0.2); }
          50% { transform: scale(1.05); box-shadow: 0 0 20px 10px rgba(59, 130, 246, 0.1); }
        }
      `}</style>
      {/* Page Header */}
      <div className="leads-header-row">
        <div>
          <div className="leads-badge-pill">
            <span>⚡ Sales & Lead Pipeline</span>
          </div>
          <h1 className="leads-page-title">Leads & Inbound Inquiries</h1>
          <p className="leads-page-subtitle">
            Discover, score, and qualify high-value shipping prospects into active commercial RFQs.
          </p>
        </div>

        <div className="leads-header-actions">
          <button 
            className="leads-btn leads-btn-ghost" 
            onClick={() => setShowImportModal(true)}
            disabled={!canCreateLead}
            title={!canCreateLead ? 'You do not have permission to create leads' : ''}
            style={!canCreateLead ? { opacity: 0.5, cursor: 'not-allowed' } : {}}
          >
            📥 Import CSV
          </button>
          <button 
            className="leads-btn leads-btn-primary" 
            onClick={() => setShowAddModal(true)}
            disabled={!canCreateLead}
            title={!canCreateLead ? 'You do not have permission to create leads' : ''}
            style={!canCreateLead ? { opacity: 0.5, cursor: 'not-allowed', pointerEvents: 'none' } : {}}
          >
            + Add Lead
          </button>
        </div>
      </div>

      {/* ─── Stat Cards (Interactive & Rich Gradient Styling) ─── */}
      <div className="leads-stats-row">
        
        {/* Total Leads */}
        <div 
          className={`leads-stat-card stat-all ${activeTab === 'all' ? 'active' : ''}`}
          onClick={() => handleTabChange('all')}
          role="button"
          tabIndex={0}
        >
          <div className="leads-stat-card-top">
            <div className="leads-stat-icon">
              <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.2" strokeLinecap="round" strokeLinejoin="round">
                <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2" />
                <circle cx="9" cy="7" r="4" />
                <path d="M23 21v-2a4 4 0 0 0-3-3.87" />
                <path d="M16 3.13a4 4 0 0 1 0 7.75" />
              </svg>
            </div>
            <span className="leads-card-pill">All Prospects</span>
          </div>
          <div className="leads-stat-info">
            <div className="leads-stat-value">{loading && dbCounts.all === 0 ? '···' : dbCounts.all}</div>
            <div className="leads-stat-label">Total Leads</div>
          </div>
          <div className="leads-stat-subtext">Entire database pipeline</div>
        </div>

        {/* New */}
        <div 
          className={`leads-stat-card stat-new ${activeTab === 'new' ? 'active' : ''}`}
          onClick={() => handleTabChange('new')}
          role="button"
          tabIndex={0}
        >
          <div className="leads-stat-card-top">
            <div className="leads-stat-icon">
              <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.2" strokeLinecap="round" strokeLinejoin="round">
                <path d="m12 3-1.912 5.813a2 2 0 0 1-1.275 1.275L3 12l5.813 1.912a2 2 0 0 1 1.275 1.275L12 21l1.912-5.813a2 2 0 0 1 1.275-1.275L21 12l-5.813-1.912a2 2 0 0 1-1.275-1.275L12 3Z" />
              </svg>
            </div>
            <span className="leads-card-pill">Fresh Inbound</span>
          </div>
          <div className="leads-stat-info">
            <div className="leads-stat-value">{loading && dbCounts.new === 0 ? '···' : dbCounts.new}</div>
            <div className="leads-stat-label">New Leads</div>
          </div>
          <div className="leads-stat-subtext">Awaiting initial qualification</div>
        </div>

        {/* Qualified */}
        <div 
          className={`leads-stat-card stat-qualified ${activeTab === 'qualified' ? 'active' : ''}`}
          onClick={() => handleTabChange('qualified')}
          role="button"
          tabIndex={0}
        >
          <div className="leads-stat-card-top">
            <div className="leads-stat-icon">
              <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.2" strokeLinecap="round" strokeLinejoin="round">
                <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z" />
                <path d="m9 11 2 2 4-4" />
              </svg>
            </div>
            <span className="leads-card-pill">Validated</span>
          </div>
          <div className="leads-stat-info">
            <div className="leads-stat-value">{loading && dbCounts.qualified === 0 ? '···' : dbCounts.qualified}</div>
            <div className="leads-stat-label">Qualified Leads</div>
          </div>
          <div className="leads-stat-subtext">Verified freight intent & readiness</div>
        </div>

        {/* In Progress */}
        <div 
          className={`leads-stat-card stat-in-progress ${activeTab === 'in_progress' ? 'active' : ''}`}
          onClick={() => handleTabChange('in_progress')}
          role="button"
          tabIndex={0}
        >
          <div className="leads-stat-card-top">
            <div className="leads-stat-icon">
              <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.2" strokeLinecap="round" strokeLinejoin="round">
                <path d="M5 2h14" />
                <path d="M5 22h14" />
                <path d="M19 2v4c0 4-4 6-4 6s4 2 4 6v4" />
                <path d="M5 2v4c0 4 4 6 4 6s-4 2-4 6v4" />
              </svg>
            </div>
            <span className="leads-card-pill">Active Sales</span>
          </div>
          <div className="leads-stat-info">
            <div className="leads-stat-value">{loading && dbCounts.in_progress === 0 ? '···' : dbCounts.in_progress}</div>
            <div className="leads-stat-label">In Progress</div>
          </div>
          <div className="leads-stat-subtext">Active negotiation & outreach</div>
        </div>

        {/* Converted to RFQ */}
        <div 
          className={`leads-stat-card stat-converted ${activeTab === 'converted' ? 'active' : ''}`}
          onClick={() => handleTabChange('converted')}
          role="button"
          tabIndex={0}
        >
          <div className="leads-stat-card-top">
            <div className="leads-stat-icon">
              <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.2" strokeLinecap="round" strokeLinejoin="round">
                <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" />
                <polyline points="14 2 14 8 20 8" />
                <line x1="16" y1="13" x2="8" y2="13" />
                <line x1="16" y1="17" x2="8" y2="17" />
              </svg>
            </div>
            <span className="leads-card-pill">Deals Won</span>
          </div>
          <div className="leads-stat-info">
            <div className="leads-stat-value">{loading && dbCounts.converted === 0 ? '···' : dbCounts.converted}</div>
            <div className="leads-stat-label">Converted to RFQ</div>
          </div>
          <div className="leads-stat-subtext">Handoff to commercial pricing</div>
        </div>

      </div>

      {loading && allLeads.length === 0 && (
        <div style={{
          display: 'flex',
          alignItems: 'center',
          gap: '8px',
          padding: '10px 16px',
          background: 'linear-gradient(90deg, #EFF6FF 0%, #EEF2FF 100%)',
          border: '1px solid #DBEAFE',
          borderRadius: '10px',
          marginBottom: '16px',
          fontSize: '12.5px',
          color: '#1E40AF',
          fontWeight: 600
        }}>
          <span style={{ animation: 'spin 1.5s linear infinite', display: 'inline-block' }}>🔄</span>
          <span>Loading sales leads from operational database...</span>
        </div>
      )}

      {/* Tabs */}
      <div className="leads-tabs">
        {TABS.map(tab => (
          <button
            key={tab.id}
            className={`leads-tab ${activeTab === tab.id ? 'active' : ''}`}
            onClick={() => handleTabChange(tab.id)}
          >
            {tab.label}
            <span className="leads-tab-count">
              {loading && dbCounts.all === 0 ? '···' : dbCounts[tab.id]}
            </span>
          </button>
        ))}
      </div>

      {/* Action Bar */}
      <div className="leads-action-bar">
        <div className="leads-search-wrap">
          <span className="leads-search-icon">🔍</span>
          <input
            type="text"
            className="leads-search-input"
            placeholder="Search by company, contact, email, or tags..."
            value={searchQuery}
            onChange={handleSearchChange}
          />
        </div>
        
        <div className="filter-wrapper">
          <button 
            className={`leads-filter-btn ${activeFilterCount > 0 ? 'active' : ''}`}
            onClick={() => setShowFilterPopover(!showFilterPopover)}
          >
            Filter {activeFilterCount > 0 ? `(${activeFilterCount})` : ''}
          </button>
          {showFilterPopover && (
            <div className="filter-popover">
              <div className="filter-popover-title">Filter Leads</div>
              <div className="filter-group">
                <label>Lead Source</label>
                <select
                  value={activeSourceFilter}
                  onChange={handleSourceChange}
                >
                  <option value="">All Sources</option>
                  <option value="LinkedIn">LinkedIn</option>
                  <option value="Website">Website</option>
                  <option value="Cold Outreach">Cold Outreach</option>
                  <option value="Referral">Referral</option>
                  <option value="Email">Email</option>
                </select>
              </div>
              {(searchQuery || activeSourceFilter) && (
                <button className="filter-reset-btn" onClick={handleResetFilters}>
                  Reset Filters
                </button>
              )}
            </div>
          )}
        </div>

        {activeFilterCount > 0 && (
          <button className="leads-reset-quick-btn" onClick={handleResetFilters}>
            ✕ Clear Filters
          </button>
        )}

        <div className="leads-bar-spacer" />
      </div>

      {/* ─── Split Container Layout ─── */}
      <div className="leads-page-container">
        
        {/* Left Main Section (Table or Hero Empty State) */}
        <div className="leads-main-section full-view">
          
          {!loading && dbCounts.all === 0 && !searchQuery && !activeSourceFilter && activeTab === 'all' ? (
            <ModuleHeroEmptyState
              icon={<Users size={28} />}
              badgeTheme="indigo"
              title="No Sales Leads in Your Pipeline"
              description="Start building your freight customer base by capturing shipper inquiries, importing contacts, or converting inbound quotes into active commercial pipelines."
              primaryAction={{
                label: 'Add First Lead',
                icon: <Plus size={15} />,
                onClick: () => setShowAddModal(true),
              }}
              secondaryAction={{
                label: 'Import CSV / Contacts',
                icon: <UploadCloud size={15} />,
                onClick: () => setShowImportModal(true),
              }}
              features={[
                {
                  icon: <Zap size={18} />,
                  iconBg: '#eef2ff',
                  iconColor: '#4f46e5',
                  title: 'Instant Lead-to-RFQ Conversion',
                  desc: 'Convert qualified shipper inquiries directly into priced shipment RFQs with one click.',
                },
                {
                  icon: <Target size={18} />,
                  iconBg: '#ecfdf5',
                  iconColor: '#059669',
                  title: 'Automated Outreach Sequences',
                  desc: 'Nurture potential shippers and exporters with multi-touch follow-up workflows.',
                },
                {
                  icon: <Building2 size={18} />,
                  iconBg: '#eff6ff',
                  iconColor: '#2563eb',
                  title: 'Shipper & Consignee Intelligence',
                  desc: 'Centralize contact details, preferred trade lanes, customs requirements, and volume history.',
                },
              ]}
            />
          ) : (
            /* Table Container */
            <div className="leads-table-wrap">
              <table className="leads-table">
                <thead>
                  <tr>
                    <th>Lead / Company</th>
                    <th>Contact</th>
                    <th>Source</th>
                    <th>Location</th>
                    <th>Status</th>
                    <th>Created</th>
                    <th style={{ width: '80px', textAlign: 'center' }}>Actions</th>
                  </tr>
                </thead>
                <tbody>
                  {loading ? (
                    <SkeletonRows count={pageSize} />
                  ) : allLeads.length === 0 ? (
                    <tr>
                      <td colSpan={7}>
                        <div className="leads-empty">
                          <div className="leads-empty-icon">
                            <svg width="36" height="36" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
                              <circle cx="11" cy="11" r="8" />
                              <line x1="21" y1="21" x2="16.65" y2="16.65" />
                              <line x1="8" y1="11" x2="14" y2="11" />
                              <line x1="11" y1="8" x2="11" y2="14" />
                            </svg>
                          </div>
                          <div className="leads-empty-title">
                            No matching leads found
                          </div>
                          <div className="leads-empty-sub">
                            Try adjusting your search query, filters, or status tab to find what you're looking for.
                          </div>
                          <div className="leads-empty-actions">
                            <button className="leads-empty-btn secondary" onClick={handleResetFilters}>
                              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><polyline points="1 4 1 10 7 10"/><path d="M3.51 15a9 9 0 1 0 2.13-9.36L1 10"/></svg>
                              Reset Filters
                            </button>
                            <button className="leads-empty-btn primary" onClick={() => setShowAddModal(true)}>
                              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
                              Add Lead
                            </button>
                          </div>
                        </div>
                      </td>
                    </tr>
                  ) : (
                    allLeads.map(lead => {
                    const statusCfg = STATUS_CFG[lead.status] || { label: lead.status, type: 'neutral' };
                    const avatar = lead.company_name?.slice(0, 2).toUpperCase() || '??';
                    const isSelected = selectedLead?.id === lead.id;

                    return (
                      <tr 
                        key={lead.id} 
                        className={`${isSelected ? 'selected' : ''} status-${lead.status?.toLowerCase() || 'new'}`}
                        onClick={() => changeSelectedLead(isSelected ? null : lead)}
                      >
                        <td>
                          <div className="lead-company-cell">
                            <div className="lead-avatar">{avatar}</div>
                            <div>
                              <div className="lead-company-name">{lead.company_name}</div>
                              <div className="lead-contact-name">{lead.contact_name || '—'}</div>
                              {lead.tags && lead.tags.length > 0 && (
                                <div className="lead-table-tags">
                                  {lead.tags.slice(0, 2).map((t, idx) => (
                                    <span key={idx} className="lead-tag-pill compact">{t}</span>
                                  ))}
                                  {lead.tags.length > 2 && (
                                    <span className="lead-tag-pill compact count">+{lead.tags.length - 2}</span>
                                  )}
                                </div>
                              )}
                            </div>
                          </div>
                        </td>
                        <td>
                          <div className="lead-contact-info">
                            <div className="lead-email">{lead.email || '—'}</div>
                            <div className="lead-phone">{lead.phone || '—'}</div>
                          </div>
                        </td>
                        <td>
                          <span style={{ fontSize: '12.5px', color: '#475569', fontWeight: 500 }}>
                            {lead.source || '—'}
                          </span>
                        </td>
                        <td>
                          <span style={{ fontSize: '12.5px', color: '#475569' }}>
                            {lead.location ? `📍 ${lead.location}` : '—'}
                          </span>
                        </td>
                        <td>
                          <StatusBadge
                             status={lead.status || 'NEW'}
                             customLabel={statusCfg.label}
                             customType={statusCfg.type}
                           />
                        </td>
                        <td style={{ color: '#94A3B8', fontSize: '12px' }}>
                          {lead.created_at ? new Date(lead.created_at).toLocaleDateString() : '—'}
                        </td>
                        <td className="actions-col" style={{ textAlign: 'center' }} onClick={(e) => e.stopPropagation()}>
                          <button 
                            className="lead-view-action-btn"
                            onClick={() => changeSelectedLead(isSelected ? null : lead)}
                          >
                            View
                          </button>
                        </td>
                      </tr>
                    );
                  })
                )}
              </tbody>
            </table>

            {/* Pagination Controls */}
            <div className="pagination-container">
              <div className="pagination-info">
                <span>
                  Showing {totalCount > 0 ? (currentPage - 1) * pageSize + 1 : 0} - {Math.min(currentPage * pageSize, totalCount)} of {totalCount} records
                </span>
                <select
                  className="pagination-size-select"
                  value={pageSize}
                  onChange={handlePageSizeChange}
                >
                  <option value={5}>5 per page</option>
                  <option value={10}>10 per page</option>
                  <option value={25}>25 per page</option>
                  <option value={50}>50 per page</option>
                </select>
              </div>
              <div className="pagination-controls">
                <button
                  className="pagination-btn"
                  disabled={currentPage <= 1 || loading}
                  onClick={() => setCurrentPage(prev => prev - 1)}
                >
                  ◀ Previous
                </button>
                <span className="pagination-pages">
                  Page {currentPage} of {totalPages}
                </span>
                <button
                  className="pagination-btn"
                  disabled={currentPage >= totalPages || loading}
                  onClick={() => setCurrentPage(prev => prev + 1)}
                >
                  Next ▶
                </button>
              </div>
            </div>

          </div>
          )}

        </div>

        {/* Right Side detail panel is now rendered at root overlay level using React Portal */}

      </div>


      {/* Modals */}
      {convertingCustomerLead && (
        <LeadConversionModal
          isOpen={true}
          lead={convertingCustomerLead}
          onClose={() => setConvertingCustomerLead(null)}
          onSuccess={() => {
            setConvertingCustomerLead(null);
            fetchLeads();
            fetchDbCounts();
          }}
        />
      )}
      {convertingLead && (
        <RFQBuilder
          lead={convertingLead}
          onClose={() => setConvertingLead(null)}
          onSuccess={async () => {
            setConvertingLead(null);
            fetchLeads();
            fetchDbCounts();
            if (selectedLead) {
              try {
                const fresh = await getLead(selectedLead.id);
                setSelectedLead(fresh?.data || fresh);
              } catch (err) {
                console.error('Failed to reload selected lead after RFQ conversion:', err);
              }
            }
          }}
        />
      )}
      {showAddModal && (
        <AddLeadModal
          users={users}
          onClose={() => setShowAddModal(false)}
          onLeadAdded={() => {
            fetchLeads();
            fetchDbCounts();
          }}
        />
      )}
      {showImportModal && (
        <ImportLeadsModal
          onClose={() => setShowImportModal(false)}
          onImportComplete={() => {
            fetchLeads();
            fetchDbCounts();
          }}
        />
      )}
      {/* Drawer Overlay (rendered at root level using React Portal for full viewport backdrop containment) */}
      {activeLeadForDrawer && createPortal(
        <>
          <div 
            className={`leads-drawer-backdrop ${isClosingDrawer ? 'closing' : ''}`}
            onClick={() => changeSelectedLead(null)}
          />
          <div className={`leads-details-sidebar ${isClosingDrawer ? 'closing' : ''}`}>
            <LeadDetailPanel
              lead={activeLeadForDrawer}
              initialTab={drawerInitialTab}
              users={users}
              onClose={() => changeSelectedLead(null)}
              onLeadUpdated={handleLeadUpdated}
              onDirtyChange={setDrawerIsDirty}
              onConvertToRFQ={(l) => {
                setConvertingLead(l);
              }}
              onConvertToCustomer={(l) => {
                setConvertingCustomerLead(l);
              }}
            />
          </div>
        </>,
        document.body
      )}

    </div>
  );
}

// ── Custom Dropdown Component ──
function CustomSelect({ name, value, onChange, options, disabled, placeholder }) {
  const [isOpen, setIsOpen] = useState(false);
  const containerRef = useRef(null);

  useEffect(() => {
    const handleOutsideClick = (e) => {
      if (containerRef.current && !containerRef.current.contains(e.target)) {
        setIsOpen(false);
      }
    };
    if (isOpen) document.addEventListener('mousedown', handleOutsideClick);
    return () => document.removeEventListener('mousedown', handleOutsideClick);
  }, [isOpen]);

  const selectedOption = options.find(o => o.value === value);

  return (
    <div className={`leads-custom-dropdown ${disabled ? 'disabled' : ''} ${isOpen ? 'open' : ''}`} ref={containerRef}>
      <div 
        className="leads-custom-dropdown-selected" 
        onClick={() => !disabled && setIsOpen(!isOpen)}
      >
        {selectedOption ? selectedOption.label : <span className="placeholder">{placeholder || 'Select...'}</span>}
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
          <polyline points="6 9 12 15 18 9"></polyline>
        </svg>
      </div>
      {isOpen && !disabled && (
        <ul className="leads-custom-dropdown-list">
          {options.map(opt => (
            <li
              key={opt.value}
              className={value === opt.value ? 'selected' : ''}
              onClick={(e) => {
                e.stopPropagation();
                onChange({ target: { name, value: opt.value } });
                setIsOpen(false);
              }}
            >
              {opt.label}
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}

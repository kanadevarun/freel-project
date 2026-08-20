import { useState, useEffect, useMemo, useCallback } from 'react';
import { listLeads } from '../../../services/leadsService';
import PageHeader from '../../../components/dashboard/PageHeader';
import StatusBadge from '../../../components/dashboard/StatusBadge';
import LeadDetailDrawer from './LeadDetailDrawer';
import AddLeadModal from './AddLeadModal';
import ImportLeadsModal from './ImportLeadsModal';
import './LeadsPage.css';

// ── Constants ─────────────────────────────────────────────────────────────────

/**
 * LEAD_STATUS — All possible status values for a Lead.
 * Simple meaning: Instead of typing the string "QUALIFIED" everywhere (and risking
 * a typo), we define it once here and use LEAD_STATUS.QUALIFIED throughout.
 * If the backend ever renames a status, we only change it in one place.
 */
export const LEAD_STATUS = {
  NEW:       'NEW',
  QUALIFIED: 'QUALIFIED',
  REJECTED:  'REJECTED',
  CONVERTED: 'CONVERTED',
};

/** All tab definitions — "all" sends no filter, others filter by status. */
const TABS = [
  { id: 'all',       label: 'All Leads',  statusFilter: null },
  { id: 'new',       label: 'New',        statusFilter: LEAD_STATUS.NEW },
  { id: 'qualified', label: 'Qualified',  statusFilter: LEAD_STATUS.QUALIFIED },
  { id: 'rejected',  label: 'Rejected',   statusFilter: LEAD_STATUS.REJECTED },
  { id: 'converted', label: 'Converted',  statusFilter: LEAD_STATUS.CONVERTED },
];

/**
 * STATUS_CFG — Maps a LEAD_STATUS value to the display label and badge color type.
 * Simple meaning: This tells the UI what text to show and what color to use
 * for each status. For example, QUALIFIED → green badge, REJECTED → red badge.
 */
const STATUS_CFG = {
  [LEAD_STATUS.NEW]:       { label: 'New',       type: 'info' },
  [LEAD_STATUS.QUALIFIED]: { label: 'Qualified', type: 'success' },
  [LEAD_STATUS.REJECTED]:  { label: 'Rejected',  type: 'danger' },
  [LEAD_STATUS.CONVERTED]: { label: 'Converted', type: 'primary' },
};

// ── Sub-components ────────────────────────────────────────────────────────────

/** Shows a color-coded AI score badge. */
function AIScoreBadge({ score }) {
  if (score == null) return <span className="ai-score-badge ai-score-none">Pending</span>;
  const cls = score >= 70 ? 'ai-score-high' : score >= 40 ? 'ai-score-medium' : 'ai-score-low';
  return <span className={`ai-score-badge ${cls}`}>⚡ {score}/100</span>;
}

/** Shows pulsing skeleton rows while loading. */
function SkeletonRows({ count = 6 }) {
  return Array.from({ length: count }).map((_, i) => (
    <tr key={i} className="leads-skeleton-row">
      {[130, 100, 80, 70, 60].map((w, j) => (
        <td key={j}><div className="leads-skeleton-cell" style={{ width: w }} /></td>
      ))}
    </tr>
  ));
}

// ── Main Component ─────────────────────────────────────────────────────────────

/**
 * LeadsPage — The main Leads module page.
 *
 * Simple meaning: This is the whole "Leads" screen. It shows:
 *   - 4 stat cards at the top (total, new, qualified, converted counts)
 *   - A row of tabs (All | New | Qualified | Rejected | Converted)
 *   - A search bar + action buttons (Add Lead, Import CSV)
 *   - The data table listing all leads
 *
 * Clicking a row opens the LeadDetailDrawer (slide-in panel).
 * Clicking "Add Lead" opens the AddLeadModal.
 * Clicking "Import CSV" opens the ImportLeadsModal.
 */
export default function LeadsPage() {
  // ── State ──────────────────────────────────────────────────────────────────
  const [allLeads, setAllLeads]           = useState([]);   // All fetched leads from API
  const [loading, setLoading]             = useState(true); // API loading state
  const [activeTab, setActiveTab]         = useState('all');
  const [searchQuery, setSearchQuery]     = useState('');
  const [selectedLead, setSelectedLead]   = useState(null); // Lead shown in drawer
  const [showAddModal, setShowAddModal]   = useState(false);
  const [showImportModal, setShowImportModal] = useState(false);

  // ── Data Fetching ─────────────────────────────────────────────────────────
  const fetchLeads = useCallback(async () => {
    setLoading(true);
    try {
      const data = await listLeads({ limit: 100 });
      // The backend returns { leads: [...], total: N }
      setAllLeads(data?.leads || []);
    } catch (err) {
      console.error('Failed to load leads:', err);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchLeads();
  }, [fetchLeads]);

  // ── Derived Data ─────────────────────────────────────────────────────────
  /** Compute status counts for the stat cards and tab badges */
  const counts = useMemo(() => ({
    all:       allLeads.length,
    new:       allLeads.filter(l => l.status === LEAD_STATUS.NEW).length,
    qualified: allLeads.filter(l => l.status === LEAD_STATUS.QUALIFIED).length,
    rejected:  allLeads.filter(l => l.status === LEAD_STATUS.REJECTED).length,
    converted: allLeads.filter(l => l.status === LEAD_STATUS.CONVERTED).length,
  }), [allLeads]);

  /** Apply active tab filter + search query */
  const visibleLeads = useMemo(() => {
    const tab = TABS.find(t => t.id === activeTab);
    let filtered = tab?.statusFilter
      ? allLeads.filter(l => l.status === tab.statusFilter)
      : allLeads;

    if (searchQuery.trim()) {
      const q = searchQuery.toLowerCase();
      filtered = filtered.filter(l =>
        l.company_name?.toLowerCase().includes(q) ||
        l.contact_name?.toLowerCase().includes(q) ||
        l.email?.toLowerCase().includes(q)
      );
    }
    return filtered;
  }, [allLeads, activeTab, searchQuery]);

  // ── Render ────────────────────────────────────────────────────────────────
  return (
    <div className="leads-page">

      {/* Page Header */}
      <PageHeader
        title="Leads"
        subtitle="Discover, score, and convert high-value shipping prospects"
      />

      {/* ─── Stat Cards ─── */}
      <div className="leads-stats-row">
        <div className="leads-stat-card">
          <div className="leads-stat-icon teal">🎯</div>
          <div>
            <div className="leads-stat-value">{counts.all}</div>
            <div className="leads-stat-label">Total Leads</div>
          </div>
        </div>
        <div className="leads-stat-card">
          <div className="leads-stat-icon indigo">✨</div>
          <div>
            <div className="leads-stat-value">{counts.new}</div>
            <div className="leads-stat-label">New</div>
          </div>
        </div>
        <div className="leads-stat-card">
          <div className="leads-stat-icon green">✅</div>
          <div>
            <div className="leads-stat-value">{counts.qualified}</div>
            <div className="leads-stat-label">Qualified</div>
          </div>
        </div>
        <div className="leads-stat-card">
          <div className="leads-stat-icon amber">📋</div>
          <div>
            <div className="leads-stat-value">{counts.converted}</div>
            <div className="leads-stat-label">Converted to RFQ</div>
          </div>
        </div>
      </div>

      {/* ─── Tabs ─── */}
      <div className="leads-tabs">
        {TABS.map(tab => (
          <button
            key={tab.id}
            className={`leads-tab ${activeTab === tab.id ? 'active' : ''}`}
            onClick={() => setActiveTab(tab.id)}
          >
            {tab.label}
            <span className="leads-tab-count">{counts[tab.id]}</span>
          </button>
        ))}
      </div>

      {/* ─── Action Bar ─── */}
      <div className="leads-action-bar">
        <div className="leads-search-wrap">
          <span className="leads-search-icon">🔍</span>
          <input
            type="text"
            className="leads-search-input"
            placeholder="Search by company, contact, email..."
            value={searchQuery}
            onChange={e => setSearchQuery(e.target.value)}
          />
        </div>
        <div className="leads-bar-spacer" />
        <button className="leads-btn leads-btn-ghost" onClick={() => setShowImportModal(true)}>
          📥 Import CSV
        </button>
        <button className="leads-btn leads-btn-primary" onClick={() => setShowAddModal(true)}>
          + Add Lead
        </button>
      </div>

      {/* ─── Table or Enterprise Empty State ─── */}
      {loading ? (
        <div className="leads-table-wrap">
          <table className="leads-table">
            <thead>
              <tr>
                <th>Company</th>
                <th>Status</th>
                <th>AI Score</th>
                <th>Source</th>
                <th>Added</th>
              </tr>
            </thead>
            <tbody>
              <SkeletonRows count={7} />
            </tbody>
          </table>
        </div>
      ) : allLeads.length === 0 ? (
        <div className="leads-empty-enterprise" style={{
          background: '#FFFFFF',
          border: '1px solid #E2E8F0',
          borderRadius: '16px',
          padding: '48px 32px 32px 32px',
          textAlign: 'center',
          boxShadow: '0 1px 3px rgba(15, 23, 42, 0.03)',
        }}>
          <div style={{
            width: '56px',
            height: '56px',
            borderRadius: '14px',
            background: '#EFF6FF',
            border: '1px solid #DBEAFE',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            fontSize: '1.6rem',
            margin: '0 auto 16px auto',
          }}>
            🎯
          </div>
          <h3 style={{ fontSize: '1.2rem', fontWeight: 800, color: '#0F172A', marginBottom: '8px' }}>
            No leads yet
          </h3>
          <p style={{ fontSize: '0.85rem', color: '#64748B', maxWidth: '440px', margin: '0 auto 24px auto', lineHeight: 1.5 }}>
            Start building your customer pipeline. Add your first prospect manually or import your existing customer list.
          </p>

          <div style={{ display: 'flex', gap: '12px', justifyContent: 'center', flexWrap: 'wrap', marginBottom: '32px' }}>
            <button
              className="leads-btn leads-btn-primary"
              onClick={() => setShowAddModal(true)}
              style={{ padding: '10px 22px', fontSize: '0.82rem', fontWeight: 700 }}
            >
              + Add Your First Lead
            </button>
            <button
              className="leads-btn leads-btn-ghost"
              onClick={() => setShowImportModal(true)}
              style={{ padding: '10px 20px', fontSize: '0.82rem', fontWeight: 600 }}
            >
              📥 Import from CSV
            </button>
          </div>

          {/* AI Lead Qualification Feature Banner */}
          <div style={{
            borderTop: '1px solid #F1F5F9',
            paddingTop: '20px',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            gap: '10px',
            color: '#475569',
            fontSize: '0.78rem',
          }}>
            <span style={{
              background: '#EEF2FF',
              color: '#4F46E5',
              fontWeight: 800,
              padding: '2px 8px',
              borderRadius: '4px',
              fontSize: '0.68rem',
            }}>
              AI FEATURE
            </span>
            <span><strong>AI Lead Qualification:</strong> LogisticsHQ automatically scores and prioritizes new leads based on shipment volume and trade lanes.</span>
          </div>
        </div>
      ) : (
        <div className="leads-table-wrap">
          <table className="leads-table">
            <thead>
              <tr>
                <th>Company</th>
                <th>Status</th>
                <th>AI Score</th>
                <th>Source</th>
                <th>Added</th>
              </tr>
            </thead>
            <tbody>
              {visibleLeads.length === 0 ? (
                <tr>
                  <td colSpan={5}>
                    <div className="leads-empty">
                      <div className="leads-empty-icon">🔍</div>
                      <div className="leads-empty-title">No leads match your search</div>
                      <div className="leads-empty-sub">Try a different search term</div>
                    </div>
                  </td>
                </tr>
              ) : (
                visibleLeads.map(lead => {
                  const statusCfg = STATUS_CFG[lead.status] || { label: lead.status, type: 'neutral' };
                  const avatar = lead.company_name?.slice(0, 2).toUpperCase() || '??';

                  return (
                    <tr key={lead.id} onClick={() => setSelectedLead(lead)}>
                      <td>
                        <div className="lead-company-cell">
                          <div className="lead-avatar">{avatar}</div>
                          <div>
                            <div className="lead-company-name">{lead.company_name}</div>
                            <div className="lead-contact-name">{lead.contact_name || lead.email || '—'}</div>
                          </div>
                        </div>
                      </td>
                      <td>
                        <StatusBadge
                          status={lead.status || 'NEW'}
                          customLabel={statusCfg.label}
                          customType={statusCfg.type}
                        />
                      </td>
                      <td><AIScoreBadge score={lead.ai_score} /></td>
                      <td style={{ color: '#475569' }}>{lead.source || '—'}</td>
                      <td style={{ color: '#94A3B8', fontSize: 12 }}>
                        {lead.created_at ? new Date(lead.created_at).toLocaleDateString() : '—'}
                      </td>
                    </tr>
                  );
                })
              )}
            </tbody>
          </table>
        </div>
      )}

      {/* ─── Modals & Drawer ─── */}
      {selectedLead && (
        <LeadDetailDrawer
          lead={selectedLead}
          onClose={() => setSelectedLead(null)}
          onLeadUpdated={fetchLeads}
        />
      )}
      {showAddModal && (
        <AddLeadModal
          onClose={() => setShowAddModal(false)}
          onLeadAdded={fetchLeads}
        />
      )}
      {showImportModal && (
        <ImportLeadsModal
          onClose={() => setShowImportModal(false)}
          onImportComplete={fetchLeads}
        />
      )}

    </div>
  );
}

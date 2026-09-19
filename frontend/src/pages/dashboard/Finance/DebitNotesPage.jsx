import React, { useState, useEffect, useCallback } from 'react';
import {
  Plus, Search, RefreshCw, FileWarning,
  ChevronLeft, ChevronRight, ChevronsLeft, ChevronsRight,
  AlertTriangle, CheckCircle2, Ban, Clock, DollarSign,
  TrendingUp,
} from 'lucide-react';
import api from '../../../services/api';
import CreateDebitNoteModal from './components/CreateDebitNoteModal';
import DebitNoteDetailsPanel from './components/DebitNoteDetailsPanel';
import './DebitNotesPage.css';

const STATUS_FILTERS = ['All', 'DRAFT', 'ISSUED', 'ACKNOWLEDGED', 'VOID'];

const STATUS_CONFIG = {
  DRAFT:        { label: 'Draft',        color: '#475569', bg: '#f1f5f9', border: '#e2e8f0', Icon: Clock },
  ISSUED:       { label: 'Issued',       color: '#1d4ed8', bg: '#eff6ff', border: '#bfdbfe', Icon: AlertTriangle },
  ACKNOWLEDGED: { label: 'Acknowledged', color: '#15803d', bg: '#f0fdf4', border: '#bbf7d0', Icon: CheckCircle2 },
  VOID:         { label: 'Void',         color: '#b91c1c', bg: '#fef2f2', border: '#fecaca', Icon: Ban },
};

function StatusPill({ status }) {
  const cfg = STATUS_CONFIG[status] || STATUS_CONFIG.DRAFT;
  const Icon = cfg.Icon;
  return (
    <span className="dn-status-pill" style={{ color: cfg.color, background: cfg.bg, borderColor: cfg.border }}>
      <Icon size={11} />
      {cfg.label}
    </span>
  );
}

function KpiCard({ title, amount, count, Icon, accent }) {
  return (
    <div className="dn-kpi-card" style={{ '--accent': accent }}>
      <div className="dn-kpi-icon"><Icon size={18} /></div>
      <div className="dn-kpi-body">
        <div className="dn-kpi-title">{title}</div>
        <div className="dn-kpi-amount">{amount || '$0.00'}</div>
        <div className="dn-kpi-count">{count} debit notes</div>
      </div>
    </div>
  );
}

export default function DebitNotesPage() {
  const [debitNotes, setDebitNotes] = useState([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [kpi, setKpi] = useState(null);

  const [statusFilter, setStatusFilter] = useState('All');
  const [searchQuery, setSearchQuery] = useState('');
  const [currentPage, setCurrentPage] = useState(1);
  const [pageSize] = useState(15);

  const [selectedDN, setSelectedDN] = useState(null);
  const [isCreateOpen, setIsCreateOpen] = useState(false);

  const [toast, setToast] = useState(null);

  const showToast = (text, type = 'info') => {
    setToast({ text, type });
    setTimeout(() => setToast(null), 4000);
  };

  const fetchKpi = useCallback(async () => {
    try {
      const res = await api.get('/api/v1/debit-notes/kpi-stats');
      const stats = res?.total_issued ? res : (res?.data?.total_issued ? res.data : (res?.data?.data || res?.data || res));
      setKpi(stats);
    } catch (_) {}
  }, []);

  const fetchDebitNotes = useCallback(async () => {
    try {
      setLoading(true);
      setError(null);
      const params = new URLSearchParams();
      params.set('page', String(currentPage));
      params.set('page_size', String(pageSize));
      if (statusFilter !== 'All') params.set('status', statusFilter);
      if (searchQuery.trim()) params.set('search', searchQuery.trim());

      const res = await api.get(`/api/v1/debit-notes?${params.toString()}`);
      const rawNotes = res?.debit_notes || res?.data?.debit_notes || res?.data?.data?.debit_notes || (Array.isArray(res) ? res : []);
      const totalCount = res?.total ?? res?.data?.total ?? res?.data?.data?.total ?? rawNotes.length;
      setDebitNotes(rawNotes);
      setTotal(totalCount);
    } catch (err) {
      setError('Failed to load debit notes.');
      setDebitNotes([]);
    } finally {
      setLoading(false);
    }
  }, [currentPage, pageSize, statusFilter, searchQuery]);

  useEffect(() => {
    fetchDebitNotes();
    fetchKpi();
  }, [fetchDebitNotes, fetchKpi]);

  // Escape closes panel
  useEffect(() => {
    const onKey = (e) => { if (e.key === 'Escape' && selectedDN) setSelectedDN(null); };
    window.addEventListener('keydown', onKey);
    return () => window.removeEventListener('keydown', onKey);
  }, [selectedDN]);

  const handleStatusFilterChange = (s) => {
    setStatusFilter(s);
    setCurrentPage(1);
  };

  const handleSearch = (e) => {
    setSearchQuery(e.target.value);
    setCurrentPage(1);
  };

  const handleIssue = async (dn) => {
    try {
      await api.post(`/api/v1/debit-notes/${dn.id}/issue`);
      showToast(`Debit Note ${dn.debit_note_number} issued!`, 'success');
      setSelectedDN(null);
      fetchDebitNotes();
      fetchKpi();
    } catch (err) {
      showToast(err?.message || 'Failed to issue debit note', 'error');
    }
  };

  const handleVoid = async (dn) => {
    if (!window.confirm(`Void debit note ${dn.debit_note_number}? This cannot be undone.`)) return;
    try {
      await api.post(`/api/v1/debit-notes/${dn.id}/void`);
      showToast(`Debit Note ${dn.debit_note_number} voided.`, 'info');
      setSelectedDN(null);
      fetchDebitNotes();
      fetchKpi();
    } catch (err) {
      showToast(err?.message || 'Failed to void debit note', 'error');
    }
  };

  const handleRowClick = async (dn) => {
    // Fetch full detail (includes line items)
    try {
      const res = await api.get(`/api/v1/debit-notes/${dn.id}`);
      setSelectedDN(res?.data || res);
    } catch (_) {
      setSelectedDN(dn);
    }
  };

  const totalPages = Math.max(1, Math.ceil(total / pageSize));

  const formatAmount = (val, currency = 'USD') =>
    `${currency} ${Number(val || 0).toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`;

  const formatDate = (d) => {
    if (!d) return '—';
    return new Date(d).toLocaleDateString('en-GB', { day: '2-digit', month: 'short', year: 'numeric' });
  };

  return (
    <div className="dn-page">
      {/* Toast */}
      {toast && (
        <div className={`dn-toast dn-toast--${toast.type}`}>
          {toast.text}
        </div>
      )}

      {/* Page header */}
      <div className="dn-page-header">
        <div className="dn-page-title-block">
          <div className="dn-page-icon"><FileWarning size={22} /></div>
          <div>
            <h1 className="dn-page-title">Debit Notes</h1>
            <p className="dn-page-subtitle">Issue additional charges to customers</p>
          </div>
        </div>
        <div className="dn-page-actions">
          <button className="dn-btn-refresh" onClick={() => { fetchDebitNotes(); fetchKpi(); }} title="Refresh">
            <RefreshCw size={15} />
          </button>
          <button className="dn-btn-create" onClick={() => setIsCreateOpen(true)} id="create-debit-note-btn">
            <Plus size={16} />
            New Debit Note
          </button>
        </div>
      </div>

      {/* KPI Cards */}
      <div className="dn-kpi-row">
        <KpiCard
          title="Total Issued"
          amount={kpi?.total_issued?.display_amount}
          count={kpi?.total_issued?.count || 0}
          Icon={TrendingUp}
          accent="#2563eb"
        />
        <KpiCard
          title="Draft"
          amount={kpi?.draft?.display_amount}
          count={kpi?.draft?.count || 0}
          Icon={Clock}
          accent="#64748b"
        />
        <KpiCard
          title="Outstanding"
          amount={kpi?.outstanding?.display_amount}
          count={kpi?.outstanding?.count || 0}
          Icon={DollarSign}
          accent="#d97706"
        />
        <KpiCard
          title="Void"
          amount={kpi?.void?.display_amount}
          count={kpi?.void?.count || 0}
          Icon={Ban}
          accent="#dc2626"
        />
      </div>

      {/* Filters & search */}
      <div className="dn-controls">
        <div className="dn-status-filters">
          {STATUS_FILTERS.map((s) => (
            <button
              key={s}
              className={`dn-filter-pill ${statusFilter === s ? 'active' : ''}`}
              onClick={() => handleStatusFilterChange(s)}
            >
              {s === 'All' ? 'All' : (STATUS_CONFIG[s]?.label || s)}
            </button>
          ))}
        </div>
        <div className="dn-search-box">
          <Search size={15} className="dn-search-icon" />
          <input
            type="text"
            placeholder="Search by DN#, customer, invoice, reason…"
            value={searchQuery}
            onChange={handleSearch}
            id="debit-notes-search"
          />
        </div>
      </div>

      {/* Main content layout */}
      <div className={`dn-content-layout ${selectedDN ? 'has-panel' : ''}`}>
        {/* Table */}
        <div className="dn-table-wrap">
          {loading ? (
            <div className="dn-loading">
              <div className="dn-spinner" />
              <span>Loading debit notes…</span>
            </div>
          ) : error ? (
            <div className="dn-error-state">
              <AlertTriangle size={32} />
              <p>{error}</p>
              <button onClick={fetchDebitNotes} className="dn-btn-refresh">Retry</button>
            </div>
          ) : debitNotes.length === 0 ? (
            <div className="dn-empty-state">
              <FileWarning size={48} className="dn-empty-icon" />
              <h3>No debit notes found</h3>
              <p>Create your first debit note to charge additional amounts to customers.</p>
              <button className="dn-btn-create" onClick={() => setIsCreateOpen(true)}>
                <Plus size={15} /> New Debit Note
              </button>
            </div>
          ) : (
            <>
              <table className="dn-table">
                <thead>
                  <tr>
                    <th>DN Number</th>
                    <th>Customer</th>
                    <th>Linked Invoice</th>
                    <th>Amount</th>
                    <th>Status</th>
                    <th>Issue Date</th>
                    <th>Due Date</th>
                    <th>Actions</th>
                  </tr>
                </thead>
                <tbody>
                  {debitNotes.map((dn) => (
                    <tr
                      key={dn.id}
                      className={`dn-table-row ${selectedDN?.id === dn.id ? 'selected' : ''}`}
                      onClick={() => handleRowClick(dn)}
                    >
                      <td>
                        <span className="dn-dn-number">{dn.debit_note_number}</span>
                      </td>
                      <td>
                        <span className="dn-customer-name">{dn.customer_name}</span>
                        {dn.customer_country && <span className="dn-customer-country">{dn.customer_country}</span>}
                      </td>
                      <td>
                        {dn.invoice_number ? (
                          <span className="dn-invoice-link">{dn.invoice_number}</span>
                        ) : (
                          <span className="dn-none">—</span>
                        )}
                      </td>
                      <td>
                        <span className="dn-amount">{formatAmount(dn.total_amount, dn.currency)}</span>
                      </td>
                      <td>
                        <StatusPill status={dn.status} />
                      </td>
                      <td className="dn-date-cell">{formatDate(dn.issue_date)}</td>
                      <td className="dn-date-cell">{formatDate(dn.due_date)}</td>
                      <td onClick={(e) => e.stopPropagation()}>
                        <div className="dn-row-actions">
                          {dn.status === 'DRAFT' && (
                            <button
                              className="dn-row-btn dn-row-btn-issue"
                              title="Issue"
                              onClick={() => handleIssue(dn)}
                            >
                              Issue
                            </button>
                          )}
                          {dn.status !== 'VOID' && (
                            <button
                              className="dn-row-btn dn-row-btn-void"
                              title="Void"
                              onClick={() => handleVoid(dn)}
                            >
                              Void
                            </button>
                          )}
                        </div>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>

              {/* Pagination */}
              <div className="dn-pagination">
                <div className="dn-pagination-info">
                  Showing {Math.min((currentPage - 1) * pageSize + 1, total)}–{Math.min(currentPage * pageSize, total)} of {total}
                </div>
                <div className="dn-pagination-controls">
                  <button className="dn-page-btn" disabled={currentPage === 1} onClick={() => setCurrentPage(1)} title="First">
                    <ChevronsLeft size={15} />
                  </button>
                  <button className="dn-page-btn" disabled={currentPage === 1} onClick={() => setCurrentPage(p => p - 1)} title="Prev">
                    <ChevronLeft size={15} />
                  </button>
                  {Array.from({ length: Math.min(totalPages, 5) }, (_, i) => {
                    const p = Math.max(1, Math.min(totalPages - 4, currentPage - 2)) + i;
                    return (
                      <button
                        key={p}
                        className={`dn-page-num ${currentPage === p ? 'active' : ''}`}
                        onClick={() => setCurrentPage(p)}
                      >
                        {p}
                      </button>
                    );
                  })}
                  <button className="dn-page-btn" disabled={currentPage === totalPages} onClick={() => setCurrentPage(p => p + 1)} title="Next">
                    <ChevronRight size={15} />
                  </button>
                  <button className="dn-page-btn" disabled={currentPage === totalPages} onClick={() => setCurrentPage(totalPages)} title="Last">
                    <ChevronsRight size={15} />
                  </button>
                </div>
              </div>
            </>
          )}
        </div>

        {/* Detail Panel */}
        {selectedDN && (
          <div className="dn-panel-wrap">
            <DebitNoteDetailsPanel
              debitNote={selectedDN}
              onClose={() => setSelectedDN(null)}
              onIssue={handleIssue}
              onVoid={handleVoid}
            />
          </div>
        )}
      </div>

      {/* Drawer overlay for smaller screens */}
      {selectedDN && (
        <div className="dn-drawer-overlay-mobile" onClick={() => setSelectedDN(null)} />
      )}

      {/* Create Modal */}
      <CreateDebitNoteModal
        isOpen={isCreateOpen}
        onClose={() => setIsCreateOpen(false)}
        showToast={showToast}
        onSuccess={() => {
          fetchDebitNotes();
          fetchKpi();
        }}
      />
    </div>
  );
}

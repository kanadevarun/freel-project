import React, { useState, useEffect, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  Plus,
  Filter,
  Download,
  Search,
  ChevronLeft,
  ChevronRight,
  ChevronsLeft,
  ChevronsRight,
  FileText,
  AlertCircle
} from 'lucide-react';

import api from '../../../services/api';
import InvoiceKpiCards from './components/InvoiceKpiCards';
import InvoiceTable from './components/InvoiceTable';
import InvoiceDetailsPanel from './components/InvoiceDetailsPanel';
import InvoiceFilterDrawer from './components/InvoiceFilterDrawer';
import CreateInvoiceModal from './components/CreateInvoiceModal';
import RecordPaymentModal from './components/RecordPaymentModal';

import { INITIAL_KPI_STATS, INVOICE_STATUSES } from './mockInvoiceData';
import './InvoicesPage.css';

export default function InvoicesPage() {
  const navigate = useNavigate();

  // State management
  const [invoices, setInvoices] = useState([]);
  const [totalResults, setTotalResults] = useState(0);
  const [kpiStats, setKpiStats] = useState(INITIAL_KPI_STATS);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  // Tabs & Filter states
  const [primaryTab, setPrimaryTab] = useState('ALL'); // 'ALL' | 'MY'
  const [statusFilter, setStatusFilter] = useState('All');
  const [searchQuery, setSearchQuery] = useState('');
  const [isFilterDrawerOpen, setIsFilterDrawerOpen] = useState(false);
  const [advancedFilters, setAdvancedFilters] = useState({
    customer: '',
    shipmentId: '',
    status: 'All',
    dateFrom: '',
    dateTo: '',
    minAmount: '',
    maxAmount: '',
    currency: ''
  });

  // Table & Panel selection
  const [selectedInvoice, setSelectedInvoice] = useState(null);

  // Pagination state
  const [currentPage, setCurrentPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);

  // Modal states
  const [isCreateModalOpen, setIsCreateModalOpen] = useState(false);
  const [editingInvoice, setEditingInvoice] = useState(null);
  const [isRecordPaymentOpen, setIsRecordPaymentOpen] = useState(false);
  const [recordingPaymentInvoice, setRecordingPaymentInvoice] = useState(null);

  // Keyboard navigation: Escape key closes invoice details drawer
  useEffect(() => {
    function handleKeyDown(e) {
      if (e.key === 'Escape' && selectedInvoice) {
        setSelectedInvoice(null);
      }
    }
    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [selectedInvoice]);

  // Action toast state
  const [toastMessage, setToastMessage] = useState(null);

  const showToast = (msg, type = 'info') => {
    setToastMessage({ text: msg, type });
    setTimeout(() => {
      setToastMessage(null);
    }, 4000);
  };

  // Fetch KPI statistics from backend
  const fetchKpiStats = useCallback(async () => {
    try {
      const res = await api.get('/api/v1/invoices/kpi-stats');
      if (res) {
        setKpiStats({
          totalInvoices: {
            amount: res.total_invoices?.amount || '$0.00',
            displayAmount: res.total_invoices?.display_amount || '$0.00',
            count: res.total_invoices?.count || 0,
            label: res.total_invoices?.label || '0 Invoices',
            trend: res.total_invoices?.trend || '18.6%',
            trendDirection: 'up',
            trendPeriod: 'vs last 7 days'
          },
          outstanding: {
            amount: res.outstanding?.amount || '$0.00',
            displayAmount: res.outstanding?.display_amount || '$0.00',
            count: res.outstanding?.count || 0,
            label: res.outstanding?.label || '0 Invoices',
            trend: res.outstanding?.trend || '12.4%',
            trendDirection: 'up',
            trendPeriod: 'vs last 7 days'
          },
          paidThisMonth: {
            amount: res.paid_this_month?.amount || '$0.00',
            displayAmount: res.paid_this_month?.display_amount || '$0.00',
            count: res.paid_this_month?.count || 0,
            label: res.paid_this_month?.label || '0 Invoices',
            trend: res.paid_this_month?.trend || '24.8%',
            trendDirection: 'up',
            trendPeriod: 'vs last 7 days'
          },
          overdue: {
            amount: res.overdue?.amount || '$0.00',
            displayAmount: res.overdue?.display_amount || '$0.00',
            count: res.overdue?.count || 0,
            label: res.overdue?.label || '0 Invoices',
            trend: res.overdue?.trend || '8.2%',
            trendDirection: 'up',
            trendPeriod: 'vs last 7 days'
          }
        });
      }
    } catch (err) {
      console.warn('Unable to load invoice KPI stats from backend:', err);
    }
  }, []);

  // Fetch real persistent invoices from backend API
  const fetchInvoices = useCallback(async () => {
    try {
      setLoading(true);
      setError(null);

      const queryParams = new URLSearchParams();
      queryParams.set('primary_tab', primaryTab);
      if (statusFilter !== 'All') queryParams.set('status', statusFilter);
      if (searchQuery.trim()) queryParams.set('search', searchQuery.trim());
      queryParams.set('page', currentPage);
      queryParams.set('page_size', pageSize);

      const res = await api.get(`/api/v1/invoices?${queryParams.toString()}`);
      
      const invList = res?.invoices || (Array.isArray(res) ? res : []);
      const total = res?.total !== undefined ? res.total : invList.length;

      // Transform backend records for UI consistency
      const formatted = invList.map((inv) => ({
        id: inv.id,
        invoiceNumber: inv.invoice_number || `INV-${inv.id}`,
        creator: inv.creator_name || 'By Varun Sharma',
        customer: inv.customer_name || '—',
        customerCountry: inv.customer_country || 'USA',
        shipmentId: inv.shipment_number || (inv.shipment_id ? `SH-2026-${inv.shipment_id}` : '—'),
        route: inv.route || 'Shanghai ➔ Los Angeles',
        invoiceDate: inv.invoice_date ? new Date(inv.invoice_date).toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' }) : 'Aug 15, 2026',
        dueDate: inv.due_date ? new Date(inv.due_date).toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' }) : 'Aug 30, 2026',
        daysLeft: inv.days_left || '15 days left',
        amount: Number(inv.total_amount || 0),
        currency: inv.currency || 'USD',
        status: inv.status || 'Draft',
        balance: Number(inv.balance_due || 0),
        type: inv.type || 'CUSTOMER_AR',
        bookmarked: Boolean(inv.bookmarked),
        isMyInvoice: Boolean(inv.is_my_invoice),
        subtotal: Number(inv.subtotal || inv.total_amount || 0),
        tax: Number(inv.tax_amount || 0),
        discount: Number(inv.discount_amount || 0),
        amountPaid: Number(inv.paid_amount || 0),
        lineItems: inv.line_items || [],
        payments: inv.payments || [],
        documents: inv.documents || [],
        history: inv.history || []
      }));

      setInvoices(formatted);
      setTotalResults(total);

      // If an invoice is currently open in the drawer, refresh its latest detail
      if (selectedInvoice) {
        const stillExists = formatted.find(i => i.id === selectedInvoice.id);
        if (stillExists) {
          fetchInvoiceDetails(selectedInvoice.id);
        }
      }
    } catch (err) {
      console.error('Failed to load invoices from backend:', err);
      setError('Unable to load persistent invoice ledger. Please try again.');
    } finally {
      setLoading(false);
    }
  }, [primaryTab, statusFilter, searchQuery, currentPage, pageSize]);

  // Fetch single invoice full detail
  const fetchInvoiceDetails = async (invId) => {
    try {
      const res = await api.get(`/api/v1/invoices/${invId}`);
      if (res) {
        const fullDetail = {
          id: res.id,
          invoiceNumber: res.invoice_number,
          creator: res.creator_name || 'By Varun Sharma',
          customer: res.customer_name,
          customerCountry: res.customer_country || 'USA',
          shipmentId: res.shipment_number || (res.shipment_id ? `SH-2026-${res.shipment_id}` : '—'),
          route: res.route || 'Shanghai ➔ Los Angeles',
          invoiceDate: res.invoice_date ? new Date(res.invoice_date).toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' }) : 'Aug 15, 2026',
          dueDate: res.due_date ? new Date(res.due_date).toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' }) : 'Aug 30, 2026',
          daysLeft: res.days_left || '15 days left',
          amount: Number(res.total_amount || 0),
          currency: res.currency || 'USD',
          status: res.status || 'Draft',
          balance: Number(res.balance_due || 0),
          type: res.type || 'CUSTOMER_AR',
          bookmarked: Boolean(res.bookmarked),
          isMyInvoice: Boolean(res.is_my_invoice),
          subtotal: Number(res.subtotal || 0),
          tax: Number(res.tax_amount || 0),
          discount: Number(res.discount_amount || 0),
          amountPaid: Number(res.paid_amount || 0),
          lineItems: (res.line_items || []).map((item, idx) => ({
            id: item.id || idx + 1,
            description: item.description,
            quantity: Number(item.quantity || 1),
            rate: Number(item.unit_price || 0),
            amount: Number(item.total_amount || 0)
          })),
          payments: (res.payments || []).map((p, idx) => ({
            id: p.id || idx + 1,
            reference: p.payment_ref,
            amount: Number(p.amount || 0),
            method: p.payment_method,
            status: p.status,
            date: p.payment_date ? new Date(p.payment_date).toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' }) : 'Aug 20, 2026'
          })),
          documents: (res.documents || []).map((doc, idx) => ({
            id: doc.id || idx + 1,
            name: doc.document_name,
            size: doc.file_size,
            type: doc.file_type,
            uploadedAt: doc.uploaded_at ? new Date(doc.uploaded_at).toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' }) : 'Aug 15, 2026'
          })),
          history: (res.history || []).map((h, idx) => ({
            id: h.id || idx + 1,
            title: h.title,
            description: h.description,
            timestamp: h.created_at ? new Date(h.created_at).toLocaleString('en-US', { month: 'short', day: 'numeric', year: 'numeric', hour: '2-digit', minute: '2-digit' }) : 'Aug 15, 2026 10:30 AM',
            user: h.user_name
          }))
        };

        setSelectedInvoice(fullDetail);
      }
    } catch (err) {
      console.warn('Failed to load invoice details:', err);
    }
  };

  useEffect(() => {
    fetchInvoices();
    fetchKpiStats();
  }, [fetchInvoices, fetchKpiStats]);

  const totalPages = Math.ceil(totalResults / pageSize) || 1;

  const handleToggleBookmark = async (id) => {
    try {
      const res = await api.post(`/api/v1/invoices/${id}/bookmark`);
      const isBookmarked = res?.bookmarked;
      setInvoices((prev) =>
        prev.map((inv) =>
          inv.id === id ? { ...inv, bookmarked: isBookmarked } : inv
        )
      );
      if (selectedInvoice && selectedInvoice.id === id) {
        setSelectedInvoice((prev) => ({ ...prev, bookmarked: isBookmarked }));
      }
    } catch (err) {
      showToast('Failed to toggle bookmark', 'warning');
    }
  };

  // Action dispatcher
  const handleActionClick = async (actionName, payload) => {
    switch (actionName) {
      case 'newInvoice':
        setEditingInvoice(null);
        setIsCreateModalOpen(true);
        break;
      case 'issueInvoice':
        try {
          await api.post(`/api/v1/invoices/${payload.id}/issue`);
          showToast(`Invoice ${payload?.invoiceNumber || ''} issued successfully`, 'success');
          fetchInvoices();
          fetchKpiStats();
          if (payload?.id) fetchInvoiceDetails(payload.id);
        } catch (err) {
          showToast(err.message || 'Failed to issue invoice', 'warning');
        }
        break;
      case 'submitApproval':
        try {
          await api.post(`/api/v1/invoices/${payload.id}/submit-approval`);
          showToast(`Invoice ${payload?.invoiceNumber || ''} submitted for approval`, 'info');
          fetchInvoices();
          fetchKpiStats();
          if (payload?.id) fetchInvoiceDetails(payload.id);
        } catch (err) {
          showToast(err.message || 'Failed to submit invoice for approval', 'warning');
        }
        break;
      case 'editInvoice':
        setEditingInvoice(payload);
        setIsCreateModalOpen(true);
        break;
      case 'export':
        showToast('Exporting invoices to CSV/PDF...', 'info');
        break;
      case 'recordPayment':
        setRecordingPaymentInvoice(payload);
        setIsRecordPaymentOpen(true);
        break;
      case 'sendInvoice':
        showToast(`Send invoice ${payload?.invoiceNumber} to ${payload?.customer}`, 'info');
        break;
      case 'downloadPdf':
        showToast(`Downloading PDF for ${payload?.invoiceNumber}...`, 'info');
        break;
      case 'cancelInvoice':
        try {
          await api.post(`/api/v1/invoices/${payload.id}/cancel`, { reason: 'User requested cancellation' });
          showToast(`Cancelled invoice ${payload?.invoiceNumber}`, 'warning');
          fetchInvoices();
          fetchKpiStats();
          if (payload?.id) fetchInvoiceDetails(payload.id);
        } catch (err) {
          showToast(err.message || 'Failed to cancel invoice', 'warning');
        }
        break;
      case 'viewDoc':
        showToast(`Opening document ${payload?.name}...`, 'info');
        break;
      case 'downloadDoc':
        showToast(`Downloading document ${payload?.name}...`, 'info');
        break;
      case 'uploadDoc':
        showToast(`Upload document entry point for ${payload?.invoiceNumber}`, 'info');
        break;
      case 'navCustomer':
        navigate('/dashboard/customers');
        break;
      case 'navShipment':
        navigate('/dashboard/shipments');
        break;
      case 'navBooking':
        navigate('/dashboard/bookings');
        break;
      case 'navQuotation':
        navigate('/dashboard/quotations');
        break;
      case 'navApproval':
        navigate('/dashboard/approvals');
        break;
      case 'navDocument':
        navigate('/dashboard/documents');
        break;
      default:
        break;
    }
  };

  return (
    <div className="invoices-module-container">
      {/* Toast Notification Banner */}
      {toastMessage && (
        <div className={`invoices-toast banner-${toastMessage.type}`}>
          <AlertCircle size={16} />
          <span>{toastMessage.text}</span>
        </div>
      )}

      {/* Main Container Layout */}
      <div className="invoices-main-layout">
        {/* Left/Center Main Column */}
        <div className="invoices-primary-content">
          
          {/* ── Page Header ── */}
          <div className="invoices-page-header">
            <div className="header-titles">
              <h1 className="invoices-title">Invoices</h1>
              <p className="invoices-subtitle">Manage all your customer invoices and billing</p>
            </div>
          </div>

          {/* ── KPI Cards ── */}
          <InvoiceKpiCards kpiData={kpiStats} />

          {/* ── Primary View Tabs & Action Toolbar ── */}
          <div className="invoices-toolbar-card">
            <div className="toolbar-top-row">
              {/* Left: View Tabs (All Invoices vs My Invoices) */}
              <div className="primary-tabs-group">
                <button
                  className={`tab-pill-btn ${primaryTab === 'ALL' ? 'active' : ''}`}
                  onClick={() => {
                    setPrimaryTab('ALL');
                    setCurrentPage(1);
                  }}
                >
                  All Invoices
                </button>
                <button
                  className={`tab-pill-btn ${primaryTab === 'MY' ? 'active' : ''}`}
                  onClick={() => {
                    setPrimaryTab('MY');
                    setCurrentPage(1);
                  }}
                >
                  My Invoices
                </button>
              </div>

              {/* Right: Actions (Filters, Export, + New Invoice) */}
              <div className="toolbar-actions-right">
                <button
                  className="btn-secondary-action"
                  onClick={() => setIsFilterDrawerOpen(true)}
                >
                  <Filter size={15} /> Filters
                </button>

                <button
                  className="btn-secondary-action"
                  onClick={() => handleActionClick('export')}
                >
                  <Download size={15} /> Export
                </button>

                <button
                  className="btn-primary-action"
                  onClick={() => handleActionClick('newInvoice')}
                >
                  <Plus size={16} /> New Invoice
                </button>
              </div>
            </div>

            {/* Status Filter Chips Row */}
            <div className="status-chips-row">
              {Object.values(INVOICE_STATUSES).map((status) => (
                <button
                  key={status}
                  className={`status-chip-btn ${statusFilter === status ? 'active' : ''}`}
                  onClick={() => {
                    setStatusFilter(status);
                    setCurrentPage(1);
                  }}
                >
                  {status}
                </button>
              ))}
            </div>

            {/* Search Bar Row */}
            <div className="search-bar-row">
              <div className="invoice-search-input-wrapper">
                <Search size={16} className="search-icon-svg" />
                <input
                  type="text"
                  placeholder="Search invoices by number, customer, shipment..."
                  value={searchQuery}
                  onChange={(e) => {
                    setSearchQuery(e.target.value);
                    setCurrentPage(1);
                  }}
                  className="invoice-search-input"
                />
                {searchQuery && (
                  <button
                    className="clear-search-btn"
                    onClick={() => {
                      setSearchQuery('');
                      setCurrentPage(1);
                    }}
                  >
                    ×
                  </button>
                )}
              </div>
            </div>
          </div>

          {/* ── Error Banner ── */}
          {error && (
            <div className="invoices-toast banner-warning" style={{ position: 'relative', top: 0, right: 0, marginBottom: '16px' }}>
              <AlertCircle size={16} />
              <span>{error}</span>
              <button onClick={fetchInvoices} style={{ marginLeft: '12px', cursor: 'pointer', background: 'transparent', border: 'none', fontWeight: 700, textDecoration: 'underline' }}>
                Retry
              </button>
            </div>
          )}

          {/* ── Invoice Table / Empty State ── */}
          {loading ? (
            <div className="invoices-skeleton-box">
              {[1, 2, 3, 4, 5].map((i) => (
                <div key={i} className="skeleton-table-row" />
              ))}
            </div>
          ) : invoices.length === 0 ? (
            <div className="invoices-empty-card">
              <div className="empty-icon-wrapper">
                <FileText size={32} className="empty-icon" />
              </div>
              <h3>No invoices found</h3>
              <p>No customer invoices match your current search or filter criteria.</p>
              <div className="empty-actions">
                <button
                  className="btn-secondary-action"
                  onClick={() => {
                    setSearchQuery('');
                    setStatusFilter('All');
                    setPrimaryTab('ALL');
                    setCurrentPage(1);
                  }}
                >
                  Clear Filters
                </button>
                <button
                  className="btn-primary-action"
                  onClick={() => handleActionClick('newInvoice')}
                >
                  <Plus size={15} /> New Invoice
                </button>
              </div>
            </div>
          ) : (
            <div className="invoices-table-card">
              <InvoiceTable
                invoices={invoices}
                selectedInvoice={selectedInvoice}
                onSelectInvoice={(inv) => fetchInvoiceDetails(inv.id)}
                onActionClick={handleActionClick}
              />

              {/* ── Pagination Footer ── */}
              <div className="invoices-pagination-bar">
                <div className="pagination-info">
                  Showing {totalResults === 0 ? 0 : (currentPage - 1) * pageSize + 1} to{' '}
                  {Math.min(currentPage * pageSize, totalResults)} of {totalResults} results
                </div>

                <div className="pagination-controls">
                  <button
                    className="page-nav-btn"
                    disabled={currentPage === 1}
                    onClick={() => setCurrentPage(1)}
                    title="First page"
                  >
                    <ChevronsLeft size={16} />
                  </button>

                  <button
                    className="page-nav-btn"
                    disabled={currentPage === 1}
                    onClick={() => setCurrentPage((prev) => Math.max(prev - 1, 1))}
                    title="Previous page"
                  >
                    <ChevronLeft size={16} />
                  </button>

                  {Array.from({ length: totalPages }, (_, i) => i + 1).map((pageNum) => (
                    <button
                      key={pageNum}
                      className={`page-num-btn ${currentPage === pageNum ? 'active' : ''}`}
                      onClick={() => setCurrentPage(pageNum)}
                    >
                      {pageNum}
                    </button>
                  ))}

                  <button
                    className="page-nav-btn"
                    disabled={currentPage === totalPages}
                    onClick={() => setCurrentPage((prev) => Math.min(prev + 1, totalPages))}
                    title="Next page"
                  >
                    <ChevronRight size={16} />
                  </button>

                  <button
                    className="page-nav-btn"
                    disabled={currentPage === totalPages}
                    onClick={() => setCurrentPage(totalPages)}
                    title="Last page"
                  >
                    <ChevronsRight size={16} />
                  </button>
                </div>

                <div className="pagination-per-page">
                  <select
                    value={pageSize}
                    onChange={(e) => {
                      setPageSize(Number(e.target.value));
                      setCurrentPage(1);
                    }}
                    className="per-page-select"
                  >
                    <option value={10}>10 / page</option>
                    <option value={20}>20 / page</option>
                    <option value={50}>50 / page</option>
                  </select>
                </div>
              </div>
            </div>
          )}
        </div>
      </div>

      {/* ── Slide-over Invoice Detail Drawer (45-50% width with dimmed background) ── */}
      {selectedInvoice && (
        <div className="invoice-drawer-overlay" onClick={() => setSelectedInvoice(null)}>
          <div className="invoice-drawer" onClick={(e) => e.stopPropagation()}>
            <InvoiceDetailsPanel
              invoice={selectedInvoice}
              onClose={() => setSelectedInvoice(null)}
              onActionClick={handleActionClick}
              onToggleBookmark={handleToggleBookmark}
            />
          </div>
        </div>
      )}

      {/* Advanced Filter Drawer */}
      <InvoiceFilterDrawer
        isOpen={isFilterDrawerOpen}
        onClose={() => setIsFilterDrawerOpen(false)}
        filters={advancedFilters}
        onApplyFilters={(newFilters) => {
          setAdvancedFilters(newFilters);
          if (newFilters.status) setStatusFilter(newFilters.status);
          setCurrentPage(1);
        }}
        onResetFilters={() => {
          setAdvancedFilters({ status: 'All' });
          setStatusFilter('All');
          setCurrentPage(1);
        }}
      />

      {/* Create / Edit Invoice Modal */}
      <CreateInvoiceModal
        isOpen={isCreateModalOpen}
        editingInvoice={editingInvoice}
        onClose={() => {
          setIsCreateModalOpen(false);
          setEditingInvoice(null);
        }}
        showToast={showToast}
        onSuccess={(newInv) => {
          fetchInvoices();
          fetchKpiStats();
          if (newInv && newInv.id) {
            fetchInvoiceDetails(newInv.id);
          }
        }}
      />

      {/* Record Payment Modal */}
      <RecordPaymentModal
        isOpen={isRecordPaymentOpen}
        invoice={recordingPaymentInvoice}
        onClose={() => {
          setIsRecordPaymentOpen(false);
          setRecordingPaymentInvoice(null);
        }}
        showToast={showToast}
        onSuccess={(updatedInv) => {
          fetchInvoices();
          fetchKpiStats();
          if (updatedInv && updatedInv.id) {
            fetchInvoiceDetails(updatedInv.id);
          }
        }}
      />
    </div>
  );
}

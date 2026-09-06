import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import toast from 'react-hot-toast';
import {
  X,
  Star,
  MoreHorizontal,
  Mail,
  CreditCard,
  ChevronDown,
  Check,
  Copy,
  Download,
  Building2,
  Truck,
  Calendar,
  Clock,
  ExternalLink,
  CheckCircle2,
  AlertCircle,
  FileText,
  Send,
  Printer,
  ShieldCheck,
  Receipt,
  FileSpreadsheet,
  ArrowRight
} from 'lucide-react';
import InvoiceSummaryTab from './tabs/InvoiceSummaryTab';
import InvoiceItemsTab from './tabs/InvoiceItemsTab';
import InvoicePaymentsTab from './tabs/InvoicePaymentsTab';
import InvoiceDocumentsTab from './tabs/InvoiceDocumentsTab';
import InvoiceHistoryTab from './tabs/InvoiceHistoryTab';
import './InvoiceDetailsPanel.css';

export default function InvoiceDetailsPanel({
  invoice,
  onClose,
  onActionClick,
  onToggleBookmark
}) {
  const navigate = useNavigate();
  const [activeTab, setActiveTab] = useState('Summary');
  const [showMoreActions, setShowMoreActions] = useState(false);
  const [copiedField, setCopiedField] = useState(null);

  if (!invoice) return null;

  const handleCopy = (text, fieldName) => {
    if (!text) return;
    navigator.clipboard.writeText(text);
    setCopiedField(fieldName);
    toast.success(`Copied ${fieldName}: ${text}`);
    setTimeout(() => setCopiedField(null), 2000);
  };

  const renderStatusBadge = (status) => {
    switch (status) {
      case 'Issued':
        return (
          <span className="panel-badge badge-issued">
            <Check size={12} /> Issued
          </span>
        );
      case 'Partially Paid':
        return (
          <span className="panel-badge badge-partially-paid">
            <span className="badge-dot" /> Partially Paid
          </span>
        );
      case 'Overdue':
        return (
          <span className="panel-badge badge-overdue">
            <AlertCircle size={12} /> Overdue
          </span>
        );
      case 'Paid':
        return (
          <span className="panel-badge badge-paid">
            <CheckCircle2 size={12} /> Paid
          </span>
        );
      case 'Pending Approval':
        return (
          <span className="panel-badge badge-pending">
            <Clock size={12} /> Pending Approval
          </span>
        );
      case 'Draft':
        return (
          <span className="panel-badge badge-draft">
            <FileText size={12} /> Draft
          </span>
        );
      case 'Cancelled':
        return (
          <span className="panel-badge badge-cancelled">
            <X size={12} /> Cancelled
          </span>
        );
      default:
        return <span className="panel-badge badge-draft">{status}</span>;
    }
  };

  const handleTabAction = (actionType, payload) => {
    if (actionType === 'switchTab') {
      setActiveTab(payload);
    } else if (actionType === 'navCustomer') {
      navigate('/dashboard/customers');
    } else if (actionType === 'navShipment') {
      navigate('/dashboard/shipments');
    } else if (actionType === 'navBooking') {
      navigate('/dashboard/bookings');
    } else if (actionType === 'navQuotation') {
      navigate('/dashboard/quotations');
    } else if (actionType === 'navApproval') {
      navigate('/dashboard/approvals');
    } else if (onActionClick) {
      onActionClick(actionType, payload || invoice);
    }
  };

  const totalAmount = Number(invoice.amount || 0);
  const amountPaid = Number(invoice.amountPaid || 0);
  const balanceDue = Number(invoice.balance || 0);
  const percentPaid = totalAmount > 0 ? Math.min(100, Math.round((amountPaid / totalAmount) * 100)) : 0;

  return (
    <div className="invoice-details-panel">
      {/* ── Top Header ── */}
      <div className="panel-header">
        <div className="panel-header-left">
          <div className="panel-header-icon-box">
            <Receipt className="w-4 h-4 text-blue-600" />
          </div>
          <div>
            <div className="flex items-center gap-2">
              <h3 className="panel-header-title">Invoice Details</h3>
              <span className="panel-type-tag">Accounts Receivable</span>
            </div>
            <p className="panel-header-sub">Commercial invoice record and settlement status</p>
          </div>
        </div>

        <div className="panel-header-actions">
          <button
            className={`panel-header-btn bookmark-btn ${invoice.bookmarked ? 'is-bookmarked' : ''}`}
            onClick={() => onToggleBookmark?.(invoice.id)}
            title={invoice.bookmarked ? 'Remove Bookmark' : 'Bookmark Invoice'}
          >
            <Star size={14} className={invoice.bookmarked ? 'fill-star' : ''} />
            <span>{invoice.bookmarked ? 'Saved' : 'Save'}</span>
          </button>

          <button
            className="panel-header-btn"
            onClick={() => handleCopy(invoice.invoiceNumber, 'Invoice Number')}
            title="Copy Invoice Number"
          >
            {copiedField === 'Invoice Number' ? <Check size={14} className="text-emerald-600" /> : <Copy size={14} />}
          </button>

          <button
            className="panel-header-btn"
            onClick={() => onActionClick?.('downloadPdf', invoice)}
            title="Download PDF"
          >
            <Download size={14} />
          </button>

          <div className="panel-header-divider" />

          <button
            className="panel-close-btn"
            onClick={onClose}
            title="Close panel (Esc)"
            aria-label="Close invoice details"
          >
            <X size={18} />
          </button>
        </div>
      </div>

      <div className="panel-scroll-body">
        {/* ── Invoice Hero Banner ── */}
        <div className="invoice-hero-card">
          <div className="invoice-hero-top">
            <div className="hero-id-col">
              <div className="flex items-center gap-2.5 flex-wrap">
                <h2 className="invoice-number-heading">{invoice.invoiceNumber}</h2>
                {renderStatusBadge(invoice.status)}
              </div>
              <div className="hero-creator-meta">
                <span>{invoice.creator || 'By Finance Team'}</span>
                <span className="meta-dot">•</span>
                <span>Created {invoice.invoiceDate}</span>
              </div>
            </div>

            {/* Quick Hero Action CTAs */}
            <div className="hero-quick-actions">
              {(invoice.status === 'Issued' || invoice.status === 'Partially Paid' || invoice.status === 'Overdue') && (
                <button
                  className="btn-quick-record-payment"
                  onClick={() => onActionClick?.('recordPayment', invoice)}
                >
                  <CreditCard size={14} />
                  <span>Record Payment</span>
                </button>
              )}

              {invoice.status === 'Draft' && (
                <button
                  className="btn-quick-primary"
                  onClick={() => onActionClick?.('issueInvoice', invoice)}
                >
                  <Check size={14} />
                  <span>Issue Invoice</span>
                </button>
              )}

              {invoice.status === 'Pending Approval' && (
                <button
                  className="btn-quick-primary"
                  onClick={() => onActionClick?.('issueInvoice', invoice)}
                >
                  <ShieldCheck size={14} />
                  <span>Approve & Issue</span>
                </button>
              )}

              {invoice.status === 'Paid' && (
                <button
                  className="btn-quick-paid-receipt"
                  onClick={() => onActionClick?.('sendInvoice', invoice)}
                >
                  <Mail size={14} />
                  <span>Email Receipt</span>
                </button>
              )}
            </div>
          </div>

          {/* ── 4-Card Hero Financial Metrics Grid ── */}
          <div className="hero-metrics-grid">
            <div className="hero-metric-card">
              <span className="metric-label">Total Amount</span>
              <span className="metric-val font-bold text-slate-900">
                ${totalAmount.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}
              </span>
              <span className="metric-currency-tag">{invoice.currency || 'USD'}</span>
            </div>

            <div className="hero-metric-card">
              <span className="metric-label">Amount Paid</span>
              <span className="metric-val font-bold text-emerald-700">
                ${amountPaid.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}
              </span>
              <div className="metric-progress-track">
                <div className="metric-progress-fill" style={{ width: `${percentPaid}%` }} />
              </div>
            </div>

            <div className="hero-metric-card">
              <span className="metric-label">Balance Due</span>
              <span className={`metric-val font-bold ${balanceDue > 0 ? 'text-rose-600' : 'text-emerald-700'}`}>
                ${balanceDue.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}
              </span>
              <span className={`metric-status-sub ${balanceDue > 0 ? 'sub-unpaid' : 'sub-settled'}`}>
                {balanceDue > 0 ? `${100 - percentPaid}% Remaining` : 'Fully Settled'}
              </span>
            </div>

            <div className="hero-metric-card">
              <span className="metric-label">Due Date</span>
              <span className="metric-val font-bold text-slate-900">
                {invoice.dueDate || 'Aug 30, 2026'}
              </span>
              {invoice.daysLeft && (
                <span className={`metric-urgency-badge ${invoice.status === 'Overdue' ? 'urgency-overdue' : 'urgency-normal'}`}>
                  {invoice.daysLeft}
                </span>
              )}
            </div>
          </div>
        </div>

        {/* ── Operational Context Badges Strip ── */}
        <div className="operational-context-strip">
          <div className="context-chip-card" onClick={() => handleTabAction('navCustomer')}>
            <div className="context-chip-icon customer-icon-bg">
              <Building2 size={16} />
            </div>
            <div className="context-chip-body">
              <span className="context-chip-label">Customer</span>
              <span className="context-chip-title">{invoice.customer || 'Customer'}</span>
              <span className="context-chip-sub">{invoice.customerCountry || 'USA'}</span>
            </div>
            <ExternalLink size={13} className="context-chip-arrow" />
          </div>

          <div className="context-chip-card" onClick={() => handleTabAction('navShipment')}>
            <div className="context-chip-icon shipment-icon-bg">
              <Truck size={16} />
            </div>
            <div className="context-chip-body">
              <span className="context-chip-label">Shipment</span>
              <span className="context-chip-title font-mono">{invoice.shipmentId || 'SH-2026-00126'}</span>
              <span className="context-chip-sub">{invoice.route || 'Shanghai ➔ Los Angeles'}</span>
            </div>
            <ExternalLink size={13} className="context-chip-arrow" />
          </div>
        </div>

        {/* ── Navigation Sub-Tabs Bar ── */}
        <div className="details-subtabs-nav">
          {[
            { id: 'Summary', label: 'Summary' },
            { id: 'Items', label: 'Line Items', count: invoice.lineItems?.length || 0 },
            { id: 'Payments', label: 'Payments', count: invoice.payments?.length || 0 },
            { id: 'Documents', label: 'Documents', count: invoice.documents?.length || 0 },
            { id: 'History', label: 'Audit Trail', count: invoice.history?.length || 0 }
          ].map((tab) => (
            <button
              key={tab.id}
              type="button"
              className={`subtab-nav-btn ${activeTab === tab.id ? 'active' : ''}`}
              onClick={() => setActiveTab(tab.id)}
            >
              <span>{tab.label}</span>
              {tab.count !== undefined && tab.count > 0 && (
                <span className="subtab-count-pill">{tab.count}</span>
              )}
            </button>
          ))}
        </div>

        {/* ── Sub-Tab Active View ── */}
        <div className="details-subtab-content">
          {activeTab === 'Summary' && (
            <InvoiceSummaryTab invoice={invoice} onActionClick={handleTabAction} />
          )}
          {activeTab === 'Items' && (
            <InvoiceItemsTab invoice={invoice} />
          )}
          {activeTab === 'Payments' && (
            <InvoicePaymentsTab invoice={invoice} onActionClick={handleTabAction} />
          )}
          {activeTab === 'Documents' && (
            <InvoiceDocumentsTab invoice={invoice} onActionClick={handleTabAction} />
          )}
          {activeTab === 'History' && (
            <InvoiceHistoryTab invoice={invoice} />
          )}
        </div>
      </div>

      {/* ── Sticky Panel Footer Actions ── */}
      <div className="panel-footer-actions">
        {invoice.status === 'Draft' && (
          <>
            <button
              className="btn-panel-primary"
              onClick={() => onActionClick?.('issueInvoice', invoice)}
            >
              <Check size={15} /> Issue Invoice
            </button>
            <button
              className="btn-panel-secondary"
              onClick={() => onActionClick?.('editInvoice', invoice)}
            >
              Edit Draft
            </button>
          </>
        )}

        {invoice.status === 'Pending Approval' && (
          <>
            <button
              className="btn-panel-primary btn-emerald"
              onClick={() => onActionClick?.('issueInvoice', invoice)}
            >
              <ShieldCheck size={15} /> Approve & Issue
            </button>
            <button
              className="btn-panel-secondary"
              onClick={() => handleTabAction('navApproval')}
            >
              View Approval Request
            </button>
          </>
        )}

        {(invoice.status === 'Issued' || invoice.status === 'Partially Paid' || invoice.status === 'Overdue') && (
          <>
            <button
              className="btn-panel-primary btn-emerald"
              onClick={() => onActionClick?.('recordPayment', invoice)}
            >
              <CreditCard size={15} /> Record Payment
            </button>
            <button
              className="btn-panel-secondary"
              onClick={() => onActionClick?.('sendInvoice', invoice)}
            >
              <Mail size={15} /> Send Invoice
            </button>
          </>
        )}

        {invoice.status === 'Paid' && (
          <>
            <button
              className="btn-panel-primary"
              onClick={() => onActionClick?.('downloadPdf', invoice)}
            >
              <Download size={15} /> Download PDF
            </button>
            <button
              className="btn-panel-secondary"
              onClick={() => onActionClick?.('sendInvoice', invoice)}
            >
              <Mail size={15} /> Email Receipt
            </button>
          </>
        )}

        {invoice.status === 'Cancelled' && (
          <div className="cancelled-notice-badge">
            <AlertCircle size={14} />
            <span>This invoice is cancelled and read-only.</span>
          </div>
        )}

        {invoice.status !== 'Cancelled' && (
          <div className="panel-more-container">
            <button
              className="btn-panel-more"
              onClick={() => setShowMoreActions(!showMoreActions)}
              title="More invoice actions"
            >
              <span>More</span>
              <ChevronDown size={14} className={`transition-transform ${showMoreActions ? 'rotate-180' : ''}`} />
            </button>

            {showMoreActions && (
              <div className="panel-more-dropdown">
                {invoice.status === 'Draft' && (
                  <button onClick={() => { setShowMoreActions(false); onActionClick?.('submitApproval', invoice); }}>
                    <Clock size={13} /> Send for Approval
                  </button>
                )}
                <button onClick={() => { setShowMoreActions(false); onActionClick?.('downloadPdf', invoice); }}>
                  <Download size={13} /> Download PDF
                </button>
                <button onClick={() => { setShowMoreActions(false); window.print(); }}>
                  <Printer size={13} /> Print Invoice
                </button>
                {invoice.status !== 'Paid' && (
                  <button
                    className="danger-item"
                    onClick={() => { setShowMoreActions(false); onActionClick?.('cancelInvoice', invoice); }}
                  >
                    <X size={13} /> Cancel Invoice
                  </button>
                )}
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  );
}

import React from 'react';
import {
  X, FileWarning, User, Ship, FileText, Calendar, DollarSign,
  CheckCircle, AlertTriangle, Clock, Ban, Download, Printer,
  Building2, Hash, ArrowUpRight, HelpCircle,
} from 'lucide-react';
import { generateDebitNotePDF } from '../utils/debitNotePdfGenerator';
import './DebitNoteDetailsPanel.css';

const STATUS_CONFIG = {
  DRAFT:        { label: 'Draft',        color: '#475569', bg: '#f1f5f9', border: '#e2e8f0', dot: '#94a3b8', Icon: Clock },
  ISSUED:       { label: 'Issued',       color: '#1d4ed8', bg: '#eff6ff', border: '#bfdbfe', dot: '#2563eb', Icon: AlertTriangle },
  ACKNOWLEDGED: { label: 'Acknowledged', color: '#15803d', bg: '#f0fdf4', border: '#bbf7d0', dot: '#16a34a', Icon: CheckCircle },
  VOID:         { label: 'Void',         color: '#b91c1c', bg: '#fef2f2', border: '#fecaca', dot: '#dc2626', Icon: Ban },
};

function StatusBadge({ status }) {
  const cfg = STATUS_CONFIG[status] || STATUS_CONFIG.DRAFT;
  const Icon = cfg.Icon;
  return (
    <span className="dn-detail-badge" style={{ color: cfg.color, background: cfg.bg, borderColor: cfg.border }}>
      <span className="dn-detail-badge-dot" style={{ background: cfg.dot }} />
      <Icon size={12} />
      {cfg.label}
    </span>
  );
}

export default function DebitNoteDetailsPanel({ debitNote: dn, onClose, onIssue, onVoid }) {
  if (!dn) return null;

  const currency = dn.currency || 'USD';

  const formatCurrency = (val) =>
    `${currency} ${Number(val || 0).toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`;

  const formatDate = (d) => {
    if (!d) return '—';
    return new Date(d).toLocaleDateString('en-GB', { day: '2-digit', month: 'short', year: 'numeric' });
  };

  const handleDownloadPDF = () => {
    generateDebitNotePDF(dn);
  };

  // Ensure line items have at least one fallback row if empty so user sees full itemization
  const items = dn.line_items && dn.line_items.length > 0 ? dn.line_items : [
    {
      id: 'default-1',
      description: dn.reason || 'Ancillary freight adjustment & demurrage/detention charge',
      service_category: 'Ancillary Surcharge',
      quantity: 1,
      unit_price: dn.subtotal || dn.total_amount || 0,
      total_amount: dn.subtotal || dn.total_amount || 0,
    }
  ];

  return (
    <div className="dn-details-panel">
      {/* Top sticky header */}
      <div className="dn-details-header">
        <div className="dn-details-header-title-block">
          <div className="dn-details-tag">COMMERCIAL DEBIT NOTE</div>
          <div className="dn-details-title-row">
            <h2 className="dn-details-dn-number">{dn.debit_note_number}</h2>
            <StatusBadge status={dn.status} />
          </div>
        </div>

        <div className="dn-details-header-actions">
          <button
            className="dn-btn-pdf-header"
            onClick={handleDownloadPDF}
            title="Download / Print PDF"
            id="dn-header-download-pdf-btn"
          >
            <Download size={15} />
            <span>Download PDF</span>
          </button>
          <button
            className="dn-details-close-btn"
            onClick={onClose}
            aria-label="Close"
            title="Close Panel (Esc)"
          >
            <X size={18} />
          </button>
        </div>
      </div>

      {/* Scrollable body content */}
      <div className="dn-details-body">
        {/* Hero Card with Total and Quick Highlights */}
        <div className="dn-hero-card">
          <div className="dn-hero-info">
            <span className="dn-hero-label">Total Amount Payable</span>
            <div className="dn-hero-amount">
              {formatCurrency(dn.total_amount)}
            </div>
            <div className="dn-hero-chips">
              <span className="dn-hero-chip">
                <Hash size={11} /> Currency: <strong>{currency}</strong>
              </span>
              <span className="dn-hero-chip">
                <Calendar size={11} /> Due: <strong>{formatDate(dn.due_date)}</strong>
              </span>
              {dn.invoice_number && (
                <span className="dn-hero-chip accent">
                  <FileText size={11} /> Inv Ref: <strong>{dn.invoice_number}</strong>
                </span>
              )}
            </div>
          </div>
        </div>

        {/* 2-Column Parties Section: Customer & Issuer */}
        <div className="dn-parties-section">
          {/* Customer / Debtor */}
          <div className="dn-party-card">
            <div className="dn-party-card-header">
              <div className="dn-party-icon debtor"><User size={15} /></div>
              <div>
                <span className="dn-party-tag">BILLED TO (DEBTOR)</span>
                <div className="dn-party-name">{dn.customer_name || 'Customer'}</div>
              </div>
            </div>
            <div className="dn-party-details">
              <div className="dn-party-meta-row">
                <span className="dn-party-meta-key">Country:</span>
                <span className="dn-party-country-badge">{dn.customer_country || 'Global Account'}</span>
              </div>
              <div className="dn-party-meta-row">
                <span className="dn-party-meta-key">Account Type:</span>
                <span className="dn-party-meta-val">Commercial Debtor</span>
              </div>
            </div>
          </div>

          {/* Issuer / Freel Logistics */}
          <div className="dn-party-card">
            <div className="dn-party-card-header">
              <div className="dn-party-icon issuer"><Building2 size={15} /></div>
              <div>
                <span className="dn-party-tag">ISSUED BY (FORWARDER)</span>
                <div className="dn-party-name">Freel Global Logistics Pvt. Ltd.</div>
              </div>
            </div>
            <div className="dn-party-details">
              <div className="dn-party-meta-row">
                <span className="dn-party-meta-key">Division:</span>
                <span className="dn-party-meta-val">Freight Billing & Recovery</span>
              </div>
              <div className="dn-party-meta-row">
                <span className="dn-party-meta-key">Tax ID / GST:</span>
                <span className="dn-party-meta-val">27AABCF1234F1Z8</span>
              </div>
            </div>
          </div>
        </div>

        {/* Key Operational References */}
        <div className="dn-meta-strip">
          <div className="dn-meta-tile">
            <span className="dn-tile-label"><Calendar size={12} /> Issue Date</span>
            <span className="dn-tile-value">{formatDate(dn.issue_date)}</span>
          </div>
          <div className="dn-meta-tile">
            <span className="dn-tile-label"><Calendar size={12} /> Due Date</span>
            <span className="dn-tile-value">{formatDate(dn.due_date)}</span>
          </div>
          <div className="dn-meta-tile">
            <span className="dn-tile-label"><FileText size={12} /> Linked Invoice</span>
            <span className="dn-tile-value accent">{dn.invoice_number || '—'}</span>
          </div>
          <div className="dn-meta-tile">
            <span className="dn-tile-label"><Ship size={12} /> Shipment No.</span>
            <span className="dn-tile-value">{dn.shipment_number || '—'}</span>
          </div>
        </div>

        {/* Commercial Reason / Justification */}
        {dn.reason && (
          <div className="dn-reason-banner">
            <div className="dn-reason-header">
              <AlertTriangle size={15} className="dn-reason-icon" />
              <span className="dn-reason-title">Adjustment Justification / Reason</span>
            </div>
            <p className="dn-reason-body">{dn.reason}</p>
          </div>
        )}

        {/* Line items Section */}
        <div className="dn-section-card">
          <div className="dn-section-header">
            <div className="dn-section-title">
              <span>Adjustment Line Items</span>
              <span className="dn-count-pill">{items.length}</span>
            </div>
          </div>

          <div className="dn-table-container">
            <table className="dn-items-table">
              <thead>
                <tr>
                  <th className="th-desc">Description / Service</th>
                  <th className="th-cat">Category</th>
                  <th className="th-qty text-center">Qty</th>
                  <th className="th-price text-right">Unit Rate</th>
                  <th className="th-total text-right">Amount</th>
                </tr>
              </thead>
              <tbody>
                {items.map((item, idx) => (
                  <tr key={item.id || idx}>
                    <td className="td-desc">
                      <div className="dn-item-desc">{item.description}</div>
                    </td>
                    <td className="td-cat">
                      <span className="dn-cat-badge">{item.service_category || 'Ancillary'}</span>
                    </td>
                    <td className="td-qty text-center font-mono">
                      {item.quantity || 1}
                    </td>
                    <td className="td-price text-right font-mono">
                      {formatCurrency(item.unit_price || item.total_amount)}
                    </td>
                    <td className="td-total text-right font-mono font-bold">
                      {formatCurrency(item.total_amount)}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

          {/* Ledger Breakdown Totals */}
          <div className="dn-totals-card">
            <div className="dn-totals-row">
              <span className="dn-totals-label">Subtotal</span>
              <span className="dn-totals-val font-mono">{formatCurrency(dn.subtotal || dn.total_amount)}</span>
            </div>
            {Number(dn.tax_amount) > 0 && (
              <div className="dn-totals-row">
                <span className="dn-totals-label">Tax / VAT</span>
                <span className="dn-totals-val font-mono">{formatCurrency(dn.tax_amount)}</span>
              </div>
            )}
            <div className="dn-totals-row grand-total">
              <span className="dn-totals-label">Total Payable</span>
              <span className="dn-totals-val grand-val font-mono">{formatCurrency(dn.total_amount)}</span>
            </div>
          </div>
        </div>

        {/* Internal Notes / Remarks */}
        {dn.notes && (
          <div className="dn-notes-box">
            <span className="dn-notes-label">Additional Instructions / Notes</span>
            <p className="dn-notes-text">{dn.notes}</p>
          </div>
        )}
      </div>

      {/* Sticky Bottom Action Footer */}
      <div className="dn-details-footer">
        <button
          className="dn-footer-btn dn-footer-btn-pdf"
          onClick={handleDownloadPDF}
          id="dn-footer-download-pdf-btn"
        >
          <Download size={16} />
          <span>Download PDF</span>
        </button>

        {dn.status === 'DRAFT' && (
          <button
            className="dn-footer-btn dn-footer-btn-issue"
            onClick={() => onIssue && onIssue(dn)}
            id="dn-footer-issue-btn"
          >
            <AlertTriangle size={15} />
            <span>Issue Debit Note</span>
          </button>
        )}

        {dn.status !== 'VOID' && (
          <button
            className="dn-footer-btn dn-footer-btn-void"
            onClick={() => onVoid && onVoid(dn)}
            id="dn-footer-void-btn"
          >
            <Ban size={15} />
            <span>Void</span>
          </button>
        )}

        <button className="dn-footer-btn dn-footer-btn-close" onClick={onClose}>
          Close
        </button>
      </div>
    </div>
  );
}

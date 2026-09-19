import React from 'react';
import {
  X, FileWarning, User, Ship, FileText, Calendar, DollarSign,
  CheckCircle, AlertTriangle, Clock, Ban, ChevronRight,
} from 'lucide-react';
import './DebitNoteDetailsPanel.css';

const STATUS_CONFIG = {
  DRAFT:        { label: 'Draft',        color: '#475569', bg: '#f1f5f9', border: '#e2e8f0', Icon: Clock },
  ISSUED:       { label: 'Issued',       color: '#1d4ed8', bg: '#eff6ff', border: '#bfdbfe', Icon: AlertTriangle },
  ACKNOWLEDGED: { label: 'Acknowledged', color: '#15803d', bg: '#f0fdf4', border: '#bbf7d0', Icon: CheckCircle },
  VOID:         { label: 'Void',         color: '#b91c1c', bg: '#fef2f2', border: '#fecaca', Icon: Ban },
};

function StatusBadge({ status }) {
  const cfg = STATUS_CONFIG[status] || STATUS_CONFIG.DRAFT;
  const Icon = cfg.Icon;
  return (
    <span className="dn-panel-status" style={{ color: cfg.color, background: cfg.bg, borderColor: cfg.border }}>
      <Icon size={13} />
      {cfg.label}
    </span>
  );
}

export default function DebitNoteDetailsPanel({ debitNote: dn, onClose, onIssue, onVoid }) {
  if (!dn) return null;

  const formatCurrency = (val) =>
    `${dn.currency || 'USD'} ${Number(val || 0).toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`;

  const formatDate = (d) => {
    if (!d) return '—';
    return new Date(d).toLocaleDateString('en-US', { day: '2-digit', month: 'short', year: 'numeric' });
  };

  return (
    <div className="dn-panel">
      {/* Panel header */}
      <div className="dn-panel-header">
        <div className="dn-panel-header-left">
          <div className="dn-panel-header-icon">
            <FileWarning size={18} />
          </div>
          <div>
            <div className="dn-panel-dn-number">{dn.debit_note_number}</div>
            <StatusBadge status={dn.status} />
          </div>
        </div>
        <button className="dn-panel-close" onClick={onClose} aria-label="Close">
          <X size={18} />
        </button>
      </div>

      {/* Customer block */}
      <div className="dn-panel-section dn-panel-customer-block">
        <div className="dn-panel-customer-icon"><User size={14} /></div>
        <div>
          <div className="dn-panel-customer-name">{dn.customer_name || '—'}</div>
          {dn.customer_country && <div className="dn-panel-customer-sub">{dn.customer_country}</div>}
        </div>
      </div>

      {/* Key fields */}
      <div className="dn-panel-section">
        <div className="dn-panel-grid-meta">
          {dn.invoice_number && (
            <div className="dn-panel-meta-item">
              <span className="dn-meta-label"><FileText size={12} /> Linked Invoice</span>
              <span className="dn-meta-value accent">{dn.invoice_number}</span>
            </div>
          )}
          {dn.shipment_number && (
            <div className="dn-panel-meta-item">
              <span className="dn-meta-label"><Ship size={12} /> Shipment</span>
              <span className="dn-meta-value">{dn.shipment_number}</span>
            </div>
          )}
          {dn.issue_date && (
            <div className="dn-panel-meta-item">
              <span className="dn-meta-label"><Calendar size={12} /> Issue Date</span>
              <span className="dn-meta-value">{formatDate(dn.issue_date)}</span>
            </div>
          )}
          {dn.due_date && (
            <div className="dn-panel-meta-item">
              <span className="dn-meta-label"><Calendar size={12} /> Due Date</span>
              <span className="dn-meta-value">{formatDate(dn.due_date)}</span>
            </div>
          )}
          <div className="dn-panel-meta-item">
            <span className="dn-meta-label"><Calendar size={12} /> Created</span>
            <span className="dn-meta-value">{formatDate(dn.created_at)}</span>
          </div>
        </div>
      </div>

      {/* Reason */}
      {dn.reason && (
        <div className="dn-panel-section">
          <div className="dn-panel-field-label">Reason</div>
          <div className="dn-panel-reason">{dn.reason}</div>
        </div>
      )}

      {/* Line items */}
      {dn.line_items && dn.line_items.length > 0 && (
        <div className="dn-panel-section">
          <div className="dn-panel-field-label">Line Items</div>
          <div className="dn-panel-items">
            {dn.line_items.map((item, idx) => (
              <div key={item.id || idx} className="dn-panel-item-row">
                <div className="dn-panel-item-info">
                  <span className="dn-panel-item-desc">{item.description}</span>
                  {item.service_category && (
                    <span className="dn-panel-item-cat">{item.service_category}</span>
                  )}
                </div>
                <div className="dn-panel-item-amounts">
                  <span className="dn-panel-item-qty">×{item.quantity}</span>
                  <span className="dn-panel-item-price">@ {formatCurrency(item.unit_price)}</span>
                  <span className="dn-panel-item-total">{formatCurrency(item.total_amount)}</span>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Totals */}
      <div className="dn-panel-section dn-panel-totals">
        <div className="dn-panel-total-row">
          <span>Subtotal</span>
          <span>{formatCurrency(dn.subtotal)}</span>
        </div>
        {dn.tax_amount > 0 && (
          <div className="dn-panel-total-row">
            <span>Tax</span>
            <span>{formatCurrency(dn.tax_amount)}</span>
          </div>
        )}
        <div className="dn-panel-total-row dn-panel-grand-total">
          <span>Total</span>
          <strong><DollarSign size={14} />{formatCurrency(dn.total_amount)}</strong>
        </div>
      </div>

      {/* Notes */}
      {dn.notes && (
        <div className="dn-panel-section">
          <div className="dn-panel-field-label">Notes</div>
          <div className="dn-panel-notes">{dn.notes}</div>
        </div>
      )}

      {/* Actions */}
      <div className="dn-panel-actions">
        {dn.status === 'DRAFT' && (
          <button className="dn-action-btn dn-action-issue" onClick={() => onIssue && onIssue(dn)}>
            <AlertTriangle size={15} />
            Issue Debit Note
            <ChevronRight size={14} />
          </button>
        )}
        {dn.status !== 'VOID' && dn.status !== 'DRAFT' && (
          <button className="dn-action-btn dn-action-acknowledge" disabled>
            <CheckCircle size={15} />
            Mark Acknowledged
          </button>
        )}
        {dn.status !== 'VOID' && (
          <button className="dn-action-btn dn-action-void" onClick={() => onVoid && onVoid(dn)}>
            <Ban size={15} />
            Void
          </button>
        )}
      </div>
    </div>
  );
}

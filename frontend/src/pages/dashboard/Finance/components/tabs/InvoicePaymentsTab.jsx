import React from 'react';
import { CreditCard, CheckCircle2, Plus, DollarSign } from 'lucide-react';
import './InvoiceSubTabs.css';

export default function InvoicePaymentsTab({ invoice, onActionClick }) {
  const payments = invoice?.payments || [];
  const totalAmount = Number(invoice?.amount || invoice?.total_amount || 0);
  const amountPaid = Number(invoice?.amountPaid || invoice?.paid_amount || 0);
  const balanceDue = Number(invoice?.balance || invoice?.balance_due || 0);
  const status = invoice?.status || 'Draft';

  const formatCurrency = (val) =>
    `$${(val || 0).toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`;

  const canRecordPayment = status !== 'Draft' && status !== 'Cancelled' && balanceDue > 0;

  return (
    <div className="invoice-subtab-container payments-tab">
      {/* Financial Breakdown Header Card */}
      <div className="payments-financial-header">
        <div className="fin-stat-item">
          <span className="fin-stat-label">Invoice Total</span>
          <span className="fin-stat-val">{formatCurrency(totalAmount)}</span>
        </div>
        <div className="fin-stat-item">
          <span className="fin-stat-label">Total Paid</span>
          <span className="fin-stat-val text-green">{formatCurrency(amountPaid)}</span>
        </div>
        <div className="fin-stat-item">
          <span className="fin-stat-label">Outstanding Balance</span>
          <span className={`fin-stat-val ${balanceDue > 0 ? 'text-red' : 'text-slate'}`}>
            {formatCurrency(balanceDue)}
          </span>
        </div>
        <div className="fin-stat-item status-col">
          <span className="fin-stat-label">Status</span>
          <span className={`payment-status-badge status-${status.toLowerCase().replace(/\s+/g, '-')}`}>
            {status.toUpperCase()}
          </span>
        </div>
      </div>

      {/* Record Payment Action Bar */}
      {canRecordPayment && (
        <div className="payments-action-bar">
          <button
            className="btn-primary-action btn-sm"
            onClick={() => onActionClick?.('recordPayment', invoice)}
          >
            <Plus size={14} /> Record Payment
          </button>
        </div>
      )}

      {/* Payment History List / Table */}
      {payments.length === 0 ? (
        <div className="empty-subtab">
          <div className="empty-subtab-icon-box">
            <CreditCard size={28} className="empty-icon-gray" />
          </div>
          <h4>No Payments Recorded</h4>
          <p className="empty-subtab-text">
            No payment transactions have been logged for this invoice yet.
          </p>
          {canRecordPayment && (
            <button
              className="btn-primary-action"
              style={{ marginTop: '12px' }}
              onClick={() => onActionClick?.('recordPayment', invoice)}
            >
              + Record Payment
            </button>
          )}
        </div>
      ) : (
        <div className="payments-table-wrapper">
          <table className="payments-history-table">
            <thead>
              <tr>
                <th>Payment Date</th>
                <th>Method</th>
                <th>Reference</th>
                <th>Amount</th>
                <th>Status</th>
              </tr>
            </thead>
            <tbody>
              {payments.map((p) => (
                <tr key={p.id || p.reference}>
                  <td className="font-medium">{p.date || 'Aug 31, 2026'}</td>
                  <td>{p.method || 'Wire Transfer'}</td>
                  <td className="font-mono">{p.reference}</td>
                  <td className="text-green font-bold">{formatCurrency(p.amount)}</td>
                  <td>
                    <span className="payment-completed-pill">
                      <CheckCircle2 size={12} /> {p.status || 'Completed'}
                    </span>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}

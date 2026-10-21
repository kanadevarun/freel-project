import React from 'react';
import {
  FileText,
  Eye,
  Download,
  Building2,
  Ship,
  Calendar,
  CheckCircle2,
  AlertTriangle,
  Clock,
  ArrowRight,
  Receipt,
  CreditCard,
  ShieldCheck,
  Package
} from 'lucide-react';
import './InvoiceSubTabs.css';

export default function InvoiceSummaryTab({ invoice, onActionClick }) {
  const lineItems = invoice?.lineItems || [];
  const documents = invoice?.documents || [];

  const formatCurrency = (val) =>
    `$${(val || 0).toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`;

  const isPaid = (invoice?.balance === 0 || invoice?.status === 'Paid') && ((invoice?.amountPaid || 0) > 0 || (invoice?.amount || 0) > 0);
  const isOverdue = invoice?.status === 'Overdue' || ((invoice?.balance || 0) > 0 && String(invoice?.daysLeft || '').toLowerCase().includes('overdue'));

  const shipmentRef = invoice?.shipmentNumber || invoice?.shipmentId;
  const bookingRef = invoice?.bookingNumber || invoice?.bookingId;
  const quoteRef = invoice?.quoteNumber || invoice?.quoteId;

  return (
    <div className="invoice-subtab-container summary-tab">

      {/* ── 1. SETTLEMENT STATUS STAMP BANNER ── */}
      {isPaid ? (
        <div className="inv-settlement-card settled">
          <div className="inv-settlement-left">
            <div className="inv-settlement-icon settled">
              <CheckCircle2 size={18} />
            </div>
            <div>
              <div className="inv-settlement-title">Invoice Settled in Full</div>
              <div className="inv-settlement-sub">
                Realized cash inflow of <strong>{formatCurrency(invoice?.amountPaid || invoice?.amount)}</strong> confirmed. Zero outstanding receivables.
              </div>
            </div>
          </div>
          <div className="inv-settlement-stamp settled">SETTLED</div>
        </div>
      ) : isOverdue ? (
        <div className="inv-settlement-card overdue">
          <div className="inv-settlement-left">
            <div className="inv-settlement-icon overdue">
              <AlertTriangle size={18} />
            </div>
            <div>
              <div className="inv-settlement-title">Payment Overdue</div>
              <div className="inv-settlement-sub">
                Outstanding balance of <strong>{formatCurrency(invoice?.balance)}</strong> is past due ({invoice?.dueDate || 'immediate'}).
              </div>
            </div>
          </div>
          <button
            type="button"
            className="inv-settlement-btn overdue"
            onClick={() => onActionClick?.('recordPayment', invoice)}
          >
            <CreditCard size={13} />
            <span>Record Payment</span>
          </button>
        </div>
      ) : (
        <div className="inv-settlement-card active">
          <div className="inv-settlement-left">
            <div className="inv-settlement-icon active">
              <Clock size={18} />
            </div>
            <div>
              <div className="inv-settlement-title">Commercial Invoice Active</div>
              <div className="inv-settlement-sub">
                Total balance <strong>{formatCurrency(invoice?.balance || invoice?.amount)}</strong> due on <strong>{invoice?.dueDate || 'Net 30'}</strong>.
              </div>
            </div>
          </div>
          <span className="inv-settlement-stamp active">{invoice?.status?.toUpperCase() || 'ISSUED'}</span>
        </div>
      )}

      {/* ── 2. EXECUTIVE PARTIES & CORRIDOR (2-COLUMN CARDS) ── */}
      <div className="inv-summary-grid">
        {/* Bill-To Customer Card */}
        <div className="inv-summary-card">
          <div className="inv-card-header">
            <div className="inv-card-header-icon customer">
              <Building2 size={14} />
            </div>
            <span className="inv-card-header-title">Bill-To Customer</span>
            <span className="inv-card-badge customer">Commercial</span>
          </div>
          <div className="inv-card-body">
            <h4 className="inv-customer-name" title={invoice?.customer || 'Global Traders Inc.'}>
              {invoice?.customer || 'Global Traders Inc.'}
            </h4>
            <div className="inv-meta-sub">
              {invoice?.customerCountry || 'United States'} • Direct Freight Account
            </div>
            <button
              type="button"
              className="inv-link-action"
              onClick={() => onActionClick?.('navCustomer', invoice)}
            >
              <span>View Customer 360</span>
              <ArrowRight size={12} />
            </button>
          </div>
        </div>

        {/* Freight Corridor & Logistics References */}
        <div className="inv-summary-card">
          <div className="inv-card-header">
            <div className="inv-card-header-icon freight">
              <Ship size={14} />
            </div>
            <span className="inv-card-header-title">Freight Route & Corridor</span>
            <span className="inv-card-badge freight">Ocean FCL</span>
          </div>
          <div className="inv-card-body">
            <div className="inv-route-row">
              <span className="inv-route-port origin">{invoice?.origin || 'Shanghai'}</span>
              <span className="inv-route-arrow">➔</span>
              <span className="inv-route-port dest">{invoice?.destination || 'Los Angeles'}</span>
            </div>
            <div className="inv-linked-chips">
              {shipmentRef && (
                <button
                  type="button"
                  className="inv-chip-link"
                  onClick={() => onActionClick?.('navShipment', invoice)}
                  title="View Freight Shipment"
                >
                  <Package size={11} /> {shipmentRef}
                </button>
              )}
              {bookingRef && (
                <button
                  type="button"
                  className="inv-chip-link"
                  onClick={() => onActionClick?.('navBooking', invoice)}
                  title="View Master Booking"
                >
                  <FileText size={11} /> {bookingRef}
                </button>
              )}
              {quoteRef && (
                <button
                  type="button"
                  className="inv-chip-link"
                  onClick={() => onActionClick?.('navQuotation', invoice)}
                  title="View Commercial Quote"
                >
                  <Receipt size={11} /> {quoteRef}
                </button>
              )}
            </div>
          </div>
        </div>
      </div>

      {/* ── 3. BILLING SCHEDULE & LINE ITEMS TABLE ── */}
      <div className="inv-items-card">
        <div className="inv-card-header">
          <div className="inv-card-header-icon items">
            <Receipt size={14} />
          </div>
          <div className="inv-items-header-text">
            <span className="inv-card-header-title">Billed Line Items & Freight Charges</span>
            <span className="inv-count-pill">{lineItems.length} charge{lineItems.length !== 1 ? 's' : ''}</span>
          </div>
          {lineItems.length > 0 && (
            <button
              type="button"
              className="inv-header-action-btn"
              onClick={() => onActionClick?.('switchTab', 'Items')}
            >
              Line item details →
            </button>
          )}
        </div>

        {lineItems.length === 0 ? (
          <div className="inv-empty-state">No line items specified on this invoice file.</div>
        ) : (
          <div className="inv-items-table-wrap">
            <table className="inv-table">
              <thead>
                <tr>
                  <th style={{ width: '56%' }}>Charge Description</th>
                  <th className="th-center" style={{ width: '16%' }}>Qty</th>
                  <th className="th-right" style={{ width: '28%' }}>Amount</th>
                </tr>
              </thead>
              <tbody>
                {lineItems.map((item, idx) => (
                  <tr key={item.id || idx}>
                    <td>
                      <div className="inv-item-desc-cell">
                        <span className="inv-item-bullet" />
                        <div>
                          <span className="inv-item-name">{item.description}</span>
                          {item.rate ? (
                            <span className="inv-item-sub">Unit Rate: {formatCurrency(item.rate)}</span>
                          ) : null}
                        </div>
                      </div>
                    </td>
                    <td className="td-center">
                      <span className="inv-qty-badge">{item.quantity || 1}</span>
                    </td>
                    <td className="td-right">
                      <span className="inv-amount-bold">{formatCurrency(item.amount)}</span>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {/* ── 4. FINANCIAL BREAKDOWN & SETTLEMENT LEDGER ── */}
      <div className="inv-ledger-card">
        <div className="inv-card-header">
          <div className="inv-card-header-icon ledger">
            <ShieldCheck size={14} />
          </div>
          <span className="inv-card-header-title">Settlement Breakdown & Ledger Totals</span>
          <span className="inv-currency-tag">{invoice?.currency || 'USD'}</span>
        </div>

        <div className="inv-ledger-rows">
          <div className="inv-ledger-row">
            <span className="ledger-label">Subtotal</span>
            <span className="ledger-val">{formatCurrency(invoice?.subtotal)}</span>
          </div>

          <div className="inv-ledger-row">
            <span className="ledger-label">Tax / GST ({invoice?.tax > 0 ? '8%' : '0%'})</span>
            <span className="ledger-val">{formatCurrency(invoice?.tax)}</span>
          </div>

          {invoice?.discount != null && invoice?.discount !== 0 && (
            <div className="inv-ledger-row discount">
              <span className="ledger-label">Contract Discount / Waiver</span>
              <span className="ledger-val discount-val">-${Math.abs(invoice.discount).toFixed(2)}</span>
            </div>
          )}

          <div className="inv-ledger-divider" />

          <div className="inv-ledger-row total">
            <span className="ledger-label total-label">Total Billed</span>
            <span className="ledger-val total-val">{formatCurrency(invoice?.amount)}</span>
          </div>

          <div className="inv-ledger-row paid">
            <span className="ledger-label">Amount Paid to Date</span>
            <span className="ledger-val paid-val">{formatCurrency(invoice?.amountPaid)}</span>
          </div>

          <div className={`inv-balance-box ${isPaid ? 'paid-zero' : 'has-due'}`}>
            <div className="inv-balance-left">
              <span className="balance-label">Balance Due</span>
              <span className="balance-status-text">
                {isPaid ? 'Settlement verified • Closed file' : `Payment due ${invoice?.dueDate || 'on receipt'}`}
              </span>
            </div>
            <span className="balance-amount">{formatCurrency(invoice?.balance)}</span>
          </div>
        </div>
      </div>

      {/* ── 5. ATTACHED DOCUMENTS ── */}
      <div className="inv-docs-card">
        <div className="inv-card-header">
          <div className="inv-card-header-icon docs">
            <FileText size={14} />
          </div>
          <div className="inv-items-header-text">
            <span className="inv-card-header-title">Commercial Documents & Attachments</span>
            <span className="inv-count-pill">{documents.length}</span>
          </div>
          {documents.length > 0 && (
            <button
              type="button"
              className="inv-header-action-btn"
              onClick={() => onActionClick?.('switchTab', 'Documents')}
            >
              All files →
            </button>
          )}
        </div>

        {documents.length === 0 ? (
          <div className="inv-empty-state">No commercial documents attached.</div>
        ) : (
          <div className="inv-docs-grid">
            {documents.map((doc) => (
              <div key={doc.id} className="inv-doc-pill">
                <div className="inv-doc-icon-wrap">
                  <FileText size={16} className="pdf-icon" />
                </div>
                <div className="inv-doc-info">
                  <span className="inv-doc-name" title={doc.name}>{doc.name}</span>
                  <span className="inv-doc-size">{doc.size || 'PDF'} • Uploaded</span>
                </div>
                <div className="inv-doc-btns">
                  <button
                    type="button"
                    className="inv-doc-btn"
                    title="View Document"
                    onClick={() => onActionClick?.('viewDoc', doc)}
                  >
                    <Eye size={13} />
                  </button>
                  <button
                    type="button"
                    className="inv-doc-btn"
                    title="Download File"
                    onClick={() => onActionClick?.('downloadDoc', doc)}
                  >
                    <Download size={13} />
                  </button>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>

    </div>
  );
}

import React from 'react';
import { FileText, Eye, Download } from 'lucide-react';
import './InvoiceSubTabs.css';

export default function InvoiceSummaryTab({ invoice, onActionClick }) {
  const lineItems = invoice?.lineItems || [];
  const documents = invoice?.documents || [];

  const formatCurrency = (val) =>
    `$${(val || 0).toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`;

  return (
    <div className="invoice-subtab-container summary-tab">
      {/* Operational Context & Cross-Module References Card */}
      <div className="summary-operational-context-card">
        <h4 className="context-card-title">Operational Context & References</h4>
        <div className="context-links-grid">
          <div className="context-link-item">
            <span className="context-label">Customer</span>
            <button
              className="context-nav-btn"
              onClick={() => onActionClick?.('navCustomer', invoice)}
              title="View Customer Profile"
            >
              {invoice?.customer || 'Global Traders Inc.'}
            </button>
          </div>

          {invoice?.shipmentNumber && (
            <div className="context-link-item">
              <span className="context-label">Shipment</span>
              <button
                className="context-nav-btn"
                onClick={() => onActionClick?.('navShipment', invoice)}
                title="View Shipment Details"
              >
                {invoice.shipmentNumber} ({invoice?.route || 'Shanghai ➔ LA'})
              </button>
            </div>
          )}

          {invoice?.bookingNumber && (
            <div className="context-link-item">
              <span className="context-label">Booking</span>
              <button
                className="context-nav-btn"
                onClick={() => onActionClick?.('navBooking', invoice)}
                title="View Freight Booking"
              >
                {invoice.bookingNumber}
              </button>
            </div>
          )}

          {invoice?.quoteNumber && (
            <div className="context-link-item">
              <span className="context-label">Quotation</span>
              <button
                className="context-nav-btn"
                onClick={() => onActionClick?.('navQuotation', invoice)}
                title="View Commercial Quote"
              >
                {invoice.quoteNumber}
              </button>
            </div>
          )}

          <div className="context-link-item">
            <span className="context-label">Approval Context</span>
            <button
              className="context-nav-btn approval-link-btn"
              onClick={() => onActionClick?.('navApproval', invoice)}
              title="View Approvals Workspace"
            >
              {invoice?.status === 'Pending Approval' ? 'Pending Approval' : 'Finance Approval Standard'}
            </button>
          </div>
        </div>
      </div>

      {/* Line Items Summary Section */}
      {lineItems.length > 0 && (
        <div className="summary-items-section" style={{ marginBottom: '20px' }}>
          <div className="doc-section-header">
            <h4 className="doc-section-title">Line Items ({lineItems.length})</h4>
            <button className="btn-view-all-docs" onClick={() => onActionClick?.('switchTab', 'Items')}>
              View details
            </button>
          </div>
          <table className="subtab-table" style={{ marginTop: '10px' }}>
            <thead>
              <tr>
                <th>Description</th>
                <th className="th-center">Qty</th>
                <th className="th-right">Amount</th>
              </tr>
            </thead>
            <tbody>
              {lineItems.map((item) => (
                <tr key={item.id}>
                  <td>
                    <span className="item-description">{item.description}</span>
                  </td>
                  <td className="td-center">{item.quantity}</td>
                  <td className="td-right bold-col">{formatCurrency(item.amount)}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {/* Financial Totals Card */}
      <div className="summary-financial-card">
        <div className="summary-row">
          <span className="summary-label">Subtotal</span>
          <span className="summary-value">{formatCurrency(invoice?.subtotal)}</span>
        </div>
        <div className="summary-row">
          <span className="summary-label">Tax ({invoice?.tax > 0 ? '8%' : '0%'})</span>
          <span className="summary-value">{formatCurrency(invoice?.tax)}</span>
        </div>
        <div className="summary-row">
          <span className="summary-label">Discount</span>
          <span className="summary-value">-${(invoice?.discount || 0).toFixed(2)}</span>
        </div>
        <div className="summary-row total-row">
          <span className="summary-label-bold">Total</span>
          <span className="summary-value-bold">{formatCurrency(invoice?.amount)}</span>
        </div>
        <div className="summary-row divider-top">
          <span className="summary-label">Amount Paid</span>
          <span className="summary-value">{formatCurrency(invoice?.amountPaid)}</span>
        </div>
        <div className="summary-row balance-row">
          <span className="summary-label-balance">Balance Due</span>
          <span className={`summary-value-balance ${invoice?.balance > 0 ? 'has-balance' : 'zero-balance'}`}>
            {formatCurrency(invoice?.balance)}
          </span>
        </div>
      </div>

      {/* Attached Documents Section */}
      <div className="summary-documents-section">
        <div className="doc-section-header">
          <h4 className="doc-section-title">Attached Documents ({documents.length})</h4>
          {documents.length > 0 && (
            <button className="btn-view-all-docs" onClick={() => onActionClick?.('switchTab', 'Documents')}>
              View all
            </button>
          )}
        </div>

        {documents.length === 0 ? (
          <p className="no-docs-text">No documents attached to this invoice.</p>
        ) : (
          <div className="attached-doc-list">
            {documents.map((doc) => (
              <div key={doc.id} className="attached-doc-item">
                <div className="doc-item-icon">
                  <FileText size={18} className="pdf-icon-red" />
                </div>
                <div className="doc-item-meta">
                  <span className="doc-filename" title={doc.name}>{doc.name}</span>
                  <span className="doc-filesize">{doc.size}</span>
                </div>
                <div className="doc-item-actions">
                  <button
                    className="doc-action-btn"
                    title="View Document"
                    onClick={() => onActionClick?.('viewDoc', doc)}
                  >
                    <Eye size={14} />
                  </button>
                  <button
                    className="doc-action-btn"
                    title="Download Document"
                    onClick={() => onActionClick?.('downloadDoc', doc)}
                  >
                    <Download size={14} />
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

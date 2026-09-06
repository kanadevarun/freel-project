import React, { useState } from 'react';
import { MoreHorizontal, CheckCircle2, Clock, AlertTriangle, FileText, Check, X, ArrowDown } from 'lucide-react';
import './InvoiceTable.css';

export default function InvoiceTable({
  invoices,
  selectedInvoice,
  onSelectInvoice,
  onActionClick
}) {
  const [activeActionMenuId, setActiveActionMenuId] = useState(null);

  const renderStatusBadge = (status) => {
    switch (status) {
      case 'Issued':
        return (
          <span className="inv-badge badge-issued">
            <Check size={12} className="badge-icon" /> Issued
          </span>
        );
      case 'Partially Paid':
        return (
          <span className="inv-badge badge-partially-paid">
            <span className="badge-diamond">◆</span> Partially Paid
          </span>
        );
      case 'Overdue':
        return (
          <span className="inv-badge badge-overdue">
            <ArrowDown size={12} className="badge-icon" /> Overdue
          </span>
        );
      case 'Paid':
        return (
          <span className="inv-badge badge-paid">
            <Check size={12} className="badge-icon" /> Paid
          </span>
        );
      case 'Pending Approval':
        return (
          <span className="inv-badge badge-pending">
            <span className="badge-diamond">◆</span> Pending Approval
          </span>
        );
      case 'Draft':
        return (
          <span className="inv-badge badge-draft">
            <Check size={12} className="badge-icon" /> Draft
          </span>
        );
      case 'Cancelled':
        return (
          <span className="inv-badge badge-cancelled">
            <X size={12} className="badge-icon" /> Cancelled
          </span>
        );
      default:
        return <span className="inv-badge badge-draft">{status}</span>;
    }
  };

  return (
    <div className="invoice-table-wrapper">
      <table className="invoice-data-table">
        <thead>
          <tr>
            <th>Invoice #</th>
            <th>Customer</th>
            <th>Shipment / Ref</th>
            <th>Invoice Date</th>
            <th>Due Date</th>
            <th>Amount</th>
            <th>Status</th>
            <th>Balance</th>
            <th className="th-actions">Actions</th>
          </tr>
        </thead>
        <tbody>
          {invoices.map((inv) => {
            const isSelected = selectedInvoice?.id === inv.id;

            return (
              <tr
                key={inv.id}
                className={`invoice-row ${isSelected ? 'row-selected' : ''}`}
                onClick={() => onSelectInvoice(inv)}
              >
                <td className="td-invoice-number">
                  <span className="inv-num-link">{inv.invoiceNumber}</span>
                  {inv.creator && <span className="inv-subtext">{inv.creator}</span>}
                </td>

                <td className="td-customer">
                  <span className="inv-customer-name">{inv.customer}</span>
                  <span className="inv-subtext">{inv.customerCountry}</span>
                </td>

                <td className="td-shipment">
                  <span className="inv-shipment-id">{inv.shipmentId}</span>
                  <span className="inv-subtext route-text">{inv.route}</span>
                </td>

                <td className="td-date">{inv.invoiceDate}</td>

                <td className="td-date">
                  <span>{inv.dueDate}</span>
                  {inv.daysLeft && (
                    <span className={`inv-subtext urgency-text ${inv.status === 'Overdue' ? 'overdue' : ''}`}>
                      {inv.daysLeft}
                    </span>
                  )}
                </td>

                <td className="td-amount">
                  <span className="inv-amount-val">${inv.amount.toLocaleString('en-US', { minimumFractionDigits: 2 })}</span>
                  <span className="inv-subtext">{inv.currency}</span>
                </td>

                <td className="td-status">
                  {renderStatusBadge(inv.status)}
                </td>

                <td className="td-balance">
                  <span className={`inv-balance-val ${inv.balance > 0 ? 'has-balance' : 'zero-balance'}`}>
                    ${inv.balance.toLocaleString('en-US', { minimumFractionDigits: 2 })}
                  </span>
                </td>

                <td className="td-actions" onClick={(e) => e.stopPropagation()}>
                  <div className="action-menu-container">
                    <button
                      className="inv-action-dots-btn"
                      onClick={() => setActiveActionMenuId(activeActionMenuId === inv.id ? null : inv.id)}
                      title="More actions"
                    >
                      <MoreHorizontal size={16} />
                    </button>

                    {activeActionMenuId === inv.id && (
                      <div className="inv-dropdown-menu">
                        <button
                          onClick={() => {
                            setActiveActionMenuId(null);
                            onSelectInvoice(inv);
                          }}
                        >
                          View Details
                        </button>
                        <button
                          onClick={() => {
                            setActiveActionMenuId(null);
                            onActionClick('recordPayment', inv);
                          }}
                        >
                          Record Payment
                        </button>
                        <button
                          onClick={() => {
                            setActiveActionMenuId(null);
                            onActionClick('sendInvoice', inv);
                          }}
                        >
                          Send Invoice
                        </button>
                        <button
                          onClick={() => {
                            setActiveActionMenuId(null);
                            onActionClick('downloadPdf', inv);
                          }}
                        >
                          Download PDF
                        </button>
                      </div>
                    )}
                  </div>
                </td>
              </tr>
            );
          })}
        </tbody>
      </table>
    </div>
  );
}

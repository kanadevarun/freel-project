import React, { useState } from 'react';
import { MoreHorizontal, CheckCircle2, Clock, AlertTriangle, FileText, Check, X, ArrowDown, Sparkles } from 'lucide-react';
import AgentStatusBadge from '../../../../components/agent/AgentStatusBadge';
import './InvoiceTable.css';

export default function InvoiceTable({
  invoices,
  selectedInvoice,
  onSelectInvoice,
  onActionClick,
  onOpenAdaptiveCollection,
  aiTasksByRef = {}
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
            const aiTask = aiTasksByRef[inv.invoiceNumber] || aiTasksByRef[inv.id] || null;

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
                  <div style={{ display: 'flex', flexDirection: 'column', gap: '4px', alignItems: 'flex-start' }}>
                    {renderStatusBadge(inv.status)}
                    {aiTask && (
                      <AgentStatusBadge
                        status={aiTask.workforce_status}
                        error={aiTask.safe_error_msg}
                        mockMode={aiTask.mock_mode}
                        providerFailover={aiTask.provider_failover}
                      />
                    )}
                  </div>
                </td>

                <td className="td-balance">
                  <span className={`inv-balance-val ${inv.balance > 0 ? 'has-balance' : 'zero-balance'}`}>
                    ${inv.balance.toLocaleString('en-US', { minimumFractionDigits: 2 })}
                  </span>
                </td>

                <td className="td-actions" onClick={(e) => e.stopPropagation()} style={{ position: 'relative', zIndex: 5 }}>
                  <div style={{ display: 'flex', alignItems: 'center', gap: '6px', justifyContent: 'flex-end' }}>
                    <button
                      type="button"
                      className="inv-action-ai-btn"
                      onClick={() => onOpenAdaptiveCollection ? onOpenAdaptiveCollection(inv) : onActionClick?.('adaptiveCollections', inv)}
                      title="Open Adaptive Collections AI"
                      data-testid={`btn-ai-collection-${inv.id}`}
                      style={{
                        display: 'inline-flex',
                        alignItems: 'center',
                        gap: '4px',
                        padding: '4px 8px',
                        fontSize: '11px',
                        fontWeight: 600,
                        borderRadius: '4px',
                        backgroundColor: '#eff6ff',
                        color: '#1d4ed8',
                        border: '1px solid #bfdbfe',
                        cursor: 'pointer'
                      }}
                    >
                      <Sparkles size={12} />
                      <span>AI Collection</span>
                    </button>

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
                              onOpenAdaptiveCollection ? onOpenAdaptiveCollection(inv) : onActionClick?.('adaptiveCollections', inv);
                            }}
                            data-testid={`menu-item-ai-collection-${inv.id}`}
                          >
                            Adaptive Collections AI
                          </button>
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

import React from 'react';
import './InvoiceSubTabs.css';

export default function InvoiceItemsTab({ invoice }) {
  const lineItems = invoice?.lineItems || [];

  const formatCurrency = (val) =>
    `$${(val || 0).toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`;

  if (lineItems.length === 0) {
    return (
      <div className="invoice-subtab-container empty-subtab">
        <p className="empty-subtab-text">No line items recorded for this invoice.</p>
      </div>
    );
  }

  return (
    <div className="invoice-subtab-container items-tab">
      <table className="subtab-table">
        <thead>
          <tr>
            <th>Description</th>
            <th className="th-center">Qty</th>
            <th className="th-right">Rate</th>
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
              <td className="td-right">{formatCurrency(item.rate)}</td>
              <td className="td-right bold-col">{formatCurrency(item.amount)}</td>
            </tr>
          ))}
        </tbody>
      </table>

      <div className="items-total-footer">
        <span>Total Line Items: {lineItems.length}</span>
        <span className="items-grand-total">Total: {formatCurrency(invoice?.amount)}</span>
      </div>
    </div>
  );
}

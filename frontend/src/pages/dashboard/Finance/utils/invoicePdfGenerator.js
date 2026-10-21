/**
 * invoicePdfGenerator.js
 * Authoritative Client-Side Commercial Invoice & PDF Generator
 * Renders an official, customer-ready commercial freight invoice formatted for A4 PDF export.
 */

export function generateInvoicePDF(inv) {
  if (!inv) return;

  const currency = inv.currency || 'USD';
  const formatMoney = (val) =>
    `${currency} ${Number(val || 0).toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`;

  const formatDate = (d) => {
    if (!d) return '—';
    try {
      const parsed = new Date(d);
      if (isNaN(parsed.getTime())) return d;
      return parsed.toLocaleDateString('en-US', { day: '2-digit', month: 'short', year: 'numeric' });
    } catch (_) {
      return d;
    }
  };

  const lineItems = inv.lineItems && inv.lineItems.length > 0 ? inv.lineItems : (
    inv.line_items && inv.line_items.length > 0 ? inv.line_items : [
      {
        description: 'International Freight Transportation & Logistics Services',
        quantity: 1,
        rate: inv.amount || inv.total_amount || 0,
        amount: inv.amount || inv.total_amount || 0,
      }
    ]
  );

  const totalAmount = Number(inv.amount || inv.total_amount || 0);
  const subtotal = Number(inv.subtotal || totalAmount);
  const tax = Number(inv.tax || inv.tax_amount || 0);
  const discount = Number(inv.discount || inv.discount_amount || 0);
  const amountPaid = Number(inv.amountPaid || inv.paid_amount || 0);
  const balanceDue = Number(inv.balance !== undefined ? inv.balance : (inv.balance_due !== undefined ? inv.balance_due : (totalAmount - amountPaid)));

  const status = (inv.status || 'ISSUED').toUpperCase();

  const getStatusColor = (s) => {
    switch (s) {
      case 'PAID':
        return { bg: '#f0fdf4', text: '#15803d', border: '#bbf7d0' };
      case 'PARTIALLY PAID':
        return { bg: '#fffbeb', text: '#b45309', border: '#fde68a' };
      case 'OVERDUE':
        return { bg: '#fef2f2', text: '#b91c1c', border: '#fecaca' };
      case 'ISSUED':
        return { bg: '#eff6ff', text: '#1d4ed8', border: '#bfdbfe' };
      default:
        return { bg: '#f1f5f9', text: '#475569', border: '#cbd5e1' };
    }
  };

  const statusStyle = getStatusColor(status);

  const html = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>Invoice ${inv.invoiceNumber || inv.invoice_number || 'INV'}</title>
  <style>
    @page {
      size: A4 portrait;
      margin: 12mm 14mm;
    }
    * {
      box-sizing: border-box;
      -webkit-print-color-adjust: exact;
      print-color-adjust: exact;
    }
    body {
      font-family: 'Segoe UI', -apple-system, BlinkMacSystemFont, Roboto, sans-serif;
      color: #0f172a;
      background: #ffffff;
      margin: 0;
      padding: 24px;
      font-size: 13px;
      line-height: 1.5;
    }
    .inv-header {
      display: flex;
      justify-content: space-between;
      align-items: flex-start;
      border-bottom: 2px solid #2563eb;
      padding-bottom: 18px;
      margin-bottom: 20px;
    }
    .inv-brand-title {
      font-size: 22px;
      font-weight: 800;
      color: #0f172a;
      letter-spacing: -0.02em;
    }
    .inv-brand-title span {
      color: #2563eb;
    }
    .inv-brand-sub {
      font-size: 11px;
      color: #64748b;
      margin-top: 3px;
      text-transform: uppercase;
      letter-spacing: 0.05em;
    }
    .inv-type-block {
      text-align: right;
    }
    .inv-type-title {
      font-size: 20px;
      font-weight: 800;
      color: #2563eb;
      text-transform: uppercase;
      letter-spacing: 0.05em;
    }
    .inv-number-text {
      font-family: ui-monospace, SFMono-Regular, monospace;
      font-size: 15px;
      font-weight: 700;
      color: #0f172a;
      margin-top: 3px;
    }
    .inv-status-pill {
      display: inline-block;
      font-size: 10px;
      font-weight: 800;
      padding: 3px 10px;
      border-radius: 999px;
      margin-top: 5px;
      text-transform: uppercase;
      background: ${statusStyle.bg};
      color: ${statusStyle.text};
      border: 1px solid ${statusStyle.border};
    }

    .inv-parties-grid {
      display: grid;
      grid-template-columns: 1fr 1fr;
      gap: 20px;
      margin-bottom: 20px;
    }
    .inv-party-card {
      background: #f8fafc;
      border: 1px solid #e2e8f0;
      border-radius: 8px;
      padding: 14px 16px;
    }
    .inv-party-tag {
      font-size: 10px;
      font-weight: 700;
      text-transform: uppercase;
      color: #64748b;
      letter-spacing: 0.06em;
      margin-bottom: 5px;
    }
    .inv-party-name {
      font-size: 15px;
      font-weight: 700;
      color: #0f172a;
      margin-bottom: 4px;
    }
    .inv-party-meta {
      font-size: 12px;
      color: #475569;
      line-height: 1.4;
    }

    .inv-ref-strip {
      display: grid;
      grid-template-columns: repeat(4, 1fr);
      gap: 12px;
      background: #f1f5f9;
      border: 1px solid #cbd5e1;
      border-radius: 8px;
      padding: 10px 14px;
      margin-bottom: 22px;
    }
    .inv-ref-item {
      display: flex;
      flex-direction: column;
    }
    .inv-ref-label {
      font-size: 10px;
      font-weight: 700;
      text-transform: uppercase;
      color: #64748b;
    }
    .inv-ref-value {
      font-size: 12px;
      font-weight: 600;
      color: #0f172a;
      font-family: ui-monospace, SFMono-Regular, monospace;
      margin-top: 2px;
    }

    .inv-table {
      width: 100%;
      border-collapse: collapse;
      margin-bottom: 20px;
      border: 1px solid #e2e8f0;
      border-radius: 8px;
      overflow: hidden;
    }
    .inv-table th {
      background: #f8fafc;
      padding: 10px 12px;
      font-size: 11px;
      font-weight: 700;
      text-transform: uppercase;
      letter-spacing: 0.04em;
      color: #475569;
      border-bottom: 1px solid #cbd5e1;
      text-align: left;
    }
    .inv-table td {
      padding: 11px 12px;
      font-size: 12.5px;
      border-bottom: 1px solid #f1f5f9;
      color: #1e293b;
    }
    .inv-table tbody tr:last-child td {
      border-bottom: none;
    }
    .text-center { text-align: center; }
    .text-right { text-align: right; }
    .font-bold { font-weight: 700; }
    .font-mono { font-family: ui-monospace, SFMono-Regular, monospace; }

    .inv-bottom-section {
      display: grid;
      grid-template-columns: 1.4fr 1fr;
      gap: 24px;
      margin-bottom: 24px;
    }
    .inv-remit-card {
      background: #f8fafc;
      border: 1px solid #e2e8f0;
      border-radius: 8px;
      padding: 14px 16px;
      font-size: 11.5px;
      color: #475569;
      line-height: 1.6;
    }
    .inv-remit-title {
      font-size: 11px;
      font-weight: 700;
      text-transform: uppercase;
      color: #0f172a;
      margin-bottom: 6px;
      letter-spacing: 0.04em;
    }
    .inv-totals-card {
      background: #f8fafc;
      border: 1px solid #e2e8f0;
      border-radius: 8px;
      padding: 14px 16px;
    }
    .inv-total-row {
      display: flex;
      justify-content: space-between;
      align-items: center;
      padding: 4px 0;
      font-size: 12.5px;
      color: #475569;
    }
    .inv-total-row.grand-total {
      border-top: 1px solid #cbd5e1;
      margin-top: 6px;
      padding-top: 8px;
      font-size: 15px;
      font-weight: 800;
      color: #0f172a;
    }
    .inv-total-row.balance-due {
      border-top: 1px dashed #cbd5e1;
      margin-top: 6px;
      padding-top: 8px;
      font-size: 14px;
      font-weight: 800;
      color: #2563eb;
    }

    .inv-footer {
      border-top: 1px solid #e2e8f0;
      padding-top: 16px;
      display: flex;
      justify-content: space-between;
      align-items: flex-end;
      font-size: 11px;
      color: #64748b;
    }
    .inv-sign-block {
      text-align: right;
    }
    .inv-sign-line {
      width: 170px;
      border-bottom: 1px solid #0f172a;
      margin-bottom: 5px;
    }

    @media print {
      body {
        padding: 0;
      }
    }
  </style>
</head>
<body>
  <!-- Header -->
  <div class="inv-header">
    <div>
      <div class="inv-brand-title">FREEL <span>LOGISTICS</span></div>
      <div class="inv-brand-sub">Global Freight Forwarding & Supply Chain Solutions</div>
    </div>
    <div class="inv-type-block">
      <div class="inv-type-title">COMMERCIAL INVOICE</div>
      <div class="inv-number-text">${inv.invoiceNumber || inv.invoice_number || 'INV-DRAFT'}</div>
      <span class="inv-status-pill">${status}</span>
    </div>
  </div>

  <!-- Parties Grid -->
  <div class="inv-parties-grid">
    <div class="inv-party-card">
      <div class="inv-party-tag">Billed To (Customer / Consignee)</div>
      <div class="inv-party-name">${inv.customer || inv.customer_name || 'Valued Commercial Client'}</div>
      <div class="inv-party-meta">
        ${inv.customerCountry || inv.customer_country ? `Country: ${inv.customerCountry || inv.customer_country}` : 'International Accounts'}<br>
        Direct Customer Ledger Reference
      </div>
    </div>

    <div class="inv-party-card">
      <div class="inv-party-tag">Issued By (Freight Forwarder)</div>
      <div class="inv-party-name">Freel Global Logistics Pvt. Ltd.</div>
      <div class="inv-party-meta">
        Corporate Billing & Receivables Department<br>
        GST / Tax ID: 27AABCF1234F1Z8<br>
        Email: billing@freellogistics.com
      </div>
    </div>
  </div>

  <!-- References Strip -->
  <div class="inv-ref-strip">
    <div class="inv-ref-item">
      <span class="inv-ref-label">Invoice Date</span>
      <span class="inv-ref-value">${formatDate(inv.invoiceDate || inv.invoice_date)}</span>
    </div>
    <div class="inv-ref-item">
      <span class="inv-ref-label">Due Date</span>
      <span class="inv-ref-value">${formatDate(inv.dueDate || inv.due_date)}</span>
    </div>
    <div class="inv-ref-item">
      <span class="inv-ref-label">Shipment Ref</span>
      <span class="inv-ref-value">${inv.shipmentId || inv.shipment_number || '—'}</span>
    </div>
    <div class="inv-ref-item">
      <span class="inv-ref-label">Routing / Ports</span>
      <span class="inv-ref-value">${inv.route || 'Global Routing'}</span>
    </div>
  </div>

  <!-- Items Table -->
  <table class="inv-table">
    <thead>
      <tr>
        <th style="width: 50%;">Service Description & Logistics Line Items</th>
        <th class="text-center" style="width: 12%;">Qty</th>
        <th class="text-right" style="width: 18%;">Unit Rate</th>
        <th class="text-right" style="width: 20%;">Total Amount</th>
      </tr>
    </thead>
    <tbody>
      ${lineItems.map(item => `
        <tr>
          <td class="font-bold">${item.description || item.name || 'Freight Forwarding Charge'}</td>
          <td class="text-center font-mono">${item.quantity || 1}</td>
          <td class="text-right font-mono">${formatMoney(item.rate || item.unit_price || item.amount || 0)}</td>
          <td class="text-right font-bold font-mono">${formatMoney(item.amount || item.total_amount || 0)}</td>
        </tr>
      `).join('')}
    </tbody>
  </table>

  <!-- Bottom Breakdown & Remittance -->
  <div class="inv-bottom-section">
    <!-- Remittance Bank Details -->
    <div class="inv-remit-card">
      <div class="inv-remit-title">Wire Transfer & Settlement Instructions</div>
      Please remit funds in <strong>${currency}</strong> to:<br>
      <strong>Beneficiary:</strong> Freel Global Logistics Pvt. Ltd.<br>
      <strong>Bank:</strong> Standard Chartered Bank<br>
      <strong>Account Number:</strong> 9812-4421-9901<br>
      <strong>SWIFT / BIC:</strong> SCBLINBBXXX<br>
      Please quote invoice number <strong>${inv.invoiceNumber || inv.invoice_number}</strong> on all wire payment advices.
    </div>

    <!-- Financial Totals -->
    <div class="inv-totals-card">
      <div class="inv-total-row">
        <span>Subtotal:</span>
        <span class="font-mono">${formatMoney(subtotal)}</span>
      </div>
      ${tax > 0 ? `
      <div class="inv-total-row">
        <span>Tax / VAT:</span>
        <span class="font-mono">${formatMoney(tax)}</span>
      </div>
      ` : ''}
      ${discount > 0 ? `
      <div class="inv-total-row">
        <span>Discount:</span>
        <span class="font-mono">-${formatMoney(discount)}</span>
      </div>
      ` : ''}
      <div class="inv-total-row grand-total">
        <span>Total Amount:</span>
        <span class="font-mono">${formatMoney(totalAmount)}</span>
      </div>
      ${amountPaid > 0 ? `
      <div class="inv-total-row">
        <span>Amount Paid:</span>
        <span class="font-mono text-emerald-600">-${formatMoney(amountPaid)}</span>
      </div>
      ` : ''}
      <div class="inv-total-row balance-due">
        <span>Balance Due:</span>
        <span class="font-mono">${formatMoney(balanceDue)}</span>
      </div>
    </div>
  </div>

  <!-- Footer Signatory -->
  <div class="inv-footer">
    <div>
      Thank you for your business.<br>
      Freel Global Logistics Operating System • All transactions subject to standard FIATA trading terms.
    </div>
    <div class="inv-sign-block">
      <div class="inv-sign-line"></div>
      <strong>Authorized Signatory</strong><br>
      <span>Finance & Commercial Collections</span>
    </div>
  </div>
</body>
</html>`;

  // Open printable document window
  const printWindow = window.open('', '_blank');
  if (printWindow) {
    printWindow.document.open();
    printWindow.document.write(html);
    printWindow.document.close();
    printWindow.focus();
    setTimeout(() => {
      printWindow.print();
    }, 400);
  } else {
    // If popups blocked, download as HTML file blob
    const blob = new Blob([html], { type: 'text/html' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `Invoice_${inv.invoiceNumber || inv.invoice_number || 'document'}.html`;
    document.body.appendChild(a);
    a.click();
    a.remove();
    URL.revokeObjectURL(url);
  }
}

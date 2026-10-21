/**
 * debitNotePdfGenerator.js
 * Authoritative Client-Side Commercial Debit Note Document & PDF Generator
 * Renders an official, customer-ready Debit Note formatted for A4 PDF export.
 */

export function generateDebitNotePDF(dn) {
  if (!dn) return;

  const currency = dn.currency || 'USD';
  const formatMoney = (val) =>
    `${currency} ${Number(val || 0).toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`;

  const formatDate = (d) => {
    if (!d) return '—';
    return new Date(d).toLocaleDateString('en-US', { day: '2-digit', month: 'short', year: 'numeric' });
  };

  const lineItems = dn.line_items && dn.line_items.length > 0 ? dn.line_items : [
    {
      description: dn.reason || 'Ancillary freight adjustment & demurrage/detention charge',
      service_category: 'Ancillary Adjustment',
      quantity: 1,
      unit_price: dn.total_amount || 0,
      total_amount: dn.total_amount || 0,
    }
  ];

  const html = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>Debit Note ${dn.debit_note_number || 'DN'}</title>
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
    .dn-header {
      display: flex;
      justify-content: space-between;
      align-items: flex-start;
      border-bottom: 2px solid #2563eb;
      padding-bottom: 18px;
      margin-bottom: 20px;
    }
    .dn-logo-title {
      font-size: 20px;
      font-weight: 800;
      color: #0f172a;
      letter-spacing: -0.02em;
    }
    .dn-logo-title span {
      color: #2563eb;
    }
    .dn-logo-sub {
      font-size: 11px;
      color: #64748b;
      margin-top: 3px;
      text-transform: uppercase;
      letter-spacing: 0.05em;
    }
    .dn-type-block {
      text-align: right;
    }
    .dn-type-badge {
      display: inline-block;
      font-size: 18px;
      font-weight: 800;
      color: #2563eb;
      text-transform: uppercase;
      letter-spacing: 0.05em;
    }
    .dn-number-text {
      font-family: ui-monospace, SFMono-Regular, monospace;
      font-size: 14px;
      font-weight: 700;
      color: #0f172a;
      margin-top: 4px;
    }
    .dn-status-tag {
      display: inline-block;
      font-size: 10px;
      font-weight: 800;
      padding: 2px 8px;
      border-radius: 4px;
      margin-top: 5px;
      text-transform: uppercase;
      background: ${dn.status === 'VOID' ? '#fef2f2' : dn.status === 'ISSUED' ? '#eff6ff' : '#f0fdf4'};
      color: ${dn.status === 'VOID' ? '#b91c1c' : dn.status === 'ISSUED' ? '#1d4ed8' : '#15803d'};
      border: 1px solid ${dn.status === 'VOID' ? '#fecaca' : dn.status === 'ISSUED' ? '#bfdbfe' : '#bbf7d0'};
    }

    .dn-parties-grid {
      display: grid;
      grid-template-columns: 1fr 1fr;
      gap: 20px;
      margin-bottom: 22px;
    }
    .dn-party-card {
      background: #f8fafc;
      border: 1px solid #e2e8f0;
      border-radius: 8px;
      padding: 14px 16px;
    }
    .dn-party-label {
      font-size: 10px;
      font-weight: 700;
      text-transform: uppercase;
      color: #64748b;
      letter-spacing: 0.06em;
      margin-bottom: 6px;
    }
    .dn-party-name {
      font-size: 15px;
      font-weight: 700;
      color: #0f172a;
      margin-bottom: 4px;
    }
    .dn-party-meta {
      font-size: 12px;
      color: #475569;
    }

    .dn-references-strip {
      display: grid;
      grid-template-columns: repeat(4, 1fr);
      gap: 12px;
      background: #f1f5f9;
      border: 1px solid #cbd5e1;
      border-radius: 8px;
      padding: 10px 14px;
      margin-bottom: 22px;
    }
    .dn-ref-item {
      display: flex;
      flex-direction: column;
    }
    .dn-ref-label {
      font-size: 10px;
      font-weight: 700;
      text-transform: uppercase;
      color: #64748b;
    }
    .dn-ref-value {
      font-size: 12px;
      font-weight: 600;
      color: #0f172a;
      font-family: ui-monospace, SFMono-Regular, monospace;
      margin-top: 2px;
    }

    .dn-reason-box {
      background: #fffbeb;
      border: 1px solid #fde68a;
      border-radius: 8px;
      padding: 10px 14px;
      margin-bottom: 22px;
    }
    .dn-reason-label {
      font-size: 11px;
      font-weight: 700;
      color: #92400e;
      text-transform: uppercase;
      letter-spacing: 0.04em;
    }
    .dn-reason-text {
      font-size: 12.5px;
      color: #78350f;
      margin-top: 4px;
      line-height: 1.4;
    }

    .dn-table {
      width: 100%;
      border-collapse: collapse;
      margin-bottom: 22px;
      border: 1px solid #e2e8f0;
      border-radius: 8px;
      overflow: hidden;
    }
    .dn-table th {
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
    .dn-table td {
      padding: 11px 12px;
      font-size: 12.5px;
      border-bottom: 1px solid #f1f5f9;
      color: #1e293b;
    }
    .dn-table tbody tr:last-child td {
      border-bottom: none;
    }
    .text-center { text-align: center; }
    .text-right { text-align: right; }
    .font-bold { font-weight: 700; }
    .font-mono { font-family: ui-monospace, SFMono-Regular, monospace; }

    .dn-totals-wrap {
      display: flex;
      justify-content: flex-end;
      margin-bottom: 24px;
    }
    .dn-totals-box {
      width: 280px;
      background: #f8fafc;
      border: 1px solid #e2e8f0;
      border-radius: 8px;
      padding: 12px 16px;
    }
    .dn-totals-row {
      display: flex;
      justify-content: space-between;
      align-items: center;
      padding: 4px 0;
      font-size: 12.5px;
    }
    .dn-totals-row.grand-total {
      border-top: 1px solid #cbd5e1;
      margin-top: 6px;
      padding-top: 8px;
      font-size: 15px;
      font-weight: 800;
      color: #0f172a;
    }

    .dn-footer-grid {
      display: grid;
      grid-template-columns: 1.5fr 1fr;
      gap: 20px;
      border-top: 1px solid #e2e8f0;
      padding-top: 16px;
      margin-top: 30px;
      font-size: 11px;
      color: #64748b;
    }
    .dn-remit-title {
      font-weight: 700;
      color: #0f172a;
      text-transform: uppercase;
      margin-bottom: 4px;
    }
    .dn-sig-box {
      text-align: right;
      display: flex;
      flex-direction: column;
      align-items: flex-end;
      justify-content: flex-end;
    }
    .dn-sig-line {
      width: 180px;
      border-bottom: 1px solid #0f172a;
      margin-bottom: 6px;
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
  <div class="dn-header">
    <div>
      <div class="dn-logo-title">FREEL <span>LOGISTICS</span></div>
      <div class="dn-logo-sub">Global Freight Operations & Ancillary Adjustments</div>
    </div>
    <div class="dn-type-block">
      <div class="dn-type-badge">DEBIT NOTE</div>
      <div class="dn-number-text">${dn.debit_note_number || 'DN-DRAFT'}</div>
      <span class="dn-status-tag">${dn.status || 'ISSUED'}</span>
    </div>
  </div>

  <!-- Parties -->
  <div class="dn-parties-grid">
    <div class="dn-party-card">
      <div class="dn-party-label">Billed To (Customer / Debtor)</div>
      <div class="dn-party-name">${dn.customer_name || 'Valued Customer'}</div>
      <div class="dn-party-meta">
        ${dn.customer_country ? `Country: ${dn.customer_country}` : 'International Commercial Account'}<br>
        Direct Customer Ledger Reference
      </div>
    </div>

    <div class="dn-party-card">
      <div class="dn-party-label">Issued By (Freight Forwarder)</div>
      <div class="dn-party-name">Freel Global Logistics Pvt. Ltd.</div>
      <div class="dn-party-meta">
        Corporate Billing & Collections Department<br>
        GST / Tax ID: 27AABCF1234F1Z8
      </div>
    </div>
  </div>

  <!-- Operational References Strip -->
  <div class="dn-references-strip">
    <div class="dn-ref-item">
      <span class="dn-ref-label">Issue Date</span>
      <span class="dn-ref-value">${formatDate(dn.issue_date)}</span>
    </div>
    <div class="dn-ref-item">
      <span class="dn-ref-label">Due Date</span>
      <span class="dn-ref-value">${formatDate(dn.due_date)}</span>
    </div>
    <div class="dn-ref-item">
      <span class="dn-ref-label">Linked Invoice</span>
      <span class="dn-ref-value">${dn.invoice_number || '—'}</span>
    </div>
    <div class="dn-ref-item">
      <span class="dn-ref-label">Shipment File</span>
      <span class="dn-ref-value">${dn.shipment_number || '—'}</span>
    </div>
  </div>

  <!-- Commercial Reason -->
  ${dn.reason ? `
  <div class="dn-reason-box">
    <div class="dn-reason-label">Adjustment Justification / Reason:</div>
    <div class="dn-reason-text">${dn.reason}</div>
  </div>
  ` : ''}

  <!-- Items Table -->
  <table class="dn-table">
    <thead>
      <tr>
        <th style="width: 45%;">Service / Charge Description</th>
        <th style="width: 20%;">Category</th>
        <th class="text-center" style="width: 10%;">Qty</th>
        <th class="text-right" style="width: 12%;">Unit Rate</th>
        <th class="text-right" style="width: 13%;">Total Amount</th>
      </tr>
    </thead>
    <tbody>
      ${lineItems.map(item => `
        <tr>
          <td class="font-bold">${item.description}</td>
          <td>${item.service_category || 'Demurrage / Ancillary'}</td>
          <td class="text-center font-mono">${item.quantity || 1}</td>
          <td class="text-right font-mono">${formatMoney(item.unit_price || item.total_amount)}</td>
          <td class="text-right font-bold font-mono">${formatMoney(item.total_amount)}</td>
        </tr>
      `).join('')}
    </tbody>
  </table>

  <!-- Totals -->
  <div class="dn-totals-wrap">
    <div class="dn-totals-box">
      <div class="dn-totals-row">
        <span>Subtotal:</span>
        <span class="font-mono">${formatMoney(dn.subtotal || dn.total_amount)}</span>
      </div>
      ${dn.tax_amount > 0 ? `
      <div class="dn-totals-row">
        <span>Tax / VAT:</span>
        <span class="font-mono">${formatMoney(dn.tax_amount)}</span>
      </div>
      ` : ''}
      <div class="dn-totals-row grand-total">
        <span>Total Payable:</span>
        <span class="font-mono">${formatMoney(dn.total_amount)}</span>
      </div>
    </div>
  </div>

  <!-- Footer & Remittance -->
  <div class="dn-footer-grid">
    <div>
      <div class="dn-remit-title">Payment & Settlement Remittance</div>
      Please remit full payment within terms to:<br>
      <strong>Standard Chartered Bank</strong> • A/C No: 9812-4421-9901<br>
      SWIFT / BIC: SCBLINBBXXX • Currency: ${currency}<br>
      Please quote debit note number <strong>${dn.debit_note_number}</strong> on wire advice.
    </div>
    <div class="dn-sig-box">
      <div class="dn-sig-line"></div>
      <strong>Authorized Signatory</strong>
      <span>Freel Global Logistics Billing Team</span>
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
    a.download = `Debit_Note_${dn.debit_note_number || 'document'}.html`;
    document.body.appendChild(a);
    a.click();
    a.remove();
    URL.revokeObjectURL(url);
  }
}

import React, { useState, useEffect } from 'react';

export default function FinanceWorkspace({ shipmentId, token, onRefreshShipment }) {
  const [invoices, setInvoices] = useState([]);
  const [discrepancies, setDiscrepancies] = useState([]);
  const [shipment, setShipment] = useState(null);
  const [quote, setQuote] = useState(null);
  const [uploading, setUploading] = useState(false);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [invoiceNo, setInvoiceNo] = useState('');
  const [vendorName, setVendorName] = useState('Maersk');

  useEffect(() => {
    fetchFinanceWorkspace();
  }, [shipmentId]);

  const fetchFinanceWorkspace = async () => {
    try {
      setLoading(true);
      const response = await fetch(`/api/v1/shipments/${shipmentId}/finance`, {
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        }
      });
      if (!response.ok) {
        throw new Error(`HTTP error ${response.status}`);
      }
      const data = await response.json();
      setInvoices(data.data.invoices || []);
      setDiscrepancies(data.data.discrepancies || []);
      setShipment(data.data.shipment || null);
      setQuote(data.data.quote || null);
      setError('');
    } catch (err) {
      console.error('Failed to fetch finance workspace:', err);
      setError('Could not load finance reconciliation workspace.');
    } finally {
      setLoading(false);
    }
  };

  const handleInvoiceUpload = async (e) => {
    e.preventDefault();
    if (!invoiceNo) {
      setError('Please provide an invoice number.');
      return;
    }

    try {
      setUploading(true);
      setError('');
      
      const uploadPayload = {
        invoice_number: invoiceNo,
        vendor_name: vendorName,
        s3_key: `uploads/invoices/${shipmentId}/inv_${invoiceNo.toLowerCase()}_${Date.now()}.pdf`,
        file_name: `invoice_${invoiceNo}.pdf`
      };

      const response = await fetch(`/api/v1/shipments/${shipmentId}/finance/invoices/upload`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify(uploadPayload)
      });

      if (!response.ok) {
        throw new Error(`Upload failed: ${response.statusText}`);
      }

      setInvoiceNo('');
      await fetchFinanceWorkspace();
      if (onRefreshShipment) onRefreshShipment();
      
      let count = 0;
      const interval = setInterval(async () => {
        count++;
        await fetchFinanceWorkspace();
        if (onRefreshShipment) onRefreshShipment();
        if (count >= 3) clearInterval(interval);
      }, 3000);

    } catch (err) {
      console.error('Invoice upload failed:', err);
      setError('Failed to upload and reconcile carrier invoice.');
    } finally {
      setUploading(false);
    }
  };

  const handleResolveDiscrepancy = async (discId) => {
    try {
      const response = await fetch(`/api/v1/finance/discrepancies/${discId}/resolve`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        }
      });
      if (response.ok) {
        fetchFinanceWorkspace();
        if (onRefreshShipment) onRefreshShipment();
      } else {
        alert('Failed to resolve financial discrepancy.');
      }
    } catch (err) {
      console.error('Error resolving discrepancy:', err);
    }
  };

  const handleApproveInvoice = async (invoiceId) => {
    try {
      const response = await fetch(`/api/v1/finance/invoices/${invoiceId}/approve`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        }
      });
      if (response.ok) {
        fetchFinanceWorkspace();
        if (onRefreshShipment) onRefreshShipment();
      } else {
        alert('Failed to manually approve invoice.');
      }
    } catch (err) {
      console.error('Error approving invoice:', err);
    }
  };

  if (loading && invoices.length === 0) {
    return <div className="workspace-loading">Loading financial reconciliation context...</div>;
  }

  // Calculate comparison statistics
  const quotedBuy = quote ? quote.buy_price : 0;
  const invoicedTotal = invoices.reduce((sum, inv) => sum + (inv.total_amount || 0), 0);
  const variance = invoicedTotal - quotedBuy;

  return (
    <div className="glass-card panel-card doc-workspace finance-workspace">
      <div className="workspace-header">
        <div>
          <h3>Finance & Three-Way Reconciliation</h3>
          <p className="panel-desc">Ingest carrier invoices, audit line items against agreed buy quotes, and manage exceptions.</p>
        </div>
        
        <form onSubmit={handleInvoiceUpload} className="invoice-upload-form">
          <input 
            type="text" 
            placeholder="Invoice Number" 
            value={invoiceNo}
            onChange={(e) => setInvoiceNo(e.target.value)}
            className="input-text invoice-no-input"
            required
            disabled={uploading}
          />
          <select 
            value={vendorName} 
            onChange={(e) => setVendorName(e.target.value)}
            className="doc-type-select vendor-select"
            disabled={uploading}
          >
            <option value="Maersk">Maersk (MAEU)</option>
            <option value="MSC">MSC (MSCU)</option>
            <option value="CMA CGM">CMA CGM (CMDU)</option>
            <option value="ONE">ONE (ONEY)</option>
            <option value="Hapag-Lloyd">Hapag-Lloyd (HLCU)</option>
          </select>
          <button type="submit" className={`btn-primary ${uploading ? 'disabled' : ''}`} disabled={uploading}>
            {uploading ? 'Auditing...' : 'Upload Carrier Invoice'}
          </button>
        </form>
      </div>

      {error && <div className="error-banner">{error}</div>}

      {/* Quote buy price vs actual invoice comparison */}
      {quote && (
        <div className="finance-comparison-summary-banner">
          <div className="summary-stat">
            <span className="stat-label">Quoted Buy Price</span>
            <span className="stat-val">${quotedBuy.toFixed(2)}</span>
          </div>
          <div className="summary-stat-divider">➔</div>
          <div className="summary-stat">
            <span className="stat-label">Total Invoiced Amount</span>
            <span className="stat-val">${invoicedTotal.toFixed(2)}</span>
          </div>
          <div className="summary-stat-divider">➔</div>
          <div className="summary-stat">
            <span className="stat-label">Computed Variance</span>
            <span className={`stat-val ${variance > 0 ? 'text-red' : variance < 0 ? 'text-green' : ''}`}>
              {variance > 0 ? `+$${variance.toFixed(2)}` : variance < 0 ? `-$${Math.abs(variance).toFixed(2)}` : '$0.00'}
            </span>
          </div>
        </div>
      )}

      <div className="workspace-split">
        {/* Invoices list */}
        <div className="docs-list-side">
          <h4>Carrier Invoices ({invoices.length})</h4>
          {invoices.length === 0 ? (
            <p className="empty-workspace-txt">No carrier invoices ingested. Upload a freight invoice to start audit reconciliation.</p>
          ) : (
            <div className="docs-grid-list">
              {invoices.map((inv) => (
                <div key={inv.id} className="doc-item-card invoice-item-card">
                  <div className="doc-meta-row">
                    <span className="doc-badge-type">{inv.vendor_name}</span>
                    <div className="invoice-status-controls">
                      <span className={`doc-status-pill status-${inv.status.toLowerCase()}`}>
                        {inv.status.replace('_', ' ')}
                      </span>
                      {inv.status !== 'APPROVED' && (
                        <button 
                          className="btn-approve-invoice"
                          onClick={() => handleApproveInvoice(inv.id)}
                        >
                          Manually Approve
                        </button>
                      )}
                    </div>
                  </div>
                  <h5 className="doc-filename">No: {inv.invoice_number}</h5>
                  {inv.ai_summary && (
                    <div className="invoice-ai-narrative">
                      <span className="narrative-badge">AI Analysis</span>
                      <p className="narrative-text">{inv.ai_summary}</p>
                    </div>
                  )}
                  
                  {inv.total_amount > 0 && (
                    <div className="invoice-grand-total">
                      Grand Total: <strong>{inv.total_amount.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })} {inv.currency}</strong>
                    </div>
                  )}
                </div>
              ))}
            </div>
          )}
        </div>

        {/* Discrepancies listing */}
        <div className="discrepancies-side">
          <h4>Finance Discrepancies ({discrepancies.length})</h4>
          {discrepancies.length === 0 ? (
            <div className="no-mismatches-box">
              <span className="clean-audit-icon">🪙</span>
              <p>Clean billing audit. All invoiced items perfectly match target contracts and quotes.</p>
            </div>
          ) : (
            <div className="discrepancies-grid">
              {discrepancies.map((disc) => (
                <div key={disc.id} className={`discrepancy-card status-${disc.status.toLowerCase()}`}>
                  <div className="discrepancy-header-row">
                    <span className="disc-field">Code: <strong>{disc.charge_code}</strong> ({disc.field_name})</span>
                    {disc.status === 'OPEN' ? (
                      <button 
                        className="btn-resolve-disc" 
                        onClick={() => handleResolveDiscrepancy(disc.id)}
                      >
                        Approve Variance
                      </button>
                    ) : (
                      <span className="disc-resolved-txt">✓ Variance Approved</span>
                    )}
                  </div>
                  <div className="disc-comparison-grid">
                    <div className="compare-block">
                      <small className="block-label">{disc.source} (Target)</small>
                      <span className="compare-val">${parseFloat(disc.expected_value).toFixed(2)}</span>
                    </div>
                    <div className="compare-arrow">➔</div>
                    <div className="compare-block">
                      <small className="block-label">Invoice (Actual)</small>
                      <span className="compare-val val-mismatch">${parseFloat(disc.actual_value).toFixed(2)}</span>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}


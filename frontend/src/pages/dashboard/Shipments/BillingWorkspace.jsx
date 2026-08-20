import React, { useState, useEffect } from 'react';
import './Shipments.css';

export default function BillingWorkspace({ shipmentId, token, onRefreshShipment }) {
  const [invoices, setInvoices] = useState([]);
  const [profitability, setProfitability] = useState(null);
  const [closureAudit, setClosureAudit] = useState(null);
  const [loading, setLoading] = useState(false);
  const [actionLoading, setActionLoading] = useState(false);
  const [error, setError] = useState(null);

  useEffect(() => {
    if (shipmentId) {
      fetchBillingWorkspace();
    }
  }, [shipmentId]);

  const fetchBillingWorkspace = async () => {
    setLoading(true);
    setError(null);
    try {
      const resp = await fetch(`/api/v1/shipments/${shipmentId}/billing`, {
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        }
      });
      const data = await resp.json();
      if (resp.ok && data.success) {
        setInvoices(data.data.invoices || []);
        setProfitability(data.data.profitability);
        setClosureAudit(data.data.closure_audit);
      } else {
        setError(data.message || 'Failed to load billing workspace');
      }
    } catch (err) {
      setError(err.message || 'Connection error to billing service');
    } finally {
      setLoading(false);
    }
  };

  const handleGenerateInvoice = async () => {
    setActionLoading(true);
    try {
      const resp = await fetch(`/api/v1/shipments/${shipmentId}/billing/invoices/generate`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        }
      });
      const data = await resp.json();
      if (resp.ok && data.success) {
        await fetchBillingWorkspace();
      } else {
        alert(data.message || 'Failed to generate customer invoice');
      }
    } catch (err) {
      alert('Error connecting to billing API');
    } finally {
      setActionLoading(false);
    }
  };

  const handleApproveInvoice = async (invoiceId) => {
    setActionLoading(true);
    try {
      const resp = await fetch(`/api/v1/billing/invoices/${invoiceId}/approve`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        }
      });
      const data = await resp.json();
      if (resp.ok && data.success) {
        await fetchBillingWorkspace();
      } else {
        alert(data.message || 'Failed to approve customer invoice');
      }
    } catch (err) {
      alert('Error connecting to billing API');
    } finally {
      setActionLoading(false);
    }
  };

  const handlePayInvoice = async (invoiceId) => {
    setActionLoading(true);
    try {
      const resp = await fetch(`/api/v1/billing/invoices/${invoiceId}/pay`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        }
      });
      const data = await resp.json();
      if (resp.ok && data.success) {
        await fetchBillingWorkspace();
      } else {
        alert(data.message || 'Failed to process payment');
      }
    } catch (err) {
      alert('Error connecting to billing API');
    } finally {
      setActionLoading(false);
    }
  };

  const handleCloseShipment = async () => {
    setActionLoading(true);
    try {
      const resp = await fetch(`/api/v1/shipments/${shipmentId}/close`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        }
      });
      const data = await resp.json();
      if (resp.ok && data.success) {
        alert('Shipment successfully closed!');
        onRefreshShipment();
        await fetchBillingWorkspace();
      } else {
        alert(data.message || 'Failed to close shipment');
      }
    } catch (err) {
      alert('Error connecting to shipment closure API');
    } finally {
      setActionLoading(false);
    }
  };

  const getStatusBadgeClass = (status) => {
    switch (status) {
      case 'DRAFT': return 'status-badge draft';
      case 'APPROVED': return 'status-badge approved';
      case 'SENT': return 'status-badge sent';
      case 'PAID': return 'status-badge paid';
      case 'OVERDUE': return 'status-badge overdue';
      default: return 'status-badge';
    }
  };

  const getProfitabilityBadgeClass = (status) => {
    switch (status) {
      case 'ON_TARGET': return 'margin-badge on-target';
      case 'UNDER_TARGET': return 'margin-badge under-target';
      case 'NEGATIVE_MARGIN': return 'margin-badge negative-margin';
      default: return 'margin-badge pending';
    }
  };

  return (
    <div className="glass-card panel-card billing-workspace-card">
      <div className="card-header">
        <h3 className="section-title">💼 Billing & Closure Workspace</h3>
        <button 
          className="refresh-btn" 
          onClick={fetchBillingWorkspace} 
          disabled={loading || actionLoading}
        >
          {loading ? '🔄 Loading...' : '🔄 Refresh'}
        </button>
      </div>

      {error && <div className="error-banner">{error}</div>}

      {/* 1. Profitability Variance Analytics */}
      {profitability && (
        <div className="profitability-section">
          <div className="profitability-header">
            <h4>📊 Shipment Profitability Margins</h4>
            <span className={getProfitabilityBadgeClass(profitability.profitability_status)}>
              Status: {profitability.profitability_status?.replace('_', ' ')}
            </span>
          </div>

          <div className="profitability-grid">
            <div className="metric-box">
              <span className="metric-label">Quoted Expected Profit</span>
              <span className="metric-value font-mono">${profitability.expected_profit?.toFixed(2)}</span>
              <span className="metric-sub">
                Sell: ${profitability.quoted_sell_price?.toFixed(2)} | Buy: ${profitability.quoted_buy_price?.toFixed(2)}
              </span>
              <span className="metric-sub text-green">
                Margin: {profitability.expected_margin_pct?.toFixed(1)}%
              </span>
            </div>

            <div className="metric-box">
              <span className="metric-label">Actual Audited Profit</span>
              <span className="metric-value font-mono">${profitability.actual_profit?.toFixed(2)}</span>
              <span className="metric-sub">
                Revenue: ${profitability.actual_revenue?.toFixed(2)} | Cost: ${profitability.actual_carrier_cost?.toFixed(2)}
              </span>
              <span className="metric-sub text-blue">
                Margin: {profitability.actual_margin_pct?.toFixed(1)}%
              </span>
            </div>

            <div className="metric-box variance-box">
              <span className="metric-label">Variance (Actual vs Expected)</span>
              <span className={`metric-value font-mono ${profitability.variance >= 0 ? 'text-green' : 'text-red'}`}>
                {profitability.variance >= 0 ? '+' : ''}${profitability.variance?.toFixed(2)}
              </span>
              <span className="metric-sub">Difference in bottom-line revenue</span>
            </div>
          </div>
        </div>
      )}

      {/* 2. Customer Invoices Management */}
      <div className="billing-list-section">
        <div className="section-header-row">
          <h4>📑 Customer Invoices</h4>
          {invoices.length === 0 && (
            <button 
              className="action-btn primary-btn"
              onClick={handleGenerateInvoice}
              disabled={actionLoading}
            >
              Generate Bill from Quote
            </button>
          )}
        </div>

        {invoices.length === 0 ? (
          <div className="empty-state">
            <p>No customer invoices generated yet for this shipment.</p>
          </div>
        ) : (
          <div className="table-wrapper">
            <table className="workspace-table">
              <thead>
                <tr>
                  <th>Invoice No.</th>
                  <th>Status</th>
                  <th>Amount</th>
                  <th>Due Date</th>
                  <th>Actions</th>
                </tr>
              </thead>
              <tbody>
                {invoices.map((inv) => (
                  <tr key={inv.id}>
                    <td className="font-mono">{inv.invoice_number}</td>
                    <td>
                      <span className={getStatusBadgeClass(inv.status)}>{inv.status}</span>
                    </td>
                    <td className="font-mono">${inv.total_amount?.toFixed(2)}</td>
                    <td>{inv.due_date ? new Date(inv.due_date).toLocaleDateString() : 'N/A'}</td>
                    <td className="actions-cell">
                      {inv.status === 'DRAFT' && (
                        <button 
                          className="table-btn approve-btn"
                          onClick={() => handleApproveInvoice(inv.id)}
                          disabled={actionLoading}
                        >
                          Approve
                        </button>
                      )}
                      {inv.status === 'APPROVED' && (
                        <button 
                          className="table-btn pay-btn"
                          onClick={() => handlePayInvoice(inv.id)}
                          disabled={actionLoading}
                        >
                          Mark Paid
                        </button>
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {/* 3. Shipment Closure Checklist */}
      {closureAudit && (
        <div className="closure-section">
          <h4>🔒 Shipment Closure Audit Checklist</h4>
          <div className="checklist-box">
            {closureAudit.checks?.map((check, idx) => (
              <div key={idx} className="checklist-item">
                <span className={`check-icon ${check.passed ? 'passed' : 'failed'}`}>
                  {check.passed ? '✅' : '❌'}
                </span>
                <div className="check-details">
                  <span className="check-title">{check.rule_name}</span>
                  <p className="check-desc">{check.description}</p>
                </div>
              </div>
            ))}
          </div>

          <div className="closure-action-row">
            <button 
              className={`close-shipment-btn ${closureAudit.ready ? 'ready' : 'blocked'}`}
              onClick={handleCloseShipment}
              disabled={!closureAudit.ready || actionLoading}
            >
              {actionLoading ? 'Closing...' : 'Close Shipment & Archive'}
            </button>
            {!closureAudit.ready && (
              <span className="closure-tooltip">
                ⚠️ All closure checklist rules must pass before archiving this shipment.
              </span>
            )}
          </div>
        </div>
      )}
    </div>
  );
}

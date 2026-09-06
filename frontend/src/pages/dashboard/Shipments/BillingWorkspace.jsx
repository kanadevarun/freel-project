import React, { useState, useEffect } from 'react';
import { RotateCw, CheckCircle2, XCircle, AlertCircle, FileText, DollarSign, Lock } from 'lucide-react';
import api from '../../../services/api';
import './Shipments.css';

export default function BillingWorkspace({ shipmentId, onRefreshShipment }) {
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
      const data = await api.get(`/api/v1/shipments/${shipmentId}/billing`);
      setInvoices(data?.invoices || []);
      setProfitability(data?.profitability || null);
      setClosureAudit(data?.closure_audit || null);
    } catch (err) {
      // Graceful fallback for uninitialized billing workspaces
      setInvoices([]);
      setProfitability(null);
      setClosureAudit(null);
      if (err.response?.status !== 404) {
        setError("We couldn't load billing information for this shipment. Please refresh and try again.");
      }
    } finally {
      setLoading(false);
    }
  };

  const handleGenerateInvoice = async () => {
    setActionLoading(true);
    try {
      await api.post(`/api/v1/shipments/${shipmentId}/billing/invoices/generate`);
      await fetchBillingWorkspace();
    } catch (err) {
      setError(err.message || 'Failed to generate customer invoice');
    } finally {
      setActionLoading(false);
    }
  };

  const handleApproveInvoice = async (invoiceId) => {
    setActionLoading(true);
    try {
      await api.post(`/api/v1/billing/invoices/${invoiceId}/approve`);
      await fetchBillingWorkspace();
    } catch (err) {
      setError(err.message || 'Failed to approve customer invoice');
    } finally {
      setActionLoading(false);
    }
  };

  const handlePayInvoice = async (invoiceId) => {
    setActionLoading(true);
    try {
      await api.post(`/api/v1/billing/invoices/${invoiceId}/pay`);
      await fetchBillingWorkspace();
    } catch (err) {
      setError(err.message || 'Failed to process payment');
    } finally {
      setActionLoading(false);
    }
  };

  const handleCloseShipment = async () => {
    setActionLoading(true);
    try {
      await api.post(`/api/v1/shipments/${shipmentId}/close`);
      if (onRefreshShipment) onRefreshShipment();
      await fetchBillingWorkspace();
    } catch (err) {
      setError(err.message || 'Failed to close shipment');
    } finally {
      setActionLoading(false);
    }
  };

  const getStatusBadgeStyle = (status) => {
    switch (status) {
      case 'PAID':
        return { bg: '#dcfce7', fg: '#15803d', border: '#bbf7d0' };
      case 'APPROVED':
        return { bg: '#dbeafe', fg: '#1d4ed8', border: '#bfdbfe' };
      case 'SENT':
        return { bg: '#fef3c7', fg: '#b45309', border: '#fde68a' };
      case 'OVERDUE':
        return { bg: '#fee2e2', fg: '#dc2626', border: '#fecaca' };
      case 'DRAFT':
      default:
        return { bg: '#f1f5f9', fg: '#475569', border: '#e2e8f0' };
    }
  };

  return (
    <div className="sd-panel">
      {/* Panel Header */}
      <div className="sd-panel-header">
        <div className="sd-panel-title-row">
          <div>
            <h3 className="sd-panel-title">Customer Billing & Invoicing</h3>
            <p className="sd-panel-desc">Customer receivable invoicing, profitability margin tracking, and financial closure audit.</p>
          </div>
          <button 
            className="sd-diagnostics-btn" 
            onClick={fetchBillingWorkspace} 
            disabled={loading || actionLoading}
            title="Refresh Billing Data"
          >
            <RotateCw size={12} className={loading ? 'sd-spin' : ''} />
            {loading ? 'Refreshing...' : 'Refresh'}
          </button>
        </div>
      </div>

      {error && (
        <div className="sd-error-banner" style={{ margin: '0 0 16px 0' }}>
          <AlertCircle size={14} />
          <span>{error}</span>
          <button className="sd-btn-ghost" onClick={fetchBillingWorkspace} style={{ marginLeft: 'auto', fontSize: '0.74rem' }}>
            Retry
          </button>
        </div>
      )}

      {/* 1. Profitability Variance Analytics */}
      {profitability && (
        <div className="sd-profitability-box" style={{ marginBottom: '16px', background: '#f8fafc', border: '1px solid #e2e8f0', borderRadius: '10px', padding: '14px' }}>
          <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: '12px' }}>
            <span style={{ fontSize: '0.78rem', fontWeight: 750, textTransform: 'uppercase', letterSpacing: '0.04em', color: '#475569' }}>
              📊 Shipment Profitability Margins
            </span>
            <span style={{ fontSize: '0.72rem', fontWeight: 800, padding: '2px 8px', borderRadius: '4px', background: profitability.profitability_status === 'ON_TARGET' ? '#dcfce7' : '#fef3c7', color: profitability.profitability_status === 'ON_TARGET' ? '#15803d' : '#b45309' }}>
              {profitability.profitability_status?.replace(/_/g, ' ') || 'Pending'}
            </span>
          </div>

          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(180px, 1fr))', gap: '10px' }}>
            <div style={{ background: '#ffffff', border: '1px solid #e2e8f0', borderRadius: '8px', padding: '10px 12px' }}>
              <span style={{ fontSize: '0.7rem', fontWeight: 700, color: '#64748b', textTransform: 'uppercase' }}>Quoted Expected</span>
              <div style={{ fontSize: '1.05rem', fontWeight: 800, color: '#0f172a', margin: '2px 0' }}>
                ${profitability.expected_profit?.toFixed(2) || '0.00'}
              </div>
              <small style={{ fontSize: '0.72rem', color: '#16a34a', fontWeight: 700 }}>
                Margin: {profitability.expected_margin_pct?.toFixed(1) || '0.0'}%
              </small>
            </div>

            <div style={{ background: '#ffffff', border: '1px solid #e2e8f0', borderRadius: '8px', padding: '10px 12px' }}>
              <span style={{ fontSize: '0.7rem', fontWeight: 700, color: '#64748b', textTransform: 'uppercase' }}>Actual Audited</span>
              <div style={{ fontSize: '1.05rem', fontWeight: 800, color: '#0f172a', margin: '2px 0' }}>
                ${profitability.actual_profit?.toFixed(2) || '0.00'}
              </div>
              <small style={{ fontSize: '0.72rem', color: '#2563eb', fontWeight: 700 }}>
                Margin: {profitability.actual_margin_pct?.toFixed(1) || '0.0'}%
              </small>
            </div>

            <div style={{ background: '#ffffff', border: '1px solid #e2e8f0', borderRadius: '8px', padding: '10px 12px' }}>
              <span style={{ fontSize: '0.7rem', fontWeight: 700, color: '#64748b', textTransform: 'uppercase' }}>Profit Variance</span>
              <div style={{ fontSize: '1.05rem', fontWeight: 800, color: (profitability.variance ?? 0) >= 0 ? '#16a34a' : '#dc2626', margin: '2px 0' }}>
                {(profitability.variance ?? 0) >= 0 ? '+' : ''}${profitability.variance?.toFixed(2) || '0.00'}
              </div>
              <small style={{ fontSize: '0.72rem', color: '#64748b' }}>
                Bottom-line variance
              </small>
            </div>
          </div>
        </div>
      )}

      {/* 2. Customer Invoices Management */}
      <div style={{ marginBottom: '16px' }}>
        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: '10px' }}>
          <span style={{ fontSize: '0.78rem', fontWeight: 750, textTransform: 'uppercase', letterSpacing: '0.04em', color: '#475569' }}>
            📑 Customer Invoices ({invoices.length})
          </span>
          {invoices.length === 0 && (
            <button 
              className="sd-action-btn sd-action-primary"
              style={{ padding: '5px 12px', fontSize: '0.76rem' }}
              onClick={handleGenerateInvoice}
              disabled={actionLoading}
            >
              Generate Bill from Quote
            </button>
          )}
        </div>

        {invoices.length === 0 ? (
          <div className="sd-panel-empty-box">
            <FileText size={24} className="sd-empty-icon-muted" />
            <span className="sd-empty-title">No customer invoices issued yet</span>
            <span className="sd-empty-desc">Invoices generated from confirmed customer quotations will appear here.</span>
          </div>
        ) : (
          <div style={{ overflowX: 'auto', border: '1px solid #e2e8f0', borderRadius: '8px' }}>
            <table style={{ width: '100%', borderCollapse: 'collapse', fontSize: '0.8rem', textAlign: 'left' }}>
              <thead>
                <tr style={{ background: '#f8fafc', borderBottom: '1px solid #e2e8f0', color: '#64748b', fontSize: '0.72rem', fontWeight: 750, textTransform: 'uppercase', letterSpacing: '0.04em' }}>
                  <th style={{ padding: '10px 12px' }}>Invoice No.</th>
                  <th style={{ padding: '10px 12px' }}>Status</th>
                  <th style={{ padding: '10px 12px' }}>Amount</th>
                  <th style={{ padding: '10px 12px' }}>Due Date</th>
                  <th style={{ padding: '10px 12px', textAlign: 'right' }}>Actions</th>
                </tr>
              </thead>
              <tbody>
                {invoices.map((inv) => {
                  const badge = getStatusBadgeStyle(inv.status);
                  return (
                    <tr key={inv.id} style={{ borderBottom: '1px solid #f1f5f9' }}>
                      <td style={{ padding: '10px 12px', fontWeight: 700, color: '#0f172a' }}>{inv.invoice_number}</td>
                      <td style={{ padding: '10px 12px' }}>
                        <span style={{ fontSize: '0.7rem', fontWeight: 800, padding: '2px 7px', borderRadius: '4px', background: badge.bg, color: badge.fg, border: `1px solid ${badge.border}` }}>
                          {inv.status}
                        </span>
                      </td>
                      <td style={{ padding: '10px 12px', fontWeight: 750, color: '#0f172a' }}>${inv.total_amount?.toFixed(2)}</td>
                      <td style={{ padding: '10px 12px', color: '#64748b' }}>{inv.due_date ? new Date(inv.due_date).toLocaleDateString() : '—'}</td>
                      <td style={{ padding: '10px 12px', textAlign: 'right' }}>
                        {inv.status === 'DRAFT' && (
                          <button 
                            className="sd-action-btn sd-action-primary"
                            style={{ padding: '3px 9px', fontSize: '0.72rem' }}
                            onClick={() => handleApproveInvoice(inv.id)}
                            disabled={actionLoading}
                          >
                            Approve
                          </button>
                        )}
                        {inv.status === 'APPROVED' && (
                          <button 
                            className="sd-action-btn sd-action-success"
                            style={{ padding: '3px 9px', fontSize: '0.72rem' }}
                            onClick={() => handlePayInvoice(inv.id)}
                            disabled={actionLoading}
                          >
                            Mark Paid
                          </button>
                        )}
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {/* 3. Shipment Closure Checklist */}
      {closureAudit && (
        <div style={{ background: '#f8fafc', border: '1px solid #e2e8f0', borderRadius: '10px', padding: '14px' }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: '6px', marginBottom: '10px' }}>
            <Lock size={14} style={{ color: '#475569' }} />
            <span style={{ fontSize: '0.78rem', fontWeight: 750, textTransform: 'uppercase', letterSpacing: '0.04em', color: '#475569' }}>
              Shipment Closure Readiness Audit
            </span>
          </div>

          <div style={{ display: 'flex', flexDirection: 'column', gap: '8px', marginBottom: '14px' }}>
            {closureAudit.checks?.map((check, idx) => (
              <div key={idx} style={{ display: 'flex', alignItems: 'flex-start', gap: '8px', background: '#ffffff', border: '1px solid #e2e8f0', borderRadius: '6px', padding: '8px 10px' }}>
                {check.passed ? (
                  <CheckCircle2 size={16} style={{ color: '#16a34a', flexShrink: 0, marginTop: '2px' }} />
                ) : (
                  <XCircle size={16} style={{ color: '#dc2626', flexShrink: 0, marginTop: '2px' }} />
                )}
                <div>
                  <div style={{ fontSize: '0.78rem', fontWeight: 700, color: '#0f172a' }}>{check.rule_name}</div>
                  <p style={{ fontSize: '0.72rem', color: '#64748b', margin: '2px 0 0 0' }}>{check.description}</p>
                </div>
              </div>
            ))}
          </div>

          <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', flexWrap: 'wrap', gap: '8px' }}>
            <button 
              className={`sd-action-btn ${closureAudit.ready ? 'sd-action-success' : 'sd-action-secondary'}`}
              style={{ padding: '7px 16px', fontSize: '0.78rem' }}
              onClick={handleCloseShipment}
              disabled={!closureAudit.ready || actionLoading}
            >
              {actionLoading ? 'Closing File...' : 'Close Shipment & Archive File'}
            </button>
            {!closureAudit.ready && (
              <span style={{ fontSize: '0.72rem', color: '#94a3b8', fontStyle: 'italic' }}>
                ⚠️ All audit requirements must pass before this shipment file can be closed.
              </span>
            )}
          </div>
        </div>
      )}
    </div>
  );
}

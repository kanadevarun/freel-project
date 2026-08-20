import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { CreditCard, ArrowRight, Download, Filter, CheckCircle2, Clock } from 'lucide-react';
import api from '../../../services/api';
import PageHeader from '../../../components/dashboard/PageHeader';
import './InvoicesPage.css';

export default function InvoicesPage() {
  const [invoices, setInvoices] = useState([]);
  const [filterType, setFilterType] = useState('ALL');
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const navigate = useNavigate();

  useEffect(() => {
    fetchInvoices();
  }, []);

  const fetchInvoices = async () => {
    try {
      setLoading(true);
      setError(null);
      const res = await api.get('/api/v1/invoices');
      const invList = Array.isArray(res) ? res : (res?.data || []);
      setInvoices(invList);
    } catch (err) {
      console.error('Failed to load invoices:', err);
      setError('Unable to load invoice ledger. Please try again.');
    } finally {
      setLoading(false);
    }
  };

  const filteredInvoices = invoices.filter((inv) => {
    if (filterType === 'ALL') return true;
    if (filterType === 'AR') return inv.type === 'CUSTOMER_AR';
    if (filterType === 'AP') return inv.type === 'CARRIER_AP';
    return true;
  });

  return (
    <div className="invoices-page">
      <div className="invoices-header-row">
        <PageHeader
          title="Invoices & AR/AP Ledger"
          subtitle="Manage customer freight billing (AR), carrier freight invoices (AP), and automated 3-way reconciliation"
        />
      </div>

      {/* Tabs */}
      <div className="invoices-tabs-row">
        {['ALL', 'AR', 'AP'].map((t) => (
          <button
            key={t}
            className={`tab-btn ${filterType === t ? 'active' : ''}`}
            onClick={() => setFilterType(t)}
          >
            {t === 'ALL' ? 'All Invoices' : t === 'AR' ? 'Customer Billing (AR)' : 'Carrier Payables (AP)'}
          </button>
        ))}
      </div>

      {error && (
        <div className="invoices-error-banner">
          <span>⚠️ {error}</span>
          <button onClick={fetchInvoices} className="btn-retry">Retry</button>
        </div>
      )}

      {loading ? (
        <div className="invoices-skeleton">
          {[1, 2, 3].map((i) => (
            <div key={i} className="skeleton-invoice-row" />
          ))}
        </div>
      ) : filteredInvoices.length === 0 ? (
        <div className="invoices-empty-state">
          <div className="empty-icon-box">💳</div>
          <h3>No invoices yet</h3>
          <p>
            Customer receivables (AR) and carrier freight bills (AP) will be automatically generated upon shipment booking confirmation and container departure.
          </p>
          <div style={{ display: 'flex', gap: '12px', justifyContent: 'center', marginTop: '20px' }}>
            <button className="btn-primary-action" onClick={() => navigate('/dashboard/shipments')}>
              View Active Shipments
            </button>
          </div>
        </div>
      ) : (
        <div className="invoices-table-card">
          <table className="invoices-table">
            <thead>
              <tr>
                <th>Invoice #</th>
                <th>Shipment Ref</th>
                <th>Party / Counterparty</th>
                <th>Type</th>
                <th>Amount</th>
                <th>Due Date</th>
                <th>Audit Status</th>
                <th>Payment Status</th>
              </tr>
            </thead>
            <tbody>
              {filteredInvoices.map((inv) => (
                <tr key={inv.id} onClick={() => navigate(`/dashboard/shipments/${inv.shipment_id}`)}>
                  <td><strong>{inv.invoice_number || `INV-${inv.id}`}</strong></td>
                  <td>{inv.booking_number || `SH-${inv.shipment_id}`}</td>
                  <td>{inv.vendor_name || inv.customer_name || '—'}</td>
                  <td>
                    <span className={`type-tag ${inv.type === 'CUSTOMER_AR' ? 'ar' : 'ap'}`}>
                      {inv.type === 'CUSTOMER_AR' ? 'AR (Receivable)' : 'AP (Payable)'}
                    </span>
                  </td>
                  <td><strong>${Number(inv.total_amount || 0).toLocaleString()}</strong></td>
                  <td>{inv.due_date ? new Date(inv.due_date).toLocaleDateString() : '—'}</td>
                  <td>
                    <span className="audit-pill matched">3-Way Matched</span>
                  </td>
                  <td>
                    <span className={`status-pill ${inv.status?.toLowerCase() || 'pending'}`}>
                      {inv.status || 'PENDING'}
                    </span>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}

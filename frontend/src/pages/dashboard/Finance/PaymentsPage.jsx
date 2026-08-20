import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { DollarSign, ArrowUpRight, ArrowDownLeft, ShieldCheck, Filter } from 'lucide-react';
import api from '../../../services/api';
import PageHeader from '../../../components/dashboard/PageHeader';
import './PaymentsPage.css';

export default function PaymentsPage() {
  const [payments, setPayments] = useState([]);
  const [filterType, setFilterType] = useState('ALL');
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const navigate = useNavigate();

  useEffect(() => {
    fetchPayments();
  }, []);

  const fetchPayments = async () => {
    try {
      setLoading(true);
      setError(null);
      const res = await api.get('/api/v1/invoices');
      const invoices = Array.isArray(res) ? res : (res?.data || []);
      const payList = invoices.filter(inv => inv.status === 'PAID' || inv.status === 'APPROVED');
      setPayments(payList);
    } catch (err) {
      console.error('Failed to load payments:', err);
      setError('Unable to load payment transactions. Please try again.');
    } finally {
      setLoading(false);
    }
  };

  const filteredPayments = payments.filter((p) => {
    if (filterType === 'ALL') return true;
    if (filterType === 'INBOUND') return p.direction === 'INBOUND';
    if (filterType === 'OUTBOUND') return p.direction === 'OUTBOUND';
    return true;
  });

  return (
    <div className="payments-page">
      <div className="payments-header-row">
        <PageHeader
          title="Payments & Freight Settlements"
          subtitle="Track inbound customer invoice settlements, outbound carrier freight disbursements, and escrow reconciliation"
        />
      </div>

      {/* Tabs */}
      <div className="payments-tabs-row">
        {['ALL', 'INBOUND', 'OUTBOUND'].map((t) => (
          <button
            key={t}
            className={`tab-btn ${filterType === t ? 'active' : ''}`}
            onClick={() => setFilterType(t)}
          >
            {t === 'ALL' ? 'All Transactions' : t === 'INBOUND' ? 'Customer Receipts (Inbound)' : 'Carrier Disbursements (Outbound)'}
          </button>
        ))}
      </div>

      {error && (
        <div className="payments-error-banner">
          <span>⚠️ {error}</span>
          <button onClick={fetchPayments} className="btn-retry">Retry</button>
        </div>
      )}

      {loading ? (
        <div className="payments-skeleton">
          {[1, 2, 3].map((i) => (
            <div key={i} className="skeleton-payment-row" />
          ))}
        </div>
      ) : filteredPayments.length === 0 ? (
        <div className="payments-empty-state">
          <div className="empty-icon-box">💰</div>
          <h3>No payment transactions yet</h3>
          <p>
            Customer payment receipts and carrier freight settlement records will appear here as invoices are paid and reconciled.
          </p>
          <div style={{ display: 'flex', gap: '12px', justifyContent: 'center', marginTop: '20px' }}>
            <button className="btn-primary-action" onClick={() => navigate('/dashboard/invoices')}>
              View Invoices
            </button>
          </div>
        </div>
      ) : (
        <div className="payments-table-card">
          <table className="payments-table">
            <thead>
              <tr>
                <th>Payment Ref</th>
                <th>Shipment / Invoice</th>
                <th>Counterparty</th>
                <th>Flow</th>
                <th>Amount</th>
                <th>Payment Method</th>
                <th>Settlement Status</th>
                <th>Date</th>
              </tr>
            </thead>
            <tbody>
              {filteredPayments.map((p) => (
                <tr key={p.id} onClick={() => navigate(`/dashboard/shipments/${p.shipment_id}`)}>
                  <td><strong>{p.payment_reference || `PAY-${p.id}`}</strong></td>
                  <td>{p.booking_number || `SH-${p.shipment_id}`}</td>
                  <td>{p.counterparty_name || '—'}</td>
                  <td>
                    <span className={`flow-tag ${p.direction === 'INBOUND' ? 'inbound' : 'outbound'}`}>
                      {p.direction === 'INBOUND' ? <ArrowDownLeft size={12} /> : <ArrowUpRight size={12} />}
                      {p.direction === 'INBOUND' ? 'Customer Inbound' : 'Carrier Outbound'}
                    </span>
                  </td>
                  <td><strong>${Number(p.amount || 0).toLocaleString()}</strong></td>
                  <td>{p.payment_method || 'Wire Transfer / ACH'}</td>
                  <td>
                    <span className={`status-pill ${p.status?.toLowerCase() || 'settled'}`}>
                      {p.status || 'SETTLED'}
                    </span>
                  </td>
                  <td>{p.created_at ? new Date(p.created_at).toLocaleDateString() : '—'}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}

import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { FileSpreadsheet, Plus, ArrowRight, ExternalLink, Filter } from 'lucide-react';
import api from '../../../services/api';
import PageHeader from '../../../components/dashboard/PageHeader';
import './QuotationsPage.css';

export default function QuotationsPage() {
  const [rfqs, setRfqs] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [filterStage, setFilterStage] = useState('ALL');
  const navigate = useNavigate();

  useEffect(() => {
    fetchQuotations();
  }, []);

  const fetchQuotations = async () => {
    try {
      setLoading(true);
      setError(null);
      const res = await api.get('/api/v1/rfqs');
      const data = Array.isArray(res) ? res : (res?.data?.rfqs || res?.rfqs || []);
      setRfqs(data);
    } catch (err) {
      console.error('Failed to load quotes:', err);
      setError('Unable to load quotations. Please try again.');
    } finally {
      setLoading(false);
    }
  };

  // Filter quotes from RFQs
  const quoteRows = rfqs.filter((r) => {
    if (filterStage === 'ALL') return true;
    return r.stage === filterStage;
  });

  const getStageBadge = (stage) => {
    switch (stage) {
      case 'STAGE_QUOTE_SENT':
        return <span className="stage-pill sent">Quote Sent</span>;
      case 'STAGE_WON':
        return <span className="stage-pill won">Won / Booked</span>;
      case 'STAGE_LOST':
        return <span className="stage-pill lost">Declined</span>;
      case 'STAGE_CARRIER_RATES_RECEIVED':
        return <span className="stage-pill draft">Carrier Rates In</span>;
      default:
        return <span className="stage-pill new">RFQ Created</span>;
    }
  };

  return (
    <div className="quotations-page">
      <div className="quotations-header-row">
        <PageHeader
          title="Quotations & Pricing"
          subtitle="Generate customer quotes, compare carrier margins, and track commercial win rates"
        />
        <button className="btn-create-quote" onClick={() => navigate('/dashboard/rfqs')}>
          <Plus size={16} />
          Create New Quote
        </button>
      </div>

      {/* Filter Tabs */}
      <div className="quote-filter-bar">
        {['ALL', 'STAGE_QUOTE_SENT', 'STAGE_WON', 'STAGE_CARRIER_RATES_RECEIVED', 'STAGE_LOST'].map((st) => (
          <button
            key={st}
            className={`filter-btn ${filterStage === st ? 'active' : ''}`}
            onClick={() => setFilterStage(st)}
          >
            {st === 'ALL'
              ? 'All Quotes'
              : st === 'STAGE_QUOTE_SENT'
              ? 'Quote Sent'
              : st === 'STAGE_WON'
              ? 'Won'
              : st === 'STAGE_CARRIER_RATES_RECEIVED'
              ? 'Drafts'
              : 'Declined'}
          </button>
        ))}
      </div>

      {error && (
        <div className="quote-error-banner">
          <span>⚠️ {error}</span>
          <button onClick={fetchQuotations} className="btn-retry">Retry</button>
        </div>
      )}

      {loading ? (
        <div className="quote-skeleton">
          {[1, 2, 3, 4].map((i) => (
            <div key={i} className="skeleton-quote-row" />
          ))}
        </div>
      ) : quoteRows.length === 0 ? (
        <div className="quote-empty-state">
          <div className="quote-empty-icon">📊</div>
          <h3>No quotations yet</h3>
          <p>
            Once you receive an RFQ, LogisticsHQ automatically compiles carrier rate options and generates customer-ready quotes.
          </p>
          <div style={{ display: 'flex', gap: '12px', justifyContent: 'center', marginTop: '20px' }}>
            <button className="btn-primary-action" onClick={() => navigate('/dashboard/rfqs')}>
              View Open RFQs
            </button>
          </div>
        </div>
      ) : (
        <div className="quote-table-card">
          <table className="quotations-table">
            <thead>
              <tr>
                <th>RFQ / Quote Ref</th>
                <th>Customer</th>
                <th>Route</th>
                <th>Cargo / Equip</th>
                <th>Status</th>
                <th>Created</th>
                <th>Action</th>
              </tr>
            </thead>
            <tbody>
              {quoteRows.map((q) => (
                <tr key={q.id} onClick={() => navigate(`/dashboard/rfqs`)}>
                  <td>
                    <strong>{q.rfq_number || `QT-${q.id}`}</strong>
                  </td>
                  <td>{q.customer_name || `Customer #${q.customer_id || '—'}`}</td>
                  <td>
                    <div className="route-cell">
                      <span>{q.origin || '—'}</span>
                      <ArrowRight size={13} className="route-arrow" />
                      <span>{q.destination || '—'}</span>
                    </div>
                  </td>
                  <td>
                    <span className="equipment-tag">{q.equipment_type || q.cargo_details || '40GP FCL'}</span>
                  </td>
                  <td>{getStageBadge(q.stage)}</td>
                  <td>{q.created_at ? new Date(q.created_at).toLocaleDateString() : '—'}</td>
                  <td>
                    <button className="btn-view-quote" onClick={(e) => { e.stopPropagation(); navigate('/dashboard/rfqs'); }}>
                      <span>View Pricing</span>
                      <ExternalLink size={12} />
                    </button>
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

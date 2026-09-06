import React, { useState, useEffect, useCallback } from 'react';
import {
  AlertTriangle,
  Clock,
  CheckCircle2,
  RefreshCw,
  FileText,
  Building2,
  Calendar,
  Layers,
  ArrowRight,
  TrendingDown,
  ShieldAlert,
  Flame,
  Check,
  ChevronRight,
  Filter,
  Eye,
  FileCheck
} from 'lucide-react';
import rateService from '../../../services/rateService';
import toast from 'react-hot-toast';
import './RateLifecycleWorkspace.css';

export default function RateLifecycleWorkspace({ onViewRate, onViewContract }) {
  const [summary, setSummary] = useState({
    total_rates: 0,
    active_rates: 0,
    expiring_soon_rates: 0,
    expired_rates: 0,
    superseded_rates: 0,
    total_contracts: 0,
    active_contracts: 0,
    expiring_contracts: 0,
    expired_contracts: 0,
    contracts_requiring_renewal: 0,
    quotations_at_risk: 0,
  });

  const [attentionBucket, setAttentionBucket] = useState('ALL'); // ALL, EXPIRED, EXPIRING_7D, EXPIRING_30D, SUPERSEDED
  const [attentionRates, setAttentionRates] = useState([]);
  const [attentionContracts, setAttentionContracts] = useState([]);
  const [lifecycleEvents, setLifecycleEvents] = useState([]);
  const [loading, setLoading] = useState(true);
  const [evaluating, setEvaluating] = useState(false);
  const [activeTab, setActiveTab] = useState('rates'); // rates, contracts, timeline

  const loadData = useCallback(async () => {
    setLoading(true);
    try {
      const [sumRes, ratesRes, contractsRes, eventsRes] = await Promise.all([
        rateService.getRateLifecycleSummary(),
        rateService.getRatesRequiringAttention(),
        rateService.getContractsRequiringAttention(),
        rateService.getRateLifecycleEvents(50),
      ]);

      setSummary(sumRes?.data || sumRes || {});
      setAttentionRates(ratesRes?.data?.rates || ratesRes?.rates || []);
      setAttentionContracts(contractsRes?.data?.contracts || contractsRes?.contracts || []);
      setLifecycleEvents(eventsRes?.data?.events || eventsRes?.events || []);
    } catch (err) {
      toast.error('Failed to load rate lifecycle intelligence data');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    loadData();
  }, [loadData]);

  const handleTriggerEvaluation = async () => {
    setEvaluating(true);
    try {
      await rateService.evaluateRateLifecycle();
      toast.success('Rate lifecycle evaluation completed successfully!');
      loadData();
    } catch (err) {
      toast.error(err?.message || 'Failed to trigger rate evaluation');
    } finally {
      setEvaluating(false);
    }
  };

  const filteredRates = attentionRates.filter((r) => {
    if (attentionBucket === 'ALL') return true;
    return r.attention_bucket === attentionBucket;
  });

  const getBucketBadge = (bucket) => {
    switch (bucket) {
      case 'EXPIRED':
        return <span className="rlc-badge rlc-badge-expired">Expired</span>;
      case 'EXPIRING_7D':
        return <span className="rlc-badge rlc-badge-critical">Expiring &le; 7 Days</span>;
      case 'EXPIRING_30D':
        return <span className="rlc-badge rlc-badge-warning">Expiring &le; 30 Days</span>;
      case 'SUPERSEDED':
        return <span className="rlc-badge rlc-badge-superseded">Superseded</span>;
      default:
        return <span className="rlc-badge rlc-badge-active">Active</span>;
    }
  };

  return (
    <div className="rlc-container">
      {/* ── Top Header Banner & Evaluation Trigger ── */}
      <div className="rlc-header">
        <div>
          <div className="rlc-header-title-row">
            <h1 className="rlc-header-title">Rate Lifecycle & Risk Intelligence</h1>
            <span className="rlc-pulse-badge">
              <span className="rlc-pulse-dot" /> Proactive Monitoring
            </span>
          </div>
          <p className="rlc-header-sub">
            Continuous commercial health monitoring for carrier rates, contracts, spot quotes, and active quotation snapshots.
          </p>
        </div>

        <button
          type="button"
          className="rlc-btn-evaluate"
          onClick={handleTriggerEvaluation}
          disabled={evaluating}
        >
          <RefreshCw size={15} className={evaluating ? 'rlc-spin' : ''} />
          {evaluating ? 'Evaluating...' : 'Run Lifecycle Evaluation'}
        </button>
      </div>

      {/* ── KPI Strip (6 Key Commercial Health Metrics) ── */}
      <div className="rlc-kpi-grid">
        <div className="rlc-kpi-card">
          <div className="rlc-kpi-top">
            <span className="rlc-kpi-label">Active Rates</span>
            <div className="rlc-kpi-icon-wrap rlc-icon-blue">
              <Layers size={18} />
            </div>
          </div>
          <div className="rlc-kpi-val-row">
            <span className="rlc-kpi-value">{summary.active_rates}</span>
            <span className="rlc-kpi-sub">of {summary.total_rates} total</span>
          </div>
        </div>

        <div className="rlc-kpi-card">
          <div className="rlc-kpi-top">
            <span className="rlc-kpi-label">Expiring Soon (&le;30d)</span>
            <div className="rlc-kpi-icon-wrap rlc-icon-amber">
              <Clock size={18} />
            </div>
          </div>
          <div className="rlc-kpi-val-row">
            <span className="rlc-kpi-value" style={{ color: '#D97706' }}>{summary.expiring_soon_rates}</span>
            <span className="rlc-kpi-sub">needs review</span>
          </div>
        </div>

        <div className="rlc-kpi-card">
          <div className="rlc-kpi-top">
            <span className="rlc-kpi-label">Expired Rates</span>
            <div className="rlc-kpi-icon-wrap rlc-icon-rose">
              <Flame size={18} />
            </div>
          </div>
          <div className="rlc-kpi-val-row">
            <span className="rlc-kpi-value" style={{ color: '#DC2626' }}>{summary.expired_rates}</span>
            <span className="rlc-kpi-sub">invalid for booking</span>
          </div>
        </div>

        <div className="rlc-kpi-card">
          <div className="rlc-kpi-top">
            <span className="rlc-kpi-label">Superseded Rates</span>
            <div className="rlc-kpi-icon-wrap rlc-icon-purple">
              <TrendingDown size={18} />
            </div>
          </div>
          <div className="rlc-kpi-val-row">
            <span className="rlc-kpi-value" style={{ color: '#7C3AED' }}>{summary.superseded_rates}</span>
            <span className="rlc-kpi-sub">newer version exists</span>
          </div>
        </div>

        <div className="rlc-kpi-card">
          <div className="rlc-kpi-top">
            <span className="rlc-kpi-label">Contract Renewals</span>
            <div className="rlc-kpi-icon-wrap rlc-icon-indigo">
              <FileCheck size={18} />
            </div>
          </div>
          <div className="rlc-kpi-val-row">
            <span className="rlc-kpi-value">{summary.contracts_requiring_renewal}</span>
            <span className="rlc-kpi-sub">expiring / expired</span>
          </div>
        </div>

        <div className="rlc-kpi-card">
          <div className="rlc-kpi-top">
            <span className="rlc-kpi-label">Quotes at Risk</span>
            <div className="rlc-kpi-icon-wrap rlc-icon-rose">
              <ShieldAlert size={18} />
            </div>
          </div>
          <div className="rlc-kpi-val-row">
            <span className="rlc-kpi-value" style={{ color: summary.quotations_at_risk > 0 ? '#DC2626' : '#16A34A' }}>
              {summary.quotations_at_risk}
            </span>
            <span className="rlc-kpi-sub">affected quotations</span>
          </div>
        </div>
      </div>

      {/* ── Main Tab Navigation ── */}
      <div className="rlc-main-card">
        <div className="rlc-tabs-header">
          <div className="rlc-tab-group">
            <button
              type="button"
              className={`rlc-tab-btn ${activeTab === 'rates' ? 'active' : ''}`}
              onClick={() => setActiveTab('rates')}
            >
              <AlertTriangle size={15} /> Rates Requiring Attention ({attentionRates.length})
            </button>
            <button
              type="button"
              className={`rlc-tab-btn ${activeTab === 'contracts' ? 'active' : ''}`}
              onClick={() => setActiveTab('contracts')}
            >
              <FileText size={15} /> Contract Renewal Intelligence ({attentionContracts.length})
            </button>
            <button
              type="button"
              className={`rlc-tab-btn ${activeTab === 'timeline' ? 'active' : ''}`}
              onClick={() => setActiveTab('timeline')}
            >
              <Clock size={15} /> Lifecycle Event Timeline ({lifecycleEvents.length})
            </button>
          </div>
        </div>

        {/* ── TAB 1: Rates Requiring Attention ── */}
        {activeTab === 'rates' && (
          <div className="rlc-body">
            {/* Filter Pills */}
            <div className="rlc-filter-bar">
              <div className="rlc-filter-pills">
                <button
                  type="button"
                  className={`rlc-filter-pill ${attentionBucket === 'ALL' ? 'active' : ''}`}
                  onClick={() => setAttentionBucket('ALL')}
                >
                  All Attention ({attentionRates.length})
                </button>
                <button
                  type="button"
                  className={`rlc-filter-pill ${attentionBucket === 'EXPIRED' ? 'active' : ''}`}
                  onClick={() => setAttentionBucket('EXPIRED')}
                >
                  🔴 Expired ({attentionRates.filter((r) => r.attention_bucket === 'EXPIRED').length})
                </button>
                <button
                  type="button"
                  className={`rlc-filter-pill ${attentionBucket === 'EXPIRING_7D' ? 'active' : ''}`}
                  onClick={() => setAttentionBucket('EXPIRING_7D')}
                >
                  🟠 Expiring &le; 7d ({attentionRates.filter((r) => r.attention_bucket === 'EXPIRING_7D').length})
                </button>
                <button
                  type="button"
                  className={`rlc-filter-pill ${attentionBucket === 'EXPIRING_30D' ? 'active' : ''}`}
                  onClick={() => setAttentionBucket('EXPIRING_30D')}
                >
                  🟡 Expiring &le; 30d ({attentionRates.filter((r) => r.attention_bucket === 'EXPIRING_30D').length})
                </button>
                <button
                  type="button"
                  className={`rlc-filter-pill ${attentionBucket === 'SUPERSEDED' ? 'active' : ''}`}
                  onClick={() => setAttentionBucket('SUPERSEDED')}
                >
                  🔵 Superseded ({attentionRates.filter((r) => r.attention_bucket === 'SUPERSEDED').length})
                </button>
              </div>
            </div>

            {loading ? (
              <div className="rlc-loading-state">
                <RefreshCw size={24} className="rlc-spin" />
                <span>Loading lifecycle intelligence data...</span>
              </div>
            ) : filteredRates.length === 0 ? (
              <div className="rlc-empty-state">
                <div className="rlc-empty-icon-wrap">
                  <CheckCircle2 size={32} color="#16A34A" />
                </div>
                <h3>All Rates Within Healthy Validity</h3>
                <p>No carrier rates in this category currently require urgent commercial attention.</p>
              </div>
            ) : (
              <div className="rlc-table-wrap">
                <table className="rlc-table">
                  <thead>
                    <tr>
                      <th>Status / Urgency</th>
                      <th>Carrier & Lane</th>
                      <th>Mode / Equipment</th>
                      <th>Base Rate</th>
                      <th>Validity Expiry</th>
                      <th>Impacted Quotes</th>
                      <th style={{ textAlign: 'right' }}>Actions</th>
                    </tr>
                  </thead>
                  <tbody>
                    {filteredRates.map((rate) => (
                      <tr key={rate.rate_id}>
                        <td>{getBucketBadge(rate.attention_bucket)}</td>
                        <td>
                          <div style={{ fontWeight: 700, color: '#0F172A' }}>{rate.carrier_name}</div>
                          <div style={{ fontSize: '12px', color: '#64748B' }}>
                            {rate.origin} &rarr; {rate.destination}
                          </div>
                        </td>
                        <td>
                          <div style={{ fontWeight: 600 }}>{rate.transport_mode}</div>
                          <div style={{ fontSize: '11.5px', color: '#64748B' }}>{rate.equipment_type || 'General'} • v{rate.version_number}</div>
                        </td>
                        <td>
                          <span style={{ fontWeight: 750, color: '#0F172A' }}>
                            {rate.currency} {Number(rate.base_amount).toLocaleString(undefined, { minimumFractionDigits: 2 })}
                          </span>
                        </td>
                        <td>
                          {rate.valid_until ? (
                            <div>
                              <div style={{ fontWeight: 650, color: rate.days_remaining < 0 ? '#DC2626' : rate.days_remaining <= 7 ? '#EA580C' : '#D97706' }}>
                                {rate.valid_until}
                              </div>
                              <div style={{ fontSize: '11px', color: '#64748B' }}>
                                {rate.days_remaining < 0
                                  ? `Expired ${Math.abs(rate.days_remaining)}d ago`
                                  : `${rate.days_remaining} days left`}
                              </div>
                            </div>
                          ) : (
                            <span style={{ color: '#94A3B8' }}>No expiry set</span>
                          )}
                        </td>
                        <td>
                          {rate.affected_quotes_count > 0 ? (
                            <span className="rlc-impact-tag">
                              <ShieldAlert size={12} /> {rate.affected_quotes_count} Active Quote{rate.affected_quotes_count > 1 ? 's' : ''}
                            </span>
                          ) : (
                            <span style={{ color: '#94A3B8', fontSize: '12px' }}>0 quotes</span>
                          )}
                        </td>
                        <td style={{ textAlign: 'right' }}>
                          <button
                            type="button"
                            className="rlc-btn-action"
                            onClick={() => onViewRate && onViewRate(rate.rate_id)}
                          >
                            <Eye size={13} /> View Rate
                          </button>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
          </div>
        )}

        {/* ── TAB 2: Contract Renewal Intelligence ── */}
        {activeTab === 'contracts' && (
          <div className="rlc-body">
            {loading ? (
              <div className="rlc-loading-state">
                <RefreshCw size={24} className="rlc-spin" />
                <span>Loading contract renewal status...</span>
              </div>
            ) : attentionContracts.length === 0 ? (
              <div className="rlc-empty-state">
                <div className="rlc-empty-icon-wrap">
                  <CheckCircle2 size={32} color="#16A34A" />
                </div>
                <h3>All Contracts Healthy</h3>
                <p>There are no carrier contracts currently requiring renewal intervention.</p>
              </div>
            ) : (
              <div className="rlc-table-wrap">
                <table className="rlc-table">
                  <thead>
                    <tr>
                      <th>Contract Code</th>
                      <th>Carrier & Title</th>
                      <th>Validity Period</th>
                      <th>Renewal Status</th>
                      <th>Linked Rates</th>
                      <th>Affected Quotes</th>
                      <th style={{ textAlign: 'right' }}>Actions</th>
                    </tr>
                  </thead>
                  <tbody>
                    {attentionContracts.map((c) => (
                      <tr key={c.contract_id}>
                        <td>
                          <span style={{ fontFamily: 'monospace', fontWeight: 750, color: '#2563EB' }}>
                            {c.contract_code}
                          </span>
                        </td>
                        <td>
                          <div style={{ fontWeight: 700, color: '#0F172A' }}>{c.carrier_name}</div>
                          <div style={{ fontSize: '12px', color: '#64748B' }}>{c.title}</div>
                        </td>
                        <td>
                          <div style={{ fontWeight: 650 }}>{c.start_date || '—'} &rarr; {c.end_date || '—'}</div>
                          <div style={{ fontSize: '11px', color: c.days_remaining < 0 ? '#DC2626' : '#D97706' }}>
                            {c.days_remaining < 0 ? `Expired ${Math.abs(c.days_remaining)}d ago` : `${c.days_remaining} days remaining`}
                          </div>
                        </td>
                        <td>
                          <span className={`rlc-renewal-badge rlc-renewal-${(c.renewal_status || 'not_started').toLowerCase().replace(/_/g, '-')}`}>
                            {(c.renewal_status || 'NOT_STARTED').replace(/_/g, ' ')}
                          </span>
                        </td>
                        <td>
                          <span style={{ fontWeight: 700 }}>{c.linked_rates_count}</span> rates
                        </td>
                        <td>
                          {c.affected_quotes_count > 0 ? (
                            <span className="rlc-impact-tag">
                              <ShieldAlert size={12} /> {c.affected_quotes_count} Active Quotes
                            </span>
                          ) : (
                            <span style={{ color: '#94A3B8', fontSize: '12px' }}>0 quotes</span>
                          )}
                        </td>
                        <td style={{ textAlign: 'right' }}>
                          <button
                            type="button"
                            className="rlc-btn-action"
                            onClick={() => onViewContract && onViewContract(c.contract_id)}
                          >
                            <FileText size={13} /> View Contract
                          </button>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
          </div>
        )}

        {/* ── TAB 3: Chronological Lifecycle Event Timeline ── */}
        {activeTab === 'timeline' && (
          <div className="rlc-body">
            {loading ? (
              <div className="rlc-loading-state">
                <RefreshCw size={24} className="rlc-spin" />
                <span>Loading lifecycle transition events...</span>
              </div>
            ) : lifecycleEvents.length === 0 ? (
              <div className="rlc-empty-state">
                <div className="rlc-empty-icon-wrap">
                  <Clock size={32} color="#64748B" />
                </div>
                <h3>No Lifecycle Transition Events Yet</h3>
                <p>State changes such as expiry detection, contract renewal, and rate supersession will appear chronologically here.</p>
              </div>
            ) : (
              <div className="rlc-timeline">
                {lifecycleEvents.map((evt) => (
                  <div key={evt.id} className="rlc-timeline-item">
                    <div className="rlc-timeline-bullet">
                      {evt.event_type.includes('EXPIRED') ? (
                        <Flame size={14} color="#DC2626" />
                      ) : evt.event_type.includes('SUPERSEDED') ? (
                        <TrendingDown size={14} color="#7C3AED" />
                      ) : evt.event_type.includes('RENEWED') ? (
                        <Check size={14} color="#16A34A" />
                      ) : (
                        <Clock size={14} color="#D97706" />
                      )}
                    </div>
                    <div className="rlc-timeline-content">
                      <div className="rlc-timeline-header">
                        <span className="rlc-timeline-event-name">{evt.event_type.replace(/_/g, ' ')}</span>
                        <span className="rlc-timeline-date">{evt.created_at}</span>
                      </div>
                      <p className="rlc-timeline-desc">{evt.description}</p>
                      <div className="rlc-timeline-tags">
                        {evt.previous_status && (
                          <span className="rlc-timeline-tag">From: {evt.previous_status}</span>
                        )}
                        <span className="rlc-timeline-tag">To: {evt.current_status}</span>
                        {evt.rate_id && <span className="rlc-timeline-tag">Rate #{evt.rate_id}</span>}
                        {evt.contract_id && <span className="rlc-timeline-tag">Contract #{evt.contract_id}</span>}
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  );
}

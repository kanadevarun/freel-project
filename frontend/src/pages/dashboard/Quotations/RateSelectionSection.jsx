import React, { useState, useEffect, useCallback } from 'react';
import {
  Tag, Search, ShieldCheck, CheckCircle2, Clock, AlertTriangle,
  ArrowRight, ExternalLink, RefreshCw, X, Layers, Award, Scale, Check, Trash2
} from 'lucide-react';
import quotationService from '../../../services/quotationService';
import QuotationCommercialRiskPanel from './QuotationCommercialRiskPanel';
import toast from 'react-hot-toast';
import './RateSelectionSection.css';

export default function RateSelectionSection({ quotation, canEdit, onRateUpdated }) {
  const [snapshot, setSnapshot] = useState(null);
  const [history, setHistory] = useState([]);
  const [candidates, setCandidates] = useState([]);
  const [loading, setLoading] = useState(false);
  const [showDrawer, setShowDrawer] = useState(false);
  const [activeTab, setActiveTab] = useState('MANAGED_RATE'); // 'MANAGED_RATE' | 'SPOT_RATE' | 'ALL'

  const loadData = useCallback(async () => {
    if (!quotation?.id) return;
    setLoading(true);
    try {
      const [snapRes, histRes] = await Promise.all([
        quotationService.getRateSnapshot(quotation.id),
        quotationService.getRateSelectionHistory(quotation.id),
      ]);
      setSnapshot(snapRes?.data || snapRes || null);
      const histData = histRes?.data || histRes || [];
      setHistory(Array.isArray(histData) ? histData : []);
    } catch (err) {
      // no snapshot yet
    } finally {
      setLoading(false);
    }
  }, [quotation?.id]);

  useEffect(() => {
    loadData();
  }, [loadData]);

  const loadCandidates = async () => {
    if (!quotation?.id) return;
    try {
      const resp = await quotationService.getRateCandidates(quotation.id);
      const data = resp?.data || resp || {};
      setCandidates(data.candidates || []);
      setShowDrawer(true);
    } catch (err) {
      toast.error('Failed to search rate candidates for this lane');
    }
  };

  const handleSelectRate = async (cand) => {
    try {
      const payload = {
        source_type: cand.source_type,
        rate_id: cand.rate_id,
        spot_rate_response_id: cand.spot_rate_response_id,
        notes: `Selected from Quotation workspace on ${new Date().toLocaleDateString()}`,
      };

      if (snapshot) {
        await quotationService.replaceRate(quotation.id, payload);
        toast.success(`Commercial rate replaced with ${cand.carrier_name}`);
      } else {
        await quotationService.selectRate(quotation.id, payload);
        toast.success(`Commercial rate from ${cand.carrier_name} attached & snapshot locked`);
      }

      setShowDrawer(false);
      loadData();
      if (onRateUpdated) onRateUpdated();
    } catch (err) {
      toast.error(err?.message || 'Failed to select rate');
    }
  };

  const handleRemoveRate = async () => {
    if (!window.confirm('Are you sure you want to remove the rate selection from this quotation?')) return;
    try {
      await quotationService.removeRate(quotation.id);
      toast.success('Rate selection removed');
      setSnapshot(null);
      loadData();
      if (onRateUpdated) onRateUpdated();
    } catch (err) {
      toast.error('Failed to remove rate selection');
    }
  };

  const filteredCandidates = candidates.filter((c) => {
    if (activeTab === 'ALL') return true;
    return c.source_type === activeTab;
  });

  return (
    <div className="rss-container">
      {/* ── Section Header ── */}
      <div className="rss-header">
        <div>
          <div className="rss-title">
            <Tag size={16} style={{ color: '#2563EB' }} />
            <span>Commercial Rate Selection & Immutable Snapshot</span>
          </div>
          <div style={{ fontSize: '12px', color: '#64748B', marginTop: '2px' }}>
            {snapshot ? 'Commercial rate is locked for this quotation.' : 'No managed carrier rate or spot quote attached yet.'}
          </div>
        </div>

        <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
          {snapshot && (
            <span className="rss-locked-badge">
              <ShieldCheck size={14} /> Snapshot Locked
            </span>
          )}
          {canEdit && (
            <button
              type="button"
              className="qt-btn qt-btn--primary qt-btn--sm"
              onClick={loadCandidates}
            >
              <Search size={14} />
              <span>{snapshot ? 'Replace Commercial Rate' : 'Find Commercial Rates'}</span>
            </button>
          )}
          {snapshot && canEdit && (
            <button
              type="button"
              className="qt-btn qt-btn--ghost qt-btn--sm"
              style={{ color: '#DC2626' }}
              onClick={handleRemoveRate}
              title="Remove rate selection"
            >
              <Trash2 size={14} />
            </button>
          )}
        </div>
      </div>

      {/* ── Selected Snapshot Card ── */}
      {snapshot ? (
        <>
          <div className="rss-card-grid">
            <div>
              <div className="rss-field-label">Selected Carrier</div>
              <div className="rss-field-val">{snapshot.carrier_name}</div>
              <div style={{ fontSize: '11px', color: '#64748B' }}>
                Ref: {snapshot.carrier_reference || 'Contract Rate'}
              </div>
            </div>

            <div>
              <div className="rss-field-label">Route & Mode</div>
              <div className="rss-field-val">{snapshot.origin} → {snapshot.destination}</div>
              <div style={{ fontSize: '11px', color: '#64748B' }}>
                {snapshot.transport_mode} ({snapshot.equipment_type || '40GP'})
              </div>
            </div>

            <div>
              <div className="rss-field-label">Base Rate</div>
              <div className="rss-field-val">
                {snapshot.currency} {Number(snapshot.base_rate).toLocaleString(undefined, { minimumFractionDigits: 2 })}
              </div>
            </div>

            <div>
              <div className="rss-field-label">Commercial Total</div>
              <div className="rss-field-val highlight">
                {snapshot.currency} {Number(snapshot.commercial_total).toLocaleString(undefined, { minimumFractionDigits: 2 })}
              </div>
            </div>

            <div>
              <div className="rss-field-label">Cross-Module Lineage</div>
              <div style={{ marginTop: '4px' }}>
                <a
                  href="/dashboard/rate-management"
                  target="_blank"
                  rel="noreferrer"
                  style={{ fontSize: '11.5px', color: '#2563EB', fontWeight: 700, display: 'inline-flex', alignItems: 'center', gap: '4px' }}
                >
                  View Source Rate <ExternalLink size={11} />
                </a>
              </div>
            </div>
          </div>

          {/* ── Task 19.6: Embedded Rate Health & Commercial Risk Panel ── */}
          <QuotationCommercialRiskPanel
            quotationId={quotation.id}
            quotationStatus={quotation.status}
            onRateReplaced={() => {
              loadData();
              if (onRateUpdated) onRateUpdated();
            }}
          />
        </>
      ) : (
        <div style={{ background: '#F8FAFC', padding: '16px', borderRadius: '8px', border: '1px dashed #CBD5E1', fontSize: '13px', color: '#64748B', textAlign: 'center' }}>
          Click <strong>"Find Commercial Rates"</strong> to search managed contract rates or spot-rate responses for lane <strong>{quotation.origin} → {quotation.destination}</strong>.
        </div>
      )}

      {/* ── Rate Selection Audit History Timeline ── */}
      {history.length > 0 && (
        <div style={{ borderTop: '1px solid #F1F5F9', paddingTop: '12px' }}>
          <div style={{ fontSize: '11.5px', fontWeight: 750, color: '#64748B', textTransform: 'uppercase', marginBottom: '8px' }}>
            Rate Selection Audit Timeline ({history.length})
          </div>
          <div style={{ display: 'flex', flexDirection: 'column', gap: '6px' }}>
            {history.map((h) => (
              <div key={h.id} style={{ display: 'flex', justifyContent: 'space-between', fontSize: '12px', background: '#F8FAFC', padding: '6px 10px', borderRadius: '6px' }}>
                <div>
                  <strong style={{ color: '#2563EB' }}>{h.event_type}</strong> — {h.description}
                </div>
                <div style={{ color: '#94A3B8', fontSize: '11px' }}>
                  {new Date(h.created_at).toLocaleString()} • {h.performed_by || 'User'}
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* ── Rate Candidates Search Modal Drawer ── */}
      {showDrawer && (
        <div className="rss-drawer-overlay" onClick={() => setShowDrawer(false)}>
          <div className="rss-modal" onClick={(e) => e.stopPropagation()}>
            <div className="rss-modal-header">
              <h3 style={{ margin: 0, fontSize: '16px', fontWeight: 800 }}>
                Select Commercial Rate for {quotation.origin} → {quotation.destination}
              </h3>
              <button className="rc-modal-close" onClick={() => setShowDrawer(false)}>
                <X size={18} />
              </button>
            </div>

            <div className="rss-modal-tabs">
              <button
                className={`rss-tab-btn ${activeTab === 'MANAGED_RATE' ? 'is-active' : ''}`}
                onClick={() => setActiveTab('MANAGED_RATE')}
              >
                <Layers size={14} /> Managed Contract Rates
              </button>
              <button
                className={`rss-tab-btn ${activeTab === 'SPOT_RATE' ? 'is-active' : ''}`}
                onClick={() => setActiveTab('SPOT_RATE')}
              >
                <Tag size={14} /> Spot Rate Responses
              </button>
              <button
                className={`rss-tab-btn ${activeTab === 'ALL' ? 'is-active' : ''}`}
                onClick={() => setActiveTab('ALL')}
              >
                All Options ({candidates.length})
              </button>
            </div>

            <div className="rss-modal-body">
              {filteredCandidates.length === 0 ? (
                <div style={{ textAlign: 'center', padding: '40px', color: '#64748B' }}>
                  No matching rate candidates found for this transport mode and lane.
                </div>
              ) : (
                <div className="sr-table-card" style={{ border: '1px solid #E2E8F0' }}>
                  <table className="sr-table">
                    <thead>
                      <tr>
                        <th>Source</th>
                        <th>Carrier</th>
                        <th>Rate / Equipment</th>
                        <th>Commercial Rate</th>
                        <th>Transit</th>
                        <th>Recommendation</th>
                        <th style={{ textAlign: 'right' }}>Action</th>
                      </tr>
                    </thead>
                    <tbody>
                      {filteredCandidates.map((cand, idx) => (
                        <tr key={idx}>
                          <td>
                            <span style={{ fontSize: '11px', fontWeight: 800, padding: '2px 6px', borderRadius: '4px', background: cand.source_type === 'MANAGED_RATE' ? '#EFF6FF' : '#FEF3C7', color: cand.source_type === 'MANAGED_RATE' ? '#1D4ED8' : '#B45309' }}>
                              {cand.source_type === 'MANAGED_RATE' ? 'CONTRACT' : 'SPOT'}
                            </span>
                          </td>
                          <td>
                            <strong>{cand.carrier_name}</strong>
                          </td>
                          <td>
                            {cand.rate_type} ({cand.equipment_type})
                          </td>
                          <td>
                            <strong style={{ color: '#0F172A', fontSize: '14px' }}>
                              {cand.currency} {Number(cand.commercial_total).toLocaleString(undefined, { minimumFractionDigits: 2 })}
                            </strong>
                          </td>
                          <td>
                            {cand.transit_days ? `${cand.transit_days} days` : '—'}
                          </td>
                          <td>
                            <div style={{ display: 'flex', gap: '4px' }}>
                              {cand.recommendation_tags?.map((t) => (
                                <span key={t} style={{ fontSize: '10px', fontWeight: 800, padding: '2px 5px', borderRadius: '4px', background: t === 'CHEAPEST' ? '#ECFDF5' : (t === 'FASTEST' ? '#EFF6FF' : '#FAF5FF'), color: t === 'CHEAPEST' ? '#047857' : (t === 'FASTEST' ? '#1D4ED8' : '#7E22CE') }}>
                                  {t}
                                </span>
                              ))}
                            </div>
                          </td>
                          <td style={{ textAlign: 'right' }}>
                            <button
                              type="button"
                              className="sr-btn-primary"
                              style={{ padding: '4px 10px', fontSize: '12px' }}
                              onClick={() => handleSelectRate(cand)}
                            >
                              Select & Snapshot
                            </button>
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              )}
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

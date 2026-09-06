import React, { useState, useEffect, useCallback } from 'react';
import {
  AlertTriangle,
  ShieldAlert,
  Flame,
  Clock,
  TrendingDown,
  CheckCircle2,
  ArrowRight,
  RefreshCw,
  Scale,
  Sparkles,
  Layers,
  ChevronRight,
  Info,
  Check,
  Building2,
  Calendar,
  Lock
} from 'lucide-react';
import quotationService from '../../../services/quotationService';
import ConfirmModal from '../Settings/ConfirmModal';
import toast from 'react-hot-toast';
import './QuotationCommercialRiskPanel.css';

export default function QuotationCommercialRiskPanel({
  quotationId,
  quotationStatus,
  onRateReplaced,
}) {
  const [riskSummary, setRiskSummary] = useState(null);
  const [candidates, setCandidates] = useState([]);
  const [selectedCandidate, setSelectedCandidate] = useState(null);
  const [impactAnalysis, setImpactAnalysis] = useState(null);
  const [loading, setLoading] = useState(true);
  const [replacing, setReplacing] = useState(false);
  const [confirmModal, setConfirmModal] = useState({ isOpen: false, data: null });

  const loadRiskData = useCallback(async () => {
    if (!quotationId) return;
    setLoading(true);
    try {
      const [risksRes, candRes] = await Promise.all([
        quotationService.getRateRisks(quotationId),
        quotationService.getRateReplacementCandidates(quotationId),
      ]);

      const rSummary = risksRes?.data || risksRes || {};
      const cands = candRes?.data || candRes || [];

      setRiskSummary(rSummary);
      setCandidates(cands);

      if (cands.length > 0) {
        setSelectedCandidate(cands[0]);
      }
    } catch (err) {
      console.error('Failed to load quotation risk data:', err);
    } finally {
      setLoading(false);
    }
  }, [quotationId]);

  useEffect(() => {
    loadRiskData();
  }, [loadRiskData]);

  // Load commercial impact whenever candidate selection changes
  useEffect(() => {
    if (!quotationId || !selectedCandidate) return;

    const fetchImpact = async () => {
      try {
        const res = await quotationService.getCommercialImpactAnalysis(quotationId, {
          replacement_rate_id: selectedCandidate.rate_id,
          replacement_spot_id: selectedCandidate.spot_rate_response_id,
        });
        setImpactAnalysis(res?.data || res || null);
      } catch (err) {
        console.error('Failed to fetch commercial impact analysis:', err);
      }
    };

    fetchImpact();
  }, [quotationId, selectedCandidate]);

  const handleResolveRisk = async (riskId) => {
    try {
      await quotationService.resolveRateRisk(quotationId, riskId);
      toast.success('Risk alert marked as resolved');
      loadRiskData();
    } catch (err) {
      toast.error(err?.message || 'Failed to resolve risk alert');
    }
  };

  const handleConfirmReplace = async () => {
    if (!selectedCandidate) return;
    setReplacing(true);
    try {
      await quotationService.replaceRate(quotationId, {
        source_type: selectedCandidate.source_type,
        rate_id: selectedCandidate.rate_id,
        spot_rate_response_id: selectedCandidate.spot_rate_response_id,
        notes: `Replaced rate from ${selectedCandidate.carrier_name} via Commercial Risk & Rate Health panel`,
      });

      toast.success(`Successfully replaced rate with ${selectedCandidate.carrier_name}! New commercial snapshot locked.`);
      setConfirmModal({ isOpen: false, data: null });
      if (onRateReplaced) onRateReplaced();
      loadRiskData();
    } catch (err) {
      toast.error(err?.message || 'Failed to replace quotation rate');
    } finally {
      setReplacing(false);
    }
  };

  const isEditable = quotationStatus === 'DRAFT' || quotationStatus === 'CHANGES_REQUESTED';
  const activeRisks = riskSummary?.risks?.filter((r) => !r.is_resolved) || [];

  if (loading) {
    return (
      <div className="qcr-panel qcr-panel-loading">
        <RefreshCw size={20} className="qcr-spin" />
        <span>Evaluating commercial rate health...</span>
      </div>
    );
  }

  if (!riskSummary?.has_active_snapshot) {
    return null; // No rate snapshot on this quotation
  }

  return (
    <div className="qcr-panel">
      <div className="qcr-header">
        <div className="qcr-header-left">
          <div className="qcr-header-title-row">
            <ShieldAlert size={18} className={activeRisks.length > 0 ? 'qcr-icon-alert' : 'qcr-icon-safe'} />
            <h3 className="qcr-header-title">Commercial Risk & Rate Health</h3>
            {activeRisks.length > 0 ? (
              <span className="qcr-badge qcr-badge-risk">
                {activeRisks.length} Risk{activeRisks.length > 1 ? 's' : ''} Detected
              </span>
            ) : (
              <span className="qcr-badge qcr-badge-safe">
                <CheckCircle2 size={12} /> Rate Healthy
              </span>
            )}
          </div>
          <p className="qcr-header-sub">
            Continuous validity, contract expiration, and supersession monitoring for this quotation's locked rate snapshot.
          </p>
        </div>

        <button
          type="button"
          className="qcr-btn-refresh"
          onClick={loadRiskData}
          title="Refresh risk evaluation"
        >
          <RefreshCw size={14} /> Refresh
        </button>
      </div>

      {/* ── Active Risk Banners ── */}
      {activeRisks.length > 0 && (
        <div className="qcr-risks-list">
          {activeRisks.map((rk) => (
            <div key={rk.id} className={`qcr-risk-banner qcr-risk-${rk.severity.toLowerCase()}`}>
              <div className="qcr-risk-icon">
                {rk.severity === 'CRITICAL' ? (
                  <Flame size={18} />
                ) : rk.severity === 'WARNING' ? (
                  <AlertTriangle size={18} />
                ) : (
                  <Info size={18} />
                )}
              </div>
              <div className="qcr-risk-content">
                <div className="qcr-risk-headline-row">
                  <span className="qcr-risk-headline">{rk.headline}</span>
                  <span className={`qcr-severity-tag qcr-sev-${rk.severity.toLowerCase()}`}>
                    {rk.severity}
                  </span>
                </div>
                <p className="qcr-risk-desc">{rk.description}</p>
                {rk.recommended_action && (
                  <div className="qcr-risk-action-text">
                    <strong>Recommended Action:</strong> {rk.recommended_action}
                  </div>
                )}
              </div>
              <button
                type="button"
                className="qcr-btn-resolve"
                onClick={() => handleResolveRisk(rk.id)}
                title="Mark this risk resolved"
              >
                <Check size={13} /> Dismiss
              </button>
            </div>
          ))}
        </div>
      )}

      {/* ── Rate Replacement Recommendations & Commercial Impact ── */}
      {candidates.length > 0 ? (
        <div className="qcr-replacement-section">
          <div className="qcr-section-header">
            <div>
              <h4 className="qcr-section-title">
                <Sparkles size={15} color="#2563EB" /> Replacement Rate Recommendations ({candidates.length})
              </h4>
              <div className="qcr-section-sub">
                Alternative carrier quotes matching this lane and transport mode. Replacing a rate requires explicit user action.
              </div>
            </div>
          </div>

          {/* Candidate Selection Cards */}
          <div className="qcr-candidates-grid">
            {candidates.map((cand, idx) => {
              const isSelected =
                selectedCandidate &&
                (selectedCandidate.rate_id === cand.rate_id &&
                  selectedCandidate.spot_rate_response_id === cand.spot_rate_response_id);

              return (
                <div
                  key={idx}
                  className={`qcr-candidate-card ${isSelected ? 'is-selected' : ''}`}
                  onClick={() => setSelectedCandidate(cand)}
                >
                  <div className="qcr-cand-top">
                    <span className="qcr-cand-carrier">{cand.carrier_name}</span>
                    <span className="qcr-cand-source">
                      {cand.source_type === 'SPOT_RATE' ? 'Spot Quote' : `Contract v${cand.version_number}`}
                    </span>
                  </div>

                  <div className="qcr-cand-price-row">
                    <span className="qcr-cand-price">
                      {cand.currency} {Number(cand.commercialTotal || cand.commercial_total || cand.base_rate).toLocaleString(undefined, { minimumFractionDigits: 2 })}
                    </span>
                    <span className="qcr-cand-transit">{cand.transit_days}d transit</span>
                  </div>

                  {cand.recommendation_tags && cand.recommendation_tags.length > 0 && (
                    <div className="qcr-cand-tags">
                      {cand.recommendation_tags.map((tag) => (
                        <span key={tag} className={`qcr-cand-tag qcr-tag-${tag.toLowerCase()}`}>
                          {tag.replace(/_/g, ' ')}
                        </span>
                      ))}
                    </div>
                  )}
                </div>
              );
            })}
          </div>

          {/* ── Side-by-Side Commercial Impact Matrix ── */}
          {impactAnalysis && (
            <div className="qcr-impact-card">
              <div className="qcr-impact-header">
                <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                  <Scale size={16} color="#0F172A" />
                  <span className="qcr-impact-title">Side-by-Side Commercial Impact</span>
                </div>
                {impactAnalysis.impact_summary && (
                  <span className="qcr-impact-summary-pill">
                    {impactAnalysis.impact_summary}
                  </span>
                )}
              </div>

              <div className="qcr-impact-matrix">
                <div className="qcr-impact-col">
                  <div className="qcr-impact-col-header">Current Snapshot</div>
                  <div className="qcr-impact-row">
                    <span className="qcr-impact-field-label">Carrier:</span>
                    <span className="qcr-impact-field-val">{impactAnalysis.current_carrier_name}</span>
                  </div>
                  <div className="qcr-impact-row">
                    <span className="qcr-impact-field-label">Commercial Total:</span>
                    <span className="qcr-impact-field-val" style={{ fontWeight: 800 }}>
                      {impactAnalysis.current_currency} {Number(impactAnalysis.current_commercial_total).toLocaleString(undefined, { minimumFractionDigits: 2 })}
                    </span>
                  </div>
                  <div className="qcr-impact-row">
                    <span className="qcr-impact-field-label">Estimated Transit:</span>
                    <span className="qcr-impact-field-val">{impactAnalysis.current_transit_days} Days</span>
                  </div>
                  <div className="qcr-impact-row">
                    <span className="qcr-impact-field-label">Valid Until:</span>
                    <span className="qcr-impact-field-val">{impactAnalysis.current_valid_until || '—'}</span>
                  </div>
                </div>

                <div className="qcr-impact-arrow">
                  <ArrowRight size={20} color="#64748B" />
                </div>

                <div className="qcr-impact-col is-replacement">
                  <div className="qcr-impact-col-header">Replacement Option</div>
                  <div className="qcr-impact-row">
                    <span className="qcr-impact-field-label">Carrier:</span>
                    <span className="qcr-impact-field-val">{impactAnalysis.replacement_carrier_name}</span>
                  </div>
                  <div className="qcr-impact-row">
                    <span className="qcr-impact-field-label">Commercial Total:</span>
                    <span className="qcr-impact-field-val" style={{ fontWeight: 800, color: '#2563EB' }}>
                      {impactAnalysis.replacement_currency} {Number(impactAnalysis.replacement_commercial_total).toLocaleString(undefined, { minimumFractionDigits: 2 })}
                    </span>
                  </div>
                  <div className="qcr-impact-row">
                    <span className="qcr-impact-field-label">Price Variance:</span>
                    <span
                      className="qcr-impact-field-val"
                      style={{
                        fontWeight: 800,
                        color: impactAnalysis.is_cheaper ? '#16A34A' : impactAnalysis.price_difference_amount > 0 ? '#DC2626' : '#0F172A',
                      }}
                    >
                      {impactAnalysis.price_difference_amount > 0 ? '+' : ''}
                      {impactAnalysis.current_currency} {Number(impactAnalysis.price_difference_amount).toLocaleString(undefined, { minimumFractionDigits: 2 })}
                      {' '}({impactAnalysis.price_difference_percentage > 0 ? '+' : ''}{impactAnalysis.price_difference_percentage.toFixed(1)}%)
                    </span>
                  </div>
                  <div className="qcr-impact-row">
                    <span className="qcr-impact-field-label">Valid Until:</span>
                    <span className="qcr-impact-field-val">{impactAnalysis.replacement_valid_until || '—'}</span>
                  </div>
                </div>
              </div>

              {/* Action Bar */}
              <div className="qcr-impact-footer">
                <div className="qcr-immutable-notice">
                  <Lock size={13} /> Replacing the rate creates a new immutable snapshot. Historical versions remain preserved in audit logs.
                </div>

                {isEditable ? (
                  <button
                    type="button"
                    className="qcr-btn-replace"
                    onClick={() =>
                      setConfirmModal({
                        isOpen: true,
                        data: selectedCandidate,
                      })
                    }
                    disabled={replacing}
                  >
                    <Check size={14} /> Replace Rate & Lock Commercial Snapshot
                  </button>
                ) : (
                  <div className="qcr-locked-quote-notice">
                    Quotation status ({quotationStatus}) is locked. Reopen as draft to replace rate.
                  </div>
                )}
              </div>
            </div>
          )}
        </div>
      ) : (
        <div className="qcr-no-candidates">
          <Info size={16} color="#64748B" />
          <span>No alternative replacement carrier rates currently available for this lane.</span>
        </div>
      )}

      {/* Confirmation Modal for Rate Replacement */}
      <ConfirmModal
        isOpen={confirmModal.isOpen}
        title="Replace Quotation Rate & Snapshot"
        message={`Are you sure you want to replace this quotation's active carrier rate with ${confirmModal.data?.carrier_name} (${confirmModal.data?.currency} ${Number(confirmModal.data?.commercialTotal || confirmModal.data?.commercial_total || confirmModal.data?.base_rate).toLocaleString(undefined, { minimumFractionDigits: 2 })})? A new immutable commercial snapshot will be created and locked for this quotation.`}
        confirmText="Replace Rate & Snapshot"
        confirmStyle="primary"
        onConfirm={handleConfirmReplace}
        onCancel={() => setConfirmModal({ isOpen: false, data: null })}
      />
    </div>
  );
}

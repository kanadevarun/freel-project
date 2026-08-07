import React, { useState, useEffect, useCallback } from 'react';
import toast from 'react-hot-toast';
import { carrierService } from '../../../services/carrierService';
import { rfqService } from '../../../services/rfqService';
import AgentStatusBadge from '../../../components/agent/AgentStatusBadge';

/**
 * PricingWorkspace — the dedicated screen for the Pricing team.
 *
 * Flow:
 *  1. Page loads → calls GET /rfqs/{id}/carrier-rates (backend fetches from
 *     FF partner / mock and ranks them with AI).
 *  2. Displays a side-by-side rate comparison table with:
 *       - Buy rate, sell rate (editable), margin %, transit days
 *       - Reliability bar, deadline status
 *       - AI reasoning badge on recommended carrier
 *  3. Pricing team clicks "Select" on their preferred carrier → adjusts markup
 *     (optional) → clicks "Approve & Send Quote".
 *  4. Calls POST /rfqs/{id}/approve-quote → stage advances to QUOTE_SENT.
 *
 * Props:
 *  rfq              — The RFQ object (has id, origin, destination, target_date, etc.)
 *  quotes           — Existing quotes already stored in DB (AI drafts or manual)
 *  onQuoteSubmitted — Callback to refresh parent state after approval
 */
export default function PricingWorkspace({ rfq, quotes, onQuoteSubmitted }) {
  // ── State ────────────────────────────────────────────────────────────────────
  const [rates, setRates]                 = useState([]);
  const [overallReasoning, setOverallReasoning] = useState('');
  const [fetchedAt, setFetchedAt]         = useState(null);
  const [isLoadingRates, setIsLoadingRates] = useState(true);
  const [rateError, setRateError]         = useState(null);

  // selectedIdx tracks which carrier row the Pricing team has highlighted
  const [selectedIdx, setSelectedIdx]     = useState(null);
  // customSellPrices lets Pricing override the suggested 20% margin per carrier
  const [customSellPrices, setCustomSellPrices] = useState({});
  const [isApproving, setIsApproving]     = useState(false);

  // Manual quote form (fallback when carrier API has no data)
  const [showManualForm, setShowManualForm] = useState(false);
  const [manualCarrier, setManualCarrier]   = useState('');
  const [manualBuyPrice, setManualBuyPrice] = useState('');
  const [manualSellPrice, setManualSellPrice] = useState('');
  const [manualTransit, setManualTransit]   = useState('');
  const [isSubmittingManual, setIsSubmittingManual] = useState(false);

  // ── Fetch carrier rates on mount ─────────────────────────────────────────────
  const fetchRates = useCallback(async () => {
    if (!rfq?.id) return;
    setIsLoadingRates(true);
    setRateError(null);
    try {
      const resp = await carrierService.getCarrierRates(rfq.id);
      // Backend returns: { rates, overall_reasoning, recommended_idx, fetched_at }
      setRates(resp.rates || []);
      setOverallReasoning(resp.overall_reasoning || '');
      setFetchedAt(resp.fetched_at);
      // Auto-select the recommended carrier so the form is pre-filled
      if (resp.recommended_idx >= 0) {
        setSelectedIdx(resp.recommended_idx);
      }
    } catch (err) {
      // RFQ might not have origin/destination yet
      if (err.status === 400) {
        setRateError('This RFQ is missing origin or destination — add them to fetch carrier rates.');
      } else {
        setRateError('Unable to fetch carrier rates. You can add a manual quote below.');
      }
      setShowManualForm(true);
    } finally {
      setIsLoadingRates(false);
    }
  }, [rfq?.id]);

  useEffect(() => {
    fetchRates();
  }, [fetchRates]);

  // ── Helpers ──────────────────────────────────────────────────────────────────
  const getSellPrice = (idx) => {
    // Custom sell price overrides the AI suggestion
    return customSellPrices[idx] !== undefined
      ? parseFloat(customSellPrices[idx])
      : rates[idx]?.sell_price;
  };

  const getMargin = (idx) => {
    const sell = getSellPrice(idx);
    const buy = rates[idx]?.buy_price;
    if (!sell || !buy || sell === 0) return 0;
    return (((sell - buy) / sell) * 100).toFixed(1);
  };

  const getDeadlineBadge = (status) => {
    switch (status) {
      case 'on_time':   return <span className="deadline-ok">✓ Meets deadline</span>;
      case 'borderline': return <span className="deadline-warn">⚠ Borderline</span>;
      case 'missed':    return <span className="deadline-miss">✗ Misses deadline</span>;
      default:          return null;
    }
  };

  // ── Approve & Send ───────────────────────────────────────────────────────────
  const handleApproveAndSend = async () => {
    if (selectedIdx === null) {
      toast.error('Please select a carrier first');
      return;
    }

    const selectedRate = rates[selectedIdx];
    const sellPrice = getSellPrice(selectedIdx);

    if (!sellPrice || sellPrice <= selectedRate.buy_price) {
      toast.error('Sell price must be higher than buy price');
      return;
    }

    setIsApproving(true);
    try {
      // Step 1: Save the selected rate as a Quote in the DB
      const quotePayload = {
        carrier_name:            selectedRate.carrier_name,
        buy_price:               selectedRate.buy_price,
        sell_price:              sellPrice,
        transit_time_days:       selectedRate.transit_days,
        reliability_score:       selectedRate.reliability_score,
        historical_success_rate: selectedRate.historical_success_rate,
        is_recommended:          selectedRate.is_recommended,
        ai_reasoning:            selectedRate.ai_reasoning || '',
      };
      const quoteResp = await rfqService.addQuote(rfq.id, quotePayload);

      // Step 2: Approve the saved quote (this advances stage to QUOTE_SENT)
      await carrierService.approveQuote(rfq.id, quoteResp.id);

      toast.success(`Quote approved — ${selectedRate.carrier_name} @ $${sellPrice.toFixed(2)}/ctr`);
      onQuoteSubmitted();
    } catch (err) {
      console.error('Approve quote failed:', err);
      toast.error(err.message || 'Failed to approve quote');
    } finally {
      setIsApproving(false);
    }
  };

  // ── Manual quote submit (fallback) ───────────────────────────────────────────
  const handleManualSubmit = async (e) => {
    e.preventDefault();
    if (!manualCarrier || !manualBuyPrice || !manualSellPrice) {
      toast.error('Please fill in all required fields');
      return;
    }
    setIsSubmittingManual(true);
    try {
      await rfqService.addQuote(rfq.id, {
        carrier_name:      manualCarrier,
        buy_price:         parseFloat(manualBuyPrice),
        sell_price:        parseFloat(manualSellPrice),
        transit_time_days: manualTransit ? parseInt(manualTransit, 10) : null,
      });
      toast.success('Manual quote submitted');
      setManualCarrier(''); setManualBuyPrice('');
      setManualSellPrice(''); setManualTransit('');
      onQuoteSubmitted();
    } catch (err) {
      toast.error('Failed to submit quote');
    } finally {
      setIsSubmittingManual(false);
    }
  };

  // ── Render ───────────────────────────────────────────────────────────────────
  return (
    <div className="pricing-workspace">
      {/* Header */}
      <div className="pricing-header">
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
          <div>
            <h3>Carrier Comparison Engine</h3>
            <p>
              {rfq?.origin && rfq?.destination
                ? `${rfq.origin} → ${rfq.destination} · Fetching rates from FF partner`
                : 'Add origin & destination to fetch live carrier rates'}
            </p>
          </div>
          <AgentStatusBadge status={rfq?.agent_status} />
        </div>
        {fetchedAt && (
          <div style={{ fontSize: '0.72rem', color: 'var(--slate-400)', marginTop: '6px' }}>
            📡 Rates last fetched: {new Date(fetchedAt).toLocaleTimeString()} · Valid 24 hours
          </div>
        )}
      </div>

      {/* ── Carrier Rate Comparison Table ──────────────────────────────────── */}
      {isLoadingRates ? (
        <div className="pricing-loading">
          <div className="spinner" />
          <span>Fetching carrier rates from FF partner…</span>
        </div>
      ) : rateError ? (
        <div className="alert-warning" style={{ marginBottom: '16px' }}>
          ⚠️ {rateError}
        </div>
      ) : rates.length > 0 ? (
        <>
          {/* AI Overall Reasoning */}
          {overallReasoning && (
            <div className="ai-reasoning-box" style={{ marginBottom: '16px' }}>
              <div style={{ fontSize: '1rem', flexShrink: 0, marginTop: '2px' }}>🤖</div>
              <div>
                <div className="ai-reasoning-label">FREEL AI RECOMMENDATION</div>
                <div className="ai-reasoning-text">{overallReasoning}</div>
              </div>
            </div>
          )}

          {/* Rate Table */}
          <div className="carrier-rates-table">
            <div className="carrier-rates-header">
              <div style={{ flex: 1.5 }}>Carrier</div>
              <div style={{ flex: 1 }}>Buy Rate</div>
              <div style={{ flex: 1 }}>Sell Rate</div>
              <div style={{ flex: 0.8 }}>Margin</div>
              <div style={{ flex: 0.8 }}>Transit</div>
              <div style={{ flex: 0.8 }}>Free Time</div>
              <div style={{ flex: 1 }}>Reliability</div>
              <div style={{ flex: 1 }}>Deadline</div>
              <div style={{ flex: 0.7 }}>Action</div>
            </div>

            {rates.map((rate, idx) => {
              const isSelected = selectedIdx === idx;
              const sellPrice  = getSellPrice(idx);
              const margin     = getMargin(idx);

              return (
                <div
                  key={rate.carrier_name}
                  className={`carrier-rate-row${rate.is_recommended ? ' recommended' : ''}${isSelected ? ' selected' : ''}`}
                  onClick={() => setSelectedIdx(idx)}
                  style={{
                    display: 'flex',
                    flexDirection: 'column',
                    padding: '16px',
                    gap: '12px',
                    transition: 'all 0.3s cubic-bezier(0.4, 0, 0.2, 1)',
                  }}
                >
                  {/* Top Layer: Column Summary Grid */}
                  <div style={{ display: 'flex', width: '100%', alignItems: 'center' }}>
                    {/* Carrier Name */}
                    <div style={{ flex: 1.5 }}>
                      <div className="carrier-name-row">
                        <div className="carrier-logo-badge">
                          {rate.carrier_name.slice(0, 2).toUpperCase()}
                        </div>
                        <div>
                          <div className="carrier-name">{rate.carrier_name}</div>
                          {rate.is_recommended && (
                            <span className="badge-ai">⭐ AI Pick</span>
                          )}
                        </div>
                      </div>
                      {rate.ai_reasoning && isSelected && (
                        <div className="carrier-mini-reasoning" style={{ marginTop: '8px' }}>
                          {rate.ai_reasoning}
                        </div>
                      )}
                    </div>

                    {/* Buy Rate */}
                    <div style={{ flex: 1 }}>
                      <div className="rate-price">${rate.buy_price.toFixed(2)}</div>
                      <div className="rate-sub">per container</div>
                    </div>

                    {/* Sell Rate (editable when selected) */}
                    <div style={{ flex: 1 }}>
                      {isSelected ? (
                        <input
                          type="number"
                          className="sell-price-input"
                          value={customSellPrices[idx] !== undefined ? customSellPrices[idx] : rate.sell_price.toFixed(2)}
                          onChange={(e) =>
                            setCustomSellPrices({ ...customSellPrices, [idx]: e.target.value })
                          }
                          onClick={(e) => e.stopPropagation()}
                          step="0.01"
                          min={rate.buy_price + 1}
                        />
                      ) : (
                        <>
                          <div className="rate-price" style={{ color: 'var(--brand-teal)' }}>
                            ${(sellPrice || rate.sell_price).toFixed(2)}
                          </div>
                          <div className="rate-sub">{rate.margin_pct}% margin</div>
                        </>
                      )}
                    </div>

                    {/* Margin % */}
                    <div style={{ flex: 0.8 }}>
                      <div
                        className="metric-value"
                        style={{
                          color: margin >= 15 ? 'var(--success)' : margin >= 10 ? 'var(--warning)' : 'var(--danger)',
                        }}
                      >
                        {margin}%
                      </div>
                    </div>

                    {/* Transit Days */}
                    <div style={{ flex: 0.8 }}>
                      <div className="metric-value">{rate.transit_days} days</div>
                    </div>

                    {/* Free Time Days */}
                    <div style={{ flex: 0.8 }}>
                      <div className="metric-value" style={{ color: 'var(--sky)', fontWeight: 600 }}>
                        {rate.free_days} days
                      </div>
                    </div>

                    {/* Reliability Bar */}
                    <div style={{ flex: 1 }}>
                      <div style={{ fontSize: '0.82rem', fontWeight: 600 }}>{rate.reliability_score}/100</div>
                      <div className="reliability-bar">
                        <div
                          className={`reliability-fill ${
                            rate.reliability_score >= 90 ? 'high' :
                            rate.reliability_score >= 75 ? 'medium' : 'low'
                          }`}
                          style={{ width: `${rate.reliability_score}%` }}
                        />
                      </div>
                      <div style={{ fontSize: '0.68rem', color: 'var(--slate-400)', marginTop: '2px' }}>
                        {rate.historical_success_rate}% success
                      </div>
                    </div>

                    {/* Deadline status */}
                    <div style={{ flex: 1, fontSize: '0.75rem' }}>
                      {getDeadlineBadge(rate.deadline_status)}
                    </div>

                    {/* Select button */}
                    <div style={{ flex: 0.7 }}>
                      <button
                        className={`action-btn${isSelected ? ' primary' : ''}`}
                        onClick={(e) => { e.stopPropagation(); setSelectedIdx(idx); }}
                      >
                        {isSelected ? '✓ Selected' : 'Select'}
                      </button>
                    </div>
                  </div>

                  {/* Bottom Layer: Expandable Details Panel (Toggles on Select, matches CMA CGM format) */}
                  {isSelected && (
                    <div
                      className="carrier-details-panel"
                      onClick={(e) => e.stopPropagation()} // prevent clicking details card from toggling row
                      style={{
                        width: '100%',
                        marginTop: '12px',
                        borderTop: '1px solid rgba(255, 255, 255, 0.08)',
                        paddingTop: '16px',
                        display: 'grid',
                        gridTemplateColumns: '1.2fr 1fr',
                        gap: '24px',
                      }}
                    >
                      {/* Left Block: Operational/Route Logistics info */}
                      <div style={{ display: 'flex', flexDirection: 'column', gap: '12px' }}>
                        <div style={{ fontSize: '0.78rem', color: 'var(--slate-400)', fontWeight: 600, letterSpacing: '0.05em' }}>
                          ROUTE & VESSEL DETAILS
                        </div>
                        
                        <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '10px' }}>
                          <div style={{ background: 'rgba(255,255,255,0.02)', padding: '10px', borderRadius: '8px', border: '1px solid rgba(255,255,255,0.04)' }}>
                            <div style={{ fontSize: '0.7rem', color: 'var(--slate-400)' }}>🚢 VESSEL NAME</div>
                            <div style={{ fontSize: '0.85rem', fontWeight: 600, marginTop: '2px' }}>{rate.vessel_name || 'TBA'}</div>
                          </div>
                          
                          <div style={{ background: 'rgba(255,255,255,0.02)', padding: '10px', borderRadius: '8px', border: '1px solid rgba(255,255,255,0.04)' }}>
                            <div style={{ fontSize: '0.7rem', color: 'var(--slate-400)' }}>🏷️ SERVICE CODE</div>
                            <div style={{ fontSize: '0.85rem', fontWeight: 600, marginTop: '2px' }}>{rate.service_code || 'Direct'}</div>
                          </div>

                          <div style={{ background: 'rgba(255,255,255,0.02)', padding: '10px', borderRadius: '8px', border: '1px solid rgba(255,255,255,0.04)' }}>
                            <div style={{ fontSize: '0.7rem', color: 'var(--slate-400)' }}>📍 TRANSSHIPMENT</div>
                            <div style={{ fontSize: '0.82rem', fontWeight: 600, color: 'var(--sky)', marginTop: '2px' }}>
                              {rate.via_port ? rate.via_port.toUpperCase() : 'DIRECT ROUTE'}
                            </div>
                          </div>

                          <div style={{ background: 'rgba(255,255,255,0.02)', padding: '10px', borderRadius: '8px', border: '1px solid rgba(255,255,255,0.04)' }}>
                            <div style={{ fontSize: '0.7rem', color: 'var(--slate-400)' }}>🌐 NAUTICAL MILES</div>
                            <div style={{ fontSize: '0.85rem', fontWeight: 600, marginTop: '2px' }}>
                              {rate.nautical_miles ? `${rate.nautical_miles.toLocaleString()} nm` : '—'}
                            </div>
                          </div>
                        </div>

                        {/* CO2 Emissions Pill badge */}
                        <div style={{ display: 'flex', alignItems: 'center', gap: '8px', background: 'rgba(16,185,129,0.06)', border: '1px solid rgba(16,185,129,0.15)', padding: '8px 12px', borderRadius: '8px', width: 'fit-content' }}>
                          <span style={{ fontSize: '1rem' }}>🍃</span>
                          <span style={{ fontSize: '0.78rem', color: 'var(--success)', fontWeight: 600 }}>
                            CO2 Emission: {rate.co2_emissions || '3.4'} tonnes CO2e per TEU
                          </span>
                        </div>
                      </div>

                      {/* Right Block: Surcharges / Detailed Cost Itemisation */}
                      <div style={{ display: 'flex', flexDirection: 'column', gap: '10px', borderLeft: '1px solid rgba(255,255,255,0.06)', paddingLeft: '24px' }}>
                        <div style={{ fontSize: '0.78rem', color: 'var(--slate-400)', fontWeight: 600, letterSpacing: '0.05em', marginBottom: '4px' }}>
                          CHARGES ITEMISATION (BUY COST)
                        </div>

                        <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: '0.82rem', padding: '6px 0' }}>
                          <span style={{ color: 'var(--slate-300)' }}>🚢 Base Ocean Freight</span>
                          <span style={{ fontWeight: 600 }}>${(rate.ocean_freight || rate.buy_price * 0.8).toFixed(2)}</span>
                        </div>

                        <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: '0.82rem', padding: '6px 0', borderBottom: '1px dashed rgba(255,255,255,0.06)' }}>
                          <span style={{ color: 'var(--slate-300)' }}>📤 Export Surcharges (Origin THC)</span>
                          <span style={{ fontWeight: 600 }}>${(rate.origin_charges || rate.buy_price * 0.12).toFixed(2)}</span>
                        </div>

                        <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: '0.82rem', padding: '6px 0', borderBottom: '1px dashed rgba(255,255,255,0.06)' }}>
                          <span style={{ color: 'var(--slate-300)' }}>📥 Import Surcharges (Dest THC/Congestion)</span>
                          <span style={{ fontWeight: 600 }}>${(rate.destination_charges || rate.buy_price * 0.08).toFixed(2)}</span>
                        </div>

                        <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: '0.9rem', padding: '8px 0', fontWeight: 700, borderTop: '1px solid rgba(255,255,255,0.1)' }}>
                          <span>Subtotal Buy Price</span>
                          <span style={{ color: 'var(--sky)' }}>${rate.buy_price.toFixed(2)}</span>
                        </div>
                      </div>
                    </div>
                  )}
                </div>
              );
            })}

          </div>

          {/* Approve & Send CTA */}
          {selectedIdx !== null && (
            <div className="approve-cta-bar">
              <div className="approve-cta-summary">
                <div>
                  <div className="rate-sub">SELECTED CARRIER</div>
                  <div style={{ fontWeight: 700, marginTop: 4 }}>{rates[selectedIdx]?.carrier_name}</div>
                </div>
                <div>
                  <div className="rate-sub">SELL PRICE</div>
                  <div style={{ fontWeight: 700, color: 'var(--brand-teal)', marginTop: 4 }}>
                    ${(getSellPrice(selectedIdx) || 0).toFixed(2)}/ctr
                  </div>
                </div>
                <div>
                  <div className="rate-sub">MARGIN</div>
                  <div style={{ fontWeight: 700, color: 'var(--success)', marginTop: 4 }}>
                    {getMargin(selectedIdx)}%
                  </div>
                </div>
                <div>
                  <div className="rate-sub">TRANSIT</div>
                  <div style={{ fontWeight: 700, marginTop: 4 }}>{rates[selectedIdx]?.transit_days} days</div>
                </div>
                <div>
                  <div className="rate-sub">FREE TIME</div>
                  <div style={{ fontWeight: 700, color: 'var(--sky)', marginTop: 4 }}>
                    {rates[selectedIdx]?.free_days} days
                  </div>
                </div>
              </div>
              <div style={{ display: 'flex', gap: '10px' }}>
                <button className="btn-secondary" onClick={() => setShowManualForm(true)}>
                  ✏️ Manual Override
                </button>
                <button
                  className="btn-primary"
                  onClick={handleApproveAndSend}
                  disabled={isApproving}
                >
                  {isApproving ? 'Approving…' : '✅ Approve & Send Quote →'}
                </button>
              </div>
            </div>
          )}
        </>
      ) : null}

      {/* ── Existing Quotes (previously submitted) ──────────────────────────── */}
      {quotes && quotes.length > 0 && (
        <div style={{ marginTop: '20px' }}>
          <h4 style={{ marginBottom: '12px', fontSize: '0.9rem', color: 'var(--slate-400)' }}>
            Previously Submitted Quotes
          </h4>
          <div className="carrier-options-list">
            {quotes.map((quote) => (
              <div key={quote.id} className={`carrier-card ${quote.is_recommended ? 'recommended' : ''}`}>
                <div className="carrier-card-header">
                  <div className="carrier-name">{quote.carrier_name}</div>
                  <div style={{ display: 'flex', gap: '8px', alignItems: 'center' }}>
                    <span
                      className="status-badge"
                      style={{
                        background: quote.status === 'APPROVED' ? 'rgba(16,185,129,.15)' : 'rgba(100,116,139,.15)',
                        color: quote.status === 'APPROVED' ? 'var(--success)' : 'var(--slate-400)',
                        padding: '3px 8px', borderRadius: '6px', fontSize: '0.72rem', fontWeight: 700,
                      }}
                    >
                      {quote.status}
                    </span>
                    {quote.is_recommended && <div className="badge-ai">✨ AI Drafted</div>}
                  </div>
                </div>

                {quote.ai_reasoning && (
                  <div className="ai-reasoning-box" style={{ margin: '8px 0', fontSize: '0.8rem' }}>
                    <strong>AI Note:</strong> {quote.ai_reasoning}
                  </div>
                )}

                <div className="carrier-metrics">
                  <div className="metric">
                    <span className="label">Transit</span>
                    <span className="value">{quote.transit_time_days || '-'} Days</span>
                  </div>
                  <div className="metric">
                    <span className="label">Buy Rate</span>
                    <span className="value">${quote.buy_price.toFixed(2)}</span>
                  </div>
                  <div className="metric highlight">
                    <span className="label">Sell Rate</span>
                    <span className="value">${quote.sell_price.toFixed(2)}</span>
                  </div>
                  <div className="metric">
                    <span className="label">Margin</span>
                    <span className="value text-green">
                      {(((quote.sell_price - quote.buy_price) / quote.sell_price) * 100).toFixed(1)}%
                    </span>
                  </div>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* ── Manual Quote Form (fallback / override) ──────────────────────────── */}
      {showManualForm && (
        <>
          <hr className="divider" />
          <form onSubmit={handleManualSubmit} className="pricing-form">
            <h4>Add Carrier Option Manually</h4>
            <div className="form-row">
              <div className="form-group">
                <label>Carrier Name *</label>
                <input
                  type="text"
                  className="form-control"
                  placeholder="e.g. MAERSK, MSC"
                  value={manualCarrier}
                  onChange={(e) => setManualCarrier(e.target.value)}
                />
              </div>
              <div className="form-group">
                <label>Transit Time (Days)</label>
                <input
                  type="number"
                  className="form-control"
                  placeholder="e.g. 14"
                  value={manualTransit}
                  onChange={(e) => setManualTransit(e.target.value)}
                />
              </div>
            </div>
            <div className="form-row">
              <div className="form-group">
                <label>Buy Price ($) *</label>
                <input
                  type="number"
                  step="0.01"
                  className="form-control"
                  value={manualBuyPrice}
                  onChange={(e) => setManualBuyPrice(e.target.value)}
                />
              </div>
              <div className="form-group">
                <label>Sell Price ($) *</label>
                <input
                  type="number"
                  step="0.01"
                  className="form-control"
                  value={manualSellPrice}
                  onChange={(e) => setManualSellPrice(e.target.value)}
                />
              </div>
            </div>
            {manualBuyPrice && (
              <div className="ai-suggestion-inline">
                🤖 AI Suggestion: 18% margin → recommended sell price: ${(parseFloat(manualBuyPrice) / 0.82).toFixed(2)}
              </div>
            )}
            <div style={{ display: 'flex', gap: '10px', marginTop: '16px' }}>
              <button
                type="button"
                className="btn-secondary"
                onClick={() => setShowManualForm(false)}
              >
                Cancel
              </button>
              <button type="submit" className="btn-primary" disabled={isSubmittingManual}>
                {isSubmittingManual ? 'Submitting…' : 'Submit Manual Quote'}
              </button>
            </div>
          </form>
        </>
      )}
    </div>
  );
}

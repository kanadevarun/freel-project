import React, { useState, useEffect, useCallback } from 'react';
import {
  TrendingUp,
  DollarSign,
  Layers,
  ShieldCheck,
  AlertTriangle,
  RefreshCw,
  Clock,
  Sparkles,
  Award,
  CheckCircle2,
  XCircle,
  HelpCircle,
  ArrowUpRight,
  ArrowDownRight,
  Percent,
  FileText,
  Ship,
  Plane,
  Truck,
  ExternalLink,
  Info
} from 'lucide-react';
import { rfqService } from '../../../../services/rfqService';
import './RFQPricingIntelligenceSection.css';

/**
 * Format currency amounts safely
 */
function formatMoney(amount, currency = 'USD') {
  if (amount === null || amount === undefined || isNaN(amount)) return '—';
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: currency || 'USD',
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  }).format(amount);
}

/**
 * Format hours into friendly duration
 */
function formatDuration(hours) {
  if (hours === null || hours === undefined) return '—';
  if (hours < 1) {
    const mins = Math.round(hours * 60);
    return `${mins} min${mins === 1 ? '' : 's'}`;
  }
  if (hours < 24) {
    const rounded = Math.round(hours * 10) / 10;
    return `${rounded} hr${rounded === 1 ? '' : 's'}`;
  }
  const days = Math.round((hours / 24) * 10) / 10;
  return `${days} day${days === 1 ? '' : 's'}`;
}

export default function RFQPricingIntelligenceSection({ rfqId }) {
  const [data, setData] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  const fetchIntelligence = useCallback(async () => {
    if (!rfqId) return;
    setLoading(true);
    setError(null);
    try {
      const res = await rfqService.getRFQ360PricingIntelligence(rfqId);
      const payload = res?.data || res;
      setData(payload);
    } catch (err) {
      console.error('Failed to load RFQ pricing intelligence:', err);
      setError(err?.response?.data?.error || err?.message || 'Failed to load pricing intelligence');
    } finally {
      setLoading(false);
    }
  }, [rfqId]);

  useEffect(() => {
    fetchIntelligence();
  }, [fetchIntelligence]);

  if (loading) {
    return (
      <div className="rfq-intel-loading" data-testid="rfq-intel-loading">
        <div className="rfq-intel-spinner" />
        <p className="rfq-intel-loading-text">Computing deterministic RFQ pricing intelligence & quotation comparisons...</p>
      </div>
    );
  }

  if (error) {
    return (
      <div className="rfq-intel-error" data-testid="rfq-intel-error">
        <AlertTriangle size={28} className="text-amber-500 mb-2" />
        <h4 className="font-semibold text-slate-800">Intelligence Unavailable</h4>
        <p className="text-sm text-slate-500 mb-4">{error}</p>
        <button className="rfq-intel-btn rfq-intel-btn--secondary" onClick={fetchIntelligence}>
          <RefreshCw size={14} className="mr-1.5" />
          Retry Calculation
        </button>
      </div>
    );
  }

  if (!data) return null;

  const {
    identity = {},
    quotation_comparison = {},
    commercial_performance = {},
    margin_and_risk = {},
    ai_summary = {},
    data_freshness_timestamp,
    correlation_id,
  } = data;

  const currency = identity.currency || 'USD';
  const marginHealth = margin_and_risk.margin_health_rating || 'UNKNOWN';

  return (
    <div className="rfq-intel-container" data-testid="rfq-pricing-intelligence-section">
      {/* Top Header Bar */}
      <div className="rfq-intel-header">
        <div className="rfq-intel-header-left">
          <div className="rfq-intel-badge rfq-intel-badge--readonly">
            <ShieldCheck size={14} className="rfq-intel-badge-icon" />
            <span>READ-ONLY PRICING INTELLIGENCE</span>
          </div>
          <div className="rfq-intel-title-row">
            <h3 className="rfq-intel-title">RFQ Commercial & Quotation Intelligence</h3>
            <span className={`rfq-intel-health-pill rfq-intel-health-pill--${marginHealth.toLowerCase()}`}>
              Margin: {marginHealth}
            </span>
            {margin_and_risk.price_anomaly_detected && (
              <span className="rfq-intel-anomaly-pill">
                <AlertTriangle size={12} />
                Price Anomaly
              </span>
            )}
          </div>
          <p className="rfq-intel-subtitle">
            Deterministic spread analytics, carrier response benchmarks, and grounded margin indicators for RFQ #{identity.rfq_number || rfqId}.
          </p>
        </div>

        <div className="rfq-intel-header-right">
          <div className="rfq-intel-meta">
            <div className="rfq-intel-meta-item">
              <Clock size={12} />
              <span>Calculated: {data_freshness_timestamp ? new Date(data_freshness_timestamp).toLocaleTimeString() : 'Just now'}</span>
            </div>
            {correlation_id && (
              <div className="rfq-intel-meta-item" title={correlation_id}>
                <span>Corr: {correlation_id.substring(0, 12)}...</span>
              </div>
            )}
          </div>
          <button
            className="rfq-intel-btn rfq-intel-btn--secondary"
            onClick={fetchIntelligence}
            title="Re-run deterministic calculations"
            data-testid="rfq-intel-refresh-btn"
          >
            <RefreshCw size={14} className="rfq-intel-btn-icon" />
            <span>Refresh</span>
          </button>
        </div>
      </div>

      {/* AI Grounded Summary Banner (if available) */}
      {ai_summary && ai_summary.executive_summary && (
        <div className="rfq-intel-ai-card" data-testid="rfq-intel-ai-card">
          <div className="rfq-intel-ai-header">
            <div className="rfq-intel-ai-header-left">
              <div className="rfq-intel-ai-sparkle">
                <Sparkles size={16} />
              </div>
              <span className="rfq-intel-ai-title">Grounded Pricing Insights</span>
              <span className="rfq-intel-ai-confidence">
                Confidence: {ai_summary.confidence_score || 'HIGH'}
              </span>
            </div>
            <span className="rfq-intel-ai-disclaimer">Read-Only Advisory</span>
          </div>

          <p className="rfq-intel-ai-text">{ai_summary.executive_summary}</p>

          <div className="rfq-intel-ai-grid">
            {ai_summary.quotation_spread_analysis && (
              <div className="rfq-intel-ai-subitem">
                <div className="rfq-intel-ai-subitem-title">
                  <TrendingUp size={14} className="text-indigo-600" />
                  <span>Quotation Spread Analysis</span>
                </div>
                <p>{ai_summary.quotation_spread_analysis}</p>
              </div>
            )}

            {ai_summary.carrier_response_insights && (
              <div className="rfq-intel-ai-subitem">
                <div className="rfq-intel-ai-subitem-title">
                  <Ship size={14} className="text-blue-600" />
                  <span>Carrier Response Insight</span>
                </div>
                <p>{ai_summary.carrier_response_insights}</p>
              </div>
            )}

            {ai_summary.commercial_risks && (
              <div className="rfq-intel-ai-subitem">
                <div className="rfq-intel-ai-subitem-title">
                  <AlertTriangle size={14} className="text-amber-600" />
                  <span>Commercial & Margin Risks</span>
                </div>
                <p>{ai_summary.commercial_risks}</p>
              </div>
            )}
          </div>

          {/* Attention Items & Operator Inquiries */}
          {(ai_summary.actionable_attention_items?.length > 0 || ai_summary.suggested_operator_inquiries?.length > 0) && (
            <div className="rfq-intel-ai-actions-row">
              {ai_summary.actionable_attention_items?.length > 0 && (
                <div className="rfq-intel-ai-action-col">
                  <span className="rfq-intel-ai-action-heading">Observations & Attention Items:</span>
                  <ul className="rfq-intel-ai-list">
                    {ai_summary.actionable_attention_items.map((item, idx) => (
                      <li key={idx} className="rfq-intel-ai-list-item">
                        <CheckCircle2 size={13} className="text-indigo-500 mr-1.5 flex-shrink-0" />
                        <span>{item}</span>
                      </li>
                    ))}
                  </ul>
                </div>
              )}

              {ai_summary.suggested_operator_inquiries?.length > 0 && (
                <div className="rfq-intel-ai-action-col">
                  <span className="rfq-intel-ai-action-heading">Suggested Inquiries for Carrier/Customer:</span>
                  <ul className="rfq-intel-ai-list">
                    {ai_summary.suggested_operator_inquiries.map((inq, idx) => (
                      <li key={idx} className="rfq-intel-ai-list-item">
                        <HelpCircle size={13} className="text-slate-500 mr-1.5 flex-shrink-0" />
                        <span>{inq}</span>
                      </li>
                    ))}
                  </ul>
                </div>
              )}
            </div>
          )}

          {/* Verifiable Source Citations */}
          {ai_summary.verifiable_source_citations?.length > 0 && (
            <div className="rfq-intel-ai-citations">
              <span className="rfq-intel-ai-citations-label">Grounded Sources:</span>
              {ai_summary.verifiable_source_citations.map((cite, idx) => (
                <span key={idx} className="rfq-intel-citation-chip">
                  <FileText size={11} className="mr-1" />
                  {cite}
                </span>
              ))}
            </div>
          )}
        </div>
      )}

      {/* 4-Quadrant KPI Grid */}
      <div className="rfq-intel-kpi-grid">
        {/* Quadrant 1: Quotation Spread & Comparison */}
        <div className="rfq-intel-card">
          <div className="rfq-intel-card-header">
            <div className="rfq-intel-card-icon-box bg-indigo-50 text-indigo-600">
              <DollarSign size={18} />
            </div>
            <div>
              <h4 className="rfq-intel-card-title">Quotation Spread & Pricing</h4>
              <p className="rfq-intel-card-desc">Deterministic comparison of received quotes</p>
            </div>
          </div>

          <div className="rfq-intel-kpi-row">
            <div className="rfq-intel-kpi-cell">
              <span className="rfq-intel-kpi-label">Quotes Received</span>
              <span className="rfq-intel-kpi-val" data-testid="rfq-intel-quotes-received">
                {quotation_comparison.quotations_received_count || 0}
                <span className="rfq-intel-kpi-subval">
                  {' '}/ {quotation_comparison.quotations_requested_count || 0} req
                </span>
              </span>
            </div>

            <div className="rfq-intel-kpi-cell">
              <span className="rfq-intel-kpi-label">Price Spread</span>
              <span className="rfq-intel-kpi-val" data-testid="rfq-intel-price-spread">
                {quotation_comparison.price_spread !== null && quotation_comparison.price_spread !== undefined
                  ? formatMoney(quotation_comparison.price_spread, currency)
                  : '—'}
                {quotation_comparison.price_spread_percentage !== null && quotation_comparison.price_spread_percentage !== undefined && (
                  <span className="rfq-intel-kpi-subval">
                    {' '}({quotation_comparison.price_spread_percentage.toFixed(1)}%)
                  </span>
                )}
              </span>
            </div>
          </div>

          <div className="rfq-intel-stat-list">
            <div className="rfq-intel-stat-row">
              <span className="rfq-intel-stat-name">Lowest Valid Quote</span>
              <span className="rfq-intel-stat-val font-semibold text-emerald-700" data-testid="rfq-intel-lowest-price">
                {formatMoney(quotation_comparison.lowest_quotation_amount, currency)}
              </span>
            </div>
            <div className="rfq-intel-stat-row">
              <span className="rfq-intel-stat-name">Average Valid Quote</span>
              <span className="rfq-intel-stat-val font-medium text-slate-700">
                {formatMoney(quotation_comparison.average_quotation_amount, currency)}
              </span>
            </div>
            <div className="rfq-intel-stat-row">
              <span className="rfq-intel-stat-name">Highest Valid Quote</span>
              <span className="rfq-intel-stat-val font-medium text-slate-700">
                {formatMoney(quotation_comparison.highest_quotation_amount, currency)}
              </span>
            </div>
            {quotation_comparison.selected_quotation_id && (
              <div className="rfq-intel-stat-row bg-slate-50 px-2 py-1 rounded">
                <span className="rfq-intel-stat-name font-medium text-indigo-900">
                  Selected: {quotation_comparison.selected_carrier_name || `Quote #${quotation_comparison.selected_quotation_id}`}
                </span>
                <span className="rfq-intel-stat-val font-bold text-indigo-700" data-testid="rfq-intel-selected-price">
                  {formatMoney(quotation_comparison.selected_quotation_amount, currency)}
                </span>
              </div>
            )}
          </div>

          {/* Response Timing */}
          <div className="rfq-intel-timing-row">
            <div className="rfq-intel-timing-item">
              <span className="text-slate-400 text-xs">Time to First Quote:</span>
              <span className="text-slate-700 font-medium text-xs ml-1">
                {formatDuration(quotation_comparison.time_to_first_quote_hours)}
              </span>
            </div>
            {quotation_comparison.time_to_final_selection_hours !== null && quotation_comparison.time_to_final_selection_hours !== undefined && (
              <div className="rfq-intel-timing-item">
                <span className="text-slate-400 text-xs">Time to Selection:</span>
                <span className="text-slate-700 font-medium text-xs ml-1">
                  {formatDuration(quotation_comparison.time_to_final_selection_hours)}
                </span>
              </div>
            )}
          </div>
        </div>

        {/* Quadrant 2: Margin & Profitability Health */}
        <div className="rfq-intel-card">
          <div className="rfq-intel-card-header">
            <div className="rfq-intel-card-icon-box bg-emerald-50 text-emerald-600">
              <Percent size={18} />
            </div>
            <div>
              <h4 className="rfq-intel-card-title">Margin & Financial Health</h4>
              <p className="rfq-intel-card-desc">Revenue, carrier cost & profit variance</p>
            </div>
          </div>

          <div className="rfq-intel-kpi-row">
            <div className="rfq-intel-kpi-cell">
              <span className="rfq-intel-kpi-label">Gross Margin</span>
              <span
                className={`rfq-intel-kpi-val ${
                  margin_and_risk.gross_margin_amount > 0 ? 'text-emerald-700' : margin_and_risk.gross_margin_amount < 0 ? 'text-rose-600' : 'text-slate-700'
                }`}
                data-testid="rfq-intel-gross-margin"
              >
                {formatMoney(margin_and_risk.gross_margin_amount, currency)}
                {margin_and_risk.gross_margin_percentage !== null && margin_and_risk.gross_margin_percentage !== undefined && (
                  <span className="rfq-intel-kpi-subval">
                    {' '}({margin_and_risk.gross_margin_percentage.toFixed(1)}%)
                  </span>
                )}
              </span>
            </div>

            <div className="rfq-intel-kpi-cell">
              <span className="rfq-intel-kpi-label">Margin Status</span>
              <span className="rfq-intel-kpi-val" data-testid="rfq-intel-margin-rating">
                {marginHealth}
              </span>
            </div>
          </div>

          <div className="rfq-intel-stat-list">
            <div className="rfq-intel-stat-row">
              <span className="rfq-intel-stat-name">Quoted Revenue (Sell)</span>
              <span className="rfq-intel-stat-val font-semibold text-slate-800" data-testid="rfq-intel-revenue">
                {formatMoney(margin_and_risk.quoted_revenue_amount, currency)}
              </span>
            </div>
            <div className="rfq-intel-stat-row">
              <span className="rfq-intel-stat-name">Estimated/Recorded Cost (Buy)</span>
              <span className="rfq-intel-stat-val font-medium text-slate-600" data-testid="rfq-intel-cost">
                {formatMoney(margin_and_risk.estimated_or_recorded_cost, currency)}
              </span>
            </div>
            {commercial_performance.lane_historical_quotes_count > 0 && (
              <div className="rfq-intel-stat-row">
                <span className="rfq-intel-stat-name">Lane Historical Avg</span>
                <span className="rfq-intel-stat-val font-medium text-slate-600">
                  {formatMoney(commercial_performance.lane_historical_average_price, currency)}
                </span>
              </div>
            )}
          </div>

          {/* Missing data warnings or notices */}
          {margin_and_risk.missing_data_reasons?.length > 0 && (
            <div className="rfq-intel-notice-box">
              <Info size={13} className="text-slate-500 mr-1.5 flex-shrink-0 mt-0.5" />
              <div className="text-xs text-slate-600">
                {margin_and_risk.missing_data_reasons.join('; ')}
              </div>
            </div>
          )}
        </div>

        {/* Quadrant 3: Commercial Performance & Conversion */}
        <div className="rfq-intel-card">
          <div className="rfq-intel-card-header">
            <div className="rfq-intel-card-icon-box bg-blue-50 text-blue-600">
              <TrendingUp size={18} />
            </div>
            <div>
              <h4 className="rfq-intel-card-title">Commercial Conversion</h4>
              <p className="rfq-intel-card-desc">Customer win rates and pipeline progress</p>
            </div>
          </div>

          <div className="rfq-intel-kpi-row">
            <div className="rfq-intel-kpi-cell">
              <span className="rfq-intel-kpi-label">Customer Win Rate</span>
              <span className="rfq-intel-kpi-val" data-testid="rfq-intel-win-rate">
                {commercial_performance.quotation_win_rate !== null && commercial_performance.quotation_win_rate !== undefined
                  ? `${(commercial_performance.quotation_win_rate * 100).toFixed(0)}%`
                  : '—'}
              </span>
            </div>

            <div className="rfq-intel-kpi-cell">
              <span className="rfq-intel-kpi-label">Lane Depth</span>
              <span className="rfq-intel-kpi-val">
                {commercial_performance.lane_historical_quotes_count || 0}
                <span className="rfq-intel-kpi-subval"> past quotes</span>
              </span>
            </div>
          </div>

          <div className="rfq-intel-stat-list">
            <div className="rfq-intel-stat-row">
              <span className="rfq-intel-stat-name">RFQ-to-Quote Status</span>
              <span className="rfq-intel-stat-val">
                {commercial_performance.rfq_to_quotation_conversion ? (
                  <span className="rfq-intel-status-badge rfq-intel-status-badge--success">Converted</span>
                ) : (
                  <span className="rfq-intel-status-badge rfq-intel-status-badge--neutral">Pending</span>
                )}
              </span>
            </div>
            <div className="rfq-intel-stat-row">
              <span className="rfq-intel-stat-name">RFQ-to-Booking Status</span>
              <span className="rfq-intel-stat-val">
                {commercial_performance.rfq_to_booking_conversion ? (
                  <span className="rfq-intel-status-badge rfq-intel-status-badge--success">Booked</span>
                ) : (
                  <span className="rfq-intel-status-badge rfq-intel-status-badge--neutral">Not Booked</span>
                )}
              </span>
            </div>
            <div className="rfq-intel-stat-row">
              <span className="rfq-intel-stat-name">Carrier Response Rate</span>
              <span className="rfq-intel-stat-val font-medium text-slate-700">
                {commercial_performance.carrier_response_rate !== null && commercial_performance.carrier_response_rate !== undefined
                  ? `${(commercial_performance.carrier_response_rate * 100).toFixed(0)}%`
                  : '—'}
              </span>
            </div>
          </div>
        </div>

        {/* Quadrant 4: RFQ Quality & Specifications */}
        <div className="rfq-intel-card">
          <div className="rfq-intel-card-header">
            <div className="rfq-intel-card-icon-box bg-amber-50 text-amber-600">
              <Layers size={18} />
            </div>
            <div>
              <h4 className="rfq-intel-card-title">RFQ Quality & Specs</h4>
              <p className="rfq-intel-card-desc">Cargo profile and completeness metrics</p>
            </div>
          </div>

          <div className="rfq-intel-kpi-row">
            <div className="rfq-intel-kpi-cell">
              <span className="rfq-intel-kpi-label">Completeness</span>
              <span className="rfq-intel-kpi-val" data-testid="rfq-intel-completeness">
                {identity.completeness_percentage || 0}%
              </span>
            </div>

            <div className="rfq-intel-kpi-cell">
              <span className="rfq-intel-kpi-label">Shipment Mode</span>
              <span className="rfq-intel-kpi-val uppercase font-semibold text-slate-700">
                {identity.shipment_mode || 'FCL'}
              </span>
            </div>
          </div>

          <div className="rfq-intel-stat-list">
            <div className="rfq-intel-stat-row">
              <span className="rfq-intel-stat-name">Origin → Destination</span>
              <span className="rfq-intel-stat-val font-medium text-slate-800 text-xs">
                {identity.origin_port || 'Origin'} → {identity.destination_port || 'Dest'}
              </span>
            </div>
            <div className="rfq-intel-stat-row">
              <span className="rfq-intel-stat-name">Incoterms</span>
              <span className="rfq-intel-stat-val font-medium text-slate-700">
                {identity.requested_incoterms || 'FOB'}
              </span>
            </div>
            <div className="rfq-intel-stat-row">
              <span className="rfq-intel-stat-name">Service Level</span>
              <span className="rfq-intel-stat-val font-medium text-slate-700">
                {identity.requested_service_level || 'STANDARD'}
              </span>
            </div>
          </div>
        </div>
      </div>

      {/* Side-by-Side Quotation Comparison Table */}
      <div className="rfq-intel-table-card" data-testid="rfq-intel-quotes-table-card">
        <div className="rfq-intel-table-header">
          <div>
            <h4 className="rfq-intel-table-title">Carrier Quotation Comparison Matrix</h4>
            <p className="rfq-intel-table-desc">
              All quotes submitted for RFQ #{identity.rfq_number || rfqId}. Read-only audit of pricing variance and transit times.
            </p>
          </div>
          <span className="rfq-intel-count-badge">
            {quotation_comparison.carrier_quotes?.length || 0} Quotes Received
          </span>
        </div>

        {quotation_comparison.carrier_quotes && quotation_comparison.carrier_quotes.length > 0 ? (
          <div className="rfq-intel-table-wrapper">
            <table className="rfq-intel-table">
              <thead>
                <tr>
                  <th>Carrier</th>
                  <th>Status</th>
                  <th>Buy Price (Cost)</th>
                  <th>Sell Price (Quoted)</th>
                  <th>Gross Margin</th>
                  <th>Transit Days</th>
                  <th>Valid Until</th>
                  <th>Selection</th>
                </tr>
              </thead>
              <tbody>
                {quotation_comparison.carrier_quotes.map((quote) => {
                  const isLowest = quote.is_lowest;
                  const isSelected = quote.is_selected;
                  const quoteCurrency = quote.currency || currency;

                  return (
                    <tr
                      key={quote.id}
                      className={`${isSelected ? 'rfq-intel-tr--selected' : ''} ${isLowest ? 'rfq-intel-tr--lowest' : ''}`}
                    >
                      <td className="font-semibold text-slate-800">
                        <div className="flex items-center gap-2">
                          <span>{quote.carrier_name || `Carrier Quote #${quote.id}`}</span>
                          {quote.carrier_code && (
                            <span className="rfq-intel-carrier-code">{quote.carrier_code}</span>
                          )}
                          {isLowest && (
                            <span className="rfq-intel-lowest-tag" title="Lowest valid price received">
                              Lowest
                            </span>
                          )}
                        </div>
                      </td>
                      <td>
                        <span className={`rfq-intel-quote-status rfq-intel-quote-status--${(quote.status || 'pending').toLowerCase()}`}>
                          {quote.status || 'PENDING'}
                        </span>
                      </td>
                      <td className="text-slate-600 font-medium">
                        {formatMoney(quote.buy_price, quoteCurrency)}
                      </td>
                      <td className="text-slate-800 font-bold">
                        {formatMoney(quote.sell_price, quoteCurrency)}
                      </td>
                      <td>
                        {quote.margin_percentage !== null && quote.margin_percentage !== undefined ? (
                          <span
                            className={`font-semibold ${
                              quote.margin_percentage >= 15 ? 'text-emerald-700' : quote.margin_percentage >= 5 ? 'text-amber-700' : 'text-rose-700'
                            }`}
                          >
                            {quote.margin_percentage.toFixed(1)}%
                            {quote.margin_amount !== null && (
                              <span className="text-xs text-slate-400 block font-normal">
                                {formatMoney(quote.margin_amount, quoteCurrency)}
                              </span>
                            )}
                          </span>
                        ) : (
                          <span className="text-slate-400">—</span>
                        )}
                      </td>
                      <td className="text-slate-600">
                        {quote.transit_days ? `${quote.transit_days} days` : '—'}
                      </td>
                      <td className="text-slate-500 text-xs">
                        {quote.valid_until ? new Date(quote.valid_until).toLocaleDateString() : 'Open'}
                      </td>
                      <td>
                        {isSelected ? (
                          <span className="rfq-intel-selected-badge">
                            <CheckCircle2 size={12} className="mr-1" />
                            Selected
                          </span>
                        ) : (
                          <span className="text-xs text-slate-400">Available</span>
                        )}
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        ) : (
          <div className="rfq-intel-empty-quotes" data-testid="rfq-intel-empty-quotes">
            <div className="rfq-intel-empty-icon">
              <Ship size={24} className="text-slate-400" />
            </div>
            <h5 className="font-semibold text-slate-800 text-sm">No Carrier Quotes Recorded</h5>
            <p className="text-xs text-slate-500 max-w-md mx-auto mt-1 mb-3">
              No carrier quotations have been submitted for this RFQ yet. Once quotes are added from carriers, spread and pricing benchmarks will compute automatically.
            </p>
          </div>
        )}
      </div>

      {/* Traceable Grounded Observations List */}
      {data.grounded_observations && data.grounded_observations.length > 0 && (
        <div className="rfq-intel-observations-card">
          <div className="rfq-intel-obs-header">
            <h5 className="rfq-intel-obs-title">Grounded Observation Ledger & Data Lineage</h5>
            <span className="text-xs text-slate-400">Deterministic Source Traceability</span>
          </div>
          <div className="rfq-intel-obs-list">
            {data.grounded_observations.map((obs, idx) => (
              <div key={idx} className="rfq-intel-obs-item">
                <div className="rfq-intel-obs-item-head">
                  <span className={`rfq-intel-obs-severity rfq-intel-obs-severity--${obs.severity}`}>
                    {obs.severity}
                  </span>
                  <span className="rfq-intel-obs-category">{obs.category}</span>
                  <span className="rfq-intel-obs-evidence">Source: {obs.evidence}</span>
                </div>
                <p className="rfq-intel-obs-msg">{obs.message}</p>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}

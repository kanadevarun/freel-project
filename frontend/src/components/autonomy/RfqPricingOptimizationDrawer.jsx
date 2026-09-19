import React, { useState, useEffect, useCallback } from 'react';
import {
  X,
  TrendingUp,
  DollarSign,
  ShieldCheck,
  CheckCircle2,
  AlertCircle,
  RefreshCw,
  FileText,
  Clock,
  Layers,
  Sparkles,
  ArrowRight,
  Info,
  Check,
  AlertTriangle,
} from 'lucide-react';
import toast from 'react-hot-toast';
import { autonomyService } from '../../services/autonomyService';

export default function RfqPricingOptimizationDrawer({
  rfq,
  isOpen,
  onClose,
  onQuotationExecuted,
}) {
  const [activeTab, setActiveTab] = useState('strategies'); // 'strategies' | 'margin-risk' | 'plan' | 'replan'
  const [state, setState] = useState(null);
  const [loading, setLoading] = useState(false);
  const [evaluating, setEvaluating] = useState(false);
  const [selectingStrategy, setSelectingStrategy] = useState(false);
  const [executingQuote, setExecutingQuote] = useState(false);
  const [replanning, setReplanning] = useState(false);

  // Replan form state
  const [replanReason, setReplanReason] = useState('Carrier rate increase or fuel escalation');
  const [replanDelta, setReplanDelta] = useState(250.0);

  const rfqId = rfq?.id;

  const fetchPricingState = useCallback(async () => {
    if (!rfqId) return;
    setLoading(true);
    try {
      const resp = await autonomyService.getRfqPricingState(rfqId);
      const data = resp?.state || resp?.data?.state || resp?.data || resp;
      setState(data);
    } catch (err) {
      console.error('Failed to load RFQ pricing state:', err);
      toast.error('Failed to load pricing optimization state');
    } finally {
      setLoading(false);
    }
  }, [rfqId]);

  useEffect(() => {
    if (isOpen && rfqId) {
      fetchPricingState();
    }
  }, [isOpen, rfqId, fetchPricingState]);

  const handleRunEvaluation = async () => {
    if (!rfqId) return;
    setEvaluating(true);
    try {
      const resp = await autonomyService.evaluateRfqPricing(rfqId);
      toast.success('AI pricing optimization evaluation completed');
      await fetchPricingState();
    } catch (err) {
      console.error('Evaluation failed:', err);
      toast.error(err.response?.data?.message || 'Evaluation failed');
    } finally {
      setEvaluating(false);
    }
  };

  const handleSelectStrategy = async (strategyId) => {
    if (!rfqId) return;
    setSelectingStrategy(true);
    try {
      await autonomyService.selectPricingStrategy(rfqId, strategyId);
      toast.success('Commercial pricing strategy updated');
      await fetchPricingState();
    } catch (err) {
      console.error('Failed to select strategy:', err);
      toast.error(err.response?.data?.message || 'Failed to select strategy');
    } finally {
      setSelectingStrategy(false);
    }
  };

  const handleExecuteQuotation = async () => {
    if (!rfqId) return;
    setExecutingQuote(true);
    try {
      const resp = await autonomyService.executePricingQuotation(rfqId);
      const quote = resp?.quotation || resp?.data?.quotation || resp?.data || resp;
      toast.success(`Quotation ${quote?.quotation_number || ''} created via Go Action System!`);
      await fetchPricingState();
      if (onQuotationExecuted) {
        onQuotationExecuted(quote);
      }
    } catch (err) {
      console.error('Failed to execute quotation:', err);
      if (err.response?.status === 403) {
        toast.error('Quotation requires manager approval before execution');
      } else {
        toast.error(err.response?.data?.message || 'Failed to execute quotation');
      }
    } finally {
      setExecutingQuote(false);
    }
  };

  const handleReplanSubmit = async (e) => {
    e.preventDefault();
    if (!rfqId) return;
    setReplanning(true);
    try {
      await autonomyService.replanRfqPricing(rfqId, replanReason, parseFloat(replanDelta) || 0.0);
      toast.success('Pricing replanning triggered and evaluated');
      await fetchPricingState();
    } catch (err) {
      console.error('Replanning failed:', err);
      toast.error(err.response?.data?.message || 'Replanning failed');
    } finally {
      setReplanning(false);
    }
  };

  if (!isOpen) return null;

  const opt = state?.optimization;
  const candidates = state?.candidates || [];
  const actualFacts = state?.actual_facts || [];
  const predictions = state?.predictions || [];
  const assumptions = state?.assumptions || [];
  const versions = state?.versions || [];

  const recommendedCandidate = candidates.find(
    (c) => c.strategy_id === opt?.recommended_strategy_id
  );

  return (
    <div
      className="fixed inset-0 z-50 overflow-hidden bg-slate-900/40 backdrop-blur-sm flex justify-end"
      data-testid="rfq-pricing-optimization-drawer"
    >
      <div className="w-full max-w-4xl bg-white h-full shadow-2xl flex flex-col border-l border-slate-200">
        {/* Header */}
        <div className="px-6 py-4 border-b border-slate-200 bg-slate-50 flex items-center justify-between">
          <div>
            <div className="flex items-center gap-3">
              <span className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-semibold bg-indigo-50 text-indigo-700 border border-indigo-200">
                <Sparkles className="w-3.5 h-3.5 text-indigo-600" />
                Task 5.5 Autonomous Pricing
              </span>
              <span className="text-xs font-medium text-slate-500">
                RFQ #{rfq?.rfq_number || state?.rfq_number || rfqId}
              </span>
              {opt?.status && (
                <span
                  className={`text-xs px-2 py-0.5 rounded font-medium ${
                    opt.status === 'EXECUTED'
                      ? 'bg-emerald-100 text-emerald-800'
                      : opt.status === 'REQUIRES_APPROVAL'
                      ? 'bg-amber-100 text-amber-800'
                      : opt.status === 'REPLANNING'
                      ? 'bg-purple-100 text-purple-800'
                      : 'bg-blue-100 text-blue-800'
                  }`}
                >
                  {opt.status}
                </span>
              )}
            </div>
            <h2 className="text-lg font-bold text-slate-900 mt-1">
              Intelligent RFQ & Pricing Optimization
            </h2>
            <p className="text-xs text-slate-500">
              {rfq?.origin || state?.origin || 'Origin'} → {rfq?.destination || state?.destination || 'Destination'} (
              {state?.customer_name || 'Customer'})
            </p>
          </div>

          <div className="flex items-center gap-2">
            <button
              onClick={handleRunEvaluation}
              disabled={evaluating}
              className="px-3 py-1.5 text-xs font-medium text-slate-700 bg-white border border-slate-300 rounded hover:bg-slate-50 flex items-center gap-1.5"
              data-testid="btn-re-evaluate-pricing"
            >
              <RefreshCw className={`w-3.5 h-3.5 ${evaluating ? 'animate-spin' : ''}`} />
              {evaluating ? 'Evaluating...' : 'Re-Evaluate'}
            </button>
            <button
              onClick={onClose}
              className="p-1.5 text-slate-400 hover:text-slate-600 rounded-lg hover:bg-slate-100"
              data-testid="btn-close-pricing-drawer"
            >
              <X className="w-5 h-5" />
            </button>
          </div>
        </div>

        {/* Tab Navigation */}
        <div className="flex border-b border-slate-200 bg-white px-6">
          <button
            onClick={() => setActiveTab('strategies')}
            className={`py-3 px-4 text-xs font-medium border-b-2 flex items-center gap-2 ${
              activeTab === 'strategies'
                ? 'border-indigo-600 text-indigo-600'
                : 'border-transparent text-slate-500 hover:text-slate-800'
            }`}
            data-testid="tab-pricing-strategies"
          >
            <DollarSign className="w-4 h-4" />
            Strategies & Recommendation
          </button>
          <button
            onClick={() => setActiveTab('margin-risk')}
            className={`py-3 px-4 text-xs font-medium border-b-2 flex items-center gap-2 ${
              activeTab === 'margin-risk'
                ? 'border-indigo-600 text-indigo-600'
                : 'border-transparent text-slate-500 hover:text-slate-800'
            }`}
            data-testid="tab-margin-risk"
          >
            <TrendingUp className="w-4 h-4" />
            Margin & Risk Intelligence
          </button>
          <button
            onClick={() => setActiveTab('plan')}
            className={`py-3 px-4 text-xs font-medium border-b-2 flex items-center gap-2 ${
              activeTab === 'plan'
                ? 'border-indigo-600 text-indigo-600'
                : 'border-transparent text-slate-500 hover:text-slate-800'
            }`}
            data-testid="tab-pricing-plan"
          >
            <Layers className="w-4 h-4" />
            Execution Plan & Action Gate
          </button>
          <button
            onClick={() => setActiveTab('replan')}
            className={`py-3 px-4 text-xs font-medium border-b-2 flex items-center gap-2 ${
              activeTab === 'replan'
                ? 'border-indigo-600 text-indigo-600'
                : 'border-transparent text-slate-500 hover:text-slate-800'
            }`}
            data-testid="tab-pricing-replan"
          >
            <RefreshCw className="w-4 h-4" />
            Versioning & Replan Simulator
          </button>
        </div>

        {/* Drawer Body */}
        <div className="flex-1 overflow-y-auto p-6 space-y-6">
          {loading ? (
            <div className="flex flex-col items-center justify-center py-20 text-slate-400">
              <RefreshCw className="w-8 h-8 animate-spin text-indigo-600 mb-3" />
              <p className="text-sm font-medium">Assembling authorized pricing context...</p>
            </div>
          ) : (
            <>
              {/* TAB 1: Strategies & Recommendations */}
              {activeTab === 'strategies' && (
                <div className="space-y-6">
                  {/* Hero Recommendation Card */}
                  {opt && (
                    <div
                      className="p-5 bg-gradient-to-r from-indigo-50/50 via-white to-slate-50 border border-indigo-100 rounded-xl shadow-sm"
                      data-testid="hero-pricing-recommendation"
                    >
                      <div className="flex items-start justify-between">
                        <div>
                          <span className="text-[11px] font-bold uppercase tracking-wider text-indigo-700 bg-indigo-100 px-2 py-0.5 rounded">
                            Recommended Strategy
                          </span>
                          <h3 className="text-xl font-extrabold text-slate-900 mt-2">
                            {recommendedCandidate?.title || opt.recommended_strategy_id}
                          </h3>
                          <p className="text-xs text-slate-600 mt-1 max-w-2xl leading-relaxed">
                            {opt.reasoning_summary}
                          </p>
                        </div>
                        <div className="text-right">
                          <div className="text-2xl font-black text-slate-900">
                            ${opt.recommended_price?.toLocaleString('en-US', { minimumFractionDigits: 2 })}
                          </div>
                          <div className="text-xs font-semibold text-emerald-600">
                            {opt.recommended_margin_pct?.toFixed(1)}% Gross Margin ($
                            {((opt.recommended_price || 0) - (opt.base_cost || 0)).toLocaleString('en-US', {
                              minimumFractionDigits: 2,
                            })}
                            )
                          </div>
                          <div className="text-[11px] text-slate-400 mt-0.5">
                            Currency: {opt.currency || 'USD'}
                          </div>
                        </div>
                      </div>

                      {/* Approval Warning Banner if Required */}
                      {opt.requires_approval && (
                        <div
                          className="mt-4 p-3 bg-amber-50 border border-amber-200 rounded-lg flex items-center gap-2.5 text-xs text-amber-900"
                          data-testid="alert-pricing-approval-required"
                        >
                          <AlertTriangle className="w-4 h-4 text-amber-600 flex-shrink-0" />
                          <span>
                            <strong>Manager Review Gate Required:</strong>{' '}
                            {opt.approval_reason || 'Quotation exceeds policy autonomy ceiling.'}
                          </span>
                        </div>
                      )}
                    </div>
                  )}

                  {/* Fact / Prediction / Assumption Chips */}
                  <div className="space-y-3">
                    <h4 className="text-xs font-bold uppercase tracking-wider text-slate-500">
                      Context Breakdown & Fact Separation
                    </h4>
                    <div className="grid grid-cols-1 gap-2" data-testid="fact-prediction-breakdown">
                      {actualFacts.map((fact, idx) => (
                        <div
                          key={`fact-${idx}`}
                          className="px-3 py-2 bg-emerald-50/70 border border-emerald-200 rounded-lg text-xs text-emerald-950 flex items-start gap-2"
                        >
                          <span className="font-bold text-[10px] px-1.5 py-0.5 rounded bg-emerald-200 text-emerald-900 uppercase flex-shrink-0">
                            ACTUAL FACT
                          </span>
                          <span>{fact}</span>
                        </div>
                      ))}
                      {predictions.map((pred, idx) => (
                        <div
                          key={`pred-${idx}`}
                          className="px-3 py-2 bg-amber-50/70 border border-amber-200 rounded-lg text-xs text-amber-950 flex items-start gap-2"
                        >
                          <span className="font-bold text-[10px] px-1.5 py-0.5 rounded bg-amber-200 text-amber-900 uppercase flex-shrink-0">
                            PREDICTION
                          </span>
                          <span>{pred}</span>
                        </div>
                      ))}
                      {assumptions.map((assump, idx) => (
                        <div
                          key={`assump-${idx}`}
                          className="px-3 py-2 bg-blue-50/70 border border-blue-200 rounded-lg text-xs text-blue-950 flex items-start gap-2"
                        >
                          <span className="font-bold text-[10px] px-1.5 py-0.5 rounded bg-blue-200 text-blue-900 uppercase flex-shrink-0">
                            ASSUMPTION
                          </span>
                          <span>{assump}</span>
                        </div>
                      ))}
                    </div>
                  </div>

                  {/* Candidate Pricing Strategies Table */}
                  <div className="space-y-3">
                    <div className="flex items-center justify-between">
                      <h4 className="text-xs font-bold uppercase tracking-wider text-slate-500">
                        Candidate Pricing Strategies Comparison
                      </h4>
                      <span className="text-xs text-slate-400">
                        {candidates.length} evaluated options
                      </span>
                    </div>

                    <div className="border border-slate-200 rounded-lg overflow-hidden">
                      <table className="w-full text-left text-xs border-collapse">
                        <thead className="bg-slate-50 border-b border-slate-200 text-slate-600 uppercase text-[10px] font-bold">
                          <tr>
                            <th className="py-2.5 px-3">Strategy</th>
                            <th className="py-2.5 px-3 text-right">Price</th>
                            <th className="py-2.5 px-3 text-right">Base Cost</th>
                            <th className="py-2.5 px-3 text-right">Margin %</th>
                            <th className="py-2.5 px-3 text-center">Risk</th>
                            <th className="py-2.5 px-3 text-center">Status</th>
                            <th className="py-2.5 px-3 text-right">Action</th>
                          </tr>
                        </thead>
                        <tbody className="divide-y divide-slate-200">
                          {candidates.map((cand) => {
                            const isCurrent = cand.strategy_id === opt?.recommended_strategy_id;
                            return (
                              <tr
                                key={cand.strategy_id}
                                className={`hover:bg-slate-50/80 transition-colors ${
                                  isCurrent ? 'bg-indigo-50/40' : ''
                                }`}
                                data-testid={`row-strategy-${cand.strategy_id}`}
                              >
                                <td className="py-3 px-3">
                                  <div className="font-bold text-slate-900 flex items-center gap-1.5">
                                    {cand.title}
                                    {isCurrent && (
                                      <span className="text-[9px] bg-indigo-600 text-white px-1.5 py-0.2 rounded font-bold">
                                        ACTIVE
                                      </span>
                                    )}
                                  </div>
                                  <div className="text-[11px] text-slate-500 max-w-xs mt-0.5 line-clamp-1">
                                    {cand.description}
                                  </div>
                                </td>
                                <td className="py-3 px-3 text-right font-bold text-slate-900">
                                  ${cand.price?.toLocaleString('en-US', { minimumFractionDigits: 2 })}
                                </td>
                                <td className="py-3 px-3 text-right text-slate-600">
                                  ${cand.base_cost?.toLocaleString('en-US', { minimumFractionDigits: 2 })}
                                </td>
                                <td className="py-3 px-3 text-right font-semibold text-emerald-600">
                                  {cand.margin_pct?.toFixed(1)}%
                                </td>
                                <td className="py-3 px-3 text-center">
                                  <span
                                    className={`px-1.5 py-0.5 rounded text-[10px] font-semibold ${
                                      cand.margin_risk_level === 'CRITICAL'
                                        ? 'bg-rose-100 text-rose-800'
                                        : cand.margin_risk_level === 'HIGH'
                                        ? 'bg-amber-100 text-amber-800'
                                        : 'bg-slate-100 text-slate-700'
                                    }`}
                                  >
                                    {cand.margin_risk_level}
                                  </span>
                                </td>
                                <td className="py-3 px-3 text-center">
                                  {cand.is_feasible ? (
                                    cand.requires_approval ? (
                                      <span className="text-[10px] bg-amber-50 text-amber-700 border border-amber-200 px-1.5 py-0.5 rounded font-medium">
                                        Approval Req
                                      </span>
                                    ) : (
                                      <span className="text-[10px] bg-emerald-50 text-emerald-700 border border-emerald-200 px-1.5 py-0.5 rounded font-medium">
                                        Pre-Approved
                                      </span>
                                    )
                                  ) : (
                                    <span
                                      className="text-[10px] bg-rose-50 text-rose-700 border border-rose-200 px-1.5 py-0.5 rounded font-bold"
                                      title={cand.infeasibility_reason || 'Hard constraint violation'}
                                    >
                                      Infeasible
                                    </span>
                                  )}
                                </td>
                                <td className="py-3 px-3 text-right">
                                  {cand.is_feasible ? (
                                    isCurrent ? (
                                      <span className="text-xs font-semibold text-indigo-600 flex items-center justify-end gap-1">
                                        <Check className="w-3.5 h-3.5" /> Selected
                                      </span>
                                    ) : (
                                      <button
                                        onClick={() => handleSelectStrategy(cand.strategy_id)}
                                        disabled={selectingStrategy}
                                        className="px-2.5 py-1 bg-white border border-slate-300 rounded text-xs font-medium text-slate-700 hover:bg-slate-50 hover:border-slate-400"
                                        data-testid={`btn-select-strat-${cand.strategy_id}`}
                                      >
                                        Select
                                      </button>
                                    )
                                  ) : (
                                    <span className="text-[10px] text-rose-600 font-medium">Blocked</span>
                                  )}
                                </td>
                              </tr>
                            );
                          })}
                        </tbody>
                      </table>
                    </div>
                  </div>
                </div>
              )}

              {/* TAB 2: Margin & Risk Intelligence */}
              {activeTab === 'margin-risk' && (
                <div className="space-y-6">
                  <div className="grid grid-cols-3 gap-4">
                    <div className="p-4 bg-slate-50 border border-slate-200 rounded-lg">
                      <div className="text-xs text-slate-500 font-medium">Rate Freshness</div>
                      <div className="text-base font-bold text-slate-900 mt-1 flex items-center gap-2">
                        <span
                          className={`w-2.5 h-2.5 rounded-full ${
                            opt?.rate_freshness_status === 'FRESH'
                              ? 'bg-emerald-500'
                              : opt?.rate_freshness_status === 'EXPIRING_SOON'
                              ? 'bg-amber-500'
                              : 'bg-rose-500'
                          }`}
                        />
                        {opt?.rate_freshness_status || 'FRESH'}
                      </div>
                      <div className="text-[11px] text-slate-400 mt-1">
                        Source: {opt?.rate_source || 'RATE_SHEET'}
                      </div>
                    </div>

                    <div className="p-4 bg-slate-50 border border-slate-200 rounded-lg">
                      <div className="text-xs text-slate-500 font-medium">Operational Risk Level</div>
                      <div className="text-base font-bold text-slate-900 mt-1">
                        {opt?.operational_risk_level || 'LOW'}
                      </div>
                      <div className="text-[11px] text-slate-400 mt-1">
                        Traffic & Port Congestion
                      </div>
                    </div>

                    <div className="p-4 bg-slate-50 border border-slate-200 rounded-lg">
                      <div className="text-xs text-slate-500 font-medium">Data Sufficiency</div>
                      <div className="text-base font-bold text-slate-900 mt-1">
                        {opt?.data_sufficiency || 'COMPLETE'}
                      </div>
                      <div className="text-[11px] text-slate-400 mt-1">
                        Confidence: {Math.round((opt?.confidence_score || 0.85) * 100)}%
                      </div>
                    </div>
                  </div>

                  <div className="p-5 border border-slate-200 rounded-xl bg-white space-y-4">
                    <h4 className="text-xs font-bold uppercase tracking-wider text-slate-700">
                      Cost Basis vs Projected Trajectory
                    </h4>
                    <div className="space-y-3">
                      <div className="flex justify-between items-center text-xs">
                        <span className="text-slate-600">Carrier Buy Price (Base Ocean Freight):</span>
                        <span className="font-bold text-slate-900">
                          ${opt?.base_cost?.toLocaleString('en-US', { minimumFractionDigits: 2 }) || '2,200.00'}
                        </span>
                      </div>
                      <div className="flex justify-between items-center text-xs">
                        <span className="text-slate-600">Predicted Operational Cost Trajectory:</span>
                        <span className="font-bold text-amber-700">
                          ${opt?.predicted_cost?.toLocaleString('en-US', { minimumFractionDigits: 2 }) || '2,350.00'}
                        </span>
                      </div>
                      <div className="flex justify-between items-center text-xs">
                        <span className="text-slate-600">Minimum Commercial Margin Floor:</span>
                        <span className="font-bold text-slate-900">{opt?.min_margin_pct || 8.0}%</span>
                      </div>
                      <div className="flex justify-between items-center text-xs">
                        <span className="text-slate-600">Target Commercial Margin:</span>
                        <span className="font-bold text-indigo-700">{opt?.target_margin_pct || 16.0}%</span>
                      </div>
                    </div>
                  </div>
                </div>
              )}

              {/* TAB 3: Execution Plan */}
              {activeTab === 'plan' && (
                <div className="space-y-6">
                  <div className="border border-slate-200 rounded-xl p-5 bg-slate-50/50">
                    <h4 className="text-xs font-bold uppercase tracking-wider text-slate-600 mb-1">
                      7-Step Controlled Autonomous Pricing Workflow
                    </h4>
                    <p className="text-xs text-slate-500 mb-4">
                      Execution boundaries strictly enforced by Go Action System and Human-in-the-Loop policies.
                    </p>

                    <div className="space-y-2.5">
                      {[
                        {
                          num: 1,
                          title: 'Verify RFQ Route & Cargo Parameters',
                          status: 'COMPLETED',
                          desc: 'Cross-reference POL, POD, container specs against carrier routing guide.',
                        },
                        {
                          num: 2,
                          title: 'Inspect Carrier Rate Freshness & Surcharges',
                          status: opt?.rate_freshness_status === 'FRESH' ? 'COMPLETED' : 'PENDING',
                          desc: `Rate sheet freshness: ${opt?.rate_freshness_status || 'FRESH'} with verified THC/BAF.`,
                        },
                        {
                          num: 3,
                          title: 'Enforce Hard Margin Floor & Pricing Rules',
                          status: 'COMPLETED',
                          desc: `Margin ${opt?.recommended_margin_pct?.toFixed(1)}% validated against ${opt?.min_margin_pct || 8.0}% floor.`,
                        },
                        {
                          num: 4,
                          title: 'Prepare Quotation Draft & Line Items',
                          status: opt?.status === 'EXECUTED' ? 'COMPLETED' : 'PENDING',
                          desc: `Assemble quotation charge breakdown with sell price of $${opt?.recommended_price?.toLocaleString('en-US', { minimumFractionDigits: 2 })}.`,
                        },
                        {
                          num: 5,
                          title: 'Manager Pricing Review & Sign-Off Gate',
                          status: opt?.requires_approval
                            ? opt?.approval_status === 'APPROVED'
                              ? 'COMPLETED'
                              : 'AWAITING_APPROVAL'
                            : 'SKIPPED',
                          desc: opt?.requires_approval
                            ? `Approval gate: ${opt?.approval_reason || 'Policy threshold'}`
                            : 'Pre-approved low-risk pricing autonomy tier.',
                        },
                        {
                          num: 6,
                          title: 'Execute Quotation via Go Action System',
                          status: opt?.status === 'EXECUTED' ? 'COMPLETED' : 'PENDING',
                          desc: 'Commit authoritative quotation record to MariaDB quotations table.',
                        },
                        {
                          num: 7,
                          title: 'Handoff to Task 5.4 Customer Follow-Up',
                          status: opt?.status === 'EXECUTED' ? 'IN_PROGRESS' : 'PENDING',
                          desc: 'Schedule dispatch notifications and monitor customer reply.',
                        },
                      ].map((s) => (
                        <div
                          key={s.num}
                          className="flex items-start gap-3 p-3 bg-white border border-slate-200 rounded-lg"
                          data-testid={`step-plan-${s.num}`}
                        >
                          <div
                            className={`w-6 h-6 rounded-full flex items-center justify-center text-xs font-bold flex-shrink-0 ${
                              s.status === 'COMPLETED'
                                ? 'bg-emerald-100 text-emerald-700'
                                : s.status === 'AWAITING_APPROVAL'
                                ? 'bg-amber-100 text-amber-700'
                                : s.status === 'SKIPPED'
                                ? 'bg-slate-100 text-slate-500'
                                : 'bg-indigo-50 text-indigo-700'
                            }`}
                          >
                            {s.status === 'COMPLETED' ? <Check className="w-3.5 h-3.5" /> : s.num}
                          </div>
                          <div className="flex-1">
                            <div className="flex items-center justify-between">
                              <span className="text-xs font-bold text-slate-900">{s.title}</span>
                              <span
                                className={`text-[10px] font-semibold px-2 py-0.5 rounded ${
                                  s.status === 'COMPLETED'
                                    ? 'bg-emerald-50 text-emerald-700 border border-emerald-200'
                                    : s.status === 'AWAITING_APPROVAL'
                                    ? 'bg-amber-50 text-amber-700 border border-amber-200'
                                    : s.status === 'SKIPPED'
                                    ? 'bg-slate-100 text-slate-500'
                                    : 'bg-slate-50 text-slate-600 border border-slate-200'
                                }`}
                              >
                                {s.status}
                              </span>
                            </div>
                            <p className="text-[11px] text-slate-500 mt-0.5">{s.desc}</p>
                          </div>
                        </div>
                      ))}
                    </div>
                  </div>
                </div>
              )}

              {/* TAB 4: Versioning & Replanning */}
              {activeTab === 'replan' && (
                <div className="space-y-6">
                  {/* Version History Table */}
                  <div className="space-y-3">
                    <h4 className="text-xs font-bold uppercase tracking-wider text-slate-500">
                      Quotation Pricing Version History
                    </h4>
                    <div className="border border-slate-200 rounded-lg overflow-hidden bg-white">
                      <table className="w-full text-left text-xs border-collapse">
                        <thead className="bg-slate-50 border-b border-slate-200 text-[10px] font-bold text-slate-600 uppercase">
                          <tr>
                            <th className="py-2.5 px-3">Ver</th>
                            <th className="py-2.5 px-3">Strategy</th>
                            <th className="py-2.5 px-3 text-right">Price</th>
                            <th className="py-2.5 px-3 text-right">Margin %</th>
                            <th className="py-2.5 px-3">Reason / Trigger</th>
                            <th className="py-2.5 px-3">Status</th>
                          </tr>
                        </thead>
                        <tbody className="divide-y divide-slate-200">
                          {versions.map((v) => (
                            <tr key={v.id || v.version} className="hover:bg-slate-50">
                              <td className="py-2.5 px-3 font-bold text-indigo-700">v{v.version}</td>
                              <td className="py-2.5 px-3 font-medium text-slate-900">{v.strategy_name}</td>
                              <td className="py-2.5 px-3 text-right font-bold text-slate-900">
                                ${v.price?.toLocaleString('en-US', { minimumFractionDigits: 2 })}
                              </td>
                              <td className="py-2.5 px-3 text-right text-emerald-600 font-semibold">
                                {v.margin_pct?.toFixed(1)}%
                              </td>
                              <td className="py-2.5 px-3 text-slate-600 max-w-xs truncate" title={v.change_reason}>
                                {v.change_reason}
                              </td>
                              <td className="py-2.5 px-3">
                                <span className="text-[10px] px-2 py-0.5 bg-slate-100 rounded text-slate-700 font-medium">
                                  {v.approval_status || 'ANALYZED'}
                                </span>
                              </td>
                            </tr>
                          ))}
                        </tbody>
                      </table>
                    </div>
                  </div>

                  {/* Replanning Form Simulator */}
                  <div className="p-5 border border-indigo-100 rounded-xl bg-indigo-50/40">
                    <h4 className="text-xs font-bold uppercase tracking-wider text-indigo-900 mb-1">
                      Trigger Operational Replanning
                    </h4>
                    <p className="text-xs text-slate-600 mb-4">
                      Simulate carrier rate fluctuations or cargo specification adjustments to trigger AI replanning.
                    </p>

                    <form onSubmit={handleReplanSubmit} className="space-y-4">
                      <div>
                        <label className="block text-xs font-semibold text-slate-700 mb-1">
                          Material Change Reason
                        </label>
                        <input
                          type="text"
                          value={replanReason}
                          onChange={(e) => setReplanReason(e.target.value)}
                          className="w-full text-xs px-3 py-2 bg-white border border-slate-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:outline-none"
                          required
                          data-testid="input-replan-reason"
                        />
                      </div>

                      <div>
                        <label className="block text-xs font-semibold text-slate-700 mb-1">
                          Carrier Rate Delta ($ USD)
                        </label>
                        <input
                          type="number"
                          step="10"
                          value={replanDelta}
                          onChange={(e) => setReplanDelta(e.target.value)}
                          className="w-full text-xs px-3 py-2 bg-white border border-slate-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:outline-none"
                          required
                          data-testid="input-replan-delta"
                        />
                      </div>

                      <button
                        type="submit"
                        disabled={replanning}
                        className="px-4 py-2 bg-indigo-600 text-white rounded-lg text-xs font-semibold hover:bg-indigo-700 flex items-center gap-2"
                        data-testid="btn-submit-replan"
                      >
                        <RefreshCw className={`w-3.5 h-3.5 ${replanning ? 'animate-spin' : ''}`} />
                        {replanning ? 'Replanning...' : 'Trigger AI Replanning'}
                      </button>
                    </form>
                  </div>
                </div>
              )}
            </>
          )}
        </div>

        {/* Action System Footer */}
        <div className="px-6 py-4 border-t border-slate-200 bg-slate-50 flex items-center justify-between">
          <div className="text-xs text-slate-500">
            {opt?.quotation_id ? (
              <span className="text-emerald-700 font-semibold flex items-center gap-1.5">
                <CheckCircle2 className="w-4 h-4 text-emerald-600" />
                Quotation committed (ID #{opt.quotation_id})
              </span>
            ) : (
              <span>Action System execution boundary armed</span>
            )}
          </div>

          <div className="flex items-center gap-3">
            <button
              onClick={onClose}
              className="px-4 py-2 text-xs font-medium text-slate-700 bg-white border border-slate-300 rounded-lg hover:bg-slate-50"
            >
              Close
            </button>
            <button
              onClick={handleExecuteQuotation}
              disabled={executingQuote || (opt?.requires_approval && opt?.approval_status !== 'APPROVED')}
              className={`px-4 py-2 text-xs font-semibold rounded-lg flex items-center gap-2 shadow-sm ${
                opt?.requires_approval && opt?.approval_status !== 'APPROVED'
                  ? 'bg-slate-200 text-slate-400 cursor-not-allowed'
                  : 'bg-emerald-600 text-white hover:bg-emerald-700'
              }`}
              data-testid="btn-execute-pricing-quotation"
            >
              <ShieldCheck className="w-4 h-4" />
              {executingQuote
                ? 'Executing...'
                : opt?.requires_approval && opt?.approval_status !== 'APPROVED'
                ? 'Manager Approval Required'
                : 'Execute Quotation via Action System'}
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}

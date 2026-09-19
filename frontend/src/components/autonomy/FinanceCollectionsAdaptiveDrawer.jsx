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
  Send,
  UserCheck,
  Calendar,
  Building2,
  Lock,
} from 'lucide-react';
import toast from 'react-hot-toast';
import { autonomyService } from '../../services/autonomyService';

export default function FinanceCollectionsAdaptiveDrawer({
  invoice,
  isOpen,
  onClose,
  onActionExecuted,
}) {
  const [activeTab, setActiveTab] = useState('strategies'); // 'strategies' | 'facts-predictions' | 'plan' | 'replan'
  const [state, setState] = useState(null);
  const [loading, setLoading] = useState(false);
  const [evaluating, setEvaluating] = useState(false);
  const [selectingStrategy, setSelectingStrategy] = useState(false);
  const [executingAction, setExecutingAction] = useState(false);
  const [replanning, setReplanning] = useState(false);

  // Replan form state
  const [replanTrigger, setReplanTrigger] = useState('PARTIAL_PAYMENT');
  const [replanAmount, setReplanAmount] = useState(1000.0);
  const [replanNotes, setReplanNotes] = useState('Customer processed wire transfer for partial invoice balance');

  const invoiceId = invoice?.id;

  const fetchCollectionState = useCallback(async () => {
    if (!invoiceId) return;
    setLoading(true);
    try {
      const resp = await autonomyService.getFinanceCollectionState(invoiceId);
      const data = resp?.state || resp?.data?.state || resp?.data || resp;
      setState(data);
    } catch (err) {
      console.error('Failed to load finance collection state:', err);
      toast.error('Failed to load collection optimization state');
    } finally {
      setLoading(false);
    }
  }, [invoiceId]);

  useEffect(() => {
    if (isOpen && invoiceId) {
      fetchCollectionState();
    }
  }, [isOpen, invoiceId, fetchCollectionState]);

  const handleRunEvaluation = async () => {
    if (!invoiceId) return;
    setEvaluating(true);
    try {
      const resp = await autonomyService.evaluateFinanceCollection(invoiceId);
      toast.success('Receivables evaluation & collection plan generated');
      const data = resp?.state || resp?.data?.state || resp?.data || resp;
      setState(data);
    } catch (err) {
      console.error('Evaluation failed:', err);
      toast.error(err.response?.data?.message || 'Collection evaluation failed');
    } finally {
      setEvaluating(false);
    }
  };

  const handleSelectStrategy = async (strategyId) => {
    if (!invoiceId) return;
    setSelectingStrategy(true);
    try {
      const resp = await autonomyService.selectFinanceCollectionStrategy(invoiceId, strategyId);
      toast.success('Collection strategy updated');
      const data = resp?.state || resp?.data?.state || resp?.data || resp;
      setState(data);
    } catch (err) {
      console.error('Failed to select strategy:', err);
      toast.error(err.response?.data?.message || 'Failed to select strategy');
    } finally {
      setSelectingStrategy(false);
    }
  };

  const handleExecuteAction = async () => {
    if (!invoiceId) return;
    setExecutingAction(true);
    try {
      const resp = await autonomyService.executeFinanceCollectionAction(invoiceId);
      toast.success('Collection action executed safely via Go Action System!');
      await fetchCollectionState();
      if (onActionExecuted) {
        onActionExecuted(resp);
      }
    } catch (err) {
      console.error('Failed to execute collection action:', err);
      toast.error(err.response?.data?.message || 'Execution failed');
    } finally {
      setExecutingAction(false);
    }
  };

  const handleReplan = async (e) => {
    e?.preventDefault();
    if (!invoiceId) return;
    setReplanning(true);
    try {
      const payload = {};
      if (replanTrigger === 'PARTIAL_PAYMENT' || replanTrigger === 'PAYMENT_RECEIVED') {
        payload.amount_paid = parseFloat(replanAmount) || 0.0;
      }
      if (replanTrigger === 'CUSTOMER_RESPONSE') {
        payload.response_text = replanNotes;
      }
      if (replanTrigger === 'DISPUTE_OPENED') {
        payload.dispute_reason = replanNotes;
      }

      const resp = await autonomyService.replanFinanceCollection(invoiceId, replanTrigger, payload);
      toast.success(`Plan adapted to v${resp?.state?.plan?.version || ''} upon ${replanTrigger}!`);
      const data = resp?.state || resp?.data?.state || resp?.data || resp;
      setState(data);
    } catch (err) {
      console.error('Replanning failed:', err);
      toast.error(err.response?.data?.message || 'Replanning failed');
    } finally {
      setReplanning(false);
    }
  };

  if (!isOpen) return null;

  const plan = state?.plan;
  const candidates = state?.candidates || [];
  const facts = state?.actual_facts || [];
  const predictions = state?.predictions || [];
  const assumptions = state?.assumptions || [];
  const planSteps = state?.plan_steps || [];
  const versions = state?.versions || [];
  const multiInvoiceSummary = state?.multi_invoice_summary;

  const invoiceNumber = state?.invoice_number || invoice?.invoiceNumber || `INV-${invoiceId}`;
  const customerName = state?.customer_name || invoice?.customer || 'Customer';
  const currency = state?.currency || invoice?.currency || 'USD';
  const totalAmount = state?.total_amount !== undefined ? state.total_amount : (invoice?.amount || 0.0);
  const balanceDue = state?.balance_due !== undefined ? state.balance_due : (invoice?.balance !== undefined ? invoice.balance : totalAmount);
  const daysOverdue = state?.days_overdue !== undefined ? state.days_overdue : (invoice?.daysOverdue || 0);
  const agingBucket = state?.aging_bucket || 'CURRENT';
  const invoiceStatus = state?.invoice_status || invoice?.status || 'Draft';
  const isSettled = balanceDue <= 0.0 || invoiceStatus.toLowerCase() === 'paid';
  const isDisputed = invoiceStatus.toLowerCase() === 'disputed';

  return (
    <div
      className="fixed inset-0 z-50 overflow-hidden flex justify-end bg-slate-900/40 backdrop-blur-sm transition-opacity"
      data-testid="finance-collection-drawer-overlay"
      onClick={onClose}
    >
      <div
        className="w-full max-w-3xl bg-white h-full shadow-2xl flex flex-col border-l border-slate-200 overflow-hidden"
        data-testid="finance-collection-drawer"
        onClick={(e) => e.stopPropagation()}
      >
        {/* Top Navigation Header */}
        <div className="p-4 border-b border-slate-200 flex items-center justify-between bg-slate-50">
          <div className="flex items-center gap-3">
            <div className="p-2 bg-blue-100 text-blue-700 rounded-lg">
              <DollarSign size={20} />
            </div>
            <div>
              <div className="flex items-center gap-2 flex-wrap">
                <h2 className="text-base font-bold text-slate-900 tracking-tight">
                  Adaptive Finance & Collections
                </h2>
                <span className="text-xs font-semibold px-2 py-0.5 rounded bg-blue-100 text-blue-800 border border-blue-200">
                  {invoiceNumber}
                </span>
                <span className={`text-xs font-semibold px-2 py-0.5 rounded border ${
                  isSettled
                    ? 'bg-emerald-100 text-emerald-800 border-emerald-200'
                    : isDisputed
                    ? 'bg-amber-100 text-amber-800 border-amber-200'
                    : daysOverdue > 0
                    ? 'bg-rose-100 text-rose-800 border-rose-200'
                    : 'bg-slate-100 text-slate-700 border-slate-200'
                }`}>
                  {invoiceStatus}
                </span>
                <span className="text-xs font-medium px-2 py-0.5 rounded bg-slate-200 text-slate-700">
                  Aging: {agingBucket} ({daysOverdue > 0 ? `${daysOverdue}d overdue` : 'Current'})
                </span>
              </div>
              <p className="text-xs text-slate-500 mt-0.5 flex items-center gap-2">
                <span>Customer: <strong className="text-slate-700">{customerName}</strong></span>
                <span>•</span>
                <span>Currency: <strong className="text-slate-700">{currency}</strong></span>
              </p>
            </div>
          </div>
          <div className="flex items-center gap-2">
            <button
              type="button"
              className="p-1.5 text-slate-400 hover:text-slate-600 hover:bg-slate-100 rounded-md transition-colors"
              onClick={handleRunEvaluation}
              disabled={evaluating || loading}
              title="Refresh AI Evaluation"
              data-testid="btn-run-evaluation"
            >
              <RefreshCw size={16} className={evaluating ? 'animate-spin text-blue-600' : ''} />
            </button>
            <button
              type="button"
              className="p-1.5 text-slate-400 hover:text-slate-600 hover:bg-slate-100 rounded-md transition-colors"
              onClick={onClose}
              data-testid="btn-close-finance-drawer"
            >
              <X size={18} />
            </button>
          </div>
        </div>

        {/* Financial KPI Strip: Segregating Actual Ledger vs Predicted vs Recommendation */}
        <div className="bg-white border-b border-slate-200 p-4 grid grid-cols-2 sm:grid-cols-4 gap-3 text-xs">
          {/* 1. Actual Ledger */}
          <div className="p-2.5 bg-slate-50 rounded-lg border border-slate-200">
            <span className="text-[10px] font-bold uppercase tracking-wider text-slate-500 block">
              [Actual] Outstanding Ledger
            </span>
            <div className="text-base font-extrabold text-slate-900 mt-0.5">
              ${Number(balanceDue).toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}
            </div>
            <span className="text-[11px] text-slate-500">
              Total: ${Number(totalAmount).toLocaleString('en-US', { minimumFractionDigits: 2 })} {currency}
            </span>
          </div>

          {/* 2. Predicted Risk */}
          <div className="p-2.5 bg-slate-50 rounded-lg border border-slate-200">
            <span className="text-[10px] font-bold uppercase tracking-wider text-slate-500 block">
              [Predicted] Collection Risk
            </span>
            <div className="flex items-center gap-1.5 mt-0.5">
              <span className={`text-base font-extrabold ${
                plan?.risk_level === 'CRITICAL' || plan?.risk_level === 'HIGH'
                  ? 'text-rose-700'
                  : plan?.risk_level === 'MEDIUM'
                  ? 'text-amber-700'
                  : 'text-emerald-700'
              }`}>
                {plan?.risk_level || 'LOW'}
              </span>
              <span className="text-[11px] text-slate-500">
                ({plan?.risk_score !== undefined ? Number(plan.risk_score).toFixed(2) : '0.25'})
              </span>
            </div>
            <span className="text-[11px] text-slate-500">
              Priority: <strong className="text-slate-700">{plan?.priority_level || 'MEDIUM'}</strong>
            </span>
          </div>

          {/* 3. Recommended Strategy */}
          <div className="p-2.5 bg-blue-50/50 rounded-lg border border-blue-200">
            <span className="text-[10px] font-bold uppercase tracking-wider text-blue-700 block">
              [Recommended] Strategy
            </span>
            <div className="text-xs font-bold text-slate-900 mt-1 truncate" title={plan?.selected_strategy_id}>
              {candidates.find(c => c.strategy_id === plan?.selected_strategy_id)?.strategy_name || 'Standard Follow-Up'}
            </div>
            <span className="text-[11px] text-blue-700 font-medium block mt-0.5">
              Action: {candidates.find(c => c.strategy_id === plan?.selected_strategy_id)?.recommended_action || 'WAIT_AND_MONITOR'}
            </span>
          </div>

          {/* 4. Autonomy & Version */}
          <div className="p-2.5 bg-slate-50 rounded-lg border border-slate-200">
            <span className="text-[10px] font-bold uppercase tracking-wider text-slate-500 block">
              [Governed] Autonomy Status
            </span>
            <div className="flex items-center gap-1 mt-1 flex-wrap">
              <span
                className={`text-[10px] font-bold px-1.5 py-0.5 rounded ${
                  plan?.status === 'EXECUTED'
                    ? 'bg-emerald-100 text-emerald-800'
                    : plan?.status === 'STOPPED'
                    ? 'bg-slate-200 text-slate-700'
                    : plan?.requires_approval
                    ? 'bg-amber-100 text-amber-800'
                    : 'bg-blue-100 text-blue-800'
                }`}
                data-testid="collection-status-badge"
              >
                {plan?.status || 'INITIAL'}
              </span>
              <span
                className="text-[10px] font-semibold px-1.5 py-0.5 rounded bg-slate-200 text-slate-700"
                data-testid="plan-version-badge"
              >
                v{plan?.version || 1}
              </span>
            </div>
            <span className="text-[10px] text-slate-500 block mt-1 truncate">
              {plan?.requires_approval ? 'Approval Required' : 'Pre-Approved Bounded'}
            </span>
          </div>
        </div>

        {/* Stop Condition Banner */}
        {plan?.stop_reason && (
          <div className="mx-4 mt-3 p-3 bg-emerald-50 border border-emerald-200 rounded-lg flex items-start gap-2.5">
            <CheckCircle2 size={16} className="text-emerald-700 shrink-0 mt-0.5" />
            <div className="text-xs text-emerald-900">
              <strong>Workflow Stopped:</strong> {plan.stop_reason}
            </div>
          </div>
        )}

        {/* Multi-Invoice Consolidation Alert */}
        {multiInvoiceSummary && (
          <div className="mx-4 mt-3 p-3 bg-blue-50 border border-blue-200 rounded-lg flex items-start gap-2.5">
            <Info size={16} className="text-blue-700 shrink-0 mt-0.5" />
            <div className="text-xs text-blue-900">
              <strong>Multi-Invoice Consolidation:</strong> {multiInvoiceSummary}
            </div>
          </div>
        )}

        {/* Dispute Alert */}
        {isDisputed && (
          <div className="mx-4 mt-3 p-3 bg-amber-50 border border-amber-200 rounded-lg flex items-start gap-2.5">
            <AlertTriangle size={16} className="text-amber-700 shrink-0 mt-0.5" />
            <div className="text-xs text-amber-900">
              <strong>Billing Dispute Active:</strong> Automated collection pressure is paused. Dispute investigation workflow active; review required before customer follow-up.
            </div>
          </div>
        )}

        {/* Tabs Navigation */}
        <div className="flex border-b border-slate-200 px-4 bg-slate-50 mt-2">
          <button
            type="button"
            className={`px-4 py-2.5 text-xs font-semibold border-b-2 transition-colors flex items-center gap-1.5 ${
              activeTab === 'strategies'
                ? 'border-blue-600 text-blue-700 bg-white'
                : 'border-transparent text-slate-600 hover:text-slate-900'
            }`}
            onClick={() => setActiveTab('strategies')}
            data-testid="tab-strategies"
          >
            <Layers size={14} />
            <span>Collection Strategies</span>
            <span className="ml-1 text-[10px] px-1.5 py-0.2 rounded-full bg-slate-200 text-slate-700">
              {candidates.length}
            </span>
          </button>

          <button
            type="button"
            className={`px-4 py-2.5 text-xs font-semibold border-b-2 transition-colors flex items-center gap-1.5 ${
              activeTab === 'facts-predictions'
                ? 'border-blue-600 text-blue-700 bg-white'
                : 'border-transparent text-slate-600 hover:text-slate-900'
            }`}
            onClick={() => setActiveTab('facts-predictions')}
            data-testid="tab-facts-predictions"
          >
            <ShieldCheck size={14} />
            <span>Facts vs Predictions</span>
          </button>

          <button
            type="button"
            className={`px-4 py-2.5 text-xs font-semibold border-b-2 transition-colors flex items-center gap-1.5 ${
              activeTab === 'plan'
                ? 'border-blue-600 text-blue-700 bg-white'
                : 'border-transparent text-slate-600 hover:text-slate-900'
            }`}
            onClick={() => setActiveTab('plan')}
            data-testid="tab-plan"
          >
            <Clock size={14} />
            <span>Autonomous Plan</span>
            <span className="ml-1 text-[10px] px-1.5 py-0.2 rounded-full bg-slate-200 text-slate-700">
              {planSteps.length} Steps
            </span>
          </button>

          <button
            type="button"
            className={`px-4 py-2.5 text-xs font-semibold border-b-2 transition-colors flex items-center gap-1.5 ${
              activeTab === 'replan'
                ? 'border-blue-600 text-blue-700 bg-white'
                : 'border-transparent text-slate-600 hover:text-slate-900'
            }`}
            onClick={() => setActiveTab('replan')}
            data-testid="tab-replan"
          >
            <RefreshCw size={14} />
            <span>Adaptation & History</span>
            <span className="ml-1 text-[10px] px-1.5 py-0.2 rounded-full bg-slate-200 text-slate-700">
              v{plan?.version || 1}
            </span>
          </button>
        </div>

        {/* Tab Content Body */}
        <div className="flex-1 overflow-y-auto p-4 space-y-4">
          {loading ? (
            <div className="flex flex-col items-center justify-center h-64 text-slate-400 gap-2">
              <RefreshCw className="animate-spin text-blue-600" size={24} />
              <span className="text-xs">Loading collection intelligence...</span>
            </div>
          ) : (
            <>
              {/* TAB 1: COLLECTION STRATEGIES */}
              {activeTab === 'strategies' && (
                <div className="space-y-3">
                  <div className="flex items-center justify-between text-xs text-slate-600">
                    <span>Ranked candidate strategies generated by Python AI reasoning</span>
                    <span className="text-slate-400">Select candidate to update active plan</span>
                  </div>

                  {candidates.map((cand) => {
                    const isSelected = plan?.selected_strategy_id === cand.strategy_id;
                    const isRecommended = plan?.recommended_strategy_id === cand.strategy_id;

                    return (
                      <div
                        key={cand.strategy_id}
                        className={`p-3.5 rounded-lg border transition-all ${
                          isSelected
                            ? 'border-blue-500 bg-blue-50/20 ring-1 ring-blue-400'
                            : 'border-slate-200 bg-white hover:border-slate-300'
                        }`}
                        data-testid={`strategy-card-${cand.strategy_id}`}
                      >
                        <div className="flex items-start justify-between gap-3">
                          <div>
                            <div className="flex items-center gap-2 flex-wrap">
                              <span className="font-bold text-sm text-slate-900">
                                {cand.strategy_name}
                              </span>
                              {isRecommended && (
                                <span className="text-[10px] font-bold px-2 py-0.5 rounded bg-blue-100 text-blue-800 border border-blue-200">
                                  ★ AI Recommended
                                </span>
                              )}
                              {isSelected && (
                                <span className="text-[10px] font-bold px-2 py-0.5 rounded bg-emerald-100 text-emerald-800 border border-emerald-200">
                                  ✓ Active Strategy
                                </span>
                              )}
                              <span className="text-[10px] font-semibold px-2 py-0.5 rounded bg-slate-100 text-slate-700">
                                Action: {cand.recommended_action}
                              </span>
                            </div>
                            <p className="text-xs text-slate-600 mt-1">
                              {cand.description}
                            </p>
                          </div>

                          {!isSelected && (
                            <button
                              type="button"
                              className="px-3 py-1 text-xs font-semibold rounded bg-slate-100 hover:bg-slate-200 text-slate-800 transition-colors shrink-0"
                              onClick={() => handleSelectStrategy(cand.strategy_id)}
                              disabled={selectingStrategy}
                              data-testid={`btn-select-strategy-${cand.strategy_id}`}
                            >
                              Select Strategy
                            </button>
                          )}
                        </div>

                        {/* Strategy Details Strip */}
                        <div className="mt-3 pt-2.5 border-t border-slate-100 grid grid-cols-2 sm:grid-cols-4 gap-2 text-xs text-slate-500">
                          <div>
                            <span className="text-[10px] block uppercase text-slate-400">Urgency</span>
                            <span className="font-medium text-slate-700">{cand.urgency || 'NORMAL'}</span>
                          </div>
                          <div>
                            <span className="text-[10px] block uppercase text-slate-400">Cooldown</span>
                            <span className="font-medium text-slate-700">{cand.cooldown_days} days</span>
                          </div>
                          <div>
                            <span className="text-[10px] block uppercase text-slate-400">Human Approval</span>
                            <span className={`font-medium ${cand.requires_approval ? 'text-amber-700' : 'text-slate-700'}`}>
                              {cand.requires_approval ? 'Required' : 'Pre-approved'}
                            </span>
                          </div>
                          <div>
                            <span className="text-[10px] block uppercase text-slate-400">Priority Score</span>
                            <span className="font-medium text-slate-700">{cand.priority_score} / 100</span>
                          </div>
                        </div>

                        {/* Draft Message Preview */}
                        {cand.draft_message && (
                          <div className="mt-2.5 p-2.5 bg-slate-50 rounded border border-slate-200 text-xs font-mono text-slate-700 whitespace-pre-wrap">
                            <div className="text-[10px] font-sans font-bold text-slate-500 uppercase tracking-wider mb-1">
                              Customer Message Draft ({cand.draft_subject}):
                            </div>
                            {cand.draft_message}
                          </div>
                        )}
                      </div>
                    );
                  })}
                </div>
              )}

              {/* TAB 2: FACTS VS PREDICTIONS SEGREGATION */}
              {activeTab === 'facts-predictions' && (
                <div className="space-y-4">
                  {/* Authoritative Facts */}
                  <div className="p-3.5 bg-slate-50 border border-slate-200 rounded-lg">
                    <div className="flex items-center gap-2 mb-2 text-slate-900 font-bold text-xs uppercase tracking-wider">
                      <ShieldCheck size={16} className="text-blue-600" />
                      <span>Authoritative Facts (Source of Truth - MariaDB Ledger)</span>
                    </div>
                    <ul className="space-y-1.5 text-xs text-slate-700">
                      {facts.map((fact, idx) => (
                        <li key={idx} className="flex items-start gap-2">
                          <span className="text-blue-600 font-bold">•</span>
                          <span>{fact}</span>
                        </li>
                      ))}
                    </ul>
                  </div>

                  {/* Predictions */}
                  <div className="p-3.5 bg-slate-50 border border-slate-200 rounded-lg">
                    <div className="flex items-center gap-2 mb-2 text-slate-900 font-bold text-xs uppercase tracking-wider">
                      <Sparkles size={16} className="text-amber-600" />
                      <span>Projections & Predictions (Estimated AI Reasoning)</span>
                    </div>
                    <ul className="space-y-1.5 text-xs text-slate-700">
                      {predictions.map((pred, idx) => (
                        <li key={idx} className="flex items-start gap-2">
                          <span className="text-amber-600 font-bold">•</span>
                          <span>{pred}</span>
                        </li>
                      ))}
                    </ul>
                  </div>

                  {/* Assumptions */}
                  <div className="p-3.5 bg-slate-50 border border-slate-200 rounded-lg">
                    <div className="flex items-center gap-2 mb-2 text-slate-900 font-bold text-xs uppercase tracking-wider">
                      <Info size={16} className="text-slate-500" />
                      <span>Operating Assumptions</span>
                    </div>
                    <ul className="space-y-1.5 text-xs text-slate-700">
                      {assumptions.map((assump, idx) => (
                        <li key={idx} className="flex items-start gap-2">
                          <span className="text-slate-400 font-bold">•</span>
                          <span>{assump}</span>
                        </li>
                      ))}
                    </ul>
                  </div>
                </div>
              )}

              {/* TAB 3: AUTONOMOUS 7-STEP EXECUTION PLAN */}
              {activeTab === 'plan' && (
                <div className="space-y-4">
                  <div className="flex items-center justify-between text-xs text-slate-600">
                    <span>Structured 7-step autonomous finance workflow</span>
                    <span className="text-slate-400">Bound by Corporate Autonomy Policies</span>
                  </div>

                  <div className="space-y-2.5">
                    {planSteps.map((step, idx) => (
                      <div
                        key={step.step_id || idx}
                        className="p-3 rounded-lg border border-slate-200 bg-white flex items-start gap-3 text-xs"
                      >
                        <div className="w-6 h-6 rounded-full bg-blue-100 text-blue-700 flex items-center justify-center font-bold text-xs shrink-0 mt-0.5">
                          {step.step_number || idx + 1}
                        </div>
                        <div className="flex-1">
                          <div className="flex items-center justify-between">
                            <span className="font-bold text-slate-900">
                              {step.title || step.action_type}
                            </span>
                            {step.requires_approval && (
                              <span className="text-[10px] font-bold px-1.5 py-0.5 rounded bg-amber-100 text-amber-800">
                                Approval Gate
                              </span>
                            )}
                          </div>
                          <p className="text-slate-600 mt-0.5">{step.description}</p>
                          {step.expected_outcome && (
                            <p className="text-[11px] text-slate-400 mt-1 italic">
                              Expected outcome: {step.expected_outcome}
                            </p>
                          )}
                        </div>
                      </div>
                    ))}
                  </div>

                  {/* Execution Action Footer */}
                  <div className="pt-3 border-t border-slate-200 flex items-center justify-between">
                    <div>
                      <span className="text-xs font-bold text-slate-900 block">
                        Action System Execution Boundary
                      </span>
                      <span className="text-[11px] text-slate-500">
                        {isSettled
                          ? 'Invoice is settled; no collection actions permitted.'
                          : plan?.requires_approval
                          ? 'Action requires finance controller sign-off.'
                          : 'Action is pre-approved for automated delivery.'}
                      </span>
                    </div>

                    <button
                      type="button"
                      className={`px-4 py-2 text-xs font-bold rounded-lg flex items-center gap-1.5 shadow-sm transition-all ${
                        isSettled
                          ? 'bg-slate-200 text-slate-400 cursor-not-allowed'
                          : 'bg-blue-600 hover:bg-blue-700 text-white'
                      }`}
                      onClick={handleExecuteAction}
                      disabled={isSettled || executingAction}
                      data-testid="btn-execute-collection-action"
                    >
                      <Send size={14} className={executingAction ? 'animate-spin' : ''} />
                      <span>{executingAction ? 'Executing Action...' : 'Execute via Action System'}</span>
                    </button>
                  </div>
                </div>
              )}

              {/* TAB 4: ADAPTATION, EVENTS & VERSION HISTORY */}
              {activeTab === 'replan' && (
                <div className="space-y-4">
                  {/* Event Simulation & Replanning Form */}
                  <form onSubmit={handleReplan} className="p-3.5 bg-slate-50 border border-slate-200 rounded-lg space-y-3">
                    <div className="flex items-center gap-2 text-slate-900 font-bold text-xs uppercase tracking-wider">
                      <RefreshCw size={14} className="text-blue-600" />
                      <span>Adapt Active Collection Plan to Business Events</span>
                    </div>

                    <div className="grid grid-cols-1 sm:grid-cols-2 gap-3 text-xs">
                      <div>
                        <label className="block text-slate-700 font-semibold mb-1">
                          Trigger Event
                        </label>
                        <select
                          className="w-full px-2.5 py-1.5 rounded border border-slate-300 bg-white text-slate-800"
                          value={replanTrigger}
                          onChange={(e) => setReplanTrigger(e.target.value)}
                          data-testid="replan-event-select"
                        >
                          <option value="PARTIAL_PAYMENT">PARTIAL_PAYMENT (Remittance Received)</option>
                          <option value="PAYMENT_RECEIVED">PAYMENT_RECEIVED (Full Settlement)</option>
                          <option value="CUSTOMER_RESPONSE">CUSTOMER_RESPONSE (Follow-up Feedback)</option>
                          <option value="DISPUTE_OPENED">DISPUTE_OPENED (Billing Dispute Flagged)</option>
                          <option value="PAYMENT_FAILED">PAYMENT_FAILED (Wire Bounced / Rejected)</option>
                        </select>
                      </div>

                      {(replanTrigger === 'PARTIAL_PAYMENT' || replanTrigger === 'PAYMENT_RECEIVED') && (
                        <div>
                          <label className="block text-slate-700 font-semibold mb-1">
                            Amount Paid ({currency})
                          </label>
                          <input
                            type="number"
                            step="0.01"
                            className="w-full px-2.5 py-1.5 rounded border border-slate-300 bg-white text-slate-800"
                            value={replanAmount}
                            onChange={(e) => setReplanAmount(e.target.value)}
                            data-testid="replan-amount-input"
                          />
                        </div>
                      )}
                    </div>

                    <div>
                      <label className="block text-slate-700 font-semibold mb-1 text-xs">
                        Event Context / Remittance Reference
                      </label>
                      <input
                        type="text"
                        className="w-full px-2.5 py-1.5 rounded border border-slate-300 bg-white text-slate-800 text-xs"
                        value={replanNotes}
                        onChange={(e) => setReplanNotes(e.target.value)}
                        data-testid="replan-notes-input"
                      />
                    </div>

                    <div className="flex justify-end">
                      <button
                        type="submit"
                        className="px-3.5 py-1.5 bg-blue-600 hover:bg-blue-700 text-white rounded text-xs font-semibold flex items-center gap-1.5 transition-colors"
                        disabled={replanning}
                        data-testid="btn-trigger-replan"
                      >
                        <RefreshCw size={12} className={replanning ? 'animate-spin' : ''} />
                        <span>{replanning ? 'Adapting Plan...' : 'Trigger Adaptation & Replan'}</span>
                      </button>
                    </div>
                  </form>

                  {/* Plan Version Audit History */}
                  <div>
                    <h3 className="text-xs font-bold uppercase tracking-wider text-slate-700 mb-2">
                      Immutable Plan Version Audit Trail ({versions.length} versions recorded)
                    </h3>

                    <div className="space-y-2">
                      {versions.map((ver) => (
                        <div
                          key={ver.id}
                          className="p-2.5 rounded border border-slate-200 bg-white text-xs flex items-center justify-between gap-3"
                          data-testid={`version-row-${ver.version_number}`}
                        >
                          <div>
                            <div className="flex items-center gap-2">
                              <span className="font-bold text-blue-700">
                                v{ver.version_number}
                              </span>
                              <span className="font-semibold px-1.5 py-0.5 rounded bg-slate-100 text-slate-700 text-[10px]">
                                {ver.trigger_event}
                              </span>
                              <span className="text-[10px] text-slate-400">
                                Status: <strong className="text-slate-600">{ver.status}</strong>
                              </span>
                            </div>
                            <p className="text-slate-600 mt-0.5 text-[11px]">
                              {ver.change_reason || 'Plan updated'}
                            </p>
                          </div>
                          <div className="text-right shrink-0">
                            <span className="font-bold text-slate-800 block">
                              ${Number(ver.balance_due).toLocaleString('en-US', { minimumFractionDigits: 2 })}
                            </span>
                            <span className="text-[10px] text-slate-400">
                              {ver.created_at ? new Date(ver.created_at).toLocaleTimeString() : ''}
                            </span>
                          </div>
                        </div>
                      ))}
                    </div>
                  </div>
                </div>
              )}
            </>
          )}
        </div>
      </div>
    </div>
  );
}

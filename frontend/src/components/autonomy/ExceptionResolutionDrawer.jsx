import React, { useState, useEffect, useCallback } from 'react';
import {
  X,
  AlertCircle,
  AlertTriangle,
  CheckCircle2,
  RefreshCw,
  Clock,
  Layers,
  ArrowRight,
  ShieldCheck,
  Check,
  UserCheck,
  HelpCircle,
  FileCheck,
  Activity,
  GitBranch,
  ShieldAlert,
  Send,
  Zap,
} from 'lucide-react';
import toast from 'react-hot-toast';
import autonomyService from '../../services/autonomyService';

export default function ExceptionResolutionDrawer({
  exception,
  isOpen,
  onClose,
  onActionExecuted,
}) {
  const [activeTab, setActiveTab] = useState('strategies'); // 'strategies' | 'evidence' | 'plan' | 'replan'
  const [state, setState] = useState(null);
  const [loading, setLoading] = useState(false);
  const [evaluating, setEvaluating] = useState(false);
  const [selectingStrategy, setSelectingStrategy] = useState(false);
  const [executingAction, setExecutingAction] = useState(false);
  const [replanning, setReplanning] = useState(false);

  // Replan form state
  const [replanTrigger, setReplanTrigger] = useState('CARRIER_UPDATE');
  const [replanPayloadValue, setReplanPayloadValue] = useState('2026-03-24');
  const [replanNotes, setReplanNotes] = useState('Carrier telemetry confirmed revised discharge ETA and berth allocation.');

  const exceptionId = exception?.id;

  const fetchExceptionState = useCallback(async () => {
    if (!exceptionId) return;
    setLoading(true);
    try {
      const resp = await autonomyService.getExceptionResolutionState(exceptionId);
      const data = resp?.state || resp?.data?.state || resp?.data || resp;
      setState(data);
    } catch (err) {
      console.error('Failed to load exception resolution state:', err);
      toast.error('Failed to load exception resolution state');
    } finally {
      setLoading(false);
    }
  }, [exceptionId]);

  useEffect(() => {
    if (isOpen && exceptionId) {
      fetchExceptionState();
    }
  }, [isOpen, exceptionId, fetchExceptionState]);

  const handleRunEvaluation = async () => {
    if (!exceptionId) return;
    setEvaluating(true);
    try {
      const resp = await autonomyService.evaluateExceptionResolution(exceptionId);
      toast.success('Operational exception reasoning evaluated successfully');
      const data = resp?.state || resp?.data?.state || resp?.data || resp;
      setState(data);
    } catch (err) {
      console.error('Exception evaluation failed:', err);
      toast.error('Failed to evaluate exception resolution');
    } finally {
      setEvaluating(false);
    }
  };

  const handleSelectStrategy = async (strategyId) => {
    if (!exceptionId || selectingStrategy) return;
    setSelectingStrategy(true);
    try {
      const resp = await autonomyService.selectExceptionResolutionStrategy(exceptionId, strategyId);
      toast.success('Recovery strategy updated successfully');
      const data = resp?.state || resp?.data?.state || resp?.data || resp;
      setState(data);
    } catch (err) {
      console.error('Failed selecting strategy:', err);
      toast.error(err?.response?.data?.message || 'Failed selecting candidate strategy');
    } finally {
      setSelectingStrategy(false);
    }
  };

  const handleExecuteAction = async () => {
    if (!exceptionId || executingAction) return;
    setExecutingAction(true);
    try {
      const resp = await autonomyService.executeExceptionResolutionAction(exceptionId);
      toast.success('Governed recovery action dispatched via Action System boundary');
      await fetchExceptionState();
      if (onActionExecuted) {
        onActionExecuted(resp);
      }
    } catch (err) {
      console.error('Action execution failed:', err);
      toast.error(err?.response?.data?.message || 'Failed executing recovery action');
    } finally {
      setExecutingAction(false);
    }
  };

  const handleReplan = async () => {
    if (!exceptionId || replanning) return;
    setReplanning(true);
    try {
      let payload = {};
      if (replanTrigger === 'CARRIER_UPDATE') {
        payload = { revised_eta: replanPayloadValue, carrier_notes: replanNotes };
      } else if (replanTrigger === 'DOCUMENT_SUBMITTED') {
        payload = { document_type: replanPayloadValue, broker_notes: replanNotes };
      } else if (replanTrigger === 'VERIFICATION_FAILED') {
        payload = { failure_reason: replanNotes };
      } else {
        payload = { shipper_notes: replanNotes };
      }

      const resp = await autonomyService.replanExceptionResolution(exceptionId, replanTrigger, payload);
      toast.success(`Exception resolution plan re-evaluated (Trigger: ${replanTrigger})`);
      const data = resp?.state || resp?.data?.state || resp?.data || resp;
      setState(data);
      setActiveTab('plan');
    } catch (err) {
      console.error('Replanning failed:', err);
      toast.error(err?.response?.data?.message || 'Failed replanning recovery workflow');
    } finally {
      setReplanning(false);
    }
  };

  if (!isOpen) return null;

  const plan = state?.plan;
  const candidates = state?.candidates || [];
  const versions = state?.versions || [];
  const steps = state?.recovery_plan_steps || [];
  const impact = state?.impact_assessment || {};
  const currentVersion = plan?.version || 1;
  const selectedStrategyId = plan?.selected_strategy_id || '';
  const waitingState = plan?.waiting_state || 'NONE';
  const lifecycleStatus = plan?.lifecycle_status || 'DETECTED';
  const requiresApproval = plan?.requires_approval || false;
  const severity = state?.severity || exception?.severity || 'MEDIUM';
  const excType = state?.exception_type || exception?.exception_type || 'OTHER';

  return (
    <div className="fixed inset-0 z-50 flex justify-end bg-slate-900/30 backdrop-blur-[1px] transition-opacity">
      <div className="flex h-full w-full max-w-4xl flex-col bg-white shadow-2xl border-l border-slate-200">
        {/* Top Bar */}
        <div className="flex items-center justify-between border-b border-slate-200 px-6 py-4 bg-slate-50/70">
          <div className="flex items-center space-x-3">
            <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-blue-50 border border-blue-200 text-blue-700">
              <ShieldAlert className="h-5 w-5" />
            </div>
            <div>
              <div className="flex items-center space-x-2">
                <span className="text-xs font-semibold uppercase tracking-wider text-slate-500">
                  Task 5.8 Controlled Autonomy
                </span>
                <span className="rounded bg-slate-200/80 px-2 py-0.5 text-xs font-mono font-medium text-slate-700">
                  Exception #{exceptionId}
                </span>
                <span
                  className={`rounded px-2 py-0.5 text-xs font-bold ${
                    severity === 'CRITICAL'
                      ? 'bg-rose-100 text-rose-800 border border-rose-200'
                      : severity === 'HIGH'
                      ? 'bg-amber-100 text-amber-800 border border-amber-200'
                      : 'bg-blue-100 text-blue-800 border border-blue-200'
                  }`}
                >
                  {severity}
                </span>
                <span className="rounded bg-slate-100 px-2 py-0.5 text-xs font-medium text-slate-700 border border-slate-200">
                  {excType}
                </span>
              </div>
              <h2 className="text-lg font-bold text-slate-900 line-clamp-1">
                {state?.title || exception?.title || 'Operational Exception Resolution'}
              </h2>
            </div>
          </div>

          <div className="flex items-center space-x-2">
            <button
              onClick={handleRunEvaluation}
              disabled={evaluating || loading}
              className="inline-flex items-center space-x-1.5 rounded-lg border border-slate-300 bg-white px-3 py-1.5 text-xs font-medium text-slate-700 shadow-sm hover:bg-slate-50 disabled:opacity-50"
              title="Re-run AI Root Cause & Recovery Synthesis"
            >
              <RefreshCw className={`h-3.5 w-3.5 ${evaluating ? 'animate-spin text-blue-600' : ''}`} />
              <span>{evaluating ? 'Evaluating...' : 'Re-Evaluate'}</span>
            </button>
            <button
              onClick={onClose}
              className="rounded-lg p-1.5 text-slate-400 hover:bg-slate-100 hover:text-slate-600"
            >
              <X className="h-5 w-5" />
            </button>
          </div>
        </div>

        {/* Operational KPI Strip */}
        <div className="grid grid-cols-4 gap-3 border-b border-slate-200 bg-white px-6 py-3 text-xs">
          <div className="rounded-md border border-slate-100 bg-slate-50/50 p-2.5">
            <div className="text-[11px] font-semibold text-slate-500 uppercase tracking-wide">
              [Authoritative] Lifecycle
            </div>
            <div className="mt-1 flex items-center space-x-1.5 font-bold text-slate-900">
              <span className="inline-block h-2 w-2 rounded-full bg-blue-500"></span>
              <span>{lifecycleStatus}</span>
            </div>
            <div className="mt-0.5 text-[10px] text-slate-500">Shipment #{state?.shipment_id || exception?.shipment_id}</div>
          </div>

          <div className="rounded-md border border-slate-100 bg-slate-50/50 p-2.5">
            <div className="text-[11px] font-semibold text-slate-500 uppercase tracking-wide">
              [Root Cause] Diagnosis
            </div>
            <div className="mt-1 font-semibold text-slate-900 line-clamp-1" title={state?.likely_root_cause}>
              {state?.likely_root_cause || 'Analyzing root cause...'}
            </div>
            <div className="mt-0.5 text-[10px] text-emerald-700 font-medium">
              Confidence: {Math.round((plan?.confidence || 0.88) * 100)}%
            </div>
          </div>

          <div className="rounded-md border border-slate-100 bg-slate-50/50 p-2.5">
            <div className="text-[11px] font-semibold text-slate-500 uppercase tracking-wide">
              [Predicted] Impact
            </div>
            <div className="mt-1 font-bold text-slate-900">
              Customer: <span className="text-amber-700">{impact?.customer_impact || 'MEDIUM'}</span>
            </div>
            <div className="mt-0.5 text-[10px] text-slate-500">
              Financial Risk: ${impact?.financial_impact_estimate || '250'}
            </div>
          </div>

          <div className="rounded-md border border-slate-100 bg-slate-50/50 p-2.5">
            <div className="text-[11px] font-semibold text-slate-500 uppercase tracking-wide">
              [Governed] Governance
            </div>
            <div className="mt-1 flex items-center space-x-1 font-bold">
              {requiresApproval ? (
                <span className="text-amber-800 flex items-center space-x-1">
                  <AlertTriangle className="h-3.5 w-3.5 text-amber-600" />
                  <span>Approval Required</span>
                </span>
              ) : (
                <span className="text-emerald-700 flex items-center space-x-1">
                  <ShieldCheck className="h-3.5 w-3.5 text-emerald-600" />
                  <span>Level 2 Governed</span>
                </span>
              )}
            </div>
            <div className="mt-0.5 text-[10px] text-blue-700 font-medium font-mono">
              Waiting: {waitingState}
            </div>
          </div>
        </div>

        {/* Tab Navigation */}
        <div className="flex border-b border-slate-200 bg-slate-50/50 px-6 text-sm">
          <button
            onClick={() => setActiveTab('strategies')}
            className={`flex items-center space-x-2 border-b-2 py-3 px-3 font-medium transition-colors ${
              activeTab === 'strategies'
                ? 'border-blue-600 text-blue-700'
                : 'border-transparent text-slate-600 hover:text-slate-900'
            }`}
          >
            <Zap className="h-4 w-4" />
            <span>Recovery Strategies ({candidates.length})</span>
          </button>
          <button
            onClick={() => setActiveTab('evidence')}
            className={`flex items-center space-x-2 border-b-2 py-3 px-3 font-medium transition-colors ${
              activeTab === 'evidence'
                ? 'border-blue-600 text-blue-700'
                : 'border-transparent text-slate-600 hover:text-slate-900'
            }`}
          >
            <Activity className="h-4 w-4" />
            <span>Root Cause & Impact Evidence</span>
          </button>
          <button
            onClick={() => setActiveTab('plan')}
            className={`flex items-center space-x-2 border-b-2 py-3 px-3 font-medium transition-colors ${
              activeTab === 'plan'
                ? 'border-blue-600 text-blue-700'
                : 'border-transparent text-slate-600 hover:text-slate-900'
            }`}
          >
            <Layers className="h-4 w-4" />
            <span>7-Step Recovery Plan (v{currentVersion})</span>
          </button>
          <button
            onClick={() => setActiveTab('replan')}
            className={`flex items-center space-x-2 border-b-2 py-3 px-3 font-medium transition-colors ${
              activeTab === 'replan'
                ? 'border-blue-600 text-blue-700'
                : 'border-transparent text-slate-600 hover:text-slate-900'
            }`}
          >
            <GitBranch className="h-4 w-4" />
            <span>Event Adaptation & Lineage ({versions.length})</span>
          </button>
        </div>

        {/* Tab Body */}
        <div className="flex-1 overflow-y-auto p-6 bg-slate-50/30">
          {loading ? (
            <div className="flex h-64 flex-col items-center justify-center space-y-3">
              <RefreshCw className="h-8 w-8 animate-spin text-blue-600" />
              <p className="text-sm font-medium text-slate-600">Loading exception intelligence...</p>
            </div>
          ) : (
            <>
              {/* TAB 1: RECOVERY STRATEGIES */}
              {activeTab === 'strategies' && (
                <div className="space-y-4">
                  <div className="rounded-lg border border-blue-200 bg-blue-50/70 p-3.5 text-xs text-blue-900">
                    <span className="font-bold">Controlled AI Decision Boundary:</span> The AI evaluates recovery
                    options based on expected resolution probability, time to recover, and risk. High-risk actions
                    strictly require human supervisor authorization.
                  </div>

                  <div className="space-y-3">
                    {candidates.map((cand) => {
                      const isSelected = cand.strategy_id === selectedStrategyId;
                      return (
                        <div
                          key={cand.strategy_id}
                          className={`rounded-lg border p-4 transition-all ${
                            isSelected
                              ? 'border-blue-500 bg-blue-50/40 shadow-sm ring-1 ring-blue-500'
                              : cand.is_feasible
                              ? 'border-slate-200 bg-white hover:border-slate-300'
                              : 'border-slate-200 bg-slate-100/70 opacity-75'
                          }`}
                        >
                          <div className="flex items-start justify-between">
                            <div className="space-y-1">
                              <div className="flex items-center space-x-2">
                                <span className="font-bold text-sm text-slate-900">{cand.strategy_name}</span>
                                <span className="rounded bg-slate-100 px-2 py-0.5 text-[10px] font-mono text-slate-600 border border-slate-200">
                                  {cand.strategy_type}
                                </span>
                                {isSelected && (
                                  <span className="inline-flex items-center rounded-full bg-blue-100 px-2 py-0.5 text-[10px] font-semibold text-blue-800">
                                    <Check className="mr-1 h-3 w-3" /> Selected Strategy
                                  </span>
                                )}
                                {!cand.is_feasible && (
                                  <span className="rounded-full bg-rose-100 px-2 py-0.5 text-[10px] font-semibold text-rose-800">
                                    Infeasible
                                  </span>
                                )}
                              </div>
                              <p className="text-xs text-slate-600">{cand.description}</p>
                            </div>

                            <div>
                              {isSelected ? (
                                <span className="rounded-md bg-blue-600 px-3 py-1.5 text-xs font-semibold text-white shadow-sm">
                                  Active
                                </span>
                              ) : cand.is_feasible ? (
                                <button
                                  onClick={() => handleSelectStrategy(cand.strategy_id)}
                                  disabled={selectingStrategy}
                                  className="rounded-md border border-slate-300 bg-white px-3 py-1.5 text-xs font-medium text-slate-700 hover:bg-slate-50 hover:text-slate-900 disabled:opacity-50"
                                >
                                  Select Plan
                                </button>
                              ) : (
                                <span className="text-[11px] text-slate-400 italic">Blocked</span>
                              )}
                            </div>
                          </div>

                          <div className="mt-3 grid grid-cols-4 gap-2 border-t border-slate-100 pt-3 text-[11px]">
                            <div>
                              <span className="text-slate-500">Expected Resolution:</span>{' '}
                              <span className="font-semibold text-emerald-700">
                                {Math.round(cand.expected_resolution_prob * 100)}%
                              </span>
                            </div>
                            <div>
                              <span className="text-slate-500">Resolution Time:</span>{' '}
                              <span className="font-medium text-slate-800">{cand.time_to_resolution}</span>
                            </div>
                            <div>
                              <span className="text-slate-500">Cost Impact:</span>{' '}
                              <span className="font-medium text-slate-800">${cand.cost_impact}</span>
                            </div>
                            <div>
                              <span className="text-slate-500">Approval Required:</span>{' '}
                              <span
                                className={`font-semibold ${
                                  cand.requires_approval ? 'text-amber-700' : 'text-slate-600'
                                }`}
                              >
                                {cand.requires_approval ? 'Yes (Supervisor)' : 'No (Pre-Authorized)'}
                              </span>
                            </div>
                          </div>

                          {cand.infeasibility_reason && (
                            <div className="mt-2 text-[11px] text-rose-700 bg-rose-50 p-2 rounded border border-rose-200">
                              <span className="font-semibold">Infeasibility Constraint:</span>{' '}
                              {cand.infeasibility_reason}
                            </div>
                          )}

                          {cand.expected_outcome && (
                            <div className="mt-2 text-[11px] text-slate-600 bg-slate-50 p-2 rounded">
                              <span className="font-semibold text-slate-700">Expected Outcome:</span>{' '}
                              {cand.expected_outcome}
                            </div>
                          )}
                        </div>
                      );
                    })}
                  </div>
                </div>
              )}

              {/* TAB 2: ROOT CAUSE & IMPACT EVIDENCE */}
              {activeTab === 'evidence' && (
                <div className="space-y-5">
                  {/* Root Cause vs Symptom */}
                  <div className="rounded-lg border border-slate-200 bg-white p-4 space-y-3">
                    <h3 className="text-xs font-bold uppercase tracking-wider text-slate-600">
                      Root Cause vs. Symptom Segregation
                    </h3>
                    <div className="grid grid-cols-2 gap-4 text-xs">
                      <div className="rounded-md border border-amber-200 bg-amber-50/50 p-3">
                        <div className="font-bold text-amber-900 flex items-center space-x-1.5">
                          <AlertTriangle className="h-4 w-4 text-amber-600" />
                          <span>Observed Symptom</span>
                        </div>
                        <p className="mt-1 text-amber-800">{state?.symptom || 'Analyzing observed symptoms...'}</p>
                      </div>

                      <div className="rounded-md border border-blue-200 bg-blue-50/50 p-3">
                        <div className="font-bold text-blue-900 flex items-center space-x-1.5">
                          <CheckCircle2 className="h-4 w-4 text-blue-600" />
                          <span>Identified Root Cause</span>
                        </div>
                        <p className="mt-1 text-blue-800">{state?.likely_root_cause || 'Deducing root cause...'}</p>
                      </div>
                    </div>

                    {state?.contributing_factors?.length > 0 && (
                      <div className="mt-2 pt-2 border-t border-slate-100">
                        <span className="text-[11px] font-semibold text-slate-700">Contributing Factors:</span>
                        <ul className="mt-1 list-disc pl-4 space-y-1 text-xs text-slate-600">
                          {state.contributing_factors.map((f, idx) => (
                            <li key={idx}>{f}</li>
                          ))}
                        </ul>
                      </div>
                    )}
                  </div>

                  {/* Corroborating Evidence */}
                  <div className="rounded-lg border border-slate-200 bg-white p-4 space-y-3">
                    <h3 className="text-xs font-bold uppercase tracking-wider text-slate-600">
                      Corroborating Evidence & Verifiable Facts
                    </h3>
                    <ul className="space-y-1.5 text-xs text-slate-700">
                      {state?.evidence?.map((ev, idx) => (
                        <li key={idx} className="flex items-start space-x-2">
                          <Check className="h-4 w-4 text-emerald-600 shrink-0 mt-0.5" />
                          <span>{ev}</span>
                        </li>
                      ))}
                    </ul>

                    {state?.unknown_factors?.length > 0 && (
                      <div className="mt-3 rounded border border-slate-100 bg-slate-50 p-2.5 text-xs text-slate-600">
                        <span className="font-semibold text-slate-700">Unknown Factors (Pending Verification):</span>
                        <ul className="mt-1 list-disc pl-4 space-y-0.5 text-[11px] text-slate-500">
                          {state.unknown_factors.map((u, idx) => (
                            <li key={idx}>{u}</li>
                          ))}
                        </ul>
                      </div>
                    )}
                  </div>

                  {/* Impact Analysis: Confirmed vs Predicted vs Possible */}
                  <div className="rounded-lg border border-slate-200 bg-white p-4 space-y-3">
                    <h3 className="text-xs font-bold uppercase tracking-wider text-slate-600">
                      Multi-Tier Impact Assessment
                    </h3>
                    <div className="grid grid-cols-3 gap-3 text-xs">
                      <div className="rounded border border-slate-200 bg-slate-50 p-3">
                        <span className="font-bold text-slate-800">[Confirmed] Authoritative</span>
                        <ul className="mt-1.5 space-y-1 text-[11px] text-slate-600">
                          {impact?.confirmed_impact?.map((c, idx) => (
                            <li key={idx}>• {c}</li>
                          )) || <li>• Active exception condition verified</li>}
                        </ul>
                      </div>

                      <div className="rounded border border-slate-200 bg-slate-50 p-3">
                        <span className="font-bold text-slate-800">[Predicted] Forecasted</span>
                        <ul className="mt-1.5 space-y-1 text-[11px] text-slate-600">
                          {impact?.predicted_impact?.map((p, idx) => (
                            <li key={idx}>• {p}</li>
                          )) || <li>• SLA buffer deviation expected</li>}
                        </ul>
                      </div>

                      <div className="rounded border border-slate-200 bg-slate-50 p-3">
                        <span className="font-bold text-slate-800">[Possible] Downstream Risk</span>
                        <ul className="mt-1.5 space-y-1 text-[11px] text-slate-600">
                          {impact?.possible_impact?.map((pos, idx) => (
                            <li key={idx}>• {pos}</li>
                          )) || <li>• Intermodal transfer risk</li>}
                        </ul>
                      </div>
                    </div>
                  </div>

                  {/* Hard Constraints */}
                  {state?.hard_constraints?.length > 0 && (
                    <div className="rounded-lg border border-rose-200 bg-rose-50/40 p-4">
                      <h3 className="text-xs font-bold uppercase tracking-wider text-rose-900 flex items-center space-x-1.5">
                        <ShieldAlert className="h-4 w-4 text-rose-600" />
                        <span>Enforced Hard Constraints</span>
                      </h3>
                      <ul className="mt-2 space-y-1 text-xs text-rose-800">
                        {state.hard_constraints.map((c, idx) => (
                          <li key={idx}>• {c}</li>
                        ))}
                      </ul>
                    </div>
                  )}
                </div>
              )}

              {/* TAB 3: 7-STEP RECOVERY PLAN */}
              {activeTab === 'plan' && (
                <div className="space-y-5">
                  <div className="flex items-center justify-between rounded-lg border border-slate-200 bg-white p-4">
                    <div>
                      <span className="text-xs font-semibold text-slate-500 uppercase tracking-wide">
                        Selected Recovery Plan
                      </span>
                      <h3 className="text-sm font-bold text-slate-900">{plan?.selected_strategy_id}</h3>
                      <p className="text-xs text-slate-600 mt-0.5">
                        Status: <span className="font-semibold text-blue-700">{lifecycleStatus}</span> • Waiting:{' '}
                        <span className="font-mono text-slate-700">{waitingState}</span>
                      </p>
                    </div>

                    <div className="flex items-center space-x-2">
                      <button
                        onClick={handleExecuteAction}
                        disabled={executingAction}
                        className="inline-flex items-center space-x-1.5 rounded-lg bg-blue-600 px-4 py-2 text-xs font-semibold text-white shadow hover:bg-blue-700 disabled:opacity-50"
                      >
                        <Send className="h-3.5 w-3.5" />
                        <span>{executingAction ? 'Dispatching...' : 'Execute Governed Action'}</span>
                      </button>
                    </div>
                  </div>

                  {/* 7 Ordered Steps */}
                  <div className="rounded-lg border border-slate-200 bg-white p-4">
                    <h3 className="text-xs font-bold uppercase tracking-wider text-slate-600 mb-3">
                      Sequential Recovery Execution Roadmap
                    </h3>

                    <div className="space-y-3">
                      {steps.map((step) => {
                        const isCompleted = step.status === 'COMPLETED';
                        return (
                          <div
                            key={step.step_id || step.step_number}
                            className={`flex items-start space-x-3 rounded-lg border p-3 text-xs ${
                              isCompleted
                                ? 'border-emerald-200 bg-emerald-50/30'
                                : step.requires_approval
                                ? 'border-amber-200 bg-amber-50/20'
                                : 'border-slate-200 bg-white'
                            }`}
                          >
                            <div
                              className={`flex h-6 w-6 shrink-0 items-center justify-center rounded-full text-xs font-bold ${
                                isCompleted
                                  ? 'bg-emerald-600 text-white'
                                  : 'bg-slate-200 text-slate-700'
                              }`}
                            >
                              {isCompleted ? <Check className="h-3.5 w-3.5" /> : step.step_number}
                            </div>

                            <div className="flex-1 space-y-0.5">
                              <div className="flex items-center justify-between">
                                <span className="font-semibold text-slate-900">{step.name}</span>
                                <span className="rounded bg-slate-100 px-1.5 py-0.5 text-[10px] font-mono text-slate-600 border">
                                  {step.action_type}
                                </span>
                              </div>
                              <p className="text-slate-600 text-[11px]">{step.description}</p>
                            </div>
                          </div>
                        );
                      })}
                    </div>
                  </div>

                  {/* Verification Criteria */}
                  {state?.verification_criteria && (
                    <div className="rounded-lg border border-slate-200 bg-slate-50 p-4 text-xs">
                      <span className="font-bold text-slate-800">Authoritative Verification Gate:</span>
                      <div className="mt-2 grid grid-cols-2 gap-2 text-[11px] text-slate-600">
                        <div>
                          <span className="text-slate-500">Required Condition:</span>{' '}
                          <span className="font-mono font-medium text-slate-800">
                            {state.verification_criteria.condition_cleared_check || 'shipment_exceptions.resolved == 1'}
                          </span>
                        </div>
                        <div>
                          <span className="text-slate-500">Fallback on Failure:</span>{' '}
                          <span className="font-medium text-slate-800">
                            {state.verification_criteria.fallback_action || 'TRIGGER_REPLANNING_OR_ESCALATE'}
                          </span>
                        </div>
                      </div>
                    </div>
                  )}
                </div>
              )}

              {/* TAB 4: EVENT ADAPTATION & REPLANNING */}
              {activeTab === 'replan' && (
                <div className="space-y-5">
                  {/* Replan Form */}
                  <div className="rounded-lg border border-slate-200 bg-white p-4 space-y-4">
                    <div>
                      <h3 className="text-xs font-bold uppercase tracking-wider text-slate-800">
                        Trigger Adaptive Replanning Loop
                      </h3>
                      <p className="text-xs text-slate-500 mt-0.5">
                        When an operational event (carrier ETA revision, document submission, verification outcome) occurs,
                        the AI adapts the recovery plan into a new version.
                      </p>
                    </div>

                    <div className="space-y-3 text-xs">
                      <div>
                        <label className="font-semibold text-slate-700">Trigger Event</label>
                        <select
                          value={replanTrigger}
                          onChange={(e) => setReplanTrigger(e.target.value)}
                          className="mt-1 w-full rounded-md border border-slate-300 bg-white p-2 text-xs text-slate-900 focus:border-blue-500 focus:outline-none"
                        >
                          <option value="CARRIER_UPDATE">CARRIER_UPDATE (Carrier revised transit ETA)</option>
                          <option value="DOCUMENT_SUBMITTED">DOCUMENT_SUBMITTED (Broker filed amended document)</option>
                          <option value="CUSTOMER_CONFIRMED">CUSTOMER_CONFIRMED (Shipper accepted schedule buffer)</option>
                          <option value="VERIFICATION_FAILED">VERIFICATION_FAILED (Remediation milestone did not clear)</option>
                        </select>
                      </div>

                      <div>
                        <label className="font-semibold text-slate-700">Event Payload / Value</label>
                        <input
                          type="text"
                          value={replanPayloadValue}
                          onChange={(e) => setReplanPayloadValue(e.target.value)}
                          placeholder="e.g. 2026-03-24 or COMMERCIAL_INVOICE_AMENDED"
                          className="mt-1 w-full rounded-md border border-slate-300 p-2 text-xs text-slate-900 focus:border-blue-500 focus:outline-none"
                        />
                      </div>

                      <div>
                        <label className="font-semibold text-slate-700">Operational Corroboration Notes</label>
                        <textarea
                          rows={2}
                          value={replanNotes}
                          onChange={(e) => setReplanNotes(e.target.value)}
                          className="mt-1 w-full rounded-md border border-slate-300 p-2 text-xs text-slate-900 focus:border-blue-500 focus:outline-none"
                        />
                      </div>

                      <div className="pt-2">
                        <button
                          onClick={handleReplan}
                          disabled={replanning}
                          className="inline-flex items-center space-x-1.5 rounded-lg bg-blue-600 px-4 py-2 text-xs font-semibold text-white shadow hover:bg-blue-700 disabled:opacity-50"
                        >
                          <RefreshCw className={`h-3.5 w-3.5 ${replanning ? 'animate-spin' : ''}`} />
                          <span>{replanning ? 'Adapting Plan...' : 'Re-Evaluate & Create Plan Version'}</span>
                        </button>
                      </div>
                    </div>
                  </div>

                  {/* Version Lineage */}
                  <div className="rounded-lg border border-slate-200 bg-white p-4 space-y-3">
                    <h3 className="text-xs font-bold uppercase tracking-wider text-slate-700">
                      Immutable Plan Version Lineage
                    </h3>
                    <div className="space-y-2">
                      {versions.map((ver) => (
                        <div
                          key={ver.id || ver.version_number}
                          className="flex items-center justify-between rounded border border-slate-100 bg-slate-50 p-2.5 text-xs"
                        >
                          <div className="flex items-center space-x-2">
                            <span className="font-bold text-blue-700 font-mono">v{ver.version_number}</span>
                            <span className="rounded bg-slate-200 px-1.5 py-0.5 text-[10px] font-mono text-slate-700">
                              {ver.trigger_event}
                            </span>
                            <span className="text-slate-600 text-[11px] line-clamp-1">
                              {ver.change_reason || 'Plan updated'}
                            </span>
                          </div>
                          <span className="text-[10px] text-slate-400 font-mono">
                            {new Date(ver.created_at).toLocaleTimeString()}
                          </span>
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

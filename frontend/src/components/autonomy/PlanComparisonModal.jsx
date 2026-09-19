import React, { useState } from 'react';
import { X, CheckCircle, AlertTriangle, ArrowRight, ShieldCheck, Zap, DollarSign, Clock, HelpCircle } from 'lucide-react';
import autonomyService from '../../services/autonomyService';

export const PlanComparisonModal = ({ isOpen, onClose, plan, onCandidateSelected }) => {
  const [selectedCandidate, setSelectedCandidate] = useState(plan?.selected_candidate_id || '');
  const [submitting, setSubmitting] = useState(false);
  const [errorMsg, setErrorMsg] = useState('');

  if (!isOpen || !plan) return null;

  let candidates = [];
  try {
    if (typeof plan.candidate_plans === 'string') {
      candidates = JSON.parse(plan.candidate_plans || '[]');
    } else if (Array.isArray(plan.candidate_plans)) {
      candidates = plan.candidate_plans;
    }
  } catch (e) {
    candidates = [];
  }

  const handleSelect = async (candidateId) => {
    setSubmitting(true);
    setErrorMsg('');
    try {
      await autonomyService.selectCandidate(plan.plan_id, candidateId, 'Selected by operations supervisor via comparison view');
      setSelectedCandidate(candidateId);
      if (onCandidateSelected) {
        onCandidateSelected(candidateId);
      }
      onClose();
    } catch (err) {
      setErrorMsg(err.response?.data?.message || err.message || 'Failed to select candidate strategy');
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 overflow-y-auto bg-slate-900/50 backdrop-blur-sm flex items-center justify-center p-4">
      <div className="bg-white rounded-xl shadow-xl border border-slate-200 w-full max-w-5xl overflow-hidden flex flex-col max-h-[90vh]">
        {/* Header */}
        <div className="flex items-center justify-between px-6 py-4 border-b border-slate-100 bg-slate-50/50">
          <div>
            <div className="flex items-center gap-2">
              <Zap className="w-5 h-5 text-indigo-600" />
              <h2 className="text-lg font-semibold text-slate-900">Compare Operational Strategies</h2>
              <span className="text-xs px-2 py-0.5 rounded-full bg-slate-200 text-slate-700 font-mono">
                {plan.plan_id}
              </span>
            </div>
            <p className="text-xs text-slate-500 mt-1">
              Multi-attribute candidate evaluation for {plan.module} entity #{plan.related_entity_id}. Hard constraints are strictly enforced.
            </p>
          </div>
          <button
            onClick={onClose}
            className="text-slate-400 hover:text-slate-600 p-1.5 rounded-lg hover:bg-slate-100 transition-colors"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        {/* Error banner */}
        {errorMsg && (
          <div className="mx-6 mt-4 p-3 rounded-lg bg-red-50 border border-red-200 flex items-center gap-2 text-xs text-red-700">
            <AlertTriangle className="w-4 h-4 flex-shrink-0" />
            <span>{errorMsg}</span>
          </div>
        )}

        {/* Candidates Grid */}
        <div className="p-6 overflow-y-auto flex-1">
          {candidates.length === 0 ? (
            <div className="text-center py-12 text-slate-500 text-sm">
              No candidate alternatives generated for this plan version.
            </div>
          ) : (
            <div className="grid grid-cols-1 md:grid-cols-3 gap-5">
              {candidates.map((cand, idx) => {
                const evalData = cand.evaluation || {};
                const isSelected = (cand.candidate_id === plan.selected_candidate_id) || (cand.candidate_id === selectedCandidate);
                const isFeasible = cand.is_feasible && evalData.hard_constraints_satisfied !== false;

                return (
                  <div
                    key={cand.candidate_id || idx}
                    className={`rounded-xl border p-5 flex flex-col justify-between transition-all ${
                      isSelected
                        ? 'border-indigo-600 bg-indigo-50/30 ring-2 ring-indigo-500/20'
                        : isFeasible
                        ? 'border-slate-200 bg-white hover:border-slate-300'
                        : 'border-red-200 bg-red-50/20 opacity-80'
                    }`}
                  >
                    <div>
                      {/* Top Bar: Rank & Status */}
                      <div className="flex items-center justify-between mb-3">
                        <span className="text-xs font-semibold px-2 py-0.5 rounded bg-slate-100 text-slate-700">
                          Rank #{cand.rank || idx + 1}
                        </span>
                        {isFeasible ? (
                          <span className="inline-flex items-center gap-1 text-[11px] font-medium px-2 py-0.5 rounded-full bg-emerald-50 text-emerald-700 border border-emerald-200">
                            <CheckCircle className="w-3 h-3" /> Feasible
                          </span>
                        ) : (
                          <span className="inline-flex items-center gap-1 text-[11px] font-medium px-2 py-0.5 rounded-full bg-red-50 text-red-700 border border-red-200">
                            <AlertTriangle className="w-3 h-3" /> Hard Constraint Exceeded
                          </span>
                        )}
                      </div>

                      {/* Strategy Title & Summary */}
                      <h3 className="text-sm font-bold text-slate-900 mb-1">{cand.strategy_name}</h3>
                      <p className="text-xs text-slate-600 mb-4 leading-relaxed">{cand.summary}</p>

                      {/* Metrics comparison */}
                      <div className="space-y-2 mb-4 bg-slate-50 rounded-lg p-3 border border-slate-100">
                        <div className="flex items-center justify-between text-xs">
                          <span className="text-slate-500 flex items-center gap-1">
                            <Clock className="w-3.5 h-3.5 text-slate-400" /> Delay Recovery:
                          </span>
                          <span className="font-semibold text-slate-900">
                            {evalData.delay_reduction_hours ? `${evalData.delay_reduction_hours.toFixed(1)}h` : '—'}
                          </span>
                        </div>
                        <div className="flex items-center justify-between text-xs">
                          <span className="text-slate-500 flex items-center gap-1">
                            <DollarSign className="w-3.5 h-3.5 text-slate-400" /> Incremental Cost:
                          </span>
                          <span className={`font-semibold ${evalData.estimated_cost > 0 ? 'text-amber-700' : 'text-slate-900'}`}>
                            ${(evalData.estimated_cost || 0).toFixed(2)}
                          </span>
                        </div>
                        <div className="flex items-center justify-between text-xs">
                          <span className="text-slate-500 flex items-center gap-1">
                            <ShieldCheck className="w-3.5 h-3.5 text-slate-400" /> Operational Risk:
                          </span>
                          <span className={`font-semibold ${evalData.operational_risk === 'LOW' ? 'text-emerald-700' : 'text-amber-700'}`}>
                            {evalData.operational_risk || 'LOW'}
                          </span>
                        </div>
                        <div className="flex items-center justify-between text-xs">
                          <span className="text-slate-500">Utility Score:</span>
                          <span className="font-bold text-indigo-700">
                            {evalData.overall_utility_score ? (evalData.overall_utility_score * 100).toFixed(0) + '%' : '0%'}
                          </span>
                        </div>
                      </div>

                      {/* Operational Rationale / Penalties */}
                      {evalData.selection_rationale && (
                        <div className="mb-4 text-[11px] text-slate-500 italic bg-white p-2.5 rounded border border-slate-100">
                          "{evalData.selection_rationale}"
                        </div>
                      )}
                      {evalData.rejection_or_penalty_reasons && evalData.rejection_or_penalty_reasons.length > 0 && (
                        <div className="mb-4 text-[11px] text-red-600 bg-red-50/60 p-2.5 rounded border border-red-100">
                          {evalData.rejection_or_penalty_reasons.join(', ')}
                        </div>
                      )}
                    </div>

                    {/* Action Button */}
                    <div className="pt-3 border-t border-slate-100">
                      {isSelected ? (
                        <div className="w-full py-2 text-center text-xs font-semibold text-indigo-700 bg-indigo-50 rounded-lg border border-indigo-200 flex items-center justify-center gap-1.5">
                          <CheckCircle className="w-4 h-4" /> Selected Plan
                        </div>
                      ) : isFeasible ? (
                        <button
                          onClick={() => handleSelect(cand.candidate_id)}
                          disabled={submitting}
                          className="w-full py-2 text-xs font-semibold text-slate-700 bg-white hover:bg-slate-50 hover:text-slate-900 border border-slate-300 rounded-lg transition-colors flex items-center justify-center gap-1.5"
                        >
                          Select this Strategy <ArrowRight className="w-3.5 h-3.5" />
                        </button>
                      ) : (
                        <div className="w-full py-2 text-center text-xs font-medium text-slate-400 bg-slate-100 rounded-lg cursor-not-allowed">
                          Ineligible (Violates Policy)
                        </div>
                      )}
                    </div>
                  </div>
                );
              })}
            </div>
          )}
        </div>

        {/* Footer */}
        <div className="px-6 py-3.5 border-t border-slate-100 bg-slate-50/50 flex items-center justify-between">
          <div className="text-xs text-slate-500 flex items-center gap-1.5">
            <HelpCircle className="w-4 h-4 text-slate-400" />
            Candidate selection is persisted durably and logged in the immutable audit trail.
          </div>
          <button
            onClick={onClose}
            className="px-4 py-2 text-xs font-medium text-slate-700 bg-white border border-slate-200 hover:bg-slate-50 rounded-lg transition-colors"
          >
            Close
          </button>
        </div>
      </div>
    </div>
  );
};

export default PlanComparisonModal;

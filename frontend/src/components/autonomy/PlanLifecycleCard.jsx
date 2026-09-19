import React, { useState } from 'react';
import { Zap, RefreshCw, CheckCircle, AlertTriangle, ShieldCheck, ArrowRight } from 'lucide-react';
import PlanComparisonModal from './PlanComparisonModal';
import autonomyService from '../../services/autonomyService';

/**
 * PlanLifecycleCard.jsx — Phase 5 Intelligent Operational Planning Card
 *
 * Displays operational plan lifecycle, candidate comparison trigger,
 * staleness status, step dependencies, condition predicates, and verification criteria.
 */

export function PlanLifecycleCard({ plan, steps = [], onApprove, onReject, onExecuteStep, onPlanUpdated }) {
  const [showComparison, setShowComparison] = useState(false);
  const [revalidating, setRevalidating] = useState(false);

  if (!plan) return null;

  const getStatusBadge = (status) => {
    switch (status) {
      case 'APPROVED':
      case 'COMPLETED':
        return 'bg-emerald-50 text-emerald-700 border-emerald-200';
      case 'REQUIRES_APPROVAL':
        return 'bg-amber-50 text-amber-700 border-amber-200';
      case 'EXECUTING':
        return 'bg-blue-50 text-blue-700 border-blue-200';
      case 'FAILED':
      case 'REJECTED':
        return 'bg-red-50 text-red-700 border-red-200';
      case 'PAUSED':
        return 'bg-gray-100 text-gray-700 border-gray-300';
      case 'REPLANNING':
        return 'bg-indigo-50 text-indigo-700 border-indigo-200';
      default:
        return 'bg-slate-50 text-slate-700 border-slate-200';
    }
  };

  const formatLevel = (lvl) => {
    switch (lvl) {
      case 'LEVEL_0_OBSERVE': return 'Level 0: Observe';
      case 'LEVEL_1_RECOMMEND': return 'Level 1: Recommend';
      case 'LEVEL_2_PREPARE': return 'Level 2: Prepare';
      case 'LEVEL_3_CONTROLLED_EXECUTION': return 'Level 3: Controlled Execution';
      case 'LEVEL_4_CONTROLLED_MULTI_STEP': return 'Level 4: Multi-Step Autonomy';
      default: return lvl;
    }
  };

  const handleRevalidate = async () => {
    setRevalidating(true);
    try {
      await autonomyService.revalidatePlan(plan.plan_id);
      if (onPlanUpdated) {
        onPlanUpdated();
      }
    } catch (err) {
      console.error('Failed to revalidate plan:', err);
    } finally {
      setRevalidating(false);
    }
  };

  let candidatesCount = 0;
  try {
    if (typeof plan.candidate_plans === 'string') {
      const parsed = JSON.parse(plan.candidate_plans || '[]');
      candidatesCount = parsed.length;
    } else if (Array.isArray(plan.candidate_plans)) {
      candidatesCount = plan.candidate_plans.length;
    }
  } catch (e) {
    candidatesCount = 0;
  }

  const isStale = plan.staleness_status === 'REVALIDATION_REQUIRED' || plan.staleness_status === 'STALE_DATA';

  return (
    <div className="bg-white border border-slate-200 rounded-lg shadow-sm p-6 space-y-5" data-testid="autonomy-plan-card">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 border-b border-slate-100 pb-4">
        <div>
          <div className="flex flex-wrap items-center gap-2">
            <span className="text-xs font-semibold px-2 py-0.5 rounded border uppercase tracking-wider bg-slate-100 text-slate-700">
              {formatLevel(plan.autonomy_level)}
            </span>
            <span className={`text-xs font-semibold px-2.5 py-0.5 rounded-full border ${getStatusBadge(plan.status)}`}>
              {plan.status}
            </span>
            <span className="text-xs text-slate-400 font-mono">v{plan.version}</span>
            {/* Staleness Pill */}
            <span className={`text-[11px] font-medium px-2 py-0.5 rounded-full border flex items-center gap-1 ${
              isStale ? 'bg-amber-50 text-amber-700 border-amber-200' : 'bg-slate-50 text-slate-600 border-slate-200'
            }`}>
              {isStale ? <AlertTriangle className="w-3 h-3 text-amber-600" /> : <CheckCircle className="w-3 h-3 text-emerald-600" />}
              {isStale ? 'Revalidation Needed' : 'Telemetry Fresh'}
            </span>
          </div>
          <h3 className="text-base font-semibold text-slate-900 mt-2">{plan.goal}</h3>
        </div>

        {/* Action Controls & Candidate Comparison */}
        <div className="flex items-center gap-2 shrink-0 flex-wrap">
          {candidatesCount > 0 && (
            <button
              onClick={() => setShowComparison(true)}
              className="px-3 py-1.5 text-xs font-medium text-indigo-700 bg-indigo-50 hover:bg-indigo-100 border border-indigo-200 rounded transition-colors flex items-center gap-1.5"
            >
              <Zap className="w-3.5 h-3.5 text-indigo-600" /> Compare Alternatives ({candidatesCount})
            </button>
          )}
          <button
            onClick={handleRevalidate}
            disabled={revalidating}
            className="px-2.5 py-1.5 text-xs font-medium text-slate-600 hover:text-slate-900 bg-slate-50 hover:bg-slate-100 border border-slate-200 rounded transition-colors flex items-center gap-1"
            title="Revalidate plan against live database state"
          >
            <RefreshCw className={`w-3.5 h-3.5 ${revalidating ? 'animate-spin text-slate-700' : 'text-slate-400'}`} />
            Revalidate
          </button>
          {plan.status === 'REQUIRES_APPROVAL' && (
            <div className="flex items-center gap-2">
              <button
                onClick={() => onReject && onReject(plan.plan_id)}
                className="px-3 py-1.5 text-xs font-medium text-red-700 bg-red-50 hover:bg-red-100 border border-red-200 rounded transition-colors"
              >
                Reject Plan
              </button>
              <button
                onClick={() => onApprove && onApprove(plan.plan_id)}
                className="px-3 py-1.5 text-xs font-medium text-white bg-slate-900 hover:bg-slate-800 rounded transition-colors shadow-xs"
              >
                Approve Plan
              </button>
            </div>
          )}
        </div>
      </div>

      {/* Rationale & State Summary */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4 text-xs">
        <div className="bg-slate-50 p-3 rounded border border-slate-100 space-y-1">
          <span className="font-semibold text-slate-600 uppercase tracking-wider block text-[10px]">Current Observed State</span>
          <p className="text-slate-700 leading-relaxed">{plan.current_state_summary || 'Observed operational telemetry consistent with active shipment parameters.'}</p>
        </div>
        <div className="bg-slate-50 p-3 rounded border border-slate-100 space-y-1">
          <span className="font-semibold text-slate-600 uppercase tracking-wider block text-[10px]">Governing Policy Decision</span>
          <p className="text-slate-700 leading-relaxed">
            <span className="font-medium text-slate-900">{plan.policy_decision}:</span> {plan.policy_reason || 'Plan verified within acceptable risk thresholds.'}
          </p>
        </div>
      </div>

      {/* Metrics Row */}
      <div className="grid grid-cols-2 sm:grid-cols-4 gap-3 py-1 border-y border-slate-100 text-xs">
        <div>
          <span className="text-slate-400 block text-[11px]">Confidence</span>
          <span className="font-semibold text-slate-800">{Math.round((plan.confidence_score || 0) * 100)}%</span>
        </div>
        <div>
          <span className="text-slate-400 block text-[11px]">Data Sufficiency</span>
          <span className="font-semibold text-slate-800">{plan.data_sufficiency ? 'Satisfied' : 'Incomplete'}</span>
        </div>
        <div>
          <span className="text-slate-400 block text-[11px]">Risk Level</span>
          <span className="font-semibold text-slate-800">{plan.risk_level || 'LOW'}</span>
        </div>
        <div>
          <span className="text-slate-400 block text-[11px]">Selected Strategy</span>
          <span className="font-semibold text-indigo-700">{plan.selected_candidate_id || 'Rank #1 Strategy'}</span>
        </div>
      </div>

      {/* Execution Steps */}
      {steps && steps.length > 0 && (
        <div className="space-y-3">
          <div className="flex items-center justify-between">
            <h4 className="text-xs font-semibold text-slate-800 uppercase tracking-wider">Ordered Execution Steps ({steps.length})</h4>
          </div>
          <div className="divide-y divide-slate-100 border border-slate-200 rounded-md overflow-hidden">
            {steps.map((step, idx) => {
              let deps = [];
              try {
                if (typeof step.dependencies === 'string') deps = JSON.parse(step.dependencies || '[]');
                else if (Array.isArray(step.dependencies)) deps = step.dependencies;
              } catch (e) {
                deps = [];
              }

              return (
                <div key={step.step_id || idx} className="p-3 bg-white flex flex-col sm:flex-row sm:items-center justify-between gap-3 text-xs hover:bg-slate-50/50 transition-colors">
                  <div className="space-y-1">
                    <div className="flex items-center gap-2 flex-wrap">
                      <span className="font-mono text-slate-400 font-semibold">{step.step_number}.</span>
                      <span className="font-medium text-slate-900">{step.title}</span>
                      <span className={`text-[10px] px-1.5 py-0.2 rounded border font-mono ${getStatusBadge(step.status)}`}>
                        {step.status}
                      </span>
                      {deps.length > 0 && (
                        <span className="text-[10px] text-slate-500 bg-slate-100 px-1.5 py-0.5 rounded border border-slate-200">
                          Depends on: {deps.join(', ')}
                        </span>
                      )}
                      {step.requires_approval && (
                        <span className="text-[10px] text-amber-700 bg-amber-50 border border-amber-200 px-1 rounded">HITL Required</span>
                      )}
                    </div>
                    <p className="text-slate-500 pl-4">{step.description}</p>
                    <div className="pl-4 text-[11px] text-slate-400 font-mono flex flex-wrap items-center gap-3">
                      <span>Action: {step.action_type}</span>
                      <span>Outcome: {step.expected_outcome}</span>
                    </div>
                  </div>

                  {/* Step manual dispatch trigger when plan is approved */}
                  {plan.status === 'APPROVED' && step.status === 'PENDING' && onExecuteStep && (
                    <button
                      onClick={() => onExecuteStep(plan.plan_id, step.step_id)}
                      className="shrink-0 px-2.5 py-1 text-xs font-medium text-slate-700 hover:text-slate-900 bg-slate-100 hover:bg-slate-200 rounded transition-colors"
                    >
                      Execute Step
                    </button>
                  )}
                </div>
              );
            })}
          </div>
        </div>
      )}

      {/* Plan Comparison Modal */}
      {showComparison && (
        <PlanComparisonModal
          isOpen={showComparison}
          onClose={() => setShowComparison(false)}
          plan={plan}
          onCandidateSelected={() => {
            if (onPlanUpdated) {
              onPlanUpdated();
            }
          }}
        />
      )}
    </div>
  );
}

export default PlanLifecycleCard;

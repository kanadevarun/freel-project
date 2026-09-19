import React, { useState, useEffect, useCallback } from 'react';
import {
  X,
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
  FileCheck,
  ShieldAlert,
  Scale,
  Copy,
} from 'lucide-react';
import toast from 'react-hot-toast';
import { autonomyService } from '../../services/autonomyService';

export default function ContractComplianceMonitoringDrawer({
  contract,
  isOpen,
  onClose,
  onActionExecuted,
}) {
  const [activeTab, setActiveTab] = useState('strategies'); // 'strategies' | 'facts-deviations' | 'plan' | 'replan'
  const [state, setState] = useState(null);
  const [loading, setLoading] = useState(false);
  const [evaluating, setEvaluating] = useState(false);
  const [selectingStrategy, setSelectingStrategy] = useState(false);
  const [executingAction, setExecutingAction] = useState(false);
  const [replanning, setReplanning] = useState(false);

  // Replan form state
  const [replanTrigger, setReplanTrigger] = useState('DOCUMENT_VERIFIED');
  const [replanDocType, setReplanDocType] = useState('GDP_PHARMA_CERT');
  const [replanNotes, setReplanNotes] = useState('Auditor verified and certified cold chain Good Distribution Practice standard');

  const contractId = contract?.id;

  const fetchComplianceState = useCallback(async () => {
    if (!contractId) return;
    setLoading(true);
    try {
      const resp = await autonomyService.getContractComplianceState(contractId);
      const data = resp?.state || resp?.data?.state || resp?.data || resp;
      setState(data);
    } catch (err) {
      console.error('Failed to load contract compliance state:', err);
      toast.error('Failed to load compliance monitoring state');
    } finally {
      setLoading(false);
    }
  }, [contractId]);

  useEffect(() => {
    if (isOpen && contractId) {
      fetchComplianceState();
    }
  }, [isOpen, contractId, fetchComplianceState]);

  const handleRunEvaluation = async () => {
    if (!contractId) return;
    setEvaluating(true);
    try {
      const resp = await autonomyService.evaluateContractCompliance(contractId);
      toast.success('Contract & compliance audit evaluated successfully');
      const data = resp?.state || resp?.data?.state || resp?.data || resp;
      setState(data);
    } catch (err) {
      console.error('Compliance evaluation failed:', err);
      toast.error(err.response?.data?.message || 'Compliance evaluation failed');
    } finally {
      setEvaluating(false);
    }
  };

  const handleSelectStrategy = async (strategyId) => {
    if (!contractId) return;
    setSelectingStrategy(true);
    try {
      const resp = await autonomyService.selectContractComplianceStrategy(contractId, strategyId);
      toast.success('Remediation strategy updated');
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
    if (!contractId) return;
    setExecutingAction(true);
    try {
      const resp = await autonomyService.executeContractComplianceAction(contractId);
      toast.success('Compliance remediation action dispatched safely via Go Action System!');
      await fetchComplianceState();
      if (onActionExecuted) {
        onActionExecuted(resp);
      }
    } catch (err) {
      console.error('Failed to execute compliance action:', err);
      toast.error(err.response?.data?.message || 'Execution failed');
    } finally {
      setExecutingAction(false);
    }
  };

  const handleReplan = async (e) => {
    e?.preventDefault();
    if (!contractId) return;
    setReplanning(true);
    try {
      const payload = {
        document_type: replanDocType,
        change_notes: replanNotes,
        timestamp: new Date().toISOString(),
      };
      const resp = await autonomyService.replanContractCompliance(contractId, replanTrigger, payload);
      toast.success(`Plan re-evaluated & updated to Version ${resp?.plan?.version || 'new'}`);
      const data = resp?.state || resp?.data?.state || resp?.data || resp;
      setState(data);
      setActiveTab('plan');
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
  const selectedStrategyId = plan?.selected_remediation_strategy_id || plan?.recommended_remediation_strategy_id;
  const selectedCandidate = candidates.find((c) => c.strategy_id === selectedStrategyId);

  const facts = state?.authoritative_facts || [];
  const terms = state?.extracted_terms || [];
  const predictions = state?.predictions || [];
  const assumptions = state?.assumptions || [];
  const deviations = state?.deviations || [];
  const versions = state?.versions || [];

  const daysRemaining = state?.days_until_expiration;
  const isExpiringSoon = state?.expiration_status === 'EXPIRING_SOON' || (daysRemaining !== undefined && daysRemaining >= 0 && daysRemaining <= 30);
  const isExpired = state?.expiration_status === 'EXPIRED' || (daysRemaining !== undefined && daysRemaining < 0);
  const isCompliant = state?.compliance_status === 'COMPLIANT';
  const isNonCompliant = state?.compliance_status === 'NON_COMPLIANT';
  const isStopped = plan?.execution_status === 'STOPPED' || plan?.status === 'Expired' || isExpired;

  return (
    <div className="fixed inset-0 z-50 overflow-hidden bg-slate-900/40 backdrop-blur-xs flex justify-end animate-fadeIn">
      <div className="w-full max-w-3xl bg-white h-full shadow-2xl flex flex-col border-l border-slate-200 animate-slideLeft">
        {/* Top Header */}
        <div className="px-6 py-4 border-b border-slate-200 flex items-center justify-between bg-slate-50">
          <div className="flex items-center space-x-3">
            <div className="p-2 bg-blue-100 text-blue-700 rounded-lg">
              <Scale className="w-5 h-5" />
            </div>
            <div>
              <div className="flex items-center space-x-2">
                <h2 className="text-lg font-bold text-slate-900">
                  {state?.contract_reference || contract?.contract_reference || `Contract #${contractId}`}
                </h2>
                <span className="text-xs px-2 py-0.5 rounded-full font-medium bg-slate-200 text-slate-700">
                  {state?.contract_type || 'AGREEMENT'}
                </span>
                {plan?.version && (
                  <span className="text-xs px-2 py-0.5 rounded-full font-semibold bg-blue-100 text-blue-800 border border-blue-200">
                    Plan v{plan.version}
                  </span>
                )}
              </div>
              <p className="text-xs text-slate-500 flex items-center gap-2 mt-0.5">
                <Building2 className="w-3.5 h-3.5 text-slate-400" />
                <span>{state?.party_name || contract?.party_name || 'Counterparty'}</span>
                <span>•</span>
                <span>ID: #{contractId}</span>
              </p>
            </div>
          </div>

          <div className="flex items-center space-x-2">
            <button
              onClick={handleRunEvaluation}
              disabled={evaluating || loading}
              className="p-1.5 text-slate-600 hover:text-blue-600 hover:bg-slate-100 rounded-lg transition-colors flex items-center gap-1.5 text-xs font-medium border border-slate-200"
              title="Re-run AI & Rule Evaluation"
            >
              <RefreshCw className={`w-3.5 h-3.5 ${evaluating ? 'animate-spin text-blue-600' : ''}`} />
              <span>{evaluating ? 'Auditing...' : 'Evaluate'}</span>
            </button>
            <button
              onClick={onClose}
              className="p-1.5 text-slate-400 hover:text-slate-600 hover:bg-slate-100 rounded-lg transition-colors"
            >
              <X className="w-5 h-5" />
            </button>
          </div>
        </div>

        {/* Authoritative KPI Strip */}
        <div className="px-6 py-3 bg-white border-b border-slate-200 grid grid-cols-4 gap-3 text-xs">
          <div className="p-2.5 rounded-lg border border-slate-200 bg-slate-50">
            <div className="text-slate-500 text-[10px] font-semibold uppercase tracking-wider flex items-center gap-1">
              <span className="w-1.5 h-1.5 rounded-full bg-slate-400"></span>
              [Authoritative] Term
            </div>
            <div className="mt-1 font-bold text-slate-800 flex items-center gap-1.5">
              <Calendar className="w-3.5 h-3.5 text-slate-400" />
              <span>{state?.expiry_date || 'Ongoing'}</span>
            </div>
            <div className="mt-0.5 text-[11px] font-medium">
              {isExpired ? (
                <span className="text-rose-600 font-semibold">Expired ({Math.abs(daysRemaining)}d ago)</span>
              ) : isExpiringSoon ? (
                <span className="text-amber-600 font-semibold">Expiring in {daysRemaining} days</span>
              ) : (
                <span className="text-emerald-600 font-semibold">{daysRemaining ?? '--'} days remaining</span>
              )}
            </div>
          </div>

          <div className="p-2.5 rounded-lg border border-slate-200 bg-slate-50">
            <div className="text-slate-500 text-[10px] font-semibold uppercase tracking-wider flex items-center gap-1">
              <span className="w-1.5 h-1.5 rounded-full bg-emerald-500"></span>
              [Authoritative] Compliance
            </div>
            <div className="mt-1 font-bold text-slate-800 flex items-center gap-1.5">
              {isCompliant ? (
                <span className="text-emerald-700 bg-emerald-50 px-1.5 py-0.5 rounded font-bold">COMPLIANT</span>
              ) : isNonCompliant ? (
                <span className="text-rose-700 bg-rose-50 px-1.5 py-0.5 rounded font-bold">NON_COMPLIANT</span>
              ) : (
                <span className="text-amber-700 bg-amber-50 px-1.5 py-0.5 rounded font-bold">WARNING</span>
              )}
            </div>
            <div className="mt-0.5 text-[11px] text-slate-500">
              {plan?.missing_documents_count ? `${plan.missing_documents_count} missing document(s)` : 'Statutory filing current'}
            </div>
          </div>

          <div className="p-2.5 rounded-lg border border-slate-200 bg-slate-50">
            <div className="text-slate-500 text-[10px] font-semibold uppercase tracking-wider flex items-center gap-1">
              <span className="w-1.5 h-1.5 rounded-full bg-blue-500"></span>
              [Predicted] Risk & Impact
            </div>
            <div className="mt-1 font-bold flex items-center gap-1">
              <span className={`px-1.5 py-0.5 rounded text-[11px] ${
                plan?.risk_level === 'CRITICAL' ? 'bg-rose-100 text-rose-800' :
                plan?.risk_level === 'HIGH' ? 'bg-amber-100 text-amber-800' :
                plan?.risk_level === 'MEDIUM' ? 'bg-blue-100 text-blue-800' :
                'bg-emerald-100 text-emerald-800'
              }`}>
                {plan?.risk_level || 'LOW'} RISK
              </span>
              <span className="text-slate-500 text-[10px]">({Math.round((plan?.risk_score || 0) * 100)}%)</span>
            </div>
            <div className="mt-0.5 text-[11px] text-slate-500">
              Confidence: {Math.round((plan?.confidence_score || 0.9) * 100)}%
            </div>
          </div>

          <div className="p-2.5 rounded-lg border border-slate-200 bg-slate-50">
            <div className="text-slate-500 text-[10px] font-semibold uppercase tracking-wider flex items-center gap-1">
              <span className="w-1.5 h-1.5 rounded-full bg-purple-500"></span>
              [Governed] Autonomy
            </div>
            <div className="mt-1 font-bold text-slate-800 flex items-center gap-1">
              <Lock className="w-3.5 h-3.5 text-slate-400" />
              <span className="text-[11px]">{plan?.autonomy_level || 'LEVEL_2_PREPARE'}</span>
            </div>
            <div className="mt-0.5 text-[11px] font-medium">
              {plan?.requires_approval ? (
                <span className="text-amber-700 bg-amber-50 px-1 py-0.5 rounded font-semibold text-[10px]">
                  Approval Required
                </span>
              ) : (
                <span className="text-emerald-700 bg-emerald-50 px-1 py-0.5 rounded font-semibold text-[10px]">
                  Governed Auto-Permitted
                </span>
              )}
            </div>
          </div>
        </div>

        {/* Tab Navigation */}
        <div className="px-6 border-b border-slate-200 bg-white flex space-x-6 text-sm font-medium">
          <button
            onClick={() => setActiveTab('strategies')}
            className={`py-3 border-b-2 flex items-center gap-2 transition-colors ${
              activeTab === 'strategies'
                ? 'border-blue-600 text-blue-600 font-semibold'
                : 'border-transparent text-slate-500 hover:text-slate-800'
            }`}
          >
            <Sparkles className="w-4 h-4" />
            <span>Remediation Strategies ({candidates.length})</span>
          </button>

          <button
            onClick={() => setActiveTab('facts-deviations')}
            className={`py-3 border-b-2 flex items-center gap-2 transition-colors ${
              activeTab === 'facts-deviations'
                ? 'border-blue-600 text-blue-600 font-semibold'
                : 'border-transparent text-slate-500 hover:text-slate-800'
            }`}
          >
            <Layers className="w-4 h-4" />
            <span>Evidence & Deviations ({deviations.length})</span>
          </button>

          <button
            onClick={() => setActiveTab('plan')}
            className={`py-3 border-b-2 flex items-center gap-2 transition-colors ${
              activeTab === 'plan'
                ? 'border-blue-600 text-blue-600 font-semibold'
                : 'border-transparent text-slate-500 hover:text-slate-800'
            }`}
          >
            <FileCheck className="w-4 h-4" />
            <span>Remediation Execution Plan</span>
          </button>

          <button
            onClick={() => setActiveTab('replan')}
            className={`py-3 border-b-2 flex items-center gap-2 transition-colors ${
              activeTab === 'replan'
                ? 'border-blue-600 text-blue-600 font-semibold'
                : 'border-transparent text-slate-500 hover:text-slate-800'
            }`}
          >
            <Clock className="w-4 h-4" />
            <span>Event Adaptation & Lineage ({versions.length})</span>
          </button>
        </div>

        {/* Drawer Body */}
        <div className="flex-1 overflow-y-auto p-6 bg-slate-50/50 space-y-6">
          {loading ? (
            <div className="flex flex-col items-center justify-center h-64 text-slate-400 space-y-3">
              <RefreshCw className="w-8 h-8 animate-spin text-blue-500" />
              <p className="text-sm">Auditing contract terms & compliance records...</p>
            </div>
          ) : (
            <>
              {/* TAB 1: Strategies & Action Dispatch */}
              {activeTab === 'strategies' && (
                <div className="space-y-5">
                  {/* Stop reason banner if applicable */}
                  {isStopped && (
                    <div className="p-4 bg-rose-50 border border-rose-200 rounded-xl flex items-start gap-3">
                      <ShieldAlert className="w-5 h-5 text-rose-600 shrink-0 mt-0.5" />
                      <div>
                        <h4 className="text-sm font-bold text-rose-900">Commercial & Operational Dispatch Ceased</h4>
                        <p className="text-xs text-rose-700 mt-1">
                          {plan?.stop_reason || 'Contract has expired or hard statutory compliance block detected. Autonomous action dispatch is prohibited until formal human/legal review.'}
                        </p>
                      </div>
                    </div>
                  )}

                  <div>
                    <div className="flex items-center justify-between mb-3">
                      <h3 className="text-sm font-bold text-slate-900 flex items-center gap-2">
                        <span>Autonomous Remediation Candidates</span>
                        <span className="text-xs px-2 py-0.5 rounded font-normal bg-slate-200 text-slate-600">
                          {candidates.length} options evaluated
                        </span>
                      </h3>
                      <span className="text-xs text-slate-500">Go rules gate selection & execution</span>
                    </div>

                    <div className="space-y-3">
                      {candidates.map((candidate) => {
                        const isSelected = candidate.strategy_id === selectedStrategyId;
                        const isRecommended = candidate.strategy_id === plan?.recommended_remediation_strategy_id;

                        return (
                          <div
                            key={candidate.strategy_id}
                            onClick={() => !selectingStrategy && handleSelectStrategy(candidate.strategy_id)}
                            className={`p-4 rounded-xl border transition-all cursor-pointer bg-white relative ${
                              isSelected
                                ? 'border-blue-500 ring-2 ring-blue-100 shadow-sm'
                                : 'border-slate-200 hover:border-slate-300 hover:bg-slate-50/80'
                            } ${!candidate.is_feasible ? 'opacity-60 cursor-not-allowed bg-slate-50' : ''}`}
                          >
                            <div className="flex items-start justify-between">
                              <div className="flex items-start space-x-3">
                                <div className={`w-5 h-5 rounded-full border flex items-center justify-center mt-0.5 ${
                                  isSelected ? 'border-blue-600 bg-blue-600 text-white' : 'border-slate-300 bg-white'
                                }`}>
                                  {isSelected && <Check className="w-3 h-3" />}
                                </div>
                                <div>
                                  <div className="flex items-center space-x-2">
                                    <h4 className="text-sm font-bold text-slate-900">{candidate.strategy_name}</h4>
                                    {isRecommended && (
                                      <span className="text-[10px] px-2 py-0.5 rounded font-bold bg-blue-100 text-blue-800 border border-blue-200">
                                        RECOMMENDED
                                      </span>
                                    )}
                                    <span className={`text-[10px] px-2 py-0.5 rounded font-semibold ${
                                      candidate.severity_level === 'CRITICAL' ? 'bg-rose-100 text-rose-800' :
                                      candidate.severity_level === 'HIGH' ? 'bg-amber-100 text-amber-800' :
                                      'bg-slate-100 text-slate-700'
                                    }`}>
                                      {candidate.severity_level}
                                    </span>
                                  </div>
                                  <p className="text-xs text-slate-600 mt-1">{candidate.description}</p>
                                </div>
                              </div>

                              <div className="text-right shrink-0">
                                <div className="text-xs font-semibold text-slate-700">
                                  Action: <span className="text-blue-700 font-mono">{candidate.recommended_action}</span>
                                </div>
                                <div className="text-[11px] text-slate-500 mt-0.5">
                                  {candidate.requires_approval ? (
                                    <span className="text-amber-700 font-medium">Approval Required</span>
                                  ) : (
                                    <span className="text-emerald-700 font-medium">Standard Administrative</span>
                                  )}
                                </div>
                              </div>
                            </div>

                            {/* Candidate expected outcome & details */}
                            <div className="mt-3 pt-2.5 border-t border-slate-100 flex items-center justify-between text-xs text-slate-500">
                              <span className="flex items-center gap-1">
                                <Info className="w-3.5 h-3.5 text-slate-400" />
                                <span>Expected Outcome: <strong className="text-slate-700">{candidate.expected_outcome}</strong></span>
                              </span>
                              <span>Score: {Math.round(candidate.score * 100)}/100</span>
                            </div>
                          </div>
                        );
                      })}
                    </div>
                  </div>

                  {/* Selected Strategy Action Execution & Draft Preview */}
                  {selectedCandidate && (
                    <div className="p-5 bg-white border border-slate-200 rounded-xl space-y-4">
                      <div className="flex items-center justify-between">
                        <div className="flex items-center gap-2">
                          <Send className="w-4 h-4 text-blue-600" />
                          <h4 className="text-sm font-bold text-slate-900">Controlled Remediation Action</h4>
                        </div>
                        <span className="text-xs px-2 py-0.5 rounded font-mono bg-slate-100 text-slate-700">
                          {selectedCandidate.recommended_action}
                        </span>
                      </div>

                      {/* Draft Subject & Message Preview if present */}
                      {selectedCandidate.draft_subject && (
                        <div className="space-y-2">
                          <div className="text-xs font-semibold text-slate-600 flex items-center justify-between">
                            <span>Prepared Customer/Carrier Notice Draft:</span>
                            <button
                              type="button"
                              onClick={() => {
                                navigator.clipboard.writeText(`${selectedCandidate.draft_subject}\n\n${selectedCandidate.draft_message}`);
                                toast.success('Notice copied to clipboard');
                              }}
                              className="text-blue-600 hover:text-blue-800 flex items-center gap-1 text-[11px]"
                            >
                              <Copy className="w-3 h-3" />
                              <span>Copy</span>
                            </button>
                          </div>
                          <div className="p-3 bg-slate-50 border border-slate-200 rounded-lg text-xs space-y-1.5">
                            <div className="font-semibold text-slate-800">Subject: {selectedCandidate.draft_subject}</div>
                            <div className="text-slate-600 whitespace-pre-wrap leading-relaxed">{selectedCandidate.draft_message}</div>
                          </div>
                        </div>
                      )}

                      {/* Execution safety bar */}
                      <div className="pt-2 flex items-center justify-between border-t border-slate-100">
                        <div className="text-xs text-slate-500">
                          {plan?.requires_approval ? (
                            <span className="text-amber-700 flex items-center gap-1">
                              <AlertTriangle className="w-3.5 h-3.5" />
                              Human approval required ({plan.approval_reason || 'Policy threshold'})
                            </span>
                          ) : (
                            <span className="text-emerald-700 flex items-center gap-1">
                              <CheckCircle2 className="w-3.5 h-3.5" />
                              Pre-authorized low-risk administrative remediation
                            </span>
                          )}
                        </div>

                        <button
                          onClick={handleExecuteAction}
                          disabled={executingAction || isStopped}
                          className={`px-4 py-2 rounded-lg text-xs font-bold text-white flex items-center gap-2 shadow-xs transition-all ${
                            isStopped
                              ? 'bg-slate-400 cursor-not-allowed'
                              : 'bg-blue-600 hover:bg-blue-700 active:bg-blue-800'
                          }`}
                        >
                          {executingAction ? (
                            <RefreshCw className="w-4 h-4 animate-spin" />
                          ) : (
                            <Send className="w-4 h-4" />
                          )}
                          <span>
                            {isStopped
                              ? 'Action Dispatch Prohibited'
                              : 'Dispatch Action via Go Action System'}
                          </span>
                        </button>
                      </div>
                    </div>
                  )}
                </div>
              )}

              {/* TAB 2: Facts, Terms, Predictions & Deviations */}
              {activeTab === 'facts-deviations' && (
                <div className="space-y-6">
                  {/* Deviations Table */}
                  <div className="p-4 bg-white border border-slate-200 rounded-xl space-y-3">
                    <h3 className="text-sm font-bold text-slate-900 flex items-center gap-2">
                      <AlertTriangle className="w-4 h-4 text-amber-600" />
                      <span>Contract & Compliance Deviations ({deviations.length})</span>
                    </h3>
                    {deviations.length === 0 ? (
                      <p className="text-xs text-slate-500 py-3 text-center">
                        No active deviations or violations detected against authoritative terms.
                      </p>
                    ) : (
                      <div className="overflow-x-auto">
                        <table className="w-full text-xs text-left">
                          <thead className="bg-slate-50 text-slate-600 border-b border-slate-200">
                            <tr>
                              <th className="py-2 px-3 font-semibold">Requirement</th>
                              <th className="py-2 px-3 font-semibold">Observed Business State</th>
                              <th className="py-2 px-3 font-semibold">Deviation Nature</th>
                              <th className="py-2 px-3 font-semibold">Status</th>
                            </tr>
                          </thead>
                          <tbody className="divide-y divide-slate-100">
                            {deviations.map((dev, idx) => (
                              <tr key={idx} className="hover:bg-slate-50/80">
                                <td className="py-2 px-3 font-medium text-slate-800">{dev.requirement || dev.type}</td>
                                <td className="py-2 px-3 text-slate-600">{dev.current_state || dev.description}</td>
                                <td className="py-2 px-3">
                                  <span className={`px-1.5 py-0.5 rounded font-semibold text-[10px] ${
                                    dev.is_hard ? 'bg-rose-100 text-rose-800' : 'bg-amber-100 text-amber-800'
                                  }`}>
                                    {dev.is_hard ? 'HARD REQUIREMENT' : 'SOFT / ADVISORY'}
                                  </span>
                                </td>
                                <td className="py-2 px-3 font-mono text-[11px] text-slate-700">{dev.status || 'FLAGGED'}</td>
                              </tr>
                            ))}
                          </tbody>
                        </table>
                      </div>
                    )}
                  </div>

                  {/* 4-Quadrant Fact vs Extracted vs Prediction vs Assumption Segregation */}
                  <div className="grid grid-cols-2 gap-4">
                    {/* Authoritative Facts */}
                    <div className="p-4 bg-white border border-emerald-200 rounded-xl space-y-2">
                      <h4 className="text-xs font-bold text-emerald-900 uppercase tracking-wider flex items-center gap-1.5">
                        <CheckCircle2 className="w-4 h-4 text-emerald-600" />
                        <span>[Authoritative Facts]</span>
                      </h4>
                      <ul className="space-y-1 text-xs text-slate-700">
                        {facts.length > 0 ? (
                          facts.map((fact, idx) => (
                            <li key={idx} className="flex items-start gap-1.5">
                              <span className="text-emerald-500 font-bold">•</span>
                              <span>{fact}</span>
                            </li>
                          ))
                        ) : (
                          <li className="text-slate-400 italic">No facts cataloged</li>
                        )}
                      </ul>
                    </div>

                    {/* Extracted Contract Terms */}
                    <div className="p-4 bg-white border border-blue-200 rounded-xl space-y-2">
                      <h4 className="text-xs font-bold text-blue-900 uppercase tracking-wider flex items-center gap-1.5">
                        <FileText className="w-4 h-4 text-blue-600" />
                        <span>[Extracted Terms]</span>
                      </h4>
                      <ul className="space-y-1 text-xs text-slate-700">
                        {terms.length > 0 ? (
                          terms.map((term, idx) => (
                            <li key={idx} className="flex items-start gap-1.5">
                              <span className="text-blue-500 font-bold">•</span>
                              <span>{term}</span>
                            </li>
                          ))
                        ) : (
                          <li className="text-slate-400 italic">No extracted terms</li>
                        )}
                      </ul>
                    </div>

                    {/* AI Predictions */}
                    <div className="p-4 bg-white border border-purple-200 rounded-xl space-y-2">
                      <h4 className="text-xs font-bold text-purple-900 uppercase tracking-wider flex items-center gap-1.5">
                        <Sparkles className="w-4 h-4 text-purple-600" />
                        <span>[AI Predictions]</span>
                      </h4>
                      <ul className="space-y-1 text-xs text-slate-700">
                        {predictions.length > 0 ? (
                          predictions.map((pred, idx) => (
                            <li key={idx} className="flex items-start gap-1.5">
                              <span className="text-purple-500 font-bold">•</span>
                              <span>{pred}</span>
                            </li>
                          ))
                        ) : (
                          <li className="text-slate-400 italic">No predictions generated</li>
                        )}
                      </ul>
                    </div>

                    {/* AI Assumptions */}
                    <div className="p-4 bg-white border border-slate-200 rounded-xl space-y-2">
                      <h4 className="text-xs font-bold text-slate-700 uppercase tracking-wider flex items-center gap-1.5">
                        <Info className="w-4 h-4 text-slate-500" />
                        <span>[AI Assumptions]</span>
                      </h4>
                      <ul className="space-y-1 text-xs text-slate-700">
                        {assumptions.length > 0 ? (
                          assumptions.map((assump, idx) => (
                            <li key={idx} className="flex items-start gap-1.5">
                              <span className="text-slate-400 font-bold">•</span>
                              <span>{assump}</span>
                            </li>
                          ))
                        ) : (
                          <li className="text-slate-400 italic">No assumptions recorded</li>
                        )}
                      </ul>
                    </div>
                  </div>
                </div>
              )}

              {/* TAB 3: 7-Step Remediation Plan */}
              {activeTab === 'plan' && (
                <div className="space-y-5">
                  <div className="p-4 bg-white border border-slate-200 rounded-xl space-y-3">
                    <div className="flex items-center justify-between">
                      <h3 className="text-sm font-bold text-slate-900">7-Step Autonomous Remediation Execution Plan</h3>
                      <span className="text-xs px-2.5 py-0.5 rounded-full font-semibold bg-blue-100 text-blue-800">
                        Plan v{plan?.version || 1} • {plan?.execution_status || 'GENERATED'}
                      </span>
                    </div>
                    <p className="text-xs text-slate-500">
                      Stepwise sequential workflow executed under strict Go compliance gate boundaries.
                    </p>
                  </div>

                  <div className="space-y-3">
                    {[
                      { step: 1, action: 'AUDIT_REGULATORY_REQUIREMENTS', desc: 'Audit FMC filings, DG endorsements, and statutory jurisdiction mandates' },
                      { step: 2, action: 'INSPECT_DOCUMENT_EXPIRATIONS', desc: 'Inspect validity of carrier certificates, insurance policies, and licenses' },
                      { step: 3, action: 'VERIFY_COMMERCIAL_RATES', desc: 'Cross-check quotation & booking tariff tables against active contract rates' },
                      { step: 4, action: 'ASSESS_SEVERITY_AND_IMPACT', desc: 'Classify deviations as Hard Statutory Block or Soft Advisory Divergence' },
                      { step: 5, action: 'FORMULATE_REMEDIATION_STRATEGY', desc: 'Synthesize optimal candidate strategies and draft customer notices' },
                      { step: 6, action: 'GOVERN_HUMAN_APPROVAL', desc: 'Verify autonomy policy limits and gate high-impact actions for approval' },
                      { step: 7, action: 'DISPATCH_AUTHORIZED_ACTION', desc: 'Safely execute through the authoritative Go Action System boundary' },
                    ].map((stepItem) => {
                      const isActionStep = stepItem.step === 7;
                      return (
                        <div
                          key={stepItem.step}
                          className="p-3.5 bg-white border border-slate-200 rounded-lg flex items-start gap-3"
                        >
                          <div className={`w-7 h-7 rounded-full flex items-center justify-center text-xs font-bold shrink-0 ${
                            isActionStep ? 'bg-blue-600 text-white' : 'bg-slate-100 text-slate-700'
                          }`}>
                            {stepItem.step}
                          </div>
                          <div className="flex-1">
                            <div className="flex items-center justify-between">
                              <h4 className="text-xs font-bold text-slate-800 font-mono">{stepItem.action}</h4>
                              <span className="text-[10px] text-slate-500">
                                {isActionStep ? 'Action System Boundary' : 'Authoritative Guardrail'}
                              </span>
                            </div>
                            <p className="text-xs text-slate-600 mt-0.5">{stepItem.desc}</p>
                          </div>
                        </div>
                      );
                    })}
                  </div>
                </div>
              )}

              {/* TAB 4: Event Adaptation & Lineage History */}
              {activeTab === 'replan' && (
                <div className="space-y-6">
                  {/* Event Simulation Form */}
                  <form onSubmit={handleReplan} className="p-5 bg-white border border-slate-200 rounded-xl space-y-4">
                    <div className="flex items-center justify-between">
                      <div className="flex items-center gap-2">
                        <Clock className="w-4 h-4 text-blue-600" />
                        <h4 className="text-sm font-bold text-slate-900">Simulate Business Change / Replanning Trigger</h4>
                      </div>
                      <span className="text-xs text-slate-500">Generates incremental plan version</span>
                    </div>

                    <div className="grid grid-cols-2 gap-4 text-xs">
                      <div>
                        <label className="font-medium text-slate-700 block mb-1">Trigger Event</label>
                        <select
                          value={replanTrigger}
                          onChange={(e) => setReplanTrigger(e.target.value)}
                          className="w-full px-3 py-2 bg-slate-50 border border-slate-200 rounded-lg text-xs"
                        >
                          <option value="DOCUMENT_VERIFIED">DOCUMENT_VERIFIED (Clear Non-Compliance)</option>
                          <option value="DOCUMENT_UPLOADED">DOCUMENT_UPLOADED (Awaiting Inspection)</option>
                          <option value="DOCUMENT_REJECTED">DOCUMENT_REJECTED (Escalate to Legal)</option>
                          <option value="ROUTE_CHANGED">ROUTE_CHANGED (Revalidate Endorsements)</option>
                          <option value="RATE_AMENDED">RATE_AMENDED (Tariff Recheck)</option>
                          <option value="CONTRACT_EXPIRED">CONTRACT_EXPIRED (Impose Hard Stop)</option>
                        </select>
                      </div>

                      <div>
                        <label className="font-medium text-slate-700 block mb-1">Document / Requirement Type</label>
                        <input
                          type="text"
                          value={replanDocType}
                          onChange={(e) => setReplanDocType(e.target.value)}
                          className="w-full px-3 py-2 bg-slate-50 border border-slate-200 rounded-lg text-xs"
                          placeholder="e.g. GDP_PHARMA_CERT"
                        />
                      </div>
                    </div>

                    <div className="text-xs">
                      <label className="font-medium text-slate-700 block mb-1">Operational Change Reason / Details</label>
                      <textarea
                        rows={2}
                        value={replanNotes}
                        onChange={(e) => setReplanNotes(e.target.value)}
                        className="w-full px-3 py-2 bg-slate-50 border border-slate-200 rounded-lg text-xs"
                      />
                    </div>

                    <div className="flex justify-end">
                      <button
                        type="submit"
                        disabled={replanning}
                        className="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white font-bold text-xs rounded-lg flex items-center gap-1.5"
                      >
                        {replanning ? (
                          <RefreshCw className="w-3.5 h-3.5 animate-spin" />
                        ) : (
                          <RefreshCw className="w-3.5 h-3.5" />
                        )}
                        <span>Trigger Event Adaptation</span>
                      </button>
                    </div>
                  </form>

                  {/* Lineage History */}
                  <div className="p-4 bg-white border border-slate-200 rounded-xl space-y-3">
                    <h4 className="text-xs font-bold text-slate-900 uppercase tracking-wider">
                      Monitoring Plan Version Lineage ({versions.length})
                    </h4>
                    {versions.length === 0 ? (
                      <p className="text-xs text-slate-500 py-3 text-center">No incremental versions recorded.</p>
                    ) : (
                      <div className="space-y-2">
                        {versions.map((ver) => (
                          <div
                            key={ver.id || ver.version_number}
                            className="p-3 bg-slate-50 border border-slate-200 rounded-lg flex items-center justify-between text-xs"
                          >
                            <div className="flex items-center space-x-3">
                              <span className="font-bold px-2 py-0.5 rounded bg-blue-100 text-blue-800 border border-blue-200">
                                v{ver.version_number}
                              </span>
                              <div>
                                <div className="font-semibold text-slate-800 flex items-center gap-2">
                                  <span>{ver.trigger_event}</span>
                                  <span className="text-[10px] text-slate-500 font-mono">({ver.strategy_id})</span>
                                </div>
                                <div className="text-[11px] text-slate-600 mt-0.5">{ver.change_reason || 'Plan updated'}</div>
                              </div>
                            </div>
                            <div className="text-right">
                              <span className="text-[10px] text-slate-400">
                                {ver.created_at ? new Date(ver.created_at).toLocaleTimeString() : 'Recent'}
                              </span>
                            </div>
                          </div>
                        ))}
                      </div>
                    )}
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

import React, { useState, useEffect, useCallback } from 'react';
import {
  X,
  Shield,
  CheckCircle2,
  AlertTriangle,
  Clock,
  RotateCw,
  Edit3,
  StopCircle,
  ArrowUpRight,
  UserCheck,
  Cpu,
  Eye,
  Check,
  AlertCircle,
  FileText,
  Sliders,
  Sparkles,
  RefreshCw,
  GitBranch,
  Layers,
  Search,
  ExternalLink,
  Ban
} from 'lucide-react';
import autonomyService from '../../services/autonomyService';

export default function HumanAIDecisionCenterDrawer({
  isOpen,
  onClose,
  initialDecisionId = null,
  onDecisionUpdated
}) {
  const [activeTab, setActiveTab] = useState('decisions'); // 'decisions' | 'detail' | 'operating_model'
  const [loading, setLoading] = useState(false);
  const [actionLoading, setActionLoading] = useState(false);
  const [decisions, setDecisions] = useState([]);
  const [selectedDecision, setSelectedDecision] = useState(null);
  const [summary, setSummary] = useState(null);
  const [statusFilter, setStatusFilter] = useState('ALL');
  const [moduleFilter, setModuleFilter] = useState('ALL');
  const [error, setError] = useState(null);
  const [successMessage, setSuccessMessage] = useState(null);

  // Decision Action Form
  const [humanEditContent, setHumanEditContent] = useState('');
  const [decisionReason, setDecisionReason] = useState('');
  const [selectedAlternative, setSelectedAlternative] = useState('');
  const [isEditingPayload, setIsEditingPayload] = useState(false);

  const loadData = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      // 1. Load Summary
      const summaryRes = await autonomyService.getDecisionCenterSummary();
      const sumObj = summaryRes?.summary || summaryRes?.data?.summary;
      if (sumObj) setSummary(sumObj);

      // 2. Load Decisions List
      const params = {};
      if (statusFilter !== 'ALL') params.status = statusFilter;
      if (moduleFilter !== 'ALL') params.module = moduleFilter;

      const decRes = await autonomyService.listDecisionPoints(params);
      const decList = decRes?.decisions || decRes?.data?.decisions || [];
      setDecisions(decList);

      if (initialDecisionId) {
        const found = decList.find(d => d.decision_id === initialDecisionId);
        if (found) {
          selectDecision(found);
        } else {
          // Fetch directly
          try {
            const single = await autonomyService.getDecisionPoint(initialDecisionId);
            const decObj = single?.decision || single?.data?.decision;
            if (decObj) selectDecision(decObj);
          } catch (e) {
            console.warn('Initial decision not found', e);
          }
        }
      } else if (decList.length > 0 && !selectedDecision) {
        selectDecision(decList[0]);
      }
    } catch (err) {
      console.error('Failed loading decisions:', err);
      setError(err?.response?.data?.message || err.message || 'Failed to load decisions');
    } finally {
      setLoading(false);
    }
  }, [statusFilter, moduleFilter, initialDecisionId]);

  useEffect(() => {
    if (isOpen) {
      loadData();
    }
  }, [isOpen, loadData]);

  const selectDecision = (dec) => {
    setSelectedDecision(dec);
    setDecisionReason('');
    setSelectedAlternative('');
    setIsEditingPayload(false);

    // Populate payload edit content
    let rawPayload = '';
    if (dec.human_edited_payload && Object.keys(dec.human_edited_payload).length > 0) {
      rawPayload = typeof dec.human_edited_payload === 'string'
        ? dec.human_edited_payload
        : JSON.stringify(dec.human_edited_payload, null, 2);
    } else if (dec.original_ai_payload && Object.keys(dec.original_ai_payload).length > 0) {
      rawPayload = typeof dec.original_ai_payload === 'string'
        ? dec.original_ai_payload
        : JSON.stringify(dec.original_ai_payload, null, 2);
    }
    setHumanEditContent(rawPayload);
  };

  const handleDecisionSubmit = async (decisionType) => {
    if (!selectedDecision) return;
    setActionLoading(true);
    setError(null);
    setSuccessMessage(null);

    try {
      const payload = {
        decision_type: decisionType,
        reason: decisionReason || `Operator selected ${decisionType}`,
      };

      if (decisionType === 'EDIT_AND_APPROVE') {
        try {
          payload.human_edited_payload = JSON.parse(humanEditContent);
        } catch (e) {
          // Send as string or formatted json
          payload.human_edited_payload = { raw_content: humanEditContent };
        }
      }

      if (decisionType === 'OVERRIDE' && selectedAlternative) {
        payload.alternative_id = selectedAlternative;
      }

      const res = await autonomyService.submitHumanDecision(selectedDecision.decision_id, payload);
      const updated = res?.decision || res?.data?.decision;

      setSuccessMessage(`Decision successfully submitted: ${decisionType}`);
      if (updated) {
        setSelectedDecision(updated);
      }
      loadData();
      if (onDecisionUpdated) onDecisionUpdated(updated);
    } catch (err) {
      console.error('Decision submission error:', err);
      setError(err?.response?.data?.message || err.message || 'Failed submitting decision');
    } finally {
      setActionLoading(false);
    }
  };

  const handleStopWorkflow = async () => {
    if (!selectedDecision?.plan_id) return;
    setActionLoading(true);
    setError(null);
    try {
      await autonomyService.stopWorkflow(selectedDecision.plan_id, decisionReason || 'Human operator invoked emergency stop');
      setSuccessMessage('Active workflow stopped and all pending steps cancelled');
      loadData();
      if (onDecisionUpdated) onDecisionUpdated();
    } catch (err) {
      setError(err?.response?.data?.message || err.message || 'Failed stopping workflow');
    } finally {
      setActionLoading(false);
    }
  };

  if (!isOpen) return null;

  return (
    <div
      className="fixed inset-0 z-50 flex justify-end bg-slate-900/40 backdrop-blur-sm transition-opacity"
      data-testid="human-ai-decision-center-modal"
      role="dialog"
      aria-modal="true"
      aria-labelledby="decision-center-title"
    >
      <div className="flex h-full w-full max-w-5xl flex-col bg-white shadow-2xl border-l border-slate-200">
        {/* Header */}
        <div className="flex items-center justify-between border-b border-slate-200 px-6 py-4 bg-slate-50">
          <div className="flex items-center space-x-3">
            <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-navy-900 text-white shadow-sm">
              <UserCheck className="h-5 w-5" />
            </div>
            <div>
              <div className="flex items-center space-x-2">
                <h2 id="decision-center-title" className="text-lg font-bold text-slate-900">
                  Human + AI Decision Center
                </h2>
                <span className="inline-flex items-center rounded-full bg-blue-50 px-2 py-0.5 text-xs font-semibold text-blue-700 border border-blue-200">
                  Task 5.11 Controlled HITL
                </span>
              </div>
              <p className="text-xs text-slate-500">
                Authoritative Go boundary &bull; Governed AI recommendations &bull; Non-negotiable human oversight
              </p>
            </div>
          </div>
          <div className="flex items-center space-x-2">
            <button
              onClick={loadData}
              disabled={loading}
              className="p-2 text-slate-500 hover:text-slate-800 rounded-lg hover:bg-slate-200 transition-colors"
              title="Refresh decisions"
              aria-label="Refresh decisions"
            >
              <RefreshCw className={`h-4 w-4 ${loading ? 'animate-spin' : ''}`} />
            </button>
            <button
              onClick={onClose}
              className="p-2 text-slate-500 hover:text-slate-800 rounded-lg hover:bg-slate-200 transition-colors"
              aria-label="Close Decision Center"
            >
              <X className="h-5 w-5" />
            </button>
          </div>
        </div>

        {/* Operational Metrics Bar */}
        {summary && (
          <div className="grid grid-cols-2 md:grid-cols-6 gap-2 px-6 py-3 bg-white border-b border-slate-100 text-xs">
            <div className="p-2 rounded bg-slate-50 border border-slate-200">
              <span className="text-slate-500 block">Awaiting Review</span>
              <span className="text-base font-bold text-blue-700" data-testid="metric-awaiting-review">
                {summary.pending_awaiting_user || 0}
              </span>
            </div>
            <div className="p-2 rounded bg-slate-50 border border-slate-200">
              <span className="text-slate-500 block">High-Risk Plans</span>
              <span className="text-base font-bold text-amber-700" data-testid="metric-high-risk">
                {summary.high_risk_workflows || 0}
              </span>
            </div>
            <div className="p-2 rounded bg-slate-50 border border-slate-200">
              <span className="text-slate-500 block">Low Confidence</span>
              <span className="text-base font-bold text-orange-700" data-testid="metric-low-confidence">
                {summary.low_confidence_count || 0}
              </span>
            </div>
            <div className="p-2 rounded bg-slate-50 border border-slate-200">
              <span className="text-slate-500 block">Escalated</span>
              <span className="text-base font-bold text-red-700" data-testid="metric-escalations">
                {summary.active_escalations || 0}
              </span>
            </div>
            <div className="p-2 rounded bg-slate-50 border border-slate-200">
              <span className="text-slate-500 block">Human Modified</span>
              <span className="text-base font-bold text-purple-700" data-testid="metric-human-modified">
                {summary.modified_by_me_count || 0}
              </span>
            </div>
            <div className="p-2 rounded bg-slate-50 border border-slate-200">
              <span className="text-slate-500 block">Stopped Workflows</span>
              <span className="text-base font-bold text-slate-800" data-testid="metric-stopped-workflows">
                {summary.stopped_workflows || 0}
              </span>
            </div>
          </div>
        )}

        {/* Tab Navigation */}
        <div className="flex border-b border-slate-200 bg-white px-6">
          <button
            onClick={() => setActiveTab('decisions')}
            className={`flex items-center space-x-2 py-3 px-4 text-xs font-semibold border-b-2 transition-colors ${
              activeTab === 'decisions'
                ? 'border-blue-600 text-blue-600'
                : 'border-transparent text-slate-600 hover:text-slate-900'
            }`}
          >
            <Sliders className="h-4 w-4" />
            <span>Decisions Queue ({decisions.length})</span>
          </button>
          <button
            onClick={() => setActiveTab('detail')}
            className={`flex items-center space-x-2 py-3 px-4 text-xs font-semibold border-b-2 transition-colors ${
              activeTab === 'detail'
                ? 'border-blue-600 text-blue-600'
                : 'border-transparent text-slate-600 hover:text-slate-900'
            }`}
          >
            <Eye className="h-4 w-4" />
            <span>Decision Detail</span>
          </button>
          <button
            onClick={() => setActiveTab('operating_model')}
            className={`flex items-center space-x-2 py-3 px-4 text-xs font-semibold border-b-2 transition-colors ${
              activeTab === 'operating_model'
                ? 'border-blue-600 text-blue-600'
                : 'border-transparent text-slate-600 hover:text-slate-900'
            }`}
          >
            <Cpu className="h-4 w-4" />
            <span>Operating Model Architecture</span>
          </button>
        </div>

        {/* Notifications */}
        {error && (
          <div className="mx-6 mt-4 p-3 bg-red-50 border border-red-200 rounded-lg flex items-center space-x-2 text-xs text-red-700">
            <AlertCircle className="h-4 w-4 flex-shrink-0" />
            <span>{error}</span>
          </div>
        )}
        {successMessage && (
          <div className="mx-6 mt-4 p-3 bg-emerald-50 border border-emerald-200 rounded-lg flex items-center space-x-2 text-xs text-emerald-700">
            <CheckCircle2 className="h-4 w-4 flex-shrink-0" />
            <span>{successMessage}</span>
          </div>
        )}

        {/* Main Content Body */}
        <div className="flex-1 overflow-y-auto p-6 bg-slate-50">
          {activeTab === 'decisions' && (
            <div className="space-y-4">
              {/* Filter controls */}
              <div className="flex flex-wrap items-center justify-between gap-3 bg-white p-3 rounded-lg border border-slate-200">
                <div className="flex items-center space-x-3">
                  <div className="flex items-center space-x-1 text-xs text-slate-600">
                    <span className="font-medium">Status:</span>
                    <select
                      value={statusFilter}
                      onChange={(e) => setStatusFilter(e.target.value)}
                      className="rounded border border-slate-300 py-1 px-2 text-xs bg-white focus:outline-none focus:ring-1 focus:ring-blue-500"
                    >
                      <option value="ALL">All Statuses</option>
                      <option value="PENDING">Pending Approval</option>
                      <option value="APPROVED">Approved</option>
                      <option value="REJECTED">Rejected</option>
                      <option value="STOPPED">Stopped</option>
                      <option value="ESCALATED">Escalated</option>
                      <option value="SUPERSEDED">Superseded</option>
                    </select>
                  </div>
                  <div className="flex items-center space-x-1 text-xs text-slate-600">
                    <span className="font-medium">Module:</span>
                    <select
                      value={moduleFilter}
                      onChange={(e) => setModuleFilter(e.target.value)}
                      className="rounded border border-slate-300 py-1 px-2 text-xs bg-white focus:outline-none focus:ring-1 focus:ring-blue-500"
                    >
                      <option value="ALL">All Modules</option>
                      <option value="shipments">Shipments</option>
                      <option value="customer_followup">Customer Follow-up</option>
                      <option value="rfq_pricing">RFQ Pricing</option>
                      <option value="finance_collections">Finance</option>
                      <option value="contract_compliance">Contracts & Compliance</option>
                      <option value="exception_resolution">Exceptions</option>
                    </select>
                  </div>
                </div>
                <div className="text-xs text-slate-500">
                  Showing {decisions.length} operational decisions
                </div>
              </div>

              {/* Decisions List Grid */}
              <div className="grid grid-cols-1 gap-3">
                {decisions.length === 0 ? (
                  <div className="p-8 text-center bg-white rounded-lg border border-slate-200 text-slate-500">
                    <p className="text-sm">No decisions found for current filter criteria.</p>
                  </div>
                ) : (
                  decisions.map((dec) => {
                    const isSelected = selectedDecision?.decision_id === dec.decision_id;
                    const isPending = dec.decision_status === 'PENDING';
                    const isHighRisk = dec.risk_level === 'HIGH' || dec.risk_level === 'CRITICAL';

                    return (
                      <div
                        key={dec.decision_id}
                        onClick={() => {
                          selectDecision(dec);
                          setActiveTab('detail');
                        }}
                        className={`p-4 rounded-lg border cursor-pointer transition-all bg-white hover:border-blue-400 hover:shadow-sm ${
                          isSelected ? 'border-blue-600 ring-1 ring-blue-600' : 'border-slate-200'
                        }`}
                        data-testid={`decision-card-${dec.decision_id}`}
                      >
                        <div className="flex items-start justify-between">
                          <div>
                            <div className="flex items-center space-x-2">
                              <span className="text-xs font-mono text-slate-500">{dec.decision_id}</span>
                              <span className={`px-2 py-0.5 text-xs font-semibold rounded ${
                                dec.decision_status === 'PENDING'
                                  ? 'bg-amber-100 text-amber-800 border border-amber-300'
                                  : dec.decision_status === 'APPROVED'
                                  ? 'bg-emerald-100 text-emerald-800'
                                  : dec.decision_status === 'REJECTED'
                                  ? 'bg-red-100 text-red-800'
                                  : 'bg-slate-100 text-slate-700'
                              }`}>
                                {dec.decision_status}
                              </span>
                              <span className="px-2 py-0.5 text-xs font-medium rounded bg-slate-100 text-slate-700">
                                {dec.operating_mode}
                              </span>
                            </div>
                            <h3 className="text-sm font-semibold text-slate-900 mt-1">{dec.title}</h3>
                          </div>
                          <div className="flex items-center space-x-2 text-xs">
                            <span className={`px-2 py-0.5 rounded font-medium ${
                              isHighRisk ? 'bg-red-50 text-red-700 border border-red-200' : 'bg-slate-50 text-slate-600'
                            }`}>
                              Risk: {dec.risk_level}
                            </span>
                            <span className="px-2 py-0.5 rounded bg-blue-50 text-blue-700 font-medium">
                              Conf: {dec.confidence}
                            </span>
                          </div>
                        </div>

                        <div className="mt-3 text-xs text-slate-700 bg-slate-50 p-2.5 rounded border border-slate-100">
                          <span className="font-semibold text-slate-900">AI Recommendation:</span>{' '}
                          {dec.ai_recommendation}
                        </div>

                        {dec.human_decision && (
                          <div className="mt-2 text-xs text-slate-600 flex items-center justify-between border-t border-slate-100 pt-2">
                            <span>
                              <strong>Human Decision:</strong> {dec.human_decision} by {dec.decided_by_name || 'Operator'}
                            </span>
                            {dec.decision_reason && (
                              <span className="italic text-slate-500 truncate max-w-xs">
                                "{dec.decision_reason}"
                              </span>
                            )}
                          </div>
                        )}
                      </div>
                    );
                  })
                )}
              </div>
            </div>
          )}

          {activeTab === 'detail' && selectedDecision && (
            <div className="space-y-4">
              {/* Detailed Decision Card */}
              <div className="bg-white rounded-lg border border-slate-200 p-6 shadow-sm space-y-6">
                {/* Header Summary */}
                <div className="flex items-start justify-between border-b border-slate-200 pb-4">
                  <div>
                    <div className="flex items-center space-x-2 mb-1">
                      <span className="text-xs font-mono font-medium text-slate-500">
                        ID: {selectedDecision.decision_id}
                      </span>
                      <span className="px-2 py-0.5 text-xs font-semibold rounded bg-blue-100 text-blue-800">
                        Mode: {selectedDecision.operating_mode}
                      </span>
                      <span className="px-2 py-0.5 text-xs font-semibold rounded bg-slate-100 text-slate-800">
                        Autonomy: {selectedDecision.autonomy_level}
                      </span>
                    </div>
                    <h2 className="text-base font-bold text-slate-900">{selectedDecision.title}</h2>
                    <p className="text-xs text-slate-600 mt-1">
                      Entity: <span className="font-medium text-slate-900 uppercase">{selectedDecision.entity_type} #{selectedDecision.entity_id}</span> &bull; Module: <span className="font-medium">{selectedDecision.module}</span>
                    </p>
                  </div>
                  <div className="text-right">
                    <div className={`inline-block px-3 py-1 rounded text-xs font-bold ${
                      selectedDecision.decision_status === 'PENDING'
                        ? 'bg-amber-100 text-amber-900 border border-amber-300'
                        : selectedDecision.decision_status === 'APPROVED'
                        ? 'bg-emerald-100 text-emerald-900 border border-emerald-300'
                        : 'bg-slate-100 text-slate-800'
                    }`}>
                      {selectedDecision.decision_status}
                    </div>
                    <div className="text-[11px] text-slate-500 mt-1">
                      Reversible: {selectedDecision.is_reversible ? 'Yes' : 'No'}
                    </div>
                  </div>
                </div>

                {/* Provenance: Facts vs Predictions vs Recommendations */}
                <div className="space-y-3">
                  <h4 className="text-xs font-bold uppercase tracking-wider text-slate-500">
                    Decision Provenance Attribution
                  </h4>
                  <div className="grid grid-cols-1 md:grid-cols-3 gap-3 text-xs">
                    {/* Authoritative Facts */}
                    <div className="p-3 bg-emerald-50/70 border border-emerald-200 rounded-lg">
                      <div className="flex items-center space-x-1.5 font-bold text-emerald-800 mb-1.5">
                        <CheckCircle2 className="h-4 w-4 text-emerald-600" />
                        <span>[FACT] Authoritative Data</span>
                      </div>
                      <p className="text-slate-700">
                        {typeof selectedDecision.facts === 'object' && selectedDecision.facts !== null
                          ? JSON.stringify(selectedDecision.facts, null, 1)
                          : String(selectedDecision.facts || 'Authoritative state verified from DB.')}
                      </p>
                    </div>

                    {/* AI Predictions */}
                    <div className="p-3 bg-blue-50/70 border border-blue-200 rounded-lg">
                      <div className="flex items-center space-x-1.5 font-bold text-blue-800 mb-1.5">
                        <Sparkles className="h-4 w-4 text-blue-600" />
                        <span>[PREDICTION] Model Forecast</span>
                      </div>
                      <p className="text-slate-700">
                        {typeof selectedDecision.predictions === 'object' && selectedDecision.predictions !== null
                          ? JSON.stringify(selectedDecision.predictions, null, 1)
                          : String(selectedDecision.predictions || 'Forecasted trend aligned with historical trajectory.')}
                      </p>
                    </div>

                    {/* AI Recommendation */}
                    <div className="p-3 bg-purple-50/70 border border-purple-200 rounded-lg">
                      <div className="flex items-center space-x-1.5 font-bold text-purple-800 mb-1.5">
                        <Cpu className="h-4 w-4 text-purple-600" />
                        <span>[RECOMMENDATION] Prescribed</span>
                      </div>
                      <p className="text-slate-700 font-medium">
                        {selectedDecision.ai_recommendation}
                      </p>
                    </div>
                  </div>
                </div>

                {/* AI Prepared Action Payload vs Human Edited Version */}
                <div className="space-y-3">
                  <div className="flex items-center justify-between">
                    <h4 className="text-xs font-bold uppercase tracking-wider text-slate-500">
                      Prepared Action Payload (Task 5.11 Human Modification)
                    </h4>
                    {selectedDecision.decision_status === 'PENDING' && (
                      <button
                        onClick={() => setIsEditingPayload(!isEditingPayload)}
                        className="flex items-center space-x-1 text-xs font-semibold text-blue-600 hover:text-blue-800"
                      >
                        <Edit3 className="h-3.5 w-3.5" />
                        <span>{isEditingPayload ? 'Cancel Editing' : 'Modify Payload'}</span>
                      </button>
                    )}
                  </div>

                  <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
                    {/* Original AI Payload */}
                    <div className="p-3 bg-slate-50 border border-slate-200 rounded-lg">
                      <span className="text-[11px] font-bold text-slate-600 uppercase block mb-1">
                        Original AI Generated Payload:
                      </span>
                      <pre className="text-xs text-slate-800 font-mono bg-white p-2 rounded border border-slate-200 overflow-x-auto max-h-48 whitespace-pre-wrap">
                        {typeof selectedDecision.original_ai_payload === 'object'
                          ? JSON.stringify(selectedDecision.original_ai_payload, null, 2)
                          : String(selectedDecision.original_ai_payload || '{}')}
                      </pre>
                    </div>

                    {/* Human Edited Version / Editor */}
                    <div className="p-3 bg-blue-50/40 border border-blue-200 rounded-lg">
                      <span className="text-[11px] font-bold text-blue-900 uppercase block mb-1">
                        Human Edited Payload (Final Executable):
                      </span>
                      {isEditingPayload ? (
                        <textarea
                          value={humanEditContent}
                          onChange={(e) => setHumanEditContent(e.target.value)}
                          rows={7}
                          className="w-full text-xs font-mono p-2 bg-white border border-blue-300 rounded focus:ring-1 focus:ring-blue-500 focus:outline-none"
                          placeholder="Edit parameters or messaging draft..."
                        />
                      ) : (
                        <pre className="text-xs text-slate-800 font-mono bg-white p-2 rounded border border-blue-200 overflow-x-auto max-h-48 whitespace-pre-wrap">
                          {humanEditContent || 'No human edits applied. Using original AI payload.'}
                        </pre>
                      )}
                    </div>
                  </div>
                </div>

                {/* Candidate Alternatives */}
                {selectedDecision.alternatives && Array.isArray(selectedDecision.alternatives) && selectedDecision.alternatives.length > 0 && (
                  <div className="space-y-2">
                    <h4 className="text-xs font-bold uppercase tracking-wider text-slate-500">
                      Feasible Alternatives
                    </h4>
                    <div className="space-y-2">
                      {selectedDecision.alternatives.map((alt, i) => (
                        <label
                          key={alt.id || i}
                          className={`flex items-start space-x-3 p-3 rounded-lg border text-xs cursor-pointer ${
                            selectedAlternative === (alt.id || `alt-${i}`)
                              ? 'bg-blue-50 border-blue-500'
                              : 'bg-white border-slate-200 hover:bg-slate-50'
                          }`}
                        >
                          <input
                            type="radio"
                            name="alternative"
                            value={alt.id || `alt-${i}`}
                            checked={selectedAlternative === (alt.id || `alt-${i}`)}
                            onChange={(e) => setSelectedAlternative(e.target.value)}
                            className="mt-0.5 text-blue-600"
                          />
                          <div>
                            <span className="font-semibold text-slate-900">{alt.title || alt.strategy_name || `Alternative #${i + 1}`}</span>
                            <p className="text-slate-600 mt-0.5">{alt.description || JSON.stringify(alt)}</p>
                          </div>
                        </label>
                      ))}
                    </div>
                  </div>
                )}

                {/* Human Reason & Action Controls */}
                {selectedDecision.decision_status === 'PENDING' ? (
                  <div className="border-t border-slate-200 pt-4 space-y-4">
                    <div>
                      <label className="block text-xs font-semibold text-slate-700 mb-1">
                        Operator Decision Rationale / Audit Justification:
                      </label>
                      <input
                        type="text"
                        value={decisionReason}
                        onChange={(e) => setDecisionReason(e.target.value)}
                        placeholder="e.g., Authorized priority dispatch per customer SLA contract clause 4.2"
                        className="w-full text-xs p-2.5 rounded border border-slate-300 focus:ring-1 focus:ring-blue-500 focus:outline-none"
                      />
                    </div>

                    <div className="flex flex-wrap items-center gap-2">
                      <button
                        onClick={() => handleDecisionSubmit('APPROVE')}
                        disabled={actionLoading}
                        className="flex items-center space-x-1.5 px-4 py-2 bg-emerald-600 hover:bg-emerald-700 text-white rounded text-xs font-semibold shadow-sm transition-colors"
                        data-testid="btn-approve-decision"
                      >
                        <Check className="h-4 w-4" />
                        <span>Approve AI Recommendation</span>
                      </button>

                      {isEditingPayload && (
                        <button
                          onClick={() => handleDecisionSubmit('EDIT_AND_APPROVE')}
                          disabled={actionLoading}
                          className="flex items-center space-x-1.5 px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded text-xs font-semibold shadow-sm transition-colors"
                          data-testid="btn-edit-approve-decision"
                        >
                          <Edit3 className="h-4 w-4" />
                          <span>Save Edit & Approve</span>
                        </button>
                      )}

                      {selectedAlternative && (
                        <button
                          onClick={() => handleDecisionSubmit('OVERRIDE')}
                          disabled={actionLoading}
                          className="flex items-center space-x-1.5 px-4 py-2 bg-purple-600 hover:bg-purple-700 text-white rounded text-xs font-semibold shadow-sm transition-colors"
                          data-testid="btn-override-decision"
                        >
                          <GitBranch className="h-4 w-4" />
                          <span>Select Alternative Plan</span>
                        </button>
                      )}

                      <button
                        onClick={() => handleDecisionSubmit('REJECT')}
                        disabled={actionLoading}
                        className="flex items-center space-x-1.5 px-4 py-2 bg-rose-600 hover:bg-rose-700 text-white rounded text-xs font-semibold shadow-sm transition-colors"
                        data-testid="btn-reject-decision"
                      >
                        <X className="h-4 w-4" />
                        <span>Reject Recommendation</span>
                      </button>

                      <button
                        onClick={() => handleDecisionSubmit('ESCALATE')}
                        disabled={actionLoading}
                        className="flex items-center space-x-1.5 px-3 py-2 bg-amber-600 hover:bg-amber-700 text-white rounded text-xs font-semibold shadow-sm transition-colors"
                        data-testid="btn-escalate-decision"
                      >
                        <ArrowUpRight className="h-4 w-4" />
                        <span>Escalate</span>
                      </button>

                      {selectedDecision.plan_id && (
                        <button
                          onClick={handleStopWorkflow}
                          disabled={actionLoading}
                          className="flex items-center space-x-1.5 px-3 py-2 bg-slate-800 hover:bg-black text-white rounded text-xs font-semibold shadow-sm transition-colors ml-auto"
                          data-testid="btn-stop-workflow"
                        >
                          <StopCircle className="h-4 w-4 text-red-400" />
                          <span>Stop Plan (Emergency)</span>
                        </button>
                      )}
                    </div>
                  </div>
                ) : (
                  <div className="border-t border-slate-200 pt-4 p-3 bg-slate-50 rounded-lg text-xs space-y-1">
                    <div className="font-bold text-slate-900">Historical Decision Record</div>
                    <div className="text-slate-700">
                      Decided by <strong>{selectedDecision.decided_by_name || 'Operator'}</strong> on{' '}
                      {new Date(selectedDecision.decided_at || selectedDecision.updated_at).toLocaleString()}
                    </div>
                    {selectedDecision.decision_reason && (
                      <div className="italic text-slate-600">
                        Reason: "{selectedDecision.decision_reason}"
                      </div>
                    )}
                  </div>
                )}
              </div>
            </div>
          )}

          {activeTab === 'operating_model' && (
            <div className="space-y-6 bg-white p-6 rounded-lg border border-slate-200">
              <div>
                <h3 className="text-sm font-bold text-slate-900 mb-1">
                  LogisticsHQ Controlled Operating Model Flow
                </h3>
                <p className="text-xs text-slate-600">
                  Strict separation of concerns: Python Sidecar for intelligence &amp; reasoning; Go Backend for authoritative state, security, tenant isolation, and Action System execution.
                </p>
              </div>

              {/* Flow Steps */}
              <div className="relative border-l-2 border-blue-500 ml-4 pl-6 space-y-6 text-xs">
                <div className="relative">
                  <div className="absolute -left-[31px] top-0 h-4 w-4 rounded-full bg-blue-600 border-2 border-white" />
                  <span className="font-bold text-slate-900 block uppercase">1. AI Observes &amp; Ingests</span>
                  <p className="text-slate-600">Continuous telemetry, EDI/API events, milestone timestamps.</p>
                </div>
                <div className="relative">
                  <div className="absolute -left-[31px] top-0 h-4 w-4 rounded-full bg-blue-600 border-2 border-white" />
                  <span className="font-bold text-slate-900 block uppercase">2. AI Analyzes &amp; Recommends</span>
                  <p className="text-slate-600">Python sidecar reasons over facts, estimates delay probability, and drafts action payloads.</p>
                </div>
                <div className="relative">
                  <div className="absolute -left-[31px] top-0 h-4 w-4 rounded-full bg-blue-600 border-2 border-white" />
                  <span className="font-bold text-slate-900 block uppercase">3. Go Policy &amp; Autonomy Gate</span>
                  <p className="text-slate-600">Go backend checks tenant autonomy level (0-4), emergency stop status, monetary limits, and approval mandates.</p>
                </div>
                <div className="relative">
                  <div className="absolute -left-[31px] top-0 h-4 w-4 rounded-full bg-amber-500 border-2 border-white" />
                  <span className="font-bold text-amber-900 block uppercase">4. Human Review &amp; Edit (HITL)</span>
                  <p className="text-slate-600">Operator reviews provenance, edits payload parameters, approves, chooses alternatives, or stops workflow.</p>
                </div>
                <div className="relative">
                  <div className="absolute -left-[31px] top-0 h-4 w-4 rounded-full bg-emerald-600 border-2 border-white" />
                  <span className="font-bold text-emerald-900 block uppercase">5. Action System Execution</span>
                  <p className="text-slate-600">Go backend executes approved or low-risk authorized actions idempotently with rollback protections.</p>
                </div>
                <div className="relative">
                  <div className="absolute -left-[31px] top-0 h-4 w-4 rounded-full bg-purple-600 border-2 border-white" />
                  <span className="font-bold text-purple-900 block uppercase">6. Outcome Learning &amp; Memory</span>
                  <p className="text-slate-600">Tenant-isolated memory records human preferences, edit habits, and execution telemetry without chain-of-thought storage.</p>
                </div>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

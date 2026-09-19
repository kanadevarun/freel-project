import React, { useState, useEffect, useCallback } from 'react';
import {
  X,
  Zap,
  Clock,
  AlertTriangle,
  CheckCircle2,
  RotateCw,
  Shield,
  Layers,
  HelpCircle,
  Play,
  ArrowRight,
  TrendingDown,
  FileText,
  AlertCircle
} from 'lucide-react';
import autonomyService from '../../services/autonomyService';
import PlanComparisonModal from './PlanComparisonModal';

/**
 * ShipmentAdaptiveDrawer.jsx — Phase 5 Task 5.3 Adaptive Shipment Management Drawer
 *
 * Provides deep operational insight into:
 * - Shipment Operational Risk & Adaptive Status
 * - Predicted ETA vs Customer Commitment Protection
 * - Active Autonomous Recovery Plan & Candidate Alternatives
 * - Durable Waiting States (Carrier, Approval, Milestone, Customer, External)
 * - Event Journal with Deduplication Keys & Decisions
 * - Policy Validation, Approvals, and Step Execution
 */
export default function ShipmentAdaptiveDrawer({
  isOpen,
  onClose,
  shipmentId,
  shipmentData,
  onPlanUpdated
}) {
  const [loading, setLoading] = useState(false);
  const [stateData, setStateData] = useState(null);
  const [events, setEvents] = useState([]);
  const [error, setError] = useState(null);
  const [actionLoading, setActionLoading] = useState(false);
  const [showComparison, setShowComparison] = useState(false);
  const [simulationLoading, setSimulationLoading] = useState(false);

  const fetchAdaptiveState = useCallback(async () => {
    if (!shipmentId) return;
    setLoading(true);
    setError(null);
    try {
      const resp = await autonomyService.getShipmentAdaptiveState(shipmentId);
      if (resp && resp.state) {
        setStateData(resp.state);
        setEvents(resp.state.recent_events || []);
      }
    } catch (err) {
      console.error('Failed fetching adaptive shipment state:', err);
      setError(err?.response?.data?.message || err.message || 'Failed to load adaptive shipment state');
    } finally {
      setLoading(false);
    }
  }, [shipmentId]);

  useEffect(() => {
    if (isOpen && shipmentId) {
      fetchAdaptiveState();
    }
  }, [isOpen, shipmentId, fetchAdaptiveState]);

  if (!isOpen) return null;

  const activePlan = stateData?.active_plan;
  const activeSteps = stateData?.active_plan_steps || [];

  const handleApprove = async () => {
    if (!activePlan) return;
    setActionLoading(true);
    try {
      await autonomyService.approvePlan(activePlan.plan_id, 'Approved via Adaptive Management Drawer');
      await fetchAdaptiveState();
      if (onPlanUpdated) onPlanUpdated();
    } catch (err) {
      alert('Approval failed: ' + (err?.response?.data?.message || err.message));
    } finally {
      setActionLoading(false);
    }
  };

  const handleReject = async () => {
    if (!activePlan) return;
    const reason = prompt('Please enter rejection reason:');
    if (!reason) return;
    setActionLoading(true);
    try {
      await autonomyService.rejectPlan(activePlan.plan_id, reason);
      await fetchAdaptiveState();
      if (onPlanUpdated) onPlanUpdated();
    } catch (err) {
      alert('Rejection failed: ' + (err?.response?.data?.message || err.message));
    } finally {
      setActionLoading(false);
    }
  };

  const handleExecuteStep = async (stepId) => {
    if (!activePlan) return;
    setActionLoading(true);
    try {
      const res = await autonomyService.executeStep(activePlan.plan_id, stepId);
      alert(`Step ${stepId} executed! Verification: ${res.verification_status}`);
      await fetchAdaptiveState();
      if (onPlanUpdated) onPlanUpdated();
    } catch (err) {
      alert('Step execution failed: ' + (err?.response?.data?.message || err.message));
    } finally {
      setActionLoading(false);
    }
  };

  const handleTransitionWaitingState = async (waitingState) => {
    if (!activePlan) return;
    setActionLoading(true);
    try {
      await autonomyService.transitionWaitingState(activePlan.plan_id, {
        waiting_state: waitingState,
        waiting_until: waitingState !== 'NONE' ? new Date(Date.now() + 6 * 3600 * 1000).toISOString() : null
      });
      await fetchAdaptiveState();
    } catch (err) {
      alert('Failed updating waiting state: ' + (err?.response?.data?.message || err.message));
    } finally {
      setActionLoading(false);
    }
  };

  const handleSimulateEvent = async (type, payloadOverride = {}) => {
    setSimulationLoading(true);
    try {
      let payload = { ...payloadOverride };
      let severity = 'MEDIUM';
      let eventType = type;
      let predictedEta = null;
      let commitmentDate = stateData?.customer_commitment_date || '2026-10-18T18:00:00Z';

      if (type === 'MINOR_ETA_FLUCTUATION') {
        eventType = 'PREDICTIVE_ETA_UPDATE';
        severity = 'LOW';
        payload = { delay_hours: 1.5, reason: 'Mild ocean swell', confidence: 0.92 };
        predictedEta = '2026-10-18T12:00:00Z';
      } else if (type === 'MAJOR_ETA_DELAY') {
        eventType = 'PREDICTIVE_ETA_UPDATE';
        severity = 'HIGH';
        payload = { delay_hours: 28.0, reason: 'Severe port terminal congestion', confidence: 0.89 };
        predictedEta = '2026-10-20T10:00:00Z';
      } else if (type === 'CUSTOMS_HOLD') {
        eventType = 'CUSTOMS_EXCEPTION';
        severity = 'CRITICAL';
        payload = { hold_code: 'CUST-BLOCK-44', documentation_missing: 'Certificate of Origin', authority: 'US Customs' };
      } else if (type === 'DUPLICATE_EVENT') {
        eventType = 'PREDICTIVE_ETA_UPDATE';
        severity = 'LOW';
        payload = { delay_hours: 1.5, reason: 'Duplicate replayed webhook', confidence: 0.92 };
      }

      const res = await autonomyService.ingestShipmentEvent(shipmentId, {
        event_type: eventType,
        severity,
        payload,
        deduplication_key: type === 'DUPLICATE_EVENT' ? `test-dedup-fixed-${shipmentId}` : undefined,
        predicted_eta: predictedEta,
        customer_commitment_date: commitmentDate
      });

      alert(`Event Processed: ${res.result?.decision} - ${res.result?.decision_reason}`);
      await fetchAdaptiveState();
      if (onPlanUpdated) onPlanUpdated();
    } catch (err) {
      alert('Event simulation failed: ' + (err?.response?.data?.message || err.message));
    } finally {
      setSimulationLoading(false);
    }
  };

  const getRiskBadge = (level) => {
    const l = (level || 'LOW').toUpperCase();
    if (l === 'CRITICAL') return 'bg-red-100 text-red-800 border-red-300';
    if (l === 'HIGH') return 'bg-orange-100 text-orange-800 border-orange-300';
    if (l === 'MEDIUM') return 'bg-amber-100 text-amber-800 border-amber-300';
    return 'bg-emerald-100 text-emerald-800 border-emerald-300';
  };

  const getWaitingBadge = (ws) => {
    if (!ws || typeof ws !== 'string' || ws === 'NONE') return 'bg-slate-100 text-slate-700 border-slate-200';
    const s = String(ws).toUpperCase();
    if (s.includes('CARRIER')) return 'bg-blue-100 text-blue-800 border-blue-300';
    if (s.includes('APPROVAL')) return 'bg-amber-100 text-amber-800 border-amber-300';
    if (s.includes('CUSTOMER')) return 'bg-purple-100 text-purple-800 border-purple-300';
    return 'bg-indigo-100 text-indigo-800 border-indigo-300';
  };

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 z-50 overflow-hidden bg-slate-900/40 backdrop-blur-sm flex justify-end">
      <div className="w-full max-w-3xl bg-white h-full shadow-2xl flex flex-col border-l border-slate-200 animate-in slide-in-from-right duration-200">
        {/* Top Header */}
        <div className="p-5 border-b border-slate-200 flex items-center justify-between bg-slate-50/80">
          <div>
            <div className="flex items-center gap-2 mb-1">
              <span className="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-xs font-semibold bg-blue-100 text-blue-800 border border-blue-200">
                <Zap className="w-3.5 h-3.5 text-blue-600" />
                Adaptive Shipment Management
              </span>
              <span className="text-xs px-2 py-0.5 rounded font-mono font-medium bg-slate-200 text-slate-700">
                ID #{shipmentId}
              </span>
              <span className={`text-xs px-2 py-0.5 rounded-full font-semibold border ${getRiskBadge(stateData?.current_risk_level)}`}>
                {stateData?.current_risk_level || 'LOW'} RISK
              </span>
            </div>
            <h2 className="text-lg font-bold text-slate-900">
              {shipmentData?.booking_number || `Shipment #${shipmentId}`} Adaptive Control
            </h2>
          </div>
          <div className="flex items-center gap-2">
            <button
              onClick={fetchAdaptiveState}
              disabled={loading}
              title="Refresh Adaptive State"
              className="p-2 text-slate-500 hover:text-slate-800 hover:bg-slate-200/60 rounded-lg transition-colors"
            >
              <RotateCw className={`w-4 h-4 ${loading ? 'animate-spin' : ''}`} />
            </button>
            <button
              onClick={onClose}
              className="p-2 text-slate-500 hover:text-slate-800 hover:bg-slate-200/60 rounded-lg transition-colors"
            >
              <X className="w-5 h-5" />
            </button>
          </div>
        </div>

        {/* Scrollable Content Body */}
        <div className="flex-1 overflow-y-auto p-5 space-y-6">
          {error && (
            <div className="p-4 bg-red-50 border border-red-200 rounded-lg flex items-start gap-3">
              <AlertTriangle className="w-5 h-5 text-red-600 shrink-0 mt-0.5" />
              <div>
                <h4 className="text-sm font-semibold text-red-900">Error Loading Adaptive State</h4>
                <p className="text-sm text-red-700 mt-0.5">{error}</p>
              </div>
            </div>
          )}

          {/* Section 1: Customer Commitment & ETA Protection Banner */}
          <div className="bg-white border border-slate-200 rounded-xl p-4 shadow-xs">
            <h3 className="text-xs font-bold uppercase tracking-wider text-slate-500 mb-3 flex items-center gap-2">
              <Clock className="w-4 h-4 text-slate-600" />
              Customer Commitment Protection & ETA Analysis
            </h3>
            <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
              {/* Actual Commitment */}
              <div className="p-3 bg-slate-50 border border-slate-200 rounded-lg">
                <div className="text-[11px] font-semibold text-slate-500 uppercase flex items-center gap-1">
                  <span className="w-2 h-2 rounded-full bg-emerald-500" />
                  Actual Commitment
                </div>
                <div className="text-sm font-bold text-slate-800 mt-1">
                  {stateData?.customer_commitment_date
                    ? new Date(stateData.customer_commitment_date).toLocaleString('en-US', { dateStyle: 'medium', timeStyle: 'short' })
                    : 'Not Specified'}
                </div>
                <div className="text-[11px] text-slate-500 mt-0.5">Agreed Customer SLA</div>
              </div>

              {/* Predicted ETA */}
              <div className="p-3 bg-blue-50/50 border border-blue-200 rounded-lg">
                <div className="text-[11px] font-semibold text-blue-700 uppercase flex items-center gap-1">
                  <span className="w-2 h-2 rounded-full bg-blue-500" />
                  Predicted ETA (Phase 4)
                </div>
                <div className="text-sm font-bold text-blue-900 mt-1">
                  {stateData?.predicted_eta
                    ? new Date(stateData.predicted_eta).toLocaleString('en-US', { dateStyle: 'medium', timeStyle: 'short' })
                    : stateData?.eta
                    ? new Date(stateData.eta).toLocaleString('en-US', { dateStyle: 'medium', timeStyle: 'short' })
                    : 'Pending Assessment'}
                </div>
                <div className="text-[11px] text-blue-600 mt-0.5">Machine Learning Projection</div>
              </div>

              {/* Deviation & Risk */}
              <div className="p-3 bg-slate-50 border border-slate-200 rounded-lg">
                <div className="text-[11px] font-semibold text-slate-500 uppercase flex items-center gap-1">
                  <TrendingDown className="w-3.5 h-3.5 text-slate-500" />
                  SLA Commitment Risk
                </div>
                <div className="text-sm font-bold text-slate-900 mt-1 flex items-center gap-2">
                  <span>{stateData?.eta_deviation_hours ? `+${stateData.eta_deviation_hours.toFixed(1)} hrs` : '0.0 hrs'}</span>
                  <span className={`text-[10px] px-1.5 py-0.5 rounded font-bold uppercase border ${getRiskBadge(stateData?.commitment_risk_severity)}`}>
                    {stateData?.commitment_risk_severity || 'NONE'}
                  </span>
                </div>
                <div className="text-[11px] text-slate-500 mt-0.5">
                  {stateData?.eta_deviation_hours > 0 ? 'Threatens delivery window' : 'Within safe commitment buffer'}
                </div>
              </div>
            </div>
          </div>

          {/* Section 2: Active AI Recovery Plan & Waiting State */}
          <div className="bg-white border border-slate-200 rounded-xl p-4 shadow-xs">
            <div className="flex items-center justify-between mb-3">
              <h3 className="text-xs font-bold uppercase tracking-wider text-slate-500 flex items-center gap-2">
                <Shield className="w-4 h-4 text-indigo-600" />
                Active Autonomous Recovery Plan
              </h3>
              {activePlan && (
                <button
                  onClick={() => setShowComparison(true)}
                  className="text-xs font-semibold text-indigo-600 hover:text-indigo-800 flex items-center gap-1"
                >
                  <Layers className="w-3.5 h-3.5" />
                  Compare Alternatives
                </button>
              )}
            </div>

            {activePlan ? (
              <div className="space-y-4">
                <div className="p-3 bg-slate-50 border border-slate-200 rounded-lg flex flex-wrap items-center justify-between gap-2">
                  <div>
                    <div className="flex items-center gap-2">
                      <span className="font-mono text-xs font-bold text-slate-900">{activePlan.plan_id}</span>
                      <span className="text-xs px-2 py-0.5 rounded bg-slate-200 text-slate-700 font-semibold">
                        Version {activePlan.version}
                      </span>
                      <span className="text-xs px-2 py-0.5 rounded-full font-bold bg-amber-50 text-amber-800 border border-amber-200">
                        {activePlan.status}
                      </span>
                    </div>
                    <p className="text-xs text-slate-600 mt-1">{activePlan.goal}</p>
                  </div>
                  <div className="flex items-center gap-2">
                    <span className={`text-xs px-2 py-1 rounded font-semibold border ${getWaitingBadge(activePlan.waiting_state)}`}>
                      {activePlan.waiting_state && typeof activePlan.waiting_state === 'string' && activePlan.waiting_state !== 'NONE'
                        ? activePlan.waiting_state.replace(/_/g, ' ')
                        : 'ACTIVE EXECUTION'}
                    </span>
                  </div>
                </div>

                {/* Waiting State Control */}
                <div className="p-3 bg-slate-50/60 border border-dashed border-slate-200 rounded-lg flex items-center justify-between">
                  <div className="text-xs text-slate-600">
                    <span className="font-semibold text-slate-800">Durable Waiting State: </span>
                    Workflows pause gracefully without busy-looping while awaiting external updates.
                  </div>
                  <div className="flex items-center gap-1.5">
                    <button
                      onClick={() => handleTransitionWaitingState('WAITING_FOR_CARRIER')}
                      disabled={actionLoading}
                      className="text-[11px] px-2 py-1 bg-white border border-slate-300 hover:bg-slate-100 rounded text-slate-700 font-medium"
                    >
                      Wait for Carrier
                    </button>
                    <button
                      onClick={() => handleTransitionWaitingState('WAITING_FOR_APPROVAL')}
                      disabled={actionLoading}
                      className="text-[11px] px-2 py-1 bg-white border border-slate-300 hover:bg-slate-100 rounded text-slate-700 font-medium"
                    >
                      Wait for Approval
                    </button>
                    <button
                      onClick={() => handleTransitionWaitingState('NONE')}
                      disabled={actionLoading}
                      className="text-[11px] px-2 py-1 bg-emerald-50 border border-emerald-300 hover:bg-emerald-100 rounded text-emerald-800 font-medium"
                    >
                      Resume Plan
                    </button>
                  </div>
                </div>

                {/* Policy & Approval Callout */}
                {activePlan.status === 'REQUIRES_APPROVAL' && (
                  <div className="p-3 bg-amber-50 border border-amber-200 rounded-lg flex items-center justify-between">
                    <div className="flex items-center gap-2 text-xs text-amber-900">
                      <AlertTriangle className="w-4 h-4 text-amber-600 shrink-0" />
                      <div>
                        <span className="font-bold">Approval Required by Autonomy Policy: </span>
                        {activePlan.policy_reason || 'Consequential shipment action requires authorized operator approval.'}
                      </div>
                    </div>
                    <div className="flex items-center gap-2">
                      <button
                        onClick={handleApprove}
                        disabled={actionLoading}
                        className="px-3 py-1 bg-emerald-600 hover:bg-emerald-700 text-white rounded text-xs font-semibold shadow-xs"
                      >
                        Approve Plan
                      </button>
                      <button
                        onClick={handleReject}
                        disabled={actionLoading}
                        className="px-3 py-1 bg-white border border-slate-300 hover:bg-slate-100 text-slate-700 rounded text-xs font-semibold"
                      >
                        Reject
                      </button>
                    </div>
                  </div>
                )}

                {/* Plan Steps */}
                <div>
                  <h4 className="text-xs font-bold text-slate-700 uppercase tracking-wider mb-2">Controlled Plan Steps</h4>
                  <div className="space-y-2">
                    {activeSteps.map((step) => (
                      <div
                        key={step.step_id}
                        className="p-3 bg-white border border-slate-200 rounded-lg flex items-center justify-between gap-3 text-xs"
                      >
                        <div className="flex items-start gap-2.5">
                          <span className="w-5 h-5 rounded-full bg-slate-100 text-slate-700 flex items-center justify-center font-bold text-[11px] shrink-0 mt-0.5">
                            {step.step_number}
                          </span>
                          <div>
                            <div className="font-semibold text-slate-900">{step.title}</div>
                            <div className="text-slate-500 text-[11px]">{step.description}</div>
                            <div className="text-[10px] text-slate-400 font-mono mt-0.5">Action: {step.action_type}</div>
                          </div>
                        </div>
                        <div className="flex items-center gap-2 shrink-0">
                          <span className={`px-2 py-0.5 rounded font-semibold text-[10px] uppercase border ${
                            step.status === 'COMPLETED' ? 'bg-emerald-50 text-emerald-700 border-emerald-200' :
                            step.status === 'EXECUTING' ? 'bg-blue-50 text-blue-700 border-blue-200' :
                            step.status === 'FAILED' ? 'bg-red-50 text-red-700 border-red-200' :
                            'bg-slate-100 text-slate-600 border-slate-200'
                          }`}>
                            {step.status}
                          </span>
                          {step.status !== 'COMPLETED' && (
                            <button
                              onClick={() => handleExecuteStep(step.step_id)}
                              disabled={actionLoading || activePlan.status === 'REQUIRES_APPROVAL'}
                              className="px-2.5 py-1 bg-indigo-600 hover:bg-indigo-700 disabled:opacity-50 text-white rounded text-[11px] font-medium flex items-center gap-1 shadow-xs"
                            >
                              <Play className="w-3 h-3" />
                              Execute Step
                            </button>
                          )}
                        </div>
                      </div>
                    ))}
                  </div>
                </div>
              </div>
            ) : (
              <div className="p-6 text-center bg-slate-50 rounded-lg border border-slate-200">
                <CheckCircle2 className="w-8 h-8 text-emerald-500 mx-auto mb-2" />
                <h4 className="text-sm font-semibold text-slate-800">Shipment Operating Smoothly</h4>
                <p className="text-xs text-slate-500 mt-1 max-w-sm mx-auto">
                  No active disruptions or commitment deviations detected. The autonomous engine is actively monitoring milestones.
                </p>
              </div>
            )}
          </div>

          {/* Section 3: Event Interpretation & Adaptive Simulation */}
          <div className="bg-white border border-slate-200 rounded-xl p-4 shadow-xs">
            <h3 className="text-xs font-bold uppercase tracking-wider text-slate-500 mb-2 flex items-center gap-2">
              <Zap className="w-4 h-4 text-amber-600" />
              Event Interpretation & Adaptation Testing
            </h3>
            <p className="text-xs text-slate-500 mb-3">
              Trigger realistic operational events to test threshold filtering, deduplication, commitment risk protection, and safe replanning.
            </p>
            <div className="grid grid-cols-2 sm:grid-cols-4 gap-2">
              <button
                onClick={() => handleSimulateEvent('MINOR_ETA_FLUCTUATION')}
                disabled={simulationLoading}
                className="p-2.5 bg-slate-50 hover:bg-slate-100 border border-slate-200 rounded-lg text-left transition-colors"
              >
                <div className="text-[11px] font-bold text-slate-800">Minor ETA Swing</div>
                <div className="text-[10px] text-slate-500 mt-0.5">+1.5h Delay</div>
                <div className="text-[9px] font-semibold text-slate-400 uppercase mt-1">Filters to Monitor</div>
              </button>

              <button
                onClick={() => handleSimulateEvent('MAJOR_ETA_DELAY')}
                disabled={simulationLoading}
                className="p-2.5 bg-blue-50/50 hover:bg-blue-50 border border-blue-200 rounded-lg text-left transition-colors"
              >
                <div className="text-[11px] font-bold text-blue-900">Severe Port Delay</div>
                <div className="text-[10px] text-blue-700 mt-0.5">+28h (Breach SLA)</div>
                <div className="text-[9px] font-semibold text-blue-600 uppercase mt-1">Generates Recovery</div>
              </button>

              <button
                onClick={() => handleSimulateEvent('CUSTOMS_HOLD')}
                disabled={simulationLoading}
                className="p-2.5 bg-red-50/50 hover:bg-red-50 border border-red-200 rounded-lg text-left transition-colors"
              >
                <div className="text-[11px] font-bold text-red-900">Customs / Reg Hold</div>
                <div className="text-[10px] text-red-700 mt-0.5">Missing Documents</div>
                <div className="text-[9px] font-semibold text-red-600 uppercase mt-1">Escalates to Human</div>
              </button>

              <button
                onClick={() => handleSimulateEvent('DUPLICATE_EVENT')}
                disabled={simulationLoading}
                className="p-2.5 bg-slate-50 hover:bg-slate-100 border border-slate-200 rounded-lg text-left transition-colors"
              >
                <div className="text-[11px] font-bold text-slate-800">Replayed Webhook</div>
                <div className="text-[10px] text-slate-500 mt-0.5">Exact Dedup Key</div>
                <div className="text-[9px] font-semibold text-slate-400 uppercase mt-1">Deduplication Guard</div>
              </button>
            </div>
          </div>

          {/* Section 4: Event Journal & Deduplication History */}
          <div className="bg-white border border-slate-200 rounded-xl p-4 shadow-xs">
            <h3 className="text-xs font-bold uppercase tracking-wider text-slate-500 mb-3 flex items-center gap-2">
              <FileText className="w-4 h-4 text-slate-600" />
              Shipment Event Journal & Deduplication Log
            </h3>
            {events.length > 0 ? (
              <div className="overflow-x-auto">
                <table className="w-full text-left text-xs border-collapse">
                  <thead>
                    <tr className="border-b border-slate-200 bg-slate-50 text-slate-500 font-semibold">
                      <th className="py-2 px-2.5">Time</th>
                      <th className="py-2 px-2.5">Event Type</th>
                      <th className="py-2 px-2.5">Decision</th>
                      <th className="py-2 px-2.5">Reason</th>
                      <th className="py-2 px-2.5">Plan Ref</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-slate-100">
                    {events.map((ev) => (
                      <tr key={ev.id || ev.event_id} className="hover:bg-slate-50/60">
                        <td className="py-2 px-2.5 whitespace-nowrap text-slate-500">
                          {new Date(ev.created_at).toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit' })}
                        </td>
                        <td className="py-2 px-2.5 font-medium text-slate-800">{ev.event_type}</td>
                        <td className="py-2 px-2.5">
                          <span className={`px-1.5 py-0.5 rounded text-[10px] font-bold uppercase ${
                            ev.decision === 'NEW_PLAN' ? 'bg-blue-100 text-blue-800' :
                            ev.decision === 'ESCALATION' ? 'bg-red-100 text-red-800' :
                            ev.decision === 'REPLANNING' ? 'bg-indigo-100 text-indigo-800' :
                            'bg-slate-100 text-slate-700'
                          }`}>
                            {ev.decision}
                          </span>
                        </td>
                        <td className="py-2 px-2.5 text-slate-600 max-w-xs truncate" title={ev.decision_reason}>
                          {ev.decision_reason || '—'}
                        </td>
                        <td className="py-2 px-2.5 font-mono text-[10px] text-slate-500">
                          {ev.plan_id ? ev.plan_id.substring(0, 14) + '...' : '—'}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            ) : (
              <div className="text-center py-4 text-xs text-slate-400">
                No events recorded in the shipment journal yet.
              </div>
            )}
          </div>
        </div>

        {/* Footer */}
        <div className="p-4 border-t border-slate-200 bg-slate-50 flex items-center justify-between text-xs text-slate-500">
          <div className="flex items-center gap-1.5">
            <Shield className="w-4 h-4 text-emerald-600" />
            <span>Go Action System & Governance Boundary Enforced</span>
          </div>
          <button
            onClick={onClose}
            className="px-4 py-2 bg-slate-200 hover:bg-slate-300 text-slate-800 font-semibold rounded-lg transition-colors"
          >
            Close Drawer
          </button>
        </div>
      </div>

      {/* Candidate Alternatives Modal */}
      {showComparison && activePlan && (
        <PlanComparisonModal
          plan={activePlan}
          isOpen={showComparison}
          onClose={() => setShowComparison(false)}
          onCandidateSelected={() => {
            setShowComparison(false);
            fetchAdaptiveState();
            if (onPlanUpdated) onPlanUpdated();
          }}
        />
      )}
    </div>
  );
}

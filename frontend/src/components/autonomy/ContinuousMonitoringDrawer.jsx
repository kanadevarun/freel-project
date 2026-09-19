import React, { useState, useEffect, useCallback } from 'react';
import {
  X,
  Activity,
  CheckCircle2,
  AlertTriangle,
  Clock,
  RotateCw,
  Shield,
  Layers,
  ArrowRight,
  GitBranch,
  RefreshCw,
  AlertCircle,
  BarChart3,
  Check,
  Pause,
  Play,
  Zap,
  Filter,
  Eye,
  ShieldCheck,
  Lock,
  Compass,
  FileText,
  DollarSign,
  TrendingDown,
  ChevronRight
} from 'lucide-react';
import autonomyService from '../../services/autonomyService';

/**
 * ContinuousMonitoringDrawer.jsx — Phase 5 Task 5.10 Continuous Monitoring & Replanning
 *
 * Controlled adaptive autonomy for LogisticsHQ:
 * MONITOR → DETECT MATERIAL CHANGE → LOAD CONTEXT → ASSESS STATE → CHECK PLAN
 * → VALIDATE ASSUMPTIONS → CONTINUE/PAUSE/REPLAN/ESCALATE/STOP → ACTION SYSTEM → VERIFY
 */
export default function ContinuousMonitoringDrawer({
  isOpen,
  onClose,
  planId,
  initialPlan,
  entityType = 'SHIPMENT',
  entityId = '101',
  onPlanUpdated
}) {
  const [activeTab, setActiveTab] = useState('health'); // 'health' | 'lineage' | 'events' | 'metrics'
  const [loading, setLoading] = useState(false);
  const [actionLoading, setActionLoading] = useState(false);
  const [plan, setPlan] = useState(initialPlan || null);
  const [steps, setSteps] = useState([]);
  const [healthData, setHealthData] = useState(null);
  const [events, setEvents] = useState([]);
  const [metrics, setMetrics] = useState(null);
  const [versions, setVersions] = useState([]);
  const [error, setError] = useState(null);
  const [successMessage, setSuccessMessage] = useState(null);

  // Simulation Form State
  const [simulateEventType, setSimulateEventType] = useState('SHIPMENT_ETA_SLIP');
  const [simulateDelayHours, setSimulateDelayHours] = useState('6.0');
  const [simulateCommitmentBreached, setSimulateCommitmentBreached] = useState(true);
  const [simulateMessage, setSimulateMessage] = useState('');
  const [simulationResult, setSimulationResult] = useState(null);

  const loadData = useCallback(async () => {
    const targetPlanId = planId || initialPlan?.plan_id;
    setLoading(true);
    setError(null);

    // 1. Fetch Plan & Steps
    if (targetPlanId) {
      try {
        const planRes = await autonomyService.getPlan(targetPlanId);
        const planObj = planRes?.plan || planRes?.data?.plan;
        if (planObj) {
          setPlan(planObj);
          setSteps(planRes?.steps || planRes?.data?.steps || []);
        }
      } catch (e) {
        console.warn('Plan not found for targetPlanId:', targetPlanId);
      }

      // 2. Fetch Plan Health
      try {
        const healthRes = await autonomyService.getPlanHealth(targetPlanId);
        const healthObj = healthRes?.health || healthRes?.data?.health;
        if (healthObj) {
          setHealthData(healthObj);
        }
      } catch (e) {
        console.warn('Health data not found for targetPlanId:', targetPlanId);
      }

      // 3. Fetch Plan Versions (Lineage)
      try {
        const verRes = await autonomyService.getPlanVersions(targetPlanId);
        const versionsList = verRes?.versions || verRes?.data?.versions;
        if (versionsList) {
          setVersions(versionsList);
        }
      } catch (e) {
        console.warn('Versions not found for targetPlanId:', targetPlanId);
      }
    }

    // 4. Fetch Monitoring Events
    try {
      const evtRes = await autonomyService.listMonitoringEvents(50);
      const eventsList = evtRes?.events || evtRes?.data?.events;
      if (eventsList) {
        setEvents(eventsList);
      }
    } catch (e) {
      console.warn('Failed listing monitoring events:', e);
    }

    // 5. Fetch Continuous Monitoring Metrics
    try {
      const metRes = await autonomyService.getContinuousMonitoringMetrics();
      const metricsObj = metRes?.metrics || metRes?.data?.metrics;
      if (metricsObj) {
        setMetrics(metricsObj);
      }
    } catch (e) {
      console.warn('Failed fetching metrics:', e);
    }

    setLoading(false);
  }, [planId, initialPlan]);

  useEffect(() => {
    if (isOpen) {
      loadData();
    }
  }, [isOpen, loadData]);

  // Handle Event Ingestion Simulation
  const handleSimulateEvent = async (e) => {
    e?.preventDefault();
    setActionLoading(true);
    setError(null);
    setSuccessMessage(null);
    setSimulationResult(null);

    const payload = {};
    if (simulateEventType === 'SHIPMENT_ETA_SLIP' || simulateEventType === 'SHIPMENT_ETA_UPDATED') {
      payload.eta_delay_hours = parseFloat(simulateDelayHours) || 0;
      payload.commitment_breached = simulateCommitmentBreached;
      payload.reason = 'Carrier AIS berth delay';
    } else if (simulateEventType === 'CUSTOMER_COMMUNICATION_RECEIVED') {
      payload.message = simulateMessage || 'I cannot accept the proposed delivery date. Please expedite or cancel.';
    } else if (simulateEventType === 'INVOICE_PAYMENT_RECEIVED') {
      payload.amount = 45000.0;
      payload.is_full_settlement = true;
    } else if (simulateEventType === 'RATE_EXPIRED') {
      payload.spot_quote_id = 'SP-8821';
      payload.expiration_time = new Date().toISOString();
    } else {
      payload.note = 'Authoritative state event simulation';
    }

    try {
      const eventId = `evt-sim-${Date.now().toString().slice(-6)}`;
      const req = {
        event_id: eventId,
        event_type: simulateEventType,
        entity_type: plan?.related_entity_type || entityType || 'SHIPMENT',
        entity_id: plan?.related_entity_id || entityId || '101',
        source: 'CONTINUOUS_MONITORING_UI',
        payload,
      };

      const res = await autonomyService.ingestMonitoringEvent(req);
      const evalResp = res?.response || res?.data?.response;
      if (evalResp) {
        setSimulationResult(evalResp);
        setSuccessMessage(`Event ${eventId} ingested: Action ${evalResp.recommended_action || 'CONTINUE'}`);
        await loadData();
        if (onPlanUpdated) onPlanUpdated();
      }
    } catch (err) {
      console.error('Failed simulating event ingestion:', err);
      setError(err?.response?.data?.message || err?.message || 'Event ingestion error');
    } finally {
      setActionLoading(false);
    }
  };

  // Manual Adaptive Replan Trigger
  const handleManualReplan = async () => {
    if (!plan) return;
    const reason = window.prompt('Provide reason for manual adaptive replan:', 'Operational priority recalibration by operator');
    if (!reason) return;

    setActionLoading(true);
    setError(null);
    try {
      const res = await autonomyService.triggerAdaptiveReplan(plan.plan_id, reason);
      const updatedPlan = res?.plan || res?.data?.plan;
      if (updatedPlan) {
        setSuccessMessage(`Adaptive Replan successful: Generated Version ${updatedPlan.version}`);
        await loadData();
        if (onPlanUpdated) onPlanUpdated();
      }
    } catch (err) {
      console.error('Manual replan failed:', err);
      setError(err?.response?.data?.message || err?.message || 'Failed initiating adaptive replan');
    } finally {
      setActionLoading(false);
    }
  };

  // Pause / Resume Plan
  const handleTogglePause = async () => {
    if (!plan) return;
    setActionLoading(true);
    setError(null);
    try {
      if (plan.status === 'PAUSED') {
        await autonomyService.resumePlan(plan.plan_id);
        setSuccessMessage('Autonomous workflow resumed successfully');
      } else {
        await autonomyService.pausePlan(plan.plan_id, 'Paused manually by operator via Monitoring Console');
        setSuccessMessage('Autonomous workflow paused');
      }
      await loadData();
      if (onPlanUpdated) onPlanUpdated();
    } catch (err) {
      setError(err.response?.data?.message || err.message || 'State change failed');
    } finally {
      setActionLoading(false);
    }
  };

  if (!isOpen) return null;

  // Format Health Badges
  const getHealthBadge = (health) => {
    switch (health) {
      case 'HEALTHY':
        return <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-semibold bg-emerald-50 text-emerald-700 border border-emerald-200"><CheckCircle2 className="w-3.5 h-3.5 mr-1" /> HEALTHY</span>;
      case 'AT_RISK':
        return <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-semibold bg-amber-50 text-amber-700 border border-amber-200"><AlertTriangle className="w-3.5 h-3.5 mr-1" /> AT RISK</span>;
      case 'STALE':
        return <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-semibold bg-yellow-50 text-yellow-700 border border-yellow-200"><Clock className="w-3.5 h-3.5 mr-1" /> STALE</span>;
      case 'BLOCKED':
        return <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-semibold bg-rose-50 text-rose-700 border border-rose-200"><AlertCircle className="w-3.5 h-3.5 mr-1" /> BLOCKED</span>;
      case 'REPLANNING':
        return <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-semibold bg-sky-50 text-sky-700 border border-sky-200"><RotateCw className="w-3.5 h-3.5 mr-1 animate-spin" /> REPLANNING</span>;
      case 'ESCALATED':
        return <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-semibold bg-red-50 text-red-700 border border-red-200"><Shield className="w-3.5 h-3.5 mr-1" /> ESCALATED</span>;
      case 'COMPLETED':
        return <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-semibold bg-slate-100 text-slate-700 border border-slate-200"><Check className="w-3.5 h-3.5 mr-1" /> COMPLETED</span>;
      default:
        return <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-semibold bg-slate-100 text-slate-600 border border-slate-200">MONITORING</span>;
    }
  };

  const formatText = (val, fallback = '') => {
    if (!val && val !== 0) return fallback;
    if (typeof val === 'object') {
      if (val.String !== undefined) return val.String || fallback;
      return JSON.stringify(val);
    }
    return String(val);
  };

  const currentHealth = formatText(healthData?.plan_health || plan?.plan_health, 'HEALTHY');

  return (
    <div className="fixed inset-0 z-50 overflow-hidden bg-slate-900/30 backdrop-blur-xs flex justify-end">
      <div className="w-full max-w-4xl bg-white h-full shadow-2xl flex flex-col border-l border-slate-200 animate-in slide-in-from-right duration-200">
        {/* Header */}
        <div className="px-6 py-4 border-b border-slate-200 flex items-center justify-between bg-slate-50/70">
          <div>
            <div className="flex items-center gap-2">
              <span className="p-1.5 bg-blue-100 text-blue-700 rounded-md">
                <Activity className="w-4 h-4" />
              </span>
              <h2 className="text-base font-semibold text-slate-900">Continuous Monitoring & Adaptive Replanning</h2>
              <span className="text-xs px-2 py-0.5 font-medium rounded-full bg-slate-200 text-slate-700">
                {plan ? `V${plan.version || 1}` : 'No Plan'}
              </span>
              {getHealthBadge(currentHealth)}
            </div>
            <p className="text-xs text-slate-500 mt-1">
              Autonomous State Observation &bull; Entity {plan?.related_entity_id || entityId} ({plan?.related_entity_type || entityType}) &bull; Controlled Autonomy Loop
            </p>
          </div>
          <button
            onClick={onClose}
            className="p-1.5 text-slate-400 hover:text-slate-600 rounded-md hover:bg-slate-100 transition-colors"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        {/* Global Notifications */}
        {error && (
          <div className="mx-6 mt-3 p-3 bg-red-50 border border-red-200 text-red-700 rounded-md text-xs flex items-center gap-2">
            <AlertCircle className="w-4 h-4 shrink-0" />
            <span className="flex-1">{formatText(error)}</span>
          </div>
        )}
        {successMessage && (
          <div className="mx-6 mt-3 p-3 bg-emerald-50 border border-emerald-200 text-emerald-700 rounded-md text-xs flex items-center gap-2">
            <CheckCircle2 className="w-4 h-4 shrink-0" />
            <span className="flex-1">{formatText(successMessage)}</span>
          </div>
        )}

        {/* Key Metrics Bar */}
        <div className="grid grid-cols-2 sm:grid-cols-4 gap-2 px-6 py-3 border-b border-slate-200 bg-white">
          <div className="p-2.5 rounded-lg border border-slate-100 bg-slate-50/50">
            <div className="text-[11px] font-medium text-slate-500 flex items-center gap-1">
              <Activity className="w-3.5 h-3.5 text-blue-600" /> Plan Health
            </div>
            <div className="text-sm font-bold text-slate-800 mt-0.5">{currentHealth}</div>
          </div>
          <div className="p-2.5 rounded-lg border border-slate-100 bg-slate-50/50">
            <div className="text-[11px] font-medium text-slate-500 flex items-center gap-1">
              <RotateCw className="w-3.5 h-3.5 text-purple-600" /> Replan Iterations
            </div>
            <div className="text-sm font-bold text-slate-800 mt-0.5">
              {formatText(healthData?.replan_count ?? plan?.replan_count ?? 0)} / 3 <span className="text-[10px] text-slate-400 font-normal">(Max Ceiling)</span>
            </div>
          </div>
          <div className="p-2.5 rounded-lg border border-slate-100 bg-slate-50/50">
            <div className="text-[11px] font-medium text-slate-500 flex items-center gap-1">
              <Filter className="w-3.5 h-3.5 text-emerald-600" /> Filtered Events
            </div>
            <div className="text-sm font-bold text-slate-800 mt-0.5">
              {metrics?.events_filtered || 0} <span className="text-[10px] text-slate-400 font-normal">AI Calls Avoided</span>
            </div>
          </div>
          <div className="p-2.5 rounded-lg border border-slate-100 bg-slate-50/50">
            <div className="text-[11px] font-medium text-slate-500 flex items-center gap-1">
              <ShieldCheck className="w-3.5 h-3.5 text-indigo-600" /> Protected Steps
            </div>
            <div className="text-sm font-bold text-slate-800 mt-0.5">
              {Array.isArray(steps) ? steps.filter(s => s.status === 'COMPLETED' || s.status === 'SUCCEEDED').length : 0} Completed
            </div>
          </div>
        </div>

        {/* Operating Control Actions */}
        <div className="px-6 py-2.5 border-b border-slate-200 bg-slate-50 flex items-center justify-between flex-wrap gap-2">
          <div className="flex items-center gap-2">
            <button
              onClick={handleTogglePause}
              disabled={actionLoading || !plan}
              className="px-3 py-1.5 text-xs font-medium rounded-md border border-slate-300 bg-white text-slate-700 hover:bg-slate-50 transition-colors flex items-center gap-1.5 shadow-xs"
            >
              {plan?.status === 'PAUSED' ? <Play className="w-3.5 h-3.5 text-emerald-600" /> : <Pause className="w-3.5 h-3.5 text-amber-600" />}
              {plan?.status === 'PAUSED' ? 'Resume Plan' : 'Pause Plan'}
            </button>
            <button
              onClick={handleManualReplan}
              disabled={actionLoading || !plan || (plan?.replan_count >= 3)}
              className="px-3 py-1.5 text-xs font-medium rounded-md border border-indigo-200 bg-indigo-50 text-indigo-700 hover:bg-indigo-100 transition-colors flex items-center gap-1.5 shadow-xs disabled:opacity-50"
            >
              <RotateCw className="w-3.5 h-3.5 text-indigo-600" />
              Adaptive Replan
            </button>
          </div>
          <div className="flex items-center gap-2">
            <button
              onClick={loadData}
              disabled={loading}
              className="p-1.5 text-slate-600 hover:bg-slate-200 rounded transition-colors"
              title="Refresh Monitoring Data"
            >
              <RefreshCw className={`w-4 h-4 ${loading ? 'animate-spin' : ''}`} />
            </button>
          </div>
        </div>

        {/* Subtabs Navigation */}
        <div className="flex border-b border-slate-200 px-6 bg-white gap-6">
          <button
            onClick={() => setActiveTab('health')}
            className={`py-3 text-xs font-medium border-b-2 flex items-center gap-1.5 transition-colors ${
              activeTab === 'health'
                ? 'border-blue-600 text-blue-600'
                : 'border-transparent text-slate-500 hover:text-slate-800'
            }`}
          >
            <Activity className="w-3.5 h-3.5" /> Plan Health & Assumptions
          </button>
          <button
            onClick={() => setActiveTab('lineage')}
            className={`py-3 text-xs font-medium border-b-2 flex items-center gap-1.5 transition-colors ${
              activeTab === 'lineage'
                ? 'border-blue-600 text-blue-600'
                : 'border-transparent text-slate-500 hover:text-slate-800'
            }`}
          >
            <GitBranch className="w-3.5 h-3.5" /> Plan Version Lineage
          </button>
          <button
            onClick={() => setActiveTab('events')}
            className={`py-3 text-xs font-medium border-b-2 flex items-center gap-1.5 transition-colors ${
              activeTab === 'events'
                ? 'border-blue-600 text-blue-600'
                : 'border-transparent text-slate-500 hover:text-slate-800'
            }`}
          >
            <Filter className="w-3.5 h-3.5" /> Material Event Stream
          </button>
          <button
            onClick={() => setActiveTab('simulate')}
            className={`py-3 text-xs font-medium border-b-2 flex items-center gap-1.5 transition-colors ${
              activeTab === 'simulate'
                ? 'border-blue-600 text-blue-600'
                : 'border-transparent text-slate-500 hover:text-slate-800'
            }`}
          >
            <Zap className="w-3.5 h-3.5 text-amber-500" /> Event Simulation
          </button>
        </div>

        {/* Content Area */}
        <div className="flex-1 overflow-y-auto p-6 space-y-6">
          {/* TAB 1: PLAN HEALTH & ASSUMPTIONS */}
          {activeTab === 'health' && (
            <div className="space-y-6">
              {/* Health Rationale Card */}
              <div className="p-4 rounded-lg border border-slate-200 bg-white shadow-xs">
                <h3 className="text-xs font-semibold text-slate-900 uppercase tracking-wider mb-2 flex items-center gap-2">
                  <Shield className="w-4 h-4 text-blue-600" />
                  Health Status & Authoritative Diagnostics
                </h3>
                <div className="p-3 bg-slate-50 rounded-md border border-slate-100 text-xs text-slate-700">
                  <span className="font-semibold text-slate-900">Current Assessment: </span>
                  {formatText(healthData?.health_reason || plan?.health_reason, 'Active workflow is healthy and aligned with authoritative telemetry.')}
                </div>
              </div>

              {/* Assumptions & Changed Invalidation */}
              <div className="p-4 rounded-lg border border-slate-200 bg-white shadow-xs">
                <h3 className="text-xs font-semibold text-slate-900 uppercase tracking-wider mb-3 flex items-center gap-2">
                  <Compass className="w-4 h-4 text-indigo-600" />
                  Assumption Validation & Invalidation Tracking
                </h3>
                <div className="space-y-2">
                  {/* Changed / Invalidated Assumptions */}
                  {Array.isArray(healthData?.changed_assumptions) && healthData.changed_assumptions.length > 0 ? (
                    healthData.changed_assumptions.map((assump, idx) => (
                      <div key={idx} className="flex items-start gap-2.5 p-2.5 bg-rose-50 border border-rose-100 rounded-md text-xs text-rose-800">
                        <AlertTriangle className="w-4 h-4 text-rose-600 shrink-0 mt-0.5" />
                        <div>
                          <div className="font-semibold">[INVALIDATED ASSUMPTION]</div>
                          <div>{formatText(assump)}</div>
                        </div>
                      </div>
                    ))
                  ) : (
                    <div className="flex items-center gap-2 p-2.5 bg-emerald-50 border border-emerald-100 rounded-md text-xs text-emerald-800">
                      <CheckCircle2 className="w-4 h-4 text-emerald-600 shrink-0" />
                      <span>All underlying operational assumptions remain valid and verified against authoritative records.</span>
                    </div>
                  )}
                </div>
              </div>

              {/* Active Plan Steps with Protection Badges */}
              <div className="p-4 rounded-lg border border-slate-200 bg-white shadow-xs">
                <h3 className="text-xs font-semibold text-slate-900 uppercase tracking-wider mb-3 flex items-center gap-2">
                  <Layers className="w-4 h-4 text-slate-700" />
                  Active Execution Graph Steps
                </h3>
                <div className="space-y-2">
                  {steps.map((st, idx) => {
                    const isCompleted = st.status === 'COMPLETED' || st.status === 'SUCCEEDED';
                    return (
                      <div
                        key={st.step_id || idx}
                        className={`p-3 rounded-lg border flex items-center justify-between text-xs ${
                          isCompleted
                            ? 'bg-slate-50/80 border-slate-200 text-slate-600'
                            : st.status === 'EXECUTING'
                            ? 'bg-blue-50/50 border-blue-200 text-blue-900'
                            : 'bg-white border-slate-200 text-slate-800'
                        }`}
                      >
                        <div className="flex items-center gap-3">
                          <span className="w-5 h-5 rounded-full flex items-center justify-center text-[11px] font-bold bg-slate-200 text-slate-700">
                            {st.step_number || idx + 1}
                          </span>
                          <div>
                            <div className="font-semibold text-slate-900 flex items-center gap-2">
                              {st.title}
                              {isCompleted && (
                                <span className="inline-flex items-center px-2 py-0.5 text-[10px] font-medium rounded-full bg-indigo-50 text-indigo-700 border border-indigo-200" title="Completed step protected from repetition">
                                  <Lock className="w-3 h-3 mr-1 text-indigo-600" /> Protected
                                </span>
                              )}
                            </div>
                            <div className="text-[11px] text-slate-500 font-mono mt-0.5">{st.action_type}</div>
                          </div>
                        </div>
                        <div className="flex items-center gap-2">
                          <span className={`px-2 py-0.5 text-[10px] font-semibold rounded-full ${
                            isCompleted ? 'bg-emerald-100 text-emerald-800' : 'bg-slate-100 text-slate-700'
                          }`}>
                            {st.status}
                          </span>
                        </div>
                      </div>
                    );
                  })}
                </div>
              </div>
            </div>
          )}

          {/* TAB 2: LINEAGE & ADAPTIVE VERSIONS */}
          {activeTab === 'lineage' && (
            <div className="space-y-6">
              <div className="p-4 rounded-lg border border-slate-200 bg-white shadow-xs">
                <h3 className="text-xs font-semibold text-slate-900 uppercase tracking-wider mb-2 flex items-center gap-2">
                  <GitBranch className="w-4 h-4 text-purple-600" />
                  Plan Version History & Immutable Lineage
                </h3>
                <p className="text-xs text-slate-500 mb-4">
                  Each material event trigger formulations of new version branches (V1 &rarr; V2 &rarr; V3) without overwriting historical execution records.
                </p>

                <div className="relative border-l-2 border-slate-200 ml-4 pl-6 space-y-6">
                  {versions.length > 0 ? (
                    versions.map((ver, idx) => (
                      <div key={ver.plan_id || idx} className="relative">
                        <div className="absolute -left-[31px] top-1.5 w-4 h-4 rounded-full bg-white border-2 border-purple-600" />
                        <div className="p-3 bg-slate-50 rounded-lg border border-slate-200 text-xs">
                          <div className="flex items-center justify-between">
                            <span className="font-bold text-slate-900">Version {ver.version}</span>
                            <span className="text-slate-400 font-mono text-[10px]">{ver.plan_id}</span>
                          </div>
                          <div className="text-slate-700 mt-1 font-medium">{ver.goal}</div>
                          <div className="mt-2 flex items-center gap-3 text-[11px] text-slate-500">
                            <span>Status: <strong className="text-slate-700">{ver.status}</strong></span>
                            <span>Confidence: <strong className="text-slate-700">{(ver.confidence_score * 100).toFixed(0)}%</strong></span>
                            <span>Replans: <strong className="text-slate-700">{ver.replan_count}</strong></span>
                          </div>
                        </div>
                      </div>
                    ))
                  ) : (
                    <div className="text-xs text-slate-500 italic">No previous plan versions recorded for this entity.</div>
                  )}
                </div>
              </div>
            </div>
          )}

          {/* TAB 3: MATERIAL EVENT STREAM */}
          {activeTab === 'events' && (
            <div className="space-y-6">
              <div className="p-4 rounded-lg border border-slate-200 bg-white shadow-xs">
                <div className="flex items-center justify-between mb-3">
                  <h3 className="text-xs font-semibold text-slate-900 uppercase tracking-wider flex items-center gap-2">
                    <Filter className="w-4 h-4 text-emerald-600" />
                    Incoming Operational Business Events
                  </h3>
                  <span className="text-xs text-slate-500">Latest {events.length} Events</span>
                </div>

                <div className="space-y-2">
                  {events.map((evt, idx) => (
                    <div
                      key={evt.event_id || idx}
                      className={`p-3 rounded-lg border text-xs ${
                        evt.is_material
                          ? 'bg-amber-50/50 border-amber-200'
                          : 'bg-slate-50 border-slate-200 text-slate-600'
                      }`}
                    >
                      <div className="flex items-center justify-between">
                        <div className="flex items-center gap-2">
                          <span className={`px-2 py-0.5 text-[10px] font-bold rounded-full ${
                            evt.is_material ? 'bg-amber-100 text-amber-800' : 'bg-slate-200 text-slate-700'
                          }`}>
                            {evt.is_material ? 'MATERIAL' : 'FILTERED'}
                          </span>
                          <span className="font-semibold text-slate-900">{evt.event_type}</span>
                          <span className="text-slate-400 font-mono text-[10px]">({evt.event_id})</span>
                        </div>
                        <span className="text-[11px] text-slate-400">
                          {evt.created_at ? new Date(evt.created_at).toLocaleTimeString() : 'Just now'}
                        </span>
                      </div>
                      <div className="mt-1 text-slate-600">
                        <strong>Analysis:</strong> {evt.filter_reason || 'Routine event ingested'}
                      </div>
                      <div className="mt-1 flex items-center gap-4 text-[10px] text-slate-500 font-mono">
                        <span>Source: {evt.source}</span>
                        <span>Entity: {evt.entity_type}:{evt.entity_id}</span>
                        {evt.replan_triggered && <span className="text-purple-600 font-bold">&bull; Replan Triggered</span>}
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            </div>
          )}

          {/* TAB 4: EVENT SIMULATION FORM */}
          {activeTab === 'simulate' && (
            <div className="space-y-6">
              <div className="p-4 rounded-lg border border-slate-200 bg-white shadow-xs">
                <h3 className="text-xs font-semibold text-slate-900 uppercase tracking-wider mb-2 flex items-center gap-2">
                  <Zap className="w-4 h-4 text-amber-500" />
                  Simulate Operational Business Event
                </h3>
                <p className="text-xs text-slate-500 mb-4">
                  Inject real business events into the Go event boundary to test deterministic filtering, material change detection, and adaptive replanning.
                </p>

                <form onSubmit={handleSimulateEvent} className="space-y-4">
                  <div>
                    <label className="block text-xs font-medium text-slate-700 mb-1">Event Type</label>
                    <select
                      value={simulateEventType}
                      onChange={(e) => setSimulateEventType(e.target.value)}
                      className="w-full text-xs rounded-md border border-slate-300 p-2 bg-white"
                    >
                      <option value="SHIPMENT_ETA_SLIP">SHIPMENT_ETA_SLIP (Material &ge;3h Delay)</option>
                      <option value="SHIPMENT_ETA_UPDATED">SHIPMENT_ETA_UPDATED (Minor &lt;3h Non-Material)</option>
                      <option value="CUSTOMER_COMMUNICATION_RECEIVED">CUSTOMER_COMMUNICATION_RECEIVED (Customer Objection)</option>
                      <option value="INVOICE_PAYMENT_RECEIVED">INVOICE_PAYMENT_RECEIVED (Full Settlement &bull; Stop Workflow)</option>
                      <option value="RATE_EXPIRED">RATE_EXPIRED (Pricing Quotation Invalidation)</option>
                      <option value="CARRIER_HOLD_CRITICAL">CARRIER_HOLD_CRITICAL (Terminal Operational Block)</option>
                    </select>
                  </div>

                  {(simulateEventType === 'SHIPMENT_ETA_SLIP' || simulateEventType === 'SHIPMENT_ETA_UPDATED') && (
                    <div className="grid grid-cols-2 gap-4">
                      <div>
                        <label className="block text-xs font-medium text-slate-700 mb-1">Delay Duration (Hours)</label>
                        <input
                          type="number"
                          step="0.5"
                          value={simulateDelayHours}
                          onChange={(e) => setSimulateDelayHours(e.target.value)}
                          className="w-full text-xs rounded-md border border-slate-300 p-2 bg-white"
                        />
                      </div>
                      <div className="flex items-center pt-5">
                        <label className="flex items-center gap-2 text-xs text-slate-700">
                          <input
                            type="checkbox"
                            checked={simulateCommitmentBreached}
                            onChange={(e) => setSimulateCommitmentBreached(e.target.checked)}
                            className="rounded border-slate-300 text-blue-600"
                          />
                          Customer SLA Commitment Breached
                        </label>
                      </div>
                    </div>
                  )}

                  {simulateEventType === 'CUSTOMER_COMMUNICATION_RECEIVED' && (
                    <div>
                      <label className="block text-xs font-medium text-slate-700 mb-1">Customer Message Payload</label>
                      <textarea
                        rows={2}
                        value={simulateMessage}
                        onChange={(e) => setSimulateMessage(e.target.value)}
                        placeholder="I cannot accept the proposed delivery date. Please expedite or cancel."
                        className="w-full text-xs rounded-md border border-slate-300 p-2 bg-white"
                      />
                    </div>
                  )}

                  <button
                    type="submit"
                    disabled={actionLoading}
                    className="w-full py-2 px-4 rounded-md text-xs font-semibold bg-blue-600 text-white hover:bg-blue-700 transition-colors flex items-center justify-center gap-2 shadow-xs"
                  >
                    {actionLoading ? <RotateCw className="w-3.5 h-3.5 animate-spin" /> : <Zap className="w-3.5 h-3.5" />}
                    Ingest & Evaluate Event
                  </button>
                </form>

                {/* Simulation Output Card */}
                {simulationResult && (
                  <div className="mt-4 p-3 bg-slate-50 border border-slate-200 rounded-md text-xs space-y-1.5">
                    <div className="font-semibold text-slate-900 flex items-center justify-between">
                      <span>Evaluation Result:</span>
                      <span className="font-mono text-slate-500">{simulationResult.event_id}</span>
                    </div>
                    <div className="text-slate-700">
                      <strong>Material:</strong> {simulationResult.is_material ? 'Yes' : 'No (Filtered Deterministically)'}
                    </div>
                    <div className="text-slate-700">
                      <strong>Recommended Action:</strong> <span className="font-bold text-blue-700">{simulationResult.recommended_action || 'CONTINUE'}</span>
                    </div>
                    <div className="text-slate-700">
                      <strong>Analysis:</strong> {simulationResult.filter_reason}
                    </div>
                    {simulationResult.replan_triggered && (
                      <div className="p-2 bg-purple-50 border border-purple-200 text-purple-800 rounded font-medium">
                        &bull; Plan Adapted to Version {simulationResult.new_version} ({simulationResult.new_plan_id})
                      </div>
                    )}
                  </div>
                )}
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

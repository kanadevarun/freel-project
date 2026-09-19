import React, { useState, useEffect, useCallback } from 'react';
import {
  X,
  Shield,
  AlertOctagon,
  CheckCircle2,
  AlertTriangle,
  RefreshCw,
  Search,
  Filter,
  Sliders,
  Lock,
  Unlock,
  Key,
  DollarSign,
  Activity,
  History,
  FileText,
  Play,
  Check,
  Zap,
  Layers,
  ChevronRight,
  Info,
  Scale,
  Settings,
  Eye,
  Plus
} from 'lucide-react';
import autonomyService from '../../services/autonomyService';

export default function ControlledAutonomyGovernanceDrawer({
  isOpen,
  onClose,
  initialTab = 'overview',
  onPolicyUpdated
}) {
  const [activeTab, setActiveTab] = useState(initialTab);
  const [loading, setLoading] = useState(false);
  const [actionLoading, setActionLoading] = useState(false);
  const [error, setError] = useState(null);
  const [successMsg, setSuccessMsg] = useState(null);

  // Core data states
  const [limits, setLimits] = useState(null);
  const [telemetry, setTelemetry] = useState(null);
  const [allowlist, setAllowlist] = useState([]);
  const [flags, setFlags] = useState([]);
  const [evaluations, setEvaluations] = useState([]);
  const [auditLogs, setAuditLogs] = useState([]);

  // Filters
  const [allowlistModuleFilter, setAllowlistModuleFilter] = useState('ALL');
  const [allowlistSearch, setAllowlistSearch] = useState('');
  const [evalDecisionFilter, setEvalDecisionFilter] = useState('ALL');

  // Kill switch modal state
  const [killSwitchModalOpen, setKillSwitchModalOpen] = useState(false);
  const [killSwitchReason, setKillSwitchReason] = useState('');

  // Edit limits state
  const [isEditingLimits, setIsEditingLimits] = useState(false);
  const [limitsForm, setLimitsForm] = useState({
    max_tenant_autonomy: 3,
    max_actions_per_hour: 100,
    max_financial_exposure_per_workflow: 5000,
    max_retries_per_step: 3,
    max_replans_per_plan: 5,
    enforce_four_eyes: true
  });

  // Simulator / Matrix State
  const [simForm, setSimForm] = useState({
    module: 'shipments',
    action_type: 'carrier.escalate',
    entity_type: 'SHIPMENT',
    entity_id: '101',
    user_role: 'operations',
    user_permissions: 'shipments:update',
    preparer_user_id: 10,
    approval_approver_id: 20,
    is_approved: false,
    compliance_status: 'COMPLIANT',
    entity_state: 'ACTIVE',
    requested_autonomy: 3,
    financial_amount: 0,
    context_text: 'Operational delay requires formal carrier escalation.'
  });
  const [simResult, setSimResult] = useState(null);
  const [simLoading, setSimLoading] = useState(false);

  // Load all governance data
  const loadData = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const [limRes, telRes, alRes, flRes, evRes, auRes] = await Promise.all([
        autonomyService.getGovernanceLimits().catch(() => ({ data: { limits: null } })),
        autonomyService.getGovernanceTelemetry().catch(() => ({ data: { telemetry: null } })),
        autonomyService.getActionAllowlist().catch(() => ({ data: { items: [] } })),
        autonomyService.getGovernanceFeatureFlags().catch(() => ({ data: { flags: [] } })),
        autonomyService.listPolicyEvaluations({ limit: 50 }).catch(() => ({ data: { evaluations: [] } })),
        autonomyService.listPolicyAuditLogs({ limit: 50 }).catch(() => ({ data: { logs: [] } }))
      ]);

      const lim = limRes.limits || limRes.data?.limits;
      if (lim) {
        setLimits(lim);
        setLimitsForm({
          max_tenant_autonomy: lim.max_tenant_autonomy ?? 3,
          max_actions_per_hour: lim.max_actions_per_hour ?? 100,
          max_financial_exposure_per_workflow: lim.max_financial_exposure_per_workflow ?? 5000,
          max_retries_per_step: lim.max_retries_per_step ?? 3,
          max_replans_per_plan: lim.max_replans_per_plan ?? 5,
          enforce_four_eyes: lim.enforce_four_eyes ?? true
        });
      }
      setTelemetry(telRes.telemetry || telRes.data?.telemetry || null);
      setAllowlist(alRes.items || alRes.data?.items || []);
      setFlags(flRes.flags || flRes.data?.flags || []);
      setEvaluations(evRes.evaluations || evRes.data?.evaluations || []);
      setAuditLogs(auRes.logs || auRes.data?.logs || []);
    } catch (err) {
      setError(err.response?.data?.message || err.message || 'Failed to load governance data');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    if (isOpen) {
      loadData();
    }
  }, [isOpen, loadData]);

  useEffect(() => {
    if (initialTab) {
      setActiveTab(initialTab);
    }
  }, [initialTab]);

  const showNotification = (msg) => {
    setSuccessMsg(msg);
    setTimeout(() => setSuccessMsg(null), 4000);
  };

  // Toggle Kill Switch
  const handleToggleKillSwitch = async () => {
    if (!limits) return;
    setActionLoading(true);
    try {
      const targetState = !limits.kill_switch_active;
      await autonomyService.toggleGovernanceKillSwitch(targetState, killSwitchReason || 'Manual administrative toggle');
      showNotification(targetState ? 'EMERGENCY KILL SWITCH ACTIVATED — Autonomous execution halted.' : 'Emergency kill switch deactivated. Normal autonomy resumed.');
      setKillSwitchModalOpen(false);
      setKillSwitchReason('');
      await loadData();
      if (onPolicyUpdated) onPolicyUpdated();
    } catch (err) {
      setError(err.response?.data?.message || err.message || 'Failed to toggle kill switch');
    } finally {
      setActionLoading(false);
    }
  };

  // Save Tenant Limits
  const handleSaveLimits = async (e) => {
    e.preventDefault();
    setActionLoading(true);
    try {
      await autonomyService.updateGovernanceLimits(limitsForm);
      showNotification('Tenant governance limits updated successfully.');
      setIsEditingLimits(false);
      await loadData();
      if (onPolicyUpdated) onPolicyUpdated();
    } catch (err) {
      setError(err.response?.data?.message || err.message || 'Failed to update tenant limits');
    } finally {
      setActionLoading(false);
    }
  };

  // Toggle Feature Flag
  const handleToggleFlag = async (flagKey, currentEnabled, maxAutonomy, reqApproval) => {
    setActionLoading(true);
    try {
      await autonomyService.updateGovernanceFeatureFlag(flagKey, {
        enabled: !currentEnabled,
        max_autonomy_level: maxAutonomy,
        requires_approval: reqApproval
      });
      showNotification(`Feature flag '${flagKey}' ${!currentEnabled ? 'enabled' : 'disabled'}.`);
      await loadData();
      if (onPolicyUpdated) onPolicyUpdated();
    } catch (err) {
      setError(err.response?.data?.message || err.message || 'Failed to update feature flag');
    } finally {
      setActionLoading(false);
    }
  };

  // Update Flag Autonomy Level
  const handleUpdateFlagAutonomy = async (flagKey, isEnabled, newLevel, reqApproval) => {
    setActionLoading(true);
    try {
      await autonomyService.updateGovernanceFeatureFlag(flagKey, {
        enabled: isEnabled,
        max_autonomy_level: parseInt(newLevel, 10),
        requires_approval: reqApproval
      });
      showNotification(`Feature flag '${flagKey}' max autonomy set to Level ${newLevel}.`);
      await loadData();
      if (onPolicyUpdated) onPolicyUpdated();
    } catch (err) {
      setError(err.response?.data?.message || err.message || 'Failed to update autonomy tier');
    } finally {
      setActionLoading(false);
    }
  };

  // Run Simulator / Matrix Check
  const handleRunSimulation = async (e) => {
    e.preventDefault();
    setSimLoading(true);
    setSimResult(null);
    try {
      const permsArray = simForm.user_permissions
        .split(',')
        .map((p) => p.trim())
        .filter(Boolean);

      const payload = {
        module: simForm.module,
        action_type: simForm.action_type,
        entity_type: simForm.entity_type,
        entity_id: simForm.entity_id,
        user_role: simForm.user_role,
        user_permissions: permsArray,
        preparer_user_id: parseInt(simForm.preparer_user_id, 10) || 0,
        approval_approver_id: parseInt(simForm.approval_approver_id, 10) || 0,
        is_approved: simForm.is_approved,
        compliance_status: simForm.compliance_status,
        entity_state: simForm.entity_state,
        requested_autonomy: parseInt(simForm.requested_autonomy, 10),
        parameters: {
          cost_impact: parseFloat(simForm.financial_amount) || 0.0,
          amount: parseFloat(simForm.financial_amount) || 0.0
        },
        context_text: simForm.context_text
      };

      const res = await autonomyService.evaluateGovernancePolicy(payload);
      setSimResult(res.result || res.data?.result || res);
    } catch (err) {
      setError(err.response?.data?.message || err.message || 'Policy simulation failed');
    } finally {
      setSimLoading(false);
    }
  };

  if (!isOpen) return null;

  const filteredAllowlist = allowlist.filter((item) => {
    const matchModule = allowlistModuleFilter === 'ALL' || item.module === allowlistModuleFilter;
    const matchSearch =
      !allowlistSearch ||
      item.action_type.toLowerCase().includes(allowlistSearch.toLowerCase()) ||
      item.action_name.toLowerCase().includes(allowlistSearch.toLowerCase());
    return matchModule && matchSearch;
  });

  const filteredEvaluations = evaluations.filter((ev) => {
    if (evalDecisionFilter === 'ALL') return true;
    return ev.decision === evalDecisionFilter;
  });

  return (
    <div className="fixed inset-0 z-50 overflow-hidden bg-slate-900/60 backdrop-blur-xs flex justify-end animate-in fade-in duration-200">
      <div className="w-full max-w-5xl bg-white h-full shadow-2xl flex flex-col border-l border-slate-200 text-slate-800">
        {/* Top Header */}
        <div className="bg-[#0B192C] text-white px-6 py-4 flex items-center justify-between border-b border-slate-700">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 rounded-lg bg-blue-600/30 border border-blue-400/40 flex items-center justify-center text-blue-400">
              <Shield className="w-6 h-6" />
            </div>
            <div>
              <div className="flex items-center gap-2">
                <h2 className="text-lg font-semibold tracking-tight text-white">Controlled Autonomy Governance</h2>
                <span className="text-xs px-2 py-0.5 rounded-full font-mono font-medium bg-blue-900/60 border border-blue-500/40 text-blue-300">
                  v5.14
                </span>
                {limits?.kill_switch_active && (
                  <span className="text-xs px-2 py-0.5 rounded-full font-bold bg-red-600 text-white animate-pulse">
                    KILL SWITCH ACTIVE
                  </span>
                )}
              </div>
              <p className="text-xs text-slate-300 mt-0.5">
                Centralized deterministic policy enforcement, blast-radius boundaries, and human-in-the-loop controls
              </p>
            </div>
          </div>

          <div className="flex items-center gap-3">
            <button
              id="gov-refresh-btn"
              onClick={loadData}
              disabled={loading}
              className="p-2 text-slate-300 hover:text-white rounded-lg hover:bg-slate-800 transition-colors"
              title="Refresh Governance State"
            >
              <RefreshCw className={`w-4 h-4 ${loading ? 'animate-spin' : ''}`} />
            </button>
            <button
              id="gov-close-drawer-btn"
              onClick={onClose}
              className="p-2 text-slate-300 hover:text-white rounded-lg hover:bg-slate-800 transition-colors"
              title="Close Drawer"
            >
              <X className="w-5 h-5" />
            </button>
          </div>
        </div>

        {/* Global Notifications */}
        {successMsg && (
          <div className="bg-emerald-50 border-b border-emerald-200 px-6 py-2.5 text-sm text-emerald-800 flex items-center gap-2 animate-in slide-in-from-top-2">
            <CheckCircle2 className="w-4 h-4 text-emerald-600 shrink-0" />
            <span className="font-medium">{successMsg}</span>
          </div>
        )}

        {error && (
          <div className="bg-red-50 border-b border-red-200 px-6 py-2.5 text-sm text-red-800 flex items-center justify-between animate-in slide-in-from-top-2">
            <div className="flex items-center gap-2">
              <AlertTriangle className="w-4 h-4 text-red-600 shrink-0" />
              <span>{error}</span>
            </div>
            <button onClick={() => setError(null)} className="text-red-600 hover:text-red-800">
              <X className="w-4 h-4" />
            </button>
          </div>
        )}

        {/* Navigation Tabs */}
        <div className="bg-slate-100 border-b border-slate-200 px-6 flex items-center gap-1 overflow-x-auto">
          {[
            { id: 'overview', label: 'Overview & Kill Switch', icon: Shield },
            { id: 'allowlist', label: 'Action Allowlist', icon: Scale },
            { id: 'flags', label: 'Feature Flags', icon: Sliders },
            { id: 'matrix', label: 'Policy Matrix Simulator', icon: Play },
            { id: 'audit', label: 'Audit & Telemetry', icon: History }
          ].map((tab) => {
            const Icon = tab.icon;
            const isActive = activeTab === tab.id;
            return (
              <button
                key={tab.id}
                id={`gov-tab-${tab.id}`}
                onClick={() => setActiveTab(tab.id)}
                className={`flex items-center gap-2 px-4 py-3 text-xs font-medium border-b-2 transition-all whitespace-nowrap ${
                  isActive
                    ? 'border-blue-600 text-blue-700 bg-white shadow-xs font-semibold'
                    : 'border-transparent text-slate-600 hover:text-slate-900 hover:bg-slate-200/60'
                }`}
              >
                <Icon className={`w-4 h-4 ${isActive ? 'text-blue-600' : 'text-slate-500'}`} />
                {tab.label}
              </button>
            );
          })}
        </div>

        {/* Tab Contents Area */}
        <div className="flex-1 overflow-y-auto p-6 bg-slate-50/50">
          {/* TAB 1: OVERVIEW & KILL SWITCH */}
          {activeTab === 'overview' && (
            <div className="space-y-6">
              {/* Emergency Kill Switch Banner */}
              <div
                className={`p-5 rounded-xl border transition-all ${
                  limits?.kill_switch_active
                    ? 'bg-red-50 border-red-300 text-red-950 shadow-md'
                    : 'bg-white border-slate-200 text-slate-900 shadow-xs'
                }`}
              >
                <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
                  <div className="flex items-start gap-3">
                    <div
                      className={`w-10 h-10 rounded-lg flex items-center justify-center shrink-0 ${
                        limits?.kill_switch_active
                          ? 'bg-red-600 text-white animate-pulse'
                          : 'bg-emerald-100 text-emerald-700 border border-emerald-300'
                      }`}
                    >
                      <AlertOctagon className="w-6 h-6" />
                    </div>
                    <div>
                      <h3 className="text-sm font-semibold flex items-center gap-2">
                        Tenant-Wide Emergency Stop (Kill Switch)
                        {limits?.kill_switch_active ? (
                          <span className="text-xs px-2 py-0.5 rounded-full bg-red-600 text-white font-bold">
                            ACTIVATED
                          </span>
                        ) : (
                          <span className="text-xs px-2 py-0.5 rounded-full bg-emerald-100 text-emerald-800 border border-emerald-300 font-medium">
                            STANDBY (NORMAL OPS)
                          </span>
                        )}
                      </h3>
                      <p className="text-xs text-slate-600 mt-1 max-w-xl leading-relaxed">
                        {limits?.kill_switch_active
                          ? `Autonomous execution is currently STOPPED. All autonomous actions are blocked immediately until deactivated. Reason: "${limits.kill_switch_reason || 'Administrative intervention'}"`
                          : 'Immediately halts all autonomous plan executions, actions, and auto-dispatches across all operational modules in this tenant.'}
                      </p>
                    </div>
                  </div>

                  <div className="flex items-center gap-3">
                    <button
                      id="gov-kill-switch-btn"
                      onClick={() => setKillSwitchModalOpen(true)}
                      className={`px-4 py-2.5 rounded-lg text-xs font-semibold flex items-center gap-2 shadow-xs transition-all ${
                        limits?.kill_switch_active
                          ? 'bg-emerald-600 hover:bg-emerald-700 text-white'
                          : 'bg-red-600 hover:bg-red-700 text-white'
                      }`}
                    >
                      <AlertOctagon className="w-4 h-4" />
                      {limits?.kill_switch_active ? 'Deactivate Emergency Stop' : 'ACTIVATE EMERGENCY STOP'}
                    </button>
                  </div>
                </div>
              </div>

              {/* Telemetry Snapshot Cards */}
              <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
                <div className="bg-white p-4 rounded-xl border border-slate-200 shadow-xs">
                  <span className="text-xs font-medium text-slate-500 block">Total Evaluations</span>
                  <span className="text-2xl font-bold text-[#0B192C] mt-1 block">
                    {telemetry?.total_evaluations ?? 0}
                  </span>
                  <span className="text-[11px] text-slate-400 mt-1 block">Recorded by policy engine</span>
                </div>

                <div className="bg-white p-4 rounded-xl border border-slate-200 shadow-xs">
                  <span className="text-xs font-medium text-slate-500 block">Allowed (Permitted)</span>
                  <span className="text-2xl font-bold text-emerald-600 mt-1 block">
                    {telemetry?.allowed_count ?? 0}
                  </span>
                  <span className="text-[11px] text-slate-400 mt-1 block">Within autonomous limits</span>
                </div>

                <div className="bg-white p-4 rounded-xl border border-slate-200 shadow-xs">
                  <span className="text-xs font-medium text-slate-500 block">Requires Review</span>
                  <span className="text-2xl font-bold text-amber-600 mt-1 block">
                    {telemetry?.review_required_count ?? 0}
                  </span>
                  <span className="text-[11px] text-slate-400 mt-1 block">Human review gated</span>
                </div>

                <div className="bg-white p-4 rounded-xl border border-slate-200 shadow-xs">
                  <span className="text-xs font-medium text-slate-500 block">Blocked (Prohibited)</span>
                  <span className="text-2xl font-bold text-red-600 mt-1 block">
                    {telemetry?.blocked_count ?? 0}
                  </span>
                  <span className="text-[11px] text-slate-400 mt-1 block">Policy or safety violations</span>
                </div>
              </div>

              {/* Tenant Autonomy Bounds & Configuration */}
              <div className="bg-white rounded-xl border border-slate-200 shadow-xs overflow-hidden">
                <div className="px-5 py-3.5 border-b border-slate-200 flex items-center justify-between bg-slate-50">
                  <div className="flex items-center gap-2">
                    <Settings className="w-4 h-4 text-blue-600" />
                    <h4 className="text-xs font-semibold text-slate-800 uppercase tracking-wider">
                      Tenant Autonomy & Safety Limits
                    </h4>
                  </div>
                  {!isEditingLimits ? (
                    <button
                      id="gov-edit-limits-btn"
                      onClick={() => setIsEditingLimits(true)}
                      className="text-xs font-medium text-blue-600 hover:text-blue-800 transition-colors"
                    >
                      Edit Limits
                    </button>
                  ) : (
                    <button
                      id="gov-cancel-limits-btn"
                      onClick={() => setIsEditingLimits(false)}
                      className="text-xs font-medium text-slate-500 hover:text-slate-700 transition-colors"
                    >
                      Cancel
                    </button>
                  )}
                </div>

                <form onSubmit={handleSaveLimits} className="p-5">
                  <div className="grid grid-cols-1 md:grid-cols-3 gap-5">
                    <div>
                      <label className="text-xs font-medium text-slate-700 block mb-1">
                        Max Tenant Autonomy Tier
                      </label>
                      <select
                        id="gov-limit-max-autonomy"
                        disabled={!isEditingLimits}
                        value={limitsForm.max_tenant_autonomy}
                        onChange={(e) =>
                          setLimitsForm({ ...limitsForm, max_tenant_autonomy: parseInt(e.target.value, 10) })
                        }
                        className="w-full text-xs p-2 rounded-lg border border-slate-300 bg-white disabled:bg-slate-100 text-slate-800"
                      >
                        <option value={0}>Level 0 — Observe Only</option>
                        <option value={1}>Level 1 — Recommendations Only</option>
                        <option value={2}>Level 2 — Plan Preparation (Draft)</option>
                        <option value={3}>Level 3 — Controlled Execution (Default)</option>
                        <option value={4}>Level 4 — Controlled Multi-Step Workflows</option>
                      </select>
                      <span className="text-[11px] text-slate-500 mt-1 block">
                        Hard ceiling for any module in this tenant
                      </span>
                    </div>

                    <div>
                      <label className="text-xs font-medium text-slate-700 block mb-1">
                        Max Actions / Hour
                      </label>
                      <input
                        id="gov-limit-max-actions"
                        type="number"
                        disabled={!isEditingLimits}
                        value={limitsForm.max_actions_per_hour}
                        onChange={(e) =>
                          setLimitsForm({ ...limitsForm, max_actions_per_hour: parseInt(e.target.value, 10) || 0 })
                        }
                        className="w-full text-xs p-2 rounded-lg border border-slate-300 bg-white disabled:bg-slate-100 text-slate-800"
                      />
                      <span className="text-[11px] text-slate-500 mt-1 block">
                        Rate limit across all autonomous agents
                      </span>
                    </div>

                    <div>
                      <label className="text-xs font-medium text-slate-700 block mb-1">
                        Max Financial Exposure ($ USD)
                      </label>
                      <input
                        id="gov-limit-max-finance"
                        type="number"
                        step="0.01"
                        disabled={!isEditingLimits}
                        value={limitsForm.max_financial_exposure_per_workflow}
                        onChange={(e) =>
                          setLimitsForm({
                            ...limitsForm,
                            max_financial_exposure_per_workflow: parseFloat(e.target.value) || 0
                          })
                        }
                        className="w-full text-xs p-2 rounded-lg border border-slate-300 bg-white disabled:bg-slate-100 text-slate-800"
                      />
                      <span className="text-[11px] text-slate-500 mt-1 block">
                        Requires human approval above this threshold
                      </span>
                    </div>

                    <div>
                      <label className="text-xs font-medium text-slate-700 block mb-1">
                        Max Retries / Step
                      </label>
                      <input
                        id="gov-limit-max-retries"
                        type="number"
                        disabled={!isEditingLimits}
                        value={limitsForm.max_retries_per_step}
                        onChange={(e) =>
                          setLimitsForm({ ...limitsForm, max_retries_per_step: parseInt(e.target.value, 10) || 0 })
                        }
                        className="w-full text-xs p-2 rounded-lg border border-slate-300 bg-white disabled:bg-slate-100 text-slate-800"
                      />
                      <span className="text-[11px] text-slate-500 mt-1 block">
                        Prevents infinite loop error cascades
                      </span>
                    </div>

                    <div>
                      <label className="text-xs font-medium text-slate-700 block mb-1">
                        Max Replans / Workflow
                      </label>
                      <input
                        id="gov-limit-max-replans"
                        type="number"
                        disabled={!isEditingLimits}
                        value={limitsForm.max_replans_per_plan}
                        onChange={(e) =>
                          setLimitsForm({ ...limitsForm, max_replans_per_plan: parseInt(e.target.value, 10) || 0 })
                        }
                        className="w-full text-xs p-2 rounded-lg border border-slate-300 bg-white disabled:bg-slate-100 text-slate-800"
                      />
                      <span className="text-[11px] text-slate-500 mt-1 block">
                        Escalates to human on excessive replanning
                      </span>
                    </div>

                    <div className="flex flex-col justify-center">
                      <label className="text-xs font-medium text-slate-700 mb-2 block">
                        Four-Eyes Control Policy
                      </label>
                      <label className="flex items-center gap-2 cursor-pointer">
                        <input
                          id="gov-limit-four-eyes"
                          type="checkbox"
                          disabled={!isEditingLimits}
                          checked={limitsForm.enforce_four_eyes}
                          onChange={(e) => setLimitsForm({ ...limitsForm, enforce_four_eyes: e.target.checked })}
                          className="w-4 h-4 rounded text-blue-600 focus:ring-blue-500"
                        />
                        <span className="text-xs text-slate-800 font-medium">
                          Enforce dual authorization (Preparer ≠ Approver)
                        </span>
                      </label>
                      <span className="text-[11px] text-slate-500 mt-1 block">
                        Strict separation on high-impact financial & commercial mutations
                      </span>
                    </div>
                  </div>

                  {isEditingLimits && (
                    <div className="mt-5 pt-4 border-t border-slate-200 flex justify-end gap-3">
                      <button
                        type="button"
                        onClick={() => setIsEditingLimits(false)}
                        className="px-3 py-1.5 rounded-lg border border-slate-300 text-xs font-medium text-slate-700 hover:bg-slate-100"
                      >
                        Cancel
                      </button>
                      <button
                        id="gov-save-limits-btn"
                        type="submit"
                        disabled={actionLoading}
                        className="px-4 py-1.5 rounded-lg bg-blue-600 hover:bg-blue-700 text-xs font-semibold text-white shadow-xs"
                      >
                        {actionLoading ? 'Saving...' : 'Save Governance Limits'}
                      </button>
                    </div>
                  )}
                </form>
              </div>
            </div>
          )}

          {/* TAB 2: ACTION ALLOWLIST */}
          {activeTab === 'allowlist' && (
            <div className="space-y-4">
              <div className="flex flex-col md:flex-row md:items-center justify-between gap-3 bg-white p-3.5 rounded-xl border border-slate-200">
                <div className="flex items-center gap-2 flex-1">
                  <div className="relative flex-1 max-w-sm">
                    <Search className="w-4 h-4 absolute left-3 top-2.5 text-slate-400" />
                    <input
                      id="gov-allowlist-search"
                      type="text"
                      placeholder="Search action types or names..."
                      value={allowlistSearch}
                      onChange={(e) => setAllowlistSearch(e.target.value)}
                      className="w-full text-xs pl-9 pr-3 py-1.5 rounded-lg border border-slate-300 focus:outline-hidden focus:border-blue-500"
                    />
                  </div>

                  <select
                    id="gov-allowlist-module-filter"
                    value={allowlistModuleFilter}
                    onChange={(e) => setAllowlistModuleFilter(e.target.value)}
                    className="text-xs p-1.5 rounded-lg border border-slate-300 bg-white"
                  >
                    <option value="ALL">All Modules</option>
                    <option value="shipments">Shipments</option>
                    <option value="finance">Finance</option>
                    <option value="pricing">Pricing / RFQ</option>
                    <option value="compliance">Compliance</option>
                    <option value="exceptions">Exceptions</option>
                    <option value="customer_followup">Customer Followup</option>
                    <option value="multistep">Multi-Step</option>
                  </select>
                </div>

                <div className="text-xs text-slate-500">
                  Showing <span className="font-semibold text-slate-800">{filteredAllowlist.length}</span> allowlisted
                  actions
                </div>
              </div>

              {/* Allowlist Table */}
              <div className="bg-white rounded-xl border border-slate-200 shadow-xs overflow-hidden">
                <div className="overflow-x-auto">
                  <table className="w-full text-left text-xs">
                    <thead className="bg-slate-50 border-b border-slate-200 text-slate-600 font-semibold uppercase tracking-wider">
                      <tr>
                        <th className="py-2.5 px-4">Action Type / Name</th>
                        <th className="py-2.5 px-3">Module</th>
                        <th className="py-2.5 px-3">Risk Class</th>
                        <th className="py-2.5 px-3">Reversibility</th>
                        <th className="py-2.5 px-3">Approval Gate</th>
                        <th className="py-2.5 px-3">Required Permission</th>
                        <th className="py-2.5 px-3">Max Limit</th>
                        <th className="py-2.5 px-3">Status</th>
                      </tr>
                    </thead>
                    <tbody className="divide-y divide-slate-100 text-slate-700">
                      {filteredAllowlist.length === 0 ? (
                        <tr>
                          <td colSpan={8} className="py-8 text-center text-slate-400">
                            No actions match current filters.
                          </td>
                        </tr>
                      ) : (
                        filteredAllowlist.map((item) => (
                          <tr key={item.id} className="hover:bg-slate-50/80 transition-colors">
                            <td className="py-2.5 px-4">
                              <span className="font-mono font-medium text-blue-900 block">{item.action_type}</span>
                              <span className="text-[11px] text-slate-500 block">{item.action_name}</span>
                            </td>
                            <td className="py-2.5 px-3 capitalize">{item.module}</td>
                            <td className="py-2.5 px-3">
                              <span
                                className={`inline-block px-2 py-0.5 rounded-full text-[10px] font-semibold ${
                                  item.risk_class === 'LOW'
                                    ? 'bg-emerald-50 text-emerald-700 border border-emerald-200'
                                    : item.risk_class === 'MEDIUM'
                                    ? 'bg-blue-50 text-blue-700 border border-blue-200'
                                    : item.risk_class === 'HIGH'
                                    ? 'bg-amber-50 text-amber-700 border border-amber-200'
                                    : 'bg-red-50 text-red-700 border border-red-200'
                                }`}
                              >
                                {item.risk_class}
                              </span>
                            </td>
                            <td className="py-2.5 px-3">
                              <span
                                className={`text-[10px] font-medium ${
                                  item.reversibility === 'REVERSIBLE'
                                    ? 'text-emerald-700'
                                    : item.reversibility === 'PARTIALLY_REVERSIBLE'
                                    ? 'text-amber-700'
                                    : 'text-rose-700 font-semibold'
                                }`}
                              >
                                {item.reversibility}
                              </span>
                            </td>
                            <td className="py-2.5 px-3 text-[11px]">
                              {item.approval_requirement === 'ALWAYS' ? (
                                <span className="text-rose-700 font-medium flex items-center gap-1">
                                  <Lock className="w-3 h-3" /> Mandatory
                                </span>
                              ) : item.approval_requirement === 'CONDITIONAL_RISK_THRESHOLD' ? (
                                <span className="text-amber-700 font-medium">Conditional</span>
                              ) : (
                                <span className="text-slate-500">Standard</span>
                              )}
                            </td>
                            <td className="py-2.5 px-3 font-mono text-[11px] text-slate-600">
                              {item.required_permission || '—'}
                            </td>
                            <td className="py-2.5 px-3 font-mono text-[11px]">
                              {item.max_financial_limit > 0 ? `$${item.max_financial_limit.toLocaleString()}` : '—'}
                            </td>
                            <td className="py-2.5 px-3">
                              <span
                                className={`inline-flex items-center gap-1 px-1.5 py-0.5 rounded text-[10px] font-medium ${
                                  item.is_enabled ? 'text-emerald-700 bg-emerald-50' : 'text-slate-400 bg-slate-100'
                                }`}
                              >
                                {item.is_enabled ? 'Active' : 'Disabled'}
                              </span>
                            </td>
                          </tr>
                        ))
                      )}
                    </tbody>
                  </table>
                </div>
              </div>
            </div>
          )}

          {/* TAB 3: FEATURE FLAGS */}
          {activeTab === 'flags' && (
            <div className="space-y-4">
              <div className="bg-white p-4 rounded-xl border border-slate-200">
                <h3 className="text-xs font-semibold text-slate-800 uppercase tracking-wider mb-1">
                  Module-Specific Autonomous Capability Flags
                </h3>
                <p className="text-xs text-slate-500">
                  Granularly disable autonomy or constrain maximum permitted autonomy levels per domain.
                </p>
              </div>

              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                {flags.map((flag) => (
                  <div
                    key={flag.id}
                    className="bg-white rounded-xl border border-slate-200 p-4 shadow-xs hover:border-slate-300 transition-colors"
                  >
                    <div className="flex items-start justify-between gap-3">
                      <div>
                        <h4 className="text-sm font-semibold text-slate-900 flex items-center gap-2">
                          {flag.flag_name}
                          <span
                            className={`text-[10px] px-2 py-0.5 rounded-full font-mono ${
                              flag.is_enabled
                                ? 'bg-emerald-50 text-emerald-700 border border-emerald-200'
                                : 'bg-red-50 text-red-700 border border-red-200'
                            }`}
                          >
                            {flag.is_enabled ? 'ENABLED' : 'DISABLED'}
                          </span>
                        </h4>
                        <span className="font-mono text-[11px] text-slate-400 block mt-0.5">{flag.flag_key}</span>
                        <p className="text-xs text-slate-600 mt-2 leading-relaxed">{flag.description}</p>
                      </div>

                      <button
                        id={`gov-flag-toggle-${flag.flag_key}`}
                        onClick={() =>
                          handleToggleFlag(
                            flag.flag_key,
                            flag.is_enabled,
                            flag.max_autonomy_level,
                            flag.requires_approval
                          )
                        }
                        disabled={actionLoading}
                        className={`px-3 py-1 rounded-lg text-xs font-semibold transition-all ${
                          flag.is_enabled
                            ? 'bg-slate-100 hover:bg-slate-200 text-slate-700 border border-slate-300'
                            : 'bg-emerald-600 hover:bg-emerald-700 text-white'
                        }`}
                      >
                        {flag.is_enabled ? 'Disable' : 'Enable'}
                      </button>
                    </div>

                    <div className="mt-4 pt-3 border-t border-slate-100 flex items-center justify-between text-xs">
                      <div className="flex items-center gap-2">
                        <span className="text-slate-500">Max Autonomy Tier:</span>
                        <select
                          id={`gov-flag-tier-${flag.flag_key}`}
                          value={flag.max_autonomy_level}
                          onChange={(e) =>
                            handleUpdateFlagAutonomy(
                              flag.flag_key,
                              flag.is_enabled,
                              e.target.value,
                              flag.requires_approval
                            )
                          }
                          disabled={!flag.is_enabled || actionLoading}
                          className="text-xs p-1 rounded border border-slate-300 bg-white font-medium"
                        >
                          <option value={0}>L0 (Observe)</option>
                          <option value={1}>L1 (Recommend)</option>
                          <option value={2}>L2 (Prepare)</option>
                          <option value={3}>L3 (Execute Low-Risk)</option>
                          <option value={4}>L4 (Multi-Step)</option>
                        </select>
                      </div>

                      <div className="flex items-center gap-1.5">
                        <span className="text-slate-500">Strict Human Approval:</span>
                        <span className="font-semibold text-slate-800">
                          {flag.requires_approval ? 'Required' : 'Conditional'}
                        </span>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* TAB 4: POLICY MATRIX SIMULATOR */}
          {activeTab === 'matrix' && (
            <div className="grid grid-cols-1 lg:grid-cols-12 gap-6">
              {/* Simulator Form */}
              <div className="lg:col-span-6 bg-white p-5 rounded-xl border border-slate-200 shadow-xs space-y-4">
                <div className="border-b border-slate-100 pb-3">
                  <h3 className="text-sm font-semibold text-slate-900 flex items-center gap-2">
                    <Play className="w-4 h-4 text-blue-600" />
                    Interactive Policy Decision Matrix
                  </h3>
                  <p className="text-xs text-slate-500 mt-0.5">
                    Test action authorization against live kill switches, flags, allowlists, four-eyes, and safety invariants.
                  </p>
                </div>

                <form onSubmit={handleRunSimulation} className="space-y-3.5">
                  <div className="grid grid-cols-2 gap-3">
                    <div>
                      <label className="text-[11px] font-medium text-slate-700 block mb-1">Module</label>
                      <select
                        id="gov-sim-module"
                        value={simForm.module}
                        onChange={(e) => setSimForm({ ...simForm, module: e.target.value })}
                        className="w-full text-xs p-2 rounded-lg border border-slate-300 bg-white text-slate-800"
                      >
                        <option value="shipments">Shipments</option>
                        <option value="finance">Finance</option>
                        <option value="pricing">Pricing / RFQ</option>
                        <option value="compliance">Compliance</option>
                        <option value="exceptions">Exceptions</option>
                        <option value="customer_followup">Customer Followup</option>
                      </select>
                    </div>

                    <div>
                      <label className="text-[11px] font-medium text-slate-700 block mb-1">Action Type</label>
                      <input
                        id="gov-sim-action-type"
                        type="text"
                        value={simForm.action_type}
                        onChange={(e) => setSimForm({ ...simForm, action_type: e.target.value })}
                        placeholder="e.g. carrier.escalate"
                        className="w-full text-xs p-2 rounded-lg border border-slate-300 font-mono text-slate-800"
                        required
                      />
                    </div>
                  </div>

                  <div className="grid grid-cols-3 gap-3">
                    <div>
                      <label className="text-[11px] font-medium text-slate-700 block mb-1">User Role</label>
                      <select
                        id="gov-sim-user-role"
                        value={simForm.user_role}
                        onChange={(e) => setSimForm({ ...simForm, user_role: e.target.value })}
                        className="w-full text-xs p-2 rounded-lg border border-slate-300 bg-white text-slate-800"
                      >
                        <option value="operations">Operations Specialist</option>
                        <option value="finance_manager">Finance Manager</option>
                        <option value="compliance_manager">Compliance Manager</option>
                        <option value="commercial_director">Commercial Director</option>
                        <option value="admin">System Administrator</option>
                      </select>
                    </div>

                    <div>
                      <label className="text-[11px] font-medium text-slate-700 block mb-1">Requested Autonomy</label>
                      <select
                        id="gov-sim-requested-autonomy"
                        value={simForm.requested_autonomy}
                        onChange={(e) => setSimForm({ ...simForm, requested_autonomy: e.target.value })}
                        className="w-full text-xs p-2 rounded-lg border border-slate-300 bg-white text-slate-800"
                      >
                        <option value={0}>L0 Observe</option>
                        <option value={1}>L1 Recommend</option>
                        <option value={2}>L2 Prepare</option>
                        <option value={3}>L3 Controlled Exec</option>
                        <option value={4}>L4 Multi-Step</option>
                      </select>
                    </div>

                    <div>
                      <label className="text-[11px] font-medium text-slate-700 block mb-1">Financial Impact ($)</label>
                      <input
                        id="gov-sim-financial-amount"
                        type="number"
                        value={simForm.financial_amount}
                        onChange={(e) => setSimForm({ ...simForm, financial_amount: e.target.value })}
                        className="w-full text-xs p-2 rounded-lg border border-slate-300 text-slate-800"
                      />
                    </div>
                  </div>

                  <div className="grid grid-cols-2 gap-3">
                    <div>
                      <label className="text-[11px] font-medium text-slate-700 block mb-1">User Permissions (comma-separated)</label>
                      <input
                        id="gov-sim-permissions"
                        type="text"
                        value={simForm.user_permissions}
                        onChange={(e) => setSimForm({ ...simForm, user_permissions: e.target.value })}
                        placeholder="shipments:update, finance:apply_discount"
                        className="w-full text-xs p-2 rounded-lg border border-slate-300 font-mono text-slate-800"
                      />
                    </div>

                    <div>
                      <label className="text-[11px] font-medium text-slate-700 block mb-1">Compliance Status</label>
                      <select
                        id="gov-sim-compliance-status"
                        value={simForm.compliance_status}
                        onChange={(e) => setSimForm({ ...simForm, compliance_status: e.target.value })}
                        className="w-full text-xs p-2 rounded-lg border border-slate-300 bg-white text-slate-800"
                      >
                        <option value="COMPLIANT">COMPLIANT (Normal)</option>
                        <option value="WARNING">WARNING (Review)</option>
                        <option value="BLOCKED">BLOCKED (Prohibited)</option>
                      </select>
                    </div>
                  </div>

                  {/* Four-Eyes and Human Approval Toggles */}
                  <div className="p-3 bg-slate-50 rounded-lg border border-slate-200 space-y-2.5">
                    <div className="flex items-center justify-between">
                      <span className="text-xs font-medium text-slate-800">Four-Eyes Dual Control Simulation</span>
                      <label className="flex items-center gap-2 cursor-pointer">
                        <input
                          id="gov-sim-is-approved"
                          type="checkbox"
                          checked={simForm.is_approved}
                          onChange={(e) => setSimForm({ ...simForm, is_approved: e.target.checked })}
                          className="w-3.5 h-3.5 rounded text-blue-600"
                        />
                        <span className="text-xs text-slate-600">Action is Approved</span>
                      </label>
                    </div>

                    <div className="grid grid-cols-2 gap-3">
                      <div>
                        <span className="text-[10px] text-slate-500 block">Preparer User ID</span>
                        <input
                          id="gov-sim-preparer-id"
                          type="number"
                          value={simForm.preparer_user_id}
                          onChange={(e) => setSimForm({ ...simForm, preparer_user_id: e.target.value })}
                          className="w-full text-xs p-1.5 rounded border border-slate-300 bg-white text-slate-800"
                        />
                      </div>
                      <div>
                        <span className="text-[10px] text-slate-500 block">Approver User ID</span>
                        <input
                          id="gov-sim-approver-id"
                          type="number"
                          value={simForm.approval_approver_id}
                          onChange={(e) => setSimForm({ ...simForm, approval_approver_id: e.target.value })}
                          className="w-full text-xs p-1.5 rounded border border-slate-300 bg-white text-slate-800"
                        />
                      </div>
                    </div>
                  </div>

                  <div>
                    <label className="text-[11px] font-medium text-slate-700 block mb-1">Operational Context / Reason</label>
                    <textarea
                      id="gov-sim-context"
                      rows={2}
                      value={simForm.context_text}
                      onChange={(e) => setSimForm({ ...simForm, context_text: e.target.value })}
                      placeholder="Enter context or test prompt injection resistance..."
                      className="w-full text-xs p-2 rounded-lg border border-slate-300 text-slate-800 focus:outline-hidden focus:border-blue-500"
                    />
                  </div>

                  <button
                    id="gov-eval-simulate-btn"
                    type="submit"
                    disabled={simLoading}
                    className="w-full py-2.5 rounded-lg bg-[#0B192C] hover:bg-slate-800 text-white font-semibold text-xs flex items-center justify-center gap-2 shadow-xs transition-colors"
                  >
                    {simLoading ? (
                      <>
                        <RefreshCw className="w-4 h-4 animate-spin" />
                        Evaluating Policy...
                      </>
                    ) : (
                      <>
                        <Play className="w-4 h-4 text-blue-400" />
                        Evaluate Policy Decision
                      </>
                    )}
                  </button>
                </form>
              </div>

              {/* Simulator Results Panel */}
              <div className="lg:col-span-6 space-y-4">
                {simResult ? (
                  <div
                    id="gov-sim-result-panel"
                    className="bg-white p-5 rounded-xl border border-slate-200 shadow-xs space-y-4 animate-in fade-in duration-150"
                  >
                    {/* Primary Verdict Banner */}
                    <div
                      className={`p-4 rounded-xl border flex items-center gap-3.5 ${
                        simResult.decision === 'ALLOW'
                          ? 'bg-emerald-50 border-emerald-200 text-emerald-950'
                          : simResult.decision === 'REQUIRE_REVIEW'
                          ? 'bg-amber-50 border-amber-200 text-amber-950'
                          : 'bg-red-50 border-red-200 text-red-950'
                      }`}
                    >
                      {simResult.decision === 'ALLOW' ? (
                        <CheckCircle2 className="w-8 h-8 text-emerald-600 shrink-0" />
                      ) : simResult.decision === 'REQUIRE_REVIEW' ? (
                        <AlertTriangle className="w-8 h-8 text-amber-600 shrink-0" />
                      ) : (
                        <AlertOctagon className="w-8 h-8 text-red-600 shrink-0" />
                      )}

                      <div>
                        <div className="flex items-center gap-2">
                          <span className="text-xs uppercase font-mono font-semibold tracking-wider text-slate-500">
                            Authoritative Verdict
                          </span>
                          <span
                            className={`text-xs px-2 py-0.5 rounded font-bold ${
                              simResult.decision === 'ALLOW'
                                ? 'bg-emerald-600 text-white'
                                : simResult.decision === 'REQUIRE_REVIEW'
                                ? 'bg-amber-600 text-white'
                                : 'bg-red-600 text-white'
                            }`}
                          >
                            {simResult.decision}
                          </span>
                        </div>
                        <span className="text-sm font-semibold block mt-0.5">
                          {simResult.decision === 'ALLOW'
                            ? 'Action permitted for immediate autonomous execution'
                            : simResult.decision === 'REQUIRE_REVIEW'
                            ? 'Action requires human approval before execution'
                            : 'Action strictly prohibited under current governance policy'}
                        </span>
                      </div>
                    </div>

                    {/* Decision Factors Grid */}
                    <div className="grid grid-cols-2 md:grid-cols-3 gap-3 text-xs">
                      <div className="bg-slate-50 p-2.5 rounded-lg border border-slate-200">
                        <span className="text-[10px] text-slate-500 block">Effective Autonomy</span>
                        <span className="font-bold text-slate-800 text-sm">
                          Level {simResult.effective_autonomy}
                        </span>
                      </div>

                      <div className="bg-slate-50 p-2.5 rounded-lg border border-slate-200">
                        <span className="text-[10px] text-slate-500 block">Risk Class</span>
                        <span className="font-bold text-slate-800 text-sm">{simResult.risk_class}</span>
                      </div>

                      <div className="bg-slate-50 p-2.5 rounded-lg border border-slate-200">
                        <span className="text-[10px] text-slate-500 block">Data Sufficiency</span>
                        <span className="font-bold text-slate-800 text-sm">
                          {simResult.data_sufficiency}
                        </span>
                      </div>

                      <div className="bg-slate-50 p-2.5 rounded-lg border border-slate-200">
                        <span className="text-[10px] text-slate-500 block">Reversibility</span>
                        <span className="font-bold text-slate-800 text-sm">{simResult.reversibility}</span>
                      </div>

                      <div className="bg-slate-50 p-2.5 rounded-lg border border-slate-200">
                        <span className="text-[10px] text-slate-500 block">Four-Eyes Enforced</span>
                        <span className="font-bold text-slate-800 text-sm">
                          {simResult.four_eyes_enforced ? 'Yes' : 'No'}
                        </span>
                      </div>

                      <div className="bg-slate-50 p-2.5 rounded-lg border border-slate-200">
                        <span className="text-[10px] text-slate-500 block">Approval Required</span>
                        <span className="font-bold text-slate-800 text-sm">
                          {simResult.requires_approval ? 'Yes' : 'No'}
                        </span>
                      </div>
                    </div>

                    {/* Reasons & Explanations */}
                    <div>
                      <h4 className="text-xs font-semibold text-slate-800 uppercase tracking-wider mb-2">
                        Evaluation Reasons & Enforced Rules
                      </h4>
                      <div className="space-y-1.5">
                        {simResult.reasons && simResult.reasons.length > 0 ? (
                          simResult.reasons.map((reason, idx) => (
                            <div
                              key={idx}
                              className="text-xs p-2.5 rounded-lg bg-slate-50 border border-slate-200 text-slate-800 flex items-start gap-2"
                            >
                              <ChevronRight className="w-3.5 h-3.5 text-blue-500 mt-0.5 shrink-0" />
                              <span>{reason}</span>
                            </div>
                          ))
                        ) : (
                          <div className="text-xs text-slate-500 italic">No specific conditions triggered.</div>
                        )}
                      </div>
                    </div>

                    <div className="text-[11px] font-mono text-slate-400 text-right">
                      Correlation ID: {simResult.correlation_id}
                    </div>
                  </div>
                ) : (
                  <div className="bg-white p-8 rounded-xl border border-slate-200 text-center shadow-xs flex flex-col items-center justify-center min-h-[300px]">
                    <Scale className="w-12 h-12 text-slate-300 mb-3" />
                    <h4 className="text-sm font-medium text-slate-700">Awaiting Policy Evaluation</h4>
                    <p className="text-xs text-slate-500 max-w-sm mt-1">
                      Configure operational parameters on the left and click "Evaluate Policy Decision" to run the central Go policy verification engine.
                    </p>
                  </div>
                )}
              </div>
            </div>
          )}

          {/* TAB 5: AUDIT & TELEMETRY */}
          {activeTab === 'audit' && (
            <div className="space-y-6">
              {/* Policy Evaluations Table */}
              <div className="bg-white rounded-xl border border-slate-200 shadow-xs overflow-hidden">
                <div className="px-5 py-3.5 border-b border-slate-200 flex items-center justify-between bg-slate-50">
                  <div className="flex items-center gap-2">
                    <Activity className="w-4 h-4 text-blue-600" />
                    <h4 className="text-xs font-semibold text-slate-800 uppercase tracking-wider">
                      Recent Policy Evaluations
                    </h4>
                  </div>

                  <select
                    id="gov-eval-decision-filter"
                    value={evalDecisionFilter}
                    onChange={(e) => setEvalDecisionFilter(e.target.value)}
                    className="text-xs p-1.5 rounded-lg border border-slate-300 bg-white"
                  >
                    <option value="ALL">All Decisions</option>
                    <option value="ALLOW">ALLOW</option>
                    <option value="REQUIRE_REVIEW">REQUIRE_REVIEW</option>
                    <option value="BLOCK">BLOCK</option>
                  </select>
                </div>

                <div className="overflow-x-auto">
                  <table className="w-full text-left text-xs">
                    <thead className="bg-slate-50 border-b border-slate-200 text-slate-600 font-semibold uppercase tracking-wider">
                      <tr>
                        <th className="py-2.5 px-4">Timestamp</th>
                        <th className="py-2.5 px-3">Entity</th>
                        <th className="py-2.5 px-3">Action Type</th>
                        <th className="py-2.5 px-3">Autonomy (Req / Eff)</th>
                        <th className="py-2.5 px-3">Decision</th>
                        <th className="py-2.5 px-3">Risk</th>
                      </tr>
                    </thead>
                    <tbody className="divide-y divide-slate-100 text-slate-700">
                      {filteredEvaluations.length === 0 ? (
                        <tr>
                          <td colSpan={6} className="py-8 text-center text-slate-400">
                            No evaluations recorded yet.
                          </td>
                        </tr>
                      ) : (
                        filteredEvaluations.map((ev) => (
                          <tr key={ev.id} className="hover:bg-slate-50 transition-colors">
                            <td className="py-2.5 px-4 font-mono text-[11px] text-slate-500 whitespace-nowrap">
                              {new Date(ev.created_at).toLocaleTimeString([], {
                                hour: '2-digit',
                                minute: '2-digit',
                                second: '2-digit'
                              })}
                            </td>
                            <td className="py-2.5 px-3">
                              <span className="font-mono text-slate-800 font-medium">{ev.entity_type}</span>
                              <span className="text-[11px] text-slate-400 block">{ev.entity_id}</span>
                            </td>
                            <td className="py-2.5 px-3 font-mono text-blue-900 font-medium">{ev.action_type}</td>
                            <td className="py-2.5 px-3 font-mono text-xs">
                              L{ev.requested_autonomy} → <span className="font-bold text-blue-700">L{ev.effective_autonomy}</span>
                            </td>
                            <td className="py-2.5 px-3">
                              <span
                                className={`inline-block px-2 py-0.5 rounded text-[10px] font-bold ${
                                  ev.decision === 'ALLOW'
                                    ? 'bg-emerald-100 text-emerald-800'
                                    : ev.decision === 'REQUIRE_REVIEW'
                                    ? 'bg-amber-100 text-amber-800'
                                    : 'bg-red-100 text-red-800'
                                }`}
                              >
                                {ev.decision}
                              </span>
                            </td>
                            <td className="py-2.5 px-3 text-[11px]">{ev.risk_level}</td>
                          </tr>
                        ))
                      )}
                    </tbody>
                  </table>
                </div>
              </div>

              {/* Administrative Policy Audit Logs */}
              <div className="bg-white rounded-xl border border-slate-200 shadow-xs overflow-hidden">
                <div className="px-5 py-3.5 border-b border-slate-200 bg-slate-50 flex items-center gap-2">
                  <History className="w-4 h-4 text-blue-600" />
                  <h4 className="text-xs font-semibold text-slate-800 uppercase tracking-wider">
                    Administrative Configuration Audit Trail
                  </h4>
                </div>

                <div className="overflow-x-auto">
                  <table className="w-full text-left text-xs">
                    <thead className="bg-slate-50 border-b border-slate-200 text-slate-600 font-semibold uppercase tracking-wider">
                      <tr>
                        <th className="py-2.5 px-4">Timestamp</th>
                        <th className="py-2.5 px-3">User ID</th>
                        <th className="py-2.5 px-3">Change Type</th>
                        <th className="py-2.5 px-3">Target</th>
                        <th className="py-2.5 px-3">Reason / Justification</th>
                      </tr>
                    </thead>
                    <tbody className="divide-y divide-slate-100 text-slate-700">
                      {auditLogs.length === 0 ? (
                        <tr>
                          <td colSpan={5} className="py-8 text-center text-slate-400">
                            No administrative audit records found.
                          </td>
                        </tr>
                      ) : (
                        auditLogs.map((log) => (
                          <tr key={log.id} className="hover:bg-slate-50 transition-colors">
                            <td className="py-2.5 px-4 font-mono text-[11px] text-slate-500 whitespace-nowrap">
                              {new Date(log.created_at).toLocaleString()}
                            </td>
                            <td className="py-2.5 px-3 font-mono">User #{log.user_id}</td>
                            <td className="py-2.5 px-3 font-semibold text-slate-800">{log.change_type}</td>
                            <td className="py-2.5 px-3 font-mono text-[11px] text-slate-600">
                              {log.target_type}: {log.target_id}
                            </td>
                            <td className="py-2.5 px-3 text-slate-600">{log.reason || '—'}</td>
                          </tr>
                        ))
                      )}
                    </tbody>
                  </table>
                </div>
              </div>
            </div>
          )}
        </div>

        {/* Footer */}
        <div className="bg-slate-100 px-6 py-3 border-t border-slate-200 flex items-center justify-between text-xs text-slate-500">
          <div className="flex items-center gap-2">
            <Shield className="w-4 h-4 text-emerald-600" />
            <span>Go Centralized Enforcement: Deterministic & Isolated</span>
          </div>
          <button
            onClick={onClose}
            className="px-4 py-1.5 rounded-lg bg-white border border-slate-300 text-slate-700 hover:bg-slate-200 font-medium"
          >
            Close
          </button>
        </div>
      </div>

      {/* Kill Switch Confirmation Modal */}
      {killSwitchModalOpen && (
        <div className="fixed inset-0 z-60 bg-slate-900/60 backdrop-blur-xs flex items-center justify-center p-4">
          <div className="bg-white rounded-xl max-w-md w-full p-6 shadow-2xl border border-slate-200 text-slate-800 space-y-4">
            <div className="flex items-center gap-3">
              <div
                className={`w-10 h-10 rounded-lg flex items-center justify-center shrink-0 ${
                  limits?.kill_switch_active ? 'bg-emerald-100 text-emerald-700' : 'bg-red-100 text-red-700'
                }`}
              >
                <AlertOctagon className="w-6 h-6" />
              </div>
              <div>
                <h3 className="text-base font-semibold">
                  {limits?.kill_switch_active ? 'Deactivate Emergency Stop' : 'Activate Emergency Kill Switch'}
                </h3>
                <p className="text-xs text-slate-500 mt-0.5">
                  {limits?.kill_switch_active
                    ? 'Resume normal controlled autonomous execution across this tenant.'
                    : 'Immediately freeze and block all autonomous execution.'}
                </p>
              </div>
            </div>

            <div>
              <label className="text-xs font-medium text-slate-700 block mb-1">
                Reason / Administrative Justification
              </label>
              <textarea
                id="gov-kill-switch-reason"
                rows={3}
                value={killSwitchReason}
                onChange={(e) => setKillSwitchReason(e.target.value)}
                placeholder="Provide a mandatory reason for this emergency intervention..."
                className="w-full text-xs p-2 rounded-lg border border-slate-300 text-slate-800 focus:outline-hidden focus:border-red-500"
              />
            </div>

            <div className="flex justify-end gap-3 pt-2">
              <button
                type="button"
                onClick={() => {
                  setKillSwitchModalOpen(false);
                  setKillSwitchReason('');
                }}
                className="px-3.5 py-2 rounded-lg border border-slate-300 text-xs font-medium text-slate-700 hover:bg-slate-100"
              >
                Cancel
              </button>
              <button
                id="gov-confirm-kill-switch-btn"
                type="button"
                onClick={handleToggleKillSwitch}
                disabled={actionLoading}
                className={`px-4 py-2 rounded-lg text-xs font-semibold text-white shadow-xs ${
                  limits?.kill_switch_active
                    ? 'bg-emerald-600 hover:bg-emerald-700'
                    : 'bg-red-600 hover:bg-red-700'
                }`}
              >
                {actionLoading
                  ? 'Processing...'
                  : limits?.kill_switch_active
                  ? 'Confirm Deactivation'
                  : 'CONFIRM EMERGENCY STOP'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

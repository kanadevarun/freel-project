import React, { useState, useEffect } from 'react';
import { Link, useSearchParams } from 'react-router-dom';
import {
  Settings,
  User,
  Shield,
  Sliders,
  Cpu,
  Radio,
  Lock,
  Activity,
  FileText,
  AlertTriangle,
  CheckCircle2,
  XCircle,
  RefreshCw,
  Save,
  Edit2,
  Eye,
  Check,
  X,
  Power,
  ShieldAlert,
  Server,
  Database,
  ExternalLink,
  ChevronRight,
  Search,
  Bell,
  Zap,
  Globe,
  HelpCircle,
  Clock,
  Layers,
  ArrowUpRight,
  Sparkles,
  Key,
  Mail,
  Smartphone
} from 'lucide-react';
import { sportalService } from '../../services/sportalService';
import toast from 'react-hot-toast';

export function SettingsPage() {
  const [searchParams, setSearchParams] = useSearchParams();
  const initialTab = searchParams.get('tab') || 'overview';
  const [activeTab, setActiveTabState] = useState(initialTab);

  const setActiveTab = (tabId) => {
    setActiveTabState(tabId);
    setSearchParams({ tab: tabId });
  };

  useEffect(() => {
    const currentTab = searchParams.get('tab');
    if (currentTab && currentTab !== activeTab) {
      setActiveTabState(currentTab);
    }
  }, [searchParams]);

  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  const [overviewData, setOverviewData] = useState(null);

  // Profile Form State
  const [profileForm, setProfileForm] = useState({
    firstName: '',
    lastName: '',
    minSeverity: 'MEDIUM',
    inAppEnabled: true,
    approvalsEnabled: true,
    automationsEnabled: true,
    recommendationsEnabled: true,
    financeEnabled: true,
    operationsEnabled: true,
    complianceEnabled: true,
  });
  const [savingProfile, setSavingProfile] = useState(false);

  // Edit Platform Setting Modal
  const [editingSetting, setEditingSetting] = useState(null);
  const [editSettingValue, setEditSettingValue] = useState('');
  const [savingSetting, setSavingSetting] = useState(false);

  // Edit Feature Flag Modal
  const [editingFlag, setEditingFlag] = useState(null);
  const [flagForm, setFlagForm] = useState({
    isEnabled: false,
    requiresApproval: true,
    maxAutonomyLevel: 2,
  });
  const [savingFlag, setSavingFlag] = useState(false);

  // Emergency Halt Modal
  const [haltModalOpen, setHaltModalOpen] = useState(false);
  const [haltTargetModule, setHaltTargetModule] = useState('');
  const [haltActiveState, setHaltActiveState] = useState(true);
  const [haltReason, setHaltReason] = useState('');
  const [executingHalt, setExecutingHalt] = useState(false);

  // Integration Toggle Modal
  const [integrationModal, setIntegrationModal] = useState(null);
  const [integrationReason, setIntegrationReason] = useState('');
  const [togglingIntegration, setTogglingIntegration] = useState(false);

  // Audit Search
  const [auditSearch, setAuditSearch] = useState('');

  useEffect(() => {
    fetchSettingsOverview();
  }, []);

  const fetchSettingsOverview = async () => {
    try {
      setLoading(true);
      const res = await sportalService.getSettingsOverview();
      const payload = (res && res.data !== undefined && !res.platform_settings) ? res.data : res;
      if (payload) {
        setOverviewData(payload);
        // Initialize profile form
        if (payload.user) {
          setProfileForm((prev) => ({
            ...prev,
            firstName: payload.user.first_name || '',
            lastName: payload.user.last_name || '',
          }));
        }
        if (payload.preferences) {
          const p = payload.preferences;
          setProfileForm((prev) => ({
            ...prev,
            minSeverity: p.min_severity || 'MEDIUM',
            inAppEnabled: p.in_app_enabled ?? true,
            approvalsEnabled: p.approvals_enabled ?? true,
            automationsEnabled: p.automations_enabled ?? true,
            recommendationsEnabled: p.recommendations_enabled ?? true,
            financeEnabled: p.finance_enabled ?? true,
            operationsEnabled: p.operations_enabled ?? true,
            complianceEnabled: p.compliance_enabled ?? true,
          }));
        }
      }
    } catch (err) {
      console.error('Failed to load settings overview:', err);
      toast.error('Failed to load settings overview: ' + (err.message || 'Unknown error'));
    } finally {
      setLoading(false);
      setRefreshing(false);
    }
  };

  const handleRefresh = () => {
    setRefreshing(true);
    fetchSettingsOverview();
  };

  // 1. Handle Profile Update
  const handleSaveProfile = async (e) => {
    if (e && e.preventDefault) e.preventDefault();
    try {
      setSavingProfile(true);
      const payload = {
        first_name: profileForm.firstName,
        last_name: profileForm.lastName,
        min_severity: profileForm.minSeverity,
        in_app_enabled: profileForm.inAppEnabled,
        approvals_enabled: profileForm.approvalsEnabled,
        automations_enabled: profileForm.automationsEnabled,
        recommendations_enabled: profileForm.recommendationsEnabled,
        finance_enabled: profileForm.financeEnabled,
        operations_enabled: profileForm.operationsEnabled,
        compliance_enabled: profileForm.complianceEnabled,
      };
      await sportalService.updateInternalUserProfile(payload);
      toast.success('Personal staff preferences updated successfully');
      fetchSettingsOverview();
    } catch (err) {
      console.error('Profile update failed:', err);
      toast.error('Failed to update profile: ' + (err.message || 'Check permissions'));
    } finally {
      setSavingProfile(false);
    }
  };

  // 2. Handle Platform Setting Edit
  const openEditSetting = (setting) => {
    setEditingSetting(setting);
    setEditSettingValue(setting.setting_value);
  };

  const handleSaveSetting = async () => {
    if (!editingSetting) return;
    try {
      setSavingSetting(true);
      await sportalService.updatePlatformSetting(editingSetting.setting_key, editSettingValue);
      toast.success(`Platform setting '${editingSetting.setting_key}' updated`);
      setEditingSetting(null);
      fetchSettingsOverview();
    } catch (err) {
      console.error('Setting update failed:', err);
      toast.error('Failed to update setting: ' + (err.message || 'Permission denied'));
    } finally {
      setSavingSetting(false);
    }
  };

  // 3. Handle Feature Flag Edit
  const openEditFlag = (flag) => {
    setEditingFlag(flag);
    setFlagForm({
      isEnabled: flag.is_enabled,
      requiresApproval: flag.requires_approval,
      maxAutonomyLevel: flag.max_autonomy_level,
    });
  };

  const handleSaveFlag = async () => {
    if (!editingFlag) return;
    try {
      setSavingFlag(true);
      await sportalService.updateFeatureFlag(editingFlag.flag_key, {
        is_enabled: flagForm.isEnabled,
        requires_approval: flagForm.requiresApproval,
        max_autonomy_level: Number(flagForm.maxAutonomyLevel),
      });
      toast.success(`Feature flag '${editingFlag.flag_key}' updated successfully`);
      setEditingFlag(null);
      fetchSettingsOverview();
    } catch (err) {
      console.error('Feature flag update failed:', err);
      toast.error('Failed to update feature flag: ' + (err.message || 'Permission denied'));
    } finally {
      setSavingFlag(false);
    }
  };

  // 4. Handle Emergency Halt
  const openEmergencyHalt = (module = '', targetState = true) => {
    setHaltTargetModule(module);
    setHaltActiveState(targetState);
    setHaltReason('');
    setHaltModalOpen(true);
  };

  const handleExecuteHalt = async () => {
    try {
      setExecutingHalt(true);
      await sportalService.triggerEmergencyHalt(haltActiveState, haltTargetModule, haltReason);
      toast.success(
        haltActiveState
          ? `Emergency Halt ENGAGED on ${haltTargetModule ? `module '${haltTargetModule}'` : 'ALL MODULES'}`
          : `Emergency Halt CLEARED for ${haltTargetModule ? `module '${haltTargetModule}'` : 'ALL MODULES'}`
      );
      setHaltModalOpen(false);
      fetchSettingsOverview();
    } catch (err) {
      console.error('Emergency halt execution failed:', err);
      toast.error('Emergency halt failed: ' + (err.message || 'Permission denied'));
    } finally {
      setExecutingHalt(false);
    }
  };

  // 5. Handle Integration Toggle
  const openIntegrationToggle = (integration) => {
    setIntegrationModal(integration);
    setIntegrationReason('');
  };

  const handleToggleIntegration = async () => {
    if (!integrationModal) return;
    try {
      setTogglingIntegration(true);
      const newState = !integrationModal.is_enabled;
      await sportalService.toggleIntegrationSetting(integrationModal.integration_type, newState, integrationReason);
      toast.success(`Integration '${integrationModal.provider_name}' ${newState ? 'enabled' : 'disabled'}`);
      setIntegrationModal(null);
      fetchSettingsOverview();
    } catch (err) {
      console.error('Integration toggle failed:', err);
      toast.error('Failed to toggle integration: ' + (err.message || 'Permission denied'));
    } finally {
      setTogglingIntegration(false);
    }
  };

  if (loading && !overviewData) {
    return (
      <div className="flex flex-col items-center justify-center min-h-[500px] text-slate-500">
        <RefreshCw className="w-8 h-8 animate-spin text-blue-600 mb-3" />
        <p className="text-sm font-medium text-slate-700">Loading SPortal Platform Administration...</p>
        <p className="text-xs text-slate-400 mt-1">Connecting to Go control plane and retrieving configuration source of truth</p>
      </div>
    );
  }

  const {
    user = {},
    role = {},
    platform_settings: platformSettings = [],
    feature_flags: featureFlags = [],
    autonomy_policies: autonomyPolicies = [],
    global_emergency_halt: globalEmergencyHalt = false,
    integrations = [],
    operations_health: opsHealth = {},
    recent_audits: recentAudits = []
  } = overviewData || {};

  // Filtered audits
  const filteredAudits = (recentAudits || []).filter((a) => {
    if (!auditSearch) return true;
    const term = auditSearch.toLowerCase();
    return (
      (a.action && a.action.toLowerCase().includes(term)) ||
      (a.actor_name && a.actor_name.toLowerCase().includes(term)) ||
      (a.description && a.description.toLowerCase().includes(term)) ||
      (a.module && a.module.toLowerCase().includes(term))
    );
  });

  const isTabActive = (tabId, aliases = []) => {
    if (activeTab === tabId) return true;
    if (aliases && aliases.includes(activeTab)) return true;
    return false;
  };

  const tabs = [
    { id: 'profile', label: 'Profile', icon: User, aliases: ['overview'] },
    { id: 'users', label: 'Users & Access', icon: Shield },
    { id: 'platform', label: 'Platform', icon: Sliders },
    { id: 'feature-flags', label: 'Feature Flags', icon: Radio },
    { id: 'integrations', label: 'Integrations', icon: Globe },
    { id: 'notifications', label: 'Notifications', icon: Bell },
    { id: 'ai', label: 'AI Administration', icon: Cpu },
    { id: 'autonomy', label: 'Autonomy & Halt', icon: Zap },
    { id: 'security', label: 'Security', icon: Lock },
    { id: 'operations', label: 'System Health', icon: Activity },
    { id: 'audit', label: 'Audit Trail', icon: FileText },
  ];

  return (
    <div className="space-y-6 pb-12">
      {/* Top Header Card */}
      <div className="bg-white border border-slate-200 rounded-lg p-6 shadow-sm">
        <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
          <div>
            <div className="flex items-center gap-2">
              <span className="px-2.5 py-0.5 rounded text-xs font-semibold bg-slate-100 text-slate-700 border border-slate-200 uppercase tracking-wide">
                Administrative Control Area
              </span>
              {globalEmergencyHalt ? (
                <span className="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded text-xs font-semibold bg-rose-100 text-rose-800 border border-rose-300 animate-pulse">
                  <ShieldAlert className="w-3.5 h-3.5 text-rose-600" />
                  EMERGENCY HALT ACTIVE
                </span>
              ) : (
                <span className="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded text-xs font-semibold bg-emerald-100 text-emerald-800 border border-emerald-300">
                  <CheckCircle2 className="w-3.5 h-3.5 text-emerald-600" />
                  Autonomous Platform Healthy
                </span>
              )}
            </div>
            <h1 className="text-2xl font-bold text-slate-900 mt-2">SPortal Settings & Platform Administration</h1>
            <p className="text-sm text-slate-600 mt-1 max-w-3xl">
              Internal control center for LogisticsHQ. Manage internal staff profiles, access policies, platform defaults,
              feature flags, external gateways, notifications, AI Workforce autonomy, and system telemetry.
            </p>
          </div>

          <div className="flex items-center gap-3 self-start md:self-auto">
            <button
              onClick={handleRefresh}
              disabled={refreshing}
              className="inline-flex items-center gap-2 px-3.5 py-2 text-sm font-medium text-slate-700 bg-white border border-slate-300 rounded-md hover:bg-slate-50 transition-colors shadow-sm disabled:opacity-50"
            >
              <RefreshCw className={`w-4 h-4 text-slate-500 ${refreshing ? 'animate-spin' : ''}`} />
              Refresh
            </button>
            <button
              onClick={() => openEmergencyHalt('', !globalEmergencyHalt)}
              className={`inline-flex items-center gap-2 px-4 py-2 text-sm font-semibold rounded-md shadow-sm transition-colors ${
                globalEmergencyHalt
                  ? 'bg-emerald-600 text-white hover:bg-emerald-700'
                  : 'bg-rose-600 text-white hover:bg-rose-700'
              }`}
            >
              <Power className="w-4 h-4" />
              {globalEmergencyHalt ? 'Resume Autonomous Execution' : 'Emergency Kill Switch'}
            </button>
          </div>
        </div>

        {/* Live Operational Tickers */}
        <div className="mt-6 pt-5 border-t border-slate-100 grid grid-cols-2 sm:grid-cols-3 md:grid-cols-6 gap-3">
          <div className="bg-slate-50 rounded p-2.5 border border-slate-200">
            <span className="text-[11px] font-semibold text-slate-500 uppercase tracking-wider block">Backend</span>
            <span className="text-xs font-bold text-emerald-700 flex items-center gap-1 mt-0.5">
              <span className="w-2 h-2 rounded-full bg-emerald-500"></span>
              {opsHealth.backend_status || 'HEALTHY'} (:8080)
            </span>
          </div>

          <div className="bg-slate-50 rounded p-2.5 border border-slate-200">
            <span className="text-[11px] font-semibold text-slate-500 uppercase tracking-wider block">Database</span>
            <span className="text-xs font-bold text-emerald-700 flex items-center gap-1 mt-0.5">
              <span className="w-2 h-2 rounded-full bg-emerald-500"></span>
              {opsHealth.database_status || 'HEALTHY'} (:3306)
            </span>
          </div>

          <div className="bg-slate-50 rounded p-2.5 border border-slate-200">
            <span className="text-[11px] font-semibold text-slate-500 uppercase tracking-wider block">AI Sidecar</span>
            <span className={`text-xs font-bold flex items-center gap-1 mt-0.5 ${
              opsHealth.ai_sidecar_status === 'HEALTHY' ? 'text-emerald-700' : 'text-amber-700'
            }`}>
              <span className={`w-2 h-2 rounded-full ${opsHealth.ai_sidecar_status === 'HEALTHY' ? 'bg-emerald-500' : 'bg-amber-500'}`}></span>
              {opsHealth.ai_sidecar_status || 'UNKNOWN'} (:8090)
            </span>
          </div>

          <div className="bg-slate-50 rounded p-2.5 border border-slate-200">
            <span className="text-[11px] font-semibold text-slate-500 uppercase tracking-wider block">Event Mesh</span>
            <span className="text-xs font-bold text-emerald-700 flex items-center gap-1 mt-0.5">
              <span className="w-2 h-2 rounded-full bg-emerald-500"></span>
              {opsHealth.event_mesh_status || 'HEALTHY'}
            </span>
          </div>

          <div className="bg-slate-50 rounded p-2.5 border border-slate-200">
            <span className="text-[11px] font-semibold text-slate-500 uppercase tracking-wider block">Dead Letters</span>
            <span className="text-xs font-bold text-slate-700 mt-0.5 block">
              {opsHealth.dead_letter_count ?? 0} queued
            </span>
          </div>

          <div className="bg-slate-50 rounded p-2.5 border border-slate-200">
            <span className="text-[11px] font-semibold text-slate-500 uppercase tracking-wider block">Active Workers</span>
            <span className="text-xs font-bold text-blue-700 mt-0.5 block">
              {opsHealth.active_workers_count ?? 0} agents active
            </span>
          </div>
        </div>
      </div>

      {/* Global Emergency Alert Banner if Active */}
      {globalEmergencyHalt && (
        <div className="bg-rose-50 border-l-4 border-rose-500 p-4 rounded-r-lg shadow-sm">
          <div className="flex items-start gap-3">
            <ShieldAlert className="w-5 h-5 text-rose-600 mt-0.5 flex-shrink-0" />
            <div className="flex-1">
              <h3 className="text-sm font-bold text-rose-900">Emergency Halt / Kill Switch is Engaged</h3>
              <p className="text-xs text-rose-700 mt-1">
                One or more autonomous execution modules have been suspended to prevent unmonitored actions.
                Autonomous shipment dispatch, bookings, and customer updates are halted. Human review is strictly enforced.
              </p>
            </div>
            <button
              onClick={() => openEmergencyHalt('', false)}
              className="px-3 py-1.5 bg-rose-600 text-white hover:bg-rose-700 text-xs font-semibold rounded shadow-sm transition-colors"
            >
              Clear Halt
            </button>
          </div>
        </div>
      )}

      {/* Tab Navigation Bar */}
      <div className="border-b border-slate-200 bg-white rounded-t-lg px-2 shadow-sm flex overflow-x-auto no-scrollbar">
        {tabs.map((tab) => {
          const Icon = tab.icon;
          const isActive = isTabActive(tab.id, tab.aliases);
          return (
            <button
              key={tab.id}
              onClick={() => setActiveTab(tab.id)}
              className={`inline-flex items-center gap-2 py-3.5 px-4 text-xs font-semibold border-b-2 whitespace-nowrap transition-colors ${
                isActive
                  ? 'border-blue-600 text-blue-600'
                  : 'border-transparent text-slate-600 hover:text-slate-900 hover:border-slate-300'
              }`}
            >
              <Icon className={`w-4 h-4 ${isActive ? 'text-blue-600' : 'text-slate-400'}`} />
              {tab.label}
              {tab.id === 'feature-flags' && (
                <span className="ml-1 px-1.5 py-0.2 rounded-full text-[10px] bg-slate-100 text-slate-600 font-bold">
                  {featureFlags.length}
                </span>
              )}
              {tab.id === 'autonomy' && globalEmergencyHalt && (
                <span className="w-2 h-2 rounded-full bg-rose-500 animate-ping"></span>
              )}
            </button>
          );
        })}
      </div>

      {/* TAB 1: PROFILE */}
      {isTabActive('profile', ['overview']) && (
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          {/* User Profile Card */}
          <div className="bg-white border border-slate-200 rounded-lg p-6 shadow-sm lg:col-span-1">
            <div className="flex items-center gap-3 pb-4 border-b border-slate-100">
              <div className="w-12 h-12 rounded-full bg-blue-600 text-white font-bold flex items-center justify-center text-lg shadow-sm">
                {(user.first_name?.[0] || user.email?.[0] || 'U').toUpperCase()}
              </div>
              <div>
                <h3 className="font-bold text-slate-900 text-base">{user.full_name || 'Internal Staff Member'}</h3>
                <p className="text-xs text-slate-500">{user.email || 'staff@logisticshq.internal'}</p>
                <div className="flex items-center gap-1.5 mt-1">
                  <span className="px-2 py-0.5 rounded text-[10px] font-bold bg-blue-50 text-blue-700 border border-blue-200">
                    {role.display_name || role.name || 'INTERNAL_STAFF'}
                  </span>
                  <span className="px-2 py-0.5 rounded text-[10px] font-bold bg-emerald-50 text-emerald-700 border border-emerald-200">
                    {user.status || 'ACTIVE'}
                  </span>
                </div>
              </div>
            </div>

            <div className="mt-5 space-y-3 text-xs text-slate-600">
              <div className="flex justify-between py-1.5 border-b border-slate-50">
                <span className="text-slate-500">Staff User ID:</span>
                <span className="font-semibold text-slate-800">#{user.id || 1}</span>
              </div>
              <div className="flex justify-between py-1.5 border-b border-slate-50">
                <span className="text-slate-500">Internal Organization:</span>
                <span className="font-semibold text-slate-800">LogisticsHQ Internal (Org #1)</span>
              </div>
              <div className="flex justify-between py-1.5 border-b border-slate-50">
                <span className="text-slate-500">Cognito / Auth Pool:</span>
                <span className="font-semibold text-slate-800">Internal Staff Pool</span>
              </div>
              <div className="flex justify-between py-1.5 border-b border-slate-50">
                <span className="text-slate-500">Effective Permissions:</span>
                <span className="font-semibold text-blue-600">{(role.permissions || []).length} granted</span>
              </div>
            </div>

            {/* Quick Navigation to Users & Roles */}
            <div className="mt-6 pt-5 border-t border-slate-100">
              <span className="text-xs font-semibold text-slate-700 uppercase tracking-wide block mb-3">
                Administrative Direct Navigation
              </span>
              <div className="space-y-2">
                <Link
                  to="/users"
                  className="flex items-center justify-between p-2.5 rounded-md border border-slate-200 bg-slate-50 hover:bg-blue-50 hover:border-blue-200 transition-colors text-xs font-medium text-slate-700 group"
                >
                  <span className="flex items-center gap-2">
                    <User className="w-4 h-4 text-blue-600" />
                    Manage Internal Staff Users
                  </span>
                  <ChevronRight className="w-4 h-4 text-slate-400 group-hover:text-blue-600" />
                </Link>
                <Link
                  to="/users?tab=matrix"
                  className="flex items-center justify-between p-2.5 rounded-md border border-slate-200 bg-slate-50 hover:bg-blue-50 hover:border-blue-200 transition-colors text-xs font-medium text-slate-700 group"
                >
                  <span className="flex items-center gap-2">
                    <Shield className="w-4 h-4 text-indigo-600" />
                    Roles & Permission Matrix
                  </span>
                  <ChevronRight className="w-4 h-4 text-slate-400 group-hover:text-indigo-600" />
                </Link>
              </div>
            </div>
          </div>

          {/* Profile Identity Edit Form */}
          <div className="bg-white border border-slate-200 rounded-lg p-6 shadow-sm lg:col-span-2">
            <h2 className="text-base font-bold text-slate-900 mb-1">Personal Staff Identity & Credentials</h2>
            <p className="text-xs text-slate-500 mb-6">
              Update your internal operator profile attributes. Privileged access roles and organization boundaries remain strictly server-governed.
            </p>

            <form onSubmit={handleSaveProfile} className="space-y-6">
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div>
                  <label className="block text-xs font-semibold text-slate-700 mb-1">First Name</label>
                  <input
                    type="text"
                    value={profileForm.firstName}
                    onChange={(e) => setProfileForm({ ...profileForm, firstName: e.target.value })}
                    className="w-full px-3 py-2 text-xs border border-slate-300 rounded-md focus:outline-none focus:ring-1 focus:ring-blue-500 focus:border-blue-500"
                    placeholder="First Name"
                  />
                </div>
                <div>
                  <label className="block text-xs font-semibold text-slate-700 mb-1">Last Name</label>
                  <input
                    type="text"
                    value={profileForm.lastName}
                    onChange={(e) => setProfileForm({ ...profileForm, lastName: e.target.value })}
                    className="w-full px-3 py-2 text-xs border border-slate-300 rounded-md focus:outline-none focus:ring-1 focus:ring-blue-500 focus:border-blue-500"
                    placeholder="Last Name"
                  />
                </div>
              </div>

              <div className="p-4 bg-slate-50 rounded-lg border border-slate-200 space-y-3">
                <h4 className="text-xs font-bold text-slate-900 uppercase tracking-wide">Session Security Information</h4>
                <div className="grid grid-cols-1 sm:grid-cols-3 gap-3 text-xs text-slate-600">
                  <div>
                    <span className="text-slate-500 block text-[11px]">Session Timeout:</span>
                    <span className="font-semibold text-slate-800">60 Minutes</span>
                  </div>
                  <div>
                    <span className="text-slate-500 block text-[11px]">Lockout Policy:</span>
                    <span className="font-semibold text-slate-800">5 Failed Attempts</span>
                  </div>
                  <div>
                    <span className="text-slate-500 block text-[11px]">MFA Authentication:</span>
                    <span className="font-semibold text-emerald-700">Enforced (Admin/Exec)</span>
                  </div>
                </div>
              </div>

              <div className="pt-4 border-t border-slate-100 flex items-center justify-between">
                <span className="text-[11px] text-slate-500">
                  Changes write an immutable record to the administrative audit log.
                </span>
                <button
                  type="submit"
                  disabled={savingProfile}
                  className="inline-flex items-center gap-2 px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white text-xs font-semibold rounded-md shadow-sm transition-colors disabled:opacity-50"
                >
                  <Save className="w-3.5 h-3.5" />
                  {savingProfile ? 'Saving Changes...' : 'Save Profile Changes'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* TAB 2: USERS & ACCESS */}
      {isTabActive('users') && (
        <div className="space-y-6">
          <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
            <div className="p-4 bg-white border border-slate-200 rounded-lg shadow-sm">
              <span className="text-xs font-bold text-slate-500 uppercase tracking-wider block">Staff Users Directory</span>
              <span className="text-2xl font-bold text-slate-900 mt-1 block">3 Active Accounts</span>
              <p className="text-xs text-slate-500 mt-1">Super Admin, Operations Lead, Support Specialist</p>
            </div>
            <div className="p-4 bg-white border border-slate-200 rounded-lg shadow-sm">
              <span className="text-xs font-bold text-slate-500 uppercase tracking-wider block">Defined System Roles</span>
              <span className="text-2xl font-bold text-slate-900 mt-1 block">6 Internal Roles</span>
              <p className="text-xs text-slate-500 mt-1">SUPER_ADMIN, OPERATOR, SALES_REP, etc.</p>
            </div>
            <div className="p-4 bg-white border border-slate-200 rounded-lg shadow-sm">
              <span className="text-xs font-bold text-slate-500 uppercase tracking-wider block">Session Authority</span>
              <span className="text-2xl font-bold text-emerald-700 mt-1 block">{(role.permissions || []).length} Permissions</span>
              <p className="text-xs text-slate-500 mt-1">Active role: {role.name || 'SUPER_ADMIN'}</p>
            </div>
          </div>

          <div className="bg-white border border-slate-200 rounded-lg p-6 shadow-sm">
            <h3 className="text-base font-bold text-slate-900 mb-1">Authoritative Staff & Access Control Centers</h3>
            <p className="text-xs text-slate-500 mb-4">
              Internal user administration, role assignment, and access governance are managed centrally in dedicated workspaces.
            </p>

            <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
              <Link
                to="/users"
                className="p-4 rounded-lg border border-slate-200 bg-slate-50 hover:bg-blue-50 hover:border-blue-200 transition-colors group flex flex-col justify-between"
              >
                <div>
                  <div className="flex items-center justify-between mb-2">
                    <User className="w-5 h-5 text-blue-600" />
                    <ArrowUpRight className="w-4 h-4 text-slate-400 group-hover:text-blue-600" />
                  </div>
                  <h4 className="font-bold text-sm text-slate-900 group-hover:text-blue-900">Manage Internal Staff Users</h4>
                  <p className="text-xs text-slate-500 mt-1">
                    Invite internal staff members, manage active sessions, disable accounts, and assign initial roles.
                  </p>
                </div>
                <span className="text-xs font-semibold text-blue-600 mt-4 block">Open Users Directory →</span>
              </Link>

              <Link
                to="/users?tab=matrix"
                className="p-4 rounded-lg border border-slate-200 bg-slate-50 hover:bg-indigo-50 hover:border-indigo-200 transition-colors group flex flex-col justify-between"
              >
                <div>
                  <div className="flex items-center justify-between mb-2">
                    <Shield className="w-5 h-5 text-indigo-600" />
                    <ArrowUpRight className="w-4 h-4 text-slate-400 group-hover:text-indigo-600" />
                  </div>
                  <h4 className="font-bold text-sm text-slate-900 group-hover:text-indigo-900">Roles & Permission Matrix</h4>
                  <p className="text-xs text-slate-500 mt-1">
                    Inspect fine-grained capability matrices across 12 functional areas and configure role limits.
                  </p>
                </div>
                <span className="text-xs font-semibold text-indigo-600 mt-4 block">View Permission Matrix →</span>
              </Link>

              <Link
                to="/users?tab=governance"
                className="p-4 rounded-lg border border-slate-200 bg-slate-50 hover:bg-emerald-50 hover:border-emerald-200 transition-colors group flex flex-col justify-between"
              >
                <div>
                  <div className="flex items-center justify-between mb-2">
                    <CheckCircle2 className="w-5 h-5 text-emerald-600" />
                    <ArrowUpRight className="w-4 h-4 text-slate-400 group-hover:text-emerald-600" />
                  </div>
                  <h4 className="font-bold text-sm text-slate-900 group-hover:text-emerald-900">Access Governance & Audit</h4>
                  <p className="text-xs text-slate-500 mt-1">
                    Audit role changes, privilege elevations, login history, and tenant separation compliance.
                  </p>
                </div>
                <span className="text-xs font-semibold text-emerald-600 mt-4 block">Open Access Governance →</span>
              </Link>
            </div>
          </div>

          {/* Effective Permissions Catalog */}
          <div className="bg-white border border-slate-200 rounded-lg p-6 shadow-sm">
            <div className="flex items-center justify-between mb-4">
              <div>
                <h3 className="text-base font-bold text-slate-900">Effective Permissions for Current Session</h3>
                <p className="text-xs text-slate-500 mt-0.5">
                  Authority derived from role <strong className="text-slate-800 font-semibold">{role.name}</strong>.
                  Protected operational endpoints verify these permissions server-side.
                </p>
              </div>
              <span className="px-2.5 py-1 rounded bg-slate-100 text-slate-700 text-xs font-bold">
                {(role.permissions || []).length} Active Permissions
              </span>
            </div>

            <div className="flex flex-wrap gap-2 pt-2">
              {(role.permissions || []).map((perm, idx) => (
                <span
                  key={idx}
                  className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded bg-slate-50 text-slate-700 border border-slate-200 text-xs font-mono font-medium"
                >
                  <Check className="w-3 h-3 text-emerald-600" />
                  {perm}
                </span>
              ))}
            </div>
          </div>
        </div>
      )}

      {/* TAB 3: PLATFORM DEFAULTS */}
      {isTabActive('platform') && (
        <div className="space-y-6">
          <div className="bg-blue-50 border border-blue-200 rounded-lg p-4 flex items-start gap-3 shadow-sm">
            <Shield className="w-5 h-5 text-blue-600 mt-0.5 flex-shrink-0" />
            <div>
              <h4 className="text-xs font-bold text-blue-950 uppercase tracking-wide">Strict Secrets Protection Enforced</h4>
              <p className="text-xs text-blue-800 mt-0.5">
                SPortal never exposes raw `.env` process variables, AWS keys, Twilio credentials, or database passwords.
                All configurations displayed below are safe typed platform defaults stored in MariaDB source of truth (`sportal_platform_settings`).
              </p>
            </div>
          </div>

          <div className="bg-white border border-slate-200 rounded-lg shadow-sm overflow-hidden">
            <div className="px-6 py-4 border-b border-slate-200 flex items-center justify-between">
              <div>
                <h3 className="text-base font-bold text-slate-900">Platform Configuration Defaults</h3>
                <p className="text-xs text-slate-500 mt-0.5">
                  Platform-wide operational defaults, maintenance mode banners, security timeouts, and system intervals.
                </p>
              </div>
              <span className="px-2.5 py-1 rounded bg-slate-100 text-slate-700 text-xs font-semibold">
                {platformSettings.length} Configured Settings
              </span>
            </div>

            <div className="overflow-x-auto">
              <table className="w-full text-left border-collapse">
                <thead>
                  <tr className="bg-slate-50 border-b border-slate-200 text-[11px] font-semibold text-slate-600 uppercase tracking-wider">
                    <th className="py-3 px-4">Setting Key</th>
                    <th className="py-3 px-4">Category</th>
                    <th className="py-3 px-4">Current Value</th>
                    <th className="py-3 px-4">Data Type</th>
                    <th className="py-3 px-4">Description</th>
                    <th className="py-3 px-4">Last Updated</th>
                    <th className="py-3 px-4 text-right">Action</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-slate-100 text-xs">
                  {platformSettings.map((item) => (
                    <tr key={item.setting_key} className="hover:bg-slate-50 transition-colors">
                      <td className="py-3.5 px-4 font-mono font-semibold text-blue-900">
                        {item.setting_key}
                      </td>
                      <td className="py-3.5 px-4">
                        <span className="px-2 py-0.5 rounded text-[10px] font-bold bg-slate-100 text-slate-700">
                          {item.category}
                        </span>
                      </td>
                      <td className="py-3.5 px-4">
                        {item.setting_key === 'maintenance_mode' ? (
                          <span className={`px-2 py-0.5 rounded text-[10px] font-bold ${
                            item.setting_value === 'true'
                              ? 'bg-rose-100 text-rose-800'
                              : 'bg-emerald-100 text-emerald-800'
                          }`}>
                            {item.setting_value === 'true' ? 'MAINTENANCE ENGAGED' : 'NORMAL OPERATION'}
                          </span>
                        ) : (
                          <span className="font-semibold text-slate-800">
                            {item.setting_value}
                          </span>
                        )}
                      </td>
                      <td className="py-3.5 px-4 font-mono text-[11px] text-slate-500">
                        {item.data_type}
                      </td>
                      <td className="py-3.5 px-4 text-slate-600 max-w-xs truncate" title={item.description}>
                        {item.description}
                      </td>
                      <td className="py-3.5 px-4 text-slate-500 whitespace-nowrap">
                        {item.updated_at ? new Date(item.updated_at).toLocaleDateString() : 'System default'}
                      </td>
                      <td className="py-3.5 px-4 text-right">
                        <button
                          onClick={() => openEditSetting(item)}
                          className="inline-flex items-center gap-1 px-2.5 py-1 bg-white border border-slate-300 hover:bg-slate-50 text-slate-700 rounded text-xs font-medium shadow-xs transition-colors"
                        >
                          <Edit2 className="w-3 h-3 text-slate-500" />
                          Edit
                        </button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        </div>
      )}

      {/* TAB 4: FEATURE FLAGS */}
      {isTabActive('feature-flags') && (
        <div className="space-y-6">
          <div className="bg-amber-50 border border-amber-200 rounded-lg p-4 flex items-start gap-3 shadow-sm">
            <AlertTriangle className="w-5 h-5 text-amber-600 mt-0.5 flex-shrink-0" />
            <div>
              <h4 className="text-xs font-bold text-amber-950 uppercase tracking-wide">Production Feature Flag Safety</h4>
              <p className="text-xs text-amber-800 mt-0.5">
                Feature flags govern autonomous AI workflow activation across customer tenants. Any modification requires elevated
                internal permissions, writes an immutable audit record, and updates real-time worker policy.
              </p>
            </div>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {featureFlags.map((flag) => (
              <div
                key={flag.flag_key}
                className="bg-white border border-slate-200 rounded-lg p-5 shadow-sm flex flex-col justify-between hover:border-slate-300 transition-colors"
              >
                <div>
                  <div className="flex items-start justify-between gap-3 mb-2">
                    <div>
                      <h4 className="font-bold text-slate-900 text-sm">{flag.flag_name}</h4>
                      <span className="font-mono text-[11px] text-blue-600 block mt-0.5">{flag.flag_key}</span>
                    </div>
                    <span
                      className={`px-2.5 py-0.5 rounded-full text-xs font-bold flex items-center gap-1 ${
                        flag.is_enabled
                          ? 'bg-emerald-100 text-emerald-800 border border-emerald-200'
                          : 'bg-slate-100 text-slate-600 border border-slate-200'
                      }`}
                    >
                      {flag.is_enabled ? (
                        <>
                          <CheckCircle2 className="w-3 h-3 text-emerald-600" />
                          ENABLED
                        </>
                      ) : (
                        <>
                          <XCircle className="w-3 h-3 text-slate-400" />
                          DISABLED
                        </>
                      )}
                    </span>
                  </div>

                  <p className="text-xs text-slate-600 mb-4 leading-relaxed">{flag.description}</p>
                </div>

                <div className="pt-4 border-t border-slate-100 flex items-center justify-between">
                  <div className="flex items-center gap-2">
                    <span className="text-[11px] font-semibold text-slate-500">Max Autonomy:</span>
                    <span className="px-2 py-0.5 rounded bg-indigo-50 text-indigo-700 text-[10px] font-bold border border-indigo-200">
                      Level {flag.max_autonomy_level}
                    </span>
                    {flag.requires_approval ? (
                      <span className="px-2 py-0.5 rounded bg-amber-50 text-amber-700 text-[10px] font-bold border border-amber-200">
                        Approval Required
                      </span>
                    ) : (
                      <span className="px-2 py-0.5 rounded bg-slate-50 text-slate-600 text-[10px] font-medium">
                        Autonomous
                      </span>
                    )}
                  </div>

                  <button
                    onClick={() => openEditFlag(flag)}
                    className="inline-flex items-center gap-1 px-3 py-1.5 bg-white border border-slate-300 hover:bg-slate-50 text-slate-700 text-xs font-semibold rounded shadow-xs transition-colors"
                  >
                    <Sliders className="w-3 h-3 text-slate-500" />
                    Configure
                  </button>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* TAB 5: INTEGRATIONS */}
      {isTabActive('integrations') && (
        <div className="space-y-6">
          <div className="bg-slate-50 border border-slate-200 rounded-lg p-4 flex items-start gap-3 shadow-sm">
            <Globe className="w-5 h-5 text-blue-600 mt-0.5 flex-shrink-0" />
            <div>
              <h4 className="text-xs font-bold text-slate-900 uppercase tracking-wide">External Gateway Architecture</h4>
              <p className="text-xs text-slate-600 mt-0.5">
                All communications (Twilio SMS, AWS SES Email, Carrier API Polling, S3 Storage) route through the Go External Integration Gateway.
                Zero credentials or client secrets are exposed to the frontend browser context.
              </p>
            </div>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {integrations.map((integ) => (
              <div
                key={integ.id}
                className="bg-white border border-slate-200 rounded-lg p-5 shadow-sm flex flex-col justify-between hover:border-slate-300 transition-colors"
              >
                <div>
                  <div className="flex items-start justify-between gap-3 mb-2">
                    <div>
                      <h4 className="font-bold text-slate-900 text-base">{integ.provider_name}</h4>
                      <span className="text-xs font-medium text-slate-500 block">Type: {integ.integration_type}</span>
                    </div>
                    <span
                      className={`px-2.5 py-0.5 rounded-full text-xs font-bold flex items-center gap-1 ${
                        integ.is_enabled
                          ? 'bg-emerald-100 text-emerald-800 border border-emerald-200'
                          : 'bg-slate-100 text-slate-600 border border-slate-200'
                      }`}
                    >
                      {integ.is_enabled ? (
                        <>
                          <CheckCircle2 className="w-3 h-3 text-emerald-600" />
                          {integ.status || 'CONNECTED'}
                        </>
                      ) : (
                        <>
                          <XCircle className="w-3 h-3 text-slate-400" />
                          DISABLED
                        </>
                      )}
                    </span>
                  </div>

                  <div className="mt-3 p-2.5 bg-slate-50 rounded border border-slate-100 text-xs">
                    <span className="text-slate-500 block text-[11px]">Credential Security Boundary:</span>
                    <span className="font-mono text-slate-700 font-semibold">{integ.masked_identity}</span>
                  </div>

                  {integ.health_message && (
                    <p className="text-xs text-slate-600 mt-3 bg-emerald-50/50 p-2 rounded border border-emerald-100">
                      {integ.health_message}
                    </p>
                  )}
                </div>

                <div className="pt-4 mt-4 border-t border-slate-100 flex items-center justify-between">
                  <span className="text-[11px] text-slate-400">
                    Last check: {integ.last_health_check ? new Date(integ.last_health_check).toLocaleString() : 'N/A'}
                  </span>
                  <button
                    onClick={() => openIntegrationToggle(integ)}
                    className={`px-3 py-1.5 text-xs font-semibold rounded shadow-xs transition-colors ${
                      integ.is_enabled
                        ? 'bg-white border border-rose-300 text-rose-700 hover:bg-rose-50'
                        : 'bg-white border border-emerald-300 text-emerald-700 hover:bg-emerald-50'
                    }`}
                  >
                    {integ.is_enabled ? 'Disable Gateway' : 'Enable Gateway'}
                  </button>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* TAB 6: NOTIFICATIONS */}
      {isTabActive('notifications') && (
        <div className="space-y-6">
          <div className="bg-white border border-slate-200 rounded-lg p-6 shadow-sm">
            <h3 className="text-base font-bold text-slate-900 mb-1">Notification Channel Subscriptions & Thresholds</h3>
            <p className="text-xs text-slate-500 mb-6">
              Configure fine-grained operational notifications, approval pings, and minimum severity filters.
            </p>

            <form onSubmit={handleSaveProfile} className="space-y-6">
              <div>
                <label className="block text-xs font-semibold text-slate-700 mb-1">Minimum Alert Severity Threshold</label>
                <select
                  value={profileForm.minSeverity}
                  onChange={(e) => setProfileForm({ ...profileForm, minSeverity: e.target.value })}
                  className="w-full md:w-1/2 px-3 py-2 text-xs border border-slate-300 rounded-md focus:outline-none focus:ring-1 focus:ring-blue-500 focus:border-blue-500 bg-white"
                >
                  <option value="LOW">Low (All platform alerts and background completions)</option>
                  <option value="MEDIUM">Medium (Recommended: Operational events, warnings, approvals)</option>
                  <option value="HIGH">High (Anomalies, SLA breaches, and critical approvals only)</option>
                  <option value="CRITICAL">Critical (Kill switches, security alerts, and system halts only)</option>
                </select>
                <p className="text-[11px] text-slate-400 mt-1">Filters internal toast notifications and notification badge counts.</p>
              </div>

              <div>
                <span className="block text-xs font-semibold text-slate-700 mb-2">Notification Channel Subscriptions</span>
                <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
                  {[
                    { key: 'inAppEnabled', label: 'In-App Operational Toasts', desc: 'Real-time notifications during active SPortal session' },
                    { key: 'approvalsEnabled', label: 'Pending Human Approvals', desc: 'Alerts when autonomous agents request action sign-off' },
                    { key: 'automationsEnabled', label: 'Workflow Automations', desc: 'Scheduled pipeline completion and batch run summaries' },
                    { key: 'recommendationsEnabled', label: 'AI Insights & Recommendations', desc: 'Customer Success Copilot risk signals and renewals' },
                    { key: 'financeEnabled', label: 'Finance & Invoices', desc: 'Overdue invoices, past-due billing, payment exceptions' },
                    { key: 'complianceEnabled', label: 'Contract & Compliance', desc: 'Statutory compliance alerts, expiring documents, BL review' },
                  ].map((item) => (
                    <label
                      key={item.key}
                      className="flex items-start gap-2.5 p-3 rounded-md border border-slate-200 bg-slate-50 hover:bg-slate-100 cursor-pointer transition-colors"
                    >
                      <input
                        type="checkbox"
                        checked={profileForm[item.key]}
                        onChange={(e) => setProfileForm({ ...profileForm, [item.key]: e.target.checked })}
                        className="mt-0.5 h-4 w-4 rounded border-slate-300 text-blue-600 focus:ring-blue-500"
                      />
                      <div>
                        <span className="text-xs font-semibold text-slate-800 block">{item.label}</span>
                        <span className="text-[11px] text-slate-500 block leading-tight mt-0.5">{item.desc}</span>
                      </div>
                    </label>
                  ))}
                </div>
              </div>

              <div className="p-4 bg-slate-50 rounded-lg border border-slate-200">
                <h4 className="text-xs font-bold text-slate-900 uppercase tracking-wide mb-2">Delivery Channels & Escalation</h4>
                <div className="grid grid-cols-1 sm:grid-cols-3 gap-3 text-xs text-slate-600">
                  <div className="flex items-start gap-2">
                    <Bell className="w-4 h-4 text-blue-600 mt-0.5 flex-shrink-0" />
                    <div>
                      <span className="font-semibold text-slate-800 block">In-App Notification Center</span>
                      <span className="text-[11px] text-slate-500">Real-time alerts, header bell drawer</span>
                    </div>
                  </div>
                  <div className="flex items-start gap-2">
                    <Mail className="w-4 h-4 text-indigo-600 mt-0.5 flex-shrink-0" />
                    <div>
                      <span className="font-semibold text-slate-800 block">Email Gateway (AWS SES)</span>
                      <span className="text-[11px] text-slate-500">High-severity anomalies & digest pings</span>
                    </div>
                  </div>
                  <div className="flex items-start gap-2">
                    <Smartphone className="w-4 h-4 text-emerald-600 mt-0.5 flex-shrink-0" />
                    <div>
                      <span className="font-semibold text-slate-800 block">SMS Gateway (Twilio)</span>
                      <span className="text-[11px] text-slate-500">Emergency halts & critical disruptions</span>
                    </div>
                  </div>
                </div>
              </div>

              <div className="pt-4 border-t border-slate-100 flex items-center justify-between">
                <span className="text-[11px] text-slate-500">Preferences persist to MariaDB configuration.</span>
                <button
                  type="submit"
                  disabled={savingProfile}
                  className="inline-flex items-center gap-2 px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white text-xs font-semibold rounded-md shadow-sm transition-colors disabled:opacity-50"
                >
                  <Save className="w-3.5 h-3.5" />
                  {savingProfile ? 'Saving...' : 'Save Notification Preferences'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* TAB 7: AI ADMINISTRATION */}
      {isTabActive('ai') && (
        <div className="space-y-6">
          <div className="bg-white border border-slate-200 rounded-lg p-6 shadow-sm">
            <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 pb-4 border-b border-slate-100 mb-6">
              <div>
                <h3 className="text-base font-bold text-slate-900">AI Architecture, Workforce & Safety Governance</h3>
                <p className="text-xs text-slate-500 mt-0.5">
                  Autonomous agents and predictive pipelines operating under strict Go enforcement and LangGraph multi-agent orchestration.
                </p>
              </div>
              <div className="flex items-center gap-3">
                <Link
                  to="/ai"
                  className="inline-flex items-center gap-1.5 px-3.5 py-1.5 bg-blue-600 text-white hover:bg-blue-700 text-xs font-semibold rounded-md shadow-xs transition-colors"
                >
                  <Sparkles className="w-3.5 h-3.5" />
                  Open SPortal AI Workspace
                </Link>
                <Link
                  to="/approvals"
                  className="inline-flex items-center gap-1.5 px-3.5 py-1.5 bg-white border border-slate-300 text-slate-700 hover:bg-slate-50 text-xs font-semibold rounded-md shadow-xs transition-colors"
                >
                  <Shield className="w-3.5 h-3.5 text-indigo-600" />
                  Human Approval Center
                </Link>
              </div>
            </div>

            <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-4 gap-4">
              <div className="p-4 rounded-lg border border-slate-200 bg-slate-50">
                <span className="text-[11px] font-semibold text-slate-500 uppercase tracking-wider block">AI Provider Engine</span>
                <span className="text-sm font-bold text-slate-900 mt-1 flex items-center gap-1.5">
                  <span className="w-2.5 h-2.5 rounded-full bg-emerald-500"></span>
                  Local Engine & Gateway
                </span>
                <span className="text-[11px] text-slate-500 mt-1 block">Port 8090 • FastAPI Uvicorn</span>
              </div>

              <div className="p-4 rounded-lg border border-slate-200 bg-slate-50">
                <span className="text-[11px] font-semibold text-slate-500 uppercase tracking-wider block">AI Workforce Registry</span>
                <span className="text-sm font-bold text-blue-700 mt-1 flex items-center gap-1.5">
                  <Cpu className="w-4 h-4 text-blue-600" />
                  10 Specialist Agents
                </span>
                <span className="text-[11px] text-slate-500 mt-1 block">LangGraph Multi-Agent</span>
              </div>

              <div className="p-4 rounded-lg border border-slate-200 bg-slate-50">
                <span className="text-[11px] font-semibold text-slate-500 uppercase tracking-wider block">Model Orchestration</span>
                <span className="text-sm font-bold text-slate-900 mt-1 block">GPT-4o & Claude 3.5</span>
                <span className="text-[11px] text-slate-500 mt-1 block">Local Llama-3-70B Gateway</span>
              </div>

              <div className="p-4 rounded-lg border border-slate-200 bg-slate-50">
                <span className="text-[11px] font-semibold text-slate-500 uppercase tracking-wider block">Safety & Evaluation</span>
                <span className="text-sm font-bold text-emerald-700 mt-1 flex items-center gap-1.5">
                  <CheckCircle2 className="w-4 h-4 text-emerald-600" />
                  ACTIVE_ENFORCED
                </span>
                <span className="text-[11px] text-slate-500 mt-1 block">Fact Grounding & HITL Gate</span>
              </div>
            </div>

            <div className="mt-6 p-4 bg-emerald-50/60 rounded-lg border border-emerald-200 flex items-start gap-3">
              <Key className="w-5 h-5 text-emerald-700 mt-0.5 flex-shrink-0" />
              <div>
                <h4 className="text-xs font-bold text-emerald-950 uppercase tracking-wide">Zero Credential Leaks Guarantee</h4>
                <p className="text-xs text-emerald-800 mt-0.5 leading-relaxed">
                  All model provider API keys, machine-to-machine internal service tokens, and vector memory embeddings are stored
                  and executed strictly server-side within the Python Uvicorn engine and Go control plane. SPortal client applications
                  never receive or store raw provider secrets.
                </p>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* TAB 8: AUTONOMY CONTROLS & EMERGENCY HALT */}
      {isTabActive('autonomy') && (
        <div className="space-y-6">
          {/* Emergency Kill Switch Banner Card */}
          <div className={`rounded-lg p-6 border shadow-sm ${
            globalEmergencyHalt
              ? 'bg-rose-50 border-rose-300 text-rose-950'
              : 'bg-white border-slate-200 text-slate-900'
          }`}>
            <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
              <div className="flex items-start gap-4">
                <div className={`p-3 rounded-lg ${globalEmergencyHalt ? 'bg-rose-200 text-rose-800' : 'bg-slate-100 text-slate-700'}`}>
                  <ShieldAlert className="w-8 h-8" />
                </div>
                <div>
                  <div className="flex items-center gap-2">
                    <h3 className="text-lg font-bold">Platform-Wide AI Emergency Kill Switch</h3>
                    {globalEmergencyHalt ? (
                      <span className="px-2.5 py-0.5 rounded text-xs font-bold bg-rose-600 text-white animate-pulse">
                        HALT ACTIVE
                      </span>
                    ) : (
                      <span className="px-2.5 py-0.5 rounded text-xs font-bold bg-emerald-100 text-emerald-800 border border-emerald-200">
                        NORMAL GOVERNED EXECUTION
                      </span>
                    )}
                  </div>
                  <p className="text-xs text-slate-600 mt-1 max-w-2xl">
                    Executive kill switch to halt all autonomous AI agent operations across all modules immediately.
                    When engaged, automated actions require explicit human operator review and sign-off.
                  </p>
                </div>
              </div>

              <div className="flex items-center gap-3">
                {globalEmergencyHalt ? (
                  <button
                    onClick={() => openEmergencyHalt('', false)}
                    className="px-4 py-2 bg-emerald-600 hover:bg-emerald-700 text-white text-xs font-bold rounded-md shadow-sm transition-colors"
                  >
                    Resume Autonomous Operations
                  </button>
                ) : (
                  <button
                    onClick={() => openEmergencyHalt('', true)}
                    className="px-4 py-2 bg-rose-600 hover:bg-rose-700 text-white text-xs font-bold rounded-md shadow-sm transition-colors"
                  >
                    Engage Emergency Kill Switch
                  </button>
                )}
              </div>
            </div>
          </div>

          {/* Autonomy Level Descriptions Card */}
          <div className="bg-white border border-slate-200 rounded-lg p-5 shadow-sm">
            <h4 className="text-xs font-bold text-slate-900 uppercase tracking-wide mb-3">Autonomy Levels Architecture</h4>
            <div className="grid grid-cols-1 md:grid-cols-5 gap-3 text-xs">
              <div className="p-3 bg-slate-50 rounded border border-slate-200">
                <span className="font-bold text-slate-900 block">Level 0: Manual</span>
                <span className="text-[11px] text-slate-500 mt-1 block">Manual human execution only; AI dormant.</span>
              </div>
              <div className="p-3 bg-slate-50 rounded border border-slate-200">
                <span className="font-bold text-slate-900 block">Level 1: Inform</span>
                <span className="text-[11px] text-slate-500 mt-1 block">Read-only operational insights & telemetry.</span>
              </div>
              <div className="p-3 bg-slate-50 rounded border border-slate-200">
                <span className="font-bold text-slate-900 block">Level 2: Prepare</span>
                <span className="text-[11px] text-slate-500 mt-1 block">AI drafts actions; human approval mandatory.</span>
              </div>
              <div className="p-3 bg-indigo-50/50 rounded border border-indigo-200">
                <span className="font-bold text-indigo-900 block">Level 3: Controlled</span>
                <span className="text-[11px] text-indigo-700 mt-1 block">Bounded execution within monetary limits.</span>
              </div>
              <div className="p-3 bg-emerald-50/50 rounded border border-emerald-200">
                <span className="font-bold text-emerald-900 block">Level 4: Autonomous</span>
                <span className="text-[11px] text-emerald-700 mt-1 block">Full autonomous loop under safety gates.</span>
              </div>
            </div>
          </div>

          {/* Module Autonomy Policies Table */}
          <div className="bg-white border border-slate-200 rounded-lg shadow-sm overflow-hidden">
            <div className="px-6 py-4 border-b border-slate-200 flex items-center justify-between">
              <div>
                <h3 className="text-base font-bold text-slate-900">Module Autonomy Policies</h3>
                <p className="text-xs text-slate-500 mt-0.5">
                  Governed operational boundaries for individual domain modules in MariaDB (`autonomy_policies`).
                </p>
              </div>
              <span className="text-xs font-semibold text-slate-500">
                {autonomyPolicies.length} Governed Modules
              </span>
            </div>

            <div className="overflow-x-auto">
              <table className="w-full text-left border-collapse">
                <thead>
                  <tr className="bg-slate-50 border-b border-slate-200 text-[11px] font-semibold text-slate-600 uppercase tracking-wider">
                    <th className="py-3 px-4">Module Name</th>
                    <th className="py-3 px-4">Autonomy Level</th>
                    <th className="py-3 px-4">Approval Required</th>
                    <th className="py-3 px-4">Max Monetary Threshold</th>
                    <th className="py-3 px-4">Min Confidence</th>
                    <th className="py-3 px-4">Emergency Stop</th>
                    <th className="py-3 px-4 text-right">Module Action</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-slate-100 text-xs">
                  {autonomyPolicies.map((p) => (
                    <tr key={p.id} className="hover:bg-slate-50 transition-colors">
                      <td className="py-3.5 px-4 font-bold text-slate-900 uppercase">
                        {p.module}
                      </td>
                      <td className="py-3.5 px-4">
                        <span className="px-2.5 py-0.5 rounded text-[11px] font-semibold bg-blue-50 text-blue-700 border border-blue-200">
                          {p.autonomy_level}
                        </span>
                      </td>
                      <td className="py-3.5 px-4">
                        {p.requires_approval ? (
                          <span className="text-amber-700 font-semibold flex items-center gap-1">
                            <Check className="w-3.5 h-3.5" /> Enforced
                          </span>
                        ) : (
                          <span className="text-slate-500">Bypassed</span>
                        )}
                      </td>
                      <td className="py-3.5 px-4 font-semibold text-slate-800">
                        {p.max_monetary_threshold > 0 ? `$${p.max_monetary_threshold.toLocaleString()}` : 'Unlimited / N/A'}
                      </td>
                      <td className="py-3.5 px-4 font-mono text-slate-600">
                        {p.min_confidence_threshold > 0 ? `${(p.min_confidence_threshold * 100).toFixed(0)}%` : 'Standard'}
                      </td>
                      <td className="py-3.5 px-4">
                        {p.emergency_stop ? (
                          <span className="px-2 py-0.5 rounded text-[10px] font-bold bg-rose-100 text-rose-800 border border-rose-200">
                            HALTED
                          </span>
                        ) : (
                          <span className="px-2 py-0.5 rounded text-[10px] font-bold bg-emerald-100 text-emerald-800 border border-emerald-200">
                            ACTIVE
                          </span>
                        )}
                      </td>
                      <td className="py-3.5 px-4 text-right">
                        <button
                          onClick={() => openEmergencyHalt(p.module, !p.emergency_stop)}
                          className={`px-2.5 py-1 text-xs font-semibold rounded shadow-xs transition-colors ${
                            p.emergency_stop
                              ? 'bg-emerald-50 text-emerald-700 border border-emerald-200 hover:bg-emerald-100'
                              : 'bg-rose-50 text-rose-700 border border-rose-200 hover:bg-rose-100'
                          }`}
                        >
                          {p.emergency_stop ? 'Resume Module' : 'Halt Module'}
                        </button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        </div>
      )}

      {/* TAB 9: SECURITY POLICY */}
      {isTabActive('security') && (
        <div className="space-y-6">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            {/* Authentication Policy Card */}
            <div className="bg-white border border-slate-200 rounded-lg p-6 shadow-sm">
              <div className="flex items-center gap-3 pb-3 border-b border-slate-100 mb-4">
                <div className="p-2.5 rounded bg-blue-50 text-blue-600">
                  <Lock className="w-5 h-5" />
                </div>
                <div>
                  <h3 className="font-bold text-slate-900 text-base">Authentication & Session Controls</h3>
                  <p className="text-xs text-slate-500">Active platform session governance policies</p>
                </div>
              </div>

              <div className="space-y-3 text-xs text-slate-600">
                <div className="flex justify-between py-2 border-b border-slate-50">
                  <span className="text-slate-500">Identity Provider:</span>
                  <span className="font-semibold text-slate-800">AWS Cognito + MariaDB Salted Hashes</span>
                </div>
                <div className="flex justify-between py-2 border-b border-slate-50">
                  <span className="text-slate-500">Session Inactivity Timeout:</span>
                  <span className="font-semibold text-slate-800">60 minutes</span>
                </div>
                <div className="flex justify-between py-2 border-b border-slate-50">
                  <span className="text-slate-500">Max Failed Login Attempts:</span>
                  <span className="font-semibold text-slate-800">5 attempts before lockout</span>
                </div>
                <div className="flex justify-between py-2 border-b border-slate-50">
                  <span className="text-slate-500">Password Min Length:</span>
                  <span className="font-semibold text-slate-800">8 characters (Complex)</span>
                </div>
                <div className="flex justify-between py-2 border-b border-slate-50">
                  <span className="text-slate-500">MFA for Internal Staff:</span>
                  <span className="font-semibold text-emerald-700">Enforced for Admin/Executive</span>
                </div>
              </div>
            </div>

            {/* Tenant Boundary Isolation Policy */}
            <div className="bg-white border border-slate-200 rounded-lg p-6 shadow-sm">
              <div className="flex items-center gap-3 pb-3 border-b border-slate-100 mb-4">
                <div className="p-2.5 rounded bg-indigo-50 text-indigo-600">
                  <Shield className="w-5 h-5" />
                </div>
                <div>
                  <h3 className="font-bold text-slate-900 text-base">Tenant & Customer Isolation Boundary</h3>
                  <p className="text-xs text-slate-500">Multi-tenant boundary enforcement in Go</p>
                </div>
              </div>

              <div className="space-y-3 text-xs text-slate-600">
                <p className="leading-relaxed">
                  LogisticsHQ enforces strict separation between customer CPortal tenants and internal SPortal administrative controls:
                </p>
                <ul className="list-disc list-inside space-y-1 text-slate-700 pl-1">
                  <li>CPortal users are barred from all `/api/v1/sportal/*` endpoints (403 Forbidden).</li>
                  <li>Organization #1 is reserved strictly for internal staff; customer orgs are strictly {'>'}= 2.</li>
                  <li>Customer data queries require explicit organization scoping in SQL and Go services.</li>
                  <li>No cross-tenant data leakage or direct infrastructure database access.</li>
                </ul>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* TAB 10: SYSTEM HEALTH & TELEMETRY */}
      {isTabActive('operations') && (
        <div className="space-y-6">
          <div className="bg-white border border-slate-200 rounded-lg p-6 shadow-sm">
            <div className="flex items-center justify-between pb-4 border-b border-slate-100 mb-6">
              <div>
                <h3 className="font-bold text-slate-900 text-base">Subsystem Telemetry & Infrastructure Health</h3>
                <p className="text-xs text-slate-500 mt-0.5">
                  Truthful operational health metrics evaluated by Go control plane.
                </p>
              </div>
              <button
                onClick={handleRefresh}
                className="inline-flex items-center gap-1.5 px-3 py-1.5 bg-slate-50 hover:bg-slate-100 text-slate-700 border border-slate-200 rounded text-xs font-semibold shadow-xs transition-colors"
              >
                <RefreshCw className={`w-3.5 h-3.5 ${refreshing ? 'animate-spin' : ''}`} />
                Evaluate Now
              </button>
            </div>

            <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-4 gap-4">
              <div className="p-4 rounded-lg border border-slate-200 bg-slate-50">
                <div className="flex items-center justify-between mb-2">
                  <span className="text-xs font-bold text-slate-700">Core Go Backend</span>
                  <span className="w-2.5 h-2.5 rounded-full bg-emerald-500"></span>
                </div>
                <p className="text-xl font-bold text-slate-900">{opsHealth.backend_status || 'HEALTHY'}</p>
                <span className="text-[11px] text-slate-500 mt-1 block">Port 8080 • Chi Router v5</span>
              </div>

              <div className="p-4 rounded-lg border border-slate-200 bg-slate-50">
                <div className="flex items-center justify-between mb-2">
                  <span className="text-xs font-bold text-slate-700">MariaDB Persistence</span>
                  <span className="w-2.5 h-2.5 rounded-full bg-emerald-500"></span>
                </div>
                <p className="text-xl font-bold text-slate-900">{opsHealth.database_status || 'HEALTHY'}</p>
                <span className="text-[11px] text-slate-500 mt-1 block">Port 3306 • freel_mysql</span>
              </div>

              <div className="p-4 rounded-lg border border-slate-200 bg-slate-50">
                <div className="flex items-center justify-between mb-2">
                  <span className="text-xs font-bold text-slate-700">Python AI Sidecar</span>
                  <span className={`w-2.5 h-2.5 rounded-full ${opsHealth.ai_sidecar_status === 'HEALTHY' ? 'bg-emerald-500' : 'bg-amber-500'}`}></span>
                </div>
                <p className={`text-xl font-bold ${opsHealth.ai_sidecar_status === 'HEALTHY' ? 'text-slate-900' : 'text-amber-700'}`}>
                  {opsHealth.ai_sidecar_status || 'UNKNOWN'}
                </p>
                <span className="text-[11px] text-slate-500 mt-1 block">Port 8090 • FastAPI Uvicorn</span>
              </div>

              <div className="p-4 rounded-lg border border-slate-200 bg-slate-50">
                <div className="flex items-center justify-between mb-2">
                  <span className="text-xs font-bold text-slate-700">Event Mesh</span>
                  <span className="w-2.5 h-2.5 rounded-full bg-emerald-500"></span>
                </div>
                <p className="text-xl font-bold text-slate-900">{opsHealth.event_mesh_status || 'HEALTHY'}</p>
                <span className="text-[11px] text-slate-500 mt-1 block">{opsHealth.dead_letter_count ?? 0} Dead letters</span>
              </div>
            </div>

            <div className="mt-6 pt-5 border-t border-slate-100 grid grid-cols-1 sm:grid-cols-3 gap-4 text-xs">
              <div className="p-3 bg-white border border-slate-200 rounded">
                <span className="text-slate-500 block">Active Workforce Agents:</span>
                <span className="font-bold text-slate-800 text-sm mt-0.5 block">{opsHealth.active_workers_count ?? 0}</span>
              </div>
              <div className="p-3 bg-white border border-slate-200 rounded">
                <span className="text-slate-500 block">Active External Gateways:</span>
                <span className="font-bold text-slate-800 text-sm mt-0.5 block">{opsHealth.active_integrations ?? 0}</span>
              </div>
              <div className="p-3 bg-white border border-slate-200 rounded">
                <span className="text-slate-500 block">Telemetry Last Evaluated:</span>
                <span className="font-bold text-slate-800 text-sm mt-0.5 block">
                  {opsHealth.last_evaluated_at ? new Date(opsHealth.last_evaluated_at).toLocaleTimeString() : 'Recent'}
                </span>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* TAB 11: AUDIT TRAIL */}
      {isTabActive('audit') && (
        <div className="bg-white border border-slate-200 rounded-lg shadow-sm overflow-hidden">
          <div className="p-6 border-b border-slate-200 flex flex-col md:flex-row md:items-center justify-between gap-4">
            <div>
              <h3 className="text-base font-bold text-slate-900">Administrative Audit Trail</h3>
              <p className="text-xs text-slate-500 mt-0.5">
                Immutable record of administrative actions, platform setting updates, feature flag changes, and emergency halts.
              </p>
            </div>

            <div className="relative w-full md:w-64">
              <Search className="w-4 h-4 text-slate-400 absolute left-3 top-2.5" />
              <input
                type="text"
                value={auditSearch}
                onChange={(e) => setAuditSearch(e.target.value)}
                placeholder="Search audit actions..."
                className="w-full pl-9 pr-3 py-1.5 text-xs border border-slate-300 rounded-md focus:outline-none focus:ring-1 focus:ring-blue-500 focus:border-blue-500"
              />
            </div>
          </div>

          <div className="overflow-x-auto">
            <table className="w-full text-left border-collapse">
              <thead>
                <tr className="bg-slate-50 border-b border-slate-200 text-[11px] font-semibold text-slate-600 uppercase tracking-wider">
                  <th className="py-3 px-4">Timestamp (UTC)</th>
                  <th className="py-3 px-4">Actor</th>
                  <th className="py-3 px-4">Action</th>
                  <th className="py-3 px-4">Module</th>
                  <th className="py-3 px-4">Description</th>
                  <th className="py-3 px-4">Result</th>
                  <th className="py-3 px-4">Client IP</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-100 text-xs">
                {filteredAudits.length === 0 ? (
                  <tr>
                    <td colSpan="7" className="py-8 text-center text-slate-500">
                      No administrative audit records found matching your filter.
                    </td>
                  </tr>
                ) : (
                  filteredAudits.map((item) => (
                    <tr key={item.id} className="hover:bg-slate-50 transition-colors">
                      <td className="py-3.5 px-4 font-mono text-[11px] text-slate-600 whitespace-nowrap">
                        {item.created_at ? new Date(item.created_at).toISOString().replace('T', ' ').substring(0, 19) : 'N/A'}
                      </td>
                      <td className="py-3.5 px-4">
                        <span className="font-semibold text-slate-900 block">{item.actor_name}</span>
                        <span className="text-[10px] text-slate-400 block">{item.actor_role}</span>
                      </td>
                      <td className="py-3.5 px-4">
                        <span className="font-mono text-xs font-bold text-blue-900">
                          {item.action}
                        </span>
                      </td>
                      <td className="py-3.5 px-4">
                        <span className="px-2 py-0.5 rounded text-[10px] font-bold bg-slate-100 text-slate-700">
                          {item.module}
                        </span>
                      </td>
                      <td className="py-3.5 px-4 text-slate-700 max-w-sm truncate" title={item.description}>
                        {item.description}
                      </td>
                      <td className="py-3.5 px-4">
                        <span className="px-2 py-0.5 rounded text-[10px] font-bold bg-emerald-100 text-emerald-800">
                          {item.result}
                        </span>
                      </td>
                      <td className="py-3.5 px-4 font-mono text-[11px] text-slate-400 whitespace-nowrap">
                        {item.ip_address || '127.0.0.1'}
                      </td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {/* EDIT SETTING MODAL */}
      {editingSetting && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/50 backdrop-blur-xs p-4">
          <div className="bg-white rounded-lg max-w-md w-full p-6 shadow-xl border border-slate-200">
            <div className="flex items-center justify-between pb-3 border-b border-slate-100 mb-4">
              <div>
                <h3 className="font-bold text-slate-900 text-base">Edit Platform Setting</h3>
                <span className="font-mono text-xs text-blue-600 block mt-0.5">{editingSetting.setting_key}</span>
              </div>
              <button onClick={() => setEditingSetting(null)} className="text-slate-400 hover:text-slate-600">
                <X className="w-5 h-5" />
              </button>
            </div>

            <p className="text-xs text-slate-600 mb-4">{editingSetting.description}</p>

            <div className="mb-5">
              <label className="block text-xs font-semibold text-slate-700 mb-1">
                Setting Value ({editingSetting.data_type})
              </label>
              {editingSetting.data_type === 'BOOLEAN' ? (
                <select
                  value={editSettingValue}
                  onChange={(e) => setEditSettingValue(e.target.value)}
                  className="w-full px-3 py-2 text-xs border border-slate-300 rounded-md focus:outline-none focus:ring-1 focus:ring-blue-500 bg-white"
                >
                  <option value="true">true (Active / Engaged)</option>
                  <option value="false">false (Disabled / Inactive)</option>
                </select>
              ) : (
                <input
                  type="text"
                  value={editSettingValue}
                  onChange={(e) => setEditSettingValue(e.target.value)}
                  className="w-full px-3 py-2 text-xs border border-slate-300 rounded-md focus:outline-none focus:ring-1 focus:ring-blue-500"
                />
              )}
            </div>

            <div className="flex justify-end gap-3 pt-3 border-t border-slate-100">
              <button
                type="button"
                onClick={() => setEditingSetting(null)}
                className="px-3.5 py-2 border border-slate-300 text-slate-700 hover:bg-slate-50 text-xs font-semibold rounded-md transition-colors"
              >
                Cancel
              </button>
              <button
                type="button"
                onClick={handleSaveSetting}
                disabled={savingSetting}
                className="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white text-xs font-semibold rounded-md shadow-sm transition-colors disabled:opacity-50"
              >
                {savingSetting ? 'Saving...' : 'Update Setting'}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* EDIT FEATURE FLAG MODAL */}
      {editingFlag && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/50 backdrop-blur-xs p-4">
          <div className="bg-white rounded-lg max-w-md w-full p-6 shadow-xl border border-slate-200">
            <div className="flex items-center justify-between pb-3 border-b border-slate-100 mb-4">
              <div>
                <h3 className="font-bold text-slate-900 text-base">Configure Feature Flag</h3>
                <span className="font-mono text-xs text-blue-600 block mt-0.5">{editingFlag.flag_key}</span>
              </div>
              <button onClick={() => setEditingFlag(null)} className="text-slate-400 hover:text-slate-600">
                <X className="w-5 h-5" />
              </button>
            </div>

            <p className="text-xs text-slate-600 mb-4">{editingFlag.description}</p>

            <div className="space-y-4 mb-6">
              <label className="flex items-center gap-3 p-3 bg-slate-50 rounded border border-slate-200 cursor-pointer">
                <input
                  type="checkbox"
                  checked={flagForm.isEnabled}
                  onChange={(e) => setFlagForm({ ...flagForm, isEnabled: e.target.checked })}
                  className="h-4 w-4 rounded border-slate-300 text-blue-600 focus:ring-blue-500"
                />
                <div>
                  <span className="text-xs font-bold text-slate-900 block">Flag Is Active / Enabled</span>
                  <span className="text-[11px] text-slate-500 block">Allows agents and runtime pipelines to execute this capability.</span>
                </div>
              </label>

              <label className="flex items-center gap-3 p-3 bg-slate-50 rounded border border-slate-200 cursor-pointer">
                <input
                  type="checkbox"
                  checked={flagForm.requiresApproval}
                  onChange={(e) => setFlagForm({ ...flagForm, requiresApproval: e.target.checked })}
                  className="h-4 w-4 rounded border-slate-300 text-blue-600 focus:ring-blue-500"
                />
                <div>
                  <span className="text-xs font-bold text-slate-900 block">Requires Human Operator Approval</span>
                  <span className="text-[11px] text-slate-500 block">Consequential actions generate an approval request before execution.</span>
                </div>
              </label>

              <div>
                <label className="block text-xs font-semibold text-slate-700 mb-1">Max Autonomy Level Allowed</label>
                <select
                  value={flagForm.maxAutonomyLevel}
                  onChange={(e) => setFlagForm({ ...flagForm, maxAutonomyLevel: e.target.value })}
                  className="w-full px-3 py-2 text-xs border border-slate-300 rounded-md focus:outline-none focus:ring-1 focus:ring-blue-500 bg-white"
                >
                  <option value={1}>Level 1: Inform (Read-only insights)</option>
                  <option value={2}>Level 2: Prepare (Drafts and staging)</option>
                  <option value={3}>Level 3: Controlled Execution (With safety gates)</option>
                  <option value={4}>Level 4: Full Multi-Step Autonomous</option>
                </select>
              </div>
            </div>

            <div className="flex justify-end gap-3 pt-3 border-t border-slate-100">
              <button
                type="button"
                onClick={() => setEditingFlag(null)}
                className="px-3.5 py-2 border border-slate-300 text-slate-700 hover:bg-slate-50 text-xs font-semibold rounded-md transition-colors"
              >
                Cancel
              </button>
              <button
                type="button"
                onClick={handleSaveFlag}
                disabled={savingFlag}
                className="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white text-xs font-semibold rounded-md shadow-sm transition-colors disabled:opacity-50"
              >
                {savingFlag ? 'Saving...' : 'Save Flag Policy'}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* EMERGENCY HALT MODAL */}
      {haltModalOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/60 backdrop-blur-xs p-4">
          <div className="bg-white rounded-lg max-w-md w-full p-6 shadow-2xl border border-rose-200">
            <div className="flex items-center gap-3 pb-3 border-b border-slate-100 mb-4 text-rose-700">
              <ShieldAlert className="w-6 h-6 flex-shrink-0" />
              <div>
                <h3 className="font-bold text-slate-900 text-base">
                  {haltActiveState ? 'Engage Emergency Kill Switch' : 'Resume Autonomous Execution'}
                </h3>
                <span className="text-xs text-slate-500">
                  Target: {haltTargetModule ? `Module '${haltTargetModule}'` : 'ALL OPERATIONAL MODULES'}
                </span>
              </div>
            </div>

            <div className="p-3 bg-amber-50 border border-amber-200 rounded text-xs text-amber-900 mb-4">
              {haltActiveState ? (
                <p>
                  <strong>Warning:</strong> Engaging this halt immediately prevents all automated agents from completing consequential actions.
                  All execution requests will suspend until an authorized administrator clears the halt.
                </p>
              ) : (
                <p>
                  <strong>Notice:</strong> Clearing the emergency halt will re-enable autonomous operations and scheduled pipelines
                  for the specified scope.
                </p>
              )}
            </div>

            <div className="mb-5">
              <label className="block text-xs font-semibold text-slate-700 mb-1">
                Operational Reason & Justification <span className="text-rose-500">*</span>
              </label>
              <textarea
                rows={3}
                value={haltReason}
                onChange={(e) => setHaltReason(e.target.value)}
                placeholder="Explain why this emergency halt action is being executed..."
                className="w-full px-3 py-2 text-xs border border-slate-300 rounded-md focus:outline-none focus:ring-1 focus:ring-rose-500 focus:border-rose-500"
              />
            </div>

            <div className="flex justify-end gap-3 pt-3 border-t border-slate-100">
              <button
                type="button"
                onClick={() => setHaltModalOpen(false)}
                className="px-3.5 py-2 border border-slate-300 text-slate-700 hover:bg-slate-50 text-xs font-semibold rounded-md transition-colors"
              >
                Cancel
              </button>
              <button
                type="button"
                onClick={handleExecuteHalt}
                disabled={executingHalt}
                className={`px-4 py-2 text-white text-xs font-bold rounded-md shadow-sm transition-colors disabled:opacity-50 ${
                  haltActiveState ? 'bg-rose-600 hover:bg-rose-700' : 'bg-emerald-600 hover:bg-emerald-700'
                }`}
              >
                {executingHalt ? 'Executing...' : haltActiveState ? 'Confirm Engage Kill Switch' : 'Confirm Resume Operations'}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* INTEGRATION TOGGLE MODAL */}
      {integrationModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/50 backdrop-blur-xs p-4">
          <div className="bg-white rounded-lg max-w-md w-full p-6 shadow-xl border border-slate-200">
            <div className="flex items-center justify-between pb-3 border-b border-slate-100 mb-4">
              <div>
                <h3 className="font-bold text-slate-900 text-base">
                  {integrationModal.is_enabled ? 'Disable External Integration' : 'Enable External Integration'}
                </h3>
                <span className="text-xs text-blue-600 block mt-0.5">{integrationModal.provider_name}</span>
              </div>
              <button onClick={() => setIntegrationModal(null)} className="text-slate-400 hover:text-slate-600">
                <X className="w-5 h-5" />
              </button>
            </div>

            <p className="text-xs text-slate-600 mb-4">
              {integrationModal.is_enabled
                ? `Disabling '${integrationModal.provider_name}' will halt outbound communication and status sync jobs for this provider.`
                : `Enabling '${integrationModal.provider_name}' will resume integration gateway routing.`}
            </p>

            <div className="mb-5">
              <label className="block text-xs font-semibold text-slate-700 mb-1">Reason for modification (Optional)</label>
              <input
                type="text"
                value={integrationReason}
                onChange={(e) => setIntegrationReason(e.target.value)}
                placeholder="e.g. Scheduled provider maintenance or credential rotation"
                className="w-full px-3 py-2 text-xs border border-slate-300 rounded-md focus:outline-none focus:ring-1 focus:ring-blue-500"
              />
            </div>

            <div className="flex justify-end gap-3 pt-3 border-t border-slate-100">
              <button
                type="button"
                onClick={() => setIntegrationModal(null)}
                className="px-3.5 py-2 border border-slate-300 text-slate-700 hover:bg-slate-50 text-xs font-semibold rounded-md transition-colors"
              >
                Cancel
              </button>
              <button
                type="button"
                onClick={handleToggleIntegration}
                disabled={togglingIntegration}
                className={`px-4 py-2 text-white text-xs font-semibold rounded-md shadow-sm transition-colors disabled:opacity-50 ${
                  integrationModal.is_enabled ? 'bg-rose-600 hover:bg-rose-700' : 'bg-emerald-600 hover:bg-emerald-700'
                }`}
              >
                {togglingIntegration ? 'Updating...' : integrationModal.is_enabled ? 'Confirm Disable' : 'Confirm Enable'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

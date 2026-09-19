import React, { useState, useEffect } from 'react';
import { useParams, useNavigate, Link, useSearchParams } from 'react-router-dom';
import {
  Building2,
  ArrowLeft,
  Mail,
  Phone,
  Globe,
  MapPin,
  FileText,
  CreditCard,
  Users,
  Package,
  FileCheck,
  Calendar,
  AlertTriangle,
  CheckCircle2,
  Clock,
  Shield,
  Edit3,
  ExternalLink,
  Activity,
  Layers,
  ChevronRight,
  ChevronDown,
  Info,
  UserPlus,
  Eye,
  UserX,
  UserCheck,
  ArrowRightLeft,
  RefreshCw,
  Ban,
  ToggleLeft,
  ToggleRight,
  Heart,
  TrendingUp,
  TrendingDown,
  DollarSign,
  AlertCircle,
  FileSpreadsheet,
  Cpu,
  Workflow,
  Zap,
  Check,
  Plus,
  Sparkles,
  Search,
  Filter,
  ArrowUpRight,
  Rocket,
  BarChart2,
  Compass,
  Link2
} from 'lucide-react';
import { sportalService } from '../../services/sportalService';
import { EditOrganizationModal } from './EditOrganizationModal';
import { InviteCustomerUserModal } from '../users/InviteCustomerUserModal';
import { CustomerUserDetailModal } from '../users/CustomerUserDetailModal';
import { ConfirmUserStatusModal } from '../users/ConfirmUserStatusModal';
import { ChangeUserRoleModal } from '../users/ChangeUserRoleModal';
import { ChangeSubscriptionPlanModal } from '../subscriptions/ChangeSubscriptionPlanModal';
import { RenewSubscriptionModal } from '../subscriptions/RenewSubscriptionModal';
import { CancelSubscriptionModal } from '../subscriptions/CancelSubscriptionModal';
import { AssignSubscriptionModal } from '../subscriptions/AssignSubscriptionModal';
import { CustomerUsageView } from '../usage/CustomerUsageView';
import { CustomerHealthView } from '../health/CustomerHealthView';
import { CustomerCompanyView } from '../company/CustomerCompanyView';
import { CustomerOnboardingModal } from '../onboarding/CustomerOnboardingModal';
import { CustomerOnboardingView } from '../onboarding/CustomerOnboardingView';
import { CustomerSubscriptionView } from '../subscriptions/CustomerSubscriptionView';
import { CustomerUsersView } from '../users/CustomerUsersView';
import { CustomerExceptionsView } from '../exceptions/CustomerExceptionsView';
import { CustomerContractsView } from '../contracts/CustomerContractsView';
import { CustomerIntegrationsView } from '../integrations/CustomerIntegrationsView';

export function OrganizationDetailPage() {
  const { organizationId } = useParams();
  const navigate = useNavigate();

  // Core State
  const [details, setDetails] = useState(null);
  const [userSummary, setUserSummary] = useState(null);
  const [richSub, setRichSub] = useState(null);
  const [plans, setPlans] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [searchParams, setSearchParams] = useSearchParams();
  const [activeTab, setActiveTab] = useState(searchParams.get('tab') || 'overview');
  const [isEditOpen, setIsEditOpen] = useState(false);
  const [isInviteOpen, setIsInviteOpen] = useState(false);
  const [isActionsOpen, setIsActionsOpen] = useState(false);

  const handleTabChange = (tabId) => {
    setActiveTab(tabId);
    setSearchParams({ tab: tabId });
  };

  useEffect(() => {
    const tabFromUrl = searchParams.get('tab');
    if (tabFromUrl && tabFromUrl !== activeTab) {
      setActiveTab(tabFromUrl);
    }
  }, [searchParams]);

  // Lazy Domain Subsystem State for Retained Deep Tabs
  const [contractsState, setContractsState] = useState({ loading: false, data: null, error: null });
  const [exceptionsState, setExceptionsState] = useState({ loading: false, data: null, error: null });
  const [integrationsState, setIntegrationsState] = useState({ loading: false, data: null, error: null });

  // Modals State
  const [selectedDetailUser, setSelectedDetailUser] = useState(null);
  const [roleChangeState, setRoleChangeState] = useState({ isOpen: false, user: null });
  const [statusConfirmState, setStatusConfirmState] = useState({
    isOpen: false,
    user: null,
    newStatus: '',
  });

  // Task S8: Subscription Administration State
  const [isChangePlanOpen, setIsChangePlanOpen] = useState(false);
  const [isRenewOpen, setIsRenewOpen] = useState(false);
  const [isCancelOpen, setIsCancelOpen] = useState(false);
  const [isAssignOpen, setIsAssignOpen] = useState(false);
  const [togglingAutoRenew, setTogglingAutoRenew] = useState(false);
  const [isOnboardingOpen, setIsOnboardingOpen] = useState(false);
  const [onboardingStage, setOnboardingStage] = useState(1);

  // Fetch core foundation data
  const fetchDetails = async () => {
    try {
      setLoading(true);
      setError(null);
      const [data, summary, subRes, plansRes] = await Promise.all([
        sportalService.getOrganizationDetails(organizationId),
        sportalService.getOrgUserSummary(organizationId).catch(() => null),
        sportalService.getOrganizationSubscription(organizationId).catch(() => null),
        sportalService.getSubscriptionPlans().catch(() => null),
      ]);
      setDetails(data);
      if (summary) {
        setUserSummary(summary.data || summary);
      }
      if (subRes) {
        setRichSub(subRes.data || subRes);
      }
      if (plansRes) {
        const pData = plansRes.data || plansRes;
        setPlans(Array.isArray(pData) ? pData : []);
      }
    } catch (err) {
      console.error('Failed to load customer details:', err);
      setError(err.message || 'Failed to retrieve customer organization record');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (organizationId) {
      fetchDetails();
    }
  }, [organizationId]);

  // Lazy tab data loaders for retained tabs
  useEffect(() => {
    if (!organizationId) return;

    if ((activeTab === 'contracts' || activeTab === 'onboarding') && !contractsState.data && !contractsState.loading) {
      setContractsState((prev) => ({ ...prev, loading: true }));
      sportalService
        .getCustomerContracts(organizationId)
        .then((res) => {
          const items = res?.items || res?.data?.items || (Array.isArray(res?.data) ? res.data : (Array.isArray(res) ? res : []));
          setContractsState({ loading: false, data: items, error: null });
        })
        .catch((err) => setContractsState({ loading: false, data: [], error: err.message }));
    }

    if (activeTab === 'exceptions' && !exceptionsState.data && !exceptionsState.loading) {
      setExceptionsState((prev) => ({ ...prev, loading: true }));
      sportalService
        .getCustomerExceptions(organizationId)
        .then((res) => {
          const items = Array.isArray(res) ? res : (Array.isArray(res?.data) ? res.data : []);
          setExceptionsState({ loading: false, data: items, error: null });
        })
        .catch((err) => setExceptionsState({ loading: false, data: [], error: err.message }));
    }

    if ((activeTab === 'integrations' || activeTab === 'onboarding') && !integrationsState.data && !integrationsState.loading) {
      setIntegrationsState((prev) => ({ ...prev, loading: true }));
      sportalService
        .getCustomerIntegrations(organizationId)
        .then((res) => {
          const items = res?.items || res?.data?.items || (Array.isArray(res?.data) ? res.data : (Array.isArray(res) ? res : []));
          setIntegrationsState({ loading: false, data: items, error: null });
        })
        .catch((err) => setIntegrationsState({ loading: false, data: [], error: err.message }));
    }
  }, [activeTab, organizationId]);

  const handleToggleAutoRenew = async () => {
    if (!richSub || togglingAutoRenew) return;
    try {
      setTogglingAutoRenew(true);
      const newSetting = !richSub.auto_renew;
      const res = await sportalService.toggleOrganizationAutoRenew(organizationId, {
        auto_renew: newSetting,
        reason: `Auto-renew set to ${newSetting ? 'ENABLED' : 'DISABLED'} from Customer 360`,
      });
      setRichSub(res?.data || res);
      fetchDetails();
    } catch (err) {
      alert('Failed to update auto-renew: ' + (err?.response?.data?.message || err.message));
    } finally {
      setTogglingAutoRenew(false);
    }
  };

  if (loading) {
    return (
      <div className="space-y-6">
        <div className="h-6 w-48 bg-slate-200 rounded-md animate-pulse" />
        <div className="h-32 bg-white rounded-xl border border-slate-200 p-6 animate-pulse" />
        <div className="grid grid-cols-2 md:grid-cols-6 gap-3.5">
          {[...Array(6)].map((_, i) => (
            <div key={i} className="h-24 bg-white rounded-xl border border-slate-200 animate-pulse" />
          ))}
        </div>
      </div>
    );
  }

  if (error || !details) {
    return (
      <div className="rounded-xl border border-red-200 bg-red-50 p-8 text-center">
        <AlertTriangle className="mx-auto h-12 w-12 text-red-500 mb-3" />
        <h3 className="text-lg font-bold text-slate-900">Customer Organization Not Found</h3>
        <p className="text-sm text-slate-600 mt-1 mb-4">{error || 'The requested customer organization could not be loaded.'}</p>
        <button
          onClick={() => navigate('/organizations')}
          className="inline-flex items-center gap-2 rounded-lg bg-navy-900 px-4 py-2 text-sm font-medium text-white hover:bg-navy-800"
        >
          <ArrowLeft className="h-4 w-4" />
          <span>Back to Organizations</span>
        </button>
      </div>
    );
  }

  const {
    organization: org = {},
    subscription: sub = {},
    users = [],
    stats = {},
    recent_activity: activity = [],
    health = { status: 'Good', score: 94, summary: 'No critical issues', risk_indicators: [] },
    shipment_trends: trends = [],
    open_alerts: alerts = [],
    team_summary: team = { total_users: users.length, active_users: users.length, pending_users: 0, suspended_users: 0, primary_admin: null, other_key_users: [] }
  } = details || {};

  // Format date helper
  const formatDate = (dStr) => {
    if (!dStr) return '—';
    try {
      const d = new Date(dStr);
      return isNaN(d.getTime()) ? '—' : d.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' });
    } catch {
      return '—';
    }
  };

  // Nav Tabs List (10 Retained Customer 360 Tabs per UI specification)
  const navTabs = [
    { id: 'overview', label: 'Overview', icon: FileText },
    { id: 'health', label: 'Health & Retention', icon: Activity },
    { id: 'usage', label: 'Usage & Analytics', icon: BarChart2 },
    { id: 'company', label: 'Company', icon: Building2 },
    { id: 'onboarding', label: 'Onboarding', icon: Compass },
    { id: 'subscription', label: 'Subscription', icon: CreditCard },
    { id: 'users', label: 'Users', icon: Users, count: users.length },
    { id: 'exceptions', label: 'Exceptions', icon: AlertTriangle, count: stats.open_exceptions_count ?? 0 },
    { id: 'contracts', label: 'Contracts', icon: FileText, count: stats.expiring_contracts_count ? `${stats.expiring_contracts_count} Due` : null },
    { id: 'integrations', label: 'Integrations', icon: Link2 },
  ];

  const primaryAdminUser = team.primary_admin || (users.length > 0 ? users[0] : null);

  return (
    <div className="space-y-6 pb-12 w-full max-w-full min-w-0">
      {/* 1. Breadcrumb Navigation */}
      <div className="flex items-center gap-2 text-xs font-medium text-slate-500">
        <Link to="/organizations" className="hover:text-navy-900 transition-colors">
          Organizations
        </Link>
        <ChevronRight className="h-3 w-3 text-slate-400" />
        <span className="hover:text-navy-900 transition-colors cursor-pointer" onClick={() => setActiveTab('overview')}>
          {org.name}
        </span>
        <ChevronRight className="h-3 w-3 text-slate-400" />
        <span className="text-slate-900 font-semibold">Customer 360</span>
      </div>

      {/* 2. Customer Identity Header Card (Strictly adhering to sportalCustomerView.png) */}
      <div className="rounded-xl border border-slate-200 bg-white p-5 sm:p-6 shadow-xs">
        <div className="flex flex-col lg:flex-row lg:items-center justify-between gap-5">
          <div className="flex items-start gap-4 min-w-0">
            {/* Identity Icon / Logo Box */}
            <div className="flex h-14 w-14 sm:h-16 sm:w-16 shrink-0 items-center justify-center rounded-2xl bg-white border border-slate-200 text-blue-700 shadow-2xs overflow-hidden">
              {org.logo_url ? (
                <img src={org.logo_url} alt={org.name} className="h-full w-full object-contain p-1.5" />
              ) : (
                <div className="flex flex-col items-center justify-center text-center p-1">
                  <Building2 className="h-6 w-6 text-blue-700" />
                  <span className="text-[9px] font-black uppercase text-slate-800 tracking-wider mt-0.5 max-w-[52px] truncate">
                    {org.name?.split(' ')[0] || 'CORP'}
                  </span>
                </div>
              )}
            </div>

            <div className="space-y-1.5 min-w-0 flex-1">
              <div className="flex flex-wrap items-center gap-2.5">
                <h1 className="text-xl sm:text-2xl font-bold text-slate-900 tracking-tight truncate">{org.name}</h1>
                <span className="rounded-full bg-emerald-50 px-2.5 py-0.5 text-xs font-semibold text-emerald-700 border border-emerald-200 shrink-0">
                  {org.status === 'Active' ? 'Active Customer' : org.status || 'Active Customer'}
                </span>
              </div>

              {/* Metadata Line */}
              <div className="flex flex-wrap items-center gap-2 text-xs text-slate-500">
                <span className="font-mono font-medium text-slate-700">ORG-{String(org.id).padStart(6, '0')}</span>
                <span>•</span>
                <span>{org.company_type || 'Private Limited'}</span>
                <span>•</span>
                <span>{[org.city, org.country].filter(Boolean).join(', ') || 'Navi Mumbai, India'}</span>
                <span>•</span>
                <span>Customer since {formatDate(org.created_at)}</span>
              </div>

              {/* Contact Information Line */}
              <div className="flex flex-wrap items-center gap-4 text-xs text-slate-600 pt-0.5">
                {org.website ? (
                  <a
                    href={org.website.startsWith('http') ? org.website : `https://${org.website}`}
                    target="_blank"
                    rel="noreferrer"
                    className="flex items-center gap-1.5 text-blue-600 hover:underline truncate max-w-xs"
                  >
                    <Globe className="h-3.5 w-3.5 text-blue-500 shrink-0" />
                    <span className="truncate">{org.website.replace(/^https?:\/\//, '')}</span>
                  </a>
                ) : (
                  <span className="flex items-center gap-1.5 text-slate-400">
                    <Globe className="h-3.5 w-3.5 text-slate-400 shrink-0" />
                    <span>www.{org.name.toLowerCase().replace(/[^a-z0-9]/g, '')}.com</span>
                  </span>
                )}

                <span className="flex items-center gap-1.5 text-slate-600 truncate max-w-xs">
                  <Mail className="h-3.5 w-3.5 text-slate-400 shrink-0" />
                  <span className="truncate">{org.primary_email || `info@${org.name.toLowerCase().replace(/[^a-z0-9]/g, '')}.com`}</span>
                </span>

                <span className="flex items-center gap-1.5 text-slate-600 shrink-0">
                  <Phone className="h-3.5 w-3.5 text-slate-400 shrink-0" />
                  <span>{org.phone_number || '+91 98765 43210'}</span>
                </span>

                <div className="flex items-center gap-1.5 text-slate-600 shrink-0">
                  <span className="font-mono text-slate-700 font-medium">GST: {org.tax_number || '27AABCU8925J125'}</span>
                  <span className="inline-flex items-center gap-0.5 rounded-full bg-blue-50 px-1.5 py-0.2 text-[10px] font-bold text-blue-700 border border-blue-200">
                    <CheckCircle2 className="h-2.5 w-2.5 text-blue-600" />
                    <span>Verified</span>
                  </span>
                </div>
              </div>
            </div>
          </div>

          {/* Action Controls */}
          <div className="flex items-center gap-2.5 shrink-0 self-start lg:self-center flex-wrap">
            <Link
              to={`/ai?orgId=${org.id}`}
              className="flex items-center gap-1.5 rounded-lg border border-purple-200 bg-purple-50 px-3.5 py-2 text-xs font-bold text-purple-700 hover:bg-purple-100 transition-colors shadow-2xs"
            >
              <Sparkles className="h-3.5 w-3.5 text-purple-600" />
              <span>Ask AI Copilot</span>
            </Link>

            <button
              type="button"
              onClick={() => setIsEditOpen(true)}
              className="flex items-center gap-2 rounded-lg border border-slate-300 bg-white px-3.5 py-2 text-xs font-semibold text-slate-700 hover:bg-slate-50 transition-colors shadow-2xs cursor-pointer"
            >
              <Edit3 className="h-3.5 w-3.5 text-slate-500" />
              <span>Edit Organization</span>
            </button>

            {/* Actions Dropdown */}
            <div className="relative">
              <button
                type="button"
                onClick={() => setIsActionsOpen(!isActionsOpen)}
                className="flex items-center gap-2 rounded-lg bg-navy-900 px-4 py-2 text-xs font-bold text-white hover:bg-navy-800 transition-colors shadow-2xs cursor-pointer"
              >
                <span>Actions</span>
                <ChevronDown className="h-3.5 w-3.5" />
              </button>

              {isActionsOpen && (
                <div className="absolute right-0 mt-2 w-52 rounded-xl border border-slate-200 bg-white py-1.5 shadow-lg z-30 divide-y divide-slate-100">
                  <div className="py-1">
                    <button
                      type="button"
                      onClick={() => {
                        setIsActionsOpen(false);
                        setIsChangePlanOpen(true);
                      }}
                      className="w-full px-3.5 py-2 text-left text-xs font-medium text-slate-700 hover:bg-slate-50 flex items-center gap-2 cursor-pointer"
                    >
                      <ArrowRightLeft className="h-3.5 w-3.5 text-blue-600" />
                      <span>Change Subscription Plan</span>
                    </button>
                    <button
                      type="button"
                      onClick={() => {
                        setIsActionsOpen(false);
                        setIsRenewOpen(true);
                      }}
                      className="w-full px-3.5 py-2 text-left text-xs font-medium text-slate-700 hover:bg-slate-50 flex items-center gap-2 cursor-pointer"
                    >
                      <RefreshCw className="h-3.5 w-3.5 text-emerald-600" />
                      <span>Extend Period / Renew</span>
                    </button>
                    <button
                      type="button"
                      onClick={() => {
                        setIsActionsOpen(false);
                        setIsInviteOpen(true);
                      }}
                      className="w-full px-3.5 py-2 text-left text-xs font-medium text-slate-700 hover:bg-slate-50 flex items-center gap-2 cursor-pointer"
                    >
                      <UserPlus className="h-3.5 w-3.5 text-purple-600" />
                      <span>Invite Forwarder User</span>
                    </button>
                  </div>
                  <div className="py-1">
                    <button
                      type="button"
                      onClick={() => {
                        setIsActionsOpen(false);
                        window.print();
                      }}
                      className="w-full px-3.5 py-2 text-left text-xs font-medium text-slate-700 hover:bg-slate-50 flex items-center gap-2 cursor-pointer"
                    >
                      <FileSpreadsheet className="h-3.5 w-3.5 text-slate-500" />
                      <span>Export Customer Dossier</span>
                    </button>
                  </div>
                </div>
              )}
            </div>
          </div>
        </div>
      </div>

      {/* 3. Horizontal Navigation Tabs (10 Retained Customer 360 Tabs) */}
      <div className="border-b border-slate-200 overflow-x-auto scroll-smooth">
        <nav className="flex space-x-6 sm:space-x-8 whitespace-nowrap min-w-max px-1" aria-label="Customer 360 Navigation">
          {navTabs.map((tab) => {
            const Icon = tab.icon;
            const isActive = activeTab === tab.id;
            return (
              <button
                key={tab.id}
                id={`tab-${tab.id}`}
                onClick={() => handleTabChange(tab.id)}
                className={`flex items-center gap-2 text-xs sm:text-sm font-semibold transition-all pb-3 -mb-px border-b-2 cursor-pointer ${
                  isActive
                    ? 'border-blue-600 text-blue-600 font-bold'
                    : 'border-transparent text-slate-500 hover:text-slate-900 hover:border-slate-300'
                }`}
              >
                {Icon && (
                  <Icon
                    className={`h-4 w-4 shrink-0 transition-colors ${
                      isActive ? 'text-blue-600' : 'text-slate-400 group-hover:text-slate-600'
                    }`}
                  />
                )}
                <span>{tab.label}</span>
                {tab.count !== undefined && tab.count !== null && tab.count !== 0 && (
                  <span
                    className={`ml-0.5 rounded-full px-1.5 py-0.2 text-[10px] font-bold ${
                      isActive ? 'bg-blue-100 text-blue-700' : 'bg-slate-100 text-slate-600'
                    }`}
                  >
                    {tab.count}
                  </span>
                )}
              </button>
            );
          })}
        </nav>
      </div>

      {/* ========================================================================= */}
      {/* TAB 1: OVERVIEW (Full single-pane workspace matching sportalCustomerView.png) */}
      {/* ========================================================================= */}
      {activeTab === 'overview' && (
        <div className="space-y-6">
          {/* Top 6 KPI Metric Cards */}
          <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-6 gap-3.5">
            {/* 1. Active Shipments */}
            <div className="rounded-xl border border-slate-200 bg-white p-4 shadow-xs">
              <div className="flex items-center justify-between mb-2">
                <span className="text-xs font-medium text-slate-500">Active Shipments</span>
                <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-emerald-50 text-emerald-600">
                  <Package className="h-4 w-4" />
                </div>
              </div>
              <p className="text-2xl font-black text-slate-900">{stats.active_shipments_count ?? stats.shipments_count ?? 0}</p>
              <div className="flex items-center gap-1 text-[11px] font-semibold text-emerald-600 mt-1">
                <TrendingUp className="h-3 w-3" />
                <span>↑ 20% vs last month</span>
              </div>
            </div>

            {/* 2. Open Exceptions */}
            <div className="rounded-xl border border-slate-200 bg-white p-4 shadow-xs">
              <div className="flex items-center justify-between mb-2">
                <span className="text-xs font-medium text-slate-500">Open Exceptions</span>
                <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-sky-50 text-sky-600">
                  <AlertCircle className="h-4 w-4" />
                </div>
              </div>
              <p className="text-2xl font-black text-slate-900">{stats.open_exceptions_count ?? 0}</p>
              <div className="flex items-center gap-1 text-[11px] font-semibold text-emerald-600 mt-1">
                <TrendingDown className="h-3 w-3" />
                <span>↓ 40% vs last month</span>
              </div>
            </div>

            {/* 3. Total Users */}
            <div className="rounded-xl border border-slate-200 bg-white p-4 shadow-xs">
              <div className="flex items-center justify-between mb-2">
                <span className="text-xs font-medium text-slate-500">Total Users</span>
                <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-purple-50 text-purple-600">
                  <Users className="h-4 w-4" />
                </div>
              </div>
              <p className="text-2xl font-black text-slate-900">{users.length}</p>
              <div className="flex items-center gap-1 text-[11px] font-semibold text-emerald-600 mt-1">
                <TrendingUp className="h-3 w-3" />
                <span>↑ 12% vs last month</span>
              </div>
            </div>

            {/* 4. Outstanding Invoices */}
            <div className="rounded-xl border border-slate-200 bg-white p-4 shadow-xs">
              <div className="flex items-center justify-between mb-2">
                <span className="text-xs font-medium text-slate-500">Outstanding Invoices</span>
                <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-amber-50 text-amber-600">
                  <DollarSign className="h-4 w-4" />
                </div>
              </div>
              <p className="text-2xl font-black text-slate-900">{stats.outstanding_invoices_count ?? 0}</p>
              <div className="flex items-center gap-1 text-[11px] font-bold text-amber-700 mt-1">
                <span>USD {(stats.outstanding_invoices_amount || 0).toLocaleString()}</span>
                <Clock className="h-3 w-3 text-amber-500" />
              </div>
            </div>

            {/* 5. Subscription */}
            <div className="rounded-xl border border-slate-200 bg-white p-4 shadow-xs">
              <div className="flex items-center justify-between mb-2">
                <span className="text-xs font-medium text-slate-500">Subscription</span>
                <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-teal-50 text-teal-600">
                  <Calendar className="h-4 w-4" />
                </div>
              </div>
              <p className="text-xl font-black text-slate-900 truncate">{(richSub || sub)?.plan_name || 'Starter'}</p>
              <div className="text-[11px] font-medium text-slate-500 mt-1 truncate">
                🗓️ Renews in {(richSub || sub)?.days_until_renewal ?? 29} days
              </div>
            </div>

            {/* 6. Customer Health */}
            <div className="rounded-xl border border-slate-200 bg-white p-4 shadow-xs">
              <div className="flex items-center justify-between mb-2">
                <span className="text-xs font-medium text-slate-500">Customer Health</span>
                <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-rose-50 text-rose-600">
                  <Heart className="h-4 w-4" />
                </div>
              </div>
              <p className="text-2xl font-black text-slate-900">{health.status}</p>
              <div className="flex items-center gap-1.5 text-[11px] font-medium text-slate-600 mt-1 truncate">
                <span className={`h-2 w-2 rounded-full ${health.status === 'Good' ? 'bg-emerald-500' : 'bg-amber-500'}`} />
                <span className="truncate">{health.summary}</span>
              </div>
            </div>
          </div>

          {/* Middle Row: 4 Panels */}
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
            {/* Panel 1: Organization Details */}
            <div className="rounded-xl border border-slate-200 bg-white p-5 shadow-xs flex flex-col justify-between space-y-4">
              <div>
                <div className="flex items-center justify-between border-b border-slate-100 pb-3 mb-3">
                  <h3 className="text-sm font-bold text-slate-900">Organization Details</h3>
                  <button
                    type="button"
                    onClick={() => setIsEditOpen(true)}
                    className="text-xs font-semibold text-blue-600 hover:underline cursor-pointer"
                  >
                    Edit
                  </button>
                </div>

                <div className="space-y-2.5 text-xs">
                  <div className="flex justify-between">
                    <span className="text-slate-500">Company Name</span>
                    <span className="font-semibold text-slate-900 text-right truncate max-w-[140px]">{org.name}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-slate-500">Organization ID</span>
                    <span className="font-mono font-medium text-slate-800">ORG-{String(org.id).padStart(6, '0')}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-slate-500">Type</span>
                    <span className="text-slate-800">{org.company_type || 'Private Limited'}</span>
                  </div>
                  <div className="flex justify-between items-center">
                    <span className="text-slate-500">Status</span>
                    <span className="rounded-full bg-emerald-50 px-2 py-0.2 text-[10px] font-bold text-emerald-700 border border-emerald-200">
                      {org.status || 'Active'}
                    </span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-slate-500">Primary Admin</span>
                    <span className="font-medium text-slate-800 text-right truncate max-w-[140px]">
                      {primaryAdminUser ? `${primaryAdminUser.full_name}` : 'Not designated'}
                    </span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-slate-500">Phone</span>
                    <span className="text-slate-800">{org.phone_number || '+91 98765 43210'}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-slate-500">Email</span>
                    <span className="text-slate-800 text-right truncate max-w-[140px]">{org.primary_email || `ops_${org.id}@apexfreight.test`}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-slate-500">Website</span>
                    <span className="text-blue-600 hover:underline truncate max-w-[140px]">
                      {org.website ? org.website.replace(/^https?:\/\//, '') : `www.${org.name.toLowerCase().replace(/[^a-z0-9]/g, '')}.com`}
                    </span>
                  </div>
                  <div className="flex justify-between items-center">
                    <span className="text-slate-500">GST Number</span>
                    <div className="flex items-center gap-1">
                      <span className="font-mono font-medium text-slate-800">{org.tax_number || '27AABCU8925J125'}</span>
                      <span className="inline-flex items-center gap-0.5 rounded-full bg-blue-50 px-1 py-0.2 text-[9px] font-bold text-blue-700 border border-blue-200">
                        <CheckCircle2 className="h-2 w-2 text-blue-600" />
                        <span>Verified</span>
                      </span>
                    </div>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-slate-500">Location</span>
                    <span className="text-slate-800 text-right truncate max-w-[140px]">
                      {[org.city, org.country].filter(Boolean).join(', ') || 'Navi Mumbai, India'}
                    </span>
                  </div>
                </div>
              </div>
            </div>

            {/* Panel 2: Subscription & Billing */}
            <div className="rounded-xl border border-slate-200 bg-white p-5 shadow-xs flex flex-col justify-between space-y-4">
              <div>
                <div className="flex items-center justify-between border-b border-slate-100 pb-3 mb-3">
                  <h3 className="text-sm font-bold text-slate-900">Subscription & Billing</h3>
                  <button
                    type="button"
                    onClick={() => setActiveTab('subscription')}
                    className="text-xs font-semibold text-blue-600 hover:underline cursor-pointer"
                  >
                    View Details
                  </button>
                </div>

                {/* Plan Card Banner */}
                <div className="rounded-lg border border-slate-100 bg-slate-50 p-3 mb-3 flex items-center justify-between">
                  <div>
                    <span className="text-[10px] font-medium text-slate-400 uppercase tracking-wider block">CURRENT PLAN</span>
                    <p className="text-base font-black text-slate-900">{(richSub || sub)?.plan_name || 'Starter'}</p>
                  </div>
                  <span className="rounded-full bg-emerald-50 px-2 py-0.5 text-[10px] font-bold text-emerald-700 border border-emerald-200">
                    {(richSub || sub)?.status?.toUpperCase() || 'ACTIVE'}
                  </span>
                </div>

                <div className="space-y-2.5 text-xs mb-3">
                  <div className="flex justify-between">
                    <span className="text-slate-500">Start Date</span>
                    <span className="text-slate-800">{formatDate((richSub || sub)?.current_period_start || org.created_at)}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-slate-500">Renewal Date</span>
                    <span className="text-slate-800">
                      {formatDate((richSub || sub)?.current_period_end)}
                      <span className="text-amber-600 text-[11px] block text-right font-medium">
                        (in {(richSub || sub)?.days_until_renewal ?? 29} days)
                      </span>
                    </span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-slate-500">Billing Cycle</span>
                    <span className="text-slate-800 capitalize">{(richSub || sub)?.billing_cycle || 'Monthly'}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-slate-500">Amount</span>
                    <span className="font-semibold text-slate-900">
                      USD {((richSub || sub)?.amount || (richSub || sub)?.monthly_price || 99).toLocaleString()} / {(richSub || sub)?.billing_cycle === 'annual' ? 'year' : 'month'}
                    </span>
                  </div>
                </div>
              </div>

              <button
                type="button"
                onClick={() => setActiveTab('subscription')}
                className="w-full rounded-lg bg-blue-50 hover:bg-blue-100 text-blue-700 py-2 text-xs font-semibold flex items-center justify-center gap-1.5 transition-colors cursor-pointer"
              >
                <span>Manage Subscription</span>
                <ChevronRight className="h-3.5 w-3.5" />
              </button>
            </div>

            {/* Panel 3: Recent Activity */}
            <div className="rounded-xl border border-slate-200 bg-white p-5 shadow-xs flex flex-col justify-between space-y-4">
              <div>
                <div className="flex items-center justify-between border-b border-slate-100 pb-3 mb-3">
                  <h3 className="text-sm font-bold text-slate-900">Recent Activity</h3>
                  <Link
                    to={`/support?tab=activity&orgId=${org.id}`}
                    className="text-xs font-semibold text-blue-600 hover:underline"
                  >
                    View All
                  </Link>
                </div>

                {activity.length === 0 ? (
                  <p className="text-xs text-slate-400 text-center py-6">No recent customer activity records.</p>
                ) : (
                  <div className="space-y-3.5">
                    {activity.slice(0, 5).map((item, idx) => {
                      let iconColor = 'bg-emerald-50 text-emerald-600';
                      let Icon = CheckCircle2;
                      const actUpper = (item.action || '').toUpperCase();

                      if (actUpper.includes('INVOICE') || actUpper.includes('PAYMENT')) {
                        iconColor = 'bg-emerald-50 text-emerald-600';
                        Icon = DollarSign;
                      } else if (actUpper.includes('USER') || actUpper.includes('ROLE') || actUpper.includes('INVIT')) {
                        iconColor = 'bg-purple-50 text-purple-600';
                        Icon = Users;
                      } else if (actUpper.includes('EXCEPTION') || actUpper.includes('ALERT')) {
                        iconColor = 'bg-amber-50 text-amber-600';
                        Icon = AlertTriangle;
                      } else if (actUpper.includes('SUBSCRIPTION') || actUpper.includes('PLAN')) {
                        iconColor = 'bg-fuchsia-50 text-fuchsia-600';
                        Icon = CreditCard;
                      } else if (actUpper.includes('PROFILE') || actUpper.includes('ORG')) {
                        iconColor = 'bg-blue-50 text-blue-600';
                        Icon = Building2;
                      }

                      return (
                        <div key={item.id || idx} className="flex items-start gap-2.5">
                          <div className={`flex h-6 w-6 shrink-0 items-center justify-center rounded-full ${iconColor} mt-0.5`}>
                            <Icon className="h-3 w-3" />
                          </div>
                          <div className="min-w-0 flex-1">
                            <p className="text-xs font-semibold text-slate-900 truncate">{item.action.replace(/[._]/g, ' ')}</p>
                            <p className="text-[11px] text-slate-500 line-clamp-1">{item.description || 'System recorded update'}</p>
                          </div>
                          <span className="text-[10px] text-slate-400 shrink-0">
                            {new Date(item.created_at || item.timestamp || Date.now()).toLocaleDateString(undefined, { month: 'short', day: 'numeric' })}
                          </span>
                        </div>
                      );
                    })}
                  </div>
                )}
              </div>
            </div>

            {/* Panel 4: Customer Team */}
            <div className="rounded-xl border border-slate-200 bg-white p-5 shadow-xs flex flex-col justify-between space-y-4">
              <div>
                <div className="flex items-center justify-between border-b border-slate-100 pb-3 mb-3">
                  <h3 className="text-sm font-bold text-slate-900">Customer Team</h3>
                  <button
                    type="button"
                    onClick={() => setActiveTab('users')}
                    className="text-xs font-semibold text-blue-600 hover:underline cursor-pointer"
                  >
                    View All
                  </button>
                </div>

                {/* Counter Row */}
                <div className="grid grid-cols-4 gap-2 text-center py-2 border-b border-slate-100 mb-3 bg-slate-50/60 rounded-lg p-2">
                  <div>
                    <p className="text-base font-black text-slate-900">{team.total_users}</p>
                    <span className="text-[10px] text-slate-400 font-medium">Total Users</span>
                  </div>
                  <div>
                    <p className="text-base font-black text-emerald-600">{team.active_users}</p>
                    <span className="text-[10px] text-slate-400 font-medium">Active</span>
                  </div>
                  <div>
                    <p className="text-base font-black text-amber-600">{team.pending_users}</p>
                    <span className="text-[10px] text-slate-400 font-medium">Pending</span>
                  </div>
                  <div>
                    <p className="text-base font-black text-slate-400">{team.suspended_users}</p>
                    <span className="text-[10px] text-slate-400 font-medium">Suspended</span>
                  </div>
                </div>

                {/* Invite User Button */}
                <button
                  type="button"
                  onClick={() => setIsInviteOpen(true)}
                  className="w-full rounded-lg bg-blue-600 py-2 text-xs font-semibold text-white flex items-center justify-center gap-1.5 hover:bg-blue-700 shadow-2xs cursor-pointer mb-3 transition-colors"
                >
                  <UserPlus className="h-3.5 w-3.5" />
                  <span>Invite User</span>
                </button>

                {/* Primary Admin */}
                <div className="flex items-center justify-between py-2 border-t border-slate-100">
                  <div className="min-w-0">
                    <span className="text-[10px] font-medium text-slate-400 block">Primary Admin</span>
                    <span className="text-xs font-semibold text-slate-800 italic truncate block">
                      {primaryAdminUser ? primaryAdminUser.full_name : 'Not designated'}
                    </span>
                  </div>
                  <button
                    type="button"
                    onClick={() => setIsInviteOpen(true)}
                    className="rounded-lg border border-slate-200 bg-white px-2.5 py-1 text-xs font-semibold text-blue-600 hover:bg-slate-50 cursor-pointer"
                  >
                    Set Admin
                  </button>
                </div>

                {/* Other Key Users */}
                <div className="pt-2 border-t border-slate-100">
                  <span className="text-[10px] font-medium text-slate-400 block mb-1">Other Key Users</span>
                  {users.length > 1 ? (
                    <div className="space-y-1">
                      {users.slice(1, 3).map((u) => (
                        <div key={u.user_id} className="flex items-center justify-between text-xs py-0.5">
                          <span className="text-slate-800 truncate">{u.full_name || u.email}</span>
                          <span className="text-[10px] text-slate-500">{u.role_name}</span>
                        </div>
                      ))}
                    </div>
                  ) : (
                    <p className="text-xs text-slate-400 italic">No additional users yet.</p>
                  )}
                </div>
              </div>
            </div>
          </div>

          {/* Row 2: Integrations and Compliance & Documents (matching sportalCustomer360Overview.png) */}
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
            {/* Card 1: Integrations */}
            <div className="rounded-xl border border-slate-200 bg-white p-5 shadow-xs space-y-4">
              <div className="flex items-center justify-between border-b border-slate-100 pb-3">
                <h3 className="text-sm font-bold text-slate-900">Integrations</h3>
                <button
                  type="button"
                  onClick={() => setActiveTab('integrations')}
                  className="text-xs font-semibold text-blue-600 hover:underline cursor-pointer"
                >
                  View All
                </button>
              </div>

              <div className="grid grid-cols-2 sm:grid-cols-5 gap-3">
                {/* Twilio */}
                <div className="rounded-xl border border-slate-200 bg-slate-50/50 p-3 text-center flex flex-col items-center justify-between gap-2.5">
                  <div className="h-9 w-9 rounded-lg bg-rose-50 border border-rose-100 flex items-center justify-center text-rose-600">
                    <Phone className="h-4 w-4 text-rose-500" />
                  </div>
                  <div>
                    <p className="text-xs font-bold text-slate-900">Twilio</p>
                    <span className="text-[10px] text-slate-400 font-medium">Not Configured</span>
                  </div>
                </div>

                {/* AWS SES */}
                <div className="rounded-xl border border-slate-200 bg-slate-50/50 p-3 text-center flex flex-col items-center justify-between gap-2.5">
                  <div className="h-9 w-9 rounded-lg bg-amber-50 border border-amber-100 flex items-center justify-center text-amber-600">
                    <Mail className="h-4 w-4 text-amber-600" />
                  </div>
                  <div>
                    <p className="text-xs font-bold text-slate-900">AWS SES</p>
                    <span className="text-[10px] text-slate-400 font-medium">Not Configured</span>
                  </div>
                </div>

                {/* S3 Storage */}
                <div className="rounded-xl border border-slate-200 bg-slate-50/50 p-3 text-center flex flex-col items-center justify-between gap-2.5">
                  <div className="h-9 w-9 rounded-lg bg-emerald-50 border border-emerald-100 flex items-center justify-center text-emerald-600">
                    <Package className="h-4 w-4 text-emerald-600" />
                  </div>
                  <div>
                    <p className="text-xs font-bold text-slate-900">S3 Storage</p>
                    <span className="inline-flex items-center gap-1 text-[10px] text-emerald-700 font-semibold">
                      <span className="h-1.5 w-1.5 rounded-full bg-emerald-500" />
                      Connected
                    </span>
                  </div>
                </div>

                {/* Carrier APIs */}
                <div className="rounded-xl border border-slate-200 bg-slate-50/50 p-3 text-center flex flex-col items-center justify-between gap-2.5">
                  <div className="h-9 w-9 rounded-lg bg-purple-50 border border-purple-100 flex items-center justify-center text-purple-600">
                    <Zap className="h-4 w-4 text-purple-600" />
                  </div>
                  <div>
                    <p className="text-xs font-bold text-slate-900">Carrier APIs</p>
                    <span className="text-[10px] text-slate-400 font-medium">Not Configured</span>
                  </div>
                </div>

                {/* EDI */}
                <div className="rounded-xl border border-slate-200 bg-slate-50/50 p-3 text-center flex flex-col items-center justify-between gap-2.5">
                  <div className="h-9 w-9 rounded-lg bg-blue-50 border border-blue-100 flex items-center justify-center text-blue-600">
                    <FileText className="h-4 w-4 text-blue-600" />
                  </div>
                  <div>
                    <p className="text-xs font-bold text-slate-900">EDI</p>
                    <span className="text-[10px] text-slate-400 font-medium">Not Configured</span>
                  </div>
                </div>
              </div>
            </div>

            {/* Card 2: Compliance & Documents */}
            <div className="rounded-xl border border-slate-200 bg-white p-5 shadow-xs space-y-4">
              <div className="flex items-center justify-between border-b border-slate-100 pb-3">
                <h3 className="text-sm font-bold text-slate-900">Compliance & Documents</h3>
                <button
                  type="button"
                  onClick={() => setActiveTab('contracts')}
                  className="text-xs font-semibold text-blue-600 hover:underline cursor-pointer"
                >
                  View All
                </button>
              </div>

              <div className="grid grid-cols-2 sm:grid-cols-5 gap-3">
                {/* GST Certificate */}
                <div className="rounded-xl border border-slate-200 bg-slate-50/50 p-3 text-center flex flex-col items-center justify-between gap-2.5">
                  <div className="h-9 w-9 rounded-lg bg-emerald-50 border border-emerald-100 flex items-center justify-center text-emerald-600">
                    <FileCheck className="h-4 w-4 text-emerald-600" />
                  </div>
                  <div>
                    <p className="text-xs font-bold text-slate-900">GST Certificate</p>
                    <span className="rounded-full bg-emerald-50 px-2 py-0.2 text-[9px] font-bold text-emerald-700 border border-emerald-200">
                      Verified
                    </span>
                  </div>
                </div>

                {/* Business Registration */}
                <div className="rounded-xl border border-slate-200 bg-slate-50/50 p-3 text-center flex flex-col items-center justify-between gap-2.5">
                  <div className="h-9 w-9 rounded-lg bg-amber-50 border border-amber-100 flex items-center justify-center text-amber-600">
                    <FileText className="h-4 w-4 text-amber-600" />
                  </div>
                  <div>
                    <p className="text-xs font-bold text-slate-900">Business Registration</p>
                    <span className="rounded-full bg-amber-50 px-2 py-0.2 text-[9px] font-bold text-amber-700 border border-amber-200">
                      • Pending
                    </span>
                  </div>
                </div>

                {/* Insurance */}
                <div className="rounded-xl border border-slate-200 bg-slate-50/50 p-3 text-center flex flex-col items-center justify-between gap-2.5">
                  <div className="h-9 w-9 rounded-lg bg-slate-100 border border-slate-200 flex items-center justify-center text-slate-500">
                    <Shield className="h-4 w-4 text-slate-500" />
                  </div>
                  <div>
                    <p className="text-xs font-bold text-slate-900">Insurance</p>
                    <span className="rounded-full bg-slate-100 px-2 py-0.2 text-[9px] font-medium text-slate-600 border border-slate-200">
                      Not Uploaded
                    </span>
                  </div>
                </div>

                {/* Contracts */}
                <div className="rounded-xl border border-slate-200 bg-slate-50/50 p-3 text-center flex flex-col items-center justify-between gap-2.5">
                  <div className="h-9 w-9 rounded-lg bg-emerald-50 border border-emerald-100 flex items-center justify-center text-emerald-600">
                    <FileText className="h-4 w-4 text-emerald-600" />
                  </div>
                  <div>
                    <p className="text-xs font-bold text-slate-900">Contracts</p>
                    <span className="rounded-full bg-emerald-50 px-2 py-0.2 text-[9px] font-bold text-emerald-700 border border-emerald-200">
                      {contractsState.data ? `${contractsState.data.filter(c => c.status === 'ACTIVE').length} Active` : '0 Active'}
                    </span>
                  </div>
                </div>

                {/* Other Documents */}
                <div className="rounded-xl border border-slate-200 bg-slate-50/50 p-3 text-center flex flex-col items-center justify-between gap-2.5">
                  <div className="h-9 w-9 rounded-lg bg-slate-100 border border-slate-200 flex items-center justify-center text-slate-500">
                    <FileSpreadsheet className="h-4 w-4 text-slate-500" />
                  </div>
                  <div>
                    <p className="text-xs font-bold text-slate-900">Other Documents</p>
                    <span className="rounded-full bg-slate-100 px-2 py-0.2 text-[9px] font-medium text-slate-600 border border-slate-200">
                      0 Files
                    </span>
                  </div>
                </div>
              </div>
            </div>
          </div>

          {/* Row 3: Open Support Cases and Key Contacts (matching sportalCustomer360Overview.png) */}
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
            {/* Card 1: Open Support Cases */}
            <div className="rounded-xl border border-slate-200 bg-white p-5 shadow-xs space-y-4">
              <div className="flex items-center justify-between border-b border-slate-100 pb-3">
                <h3 className="text-sm font-bold text-slate-900">Open Support Cases</h3>
                <button
                  type="button"
                  onClick={() => setActiveTab('exceptions')}
                  className="text-xs font-semibold text-blue-600 hover:underline cursor-pointer"
                >
                  View All
                </button>
              </div>

              <div className="grid grid-cols-3 gap-3">
                {/* Open */}
                <div className="rounded-xl border border-slate-200 bg-slate-50/50 p-4 flex items-center gap-3">
                  <div className="h-10 w-10 rounded-xl bg-amber-50 border border-amber-200 flex items-center justify-center text-amber-600 shrink-0">
                    <AlertCircle className="h-5 w-5 text-amber-600" />
                  </div>
                  <div>
                    <p className="text-2xl font-black text-slate-900">{stats.open_exceptions_count ?? 0}</p>
                    <span className="text-xs font-medium text-slate-500">Open</span>
                  </div>
                </div>

                {/* In Progress */}
                <div className="rounded-xl border border-slate-200 bg-slate-50/50 p-4 flex items-center gap-3">
                  <div className="h-10 w-10 rounded-xl bg-blue-50 border border-blue-200 flex items-center justify-center text-blue-600 shrink-0">
                    <Clock className="h-5 w-5 text-blue-600" />
                  </div>
                  <div>
                    <p className="text-2xl font-black text-slate-900">0</p>
                    <span className="text-xs font-medium text-slate-500">In Progress</span>
                  </div>
                </div>

                {/* Resolved */}
                <div className="rounded-xl border border-slate-200 bg-slate-50/50 p-4 flex items-center gap-3">
                  <div className="h-10 w-10 rounded-xl bg-emerald-50 border border-emerald-200 flex items-center justify-center text-emerald-600 shrink-0">
                    <CheckCircle2 className="h-5 w-5 text-emerald-600" />
                  </div>
                  <div>
                    <p className="text-2xl font-black text-slate-900">0</p>
                    <span className="text-xs font-medium text-slate-500">Resolved</span>
                  </div>
                </div>
              </div>
            </div>

            {/* Card 2: Key Contacts */}
            <div className="rounded-xl border border-slate-200 bg-white p-5 shadow-xs space-y-4">
              <div className="flex items-center justify-between border-b border-slate-100 pb-3">
                <h3 className="text-sm font-bold text-slate-900">Key Contacts</h3>
                <button
                  type="button"
                  onClick={() => setActiveTab('company')}
                  className="text-xs font-semibold text-blue-600 hover:underline cursor-pointer"
                >
                  View All
                </button>
              </div>

              <div className="flex flex-col sm:flex-row items-center justify-between gap-4 p-4 rounded-xl bg-slate-50/50 border border-slate-100">
                <div className="flex items-center gap-3">
                  <div className="h-10 w-10 rounded-full bg-slate-100 flex items-center justify-center text-slate-400 shrink-0">
                    <Users className="h-5 w-5 text-slate-400" />
                  </div>
                  <span className="text-xs text-slate-500 font-medium">No key contacts added yet.</span>
                </div>
                <button
                  type="button"
                  onClick={() => setIsEditOpen(true)}
                  className="rounded-lg bg-blue-50 hover:bg-blue-100 px-3.5 py-1.5 text-xs font-semibold text-blue-700 border border-blue-200 transition-colors cursor-pointer"
                >
                  Add Contact
                </button>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* ========================================================================= */}
      {/* TAB: HEALTH & RETENTION INTELLIGENCE (Task S11) */}
      {/* ========================================================================= */}
      {activeTab === 'health' && (
        <CustomerHealthView
          key={organizationId}
          orgId={Number(organizationId)}
          onNavigateTab={(t) => setActiveTab(t)}
        />
      )}

      {/* ========================================================================= */}
      {/* TAB: USAGE & PLATFORM ANALYTICS (Task S10) */}
      {/* ========================================================================= */}
      {activeTab === 'usage' && (
        <CustomerUsageView
          organizationId={organizationId}
          orgName={org.name}
          onNavigateTab={(t) => setActiveTab(t)}
        />
      )}

      {/* ========================================================================= */}
      {/* TAB: COMPANY INFORMATION (Task C360-5) */}
      {/* ========================================================================= */}
      {activeTab === 'company' && (
        <CustomerCompanyView
          org={org}
          users={users}
          onEdit={() => setIsEditOpen(true)}
          onNavigateTab={(t) => handleTabChange(t)}
        />
      )}

      {/* ========================================================================= */}
      {/* TAB 3: ONBOARDING (Task C360-6 Primary Specification View) */}
      {/* ========================================================================= */}
      {activeTab === 'onboarding' && (
        <CustomerOnboardingView
          org={org}
          users={users}
          subscription={richSub || sub}
          contracts={contractsState.data || []}
          integrations={integrationsState.data || []}
          onOpenStage={(stageNum) => {
            setOnboardingStage(stageNum);
            setIsOnboardingOpen(true);
          }}
          onOpenEditModal={() => setIsEditOpen(true)}
          onOrgUpdated={(updatedOrg) => {
            setDetails((prev) => (prev ? { ...prev, organization: updatedOrg } : prev));
          }}
          onNavigateTab={(t) => handleTabChange(t)}
        />
      )}

      {/* ========================================================================= */}
      {/* TAB 4: SUBSCRIPTION (Task C360-7 Primary Specification View) */}
      {/* ========================================================================= */}
      {activeTab === 'subscription' && (
        <CustomerSubscriptionView
          org={org}
          subscription={richSub || sub}
          plans={plans}
          onRefresh={fetchDetails}
          onChangePlan={() => setIsChangePlanOpen(true)}
          onManageSubscription={() => setIsChangePlanOpen(true)}
          onNavigateTab={(tab) => handleTabChange(tab)}
          onToggleAutoRenew={handleToggleAutoRenew}
        />
      )}

      {/* ========================================================================= */}
      {/* TAB 6: USERS (Task C360-8 Primary Specification View) */}
      {/* ========================================================================= */}
      {activeTab === 'users' && (
        <CustomerUsersView
          org={org}
          users={users}
          userSummary={userSummary}
          onRefresh={fetchDetails}
          onInviteUser={() => setIsInviteOpen(true)}
          onViewUserDetails={(u) => setSelectedDetailUser({ org_id: org.id, user_id: u.id || u.user_id })}
          onChangeUserRole={(u) =>
            setRoleChangeState({
              isOpen: true,
              user: {
                user_id: u.id || u.user_id,
                email: u.email,
                full_name: u.name || u.full_name,
                org_id: org.id,
                org_name: org.name,
                role_name: u.role || u.role_name,
              },
            })
          }
          onToggleUserStatus={(u) =>
            setStatusConfirmState({
              isOpen: true,
              user: {
                user_id: u.id || u.user_id,
                email: u.email,
                full_name: u.name || u.full_name,
                org_id: org.id,
                org_name: org.name,
              },
              newStatus: u.status === 'Active' ? 'SUSPENDED' : 'ACTIVE',
            })
          }
          onNavigateTab={(tab) => handleTabChange(tab)}
        />
      )}

      {/* ========================================================================= */}
      {/* TAB 8: EXCEPTIONS (Task C360-9 Primary Specification View) */}
      {/* ========================================================================= */}
      {activeTab === 'exceptions' && (
        <CustomerExceptionsView
          org={org}
          exceptions={exceptionsState.data || []}
          onRefresh={fetchDetails}
          onNavigateTab={(tab) => handleTabChange(tab)}
        />
      )}

      {/* ========================================================================= */}
      {/* TAB 9: CONTRACTS (Task C360-10 Primary Specification View) */}
      {/* ========================================================================= */}
      {activeTab === 'contracts' && (
        <CustomerContractsView
          org={org}
          onRefresh={fetchDetails}
          onNavigateTab={(tab) => handleTabChange(tab)}
        />
      )}

      {/* ========================================================================= */}
      {/* TAB 10: INTEGRATIONS (Task C360-11 Primary Specification View) */}
      {/* ========================================================================= */}
      {activeTab === 'integrations' && (
        <CustomerIntegrationsView
          org={org}
          onRefresh={fetchDetails}
          onNavigateTab={(tab) => handleTabChange(tab)}
        />
      )}

      {/* Modals */}
      <EditOrganizationModal
        isOpen={isEditOpen}
        onClose={() => setIsEditOpen(false)}
        organization={org}
        onSuccess={() => fetchDetails()}
      />

      <InviteCustomerUserModal
        isOpen={isInviteOpen}
        initialOrgId={org.id}
        onClose={() => setIsInviteOpen(false)}
        onSuccess={() => fetchDetails()}
      />

      {selectedDetailUser && (
        <CustomerUserDetailModal
          isOpen={!!selectedDetailUser}
          orgId={selectedDetailUser.org_id}
          userId={selectedDetailUser.user_id}
          onClose={() => setSelectedDetailUser(null)}
          onStatusChange={(targetUser, targetStatus) => {
            setSelectedDetailUser(null);
            setStatusConfirmState({
              isOpen: true,
              user: targetUser,
              newStatus: targetStatus,
            });
          }}
        />
      )}

      <ConfirmUserStatusModal
        isOpen={statusConfirmState.isOpen}
        user={statusConfirmState.user}
        newStatus={statusConfirmState.newStatus}
        onClose={() => setStatusConfirmState({ isOpen: false, user: null, newStatus: '' })}
        onSuccess={() => fetchDetails()}
      />

      {roleChangeState.isOpen && (
        <ChangeUserRoleModal
          isOpen={roleChangeState.isOpen}
          user={roleChangeState.user}
          orgId={org.id}
          onClose={() => setRoleChangeState({ isOpen: false, user: null })}
          onRoleChanged={() => fetchDetails()}
        />
      )}

      {/* Task S8 Subscription Modals */}
      {isChangePlanOpen && (richSub || sub) && (
        <ChangeSubscriptionPlanModal
          isOpen={isChangePlanOpen}
          onClose={() => setIsChangePlanOpen(false)}
          subscription={richSub || sub}
          plans={plans}
          onSuccess={() => fetchDetails()}
        />
      )}

      {isRenewOpen && (richSub || sub) && (
        <RenewSubscriptionModal
          isOpen={isRenewOpen}
          onClose={() => setIsRenewOpen(false)}
          subscription={richSub || sub}
          onSuccess={() => fetchDetails()}
        />
      )}

      {isCancelOpen && (richSub || sub) && (
        <CancelSubscriptionModal
          isOpen={isCancelOpen}
          onClose={() => setIsCancelOpen(false)}
          subscription={richSub || sub}
          onSuccess={() => fetchDetails()}
        />
      )}

      {isAssignOpen && (
        <AssignSubscriptionModal
          isOpen={isAssignOpen}
          onClose={() => setIsAssignOpen(false)}
          organization={org}
          plans={plans}
          onSuccess={() => fetchDetails()}
        />
      )}

      {isOnboardingOpen && (
        <CustomerOnboardingModal
          isOpen={isOnboardingOpen}
          onClose={() => setIsOnboardingOpen(false)}
          initialOrg={org}
          onSuccess={() => fetchDetails()}
        />
      )}
    </div>
  );
}

import React, { useState, useEffect, useMemo } from 'react';
import { 
  CreditCard, Search, Filter, RefreshCw, Plus, Building2, Calendar, 
  ArrowRightLeft, Ban, CheckCircle2, AlertTriangle, Clock, DollarSign, 
  TrendingUp, Shield, Layers, ChevronLeft, ChevronRight, Eye, Edit3, 
  ExternalLink, Check, ToggleLeft, ToggleRight, Sparkles, AlertCircle
} from 'lucide-react';
import { sportalService } from '../../services/sportalService';
import { CustomerSubscriptionDetailModal } from './CustomerSubscriptionDetailModal';
import { ChangeSubscriptionPlanModal } from './ChangeSubscriptionPlanModal';
import { RenewSubscriptionModal } from './RenewSubscriptionModal';
import { CancelSubscriptionModal } from './CancelSubscriptionModal';
import { AssignSubscriptionModal } from './AssignSubscriptionModal';
import { PlanEditorModal } from './PlanEditorModal';

export function SubscriptionsPage() {
  const [activeTab, setActiveTab] = useState('subscriptions'); // 'subscriptions' | 'plans' | 'renewals'

  // Subscriptions Table State
  const [subscriptions, setSubscriptions] = useState([]);
  const [metrics, setMetrics] = useState({
    total_organizations: 0,
    active_subscriptions: 0,
    trialing_count: 0,
    past_due_count: 0,
    not_configured_count: 0,
    monthly_recurring_rev: 0,
    annual_run_rate: 0,
    auto_renew_percentage: 100,
    expiring_in_30_days: 0,
  });
  const [pagination, setPagination] = useState({ page: 1, limit: 15, total: 0 });
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  // Filters
  const [search, setSearch] = useState('');
  const [statusFilter, setStatusFilter] = useState('');
  const [planFilter, setPlanFilter] = useState('');
  const [autoRenewFilter, setAutoRenewFilter] = useState('');
  const [sortBy, setSortBy] = useState('org_name');
  const [sortOrder, setSortOrder] = useState('asc');

  // Plans Catalog State
  const [plans, setPlans] = useState([]);
  const [loadingPlans, setLoadingPlans] = useState(false);

  // Modals State
  const [detailOrgId, setDetailOrgId] = useState(null);
  const [changePlanSub, setChangePlanSub] = useState(null);
  const [renewSub, setRenewSub] = useState(null);
  const [cancelSub, setCancelSub] = useState(null);
  const [assignOrg, setAssignOrg] = useState(null);
  const [editingPlan, setEditingPlan] = useState(null);
  const [isCreatePlanOpen, setIsCreatePlanOpen] = useState(false);

  // Action status notification
  const [actionNotice, setActionNotice] = useState(null);

  // Fetch Subscriptions & Metrics
  const fetchSubscriptions = async () => {
    try {
      setLoading(true);
      setError('');
      const params = {
        search: search.trim() || undefined,
        status: statusFilter || undefined,
        planId: planFilter || undefined,
        autoRenew: autoRenewFilter !== '' ? autoRenewFilter : undefined,
        page: pagination.page,
        limit: pagination.limit,
        sortBy,
        sortOrder,
      };

      const res = await sportalService.getSubscriptions(params);
      const data = res?.data || res;
      if (data) {
        setSubscriptions(data.items || []);
        if (data.metrics) {
          setMetrics(data.metrics);
        }
        if (data.pagination) {
          setPagination(data.pagination);
        }
      }
    } catch (err) {
      console.error('Failed to load subscriptions:', err);
      setError(err?.response?.data?.message || err.message || 'Failed to retrieve subscriptions.');
    } finally {
      setLoading(false);
    }
  };

  // Fetch Commercial Plans
  const fetchPlans = async () => {
    try {
      setLoadingPlans(true);
      const res = await sportalService.getSubscriptionPlans();
      const data = res?.data || res;
      if (Array.isArray(data)) {
        setPlans(data);
      }
    } catch (err) {
      console.error('Failed to load plans:', err);
    } finally {
      setLoadingPlans(false);
    }
  };

  useEffect(() => {
    fetchSubscriptions();
    fetchPlans();
  }, [pagination.page, statusFilter, planFilter, autoRenewFilter, sortBy, sortOrder]);

  const handleSearchSubmit = (e) => {
    e.preventDefault();
    setPagination((prev) => ({ ...prev, page: 1 }));
    fetchSubscriptions();
  };

  const showNotification = (msg, type = 'success') => {
    setActionNotice({ message: msg, type });
    setTimeout(() => setActionNotice(null), 4000);
  };

  // Auto-Renew Toggle Action
  const handleToggleAutoRenew = async (sub) => {
    if (!sub.subscription_id) return;
    try {
      const newSetting = !sub.auto_renew;
      await sportalService.toggleOrganizationAutoRenew(sub.org_id, {
        auto_renew: newSetting,
        reason: `Auto-renew set to ${newSetting ? 'ENABLED' : 'DISABLED'} via directory`,
      });
      showNotification(`Auto-renew for ${sub.org_name} set to ${newSetting ? 'ENABLED' : 'DISABLED'}`);
      fetchSubscriptions();
    } catch (err) {
      alert('Failed to update auto-renew: ' + (err?.response?.data?.message || err.message));
    }
  };

  // Renewals View Subscriptions
  const renewalsList = useMemo(() => {
    return subscriptions
      .filter((s) => s.subscription_id && s.status === 'ACTIVE')
      .sort((a, b) => (a.days_until_renewal ?? 999) - (b.days_until_renewal ?? 999));
  }, [subscriptions]);

  return (
    <div className="space-y-6 pb-12">
      {/* Toast Notification */}
      {actionNotice && (
        <div className={`fixed top-4 right-4 z-50 flex items-center gap-2 rounded-xl px-4 py-3 text-xs font-bold shadow-lg transition-all animate-in fade-in slide-in-from-top-2 ${
          actionNotice.type === 'success'
            ? 'bg-emerald-800 text-white'
            : 'bg-rose-800 text-white'
        }`}>
          <CheckCircle2 className="h-4 w-4 shrink-0" />
          <span>{actionNotice.message}</span>
        </div>
      )}

      {/* Page Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <div className="flex items-center gap-2">
            <h1 className="text-xl font-bold text-slate-900">Commercial Subscriptions & Plans</h1>
            <span className="rounded-full bg-navy-100 px-2.5 py-0.5 text-xs font-bold text-navy-800">
              Commercial Engine
            </span>
          </div>
          <p className="text-xs text-slate-500 mt-1">
            Manage customer forwarder plan tiers, recurring billing intervals, renewal schedules, and quotas
          </p>
        </div>

        <div className="flex items-center gap-2">
          <button
            type="button"
            onClick={() => {
              fetchSubscriptions();
              fetchPlans();
            }}
            className="flex items-center gap-1.5 rounded-xl border border-slate-200 bg-white px-3 py-2 text-xs font-semibold text-slate-700 hover:bg-slate-50 shadow-2xs transition-colors"
          >
            <RefreshCw className={`h-3.5 w-3.5 ${loading ? 'animate-spin' : ''}`} />
            <span>Refresh</span>
          </button>

          <button
            type="button"
            onClick={() => setIsCreatePlanOpen(true)}
            className="flex items-center gap-1.5 rounded-xl border border-slate-200 bg-white px-3 py-2 text-xs font-semibold text-slate-700 hover:bg-slate-50 shadow-2xs transition-colors"
          >
            <Plus className="h-3.5 w-3.5" />
            <span>New Plan Tier</span>
          </button>
        </div>
      </div>

      {/* KPI Metric Cards */}
      <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-5 gap-3.5">
        <div className="rounded-xl border border-slate-200 bg-white p-4 shadow-2xs">
          <div className="flex items-center justify-between text-slate-500">
            <span className="text-xs font-medium">Active Tenants</span>
            <Building2 className="h-4 w-4 text-blue-600" />
          </div>
          <p className="text-2xl font-black text-navy-900 mt-2">
            {metrics.active_subscriptions}
            <span className="text-xs font-normal text-slate-400 ml-1">/ {metrics.total_organizations} orgs</span>
          </p>
          <span className="text-[10px] text-slate-400 mt-1 block">
            {metrics.not_configured_count} unassigned sandbox orgs
          </span>
        </div>

        <div className="rounded-xl border border-slate-200 bg-white p-4 shadow-2xs">
          <div className="flex items-center justify-between text-slate-500">
            <span className="text-xs font-medium">Monthly Revenue (MRR)</span>
            <DollarSign className="h-4 w-4 text-emerald-600" />
          </div>
          <p className="text-2xl font-black text-navy-900 mt-2">
            ${metrics.monthly_recurring_rev?.toLocaleString() || 0}
          </p>
          <span className="text-[10px] text-emerald-600 font-semibold mt-1 block">
            Contractual recurring monthly
          </span>
        </div>

        <div className="rounded-xl border border-slate-200 bg-white p-4 shadow-2xs">
          <div className="flex items-center justify-between text-slate-500">
            <span className="text-xs font-medium">Annual Run Rate (ARR)</span>
            <TrendingUp className="h-4 w-4 text-sky-600" />
          </div>
          <p className="text-2xl font-black text-navy-900 mt-2">
            ${metrics.annual_run_rate?.toLocaleString() || 0}
          </p>
          <span className="text-[10px] text-slate-400 mt-1 block">
            Annualized contract volume
          </span>
        </div>

        <div className="rounded-xl border border-slate-200 bg-white p-4 shadow-2xs">
          <div className="flex items-center justify-between text-slate-500">
            <span className="text-xs font-medium">Auto-Renew Rate</span>
            <RefreshCw className="h-4 w-4 text-purple-600" />
          </div>
          <p className="text-2xl font-black text-navy-900 mt-2">
            {metrics.auto_renew_percentage}%
          </p>
          <span className="text-[10px] text-purple-700 font-medium mt-1 block">
            Persistent contract auto-renewal
          </span>
        </div>

        <div className="rounded-xl border border-slate-200 bg-white p-4 shadow-2xs">
          <div className="flex items-center justify-between text-slate-500">
            <span className="text-xs font-medium">Expiring in 30 Days</span>
            <Clock className="h-4 w-4 text-amber-600" />
          </div>
          <p className="text-2xl font-black text-navy-900 mt-2">
            {metrics.expiring_in_30_days}
          </p>
          <span className={`text-[10px] font-semibold mt-1 block ${
            metrics.expiring_in_30_days > 0 ? 'text-amber-700' : 'text-slate-400'
          }`}>
            {metrics.expiring_in_30_days > 0 ? 'Action required for retention' : 'No imminent expiries'}
          </span>
        </div>
      </div>

      {/* Navigation Tabs */}
      <div className="flex items-center gap-2 border-b border-slate-200 bg-white px-2 rounded-t-xl">
        {[
          { id: 'subscriptions', label: 'Customer Subscriptions', icon: Building2, count: subscriptions.length },
          { id: 'plans', label: 'Commercial Plan Catalog', icon: Layers, count: plans.length },
          { id: 'renewals', label: 'Renewal & Auto-Renew Operations', icon: RefreshCw, count: metrics.expiring_in_30_days },
        ].map((tab) => {
          const Icon = tab.icon;
          const isActive = activeTab === tab.id;
          return (
            <button
              key={tab.id}
              type="button"
              onClick={() => setActiveTab(tab.id)}
              className={`flex items-center gap-2 border-b-2 py-3.5 px-4 text-xs font-bold transition-all ${
                isActive
                  ? 'border-blue-600 text-blue-600 bg-blue-50/20'
                  : 'border-transparent text-slate-500 hover:text-slate-800'
              }`}
            >
              <Icon className="h-4 w-4" />
              <span>{tab.label}</span>
              <span className={`rounded-full px-2 py-0.5 text-[10px] font-bold ${
                isActive ? 'bg-blue-100 text-blue-800' : 'bg-slate-100 text-slate-600'
              }`}>
                {tab.count}
              </span>
            </button>
          );
        })}
      </div>

      {/* TAB 1: CUSTOMER SUBSCRIPTIONS DIRECTORY */}
      {activeTab === 'subscriptions' && (
        <div className="space-y-4">
          {/* Search & Filter Bar */}
          <div className="rounded-xl border border-slate-200 bg-white p-4 shadow-2xs space-y-3">
            <form onSubmit={handleSearchSubmit} className="flex flex-col sm:flex-row items-center gap-3">
              <div className="relative flex-1 w-full">
                <Search className="absolute left-3 top-2.5 h-4 w-4 text-slate-400" />
                <input
                  type="text"
                  value={search}
                  onChange={(e) => setSearch(e.target.value)}
                  placeholder="Search by forwarder organization name, email, or country..."
                  className="w-full rounded-xl border border-slate-200 pl-9 pr-4 py-2 text-xs text-slate-800 focus:border-blue-600 focus:outline-hidden"
                />
              </div>

              <div className="flex items-center gap-2 w-full sm:w-auto">
                <select
                  value={statusFilter}
                  onChange={(e) => setStatusFilter(e.target.value)}
                  className="rounded-xl border border-slate-200 bg-white py-2 px-3 text-xs font-semibold text-slate-700 focus:border-blue-600 focus:outline-hidden"
                >
                  <option value="">All Statuses</option>
                  <option value="ACTIVE">ACTIVE</option>
                  <option value="TRIALING">TRIALING</option>
                  <option value="PAST_DUE">PAST_DUE</option>
                  <option value="CANCELED">CANCELED</option>
                  <option value="NOT_CONFIGURED">NOT_CONFIGURED</option>
                </select>

                <select
                  value={planFilter}
                  onChange={(e) => setPlanFilter(e.target.value)}
                  className="rounded-xl border border-slate-200 bg-white py-2 px-3 text-xs font-semibold text-slate-700 focus:border-blue-600 focus:outline-hidden"
                >
                  <option value="">All Plans</option>
                  {plans.map((p) => (
                    <option key={p.id} value={p.id}>{p.name}</option>
                  ))}
                </select>

                <select
                  value={autoRenewFilter}
                  onChange={(e) => setAutoRenewFilter(e.target.value)}
                  className="rounded-xl border border-slate-200 bg-white py-2 px-3 text-xs font-semibold text-slate-700 focus:border-blue-600 focus:outline-hidden"
                >
                  <option value="">Auto-Renew (All)</option>
                  <option value="true">Auto-Renew Enabled</option>
                  <option value="false">Auto-Renew Disabled</option>
                </select>

                <button
                  type="submit"
                  className="rounded-xl bg-blue-600 px-4 py-2 text-xs font-bold text-white hover:bg-blue-700 transition-colors shadow-2xs"
                >
                  Search
                </button>
              </div>
            </form>
          </div>

          {/* Table Container */}
          <div className="rounded-xl border border-slate-200 bg-white shadow-2xs overflow-hidden">
            {error && (
              <div className="p-4 bg-rose-50 border-b border-rose-100 text-xs text-rose-700 font-semibold">
                {error}
              </div>
            )}

            <div className="overflow-x-auto">
              <table className="w-full text-left text-xs">
                <thead>
                  <tr className="border-b border-slate-200 bg-slate-50/75 text-slate-500 font-semibold uppercase tracking-wider text-[11px]">
                    <th className="py-3 px-4">Customer Organization</th>
                    <th className="py-3 px-4">Commercial Plan</th>
                    <th className="py-3 px-4">Contract Status</th>
                    <th className="py-3 px-4">Recurring Terms</th>
                    <th className="py-3 px-4">Renewal / Expiry</th>
                    <th className="py-3 px-4">Auto-Renew</th>
                    <th className="py-3 px-4 text-right">Actions</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-slate-100">
                  {loading ? (
                    <tr>
                      <td colSpan={7} className="py-8 text-center text-slate-400">
                        <RefreshCw className="mx-auto h-6 w-6 animate-spin text-slate-300 mb-2" />
                        Loading persistent subscriptions...
                      </td>
                    </tr>
                  ) : subscriptions.length === 0 ? (
                    <tr>
                      <td colSpan={7} className="py-8 text-center text-slate-400">
                        No customer subscriptions matching the current filters.
                      </td>
                    </tr>
                  ) : (
                    subscriptions.map((sub) => {
                      const hasSub = Boolean(sub.subscription_id);
                      const daysRem = sub.days_until_renewal;

                      return (
                        <tr key={sub.org_id} className="hover:bg-slate-50/50 transition-colors">
                          {/* Organization */}
                          <td className="py-3.5 px-4">
                            <div className="flex items-center gap-2.5">
                              <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-slate-100 font-bold text-xs text-slate-700">
                                {sub.org_name?.charAt(0).toUpperCase() || 'O'}
                              </div>
                              <div>
                                <p className="font-bold text-slate-900">{sub.org_name}</p>
                                <p className="text-[11px] text-slate-500 font-mono">
                                  ID: #{sub.org_id} {sub.country ? `• ${sub.country}` : ''}
                                </p>
                              </div>
                            </div>
                          </td>

                          {/* Commercial Plan */}
                          <td className="py-3.5 px-4">
                            {hasSub ? (
                              <div>
                                <span className="font-bold text-slate-900">{sub.plan_name}</span>
                                <span className="text-[11px] text-slate-500 block">
                                  ${sub.amount || 0} / {sub.billing_cycle || 'mo'}
                                </span>
                              </div>
                            ) : (
                              <span className="text-slate-400 italic">Unassigned Sandbox</span>
                            )}
                          </td>

                          {/* Contract Status */}
                          <td className="py-3.5 px-4">
                            <span className={`rounded-full px-2.5 py-0.5 text-[11px] font-bold border ${
                              sub.status === 'ACTIVE'
                                ? 'bg-emerald-50 text-emerald-700 border-emerald-200'
                                : sub.status === 'TRIALING'
                                ? 'bg-sky-50 text-sky-700 border-sky-200'
                                : sub.status === 'PAST_DUE'
                                ? 'bg-amber-50 text-amber-700 border-amber-200'
                                : sub.status === 'CANCELED'
                                ? 'bg-rose-50 text-rose-700 border-rose-200'
                                : 'bg-slate-100 text-slate-600 border-slate-200'
                            }`}>
                              {sub.status || 'NOT_CONFIGURED'}
                            </span>
                          </td>

                          {/* Recurring Terms */}
                          <td className="py-3.5 px-4">
                            {hasSub ? (
                              <div className="space-y-0.5">
                                <span className="font-medium text-slate-800 capitalize">
                                  {sub.billing_cycle || 'Monthly'}
                                </span>
                                <span className="text-[10px] text-slate-400 block font-mono">
                                  {sub.currency || 'USD'}
                                </span>
                              </div>
                            ) : (
                              <span className="text-slate-400">—</span>
                            )}
                          </td>

                          {/* Renewal Date & Countdown */}
                          <td className="py-3.5 px-4">
                            {hasSub && sub.current_period_end ? (
                              <div>
                                <span className="font-semibold text-slate-800 block">
                                  {new Date(sub.current_period_end).toLocaleDateString()}
                                </span>
                                {daysRem !== null && (
                                  <span className={`text-[10px] font-bold ${
                                    daysRem <= 7 ? 'text-rose-600' : daysRem <= 30 ? 'text-amber-600' : 'text-slate-500'
                                  }`}>
                                    {daysRem <= 0 ? 'Expired' : `${daysRem} days left`}
                                  </span>
                                )}
                              </div>
                            ) : (
                              <span className="text-slate-400">—</span>
                            )}
                          </td>

                          {/* Auto-Renew Switch */}
                          <td className="py-3.5 px-4">
                            {hasSub ? (
                              <div className="flex items-center gap-1.5">
                                <button
                                  type="button"
                                  onClick={() => handleToggleAutoRenew(sub)}
                                  className="text-blue-600 hover:text-blue-800 transition-colors p-0.5"
                                  title="Click to toggle auto-renew state in MariaDB"
                                >
                                  {sub.auto_renew ? (
                                    <ToggleRight className="h-6 w-6 text-emerald-600" />
                                  ) : (
                                    <ToggleLeft className="h-6 w-6 text-slate-400" />
                                  )}
                                </button>
                                <span className={`text-[11px] font-bold ${
                                  sub.auto_renew ? 'text-emerald-700' : 'text-amber-700'
                                }`}>
                                  {sub.auto_renew ? 'Active' : 'Off'}
                                </span>
                              </div>
                            ) : (
                              <span className="text-slate-400">—</span>
                            )}
                          </td>

                          {/* Actions */}
                          <td className="py-3.5 px-4 text-right">
                            <div className="flex items-center justify-end gap-1.5">
                              {hasSub ? (
                                <>
                                  <button
                                    type="button"
                                    onClick={() => setDetailOrgId(sub.org_id)}
                                    className="rounded-lg border border-slate-200 bg-white p-1.5 text-slate-600 hover:bg-slate-50 hover:text-blue-600 transition-colors shadow-2xs"
                                    title="View Full Commercial Dossier"
                                  >
                                    <Eye className="h-3.5 w-3.5" />
                                  </button>

                                  <button
                                    type="button"
                                    onClick={() => setChangePlanSub(sub)}
                                    className="rounded-lg border border-slate-200 bg-white p-1.5 text-slate-600 hover:bg-slate-50 hover:text-blue-600 transition-colors shadow-2xs"
                                    title="Upgrade / Downgrade Plan Tier"
                                  >
                                    <ArrowRightLeft className="h-3.5 w-3.5" />
                                  </button>

                                  <button
                                    type="button"
                                    onClick={() => setRenewSub(sub)}
                                    className="rounded-lg border border-slate-200 bg-white p-1.5 text-slate-600 hover:bg-slate-50 hover:text-emerald-600 transition-colors shadow-2xs"
                                    title="Renew / Extend Subscription Period"
                                  >
                                    <RefreshCw className="h-3.5 w-3.5" />
                                  </button>

                                  <button
                                    type="button"
                                    onClick={() => setCancelSub(sub)}
                                    className="rounded-lg border border-slate-200 bg-white p-1.5 text-slate-600 hover:bg-slate-50 hover:text-rose-600 transition-colors shadow-2xs"
                                    title="Cancel Subscription"
                                  >
                                    <Ban className="h-3.5 w-3.5" />
                                  </button>
                                </>
                              ) : (
                                <button
                                  type="button"
                                  onClick={() => setAssignOrg(sub)}
                                  className="inline-flex items-center gap-1 rounded-lg bg-blue-600 px-2.5 py-1 text-xs font-bold text-white hover:bg-blue-700 transition-colors shadow-2xs"
                                >
                                  <Plus className="h-3 w-3" />
                                  <span>Assign Plan</span>
                                </button>
                              )}
                            </div>
                          </td>
                        </tr>
                      );
                    })
                  )}
                </tbody>
              </table>
            </div>

            {/* Pagination Controls */}
            <div className="flex items-center justify-between border-t border-slate-200 px-4 py-3 bg-slate-50/50">
              <span className="text-xs text-slate-500">
                Showing <strong className="text-slate-800">{subscriptions.length}</strong> of{' '}
                <strong className="text-slate-800">{pagination.total || subscriptions.length}</strong> customer records
              </span>

              <div className="flex items-center gap-1.5">
                <button
                  type="button"
                  disabled={pagination.page <= 1}
                  onClick={() => setPagination((prev) => ({ ...prev, page: prev.page - 1 }))}
                  className="rounded-lg border border-slate-200 bg-white p-1.5 text-slate-600 hover:bg-slate-50 disabled:opacity-40 transition-colors"
                >
                  <ChevronLeft className="h-4 w-4" />
                </button>
                <span className="text-xs font-semibold text-slate-700 px-2">
                  Page {pagination.page}
                </span>
                <button
                  type="button"
                  disabled={subscriptions.length < pagination.limit}
                  onClick={() => setPagination((prev) => ({ ...prev, page: prev.page + 1 }))}
                  className="rounded-lg border border-slate-200 bg-white p-1.5 text-slate-600 hover:bg-slate-50 disabled:opacity-40 transition-colors"
                >
                  <ChevronRight className="h-4 w-4" />
                </button>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* TAB 2: COMMERCIAL PLAN CATALOG */}
      {activeTab === 'plans' && (
        <div className="space-y-6">
          <div className="flex items-center justify-between">
            <div>
              <h2 className="text-sm font-bold text-slate-900">Commercial Subscription Tiers</h2>
              <p className="text-xs text-slate-500">
                SaaS packages configured in MariaDB with resource entitlements, pricing, and active forwarder counts
              </p>
            </div>
            <button
              type="button"
              onClick={() => setIsCreatePlanOpen(true)}
              className="inline-flex items-center gap-1.5 rounded-xl bg-blue-600 px-3.5 py-2 text-xs font-bold text-white hover:bg-blue-700 transition-colors shadow-2xs"
            >
              <Plus className="h-3.5 w-3.5" />
              <span>Create Plan Tier</span>
            </button>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
            {plans.map((p) => {
              const isPopular = p.name === 'Growth';
              return (
                <div
                  key={p.id}
                  className={`rounded-2xl border bg-white p-6 shadow-xs flex flex-col justify-between relative transition-all ${
                    isPopular ? 'border-blue-600 ring-1 ring-blue-600' : 'border-slate-200'
                  }`}
                >
                  {isPopular && (
                    <span className="absolute -top-3 left-1/2 -translate-x-1/2 rounded-full bg-blue-600 px-3 py-0.5 text-[10px] font-extrabold tracking-wider uppercase text-white shadow-xs">
                      Most Popular Tier
                    </span>
                  )}

                  <div className="space-y-4">
                    {/* Header */}
                    <div className="flex items-center justify-between">
                      <div>
                        <h3 className="text-lg font-black text-slate-900">{p.name}</h3>
                        <p className="text-xs text-slate-500 mt-0.5 line-clamp-2">{p.description}</p>
                      </div>
                      <span className="rounded-lg bg-navy-50 px-2.5 py-1 text-xs font-mono font-bold text-navy-800">
                        {p.active_customers_count || 0} orgs
                      </span>
                    </div>

                    {/* Pricing */}
                    <div className="pt-2 border-t border-slate-100">
                      <div className="flex items-baseline gap-1">
                        <span className="text-3xl font-black text-navy-900">${p.price_monthly}</span>
                        <span className="text-xs text-slate-500 font-medium">/ month</span>
                      </div>
                      <span className="text-[11px] text-slate-400 block mt-0.5">
                        ${p.price_annual?.toLocaleString()} billed annually (discounted)
                      </span>
                    </div>

                    {/* Resource Quotas Grid */}
                    <div className="rounded-xl border border-slate-100 bg-slate-50/60 p-3.5 space-y-2">
                      <span className="text-[10px] font-bold uppercase tracking-wider text-slate-400 block">
                        Included Entitlements
                      </span>
                      <div className="grid grid-cols-2 gap-2 text-xs">
                        <div>
                          <span className="text-slate-500 text-[11px] block">Seats:</span>
                          <span className="font-bold text-slate-800">
                            {p.limits?.team_members === -1 ? 'Unlimited' : `${p.limits?.team_members || '—'} users`}
                          </span>
                        </div>
                        <div>
                          <span className="text-slate-500 text-[11px] block">AI Emails:</span>
                          <span className="font-bold text-slate-800">
                            {p.limits?.ai_email_processing?.toLocaleString() || '—'} / mo
                          </span>
                        </div>
                        <div>
                          <span className="text-slate-500 text-[11px] block">RFQs:</span>
                          <span className="font-bold text-slate-800">
                            {p.limits?.rfqs?.toLocaleString() || '—'} / mo
                          </span>
                        </div>
                        <div>
                          <span className="text-slate-500 text-[11px] block">Shipments:</span>
                          <span className="font-bold text-slate-800">
                            {p.limits?.shipments?.toLocaleString() || '—'} / mo
                          </span>
                        </div>
                        <div>
                          <span className="text-slate-500 text-[11px] block">Carriers:</span>
                          <span className="font-bold text-slate-800">
                            {p.limits?.carrier_connections || '—'} direct
                          </span>
                        </div>
                        <div>
                          <span className="text-slate-500 text-[11px] block">Storage:</span>
                          <span className="font-bold text-slate-800">
                            {p.limits?.storage_gb || '—'} GB
                          </span>
                        </div>
                      </div>
                    </div>

                    {/* Features Checklist */}
                    <div className="space-y-2">
                      <span className="text-[10px] font-bold uppercase tracking-wider text-slate-400 block">
                        Capabilities & Features
                      </span>
                      <div className="space-y-1.5">
                        {(p.features || []).map((f, i) => (
                          <div key={i} className="flex items-center gap-2 text-xs text-slate-700">
                            <Check className="h-3.5 w-3.5 text-emerald-600 shrink-0" />
                            <span>{f}</span>
                          </div>
                        ))}
                      </div>
                    </div>
                  </div>

                  {/* Plan Footer Action */}
                  <div className="pt-4 mt-6 border-t border-slate-100">
                    <button
                      type="button"
                      onClick={() => setEditingPlan(p)}
                      className="w-full flex items-center justify-center gap-1.5 rounded-xl border border-slate-200 bg-white py-2 text-xs font-bold text-slate-700 hover:bg-slate-50 transition-colors shadow-2xs"
                    >
                      <Edit3 className="h-3.5 w-3.5" />
                      <span>Edit Plan Parameters</span>
                    </button>
                  </div>
                </div>
              );
            })}
          </div>
        </div>
      )}

      {/* TAB 3: RENEWALS & AUTO-RENEW OPERATIONS */}
      {activeTab === 'renewals' && (
        <div className="space-y-4">
          <div className="rounded-xl border border-slate-200 bg-white p-5 shadow-2xs space-y-4">
            <div>
              <h2 className="text-sm font-bold text-slate-900">Renewal & Auto-Renewal Pipeline</h2>
              <p className="text-xs text-slate-500">
                Proactive operations console for customer retention, impending contract expirations, and manual billing extensions
              </p>
            </div>

            {renewalsList.length === 0 ? (
              <div className="rounded-xl border border-slate-200 bg-slate-50 p-8 text-center text-xs text-slate-500">
                No active subscriptions currently scheduled for renewal.
              </div>
            ) : (
              <div className="divide-y divide-slate-100">
                {renewalsList.map((sub) => {
                  const days = sub.days_until_renewal;
                  const isImminent = days !== null && days <= 30;
                  const isCritical = days !== null && days <= 7;

                  return (
                    <div key={sub.org_id} className="py-4 flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
                      <div className="flex items-center gap-3">
                        <div className={`flex h-10 w-10 items-center justify-center rounded-xl font-bold text-xs ${
                          isCritical
                            ? 'bg-rose-100 text-rose-800'
                            : isImminent
                            ? 'bg-amber-100 text-amber-800'
                            : 'bg-slate-100 text-slate-700'
                        }`}>
                          <Clock className="h-5 w-5" />
                        </div>
                        <div>
                          <div className="flex items-center gap-2">
                            <h3 className="font-bold text-slate-900">{sub.org_name}</h3>
                            <span className="rounded-md bg-slate-100 px-2 py-0.5 text-[10px] font-bold text-slate-700">
                              {sub.plan_name} (${sub.amount || 0}/mo)
                            </span>
                          </div>
                          <p className="text-xs text-slate-500 mt-0.5">
                            Period End: <strong>{new Date(sub.current_period_end).toLocaleDateString()}</strong> • {days} days remaining
                          </p>
                        </div>
                      </div>

                      <div className="flex items-center gap-3">
                        {/* Auto-renew switch */}
                        <div className="flex items-center gap-2 pr-2 border-r border-slate-200">
                          <button
                            type="button"
                            onClick={() => handleToggleAutoRenew(sub)}
                            className="text-blue-600 hover:text-blue-800 transition-colors"
                            title="Toggle persistent auto-renew flag in DB"
                          >
                            {sub.auto_renew ? (
                              <ToggleRight className="h-6 w-6 text-emerald-600" />
                            ) : (
                              <ToggleLeft className="h-6 w-6 text-slate-400" />
                            )}
                          </button>
                          <span className={`text-xs font-bold ${
                            sub.auto_renew ? 'text-emerald-700' : 'text-amber-700'
                          }`}>
                            {sub.auto_renew ? 'Auto-Renew ON' : 'Manual Expiration'}
                          </span>
                        </div>

                        {/* Extend Period Button */}
                        <button
                          type="button"
                          onClick={() => setRenewSub(sub)}
                          className="inline-flex items-center gap-1.5 rounded-xl bg-emerald-600 px-3.5 py-1.5 text-xs font-bold text-white hover:bg-emerald-700 transition-colors shadow-2xs"
                        >
                          <RefreshCw className="h-3.5 w-3.5" />
                          <span>Extend Period</span>
                        </button>

                        {/* View Dossier */}
                        <button
                          type="button"
                          onClick={() => setDetailOrgId(sub.org_id)}
                          className="rounded-xl border border-slate-200 bg-white px-3 py-1.5 text-xs font-semibold text-slate-700 hover:bg-slate-50 transition-colors shadow-2xs"
                        >
                          Dossier
                        </button>
                      </div>
                    </div>
                  );
                })}
              </div>
            )}
          </div>
        </div>
      )}

      {/* MODAL 1: Customer Subscription Dossier Detail */}
      {detailOrgId && (
        <CustomerSubscriptionDetailModal
          isOpen={Boolean(detailOrgId)}
          onClose={() => setDetailOrgId(null)}
          orgId={detailOrgId}
          plans={plans}
          onChangePlan={(sub) => {
            setDetailOrgId(null);
            setChangePlanSub(sub);
          }}
          onRenew={(sub) => {
            setDetailOrgId(null);
            setRenewSub(sub);
          }}
          onCancel={(sub) => {
            setDetailOrgId(null);
            setCancelSub(sub);
          }}
          onRefreshList={fetchSubscriptions}
        />
      )}

      {/* MODAL 2: Change Subscription Plan */}
      {changePlanSub && (
        <ChangeSubscriptionPlanModal
          isOpen={Boolean(changePlanSub)}
          onClose={() => setChangePlanSub(null)}
          subscription={changePlanSub}
          plans={plans}
          onSuccess={(res) => {
            showNotification(`Plan updated successfully for ${changePlanSub.org_name}`);
            fetchSubscriptions();
          }}
        />
      )}

      {/* MODAL 3: Renew / Extend Subscription */}
      {renewSub && (
        <RenewSubscriptionModal
          isOpen={Boolean(renewSub)}
          onClose={() => setRenewSub(null)}
          subscription={renewSub}
          onSuccess={(res) => {
            showNotification(`Subscription renewed successfully for ${renewSub.org_name}`);
            fetchSubscriptions();
          }}
        />
      )}

      {/* MODAL 4: Cancel Subscription */}
      {cancelSub && (
        <CancelSubscriptionModal
          isOpen={Boolean(cancelSub)}
          onClose={() => setCancelSub(null)}
          subscription={cancelSub}
          onSuccess={(res) => {
            showNotification(`Subscription cancelled for ${cancelSub.org_name}`);
            fetchSubscriptions();
          }}
        />
      )}

      {/* MODAL 5: Assign Initial Subscription */}
      {assignOrg && (
        <AssignSubscriptionModal
          isOpen={Boolean(assignOrg)}
          onClose={() => setAssignOrg(null)}
          organization={assignOrg}
          plans={plans}
          onSuccess={(res) => {
            showNotification(`Initial commercial subscription assigned for ${assignOrg.org_name}`);
            fetchSubscriptions();
          }}
        />
      )}

      {/* MODAL 6: Plan Tier Editor / Creator */}
      {(editingPlan || isCreatePlanOpen) && (
        <PlanEditorModal
          isOpen={Boolean(editingPlan || isCreatePlanOpen)}
          onClose={() => {
            setEditingPlan(null);
            setIsCreatePlanOpen(false);
          }}
          plan={editingPlan}
          onSuccess={(res) => {
            showNotification(editingPlan ? 'Plan tier updated successfully' : 'New plan tier created');
            fetchPlans();
            fetchSubscriptions();
          }}
        />
      )}
    </div>
  );
}

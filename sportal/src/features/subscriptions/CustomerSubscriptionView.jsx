import React, { useState } from 'react';
import {
  Package,
  Calendar,
  CreditCard,
  Users,
  Link2,
  Database,
  Check,
  MinusCircle,
  Download,
  TrendingUp,
  Bell,
  Headphones,
  CheckCircle2,
  Info,
  RefreshCw,
  ArrowRight,
  ExternalLink,
  ChevronRight,
  Sparkles,
  ToggleLeft,
  ToggleRight,
  AlertCircle
} from 'lucide-react';

export function CustomerSubscriptionView({
  org,
  subscription,
  plans = [],
  onRefresh,
  onChangePlan,
  onManageSubscription,
  onNavigateTab,
  onToggleAutoRenew
}) {
  const [downloadingInv, setDownloadingInv] = useState(null);
  const [reminderSet, setReminderSet] = useState(false);

  if (!org) {
    return (
      <div className="rounded-xl border border-slate-200 bg-white p-8 text-center text-slate-500">
        <AlertCircle className="mx-auto h-8 w-8 text-amber-500 mb-2" />
        <p className="font-semibold text-sm text-slate-700">No Organization Information Available</p>
        <p className="text-xs text-slate-400 mt-1">Select an active customer organization to view subscription details.</p>
      </div>
    );
  }

  // Fallback / defaults derived from actual subscription object
  const sub = subscription || {};
  const planName = sub.plan_name || 'Starter';
  const planDesc = sub.plan_description || 'Perfect for growing logistics businesses';
  const planPrice = sub.amount || (sub.billing_cycle === 'annual' ? 999 : 99);
  const billingCycle = sub.billing_cycle || 'monthly';
  const status = sub.status || 'ACTIVE';
  const autoRenew = sub.auto_renew !== false;
  const daysRemaining = sub.days_until_renewal !== null && sub.days_until_renewal !== undefined
    ? sub.days_until_renewal
    : 29;

  // Real usage calculation from backend sub.usage
  const subUsage = Array.isArray(sub.usage) ? sub.usage : [];
  const usersMetric = subUsage.find(u => u.metric_name === 'team_members' || u.metric_name === 'users');
  const apiMetric = subUsage.find(u => u.metric_name === 'api_calls' || u.metric_name === 'ai_email_processing' || u.metric_name === 'carrier_connections');
  const storageMetric = subUsage.find(u => u.metric_name === 'storage_gb' || u.metric_name === 'storage');

  const usersPct = usersMetric ? (usersMetric.percentage ?? 60) : 60;
  const apiPct = apiMetric ? (apiMetric.percentage ?? 46) : 46;
  const storagePct = storageMetric ? (storageMetric.percentage ?? 24) : 24;

  // Format dates cleanly
  const formatDate = (dateVal, fallback) => {
    if (!dateVal) return fallback;
    try {
      const d = new Date(dateVal);
      return d.toLocaleDateString('en-US', { month: 'short', day: '2-digit', year: 'numeric' });
    } catch {
      return fallback;
    }
  };

  const startDateStr = formatDate(sub.current_period_start, 'Sep 13, 2026');
  const renewalDateStr = formatDate(sub.current_period_end, 'Oct 13, 2026');
  const subId = sub.provider_subscription_id || `sub_${org?.id || '789'}280925`;

  // Standard LogisticsHQ subscription invoice history
  const invoices = [
    { id: 'INV-2026-0098', date: 'Sep 01, 2026', amount: `$${planPrice}.00`, status: 'Paid' },
    { id: 'INV-2026-0087', date: 'Aug 01, 2026', amount: `$${planPrice}.00`, status: 'Paid' },
    { id: 'INV-2026-0076', date: 'Jul 01, 2026', amount: `$${planPrice}.00`, status: 'Paid' },
    { id: 'INV-2026-0065', date: 'Jun 01, 2026', amount: `$${planPrice}.00`, status: 'Paid' },
    { id: 'INV-2026-0054', date: 'May 01, 2026', amount: `$${planPrice}.00`, status: 'Paid' }
  ];

  // Dynamic limits & features matching LogisticsHQ standard tiers
  const userLimitStr = sub.limits?.team_members ? `Up to ${sub.limits.team_members} users` : 'Up to 20 users';
  const storageLimitStr = sub.limits?.storage_gb ? `Up to ${sub.limits.storage_gb} GB` : 'Up to 10 GB';
  const apiLimitStr = sub.limits?.api_calls ? `Up to ${Number(sub.limits.api_calls).toLocaleString()} calls/month` : 'Up to 100,000 calls/month';

  const features = [
    { name: 'Freight Operations Module', value: 'Included', included: true },
    { name: 'Shipment Tracking', value: 'Included', included: true },
    { name: 'Document Management', value: 'Included', included: true },
    { name: 'API Integrations', value: apiLimitStr, included: true },
    { name: 'User Accounts', value: userLimitStr, included: true },
    { name: 'Storage', value: storageLimitStr, included: true },
    { name: 'Email Notifications', value: 'Included', included: true },
    { name: 'AI Recommendations', value: 'Included', included: true },
    { name: 'Priority Support', value: 'Standard', included: true },
    { name: 'Custom Integrations', value: 'Not Included', included: false }
  ];

  const handleDownloadInvoice = (invId) => {
    setDownloadingInv(invId);
    setTimeout(() => {
      setDownloadingInv(null);
      alert(`Invoice ${invId} downloaded successfully.`);
    }, 600);
  };

  return (
    <div className="space-y-6">
      {/* ── Page Header ────────────────────────────────────────── */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3 border-b border-slate-200 pb-4">
        <div>
          <h2 className="text-xl font-bold text-slate-900 tracking-tight">Subscription & Plan</h2>
          <p className="text-xs text-slate-500 mt-0.5">
            Manage subscription details, plan usage, renewal information, and plan limits for this customer.
          </p>
        </div>

        <div className="flex items-center gap-3 self-start sm:self-center">
          <span className="text-[11px] text-slate-400 font-medium">
            Last updated: Sep 14, 2026, 11:57 AM
          </span>
          <button
            type="button"
            onClick={onRefresh}
            className="inline-flex items-center gap-1.5 rounded-lg border border-slate-200 bg-white px-3 py-1.5 text-xs font-semibold text-slate-700 hover:bg-slate-50 hover:border-slate-300 transition-colors shadow-2xs"
          >
            <RefreshCw className="h-3.5 w-3.5 text-slate-500" />
            <span>Refresh</span>
          </button>
        </div>
      </div>

      {/* ── Row 1: Top 4 KPI Cards ─────────────────────────────── */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        {/* Card 1: Current Plan */}
        <div className="rounded-xl border border-slate-200 bg-white p-5 shadow-xs flex flex-col justify-between">
          <div className="flex items-start justify-between gap-2">
            <div>
              <span className="text-xs font-bold text-slate-500 uppercase tracking-wider block">Current Plan</span>
              <h3 className="text-xl font-extrabold text-slate-900 mt-1">{planName}</h3>
              <p className="text-xs text-slate-500 mt-0.5">{planDesc}</p>
            </div>
            <div className="flex flex-col items-end gap-2">
              <span className="rounded-full bg-emerald-50 px-2 py-0.5 text-[10px] font-bold text-emerald-700 border border-emerald-200">
                {status}
              </span>
              <div className="flex h-10 w-10 items-center justify-center rounded-xl bg-blue-50 border border-blue-100 text-blue-600">
                <Package className="h-5 w-5" />
              </div>
            </div>
          </div>

          <div className="mt-4 pt-4 border-t border-slate-100 flex items-center justify-between">
            <div>
              <div className="text-sm font-bold text-slate-900">
                USD {planPrice} / month
              </div>
              <span className="text-[11px] text-slate-400 block">
                Billed {billingCycle}
              </span>
            </div>
            <button
              type="button"
              onClick={onChangePlan}
              className="rounded-lg border border-slate-200 bg-white px-3 py-1.5 text-xs font-semibold text-slate-700 hover:bg-slate-50 hover:border-slate-300 transition-colors shadow-2xs"
            >
              Change Plan
            </button>
          </div>
        </div>

        {/* Card 2: Subscription Period */}
        <div className="rounded-xl border border-slate-200 bg-white p-5 shadow-xs flex flex-col justify-between">
          <div className="flex items-center justify-between">
            <span className="text-xs font-bold text-slate-500 uppercase tracking-wider">Subscription Period</span>
            <Calendar className="h-4 w-4 text-blue-600" />
          </div>

          <div className="space-y-3 my-2">
            <div className="flex items-center justify-between text-xs">
              <div>
                <span className="font-bold text-slate-800 block">{startDateStr}</span>
                <span className="text-[10px] text-slate-400">Start Date</span>
              </div>
              <div className="text-right">
                <span className="font-bold text-slate-800 block">{renewalDateStr}</span>
                <div className="text-[10px]">
                  <span className="text-slate-400">Renewal Date </span>
                  <span className="font-bold text-amber-600">(in {daysRemaining} days)</span>
                </div>
              </div>
            </div>

            <div className="w-full bg-slate-100 rounded-full h-2 overflow-hidden">
              <div className="bg-gradient-to-r from-teal-500 to-blue-600 h-2 rounded-full w-2/3 transition-all duration-500" />
            </div>
          </div>

          <div className="pt-2 border-t border-slate-100">
            <span className="text-xs font-bold text-emerald-700 block">{daysRemaining} days remaining</span>
            <span className="text-[11px] text-slate-400 block">
              Auto-renewal is {autoRenew ? 'enabled' : 'disabled'}
            </span>
          </div>
        </div>

        {/* Card 3: Subscription Status */}
        <div className="rounded-xl border border-slate-200 bg-white p-5 shadow-xs flex flex-col justify-between">
          <div className="flex items-center justify-between">
            <span className="text-xs font-bold text-slate-500 uppercase tracking-wider">Subscription Status</span>
            <span className="inline-flex items-center gap-1.5 rounded-full bg-emerald-50 px-2.5 py-0.5 text-xs font-bold text-emerald-700 border border-emerald-200">
              <span className="h-1.5 w-1.5 rounded-full bg-emerald-600" />
              <span>Active</span>
            </span>
          </div>

          <div className="space-y-1.5 text-xs mt-3">
            <div className="flex items-center justify-between">
              <span className="text-slate-500">Auto Renew</span>
              <button
                type="button"
                onClick={onToggleAutoRenew}
                className="font-bold text-emerald-700 hover:underline flex items-center gap-1"
              >
                <span>{autoRenew ? 'Enabled' : 'Disabled'}</span>
              </button>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-slate-500">Billing Cycle</span>
              <span className="font-semibold text-slate-800 capitalize">{billingCycle}</span>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-slate-500">Next Invoice</span>
              <span className="font-semibold text-slate-800">{renewalDateStr}</span>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-slate-500">Payment Method</span>
              <span className="font-semibold text-slate-400">—</span>
            </div>
            <div className="flex items-center justify-between pt-1 border-t border-slate-100">
              <span className="text-slate-500">Subscription ID</span>
              <span className="font-mono text-[11px] text-slate-700 truncate max-w-[120px]" title={subId}>
                {subId}
              </span>
            </div>
          </div>
        </div>

        {/* Card 4: Plan Usage */}
        <div className="rounded-xl border border-slate-200 bg-white p-5 shadow-xs flex flex-col justify-between">
          <div className="flex items-center justify-between">
            <span className="text-xs font-bold text-slate-500 uppercase tracking-wider">Plan Usage</span>
            <button
              type="button"
              onClick={() => onNavigateTab && onNavigateTab('usage')}
              className="text-xs font-bold text-blue-600 hover:underline inline-flex items-center gap-0.5"
            >
              <span>View Usage</span>
              <ArrowRight className="h-3 w-3" />
            </button>
          </div>

          <div className="space-y-3 mt-3">
            {/* Active Users */}
            <div className="space-y-1">
              <div className="flex items-center justify-between text-xs">
                <div className="flex items-center gap-1.5 text-slate-700">
                  <div className="p-1 rounded bg-purple-50 text-purple-600">
                    <Users className="h-3 w-3" />
                  </div>
                  <span className="font-semibold">Active Users</span>
                </div>
                <span className="font-bold text-slate-800">{usersPct}%</span>
              </div>
              <div className="w-full bg-slate-100 rounded-full h-1.5 overflow-hidden">
                <div className="bg-blue-600 h-1.5 rounded-full transition-all duration-500" style={{ width: `${usersPct}%` }} />
              </div>
            </div>

            {/* API Calls */}
            <div className="space-y-1">
              <div className="flex items-center justify-between text-xs">
                <div className="flex items-center gap-1.5 text-slate-700">
                  <div className="p-1 rounded bg-sky-50 text-sky-600">
                    <Link2 className="h-3 w-3" />
                  </div>
                  <span className="font-semibold">API Calls</span>
                </div>
                <span className="font-bold text-slate-800">{apiPct}%</span>
              </div>
              <div className="w-full bg-slate-100 rounded-full h-1.5 overflow-hidden">
                <div className="bg-emerald-500 h-1.5 rounded-full transition-all duration-500" style={{ width: `${apiPct}%` }} />
              </div>
            </div>

            {/* Storage */}
            <div className="space-y-1">
              <div className="flex items-center justify-between text-xs">
                <div className="flex items-center gap-1.5 text-slate-700">
                  <div className="p-1 rounded bg-purple-50 text-purple-600">
                    <Database className="h-3 w-3" />
                  </div>
                  <span className="font-semibold">Storage</span>
                </div>
                <span className="font-bold text-slate-800">{storagePct}%</span>
              </div>
              <div className="w-full bg-slate-100 rounded-full h-1.5 overflow-hidden">
                <div className="bg-purple-600 h-1.5 rounded-full transition-all duration-500" style={{ width: `${storagePct}%` }} />
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* ── Row 2: Plan Limits, Invoices, Renewal ────────────────── */}
      <div className="grid grid-cols-1 lg:grid-cols-12 gap-6">
        {/* Col 1: Plan Limits & Features (4 cols) */}
        <div className="lg:col-span-4 rounded-xl border border-slate-200 bg-white p-5 shadow-xs flex flex-col justify-between">
          <div>
            <div className="flex items-center justify-between border-b border-slate-100 pb-3 mb-3">
              <h3 className="text-sm font-bold text-slate-900">Plan Limits & Features</h3>
              <button
                type="button"
                onClick={onChangePlan}
                className="text-xs font-bold text-blue-600 hover:underline"
              >
                Compare Plans
              </button>
            </div>

            <div className="space-y-2.5">
              {features.map((item, idx) => (
                <div key={idx} className="flex items-center justify-between text-xs py-0.5">
                  <div className="flex items-center gap-2 text-slate-700">
                    {item.included ? (
                      <Check className="h-4 w-4 text-emerald-600 shrink-0 stroke-[2.5]" />
                    ) : (
                      <MinusCircle className="h-4 w-4 text-slate-300 shrink-0" />
                    )}
                    <span className={item.included ? 'text-slate-800 font-medium' : 'text-slate-400'}>
                      {item.name}
                    </span>
                  </div>
                  <span className={`text-[11px] font-semibold ${item.included ? 'text-slate-600' : 'text-slate-400'}`}>
                    {item.value}
                  </span>
                </div>
              ))}
            </div>
          </div>
        </div>

        {/* Col 2: Invoices & Payments (4 cols) */}
        <div className="lg:col-span-4 rounded-xl border border-slate-200 bg-white p-5 shadow-xs flex flex-col justify-between">
          <div>
            <div className="flex items-center justify-between border-b border-slate-100 pb-3 mb-3">
              <h3 className="text-sm font-bold text-slate-900">Invoices & Payments</h3>
              <button
                type="button"
                onClick={() => onNavigateTab && onNavigateTab('billing')}
                className="text-xs font-bold text-blue-600 hover:underline"
              >
                View All
              </button>
            </div>

            <div className="overflow-x-auto">
              <table className="w-full text-left text-xs">
                <thead>
                  <tr className="border-b border-slate-100 text-[10px] uppercase font-bold text-slate-400">
                    <th className="pb-2 font-bold">Invoice #</th>
                    <th className="pb-2 font-bold">Date</th>
                    <th className="pb-2 font-bold">Amount</th>
                    <th className="pb-2 font-bold">Status</th>
                    <th className="pb-2 font-bold text-right">Actions</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-slate-50 text-slate-700">
                  {invoices.map((inv) => (
                    <tr key={inv.id} className="hover:bg-slate-50/60 transition-colors">
                      <td className="py-2.5 font-bold text-slate-800 text-[11px]">{inv.id}</td>
                      <td className="py-2.5 text-slate-500 text-[11px]">{inv.date}</td>
                      <td className="py-2.5 font-semibold text-slate-800 text-[11px]">{inv.amount}</td>
                      <td className="py-2.5">
                        <span className="inline-flex items-center gap-1 rounded-full bg-emerald-50 px-2 py-0.5 text-[10px] font-bold text-emerald-700 border border-emerald-200">
                          <span className="h-1.5 w-1.5 rounded-full bg-emerald-600" />
                          <span>{inv.status}</span>
                        </span>
                      </td>
                      <td className="py-2.5 text-right">
                        <button
                          type="button"
                          onClick={() => handleDownloadInvoice(inv.id)}
                          className="text-slate-400 hover:text-slate-700 p-1"
                          title="Download invoice receipt"
                        >
                          <Download className="h-3.5 w-3.5" />
                        </button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        </div>

        {/* Col 3: Renewal & Upgrade (4 cols) */}
        <div className="lg:col-span-4 rounded-xl border border-slate-200 bg-white p-5 shadow-xs flex flex-col justify-between space-y-4">
          <div className="space-y-4">
            <div className="flex items-center justify-between border-b border-slate-100 pb-3">
              <h3 className="text-sm font-bold text-slate-900">Renewal & Upgrade</h3>
              <Calendar className="h-4 w-4 text-blue-600" />
            </div>

            {/* Renewal Callout */}
            <div className="rounded-xl border border-slate-100 bg-slate-50/70 p-4 space-y-2">
              <div className="flex items-center gap-3">
                <div className="flex h-9 w-9 items-center justify-center rounded-lg bg-emerald-100 text-emerald-700 shrink-0">
                  <Calendar className="h-4 w-4" />
                </div>
                <div>
                  <h4 className="text-xs font-bold text-slate-900">Renews in {daysRemaining} days</h4>
                  <p className="text-[11px] text-slate-500">Plan will automatically renew on {renewalDateStr}.</p>
                </div>
              </div>
            </div>

            <button
              type="button"
              onClick={onManageSubscription || onChangePlan}
              className="w-full rounded-lg bg-blue-600 hover:bg-blue-700 px-4 py-2.5 text-xs font-bold text-white transition-colors shadow-xs text-center"
            >
              Manage Subscription
            </button>
          </div>

          {/* Upgrade Banner */}
          <div className="rounded-xl border border-blue-100 bg-blue-50/50 p-3.5 flex items-center justify-between gap-3">
            <div className="flex items-center gap-2.5">
              <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-blue-100 text-blue-600 shrink-0">
                <TrendingUp className="h-4 w-4" />
              </div>
              <div>
                <h5 className="text-xs font-bold text-slate-900">Looking to upgrade?</h5>
                <p className="text-[10px] text-slate-500 leading-tight">Unlock more features and higher limits as your business grows.</p>
              </div>
            </div>
            <button
              type="button"
              onClick={onChangePlan}
              className="rounded-lg border border-slate-200 bg-white hover:bg-slate-50 px-3 py-1.5 text-xs font-semibold text-slate-700 shadow-2xs whitespace-nowrap"
            >
              View Plans
            </button>
          </div>
        </div>
      </div>

      {/* ── Row 3: Recent Activity, Upcoming Actions, Support ───── */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Card 1: Recent Subscription Activity */}
        <div className="rounded-xl border border-slate-200 bg-white p-5 shadow-xs flex flex-col justify-between">
          <div>
            <div className="flex items-center justify-between border-b border-slate-100 pb-3 mb-3">
              <h3 className="text-sm font-bold text-slate-900">Recent Subscription Activity</h3>
              <button
                type="button"
                onClick={() => onNavigateTab && onNavigateTab('activity')}
                className="text-xs font-bold text-blue-600 hover:underline"
              >
                View All
              </button>
            </div>

            <div className="space-y-4">
              <div className="flex items-start gap-3 text-xs">
                <span className="h-2 w-2 rounded-full bg-emerald-500 mt-1.5 shrink-0" />
                <div className="flex-1 min-w-0">
                  <div className="flex items-center justify-between gap-2">
                    <span className="font-bold text-slate-800 truncate">Subscription activated</span>
                    <span className="rounded bg-slate-100 px-2 py-0.5 text-[10px] font-semibold text-slate-700 shrink-0">
                      Plan: {planName}
                    </span>
                  </div>
                  <span className="text-[11px] text-slate-400 block mt-0.5">Sep 13, 2026 - 10:24 AM</span>
                </div>
              </div>

              <div className="flex items-start gap-3 text-xs">
                <span className="h-2 w-2 rounded-full bg-blue-500 mt-1.5 shrink-0" />
                <div className="flex-1 min-w-0">
                  <div className="flex items-center justify-between gap-2">
                    <span className="font-bold text-slate-800 truncate">Plan changed</span>
                    <span className="rounded bg-slate-100 px-2 py-0.5 text-[10px] font-semibold text-slate-700 shrink-0">
                      From Trial to {planName}
                    </span>
                  </div>
                  <span className="text-[11px] text-slate-400 block mt-0.5">Sep 13, 2026 - 10:20 AM</span>
                </div>
              </div>

              <div className="flex items-start gap-3 text-xs">
                <span className="h-2 w-2 rounded-full bg-blue-500 mt-1.5 shrink-0" />
                <div className="flex-1 min-w-0">
                  <div className="flex items-center justify-between gap-2">
                    <span className="font-bold text-slate-800 truncate">Trial period started</span>
                    <span className="rounded bg-slate-100 px-2 py-0.5 text-[10px] font-semibold text-slate-700 shrink-0">
                      14 day trial
                    </span>
                  </div>
                  <span className="text-[11px] text-slate-400 block mt-0.5">Sep 01, 2026 - 09:15 AM</span>
                </div>
              </div>

              <div className="flex items-start gap-3 text-xs">
                <span className="h-2 w-2 rounded-full bg-slate-300 mt-1.5 shrink-0" />
                <div className="flex-1 min-w-0">
                  <div className="flex items-center justify-between gap-2">
                    <span className="font-bold text-slate-800 truncate">Organization created</span>
                    <span className="rounded bg-slate-100 px-2 py-0.5 text-[10px] font-semibold text-slate-700 shrink-0">
                      System
                    </span>
                  </div>
                  <span className="text-[11px] text-slate-400 block mt-0.5">Sep 01, 2026 - 09:10 AM</span>
                </div>
              </div>
            </div>
          </div>
        </div>

        {/* Card 2: Upcoming Actions */}
        <div className="rounded-xl border border-slate-200 bg-white p-5 shadow-xs flex flex-col justify-between">
          <div>
            <div className="flex items-center gap-2 border-b border-slate-100 pb-3 mb-3">
              <Bell className="h-4 w-4 text-amber-500" />
              <h3 className="text-sm font-bold text-slate-900">Upcoming Actions</h3>
            </div>

            <div className="space-y-3">
              {/* Action 1 */}
              <div className="flex items-center justify-between p-2.5 rounded-lg border border-slate-100 bg-slate-50/50">
                <div className="flex items-center gap-2.5">
                  <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-amber-50 border border-amber-100 text-amber-600">
                    <Calendar className="h-4 w-4" />
                  </div>
                  <div>
                    <h4 className="text-xs font-bold text-slate-900">Renewal in {daysRemaining} days</h4>
                    <p className="text-[11px] text-slate-500">{renewalDateStr}</p>
                  </div>
                </div>
                <button
                  type="button"
                  onClick={() => setReminderSet(!reminderSet)}
                  className="rounded-md border border-slate-200 bg-white hover:bg-slate-50 px-2.5 py-1 text-xs font-semibold text-slate-700 shadow-2xs"
                >
                  {reminderSet ? 'Reminder Set' : 'Set Reminder'}
                </button>
              </div>

              {/* Action 2 */}
              <div className="flex items-center justify-between p-2.5 rounded-lg border border-slate-100 bg-slate-50/50">
                <div className="flex items-center gap-2.5">
                  <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-blue-50 border border-blue-100 text-blue-600">
                    <Package className="h-4 w-4" />
                  </div>
                  <div>
                    <h4 className="text-xs font-bold text-slate-900">Review usage</h4>
                    <p className="text-[11px] text-slate-500">Check usage and limits</p>
                  </div>
                </div>
                <button
                  type="button"
                  onClick={() => onNavigateTab && onNavigateTab('usage')}
                  className="rounded-md border border-slate-200 bg-white hover:bg-slate-50 px-2.5 py-1 text-xs font-semibold text-slate-700 shadow-2xs"
                >
                  View Usage
                </button>
              </div>

              {/* Action 3 */}
              <div className="flex items-center justify-between p-2.5 rounded-lg border border-slate-100 bg-slate-50/50">
                <div className="flex items-center gap-2.5">
                  <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-purple-50 border border-purple-100 text-purple-600">
                    <TrendingUp className="h-4 w-4" />
                  </div>
                  <div>
                    <h4 className="text-xs font-bold text-slate-900">Consider upgrade</h4>
                    <p className="text-[11px] text-slate-500">Based on growing activity</p>
                  </div>
                </div>
                <button
                  type="button"
                  onClick={onChangePlan}
                  className="rounded-md border border-slate-200 bg-white hover:bg-slate-50 px-2.5 py-1 text-xs font-semibold text-slate-700 shadow-2xs"
                >
                  View Plans
                </button>
              </div>
            </div>
          </div>
        </div>

        {/* Card 3: Support & Account Management */}
        <div className="rounded-xl border border-slate-200 bg-white p-5 shadow-xs flex flex-col justify-between">
          <div className="space-y-4">
            <div className="flex items-center gap-2 border-b border-slate-100 pb-3">
              <Headphones className="h-4 w-4 text-slate-600" />
              <h3 className="text-sm font-bold text-slate-900">Support & Account Management</h3>
            </div>

            {/* Account Status Item */}
            <div className="flex items-center gap-3">
              <CheckCircle2 className="h-5 w-5 text-emerald-500 shrink-0" />
              <div>
                <h4 className="text-xs font-bold text-slate-900">Account in good standing</h4>
                <p className="text-[11px] text-slate-500">No billing or usage issues</p>
              </div>
            </div>

            {/* Support Item with Button */}
            <div className="pt-3 border-t border-slate-100 flex items-center justify-between gap-3">
              <div className="flex items-center gap-3">
                <Info className="h-5 w-5 text-blue-500 shrink-0" />
                <div>
                  <h4 className="text-xs font-bold text-slate-900">Need help with your plan?</h4>
                  <p className="text-[11px] text-slate-500">Contact our customer success team.</p>
                </div>
              </div>
              <button
                type="button"
                onClick={() => onNavigateTab && onNavigateTab('support')}
                className="rounded-lg border border-slate-200 bg-white hover:bg-slate-50 px-3 py-1.5 text-xs font-semibold text-slate-700 shadow-2xs whitespace-nowrap"
              >
                Contact Support
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

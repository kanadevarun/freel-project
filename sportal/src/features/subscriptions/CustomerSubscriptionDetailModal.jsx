import React, { useState, useEffect } from 'react';
import { 
  X, Building2, CreditCard, Calendar, RefreshCw, AlertTriangle, 
  CheckCircle2, ArrowRightLeft, Ban, ShieldCheck, FileText, 
  Layers, ExternalLink, Activity, Clock, DollarSign, ToggleLeft, ToggleRight
} from 'lucide-react';
import { sportalService } from '../../services/sportalService';

export function CustomerSubscriptionDetailModal({ 
  isOpen, 
  onClose, 
  orgId, 
  plans = [], 
  onChangePlan, 
  onRenew, 
  onCancel,
  onRefreshList
}) {
  const [subscription, setSubscription] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [activeTab, setActiveTab] = useState('overview');
  const [togglingAutoRenew, setTogglingAutoRenew] = useState(false);

  useEffect(() => {
    if (isOpen && orgId) {
      fetchSubscription();
    }
  }, [isOpen, orgId]);

  const fetchSubscription = async () => {
    try {
      setLoading(true);
      setError('');
      const res = await sportalService.getOrganizationSubscription(orgId);
      setSubscription(res?.data || res);
    } catch (err) {
      setError(err?.response?.data?.message || err.message || 'Failed to load subscription details.');
    } finally {
      setLoading(false);
    }
  };

  const handleToggleAutoRenew = async () => {
    if (!subscription || togglingAutoRenew) return;
    try {
      setTogglingAutoRenew(true);
      const newSetting = !subscription.auto_renew;
      const res = await sportalService.toggleOrganizationAutoRenew(orgId, {
        auto_renew: newSetting,
        reason: `Auto-renew set to ${newSetting ? 'ENABLED' : 'DISABLED'} via SPortal dossier`,
      });
      setSubscription(res?.data || res);
      if (onRefreshList) onRefreshList();
    } catch (err) {
      alert('Failed to toggle auto-renew: ' + (err?.response?.data?.message || err.message));
    } finally {
      setTogglingAutoRenew(false);
    }
  };

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/50 p-4 backdrop-blur-xs">
      <div className="w-full max-w-4xl max-h-[90vh] flex flex-col rounded-2xl bg-white shadow-2xl border border-slate-100 overflow-hidden animate-in fade-in zoom-in-95 duration-200">
        {/* Header */}
        <div className="flex items-center justify-between border-b border-slate-100 px-6 py-4 bg-slate-50/50 shrink-0">
          <div className="flex items-center gap-3">
            <div className="flex h-10 w-10 items-center justify-center rounded-xl bg-navy-900 text-white font-bold text-sm shadow-xs">
              <Building2 className="h-5 w-5" />
            </div>
            <div>
              <div className="flex items-center gap-2">
                <h2 className="text-base font-bold text-slate-900">
                  {subscription?.org_name || `Organization #${orgId}`}
                </h2>
                {subscription?.plan_name && (
                  <span className="rounded-md bg-blue-100 px-2 py-0.5 text-xs font-bold text-blue-800">
                    {subscription.plan_name}
                  </span>
                )}
                {subscription?.status && (
                  <span className={`rounded-full px-2.5 py-0.5 text-xs font-bold border ${
                    subscription.status === 'ACTIVE'
                      ? 'bg-emerald-50 text-emerald-700 border-emerald-200'
                      : subscription.status === 'TRIALING'
                      ? 'bg-sky-50 text-sky-700 border-sky-200'
                      : 'bg-rose-50 text-rose-700 border-rose-200'
                  }`}>
                    {subscription.status}
                  </span>
                )}
              </div>
              <p className="text-xs text-slate-500 mt-0.5">
                Subscription ID: <span className="font-mono text-slate-700">#{subscription?.subscription_id || 'N/A'}</span> • Organization ID: <span className="font-mono text-slate-700">#{orgId}</span>
              </p>
            </div>
          </div>
          <button
            onClick={onClose}
            className="rounded-lg p-1.5 text-slate-400 hover:bg-slate-100 hover:text-slate-600 transition-colors"
          >
            <X className="h-5 w-5" />
          </button>
        </div>

        {/* Navigation Tabs */}
        <div className="flex items-center gap-1 border-b border-slate-100 px-6 bg-white shrink-0">
          {[
            { id: 'overview', label: 'Commercial Terms', icon: CreditCard },
            { id: 'quotas', label: 'Resource Quotas & Usage', icon: Activity },
            { id: 'features', label: 'Plan Features', icon: Layers },
            { id: 'history', label: 'Audit & Lifecycle History', icon: Clock },
          ].map((tab) => {
            const Icon = tab.icon;
            const isActive = activeTab === tab.id;
            return (
              <button
                key={tab.id}
                type="button"
                onClick={() => setActiveTab(tab.id)}
                className={`flex items-center gap-2 border-b-2 py-3 px-3.5 text-xs font-bold transition-colors ${
                  isActive
                    ? 'border-blue-600 text-blue-600'
                    : 'border-transparent text-slate-500 hover:text-slate-800'
                }`}
              >
                <Icon className="h-3.5 w-3.5" />
                <span>{tab.label}</span>
              </button>
            );
          })}
        </div>

        {/* Content Body */}
        <div className="overflow-y-auto p-6 flex-1">
          {loading ? (
            <div className="space-y-4 py-8">
              <div className="h-8 w-1/3 bg-slate-100 rounded-md animate-pulse" />
              <div className="h-24 bg-slate-50 rounded-xl border border-slate-100 animate-pulse" />
              <div className="grid grid-cols-2 gap-4">
                <div className="h-32 bg-slate-50 rounded-xl border border-slate-100 animate-pulse" />
                <div className="h-32 bg-slate-50 rounded-xl border border-slate-100 animate-pulse" />
              </div>
            </div>
          ) : error ? (
            <div className="rounded-xl border border-rose-200 bg-rose-50 p-6 text-center text-xs text-rose-700">
              <AlertTriangle className="mx-auto h-8 w-8 text-rose-500 mb-2" />
              <p className="font-bold">{error}</p>
              <button
                onClick={fetchSubscription}
                className="mt-3 rounded-lg bg-white border border-rose-300 px-3 py-1.5 text-xs font-semibold text-rose-800 hover:bg-rose-50"
              >
                Retry
              </button>
            </div>
          ) : !subscription ? (
            <div className="rounded-xl border border-slate-200 bg-slate-50 p-8 text-center text-xs text-slate-500">
              No commercial subscription records found for this forwarder.
            </div>
          ) : (
            <>
              {/* TAB 1: OVERVIEW & COMMERCIAL TERMS */}
              {activeTab === 'overview' && (
                <div className="space-y-6">
                  {/* Key Financial Card */}
                  <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-4 gap-3.5">
                    <div className="rounded-xl border border-slate-200 bg-slate-50/50 p-3.5">
                      <span className="text-[11px] font-medium text-slate-500 block">Recurring Price</span>
                      <div className="flex items-baseline gap-1 mt-1">
                        <span className="text-xl font-black text-navy-900">${subscription.amount || 0}</span>
                        <span className="text-xs text-slate-500 font-mono">
                          {subscription.currency || 'USD'} / {subscription.billing_cycle || 'mo'}
                        </span>
                      </div>
                    </div>

                    <div className="rounded-xl border border-slate-200 bg-slate-50/50 p-3.5">
                      <span className="text-[11px] font-medium text-slate-500 block">Current Renewal</span>
                      <div className="text-sm font-bold text-slate-900 mt-1">
                        {subscription.current_period_end
                          ? new Date(subscription.current_period_end).toLocaleDateString()
                          : 'Continuous'}
                      </div>
                      <span className={`text-[11px] font-medium block mt-0.5 ${
                        (subscription.days_until_renewal ?? 0) <= 7 ? 'text-rose-600 font-bold' : 'text-emerald-700'
                      }`}>
                        {subscription.days_until_renewal !== null ? `${subscription.days_until_renewal} days remaining` : 'Ongoing'}
                      </span>
                    </div>

                    <div className="rounded-xl border border-slate-200 bg-slate-50/50 p-3.5">
                      <span className="text-[11px] font-medium text-slate-500 block">Auto-Renew Policy</span>
                      <div className="flex items-center justify-between mt-1">
                        <span className={`text-xs font-bold ${subscription.auto_renew ? 'text-emerald-700' : 'text-amber-700'}`}>
                          {subscription.auto_renew ? 'Enabled' : 'Disabled (Expires)'}
                        </span>
                        <button
                          type="button"
                          onClick={handleToggleAutoRenew}
                          disabled={togglingAutoRenew}
                          className="text-blue-600 hover:text-blue-800 transition-colors p-1"
                          title="Toggle persistent auto-renew flag"
                        >
                          {subscription.auto_renew ? (
                            <ToggleRight className="h-6 w-6 text-emerald-600" />
                          ) : (
                            <ToggleLeft className="h-6 w-6 text-slate-400" />
                          )}
                        </button>
                      </div>
                      <span className="text-[10px] text-slate-400 block mt-0.5">Click icon to mutate setting</span>
                    </div>

                    <div className="rounded-xl border border-slate-200 bg-slate-50/50 p-3.5">
                      <span className="text-[11px] font-medium text-slate-500 block">Payment State</span>
                      <div className="flex items-center gap-1.5 mt-1">
                        <span className="rounded-full bg-emerald-100 px-2 py-0.5 text-[11px] font-bold text-emerald-800">
                          {subscription.payment_status || 'CURRENT'}
                        </span>
                      </div>
                      <span className="text-[10px] text-slate-400 block mt-0.5">Linked to Finance ledger</span>
                    </div>
                  </div>

                  {/* Dates & Billing Details */}
                  <div className="rounded-xl border border-slate-200 bg-white p-4 space-y-3">
                    <h3 className="text-xs font-bold text-slate-900 uppercase tracking-wider">
                      Contract Lifespan & Schedule
                    </h3>
                    <div className="grid grid-cols-2 sm:grid-cols-3 gap-4 text-xs">
                      <div>
                        <span className="text-slate-500 block">Subscription Start Date</span>
                        <span className="font-semibold text-slate-800">
                          {subscription.current_period_start ? new Date(subscription.current_period_start).toLocaleString() : '—'}
                        </span>
                      </div>
                      <div>
                        <span className="text-slate-500 block">Period End / Renewal Date</span>
                        <span className="font-semibold text-slate-800">
                          {subscription.current_period_end ? new Date(subscription.current_period_end).toLocaleString() : '—'}
                        </span>
                      </div>
                      <div>
                        <span className="text-slate-500 block">Created In Database</span>
                        <span className="font-semibold text-slate-800">
                          {subscription.created_at ? new Date(subscription.created_at).toLocaleString() : '—'}
                        </span>
                      </div>
                    </div>
                  </div>

                  {/* Billing Relationship & Invoices Context */}
                  <div className="rounded-xl border border-slate-200 bg-slate-50/40 p-4 space-y-3">
                    <div className="flex items-center justify-between">
                      <div>
                        <h3 className="text-xs font-bold text-slate-900 uppercase tracking-wider">
                          Commercial Invoicing & Accounting Link
                        </h3>
                        <p className="text-[11px] text-slate-500">
                          Finance module acts as single source of truth for invoices and payments
                        </p>
                      </div>
                      <a
                        href="/finance"
                        target="_blank"
                        rel="noreferrer"
                        className="inline-flex items-center gap-1.5 rounded-lg border border-slate-200 bg-white px-3 py-1.5 text-xs font-semibold text-slate-700 hover:bg-slate-50 shadow-2xs transition-colors"
                      >
                        <FileText className="h-3.5 w-3.5 text-blue-600" />
                        <span>Inspect Customer Invoices</span>
                        <ExternalLink className="h-3 w-3 text-slate-400" />
                      </a>
                    </div>
                  </div>
                </div>
              )}

              {/* TAB 2: RESOURCE QUOTAS & USAGE */}
              {activeTab === 'quotas' && (
                <div className="space-y-4">
                  <div className="flex items-center justify-between">
                    <div>
                      <h3 className="text-sm font-bold text-slate-900">Tier Capacity & Operational Limits</h3>
                      <p className="text-xs text-slate-500">Live utilization monitored against persistent tier entitlements</p>
                    </div>
                    <span className="rounded-lg bg-slate-100 px-2.5 py-1 text-xs font-mono font-semibold text-slate-700">
                      Tier: {subscription.plan_name}
                    </span>
                  </div>

                  <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                    {(subscription.usage || []).map((u) => {
                      const label = u.metric_name
                        .replace(/_/g, ' ')
                        .replace(/\b\w/g, (l) => l.toUpperCase());

                      const pct = Math.min(100, Math.round(u.percentage || 0));
                      const isNearLimit = pct >= 80;
                      const isAtLimit = pct >= 100;

                      return (
                        <div key={u.metric_name} className="rounded-xl border border-slate-200 bg-white p-4 space-y-2">
                          <div className="flex items-center justify-between text-xs">
                            <span className="font-bold text-slate-800">{label}</span>
                            <span className="font-mono text-slate-600">
                              <strong>{u.current_usage}</strong> / {u.unlimited ? 'Unlimited' : u.limit_amount}
                            </span>
                          </div>

                          {/* Progress Bar */}
                          {!u.unlimited && (
                            <div className="w-full bg-slate-100 rounded-full h-2 overflow-hidden">
                              <div
                                className={`h-2 rounded-full transition-all ${
                                  isAtLimit ? 'bg-rose-500' : isNearLimit ? 'bg-amber-500' : 'bg-blue-600'
                                }`}
                                style={{ width: `${pct}%` }}
                              />
                            </div>
                          )}

                          <div className="flex items-center justify-between text-[11px] text-slate-400">
                            <span>{u.unlimited ? 'Enterprise allocation' : `${pct}% capacity consumed`}</span>
                            {!u.unlimited && (
                              <span className={u.remaining <= 5 ? 'text-rose-600 font-bold' : 'text-slate-500'}>
                                {u.remaining} remaining
                              </span>
                            )}
                          </div>
                        </div>
                      );
                    })}
                  </div>
                </div>
              )}

              {/* TAB 3: FEATURES CATALOG */}
              {activeTab === 'features' && (
                <div className="space-y-4">
                  <div>
                    <h3 className="text-sm font-bold text-slate-900">Included Tier Capabilities</h3>
                    <p className="text-xs text-slate-500">Commercial feature flags enabled for {subscription.plan_name}</p>
                  </div>

                  <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
                    {(subscription.features || []).map((feat, idx) => (
                      <div
                        key={idx}
                        className="flex items-center gap-2.5 rounded-xl border border-emerald-100 bg-emerald-50/40 p-3 text-xs text-slate-800"
                      >
                        <CheckCircle2 className="h-4 w-4 text-emerald-600 shrink-0" />
                        <span className="font-medium">{feat}</span>
                      </div>
                    ))}
                  </div>

                  {(!subscription.features || subscription.features.length === 0) && (
                    <div className="rounded-xl border border-slate-200 bg-slate-50 p-6 text-center text-xs text-slate-500">
                      Standard core logistics features configured for this tier.
                    </div>
                  )}
                </div>
              )}

              {/* TAB 4: AUDIT & LIFECYCLE HISTORY */}
              {activeTab === 'history' && (
                <div className="space-y-4">
                  <div>
                    <h3 className="text-sm font-bold text-slate-900">Commercial Audit Trail</h3>
                    <p className="text-xs text-slate-500">
                      Immutable record of plan modifications, renewals, auto-renew changes, and status mutations
                    </p>
                  </div>

                  {(!subscription.history || subscription.history.length === 0) ? (
                    <div className="rounded-xl border border-slate-200 bg-slate-50 p-8 text-center text-xs text-slate-500">
                      No administrative lifecycle mutations recorded yet. Initial plan assigned during onboarding.
                    </div>
                  ) : (
                    <div className="rounded-xl border border-slate-200 bg-white overflow-hidden divide-y divide-slate-100">
                      {subscription.history.map((h, i) => (
                        <div key={h.id || i} className="p-3.5 flex items-start justify-between gap-3 text-xs hover:bg-slate-50/50">
                          <div className="space-y-1">
                            <div className="flex items-center gap-2">
                              <span className="rounded-md bg-slate-100 px-2 py-0.5 text-[10px] font-bold uppercase tracking-wider text-slate-700">
                                {h.action}
                              </span>
                              <span className="text-slate-800 font-semibold">{h.description}</span>
                            </div>
                            <div className="text-[11px] text-slate-400">
                              Actor: <span className="font-mono text-slate-600">{h.actor}</span>
                            </div>
                          </div>
                          <span className="text-[11px] text-slate-400 whitespace-nowrap">
                            {new Date(h.created_at).toLocaleString()}
                          </span>
                        </div>
                      ))}
                    </div>
                  )}
                </div>
              )}
            </>
          )}
        </div>

        {/* Footer Actions */}
        <div className="flex items-center justify-between border-t border-slate-100 px-6 py-3 bg-slate-50/50 shrink-0">
          <div className="flex items-center gap-2">
            <button
              type="button"
              onClick={handleToggleAutoRenew}
              disabled={togglingAutoRenew || !subscription}
              className="inline-flex items-center gap-1.5 rounded-xl border border-slate-200 bg-white px-3 py-1.5 text-xs font-semibold text-slate-700 hover:bg-slate-50 transition-colors shadow-2xs"
            >
              <RefreshCw className={`h-3.5 w-3.5 text-slate-500 ${togglingAutoRenew ? 'animate-spin' : ''}`} />
              <span>{subscription?.auto_renew ? 'Disable Auto-Renew' : 'Enable Auto-Renew'}</span>
            </button>
          </div>

          <div className="flex items-center gap-2">
            <button
              type="button"
              onClick={() => onChangePlan(subscription)}
              disabled={!subscription}
              className="inline-flex items-center gap-1.5 rounded-xl border border-blue-200 bg-blue-50/80 px-3.5 py-1.5 text-xs font-bold text-blue-700 hover:bg-blue-100 transition-colors shadow-2xs"
            >
              <ArrowRightLeft className="h-3.5 w-3.5" />
              <span>Change Plan</span>
            </button>

            <button
              type="button"
              onClick={() => onRenew(subscription)}
              disabled={!subscription}
              className="inline-flex items-center gap-1.5 rounded-xl border border-emerald-200 bg-emerald-50/80 px-3.5 py-1.5 text-xs font-bold text-emerald-700 hover:bg-emerald-100 transition-colors shadow-2xs"
            >
              <RefreshCw className="h-3.5 w-3.5" />
              <span>Renew / Extend</span>
            </button>

            <button
              type="button"
              onClick={() => onCancel(subscription)}
              disabled={!subscription}
              className="inline-flex items-center gap-1.5 rounded-xl border border-rose-200 bg-rose-50/80 px-3.5 py-1.5 text-xs font-bold text-rose-700 hover:bg-rose-100 transition-colors shadow-2xs"
            >
              <Ban className="h-3.5 w-3.5" />
              <span>Cancel</span>
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}

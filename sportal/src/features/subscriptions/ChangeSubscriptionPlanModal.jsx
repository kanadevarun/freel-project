import React, { useState, useEffect } from 'react';
import { X, ArrowRightLeft, AlertCircle, ShieldCheck, CheckCircle2, DollarSign, Calendar } from 'lucide-react';
import { sportalService } from '../../services/sportalService';

export function ChangeSubscriptionPlanModal({ isOpen, onClose, subscription, plans = [], onSuccess }) {
  const [selectedPlanId, setSelectedPlanId] = useState('');
  const [billingCycle, setBillingCycle] = useState('monthly');
  const [reason, setReason] = useState('');
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState('');

  useEffect(() => {
    if (subscription) {
      setSelectedPlanId(subscription.plan_id ? String(subscription.plan_id) : '');
      setBillingCycle(subscription.billing_cycle || 'monthly');
      setReason('');
      setError('');
    }
  }, [subscription, isOpen]);

  if (!isOpen || !subscription) return null;

  const currentPlan = plans.find((p) => p.id === subscription.plan_id) || {
    id: subscription.plan_id,
    name: subscription.plan_name || 'Current Plan',
    price_monthly: subscription.amount || 0,
    price_annual: (subscription.amount || 0) * 12 * 0.8,
  };

  const selectedPlan = plans.find((p) => String(p.id) === String(selectedPlanId));

  const isUpgrade = selectedPlan && selectedPlan.price_monthly > (currentPlan.price_monthly || 0);
  const isDowngrade = selectedPlan && selectedPlan.price_monthly < (currentPlan.price_monthly || 0);

  const handleSubmit = async (e) => {
    e.preventDefault();
    if (!selectedPlanId) {
      setError('Please select a commercial plan tier.');
      return;
    }

    try {
      setSubmitting(true);
      setError('');
      const payload = {
        plan_id: parseInt(selectedPlanId, 10),
        billing_cycle: billingCycle,
        reason: reason.trim() || 'Plan tier modified via SPortal administration',
      };

      const res = await sportalService.changeOrganizationPlan(subscription.org_id, payload);
      if (onSuccess) {
        onSuccess(res?.data || res);
      }
      onClose();
    } catch (err) {
      setError(err?.response?.data?.message || err.message || 'Failed to change subscription plan.');
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/50 p-4 backdrop-blur-xs">
      <div className="w-full max-w-lg rounded-2xl bg-white shadow-2xl border border-slate-100 overflow-hidden animate-in fade-in zoom-in-95 duration-200">
        {/* Header */}
        <div className="flex items-center justify-between border-b border-slate-100 px-6 py-4 bg-slate-50/50">
          <div className="flex items-center gap-2.5">
            <div className="flex h-9 w-9 items-center justify-center rounded-xl bg-blue-50 text-blue-700">
              <ArrowRightLeft className="h-5 w-5" />
            </div>
            <div>
              <h2 className="text-sm font-bold text-slate-900">Change Commercial Plan Tier</h2>
              <p className="text-xs text-slate-500">
                {subscription.org_name || `Organization #${subscription.org_id}`}
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

        <form onSubmit={handleSubmit} className="p-6 space-y-5">
          {error && (
            <div className="flex items-start gap-2.5 rounded-xl border border-rose-200 bg-rose-50/60 p-3.5 text-xs text-rose-700">
              <AlertCircle className="h-4 w-4 shrink-0 text-rose-600 mt-0.5" />
              <div>
                <span className="font-semibold">Mutation Failed:</span> {error}
              </div>
            </div>
          )}

          {/* Current Plan vs Target Plan Comparison */}
          <div className="rounded-xl border border-slate-200 bg-slate-50/50 p-3.5 space-y-2">
            <div className="flex items-center justify-between text-xs">
              <span className="text-slate-500">Active Tier:</span>
              <span className="font-bold text-slate-800">
                {subscription.plan_name || 'Not Configured'} (${subscription.amount || 0}/mo)
              </span>
            </div>
            <div className="flex items-center justify-between text-xs">
              <span className="text-slate-500">Current Cycle:</span>
              <span className="font-medium text-slate-700 capitalize">
                {subscription.billing_cycle || 'Monthly'}
              </span>
            </div>
          </div>

          {/* Target Plan Selection */}
          <div className="space-y-1.5">
            <label className="text-xs font-bold text-slate-700 uppercase tracking-wider block">
              Target Subscription Tier *
            </label>
            <div className="grid grid-cols-1 gap-2.5">
              {plans.map((p) => {
                const isCurrent = p.id === subscription.plan_id;
                const isSelected = String(p.id) === String(selectedPlanId);
                const price = billingCycle === 'annual' ? p.price_annual : p.price_monthly;
                const periodLabel = billingCycle === 'annual' ? '/yr' : '/mo';

                return (
                  <div
                    key={p.id}
                    onClick={() => setSelectedPlanId(String(p.id))}
                    className={`cursor-pointer rounded-xl border p-3.5 transition-all flex items-center justify-between ${
                      isSelected
                        ? 'border-blue-600 bg-blue-50/40 ring-1 ring-blue-600'
                        : 'border-slate-200 hover:border-slate-300 bg-white'
                    }`}
                  >
                    <div className="space-y-1">
                      <div className="flex items-center gap-2">
                        <span className="text-sm font-bold text-slate-900">{p.name}</span>
                        {isCurrent && (
                          <span className="rounded-md bg-slate-100 px-2 py-0.5 text-[10px] font-semibold text-slate-600">
                            Current
                          </span>
                        )}
                      </div>
                      <p className="text-xs text-slate-500 line-clamp-1">{p.description}</p>
                    </div>
                    <div className="text-right">
                      <div className="text-sm font-extrabold text-navy-900">
                        ${price?.toLocaleString() || 0}
                        <span className="text-xs font-normal text-slate-500">{periodLabel}</span>
                      </div>
                    </div>
                  </div>
                );
              })}
            </div>
          </div>

          {/* Billing Cycle Option */}
          <div className="space-y-1.5">
            <label className="text-xs font-bold text-slate-700 uppercase tracking-wider block">
              Billing Interval
            </label>
            <div className="grid grid-cols-2 gap-3">
              <button
                type="button"
                onClick={() => setBillingCycle('monthly')}
                className={`rounded-xl border py-2.5 px-3 text-xs font-bold transition-all ${
                  billingCycle === 'monthly'
                    ? 'border-blue-600 bg-blue-50/50 text-blue-800 ring-1 ring-blue-600'
                    : 'border-slate-200 bg-white text-slate-700 hover:bg-slate-50'
                }`}
              >
                Monthly Billing
              </button>
              <button
                type="button"
                onClick={() => setBillingCycle('annual')}
                className={`rounded-xl border py-2.5 px-3 text-xs font-bold transition-all ${
                  billingCycle === 'annual'
                    ? 'border-blue-600 bg-blue-50/50 text-blue-800 ring-1 ring-blue-600'
                    : 'border-slate-200 bg-white text-slate-700 hover:bg-slate-50'
                }`}
              >
                Annual Billing (Discounted)
              </button>
            </div>
          </div>

          {/* Change Type Indicator */}
          {selectedPlan && selectedPlan.id !== subscription.plan_id && (
            <div className={`rounded-xl p-3 text-xs flex items-center gap-2 ${
              isUpgrade ? 'bg-emerald-50 text-emerald-800 border border-emerald-200' : 'bg-amber-50 text-amber-800 border border-amber-200'
            }`}>
              <CheckCircle2 className="h-4 w-4 shrink-0" />
              <span>
                <strong>{isUpgrade ? 'Plan Upgrade:' : 'Plan Downgrade:'}</strong> Entitlements and quotas will be instantly updated in MariaDB and reflected immediately in CPortal.
              </span>
            </div>
          )}

          {/* Audit Reason */}
          <div className="space-y-1.5">
            <label className="text-xs font-bold text-slate-700 uppercase tracking-wider block">
              Administrative Reason / Audit Note
            </label>
            <textarea
              value={reason}
              onChange={(e) => setReason(e.target.value)}
              placeholder="E.g., Customer requested commercial tier upgrade following sales negotiation..."
              rows={2}
              className="w-full rounded-xl border border-slate-200 p-2.5 text-xs text-slate-800 focus:border-blue-600 focus:outline-hidden"
            />
          </div>

          {/* Actions */}
          <div className="flex items-center justify-end gap-2.5 pt-3 border-t border-slate-100">
            <button
              type="button"
              onClick={onClose}
              disabled={submitting}
              className="rounded-xl border border-slate-200 px-4 py-2 text-xs font-semibold text-slate-700 hover:bg-slate-50 transition-colors"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={submitting}
              className="inline-flex items-center gap-1.5 rounded-xl bg-blue-600 px-4 py-2 text-xs font-bold text-white hover:bg-blue-700 disabled:opacity-50 transition-colors shadow-xs"
            >
              {submitting ? 'Applying Change...' : 'Confirm Plan Change'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}

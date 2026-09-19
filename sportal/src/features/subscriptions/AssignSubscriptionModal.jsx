import React, { useState, useEffect } from 'react';
import { X, PlusCircle, AlertCircle, ShieldCheck, CheckCircle2 } from 'lucide-react';
import { sportalService } from '../../services/sportalService';

export function AssignSubscriptionModal({ isOpen, onClose, organization, plans = [], onSuccess }) {
  const [selectedPlanId, setSelectedPlanId] = useState('');
  const [billingCycle, setBillingCycle] = useState('monthly');
  const [status, setStatus] = useState('ACTIVE');
  const [reason, setReason] = useState('');
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState('');

  useEffect(() => {
    if (plans.length > 0 && !selectedPlanId) {
      setSelectedPlanId(String(plans[0].id));
    }
    setReason('');
    setError('');
  }, [isOpen, plans]);

  if (!isOpen || !organization) return null;

  const orgId = organization.id || organization.org_id;
  const orgName = organization.name || organization.org_name || `Organization #${orgId}`;

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
        status,
        reason: reason.trim() || 'Commercial plan assigned via SPortal onboarding administration',
      };

      const res = await sportalService.assignOrganizationSubscription(orgId, payload);
      if (onSuccess) {
        onSuccess(res?.data || res);
      }
      onClose();
    } catch (err) {
      setError(err?.response?.data?.message || err.message || 'Failed to assign subscription.');
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
              <PlusCircle className="h-5 w-5" />
            </div>
            <div>
              <h2 className="text-sm font-bold text-slate-900">Assign Commercial Subscription</h2>
              <p className="text-xs text-slate-500">{orgName}</p>
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
                <span className="font-semibold">Assignment Failed:</span> {error}
              </div>
            </div>
          )}

          {/* Plan Options */}
          <div className="space-y-1.5">
            <label className="text-xs font-bold text-slate-700 uppercase tracking-wider block">
              Commercial Tier *
            </label>
            <div className="grid grid-cols-1 gap-2.5">
              {plans.map((p) => {
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
                      <span className="text-sm font-bold text-slate-900">{p.name}</span>
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

          {/* Billing Cycle & Status in 2 cols */}
          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-1.5">
              <label className="text-xs font-bold text-slate-700 uppercase tracking-wider block">
                Billing Cycle
              </label>
              <select
                value={billingCycle}
                onChange={(e) => setBillingCycle(e.target.value)}
                className="w-full rounded-xl border border-slate-200 bg-white py-2.5 px-3 text-xs font-semibold text-slate-800 focus:border-blue-600 focus:outline-hidden"
              >
                <option value="monthly">Monthly ($)</option>
                <option value="annual">Annual ($ Discounted)</option>
              </select>
            </div>

            <div className="space-y-1.5">
              <label className="text-xs font-bold text-slate-700 uppercase tracking-wider block">
                Initial Status
              </label>
              <select
                value={status}
                onChange={(e) => setStatus(e.target.value)}
                className="w-full rounded-xl border border-slate-200 bg-white py-2.5 px-3 text-xs font-semibold text-slate-800 focus:border-blue-600 focus:outline-hidden"
              >
                <option value="ACTIVE">ACTIVE (Full Access)</option>
                <option value="TRIALING">TRIALING (Evaluation)</option>
              </select>
            </div>
          </div>

          {/* Administrative Note */}
          <div className="space-y-1.5">
            <label className="text-xs font-bold text-slate-700 uppercase tracking-wider block">
              Audit Justification
            </label>
            <textarea
              value={reason}
              onChange={(e) => setReason(e.target.value)}
              placeholder="E.g., Onboarding customer approved for Starter tier contract..."
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
              {submitting ? 'Assigning...' : 'Assign Subscription'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}

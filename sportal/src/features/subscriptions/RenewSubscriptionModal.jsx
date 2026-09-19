import React, { useState, useEffect } from 'react';
import { X, RefreshCw, AlertCircle, Calendar, CheckCircle2, Clock } from 'lucide-react';
import { sportalService } from '../../services/sportalService';

export function RenewSubscriptionModal({ isOpen, onClose, subscription, onSuccess }) {
  const [extendMonths, setExtendMonths] = useState(1);
  const [reason, setReason] = useState('');
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState('');

  useEffect(() => {
    if (subscription) {
      setExtendMonths(1);
      setReason('');
      setError('');
    }
  }, [subscription, isOpen]);

  if (!isOpen || !subscription) return null;

  // Calculate projected new renewal date
  const currentEnd = subscription.current_period_end ? new Date(subscription.current_period_end) : new Date();
  const projectedEnd = new Date(currentEnd);
  projectedEnd.setMonth(projectedEnd.getMonth() + parseInt(extendMonths, 10));

  const handleSubmit = async (e) => {
    e.preventDefault();
    try {
      setSubmitting(true);
      setError('');
      const payload = {
        extend_months: parseInt(extendMonths, 10),
        reason: reason.trim() || `Manual subscription period extension (+${extendMonths} months) executed via SPortal`,
      };

      const res = await sportalService.renewOrganizationSubscription(subscription.org_id, payload);
      if (onSuccess) {
        onSuccess(res?.data || res);
      }
      onClose();
    } catch (err) {
      setError(err?.response?.data?.message || err.message || 'Failed to renew subscription.');
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/50 p-4 backdrop-blur-xs">
      <div className="w-full max-w-md rounded-2xl bg-white shadow-2xl border border-slate-100 overflow-hidden animate-in fade-in zoom-in-95 duration-200">
        {/* Header */}
        <div className="flex items-center justify-between border-b border-slate-100 px-6 py-4 bg-slate-50/50">
          <div className="flex items-center gap-2.5">
            <div className="flex h-9 w-9 items-center justify-center rounded-xl bg-emerald-50 text-emerald-700">
              <RefreshCw className="h-5 w-5" />
            </div>
            <div>
              <h2 className="text-sm font-bold text-slate-900">Renew / Extend Subscription</h2>
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
                <span className="font-semibold">Renewal Failed:</span> {error}
              </div>
            </div>
          )}

          {/* Current Status Overview */}
          <div className="rounded-xl border border-slate-200 bg-slate-50/50 p-3.5 space-y-2">
            <div className="flex items-center justify-between text-xs">
              <span className="text-slate-500">Active Plan:</span>
              <span className="font-bold text-slate-800">{subscription.plan_name}</span>
            </div>
            <div className="flex items-center justify-between text-xs">
              <span className="text-slate-500">Current Expiry:</span>
              <span className="font-medium text-slate-800">
                {subscription.current_period_end ? new Date(subscription.current_period_end).toLocaleDateString() : 'N/A'}
              </span>
            </div>
            <div className="flex items-center justify-between text-xs">
              <span className="text-slate-500">Days Remaining:</span>
              <span className={`font-bold ${subscription.days_until_renewal < 15 ? 'text-amber-600' : 'text-emerald-700'}`}>
                {subscription.days_until_renewal ?? '—'} days
              </span>
            </div>
          </div>

          {/* Extension Duration */}
          <div className="space-y-1.5">
            <label className="text-xs font-bold text-slate-700 uppercase tracking-wider block">
              Extension Term
            </label>
            <div className="grid grid-cols-4 gap-2">
              {[
                { months: 1, label: '+1 Mo' },
                { months: 3, label: '+3 Mo' },
                { months: 6, label: '+6 Mo' },
                { months: 12, label: '+1 Yr' },
              ].map((opt) => (
                <button
                  key={opt.months}
                  type="button"
                  onClick={() => setExtendMonths(opt.months)}
                  className={`rounded-xl border py-2.5 text-xs font-bold transition-all text-center ${
                    extendMonths === opt.months
                      ? 'border-emerald-600 bg-emerald-50/60 text-emerald-800 ring-1 ring-emerald-600'
                      : 'border-slate-200 bg-white text-slate-700 hover:bg-slate-50'
                  }`}
                >
                  {opt.label}
                </button>
              ))}
            </div>
          </div>

          {/* New Expiry Preview */}
          <div className="rounded-xl border border-emerald-200 bg-emerald-50/50 p-3.5 flex items-center justify-between text-xs">
            <div className="flex items-center gap-2 text-emerald-800 font-semibold">
              <Calendar className="h-4 w-4 text-emerald-600" />
              <span>Projected Expiration:</span>
            </div>
            <span className="font-bold text-emerald-900 text-sm">
              {projectedEnd.toLocaleDateString(undefined, { year: 'numeric', month: 'short', day: 'numeric' })}
            </span>
          </div>

          {/* Administrative Note */}
          <div className="space-y-1.5">
            <label className="text-xs font-bold text-slate-700 uppercase tracking-wider block">
              Administrative Justification
            </label>
            <textarea
              value={reason}
              onChange={(e) => setReason(e.target.value)}
              placeholder="Reason for extension (e.g. offline invoice payment cleared, grace period grant)..."
              rows={2}
              className="w-full rounded-xl border border-slate-200 p-2.5 text-xs text-slate-800 focus:border-emerald-600 focus:outline-hidden"
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
              className="inline-flex items-center gap-1.5 rounded-xl bg-emerald-600 px-4 py-2 text-xs font-bold text-white hover:bg-emerald-700 disabled:opacity-50 transition-colors shadow-xs"
            >
              {submitting ? 'Renewing...' : 'Extend Subscription'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}

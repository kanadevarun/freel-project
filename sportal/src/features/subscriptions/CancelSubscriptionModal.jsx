import React, { useState, useEffect } from 'react';
import { X, Ban, AlertTriangle, AlertCircle, Calendar } from 'lucide-react';
import { sportalService } from '../../services/sportalService';

export function CancelSubscriptionModal({ isOpen, onClose, subscription, onSuccess }) {
  const [immediate, setImmediate] = useState(false);
  const [reason, setReason] = useState('');
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState('');

  useEffect(() => {
    if (subscription) {
      setImmediate(false);
      setReason('');
      setError('');
    }
  }, [subscription, isOpen]);

  if (!isOpen || !subscription) return null;

  const handleSubmit = async (e) => {
    e.preventDefault();
    if (!reason.trim()) {
      setError('A cancellation reason is required for compliance and audit records.');
      return;
    }

    try {
      setSubmitting(true);
      setError('');
      const payload = {
        immediate,
        reason: reason.trim(),
      };

      const res = await sportalService.cancelOrganizationSubscription(subscription.org_id, payload);
      if (onSuccess) {
        onSuccess(res?.data || res);
      }
      onClose();
    } catch (err) {
      setError(err?.response?.data?.message || err.message || 'Failed to cancel subscription.');
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
            <div className="flex h-9 w-9 items-center justify-center rounded-xl bg-rose-50 text-rose-700">
              <Ban className="h-5 w-5" />
            </div>
            <div>
              <h2 className="text-sm font-bold text-slate-900">Cancel Customer Subscription</h2>
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
                <span className="font-semibold">Action Blocked:</span> {error}
              </div>
            </div>
          )}

          {/* Warning Banner */}
          <div className="flex items-start gap-3 rounded-xl border border-amber-200 bg-amber-50/80 p-3.5 text-xs text-amber-900">
            <AlertTriangle className="h-4 w-4 shrink-0 text-amber-700 mt-0.5" />
            <div className="space-y-1">
              <p className="font-bold">Commercial Impact Warning</p>
              <p className="text-amber-800">
                Cancelling this subscription directly modifies the tenant contract in MariaDB. Customer freight forwarder users in CPortal will see their tier changed according to the schedule selected below.
              </p>
            </div>
          </div>

          {/* Cancellation Schedule */}
          <div className="space-y-1.5">
            <label className="text-xs font-bold text-slate-700 uppercase tracking-wider block">
              Cancellation Timing
            </label>
            <div className="space-y-2">
              <label
                className={`flex items-start gap-3 rounded-xl border p-3 cursor-pointer transition-all ${
                  !immediate
                    ? 'border-blue-600 bg-blue-50/40 ring-1 ring-blue-600'
                    : 'border-slate-200 bg-white hover:bg-slate-50'
                }`}
              >
                <input
                  type="radio"
                  name="cancelType"
                  checked={!immediate}
                  onChange={() => setImmediate(false)}
                  className="mt-0.5 text-blue-600 focus:ring-blue-500"
                />
                <div className="space-y-0.5">
                  <span className="text-xs font-bold text-slate-900 block">
                    Cancel at Period End (Graceful Expiration)
                  </span>
                  <p className="text-[11px] text-slate-500">
                    Auto-renew will be turned off. Service remains active until{' '}
                    <strong>
                      {subscription.current_period_end
                        ? new Date(subscription.current_period_end).toLocaleDateString()
                        : 'end of billing cycle'}
                    </strong>
                    .
                  </p>
                </div>
              </label>

              <label
                className={`flex items-start gap-3 rounded-xl border p-3 cursor-pointer transition-all ${
                  immediate
                    ? 'border-rose-600 bg-rose-50/40 ring-1 ring-rose-600'
                    : 'border-slate-200 bg-white hover:bg-slate-50'
                }`}
              >
                <input
                  type="radio"
                  name="cancelType"
                  checked={immediate}
                  onChange={() => setImmediate(true)}
                  className="mt-0.5 text-rose-600 focus:ring-rose-500"
                />
                <div className="space-y-0.5">
                  <span className="text-xs font-bold text-rose-900 block">
                    Cancel Immediately (Revoke Commercial Access)
                  </span>
                  <p className="text-[11px] text-slate-500">
                    Immediately marks subscription as <span className="font-mono text-rose-700">CANCELED</span>. Paid features are disabled instantly.
                  </p>
                </div>
              </label>
            </div>
          </div>

          {/* Audit Reason */}
          <div className="space-y-1.5">
            <label className="text-xs font-bold text-slate-700 uppercase tracking-wider block">
              Cancellation Justification *
            </label>
            <textarea
              value={reason}
              onChange={(e) => setReason(e.target.value)}
              placeholder="Mandatory reason for cancellation (e.g., customer churn, company dissolved, non-payment)..."
              rows={3}
              required
              className="w-full rounded-xl border border-slate-200 p-2.5 text-xs text-slate-800 focus:border-rose-600 focus:outline-hidden"
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
              Back
            </button>
            <button
              type="submit"
              disabled={submitting || !reason.trim()}
              className="inline-flex items-center gap-1.5 rounded-xl bg-rose-600 px-4 py-2 text-xs font-bold text-white hover:bg-rose-700 disabled:opacity-50 transition-colors shadow-xs"
            >
              {submitting ? 'Cancelling...' : immediate ? 'Cancel Immediately' : 'Schedule Cancellation'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}

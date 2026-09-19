import React, { useState } from 'react';
import { X, AlertTriangle, CheckCircle, ShieldAlert } from 'lucide-react';
import { sportalService } from '../../services/sportalService';

export function ConfirmUserStatusModal({ isOpen, onClose, user, newStatus, onSuccess }) {
  const [reason, setReason] = useState('');
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState('');

  if (!isOpen || !user) return null;

  const isDeactivating = newStatus !== 'ACTIVE';

  const handleConfirm = async (e) => {
    e.preventDefault();
    setSubmitting(true);
    setError('');

    try {
      await sportalService.updateCustomerUserStatus(user.org_id, user.user_id, {
        status: newStatus,
        reason: reason.trim() || undefined,
      });

      if (onSuccess) {
        onSuccess();
      }
      onClose();
    } catch (err) {
      setError(err?.response?.data?.message || err.message || 'Failed to update user status.');
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/50 p-4 backdrop-blur-xs">
      <div className="w-full max-w-md rounded-2xl bg-white shadow-2xl border border-slate-100 overflow-hidden animate-in fade-in zoom-in-95 duration-200">
        <div className="flex items-center justify-between border-b border-slate-100 px-6 py-4 bg-slate-50/50">
          <div className="flex items-center gap-2.5">
            <div
              className={`flex h-9 w-9 items-center justify-center rounded-xl text-white ${
                isDeactivating ? 'bg-rose-600' : 'bg-emerald-600'
              }`}
            >
              {isDeactivating ? <AlertTriangle className="h-4 w-4" /> : <CheckCircle className="h-4 w-4" />}
            </div>
            <div>
              <h2 className="text-base font-bold text-slate-900">
                {isDeactivating ? 'Deactivate Customer User' : 'Reactivate Customer User'}
              </h2>
              <p className="text-xs text-slate-500">Access lifecycle administration</p>
            </div>
          </div>
          <button
            onClick={onClose}
            className="rounded-lg p-1.5 text-slate-400 hover:bg-slate-100 hover:text-slate-600 transition-colors"
          >
            <X className="h-4 w-4" />
          </button>
        </div>

        <form onSubmit={handleConfirm} className="p-6 space-y-4">
          {error && (
            <div className="rounded-xl border border-rose-200 bg-rose-50 p-3 text-xs text-rose-700">
              {error}
            </div>
          )}

          <div className="rounded-xl border border-slate-200 bg-slate-50 p-3.5 space-y-1.5 text-xs">
            <div className="flex justify-between">
              <span className="text-slate-500">User:</span>
              <span className="font-semibold text-slate-900">{user.full_name || user.email}</span>
            </div>
            <div className="flex justify-between">
              <span className="text-slate-500">Email:</span>
              <span className="font-mono text-slate-800">{user.email}</span>
            </div>
            <div className="flex justify-between">
              <span className="text-slate-500">Organization:</span>
              <span className="font-medium text-slate-800">{user.org_name} (#{user.org_id})</span>
            </div>
            <div className="flex justify-between">
              <span className="text-slate-500">Role:</span>
              <span className="font-medium text-navy-900">{user.role_name}</span>
            </div>
          </div>

          <div className="text-xs text-slate-600 leading-relaxed">
            {isDeactivating ? (
              <p>
                Are you sure you want to <strong>deactivate</strong> this account? Setting the account to <strong>INACTIVE</strong> immediately denies the user ability to authenticate into CPortal.
              </p>
            ) : (
              <p>
                Are you sure you want to <strong>reactivate</strong> this account? Setting the account to <strong>ACTIVE</strong> restores normal authentication and access to CPortal.
              </p>
            )}
          </div>

          <div>
            <label className="block text-xs font-semibold text-slate-700 mb-1">
              Administrative Reason (recorded in audit log)
            </label>
            <textarea
              rows={2}
              placeholder={isDeactivating ? "e.g. Employee offboarding, security lockout..." : "e.g. Identity verified, reactivation requested..."}
              value={reason}
              onChange={(e) => setReason(e.target.value)}
              className="w-full rounded-xl border border-slate-200 bg-white p-2.5 text-xs text-slate-800 placeholder:text-slate-400 focus:border-navy-900 focus:outline-hidden"
            />
          </div>

          <div className="border-t border-slate-100 pt-4 flex items-center justify-end gap-2.5">
            <button
              type="button"
              onClick={onClose}
              disabled={submitting}
              className="rounded-xl border border-slate-200 px-4 py-2 text-xs font-semibold text-slate-600 hover:bg-slate-50 transition-colors"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={submitting}
              className={`rounded-xl px-5 py-2 text-xs font-semibold text-white shadow-xs transition-colors disabled:opacity-50 ${
                isDeactivating ? 'bg-rose-600 hover:bg-rose-700' : 'bg-emerald-600 hover:bg-emerald-700'
              }`}
            >
              {submitting
                ? 'Updating...'
                : isDeactivating
                ? 'Confirm Deactivation'
                : 'Confirm Reactivation'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}

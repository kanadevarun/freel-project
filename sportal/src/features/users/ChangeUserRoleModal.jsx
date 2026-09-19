import React, { useState } from 'react';
import { X, Shield, AlertTriangle, CheckCircle, ArrowRight } from 'lucide-react';
import { sportalService } from '../../services/sportalService';

const AVAILABLE_ROLES = [
  { id: 'SUPER_ADMIN', name: 'SUPER_ADMIN', label: 'Super Admin', desc: 'Full administrative control over tenant resources, users, and settings' },
  { id: 'OPERATIONS', name: 'OPERATIONS', label: 'Operations Specialist', desc: 'Operational shipment tracking, exceptions, and container workflows' },
  { id: 'SALES', name: 'SALES', label: 'Sales Representative', desc: 'Commercial quotations, RFQs, CRM leads, and customer bookings' },
  { id: 'PRICING', name: 'PRICING', label: 'Pricing Analyst', desc: 'Freight tariffs, rate cards, and spot quote pricing calculations' },
  { id: 'FINANCE', name: 'FINANCE', label: 'Finance Officer', desc: 'Invoices, credit lines, payment reconciliation, and billing records' },
  { id: 'DOCUMENTATION', name: 'DOCUMENTATION', label: 'Documentation Specialist', desc: 'Bills of Lading, customs declarations, certificates, and compliance manifests' },
  { id: 'HR', name: 'HR', label: 'HR / Personnel', desc: 'Tenant staff directories and personnel records' }
];

export function ChangeUserRoleModal({ isOpen, onClose, user, orgId, onRoleChanged }) {
  const [selectedRole, setSelectedRole] = useState(user?.role_name || 'OPERATIONS');
  const [reason, setReason] = useState('');
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState('');

  if (!isOpen || !user) return null;

  const currentRole = user.role_name;
  const isSuperAdminCurrent = currentRole === 'SUPER_ADMIN';
  const isSuperAdminTarget = selectedRole === 'SUPER_ADMIN';

  const handleSubmit = async (e) => {
    e.preventDefault();
    if (!reason.trim()) {
      setError('Please provide an administrative reason for the role change to satisfy governance compliance.');
      return;
    }
    if (selectedRole === currentRole) {
      setError('Selected role is identical to current role. Please choose a different role.');
      return;
    }

    try {
      setSubmitting(true);
      setError('');
      const targetOrgId = orgId || user.org_id;
      const targetUserId = user.user_id || user.id;

      await sportalService.updateCustomerUserRole(targetOrgId, targetUserId, {
        role_id: selectedRole,
        reason: reason.trim()
      });

      if (onRoleChanged) {
        onRoleChanged({ ...user, role_name: selectedRole });
      }
      onClose();
    } catch (err) {
      setError(err?.response?.data?.message || err.message || 'Failed to update customer user role.');
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="fixed inset-0 z-60 flex items-center justify-center bg-slate-900/50 p-4 backdrop-blur-xs">
      <div className="w-full max-w-lg rounded-2xl bg-white shadow-2xl border border-slate-100 overflow-hidden animate-in fade-in zoom-in-95 duration-200">
        {/* Header */}
        <div className="flex items-center justify-between border-b border-slate-100 px-6 py-4 bg-slate-50/50">
          <div className="flex items-center gap-2.5">
            <div className="flex h-9 w-9 items-center justify-center rounded-xl bg-navy-900 text-white shadow-xs">
              <Shield className="h-4 w-4" />
            </div>
            <div>
              <h2 className="text-sm font-bold text-slate-900">Reassign Customer User Role</h2>
              <p className="text-xs text-slate-500">
                {user.full_name || user.email} • #{user.user_id || user.id}
              </p>
            </div>
          </div>
          <button
            onClick={onClose}
            className="rounded-lg p-1 text-slate-400 hover:bg-slate-100 hover:text-slate-600 transition-colors"
          >
            <X className="h-4 w-4" />
          </button>
        </div>

        <form onSubmit={handleSubmit} className="p-6 space-y-5">
          {error && (
            <div className="flex items-start gap-2.5 rounded-xl border border-rose-200 bg-rose-50 p-3.5 text-xs text-rose-700">
              <AlertTriangle className="h-4 w-4 shrink-0 text-rose-500 mt-0.5" />
              <span>{error}</span>
            </div>
          )}

          {/* Current vs Target Preview */}
          <div className="flex items-center justify-between p-3.5 rounded-xl bg-slate-50 border border-slate-200/80 text-xs">
            <div>
              <span className="text-[11px] text-slate-400 uppercase font-semibold block">Current Role</span>
              <span className="font-bold text-navy-900">{currentRole}</span>
            </div>
            <ArrowRight className="h-4 w-4 text-slate-400" />
            <div>
              <span className="text-[11px] text-slate-400 uppercase font-semibold block">Target Role</span>
              <span className="font-bold text-blue-700">{selectedRole}</span>
            </div>
          </div>

          {/* Super Admin Notice */}
          {isSuperAdminCurrent && selectedRole !== 'SUPER_ADMIN' && (
            <div className="rounded-xl border border-amber-200 bg-amber-50/70 p-3 text-xs text-amber-800 space-y-1">
              <p className="font-semibold flex items-center gap-1.5">
                <AlertTriangle className="h-3.5 w-3.5 text-amber-600" />
                Sole Super Admin Protection Guard
              </p>
              <p className="text-[11px] text-amber-700 leading-relaxed">
                If this user is the sole active Super Admin in {user.org_name || 'the organization'}, demotion will be rejected by the backend to prevent customer tenant lockout.
              </p>
            </div>
          )}

          {/* Target Role Selector */}
          <div className="space-y-2">
            <label className="text-xs font-semibold text-slate-700">
              Select New Assigned Role <span className="text-rose-500">*</span>
            </label>
            <div className="space-y-1.5 max-h-48 overflow-y-auto pr-1">
              {AVAILABLE_ROLES.map((r) => {
                const isSelected = selectedRole === r.id;
                return (
                  <label
                    key={r.id}
                    className={`flex items-start gap-3 p-2.5 rounded-xl border cursor-pointer transition-colors ${
                      isSelected
                        ? 'border-navy-900 bg-navy-900/5 ring-1 ring-navy-900'
                        : 'border-slate-200 hover:border-slate-300 hover:bg-slate-50'
                    }`}
                  >
                    <input
                      type="radio"
                      name="role"
                      value={r.id}
                      checked={isSelected}
                      onChange={(e) => setSelectedRole(e.target.value)}
                      className="mt-0.5 text-navy-900 focus:ring-navy-900"
                    />
                    <div className="text-xs">
                      <div className="flex items-center gap-2">
                        <span className="font-bold text-slate-900">{r.id}</span>
                        <span className="text-[11px] text-slate-500 font-medium">({r.label})</span>
                      </div>
                      <p className="text-[11px] text-slate-500 mt-0.5 leading-snug">{r.desc}</p>
                    </div>
                  </label>
                );
              })}
            </div>
          </div>

          {/* Reason Input */}
          <div className="space-y-1.5">
            <label className="text-xs font-semibold text-slate-700">
              Administrative Reason <span className="text-rose-500">*</span>
            </label>
            <textarea
              value={reason}
              onChange={(e) => setReason(e.target.value)}
              placeholder="E.g., Customer requested role adjustment per support ticket #HQ-492..."
              rows={2}
              required
              className="w-full rounded-xl border border-slate-200 px-3 py-2 text-xs text-slate-900 placeholder:text-slate-400 focus:border-navy-900 focus:outline-none focus:ring-1 focus:ring-navy-900"
            />
            <p className="text-[11px] text-slate-400">
              Reason is permanently logged in the immutable audit ledger with your SPortal admin signature.
            </p>
          </div>

          {/* Footer Controls */}
          <div className="flex items-center justify-end gap-2 pt-2 border-t border-slate-100">
            <button
              type="button"
              onClick={onClose}
              disabled={submitting}
              className="rounded-xl border border-slate-200 px-4 py-2 text-xs font-semibold text-slate-600 hover:bg-slate-50 transition-colors disabled:opacity-50"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={submitting || selectedRole === currentRole}
              className="rounded-xl bg-navy-900 px-4 py-2 text-xs font-semibold text-white hover:bg-navy-800 transition-colors shadow-xs disabled:opacity-50 flex items-center gap-1.5"
            >
              {submitting ? 'Applying Change...' : 'Confirm Role Reassignment'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}

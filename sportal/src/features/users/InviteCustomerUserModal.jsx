import React, { useState, useEffect } from 'react';
import { X, Mail, Building, Shield, User, AlertCircle, CheckCircle2 } from 'lucide-react';
import { sportalService } from '../../services/sportalService';

export function InviteCustomerUserModal({ isOpen, onClose, onSuccess, initialOrgId = null }) {
  const [orgs, setOrgs] = useState([]);
  const [loadingOrgs, setLoadingOrgs] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState('');
  const [successInfo, setSuccessInfo] = useState(null);

  const [formData, setFormData] = useState({
    orgId: initialOrgId || '',
    email: '',
    firstName: '',
    lastName: '',
    roleName: 'SUPER_ADMIN',
  });

  useEffect(() => {
    if (isOpen) {
      setError('');
      setSuccessInfo(null);
      fetchOrgs();
      if (initialOrgId) {
        setFormData((prev) => ({ ...prev, orgId: String(initialOrgId) }));
      }
    }
  }, [isOpen, initialOrgId]);

  const fetchOrgs = async () => {
    try {
      setLoadingOrgs(true);
      const res = await sportalService.getOrganizations({ pageSize: 100 });
      const orgItems = res?.items || res?.data?.items || [];
      const customerOrgs = orgItems.filter((o) => o.id !== 1);
      setOrgs(customerOrgs);
      if (!initialOrgId && customerOrgs.length > 0) {
        setFormData((prev) => ({ ...prev, orgId: prev.orgId || String(customerOrgs[0].id) }));
      }
    } catch (err) {
      console.error('Failed to fetch organizations for invite modal:', err);
    } finally {
      setLoadingOrgs(false);
    }
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    if (!formData.orgId) {
      setError('Please select a customer organization.');
      return;
    }
    if (!formData.email || !formData.email.includes('@')) {
      setError('Please enter a valid email address.');
      return;
    }

    setSubmitting(true);
    setError('');

    try {
      const res = await sportalService.inviteCustomerUser(formData.orgId, {
        org_id: Number(formData.orgId),
        email: formData.email.trim(),
        first_name: formData.firstName.trim(),
        last_name: formData.lastName.trim(),
        role_name: formData.roleName,
      });

      const data = res?.data || res;
      if (data) {
        setSuccessInfo(data);
        if (onSuccess) {
          onSuccess(data);
        }
      }
    } catch (err) {
      setError(err?.response?.data?.message || err.message || 'Failed to send customer invitation.');
    } finally {
      setSubmitting(false);
    }
  };

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/50 p-4 backdrop-blur-xs">
      <div className="w-full max-w-lg rounded-2xl bg-white shadow-2xl border border-slate-100 overflow-hidden animate-in fade-in zoom-in-95 duration-200">
        {/* Header */}
        <div className="flex items-center justify-between border-b border-slate-100 px-6 py-4 bg-slate-50/50">
          <div className="flex items-center gap-2.5">
            <div className="flex h-9 w-9 items-center justify-center rounded-lg bg-navy-900 text-white shadow-xs">
              <Mail className="h-4 w-4" />
            </div>
            <div>
              <h2 className="text-base font-bold text-slate-900">Invite Customer User</h2>
              <p className="text-xs text-slate-500">Provision customer Super Admin or freight forwarder team member</p>
            </div>
          </div>
          <button
            onClick={onClose}
            className="rounded-lg p-1.5 text-slate-400 hover:bg-slate-100 hover:text-slate-600 transition-colors"
          >
            <X className="h-4 w-4" />
          </button>
        </div>

        {/* Content */}
        {successInfo ? (
          <div className="p-6 space-y-4">
            <div className="rounded-xl border border-emerald-200 bg-emerald-50/60 p-4 text-emerald-900">
              <div className="flex items-center gap-2.5">
                <CheckCircle2 className="h-5 w-5 text-emerald-600 shrink-0" />
                <p className="text-sm font-semibold">Invitation Generated Successfully</p>
              </div>
              <p className="mt-1 text-xs text-emerald-700">
                Authoritative invitation record created for <strong className="font-semibold">{successInfo.email}</strong> under organization <strong className="font-semibold">{successInfo.org_name}</strong> with role <strong className="font-semibold">{successInfo.role_name}</strong>.
              </p>
            </div>

            <div className="rounded-lg bg-slate-50 p-3 text-xs space-y-1.5 border border-slate-200">
              <div className="flex justify-between">
                <span className="text-slate-500">Invitation ID:</span>
                <span className="font-mono font-medium text-slate-800">#{successInfo.id}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-slate-500">Expires At:</span>
                <span className="font-medium text-slate-800">{new Date(successInfo.expires_at).toLocaleString()}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-slate-500">Initial Status:</span>
                <span className="inline-flex items-center px-2 py-0.5 rounded text-[11px] font-semibold bg-blue-100 text-blue-800">
                  {successInfo.status || 'PENDING'}
                </span>
              </div>
            </div>

            <div className="pt-2 flex justify-end">
              <button
                type="button"
                onClick={onClose}
                className="rounded-lg bg-navy-900 px-5 py-2 text-xs font-semibold text-white hover:bg-navy-800 shadow-xs transition-colors"
              >
                Done
              </button>
            </div>
          </div>
        ) : (
          <form onSubmit={handleSubmit} className="p-6 space-y-4">
            {error && (
              <div className="flex items-center gap-2 rounded-xl border border-rose-200 bg-rose-50 p-3 text-xs text-rose-700">
                <AlertCircle className="h-4 w-4 shrink-0 text-rose-500" />
                <span>{error}</span>
              </div>
            )}

            {/* Organization Select */}
            <div>
              <label className="block text-xs font-semibold text-slate-700 mb-1">
                Customer Organization <span className="text-rose-500">*</span>
              </label>
              <div className="relative">
                <Building className="absolute left-3 top-2.5 h-4 w-4 text-slate-400" />
                <select
                  disabled={loadingOrgs || !!initialOrgId}
                  value={formData.orgId}
                  onChange={(e) => setFormData({ ...formData, orgId: e.target.value })}
                  className="w-full rounded-xl border border-slate-200 bg-white py-2 pl-9 pr-3 text-xs text-slate-800 focus:border-navy-900 focus:outline-hidden disabled:bg-slate-50"
                  required
                >
                  <option value="">Select target customer...</option>
                  {orgs.map((org) => (
                    <option key={org.id} value={org.id}>
                      {org.name} (#{org.id})
                    </option>
                  ))}
                </select>
              </div>
            </div>

            {/* Email Address */}
            <div>
              <label className="block text-xs font-semibold text-slate-700 mb-1">
                Work Email Address <span className="text-rose-500">*</span>
              </label>
              <div className="relative">
                <Mail className="absolute left-3 top-2.5 h-4 w-4 text-slate-400" />
                <input
                  type="email"
                  required
                  placeholder="admin@customer-domain.com"
                  value={formData.email}
                  onChange={(e) => setFormData({ ...formData, email: e.target.value })}
                  className="w-full rounded-xl border border-slate-200 bg-white py-2 pl-9 pr-3 text-xs text-slate-800 placeholder:text-slate-400 focus:border-navy-900 focus:outline-hidden"
                />
              </div>
            </div>

            {/* Name Fields */}
            <div className="grid grid-cols-2 gap-3">
              <div>
                <label className="block text-xs font-semibold text-slate-700 mb-1">First Name</label>
                <div className="relative">
                  <User className="absolute left-3 top-2.5 h-4 w-4 text-slate-400" />
                  <input
                    type="text"
                    placeholder="Jane"
                    value={formData.firstName}
                    onChange={(e) => setFormData({ ...formData, firstName: e.target.value })}
                    className="w-full rounded-xl border border-slate-200 bg-white py-2 pl-9 pr-3 text-xs text-slate-800 placeholder:text-slate-400 focus:border-navy-900 focus:outline-hidden"
                  />
                </div>
              </div>
              <div>
                <label className="block text-xs font-semibold text-slate-700 mb-1">Last Name</label>
                <input
                  type="text"
                  placeholder="Doe"
                  value={formData.lastName}
                  onChange={(e) => setFormData({ ...formData, lastName: e.target.value })}
                  className="w-full rounded-xl border border-slate-200 bg-white py-2 px-3 text-xs text-slate-800 placeholder:text-slate-400 focus:border-navy-900 focus:outline-hidden"
                />
              </div>
            </div>

            {/* Role Selection */}
            <div>
              <label className="block text-xs font-semibold text-slate-700 mb-1">
                Initial Tenant Role <span className="text-rose-500">*</span>
              </label>
              <div className="relative">
                <Shield className="absolute left-3 top-2.5 h-4 w-4 text-slate-400" />
                <select
                  value={formData.roleName}
                  onChange={(e) => setFormData({ ...formData, roleName: e.target.value })}
                  className="w-full rounded-xl border border-slate-200 bg-white py-2 pl-9 pr-3 text-xs text-slate-800 focus:border-navy-900 focus:outline-hidden"
                >
                  <option value="SUPER_ADMIN">Super Admin (Primary Freight Forwarder Admin)</option>
                  <option value="SALES">Sales Representative</option>
                  <option value="OPERATIONS">Operations & Logistics Manager</option>
                  <option value="FINANCE">Finance & Billing Specialist</option>
                  <option value="PRICING">Pricing & Procurement Specialist</option>
                  <option value="DOCUMENTATION">Documentation & Compliance Officer</option>
                  <option value="HR">HR & Team Administrator</option>
                </select>
              </div>
              <p className="mt-1 text-[11px] text-slate-500">
                The Customer Super Admin will have full authority within CPortal to invite their operational team and configure custom roles.
              </p>
            </div>

            {/* Actions */}
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
                className="rounded-xl bg-navy-900 px-5 py-2 text-xs font-semibold text-white hover:bg-navy-800 shadow-xs transition-colors disabled:opacity-50"
              >
                {submitting ? 'Generating Invitation...' : 'Send Invitation'}
              </button>
            </div>
          </form>
        )}
      </div>
    </div>
  );
}

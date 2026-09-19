import React, { useState, useEffect } from 'react';
import { 
  X, User, Building, Shield, Clock, KeyRound, Activity, AlertCircle, 
  CheckCircle, Ban, RefreshCw, Layers, Check, Lock, Edit3, ArrowRightLeft, ShieldAlert
} from 'lucide-react';
import { sportalService } from '../../services/sportalService';
import { StatusBadge } from '../../components/common/StatusBadge';
import { ChangeUserRoleModal } from './ChangeUserRoleModal';

export function CustomerUserDetailModal({ isOpen, onClose, orgId, userId, onStatusChange }) {
  const [detail, setDetail] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [activeTab, setActiveTab] = useState('overview');
  const [isRoleModalOpen, setIsRoleModalOpen] = useState(false);

  useEffect(() => {
    if (isOpen && orgId && userId) {
      fetchDetail();
    }
  }, [isOpen, orgId, userId]);

  const fetchDetail = async () => {
    try {
      setLoading(true);
      setError('');
      const res = await sportalService.getCustomerUserDetail(orgId, userId);
      const data = res?.data || res;
      if (data) {
        setDetail(data);
      }
    } catch (err) {
      setError(err?.response?.data?.message || err.message || 'Failed to load user details.');
    } finally {
      setLoading(false);
    }
  };

  const handleRoleChanged = (updatedUser) => {
    fetchDetail();
    if (onStatusChange) {
      onStatusChange(updatedUser);
    }
  };

  if (!isOpen) return null;

  return (
    <>
      <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/50 p-4 backdrop-blur-xs">
        <div className="w-full max-w-4xl max-h-[90vh] flex flex-col rounded-2xl bg-white shadow-2xl border border-slate-100 overflow-hidden animate-in fade-in zoom-in-95 duration-200">
          {/* Header */}
          <div className="flex items-center justify-between border-b border-slate-100 px-6 py-4 bg-slate-50/50 shrink-0">
            <div className="flex items-center gap-3">
              <div className="flex h-10 w-10 items-center justify-center rounded-xl bg-navy-900 text-white font-bold text-sm shadow-xs">
                {detail?.full_name ? detail.full_name.charAt(0).toUpperCase() : detail?.email?.charAt(0).toUpperCase() || 'U'}
              </div>
              <div>
                <div className="flex items-center gap-2">
                  <h2 className="text-base font-bold text-slate-900">
                    {detail?.full_name || detail?.email || 'Customer User Dossier'}
                  </h2>
                  {detail && (
                    <span className="font-mono text-xs px-2 py-0.5 rounded bg-slate-100 text-slate-600 font-medium">
                      #{detail.user_id}
                    </span>
                  )}
                </div>
                <p className="text-xs text-slate-500">
                  {detail?.org_name ? `Organization: ${detail.org_name} (#${detail.org_id})` : 'Cross-tenant customer account'}
                </p>
              </div>
            </div>
            <button
              onClick={onClose}
              className="rounded-lg p-1.5 text-slate-400 hover:bg-slate-100 hover:text-slate-600 transition-colors"
            >
              <X className="h-4 w-4" />
            </button>
          </div>

          {/* Tab Navigation */}
          <div className="border-b border-slate-200 px-6 bg-white shrink-0">
            <div className="flex space-x-6">
              {[
                { id: 'overview', label: 'Identity & Access' },
                { id: 'effective', label: 'Effective Module Access' },
                { id: 'permissions', label: `Role Permissions (${detail?.permissions?.length || 0})` },
                { id: 'activity', label: `Audit Log (${detail?.audit_activity?.length || 0})` },
              ].map((tab) => (
                <button
                  key={tab.id}
                  onClick={() => setActiveTab(tab.id)}
                  className={`py-3 text-xs font-semibold border-b-2 transition-colors ${
                    activeTab === tab.id
                      ? 'border-navy-900 text-navy-900'
                      : 'border-transparent text-slate-500 hover:text-slate-800'
                  }`}
                >
                  {tab.label}
                </button>
              ))}
            </div>
          </div>

          {/* Body Content */}
          <div className="p-6 overflow-y-auto grow space-y-6">
            {loading ? (
              <div className="py-16 text-center space-y-3">
                <div className="mx-auto h-7 w-7 animate-spin rounded-full border-2 border-navy-900 border-t-transparent"></div>
                <p className="text-xs text-slate-500">Loading customer user dossier...</p>
              </div>
            ) : error ? (
              <div className="flex items-center gap-2 rounded-xl border border-rose-200 bg-rose-50 p-4 text-xs text-rose-700">
                <AlertCircle className="h-4 w-4 shrink-0 text-rose-500" />
                <span>{error}</span>
              </div>
            ) : detail ? (
              <>
                {/* TAB 1: OVERVIEW */}
                {activeTab === 'overview' && (
                  <div className="space-y-6">
                    {/* Status Badges Header */}
                    <div className="flex flex-wrap items-center justify-between gap-3 p-3.5 rounded-xl bg-slate-50 border border-slate-100">
                      <div className="flex flex-wrap items-center gap-3">
                        <div className="flex items-center gap-2">
                          <span className="text-xs text-slate-500">Account Status:</span>
                          <StatusBadge status={detail.status} />
                        </div>
                        <div className="h-4 w-px bg-slate-200 hidden sm:block"></div>
                        <div className="flex items-center gap-2">
                          <span className="text-xs text-slate-500">Invitation:</span>
                          <span className="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-semibold bg-emerald-50 text-emerald-700 border border-emerald-200">
                            {detail.invitation_status}
                          </span>
                        </div>
                        <div className="h-4 w-px bg-slate-200 hidden sm:block"></div>
                        <div className="flex items-center gap-2">
                          <span className="text-xs text-slate-500">Role:</span>
                          <span className="inline-flex items-center px-2.5 py-0.5 rounded-md text-xs font-bold bg-navy-900 text-white">
                            {detail.role_name}
                          </span>
                        </div>
                      </div>

                      <button
                        type="button"
                        onClick={() => setIsRoleModalOpen(true)}
                        className="inline-flex items-center gap-1.5 px-3 py-1 rounded-lg border border-slate-200 bg-white text-xs font-semibold text-slate-700 hover:bg-slate-50 transition-colors shadow-2xs"
                      >
                        <Edit3 className="h-3.5 w-3.5 text-slate-500" />
                        Reassign Role
                      </button>
                    </div>

                    {/* Two column grid */}
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-6 text-xs">
                      {/* Identity Details */}
                      <div className="space-y-3.5 rounded-xl border border-slate-200 bg-white p-4">
                        <h4 className="font-bold text-slate-900 flex items-center gap-2 uppercase tracking-wider text-[11px] text-slate-400">
                          <User className="h-3.5 w-3.5 text-slate-500" />
                          Identity & Contact Information
                        </h4>
                        <div>
                          <span className="text-slate-500 block">Full Name</span>
                          <span className="font-semibold text-slate-900">{detail.full_name || '—'}</span>
                        </div>
                        <div>
                          <span className="text-slate-500 block">Email Address</span>
                          <span className="font-mono text-slate-900">{detail.email}</span>
                        </div>
                        <div>
                          <span className="text-slate-500 block">Phone</span>
                          <span className="text-slate-800">{detail.phone || '—'}</span>
                        </div>
                        <div>
                          <span className="text-slate-500 block">Registration Timestamp</span>
                          <span className="text-slate-700">{new Date(detail.created_at).toLocaleString()}</span>
                        </div>
                      </div>

                      {/* Governance & Access */}
                      <div className="space-y-3.5 rounded-xl border border-slate-200 bg-white p-4">
                        <h4 className="font-bold text-slate-900 flex items-center gap-2 uppercase tracking-wider text-[11px] text-slate-400">
                          <Shield className="h-3.5 w-3.5 text-slate-500" />
                          Access & Authentication Boundary
                        </h4>
                        <div>
                          <span className="text-slate-500 block">Assigned Tenant Role</span>
                          <span className="font-bold text-navy-900">{detail.role_name}</span>
                          {detail.role_description && (
                            <p className="text-[11px] text-slate-400 mt-0.5">{detail.role_description}</p>
                          )}
                        </div>
                        <div>
                          <span className="text-slate-500 block">Last Successful Login</span>
                          <span className="text-slate-800 font-medium">
                            {detail.last_login ? new Date(detail.last_login).toLocaleString() : 'Never logged in'}
                          </span>
                        </div>
                        <div>
                          <span className="text-slate-500 block">Multi-Factor Authentication (MFA)</span>
                          <span className="inline-flex items-center px-2 py-0.5 rounded text-[11px] font-medium bg-slate-100 text-slate-700">
                            {detail.mfa_status || 'NOT_ENABLED'}
                          </span>
                        </div>
                        <div>
                          <span className="text-slate-500 block">Tenant Isolation Boundary</span>
                          <span className="text-slate-700">
                            Scoped exclusively to {detail.org_name} (Org #{detail.org_id})
                          </span>
                        </div>
                      </div>
                    </div>

                    {/* Security Note */}
                    <div className="rounded-xl border border-blue-200 bg-blue-50/60 p-3.5 text-xs text-blue-900 space-y-1">
                      <p className="font-semibold flex items-center gap-1.5">
                        <KeyRound className="h-3.5 w-3.5 text-blue-700" />
                        Platform Security Assurance
                      </p>
                      <p className="text-[11px] text-blue-800 leading-relaxed">
                        Passwords, hashes, session keys, and MFA secrets are cryptographically protected and never exposed through SPortal APIs. Account lifecycle modifications are permanently recorded in the immutable audit log.
                      </p>
                    </div>
                  </div>
                )}

                {/* TAB 2: EFFECTIVE MODULE ACCESS (TASK S7) */}
                {activeTab === 'effective' && (
                  <div className="space-y-6">
                    {/* Role Access Scope Banner */}
                    <div className="rounded-xl border border-slate-200 bg-slate-50/60 p-4 space-y-2">
                      <div className="flex items-center justify-between">
                        <div className="flex items-center gap-2">
                          <span className="px-2.5 py-1 rounded font-mono text-xs font-bold bg-navy-900 text-white">
                            {detail.effective_access?.role_name || detail.role_name}
                          </span>
                          <span className="text-xs font-bold text-slate-800">Effective Operational Scope</span>
                        </div>
                        <span className="text-[11px] text-slate-500">
                          Tenant: <strong>{detail.org_name}</strong>
                        </span>
                      </div>
                      <p className="text-xs text-slate-600 leading-relaxed">
                        {detail.effective_access?.explanation || 'Canonical operational scope determined by role permissions.'}
                      </p>
                      <div className="pt-2 border-t border-slate-200/60 flex items-center gap-2 text-[11px] text-slate-500">
                        <ArrowRightLeft className="h-3.5 w-3.5 text-blue-600 shrink-0" />
                        <span>{detail.effective_access?.sportal_vs_cportal_boundary}</span>
                      </div>
                    </div>

                    {/* Capabilities & Denials */}
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                      {/* Key Capabilities */}
                      <div className="rounded-xl border border-emerald-200 bg-emerald-50/30 p-4 space-y-2.5">
                        <h4 className="text-xs font-bold text-emerald-900 flex items-center gap-1.5 uppercase tracking-wider">
                          <CheckCircle className="h-3.5 w-3.5 text-emerald-600" />
                          Key Authorized Capabilities
                        </h4>
                        <ul className="space-y-1.5 text-xs text-slate-700">
                          {(detail.effective_access?.key_capabilities || []).map((cap, idx) => (
                            <li key={idx} className="flex items-start gap-2">
                              <span className="text-emerald-600 font-bold shrink-0">•</span>
                              <span className="leading-snug">{cap}</span>
                            </li>
                          ))}
                        </ul>
                      </div>

                      {/* Explicit Denials */}
                      <div className="rounded-xl border border-rose-200 bg-rose-50/30 p-4 space-y-2.5">
                        <h4 className="text-xs font-bold text-rose-900 flex items-center gap-1.5 uppercase tracking-wider">
                          <ShieldAlert className="h-3.5 w-3.5 text-rose-600" />
                          Explicit Restrictions & Denials
                        </h4>
                        <ul className="space-y-1.5 text-xs text-slate-700">
                          {(detail.effective_access?.explicit_denials || []).map((denial, idx) => (
                            <li key={idx} className="flex items-start gap-2">
                              <span className="text-rose-500 font-bold shrink-0">✕</span>
                              <span className="leading-snug">{denial}</span>
                            </li>
                          ))}
                        </ul>
                      </div>
                    </div>

                    {/* Module Cards Grid */}
                    <div className="space-y-3">
                      <h4 className="text-xs font-bold text-slate-900 uppercase tracking-wider text-[11px] text-slate-400">
                        Granular Platform Resource Breakdown
                      </h4>
                      <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
                        {(Array.isArray(detail.effective_access?.modules)
                          ? detail.effective_access.modules
                          : Object.values(detail.effective_access?.modules || {})
                        ).map((mod, idx) => {
                          const canCreate = mod?.allowed_actions ? mod.allowed_actions.includes('CREATE') : Boolean(mod?.can_create);
                          const canRead = mod?.allowed_actions ? mod.allowed_actions.includes('READ') : Boolean(mod?.can_read);
                          const canUpdate = mod?.allowed_actions ? mod.allowed_actions.includes('UPDATE') : Boolean(mod?.can_update);
                          const canDelete = mod?.allowed_actions ? mod.allowed_actions.includes('DELETE') : Boolean(mod?.can_delete);

                          let badgeBg = 'bg-slate-100 text-slate-700 border-slate-200';
                          if (mod.access_level === 'FULL') badgeBg = 'bg-emerald-50 text-emerald-700 border-emerald-200';
                          else if (mod.access_level === 'MANAGE') badgeBg = 'bg-blue-50 text-blue-700 border-blue-200';
                          else if (mod.access_level === 'VIEW_ONLY') badgeBg = 'bg-amber-50 text-amber-700 border-amber-200';

                          return (
                            <div key={mod.resource || idx} className="rounded-xl border border-slate-200 bg-white p-3.5 space-y-2 hover:border-slate-300 transition-colors">
                              <div className="flex items-center justify-between">
                                <div className="flex items-center gap-1.5 min-w-0">
                                  <span className="font-semibold text-slate-900 text-xs truncate">
                                    {mod.module || mod.name || mod.resource}
                                  </span>
                                  <span className="font-mono text-[9px] text-slate-400 bg-slate-100 px-1 py-0.5 rounded shrink-0">
                                    {mod.resource}
                                  </span>
                                </div>
                                <span className={`px-2 py-0.5 rounded-full text-[10px] font-bold border ${badgeBg}`}>
                                  {mod.access_level}
                                </span>
                              </div>
                              <p className="text-[11px] text-slate-500 line-clamp-2 leading-relaxed">
                                {mod.explanation || mod.description}
                              </p>
                              
                              {/* CRUD Badges */}
                              <div className="pt-2 border-t border-slate-100 flex items-center gap-1.5 text-[10px] font-mono">
                                <span className={`px-1.5 py-0.5 rounded ${canCreate ? 'bg-emerald-50 text-emerald-700 font-bold border border-emerald-200' : 'bg-slate-100 text-slate-400'}`}>
                                  C
                                </span>
                                <span className={`px-1.5 py-0.5 rounded ${canRead ? 'bg-emerald-50 text-emerald-700 font-bold border border-emerald-200' : 'bg-slate-100 text-slate-400'}`}>
                                  R
                                </span>
                                <span className={`px-1.5 py-0.5 rounded ${canUpdate ? 'bg-emerald-50 text-emerald-700 font-bold border border-emerald-200' : 'bg-slate-100 text-slate-400'}`}>
                                  U
                                </span>
                                <span className={`px-1.5 py-0.5 rounded ${canDelete ? 'bg-emerald-50 text-emerald-700 font-bold border border-emerald-200' : 'bg-slate-100 text-slate-400'}`}>
                                  D
                                </span>
                              </div>
                            </div>
                          );
                        })}
                      </div>
                    </div>
                  </div>
                )}

                {/* TAB 3: PERMISSIONS */}
                {activeTab === 'permissions' && (
                  <div className="space-y-4">
                    <div>
                      <h3 className="text-xs font-bold text-slate-900">
                        Authorized Capabilities ({detail.permissions?.length || 0})
                      </h3>
                      <p className="text-[11px] text-slate-500">
                        Permissions inherited by {detail.full_name || detail.email} via {detail.role_name} role
                      </p>
                    </div>

                    {detail.permissions?.length === 0 ? (
                      <div className="rounded-xl border border-slate-200 bg-slate-50 p-8 text-center text-xs text-slate-500">
                        No explicit permissions configured for role {detail.role_name}.
                      </div>
                    ) : (
                      <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 gap-2">
                        {detail.permissions.map((p, idx) => (
                          <div
                            key={idx}
                            className="flex items-center gap-2 rounded-lg border border-slate-200 bg-white px-3 py-2 text-xs font-mono text-slate-800 hover:border-slate-300 transition-colors"
                          >
                            <CheckCircle className="h-3.5 w-3.5 text-emerald-600 shrink-0" />
                            <span className="truncate">{p}</span>
                          </div>
                        ))}
                      </div>
                    )}
                  </div>
                )}

                {/* TAB 4: AUDIT ACTIVITY */}
                {activeTab === 'activity' && (
                  <div className="space-y-4">
                    <div>
                      <h3 className="text-xs font-bold text-slate-900">Audit & Governance Ledger</h3>
                      <p className="text-[11px] text-slate-500">
                        Recent chronological records captured for user #{detail.user_id}
                      </p>
                    </div>

                    {detail.audit_activity?.length === 0 ? (
                      <div className="rounded-xl border border-slate-200 bg-slate-50 p-8 text-center text-xs text-slate-500">
                        No audit activities logged for this user yet.
                      </div>
                    ) : (
                      <div className="divide-y divide-slate-100 rounded-xl border border-slate-200 bg-white overflow-hidden text-xs">
                        {detail.audit_activity.map((act) => (
                          <div key={act.id} className="p-3.5 flex items-start justify-between hover:bg-slate-50/50 transition-colors">
                            <div className="space-y-1">
                              <div className="flex items-center gap-2">
                                <span className="font-semibold text-slate-900 font-mono text-[11px]">
                                  {act.action}
                                </span>
                                <span className="rounded bg-slate-100 px-1.5 py-0.2 text-[10px] text-slate-600 font-medium">
                                  {act.module}
                                </span>
                                <span className={`text-[10px] font-bold ${act.result === 'SUCCESS' ? 'text-emerald-700' : 'text-rose-700'}`}>
                                  {act.result}
                                </span>
                              </div>
                              <p className="text-slate-600 line-clamp-1">{act.description}</p>
                            </div>
                            <span className="text-[10px] text-slate-400 whitespace-nowrap ml-4">
                              {new Date(act.created_at).toLocaleString()}
                            </span>
                          </div>
                        ))}
                      </div>
                    )}
                  </div>
                )}
              </>
            ) : null}
          </div>

          {/* Footer */}
          <div className="border-t border-slate-100 px-6 py-3 bg-slate-50/50 flex items-center justify-between shrink-0">
            <div className="text-[11px] text-slate-400">
              Internal SPortal Governance Boundary
            </div>
            <div className="flex items-center gap-2">
              {detail && (
                <button
                  type="button"
                  onClick={() => setIsRoleModalOpen(true)}
                  className="rounded-xl border border-slate-200 bg-white px-3.5 py-1.5 text-xs font-semibold text-slate-700 hover:bg-slate-50 transition-colors"
                >
                  Reassign Role
                </button>
              )}
              {detail && onStatusChange && (
                detail.status === 'ACTIVE' ? (
                  <button
                    type="button"
                    onClick={() => onStatusChange(detail, 'INACTIVE')}
                    className="rounded-xl border border-rose-200 bg-rose-50 px-3.5 py-1.5 text-xs font-semibold text-rose-700 hover:bg-rose-100 transition-colors"
                  >
                    Deactivate Account
                  </button>
                ) : (
                  <button
                    type="button"
                    onClick={() => onStatusChange(detail, 'ACTIVE')}
                    className="rounded-xl border border-emerald-200 bg-emerald-50 px-3.5 py-1.5 text-xs font-semibold text-emerald-700 hover:bg-emerald-100 transition-colors"
                  >
                    Reactivate Account
                  </button>
                )
              )}
              <button
                type="button"
                onClick={onClose}
                className="rounded-xl bg-slate-800 px-4 py-1.5 text-xs font-semibold text-white hover:bg-slate-700 transition-colors"
              >
                Close
              </button>
            </div>
          </div>
        </div>
      </div>

      {/* Role Change Modal */}
      {detail && (
        <ChangeUserRoleModal
          isOpen={isRoleModalOpen}
          onClose={() => setIsRoleModalOpen(false)}
          user={{
            user_id: detail.user_id,
            email: detail.email,
            full_name: detail.full_name,
            org_id: detail.org_id,
            org_name: detail.org_name,
            role_name: detail.role_name
          }}
          orgId={detail.org_id}
          onRoleChanged={handleRoleChanged}
        />
      )}
    </>
  );
}

import React from 'react';
import { ShieldCheck, Database, Lock, AlertTriangle, CheckCircle2, ArrowRightLeft, Layers, UserCheck } from 'lucide-react';

export function AccessGovernanceBoundaryView() {
  const boundaries = [
    {
      area: 'Customer Super Admin Provisioning',
      sportal: 'Full authority to invite, verify, and provision initial customer Super Admin upon tenant onboarding.',
      cportal: 'Cannot provision initial Super Admin; subsequent Super Admins can be promoted by existing customer Super Admins.',
      sync: 'Same user and org_members record in MariaDB; immediately valid across both portals.'
    },
    {
      area: 'Team Member Lifecycle & Onboarding',
      sportal: 'Can view all tenant personnel, inspect invitations, resend or revoke invitations, and adjust status.',
      cportal: 'Primary customer responsibility: Customer Super Admin sends invitations, assigns departmental roles, and manages staff.',
      sync: 'Real-time synchronization on invitations table; token lifecycle shared.'
    },
    {
      area: 'Customer Role Assignment & Changes',
      sportal: 'Can reassign roles with required administrative reason for compliance, escalations, or emergency remediation.',
      cportal: 'Standard operational self-service: Customer Super Admin reassigns operational roles (Sales, Ops, Finance, etc.).',
      sync: 'Instant update to org_members.role_id; affects user permissions immediately upon next token refresh.'
    },
    {
      area: 'Account Suspension & Reactivation',
      sportal: 'Internal SaaS authority to deactivate customer accounts for billing default, terms violations, or security incidents.',
      cportal: 'Customer Super Admin can deactivate team members who have left the organization.',
      sync: 'Status flag in users table (ACTIVE/INACTIVE); blocked at middleware authentication boundary.'
    },
    {
      area: 'Permission Definition & RBAC Matrix',
      sportal: 'Defines canonical platform resource catalog and system roles; enforces immutable platform boundaries.',
      cportal: 'Configures granular module permissions for non-protected custom roles within tenant constraints.',
      sync: 'Canonical resources and role_permissions tables shared; tenant overrides scoped strictly by tenant_id.'
    },
    {
      area: 'Audit & Compliance Logging',
      sportal: 'Global cross-tenant audit ledger tracking all administrative actions with actor identity and correlation ID.',
      cportal: 'Tenant-scoped activity logs visible to Customer Super Admin for organization events.',
      sync: 'Unified audit_logs table; SPortal queries across orgs; CPortal strictly scoped by tenant_id.'
    }
  ];

  const securityAssurances = [
    {
      title: 'Tenant Isolation & IDOR Shield',
      desc: 'All Go backend endpoints enforce tenant boundary checks. Cross-tenant user access, role manipulation, or invitation tampering returns HTTP 403/404.',
      icon: ShieldCheck,
      status: 'VERIFIED'
    },
    {
      title: 'Sole Super Admin Protection',
      desc: 'SPortal and CPortal backend logic prevents demoting or deactivating the last remaining active Super Admin in any customer organization.',
      icon: UserCheck,
      status: 'ENFORCED'
    },
    {
      title: 'Environment-Gated Test Tokens',
      desc: 'Historical dev/test tokens (e.g. test-token) are strictly gated behind non-production environment flags and return HTTP 401 in production.',
      icon: Lock,
      status: 'HARDENED'
    },
    {
      title: 'Zero Credential Exposure',
      desc: 'Password hashes, authentication tokens, session secrets, and MFA keys are never exposed in SPortal or CPortal API payloads.',
      icon: Database,
      status: 'GUARANTEED'
    }
  ];

  return (
    <div className="space-y-6">
      {/* Header Banner */}
      <div className="rounded-2xl border border-slate-200/80 bg-white p-6 shadow-xs">
        <div className="flex items-center gap-3">
          <div className="flex h-10 w-10 items-center justify-center rounded-xl bg-navy-900 text-white shadow-xs">
            <ArrowRightLeft className="h-5 w-5" />
          </div>
          <div>
            <h2 className="text-base font-bold text-slate-900">
              SPortal vs CPortal Governance & Access Administration Boundary
            </h2>
            <p className="text-xs text-slate-500 max-w-3xl leading-relaxed mt-0.5">
              Clear segregation of administrative duties between LogisticsHQ internal operations (SPortal) and Freight Forwarder customer tenant management (CPortal), unified through a single source-of-truth MariaDB architecture.
            </p>
          </div>
        </div>

        {/* Security Assurances Grid */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 mt-6">
          {securityAssurances.map((item, idx) => {
            const Icon = item.icon;
            return (
              <div key={idx} className="rounded-xl border border-slate-200/90 bg-slate-50/50 p-4 space-y-2">
                <div className="flex items-center justify-between">
                  <div className="flex h-7 w-7 items-center justify-center rounded-lg bg-navy-900/10 text-navy-900">
                    <Icon className="h-4 w-4" />
                  </div>
                  <span className="text-[10px] font-bold px-1.5 py-0.5 rounded bg-emerald-50 text-emerald-700 border border-emerald-200">
                    {item.status}
                  </span>
                </div>
                <h4 className="text-xs font-bold text-slate-900">{item.title}</h4>
                <p className="text-[11px] text-slate-500 leading-relaxed">{item.desc}</p>
              </div>
            );
          })}
        </div>
      </div>

      {/* Responsibility Matrix Table */}
      <div className="rounded-2xl border border-slate-200/80 bg-white shadow-xs overflow-hidden">
        <div className="p-5 border-b border-slate-100 bg-slate-50/50">
          <h3 className="text-xs font-bold text-slate-900 flex items-center gap-2">
            <Layers className="h-4 w-4 text-slate-500" />
            Responsibility Division & Synchronization Matrix
          </h3>
          <p className="text-[11px] text-slate-500 mt-0.5">
            Cross-portal operational capabilities and synchronization mechanisms on persistent MariaDB storage
          </p>
        </div>

        <div className="overflow-x-auto">
          <table className="w-full text-left border-collapse text-xs">
            <thead>
              <tr className="border-b border-slate-200/80 bg-slate-50/80 text-[11px] font-semibold text-slate-600 uppercase tracking-wider">
                <th className="py-3 px-6 w-1/4">Administration Domain</th>
                <th className="py-3 px-6 w-1/3">LogisticsHQ SPortal Scope (Internal)</th>
                <th className="py-3 px-6 w-1/3">Customer CPortal Scope (Tenant)</th>
                <th className="py-3 px-6 w-1/4">Underlying MariaDB State</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-100">
              {boundaries.map((b, idx) => (
                <tr key={idx} className="hover:bg-slate-50/50 transition-colors">
                  <td className="py-4 px-6 font-semibold text-slate-900 align-top">
                    {b.area}
                  </td>
                  <td className="py-4 px-6 text-slate-600 align-top leading-relaxed text-[11px]">
                    <div className="flex items-start gap-1.5">
                      <span className="text-navy-900 font-bold shrink-0">•</span>
                      <span>{b.sportal}</span>
                    </div>
                  </td>
                  <td className="py-4 px-6 text-slate-600 align-top leading-relaxed text-[11px]">
                    <div className="flex items-start gap-1.5">
                      <span className="text-blue-700 font-bold shrink-0">•</span>
                      <span>{b.cportal}</span>
                    </div>
                  </td>
                  <td className="py-4 px-6 text-slate-500 align-top leading-relaxed text-[11px] font-mono bg-slate-50/40">
                    {b.sync}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}

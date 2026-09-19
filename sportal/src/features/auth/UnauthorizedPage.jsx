import React from 'react';
import { Link, useLocation } from 'react-router-dom';
import { ShieldAlert, ArrowLeft, LogOut } from 'lucide-react';
import { useAuth } from '../../context/AuthContext';

export function UnauthorizedPage() {
  const { user, role, logout } = useAuth();
  const location = useLocation();
  const requiredPermission = location.state?.requiredPermission || 'administrative';

  return (
    <div className="flex min-h-screen items-center justify-center bg-slate-50 p-6">
      <div className="w-full max-w-md rounded-2xl border border-slate-200 bg-white p-8 text-center shadow-xl shadow-slate-200/50">
        <div className="mx-auto mb-4 flex h-16 w-16 items-center justify-center rounded-2xl bg-rose-50 text-rose-600 ring-8 ring-rose-50/50">
          <ShieldAlert className="h-8 w-8" />
        </div>

        <h1 className="text-xl font-bold text-slate-900">Access Restricted</h1>
        <p className="mt-2 text-sm text-slate-600">
          Your current SPortal role does not possess the required authority to access this administration module.
        </p>

        <div className="my-6 rounded-xl border border-slate-100 bg-slate-50/75 p-4 text-left">
          <div className="flex items-center justify-between text-xs">
            <span className="text-slate-500">Authenticated Staff:</span>
            <span className="font-semibold text-slate-800">{user?.full_name || user?.email}</span>
          </div>
          <div className="mt-2 flex items-center justify-between text-xs">
            <span className="text-slate-500">Assigned Role:</span>
            <span className="inline-flex items-center rounded-md bg-slate-200 px-2 py-0.5 font-medium text-slate-800">
              {role?.display_name || role?.name || 'Staff'}
            </span>
          </div>
          <div className="mt-2 flex items-center justify-between text-xs">
            <span className="text-slate-500">Required Privilege:</span>
            <span className="font-mono font-medium text-rose-600">{requiredPermission}</span>
          </div>
        </div>

        <div className="flex flex-col gap-2.5">
          <Link
            to="/"
            className="flex w-full items-center justify-center gap-2 rounded-xl bg-slate-900 py-2.5 text-sm font-semibold text-white shadow hover:bg-slate-800 transition-colors"
          >
            <ArrowLeft className="h-4 w-4" /> Return to SPortal Overview
          </Link>

          <button
            type="button"
            onClick={logout}
            className="flex w-full items-center justify-center gap-2 rounded-xl border border-slate-200 py-2.5 text-sm font-semibold text-slate-600 hover:bg-slate-50 hover:text-slate-900 transition-colors"
          >
            <LogOut className="h-4 w-4" /> Sign Out & Switch Account
          </button>
        </div>
      </div>
    </div>
  );
}

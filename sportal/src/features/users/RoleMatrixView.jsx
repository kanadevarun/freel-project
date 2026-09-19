import React, { useState, useEffect } from 'react';
import { Shield, Check, X, Search, Filter, Lock, Settings, Users, Info, Building2 } from 'lucide-react';
import { sportalService } from '../../services/sportalService';

export function RoleMatrixView({ orgId = null, onSelectRoleForInspection }) {
  const [data, setData] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [selectedRole, setSelectedRole] = useState('SUPER_ADMIN');
  const [searchFilter, setSearchFilter] = useState('');

  useEffect(() => {
    fetchMatrix();
  }, [orgId]);

  const fetchMatrix = async () => {
    try {
      setLoading(true);
      setError('');
      const res = await sportalService.getPermissionMatrix(orgId);
      const resData = res?.data || res;
      setData(resData);
      if (resData?.roles?.length > 0) {
        const exists = resData.roles.some((r) => (r.name || r.role_id) === selectedRole);
        if (!exists) {
          setSelectedRole(resData.roles[0].name || resData.roles[0].role_id);
        }
      }
    } catch (err) {
      setError(err?.response?.data?.message || err.message || 'Failed to load permission matrix.');
    } finally {
      setLoading(false);
    }
  };

  const currentRoleObj =
    data?.roles?.find((r) => (r.name || r.role_id) === selectedRole) || data?.roles?.[0];

  const filteredResources = (data?.resources || []).filter((res) => {
    if (!searchFilter.trim()) return true;
    const q = searchFilter.toLowerCase();
    return (
      res.resource.toLowerCase().includes(q) ||
      res.name.toLowerCase().includes(q) ||
      res.description.toLowerCase().includes(q) ||
      res.category.toLowerCase().includes(q)
    );
  });

  return (
    <div className="space-y-6">
      {/* Overview Banner */}
      <div className="rounded-2xl border border-slate-200/80 bg-white p-6 shadow-xs">
        <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
          <div className="space-y-1">
            <div className="flex items-center gap-2.5">
              <div className="flex h-9 w-9 items-center justify-center rounded-xl bg-navy-900 text-white shadow-xs">
                <Shield className="h-5 w-5" />
              </div>
              <h2 className="text-base font-bold text-slate-900">Role Catalog & Canonical Permission Matrix</h2>
            </div>
            <p className="text-xs text-slate-500 max-w-2xl leading-relaxed">
              LogisticsHQ RBAC architecture enforcing strict multi-tenant authorization across freight-forwarder organizations. Inspect canonical system roles, active user distributions, and granular CRUD permissions.
            </p>
          </div>
          {data?.org_name && (
            <div className="flex items-center gap-2 px-3 py-1.5 rounded-xl bg-slate-50 border border-slate-200 text-xs text-slate-700">
              <Building2 className="h-4 w-4 text-slate-500" />
              <span>Tenant Scope: <strong>{data.org_name}</strong></span>
            </div>
          )}
        </div>

        {/* Role Cards Grid */}
        <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-7 gap-3 mt-6">
          {(data?.roles || []).map((r) => {
            const roleIdentifier = r.name || r.role_name || r.role_id;
            const isSelected = roleIdentifier === selectedRole;
            return (
              <button
                key={roleIdentifier}
                onClick={() => setSelectedRole(roleIdentifier)}
                className={`flex flex-col text-left p-3.5 rounded-xl border transition-all ${
                  isSelected
                    ? 'border-navy-900 bg-navy-900/5 ring-2 ring-navy-900/20 shadow-xs'
                    : 'border-slate-200/80 bg-white hover:border-slate-300 hover:bg-slate-50/50'
                }`}
              >
                <div className="flex items-center justify-between w-full">
                  <span className="font-mono text-xs font-bold text-slate-900 truncate">
                    {roleIdentifier}
                  </span>
                  {r.is_protected ? (
                    <Lock className="h-3 w-3 text-amber-500 shrink-0" title="Protected System Role" />
                  ) : (
                    <Settings className="h-3 w-3 text-slate-400 shrink-0" title="Configurable Role" />
                  )}
                </div>
                <p className="text-[11px] text-slate-500 line-clamp-1 mt-1">{r.description}</p>
                <div className="mt-3 flex items-center justify-between text-[10px] pt-2 border-t border-slate-100">
                  <span className="text-slate-400 flex items-center gap-1">
                    <Users className="h-3 w-3" />
                    {r.user_count} {r.user_count === 1 ? 'user' : 'users'}
                  </span>
                  <span
                    className={`font-semibold px-1.5 py-0.5 rounded ${
                      r.is_system ? 'bg-slate-100 text-slate-700' : 'bg-blue-50 text-blue-700'
                    }`}
                  >
                    {r.is_system ? 'SYSTEM' : 'CUSTOM'}
                  </span>
                </div>
              </button>
            );
          })}
        </div>
      </div>

      {/* Selected Role Detail & Module Matrix */}
      {currentRoleObj && (
        <div className="rounded-2xl border border-slate-200/80 bg-white shadow-xs overflow-hidden">
          {/* Role Header Info */}
          <div className="p-6 border-b border-slate-100 bg-slate-50/50 flex flex-col md:flex-row md:items-center justify-between gap-4">
            <div className="space-y-1">
              <div className="flex items-center gap-3">
                <span className="px-2.5 py-1 rounded-md text-xs font-bold font-mono bg-navy-900 text-white">
                  {currentRoleObj.name || currentRoleObj.role_name}
                </span>
                <h3 className="text-sm font-bold text-slate-900">
                  {(currentRoleObj.name || currentRoleObj.role_name) === 'SUPER_ADMIN'
                    ? 'Customer Super Administrator'
                    : currentRoleObj.name || currentRoleObj.role_name}
                </h3>
                <span className="text-xs text-slate-400">•</span>
                <span className="text-xs text-slate-500">
                  {currentRoleObj.user_count}{' '}
                  {currentRoleObj.user_count === 1 ? 'active tenant member' : 'active tenant members'}
                </span>
              </div>
              <p className="text-xs text-slate-600 mt-1 leading-relaxed">
                {currentRoleObj.description}
              </p>
            </div>

            {/* Search Filter */}
            <div className="relative w-full md:w-64">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-3.5 w-3.5 text-slate-400" />
              <input
                type="text"
                value={searchFilter}
                onChange={(e) => setSearchFilter(e.target.value)}
                placeholder="Filter platform areas..."
                className="w-full rounded-xl border border-slate-200 bg-white pl-9 pr-3 py-1.5 text-xs text-slate-800 placeholder:text-slate-400 focus:border-navy-900 focus:outline-none focus:ring-1 focus:ring-navy-900"
              />
            </div>
          </div>

          {/* Matrix Table */}
          <div className="overflow-x-auto">
            <table className="w-full text-left border-collapse text-xs">
              <thead>
                <tr className="border-b border-slate-200/80 bg-slate-50/80 text-[11px] font-semibold text-slate-600 uppercase tracking-wider">
                  <th className="py-3 px-6">Platform Resource Area</th>
                  <th className="py-3 px-4 text-center">Create</th>
                  <th className="py-3 px-4 text-center">Read</th>
                  <th className="py-3 px-4 text-center">Update</th>
                  <th className="py-3 px-4 text-center">Delete</th>
                  <th className="py-3 px-6">Effective Capability Level</th>
                  <th className="py-3 px-6">Administration Boundary</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-100">
                {filteredResources.map((res) => {
                  let mod = null;
                  if (Array.isArray(currentRoleObj.effective_access)) {
                    mod = currentRoleObj.effective_access.find((m) => m.resource === res.resource);
                  } else if (currentRoleObj.effective_access) {
                    mod = currentRoleObj.effective_access[res.resource];
                  }

                  const canCreate = mod?.allowed_actions
                    ? mod.allowed_actions.includes('CREATE')
                    : Boolean(mod?.can_create);
                  const canRead = mod?.allowed_actions
                    ? mod.allowed_actions.includes('READ')
                    : Boolean(mod?.can_read);
                  const canUpdate = mod?.allowed_actions
                    ? mod.allowed_actions.includes('UPDATE')
                    : Boolean(mod?.can_update);
                  const canDelete = mod?.allowed_actions
                    ? mod.allowed_actions.includes('DELETE')
                    : Boolean(mod?.can_delete);

                  const accessLevel = mod?.access_level || 'NONE';

                  const isFull = accessLevel === 'FULL';
                  const isManage = accessLevel === 'MANAGE';
                  const isReadOnly = accessLevel === 'VIEW_ONLY';

                  let badgeColor = 'bg-slate-100 text-slate-600 border-slate-200';
                  if (isFull) badgeColor = 'bg-emerald-50 text-emerald-700 border-emerald-200';
                  else if (isManage) badgeColor = 'bg-blue-50 text-blue-700 border-blue-200';
                  else if (isReadOnly) badgeColor = 'bg-amber-50 text-amber-700 border-amber-200';

                  return (
                    <tr key={res.resource} className="hover:bg-slate-50/50 transition-colors">
                      <td className="py-3.5 px-6">
                        <div className="font-semibold text-slate-900 flex items-center gap-2">
                          <span>{res.name}</span>
                          <span className="font-mono text-[10px] text-slate-400 bg-slate-100 px-1.5 py-0.5 rounded">
                            {res.resource}
                          </span>
                        </div>
                        <p className="text-[11px] text-slate-500 mt-0.5">{res.description}</p>
                      </td>

                      {/* CREATE */}
                      <td className="py-3.5 px-4 text-center">
                        {canCreate ? (
                          <span className="inline-flex items-center justify-center h-6 w-6 rounded-full bg-emerald-50 text-emerald-600 border border-emerald-200">
                            <Check className="h-3.5 w-3.5" />
                          </span>
                        ) : (
                          <span className="inline-flex items-center justify-center h-6 w-6 rounded-full bg-slate-50 text-slate-300">
                            <X className="h-3 w-3" />
                          </span>
                        )}
                      </td>

                      {/* READ */}
                      <td className="py-3.5 px-4 text-center">
                        {canRead ? (
                          <span className="inline-flex items-center justify-center h-6 w-6 rounded-full bg-emerald-50 text-emerald-600 border border-emerald-200">
                            <Check className="h-3.5 w-3.5" />
                          </span>
                        ) : (
                          <span className="inline-flex items-center justify-center h-6 w-6 rounded-full bg-slate-50 text-slate-300">
                            <X className="h-3 w-3" />
                          </span>
                        )}
                      </td>

                      {/* UPDATE */}
                      <td className="py-3.5 px-4 text-center">
                        {canUpdate ? (
                          <span className="inline-flex items-center justify-center h-6 w-6 rounded-full bg-emerald-50 text-emerald-600 border border-emerald-200">
                            <Check className="h-3.5 w-3.5" />
                          </span>
                        ) : (
                          <span className="inline-flex items-center justify-center h-6 w-6 rounded-full bg-slate-50 text-slate-300">
                            <X className="h-3 w-3" />
                          </span>
                        )}
                      </td>

                      {/* DELETE */}
                      <td className="py-3.5 px-4 text-center">
                        {canDelete ? (
                          <span className="inline-flex items-center justify-center h-6 w-6 rounded-full bg-emerald-50 text-emerald-600 border border-emerald-200">
                            <Check className="h-3.5 w-3.5" />
                          </span>
                        ) : (
                          <span className="inline-flex items-center justify-center h-6 w-6 rounded-full bg-slate-50 text-slate-300">
                            <X className="h-3 w-3" />
                          </span>
                        )}
                      </td>

                      {/* LEVEL */}
                      <td className="py-3.5 px-6">
                        <span
                          className={`inline-flex items-center px-2.5 py-1 rounded-full text-[11px] font-bold border ${badgeColor}`}
                        >
                          {accessLevel}
                        </span>
                      </td>

                      {/* BOUNDARY */}
                      <td className="py-3.5 px-6 text-slate-500 text-[11px]">
                        {res.resource === 'USERS' || res.resource === 'SETTINGS' ? (
                          <span className="text-amber-700 font-medium">Customer Admin (CPortal)</span>
                        ) : (
                          <span>Operational Module (CPortal)</span>
                        )}
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>

          {/* Footer Notice */}
          <div className="p-4 bg-slate-50/60 border-t border-slate-100 flex items-center justify-between text-[11px] text-slate-500">
            <div className="flex items-center gap-2">
              <Info className="h-3.5 w-3.5 text-blue-600 shrink-0" />
              <span>
                System roles are protected by canonical security constraints. Custom permission overrides are configured via customer organization administrative policies.
              </span>
            </div>
            <span className="font-mono text-slate-400">Total Canonical Areas: {filteredResources.length}</span>
          </div>
        </div>
      )}
    </div>
  );
}

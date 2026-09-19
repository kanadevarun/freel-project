import React, { useState, useEffect, useCallback } from 'react';
import { useParams, Link, useSearchParams } from 'react-router-dom';
import {
  Users,
  Search,
  Filter,
  UserPlus,
  RefreshCw,
  Eye,
  UserX,
  UserCheck,
  Building,
  Shield,
  Clock,
  CheckCircle,
  AlertCircle,
  Mail,
  Send,
  Trash2,
  ChevronLeft,
  ChevronRight,
  RotateCcw,
  Layers,
  ArrowRightLeft,
  Edit3
} from 'lucide-react';
import { sportalService } from '../../services/sportalService';
import { StatusBadge } from '../../components/common/StatusBadge';
import { InviteCustomerUserModal } from './InviteCustomerUserModal';
import { CustomerUserDetailModal } from './CustomerUserDetailModal';
import { ConfirmUserStatusModal } from './ConfirmUserStatusModal';
import { RoleMatrixView } from './RoleMatrixView';
import { AccessGovernanceBoundaryView } from './AccessGovernanceBoundaryView';
import { ChangeUserRoleModal } from './ChangeUserRoleModal';

export function UsersPage() {
  const { organizationId } = useParams();
  const [searchParams] = useSearchParams();
  const urlTab = searchParams.get('tab');

  const [activeMainTab, setActiveMainTab] = useState(
    urlTab === 'matrix' ? 'matrix' : urlTab === 'governance' ? 'governance' : 'directory'
  );

  useEffect(() => {
    if (urlTab && ['directory', 'matrix', 'governance'].includes(urlTab)) {
      setActiveMainTab(urlTab);
    }
  }, [urlTab]);

  const [users, setUsers] = useState([]);
  const [metrics, setMetrics] = useState({
    total_users: 0,
    active_users: 0,
    inactive_users: 0,
    pending_invitations: 0,
    super_admins_count: 0,
  });
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  // Organization List for Dropdown
  const [orgs, setOrgs] = useState([]);

  // Search & Filter State
  const [search, setSearch] = useState('');
  const [selectedOrg, setSelectedOrg] = useState(organizationId || '');
  const [selectedRole, setSelectedRole] = useState('');
  const [selectedStatus, setSelectedStatus] = useState('');
  const [selectedInvStatus, setSelectedInvStatus] = useState('');

  // Pagination State
  const [page, setPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [totalCount, setTotalCount] = useState(0);
  const limit = 15;

  // Modals
  const [isInviteModalOpen, setIsInviteModalOpen] = useState(false);
  const [selectedDetailUser, setSelectedDetailUser] = useState(null);
  const [statusConfirmState, setStatusConfirmState] = useState({
    isOpen: false,
    user: null,
    newStatus: '',
  });
  const [roleChangeState, setRoleChangeState] = useState({
    isOpen: false,
    user: null,
  });

  // Action loading state (resending, revoking)
  const [actionLoadingId, setActionLoadingId] = useState(null);
  const [actionFeedback, setActionFeedback] = useState({ message: '', type: '' });

  // Fetch Organizations once
  useEffect(() => {
    async function loadOrgs() {
      try {
        const res = await sportalService.getOrganizations({ pageSize: 100 });
        const orgItems = res?.items || res?.data?.items || [];
        const customerOrgs = orgItems.filter((o) => o.id !== 1);
        setOrgs(customerOrgs);
      } catch (err) {
        console.error('Failed to load organizations for filter:', err);
      }
    }
    loadOrgs();
  }, []);

  // Update selectedOrg if URL param changes
  useEffect(() => {
    if (organizationId) {
      setSelectedOrg(organizationId);
    }
  }, [organizationId]);

  // Fetch customer users
  const fetchUsers = useCallback(async () => {
    try {
      setLoading(true);
      setError('');

      const res = await sportalService.getCustomerUsers({
        search: search.trim() || undefined,
        orgId: selectedOrg || undefined,
        role: selectedRole || undefined,
        status: selectedStatus || undefined,
        invitationStatus: selectedInvStatus || undefined,
        page,
        limit,
        sortBy: 'created_at',
        sortOrder: 'desc',
      });

      const payload = res?.data || res || {};
      setUsers(payload.items || []);
      setTotalCount(payload.total || 0);
      setTotalPages(payload.total_pages || 1);
      if (payload.metrics) {
        setMetrics(payload.metrics);
      }
    } catch (err) {
      setError(err?.response?.data?.message || err.message || 'Failed to load customer user directory.');
    } finally {
      setLoading(false);
    }
  }, [search, selectedOrg, selectedRole, selectedStatus, selectedInvStatus, page]);

  useEffect(() => {
    fetchUsers();
  }, [fetchUsers]);

  const handleResetFilters = () => {
    setSearch('');
    setSelectedOrg(organizationId || '');
    setSelectedRole('');
    setSelectedStatus('');
    setSelectedInvStatus('');
    setPage(1);
  };

  const handleResendInvite = async (invitationId) => {
    if (!invitationId) return;
    try {
      setActionLoadingId(invitationId);
      setActionFeedback({ message: '', type: '' });
      await sportalService.resendCustomerInvitation(invitationId);
      setActionFeedback({ message: `Invitation #${invitationId} refreshed and resent.`, type: 'success' });
      fetchUsers();
    } catch (err) {
      setActionFeedback({
        message: err?.response?.data?.message || err.message || 'Failed to resend invitation.',
        type: 'error',
      });
    } finally {
      setActionLoadingId(null);
    }
  };

  const handleRevokeInvite = async (invitationId) => {
    if (!invitationId) return;
    if (!window.confirm(`Are you sure you want to revoke invitation #${invitationId}?`)) return;

    try {
      setActionLoadingId(invitationId);
      setActionFeedback({ message: '', type: '' });
      await sportalService.revokeCustomerInvitation(invitationId);
      setActionFeedback({ message: `Invitation #${invitationId} revoked.`, type: 'success' });
      fetchUsers();
    } catch (err) {
      setActionFeedback({
        message: err?.response?.data?.message || err.message || 'Failed to revoke invitation.',
        type: 'error',
      });
    } finally {
      setActionLoadingId(null);
    }
  };

  return (
    <div className="space-y-6">
      {/* Top Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-slate-900 tracking-tight">
            Customer Users & Access Lifecycle
          </h1>
          <p className="text-xs text-slate-500 mt-1">
            Authoritative cross-tenant directory of freight forwarding personnel, initial Super Admins, and pending invitations.
          </p>
        </div>
        <div className="flex items-center gap-2.5">
          <button
            type="button"
            onClick={fetchUsers}
            disabled={loading}
            className="flex items-center gap-1.5 rounded-xl border border-slate-200 bg-white px-3.5 py-2 text-xs font-semibold text-slate-700 hover:bg-slate-50 transition-colors shadow-2xs"
            title="Refresh User Directory"
          >
            <RefreshCw className={`h-3.5 w-3.5 ${loading ? 'animate-spin' : ''}`} />
            <span className="hidden sm:inline">Refresh</span>
          </button>
          <button
            type="button"
            onClick={() => setIsInviteModalOpen(true)}
            className="flex items-center gap-2 rounded-xl bg-navy-900 px-4 py-2 text-xs font-semibold text-white hover:bg-navy-800 shadow-xs transition-colors"
          >
            <UserPlus className="h-4 w-4" />
            <span>Invite Customer User</span>
          </button>
        </div>
      </div>

      {/* Primary Navigation Tabs */}
      <div className="border-b border-slate-200">
        <div className="flex space-x-8">
          {[
            { id: 'directory', label: 'Customer Personnel Directory', icon: Users },
            { id: 'matrix', label: 'Role Catalog & Permission Matrix', icon: Layers },
            { id: 'governance', label: 'Access Governance & SPortal/CPortal Boundary', icon: ArrowRightLeft },
          ].map((tab) => {
            const Icon = tab.icon;
            const isActive = activeMainTab === tab.id;
            return (
              <button
                key={tab.id}
                onClick={() => setActiveMainTab(tab.id)}
                className={`flex items-center gap-2 py-3 text-xs font-semibold border-b-2 transition-colors ${
                  isActive
                    ? 'border-navy-900 text-navy-900'
                    : 'border-transparent text-slate-500 hover:text-slate-800'
                }`}
              >
                <Icon className="h-4 w-4" />
                <span>{tab.label}</span>
              </button>
            );
          })}
        </div>
      </div>

      {/* Action Notification Banner */}
      {actionFeedback.message && (
        <div
          className={`flex items-center justify-between rounded-xl border p-3 text-xs transition-all ${
            actionFeedback.type === 'success'
              ? 'border-emerald-200 bg-emerald-50 text-emerald-800'
              : 'border-rose-200 bg-rose-50 text-rose-800'
          }`}
        >
          <div className="flex items-center gap-2">
            {actionFeedback.type === 'success' ? (
              <CheckCircle className="h-4 w-4 text-emerald-600 shrink-0" />
            ) : (
              <AlertCircle className="h-4 w-4 text-rose-600 shrink-0" />
            )}
            <span>{actionFeedback.message}</span>
          </div>
          <button
            onClick={() => setActionFeedback({ message: '', type: '' })}
            className="text-slate-400 hover:text-slate-600"
          >
            ×
          </button>
        </div>
      )}

      {/* TAB 1: CUSTOMER PERSONNEL DIRECTORY */}
      {activeMainTab === 'directory' && (
        <>
          {/* KPI Workforce Metrics */}
          <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
        <div className="rounded-xl border border-slate-200 bg-white p-4 shadow-xs">
          <div className="flex items-center justify-between">
            <span className="text-xs font-semibold text-slate-500">Total Accounts</span>
            <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-blue-50 text-blue-600">
              <Users className="h-4 w-4" />
            </div>
          </div>
          <p className="text-2xl font-bold text-slate-900 mt-2">{metrics.total_users}</p>
          <span className="text-[11px] text-slate-400">All registered & invited</span>
        </div>

        <div className="rounded-xl border border-slate-200 bg-white p-4 shadow-xs">
          <div className="flex items-center justify-between">
            <span className="text-xs font-semibold text-slate-500">Active Workforce</span>
            <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-emerald-50 text-emerald-600">
              <CheckCircle className="h-4 w-4" />
            </div>
          </div>
          <p className="text-2xl font-bold text-slate-900 mt-2">{metrics.active_users}</p>
          <span className="text-[11px] text-slate-400">Normal operational status</span>
        </div>

        <div className="rounded-xl border border-slate-200 bg-white p-4 shadow-xs">
          <div className="flex items-center justify-between">
            <span className="text-xs font-semibold text-slate-500">Pending Invitations</span>
            <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-amber-50 text-amber-600">
              <Mail className="h-4 w-4" />
            </div>
          </div>
          <p className="text-2xl font-bold text-slate-900 mt-2">{metrics.pending_invitations}</p>
          <span className="text-[11px] text-slate-400">Awaiting user activation</span>
        </div>

        <div className="rounded-xl border border-slate-200 bg-white p-4 shadow-xs">
          <div className="flex items-center justify-between">
            <span className="text-xs font-semibold text-slate-500">Customer Super Admins</span>
            <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-purple-50 text-purple-600">
              <Shield className="h-4 w-4" />
            </div>
          </div>
          <p className="text-2xl font-bold text-slate-900 mt-2">{metrics.super_admins_count}</p>
          <span className="text-[11px] text-slate-400">Authorized primary admins</span>
        </div>
      </div>

      {/* Filter & Search Toolbar */}
      <div className="rounded-xl border border-slate-200 bg-white p-4 shadow-xs space-y-3">
        <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-5 gap-3">
          {/* Search Input */}
          <div className="relative md:col-span-2">
            <Search className="absolute left-3 top-2.5 h-4 w-4 text-slate-400" />
            <input
              type="text"
              placeholder="Search by name, email, or organization..."
              value={search}
              onChange={(e) => {
                setSearch(e.target.value);
                setPage(1);
              }}
              className="w-full rounded-xl border border-slate-200 bg-white py-2 pl-9 pr-3 text-xs text-slate-800 placeholder:text-slate-400 focus:border-navy-900 focus:outline-hidden"
            />
          </div>

          {/* Organization Filter */}
          <div>
            <select
              value={selectedOrg}
              onChange={(e) => {
                setSelectedOrg(e.target.value);
                setPage(1);
              }}
              className="w-full rounded-xl border border-slate-200 bg-white py-2 px-3 text-xs text-slate-800 focus:border-navy-900 focus:outline-hidden"
            >
              <option value="">All Organizations</option>
              {orgs.map((o) => (
                <option key={o.id} value={o.id}>
                  {o.name}
                </option>
              ))}
            </select>
          </div>

          {/* Role Filter */}
          <div>
            <select
              value={selectedRole}
              onChange={(e) => {
                setSelectedRole(e.target.value);
                setPage(1);
              }}
              className="w-full rounded-xl border border-slate-200 bg-white py-2 px-3 text-xs text-slate-800 focus:border-navy-900 focus:outline-hidden"
            >
              <option value="">All Roles</option>
              <option value="SUPER_ADMIN">Super Admin</option>
              <option value="SALES">Sales</option>
              <option value="OPERATIONS">Operations</option>
              <option value="FINANCE">Finance</option>
              <option value="PRICING">Pricing</option>
              <option value="DOCUMENTATION">Documentation</option>
              <option value="HR">HR</option>
            </select>
          </div>

          {/* Account Status Filter */}
          <div>
            <select
              value={selectedStatus}
              onChange={(e) => {
                setSelectedStatus(e.target.value);
                setPage(1);
              }}
              className="w-full rounded-xl border border-slate-200 bg-white py-2 px-3 text-xs text-slate-800 focus:border-navy-900 focus:outline-hidden"
            >
              <option value="">All Statuses</option>
              <option value="ACTIVE">Active</option>
              <option value="INACTIVE">Inactive</option>
              <option value="PENDING">Pending</option>
              <option value="DISABLED">Disabled</option>
              <option value="SUSPENDED">Suspended</option>
            </select>
          </div>
        </div>

        {/* Second row filters & reset */}
        <div className="flex items-center justify-between pt-1 border-t border-slate-100 text-xs text-slate-500">
          <div className="flex items-center gap-3">
            <span className="font-semibold text-slate-600 flex items-center gap-1">
              <Filter className="h-3 w-3" /> Filters:
            </span>
            <select
              value={selectedInvStatus}
              onChange={(e) => {
                setSelectedInvStatus(e.target.value);
                setPage(1);
              }}
              className="rounded-lg border border-slate-200 bg-white py-1 px-2 text-[11px] text-slate-700"
            >
              <option value="">Invitation: All</option>
              <option value="ACCEPTED">Accepted</option>
              <option value="PENDING">Pending</option>
              <option value="EXPIRED">Expired</option>
            </select>
          </div>

          {(search || selectedOrg || selectedRole || selectedStatus || selectedInvStatus) && (
            <button
              onClick={handleResetFilters}
              className="flex items-center gap-1 text-slate-500 hover:text-navy-900 transition-colors"
            >
              <RotateCcw className="h-3 w-3" />
              <span>Reset Filters</span>
            </button>
          )}
        </div>
      </div>

      {/* Directory Table */}
      <div className="rounded-xl border border-slate-200 bg-white shadow-xs overflow-hidden">
        {loading ? (
          <div className="py-20 text-center space-y-3">
            <div className="mx-auto h-8 w-8 animate-spin rounded-full border-2 border-navy-900 border-t-transparent"></div>
            <p className="text-xs text-slate-500">Querying persistent customer workforce records...</p>
          </div>
        ) : error ? (
          <div className="p-8 text-center space-y-2">
            <AlertCircle className="mx-auto h-8 w-8 text-rose-500" />
            <p className="text-xs font-semibold text-rose-700">{error}</p>
            <button
              onClick={fetchUsers}
              className="rounded-lg bg-slate-100 px-3 py-1.5 text-xs font-semibold text-slate-700 hover:bg-slate-200 transition-colors"
            >
              Try Again
            </button>
          </div>
        ) : users.length === 0 ? (
          <div className="py-20 text-center space-y-2">
            <Users className="mx-auto h-8 w-8 text-slate-300" />
            <p className="text-sm font-semibold text-slate-800">No customer users found</p>
            <p className="text-xs text-slate-400 max-w-sm mx-auto">
              No persistent users match your current filter parameters. Try adjusting your search query or reset filters.
            </p>
            {(search || selectedOrg || selectedRole || selectedStatus || selectedInvStatus) && (
              <button
                onClick={handleResetFilters}
                className="mt-2 rounded-lg bg-navy-900 px-3 py-1.5 text-xs font-semibold text-white hover:bg-navy-800 transition-colors"
              >
                Clear Filters
              </button>
            )}
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-left text-xs">
              <thead>
                <tr className="border-b border-slate-200 bg-slate-50/75 text-slate-500 font-semibold uppercase tracking-wider text-[11px]">
                  <th className="py-3 px-4">User</th>
                  <th className="py-3 px-4">Customer Organization</th>
                  <th className="py-3 px-4">Assigned Role</th>
                  <th className="py-3 px-4">Account Status</th>
                  <th className="py-3 px-4">Invitation</th>
                  <th className="py-3 px-4">Last Login</th>
                  <th className="py-3 px-4 text-right">Actions</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-100">
                {users.map((u, idx) => {
                  const isPendingInvite = u.user_id === 0 || u.status === 'PENDING';
                  const isSuperAdmin = u.role_name === 'SUPER_ADMIN';

                  return (
                    <tr key={`${u.org_id}-${u.user_id}-${u.email}-${idx}`} className="hover:bg-slate-50/60 transition-colors">
                      {/* User Column */}
                      <td className="py-3.5 px-4">
                        <div className="flex items-center gap-3">
                          <div
                            className={`flex h-8 w-8 shrink-0 items-center justify-center rounded-lg font-bold text-xs ${
                              isSuperAdmin
                                ? 'bg-purple-100 text-purple-800 border border-purple-200'
                                : 'bg-slate-100 text-slate-700 border border-slate-200'
                            }`}
                          >
                            {u.full_name ? u.full_name.charAt(0).toUpperCase() : u.email.charAt(0).toUpperCase()}
                          </div>
                          <div className="min-w-0">
                            <div className="flex items-center gap-1.5">
                              <p className="font-semibold text-slate-900 truncate">
                                {u.full_name || u.email.split('@')[0]}
                              </p>
                              {u.user_id > 0 && (
                                <span className="font-mono text-[10px] text-slate-400 font-normal">
                                  #{u.user_id}
                                </span>
                              )}
                            </div>
                            <p className="text-[11px] text-slate-500 font-mono truncate">{u.email}</p>
                          </div>
                        </div>
                      </td>

                      {/* Customer Organization Column */}
                      <td className="py-3.5 px-4">
                        <Link
                          to={`/organizations/${u.org_id}`}
                          className="font-medium text-navy-900 hover:underline flex items-center gap-1"
                        >
                          <Building className="h-3 w-3 text-slate-400 shrink-0" />
                          <span className="truncate max-w-[180px]">{u.org_name}</span>
                        </Link>
                        <span className="text-[10px] text-slate-400 block font-mono">Org #{u.org_id}</span>
                      </td>

                      {/* Role Column */}
                      <td className="py-3.5 px-4">
                        <span
                          className={`inline-flex items-center px-2 py-0.5 rounded text-[11px] font-semibold border ${
                            isSuperAdmin
                              ? 'bg-purple-50 text-purple-800 border-purple-200'
                              : 'bg-slate-100 text-slate-800 border-slate-200'
                          }`}
                        >
                          {u.role_name}
                        </span>
                      </td>

                      {/* Status Column */}
                      <td className="py-3.5 px-4">
                        <StatusBadge status={u.status} />
                      </td>

                      {/* Invitation Column */}
                      <td className="py-3.5 px-4">
                        <span
                          className={`inline-flex items-center px-2 py-0.5 rounded-full text-[11px] font-medium border ${
                            u.invitation_status === 'ACCEPTED'
                              ? 'bg-emerald-50 text-emerald-700 border-emerald-200'
                              : u.invitation_status === 'PENDING'
                              ? 'bg-blue-50 text-blue-700 border-blue-200'
                              : 'bg-rose-50 text-rose-700 border-rose-200'
                          }`}
                        >
                          {u.invitation_status}
                        </span>
                      </td>

                      {/* Last Login Column */}
                      <td className="py-3.5 px-4 text-slate-600">
                        {u.last_login ? (
                          <div className="space-y-0.5">
                            <span className="font-medium text-slate-800 block">
                              {new Date(u.last_login).toLocaleDateString()}
                            </span>
                            <span className="text-[10px] text-slate-400 block">
                              {new Date(u.last_login).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
                            </span>
                          </div>
                        ) : (
                          <span className="text-slate-400 italic">Never logged in</span>
                        )}
                      </td>

                      {/* Actions Column */}
                      <td className="py-3.5 px-4 text-right">
                        <div className="flex items-center justify-end gap-1.5">
                          {/* View details */}
                          {u.user_id > 0 && (
                            <button
                              type="button"
                              onClick={() => setSelectedDetailUser(u)}
                              className="rounded-lg p-1.5 text-slate-500 hover:bg-slate-100 hover:text-navy-900 transition-colors"
                              title="View Customer User Dossier"
                            >
                              <Eye className="h-4 w-4" />
                            </button>
                          )}

                          {/* Reassign Role (Task S7) */}
                          {u.user_id > 0 && (
                            <button
                              type="button"
                              onClick={() => setRoleChangeState({ isOpen: true, user: u })}
                              className="rounded-lg p-1.5 text-slate-500 hover:bg-slate-100 hover:text-navy-900 transition-colors"
                              title="Reassign Customer User Role"
                            >
                              <Edit3 className="h-4 w-4" />
                            </button>
                          )}

                          {/* Pending Invite Actions */}
                          {isPendingInvite && u.invitation_id && (
                            <>
                              <button
                                type="button"
                                disabled={actionLoadingId === u.invitation_id}
                                onClick={() => handleResendInvite(u.invitation_id)}
                                className="rounded-lg p-1.5 text-sky-600 hover:bg-sky-50 transition-colors"
                                title="Resend Invitation Token"
                              >
                                <Send className={`h-4 w-4 ${actionLoadingId === u.invitation_id ? 'animate-pulse' : ''}`} />
                              </button>
                              <button
                                type="button"
                                disabled={actionLoadingId === u.invitation_id}
                                onClick={() => handleRevokeInvite(u.invitation_id)}
                                className="rounded-lg p-1.5 text-rose-600 hover:bg-rose-50 transition-colors"
                                title="Revoke Pending Invitation"
                              >
                                <Trash2 className="h-4 w-4" />
                              </button>
                            </>
                          )}

                          {/* Active / Inactive Status Actions */}
                          {u.user_id > 0 && u.status === 'ACTIVE' && (
                            <button
                              type="button"
                              onClick={() =>
                                setStatusConfirmState({
                                  isOpen: true,
                                  user: u,
                                  newStatus: 'INACTIVE',
                                })
                              }
                              className="rounded-lg p-1.5 text-slate-400 hover:bg-rose-50 hover:text-rose-600 transition-colors"
                              title="Deactivate User Access"
                            >
                              <UserX className="h-4 w-4" />
                            </button>
                          )}

                          {u.user_id > 0 && u.status !== 'ACTIVE' && (
                            <button
                              type="button"
                              onClick={() =>
                                setStatusConfirmState({
                                  isOpen: true,
                                  user: u,
                                  newStatus: 'ACTIVE',
                                })
                              }
                              className="rounded-lg p-1.5 text-emerald-600 hover:bg-emerald-50 transition-colors"
                              title="Reactivate User Access"
                            >
                              <UserCheck className="h-4 w-4" />
                            </button>
                          )}
                        </div>
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        )}

        {/* Pagination Footer */}
        {!loading && users.length > 0 && (
          <div className="border-t border-slate-100 px-4 py-3 bg-slate-50/50 flex flex-col sm:flex-row items-center justify-between gap-3 text-xs text-slate-500">
            <div>
              Showing <span className="font-semibold text-slate-800">{(page - 1) * limit + 1}</span> to{' '}
              <span className="font-semibold text-slate-800">{Math.min(page * limit, totalCount)}</span> of{' '}
              <span className="font-semibold text-slate-800">{totalCount}</span> customer personnel
            </div>
            <div className="flex items-center gap-1.5">
              <button
                type="button"
                disabled={page <= 1}
                onClick={() => setPage((p) => Math.max(1, p - 1))}
                className="flex items-center gap-1 rounded-lg border border-slate-200 bg-white px-3 py-1 text-slate-600 hover:bg-slate-50 disabled:opacity-40 transition-colors"
              >
                <ChevronLeft className="h-3.5 w-3.5" />
                <span>Previous</span>
              </button>
              <span className="px-2 font-medium text-slate-700">
                Page {page} of {totalPages}
              </span>
              <button
                type="button"
                disabled={page >= totalPages}
                onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
                className="flex items-center gap-1 rounded-lg border border-slate-200 bg-white px-3 py-1 text-slate-600 hover:bg-slate-50 disabled:opacity-40 transition-colors"
              >
                <span>Next</span>
                <ChevronRight className="h-3.5 w-3.5" />
              </button>
            </div>
          </div>
        )}
      </div>
      </>
      )}

      {/* TAB 2: ROLE CATALOG & PERMISSION MATRIX (TASK S7) */}
      {activeMainTab === 'matrix' && (
        <RoleMatrixView
          orgId={selectedOrg || undefined}
        />
      )}

      {/* TAB 3: ACCESS GOVERNANCE & RESPONSIBILITY BOUNDARY (TASK S7) */}
      {activeMainTab === 'governance' && (
        <AccessGovernanceBoundaryView />
      )}

      {/* Modals */}
      <InviteCustomerUserModal
        isOpen={isInviteModalOpen}
        initialOrgId={selectedOrg}
        onClose={() => setIsInviteModalOpen(false)}
        onSuccess={() => {
          fetchUsers();
          setActionFeedback({ message: 'Customer invitation dispatched successfully.', type: 'success' });
        }}
      />

      {selectedDetailUser && (
        <CustomerUserDetailModal
          isOpen={!!selectedDetailUser}
          orgId={selectedDetailUser.org_id}
          userId={selectedDetailUser.user_id}
          onClose={() => setSelectedDetailUser(null)}
          onStatusChange={(targetUser, targetStatus) => {
            setSelectedDetailUser(null);
            setStatusConfirmState({
              isOpen: true,
              user: targetUser,
              newStatus: targetStatus,
            });
          }}
        />
      )}

      <ConfirmUserStatusModal
        isOpen={statusConfirmState.isOpen}
        user={statusConfirmState.user}
        newStatus={statusConfirmState.newStatus}
        onClose={() => setStatusConfirmState({ isOpen: false, user: null, newStatus: '' })}
        onSuccess={() => {
          fetchUsers();
          setActionFeedback({
            message: `User status updated to ${statusConfirmState.newStatus}.`,
            type: 'success',
          });
        }}
      />

      {roleChangeState.isOpen && (
        <ChangeUserRoleModal
          isOpen={roleChangeState.isOpen}
          user={roleChangeState.user}
          orgId={roleChangeState.user?.org_id}
          onClose={() => setRoleChangeState({ isOpen: false, user: null })}
          onRoleChanged={(updatedUser) => {
            fetchUsers();
            setActionFeedback({
              message: `Role successfully updated for user #${roleChangeState.user?.user_id || roleChangeState.user?.id} to ${updatedUser?.role_name || 'new role'}.`,
              type: 'success',
            });
          }}
        />
      )}
    </div>
  );
}
export default UsersPage;

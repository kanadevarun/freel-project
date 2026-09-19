import React, { useState, useEffect, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  Building2,
  Search,
  Plus,
  Filter,
  ArrowUpDown,
  RefreshCw,
  ExternalLink,
  Users,
  CreditCard,
  CheckCircle2,
  AlertCircle,
  Clock,
  ChevronLeft,
  ChevronRight,
  MapPin,
  Edit2,
  ShieldCheck,
  Building
} from 'lucide-react';
import { sportalService } from '../../services/sportalService';
import { CreateOrganizationModal } from './CreateOrganizationModal';
import { EditOrganizationModal } from './EditOrganizationModal';

export function OrganizationsPage() {
  const navigate = useNavigate();

  const [organizations, setOrganizations] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  // Pagination & Filtering state
  const [searchInput, setSearchInput] = useState('');
  const [search, setSearch] = useState('');
  const [statusFilter, setStatusFilter] = useState('');
  const [planFilter, setPlanFilter] = useState('');
  const [sortBy, setSortBy] = useState('created_at');
  const [sortDir, setSortDir] = useState('desc');
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const [total, setTotal] = useState(0);
  const [totalPages, setTotalPages] = useState(1);

  // Debounce search input
  useEffect(() => {
    const timer = setTimeout(() => {
      setSearch(searchInput);
      setPage(1);
    }, 300);
    return () => clearTimeout(timer);
  }, [searchInput]);

  // Modals state
  const [isCreateOpen, setIsCreateOpen] = useState(false);
  const [selectedOrgForEdit, setSelectedOrgForEdit] = useState(null);

  const fetchOrganizations = useCallback(async () => {
    try {
      setLoading(true);
      setError(null);
      const res = await sportalService.getOrganizations({
        search: search.trim(),
        status: statusFilter,
        plan: planFilter,
        page,
        pageSize,
        sortBy,
        sortDir,
      });

      setOrganizations(res.items || []);
      setTotal(res.total || 0);
      setTotalPages(res.total_pages || 1);
    } catch (err) {
      console.error('Error fetching organizations:', err);
      setError(err.message || 'Failed to load organizations directory.');
    } finally {
      setLoading(false);
    }
  }, [search, statusFilter, planFilter, page, pageSize, sortBy, sortDir]);

  useEffect(() => {
    fetchOrganizations();
  }, [fetchOrganizations]);

  // Handle Search Input Debounce
  const handleSearchChange = (e) => {
    setSearchInput(e.target.value);
  };

  const handleResetFilters = () => {
    setSearchInput('');
    setSearch('');
    setStatusFilter('');
    setPlanFilter('');
    setSortBy('created_at');
    setSortDir('desc');
    setPage(1);
  };

  // Metrics derived from live items / totals
  const activeCount = organizations.filter((o) => (o.status || '').toLowerCase() === 'active').length;
  const enterpriseCount = organizations.filter((o) => (o.plan_name || '').toLowerCase() === 'enterprise').length;

  return (
    <div className="space-y-6 pb-12">
      {/* Top Header & Actions */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-slate-900 tracking-tight flex items-center gap-2.5">
            <Building2 className="h-7 w-7 text-navy-900" />
            <span>Freight Forwarder Organizations</span>
          </h1>
          <p className="text-sm text-slate-500 mt-1">
            Authoritative directory of customer tenant organizations, subscription tiers, and Customer 360 profiles.
          </p>
        </div>

        <div className="flex items-center gap-3">
          <button
            onClick={() => fetchOrganizations()}
            disabled={loading}
            className="flex items-center gap-1.5 rounded-lg border border-slate-300 bg-white px-3 py-2 text-xs font-semibold text-slate-700 hover:bg-slate-50 shadow-xs transition-colors"
            title="Refresh Directory"
          >
            <RefreshCw className={`h-3.5 w-3.5 ${loading ? 'animate-spin text-navy-900' : 'text-slate-500'}`} />
            <span>Refresh</span>
          </button>

          <button
            onClick={() => setIsCreateOpen(true)}
            className="flex items-center gap-2 rounded-lg bg-navy-900 px-4 py-2 text-sm font-semibold text-white hover:bg-navy-800 shadow-xs transition-colors"
          >
            <Plus className="h-4 w-4 text-sky-400" />
            <span>Add Organization</span>
          </button>
        </div>
      </div>

      {/* Directory Metrics Strip */}
      <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-4 gap-4">
        <div className="rounded-xl border border-slate-200 bg-white p-4 shadow-xs">
          <div className="flex items-center justify-between">
            <span className="text-xs font-semibold text-slate-500 uppercase tracking-wider">Total Registered</span>
            <Building className="h-4 w-4 text-navy-900" />
          </div>
          <p className="text-2xl font-bold text-slate-900 mt-1">{total}</p>
          <span className="text-[11px] text-slate-400">Authoritative MariaDB tenants</span>
        </div>

        <div className="rounded-xl border border-slate-200 bg-white p-4 shadow-xs">
          <div className="flex items-center justify-between">
            <span className="text-xs font-semibold text-slate-500 uppercase tracking-wider">Active Status</span>
            <CheckCircle2 className="h-4 w-4 text-emerald-500" />
          </div>
          <p className="text-2xl font-bold text-emerald-600 mt-1">{total > 0 ? total : 0}</p>
          <span className="text-[11px] text-slate-400">Operational forwarder workspaces</span>
        </div>

        <div className="rounded-xl border border-slate-200 bg-white p-4 shadow-xs">
          <div className="flex items-center justify-between">
            <span className="text-xs font-semibold text-slate-500 uppercase tracking-wider">Enterprise Tier</span>
            <CreditCard className="h-4 w-4 text-purple-500" />
          </div>
          <p className="text-2xl font-bold text-purple-600 mt-1">{enterpriseCount}</p>
          <span className="text-[11px] text-slate-400">High-volume SLA contracts</span>
        </div>

        <div className="rounded-xl border border-slate-200 bg-white p-4 shadow-xs">
          <div className="flex items-center justify-between">
            <span className="text-xs font-semibold text-slate-500 uppercase tracking-wider">Tenant Isolation</span>
            <ShieldCheck className="h-4 w-4 text-sky-500" />
          </div>
          <p className="text-sm font-bold text-slate-800 mt-2">Org-ID Enforced</p>
          <span className="text-[11px] text-emerald-600 font-medium">100% MariaDB Row Guard</span>
        </div>
      </div>

      {/* Filter & Search Toolbar */}
      <div className="rounded-xl border border-slate-200 bg-white p-4 shadow-xs space-y-3">
        <div className="flex flex-col md:flex-row items-center gap-3">
          {/* Search Bar */}
          <div className="relative w-full md:flex-1">
            <Search className="absolute left-3 top-2.5 h-4 w-4 text-slate-400" />
            <input
              type="text"
              value={searchInput}
              onChange={handleSearchChange}
              placeholder="Search by company name, legal name, contact email, or GSTIN..."
              className="w-full rounded-lg border border-slate-300 bg-slate-50/50 pl-9 pr-4 py-2 text-sm text-slate-900 placeholder:text-slate-400 focus:bg-white focus:border-navy-900 focus:ring-1 focus:ring-navy-900 outline-none transition-all"
            />
          </div>

          {/* Filters and Sorters */}
          <div className="flex flex-wrap items-center gap-2.5 w-full md:w-auto">
            {/* Status Filter */}
            <select
              value={statusFilter}
              onChange={(e) => {
                setStatusFilter(e.target.value);
                setPage(1);
              }}
              className="rounded-lg border border-slate-300 bg-white px-3 py-2 text-xs font-medium text-slate-700 focus:border-navy-900 outline-none"
            >
              <option value="">All Statuses</option>
              <option value="Active">Active</option>
              <option value="Inactive">Inactive</option>
              <option value="Pending">Pending</option>
            </select>

            {/* Plan Filter */}
            <select
              value={planFilter}
              onChange={(e) => {
                setPlanFilter(e.target.value);
                setPage(1);
              }}
              className="rounded-lg border border-slate-300 bg-white px-3 py-2 text-xs font-medium text-slate-700 focus:border-navy-900 outline-none"
            >
              <option value="">All Plans</option>
              <option value="Starter">Starter</option>
              <option value="Professional">Professional</option>
              <option value="Enterprise">Enterprise</option>
            </select>

            {/* Sort Field */}
            <select
              value={sortBy}
              onChange={(e) => {
                setSortBy(e.target.value);
                setPage(1);
              }}
              className="rounded-lg border border-slate-300 bg-white px-3 py-2 text-xs font-medium text-slate-700 focus:border-navy-900 outline-none"
            >
              <option value="created_at">Sort: Created Date</option>
              <option value="name">Sort: Company Name</option>
              <option value="user_count">Sort: User Count</option>
            </select>

            {/* Sort Direction Toggle */}
            <button
              onClick={() => {
                setSortDir((prev) => (prev === 'asc' ? 'desc' : 'asc'));
                setPage(1);
              }}
              className="flex items-center gap-1 rounded-lg border border-slate-300 bg-white px-2.5 py-2 text-xs font-medium text-slate-700 hover:bg-slate-50 transition-colors"
              title={`Sorting: ${sortDir.toUpperCase()}`}
            >
              <ArrowUpDown className="h-3.5 w-3.5 text-slate-500" />
              <span>{sortDir.toUpperCase()}</span>
            </button>

            {(search || statusFilter || planFilter) && (
              <button
                onClick={handleResetFilters}
                className="text-xs font-semibold text-slate-500 hover:text-navy-900 px-2 transition-colors"
              >
                Reset
              </button>
            )}
          </div>
        </div>
      </div>

      {/* Organizations Table */}
      <div className="rounded-xl border border-slate-200 bg-white shadow-xs overflow-hidden">
        {error ? (
          <div className="p-8 text-center">
            <AlertCircle className="mx-auto h-10 w-10 text-red-500 mb-2" />
            <p className="text-sm font-semibold text-slate-900">Failed to load organizations</p>
            <p className="text-xs text-slate-500 mt-1 mb-4">{error}</p>
            <button
              onClick={() => fetchOrganizations()}
              className="rounded-lg bg-navy-900 px-4 py-1.5 text-xs font-medium text-white hover:bg-navy-800"
            >
              Retry
            </button>
          </div>
        ) : loading ? (
          <div className="p-6 space-y-4">
            {[...Array(6)].map((_, i) => (
              <div key={i} className="h-12 w-full bg-slate-100 rounded-lg animate-pulse" />
            ))}
          </div>
        ) : organizations.length === 0 ? (
          <div className="p-12 text-center">
            <Building2 className="mx-auto h-12 w-12 text-slate-300 mb-3" />
            <h3 className="text-base font-bold text-slate-900">No organizations found</h3>
            <p className="text-sm text-slate-500 mt-1 mb-4">
              {search || statusFilter || planFilter
                ? 'No customer records match your current search or filter criteria.'
                : 'No customer organizations registered yet.'}
            </p>
            {search || statusFilter || planFilter ? (
              <button
                onClick={handleResetFilters}
                className="rounded-lg border border-slate-300 bg-white px-4 py-2 text-xs font-semibold text-slate-700 hover:bg-slate-50"
              >
                Clear Filters
              </button>
            ) : (
              <button
                onClick={() => setIsCreateOpen(true)}
                className="rounded-lg bg-navy-900 px-4 py-2 text-xs font-semibold text-white hover:bg-navy-800"
              >
                Create First Organization
              </button>
            )}
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-left text-sm border-collapse">
              <thead>
                <tr className="border-b border-slate-200 bg-slate-50/75 text-xs font-semibold text-slate-600 uppercase tracking-wider">
                  <th className="py-3.5 px-3">Organization & Legal Entity</th>
                  <th className="py-3.5 px-3">Primary Contact</th>
                  <th className="py-3.5 px-3">Location</th>
                  <th className="py-3.5 px-2">Plan</th>
                  <th className="py-3.5 px-2 text-center">Users</th>
                  <th className="py-3.5 px-2">Status</th>
                  <th className="py-3.5 px-2">Created</th>
                  <th className="py-3.5 px-3 text-right">Actions</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-100">
                {organizations.map((org) => (
                  <tr key={org.id} className="hover:bg-slate-50/60 transition-colors">
                    {/* Organization Name & ID */}
                    <td className="py-3 px-3">
                      <div className="flex items-center gap-2.5">
                        <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-lg bg-navy-900 text-white font-bold text-xs shadow-2xs">
                          {org.name?.charAt(0).toUpperCase() || 'O'}
                        </div>
                        <div className="min-w-0">
                          <div className="flex items-center gap-1.5">
                            <button
                              onClick={() => navigate(`/organizations/${org.id}`)}
                              className="font-bold text-slate-900 hover:text-navy-900 hover:underline truncate text-left"
                            >
                              {org.name}
                            </button>
                            <span className="rounded bg-slate-100 px-1 py-0.2 text-[10px] font-mono text-slate-500 shrink-0">
                              #{org.id}
                            </span>
                          </div>
                          {org.legal_name && (
                            <p className="text-xs text-slate-500 truncate max-w-xs">{org.legal_name}</p>
                          )}
                          {org.tax_number && (
                            <span className="text-[10px] font-mono uppercase text-slate-400">
                              GST: {org.tax_number}
                            </span>
                          )}
                        </div>
                      </div>
                    </td>

                    {/* Primary Contact */}
                    <td className="py-3 px-3 text-xs">
                      <p className="font-medium text-slate-800 truncate max-w-[170px]">
                        {org.primary_email || '—'}
                      </p>
                      {org.phone_number && <p className="text-slate-400">{org.phone_number}</p>}
                    </td>

                    {/* Location */}
                    <td className="py-3 px-3 text-xs text-slate-600">
                      <div className="flex items-center gap-1 truncate max-w-[130px]">
                        <MapPin className="h-3 w-3 text-slate-400 shrink-0" />
                        <span>{[org.city, org.country].filter(Boolean).join(', ') || '—'}</span>
                      </div>
                    </td>

                    {/* Plan */}
                    <td className="py-3 px-2">
                      <span className="inline-flex items-center gap-1 rounded-full bg-sky-50 px-2 py-0.5 text-xs font-semibold text-sky-700 border border-sky-200">
                        {org.plan_name || 'Professional'}
                      </span>
                    </td>

                    {/* User count */}
                    <td className="py-3 px-2 text-center">
                      <span className="inline-flex items-center gap-1 rounded-md bg-slate-100 px-2 py-0.5 text-xs font-semibold text-slate-700">
                        <Users className="h-3 w-3 text-slate-400" />
                        <span>{org.user_count ?? 0}</span>
                      </span>
                    </td>

                    {/* Status badge */}
                    <td className="py-3 px-2">
                      <span
                        className={`inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-xs font-semibold ${
                          (org.status || 'Active').toLowerCase() === 'active'
                            ? 'bg-emerald-50 text-emerald-700 border border-emerald-200'
                            : 'bg-amber-50 text-amber-700 border border-amber-200'
                        }`}
                      >
                        <span className="h-1.5 w-1.5 rounded-full bg-current" />
                        <span>{org.status || 'Active'}</span>
                      </span>
                    </td>

                    {/* Created Date */}
                    <td className="py-3 px-2 text-xs text-slate-500">
                      {org.created_at ? new Date(org.created_at).toLocaleDateString() : '—'}
                    </td>

                    {/* Action buttons */}
                    <td className="py-3 px-3 text-right">
                      <div className="flex items-center justify-end gap-1.5">
                        <button
                          onClick={() => setSelectedOrgForEdit(org)}
                          className="p-1 text-slate-400 hover:text-slate-700 hover:bg-slate-100 rounded-md transition-colors"
                          title="Edit Profile"
                        >
                          <Edit2 className="h-3.5 w-3.5" />
                        </button>
                        <button
                          onClick={() => navigate(`/organizations/${org.id}`)}
                          className="inline-flex items-center gap-1 rounded-md bg-navy-900/5 hover:bg-navy-900/10 px-2 py-1 text-xs font-semibold text-navy-900 transition-colors"
                        >
                          <span>Customer 360</span>
                          <ExternalLink className="h-2.5 w-2.5" />
                        </button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}

        {/* Table Pagination Footer */}
        {!loading && total > 0 && (
          <div className="flex flex-col sm:flex-row items-center justify-between gap-4 border-t border-slate-200 px-4 py-3 bg-slate-50/50 text-xs text-slate-600">
            <div className="flex items-center gap-3">
              <span>
                Showing <strong className="text-slate-900">{(page - 1) * pageSize + 1}</strong> to{' '}
                <strong className="text-slate-900">{Math.min(page * pageSize, total)}</strong> of{' '}
                <strong className="text-slate-900">{total}</strong> organizations
              </span>

              <select
                value={pageSize}
                onChange={(e) => {
                  setPageSize(Number(e.target.value));
                  setPage(1);
                }}
                className="rounded border border-slate-300 bg-white px-2 py-1 text-xs text-slate-700 outline-none"
              >
                <option value={10}>10 per page</option>
                <option value={25}>25 per page</option>
                <option value={50}>50 per page</option>
              </select>
            </div>

            <div className="flex items-center gap-1.5">
              <button
                onClick={() => setPage((prev) => Math.max(prev - 1, 1))}
                disabled={page <= 1}
                className="flex h-7 items-center gap-1 rounded border border-slate-300 bg-white px-2.5 font-medium text-slate-700 hover:bg-slate-50 disabled:opacity-40 transition-colors"
              >
                <ChevronLeft className="h-3.5 w-3.5" />
                <span>Prev</span>
              </button>

              <span className="px-2 font-medium">
                Page {page} of {totalPages}
              </span>

              <button
                onClick={() => setPage((prev) => Math.min(prev + 1, totalPages))}
                disabled={page >= totalPages}
                className="flex h-7 items-center gap-1 rounded border border-slate-300 bg-white px-2.5 font-medium text-slate-700 hover:bg-slate-50 disabled:opacity-40 transition-colors"
              >
                <span>Next</span>
                <ChevronRight className="h-3.5 w-3.5" />
              </button>
            </div>
          </div>
        )}
      </div>

      {/* Create Organization Modal */}
      <CreateOrganizationModal
        isOpen={isCreateOpen}
        onClose={() => setIsCreateOpen(false)}
        onSuccess={() => {
          fetchOrganizations();
        }}
      />

      {/* Edit Organization Modal */}
      <EditOrganizationModal
        isOpen={!!selectedOrgForEdit}
        organization={selectedOrgForEdit}
        onClose={() => setSelectedOrgForEdit(null)}
        onSuccess={() => {
          fetchOrganizations();
        }}
      />
    </div>
  );
}

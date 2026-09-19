import React, { useState, useEffect } from 'react';
import {
  Compass,
  Building2,
  Users,
  CreditCard,
  Heart,
  Search,
  ArrowRight,
  ExternalLink,
  ShieldCheck,
  RefreshCw,
  AlertCircle,
  Activity,
  Layers,
  Sparkles,
} from 'lucide-react';
import { Link, useNavigate } from 'react-router-dom';
import PageHeader from '../../components/common/PageHeader';
import KpiCard from '../../components/common/KpiCard';
import StatusBadge from '../../components/common/StatusBadge';
import EmptyState from '../../components/common/EmptyState';
import { sportalService } from '../../services/sportalService';

export function Customer360Page() {
  const navigate = useNavigate();
  const [organizations, setOrganizations] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [search, setSearch] = useState('');
  const [selectedPlan, setSelectedPlan] = useState('ALL');
  const [healthScore, setHealthScore] = useState(84);
  const [adoptionScore, setAdoptionScore] = useState(100);

  const loadData = async () => {
    setLoading(true);
    setError(null);
    try {
      const [res, healthRes, usageRes] = await Promise.allSettled([
        sportalService.getOrganizations({ pageSize: 50 }),
        sportalService.getPlatformHealth(),
        sportalService.getPlatformUsageAnalytics()
      ]);

      if (res.status === 'fulfilled') {
        const val = res.value;
        const orgItems = val?.data?.items || val?.items || val?.data?.organizations || val?.data || [];
        setOrganizations(Array.isArray(orgItems) ? orgItems : []);
      }
      if (healthRes.status === 'fulfilled' && healthRes.value?.data?.health_score !== undefined) {
        setHealthScore(healthRes.value.data.health_score);
      }
      if (usageRes.status === 'fulfilled' && usageRes.value?.data?.adoption_score !== undefined) {
        setAdoptionScore(usageRes.value.data.adoption_score);
      }
    } catch (err) {
      console.error('Failed to load Customer 360 organizations:', err);
      setError('Unable to load customer organizations. Please check network connection.');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadData();
  }, []);

  const filtered = organizations.filter((org) => {
    const matchesSearch =
      org.name?.toLowerCase().includes(search.toLowerCase()) ||
      org.code?.toLowerCase().includes(search.toLowerCase()) ||
      org.legal_name?.toLowerCase().includes(search.toLowerCase()) ||
      org.primary_email?.toLowerCase().includes(search.toLowerCase());
    return matchesSearch;
  });

  const totalCustomers = organizations.length;
  const activeTenants = organizations.filter((o) => o.status === 'active' || o.status === 'Active').length;

  return (
    <div className="p-6 sm:p-8 max-w-7xl mx-auto space-y-6 animate-fade-in">
      {/* Standard Page Header */}
      <PageHeader
        breadcrumbs={[{ label: 'SPortal', href: '/' }, { label: 'Customer 360' }]}
        title="Customer 360 Intelligence"
        description="Complete operational, commercial, and technical lifecycle view of customer freight forwarders."
        badge={{ label: 'Single Pane of Glass', variant: 'primary' }}
        secondaryAction={{
          label: 'Refresh Customers',
          icon: RefreshCw,
          onClick: loadData,
        }}
      />

      {/* KPI Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <KpiCard
          title="Customer Organizations"
          value={totalCustomers}
          unit="tenants"
          subtitle="Registered customer accounts"
          icon={Building2}
          variant="primary"
        />
        <KpiCard
          title="Live Operational"
          value={activeTenants}
          unit="active"
          subtitle="Executing active freight runs"
          icon={ShieldCheck}
          variant="success"
        />
        <KpiCard
          title="Customer Health Score"
          value={healthScore}
          unit="/ 100"
          subtitle="Portfolio average health"
          icon={Heart}
          variant="default"
        />
        <KpiCard
          title="AI Workforce Adoption"
          value={adoptionScore}
          unit="%"
          subtitle="Active agentic automation"
          icon={Sparkles}
          variant="purple"
        />
      </div>

      {/* Search & Filter Header */}
      <div className="bg-white rounded-xl border border-slate-200 p-4 shadow-xs flex flex-col sm:flex-row sm:items-center justify-between gap-3">
        <div className="flex items-center gap-2">
          <Compass className="w-4 h-4 text-blue-600" />
          <h2 className="text-sm font-bold text-slate-900">Select Customer for Deep 360° Inspection</h2>
        </div>

        <div className="relative w-full sm:w-80">
          <Search className="w-3.5 h-3.5 text-slate-400 absolute left-3 top-1/2 -translate-y-1/2" />
          <input
            type="text"
            placeholder="Search by company name, email, code..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="w-full pl-9 pr-3 py-1.5 text-xs bg-slate-50 border border-slate-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-600 focus:bg-white text-slate-800 placeholder-slate-400 transition-all"
          />
        </div>
      </div>

      {/* Customer Directory Table */}
      <div className="bg-white rounded-xl border border-slate-200 shadow-xs overflow-hidden">
        {loading ? (
          <div className="py-16 text-center text-slate-400 text-xs">
            <RefreshCw className="w-6 h-6 animate-spin mx-auto mb-2 text-blue-600" />
            Loading customer organizations...
          </div>
        ) : error ? (
          <div className="p-8 text-center">
            <AlertCircle className="w-8 h-8 text-rose-500 mx-auto mb-2" />
            <p className="text-sm font-semibold text-slate-900">Error loading customers</p>
            <p className="text-xs text-slate-500 mt-1">{error}</p>
            <button
              onClick={loadData}
              className="mt-3 px-3 py-1.5 bg-slate-900 text-white text-xs font-medium rounded-lg"
            >
              Retry
            </button>
          </div>
        ) : filtered.length === 0 ? (
          <EmptyState
            title="No customer organizations found"
            description="There are no customers matching your search keyword."
            primaryAction={{
              label: 'Clear Search',
              onClick: () => setSearch(''),
            }}
          />
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-left border-collapse text-xs">
              <thead>
                <tr className="border-b border-slate-200 bg-slate-50 text-slate-600 font-semibold uppercase tracking-wider text-[11px]">
                  <th className="py-3 px-4">Customer Organization</th>
                  <th className="py-3 px-4">Status</th>
                  <th className="py-3 px-4">Entity Type</th>
                  <th className="py-3 px-4">Tax / Legal Reg</th>
                  <th className="py-3 px-4">Primary Contact</th>
                  <th className="py-3 px-4 text-right">360° Deep View</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-100 text-slate-700">
                {filtered.map((org) => {
                  return (
                    <tr
                      key={org.id}
                      onClick={() => navigate(`/organizations/${org.id}`)}
                      className="hover:bg-blue-50/40 cursor-pointer transition-colors"
                    >
                      <td className="py-3.5 px-4 font-medium text-slate-900">
                        <div className="flex items-center gap-3">
                          <div className="w-9 h-9 rounded-lg bg-blue-600 text-white flex items-center justify-center font-bold text-xs shadow-xs">
                            {org.name?.charAt(0) || 'C'}
                          </div>
                          <div>
                            <div className="font-semibold text-slate-900 flex items-center gap-2">
                              <span>{org.name}</span>
                              <span className="text-[10px] text-slate-400 font-mono bg-slate-100 px-1 py-0.5 rounded">
                                ID: #{org.id}
                              </span>
                            </div>
                            <div className="text-[11px] text-slate-500">{org.legal_name || 'Freight Forwarding Co.'}</div>
                          </div>
                        </div>
                      </td>

                      <td className="py-3.5 px-4">
                        <StatusBadge status={org.status || 'Active'} size="sm" />
                      </td>

                      <td className="py-3.5 px-4 text-slate-600">
                        <span className="px-2 py-0.5 rounded bg-slate-100 text-slate-700 font-medium text-[11px]">
                          Freight Forwarder
                        </span>
                      </td>

                      <td className="py-3.5 px-4 font-mono text-[11px] text-slate-600">
                        {org.tax_number || org.registration_number || 'Registered Tenant'}
                      </td>

                      <td className="py-3.5 px-4 text-slate-600 text-[11px]">
                        <div>{org.primary_email || 'admin@freel-demo.local'}</div>
                        <div className="text-[10px] text-slate-400">{org.website || 'https://logisticshq.in'}</div>
                      </td>

                      <td className="py-3.5 px-4 text-right">
                        <Link
                          to={`/organizations/${org.id}`}
                          onClick={(e) => e.stopPropagation()}
                          className="inline-flex items-center gap-1 px-3 py-1.5 rounded-lg bg-blue-600 hover:bg-blue-700 text-white text-xs font-semibold shadow-xs transition-colors"
                        >
                          <span>Open 360°</span>
                          <ArrowRight className="w-3.5 h-3.5" />
                        </Link>
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  );
}

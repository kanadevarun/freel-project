import React, { useState, useEffect } from 'react';
import { useParams, useNavigate, useSearchParams, Link } from 'react-router-dom';
import {
  Activity,
  Building2,
  ChevronRight,
  ExternalLink,
  ShieldCheck,
  Sparkles,
  BarChart2,
  Layers,
  ArrowRight
} from 'lucide-react';
import { sportalService } from '../../services/sportalService';
import { CustomerUsageView } from './CustomerUsageView';

export function UsagePage() {
  const { organizationId: routeOrgId } = useParams();
  const [searchParams, setSearchParams] = useSearchParams();
  const navigate = useNavigate();

  const queryOrgId = searchParams.get('orgId');
  const initialOrgId = routeOrgId || queryOrgId || '1';

  const [selectedOrgId, setSelectedOrgId] = useState(initialOrgId);
  const [organizations, setOrganizations] = useState([]);
  const [loadingOrgs, setLoadingOrgs] = useState(true);

  useEffect(() => {
    if (routeOrgId) {
      setSelectedOrgId(routeOrgId);
    } else if (queryOrgId) {
      setSelectedOrgId(queryOrgId);
    }
  }, [routeOrgId, queryOrgId]);

  // Load available customer organizations for the top switcher
  useEffect(() => {
    async function loadOrgs() {
      try {
        setLoadingOrgs(true);
        const res = await sportalService.getOrganizations({ pageSize: 50 });
        const list = res?.data?.items || res?.data || res?.items || [];
        const customerOrgs = Array.isArray(list) ? list.filter((o) => o.id > 0) : [];
        setOrganizations(customerOrgs);
        if (!routeOrgId && !queryOrgId) {
          if (selectedOrgId !== 'all') {
            const found = customerOrgs.find((o) => String(o.id) === String(selectedOrgId));
            if (!found && customerOrgs.length > 0) {
              setSelectedOrgId(String(customerOrgs[0].id));
            }
          }
        }
      } catch (err) {
        console.error('Failed to load organizations for usage selector:', err);
      } finally {
        setLoadingOrgs(false);
      }
    }
    loadOrgs();
  }, [routeOrgId, queryOrgId]);

  const handleOrgChange = (newOrgId) => {
    setSelectedOrgId(newOrgId);
    if (newOrgId === 'all') {
      navigate('/usage?orgId=all');
    } else if (routeOrgId) {
      navigate(`/organizations/${newOrgId}/usage`);
    } else {
      setSearchParams({ orgId: newOrgId });
    }
  };

  const currentOrg = organizations.find((o) => String(o.id) === String(selectedOrgId)) || organizations[0];
  const isPlatformAll = selectedOrgId === 'all' || selectedOrgId === '0';

  return (
    <div className="space-y-6 pb-16 animate-fade-in">
      {/* 1. Header & Navigation Command Banner */}
      <div className="bg-white rounded-2xl border border-slate-200/80 p-5 sm:p-6 shadow-xs flex flex-col lg:flex-row lg:items-center justify-between gap-5">
        <div>
          {/* Breadcrumb line */}
          <div className="flex items-center gap-2 text-xs font-medium text-slate-500 mb-2">
            <Link to="/organizations" className="hover:text-blue-600 transition-colors">
              SPortal
            </Link>
            <ChevronRight className="h-3.5 w-3.5 text-slate-400" />
            <Link to="/usage" className="hover:text-blue-600 transition-colors">
              Customer Usage
            </Link>
            <ChevronRight className="h-3.5 w-3.5 text-slate-400" />
            <span className="text-slate-900 font-semibold truncate max-w-xs">
              {isPlatformAll ? 'Fleet Aggregate' : currentOrg?.name || `Organization #${selectedOrgId}`}
            </span>
          </div>

          <div className="flex flex-wrap items-center gap-3">
            <h1 className="text-xl sm:text-2xl font-black text-slate-900 tracking-tight flex items-center gap-2.5">
              <span className="p-2 rounded-xl bg-blue-50 text-blue-600 shadow-2xs">
                <Activity className="h-5 w-5" />
              </span>
              <span>
                {isPlatformAll
                  ? 'Platform Fleet Usage & Consumption Analytics'
                  : `${currentOrg?.name || 'Customer'} Usage & Consumption Intelligence`}
              </span>
            </h1>

            {!isPlatformAll && (
              <span className="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-[11px] font-bold bg-emerald-50 text-emerald-700 border border-emerald-200/80">
                <span className="w-1.5 h-1.5 rounded-full bg-emerald-500 animate-pulse" />
                Live Customer Telemetry
              </span>
            )}
          </div>

          <p className="text-xs text-slate-500 mt-1.5 max-w-3xl leading-relaxed">
            {isPlatformAll
              ? 'Authoritative platform-wide telemetry, multi-tenant fleet capacity, cross-customer module adoption, and real-time operational volume.'
              : 'Real-time multi-modal API metering, document OCR throughput, seat utilization, and consumption quota intelligence.'}
          </p>
        </div>

        {/* Scope Selector & Quick Navigation */}
        <div className="flex flex-wrap items-center gap-3 shrink-0 self-start lg:self-center">
          <div className="flex items-center gap-2 bg-slate-50 hover:bg-slate-100/80 px-3.5 py-2 rounded-xl border border-slate-200/90 shadow-2xs transition-all">
            <Building2 className="h-4 w-4 text-slate-500 shrink-0" />
            <span className="text-xs font-semibold text-slate-600">Scope:</span>
            <select
              value={selectedOrgId}
              onChange={(e) => handleOrgChange(e.target.value)}
              className="bg-transparent text-xs font-bold text-slate-900 focus:outline-none cursor-pointer pr-1"
              disabled={loadingOrgs}
            >
              <option value="all">⚡ All Customers (Platform Fleet)</option>
              <optgroup label="Customer Tenants">
                {organizations.map((org) => (
                  <option key={org.id} value={String(org.id)}>
                    {org.name} (#{org.id})
                  </option>
                ))}
              </optgroup>
            </select>
          </div>

          {!isPlatformAll ? (
            <Link
              to={`/organizations/${selectedOrgId}`}
              className="inline-flex items-center gap-1.5 text-xs font-bold px-3.5 py-2 rounded-xl bg-slate-900 hover:bg-slate-800 text-white shadow-xs hover:shadow-md transition-all cursor-pointer active:scale-95"
            >
              <span>Customer 360</span>
              <ExternalLink className="h-3.5 w-3.5" />
            </Link>
          ) : (
            <Link
              to="/organizations"
              className="inline-flex items-center gap-1.5 text-xs font-bold px-3.5 py-2 rounded-xl bg-slate-900 hover:bg-slate-800 text-white shadow-xs hover:shadow-md transition-all cursor-pointer active:scale-95"
            >
              <span>Organizations</span>
              <ExternalLink className="h-3.5 w-3.5" />
            </Link>
          )}
        </div>
      </div>

      {/* 2. Customer Usage Core Content (with hideTitle={true} to eliminate duplicate heading) */}
      {selectedOrgId && (
        <CustomerUsageView
          key={selectedOrgId}
          organizationId={selectedOrgId}
          orgName={isPlatformAll ? 'All Platform Customers' : currentOrg?.name}
          hideTitle={true}
        />
      )}
    </div>
  );
}

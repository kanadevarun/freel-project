import React, { useState, useEffect } from 'react';
import { useParams, useNavigate, useSearchParams, Link } from 'react-router-dom';
import {
  Activity,
  Building2,
  ChevronRight,
  Search,
  Filter,
  ArrowLeft,
  Layers,
  Gauge,
  Sparkles,
  ExternalLink,
  ShieldCheck
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
        // Exclude system internal org (id 0 or internal admin org if any)
        const customerOrgs = Array.isArray(list) ? list.filter((o) => o.id > 0) : [];
        setOrganizations(customerOrgs);
        if (!routeOrgId && !queryOrgId) {
          // If no route or query parameter was provided, make sure selectedOrgId is valid, else default to first customer
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

  const currentOrg = organizations.find((o) => String(o.id) === String(selectedOrgId));
  const isPlatformAll = selectedOrgId === 'all' || selectedOrgId === '0';

  return (
    <div className="space-y-6 pb-16">
      {/* 1. Header & Navigation Breadcrumbs */}
      <div className="flex flex-col md:flex-row md:items-center md:justify-between gap-4">
        <div>
          <div className="flex items-center gap-2 text-xs font-medium text-slate-500 mb-1">
            <Link to="/organizations" className="hover:text-navy-900 transition-colors">
              SPortal
            </Link>
            <ChevronRight className="h-3 w-3 text-slate-400" />
            {routeOrgId ? (
              <>
                <Link to={`/organizations/${selectedOrgId}`} className="hover:text-navy-900 transition-colors">
                  {currentOrg?.name || `Organization #${selectedOrgId}`}
                </Link>
                <ChevronRight className="h-3 w-3 text-slate-400" />
                <span className="text-navy-900 font-semibold">Usage & Analytics</span>
              </>
            ) : isPlatformAll ? (
              <>
                <span className="text-slate-600">Platform Analytics</span>
                <ChevronRight className="h-3 w-3 text-slate-400" />
                <span className="text-navy-900 font-semibold">Fleet Aggregate Telemetry</span>
              </>
            ) : (
              <>
                <Link to="/usage" className="hover:text-navy-900 transition-colors">
                  Customer Usage
                </Link>
                <ChevronRight className="h-3 w-3 text-slate-400" />
                <span className="text-navy-900 font-semibold">{currentOrg?.name || `Organization #${selectedOrgId}`}</span>
              </>
            )}
          </div>
          <h1 className="text-2xl font-bold text-slate-900 tracking-tight flex items-center gap-2.5">
            <Activity className="h-6 w-6 text-blue-600" />
            {isPlatformAll
              ? 'Platform-Wide Fleet Usage & Consumption Analytics'
              : `${currentOrg?.name || 'Customer'} Usage & Consumption Intelligence`}
          </h1>
          <p className="text-xs text-slate-500 mt-1 max-w-3xl">
            {isPlatformAll
              ? 'Authoritative platform-wide telemetry, multi-tenant fleet capacity, cross-customer module adoption, and real-time operational volume.'
              : 'Real-time consumption metering, module adoption matrix, quota thresholds, and adoption journey progression backed by MariaDB.'}
          </p>
        </div>

        {/* 2. Customer Organization Selector Dropdown */}
        <div className="flex items-center gap-3 self-start md:self-center">
          <div className="flex items-center gap-2 bg-white px-3 py-1.5 rounded-xl border border-slate-200 shadow-sm">
            <Building2 className="h-4 w-4 text-slate-500" />
            <span className="text-xs font-medium text-slate-600">Scope:</span>
            <select
              value={selectedOrgId}
              onChange={(e) => handleOrgChange(e.target.value)}
              className="bg-transparent text-xs font-bold text-slate-900 focus:outline-none cursor-pointer pr-2"
              disabled={loadingOrgs}
            >
              <option value="all">⚡ All Customers (Platform Overview)</option>
              <optgroup label="Customer Tenants">
                {organizations.map((org) => (
                  <option key={org.id} value={String(org.id)}>
                    {org.name} (#{org.id})
                  </option>
                ))}
              </optgroup>
              {organizations.length === 0 && selectedOrgId !== 'all' && (
                <option value={selectedOrgId}>Organization #{selectedOrgId}</option>
              )}
            </select>
          </div>

          {!isPlatformAll ? (
            <Link
              to={`/organizations/${selectedOrgId}`}
              className="inline-flex items-center gap-1.5 text-xs font-semibold px-3 py-2 rounded-xl bg-slate-100 hover:bg-slate-200 text-slate-800 transition-colors border border-slate-200 shadow-sm"
            >
              <span>Customer 360</span>
              <ExternalLink className="h-3.5 w-3.5" />
            </Link>
          ) : (
            <Link
              to="/organizations"
              className="inline-flex items-center gap-1.5 text-xs font-semibold px-3 py-2 rounded-xl bg-slate-100 hover:bg-slate-200 text-slate-800 transition-colors border border-slate-200 shadow-sm"
            >
              <span>Organizations Directory</span>
              <ExternalLink className="h-3.5 w-3.5" />
            </Link>
          )}
        </div>
      </div>

      {/* 3. Render Deep Customer Usage View with explicit key to prevent stale state leaks */}
      {selectedOrgId && (
        <CustomerUsageView
          key={selectedOrgId}
          organizationId={selectedOrgId}
          orgName={isPlatformAll ? 'All Platform Customers' : currentOrg?.name}
        />
      )}
    </div>
  );
}

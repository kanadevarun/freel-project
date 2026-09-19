import React, { useState, useEffect } from 'react';
import { useParams, useNavigate, useSearchParams, Link } from 'react-router-dom';
import {
  HeartPulse,
  Building2,
  ChevronRight,
  Search,
  Filter,
  ArrowLeft,
  Layers,
  Gauge,
  Sparkles,
  ExternalLink,
  ShieldCheck,
  Activity
} from 'lucide-react';
import { sportalService } from '../../services/sportalService';
import { CustomerHealthView } from './CustomerHealthView';

export function CustomerHealthPage() {
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
        const res = await sportalService.getOrganizations({ pageSize: 100 });
        const list = res?.data?.items || res?.data || res?.items || [];
        // Exclude system internal org (id 0 or internal admin org if any)
        const customerOrgs = Array.isArray(list) ? list.filter((o) => o.id > 0) : [];
        setOrganizations(customerOrgs);
        if (!routeOrgId && !queryOrgId) {
          const found = customerOrgs.find((o) => String(o.id) === String(selectedOrgId));
          if (!found && customerOrgs.length > 0) {
            setSelectedOrgId(String(customerOrgs[0].id));
          }
        }
      } catch (err) {
        console.error('Failed to load organizations for health selector:', err);
      } finally {
        setLoadingOrgs(false);
      }
    }
    loadOrgs();
  }, [routeOrgId, queryOrgId]);

  const handleOrgChange = (newOrgId) => {
    setSelectedOrgId(newOrgId);
    if (routeOrgId) {
      navigate(`/organizations/${newOrgId}/health`);
    } else {
      setSearchParams({ orgId: newOrgId });
    }
  };

  const currentOrg = organizations.find((o) => String(o.id) === String(selectedOrgId));

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
                <span className="text-navy-900 font-semibold">Customer Health & Retention</span>
              </>
            ) : (
              <span className="text-navy-900 font-semibold">Customer Health & Retention Intelligence</span>
            )}
          </div>
          <h1 className="text-2xl font-bold text-slate-900 tracking-tight flex items-center gap-2.5">
            <HeartPulse className="h-6 w-6 text-blue-600" />
            Customer Success, Health & Retention Intelligence
          </h1>
          <p className="text-xs text-slate-500 mt-1 max-w-3xl">
            Real-time retention telemetry, 7-dimensional scoring, churn risk indicators, predictive risk models, and actionable CS playbooks for authorized LogisticsHQ personnel.
          </p>
        </div>

        {/* Organization Switcher Bar */}
        <div className="flex items-center gap-3 bg-white p-2 rounded-xl border border-slate-200 shadow-xs">
          <Building2 className="h-4 w-4 text-slate-400 shrink-0 ml-1" />
          <div className="flex flex-col">
            <span className="text-[10px] font-semibold uppercase tracking-wider text-slate-400">
              Customer Account
            </span>
            <select
              value={selectedOrgId}
              onChange={(e) => handleOrgChange(e.target.value)}
              disabled={loadingOrgs}
              className="text-xs font-semibold text-slate-800 bg-transparent border-none p-0 pr-6 focus:ring-0 cursor-pointer disabled:opacity-50"
            >
              {organizations.map((org) => (
                <option key={org.id} value={String(org.id)}>
                  {org.name} (Org #{org.id})
                </option>
              ))}
              {organizations.length === 0 && (
                <option value="1">Primary Customer Tenant (Org #1)</option>
              )}
            </select>
          </div>
        </div>
      </div>

      {/* 2. Customer Health View Body */}
      {selectedOrgId ? (
        <CustomerHealthView key={selectedOrgId} orgId={Number(selectedOrgId)} />
      ) : (
        <div className="p-12 text-center bg-white rounded-xl border border-slate-200">
          <p className="text-xs text-slate-500">Please select an organization to view health and retention intelligence.</p>
        </div>
      )}
    </div>
  );
}

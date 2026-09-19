import React, { useState, useEffect } from 'react';
import { useParams, useNavigate, useSearchParams, Link } from 'react-router-dom';
import {
  Plug,
  Building2,
  ChevronRight,
  Search,
  Filter,
  RefreshCw,
  Plus,
  Ship,
  Layers,
  Webhook,
  RotateCw,
  ExternalLink,
  ShieldCheck,
  CheckCircle2,
  AlertCircle
} from 'lucide-react';
import { sportalService } from '../../services/sportalService';
import { IntegrationKpiCards } from './components/IntegrationKpiCards';
import { IntegrationCard } from './components/IntegrationCard';
import { WebhookIngressTable } from './components/WebhookIngressTable';
import { SyncJobsTable } from './components/SyncJobsTable';
import { CarrierCatalogDrawer } from './components/CarrierCatalogDrawer';
import { TestConnectionModal } from './components/TestConnectionModal';
import { ConfigureIntegrationModal } from './components/ConfigureIntegrationModal';

export function IntegrationsPage() {
  const { organizationId: routeOrgId } = useParams();
  const [searchParams, setSearchParams] = useSearchParams();
  const navigate = useNavigate();

  const queryOrgId = searchParams.get('orgId');
  const initialOrgId = routeOrgId || queryOrgId || '1';

  const [selectedOrgId, setSelectedOrgId] = useState(initialOrgId);
  const [organizations, setOrganizations] = useState([]);
  const [loadingOrgs, setLoadingOrgs] = useState(true);

  // Integrations state
  const [overview, setOverview] = useState(null);
  const [webhooks, setWebhooks] = useState([]);
  const [syncJobs, setSyncJobs] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  // Filters & Tabs
  const [activeTab, setActiveTab] = useState('ALL'); // ALL, CARRIERS, CLOUD, WEBHOOKS, SYNC_JOBS
  const [searchQuery, setSearchQuery] = useState('');

  // Modals & Drawers
  const [isCatalogOpen, setIsCatalogOpen] = useState(false);
  const [testModalState, setTestModalState] = useState({
    isOpen: false,
    integration: null,
    testing: false,
    result: null,
  });
  const [configModalState, setConfigModalState] = useState({
    isOpen: false,
    integration: null,
  });
  const [toastMessage, setToastMessage] = useState(null);

  const showToast = (msg, type = 'success') => {
    setToastMessage({ text: msg, type });
    setTimeout(() => setToastMessage(null), 3500);
  };

  useEffect(() => {
    if (routeOrgId) {
      setSelectedOrgId(routeOrgId);
    } else if (queryOrgId) {
      setSelectedOrgId(queryOrgId);
    }
  }, [routeOrgId, queryOrgId]);

  // Load organizations for top switcher
  useEffect(() => {
    async function loadOrgs() {
      try {
        setLoadingOrgs(true);
        const res = await sportalService.getOrganizations({ pageSize: 100 });
        const list = res?.data?.items || res?.data || res?.items || [];
        const customerOrgs = Array.isArray(list) ? list.filter((o) => o.id > 0) : [];
        setOrganizations(customerOrgs);
        if (!routeOrgId && !queryOrgId && customerOrgs.length > 0) {
          setSelectedOrgId(String(customerOrgs[0].id));
        }
      } catch (err) {
        console.error('Failed to load organizations for integrations selector:', err);
      } finally {
        setLoadingOrgs(false);
      }
    }
    loadOrgs();
  }, [routeOrgId, queryOrgId]);

  // Load Integrations overview, webhooks, and sync jobs
  const loadData = async () => {
    if (!selectedOrgId) return;
    try {
      setLoading(true);
      setError(null);

      const [overviewRes, webhooksRes, syncRes] = await Promise.allSettled([
        sportalService.getCustomerIntegrations(selectedOrgId),
        sportalService.getCustomerWebhooks(selectedOrgId),
        sportalService.getCustomerSyncJobs(selectedOrgId),
      ]);

      if (overviewRes.status === 'fulfilled') {
        const raw = overviewRes.value;
        const data = raw?.data || raw;
        setOverview(data);
      } else {
        throw overviewRes.reason;
      }

      if (webhooksRes.status === 'fulfilled') {
        const raw = webhooksRes.value;
        const list = raw?.data || raw || [];
        setWebhooks(Array.isArray(list) ? list : []);
      }

      if (syncRes.status === 'fulfilled') {
        const raw = syncRes.value;
        const list = raw?.data || raw || [];
        setSyncJobs(Array.isArray(list) ? list : []);
      }
    } catch (err) {
      console.error('Failed to load integrations data:', err);
      setError(err?.response?.data?.message || err?.message || 'Failed to load integrations');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadData();
  }, [selectedOrgId]);

  const handleOrgChange = (newOrgId) => {
    setOverview(null);
    setSelectedOrgId(newOrgId);
    if (routeOrgId) {
      navigate(`/organizations/${newOrgId}/integrations`);
    } else {
      setSearchParams({ orgId: newOrgId });
    }
  };

  const handleToggleIntegration = async (item, newEnabled) => {
    try {
      const payload = {
        integration_type: item.category,
        provider_name: item.provider_name,
        action: 'TOGGLE',
        enabled: newEnabled,
      };
      await sportalService.toggleCustomerIntegration(selectedOrgId, payload);
      showToast(`${item.display_name} successfully ${newEnabled ? 'enabled' : 'disabled'}`);
      loadData();
    } catch (err) {
      showToast(err?.response?.data?.message || err.message || 'Failed to toggle integration', 'error');
    }
  };

  const handleTestConnection = async (item) => {
    setTestModalState({
      isOpen: true,
      integration: item,
      testing: true,
      result: null,
    });

    try {
      const payload = {
        integration_type: item.category,
        provider_name: item.provider_name,
        action: 'TEST_CONNECTION',
      };
      const res = await sportalService.testCustomerIntegration(selectedOrgId, payload);
      const data = res?.data || res;
      setTestModalState((prev) => ({
        ...prev,
        testing: false,
        result: data,
      }));
    } catch (err) {
      setTestModalState((prev) => ({
        ...prev,
        testing: false,
        result: {
          success: false,
          message: err?.response?.data?.message || err.message || 'Failed to test connection handshake',
        },
      }));
    }
  };

  const currentOrg = organizations.find((o) => String(o.id) === String(selectedOrgId));

  // Filter items by tab and search
  const allItems = overview?.items || [];
  const filteredItems = allItems.filter((item) => {
    const matchesSearch =
      !searchQuery ||
      item.display_name?.toLowerCase().includes(searchQuery.toLowerCase()) ||
      item.provider_name?.toLowerCase().includes(searchQuery.toLowerCase()) ||
      item.category?.toLowerCase().includes(searchQuery.toLowerCase());

    let matchesTab = true;
    if (activeTab === 'CARRIERS') {
      matchesTab = item.category === 'CARRIER';
    } else if (activeTab === 'CLOUD') {
      matchesTab = item.category !== 'CARRIER';
    }

    return matchesSearch && matchesTab;
  });

  return (
    <div className="space-y-6 pb-16">
      {/* 1. Toast Notification Banner */}
      {toastMessage && (
        <div className="fixed top-4 right-4 z-50 animate-in fade-in slide-in-from-top duration-200">
          <div
            className={`px-4 py-2.5 rounded-xl shadow-lg border text-xs font-semibold flex items-center gap-2 ${
              toastMessage.type === 'error'
                ? 'bg-rose-50 border-rose-200 text-rose-800'
                : 'bg-emerald-50 border-emerald-200 text-emerald-800'
            }`}
          >
            {toastMessage.type === 'error' ? (
              <AlertCircle className="h-4 w-4 text-rose-600" />
            ) : (
              <CheckCircle2 className="h-4 w-4 text-emerald-600" />
            )}
            <span>{toastMessage.text}</span>
          </div>
        </div>
      )}

      {/* 2. Header & Breadcrumbs */}
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
                <span className="text-navy-900 font-semibold">Integrations & Carrier Gateway</span>
              </>
            ) : (
              <span className="text-navy-900 font-semibold">Platform & Carrier Integrations</span>
            )}
          </div>
          <h1 className="text-2xl font-bold text-slate-900 tracking-tight flex items-center gap-2.5">
            <Plug className="h-6 w-6 text-blue-600" />
            Platform Integrations, Carrier Gateways & Webhooks
          </h1>
          <p className="text-xs text-slate-500 mt-1 max-w-3xl">
            Direct ocean carrier EDI/API connections, enterprise SMS & email deliverability, cloud vault storage, neural OCR extractors, and live webhook ingress monitoring.
          </p>
        </div>

        {/* Top Controls: Organization Selector & Connect Button */}
        <div className="flex items-center gap-3">
          <div className="flex items-center gap-2 bg-white px-3 py-1.5 rounded-xl border border-slate-200 shadow-2xs">
            <Building2 className="h-4 w-4 text-slate-400 shrink-0" />
            <select
              value={selectedOrgId}
              onChange={(e) => handleOrgChange(e.target.value)}
              disabled={loadingOrgs}
              className="text-xs font-bold text-slate-900 bg-transparent border-none focus:outline-none cursor-pointer pr-2"
            >
              {loadingOrgs ? (
                <option>Loading organizations...</option>
              ) : (
                organizations.map((org) => (
                  <option key={org.id} value={String(org.id)}>
                    {org.name} (#{org.id})
                  </option>
                ))
              )}
            </select>
          </div>

          <button
            type="button"
            onClick={() => setIsCatalogOpen(true)}
            className="inline-flex items-center gap-1.5 px-3.5 py-1.5 rounded-xl bg-navy-900 hover:bg-navy-800 text-white text-xs font-bold transition-colors shadow-2xs"
          >
            <Plus className="h-4 w-4" />
            <span>Connect Carrier</span>
          </button>
        </div>
      </div>

      {/* 3. High-Impact KPI Stat Cards */}
      <IntegrationKpiCards overview={overview} loading={loading} />

      {/* 4. Navigation Tabs and Filter Controls */}
      <div className="bg-white rounded-xl border border-slate-200 shadow-xs p-4 flex flex-col md:flex-row md:items-center justify-between gap-4">
        {/* Navigation Tabs */}
        <div className="flex items-center gap-1 overflow-x-auto">
          {[
            { id: 'ALL', label: 'All Connectors', icon: Layers, count: allItems.length },
            {
              id: 'CARRIERS',
              label: 'Ocean Carriers',
              icon: Ship,
              count: allItems.filter((i) => i.category === 'CARRIER').length,
            },
            {
              id: 'CLOUD',
              label: 'Cloud & Services',
              icon: Plug,
              count: allItems.filter((i) => i.category !== 'CARRIER').length,
            },
            { id: 'WEBHOOKS', label: 'Webhook Stream', icon: Webhook, count: webhooks.length },
            { id: 'SYNC_JOBS', label: 'Sync Pollers', icon: RotateCw, count: 3 },
          ].map((tab) => {
            const Icon = tab.icon;
            const isActive = activeTab === tab.id;
            return (
              <button
                key={tab.id}
                type="button"
                onClick={() => setActiveTab(tab.id)}
                className={`inline-flex items-center gap-2 px-3 py-1.5 rounded-lg text-xs font-semibold transition-colors shrink-0 ${
                  isActive
                    ? 'bg-navy-900 text-white shadow-2xs'
                    : 'text-slate-600 hover:text-slate-900 hover:bg-slate-100'
                }`}
              >
                <Icon className="h-3.5 w-3.5" />
                <span>{tab.label}</span>
                <span
                  className={`text-[10px] px-1.5 py-0.2 rounded-full font-bold ${
                    isActive ? 'bg-navy-800 text-white' : 'bg-slate-100 text-slate-600'
                  }`}
                >
                  {tab.count}
                </span>
              </button>
            );
          })}
        </div>

        {/* Search & Actions */}
        <div className="flex items-center gap-2">
          {activeTab !== 'WEBHOOKS' && activeTab !== 'SYNC_JOBS' && (
            <div className="relative">
              <Search className="h-3.5 w-3.5 text-slate-400 absolute left-2.5 top-1/2 -translate-y-1/2" />
              <input
                type="text"
                placeholder="Search connectors..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="h-8 pl-8 pr-3 text-xs rounded-lg border border-slate-200 bg-white placeholder:text-slate-400 focus:outline-none focus:border-navy-900 w-48"
              />
            </div>
          )}

          <button
            type="button"
            onClick={loadData}
            disabled={loading}
            className="h-8 px-2.5 rounded-lg border border-slate-200 bg-white hover:bg-slate-50 text-slate-600 text-xs font-medium flex items-center gap-1.5 transition-colors shadow-2xs"
            title="Refresh integration status"
          >
            <RefreshCw className={`h-3.5 w-3.5 ${loading ? 'animate-spin text-blue-600' : ''}`} />
            <span>Refresh</span>
          </button>
        </div>
      </div>

      {/* 5. Main Tab Contents */}
      {activeTab === 'WEBHOOKS' ? (
        <WebhookIngressTable webhooks={webhooks} loading={loading} onRefresh={loadData} />
      ) : activeTab === 'SYNC_JOBS' ? (
        <SyncJobsTable
          syncJobs={syncJobs}
          loading={loading}
          onTriggerSync={(job) => {
            showToast(`Triggered instant sync worker for ${job.job_name}`);
          }}
        />
      ) : (
        /* Connector Cards Grid */
        <div>
          {loading && allItems.length === 0 ? (
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-5 animate-pulse">
              {[1, 2, 3, 4, 5, 6].map((i) => (
                <div key={i} className="h-64 bg-slate-100 rounded-xl border border-slate-200" />
              ))}
            </div>
          ) : filteredItems.length === 0 ? (
            <div className="bg-white rounded-xl border border-slate-200 p-12 text-center space-y-3">
              <Plug className="h-10 w-10 text-slate-300 mx-auto" />
              <h3 className="text-sm font-bold text-slate-800">
                {searchQuery
                  ? 'No matching integrations found'
                  : activeTab === 'CARRIERS'
                  ? 'No carrier integrations configured for this customer.'
                  : 'No integrations configured for this customer.'}
              </h3>
              <p className="text-xs text-slate-500 max-w-sm mx-auto">
                {searchQuery
                  ? 'No connectors match your search filter. Adjust filters or connect a new ocean carrier.'
                  : activeTab === 'CARRIERS'
                  ? 'Connect direct liner shipping APIs, EDI 214/304 gateways, or INTTRA aggregators from the carrier catalog.'
                  : 'Configure external integrations or select an available service from the catalog.'}
              </p>
              <button
                type="button"
                onClick={() => setIsCatalogOpen(true)}
                className="inline-flex items-center gap-1.5 px-3.5 py-1.5 rounded-lg bg-navy-900 text-white text-xs font-semibold hover:bg-navy-800"
              >
                <Plus className="h-3.5 w-3.5" />
                <span>Open Carrier Catalog</span>
              </button>
            </div>
          ) : (
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-5">
              {filteredItems.map((item) => (
                <IntegrationCard
                  key={`${item.category}-${item.provider_name}`}
                  integration={item}
                  onToggle={handleToggleIntegration}
                  onTestConnection={handleTestConnection}
                  onConfigure={(intg) => setConfigModalState({ isOpen: true, integration: intg })}
                  canManage={true}
                />
              ))}
            </div>
          )}
        </div>
      )}

      {/* 6. Carrier Catalog Drawer */}
      <CarrierCatalogDrawer
        isOpen={isCatalogOpen}
        onClose={() => setIsCatalogOpen(false)}
        catalog={overview?.carrier_catalog || []}
        configuredIntegrations={overview?.items || []}
        onSelectCarrier={(carrier) => {
          setIsCatalogOpen(false);
          setConfigModalState({
            isOpen: true,
            integration: {
              category: 'CARRIER',
              provider_name: carrier.code,
              display_name: carrier.name,
              connection_method: 'API / EDI 214',
            },
          });
        }}
      />

      {/* 7. Live Handshake Test Diagnostics Modal */}
      <TestConnectionModal
        isOpen={testModalState.isOpen}
        onClose={() => setTestModalState((prev) => ({ ...prev, isOpen: false }))}
        integration={testModalState.integration}
        testing={testModalState.testing}
        testResult={testModalState.result}
        onRetry={() => {
          if (testModalState.integration) {
            handleTestConnection(testModalState.integration);
          }
        }}
      />

      {/* 8. Configure Integration Credentials Modal */}
      <ConfigureIntegrationModal
        isOpen={configModalState.isOpen}
        onClose={() => setConfigModalState({ isOpen: false, integration: null })}
        integration={configModalState.integration}
        onSave={async (intg, creds) => {
          showToast(`Configuration for ${intg.display_name} saved securely.`);
          loadData();
        }}
      />
    </div>
  );
}

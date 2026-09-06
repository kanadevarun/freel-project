import { useState, useEffect } from 'react';
import { createPortal } from 'react-dom';
import { 
  Ship, Network, AlertTriangle, History, Search, MoreHorizontal, 
  Info, CheckCircle2, XCircle, RefreshCw, Radio, Lock, 
  Activity, Globe, Eye, EyeOff, Edit3, Trash2, Power, 
  ExternalLink, Layers, ArrowRight, ShieldCheck
} from 'lucide-react';
import { carrierIntegrationsService } from '../../../services/carrierIntegrationsService';
import CarrierSelectDropdown from './CarrierSelectDropdown';
import StatusFilterDropdown from './StatusFilterDropdown';
import './CarrierIntegrationsPage.css';

const ALL_CAPABILITIES = [
  { key: 'TRACKING', label: 'Tracking & Telemetry', desc: 'Real-time container & milestone events' },
  { key: 'RATES', label: 'Tariff & Spot Rates', desc: 'Port-pair dynamic rate discovery' },
  { key: 'CONTRACT_RATES', label: 'Contract Rates', desc: 'Confidential client contract rates' },
  { key: 'SPOT_RATES', label: 'Guaranteed Spot', desc: 'Instant booking spot pricing' },
  { key: 'BOOKING', label: 'Space Booking', desc: 'Direct EDI/API allocation reservations' },
  { key: 'DOCUMENTS', label: 'Transport Documents', desc: 'B/Ls, arrival notices & manifests' }
];

export default function CarrierIntegrationsPage() {
  const [carriers, setCarriers] = useState([]);
  const [providers, setProviders] = useState([]);
  const [isLoading, setIsLoading] = useState(true);
  const [openMenuId, setOpenMenuId] = useState(null);
  const [searchQuery, setSearchQuery] = useState('');
  const [statusFilter, setStatusFilter] = useState('ALL');
  const [feedbackMsg, setFeedbackMsg] = useState(null);
  
  // Modals & Drawers
  const [showConnectModal, setShowConnectModal] = useState(false);
  const [showEditModal, setShowEditModal] = useState(false);
  const [showDisconnectModal, setShowDisconnectModal] = useState(false);
  const [showHistoryModal, setShowHistoryModal] = useState(false);
  const [showHealthModal, setShowHealthModal] = useState(false);
  const [selectedIntegration, setSelectedIntegration] = useState(null);
  const [syncingId, setSyncingId] = useState(null);

  // History & Health State
  const [historyCarrier, setHistoryCarrier] = useState(null);
  const [historyItems, setHistoryItems] = useState([]);
  const [historyTotal, setHistoryTotal] = useState(0);
  const [isLoadingHistory, setIsLoadingHistory] = useState(false);
  const [healthCarrier, setHealthCarrier] = useState(null);
  const [healthData, setHealthData] = useState(null);
  const [isLoadingHealth, setIsLoadingHealth] = useState(false);
  
  // Connect / Edit Form State
  const [selectedProviderCode, setSelectedProviderCode] = useState('');
  const [useRawJson, setUseRawJson] = useState(false);
  const [showSecrets, setShowSecrets] = useState({});
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [isTesting, setIsTesting] = useState(false);
  const [testResult, setTestResult] = useState(null);
  const [rotateCredentials, setRotateCredentials] = useState(false);

  const initialFormData = {
    carrier_scac: '',
    carrier_provider_id: null,
    carrier_name: '',
    connection_method: 'API',
    environment: 'PRODUCTION',
    credentials: {},
    credentials_json: '{\n  "api_key": "",\n  "api_secret": ""\n}',
    capabilities: {
      TRACKING: true,
      RATES: false,
      CONTRACT_RATES: false,
      SPOT_RATES: false,
      BOOKING: false,
      DOCUMENTS: false
    }
  };

  const [formData, setFormData] = useState(initialFormData);

  const fetchCarriers = async () => {
    try {
      setIsLoading(true);
      const res = await carrierIntegrationsService.getIntegrations();
      const list = Array.isArray(res) ? res : (Array.isArray(res?.data) ? res.data : (Array.isArray(res?.data?.data) ? res.data.data : []));
      setCarriers(list);
    } catch (err) {
      console.error('Failed to fetch carrier integrations:', err);
      setCarriers([]);
    } finally {
      setIsLoading(false);
    }
  };

  const fetchProviders = async () => {
    try {
      const res = await carrierIntegrationsService.getProviders();
      const list = Array.isArray(res) ? res : (Array.isArray(res?.data) ? res.data : (Array.isArray(res?.data?.data) ? res.data.data : []));
      setProviders(list);
    } catch (err) {
      console.error('Failed to fetch carrier providers:', err);
    }
  };

  useEffect(() => {
    fetchCarriers();
    fetchProviders();
  }, []);

  // Close dropdown menu on outside click
  useEffect(() => {
    const handleClickOutside = (e) => {
      if (!e.target.closest('.ci-actions-dropdown-wrapper')) {
        setOpenMenuId(null);
      }
    };
    document.addEventListener('click', handleClickOutside);
    return () => document.removeEventListener('click', handleClickOutside);
  }, []);

  const showToast = (msg, type = 'success') => {
    setFeedbackMsg({ msg, type });
    setTimeout(() => {
      setFeedbackMsg(null);
    }, 4000);
  };

  const toggleSecretVisibility = (fieldKey) => {
    setShowSecrets(prev => ({ ...prev, [fieldKey]: !prev[fieldKey] }));
  };

  const handleProviderSelect = (providerCode) => {
    setSelectedProviderCode(providerCode);
    setTestResult(null);

    if (providerCode === 'CUSTOM') {
      setFormData(prev => ({
        ...prev,
        carrier_scac: '',
        carrier_provider_id: null,
        carrier_name: '',
        credentials: { api_key: '', api_secret: '', base_url: '' },
        credentials_json: '{\n  "api_key": "",\n  "api_secret": "",\n  "base_url": ""\n}',
        capabilities: {
          TRACKING: true,
          RATES: true,
          CONTRACT_RATES: false,
          SPOT_RATES: false,
          BOOKING: true,
          DOCUMENTS: true
        }
      }));
      return;
    }

    const found = providers.find(p => p.code === providerCode);
    if (found) {
      const caps = {
        TRACKING: false,
        RATES: false,
        CONTRACT_RATES: false,
        SPOT_RATES: false,
        BOOKING: false,
        DOCUMENTS: false
      };
      if (Array.isArray(found.supported_capabilities)) {
        found.supported_capabilities.forEach(c => {
          caps[c] = true;
        });
      } else {
        caps.TRACKING = true;
      }

      const creds = {};
      if (Array.isArray(found.credential_fields)) {
        found.credential_fields.forEach(f => {
          creds[f.key] = '';
        });
      } else {
        creds.api_key = '';
        creds.api_secret = '';
      }

      setFormData(prev => ({
        ...prev,
        carrier_scac: found.scac,
        carrier_provider_id: found.id,
        carrier_name: found.name,
        credentials: creds,
        credentials_json: JSON.stringify(creds, null, 2),
        capabilities: caps
      }));
    }
  };

  const handleCredentialChange = (key, value) => {
    const updated = { ...formData.credentials, [key]: value };
    setFormData(prev => ({
      ...prev,
      credentials: updated,
      credentials_json: JSON.stringify(updated, null, 2)
    }));
  };

  const handleRawJsonChange = (raw) => {
    let parsed = {};
    try {
      parsed = JSON.parse(raw);
    } catch (e) {}
    setFormData(prev => ({
      ...prev,
      credentials_json: raw,
      credentials: typeof parsed === 'object' && parsed !== null ? parsed : prev.credentials
    }));
  };

  const toggleCapability = (cap) => {
    setFormData(prev => ({
      ...prev,
      capabilities: { ...prev.capabilities, [cap]: !prev.capabilities[cap] }
    }));
  };

  // Pre-flight test in modal
  const handleModalTest = async () => {
    setIsTesting(true);
    setTestResult(null);
    try {
      if (!formData.carrier_scac) {
        throw new Error('Carrier selection or SCAC code is required.');
      }

      let payloadCreds = formData.credentials;
      if (useRawJson) {
        try {
          payloadCreds = JSON.parse(formData.credentials_json);
        } catch (e) {
          throw new Error('Invalid JSON format in credentials field.');
        }
      }

      const payload = {
        carrier_scac: formData.carrier_scac,
        environment: formData.environment,
        connection_method: formData.connection_method,
        credentials: payloadCreds,
        credentials_json: JSON.stringify(payloadCreds)
      };

      const res = await carrierIntegrationsService.testDirectConnection(payload);
      const data = res.data || res;
      setTestResult({
        success: data.success !== false,
        message: data.message || 'Connection verified successfully with carrier adapter.',
        latency: data.latency_ms,
        environment: data.tested_environment,
        caps: data.tested_capabilities
      });
    } catch (err) {
      setTestResult({
        success: false,
        message: err.response?.data?.message || err.message
      });
    } finally {
      setIsTesting(false);
    }
  };

  // Connect Submit
  const handleConnectSubmit = async (e) => {
    e.preventDefault();
    setIsSubmitting(true);
    setTestResult(null);

    try {
      if (!formData.carrier_scac) {
        throw new Error('Carrier SCAC code is required.');
      }

      let payloadCreds = formData.credentials;
      if (useRawJson) {
        try {
          payloadCreds = JSON.parse(formData.credentials_json);
        } catch (e) {
          throw new Error('Invalid JSON format in credentials field.');
        }
      }

      const activeCaps = Object.keys(formData.capabilities).filter(k => formData.capabilities[k]);

      const payload = {
        carrier_scac: formData.carrier_scac,
        environment: formData.environment,
        connection_method: formData.connection_method,
        credentials: payloadCreds,
        credentials_json: JSON.stringify(payloadCreds),
        capabilities: activeCaps
      };

      await carrierIntegrationsService.connectCarrier(payload);
      setShowConnectModal(false);
      setFormData(initialFormData);
      showToast(`Successfully connected ${formData.carrier_name || formData.carrier_scac}!`, 'success');
      fetchCarriers();
    } catch (err) {
      setTestResult({
        success: false,
        message: err.response?.data?.message || err.message
      });
    } finally {
      setIsSubmitting(false);
    }
  };

  // Open Edit Modal
  const handleOpenEdit = (carrier) => {
    setOpenMenuId(null);
    setSelectedIntegration(carrier);
    setTestResult(null);
    setRotateCredentials(false);
    setUseRawJson(false);

    // Parse capabilities
    const caps = {
      TRACKING: false,
      RATES: false,
      CONTRACT_RATES: false,
      SPOT_RATES: false,
      BOOKING: false,
      DOCUMENTS: false
    };
    if (Array.isArray(carrier.capabilities)) {
      carrier.capabilities.forEach(c => { caps[c] = true; });
    }

    const foundProvider = providers.find(p => p.scac === carrier.carrier_scac || p.code === carrier.carrier_code);
    setSelectedProviderCode(foundProvider ? foundProvider.code : 'CUSTOM');

    setFormData({
      carrier_scac: carrier.carrier_scac,
      carrier_provider_id: carrier.carrier_provider_id || (foundProvider ? foundProvider.id : null),
      carrier_name: carrier.carrier_name || carrier.carrier_scac,
      connection_method: carrier.connection_method || 'API',
      environment: carrier.environment || 'PRODUCTION',
      credentials: {},
      credentials_json: '{\n  "api_key": "",\n  "api_secret": ""\n}',
      capabilities: caps
    });

    setShowEditModal(true);
  };

  // Edit Submit
  const handleEditSubmit = async (e) => {
    e.preventDefault();
    if (!selectedIntegration) return;
    setIsSubmitting(true);
    setTestResult(null);

    try {
      const activeCaps = Object.keys(formData.capabilities).filter(k => formData.capabilities[k]);
      const payload = {
        environment: formData.environment,
        connection_method: formData.connection_method,
        capabilities: activeCaps
      };

      if (rotateCredentials) {
        let payloadCreds = formData.credentials;
        if (useRawJson) {
          try {
            payloadCreds = JSON.parse(formData.credentials_json);
          } catch (e) {
            throw new Error('Invalid JSON format in rotated credentials.');
          }
        }
        payload.credentials = payloadCreds;
        payload.credentials_json = JSON.stringify(payloadCreds);
      }

      await carrierIntegrationsService.updateCarrier(selectedIntegration.id, payload);
      setShowEditModal(false);
      showToast(`Updated configuration for ${selectedIntegration.carrier_name}!`, 'success');
      fetchCarriers();
    } catch (err) {
      setTestResult({
        success: false,
        message: err.response?.data?.message || err.message
      });
    } finally {
      setIsSubmitting(false);
    }
  };

  // Test Existing Connection
  const handleTestConnection = async (id) => {
    setOpenMenuId(null);
    try {
      showToast('Testing carrier connection...', 'info');
      const res = await carrierIntegrationsService.testConnection(id);
      const data = res.data || res;
      showToast(data.message || 'Connection test successful!', 'success');
      fetchCarriers();
    } catch (err) {
      showToast('Test failed: ' + (err.response?.data?.message || err.message), 'error');
      fetchCarriers();
    }
  };

  // Sync Telemetry & Bookings (Manual Trigger)
  const handleSyncCarrier = async (id, carrierName = 'Carrier') => {
    setOpenMenuId(null);
    setSyncingId(id);
    try {
      showToast(`Syncing ${carrierName} tracking & booking allocations...`, 'info');
      const res = await carrierIntegrationsService.syncNow(id, { operation: 'FULL_SYNC' });
      const data = res.data || res;
      const count = data.records_processed ?? 0;
      showToast(`${carrierName} sync completed (${count} record${count === 1 ? '' : 's'} synchronized)`, 'success');
      fetchCarriers();
      if (showHistoryModal && historyCarrier && historyCarrier.id === id) {
        fetchSyncHistory(id);
      }
    } catch (err) {
      if (err.response?.status === 409) {
        showToast('Sync already in progress for this carrier.', 'info');
      } else {
        const errMsg = err.response?.data?.error?.message || err.response?.data?.message || err.message;
        showToast(`Sync failed: ${errMsg}`, 'error');
      }
      fetchCarriers();
    } finally {
      setSyncingId(null);
    }
  };

  // Fetch Sync History Logs
  const fetchSyncHistory = async (integrationId) => {
    try {
      setIsLoadingHistory(true);
      const res = await carrierIntegrationsService.getSyncHistory(integrationId, { limit: 20 });
      const data = res.data || res || {};
      setHistoryItems(data.jobs || []);
      setHistoryTotal(data.total || 0);
    } catch (err) {
      console.error('Failed to load sync history:', err);
      showToast('Failed to load sync history logs.', 'error');
    } finally {
      setIsLoadingHistory(false);
    }
  };

  const handleOpenHistory = (carrier) => {
    setOpenMenuId(null);
    setHistoryCarrier(carrier);
    setShowHistoryModal(true);
    fetchSyncHistory(carrier.id);
  };

  // Open Health Telemetry Modal
  const handleOpenHealth = async (carrier) => {
    setOpenMenuId(null);
    setHealthCarrier(carrier);
    setShowHealthModal(true);
    try {
      setIsLoadingHealth(true);
      const res = await carrierIntegrationsService.getIntegrationHealth(carrier.id);
      const data = res.data || res || {};
      setHealthData(data.health || data);
    } catch (err) {
      console.error('Failed to load health telemetry:', err);
      showToast('Failed to load health diagnostics.', 'error');
    } finally {
      setIsLoadingHealth(false);
    }
  };

  // Toggle Active / Disabled
  const handleToggleActive = async (id, currentStatus) => {
    setOpenMenuId(null);
    try {
      const nextActive = currentStatus === 'DISABLED';
      await carrierIntegrationsService.toggleCarrier(id, nextActive);
      showToast(`Carrier integration ${nextActive ? 'enabled' : 'disabled'}.`, 'success');
      fetchCarriers();
    } catch (err) {
      showToast('Failed to toggle status: ' + err.message, 'error');
    }
  };

  // Open Disconnect Modal
  const handleOpenDisconnect = (carrier) => {
    setOpenMenuId(null);
    setSelectedIntegration(carrier);
    setShowDisconnectModal(true);
  };

  // Confirm Disconnect
  const handleConfirmDisconnect = async () => {
    if (!selectedIntegration) return;
    setIsSubmitting(true);
    try {
      await carrierIntegrationsService.removeCarrier(selectedIntegration.id);
      setShowDisconnectModal(false);
      showToast(`${selectedIntegration.carrier_name} disconnected successfully.`, 'success');
      fetchCarriers();
    } catch (err) {
      showToast('Failed to disconnect carrier: ' + err.message, 'error');
    } finally {
      setIsSubmitting(false);
    }
  };

  // Compute Summary KPI Counts
  const activeCount = carriers.filter(c => (c.connection_status === 'CONNECTED' || c.connection_status === 'Connected') && c.is_active !== false).length;
  const needsAttentionCount = carriers.filter(c => 
    c.health_state === 'ATTENTION' || 
    c.health_state === 'ERROR' || 
    c.connection_status === 'ERROR' || 
    c.connection_status === 'Error' || 
    c.failed_attempts >= 1 || 
    c.sync_status === 'Failed'
  ).length;
  
  let lastSyncTime = 'Never';
  const validSyncs = carriers
    .map(c => new Date(c.last_success_at || c.last_synced_at).getTime())
    .filter(t => !isNaN(t) && t > 0);
  if (validSyncs.length > 0) {
    const maxTime = Math.max(...validSyncs);
    const diffMins = Math.floor((Date.now() - maxTime) / (1000 * 60));
    if (diffMins < 1) {
      lastSyncTime = 'Just now';
    } else if (diffMins < 60) {
      lastSyncTime = `${diffMins}m ago`;
    } else {
      lastSyncTime = new Date(maxTime).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
    }
  }

  // Filter & Search List
  const filteredCarriers = carriers.filter(c => {
    const q = searchQuery.toLowerCase();
    const matchQuery = !q || 
      (c.carrier_name || '').toLowerCase().includes(q) ||
      (c.carrier_scac || '').toLowerCase().includes(q) ||
      (c.environment || '').toLowerCase().includes(q) ||
      (c.connection_method || '').toLowerCase().includes(q);
    
    if (!matchQuery) return false;
    if (statusFilter === 'ACTIVE') return c.is_active && (c.connection_status === 'CONNECTED' || c.connection_status === 'Connected');
    if (statusFilter === 'ERROR') return c.connection_status === 'ERROR' || c.connection_status === 'Error' || c.sync_status === 'Failed';
    if (statusFilter === 'DISABLED') return c.connection_status === 'DISABLED' || c.is_active === false;
    return true;
  });

  // Current Provider definition for dynamic form rendering
  const currentProvider = providers.find(p => p.code === selectedProviderCode);

  return (
    <div className="ci-page">
      {/* Toast Notification Banner */}
      {feedbackMsg && (
        <div style={{
          position: 'fixed',
          top: '24px',
          right: '24px',
          zIndex: 9999,
          padding: '12px 20px',
          borderRadius: '10px',
          background: feedbackMsg.type === 'error' ? '#ef4444' : feedbackMsg.type === 'info' ? '#3b82f6' : '#10b981',
          color: 'white',
          boxShadow: '0 12px 24px -4px rgba(0,0,0,0.25)',
          fontSize: '14px',
          fontWeight: 600,
          display: 'flex',
          alignItems: 'center',
          gap: '10px'
        }}>
          {feedbackMsg.type === 'error' ? <XCircle size={18} /> : feedbackMsg.type === 'info' ? <Activity size={18} /> : <CheckCircle2 size={18} />}
          {feedbackMsg.msg}
        </div>
      )}

      {/* Header */}
      <div className="ci-header-row">
        <div className="ci-header-left">
          <h1 className="ci-title">
            <div className="ci-title-icon">
              <Ship size={24} />
            </div>
            Carrier Integrations
          </h1>
          <p className="ci-subtitle">
            Connect and manage API & EDI integrations to ocean shipping lines, air carriers, and tracking providers.
          </p>
        </div>
        <button 
          className="ci-btn-primary" 
          onClick={() => {
            setFormData(initialFormData);
            setSelectedProviderCode('');
            setTestResult(null);
            setShowConnectModal(true);
          }}
        >
          + Connect Carrier
        </button>
      </div>

      {/* Summary Cards */}
      <div className="ci-summary-grid">
        <div className="ci-summary-card">
          <div className="ci-summary-icon-wrapper blue">
            <Network size={24} />
          </div>
          <div className="ci-summary-content">
            <span className="ci-summary-label">Connected Carriers</span>
            <span className="ci-summary-value">{carriers.length}</span>
            <span className="ci-summary-desc">Configured shipping lines</span>
          </div>
        </div>
        
        <div className="ci-summary-card">
          <div className="ci-summary-icon-wrapper green">
            <ShieldCheck size={24} />
          </div>
          <div className="ci-summary-content">
            <span className="ci-summary-label">Active Connections</span>
            <span className="ci-summary-value">{activeCount}</span>
            <span className="ci-summary-desc">Healthy & verified</span>
          </div>
        </div>

        <div className="ci-summary-card">
          <div className="ci-summary-icon-wrapper orange">
            <AlertTriangle size={24} />
          </div>
          <div className="ci-summary-content">
            <span className="ci-summary-label">Needs Attention</span>
            <span className="ci-summary-value">{needsAttentionCount}</span>
            <span className="ci-summary-desc">Auth or sync alerts</span>
          </div>
        </div>

        <div className="ci-summary-card">
          <div className="ci-summary-icon-wrapper purple">
            <History size={24} />
          </div>
          <div className="ci-summary-content">
            <span className="ci-summary-label">Last Sync</span>
            <span className="ci-summary-value">{lastSyncTime}</span>
            <span className="ci-summary-desc">Background scheduler</span>
          </div>
        </div>
      </div>

      {/* Connected Carriers Table Card */}
      <div className="ci-card">
        <div className="ci-card-header">
          <div className="ci-card-title-row">
            <div>
              <h2 className="ci-card-title">Connected Carriers</h2>
              <p className="ci-card-desc">Manage API credentials, capabilities, sync health, and environments.</p>
            </div>
            <div className="ci-search-box">
              <div className="ci-search-input-wrapper">
                <Search size={16} className="ci-search-icon" />
                <input 
                  type="text" 
                  className="ci-search-input" 
                  placeholder="Search by carrier, SCAC, env..." 
                  value={searchQuery}
                  onChange={e => setSearchQuery(e.target.value)}
                />
              </div>
              <StatusFilterDropdown 
                value={statusFilter}
                onChange={setStatusFilter}
              />
            </div>
          </div>
        </div>

        <div className="ci-table-responsive">
          <table className="ci-table">
            <thead>
              <tr>
                <th>Carrier</th>
                <th>Connection Type</th>
                <th>Capabilities</th>
                <th>Integration Health</th>
                <th>Last Synchronization</th>
                <th>Actions</th>
              </tr>
            </thead>
            <tbody>
              {isLoading ? (
                <tr>
                  <td colSpan="6" className="ci-empty-state">
                    <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', gap: '8px', padding: '32px' }}>
                      <RefreshCw size={18} className="spin" />
                      <span>Loading carrier integrations...</span>
                    </div>
                  </td>
                </tr>
              ) : filteredCarriers.length === 0 ? (
                <tr>
                  <td colSpan="6" className="ci-empty-state">
                    <div style={{ padding: '40px 20px', textAlign: 'center' }}>
                      <div style={{
                        width: '56px',
                        height: '56px',
                        borderRadius: '16px',
                        background: 'rgba(59, 130, 246, 0.08)',
                        color: '#3b82f6',
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'center',
                        margin: '0 auto 16px auto'
                      }}>
                        <Globe size={28} />
                      </div>
                      <h3 style={{ fontSize: '16px', fontWeight: 600, color: '#0f172a', marginBottom: '6px' }}>
                        No Carrier Connections Yet
                      </h3>
                      <p style={{ fontSize: '13px', color: '#64748b', maxWidth: '460px', margin: '0 auto 16px auto', lineHeight: 1.5 }}>
                        Connect Maersk, MSC, Hapag-Lloyd, CMA CGM, or custom shipping lines via API to automate container tracking, rate discovery, and space bookings.
                      </p>
                      <button 
                        className="ci-btn-primary" 
                        onClick={() => {
                          setFormData(initialFormData);
                          setSelectedProviderCode('');
                          setTestResult(null);
                          setShowConnectModal(true);
                        }}
                      >
                        + Connect Carrier
                      </button>
                    </div>
                  </td>
                </tr>
              ) : (
                filteredCarriers.map(carrier => {
                  let capabilitiesList = [];
                  if (Array.isArray(carrier.capabilities)) {
                    capabilitiesList = carrier.capabilities;
                  } else if (typeof carrier.capabilities === 'string') {
                    try {
                      const parsed = JSON.parse(carrier.capabilities);
                      if (Array.isArray(parsed)) {
                        capabilitiesList = parsed;
                      } else if (typeof parsed === 'object') {
                        capabilitiesList = Object.keys(parsed).filter(k => parsed[k]);
                      }
                    } catch(e) {}
                  }
                  if (capabilitiesList.length === 0) capabilitiesList = ['TRACKING'];

                  let formattedSync = 'Never synced';
                  const syncTimestamp = carrier.last_success_at || carrier.last_synced_at;
                  if (syncTimestamp && syncTimestamp !== '0001-01-01T00:00:00Z') {
                    formattedSync = new Date(syncTimestamp).toLocaleString([], { dateStyle: 'short', timeStyle: 'short' });
                  }

                  const isSyncing = syncingId === carrier.id || carrier.is_syncing || carrier.sync_status === 'Syncing' || carrier.sync_status === 'Running';
                  const healthState = carrier.health_state || (
                    carrier.is_active === false ? 'DISABLED' :
                    carrier.connection_status === 'DISCONNECTED' ? 'DISCONNECTED' :
                    carrier.failed_attempts >= 5 ? 'ERROR' :
                    (carrier.failed_attempts > 0 || carrier.connection_status === 'ERROR') ? 'ATTENTION' :
                    'HEALTHY'
                  );

                  const healthBadgeStyle = {
                    HEALTHY: { bg: '#ecfdf5', text: '#059669', border: '#a7f3d0', label: 'Healthy', dot: '#10b981' },
                    ATTENTION: { bg: '#fffbeb', text: '#b45309', border: '#fde68a', label: 'Attention', dot: '#f59e0b' },
                    ERROR: { bg: '#fef2f2', text: '#dc2626', border: '#fecaca', label: 'Action Required', dot: '#ef4444' },
                    DISABLED: { bg: '#f8fafc', text: '#64748b', border: '#e2e8f0', label: 'Disabled', dot: '#94a3b8' },
                    DISCONNECTED: { bg: '#f1f5f9', text: '#475569', border: '#cbd5e1', label: 'Disconnected', dot: '#64748b' }
                  }[healthState] || { bg: '#f8fafc', text: '#64748b', border: '#e2e8f0', label: healthState, dot: '#94a3b8' };

                  const isConnected = carrier.connection_status === 'CONNECTED' || carrier.connection_status === 'Connected';
                  const isDisabled = carrier.connection_status === 'DISABLED' || carrier.is_active === false;

                  return (
                    <tr key={carrier.id}>
                      <td>
                        <div className="ci-carrier-cell">
                          <div className="ci-carrier-logo" style={{
                            background: carrier.carrier_scac === 'MAEU' ? '#e0f2fe' : carrier.carrier_scac === 'MSCU' ? '#fef3c7' : carrier.carrier_scac === 'HLCU' ? '#fee2e2' : '#f1f5f9',
                            color: carrier.carrier_scac === 'MAEU' ? '#0284c7' : carrier.carrier_scac === 'MSCU' ? '#b45309' : carrier.carrier_scac === 'HLCU' ? '#dc2626' : '#334155',
                            fontWeight: 700
                          }}>
                            {carrier.carrier_scac?.substring(0, 2) || 'CA'}
                          </div>
                          <div className="ci-carrier-info">
                            <span className="ci-carrier-name">{carrier.carrier_name || carrier.carrier_scac}</span>
                            <span className="ci-carrier-type">SCAC: {carrier.carrier_scac}</span>
                          </div>
                        </div>
                      </td>
                      <td>
                        <div className="ci-conn-cell">
                          <span className="ci-conn-type">{carrier.connection_method || 'API'}</span>
                          <span className="ci-conn-env" style={{
                            background: carrier.environment === 'SANDBOX' ? '#fef3c7' : '#e0e7ff',
                            color: carrier.environment === 'SANDBOX' ? '#92400e' : '#3730a3'
                          }}>
                            {carrier.environment || 'PRODUCTION'}
                          </span>
                        </div>
                      </td>
                      <td>
                        <div className="ci-caps-cell">
                          {capabilitiesList.slice(0, 3).map(c => (
                            <span key={c} className="ci-cap-badge">✓ {c.replace('_', ' ')}</span>
                          ))}
                          {capabilitiesList.length > 3 && (
                            <span className="ci-cap-badge" style={{ background: '#f1f5f9', color: '#64748b' }}>
                              +{capabilitiesList.length - 3} more
                            </span>
                          )}
                        </div>
                      </td>
                      <td>
                        <div className="ci-status-cell">
                          <button
                            type="button"
                            onClick={() => handleOpenHealth(carrier)}
                            style={{
                              display: 'inline-flex',
                              alignItems: 'center',
                              gap: '6px',
                              padding: '4px 10px',
                              borderRadius: '12px',
                              fontSize: '12px',
                              fontWeight: 600,
                              background: healthBadgeStyle.bg,
                              color: healthBadgeStyle.text,
                              border: `1px solid ${healthBadgeStyle.border}`,
                              cursor: 'pointer'
                            }}
                            title="Click to view health diagnostics & telemetry"
                          >
                            <div style={{
                              width: '6px',
                              height: '6px',
                              borderRadius: '50%',
                              backgroundColor: healthBadgeStyle.dot
                            }}></div>
                            <span>{healthBadgeStyle.label}</span>
                          </button>
                          {carrier.health_reason && (
                            <div style={{ fontSize: '11px', color: '#64748b', marginTop: '3px', maxWidth: '170px', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }} title={carrier.health_reason}>
                              {carrier.health_reason}
                            </div>
                          )}
                        </div>
                      </td>
                      <td>
                        <div className="ci-sync-cell">
                          <div style={{ display: 'flex', alignItems: 'center', gap: '6px', fontSize: '12px', color: '#475569' }}>
                            <Lock size={12} color="#10b981" />
                            <span>Encrypted at rest</span>
                          </div>
                          <span className="ci-sync-relative" style={{ marginTop: '2px', fontSize: '11.5px' }}>
                            {isSyncing ? (
                              <span style={{ color: '#2563eb', fontWeight: 600, display: 'inline-flex', alignItems: 'center', gap: '4px' }}>
                                <RefreshCw size={11} className="spin" /> Syncing...
                              </span>
                            ) : (
                              `Last Sync: ${formattedSync}`
                            )}
                          </span>
                          {carrier.failed_attempts > 0 && (
                            <div style={{ fontSize: '11px', color: '#ef4444', marginTop: '2px', fontWeight: 500 }}>
                              ⚠️ {carrier.failed_attempts} failed attempt{carrier.failed_attempts === 1 ? '' : 's'}
                            </div>
                          )}
                        </div>
                      </td>
                      <td>
                        <div className="ci-actions-cell ci-actions-dropdown-wrapper" style={{ position: 'relative', display: 'flex', alignItems: 'center', gap: '6px' }}>
                          <button 
                            className="ci-btn-outline" 
                            style={{ 
                              fontSize: '12px', 
                              padding: '4px 10px', 
                              display: 'inline-flex', 
                              alignItems: 'center', 
                              gap: '4px',
                              color: isSyncing ? '#2563eb' : undefined,
                              borderColor: isSyncing ? '#93c5fd' : undefined
                            }}
                            onClick={() => handleSyncCarrier(carrier.id, carrier.carrier_name || carrier.carrier_scac)}
                            disabled={isSyncing}
                            title="Synchronize real-time tracking milestones & booking allocations"
                          >
                            <RefreshCw size={12} className={isSyncing ? 'spin' : ''} />
                            <span>{isSyncing ? 'Syncing...' : 'Sync'}</span>
                          </button>
                          
                          <button 
                            className="ci-btn-outline"
                            style={{ fontSize: '12px', padding: '4px 8px', color: '#475569' }}
                            onClick={() => handleOpenHistory(carrier)}
                            title="View synchronization job history logs"
                          >
                            <History size={13} />
                          </button>

                          <button 
                            className="ci-btn-icon" 
                            onClick={(e) => {
                              e.stopPropagation();
                              setOpenMenuId(openMenuId === carrier.id ? null : carrier.id);
                            }}
                          >
                            <MoreHorizontal size={16} />
                          </button>
                          
                          {openMenuId === carrier.id && (
                            <div className="ci-dropdown-menu" style={{
                              position: 'absolute',
                              right: 0,
                              top: '100%',
                              zIndex: 50,
                              background: 'white',
                              border: '1px solid #e2e8f0',
                              borderRadius: '8px',
                              boxShadow: '0 10px 15px -3px rgba(0,0,0,0.1)',
                              padding: '6px',
                              minWidth: '175px'
                            }}>
                              <button 
                                onClick={(e) => { e.stopPropagation(); handleSyncCarrier(carrier.id, carrier.carrier_name || carrier.carrier_scac); }}
                                style={{ width: '100%', padding: '8px 12px', textAlign: 'left', background: 'none', border: 'none', cursor: 'pointer', fontSize: '13px', borderRadius: '4px', color: '#1e293b', display: 'flex', alignItems: 'center', gap: '8px' }}
                              >
                                <RefreshCw size={14} /> Sync Now
                              </button>
                              <button 
                                onClick={(e) => { e.stopPropagation(); handleOpenHistory(carrier); }}
                                style={{ width: '100%', padding: '8px 12px', textAlign: 'left', background: 'none', border: 'none', cursor: 'pointer', fontSize: '13px', borderRadius: '4px', color: '#1e293b', display: 'flex', alignItems: 'center', gap: '8px' }}
                              >
                                <History size={14} /> Sync History Logs
                              </button>
                              <button 
                                onClick={(e) => { e.stopPropagation(); handleOpenHealth(carrier); }}
                                style={{ width: '100%', padding: '8px 12px', textAlign: 'left', background: 'none', border: 'none', cursor: 'pointer', fontSize: '13px', borderRadius: '4px', color: '#1e293b', display: 'flex', alignItems: 'center', gap: '8px' }}
                              >
                                <Activity size={14} /> Health Diagnostics
                              </button>
                              <button 
                                onClick={(e) => { e.stopPropagation(); handleTestConnection(carrier.id); }}
                                style={{ width: '100%', padding: '8px 12px', textAlign: 'left', background: 'none', border: 'none', cursor: 'pointer', fontSize: '13px', borderRadius: '4px', color: '#1e293b', display: 'flex', alignItems: 'center', gap: '8px' }}
                              >
                                <Radio size={14} /> Test Connection
                              </button>
                              <button 
                                onClick={(e) => { e.stopPropagation(); handleOpenEdit(carrier); }}
                                style={{ width: '100%', padding: '8px 12px', textAlign: 'left', background: 'none', border: 'none', cursor: 'pointer', fontSize: '13px', borderRadius: '4px', color: '#1e293b', display: 'flex', alignItems: 'center', gap: '8px' }}
                              >
                                <Edit3 size={14} /> Edit Configuration
                              </button>
                              <button 
                                onClick={(e) => { e.stopPropagation(); handleToggleActive(carrier.id, carrier.connection_status); }}
                                style={{ width: '100%', padding: '8px 12px', textAlign: 'left', background: 'none', border: 'none', cursor: 'pointer', fontSize: '13px', borderRadius: '4px', color: '#1e293b', display: 'flex', alignItems: 'center', gap: '8px' }}
                              >
                                <Power size={14} /> {isDisabled ? 'Enable Integration' : 'Disable Integration'}
                              </button>
                              <div style={{ height: '1px', background: '#e2e8f0', margin: '4px 0' }}></div>
                              <button 
                                onClick={(e) => { e.stopPropagation(); handleOpenDisconnect(carrier); }}
                                style={{ width: '100%', padding: '8px 12px', textAlign: 'left', background: 'none', border: 'none', cursor: 'pointer', fontSize: '13px', borderRadius: '4px', color: '#ef4444', display: 'flex', alignItems: 'center', gap: '8px' }}
                              >
                                <Trash2 size={14} /> Disconnect
                              </button>
                            </div>
                          )}
                        </div>
                      </td>
                    </tr>
                  );
                })
              )}
            </tbody>
          </table>
        </div>

        {/* Pagination */}
        <div className="ci-pagination">
          <span className="ci-pagination-text">
            {filteredCarriers.length === 0 
              ? 'Showing 0 to 0 of 0 carriers' 
              : `Showing 1 to ${filteredCarriers.length} of ${filteredCarriers.length} carriers`}
          </span>
        </div>
      </div>

      {/* Info Bar */}
      <div className="ci-info-bar">
        <div className="ci-info-left">
          <Info size={18} className="ci-info-icon" />
          <div className="ci-info-content">
            <span className="ci-info-title">Carrier Integration Architecture & Security</span>
            <span className="ci-info-desc">
              All credentials are encrypted using AES-256-GCM at rest. Adapter telemetry runs in isolated tenant sandboxes without secret exposure in frontend responses or logs.
            </span>
          </div>
        </div>
      </div>

      {/* ─────────────────────────────────────────────────────────────────── */}
      {/* Connect Carrier Modal                                               */}
      {/* ─────────────────────────────────────────────────────────────────── */}
      {showConnectModal && createPortal(
        <div className="ci-modal-overlay">
          <div className="ci-modal" style={{ maxWidth: '640px' }}>
            <div className="ci-modal-header">
              <div>
                <h2>Connect Carrier Integration</h2>
                <p style={{ fontSize: '13px', color: '#64748b', marginTop: '2px' }}>
                  Select a carrier provider and configure encrypted API credentials.
                </p>
              </div>
              <button className="ci-modal-close" onClick={() => setShowConnectModal(false)}>×</button>
            </div>
            
            <form onSubmit={handleConnectSubmit} className="ci-modal-form">
              {/* Carrier Selection */}
              <div className="ci-form-group">
                <label>Carrier Provider *</label>
                <CarrierSelectDropdown
                  providers={providers}
                  selectedProviderCode={selectedProviderCode}
                  onSelectProvider={handleProviderSelect}
                />
              </div>

              {selectedProviderCode && (
                <>
                  <div className="ci-form-row">
                    <div className="ci-form-group">
                      <label>Carrier SCAC Code *</label>
                      <input 
                        type="text" 
                        required 
                        placeholder="e.g. MAEU, MSCU, HLCU" 
                        value={formData.carrier_scac}
                        onChange={e => setFormData({ ...formData, carrier_scac: e.target.value.toUpperCase() })}
                        className="ci-form-input"
                        readOnly={selectedProviderCode !== 'CUSTOM'}
                        style={selectedProviderCode !== 'CUSTOM' ? { background: '#f1f5f9', cursor: 'not-allowed' } : {}}
                      />
                    </div>

                    <div className="ci-form-group">
                      <label>Environment Mode</label>
                      <div className="ci-env-selector">
                        <button
                          type="button"
                          className={`ci-env-btn ${formData.environment === 'PRODUCTION' ? 'active-prod' : ''}`}
                          onClick={() => setFormData({ ...formData, environment: 'PRODUCTION' })}
                        >
                          <span className="ci-env-dot prod" />
                          <div className="ci-env-text">
                            <span className="ci-env-title">Production</span>
                            <span className="ci-env-sub">Live DCSA APIs</span>
                          </div>
                        </button>
                        <button
                          type="button"
                          className={`ci-env-btn ${formData.environment === 'SANDBOX' ? 'active-sand' : ''}`}
                          onClick={() => setFormData({ ...formData, environment: 'SANDBOX' })}
                        >
                          <span className="ci-env-dot sand" />
                          <div className="ci-env-text">
                            <span className="ci-env-title">Sandbox</span>
                            <span className="ci-env-sub">Test / Staging</span>
                          </div>
                        </button>
                      </div>
                    </div>
                  </div>

                  {/* Credentials Form */}
                  <div className="ci-form-group">
                    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                      <label>API Credentials (Encrypted at Rest) *</label>
                      <button 
                        type="button" 
                        onClick={() => setUseRawJson(!useRawJson)} 
                        style={{ background: 'none', border: 'none', color: '#3b82f6', fontSize: '12px', cursor: 'pointer', textDecoration: 'underline' }}
                      >
                        {useRawJson ? 'Switch to Form Fields' : 'Switch to Raw JSON'}
                      </button>
                    </div>

                    {useRawJson ? (
                      <textarea 
                        required
                        rows={4}
                        value={formData.credentials_json}
                        onChange={e => handleRawJsonChange(e.target.value)}
                        className="ci-form-textarea"
                        style={{ fontFamily: 'monospace', fontSize: '0.85rem' }}
                      />
                    ) : (
                      <div style={{ display: 'flex', flexDirection: 'column', gap: '10px', marginTop: '4px' }}>
                        {currentProvider && Array.isArray(currentProvider.credential_fields) && currentProvider.credential_fields.length > 0 ? (
                          currentProvider.credential_fields.map(field => (
                            <div key={field.key} style={{ display: 'flex', flexDirection: 'column', gap: '4px' }}>
                              <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                                <span style={{ fontSize: '12px', fontWeight: 600, color: '#475569' }}>
                                  {field.label} {field.required && <span style={{ color: '#ef4444' }}>*</span>}
                                </span>
                                {field.description && (
                                  <span style={{ fontSize: '11px', color: '#94a3b8' }}>{field.description}</span>
                                )}
                              </div>
                              <div style={{ position: 'relative' }}>
                                <input 
                                  type={field.type === 'password' && !showSecrets[field.key] ? 'password' : 'text'}
                                  required={field.required}
                                  placeholder={field.placeholder}
                                  value={formData.credentials[field.key] || ''}
                                  onChange={e => handleCredentialChange(field.key, e.target.value)}
                                  className="ci-form-input"
                                  style={{ paddingRight: field.type === 'password' ? '36px' : '16px' }}
                                />
                                {field.type === 'password' && (
                                  <button 
                                    type="button" 
                                    onClick={() => toggleSecretVisibility(field.key)}
                                    style={{ position: 'absolute', right: '10px', top: '50%', transform: 'translateY(-50%)', background: 'none', border: 'none', color: '#94a3b8', cursor: 'pointer' }}
                                  >
                                    {showSecrets[field.key] ? <EyeOff size={16} /> : <Eye size={16} />}
                                  </button>
                                )}
                              </div>
                            </div>
                          ))
                        ) : (
                          <>
                            <div style={{ display: 'flex', flexDirection: 'column', gap: '4px' }}>
                              <span style={{ fontSize: '12px', fontWeight: 600, color: '#475569' }}>API Key / Token *</span>
                              <input 
                                type="password"
                                required
                                placeholder="Enter API Key or Access Token"
                                value={formData.credentials.api_key || ''}
                                onChange={e => handleCredentialChange('api_key', e.target.value)}
                                className="ci-form-input"
                              />
                            </div>
                            <div style={{ display: 'flex', flexDirection: 'column', gap: '4px' }}>
                              <span style={{ fontSize: '12px', fontWeight: 600, color: '#475569' }}>API Secret (Optional)</span>
                              <input 
                                type="password"
                                placeholder="Enter API Secret if required"
                                value={formData.credentials.api_secret || ''}
                                onChange={e => handleCredentialChange('api_secret', e.target.value)}
                                className="ci-form-input"
                              />
                            </div>
                          </>
                        )}
                      </div>
                    )}
                    <span className="ci-form-help">
                      🔒 Credentials are encrypted using AES-256-GCM before storage. Plaintext secrets are never stored or logged.
                    </span>
                  </div>

                  {/* Capabilities Selection */}
                  <div className="ci-form-group">
                    <label>Enabled Capabilities</label>
                    <div className="ci-caps-grid">
                      {ALL_CAPABILITIES.map(cap => {
                        const isSupported = !currentProvider || !currentProvider.supported_capabilities || 
                          currentProvider.supported_capabilities.includes(cap.key);
                        const isEnabled = !!formData.capabilities[cap.key];

                        return (
                          <label 
                            key={cap.key} 
                            className={`ci-cap-checkbox ${isEnabled && isSupported ? 'active' : ''}`}
                            style={!isSupported ? { opacity: 0.5, cursor: 'not-allowed', background: '#f8fafc' } : {}}
                          >
                            <input 
                              type="checkbox" 
                              checked={isEnabled && isSupported}
                              disabled={!isSupported}
                              onChange={() => isSupported && toggleCapability(cap.key)}
                            />
                            <div>
                              <div style={{ fontSize: '13px', fontWeight: 600 }}>{cap.label}</div>
                              {!isSupported && <div style={{ fontSize: '10px', color: '#94a3b8' }}>Not supported by carrier</div>}
                            </div>
                          </label>
                        );
                      })}
                    </div>
                  </div>

                  {/* Test Result Feedback */}
                  {testResult && (
                    <div className={`ci-test-result ${testResult.success ? 'success' : 'error'}`} style={{
                      padding: '12px 16px',
                      borderRadius: '8px',
                      fontSize: '13px',
                      background: testResult.success ? '#ecfdf5' : '#fef2f2',
                      color: testResult.success ? '#065f46' : '#991b1b',
                      border: `1px solid ${testResult.success ? '#a7f3d0' : '#fecaca'}`,
                      display: 'flex',
                      flexDirection: 'column',
                      gap: '4px'
                    }}>
                      <div style={{ display: 'flex', alignItems: 'center', gap: '8px', fontWeight: 600 }}>
                        {testResult.success ? <CheckCircle2 size={16} /> : <XCircle size={16} />}
                        {testResult.success ? 'Connection Verified' : 'Connection Failed'}
                      </div>
                      <div style={{ fontSize: '12px', marginTop: '2px' }}>{testResult.message}</div>
                      {testResult.latency && (
                        <div style={{ fontSize: '11px', color: '#059669' }}>Ping latency: {testResult.latency}ms</div>
                      )}
                    </div>
                  )}
                </>
              )}

              <div className="ci-modal-footer">
                {selectedProviderCode && (
                  <button 
                    type="button" 
                    className="ci-btn-outline" 
                    onClick={handleModalTest} 
                    disabled={isTesting || isSubmitting}
                  >
                    {isTesting ? 'Testing Adapter...' : 'Pre-flight Test'}
                  </button>
                )}
                <div style={{ flex: 1 }}></div>
                <button 
                  type="button" 
                  className="ci-btn-text" 
                  onClick={() => setShowConnectModal(false)} 
                  disabled={isSubmitting}
                >
                  Cancel
                </button>
                <button 
                  type="submit" 
                  className="ci-btn-primary" 
                  disabled={isSubmitting || isTesting || !selectedProviderCode}
                >
                  {isSubmitting ? 'Encrypting & Saving...' : 'Save & Connect'}
                </button>
              </div>
            </form>
          </div>
        </div>,
        document.body
      )}

      {/* ─────────────────────────────────────────────────────────────────── */}
      {/* Edit Carrier Modal                                                  */}
      {/* ─────────────────────────────────────────────────────────────────── */}
      {showEditModal && selectedIntegration && createPortal(
        <div className="ci-modal-overlay">
          <div className="ci-modal" style={{ maxWidth: '640px' }}>
            <div className="ci-modal-header">
              <div>
                <h2>Edit Carrier Integration</h2>
                <p style={{ fontSize: '13px', color: '#64748b', marginTop: '2px' }}>
                  Update environment, capabilities, or rotate encrypted credentials for {selectedIntegration.carrier_name}.
                </p>
              </div>
              <button className="ci-modal-close" onClick={() => setShowEditModal(false)}>×</button>
            </div>
            
            <form onSubmit={handleEditSubmit} className="ci-modal-form">
              <div className="ci-form-row">
                <div className="ci-form-group">
                  <label>Carrier</label>
                  <input 
                    type="text" 
                    value={`${selectedIntegration.carrier_name} (${selectedIntegration.carrier_scac})`} 
                    disabled 
                    className="ci-form-input" 
                    style={{ background: '#f1f5f9', cursor: 'not-allowed' }} 
                  />
                </div>

                <div className="ci-form-group">
                  <label>Environment Mode</label>
                  <div className="ci-env-selector">
                    <button
                      type="button"
                      className={`ci-env-btn ${formData.environment === 'PRODUCTION' ? 'active-prod' : ''}`}
                      onClick={() => setFormData({ ...formData, environment: 'PRODUCTION' })}
                    >
                      <span className="ci-env-dot prod" />
                      <div className="ci-env-text">
                        <span className="ci-env-title">Production</span>
                        <span className="ci-env-sub">Live DCSA APIs</span>
                      </div>
                    </button>
                    <button
                      type="button"
                      className={`ci-env-btn ${formData.environment === 'SANDBOX' ? 'active-sand' : ''}`}
                      onClick={() => setFormData({ ...formData, environment: 'SANDBOX' })}
                    >
                      <span className="ci-env-dot sand" />
                      <div className="ci-env-text">
                        <span className="ci-env-title">Sandbox</span>
                        <span className="ci-env-sub">Test / Staging</span>
                      </div>
                    </button>
                  </div>
                </div>
              </div>

              {/* Credentials Section with Safe Masking & Rotation */}
              <div className="ci-form-group">
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                  <label>Stored Credentials</label>
                  <label style={{ fontSize: '12px', color: '#3b82f6', cursor: 'pointer', display: 'flex', alignItems: 'center', gap: '6px' }}>
                    <input 
                      type="checkbox" 
                      checked={rotateCredentials} 
                      onChange={e => setRotateCredentials(e.target.checked)} 
                    />
                    Rotate / Enter New Credentials
                  </label>
                </div>

                {!rotateCredentials ? (
                  <div style={{
                    padding: '14px 18px',
                    borderRadius: '10px',
                    background: '#f8fafc',
                    border: '1px solid #e2e8f0',
                    display: 'flex',
                    flexDirection: 'column',
                    gap: '6px'
                  }}>
                    <div style={{ display: 'flex', alignItems: 'center', gap: '6px', fontSize: '13px', color: '#059669', fontWeight: 600 }}>
                      <Lock size={14} /> Credentials safely encrypted with AES-256-GCM
                    </div>
                    <div style={{ fontSize: '12px', color: '#64748b' }}>
                      Existing secrets are protected. Check "Rotate / Enter New Credentials" if you need to update them.
                    </div>
                    {selectedIntegration.credentials_mask && Object.keys(selectedIntegration.credentials_mask).length > 0 && (
                      <div style={{ marginTop: '6px', display: 'flex', flexDirection: 'column', gap: '4px' }}>
                        {Object.entries(selectedIntegration.credentials_mask).map(([k, v]) => (
                          <div key={k} style={{ display: 'flex', justifyContent: 'space-between', fontSize: '12px', color: '#475569' }}>
                            <span style={{ fontWeight: 600 }}>{k}:</span>
                            <span style={{ fontFamily: 'monospace' }}>{v}</span>
                          </div>
                        ))}
                      </div>
                    )}
                  </div>
                ) : (
                  <div style={{ display: 'flex', flexDirection: 'column', gap: '10px', marginTop: '4px' }}>
                    <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                      <span style={{ fontSize: '12px', color: '#059669', fontWeight: 600 }}>Enter new credentials to overwrite:</span>
                      <button 
                        type="button" 
                        onClick={() => setUseRawJson(!useRawJson)} 
                        style={{ background: 'none', border: 'none', color: '#3b82f6', fontSize: '12px', cursor: 'pointer', textDecoration: 'underline' }}
                      >
                        {useRawJson ? 'Use Form Fields' : 'Use Raw JSON'}
                      </button>
                    </div>

                    {useRawJson ? (
                      <textarea 
                        required
                        rows={4}
                        value={formData.credentials_json}
                        onChange={e => handleRawJsonChange(e.target.value)}
                        className="ci-form-textarea"
                        style={{ fontFamily: 'monospace', fontSize: '0.85rem' }}
                      />
                    ) : (
                      currentProvider && Array.isArray(currentProvider.credential_fields) && currentProvider.credential_fields.length > 0 ? (
                        currentProvider.credential_fields.map(field => (
                          <div key={field.key} style={{ display: 'flex', flexDirection: 'column', gap: '4px' }}>
                            <span style={{ fontSize: '12px', fontWeight: 600, color: '#475569' }}>
                              {field.label} {field.required && <span style={{ color: '#ef4444' }}>*</span>}
                            </span>
                            <input 
                              type={field.type === 'password' && !showSecrets[field.key] ? 'password' : 'text'}
                              required={field.required}
                              placeholder={field.placeholder}
                              value={formData.credentials[field.key] || ''}
                              onChange={e => handleCredentialChange(field.key, e.target.value)}
                              className="ci-form-input"
                            />
                          </div>
                        ))
                      ) : (
                        <div style={{ display: 'flex', flexDirection: 'column', gap: '4px' }}>
                          <span style={{ fontSize: '12px', fontWeight: 600, color: '#475569' }}>API Key / Token *</span>
                          <input 
                            type="password"
                            required
                            placeholder="Enter new API key"
                            value={formData.credentials.api_key || ''}
                            onChange={e => handleCredentialChange('api_key', e.target.value)}
                            className="ci-form-input"
                          />
                        </div>
                      )
                    )}
                  </div>
                )}
              </div>

              {/* Capabilities Selection */}
              <div className="ci-form-group">
                <label>Enabled Capabilities</label>
                <div className="ci-caps-grid">
                  {ALL_CAPABILITIES.map(cap => {
                    const isSupported = !currentProvider || !currentProvider.supported_capabilities || 
                      currentProvider.supported_capabilities.includes(cap.key);
                    const isEnabled = !!formData.capabilities[cap.key];

                    return (
                      <label 
                        key={cap.key} 
                        className={`ci-cap-checkbox ${isEnabled && isSupported ? 'active' : ''}`}
                        style={!isSupported ? { opacity: 0.5, cursor: 'not-allowed', background: '#f8fafc' } : {}}
                      >
                        <input 
                          type="checkbox" 
                          checked={isEnabled && isSupported}
                          disabled={!isSupported}
                          onChange={() => isSupported && toggleCapability(cap.key)}
                        />
                        <div>
                          <div style={{ fontSize: '13px', fontWeight: 600 }}>{cap.label}</div>
                          {!isSupported && <div style={{ fontSize: '10px', color: '#94a3b8' }}>Not supported</div>}
                        </div>
                      </label>
                    );
                  })}
                </div>
              </div>

              {testResult && (
                <div className={`ci-test-result ${testResult.success ? 'success' : 'error'}`} style={{
                  padding: '12px 16px',
                  borderRadius: '8px',
                  fontSize: '13px',
                  background: testResult.success ? '#ecfdf5' : '#fef2f2',
                  color: testResult.success ? '#065f46' : '#991b1b',
                  border: `1px solid ${testResult.success ? '#a7f3d0' : '#fecaca'}`
                }}>
                  {testResult.success ? '✅ ' : '❌ '} {testResult.message}
                </div>
              )}

              <div className="ci-modal-footer">
                {rotateCredentials && (
                  <button 
                    type="button" 
                    className="ci-btn-outline" 
                    onClick={handleModalTest} 
                    disabled={isTesting || isSubmitting}
                  >
                    {isTesting ? 'Testing...' : 'Test New Credentials'}
                  </button>
                )}
                <div style={{ flex: 1 }}></div>
                <button 
                  type="button" 
                  className="ci-btn-text" 
                  onClick={() => setShowEditModal(false)} 
                  disabled={isSubmitting}
                >
                  Cancel
                </button>
                <button 
                  type="submit" 
                  className="ci-btn-primary" 
                  disabled={isSubmitting || isTesting}
                >
                  {isSubmitting ? 'Saving Changes...' : 'Save Configuration'}
                </button>
              </div>
            </form>
          </div>
        </div>,
        document.body
      )}

      {/* ─────────────────────────────────────────────────────────────────── */}
      {/* Disconnect Confirmation Modal                                       */}
      {/* ─────────────────────────────────────────────────────────────────── */}
      {showDisconnectModal && selectedIntegration && createPortal(
        <div className="ci-modal-overlay">
          <div className="ci-modal" style={{ maxWidth: '480px' }}>
            <div className="ci-modal-header" style={{ borderBottom: 'none', paddingBottom: '0' }}>
              <div style={{ display: 'flex', alignItems: 'center', gap: '12px' }}>
                <div style={{
                  width: '40px',
                  height: '40px',
                  borderRadius: '10px',
                  background: '#fef2f2',
                  color: '#ef4444',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center'
                }}>
                  <AlertTriangle size={20} />
                </div>
                <div>
                  <h2 style={{ fontSize: '1.125rem' }}>Disconnect Carrier</h2>
                </div>
              </div>
              <button className="ci-modal-close" onClick={() => setShowDisconnectModal(false)}>×</button>
            </div>
            
            <div style={{ padding: '20px 32px' }}>
              <p style={{ fontSize: '14px', color: '#334155', lineHeight: 1.5, margin: 0 }}>
                Are you sure you want to disconnect <strong>{selectedIntegration.carrier_name} ({selectedIntegration.carrier_scac})</strong>?
              </p>
              
              <div style={{
                marginTop: '16px',
                padding: '12px 16px',
                borderRadius: '8px',
                background: '#f8fafc',
                border: '1px solid #e2e8f0',
                fontSize: '12px',
                color: '#64748b',
                lineHeight: 1.4
              }}>
                ℹ️ <strong>Data Preservation:</strong> Disconnecting will remove the carrier's API authentication and automated sync. All existing shipments, tracking events, quotations, and bookings in LogisticsHQ will remain preserved.
              </div>
            </div>

            <div className="ci-modal-footer" style={{ padding: '16px 32px', background: '#f8fafc' }}>
              <div style={{ flex: 1 }}></div>
              <button 
                type="button" 
                className="ci-btn-text" 
                onClick={() => setShowDisconnectModal(false)} 
                disabled={isSubmitting}
              >
                Cancel
              </button>
              <button 
                type="button" 
                style={{
                  background: '#ef4444',
                  color: 'white',
                  border: 'none',
                  borderRadius: '8px',
                  padding: '10px 18px',
                  fontSize: '13px',
                  fontWeight: 600,
                  cursor: 'pointer'
                }}
                onClick={handleConfirmDisconnect}
                disabled={isSubmitting}
              >
                {isSubmitting ? 'Disconnecting...' : 'Confirm Disconnect'}
              </button>
            </div>
          </div>
        </div>,
        document.body
      )}

      {/* ─────────────────────────────────────────────────────────────────── */}
      {/* Sync History Modal / Drawer                                         */}
      {/* ─────────────────────────────────────────────────────────────────── */}
      {showHistoryModal && historyCarrier && createPortal(
        <div className="ci-modal-overlay" onClick={() => setShowHistoryModal(false)}>
          <div className="ci-modal" style={{ maxWidth: '850px', width: '90%' }} onClick={e => e.stopPropagation()}>
            <div className="ci-modal-header">
              <div style={{ display: 'flex', alignItems: 'center', gap: '12px' }}>
                <div style={{
                  width: '40px',
                  height: '40px',
                  borderRadius: '10px',
                  background: '#eff6ff',
                  color: '#2563eb',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center'
                }}>
                  <History size={20} />
                </div>
                <div>
                  <h2 style={{ fontSize: '1.125rem', margin: 0 }}>
                    Sync History — {historyCarrier.carrier_name || historyCarrier.carrier_scac}
                  </h2>
                  <p style={{ margin: '2px 0 0 0', fontSize: '12px', color: '#64748b' }}>
                    SCAC: <strong>{historyCarrier.carrier_scac}</strong> • Environment: <strong>{historyCarrier.environment || 'PRODUCTION'}</strong> • Total Runs: <strong>{historyTotal}</strong>
                  </p>
                </div>
              </div>
              <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                <button
                  type="button"
                  className="ci-btn-outline"
                  style={{ fontSize: '12px', padding: '6px 12px', display: 'inline-flex', alignItems: 'center', gap: '6px' }}
                  onClick={() => handleSyncCarrier(historyCarrier.id, historyCarrier.carrier_name || historyCarrier.carrier_scac)}
                  disabled={syncingId === historyCarrier.id}
                >
                  <RefreshCw size={13} className={syncingId === historyCarrier.id ? 'spin' : ''} />
                  <span>{syncingId === historyCarrier.id ? 'Syncing...' : 'Trigger Sync Now'}</span>
                </button>
                <button className="ci-modal-close" onClick={() => setShowHistoryModal(false)}>×</button>
              </div>
            </div>

            <div style={{ padding: '24px 32px', maxHeight: '65vh', overflowY: 'auto' }}>
              {isLoadingHistory ? (
                <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', gap: '8px', padding: '48px 0', color: '#64748b' }}>
                  <RefreshCw size={20} className="spin" />
                  <span>Loading synchronization history records...</span>
                </div>
              ) : historyItems.length === 0 ? (
                <div style={{ textAlign: 'center', padding: '48px 20px', color: '#64748b' }}>
                  <History size={36} color="#94a3b8" style={{ marginBottom: '12px' }} />
                  <h4 style={{ margin: '0 0 6px 0', fontSize: '15px', color: '#1e293b' }}>No Sync Jobs Recorded</h4>
                  <p style={{ margin: 0, fontSize: '13px' }}>
                    Click "Trigger Sync Now" to synchronize active tracking milestones and space booking allocations with {historyCarrier.carrier_name || historyCarrier.carrier_scac}.
                  </p>
                </div>
              ) : (
                <div style={{ display: 'flex', flexDirection: 'column', gap: '12px' }}>
                  {historyItems.map(job => {
                    const isSuccess = job.status === 'SUCCESS';
                    const isPartial = job.status === 'PARTIAL_SUCCESS';
                    const isFailed = job.status === 'FAILED';
                    const isRunning = job.status === 'RUNNING';

                    const statusBg = isSuccess ? '#ecfdf5' : isPartial ? '#fffbeb' : isFailed ? '#fef2f2' : '#eff6ff';
                    const statusText = isSuccess ? '#059669' : isPartial ? '#b45309' : isFailed ? '#dc2626' : '#2563eb';
                    const statusBorder = isSuccess ? '#a7f3d0' : isPartial ? '#fde68a' : isFailed ? '#fecaca' : '#bfdbfe';

                    let formattedStart = job.started_at ? new Date(job.started_at).toLocaleString([], { dateStyle: 'short', timeStyle: 'medium' }) : '-';

                    return (
                      <div key={job.id} style={{
                        border: '1px solid #e2e8f0',
                        borderRadius: '10px',
                        padding: '16px',
                        background: '#ffffff',
                        boxShadow: '0 1px 3px rgba(0,0,0,0.03)'
                      }}>
                        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: '10px' }}>
                          <div style={{ display: 'flex', alignItems: 'center', gap: '10px' }}>
                            <span style={{
                              padding: '3px 8px',
                              borderRadius: '6px',
                              fontSize: '11.5px',
                              fontWeight: 700,
                              background: '#f1f5f9',
                              color: '#334155'
                            }}>
                              {job.operation || 'FULL_SYNC'}
                            </span>
                            <span style={{
                              padding: '3px 10px',
                              borderRadius: '12px',
                              fontSize: '11.5px',
                              fontWeight: 600,
                              background: statusBg,
                              color: statusText,
                              border: `1px solid ${statusBorder}`,
                              display: 'inline-flex',
                              alignItems: 'center',
                              gap: '4px'
                            }}>
                              {isRunning && <RefreshCw size={10} className="spin" />}
                              {job.status}
                            </span>
                            <span style={{ fontSize: '12px', color: '#64748b' }}>
                              Started: {formattedStart}
                            </span>
                          </div>
                          {job.duration_ms !== undefined && (
                            <span style={{ fontSize: '12px', color: '#475569', fontWeight: 500 }}>
                              ⏱️ {job.duration_ms}ms
                            </span>
                          )}
                        </div>

                        <div style={{
                          display: 'grid',
                          gridTemplateColumns: 'repeat(4, 1fr)',
                          gap: '8px',
                          background: '#f8fafc',
                          padding: '10px 14px',
                          borderRadius: '8px',
                          fontSize: '12px'
                        }}>
                          <div>
                            <span style={{ color: '#64748b', display: 'block', fontSize: '11px' }}>Processed</span>
                            <strong style={{ color: '#0f172a', fontSize: '13px' }}>{job.records_processed ?? 0}</strong>
                          </div>
                          <div>
                            <span style={{ color: '#64748b', display: 'block', fontSize: '11px' }}>Created</span>
                            <strong style={{ color: '#059669', fontSize: '13px' }}>+{job.records_created ?? 0}</strong>
                          </div>
                          <div>
                            <span style={{ color: '#64748b', display: 'block', fontSize: '11px' }}>Updated</span>
                            <strong style={{ color: '#2563eb', fontSize: '13px' }}>{job.records_updated ?? 0}</strong>
                          </div>
                          <div>
                            <span style={{ color: '#64748b', display: 'block', fontSize: '11px' }}>Failed</span>
                            <strong style={{ color: job.records_failed > 0 ? '#ef4444' : '#64748b', fontSize: '13px' }}>{job.records_failed ?? 0}</strong>
                          </div>
                        </div>

                        {job.error_message && (
                          <div style={{
                            marginTop: '10px',
                            padding: '8px 12px',
                            borderRadius: '6px',
                            background: '#fef2f2',
                            border: '1px solid #fecaca',
                            fontSize: '12px',
                            color: '#b91c1c'
                          }}>
                            <strong>Error:</strong> {job.error_message}
                          </div>
                        )}

                        {job.correlation_id && (
                          <div style={{ marginTop: '8px', fontSize: '11px', color: '#94a3b8' }}>
                            Correlation ID: <code>{job.correlation_id}</code>
                          </div>
                        )}
                      </div>
                    );
                  })}
                </div>
              )}
            </div>

            <div className="ci-modal-footer" style={{ background: '#f8fafc', padding: '14px 32px' }}>
              <div style={{ flex: 1 }}></div>
              <button
                type="button"
                className="ci-btn-primary"
                onClick={() => setShowHistoryModal(false)}
              >
                Close Logs
              </button>
            </div>
          </div>
        </div>,
        document.body
      )}

      {/* ─────────────────────────────────────────────────────────────────── */}
      {/* Health Diagnostics Modal                                            */}
      {/* ─────────────────────────────────────────────────────────────────── */}
      {showHealthModal && healthCarrier && createPortal(
        <div className="ci-modal-overlay" onClick={() => setShowHealthModal(false)}>
          <div className="ci-modal" style={{ maxWidth: '620px', width: '90%' }} onClick={e => e.stopPropagation()}>
            <div className="ci-modal-header">
              <div style={{ display: 'flex', alignItems: 'center', gap: '12px' }}>
                <div style={{
                  width: '40px',
                  height: '40px',
                  borderRadius: '10px',
                  background: '#ecfdf5',
                  color: '#059669',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center'
                }}>
                  <Activity size={20} />
                </div>
                <div>
                  <h2 style={{ fontSize: '1.125rem', margin: 0 }}>
                    Health Diagnostics — {healthCarrier.carrier_name || healthCarrier.carrier_scac}
                  </h2>
                  <p style={{ margin: '2px 0 0 0', fontSize: '12px', color: '#64748b' }}>
                    Real-time adapter connectivity and failure threshold telemetry
                  </p>
                </div>
              </div>
              <button className="ci-modal-close" onClick={() => setShowHealthModal(false)}>×</button>
            </div>

            <div style={{ padding: '24px 32px' }}>
              {isLoadingHealth ? (
                <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', gap: '8px', padding: '36px 0', color: '#64748b' }}>
                  <RefreshCw size={18} className="spin" />
                  <span>Evaluating integration health metrics...</span>
                </div>
              ) : healthData ? (
                <div style={{ display: 'flex', flexDirection: 'column', gap: '16px' }}>
                  <div style={{
                    padding: '16px',
                    borderRadius: '10px',
                    background: healthData.health_state === 'HEALTHY' ? '#ecfdf5' : healthData.health_state === 'ATTENTION' ? '#fffbeb' : '#fef2f2',
                    border: `1px solid ${healthData.health_state === 'HEALTHY' ? '#a7f3d0' : healthData.health_state === 'ATTENTION' ? '#fde68a' : '#fecaca'}`
                  }}>
                    <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: '6px' }}>
                      <span style={{ fontSize: '14px', fontWeight: 700, color: healthData.health_state === 'HEALTHY' ? '#065f46' : healthData.health_state === 'ATTENTION' ? '#92400e' : '#991b1b' }}>
                        Status: {healthData.health_state}
                      </span>
                      <span style={{ fontSize: '12px', fontWeight: 600, color: '#475569' }}>
                        Failures: {healthData.consecutive_failures ?? 0} consecutive
                      </span>
                    </div>
                    <p style={{ margin: 0, fontSize: '13px', color: '#334155', lineHeight: 1.4 }}>
                      {healthData.reason || (healthData.health_state === 'HEALTHY' ? 'All sync and connectivity checks operating normally within thresholds.' : 'Adapter requires attention.')}
                    </p>
                  </div>

                  <div style={{
                    display: 'grid',
                    gridTemplateColumns: '1fr 1fr',
                    gap: '12px'
                  }}>
                    <div style={{ padding: '12px 16px', background: '#f8fafc', borderRadius: '8px', border: '1px solid #e2e8f0' }}>
                      <span style={{ fontSize: '11px', color: '#64748b', display: 'block' }}>Connection Status</span>
                      <strong style={{ fontSize: '13px', color: '#0f172a' }}>{healthData.connection_status || 'CONNECTED'}</strong>
                    </div>
                    <div style={{ padding: '12px 16px', background: '#f8fafc', borderRadius: '8px', border: '1px solid #e2e8f0' }}>
                      <span style={{ fontSize: '11px', color: '#64748b', display: 'block' }}>Active State</span>
                      <strong style={{ fontSize: '13px', color: healthData.is_enabled ? '#059669' : '#dc2626' }}>
                        {healthData.is_enabled ? 'Enabled' : 'Disabled'}
                      </strong>
                    </div>
                    <div style={{ padding: '12px 16px', background: '#f8fafc', borderRadius: '8px', border: '1px solid #e2e8f0' }}>
                      <span style={{ fontSize: '11px', color: '#64748b', display: 'block' }}>Last Successful Sync</span>
                      <strong style={{ fontSize: '12.5px', color: '#0f172a' }}>
                        {healthData.last_success_at ? new Date(healthData.last_success_at).toLocaleString([], { dateStyle: 'short', timeStyle: 'short' }) : 'Never'}
                      </strong>
                    </div>
                    <div style={{ padding: '12px 16px', background: '#f8fafc', borderRadius: '8px', border: '1px solid #e2e8f0' }}>
                      <span style={{ fontSize: '11px', color: '#64748b', display: 'block' }}>Last Failed Sync</span>
                      <strong style={{ fontSize: '12.5px', color: healthData.last_failure_at ? '#dc2626' : '#64748b' }}>
                        {healthData.last_failure_at ? new Date(healthData.last_failure_at).toLocaleString([], { dateStyle: 'short', timeStyle: 'short' }) : 'None'}
                      </strong>
                    </div>
                  </div>

                  {healthData.last_error && (
                    <div style={{
                      padding: '12px 16px',
                      borderRadius: '8px',
                      background: '#fef2f2',
                      border: '1px solid #fecaca',
                      fontSize: '12px',
                      color: '#991b1b'
                    }}>
                      <strong>Last Error Trace:</strong>
                      <div style={{ marginTop: '4px', fontFamily: 'monospace', fontSize: '11.5px', wordBreak: 'break-all' }}>
                        {healthData.last_error}
                      </div>
                    </div>
                  )}
                </div>
              ) : null}
            </div>

            <div className="ci-modal-footer" style={{ background: '#f8fafc', padding: '14px 32px' }}>
              <button
                type="button"
                className="ci-btn-outline"
                onClick={() => { setShowHealthModal(false); handleOpenHistory(healthCarrier); }}
              >
                View Sync Logs
              </button>
              <div style={{ flex: 1 }}></div>
              <button
                type="button"
                className="ci-btn-primary"
                onClick={() => setShowHealthModal(false)}
              >
                Done
              </button>
            </div>
          </div>
        </div>,
        document.body
      )}
    </div>
  );
}

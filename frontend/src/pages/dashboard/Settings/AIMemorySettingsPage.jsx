import React, { useState, useEffect, useCallback } from 'react';
import { 
  Brain, 
  Plus, 
  Trash2, 
  Edit3, 
  CheckCircle2, 
  AlertCircle, 
  Shield, 
  Sparkles, 
  Clock, 
  Power, 
  RefreshCw, 
  HelpCircle,
  FileText,
  Building,
  User,
  Sliders
} from 'lucide-react';
import toast from 'react-hot-toast';
import { useRBAC } from '../../../context/RBACContext';
import memoryService from '../../../services/memoryService';
import './AIMemorySettingsPage.css';

export default function AIMemorySettingsPage() {
  const { userRole } = useRBAC();
  const isAdmin = ['SUPER_ADMIN', 'ADMIN', 'ORGANIZATION_ADMIN', 'MANAGER'].includes(
    typeof userRole === 'string' ? userRole.toUpperCase() : userRole?.name?.toUpperCase()
  );

  // Active Tab
  const [activeTab, setActiveTab] = useState('PERSONAL'); // 'PERSONAL' | 'ORGANIZATION' | 'SETTINGS' | 'AUDIT'

  // Data States
  const [stats, setStats] = useState(null);
  const [memories, setMemories] = useState([]);
  const [auditEvents, setAuditEvents] = useState([]);
  const [loading, setLoading] = useState(true);
  const [filterType, setFilterType] = useState('');
  const [filterStatus, setFilterStatus] = useState('');
  const [searchTerm, setSearchTerm] = useState('');

  // Settings Form State
  const [settingsForm, setSettingsForm] = useState({
    personalization_enabled: true,
    preferred_response_style: 'CONCISE',
    preferred_summary_depth: 'STANDARD',
    preferred_currency: 'USD',
    preferred_timezone: 'UTC',
    preferred_date_format: 'YYYY-MM-DD',
    preferred_default_module: 'DASHBOARD',
    explanation_level: 'STANDARD'
  });
  const [savingSettings, setSavingSettings] = useState(false);

  // Modal States
  const [showModal, setShowModal] = useState(false);
  const [modalMode, setModalMode] = useState('CREATE'); // 'CREATE' | 'EDIT'
  const [selectedItem, setSelectedItem] = useState(null);
  const [modalForm, setModalForm] = useState({
    title: '',
    content: '',
    memory_type: 'RESPONSE_STYLE',
    scope: 'USER',
    expires_in_days: 0,
    explicitly_confirmed: true
  });
  const [showClearConfirm, setShowClearConfirm] = useState(false);
  const [submittingModal, setSubmittingModal] = useState(false);

  // Load Data
  const loadData = useCallback(async () => {
    setLoading(true);
    try {
      const [statsRes, memRes, settingsRes, auditRes] = await Promise.all([
        memoryService.getStats().catch(() => null),
        memoryService.listMemories({
          scope: activeTab === 'ORGANIZATION' ? 'ORGANIZATION' : 'USER',
          memory_type: filterType,
          status: filterStatus,
          search: searchTerm,
          limit: 50
        }).catch(() => ({ data: [] })),
        memoryService.getUserSettings().catch(() => null),
        memoryService.listAuditEvents({ limit: 30 }).catch(() => ({ data: [] }))
      ]);

      if (statsRes?.data) setStats(statsRes.data);
      if (memRes?.data) setMemories(memRes.data);
      if (settingsRes?.data) setSettingsForm(settingsRes.data);
      if (auditRes?.data) setAuditEvents(auditRes.data);
    } catch (err) {
      console.error('Error loading memory data:', err);
      toast.error('Failed to load memory configuration');
    } finally {
      setLoading(false);
    }
  }, [activeTab, filterType, filterStatus, searchTerm]);

  useEffect(() => {
    loadData();
  }, [loadData]);

  // Toggle Master Personalization
  const handleTogglePersonalization = async () => {
    const nextVal = !settingsForm.personalization_enabled;
    try {
      await memoryService.togglePersonalization(nextVal);
      setSettingsForm(prev => ({ ...prev, personalization_enabled: nextVal }));
      toast.success(`Personalization ${nextVal ? 'enabled' : 'disabled'}`);
      loadData();
    } catch (err) {
      toast.error('Failed to update personalization toggle');
    }
  };

  // Save Settings
  const handleSaveSettings = async (e) => {
    e.preventDefault();
    setSavingSettings(true);
    try {
      await memoryService.updateUserSettings(settingsForm);
      toast.success('AI preferences saved successfully');
      loadData();
    } catch (err) {
      toast.error(err.response?.data?.message || 'Failed to update preferences');
    } finally {
      setSavingSettings(false);
    }
  };

  // Open Create Modal
  const handleOpenCreate = () => {
    setModalMode('CREATE');
    setModalForm({
      title: '',
      content: '',
      memory_type: activeTab === 'ORGANIZATION' ? 'TERMINOLOGY' : 'RESPONSE_STYLE',
      scope: activeTab === 'ORGANIZATION' ? 'ORGANIZATION' : 'USER',
      expires_in_days: 0,
      explicitly_confirmed: true
    });
    setShowModal(true);
  };

  // Open Edit Modal
  const handleOpenEdit = (item) => {
    setModalMode('EDIT');
    setSelectedItem(item);
    setModalForm({
      title: item.title,
      content: item.content,
      memory_type: item.memory_type,
      scope: item.scope,
      expires_in_days: 0,
      explicitly_confirmed: true
    });
    setShowModal(true);
  };

  // Submit Modal
  const handleSubmitModal = async (e) => {
    e.preventDefault();
    if (!modalForm.explicitly_confirmed) {
      toast.error('Explicit confirmation is required to save memory');
      return;
    }

    setSubmittingModal(true);
    try {
      if (modalMode === 'CREATE') {
        const payload = {
          scope: modalForm.scope,
          memory_type: modalForm.memory_type,
          title: modalForm.title,
          content: modalForm.content,
          explicitly_confirmed: true
        };
        if (modalForm.expires_in_days > 0) {
          const d = new Date();
          d.setDate(d.getDate() + Number(modalForm.expires_in_days));
          payload.expires_at = d.toISOString();
        }
        await memoryService.createMemory(payload);
        toast.success('Memory item explicitly saved');
      } else {
        await memoryService.updateMemory(selectedItem.id, {
          title: modalForm.title,
          content: modalForm.content
        });
        toast.success('Memory item updated');
      }
      setShowModal(false);
      loadData();
    } catch (err) {
      const msg = err.response?.data?.message || err.message || 'Failed to save memory item';
      toast.error(msg);
    } finally {
      setSubmittingModal(false);
    }
  };

  // Disable / Enable
  const handleToggleStatus = async (item) => {
    try {
      if (item.status === 'ACTIVE') {
        await memoryService.disableMemory(item.id);
        toast.success('Memory item disabled');
      } else {
        await memoryService.enableMemory(item.id);
        toast.success('Memory item re-enabled');
      }
      loadData();
    } catch (err) {
      toast.error('Failed to change item status');
    }
  };

  // Delete
  const handleDelete = async (item) => {
    if (!window.confirm(`Are you sure you want to delete "${item.title}"?`)) return;
    try {
      await memoryService.deleteMemory(item.id);
      toast.success('Memory item removed');
      loadData();
    } catch (err) {
      toast.error('Failed to delete memory item');
    }
  };

  // Clear Personal
  const handleClearPersonal = async () => {
    try {
      const res = await memoryService.clearPersonalMemories();
      toast.success(`Cleared ${res?.cleared || 0} personal memory items`);
      setShowClearConfirm(false);
      loadData();
    } catch (err) {
      toast.error('Failed to clear personal memories');
    }
  };

  return (
    <div className="ai-memory-page">
      {/* Header */}
      <div className="ai-memory-header">
        <div className="ai-memory-header-info">
          <h1>
            <Brain size={24} color="#2563eb" />
            AI Memory & Personalization
          </h1>
          <p>
            User-controlled memory, explicit preferences, and operational terminology. Authoritative data always takes precedence.
          </p>
        </div>
        <div className="ai-memory-header-actions">
          <button 
            type="button" 
            className="btn-danger-outline"
            onClick={() => setShowClearConfirm(true)}
            title="Clear all personal preferences and memories"
          >
            <Trash2 size={16} />
            Clear Personal Memory
          </button>
          <button 
            type="button" 
            className="btn-primary-memory"
            onClick={handleOpenCreate}
          >
            <Plus size={16} />
            New Memory Item
          </button>
        </div>
      </div>

      {/* KPI Cards */}
      <div className="ai-memory-kpis">
        <div className="memory-kpi-card">
          <div className="memory-kpi-title">
            Personalization
            <Sliders size={14} />
          </div>
          <div className="toggle-switch-container">
            <label className="toggle-switch">
              <input 
                type="checkbox" 
                checked={settingsForm.personalization_enabled}
                onChange={handleTogglePersonalization}
              />
              <span className="toggle-slider"></span>
            </label>
            <span style={{ fontSize: '0.875rem', fontWeight: 600, color: settingsForm.personalization_enabled ? '#10b981' : '#64748b' }}>
              {settingsForm.personalization_enabled ? 'Active' : 'Disabled'}
            </span>
          </div>
          <div className="memory-kpi-subtext" style={{ marginTop: '8px' }}>
            {settingsForm.personalization_enabled ? 'Personalizing AI responses' : 'Default system defaults applied'}
          </div>
        </div>

        <div className="memory-kpi-card">
          <div className="memory-kpi-title">
            Personal Memories
            <User size={14} />
          </div>
          <div className="memory-kpi-value">
            {stats?.active_personal_count ?? 0}
            <span className="memory-kpi-subtext">Active</span>
          </div>
          <div className="memory-kpi-subtext">Scoped strictly to your user account</div>
        </div>

        <div className="memory-kpi-card">
          <div className="memory-kpi-title">
            Organization Policies
            <Building size={14} />
          </div>
          <div className="memory-kpi-value">
            {stats?.active_org_count ?? 0}
            <span className="memory-kpi-subtext">Org-wide</span>
          </div>
          <div className="memory-kpi-subtext">Shared business facts & terminology</div>
        </div>

        <div className="memory-kpi-card">
          <div className="memory-kpi-title">
            Auditable Events
            <Shield size={14} />
          </div>
          <div className="memory-kpi-value">
            {stats?.total_audit_events ?? auditEvents.length}
            <span className="memory-kpi-subtext">Logged</span>
          </div>
          <div className="memory-kpi-subtext">Immutable audit log transitions</div>
        </div>
      </div>

      {/* Tabs */}
      <div className="ai-memory-tabs">
        <button 
          className={`memory-tab-btn ${activeTab === 'PERSONAL' ? 'active' : ''}`}
          onClick={() => setActiveTab('PERSONAL')}
        >
          <User size={16} />
          Personal Memories
          <span className="memory-tab-badge">{stats?.active_personal_count ?? 0}</span>
        </button>

        <button 
          className={`memory-tab-btn ${activeTab === 'ORGANIZATION' ? 'active' : ''}`}
          onClick={() => setActiveTab('ORGANIZATION')}
        >
          <Building size={16} />
          Organization Terminology & Facts
          <span className="memory-tab-badge">{stats?.active_org_count ?? 0}</span>
        </button>

        <button 
          className={`memory-tab-btn ${activeTab === 'SETTINGS' ? 'active' : ''}`}
          onClick={() => setActiveTab('SETTINGS')}
        >
          <Sliders size={16} />
          AI Preferences & Formats
        </button>

        <button 
          className={`memory-tab-btn ${activeTab === 'AUDIT' ? 'active' : ''}`}
          onClick={() => setActiveTab('AUDIT')}
        >
          <Shield size={16} />
          Audit & Security Trail
        </button>
      </div>

      {/* TAB CONTENT: PERSONAL or ORGANIZATION MEMORIES */}
      {(activeTab === 'PERSONAL' || activeTab === 'ORGANIZATION') && (
        <>
          {/* Filter Strip */}
          <div className="memory-filter-strip">
            <input 
              type="text" 
              placeholder="Search memory title or content..."
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              className="memory-search-input"
            />

            <select 
              value={filterType} 
              onChange={(e) => setFilterType(e.target.value)}
              className="memory-select-filter"
            >
              <option value="">All Memory Types</option>
              <option value="RESPONSE_STYLE">Response Style</option>
              <option value="SUMMARY_PREFERENCE">Summary Preference</option>
              <option value="TERMINOLOGY">Terminology</option>
              <option value="BUSINESS_FACT">Business Fact</option>
              <option value="MODULE_PREFERENCE">Module Preference</option>
              <option value="EXPLANATION_DEPTH">Explanation Depth</option>
            </select>

            <select 
              value={filterStatus} 
              onChange={(e) => setFilterStatus(e.target.value)}
              className="memory-select-filter"
            >
              <option value="">All Statuses</option>
              <option value="ACTIVE">Active</option>
              <option value="DISABLED">Disabled</option>
              <option value="EXPIRED">Expired</option>
            </select>

            <button 
              type="button" 
              onClick={loadData}
              className="btn-icon-action"
              title="Refresh list"
            >
              <RefreshCw size={14} />
            </button>
          </div>

          {/* List */}
          {loading ? (
            <div style={{ textAlign: 'center', padding: '40px', color: '#64748b' }}>
              Loading memory items...
            </div>
          ) : memories.length === 0 ? (
            <div style={{ background: '#ffffff', padding: '48px', borderRadius: '12px', border: '1px solid #e2e8f0', textAlign: 'center' }}>
              <Brain size={40} color="#94a3b8" style={{ marginBottom: '12px' }} />
              <h3 style={{ margin: '0 0 6px 0', fontSize: '1.1rem', color: '#0f172a' }}>No memories saved</h3>
              <p style={{ color: '#64748b', fontSize: '0.875rem', maxWidth: '440px', margin: '0 auto 18px auto' }}>
                {activeTab === 'ORGANIZATION' 
                  ? 'No organization-wide business facts or terminology have been defined yet.' 
                  : 'You have not saved any personal AI preferences or instructions yet.'}
              </p>
              <button 
                type="button" 
                className="btn-primary-memory" 
                onClick={handleOpenCreate}
                style={{ margin: '0 auto' }}
              >
                <Plus size={16} />
                Save New Memory
              </button>
            </div>
          ) : (
            <div className="memory-list-container">
              {memories.map((item) => (
                <div key={item.id} className="memory-item-card">
                  <div className="memory-item-top">
                    <div className="memory-item-title-wrap">
                      <span className={`memory-badge ${item.scope === 'ORGANIZATION' ? 'badge-scope-org' : 'badge-scope-user'}`}>
                        {item.scope}
                      </span>
                      <h4 className="memory-item-title">{item.title}</h4>
                      <span className={`memory-badge badge-status-${item.status.toLowerCase()}`}>
                        {item.status}
                      </span>
                    </div>

                    <div className="memory-item-actions">
                      <button 
                        type="button"
                        className="btn-icon-action"
                        onClick={() => handleToggleStatus(item)}
                        title={item.status === 'ACTIVE' ? 'Disable memory' : 'Re-enable memory'}
                      >
                        <Power size={13} />
                        {item.status === 'ACTIVE' ? 'Disable' : 'Enable'}
                      </button>

                      <button 
                        type="button"
                        className="btn-icon-action"
                        onClick={() => handleOpenEdit(item)}
                        title="Edit memory"
                      >
                        <Edit3 size={13} />
                        Edit
                      </button>

                      <button 
                        type="button"
                        className="btn-icon-action btn-icon-danger"
                        onClick={() => handleDelete(item)}
                        title="Delete memory"
                      >
                        <Trash2 size={13} />
                      </button>
                    </div>
                  </div>

                  <div className="memory-item-content">
                    {item.content}
                  </div>

                  <div className="memory-item-meta">
                    <span><strong>Type:</strong> {item.memory_type}</span>
                    <span><strong>Source:</strong> {item.source_type}</span>
                    <span><strong>Updated:</strong> {new Date(item.updated_at).toLocaleDateString()}</span>
                    {item.expires_at && (
                      <span style={{ color: '#b91c1c' }}>
                        <Clock size={12} style={{ display: 'inline', marginRight: '4px' }} />
                        Expires: {new Date(item.expires_at).toLocaleDateString()}
                      </span>
                    )}
                    {item.created_by && <span>By: {item.created_by}</span>}
                  </div>
                </div>
              ))}
            </div>
          )}
        </>
      )}

      {/* TAB CONTENT: SETTINGS & FORMATS */}
      {activeTab === 'SETTINGS' && (
        <form onSubmit={handleSaveSettings} className="preferences-form-grid">
          <div className="pref-form-group">
            <label>Preferred Response Style</label>
            <select 
              value={settingsForm.preferred_response_style}
              onChange={(e) => setSettingsForm({ ...settingsForm, preferred_response_style: e.target.value })}
            >
              <option value="CONCISE">Concise — Direct bullet points, minimal elaboration</option>
              <option value="DETAILED">Detailed — Comprehensive context & explanations</option>
              <option value="EXECUTIVE">Executive — High-level KPI and decision focus</option>
            </select>
            <span className="pref-form-hint">Governs how AI summaries and assistants format narrative responses.</span>
          </div>

          <div className="pref-form-group">
            <label>Preferred Summary Depth</label>
            <select 
              value={settingsForm.preferred_summary_depth}
              onChange={(e) => setSettingsForm({ ...settingsForm, preferred_summary_depth: e.target.value })}
            >
              <option value="BRIEF">Brief — 1 to 2 sentences</option>
              <option value="STANDARD">Standard — Balanced 3 to 4 points</option>
              <option value="COMPREHENSIVE">Comprehensive — Deep multi-section analysis</option>
            </select>
            <span className="pref-form-hint">Determines default brevity across shipment, RFQ, and invoice summaries.</span>
          </div>

          <div className="pref-form-group">
            <label>Preferred Currency Display</label>
            <select 
              value={settingsForm.preferred_currency}
              onChange={(e) => setSettingsForm({ ...settingsForm, preferred_currency: e.target.value })}
            >
              <option value="USD">USD ($) — US Dollar</option>
              <option value="EUR">EUR (€) — Euro</option>
              <option value="GBP">GBP (£) — British Pound</option>
              <option value="INR">INR (₹) — Indian Rupee</option>
              <option value="SGD">SGD (S$) — Singapore Dollar</option>
              <option value="AED">AED (د.إ) — UAE Dirham</option>
            </select>
            <span className="pref-form-hint">Display preference for presentation; does not alter underlying ledger values.</span>
          </div>

          <div className="pref-form-group">
            <label>Preferred Date Format</label>
            <select 
              value={settingsForm.preferred_date_format}
              onChange={(e) => setSettingsForm({ ...settingsForm, preferred_date_format: e.target.value })}
            >
              <option value="YYYY-MM-DD">YYYY-MM-DD (e.g. 2026-09-08)</option>
              <option value="DD/MM/YYYY">DD/MM/YYYY (e.g. 08/09/2026)</option>
              <option value="MM/DD/YYYY">MM/DD/YYYY (e.g. 09/08/2026)</option>
            </select>
            <span className="pref-form-hint">Preferred date notation in AI-generated drafts and reports.</span>
          </div>

          <div className="pref-form-group">
            <label>Default Starting Module</label>
            <select 
              value={settingsForm.preferred_default_module}
              onChange={(e) => setSettingsForm({ ...settingsForm, preferred_default_module: e.target.value })}
            >
              <option value="DASHBOARD">Dashboard / Mission Control</option>
              <option value="SHIPMENTS">Shipments & Tracking</option>
              <option value="INVOICES">Invoices & Receivables</option>
              <option value="RFQS">RFQs & Quotations</option>
            </select>
            <span className="pref-form-hint">Directs AI navigation shortcuts and context suggestions.</span>
          </div>

          <div className="pref-form-group">
            <label>Explanation Level</label>
            <select 
              value={settingsForm.explanation_level}
              onChange={(e) => setSettingsForm({ ...settingsForm, explanation_level: e.target.value })}
            >
              <option value="DIRECT">Direct — Just the facts, no explanatory preamble</option>
              <option value="STANDARD">Standard — Factual with brief reasoning</option>
              <option value="IN_DEPTH">In-Depth — Includes calculation steps & formula citations</option>
            </select>
            <span className="pref-form-hint">Degree of operational explanation provided in AI Copilot responses.</span>
          </div>

          <div style={{ gridColumn: '1 / -1', marginTop: '12px', display: 'flex', justifyContent: 'flex-end' }}>
            <button 
              type="submit" 
              className="btn-primary-memory"
              disabled={savingSettings}
            >
              {savingSettings ? 'Saving...' : 'Save AI Preferences'}
            </button>
          </div>
        </form>
      )}

      {/* TAB CONTENT: AUDIT TRAIL */}
      {activeTab === 'AUDIT' && (
        <div style={{ background: '#ffffff', borderRadius: '12px', border: '1px solid #e2e8f0', overflow: 'hidden' }}>
          <table className="memory-audit-table">
            <thead>
              <tr>
                <th>Timestamp</th>
                <th>Event Type</th>
                <th>Scope</th>
                <th>Actor</th>
                <th>Details</th>
              </tr>
            </thead>
            <tbody>
              {auditEvents.length === 0 ? (
                <tr>
                  <td colSpan={5} style={{ textAlign: 'center', padding: '24px', color: '#64748b' }}>
                    No audit records recorded yet.
                  </td>
                </tr>
              ) : (
                auditEvents.map((evt) => (
                  <tr key={evt.id}>
                    <td>{new Date(evt.created_at).toLocaleString()}</td>
                    <td>
                      <span className="memory-badge" style={{ background: '#eff6ff', color: '#1e40af' }}>
                        {evt.event_type}
                      </span>
                    </td>
                    <td>{evt.scope}</td>
                    <td>{evt.actor_name}</td>
                    <td>{evt.details || '—'}</td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
      )}

      {/* CREATE / EDIT MODAL */}
      {showModal && (
        <div className="modal-overlay" onClick={() => setShowModal(false)}>
          <div className="memory-modal-card" onClick={(e) => e.stopPropagation()}>
            <div className="memory-modal-header">
              <h3>{modalMode === 'CREATE' ? 'Save New Memory' : 'Edit Memory Item'}</h3>
              <button 
                type="button" 
                onClick={() => setShowModal(false)}
                style={{ background: 'none', border: 'none', cursor: 'pointer', fontSize: '1.25rem', color: '#64748b' }}
              >
                ×
              </button>
            </div>

            <form onSubmit={handleSubmitModal}>
              <div className="memory-modal-body">
                <div className="safety-notice-banner">
                  <Shield size={18} style={{ flexShrink: 0, marginTop: '2px' }} />
                  <div>
                    <strong>Zero-Bypass Safety Shield:</strong> LogisticsHQ screens all memory for credentials, API tokens, payment credentials, and prompt injections. Passwords or sensitive information will be immediately rejected.
                  </div>
                </div>

                <div className="pref-form-group">
                  <label>Title *</label>
                  <input 
                    type="text" 
                    placeholder="e.g. Concise Ocean Freight Summaries"
                    value={modalForm.title}
                    onChange={(e) => setModalForm({ ...modalForm, title: e.target.value })}
                    required
                  />
                </div>

                <div className="pref-form-group">
                  <label>Memory Type *</label>
                  <select 
                    value={modalForm.memory_type}
                    onChange={(e) => setModalForm({ ...modalForm, memory_type: e.target.value })}
                    required
                  >
                    <option value="RESPONSE_STYLE">Response Style — Formatting & tone preferences</option>
                    <option value="SUMMARY_PREFERENCE">Summary Preference — Brevity & bullet counts</option>
                    <option value="TERMINOLOGY">Terminology — Preferred terms (e.g. shipper vs client)</option>
                    <option value="BUSINESS_FACT">Business Fact — Working hours, non-operating days</option>
                    <option value="MODULE_PREFERENCE">Module Preference — Primary workspace focus</option>
                    <option value="EXPLANATION_DEPTH">Explanation Depth — Detail level for calculations</option>
                  </select>
                </div>

                <div className="pref-form-group">
                  <label>Scope *</label>
                  <select 
                    value={modalForm.scope}
                    onChange={(e) => setModalForm({ ...modalForm, scope: e.target.value })}
                    disabled={modalMode === 'EDIT'}
                  >
                    <option value="USER">Personal — Only used for your account</option>
                    <option value="ORGANIZATION" disabled={!isAdmin}>
                      Organization-Wide {!isAdmin ? '(Admin Only)' : '— Applies to all users'}
                    </option>
                  </select>
                </div>

                <div className="pref-form-group">
                  <label>Memory Instruction / Content *</label>
                  <textarea 
                    rows={4}
                    placeholder="Explain the explicit preference or fact clearly..."
                    value={modalForm.content}
                    onChange={(e) => setModalForm({ ...modalForm, content: e.target.value })}
                    style={{ border: '1px solid #cbd5e1', borderRadius: '6px', padding: '10px', fontSize: '0.875rem' }}
                    required
                  />
                </div>

                {modalMode === 'CREATE' && (
                  <div className="pref-form-group">
                    <label>Expiration (Optional)</label>
                    <select 
                      value={modalForm.expires_in_days}
                      onChange={(e) => setModalForm({ ...modalForm, expires_in_days: e.target.value })}
                    >
                      <option value="0">Never Expires (Permanent until deleted)</option>
                      <option value="7">Expires in 7 days</option>
                      <option value="30">Expires in 30 days</option>
                      <option value="90">Expires in 90 days</option>
                    </select>
                  </div>
                )}

                <div style={{ marginTop: '8px', display: 'flex', alignItems: 'flex-start', gap: '8px' }}>
                  <input 
                    type="checkbox" 
                    id="explicit_confirm"
                    checked={modalForm.explicitly_confirmed}
                    onChange={(e) => setModalForm({ ...modalForm, explicitly_confirmed: e.target.checked })}
                    style={{ marginTop: '3px' }}
                    required
                  />
                  <label htmlFor="explicit_confirm" style={{ fontSize: '0.8125rem', color: '#334155' }}>
                    I explicitly authorize LogisticsHQ AI to store and reference this operational preference.
                  </label>
                </div>
              </div>

              <div className="memory-modal-footer">
                <button 
                  type="button" 
                  className="btn-icon-action"
                  onClick={() => setShowModal(false)}
                >
                  Cancel
                </button>
                <button 
                  type="submit" 
                  className="btn-primary-memory"
                  disabled={submittingModal}
                >
                  {submittingModal ? 'Saving...' : 'Confirm & Save'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* CONFIRM CLEAR MODAL */}
      {showClearConfirm && (
        <div className="modal-overlay" onClick={() => setShowClearConfirm(false)}>
          <div className="memory-modal-card" onClick={(e) => e.stopPropagation()}>
            <div className="memory-modal-header">
              <h3 style={{ color: '#dc2626' }}>Clear Personal Memories</h3>
              <button 
                type="button" 
                onClick={() => setShowClearConfirm(false)}
                style={{ background: 'none', border: 'none', cursor: 'pointer', fontSize: '1.25rem', color: '#64748b' }}
              >
                ×
              </button>
            </div>
            <div className="memory-modal-body">
              <p style={{ margin: 0, fontSize: '0.875rem', color: '#334155', lineHeight: 1.5 }}>
                Are you sure you want to clear <strong>all personal memory items</strong>?
              </p>
              <div className="safety-notice-banner" style={{ background: '#fef2f2', borderColor: '#fca5a5', color: '#b91c1c' }}>
                <AlertCircle size={18} style={{ flexShrink: 0, marginTop: '2px' }} />
                <div>
                  This action will soft-delete all personal preferences saved for your account. Organization-wide policies and system settings will remain unaffected.
                </div>
              </div>
            </div>
            <div className="memory-modal-footer">
              <button 
                type="button" 
                className="btn-icon-action"
                onClick={() => setShowClearConfirm(false)}
              >
                Cancel
              </button>
              <button 
                type="button" 
                className="btn-danger-outline"
                style={{ background: '#dc2626', color: '#ffffff', borderColor: '#b91c1c' }}
                onClick={handleClearPersonal}
              >
                Yes, Clear Memories
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

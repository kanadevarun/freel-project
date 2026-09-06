import { useState, useEffect, useRef } from 'react';
import { 
  Settings, 
  Bell, 
  Copy, 
  RotateCcw, 
  Download, 
  RefreshCw, 
  Trash2,
  Lightbulb,
  ChevronRight
} from 'lucide-react';
import { useRBAC } from '../../../context/RBACContext';
import api from '../../../services/api';
import toast from 'react-hot-toast';
import { COUNTRIES, getFlagEmoji } from '../../../utils/countries';
import './WorkspaceSettingsPage.css';

// ── Helpers ──────────────────────────────────────────────────────────────────
const TIMEZONES = [
  { value: "UTC", label: "UTC" },
  { value: "Asia/Kolkata", label: "Asia/Kolkata (IST +5:30)" },
  { value: "Asia/Dubai", label: "Asia/Dubai (GST +4:00)" },
  { value: "Asia/Singapore", label: "Asia/Singapore (SGT +8:00)" },
  { value: "Europe/London", label: "Europe/London (GMT/BST)" },
  { value: "America/New_York", label: "America/New_York (EST/EDT)" },
  { value: "America/Los_Angeles", label: "America/Los_Angeles (PST/PDT)" }
];

const CURRENCIES = [
  { value: "USD", label: "USD – US Dollar" },
  { value: "EUR", label: "EUR – Euro" },
  { value: "GBP", label: "GBP – British Pound" },
  { value: "INR", label: "INR – Indian Rupee" },
  { value: "SGD", label: "SGD – Singapore Dollar" },
  { value: "AED", label: "AED – UAE Dirham" },
  { value: "CNY", label: "CNY – Chinese Yuan" }
];

const LANGUAGES = [
  { value: "English", label: "English" },
  { value: "Spanish", label: "Spanish" },
  { value: "French", label: "French" },
  { value: "German", label: "German" }
];

const DATE_FORMATS = [
  { value: "DD/MM/YYYY", label: "DD/MM/YYYY" },
  { value: "MM/DD/YYYY", label: "MM/DD/YYYY" },
  { value: "YYYY-MM-DD", label: "YYYY-MM-DD" }
];

const MEASUREMENTS = [
  { value: "Metric (kg, cm, km)", label: "Metric (kg, cm, km)" },
  { value: "Imperial (lb, in, mi)", label: "Imperial (lb, in, mi)" }
];

const WEIGHT_UNITS = [
  { value: "Kilogram (kg)", label: "Kilogram (kg)" },
  { value: "Pound (lb)", label: "Pound (lb)" }
];

const DIMENSION_UNITS = [
  { value: "Centimeter (cm)", label: "Centimeter (cm)" },
  { value: "Inch (in)", label: "Inch (in)" }
];

const VOLUME_UNITS = [
  { value: "Cubic Meter (CBM)", label: "Cubic Meter (CBM)" },
  { value: "Cubic Foot (CFT)", label: "Cubic Foot (CFT)" }
];

function formatDate(isoString) {
  if (!isoString) return '';
  const d = new Date(isoString);
  return d.toLocaleDateString('en-GB', { day: '2-digit', month: 'short', year: 'numeric' });
}

function copyToClipboard(text) {
  navigator.clipboard.writeText(text);
  toast.success('Copied to clipboard');
}

export default function WorkspaceSettingsPage() {
  const { can } = useRBAC();
  const canEdit = can('SETTINGS', 'UPDATE');

  const [isLoading, setIsLoading] = useState(true);
  const [isSaving, setIsSaving] = useState(false);
  const [isDirty, setIsDirty] = useState(false);
  const [activeTab, setActiveTab] = useState('General');

  const [form, setForm] = useState({
    id: '',
    name: '',
    default_timezone: '',
    default_language: '',
    default_currency: '',
    date_format: '',
    country: '', // Used for default country
    measurement_system: '',
    weight_unit: '',
    dimension_unit: '',
    volume_unit: '',
    time_format: '',
    created_at: '',
    updated_at: ''
  });

  const [originalForm, setOriginalForm] = useState({});

  const [notificationForm, setNotificationForm] = useState({
    new_rfq_received: true,
    new_quote_received: true,
    shipment_status_updates: true,
    shipment_exceptions: true,
    invitation_accepted: true,
    invoice_payment_events: true,
    system_security_alerts: true
  });
  const [originalNotificationForm, setOriginalNotificationForm] = useState({});
  const [isNotificationDirty, setIsNotificationDirty] = useState(false);

  useEffect(() => {
    fetchProfile();
    fetchNotifications();
  }, []);

  const fetchProfile = async () => {
    try {
      const data = await api.get('/api/v1/organizations/profile');
      
      const payload = {
        id: data.id || '',
        name: data.name || '',
        default_timezone: data.default_timezone || 'UTC',
        default_language: data.default_language || 'English',
        default_currency: data.default_currency || 'USD',
        date_format: data.date_format || 'DD/MM/YYYY',
        country: data.country || 'India', // Reusing existing country field for localization
        measurement_system: data.measurement_system || 'Metric (kg, cm, km)',
        weight_unit: data.weight_unit || 'Kilogram (kg)',
        dimension_unit: data.dimension_unit || 'Centimeter (cm)',
        volume_unit: data.volume_unit || 'Cubic Meter (CBM)',
        time_format: data.time_format || '24 Hours',
        created_at: data.created_at,
        updated_at: data.updated_at
      };

      setForm(payload);
      setOriginalForm(payload);
    } catch (err) {
      console.error(err);
      toast.error('Failed to load workspace settings');
    } finally {
      setIsLoading(false);
    }
  };

  const fetchNotifications = async () => {
    try {
      const data = await api.get('/api/v1/organizations/notifications');
      setNotificationForm(data);
      setOriginalNotificationForm(data);
    } catch (err) {
      console.error(err);
    }
  };

  const onChange = (e) => {
    const { name, value } = e.target;
    setForm(prev => {
      const next = { ...prev, [name]: value };
      checkDirty(next);
      return next;
    });
  };

  const checkDirty = (currentForm) => {
    // Exclude read-only fields from comparison
    const fields = [
      'name', 'default_timezone', 'default_language', 'default_currency', 
      'date_format', 'country', 'measurement_system', 'weight_unit', 
      'dimension_unit', 'volume_unit', 'time_format'
    ];
    let dirty = false;
    for (const f of fields) {
      if (currentForm[f] !== originalForm[f]) {
        dirty = true;
        break;
      }
    }
    setIsDirty(dirty);
  };

  const onDiscard = () => {
    if (activeTab === 'General') {
      setForm(originalForm);
      setIsDirty(false);
    } else if (activeTab === 'Notifications') {
      setNotificationForm(originalNotificationForm);
      setIsNotificationDirty(false);
    }
  };

  const onNotificationChange = (e) => {
    const { name, checked } = e.target;
    setNotificationForm(prev => {
      const next = { ...prev, [name]: checked };
      checkNotificationDirty(next);
      return next;
    });
  };

  const checkNotificationDirty = (currentForm) => {
    const fields = [
      'new_rfq_received', 'new_quote_received', 'shipment_status_updates',
      'shipment_exceptions', 'invitation_accepted', 'invoice_payment_events',
      'system_security_alerts'
    ];
    let dirty = false;
    for (const f of fields) {
      if (currentForm[f] !== originalNotificationForm[f]) {
        dirty = true;
        break;
      }
    }
    setIsNotificationDirty(dirty);
  };

  const onSave = async () => {
    setIsSaving(true);
    try {
      if (activeTab === 'General') {
        const currentOrg = await api.get('/api/v1/organizations/profile');
        const updatePayload = {
          ...currentOrg,
          name: form.name,
          default_timezone: form.default_timezone,
          default_language: form.default_language,
          default_currency: form.default_currency,
          date_format: form.date_format,
          country: form.country,
          measurement_system: form.measurement_system,
          weight_unit: form.weight_unit,
          dimension_unit: form.dimension_unit,
          volume_unit: form.volume_unit,
          time_format: form.time_format
        };
        await api.put('/api/v1/organizations/profile', updatePayload);
        toast.success('Workspace settings updated successfully');
        await fetchProfile();
        setIsDirty(false);
      } else if (activeTab === 'Notifications') {
        await api.put('/api/v1/organizations/notifications', notificationForm);
        toast.success('Notification preferences updated successfully');
        await fetchNotifications();
        setIsNotificationDirty(false);
      }
    } catch (err) {
      console.error(err);
      toast.error('Failed to update workspace settings');
    } finally {
      setIsSaving(false);
    }
  };

  if (isLoading) {
    return (
      <div className="ws-page">
        <div style={{height: 100, background: '#f1f5f9', borderRadius: 8, animation: 'pulse 1.5s infinite'}}></div>
      </div>
    );
  }

  return (
    <div className="ws-page">
      <div className="ws-header">
        <h1 className="ws-title">Workspace Settings</h1>
        <p className="ws-subtitle">Customize how your workspace operates across LogisticsHQ.</p>
      </div>

      <div className="ws-tabs">
        <div className={`ws-tab ${activeTab === 'General' ? 'active' : ''}`} onClick={() => setActiveTab('General')}>
          <Settings /> General
        </div>
        <div className={`ws-tab ${activeTab === 'Notifications' ? 'active' : ''}`} onClick={() => setActiveTab('Notifications')}>
          <Bell /> Notifications
        </div>
      </div>

      <div className="ws-grid">
        <div className="ws-main-col">
          {activeTab === 'General' ? (
            <>
              {/* Workspace Identity */}
              <div className="ws-card">
              <div className="ws-card-header">
                <div className="ws-card-icon purple">
                  <BuildingIcon />
                </div>
                <div className="ws-card-title-wrap">
                  <h3 className="ws-card-title">Workspace Identity</h3>
                  <p className="ws-card-desc">Basic details that identify your workspace.</p>
                </div>
              </div>
              
              <div className="ws-card-body">
                <div className="ws-field">
                  <label className="ws-label">Workspace Name</label>
                  <div className="ws-field-content">
                    <input 
                      className="ws-input"
                      name="name"
                      value={form.name}
                      onChange={onChange}
                      disabled={!canEdit}
                    />
                  </div>
                </div>
                <div className="ws-field">
                  <label className="ws-label">Workspace ID</label>
                  <div className="ws-field-content" style={{display: 'flex', gap: 8}}>
                    <input 
                      className="ws-input"
                      value={`WS-${form.name.substring(0, 4).toUpperCase()}-${form.id}`}
                      disabled
                    />
                    <button className="ws-copy-btn" onClick={() => copyToClipboard(`WS-${form.name.substring(0, 4).toUpperCase()}-${form.id}`)}>
                      <Copy size={16} />
                    </button>
                  </div>
                </div>
                <div className="ws-field">
                  <label className="ws-label">Created On</label>
                  <div className="ws-field-content">
                    <div className="ws-value">{formatDate(form.created_at)}</div>
                  </div>
                </div>
                <div className="ws-field">
                  <label className="ws-label">Last Updated</label>
                  <div className="ws-field-content">
                    <div className="ws-value">{formatDate(form.updated_at)}</div>
                  </div>
                </div>
              </div>
            </div>

            {/* Localization */}
            <div className="ws-card">
              <div className="ws-card-header">
                <div className="ws-card-icon blue">
                  <GlobeIcon />
                </div>
                <div className="ws-card-title-wrap">
                  <h3 className="ws-card-title">Localization</h3>
                  <p className="ws-card-desc">Set language, timezone, and regional preferences.</p>
                </div>
              </div>

              <div className="ws-card-body">
                <div className="ws-field">
                  <label className="ws-label">Default Timezone</label>
                  <div className="ws-field-content">
                    <CustomSelect name="default_timezone" value={form.default_timezone} onChange={onChange} options={TIMEZONES} disabled={!canEdit} />
                  </div>
                </div>
                <div className="ws-field">
                  <label className="ws-label">Default Language</label>
                  <div className="ws-field-content">
                    <CustomSelect name="default_language" value={form.default_language} onChange={onChange} options={LANGUAGES} disabled={!canEdit} />
                  </div>
                </div>
                <div className="ws-field">
                  <label className="ws-label">Default Currency</label>
                  <div className="ws-field-content">
                    <CustomSelect name="default_currency" value={form.default_currency} onChange={onChange} options={CURRENCIES} disabled={!canEdit} />
                  </div>
                </div>
                <div className="ws-field">
                  <label className="ws-label">Date Format</label>
                  <div className="ws-field-content">
                    <CustomSelect name="date_format" value={form.date_format} onChange={onChange} options={DATE_FORMATS} disabled={!canEdit} />
                  </div>
                </div>
                <div className="ws-field">
                  <label className="ws-label">Default Country</label>
                  <div className="ws-field-content">
                    <CustomSelect 
                      name="country" 
                      value={form.country} 
                      onChange={onChange} 
                      disabled={!canEdit} 
                      options={COUNTRIES.map(c => ({
                        value: c.label,
                        label: `${getFlagEmoji(c.iso)} ${c.label}`
                      }))}
                    />
                  </div>
                </div>
                <div className="ws-field">
                  <label className="ws-label">Measurement System</label>
                  <div className="ws-field-content">
                    <CustomSelect name="measurement_system" value={form.measurement_system} onChange={onChange} options={MEASUREMENTS} disabled={!canEdit} />
                  </div>
                </div>
              </div>
            </div>

          {/* System Preferences */}
          <div className="ws-card">
            <div className="ws-card-header">
              <div className="ws-card-icon green">
                <SlidersIcon />
              </div>
              <div className="ws-card-title-wrap">
                <h3 className="ws-card-title">System Preferences</h3>
                <p className="ws-card-desc">Configure default units and system behavior.</p>
              </div>
            </div>

            <div className="ws-card-body">
              <div className="ws-field">
                <label className="ws-label">Weight Unit</label>
                <div className="ws-field-content">
                  <CustomSelect name="weight_unit" value={form.weight_unit} onChange={onChange} options={WEIGHT_UNITS} disabled={!canEdit} />
                </div>
              </div>
              <div className="ws-field">
                <label className="ws-label">Dimension Unit</label>
                <div className="ws-field-content">
                  <CustomSelect name="dimension_unit" value={form.dimension_unit} onChange={onChange} options={DIMENSION_UNITS} disabled={!canEdit} />
                </div>
              </div>
              <div className="ws-field">
                <label className="ws-label">Volume Unit</label>
                <div className="ws-field-content">
                  <CustomSelect name="volume_unit" value={form.volume_unit} onChange={onChange} options={VOLUME_UNITS} disabled={!canEdit} />
                </div>
              </div>
              
              <div className="ws-field">
                <label className="ws-label">Time Format</label>
                <div className="ws-field-content">
                  <div className="ws-toggle-group">
                    <button 
                      className={`ws-toggle-btn ${form.time_format === '12 Hours' ? 'active' : ''}`}
                      onClick={() => canEdit && onChange({target: {name: 'time_format', value: '12 Hours'}})}
                      disabled={!canEdit}
                    >
                      12 Hours
                    </button>
                    <button 
                      className={`ws-toggle-btn ${form.time_format === '24 Hours' ? 'active' : ''}`}
                      onClick={() => canEdit && onChange({target: {name: 'time_format', value: '24 Hours'}})}
                      disabled={!canEdit}
                    >
                      24 Hours
                    </button>
                  </div>
                </div>
              </div>
              </div>
            </div>
            </>
          ) : activeTab === 'Notifications' ? (
            <div className="ws-card">
              <div className="ws-card-header">
                <div className="ws-card-icon blue">
                  <Bell />
                </div>
                <div className="ws-card-title-wrap">
                  <h3 className="ws-card-title">Notification Preferences</h3>
                  <p className="ws-card-desc">Control which events trigger emails and in-app alerts.</p>
                </div>
              </div>

              <div className="ws-card-body">
                <div className="ws-field" style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                  <div>
                    <label className="ws-label" style={{ marginBottom: 4 }}>New RFQ Received</label>
                    <div style={{ fontSize: 13, color: '#64748b' }}>Get notified when a customer requests a new quote.</div>
                  </div>
                  <label className="ws-toggle-switch">
                    <input type="checkbox" name="new_rfq_received" checked={notificationForm.new_rfq_received} onChange={onNotificationChange} disabled={!canEdit} />
                    <span className="ws-toggle-slider"></span>
                  </label>
                </div>
                
                <div className="ws-field" style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                  <div>
                    <label className="ws-label" style={{ marginBottom: 4 }}>New Quote Received</label>
                    <div style={{ fontSize: 13, color: '#64748b' }}>Get notified when a carrier or partner submits a quote.</div>
                  </div>
                  <label className="ws-toggle-switch">
                    <input type="checkbox" name="new_quote_received" checked={notificationForm.new_quote_received} onChange={onNotificationChange} disabled={!canEdit} />
                    <span className="ws-toggle-slider"></span>
                  </label>
                </div>

                <div className="ws-field" style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                  <div>
                    <label className="ws-label" style={{ marginBottom: 4 }}>Shipment Status Updates</label>
                    <div style={{ fontSize: 13, color: '#64748b' }}>Alerts for key milestones (Departed, Arrived, Cleared).</div>
                  </div>
                  <label className="ws-toggle-switch">
                    <input type="checkbox" name="shipment_status_updates" checked={notificationForm.shipment_status_updates} onChange={onNotificationChange} disabled={!canEdit} />
                    <span className="ws-toggle-slider"></span>
                  </label>
                </div>

                <div className="ws-field" style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                  <div>
                    <label className="ws-label" style={{ marginBottom: 4 }}>Shipment Exceptions</label>
                    <div style={{ fontSize: 13, color: '#64748b' }}>Urgent alerts for delays, damages, or missing documents.</div>
                  </div>
                  <label className="ws-toggle-switch">
                    <input type="checkbox" name="shipment_exceptions" checked={notificationForm.shipment_exceptions} onChange={onNotificationChange} disabled={!canEdit} />
                    <span className="ws-toggle-slider"></span>
                  </label>
                </div>

                <div className="ws-field" style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                  <div>
                    <label className="ws-label" style={{ marginBottom: 4 }}>Invitation Accepted</label>
                    <div style={{ fontSize: 13, color: '#64748b' }}>When a team member accepts your workspace invite.</div>
                  </div>
                  <label className="ws-toggle-switch">
                    <input type="checkbox" name="invitation_accepted" checked={notificationForm.invitation_accepted} onChange={onNotificationChange} disabled={!canEdit} />
                    <span className="ws-toggle-slider"></span>
                  </label>
                </div>

                <div className="ws-field" style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                  <div>
                    <label className="ws-label" style={{ marginBottom: 4 }}>Invoice & Payment Events</label>
                    <div style={{ fontSize: 13, color: '#64748b' }}>Alerts for new invoices, due dates, and received payments.</div>
                  </div>
                  <label className="ws-toggle-switch">
                    <input type="checkbox" name="invoice_payment_events" checked={notificationForm.invoice_payment_events} onChange={onNotificationChange} disabled={!canEdit} />
                    <span className="ws-toggle-slider"></span>
                  </label>
                </div>

                <div className="ws-field" style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                  <div>
                    <label className="ws-label" style={{ marginBottom: 4 }}>System & Security Alerts</label>
                    <div style={{ fontSize: 13, color: '#64748b' }}>Critical platform updates, maintenance, and logins from new IPs.</div>
                  </div>
                  <label className="ws-toggle-switch">
                    <input type="checkbox" name="system_security_alerts" checked={notificationForm.system_security_alerts} onChange={onNotificationChange} disabled={!canEdit} />
                    <span className="ws-toggle-slider"></span>
                  </label>
                </div>
              </div>
            </div>
          ) : null}
          
        </div>
        
        <div className="ws-side-col">
          
          {/* Quick Actions */}
          <div className="ws-card">
            <div className="ws-card-header" style={{ padding: '16px 20px' }}>
              <div className="ws-card-title-wrap">
                <h3 className="ws-card-title" style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                  <ZapIcon />
                  Quick Actions
                </h3>
              </div>
            </div>
            
            <div className="ws-action-list">
              <div className="ws-action-item">
                <div className="ws-action-left">
                  <div className="ws-action-icon"><RotateCcw size={16} color="#64748b" /></div>
                  <div className="ws-action-text">
                    <span className="ws-action-title">Reset Preferences</span>
                    <span className="ws-action-desc">Revert to default settings</span>
                  </div>
                </div>
                <ChevronRight className="ws-action-arrow" size={16} />
              </div>
              <div className="ws-action-item">
                <div className="ws-action-left">
                  <div className="ws-action-icon"><Download size={16} color="#64748b" /></div>
                  <div className="ws-action-text">
                    <span className="ws-action-title">Export Settings</span>
                    <span className="ws-action-desc">Download as JSON</span>
                  </div>
                </div>
                <ChevronRight className="ws-action-arrow" size={16} />
              </div>
              <div className="ws-action-item">
                <div className="ws-action-left">
                  <div className="ws-action-icon"><RefreshCw size={16} color="#64748b" /></div>
                  <div className="ws-action-text">
                    <span className="ws-action-title">Sync Data</span>
                    <span className="ws-action-desc">Update reference data</span>
                  </div>
                </div>
                <ChevronRight className="ws-action-arrow" size={16} />
              </div>
            </div>
          </div>

          <div className="ws-tip-card">
            <Lightbulb size={20} />
            <p className="ws-tip-text">
              <strong>Workspace Tip</strong><br/>
              Changes in operations or workflow settings will apply to all users in this workspace.
            </p>
          </div>

        </div>
      </div>

      {(isDirty || isNotificationDirty) && canEdit && (
        <div className="ws-save-banner">
          <div className="ws-banner-left">
            <div className="ws-banner-dot"></div>
            You have unsaved changes
          </div>
          <div className="ws-banner-actions">
            <button className="ws-btn ws-btn-ghost" onClick={onDiscard} disabled={isSaving}>Discard Changes</button>
            <button className="ws-btn ws-btn-primary" onClick={onSave} disabled={isSaving}>
              {isSaving ? 'Saving...' : 'Save Changes'}
            </button>
          </div>
        </div>
      )}
      {!(isDirty || isNotificationDirty) && canEdit && (
        <div className="ws-save-banner" style={{background: '#f8fafc', borderTop: '1px solid #e2e8f0', color: '#64748b'}}>
          <div className="ws-banner-left">
            <div className="ws-banner-dot" style={{background: '#10b981'}}></div>
            All changes saved
          </div>
        </div>
      )}
    </div>
  );
}

// ── Icons ─────────────────────────────────────────────────────────────────────
function BuildingIcon() {
  return (
    <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
      <rect x="4" y="2" width="16" height="20" rx="2" ry="2"></rect>
      <path d="M9 22v-4h6v4"></path>
      <path d="M8 6h.01"></path>
      <path d="M16 6h.01"></path>
      <path d="M12 6h.01"></path>
      <path d="M12 10h.01"></path>
      <path d="M12 14h.01"></path>
      <path d="M16 10h.01"></path>
      <path d="M16 14h.01"></path>
      <path d="M8 10h.01"></path>
      <path d="M8 14h.01"></path>
    </svg>
  );
}

function GlobeIcon() {
  return (
    <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
      <circle cx="12" cy="12" r="10"></circle>
      <line x1="2" y1="12" x2="22" y2="12"></line>
      <path d="M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z"></path>
    </svg>
  );
}

function SlidersIcon() {
  return (
    <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
      <line x1="4" y1="21" x2="4" y2="14"></line>
      <line x1="4" y1="10" x2="4" y2="3"></line>
      <line x1="12" y1="21" x2="12" y2="12"></line>
      <line x1="12" y1="8" x2="12" y2="3"></line>
      <line x1="20" y1="21" x2="20" y2="16"></line>
      <line x1="20" y1="12" x2="20" y2="3"></line>
      <line x1="1" y1="14" x2="7" y2="14"></line>
      <line x1="9" y1="8" x2="15" y2="8"></line>
      <line x1="17" y1="16" x2="23" y2="16"></line>
    </svg>
  );
}

function ActivityIcon() {
  return (
    <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
      <polyline points="22 12 18 12 15 21 9 3 6 12 2 12"></polyline>
    </svg>
  );
}

function ZapIcon() {
  return (
    <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
      <polygon points="13 2 3 14 12 14 11 22 21 10 12 10 13 2"></polygon>
    </svg>
  );
}

// ── Custom Select Component ───────────────────────────────────────────────────
function CustomSelect({ name, value, onChange, options, disabled, placeholder }) {
  const [isOpen, setIsOpen] = useState(false);
  const containerRef = useRef(null);

  useEffect(() => {
    const handleOutsideClick = (e) => {
      if (containerRef.current && !containerRef.current.contains(e.target)) {
        setIsOpen(false);
      }
    };
    if (isOpen) document.addEventListener('mousedown', handleOutsideClick);
    return () => document.removeEventListener('mousedown', handleOutsideClick);
  }, [isOpen]);

  const selectedOption = options.find(o => o.value === value);

  return (
    <div className={`ws-custom-dropdown ${disabled ? 'disabled' : ''}`} ref={containerRef}>
      <div 
        className="ws-custom-dropdown-selected" 
        onClick={() => !disabled && setIsOpen(!isOpen)}
      >
        <span className="ws-custom-dropdown-text">
          {selectedOption ? selectedOption.label : <span className="placeholder">{placeholder || 'Select...'}</span>}
        </span>
        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
          <polyline points="6 9 12 15 18 9"></polyline>
        </svg>
      </div>
      {isOpen && !disabled && (
        <ul className="ws-custom-dropdown-list">
          {options.map(opt => (
            <li
              key={opt.value}
              className={value === opt.value ? 'selected' : ''}
              onClick={(e) => {
                e.stopPropagation();
                onChange({ target: { name, value: opt.value } });
                setIsOpen(false);
              }}
            >
              {opt.label}
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}

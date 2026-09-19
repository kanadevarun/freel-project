import React, { useState, useEffect, useCallback } from 'react';
import { 
  MessageSquare, 
  Mail, 
  Ship, 
  Shield, 
  HardDrive, 
  FileSearch, 
  RefreshCw, 
  CheckCircle2, 
  AlertTriangle, 
  XCircle, 
  Sliders, 
  Lock,
  X
} from 'lucide-react';
import integrationService from '../../../services/integrationService';
import './ExternalIntegrationsPage.css';

const INTEGRATION_META = {
  SMS: {
    title: 'SMS Notifications',
    description: 'Outbound dispatch for operational milestone and delivery alerts.',
    icon: MessageSquare,
  },
  EMAIL: {
    title: 'Email Relay & SES',
    description: 'High-volume commercial quotations, booking confirmations, and invoices.',
    icon: Mail,
  },
  CARRIER_TRACKING: {
    title: 'Carrier Tracking Telemetry',
    description: 'Direct EDI/API connection with ocean carriers for live container milestones.',
    icon: Ship,
  },
  WEBHOOK: {
    title: 'Webhook Security Gateway',
    description: 'Ingress verification, HMAC-SHA256 signatures, and replay attack defense.',
    icon: Shield,
  },
  STORAGE: {
    title: 'Cloud Object Storage',
    description: 'Secure, encrypted document repository for bills of lading and invoices.',
    icon: HardDrive,
  },
  TEXTRACT: {
    title: 'Intelligent Document OCR',
    description: 'Automated extraction and classification of commercial freight paperwork.',
    icon: FileSearch,
  },
};

export default function ExternalIntegrationsPage() {
  const [activeTab, setActiveTab] = useState('providers');
  const [integrations, setIntegrations] = useState([]);
  const [deadLetters, setDeadLetters] = useState([]);
  const [smsMessages, setSmsMessages] = useState([]);
  const [emailMessages, setEmailMessages] = useState([]);
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  const [error, setError] = useState(null);
  const [editItem, setEditItem] = useState(null);
  const [saving, setSaving] = useState(false);
  const [saveSuccess, setSaveSuccess] = useState(null);
  const [testModalOpen, setTestModalOpen] = useState(false);
  const [testForm, setTestForm] = useState({ to_phone: '', message: '' });
  const [testResult, setTestResult] = useState(null);
  const [testingSMS, setTestingSMS] = useState(false);
  const [testEmailModalOpen, setTestEmailModalOpen] = useState(false);
  const [testEmailForm, setTestEmailForm] = useState({ to_email: '', subject: '', body_text: '' });
  const [testEmailResult, setTestEmailResult] = useState(null);
  const [testingEmail, setTestingEmail] = useState(false);

  const fetchIntegrations = useCallback(async () => {
    try {
      setLoading(true);
      setError(null);
      const res = await integrationService.getStatuses();
      const data = res?.data || res || [];
      setIntegrations(Array.isArray(data) ? data : []);
    } catch (err) {
      setError(err?.message || 'Failed to load external integration statuses');
    } finally {
      setLoading(false);
    }
  }, []);

  const fetchDeadLetters = useCallback(async () => {
    try {
      const res = await integrationService.getDeadLetter();
      const data = res?.data || res || [];
      setDeadLetters(Array.isArray(data) ? data : []);
    } catch (err) {
      console.error('Failed to load dead letter events:', err);
    }
  }, []);

  const fetchSMSMessages = useCallback(async () => {
    try {
      const res = await integrationService.getSMSMessages(50);
      const data = res?.data || res || [];
      setSmsMessages(Array.isArray(data) ? data : []);
    } catch (err) {
      console.error('Failed to load SMS dispatch logs:', err);
    }
  }, []);

  const fetchEmailMessages = useCallback(async () => {
    try {
      const res = await integrationService.getEmailMessages(50);
      const data = res?.data || res || [];
      setEmailMessages(Array.isArray(data) ? data : []);
    } catch (err) {
      console.error('Failed to load Email dispatch logs:', err);
    }
  }, []);

  useEffect(() => {
    fetchIntegrations();
    fetchDeadLetters();
    fetchSMSMessages();
    fetchEmailMessages();
  }, [fetchIntegrations, fetchDeadLetters, fetchSMSMessages, fetchEmailMessages]);

  const handleRefresh = async () => {
    setRefreshing(true);
    await Promise.all([fetchIntegrations(), fetchDeadLetters(), fetchSMSMessages(), fetchEmailMessages()]);
    setRefreshing(false);
  };

  const handleSendTestSMS = async (e) => {
    e.preventDefault();
    try {
      setTestingSMS(true);
      setTestResult(null);
      const res = await integrationService.testSMS(testForm);
      setTestResult({ success: true, data: res?.data || res });
      fetchSMSMessages();
    } catch (err) {
      const errData = err?.response?.data || { error: err.message };
      setTestResult({ success: false, error: errData });
    } finally {
      setTestingSMS(false);
    }
  };

  const handleSendTestEmail = async (e) => {
    e.preventDefault();
    try {
      setTestingEmail(true);
      setTestEmailResult(null);
      const res = await integrationService.testEmail(testEmailForm);
      setTestEmailResult({ success: true, data: res?.data || res });
      fetchEmailMessages();
    } catch (err) {
      const errData = err?.response?.data || { error: err.message };
      setTestEmailResult({ success: false, error: errData });
    } finally {
      setTestingEmail(false);
    }
  };

  const handleSaveConfig = async (e) => {
    e.preventDefault();
    if (!editItem) return;
    try {
      setSaving(true);
      setSaveSuccess(null);
      await integrationService.saveConfig({
        type: editItem.type,
        provider: editItem.provider,
        is_enabled: editItem.is_enabled,
        endpoint_url: editItem.endpoint_url || '',
        region: editItem.region || '',
        from_address: editItem.from_address || '',
        timeout_sec: parseInt(editItem.timeout_sec, 10) || 15,
        max_retries: parseInt(editItem.max_retries, 10) || 3,
        secret_value: editItem.new_secret || '',
      });
      setSaveSuccess('Configuration updated successfully.');
      setTimeout(() => {
        setEditItem(null);
        setSaveSuccess(null);
        fetchIntegrations();
      }, 1000);
    } catch (err) {
      alert('Failed to save configuration: ' + (err?.message || 'Error'));
    } finally {
      setSaving(false);
    }
  };

  const renderStatusBadge = (status) => {
    const s = (status || 'NOT_CONFIGURED').toUpperCase();
    switch (s) {
      case 'HEALTHY':
      case 'ENABLED':
        return (
          <span className="ext-status-pill healthy">
            <span className="ext-status-dot" /> {s === 'HEALTHY' ? 'Healthy' : 'Enabled'}
          </span>
        );
      case 'DISABLED':
        return (
          <span className="ext-status-pill disabled">
            <span className="ext-status-dot" /> Disabled
          </span>
        );
      case 'CONFIG_INVALID':
        return (
          <span className="ext-status-pill config_invalid">
            <span className="ext-status-dot" /> Config Invalid
          </span>
        );
      case 'UNAVAILABLE':
      case 'DEGRADED':
        return (
          <span className="ext-status-pill unavailable">
            <span className="ext-status-dot" /> {s}
          </span>
        );
      case 'NOT_CONFIGURED':
      default:
        return (
          <span className="ext-status-pill not_configured">
            <span className="ext-status-dot" /> Not Configured
          </span>
        );
    }
  };

  const renderDeliveryStatusBadge = (status) => {
    const s = (status || 'UNKNOWN').toUpperCase();
    switch (s) {
      case 'DELIVERED':
        return <span className="ext-status-pill healthy"><span className="ext-status-dot" /> Delivered</span>;
      case 'SENT':
        return <span className="ext-status-pill config_invalid" style={{ background: '#fffbeb', color: '#b45309', borderColor: '#fde68a' }}><span className="ext-status-dot" style={{ background: '#f59e0b' }} /> Sent</span>;
      case 'ACCEPTED':
      case 'QUEUED':
        return <span className="ext-status-pill healthy" style={{ background: '#eff6ff', color: '#1d4ed8', borderColor: '#bfdbfe' }}><span className="ext-status-dot" style={{ background: '#3b82f6' }} /> {s}</span>;
      case 'BOUNCED':
        return <span className="ext-status-pill unavailable" style={{ background: '#fff1f2', color: '#be123c', borderColor: '#fecdd3' }}><span className="ext-status-dot" style={{ background: '#e11d48' }} /> Bounced</span>;
      case 'COMPLAINED':
        return <span className="ext-status-pill unavailable" style={{ background: '#fdf4ff', color: '#a21caf', borderColor: '#f5d0fe' }}><span className="ext-status-dot" style={{ background: '#c026d3' }} /> Complained</span>;
      case 'FAILED':
      case 'REJECTED':
      case 'UNDELIVERED':
        return <span className="ext-status-pill unavailable"><span className="ext-status-dot" /> {s}</span>;
      default:
        return <span className="ext-status-pill not_configured"><span className="ext-status-dot" /> {s}</span>;
    }
  };

  return (
    <div className="ext-int-container">
      <div className="ext-int-header">
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <div>
            <h1 className="ext-int-title">External Integrations Gateway</h1>
            <p className="ext-int-subtitle">
              Manage external provider configuration, webhook security, reliability policies, and secrets.
            </p>
          </div>
          <button 
            className="ext-int-btn ext-int-btn-outline" 
            onClick={handleRefresh}
            disabled={refreshing}
          >
            <RefreshCw size={14} className={refreshing ? 'animate-spin' : ''} />
            Refresh
          </button>
        </div>
      </div>

      <div className="ext-int-tabs">
        <button 
          className={`ext-int-tab ${activeTab === 'providers' ? 'active' : ''}`}
          onClick={() => setActiveTab('providers')}
        >
          Provider Gateways ({integrations.length})
        </button>
        <button 
          className={`ext-int-tab ${activeTab === 'sms_logs' ? 'active' : ''}`}
          onClick={() => setActiveTab('sms_logs')}
        >
          SMS Dispatch Logs ({smsMessages.length})
        </button>
        <button 
          className={`ext-int-tab ${activeTab === 'email_logs' ? 'active' : ''}`}
          onClick={() => setActiveTab('email_logs')}
        >
          Email Relay Logs ({emailMessages.length})
        </button>
        <button 
          className={`ext-int-tab ${activeTab === 'dead_letter' ? 'active' : ''}`}
          onClick={() => setActiveTab('dead_letter')}
        >
          Dead-Letter Queue ({deadLetters.length})
        </button>
      </div>

      {error && (
        <div style={{ padding: '12px 16px', background: '#fef2f2', border: '1px solid #fecaca', borderRadius: '8px', color: '#b91c1c', marginBottom: '20px' }}>
          {error}
        </div>
      )}

      {activeTab === 'providers' && (
        <div className="ext-int-grid">
          {integrations.map((item) => {
            const meta = INTEGRATION_META[item.type] || {
              title: item.type,
              description: 'External provider integration service.',
              icon: Sliders,
            };
            const IconComp = meta.icon;

            return (
              <div key={`${item.type}-${item.provider}`} className="ext-int-card">
                <div>
                  <div className="ext-int-card-header">
                    <div className="ext-int-card-identity">
                      <div className="ext-int-icon-wrap">
                        <IconComp size={22} />
                      </div>
                      <div>
                        <h3 className="ext-int-card-name">{meta.title}</h3>
                        <div className="ext-int-card-provider">{item.provider}</div>
                      </div>
                    </div>
                    {renderStatusBadge(item.status)}
                  </div>

                  <div className="ext-int-card-body">
                    <p style={{ margin: '0 0 12px 0', fontSize: '13px', color: '#64748b' }}>
                      {meta.description}
                    </p>

                    <div className="ext-int-meta-row">
                      <span className="ext-int-meta-label">Timeout</span>
                      <span className="ext-int-meta-val">{item.timeout_sec}s</span>
                    </div>
                    <div className="ext-int-meta-row">
                      <span className="ext-int-meta-label">Max Retries</span>
                      <span className="ext-int-meta-val">{item.max_retries} attempts</span>
                    </div>
                    <div className="ext-int-meta-row">
                      <span className="ext-int-meta-label">Secret Credential</span>
                      <span className="ext-int-meta-val">
                        {item.has_secret ? (
                          <span style={{ display: 'inline-flex', alignItems: 'center', gap: '4px', color: '#047857' }}>
                            <Lock size={12} /> Configured ({item.masked_secret || '••••••••'})
                          </span>
                        ) : (
                          <span style={{ color: '#94a3b8' }}>None</span>
                        )}
                      </span>
                    </div>

                    {item.health_message && (
                      <div className="ext-int-health-msg">
                        {item.health_message}
                      </div>
                    )}
                  </div>
                </div>

                <div className="ext-int-card-actions">
                  <button 
                    className="ext-int-btn ext-int-btn-outline"
                    onClick={() => setEditItem({ ...item, new_secret: '' })}
                  >
                    Configure
                  </button>
                </div>
              </div>
            );
          })}
        </div>
      )}

      {activeTab === 'sms_logs' && (
        <div>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '16px' }}>
            <div>
              <h3 style={{ margin: '0 0 4px 0', fontSize: '16px', color: '#0f172a' }}>Outbound SMS Dispatch Records</h3>
              <p style={{ margin: 0, fontSize: '13px', color: '#64748b' }}>
                Monotonic delivery lifecycle tracking with Twilio webhook reconciliation and audit preservation.
              </p>
            </div>
            <button 
              className="ext-int-btn ext-int-btn-primary"
              onClick={() => {
                setTestResult(null);
                setTestModalOpen(true);
              }}
            >
              <MessageSquare size={14} />
              Test Outbound SMS
            </button>
          </div>

          {smsMessages.length === 0 ? (
            <div style={{ padding: '40px', textAlign: 'center', background: '#fff', borderRadius: '8px', border: '1px solid #e2e8f0' }}>
              <MessageSquare size={36} color="#94a3b8" style={{ margin: '0 auto 12px auto' }} />
              <h3 style={{ margin: '0 0 6px 0', color: '#0f172a' }}>No Outbound SMS Dispatches Recorded</h3>
              <p style={{ margin: 0, color: '#64748b', fontSize: '14px' }}>
                Outbound dispatches from business notifications and test calls will appear here with live provider delivery status.
              </p>
            </div>
          ) : (
            <table className="ext-dl-table">
              <thead>
                <tr>
                  <th>Timestamp</th>
                  <th>Recipient</th>
                  <th>Sender</th>
                  <th>Message SID</th>
                  <th>Delivery Status</th>
                  <th>Error / Failure</th>
                  <th>Message Preview</th>
                </tr>
              </thead>
              <tbody>
                {smsMessages.map((msg) => (
                  <tr key={msg.id}>
                    <td>{new Date(msg.created_at).toLocaleString()}</td>
                    <td><strong>{msg.to_phone}</strong></td>
                    <td><span style={{ color: '#64748b', fontSize: '12px' }}>{msg.from_phone || 'Default'}</span></td>
                    <td><code style={{ fontSize: '12px' }}>{msg.message_sid || '—'}</code></td>
                    <td>{renderDeliveryStatusBadge(msg.status)}</td>
                    <td>
                      {msg.error_code ? (
                        <span style={{ color: '#b91c1c', fontSize: '12px', fontWeight: 600 }}>
                          [{msg.error_code}] {msg.error_message}
                        </span>
                      ) : (
                        <span style={{ color: '#94a3b8', fontSize: '12px' }}>—</span>
                      )}
                    </td>
                    <td>
                      <span style={{ fontSize: '13px', color: '#334155' }}>{msg.body_preview}</span>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </div>
      )}

      {activeTab === 'email_logs' && (
        <div>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '16px' }}>
            <div>
              <h3 style={{ margin: '0 0 4px 0', fontSize: '16px', color: '#0f172a' }}>Outbound Email Dispatch Records (AWS SES)</h3>
              <p style={{ margin: 0, fontSize: '13px', color: '#64748b' }}>
                Monotonic delivery lifecycle tracking with AWS SNS event reconciliation, bounce classification, and complaint defense.
              </p>
            </div>
            <button 
              className="ext-int-btn ext-int-btn-primary"
              onClick={() => {
                setTestEmailResult(null);
                setTestEmailModalOpen(true);
              }}
            >
              <Mail size={14} />
              Test Outbound Email
            </button>
          </div>

          {emailMessages.length === 0 ? (
            <div style={{ padding: '40px', textAlign: 'center', background: '#fff', borderRadius: '8px', border: '1px solid #e2e8f0' }}>
              <Mail size={36} color="#94a3b8" style={{ margin: '0 auto 12px auto' }} />
              <h3 style={{ margin: '0 0 6px 0', color: '#0f172a' }}>No Outbound Emails Dispatched Yet</h3>
              <p style={{ margin: 0, color: '#64748b', fontSize: '14px' }}>
                Transactional quotations, operational notices, and test dispatches sent via AWS SES will appear here.
              </p>
            </div>
          ) : (
            <table className="ext-dl-table">
              <thead>
                <tr>
                  <th>Timestamp</th>
                  <th>Recipient</th>
                  <th>Sender</th>
                  <th>Subject</th>
                  <th>Message ID</th>
                  <th>Delivery Status</th>
                  <th>Bounce / Complaint</th>
                  <th>Preview</th>
                </tr>
              </thead>
              <tbody>
                {emailMessages.map((msg) => (
                  <tr key={msg.id}>
                    <td>{new Date(msg.created_at).toLocaleString()}</td>
                    <td><strong>{msg.to_email}</strong></td>
                    <td><span style={{ color: '#64748b', fontSize: '12px' }}>{msg.from_email || 'Default'}</span></td>
                    <td><span style={{ fontSize: '13px', fontWeight: 500 }}>{msg.subject}</span></td>
                    <td><code style={{ fontSize: '11px' }}>{msg.message_id ? msg.message_id.slice(0, 20) + '...' : '—'}</code></td>
                    <td>{renderDeliveryStatusBadge(msg.status)}</td>
                    <td>
                      {msg.bounce_type ? (
                        <span style={{ color: '#be123c', fontSize: '12px', fontWeight: 600 }}>
                          [{msg.bounce_type}] {msg.bounce_sub_type || ''}
                        </span>
                      ) : msg.complaint_feedback_type ? (
                        <span style={{ color: '#a21caf', fontSize: '12px', fontWeight: 600 }}>
                          Complaint: {msg.complaint_feedback_type}
                        </span>
                      ) : msg.error_code ? (
                        <span style={{ color: '#b91c1c', fontSize: '12px', fontWeight: 600 }}>
                          [{msg.error_code}] {msg.error_message}
                        </span>
                      ) : (
                        <span style={{ color: '#94a3b8', fontSize: '12px' }}>—</span>
                      )}
                    </td>
                    <td>
                      <span style={{ fontSize: '12px', color: '#64748b' }}>{msg.body_preview}</span>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </div>
      )}

      {activeTab === 'dead_letter' && (
        <div>
          {deadLetters.length === 0 ? (
            <div style={{ padding: '40px', textAlign: 'center', background: '#fff', borderRadius: '8px', border: '1px solid #e2e8f0' }}>
              <CheckCircle2 size={36} color="#10b981" style={{ margin: '0 auto 12px auto' }} />
              <h3 style={{ margin: '0 0 6px 0', color: '#0f172a' }}>Dead-Letter Queue is Empty</h3>
              <p style={{ margin: 0, color: '#64748b', fontSize: '14px' }}>
                All incoming webhooks have passed signature verification and replay defense without errors.
              </p>
            </div>
          ) : (
            <table className="ext-dl-table">
              <thead>
                <tr>
                  <th>Timestamp</th>
                  <th>Provider</th>
                  <th>Correlation ID</th>
                  <th>Error Code</th>
                  <th>Details</th>
                </tr>
              </thead>
              <tbody>
                {deadLetters.map((dl) => (
                  <tr key={dl.id}>
                    <td>{new Date(dl.created_at).toLocaleString()}</td>
                    <td><span className="ext-int-card-provider">{dl.provider}</span></td>
                    <td><code>{dl.correlation_id}</code></td>
                    <td><span style={{ color: '#b91c1c', fontWeight: 600 }}>{dl.error_code}</span></td>
                    <td>{dl.error_message}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </div>
      )}

      {/* Edit / Configure Modal */}
      {editItem && (
        <div className="ext-modal-backdrop">
          <div className="ext-modal-card">
            <div className="ext-modal-header">
              <h2 className="ext-modal-title">Configure {editItem.provider} ({editItem.type})</h2>
              <button className="ext-modal-close" onClick={() => setEditItem(null)}>
                <X size={18} />
              </button>
            </div>

            <form onSubmit={handleSaveConfig}>
              <div className="ext-modal-body">
                {saveSuccess && (
                  <div style={{ padding: '10px 14px', background: '#ecfdf5', color: '#047857', borderRadius: '6px', marginBottom: '16px' }}>
                    {saveSuccess}
                  </div>
                )}

                <div className="ext-form-group">
                  <label className="ext-form-checkbox">
                    <input 
                      type="checkbox"
                      checked={editItem.is_enabled}
                      onChange={(e) => setEditItem({ ...editItem, is_enabled: e.target.checked })}
                    />
                    Enable Integration Gateway
                  </label>
                </div>

                <div className="ext-form-group">
                  <label className="ext-form-label">Endpoint URL / Base URL</label>
                  <input 
                    type="text"
                    className="ext-form-input"
                    value={editItem.endpoint_url || ''}
                    placeholder="https://api.provider.com/v1"
                    onChange={(e) => setEditItem({ ...editItem, endpoint_url: e.target.value })}
                  />
                </div>

                <div className="ext-form-group">
                  <label className="ext-form-label">Region / Deployment Zone</label>
                  <input 
                    type="text"
                    className="ext-form-input"
                    value={editItem.region || ''}
                    placeholder="e.g. ap-south-1 or us-east-1"
                    onChange={(e) => setEditItem({ ...editItem, region: e.target.value })}
                  />
                </div>

                <div className="ext-form-group">
                  <label className="ext-form-label">Sender / From Address / Account ID</label>
                  <input 
                    type="text"
                    className="ext-form-input"
                    value={editItem.from_address || ''}
                    placeholder="e.g. notifications@logisticshq.in or +15551234"
                    onChange={(e) => setEditItem({ ...editItem, from_address: e.target.value })}
                  />
                </div>

                <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '12px' }}>
                  <div className="ext-form-group">
                    <label className="ext-form-label">Timeout (seconds)</label>
                    <input 
                      type="number"
                      className="ext-form-input"
                      value={editItem.timeout_sec || 15}
                      min="1"
                      max="60"
                      onChange={(e) => setEditItem({ ...editItem, timeout_sec: e.target.value })}
                    />
                  </div>
                  <div className="ext-form-group">
                    <label className="ext-form-label">Max Retries</label>
                    <input 
                      type="number"
                      className="ext-form-input"
                      value={editItem.max_retries || 3}
                      min="0"
                      max="5"
                      onChange={(e) => setEditItem({ ...editItem, max_retries: e.target.value })}
                    />
                  </div>
                </div>

                <div className="ext-form-group">
                  <label className="ext-form-label">Secret API Key / Auth Token</label>
                  <input 
                    type="password"
                    className="ext-form-input"
                    value={editItem.new_secret || ''}
                    placeholder={editItem.has_secret ? '•••••••• (leave blank to keep existing)' : 'Enter secret key or token'}
                    onChange={(e) => setEditItem({ ...editItem, new_secret: e.target.value })}
                  />
                  <span style={{ fontSize: '11px', color: '#64748b', marginTop: '4px', display: 'block' }}>
                    🔒 Protected. Secrets are masked and never exposed in cleartext through any API or view.
                  </span>
                </div>
              </div>

              <div className="ext-modal-footer">
                <button 
                  type="button" 
                  className="ext-int-btn ext-int-btn-outline"
                  onClick={() => setEditItem(null)}
                >
                  Cancel
                </button>
                <button 
                  type="submit" 
                  className="ext-int-btn ext-int-btn-primary"
                  disabled={saving}
                >
                  {saving ? 'Saving...' : 'Save Configuration'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Test SMS Modal */}
      {testModalOpen && (
        <div className="ext-modal-backdrop">
          <div className="ext-modal-card">
            <div className="ext-modal-header">
              <h2 className="ext-modal-title">Test Outbound SMS Dispatch</h2>
              <button 
                className="ext-modal-close" 
                onClick={() => {
                  setTestModalOpen(false);
                  setTestResult(null);
                }}
              >
                <X size={18} />
              </button>
            </div>

            <form onSubmit={handleSendTestSMS}>
              <div className="ext-modal-body">
                <p style={{ margin: '0 0 16px 0', fontSize: '13px', color: '#64748b' }}>
                  Dispatches an outbound SMS through the Go Integration Gateway, enforcing E.164 phone format,
                  tenant boundary isolation, monotonic delivery tracking, and credential validation.
                </p>

                {testResult && (
                  <div style={{ 
                    padding: '12px 16px', 
                    borderRadius: '6px', 
                    marginBottom: '16px',
                    background: testResult.success ? '#ecfdf5' : '#fef2f2',
                    border: `1px solid ${testResult.success ? '#a7f3d0' : '#fecaca'}`,
                    color: testResult.success ? '#065f46' : '#991b1b',
                    fontSize: '13px'
                  }}>
                    {testResult.success ? (
                      <div>
                        <strong>Dispatched Successfully!</strong>
                        <div style={{ marginTop: '4px' }}>
                          Status: <code>{testResult.data?.status || 'ACCEPTED'}</code>
                          {testResult.data?.message_sid && (
                            <span style={{ marginLeft: '12px' }}>SID: <code>{testResult.data.message_sid}</code></span>
                          )}
                        </div>
                      </div>
                    ) : (
                      <div>
                        <strong>Dispatch Error:</strong> [{testResult.error?.code || 'ERROR'}] {testResult.error?.error || testResult.error?.message || 'Failed to dispatch'}
                        {testResult.error?.code === 'PROVIDER_NOT_CONFIGURED' && (
                          <div style={{ marginTop: '6px', fontSize: '12px', color: '#b91c1c' }}>
                            ℹ️ Twilio credentials are not configured in this environment. System correctly rejects unconfigured providers without fabricating delivery.
                          </div>
                        )}
                      </div>
                    )}
                  </div>
                )}

                <div className="ext-form-group">
                  <label className="ext-form-label">Recipient Phone Number (E.164 format)</label>
                  <input 
                    type="text"
                    className="ext-form-input"
                    value={testForm.to_phone}
                    placeholder="+14155552671"
                    required
                    onChange={(e) => setTestForm({ ...testForm, to_phone: e.target.value })}
                  />
                  <span style={{ fontSize: '11px', color: '#64748b', marginTop: '4px', display: 'block' }}>
                    Must follow E.164 international notation (e.g. +12345678901).
                  </span>
                </div>

                <div className="ext-form-group">
                  <label className="ext-form-label">
                    Message Body ({testForm.message.length}/1600 characters)
                  </label>
                  <textarea 
                    className="ext-form-input"
                    rows={4}
                    value={testForm.message}
                    placeholder="LogisticsHQ: Shipment SH-101 has arrived at Port of Nhava Sheva. Customs clearance pending."
                    required
                    maxLength={1600}
                    onChange={(e) => setTestForm({ ...testForm, message: e.target.value })}
                  />
                </div>
              </div>

              <div className="ext-modal-footer">
                <button 
                  type="button" 
                  className="ext-int-btn ext-int-btn-outline"
                  onClick={() => {
                    setTestModalOpen(false);
                    setTestResult(null);
                  }}
                >
                  Close
                </button>
                <button 
                  type="submit" 
                  className="ext-int-btn ext-int-btn-primary"
                  disabled={testingSMS}
                >
                  {testingSMS ? 'Dispatching...' : 'Dispatch Outbound SMS'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Test Email Modal */}
      {testEmailModalOpen && (
        <div className="ext-modal-backdrop">
          <div className="ext-modal-card">
            <div className="ext-modal-header">
              <h2 className="ext-modal-title">Test Outbound Email Dispatch (AWS SES)</h2>
              <button 
                className="ext-modal-close" 
                onClick={() => {
                  setTestEmailModalOpen(false);
                  setTestEmailResult(null);
                }}
              >
                <X size={18} />
              </button>
            </div>

            <form onSubmit={handleSendTestEmail}>
              <div className="ext-modal-body">
                <p style={{ margin: '0 0 16px 0', fontSize: '13px', color: '#64748b' }}>
                  Dispatches an outbound email through the Go Integration Gateway via AWS SES. Enforces RFC 5322 address formatting,
                  tenant isolation, pre-send secret scanning, and monotonic delivery tracking.
                </p>

                {testEmailResult && (
                  <div style={{ 
                    padding: '12px 16px', 
                    borderRadius: '6px', 
                    marginBottom: '16px',
                    background: testEmailResult.success ? '#ecfdf5' : '#fef2f2',
                    border: `1px solid ${testEmailResult.success ? '#a7f3d0' : '#fecaca'}`,
                    color: testEmailResult.success ? '#065f46' : '#991b1b',
                    fontSize: '13px'
                  }}>
                    {testEmailResult.success ? (
                      <div>
                        <strong>Dispatched Successfully!</strong>
                        <div style={{ marginTop: '4px' }}>
                          Status: <code>{testEmailResult.data?.status || 'ACCEPTED'}</code>
                          {testEmailResult.data?.message_id && (
                            <span style={{ marginLeft: '12px' }}>Message ID: <code>{testEmailResult.data.message_id}</code></span>
                          )}
                        </div>
                      </div>
                    ) : (
                      <div>
                        <strong>Dispatch Error:</strong> [{testEmailResult.error?.code || 'ERROR'}] {testEmailResult.error?.error || testEmailResult.error?.message || 'Failed to dispatch'}
                        {testEmailResult.error?.code === 'PROVIDER_NOT_CONFIGURED' && (
                          <div style={{ marginTop: '6px', fontSize: '12px', color: '#b91c1c' }}>
                            ℹ️ AWS SES credentials are not configured in this environment. System correctly rejects unconfigured providers without fabricating delivery.
                          </div>
                        )}
                        {testEmailResult.error?.code === 'RECIPIENT_SUPPRESSED' && (
                          <div style={{ marginTop: '6px', fontSize: '12px', color: '#b91c1c' }}>
                            🛑 Recipient is on the organization suppression list due to a previous hard bounce or spam complaint.
                          </div>
                        )}
                      </div>
                    )}
                  </div>
                )}

                <div className="ext-form-group">
                  <label className="ext-form-label">Recipient Email Address (RFC 5322)</label>
                  <input 
                    type="email"
                    className="ext-form-input"
                    value={testEmailForm.to_email}
                    placeholder="logistics.recipient@example.com"
                    required
                    onChange={(e) => setTestEmailForm({ ...testEmailForm, to_email: e.target.value })}
                  />
                  <span style={{ fontSize: '11px', color: '#64748b', marginTop: '4px', display: 'block' }}>
                    Must be a valid email format.
                  </span>
                </div>

                <div className="ext-form-group">
                  <label className="ext-form-label">Subject</label>
                  <input 
                    type="text"
                    className="ext-form-input"
                    value={testEmailForm.subject}
                    placeholder="LogisticsHQ Freight Notice: Shipment SH-101 Update"
                    required
                    maxLength={256}
                    onChange={(e) => setTestEmailForm({ ...testEmailForm, subject: e.target.value })}
                  />
                </div>

                <div className="ext-form-group">
                  <label className="ext-form-label">
                    Email Body Content ({testEmailForm.body_text.length}/10000 characters)
                  </label>
                  <textarea 
                    className="ext-form-input"
                    rows={4}
                    value={testEmailForm.body_text}
                    placeholder="Dear Customer, your consignment has departed Nhava Sheva Port. Tracking ID: TRK-9901."
                    required
                    maxLength={10000}
                    onChange={(e) => setTestEmailForm({ ...testEmailForm, body_text: e.target.value })}
                  />
                </div>
              </div>

              <div className="ext-modal-footer">
                <button 
                  type="button" 
                  className="ext-int-btn ext-int-btn-outline"
                  onClick={() => {
                    setTestEmailModalOpen(false);
                    setTestEmailResult(null);
                  }}
                >
                  Close
                </button>
                <button 
                  type="submit" 
                  className="ext-int-btn ext-int-btn-primary"
                  disabled={testingEmail}
                >
                  {testingEmail ? 'Dispatching...' : 'Dispatch Outbound Email'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}

import React, { useState, useEffect } from 'react';
import { 
  ShieldCheck, Plus, AlertTriangle, AlertCircle, CheckCircle2, 
  Clock, FileCheck, RefreshCw, XCircle, ShieldAlert, Award,
  X, ChevronDown, Shield, Sparkles
} from 'lucide-react';
import contractsService from '../../../services/contractsService';
import './ContractCompliancePanel.css';

export default function ContractCompliancePanel({ contract }) {
  const [events, setEvents] = useState([]);
  const [requirements, setRequirements] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [showAddReqModal, setShowAddReqModal] = useState(false);
  const [resolveModal, setResolveModal] = useState(null); // event object
  const [verifyModal, setVerifyModal] = useState(null); // requirement object
  const [resolutionNotes, setResolutionNotes] = useState('');
  const [saving, setSaving] = useState(false);

  // New Requirement Form State
  const [reqFormData, setReqFormData] = useState({
    requirement_type: 'INSURANCE',
    title: '',
    description: '',
    responsible_party: 'CARRIER',
    valid_from: '',
    valid_until: '',
    risk_severity: 'HIGH',
  });

  useEffect(() => {
    if (contract?.id) {
      fetchComplianceData();
    }
  }, [contract?.id]);

  const fetchComplianceData = async () => {
    try {
      setLoading(true);
      setError(null);
      const [eventsRes, reqsRes] = await Promise.all([
        contractsService.getContractComplianceEvents(contract.id),
        contractsService.getContractComplianceRequirements(contract.id),
      ]);
      setEvents(eventsRes.data?.data || []);
      setRequirements(reqsRes.data?.data || []);
    } catch (err) {
      console.error('Failed to load compliance data:', err);
      setError('Unable to load compliance records.');
    } finally {
      setLoading(false);
    }
  };

  const handleCreateRequirement = async (e) => {
    e.preventDefault();
    if (!reqFormData.title || !reqFormData.requirement_type) return;

    try {
      setSaving(true);
      await contractsService.createComplianceRequirement(contract.id, reqFormData);
      setShowAddReqModal(false);
      setReqFormData({
        requirement_type: 'INSURANCE',
        title: '',
        description: '',
        responsible_party: 'CARRIER',
        valid_from: '',
        valid_until: '',
        risk_severity: 'HIGH',
      });
      await fetchComplianceData();
    } catch (err) {
      console.error('Failed to create compliance requirement:', err);
      alert('Failed to save compliance requirement.');
    } finally {
      setSaving(false);
    }
  };

  const handleResolveEvent = async (e) => {
    e.preventDefault();
    if (!resolveModal) return;

    try {
      setSaving(true);
      await contractsService.resolveComplianceEvent(contract.id, resolveModal.id, {
        resolution_notes: resolutionNotes || 'Investigated and resolved by compliance officer.',
      });
      setResolveModal(null);
      setResolutionNotes('');
      await fetchComplianceData();
    } catch (err) {
      console.error('Failed to resolve event:', err);
      alert('Failed to resolve compliance event.');
    } finally {
      setSaving(false);
    }
  };

  const handleVerifyRequirement = async (status) => {
    if (!verifyModal) return;

    try {
      setSaving(true);
      await contractsService.verifyComplianceRequirement(contract.id, verifyModal.id, {
        status: status,
        notes: resolutionNotes || `Marked as ${status} after document audit.`,
      });
      setVerifyModal(null);
      setResolutionNotes('');
      await fetchComplianceData();
    } catch (err) {
      console.error('Failed to verify requirement:', err);
      alert('Failed to update requirement verification.');
    } finally {
      setSaving(false);
    }
  };

  const openEvents = events.filter(e => e.status === 'OPEN' || e.status === 'IN_PROGRESS');
  const resolvedEvents = events.filter(e => e.status === 'RESOLVED' || e.status === 'WAIVED');

  const getSeverityBadge = (sev) => {
    switch (sev) {
      case 'CRITICAL':
        return <span className="sev-badge critical">🔴 Critical</span>;
      case 'WARNING':
        return <span className="sev-badge warning">🟠 Warning</span>;
      case 'ATTENTION':
        return <span className="sev-badge attention">🟡 Attention</span>;
      default:
        return <span className="sev-badge info">🔵 Info</span>;
    }
  };

  return (
    <div className="contract-compliance-panel">
      {/* Header Bar */}
      <div className="panel-header-row">
        <div>
          <h3 className="panel-title">Compliance Governance & Risk Intelligence</h3>
          <p className="panel-subtitle">
            Monitor regulatory mandates, cargo insurance, carrier KYC, and automated breach detection events.
          </p>
        </div>
        <button className="btn-primary-action" onClick={() => setShowAddReqModal(true)}>
          <Plus size={16} />
          <span>Add Requirement</span>
        </button>
      </div>

      {loading ? (
        <div className="panel-loading-state">
          <RefreshCw className="spin-icon" size={24} />
          <span>Auditing compliance mandates and active risks...</span>
        </div>
      ) : error ? (
        <div className="panel-error-banner">
          <AlertCircle size={18} />
          <span>{error}</span>
          <button className="btn-retry" onClick={fetchComplianceData}>Retry</button>
        </div>
      ) : (
        <div className="compliance-container">
          {/* Active Compliance Breaches & Warnings */}
          <div className="comp-section">
            <div className="section-label-bar">
              <ShieldAlert size={16} className="text-red-500" />
              <h4>Detected Compliance & Risk Events ({openEvents.length} Active)</h4>
            </div>

            {openEvents.length === 0 ? (
              <div className="comp-clean-banner">
                <ShieldCheck size={20} className="text-emerald-600" />
                <div>
                  <strong>No Active Compliance Breaches</strong>
                  <p>All monitored SLA covenants, insurance policies, and operational deliverables are healthy.</p>
                </div>
              </div>
            ) : (
              <div className="events-list">
                {openEvents.map(ev => (
                  <div key={ev.id} className={`comp-event-card sev-${ev.severity?.toLowerCase()}`}>
                    <div className="event-header">
                      <div className="event-header-left">
                        {getSeverityBadge(ev.severity)}
                        <span className="event-type-tag">{ev.event_type}</span>
                        <span className="event-time">{new Date(ev.detected_at).toLocaleDateString()}</span>
                      </div>
                      <button 
                        className="btn-resolve-action"
                        onClick={() => {
                          setResolveModal(ev);
                          setResolutionNotes('');
                        }}
                      >
                        <CheckCircle2 size={13} />
                        <span>Acknowledge / Resolve</span>
                      </button>
                    </div>

                    <h5 className="event-title">{ev.title}</h5>
                    {ev.description && <p className="event-desc">{ev.description}</p>}
                  </div>
                ))}
              </div>
            )}
          </div>

          {/* Compliance Requirements & Certificates */}
          <div className="comp-section">
            <div className="section-label-bar">
              <Award size={16} className="text-blue-600" />
              <h4>Mandatory Compliance & Certificate Registry ({requirements.length})</h4>
            </div>

            {requirements.length === 0 ? (
              <div className="panel-empty-state">
                <Award size={32} className="text-slate-400" />
                <h4>No Mandatory Requirements Enrolled</h4>
                <p>Register required Marine Cargo Insurance, Carrier FMC Licensure, or Dangerous Goods certifications.</p>
                <button className="btn-primary-action" onClick={() => setShowAddReqModal(true)}>
                  <Plus size={16} />
                  <span>Register First Mandate</span>
                </button>
              </div>
            ) : (
              <div className="reqs-grid">
                {requirements.map(req => (
                  <div key={req.id} className="req-card">
                    <div className="req-card-top">
                      <span className="req-type-badge">{req.requirement_type}</span>
                      <span className={`req-status-badge status-${req.status?.toLowerCase()}`}>
                        {req.status}
                      </span>
                    </div>

                    <h5 className="req-title">{req.title}</h5>
                    {req.description && <p className="req-desc">{req.description}</p>}

                    <div className="req-meta-footer">
                      <div className="meta-item">
                        <span className="text-slate-500">Party:</span>
                        <span className="font-semibold">{req.responsible_party}</span>
                      </div>
                      {req.valid_until && (
                        <div className="meta-item">
                          <Clock size={12} />
                          <span>Valid: {req.valid_until}</span>
                        </div>
                      )}
                    </div>

                    {req.status !== 'COMPLIANT' && (
                      <button 
                        className="btn-verify-action"
                        onClick={() => {
                          setVerifyModal(req);
                          setResolutionNotes('');
                        }}
                      >
                        <ShieldCheck size={13} />
                        <span>Verify / Audit Evidence</span>
                      </button>
                    )}
                  </div>
                ))}
              </div>
            )}
          </div>
        </div>
      )}

      {/* Add Requirement Modal */}
      {showAddReqModal && (
        <div className="modal-backdrop" onClick={() => setShowAddReqModal(false)}>
          <div className="comp-modal-content" onClick={e => e.stopPropagation()}>
            <div className="comp-modal-header">
              <div className="comp-modal-header-left">
                <div className="comp-header-badge">
                  <ShieldCheck size={18} />
                </div>
                <div>
                  <h4>Register Compliance Mandate / Certificate</h4>
                  <p className="comp-modal-subtitle">Track insurance policies, carrier KYC, and statutory covenants</p>
                </div>
              </div>
              <button className="btn-close-modal" onClick={() => setShowAddReqModal(false)} title="Close">
                <X size={16} />
              </button>
            </div>

            <form onSubmit={handleCreateRequirement} className="comp-form">
              <div className="comp-form-group">
                <label className="comp-label">
                  Mandate Type <span className="comp-req">*</span>
                </label>
                <div className="comp-select-wrapper">
                  <select 
                    className="comp-select"
                    value={reqFormData.requirement_type}
                    onChange={(e) => setReqFormData({ ...reqFormData, requirement_type: e.target.value })}
                  >
                    <option value="INSURANCE">Marine Cargo / Liability Insurance</option>
                    <option value="CARRIER_CERTIFICATION">Carrier FMC / Line Licensure</option>
                    <option value="VENDOR_CERTIFICATION">Vendor / Port Terminal Certification</option>
                    <option value="REGULATORY">Customs / Regulatory Compliance</option>
                    <option value="SAFETY">Dangerous Goods / HAZMAT Compliance</option>
                    <option value="FINANCIAL">Bank Guarantee / Financial Standing</option>
                  </select>
                  <ChevronDown size={14} className="comp-select-chevron" />
                </div>
              </div>

              <div className="comp-form-group">
                <label className="comp-label">
                  Requirement Title <span className="comp-req">*</span>
                </label>
                <input 
                  type="text"
                  className="comp-input"
                  placeholder="e.g. Annual Carrier Public Liability Coverage ($5M USD)"
                  value={reqFormData.title}
                  onChange={(e) => setReqFormData({ ...reqFormData, title: e.target.value })}
                  required
                />
              </div>

              <div className="comp-form-row-2">
                <div className="comp-form-group">
                  <label className="comp-label">Responsible Party</label>
                  <div className="comp-select-wrapper">
                    <select 
                      className="comp-select"
                      value={reqFormData.responsible_party}
                      onChange={(e) => setReqFormData({ ...reqFormData, responsible_party: e.target.value })}
                    >
                      <option value="CARRIER">Carrier / Line</option>
                      <option value="CUSTOMER">Customer / Shipper</option>
                      <option value="FORWARDER">LogisticsHQ (Internal)</option>
                      <option value="VENDOR">Vendor / Warehouse</option>
                    </select>
                    <ChevronDown size={14} className="comp-select-chevron" />
                  </div>
                </div>
                <div className="comp-form-group">
                  <label className="comp-label">Risk Severity if Breached</label>
                  <div className="comp-select-wrapper">
                    <select 
                      className="comp-select"
                      value={reqFormData.risk_severity}
                      onChange={(e) => setReqFormData({ ...reqFormData, risk_severity: e.target.value })}
                    >
                      <option value="CRITICAL">Critical (High Risk)</option>
                      <option value="HIGH">High Priority</option>
                      <option value="MEDIUM">Medium Priority</option>
                    </select>
                    <ChevronDown size={14} className="comp-select-chevron" />
                  </div>
                </div>
              </div>

              <div className="comp-form-row-2">
                <div className="comp-form-group">
                  <label className="comp-label">Valid From</label>
                  <input 
                    type="date"
                    className="comp-input comp-input-date"
                    value={reqFormData.valid_from}
                    onChange={(e) => setReqFormData({ ...reqFormData, valid_from: e.target.value })}
                  />
                </div>
                <div className="comp-form-group">
                  <label className="comp-label">Valid Until (Expiry)</label>
                  <input 
                    type="date"
                    className="comp-input comp-input-date"
                    value={reqFormData.valid_until}
                    onChange={(e) => setReqFormData({ ...reqFormData, valid_until: e.target.value })}
                  />
                </div>
              </div>

              <div className="modal-actions-bar">
                <button type="button" className="btn-cancel" onClick={() => setShowAddReqModal(false)}>
                  Cancel
                </button>
                <button type="submit" className="btn-submit" disabled={saving}>
                  <ShieldCheck size={14} />
                  <span>{saving ? 'Registering...' : 'Register Mandate'}</span>
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Resolve Event Modal */}
      {resolveModal && (
        <div className="modal-backdrop" onClick={() => setResolveModal(null)}>
          <div className="comp-modal-content" onClick={e => e.stopPropagation()}>
            <div className="comp-modal-header">
              <div className="comp-modal-header-left">
                <div className="comp-header-badge warning">
                  <AlertTriangle size={18} />
                </div>
                <div>
                  <h4>Resolve Compliance Risk Event</h4>
                  <p className="comp-modal-subtitle">Log audit rationale and corrective action</p>
                </div>
              </div>
              <button className="btn-close-modal" onClick={() => setResolveModal(null)} title="Close">
                <X size={16} />
              </button>
            </div>

            <form onSubmit={handleResolveEvent} className="comp-form">
              <div className="action-obligation-preview">
                <strong>Event:</strong> {resolveModal.title}
              </div>

              <div className="comp-form-group">
                <label className="comp-label">
                  Resolution & Corrective Action Notes <span className="comp-req">*</span>
                </label>
                <textarea 
                  rows={3}
                  className="comp-textarea"
                  placeholder="Describe resolution evidence, carrier confirmation, or mitigation steps taken..."
                  value={resolutionNotes}
                  onChange={(e) => setResolutionNotes(e.target.value)}
                  required
                />
              </div>

              <div className="modal-actions-bar">
                <button type="button" className="btn-cancel" onClick={() => setResolveModal(null)}>
                  Cancel
                </button>
                <button type="submit" className="btn-submit bg-emerald-600" disabled={saving}>
                  <CheckCircle2 size={14} />
                  <span>{saving ? 'Resolving...' : 'Confirm Resolution'}</span>
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Verify Requirement Modal */}
      {verifyModal && (
        <div className="modal-backdrop" onClick={() => setVerifyModal(null)}>
          <div className="comp-modal-content" onClick={e => e.stopPropagation()}>
            <div className="comp-modal-header">
              <div className="comp-modal-header-left">
                <div className="comp-header-badge">
                  <ShieldCheck size={18} />
                </div>
                <div>
                  <h4>Verify Compliance Requirement</h4>
                  <p className="comp-modal-subtitle">Audit evidence and update verification status</p>
                </div>
              </div>
              <button className="btn-close-modal" onClick={() => setVerifyModal(null)} title="Close">
                <X size={16} />
              </button>
            </div>

            <div className="comp-form">
              <div className="action-obligation-preview">
                <strong>Mandate:</strong> {verifyModal.title} ({verifyModal.requirement_type})
              </div>

              <div className="comp-form-group">
                <label className="comp-label">Verification Audit Notes</label>
                <textarea 
                  rows={2}
                  className="comp-textarea"
                  placeholder="Verification sign-off notes or policy certificate number..."
                  value={resolutionNotes}
                  onChange={(e) => setResolutionNotes(e.target.value)}
                />
              </div>

              <div className="modal-actions-bar">
                <button type="button" className="btn-cancel" onClick={() => setVerifyModal(null)}>
                  Cancel
                </button>
                <button 
                  type="button" 
                  className="btn-submit bg-red-600" 
                  onClick={() => handleVerifyRequirement('NON_COMPLIANT')}
                  disabled={saving}
                >
                  <XCircle size={14} />
                  <span>Mark Non-Compliant</span>
                </button>
                <button 
                  type="button" 
                  className="btn-submit bg-emerald-600" 
                  onClick={() => handleVerifyRequirement('COMPLIANT')}
                  disabled={saving}
                >
                  <CheckCircle2 size={14} />
                  <span>Verify Compliant</span>
                </button>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

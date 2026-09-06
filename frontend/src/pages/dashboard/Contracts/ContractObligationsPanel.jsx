import React, { useState, useEffect } from 'react';
import { 
  CheckSquare, Plus, AlertCircle, Clock, CheckCircle2, 
  XCircle, User, Flag, RefreshCw, Award, ArrowUpRight, ShieldCheck,
  X, ChevronDown, Sparkles, Shield
} from 'lucide-react';
import contractsService from '../../../services/contractsService';
import './ContractObligationsPanel.css';

export default function ContractObligationsPanel({ contract }) {
  const [obligations, setObligations] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [showAddModal, setShowAddModal] = useState(false);
  const [actionModal, setActionModal] = useState(null); // { type: 'fulfill' | 'waive', obligation: ob }
  const [actionNotes, setActionNotes] = useState('');
  const [saving, setSaving] = useState(false);

  // Form State
  const [formData, setFormData] = useState({
    obligation_reference: '',
    title: '',
    description: '',
    obligation_type: 'VOLUME_COMMITMENT',
    category: 'SHIPMENT',
    responsible_party: 'CARRIER',
    owner: '',
    priority: 'HIGH',
    due_date: '',
    is_recurring: false,
    recurrence_type: 'NONE',
    target_value: '',
    target_unit: 'TEU',
    warning_threshold: '',
    critical_threshold: '',
    notes: '',
  });

  useEffect(() => {
    if (contract?.id) {
      fetchObligations();
    }
  }, [contract?.id]);

  const fetchObligations = async () => {
    try {
      setLoading(true);
      setError(null);
      const res = await contractsService.getContractObligations(contract.id);
      setObligations(res.data?.data || []);
    } catch (err) {
      console.error('Failed to load contract obligations:', err);
      setError('Unable to load contractual obligations.');
    } finally {
      setLoading(false);
    }
  };

  const handleCreateObligation = async (e) => {
    e.preventDefault();
    if (!formData.title || !formData.obligation_reference) return;

    try {
      setSaving(true);
      const payload = {
        ...formData,
        target_value: formData.target_value ? parseFloat(formData.target_value) : null,
        warning_threshold: formData.warning_threshold ? parseFloat(formData.warning_threshold) : null,
        critical_threshold: formData.critical_threshold ? parseFloat(formData.critical_threshold) : null,
        due_date: formData.due_date || null,
      };

      await contractsService.createContractObligation(contract.id, payload);
      setShowAddModal(false);
      setFormData({
        obligation_reference: '',
        title: '',
        description: '',
        obligation_type: 'VOLUME_COMMITMENT',
        category: 'SHIPMENT',
        responsible_party: 'CARRIER',
        owner: '',
        priority: 'HIGH',
        due_date: '',
        is_recurring: false,
        recurrence_type: 'NONE',
        target_value: '',
        target_unit: 'TEU',
        warning_threshold: '',
        critical_threshold: '',
        notes: '',
      });
      await fetchObligations();
    } catch (err) {
      console.error('Failed to add obligation:', err);
      alert('Failed to add obligation.');
    } finally {
      setSaving(false);
    }
  };

  const handleActionSubmit = async (e) => {
    e.preventDefault();
    if (!actionModal) return;

    try {
      setSaving(true);
      if (actionModal.type === 'fulfill') {
        await contractsService.fulfillContractObligation(contract.id, actionModal.obligation.id, {
          notes: actionNotes || 'Fulfilled and verified by commercial lead.',
        });
      } else if (actionModal.type === 'waive') {
        await contractsService.waiveContractObligation(contract.id, actionModal.obligation.id, {
          reason: actionNotes || 'Formally waived per operational amendment.',
        });
      }
      setActionModal(null);
      setActionNotes('');
      await fetchObligations();
    } catch (err) {
      console.error('Failed to update obligation:', err);
      alert('Failed to process obligation action.');
    } finally {
      setSaving(false);
    }
  };

  // Metrics
  const total = obligations.length;
  const fulfilled = obligations.filter(o => o.status === 'FULFILLED' || o.status === 'COMPLETED').length;
  const overdue = obligations.filter(o => o.status === 'OVERDUE' || o.status === 'BREACHED').length;
  const dueSoon = obligations.filter(o => o.status === 'DUE_SOON').length;
  const active = obligations.filter(o => o.status === 'ACTIVE').length;

  const getPriorityBadgeClass = (p) => {
    switch (p) {
      case 'CRITICAL': return 'priority-critical';
      case 'HIGH': return 'priority-high';
      case 'MEDIUM': return 'priority-medium';
      default: return 'priority-low';
    }
  };

  const getStatusBadgeClass = (s) => {
    switch (s) {
      case 'FULFILLED':
      case 'COMPLETED':
        return 'status-fulfilled';
      case 'OVERDUE':
      case 'BREACHED':
        return 'status-overdue';
      case 'DUE_SOON':
        return 'status-duesoon';
      case 'WAIVED':
        return 'status-waived';
      default:
        return 'status-active';
    }
  };

  return (
    <div className="contract-obligations-panel">
      {/* Header Bar */}
      <div className="panel-header-row">
        <div>
          <h3 className="panel-title">Contractual Obligations & SLA Tracking</h3>
          <p className="panel-subtitle">
            Track commitments, minimum volume covenants, free time guarantees, and operational deliverables.
          </p>
        </div>
        <button 
          className="btn-primary-action"
          onClick={() => {
            setFormData({
              ...formData,
              obligation_reference: `OBL-${contract?.contract_reference || 'CTR'}-${obligations.length + 1}`,
              owner: contract?.owner || '',
            });
            setShowAddModal(true);
          }}
        >
          <Plus size={16} />
          <span>New Obligation</span>
        </button>
      </div>

      {/* KPI Status Strip */}
      <div className="obligations-kpi-strip">
        <div className="ob-kpi-card">
          <span className="ob-kpi-label">Total Tracked</span>
          <span className="ob-kpi-val">{total}</span>
        </div>
        <div className="ob-kpi-card active-border">
          <span className="ob-kpi-label">Active & Ongoing</span>
          <span className="ob-kpi-val text-blue-600">{active}</span>
        </div>
        <div className="ob-kpi-card duesoon-border">
          <span className="ob-kpi-label">Due Soon (&lt;7d)</span>
          <span className="ob-kpi-val text-amber-600">{dueSoon}</span>
        </div>
        <div className="ob-kpi-card overdue-border">
          <span className="ob-kpi-label">Overdue / Breached</span>
          <span className="ob-kpi-val text-red-600">{overdue}</span>
        </div>
        <div className="ob-kpi-card fulfilled-border">
          <span className="ob-kpi-label">Fulfilled</span>
          <span className="ob-kpi-val text-emerald-600">{fulfilled}</span>
        </div>
      </div>

      {loading ? (
        <div className="panel-loading-state">
          <RefreshCw className="spin-icon" size={24} />
          <span>Evaluating obligation timelines and SLAs...</span>
        </div>
      ) : error ? (
        <div className="panel-error-banner">
          <AlertCircle size={18} />
          <span>{error}</span>
          <button className="btn-retry" onClick={fetchObligations}>Retry</button>
        </div>
      ) : obligations.length === 0 ? (
        <div className="panel-empty-state">
          <div className="empty-icon-circle">
            <CheckSquare size={32} />
          </div>
          <h4>No Obligations Defined</h4>
          <p>Commitments like volume milestones, insurance renewals, and payment deadlines can be tracked here with automated alerts.</p>
          <button 
            className="btn-primary-action"
            onClick={() => {
              setFormData({
                ...formData,
                obligation_reference: `OBL-${contract?.contract_reference || 'CTR'}-1`,
                owner: contract?.owner || '',
              });
              setShowAddModal(true);
            }}
          >
            <Plus size={16} />
            <span>Create First Obligation</span>
          </button>
        </div>
      ) : (
        <div className="obligations-list">
          {obligations.map((ob) => (
            <div key={ob.id} className={`obligation-card ${getStatusBadgeClass(ob.status)}-card`}>
              <div className="obligation-card-header">
                <div className="ob-header-left">
                  <span className="ob-ref">{ob.obligation_reference}</span>
                  <span className={`ob-priority-badge ${getPriorityBadgeClass(ob.priority)}`}>
                    <Flag size={11} />
                    {ob.priority}
                  </span>
                  <span className={`ob-status-badge ${getStatusBadgeClass(ob.status)}`}>
                    {ob.status?.replace(/_/g, ' ')}
                  </span>
                </div>

                <div className="ob-header-right">
                  {ob.status !== 'FULFILLED' && ob.status !== 'WAIVED' && (
                    <div className="ob-actions-group">
                      <button 
                        className="btn-ob-action fulfill"
                        onClick={() => {
                          setActionModal({ type: 'fulfill', obligation: ob });
                          setActionNotes('');
                        }}
                        title="Mark obligation fulfilled / completed"
                      >
                        <CheckCircle2 size={13} />
                        <span>Fulfill</span>
                      </button>
                      <button 
                        className="btn-ob-action waive"
                        onClick={() => {
                          setActionModal({ type: 'waive', obligation: ob });
                          setActionNotes('');
                        }}
                        title="Waive requirement per mutual consent"
                      >
                        <XCircle size={13} />
                        <span>Waive</span>
                      </button>
                    </div>
                  )}
                </div>
              </div>

              <h4 className="ob-title">{ob.title}</h4>
              {ob.description && <p className="ob-desc">{ob.description}</p>}

              {/* Progress Bar for Volume / Target commitments */}
              {ob.target_value && (
                <div className="ob-progress-section">
                  <div className="ob-progress-header">
                    <span>Target Commitment: {ob.target_value} {ob.target_unit || 'Units'}</span>
                    <span>Current: {ob.current_value || 0} {ob.target_unit || 'Units'} ({Math.min(100, Math.round(((ob.current_value || 0) / ob.target_value) * 100))}%)</span>
                  </div>
                  <div className="ob-progress-track">
                    <div 
                      className="ob-progress-fill" 
                      style={{ width: `${Math.min(100, Math.round(((ob.current_value || 0) / ob.target_value) * 100))}%` }}
                    />
                  </div>
                </div>
              )}

              <div className="ob-meta-grid">
                <div className="ob-meta-col">
                  <span className="meta-label">Responsible Party</span>
                  <span className="meta-val font-semibold">{ob.responsible_party}</span>
                </div>
                <div className="ob-meta-col">
                  <span className="meta-label">Assigned Owner</span>
                  <span className="meta-val">{ob.owner || 'Unassigned'}</span>
                </div>
                <div className="ob-meta-col">
                  <span className="meta-label">Due Date</span>
                  <span className={`meta-val ${ob.status === 'OVERDUE' ? 'text-red-600 font-bold' : ''}`}>
                    {ob.due_date ? ob.due_date : 'No deadline'}
                  </span>
                </div>
                <div className="ob-meta-col">
                  <span className="meta-label">Type</span>
                  <span className="meta-val">{ob.obligation_type?.replace(/_/g, ' ')}</span>
                </div>
              </div>

              {ob.notes && (
                <div className="ob-notes-footer">
                  <span className="notes-label">Audit Notes:</span>
                  <span>{ob.notes}</span>
                </div>
              )}
            </div>
          ))}
        </div>
      )}

      {/* Create Obligation Modal */}
      {showAddModal && (
        <div className="modal-backdrop" onClick={() => setShowAddModal(false)}>
          <div className="ob-modal-content" onClick={e => e.stopPropagation()}>
            <div className="ob-modal-header">
              <div className="ob-modal-header-left">
                <div className="ob-header-badge">
                  <ShieldCheck size={18} />
                </div>
                <div>
                  <h4>Create Contractual Obligation / SLA</h4>
                  <p className="ob-modal-subtitle">Define operational covenants, performance deliverables, and volume targets</p>
                </div>
              </div>
              <button className="btn-close-modal" onClick={() => setShowAddModal(false)} title="Close">
                <X size={16} />
              </button>
            </div>

            <form onSubmit={handleCreateObligation} className="ob-form">
              {/* Row 1: Reference & Type */}
              <div className="ob-form-row-2">
                <div className="ob-form-group">
                  <label className="ob-label">
                    Obligation Reference <span className="ob-req">*</span>
                  </label>
                  <input 
                    type="text"
                    className="ob-input ob-input-mono"
                    value={formData.obligation_reference}
                    onChange={(e) => setFormData({ ...formData, obligation_reference: e.target.value })}
                    placeholder="e.g. OBL-2026-001"
                    required
                  />
                </div>
                <div className="ob-form-group">
                  <label className="ob-label">
                    Obligation Type <span className="ob-req">*</span>
                  </label>
                  <div className="ob-select-wrapper">
                    <select 
                      className="ob-select"
                      value={formData.obligation_type}
                      onChange={(e) => setFormData({ ...formData, obligation_type: e.target.value })}
                    >
                      <option value="VOLUME_COMMITMENT">Minimum Volume Commitment (FCL / TEU)</option>
                      <option value="TRANSIT_TIME">Transit Time Guarantee (Days)</option>
                      <option value="SLA">Service Level Agreement (SLA)</option>
                      <option value="DOCUMENTATION">Documentation Deliverable (BL, Invoices)</option>
                      <option value="INSURANCE">Insurance Policy Renewal</option>
                      <option value="FREE_DAYS">Free Time Guarantee (Detention / Demurrage)</option>
                      <option value="PAYMENT">Payment Term Compliance</option>
                      <option value="PRICING_REVIEW">Quarterly Pricing Review Notice</option>
                      <option value="OPERATIONAL">Operational Protocol</option>
                    </select>
                    <ChevronDown size={14} className="ob-select-chevron" />
                  </div>
                </div>
              </div>

              {/* Row 2: Title */}
              <div className="ob-form-group">
                <label className="ob-label">
                  Obligation Title <span className="ob-req">*</span>
                </label>
                <input 
                  type="text"
                  className="ob-input"
                  placeholder="e.g. Annual committed volume of 2,500 TEU on Asia-US West Coast"
                  value={formData.title}
                  onChange={(e) => setFormData({ ...formData, title: e.target.value })}
                  required
                />
              </div>

              {/* Row 3: Description */}
              <div className="ob-form-group">
                <label className="ob-label">Detailed Clause Description</label>
                <textarea 
                  rows={2}
                  className="ob-textarea"
                  placeholder="Specific covenants, milestone requirements, or breach thresholds..."
                  value={formData.description}
                  onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                />
              </div>

              {/* Row 4: Responsibility & Governance */}
              <div className="ob-form-row-3">
                <div className="ob-form-group">
                  <label className="ob-label">
                    Responsible Party <span className="ob-req">*</span>
                  </label>
                  <div className="ob-select-wrapper">
                    <select 
                      className="ob-select"
                      value={formData.responsible_party}
                      onChange={(e) => setFormData({ ...formData, responsible_party: e.target.value })}
                    >
                      <option value="CARRIER">Carrier / Line</option>
                      <option value="CUSTOMER">Customer / Shipper</option>
                      <option value="FORWARDER">LogisticsHQ (Forwarder)</option>
                      <option value="VENDOR">Vendor / Warehouse</option>
                    </select>
                    <ChevronDown size={14} className="ob-select-chevron" />
                  </div>
                </div>
                <div className="ob-form-group">
                  <label className="ob-label">Priority</label>
                  <div className="ob-select-wrapper">
                    <select 
                      className="ob-select"
                      value={formData.priority}
                      onChange={(e) => setFormData({ ...formData, priority: e.target.value })}
                    >
                      <option value="CRITICAL">Critical (High Risk)</option>
                      <option value="HIGH">High Priority</option>
                      <option value="MEDIUM">Medium Priority</option>
                      <option value="LOW">Low / Routine</option>
                    </select>
                    <ChevronDown size={14} className="ob-select-chevron" />
                  </div>
                </div>
                <div className="ob-form-group">
                  <label className="ob-label">Assigned Owner</label>
                  <input 
                    type="text"
                    className="ob-input"
                    placeholder="e.g. Varun Kanade"
                    value={formData.owner}
                    onChange={(e) => setFormData({ ...formData, owner: e.target.value })}
                  />
                </div>
              </div>

              {/* Row 5: Metric Targets & Due Date */}
              <div className="ob-form-row-3">
                <div className="ob-form-group">
                  <label className="ob-label">Target Value</label>
                  <input 
                    type="number"
                    step="0.1"
                    className="ob-input"
                    placeholder="e.g. 2500"
                    value={formData.target_value}
                    onChange={(e) => setFormData({ ...formData, target_value: e.target.value })}
                  />
                </div>
                <div className="ob-form-group">
                  <label className="ob-label">Unit</label>
                  <input 
                    type="text"
                    className="ob-input"
                    placeholder="e.g. TEU, Days, USD"
                    value={formData.target_unit}
                    onChange={(e) => setFormData({ ...formData, target_unit: e.target.value })}
                  />
                </div>
                <div className="ob-form-group">
                  <label className="ob-label">Target Due Date</label>
                  <input 
                    type="date"
                    className="ob-input ob-input-date"
                    value={formData.due_date}
                    onChange={(e) => setFormData({ ...formData, due_date: e.target.value })}
                  />
                </div>
              </div>

              <div className="modal-actions-bar">
                <button type="button" className="btn-cancel" onClick={() => setShowAddModal(false)}>
                  Cancel
                </button>
                <button type="submit" className="btn-submit" disabled={saving}>
                  <CheckCircle2 size={14} />
                  <span>{saving ? 'Creating...' : 'Create Obligation'}</span>
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Action Modal (Fulfill / Waive) */}
      {actionModal && (
        <div className="modal-backdrop" onClick={() => setActionModal(null)}>
          <div className="ob-modal-content" onClick={e => e.stopPropagation()}>
            <div className="ob-modal-header">
              <div className="ob-modal-header-left">
                <div className={`ob-header-badge ${actionModal.type === 'fulfill' ? 'fulfill' : 'waive'}`}>
                  {actionModal.type === 'fulfill' ? <CheckCircle2 size={18} /> : <AlertCircle size={18} />}
                </div>
                <div>
                  <h4>{actionModal.type === 'fulfill' ? 'Fulfill Contract Obligation' : 'Waive Obligation'}</h4>
                  <p className="ob-modal-subtitle">Log audit rationale and operational outcome</p>
                </div>
              </div>
              <button className="btn-close-modal" onClick={() => setActionModal(null)} title="Close">
                <X size={16} />
              </button>
            </div>

            <form onSubmit={handleActionSubmit} className="ob-form">
              <div className="action-obligation-preview">
                <strong>{actionModal.obligation.obligation_reference}:</strong> {actionModal.obligation.title}
              </div>

              <div className="ob-form-group">
                <label className="ob-label">{actionModal.type === 'fulfill' ? 'Verification & Fulfillment Audit Notes' : 'Reason for Waiver *'}</label>
                <textarea 
                  rows={3}
                  className="ob-textarea"
                  placeholder={actionModal.type === 'fulfill' ? 'Details of fulfillment evidence, operational compliance, or sign-off...' : 'Reason for formal waiver per mutual agreement...'}
                  value={actionNotes}
                  onChange={(e) => setActionNotes(e.target.value)}
                  required={actionModal.type === 'waive'}
                />
              </div>

              <div className="modal-actions-bar">
                <button type="button" className="btn-cancel" onClick={() => setActionModal(null)}>Cancel</button>
                <button 
                  type="submit" 
                  className={`btn-submit ${actionModal.type === 'waive' ? 'bg-amber-600' : 'bg-emerald-600'}`} 
                  disabled={saving}
                >
                  {saving ? 'Processing...' : actionModal.type === 'fulfill' ? 'Mark Fulfilled' : 'Confirm Waiver'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}

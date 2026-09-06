import React, { useState, useEffect, useCallback } from 'react';
import { 
  FileEdit, Plus, Send, CheckCircle2, XCircle, Clock, 
  Trash2, RefreshCw, AlertTriangle, ArrowRight, ShieldCheck, 
  FileText, Sparkles, ChevronDown, ChevronRight, X
} from 'lucide-react';
import { contractsService } from '../../../services/contractsService';
import ConfirmModal from '../Settings/ConfirmModal';
import toast from 'react-hot-toast';
import './ContractAmendmentsPanel.css';

export default function ContractAmendmentsPanel({ contractId, contractStatus, onAmendmentImplemented }) {
  const [amendments, setAmendments] = useState([]);
  const [isLoading, setIsLoading] = useState(false);
  const [expandedId, setExpandedId] = useState(null);
  const [isProcessingAction, setIsProcessingAction] = useState(false);
  const [confirmAction, setConfirmAction] = useState({
    isOpen: false,
    title: '',
    message: '',
    confirmText: '',
    confirmStyle: 'primary',
    action: null
  });

  // New Amendment Modal State
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [createForm, setCreateForm] = useState({
    title: '',
    amendment_type: 'COMMERCIAL_TERMS',
    description: '',
    change_summary: '',
    proposed_effective_date: '',
    changes: [
      { field_name: 'contract_value', previous_value: '', proposed_value: '', change_type: 'MODIFY' }
    ]
  });
  const [isSubmitting, setIsSubmitting] = useState(false);

  const fetchAmendments = useCallback(async () => {
    if (!contractId) return;
    setIsLoading(true);
    try {
      const res = await contractsService.getContractAmendments(contractId);
      const list = res.data?.amendments || res.data || [];
      setAmendments(list);
      if (list.length > 0 && !expandedId) {
        setExpandedId(list[0].id);
      }
    } catch (err) {
      console.error('Failed to load amendments', err);
      toast.error('Failed to load amendments history');
    } finally {
      setIsLoading(false);
    }
  }, [contractId, expandedId]);

  useEffect(() => {
    fetchAmendments();
  }, [fetchAmendments]);

  const handleSubmitForReview = (amendment) => {
    setConfirmAction({
      isOpen: true,
      title: 'Submit Amendment for Review',
      message: `Are you sure you want to submit amendment "${amendment.amendment_reference}" for approval review?`,
      confirmText: 'Submit for Review',
      confirmStyle: 'primary',
      action: async () => {
        await contractsService.submitContractAmendment(contractId, amendment.id, {});
        toast.success(`Amendment ${amendment.amendment_reference} submitted for approval`);
        fetchAmendments();
      }
    });
  };

  const handleImplement = (amendment) => {
    setConfirmAction({
      isOpen: true,
      title: 'Implement Approved Amendment',
      message: `Implement approved amendment "${amendment.amendment_reference}"? This will generate a new EFFECTIVE version snapshot.`,
      confirmText: 'Implement Amendment',
      confirmStyle: 'primary',
      action: async () => {
        await contractsService.implementContractAmendment(contractId, amendment.id);
        toast.success('Amendment implemented into new EFFECTIVE version!');
        fetchAmendments();
        if (onAmendmentImplemented) onAmendmentImplemented();
      }
    });
  };

  const handleCancel = (amendment) => {
    setConfirmAction({
      isOpen: true,
      title: 'Cancel Amendment Draft',
      message: `Are you sure you want to cancel amendment "${amendment.amendment_reference}"?`,
      confirmText: 'Cancel Amendment',
      confirmStyle: 'danger',
      action: async () => {
        await contractsService.cancelContractAmendment(contractId, amendment.id);
        toast.success(`Amendment cancelled`);
        fetchAmendments();
      }
    });
  };

  const handleExecuteConfirmedAction = async () => {
    if (!confirmAction.action) return;
    setIsProcessingAction(true);
    try {
      await confirmAction.action();
      setConfirmAction(prev => ({ ...prev, isOpen: false }));
    } catch (err) {
      console.error('Action error', err);
      toast.error(err.response?.data?.error?.message || 'Action failed');
    } finally {
      setIsProcessingAction(false);
    }
  };

  const handleAddChangeRow = () => {
    setCreateForm({
      ...createForm,
      changes: [
        ...createForm.changes,
        { field_name: 'contract_value', previous_value: '', proposed_value: '', change_type: 'MODIFY' }
      ]
    });
  };

  const handleRemoveChangeRow = (idx) => {
    const updated = createForm.changes.filter((_, i) => i !== idx);
    setCreateForm({ ...createForm, changes: updated });
  };

  const handleCreateAmendment = async (e) => {
    e.preventDefault();
    if (!createForm.title) {
      toast.error('Please enter an amendment title');
      return;
    }

    setIsSubmitting(true);
    try {
      await contractsService.createContractAmendment(contractId, createForm);
      toast.success('Amendment draft created successfully');
      setShowCreateModal(false);
      setCreateForm({
        title: '',
        amendment_type: 'COMMERCIAL_TERMS',
        description: '',
        change_summary: '',
        proposed_effective_date: '',
        changes: [{ field_name: 'contract_value', previous_value: '', proposed_value: '', change_type: 'MODIFY' }]
      });
      fetchAmendments();
    } catch (err) {
      console.error('Create amendment error', err);
      toast.error('Failed to create amendment draft');
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className="amendments-panel-container">
      {/* Top Action Header */}
      <div className="amendments-top-bar">
        <div className="amendments-title-group">
          <div className="amd-icon-badge">
            <FileEdit size={16} />
          </div>
          <div>
            <h4>Commercial Amendments & Change Governance</h4>
            <p>Amendments govern structured commercial adjustments without mutating historical agreement baseline.</p>
          </div>
        </div>

        <div className="amendments-actions-group">
          <button 
            className="amd-refresh-btn" 
            onClick={fetchAmendments} 
            title="Refresh Amendments"
            disabled={isLoading}
          >
            <RefreshCw size={14} className={isLoading ? 'spin-icon' : ''} />
          </button>
          <button 
            className="amd-create-btn"
            onClick={() => setShowCreateModal(true)}
          >
            <Plus size={14} /> Propose Amendment
          </button>
        </div>
      </div>

      {/* Amendments List */}
      {isLoading ? (
        <div className="amd-loading">
          <RefreshCw size={24} className="spin-icon text-blue" />
          <span>Loading agreement amendments...</span>
        </div>
      ) : amendments.length === 0 ? (
        <div className="amd-empty-state">
          <FileEdit size={36} className="text-muted" />
          <h5>No Amendments Recorded</h5>
          <p>This contract has no active or historical amendments. Click "Propose Amendment" to draft a rate, scope, or clause change.</p>
        </div>
      ) : (
        <div className="amd-list-wrapper">
          {amendments.map((amd) => {
            const isExpanded = expandedId === amd.id;
            const isDraft = amd.status === 'DRAFT';
            const isSubmitted = amd.status === 'SUBMITTED';
            const isApproved = amd.status === 'APPROVED';
            const isImplemented = amd.status === 'IMPLEMENTED';
            const isRejected = amd.status === 'REJECTED';
            const isCancelled = amd.status === 'CANCELLED';

            return (
              <div 
                key={amd.id} 
                className={`amd-card ${amd.status.toLowerCase()}`}
              >
                {/* Header Row */}
                <div 
                  className="amd-card-header"
                  onClick={() => setExpandedId(isExpanded ? null : amd.id)}
                >
                  <div className="amd-header-left">
                    <button className="amd-expand-toggle">
                      {isExpanded ? <ChevronDown size={15} /> : <ChevronRight size={15} />}
                    </button>
                    <span className="amd-ref">{amd.amendment_reference}</span>
                    <h5 className="amd-title">{amd.title}</h5>
                    <span className={`amd-status-badge ${amd.status.toLowerCase()}`}>
                      {amd.status}
                    </span>
                    <span className="amd-type-tag">
                      {amd.amendment_type?.replace(/_/g, ' ')}
                    </span>
                  </div>

                  <div className="amd-header-right" onClick={e => e.stopPropagation()}>
                    {isDraft && (
                      <button 
                        className="amd-btn-submit"
                        onClick={() => handleSubmitForReview(amd)}
                      >
                        <Send size={12} /> Submit Review
                      </button>
                    )}

                    {isApproved && (
                      <button 
                        className="amd-btn-implement"
                        onClick={() => handleImplement(amd)}
                      >
                        <Sparkles size={12} /> Implement to Version
                      </button>
                    )}

                    {(isDraft || isSubmitted) && (
                      <button 
                        className="amd-btn-cancel"
                        onClick={() => handleCancel(amd)}
                        title="Cancel Amendment"
                      >
                        <XCircle size={13} />
                      </button>
                    )}
                  </div>
                </div>

                {/* Expanded Details */}
                {isExpanded && (
                  <div className="amd-card-details">
                    {amd.description && (
                      <div className="amd-desc-box">
                        <strong>Description:</strong> {amd.description}
                      </div>
                    )}

                    {/* Structured Field Changes */}
                    {amd.changes && amd.changes.length > 0 && (
                      <div className="amd-changes-section">
                        <h6>Structured Field Adjustments</h6>
                        <div className="amd-changes-table">
                          <div className="table-hdr">
                            <span>Field Name</span>
                            <span>Current Baseline Value</span>
                            <span>Proposed Amendment Value</span>
                            <span>Type</span>
                          </div>
                          {amd.changes.map((c, i) => (
                            <div key={i} className="table-row">
                              <span className="col-fld"><strong>{c.field_name}</strong></span>
                              <span className="col-prev">{c.previous_value || '— None —'}</span>
                              <span className="col-prop highlight">{c.proposed_value || '— Removed —'}</span>
                              <span className={`col-chg-type ${c.change_type.toLowerCase()}`}>{c.change_type}</span>
                            </div>
                          ))}
                        </div>
                      </div>
                    )}

                    {/* Footer Metadata */}
                    <div className="amd-card-footer">
                      {amd.proposed_effective_date && (
                        <span><strong>Target Effective:</strong> {amd.proposed_effective_date}</span>
                      )}
                      {amd.created_at && (
                        <span><strong>Drafted:</strong> {new Date(amd.created_at).toLocaleDateString()}</span>
                      )}
                      {amd.submitted_at && (
                        <span><strong>Submitted:</strong> {new Date(amd.submitted_at).toLocaleDateString()}</span>
                      )}
                      {amd.approved_at && (
                        <span className="text-emerald"><strong>Approved:</strong> {new Date(amd.approved_at).toLocaleDateString()}</span>
                      )}
                    </div>
                  </div>
                )}
              </div>
            );
          })}
        </div>
      )}

      {/* Create Amendment Modal */}
      {showCreateModal && (
        <div className="amd-modal-overlay" onClick={() => setShowCreateModal(false)}>
          <div className="amd-modal-card" onClick={e => e.stopPropagation()}>
            <div className="amd-modal-header">
              <div className="modal-title-left">
                <div className="modal-icon-badge">
                  <FileEdit size={16} />
                </div>
                <div>
                  <h5>Propose Commercial Amendment</h5>
                  <p className="modal-subtitle">Draft a formal rate revision, clause update, or validity change</p>
                </div>
              </div>
              <button className="amd-close-modal" onClick={() => setShowCreateModal(false)} title="Close">
                <X size={18} />
              </button>
            </div>

            <form onSubmit={handleCreateAmendment} className="amd-modal-form">
              <div className="amd-modal-body">
                <div className="form-group">
                  <label className="form-label">
                    Amendment Title <span className="req">*</span>
                  </label>
                  <input 
                    type="text" 
                    className="form-control"
                    required
                    placeholder="e.g. Q3 Rate Adjustment & Fuel Surcharge Cap"
                    value={createForm.title}
                    onChange={e => setCreateForm({...createForm, title: e.target.value})}
                  />
                </div>

                <div className="form-row-2col">
                  <div className="form-group">
                    <label className="form-label">Amendment Type</label>
                    <div className="select-wrapper">
                      <select
                        className="form-control select-control"
                        value={createForm.amendment_type}
                        onChange={e => setCreateForm({...createForm, amendment_type: e.target.value})}
                      >
                        <option value="COMMERCIAL_TERMS">Commercial Terms</option>
                        <option value="RATE_REVISION">Rate Revision</option>
                        <option value="SCOPE_EXTENSION">Scope Extension</option>
                        <option value="CLAUSE_UPDATE">Clause Update</option>
                        <option value="VALIDITY_EXTENSION">Validity Extension</option>
                      </select>
                      <ChevronDown size={15} className="select-chevron" />
                    </div>
                  </div>

                  <div className="form-group">
                    <label className="form-label">Target Effective Date</label>
                    <input 
                      type="date" 
                      className="form-control date-control"
                      value={createForm.proposed_effective_date}
                      onChange={e => setCreateForm({...createForm, proposed_effective_date: e.target.value})}
                    />
                  </div>
                </div>

                <div className="form-group">
                  <label className="form-label">Commercial Rationale & Description</label>
                  <textarea 
                    className="form-control textarea-control"
                    rows={3}
                    placeholder="Provide detailed commercial context, carrier justification, or operational background for this amendment..."
                    value={createForm.description}
                    onChange={e => setCreateForm({...createForm, description: e.target.value})}
                  />
                </div>

                {/* Structured Changes Builder */}
                <div className="changes-builder">
                  <div className="builder-header">
                    <div>
                      <h6 className="builder-title">Structured Field Changes</h6>
                      <span className="builder-sub">Specify exact clause/value modifications</span>
                    </div>
                    <button type="button" className="btn-add-row" onClick={handleAddChangeRow}>
                      <Plus size={13} /> Add Field
                    </button>
                  </div>

                  <div className="builder-table-container">
                    <div className="builder-table-header">
                      <span>Target Field</span>
                      <span>Previous / Current Value</span>
                      <span>Proposed New Value</span>
                      <span></span>
                    </div>

                    <div className="builder-rows-list">
                      {createForm.changes.map((row, idx) => (
                        <div key={idx} className="builder-row">
                          <div className="select-wrapper">
                            <select
                              className="form-control select-control table-select"
                              value={row.field_name}
                              onChange={e => {
                                const updated = [...createForm.changes];
                                updated[idx].field_name = e.target.value;
                                setCreateForm({...createForm, changes: updated});
                              }}
                            >
                              <option value="contract_value">Contract Value</option>
                              <option value="currency">Currency</option>
                              <option value="effective_date">Effective Date</option>
                              <option value="expiry_date">Expiry Date</option>
                              <option value="transport_mode">Transport Mode</option>
                              <option value="payment_terms">Payment Terms</option>
                              <option value="free_days">Free Time (Detention)</option>
                              <option value="description">Description / Scope</option>
                            </select>
                            <ChevronDown size={13} className="select-chevron" />
                          </div>

                          <input 
                            type="text" 
                            className="form-control table-input"
                            placeholder="Previous / Current Value"
                            value={row.previous_value}
                            onChange={e => {
                              const updated = [...createForm.changes];
                              updated[idx].previous_value = e.target.value;
                              setCreateForm({...createForm, changes: updated});
                            }}
                          />

                          <input 
                            type="text" 
                            className="form-control table-input proposed-input"
                            placeholder="Proposed New Value"
                            value={row.proposed_value}
                            onChange={e => {
                              const updated = [...createForm.changes];
                              updated[idx].proposed_value = e.target.value;
                              setCreateForm({...createForm, changes: updated});
                            }}
                          />

                          <div className="del-btn-cell">
                            {createForm.changes.length > 1 ? (
                              <button 
                                type="button" 
                                className="btn-del-row"
                                onClick={() => handleRemoveChangeRow(idx)}
                                title="Remove change row"
                              >
                                <Trash2 size={14} />
                              </button>
                            ) : (
                              <span className="del-placeholder"></span>
                            )}
                          </div>
                        </div>
                      ))}
                    </div>
                  </div>
                </div>
              </div>

              <div className="amd-modal-footer">
                <button type="button" className="btn-cancel" onClick={() => setShowCreateModal(false)}>
                  Cancel
                </button>
                <button type="submit" className="btn-confirm" disabled={isSubmitting}>
                  {isSubmitting ? (
                    <>
                      <RefreshCw size={14} className="spin-icon" />
                      <span>Creating Draft...</span>
                    </>
                  ) : (
                    <>
                      <Sparkles size={14} />
                      <span>Create Amendment Draft</span>
                    </>
                  )}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* ── Amendment Action Confirmation Modal ── */}
      <ConfirmModal
        isOpen={confirmAction.isOpen}
        title={confirmAction.title}
        message={confirmAction.message}
        confirmText={confirmAction.confirmText}
        confirmStyle={confirmAction.confirmStyle}
        isLoading={isProcessingAction}
        onConfirm={handleExecuteConfirmedAction}
        onCancel={() => setConfirmAction(prev => ({ ...prev, isOpen: false }))}
      />
    </div>
  );
}

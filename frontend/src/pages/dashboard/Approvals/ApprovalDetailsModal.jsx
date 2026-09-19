import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  X, Check, XCircle, User, Calendar, ExternalLink,
  MessageSquare, History, ShieldAlert, CheckCircle2, Clock, Ban, Bot, AlertTriangle, Code,
  RotateCcw, RefreshCw, Send, Layers, HelpCircle, ArrowRight
} from 'lucide-react';
import { approvalsService } from '../../../services/approvalsService';

export default function ApprovalDetailsModal({
  item,
  currentUser,
  onClose,
  onApprove,
  onOpenRejectModal,
  onOpenReturnModal,
  onCancel,
  onRetryExecution,
}) {
  const navigate = useNavigate();
  const [notes, setNotes] = useState('');
  const [submitting, setSubmitting] = useState(false);
  const [cancelling, setCancelling] = useState(false);
  const [retrying, setRetrying] = useState(false);

  // Dynamic preview and history states
  const [preview, setPreview] = useState(null);
  const [history, setHistory] = useState([]);
  const [loadingPreview, setLoadingPreview] = useState(false);
  const [editableMessage, setEditableMessage] = useState('');

  useEffect(() => {
    if (item?.id) {
      loadPreviewAndHistory(item.id);
    }
  }, [item?.id]);

  const loadPreviewAndHistory = async (id) => {
    try {
      setLoadingPreview(true);
      const [previewRes, historyRes] = await Promise.allSettled([
        approvalsService.getActionPreview(id),
        approvalsService.getDecisionHistory(id),
      ]);

      if (previewRes.status === 'fulfilled' && previewRes.value) {
        setPreview(previewRes.value);
        if (previewRes.value.message_preview?.body) {
          setEditableMessage(previewRes.value.message_preview.body);
        }
      }
      if (historyRes.status === 'fulfilled' && Array.isArray(historyRes.value)) {
        setHistory(historyRes.value);
      }
    } catch (err) {
      console.warn('Could not load rich preview or history:', err);
    } finally {
      setLoadingPreview(false);
    }
  };

  if (!item) return null;

  // Separation of duties detection: is the current user the requester?
  const isRequester = (item.requesterName && currentUser && item.requesterName.trim().toLowerCase() === currentUser.trim().toLowerCase()) ||
    (item.requested_by_name && currentUser && item.requested_by_name.trim().toLowerCase() === currentUser.trim().toLowerCase());

  const handleApproveClick = async () => {
    if (submitting || cancelling) return;
    try {
      setSubmitting(true);
      await onApprove(item, notes);
      setSubmitting(false);
      onClose();
    } catch (err) {
      console.error(err);
      setSubmitting(false);
    }
  };

  const handleCancelClick = async () => {
    if (submitting || cancelling) return;
    try {
      setCancelling(true);
      if (onCancel) await onCancel(item, notes);
      setCancelling(false);
      onClose();
    } catch (err) {
      console.error(err);
      setCancelling(false);
    }
  };

  const handleRetryClick = async () => {
    if (retrying) return;
    try {
      setRetrying(true);
      if (onRetryExecution) {
        await onRetryExecution(item);
      } else {
        await approvalsService.retryExecution(item.id);
      }
      await loadPreviewAndHistory(item.id);
      setRetrying(false);
    } catch (err) {
      console.error('Retry failed:', err);
      setRetrying(false);
    }
  };

  const handleNavigateEntity = () => {
    const ref = (item.relatedRef || '').toLowerCase();
    const entityType = (item.relatedEntityType || '').toLowerCase();
    const id = item.relatedEntityId;

    if (ref.includes('rfq') || entityType.includes('rfq')) {
      navigate(id ? `/dashboard/rfq/${id}` : '/dashboard/rfq');
    } else if (ref.includes('contract') || entityType.includes('contract')) {
      navigate(id ? `/dashboard/contracts/${id}` : '/dashboard/contracts');
    } else if (ref.includes('shipment') || entityType.includes('shipment')) {
      navigate(id ? `/dashboard/shipments/${id}` : '/dashboard/shipments');
    } else if (ref.includes('customer') || entityType.includes('customer')) {
      navigate(id ? `/dashboard/customers/${id}` : '/dashboard/customers');
    } else if (ref.includes('invoice') || entityType.includes('invoice')) {
      navigate('/dashboard/invoices');
    } else if (ref.includes('document') || ref.includes('bill of lading') || entityType.includes('doc')) {
      navigate('/dashboard/documents');
    } else if (ref.includes('lead') || entityType.includes('lead')) {
      navigate(id ? `/dashboard/leads/${id}` : '/dashboard/leads');
    }
  };

  const getStatusBadgeClass = (status) => {
    switch (status) {
      case 'Approved':
      case 'APPROVED':
      case 'Completed':
      case 'COMPLETED': return 'status-badge approved';
      case 'Rejected':
      case 'REJECTED': return 'status-badge rejected';
      case 'Returned for Changes':
      case 'RETURNED_FOR_CHANGES': return 'status-badge in-progress';
      case 'Cancelled':
      case 'CANCELLED': return 'status-badge cancelled';
      case 'Expired':
      case 'EXPIRED':
      case 'Overdue': return 'status-badge overdue';
      case 'Executing':
      case 'EXECUTING': return 'status-badge in-progress';
      case 'Failed':
      case 'FAILED': return 'status-badge rejected';
      default: return 'status-badge pending';
    }
  };

  const getTypeBadgeClass = (type) => {
    switch (type) {
      case 'Document Approval': return 'type-badge document';
      case 'Commercial Approval': return 'type-badge commercial';
      case 'Clarification Email Approval': return 'type-badge commercial';
      case 'Operations Approval': return 'type-badge operations';
      case 'Finance Approval': return 'type-badge finance';
      default: return 'type-badge default';
    }
  };

  const isTerminal = ['Approved', 'APPROVED', 'Rejected', 'REJECTED', 'Cancelled', 'CANCELLED', 'Expired', 'EXPIRED', 'Completed', 'COMPLETED'].includes(item.status);

  return (
    <div className="approval-modal-overlay" onClick={onClose}>
      <div className="approval-modal-container details-container" onClick={(e) => e.stopPropagation()} style={{ maxWidth: 840 }}>
        {/* Header */}
        <div className="approval-modal-header">
          <div>
            <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 6 }}>
              <span className={getTypeBadgeClass(item.type)}>{item.type}</span>
              <span className={getStatusBadgeClass(item.status)}>{item.status}</span>
              {item.priority && (
                <span style={{ fontSize: '0.7rem', fontWeight: 800, padding: '2px 7px', borderRadius: 10, background: '#F1F5F9', color: '#475569' }}>
                  {item.priority} PRIORITY
                </span>
              )}
              {item.riskLevel && (
                <span style={{
                  fontSize: '0.7rem',
                  fontWeight: 800,
                  padding: '2px 7px',
                  borderRadius: 10,
                  background: item.riskLevel === 'HIGH_RISK' || item.riskLevel === 'CRITICAL' ? '#FEE2E2' : '#EFF6FF',
                  color: item.riskLevel === 'HIGH_RISK' || item.riskLevel === 'CRITICAL' ? '#B91C1C' : '#1D4ED8',
                }}>
                  {item.riskLevel}
                </span>
              )}
            </div>
            <h3 style={{ margin: 0, fontSize: '1.2rem', fontWeight: 800, color: '#0F172A' }}>{item.title}</h3>
            <div style={{ display: 'flex', alignItems: 'center', gap: 10, marginTop: 3 }}>
              <span style={{ fontSize: '0.76rem', color: '#64748B', fontWeight: 600 }}>ID: {item.id}</span>
              {item.actionName && (
                <span style={{ fontSize: '0.74rem', color: '#475569', fontFamily: 'monospace', background: '#F1F5F9', padding: '1px 6px', borderRadius: 4 }}>
                  Action: {item.actionName}
                </span>
              )}
            </div>
          </div>
          <button type="button" className="modal-close-btn" onClick={onClose}>
            <X size={18} />
          </button>
        </div>

        {/* Body */}
        <div className="details-body" style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
          {/* STALE ACTION WARNING BANNER */}
          {preview?.is_stale && (
            <div style={{ background: '#FFFBEB', border: '1px solid #FCD34D', borderRadius: 10, padding: '14px 16px', display: 'flex', gap: 12, alignItems: 'flex-start' }}>
              <AlertTriangle size={20} style={{ color: '#D97706', flexShrink: 0, marginTop: 2 }} />
              <div>
                <strong style={{ fontSize: '0.86rem', color: '#92400E', display: 'block', marginBottom: 2 }}>
                  Stale Action Warning - Re-Review Required
                </strong>
                <p style={{ margin: 0, fontSize: '0.82rem', color: '#B45309', lineHeight: 1.45 }}>
                  {preview.stale_reason || 'The underlying source record has changed materially since this approval was created. You cannot execute this action without returning for changes or requesting updated inputs.'}
                </p>
              </div>
            </div>
          )}

          {/* SEPARATION OF DUTIES WARNING (If logged in user created this high-risk item) */}
          {isRequester && (item.riskLevel === 'HIGH_RISK' || item.category === 'FINANCE' || item.category === 'COMMERCIAL') && !isTerminal && (
            <div style={{ background: '#F0F9FF', border: '1px solid #BAE6FD', borderRadius: 10, padding: '12px 16px', display: 'flex', gap: 10, alignItems: 'center' }}>
              <ShieldAlert size={18} style={{ color: '#0284C7', flexShrink: 0 }} />
              <span style={{ fontSize: '0.82rem', color: '#0369A1', lineHeight: 1.4 }}>
                <strong>Separation of Duties Policy:</strong> You requested this high-risk action. Another authorized manager or director must provide final sign-off.
              </span>
            </div>
          )}

          {/* Metadata Cards Grid */}
          <div className="details-meta-grid">
            {/* Related Reference with Navigation */}
            <div className="meta-card clickable-meta" onClick={handleNavigateEntity} style={{ cursor: 'pointer' }}>
              <span className="meta-label">
                <ExternalLink size={13} /> Related Context
              </span>
              <strong className="meta-val" style={{ color: '#2563EB', textDecoration: 'underline' }}>
                {item.relatedRef || 'View Business Record'}
              </strong>
              <small className="meta-sub">{item.customerName || 'Direct Entity Link'}</small>
            </div>

            {/* Requester */}
            <div className="meta-card">
              <span className="meta-label">
                <User size={13} /> Requester
              </span>
              <strong className="meta-val">{item.requesterName}</strong>
              <small className="meta-sub">{item.department || 'Operations'} Department</small>
            </div>

            {/* Due Date */}
            <div className="meta-card">
              <span className="meta-label">
                <Calendar size={13} /> Due Date
              </span>
              <strong className="meta-val">{item.dueDate}</strong>
              <small className={`meta-sub ${item.dueText === 'Overdue' ? 'overdue-text' : ''}`}>
                {item.dueText}
              </small>
            </div>
          </div>

          {/* Rejection / Decision Banner */}
          {(item.status === 'Rejected' || item.status === 'REJECTED') && (
            <div style={{ background: '#FEF2F2', border: '1px solid #FECACA', borderRadius: 10, padding: '14px 16px', display: 'flex', flexDirection: 'column', gap: 4 }}>
              <div style={{ display: 'flex', alignItems: 'center', gap: 6, color: '#DC2626', fontWeight: 800, fontSize: '0.84rem' }}>
                <ShieldAlert size={16} /> Rejection Details
              </div>
              <p style={{ margin: 0, fontSize: '0.84rem', color: '#991B1B', fontWeight: 650 }}>
                Reason: {item.rejectionReason || 'Declined by authorized approver'}
              </p>
            </div>
          )}

          {/* Returned for Changes Banner */}
          {(item.status === 'Returned for Changes' || item.status === 'RETURNED_FOR_CHANGES') && (
            <div style={{ background: '#FFFBEB', border: '1px solid #FDE68A', borderRadius: 10, padding: '14px 16px', display: 'flex', flexDirection: 'column', gap: 4 }}>
              <div style={{ display: 'flex', alignItems: 'center', gap: 6, color: '#D97706', fontWeight: 800, fontSize: '0.84rem' }}>
                <RotateCcw size={16} /> Returned for Changes
              </div>
              <p style={{ margin: 0, fontSize: '0.84rem', color: '#92400E', fontWeight: 650 }}>
                Requested Revision: {item.returnedReason || 'Please review comments and update action payload'}
              </p>
            </div>
          )}

          {/* Execution Failure Banner */}
          {(item.executionStatus === 'FAILED' || item.status === 'Failed') && (
            <div style={{ background: '#FEF2F2', border: '1px solid #FECACA', borderRadius: 10, padding: '14px 16px', display: 'flex', alignItems: 'center', justifyContent: 'space-between', gap: 12 }}>
              <div>
                <strong style={{ fontSize: '0.86rem', color: '#991B1B', display: 'block' }}>
                  Execution Failed
                </strong>
                <span style={{ fontSize: '0.80rem', color: '#B91C1C' }}>
                  {item.executionError || 'Action execution could not be completed by the Action System.'}
                </span>
              </div>
              <button
                type="button"
                className="btn-cancel"
                onClick={handleRetryClick}
                disabled={retrying}
                style={{ background: '#DC2626', color: '#FFFFFF', border: 'none', display: 'flex', alignItems: 'center', gap: 6, fontWeight: 700 }}
              >
                <RefreshCw size={14} className={retrying ? 'animate-spin' : ''} /> {retrying ? 'Retrying...' : 'Retry Execution'}
              </button>
            </div>
          )}

          {/* ACTION PREVIEW: What will happen, Affected Record, Current vs Proposed Value */}
          <div style={{ background: '#FFFFFF', border: '1px solid #E2E8F0', borderRadius: 10, padding: '16px', display: 'flex', flexDirection: 'column', gap: 12 }}>
            <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', borderBottom: '1px solid #F1F5F9', paddingBottom: 10 }}>
              <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                <Layers size={17} style={{ color: '#2563EB' }} />
                <h4 style={{ margin: 0, fontSize: '0.90rem', fontWeight: 800, color: '#0F172A' }}>
                  Action Preview & Target Impact
                </h4>
              </div>
              <div style={{ display: 'flex', gap: 6 }}>
                <span style={{
                  fontSize: '0.72rem',
                  fontWeight: 700,
                  padding: '2px 8px',
                  borderRadius: 6,
                  background: preview?.is_reversible ? '#F0FDF4' : '#FFF7ED',
                  color: preview?.is_reversible ? '#166534' : '#C2410C',
                }}>
                  {preview?.is_reversible ? '✓ Reversible' : '⚠ Non-Reversible'}
                </span>
                <span style={{
                  fontSize: '0.72rem',
                  fontWeight: 700,
                  padding: '2px 8px',
                  borderRadius: 6,
                  background: preview?.external_communication ? '#EFF6FF' : '#F8FAFC',
                  color: preview?.external_communication ? '#1E40AF' : '#475569',
                }}>
                  {preview?.external_communication ? '✉ External Communication' : '🔒 Internal Mutation'}
                </span>
              </div>
            </div>

            {/* Impact & Evidence Summary */}
            <div style={{ display: 'grid', gridTemplateColumns: 'repeat(2, 1fr)', gap: 12, fontSize: '0.82rem' }}>
              <div>
                <strong style={{ display: 'block', color: '#64748B', fontSize: '0.74rem', textTransform: 'uppercase', marginBottom: 2 }}>
                  Expected Business Impact
                </strong>
                <p style={{ margin: 0, color: '#1E293B', lineHeight: 1.45 }}>
                  {preview?.expected_impact || item.impactSummary || item.description || 'Target record will be updated upon authorization.'}
                </p>
              </div>
              <div>
                <strong style={{ display: 'block', color: '#64748B', fontSize: '0.74rem', textTransform: 'uppercase', marginBottom: 2 }}>
                  Evidence & Justification
                </strong>
                <p style={{ margin: 0, color: '#1E293B', lineHeight: 1.45 }}>
                  {preview?.evidence || item.evidence || 'Grounding derived from operational logs and active business rules.'}
                </p>
              </div>
            </div>

            {/* Current vs Proposed Values Comparison Table */}
            {preview && (
              <div style={{ marginTop: 4 }}>
                <strong style={{ display: 'block', color: '#64748B', fontSize: '0.74rem', textTransform: 'uppercase', marginBottom: 6 }}>
                  State Mutation Comparison (Current vs Proposed)
                </strong>
                <div style={{ display: 'grid', gridTemplateColumns: '1fr auto 1fr', gap: 10, alignItems: 'center' }}>
                  <div style={{ background: '#F8FAFC', border: '1px solid #E2E8F0', borderRadius: 8, padding: '10px 12px' }}>
                    <div style={{ fontSize: '0.72rem', fontWeight: 700, color: '#64748B', marginBottom: 4, textTransform: 'uppercase' }}>
                      Current State
                    </div>
                    <pre style={{ margin: 0, fontFamily: 'monospace', fontSize: '0.76rem', color: '#334155', whiteSpace: 'pre-wrap' }}>
                      {preview.current_value && Object.keys(preview.current_value).length > 0
                        ? JSON.stringify(preview.current_value, null, 2)
                        : '(Initial / Unset Record)'}
                    </pre>
                  </div>

                  <ArrowRight size={18} style={{ color: '#94A3B8' }} />

                  <div style={{ background: '#EFF6FF', border: '1px solid #BFDBFE', borderRadius: 8, padding: '10px 12px' }}>
                    <div style={{ fontSize: '0.72rem', fontWeight: 700, color: '#1D4ED8', marginBottom: 4, textTransform: 'uppercase' }}>
                      Proposed State (Post-Approval)
                    </div>
                    <pre style={{ margin: 0, fontFamily: 'monospace', fontSize: '0.76rem', color: '#1E3A8A', whiteSpace: 'pre-wrap' }}>
                      {preview.proposed_value && Object.keys(preview.proposed_value).length > 0
                        ? JSON.stringify(preview.proposed_value, null, 2)
                        : (item.proposedPayload || '{}')}
                    </pre>
                  </div>
                </div>
              </div>
            )}

            {/* Outbound Message Draft Preview if Applicable */}
            {preview?.message_preview && (
              <div style={{ marginTop: 6, background: '#F8FAFC', border: '1px solid #CBD5E1', borderRadius: 8, padding: '12px 14px', display: 'flex', flexDirection: 'column', gap: 8 }}>
                <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
                  <span style={{ fontSize: '0.80rem', fontWeight: 800, color: '#0F172A', display: 'flex', alignItems: 'center', gap: 6 }}>
                    <Send size={14} style={{ color: '#2563EB' }} /> Outbound Message Preview
                  </span>
                  <span style={{ fontSize: '0.72rem', background: '#DBEAFE', color: '#1E40AF', padding: '1px 6px', borderRadius: 4, fontWeight: 700 }}>
                    {preview.message_preview.channel || 'EMAIL'}
                  </span>
                </div>
                <div style={{ fontSize: '0.78rem', color: '#475569' }}>
                  <strong>To:</strong> {preview.message_preview.recipient || 'N/A'}{' '}
                  {preview.message_preview.subject && (
                    <>
                      • <strong>Subject:</strong> {preview.message_preview.subject}
                    </>
                  )}
                </div>
                <div>
                  <textarea
                    className="form-textarea"
                    rows={3}
                    value={editableMessage}
                    onChange={(e) => setEditableMessage(e.target.value)}
                    style={{ fontSize: '0.80rem', background: '#FFFFFF', color: '#1E293B' }}
                  />
                </div>
              </div>
            )}
          </div>

          {/* Immutable Decision History Trail */}
          <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
            <h4 style={{ margin: 0, fontSize: '0.84rem', fontWeight: 800, color: '#334155', display: 'flex', alignItems: 'center', gap: 6 }}>
              <History size={14} /> Immutable Decision History & Audit Trail
            </h4>
            <div className="approval-timeline">
              <div className="timeline-item">
                <div className="timeline-dot blue" />
                <div className="timeline-content">
                  <span className="timeline-title">Approval Request Initiated</span>
                  <span className="timeline-sub">By {item.requesterName} ({item.department})</span>
                  <span className="timeline-time">{item.createdAt ? new Date(item.createdAt).toLocaleString() : 'Recent'}</span>
                </div>
              </div>

              {history.map((dh) => (
                <div key={dh.id} className="timeline-item">
                  <div className={`timeline-dot ${dh.decision === 'APPROVE' ? 'green' : dh.decision === 'REJECT' ? 'red' : 'yellow'}`} />
                  <div className="timeline-content">
                    <span className="timeline-title">Decision: {dh.decision}</span>
                    <span className="timeline-sub">
                      Actor: {dh.actor_name} {dh.reason ? `• Reason: ${dh.reason}` : ''}
                    </span>
                    <span className="timeline-time">{new Date(dh.created_at).toLocaleString()}</span>
                  </div>
                </div>
              ))}

              {history.length === 0 && (
                <div className="timeline-item">
                  <div className="timeline-dot yellow" />
                  <div className="timeline-content">
                    <span className="timeline-title">Awaiting Authorized Decision</span>
                    <span className="timeline-sub">Pending review in {item.category || 'OPERATIONS'} queue</span>
                    <span className="timeline-time">Due: {item.dueDate}</span>
                  </div>
                </div>
              )}
            </div>
          </div>

          {/* Action Comments Input for Pending Items */}
          {!isTerminal && (
            <div className="details-notes-box">
              <label className="form-label" style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
                <MessageSquare size={14} /> Approval Sign-off Comments (Optional)
              </label>
              <textarea
                className="form-textarea"
                rows={2}
                placeholder="Add audit comments before approving or returning..."
                value={notes}
                onChange={(e) => setNotes(e.target.value)}
              />
            </div>
          )}
        </div>

        {/* Footer Actions */}
        <div className="approval-modal-footer" style={{ justifyContent: 'space-between' }}>
          <button type="button" className="btn-cancel" onClick={onClose} disabled={submitting || cancelling}>
            Close
          </button>

          {!isTerminal && (
            <div style={{ display: 'flex', gap: 10, alignItems: 'center' }}>
              <button
                type="button"
                className="btn-cancel"
                style={{ background: '#F1F5F9', color: '#475569', border: '1px solid #CBD5E1', display: 'flex', alignItems: 'center', gap: 6 }}
                onClick={handleCancelClick}
                disabled={submitting || cancelling}
              >
                <Ban size={15} /> {cancelling ? 'Cancelling...' : 'Cancel'}
              </button>

              <button
                type="button"
                className="btn-cancel"
                style={{ background: '#FFFBEB', color: '#92400E', border: '1px solid #FCD34D', display: 'flex', alignItems: 'center', gap: 6, fontWeight: 700 }}
                onClick={() => {
                  onClose();
                  if (onOpenReturnModal) onOpenReturnModal(item);
                }}
                disabled={submitting || cancelling}
              >
                <RotateCcw size={15} /> Return for Changes
              </button>

              <button
                type="button"
                className="btn-action-reject"
                onClick={() => {
                  onClose();
                  onOpenRejectModal(item);
                }}
                disabled={submitting || cancelling}
              >
                <XCircle size={15} /> Reject
              </button>

              <button
                type="button"
                className="btn-action-approve"
                onClick={handleApproveClick}
                disabled={submitting || cancelling || preview?.is_stale}
                title={preview?.is_stale ? 'Action is stale and cannot be approved until re-reviewed' : 'Approve and execute action'}
              >
                {submitting ? (
                  'Approving...'
                ) : (
                  <>
                    <Check size={15} /> Approve & Execute
                  </>
                )}
              </button>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}


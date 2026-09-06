import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  X, Check, XCircle, User, Calendar, ExternalLink,
  MessageSquare, History, ShieldAlert, CheckCircle2, Clock
} from 'lucide-react';

export default function ApprovalDetailsModal({ item, onClose, onApprove, onOpenRejectModal }) {
  const navigate = useNavigate();
  const [notes, setNotes] = useState('');
  const [submitting, setSubmitting] = useState(false);

  if (!item) return null;

  const handleApproveClick = async () => {
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

  const handleNavigateEntity = () => {
    const ref = (item.relatedRef || '').toLowerCase();
    if (ref.includes('shipment')) {
      navigate('/dashboard/shipments');
    } else if (ref.includes('customer')) {
      navigate('/dashboard/customers');
    } else if (ref.includes('invoice')) {
      navigate('/dashboard/invoices');
    } else if (ref.includes('document') || ref.includes('bill of lading')) {
      navigate('/dashboard/documents');
    }
  };

  const getStatusBadgeClass = (status) => {
    switch (status) {
      case 'Approved': return 'status-badge approved';
      case 'Rejected': return 'status-badge rejected';
      case 'Overdue': return 'status-badge overdue';
      default: return 'status-badge pending';
    }
  };

  const getTypeBadgeClass = (type) => {
    switch (type) {
      case 'Document Approval': return 'type-badge document';
      case 'Commercial Approval': return 'type-badge commercial';
      case 'Operations Approval': return 'type-badge operations';
      case 'Finance Approval': return 'type-badge finance';
      default: return 'type-badge default';
    }
  };

  return (
    <div className="approval-modal-overlay" onClick={onClose}>
      <div className="approval-modal-container details-container" onClick={(e) => e.stopPropagation()}>
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
            </div>
            <h3 style={{ margin: 0, fontSize: '1.15rem', fontWeight: 800, color: '#0F172A' }}>{item.title}</h3>
            <span style={{ fontSize: '0.76rem', color: '#64748B', fontWeight: 600 }}>ID: {item.id}</span>
          </div>
          <button type="button" className="modal-close-btn" onClick={onClose}>
            <X size={18} />
          </button>
        </div>

        {/* Body */}
        <div className="details-body">
          {/* Metadata Cards Grid */}
          <div className="details-meta-grid">
            {/* Related Reference with Navigation */}
            <div className="meta-card clickable-meta" onClick={handleNavigateEntity} style={{ cursor: 'pointer' }}>
              <span className="meta-label">
                <ExternalLink size={13} /> Related Context
              </span>
              <strong className="meta-val" style={{ color: '#2563EB', textDecoration: 'underline' }}>
                {item.relatedRef}
              </strong>
              <small className="meta-sub">{item.customerName}</small>
            </div>

            {/* Requester */}
            <div className="meta-card">
              <span className="meta-label">
                <User size={13} /> Requester
              </span>
              <strong className="meta-val">{item.requesterName}</strong>
              <small className="meta-sub">{item.department} Department</small>
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
          {item.status === 'Rejected' && (
            <div style={{ background: '#FEF2F2', border: '1px solid #FECACA', borderRadius: 10, padding: '14px 16px', display: 'flex', flexDirection: 'column', gap: 4 }}>
              <div style={{ display: 'flex', alignItems: 'center', gap: 6, color: '#DC2626', fontWeight: 800, fontSize: '0.84rem' }}>
                <ShieldAlert size={16} /> Rejection Details
              </div>
              <p style={{ margin: 0, fontSize: '0.84rem', color: '#991B1B', fontWeight: 650 }}>
                Reason: {item.rejectionReason || 'Margin threshold or compliance flag'}
              </p>
              {item.description && (
                <p style={{ margin: '4px 0 0 0', fontSize: '0.8rem', color: '#B91C1C', lineHeight: 1.4 }}>
                  Notes: {item.description}
                </p>
              )}
            </div>
          )}

          {item.status === 'Approved' && (
            <div style={{ background: '#F0FDF4', border: '1px solid #BBF7D0', borderRadius: 10, padding: '14px 16px', display: 'flex', alignItems: 'center', gap: 10 }}>
              <CheckCircle2 size={20} style={{ color: '#16A34A', flexShrink: 0 }} />
              <div>
                <strong style={{ display: 'block', fontSize: '0.84rem', color: '#14532D' }}>
                  Approved Request
                </strong>
                <span style={{ fontSize: '0.78rem', color: '#15803D' }}>
                  Sign-off completed and recorded in audit log.
                </span>
              </div>
            </div>
          )}

          {/* Description Block */}
          {item.description && item.status !== 'Rejected' && (
            <div className="details-desc-box">
              <h4 style={{ margin: '0 0 6px 0', fontSize: '0.84rem', fontWeight: 800, color: '#334155' }}>
                Request Context & Description
              </h4>
              <p style={{ margin: 0, fontSize: '0.84rem', color: '#475569', lineHeight: 1.5 }}>
                {item.description}
              </p>
            </div>
          )}

          {/* Activity / Audit History Timeline */}
          <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
            <h4 style={{ margin: 0, fontSize: '0.84rem', fontWeight: 800, color: '#334155', display: 'flex', alignItems: 'center', gap: 6 }}>
              <History size={14} /> Approval Audit Trail
            </h4>
            <div className="approval-timeline">
              <div className="timeline-item">
                <div className="timeline-dot blue" />
                <div className="timeline-content">
                  <span className="timeline-title">Approval Requested</span>
                  <span className="timeline-sub">By {item.requesterName} ({item.department})</span>
                  <span className="timeline-time">{item.createdAt ? new Date(item.createdAt).toLocaleString() : 'Aug 10, 2026 10:30 AM'}</span>
                </div>
              </div>

              <div className="timeline-item">
                <div className="timeline-dot yellow" />
                <div className="timeline-content">
                  <span className="timeline-title">Routed to Manager Queue</span>
                  <span className="timeline-sub">Pending review in {item.category || 'DOCUMENTS'} queue</span>
                  <span className="timeline-time">Due: {item.dueDate}</span>
                </div>
              </div>

              {item.status === 'Approved' && (
                <div className="timeline-item">
                  <div className="timeline-dot green" />
                  <div className="timeline-content">
                    <span className="timeline-title">Request Approved</span>
                    <span className="timeline-sub">Authorized manager sign-off verified</span>
                    <span className="timeline-time">Completed</span>
                  </div>
                </div>
              )}

              {item.status === 'Rejected' && (
                <div className="timeline-item">
                  <div className="timeline-dot red" />
                  <div className="timeline-content">
                    <span className="timeline-title">Request Rejected</span>
                    <span className="timeline-sub">Reason: {item.rejectionReason || 'Declined'}</span>
                    <span className="timeline-time">Completed</span>
                  </div>
                </div>
              )}
            </div>
          </div>

          {/* Action Comments Input for Pending Items */}
          {item.status !== 'Approved' && item.status !== 'Rejected' && (
            <div className="details-notes-box">
              <label className="form-label" style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
                <MessageSquare size={14} /> Approval Sign-off Comments (Optional)
              </label>
              <textarea
                className="form-textarea"
                rows={2}
                placeholder="Add audit comments before approving..."
                value={notes}
                onChange={(e) => setNotes(e.target.value)}
              />
            </div>
          )}
        </div>

        {/* Footer Actions */}
        <div className="approval-modal-footer" style={{ justifyContent: 'space-between' }}>
          <button type="button" className="btn-cancel" onClick={onClose}>
            Close
          </button>

          {item.status !== 'Approved' && item.status !== 'Rejected' && (
            <div style={{ display: 'flex', gap: 10 }}>
              <button
                type="button"
                className="btn-action-reject"
                onClick={() => {
                  onClose();
                  onOpenRejectModal(item);
                }}
              >
                <XCircle size={15} /> Reject Request
              </button>
              <button
                type="button"
                className="btn-action-approve"
                onClick={handleApproveClick}
                disabled={submitting}
              >
                {submitting ? (
                  'Approving...'
                ) : (
                  <>
                    <Check size={15} /> Approve Request
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

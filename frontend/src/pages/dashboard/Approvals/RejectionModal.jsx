import React, { useState } from 'react';
import { X, XCircle, AlertTriangle } from 'lucide-react';

export default function RejectionModal({ isOpen, item, onClose, onConfirmReject }) {
  const [reasonCategory, setReasonCategory] = useState('Margin Below Minimum Threshold');
  const [customNotes, setCustomNotes] = useState('');
  const [error, setError] = useState('');
  const [submitting, setSubmitting] = useState(false);

  if (!isOpen || !item) return null;

  const REASON_OPTIONS = [
    'Margin Below Minimum Threshold',
    'Discrepancy in Invoice/Document Data',
    'Unverified Carrier Release / Compliance Flag',
    'Demurrage Waiver Request Declined',
    'Customer Credit Limit Exceeded',
    'Other / Custom Reason',
  ];

  const handleSubmit = async (e) => {
    e.preventDefault();
    const finalReason = reasonCategory === 'Other / Custom Reason' ? customNotes.trim() : reasonCategory;
    
    if (!finalReason) {
      setError('Please select or provide a valid rejection reason.');
      return;
    }

    try {
      setSubmitting(true);
      setError('');
      await onConfirmReject(item, finalReason, customNotes);
      setSubmitting(false);
      onClose();
    } catch (err) {
      console.error(err);
      setError(err?.message || 'Failed to submit rejection.');
      setSubmitting(false);
    }
  };

  return (
    <div className="approval-modal-overlay" onClick={onClose}>
      <div className="approval-modal-container" onClick={(e) => e.stopPropagation()} style={{ maxWidth: 520 }}>
        {/* Header */}
        <div className="approval-modal-header" style={{ background: '#FEF2F2', borderBottom: '1px solid #FECACA' }}>
          <div>
            <div style={{ display: 'flex', alignItems: 'center', gap: 6, color: '#DC2626', marginBottom: 2 }}>
              <AlertTriangle size={16} />
              <strong style={{ fontSize: '0.82rem', uppercase: true }}>Reject Approval Request</strong>
            </div>
            <h3 style={{ margin: 0, fontSize: '1.05rem', fontWeight: 800, color: '#991B1B' }}>
              {item.title}
            </h3>
            <span style={{ fontSize: '0.74rem', color: '#B91C1C', fontWeight: 600 }}>ID: {item.id}</span>
          </div>
          <button type="button" className="modal-close-btn" onClick={onClose}>
            <X size={18} />
          </button>
        </div>

        {error && (
          <div className="approval-modal-error" style={{ margin: '16px 24px 0 24px' }}>
            <AlertTriangle size={16} />
            <span>{error}</span>
          </div>
        )}

        {/* Body */}
        <form onSubmit={handleSubmit} className="approval-modal-form">
          <p style={{ margin: '0 0 8px 0', fontSize: '0.84rem', color: '#475569', lineHeight: 1.5 }}>
            Rejecting this request will set its status to <strong>Rejected</strong> and notify the requester. Please select a rejection reason:
          </p>

          <div className="form-group">
            <label className="form-label">Primary Rejection Reason *</label>
            <select
              className="form-select"
              value={reasonCategory}
              onChange={(e) => setReasonCategory(e.target.value)}
            >
              {REASON_OPTIONS.map((opt) => (
                <option key={opt} value={opt}>
                  {opt}
                </option>
              ))}
            </select>
          </div>

          <div className="form-group">
            <label className="form-label">Detailed Audit Notes / Explanation *</label>
            <textarea
              className="form-textarea"
              rows={3}
              placeholder="Provide specific feedback or reasons for rejecting this approval..."
              value={customNotes}
              onChange={(e) => setCustomNotes(e.target.value)}
              required
            />
          </div>

          {/* Footer */}
          <div className="approval-modal-footer">
            <button type="button" className="btn-cancel" onClick={onClose} disabled={submitting}>
              Cancel
            </button>
            <button
              type="submit"
              className="btn-action-reject"
              disabled={submitting}
              style={{ padding: '8px 18px' }}
            >
              {submitting ? (
                'Rejecting...'
              ) : (
                <>
                  <XCircle size={15} /> Confirm Rejection
                </>
              )}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}

import React, { useState } from 'react';
import { X, RotateCcw, AlertCircle } from 'lucide-react';

export default function ReturnModal({ isOpen, item, onClose, onConfirmReturn }) {
  const [reasonCategory, setReasonCategory] = useState('Missing or Incomplete Documentation');
  const [customNotes, setCustomNotes] = useState('');
  const [error, setError] = useState('');
  const [submitting, setSubmitting] = useState(false);

  if (!isOpen || !item) return null;

  const REASON_OPTIONS = [
    'Missing or Incomplete Documentation',
    'Financial / Pricing Discrepancy',
    'Indemnification / Legal Clause Revision Required',
    'Route or Milestone Adjustment Required',
    'Need Customer Written Confirmation',
    'Compliance Verification Needed',
    'Other / Custom Reason',
  ];

  const handleSubmit = async (e) => {
    e.preventDefault();
    const finalReason = reasonCategory === 'Other / Custom Reason' ? customNotes.trim() : reasonCategory;

    if (!finalReason) {
      setError('Please select or provide a valid reason for returning this request.');
      return;
    }

    try {
      setSubmitting(true);
      setError('');
      await onConfirmReturn(item, finalReason, customNotes);
      setSubmitting(false);
      onClose();
    } catch (err) {
      console.error(err);
      setError(err?.message || 'Failed to return request for changes.');
      setSubmitting(false);
    }
  };

  return (
    <div className="approval-modal-overlay" onClick={onClose}>
      <div className="approval-modal-container" onClick={(e) => e.stopPropagation()} style={{ maxWidth: 520 }}>
        {/* Header */}
        <div className="approval-modal-header" style={{ background: '#FFFBEB', borderBottom: '1px solid #FDE68A' }}>
          <div>
            <div style={{ display: 'flex', alignItems: 'center', gap: 6, color: '#D97706', marginBottom: 2 }}>
              <RotateCcw size={16} />
              <strong style={{ fontSize: '0.82rem', textTransform: 'uppercase' }}>Return for Changes</strong>
            </div>
            <h3 style={{ margin: 0, fontSize: '1.05rem', fontWeight: 800, color: '#92400E' }}>
              {item.title}
            </h3>
            <span style={{ fontSize: '0.74rem', color: '#B45309', fontWeight: 600 }}>ID: {item.id}</span>
          </div>
          <button type="button" className="modal-close-btn" onClick={onClose}>
            <X size={18} />
          </button>
        </div>

        {error && (
          <div className="approval-modal-error" style={{ margin: '16px 24px 0 24px' }}>
            <AlertCircle size={16} />
            <span>{error}</span>
          </div>
        )}

        {/* Body */}
        <form onSubmit={handleSubmit} className="approval-modal-form">
          <p style={{ margin: '0 0 8px 0', fontSize: '0.84rem', color: '#475569', lineHeight: 1.5 }}>
            Returning this action will set its status to <strong>Returned for Changes</strong> and send actionable feedback to the requester to update and re-submit:
          </p>

          <div className="form-group">
            <label className="form-label">Primary Return Reason *</label>
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
            <label className="form-label">Specific Changes or Clarifications Requested *</label>
            <textarea
              className="form-textarea"
              rows={3}
              placeholder="Describe exactly what needs to be amended or provided before this can be approved..."
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
              className="btn-action-return"
              disabled={submitting}
              style={{
                padding: '8px 18px',
                background: '#D97706',
                color: '#FFFFFF',
                border: 'none',
                borderRadius: 6,
                fontWeight: 700,
                cursor: 'pointer',
                display: 'flex',
                alignItems: 'center',
                gap: 6,
              }}
            >
              {submitting ? (
                'Returning...'
              ) : (
                <>
                  <RotateCcw size={15} /> Confirm Return
                </>
              )}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}

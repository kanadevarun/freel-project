import { useEffect } from 'react';
import './ConfirmModal.css';

export default function ConfirmModal({ isOpen, title, message, onConfirm, onCancel, confirmText = 'Confirm', confirmStyle = 'danger', isLoading = false, previewData = null }) {
  // Close on Escape key
  useEffect(() => {
    const handleEsc = (e) => {
      if (e.key === 'Escape' && isOpen && !isLoading) {
        onCancel();
      }
    };
    window.addEventListener('keydown', handleEsc);
    return () => window.removeEventListener('keydown', handleEsc);
  }, [isOpen, isLoading, onCancel]);

  if (!isOpen) return null;

  const handleOverlayClick = (e) => {
    if (e.target.classList.contains('confirm-modal-overlay') && !isLoading) {
      onCancel();
    }
  };

  return (
    <div className="confirm-modal-overlay" onClick={handleOverlayClick}>
      <div className="confirm-modal-container">
        
        {isLoading && (
          <div className="confirm-modal-loading-overlay">
            <div className="confirm-spinner"></div>
          </div>
        )}

        <div className={`confirm-modal-body ${isLoading ? 'confirm-modal-body-submitting' : ''}`}>
          
          <div className={`confirm-icon-circle ${confirmStyle}`}>
            {confirmStyle === 'danger' ? (
              <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                <path d="M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
            ) : (
              <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                <circle cx="12" cy="12" r="10" />
                <line x1="12" y1="8" x2="12" y2="12" />
                <line x1="12" y1="16" x2="12.01" y2="16" />
              </svg>
            )}
          </div>

          <h3 className="confirm-modal-title">{title}</h3>
          <p className="confirm-modal-message">{message}</p>

          {previewData && (
            <div className="confirm-modal-preview">
              <div className="preview-row">
                <span>Current Plan:</span>
                <strong>{previewData.current_plan_name}</strong>
              </div>
              <div className="preview-row">
                <span>New Plan:</span>
                <strong>{previewData.new_plan_name} ({previewData.billing_cycle})</strong>
              </div>
              <div className="preview-row">
                <span>New Price:</span>
                <strong>${previewData.new_price}</strong>
              </div>
              <div className="preview-row">
                <span>Effective Date:</span>
                <strong>{previewData.effective_date}</strong>
              </div>
              <div className="preview-row proration">
                <span>Prorated {previewData.prorated_charge < 0 ? 'Credit' : 'Charge'}:</span>
                <strong>${Math.abs(previewData.prorated_charge).toFixed(2)}</strong>
              </div>
            </div>
          )}

          <div className="confirm-modal-actions">
            <button 
              className="confirm-btn-cancel" 
              onClick={onCancel} 
              disabled={isLoading}
            >
              Cancel
            </button>
            <button 
              className={`confirm-btn-submit ${confirmStyle}`} 
              onClick={onConfirm} 
              disabled={isLoading}
            >
              {isLoading ? 'Processing...' : confirmText}
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}

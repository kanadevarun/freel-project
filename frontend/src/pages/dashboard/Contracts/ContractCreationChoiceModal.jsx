import React from 'react';
import { 
  FileEdit, UploadCloud, Sparkles, X, ChevronRight, 
  ArrowRight 
} from 'lucide-react';
import './ContractCreationChoiceModal.css';

export default function ContractCreationChoiceModal({ onClose, onSelectManual, onSelectImport }) {
  return (
    <div className="choice-modal-overlay" onClick={onClose}>
      <div className="choice-modal-card" onClick={e => e.stopPropagation()}>
        <div className="choice-modal-header">
          <div className="choice-header-left">
            <div className="choice-icon-badge">
              <Sparkles size={18} />
            </div>
            <div>
              <h4>Create Commercial Contract</h4>
              <p className="choice-subtitle">Select how you want to register this agreement in LogisticsHQ</p>
            </div>
          </div>
          <button className="choice-close-btn" onClick={onClose} title="Close">
            <X size={18} />
          </button>
        </div>

        <div className="choice-modal-body">
          {/* Option A: AI Document Import (Recommended) */}
          <div 
            className="choice-option-card recommended"
            onClick={onSelectImport}
          >
            <div className="option-badge-tag">
              <Sparkles size={11} />
              <span>Recommended (Agentic AI)</span>
            </div>
            <div className="option-card-content">
              <div className="option-icon-box import-icon">
                <UploadCloud size={24} />
              </div>
              <div className="option-text-group">
                <h5>Import Agreement Document</h5>
                <p>
                  Upload an executed PDF or DOCX contract. Our AI Sidecar extracts counterparty names,
                  rates, validity windows, payment terms, and detention free time into an editable draft for your review.
                </p>
                <div className="option-feature-pills">
                  <span>PDF & Word Support</span>
                  <span>Automatic Clause Extraction</span>
                  <span>Safe Draft Review</span>
                </div>
              </div>
              <div className="option-arrow-box">
                <ArrowRight size={18} />
              </div>
            </div>
          </div>

          {/* Option B: Manual Entry */}
          <div 
            className="choice-option-card manual"
            onClick={onSelectManual}
          >
            <div className="option-card-content">
              <div className="option-icon-box manual-icon">
                <FileEdit size={22} />
              </div>
              <div className="option-text-group">
                <h5>Create Manually</h5>
                <p>
                  Standard entry form to manually define contract reference, counterparty, transport mode,
                  validity dates, and commercial notes step-by-step.
                </p>
                <div className="option-feature-pills">
                  <span>Custom References</span>
                  <span>Instant Draft Creation</span>
                </div>
              </div>
              <div className="option-arrow-box">
                <ChevronRight size={18} />
              </div>
            </div>
          </div>
        </div>

        <div className="choice-modal-footer">
          <button type="button" className="btn-choice-cancel" onClick={onClose}>
            Cancel
          </button>
        </div>
      </div>
    </div>
  );
}

import React, { useState, useRef } from 'react';
import { 
  UploadCloud, FileText, Sparkles, X, AlertCircle, 
  CheckCircle2, Loader2, ArrowRight, FileCheck
} from 'lucide-react';
import toast from 'react-hot-toast';
import { contractsService } from '../../../services/contractsService';
import './ContractImportModal.css';

const CONTRACT_TYPE_HINTS = [
  { value: 'CARRIER_AGREEMENT', label: 'Carrier Rate & Service Agreement', sub: 'Ocean / Air liner contracts & allocation commitments' },
  { value: 'CUSTOMER_SLA', label: 'Customer Master SLA / Agreement', sub: 'Client pricing, guaranteed transit times, and payment terms' },
  { value: 'VENDOR_CONTRACT', label: 'Vendor Service Agreement', sub: 'Drayage, warehouse storage, customs brokerage' },
  { value: 'FORWARDER_PARTNERSHIP', label: 'Forwarder Partnership Agreement', sub: 'Inter-forwarder agency and co-loading agreement' }
];

export default function ContractImportModal({ onClose, onExtractionComplete }) {
  const [file, setFile] = useState(null);
  const [contractTypeHint, setContractTypeHint] = useState('CARRIER_AGREEMENT');
  const [isProcessing, setIsProcessing] = useState(false);
  const [currentStep, setCurrentStep] = useState(1);
  const [isDragOver, setIsDragOver] = useState(false);

  const fileInputRef = useRef(null);

  const handleFileDrop = (e) => {
    e.preventDefault();
    setIsDragOver(false);
    if (e.dataTransfer.files && e.dataTransfer.files[0]) {
      validateAndSetFile(e.dataTransfer.files[0]);
    }
  };

  const handleFileSelect = (e) => {
    if (e.target.files && e.target.files[0]) {
      validateAndSetFile(e.target.files[0]);
    }
  };

  const validateAndSetFile = (selectedFile) => {
    const validExtensions = ['.pdf', '.docx', '.doc', '.xlsx', '.xls'];
    const hasValidExt = validExtensions.some(ext => selectedFile.name.toLowerCase().endsWith(ext));
    if (!hasValidExt) {
      toast.error('Please upload a valid PDF, Word document (.docx), or Excel sheet.');
      return;
    }
    if (selectedFile.size > 25 * 1024 * 1024) {
      toast.error('File size exceeds 25MB limit.');
      return;
    }
    setFile(selectedFile);
  };

  const handleStartExtraction = async () => {
    if (!file) {
      toast.error('Please select a contract agreement document to upload.');
      return;
    }

    setIsProcessing(true);
    setCurrentStep(1);

    // Step simulation timers for user engagement
    const t1 = setTimeout(() => setCurrentStep(2), 700);
    const t2 = setTimeout(() => setCurrentStep(3), 1600);

    try {
      const res = await contractsService.importContractDocument(file, contractTypeHint);
      clearTimeout(t1);
      clearTimeout(t2);
      setCurrentStep(4);
      
      setTimeout(() => {
        toast.success('Contract document parsed successfully! Please review the extracted draft.');
        onExtractionComplete(res);
      }, 500);
    } catch (err) {
      clearTimeout(t1);
      clearTimeout(t2);
      setIsProcessing(false);
      console.error(err);
      toast.error(err.message || 'Failed to analyze contract document');
    }
  };

  const handleLoadSampleContract = (e) => {
    e.stopPropagation();
    const sampleContent = `OCEAN TRANSPORTATION MASTER SERVICE AGREEMENT
Between Maersk Line A/S (SCAC: MAEU) and Logistics Provider Inc.
Contract Reference: MAEU-2026-0889
Effective Date: September 1, 2026
Expiration Date: August 31, 2027
Total Estimated Agreement Value: $250,000 USD
Payment terms: Net 30 days from bill of lading issuance.
Detention & Demurrage: 14 free days at all destination discharge ports.
Minimum Volume Commitment (MVC): 600 FFE.
Cargo Insurance: Carrier warrants minimum $1,000,000 cargo liability coverage.`;
    const blob = new Blob([sampleContent], { type: 'application/pdf' });
    const sampleFile = new File([blob], 'Maersk_Line_Master_Agreement_2026.pdf', { type: 'application/pdf' });
    setFile(sampleFile);
    setContractTypeHint('CARRIER_AGREEMENT');
    toast.success('Sample executed carrier agreement loaded!');
  };

  return (
    <div className="import-modal-overlay" onClick={isProcessing ? undefined : onClose}>
      <div className="import-modal-card" onClick={e => e.stopPropagation()}>
        <div className="import-modal-header">
          <div className="import-header-left">
            <div className="import-icon-badge">
              <UploadCloud size={18} />
            </div>
            <div>
              <h4>Import Contract Agreement</h4>
              <p className="import-subtitle">Upload executed document for AI clause and rate extraction</p>
            </div>
          </div>
          {!isProcessing && (
            <button className="import-close-btn" onClick={onClose} title="Close">
              <X size={18} />
            </button>
          )}
        </div>

        <div className="import-modal-body">
          {!isProcessing ? (
            <>
              {/* Document Drop Zone */}
              <div 
                className={`drop-zone ${isDragOver ? 'drag-over' : ''} ${file ? 'has-file' : ''}`}
                onDragOver={(e) => { e.preventDefault(); setIsDragOver(true); }}
                onDragLeave={() => setIsDragOver(false)}
                onDrop={handleFileDrop}
                onClick={() => fileInputRef.current?.click()}
              >
                <input 
                  type="file" 
                  ref={fileInputRef} 
                  onChange={handleFileSelect} 
                  accept=".pdf,.docx,.doc,.xlsx,.xls" 
                  style={{ display: 'none' }} 
                />

                {file ? (
                  <div className="selected-file-info">
                    <div className="file-icon-box">
                      <FileCheck size={28} className="text-blue" />
                    </div>
                    <div className="file-meta">
                      <h6>{file.name}</h6>
                      <span>{(file.size / (1024 * 1024)).toFixed(2)} MB • Ready for AI extraction</span>
                    </div>
                    <button 
                      type="button" 
                      className="btn-change-file"
                      onClick={(e) => {
                        e.stopPropagation();
                        setFile(null);
                      }}
                    >
                      Change
                    </button>
                  </div>
                ) : (
                  <div className="empty-drop-content">
                    <div className="upload-circle-icon">
                      <UploadCloud size={28} />
                    </div>
                    <h5>Click to browse or drag contract document here</h5>
                    <p>Supported formats: PDF, DOCX, Word, Excel (Max 25MB)</p>
                    <button
                      type="button"
                      className="btn-load-sample"
                      onClick={handleLoadSampleContract}
                    >
                      <Sparkles size={12} />
                      <span>Load Sample Carrier Agreement (Demo)</span>
                    </button>
                  </div>
                )}
              </div>

              {/* Agreement Type Hint Selector */}
              <div className="form-group-hint">
                <label className="hint-label">Agreement Category (Hint for AI Parser)</label>
                <div className="hints-grid">
                  {CONTRACT_TYPE_HINTS.map(hint => (
                    <div 
                      key={hint.value}
                      className={`hint-card ${contractTypeHint === hint.value ? 'selected' : ''}`}
                      onClick={() => setContractTypeHint(hint.value)}
                    >
                      <h6>{hint.label}</h6>
                      <p>{hint.sub}</p>
                    </div>
                  ))}
                </div>
              </div>

              <div className="import-notice-box">
                <Sparkles size={16} className="notice-icon" />
                <p>
                  <strong>No automatic activation:</strong> Uploading will extract terms into an editable draft.
                  You will have full control to review, edit, or cancel before creating the contract record.
                </p>
              </div>
            </>
          ) : (
            /* Animated Processing State */
            <div className="processing-container">
              <div className="processing-spinner-box">
                <Loader2 size={36} className="processing-spinner" />
                <div className="processing-sparkle">
                  <Sparkles size={16} />
                </div>
              </div>

              <h5>Extracting Contract Intelligence...</h5>
              <p className="processing-sub">
                Our Agentic AI Sidecar is analyzing <strong>{file?.name}</strong>
              </p>

              <div className="stepper-list">
                <div className={`stepper-item ${currentStep >= 1 ? (currentStep > 1 ? 'done' : 'active') : ''}`}>
                  <div className="step-dot">
                    {currentStep > 1 ? <CheckCircle2 size={14} /> : '1'}
                  </div>
                  <span>Uploading & storing agreement securely</span>
                </div>

                <div className={`stepper-item ${currentStep >= 2 ? (currentStep > 2 ? 'done' : 'active') : ''}`}>
                  <div className="step-dot">
                    {currentStep > 2 ? <CheckCircle2 size={14} /> : '2'}
                  </div>
                  <span>OCR text analysis & legal clause recognition</span>
                </div>

                <div className={`stepper-item ${currentStep >= 3 ? (currentStep > 3 ? 'done' : 'active') : ''}`}>
                  <div className="step-dot">
                    {currentStep > 3 ? <CheckCircle2 size={14} /> : '3'}
                  </div>
                  <span>Extracting commercial terms, counterparty & free days</span>
                </div>

                <div className={`stepper-item ${currentStep >= 4 ? 'done' : ''}`}>
                  <div className="step-dot">
                    {currentStep >= 4 ? <CheckCircle2 size={14} /> : '4'}
                  </div>
                  <span>Generating editable draft preview</span>
                </div>
              </div>
            </div>
          )}
        </div>

        <div className="import-modal-footer">
          {!isProcessing && (
            <>
              <button type="button" className="btn-cancel" onClick={onClose}>
                Cancel
              </button>
              <button 
                type="button" 
                className="btn-confirm-import" 
                onClick={handleStartExtraction}
                disabled={!file}
              >
                <Sparkles size={14} />
                <span>Extract Contract Draft</span>
                <ArrowRight size={14} />
              </button>
            </>
          )}
        </div>
      </div>
    </div>
  );
}

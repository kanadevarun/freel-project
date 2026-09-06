import React, { useState } from 'react';
import { 
  FileText, CheckCircle2, AlertTriangle, Sparkles, X, 
  Building2, Calendar, DollarSign, Clock, ShieldCheck, 
  ArrowLeft, ChevronDown, Check, Layers, AlertCircle
} from 'lucide-react';
import toast from 'react-hot-toast';
import { contractsService } from '../../../services/contractsService';
import './ContractImportReviewModal.css';

const CONTRACT_TYPES = [
  { value: 'CARRIER_AGREEMENT', label: 'Carrier Agreement (Supply)', sub: 'Ocean / Air rate agreement & committed allocations' },
  { value: 'CUSTOMER_SLA', label: 'Customer SLA (Client Master)', sub: 'Customer pricing commitments & guaranteed transit SLAs' },
  { value: 'VENDOR_CONTRACT', label: 'Vendor Agreement', sub: 'Drayage, warehouse storage, customs brokerage terms' },
  { value: 'FORWARDER_PARTNERSHIP', label: 'Forwarder Partnership', sub: 'Inter-forwarder agency and co-loading agreement' }
];

const TRANSPORT_MODES = [
  { value: 'OCEAN', label: 'Ocean Freight (FCL / LCL)' },
  { value: 'AIR', label: 'Air Cargo Express' },
  { value: 'ROAD', label: 'Road Transport / Drayage' },
  { value: 'RAIL', label: 'Rail Intermodal' },
  { value: 'MULTIMODAL', label: 'Multimodal / Combined' }
];

const CURRENCIES = [
  { value: 'USD', label: 'USD ($)' },
  { value: 'EUR', label: 'EUR (€)' },
  { value: 'GBP', label: 'GBP (£)' },
  { value: 'SGD', label: 'SGD (S$)' }
];

export default function ContractImportReviewModal({ 
  importData, 
  onClose, 
  onReupload, 
  onSuccess 
}) {
  const extracted = importData?.extracted_draft || {};
  const candidates = importData?.candidate_parties || [];
  const initialDupWarning = importData?.duplicate_detection?.is_duplicate 
    ? importData.duplicate_detection.message 
    : (extracted?.duplicate_warning || null);

  // Form State initialized from AI extracted draft
  const [formData, setFormData] = useState({
    contract_reference: extracted.contract_reference || '',
    contract_name: extracted.contract_name || '',
    contract_type: extracted.contract_type || 'CARRIER_AGREEMENT',
    party_id: extracted.matched_party_id || (candidates[0]?.id || 1),
    party_name: extracted.party_name || '',
    transport_mode: extracted.transport_mode || 'OCEAN',
    status: 'DRAFT',
    currency: extracted.currency || 'USD',
    contract_value: extracted.contract_value || '',
    effective_date: extracted.effective_date || '',
    expiry_date: extracted.expiry_date || '',
    owner: 'Varun Kanade',
    description: extracted.description || '',
    notes: extracted.notes || `Created via AI Document Import from ${extracted.file_name || 'agreement.pdf'}`
  });

  // Checkbox inclusions for extracted terms and obligations
  const [selectedTerms, setSelectedTerms] = useState(
    (extracted.extracted_terms || []).map((t, idx) => ({ ...t, enabled: true, index: idx }))
  );
  const [selectedObligations, setSelectedObligations] = useState(
    (extracted.extracted_obligations || []).map((o, idx) => ({ ...o, enabled: true, index: idx }))
  );

  const [duplicateWarning, setDuplicateWarning] = useState(initialDupWarning);
  const [isSubmitting, setIsSubmitting] = useState(false);

  const toggleTerm = (idx) => {
    setSelectedTerms(prev => prev.map((t, i) => i === idx ? { ...t, enabled: !t.enabled } : t));
  };

  const toggleObligation = (idx) => {
    setSelectedObligations(prev => prev.map((o, i) => i === idx ? { ...o, enabled: !o.enabled } : o));
  };

  const handlePartySelect = (e) => {
    const pId = parseInt(e.target.value, 10);
    const matched = candidates.find(c => c.id === pId);
    setFormData(prev => ({
      ...prev,
      party_id: pId,
      party_name: matched ? matched.name : prev.party_name
    }));
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    if (!formData.contract_reference.trim() || !formData.contract_name.trim() || !formData.party_name.trim()) {
      toast.error('Contract Reference, Contract Name, and Counterparty Name are mandatory.');
      return;
    }

    setIsSubmitting(true);
    const saveToast = toast.loading('Creating commercial contract and version baseline...');

    try {
      const payload = {
        document_id: importData.document_id,
        contract_reference: formData.contract_reference.trim(),
        contract_name: formData.contract_name.trim(),
        contract_type: formData.contract_type,
        party_id: Number(formData.party_id) || 1,
        party_name: formData.party_name.trim(),
        transport_mode: formData.transport_mode,
        status: 'DRAFT',
        currency: formData.currency,
        contract_value: formData.contract_value ? parseFloat(formData.contract_value) : null,
        effective_date: formData.effective_date || null,
        expiry_date: formData.expiry_date || null,
        owner: formData.owner,
        description: formData.description,
        notes: formData.notes,
        included_terms: selectedTerms.filter(t => t.enabled).map(({ enabled, index, ...rest }) => rest),
        included_obligations: selectedObligations.filter(o => o.enabled).map(({ enabled, index, ...rest }) => rest)
      };

      const created = await contractsService.confirmContractImport(payload);
      toast.success('Contract created successfully in Draft status with Version 1.0 baseline!', { id: saveToast });
      onSuccess(created.data || created);
    } catch (err) {
      console.error(err);
      toast.error(err.message || 'Failed to create contract from document', { id: saveToast });
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className="review-modal-overlay">
      <div className="review-modal-card" onClick={e => e.stopPropagation()}>
        {/* Header */}
        <div className="review-modal-header">
          <div className="review-header-left">
            <button className="btn-back" onClick={onReupload} title="Re-upload Document">
              <ArrowLeft size={16} />
            </button>
            <div className="review-title-badge">
              <Sparkles size={16} />
            </div>
            <div>
              <h4>Review & Confirm AI Extracted Contract</h4>
              <p className="review-subtitle">
                Verify extracted commercial clauses, trade lanes, and validity before creating the contract record
              </p>
            </div>
          </div>
          <button className="review-close-btn" onClick={onClose} title="Close">
            <X size={18} />
          </button>
        </div>

        {/* Duplicate Detection Warning Banner */}
        {duplicateWarning && (
          <div className="dup-warning-banner">
            <AlertTriangle size={18} className="dup-icon" />
            <div className="dup-text">
              <h6>Duplicate Reference Warning</h6>
              <p>{duplicateWarning}</p>
            </div>
          </div>
        )}

        <form onSubmit={handleSubmit} className="review-form-wrapper">
          <div className="review-modal-body">
            {/* Left Column: Source Document Card & AI Summary */}
            <div className="review-left-panel">
              <div className="source-doc-card">
                <div className="source-card-hdr">
                  <FileText size={18} className="text-blue" />
                  <div>
                    <h6>Source Agreement</h6>
                    <span className="source-file-name">{extracted.file_name || 'contract_document.pdf'}</span>
                  </div>
                </div>
                <div className="source-confidence-box">
                  <div className="conf-score-pill">
                    <CheckCircle2 size={13} />
                    <span>{extracted.overall_confidence || 92}% Extraction Confidence</span>
                  </div>
                </div>
              </div>

              {/* AI Document Abstract */}
              <div className="ai-abstract-box">
                <div className="abstract-hdr">
                  <Sparkles size={14} className="text-blue" />
                  <h6>AI Clause Summary</h6>
                </div>
                <p>
                  {extracted.ai_summary || 
                    '12-month master rate and service level agreement covering primary container trade lanes, committed allocations, and standard demurrage/detention terms.'
                  }
                </p>
              </div>

              {/* Extracted Structured Terms Checklist */}
              {selectedTerms.length > 0 && (
                <div className="clauses-checklist-box">
                  <h6>Extracted Terms ({selectedTerms.filter(t => t.enabled).length}/{selectedTerms.length})</h6>
                  <div className="terms-mini-list">
                    {selectedTerms.map((term, idx) => (
                      <div 
                        key={idx} 
                        className={`term-mini-item ${term.enabled ? 'included' : 'excluded'}`}
                        onClick={() => toggleTerm(idx)}
                      >
                        <div className={`term-check-box ${term.enabled ? 'checked' : ''}`}>
                          {term.enabled && <Check size={11} />}
                        </div>
                        <div className="term-mini-text">
                          <strong>{term.term_title}</strong>
                          <span>{term.term_value}</span>
                        </div>
                      </div>
                    ))}
                  </div>
                </div>
              )}

              {/* Extracted Obligations Checklist */}
              {selectedObligations.length > 0 && (
                <div className="clauses-checklist-box">
                  <h6>Key Covenants ({selectedObligations.filter(o => o.enabled).length}/{selectedObligations.length})</h6>
                  <div className="terms-mini-list">
                    {selectedObligations.map((ob, idx) => (
                      <div 
                        key={idx} 
                        className={`term-mini-item ${ob.enabled ? 'included' : 'excluded'}`}
                        onClick={() => toggleObligation(idx)}
                      >
                        <div className={`term-check-box ${ob.enabled ? 'checked' : ''}`}>
                          {ob.enabled && <Check size={11} />}
                        </div>
                        <div className="term-mini-text">
                          <strong>{ob.obligation_title}</strong>
                          <span>{ob.target_value ? `${ob.target_value} ${ob.metric_unit || ''}` : ob.party_responsible}</span>
                        </div>
                      </div>
                    ))}
                  </div>
                </div>
              )}
            </div>

            {/* Right Column: Editable Contract Fields */}
            <div className="review-right-panel">
              {/* Section 1: General Info */}
              <div className="form-section-card">
                <h6 className="section-title">General Agreement Metadata</h6>

                <div className="form-group">
                  <label className="field-label">
                    Contract Name <span className="req">*</span>
                  </label>
                  <input 
                    type="text" 
                    className="form-control"
                    required
                    value={formData.contract_name}
                    onChange={e => setFormData({ ...formData, contract_name: e.target.value })}
                  />
                </div>

                <div className="form-grid-2col">
                  <div className="form-group">
                    <label className="field-label">
                      Contract Reference Number <span className="req">*</span>
                    </label>
                    <input 
                      type="text" 
                      className="form-control"
                      required
                      value={formData.contract_reference}
                      onChange={e => setFormData({ ...formData, contract_reference: e.target.value })}
                    />
                  </div>

                  <div className="form-group">
                    <label className="field-label">Contract Type</label>
                    <div className="select-wrapper">
                      <select
                        className="form-control select-control"
                        value={formData.contract_type}
                        onChange={e => setFormData({ ...formData, contract_type: e.target.value })}
                      >
                        {CONTRACT_TYPES.map(t => (
                          <option key={t.value} value={t.value}>{t.label}</option>
                        ))}
                      </select>
                      <ChevronDown size={14} className="select-chevron" />
                    </div>
                  </div>
                </div>

                <div className="form-group">
                  <label className="field-label">Description & Scope</label>
                  <textarea 
                    className="form-control textarea-control"
                    rows={2}
                    value={formData.description}
                    onChange={e => setFormData({ ...formData, description: e.target.value })}
                  />
                </div>
              </div>

              {/* Section 2: Counterparty Information */}
              <div className="form-section-card">
                <h6 className="section-title">Counterparty & Registered Entity</h6>
                <div className="form-grid-2col">
                  <div className="form-group">
                    <label className="field-label">
                      Counterparty Display Name <span className="req">*</span>
                    </label>
                    <input 
                      type="text" 
                      className="form-control"
                      required
                      value={formData.party_name}
                      onChange={e => setFormData({ ...formData, party_name: e.target.value })}
                    />
                  </div>

                  <div className="form-group">
                    <label className="field-label">Link to Registered Master Entity</label>
                    <div className="select-wrapper">
                      <select
                        className="form-control select-control"
                        value={formData.party_id}
                        onChange={handlePartySelect}
                      >
                        {candidates.map(c => (
                          <option key={c.id} value={c.id}>
                            {c.name} ({c.PartyType || c.party_type}{c.SCAC ? ` - ${c.SCAC}` : ''})
                          </option>
                        ))}
                        <option value="1">Other / Custom Entity</option>
                      </select>
                      <ChevronDown size={14} className="select-chevron" />
                    </div>
                  </div>
                </div>
              </div>

              {/* Section 3: Commercial Terms & Pricing */}
              <div className="form-section-card">
                <h6 className="section-title">Commercial Terms & Free Time</h6>
                <div className="form-grid-3col">
                  <div className="form-group">
                    <label className="field-label">Contract Value (Est.)</label>
                    <input 
                      type="number" 
                      className="form-control"
                      step="0.01"
                      placeholder="0.00"
                      value={formData.contract_value}
                      onChange={e => setFormData({ ...formData, contract_value: e.target.value })}
                    />
                  </div>

                  <div className="form-group">
                    <label className="field-label">Currency</label>
                    <div className="select-wrapper">
                      <select
                        className="form-control select-control"
                        value={formData.currency}
                        onChange={e => setFormData({ ...formData, currency: e.target.value })}
                      >
                        {CURRENCIES.map(c => (
                          <option key={c.value} value={c.value}>{c.label}</option>
                        ))}
                      </select>
                      <ChevronDown size={14} className="select-chevron" />
                    </div>
                  </div>

                  <div className="form-group">
                    <label className="field-label">Transport Mode</label>
                    <div className="select-wrapper">
                      <select
                        className="form-control select-control"
                        value={formData.transport_mode}
                        onChange={e => setFormData({ ...formData, transport_mode: e.target.value })}
                      >
                        {TRANSPORT_MODES.map(m => (
                          <option key={m.value} value={m.value}>{m.label}</option>
                        ))}
                      </select>
                      <ChevronDown size={14} className="select-chevron" />
                    </div>
                  </div>
                </div>
              </div>

              {/* Section 4: Validity Dates */}
              <div className="form-section-card">
                <h6 className="section-title">Validity & Period</h6>
                <div className="form-grid-2col">
                  <div className="form-group">
                    <label className="field-label">Effective Date</label>
                    <input 
                      type="date" 
                      className="form-control date-control"
                      value={formData.effective_date}
                      onChange={e => setFormData({ ...formData, effective_date: e.target.value })}
                    />
                  </div>

                  <div className="form-group">
                    <label className="field-label">Expiry Date</label>
                    <input 
                      type="date" 
                      className="form-control date-control"
                      value={formData.expiry_date}
                      onChange={e => setFormData({ ...formData, expiry_date: e.target.value })}
                    />
                  </div>
                </div>
              </div>
            </div>
          </div>

          {/* Footer */}
          <div className="review-modal-footer">
            <div className="footer-left-info">
              <ShieldCheck size={15} className="text-muted" />
              <span>Contract will be registered in <strong>DRAFT</strong> status with <strong>Version 1.0</strong> snapshot.</span>
            </div>

            <div className="footer-actions">
              <button type="button" className="btn-cancel" onClick={onClose} disabled={isSubmitting}>
                Cancel
              </button>
              <button type="submit" className="btn-confirm-create" disabled={isSubmitting}>
                {isSubmitting ? (
                  <>
                    <Sparkles size={14} className="spin-icon" />
                    <span>Creating Contract...</span>
                  </>
                ) : (
                  <>
                    <CheckCircle2 size={15} />
                    <span>Create Contract</span>
                  </>
                )}
              </button>
            </div>
          </div>
        </form>
      </div>
    </div>
  );
}

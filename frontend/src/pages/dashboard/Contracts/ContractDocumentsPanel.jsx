import React, { useState, useEffect } from 'react';
import { 
  FileText, Plus, Shield, Award, Calendar, CheckCircle2, 
  AlertCircle, RefreshCw, Upload, History, FileCheck, Layers, Eye,
  X, ChevronDown, UploadCloud, Sparkles
} from 'lucide-react';
import contractsService from '../../../services/contractsService';
import './ContractDocumentsPanel.css';

export default function ContractDocumentsPanel({ contract, versions = [] }) {
  const [documents, setDocuments] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [showAddModal, setShowAddModal] = useState(false);
  const [supersedingDoc, setSupersedingDoc] = useState(null);

  // Form State
  const [formData, setFormData] = useState({
    document_type: 'MAIN_AGREEMENT',
    document_name: '',
    file_name: '',
    contract_version_id: '',
    effective_date: '',
    expiry_date: '',
    description: '',
  });
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    if (contract?.id) {
      fetchDocuments();
    }
  }, [contract?.id]);

  const fetchDocuments = async () => {
    try {
      setLoading(true);
      setError(null);
      const res = await contractsService.getContractDocuments(contract.id);
      setDocuments(res.data?.data || []);
    } catch (err) {
      console.error('Failed to load contract documents:', err);
      setError('Unable to load documents repository.');
    } finally {
      setLoading(false);
    }
  };

  const handleCreateDocument = async (e) => {
    e.preventDefault();
    if (!formData.document_name || !formData.file_name) return;

    try {
      setSaving(true);
      const payload = {
        document_type: formData.document_type,
        document_name: formData.document_name,
        file_name: formData.file_name,
        file_type: formData.file_name.endsWith('.pdf') ? 'PDF' : 'DOCX',
        file_size_bytes: 1024 * 450, // Realistic simulated file size
        contract_version_id: formData.contract_version_id ? parseInt(formData.contract_version_id, 10) : null,
        effective_date: formData.effective_date || null,
        expiry_date: formData.expiry_date || null,
        description: formData.description || null,
        supersedes_document_id: supersedingDoc ? supersedingDoc.id : null,
      };

      await contractsService.createContractDocument(contract.id, payload);
      setShowAddModal(false);
      setSupersedingDoc(null);
      setFormData({
        document_type: 'MAIN_AGREEMENT',
        document_name: '',
        file_name: '',
        contract_version_id: '',
        effective_date: '',
        expiry_date: '',
        description: '',
      });
      await fetchDocuments();
    } catch (err) {
      console.error('Failed to add document:', err);
      alert('Failed to save document. Please check required fields.');
    } finally {
      setSaving(false);
    }
  };

  const getDocTypeIcon = (type) => {
    switch (type) {
      case 'INSURANCE_CERTIFICATE':
      case 'COMPLIANCE_CERTIFICATE':
        return <Shield className="doc-icon shield" />;
      case 'SLA':
      case 'RATE_SCHEDULE':
        return <Award className="doc-icon award" />;
      case 'AMENDMENT':
      case 'ANNEXURE':
        return <Layers className="doc-icon layer" />;
      default:
        return <FileText className="doc-icon file" />;
    }
  };

  const activeDocs = documents.filter(d => d.is_current);
  const historicalDocs = documents.filter(d => !d.is_current);

  return (
    <div className="contract-documents-panel">
      {/* Header Bar */}
      <div className="panel-header-row">
        <div>
          <h3 className="panel-title">Contract Documents & Attachments</h3>
          <p className="panel-subtitle">
            Manage executed master agreements, annexures, rate schedules, SLAs, and certificates with immutable lineage tracking.
          </p>
        </div>
        <button 
          className="btn-primary-action"
          onClick={() => {
            setSupersedingDoc(null);
            setShowAddModal(true);
          }}
        >
          <Plus size={16} />
          <span>Upload Agreement File</span>
        </button>
      </div>

      {loading ? (
        <div className="panel-loading-state">
          <RefreshCw className="spin-icon" size={24} />
          <span>Loading verified document repository...</span>
        </div>
      ) : error ? (
        <div className="panel-error-banner">
          <AlertCircle size={18} />
          <span>{error}</span>
          <button className="btn-retry" onClick={fetchDocuments}>Retry</button>
        </div>
      ) : documents.length === 0 ? (
        <div className="panel-empty-state">
          <div className="empty-icon-circle">
            <Upload size={32} />
          </div>
          <h4>No Documents Attached</h4>
          <p>Attach the primary signed commercial agreement, insurance certificates, or rate annexures to establish complete governance.</p>
          <button 
            className="btn-primary-action"
            onClick={() => {
              setSupersedingDoc(null);
              setShowAddModal(true);
            }}
          >
            <Plus size={16} />
            <span>Upload First Document</span>
          </button>
        </div>
      ) : (
        <div className="documents-container">
          {/* Active Documents Grid */}
          <div className="docs-section">
            <div className="section-label-bar">
              <FileCheck size={16} className="text-emerald-500" />
              <h4>Active Agreement Documents ({activeDocs.length})</h4>
            </div>

            <div className="docs-grid">
              {activeDocs.map((doc) => (
                <div key={doc.id} className="doc-card active-card">
                  <div className="doc-card-top">
                    <div className="doc-type-badge-row">
                      <span className={`doc-badge type-${doc.document_type?.toLowerCase()}`}>
                        {doc.document_type?.replace(/_/g, ' ')}
                      </span>
                      {doc.is_current ? (
                        <span className="doc-status-badge current">Current</span>
                      ) : (
                        <span className="doc-status-badge superseded">Superseded</span>
                      )}
                    </div>
                    {getDocTypeIcon(doc.document_type)}
                  </div>

                  <h5 className="doc-title">{doc.document_name || doc.file_name}</h5>
                  <p className="doc-filename">{doc.file_name}</p>

                  {doc.description && (
                    <p className="doc-desc">{doc.description}</p>
                  )}

                  <div className="doc-meta-footer">
                    <div className="meta-item">
                      <Calendar size={13} />
                      <span>{doc.effective_date ? `Eff: ${doc.effective_date}` : 'No date'}</span>
                    </div>
                    {doc.expiry_date && (
                      <div className="meta-item expiry">
                        <span>Exp: {doc.expiry_date}</span>
                      </div>
                    )}
                  </div>

                  <div className="doc-card-actions">
                    <button 
                      className="btn-action-supersede"
                      title="Upload a newer revision to supersede this document"
                      onClick={() => {
                        setSupersedingDoc(doc);
                        setFormData({
                          ...formData,
                          document_type: doc.document_type,
                          document_name: `${doc.document_name || doc.file_name} (Revised)`,
                        });
                        setShowAddModal(true);
                      }}
                    >
                      <RefreshCw size={13} />
                      <span>Supersede Revision</span>
                    </button>
                  </div>
                </div>
              ))}
            </div>
          </div>

          {/* Historical Lineage Section */}
          {historicalDocs.length > 0 && (
            <div className="docs-section historical-section">
              <div className="section-label-bar">
                <History size={16} className="text-slate-400" />
                <h4>Superseded & Historical Documents ({historicalDocs.length})</h4>
              </div>

              <div className="docs-grid">
                {historicalDocs.map((doc) => (
                  <div key={doc.id} className="doc-card historical-card">
                    <div className="doc-card-top">
                      <span className="doc-badge type-historical">
                        {doc.document_type?.replace(/_/g, ' ')}
                      </span>
                      <span className="doc-status-badge superseded">Superseded</span>
                    </div>

                    <h5 className="doc-title">{doc.document_name || doc.file_name}</h5>
                    <p className="doc-filename">{doc.file_name}</p>

                    <div className="doc-meta-footer">
                      <div className="meta-item">
                        <History size={13} />
                        <span>Preserved Historical Audit</span>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )}
        </div>
      )}

      {/* Add / Supersede Document Modal */}
      {showAddModal && (
        <div className="modal-backdrop" onClick={() => setShowAddModal(false)}>
          <div className="doc-modal-content" onClick={e => e.stopPropagation()}>
            <div className="doc-modal-header">
              <div className="doc-modal-header-left">
                <div className="doc-header-badge">
                  <UploadCloud size={18} />
                </div>
                <div>
                  <h4>{supersedingDoc ? `Supersede: ${supersedingDoc.document_name || supersedingDoc.file_name}` : 'Attach Contract Document'}</h4>
                  <p className="doc-modal-subtitle">Upload executed commercial agreements, annexures, rate sheets, and compliance certificates</p>
                </div>
              </div>
              <button className="btn-close-modal" onClick={() => setShowAddModal(false)} title="Close">
                <X size={16} />
              </button>
            </div>

            <form onSubmit={handleCreateDocument} className="doc-form">
              {supersedingDoc && (
                <div className="supersede-alert-banner">
                  <AlertCircle size={16} />
                  <span>The new document will become the active version, while <strong>{supersedingDoc.file_name}</strong> will be archived into historical audit lineage.</span>
                </div>
              )}

              <div className="doc-form-group">
                <label className="doc-label">
                  Document Category <span className="doc-req">*</span>
                </label>
                <div className="doc-select-wrapper">
                  <select 
                    className="doc-select"
                    value={formData.document_type}
                    onChange={(e) => setFormData({ ...formData, document_type: e.target.value })}
                    required
                  >
                    <option value="MAIN_AGREEMENT">Main Agreement</option>
                    <option value="AMENDMENT">Amendment Addendum</option>
                    <option value="ANNEXURE">Annexure / Appendix</option>
                    <option value="RATE_SCHEDULE">Rate Schedule Sheet</option>
                    <option value="SLA">Service Level Agreement (SLA)</option>
                    <option value="INSURANCE_CERTIFICATE">Insurance Certificate</option>
                    <option value="COMPLIANCE_CERTIFICATE">Compliance / KYC Certificate</option>
                    <option value="TERMS_AND_CONDITIONS">Standard Terms & Conditions</option>
                    <option value="OTHER">Other Contractual File</option>
                  </select>
                  <ChevronDown size={14} className="doc-select-chevron" />
                </div>
              </div>

              <div className="doc-form-group">
                <label className="doc-label">
                  Document Display Name <span className="doc-req">*</span>
                </label>
                <input 
                  type="text"
                  className="doc-input"
                  placeholder="e.g. Master Freight Agreement 2026 Executed"
                  value={formData.document_name}
                  onChange={(e) => setFormData({ ...formData, document_name: e.target.value })}
                  required
                />
              </div>

              <div className="doc-form-group">
                <label className="doc-label">
                  File Name <span className="doc-req">*</span>
                </label>
                <input 
                  type="text"
                  className="doc-input doc-input-mono"
                  placeholder="e.g. Carrier_Master_Agreement_Signed.pdf"
                  value={formData.file_name}
                  onChange={(e) => setFormData({ ...formData, file_name: e.target.value })}
                  required
                />
              </div>

              {versions && versions.length > 0 && (
                <div className="doc-form-group">
                  <label className="doc-label">Associate with Version</label>
                  <div className="doc-select-wrapper">
                    <select
                      className="doc-select"
                      value={formData.contract_version_id}
                      onChange={(e) => setFormData({ ...formData, contract_version_id: e.target.value })}
                    >
                      <option value="">Default (Current Effective Contract)</option>
                      {versions.map(v => (
                        <option key={v.id} value={v.id}>
                          {v.version_label} (v{v.version_number}) - {v.status}
                        </option>
                      ))}
                    </select>
                    <ChevronDown size={14} className="doc-select-chevron" />
                  </div>
                </div>
              )}

              <div className="doc-form-row-2">
                <div className="doc-form-group">
                  <label className="doc-label">Effective Date</label>
                  <input 
                    type="date"
                    className="doc-input doc-input-date"
                    value={formData.effective_date}
                    onChange={(e) => setFormData({ ...formData, effective_date: e.target.value })}
                  />
                </div>
                <div className="doc-form-group">
                  <label className="doc-label">Expiry Date</label>
                  <input 
                    type="date"
                    className="doc-input doc-input-date"
                    value={formData.expiry_date}
                    onChange={(e) => setFormData({ ...formData, expiry_date: e.target.value })}
                  />
                </div>
              </div>

              <div className="doc-form-group">
                <label className="doc-label">Notes / Document Scope</label>
                <textarea 
                  rows={2}
                  className="doc-textarea"
                  placeholder="Summary of terms, clauses, or annexure scope contained in this file..."
                  value={formData.description}
                  onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                />
              </div>

              <div className="modal-actions-bar">
                <button type="button" className="btn-cancel" onClick={() => setShowAddModal(false)}>
                  Cancel
                </button>
                <button type="submit" className="btn-submit" disabled={saving}>
                  <UploadCloud size={14} />
                  <span>{saving ? 'Saving...' : supersedingDoc ? 'Supersede & Attach' : 'Attach Document'}</span>
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}

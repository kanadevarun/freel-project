import React, { useState, useEffect, useCallback } from 'react';
import toast from 'react-hot-toast';
import { contractsService } from '../../../services/contractsService';
import PageHeader from '../../../components/dashboard/PageHeader';
import ReviewModal from './ReviewModal';
import AgentStatusTimeline from './AgentStatusTimeline';
import './ContractDocumentsPage.css';


export default function ContractDocumentsPage() {
  const [activeTab, setActiveTab] = useState('CONTRACTS');
  const [documents, setDocuments] = useState([]);
  const [reviewItems, setReviewItems] = useState([]);
  const [isLoadingDocs, setIsLoadingDocs] = useState(true);
  const [isLoadingReview, setIsLoadingReview] = useState(true);
  const [selectedDoc, setSelectedDoc] = useState(null);
  const [selectedReviewItem, setSelectedReviewItem] = useState(null);

  // File Upload State
  const [file, setFile] = useState(null);
  const [carrierSCAC, setCarrierSCAC] = useState('');
  const [isUploading, setIsUploading] = useState(false);

  const fetchDocuments = useCallback(async () => {
    setIsLoadingDocs(true);
    try {
      const docs = await contractsService.listDocuments();
      setDocuments(docs || []);
    } catch (err) {
      console.error(err);
      toast.error('Failed to load contract documents');
    } finally {
      setIsLoadingDocs(false);
    }
  }, []);

  const fetchReviewItems = useCallback(async () => {
    setIsLoadingReview(true);
    try {
      const items = await contractsService.listReviewItems('PENDING');
      setReviewItems(items || []);
    } catch (err) {
      console.error(err);
      toast.error('Failed to load review items');
    } finally {
      setIsLoadingReview(false);
    }
  }, []);

  useEffect(() => {
    fetchDocuments();
    fetchReviewItems();
  }, [fetchDocuments, fetchReviewItems]);

  const handleFileChange = (e) => {
    if (e.target.files && e.target.files[0]) {
      setFile(e.target.files[0]);
    }
  };

  const handleUploadSubmit = async (e) => {
    e.preventDefault();
    if (!file) {
      toast.error('Please select a PDF or Excel document to upload.');
      return;
    }
    setIsUploading(true);
    const uploadToast = toast.loading('Uploading contract to secure storage...');
    try {
      await contractsService.uploadContract(file, carrierSCAC || null);
      toast.success('Contract uploaded successfully! AI Sidecar pipeline triggered.', { id: uploadToast });
      setFile(null);
      setCarrierSCAC('');
      fetchDocuments();
    } catch (err) {
      console.error(err);
      toast.error(err.message || 'Upload failed', { id: uploadToast });
    } finally {
      setIsUploading(false);
    }
  };

  const handleReprocess = async (docId) => {
    const reprocessToast = toast.loading('Re-queuing document for AI extraction...');
    try {
      await contractsService.reprocessDocument(docId);
      toast.success('Reprocessing pipeline triggered.', { id: reprocessToast });
      fetchDocuments();
    } catch (err) {
      console.error(err);
      toast.error('Failed to trigger reprocessing', { id: reprocessToast });
    }
  };

  // Stats
  const totalDocs = documents.length;
  const processingDocs = documents.filter(d => ['QUEUED', 'OCR_PROCESSING', 'AI_EXTRACTING'].includes(d.status)).length;
  const completedDocs = documents.filter(d => d.status === 'CONFIRMED').length;
  const pendingReviewCount = reviewItems.length;

  return (
    <div className="contracts-page-container">
      <div className="page-header-wrapper">
        <PageHeader
          title="Contract Rate Intelligence"
          subtitle="AI-powered multi-agent extraction, note matching, and human review for ocean carrier contracts"
        />
      </div>

      {/* ── KPI Widgets ─────────────────────────────────────────────────── */}
      <div className="contracts-stats-grid">
        <div className="kpi-card teal-glow">
          <div className="kpi-icon">📜</div>
          <div className="kpi-data">
            <span className="kpi-value">{totalDocs}</span>
            <span className="kpi-label">Total Documents</span>
          </div>
        </div>
        <div className="kpi-card indigo-glow">
          <div className="kpi-icon">⚙️</div>
          <div className="kpi-data">
            <span className="kpi-value">{processingDocs}</span>
            <span className="kpi-label">AI Processing</span>
          </div>
        </div>
        <div className="kpi-card green-glow">
          <div className="kpi-icon">✓</div>
          <div className="kpi-data">
            <span className="kpi-value">{completedDocs}</span>
            <span className="kpi-label">Confirmed Contracts</span>
          </div>
        </div>
        <div className="kpi-card red-glow">
          <div className="kpi-icon">⚠️</div>
          <div className="kpi-data">
            <span className="kpi-value">{pendingReviewCount}</span>
            <span className="kpi-label">Awaiting Verification</span>
          </div>
        </div>
      </div>

      {/* ── Tab Switcher ───────────────────────────────────────────────── */}
      <div className="contracts-tabs">
        <button
          className={`contracts-tab-btn ${activeTab === 'CONTRACTS' ? 'active' : ''}`}
          onClick={() => setActiveTab('CONTRACTS')}
        >
          📂 Carrier Contracts
        </button>
        <button
          className={`contracts-tab-btn ${activeTab === 'REVIEW' ? 'active' : ''}`}
          onClick={() => setActiveTab('REVIEW')}
        >
          📋 Human Review Queue {pendingReviewCount > 0 && <span className="tab-badge">{pendingReviewCount}</span>}
        </button>
      </div>

      {activeTab === 'CONTRACTS' ? (
        <div className="contracts-tab-content">
          <div className="contracts-workspace-layout">
            
            {/* Left Column: Upload Form */}
            <div className="workspace-sidebar">
              <div className="upload-card card-premium">
                <div className="card-header">
                  <h4>Ingest New Contract</h4>
                  <p>Upload a carrier agreement or tariff schedule (PDF or XLSX)</p>
                </div>
                <form onSubmit={handleUploadSubmit} className="upload-form">
                  <div className="form-group">
                    <label>Carrier SCAC (Optional)</label>
                    <select
                      value={carrierSCAC}
                      onChange={(e) => setCarrierSCAC(e.target.value)}
                      className="form-select"
                    >
                      <option value="">Auto-Detect SCAC</option>
                      <option value="MAEU">MAEU — Maersk Line</option>
                      <option value="MSCU">MSCU — Mediterranean Shipping Co</option>
                      <option value="CMDU">CMDU — CMA CGM</option>
                      <option value="ONEY">ONEY — Ocean Network Express</option>
                      <option value="HLCU">HLCU — Hapag-Lloyd</option>
                    </select>
                  </div>

                  <div className="form-group">
                    <label>Document File</label>
                    <div className="file-drop-area">
                      <input
                        type="file"
                        accept=".pdf,.xlsx,.xls"
                        onChange={handleFileChange}
                        id="contract-file-input"
                        className="file-input-hidden"
                      />
                      <label htmlFor="contract-file-input" className="file-drop-label">
                        <span className="file-drop-icon">📤</span>
                        <span className="file-drop-text">
                          {file ? file.name : 'Choose file or drag here'}
                        </span>
                        <span className="file-drop-subtext">PDF or Excel up to 50MB</span>
                      </label>
                    </div>
                  </div>

                  <button
                    type="submit"
                    disabled={isUploading || !file}
                    className="btn-primary w-full upload-btn"
                  >
                    {isUploading ? 'Ingesting document...' : '🚀 Upload & Run AI Analysis'}
                  </button>
                </form>
              </div>
            </div>

            {/* Right Column: Documents Table */}
            <div className="workspace-main">
              <div className="documents-list-card card-premium">
                <div className="card-header">
                  <h4>Uploaded Contracts Log</h4>
                  <p>Monitor layout extraction stages, validity dates, and logs</p>
                </div>

                {isLoadingDocs ? (
                  <div className="contracts-loading">
                    <div className="spinner" />
                    <span>Querying contract archive...</span>
                  </div>
                ) : documents.length === 0 ? (
                  <div className="empty-state">
                    <div className="empty-icon">📁</div>
                    <h5>No documents found</h5>
                    <p>Get started by uploading your first shipping contract PDF on the left panel.</p>
                  </div>
                ) : (
                  <div className="table-responsive">
                    <table className="contracts-table">
                      <thead>
                        <tr>
                          <th>File Details</th>
                          <th>Carrier</th>
                          <th>Status</th>
                          <th>Confidence</th>
                          <th>Extracted Rates</th>
                          <th>Processed Date</th>
                          <th>Actions</th>
                        </tr>
                      </thead>
                      <tbody>
                        {documents.map((doc) => (
                          <tr key={doc.id} className="document-row">
                            <td onClick={() => setSelectedDoc(doc)}>
                              <div className="file-name-cell">
                                <span className="doc-icon">{doc.file_type === 'XLSX' ? '📊' : '📕'}</span>
                                <div className="file-meta">
                                  <span className="file-title">{doc.file_name}</span>
                                  <span className="file-size">{(doc.file_size / 1024 / 1024).toFixed(2)} MB</span>
                                </div>
                              </div>
                            </td>
                            <td>
                              <span className="carrier-badge">
                                {doc.carrier_scac || 'Detecting...'}
                              </span>
                            </td>
                            <td>
                              <span className={`status-pill status-${doc.status.toLowerCase()}`}>
                                {doc.status.replace('_', ' ')}
                              </span>
                            </td>
                            <td>
                              <div className="confidence-indicator">
                                <div className="confidence-bar-bg">
                                  <div
                                    className="confidence-bar-fill"
                                    style={{
                                      width: `${doc.status === 'CONFIRMED' ? 95 : doc.status === 'FAILED' ? 0 : 50}%`,
                                      backgroundColor: doc.status === 'CONFIRMED' ? 'var(--color-success)' : 'var(--color-warning)'
                                    }}
                                  />
                                </div>
                                <span className="confidence-text">
                                  {doc.status === 'CONFIRMED' ? '95%' : doc.status === 'FAILED' ? '0%' : 'Analyzing...'}
                                </span>
                              </div>
                            </td>
                            <td>
                              <div className="rates-count-summary">
                                <span className="count-confirmed" title="Auto-confirmed rate entries">{doc.confirmed_rates_count} ✓</span>
                                <span className="count-flagged" title="Flagged items in review queue">{doc.flagged_rates_count} ⚠️</span>
                              </div>
                            </td>
                            <td>{new Date(doc.created_at).toLocaleDateString()}</td>
                            <td>
                              <div className="actions-cell">
                                <button
                                  onClick={() => handleReprocess(doc.id)}
                                  className="btn-icon"
                                  title="Reprocess Document"
                                >
                                  🔄
                                </button>
                              </div>
                            </td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>
                )}
              </div>
            </div>
          </div>

          {/* Drawer Detail View for Selected Document */}
          {selectedDoc && (
            <div className="document-drawer-backdrop" onClick={() => setSelectedDoc(null)}>
              <div className="document-drawer" onClick={(e) => e.stopPropagation()}>
                <div className="drawer-header">
                  <h3>Document Detail</h3>
                  <button className="btn-close-drawer" onClick={() => setSelectedDoc(null)}>✕</button>
                </div>
                <div className="drawer-content">
                  <div className="meta-grid">
                    <div className="meta-item">
                      <span className="meta-label">File Name</span>
                      <span className="meta-val">{selectedDoc.file_name}</span>
                    </div>
                    <div className="meta-item">
                      <span className="meta-label">SCAC Code</span>
                      <span className="meta-val">{selectedDoc.carrier_scac || 'N/A'}</span>
                    </div>
                    <div className="meta-item">
                      <span className="meta-label">File Type</span>
                      <span className="meta-val">{selectedDoc.file_type}</span>
                    </div>
                    <div className="meta-item">
                      <span className="meta-label">Status</span>
                      <span className={`status-pill status-${selectedDoc.status.toLowerCase()}`}>{selectedDoc.status}</span>
                    </div>
                  </div>

                  {selectedDoc.ai_summary && (
                    <div className="summary-section">
                      <h5>AI Document Abstract</h5>
                      <p className="summary-text">{selectedDoc.ai_summary}</p>
                    </div>
                  )}

                  <AgentStatusTimeline
                    logs={selectedDoc.processing_log || []}
                    status={selectedDoc.status}
                  />
                </div>
              </div>
            </div>
          )}
        </div>
      ) : (
        /* Tab 2: Verification Queue */
        <div className="contracts-tab-content">
          <div className="review-queue-card card-premium">
            <div className="card-header">
              <h4>Flagged Rate Anomalies</h4>
              <p>Resolve price validations and port mapping ambiguities before rates go live</p>
            </div>

            {isLoadingReview ? (
              <div className="contracts-loading">
                <div className="spinner" />
                <span>Checking review queue...</span>
              </div>
            ) : reviewItems.length === 0 ? (
              <div className="empty-state">
                <div className="empty-icon">✓</div>
                <h5>Queue is clean</h5>
                <p>No rates are currently flagged for verification. High-confidence rates auto-confirm.</p>
              </div>
            ) : (
              <div className="table-responsive">
                <table className="contracts-table">
                  <thead>
                    <tr>
                      <th>Carrier</th>
                      <th>Origin</th>
                      <th>Destination</th>
                      <th>Equipment</th>
                      <th>Flagged Price</th>
                      <th>Review Flag</th>
                      <th>AI Reasoning</th>
                      <th>Confidence</th>
                      <th>Action</th>
                    </tr>
                  </thead>
                  <tbody>
                    {reviewItems.map((item) => {
                      let rateData = {};
                      try {
                        rateData = typeof item.extracted_data === 'string'
                          ? JSON.parse(item.extracted_data)
                          : item.extracted_data || {};
                      } catch (e) {
                        rateData = {};
                      }

                      return (
                        <tr key={item.id} className="review-row">
                          <td>
                            <div className="carrier-meta">
                              <strong>{rateData.carrier_name || 'Maersk'}</strong>
                              <span>{rateData.carrier_scac || 'MAEU'}</span>
                            </div>
                          </td>
                          <td><code>{rateData.origin_port}</code></td>
                          <td><code>{rateData.destination_port}</code></td>
                          <td>{rateData.equipment_type}</td>
                          <td><span className="text-danger font-semibold">${rateData.ocean_freight}</span></td>
                          <td>
                            {(item.review_flags || []).map((flag) => (
                              <span key={flag} className="flag-badge">
                                ⚠️ {flag}
                              </span>
                            ))}
                          </td>
                          <td className="reasoning-cell" title={item.ai_reasoning}>
                            {item.ai_reasoning}
                          </td>
                          <td>
                            <span className="confidence-pct text-yellow-600 font-medium">
                              {item.confidence}%
                            </span>
                          </td>
                          <td>
                            <button
                              onClick={() => setSelectedReviewItem(item)}
                              className="btn-secondary btn-sm"
                            >
                              Verify Rate
                            </button>
                          </td>
                        </tr>
                      );
                    })}
                  </tbody>
                </table>
              </div>
            )}
          </div>
        </div>
      )}

      {/* Review Modal */}
      {selectedReviewItem && (
        <ReviewModal
          item={selectedReviewItem}
          onClose={() => setSelectedReviewItem(null)}
          onSuccess={() => {
            setSelectedReviewItem(null);
            fetchReviewItems();
            fetchDocuments();
          }}
        />
      )}
    </div>
  );
}

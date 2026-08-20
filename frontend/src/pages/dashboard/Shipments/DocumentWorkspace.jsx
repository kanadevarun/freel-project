import React, { useState, useEffect } from 'react';

export default function DocumentWorkspace({ shipmentId, token, onRefreshShipment }) {
  const [documents, setDocuments] = useState([]);
  const [discrepancies, setDiscrepancies] = useState([]);
  const [uploading, setUploading] = useState(false);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [uploadType, setUploadType] = useState('MBL');

  useEffect(() => {
    fetchDocuments();
  }, [shipmentId]);

  const fetchDocuments = async () => {
    try {
      setLoading(true);
      const response = await fetch(`/api/v1/shipments/${shipmentId}/documents`, {
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        }
      });
      if (!response.ok) {
        throw new Error(`HTTP error ${response.status}`);
      }
      const data = await response.json();
      setDocuments(data.data.documents || []);
      setDiscrepancies(data.data.discrepancies || []);
      setError('');
    } catch (err) {
      console.error('Failed to fetch documents:', err);
      setError('Could not load compliance documentation workspace.');
    } finally {
      setLoading(false);
    }
  };

  const handleFileUpload = async (e) => {
    const file = e.target.files[0];
    if (!file) return;

    try {
      setUploading(true);
      setError('');
      // Simulate file uploading to S3 - setting key and file name mock values
      const uploadPayload = {
        doc_type: uploadType,
        s3_key: `uploads/shipments/${shipmentId}/${uploadType.toLowerCase()}_${Date.now()}.pdf`,
        file_name: file.name,
        file_type: 'PDF'
      };

      const response = await fetch(`/api/v1/shipments/${shipmentId}/documents/upload`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify(uploadPayload)
      });

      if (!response.ok) {
        throw new Error(`Upload failed: ${response.statusText}`);
      }

      // Refresh documents list
      await fetchDocuments();
      if (onRefreshShipment) onRefreshShipment();
      
      // Auto-poll verification details 3 times
      let count = 0;
      const interval = setInterval(async () => {
        count++;
        await fetchDocuments();
        if (onRefreshShipment) onRefreshShipment();
        if (count >= 3) clearInterval(interval);
      }, 3000);

    } catch (err) {
      console.error('Upload failed:', err);
      setError('Failed to upload and verify compliance document.');
    } finally {
      setUploading(false);
    }
  };

  const handleResolveDiscrepancy = async (discId) => {
    try {
      const response = await fetch(`/api/v1/shipments/discrepancies/${discId}/resolve`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        }
      });
      if (response.ok) {
        fetchDocuments();
        if (onRefreshShipment) onRefreshShipment();
      } else {
        alert('Failed to resolve discrepancy.');
      }
    } catch (err) {
      console.error('Error resolving discrepancy:', err);
    }
  };

  if (loading && documents.length === 0) {
    return <div className="workspace-loading">Loading documentation verification context...</div>;
  }

  return (
    <div className="glass-card panel-card doc-workspace">
      <div className="workspace-header">
        <div>
          <h3>Compliance Workspace</h3>
          <p className="panel-desc">Manage Bills of Lading, Commercial Invoices, and cross-document reconciliation audits.</p>
        </div>
        <div className="upload-controls">
          <select 
            value={uploadType} 
            onChange={(e) => setUploadType(e.target.value)}
            className="doc-type-select"
          >
            <option value="MBL">Master Bill of Lading (MBL)</option>
            <option value="HBL">House Bill of Lading (HBL)</option>
            <option value="CI">Commercial Invoice (CI)</option>
            <option value="PL">Packing List (PL)</option>
          </select>
          <label className={`btn-primary btn-upload ${uploading ? 'disabled' : ''}`}>
            {uploading ? 'Processing OCR...' : 'Upload File'}
            <input 
              type="file" 
              accept=".pdf,.png,.jpg,.jpeg" 
              onChange={handleFileUpload} 
              style={{ display: 'none' }}
              disabled={uploading}
            />
          </label>
        </div>
      </div>

      {error && <div className="error-banner">{error}</div>}

      <div className="workspace-split">
        {/* Documents list */}
        <div className="docs-list-side">
          <h4>Uploaded Documentation ({documents.length})</h4>
          {documents.length === 0 ? (
            <p className="empty-workspace-txt">No documents uploaded. Please upload a Bill of Lading to verify details.</p>
          ) : (
            <div className="docs-grid-list">
              {documents.map((doc) => (
                <div key={doc.id} className="doc-item-card">
                  <div className="doc-meta-row">
                    <span className="doc-badge-type">{doc.doc_type}</span>
                    <span className={`doc-status-pill status-${doc.status.toLowerCase()}`}>
                      {doc.status.replace('_', ' ')}
                    </span>
                  </div>
                  <h5 className="doc-filename">{doc.file_name}</h5>
                  
                  {doc.extracted_data && Object.keys(doc.extracted_data).length > 0 && (
                    <div className="doc-extracted-summary">
                      <h6>Extracted Parameters:</h6>
                      <ul>
                        {doc.extracted_data.gross_weight && (
                          <li>Weight: <strong>{doc.extracted_data.gross_weight} KGS</strong></li>
                        )}
                        {doc.extracted_data.container_number && (
                          <li>Container: <strong>{doc.extracted_data.container_number}</strong></li>
                        )}
                        {doc.extracted_data.package_count && (
                          <li>Packages: <strong>{doc.extracted_data.package_count}</strong></li>
                        )}
                      </ul>
                    </div>
                  )}

                  {doc.raw_ocr_text && (
                    <details className="doc-ocr-details">
                      <summary>View raw OCR output</summary>
                      <pre className="ocr-raw-display">{doc.raw_ocr_text}</pre>
                    </details>
                  )}
                </div>
              ))}
            </div>
          )}
        </div>

        {/* Discrepancies listing */}
        <div className="discrepancies-side">
          <h4>Discrepancies & Mismatches ({discrepancies.length})</h4>
          {discrepancies.length === 0 ? (
            <div className="no-mismatches-box">
              <span className="clean-audit-icon">🛡️</span>
              <p>Clean compliance audit. No discrepancies detected between uploaded files.</p>
            </div>
          ) : (
            <div className="discrepancies-grid">
              {discrepancies.map((disc) => (
                <div key={disc.id} className={`discrepancy-card status-${disc.status.toLowerCase()}`}>
                  <div className="discrepancy-header-row">
                    <span className="disc-field">Field: <strong>{disc.field_name.replace('_', ' ')}</strong></span>
                    {disc.status === 'OPEN' ? (
                      <button 
                        className="btn-resolve-disc" 
                        onClick={() => handleResolveDiscrepancy(disc.id)}
                      >
                        Override / Resolve
                      </button>
                    ) : (
                      <span className="disc-resolved-txt">✓ Resolved</span>
                    )}
                  </div>
                  <div className="disc-comparison-grid">
                    <div className="compare-block">
                      <small className="block-label">{disc.source_document} (Expected)</small>
                      <span className="compare-val">{disc.expected_value}</span>
                    </div>
                    <div className="compare-arrow">➔</div>
                    <div className="compare-block">
                      <small className="block-label">{disc.target_document} (Actual)</small>
                      <span className="compare-val val-mismatch">{disc.actual_value}</span>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

import { useState, useRef } from 'react';
import PropTypes from 'prop-types';
import { importLeads } from '../../../services/leadsService';

/**
 * ImportLeadsModal — A CSV file upload modal for bulk importing leads.
 *
 * Simple meaning: When the user clicks "Import CSV", this modal appears with
 * a drag-and-drop zone. They can drag their CSV file onto it (or click to browse).
 * Once they pick a file, clicking "Upload & Import" sends it to the backend which
 * will parse all the rows and kick off AI scoring for every single lead automatically.
 *
 * Expected CSV columns (case-insensitive):
 *   Company Name | Contact Name | Email | Source
 *
 * @param {{ onClose: () => void, onImportComplete: () => void }}
 */
export default function ImportLeadsModal({ onClose, onImportComplete }) {
  const [selectedFile, setSelectedFile] = useState(null);   // The File object the user picked
  const [dragging, setDragging] = useState(false);          // Is the user dragging a file over?
  const [loading, setLoading] = useState(false);            // Is the upload in progress?
  const [result, setResult] = useState(null);               // The import result from backend
  const [error, setError] = useState('');                   // Any error message
  const fileInputRef = useRef(null);                        // Reference to the hidden file input

  // ── File selection ──────────────────────────────────────────────────────────

  function handleFileChange(e) {
    const file = e.target.files?.[0];
    if (file) handleFile(file);
  }

  function handleFile(file) {
    // Only accept CSV files
    if (!file.name.endsWith('.csv')) {
      setError('Please upload a .csv file.');
      return;
    }
    setError('');
    setResult(null);
    setSelectedFile(file);
  }

  // ── Drag and drop handlers ─────────────────────────────────────────────────

  function handleDragOver(e) {
    e.preventDefault(); // Required to allow drop
    setDragging(true);
  }

  function handleDragLeave() {
    setDragging(false);
  }

  function handleDrop(e) {
    e.preventDefault();
    setDragging(false);
    const file = e.dataTransfer.files?.[0];
    if (file) handleFile(file);
  }

  // ── Upload ─────────────────────────────────────────────────────────────────

  async function handleUpload() {
    if (!selectedFile) return;

    setLoading(true);
    setError('');

    try {
      // Send the file to the backend (POST /api/v1/leads/import)
      const importResult = await importLeads(selectedFile);
      setResult(importResult);
      onImportComplete(); // Tell the parent to reload the leads table
    } catch (err) {
      setError(err.message || 'Import failed. Please check the file format and try again.');
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="leads-modal-overlay" onClick={onClose}>
      <div className="leads-modal" onClick={e => e.stopPropagation()}>

        {/* Header */}
        <div className="modal-header">
          <div className="modal-title">📥 Import Leads from CSV</div>
          <button className="modal-close-btn" onClick={onClose}>✕</button>
        </div>

        {/* Success state — show after successful import */}
        {result ? (
          <div>
            <div className="import-result">
              ✅ Successfully imported <strong>{result.imported_count}</strong> leads out of {result.total_attempted} rows!
              The AI scoring worker is now running in the background.
            </div>
            <div style={{ textAlign: 'center', marginTop: 20 }}>
              <button className="leads-btn leads-btn-primary" onClick={onClose}>
                🎉 Done! View Leads
              </button>
            </div>
          </div>
        ) : (
          <>
            {/* Drag and Drop Zone */}
            <div
              className={`import-dropzone ${dragging ? 'dragging' : ''}`}
              onClick={() => fileInputRef.current?.click()} // clicking the zone opens file picker
              onDragOver={handleDragOver}
              onDragLeave={handleDragLeave}
              onDrop={handleDrop}
            >
              {/* Hidden real file input */}
              <input
                ref={fileInputRef}
                type="file"
                accept=".csv"
                style={{ display: 'none' }}
                onChange={handleFileChange}
              />

              <div className="import-dropzone-icon">📂</div>
              {selectedFile ? (
                <>
                  <div className="import-dropzone-title">File selected!</div>
                  <div className="import-file-name">{selectedFile.name}</div>
                </>
              ) : (
                <>
                  <div className="import-dropzone-title">Drag & drop your CSV file here</div>
                  <div className="import-dropzone-sub">or click to browse</div>
                </>
              )}
            </div>

            {/* Expected format hint */}
            <div className="import-sample-note">
              <strong>Expected CSV column headers:</strong><br />
              Company Name, Contact Name, Email, Source<br />
              <em>Column order doesn&apos;t matter. Extra columns will be ignored.</em>
            </div>

            {/* Error */}
            {error && (
              <div style={{ background: 'rgba(239,68,68,0.08)', border: '1px solid rgba(239,68,68,0.2)', borderRadius: 8, padding: '10px 14px', fontSize: 13, color: '#dc2626', marginTop: 12 }}>
                {error}
              </div>
            )}

            {/* Actions */}
            <div className="modal-actions" style={{ marginTop: 16 }}>
              <button className="leads-btn leads-btn-ghost" onClick={onClose}>Cancel</button>
              <button
                className="leads-btn leads-btn-primary"
                onClick={handleUpload}
                disabled={!selectedFile || loading}
              >
                {loading ? '⏳ Uploading...' : '⬆️ Upload & Import'}
              </button>
            </div>
          </>
        )}

      </div>
    </div>
  );
}

ImportLeadsModal.propTypes = {
  onClose: PropTypes.func.isRequired,
  onImportComplete: PropTypes.func.isRequired,
};

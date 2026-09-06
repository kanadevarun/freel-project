import React from 'react';
import { FileText, Eye, Download, Upload } from 'lucide-react';
import './InvoiceSubTabs.css';

export default function InvoiceDocumentsTab({ invoice, onActionClick }) {
  const documents = invoice?.documents || [];

  return (
    <div className="invoice-subtab-container documents-tab">
      <div className="tab-top-actions">
        <span className="docs-count-title">All Documents ({documents.length})</span>
        <button
          className="btn-secondary-action btn-sm"
          onClick={() => onActionClick?.('uploadDoc', invoice)}
        >
          <Upload size={13} /> Upload Document
        </button>
      </div>

      {documents.length === 0 ? (
        <div className="empty-subtab">
          <FileText size={28} className="empty-icon-gray" />
          <p className="empty-subtab-text">No documents attached.</p>
        </div>
      ) : (
        <div className="full-doc-list">
          {documents.map((doc) => (
            <div key={doc.id} className="full-doc-card">
              <div className="doc-icon-box">
                <FileText size={20} className="pdf-icon-red" />
              </div>
              <div className="doc-info">
                <span className="doc-name">{doc.name}</span>
                <span className="doc-sub">{doc.size} • Uploaded {doc.uploadedAt}</span>
              </div>
              <div className="doc-actions-group">
                <button
                  className="doc-action-btn"
                  title="View Document"
                  onClick={() => onActionClick?.('viewDoc', doc)}
                >
                  <Eye size={14} />
                </button>
                <button
                  className="doc-action-btn"
                  title="Download Document"
                  onClick={() => onActionClick?.('downloadDoc', doc)}
                >
                  <Download size={14} />
                </button>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}

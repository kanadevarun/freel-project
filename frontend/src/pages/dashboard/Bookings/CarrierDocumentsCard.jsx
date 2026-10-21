import React, { useState, useEffect, useCallback } from 'react';
import { 
  FileText, ExternalLink, Download, RefreshCw, 
  AlertCircle, Info, ShieldAlert, CheckCircle2 
} from 'lucide-react';
import bookingService from '../../../services/bookingService';

export default function CarrierDocumentsCard({ bookingId, carrierName, carrierBookingRef }) {
  const [documents, setDocuments] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [errorCode, setErrorCode] = useState(null);
  const [isUnsupported, setIsUnsupported] = useState(false);

  const fetchDocuments = useCallback(async () => {
    if (!bookingId) return;
    setLoading(true);
    setError(null);
    setErrorCode(null);
    setIsUnsupported(false);

    try {
      const res = await bookingService.getCarrierDocuments(bookingId);
      const data = res?.data?.data ?? res?.data ?? [];
      setDocuments(Array.isArray(data) ? data : []);
    } catch (err) {
      const respData = err.response?.data;
      const code = respData?.error?.code || respData?.code || '';
      const message = respData?.error?.message || respData?.message || err.message || 'Failed to load documents';

      setErrorCode(code);

      if (code === 'UNSUPPORTED_OPERATION' || message.includes('does not support capability') || message.includes('DOCUMENTS')) {
        setIsUnsupported(true);
      } else if (err.response?.status === 401 || code === 'AUTHENTICATION_FAILED') {
        setError('Carrier authentication failed. Please verify carrier credentials in Settings.');
      } else if (err.response?.status === 403 || code === 'AUTHORIZATION_FAILED') {
        setError('Permission denied. You do not have authorization to view documents for this carrier.');
      } else {
        setError(message);
      }
    } finally {
      setLoading(false);
    }
  }, [bookingId]);

  useEffect(() => {
    fetchDocuments();
  }, [fetchDocuments]);

  return (
    <div className="detail-card carrier-documents-card" data-testid="carrier-documents-card">
      <div className="detail-card-header" style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <h3 className="detail-card-title" style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
          <FileText size={18} className="text-blue-500" />
          <span>Carrier Transport Documents</span>
        </h3>
        <button
          type="button"
          className="btn-refresh"
          data-testid="refresh-documents-btn"
          style={{ 
            fontSize: '0.78rem', 
            padding: '4px 10px', 
            display: 'inline-flex', 
            alignItems: 'center', 
            gap: '5px',
            cursor: loading ? 'not-allowed' : 'pointer'
          }}
          onClick={fetchDocuments}
          disabled={loading}
          title="Refresh documents from carrier"
        >
          <RefreshCw size={13} className={loading ? 'animate-spin' : ''} />
          <span>{loading ? 'Fetching...' : 'Refresh'}</span>
        </button>
      </div>

      <div style={{ padding: '16px' }}>
        {/* Loading State */}
        {loading && (
          <div data-testid="documents-loading" style={{ textAlign: 'center', padding: '24px 16px', color: '#64748b' }}>
            <div className="skeleton-pulse" style={{ display: 'inline-block', fontSize: '1.4rem' }}>···</div>
            <p style={{ marginTop: '8px', fontSize: '0.88rem', fontWeight: 600 }}>
              Retrieving live transport documents from {carrierName || 'carrier'}...
            </p>
          </div>
        )}

        {/* Unsupported Capability State */}
        {!loading && isUnsupported && (
          <div 
            data-testid="documents-unsupported" 
            style={{ 
              background: '#f8fafc', 
              border: '1px solid #e2e8f0', 
              borderRadius: '8px', 
              padding: '16px', 
              display: 'flex', 
              gap: '12px', 
              alignItems: 'flex-start' 
            }}
          >
            <Info size={20} style={{ color: '#64748b', flexShrink: 0, marginTop: '2px' }} />
            <div>
              <div style={{ fontWeight: 700, fontSize: '0.88rem', color: '#334155' }}>
                Documents Capability Not Supported
              </div>
              <p style={{ margin: '4px 0 0 0', fontSize: '0.82rem', color: '#64748b' }}>
                Electronic transport document retrieval is not supported by {carrierName || 'this carrier'}. Documents must be obtained directly through the carrier portal.
              </p>
            </div>
          </div>
        )}

        {/* Error State */}
        {!loading && !isUnsupported && error && (
          <div 
            data-testid="documents-error" 
            style={{ 
              background: '#fef2f2', 
              border: '1px solid #fecaca', 
              borderRadius: '8px', 
              padding: '16px', 
              display: 'flex', 
              flexDirection: 'column', 
              gap: '10px' 
            }}
          >
            <div style={{ display: 'flex', gap: '10px', alignItems: 'flex-start' }}>
              <AlertCircle size={20} style={{ color: '#dc2626', flexShrink: 0, marginTop: '2px' }} />
              <div>
                <div style={{ fontWeight: 700, fontSize: '0.88rem', color: '#991b1b' }}>
                  Unable to Retrieve Documents
                </div>
                <p style={{ margin: '4px 0 0 0', fontSize: '0.82rem', color: '#b91c1c' }}>
                  {error}
                </p>
                {errorCode && (
                  <span style={{ fontSize: '0.74rem', color: '#7f1d1d', fontFamily: 'monospace' }}>
                    Code: {errorCode}
                  </span>
                )}
              </div>
            </div>
            <div style={{ alignSelf: 'flex-start', marginLeft: '30px' }}>
              <button
                type="button"
                data-testid="retry-documents-btn"
                className="btn-detail-action"
                style={{ 
                  background: '#dc2626', 
                  color: '#ffffff', 
                  fontSize: '0.78rem', 
                  padding: '4px 12px',
                  borderRadius: '4px'
                }}
                onClick={fetchDocuments}
              >
                Retry Request
              </button>
            </div>
          </div>
        )}

        {/* Empty State */}
        {!loading && !isUnsupported && !error && documents.length === 0 && (
          <div data-testid="documents-empty" style={{ textAlign: 'center', padding: '24px 16px', color: '#64748b' }}>
            <FileText size={32} style={{ color: '#94a3b8', margin: '0 auto 8px auto', opacity: 0.6 }} />
            <h4 style={{ margin: 0, fontSize: '0.92rem', fontWeight: 700, color: '#334155' }}>
              No documents available.
            </h4>
            <p style={{ margin: '6px 0 0 0', fontSize: '0.82rem', color: '#64748b' }}>
              No electronic transport documents have been issued by {carrierName || 'the carrier'} yet
              {carrierBookingRef ? ` for booking ${carrierBookingRef}` : ''}.
            </p>
          </div>
        )}

        {/* Success State: Documents List */}
        {!loading && !isUnsupported && !error && documents.length > 0 && (
          <div data-testid="documents-list" style={{ display: 'flex', flexDirection: 'column', gap: '10px' }}>
            {documents.map((doc, idx) => {
              const docType = doc.document_type || doc.DocumentType || 'Transport Document';
              const docNum = doc.reference || doc.document_id || doc.DocumentID || doc.Reference || '—';
              const docUrl = doc.document_url || doc.download_url || doc.DocumentURL || '';
              const docDate = doc.created_at || doc.issue_date || doc.CreatedAt;
              const formattedDate = docDate ? new Date(docDate).toLocaleDateString(undefined, {
                year: 'numeric', month: 'short', day: 'numeric'
              }) : 'Available';

              return (
                <div 
                  key={doc.document_id || idx}
                  data-testid="document-item"
                  style={{
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'space-between',
                    padding: '12px 14px',
                    background: '#f8fafc',
                    borderRadius: '8px',
                    border: '1px solid #e2e8f0',
                    gap: '12px',
                    flexWrap: 'wrap'
                  }}
                >
                  <div style={{ display: 'flex', alignItems: 'center', gap: '12px' }}>
                    <div style={{
                      width: '36px',
                      height: '36px',
                      borderRadius: '8px',
                      background: '#eff6ff',
                      color: '#2563eb',
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'center',
                      flexShrink: 0
                    }}>
                      <FileText size={18} />
                    </div>
                    <div>
                      <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                        <span style={{ fontWeight: 700, fontSize: '0.88rem', color: '#0f172a' }}>
                          {docType}
                        </span>
                        <span style={{
                          background: '#e0f2fe',
                          color: '#0369a1',
                          padding: '1px 6px',
                          borderRadius: '4px',
                          fontSize: '0.72rem',
                          fontWeight: 600,
                          fontFamily: 'monospace'
                        }}>
                          {docNum}
                        </span>
                      </div>
                      <div style={{ fontSize: '0.78rem', color: '#64748b', marginTop: '2px' }}>
                        Carrier: <strong>{carrierName || doc.carrier_scac || 'Maersk'}</strong> • Issued: {formattedDate}
                        {doc.mime_type && ` • ${doc.mime_type}`}
                      </div>
                    </div>
                  </div>

                  <div>
                    {docUrl ? (
                      <a
                        href={docUrl}
                        target="_blank"
                        rel="noopener noreferrer"
                        data-testid="document-download-link"
                        className="btn-detail-action"
                        style={{
                          background: '#2563eb',
                          color: '#ffffff',
                          fontSize: '0.78rem',
                          padding: '6px 12px',
                          borderRadius: '6px',
                          textDecoration: 'none',
                          display: 'inline-flex',
                          alignItems: 'center',
                          gap: '6px',
                          fontWeight: 600
                        }}
                      >
                        <Download size={13} />
                        <span>Download / View</span>
                      </a>
                    ) : (
                      <span style={{ fontSize: '0.78rem', color: '#94a3b8', fontStyle: 'italic' }}>
                        Pending Generation
                      </span>
                    )}
                  </div>
                </div>
              );
            })}
          </div>
        )}
      </div>
    </div>
  );
}

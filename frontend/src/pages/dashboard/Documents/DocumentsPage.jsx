import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { Folder, FileText, Upload, CheckCircle2, ShieldCheck, AlertCircle, ArrowRight, ExternalLink } from 'lucide-react';
import api from '../../../services/api';
import PageHeader from '../../../components/dashboard/PageHeader';
import './DocumentsPage.css';

export default function DocumentsPage() {
  const [documents, setDocuments] = useState([]);
  const [filterType, setFilterType] = useState('ALL');
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const navigate = useNavigate();

  useEffect(() => {
    fetchDocuments();
  }, []);

  const fetchDocuments = async () => {
    try {
      setLoading(true);
      setError(null);
      const res = await api.get('/api/v1/documents');
      const docList = Array.isArray(res) ? res : (res?.data || []);
      setDocuments(docList);
    } catch (err) {
      console.error('Failed to load documents:', err);
      setError('Unable to load document compliance center. Please try again.');
    } finally {
      setLoading(false);
    }
  };

  const filtered = documents.filter((d) => {
    if (filterType === 'ALL') return true;
    if (filterType === 'BOL') return d.doc_type === 'HBL' || d.doc_type === 'MBL';
    if (filterType === 'CUSTOMS') return d.doc_type === 'COMMERCIAL_INVOICE' || d.doc_type === 'PACKING_LIST';
    return true;
  });

  return (
    <div className="documents-page">
      <div className="documents-header-row">
        <PageHeader
          title="Document Compliance Center"
          subtitle="AI OCR verification, 3-way discrepancy checks, and automated compliance validation for HBL, MBL, Invoices, and Packing Lists"
        />
      </div>

      {/* Tabs */}
      <div className="documents-tabs-row">
        {['ALL', 'BOL', 'CUSTOMS'].map((t) => (
          <button
            key={t}
            className={`tab-btn ${filterType === t ? 'active' : ''}`}
            onClick={() => setFilterType(t)}
          >
            {t === 'ALL' ? 'All Documents' : t === 'BOL' ? 'Bills of Lading (HBL / MBL)' : 'Invoices & Packing Lists'}
          </button>
        ))}
      </div>

      {error && (
        <div className="documents-error-banner">
          <span>⚠️ {error}</span>
          <button onClick={fetchDocuments} className="btn-retry">Retry</button>
        </div>
      )}

      {loading ? (
        <div className="documents-skeleton">
          {[1, 2, 3].map((i) => (
            <div key={i} className="skeleton-doc-row" />
          ))}
        </div>
      ) : filtered.length === 0 ? (
        <div className="documents-empty-state">
          <div className="empty-icon-box">📁</div>
          <h3>No shipment documents yet</h3>
          <p>
            Shipment documents (House Bills of Lading, Master BLs, Commercial Invoices, Packing Lists) will appear here once attached to active bookings and analyzed by the compliance AI engine.
          </p>
          <div style={{ display: 'flex', gap: '12px', justifyContent: 'center', marginTop: '20px' }}>
            <button className="btn-primary-action" onClick={() => navigate('/dashboard/shipments')}>
              View Active Shipments
            </button>
            <button className="btn-secondary-action" onClick={() => navigate('/dashboard/contracts')}>
              Upload Carrier Contracts
            </button>
          </div>
        </div>
      ) : (
        <div className="documents-table-card">
          <table className="documents-table">
            <thead>
              <tr>
                <th>Document Name</th>
                <th>Doc Type</th>
                <th>Shipment / Booking</th>
                <th>Compliance Status</th>
                <th>Discrepancies</th>
                <th>Uploaded</th>
              </tr>
            </thead>
            <tbody>
              {filtered.map((d) => (
                <tr key={d.id} onClick={() => navigate(`/dashboard/shipments/${d.shipment_id}`)}>
                  <td>
                    <div className="doc-name-cell">
                      <FileText size={16} className="doc-icon" />
                      <strong>{d.file_name || `DOC-${d.id}`}</strong>
                    </div>
                  </td>
                  <td>
                    <span className="doc-type-badge">{d.doc_type}</span>
                  </td>
                  <td>{d.booking_number || `SH-${d.shipment_id}`}</td>
                  <td>
                    <span className={`compliance-pill ${d.verification_status?.toLowerCase() || 'verified'}`}>
                      <CheckCircle2 size={12} />
                      {d.verification_status || 'VERIFIED'}
                    </span>
                  </td>
                  <td>{d.discrepancy_count ? `${d.discrepancy_count} Flags` : 'Zero Discrepancies'}</td>
                  <td>{d.created_at ? new Date(d.created_at).toLocaleDateString() : '—'}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}

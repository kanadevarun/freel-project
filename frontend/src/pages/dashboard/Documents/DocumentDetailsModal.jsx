import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  X, FileText, Download, Trash2, ExternalLink, Calendar, HardDrive,
  ShieldCheck, AlertTriangle, Building, Truck, FileCode, CheckCircle2,
  Clock, PlusCircle
} from 'lucide-react';
import api from '../../../services/api';
import { approvalsService } from '../../../services/approvalsService';

export default function DocumentDetailsModal({ isOpen, onClose, document: doc, onDeleteSuccess, onRequestApprovalSuccess }) {
  const [isDeleting, setIsDeleting] = useState(false);
  const [showConfirmDelete, setShowConfirmDelete] = useState(false);
  const [requestingApproval, setRequestingApproval] = useState(false);
  const [approvalSuccessMsg, setApprovalSuccessMsg] = useState('');
  const [error, setError] = useState('');
  const navigate = useNavigate();

  if (!isOpen || !doc) return null;

  const getDownloadUrl = () => {
    if (!doc) return '#';
    if (doc.file_path && doc.file_path.startsWith('http')) return doc.file_path;
    return `http://localhost:8080/uploads/${doc.s3_key || doc.file_name}`;
  };

  const formatFileSize = (bytes) => {
    if (!bytes) return '—';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i];
  };

  const handleDelete = async () => {
    try {
      setIsDeleting(true);
      setError('');
      await api.delete(`/api/v1/documents/${doc.id}`);
      if (onDeleteSuccess) {
        onDeleteSuccess(doc.id);
      }
      setShowConfirmDelete(false);
      onClose();
    } catch (err) {
      console.error('Delete document failed:', err);
      setError(err?.response?.data?.message || 'Failed to delete document. Please try again.');
    } finally {
      setIsDeleting(false);
    }
  };

  const handleRequestApproval = async () => {
    try {
      setRequestingApproval(true);
      setError('');
      setApprovalSuccessMsg('');

      const docName = doc.original_file_name || doc.file_name || `DOC-${doc.id}`;
      const parsedDocId = parseInt(doc.id, 10) || 0;
      const parsedShipmentId = parseInt(doc.shipment_id, 10) || 0;
      const parsedCustomerId = parseInt(doc.customer_id, 10) || 0;

      await approvalsService.createApproval({
        title: `${doc.doc_type || 'Document'} Approval - ${docName}`,
        category: 'DOCUMENTS',
        type: 'Document Approval',
        priority: 'HIGH',
        related_ref: `Document: ${docName}`,
        customer_name: doc.customer_name || 'Associated Customer',
        document_id: parsedDocId,
        shipment_id: parsedShipmentId,
        customer_id: parsedCustomerId,
        description: `Compliance sign-off requested for document: ${docName} (${doc.doc_type || 'DOCUMENT'}).`,
      });

      setApprovalSuccessMsg('Approval request submitted successfully to Approval Center.');
      if (onRequestApprovalSuccess) {
        onRequestApprovalSuccess(doc.id);
      }
      setTimeout(() => setApprovalSuccessMsg(''), 4000);
    } catch (err) {
      console.error('Failed to request approval:', err);
      setError(err?.message || 'Failed to submit approval request.');
    } finally {
      setRequestingApproval(false);
    }
  };

  return (
    <div className="doc-upload-modal-overlay">
      <div className="doc-upload-modal-container" style={{ maxWidth: '620px' }}>
        {/* Header */}
        <div className="doc-upload-modal-header" style={{ background: '#F8FAFC' }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
            <div style={{ width: 42, height: 42, borderRadius: 12, background: '#EFF6FF', border: '1px solid #BFDBFE', display: 'flex', alignItems: 'center', justifyContent: 'center', color: '#2563EB' }}>
              <FileText size={22} />
            </div>
            <div>
              <h3 style={{ margin: 0, fontSize: '1.1rem', fontWeight: 800, color: '#0F172A' }}>
                {doc.original_file_name || doc.file_name || `DOC-${doc.id}`}
              </h3>
              <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginTop: 3 }}>
                <span className="doc-type-badge">{doc.doc_type}</span>
                <span style={{ fontSize: '0.76rem', color: '#64748B' }}>ID: #{doc.id}</span>
              </div>
            </div>
          </div>
          <button className="doc-upload-modal-close" onClick={onClose} disabled={isDeleting}>
            <X size={18} />
          </button>
        </div>

        {error && (
          <div className="doc-upload-modal-error">
            <AlertTriangle size={16} />
            <span>{error}</span>
          </div>
        )}

        {approvalSuccessMsg && (
          <div style={{ background: '#ECFDF5', border: '1px solid #A7F3D0', color: '#047857', padding: '10px 16px', margin: '16px 24px 0 24px', borderRadius: 8, display: 'flex', alignItems: 'center', gap: 8, fontSize: '0.84rem', fontWeight: 650 }}>
            <CheckCircle2 size={16} />
            <span>{approvalSuccessMsg}</span>
          </div>
        )}

        <div className="doc-upload-modal-body" style={{ gap: 20 }}>
          {/* File Metadata Grid */}
          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 12, background: '#F8FAFC', padding: 16, borderRadius: 12, border: '1px solid #F1F5F9' }}>
            <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
              <HardDrive size={16} style={{ color: '#64748B' }} />
              <div>
                <div style={{ fontSize: '0.72rem', color: '#64748B', fontWeight: 600 }}>File Size</div>
                <div style={{ fontSize: '0.84rem', fontWeight: 750, color: '#0F172A' }}>{formatFileSize(doc.file_size)}</div>
              </div>
            </div>

            <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
              <FileCode size={16} style={{ color: '#64748B' }} />
              <div>
                <div style={{ fontSize: '0.72rem', color: '#64748B', fontWeight: 600 }}>Format / MIME</div>
                <div style={{ fontSize: '0.84rem', fontWeight: 750, color: '#0F172A' }}>{doc.file_type || 'PDF'} ({doc.mime_type || 'application/pdf'})</div>
              </div>
            </div>

            <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
              <Calendar size={16} style={{ color: '#64748B' }} />
              <div>
                <div style={{ fontSize: '0.72rem', color: '#64748B', fontWeight: 600 }}>Upload Date</div>
                <div style={{ fontSize: '0.84rem', fontWeight: 750, color: '#0F172A' }}>
                  {doc.created_at ? new Date(doc.created_at).toLocaleString() : '—'}
                </div>
              </div>
            </div>

            <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
              <ShieldCheck size={16} style={{ color: '#059669' }} />
              <div>
                <div style={{ fontSize: '0.72rem', color: '#64748B', fontWeight: 600 }}>Compliance Status</div>
                <div style={{ fontSize: '0.84rem', fontWeight: 750, color: '#059669', display: 'flex', alignItems: 'center', gap: 4 }}>
                  <CheckCircle2 size={13} /> {doc.status || 'VERIFIED'}
                </div>
              </div>
            </div>
          </div>

          {/* Linked Business Relationships & Approval Trigger */}
          <div style={{ background: '#FFFFFF', border: '1px solid #E2E8F0', borderRadius: 12, padding: 16 }}>
            <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: 12 }}>
              <h4 style={{ margin: 0, fontSize: '0.85rem', fontWeight: 800, color: '#334155' }}>
                Linked Business Context & Approvals
              </h4>
              <button
                type="button"
                onClick={() => { onClose(); navigate('/dashboard/approvals'); }}
                style={{ background: 'none', border: 'none', color: '#2563EB', fontSize: '0.76rem', fontWeight: 750, cursor: 'pointer', display: 'flex', alignItems: 'center', gap: 4 }}
              >
                Approval Center <ExternalLink size={12} />
              </button>
            </div>

            <div style={{ display: 'flex', flexDirection: 'column', gap: 10, fontSize: '0.82rem' }}>
              <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', padding: '8px 12px', background: '#F8FAFC', borderRadius: 8 }}>
                <span style={{ display: 'flex', alignItems: 'center', gap: 6, color: '#64748B', fontWeight: 600 }}>
                  <Truck size={15} /> Shipment
                </span>
                {doc.shipment_id ? (
                  <button
                    type="button"
                    onClick={() => { onClose(); navigate(`/dashboard/shipments/${doc.shipment_id}`); }}
                    style={{ background: 'none', border: 'none', color: '#2563EB', fontWeight: 750, cursor: 'pointer', display: 'flex', alignItems: 'center', gap: 4, padding: 0 }}
                  >
                    {doc.shipment_ref || `SH-${doc.shipment_id}`} <ExternalLink size={12} />
                  </button>
                ) : (
                  <span style={{ fontWeight: 750, color: '#94A3B8' }}>Not linked</span>
                )}
              </div>

              <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', padding: '8px 12px', background: '#F8FAFC', borderRadius: 8 }}>
                <span style={{ display: 'flex', alignItems: 'center', gap: 6, color: '#64748B', fontWeight: 600 }}>
                  <Building size={15} /> Customer
                </span>
                {doc.customer_id ? (
                  <button
                    type="button"
                    onClick={() => { onClose(); navigate(`/dashboard/customers/${doc.customer_id}`); }}
                    style={{ background: 'none', border: 'none', color: '#2563EB', fontWeight: 750, cursor: 'pointer', display: 'flex', alignItems: 'center', gap: 4, padding: 0 }}
                  >
                    {doc.customer_name || `Customer #${doc.customer_id}`} <ExternalLink size={12} />
                  </button>
                ) : (
                  <span style={{ fontWeight: 750, color: '#94A3B8' }}>Not linked</span>
                )}
              </div>
            </div>

            {/* Request Approval Button */}
            <div style={{ marginTop: 14, paddingTop: 12, borderTop: '1px solid #F1F5F9', display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
              <div style={{ fontSize: '0.78rem', color: '#64748B' }}>
                Need sign-off from compliance or operations manager?
              </div>
              <button
                type="button"
                onClick={handleRequestApproval}
                disabled={requestingApproval}
                style={{ background: '#FFFBEB', border: '1px solid #FDE68A', color: '#D97706', padding: '6px 12px', borderRadius: 8, fontSize: '0.78rem', fontWeight: 750, cursor: 'pointer', display: 'flex', alignItems: 'center', gap: 6 }}
              >
                <Clock size={14} /> {requestingApproval ? 'Submitting...' : 'Request Approval'}
              </button>
            </div>
          </div>

          {/* Delete Confirmation Step */}
          {showConfirmDelete ? (
            <div style={{ background: '#FEF2F2', border: '1px solid #FECACA', borderRadius: 12, padding: 14, display: 'flex', flexDirection: 'column', gap: 10 }}>
              <div style={{ fontSize: '0.84rem', fontWeight: 750, color: '#991B1B' }}>
                ⚠️ Are you sure you want to delete this document record?
              </div>
              <div style={{ fontSize: '0.78rem', color: '#B91C1C' }}>
                This will permanently remove the document from the compliance database.
              </div>
              <div style={{ display: 'flex', gap: 8, justifyContent: 'flex-end', marginTop: 4 }}>
                <button
                  type="button"
                  className="btn-cancel"
                  onClick={() => setShowConfirmDelete(false)}
                  disabled={isDeleting}
                >
                  Cancel
                </button>
                <button
                  type="button"
                  style={{ background: '#DC2626', color: '#FFFFFF', border: 'none', padding: '7px 14px', borderRadius: 8, fontSize: '0.82rem', fontWeight: 750, cursor: 'pointer' }}
                  onClick={handleDelete}
                  disabled={isDeleting}
                >
                  {isDeleting ? 'Deleting...' : 'Confirm Delete'}
                </button>
              </div>
            </div>
          ) : null}

          {/* Action Buttons */}
          <div className="doc-upload-modal-footer" style={{ borderTop: '1px solid #F1F5F9', paddingTop: 14 }}>
            <button
              type="button"
              style={{ background: '#FEF2F2', border: '1px solid #FECACA', color: '#DC2626', padding: '8px 14px', borderRadius: 8, fontSize: '0.82rem', fontWeight: 700, cursor: 'pointer', display: 'flex', alignItems: 'center', gap: 6, marginRight: 'auto' }}
              onClick={() => setShowConfirmDelete(true)}
              disabled={showConfirmDelete || isDeleting}
            >
              <Trash2 size={15} /> Delete Document
            </button>

            <a
              href={getDownloadUrl()}
              target="_blank"
              rel="noreferrer"
              style={{ background: '#2563EB', color: '#FFFFFF', border: 'none', padding: '8px 18px', borderRadius: 8, fontSize: '0.82rem', fontWeight: 750, textDecoration: 'none', display: 'inline-flex', alignItems: 'center', gap: 6 }}
            >
              <Download size={16} /> Download File
            </a>
          </div>
        </div>
      </div>
    </div>
  );
}

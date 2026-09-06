import React, { useState, useEffect, useMemo } from 'react';
import { Link } from 'react-router-dom';
import {
  Upload,
  FileText,
  AlertTriangle,
  ShieldCheck,
  AlertCircle,
  Plus,
  Search,
  Check,
  X,
  Trash2,
  Download,
  ShieldAlert,
  ArrowUpRight,
  Info,
  Calendar,
  RotateCw,
  Edit2,
  Clock,
  CheckCircle2,
  FileCheck,
  FileSpreadsheet,
  FileCode,
  Tag,
  Filter
} from 'lucide-react';
import shipmentService from '../../../services/shipmentService';
import CustomDropdown from './CustomDropdown';
import {
  DOC_CATEGORIES,
  DOC_TYPES,
  DOC_STATUSES,
  DOC_STATUS_CONFIG,
  COMPLIANCE_STATES,
  COMPLIANCE_STATE_CONFIG,
  REQUIREMENT_LEVELS,
  REQUIREMENT_LEVEL_CONFIG,
  VALIDITY_STATES,
  VALIDITY_STATE_CONFIG,
  DOC_SOURCES,
  CATEGORY_TABS,
  CATEGORY_DROPDOWN_OPTIONS,
  DOC_TYPE_OPTIONS
} from './shipmentDocumentConstants';
import './Shipments.css';

export default function DocumentWorkspace({ shipmentId, shipment, onRefreshShipment }) {
  const [documents, setDocuments] = useState([]);
  const [compliance, setCompliance] = useState(null);
  const [discrepancies, setDiscrepancies] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [activeTab, setActiveTab] = useState(DOC_CATEGORIES.ALL);
  const [statusFilter, setStatusFilter] = useState('ALL');
  const [searchQuery, setSearchQuery] = useState('');

  // Upload Modal State
  const [showUploadModal, setShowUploadModal] = useState(false);
  const [uploadCategory, setUploadCategory] = useState(DOC_CATEGORIES.TRANSPORT);
  const [uploadType, setUploadType] = useState(DOC_TYPES.MBL);
  const [uploadDocName, setUploadDocName] = useState('');
  const [uploadRefNum, setUploadRefNum] = useState('');
  const [uploadDocDate, setUploadDocDate] = useState('');
  const [uploadExpiry, setUploadExpiry] = useState('');
  const [uploadDesc, setUploadDesc] = useState('');
  const [selectedFile, setSelectedFile] = useState(null);
  const [submittingUpload, setSubmittingUpload] = useState(false);
  const [uploadError, setUploadError] = useState('');

  // Edit Metadata Modal State
  const [editingDoc, setEditingDoc] = useState(null);
  const [editName, setEditName] = useState('');
  const [editCategory, setEditCategory] = useState(DOC_CATEGORIES.TRANSPORT);
  const [editRefNum, setEditRefNum] = useState('');
  const [editDocDate, setEditDocDate] = useState('');
  const [editExpiry, setEditExpiry] = useState('');
  const [editDesc, setEditDesc] = useState('');
  const [submittingEdit, setSubmittingEdit] = useState(false);
  const [editError, setEditError] = useState('');

  // Reject Modal State
  const [rejectingDoc, setRejectingDoc] = useState(null);
  const [rejectReason, setRejectReason] = useState('');
  const [submittingReject, setSubmittingReject] = useState(false);

  // Action in progress ID
  const [actionInProgressId, setActionInProgressId] = useState(null);

  useEffect(() => {
    fetchDocuments();
  }, [shipmentId]);

  const [refreshing, setRefreshing] = useState(false);
  const [justRefreshed, setJustRefreshed] = useState(false);

  const fetchDocuments = async (isInitial = false) => {
    const startTime = Date.now();
    try {
      if (isInitial) {
        setLoading(true);
      } else {
        setRefreshing(true);
      }
      setError('');
      const res = await shipmentService.getShipmentDocuments(shipmentId);
      const data = res?.data || res;
      if (data && Array.isArray(data.documents)) {
        setDocuments(data.documents);
        setCompliance(data.compliance_summary || null);
        setDiscrepancies(Array.isArray(data.discrepancies) ? data.discrepancies : []);
      }

      // Maintain deliberate visible refresh duration (650ms) so the refresh state is clearly experienced
      if (!isInitial) {
        const elapsed = Date.now() - startTime;
        if (elapsed < 650) {
          await new Promise(resolve => setTimeout(resolve, 650 - elapsed));
        }
        setJustRefreshed(true);
        setTimeout(() => setJustRefreshed(false), 2200);
      }
    } catch (err) {
      console.warn('Document workspace fetch error:', err?.message || err);
      // Only set error banner if we have no documents loaded yet
      if (documents.length === 0) {
        setError("Unable to load shipment documents. We couldn't retrieve the latest document list. Please try again.");
      }
    } finally {
      setLoading(false);
      setRefreshing(false);
    }
  };

  const handleOpenUploadModal = (
    preselectedCategory = DOC_CATEGORIES.TRANSPORT,
    preselectedType = DOC_TYPES.MBL,
    suggestedTitle = ''
  ) => {
    setUploadCategory(preselectedCategory || DOC_CATEGORIES.TRANSPORT);
    setUploadType(preselectedType || DOC_TYPES.MBL);
    setUploadDocName(suggestedTitle || '');
    setUploadRefNum('');
    setUploadDocDate(new Date().toISOString().split('T')[0]);
    setUploadExpiry('');
    setUploadDesc('');
    setSelectedFile(null);
    setUploadError('');
    setShowUploadModal(true);
  };

  const handleCategoryChange = (cat) => {
    setUploadCategory(cat);
    const firstOption = DOC_TYPE_OPTIONS.find((o) => o.category === cat);
    if (firstOption) {
      setUploadType(firstOption.value);
    }
  };

  const handleFileChange = (e) => {
    const file = e.target.files[0];
    if (file) {
      setSelectedFile(file);
      if (!uploadDocName) {
        setUploadDocName(file.name.replace(/\.[^/.]+$/, ''));
      }
    }
  };

  const handleUploadSubmit = async (e) => {
    e.preventDefault();
    if (!selectedFile && !uploadDocName.trim()) {
      setUploadError('Please select a file or enter a document title.');
      return;
    }

    try {
      setSubmittingUpload(true);
      setUploadError('');

      const fileName = selectedFile ? selectedFile.name : `${uploadDocName.trim() || uploadType}.pdf`;
      const fileSize = selectedFile ? selectedFile.size : 1024;
      const mimeType = selectedFile?.type || 'application/pdf';

      const payload = {
        doc_type: uploadType,
        document_name: uploadDocName.trim() || fileName,
        category: uploadCategory,
        description: uploadDesc.trim(),
        file_name: fileName,
        file_size: fileSize,
        mime_type: mimeType,
        file_url: `/uploads/shipments/${shipmentId}/${uploadType}_${Date.now()}_${fileName}`,
        reference_number: uploadRefNum.trim() || undefined,
        document_date: uploadDocDate ? new Date(uploadDocDate).toISOString() : new Date().toISOString(),
        expires_at: uploadExpiry ? new Date(uploadExpiry).toISOString() : undefined,
      };

      await shipmentService.uploadShipmentDocument(shipmentId, payload);
      setShowUploadModal(false);
      await fetchDocuments();
      if (onRefreshShipment) onRefreshShipment();
    } catch (err) {
      console.error('Failed to upload document:', err);
      setUploadError('Unable to upload document. Please verify the file and try again.');
    } finally {
      setSubmittingUpload(false);
    }
  };

  const handleOpenEditModal = (doc) => {
    setEditingDoc(doc);
    setEditName(doc.document_name || doc.file_name || '');
    setEditCategory(doc.category || DOC_CATEGORIES.TRANSPORT);
    setEditRefNum(doc.reference_number || '');
    setEditDocDate(doc.document_date ? new Date(doc.document_date).toISOString().split('T')[0] : '');
    setEditExpiry(doc.expires_at ? new Date(doc.expires_at).toISOString().split('T')[0] : '');
    setEditDesc(doc.description || '');
    setEditError('');
  };

  const handleEditSubmit = async (e) => {
    e.preventDefault();
    if (!editingDoc) return;

    try {
      setSubmittingEdit(true);
      setEditError('');

      const payload = {
        document_name: editName.trim(),
        category: editCategory,
        reference_number: editRefNum.trim() || undefined,
        document_date: editDocDate ? new Date(editDocDate).toISOString() : undefined,
        expires_at: editExpiry ? new Date(editExpiry).toISOString() : undefined,
        description: editDesc.trim(),
      };

      await shipmentService.updateShipmentDocument(shipmentId, editingDoc.id, payload);
      setEditingDoc(null);
      await fetchDocuments();
      if (onRefreshShipment) onRefreshShipment();
    } catch (err) {
      console.error('Failed to update document:', err);
      setEditError('Unable to update document metadata. Please try again.');
    } finally {
      setSubmittingEdit(false);
    }
  };

  const handleApproveDoc = async (doc) => {
    try {
      setActionInProgressId(doc.id);
      await shipmentService.approveShipmentDocument(shipmentId, doc.id);
      await fetchDocuments();
      if (onRefreshShipment) onRefreshShipment();
    } catch (err) {
      alert('Unable to approve document. Please try again.');
    } finally {
      setActionInProgressId(null);
    }
  };

  const handleOpenRejectModal = (doc) => {
    setRejectingDoc(doc);
    setRejectReason('');
  };

  const handleRejectSubmit = async (e) => {
    e.preventDefault();
    if (!rejectReason.trim()) {
      alert('Please enter an explicit reason for rejecting this document.');
      return;
    }

    try {
      setSubmittingReject(true);
      await shipmentService.rejectShipmentDocument(shipmentId, rejectingDoc.id, rejectReason.trim());
      setRejectingDoc(null);
      await fetchDocuments();
      if (onRefreshShipment) onRefreshShipment();
    } catch (err) {
      alert('Unable to reject document. Please try again.');
    } finally {
      setSubmittingReject(false);
    }
  };

  const handleDeleteDoc = async (doc) => {
    if (!window.confirm(`Are you sure you want to remove document "${doc.document_name || doc.file_name}"?`)) {
      return;
    }
    try {
      setActionInProgressId(doc.id);
      await shipmentService.deleteShipmentDocument(shipmentId, doc.id);
      await fetchDocuments();
      if (onRefreshShipment) onRefreshShipment();
    } catch (err) {
      alert('Unable to delete document. Please try again.');
    } finally {
      setActionInProgressId(null);
    }
  };

  // Filtered documents
  const filteredDocs = useMemo(() => {
    return documents.filter((doc) => {
      // Category filter
      if (activeTab !== DOC_CATEGORIES.ALL) {
        const cat = doc.category || DOC_CATEGORIES.TRANSPORT;
        if (cat !== activeTab) return false;
      }
      // Status filter
      if (statusFilter !== 'ALL') {
        const st = (doc.status || '').toUpperCase();
        if (st !== statusFilter) return false;
      }
      // Search filter
      if (searchQuery.trim()) {
        const q = searchQuery.toLowerCase();
        const name = (doc.document_name || '').toLowerCase();
        const fileName = (doc.file_name || '').toLowerCase();
        const docType = (doc.doc_type || '').toLowerCase();
        const uploader = (doc.uploaded_by || '').toLowerCase();
        const refNum = (doc.reference_number || '').toLowerCase();
        return name.includes(q) || fileName.includes(q) || docType.includes(q) || uploader.includes(q) || refNum.includes(q);
      }
      return true;
    });
  }, [documents, activeTab, statusFilter, searchQuery]);

  const formatDate = (dateStr) => {
    if (!dateStr) return '—';
    try {
      const d = new Date(dateStr);
      return d.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' });
    } catch {
      return dateStr;
    }
  };

  const formatFileSize = (bytes) => {
    if (!bytes) return '—';
    if (bytes < 1024) return bytes + ' B';
    if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB';
    return (bytes / (1024 * 1024)).toFixed(1) + ' MB';
  };

  const complianceState = compliance?.compliance_state || COMPLIANCE_STATES.READY;
  const compCfg = COMPLIANCE_STATE_CONFIG[complianceState] || COMPLIANCE_STATE_CONFIG[COMPLIANCE_STATES.READY];

  const getDocStatusBadge = (status) => {
    const s = (status || DOC_STATUSES.UPLOADED).toUpperCase();
    return DOC_STATUS_CONFIG[s] || { label: s, bg: '#f1f5f9', fg: '#475569', border: '#e2e8f0' };
  };

  const getValidityBadge = (validityStatus, expiresAt) => {
    if (!expiresAt) return null;
    const v = (validityStatus || VALIDITY_STATES.VALID).toUpperCase();
    return VALIDITY_STATE_CONFIG[v] || VALIDITY_STATE_CONFIG[VALIDITY_STATES.VALID];
  };

  const missingRequirements = (compliance?.missing_documents || []).filter((r) => r.is_missing);

  // Recommended default documents for guided empty state
  const recommendedDocs = [
    { name: 'Master Bill of Lading (MBL)', cat: DOC_CATEGORIES.TRANSPORT, type: DOC_TYPES.MBL },
    { name: 'Commercial Invoice (CI)', cat: DOC_CATEGORIES.COMMERCIAL, type: DOC_TYPES.COMMERCIAL_INVOICE },
    { name: 'Packing List (PL)', cat: DOC_CATEGORIES.COMMERCIAL, type: DOC_TYPES.PACKING_LIST },
    { name: 'Customs Declaration', cat: DOC_CATEGORIES.CUSTOMS, type: DOC_TYPES.CUSTOMS_DECLARATION },
    { name: 'Cargo Insurance Certificate', cat: DOC_CATEGORIES.INSURANCE, type: DOC_TYPES.INSURANCE_CERTIFICATE },
  ];

  return (
    <div className={`sd-panel ${refreshing ? 'sd-panel-refreshing' : ''}`} id="documents-workspace">
      {refreshing && <div className="sd-refresh-shimmer-bar" />}
      {/* ── 1. WORKSPACE HEADER & COMPLIANCE HEALTH ────────────────────────── */}
      <div className="sd-panel-header">
        <div className="sd-panel-title-row">
          <div>
            <div className="sd-doc-header-title-wrap">
              <h3 className="sd-panel-title">Shipment Documents & Compliance Intelligence</h3>
              <span
                className="sd-doc-comp-badge"
                style={{ background: compCfg.bg, color: compCfg.fg, border: `1px solid ${compCfg.border}` }}
              >
                {complianceState === COMPLIANCE_STATES.COMPLIANT && <ShieldCheck size={13} />}
                {complianceState === COMPLIANCE_STATES.READY && <ShieldCheck size={13} />}
                {complianceState === COMPLIANCE_STATES.ATTENTION_REQUIRED && <AlertTriangle size={13} />}
                {complianceState === COMPLIANCE_STATES.AT_RISK && <Clock size={13} />}
                {complianceState === COMPLIANCE_STATES.BLOCKED && <ShieldAlert size={13} />}
                {compCfg.label} · {compliance?.approved || 0}/{compliance?.total_required || 4} Verified
              </span>
            </div>
            <p className="sd-panel-desc">
              Authoritative operational document register, border clearance verification, expiry risk audit, and OCR cross-checks.
            </p>
          </div>

          <div className="sd-doc-header-actions">
            <button
              className={`sd-action-btn sd-action-secondary ${justRefreshed ? 'sd-btn-refreshed' : ''}`}
              onClick={() => fetchDocuments(false)}
              disabled={refreshing || loading}
              title="Refresh document status and re-evaluate compliance"
            >
              <RotateCw size={13} className={refreshing || loading ? 'sd-spin' : ''} />
              <span>{refreshing ? 'Refreshing...' : justRefreshed ? '✓ Updated' : 'Refresh'}</span>
            </button>
            <button
              className="sd-action-btn sd-action-primary"
              onClick={() =>
                handleOpenUploadModal(
                  activeTab === DOC_CATEGORIES.ALL ? DOC_CATEGORIES.TRANSPORT : activeTab
                )
              }
            >
              <Plus size={14} /> Add Document
            </button>
          </div>
        </div>
      </div>

      {error && (
        <div className="sd-error-banner" style={{ marginBottom: '12px' }}>
          <AlertCircle size={14} />
          <span>{error}</span>
        </div>
      )}

      {/* ── 2. DOCUMENT HEALTH SUMMARY STRIP (KPI CARDS WITH CONTEXT) ──────── */}
      <div className="sd-doc-summary-grid">
        <div className="sd-doc-kpi-card">
          <div className="sd-doc-kpi-card-header">
            <span className="sd-doc-kpi-card-label">REQUIRED</span>
            <FileText size={14} style={{ color: '#64748b' }} />
          </div>
          <div className="sd-doc-kpi-card-val">{compliance?.total_required ?? 4}</div>
          <div className="sd-doc-kpi-card-sub">Mandatory for transit & closure</div>
        </div>

        <div className="sd-doc-kpi-card">
          <div className="sd-doc-kpi-card-header">
            <span className="sd-doc-kpi-card-label">READY</span>
            <FileCheck size={14} style={{ color: '#2563eb' }} />
          </div>
          <div className="sd-doc-kpi-card-val" style={{ color: '#2563eb' }}>
            {compliance?.available ?? documents.length}
          </div>
          <div className="sd-doc-kpi-card-sub">Uploaded & available</div>
        </div>

        <div className="sd-doc-kpi-card">
          <div className="sd-doc-kpi-card-header">
            <span className="sd-doc-kpi-card-label">APPROVED</span>
            <CheckCircle2 size={14} style={{ color: '#16a34a' }} />
          </div>
          <div className="sd-doc-kpi-card-val" style={{ color: '#16a34a' }}>
            {compliance?.approved ?? documents.filter((d) => d.status === DOC_STATUSES.APPROVED).length}
          </div>
          <div className="sd-doc-kpi-card-sub">Compliance verified</div>
        </div>

        <div className="sd-doc-kpi-card">
          <div className="sd-doc-kpi-card-header">
            <span className="sd-doc-kpi-card-label">UNDER REVIEW</span>
            <Clock size={14} style={{ color: '#0284c7' }} />
          </div>
          <div className="sd-doc-kpi-card-val" style={{ color: '#0284c7' }}>
            {compliance?.under_review ?? documents.filter((d) => d.status === DOC_STATUSES.UNDER_REVIEW).length}
          </div>
          <div className="sd-doc-kpi-card-sub">Awaiting operator sign-off</div>
        </div>

        <div className={`sd-doc-kpi-card ${(compliance?.missing ?? 0) > 0 ? 'sd-kpi-alert' : ''}`}>
          <div className="sd-doc-kpi-card-header">
            <span className="sd-doc-kpi-card-label">MISSING</span>
            <AlertTriangle size={14} style={{ color: (compliance?.missing ?? 0) > 0 ? '#d97706' : '#94a3b8' }} />
          </div>
          <div className="sd-doc-kpi-card-val" style={{ color: (compliance?.missing ?? 0) > 0 ? '#d97706' : '#64748b' }}>
            {compliance?.missing ?? 0}
          </div>
          <div className="sd-doc-kpi-card-sub">
            {(compliance?.missing ?? 0) > 0 ? 'Action required' : 'All requirements met'}
          </div>
        </div>

        <div
          className={`sd-doc-kpi-card ${
            (compliance?.expired ?? 0) > 0 || (compliance?.expiring_soon ?? 0) > 0 ? 'sd-kpi-warning' : ''
          }`}
        >
          <div className="sd-doc-kpi-card-header">
            <span className="sd-doc-kpi-card-label">EXPIRING / EXPIRED</span>
            <Calendar
              size={14}
              style={{
                color: (compliance?.expired ?? 0) > 0 ? '#dc2626' : (compliance?.expiring_soon ?? 0) > 0 ? '#d97706' : '#94a3b8',
              }}
            />
          </div>
          <div
            className="sd-doc-kpi-card-val"
            style={{
              color:
                (compliance?.expired ?? 0) > 0
                  ? '#dc2626'
                  : (compliance?.expiring_soon ?? 0) > 0
                    ? '#d97706'
                    : '#64748b',
            }}
          >
            {(compliance?.expiring_soon ?? 0) + (compliance?.expired ?? 0)}
          </div>
          <div className="sd-doc-kpi-card-sub">
            {(compliance?.expired ?? 0) > 0
              ? `${compliance.expired} expired`
              : (compliance?.expiring_soon ?? 0) > 0
                ? `${compliance.expiring_soon} expiring < 14d`
                : 'Validity healthy'}
          </div>
        </div>
      </div>

      {/* ── 3. OPERATIONAL ATTENTION AREA (MISSING & EXPIRING QUEUE) ────────── */}
      {missingRequirements.length > 0 ? (
        <div className="sd-doc-attention-box">
          <div className="sd-doc-attention-header">
            <div className="sd-doc-attention-title">
              <AlertTriangle size={15} />
              <span>Operational Attention: {missingRequirements.length} Required Document(s) Missing</span>
            </div>
            <span className="sd-doc-attention-sub">
              Upload missing documents to clear compliance holds and unblock destination release.
            </span>
          </div>

          <div className="sd-doc-attention-grid">
            {missingRequirements.map((req) => {
              const reqCfg = REQUIREMENT_LEVEL_CONFIG[req.requirement_level] || REQUIREMENT_LEVEL_CONFIG.REQUIRED;
              return (
                <div key={req.doc_type} className="sd-doc-attention-card">
                  <div className="sd-doc-attention-card-top">
                    <div style={{ display: 'flex', alignItems: 'center', gap: '6px' }}>
                      <span
                        className="sd-source-pill"
                        style={{ background: reqCfg.bg, color: reqCfg.fg, border: `1px solid ${reqCfg.border}` }}
                      >
                        {req.requirement_level}
                      </span>
                      <strong className="sd-doc-attention-name">{req.name}</strong>
                    </div>
                    <button
                      className="sd-action-btn sd-action-primary"
                      style={{ padding: '3px 9px', fontSize: '0.72rem' }}
                      onClick={() => handleOpenUploadModal(req.category, req.doc_type, req.name)}
                    >
                      <Plus size={11} /> Upload Document
                    </button>
                  </div>
                  <p className="sd-doc-attention-reason">{req.reason}</p>
                </div>
              );
            })}
          </div>
        </div>
      ) : (
        <div className="sd-doc-healthy-banner">
          <ShieldCheck size={16} />
          <span>
            <strong>Document Compliance Healthy:</strong> All mandatory operational and customs documents are uploaded and verified.
          </span>
        </div>
      )}

      {/* ── 4. COMPLIANCE OCR DISCREPANCIES ALERT ────────────────────────────── */}
      {discrepancies.length > 0 && (
        <div className="sd-doc-discrepancy-banner">
          <div className="sd-disc-banner-header">
            <AlertTriangle size={15} />
            <span>Automated OCR Cross-Document Discrepancies ({discrepancies.length})</span>
          </div>
          <div className="sd-disc-items-wrap">
            {discrepancies.map((disc) => (
              <div key={disc.id} className="sd-disc-row">
                <div className="sd-disc-desc">
                  <strong>{disc.field_name?.replace(/_/g, ' ')}:</strong>{' '}
                  <span className="sd-disc-vals">
                    {disc.source_document}: <code>{disc.source_value || '—'}</code> vs {disc.target_document}: <code>{disc.target_value || '—'}</code>
                  </span>
                </div>
                <span className={`sd-disc-status-pill ${disc.resolved ? 'sd-pill-resolved' : 'sd-pill-open'}`}>
                  {disc.resolved ? 'RESOLVED' : 'UNRESOLVED'}
                </span>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* ── 5. DOCUMENT TOOLBAR (TABS, SEARCH, STATUS FILTER) ───────────────── */}
      <div className="sd-doc-toolbar">
        <div className="sd-doc-tabs">
          {CATEGORY_TABS.map((tab) => {
            const count =
              tab.id === DOC_CATEGORIES.ALL
                ? documents.length
                : documents.filter((d) => (d.category || DOC_CATEGORIES.TRANSPORT) === tab.id).length;
            return (
              <button
                key={tab.id}
                className={`sd-doc-tab ${activeTab === tab.id ? 'sd-doc-tab-active' : ''}`}
                onClick={() => setActiveTab(tab.id)}
              >
                <span>{tab.label}</span>
                <span className="sd-doc-tab-count">{count}</span>
              </button>
            );
          })}
        </div>

        <div className="sd-doc-toolbar-right">
          <div className="sd-doc-search-box">
            <Search size={13} className="sd-search-icon" />
            <input
              type="text"
              placeholder="Search by name, ref, type, or uploader..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
            />
            {searchQuery && (
              <button className="sd-search-clear-btn" onClick={() => setSearchQuery('')}>
                <X size={12} />
              </button>
            )}
          </div>

          <CustomDropdown
            value={statusFilter}
            onChange={(val) => setStatusFilter(val)}
            options={[
              { value: 'ALL', label: 'All Statuses' },
              { value: DOC_STATUSES.APPROVED, label: 'Approved' },
              { value: DOC_STATUSES.UNDER_REVIEW, label: 'Under Review' },
              { value: DOC_STATUSES.UPLOADED, label: 'Uploaded' },
              { value: DOC_STATUSES.REJECTED, label: 'Rejected' },
              { value: DOC_STATUSES.EXPIRED, label: 'Expired' },
            ]}
            size="small"
          />
        </div>
      </div>

      {/* ── 6. DOCUMENT CONTENT (TABLE OR COMPACT GUIDED EMPTY STATE) ───────── */}
      {filteredDocs.length === 0 ? (
        <div className="sd-doc-compact-empty">
          <div className="sd-doc-compact-empty-header">
            <FileText size={22} className="sd-empty-icon-muted" />
            <div>
              <h4 className="sd-doc-empty-h4">No shipment documents in this view</h4>
              <p className="sd-doc-empty-p">
                Upload transport, commercial, customs, or insurance documents to maintain the operational compliance record.
              </p>
            </div>
            <button
              className="sd-action-btn sd-action-primary"
              onClick={() =>
                handleOpenUploadModal(
                  activeTab === DOC_CATEGORIES.ALL ? DOC_CATEGORIES.TRANSPORT : activeTab
                )
              }
            >
              <Plus size={13} /> Upload Document
            </button>
          </div>

          <div className="sd-doc-recommended-row">
            <span className="sd-doc-rec-title">Recommended Next Documents:</span>
            <div className="sd-doc-rec-chips">
              {recommendedDocs.map((rec) => (
                <button
                  key={rec.type}
                  className="sd-doc-rec-chip"
                  onClick={() => handleOpenUploadModal(rec.cat, rec.type, rec.name)}
                >
                  <Plus size={11} /> {rec.name}
                </button>
              ))}
            </div>
          </div>
        </div>
      ) : (
        <div className="sd-doc-table-wrap">
          <table className="sd-doc-table">
            <thead>
              <tr>
                <th>Document & Type</th>
                <th>Category</th>
                <th>Reference #</th>
                <th>Status</th>
                <th>Issue Date</th>
                <th>Expiry / Risk</th>
                <th>Source</th>
                <th>Uploaded By</th>
                <th style={{ textAlign: 'right' }}>Actions</th>
              </tr>
            </thead>
            <tbody>
              {filteredDocs.map((doc) => {
                const sBadge = getDocStatusBadge(doc.status);
                const vBadge = getValidityBadge(doc.validity_status, doc.expires_at);
                const isWorking = actionInProgressId === doc.id;

                return (
                  <tr key={doc.id} className={doc.status === DOC_STATUSES.REJECTED ? 'sd-doc-row-rejected' : ''}>
                    {/* Document Name & Type */}
                    <td>
                      <div className="sd-doc-main-cell">
                        <div className="sd-doc-icon-wrap">
                          <FileText size={16} />
                        </div>
                        <div className="sd-doc-info">
                          <span className="sd-doc-title" title={doc.document_name || doc.file_name}>
                            {doc.document_name || doc.file_name}
                          </span>
                          <div className="sd-doc-meta-row">
                            <span className="sd-doc-type-tag">{doc.doc_type?.replace(/_/g, ' ')}</span>
                            {doc.file_size && <span className="sd-doc-size-tag">{formatFileSize(doc.file_size)}</span>}
                          </div>
                          {doc.rejection_reason && (
                            <div className="sd-doc-rejection-note">
                              <strong>Rejection reason:</strong> {doc.rejection_reason}
                            </div>
                          )}
                          {doc.description && (
                            <div style={{ fontSize: '0.72rem', color: '#64748b', marginTop: '2px' }}>
                              <em>{doc.description}</em>
                            </div>
                          )}
                        </div>
                      </div>
                    </td>

                    {/* Category */}
                    <td>
                      <span className="sd-doc-type-tag" style={{ textTransform: 'capitalize' }}>
                        {doc.category?.toLowerCase().replace(/_/g, ' ')}
                      </span>
                    </td>

                    {/* Reference # */}
                    <td>
                      <span style={{ fontSize: '0.78rem', color: '#334155', fontWeight: 600 }}>
                        {doc.reference_number || '—'}
                      </span>
                    </td>

                    {/* Status */}
                    <td>
                      <span
                        className="sd-doc-status-badge"
                        style={{ background: sBadge.bg, color: sBadge.fg, border: `1px solid ${sBadge.border}` }}
                      >
                        {sBadge.label}
                      </span>
                    </td>

                    {/* Issue Date */}
                    <td>
                      <span className="sd-doc-date">{formatDate(doc.document_date)}</span>
                    </td>

                    {/* Expiry / Risk */}
                    <td>
                      {doc.expires_at ? (
                        <div style={{ display: 'flex', flexDirection: 'column', gap: '2px' }}>
                          <span className="sd-doc-date">{formatDate(doc.expires_at)}</span>
                          {vBadge && (
                            <span
                              className="sd-source-pill"
                              style={{
                                background: vBadge.bg,
                                color: vBadge.fg,
                                border: `1px solid ${vBadge.border}`,
                                width: 'fit-content',
                                padding: '1px 5px',
                                fontSize: '0.64rem',
                              }}
                            >
                              {vBadge.label}
                            </span>
                          )}
                        </div>
                      ) : (
                        <span className="sd-doc-date-muted">—</span>
                      )}
                    </td>

                    {/* Source & Lineage */}
                    <td>
                      <div className="sd-doc-source-cell">
                        {doc.source === DOC_SOURCES.RFQ ? (
                          <Link
                            to={`/dashboard/rfqs/${doc.source_id || shipment?.rfq_id}`}
                            className="sd-doc-source-link"
                            title="View original RFQ document"
                          >
                            <span className="sd-source-pill sd-src-rfq">RFQ</span>
                            <ArrowUpRight size={11} />
                          </Link>
                        ) : doc.source === DOC_SOURCES.BOOKING ? (
                          <Link
                            to={`/dashboard/bookings/${doc.source_id || shipment?.booking_id}`}
                            className="sd-doc-source-link"
                            title="View original Booking document"
                          >
                            <span className="sd-source-pill sd-src-booking">Booking</span>
                            <ArrowUpRight size={11} />
                          </Link>
                        ) : (
                          <span className="sd-source-pill sd-src-shipment">Shipment</span>
                        )}
                      </div>
                    </td>

                    {/* Uploaded By */}
                    <td>
                      <div className="sd-doc-user-stack">
                        <span className="sd-doc-user-name">{doc.uploaded_by || 'Operations Team'}</span>
                        <span className="sd-doc-date">{formatDate(doc.uploaded_at || doc.created_at)}</span>
                      </div>
                    </td>

                    {/* Actions */}
                    <td>
                      <div className="sd-doc-actions-cell">
                        {doc.status !== DOC_STATUSES.APPROVED && doc.source === DOC_SOURCES.SHIPMENT && (
                          <button
                            className="sd-table-action-btn sd-act-approve"
                            title="Approve Document"
                            disabled={isWorking}
                            onClick={() => handleApproveDoc(doc)}
                          >
                            <Check size={12} /> Approve
                          </button>
                        )}

                        {doc.status !== DOC_STATUSES.REJECTED && doc.source === DOC_SOURCES.SHIPMENT && (
                          <button
                            className="sd-table-action-btn sd-act-reject"
                            title="Reject Document with Reason"
                            disabled={isWorking}
                            onClick={() => handleOpenRejectModal(doc)}
                          >
                            <X size={12} /> Reject
                          </button>
                        )}

                        {doc.source === DOC_SOURCES.SHIPMENT && (
                          <button
                            className="sd-table-action-btn"
                            title="Edit Metadata"
                            disabled={isWorking}
                            onClick={() => handleOpenEditModal(doc)}
                          >
                            <Edit2 size={12} />
                          </button>
                        )}

                        {doc.file_url && (
                          <a
                            href={doc.file_url}
                            target="_blank"
                            rel="noopener noreferrer"
                            className="sd-table-action-btn sd-act-view"
                            title="Download / View Document"
                          >
                            <Download size={12} />
                          </a>
                        )}

                        {doc.source === DOC_SOURCES.SHIPMENT && (
                          <button
                            className="sd-table-action-btn sd-act-delete"
                            title="Delete Document"
                            disabled={isWorking}
                            onClick={() => handleDeleteDoc(doc)}
                          >
                            <Trash2 size={12} />
                          </button>
                        )}
                      </div>
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>
      )}

      {/* ── 7. UPSTREAM RFQ & BOOKING LINEAGE BAR ───────────────────────────── */}
      {(shipment?.rfq_id || shipment?.booking_id) && (
        <div className="sd-doc-lineage-footer">
          <Info size={14} className="sd-lineage-info-icon" />
          <div className="sd-lineage-text">
            <span className="sd-lineage-label">Synchronized Operational Lineage:</span>
            {shipment?.rfq_id && (
              <Link to={`/dashboard/rfqs/${shipment.rfq_id}`} className="sd-doc-lineage-nav">
                View Source RFQ ({shipment.rfq_number || `RFQ-${shipment.rfq_id}`}) →
              </Link>
            )}
            {shipment?.booking_id && (
              <Link to={`/dashboard/bookings/${shipment.booking_id}`} className="sd-doc-lineage-nav">
                View Carrier Booking ({shipment.booking_number || `BK-${shipment.booking_id}`}) →
              </Link>
            )}
          </div>
        </div>
      )}

      {/* ── 8. ADD DOCUMENT MODAL ───────────────────────────────────────────── */}
      {showUploadModal && (
        <div className="sd-modal-overlay" onClick={() => setShowUploadModal(false)}>
          <div className="sd-modal-card sd-modal-compact" onClick={(e) => e.stopPropagation()}>
            <div className="sd-modal-header">
              <div className="sd-modal-title-row">
                <FileText size={18} className="sd-modal-icon" />
                <h3>Upload Shipment Document</h3>
              </div>
              <button className="sd-modal-close" onClick={() => setShowUploadModal(false)}>
                <X size={16} />
              </button>
            </div>

            <form onSubmit={handleUploadSubmit} className="sd-modal-form">
              <div className="sd-form-grid-2">
                <div className="sd-form-field">
                  <label>Category *</label>
                  <CustomDropdown
                    value={uploadCategory}
                    onChange={handleCategoryChange}
                    options={CATEGORY_DROPDOWN_OPTIONS}
                  />
                </div>

                <div className="sd-form-field">
                  <label>Document Type *</label>
                  <CustomDropdown
                    value={uploadType}
                    onChange={(val) => setUploadType(val)}
                    options={DOC_TYPE_OPTIONS.filter((o) => o.category === uploadCategory)}
                  />
                </div>
              </div>

              <div className="sd-form-grid-2">
                <div className="sd-form-field">
                  <label>Document Title *</label>
                  <input
                    type="text"
                    placeholder="e.g. Master Bill of Lading - MAEU12345"
                    value={uploadDocName}
                    onChange={(e) => setUploadDocName(e.target.value)}
                    required
                  />
                </div>

                <div className="sd-form-field">
                  <label>Reference # / Identifier</label>
                  <input
                    type="text"
                    placeholder="e.g. MBL-98421 / INV-2026-01"
                    value={uploadRefNum}
                    onChange={(e) => setUploadRefNum(e.target.value)}
                  />
                </div>
              </div>

              <div className="sd-form-grid-2">
                <div className="sd-form-field">
                  <label>Issue / Document Date</label>
                  <input
                    type="date"
                    value={uploadDocDate}
                    onChange={(e) => setUploadDocDate(e.target.value)}
                  />
                </div>

                <div className="sd-form-field">
                  <label>Expiration Date <small>(Optional)</small></label>
                  <input
                    type="date"
                    value={uploadExpiry}
                    onChange={(e) => setUploadExpiry(e.target.value)}
                  />
                </div>
              </div>

              <div className="sd-form-field">
                <label>Operational Notes <small>(Optional)</small></label>
                <input
                  type="text"
                  placeholder="e.g. Endorsed by customs broker / stamped"
                  value={uploadDesc}
                  onChange={(e) => setUploadDesc(e.target.value)}
                />
              </div>

              <div className="sd-form-field">
                <label>Select File (PDF, PNG, JPG, CSV, XLSX) *</label>
                <div className="sd-file-dropzone">
                  <input
                    type="file"
                    id="shipment-doc-file"
                    accept=".pdf,.png,.jpg,.jpeg,.csv,.xlsx,.docx"
                    onChange={handleFileChange}
                    style={{ display: 'none' }}
                  />
                  <label htmlFor="shipment-doc-file" className="sd-dropzone-label">
                    <Upload size={20} />
                    {selectedFile ? (
                      <div>
                        <strong>{selectedFile.name}</strong>
                        <span className="sd-dropzone-sub">{formatFileSize(selectedFile.size)}</span>
                      </div>
                    ) : (
                      <div>
                        <strong>Click to select file from computer</strong>
                        <span className="sd-dropzone-sub">Supported formats: PDF, PNG, JPG, CSV, XLSX (Max 25MB)</span>
                      </div>
                    )}
                  </label>
                </div>
              </div>

              {uploadError && (
                <div className="sd-error-banner" style={{ margin: '8px 0' }}>
                  <AlertCircle size={13} />
                  <span>{uploadError}</span>
                </div>
              )}

              <div className="sd-modal-footer">
                <button
                  type="button"
                  className="sd-btn sd-btn-ghost"
                  onClick={() => setShowUploadModal(false)}
                  disabled={submittingUpload}
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  className="sd-action-btn sd-action-primary"
                  disabled={submittingUpload}
                >
                  {submittingUpload ? 'Uploading & Enqueuing...' : 'Save Document'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* ── 9. EDIT METADATA MODAL ──────────────────────────────────────────── */}
      {editingDoc && (
        <div className="sd-modal-overlay" onClick={() => setEditingDoc(null)}>
          <div className="sd-modal-card sd-modal-compact" onClick={(e) => e.stopPropagation()}>
            <div className="sd-modal-header">
              <div className="sd-modal-title-row">
                <Edit2 size={18} className="sd-modal-icon" />
                <h3>Edit Document Metadata</h3>
              </div>
              <button className="sd-modal-close" onClick={() => setEditingDoc(null)}>
                <X size={16} />
              </button>
            </div>

            <form onSubmit={handleEditSubmit} className="sd-modal-form">
              <div className="sd-form-field">
                <label>Document Title *</label>
                <input
                  type="text"
                  value={editName}
                  onChange={(e) => setEditName(e.target.value)}
                  required
                />
              </div>

              <div className="sd-form-grid-2">
                <div className="sd-form-field">
                  <label>Category *</label>
                  <CustomDropdown
                    value={editCategory}
                    onChange={(val) => setEditCategory(val)}
                    options={CATEGORY_DROPDOWN_OPTIONS}
                  />
                </div>

                <div className="sd-form-field">
                  <label>Reference # / Identifier</label>
                  <input
                    type="text"
                    placeholder="e.g. MBL-98421"
                    value={editRefNum}
                    onChange={(e) => setEditRefNum(e.target.value)}
                  />
                </div>
              </div>

              <div className="sd-form-grid-2">
                <div className="sd-form-field">
                  <label>Issue Date</label>
                  <input
                    type="date"
                    value={editDocDate}
                    onChange={(e) => setEditDocDate(e.target.value)}
                  />
                </div>

                <div className="sd-form-field">
                  <label>Expiration Date</label>
                  <input
                    type="date"
                    value={editExpiry}
                    onChange={(e) => setEditExpiry(e.target.value)}
                  />
                </div>
              </div>

              <div className="sd-form-field">
                <label>Operational Notes</label>
                <input
                  type="text"
                  placeholder="e.g. Endorsed by customs broker"
                  value={editDesc}
                  onChange={(e) => setEditDesc(e.target.value)}
                />
              </div>

              {editError && (
                <div className="sd-error-banner" style={{ margin: '8px 0' }}>
                  <AlertCircle size={13} />
                  <span>{editError}</span>
                </div>
              )}

              <div className="sd-modal-footer">
                <button
                  type="button"
                  className="sd-btn sd-btn-ghost"
                  onClick={() => setEditingDoc(null)}
                  disabled={submittingEdit}
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  className="sd-action-btn sd-action-primary"
                  disabled={submittingEdit}
                >
                  {submittingEdit ? 'Saving...' : 'Update Metadata'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* ── 10. REJECT DOCUMENT MODAL ───────────────────────────────────────── */}
      {rejectingDoc && (
        <div className="sd-modal-overlay" onClick={() => setRejectingDoc(null)}>
          <div className="sd-modal-card sd-modal-compact" onClick={(e) => e.stopPropagation()}>
            <div className="sd-modal-header">
              <div className="sd-modal-title-row">
                <ShieldAlert size={18} style={{ color: '#dc2626' }} />
                <h3>Reject Document: {rejectingDoc.document_name || rejectingDoc.file_name}</h3>
              </div>
              <button className="sd-modal-close" onClick={() => setRejectingDoc(null)}>
                <X size={16} />
              </button>
            </div>

            <form onSubmit={handleRejectSubmit} className="sd-modal-form">
              <div className="sd-form-field">
                <label>Rejection Reason & Compliance Audit Notes *</label>
                <textarea
                  rows={3}
                  required
                  placeholder="Specify discrepancies, expired validity, illegible stamps, or missing required endorsements..."
                  value={rejectReason}
                  onChange={(e) => setRejectReason(e.target.value)}
                />
              </div>

              <div className="sd-doc-reject-presets">
                <span className="sd-doc-rec-title">Quick Presets:</span>
                <div className="sd-doc-rec-chips">
                  {[
                    'Consignee details mismatch with customs entry',
                    'Illegible carrier seal / stamp endorsement',
                    'Document validity date expired',
                    'Missing certified signature / endorsement',
                  ].map((preset) => (
                    <button
                      key={preset}
                      type="button"
                      className="sd-doc-rec-chip"
                      onClick={() => setRejectReason(preset)}
                    >
                      {preset}
                    </button>
                  ))}
                </div>
              </div>

              <div className="sd-modal-footer">
                <button
                  type="button"
                  className="sd-btn sd-btn-ghost"
                  onClick={() => setRejectingDoc(null)}
                  disabled={submittingReject}
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  className="sd-action-btn sd-action-danger"
                  disabled={submittingReject}
                >
                  {submittingReject ? 'Rejecting...' : 'Confirm Document Rejection'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}

import React, { useState, useRef } from 'react';
import {
  FileText,
  Upload,
  CheckCircle,
  AlertTriangle,
  Clock,
  Eye,
  Trash2,
  Plus,
  Shield,
  FileCheck,
  XCircle,
  ExternalLink,
  Layers,
  HelpCircle,
  MoreVertical,
  ChevronDown,
  ChevronRight,
  X,
  Calendar,
  User,
  Check,
  RotateCcw
} from 'lucide-react';
import { rfqService } from '../../../../services/rfqService';

const STATUS_CONFIG = {
  APPROVED: {
    label: 'Approved',
    badgeClass: 'bg-emerald-50 text-emerald-700 border-emerald-200',
    iconClass: 'text-emerald-600 bg-emerald-100/60 border-emerald-200',
    icon: CheckCircle,
  },
  UNDER_REVIEW: {
    label: 'Under Review',
    badgeClass: 'bg-amber-50 text-amber-700 border-amber-200',
    iconClass: 'text-amber-600 bg-amber-100/60 border-amber-200',
    icon: Clock,
  },
  UPLOADED: {
    label: 'Uploaded',
    badgeClass: 'bg-blue-50 text-blue-700 border-blue-200',
    iconClass: 'text-blue-600 bg-blue-100/60 border-blue-200',
    icon: FileCheck,
  },
  REQUESTED: {
    label: 'Requested',
    badgeClass: 'bg-purple-50 text-purple-700 border-purple-200',
    iconClass: 'text-purple-600 bg-purple-100/60 border-purple-200',
    icon: HelpCircle,
  },
  MISSING: {
    label: 'Missing',
    badgeClass: 'bg-rose-50 text-rose-700 border-rose-200',
    iconClass: 'text-rose-600 bg-rose-100/60 border-rose-200',
    icon: AlertTriangle,
  },
  REJECTED: {
    label: 'Rejected',
    badgeClass: 'bg-rose-50 text-rose-700 border-rose-200',
    iconClass: 'text-rose-600 bg-rose-100/60 border-rose-200',
    icon: XCircle,
  },
  EXPIRED: {
    label: 'Expired',
    badgeClass: 'bg-orange-50 text-orange-700 border-orange-200',
    iconClass: 'text-orange-600 bg-orange-100/60 border-orange-200',
    icon: AlertTriangle,
  },
  NOT_APPLICABLE: {
    label: 'Planned (Future)',
    badgeClass: 'bg-slate-100 text-slate-600 border-slate-200',
    iconClass: 'text-slate-500 bg-slate-100 border-slate-200',
    icon: Layers,
  }
};

const DOC_TYPE_OPTIONS = [
  { value: 'COMMERCIAL_INVOICE', label: 'Commercial Invoice' },
  { value: 'PACKING_LIST', label: 'Packing List' },
  { value: 'PROFORMA_INVOICE', label: 'Proforma Invoice' },
  { value: 'PURCHASE_ORDER', label: 'Purchase Order' },
  { value: 'SALES_CONTRACT', label: 'Sales Contract' },
  { value: 'MSDS', label: 'Material Safety Data Sheet (MSDS)' },
  { value: 'DANGEROUS_GOODS_DECLARATION', label: 'Dangerous Goods Declaration' },
  { value: 'CERTIFICATE_OF_ORIGIN', label: 'Certificate of Origin' },
  { value: 'EXPORT_DECLARATION', label: 'Export Declaration' },
  { value: 'IMPORT_DECLARATION', label: 'Import Declaration' },
  { value: 'BILL_OF_LADING', label: 'Bill of Lading (OBL)' },
  { value: 'HBL', label: 'House Bill of Lading (HBL)' },
  { value: 'MBL', label: 'Master Bill of Lading (MBL)' },
  { value: 'AIR_WAYBILL', label: 'Air Waybill (HAWB/MAWB)' },
  { value: 'INSURANCE_CERTIFICATE', label: 'Insurance Certificate' },
  { value: 'BOOKING_CONFIRMATION', label: 'Booking Confirmation' },
  { value: 'OTHER', label: 'Other / Custom Document' }
];

export default function RFQDocuments({ rfq, documentsData, onMutationSuccess }) {
  const [isAddModalOpen, setIsAddModalOpen] = useState(false);
  const [isRejectModalOpen, setIsRejectModalOpen] = useState(false);
  const [selectedDocForReject, setSelectedDocForReject] = useState(null);
  const [rejectionReason, setRejectionReason] = useState('');
  const [drawerDoc, setDrawerDoc] = useState(null);
  const [actionLoading, setActionLoading] = useState(false);
  const [errorMsg, setErrorMsg] = useState(null);

  // Form State
  const [docType, setDocType] = useState('COMMERCIAL_INVOICE');
  const [docName, setDocName] = useState('');
  const [docDesc, setDocDesc] = useState('');
  const [selectedFile, setSelectedFile] = useState(null);
  const [expiryDate, setExpiryDate] = useState('');
  const [showAdvancedFields, setShowAdvancedFields] = useState(false);
  const fileInputRef = useRef(null);

  const summary = documentsData?.summary || {
    total_documents: 0,
    required_documents: 2,
    received_documents: 0,
    missing_documents: 2,
    under_review_documents: 0,
    approved_documents: 0,
    rejected_documents: 0,
    future_stage_documents: 4,
    readiness_percentage: 0
  };

  const currentStage = documentsData?.current_stage_documents || [];
  const conditional = documentsData?.conditional_documents || [];
  const futureStage = documentsData?.future_stage_documents || [];

  const handleOpenAddModal = (prefillType = 'COMMERCIAL_INVOICE', prefillName = '') => {
    setDocType(prefillType);
    const matched = DOC_TYPE_OPTIONS.find(o => o.value === prefillType);
    setDocName(prefillName || (matched ? matched.label : ''));
    setDocDesc('');
    setSelectedFile(null);
    setExpiryDate('');
    setShowAdvancedFields(false);
    setErrorMsg(null);
    setIsAddModalOpen(true);
  };

  const handleFileChange = (e) => {
    if (e.target.files && e.target.files[0]) {
      const file = e.target.files[0];
      setSelectedFile(file);
      if (!docName) {
        setDocName(file.name.replace(/\.[^/.]+$/, ""));
      }
    }
  };

  const handleCreateDocument = async (e) => {
    e.preventDefault();
    if (!docName.trim()) {
      setErrorMsg('Please specify a document name.');
      return;
    }
    try {
      setActionLoading(true);
      setErrorMsg(null);

      let finalFileName = selectedFile ? selectedFile.name : `${docType.toLowerCase()}_${Date.now()}.pdf`;
      let finalFileSize = selectedFile ? selectedFile.size : 1024 * 350; // fallback standard size
      let finalMimeType = selectedFile ? selectedFile.type : 'application/pdf';
      let finalFileUrl = `http://localhost:8080/uploads/${finalFileName}`;

      const payload = {
        document_type: docType,
        document_name: docName.trim(),
        description: docDesc.trim() || undefined,
        file_name: finalFileName,
        file_url: finalFileUrl,
        file_size: finalFileSize,
        mime_type: finalMimeType,
        expires_at: expiryDate ? new Date(expiryDate).toISOString() : undefined
      };

      await rfqService.createDocument(rfq.id, payload);
      setIsAddModalOpen(false);
      if (onMutationSuccess) onMutationSuccess();
    } catch (err) {
      console.error('Failed to create document:', err);
      setErrorMsg(err.response?.data?.error?.message || err.message || 'Failed to create document.');
    } finally {
      setActionLoading(false);
    }
  };

  const handleUpdateStatus = async (documentId, newStatus, reason = null) => {
    if (!documentId) return;
    try {
      setActionLoading(true);
      setErrorMsg(null);
      await rfqService.updateDocumentStatus(rfq.id, documentId, {
        status: newStatus,
        rejection_reason: reason || undefined
      });
      setIsRejectModalOpen(false);
      setSelectedDocForReject(null);
      setRejectionReason('');
      if (drawerDoc && drawerDoc.id === documentId) {
        setDrawerDoc(prev => ({ ...prev, status: newStatus, rejection_reason: reason || prev?.rejection_reason }));
      }
      if (onMutationSuccess) onMutationSuccess();
    } catch (err) {
      console.error('Failed to update document status:', err);
      setErrorMsg(err.response?.data?.error?.message || err.message || 'Status transition failed.');
    } finally {
      setActionLoading(false);
    }
  };

  const handleDeleteDocument = async (documentId) => {
    if (!window.confirm('Are you sure you want to delete this document record?')) return;
    try {
      setActionLoading(true);
      await rfqService.deleteDocument(rfq.id, documentId);
      if (drawerDoc && drawerDoc.id === documentId) {
        setDrawerDoc(null);
      }
      if (onMutationSuccess) onMutationSuccess();
    } catch (err) {
      console.error('Failed to delete document:', err);
      setErrorMsg(err.response?.data?.error?.message || err.message || 'Failed to delete document.');
    } finally {
      setActionLoading(false);
    }
  };

  const openRejectDialog = (doc) => {
    setSelectedDocForReject(doc);
    setRejectionReason('');
    setIsRejectModalOpen(true);
  };

  if (documentsData === null) {
    return (
      <div className="bg-white border border-slate-200/80 rounded-xl p-12 text-center shadow-xs" data-testid="rfq-documents-loading">
        <div className="w-12 h-12 bg-indigo-50 border border-indigo-100 rounded-2xl flex items-center justify-center mx-auto mb-3.5 text-indigo-600">
          <RefreshCw className="w-6 h-6 animate-spin text-indigo-600" />
        </div>
        <h3 className="text-sm font-bold text-slate-900 mb-1">Loading Documents & Compliance...</h3>
        <p className="text-xs text-slate-500 max-w-md mx-auto leading-relaxed">
          Synchronizing shipping documentation, verification checklists, and compliance requirements from the backend.
        </p>
      </div>
    );
  }

  return (
    <div className="space-y-5 animate-fadeIn" data-testid="rfq-documents-workspace">
      
      {/* ── 1. Clean SaaS Header & Inline Readiness Indicator ── */}
      <div className="bg-white border border-slate-200 rounded-xl p-5 shadow-xs">
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
          <div>
            <div className="flex items-center gap-2">
              <div className="w-8 h-8 rounded-lg bg-indigo-50 border border-indigo-100 flex items-center justify-center text-indigo-600">
                <FileText className="w-4 h-4" />
              </div>
              <h2 className="text-lg font-bold text-slate-900 tracking-tight">
                Document Management
              </h2>
            </div>
            <p className="text-xs text-slate-500 mt-1 pl-10">
              Track required, conditional, and future-stage documentation for this RFQ.
            </p>
          </div>

          <div className="flex items-center gap-3">
            <button
              onClick={() => handleOpenAddModal()}
              className="inline-flex items-center gap-2 px-4 py-2 bg-indigo-600 hover:bg-indigo-700 text-white text-xs font-semibold rounded-lg shadow-xs transition-all active:scale-95"
              data-testid="add-document-btn"
            >
              <Plus className="w-4 h-4" />
              Add / Attach Document
            </button>
          </div>
        </div>

        {/* Inline Readiness Bar */}
        <div className="mt-4 pt-4 border-t border-slate-100">
          <div className="flex items-center justify-between text-xs mb-1.5">
            <div className="flex items-center gap-2">
              <span className="font-semibold text-slate-700">Document Readiness:</span>
              <span className={`font-bold px-1.5 py-0.5 rounded text-[11px] ${summary.readiness_percentage === 100 ? 'bg-emerald-50 text-emerald-700 border border-emerald-200' : 'bg-amber-50 text-amber-700 border border-amber-200'}`}>
                {summary.readiness_percentage}% Complete
              </span>
            </div>
            <span className="text-slate-500 font-medium">
              {summary.approved_documents} of {summary.required_documents} required documents approved
            </span>
          </div>
          <div className="w-full bg-slate-100 rounded-full h-2 overflow-hidden">
            <div
              className={`h-2 rounded-full transition-all duration-500 ${summary.readiness_percentage === 100 ? 'bg-emerald-500' : 'bg-gradient-to-r from-indigo-500 to-emerald-500'}`}
              style={{ width: `${summary.readiness_percentage}%` }}
            />
          </div>
        </div>
      </div>

      {errorMsg && (
        <div className="p-3.5 bg-rose-50 border border-rose-200 rounded-xl text-rose-700 text-xs flex items-center justify-between">
          <div className="flex items-center gap-2">
            <AlertTriangle className="w-4 h-4 text-rose-600 flex-shrink-0" />
            <span>{errorMsg}</span>
          </div>
          <button onClick={() => setErrorMsg(null)} className="text-xs font-semibold text-rose-800 hover:underline">
            Dismiss
          </button>
        </div>
      )}

      {/* ── 2. Unified Compact Horizontal KPI Strip ── */}
      <div className="rfq-kpi-strip">
        <div className="rfq-kpi-segment">
          <div className="rfq-kpi-label">
            <span>Readiness Score</span>
            <span className="w-2 h-2 rounded-full bg-indigo-500"></span>
          </div>
          <div className="rfq-kpi-value text-indigo-600">
            {summary.readiness_percentage}%
          </div>
          <div className="rfq-kpi-subtext">Quotation milestone</div>
        </div>

        <div className="rfq-kpi-segment">
          <div className="rfq-kpi-label">
            <span>Required Documents</span>
            <CheckCircle className="w-3.5 h-3.5 text-emerald-500" />
          </div>
          <div className="rfq-kpi-value">
            {summary.approved_documents} <span className="text-xs font-normal text-slate-400">/ {summary.required_documents} Approved</span>
          </div>
          <div className="rfq-kpi-subtext">Current RFQ stage</div>
        </div>

        <div className="rfq-kpi-segment">
          <div className="rfq-kpi-label">
            <span>Need Attention</span>
            <AlertTriangle className="w-3.5 h-3.5 text-rose-500" />
          </div>
          <div className="rfq-kpi-value text-rose-600">
            {summary.missing_documents}
          </div>
          <div className="rfq-kpi-subtext">Missing before quote</div>
        </div>

        <div className="rfq-kpi-segment">
          <div className="rfq-kpi-label">
            <span>Under Review</span>
            <Clock className="w-3.5 h-3.5 text-amber-500" />
          </div>
          <div className="rfq-kpi-value text-amber-600">
            {summary.under_review_documents}
          </div>
          <div className="rfq-kpi-subtext">Awaiting team verification</div>
        </div>

        <div className="rfq-kpi-segment">
          <div className="rfq-kpi-label">
            <span>Future Stage</span>
            <Layers className="w-3.5 h-3.5 text-slate-400" />
          </div>
          <div className="rfq-kpi-value text-slate-700">
            {summary.future_stage_documents}
          </div>
          <div className="rfq-kpi-subtext">Tracked for booking</div>
        </div>
      </div>

      {/* ── 3. Current RFQ Stage Documents ── */}
      <div className="bg-white border border-slate-200 rounded-xl overflow-hidden shadow-xs">
        <div className="px-5 py-3.5 bg-slate-50/80 border-b border-slate-200 flex items-center justify-between">
          <div className="flex items-center gap-2">
            <span className="w-2 h-2 rounded-full bg-indigo-600"></span>
            <h3 className="text-xs font-bold text-slate-800 uppercase tracking-wider">
              Current Stage Documents ({currentStage.length})
            </h3>
          </div>
          <span className="text-xs text-slate-500 font-medium">
            Required before quotation dispatch
          </span>
        </div>

        <div className="divide-y divide-slate-100">
          {currentStage.length === 0 ? (
            <div className="p-8 text-center text-xs text-slate-500">No current stage documents configured.</div>
          ) : (
            currentStage.map((item, idx) => (
              <DocumentRow
                key={item.doc_type || idx}
                item={item}
                onUpload={() => handleOpenAddModal(item.doc_type, item.title)}
                onUpdateStatus={handleUpdateStatus}
                onReject={() => openRejectDialog(item.document_record)}
                onDelete={handleDeleteDocument}
                onOpenDrawer={setDrawerDoc}
              />
            ))
          )}
        </div>
      </div>

      {/* ── 4. Conditional & Compliance Documents ── */}
      {conditional.length > 0 && (
        <div className="bg-white border border-amber-200/90 rounded-xl overflow-hidden shadow-xs">
          <div className="px-5 py-3.5 bg-gradient-to-r from-amber-50/80 to-amber-50/30 border-b border-amber-200 flex items-center justify-between">
            <div className="flex items-center gap-2.5">
              <div className="w-7 h-7 rounded-lg bg-amber-100/80 border border-amber-200 flex items-center justify-center text-amber-700">
                <Shield className="w-3.5 h-3.5" />
              </div>
              <div>
                <h3 className="text-xs font-bold text-amber-950 uppercase tracking-wider">
                  Conditional Compliance Requirements ({conditional.length})
                </h3>
                <p className="text-[11px] text-amber-800 font-medium">
                  Triggered dynamically by commodity characteristics or trade lane regulations
                </p>
              </div>
            </div>
            <span className="px-2 py-0.5 rounded-full bg-amber-100/80 text-amber-800 text-[11px] font-bold border border-amber-200">
              Rule-Based
            </span>
          </div>

          <div className="divide-y divide-amber-100/70">
            {conditional.map((item, idx) => (
              <DocumentRow
                key={item.doc_type || idx}
                item={item}
                isConditionalSection
                onUpload={() => handleOpenAddModal(item.doc_type, item.title)}
                onUpdateStatus={handleUpdateStatus}
                onReject={() => openRejectDialog(item.document_record)}
                onDelete={handleDeleteDocument}
                onOpenDrawer={setDrawerDoc}
              />
            ))}
          </div>
        </div>
      )}

      {/* ── 5. Future Stage Documents (Tracked & Non-blocking) ── */}
      <div className="bg-white border border-slate-200 rounded-xl overflow-hidden shadow-xs">
        <div className="px-5 py-3.5 bg-slate-50/90 border-b border-slate-200 flex flex-col sm:flex-row sm:items-center justify-between gap-2">
          <div className="flex items-center gap-2.5">
            <div className="w-7 h-7 rounded-lg bg-slate-100 border border-slate-200 flex items-center justify-center text-slate-600">
              <Layers className="w-3.5 h-3.5" />
            </div>
            <div>
              <div className="flex items-center gap-2">
                <h3 className="text-xs font-bold text-slate-800 uppercase tracking-wider">
                  Future Stage Documents ({futureStage.length} Tracked)
                </h3>
                <span className="px-2 py-0.2 rounded-full bg-slate-100 text-slate-600 text-[10px] font-bold border border-slate-200">
                  Non-Blocking for Quote
                </span>
              </div>
              <p className="text-[11px] text-slate-500 font-medium">
                These documents become mandatory at booking or shipment execution stages.
              </p>
            </div>
          </div>
          <span className="text-[11px] text-slate-400">
            Lifecycle: Booking → Customs → Delivery
          </span>
        </div>

        {/* Clean Operational Table */}
        <div className="overflow-x-auto">
          <table className="rfq-table">
            <thead>
              <tr>
                <th style={{ minWidth: '220px' }}>Document Name</th>
                <th style={{ minWidth: '160px' }}>Required Lifecycle Stage</th>
                <th style={{ minWidth: '280px' }}>Trigger / Purpose</th>
                <th style={{ minWidth: '120px' }}>Readiness State</th>
                <th style={{ textAlign: 'right', minWidth: '120px' }}>Action</th>
              </tr>
            </thead>
            <tbody>
              {futureStage.map((item, idx) => {
                const doc = item.document_record;
                const statusKey = doc ? doc.status : (item.status || 'NOT_APPLICABLE');
                const config = STATUS_CONFIG[statusKey] || STATUS_CONFIG.NOT_APPLICABLE;
                const stageLabel = item.applicable_stage === 'BOOKING_CONFIRMED'
                  ? 'Booking Confirmation'
                  : item.applicable_stage === 'SHIPMENT_EXECUTION'
                  ? 'Shipment Execution'
                  : 'Customs Clearance';

                const stageBadgeClass = item.applicable_stage === 'BOOKING_CONFIRMED'
                  ? 'bg-blue-50 text-blue-700 border-blue-200'
                  : item.applicable_stage === 'SHIPMENT_EXECUTION'
                  ? 'bg-indigo-50 text-indigo-700 border-indigo-200'
                  : 'bg-purple-50 text-purple-700 border-purple-200';

                return (
                  <tr key={item.doc_type || idx} className="hover:bg-slate-50/80 transition-colors">
                    <td>
                      <div className="flex items-center gap-2.5">
                        <div className="w-7 h-7 rounded-lg bg-slate-100 border border-slate-200 flex items-center justify-center text-slate-500 flex-shrink-0">
                          <FileText className="w-3.5 h-3.5" />
                        </div>
                        <div>
                          <div className="font-bold text-slate-900 text-xs">
                            {item.title}
                          </div>
                          {doc?.file_name ? (
                            <div className="text-[11px] text-indigo-600 font-medium truncate max-w-[180px]">
                              📎 {doc.file_name}
                            </div>
                          ) : (
                            <div className="text-[10px] text-slate-400">
                              No file attached yet
                            </div>
                          )}
                        </div>
                      </div>
                    </td>
                    <td>
                      <span className={`inline-flex items-center gap-1 px-2.5 py-0.5 rounded-md text-[11px] font-bold border ${stageBadgeClass}`}>
                        {stageLabel}
                      </span>
                    </td>
                    <td>
                      <span className="text-xs text-slate-600 leading-snug">
                        {item.reason || 'Standard operational document for forwarder & carrier issuance.'}
                      </span>
                    </td>
                    <td>
                      <span className={`inline-flex items-center gap-1 px-2.5 py-1 rounded-full text-[11px] font-bold border ${config.badgeClass}`}>
                        {config.label}
                      </span>
                    </td>
                    <td style={{ textAlign: 'right' }}>
                      {!doc ? (
                        <button
                          onClick={() => handleOpenAddModal(item.doc_type, item.title)}
                          className="px-3 py-1.5 text-xs font-bold text-indigo-600 hover:text-indigo-700 bg-indigo-50/80 hover:bg-indigo-100 border border-indigo-200 rounded-lg transition-colors inline-flex items-center gap-1 shadow-2xs"
                        >
                          <Plus className="w-3 h-3" /> Upload Early
                        </button>
                      ) : (
                        <button
                          onClick={() => setDrawerDoc(doc)}
                          className="px-3 py-1.5 text-xs font-bold text-slate-700 hover:text-slate-900 bg-slate-100 hover:bg-slate-200 border border-slate-200 rounded-lg transition-colors inline-flex items-center gap-1"
                        >
                          <Eye className="w-3 h-3 text-slate-500" /> View Details
                        </button>
                      )}
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>
      </div>


      {/* ── 6. Add / Attach Document Modal (Structured 3-Step) ── */}
      {isAddModalOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-900/50 backdrop-blur-xs animate-fadeIn">
          <div className="bg-white border border-slate-200 rounded-2xl w-full max-w-lg overflow-hidden shadow-2xl">
            <div className="px-6 py-4 border-b border-slate-100 flex items-center justify-between bg-slate-50/50">
              <div className="flex items-center gap-2.5">
                <div className="w-8 h-8 rounded-lg bg-indigo-50 border border-indigo-100 flex items-center justify-center text-indigo-600">
                  <Upload className="w-4 h-4" />
                </div>
                <h3 className="text-base font-bold text-slate-900">
                  Attach / Upload Document
                </h3>
              </div>
              <button
                onClick={() => setIsAddModalOpen(false)}
                className="text-slate-400 hover:text-slate-600 transition-colors p-1"
              >
                <X className="w-5 h-5" />
              </button>
            </div>

            <form onSubmit={handleCreateDocument} className="p-6 space-y-4">
              {/* Step 1: Document Information */}
              <div>
                <label className="block text-xs font-bold text-slate-700 uppercase tracking-wider mb-1.5">
                  Document Type *
                </label>
                <select
                  value={docType}
                  onChange={(e) => {
                    setDocType(e.target.value);
                    const opt = DOC_TYPE_OPTIONS.find(o => o.value === e.target.value);
                    if (opt) setDocName(opt.label);
                  }}
                  className="w-full bg-white border border-slate-300 rounded-lg px-3.5 py-2 text-sm text-slate-900 focus:outline-none focus:ring-2 focus:ring-indigo-500/20 focus:border-indigo-600 font-medium"
                >
                  {DOC_TYPE_OPTIONS.map((opt) => (
                    <option key={opt.value} value={opt.value}>
                      {opt.label}
                    </option>
                  ))}
                </select>
              </div>

              <div>
                <label className="block text-xs font-bold text-slate-700 uppercase tracking-wider mb-1.5">
                  Document Display Name *
                </label>
                <input
                  type="text"
                  value={docName}
                  onChange={(e) => setDocName(e.target.value)}
                  placeholder="e.g. Commercial Invoice #INV-2026-992"
                  required
                  className="w-full bg-white border border-slate-300 rounded-lg px-3.5 py-2 text-sm text-slate-900 placeholder-slate-400 focus:outline-none focus:ring-2 focus:ring-indigo-500/20 focus:border-indigo-600"
                />
              </div>

              {/* Step 2: File Attachment Zone */}
              <div>
                <label className="block text-xs font-bold text-slate-700 uppercase tracking-wider mb-1.5">
                  Attach File
                </label>
                
                <input
                  ref={fileInputRef}
                  type="file"
                  onChange={handleFileChange}
                  className="hidden"
                  accept=".pdf,.png,.jpg,.jpeg,.xlsx,.docx"
                />

                {!selectedFile ? (
                  <div
                    onClick={() => fileInputRef.current?.click()}
                    className="border-2 border-dashed border-slate-300 hover:border-indigo-500 rounded-xl p-5 text-center cursor-pointer transition-colors bg-slate-50/50 hover:bg-indigo-50/20"
                  >
                    <Upload className="w-6 h-6 text-slate-400 mx-auto mb-2" />
                    <div className="text-xs font-semibold text-slate-800">
                      Click to browse or drag file here
                    </div>
                    <p className="text-[11px] text-slate-400 mt-1">
                      PDF, XLSX, DOCX, PNG up to 25 MB
                    </p>
                  </div>
                ) : (
                  <div className="flex items-center justify-between p-3 bg-indigo-50/50 border border-indigo-200 rounded-xl">
                    <div className="flex items-center gap-2.5 truncate">
                      <FileText className="w-5 h-5 text-indigo-600 flex-shrink-0" />
                      <div className="truncate">
                        <div className="text-xs font-bold text-slate-900 truncate">{selectedFile.name}</div>
                        <div className="text-[10px] text-slate-500">{(selectedFile.size / 1024).toFixed(1)} KB</div>
                      </div>
                    </div>
                    <button
                      type="button"
                      onClick={() => setSelectedFile(null)}
                      className="text-xs font-semibold text-rose-600 hover:text-rose-700 px-2 py-1"
                    >
                      Remove
                    </button>
                  </div>
                )}
              </div>

              {/* Step 3: Expandable Additional Details */}
              <div className="pt-2">
                <button
                  type="button"
                  onClick={() => setShowAdvancedFields(!showAdvancedFields)}
                  className="flex items-center gap-1.5 text-xs font-bold text-slate-600 hover:text-slate-900"
                >
                  {showAdvancedFields ? <ChevronDown className="w-3.5 h-3.5" /> : <ChevronRight className="w-3.5 h-3.5" />}
                  <span>Additional Details (Expiry date, operational notes)</span>
                </button>

                {showAdvancedFields && (
                  <div className="mt-3 space-y-3 pl-4 border-l-2 border-slate-200">
                    <div>
                      <label className="block text-[11px] font-bold text-slate-600 uppercase tracking-wider mb-1">
                        Expiry Date
                      </label>
                      <input
                        type="date"
                        value={expiryDate}
                        onChange={(e) => setExpiryDate(e.target.value)}
                        className="w-full bg-white border border-slate-300 rounded-lg px-3 py-1.5 text-xs text-slate-900 focus:outline-none focus:border-indigo-600"
                      />
                    </div>
                    <div>
                      <label className="block text-[11px] font-bold text-slate-600 uppercase tracking-wider mb-1">
                        Operational Notes
                      </label>
                      <textarea
                        value={docDesc}
                        onChange={(e) => setDocDesc(e.target.value)}
                        rows={2}
                        placeholder="Context for customs or carrier compliance..."
                        className="w-full bg-white border border-slate-300 rounded-lg px-3 py-1.5 text-xs text-slate-900 placeholder-slate-400 focus:outline-none focus:border-indigo-600"
                      />
                    </div>
                  </div>
                )}
              </div>

              {/* Action Buttons */}
              <div className="pt-4 border-t border-slate-100 flex justify-end gap-2.5">
                <button
                  type="button"
                  onClick={() => setIsAddModalOpen(false)}
                  className="px-4 py-2 bg-slate-100 hover:bg-slate-200 text-slate-700 text-xs font-semibold rounded-lg transition-colors"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  disabled={actionLoading}
                  className="px-4 py-2 bg-indigo-600 hover:bg-indigo-700 text-white text-xs font-semibold rounded-lg shadow-xs transition-all disabled:opacity-50"
                  data-testid="submit-add-doc-btn"
                >
                  {actionLoading ? 'Attaching...' : 'Save & Attach Document'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* ── 7. Reject Document Modal ── */}
      {isRejectModalOpen && selectedDocForReject && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-900/50 backdrop-blur-xs animate-fadeIn">
          <div className="bg-white border border-slate-200 rounded-2xl w-full max-w-md overflow-hidden shadow-2xl">
            <div className="px-6 py-4 border-b border-rose-100 bg-rose-50/50 flex items-center justify-between">
              <div className="flex items-center gap-2">
                <XCircle className="w-5 h-5 text-rose-600" />
                <h3 className="text-sm font-bold text-rose-900">
                  Reject Document Record
                </h3>
              </div>
              <button
                onClick={() => setIsRejectModalOpen(false)}
                className="text-slate-400 hover:text-slate-600 transition-colors"
              >
                <X className="w-5 h-5" />
              </button>
            </div>

            <div className="p-6 space-y-4">
              <p className="text-xs text-slate-600 leading-relaxed">
                Rejecting <span className="font-bold text-slate-900">{selectedDocForReject.document_name}</span>. Provide the operational reason for audit trail:
              </p>

              {/* Quick reason chips */}
              <div className="flex flex-wrap gap-1.5">
                {[
                  'Missing signature / stamp',
                  'Illegible scan',
                  'Incorrect consignee details',
                  'Expired certificate',
                  'Commodity mismatch'
                ].map((chip) => (
                  <button
                    key={chip}
                    type="button"
                    onClick={() => setRejectionReason(chip)}
                    className="px-2 py-0.5 bg-slate-100 hover:bg-slate-200 text-slate-700 text-[11px] rounded-md font-medium transition-colors"
                  >
                    {chip}
                  </button>
                ))}
              </div>

              <textarea
                value={rejectionReason}
                onChange={(e) => setRejectionReason(e.target.value)}
                placeholder="Explain what needs correction..."
                rows={3}
                className="w-full bg-white border border-slate-300 rounded-lg p-3 text-xs text-slate-900 placeholder-slate-400 focus:outline-none focus:border-rose-500"
              />

              <div className="pt-3 border-t border-slate-100 flex justify-end gap-2.5">
                <button
                  type="button"
                  onClick={() => setIsRejectModalOpen(false)}
                  className="px-4 py-2 bg-slate-100 hover:bg-slate-200 text-slate-700 text-xs font-semibold rounded-lg transition-colors"
                >
                  Cancel
                </button>
                <button
                  type="button"
                  disabled={actionLoading || !rejectionReason.trim()}
                  onClick={() => handleUpdateStatus(selectedDocForReject.id, 'REJECTED', rejectionReason)}
                  className="px-4 py-2 bg-rose-600 hover:bg-rose-700 text-white text-xs font-semibold rounded-lg shadow-xs transition-all disabled:opacity-50"
                  data-testid="confirm-reject-doc-btn"
                >
                  {actionLoading ? 'Rejecting...' : 'Confirm Rejection'}
                </button>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* ── 8. Right-Side Slide-Over Document Detail Drawer ── */}
      {drawerDoc && (
        <div className="rfq-drawer-overlay" onClick={() => setDrawerDoc(null)}>
          <div className="rfq-drawer-panel p-6" onClick={(e) => e.stopPropagation()}>
            {/* Drawer Header */}
            <div className="flex items-center justify-between pb-4 border-b border-slate-200">
              <div className="flex items-center gap-2">
                <FileText className="w-5 h-5 text-indigo-600" />
                <h3 className="text-base font-bold text-slate-900 truncate max-w-[320px]">
                  {drawerDoc.document_name}
                </h3>
              </div>
              <button
                onClick={() => setDrawerDoc(null)}
                className="text-slate-400 hover:text-slate-700 p-1"
              >
                <X className="w-5 h-5" />
              </button>
            </div>

            {/* Status Pill */}
            <div className="py-4 border-b border-slate-100 flex items-center justify-between">
              <span className="text-xs font-bold text-slate-500 uppercase tracking-wider">Status</span>
              <span className={`inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-bold border ${STATUS_CONFIG[drawerDoc.status]?.badgeClass || 'bg-slate-100 text-slate-700'}`}>
                {drawerDoc.status}
              </span>
            </div>

            {/* File Info Card */}
            <div className="my-4 p-3.5 bg-slate-50 border border-slate-200 rounded-xl space-y-2">
              <div className="text-xs font-bold text-slate-700 uppercase tracking-wider">Attached File</div>
              <div className="flex items-center justify-between">
                <span className="text-xs font-semibold text-slate-900 truncate max-w-[260px]">
                  📄 {drawerDoc.file_name || 'Attached file'}
                </span>
                {drawerDoc.file_url && (
                  <a
                    href={drawerDoc.file_url}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="inline-flex items-center gap-1 text-xs font-bold text-indigo-600 hover:text-indigo-700"
                  >
                    Open <ExternalLink className="w-3 h-3" />
                  </a>
                )}
              </div>
              {drawerDoc.file_size && (
                <div className="text-[11px] text-slate-400">
                  Size: {(drawerDoc.file_size / 1024).toFixed(1)} KB • Type: {drawerDoc.mime_type || 'application/pdf'}
                </div>
              )}
            </div>

            {/* Rejection notice if present */}
            {drawerDoc.rejection_reason && (
              <div className="p-3 bg-rose-50 border border-rose-200 rounded-xl mb-4 text-xs">
                <span className="font-bold text-rose-800 block mb-0.5">Rejection Reason:</span>
                <p className="text-rose-700">{drawerDoc.rejection_reason}</p>
              </div>
            )}

            {/* Operational Metadata */}
            <div className="space-y-3 py-2 text-xs border-b border-slate-100 pb-4">
              <div className="flex justify-between py-1">
                <span className="text-slate-500">Document Type</span>
                <span className="font-semibold text-slate-800">{drawerDoc.document_type}</span>
              </div>
              <div className="flex justify-between py-1">
                <span className="text-slate-500">Uploaded By</span>
                <span className="font-medium text-slate-700">{drawerDoc.uploaded_by || 'Operations Team'}</span>
              </div>
              <div className="flex justify-between py-1">
                <span className="text-slate-500">Uploaded At</span>
                <span className="font-medium text-slate-700">
                  {drawerDoc.uploaded_at ? new Date(drawerDoc.uploaded_at).toLocaleString() : 'N/A'}
                </span>
              </div>
              {drawerDoc.reviewed_by && (
                <div className="flex justify-between py-1">
                  <span className="text-slate-500">Reviewed By</span>
                  <span className="font-medium text-slate-700">{drawerDoc.reviewed_by}</span>
                </div>
              )}
              {drawerDoc.reviewed_at && (
                <div className="flex justify-between py-1">
                  <span className="text-slate-500">Reviewed At</span>
                  <span className="font-medium text-slate-700">
                    {new Date(drawerDoc.reviewed_at).toLocaleString()}
                  </span>
                </div>
              )}
              {drawerDoc.expires_at && (
                <div className="flex justify-between py-1">
                  <span className="text-slate-500">Expires At</span>
                  <span className="font-medium text-slate-700">
                    {new Date(drawerDoc.expires_at).toLocaleDateString()}
                  </span>
                </div>
              )}
              {drawerDoc.description && (
                <div className="pt-2">
                  <span className="text-slate-500 block mb-1">Description</span>
                  <p className="text-slate-700 bg-slate-50 p-2.5 rounded-lg border border-slate-200">
                    {drawerDoc.description}
                  </p>
                </div>
              )}
            </div>

            {/* Direct Drawer Actions */}
            <div className="mt-auto pt-4 flex flex-col gap-2">
              {drawerDoc.status === 'UPLOADED' && (
                <div className="grid grid-cols-2 gap-2">
                  <button
                    onClick={() => handleUpdateStatus(drawerDoc.id, 'APPROVED')}
                    className="py-2 px-3 bg-emerald-600 hover:bg-emerald-700 text-white font-bold text-xs rounded-lg shadow-xs transition-colors"
                  >
                    ✓ Approve Document
                  </button>
                  <button
                    onClick={() => openRejectDialog(drawerDoc)}
                    className="py-2 px-3 bg-rose-50 hover:bg-rose-100 text-rose-700 font-bold text-xs rounded-lg border border-rose-200 transition-colors"
                  >
                    ✕ Reject Document
                  </button>
                </div>
              )}

              {drawerDoc.status === 'UNDER_REVIEW' && (
                <div className="grid grid-cols-2 gap-2">
                  <button
                    onClick={() => handleUpdateStatus(drawerDoc.id, 'APPROVED')}
                    className="py-2 px-3 bg-emerald-600 hover:bg-emerald-700 text-white font-bold text-xs rounded-lg shadow-xs transition-colors"
                  >
                    ✓ Approve
                  </button>
                  <button
                    onClick={() => openRejectDialog(drawerDoc)}
                    className="py-2 px-3 bg-rose-50 hover:bg-rose-100 text-rose-700 font-bold text-xs rounded-lg border border-rose-200 transition-colors"
                  >
                    ✕ Reject
                  </button>
                </div>
              )}

              <div className="flex justify-between items-center pt-2">
                <button
                  onClick={() => {
                    handleOpenAddModal(drawerDoc.document_type, drawerDoc.document_name);
                  }}
                  className="text-xs font-semibold text-indigo-600 hover:text-indigo-700"
                >
                  Replace File
                </button>
                <button
                  onClick={() => handleDeleteDocument(drawerDoc.id)}
                  className="text-xs font-semibold text-rose-600 hover:text-rose-700"
                >
                  Delete Record
                </button>
              </div>
            </div>
          </div>
        </div>
      )}

    </div>
  );
}

// ── Reusable Operational Document Row Component ──
function DocumentRow({ item, isConditionalSection = false, onUpload, onUpdateStatus, onReject, onDelete, onOpenDrawer }) {
  const [menuOpen, setMenuOpen] = useState(false);
  const doc = item.document_record;
  const statusKey = doc ? doc.status : (item.status || 'MISSING');
  const config = STATUS_CONFIG[statusKey] || STATUS_CONFIG.MISSING;
  const StatusIcon = config.icon;

  return (
    <div className="px-5 py-4 hover:bg-slate-50/60 transition-colors flex flex-col md:flex-row md:items-center justify-between gap-4">
      
      {/* Left: Status Icon & Document Hierarchy */}
      <div className="flex items-start gap-3 min-w-0">
        <div className={`w-9 h-9 rounded-xl border flex items-center justify-center flex-shrink-0 mt-0.5 ${config.iconClass}`}>
          <StatusIcon className="w-4 h-4" />
        </div>
        
        <div className="min-w-0 space-y-0.5">
          {/* Level 1: Document Name & Requirement Tags */}
          <div className="flex items-center gap-2 flex-wrap">
            <h4 className="text-sm font-bold text-slate-900 truncate">
              {doc?.document_name || item.title}
            </h4>
            
            {item.is_required && (
              <span className="px-1.5 py-0.5 rounded bg-rose-50 text-rose-700 text-[10px] font-extrabold border border-rose-200">
                REQUIRED
              </span>
            )}
            
            {item.is_conditional && (
              <span className="px-1.5 py-0.5 rounded bg-amber-50 text-amber-700 text-[10px] font-extrabold border border-amber-200">
                CONDITIONAL
              </span>
            )}
          </div>

          {/* Level 2: Requirement Rationale (from real backend engine) */}
          <p className="text-xs text-slate-500 line-clamp-2">
            {doc?.description || item.reason || item.condition_reason || 'Required for customs clearance and operational booking.'}
          </p>

          {/* Level 3: Supporting File Metadata */}
          {doc && (
            <div className="flex items-center gap-2.5 text-[11px] text-slate-400 pt-0.5">
              {doc.file_name && (
                <span className="text-slate-600 font-medium truncate max-w-[220px]">
                  📄 {doc.file_name}
                </span>
              )}
              {doc.uploaded_at && (
                <span>• Uploaded {new Date(doc.uploaded_at).toLocaleDateString()}</span>
              )}
              {doc.uploaded_by && (
                <span>by {doc.uploaded_by}</span>
              )}
            </div>
          )}
        </div>
      </div>

      {/* Right: Operational Status & Action Hierarchy */}
      <div className="flex items-center gap-2.5 flex-wrap md:flex-nowrap justify-end flex-shrink-0">
        
        {/* Compact Status Badge */}
        <span className={`inline-flex items-center gap-1 px-2.5 py-1 rounded-full text-xs font-bold border ${config.badgeClass}`}>
          {config.label}
        </span>

        {/* Primary Operational Action */}
        {!doc ? (
          <button
            onClick={onUpload}
            className="inline-flex items-center gap-1.5 px-3 py-1.5 bg-indigo-600 hover:bg-indigo-700 text-white text-xs font-semibold rounded-lg shadow-xs transition-all active:scale-95"
            data-testid={`upload-${item.doc_type}-btn`}
          >
            <Upload className="w-3.5 h-3.5" />
            Upload Document
          </button>
        ) : (
          <div className="flex items-center gap-1.5 relative">
            
            {/* View / Detail action */}
            <button
              onClick={() => onOpenDrawer(doc)}
              className="px-2.5 py-1 bg-slate-100 hover:bg-slate-200 text-slate-700 text-xs font-semibold rounded-lg transition-colors inline-flex items-center gap-1"
            >
              <Eye className="w-3.5 h-3.5 text-slate-500" />
              View
            </button>

            {/* Lifecycle quick transitions */}
            {doc.status === 'UPLOADED' && (
              <button
                onClick={() => onUpdateStatus(doc.id, 'UNDER_REVIEW')}
                className="px-2.5 py-1 bg-amber-50 hover:bg-amber-100 text-amber-700 border border-amber-200 text-xs font-semibold rounded-lg transition-colors"
              >
                Start Review
              </button>
            )}

            {doc.status === 'UNDER_REVIEW' && (
              <>
                <button
                  onClick={() => onUpdateStatus(doc.id, 'APPROVED')}
                  className="px-2.5 py-1 bg-emerald-600 hover:bg-emerald-700 text-white text-xs font-bold rounded-lg shadow-xs transition-colors"
                  data-testid={`approve-${doc.id}-btn`}
                >
                  Approve
                </button>
                <button
                  onClick={onReject}
                  className="px-2 py-1 bg-rose-50 hover:bg-rose-100 text-rose-700 border border-rose-200 text-xs font-semibold rounded-lg transition-colors"
                  data-testid={`reject-${doc.id}-btn`}
                >
                  Reject
                </button>
              </>
            )}

            {doc.status === 'APPROVED' && (
              <button
                onClick={onUpload}
                className="px-2.5 py-1 bg-slate-50 hover:bg-slate-100 text-slate-600 border border-slate-200 text-xs font-medium rounded-lg transition-colors"
              >
                Replace
              </button>
            )}

            {/* Overflow Menu [ ⋮ ] */}
            <div className="relative">
              <button
                onClick={() => setMenuOpen(!menuOpen)}
                className="p-1 text-slate-400 hover:text-slate-700 hover:bg-slate-100 rounded-lg transition-colors"
              >
                <MoreVertical className="w-4 h-4" />
              </button>

              {menuOpen && (
                <>
                  <div className="fixed inset-0 z-20" onClick={() => setMenuOpen(false)} />
                  <div className="absolute right-0 mt-1 w-44 bg-white border border-slate-200 rounded-xl shadow-lg z-30 py-1.5 text-xs">
                    <button
                      onClick={() => {
                        setMenuOpen(false);
                        onOpenDrawer(doc);
                      }}
                      className="w-full text-left px-3.5 py-1.5 hover:bg-slate-50 text-slate-700 font-medium flex items-center gap-2"
                    >
                      <FileText className="w-3.5 h-3.5 text-slate-400" />
                      View Audit Log
                    </button>

                    <button
                      onClick={() => {
                        setMenuOpen(false);
                        onUpload();
                      }}
                      className="w-full text-left px-3.5 py-1.5 hover:bg-slate-50 text-slate-700 font-medium flex items-center gap-2"
                    >
                      <RotateCcw className="w-3.5 h-3.5 text-slate-400" />
                      Replace Document
                    </button>

                    {doc.status !== 'APPROVED' && (
                      <button
                        onClick={() => {
                          setMenuOpen(false);
                          onUpdateStatus(doc.id, 'APPROVED');
                        }}
                        className="w-full text-left px-3.5 py-1.5 hover:bg-emerald-50 text-emerald-700 font-medium flex items-center gap-2"
                      >
                        <Check className="w-3.5 h-3.5" />
                        Mark Approved
                      </button>
                    )}

                    <div className="border-t border-slate-100 my-1"></div>

                    <button
                      onClick={() => {
                        setMenuOpen(false);
                        onDelete(doc.id);
                      }}
                      className="w-full text-left px-3.5 py-1.5 hover:bg-rose-50 text-rose-600 font-medium flex items-center gap-2"
                    >
                      <Trash2 className="w-3.5 h-3.5 text-rose-500" />
                      Delete Record
                    </button>
                  </div>
                </>
              )}
            </div>

          </div>
        )}

      </div>
    </div>
  );
}

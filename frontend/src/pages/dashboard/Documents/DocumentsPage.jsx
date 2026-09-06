import React, { useState, useEffect, useMemo } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  FileText, Upload, CheckCircle2, ShieldCheck, AlertCircle, ExternalLink,
  Search, Filter, Eye, Download, Trash2, Clock, AlertTriangle, Layers, Building, Truck,
  FileCheck, Anchor, Home, Receipt, Package, DollarSign
} from 'lucide-react';
import api from '../../../services/api';
import { approvalsService } from '../../../services/approvalsService';
import PageHeader from '../../../components/dashboard/PageHeader';
import DocumentUploadModal from './DocumentUploadModal';
import DocumentDetailsModal from './DocumentDetailsModal';
import ModuleHeroEmptyState from '../../../components/dashboard/ModuleHeroEmptyState';
import './DocumentsPage.css';

export default function DocumentsPage() {
  const [documents, setDocuments] = useState([]);
  const [tabFilter, setTabFilter] = useState('ALL'); // ALL, BOL, CUSTOMS

  // Filters state
  const [docTypeFilter, setDocTypeFilter] = useState('ALL');
  const [bolSubFilter, setBolSubFilter] = useState('ALL_BILLS'); // ALL_BILLS, HBL, MBL
  const [customsSubFilter, setCustomsSubFilter] = useState('ALL_CUSTOMS'); // ALL_CUSTOMS, COMMERCIAL_INVOICE, PACKING_LIST
  const [statusFilter, setStatusFilter] = useState('ALL');
  const [searchQuery, setSearchQuery] = useState('');

  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  // Modals state
  const [isUploadModalOpen, setIsUploadModalOpen] = useState(false);
  const [selectedDocForDetails, setSelectedDocForDetails] = useState(null);

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

  const handleUploadSuccess = () => {
    fetchDocuments();
  };

  const handleDeleteSuccess = (deletedId) => {
    setDocuments((prev) => prev.filter((d) => d.id !== deletedId));
  };

  // Tab counts for main tab buttons
  const tabCounts = useMemo(() => {
    const total = documents.length;
    const bol = documents.filter((d) => d.doc_type === 'HBL' || d.doc_type === 'MBL').length;
    const customs = documents.filter(
      (d) => d.doc_type === 'COMMERCIAL_INVOICE' || d.doc_type === 'INVOICE' || d.doc_type === 'PACKING_LIST'
    ).length;
    return { total, bol, customs };
  }, [documents]);

  // Dynamic KPI Metrics derived directly from real backend document records
  const kpis = useMemo(() => {
    if (tabFilter === 'BOL') {
      const bolDocs = documents.filter((d) => d.doc_type === 'HBL' || d.doc_type === 'MBL');
      const totalBol = bolDocs.length;
      const hblCount = bolDocs.filter((d) => d.doc_type === 'HBL').length;
      const mblCount = bolDocs.filter((d) => d.doc_type === 'MBL').length;
      const pendingBol = bolDocs.filter(
        (d) => d.status === 'PENDING_VERIFICATION' || d.status === 'PENDING'
      ).length;
      const verifiedBol = bolDocs.filter(
        (d) => !d.status || d.status === 'VERIFIED' || d.status === 'APPROVED'
      ).length;

      return {
        total: totalBol,
        hbl: hblCount,
        mbl: mblCount,
        pending: pendingBol,
        verified: verifiedBol
      };
    }

    if (tabFilter === 'CUSTOMS') {
      const customsDocs = documents.filter(
        (d) => d.doc_type === 'COMMERCIAL_INVOICE' || d.doc_type === 'INVOICE' || d.doc_type === 'PACKING_LIST'
      );
      const totalCustoms = customsDocs.length;
      const invoiceCount = customsDocs.filter((d) => d.doc_type === 'COMMERCIAL_INVOICE' || d.doc_type === 'INVOICE').length;
      const packingCount = customsDocs.filter((d) => d.doc_type === 'PACKING_LIST').length;
      const pendingCustoms = customsDocs.filter(
        (d) => d.status === 'PENDING_VERIFICATION' || d.status === 'PENDING'
      ).length;
      const verifiedCustoms = customsDocs.filter(
        (d) => !d.status || d.status === 'VERIFIED' || d.status === 'APPROVED'
      ).length;

      return {
        total: totalCustoms,
        invoices: invoiceCount,
        packing: packingCount,
        pending: pendingCustoms,
        verified: verifiedCustoms
      };
    }

    // Default All Documents KPIs
    const total = documents.length;
    const pending = documents.filter(
      (d) => d.status === 'PENDING_VERIFICATION' || d.status === 'PENDING'
    ).length;
    const verified = documents.filter(
      (d) => !d.status || d.status === 'VERIFIED' || d.status === 'APPROVED'
    ).length;
    const attention = documents.filter(
      (d) => d.status === 'REQUIRES_ATTENTION' || d.status === 'EXCEPTION' || (d.discrepancy_count && d.discrepancy_count > 0)
    ).length;

    return { total, pending, verified, attention };
  }, [documents, tabFilter]);

  // Multi-criteria Filtering & Searching
  const filteredDocuments = useMemo(() => {
    return documents.filter((d) => {
      // 1. Top Tab Filter
      if (tabFilter === 'BOL') {
        if (d.doc_type !== 'HBL' && d.doc_type !== 'MBL') return false;
        if (bolSubFilter === 'HBL' && d.doc_type !== 'HBL') return false;
        if (bolSubFilter === 'MBL' && d.doc_type !== 'MBL') return false;
      } else if (tabFilter === 'CUSTOMS') {
        if (d.doc_type !== 'COMMERCIAL_INVOICE' && d.doc_type !== 'INVOICE' && d.doc_type !== 'PACKING_LIST') return false;
        if (customsSubFilter === 'COMMERCIAL_INVOICE' && d.doc_type !== 'COMMERCIAL_INVOICE' && d.doc_type !== 'INVOICE') return false;
        if (customsSubFilter === 'PACKING_LIST' && d.doc_type !== 'PACKING_LIST') return false;
      } else {
        // ALL tab
        if (docTypeFilter !== 'ALL' && d.doc_type !== docTypeFilter) return false;
      }

      // 2. Status Select Filter
      if (statusFilter !== 'ALL') {
        const curStatus = d.status || 'VERIFIED';
        if (statusFilter === 'VERIFIED' && curStatus !== 'VERIFIED' && curStatus !== 'APPROVED') return false;
        if (statusFilter === 'PENDING' && curStatus !== 'PENDING_VERIFICATION' && curStatus !== 'PENDING') return false;
        if (statusFilter === 'ATTENTION' && curStatus !== 'REQUIRES_ATTENTION' && curStatus !== 'EXCEPTION') return false;
      }

      // 3. Search Query Match
      if (searchQuery.trim()) {
        const q = searchQuery.toLowerCase();
        const docName = (d.file_name || '').toLowerCase();
        const origName = (d.original_file_name || '').toLowerCase();
        const docType = (d.doc_type || '').toLowerCase();
        const shipRef = (d.shipment_ref || '').toLowerCase();
        const custName = (d.customer_name || '').toLowerCase();
        const bookRef = (d.booking_ref || '').toLowerCase();
        const leadName = (d.lead_name || '').toLowerCase();

        return (
          docName.includes(q) ||
          origName.includes(q) ||
          docType.includes(q) ||
          shipRef.includes(q) ||
          custName.includes(q) ||
          bookRef.includes(q) ||
          leadName.includes(q)
        );
      }

      return true;
    });
  }, [documents, tabFilter, docTypeFilter, bolSubFilter, customsSubFilter, statusFilter, searchQuery]);

  const formatFileSize = (bytes) => {
    if (!bytes) return '';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i];
  };

  const getDocTypeBadgeClass = (type) => {
    switch (type) {
      case 'HBL': return 'doc-type-badge hbl';
      case 'MBL': return 'doc-type-badge mbl';
      case 'COMMERCIAL_INVOICE':
      case 'INVOICE': return 'doc-type-badge invoice';
      case 'PACKING_LIST': return 'doc-type-badge packing';
      case 'CONTRACT': return 'doc-type-badge contract';
      case 'QUOTATION': return 'doc-type-badge quotation';
      default: return 'doc-type-badge other';
    }
  };

  const getDownloadUrl = (doc) => {
    if (doc.file_path && doc.file_path.startsWith('http')) return doc.file_path;
    return `http://localhost:8080/uploads/${doc.s3_key || doc.file_name}`;
  };

  const getUploadInitialType = () => {
    if (tabFilter === 'BOL') return bolSubFilter === 'MBL' ? 'MBL' : 'HBL';
    if (tabFilter === 'CUSTOMS') return customsSubFilter === 'PACKING_LIST' ? 'PACKING_LIST' : 'COMMERCIAL_INVOICE';
    return 'HBL';
  };

  return (
    <div className="documents-page">
      {/* Header Row */}
      <div className="documents-header-row">
        <PageHeader
          title="Document Compliance Center"
          subtitle="AI OCR verification, 3-way discrepancy checks, and automated compliance validation for HBL, MBL, Invoices, and Packing Lists"
        />
        <button
          className="btn-upload-trigger"
          onClick={() => setIsUploadModalOpen(true)}
        >
          <Upload size={16} /> {tabFilter === 'BOL' ? 'Upload Bill of Lading' : tabFilter === 'CUSTOMS' ? 'Upload Invoice / Packing List' : 'Upload Document'}
        </button>
      </div>

      {/* KPI / Summary Cards Area */}
      {tabFilter === 'BOL' ? (
        <div className="doc-kpi-grid">
          <div className="doc-kpi-card blue">
            <div className="doc-kpi-icon-box">
              <FileCheck size={20} />
            </div>
            <div className="doc-kpi-content">
              <span className="doc-kpi-value">{loading ? '—' : kpis.total}</span>
              <span className="doc-kpi-label">Total Bills of Lading</span>
            </div>
          </div>

          <div className="doc-kpi-card purple">
            <div className="doc-kpi-icon-box">
              <Home size={20} />
            </div>
            <div className="doc-kpi-content">
              <span className="doc-kpi-value">{loading ? '—' : kpis.hbl}</span>
              <span className="doc-kpi-label">House Bills (HBL)</span>
            </div>
          </div>

          <div className="doc-kpi-card indigo">
            <div className="doc-kpi-icon-box">
              <Anchor size={20} />
            </div>
            <div className="doc-kpi-content">
              <span className="doc-kpi-value">{loading ? '—' : kpis.mbl}</span>
              <span className="doc-kpi-label">Master Bills (MBL)</span>
            </div>
          </div>

          <div className="doc-kpi-card green">
            <div className="doc-kpi-icon-box">
              <ShieldCheck size={20} />
            </div>
            <div className="doc-kpi-content">
              <span className="doc-kpi-value">{loading ? '—' : kpis.verified}</span>
              <span className="doc-kpi-label">Verified & Approved</span>
            </div>
          </div>
        </div>
      ) : tabFilter === 'CUSTOMS' ? (
        <div className="doc-kpi-grid">
          <div className="doc-kpi-card blue">
            <div className="doc-kpi-icon-box">
              <Receipt size={20} />
            </div>
            <div className="doc-kpi-content">
              <span className="doc-kpi-value">{loading ? '—' : kpis.total}</span>
              <span className="doc-kpi-label">Total Invoices & Packing</span>
            </div>
          </div>

          <div className="doc-kpi-card green">
            <div className="doc-kpi-icon-box">
              <Receipt size={20} />
            </div>
            <div className="doc-kpi-content">
              <span className="doc-kpi-value">{loading ? '—' : kpis.invoices}</span>
              <span className="doc-kpi-label">Commercial Invoices</span>
            </div>
          </div>

          <div className="doc-kpi-card orange">
            <div className="doc-kpi-icon-box">
              <Package size={20} />
            </div>
            <div className="doc-kpi-content">
              <span className="doc-kpi-value">{loading ? '—' : kpis.packing}</span>
              <span className="doc-kpi-label">Packing Lists</span>
            </div>
          </div>

          <div className="doc-kpi-card green">
            <div className="doc-kpi-icon-box">
              <ShieldCheck size={20} />
            </div>
            <div className="doc-kpi-content">
              <span className="doc-kpi-value">{loading ? '—' : kpis.verified}</span>
              <span className="doc-kpi-label">Verified & Approved</span>
            </div>
          </div>
        </div>
      ) : (
        <div className="doc-kpi-grid">
          <div className="doc-kpi-card blue">
            <div className="doc-kpi-icon-box">
              <Layers size={20} />
            </div>
            <div className="doc-kpi-content">
              <span className="doc-kpi-value">{loading ? '—' : kpis.total}</span>
              <span className="doc-kpi-label">Total Documents</span>
            </div>
          </div>

          <div className="doc-kpi-card yellow">
            <div className="doc-kpi-icon-box">
              <Clock size={20} />
            </div>
            <div className="doc-kpi-content">
              <span className="doc-kpi-value">{loading ? '—' : kpis.pending}</span>
              <span className="doc-kpi-label">Pending Verification</span>
            </div>
          </div>

          <div className="doc-kpi-card green">
            <div className="doc-kpi-icon-box">
              <ShieldCheck size={20} />
            </div>
            <div className="doc-kpi-content">
              <span className="doc-kpi-value">{loading ? '—' : kpis.verified}</span>
              <span className="doc-kpi-label">Verified & Approved</span>
            </div>
          </div>

          <div className="doc-kpi-card red">
            <div className="doc-kpi-icon-box">
              <AlertTriangle size={20} />
            </div>
            <div className="doc-kpi-content">
              <span className="doc-kpi-value">{loading ? '—' : kpis.attention}</span>
              <span className="doc-kpi-label">Requires Attention</span>
            </div>
          </div>
        </div>
      )}

      {/* Tabs Row */}
      <div className="documents-tabs-row">
        {[
          { id: 'ALL', label: `All Documents (${loading ? '—' : tabCounts.total})` },
          { id: 'BOL', label: `Bills of Lading (${loading ? '—' : tabCounts.bol})` },
          { id: 'CUSTOMS', label: `Invoices & Packing Lists (${loading ? '—' : tabCounts.customs})` }
        ].map((t) => (
          <button
            key={t.id}
            className={`tab-btn ${tabFilter === t.id ? 'active' : ''}`}
            onClick={() => setTabFilter(t.id)}
          >
            {t.label}
          </button>
        ))}
      </div>

      {/* Document Toolbar: Search & Select Filters */}
      <div className="doc-toolbar-card">
        <div className="doc-toolbar-left">
          {/* Search Box */}
          <div className="doc-search-box">
            <Search size={15} className="doc-search-icon" />
            <input
              type="text"
              placeholder={
                tabFilter === 'BOL'
                  ? 'Search BL number, shipment, customer...'
                  : tabFilter === 'CUSTOMS'
                    ? 'Search invoice #, packing list, customer...'
                    : 'Search by file name, shipment, customer, booking...'
              }
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
            />
          </div>

          {/* Tab Specific Filter Dropdowns */}
          {tabFilter === 'BOL' ? (
            <select
              className="doc-filter-select"
              value={bolSubFilter}
              onChange={(e) => setBolSubFilter(e.target.value)}
            >
              <option value="ALL_BILLS">All Bills of Lading (HBL + MBL)</option>
              <option value="HBL">House Bills of Lading (HBL)</option>
              <option value="MBL">Master Bills of Lading (MBL)</option>
            </select>
          ) : tabFilter === 'CUSTOMS' ? (
            <select
              className="doc-filter-select"
              value={customsSubFilter}
              onChange={(e) => setCustomsSubFilter(e.target.value)}
            >
              <option value="ALL_CUSTOMS">All Invoices & Packing Lists</option>
              <option value="COMMERCIAL_INVOICE">Commercial Invoices</option>
              <option value="PACKING_LIST">Packing Lists</option>
            </select>
          ) : (
            <select
              className="doc-filter-select"
              value={docTypeFilter}
              onChange={(e) => setDocTypeFilter(e.target.value)}
            >
              <option value="ALL">All Categories</option>
              <option value="HBL">HBL (House Bill)</option>
              <option value="MBL">Master Bill (MBL)</option>
              <option value="COMMERCIAL_INVOICE">Commercial Invoice</option>
              <option value="PACKING_LIST">Packing List</option>
              <option value="CONTRACT">Carrier Contract</option>
              <option value="QUOTATION">Quotation / Rate</option>
              <option value="OTHER">Other / General</option>
            </select>
          )}

          {/* Status Select */}
          <select
            className="doc-filter-select"
            value={statusFilter}
            onChange={(e) => setStatusFilter(e.target.value)}
          >
            <option value="ALL">All Statuses</option>
            <option value="VERIFIED">Verified / Approved</option>
            <option value="PENDING">Pending Review</option>
            <option value="ATTENTION">Requires Attention</option>
          </select>
        </div>

        <div style={{ fontSize: '0.8rem', color: '#64748B', fontWeight: 600 }}>
          Showing <strong>{filteredDocuments.length}</strong> {tabFilter === 'BOL' ? 'Bills of Lading' : tabFilter === 'CUSTOMS' ? 'Invoices & Packing Lists' : 'documents'}
        </div>
      </div>

      {/* Error Banner */}
      {error && (
        <div className="documents-error-banner">
          <span>⚠️ {error}</span>
          <button onClick={fetchDocuments} className="btn-retry">Retry</button>
        </div>
      )}

      {/* Main Content Area */}
      {loading ? (
        <div className="documents-skeleton">
          {[1, 2, 3, 4].map((i) => (
            <div key={i} className="skeleton-doc-row" />
          ))}
        </div>
      ) : filteredDocuments.length === 0 && !searchQuery && statusFilter === 'ALL' && tabFilter === 'ALL' ? (
        <ModuleHeroEmptyState
          icon={<FileText size={28} />}
          badgeTheme="indigo"
          title="No Freight Documents Uploaded Yet"
          description="Centralized document compliance vault for Master/House Bills of Lading (MBL/HBL), commercial invoices, packing lists, customs entries, and carrier release certificates."
          primaryAction={{
            label: 'Upload First Document',
            icon: <Upload size={15} />,
            onClick: () => setIsUploadModalOpen(true),
          }}
          secondaryAction={{
            label: 'Explore Active Shipments',
            icon: <Package size={15} />,
            onClick: () => navigate('/dashboard/shipments'),
          }}
          features={[
            {
              icon: <ShieldCheck size={18} />,
              iconBg: '#eff6ff',
              iconColor: '#2563eb',
              title: 'Automated 3-Way Compliance Check',
              desc: 'Cross-validate piece counts, gross weight, and cargo descriptions across MBL, HBL, and packing lists.',
            },
            {
              icon: <FileCheck size={18} />,
              iconBg: '#ecfdf5',
              iconColor: '#059669',
              title: 'Verification & Audit Workflow',
              desc: 'Enforce team verification approvals, customs release timestamps, and document compliance checks.',
            },
            {
              icon: <Layers size={18} />,
              iconBg: '#f5f3ff',
              iconColor: '#7c3aed',
              title: 'Contextual Freight File Linking',
              desc: 'Automatically index files under corresponding customer accounts, bookings, and active operations.',
            },
          ]}
        />
      ) : filteredDocuments.length === 0 ? (
        <div className="documents-empty-state">
          <div className="empty-icon-box">{tabFilter === 'BOL' ? '📜' : tabFilter === 'CUSTOMS' ? '📄' : '📁'}</div>
          <h3>
            {searchQuery || statusFilter !== 'ALL' || (tabFilter === 'BOL' ? bolSubFilter !== 'ALL_BILLS' : tabFilter === 'CUSTOMS' ? customsSubFilter !== 'ALL_CUSTOMS' : docTypeFilter !== 'ALL')
              ? 'No matching documents found'
              : tabFilter === 'BOL' ? 'No Bills of Lading yet' : tabFilter === 'CUSTOMS' ? 'No Commercial Invoices or Packing Lists yet' : 'No shipment documents yet'}
          </h3>
          <p>
            {searchQuery || statusFilter !== 'ALL' || (tabFilter === 'BOL' ? bolSubFilter !== 'ALL_BILLS' : tabFilter === 'CUSTOMS' ? customsSubFilter !== 'ALL_CUSTOMS' : docTypeFilter !== 'ALL')
              ? 'Try clearing your search query or adjusting filter dropdowns to find document records.'
              : tabFilter === 'BOL'
                ? 'House Bills of Lading (HBL) and Master Bills of Lading (MBL) uploaded to active bookings will appear here.'
                : tabFilter === 'CUSTOMS'
                  ? 'Commercial Invoices and Packing Lists uploaded for customs clearance and 3-way compliance checks will appear here.'
                  : 'Shipment documents (HBL, MBL, Commercial Invoices, Packing Lists) will appear here once attached to active bookings.'}
          </p>
          <div style={{ display: 'flex', gap: '12px', justifyContent: 'center', marginTop: '20px' }}>
            <button className="btn-primary-action" onClick={() => setIsUploadModalOpen(true)}>
              <Upload size={15} style={{ marginRight: 6 }} /> {tabFilter === 'BOL' ? 'Upload Bill of Lading' : tabFilter === 'CUSTOMS' ? 'Upload Invoice / Packing List' : 'Upload Document'}
            </button>
            {(searchQuery || statusFilter !== 'ALL' || (tabFilter === 'BOL' ? bolSubFilter !== 'ALL_BILLS' : tabFilter === 'CUSTOMS' ? customsSubFilter !== 'ALL_CUSTOMS' : docTypeFilter !== 'ALL')) && (
              <button
                className="btn-secondary-action"
                onClick={() => {
                  setSearchQuery('');
                  setDocTypeFilter('ALL');
                  setBolSubFilter('ALL_BILLS');
                  setCustomsSubFilter('ALL_CUSTOMS');
                  setStatusFilter('ALL');
                }}
              >
                Reset Filters
              </button>
            )}
          </div>
        </div>
      ) : (
        <div className="documents-table-card">
          <table className="documents-table">
            <thead>
              <tr>
                <th>{tabFilter === 'BOL' ? 'Bill of Lading Document' : tabFilter === 'CUSTOMS' ? 'Invoice / Packing List' : 'Document'}</th>
                <th>Type</th>
                <th>Related Context</th>
                <th>Customer</th>
                <th>Status</th>
                <th>Uploaded</th>
                <th style={{ textAlign: 'right' }}>Actions</th>
              </tr>
            </thead>
            <tbody>
              {filteredDocuments.map((d) => (
                <tr key={d.id} onClick={() => setSelectedDocForDetails(d)}>
                  <td>
                    <div className="doc-name-cell">
                      <FileText size={18} className="doc-icon" />
                      <div>
                        <strong className="doc-name-text" style={{ color: '#0F172A', display: 'block', fontSize: '0.86rem' }} title={d.original_file_name || d.file_name}>
                          {d.original_file_name || d.file_name || `DOC-${d.id}`}
                        </strong>
                        <span style={{ fontSize: '0.74rem', color: '#64748B' }}>
                          {d.file_type || 'PDF'} {d.file_size ? `· ${formatFileSize(d.file_size)}` : ''}
                        </span>
                      </div>
                    </div>
                  </td>
                  <td>
                    <span className={getDocTypeBadgeClass(d.doc_type)}>
                      {d.doc_type === 'HBL' ? (
                        <>
                          <Home size={11} /> House BL
                        </>
                      ) : d.doc_type === 'MBL' ? (
                        <>
                          <Anchor size={11} /> Master BL
                        </>
                      ) : d.doc_type === 'COMMERCIAL_INVOICE' || d.doc_type === 'INVOICE' ? (
                        <>
                          <Receipt size={11} /> Commercial Invoice
                        </>
                      ) : d.doc_type === 'PACKING_LIST' ? (
                        <>
                          <Package size={11} /> Packing List
                        </>
                      ) : (
                        d.doc_type
                      )}
                    </span>
                  </td>
                  <td>
                    <div style={{ display: 'flex', flexDirection: 'column', gap: 2, fontSize: '0.78rem' }}>
                      {d.shipment_ref && <span>🚢 Shipment: <strong>{d.shipment_ref}</strong></span>}
                      {d.booking_ref && <span>📋 Booking: <strong>{d.booking_ref}</strong></span>}
                      {d.lead_name && <span>👤 Lead: <strong>{d.lead_name}</strong></span>}
                      {!d.shipment_ref && !d.booking_ref && !d.lead_name && <span style={{ color: '#94A3B8' }}>Standalone Document</span>}
                    </div>
                  </td>
                  <td>
                    {d.customer_name ? (
                      <span className="doc-customer-text" style={{ fontSize: '0.8rem', fontWeight: 650, color: '#334155', display: 'inline-block' }} title={d.customer_name}>{d.customer_name}</span>
                    ) : (
                      <span style={{ color: '#94A3B8' }}>—</span>
                    )}
                  </td>
                  <td>
                    <span className={`compliance-pill ${(d.status || 'VERIFIED').toLowerCase()}`}>
                      <CheckCircle2 size={12} />
                      {d.status || 'VERIFIED'}
                    </span>
                  </td>
                  <td>{d.created_at ? new Date(d.created_at).toLocaleDateString() : '—'}</td>
                  <td style={{ textAlign: 'right' }} onClick={(e) => e.stopPropagation()}>
                    <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'flex-end', gap: 6 }}>
                      <button
                        className="doc-action-btn"
                        title="Request Approval"
                        onClick={async (e) => {
                          e.stopPropagation();
                          try {
                            const docName = d.original_file_name || d.file_name || `DOC-${d.id}`;
                            const parsedDocId = parseInt(d.id, 10) || 0;
                            const parsedShipmentId = parseInt(d.shipment_id, 10) || 0;
                            const parsedCustomerId = parseInt(d.customer_id, 10) || 0;

                            await approvalsService.createApproval({
                              title: `${d.doc_type || 'Document'} Approval - ${docName}`,
                              category: 'DOCUMENTS',
                              type: 'Document Approval',
                              priority: 'HIGH',
                              related_ref: `Document: ${docName}`,
                              customer_name: d.customer_name || 'Associated Customer',
                              document_id: parsedDocId,
                              shipment_id: parsedShipmentId,
                              customer_id: parsedCustomerId,
                              description: `Compliance sign-off requested for document: ${docName} (${d.doc_type || 'DOCUMENT'}).`,
                            });
                            alert(`Approval request submitted to Approval Center for ${docName}`);
                          } catch (err) {
                            console.error(err);
                            alert('Failed to request approval. Please try again.');
                          }
                        }}
                      >
                        <Clock size={15} style={{ color: '#D97706' }} />
                      </button>

                      <button
                        className="doc-action-btn"
                        title="View Details"
                        onClick={() => setSelectedDocForDetails(d)}
                      >
                        <Eye size={15} />
                      </button>

                      <a
                        href={getDownloadUrl(d)}
                        target="_blank"
                        rel="noreferrer"
                        className="doc-action-btn"
                        title="Download Document"
                      >
                        <Download size={15} />
                      </a>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {/* Upload Document Modal */}
      <DocumentUploadModal
        isOpen={isUploadModalOpen}
        onClose={() => setIsUploadModalOpen(false)}
        onUploadSuccess={handleUploadSuccess}
        initialDocType={getUploadInitialType()}
      />

      {/* View Document Details Modal */}
      <DocumentDetailsModal
        isOpen={!!selectedDocForDetails}
        onClose={() => setSelectedDocForDetails(null)}
        document={selectedDocForDetails}
        onDeleteSuccess={handleDeleteSuccess}
      />
    </div>
  );
}

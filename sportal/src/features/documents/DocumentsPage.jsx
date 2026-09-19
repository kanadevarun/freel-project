import React, { useState, useEffect, useMemo } from 'react';
import { useSearchParams, Link } from 'react-router-dom';
import {
  FileText,
  FileCheck,
  ShieldCheck,
  HardDrive,
  Search,
  RefreshCw,
  AlertCircle,
  Download,
  Eye,
  CheckCircle2,
  Clock,
  Filter,
  FileSpreadsheet,
  Building2,
  AlertTriangle,
  XCircle,
  ExternalLink,
  ChevronRight,
  Sparkles,
  Layers,
  X,
  Check,
  Ban,
  FileCode,
  Calendar,
  ArrowUpRight
} from 'lucide-react';
import PageHeader from '../../components/common/PageHeader';
import KpiCard from '../../components/common/KpiCard';
import StatusBadge from '../../components/common/StatusBadge';
import EmptyState from '../../components/common/EmptyState';
import { sportalService } from '../../services/sportalService';

export function DocumentsPage() {
  const [searchParams, setSearchParams] = useSearchParams();
  const initialOrgId = searchParams.get('orgId') || '2';

  const [organizations, setOrganizations] = useState([]);
  const [selectedOrgId, setSelectedOrgId] = useState(initialOrgId);
  const [documents, setDocuments] = useState([]);
  const [summary, setSummary] = useState(null);
  const [platformOverview, setPlatformOverview] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  // Filters
  const [search, setSearch] = useState('');
  const [typeFilter, setTypeFilter] = useState('ALL');
  const [statusFilter, setStatusFilter] = useState('ALL');
  const [expiryFilter, setExpiryFilter] = useState('ALL');

  // Modal / Drawer State
  const [inspectDoc, setInspectDoc] = useState(null);
  const [inspectLoading, setInspectLoading] = useState(false);
  const [activeInspectTab, setActiveInspectTab] = useState('metadata');
  const [actionLoading, setActionLoading] = useState(false);
  const [actionSuccessMsg, setActionSuccessMsg] = useState('');
  const [actionErrorMsg, setActionErrorMsg] = useState('');

  // 1. Fetch organization list for the switcher
  useEffect(() => {
    sportalService
      .getOrganizations({ pageSize: 100 })
      .then((res) => {
        const orgs = res?.items || res?.data?.items || res?.data?.organizations || res?.organizations || [];
        setOrganizations(orgs);
        if (!searchParams.get('orgId') && orgs.length > 0) {
          const found = orgs.find((o) => String(o.id) === String(selectedOrgId));
          if (!found) {
            setSelectedOrgId(String(orgs[0].id));
          }
        }
      })
      .catch((err) => {
        console.error('Failed to load organizations for document selector:', err);
      });
  }, [searchParams, selectedOrgId]);

  // Update URL search params when selectedOrgId changes
  useEffect(() => {
    if (selectedOrgId && selectedOrgId !== 'all') {
      setSearchParams({ orgId: selectedOrgId });
    } else {
      setSearchParams({});
    }
  }, [selectedOrgId, setSearchParams]);

  // 2. Load Documents & Summary for selected organization or platform overview
  const loadDocuments = async () => {
    setLoading(true);
    setError(null);
    try {
      if (selectedOrgId === 'all') {
        const pRes = await sportalService.getPlatformDocumentsOverview();
        const pData = pRes?.data || pRes || {};
        setPlatformOverview(pData);
        setDocuments([]);
        setSummary(null);
      } else {
        const orgIdNum = Number(selectedOrgId);
        const res = await sportalService.getCustomerDocumentsPaginated(orgIdNum, {
          search,
          doc_type: typeFilter,
          status: statusFilter,
          expiry_filter: expiryFilter,
          limit: 100,
        });

        const data = res?.data || res || {};
        const docList = data.documents || data.items || [];
        setDocuments(docList);
        setSummary(data.summary || null);
        setPlatformOverview(null);
      }
    } catch (err) {
      console.error('Failed to load documents:', err);
      setError(err?.message || 'Unable to load document vault records. Please verify connection.');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadDocuments();
  }, [selectedOrgId, typeFilter, statusFilter, expiryFilter]);

  // Handle client-side search query
  const filteredDocs = useMemo(() => {
    if (!search.trim()) return documents;
    const term = search.toLowerCase().trim();
    return documents.filter((doc) => {
      const fileName = (doc.file_name || doc.original_file_name || '').toLowerCase();
      const docName = (doc.document_name || '').toLowerCase();
      const docType = (doc.doc_type || '').toLowerCase();
      const status = (doc.status || '').toLowerCase();
      const shipRef = (doc.shipment_ref || '').toLowerCase();
      const bookRef = (doc.booking_ref || '').toLowerCase();
      const custName = (doc.customer_name || '').toLowerCase();
      const idStr = String(doc.id);

      return (
        fileName.includes(term) ||
        docName.includes(term) ||
        docType.includes(term) ||
        status.includes(term) ||
        shipRef.includes(term) ||
        bookRef.includes(term) ||
        custName.includes(term) ||
        idStr.includes(term)
      );
    });
  }, [documents, search]);

  // Open Document Detail & OCR Inspector
  const handleOpenInspector = async (doc) => {
    setInspectDoc(doc);
    setInspectLoading(true);
    setActiveInspectTab('metadata');
    setActionSuccessMsg('');
    setActionErrorMsg('');
    try {
      const res = await sportalService.getCustomerDocumentDetail(doc.org_id, doc.id);
      const detail = res?.data || res;
      if (detail && detail.id) {
        setInspectDoc(detail);
      }
    } catch (err) {
      console.error('Failed to load deep document details:', err);
    } finally {
      setInspectLoading(false);
    }
  };

  // Status Update (Verify / Reject)
  const handleUpdateStatus = async (newStatus, reason = '') => {
    if (!inspectDoc) return;
    setActionLoading(true);
    setActionSuccessMsg('');
    setActionErrorMsg('');
    try {
      const res = await sportalService.updateCustomerDocumentStatus(
        inspectDoc.org_id,
        inspectDoc.id,
        { status: newStatus, reason: reason || `Updated via SPortal Document Vault` }
      );
      const updated = res?.data || res;
      setInspectDoc(updated);
      setActionSuccessMsg(`Document #${inspectDoc.id} status updated to ${newStatus}`);
      // Refresh list
      loadDocuments();
    } catch (err) {
      console.error('Failed to update document status:', err);
      setActionErrorMsg(err?.message || 'Failed to update document status');
    } finally {
      setActionLoading(false);
    }
  };

  // Download Action
  const handleDownload = async (doc) => {
    try {
      const res = await sportalService.downloadCustomerDocument(doc.org_id, doc.id);
      const blob = new Blob([res], { type: doc.mime_type || 'application/pdf' });
      const url = window.URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = url;
      link.download = doc.file_name || doc.original_file_name || `document_${doc.id}.pdf`;
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      window.URL.revokeObjectURL(url);
    } catch (err) {
      console.error('Failed to download document:', err);
      alert('Failed to download document. Please check permissions.');
    }
  };

  // Helpers
  const formatBytes = (bytes) => {
    if (!bytes || bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i];
  };

  const formatDate = (dStr) => {
    if (!dStr) return '—';
    try {
      const d = new Date(dStr);
      return isNaN(d.getTime()) ? '—' : d.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' });
    } catch {
      return '—';
    }
  };

  const currentOrgName = useMemo(() => {
    if (selectedOrgId === 'all') return 'All Organizations (Platform-Wide)';
    const found = organizations.find((o) => String(o.id) === String(selectedOrgId));
    return found ? found.name : `Organization #${selectedOrgId}`;
  }, [organizations, selectedOrgId]);

  // Derived KPI values
  const totalCount = summary?.total_documents ?? documents.length;
  const verifiedCount = summary?.verified_count ?? documents.filter((d) => d.status === 'VERIFIED').length;
  const discrepancyCount = summary?.discrepancies_count ?? documents.filter((d) => d.status === 'DISCREPANCY' || d.has_discrepancy).length;
  const expiringCount = (summary?.expiring_soon_count ?? 0) + (summary?.expired_count ?? 0);

  return (
    <div className="p-6 sm:p-8 max-w-7xl mx-auto space-y-6 animate-fade-in">
      {/* Standard Page Header */}
      <PageHeader
        breadcrumbs={[{ label: 'SPortal', href: '/' }, { label: 'Documents & Compliance' }]}
        title="Documents & Compliance Repository"
        description="Authoritative oversight of customer freight documentation, KYC/KYB records, Textract OCR parsing, contract validation, and audit governance."
        badge={{ label: 'Document Repository', variant: 'info' }}
        secondaryAction={{
          label: 'Refresh Records',
          icon: RefreshCw,
          onClick: loadDocuments,
        }}
      />

      {/* Organization Switcher & Context Bar */}
      <div className="rounded-xl border border-slate-200 bg-white p-4 shadow-xs flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div className="flex items-center gap-3">
          <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-navy-900 text-white font-bold">
            <Building2 className="h-5 w-5" />
          </div>
          <div>
            <span className="text-[11px] font-bold text-slate-400 uppercase tracking-wider block">Customer Scope</span>
            <div className="flex items-center gap-2">
              <span className="text-sm font-bold text-slate-900">{currentOrgName}</span>
              {selectedOrgId !== 'all' && (
                <Link
                  to={`/organizations/${selectedOrgId}#tab-documents`}
                  className="inline-flex items-center gap-1 text-[11px] font-bold text-blue-600 hover:underline"
                >
                  <span>Customer 360</span>
                  <ExternalLink className="h-3 w-3" />
                </Link>
              )}
            </div>
          </div>
        </div>

        {/* Organization Select Dropdown */}
        <div className="flex items-center gap-2.5">
          <label htmlFor="customer-org-select" className="text-xs font-semibold text-slate-600 whitespace-nowrap">
            Switch Customer:
          </label>
          <select
            id="customer-org-select"
            value={selectedOrgId}
            onChange={(e) => setSelectedOrgId(e.target.value)}
            className="rounded-lg border border-slate-300 bg-slate-50 px-3 py-1.5 text-xs font-semibold text-slate-800 focus:border-blue-500 focus:bg-white focus:outline-none"
          >
            <option value="2">Org #2 - LogisticsHQ Dev Org - Varun Logistics</option>
            <option value="1">Org #1 - Freel Global Logistics Pvt Ltd</option>
            <option value="all">🌐 All Organizations (Platform Aggregate)</option>
            <optgroup label="Registered Customer Accounts">
              {organizations
                .filter((o) => o.id !== 1 && o.id !== 2)
                .map((o) => (
                  <option key={o.id} value={String(o.id)}>
                    Org #{o.id} - {o.name}
                  </option>
                ))}
            </optgroup>
          </select>
        </div>
      </div>

      {/* KPI Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <KpiCard
          title="Total Documents Stored"
          value={selectedOrgId === 'all' ? platformOverview?.total_documents ?? 9 : totalCount}
          unit="files"
          subtitle="Indexed in encrypted S3 repository"
          icon={HardDrive}
          variant="primary"
        />
        <KpiCard
          title="Verified Compliance"
          value={selectedOrgId === 'all' ? platformOverview?.verified_documents ?? 7 : verifiedCount}
          unit="verified"
          subtitle="Passed legal & OCR verification"
          icon={FileCheck}
          variant="success"
        />
        <KpiCard
          title="Discrepancies & Attention"
          value={selectedOrgId === 'all' ? platformOverview?.discrepancy_documents ?? 1 : discrepancyCount}
          unit="alerts"
          subtitle="Validation mismatches requiring review"
          icon={AlertTriangle}
          variant={discrepancyCount > 0 ? 'warning' : 'default'}
        />
        <KpiCard
          title="Expiring / Expired"
          value={selectedOrgId === 'all' ? platformOverview?.expiring_documents ?? 0 : expiringCount}
          unit="documents"
          subtitle="Within 30-day renewal threshold"
          icon={Clock}
          variant={expiringCount > 0 ? 'warning' : 'default'}
        />
      </div>

      {/* Main Document Repository Section */}
      {selectedOrgId === 'all' ? (
        /* Platform Overview View */
        <div className="rounded-xl border border-slate-200 bg-white shadow-xs p-6 space-y-6">
          <div className="border-b border-slate-100 pb-4">
            <h2 className="text-sm font-bold text-slate-900">Platform-Wide Tenant Vault Summary</h2>
            <p className="text-xs text-slate-500">Global distribution of customer documentation across all active business accounts.</p>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <div className="p-4 rounded-xl border border-slate-200 bg-slate-50">
              <span className="text-xs font-bold text-slate-500 block">Total Active Tenant Accounts</span>
              <p className="text-2xl font-black text-slate-900 mt-1">{platformOverview?.total_organizations ?? organizations.length}</p>
            </div>
            <div className="p-4 rounded-xl border border-slate-200 bg-slate-50">
              <span className="text-xs font-bold text-slate-500 block">Average Compliance Score</span>
              <p className="text-2xl font-black text-emerald-600 mt-1">{platformOverview?.average_compliance_score ?? 100}%</p>
            </div>
            <div className="p-4 rounded-xl border border-slate-200 bg-slate-50">
              <span className="text-xs font-bold text-slate-500 block">Active Contract Cover</span>
              <p className="text-2xl font-black text-blue-600 mt-1">{platformOverview?.total_active_contracts ?? 8}</p>
            </div>
          </div>

          {platformOverview?.tenant_summaries && platformOverview.tenant_summaries.length > 0 && (
            <div className="overflow-x-auto">
              <table className="w-full text-left text-xs">
                <thead>
                  <tr className="border-b border-slate-200 bg-slate-50/75 text-slate-500 font-semibold uppercase tracking-wider text-[11px]">
                    <th className="py-3 px-3">Organization</th>
                    <th className="py-3 px-3">Total Docs</th>
                    <th className="py-3 px-3">Verified</th>
                    <th className="py-3 px-3">Discrepancies</th>
                    <th className="py-3 px-3">Contracts</th>
                    <th className="py-3 px-3">Score</th>
                    <th className="py-3 px-3 text-right">Actions</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-slate-100">
                  {platformOverview.tenant_summaries.map((ts) => (
                    <tr key={ts.organization_id} className="hover:bg-slate-50/50">
                      <td className="py-3 px-3 font-semibold text-slate-900">
                        {ts.organization_name} <span className="font-mono text-slate-400">#{ts.organization_id}</span>
                      </td>
                      <td className="py-3 px-3 font-semibold text-slate-800">{ts.total_documents}</td>
                      <td className="py-3 px-3 text-emerald-600 font-medium">{ts.verified_count}</td>
                      <td className="py-3 px-3 text-amber-600 font-medium">{ts.discrepancies_count}</td>
                      <td className="py-3 px-3 text-slate-700">{ts.total_contracts}</td>
                      <td className="py-3 px-3 font-bold text-slate-900">{ts.compliance_score}%</td>
                      <td className="py-3 px-3 text-right">
                        <button
                          type="button"
                          onClick={() => setSelectedOrgId(String(ts.organization_id))}
                          className="inline-flex items-center gap-1 rounded bg-slate-100 px-2.5 py-1 text-xs font-semibold text-slate-700 hover:bg-slate-200"
                        >
                          <span>Open Vault</span>
                          <ChevronRight className="h-3 w-3" />
                        </button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>
      ) : (
        /* Customer Specific Document Vault */
        <div className="bg-white rounded-xl border border-slate-200 shadow-xs overflow-hidden">
          {/* Controls Bar */}
          <div className="p-4 border-b border-slate-200 flex flex-col sm:flex-row sm:items-center justify-between gap-3 bg-slate-50/50">
            <div className="flex items-center gap-2">
              <h2 className="text-sm font-bold text-slate-900">Customer Document Records</h2>
              <span className="text-xs font-semibold text-slate-500 bg-slate-200/70 px-2 py-0.5 rounded-full">
                {filteredDocs.length} Records
              </span>
            </div>

            <div className="flex flex-wrap items-center gap-2.5">
              {/* Search */}
              <div className="relative">
                <Search className="w-3.5 h-3.5 text-slate-400 absolute left-2.5 top-1/2 -translate-y-1/2" />
                <input
                  type="text"
                  placeholder="Search file name, type, reference..."
                  value={search}
                  onChange={(e) => setSearch(e.target.value)}
                  className="pl-8 pr-3 py-1.5 text-xs bg-white border border-slate-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-600 text-slate-800 w-44 sm:w-60"
                />
              </div>

              {/* Status Filter Tabs */}
              <div className="flex items-center bg-slate-200/60 p-0.5 rounded-lg text-xs">
                <button
                  type="button"
                  onClick={() => setStatusFilter('ALL')}
                  className={`px-2.5 py-1 rounded-md font-medium transition-all ${
                    statusFilter === 'ALL' ? 'bg-white text-slate-900 shadow-2xs font-semibold' : 'text-slate-600 hover:text-slate-900'
                  }`}
                >
                  All
                </button>
                <button
                  type="button"
                  onClick={() => setStatusFilter('VERIFIED')}
                  className={`px-2.5 py-1 rounded-md font-medium transition-all ${
                    statusFilter === 'VERIFIED' ? 'bg-white text-emerald-700 shadow-2xs font-semibold' : 'text-slate-600 hover:text-slate-900'
                  }`}
                >
                  Verified
                </button>
                <button
                  type="button"
                  onClick={() => setStatusFilter('PENDING_REVIEW')}
                  className={`px-2.5 py-1 rounded-md font-medium transition-all ${
                    statusFilter === 'PENDING_REVIEW' ? 'bg-white text-blue-700 shadow-2xs font-semibold' : 'text-slate-600 hover:text-slate-900'
                  }`}
                >
                  Pending
                </button>
                <button
                  type="button"
                  onClick={() => setStatusFilter('DISCREPANCY')}
                  className={`px-2.5 py-1 rounded-md font-medium transition-all ${
                    statusFilter === 'DISCREPANCY' ? 'bg-white text-amber-700 shadow-2xs font-semibold' : 'text-slate-600 hover:text-slate-900'
                  }`}
                >
                  Discrepancies
                </button>
              </div>

              {/* Document Type Dropdown */}
              <select
                value={typeFilter}
                onChange={(e) => setTypeFilter(e.target.value)}
                className="rounded-lg border border-slate-200 bg-white px-2.5 py-1.5 text-xs font-medium text-slate-700 focus:outline-none focus:ring-2 focus:ring-blue-600"
              >
                <option value="ALL">All Document Types</option>
                <option value="BILL_OF_LADING">Bill of Lading</option>
                <option value="HBL">House Bill of Lading (HBL)</option>
                <option value="MBL">Master Bill of Lading (MBL)</option>
                <option value="COMMERCIAL_INVOICE">Commercial Invoice</option>
                <option value="PACKING_LIST">Packing List</option>
                <option value="CERTIFICATE_OF_ORIGIN">Certificate of Origin</option>
                <option value="KYC">KYC / Identification</option>
                <option value="INSURANCE">Insurance Certificate</option>
              </select>
            </div>
          </div>

          {/* Content */}
          {loading ? (
            <div className="py-16 text-center text-slate-400 text-xs">
              <RefreshCw className="w-6 h-6 animate-spin mx-auto mb-2 text-blue-600" />
              Loading customer documents from database...
            </div>
          ) : error ? (
            <div className="p-8 text-center">
              <AlertCircle className="w-8 h-8 text-rose-500 mx-auto mb-2" />
              <p className="text-sm font-semibold text-slate-900">Failed to load documents</p>
              <p className="text-xs text-slate-500 mt-1">{error}</p>
              <button
                type="button"
                onClick={loadDocuments}
                className="mt-3 px-3 py-1.5 bg-slate-900 text-white text-xs font-medium rounded-lg"
              >
                Retry
              </button>
            </div>
          ) : filteredDocs.length === 0 ? (
            <EmptyState
              title="No documents found"
              description={`There are currently no documents matching your criteria for ${currentOrgName}.`}
              primaryAction={{
                label: 'Clear Search Filters',
                onClick: () => {
                  setSearch('');
                  setTypeFilter('ALL');
                  setStatusFilter('ALL');
                  setExpiryFilter('ALL');
                },
              }}
            />
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full text-left border-collapse text-xs">
                <thead>
                  <tr className="border-b border-slate-200 bg-slate-50 text-slate-600 font-semibold uppercase tracking-wider text-[11px]">
                    <th className="py-3 px-4">Document Title / File</th>
                    <th className="py-3 px-4">Type & Category</th>
                    <th className="py-3 px-4">Associated Entity</th>
                    <th className="py-3 px-4">Status</th>
                    <th className="py-3 px-4">Expiry Date</th>
                    <th className="py-3 px-4">File Size</th>
                    <th className="py-3 px-4">Uploaded</th>
                    <th className="py-3 px-4 text-right">Actions</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-slate-100 text-slate-700">
                  {filteredDocs.map((doc) => {
                    const hasDiscrepancy = doc.status === 'DISCREPANCY' || doc.has_discrepancy;
                    return (
                      <tr key={doc.id} className="hover:bg-slate-50/80 transition-colors">
                        {/* File Name */}
                        <td className="py-3.5 px-4 font-medium text-slate-900">
                          <div className="flex items-center gap-2.5">
                            <div className={`w-8 h-8 rounded-lg flex items-center justify-center font-bold text-xs border ${
                              hasDiscrepancy
                                ? 'bg-amber-50 text-amber-700 border-amber-200'
                                : doc.status === 'VERIFIED'
                                ? 'bg-emerald-50 text-emerald-700 border-emerald-200'
                                : 'bg-blue-50 text-blue-700 border-blue-200'
                            }`}>
                              <FileText className="w-4 h-4" />
                            </div>
                            <div>
                              <div className="font-semibold text-slate-900">
                                {doc.original_file_name || doc.file_name || doc.document_name || `Document_${doc.id}.pdf`}
                              </div>
                              <div className="text-[11px] text-slate-400 font-mono">ID: #{doc.id}</div>
                            </div>
                          </div>
                        </td>

                        {/* Type & Category */}
                        <td className="py-3.5 px-4">
                          <span className="px-2 py-0.5 rounded-md font-semibold text-[11px] bg-slate-100 text-slate-700 uppercase">
                            {doc.doc_type ? doc.doc_type.replace(/_/g, ' ') : 'PDF DOCUMENT'}
                          </span>
                        </td>

                        {/* Associated Entity */}
                        <td className="py-3.5 px-4 text-slate-600">
                          {doc.shipment_ref ? (
                            <span className="rounded bg-sky-50 border border-sky-200 px-1.5 py-0.5 font-mono text-[10px] font-bold text-sky-700">
                              {doc.shipment_ref}
                            </span>
                          ) : doc.booking_ref ? (
                            <span className="rounded bg-indigo-50 border border-indigo-200 px-1.5 py-0.5 font-mono text-[10px] font-bold text-indigo-700">
                              {doc.booking_ref}
                            </span>
                          ) : (
                            <span className="text-slate-400">—</span>
                          )}
                        </td>

                        {/* Status */}
                        <td className="py-3.5 px-4">
                          <span className={`inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-[10px] font-bold border ${
                            doc.status === 'VERIFIED'
                              ? 'bg-emerald-50 text-emerald-700 border-emerald-200'
                              : doc.status === 'DISCREPANCY'
                              ? 'bg-amber-50 text-amber-700 border-amber-200'
                              : doc.status === 'REJECTED'
                              ? 'bg-rose-50 text-rose-700 border-rose-200'
                              : 'bg-blue-50 text-blue-700 border-blue-200'
                          }`}>
                            {doc.status === 'VERIFIED' && <CheckCircle2 className="w-2.5 h-2.5" />}
                            {doc.status === 'DISCREPANCY' && <AlertTriangle className="w-2.5 h-2.5" />}
                            {doc.status === 'REJECTED' && <XCircle className="w-2.5 h-2.5" />}
                            {doc.status || 'PENDING_REVIEW'}
                          </span>
                        </td>

                        {/* Expiry Date */}
                        <td className="py-3.5 px-4 font-mono text-[11px] text-slate-600">
                          {doc.expires_at ? (
                            <span className={doc.is_expired ? 'text-rose-600 font-bold' : doc.is_expiring_soon ? 'text-amber-600 font-bold' : ''}>
                              {formatDate(doc.expires_at)}
                              {doc.is_expiring_soon && ' (Expiring)'}
                              {doc.is_expired && ' (Expired)'}
                            </span>
                          ) : (
                            <span className="text-slate-400">Standard / No Expiry</span>
                          )}
                        </td>

                        {/* File Size */}
                        <td className="py-3.5 px-4 font-mono text-[11px] text-slate-600">
                          {formatBytes(doc.file_size)}
                        </td>

                        {/* Upload Date */}
                        <td className="py-3.5 px-4 font-mono text-[11px] text-slate-600">
                          {formatDate(doc.created_at)}
                        </td>

                        {/* Actions */}
                        <td className="py-3.5 px-4 text-right">
                          <div className="flex items-center justify-end gap-1.5">
                            <button
                              type="button"
                              onClick={() => handleOpenInspector(doc)}
                              className="inline-flex items-center gap-1 px-2.5 py-1 rounded-md bg-white border border-slate-200 text-slate-700 hover:bg-slate-50 text-xs font-medium transition-colors shadow-2xs"
                              title="Inspect OCR & Metadata"
                            >
                              <Eye className="w-3.5 h-3.5 text-slate-500" />
                              <span>Inspect</span>
                            </button>
                            <button
                              type="button"
                              onClick={() => handleDownload(doc)}
                              className="inline-flex items-center gap-1 px-2 py-1 rounded-md bg-white border border-slate-200 text-slate-700 hover:bg-slate-50 text-xs font-medium transition-colors shadow-2xs"
                              title="Download File"
                            >
                              <Download className="w-3.5 h-3.5 text-slate-500" />
                            </button>
                          </div>
                        </td>
                      </tr>
                    );
                  })}
                </tbody>
              </table>
            </div>
          )}
        </div>
      )}

      {/* ========================================================================= */}
      {/* DOCUMENT INSPECTION & OCR MODAL DRAWER */}
      {/* ========================================================================= */}
      {inspectDoc && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/50 backdrop-blur-xs p-4 overflow-y-auto animate-fade-in">
          <div className="relative w-full max-w-3xl rounded-2xl border border-slate-200 bg-white shadow-2xl overflow-hidden my-8">
            {/* Header */}
            <div className="flex items-center justify-between border-b border-slate-200 bg-slate-50/80 px-6 py-4">
              <div className="flex items-center gap-3">
                <div className="flex h-9 w-9 items-center justify-center rounded-xl bg-blue-50 text-blue-600 border border-blue-100">
                  <FileText className="h-5 w-5" />
                </div>
                <div>
                  <h3 className="text-base font-bold text-slate-900">
                    {inspectDoc.original_file_name || inspectDoc.file_name || `Document #${inspectDoc.id}`}
                  </h3>
                  <div className="flex items-center gap-2 text-xs text-slate-500 font-mono">
                    <span>ID: #{inspectDoc.id}</span>
                    <span>•</span>
                    <span className="uppercase font-semibold">{inspectDoc.doc_type || 'DOCUMENT'}</span>
                    <span>•</span>
                    <span>{inspectDoc.status}</span>
                  </div>
                </div>
              </div>

              <div className="flex items-center gap-2">
                <button
                  type="button"
                  onClick={() => handleDownload(inspectDoc)}
                  className="inline-flex items-center gap-1 px-3 py-1.5 rounded-lg border border-slate-200 bg-white text-xs font-semibold text-slate-700 hover:bg-slate-50 shadow-2xs"
                >
                  <Download className="w-3.5 h-3.5 text-slate-500" />
                  <span>Download</span>
                </button>
                <button
                  type="button"
                  onClick={() => setInspectDoc(null)}
                  className="rounded-lg p-1.5 text-slate-400 hover:bg-slate-100 hover:text-slate-600"
                >
                  <X className="h-5 w-5" />
                </button>
              </div>
            </div>

            {/* Notification messages */}
            {actionSuccessMsg && (
              <div className="bg-emerald-50 border-b border-emerald-200 px-6 py-2.5 text-xs font-semibold text-emerald-800 flex items-center gap-2">
                <Check className="h-4 w-4 text-emerald-600" />
                <span>{actionSuccessMsg}</span>
              </div>
            )}
            {actionErrorMsg && (
              <div className="bg-rose-50 border-b border-rose-200 px-6 py-2.5 text-xs font-semibold text-rose-800 flex items-center gap-2">
                <AlertCircle className="h-4 w-4 text-rose-600" />
                <span>{actionErrorMsg}</span>
              </div>
            )}

            {/* Verification Action Bar */}
            <div className="bg-slate-100/70 border-b border-slate-200 px-6 py-3 flex flex-wrap items-center justify-between gap-3">
              <div className="flex items-center gap-2">
                <span className="text-xs font-bold text-slate-700">Verification State:</span>
                <span className={`rounded-full px-2.5 py-0.5 text-xs font-bold border ${
                  inspectDoc.status === 'VERIFIED'
                    ? 'bg-emerald-50 text-emerald-700 border-emerald-200'
                    : inspectDoc.status === 'DISCREPANCY'
                    ? 'bg-amber-50 text-amber-700 border-amber-200'
                    : inspectDoc.status === 'REJECTED'
                    ? 'bg-rose-50 text-rose-700 border-rose-200'
                    : 'bg-blue-50 text-blue-700 border-blue-200'
                }`}>
                  {inspectDoc.status || 'PENDING_REVIEW'}
                </span>
              </div>

              <div className="flex items-center gap-2">
                {inspectDoc.status !== 'VERIFIED' && (
                  <button
                    type="button"
                    disabled={actionLoading}
                    onClick={() => handleUpdateStatus('VERIFIED', 'Verified by Customs Administrator in SPortal')}
                    className="inline-flex items-center gap-1 rounded-lg bg-emerald-600 px-3 py-1.5 text-xs font-bold text-white hover:bg-emerald-700 transition-colors shadow-2xs disabled:opacity-50"
                  >
                    <CheckCircle2 className="h-3.5 w-3.5" />
                    <span>Approve & Verify</span>
                  </button>
                )}
                {inspectDoc.status !== 'REJECTED' && (
                  <button
                    type="button"
                    disabled={actionLoading}
                    onClick={() => {
                      const reason = prompt('Enter reason for document rejection:');
                      if (reason !== null) {
                        handleUpdateStatus('REJECTED', reason || 'Rejected by compliance admin');
                      }
                    }}
                    className="inline-flex items-center gap-1 rounded-lg border border-rose-200 bg-rose-50 px-3 py-1.5 text-xs font-bold text-rose-700 hover:bg-rose-100 transition-colors shadow-2xs disabled:opacity-50"
                  >
                    <XCircle className="h-3.5 w-3.5 text-rose-600" />
                    <span>Reject Document</span>
                  </button>
                )}
              </div>
            </div>

            {/* Tabs Bar */}
            <div className="border-b border-slate-200 px-6 bg-white">
              <nav className="flex space-x-6">
                <button
                  type="button"
                  onClick={() => setActiveInspectTab('metadata')}
                  className={`py-3 text-xs font-semibold border-b-2 transition-colors ${
                    activeInspectTab === 'metadata'
                      ? 'border-blue-600 text-blue-600'
                      : 'border-transparent text-slate-500 hover:text-slate-700'
                  }`}
                >
                  Metadata & Vault Details
                </button>
                <button
                  type="button"
                  onClick={() => setActiveInspectTab('extracted')}
                  className={`py-3 text-xs font-semibold border-b-2 transition-colors ${
                    activeInspectTab === 'extracted'
                      ? 'border-blue-600 text-blue-600'
                      : 'border-transparent text-slate-500 hover:text-slate-700'
                  }`}
                >
                  Extracted Data (OCR)
                </button>
                <button
                  type="button"
                  onClick={() => setActiveInspectTab('raw_ocr')}
                  className={`py-3 text-xs font-semibold border-b-2 transition-colors ${
                    activeInspectTab === 'raw_ocr'
                      ? 'border-blue-600 text-blue-600'
                      : 'border-transparent text-slate-500 hover:text-slate-700'
                  }`}
                >
                  Raw OCR Stream
                </button>
                <button
                  type="button"
                  onClick={() => setActiveInspectTab('audit')}
                  className={`py-3 text-xs font-semibold border-b-2 transition-colors ${
                    activeInspectTab === 'audit'
                      ? 'border-blue-600 text-blue-600'
                      : 'border-transparent text-slate-500 hover:text-slate-700'
                  }`}
                >
                  Audit & Discrepancies
                </button>
              </nav>
            </div>

            {/* Modal Body */}
            <div className="p-6 max-h-[60vh] overflow-y-auto space-y-4">
              {inspectLoading ? (
                <div className="py-12 text-center text-xs text-slate-400">
                  <RefreshCw className="w-5 h-5 animate-spin mx-auto mb-2 text-blue-600" />
                  Loading document details...
                </div>
              ) : activeInspectTab === 'metadata' ? (
                <div className="space-y-4">
                  <div className="grid grid-cols-2 gap-3 text-xs">
                    <div className="p-3 rounded-lg border border-slate-100 bg-slate-50">
                      <span className="text-slate-400 font-medium block">Document Type</span>
                      <span className="font-semibold text-slate-900 mt-0.5 block">{inspectDoc.doc_type || 'OTHER'}</span>
                    </div>
                    <div className="p-3 rounded-lg border border-slate-100 bg-slate-50">
                      <span className="text-slate-400 font-medium block">File Size</span>
                      <span className="font-semibold text-slate-900 mt-0.5 block">{formatBytes(inspectDoc.file_size)}</span>
                    </div>
                    <div className="p-3 rounded-lg border border-slate-100 bg-slate-50">
                      <span className="text-slate-400 font-medium block">Storage Vault</span>
                      <span className="font-semibold text-emerald-700 mt-0.5 block">Amazon S3 Encrypted Bucket</span>
                    </div>
                    <div className="p-3 rounded-lg border border-slate-100 bg-slate-50">
                      <span className="text-slate-400 font-medium block">MIME Type</span>
                      <span className="font-semibold text-slate-900 mt-0.5 block font-mono">{inspectDoc.mime_type || 'application/pdf'}</span>
                    </div>
                    <div className="p-3 rounded-lg border border-slate-100 bg-slate-50">
                      <span className="text-slate-400 font-medium block">Associated Shipment</span>
                      <span className="font-semibold text-slate-900 mt-0.5 block font-mono">{inspectDoc.shipment_ref || '—'}</span>
                    </div>
                    <div className="p-3 rounded-lg border border-slate-100 bg-slate-50">
                      <span className="text-slate-400 font-medium block">Associated Booking</span>
                      <span className="font-semibold text-slate-900 mt-0.5 block font-mono">{inspectDoc.booking_ref || '—'}</span>
                    </div>
                    <div className="p-3 rounded-lg border border-slate-100 bg-slate-50">
                      <span className="text-slate-400 font-medium block">Created Timestamp</span>
                      <span className="font-semibold text-slate-900 mt-0.5 block font-mono">{formatDate(inspectDoc.created_at)}</span>
                    </div>
                    <div className="p-3 rounded-lg border border-slate-100 bg-slate-50">
                      <span className="text-slate-400 font-medium block">Expiry Timestamp</span>
                      <span className="font-semibold text-slate-900 mt-0.5 block font-mono">{formatDate(inspectDoc.expires_at)}</span>
                    </div>
                  </div>

                  {inspectDoc.ai_summary && (
                    <div className="p-3.5 rounded-xl border border-purple-100 bg-purple-50/50">
                      <div className="flex items-center gap-1.5 text-xs font-bold text-purple-900 mb-1">
                        <Sparkles className="h-3.5 w-3.5 text-purple-600" />
                        <span>Intelligence Assessment</span>
                      </div>
                      <p className="text-xs text-purple-800">{inspectDoc.ai_summary}</p>
                    </div>
                  )}
                </div>
              ) : activeInspectTab === 'extracted' ? (
                <div>
                  <h4 className="text-xs font-bold text-slate-700 uppercase tracking-wider mb-2">
                    Textract Structured Extraction
                  </h4>
                  {inspectDoc.extracted_data && Object.keys(inspectDoc.extracted_data).length > 0 ? (
                    <div className="border border-slate-200 rounded-xl overflow-hidden">
                      <table className="w-full text-left text-xs">
                        <thead className="bg-slate-50 border-b border-slate-200 text-slate-500 font-semibold uppercase text-[10px]">
                          <tr>
                            <th className="py-2.5 px-3">Field Key</th>
                            <th className="py-2.5 px-3">Extracted Entity Value</th>
                          </tr>
                        </thead>
                        <tbody className="divide-y divide-slate-100">
                          {Object.entries(inspectDoc.extracted_data).map(([key, val]) => (
                            <tr key={key} className="hover:bg-slate-50/50">
                              <td className="py-2 px-3 font-mono font-semibold text-slate-700">{key}</td>
                              <td className="py-2 px-3 font-mono text-slate-900">
                                {typeof val === 'object' ? JSON.stringify(val) : String(val)}
                              </td>
                            </tr>
                          ))}
                        </tbody>
                      </table>
                    </div>
                  ) : (
                    <div className="p-6 text-center text-xs text-slate-500 bg-slate-50 rounded-xl border border-slate-100">
                      No structured fields extracted for this document.
                    </div>
                  )}
                </div>
              ) : activeInspectTab === 'raw_ocr' ? (
                <div>
                  <h4 className="text-xs font-bold text-slate-700 uppercase tracking-wider mb-2">
                    Raw OCR Unstructured Text Stream
                  </h4>
                  {inspectDoc.raw_ocr_text ? (
                    <pre className="p-4 bg-slate-900 text-slate-100 rounded-xl text-xs font-mono overflow-x-auto whitespace-pre-wrap max-h-60">
                      {inspectDoc.raw_ocr_text}
                    </pre>
                  ) : (
                    <div className="p-6 text-center text-xs text-slate-500 bg-slate-50 rounded-xl border border-slate-100">
                      No raw OCR stream available for this file.
                    </div>
                  )}
                </div>
              ) : (
                <div className="space-y-4">
                  {inspectDoc.discrepancies && inspectDoc.discrepancies.length > 0 ? (
                    <div className="space-y-2">
                      <h4 className="text-xs font-bold text-amber-900 uppercase tracking-wider">
                        Document Discrepancies ({inspectDoc.discrepancies.length})
                      </h4>
                      {inspectDoc.discrepancies.map((disc) => (
                        <div key={disc.id} className="p-3 rounded-lg border border-amber-200 bg-amber-50 text-xs">
                          <div className="flex items-center justify-between">
                            <span className="font-bold text-amber-900">{disc.discrepancy_type}</span>
                            <span className="rounded bg-amber-200 px-1.5 py-0.5 text-[10px] font-bold text-amber-800">
                              {disc.severity}
                            </span>
                          </div>
                          <p className="text-amber-800 mt-1">{disc.description}</p>
                          {disc.field && (
                            <div className="mt-2 grid grid-cols-2 gap-2 text-[11px] font-mono">
                              <span className="text-slate-600">Expected: {disc.expected_value || 'N/A'}</span>
                              <span className="text-slate-900 font-bold">Actual: {disc.actual_value || 'N/A'}</span>
                            </div>
                          )}
                        </div>
                      ))}
                    </div>
                  ) : (
                    <div className="p-4 rounded-xl border border-emerald-100 bg-emerald-50/60 text-xs text-emerald-800 flex items-center gap-2">
                      <CheckCircle2 className="h-4 w-4 text-emerald-600" />
                      <span>No validation discrepancies recorded for this document.</span>
                    </div>
                  )}

                  {/* Audit Timeline */}
                  <div>
                    <h4 className="text-xs font-bold text-slate-700 uppercase tracking-wider mb-2">
                      Governance Audit History
                    </h4>
                    {inspectDoc.audit_log && inspectDoc.audit_log.length > 0 ? (
                      <div className="divide-y divide-slate-100 border border-slate-200 rounded-xl">
                        {inspectDoc.audit_log.map((item, idx) => (
                          <div key={idx} className="p-3 text-xs flex items-start justify-between">
                            <div>
                              <span className="font-bold text-slate-900">{item.action}</span>
                              <p className="text-slate-600 mt-0.5">{item.details}</p>
                              <span className="text-[10px] text-slate-400">Actor: {item.actor_name}</span>
                            </div>
                            <span className="text-[11px] text-slate-400 font-mono">
                              {formatDate(item.timestamp)}
                            </span>
                          </div>
                        ))}
                      </div>
                    ) : (
                      <div className="p-4 rounded-xl border border-slate-100 bg-slate-50 text-xs text-slate-500 text-center">
                        Standard upload recorded.
                      </div>
                    )}
                  </div>
                </div>
              )}
            </div>

            {/* Footer */}
            <div className="border-t border-slate-200 bg-slate-50 px-6 py-3 flex justify-end">
              <button
                type="button"
                onClick={() => setInspectDoc(null)}
                className="rounded-lg border border-slate-300 bg-white px-4 py-2 text-xs font-semibold text-slate-700 hover:bg-slate-100 shadow-2xs"
              >
                Close Inspector
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

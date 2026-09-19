import { useState, useEffect, useCallback } from 'react';
import { Link, useSearchParams } from 'react-router-dom';
import {
  Compass,
  RefreshCw,
  Search,
  Filter,
  CheckCircle2,
  AlertTriangle,
  Clock,
  ExternalLink,
  User,
  ShieldCheck,
  X,
  FileCheck,
  Check,
  ChevronRight,
  UserCheck,
  Info,
  Mail,
  ListTodo,
  Send,
  Edit3,
  Eye,
} from 'lucide-react';
import { recommendationService } from '../../../services/recommendationService';
import AIConfidenceIndicator from '../../../components/ai/AIConfidenceIndicator';
import AIEmptyState from '../../../components/ai/AIEmptyState';
import AIErrorState from '../../../components/ai/AIErrorState';
import AILoadingState from '../../../components/ai/AILoadingState';
import AIEvidenceList from '../../../components/ai/AIEvidenceList';
import './RecommendationCenterPage.css';

export default function RecommendationCenterPage() {
  const [searchParams, setSearchParams] = useSearchParams();
  const initialShipmentId = searchParams.get('shipmentId') || searchParams.get('shipment_id') || '';
  const initialInvoiceId = searchParams.get('invoiceId') || searchParams.get('invoice_id') || '';
  const initialContractId = searchParams.get('contractId') || searchParams.get('contract_id') || '';
  const initialDocumentId = searchParams.get('documentId') || searchParams.get('document_id') || '';
  const initialComplianceId = searchParams.get('complianceId') || searchParams.get('compliance_id') || '';
  const initialCategory = searchParams.get('category') || '';
  const initialDelayedMilestones = searchParams.get('delayed_milestone') === 'true';
  const initialActiveExceptions = searchParams.get('active_exception') === 'true';
  const initialMissingOpsInfo = searchParams.get('missing_ops_info') === 'true';
  const initialOverdueOnly = searchParams.get('overdue_only') === 'true';
  const initialDataQualityOnly = searchParams.get('data_quality_only') === 'true';
  const initialUpcomingOnly = searchParams.get('upcoming_only') === 'true';
  const initialExpiredOnly = searchParams.get('expired_only') === 'true';
  const initialAutomationId = searchParams.get('automation_id') || '';
  const initialExecutionId = searchParams.get('execution_id') || '';

  const [recommendations, setRecommendations] = useState([]);
  const [stats, setStats] = useState(null);
  const [pagination, setPagination] = useState({ page: 1, limit: 10, total: 0, total_pages: 1 });
  const [loading, setLoading] = useState(true);
  const [generating, setGenerating] = useState(false);
  const [error, setError] = useState(null);

  // Filter states
  const [search, setSearch] = useState('');
  const [shipmentIdFilter, setShipmentIdFilter] = useState(initialShipmentId);
  const [invoiceIdFilter, setInvoiceIdFilter] = useState(initialInvoiceId);
  const [contractIdFilter, setContractIdFilter] = useState(initialContractId);
  const [documentIdFilter, setDocumentIdFilter] = useState(initialDocumentId);
  const [complianceIdFilter, setComplianceIdFilter] = useState(initialComplianceId);
  const [automationIdFilter, setAutomationIdFilter] = useState(initialAutomationId);
  const [executionIdFilter, setExecutionIdFilter] = useState(initialExecutionId);
  const [category, setCategory] = useState(initialCategory);
  const [priority, setPriority] = useState('');
  const [riskLevel, setRiskLevel] = useState('');
  const [status, setStatus] = useState('');
  const [followupType, setFollowupType] = useState('');
  const [requiresApproval, setRequiresApproval] = useState(false);
  const [missingInfoOnly, setMissingInfoOnly] = useState(false);
  const [expiringSoonOnly, setExpiringSoonOnly] = useState(false);
  const [pricingConcernOnly, setPricingConcernOnly] = useState(false);
  const [delayedMilestonesOnly, setDelayedMilestonesOnly] = useState(initialDelayedMilestones);
  const [activeExceptionsOnly, setActiveExceptionsOnly] = useState(initialActiveExceptions);
  const [missingOpsInfoOnly, setMissingOpsInfoOnly] = useState(initialMissingOpsInfo);
  const [overdueOnly, setOverdueOnly] = useState(initialOverdueOnly);
  const [dataQualityOnly, setDataQualityOnly] = useState(initialDataQualityOnly);
  const [upcomingOnly, setUpcomingOnly] = useState(initialUpcomingOnly);
  const [expiredOnly, setExpiredOnly] = useState(initialExpiredOnly);
  const [agingBandFilter, setAgingBandFilter] = useState('');
  const [sortBy, setSortBy] = useState('created_at');
  const [sortDir, setSortDir] = useState('desc');

  // Follow-Up stats
  const [followupStats, setFollowupStats] = useState(null);

  // Detail Drawer & Modal States
  const [selectedRec, setSelectedRec] = useState(null);
  const [dismissModalRec, setDismissModalRec] = useState(null);
  const [dismissReason, setDismissReason] = useState('');
  const [assignModalRec, setAssignModalRec] = useState(null);
  const [assigneeName, setAssigneeName] = useState('');
  const [actionSuccessMsg, setActionSuccessMsg] = useState(null);

  // Task 2.2 Draft & Task Modal States
  const [draftModalRec, setDraftModalRec] = useState(null);
  const [draftSubject, setDraftSubject] = useState('');
  const [draftBody, setDraftBody] = useState('');
  const [generatingDraft, setGeneratingDraft] = useState(false);
  const [savingDraft, setSavingDraft] = useState(false);

  const [taskModalRec, setTaskModalRec] = useState(null);
  const [taskAssigneeName, setTaskAssigneeName] = useState('');
  const [taskPriority, setTaskPriority] = useState('high');
  const [taskNotes, setTaskNotes] = useState('');
  const [taskDueDate, setTaskDueDate] = useState('');
  const [creatingTask, setCreatingTask] = useState(false);

  // Task 2.3 Action Preview & HITL Approval Modal States
  const [previewModalRec, setPreviewModalRec] = useState(null);
  const [actionPreviewData, setActionPreviewData] = useState(null);
  const [loadingPreview, setLoadingPreview] = useState(false);
  const [submittingApproval, setSubmittingApproval] = useState(false);
  const [approvalNotes, setApprovalNotes] = useState('');

  const fetchRecommendations = useCallback(async (page = 1) => {
    setLoading(true);
    setError(null);
    try {
      const res = await recommendationService.listRecommendations({
        page,
        limit: pagination.limit,
        category,
        priority,
        risk_level: riskLevel,
        status,
        search,
        sort_by: sortBy,
        sort_dir: sortDir,
        requires_approval: requiresApproval ? 'true' : '',
        followup_type: followupType,
        missing_info: missingInfoOnly,
        expiring_soon: expiringSoonOnly,
        pricing_concern: pricingConcernOnly,
        delayed_milestone: delayedMilestonesOnly ? 'true' : '',
        active_exception: activeExceptionsOnly ? 'true' : '',
        missing_ops_info: missingOpsInfoOnly ? 'true' : '',
        shipment_id: shipmentIdFilter || undefined,
        invoice_id: invoiceIdFilter || undefined,
        contract_id: contractIdFilter || undefined,
        document_id: documentIdFilter || undefined,
        compliance_id: complianceIdFilter || undefined,
        overdue_only: overdueOnly ? 'true' : undefined,
        data_quality_only: dataQualityOnly ? 'true' : undefined,
        upcoming_only: upcomingOnly ? 'true' : undefined,
        expired_only: expiredOnly ? 'true' : undefined,
        aging_band: agingBandFilter || undefined,
        automation_id: automationIdFilter || undefined,
        execution_id: executionIdFilter || undefined,
      });

      if (res?.success) {
        setRecommendations(res.recommendations || []);
        if (res.pagination) {
          setPagination(res.pagination);
        }
        if (res.stats) {
          setStats(res.stats);
        }
      }
    } catch (err) {
      setError(err.message || 'Failed to load recommendations.');
    } finally {
      setLoading(false);
    }
  }, [category, priority, riskLevel, status, search, sortBy, sortDir, requiresApproval, followupType, missingInfoOnly, expiringSoonOnly, pricingConcernOnly, delayedMilestonesOnly, activeExceptionsOnly, missingOpsInfoOnly, shipmentIdFilter, invoiceIdFilter, contractIdFilter, documentIdFilter, complianceIdFilter, overdueOnly, dataQualityOnly, upcomingOnly, expiredOnly, agingBandFilter, automationIdFilter, executionIdFilter, pagination.limit]);

  const fetchStats = useCallback(async () => {
    try {
      const [statsRes, followRes] = await Promise.all([
        recommendationService.getStats(),
        recommendationService.getFollowupStats().catch(() => null),
      ]);
      if (statsRes?.success && statsRes?.stats) {
        setStats(statsRes.stats);
      }
      if (followRes?.success && followRes?.stats) {
        setFollowupStats(followRes.stats);
      }
    } catch {
      // Non-blocking
    }
  }, []);

  useEffect(() => {
    fetchRecommendations(1);
    fetchStats();
  }, [fetchRecommendations, fetchStats]);

  const handleGenerate = async () => {
    setGenerating(true);
    setError(null);
    setActionSuccessMsg(null);
    try {
      const res = await recommendationService.generateRecommendations();
      if (res?.success) {
        const result = res.result;
        setActionSuccessMsg(`Generation cycle completed. Evaluated ${result.total_evaluated} records. (${result.created_count} new, ${result.updated_count} refreshed)`);
        await fetchRecommendations(1);
        await fetchStats();
      }
    } catch (err) {
      setError(err.message || 'Failed to run recommendation generation cycle.');
    } finally {
      setGenerating(false);
    }
  };

  const handleMarkReviewed = async (rec) => {
    try {
      const res = await recommendationService.markReviewed(rec.id);
      if (res?.success) {
        setActionSuccessMsg(`Recommendation ${rec.id} marked as reviewed.`);
        if (selectedRec?.id === rec.id) {
          setSelectedRec(res.recommendation);
        }
        await fetchRecommendations(pagination.page);
        await fetchStats();
      }
    } catch (err) {
      alert(`Review action failed: ${err.message}`);
    }
  };

  const handleConfirmAssign = async () => {
    if (!assignModalRec || !assigneeName.trim()) return;
    try {
      const res = await recommendationService.assignRecommendation(assignModalRec.id, 1, assigneeName.trim());
      if (res?.success) {
        setActionSuccessMsg(`Assigned recommendation ${assignModalRec.id} to ${assigneeName.trim()}.`);
        setAssignModalRec(null);
        setAssigneeName('');
        if (selectedRec?.id === assignModalRec.id) {
          setSelectedRec(res.recommendation);
        }
        await fetchRecommendations(pagination.page);
        await fetchStats();
      }
    } catch (err) {
      alert(`Assignment failed: ${err.message}`);
    }
  };

  const handleConfirmDismiss = async () => {
    if (!dismissModalRec || !dismissReason.trim()) return;
    try {
      const res = await recommendationService.dismissRecommendation(dismissModalRec.id, dismissReason.trim());
      if (res?.success) {
        setActionSuccessMsg(`Recommendation ${dismissModalRec.id} dismissed.`);
        setDismissModalRec(null);
        setDismissReason('');
        if (selectedRec?.id === dismissModalRec.id) {
          setSelectedRec(res.recommendation);
        }
        await fetchRecommendations(pagination.page);
        await fetchStats();
      }
    } catch (err) {
      alert(`Dismissal failed: ${err.message}`);
    }
  };

  // Explicit Draft Generation Handler (Task 2.2)
  const handleOpenDraftModal = async (rec) => {
    setDraftModalRec(rec);
    if (rec.draft_subject && rec.draft_body) {
      setDraftSubject(rec.draft_subject);
      setDraftBody(rec.draft_body);
    } else {
      setGeneratingDraft(true);
      try {
        const res = await recommendationService.generateDraft(rec.id);
        if (res?.success && res.recommendation) {
          setDraftSubject(res.recommendation.draft_subject || '');
          setDraftBody(res.recommendation.draft_body || '');
          setDraftModalRec(res.recommendation);
          await fetchRecommendations(pagination.page);
        }
      } catch (err) {
        alert(`Draft generation failed: ${err.message}`);
      } finally {
        setGeneratingDraft(false);
      }
    }
  };

  const handleSaveDraft = async () => {
    if (!draftModalRec) return;
    setSavingDraft(true);
    try {
      const res = await recommendationService.saveDraft(draftModalRec.id, draftSubject, draftBody);
      if (res?.success) {
        setActionSuccessMsg(`Draft saved for customer ${draftModalRec.customer_name || 'follow-up'}.`);
        setDraftModalRec(null);
        await fetchRecommendations(pagination.page);
      }
    } catch (err) {
      alert(`Save draft failed: ${err.message}`);
    } finally {
      setSavingDraft(false);
    }
  };

  // Controlled Follow-Up Task Creation Handler (Task 2.2)
  const handleOpenTaskModal = (rec) => {
    setTaskModalRec(rec);
    setTaskAssigneeName(rec.suggested_owner_name || rec.assignee_name || '');
    setTaskPriority(rec.priority || 'high');
    setTaskNotes('');
    setTaskDueDate('');
  };

  const handleConfirmCreateTask = async () => {
    if (!taskModalRec) return;
    setCreatingTask(true);
    try {
      const res = await recommendationService.createFollowupTask(taskModalRec.id, {
        assignee_name: taskAssigneeName.trim() || undefined,
        priority: taskPriority,
        notes: taskNotes,
        due_date: taskDueDate ? `${taskDueDate}T23:59:59Z` : undefined,
      });
      if (res?.success) {
        setActionSuccessMsg(`Follow-up task #${res.task?.id} created for ${taskModalRec.customer_name || 'customer'}.`);
        setTaskModalRec(null);
        await fetchRecommendations(pagination.page);
        await fetchStats();
      }
    } catch (err) {
      alert(`Task creation failed: ${err.message}`);
    } finally {
      setCreatingTask(false);
    }
  };

  // Controlled Action Preview & Approval Integration (Task 2.3)
  const handleOpenActionPreview = async (rec) => {
    setPreviewModalRec(rec);
    setActionPreviewData(null);
    setLoadingPreview(true);
    setApprovalNotes('');
    try {
      const res = await recommendationService.getActionPreview(rec.id);
      if (res?.success && res.action_preview) {
        setActionPreviewData(res.action_preview);
      }
    } catch (err) {
      alert(`Failed to load action preview: ${err.message}`);
    } finally {
      setLoadingPreview(false);
    }
  };

  const handleSubmitApproval = async () => {
    if (!previewModalRec) return;
    setSubmittingApproval(true);
    try {
      const res = await recommendationService.requestApproval(previewModalRec.id, approvalNotes.trim());
      if (res?.success) {
        setActionSuccessMsg(`Approval request #${res.recommendation?.approval_id || 'HITL'} submitted for recommendation #${previewModalRec.id}.`);
        if (actionPreviewData) {
          setActionPreviewData({
            ...actionPreviewData,
            approval_id: res.recommendation?.approval_id,
          });
        }
        await fetchRecommendations(pagination.page);
        await fetchStats();
      }
    } catch (err) {
      alert(`Failed to submit approval request: ${err.message}`);
    } finally {
      setSubmittingApproval(false);
    }
  };

  const resetFilters = () => {
    setSearch('');
    setShipmentIdFilter('');
    setInvoiceIdFilter('');
    setContractIdFilter('');
    setDocumentIdFilter('');
    setComplianceIdFilter('');
    setCategory('');
    setPriority('');
    setRiskLevel('');
    setStatus('');
    setFollowupType('');
    setRequiresApproval(false);
    setMissingInfoOnly(false);
    setExpiringSoonOnly(false);
    setPricingConcernOnly(false);
    setDelayedMilestonesOnly(false);
    setActiveExceptionsOnly(false);
    setMissingOpsInfoOnly(false);
    setOverdueOnly(false);
    setDataQualityOnly(false);
    setUpcomingOnly(false);
    setExpiredOnly(false);
    setAutomationIdFilter('');
    setExecutionIdFilter('');
    setAgingBandFilter('');
    setSortBy('created_at');
    setSortDir('desc');
  };

  const resolveSourceRoute = (sourceType, sourceId, shipmentId, invoiceId, contractId, documentId) => {
    if (contractId) {
      return `/dashboard/contracts?contract_id=${contractId}`;
    }
    if (documentId) {
      return `/dashboard/documents`;
    }
    if (invoiceId) {
      return `/dashboard/invoices?invoice_id=${invoiceId}`;
    }
    if (shipmentId) {
      return `/dashboard/shipments/${shipmentId}`;
    }
    switch (sourceType?.toUpperCase()) {
      case 'CONTRACT':
        return `/dashboard/contracts?contract_id=${sourceId}`;
      case 'DOCUMENT':
        return `/dashboard/documents`;
      case 'COMPLIANCE':
        return `/dashboard/compliance`;
      case 'SHIPMENT':
        return `/dashboard/shipments/${sourceId}`;
      case 'MILESTONE':
      case 'EXCEPTION':
        return shipmentId ? `/dashboard/shipments/${shipmentId}` : `/dashboard/shipments`;
      case 'CUSTOMER':
        return `/dashboard/customers/${sourceId}`;
      case 'INVOICE':
        return `/dashboard/invoices?invoice_id=${sourceId}`;
      case 'RFQ':
        return `/dashboard/rfqs/${sourceId}`;
      case 'QUOTATION':
        return `/dashboard/quotations`;
      case 'LEAD':
        return `/dashboard/leads`;
      default:
        return null;
    }
  };

  return (
    <div className="rec-center-container">
      {/* ── Page Header ── */}
      <div className="rec-center-header">
        <div className="rec-center-title-area">
          <h1>
            <Compass size={24} color="#0f172a" />
            AI Action & Recommendation Center
          </h1>
          <p>
            Deterministic, evidence-grounded operational guidance synthesized from active shipments,
            financial ledgers, commercial contracts, and customer signals.
          </p>
        </div>
        <div className="rec-center-header-actions">
          <button
            type="button"
            className="btn-generate-rec"
            onClick={handleGenerate}
            disabled={generating}
            aria-label="Generate and refresh recommendations"
          >
            <RefreshCw size={15} className={generating ? 'animate-spin' : ''} />
            {generating ? 'Evaluating Rules...' : 'Refresh Intelligence'}
          </button>
        </div>
      </div>

      {actionSuccessMsg && (
        <div style={{
          background: '#ecfdf5',
          border: '1px solid #a7f3d0',
          color: '#065f46',
          padding: '10px 16px',
          borderRadius: '8px',
          fontSize: '0.875rem',
          marginBottom: '16px',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
        }}>
          <span>{actionSuccessMsg}</span>
          <button
            type="button"
            onClick={() => setActionSuccessMsg(null)}
            style={{ background: 'transparent', border: 'none', cursor: 'pointer', color: '#065f46' }}
          >
            <X size={14} />
          </button>
        </div>
      )}

      {/* ── Metrics Stats Grid ── */}
      <div className="rec-stats-grid">
        <div className="rec-stat-card">
          <div className="rec-stat-icon-wrapper total">
            <Compass size={20} />
          </div>
          <div className="rec-stat-info">
            <div className="rec-stat-value">{stats?.total_active ?? 0}</div>
            <div className="rec-stat-label">Active Signals</div>
          </div>
        </div>

        <div className="rec-stat-card">
          <div className="rec-stat-icon-wrapper critical">
            <AlertTriangle size={20} />
          </div>
          <div className="rec-stat-info">
            <div className="rec-stat-value">{stats?.critical_count ?? 0}</div>
            <div className="rec-stat-label">Critical Priority</div>
          </div>
        </div>

        <div className="rec-stat-card">
          <div className="rec-stat-icon-wrapper high">
            <Clock size={20} />
          </div>
          <div className="rec-stat-info">
            <div className="rec-stat-value">{stats?.high_count ?? 0}</div>
            <div className="rec-stat-label">High Priority</div>
          </div>
        </div>

        <div className="rec-stat-card">
          <div className="rec-stat-icon-wrapper review">
            <FileCheck size={20} />
          </div>
          <div className="rec-stat-info">
            <div className="rec-stat-value">{stats?.requires_review ?? 0}</div>
            <div className="rec-stat-label">Awaiting Review</div>
          </div>
        </div>

        <div className="rec-stat-card">
          <div className="rec-stat-icon-wrapper approval">
            <ShieldCheck size={20} />
          </div>
          <div className="rec-stat-info">
            <div className="rec-stat-value">{stats?.requires_approval ?? 0}</div>
            <div className="rec-stat-label">Requires Approval</div>
          </div>
        </div>
      </div>

      {/* ── Filters Toolbar ── */}
      <div className="rec-filters-toolbar">
        <div className="rec-filters-row">
          <div className="rec-search-wrapper">
            <Search size={16} className="rec-search-icon" />
            <input
              type="text"
              className="rec-search-input"
              placeholder="Search by title, description, or reference..."
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              aria-label="Search recommendations"
            />
          </div>

          <select
            className="rec-select"
            value={category}
            onChange={(e) => setCategory(e.target.value)}
            aria-label="Filter by Category"
          >
            <option value="">All Categories</option>
            <option value="contract">Contracts & Legal</option>
            <option value="compliance">Document Compliance</option>
            <option value="finance">Finance & Collections</option>
            <option value="data_quality">Invoice Data Quality</option>
            <option value="rfq">RFQ Attention</option>
            <option value="quotation">Quotations</option>
            <option value="pricing">Pricing & Margins</option>
            <option value="operations">Operations</option>
            <option value="customer">Customer Health</option>
            <option value="sales">Sales / Leads</option>
          </select>

          <select
            className="rec-select"
            value={priority}
            onChange={(e) => setPriority(e.target.value)}
            aria-label="Filter by Priority"
          >
            <option value="">All Priorities</option>
            <option value="critical">Critical</option>
            <option value="high">High</option>
            <option value="medium">Medium</option>
            <option value="low">Low</option>
          </select>

          <select
            className="rec-select"
            value={status}
            onChange={(e) => setStatus(e.target.value)}
            aria-label="Filter by Status"
          >
            <option value="">All Statuses</option>
            <option value="new">New</option>
            <option value="reviewed">Reviewed</option>
            <option value="assigned">Assigned</option>
            <option value="approved">Approved</option>
            <option value="dismissed">Dismissed</option>
            <option value="completed">Completed</option>
          </select>

          <select
            className="rec-select"
            value={followupType}
            onChange={(e) => setFollowupType(e.target.value)}
            aria-label="Filter by Follow-Up Type"
          >
            <option value="">All Follow-Up Types</option>
            <option value="General check-in">General check-in</option>
            <option value="RFQ follow-up">RFQ follow-up</option>
            <option value="Quotation follow-up">Quotation follow-up</option>
            <option value="Shipment update">Shipment update</option>
            <option value="Exception resolution">Exception resolution</option>
            <option value="Invoice reminder">Invoice reminder</option>
            <option value="Contract renewal">Contract renewal</option>
            <option value="Service recovery">Service recovery</option>
            <option value="Account review">Account review</option>
          </select>

          <label className="rec-toggle-label">
            <input
              type="checkbox"
              checked={delayedMilestonesOnly}
              onChange={(e) => setDelayedMilestonesOnly(e.target.checked)}
            />
            Delayed Milestones
          </label>

          <label className="rec-toggle-label">
            <input
              type="checkbox"
              checked={activeExceptionsOnly}
              onChange={(e) => setActiveExceptionsOnly(e.target.checked)}
            />
            Active Exceptions
          </label>

          <label className="rec-toggle-label">
            <input
              type="checkbox"
              checked={missingOpsInfoOnly}
              onChange={(e) => setMissingOpsInfoOnly(e.target.checked)}
            />
            Missing Ops Info
          </label>

          <label className="rec-toggle-label">
            <input
              type="checkbox"
              checked={missingInfoOnly}
              onChange={(e) => setMissingInfoOnly(e.target.checked)}
            />
            Missing Commercial Info
          </label>

          <label className="rec-toggle-label">
            <input
              type="checkbox"
              checked={expiringSoonOnly}
              onChange={(e) => setExpiringSoonOnly(e.target.checked)}
            />
            Expiring Soon
          </label>

          <label className="rec-toggle-label">
            <input
              type="checkbox"
              checked={pricingConcernOnly}
              onChange={(e) => setPricingConcernOnly(e.target.checked)}
            />
            Pricing Warning
          </label>

          <label className="rec-toggle-label">
            <input
              type="checkbox"
              checked={overdueOnly}
              onChange={(e) => setOverdueOnly(e.target.checked)}
            />
            Overdue Invoices
          </label>

          <label className="rec-toggle-label">
            <input
              type="checkbox"
              checked={upcomingOnly}
              onChange={(e) => setUpcomingOnly(e.target.checked)}
            />
            Upcoming Due
          </label>

          <label className="rec-toggle-label">
            <input
              type="checkbox"
              checked={dataQualityOnly}
              onChange={(e) => setDataQualityOnly(e.target.checked)}
            />
            Invoice Data Quality
          </label>

          <label className="rec-toggle-label">
            <input
              type="checkbox"
              checked={requiresApproval}
              onChange={(e) => setRequiresApproval(e.target.checked)}
            />
            Requires Approval
          </label>

          <label className="rec-toggle-label">
            <input
              type="checkbox"
              checked={expiredOnly}
              onChange={(e) => setExpiredOnly(e.target.checked)}
            />
            Expired / Expiring
          </label>

          {contractIdFilter && (
            <div style={{ display: 'inline-flex', alignItems: 'center', gap: '6px', background: '#f5f3ff', color: '#6d28d9', padding: '4px 10px', borderRadius: '6px', fontSize: '0.8rem', fontWeight: 500 }}>
              <span>Contract: #{contractIdFilter}</span>
              <button type="button" onClick={() => setContractIdFilter('')} style={{ background: 'none', border: 'none', cursor: 'pointer', padding: 0, color: '#6d28d9' }}>
                <X size={12} />
              </button>
            </div>
          )}

          {documentIdFilter && (
            <div style={{ display: 'inline-flex', alignItems: 'center', gap: '6px', background: '#fdf2f8', color: '#be185d', padding: '4px 10px', borderRadius: '6px', fontSize: '0.8rem', fontWeight: 500 }}>
              <span>Document: {documentIdFilter}</span>
              <button type="button" onClick={() => setDocumentIdFilter('')} style={{ background: 'none', border: 'none', cursor: 'pointer', padding: 0, color: '#be185d' }}>
                <X size={12} />
              </button>
            </div>
          )}

          {complianceIdFilter && (
            <div style={{ display: 'inline-flex', alignItems: 'center', gap: '6px', background: '#fffbeb', color: '#b45309', padding: '4px 10px', borderRadius: '6px', fontSize: '0.8rem', fontWeight: 500 }}>
              <span>Compliance: #{complianceIdFilter}</span>
              <button type="button" onClick={() => setComplianceIdFilter('')} style={{ background: 'none', border: 'none', cursor: 'pointer', padding: 0, color: '#b45309' }}>
                <X size={12} />
              </button>
            </div>
          )}

          {invoiceIdFilter && (
            <div style={{ display: 'inline-flex', alignItems: 'center', gap: '6px', background: '#fef2f2', color: '#b91c1c', padding: '4px 10px', borderRadius: '6px', fontSize: '0.8rem', fontWeight: 500 }}>
              <span>Invoice: #{invoiceIdFilter}</span>
              <button type="button" onClick={() => setInvoiceIdFilter('')} style={{ background: 'none', border: 'none', cursor: 'pointer', padding: 0, color: '#b91c1c' }}>
                <X size={12} />
              </button>
            </div>
          )}

          {shipmentIdFilter && (
            <div style={{ display: 'inline-flex', alignItems: 'center', gap: '6px', background: '#eff6ff', color: '#1d4ed8', padding: '4px 10px', borderRadius: '6px', fontSize: '0.8rem', fontWeight: 500 }}>
              <span>Shipment: SH-{shipmentIdFilter}</span>
              <button type="button" onClick={() => setShipmentIdFilter('')} style={{ background: 'none', border: 'none', cursor: 'pointer', padding: 0, color: '#1d4ed8' }}>
                <X size={12} />
              </button>
            </div>
          )}

          {automationIdFilter && (
            <div style={{ display: 'inline-flex', alignItems: 'center', gap: '6px', background: '#f0fdf4', color: '#166534', padding: '4px 10px', borderRadius: '6px', fontSize: '0.8rem', fontWeight: 500 }}>
              <span>Automation: AUTO-{automationIdFilter}</span>
              <button type="button" onClick={() => setAutomationIdFilter('')} style={{ background: 'none', border: 'none', cursor: 'pointer', padding: 0, color: '#166534' }}>
                <X size={12} />
              </button>
            </div>
          )}

          {executionIdFilter && (
            <div style={{ display: 'inline-flex', alignItems: 'center', gap: '6px', background: '#f8fafc', color: '#334155', padding: '4px 10px', borderRadius: '6px', fontSize: '0.8rem', fontWeight: 500, border: '1px solid #e2e8f0' }}>
              <span>Execution: EXEC-{executionIdFilter}</span>
              <button type="button" onClick={() => setExecutionIdFilter('')} style={{ background: 'none', border: 'none', cursor: 'pointer', padding: 0, color: '#334155' }}>
                <X size={12} />
              </button>
            </div>
          )}

          <button
            type="button"
            className="btn-reset-filters"
            onClick={resetFilters}
          >
            Reset Filters
          </button>
        </div>
      </div>

      {/* ── Content Body ── */}
      {loading ? (
        <AILoadingState title="Synthesizing recommendations..." count={3} />
      ) : error ? (
        <AIErrorState
          title="Failed to Load Recommendations"
          error={error}
          onRetry={() => fetchRecommendations(pagination.page)}
        />
      ) : recommendations.length === 0 ? (
        <AIEmptyState
          title="No Pending Recommendations"
          message="All operational milestones, customer accounts, and billing ledgers are operating within standard parameters."
          shieldText="Verified Deterministic Grounding"
        />
      ) : (
        <>
          <div className="rec-list">
            {recommendations.map((rec) => {
              const sourceRoute = resolveSourceRoute(rec.source_type, rec.source_id, rec.shipment_id, rec.invoice_id, rec.contract_id, rec.document_id);

              return (
                <div key={rec.id} className="rec-card">
                  <div className="rec-card-header">
                    <div className="rec-card-badges">
                      <span className={`rec-badge priority-${rec.priority}`}>
                        {rec.priority}
                      </span>
                      <span className="rec-badge category">
                        {rec.category}
                      </span>
                      {rec.requires_approval && (
                        <span className="rec-badge approval">
                          Requires Approval
                        </span>
                      )}
                      {rec.contract_id && (
                        <Link to={`/dashboard/contracts?contract_id=${rec.contract_id}`} className="rec-badge" style={{ background: '#f5f3ff', color: '#6d28d9', border: '1px solid #ddd6fe', textDecoration: 'none' }}>
                          CNT-{rec.contract_id}
                        </Link>
                      )}
                      {rec.document_id && (
                        <Link to={`/dashboard/documents`} className="rec-badge" style={{ background: '#fdf2f8', color: '#be185d', border: '1px solid #fbcfe8', textDecoration: 'none' }}>
                          DOC: {rec.document_id}
                        </Link>
                      )}
                      {rec.compliance_id && (
                        <span className="rec-badge" style={{ background: '#fffbeb', color: '#b45309', border: '1px solid #fde68a' }}>
                          COMP-{rec.compliance_id}
                        </span>
                      )}
                      {rec.invoice_id && (
                        <Link to={`/dashboard/invoices?invoice_id=${rec.invoice_id}`} className="rec-badge" style={{ background: '#fef2f2', color: '#b91c1c', border: '1px solid #fca5a5', textDecoration: 'none' }}>
                          INV-{rec.invoice_id}
                        </Link>
                      )}
                      {rec.delayed_milestone && (
                        <span className="rec-badge" style={{ background: '#fef3c7', color: '#92400e', border: '1px solid #fde68a' }}>
                          Delayed Milestone
                        </span>
                      )}
                      {rec.active_exception && (
                        <span className="rec-badge" style={{ background: '#fee2e2', color: '#991b1b', border: '1px solid #fecaca' }}>
                          Active Exception
                        </span>
                      )}
                      {rec.missing_ops_info && (
                        <span className="rec-badge" style={{ background: '#e0e7ff', color: '#3730a3', border: '1px solid #c7d2fe' }}>
                          Missing Ops Info
                        </span>
                      )}
                      {rec.shipment_id && rec.source_type !== 'SHIPMENT' && (
                        <Link to={`/dashboard/shipments/${rec.shipment_id}`} className="rec-badge" style={{ background: '#f0fdf4', color: '#166534', border: '1px solid #bbf7d0', textDecoration: 'none' }}>
                          SH-{rec.shipment_id}
                        </Link>
                      )}
                      {rec.automation_id && (
                        <Link to={`/dashboard/automations`} className="rec-badge" style={{ background: '#f0fdf4', color: '#166534', border: '1px solid #bbf7d0', textDecoration: 'none' }} title="Generated by Workflow Automation">
                          AUTO-{rec.automation_id}
                        </Link>
                      )}
                      {rec.execution_id && (
                        <span className="rec-badge" style={{ background: '#f8fafc', color: '#475569', border: '1px solid #e2e8f0' }} title="Execution ID">
                          EXEC-{rec.execution_id}
                        </span>
                      )}
                      {sourceRoute ? (
                        <Link to={sourceRoute} className="rec-source-link">
                          <span>{rec.source_reference}</span>
                          <ExternalLink size={12} />
                        </Link>
                      ) : (
                        <span className="rec-source-link">{rec.source_reference}</span>
                      )}
                    </div>

                    <div className="rec-card-time">
                      Created {new Date(rec.created_at).toLocaleDateString()}
                    </div>
                  </div>

                  <div className="rec-card-body">
                    <h3>{rec.title}</h3>
                    <p>{rec.description}</p>

                    <div className="rec-action-box">
                      <ChevronRight size={16} className="rec-action-icon" />
                      <div className="rec-action-text">
                        <strong>Recommended Step:</strong> {rec.recommended_action}
                      </div>
                    </div>
                  </div>

                  <div className="rec-card-footer">
                    <div className="rec-card-meta">
                      <div className="rec-meta-item">
                        <AIConfidenceIndicator confidence={rec.confidence} score={rec.confidence_score} />
                      </div>
                      <div className="rec-meta-item">
                        <User size={13} />
                        <span>{rec.assignee_name || 'Unassigned'}</span>
                      </div>
                      <div className="rec-meta-item">
                        <Info size={13} />
                        <span>Status: <strong style={{ textTransform: 'capitalize' }}>{rec.status}</strong></span>
                      </div>
                    </div>

                    <div className="rec-card-actions">
                      <button
                        type="button"
                        className="btn-rec-action"
                        onClick={() => setSelectedRec(rec)}
                      >
                        View Details
                      </button>

                      {rec.status === 'new' && (
                        <button
                          type="button"
                          className="btn-rec-action"
                          onClick={() => handleMarkReviewed(rec)}
                        >
                          <Check size={14} />
                          Mark Reviewed
                        </button>
                      )}

                      {/* Controlled Action Preview (Task 2.3) */}
                      <button
                        type="button"
                        className="btn-rec-action"
                        style={{ background: '#f8fafc', color: '#0f172a', borderColor: '#cbd5e1' }}
                        onClick={() => handleOpenActionPreview(rec)}
                        title="Inspect proposed action, expected effect, risk level, and evidence"
                      >
                        <Eye size={13} />
                        Action Preview
                      </button>

                      {/* Draft Message Controls (Task 2.2, 2.3, 2.5, 2.6) */}
                      {(rec.category === 'customer' || rec.category === 'sales' || rec.category === 'rfq' || rec.category === 'quotation' || rec.category === 'pricing' || rec.category === 'finance' || rec.category === 'data_quality' || rec.category === 'contract' || rec.category === 'compliance' || rec.followup_type) && rec.status !== 'dismissed' && rec.status !== 'completed' && (
                        <>
                          <button
                            type="button"
                            className="btn-rec-action"
                            style={{ background: '#eff6ff', color: '#1d4ed8', borderColor: '#bfdbfe' }}
                            onClick={() => handleOpenDraftModal(rec)}
                            title="Generate/review editable communication draft"
                          >
                            <Mail size={13} />
                            {rec.draft_status === 'SAVED' ? 'Edit Draft' : rec.draft_status === 'DRAFTED' ? 'Review Draft' : 'Draft Message'}
                          </button>

                          {!rec.followup_task_id && (rec.category === 'customer' || rec.category === 'sales') && (
                            <button
                              type="button"
                              className="btn-rec-action"
                              style={{ background: '#f8fafc', color: '#0f172a', borderColor: '#cbd5e1' }}
                              onClick={() => handleOpenTaskModal(rec)}
                              title="Create controlled internal follow-up task"
                            >
                              <ListTodo size={13} />
                              Create Task
                            </button>
                          )}
                          {rec.followup_task_id && (
                            <span style={{ fontSize: '0.6875rem', fontWeight: 600, color: '#059669', background: '#ecfdf5', padding: '4px 8px', borderRadius: '4px', border: '1px solid #a7f3d0' }}>
                              Task #{rec.followup_task_id} Active
                            </span>
                          )}
                        </>
                      )}

                      {/* Approval Status Badge or Request Button */}
                      {rec.approval_id ? (
                        <span style={{ fontSize: '0.6875rem', fontWeight: 600, color: '#4338ca', background: '#e0e7ff', padding: '4px 8px', borderRadius: '4px', border: '1px solid #c7d2fe' }}>
                          Approval #{rec.approval_id} Pending
                        </span>
                      ) : rec.requires_approval && rec.status !== 'dismissed' && rec.status !== 'completed' && (
                        <button
                          type="button"
                          className="btn-rec-action"
                          style={{ background: '#fffbeb', color: '#b45309', borderColor: '#fde68a' }}
                          onClick={() => handleOpenActionPreview(rec)}
                          title="Submit to centralized approval system"
                        >
                          <ShieldCheck size={13} />
                          Request Approval
                        </button>
                      )}

                      {rec.status !== 'completed' && rec.status !== 'dismissed' && (
                        <>
                          <button
                            type="button"
                            className="btn-rec-action"
                            onClick={() => {
                              setAssignModalRec(rec);
                              setAssigneeName(rec.assignee_name || '');
                            }}
                          >
                            <UserCheck size={14} />
                            Assign
                          </button>

                          <button
                            type="button"
                            className="btn-rec-action dismiss"
                            onClick={() => {
                              setDismissModalRec(rec);
                              setDismissReason('');
                            }}
                          >
                            Dismiss
                          </button>
                        </>
                      )}
                    </div>
                  </div>
                </div>
              );
            })}
          </div>

          {/* ── Pagination ── */}
          {pagination.total_pages > 1 && (
            <div className="rec-pagination">
              <div>
                Showing {((pagination.page - 1) * pagination.limit) + 1} to{' '}
                {Math.min(pagination.page * pagination.limit, pagination.total)} of{' '}
                {pagination.total} recommendations
              </div>
              <div className="rec-pagination-controls">
                <button
                  type="button"
                  className="btn-page"
                  disabled={pagination.page <= 1}
                  onClick={() => fetchRecommendations(pagination.page - 1)}
                >
                  Previous
                </button>
                <span style={{ margin: '0 8px', fontWeight: 600 }}>
                  Page {pagination.page} of {pagination.total_pages}
                </span>
                <button
                  type="button"
                  className="btn-page"
                  disabled={pagination.page >= pagination.total_pages}
                  onClick={() => fetchRecommendations(pagination.page + 1)}
                >
                  Next
                </button>
              </div>
            </div>
          )}
        </>
      )}

      {/* ── Detail Drawer ── */}
      {selectedRec && (
        <div className="rec-drawer-overlay" onClick={() => setSelectedRec(null)}>
          <div className="rec-drawer" onClick={(e) => e.stopPropagation()}>
            <div className="rec-drawer-header">
              <div>
                <div style={{ display: 'flex', gap: '8px', marginBottom: '8px' }}>
                  <span className={`rec-badge priority-${selectedRec.priority}`}>{selectedRec.priority}</span>
                  <span className="rec-badge category">{selectedRec.category}</span>
                  <span className="rec-badge" style={{ background: '#f1f5f9', color: '#475569' }}>Status: {selectedRec.status}</span>
                </div>
                <h2 style={{ fontSize: '1.25rem', fontWeight: 700, margin: 0, color: '#0f172a' }}>
                  {selectedRec.title}
                </h2>
              </div>
              <button type="button" className="rec-drawer-close" onClick={() => setSelectedRec(null)}>
                <X size={20} />
              </button>
            </div>

            <div className="rec-drawer-content">
              <div className="rec-section-block">
                <h4>Operational Summary</h4>
                <p>{selectedRec.description}</p>
              </div>

              <div className="rec-section-block">
                <h4>Recommended Action</h4>
                <div className="rec-action-box">
                  <ChevronRight size={16} className="rec-action-icon" />
                  <div className="rec-action-text">{selectedRec.recommended_action}</div>
                </div>
              </div>

              {selectedRec.requires_approval && (
                <div style={{
                  background: '#faf5ff',
                  border: '1px solid #e9d5ff',
                  borderRadius: '8px',
                  padding: '12px 16px',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'space-between',
                  gap: '12px',
                }}>
                  <div style={{ fontSize: '0.8125rem', color: '#581c87', fontWeight: 500 }}>
                    <strong>Approval Required:</strong> This action involves contractual terms or credit limits and cannot be automated without an approval workflow.
                  </div>
                  <Link
                    to="/dashboard/approvals"
                    className="btn-rec-action primary"
                    style={{ textDecoration: 'none', whiteSpace: 'nowrap' }}
                  >
                    Open Approvals
                  </Link>
                </div>
              )}

              <div className="rec-section-block">
                <h4>Supporting Factual Evidence</h4>
                {selectedRec.evidence && selectedRec.evidence.length > 0 ? (
                  <table className="rec-evidence-table">
                    <thead>
                      <tr>
                        <th>Source Entity</th>
                        <th>Field Tracked</th>
                        <th>Observed Value</th>
                        <th>Evidence Context</th>
                      </tr>
                    </thead>
                    <tbody>
                      {selectedRec.evidence.map((ev, idx) => (
                        <tr key={idx}>
                          <td><strong>{ev.source_ref}</strong></td>
                          <td><code>{ev.field_name}</code></td>
                          <td>{String(ev.observed_value)}</td>
                          <td>{ev.description}</td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                ) : (
                  <p style={{ color: '#94a3b8', fontSize: '0.8125rem' }}>No granular evidence fields attached.</p>
                )}
              </div>

              <div className="rec-section-block">
                <h4>Audit & Traceability Details</h4>
                <div style={{
                  background: '#f8fafc',
                  border: '1px solid #e2e8f0',
                  borderRadius: '8px',
                  padding: '12px',
                  fontSize: '0.75rem',
                  display: 'grid',
                  gridTemplateColumns: '1fr 1fr',
                  gap: '8px',
                  color: '#475569',
                }}>
                  <div><strong>Rule Applied:</strong> {selectedRec.rule_applied}</div>
                  <div><strong>Generated By:</strong> {selectedRec.generated_by}</div>
                  <div><strong>Correlation ID:</strong> <code>{selectedRec.correlation_id}</code></div>
                  <div><strong>Freshness:</strong> {new Date(selectedRec.freshness).toLocaleTimeString()}</div>
                  {selectedRec.dismissed_reason && (
                    <div style={{ gridColumn: 'span 2', color: '#dc2626' }}>
                      <strong>Dismissal Reason:</strong> {selectedRec.dismissed_reason}
                    </div>
                  )}
                </div>
              </div>
            </div>

            <div className="rec-drawer-footer">
              <button type="button" className="btn-rec-action" onClick={() => setSelectedRec(null)}>
                Close
              </button>
            </div>
          </div>
        </div>
      )}

      {/* ── Dismiss Reason Modal ── */}
      {dismissModalRec && (
        <div className="rec-modal-overlay" onClick={() => setDismissModalRec(null)}>
          <div className="rec-modal" onClick={(e) => e.stopPropagation()}>
            <h3>Dismiss Recommendation</h3>
            <p>
              Please supply an operational reason for dismissing recommendation #{dismissModalRec.id} ({dismissModalRec.source_reference}).
              This will be permanently recorded in the audit log.
            </p>
            <textarea
              placeholder="e.g. Issue resolved externally with port terminal; carrier confirms GPS updated."
              value={dismissReason}
              onChange={(e) => setDismissReason(e.target.value)}
              rows={3}
            />
            <div className="rec-modal-actions">
              <button
                type="button"
                className="btn-rec-action"
                onClick={() => setDismissModalRec(null)}
              >
                Cancel
              </button>
              <button
                type="button"
                className="btn-rec-action dismiss"
                disabled={!dismissReason.trim()}
                onClick={handleConfirmDismiss}
              >
                Confirm Dismissal
              </button>
            </div>
          </div>
        </div>
      )}

      {/* ── Assign Operator Modal ── */}
      {assignModalRec && (
        <div className="rec-modal-overlay" onClick={() => setAssignModalRec(null)}>
          <div className="rec-modal" onClick={(e) => e.stopPropagation()}>
            <h3>Assign Recommendation</h3>
            <p>
              Assign recommendation #{assignModalRec.id} ({assignModalRec.title}) to an operations or sales team member.
            </p>
            <input
              type="text"
              className="rec-search-input"
              style={{ marginBottom: '16px' }}
              placeholder="Enter team member name..."
              value={assigneeName}
              onChange={(e) => setAssigneeName(e.target.value)}
            />
            <div className="rec-modal-actions">
              <button
                type="button"
                className="btn-rec-action"
                onClick={() => setAssignModalRec(null)}
              >
                Cancel
              </button>
              <button
                type="button"
                className="btn-rec-action primary"
                disabled={!assigneeName.trim()}
                onClick={handleConfirmAssign}
              >
                Assign
              </button>
            </div>
          </div>
        </div>
      )}

      {/* ── Task 2.2: Draft Communication Modal (100% White/Light Theme) ── */}
      {draftModalRec && (
        <div className="rec-modal-overlay" onClick={() => setDraftModalRec(null)}>
          <div className="rec-modal" style={{ maxWidth: '640px' }} onClick={(e) => e.stopPropagation()}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '12px' }}>
              <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                <Mail size={18} color="#2563eb" />
                <h3 style={{ margin: 0 }}>Follow-Up Communication Draft</h3>
              </div>
              <button
                type="button"
                style={{ background: 'transparent', border: 'none', cursor: 'pointer', color: '#64748b' }}
                onClick={() => setDraftModalRec(null)}
              >
                <X size={18} />
              </button>
            </div>

            <p style={{ fontSize: '0.8125rem', color: '#64748b', marginBottom: '16px' }}>
              Customer: <strong>{draftModalRec.customer_name || 'Customer'}</strong> • Ref: <strong>{draftModalRec.source_reference}</strong>.
              Drafts are grounded in verified records only. Review and edit before saving. <em>Messages are never sent automatically.</em>
            </p>

            {generatingDraft ? (
              <div style={{ padding: '32px 0', textAlign: 'center', color: '#64748b' }}>
                <RefreshCw size={24} className="animate-spin" style={{ margin: '0 auto 12px auto' }} />
                <div>Synthesizing verified factual draft...</div>
              </div>
            ) : (
              <>
                <div style={{ marginBottom: '12px' }}>
                  <label style={{ display: 'block', fontSize: '0.75rem', fontWeight: 600, color: '#475569', marginBottom: '4px' }}>
                    Subject Line
                  </label>
                  <input
                    type="text"
                    className="rec-search-input"
                    style={{ width: '100%' }}
                    value={draftSubject}
                    onChange={(e) => setDraftSubject(e.target.value)}
                    placeholder="Enter message subject..."
                  />
                </div>

                <div style={{ marginBottom: '16px' }}>
                  <label style={{ display: 'block', fontSize: '0.75rem', fontWeight: 600, color: '#475569', marginBottom: '4px' }}>
                    Message Body (Editable)
                  </label>
                  <textarea
                    rows={8}
                    style={{
                      width: '100%',
                      fontFamily: 'inherit',
                      fontSize: '0.875rem',
                      lineHeight: 1.5,
                      padding: '10px',
                      borderRadius: '8px',
                      border: '1px solid #cbd5e1',
                      resize: 'vertical',
                      color: '#0f172a',
                    }}
                    value={draftBody}
                    onChange={(e) => setDraftBody(e.target.value)}
                  />
                </div>

                <div style={{
                  background: '#f8fafc',
                  border: '1px solid #e2e8f0',
                  borderRadius: '6px',
                  padding: '8px 12px',
                  fontSize: '0.75rem',
                  color: '#64748b',
                  marginBottom: '16px',
                }}>
                  Status: <strong>{draftModalRec.draft_status || 'DRAFTED'}</strong> • Controlled Assistant mode: External sending requires human dispatch.
                </div>

                <div className="rec-modal-actions">
                  <button
                    type="button"
                    className="btn-rec-action"
                    onClick={() => setDraftModalRec(null)}
                  >
                    Cancel
                  </button>
                  <button
                    type="button"
                    className="btn-rec-action primary"
                    disabled={savingDraft || !draftSubject.trim() || !draftBody.trim()}
                    onClick={handleSaveDraft}
                  >
                    <Check size={14} />
                    {savingDraft ? 'Saving...' : 'Save Draft'}
                  </button>
                </div>
              </>
            )}
          </div>
        </div>
      )}

      {/* ── Task 2.2: Create Follow-Up Task Modal (100% White/Light Theme) ── */}
      {taskModalRec && (
        <div className="rec-modal-overlay" onClick={() => setTaskModalRec(null)}>
          <div className="rec-modal" style={{ maxWidth: '520px' }} onClick={(e) => e.stopPropagation()}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '12px' }}>
              <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                <ListTodo size={18} color="#2563eb" />
                <h3 style={{ margin: 0 }}>Create Internal Follow-Up Task</h3>
              </div>
              <button
                type="button"
                style={{ background: 'transparent', border: 'none', cursor: 'pointer', color: '#64748b' }}
                onClick={() => setTaskModalRec(null)}
              >
                <X size={18} />
              </button>
            </div>

            <p style={{ fontSize: '0.8125rem', color: '#64748b', marginBottom: '16px' }}>
              Create an idempotent internal follow-up task for customer <strong>{taskModalRec.customer_name || 'Customer'}</strong> ({taskModalRec.source_reference}).
            </p>

            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '12px', marginBottom: '12px' }}>
              <div>
                <label style={{ display: 'block', fontSize: '0.75rem', fontWeight: 600, color: '#475569', marginBottom: '4px' }}>
                  Assignee
                </label>
                <input
                  type="text"
                  className="rec-search-input"
                  style={{ width: '100%' }}
                  value={taskAssigneeName}
                  onChange={(e) => setTaskAssigneeName(e.target.value)}
                  placeholder="Owner name..."
                />
              </div>

              <div>
                <label style={{ display: 'block', fontSize: '0.75rem', fontWeight: 600, color: '#475569', marginBottom: '4px' }}>
                  Priority
                </label>
                <select
                  className="rec-select"
                  style={{ width: '100%' }}
                  value={taskPriority}
                  onChange={(e) => setTaskPriority(e.target.value)}
                >
                  <option value="critical">Critical</option>
                  <option value="high">High</option>
                  <option value="medium">Medium</option>
                  <option value="low">Low</option>
                </select>
              </div>
            </div>

            <div style={{ marginBottom: '12px' }}>
              <label style={{ display: 'block', fontSize: '0.75rem', fontWeight: 600, color: '#475569', marginBottom: '4px' }}>
                Due Date
              </label>
              <input
                type="date"
                className="rec-search-input"
                style={{ width: '100%' }}
                value={taskDueDate}
                onChange={(e) => setTaskDueDate(e.target.value)}
              />
            </div>

            <div style={{ marginBottom: '16px' }}>
              <label style={{ display: 'block', fontSize: '0.75rem', fontWeight: 600, color: '#475569', marginBottom: '4px' }}>
                Instructions / Notes
              </label>
              <textarea
                rows={3}
                style={{
                  width: '100%',
                  fontFamily: 'inherit',
                  fontSize: '0.875rem',
                  padding: '8px',
                  borderRadius: '8px',
                  border: '1px solid #cbd5e1',
                  color: '#0f172a',
                }}
                placeholder="Optional notes for the assigned owner..."
                value={taskNotes}
                onChange={(e) => setTaskNotes(e.target.value)}
              />
            </div>

            <div className="rec-modal-actions">
              <button
                type="button"
                className="btn-rec-action"
                onClick={() => setTaskModalRec(null)}
              >
                Cancel
              </button>
              <button
                type="button"
                className="btn-rec-action primary"
                disabled={creatingTask}
                onClick={handleConfirmCreateTask}
              >
                <Check size={14} />
                {creatingTask ? 'Creating...' : 'Create Controlled Task'}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* ── Task 2.3: Controlled Action Preview Modal (100% Native White/Light Theme) ── */}
      {previewModalRec && (
        <div className="rec-modal-overlay" onClick={() => setPreviewModalRec(null)}>
          <div className="rec-modal" style={{ maxWidth: '640px' }} onClick={(e) => e.stopPropagation()}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '14px', borderBottom: '1px solid #f1f5f9', paddingBottom: '12px' }}>
              <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                <Eye size={18} color="#2563eb" />
                <h3 style={{ margin: 0, fontSize: '1.125rem', color: '#0f172a' }}>Controlled Action Preview</h3>
              </div>
              <button
                type="button"
                style={{ background: 'transparent', border: 'none', cursor: 'pointer', color: '#64748b' }}
                onClick={() => setPreviewModalRec(null)}
              >
                <X size={18} />
              </button>
            </div>

            {loadingPreview ? (
              <AILoadingState title="Generating action preview and verifying safety constraints..." count={2} />
            ) : actionPreviewData ? (
              <div>
                <div style={{ display: 'flex', gap: '8px', alignItems: 'center', marginBottom: '16px' }}>
                  <span style={{
                    fontSize: '0.75rem',
                    fontWeight: 700,
                    padding: '3px 8px',
                    borderRadius: '4px',
                    background: '#f1f5f9',
                    color: '#334155',
                  }}>
                    {actionPreviewData.source_type} #{actionPreviewData.source_id}
                  </span>
                  <span style={{
                    fontSize: '0.75rem',
                    fontWeight: 700,
                    padding: '3px 8px',
                    borderRadius: '4px',
                    background: actionPreviewData.risk_level === 'CRITICAL' ? '#fef2f2' : actionPreviewData.risk_level === 'HIGH' ? '#fffbeb' : '#eff6ff',
                    color: actionPreviewData.risk_level === 'CRITICAL' ? '#991b1b' : actionPreviewData.risk_level === 'HIGH' ? '#92400e' : '#1e40af',
                  }}>
                    Risk: {actionPreviewData.risk_level}
                  </span>
                  {actionPreviewData.required_approval && (
                    <span style={{
                      fontSize: '0.75rem',
                      fontWeight: 700,
                      padding: '3px 8px',
                      borderRadius: '4px',
                      background: '#e0e7ff',
                      color: '#4338ca',
                    }}>
                      Requires Approval
                    </span>
                  )}
                </div>

                <div style={{
                  background: '#f8fafc',
                  border: '1px solid #e2e8f0',
                  borderRadius: '8px',
                  padding: '14px',
                  marginBottom: '16px',
                }}>
                  <div style={{ fontSize: '0.75rem', fontWeight: 600, color: '#64748b', textTransform: 'uppercase', letterSpacing: '0.03em', marginBottom: '4px' }}>
                    Proposed Action
                  </div>
                  <div style={{ fontSize: '0.9375rem', fontWeight: 700, color: '#0f172a', marginBottom: '10px' }}>
                    {actionPreviewData.proposed_action}
                  </div>

                  <div style={{ fontSize: '0.75rem', fontWeight: 600, color: '#64748b', textTransform: 'uppercase', letterSpacing: '0.03em', marginBottom: '4px' }}>
                    Expected Effect
                  </div>
                  <div style={{ fontSize: '0.875rem', color: '#334155', lineHeight: 1.5 }}>
                    {actionPreviewData.expected_effect}
                  </div>
                </div>

                {/* Supporting Factual Evidence */}
                {actionPreviewData.evidence && actionPreviewData.evidence.length > 0 && (
                  <div style={{ marginBottom: '16px' }}>
                    <div style={{ fontSize: '0.8125rem', fontWeight: 700, color: '#1e293b', marginBottom: '8px' }}>
                      Verified Grounding Evidence ({actionPreviewData.evidence.length})
                    </div>
                    <div style={{ maxHeight: '160px', overflowY: 'auto', border: '1px solid #e2e8f0', borderRadius: '6px' }}>
                      <AIEvidenceList evidence={actionPreviewData.evidence} />
                    </div>
                  </div>
                )}

                {/* Human-in-the-Loop Approval Action */}
                {actionPreviewData.required_approval && !actionPreviewData.approval_id && (
                  <div style={{
                    background: '#fffbeb',
                    border: '1px solid #fde68a',
                    borderRadius: '8px',
                    padding: '12px',
                    marginBottom: '16px',
                  }}>
                    <div style={{ display: 'flex', alignItems: 'center', gap: '6px', fontSize: '0.875rem', fontWeight: 600, color: '#92400e', marginBottom: '6px' }}>
                      <ShieldCheck size={16} />
                      Commercial / Operational Approval Required
                    </div>
                    <p style={{ fontSize: '0.75rem', color: '#78350f', margin: '0 0 8px 0' }}>
                      This action has potential pricing, margin, or contractual implications. Route to the HITL approvals system for authorized signoff.
                    </p>
                    <textarea
                      rows={2}
                      style={{
                        width: '100%',
                        fontFamily: 'inherit',
                        fontSize: '0.8125rem',
                        padding: '6px 8px',
                        borderRadius: '6px',
                        border: '1px solid #fcd34d',
                        background: '#ffffff',
                        color: '#0f172a',
                        marginBottom: '8px',
                      }}
                      placeholder="Optional justification note for the approver..."
                      value={approvalNotes}
                      onChange={(e) => setApprovalNotes(e.target.value)}
                    />
                    <button
                      type="button"
                      className="btn-rec-action primary"
                      style={{ background: '#d97706', borderColor: '#b45309', color: '#ffffff' }}
                      disabled={submittingApproval}
                      onClick={handleSubmitApproval}
                    >
                      <ShieldCheck size={14} />
                      {submittingApproval ? 'Submitting...' : 'Submit to HITL Approval System'}
                    </button>
                  </div>
                )}

                {actionPreviewData.approval_id && (
                  <div style={{
                    background: '#ecfdf5',
                    border: '1px solid #a7f3d0',
                    color: '#065f46',
                    padding: '10px 14px',
                    borderRadius: '8px',
                    fontSize: '0.8125rem',
                    marginBottom: '16px',
                    display: 'flex',
                    alignItems: 'center',
                    gap: '8px',
                  }}>
                    <CheckCircle2 size={16} />
                    <span>Approval request <strong>#{actionPreviewData.approval_id}</strong> is active in the approvals workflow.</span>
                  </div>
                )}

                <div style={{
                  display: 'flex',
                  justifyContent: 'space-between',
                  alignItems: 'center',
                  fontSize: '0.6875rem',
                  color: '#94a3b8',
                  borderTop: '1px solid #f1f5f9',
                  paddingTop: '10px',
                  marginTop: '10px',
                }}>
                  <span>Initiated by: {actionPreviewData.initiated_by_user || 'Operator'}</span>
                  <span>Correlation: {actionPreviewData.correlation_id || 'preview-corr'}</span>
                </div>

                <div className="rec-modal-actions" style={{ marginTop: '16px' }}>
                  <button
                    type="button"
                    className="btn-rec-action"
                    onClick={() => setPreviewModalRec(null)}
                  >
                    Close Preview
                  </button>
                </div>
              </div>
            ) : (
              <AIErrorState title="Preview Unavailable" message="Could not load preview details for this recommendation." />
            )}
          </div>
        </div>
      )}
    </div>
  );
}

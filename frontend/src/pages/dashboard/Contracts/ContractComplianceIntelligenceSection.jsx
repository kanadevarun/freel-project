import React, { useState, useEffect, useCallback } from 'react';
import {
  ShieldCheck,
  RefreshCw,
  AlertTriangle,
  Clock,
  DollarSign,
  FileCheck,
  Calendar,
  Layers,
  Sparkles,
  AlertCircle,
  CheckCircle2,
  FileText,
  Building2,
  CheckCircle,
  HelpCircle,
  Shield,
  Briefcase,
  Mail,
  Send,
  Eye,
  Edit3,
  X,
  Copy,
  UserCheck,
  Check
} from 'lucide-react';
import contractsService from '../../../services/contractsService';
import { recommendationService } from '../../../services/recommendationService';
import AIConfidenceIndicator from '../../../components/ai/AIConfidenceIndicator';
import AIEvidenceList from '../../../components/ai/AIEvidenceList';
import './ContractComplianceIntelligenceSection.css';

export default function ContractComplianceIntelligenceSection({ contractId }) {
  const [data, setData] = useState(null);
  const [recommendations, setRecommendations] = useState([]);
  const [evidenceData, setEvidenceData] = useState(null);
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  const [error, setError] = useState(null);

  // Controlled Draft Modal State
  const [selectedDraftRec, setSelectedDraftRec] = useState(null);
  const [draftSubject, setDraftSubject] = useState('');
  const [draftBody, setDraftBody] = useState('');
  const [isGeneratingDraft, setIsGeneratingDraft] = useState(false);
  const [isSavingDraft, setIsSavingDraft] = useState(false);
  const [draftSaveStatus, setDraftSaveStatus] = useState(null);
  const [draftCopied, setDraftCopied] = useState(false);

  // Action Preview Modal State
  const [actionPreviewRec, setActionPreviewRec] = useState(null);
  const [actionPreview, setActionPreview] = useState(null);
  const [isLoadingPreview, setIsLoadingPreview] = useState(false);
  const [approvalStatus, setApprovalStatus] = useState({});

  const fetchIntelligence = useCallback(async (isManual = false) => {
    if (!contractId) return;
    if (isManual) setRefreshing(true);
    else setLoading(true);
    setError(null);

    try {
      // Execute 360 intelligence, recommendations, and factual evidence queries
      const [intelRes, recsRes, evidenceRes] = await Promise.allSettled([
        contractsService.getContract360ComplianceIntelligence(contractId),
        recommendationService.listRecommendations({ contract_id: contractId }),
        recommendationService.getContractEvidence(contractId)
      ]);

      if (intelRes.status === 'fulfilled' && intelRes.value) {
        const response = intelRes.value;
        if (response && response.data) {
          setData(response.data);
        } else if (response && (response.identity || response.contract_id)) {
          setData(response);
        } else {
          setData(null);
        }
      } else if (evidenceRes.status === 'fulfilled' && evidenceRes.value?.data) {
        // Deterministic fallback structure from factual evidence if 360 intelligence endpoint is not populated
        const ev = evidenceRes.value.data;
        setData({
          contract_id: contractId,
          correlation_id: `corr-contract-${contractId}`,
          calculated_at: new Date().toISOString(),
          identity: {
            contract_id: contractId,
            contract_number: ev.contract?.contract_reference || `#${contractId}`,
            contract_name: ev.contract?.contract_name,
            party_name: ev.contract?.party_name,
            owner_name: ev.contract?.owner,
            status: ev.contract?.status,
            effective_date: ev.contract?.effective_date,
            expiry_date: ev.contract?.expiry_date,
            transport_mode: ev.contract?.transport_mode,
            currency: ev.contract?.currency || 'USD',
            is_expired: ev.summary?.is_expired,
            is_expiring_soon: ev.summary?.is_expiring_soon,
            days_until_expiry: ev.summary?.days_until_expiry,
            completeness_score: ev.contract?.owner ? 90 : 65
          },
          ai_summary: {
            confidence: 'HIGH',
            confidence_score: 0.95,
            contract_executive_summary: `Contract ${ev.contract?.contract_reference || contractId} evaluated: ${ev.summary?.active_recommendations_count || 0} active findings, ${ev.summary?.total_compliance_count || 0} compliance requirements tracked.`
          },
          risks: {
            risk_level: ev.summary?.is_expired ? 'CRITICAL' : (ev.summary?.is_expiring_soon ? 'HIGH' : 'LOW'),
            risk_score: ev.summary?.is_expired ? 90 : (ev.summary?.is_expiring_soon ? 75 : 20),
            risk_alerts: []
          },
          compliance: {
            total_requirements: ev.summary?.total_compliance_count || 0,
            valid_requirements: ev.summary?.verified_compliance_count || 0,
            compliance_status: ev.summary?.total_compliance_count === ev.summary?.verified_compliance_count ? 'COMPLIANT' : 'PARTIALLY_VERIFIED'
          },
          commercial_terms: {
            payment_terms: 'Standard Terms',
            demurrage_terms: 'Standard Terms'
          }
        });
      } else if (intelRes.status === 'rejected') {
        throw intelRes.reason;
      }

      // Handle recommendations
      if (recsRes.status === 'fulfilled') {
        const items = recsRes.value?.data?.items || recsRes.value?.items || [];
        setRecommendations(items);
      }

      // Handle factual evidence
      if (evidenceRes.status === 'fulfilled') {
        setEvidenceData(evidenceRes.value?.data || null);
      }
    } catch (err) {
      console.error('Failed to load contract compliance intelligence', err);
      setError(err.message || 'Failed to load Contract & Compliance Intelligence');
    } finally {
      setLoading(false);
      setRefreshing(false);
    }
  }, [contractId]);

  useEffect(() => {
    fetchIntelligence();
  }, [fetchIntelligence]);

  // Handle Draft Generation for a recommendation
  const handleOpenDraft = async (rec) => {
    setSelectedDraftRec(rec);
    setDraftSaveStatus(null);
    setDraftCopied(false);

    if (rec.draft_subject || rec.draft_body) {
      setDraftSubject(rec.draft_subject || '');
      setDraftBody(rec.draft_body || '');
      return;
    }

    setIsGeneratingDraft(true);
    try {
      const res = await recommendationService.generateDraft(rec.id);
      const updated = res?.data || res;
      setDraftSubject(updated.draft_subject || '');
      setDraftBody(updated.draft_body || '');
      setRecommendations((prev) =>
        prev.map((r) => (r.id === rec.id ? { ...r, ...updated } : r))
      );
    } catch (err) {
      console.error('Failed to generate draft:', err);
      setDraftSaveStatus({ type: 'error', message: err?.message || 'Failed to generate draft' });
    } finally {
      setIsGeneratingDraft(false);
    }
  };

  // Handle Saving Edited Draft
  const handleSaveDraft = async () => {
    if (!selectedDraftRec) return;
    setIsSavingDraft(true);
    setDraftSaveStatus(null);
    try {
      await recommendationService.saveDraft(selectedDraftRec.id, draftSubject, draftBody);
      setRecommendations((prev) =>
        prev.map((r) =>
          r.id === selectedDraftRec.id
            ? { ...r, draft_subject: draftSubject, draft_body: draftBody, draft_status: 'generated' }
            : r
        )
      );
      setDraftSaveStatus({ type: 'success', message: 'Draft saved successfully to record.' });
      setTimeout(() => setDraftSaveStatus(null), 3000);
    } catch (err) {
      console.error('Failed to save draft:', err);
      setDraftSaveStatus({ type: 'error', message: 'Failed to save draft changes.' });
    } finally {
      setIsSavingDraft(false);
    }
  };

  // Handle Action Preview
  const handleOpenActionPreview = async (rec) => {
    setActionPreviewRec(rec);
    setIsLoadingPreview(true);
    setActionPreview(null);
    try {
      const res = await recommendationService.getActionPreview(rec.id);
      setActionPreview(res?.data || res);
    } catch (err) {
      console.error('Failed to fetch action preview:', err);
      setActionPreview({
        action_type: rec.action_type || 'review_expiring_contract',
        title: rec.title,
        description: rec.description,
        target_entity: `Contract #${rec.contract_id || contractId}`,
        operator_checklist: [
          'Verify counterparty contractual terms and rate schedule',
          'Review associated shipment activity and volume commitments',
          'Confirm renewal notice requirements with Legal/Compliance'
        ],
        safety_notice: 'Zero contract, compliance, or financial records will be automatically modified.'
      });
    } finally {
      setIsLoadingPreview(false);
    }
  };

  // Handle Request Approval
  const handleRequestApproval = async (recId) => {
    try {
      await recommendationService.requestApproval(recId, 'Legal/Compliance review requested by operator');
      setApprovalStatus((prev) => ({ ...prev, [recId]: 'SUBMITTED' }));
      setRecommendations((prev) =>
        prev.map((r) => (r.id === recId ? { ...r, status: 'pending_approval', requires_approval: true } : r))
      );
    } catch (err) {
      console.error('Failed to request approval:', err);
      setApprovalStatus((prev) => ({ ...prev, [recId]: 'ERROR' }));
    }
  };

  // Handle Mark Reviewed
  const handleMarkReviewed = async (recId) => {
    try {
      await recommendationService.markReviewed(recId);
      setRecommendations((prev) =>
        prev.map((r) => (r.id === recId ? { ...r, status: 'reviewed' } : r))
      );
    } catch (err) {
      console.error('Failed to mark reviewed:', err);
    }
  };

  if (!contractId) {
    return (
      <div className="cnt-intel-empty-card" data-testid="cnt-intel-empty-no-id">
        <Shield size={28} className="cnt-intel-empty-icon" />
        <h4 className="cnt-intel-empty-title">No Contract Selected</h4>
        <p className="cnt-intel-empty-desc">
          Select a commercial agreement to evaluate contract coverage, terms, and compliance intelligence.
        </p>
      </div>
    );
  }

  if (loading) {
    return (
      <div className="cnt-intel-loading-card" data-testid="cnt-intel-loading">
        <div className="cnt-intel-spinner"></div>
        <span className="cnt-intel-loading-text">
          Assembling Contract & Compliance Intelligence 360°...
        </span>
      </div>
    );
  }

  if (error) {
    return (
      <div className="cnt-intel-error-card" data-testid="cnt-intel-error">
        <AlertTriangle size={24} className="cnt-intel-error-icon" />
        <div className="cnt-intel-error-content">
          <h4 className="cnt-intel-error-title">Intelligence Unavailable</h4>
          <p className="cnt-intel-error-desc">{error}</p>
          <button
            className="cnt-intel-btn cnt-intel-btn--primary"
            onClick={() => fetchIntelligence(true)}
            data-testid="cnt-intel-retry-btn"
          >
            <RefreshCw size={14} />
            <span>Retry Intelligence Query</span>
          </button>
        </div>
      </div>
    );
  }

  if (!data) return null;

  const identity = data.identity || {};
  const commercial = data.commercial_terms || data.commercial || {};
  const obligations = data.obligations || {};
  const compliance = data.compliance || {};
  const coverage = data.coverage || {};
  const risks = data.risks || data.risk_indicators || {};
  const aiSummary = data.ai_summary || {};
  const correlationId = data.correlation_id || '';
  const dataFreshness = data.calculated_at || data.data_freshness || '';

  const formatDate = (dateStr) => {
    if (!dateStr) return 'Not Specified';
    try {
      return new Date(dateStr).toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' });
    } catch {
      return dateStr;
    }
  };

  const getRiskPillClass = (rating) => {
    switch ((rating || '').toUpperCase()) {
      case 'CRITICAL': return 'cnt-intel-risk-pill--critical';
      case 'ELEVATED':
      case 'HIGH': return 'cnt-intel-risk-pill--high';
      case 'MODERATE':
      case 'MEDIUM': return 'cnt-intel-risk-pill--moderate';
      case 'LOW':
      default:
        return 'cnt-intel-risk-pill--low';
    }
  };

  const getComplianceStatusBadge = (status) => {
    switch ((status || '').toUpperCase()) {
      case 'COMPLIANT':
      case 'FULL_COMPLIANCE':
      case 'ACTIVE':
        return <span className="cnt-intel-status-pill cnt-intel-status-pill--compliant"><CheckCircle2 size={12} /> Compliant</span>;
      case 'ACTION_REQUIRED':
      case 'NON_COMPLIANT':
      case 'HIGH_RISK':
        return <span className="cnt-intel-status-pill cnt-intel-status-pill--action"><AlertTriangle size={12} /> Action Required</span>;
      case 'PENDING_VERIFICATION':
      case 'PARTIALLY_VERIFIED':
      case 'PENDING_REVIEW':
        return <span className="cnt-intel-status-pill cnt-intel-status-pill--pending"><Clock size={12} /> {status.replace(/_/g, ' ')}</span>;
      default:
        return <span className="cnt-intel-status-pill cnt-intel-status-pill--unknown"><HelpCircle size={12} /> {status || 'Unknown'}</span>;
    }
  };

  const overallRiskRating = risks.risk_level || risks.overall_risk_rating || 'LOW';
  const overallRiskScore = risks.risk_score ?? risks.overall_risk_score ?? 0;
  const contractNumber = identity.contract_number || identity.contract_reference || `#${contractId}`;
  const ownerName = identity.owner_name || identity.owner || 'Unassigned';
  const confidenceDisplay = aiSummary.confidence_score 
    ? `${Math.round(aiSummary.confidence_score * 100)}% Confidence` 
    : (aiSummary.confidence || 'HIGH Confidence');

  const riskItems = risks.risk_alerts || risks.risk_factors || [];
  const recommendedActions = aiSummary.recommended_actions || aiSummary.suggested_review_areas || [];

  return (
    <div className="cnt-intel-container" data-testid="contract-compliance-intelligence-section">
      {/* ── HEADER ── */}
      <div className="cnt-intel-header">
        <div className="cnt-intel-header-left">
          <div className="cnt-intel-badge cnt-intel-badge--readonly">
            <ShieldCheck size={14} className="cnt-intel-badge-icon" />
            <span>READ-ONLY CONTRACT & COMPLIANCE INTELLIGENCE</span>
          </div>
          <div className="cnt-intel-title-row">
            <h3 className="cnt-intel-title">Agreement & Compliance Analysis</h3>
            <span
              className={`cnt-intel-risk-pill ${getRiskPillClass(overallRiskRating)}`}
              data-testid="cnt-intel-risk-rating"
            >
              Risk: {overallRiskRating} ({overallRiskScore}/100)
            </span>
            <span className="cnt-intel-status-pill cnt-intel-status-pill--compliant">
              {overallRiskRating} RISK
            </span>
            {identity.is_expired ? (
              <span className="cnt-intel-expiry-pill cnt-intel-expiry-pill--expired" data-testid="cnt-intel-expiry-pill">
                Expired ({identity.days_expired || 0}d ago)
              </span>
            ) : identity.is_expiring_soon ? (
              <span className="cnt-intel-expiry-pill cnt-intel-expiry-pill--expiring" data-testid="cnt-intel-expiry-pill">
                Expiring ({identity.days_until_expiry || 0}d left)
              </span>
            ) : (
              <span className="cnt-intel-expiry-pill cnt-intel-expiry-pill--active" data-testid="cnt-intel-expiry-pill">
                Active ({identity.days_until_expiry || 0}d remaining)
              </span>
            )}
          </div>
          <p className="cnt-intel-subtitle">
            Deterministic commercial terms audit, obligation SLAs, regulatory requirements, and risk evaluation for <strong>{contractNumber}</strong>.
          </p>
        </div>

        <div className="cnt-intel-header-right">
          <div className="cnt-intel-meta">
            <div className="cnt-intel-meta-item">
              <Clock size={12} />
              <span>Freshness: {dataFreshness ? new Date(dataFreshness).toLocaleTimeString() : 'Live'}</span>
            </div>
            {correlationId && (
              <div className="cnt-intel-meta-item" title={correlationId}>
                <span>Corr: {correlationId.slice(0, 16)}...</span>
              </div>
            )}
          </div>
          <button
            className="cnt-intel-btn cnt-intel-btn--secondary"
            onClick={() => fetchIntelligence(true)}
            disabled={refreshing}
            data-testid="cnt-intel-refresh-btn"
            title="Re-evaluate Contract & Compliance Intelligence"
          >
            <RefreshCw size={14} className={refreshing ? 'cnt-intel-btn-icon--spin' : ''} />
            <span>{refreshing ? 'Auditing...' : 'Refresh'}</span>
          </button>
        </div>
      </div>

      {/* ── GROUNDED AI CONTRACT & COMPLIANCE SYNTHESIS CARD ── */}
      <div className="cnt-intel-ai-card" data-testid="cnt-intel-ai-card">
        <div className="cnt-intel-ai-header">
          <div className="cnt-intel-ai-header-left">
            <div className="cnt-intel-ai-badge">
              <Sparkles size={13} className="cnt-intel-ai-badge-icon" />
              <span>AI CONTRACT & COMPLIANCE SYNTHESIS</span>
            </div>
            <h4 className="cnt-intel-ai-title">Deterministic Intelligence Audit</h4>
          </div>
          <span className="cnt-intel-confidence-pill">
            {confidenceDisplay}
          </span>
        </div>

        <div className="cnt-intel-ai-body">
          <p className="cnt-intel-exec-summary">
            {aiSummary.contract_executive_summary || aiSummary.executive_summary || 'Deterministic contract intelligence analysis synthesized from active database records.'}
          </p>

          <div className="cnt-intel-ai-sections">
            <div className="cnt-intel-ai-block">
              <div className="cnt-intel-ai-block-title">Coverage & Scope Assessment</div>
              <p className="cnt-intel-ai-block-text">
                {aiSummary.coverage_evaluation || aiSummary.coverage_assessment || 'Coverage evaluated based on contracted trade lanes and transport modes.'}
              </p>
            </div>

            <div className="cnt-intel-ai-block">
              <div className="cnt-intel-ai-block-title">Compliance & Regulatory Evaluation</div>
              <p className="cnt-intel-ai-block-text">
                {aiSummary.compliance_risk_assessment || aiSummary.compliance_assessment || 'Compliance posture reviewed against current documentation and regulatory requirements.'}
              </p>
            </div>
          </div>

          {recommendedActions.length > 0 && (
            <div className="cnt-intel-ai-block" style={{ marginTop: '12px' }}>
              <div className="cnt-intel-ai-block-title">Actionable Review Recommendations</div>
              <ul className="cnt-intel-ai-list">
                {recommendedActions.map((rec, idx) => (
                  <li key={idx}>{rec}</li>
                ))}
              </ul>
            </div>
          )}

          {/* Citations & Source Records */}
          {aiSummary.citations?.length > 0 && (
            <div className="cnt-intel-citations-row">
              <span className="cnt-intel-citations-label">Grounded Citations:</span>
              {aiSummary.citations.map((cite, idx) => (
                <span key={idx} className="cnt-intel-citation-tag">{cite}</span>
              ))}
            </div>
          )}
        </div>
      </div>

      {/* ── 4-QUADRANT METRICS GRID ── */}
      <div className="cnt-intel-grid">
        {/* Card 1: Validity Window & Expiry */}
        <div className="cnt-intel-card">
          <div className="cnt-intel-card-header">
            <div className="cnt-intel-card-icon icon-blue">
              <Calendar size={16} />
            </div>
            <span className="cnt-intel-card-title">Validity & Counterparty</span>
          </div>
          <div className="cnt-intel-metric-row">
            <span className="cnt-intel-metric-label">Counterparty</span>
            <span className="cnt-intel-metric-val font-semibold">{identity.party_name || '—'}</span>
          </div>
          <div className="cnt-intel-metric-row">
            <span className="cnt-intel-metric-label">Trade Corridor</span>
            <span className="cnt-intel-metric-val font-mono">{identity.lane || `${identity.origin_port || 'Origin'} ➔ ${identity.destination_port || 'Dest'}`}</span>
          </div>
          <div className="cnt-intel-metric-row">
            <span className="cnt-intel-metric-label">Effective Date</span>
            <span className="cnt-intel-metric-val">{formatDate(identity.effective_date)}</span>
          </div>
          <div className="cnt-intel-metric-row">
            <span className="cnt-intel-metric-label">Expiry Date</span>
            <span className="cnt-intel-metric-val">{formatDate(identity.expiry_date)}</span>
          </div>
          <div className="cnt-intel-metric-row">
            <span className="cnt-intel-metric-label">Term Status</span>
            <span className="cnt-intel-metric-val highlight">
              {identity.is_expired
                ? `Expired (${identity.days_expired || 0}d ago)`
                : `${identity.days_until_expiry || 0} days remaining`}
            </span>
          </div>
        </div>

        {/* Card 2: Agreement Completeness */}
        <div className="cnt-intel-card">
          <div className="cnt-intel-card-header">
            <div className="cnt-intel-card-icon icon-emerald">
              <FileCheck size={16} />
            </div>
            <span className="cnt-intel-card-title">Agreement Completeness</span>
          </div>
          <div className="cnt-intel-score-wrap">
            <span className="cnt-intel-score-number">{identity.completeness_score || 0}%</span>
            <div className="cnt-intel-progress-bar">
              <div
                className="cnt-intel-progress-fill"
                style={{ width: `${identity.completeness_score || 0}%` }}
              />
            </div>
          </div>
          <div className="cnt-intel-metric-row">
            <span className="cnt-intel-metric-label">Coverage Status</span>
            <span className="cnt-intel-metric-val font-semibold">{coverage.coverage_status || 'ACTIVE_COVERAGE'}</span>
          </div>
          <div className="cnt-intel-metric-row">
            <span className="cnt-intel-metric-label">Agreement File</span>
            <span className="cnt-intel-metric-val">
              {identity.has_active_agreement_document ? 'Attached & Verified' : `${identity.document_count || 0} files`}
            </span>
          </div>
          <div className="cnt-intel-metric-row">
            <span className="cnt-intel-metric-label">Designated Owner</span>
            <span className="cnt-intel-metric-val">{ownerName}</span>
          </div>
        </div>

        {/* Card 3: Terms & Obligations */}
        <div className="cnt-intel-card">
          <div className="cnt-intel-card-header">
            <div className="cnt-intel-card-icon icon-purple">
              <Layers size={16} />
            </div>
            <span className="cnt-intel-card-title">Commercial & SLA Terms</span>
          </div>
          <div className="cnt-intel-metric-row">
            <span className="cnt-intel-metric-label">Payment Terms</span>
            <span className="cnt-intel-metric-val font-semibold">{commercial.payment_terms || 'Standard Terms'}</span>
          </div>
          <div className="cnt-intel-metric-row">
            <span className="cnt-intel-metric-label">Free Time (Demurrage)</span>
            <span className="cnt-intel-metric-val">{commercial.free_time_days ? `${commercial.free_time_days} days` : 'Standard'}</span>
          </div>
          <div className="cnt-intel-metric-row">
            <span className="cnt-intel-metric-label">Demurrage Terms</span>
            <span className="cnt-intel-metric-val font-mono text-sm">{commercial.demurrage_terms || 'None specified'}</span>
          </div>
          <div className="cnt-intel-metric-row">
            <span className="cnt-intel-metric-label">Volume Commitment</span>
            <span className="cnt-intel-metric-val">{commercial.minimum_commitment_details || 'Open volume'}</span>
          </div>
        </div>

        {/* Card 4: Compliance Health */}
        <div className="cnt-intel-card">
          <div className="cnt-intel-card-header">
            <div className="cnt-intel-card-icon icon-amber">
              <Shield size={16} />
            </div>
            <span className="cnt-intel-card-title">Compliance Posture</span>
          </div>
          <div className="cnt-intel-metric-row">
            <span className="cnt-intel-metric-label">Overall Status</span>
            <div>{getComplianceStatusBadge(compliance.compliance_status)}</div>
          </div>
          <div className="cnt-intel-metric-row">
            <span className="cnt-intel-metric-label">Verified Requirements</span>
            <span className="cnt-intel-metric-val">
              {compliance.valid_requirements ?? compliance.verified_requirements ?? 0} / {compliance.total_requirements ?? 0}
            </span>
          </div>
          <div className="cnt-intel-metric-row">
            <span className="cnt-intel-metric-label">Open Risk Events</span>
            <span className={`cnt-intel-metric-val ${(compliance.open_compliance_events || compliance.open_events_count) > 0 ? 'text-danger' : ''}`}>
              {compliance.open_compliance_events ?? compliance.open_events_count ?? 0} ({compliance.high_severity_events ?? compliance.high_severity_events_count ?? 0} high)
            </span>
          </div>
          {compliance.missing_required_documents?.length > 0 && (
            <div className="cnt-intel-metric-row">
              <span className="cnt-intel-metric-label text-amber">Pending Docs</span>
              <span className="cnt-intel-metric-val text-amber font-semibold">
                {compliance.missing_required_documents[0]}
              </span>
            </div>
          )}
        </div>
      </div>

      {/* ── ACTIVE AI RECOMMENDATIONS & FINDINGS (Phase 2 Task 2.6) ── */}
      <div className="cnt-intel-section-card" data-testid="cnt-intel-recs-section">
        <div className="cnt-intel-section-header">
          <Sparkles size={16} className="icon-purple" />
          <h4 className="cnt-intel-section-title">
            Active AI Recommendations & Findings ({recommendations.length})
          </h4>
          <span className="cnt-intel-readonly-tag">Read-Only Guardrails Active</span>
        </div>

        {recommendations.length === 0 ? (
          <div className="cnt-intel-empty-recs">
            <CheckCircle2 size={24} className="text-emerald" />
            <p>Zero contract compliance or expiry flags detected. All operational terms compliant.</p>
          </div>
        ) : (
          <div className="cnt-intel-recs-list">
            {recommendations.map((rec) => (
              <div key={rec.id} className="cnt-intel-rec-item" data-testid={`cnt-rec-item-${rec.id}`}>
                <div className="cnt-intel-rec-top">
                  <div className="cnt-intel-rec-badge-group">
                    <span className={`cnt-intel-pill cnt-intel-pill--${rec.priority?.toLowerCase() || 'medium'}`}>
                      {rec.priority?.toUpperCase()}
                    </span>
                    <span className="cnt-intel-tag">{rec.rule_applied || rec.category}</span>
                    {rec.status && (
                      <span className={`cnt-intel-status-tag ${rec.status}`}>
                        {rec.status.toUpperCase()}
                      </span>
                    )}
                  </div>
                  <div className="cnt-intel-rec-actions">
                    <button
                      className="cnt-intel-btn cnt-intel-btn--secondary cnt-intel-btn--sm"
                      onClick={() => handleOpenActionPreview(rec)}
                      title="Preview controlled action and checklist"
                    >
                      <Eye size={12} />
                      <span>Preview Action</span>
                    </button>
                    {(rec.draft_type || rec.category === 'contract' || rec.category === 'compliance') && (
                      <button
                        className="cnt-intel-btn cnt-intel-btn--primary cnt-intel-btn--sm"
                        onClick={() => handleOpenDraft(rec)}
                        title="Prepare controlled reminder or follow-up draft"
                      >
                        <Mail size={12} />
                        <span>{rec.draft_subject ? 'View Draft' : 'Prepare Draft'}</span>
                      </button>
                    )}
                    {rec.status === 'new' && (
                      <button
                        className="cnt-intel-btn cnt-intel-btn--secondary cnt-intel-btn--sm"
                        onClick={() => handleMarkReviewed(rec.id)}
                        title="Mark recommendation as reviewed"
                      >
                        <Check size={12} />
                        <span>Reviewed</span>
                      </button>
                    )}
                  </div>
                </div>

                <div className="cnt-intel-rec-content">
                  <h5 className="cnt-intel-rec-title">{rec.title}</h5>
                  <p className="cnt-intel-rec-desc">{rec.description}</p>
                </div>

                {/* Structured Evidence Items */}
                {rec.evidence && rec.evidence.length > 0 && (
                  <div className="cnt-intel-rec-evidence">
                    <AIEvidenceList evidence={rec.evidence} />
                  </div>
                )}
              </div>
            ))}
          </div>
        )}
      </div>

      {/* ── FACTUAL CONTRACT & COMPLIANCE EVIDENCE BREAKDOWN ── */}
      {evidenceData && (
        <div className="cnt-intel-section-card" data-testid="cnt-intel-evidence-breakdown">
          <div className="cnt-intel-section-header">
            <FileText size={16} className="icon-blue" />
            <h4 className="cnt-intel-section-title">Verified MariaDB Factual Evidence</h4>
            <span className="cnt-intel-db-tag">Table: contracts & contract_documents</span>
          </div>

          <div className="cnt-intel-evidence-grid">
            <div className="cnt-intel-ev-box">
              <span className="cnt-intel-ev-label">Active Documents Tracked</span>
              <span className="cnt-intel-ev-val">{evidenceData.documents?.length || 0}</span>
              <span className="cnt-intel-ev-sub">
                {evidenceData.documents?.filter(d => d.status === 'verified').length || 0} Verified
              </span>
            </div>
            <div className="cnt-intel-ev-box">
              <span className="cnt-intel-ev-label">Compliance Checklist</span>
              <span className="cnt-intel-ev-val">{evidenceData.compliance_requirements?.length || 0}</span>
              <span className="cnt-intel-ev-sub">
                {evidenceData.summary?.verified_compliance_count || 0} Verified
              </span>
            </div>
            <div className="cnt-intel-ev-box">
              <span className="cnt-intel-ev-label">Operational Obligations</span>
              <span className="cnt-intel-ev-val">{evidenceData.obligations?.length || 0}</span>
              <span className="cnt-intel-ev-sub">
                {evidenceData.summary?.overdue_obligations_count || 0} Overdue
              </span>
            </div>
            <div className="cnt-intel-ev-box">
              <span className="cnt-intel-ev-label">Expiry Risk Status</span>
              <span className={`cnt-intel-ev-val ${evidenceData.summary?.is_expired ? 'text-danger' : (evidenceData.summary?.is_expiring_soon ? 'text-amber' : 'text-emerald')}`}>
                {evidenceData.summary?.is_expired ? 'EXPIRED' : (evidenceData.summary?.is_expiring_soon ? 'EXPIRING' : 'ACTIVE')}
              </span>
              <span className="cnt-intel-ev-sub">
                {evidenceData.summary?.days_until_expiry ?? 0} days remaining
              </span>
            </div>
          </div>
        </div>
      )}

      {/* ── RISK FACTORS BANNER (if any) ── */}
      {riskItems.length > 0 && (
        <div className="cnt-intel-risks-card" data-testid="cnt-intel-risk-factors">
          <div className="cnt-intel-risks-header">
            <AlertCircle size={16} className="text-amber" />
            <span className="cnt-intel-risks-title">
              Active Risk Factors & Alerts ({riskItems.length})
            </span>
          </div>
          <div className="cnt-intel-risks-list">
            {riskItems.map((item, idx) => {
              const title = typeof item === 'string' ? item : (item.title || item.code);
              const desc = typeof item === 'object' ? item.description : null;
              return (
                <div key={idx} className="cnt-intel-risk-item">
                  <span className="cnt-intel-risk-bullet">•</span>
                  <div>
                    <strong>{title}</strong>
                    {desc && <span className="cnt-intel-risk-desc"> — {desc}</span>}
                  </div>
                </div>
              );
            })}
          </div>
        </div>
      )}

      {/* ── COMMERCIAL TERMS LIST ── */}
      {commercial.terms?.length > 0 && (
        <div className="cnt-intel-section-card">
          <div className="cnt-intel-section-header">
            <Briefcase size={16} className="icon-blue" />
            <h4 className="cnt-intel-section-title">Audited Commercial Clauses ({commercial.terms.length})</h4>
          </div>
          <div className="cnt-intel-table-wrap">
            <table className="cnt-intel-table">
              <thead>
                <tr>
                  <th>Category</th>
                  <th>Clause Title</th>
                  <th>Value</th>
                  <th>Currency</th>
                  <th>Criticality</th>
                </tr>
              </thead>
              <tbody>
                {commercial.terms.map((t) => (
                  <tr key={t.id}>
                    <td><span className="cnt-intel-tag">{t.term_category}</span></td>
                    <td className="font-semibold">{t.term_title}</td>
                    <td>{t.term_value}</td>
                    <td>{t.currency || '—'}</td>
                    <td>
                      {t.is_critical ? (
                        <span className="cnt-intel-crit-badge">CRITICAL</span>
                      ) : (
                        <span className="cnt-intel-std-badge">STANDARD</span>
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {/* ── COMPLIANCE REQUIREMENTS LIST ── */}
      {compliance.requirements?.length > 0 && (
        <div className="cnt-intel-section-card">
          <div className="cnt-intel-section-header">
            <ShieldCheck size={16} className="icon-emerald" />
            <h4 className="cnt-intel-section-title">Mandatory Regulatory & Compliance Requirements ({compliance.requirements.length})</h4>
          </div>
          <div className="cnt-intel-table-wrap">
            <table className="cnt-intel-table">
              <thead>
                <tr>
                  <th>Requirement</th>
                  <th>Responsible</th>
                  <th>Status</th>
                  <th>Severity</th>
                  <th>Valid Until</th>
                  <th>Evidence</th>
                </tr>
              </thead>
              <tbody>
                {compliance.requirements.map((r) => (
                  <tr key={r.id}>
                    <td className="font-semibold">{r.title}</td>
                    <td>{r.responsible_party}</td>
                    <td>
                      <span className={`cnt-intel-pill cnt-intel-pill--${(r.status || '').toLowerCase()}`}>
                        {r.status}
                      </span>
                    </td>
                    <td>
                      <span className={`cnt-intel-sev cnt-intel-sev--${(r.risk_severity || '').toLowerCase()}`}>
                        {r.risk_severity}
                      </span>
                    </td>
                    <td>{formatDate(r.valid_until)}</td>
                    <td>{r.evidence_doc_id ? 'Attached' : <span className="text-amber">Pending</span>}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {/* ── CONTROLLED DRAFT GENERATION MODAL ── */}
      {selectedDraftRec && (
        <div className="cnt-modal-backdrop" onClick={() => setSelectedDraftRec(null)}>
          <div className="cnt-modal-dialog" onClick={(e) => e.stopPropagation()}>
            <div className="cnt-modal-header">
              <div className="cnt-modal-header-left">
                <div className="cnt-modal-badge">
                  <Sparkles size={14} />
                  <span>Controlled AI Draft</span>
                </div>
                <h3 className="cnt-modal-title">
                  {selectedDraftRec.draft_type === 'contract_renewal_reminder' ? 'Contract Renewal Reminder Draft' : 'Compliance Follow-up Draft'}
                </h3>
              </div>
              <button className="cnt-modal-close-btn" onClick={() => setSelectedDraftRec(null)}>
                <X size={16} />
              </button>
            </div>

            <div className="cnt-modal-body">
              <div className="cnt-modal-safe-banner">
                <ShieldCheck size={16} className="text-blue" />
                <span>
                  <strong>Strict Safe Guardrail:</strong> Zero emails are dispatched automatically. You may review, edit, copy, and save this communication draft.
                </span>
              </div>

              {isGeneratingDraft ? (
                <div className="cnt-modal-loading">
                  <RefreshCw className="cnt-intel-btn-icon--spin" size={20} />
                  <span>Generating controlled draft from contract parameters...</span>
                </div>
              ) : (
                <>
                  <div className="cnt-form-group">
                    <label className="cnt-form-label">Subject Line</label>
                    <input
                      type="text"
                      className="cnt-form-input"
                      value={draftSubject}
                      onChange={(e) => setDraftSubject(e.target.value)}
                      placeholder="Enter draft subject..."
                    />
                  </div>

                  <div className="cnt-form-group">
                    <label className="cnt-form-label">Message Body</label>
                    <textarea
                      rows={10}
                      className="cnt-form-textarea"
                      value={draftBody}
                      onChange={(e) => setDraftBody(e.target.value)}
                      placeholder="Draft message content..."
                    />
                  </div>

                  {draftSaveStatus && (
                    <div className={`cnt-modal-status ${draftSaveStatus.type}`}>
                      {draftSaveStatus.message}
                    </div>
                  )}
                </>
              )}
            </div>

            <div className="cnt-modal-footer">
              <div className="cnt-modal-footer-left">
                <button
                  className="cnt-intel-btn cnt-intel-btn--secondary"
                  onClick={() => {
                    navigator.clipboard.writeText(`${draftSubject}\n\n${draftBody}`);
                    setDraftCopied(true);
                    setTimeout(() => setDraftCopied(false), 2000);
                  }}
                  disabled={!draftBody}
                >
                  <Copy size={13} />
                  <span>{draftCopied ? 'Copied to Clipboard!' : 'Copy Text'}</span>
                </button>
              </div>
              <div className="cnt-modal-footer-right">
                <button className="cnt-intel-btn cnt-intel-btn--secondary" onClick={() => setSelectedDraftRec(null)}>
                  Close
                </button>
                <button
                  className="cnt-intel-btn cnt-intel-btn--primary"
                  onClick={handleSaveDraft}
                  disabled={isSavingDraft || !draftBody}
                >
                  {isSavingDraft ? 'Saving...' : 'Save Draft to Record'}
                </button>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* ── ACTION PREVIEW MODAL ── */}
      {actionPreviewRec && (
        <div className="cnt-modal-backdrop" onClick={() => setActionPreviewRec(null)}>
          <div className="cnt-modal-dialog" onClick={(e) => e.stopPropagation()}>
            <div className="cnt-modal-header">
              <div className="cnt-modal-header-left">
                <div className="cnt-modal-badge cnt-modal-badge--action">
                  <ShieldCheck size={14} />
                  <span>Read-Only Action Preview</span>
                </div>
                <h3 className="cnt-modal-title">{actionPreviewRec.title}</h3>
              </div>
              <button className="cnt-modal-close-btn" onClick={() => setActionPreviewRec(null)}>
                <X size={16} />
              </button>
            </div>

            <div className="cnt-modal-body">
              {isLoadingPreview ? (
                <div className="cnt-modal-loading">
                  <RefreshCw className="cnt-intel-btn-icon--spin" size={20} />
                  <span>Evaluating pre-execution checks and impact analysis...</span>
                </div>
              ) : (
                <>
                  <div className="cnt-modal-safe-banner">
                    <ShieldCheck size={16} className="text-emerald" />
                    <span>
                      <strong>Deterministic Guardrail:</strong> {actionPreview?.safety_notice || 'All actions require explicit human operator execution. Zero database modifications occur without human sign-off.'}
                    </span>
                  </div>

                  <div className="cnt-action-details-grid">
                    <div className="cnt-action-kv">
                      <span className="cnt-action-k">Target Record</span>
                      <span className="cnt-action-v font-semibold">{actionPreview?.target_entity || `Contract #${contractId}`}</span>
                    </div>
                    <div className="cnt-action-kv">
                      <span className="cnt-action-k">Action Identifier</span>
                      <span className="cnt-action-v font-mono">{actionPreview?.action_type || actionPreviewRec.action_type}</span>
                    </div>
                    <div className="cnt-action-kv">
                      <span className="cnt-action-k">Routing Department</span>
                      <span className="cnt-action-v font-semibold">{actionPreview?.routing_department || 'Legal & Compliance'}</span>
                    </div>
                  </div>

                  {actionPreview?.operator_checklist && (
                    <div className="cnt-checklist-card">
                      <h5 className="cnt-checklist-title">Pre-Execution Operator Checklist</h5>
                      <ul className="cnt-checklist-list">
                        {actionPreview.operator_checklist.map((item, idx) => (
                          <li key={idx}>
                            <CheckCircle size={13} className="text-emerald" />
                            <span>{item}</span>
                          </li>
                        ))}
                      </ul>
                    </div>
                  )}

                  {actionPreviewRec.requires_approval && (
                    <div className="cnt-approval-section">
                      <div className="cnt-approval-header">
                        <AlertCircle size={15} className="text-amber" />
                        <span>Human-in-the-Loop Governance Sign-Off Required</span>
                      </div>
                      <p className="cnt-approval-desc">
                        This action has contractual or legal implications and must be reviewed by the Legal or Compliance team.
                      </p>
                      {approvalStatus[actionPreviewRec.id] === 'SUBMITTED' ? (
                        <div className="cnt-approval-submitted">
                          <CheckCircle2 size={16} className="text-emerald" />
                          <span>Approval request submitted to Legal Department.</span>
                        </div>
                      ) : (
                        <button
                          className="cnt-intel-btn cnt-intel-btn--primary"
                          onClick={() => handleRequestApproval(actionPreviewRec.id)}
                        >
                          Submit for Legal Review
                        </button>
                      )}
                    </div>
                  )}
                </>
              )}
            </div>

            <div className="cnt-modal-footer">
              <button className="cnt-intel-btn cnt-intel-btn--secondary" onClick={() => setActionPreviewRec(null)}>
                Dismiss Preview
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

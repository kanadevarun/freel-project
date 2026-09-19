import { useState, useEffect, useCallback } from 'react';
import { Link } from 'react-router-dom';
import {
  Compass,
  ExternalLink,
  ChevronRight,
  CheckCircle2,
  AlertTriangle,
  ShieldCheck,
  Eye,
  Mail,
  Check,
  X,
  RefreshCw,
  User,
  UserCheck,
} from 'lucide-react';
import { recommendationService } from '../../services/recommendationService';
import AIConfidenceIndicator from '../ai/AIConfidenceIndicator';
import AIEvidenceList from '../ai/AIEvidenceList';
import AILoadingState from '../ai/AILoadingState';

export default function ModuleRecommendationsWidget({ sourceType, sourceId, sourceRef, onRefreshNeeded }) {
  const [recommendations, setRecommendations] = useState([]);
  const [loading, setLoading] = useState(true);

  // Modal States
  const [previewRec, setPreviewRec] = useState(null);
  const [actionPreviewData, setActionPreviewData] = useState(null);
  const [loadingPreview, setLoadingPreview] = useState(false);
  const [submittingApproval, setSubmittingApproval] = useState(false);
  const [approvalNotes, setApprovalNotes] = useState('');

  const [draftRec, setDraftRec] = useState(null);
  const [draftSubject, setDraftSubject] = useState('');
  const [draftBody, setDraftBody] = useState('');
  const [generatingDraft, setGeneratingDraft] = useState(false);
  const [savingDraft, setSavingDraft] = useState(false);

  // Task 2.4 Assign & Dismiss States
  const [assignModalRec, setAssignModalRec] = useState(null);
  const [assigneeName, setAssigneeName] = useState('');
  const [submittingAssign, setSubmittingAssign] = useState(false);

  const [dismissModalRec, setDismissModalRec] = useState(null);
  const [dismissReason, setDismissReason] = useState('');
  const [submittingDismiss, setSubmittingDismiss] = useState(false);

  const [statusMsg, setStatusMsg] = useState(null);

  const fetchWidgetRecs = useCallback(() => {
    if (!sourceType || !sourceId) {
      setLoading(false);
      return;
    }
    recommendationService.listBySource(sourceType, sourceId)
      .then((res) => {
        const items = res?.data?.items || res?.items || res?.recommendations || [];
        setRecommendations(items);
      })
      .catch(() => {})
      .finally(() => setLoading(false));
  }, [sourceType, sourceId]);

  useEffect(() => {
    fetchWidgetRecs();
  }, [fetchWidgetRecs]);

  const handleMarkReviewed = async (rec) => {
    try {
      const res = await recommendationService.markReviewed(rec.id);
      if (res?.success) {
        setStatusMsg(`Recommendation #${rec.id} marked as reviewed.`);
        fetchWidgetRecs();
        if (onRefreshNeeded) onRefreshNeeded();
      }
    } catch (err) {
      alert(`Review failed: ${err.message}`);
    }
  };

  const handleOpenPreview = async (rec) => {
    setPreviewRec(rec);
    setActionPreviewData(null);
    setLoadingPreview(true);
    setApprovalNotes('');
    try {
      const res = await recommendationService.getActionPreview(rec.id);
      if (res?.success && res.action_preview) {
        setActionPreviewData(res.action_preview);
      }
    } catch (err) {
      alert(`Preview failed: ${err.message}`);
    } finally {
      setLoadingPreview(false);
    }
  };

  const handleSubmitApproval = async () => {
    if (!previewRec) return;
    setSubmittingApproval(true);
    try {
      const res = await recommendationService.requestApproval(previewRec.id, approvalNotes.trim());
      if (res?.success) {
        setStatusMsg(`Approval request #${res.recommendation?.approval_id || 'HITL'} submitted.`);
        if (actionPreviewData) {
          setActionPreviewData({ ...actionPreviewData, approval_id: res.recommendation?.approval_id });
        }
        fetchWidgetRecs();
        if (onRefreshNeeded) onRefreshNeeded();
      }
    } catch (err) {
      alert(`Approval submission failed: ${err.message}`);
    } finally {
      setSubmittingApproval(false);
    }
  };

  const handleOpenDraft = async (rec) => {
    setDraftRec(rec);
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
          setDraftRec(res.recommendation);
          fetchWidgetRecs();
        }
      } catch (err) {
        alert(`Draft generation failed: ${err.message}`);
      } finally {
        setGeneratingDraft(false);
      }
    }
  };

  const handleSaveDraft = async () => {
    if (!draftRec) return;
    setSavingDraft(true);
    try {
      const res = await recommendationService.saveDraft(draftRec.id, draftSubject, draftBody);
      if (res?.success) {
        setStatusMsg(`Draft saved for ${draftRec.source_reference}.`);
        setDraftRec(null);
        fetchWidgetRecs();
      }
    } catch (err) {
      alert(`Save draft failed: ${err.message}`);
    } finally {
      setSavingDraft(false);
    }
  };

  const handleAssign = async () => {
    if (!assignModalRec || !assigneeName.trim()) return;
    setSubmittingAssign(true);
    try {
      const res = await recommendationService.assignRecommendation(assignModalRec.id, 0, assigneeName.trim());
      if (res?.success) {
        setStatusMsg(`Recommendation #${assignModalRec.id} assigned to ${assigneeName.trim()}.`);
        setAssignModalRec(null);
        setAssigneeName('');
        fetchWidgetRecs();
        if (onRefreshNeeded) onRefreshNeeded();
      }
    } catch (err) {
      alert(`Assignment failed: ${err.message}`);
    } finally {
      setSubmittingAssign(false);
    }
  };

  const handleDismiss = async () => {
    if (!dismissModalRec || !dismissReason.trim()) return;
    setSubmittingDismiss(true);
    try {
      const res = await recommendationService.dismissRecommendation(dismissModalRec.id, dismissReason.trim());
      if (res?.success) {
        setStatusMsg(`Recommendation #${dismissModalRec.id} dismissed.`);
        setDismissModalRec(null);
        setDismissReason('');
        fetchWidgetRecs();
        if (onRefreshNeeded) onRefreshNeeded();
      }
    } catch (err) {
      alert(`Dismissal failed: ${err.message}`);
    } finally {
      setSubmittingDismiss(false);
    }
  };

  if (loading) {
    return null;
  }

  if (recommendations.length === 0) {
    return (
      <div style={{
        background: '#ffffff',
        border: '1px dashed #cbd5e1',
        borderRadius: '12px',
        padding: '16px 20px',
        marginTop: '16px',
        marginBottom: '16px',
        color: '#64748b',
        fontSize: '0.875rem',
        display: 'flex',
        alignItems: 'center',
        gap: '10px',
      }}>
        <Compass size={16} color="#94a3b8" />
        <span>No active AI recommendations for this {sourceType}. All signals are verified within normal operating parameters.</span>
      </div>
    );
  }

  return (
    <div style={{
      background: '#ffffff',
      border: '1px solid #e2e8f0',
      borderRadius: '12px',
      padding: '18px 20px',
      marginTop: '16px',
      marginBottom: '16px',
      boxShadow: '0 1px 3px rgba(15, 23, 42, 0.04)',
    }}>
      <div style={{
        display: 'flex',
        justifyContent: 'space-between',
        alignItems: 'center',
        marginBottom: '14px',
        borderBottom: '1px solid #f1f5f9',
        paddingBottom: '10px',
      }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
          <Compass size={18} color="#2563eb" />
          <h4 style={{ margin: 0, fontSize: '0.9375rem', fontWeight: 700, color: '#0f172a' }}>
            Workflow Assistant Intelligence ({recommendations.length})
          </h4>
        </div>
        <div style={{ display: 'flex', alignItems: 'center', gap: '12px' }}>
          <Link
            to="/dashboard/recommendations"
            style={{
              fontSize: '0.75rem',
              fontWeight: 600,
              color: '#2563eb',
              textDecoration: 'none',
              display: 'flex',
              alignItems: 'center',
              gap: '4px',
            }}
          >
            <span>Open Center</span>
            <ExternalLink size={12} />
          </Link>
        </div>
      </div>

      {statusMsg && (
        <div style={{
          background: '#ecfdf5',
          border: '1px solid #a7f3d0',
          color: '#065f46',
          padding: '8px 12px',
          borderRadius: '6px',
          fontSize: '0.75rem',
          marginBottom: '12px',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
        }}>
          <span>{statusMsg}</span>
          <button
            type="button"
            onClick={() => setStatusMsg(null)}
            style={{ background: 'transparent', border: 'none', cursor: 'pointer', color: '#065f46' }}
          >
            <X size={13} />
          </button>
        </div>
      )}

      <div style={{ display: 'flex', flexDirection: 'column', gap: '12px' }}>
        {recommendations.map((rec) => {
          const isMissingInfo = rec.rule_applied === 'RFQ_MISSING_INFO';
          const isMarginAlert = rec.rule_applied === 'QUOTATION_MARGIN_ALERT' || rec.category === 'pricing';
          const isExpiryRisk = rec.rule_applied === 'QUOTATION_EXPIRY_RISK';
          const isPendingApproval = rec.rule_applied === 'QUOTATION_PENDING_APPROVAL' || rec.requires_approval;
          const isDelayedMilestone = rec.rule_applied === 'SHIPMENT_MILESTONE_DELAYED' || rec.milestone_id;
          const isException = rec.rule_applied === 'UNRESOLVED_SHIPMENT_EXCEPTIONS' || rec.exception_id;
          const isMissingOpsInfo = rec.rule_applied === 'SHIPMENT_MISSING_OPERATIONAL_INFO';
          const isInactiveTracking = rec.rule_applied === 'SHIPMENT_INACTIVE_TRACKING';
          const isOverdueDelivery = rec.rule_applied === 'SHIPMENT_OVERDUE_DELIVERY';

          return (
            <div
              key={rec.id}
              style={{
                background: '#f8fafc',
                border: '1px solid #e2e8f0',
                borderLeft: rec.priority === 'critical' ? '3px solid #dc2626' : rec.priority === 'high' ? '3px solid #d97706' : '3px solid #2563eb',
                borderRadius: '8px',
                padding: '14px 16px',
                display: 'flex',
                flexDirection: 'column',
                gap: '8px',
              }}
            >
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', flexWrap: 'wrap', gap: '6px' }}>
                <div style={{ display: 'flex', alignItems: 'center', gap: '6px', flexWrap: 'wrap' }}>
                  <span style={{
                    fontSize: '0.6875rem',
                    fontWeight: 700,
                    textTransform: 'uppercase',
                    padding: '2px 6px',
                    borderRadius: '4px',
                    background: rec.priority === 'critical' ? '#fef2f2' : rec.priority === 'high' ? '#fffbeb' : '#eff6ff',
                    color: rec.priority === 'critical' ? '#991b1b' : rec.priority === 'high' ? '#92400e' : '#1e40af',
                  }}>
                    {rec.priority} Priority
                  </span>

                  {isMissingInfo && (
                    <span style={{ fontSize: '0.6875rem', fontWeight: 700, padding: '2px 6px', borderRadius: '4px', background: '#fee2e2', color: '#991b1b' }}>
                      Missing Information
                    </span>
                  )}
                  {isMarginAlert && (
                    <span style={{ fontSize: '0.6875rem', fontWeight: 700, padding: '2px 6px', borderRadius: '4px', background: '#fef3c7', color: '#92400e' }}>
                      Pricing / Margin Warning
                    </span>
                  )}
                  {isExpiryRisk && (
                    <span style={{ fontSize: '0.6875rem', fontWeight: 700, padding: '2px 6px', borderRadius: '4px', background: '#ffedd5', color: '#9a3412' }}>
                      Expiring Soon
                    </span>
                  )}
                  {isPendingApproval && (
                    <span style={{ fontSize: '0.6875rem', fontWeight: 700, padding: '2px 6px', borderRadius: '4px', background: '#e0e7ff', color: '#4338ca' }}>
                      Approval Required
                    </span>
                  )}
                  {isDelayedMilestone && (
                    <span style={{ fontSize: '0.6875rem', fontWeight: 700, padding: '2px 6px', borderRadius: '4px', background: '#fee2e2', color: '#b91c1c' }}>
                      Delayed Milestone
                    </span>
                  )}
                  {isException && (
                    <span style={{ fontSize: '0.6875rem', fontWeight: 700, padding: '2px 6px', borderRadius: '4px', background: '#fef2f2', color: '#991b1b' }}>
                      Active Exception
                    </span>
                  )}
                  {isMissingOpsInfo && (
                    <span style={{ fontSize: '0.6875rem', fontWeight: 700, padding: '2px 6px', borderRadius: '4px', background: '#fef3c7', color: '#92400e' }}>
                      Missing Ops Info
                    </span>
                  )}
                  {isInactiveTracking && (
                    <span style={{ fontSize: '0.6875rem', fontWeight: 700, padding: '2px 6px', borderRadius: '4px', background: '#f1f5f9', color: '#475569' }}>
                      Inactive Tracking
                    </span>
                  )}
                  {isOverdueDelivery && (
                    <span style={{ fontSize: '0.6875rem', fontWeight: 700, padding: '2px 6px', borderRadius: '4px', background: '#fee2e2', color: '#dc2626' }}>
                      Overdue Delivery
                    </span>
                  )}

                  <span style={{ fontSize: '0.75rem', color: '#64748b', fontWeight: 500 }}>
                    Status: <strong style={{ textTransform: 'capitalize', color: '#334155' }}>{rec.status}</strong>
                  </span>
                </div>

                <AIConfidenceIndicator confidence={rec.confidence} score={rec.confidence_score} />
              </div>

              <div style={{ fontSize: '0.9375rem', fontWeight: 600, color: '#0f172a' }}>
                {rec.title}
              </div>

              <div style={{ fontSize: '0.8125rem', color: '#475569', lineHeight: 1.45 }}>
                {rec.description}
              </div>

              <div style={{
                background: '#ffffff',
                border: '1px solid #e2e8f0',
                borderRadius: '6px',
                padding: '8px 12px',
                display: 'flex',
                alignItems: 'center',
                gap: '8px',
                fontSize: '0.8125rem',
                color: '#1e293b',
              }}>
                <ChevronRight size={14} color="#2563eb" />
                <span><strong>Recommended Step:</strong> {rec.recommended_action}</span>
              </div>

              <div style={{
                display: 'flex',
                justifyContent: 'space-between',
                alignItems: 'center',
                flexWrap: 'wrap',
                gap: '8px',
                marginTop: '4px',
                borderTop: '1px solid #f1f5f9',
                paddingTop: '8px',
              }}>
                <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                  <button
                    type="button"
                    style={{
                      background: '#ffffff',
                      border: '1px solid #cbd5e1',
                      borderRadius: '6px',
                      padding: '4px 10px',
                      fontSize: '0.75rem',
                      fontWeight: 600,
                      color: '#0f172a',
                      cursor: 'pointer',
                      display: 'flex',
                      alignItems: 'center',
                      gap: '4px',
                    }}
                    onClick={() => handleOpenPreview(rec)}
                  >
                    <Eye size={12} />
                    Action Preview
                  </button>

                  <button
                    type="button"
                    style={{
                      background: '#eff6ff',
                      border: '1px solid #bfdbfe',
                      borderRadius: '6px',
                      padding: '4px 10px',
                      fontSize: '0.75rem',
                      fontWeight: 600,
                      color: '#1d4ed8',
                      cursor: 'pointer',
                      display: 'flex',
                      alignItems: 'center',
                      gap: '4px',
                    }}
                    onClick={() => handleOpenDraft(rec)}
                  >
                    <Mail size={12} />
                    {rec.draft_status === 'EDITED' ? 'Edit Draft' : rec.draft_status === 'DRAFTED' ? 'Review Draft' : 'Draft Clarification'}
                  </button>

                  <button
                    type="button"
                    style={{
                      background: '#ffffff',
                      border: '1px solid #cbd5e1',
                      borderRadius: '6px',
                      padding: '4px 10px',
                      fontSize: '0.75rem',
                      fontWeight: 600,
                      color: '#0f172a',
                      cursor: 'pointer',
                      display: 'flex',
                      alignItems: 'center',
                      gap: '4px',
                    }}
                    onClick={() => {
                      setAssignModalRec(rec);
                      setAssigneeName('');
                    }}
                  >
                    <User size={12} />
                    {rec.assignee_name ? `Assigned: ${rec.assignee_name}` : 'Assign'}
                  </button>

                  <button
                    type="button"
                    style={{
                      background: '#ffffff',
                      border: '1px solid #fecaca',
                      borderRadius: '6px',
                      padding: '4px 10px',
                      fontSize: '0.75rem',
                      fontWeight: 600,
                      color: '#dc2626',
                      cursor: 'pointer',
                      display: 'flex',
                      alignItems: 'center',
                      gap: '4px',
                    }}
                    onClick={() => {
                      setDismissModalRec(rec);
                      setDismissReason('');
                    }}
                  >
                    <X size={12} />
                    Dismiss
                  </button>

                  {rec.status === 'new' && (
                    <button
                      type="button"
                      style={{
                        background: '#ffffff',
                        border: '1px solid #cbd5e1',
                        borderRadius: '6px',
                        padding: '4px 10px',
                        fontSize: '0.75rem',
                        fontWeight: 600,
                        color: '#0f172a',
                        cursor: 'pointer',
                        display: 'flex',
                        alignItems: 'center',
                        gap: '4px',
                      }}
                      onClick={() => handleMarkReviewed(rec)}
                    >
                      <Check size={12} />
                      Mark Reviewed
                    </button>
                  )}
                </div>

                <Link
                  to="/dashboard/recommendations"
                  style={{
                    color: '#2563eb',
                    fontWeight: 600,
                    fontSize: '0.75rem',
                    textDecoration: 'none',
                    display: 'flex',
                    alignItems: 'center',
                    gap: '4px',
                  }}
                >
                  Manage in Center →
                </Link>
              </div>
            </div>
          );
        })}
      </div>

      {/* Action Preview Modal */}
      {previewRec && (
        <div style={{
          position: 'fixed',
          top: 0,
          left: 0,
          right: 0,
          bottom: 0,
          background: 'rgba(15, 23, 42, 0.45)',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          zIndex: 1000,
          padding: '16px',
        }} onClick={() => setPreviewRec(null)}>
          <div style={{
            background: '#ffffff',
            borderRadius: '12px',
            boxShadow: '0 20px 25px -5px rgba(0, 0, 0, 0.1)',
            maxWidth: '600px',
            width: '100%',
            padding: '20px',
            maxHeight: '90vh',
            overflowY: 'auto',
          }} onClick={(e) => e.stopPropagation()}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '14px', borderBottom: '1px solid #f1f5f9', paddingBottom: '10px' }}>
              <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                <Eye size={18} color="#2563eb" />
                <h3 style={{ margin: 0, fontSize: '1rem', fontWeight: 700, color: '#0f172a' }}>Controlled Action Preview</h3>
              </div>
              <button
                type="button"
                style={{ background: 'transparent', border: 'none', cursor: 'pointer', color: '#64748b' }}
                onClick={() => setPreviewRec(null)}
              >
                <X size={18} />
              </button>
            </div>

            {loadingPreview ? (
              <AILoadingState title="Verifying safety constraints..." count={2} />
            ) : actionPreviewData ? (
              <div>
                <div style={{
                  background: '#f8fafc',
                  border: '1px solid #e2e8f0',
                  borderRadius: '8px',
                  padding: '12px',
                  marginBottom: '14px',
                }}>
                  <div style={{ fontSize: '0.6875rem', fontWeight: 700, color: '#64748b', textTransform: 'uppercase', marginBottom: '4px' }}>
                    Proposed Action
                  </div>
                  <div style={{ fontSize: '0.875rem', fontWeight: 700, color: '#0f172a', marginBottom: '8px' }}>
                    {actionPreviewData.proposed_action}
                  </div>
                  <div style={{ fontSize: '0.6875rem', fontWeight: 700, color: '#64748b', textTransform: 'uppercase', marginBottom: '4px' }}>
                    Expected Effect
                  </div>
                  <div style={{ fontSize: '0.8125rem', color: '#334155' }}>
                    {actionPreviewData.expected_effect}
                  </div>
                </div>

                {actionPreviewData.evidence && actionPreviewData.evidence.length > 0 && (
                  <div style={{ marginBottom: '14px' }}>
                    <div style={{ fontSize: '0.75rem', fontWeight: 700, color: '#1e293b', marginBottom: '6px' }}>
                      Grounding Evidence
                    </div>
                    <div style={{ maxHeight: '140px', overflowY: 'auto', border: '1px solid #e2e8f0', borderRadius: '6px' }}>
                      <AIEvidenceList evidence={actionPreviewData.evidence} />
                    </div>
                  </div>
                )}

                {actionPreviewData.required_approval && !actionPreviewData.approval_id && (
                  <div style={{
                    background: '#fffbeb',
                    border: '1px solid #fde68a',
                    borderRadius: '8px',
                    padding: '12px',
                    marginBottom: '14px',
                  }}>
                    <div style={{ display: 'flex', alignItems: 'center', gap: '6px', fontSize: '0.8125rem', fontWeight: 600, color: '#92400e', marginBottom: '6px' }}>
                      <ShieldCheck size={14} />
                      Commercial Approval Required
                    </div>
                    <textarea
                      rows={2}
                      style={{
                        width: '100%',
                        fontFamily: 'inherit',
                        fontSize: '0.75rem',
                        padding: '6px',
                        borderRadius: '6px',
                        border: '1px solid #fcd34d',
                        marginBottom: '8px',
                      }}
                      placeholder="Optional notes for approver..."
                      value={approvalNotes}
                      onChange={(e) => setApprovalNotes(e.target.value)}
                    />
                    <button
                      type="button"
                      style={{
                        background: '#d97706',
                        border: '1px solid #b45309',
                        color: '#ffffff',
                        padding: '6px 12px',
                        borderRadius: '6px',
                        fontSize: '0.75rem',
                        fontWeight: 600,
                        cursor: 'pointer',
                      }}
                      disabled={submittingApproval}
                      onClick={handleSubmitApproval}
                    >
                      {submittingApproval ? 'Submitting...' : 'Submit to Approvals System'}
                    </button>
                  </div>
                )}

                <div style={{ display: 'flex', justifyContent: 'flex-end', marginTop: '12px' }}>
                  <button
                    type="button"
                    style={{
                      background: '#f1f5f9',
                      border: '1px solid #cbd5e1',
                      borderRadius: '6px',
                      padding: '6px 14px',
                      fontSize: '0.8125rem',
                      fontWeight: 600,
                      cursor: 'pointer',
                    }}
                    onClick={() => setPreviewRec(null)}
                  >
                    Close
                  </button>
                </div>
              </div>
            ) : null}
          </div>
        </div>
      )}

      {/* Draft Modal */}
      {draftRec && (
        <div style={{
          position: 'fixed',
          top: 0,
          left: 0,
          right: 0,
          bottom: 0,
          background: 'rgba(15, 23, 42, 0.45)',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          zIndex: 1000,
          padding: '16px',
        }} onClick={() => setDraftRec(null)}>
          <div style={{
            background: '#ffffff',
            borderRadius: '12px',
            boxShadow: '0 20px 25px -5px rgba(0, 0, 0, 0.1)',
            maxWidth: '600px',
            width: '100%',
            padding: '20px',
            maxHeight: '90vh',
            overflowY: 'auto',
          }} onClick={(e) => e.stopPropagation()}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '14px', borderBottom: '1px solid #f1f5f9', paddingBottom: '10px' }}>
              <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                <Mail size={18} color="#2563eb" />
                <h3 style={{ margin: 0, fontSize: '1rem', fontWeight: 700, color: '#0f172a' }}>Editable Message Draft</h3>
              </div>
              <button
                type="button"
                style={{ background: 'transparent', border: 'none', cursor: 'pointer', color: '#64748b' }}
                onClick={() => setDraftRec(null)}
              >
                <X size={18} />
              </button>
            </div>

            {generatingDraft ? (
              <AILoadingState title="Synthesizing verified draft..." count={2} />
            ) : (
              <div>
                <div style={{ marginBottom: '12px' }}>
                  <label style={{ display: 'block', fontSize: '0.75rem', fontWeight: 600, color: '#475569', marginBottom: '4px' }}>
                    Subject Line
                  </label>
                  <input
                    type="text"
                    style={{
                      width: '100%',
                      fontFamily: 'inherit',
                      fontSize: '0.8125rem',
                      padding: '8px',
                      borderRadius: '6px',
                      border: '1px solid #cbd5e1',
                    }}
                    value={draftSubject}
                    onChange={(e) => setDraftSubject(e.target.value)}
                  />
                </div>

                <div style={{ marginBottom: '12px' }}>
                  <label style={{ display: 'block', fontSize: '0.75rem', fontWeight: 600, color: '#475569', marginBottom: '4px' }}>
                    Body (Editable)
                  </label>
                  <textarea
                    rows={7}
                    style={{
                      width: '100%',
                      fontFamily: 'inherit',
                      fontSize: '0.8125rem',
                      lineHeight: 1.45,
                      padding: '8px',
                      borderRadius: '6px',
                      border: '1px solid #cbd5e1',
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
                  marginBottom: '14px',
                }}>
                  Controlled Assistant mode: Draft is saved locally and cannot be sent automatically.
                </div>

                <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '8px' }}>
                  <button
                    type="button"
                    style={{
                      background: '#f1f5f9',
                      border: '1px solid #cbd5e1',
                      borderRadius: '6px',
                      padding: '6px 14px',
                      fontSize: '0.8125rem',
                      fontWeight: 600,
                      cursor: 'pointer',
                    }}
                    onClick={() => setDraftRec(null)}
                  >
                    Cancel
                  </button>
                  <button
                    type="button"
                    style={{
                      background: '#2563eb',
                      border: '1px solid #1d4ed8',
                      color: '#ffffff',
                      borderRadius: '6px',
                      padding: '6px 14px',
                      fontSize: '0.8125rem',
                      fontWeight: 600,
                      cursor: 'pointer',
                      display: 'flex',
                      alignItems: 'center',
                      gap: '4px',
                    }}
                    disabled={savingDraft || !draftSubject.trim() || !draftBody.trim()}
                    onClick={handleSaveDraft}
                  >
                    <Check size={14} />
                    {savingDraft ? 'Saving...' : 'Save Draft'}
                  </button>
                </div>
              </div>
            )}
          </div>
        </div>
      )}

      {/* Assign Modal */}
      {assignModalRec && (
        <div style={{
          position: 'fixed',
          top: 0,
          left: 0,
          right: 0,
          bottom: 0,
          background: 'rgba(15, 23, 42, 0.45)',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          zIndex: 1000,
          padding: '16px',
        }} onClick={() => setAssignModalRec(null)}>
          <div style={{
            background: '#ffffff',
            borderRadius: '12px',
            boxShadow: '0 20px 25px -5px rgba(0, 0, 0, 0.1)',
            maxWidth: '460px',
            width: '100%',
            padding: '20px',
          }} onClick={(e) => e.stopPropagation()}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '14px', borderBottom: '1px solid #f1f5f9', paddingBottom: '10px' }}>
              <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                <User size={18} color="#2563eb" />
                <h4 style={{ margin: 0, fontSize: '1rem', fontWeight: 600, color: '#0f172a' }}>
                  Assign Operational Owner
                </h4>
              </div>
              <button
                type="button"
                style={{ background: 'transparent', border: 'none', cursor: 'pointer', color: '#64748b' }}
                onClick={() => setAssignModalRec(null)}
              >
                <X size={16} />
              </button>
            </div>

            <p style={{ fontSize: '0.8125rem', color: '#475569', marginBottom: '12px' }}>
              Assign recommendation <strong>#{assignModalRec.id}</strong> ({assignModalRec.source_reference}) to an operations specialist for follow-up.
            </p>

            <div style={{ marginBottom: '16px' }}>
              <label style={{ display: 'block', fontSize: '0.75rem', fontWeight: 600, color: '#334155', marginBottom: '4px' }}>
                Assignee Name / Role *
              </label>
              <input
                type="text"
                placeholder="e.g. Varun Kanade, Ocean Desk, Dispatch Specialist"
                value={assigneeName}
                onChange={(e) => setAssigneeName(e.target.value)}
                style={{
                  width: '100%',
                  padding: '8px 12px',
                  borderRadius: '6px',
                  border: '1px solid #cbd5e1',
                  fontSize: '0.8125rem',
                }}
              />
            </div>

            <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '8px' }}>
              <button
                type="button"
                style={{
                  background: '#f1f5f9',
                  border: '1px solid #cbd5e1',
                  borderRadius: '6px',
                  padding: '6px 14px',
                  fontSize: '0.8125rem',
                  fontWeight: 600,
                  cursor: 'pointer',
                }}
                onClick={() => setAssignModalRec(null)}
              >
                Cancel
              </button>
              <button
                type="button"
                style={{
                  background: '#2563eb',
                  border: '1px solid #1d4ed8',
                  color: '#ffffff',
                  borderRadius: '6px',
                  padding: '6px 14px',
                  fontSize: '0.8125rem',
                  fontWeight: 600,
                  cursor: 'pointer',
                  display: 'flex',
                  alignItems: 'center',
                  gap: '4px',
                }}
                disabled={submittingAssign || !assigneeName.trim()}
                onClick={handleAssign}
              >
                <UserCheck size={14} />
                {submittingAssign ? 'Assigning...' : 'Confirm Assignment'}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Dismiss Modal */}
      {dismissModalRec && (
        <div style={{
          position: 'fixed',
          top: 0,
          left: 0,
          right: 0,
          bottom: 0,
          background: 'rgba(15, 23, 42, 0.45)',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          zIndex: 1000,
          padding: '16px',
        }} onClick={() => setDismissModalRec(null)}>
          <div style={{
            background: '#ffffff',
            borderRadius: '12px',
            boxShadow: '0 20px 25px -5px rgba(0, 0, 0, 0.1)',
            maxWidth: '460px',
            width: '100%',
            padding: '20px',
          }} onClick={(e) => e.stopPropagation()}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '14px', borderBottom: '1px solid #f1f5f9', paddingBottom: '10px' }}>
              <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                <X size={18} color="#dc2626" />
                <h4 style={{ margin: 0, fontSize: '1rem', fontWeight: 600, color: '#0f172a' }}>
                  Dismiss Recommendation
                </h4>
              </div>
              <button
                type="button"
                style={{ background: 'transparent', border: 'none', cursor: 'pointer', color: '#64748b' }}
                onClick={() => setDismissModalRec(null)}
              >
                <X size={16} />
              </button>
            </div>

            <p style={{ fontSize: '0.8125rem', color: '#475569', marginBottom: '12px' }}>
              Dismiss recommendation <strong>#{dismissModalRec.id}</strong> ({dismissModalRec.source_reference}). An audited justification is mandatory.
            </p>

            <div style={{ marginBottom: '16px' }}>
              <label style={{ display: 'block', fontSize: '0.75rem', fontWeight: 600, color: '#334155', marginBottom: '4px' }}>
                Dismissal Reason *
              </label>
              <textarea
                rows={3}
                placeholder="State the operational justification for dismissing this signal..."
                value={dismissReason}
                onChange={(e) => setDismissReason(e.target.value)}
                style={{
                  width: '100%',
                  padding: '8px 12px',
                  borderRadius: '6px',
                  border: '1px solid #cbd5e1',
                  fontSize: '0.8125rem',
                  resize: 'vertical',
                }}
              />
            </div>

            <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '8px' }}>
              <button
                type="button"
                style={{
                  background: '#f1f5f9',
                  border: '1px solid #cbd5e1',
                  borderRadius: '6px',
                  padding: '6px 14px',
                  fontSize: '0.8125rem',
                  fontWeight: 600,
                  cursor: 'pointer',
                }}
                onClick={() => setDismissModalRec(null)}
              >
                Cancel
              </button>
              <button
                type="button"
                style={{
                  background: '#dc2626',
                  border: '1px solid #b91c1c',
                  color: '#ffffff',
                  borderRadius: '6px',
                  padding: '6px 14px',
                  fontSize: '0.8125rem',
                  fontWeight: 600,
                  cursor: 'pointer',
                  display: 'flex',
                  alignItems: 'center',
                  gap: '4px',
                }}
                disabled={submittingDismiss || !dismissReason.trim()}
                onClick={handleDismiss}
              >
                <Check size={14} />
                {submittingDismiss ? 'Dismissing...' : 'Confirm Dismissal'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

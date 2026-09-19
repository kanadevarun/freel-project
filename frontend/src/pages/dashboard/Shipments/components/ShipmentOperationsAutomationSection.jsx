import React, { useState, useEffect, useCallback } from 'react';
import {
  ShieldAlert,
  AlertTriangle,
  Clock,
  FileCheck,
  Send,
  Sparkles,
  RefreshCw,
  CheckCircle2,
  AlertCircle,
  FileText,
  Mail,
  ChevronRight,
  ExternalLink,
  Info,
  Layers,
  ArrowRight
} from 'lucide-react';
import shipmentOperationsAutomationService from '../../../../services/shipmentOperationsAutomationService';
import './ShipmentOperationsAutomationSection.css';

export default function ShipmentOperationsAutomationSection({ shipmentId, onRefreshShipment }) {
  const [overview, setOverview] = useState(null);
  const [loading, setLoading] = useState(true);
  const [analyzing, setAnalyzing] = useState(false);
  const [generatingDraft, setGeneratingDraft] = useState(false);
  const [submittingApproval, setSubmittingApproval] = useState(false);
  const [error, setError] = useState(null);
  const [feedback, setFeedback] = useState(null);

  // Drafting Studio state
  const [draftType, setDraftType] = useState('CUSTOMER_UPDATE');
  const [userInstructions, setUserInstructions] = useState('');
  const [currentDraft, setCurrentDraft] = useState(null);
  const [editableSubject, setEditableSubject] = useState('');
  const [editableWording, setEditableWording] = useState('');
  const [editableNotes, setEditableNotes] = useState('');
  const [editableEmail, setEditableEmail] = useState('');
  const [editableRecipientName, setEditableRecipientName] = useState('');
  const [approvalReason, setApprovalReason] = useState('Customer update regarding port operations status');

  const fetchOverview = useCallback(async () => {
    if (!shipmentId) return;
    setLoading(true);
    setError(null);
    try {
      const res = await shipmentOperationsAutomationService.getOverview(shipmentId);
      const data = res?.data?.data || res?.data;
      setOverview(data);

      if (data?.latest_draft) {
        syncDraftState(data.latest_draft);
      }
    } catch (err) {
      console.error('Error fetching shipment operations automation overview:', err);
      setError(err?.response?.data?.message || err.message || 'Failed to load operations automation overview');
    } finally {
      setLoading(false);
    }
  }, [shipmentId]);

  useEffect(() => {
    fetchOverview();
  }, [fetchOverview]);

  const syncDraftState = (draft) => {
    setCurrentDraft(draft);
    setEditableSubject(draft.subject || '');
    setEditableWording(draft.customer_wording || '');
    setEditableNotes(draft.internal_notes || '');
    setEditableEmail(draft.recipient_email || '');
    setEditableRecipientName(draft.recipient_name || '');
  };

  const handleAnalyzeRisks = async () => {
    setAnalyzing(true);
    setError(null);
    try {
      const res = await shipmentOperationsAutomationService.analyzeRisks(shipmentId);
      setFeedback({ type: 'success', message: 'Operational risk analysis refreshed successfully.' });
      setTimeout(() => setFeedback(null), 4000);
      await fetchOverview();
      if (onRefreshShipment) onRefreshShipment();
    } catch (err) {
      console.error('Failed to run shipment risk analysis:', err);
      setError(err?.response?.data?.message || err.message || 'Risk analysis failed');
    } finally {
      setAnalyzing(false);
    }
  };

  const handleGenerateDraft = async () => {
    setGeneratingDraft(true);
    setError(null);
    try {
      const res = await shipmentOperationsAutomationService.generateDraft(shipmentId, {
        draft_type: draftType,
        recipient_name: editableRecipientName,
        recipient_email: editableEmail,
        user_instructions: userInstructions
      });
      const newDraft = res?.data?.data || res?.data;
      syncDraftState(newDraft);
      setFeedback({ type: 'success', message: `Generated ${draftType.replace('_', ' ')} draft.` });
      setTimeout(() => setFeedback(null), 4000);
      await fetchOverview();
    } catch (err) {
      console.error('Failed to generate communication draft:', err);
      setError(err?.response?.data?.message || err.message || 'Draft generation failed');
    } finally {
      setGeneratingDraft(false);
    }
  };

  const handleSaveDraft = async () => {
    if (!currentDraft?.id) return;
    try {
      const res = await shipmentOperationsAutomationService.updateDraft(shipmentId, currentDraft.id, {
        subject: editableSubject,
        customer_wording: editableWording,
        internal_notes: editableNotes,
        recipient_name: editableRecipientName,
        recipient_email: editableEmail
      });
      const updated = res?.data?.data || res?.data;
      syncDraftState(updated);
      setFeedback({ type: 'success', message: 'Draft changes saved.' });
      setTimeout(() => setFeedback(null), 3000);
    } catch (err) {
      console.error('Failed to update draft:', err);
      setError(err?.response?.data?.message || err.message || 'Save draft failed');
    }
  };

  const handleSubmitForApproval = async () => {
    if (!currentDraft?.id) return;
    setSubmittingApproval(true);
    try {
      const res = await shipmentOperationsAutomationService.submitForApproval(
        shipmentId,
        currentDraft.id,
        approvalReason
      );
      const submitted = res?.data?.data || res?.data;
      syncDraftState(submitted);
      setFeedback({
        type: 'success',
        message: `Draft submitted for managerial review (Approval Request #${submitted.approval_id || 'Enqueued'}).`
      });
      setTimeout(() => setFeedback(null), 5000);
      await fetchOverview();
    } catch (err) {
      console.error('Failed to submit draft for approval:', err);
      setError(err?.response?.data?.message || err.message || 'Submission failed');
    } finally {
      setSubmittingApproval(false);
    }
  };

  if (loading) {
    return (
      <div className="sh-auto-container" data-testid="sh-auto-loading" style={{ textAlign: 'center', padding: '40px' }}>
        <RefreshCw size={24} className="sh-auto-spinner" style={{ display: 'inline-block', borderColor: '#2563eb', borderTopColor: 'transparent' }} />
        <p style={{ marginTop: '12px', color: '#64748b', fontSize: '0.9rem' }}>
          Loading shipment operations automation and deterministic risk models...
        </p>
      </div>
    );
  }

  const signals = overview?.signals || {};
  const analysis = overview?.latest_analysis;
  const prioritizedExceptions = overview?.prioritized_exceptions || [];
  const recommendations = overview?.recommendations || [];

  const getRiskBadgeClass = (level) => {
    switch (String(level).toUpperCase()) {
      case 'CRITICAL': return 'sh-auto-badge-critical';
      case 'HIGH': return 'sh-auto-badge-high';
      case 'MEDIUM': return 'sh-auto-badge-medium';
      default: return 'sh-auto-badge-low';
    }
  };

  return (
    <div className="sh-auto-container" data-testid="sh-operations-automation-section">
      {/* ── Section Header ── */}
      <div className="sh-auto-header">
        <div className="sh-auto-title-area">
          <h3>
            <ShieldAlert size={20} color="#1e3a8a" />
            Shipment Operations Automation & Exception Response
          </h3>
          <p>
            Deterministic risk detection, exception prioritization, and supervised communication draft dispatch.
          </p>
        </div>
        <div className="sh-auto-header-actions">
          <button
            type="button"
            className="sh-auto-btn sh-auto-btn-secondary"
            onClick={fetchOverview}
            disabled={analyzing}
            title="Refresh signals from backend"
          >
            <RefreshCw size={13} /> Refresh Signals
          </button>
          <button
            type="button"
            className="sh-auto-btn sh-auto-btn-primary"
            onClick={handleAnalyzeRisks}
            disabled={analyzing}
          >
            {analyzing ? <span className="sh-auto-spinner" /> : <Sparkles size={14} />}
            {analyzing ? 'Analyzing Risks...' : 'Run Risk Analysis'}
          </button>
        </div>
      </div>

      {/* Feedback & Error Notices */}
      {error && (
        <div className="sh-auto-notice sh-auto-notice-warning">
          <AlertTriangle size={16} />
          <span>{error}</span>
        </div>
      )}
      {feedback && (
        <div className={`sh-auto-notice sh-auto-notice-${feedback.type}`}>
          <CheckCircle2 size={16} />
          <span>{feedback.message}</span>
        </div>
      )}

      {/* ── 1. Deterministic Operational Signals KPI Grid ── */}
      <div className="sh-auto-kpi-grid">
        {/* Risk Level & Score */}
        <div className="sh-auto-kpi-card">
          <div className="sh-auto-kpi-label">
            <span>OPERATIONAL RISK</span>
            <span className={`sh-auto-badge ${getRiskBadgeClass(signals.risk_level || 'LOW')}`}>
              {signals.risk_level || 'LOW'}
            </span>
          </div>
          <div className="sh-auto-kpi-val">
            {signals.risk_score != null ? `${signals.risk_score} / 100` : '0 / 100'}
          </div>
          <div className="sh-auto-kpi-sub">
            {signals.requires_approval ? 'Consequential approval gate active' : 'Standard operations flow'}
          </div>
        </div>

        {/* Milestone Health */}
        <div className="sh-auto-kpi-card">
          <div className="sh-auto-kpi-label">
            <span>MILESTONE HEALTH</span>
            <Clock size={14} color="#64748b" />
          </div>
          <div className="sh-auto-kpi-val" style={{ color: signals.has_overdue_milestone ? '#b91c1c' : '#0f172a' }}>
            {signals.has_overdue_milestone ? `${signals.overdue_milestones?.length || 1} Overdue` : 'On Schedule'}
          </div>
          <div className="sh-auto-kpi-sub">
            {signals.approaching_milestones?.length > 0
              ? `${signals.approaching_milestones.length} milestone(s) approaching`
              : 'Next milestones within normal window'}
          </div>
        </div>

        {/* Tracking Freshness */}
        <div className="sh-auto-kpi-card">
          <div className="sh-auto-kpi-label">
            <span>TRACKING FRESHNESS</span>
            <Layers size={14} color="#64748b" />
          </div>
          <div className="sh-auto-kpi-val" style={{ color: signals.is_tracking_stale ? '#b45309' : '#0f172a' }}>
            {signals.hours_since_last_tracking != null
              ? `${signals.hours_since_last_tracking}h ago`
              : 'Current'}
          </div>
          <div className="sh-auto-kpi-sub">
            {signals.is_tracking_stale ? 'Tracking update gap > 24 hours' : 'Carrier telemetry active'}
          </div>
        </div>

        {/* Active Exceptions */}
        <div className="sh-auto-kpi-card">
          <div className="sh-auto-kpi-label">
            <span>ACTIVE EXCEPTIONS</span>
            <AlertCircle size={14} color="#64748b" />
          </div>
          <div className="sh-auto-kpi-val" style={{ color: signals.active_exceptions_count > 0 ? '#b91c1c' : '#059669' }}>
            {signals.active_exceptions_count || 0} Open
          </div>
          <div className="sh-auto-kpi-sub">
            {signals.critical_exceptions_count > 0
              ? `${signals.critical_exceptions_count} critical blocking issue(s)`
              : 'No unresolved critical blocks'}
          </div>
        </div>

        {/* Missing Documentation */}
        <div className="sh-auto-kpi-card">
          <div className="sh-auto-kpi-label">
            <span>DOCUMENTATION</span>
            <FileText size={14} color="#64748b" />
          </div>
          <div className="sh-auto-kpi-val" style={{ color: signals.missing_documents_count > 0 ? '#b45309' : '#059669' }}>
            {signals.missing_documents_count > 0 ? `${signals.missing_documents_count} Missing` : 'Complete'}
          </div>
          <div className="sh-auto-kpi-sub">
            {signals.missing_documents?.length > 0
              ? signals.missing_documents.join(', ')
              : 'Mandatory shipping files verified'}
          </div>
        </div>
      </div>

      {/* ── 2. AI Operational Risk Analysis & Fact Grounding ── */}
      {analysis && (
        <div className="sh-auto-card">
          <div className="sh-auto-card-header">
            <h4>
              <Sparkles size={16} color="#1e3a8a" />
              Grounded AI Operational Summary
            </h4>
            <span style={{ fontSize: '0.75rem', color: '#64748b' }}>
              Confidence: {Math.round((analysis.confidence_score || 0.85) * 100)}% · Correlation: {analysis.correlation_id || 'N/A'}
            </span>
          </div>
          <div className="sh-auto-card-body">
            <div className="sh-auto-summary-text">
              {analysis.operational_summary || 'No operational summary generated yet.'}
            </div>

            {analysis.key_risks?.length > 0 && (
              <div>
                <div style={{ fontSize: '0.8rem', fontWeight: '700', color: '#475569', marginBottom: '6px', textTransform: 'uppercase' }}>
                  Identified Key Risks
                </div>
                <div className="sh-auto-tag-list">
                  {analysis.key_risks.map((risk, idx) => (
                    <span key={idx} className="sh-auto-tag">
                      ⚠️ {risk}
                    </span>
                  ))}
                </div>
              </div>
            )}
          </div>
        </div>
      )}

      {/* ── 3. Active Exception Prioritization Matrix ── */}
      {prioritizedExceptions.length > 0 && (
        <div className="sh-auto-card">
          <div className="sh-auto-card-header">
            <h4>
              <AlertTriangle size={16} color="#d97706" />
              Prioritized Exceptions Matrix
            </h4>
            <span style={{ fontSize: '0.75rem', color: '#64748b' }}>
              Ranked by downstream operational impact & deadline
            </span>
          </div>
          <div className="sh-auto-card-body" style={{ padding: 0 }}>
            <table className="sh-auto-table">
              <thead>
                <tr>
                  <th style={{ width: '60px' }}>Rank</th>
                  <th>Urgency</th>
                  <th>Business Impact</th>
                  <th>Action Deadline</th>
                  <th>Recommended Action</th>
                </tr>
              </thead>
              <tbody>
                {prioritizedExceptions.map((ex, idx) => (
                  <tr key={idx}>
                    <td>
                      <span style={{ fontWeight: '800', color: '#1e3a8a' }}>#{ex.priority_rank || idx + 1}</span>
                    </td>
                    <td>
                      <span className={`sh-auto-badge ${getRiskBadgeClass(ex.urgency || 'MEDIUM')}`}>
                        {ex.urgency || 'MEDIUM'}
                      </span>
                    </td>
                    <td>
                      <strong>{ex.business_impact}</strong>
                      <div style={{ fontSize: '0.78rem', color: '#64748b', marginTop: '2px' }}>{ex.rationale}</div>
                    </td>
                    <td>
                      <span style={{ color: '#b91c1c', fontWeight: '600' }}>{ex.action_deadline || 'Prompt Review'}</span>
                    </td>
                    <td>
                      <span style={{ color: '#1e293b' }}>{ex.recommended_action}</span>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {/* ── 4. Operational Action Recommendations ── */}
      {recommendations.length > 0 && (
        <div className="sh-auto-card">
          <div className="sh-auto-card-header">
            <h4>
              <FileCheck size={16} color="#059669" />
              Operational Recommendations (Action System Mapped)
            </h4>
            <span style={{ fontSize: '0.75rem', color: '#64748b' }}>
              Action boundaries enforce approval for consequential updates
            </span>
          </div>
          <div className="sh-auto-card-body">
            <div className="sh-auto-rec-grid">
              {recommendations.map((rec, idx) => (
                <div key={idx} className="sh-auto-rec-card">
                  <div>
                    <div className="sh-auto-rec-header">
                      <span className="sh-auto-rec-title">{rec.title}</span>
                      <span className={`sh-auto-badge ${getRiskBadgeClass(rec.risk_rating || 'LOW')}`}>
                        {rec.risk_rating || 'LOW'}
                      </span>
                    </div>
                    <div className="sh-auto-rec-desc">{rec.description}</div>
                  </div>
                  <div className="sh-auto-rec-footer">
                    <span>Target: <code>{rec.target_action}</code></span>
                    <span style={{ color: rec.requires_approval ? '#b45309' : '#059669', fontWeight: '600' }}>
                      {rec.requires_approval ? 'Approval Required' : 'Safe Internal Action'}
                    </span>
                  </div>
                </div>
              ))}
            </div>
          </div>
        </div>
      )}

      {/* ── 5. Communication Drafts Studio ── */}
      <div className="sh-auto-card">
        <div className="sh-auto-card-header">
          <h4>
            <Mail size={16} color="#1e3a8a" />
            Communication Drafts Studio
          </h4>
          <span style={{ fontSize: '0.75rem', color: '#64748b' }}>
            Human-in-the-loop: All consequential customer & carrier messages require manager sign-off
          </span>
        </div>
        <div className="sh-auto-card-body">
          {/* Draft Type Selector & Prompt */}
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(240px, 1fr))', gap: '14px', marginBottom: '16px' }}>
            <div className="sh-auto-form-group">
              <label>Draft Category</label>
              <select
                className="sh-auto-form-select"
                value={draftType}
                onChange={(e) => setDraftType(e.target.value)}
              >
                <option value="CUSTOMER_UPDATE">Customer Proactive Delay / Status Update</option>
                <option value="CARRIER_FOLLOWUP">Carrier Schedule & Telemetry Follow-up</option>
                <option value="INTERNAL_ESCALATION">Internal Operations Escalation Notice</option>
              </select>
            </div>
            <div className="sh-auto-form-group">
              <label>Custom Drafting Instructions (Optional)</label>
              <input
                type="text"
                className="sh-auto-form-input"
                placeholder="e.g. Highlight terminal berthing delays and reassure customer regarding discharge"
                value={userInstructions}
                onChange={(e) => setUserInstructions(e.target.value)}
              />
            </div>
          </div>

          <div style={{ marginBottom: '16px' }}>
            <button
              type="button"
              className="sh-auto-btn sh-auto-btn-primary"
              onClick={handleGenerateDraft}
              disabled={generatingDraft}
            >
              {generatingDraft ? <span className="sh-auto-spinner" /> : <Sparkles size={14} />}
              {generatingDraft ? 'Synthesizing Grounded Draft...' : 'Generate Communication Draft'}
            </button>
          </div>

          {/* Current / Active Draft Editor */}
          {currentDraft ? (
            <div style={{ background: '#f8fafc', border: '1px solid #cbd5e1', borderRadius: '8px', padding: '16px' }}>
              <div className="sh-auto-draft-badge-row">
                <span className="sh-auto-badge sh-auto-badge-medium">
                  {currentDraft.draft_type}
                </span>
                <span className={`sh-auto-badge ${currentDraft.status === 'PENDING_APPROVAL' ? 'sh-auto-badge-high' : 'sh-auto-badge-low'}`}>
                  STATUS: {currentDraft.status}
                </span>
                {currentDraft.approval_id && (
                  <span style={{ fontSize: '0.8rem', color: '#1e3a8a', fontWeight: '600' }}>
                    Approval Request #{currentDraft.approval_id}
                  </span>
                )}
              </div>

              <div className="sh-auto-draft-form">
                <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '12px' }}>
                  <div className="sh-auto-form-group">
                    <label>Recipient Name</label>
                    <input
                      type="text"
                      className="sh-auto-form-input"
                      value={editableRecipientName}
                      onChange={(e) => setEditableRecipientName(e.target.value)}
                      disabled={currentDraft.status === 'PENDING_APPROVAL' || currentDraft.status === 'APPROVED'}
                    />
                  </div>
                  <div className="sh-auto-form-group">
                    <label>Recipient Email</label>
                    <input
                      type="email"
                      className="sh-auto-form-input"
                      value={editableEmail}
                      onChange={(e) => setEditableEmail(e.target.value)}
                      disabled={currentDraft.status === 'PENDING_APPROVAL' || currentDraft.status === 'APPROVED'}
                    />
                  </div>
                </div>

                <div className="sh-auto-form-group">
                  <label>Subject</label>
                  <input
                    type="text"
                    className="sh-auto-form-input"
                    value={editableSubject}
                    onChange={(e) => setEditableSubject(e.target.value)}
                    disabled={currentDraft.status === 'PENDING_APPROVAL' || currentDraft.status === 'APPROVED'}
                  />
                </div>

                <div className="sh-auto-form-group">
                  <label>Message Body (Customer/Carrier Wording)</label>
                  <textarea
                    className="sh-auto-form-textarea"
                    rows={5}
                    value={editableWording}
                    onChange={(e) => setEditableWording(e.target.value)}
                    disabled={currentDraft.status === 'PENDING_APPROVAL' || currentDraft.status === 'APPROVED'}
                  />
                </div>

                <div className="sh-auto-form-group">
                  <label>Internal Operational Notes (Audit / Team Visibility)</label>
                  <textarea
                    className="sh-auto-form-textarea"
                    rows={2}
                    value={editableNotes}
                    onChange={(e) => setEditableNotes(e.target.value)}
                    disabled={currentDraft.status === 'PENDING_APPROVAL' || currentDraft.status === 'APPROVED'}
                  />
                </div>

                {/* Submission Controls */}
                {currentDraft.status === 'DRAFT' && (
                  <div style={{ borderTop: '1px solid #e2e8f0', paddingTop: '14px', marginTop: '6px' }}>
                    <div className="sh-auto-form-group" style={{ marginBottom: '12px' }}>
                      <label>Managerial Approval Justification Reason</label>
                      <input
                        type="text"
                        className="sh-auto-form-input"
                        value={approvalReason}
                        onChange={(e) => setApprovalReason(e.target.value)}
                        placeholder="State reason for customer/carrier communication"
                      />
                    </div>
                    <div style={{ display: 'flex', gap: '10px' }}>
                      <button
                        type="button"
                        className="sh-auto-btn sh-auto-btn-secondary"
                        onClick={handleSaveDraft}
                      >
                        Save Draft Updates
                      </button>
                      <button
                        type="button"
                        className="sh-auto-btn sh-auto-btn-success"
                        onClick={handleSubmitForApproval}
                        disabled={submittingApproval}
                      >
                        {submittingApproval ? <span className="sh-auto-spinner" /> : <Send size={14} />}
                        {submittingApproval ? 'Submitting...' : 'Submit for Managerial Approval'}
                      </button>
                    </div>
                  </div>
                )}

                {currentDraft.status === 'PENDING_APPROVAL' && (
                  <div className="sh-auto-notice sh-auto-notice-warning" style={{ marginTop: '10px' }}>
                    <Clock size={16} />
                    <span>
                      Draft is locked in <strong>PENDING_APPROVAL</strong> state (Approval Request #{currentDraft.approval_id}).
                      A managerial reviewer must approve this action before transmission.
                    </span>
                  </div>
                )}
              </div>
            </div>
          ) : (
            <div style={{ textAlign: 'center', padding: '24px', color: '#64748b', fontSize: '0.85rem' }}>
              No active communication draft generated for this shipment yet. Click above to synthesize a draft.
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

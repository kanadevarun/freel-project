import React, { useState, useEffect, useCallback } from 'react';
import {
  FileText,
  ShieldCheck,
  AlertTriangle,
  Clock,
  Sparkles,
  Send,
  CheckCircle2,
  AlertCircle,
  HelpCircle,
  RefreshCw,
  Edit3,
  Check,
  ChevronRight,
  Info
} from 'lucide-react';
import toast from 'react-hot-toast';
import contractComplianceAutomationService from '../../../services/contractComplianceAutomationService';
import './ContractComplianceAutomationSection.css';

/**
 * Phase 3 Task 3.7: Contract and Compliance Automation and Intelligent Document Review
 * 100% Light LogisticsHQ Design System
 */
export const ContractComplianceAutomationSection = ({ contract, onUpdate }) => {
  const [loading, setLoading] = useState(true);
  const [actionLoading, setActionLoading] = useState(false);
  const [overview, setOverview] = useState(null);
  const [activeSubTab, setActiveSubTab] = useState('CLAUSES'); // CLAUSES | DISCREPANCIES | CHECKLIST | DRAFTS

  // Active AI review results
  const [reviewResult, setReviewResult] = useState(null);
  const [clausesResult, setClausesResult] = useState(null);
  const [discrepanciesResult, setDiscrepanciesResult] = useState(null);
  const [checklistResult, setChecklistResult] = useState(null);

  // Drafts state
  const [drafts, setDrafts] = useState([]);
  const [draftForm, setDraftForm] = useState({
    draft_type: 'MISSING_DOCUMENT_REQUEST',
    tone: 'PROFESSIONAL',
    specific_topics: '',
    custom_instructions: ''
  });
  const [editingDraftId, setEditingDraftId] = useState(null);
  const [editFields, setEditFields] = useState({ subject: '', message_body: '' });

  const fetchOverview = useCallback(async () => {
    if (!contract?.id) return;
    try {
      setLoading(true);
      const res = await contractComplianceAutomationService.getOverview(contract.id);
      const data = res?.data?.data || res?.data || res;
      setOverview(data);

      if (data?.recent_reviews?.length > 0) {
        const latest = data.recent_reviews[0];
        setReviewResult({
          risk_level: latest.risk_level,
          risk_score: latest.risk_score,
          compliance_status: latest.compliance_status,
          executive_summary: latest.executive_summary,
          extracted_clauses: latest.extracted_clauses,
          structured_discrepancies: latest.structured_discrepancies,
          compliance_obligations: latest.compliance_obligations,
          recommendations: latest.recommendations,
          confidence_score: latest.confidence_score
        });
      }
      if (data?.recent_drafts) {
        setDrafts(data.recent_drafts);
      }
    } catch (err) {
      console.error('Failed to load contract compliance overview:', err);
      toast.error('Failed to load compliance automation overview');
    } finally {
      setLoading(false);
    }
  }, [contract?.id]);

  useEffect(() => {
    fetchOverview();
  }, [fetchOverview]);

  const handleRunReview = async () => {
    try {
      setActionLoading(true);
      const res = await contractComplianceAutomationService.reviewContract(contract.id, {});
      const data = res?.data?.data || res?.data || res;
      setReviewResult(data);
      toast.success('Contract and compliance review completed');
      fetchOverview();
      if (onUpdate) onUpdate();
    } catch (err) {
      console.error(err);
      toast.error(err.response?.data?.message || 'Failed to complete contract review');
    } finally {
      setActionLoading(false);
    }
  };

  const handleExtractClauses = async () => {
    try {
      setActionLoading(true);
      const res = await contractComplianceAutomationService.extractClauses(contract.id, {});
      const data = res?.data?.data || res?.data || res;
      setClausesResult(data);
      toast.success(`Extracted ${data?.extracted_clauses?.length || 0} contract clauses`);
      setActiveSubTab('CLAUSES');
    } catch (err) {
      console.error(err);
      toast.error('Failed to extract clauses');
    } finally {
      setActionLoading(false);
    }
  };

  const handleVerifyTerms = async () => {
    try {
      setActionLoading(true);
      const res = await contractComplianceAutomationService.verifyStructuredTerms(contract.id, {});
      const data = res?.data?.data || res?.data || res;
      setDiscrepanciesResult(data);
      toast.success('Verified structured terms against document');
      setActiveSubTab('DISCREPANCIES');
    } catch (err) {
      console.error(err);
      toast.error('Failed to verify structured terms');
    } finally {
      setActionLoading(false);
    }
  };

  const handleLoadChecklist = async () => {
    try {
      setActionLoading(true);
      const res = await contractComplianceAutomationService.getComplianceChecklist(contract.id);
      const data = res?.data?.data || res?.data || res;
      setChecklistResult(data);
      setActiveSubTab('CHECKLIST');
    } catch (err) {
      console.error(err);
      toast.error('Failed to load compliance checklist');
    } finally {
      setActionLoading(false);
    }
  };

  const handleGenerateDraft = async (e) => {
    e.preventDefault();
    try {
      setActionLoading(true);
      const topics = draftForm.specific_topics
        ? draftForm.specific_topics.split(',').map((s) => s.trim()).filter(Boolean)
        : [];
      const res = await contractComplianceAutomationService.generateDraft(contract.id, {
        draft_type: draftForm.draft_type,
        tone: draftForm.tone,
        specific_clauses_or_topics: topics,
        custom_instructions: draftForm.custom_instructions
      });
      const data = res?.data?.data || res?.data || res;
      toast.success('Draft communication generated');
      setDrafts([data, ...drafts]);
      setActiveSubTab('DRAFTS');
    } catch (err) {
      console.error(err);
      toast.error(err.response?.data?.message || 'Failed to generate draft');
    } finally {
      setActionLoading(false);
    }
  };

  const handleSaveEditDraft = async (draftId) => {
    try {
      setActionLoading(true);
      const res = await contractComplianceAutomationService.updateDraft(contract.id, draftId, editFields);
      const data = res?.data?.data || res?.data || res;
      toast.success('Draft updated successfully');
      setDrafts(drafts.map((d) => (d.id === draftId ? data : d)));
      setEditingDraftId(null);
    } catch (err) {
      console.error(err);
      toast.error('Failed to update draft');
    } finally {
      setActionLoading(false);
    }
  };

  const handleSubmitForApproval = async (draftId) => {
    try {
      setActionLoading(true);
      const res = await contractComplianceAutomationService.submitApproval(contract.id, draftId, {
        notes: 'Submitted for compliance officer review & approval before communication dispatch.'
      });
      const data = res?.data?.data || res?.data || res;
      toast.success(`Draft #${draftId} submitted to Approvals Center!`);
      setDrafts(drafts.map((d) => (d.id === draftId ? data : d)));
      fetchOverview();
    } catch (err) {
      console.error(err);
      toast.error(err.response?.data?.message || 'Failed to submit approval request');
    } finally {
      setActionLoading(false);
    }
  };

  if (loading) {
    return (
      <div className="cca-container" style={{ padding: '40px 20px', textAlign: 'center' }}>
        <RefreshCw size={24} className="animate-spin text-sky" style={{ margin: '0 auto 12px auto' }} />
        <p style={{ color: '#64748b', fontSize: '0.9rem' }}>Loading Contract & Compliance Intelligence...</p>
      </div>
    );
  }

  const signals = overview?.signals || {};
  const clauses = clausesResult?.extracted_clauses || reviewResult?.extracted_clauses || [];
  const discrepancies = discrepanciesResult?.discrepancies || reviewResult?.structured_discrepancies || [];
  const checklist = checklistResult?.checklist_items || reviewResult?.compliance_obligations || [];

  return (
    <div className="cca-container" data-testid="contract-compliance-automation-section">
      {/* 1. Header Card with Deterministic Facts */}
      <div className="cca-header-card">
        <div className="cca-header-top">
          <div className="cca-title-group">
            <h3>
              <ShieldCheck size={20} className="text-sky" />
              Contract Compliance & Intelligent Document Review
            </h3>
            <p className="cca-party-subtitle">
              {contract.contract_reference} · {contract.party_name} ({contract.contract_type})
            </p>
          </div>
          <div className="cca-badges-group">
            <span className={`cca-pill ${contract.status === 'ACTIVE' ? 'active' : contract.status === 'EXPIRED' ? 'danger' : 'neutral'}`}>
              Backend: {contract.status}
            </span>
            {signals.is_expiring_soon && (
              <span className="cca-pill warning">
                <Clock size={12} /> Expiring in {signals.days_until_expiry} days
              </span>
            )}
            {signals.is_expired && (
              <span className="cca-pill danger">
                <AlertCircle size={12} /> Expired
              </span>
            )}
            {reviewResult?.risk_level && (
              <span className={`cca-pill ${reviewResult.risk_level === 'LOW' ? 'active' : reviewResult.risk_level === 'MEDIUM' ? 'warning' : 'danger'}`}>
                AI Risk: {reviewResult.risk_level}
              </span>
            )}
            <span className="cca-pill verified">
              <CheckCircle2 size={12} /> Organization Isolated
            </span>
          </div>
        </div>

        {/* Deterministic Signals Grid */}
        <div className="cca-signals-grid">
          <div className="cca-signal-metric">
            <div className="cca-signal-label">Effective Date</div>
            <div className="cca-signal-val">
              {contract.effective_date ? new Date(contract.effective_date).toLocaleDateString() : 'Missing'}
            </div>
          </div>
          <div className="cca-signal-metric">
            <div className="cca-signal-label">Expiry Date</div>
            <div className="cca-signal-val">
              {contract.expiry_date ? new Date(contract.expiry_date).toLocaleDateString() : 'Missing'}
            </div>
          </div>
          <div className="cca-signal-metric">
            <div className="cca-signal-label">Attached Document</div>
            <div className="cca-signal-val">
              {overview?.primary_document_title || (overview?.document_count > 0 ? `${overview.document_count} file(s)` : 'None Attached')}
            </div>
          </div>
          <div className="cca-signal-metric">
            <div className="cca-signal-label">Compliance Status</div>
            <div className="cca-signal-val font-semibold">
              {reviewResult?.compliance_status || (signals.compliance_risk_detected ? 'ACTION REQUIRED' : 'NORMAL')}
            </div>
          </div>
        </div>
      </div>

      {/* 2. Actions Toolbar */}
      <div className="cca-toolbar">
        <button
          className="cca-btn cca-btn-primary"
          onClick={handleRunReview}
          disabled={actionLoading}
        >
          <Sparkles size={15} />
          {actionLoading ? 'Analyzing...' : 'Run Full AI Document Review'}
        </button>
        <button
          className="cca-btn cca-btn-secondary"
          onClick={handleExtractClauses}
          disabled={actionLoading}
        >
          <FileText size={15} />
          Extract Clauses
        </button>
        <button
          className="cca-btn cca-btn-secondary"
          onClick={handleVerifyTerms}
          disabled={actionLoading}
        >
          <CheckCircle2 size={15} />
          Verify Terms vs Database
        </button>
        <button
          className="cca-btn cca-btn-secondary"
          onClick={handleLoadChecklist}
          disabled={actionLoading}
        >
          <ShieldCheck size={15} />
          Checklist
        </button>
      </div>

      {/* 3. Executive AI Summary Card */}
      {reviewResult?.executive_summary && (
        <div className="cca-summary-card">
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 8 }}>
            <span style={{ fontSize: '0.8rem', fontWeight: 700, color: '#0369a1', textTransform: 'uppercase' }}>
              AI Executive Summary
            </span>
            <span style={{ fontSize: '0.75rem', color: '#64748b' }}>
              Confidence: {Math.round((reviewResult.confidence_score || 0.92) * 100)}%
            </span>
          </div>
          <p style={{ margin: 0, fontSize: '0.88rem', lineHeight: 1.5, color: '#334155' }}>
            {reviewResult.executive_summary}
          </p>
          <div className="cca-disclaimer">
            <Info size={13} />
            <span>AI recommendation for operational support only. Not authoritative legal or regulatory advice.</span>
          </div>
        </div>
      )}

      {/* 4. Sub-Tabs */}
      <div className="cca-nav-tabs">
        <button
          className={`cca-nav-tab ${activeSubTab === 'CLAUSES' ? 'active' : ''}`}
          onClick={() => setActiveSubTab('CLAUSES')}
        >
          Extracted Clauses ({clauses.length})
        </button>
        <button
          className={`cca-nav-tab ${activeSubTab === 'DISCREPANCIES' ? 'active' : ''}`}
          onClick={() => setActiveSubTab('DISCREPANCIES')}
        >
          Database vs Document ({discrepancies.length})
        </button>
        <button
          className={`cca-nav-tab ${activeSubTab === 'CHECKLIST' ? 'active' : ''}`}
          onClick={() => setActiveSubTab('CHECKLIST')}
        >
          Mandatory Checklist ({checklist.length})
        </button>
        <button
          className={`cca-nav-tab ${activeSubTab === 'DRAFTS' ? 'active' : ''}`}
          onClick={() => setActiveSubTab('DRAFTS')}
        >
          Clarification Drafts ({drafts.length})
        </button>
      </div>

      {/* Sub-Tab 1: Extracted Clauses */}
      {activeSubTab === 'CLAUSES' && (
        <div className="cca-clauses-grid">
          {clauses.length === 0 ? (
            <div style={{ padding: 24, textAlign: 'center', background: '#fff', borderRadius: 8, border: '1px solid #e2e8f0', color: '#64748b' }}>
              No clauses extracted yet. Click "Extract Clauses" or "Run Full AI Document Review" above.
            </div>
          ) : (
            clauses.map((clause, idx) => (
              <div key={idx} className="cca-clause-card">
                <div className="cca-clause-header">
                  <span className="cca-clause-name">{clause.clause_type?.replace(/_/g, ' ')}</span>
                  <div style={{ display: 'flex', gap: 6, alignItems: 'center' }}>
                    <span className="cca-clause-source">{clause.source_document || 'Primary Agreement'}</span>
                    {clause.section_reference && (
                      <span className="cca-clause-source">{clause.section_reference}</span>
                    )}
                  </div>
                </div>
                <div className="cca-clause-text">"{clause.raw_text_snippet}"</div>
                <div className="cca-clause-footer">
                  <span>Summary: {clause.summary}</span>
                  <span>Confidence: {Math.round((clause.confidence || 0.9) * 100)}%</span>
                </div>
              </div>
            ))
          )}
        </div>
      )}

      {/* Sub-Tab 2: Discrepancies Table */}
      {activeSubTab === 'DISCREPANCIES' && (
        <div className="cca-table-card">
          <table className="cca-table">
            <thead>
              <tr>
                <th>Field Name</th>
                <th>Structured System Record</th>
                <th>Extracted Document Text</th>
                <th>Discrepancy Status</th>
                <th>AI Note</th>
              </tr>
            </thead>
            <tbody>
              {discrepancies.length === 0 ? (
                <tr>
                  <td colSpan={5} style={{ textAlign: 'center', padding: 24, color: '#64748b' }}>
                    No discrepancies detected between structured records and contract document.
                  </td>
                </tr>
              ) : (
                discrepancies.map((d, idx) => (
                  <tr key={idx}>
                    <td style={{ fontWeight: 600 }}>{d.field_name}</td>
                    <td style={{ color: '#0f172a' }}>{d.structured_value || '—'}</td>
                    <td style={{ fontStyle: 'italic' }}>{d.document_value || '—'}</td>
                    <td>
                      <span className={`cca-pill ${d.is_discrepancy ? 'warning' : 'active'}`}>
                        {d.is_discrepancy ? 'DISCREPANCY' : 'MATCH'}
                      </span>
                    </td>
                    <td style={{ fontSize: '0.8rem', color: '#64748b' }}>{d.explanation}</td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
      )}

      {/* Sub-Tab 3: Compliance Checklist */}
      {activeSubTab === 'CHECKLIST' && (
        <div className="cca-checklist-grid">
          {checklist.length === 0 ? (
            <div style={{ gridColumn: '1 / -1', padding: 24, textAlign: 'center', background: '#fff', borderRadius: 8, border: '1px solid #e2e8f0', color: '#64748b' }}>
              No compliance checklist loaded. Click "Checklist" button to evaluate.
            </div>
          ) : (
            checklist.map((item, idx) => (
              <div key={idx} className="cca-check-card">
                <div>
                  <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 6 }}>
                    <span style={{ fontWeight: 600, fontSize: '0.9rem', color: '#1e293b' }}>
                      {item.document_type?.replace(/_/g, ' ') || item.obligation_name}
                    </span>
                    <span className={`cca-pill ${item.status === 'VERIFIED' ? 'active' : item.status === 'MISSING' ? 'danger' : 'warning'}`}>
                      {item.status}
                    </span>
                  </div>
                  <p style={{ margin: 0, fontSize: '0.82rem', color: '#64748b' }}>
                    {item.description || item.governing_rule}
                  </p>
                </div>
                <div style={{ fontSize: '0.75rem', color: '#94a3b8', borderTop: '1px solid #f1f5f9', paddingTop: 8 }}>
                  Requirement: Mandatory for Carrier / Partner onboarding
                </div>
              </div>
            ))
          )}
        </div>
      )}

      {/* Sub-Tab 4: Clarification Drafts & Approvals */}
      {activeSubTab === 'DRAFTS' && (
        <div style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
          {/* Draft Generator Form */}
          <form onSubmit={handleGenerateDraft} className="cca-draft-builder">
            <h4 style={{ margin: '0 0 12px 0', fontSize: '1rem', fontWeight: 600, color: '#1e293b' }}>
              Generate Clarification or Missing Document Draft
            </h4>
            <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))', gap: 12 }}>
              <div className="cca-form-group">
                <label className="cca-form-label">Draft Type</label>
                <select
                  className="cca-select"
                  value={draftForm.draft_type}
                  onChange={(e) => setDraftForm({ ...draftForm, draft_type: e.target.value })}
                >
                  <option value="MISSING_DOCUMENT_REQUEST">Missing Document Request</option>
                  <option value="TERM_CLARIFICATION">Contract Term Clarification</option>
                  <option value="RENEWAL_FOLLOWUP">Renewal Follow-Up Inquiry</option>
                  <option value="COMPLIANCE_NOTICE">Compliance Obligation Notice</option>
                </select>
              </div>
              <div className="cca-form-group">
                <label className="cca-form-label">Communication Tone</label>
                <select
                  className="cca-select"
                  value={draftForm.tone}
                  onChange={(e) => setDraftForm({ ...draftForm, tone: e.target.value })}
                >
                  <option value="PROFESSIONAL">Professional & Collaborative</option>
                  <option value="URGENT">Urgent (Nearing Expiry / Hold)</option>
                  <option value="FIRM">Formal Legal Notice</option>
                </select>
              </div>
            </div>

            <div className="cca-form-group">
              <label className="cca-form-label">Specific Clauses or Documents (comma separated)</label>
              <input
                type="text"
                className="cca-input"
                placeholder="e.g. Certificate of Insurance, Section 8 Liability Limit"
                value={draftForm.specific_topics}
                onChange={(e) => setDraftForm({ ...draftForm, specific_topics: e.target.value })}
              />
            </div>

            <div className="cca-form-group">
              <label className="cca-form-label">Custom Instructions for AI Draft</label>
              <input
                type="text"
                className="cca-input"
                placeholder="e.g. Request $250k cargo policy endorsement and proof of carrier authority"
                value={draftForm.custom_instructions}
                onChange={(e) => setDraftForm({ ...draftForm, custom_instructions: e.target.value })}
              />
            </div>

            <button type="submit" className="cca-btn cca-btn-primary" disabled={actionLoading}>
              <Sparkles size={14} /> Generate Editable Draft
            </button>
          </form>

          {/* Drafts List */}
          <div className="cca-drafts-list">
            <h4 style={{ margin: 0, fontSize: '0.95rem', fontWeight: 600, color: '#1e293b' }}>
              Communication Drafts ({drafts.length})
            </h4>

            {drafts.length === 0 ? (
              <div style={{ padding: 20, textAlign: 'center', background: '#fff', borderRadius: 8, border: '1px solid #e2e8f0', color: '#64748b' }}>
                No drafts generated yet for this contract.
              </div>
            ) : (
              drafts.map((draft) => {
                const isEditing = editingDraftId === draft.id;
                return (
                  <div key={draft.id} className="cca-draft-item">
                    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 8, flexWrap: 'wrap', gap: 8 }}>
                      <div>
                        <span style={{ fontWeight: 600, fontSize: '0.95rem', color: '#0f172a', marginRight: 8 }}>
                          Draft #{draft.id}: {draft.draft_type?.replace(/_/g, ' ')}
                        </span>
                        <span style={{ fontSize: '0.8rem', color: '#64748b' }}>
                          To: {draft.recipient_name} &lt;{draft.recipient_email}&gt;
                        </span>
                      </div>
                      <div style={{ display: 'flex', gap: 6, alignItems: 'center' }}>
                        <span className={`cca-pill ${draft.status === 'PENDING_APPROVAL' ? 'warning' : draft.status === 'DISPATCHED' ? 'active' : 'neutral'}`}>
                          {draft.status}
                        </span>
                        {draft.approval_id && (
                          <span className="cca-pill verified">
                            Approval #{draft.approval_id}
                          </span>
                        )}
                      </div>
                    </div>

                    {isEditing ? (
                      <div style={{ display: 'flex', flexDirection: 'column', gap: 10, marginTop: 10 }}>
                        <input
                          type="text"
                          className="cca-input"
                          value={editFields.subject}
                          onChange={(e) => setEditFields({ ...editFields, subject: e.target.value })}
                        />
                        <textarea
                          className="cca-textarea"
                          value={editFields.message_body}
                          onChange={(e) => setEditFields({ ...editFields, message_body: e.target.value })}
                        />
                        <div style={{ display: 'flex', gap: 8 }}>
                          <button
                            className="cca-btn cca-btn-primary"
                            onClick={() => handleSaveEditDraft(draft.id)}
                            disabled={actionLoading}
                          >
                            <Check size={14} /> Save Changes
                          </button>
                          <button
                            className="cca-btn cca-btn-secondary"
                            onClick={() => setEditingDraftId(null)}
                          >
                            Cancel
                          </button>
                        </div>
                      </div>
                    ) : (
                      <>
                        <div style={{ fontWeight: 600, fontSize: '0.88rem', color: '#1e293b', marginBottom: 6 }}>
                          Subject: {draft.subject}
                        </div>
                        <div style={{ background: '#f8fafc', padding: 12, borderRadius: 6, fontSize: '0.85rem', color: '#334155', whiteSpace: 'pre-wrap', lineHeight: 1.5, border: '1px solid #e2e8f0' }}>
                          {draft.message_body}
                        </div>
                        <div style={{ display: 'flex', gap: 8, marginTop: 12 }}>
                          {draft.status === 'DRAFT' && (
                            <>
                              <button
                                className="cca-btn cca-btn-secondary"
                                onClick={() => {
                                  setEditingDraftId(draft.id);
                                  setEditFields({ subject: draft.subject, message_body: draft.message_body });
                                }}
                              >
                                <Edit3 size={14} /> Edit Message
                              </button>
                              <button
                                className="cca-btn cca-btn-primary"
                                onClick={() => handleSubmitForApproval(draft.id)}
                                disabled={actionLoading}
                              >
                                <Send size={14} /> Submit for HITL Approval
                              </button>
                            </>
                          )}
                          {draft.status === 'PENDING_APPROVAL' && (
                            <span style={{ fontSize: '0.82rem', color: '#b45309', display: 'flex', alignItems: 'center', gap: 5 }}>
                              <Clock size={14} /> Awaiting Manager Approval in Approvals Center
                            </span>
                          )}
                        </div>
                      </>
                    )}
                  </div>
                );
              })
            )}
          </div>
        </div>
      )}
    </div>
  );
};

export default ContractComplianceAutomationSection;

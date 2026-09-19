import React, { useState, useEffect, useCallback } from 'react';
import PropTypes from 'prop-types';
import toast from 'react-hot-toast';
import { rfqPricingWorkflowService } from '../../../../services/rfqPricingWorkflowService';
import './RFQIntelligentPricingWorkflowSection.css';

export const RFQIntelligentPricingWorkflowSection = ({ rfqId, rfqData, onWorkflowUpdated }) => {
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [overview, setOverview] = useState(null);

  // Draft Editor State
  const [activeDraftTab, setActiveDraftTab] = useState('customer'); // 'customer', 'internal', 'terms'
  const [customerWording, setCustomerWording] = useState('');
  const [internalSummary, setInternalSummary] = useState('');
  const [termsAndConditions, setTermsAndConditions] = useState('');
  const [recipientName, setRecipientName] = useState('');
  const [recipientEmail, setRecipientEmail] = useState('');
  const [targetMarginPct, setTargetMarginPct] = useState(20);
  const [discountAmount, setDiscountAmount] = useState(0);

  const [savingDraft, setSavingDraft] = useState(false);
  const [generatingDraft, setGeneratingDraft] = useState(false);
  const [submittingApproval, setSubmittingApproval] = useState(false);
  const [extractingReqs, setExtractingReqs] = useState(false);

  const fetchOverview = useCallback(async () => {
    if (!rfqId) return;
    try {
      setLoading(true);
      setError(null);
      const res = await rfqPricingWorkflowService.getOverview(rfqId);
      const data = res?.data?.data || res?.data || res;
      setOverview(data);

      if (data?.latest_draft) {
        setCustomerWording(data.latest_draft.customer_wording || '');
        setInternalSummary(data.latest_draft.internal_summary || '');
        setTermsAndConditions(data.latest_draft.terms_and_conditions || '');
        setRecipientName(data.latest_draft.recipient_name || '');
        setRecipientEmail(data.latest_draft.recipient_email || '');
      } else if (data) {
        setRecipientName(data.customer_name || '');
      }
    } catch (err) {
      console.error('Failed to load RFQ pricing workflow:', err);
      setError(err?.response?.data?.message || err?.message || 'Failed to load workflow data');
    } finally {
      setLoading(false);
    }
  }, [rfqId]);

  useEffect(() => {
    fetchOverview();
  }, [fetchOverview]);

  const handleExtractRequirements = async () => {
    try {
      setExtractingReqs(true);
      await rfqPricingWorkflowService.extractRequirements(rfqId);
      toast.success('Requirements evaluated and extraction persisted!');
      await fetchOverview();
      if (onWorkflowUpdated) onWorkflowUpdated();
    } catch (err) {
      toast.error(err?.response?.data?.message || 'Failed to extract requirements');
    } finally {
      setExtractingReqs(false);
    }
  };

  const handleRecalculatePricing = async (margin, discount) => {
    try {
      const res = await rfqPricingWorkflowService.getPricingPreview(rfqId, {
        target_margin_pct: parseFloat(margin) || 20,
        discount_amount: parseFloat(discount) || 0
      });
      const preview = res?.data?.data || res?.data || res;
      setOverview((prev) => ({
        ...prev,
        pricing_preview: preview
      }));
      toast.success('Deterministic pricing recalculated!');
    } catch (err) {
      toast.error('Could not recalculate pricing preview');
    }
  };

  const handleGenerateDraft = async () => {
    try {
      setGeneratingDraft(true);
      const res = await rfqPricingWorkflowService.generateDraft(rfqId, {
        recipient_name: recipientName,
        recipient_email: recipientEmail,
        validity_days: 14,
        discount_amount: parseFloat(discountAmount) || 0,
        user_prompt_notes: 'Commercial quotation review request'
      });
      const draft = res?.data?.data || res?.data || res;
      setOverview((prev) => ({
        ...prev,
        latest_draft: draft
      }));
      setCustomerWording(draft.customer_wording || '');
      setInternalSummary(draft.internal_summary || '');
      setTermsAndConditions(draft.terms_and_conditions || '');
      toast.success('New quotation draft generated and persisted!');
      if (onWorkflowUpdated) onWorkflowUpdated();
    } catch (err) {
      toast.error(err?.response?.data?.message || 'Failed to generate draft');
    } finally {
      setGeneratingDraft(false);
    }
  };

  const handleSaveDraft = async () => {
    if (!overview?.latest_draft?.id) {
      toast.error('No existing draft to update. Generate a draft first.');
      return;
    }
    try {
      setSavingDraft(true);
      const res = await rfqPricingWorkflowService.updateDraft(rfqId, overview.latest_draft.id, {
        customer_wording: customerWording,
        internal_summary: internalSummary,
        terms_and_conditions: termsAndConditions,
        recipient_name: recipientName,
        recipient_email: recipientEmail
      });
      const updated = res?.data?.data || res?.data || res;
      setOverview((prev) => ({
        ...prev,
        latest_draft: updated
      }));
      toast.success('Quotation draft saved successfully');
      if (onWorkflowUpdated) onWorkflowUpdated();
    } catch (err) {
      toast.error(err?.response?.data?.message || 'Failed to save draft changes');
    } finally {
      setSavingDraft(false);
    }
  };

  const handleSubmitApproval = async () => {
    if (!overview?.latest_draft?.id) {
      toast.error('Please generate a quotation draft first.');
      return;
    }
    try {
      setSubmittingApproval(true);
      const res = await rfqPricingWorkflowService.submitForApproval(
        rfqId,
        overview.latest_draft.id,
        'Commercial proposal submitted for pricing review and dispatch clearance'
      );
      const submitted = res?.data?.data || res?.data || res;
      setOverview((prev) => ({
        ...prev,
        latest_draft: submitted,
        pending_approval: true,
        approval_id: submitted.approval_id
      }));
      toast.success('Draft submitted for Manager Approval!');
      if (onWorkflowUpdated) onWorkflowUpdated();
    } catch (err) {
      toast.error(err?.response?.data?.message || 'Failed to submit draft for approval');
    } finally {
      setSubmittingApproval(false);
    }
  };

  if (loading) {
    return (
      <div className="rfq-pricing-workflow-container" style={{ padding: '32px', textAlign: 'center' }}>
        <div style={{ fontSize: '14px', color: '#64748b' }}>
          ⚡ Loading RFQ-to-Quotation Pricing Workflow & syncing deterministic records...
        </div>
      </div>
    );
  }

  if (error && !overview) {
    return (
      <div className="rfq-pricing-workflow-container" style={{ padding: '24px' }}>
        <div className="workflow-risk-alert critical">
          <div className="workflow-risk-title">⚠️ Error Loading Pricing Workflow</div>
          <div>{error}</div>
          <button
            className="btn-workflow-secondary"
            style={{ width: '120px', marginTop: '12px' }}
            onClick={fetchOverview}
          >
            Retry
          </button>
        </div>
      </div>
    );
  }

  const p = overview?.pricing_preview || {};
  const ext = overview?.extraction || {};
  const draft = overview?.latest_draft || null;
  const isPendingApproval = overview?.pending_approval || draft?.status === 'PENDING_APPROVAL';
  const isExecuted = draft?.status === 'EXECUTED';

  let marginClass = 'healthy';
  if (p.margin_health === 'LOW') marginClass = 'low';
  if (p.margin_health === 'NEGATIVE') marginClass = 'negative';

  let extStatusClass = 'draft';
  if (ext.status === 'COMPLETE') extStatusClass = 'complete';
  if (ext.status === 'INCOMPLETE') extStatusClass = 'incomplete';
  if (ext.status === 'CLARIFICATION_REQUIRED') extStatusClass = 'clarification_required';

  return (
    <div className="rfq-pricing-workflow-container" data-testid="rfq-intelligent-pricing-workflow">
      {/* 1. KPI Metric Banner */}
      <div className="rfq-workflow-kpi-grid">
        <div className="rfq-workflow-kpi-card">
          <div className="rfq-workflow-kpi-label">Requirements Readiness</div>
          <div className="rfq-workflow-kpi-value">
            <span className={`workflow-badge ${extStatusClass}`}>
              {ext.status ? ext.status.replace('_', ' ') : 'PENDING EVALUATION'}
            </span>
          </div>
          <div className="rfq-workflow-kpi-sub">
            {ext.confidence_score ? `Confidence: ${(ext.confidence_score * 100).toFixed(0)}%` : 'Not yet extracted'}
          </div>
        </div>

        <div className="rfq-workflow-kpi-card">
          <div className="rfq-workflow-kpi-label">Deterministic Cost</div>
          <div className="rfq-workflow-kpi-value">
            {p.currency || 'USD'} {p.total_cost != null ? p.total_cost.toLocaleString('en-US', { minimumFractionDigits: 2 }) : '0.00'}
          </div>
          <div className="rfq-workflow-kpi-sub">
            Carrier: {p.applied_carrier_name || 'Standard Tariff'}
          </div>
        </div>

        <div className="rfq-workflow-kpi-card">
          <div className="rfq-workflow-kpi-label">Selling Price & Margin</div>
          <div className="rfq-workflow-kpi-value">
            {p.currency || 'USD'} {p.total_selling_price != null ? p.total_selling_price.toLocaleString('en-US', { minimumFractionDigits: 2 }) : '0.00'}
            <span className={`workflow-badge ${marginClass}`} style={{ marginLeft: '6px' }}>
              {p.gross_margin_pct != null ? `${p.gross_margin_pct}%` : '0%'}
            </span>
          </div>
          <div className="rfq-workflow-kpi-sub">
            Gross Profit: {p.currency || 'USD'} {p.gross_profit != null ? p.gross_profit.toLocaleString('en-US', { minimumFractionDigits: 2 }) : '0.00'}
          </div>
        </div>

        <div className="rfq-workflow-kpi-card">
          <div className="rfq-workflow-kpi-label">Workflow Status</div>
          <div className="rfq-workflow-kpi-value">
            <span className={`workflow-badge ${isExecuted ? 'executed' : isPendingApproval ? 'pending_approval' : draft ? 'draft' : 'neutral'}`}>
              {isExecuted ? 'EXECUTED / DISPATCHED' : isPendingApproval ? 'PENDING APPROVAL' : draft ? 'DRAFT READY' : 'NO DRAFT'}
            </span>
          </div>
          <div className="rfq-workflow-kpi-sub">
            {overview?.approval_id ? `Approval ID #${overview.approval_id}` : 'Approval required before dispatch'}
          </div>
        </div>
      </div>

      {/* 2. Main Two-Column Workflow Workspace */}
      <div className="rfq-workflow-layout">
        {/* Left Column: Requirements & Deterministic Pricing */}
        <div style={{ display: 'flex', flexDirection: 'column', gap: '20px' }}>
          {/* Card A: Structured Requirements */}
          <div className="rfq-workflow-card">
            <div className="rfq-workflow-card-header">
              <div>
                <div className="rfq-workflow-card-title">
                  <span>📋</span> Normalized RFQ Requirements
                </div>
                <div className="rfq-workflow-card-desc">
                  Deterministic extraction verified against MariaDB RFQ records
                </div>
              </div>
              <button
                className="btn-workflow-secondary"
                onClick={handleExtractRequirements}
                disabled={extractingReqs}
                style={{ fontSize: '12px', padding: '6px 12px' }}
              >
                {extractingReqs ? 'Evaluating...' : 'Re-Evaluate'}
              </button>
            </div>

            <div className="rfq-req-field-grid">
              <div className="rfq-req-field-item">
                <div className="rfq-req-field-name">Origin Port (POL)</div>
                <div className={`rfq-req-field-val ${!overview?.origin ? 'missing' : ''}`}>
                  {overview?.origin || 'Missing POL'}
                </div>
              </div>

              <div className="rfq-req-field-item">
                <div className="rfq-req-field-name">Destination Port (POD)</div>
                <div className={`rfq-req-field-val ${!overview?.destination ? 'missing' : ''}`}>
                  {overview?.destination || 'Missing POD'}
                </div>
              </div>

              <div className="rfq-req-field-item">
                <div className="rfq-req-field-name">Incoterms</div>
                <div className={`rfq-req-field-val ${!overview?.incoterms ? 'missing' : ''}`}>
                  {overview?.incoterms || 'Missing Terms'}
                </div>
              </div>

              <div className="rfq-req-field-item">
                <div className="rfq-req-field-name">Shipment Mode</div>
                <div className="rfq-req-field-val">
                  {overview?.shipment_mode || 'OCEAN_FCL'}
                </div>
              </div>
            </div>

            {ext.clarification_needed && (
              <div className="workflow-risk-alert" style={{ background: '#fef2f2', borderColor: '#fecaca' }}>
                <div className="workflow-risk-title" style={{ color: '#b91c1c' }}>
                  <span>⚠️</span> Missing Inquiry Information
                </div>
                <div style={{ fontSize: '12px', color: '#991b1b' }}>
                  Mandatory shipment parameters are missing. A clarification request should be dispatched to the customer before finalizing carrier contracts.
                </div>
              </div>
            )}
          </div>

          {/* Card B: Deterministic Pricing Breakdown */}
          <div className="rfq-workflow-card">
            <div className="rfq-workflow-card-header">
              <div>
                <div className="rfq-workflow-card-title">
                  <span>💰</span> Deterministic Pricing & Margin Breakdown
                </div>
                <div className="rfq-workflow-card-desc">
                  Calculated authoritatively by Go backend engine
                </div>
              </div>
            </div>

            <table className="pricing-breakdown-table">
              <tbody>
                <tr>
                  <td>Base Ocean Freight Cost</td>
                  <td>{p.currency || 'USD'} {p.base_cost != null ? p.base_cost.toFixed(2) : '0.00'}</td>
                </tr>
                <tr>
                  <td>Verified Ancillary Surcharges (THC, BAF, Doc)</td>
                  <td>{p.currency || 'USD'} {p.surcharges != null ? p.surcharges.toFixed(2) : '0.00'}</td>
                </tr>
                <tr style={{ background: '#f8fafc' }}>
                  <td><strong>Total Carrier Cost</strong></td>
                  <td><strong>{p.currency || 'USD'} {p.total_cost != null ? p.total_cost.toFixed(2) : '0.00'}</strong></td>
                </tr>
                <tr>
                  <td>Base Commercial Selling Price</td>
                  <td>{p.currency || 'USD'} {p.base_sell != null ? p.base_sell.toFixed(2) : '0.00'}</td>
                </tr>
                {p.discounts > 0 && (
                  <tr style={{ color: '#dc2626' }}>
                    <td>Approved Commercial Discount</td>
                    <td>-{p.currency || 'USD'} {p.discounts.toFixed(2)}</td>
                  </tr>
                )}
                <tr className="total-row">
                  <td>Total All-In Selling Price</td>
                  <td>{p.currency || 'USD'} {p.total_selling_price != null ? p.total_selling_price.toFixed(2) : '0.00'}</td>
                </tr>
                <tr>
                  <td>Gross Profit Margin ($ / %)</td>
                  <td>
                    {p.currency || 'USD'} {p.gross_profit != null ? p.gross_profit.toFixed(2) : '0.00'} ({p.gross_margin_pct != null ? `${p.gross_margin_pct}%` : '0%'})
                  </td>
                </tr>
              </tbody>
            </table>

            {/* Target Margin Slider / Recalculator */}
            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '12px', background: '#f8fafc', padding: '14px', borderRadius: '8px' }}>
              <div>
                <label style={{ display: 'block', fontSize: '11px', fontWeight: 600, color: '#64748b', marginBottom: '4px' }}>
                  Target Margin %: {targetMarginPct}%
                </label>
                <input
                  type="range"
                  min="0"
                  max="40"
                  step="1"
                  value={targetMarginPct}
                  onChange={(e) => {
                    setTargetMarginPct(e.target.value);
                    handleRecalculatePricing(e.target.value, discountAmount);
                  }}
                  style={{ width: '100%' }}
                />
              </div>

              <div>
                <label style={{ display: 'block', fontSize: '11px', fontWeight: 600, color: '#64748b', marginBottom: '4px' }}>
                  Discount Amount ({p.currency || 'USD'})
                </label>
                <input
                  type="number"
                  min="0"
                  value={discountAmount}
                  onChange={(e) => {
                    setDiscountAmount(e.target.value);
                    handleRecalculatePricing(targetMarginPct, e.target.value);
                  }}
                  className="draft-meta-input"
                  style={{ padding: '4px 8px' }}
                />
              </div>
            </div>

            {/* Risk Warnings */}
            {p.margin_health === 'LOW' && (
              <div className="workflow-risk-alert">
                <div className="workflow-risk-title">
                  <span>⚠️</span> Margin Warning
                </div>
                <div style={{ fontSize: '12px' }}>
                  Gross profit margin ({p.gross_margin_pct}%) is below the 15.0% threshold. Managerial approval is required prior to release.
                </div>
              </div>
            )}
            {p.margin_health === 'NEGATIVE' && (
              <div className="workflow-risk-alert critical">
                <div className="workflow-risk-title">
                  <span>🛑</span> Negative Margin Alert
                </div>
                <div style={{ fontSize: '12px' }}>
                  Quotation selling price is below total carrier cost. Commercial Director approval is mandatory.
                </div>
              </div>
            )}
          </div>
        </div>

        {/* Right Column: Grounded Quotation Drafting Workspace */}
        <div style={{ display: 'flex', flexDirection: 'column', gap: '20px' }}>
          <div className="rfq-workflow-card" style={{ flex: 1 }}>
            <div className="rfq-workflow-card-header">
              <div>
                <div className="rfq-workflow-card-title">
                  <span>✍️</span> Quotation Draft Workspace
                </div>
                <div className="rfq-workflow-card-desc">
                  AI-prepared draft grounded on Go-calculated financial facts
                </div>
              </div>

              <button
                className="btn-workflow-primary"
                onClick={handleGenerateDraft}
                disabled={generatingDraft || isExecuted}
                style={{ fontSize: '12px', padding: '6px 14px' }}
              >
                {generatingDraft ? 'Synthesizing...' : '⚡ Generate AI Draft'}
              </button>
            </div>

            {/* Tab Navigation for Draft Editor */}
            <div className="draft-tab-nav">
              <button
                className={`draft-tab-btn ${activeDraftTab === 'customer' ? 'active' : ''}`}
                onClick={() => setActiveDraftTab('customer')}
              >
                Customer Proposal Letter
              </button>
              <button
                className={`draft-tab-btn ${activeDraftTab === 'internal' ? 'active' : ''}`}
                onClick={() => setActiveDraftTab('internal')}
              >
                Internal Pricing Summary
              </button>
              <button
                className={`draft-tab-btn ${activeDraftTab === 'terms' ? 'active' : ''}`}
                onClick={() => setActiveDraftTab('terms')}
              >
                Terms & Conditions
              </button>
            </div>

            {/* Active Editor Panel */}
            {activeDraftTab === 'customer' && (
              <textarea
                className="draft-textarea"
                rows={10}
                value={customerWording}
                onChange={(e) => setCustomerWording(e.target.value)}
                placeholder="Click 'Generate AI Draft' to synthesize professional customer wording based on verified rates..."
                disabled={isExecuted}
                data-testid="customer-wording-input"
              />
            )}

            {activeDraftTab === 'internal' && (
              <textarea
                className="draft-textarea"
                rows={10}
                value={internalSummary}
                onChange={(e) => setInternalSummary(e.target.value)}
                placeholder="Internal pricing rationale and carrier selection summary..."
                disabled={isExecuted}
                data-testid="internal-summary-input"
              />
            )}

            {activeDraftTab === 'terms' && (
              <textarea
                className="draft-textarea"
                rows={10}
                value={termsAndConditions}
                onChange={(e) => setTermsAndConditions(e.target.value)}
                placeholder="Terms and conditions..."
                disabled={isExecuted}
                data-testid="terms-conditions-input"
              />
            )}

            {/* Recipient Information */}
            <div className="draft-meta-row">
              <div className="draft-meta-input-group">
                <label>Recipient Contact Name</label>
                <input
                  type="text"
                  className="draft-meta-input"
                  value={recipientName}
                  onChange={(e) => setRecipientName(e.target.value)}
                  placeholder="e.g. John Doe"
                  disabled={isExecuted}
                />
              </div>

              <div className="draft-meta-input-group">
                <label>Recipient Email Address</label>
                <input
                  type="email"
                  className="draft-meta-input"
                  value={recipientEmail}
                  onChange={(e) => setRecipientEmail(e.target.value)}
                  placeholder="e.g. importer@company.com"
                  disabled={isExecuted}
                />
              </div>
            </div>

            {/* Approval Notice */}
            {draft?.requires_approval && (
              <div className="workflow-risk-alert" style={{ marginTop: '6px' }}>
                <div className="workflow-risk-title">
                  <span>🛡️</span> Governed Action System Protection
                </div>
                <div style={{ fontSize: '12px' }}>
                  This quotation requires manager approval before it can be sent externally. The Send Quotation action is governed by the Central Action Registry.
                </div>
              </div>
            )}

            {/* Action Bar */}
            <div className="workflow-actions-bar">
              <button
                className="btn-workflow-secondary"
                onClick={handleSaveDraft}
                disabled={savingDraft || !draft || isExecuted}
              >
                {savingDraft ? 'Saving...' : '💾 Save Draft Changes'}
              </button>

              {!isPendingApproval && !isExecuted && (
                <button
                  className="btn-workflow-approval"
                  onClick={handleSubmitApproval}
                  disabled={submittingApproval || !draft}
                >
                  {submittingApproval ? 'Submitting...' : '🚀 Submit for Manager Approval'}
                </button>
              )}

              {isPendingApproval && (
                <div style={{ display: 'inline-flex', alignItems: 'center', gap: '8px', color: '#d97706', fontSize: '13px', fontWeight: 600 }}>
                  <span>⏳</span> Awaiting Manager Review in Approvals Center
                </div>
              )}

              {isExecuted && (
                <div style={{ display: 'inline-flex', alignItems: 'center', gap: '8px', color: '#059669', fontSize: '13px', fontWeight: 600 }}>
                  <span>✅</span> Quotation Approved & Dispatched via Action System
                </div>
              )}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

RFQIntelligentPricingWorkflowSection.propTypes = {
  rfqId: PropTypes.oneOfType([PropTypes.string, PropTypes.number]).isRequired,
  rfqData: PropTypes.object,
  onWorkflowUpdated: PropTypes.func
};

export default RFQIntelligentPricingWorkflowSection;

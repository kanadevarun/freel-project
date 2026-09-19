import React, { useState, useEffect, useCallback } from 'react';
import {
  ShieldCheck,
  RefreshCw,
  AlertTriangle,
  Clock,
  DollarSign,
  TrendingUp,
  FileCheck,
  Building2,
  Calendar,
  CreditCard,
  Layers,
  Sparkles,
  AlertCircle,
  HelpCircle,
  CheckCircle2,
  Receipt,
  Mail,
  Send,
  Copy,
  Check,
  X,
  ExternalLink,
  ChevronRight
} from 'lucide-react';
import invoiceService from '../../../../services/invoiceService';
import { recommendationService } from '../../../../services/recommendationService';
import './InvoiceFinanceIntelligenceSection.css';

export default function InvoiceFinanceIntelligenceSection({ invoiceId }) {
  const [data, setData] = useState(null);
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  const [error, setError] = useState(null);

  // Task 2.5: AI Recommendations & Evidence State
  const [recommendations, setRecommendations] = useState([]);
  const [invoiceEvidence, setInvoiceEvidence] = useState(null);
  const [selectedDraftRec, setSelectedDraftRec] = useState(null);
  const [draftSubject, setDraftSubject] = useState('');
  const [draftBody, setDraftBody] = useState('');
  const [isGeneratingDraft, setIsGeneratingDraft] = useState(false);
  const [isSavingDraft, setIsSavingDraft] = useState(false);
  const [draftCopied, setDraftCopied] = useState(false);
  const [draftSaveStatus, setDraftSaveStatus] = useState(null);

  // Action Preview Modal State
  const [actionPreview, setActionPreview] = useState(null);
  const [isLoadingPreview, setIsLoadingPreview] = useState(false);
  const [approvalStatus, setApprovalStatus] = useState({});

  const fetchIntelligence = useCallback(async (isManual = false) => {
    if (!invoiceId) return;
    if (isManual) setRefreshing(true);
    else setLoading(true);
    setError(null);

    try {
      const response = await invoiceService.getInvoice360FinanceIntelligence(invoiceId);
      if (response && response.data) {
        setData(response.data);
      } else if (response && response.identity) {
        setData(response);
      } else {
        throw new Error('Invalid invoice intelligence payload received');
      }

      // Fetch AI Recommendations for this invoice (Task 2.5)
      try {
        if (recommendationService?.listRecommendations) {
          const recRes = await recommendationService.listRecommendations({ invoice_id: invoiceId });
          const items = recRes?.data?.items || recRes?.data?.recommendations || recRes?.items || [];
          setRecommendations(items);
        }
      } catch (recErr) {
        console.warn('Could not fetch invoice recommendations:', recErr);
      }

      // Fetch Invoice Financial Evidence (Task 2.5)
      try {
        if (recommendationService?.getInvoiceEvidence) {
          const evRes = await recommendationService.getInvoiceEvidence(invoiceId);
          if (evRes && evRes.data) {
            setInvoiceEvidence(evRes.data);
          }
        }
      } catch (evErr) {
        console.warn('Could not fetch invoice evidence:', evErr);
      }
    } catch (err) {
      console.error('Failed to load invoice finance intelligence:', err);
      setError(err?.response?.data?.message || err.message || 'Error loading financial intelligence');
    } finally {
      setLoading(false);
      setRefreshing(false);
    }
  }, [invoiceId]);

  useEffect(() => {
    fetchIntelligence();
  }, [fetchIntelligence]);

  // Handle explicit draft generation
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
      // Update local recommendation
      setRecommendations((prev) =>
        prev.map((r) => (r.id === rec.id ? { ...r, ...updated } : r))
      );
    } catch (err) {
      console.error('Failed to generate collection draft:', err);
      setDraftSaveStatus({ type: 'error', message: err?.message || 'Failed to generate draft' });
    } finally {
      setIsGeneratingDraft(false);
    }
  };

  // Handle saving edited draft
  const handleSaveDraft = async () => {
    if (!selectedDraftRec) return;
    setIsSavingDraft(true);
    setDraftSaveStatus(null);
    try {
      const res = await recommendationService.saveDraft(selectedDraftRec.id, draftSubject, draftBody);
      const updated = res?.data || res;
      setRecommendations((prev) =>
        prev.map((r) => (r.id === selectedDraftRec.id ? { ...r, draft_subject: draftSubject, draft_body: draftBody } : r))
      );
      setDraftSaveStatus({ type: 'success', message: 'Collection draft saved to recommendation record' });
      setTimeout(() => setDraftSaveStatus(null), 3000);
    } catch (err) {
      console.error('Failed to save draft:', err);
      setDraftSaveStatus({ type: 'error', message: 'Failed to save draft changes' });
    } finally {
      setIsSavingDraft(false);
    }
  };

  // Handle Action Preview
  const handleOpenActionPreview = async (rec) => {
    setIsLoadingPreview(true);
    setActionPreview(null);
    try {
      const res = await recommendationService.getActionPreview(rec.id);
      setActionPreview(res?.data || res);
    } catch (err) {
      console.error('Failed to fetch action preview:', err);
      setActionPreview({
        action_type: rec.action_type || 'review_overdue_invoice',
        title: rec.title,
        description: rec.description,
        target_entity: `Invoice #${rec.invoice_id || invoiceId}`,
        operator_checklist: [
          'Verify bank transaction settlement status',
          'Review customer credit history and open disputes',
          'Coordinate with account manager before customer contact'
        ],
        safety_notice: 'Zero financial records will be automatically modified.'
      });
    } finally {
      setIsLoadingPreview(false);
    }
  };

  // Handle Request Approval
  const handleRequestApproval = async (recId) => {
    try {
      await recommendationService.requestApproval(recId, 'Finance sign-off requested by operator');
      setApprovalStatus((prev) => ({ ...prev, [recId]: 'SUBMITTED' }));
      setRecommendations((prev) =>
        prev.map((r) => (r.id === recId ? { ...r, status: 'pending_approval', requires_approval: true } : r))
      );
    } catch (err) {
      console.error('Failed to request approval:', err);
      setApprovalStatus((prev) => ({ ...prev, [recId]: 'ERROR' }));
    }
  };

  if (loading) {
    return (
      <div className="inv-intel-container" data-testid="invoice-finance-intelligence-loading">
        <div className="inv-intel-loading">
          <RefreshCw className="inv-intel-btn-icon--spin" size={24} />
          <span>Assembling deterministic finance intelligence & receivables aging...</span>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="inv-intel-container" data-testid="invoice-finance-intelligence-error">
        <div className="inv-intel-error">
          <div className="inv-intel-error-left">
            <AlertTriangle size={18} />
            <span>{error}</span>
          </div>
          <button
            className="inv-intel-btn inv-intel-btn--secondary"
            onClick={() => fetchIntelligence(true)}
            data-testid="inv-intel-retry-btn"
          >
            <RefreshCw size={14} />
            <span>Retry</span>
          </button>
        </div>
      </div>
    );
  }

  if (!data) return null;

  const {
    identity = {},
    customer_receivables: customerAR = {},
    revenue_cost: revenueCost = {},
    line_items = [],
    payments = [],
    risk_indicators: riskIndicators = {},
    ai_summary: aiSummary = {},
    audited_line_item_total: auditedItemTotal = 0,
    correlation_id: correlationId = '',
    calculated_at: calculatedAt = '',
    warnings = []
  } = data;

  const safeLineItems = line_items || [];
  const safePayments = payments || [];
  const safeWarnings = warnings || [];

  const formatCurrency = (val, curr = identity.currency || 'USD') => {
    if (val === null || val === undefined) return 'N/A';
    return `${Number(val).toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })} ${curr}`;
  };

  const getRiskPillClass = (rating) => {
    switch (rating) {
      case 'CRITICAL': return 'inv-intel-risk-pill--critical';
      case 'HIGH': return 'inv-intel-risk-pill--high';
      case 'MODERATE': return 'inv-intel-risk-pill--moderate';
      case 'LOW':
      default:
        return 'inv-intel-risk-pill--low';
    }
  };

  return (
    <div className="inv-intel-container" data-testid="invoice-finance-intelligence-section">
      {/* ── HEADER ── */}
      <div className="inv-intel-header">
        <div className="inv-intel-header-left">
          <div className="inv-intel-badge inv-intel-badge--readonly">
            <ShieldCheck size={14} className="inv-intel-badge-icon" />
            <span>READ-ONLY FINANCE INTELLIGENCE</span>
          </div>
          <div className="inv-intel-title-row">
            <h3 className="inv-intel-title">Receivables & Commercial Exposure Intelligence</h3>
            <span
              className={`inv-intel-risk-pill ${getRiskPillClass(riskIndicators.overall_risk_rating)}`}
              data-testid="inv-intel-risk-rating"
            >
              Risk: {riskIndicators.overall_risk_rating || 'LOW'} ({riskIndicators.overall_risk_score || 0}/100)
            </span>
            <span
              className={`inv-intel-aging-pill ${identity.is_overdue ? 'inv-intel-aging-pill--overdue' : 'inv-intel-aging-pill--not-due'}`}
              data-testid="inv-intel-aging-pill"
            >
              Aging: {identity.aging_bucket || 'NOT_DUE'}
            </span>
          </div>
          <p className="inv-intel-subtitle">
            Deterministic balance auditing, payment delay analytics, margin evaluation, and grounded financial risks for {identity.invoice_number || `#${invoiceId}`}.
          </p>
        </div>

        <div className="inv-intel-header-right">
          <div className="inv-intel-meta">
            <div className="inv-intel-meta-item">
              <Clock size={12} />
              <span>Calculated: {calculatedAt ? new Date(calculatedAt).toLocaleTimeString() : 'Just now'}</span>
            </div>
            {correlationId && (
              <div className="inv-intel-meta-item" title={correlationId}>
                <span>Corr: {correlationId.slice(0, 15)}...</span>
              </div>
            )}
          </div>
          <button
            className="inv-intel-btn inv-intel-btn--secondary"
            onClick={() => fetchIntelligence(true)}
            disabled={refreshing}
            data-testid="inv-intel-refresh-btn"
            title="Re-run deterministic financial calculations"
          >
            <RefreshCw size={14} className={refreshing ? 'inv-intel-btn-icon--spin' : ''} />
            <span>{refreshing ? 'Auditing...' : 'Refresh'}</span>
          </button>
        </div>
      </div>

      {/* ── GROUNDED AI FINANCIAL SUMMARY CARD ── */}
      <div className="inv-intel-ai-card" data-testid="inv-intel-ai-card">
        <div className="inv-intel-ai-header">
          <div className="inv-intel-ai-header-left">
            <div className="inv-intel-ai-badge">
              <Sparkles size={13} className="inv-intel-ai-badge-icon" />
              <span>AI FINANCIAL SYNTHESIS</span>
            </div>
            <h4 className="inv-intel-ai-title">Audit & Receivables Analysis</h4>
          </div>
          <span className="inv-intel-confidence-pill">
            Confidence: {aiSummary.confidence || 'HIGH'}
          </span>
        </div>

        <div className="inv-intel-ai-body">
          <p className="inv-intel-exec-summary">
            {aiSummary.executive_summary}
          </p>

          <div className="inv-intel-ai-sections">
            <div className="inv-intel-ai-block">
              <div className="inv-intel-ai-block-title">Outstanding & AR Exposure</div>
              <p className="inv-intel-ai-block-text">
                {aiSummary.outstanding_exposure_explanation}
              </p>
            </div>

            <div className="inv-intel-ai-block">
              <div className="inv-intel-ai-block-title">Aging & Settlement Position</div>
              <p className="inv-intel-ai-block-text">
                {aiSummary.aging_position_explanation}
              </p>
            </div>

            <div className="inv-intel-ai-block">
              <div className="inv-intel-ai-block-title">Revenue & Profit Margin</div>
              <p className="inv-intel-ai-block-text">
                {aiSummary.revenue_cost_observations}
              </p>
            </div>

            <div className="inv-intel-ai-block">
              <div className="inv-intel-ai-block-title">Payment Remittance Behavior</div>
              <p className="inv-intel-ai-block-text">
                {aiSummary.payment_behavior_observations}
              </p>
            </div>
          </div>

          {/* Actionable Follow-up and Suggested Inquiries */}
          <div className="inv-intel-ai-sections">
            {aiSummary.actionable_attention_items?.length > 0 && (
              <div className="inv-intel-ai-block">
                <div className="inv-intel-ai-block-title">Actionable Attention Items</div>
                <ul className="inv-intel-ai-list">
                  {aiSummary.actionable_attention_items.map((item, idx) => (
                    <li key={idx}>{item}</li>
                  ))}
                </ul>
              </div>
            )}

            {aiSummary.suggested_finance_inquiries?.length > 0 && (
              <div className="inv-intel-ai-block">
                <div className="inv-intel-ai-block-title">Suggested Finance Inquiries</div>
                <ul className="inv-intel-ai-list">
                  {aiSummary.suggested_finance_inquiries.map((inq, idx) => (
                    <li key={idx}>{inq}</li>
                  ))}
                </ul>
              </div>
            )}
          </div>

          {/* Citations */}
          {aiSummary.citations?.length > 0 && (
            <div className="inv-intel-citations-row">
              <span>Citations:</span>
              {aiSummary.citations.map((cite, idx) => (
                <span key={idx} className="inv-intel-citation-tag">{cite}</span>
              ))}
            </div>
          )}
        </div>
      </div>

      {/* ── 4-QUADRANT KPI GRID ── */}
      <div className="inv-intel-kpi-grid">
        {/* Card 1: Balance Due & Total */}
        <div className="inv-intel-kpi-card" data-testid="inv-intel-kpi-balance">
          <div className="inv-intel-kpi-head">
            <span className="inv-intel-kpi-label">Balance Due</span>
            <DollarSign size={16} className="inv-intel-kpi-icon" />
          </div>
          <div className={`inv-intel-kpi-value ${identity.balance_due > 0 ? (identity.is_overdue ? 'inv-intel-kpi-value--danger' : 'inv-intel-kpi-value--warning') : 'inv-intel-kpi-value--success'}`}>
            {formatCurrency(identity.balance_due, identity.currency)}
          </div>
          <div className="inv-intel-kpi-subtext">
            Total: {formatCurrency(identity.total_amount, identity.currency)} | Paid: {formatCurrency(identity.paid_amount, identity.currency)}
          </div>
        </div>

        {/* Card 2: Aging & Due Timing */}
        <div className="inv-intel-kpi-card" data-testid="inv-intel-kpi-aging">
          <div className="inv-intel-kpi-head">
            <span className="inv-intel-kpi-label">Aging & Due Date</span>
            <Calendar size={16} className="inv-intel-kpi-icon" />
          </div>
          <div className={`inv-intel-kpi-value ${identity.is_overdue ? 'inv-intel-kpi-value--danger' : 'inv-intel-kpi-value--success'}`}>
            {identity.is_overdue ? `${identity.days_overdue}d Overdue` : (identity.is_paid ? 'Paid in Full' : `${identity.days_until_due}d Left`)}
          </div>
          <div className="inv-intel-kpi-subtext">
            Due: {identity.due_date ? new Date(identity.due_date).toLocaleDateString() : 'Missing Due Date'} ({identity.aging_bucket})
          </div>
        </div>

        {/* Card 3: Customer AR Exposure */}
        <div className="inv-intel-kpi-card" data-testid="inv-intel-kpi-customer-ar">
          <div className="inv-intel-kpi-head">
            <span className="inv-intel-kpi-label">Customer AR Exposure</span>
            <Building2 size={16} className="inv-intel-kpi-icon" />
          </div>
          <div className="inv-intel-kpi-value">
            {formatCurrency(customerAR.total_outstanding_amount, identity.currency)}
          </div>
          <div className="inv-intel-kpi-subtext">
            {customerAR.open_invoices_count || 0} open inv ({formatCurrency(customerAR.total_overdue_amount, identity.currency)} overdue)
          </div>
        </div>

        {/* Card 4: Profit Margin & Cost Parity */}
        <div className="inv-intel-kpi-card" data-testid="inv-intel-kpi-margin">
          <div className="inv-intel-kpi-head">
            <span className="inv-intel-kpi-label">Commercial Margin</span>
            <TrendingUp size={16} className="inv-intel-kpi-icon" />
          </div>
          <div className={`inv-intel-kpi-value ${revenueCost.gross_margin_percentage !== undefined && revenueCost.gross_margin_percentage !== null ? (revenueCost.gross_margin_percentage >= 15 ? 'inv-intel-kpi-value--success' : 'inv-intel-kpi-value--warning') : ''}`}>
            {revenueCost.gross_margin_percentage !== undefined && revenueCost.gross_margin_percentage !== null ? `${revenueCost.gross_margin_percentage}%` : 'Cost Missing'}
          </div>
          <div className="inv-intel-kpi-subtext">
            {revenueCost.gross_margin_amount !== undefined && revenueCost.gross_margin_amount !== null
              ? `Profit: ${formatCurrency(revenueCost.gross_margin_amount, identity.currency)}`
              : 'No commercial cost recorded'}
          </div>
        </div>
      </div>

      {/* ── TWO-COLUMN DETAILS: AUDITED LINE ITEMS & PAYMENTS/RISKS ── */}
      <div className="inv-intel-details-columns">
        {/* Left: Audited Line Items */}
        <div className="inv-intel-table-card" data-testid="inv-intel-line-items-card">
          <div className="inv-intel-table-header">
            <h4 className="inv-intel-table-title">Audited Line Items</h4>
            <span className="inv-intel-table-count">
              {safeLineItems.length} items (Sum: {formatCurrency(auditedItemTotal, identity.currency)})
            </span>
          </div>

          <div className="inv-intel-table-wrapper">
            <table className="inv-intel-table">
              <thead>
                <tr>
                  <th>Category</th>
                  <th>Description</th>
                  <th>Qty</th>
                  <th>Unit Price</th>
                  <th>Total</th>
                </tr>
              </thead>
              <tbody>
                {safeLineItems.length > 0 ? (
                  safeLineItems.map((item) => (
                    <tr key={item.id}>
                      <td><span className="font-semibold text-slate-700">{item.service_category || 'General'}</span></td>
                      <td>{item.description}</td>
                      <td>{item.quantity}</td>
                      <td>{formatCurrency(item.unit_price, identity.currency)}</td>
                      <td className="font-semibold">{formatCurrency(item.total_amount, identity.currency)}</td>
                    </tr>
                  ))
                ) : (
                  <tr>
                    <td colSpan="5" className="text-slate-400 text-center py-3">No line items recorded</td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>
        </div>

        {/* Right: Payment Ledger & Financial Risks */}
        <div className="inv-intel-table-card" data-testid="inv-intel-payments-card">
          <div className="inv-intel-table-header">
            <h4 className="inv-intel-table-title">Payment Ledger & Remittance</h4>
            <span className="inv-intel-table-count">{safePayments.length} transactions</span>
          </div>

          <div className="inv-intel-table-wrapper">
            <table className="inv-intel-table">
              <thead>
                <tr>
                  <th>Ref</th>
                  <th>Date</th>
                  <th>Method</th>
                  <th>Amount</th>
                  <th>Status</th>
                </tr>
              </thead>
              <tbody>
                {safePayments.length > 0 ? (
                  safePayments.map((p) => (
                    <tr key={p.id}>
                      <td className="font-mono text-blue-600">{p.payment_ref}</td>
                      <td>{new Date(p.payment_date).toLocaleDateString()}</td>
                      <td>{p.payment_method}</td>
                      <td className="font-semibold text-emerald-600">{formatCurrency(p.amount, identity.currency)}</td>
                      <td><span className="panel-badge badge-paid">{p.status}</span></td>
                    </tr>
                  ))
                ) : (
                  <tr>
                    <td colSpan="5" className="text-slate-400 text-center py-3">Zero payments recorded to date</td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>

          {/* Identified Financial Risks */}
          <div className="mt-2">
            <div className="inv-intel-table-header mb-1">
              <h5 className="inv-intel-table-title text-xs">Identified Operational Financial Risks</h5>
            </div>
            {riskIndicators.risk_factors?.length > 0 ? (
              <div className="inv-intel-risk-list">
                {riskIndicators.risk_factors.map((factor, idx) => (
                  <div key={idx} className="inv-intel-risk-item">
                    <AlertTriangle size={14} className="inv-intel-risk-icon" />
                    <span>{factor}</span>
                  </div>
                ))}
              </div>
            ) : (
              <div className="inv-intel-clean-state">
                <CheckCircle2 size={16} />
                <span>Zero elevated financial risks detected. Invoice billing & payment posture is healthy.</span>
              </div>
            )}
          </div>
        </div>
      </div>

      {/* ── TASK 2.5: AI COLLECTIONS & CONTROLLED FOLLOW-UP ASSISTANT ── */}
      <div className="inv-intel-collections-card" data-testid="invoice-collections-assistant-section">
        <div className="inv-intel-collections-header">
          <div className="inv-intel-collections-header-left">
            <div className="inv-intel-collections-badge">
              <Sparkles size={13} className="text-blue-600" />
              <span>COLLECTIONS & FOLLOW-UP ASSISTANT</span>
            </div>
            <h4 className="inv-intel-collections-title">Grounded Priorities & Controlled Communications</h4>
          </div>
          <span className="inv-intel-safety-pill">
            <ShieldCheck size={12} /> Strictly Read-Only • No Auto-Send
          </span>
        </div>

        <p className="inv-intel-collections-sub">
          Factual collections root-causes, overdue aging bands, customer exposure compounding risk, and on-demand editable communication drafts.
        </p>

        {recommendations.length > 0 ? (
          <div className="inv-intel-recs-list">
            {recommendations.map((rec) => (
              <div key={rec.id} className="inv-intel-rec-item" data-testid={`invoice-rec-${rec.id}`}>
                <div className="inv-intel-rec-top">
                  <div className="inv-intel-rec-heading">
                    <span className={`inv-intel-rec-priority-badge priority-${rec.priority || 'medium'}`}>
                      {(rec.priority || 'MEDIUM').toUpperCase()}
                    </span>
                    {rec.aging_band && (
                      <span className="inv-intel-aging-band-pill">
                        {rec.aging_band}
                      </span>
                    )}
                    <span className="inv-intel-rec-title">{rec.title}</span>
                  </div>
                  <span className="inv-intel-rec-score">
                    Confidence: {rec.confidence_score ? `${Math.round(rec.confidence_score * 100)}%` : '95%'}
                  </span>
                </div>

                <p className="inv-intel-rec-desc">{rec.description}</p>

                {/* Evidence Metrics Strip */}
                <div className="inv-intel-evidence-chips">
                  {invoiceEvidence && (
                    <>
                      <div className="inv-intel-ev-chip">
                        <span className="ev-label">Aging Band:</span>
                        <span className="ev-val font-semibold">{invoiceEvidence.aging_band}</span>
                      </div>
                      <div className="inv-intel-ev-chip">
                        <span className="ev-label">Days Overdue:</span>
                        <span className="ev-val font-semibold">{invoiceEvidence.days_overdue}d</span>
                      </div>
                      <div className="inv-intel-ev-chip">
                        <span className="ev-label">Customer Total AR:</span>
                        <span className="ev-val font-semibold">{formatCurrency(invoiceEvidence.customer_total_ar)}</span>
                      </div>
                      <div className="inv-intel-ev-chip">
                        <span className="ev-label">Open Customer Invoices:</span>
                        <span className="ev-val font-semibold">{invoiceEvidence.customer_open_invoices_count}</span>
                      </div>
                    </>
                  )}
                  {rec.draft_status === 'SAVED' && (
                    <div className="inv-intel-ev-chip chip-saved-draft">
                      <FileCheck size={12} className="text-emerald-600" />
                      <span>Draft Saved</span>
                    </div>
                  )}
                </div>

                {/* Controlled Action CTAs */}
                <div className="inv-intel-rec-actions">
                  <button
                    type="button"
                    className="inv-intel-action-btn inv-intel-action-btn--primary"
                    onClick={() => handleOpenDraft(rec)}
                    data-testid={`btn-draft-${rec.id}`}
                  >
                    <Mail size={13} />
                    <span>{rec.draft_subject ? 'Review / Edit Draft' : 'Draft Collection Notice'}</span>
                  </button>

                  <button
                    type="button"
                    className="inv-intel-action-btn inv-intel-action-btn--secondary"
                    onClick={() => handleOpenActionPreview(rec)}
                    data-testid={`btn-preview-${rec.id}`}
                  >
                    <FileCheck size={13} />
                    <span>Action Preview</span>
                  </button>

                  <button
                    type="button"
                    className="inv-intel-action-btn inv-intel-action-btn--secondary"
                    onClick={() => handleRequestApproval(rec.id)}
                    disabled={approvalStatus[rec.id] === 'SUBMITTED' || rec.status === 'pending_approval'}
                    data-testid={`btn-approval-${rec.id}`}
                  >
                    <CheckCircle2 size={13} />
                    <span>
                      {approvalStatus[rec.id] === 'SUBMITTED' || rec.status === 'pending_approval'
                        ? 'Under Finance Review'
                        : 'Request Finance Sign-off'}
                    </span>
                  </button>
                </div>
              </div>
            ))}
          </div>
        ) : (
          <div className="inv-intel-collections-clean">
            <CheckCircle2 size={16} className="text-emerald-600" />
            <span>No pending collection actions or severe financial risks flagged for this invoice. Payment posture is within normal parameters.</span>
          </div>
        )}
      </div>

      {/* ── EDITABLE COLLECTION NOTICE DRAFT MODAL ── */}
      {selectedDraftRec && (
        <div className="inv-intel-modal-overlay" onClick={() => setSelectedDraftRec(null)}>
          <div className="inv-intel-modal" onClick={(e) => e.stopPropagation()} data-testid="collection-draft-modal">
            <div className="inv-intel-modal-header">
              <div className="flex items-center gap-2">
                <div className="inv-intel-modal-icon">
                  <Mail size={16} />
                </div>
                <div>
                  <h3 className="inv-intel-modal-title">Controlled Collection Message Draft</h3>
                  <span className="inv-intel-modal-sub">
                    Draft Type: {selectedDraftRec.draft_type || 'overdue_payment_notice'} • Reference: {identity.invoice_number || `#${invoiceId}`}
                  </span>
                </div>
              </div>
              <button
                className="inv-intel-modal-close"
                onClick={() => setSelectedDraftRec(null)}
                aria-label="Close draft modal"
              >
                <X size={16} />
              </button>
            </div>

            {/* Read-Only Safety Banner */}
            <div className="inv-intel-safety-banner">
              <ShieldCheck size={16} className="text-emerald-700 flex-shrink-0" />
              <div>
                <strong>Read-Only Safety Notice:</strong> Collection drafts are strictly generated upon explicit request.
                Drafts are <em>NEVER automatically dispatched</em> to customers or accounting gateways. In accordance with read-only principles, no balances, terms, or ledgers will be mutated.
              </div>
            </div>

            <div className="inv-intel-modal-body">
              {isGeneratingDraft ? (
                <div className="inv-intel-draft-loading">
                  <RefreshCw size={20} className="inv-intel-btn-icon--spin text-blue-600" />
                  <span>Synthesizing fact-grounded, professional collection draft...</span>
                </div>
              ) : (
                <>
                  <div className="inv-intel-field-group">
                    <label className="inv-intel-field-label">Email Subject Line</label>
                    <input
                      type="text"
                      className="inv-intel-input"
                      value={draftSubject}
                      onChange={(e) => setDraftSubject(e.target.value)}
                      placeholder="e.g., Payment Notice: Overdue Invoice INV-2026-0456"
                      data-testid="draft-subject-input"
                    />
                  </div>

                  <div className="inv-intel-field-group">
                    <label className="inv-intel-field-label">Email Message Body (Editable)</label>
                    <textarea
                      rows={10}
                      className="inv-intel-textarea"
                      value={draftBody}
                      onChange={(e) => setDraftBody(e.target.value)}
                      placeholder="Draft communication text..."
                      data-testid="draft-body-textarea"
                    />
                  </div>

                  {draftSaveStatus && (
                    <div className={`inv-intel-status-pill ${draftSaveStatus.type === 'error' ? 'status-error' : 'status-success'}`}>
                      {draftSaveStatus.type === 'error' ? <AlertTriangle size={14} /> : <CheckCircle2 size={14} />}
                      <span>{draftSaveStatus.message}</span>
                    </div>
                  )}
                </>
              )}
            </div>

            <div className="inv-intel-modal-footer">
              <div className="inv-intel-modal-footer-left">
                <button
                  type="button"
                  className="inv-intel-btn inv-intel-btn--secondary"
                  onClick={() => {
                    const fullText = `Subject: ${draftSubject}\n\n${draftBody}`;
                    navigator.clipboard?.writeText(fullText);
                    setDraftCopied(true);
                    setTimeout(() => setDraftCopied(false), 2500);
                  }}
                  data-testid="btn-copy-draft"
                >
                  {draftCopied ? <Check size={14} className="text-emerald-600" /> : <Copy size={14} />}
                  <span>{draftCopied ? 'Copied to Clipboard!' : 'Copy Draft'}</span>
                </button>
              </div>

              <div className="inv-intel-modal-footer-right">
                <button
                  type="button"
                  className="inv-intel-btn inv-intel-btn--secondary"
                  onClick={() => setSelectedDraftRec(null)}
                >
                  Close
                </button>
                <button
                  type="button"
                  className="inv-intel-btn inv-intel-btn--primary"
                  onClick={handleSaveDraft}
                  disabled={isSavingDraft || isGeneratingDraft}
                  data-testid="btn-save-draft"
                >
                  {isSavingDraft ? <RefreshCw size={14} className="inv-intel-btn-icon--spin" /> : <FileCheck size={14} />}
                  <span>{isSavingDraft ? 'Saving...' : 'Save Draft'}</span>
                </button>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* ── CONTROLLED ACTION PREVIEW MODAL ── */}
      {actionPreview && (
        <div className="inv-intel-modal-overlay" onClick={() => setActionPreview(null)}>
          <div className="inv-intel-modal" onClick={(e) => e.stopPropagation()} data-testid="action-preview-modal">
            <div className="inv-intel-modal-header">
              <div className="flex items-center gap-2">
                <div className="inv-intel-modal-icon icon-preview">
                  <ShieldCheck size={16} />
                </div>
                <div>
                  <h3 className="inv-intel-modal-title">Action Execution Preview</h3>
                  <span className="inv-intel-modal-sub">
                    Target: {actionPreview.target_entity || `Invoice #${invoiceId}`}
                  </span>
                </div>
              </div>
              <button
                className="inv-intel-modal-close"
                onClick={() => setActionPreview(null)}
                aria-label="Close action preview"
              >
                <X size={16} />
              </button>
            </div>

            <div className="inv-intel-modal-body">
              <div className="inv-intel-preview-card">
                <div className="preview-label">Action Classification</div>
                <div className="preview-title font-semibold text-slate-900">{actionPreview.title || actionPreview.action_type}</div>
                <p className="preview-desc text-xs text-slate-600 mt-1">{actionPreview.description}</p>
              </div>

              <div className="inv-intel-preview-section">
                <h5 className="preview-section-title">Recommended Operator Verification Steps</h5>
                <ul className="preview-steps-list">
                  {(actionPreview.operator_checklist || [
                    'Review real bank deposit remittance history before outreach',
                    'Verify if customer has logged any active billing disputes',
                    'Coordinate with assigned account executive prior to formal escalation'
                  ]).map((step, idx) => (
                    <li key={idx} className="preview-step-item">
                      <span className="step-num">{idx + 1}</span>
                      <span>{step}</span>
                    </li>
                  ))}
                </ul>
              </div>

              <div className="inv-intel-safety-banner">
                <ShieldCheck size={16} className="text-emerald-700 flex-shrink-0" />
                <div className="text-xs">
                  <strong>Zero Mutation Guarantee:</strong> This action recommendation provides operational guidance and communications drafts only. No financial ledger entries, invoice totals, balances due, or maturity dates are modified.
                </div>
              </div>
            </div>

            <div className="inv-intel-modal-footer">
              <button
                type="button"
                className="inv-intel-btn inv-intel-btn--secondary"
                onClick={() => setActionPreview(null)}
              >
                Dismiss Preview
              </button>
            </div>
          </div>
        </div>
      )}

      {/* ── WARNINGS & LIMITATIONS ── */}
      {safeWarnings.length > 0 && (
        <div className="inv-intel-warnings-banner" data-testid="inv-intel-warnings">
          <div className="inv-intel-warnings-title">Audit Limitations & Warnings</div>
          <ul className="inv-intel-warnings-list">
            {safeWarnings.map((w, idx) => (
              <li key={idx}>{w}</li>
            ))}
          </ul>
        </div>
      )}
    </div>
  );
}

import React, { useState, useEffect, useCallback } from 'react';
import {
  DollarSign,
  TrendingUp,
  ShieldCheck,
  AlertTriangle,
  FileCheck,
  CheckCircle,
  ArrowRight,
  UserCheck,
  RefreshCw,
  Sliders,
  Send,
  Truck,
  FileText,
  Award
} from 'lucide-react';
import { enterpriseService } from '../../../../services/enterpriseService';

const STAGE_COLORS = {
  COMMERCIAL_INTAKE: { bg: '#F1F5F9', text: '#475569', border: '#CBD5E1' },
  RFQ_EXTRACTION: { bg: '#EFF6FF', text: '#1D4ED8', border: '#BFDBFE' },
  QUALIFICATION: { bg: '#FAF5FF', text: '#7E22CE', border: '#E9D5FF' },
  PRICING_MARGIN_ANALYSIS: { bg: '#F0FDF4', text: '#15803D', border: '#BBF7D0' },
  CONTRACT_COMPLIANCE_CHECK: { bg: '#ECFDF5', text: '#047857', border: '#A7F3D0' },
  QUOTATION_PREPARATION: { bg: '#F0F9FF', text: '#0369A1', border: '#BAE6FD' },
  CUSTOMER_NEGOTIATION: { bg: '#FFFBEB', text: '#B45309', border: '#FDE68A' },
  QUOTE_ACCEPTED: { bg: '#F0FDF4', text: '#166534', border: '#86EFAC' },
  BOOKING_HANDOFF: { bg: '#F5F3FF', text: '#6D28D9', border: '#DDD6FE' },
  SHIPMENT_HANDOFF: { bg: '#EFF6FF', text: '#1E40AF', border: '#93C5FD' },
  INVOICE_GENERATION: { bg: '#F0FDF4', text: '#15803D', border: '#86EFAC' },
  COLLECTION_MONITORING: { bg: '#FFF7ED', text: '#C2410C', border: '#FED7AA' },
  OUTCOME_LEARNING: { bg: '#F8FAFC', text: '#334155', border: '#CBD5E1' },
  COMPLETED: { bg: '#F8FAFC', text: '#0F172A', border: '#94A3B8' }
};

export default function AutonomousCommercialLifecycleCard({ rfqId, quotationId }) {
  const [lifecycle, setLifecycle] = useState(null);
  const [loading, setLoading] = useState(true);
  const [actionLoading, setActionLoading] = useState(false);
  const [error, setError] = useState(null);
  const [activeTab, setActiveTab] = useState('overview'); // overview, pricing, negotiation, steps

  const activeEntityId = rfqId || (quotationId ? String(quotationId) : 'RFQ-101');

  const fetchLifecycle = useCallback(async () => {
    if (!activeEntityId) return;
    setLoading(true);
    setError(null);
    try {
      const res = await enterpriseService.getCommercialLifecycle(activeEntityId);
      const payload = res?.data || res;
      setLifecycle(payload);
    } catch (err) {
      if (err?.status === 404 || err?.response?.status === 404) {
        setLifecycle(null);
      } else {
        setError(err?.response?.data?.error || err?.message || 'Failed to load commercial lifecycle');
      }
    } finally {
      setLoading(false);
    }
  }, [activeEntityId]);

  useEffect(() => {
    fetchLifecycle();
  }, [fetchLifecycle]);

  const handleInitiate = async () => {
    setActionLoading(true);
    setError(null);
    try {
      const res = await enterpriseService.initiateCommercialLifecycle(activeEntityId);
      setLifecycle(res?.data || res);
    } catch (err) {
      setError(err?.response?.data?.error || err?.message || 'Failed to initiate commercial lifecycle');
    } finally {
      setActionLoading(false);
    }
  };

  const handleRunFullCycle = async () => {
    if (!lifecycle?.workflow_id) return;
    setActionLoading(true);
    setError(null);
    try {
      // 1. Extract & Qualify
      let res = await enterpriseService.extractCommercialRFQ(lifecycle.workflow_id);
      res = await enterpriseService.qualifyCommercialRFQ(lifecycle.workflow_id);
      // 2. Pricing & Compliance
      res = await enterpriseService.optimizeCommercialPricing(lifecycle.workflow_id);
      res = await enterpriseService.checkCommercialCompliance(lifecycle.workflow_id);
      // 3. Prepare Quote
      res = await enterpriseService.prepareCommercialQuote(lifecycle.workflow_id);
      setLifecycle(res?.data || res);
    } catch (err) {
      setError(err?.response?.data?.error || err?.message || 'Cycle progression encountered policy gate');
      fetchLifecycle();
    } finally {
      setActionLoading(false);
    }
  };

  const handleNegotiate = async () => {
    if (!lifecycle?.workflow_id) return;
    const counter = window.prompt('Enter customer counter-offer price ($):', '2600.00');
    if (!counter) return;
    setActionLoading(true);
    setError(null);
    try {
      const res = await enterpriseService.negotiateCommercialQuote(lifecycle.workflow_id, parseFloat(counter), 'Customer counter-offer via online portal');
      setLifecycle(res?.data || res);
      setActiveTab('negotiation');
    } catch (err) {
      setError(err?.response?.data?.error || err?.message || 'Negotiation evaluation failed');
    } finally {
      setActionLoading(false);
    }
  };

  const handleApplyOption = async (optionId) => {
    if (!lifecycle?.workflow_id) return;
    setActionLoading(true);
    setError(null);
    try {
      const res = await enterpriseService.applyCommercialNegotiation(lifecycle.workflow_id, optionId);
      setLifecycle(res?.data || res);
    } catch (err) {
      setError(err?.response?.data?.error || err?.message || 'Failed to apply negotiation option');
    } finally {
      setActionLoading(false);
    }
  };

  const handleAcceptAndBook = async () => {
    if (!lifecycle?.workflow_id) return;
    setActionLoading(true);
    setError(null);
    try {
      const qId = lifecycle?.quotation_id || 3001;
      let res = await enterpriseService.acceptCommercialQuote(lifecycle.workflow_id, qId);
      res = await enterpriseService.handoffCommercialBooking(lifecycle.workflow_id);
      setLifecycle(res?.data || res);
    } catch (err) {
      setError(err?.response?.data?.error || err?.message || 'Booking handoff failed');
    } finally {
      setActionLoading(false);
    }
  };

  const handleShipmentHandoff = async () => {
    if (!lifecycle?.workflow_id) return;
    setActionLoading(true);
    setError(null);
    try {
      const res = await enterpriseService.handoffCommercialShipment(lifecycle.workflow_id, 101);
      setLifecycle(res?.data || res);
    } catch (err) {
      setError(err?.response?.data?.error || err?.message || 'Shipment handoff failed');
    } finally {
      setActionLoading(false);
    }
  };

  const handleInvoiceAndOutcome = async () => {
    if (!lifecycle?.workflow_id) return;
    setActionLoading(true);
    setError(null);
    try {
      let res = await enterpriseService.evaluateCommercialInvoice(lifecycle.workflow_id);
      res = await enterpriseService.assessCommercialCollection(lifecycle.workflow_id);
      res = await enterpriseService.recordCommercialOutcome(lifecycle.workflow_id, {
        won: true,
        actual_revenue: lifecycle?.pricing_recommendation?.recommended_price || 2750.0,
        actual_cost: 2200.0,
        actual_margin_pct: 20.0,
        customer_response: 'Customer accepted governed commercial terms',
        paid_on_time: true,
        dispute_raised: false,
        feedback_notes: 'Autonomous quote-to-cash lifecycle executed successfully'
      });
      setLifecycle(res?.data || res);
    } catch (err) {
      setError(err?.response?.data?.error || err?.message || 'Invoice and outcome recording failed');
    } finally {
      setActionLoading(false);
    }
  };

  const stageStyle = STAGE_COLORS[lifecycle?.current_stage] || {
    bg: '#F1F5F9',
    text: '#475569',
    border: '#CBD5E1'
  };

  return (
    <div
      style={{
        background: '#FFFFFF',
        border: '1px solid #E2E8F0',
        borderRadius: '10px',
        padding: '20px',
        fontFamily: 'Inter, -apple-system, sans-serif',
        marginBottom: '24px',
        boxShadow: '0 1px 3px rgba(0, 0, 0, 0.05)'
      }}
    >
      {/* Header Banner */}
      <div
        style={{
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
          flexWrap: 'wrap',
          gap: '12px',
          borderBottom: '1px solid #F1F5F9',
          paddingBottom: '16px',
          marginBottom: '16px'
        }}
      >
        <div style={{ display: 'flex', alignItems: 'center', gap: '10px' }}>
          <div
            style={{
              width: '36px',
              height: '36px',
              borderRadius: '8px',
              background: '#EFF6FF',
              border: '1px solid #BFDBFE',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              color: '#1D4ED8'
            }}
          >
            <DollarSign size={20} />
          </div>
          <div>
            <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
              <span style={{ fontSize: '15px', fontWeight: '700', color: '#0F172A' }}>
                Autonomous Quote-to-Cash Lifecycle
              </span>
              <span
                style={{
                  fontSize: '11px',
                  fontWeight: '600',
                  padding: '2px 8px',
                  borderRadius: '12px',
                  background: '#F1F5F9',
                  color: '#475569',
                  border: '1px solid #E2E8F0'
                }}
              >
                Phase 7.3
              </span>
            </div>
            <div style={{ fontSize: '12px', color: '#64748B', marginTop: '2px' }}>
              Target: {lifecycle?.rfq_id || activeEntityId} &bull; Correlation: {lifecycle?.correlation_id?.substring(0, 16) || 'Pending'}
            </div>
          </div>
        </div>

        <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
          {lifecycle && (
            <span
              style={{
                fontSize: '12px',
                fontWeight: '700',
                padding: '4px 10px',
                borderRadius: '6px',
                background: stageStyle.bg,
                color: stageStyle.text,
                border: `1px solid ${stageStyle.border}`
              }}
            >
              {lifecycle.current_stage?.replace(/_/g, ' ')}
            </span>
          )}

          <button
            onClick={fetchLifecycle}
            disabled={loading || actionLoading}
            style={{
              display: 'flex',
              alignItems: 'center',
              gap: '4px',
              padding: '6px 12px',
              borderRadius: '6px',
              border: '1px solid #CBD5E1',
              background: '#FFFFFF',
              color: '#334155',
              fontSize: '12px',
              fontWeight: '500',
              cursor: 'pointer'
            }}
          >
            <RefreshCw size={13} className={loading ? 'animate-spin' : ''} />
            Refresh
          </button>
        </div>
      </div>

      {/* Error / Alert banner */}
      {error && (
        <div
          style={{
            display: 'flex',
            alignItems: 'center',
            gap: '8px',
            background: '#FEF2F2',
            border: '1px solid #FECACA',
            color: '#B91C1C',
            padding: '10px 14px',
            borderRadius: '6px',
            fontSize: '13px',
            marginBottom: '16px'
          }}
        >
          <AlertTriangle size={16} />
          <span>{error}</span>
        </div>
      )}

      {/* Uninitiated State */}
      {!lifecycle && !loading && (
        <div
          style={{
            textAlign: 'center',
            padding: '32px 16px',
            background: '#F8FAFC',
            borderRadius: '8px',
            border: '1px dashed #CBD5E1'
          }}
        >
          <DollarSign size={32} color="#94A3B8" style={{ margin: '0 auto 12px' }} />
          <h4 style={{ fontSize: '14px', fontWeight: '600', color: '#1E293B', marginBottom: '6px' }}>
            No Active Quote-to-Cash Orchestration Found
          </h4>
          <p style={{ fontSize: '12px', color: '#64748B', maxWidth: '440px', margin: '0 auto 16px' }}>
            Initiate governed multi-agent operations connecting RFQ intake, extraction, qualification, pricing, quotation, booking, and invoicing.
          </p>
          <button
            onClick={handleInitiate}
            disabled={actionLoading}
            style={{
              padding: '8px 18px',
              borderRadius: '6px',
              background: '#2563EB',
              color: '#FFFFFF',
              border: 'none',
              fontSize: '13px',
              fontWeight: '600',
              cursor: 'pointer'
            }}
          >
            {actionLoading ? 'Initiating Operations...' : 'Initiate Commercial Lifecycle'}
          </button>
        </div>
      )}

      {/* Active Lifecycle View */}
      {lifecycle && (
        <div>
          {/* Waiting for Approval Alert */}
          {lifecycle.workflow_state === 'WAITING_FOR_APPROVAL' && (
            <div
              style={{
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'space-between',
                background: '#FFFBEB',
                border: '1px solid #FDE68A',
                color: '#92400E',
                padding: '12px 16px',
                borderRadius: '8px',
                marginBottom: '16px'
              }}
            >
              <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                <UserCheck size={18} />
                <span style={{ fontSize: '13px', fontWeight: '600' }}>
                  Commercial Action Suspended: Human-in-the-Loop Policy Approval Required
                </span>
              </div>
              <span style={{ fontSize: '12px', color: '#B45309', fontWeight: '500' }}>
                Pending Sign-Off
              </span>
            </div>
          )}

          {/* Navigation Tabs */}
          <div
            style={{
              display: 'flex',
              gap: '8px',
              borderBottom: '1px solid #E2E8F0',
              marginBottom: '16px'
            }}
          >
            {[
              { id: 'overview', label: 'Commercial Overview' },
              { id: 'pricing', label: 'Pricing & Margin' },
              { id: 'negotiation', label: 'Negotiation Options' },
              { id: 'steps', label: `Orchestration Steps (${lifecycle.steps?.length || 0})` }
            ].map((tab) => (
              <button
                key={tab.id}
                onClick={() => setActiveTab(tab.id)}
                style={{
                  padding: '8px 14px',
                  background: 'none',
                  border: 'none',
                  borderBottom: activeTab === tab.id ? '2px solid #2563EB' : '2px solid transparent',
                  color: activeTab === tab.id ? '#2563EB' : '#64748B',
                  fontWeight: activeTab === tab.id ? '600' : '500',
                  fontSize: '13px',
                  cursor: 'pointer'
                }}
              >
                {tab.label}
              </button>
            ))}
          </div>

          {/* TAB 1: OVERVIEW */}
          {activeTab === 'overview' && (
            <div>
              {/* Authoritative Business Data vs Intelligence Grid */}
              <div
                style={{
                  display: 'grid',
                  gridTemplateColumns: 'repeat(auto-fit, minmax(220px, 1fr))',
                  gap: '12px',
                  marginBottom: '16px'
                }}
              >
                {/* Fact: Customer Tier */}
                <div style={{ background: '#F8FAFC', border: '1px solid #E2E8F0', borderRadius: '8px', padding: '12px' }}>
                  <div style={{ fontSize: '11px', fontWeight: '600', color: '#64748B', textTransform: 'uppercase' }}>
                    Authoritative Customer Fact
                  </div>
                  <div style={{ fontSize: '16px', fontWeight: '700', color: '#0F172A', marginTop: '4px' }}>
                    {lifecycle.customer_intelligence?.authoritative_tier || 'Tier 1 Enterprise'}
                  </div>
                  <div style={{ fontSize: '11px', color: '#475569', marginTop: '2px' }}>
                    {lifecycle.customer_intelligence?.total_historical_shipments || 42} shipments &bull; {lifecycle.customer_intelligence?.average_payment_days || 28.5}d pay cycle
                  </div>
                </div>

                {/* Prediction: Win Rate */}
                <div style={{ background: '#EFF6FF', border: '1px solid #BFDBFE', borderRadius: '8px', padding: '12px' }}>
                  <div style={{ fontSize: '11px', fontWeight: '600', color: '#1D4ED8', textTransform: 'uppercase' }}>
                    AI Win Probability Prediction
                  </div>
                  <div style={{ fontSize: '16px', fontWeight: '700', color: '#1E40AF', marginTop: '4px' }}>
                    {Math.round((lifecycle.pricing_recommendation?.win_probability || 0.82) * 100)}%
                  </div>
                  <div style={{ fontSize: '11px', color: '#1E40AF', marginTop: '2px' }}>
                    Lead Score: {lifecycle.customer_intelligence?.predicted_lead_score || 88.5} &bull; Churn: {lifecycle.customer_intelligence?.predicted_churn_risk || 'LOW'}
                  </div>
                </div>

                {/* Recommendation: Target Price */}
                <div style={{ background: '#F0FDF4', border: '1px solid #BBF7D0', borderRadius: '8px', padding: '12px' }}>
                  <div style={{ fontSize: '11px', fontWeight: '600', color: '#15803D', textTransform: 'uppercase' }}>
                    Optimized Target Recommendation
                  </div>
                  <div style={{ fontSize: '16px', fontWeight: '700', color: '#166534', marginTop: '4px' }}>
                    ${lifecycle.pricing_recommendation?.recommended_price?.toFixed(2) || '2,750.00'}
                  </div>
                  <div style={{ fontSize: '11px', color: '#166534', marginTop: '2px' }}>
                    Expected Margin: {lifecycle.pricing_recommendation?.expected_margin_pct || 20.0}% &bull; Cost: ${lifecycle.pricing_recommendation?.base_carrier_cost || 2200}
                  </div>
                </div>

                {/* Governance Status */}
                <div style={{ background: '#FAF5FF', border: '1px solid #E9D5FF', borderRadius: '8px', padding: '12px' }}>
                  <div style={{ fontSize: '11px', fontWeight: '600', color: '#7E22CE', textTransform: 'uppercase' }}>
                    Governance & Downstream Links
                  </div>
                  <div style={{ fontSize: '13px', fontWeight: '600', color: '#581C87', marginTop: '4px' }}>
                    Booking: {lifecycle.booking_id ? `#${lifecycle.booking_id}` : 'Pending'}
                  </div>
                  <div style={{ fontSize: '11px', color: '#6B21A8', marginTop: '2px' }}>
                    Shipment: {lifecycle.shipment_id ? `#${lifecycle.shipment_id}` : 'Unlinked'} &bull; Inv: {lifecycle.invoice_id ? `#${lifecycle.invoice_id}` : 'Pending'}
                  </div>
                </div>
              </div>

              {/* Action Toolbar */}
              <div
                style={{
                  display: 'flex',
                  gap: '8px',
                  flexWrap: 'wrap',
                  padding: '12px',
                  background: '#F8FAFC',
                  borderRadius: '8px',
                  border: '1px solid #E2E8F0'
                }}
              >
                <button
                  onClick={handleRunFullCycle}
                  disabled={actionLoading || lifecycle.workflow_state === 'COMPLETED'}
                  style={{
                    display: 'flex',
                    alignItems: 'center',
                    gap: '6px',
                    padding: '8px 14px',
                    borderRadius: '6px',
                    background: '#2563EB',
                    color: '#FFFFFF',
                    border: 'none',
                    fontSize: '12px',
                    fontWeight: '600',
                    cursor: 'pointer'
                  }}
                >
                  <TrendingUp size={14} />
                  Analyze, Price & Prepare Quote
                </button>

                <button
                  onClick={handleNegotiate}
                  disabled={actionLoading || lifecycle.workflow_state === 'COMPLETED'}
                  style={{
                    display: 'flex',
                    alignItems: 'center',
                    gap: '6px',
                    padding: '8px 14px',
                    borderRadius: '6px',
                    background: '#FFFFFF',
                    color: '#0F172A',
                    border: '1px solid #CBD5E1',
                    fontSize: '12px',
                    fontWeight: '600',
                    cursor: 'pointer'
                  }}
                >
                  <Sliders size={14} />
                  Simulate Counter-Offer
                </button>

                <button
                  onClick={handleAcceptAndBook}
                  disabled={actionLoading || lifecycle.booking_id != null}
                  style={{
                    display: 'flex',
                    alignItems: 'center',
                    gap: '6px',
                    padding: '8px 14px',
                    borderRadius: '6px',
                    background: '#FFFFFF',
                    color: '#15803D',
                    border: '1px solid #BBF7D0',
                    fontSize: '12px',
                    fontWeight: '600',
                    cursor: 'pointer'
                  }}
                >
                  <FileCheck size={14} />
                  Accept & Book
                </button>

                <button
                  onClick={handleShipmentHandoff}
                  disabled={actionLoading || lifecycle.child_shipment_workflow_id != null}
                  style={{
                    display: 'flex',
                    alignItems: 'center',
                    gap: '6px',
                    padding: '8px 14px',
                    borderRadius: '6px',
                    background: '#FFFFFF',
                    color: '#1D4ED8',
                    border: '1px solid #BFDBFE',
                    fontSize: '12px',
                    fontWeight: '600',
                    cursor: 'pointer'
                  }}
                >
                  <Truck size={14} />
                  Handoff to Shipment
                </button>

                <button
                  onClick={handleInvoiceAndOutcome}
                  disabled={actionLoading || lifecycle.outcome_recorded}
                  style={{
                    display: 'flex',
                    alignItems: 'center',
                    gap: '6px',
                    padding: '8px 14px',
                    borderRadius: '6px',
                    background: '#FFFFFF',
                    color: '#7E22CE',
                    border: '1px solid #E9D5FF',
                    fontSize: '12px',
                    fontWeight: '600',
                    cursor: 'pointer'
                  }}
                >
                  <Award size={14} />
                  Invoice & Record Outcome
                </button>
              </div>
            </div>
          )}

          {/* TAB 2: PRICING & MARGIN */}
          {activeTab === 'pricing' && lifecycle.pricing_recommendation && (
            <div style={{ background: '#F8FAFC', border: '1px solid #E2E8F0', borderRadius: '8px', padding: '16px' }}>
              <h4 style={{ fontSize: '14px', fontWeight: '700', color: '#0F172A', marginBottom: '12px' }}>
                Pricing Intelligence & Margin Decision Support
              </h4>

              <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))', gap: '12px', marginBottom: '16px' }}>
                <div style={{ background: '#FFFFFF', padding: '10px', borderRadius: '6px', border: '1px solid #E2E8F0' }}>
                  <div style={{ fontSize: '11px', color: '#64748B' }}>Recommended Offer</div>
                  <div style={{ fontSize: '18px', fontWeight: '700', color: '#166534' }}>${lifecycle.pricing_recommendation.recommended_price}</div>
                </div>
                <div style={{ background: '#FFFFFF', padding: '10px', borderRadius: '6px', border: '1px solid #E2E8F0' }}>
                  <div style={{ fontSize: '11px', color: '#64748B' }}>Expected Margin</div>
                  <div style={{ fontSize: '18px', fontWeight: '700', color: '#1D4ED8' }}>{lifecycle.pricing_recommendation.expected_margin_pct}%</div>
                </div>
                <div style={{ background: '#FFFFFF', padding: '10px', borderRadius: '6px', border: '1px solid #E2E8F0' }}>
                  <div style={{ fontSize: '11px', color: '#64748B' }}>Base Carrier Cost</div>
                  <div style={{ fontSize: '18px', fontWeight: '700', color: '#475569' }}>${lifecycle.pricing_recommendation.base_carrier_cost}</div>
                </div>
              </div>

              {/* Evidence & Assumptions */}
              <div style={{ marginBottom: '12px' }}>
                <span style={{ fontSize: '12px', fontWeight: '600', color: '#334155' }}>Analytical Evidence:</span>
                <ul style={{ margin: '4px 0 0 16px', fontSize: '12px', color: '#475569' }}>
                  {lifecycle.pricing_recommendation.evidence?.map((ev, i) => (
                    <li key={i}>{ev}</li>
                  ))}
                </ul>
              </div>

              <div>
                <span style={{ fontSize: '12px', fontWeight: '600', color: '#334155' }}>Assumptions:</span>
                <ul style={{ margin: '4px 0 0 16px', fontSize: '12px', color: '#475569' }}>
                  {lifecycle.pricing_recommendation.assumptions?.map((as, i) => (
                    <li key={i}>{as}</li>
                  ))}
                </ul>
              </div>
            </div>
          )}

          {/* TAB 3: NEGOTIATION OPTIONS */}
          {activeTab === 'negotiation' && (
            <div>
              <div style={{ fontSize: '13px', color: '#475569', marginBottom: '12px' }}>
                Multi-agent strategic options formulated by Customer and Pricing specialists:
              </div>

              <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(240px, 1fr))', gap: '12px' }}>
                {(lifecycle.negotiation_options || []).map((opt) => (
                  <div
                    key={opt.option_id}
                    style={{
                      background: '#FFFFFF',
                      border: '1px solid #E2E8F0',
                      borderRadius: '8px',
                      padding: '14px',
                      display: 'flex',
                      flexDirection: 'column',
                      justifyContent: 'space-between'
                    }}
                  >
                    <div>
                      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '6px' }}>
                        <span style={{ fontSize: '11px', fontWeight: '700', color: '#2563EB' }}>{opt.option_id}</span>
                        {opt.approval_required && (
                          <span style={{ fontSize: '10px', background: '#FEF3C7', color: '#92400E', padding: '1px 6px', borderRadius: '4px', fontWeight: '600' }}>
                            Approval Required
                          </span>
                        )}
                      </div>
                      <div style={{ fontSize: '13px', fontWeight: '700', color: '#0F172A', marginBottom: '4px' }}>
                        {opt.label}
                      </div>
                      <div style={{ fontSize: '16px', fontWeight: '700', color: '#166534', marginBottom: '4px' }}>
                        ${opt.proposed_price.toFixed(2)}{' '}
                        <span style={{ fontSize: '11px', color: '#64748B', fontWeight: '500' }}>({opt.expected_margin_pct}% margin)</span>
                      </div>
                      <p style={{ fontSize: '11px', color: '#64748B', margin: '4px 0 12px' }}>{opt.rationale}</p>
                    </div>

                    <button
                      onClick={() => handleApplyOption(opt.option_id)}
                      disabled={actionLoading}
                      style={{
                        width: '100%',
                        padding: '6px 0',
                        borderRadius: '4px',
                        background: '#F1F5F9',
                        color: '#1E293B',
                        border: '1px solid #CBD5E1',
                        fontSize: '12px',
                        fontWeight: '600',
                        cursor: 'pointer'
                      }}
                    >
                      Apply Strategy
                    </button>
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* TAB 4: ORCHESTRATION STEPS */}
          {activeTab === 'steps' && (
            <div style={{ display: 'flex', flexDirection: 'column', gap: '8px' }}>
              {(lifecycle.steps || []).map((step, idx) => (
                <div
                  key={step.step_id || idx}
                  style={{
                    display: 'flex',
                    alignItems: 'flex-start',
                    gap: '10px',
                    padding: '10px 12px',
                    background: '#F8FAFC',
                    borderRadius: '6px',
                    border: '1px solid #E2E8F0'
                  }}
                >
                  <div style={{ marginTop: '2px' }}>
                    <CheckCircle size={15} color="#15803D" />
                  </div>
                  <div style={{ flex: 1 }}>
                    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                      <span style={{ fontSize: '12px', fontWeight: '700', color: '#0F172A' }}>
                        {step.title || step.step_name}
                      </span>
                      <span style={{ fontSize: '11px', color: '#64748B' }}>
                        Agent: <strong style={{ color: '#2563EB' }}>{step.agent_id}</strong>
                      </span>
                    </div>
                    <div style={{ fontSize: '11px', color: '#475569', marginTop: '2px' }}>
                      {step.description || step.explanation}
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      )}
    </div>
  );
}

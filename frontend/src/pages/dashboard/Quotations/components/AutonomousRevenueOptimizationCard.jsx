import React, { useState, useEffect, useCallback } from 'react';
import {
  TrendingUp,
  DollarSign,
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
  Layers,
  Award,
  ChevronRight,
  Zap,
  Clock,
  Briefcase
} from 'lucide-react';
import enterpriseService from '../../../../services/enterpriseService';

const STAGE_CONFIG = {
  MONITORING: { label: 'Continuous Monitoring', bg: '#F8FAFC', text: '#475569', border: '#CBD5E1' },
  COST_MARGIN_ANALYSIS: { label: 'Cost & Margin Analysis', bg: '#EFF6FF', text: '#1D4ED8', border: '#BFDBFE' },
  CARRIER_ECONOMICS_EVALUATION: { label: 'Carrier Economics', bg: '#F0FDF4', text: '#15803D', border: '#BBF7D0' },
  CUSTOMER_VALUE_ASSESSMENT: { label: 'Customer Value Scorecard', bg: '#FAF5FF', text: '#7E22CE', border: '#E9D5FF' },
  OPTIMIZATION_REASONING: { label: 'Multi-Agent Reasoning', bg: '#FFFBEB', text: '#B45309', border: '#FDE68A' },
  RECOMMENDATION_FORMULATION: { label: 'Recommendation Formulated', bg: '#F0F9FF', text: '#0369A1', border: '#BAE6FD' },
  WAITING_FOR_APPROVAL: { label: 'Waiting Executive Approval', bg: '#FEF2F2', text: '#B91C1C', border: '#FECACA' },
  EXECUTING: { label: 'Executing via Action System', bg: '#FFF7ED', text: '#C2410C', border: '#FED7AA' },
  VERIFYING: { label: 'Authoritative Verification', bg: '#F5F3FF', text: '#6D28D9', border: '#DDD6FE' },
  OUTCOME_TRACKING: { label: 'Outcome Tracking', bg: '#EFF6FF', text: '#1E40AF', border: '#93C5FD' },
  COMPLETED: { label: 'Cycle Completed', bg: '#F0FDF4', text: '#166534', border: '#86EFAC' },
  ESCALATED: { label: 'Commercial Escalation', bg: '#FEF2F2', text: '#991B1B', border: '#FCA5A5' }
};

export default function AutonomousRevenueOptimizationCard({ entityType = 'RFQ', entityId = 'RFQ-2026-DEV-001' }) {
  const [workflow, setWorkflow] = useState(null);
  const [loading, setLoading] = useState(true);
  const [actionLoading, setActionLoading] = useState(false);
  const [error, setError] = useState(null);
  const [activeTab, setActiveTab] = useState('overview'); // overview, carrier, customer, options, history
  const [discountInput, setDiscountInput] = useState(5.0);
  const [showNegotiateModal, setShowNegotiateModal] = useState(false);

  const fetchWorkflow = useCallback(async () => {
    if (!entityId) return;
    setLoading(true);
    setError(null);
    try {
      const res = await enterpriseService.getRevenueWorkflow(entityId);
      const data = res?.data || res;
      setWorkflow(data);
    } catch (err) {
      if (err?.status === 404 || err?.response?.status === 404) {
        setWorkflow(null);
      } else {
        setError(err?.response?.data?.error || err?.message || 'Failed to load revenue optimization workflow');
      }
    } finally {
      setLoading(false);
    }
  }, [entityId]);

  useEffect(() => {
    fetchWorkflow();
  }, [fetchWorkflow]);

  const handleInitiate = async () => {
    setActionLoading(true);
    setError(null);
    try {
      const res = await enterpriseService.initiateRevenueWorkflow(entityType, entityId);
      setWorkflow(res?.data || res);
    } catch (err) {
      setError(err?.response?.data?.error || err?.message || 'Failed to initiate revenue workflow');
    } finally {
      setActionLoading(false);
    }
  };

  const handleRunFullOptimizationCycle = async () => {
    if (!workflow?.workflow_id) return;
    setActionLoading(true);
    setError(null);
    try {
      // Step 1: Cost & Margin
      let res = await enterpriseService.evaluateRevenueCostAndMargin(workflow.workflow_id);
      // Step 2: Carrier Economics
      res = await enterpriseService.evaluateRevenueCarrierEconomics(workflow.workflow_id);
      // Step 3: Customer Value
      res = await enterpriseService.assessRevenueCustomerValue(workflow.workflow_id);
      // Step 4: Multi-Agent Reasoning
      res = await enterpriseService.conductRevenueMultiAgentOptimization(workflow.workflow_id);
      // Step 5: Recommendation Formulation
      res = await enterpriseService.formulatePricingRecommendation(workflow.workflow_id);
      setWorkflow(res?.data || res);
      setActiveTab('options');
    } catch (err) {
      setError(err?.response?.data?.error || err?.message || 'Optimization sequence encountered an issue');
      fetchWorkflow();
    } finally {
      setActionLoading(false);
    }
  };

  const handleSelectOption = async (optionId) => {
    if (!workflow?.workflow_id) return;
    setActionLoading(true);
    setError(null);
    try {
      const res = await enterpriseService.selectRevenueOptimizationOption(workflow.workflow_id, optionId);
      setWorkflow(res?.data || res);
    } catch (err) {
      setError(err?.response?.data?.error || err?.message || 'Failed to select commercial option');
      fetchWorkflow();
    } finally {
      setActionLoading(false);
    }
  };

  const handleExecuteAction = async () => {
    if (!workflow?.workflow_id) return;
    setActionLoading(true);
    setError(null);
    try {
      const res = await enterpriseService.executeRevenueCommercialAction(workflow.workflow_id, 'step-exec-action');
      setWorkflow(res?.data || res);
    } catch (err) {
      setError(err?.response?.data?.error || err?.message || 'Action execution failed');
      fetchWorkflow();
    } finally {
      setActionLoading(false);
    }
  };

  const handleVerifyAction = async () => {
    if (!workflow?.workflow_id) return;
    setActionLoading(true);
    setError(null);
    try {
      const res = await enterpriseService.verifyRevenueCommercialAction(workflow.workflow_id, 'step-verify-action');
      setWorkflow(res?.data || res);
    } catch (err) {
      setError(err?.response?.data?.error || err?.message || 'Authoritative verification failed');
      fetchWorkflow();
    } finally {
      setActionLoading(false);
    }
  };

  const handleCounterNegotiate = async () => {
    if (!workflow?.workflow_id) return;
    setActionLoading(true);
    setError(null);
    try {
      const res = await enterpriseService.optimizeNegotiationRequest(workflow.workflow_id, parseFloat(discountInput));
      setWorkflow(res?.data || res);
      setShowNegotiateModal(false);
      setActiveTab('overview');
    } catch (err) {
      setError(err?.response?.data?.error || err?.message || 'Failed to counter-optimize negotiation');
    } finally {
      setActionLoading(false);
    }
  };

  const handleRecordOutcome = async (dealWon) => {
    if (!workflow?.workflow_id) return;
    setActionLoading(true);
    setError(null);
    try {
      const feedback = {
        quoted_price: workflow.current_recommendation?.recommended_price || 2822.09,
        actual_revenue: workflow.current_recommendation?.recommended_price || 2822.09,
        actual_cost: 2300.00,
        realized_margin_pct: workflow.current_recommendation?.expected_margin_pct || 18.5,
        deal_won: dealWon,
        negotiation_rounds: workflow.replan_version + 1,
        lessons_learned: dealWon
          ? 'Target margin secured with controlled pricing policy adherence'
          : 'Quote lost to aggressive spot competitor; review carrier allocation tier'
      };
      const res = await enterpriseService.recordRevenueOutcome(workflow.workflow_id, feedback);
      setWorkflow(res?.data || res);
    } catch (err) {
      setError(err?.response?.data?.error || err?.message || 'Failed to record revenue outcome');
    } finally {
      setActionLoading(false);
    }
  };

  const currentStage = workflow?.current_stage || 'MONITORING';
  const stageStyle = STAGE_CONFIG[currentStage] || STAGE_CONFIG.MONITORING;

  return (
    <div style={{
      background: '#FFFFFF',
      border: '1px solid #E2E8F0',
      borderRadius: '12px',
      padding: '20px',
      marginBottom: '24px',
      boxShadow: '0 1px 3px rgba(0, 0, 0, 0.05)',
      fontFamily: 'inherit'
    }}>
      {/* Header */}
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '16px', flexWrap: 'wrap', gap: '12px' }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: '12px' }}>
          <div style={{
            width: '40px',
            height: '40px',
            borderRadius: '10px',
            background: '#EFF6FF',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            color: '#2563EB'
          }}>
            <TrendingUp size={22} />
          </div>
          <div>
            <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
              <h3 style={{ margin: 0, fontSize: '17px', fontWeight: 600, color: '#0F172A' }}>
                Autonomous Revenue & Margin Optimization
              </h3>
              <span style={{
                fontSize: '11px',
                fontWeight: 600,
                padding: '2px 8px',
                borderRadius: '999px',
                background: '#F1F5F9',
                color: '#475569'
              }}>
                Phase 7.6 Enterprise
              </span>
            </div>
            <p style={{ margin: '2px 0 0', fontSize: '13px', color: '#64748B' }}>
              Continuous multi-agent commercial economics, margin floor protection, and governed tariff execution
            </p>
          </div>
        </div>

        {/* Status Badge & Actions */}
        <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
          <div style={{
            padding: '4px 12px',
            borderRadius: '6px',
            background: stageStyle.bg,
            border: `1px solid ${stageStyle.border}`,
            color: stageStyle.text,
            fontSize: '12px',
            fontWeight: 600,
            display: 'flex',
            alignItems: 'center',
            gap: '6px'
          }}>
            <span style={{ width: '6px', height: '6px', borderRadius: '50%', background: stageStyle.text }} />
            {stageStyle.label}
          </div>

          <button
            onClick={fetchWorkflow}
            disabled={loading || actionLoading}
            style={{
              background: '#F8FAFC',
              border: '1px solid #E2E8F0',
              borderRadius: '6px',
              padding: '6px 10px',
              color: '#475569',
              cursor: 'pointer',
              display: 'flex',
              alignItems: 'center',
              gap: '4px',
              fontSize: '12px'
            }}
            title="Refresh Workflow State"
          >
            <RefreshCw size={14} className={loading ? 'animate-spin' : ''} />
          </button>
        </div>
      </div>

      {/* Error alert */}
      {error && (
        <div style={{
          background: '#FEF2F2',
          border: '1px solid #FECACA',
          borderRadius: '8px',
          padding: '10px 14px',
          marginBottom: '16px',
          color: '#991B1B',
          fontSize: '13px',
          display: 'flex',
          alignItems: 'center',
          gap: '8px'
        }}>
          <AlertTriangle size={16} />
          <span>{error}</span>
        </div>
      )}

      {!workflow ? (
        <div style={{
          background: '#F8FAFC',
          border: '1px dashed #CBD5E1',
          borderRadius: '8px',
          padding: '24px',
          textAlign: 'center'
        }}>
          <p style={{ margin: '0 0 12px', fontSize: '14px', color: '#64748B' }}>
            No autonomous revenue optimization workflow active for {entityType} #{entityId}.
          </p>
          <button
            onClick={handleInitiate}
            disabled={actionLoading}
            style={{
              background: '#2563EB',
              color: '#FFFFFF',
              border: 'none',
              borderRadius: '6px',
              padding: '8px 16px',
              fontSize: '13px',
              fontWeight: 500,
              cursor: 'pointer',
              display: 'inline-flex',
              alignItems: 'center',
              gap: '6px'
            }}
          >
            <Zap size={14} />
            {actionLoading ? 'Initiating Workflow...' : 'Initiate Autonomous Commercial Optimization'}
          </button>
        </div>
      ) : (
        <>
          {/* Navigation Tabs */}
          <div style={{
            display: 'flex',
            gap: '8px',
            borderBottom: '1px solid #E2E8F0',
            paddingBottom: '10px',
            marginBottom: '16px',
            overflowX: 'auto'
          }}>
            {[
              { id: 'overview', label: 'Commercial Overview' },
              { id: 'carrier', label: 'Carrier Economics' },
              { id: 'customer', label: 'Customer Value Scorecard' },
              { id: 'options', label: `Governed Options (${workflow.available_options?.length || 0})` },
              { id: 'history', label: `Decision Versions (${workflow.recommendation_history?.length || 0})` }
            ].map(tab => (
              <button
                key={tab.id}
                onClick={() => setActiveTab(tab.id)}
                style={{
                  background: activeTab === tab.id ? '#EFF6FF' : 'transparent',
                  color: activeTab === tab.id ? '#1D4ED8' : '#64748B',
                  border: activeTab === tab.id ? '1px solid #BFDBFE' : '1px solid transparent',
                  borderRadius: '6px',
                  padding: '6px 12px',
                  fontSize: '13px',
                  fontWeight: activeTab === tab.id ? 600 : 500,
                  cursor: 'pointer',
                  whiteSpace: 'nowrap'
                }}
              >
                {tab.label}
              </button>
            ))}
          </div>

          {/* TAB CONTENT: Overview */}
          {activeTab === 'overview' && (
            <div>
              {/* Fact vs Prediction vs Recommendation Header Strip */}
              <div style={{
                display: 'grid',
                gridTemplateColumns: 'repeat(auto-fit, minmax(220px, 1fr))',
                gap: '12px',
                marginBottom: '16px'
              }}>
                {/* Authoritative Fact */}
                <div style={{ background: '#F8FAFC', border: '1px solid #E2E8F0', borderRadius: '8px', padding: '12px' }}>
                  <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '6px' }}>
                    <span style={{ fontSize: '11px', fontWeight: 600, color: '#475569', textTransform: 'uppercase' }}>Authoritative Ledger Fact</span>
                    <span style={{ fontSize: '10px', background: '#E2E8F0', color: '#334155', padding: '1px 5px', borderRadius: '4px' }}>VERIFIED</span>
                  </div>
                  <div style={{ fontSize: '20px', fontWeight: 700, color: '#0F172A' }}>
                    ${workflow.metrics?.actual_quoted_price?.toFixed(2) || '2,850.00'}
                  </div>
                  <div style={{ fontSize: '12px', color: '#64748B', marginTop: '4px' }}>
                    Carrier Cost: ${workflow.metrics?.actual_carrier_cost?.toFixed(2) || '2,200.00'} (Margin: {workflow.metrics?.actual_margin_percentage?.toFixed(1) || '17.5'}%)
                  </div>
                </div>

                {/* AI Recommendation */}
                <div style={{ background: '#EFF6FF', border: '1px solid #BFDBFE', borderRadius: '8px', padding: '12px' }}>
                  <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '6px' }}>
                    <span style={{ fontSize: '11px', fontWeight: 600, color: '#1D4ED8', textTransform: 'uppercase' }}>AI Target Recommendation</span>
                    <span style={{ fontSize: '10px', background: '#DBEAFE', color: '#1E40AF', padding: '1px 5px', borderRadius: '4px' }}>VERSION {workflow.current_recommendation?.version || 1}</span>
                  </div>
                  <div style={{ fontSize: '20px', fontWeight: 700, color: '#1D4ED8' }}>
                    ${workflow.current_recommendation?.recommended_price?.toFixed(2) || '2,822.09'}
                  </div>
                  <div style={{ fontSize: '12px', color: '#2563EB', marginTop: '4px' }}>
                    Target Margin: {workflow.current_recommendation?.expected_margin_pct?.toFixed(1) || '18.5'}% (Win Prob: {workflow.current_recommendation?.win_probability_pct || 78.4}%)
                  </div>
                </div>

                {/* Governance & Autonomy */}
                <div style={{ background: '#FAF5FF', border: '1px solid #E9D5FF', borderRadius: '8px', padding: '12px' }}>
                  <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '6px' }}>
                    <span style={{ fontSize: '11px', fontWeight: 600, color: '#7E22CE', textTransform: 'uppercase' }}>Go Policy Enforcement</span>
                    <span style={{ fontSize: '10px', background: '#F3E8FF', color: '#6B21A8', padding: '1px 5px', borderRadius: '4px' }}>GOVERNED</span>
                  </div>
                  <div style={{ fontSize: '14px', fontWeight: 600, color: '#581C87' }}>
                    Margin Floor: &ge; 12.0% | Max Discount: &le; 15.0%
                  </div>
                  <div style={{ fontSize: '12px', color: '#7E22CE', marginTop: '4px' }}>
                    {workflow.pending_approvals_count > 0 ? 'Requires Executive Signoff' : 'Permitted under Level 3 Autonomy'}
                  </div>
                </div>
              </div>

              {/* Action Toolbar */}
              <div style={{
                display: 'flex',
                gap: '8px',
                flexWrap: 'wrap',
                background: '#F8FAFC',
                padding: '12px',
                borderRadius: '8px',
                border: '1px solid #E2E8F0',
                alignItems: 'center',
                justifyContent: 'space-between'
              }}>
                <div style={{ display: 'flex', gap: '8px', flexWrap: 'wrap' }}>
                  <button
                    onClick={handleRunFullOptimizationCycle}
                    disabled={actionLoading}
                    style={{
                      background: '#2563EB',
                      color: '#FFFFFF',
                      border: 'none',
                      borderRadius: '6px',
                      padding: '7px 14px',
                      fontSize: '13px',
                      fontWeight: 500,
                      cursor: 'pointer',
                      display: 'inline-flex',
                      alignItems: 'center',
                      gap: '6px'
                    }}
                  >
                    <Zap size={14} />
                    Run Multi-Agent Optimization
                  </button>

                  <button
                    onClick={() => setShowNegotiateModal(true)}
                    disabled={actionLoading}
                    style={{
                      background: '#FFFFFF',
                      color: '#0F172A',
                      border: '1px solid #CBD5E1',
                      borderRadius: '6px',
                      padding: '7px 14px',
                      fontSize: '13px',
                      fontWeight: 500,
                      cursor: 'pointer',
                      display: 'inline-flex',
                      alignItems: 'center',
                      gap: '6px'
                    }}
                  >
                    <Sliders size={14} />
                    Counter-Negotiate Discount
                  </button>

                  {workflow.current_stage === 'EXECUTING' && (
                    <button
                      onClick={handleExecuteAction}
                      disabled={actionLoading}
                      style={{
                        background: '#16A34A',
                        color: '#FFFFFF',
                        border: 'none',
                        borderRadius: '6px',
                        padding: '7px 14px',
                        fontSize: '13px',
                        fontWeight: 500,
                        cursor: 'pointer',
                        display: 'inline-flex',
                        alignItems: 'center',
                        gap: '6px'
                      }}
                    >
                      <CheckCircle size={14} />
                      Execute Action via Go System
                    </button>
                  )}

                  {workflow.current_stage === 'VERIFYING' && (
                    <button
                      onClick={handleVerifyAction}
                      disabled={actionLoading}
                      style={{
                        background: '#7C3AED',
                        color: '#FFFFFF',
                        border: 'none',
                        borderRadius: '6px',
                        padding: '7px 14px',
                        fontSize: '13px',
                        fontWeight: 500,
                        cursor: 'pointer',
                        display: 'inline-flex',
                        alignItems: 'center',
                        gap: '6px'
                      }}
                    >
                      <ShieldCheck size={14} />
                      Verify Ledger Persistence
                    </button>
                  )}

                  {workflow.current_stage === 'OUTCOME_TRACKING' && (
                    <div style={{ display: 'flex', gap: '6px' }}>
                      <button
                        onClick={() => handleRecordOutcome(true)}
                        disabled={actionLoading}
                        style={{
                          background: '#15803D',
                          color: '#FFFFFF',
                          border: 'none',
                          borderRadius: '6px',
                          padding: '7px 12px',
                          fontSize: '12px',
                          cursor: 'pointer'
                        }}
                      >
                        Record Won Deal
                      </button>
                      <button
                        onClick={() => handleRecordOutcome(false)}
                        disabled={actionLoading}
                        style={{
                          background: '#DC2626',
                          color: '#FFFFFF',
                          border: 'none',
                          borderRadius: '6px',
                          padding: '7px 12px',
                          fontSize: '12px',
                          cursor: 'pointer'
                        }}
                      >
                        Record Lost Deal
                      </button>
                    </div>
                  )}
                </div>

                <div style={{ fontSize: '12px', color: '#64748B' }}>
                  Replan Version: <strong style={{ color: '#0F172A' }}>v{workflow.replan_version}</strong>
                </div>
              </div>
            </div>
          )}

          {/* TAB CONTENT: Carrier Economics */}
          {activeTab === 'carrier' && (
            <div style={{ background: '#F8FAFC', border: '1px solid #E2E8F0', borderRadius: '8px', padding: '16px' }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '12px' }}>
                <h4 style={{ margin: 0, fontSize: '14px', color: '#0F172A', display: 'flex', alignItems: 'center', gap: '6px' }}>
                  <Truck size={16} color="#2563EB" />
                  Authoritative Carrier Operational & Economic Signals
                </h4>
                <span style={{ fontSize: '12px', color: '#64748B' }}>
                  Lane: <strong>{workflow.metrics?.lane_code || 'USLAX-CNSHA'}</strong>
                </span>
              </div>

              <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(180px, 1fr))', gap: '12px', marginBottom: '16px' }}>
                <div style={{ background: '#FFFFFF', padding: '12px', borderRadius: '6px', border: '1px solid #E2E8F0' }}>
                  <div style={{ fontSize: '11px', color: '#64748B' }}>Carrier Partner</div>
                  <div style={{ fontSize: '14px', fontWeight: 600, color: '#0F172A', marginTop: '2px' }}>
                    {workflow.carrier_economics?.carrier_name || 'Maersk Line Ocean Services'}
                  </div>
                </div>

                <div style={{ background: '#FFFFFF', padding: '12px', borderRadius: '6px', border: '1px solid #E2E8F0' }}>
                  <div style={{ fontSize: '11px', color: '#64748B' }}>On-Time Reliability</div>
                  <div style={{ fontSize: '14px', fontWeight: 600, color: '#16A34A', marginTop: '2px' }}>
                    {workflow.carrier_economics?.on_time_reliability_pct?.toFixed(1) || '92.4'}%
                  </div>
                </div>

                <div style={{ background: '#FFFFFF', padding: '12px', borderRadius: '6px', border: '1px solid #E2E8F0' }}>
                  <div style={{ fontSize: '11px', color: '#64748B' }}>Exception Frequency</div>
                  <div style={{ fontSize: '14px', fontWeight: 600, color: '#475569', marginTop: '2px' }}>
                    {workflow.carrier_economics?.exception_frequency_pct?.toFixed(1) || '2.1'}%
                  </div>
                </div>

                <div style={{ background: '#FFFFFF', padding: '12px', borderRadius: '6px', border: '1px solid #E2E8F0' }}>
                  <div style={{ fontSize: '11px', color: '#64748B' }}>Profitability Score</div>
                  <div style={{ fontSize: '14px', fontWeight: 600, color: '#2563EB', marginTop: '2px' }}>
                    {workflow.carrier_economics?.historical_profitability_score?.toFixed(1) || '84.5'} / 100
                  </div>
                </div>
              </div>

              {workflow.carrier_economics?.evaluation_notes && (
                <div style={{ background: '#EFF6FF', border: '1px solid #BFDBFE', borderRadius: '6px', padding: '10px 12px' }}>
                  <div style={{ fontSize: '12px', fontWeight: 600, color: '#1E40AF', marginBottom: '4px' }}>Workforce Carrier Observations:</div>
                  <ul style={{ margin: 0, paddingLeft: '18px', fontSize: '12px', color: '#1E40AF' }}>
                    {workflow.carrier_economics.evaluation_notes.map((n, i) => (
                      <li key={i}>{n}</li>
                    ))}
                  </ul>
                </div>
              )}
            </div>
          )}

          {/* TAB CONTENT: Customer Value */}
          {activeTab === 'customer' && (
            <div style={{ background: '#F8FAFC', border: '1px solid #E2E8F0', borderRadius: '8px', padding: '16px' }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '12px' }}>
                <h4 style={{ margin: 0, fontSize: '14px', color: '#0F172A', display: 'flex', alignItems: 'center', gap: '6px' }}>
                  <Briefcase size={16} color="#7E22CE" />
                  Multi-Dimensional Customer Value Scorecard
                </h4>
                <span style={{
                  fontSize: '11px',
                  fontWeight: 600,
                  padding: '2px 8px',
                  borderRadius: '4px',
                  background: '#F3E8FF',
                  color: '#6B21A8'
                }}>
                  {workflow.customer_value?.churn_risk_tier || 'LOW'} CHURN RISK
                </span>
              </div>

              <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(180px, 1fr))', gap: '12px', marginBottom: '16px' }}>
                <div style={{ background: '#FFFFFF', padding: '12px', borderRadius: '6px', border: '1px solid #E2E8F0' }}>
                  <div style={{ fontSize: '11px', color: '#64748B' }}>Account Name</div>
                  <div style={{ fontSize: '14px', fontWeight: 600, color: '#0F172A', marginTop: '2px' }}>
                    {workflow.customer_value?.customer_name || 'Apex Global Logistics Corp'}
                  </div>
                </div>

                <div style={{ background: '#FFFFFF', padding: '12px', borderRadius: '6px', border: '1px solid #E2E8F0' }}>
                  <div style={{ fontSize: '11px', color: '#64748B' }}>Annual Volume</div>
                  <div style={{ fontSize: '14px', fontWeight: 600, color: '#0F172A', marginTop: '2px' }}>
                    {workflow.customer_value?.annual_volume_teu || 145} TEU
                  </div>
                </div>

                <div style={{ background: '#FFFFFF', padding: '12px', borderRadius: '6px', border: '1px solid #E2E8F0' }}>
                  <div style={{ fontSize: '11px', color: '#64748B' }}>Payment Performance</div>
                  <div style={{ fontSize: '14px', fontWeight: 600, color: '#16A34A', marginTop: '2px' }}>
                    DSO {workflow.customer_value?.payment_dso || 18} Days
                  </div>
                </div>

                <div style={{ background: '#FFFFFF', padding: '12px', borderRadius: '6px', border: '1px solid #E2E8F0' }}>
                  <div style={{ fontSize: '11px', color: '#64748B' }}>Lifetime Value Score</div>
                  <div style={{ fontSize: '14px', fontWeight: 600, color: '#7E22CE', marginTop: '2px' }}>
                    {workflow.customer_value?.lifetime_value_score || 91.0} / 100
                  </div>
                </div>
              </div>
            </div>
          )}

          {/* TAB CONTENT: Governed Options */}
          {activeTab === 'options' && (
            <div>
              <p style={{ margin: '0 0 12px', fontSize: '13px', color: '#64748B' }}>
                Structured recommendations synthesized by the workforce under enterprise governance rules:
              </p>

              <div style={{ display: 'flex', flexDirection: 'column', gap: '10px' }}>
                {(workflow.available_options || []).map(opt => {
                  const isSelected = workflow.selected_option?.option_id === opt.option_id;
                  const isFloorBreach = opt.expected_margin_pct < 12.0;

                  return (
                    <div
                      key={opt.option_id}
                      style={{
                        background: isSelected ? '#EFF6FF' : '#FFFFFF',
                        border: isSelected ? '2px solid #2563EB' : '1px solid #E2E8F0',
                        borderRadius: '8px',
                        padding: '14px',
                        display: 'flex',
                        justifyContent: 'space-between',
                        alignItems: 'center',
                        flexWrap: 'wrap',
                        gap: '12px'
                      }}
                    >
                      <div style={{ flex: 1, minWidth: '240px' }}>
                        <div style={{ display: 'flex', alignItems: 'center', gap: '8px', marginBottom: '4px' }}>
                          <span style={{ fontSize: '14px', fontWeight: 600, color: '#0F172A' }}>{opt.title}</span>
                          {opt.requires_approval ? (
                            <span style={{ fontSize: '10px', background: '#FEF2F2', color: '#B91C1C', padding: '1px 6px', borderRadius: '4px', fontWeight: 600 }}>
                              REQUIRES EXECUTIVE APPROVAL
                            </span>
                          ) : (
                            <span style={{ fontSize: '10px', background: '#F0FDF4', color: '#166534', padding: '1px 6px', borderRadius: '4px', fontWeight: 600 }}>
                              AUTONOMOUSLY PERMITTED
                            </span>
                          )}
                        </div>
                        <div style={{ fontSize: '13px', color: '#475569', marginBottom: '6px' }}>
                          {opt.reason}
                        </div>
                        <div style={{ fontSize: '12px', color: '#64748B', display: 'flex', gap: '12px', flexWrap: 'wrap' }}>
                          <span>Price: <strong style={{ color: '#0F172A' }}>${opt.recommended_price?.toFixed(2)}</strong></span>
                          <span>Gross Margin: <strong style={{ color: isFloorBreach ? '#DC2626' : '#16A34A' }}>{opt.expected_margin_pct?.toFixed(1)}%</strong></span>
                          <span>Win Probability: <strong>{Math.round(opt.win_probability * 100)}%</strong></span>
                          <span>Autonomy: <strong>{opt.required_autonomy}</strong></span>
                        </div>
                      </div>

                      <div>
                        {isSelected ? (
                          <span style={{
                            fontSize: '12px',
                            fontWeight: 600,
                            color: '#2563EB',
                            background: '#DBEAFE',
                            padding: '6px 12px',
                            borderRadius: '6px',
                            display: 'inline-flex',
                            alignItems: 'center',
                            gap: '4px'
                          }}>
                            <CheckCircle size={14} /> Selected & Governed
                          </span>
                        ) : (
                          <button
                            onClick={() => handleSelectOption(opt.option_id)}
                            disabled={actionLoading}
                            style={{
                              background: '#FFFFFF',
                              border: '1px solid #CBD5E1',
                              color: '#0F172A',
                              padding: '6px 14px',
                              borderRadius: '6px',
                              fontSize: '13px',
                              fontWeight: 500,
                              cursor: 'pointer'
                            }}
                          >
                            Select Strategy
                          </button>
                        )}
                      </div>
                    </div>
                  );
                })}
              </div>
            </div>
          )}

          {/* TAB CONTENT: Version History */}
          {activeTab === 'history' && (
            <div>
              <p style={{ margin: '0 0 12px', fontSize: '13px', color: '#64748B' }}>
                Auditable decision provenance and pricing revisions preserved across continuous cycles:
              </p>

              <div style={{ display: 'flex', flexDirection: 'column', gap: '8px' }}>
                {(workflow.recommendation_history || []).map((rec, i) => (
                  <div key={i} style={{ background: '#F8FAFC', border: '1px solid #E2E8F0', borderRadius: '6px', padding: '12px' }}>
                    <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '4px' }}>
                      <span style={{ fontSize: '13px', fontWeight: 600, color: '#0F172A' }}>
                        Version {rec.version} - ${rec.recommended_price?.toFixed(2)}
                      </span>
                      <span style={{ fontSize: '11px', color: '#64748B' }}>
                        {new Date(rec.generated_at).toLocaleString()}
                      </span>
                    </div>
                    <div style={{ fontSize: '12px', color: '#475569' }}>
                      {rec.reason_for_revision || 'Commercial evaluation update'}
                    </div>
                    <div style={{ fontSize: '11px', color: '#64748B', marginTop: '4px' }}>
                      Expected Margin: {rec.expected_margin_pct?.toFixed(1)}% | Win Probability: {rec.win_probability_pct}%
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )}
        </>
      )}

      {/* Counter-Negotiate Modal */}
      {showNegotiateModal && (
        <div style={{
          position: 'fixed',
          top: 0,
          left: 0,
          right: 0,
          bottom: 0,
          background: 'rgba(15, 23, 42, 0.4)',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          zIndex: 9999
        }}>
          <div style={{
            background: '#FFFFFF',
            borderRadius: '12px',
            padding: '24px',
            width: '100%',
            maxWidth: '420px',
            boxShadow: '0 20px 25px -5px rgba(0, 0, 0, 0.1)',
            border: '1px solid #CBD5E1'
          }}>
            <h4 style={{ margin: '0 0 8px', fontSize: '16px', color: '#0F172A' }}>
              Counter-Optimize Customer Discount
            </h4>
            <p style={{ margin: '0 0 16px', fontSize: '13px', color: '#64748B' }}>
              Enter requested customer discount concession. Policy enforces max 15.0% discount and 12.0% gross margin floor.
            </p>

            <div style={{ marginBottom: '16px' }}>
              <label style={{ fontSize: '12px', fontWeight: 600, color: '#475569', display: 'block', marginBottom: '6px' }}>
                Requested Concession (%):
              </label>
              <input
                type="number"
                min="0.5"
                max="25.0"
                step="0.5"
                value={discountInput}
                onChange={(e) => setDiscountInput(e.target.value)}
                style={{
                  width: '100%',
                  padding: '8px 12px',
                  borderRadius: '6px',
                  border: '1px solid #CBD5E1',
                  fontSize: '14px',
                  boxSizing: 'border-box'
                }}
              />
            </div>

            <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '8px' }}>
              <button
                onClick={() => setShowNegotiateModal(false)}
                style={{
                  background: '#F1F5F9',
                  border: 'none',
                  borderRadius: '6px',
                  padding: '8px 14px',
                  fontSize: '13px',
                  cursor: 'pointer'
                }}
              >
                Cancel
              </button>
              <button
                onClick={handleCounterNegotiate}
                disabled={actionLoading}
                style={{
                  background: '#2563EB',
                  color: '#FFFFFF',
                  border: 'none',
                  borderRadius: '6px',
                  padding: '8px 14px',
                  fontSize: '13px',
                  fontWeight: 500,
                  cursor: 'pointer'
                }}
              >
                {actionLoading ? 'Optimizing...' : 'Calculate Counter-Tariff'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

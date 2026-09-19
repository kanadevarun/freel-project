import React, { useState, useEffect, useCallback } from 'react';
import {
  Users,
  ShieldCheck,
  Cpu,
  Search,
  CheckCircle,
  FileCheck,
  RotateCw,
  AlertCircle,
  TrendingUp,
  Clock,
  ArrowRight,
  Zap,
  Activity,
  HeartHandshake,
  MessageSquare,
  DollarSign
} from 'lucide-react';
import { enterpriseService } from '../../../../services/enterpriseService';

const HEALTH_COLORS = {
  HEALTHY: { bg: '#F0FDF4', text: '#15803D', border: '#BBF7D0' },
  STABLE: { bg: '#EFF6FF', text: '#1D4ED8', border: '#BFDBFE' },
  DECLINING: { bg: '#FFFBEB', text: '#B45309', border: '#FDE68A' },
  AT_RISK: { bg: '#FEF2F2', text: '#B91C1C', border: '#FECACA' },
  HIGH_RISK: { bg: '#450A0A', text: '#FEF2F2', border: '#991B1B' },
  OPPORTUNITY: { bg: '#FAF5FF', text: '#7E22CE', border: '#E9D5FF' },
  INACTIVE: { bg: '#F1F5F9', text: '#475569', border: '#CBD5E1' }
};

const STAGE_LABELS = {
  MONITORING: { label: 'Active Monitoring', bg: '#F8FAFC', text: '#475569' },
  HEALTH_ASSESSMENT: { label: 'Health Assessing', bg: '#EFF6FF', text: '#2563EB' },
  RISK_OPPORTUNITY_DETECTION: { label: 'Signals Detected', bg: '#FAF5FF', text: '#7C3AED' },
  INVESTIGATING: { label: 'Multi-Agent Investigation', bg: '#F0FDF4', text: '#16A34A' },
  PLANNING_INTERVENTION: { label: 'Planning Interventions', bg: '#FFFBEB', text: '#D97706' },
  WAITING_FOR_APPROVAL: { label: 'Waiting Approval', bg: '#FEF2F2', text: '#DC2626' },
  EXECUTING: { label: 'Executing Intervention', bg: '#EFF6FF', text: '#0284C7' },
  VERIFYING: { label: 'Verifying Delivery', bg: '#F5F3FF', text: '#6D28D9' },
  OUTCOME_TRACKING: { label: 'Outcome Tracking', bg: '#ECFDF5', text: '#059669' },
  COMPLETED: { label: 'Cycle Completed', bg: '#F0FDF4', text: '#15803D' },
  ESCALATED: { label: 'Escalated to Director', bg: '#FEF2F2', text: '#DC2626' }
};

export default function AutonomousCustomerRelationshipCard({ customerId }) {
  const [workflow, setWorkflow] = useState(null);
  const [loading, setLoading] = useState(true);
  const [actionLoading, setActionLoading] = useState(false);
  const [error, setError] = useState(null);
  const [selectedOptionId, setSelectedOptionId] = useState('');

  const fetchWorkflow = useCallback(async () => {
    if (!customerId) return;
    setLoading(true);
    setError(null);
    try {
      const res = await enterpriseService.getCustomerWorkflow(customerId);
      const data = res?.data || res;
      setWorkflow(data);
      if (data?.intervention_options?.length > 0) {
        setSelectedOptionId(data.intervention_options[0].option_id);
      }
    } catch (err) {
      if (err?.response?.status === 404) {
        setWorkflow(null);
      } else {
        setError(err?.response?.data?.error || err?.message || 'Failed to load autonomous CRM workflow');
      }
    } finally {
      setLoading(false);
    }
  }, [customerId]);

  useEffect(() => {
    fetchWorkflow();
  }, [fetchWorkflow]);

  const handleInitiate = async () => {
    setActionLoading(true);
    setError(null);
    try {
      const res = await enterpriseService.initiateCustomerWorkflow(customerId);
      setWorkflow(res?.data || res);
    } catch (err) {
      setError(err?.response?.data?.error || err?.message || 'Failed to initiate CRM orchestrator');
    } finally {
      setActionLoading(false);
    }
  };

  const handleEvaluateHealth = async () => {
    if (!workflow?.workflow_id) return;
    setActionLoading(true);
    setError(null);
    try {
      const res = await enterpriseService.evaluateCustomerHealth(workflow.workflow_id);
      setWorkflow(res?.data || res);
    } catch (err) {
      setError(err?.response?.data?.error || err?.message || 'Health evaluation failed');
    } finally {
      setActionLoading(false);
    }
  };

  const handleDetectSignals = async () => {
    if (!workflow?.workflow_id) return;
    setActionLoading(true);
    setError(null);
    try {
      const res = await enterpriseService.detectCustomerRisksAndOpportunities(workflow.workflow_id);
      setWorkflow(res?.data || res);
    } catch (err) {
      setError(err?.response?.data?.error || err?.message || 'Risk/Opportunity detection failed');
    } finally {
      setActionLoading(false);
    }
  };

  const handleInvestigate = async () => {
    if (!workflow?.workflow_id) return;
    setActionLoading(true);
    setError(null);
    try {
      const res = await enterpriseService.investigateCustomerRootCauses(workflow.workflow_id);
      setWorkflow(res?.data || res);
    } catch (err) {
      setError(err?.response?.data?.error || err?.message || 'Multi-agent investigation failed');
    } finally {
      setActionLoading(false);
    }
  };

  const handlePlanInterventions = async () => {
    if (!workflow?.workflow_id) return;
    setActionLoading(true);
    setError(null);
    try {
      const res = await enterpriseService.planCustomerInterventions(workflow.workflow_id);
      const data = res?.data || res;
      setWorkflow(data);
      if (data?.intervention_options?.length > 0) {
        setSelectedOptionId(data.intervention_options[0].option_id);
      }
    } catch (err) {
      setError(err?.response?.data?.error || err?.message || 'Intervention planning failed');
    } finally {
      setActionLoading(false);
    }
  };

  const handleSelectOption = async () => {
    if (!workflow?.workflow_id || !selectedOptionId) return;
    setActionLoading(true);
    setError(null);
    try {
      const res = await enterpriseService.selectCustomerInterventionOption(workflow.workflow_id, selectedOptionId);
      setWorkflow(res?.data || res);
    } catch (err) {
      setError(err?.response?.data?.error || err?.message || 'Selecting option failed');
    } finally {
      setActionLoading(false);
    }
  };

  const handleExecute = async () => {
    if (!workflow?.workflow_id || !workflow?.selected_option) return;
    setActionLoading(true);
    setError(null);
    try {
      const res = await enterpriseService.executeCustomerIntervention(workflow.workflow_id, workflow.selected_option.option_id);
      setWorkflow(res?.data || res);
    } catch (err) {
      setError(err?.response?.data?.error || err?.message || 'Executing intervention failed');
    } finally {
      setActionLoading(false);
    }
  };

  const handleVerify = async () => {
    if (!workflow?.workflow_id || !workflow?.selected_option) return;
    setActionLoading(true);
    setError(null);
    try {
      const res = await enterpriseService.verifyCustomerIntervention(workflow.workflow_id, workflow.selected_option.option_id);
      setWorkflow(res?.data || res);
    } catch (err) {
      setError(err?.response?.data?.error || err?.message || 'Verification failed');
    } finally {
      setActionLoading(false);
    }
  };

  const handleSimulateResponse = async () => {
    if (!workflow?.workflow_id) return;
    const msg = window.prompt(
      'Simulate incoming customer message (tested against prompt-injection):',
      'Thank you for the prompt courtesy credit on our invoice. We would also like to explore reefer rates on USLAX-CNSHA.'
    );
    if (!msg) return;
    setActionLoading(true);
    setError(null);
    try {
      const res = await enterpriseService.processCustomerResponse(workflow.workflow_id, msg);
      setWorkflow(res?.data || res);
    } catch (err) {
      setError(err?.response?.data?.error || err?.message || 'Processing response failed');
    } finally {
      setActionLoading(false);
    }
  };

  const handleRecordOutcome = async () => {
    if (!workflow?.workflow_id) return;
    setActionLoading(true);
    setError(null);
    try {
      const res = await enterpriseService.recordCustomerOutcome(workflow.workflow_id, {
        customer_id: workflow.customer_id,
        intervention_success: true,
        customer_retention_status: 'EXPANDED',
        revenue_impact: 34500.00,
        customer_satisfaction_score: 'HIGH',
        lessons_learned: 'Proactive service adjustment eliminated billing dispute and unlocked high-volume reefer quotation'
      });
      setWorkflow(res?.data || res);
    } catch (err) {
      setError(err?.response?.data?.error || err?.message || 'Recording outcome failed');
    } finally {
      setActionLoading(false);
    }
  };

  const healthStyle = HEALTH_COLORS[workflow?.health_profile?.category] || HEALTH_COLORS.STABLE;
  const stageStyle = STAGE_LABELS[workflow?.current_stage] || STAGE_LABELS.MONITORING;

  return (
    <div
      style={{
        backgroundColor: '#FFFFFF',
        borderRadius: '12px',
        border: '1px solid #E2E8F0',
        padding: '24px',
        boxShadow: '0 1px 3px rgba(0, 0, 0, 0.05)',
        fontFamily: 'Inter, system-ui, sans-serif',
        marginBottom: '24px'
      }}
    >
      {/* Header */}
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: '20px' }}>
        <div>
          <div style={{ display: 'flex', alignItems: 'center', gap: '8px', marginBottom: '6px' }}>
            <HeartHandshake size={20} color="#2563EB" />
            <h3 style={{ margin: 0, fontSize: '18px', fontWeight: 600, color: '#0F172A' }}>
              Autonomous Customer Relationship Management
            </h3>
            <span
              style={{
                fontSize: '11px',
                fontWeight: 600,
                padding: '2px 8px',
                borderRadius: '9999px',
                backgroundColor: '#EFF6FF',
                color: '#2563EB',
                border: '1px solid #BFDBFE'
              }}
            >
              Phase 7.5 Governed
            </span>
          </div>
          <p style={{ margin: 0, fontSize: '13px', color: '#64748B' }}>
            Continuous account telemetry, predictive health analysis, governed intervention planning, and memory learning
          </p>
        </div>

        {workflow && (
          <div style={{ display: 'flex', gap: '8px', alignItems: 'center' }}>
            {workflow.health_profile && (
              <span
                style={{
                  fontSize: '12px',
                  fontWeight: 600,
                  padding: '4px 10px',
                  borderRadius: '6px',
                  backgroundColor: healthStyle.bg,
                  color: healthStyle.text,
                  border: `1px solid ${healthStyle.border}`
                }}
              >
                {workflow.health_profile.category} ({workflow.health_profile.health_score.toFixed(0)}/100)
              </span>
            )}
            <span
              style={{
                fontSize: '12px',
                fontWeight: 600,
                padding: '4px 10px',
                borderRadius: '6px',
                backgroundColor: stageStyle.bg,
                color: stageStyle.text,
                border: '1px solid #CBD5E1'
              }}
            >
              {stageStyle.label}
            </span>
          </div>
        )}
      </div>

      {error && (
        <div
          style={{
            display: 'flex',
            alignItems: 'center',
            gap: '8px',
            padding: '12px 16px',
            backgroundColor: '#FEF2F2',
            border: '1px solid #FECACA',
            borderRadius: '8px',
            color: '#B91C1C',
            fontSize: '13px',
            marginBottom: '16px'
          }}
        >
          <AlertCircle size={16} />
          <span>{error}</span>
        </div>
      )}

      {loading ? (
        <div style={{ textAlign: 'center', padding: '32px 0', color: '#64748B', fontSize: '14px' }}>
          Loading customer relationship telemetry...
        </div>
      ) : !workflow ? (
        <div
          style={{
            textAlign: 'center',
            padding: '32px 16px',
            backgroundColor: '#F8FAFC',
            borderRadius: '8px',
            border: '1px dashed #CBD5E1'
          }}
        >
          <Users size={32} color="#3B82F6" style={{ margin: '0 auto 12px' }} />
          <h4 style={{ margin: '0 0 6px', fontSize: '15px', fontWeight: 600, color: '#1E293B' }}>
            Autonomous CRM Orchestration Inactive
          </h4>
          <p style={{ margin: '0 0 16px', fontSize: '13px', color: '#64748B' }}>
            Activate continuous customer relationship management to monitor account health, detect risks/opportunities, and govern interventions.
          </p>
          <button
            onClick={handleInitiate}
            disabled={actionLoading}
            style={{
              padding: '8px 16px',
              backgroundColor: '#2563EB',
              color: '#FFFFFF',
              border: 'none',
              borderRadius: '6px',
              fontSize: '13px',
              fontWeight: 500,
              cursor: actionLoading ? 'not-allowed' : 'pointer',
              display: 'inline-flex',
              alignItems: 'center',
              gap: '6px'
            }}
          >
            <Zap size={14} />
            Initialize Autonomous CRM Workflow
          </button>
        </div>
      ) : (
        <div>
          {/* Section 1: Customer Health Breakdown (Facts vs Predictions vs Recommendations) */}
          {workflow.health_profile && (
            <div style={{ marginBottom: '20px', padding: '16px', borderRadius: '8px', border: '1px solid #E2E8F0', backgroundColor: '#FFFFFF' }}>
              <div style={{ display: 'flex', alignItems: 'center', gap: '8px', marginBottom: '12px' }}>
                <Activity size={16} color="#2563EB" />
                <h4 style={{ margin: 0, fontSize: '14px', fontWeight: 600, color: '#0F172A' }}>
                  Customer Health Profile (Confidence: {(workflow.health_profile.confidence * 100).toFixed(0)}%)
                </h4>
              </div>
              <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(240px, 1fr))', gap: '16px' }}>
                <div>
                  <div style={{ fontSize: '12px', fontWeight: 600, color: '#059669', marginBottom: '6px' }}>
                    Authoritative Facts
                  </div>
                  <ul style={{ margin: 0, paddingLeft: '18px', fontSize: '12px', color: '#334155' }}>
                    {workflow.health_profile.confirmed_facts?.map((fact, idx) => (
                      <li key={idx} style={{ marginBottom: '3px' }}>{fact}</li>
                    ))}
                  </ul>
                </div>
                <div>
                  <div style={{ fontSize: '12px', fontWeight: 600, color: '#2563EB', marginBottom: '6px' }}>
                    AI Predictions
                  </div>
                  <ul style={{ margin: 0, paddingLeft: '18px', fontSize: '12px', color: '#334155' }}>
                    {workflow.health_profile.ai_predictions?.map((pred, idx) => (
                      <li key={idx} style={{ marginBottom: '3px' }}>{pred}</li>
                    ))}
                  </ul>
                </div>
                <div>
                  <div style={{ fontSize: '12px', fontWeight: 600, color: '#7C3AED', marginBottom: '6px' }}>
                    Governed Recommendations
                  </div>
                  <ul style={{ margin: 0, paddingLeft: '18px', fontSize: '12px', color: '#334155' }}>
                    {workflow.health_profile.recommendations?.map((rec, idx) => (
                      <li key={idx} style={{ marginBottom: '3px' }}>{rec}</li>
                    ))}
                  </ul>
                </div>
              </div>
            </div>
          )}

          {/* Section 2: Detected Risks & Opportunities */}
          {(workflow.detected_risks?.length > 0 || workflow.detected_opportunities?.length > 0) && (
            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '16px', marginBottom: '20px' }}>
              {/* Risks */}
              <div style={{ padding: '16px', borderRadius: '8px', border: '1px solid #FECACA', backgroundColor: '#FEF2F2' }}>
                <div style={{ fontSize: '13px', fontWeight: 600, color: '#991B1B', marginBottom: '8px' }}>
                  Identified Account Risks ({workflow.detected_risks?.length || 0})
                </div>
                {workflow.detected_risks?.map((risk) => (
                  <div key={risk.signal_id} style={{ fontSize: '12px', color: '#7F1D1D', marginBottom: '6px' }}>
                    <strong>{risk.title}:</strong> {risk.description}
                  </div>
                ))}
              </div>

              {/* Opportunities */}
              <div style={{ padding: '16px', borderRadius: '8px', border: '1px solid #BBF7D0', backgroundColor: '#F0FDF4' }}>
                <div style={{ fontSize: '13px', fontWeight: 600, color: '#166534', marginBottom: '8px' }}>
                  Commercial Growth Opportunities ({workflow.detected_opportunities?.length || 0})
                </div>
                {workflow.detected_opportunities?.map((opp) => (
                  <div key={opp.signal_id} style={{ fontSize: '12px', color: '#14532D', marginBottom: '6px' }}>
                    <strong>{opp.title} (${opp.potential_value.toLocaleString()}):</strong> {opp.description}
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* Section 3: Governed Intervention Options */}
          {workflow.intervention_options?.length > 0 && (
            <div style={{ marginBottom: '20px', padding: '16px', borderRadius: '8px', border: '1px solid #E2E8F0', backgroundColor: '#FFFFFF' }}>
              <div style={{ display: 'flex', alignItems: 'center', gap: '8px', marginBottom: '12px' }}>
                <ShieldCheck size={16} color="#16A34A" />
                <h4 style={{ margin: 0, fontSize: '14px', fontWeight: 600, color: '#0F172A' }}>
                  Structured Intervention Options
                </h4>
              </div>
              <div style={{ display: 'flex', flexDirection: 'column', gap: '10px' }}>
                {workflow.intervention_options.map((opt) => (
                  <div
                    key={opt.option_id}
                    onClick={() => setSelectedOptionId(opt.option_id)}
                    style={{
                      padding: '12px',
                      borderRadius: '6px',
                      border: `1px solid ${selectedOptionId === opt.option_id ? '#2563EB' : '#E2E8F0'}`,
                      backgroundColor: selectedOptionId === opt.option_id ? '#EFF6FF' : '#FFFFFF',
                      cursor: 'pointer',
                      display: 'flex',
                      justifyContent: 'space-between',
                      alignItems: 'center'
                    }}
                  >
                    <div>
                      <div style={{ fontSize: '13px', fontWeight: 600, color: '#0F172A' }}>
                        {opt.title} <span style={{ fontSize: '11px', color: '#64748B' }}>({opt.action_type})</span>
                      </div>
                      <div style={{ fontSize: '12px', color: '#475569', marginTop: '2px' }}>{opt.reason}</div>
                    </div>
                    <div style={{ display: 'flex', gap: '8px', alignItems: 'center' }}>
                      <span
                        style={{
                          fontSize: '11px',
                          fontWeight: 500,
                          padding: '2px 6px',
                          borderRadius: '4px',
                          backgroundColor: opt.requires_approval ? '#FFFBEB' : '#F0FDF4',
                          color: opt.requires_approval ? '#B45309' : '#15803D'
                        }}
                      >
                        {opt.requires_approval ? 'Approval Required' : 'Autonomous'}
                      </span>
                      <span
                        style={{
                          fontSize: '11px',
                          fontWeight: 600,
                          color: '#2563EB',
                          backgroundColor: '#DBEAFE',
                          padding: '2px 6px',
                          borderRadius: '4px'
                        }}
                      >
                        {(opt.confidence * 100).toFixed(0)}%
                      </span>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* Section 4: Customer Response & Secure Parsing */}
          {workflow.last_response && (
            <div style={{ marginBottom: '20px', padding: '12px 16px', borderRadius: '8px', backgroundColor: '#F8FAFC', border: '1px solid #CBD5E1' }}>
              <div style={{ fontSize: '11px', color: '#64748B', fontWeight: 600, textTransform: 'uppercase' }}>
                Processed Customer Response ({workflow.last_response.intent} / {workflow.last_response.sentiment})
              </div>
              <div style={{ fontSize: '13px', color: '#1E293B', marginTop: '2px' }}>
                "{workflow.last_response.raw_message}"
              </div>
              <div style={{ fontSize: '12px', color: '#2563EB', marginTop: '4px' }}>
                Recommended Next Step: {workflow.last_response.recommended_next_step}
              </div>
            </div>
          )}

          {/* Action Toolbar */}
          <div style={{ display: 'flex', flexWrap: 'wrap', gap: '8px', paddingTop: '12px', borderTop: '1px solid #F1F5F9' }}>
            {workflow.current_stage === 'MONITORING' && (
              <button
                onClick={handleEvaluateHealth}
                disabled={actionLoading}
                style={{
                  padding: '7px 14px',
                  backgroundColor: '#2563EB',
                  color: '#FFFFFF',
                  border: 'none',
                  borderRadius: '6px',
                  fontSize: '12px',
                  fontWeight: 500,
                  cursor: 'pointer'
                }}
              >
                Evaluate Account Health
              </button>
            )}

            {workflow.current_stage === 'HEALTH_ASSESSMENT' && (
              <button
                onClick={handleDetectSignals}
                disabled={actionLoading}
                style={{
                  padding: '7px 14px',
                  backgroundColor: '#7C3AED',
                  color: '#FFFFFF',
                  border: 'none',
                  borderRadius: '6px',
                  fontSize: '12px',
                  fontWeight: 500,
                  cursor: 'pointer'
                }}
              >
                Detect Risks & Opportunities
              </button>
            )}

            {workflow.current_stage === 'RISK_OPPORTUNITY_DETECTION' && (
              <button
                onClick={handleInvestigate}
                disabled={actionLoading}
                style={{
                  padding: '7px 14px',
                  backgroundColor: '#059669',
                  color: '#FFFFFF',
                  border: 'none',
                  borderRadius: '6px',
                  fontSize: '12px',
                  fontWeight: 500,
                  cursor: 'pointer'
                }}
              >
                Investigate Causes
              </button>
            )}

            {workflow.current_stage === 'INVESTIGATING' && (
              <button
                onClick={handlePlanInterventions}
                disabled={actionLoading}
                style={{
                  padding: '7px 14px',
                  backgroundColor: '#2563EB',
                  color: '#FFFFFF',
                  border: 'none',
                  borderRadius: '6px',
                  fontSize: '12px',
                  fontWeight: 500,
                  cursor: 'pointer'
                }}
              >
                Plan Interventions
              </button>
            )}

            {workflow.current_stage === 'PLANNING_INTERVENTION' && (
              <button
                onClick={handleSelectOption}
                disabled={actionLoading || !selectedOptionId}
                style={{
                  padding: '7px 14px',
                  backgroundColor: '#2563EB',
                  color: '#FFFFFF',
                  border: 'none',
                  borderRadius: '6px',
                  fontSize: '12px',
                  fontWeight: 500,
                  cursor: 'pointer'
                }}
              >
                Govern & Select Option
              </button>
            )}

            {workflow.current_stage === 'EXECUTING' && (
              <>
                <button
                  onClick={handleExecute}
                  disabled={actionLoading}
                  style={{
                    padding: '7px 14px',
                    backgroundColor: '#0284C7',
                    color: '#FFFFFF',
                    border: 'none',
                    borderRadius: '6px',
                    fontSize: '12px',
                    fontWeight: 500,
                    cursor: 'pointer'
                  }}
                >
                  Execute Intervention
                </button>
                <button
                  onClick={handleVerify}
                  disabled={actionLoading}
                  style={{
                    padding: '7px 14px',
                    backgroundColor: '#6D28D9',
                    color: '#FFFFFF',
                    border: 'none',
                    borderRadius: '6px',
                    fontSize: '12px',
                    fontWeight: 500,
                    cursor: 'pointer'
                  }}
                >
                  Verify Delivery
                </button>
              </>
            )}

            {workflow.current_stage === 'VERIFYING' && (
              <button
                onClick={handleSimulateResponse}
                disabled={actionLoading}
                style={{
                  padding: '7px 14px',
                  backgroundColor: '#16A34A',
                  color: '#FFFFFF',
                  border: 'none',
                  borderRadius: '6px',
                  fontSize: '12px',
                  fontWeight: 500,
                  cursor: 'pointer'
                }}
              >
                Simulate Customer Response
              </button>
            )}

            {workflow.current_stage !== 'COMPLETED' && (
              <button
                onClick={handleRecordOutcome}
                disabled={actionLoading}
                style={{
                  padding: '7px 14px',
                  backgroundColor: '#334155',
                  color: '#FFFFFF',
                  border: 'none',
                  borderRadius: '6px',
                  fontSize: '12px',
                  fontWeight: 500,
                  cursor: 'pointer'
                }}
              >
                Record Outcome & Train Memory
              </button>
            )}

            <button
              onClick={fetchWorkflow}
              disabled={loading || actionLoading}
              style={{
                padding: '7px 14px',
                backgroundColor: '#F8FAFC',
                color: '#475569',
                border: '1px solid #E2E8F0',
                borderRadius: '6px',
                fontSize: '12px',
                fontWeight: 500,
                cursor: 'pointer',
                marginLeft: 'auto'
              }}
            >
              Refresh
            </button>
          </div>
        </div>
      )}
    </div>
  );
}

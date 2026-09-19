import React, { useState, useEffect, useCallback } from 'react';
import {
  AlertTriangle,
  ShieldCheck,
  Cpu,
  Search,
  CheckCircle,
  FileCheck,
  RotateCw,
  AlertCircle,
  UserCheck,
  TrendingUp,
  Clock,
  ArrowRight,
  Zap,
  Activity
} from 'lucide-react';
import { enterpriseService } from '../../../../services/enterpriseService';

const SEVERITY_BADGES = {
  LOW: { bg: '#F1F5F9', text: '#475569', border: '#CBD5E1' },
  MEDIUM: { bg: '#EFF6FF', text: '#1D4ED8', border: '#BFDBFE' },
  HIGH: { bg: '#FFFBEB', text: '#B45309', border: '#FDE68A' },
  CRITICAL: { bg: '#FEF2F2', text: '#B91C1C', border: '#FECACA' }
};

const STAGE_BADGES = {
  DETECTED: { label: 'Detected', bg: '#F8FAFC', text: '#475569' },
  INVESTIGATING: { label: 'Investigating', bg: '#EFF6FF', text: '#2563EB' },
  IMPACT_ASSESSMENT: { label: 'Impact Assessing', bg: '#FAF5FF', text: '#7C3AED' },
  PLANNING_RECOVERY: { label: 'Planning Recovery', bg: '#F0FDF4', text: '#16A34A' },
  WAITING_FOR_APPROVAL: { label: 'Waiting Approval', bg: '#FFFBEB', text: '#D97706' },
  EXECUTING: { label: 'Executing Action', bg: '#EFF6FF', text: '#0284C7' },
  VERIFYING: { label: 'Verifying Result', bg: '#F5F3FF', text: '#6D28D9' },
  MONITORING: { label: 'Monitoring Stable', bg: '#ECFDF5', text: '#059669' },
  RESOLVED: { label: 'Resolved', bg: '#F0FDF4', text: '#15803D' },
  ESCALATED: { label: 'Escalated', bg: '#FEF2F2', text: '#DC2626' },
  CLOSED: { label: 'Closed', bg: '#F1F5F9', text: '#334155' }
};

export default function AutonomousExceptionLifecycleCard({ entityType = 'SHIPMENT', entityId }) {
  const [workflow, setWorkflow] = useState(null);
  const [loading, setLoading] = useState(true);
  const [actionLoading, setActionLoading] = useState(false);
  const [error, setError] = useState(null);
  const [selectedOptionId, setSelectedOptionId] = useState('');

  const excKey = `EXC-${entityType}-${entityId}`;

  const fetchWorkflow = useCallback(async () => {
    if (!entityId) return;
    setLoading(true);
    setError(null);
    try {
      const res = await enterpriseService.getExceptionWorkflow(excKey);
      setWorkflow(res?.data || res);
    } catch (err) {
      if (err?.response?.status === 404) {
        setWorkflow(null);
      } else {
        setError(err?.response?.data?.error || err?.message || 'Failed to load exception lifecycle');
      }
    } finally {
      setLoading(false);
    }
  }, [entityId, excKey]);

  useEffect(() => {
    fetchWorkflow();
  }, [fetchWorkflow]);

  const handleDetect = async () => {
    setActionLoading(true);
    setError(null);
    try {
      const res = await enterpriseService.detectException({
        event_id: `evt-man-${Date.now()}`,
        event_type: `${entityType}_OPERATIONAL_EXCEPTION`,
        domain: entityType === 'SHIPMENT' ? 'SHIPMENT' : 'COMMERCIAL',
        related_entity_type: entityType,
        related_entity_id: String(entityId),
        severity: 'HIGH',
        title: `${entityType} Transit Milestone Variance`,
        description: `Operational telemetry reported exception on ${entityType} #${entityId}. Autonomous investigation initiated.`,
        source: 'manual_trigger',
        payload: { entity_id: entityId }
      });
      setWorkflow(res?.data || res);
    } catch (err) {
      setError(err?.response?.data?.error || err?.message || 'Failed to trigger exception detection');
    } finally {
      setActionLoading(false);
    }
  };

  const handleInvestigate = async () => {
    if (!workflow?.workflow_id) return;
    setActionLoading(true);
    setError(null);
    try {
      const res = await enterpriseService.investigateException(workflow.workflow_id);
      setWorkflow(res?.data || res);
    } catch (err) {
      setError(err?.response?.data?.error || err?.message || 'Root-cause investigation failed');
    } finally {
      setActionLoading(false);
    }
  };

  const handleAssessImpact = async () => {
    if (!workflow?.workflow_id) return;
    setActionLoading(true);
    setError(null);
    try {
      const res = await enterpriseService.assessExceptionImpact(workflow.workflow_id);
      setWorkflow(res?.data || res);
    } catch (err) {
      setError(err?.response?.data?.error || err?.message || 'Cross-module impact analysis failed');
    } finally {
      setActionLoading(false);
    }
  };

  const handlePlanRecovery = async () => {
    if (!workflow?.workflow_id) return;
    setActionLoading(true);
    setError(null);
    try {
      const res = await enterpriseService.planExceptionRecovery(workflow.workflow_id);
      const data = res?.data || res;
      setWorkflow(data);
      if (data?.recovery_options?.length > 0) {
        setSelectedOptionId(data.recovery_options[0].option_id);
      }
    } catch (err) {
      setError(err?.response?.data?.error || err?.message || 'Recovery planning failed');
    } finally {
      setActionLoading(false);
    }
  };

  const handleSelectOption = async () => {
    if (!workflow?.workflow_id || !selectedOptionId) return;
    setActionLoading(true);
    setError(null);
    try {
      const res = await enterpriseService.selectExceptionRecoveryOption(workflow.workflow_id, selectedOptionId);
      setWorkflow(res?.data || res);
    } catch (err) {
      setError(err?.response?.data?.error || err?.message || 'Selecting recovery option failed');
    } finally {
      setActionLoading(false);
    }
  };

  const handleExecuteAction = async () => {
    if (!workflow?.workflow_id || !workflow?.selected_option) return;
    setActionLoading(true);
    setError(null);
    try {
      const res = await enterpriseService.executeExceptionRecoveryStep(workflow.workflow_id, workflow.selected_option.option_id);
      setWorkflow(res?.data || res);
    } catch (err) {
      setError(err?.response?.data?.error || err?.message || 'Executing recovery action failed');
    } finally {
      setActionLoading(false);
    }
  };

  const handleVerify = async () => {
    if (!workflow?.workflow_id || !workflow?.selected_option) return;
    setActionLoading(true);
    setError(null);
    try {
      await enterpriseService.verifyExceptionRecovery(workflow.workflow_id, workflow.selected_option.option_id);
      await fetchWorkflow();
    } catch (err) {
      setError(err?.response?.data?.error || err?.message || 'Action verification failed');
    } finally {
      setActionLoading(false);
    }
  };

  const handleReplan = async () => {
    if (!workflow?.workflow_id) return;
    const reason = window.prompt('Enter failure or reassessment reason for adaptive replanning:', 'Carrier secondary delay detected');
    if (!reason) return;
    setActionLoading(true);
    setError(null);
    try {
      const res = await enterpriseService.triggerAdaptiveRecovery(workflow.workflow_id, reason);
      setWorkflow(res?.data || res);
    } catch (err) {
      setError(err?.response?.data?.error || err?.message || 'Adaptive replan failed');
    } finally {
      setActionLoading(false);
    }
  };

  const handleTransitionMonitoring = async () => {
    if (!workflow?.workflow_id) return;
    setActionLoading(true);
    setError(null);
    try {
      const res = await enterpriseService.transitionExceptionToMonitoring(workflow.workflow_id, 'next_connecting_carrier_check');
      setWorkflow(res?.data || res);
    } catch (err) {
      setError(err?.response?.data?.error || err?.message || 'Monitoring transition failed');
    } finally {
      setActionLoading(false);
    }
  };

  const handleResolve = async () => {
    if (!workflow?.workflow_id) return;
    const proof = window.prompt('Enter authoritative resolution proof (required by Go governance):', 'Authoritative carrier EDI confirmation & tracking status normal');
    if (!proof) return;
    setActionLoading(true);
    setError(null);
    try {
      const res = await enterpriseService.resolveException(workflow.workflow_id, proof);
      setWorkflow(res?.data || res);
    } catch (err) {
      setError(err?.response?.data?.error || err?.message || 'Resolution failed');
    } finally {
      setActionLoading(false);
    }
  };

  const handleEscalate = async () => {
    if (!workflow?.workflow_id) return;
    const reason = window.prompt('Enter escalation reason for human operations:', 'Exceeded automated resolution threshold');
    if (!reason) return;
    setActionLoading(true);
    setError(null);
    try {
      const res = await enterpriseService.escalateException(workflow.workflow_id, reason);
      setWorkflow(res?.data || res);
    } catch (err) {
      setError(err?.response?.data?.error || err?.message || 'Escalation failed');
    } finally {
      setActionLoading(false);
    }
  };

  const handleRecordOutcome = async () => {
    if (!workflow?.workflow_id) return;
    setActionLoading(true);
    setError(null);
    try {
      const res = await enterpriseService.recordExceptionOutcome(workflow.workflow_id, {
        exception_id: workflow.exception_id,
        resolved_successfully: true,
        operational_recovery_time_hours: 12.0,
        actual_financial_cost: workflow.impact_assessment?.estimated_cost_impact || 0.0,
        customer_satisfaction: 'high',
        human_intervention: false,
        lesson_learned: 'Proactive customer communication and expedited corridor prevented dispute'
      });
      setWorkflow(res?.data || res);
    } catch (err) {
      setError(err?.response?.data?.error || err?.message || 'Recording outcome failed');
    } finally {
      setActionLoading(false);
    }
  };

  const sevStyle = SEVERITY_BADGES[workflow?.severity] || SEVERITY_BADGES.MEDIUM;
  const stageStyle = STAGE_BADGES[workflow?.current_stage] || STAGE_BADGES.DETECTED;

  return (
    <div
      style={{
        backgroundColor: '#FFFFFF',
        borderRadius: '12px',
        border: '1px solid #E2E8F0',
        padding: '24px',
        boxShadow: '0 1px 3px rgba(0, 0, 0, 0.05)',
        fontFamily: 'Inter, system-ui, sans-serif'
      }}
    >
      {/* Header */}
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: '20px' }}>
        <div>
          <div style={{ display: 'flex', alignItems: 'center', gap: '8px', marginBottom: '6px' }}>
            <AlertTriangle size={20} color="#D97706" />
            <h3 style={{ margin: 0, fontSize: '18px', fontWeight: 600, color: '#0F172A' }}>
              Autonomous Enterprise Exception Management
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
              Phase 7.4 Governed
            </span>
          </div>
          <p style={{ margin: 0, fontSize: '13px', color: '#64748B' }}>
            Detect → Investigate → Assess Impact → Plan Recovery → Govern → Act → Verify → Learn
          </p>
        </div>

        {workflow && (
          <div style={{ display: 'flex', gap: '8px', alignItems: 'center' }}>
            <span
              style={{
                fontSize: '12px',
                fontWeight: 600,
                padding: '4px 10px',
                borderRadius: '6px',
                backgroundColor: sevStyle.bg,
                color: sevStyle.text,
                border: `1px solid ${sevStyle.border}`
              }}
            >
              {workflow.severity} SEVERITY
            </span>
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
          Loading exception intelligence & governance state...
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
          <CheckCircle size={32} color="#10B981" style={{ margin: '0 auto 12px' }} />
          <h4 style={{ margin: '0 0 6px', fontSize: '15px', fontWeight: 600, color: '#1E293B' }}>
            No Active Operational Exception
          </h4>
          <p style={{ margin: '0 0 16px', fontSize: '13px', color: '#64748B' }}>
            Telemetry and milestone tracking show normal operations. You can ingest an exception signal to test governed autonomous recovery.
          </p>
          <button
            onClick={handleDetect}
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
            Simulate Governed Exception Intake
          </button>
        </div>
      ) : (
        <div>
          {/* Metadata Grid */}
          <div
            style={{
              display: 'grid',
              gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))',
              gap: '12px',
              marginBottom: '20px',
              backgroundColor: '#F8FAFC',
              padding: '16px',
              borderRadius: '8px',
              border: '1px solid #E2E8F0'
            }}
          >
            <div>
              <div style={{ fontSize: '11px', color: '#64748B', textTransform: 'uppercase', fontWeight: 600 }}>Exception ID</div>
              <div style={{ fontSize: '13px', fontWeight: 600, color: '#0F172A', marginTop: '2px' }}>{workflow.exception_id}</div>
            </div>
            <div>
              <div style={{ fontSize: '11px', color: '#64748B', textTransform: 'uppercase', fontWeight: 600 }}>Domain / Type</div>
              <div style={{ fontSize: '13px', fontWeight: 500, color: '#0F172A', marginTop: '2px' }}>
                {workflow.domain} / {workflow.exception_type}
              </div>
            </div>
            <div>
              <div style={{ fontSize: '11px', color: '#64748B', textTransform: 'uppercase', fontWeight: 600 }}>Assigned Workforce</div>
              <div style={{ fontSize: '13px', fontWeight: 500, color: '#2563EB', marginTop: '2px' }}>
                {workflow.assigned_specialists?.join(', ') || 'exception_agent'}
              </div>
            </div>
            <div>
              <div style={{ fontSize: '11px', color: '#64748B', textTransform: 'uppercase', fontWeight: 600 }}>Replan & Retries</div>
              <div style={{ fontSize: '13px', fontWeight: 500, color: '#0F172A', marginTop: '2px' }}>
                Version {workflow.replan_version} (Retries: {workflow.retry_count}/{workflow.max_retries})
              </div>
            </div>
          </div>

          {/* Section 1: Root Cause Analysis (Facts vs Inferences) */}
          {workflow.root_cause_analysis && (
            <div style={{ marginBottom: '20px', padding: '16px', borderRadius: '8px', border: '1px solid #E2E8F0', backgroundColor: '#FFFFFF' }}>
              <div style={{ display: 'flex', alignItems: 'center', gap: '8px', marginBottom: '12px' }}>
                <Search size={16} color="#2563EB" />
                <h4 style={{ margin: 0, fontSize: '14px', fontWeight: 600, color: '#0F172A' }}>
                  Root Cause Investigation (Confidence: {(workflow.root_cause_analysis.confidence * 100).toFixed(0)}%)
                </h4>
              </div>
              <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '16px' }}>
                <div>
                  <div style={{ fontSize: '12px', fontWeight: 600, color: '#059669', marginBottom: '4px' }}>Authoritative Facts</div>
                  <ul style={{ margin: 0, paddingLeft: '18px', fontSize: '12px', color: '#334155' }}>
                    {workflow.root_cause_analysis.confirmed_facts?.map((fact, idx) => (
                      <li key={idx} style={{ marginBottom: '2px' }}>{fact}</li>
                    ))}
                  </ul>
                </div>
                <div>
                  <div style={{ fontSize: '12px', fontWeight: 600, color: '#D97706', marginBottom: '4px' }}>Likely Causes (Inferences)</div>
                  <ul style={{ margin: 0, paddingLeft: '18px', fontSize: '12px', color: '#334155' }}>
                    {workflow.root_cause_analysis.likely_causes?.map((cause, idx) => (
                      <li key={idx} style={{ marginBottom: '2px' }}>{cause}</li>
                    ))}
                  </ul>
                </div>
              </div>
            </div>
          )}

          {/* Section 2: Cross-Module Impact Assessment */}
          {workflow.impact_assessment && (
            <div style={{ marginBottom: '20px', padding: '16px', borderRadius: '8px', border: '1px solid #E2E8F0', backgroundColor: '#FFFFFF' }}>
              <div style={{ display: 'flex', alignItems: 'center', gap: '8px', marginBottom: '12px' }}>
                <TrendingUp size={16} color="#7C3AED" />
                <h4 style={{ margin: 0, fontSize: '14px', fontWeight: 600, color: '#0F172A' }}>
                  Cross-Module Impact Assessment (Score: {workflow.impact_assessment.overall_impact_score}/100)
                </h4>
              </div>
              <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(180px, 1fr))', gap: '12px' }}>
                <div style={{ padding: '8px 12px', backgroundColor: '#F8FAFC', borderRadius: '6px' }}>
                  <div style={{ fontSize: '11px', color: '#64748B' }}>Operational Delay</div>
                  <div style={{ fontSize: '14px', fontWeight: 600, color: '#0F172A' }}>
                    +{workflow.impact_assessment.operational_delay_hours} Hours
                  </div>
                </div>
                <div style={{ padding: '8px 12px', backgroundColor: '#F8FAFC', borderRadius: '6px' }}>
                  <div style={{ fontSize: '11px', color: '#64748B' }}>Customer Impact</div>
                  <div style={{ fontSize: '14px', fontWeight: 600, color: '#D97706' }}>
                    {workflow.impact_assessment.customer_service_impact}
                  </div>
                </div>
                <div style={{ padding: '8px 12px', backgroundColor: '#F8FAFC', borderRadius: '6px' }}>
                  <div style={{ fontSize: '11px', color: '#64748B' }}>Cost Exposure</div>
                  <div style={{ fontSize: '14px', fontWeight: 600, color: '#DC2626' }}>
                    ${workflow.impact_assessment.estimated_cost_impact?.toFixed(2)}
                  </div>
                </div>
                <div style={{ padding: '8px 12px', backgroundColor: '#F8FAFC', borderRadius: '6px' }}>
                  <div style={{ fontSize: '11px', color: '#64748B' }}>SLA Breached</div>
                  <div style={{ fontSize: '14px', fontWeight: 600, color: workflow.impact_assessment.sla_breached ? '#DC2626' : '#16A34A' }}>
                    {workflow.impact_assessment.sla_breached ? 'YES' : 'NO'}
                  </div>
                </div>
                <div style={{ padding: '8px 12px', backgroundColor: '#F8FAFC', borderRadius: '6px' }}>
                  <div style={{ fontSize: '11px', color: '#64748B' }}>Regulatory Risk</div>
                  <div style={{ fontSize: '14px', fontWeight: 600, color: '#0F172A' }}>
                    {workflow.impact_assessment.regulatory_risk}
                  </div>
                </div>
              </div>
            </div>
          )}

          {/* Section 3: Recovery Options & Governance Selection */}
          {workflow.recovery_options?.length > 0 && (
            <div style={{ marginBottom: '20px', padding: '16px', borderRadius: '8px', border: '1px solid #E2E8F0', backgroundColor: '#FFFFFF' }}>
              <div style={{ display: 'flex', alignItems: 'center', gap: '8px', marginBottom: '12px' }}>
                <ShieldCheck size={16} color="#16A34A" />
                <h4 style={{ margin: 0, fontSize: '14px', fontWeight: 600, color: '#0F172A' }}>
                  Governed Recovery Options
                </h4>
              </div>
              <div style={{ display: 'flex', flexDirection: 'column', gap: '10px' }}>
                {workflow.recovery_options.map((opt) => (
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

          {/* Section 4: Authoritative Verification Result */}
          {workflow.last_verification_result && (
            <div style={{ marginBottom: '20px', padding: '12px 16px', borderRadius: '8px', backgroundColor: '#F0FDF4', border: '1px solid #BBF7D0' }}>
              <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                <CheckCircle size={16} color="#16A34A" />
                <span style={{ fontSize: '13px', fontWeight: 600, color: '#166534' }}>
                  Authoritative Verification Succeeded
                </span>
              </div>
              <div style={{ fontSize: '12px', color: '#15803D', marginTop: '4px' }}>
                {workflow.last_verification_result.verification_message} (Source: {workflow.last_verification_result.authoritative_source})
              </div>
            </div>
          )}

          {/* Section 5: Resolution Proof */}
          {workflow.resolution_proof && (
            <div style={{ marginBottom: '20px', padding: '12px 16px', borderRadius: '8px', backgroundColor: '#F8FAFC', border: '1px solid #CBD5E1' }}>
              <div style={{ fontSize: '11px', color: '#64748B', fontWeight: 600, textTransform: 'uppercase' }}>Authoritative Closure Proof</div>
              <div style={{ fontSize: '13px', color: '#1E293B', marginTop: '2px' }}>{workflow.resolution_proof}</div>
            </div>
          )}

          {/* Action Toolbar */}
          <div style={{ display: 'flex', flexWrap: 'wrap', gap: '8px', paddingTop: '12px', borderTop: '1px solid #F1F5F9' }}>
            {workflow.current_stage === 'DETECTED' && (
              <button
                onClick={handleInvestigate}
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
                Investigate Root Cause
              </button>
            )}

            {workflow.current_stage === 'INVESTIGATING' && (
              <button
                onClick={handleAssessImpact}
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
                Assess Downstream Impact
              </button>
            )}

            {workflow.current_stage === 'IMPACT_ASSESSMENT' && (
              <button
                onClick={handlePlanRecovery}
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
                Plan Recovery Options
              </button>
            )}

            {workflow.current_stage === 'PLANNING_RECOVERY' && (
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
                  onClick={handleExecuteAction}
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
                  Execute Action Boundary
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
                  Verify Action Execution
                </button>
              </>
            )}

            {workflow.current_stage === 'VERIFYING' && (
              <>
                <button
                  onClick={handleTransitionMonitoring}
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
                  Transition to Monitoring
                </button>
                <button
                  onClick={handleResolve}
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
                  Resolve with Authoritative Proof
                </button>
              </>
            )}

            {workflow.current_stage === 'MONITORING' && (
              <button
                onClick={handleResolve}
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
                Confirm Stability & Resolve
              </button>
            )}

            {workflow.current_stage === 'RESOLVED' && !workflow.outcome_recorded && (
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
                Record Learning into Workforce Memory
              </button>
            )}

            {/* Adaptive Replanning & Escalation available during active recovery */}
            {['PLANNING_RECOVERY', 'EXECUTING', 'VERIFYING', 'MONITORING'].includes(workflow.current_stage) && (
              <button
                onClick={handleReplan}
                disabled={actionLoading}
                style={{
                  padding: '7px 14px',
                  backgroundColor: '#FFFFFF',
                  color: '#475569',
                  border: '1px solid #CBD5E1',
                  borderRadius: '6px',
                  fontSize: '12px',
                  fontWeight: 500,
                  cursor: 'pointer'
                }}
              >
                Adaptive Replan
              </button>
            )}

            {workflow.current_stage !== 'CLOSED' && workflow.current_stage !== 'ESCALATED' && (
              <button
                onClick={handleEscalate}
                disabled={actionLoading}
                style={{
                  padding: '7px 14px',
                  backgroundColor: '#FFFFFF',
                  color: '#DC2626',
                  border: '1px solid #FECACA',
                  borderRadius: '6px',
                  fontSize: '12px',
                  fontWeight: 500,
                  cursor: 'pointer'
                }}
              >
                Escalate
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

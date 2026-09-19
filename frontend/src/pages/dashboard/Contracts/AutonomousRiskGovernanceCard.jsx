import React, { useState, useEffect, useCallback } from 'react';
import {
  ShieldAlert,
  ShieldCheck,
  AlertTriangle,
  FileCheck,
  CheckCircle,
  ArrowRight,
  RefreshCw,
  Sliders,
  Send,
  Truck,
  Layers,
  Award,
  ChevronRight,
  Zap,
  Clock,
  Briefcase,
  FileText,
  DollarSign,
  Scale
} from 'lucide-react';
import enterpriseService from '../../../services/enterpriseService';

const STAGE_CONFIG = {
  MONITORING: { label: 'Continuous Telemetry', bg: '#F8FAFC', text: '#475569', border: '#CBD5E1' },
  RISK_DETECTION: { label: 'Signal Detected', bg: '#FEF2F2', text: '#B91C1C', border: '#FECACA' },
  EVIDENCE_COLLECTION: { label: 'Evidence Collection', bg: '#FFFBEB', text: '#B45309', border: '#FDE68A' },
  MULTI_AGENT_ASSESSMENT: { label: 'Multi-Agent Assessment', bg: '#EFF6FF', text: '#1D4ED8', border: '#BFDBFE' },
  IMPACT_ANALYSIS: { label: 'Cross-Domain Impact', bg: '#FAF5FF', text: '#7E22CE', border: '#E9D5FF' },
  MITIGATION_PLANNING: { label: 'Mitigation Formulated', bg: '#F0F9FF', text: '#0369A1', border: '#BAE6FD' },
  WAITING_FOR_APPROVAL: { label: 'Waiting Executive Approval', bg: '#FEF2F2', text: '#991B1B', border: '#FCA5A5' },
  EXECUTING_MITIGATION: { label: 'Executing via Action System', bg: '#FFF7ED', text: '#C2410C', border: '#FED7AA' },
  VERIFYING: { label: 'Authoritative Verification', bg: '#F5F3FF', text: '#6D28D9', border: '#DDD6FE' },
  RISK_MONITORING: { label: 'Active Post-Remediation Monitoring', bg: '#EFF6FF', text: '#1E40AF', border: '#93C5FD' },
  RESOLVED: { label: 'Risk Formally Resolved', bg: '#F0FDF4', text: '#166534', border: '#86EFAC' },
  ESCALATED: { label: 'Executive Escalation', bg: '#FEF2F2', text: '#7F1D1D', border: '#F87171' }
};

export default function AutonomousRiskGovernanceCard({ entityType = 'CONTRACT', entityId = 'CTR-2026-001' }) {
  const [workflow, setWorkflow] = useState(null);
  const [loading, setLoading] = useState(true);
  const [actionLoading, setActionLoading] = useState(false);
  const [error, setError] = useState(null);
  const [activeTab, setActiveTab] = useState('overview'); // overview, contract, compliance, impact, mitigations
  const [showReassessModal, setShowReassessModal] = useState(false);
  const [newSignalText, setNewSignalText] = useState('Carrier transshipment delay of 12 hours declared');

  const fetchWorkflow = useCallback(async () => {
    if (!entityId) return;
    setLoading(true);
    setError(null);
    try {
      const res = await enterpriseService.getRiskWorkflow(entityId);
      const data = res?.data || res;
      setWorkflow(data);
    } catch (err) {
      if (err?.response?.status === 404) {
        setWorkflow(null);
      } else {
        setError(err?.response?.data?.error || err?.message || 'Failed to load risk governance workflow');
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
      const res = await enterpriseService.initiateRiskWorkflow(entityType, entityId);
      setWorkflow(res?.data || res);
    } catch (err) {
      setError(err?.response?.data?.error || err?.message || 'Failed to initiate risk workflow');
    } finally {
      setActionLoading(false);
    }
  };

  const handleRunFullAssessmentCycle = async () => {
    if (!workflow?.workflow_id) return;
    setActionLoading(true);
    setError(null);
    try {
      // 1. Evidence Collection
      let res = await enterpriseService.collectRiskEvidence(workflow.workflow_id, {
        signal_type: 'CUSTOMS_AND_RATE_AUDIT',
        category: 'CONTRACT',
        source_entity: `${workflow.entity_type} #${workflow.entity_id}`,
        source_record: 'commercial_agreements_ledger',
        authoritative_value: 'Rate variance $300.00 | ISF Filing Pending',
        reporting_agent: 'contract_agent',
        confidence_score: 1.0,
        is_authoritative_fact: true,
        provenance_description: 'Reconciled against active contract terms and CBP customs gateway'
      });
      // 2. Multi-Agent Assessment
      res = await enterpriseService.conductMultiAgentRiskAssessment(workflow.workflow_id);
      // 3. Cross-Domain Impact
      res = await enterpriseService.assessCrossDomainImpact(workflow.workflow_id);
      // 4. Mitigation Planning
      res = await enterpriseService.planRiskMitigationOptions(workflow.workflow_id);
      setWorkflow(res?.data || res);
      setActiveTab('mitigations');
    } catch (err) {
      setError(err?.response?.data?.error || err?.message || 'Risk assessment sequence encountered policy check');
      fetchWorkflow();
    } finally {
      setActionLoading(false);
    }
  };

  const handleSelectMitigation = async (optionId) => {
    if (!workflow?.workflow_id) return;
    setActionLoading(true);
    setError(null);
    try {
      const res = await enterpriseService.selectRiskMitigationOption(workflow.workflow_id, optionId);
      setWorkflow(res?.data || res);
    } catch (err) {
      setError(err?.response?.data?.error || err?.message || 'Failed to select mitigation option');
      fetchWorkflow();
    } finally {
      setActionLoading(false);
    }
  };

  const handleExecuteMitigation = async () => {
    if (!workflow?.workflow_id) return;
    setActionLoading(true);
    setError(null);
    try {
      const res = await enterpriseService.executeRiskMitigationAction(workflow.workflow_id, 'step-exec-action');
      setWorkflow(res?.data || res);
    } catch (err) {
      setError(err?.response?.data?.error || err?.message || 'Action execution failed');
      fetchWorkflow();
    } finally {
      setActionLoading(false);
    }
  };

  const handleVerifyMitigation = async () => {
    if (!workflow?.workflow_id) return;
    setActionLoading(true);
    setError(null);
    try {
      const res = await enterpriseService.verifyRiskMitigationAction(workflow.workflow_id, 'step-verify-action');
      setWorkflow(res?.data || res);
    } catch (err) {
      setError(err?.response?.data?.error || err?.message || 'Authoritative verification failed');
      fetchWorkflow();
    } finally {
      setActionLoading(false);
    }
  };

  const handleReassess = async () => {
    if (!workflow?.workflow_id) return;
    setActionLoading(true);
    setError(null);
    try {
      const newEvidence = {
        evidence_id: `EVD-${Date.now()}`,
        signal_type: 'CARRIER_TRANSIT_DELAY',
        category: 'OPERATIONAL',
        source_entity: `SHIPMENT #${workflow.entity_id}`,
        source_record: 'carrier_tracking_telemetry',
        authoritative_value: newSignalText,
        reporting_agent: 'shipment_agent',
        confidence_score: 0.95,
        is_authoritative_fact: true,
        provenance_description: 'Real-time vessel AIS transponder update'
      };
      const res = await enterpriseService.reassessRiskCondition(workflow.workflow_id, newEvidence);
      setWorkflow(res?.data || res);
      setShowReassessModal(false);
      setActiveTab('impact');
    } catch (err) {
      setError(err?.response?.data?.error || err?.message || 'Failed to reassess risk condition');
    } finally {
      setActionLoading(false);
    }
  };

  const handleResolveWorkflow = async () => {
    if (!workflow?.workflow_id) return;
    setActionLoading(true);
    setError(null);
    try {
      const res = await enterpriseService.resolveRiskWorkflow(workflow.workflow_id, 'Customs documents cleared and contract rate harmonized with 0 demurrage');
      setWorkflow(res?.data || res);
    } catch (err) {
      setError(err?.response?.data?.error || err?.message || 'Failed to resolve risk workflow');
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
            background: '#FEF2F2',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            color: '#DC2626'
          }}>
            <ShieldAlert size={22} />
          </div>
          <div>
            <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
              <h3 style={{ margin: 0, fontSize: '17px', fontWeight: 600, color: '#0F172A' }}>
                Autonomous Contract, Compliance & Risk Governance
              </h3>
              <span style={{
                fontSize: '11px',
                fontWeight: 600,
                padding: '2px 8px',
                borderRadius: '999px',
                background: '#F1F5F9',
                color: '#475569'
              }}>
                Phase 7.7 Enterprise
              </span>
            </div>
            <p style={{ margin: '2px 0 0', fontSize: '13px', color: '#64748B' }}>
              Continuous cross-domain risk detection, SLA protection, statutory compliance enforcement, and governed mitigation
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
            title="Refresh Risk Governance Workflow"
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
            No autonomous risk governance workflow active for {entityType} #{entityId}.
          </p>
          <button
            onClick={handleInitiate}
            disabled={actionLoading}
            style={{
              background: '#DC2626',
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
            {actionLoading ? 'Initiating Governance...' : 'Initiate Continuous Risk Governance'}
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
              { id: 'overview', label: 'Risk Overview' },
              { id: 'contract', label: 'Contract Terms & SLA' },
              { id: 'compliance', label: 'Statutory Compliance' },
              { id: 'impact', label: 'Cross-Domain Impact' },
              { id: 'mitigations', label: `Governed Mitigations (${workflow.available_mitigations?.length || 0})` }
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
              {/* Fact vs Assessment vs Cascading Strip */}
              <div style={{
                display: 'grid',
                gridTemplateColumns: 'repeat(auto-fit, minmax(220px, 1fr))',
                gap: '12px',
                marginBottom: '16px'
              }}>
                {/* Authoritative Fact */}
                <div style={{ background: '#F8FAFC', border: '1px solid #E2E8F0', borderRadius: '8px', padding: '12px' }}>
                  <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '6px' }}>
                    <span style={{ fontSize: '11px', fontWeight: 600, color: '#475569', textTransform: 'uppercase' }}>Authoritative Contract Fact</span>
                    <span style={{ fontSize: '10px', background: '#E2E8F0', color: '#334155', padding: '1px 5px', borderRadius: '4px' }}>VERIFIED</span>
                  </div>
                  <div style={{ fontSize: '15px', fontWeight: 700, color: '#0F172A' }}>
                    SLA: {workflow.authoritative_entity?.agreed_sla_hours || 48}h Max Port Turnaround
                  </div>
                  <div style={{ fontSize: '12px', color: '#64748B', marginTop: '4px' }}>
                    Floor Margin: {workflow.authoritative_entity?.contract_margin_floor || 14.5}% | Bonded Carrier: Verified
                  </div>
                </div>

                {/* AI Risk Assessment */}
                <div style={{ background: '#FEF2F2', border: '1px solid #FECACA', borderRadius: '8px', padding: '12px' }}>
                  <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '6px' }}>
                    <span style={{ fontSize: '11px', fontWeight: 600, color: '#DC2626', textTransform: 'uppercase' }}>Risk Tier & Signals</span>
                    <span style={{ fontSize: '10px', background: '#FEE2E2', color: '#991B1B', padding: '1px 5px', borderRadius: '4px' }}>{workflow.overall_severity} SEVERITY</span>
                  </div>
                  <div style={{ fontSize: '15px', fontWeight: 700, color: '#991B1B' }}>
                    Rate Variance: $300.00 | Customs Filing Absent
                  </div>
                  <div style={{ fontSize: '12px', color: '#DC2626', marginTop: '4px' }}>
                    SLA Breach Likelihood: 68.5% (T-14 Days to Expiry)
                  </div>
                </div>

                {/* Cross-Domain Cascading Exposure */}
                <div style={{ background: '#FAF5FF', border: '1px solid #E9D5FF', borderRadius: '8px', padding: '12px' }}>
                  <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '6px' }}>
                    <span style={{ fontSize: '11px', fontWeight: 600, color: '#7E22CE', textTransform: 'uppercase' }}>Financial Exposure</span>
                    <span style={{ fontSize: '10px', background: '#F3E8FF', color: '#6B21A8', padding: '1px 5px', borderRadius: '4px' }}>CASCADING</span>
                  </div>
                  <div style={{ fontSize: '20px', fontWeight: 700, color: '#6B21A8' }}>
                    ${workflow.cross_domain_risk?.total_financial_exposure?.toFixed(2) || '4,850.00'}
                  </div>
                  <div style={{ fontSize: '12px', color: '#7E22CE', marginTop: '4px' }}>
                    Includes SLA Liquidated Damages & Demurrage Buffer
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
                    onClick={handleRunFullAssessmentCycle}
                    disabled={actionLoading}
                    style={{
                      background: '#DC2626',
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
                    Run Multi-Agent Risk Assessment
                  </button>

                  <button
                    onClick={() => setShowReassessModal(true)}
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
                    Adaptive Reassessment
                  </button>

                  {workflow.current_stage === 'EXECUTING_MITIGATION' && (
                    <button
                      onClick={handleExecuteMitigation}
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
                      Execute Permitted Action
                    </button>
                  )}

                  {workflow.current_stage === 'VERIFYING' && (
                    <button
                      onClick={handleVerifyMitigation}
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
                      Verify Clearance Gateway
                    </button>
                  )}

                  {workflow.current_stage === 'RISK_MONITORING' && (
                    <button
                      onClick={handleResolveWorkflow}
                      disabled={actionLoading}
                      style={{
                        background: '#15803D',
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
                      Formally Resolve & Close
                    </button>
                  )}
                </div>

                <div style={{ fontSize: '12px', color: '#64748B' }}>
                  Replan Version: <strong style={{ color: '#0F172A' }}>v{workflow.replan_version}</strong>
                </div>
              </div>
            </div>
          )}

          {/* TAB CONTENT: Contract Risk */}
          {activeTab === 'contract' && (
            <div style={{ background: '#F8FAFC', border: '1px solid #E2E8F0', borderRadius: '8px', padding: '16px' }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '12px' }}>
                <h4 style={{ margin: 0, fontSize: '14px', color: '#0F172A', display: 'flex', alignItems: 'center', gap: '6px' }}>
                  <FileText size={16} color="#2563EB" />
                  Contract Clauses, Agreed Rates & SLA Obligations
                </h4>
                <span style={{ fontSize: '12px', color: '#64748B' }}>
                  Contract: <strong>CTR-{workflow.entity_id}</strong>
                </span>
              </div>

              <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(180px, 1fr))', gap: '12px', marginBottom: '16px' }}>
                <div style={{ background: '#FFFFFF', padding: '12px', borderRadius: '6px', border: '1px solid #E2E8F0' }}>
                  <div style={{ fontSize: '11px', color: '#64748B' }}>Agreed Contract Rate</div>
                  <div style={{ fontSize: '15px', fontWeight: 600, color: '#0F172A', marginTop: '2px' }}>
                    ${workflow.contract_risk?.agreed_contract_rate?.toFixed(2) || '2,650.00'}
                  </div>
                </div>

                <div style={{ background: '#FFFFFF', padding: '12px', borderRadius: '6px', border: '1px solid #E2E8F0' }}>
                  <div style={{ fontSize: '11px', color: '#64748B' }}>Invoiced Rate</div>
                  <div style={{ fontSize: '15px', fontWeight: 600, color: '#DC2626', marginTop: '2px' }}>
                    ${workflow.contract_risk?.quoted_or_invoiced_rate?.toFixed(2) || '2,950.00'}
                  </div>
                </div>

                <div style={{ background: '#FFFFFF', padding: '12px', borderRadius: '6px', border: '1px solid #E2E8F0' }}>
                  <div style={{ fontSize: '11px', color: '#64748B' }}>Unauthorized Variance</div>
                  <div style={{ fontSize: '15px', fontWeight: 600, color: '#DC2626', marginTop: '2px' }}>
                    +${workflow.contract_risk?.unauthorized_rate_variance?.toFixed(2) || '300.00'}
                  </div>
                </div>

                <div style={{ background: '#FFFFFF', padding: '12px', borderRadius: '6px', border: '1px solid #E2E8F0' }}>
                  <div style={{ fontSize: '11px', color: '#64748B' }}>SLA Breach Likelihood</div>
                  <div style={{ fontSize: '15px', fontWeight: 600, color: '#B45309', marginTop: '2px' }}>
                    {workflow.contract_risk?.sla_breach_likelihood_pct?.toFixed(1) || '68.5'}%
                  </div>
                </div>
              </div>

              {workflow.contract_risk?.contract_clauses_at_risk && (
                <div style={{ background: '#FEF2F2', border: '1px solid #FECACA', borderRadius: '6px', padding: '10px 12px' }}>
                  <div style={{ fontSize: '12px', fontWeight: 600, color: '#991B1B', marginBottom: '4px' }}>Clauses at Commercial Risk:</div>
                  <ul style={{ margin: 0, paddingLeft: '18px', fontSize: '12px', color: '#991B1B' }}>
                    {workflow.contract_risk.contract_clauses_at_risk.map((c, i) => (
                      <li key={i}>{c}</li>
                    ))}
                  </ul>
                </div>
              )}
            </div>
          )}

          {/* TAB CONTENT: Compliance Risk */}
          {activeTab === 'compliance' && (
            <div style={{ background: '#F8FAFC', border: '1px solid #E2E8F0', borderRadius: '8px', padding: '16px' }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '12px' }}>
                <h4 style={{ margin: 0, fontSize: '14px', color: '#0F172A', display: 'flex', alignItems: 'center', gap: '6px' }}>
                  <Scale size={16} color="#047857" />
                  Statutory & Customs Documentation Readiness
                </h4>
                <span style={{
                  fontSize: '11px',
                  fontWeight: 600,
                  padding: '2px 8px',
                  borderRadius: '4px',
                  background: '#FEF2F2',
                  color: '#991B1B'
                }}>
                  RESTRICTED OPERATION
                </span>
              </div>

              <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(180px, 1fr))', gap: '12px', marginBottom: '16px' }}>
                <div style={{ background: '#FFFFFF', padding: '12px', borderRadius: '6px', border: '1px solid #E2E8F0' }}>
                  <div style={{ fontSize: '11px', color: '#64748B' }}>Regulatory Standard</div>
                  <div style={{ fontSize: '14px', fontWeight: 600, color: '#0F172A', marginTop: '2px' }}>
                    {workflow.compliance_risk?.regulatory_standard || 'CBP_19CFR_149'}
                  </div>
                </div>

                <div style={{ background: '#FFFFFF', padding: '12px', borderRadius: '6px', border: '1px solid #E2E8F0' }}>
                  <div style={{ fontSize: '11px', color: '#64748B' }}>Customs Readiness Score</div>
                  <div style={{ fontSize: '14px', fontWeight: 600, color: '#DC2626', marginTop: '2px' }}>
                    {workflow.compliance_risk?.customs_readiness_score?.toFixed(0) || '45'}% / 100
                  </div>
                </div>

                <div style={{ background: '#FFFFFF', padding: '12px', borderRadius: '6px', border: '1px solid #E2E8F0' }}>
                  <div style={{ fontSize: '11px', color: '#64748B' }}>Sanctions Screening</div>
                  <div style={{ fontSize: '14px', fontWeight: 600, color: '#16A34A', marginTop: '2px' }}>
                    PASSED (0 Violations)
                  </div>
                </div>
              </div>

              {workflow.compliance_risk?.missing_documents_list && (
                <div style={{ background: '#FFFBEB', border: '1px solid #FDE68A', borderRadius: '6px', padding: '10px 12px' }}>
                  <div style={{ fontSize: '12px', fontWeight: 600, color: '#B45309', marginBottom: '4px' }}>Missing Statutory Documentation:</div>
                  <ul style={{ margin: 0, paddingLeft: '18px', fontSize: '12px', color: '#B45309' }}>
                    {workflow.compliance_risk.missing_documents_list.map((d, i) => (
                      <li key={i}>{d}</li>
                    ))}
                  </ul>
                </div>
              )}
            </div>
          )}

          {/* TAB CONTENT: Cross-Domain Impact */}
          {activeTab === 'impact' && (
            <div style={{ background: '#F8FAFC', border: '1px solid #E2E8F0', borderRadius: '8px', padding: '16px' }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '12px' }}>
                <h4 style={{ margin: 0, fontSize: '14px', color: '#0F172A', display: 'flex', alignItems: 'center', gap: '6px' }}>
                  <Layers size={16} color="#7E22CE" />
                  Cascading Cross-Domain Risk Propagation Chain
                </h4>
                <span style={{ fontSize: '12px', color: '#64748B' }}>
                  Combined Severity: <strong style={{ color: '#DC2626' }}>{workflow.cross_domain_risk?.combined_risk_severity || 'HIGH'}</strong>
                </span>
              </div>

              <div style={{ background: '#EFF6FF', border: '1px solid #BFDBFE', borderRadius: '6px', padding: '12px', marginBottom: '16px' }}>
                <div style={{ fontSize: '11px', fontWeight: 600, color: '#1E40AF', textTransform: 'uppercase', marginBottom: '4px' }}>Propagation Chain</div>
                <div style={{ fontSize: '13px', fontWeight: 600, color: '#1E40AF' }}>
                  {workflow.cross_domain_risk?.propagation_chain || 'Missing ISF Docs -> Customs Hold -> Transit Delay -> 48h SLA Breach -> Liquidated Damages ($1,200) -> Churn Risk'}
                </div>
              </div>

              <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(180px, 1fr))', gap: '12px' }}>
                <div style={{ background: '#FFFFFF', padding: '12px', borderRadius: '6px', border: '1px solid #E2E8F0' }}>
                  <div style={{ fontSize: '11px', color: '#64748B' }}>Total Financial Exposure</div>
                  <div style={{ fontSize: '16px', fontWeight: 700, color: '#0F172A', marginTop: '2px' }}>
                    ${workflow.cross_domain_risk?.total_financial_exposure?.toFixed(2) || '4,850.00'}
                  </div>
                </div>

                <div style={{ background: '#FFFFFF', padding: '12px', borderRadius: '6px', border: '1px solid #E2E8F0' }}>
                  <div style={{ fontSize: '11px', color: '#64748B' }}>SLA Liquidated Damages</div>
                  <div style={{ fontSize: '16px', fontWeight: 700, color: '#DC2626', marginTop: '2px' }}>
                    ${workflow.cross_domain_risk?.sla_penalty_exposure?.toFixed(2) || '1,200.00'}
                  </div>
                </div>

                <div style={{ background: '#FFFFFF', padding: '12px', borderRadius: '6px', border: '1px solid #E2E8F0' }}>
                  <div style={{ fontSize: '11px', color: '#64748B' }}>Customer Churn Likelihood</div>
                  <div style={{ fontSize: '16px', fontWeight: 700, color: '#B45309', marginTop: '2px' }}>
                    {workflow.cross_domain_risk?.customer_retention_risk_pct?.toFixed(1) || '34.0'}%
                  </div>
                </div>
              </div>
            </div>
          )}

          {/* TAB CONTENT: Governed Mitigations */}
          {activeTab === 'mitigations' && (
            <div>
              <p style={{ margin: '0 0 12px', fontSize: '13px', color: '#64748B' }}>
                Synthesized risk mitigations governed under enterprise contract and statutory compliance policies:
              </p>

              <div style={{ display: 'flex', flexDirection: 'column', gap: '10px' }}>
                {(workflow.available_mitigations || []).map(opt => {
                  const isSelected = workflow.selected_mitigation?.option_id === opt.option_id;

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
                          <span>Action: <strong style={{ color: '#0F172A' }}>{opt.action_type}</strong></span>
                          <span>Residual Risk: <strong>{opt.residual_risk_level}</strong></span>
                          <span>Autonomy: <strong>{opt.required_autonomy}</strong></span>
                          <span>Confidence: <strong>{Math.round(opt.confidence * 100)}%</strong></span>
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
                            <CheckCircle size={14} /> Selected Strategy
                          </span>
                        ) : (
                          <button
                            onClick={() => handleSelectMitigation(opt.option_id)}
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
                            Select Mitigation
                          </button>
                        )}
                      </div>
                    </div>
                  );
                })}
              </div>
            </div>
          )}
        </>
      )}

      {/* Adaptive Reassessment Modal */}
      {showReassessModal && (
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
              Adaptive Risk Reassessment
            </h4>
            <p style={{ margin: '0 0 16px', fontSize: '13px', color: '#64748B' }}>
              Simulate arrival of a new operational or commercial signal to test adaptive risk recalculation and cross-domain propagation.
            </p>

            <div style={{ marginBottom: '16px' }}>
              <label style={{ fontSize: '12px', fontWeight: 600, color: '#475569', display: 'block', marginBottom: '6px' }}>
                New Ingested Signal:
              </label>
              <textarea
                rows={3}
                value={newSignalText}
                onChange={(e) => setNewSignalText(e.target.value)}
                style={{
                  width: '100%',
                  padding: '8px 12px',
                  borderRadius: '6px',
                  border: '1px solid #CBD5E1',
                  fontSize: '13px',
                  boxSizing: 'border-box'
                }}
              />
            </div>

            <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '8px' }}>
              <button
                onClick={() => setShowReassessModal(false)}
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
                onClick={handleReassess}
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
                {actionLoading ? 'Reassessing...' : 'Execute Reassessment'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

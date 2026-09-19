import React, { useState, useEffect, useCallback } from 'react';
import {
  Compass,
  Cpu,
  ShieldCheck,
  AlertTriangle,
  Clock,
  RotateCw,
  CheckCircle,
  FileCheck,
  ArrowRight,
  UserCheck,
  AlertCircle,
  RefreshCw
} from 'lucide-react';
import { enterpriseService } from '../../../../services/enterpriseService';

const STAGE_COLORS = {
  SHIPMENT_CREATED: { bg: '#F1F5F9', text: '#475569', border: '#CBD5E1' },
  PLANNING: { bg: '#EFF6FF', text: '#1D4ED8', border: '#BFDBFE' },
  BOOKED: { bg: '#F0FDF4', text: '#15803D', border: '#BBF7D0' },
  IN_TRANSIT: { bg: '#F0F9FF', text: '#0369A1', border: '#BAE6FD' },
  MONITORING: { bg: '#FAF5FF', text: '#7E22CE', border: '#E9D5FF' },
  DELIVERED: { bg: '#ECFDF5', text: '#047857', border: '#A7F3D0' },
  POST_DELIVERY: { bg: '#FFFBEB', text: '#B45309', border: '#FDE68A' },
  COMPLETED: { bg: '#F8FAFC', text: '#334155', border: '#E2E8F0' }
};

export default function AutonomousShipmentLifecycleCard({ shipmentId }) {
  const [lifecycle, setLifecycle] = useState(null);
  const [loading, setLoading] = useState(true);
  const [actionLoading, setActionLoading] = useState(false);
  const [error, setError] = useState(null);

  const fetchLifecycle = useCallback(async () => {
    if (!shipmentId) return;
    setLoading(true);
    setError(null);
    try {
      const res = await enterpriseService.getShipmentLifecycle(shipmentId);
      const payload = res?.data || res;
      setLifecycle(payload);
    } catch (err) {
      // If not yet initiated, attempt initial retrieval or provide initiate option
      if (err?.response?.status === 404) {
        setLifecycle(null);
      } else {
        setError(err?.response?.data?.error || err?.message || 'Failed to load autonomous lifecycle');
      }
    } finally {
      setLoading(false);
    }
  }, [shipmentId]);

  useEffect(() => {
    fetchLifecycle();
  }, [fetchLifecycle]);

  const handleInitiate = async () => {
    setActionLoading(true);
    try {
      const res = await enterpriseService.initiateShipmentLifecycle(shipmentId);
      setLifecycle(res?.data || res);
    } catch (err) {
      setError(err?.response?.data?.error || err?.message || 'Failed to initiate autonomous lifecycle');
    } finally {
      setActionLoading(false);
    }
  };

  const handleEvaluateETA = async () => {
    setActionLoading(true);
    try {
      const res = await enterpriseService.evaluateShipmentETA(shipmentId);
      setLifecycle(res?.data || res);
    } catch (err) {
      setError(err?.response?.data?.error || err?.message || 'ETA evaluation failed');
    } finally {
      setActionLoading(false);
    }
  };

  const handleReplan = async () => {
    const reason = window.prompt('Enter operational replanning reason:', 'Operational buffer adjustment');
    if (!reason) return;
    setActionLoading(true);
    try {
      const res = await enterpriseService.replanShipmentLifecycle(shipmentId, reason);
      setLifecycle(res?.data || res);
    } catch (err) {
      setError(err?.response?.data?.error || err?.message || 'Adaptive replan failed');
    } finally {
      setActionLoading(false);
    }
  };

  const handlePostDeliveryAudit = async () => {
    setActionLoading(true);
    try {
      const res = await enterpriseService.auditPostDelivery(shipmentId);
      setLifecycle(res?.data || res);
    } catch (err) {
      setError(err?.response?.data?.error || err?.message || 'Post-delivery audit failed');
    } finally {
      setActionLoading(false);
    }
  };

  if (loading) {
    return (
      <div style={{
        padding: '16px',
        backgroundColor: '#FFFFFF',
        border: '1px solid #E2E8F0',
        borderRadius: '8px',
        display: 'flex',
        alignItems: 'center',
        gap: '10px',
        color: '#64748B',
        fontSize: '13px'
      }}>
        <RotateCw size={16} className="animate-spin text-blue-600" />
        <span>Synchronizing autonomous shipment lifecycle state...</span>
      </div>
    );
  }

  if (!lifecycle) {
    return (
      <div style={{
        padding: '20px',
        backgroundColor: '#F8FAFC',
        border: '1px dashed #CBD5E1',
        borderRadius: '8px',
        display: 'flex',
        justifyContent: 'space-between',
        alignItems: 'center'
      }}>
        <div>
          <h4 style={{ margin: 0, fontSize: '14px', fontWeight: 600, color: '#1E293B', display: 'flex', alignItems: 'center', gap: '8px' }}>
            <Compass size={18} color="#2563EB" />
            Autonomous Shipment Lifecycle (Phase 7.2)
          </h4>
          <p style={{ margin: '4px 0 0', fontSize: '12px', color: '#64748B' }}>
            Continuous multi-agent orchestration is ready to be initiated for this shipment.
          </p>
        </div>
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
            display: 'flex',
            alignItems: 'center',
            gap: '6px'
          }}
        >
          {actionLoading ? <RotateCw size={14} className="animate-spin" /> : <Cpu size={14} />}
          Initiate Autonomous Lifecycle
        </button>
      </div>
    );
  }

  const stageTheme = STAGE_COLORS[lifecycle.current_stage] || STAGE_COLORS.PLANNING;

  return (
    <div style={{
      backgroundColor: '#FFFFFF',
      border: '1px solid #E2E8F0',
      borderRadius: '8px',
      padding: '20px',
      boxShadow: '0 1px 3px rgba(0,0,0,0.04)',
      fontFamily: 'inherit'
    }}>
      {/* Header */}
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: '16px' }}>
        <div>
          <div style={{ display: 'flex', alignItems: 'center', gap: '10px' }}>
            <span style={{
              display: 'inline-flex',
              alignItems: 'center',
              justifyContent: 'center',
              width: '30px',
              height: '30px',
              borderRadius: '6px',
              backgroundColor: '#EFF6FF',
              color: '#2563EB'
            }}>
              <Compass size={18} />
            </span>
            <div>
              <h3 style={{ margin: 0, fontSize: '15px', fontWeight: 600, color: '#0F172A' }}>
                Autonomous Shipment Lifecycle
              </h3>
              <span style={{ fontSize: '11px', color: '#64748B' }}>
                Governed Multi-Agent Coordination & Durable State Machine
              </span>
            </div>
          </div>
        </div>

        {/* Badges */}
        <div style={{ display: 'flex', gap: '8px', alignItems: 'center' }}>
          {/* Authoritative State Separation */}
          <span style={{
            fontSize: '11px',
            padding: '4px 8px',
            borderRadius: '4px',
            backgroundColor: '#F1F5F9',
            color: '#475569',
            border: '1px solid #E2E8F0',
            fontWeight: 500
          }}>
            Authoritative: <strong>{lifecycle.authoritative_status}</strong>
          </span>

          {/* Autonomous Lifecycle Stage */}
          <span style={{
            fontSize: '11px',
            padding: '4px 10px',
            borderRadius: '4px',
            backgroundColor: stageTheme.bg,
            color: stageTheme.text,
            border: `1px solid ${stageTheme.border}`,
            fontWeight: 600,
            textTransform: 'uppercase'
          }}>
            AI Stage: {lifecycle.current_stage?.replace(/_/g, ' ')}
          </span>

          <span style={{
            fontSize: '11px',
            padding: '4px 8px',
            borderRadius: '4px',
            backgroundColor: lifecycle.workflow_state === 'WAITING_FOR_APPROVAL' ? '#FEF2F2' : '#F8FAFC',
            color: lifecycle.workflow_state === 'WAITING_FOR_APPROVAL' ? '#DC2626' : '#64748B',
            border: '1px solid #E2E8F0',
            fontWeight: 500
          }}>
            State: {lifecycle.workflow_state}
          </span>
        </div>
      </div>

      {error && (
        <div style={{
          padding: '10px 14px',
          backgroundColor: '#FEF2F2',
          border: '1px solid #FCA5A5',
          borderRadius: '6px',
          color: '#B91C1C',
          fontSize: '12px',
          marginBottom: '14px',
          display: 'flex',
          alignItems: 'center',
          gap: '8px'
        }}>
          <AlertCircle size={14} />
          <span>{error}</span>
        </div>
      )}

      {/* Grid: Metrics, Intelligence, Specialists, Actions */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(220px, 1fr))', gap: '16px', marginBottom: '16px' }}>
        {/* Metric 1: ETA Intelligence */}
        <div style={{ padding: '12px', backgroundColor: '#F8FAFC', borderRadius: '6px', border: '1px solid #E2E8F0' }}>
          <div style={{ fontSize: '11px', color: '#64748B', display: 'flex', alignItems: 'center', gap: '5px', marginBottom: '4px' }}>
            <Clock size={13} color="#2563EB" />
            <span>ETA Intelligence</span>
          </div>
          <div style={{ fontSize: '14px', fontWeight: 600, color: '#0F172A' }}>
            {lifecycle.predicted_delay_hours > 0 ? (
              <span style={{ color: lifecycle.is_delay_operationally_significant ? '#DC2626' : '#D97706' }}>
                +{lifecycle.predicted_delay_hours.toFixed(1)}h Delay
              </span>
            ) : (
              <span style={{ color: '#16A34A' }}>On Schedule</span>
            )}
          </div>
          <div style={{ fontSize: '11px', color: '#64748B', marginTop: '2px' }}>
            Confidence: {(lifecycle.eta_confidence * 100).toFixed(0)}%
            {lifecycle.is_delay_operationally_significant && (
              <span style={{ color: '#DC2626', marginLeft: '6px', fontWeight: 500 }}>Operational Risk</span>
            )}
          </div>
        </div>

        {/* Metric 2: Autonomous Exceptions */}
        <div style={{ padding: '12px', backgroundColor: '#F8FAFC', borderRadius: '6px', border: '1px solid #E2E8F0' }}>
          <div style={{ fontSize: '11px', color: '#64748B', display: 'flex', alignItems: 'center', gap: '5px', marginBottom: '4px' }}>
            <AlertTriangle size={13} color="#D97706" />
            <span>Operational Exceptions</span>
          </div>
          <div style={{ fontSize: '14px', fontWeight: 600, color: lifecycle.active_exceptions_count > 0 ? '#DC2626' : '#16A34A' }}>
            {lifecycle.active_exceptions_count > 0 ? `${lifecycle.active_exceptions_count} Active (${lifecycle.exception_severity || 'HIGH'})` : 'None Detected'}
          </div>
          <div style={{ fontSize: '11px', color: '#64748B', marginTop: '2px' }}>
            {lifecycle.exception_root_cause ? `Cause: ${lifecycle.exception_root_cause}` : 'Telemetry nominal'}
          </div>
        </div>

        {/* Metric 3: Action Execution & Verification */}
        <div style={{ padding: '12px', backgroundColor: '#F8FAFC', borderRadius: '6px', border: '1px solid #E2E8F0' }}>
          <div style={{ fontSize: '11px', color: '#64748B', display: 'flex', alignItems: 'center', gap: '5px', marginBottom: '4px' }}>
            <ShieldCheck size={13} color="#16A34A" />
            <span>Action Verification</span>
          </div>
          <div style={{ fontSize: '14px', fontWeight: 600, color: '#0F172A' }}>
            {lifecycle.last_action_verified ? (
              <span style={{ color: '#16A34A', display: 'flex', alignItems: 'center', gap: '4px' }}>
                <CheckCircle size={14} /> Verified Grounded
              </span>
            ) : (
              <span style={{ color: '#64748B' }}>Awaiting Execution</span>
            )}
          </div>
          <div style={{ fontSize: '11px', color: '#64748B', marginTop: '2px' }}>
            Last: {lifecycle.last_executed_action || 'None'}
          </div>
        </div>

        {/* Metric 4: Adaptive Replan Version */}
        <div style={{ padding: '12px', backgroundColor: '#F8FAFC', borderRadius: '6px', border: '1px solid #E2E8F0' }}>
          <div style={{ fontSize: '11px', color: '#64748B', display: 'flex', alignItems: 'center', gap: '5px', marginBottom: '4px' }}>
            <RefreshCw size={13} color="#6366F1" />
            <span>Replan Version</span>
          </div>
          <div style={{ fontSize: '14px', fontWeight: 600, color: '#0F172A' }}>
            v{lifecycle.replan_version || 1}
          </div>
          <div style={{ fontSize: '11px', color: '#64748B', marginTop: '2px' }}>
            {lifecycle.pending_approvals_count > 0 ? (
              <span style={{ color: '#DC2626', fontWeight: 600 }}>{lifecycle.pending_approvals_count} Pending Approval</span>
            ) : (
              'Governed Continuous Plan'
            )}
          </div>
        </div>
      </div>

      {/* Specialist Agents Involved */}
      {lifecycle.assigned_specialists && lifecycle.assigned_specialists.length > 0 && (
        <div style={{ marginBottom: '14px' }}>
          <div style={{ fontSize: '11px', fontWeight: 600, color: '#475569', textTransform: 'uppercase', letterSpacing: '0.04em', marginBottom: '6px' }}>
            Multi-Agent Workforce Assigned
          </div>
          <div style={{ display: 'flex', flexWrap: 'wrap', gap: '6px' }}>
            {lifecycle.assigned_specialists.map((agent) => (
              <span key={agent} style={{
                fontSize: '11px',
                padding: '2px 8px',
                borderRadius: '12px',
                backgroundColor: '#F1F5F9',
                color: '#334155',
                border: '1px solid #E2E8F0',
                display: 'inline-flex',
                alignItems: 'center',
                gap: '4px'
              }}>
                <Cpu size={10} color="#64748B" />
                {agent.replace('_agent', '').toUpperCase()} AGENT
              </span>
            ))}
          </div>
        </div>
      )}

      {/* Action Controls */}
      <div style={{
        display: 'flex',
        flexWrap: 'wrap',
        gap: '8px',
        borderTop: '1px solid #F1F5F9',
        paddingTop: '14px',
        justifyContent: 'flex-end'
      }}>
        <button
          onClick={handleEvaluateETA}
          disabled={actionLoading}
          style={{
            padding: '6px 12px',
            backgroundColor: '#FFFFFF',
            color: '#334155',
            border: '1px solid #CBD5E1',
            borderRadius: '6px',
            fontSize: '12px',
            fontWeight: 500,
            cursor: actionLoading ? 'not-allowed' : 'pointer',
            display: 'flex',
            alignItems: 'center',
            gap: '5px'
          }}
        >
          <Clock size={13} />
          Evaluate Predictive ETA
        </button>

        <button
          onClick={handleReplan}
          disabled={actionLoading}
          style={{
            padding: '6px 12px',
            backgroundColor: '#FFFFFF',
            color: '#334155',
            border: '1px solid #CBD5E1',
            borderRadius: '6px',
            fontSize: '12px',
            fontWeight: 500,
            cursor: actionLoading ? 'not-allowed' : 'pointer',
            display: 'flex',
            alignItems: 'center',
            gap: '5px'
          }}
        >
          <RefreshCw size={13} />
          Adaptive Replan
        </button>

        {lifecycle.current_stage === 'DELIVERED' && (
          <button
            onClick={handlePostDeliveryAudit}
            disabled={actionLoading}
            style={{
              padding: '6px 12px',
              backgroundColor: '#047857',
              color: '#FFFFFF',
              border: 'none',
              borderRadius: '6px',
              fontSize: '12px',
              fontWeight: 500,
              cursor: actionLoading ? 'not-allowed' : 'pointer',
              display: 'flex',
              alignItems: 'center',
              gap: '5px'
            }}
          >
            <FileCheck size={13} />
            Post-Delivery Audit
          </button>
        )}
      </div>
    </div>
  );
}

import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import toast from 'react-hot-toast';

export default function RFQRequirements({ rfq, completeness, requirements = null, onSwitchTab }) {
  const navigate = useNavigate();
  const [collapsedGroups, setCollapsedGroups] = useState({});

  // Fallback calculations if requirements endpoint is loading or unavailable
  const opReadiness = requirements?.operational_readiness;
  const overallStatus = opReadiness?.overall_status || (completeness.isComplete ? 'READY_FOR_QUOTATION' : 'INFORMATION_REQUIRED');
  const blockingCount = opReadiness?.blocking_count ?? (completeness.isComplete ? 0 : completeness.missingFields?.length || 1);
  const missingRequired = opReadiness?.missing_required_count ?? 0;
  const conditionalAttn = opReadiness?.conditional_attention_count ?? 0;
  const totalCount = opReadiness?.total_count || 12;
  const completeCount = opReadiness?.complete_count || (completeness.isComplete ? 12 : 7);
  const readinessScore = opReadiness?.readiness_score ?? (completeness.isComplete ? 100 : completeness.percentage || 60);
  const nextBestAction = opReadiness?.next_best_action || (completeness.isComplete
    ? 'All quotation-stage requirements are complete. Proceed to generate and send quotation.'
    : 'Complete missing mandatory shipment parameters before proceeding.');

  const groups = requirements?.groups || [];
  const docRequirements = requirements?.document_requirements || [];
  const aiFindings = requirements?.ai_findings || [];
  const leadId = rfq?.lead_id;

  const isReadyToQuote = overallStatus === 'READY_FOR_QUOTATION' || overallStatus === 'ATTENTION_REQUIRED';

  const toggleGroup = (category) => {
    setCollapsedGroups(prev => ({ ...prev, [category]: !prev[category] }));
  };

  // Status badge styling helper
  const getStatusBadge = (status) => {
    switch (status) {
      case 'READY_FOR_QUOTATION':
        return { label: 'Ready for Quotation', bg: '#DCFCE7', color: '#15803D', border: '#86EFAC', icon: '🟢' };
      case 'ATTENTION_REQUIRED':
        return { label: 'Attention Required', bg: '#FEF3C7', color: '#B45309', border: '#FCD34D', icon: '🟡' };
      case 'INFORMATION_REQUIRED':
        return { label: 'Information Required', bg: '#FEE2E2', color: '#B91C1C', border: '#FCA5A5', icon: '🔴' };
      case 'REQUIREMENTS_INCOMPLETE':
        return { label: 'Requirements Incomplete', bg: '#FFEDD5', color: '#C2410C', border: '#FDBA74', icon: '🟠' };
      default:
        return { label: 'Under Review', bg: '#F1F5F9', color: '#475569', border: '#CBD5E1', icon: '⚪' };
    }
  };

  const getSeverityBadge = (severity) => {
    switch (severity) {
      case 'BLOCKING':
        return { label: 'BLOCKING', bg: '#FEF2F2', color: '#DC2626', border: '#FECACA' };
      case 'REQUIRED':
        return { label: 'REQUIRED', bg: '#FFFBEB', color: '#D97706', border: '#FDE68A' };
      case 'CONDITIONAL':
        return { label: 'CONDITIONAL', bg: '#FAF5FF', color: '#7C3AED', border: '#E9D5FF' };
      case 'OPTIONAL':
        return { label: 'OPTIONAL', bg: '#F8FAFC', color: '#64748B', border: '#E2E8F0' };
      case 'INFORMATIONAL':
        return { label: 'INFO', bg: '#EFF6FF', color: '#2563EB', border: '#BFDBFE' };
      default:
        return { label: severity, bg: '#F8FAFC', color: '#64748B', border: '#E2E8F0' };
    }
  };

  const getReqStatusIcon = (status) => {
    switch (status) {
      case 'SATISFIED':
        return <span style={{ color: '#10B981', fontWeight: 900, fontSize: '14px' }}>✓</span>;
      case 'MISSING':
        return <span style={{ color: '#EF4444', fontWeight: 900, fontSize: '13px' }}>✕</span>;
      case 'UNDER_REVIEW':
        return <span style={{ color: '#F59E0B', fontWeight: 900, fontSize: '13px' }}>⏳</span>;
      case 'NOT_APPLICABLE':
        return <span style={{ color: '#94A3B8', fontWeight: 700, fontSize: '12px' }}>—</span>;
      default:
        return <span style={{ color: '#64748B', fontWeight: 700, fontSize: '12px' }}>○</span>;
    }
  };

  const statusBadge = getStatusBadge(overallStatus);

  // Group current vs future stage documents
  const currentStageDocs = docRequirements.filter(d => d.applicable_stage === 'RFQ_STAGE');
  const futureStageDocs = docRequirements.filter(d => d.applicable_stage !== 'RFQ_STAGE');

  if (requirements === null) {
    return (
      <div style={{
        background: '#FFFFFF',
        borderRadius: '16px',
        border: '1px solid #E2E8F0',
        padding: '48px 24px',
        textAlign: 'center',
        boxShadow: '0 2px 8px rgba(15, 23, 42, 0.04)'
      }} data-testid="rfq-requirements-loading">
        <div style={{
          width: '48px',
          height: '48px',
          background: '#EEF2FF',
          border: '1px solid #E0E7FF',
          borderRadius: '16px',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          margin: '0 auto 14px auto',
          fontSize: '22px'
        }}>
          <span style={{ animation: 'spin 1.5s linear infinite', display: 'inline-block' }}>🔄</span>
        </div>
        <h3 style={{ fontSize: '15px', fontWeight: 800, color: '#0F172A', marginBottom: '4px' }}>
          Evaluating Operational Requirements...
        </h3>
        <p style={{ fontSize: '12px', color: '#64748B', maxWidth: '420px', margin: '0 auto', lineHeight: 1.5 }}>
          Running server-side rules engine, evaluating trade compliance criteria, and validating documentation gates.
        </p>
      </div>
    );
  }

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: '22px' }}>
      
      {/* ── HEADER BANNER ─────────────────────────────────────────────────── */}
      <div
        style={{
          background: '#FFFFFF',
          borderRadius: '14px',
          border: '1px solid #E2E8F0',
          padding: '24px 28px',
          boxShadow: '0 2px 8px rgba(15, 23, 42, 0.04)',
          position: 'relative',
          overflow: 'hidden',
        }}
      >
        <div style={{ position: 'absolute', top: 0, left: 0, right: 0, height: '4px', background: isReadyToQuote ? 'linear-gradient(90deg, #10B981, #3B82F6)' : 'linear-gradient(90deg, #F59E0B, #EF4444)' }} />
        
        <div style={{ display: 'flex', alignItems: 'flex-start', justifyContent: 'space-between', flexWrap: 'wrap', gap: '16px' }}>
          <div>
            <div style={{ display: 'flex', alignItems: 'center', gap: '10px', marginBottom: '6px' }}>
              <h2 style={{ fontSize: '20px', fontWeight: 900, color: '#0F172A', margin: 0, letterSpacing: '-0.02em' }}>
                Requirements & Operational Readiness
              </h2>
              <span
                style={{
                  background: statusBadge.bg,
                  color: statusBadge.color,
                  border: `1px solid ${statusBadge.border}`,
                  padding: '3px 10px',
                  borderRadius: '12px',
                  fontSize: '12px',
                  fontWeight: 800,
                  display: 'flex',
                  alignItems: 'center',
                  gap: '5px',
                }}
              >
                <span>{statusBadge.icon}</span>
                <span>{statusBadge.label}</span>
              </span>
            </div>
            <p style={{ fontSize: '13px', color: '#64748B', margin: 0, maxWidth: '780px', lineHeight: 1.5 }}>
              Deterministic compliance, operational readiness, and documentation intelligence evaluating whether this RFQ is fully validated for carrier quotation, rate comparison, and downstream booking.
            </p>
          </div>

          {/* Quick Actions */}
          <div style={{ display: 'flex', alignItems: 'center', gap: '10px' }}>
            {leadId && (
              <button
                onClick={() => navigate(`/dashboard/leads?leadId=${leadId}&tab=emails`)}
                style={{
                  background: '#F8FAFC',
                  color: '#475569',
                  border: '1px solid #CBD5E1',
                  borderRadius: '8px',
                  padding: '9px 14px',
                  fontSize: '12.5px',
                  fontWeight: 700,
                  cursor: 'pointer',
                  display: 'flex',
                  alignItems: 'center',
                  gap: '6px',
                  transition: 'all 0.15s',
                }}
                onMouseEnter={e => e.currentTarget.style.background = '#F1F5F9'}
                onMouseLeave={e => e.currentTarget.style.background = '#F8FAFC'}
              >
                <span>💬</span>
                <span>Source Lead #{leadId}</span>
              </button>
            )}

            <button
              onClick={() => {
                if (onSwitchTab) {
                  onSwitchTab('quotes');
                } else {
                  toast.success('Navigating to Quotation Workspace...');
                }
              }}
              style={{
                background: isReadyToQuote ? 'linear-gradient(135deg, #4F46E5 0%, #4338CA 100%)' : '#94A3B8',
                color: '#FFFFFF',
                border: 'none',
                borderRadius: '8px',
                padding: '9px 18px',
                fontSize: '13px',
                fontWeight: 800,
                cursor: isReadyToQuote ? 'pointer' : 'not-allowed',
                boxShadow: isReadyToQuote ? '0 2px 6px rgba(79, 70, 229, 0.3)' : 'none',
                display: 'flex',
                alignItems: 'center',
                gap: '6px',
                transition: 'all 0.15s',
              }}
            >
              <span>{isReadyToQuote ? '⚡ Proceed to Quotes →' : '🔒 Quoting Locked'}</span>
            </button>
          </div>
        </div>
      </div>

      {/* ── METRICS SUMMARY BAR ───────────────────────────────────────────── */}
      <div
        style={{
          display: 'grid',
          gridTemplateColumns: 'repeat(auto-fit, minmax(210px, 1fr))',
          gap: '16px',
        }}
      >
        {/* Metric 1: Overall Score */}
        <div style={{ background: '#FFFFFF', border: '1px solid #E2E8F0', borderRadius: '12px', padding: '16px 20px', display: 'flex', alignItems: 'center', gap: '16px', boxShadow: '0 1px 3px rgba(0,0,0,0.02)' }}>
          <div style={{ width: '48px', height: '48px', borderRadius: '50%', background: isReadyToQuote ? '#ECFDF5' : '#FFFBEB', border: `3px solid ${isReadyToQuote ? '#10B981' : '#F59E0B'}`, display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: '14px', fontWeight: 900, color: isReadyToQuote ? '#065F46' : '#92400E', flexShrink: 0 }}>
            {readinessScore}%
          </div>
          <div>
            <div style={{ fontSize: '11px', fontWeight: 800, color: '#64748B', textTransform: 'uppercase', letterSpacing: '0.05em' }}>Readiness Score</div>
            <div style={{ fontSize: '17px', fontWeight: 900, color: '#0F172A', marginTop: '2px' }}>
              {completeCount} / {totalCount} Items
            </div>
          </div>
        </div>

        {/* Metric 2: Critical Blockers */}
        <div style={{ background: '#FFFFFF', border: `1px solid ${blockingCount > 0 ? '#FECACA' : '#E2E8F0'}`, borderRadius: '12px', padding: '16px 20px', display: 'flex', alignItems: 'center', gap: '14px', boxShadow: '0 1px 3px rgba(0,0,0,0.02)' }}>
          <div style={{ width: '40px', height: '40px', borderRadius: '10px', background: blockingCount > 0 ? '#FEF2F2' : '#F8FAFC', color: blockingCount > 0 ? '#DC2626' : '#10B981', display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: '18px', fontWeight: 900, flexShrink: 0 }}>
            {blockingCount > 0 ? '✕' : '✓'}
          </div>
          <div>
            <div style={{ fontSize: '11px', fontWeight: 800, color: '#64748B', textTransform: 'uppercase', letterSpacing: '0.05em' }}>Critical Blockers</div>
            <div style={{ fontSize: '17px', fontWeight: 900, color: blockingCount > 0 ? '#DC2626' : '#059669', marginTop: '2px' }}>
              {blockingCount === 0 ? '0 Blockers' : `${blockingCount} Blocking`}
            </div>
          </div>
        </div>

        {/* Metric 3: Required / Attention Items */}
        <div style={{ background: '#FFFFFF', border: '1px solid #E2E8F0', borderRadius: '12px', padding: '16px 20px', display: 'flex', alignItems: 'center', gap: '14px', boxShadow: '0 1px 3px rgba(0,0,0,0.02)' }}>
          <div style={{ width: '40px', height: '40px', borderRadius: '10px', background: conditionalAttn > 0 ? '#FFFBEB' : '#F8FAFC', color: conditionalAttn > 0 ? '#D97706' : '#64748B', display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: '18px', fontWeight: 900, flexShrink: 0 }}>
            {conditionalAttn > 0 ? '⚠️' : '✦'}
          </div>
          <div>
            <div style={{ fontSize: '11px', fontWeight: 800, color: '#64748B', textTransform: 'uppercase', letterSpacing: '0.05em' }}>Conditional & Attention</div>
            <div style={{ fontSize: '17px', fontWeight: 900, color: '#0F172A', marginTop: '2px' }}>
              {conditionalAttn} Item{conditionalAttn !== 1 ? 's' : ''}
            </div>
          </div>
        </div>

        {/* Metric 4: Document Stage Status */}
        <div style={{ background: '#FFFFFF', border: '1px solid #E2E8F0', borderRadius: '12px', padding: '16px 20px', display: 'flex', alignItems: 'center', gap: '14px', boxShadow: '0 1px 3px rgba(0,0,0,0.02)' }}>
          <div style={{ width: '40px', height: '40px', borderRadius: '10px', background: '#EFF6FF', color: '#2563EB', display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: '18px', fontWeight: 900, flexShrink: 0 }}>
            📄
          </div>
          <div>
            <div style={{ fontSize: '11px', fontWeight: 800, color: '#64748B', textTransform: 'uppercase', letterSpacing: '0.05em' }}>Current Stage Docs</div>
            <div style={{ fontSize: '17px', fontWeight: 900, color: '#0F172A', marginTop: '2px' }}>
              {currentStageDocs.filter(d => d.status === 'SATISFIED').length} / {currentStageDocs.length || 2} Ready
            </div>
          </div>
        </div>
      </div>

      {/* ── NEXT BEST ACTION OPERATIONAL BANNER ────────────────────────────── */}
      <div
        style={{
          background: isReadyToQuote ? 'linear-gradient(135deg, #F0FDF4 0%, #ECFDF5 100%)' : 'linear-gradient(135deg, #FFFBEB 0%, #FEF3C7 100%)',
          border: `1px solid ${isReadyToQuote ? '#A7F3D0' : '#FDE68A'}`,
          borderRadius: '12px',
          padding: '16px 20px',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          flexWrap: 'wrap',
          gap: '12px',
        }}
      >
        <div style={{ display: 'flex', alignItems: 'center', gap: '12px' }}>
          <span style={{ fontSize: '22px' }}>{isReadyToQuote ? '💡' : '⚠️'}</span>
          <div>
            <strong style={{ fontSize: '13px', fontWeight: 800, color: isReadyToQuote ? '#065F46' : '#92400E', display: 'block' }}>
              Next Best Action
            </strong>
            <span style={{ fontSize: '12.5px', color: isReadyToQuote ? '#047857' : '#78350F', lineHeight: 1.4 }}>
              {nextBestAction}
            </span>
          </div>
        </div>

        {!isReadyToQuote ? (
          <button
            onClick={() => onSwitchTab ? onSwitchTab('cargo') : null}
            style={{
              background: '#FFFFFF',
              color: '#B45309',
              border: '1px solid #FCD34D',
              borderRadius: '7px',
              padding: '6px 12px',
              fontSize: '12px',
              fontWeight: 700,
              cursor: 'pointer',
              boxShadow: '0 1px 2px rgba(0,0,0,0.04)',
            }}
          >
            Fix Missing Details →
          </button>
        ) : (
          <button
            onClick={() => onSwitchTab ? onSwitchTab('quotes') : null}
            style={{
              background: '#059669',
              color: '#FFFFFF',
              border: 'none',
              borderRadius: '7px',
              padding: '6px 14px',
              fontSize: '12px',
              fontWeight: 700,
              cursor: 'pointer',
              boxShadow: '0 1px 2px rgba(0,0,0,0.06)',
            }}
          >
            View Carrier Quotes →
          </button>
        )}
      </div>

      {/* ── 6 REQUIREMENT GROUPS LIST ─────────────────────────────────────── */}
      <div style={{ display: 'flex', flexDirection: 'column', gap: '18px' }}>
        
        {/* Render each dynamic group from the backend engine */}
        {groups.length > 0 ? (
          groups.map((group) => {
            const isCollapsed = collapsedGroups[group.category];
            const isGroupComplete = group.status === 'COMPLETE';
            const isAIFindings = group.category === 'AI_FINDINGS';

            return (
              <div
                key={group.category}
                style={{
                  background: isAIFindings ? 'linear-gradient(180deg, #FAF5FF 0%, #FFFFFF 100%)' : '#FFFFFF',
                  borderRadius: '14px',
                  border: `1px solid ${isAIFindings ? '#E9D5FF' : '#E2E8F0'}`,
                  boxShadow: '0 1px 4px rgba(15, 23, 42, 0.03)',
                  overflow: 'hidden',
                }}
              >
                {/* Group Header */}
                <div
                  onClick={() => toggleGroup(group.category)}
                  style={{
                    padding: '16px 22px',
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'space-between',
                    cursor: 'pointer',
                    background: isAIFindings ? '#FAF5FF' : '#F8FAFC',
                    borderBottom: isCollapsed ? 'none' : '1px solid #E2E8F0',
                    userSelect: 'none',
                  }}
                >
                  <div style={{ display: 'flex', alignItems: 'center', gap: '12px' }}>
                    <span style={{ fontSize: '18px' }}>{group.icon || '📌'}</span>
                    <div>
                      <h3 style={{ fontSize: '14.5px', fontWeight: 800, color: '#0F172A', margin: 0 }}>
                        {group.title}
                      </h3>
                      <span style={{ fontSize: '12px', color: '#64748B', fontWeight: 600 }}>
                        {group.complete_count} of {group.total_count} satisfied
                      </span>
                    </div>
                  </div>

                  <div style={{ display: 'flex', alignItems: 'center', gap: '10px' }}>
                    <span
                      style={{
                        fontSize: '11px',
                        fontWeight: 800,
                        padding: '3px 8px',
                        borderRadius: '6px',
                        background: isGroupComplete ? '#DCFCE7' : (group.status === 'ATTENTION' ? '#FEF3C7' : '#FEE2E2'),
                        color: isGroupComplete ? '#15803D' : (group.status === 'ATTENTION' ? '#B45309' : '#B91C1C'),
                      }}
                    >
                      {group.status}
                    </span>
                    <span style={{ fontSize: '14px', color: '#64748B', transform: isCollapsed ? 'rotate(-90deg)' : 'rotate(0deg)', transition: 'transform 0.15s' }}>
                      ▼
                    </span>
                  </div>
                </div>

                {/* Group Items Table */}
                {!isCollapsed && (
                  <div style={{ padding: '8px 16px 16px 16px' }}>
                    {group.requirements.map((req) => {
                      const sevBadge = getSeverityBadge(req.severity);
                      return (
                        <div
                          key={req.id}
                          style={{
                            display: 'flex',
                            alignItems: 'flex-start',
                            justifyContent: 'space-between',
                            padding: '12px 14px',
                            borderBottom: '1px solid #F1F5F9',
                            gap: '14px',
                          }}
                        >
                          {/* Left: Status Icon + Title + Description */}
                          <div style={{ display: 'flex', alignItems: 'flex-start', gap: '12px', flex: '1 1 auto' }}>
                            <div style={{ width: '22px', height: '22px', borderRadius: '50%', background: '#F8FAFC', border: '1px solid #E2E8F0', display: 'flex', alignItems: 'center', justifyContent: 'center', flexShrink: 0, marginTop: '2px' }}>
                              {getReqStatusIcon(req.status)}
                            </div>
                            <div>
                              <div style={{ display: 'flex', alignItems: 'center', gap: '8px', flexWrap: 'wrap' }}>
                                <strong style={{ fontSize: '13px', fontWeight: 700, color: '#0F172A' }}>
                                  {req.title}
                                </strong>
                                <span
                                  style={{
                                    fontSize: '10px',
                                    fontWeight: 800,
                                    padding: '1px 6px',
                                    borderRadius: '4px',
                                    background: sevBadge.bg,
                                    color: sevBadge.color,
                                    border: `1px solid ${sevBadge.border}`,
                                    letterSpacing: '0.04em',
                                  }}
                                >
                                  {sevBadge.label}
                                </span>
                              </div>
                              <p style={{ fontSize: '12px', color: '#64748B', margin: '3px 0 0 0', lineHeight: 1.4 }}>
                                {req.description}
                              </p>
                              {req.condition_reason && (
                                <div style={{ marginTop: '4px', fontSize: '11.5px', color: '#B45309', background: '#FFFBEB', padding: '3px 8px', borderRadius: '4px', border: '1px solid #FEF3C7' }}>
                                  💡 <em>{req.condition_reason}</em>
                                </div>
                              )}
                              {req.source_context && (
                                <div style={{ marginTop: '3px', fontSize: '11px', color: '#6D28D9', fontWeight: 600 }}>
                                  ✦ {req.source_context}
                                </div>
                              )}
                            </div>
                          </div>

                          {/* Right: Current Value */}
                          <div style={{ textAlign: 'right', flexShrink: 0, minWidth: '130px' }}>
                            <div style={{ fontSize: '12.5px', fontWeight: 700, color: req.value ? '#0F172A' : '#94A3B8' }}>
                              {req.value || 'Not provided'}
                            </div>
                            <div style={{ fontSize: '11px', color: req.status === 'SATISFIED' ? '#059669' : (req.status === 'MISSING' ? '#DC2626' : '#64748B'), fontWeight: 600, marginTop: '2px' }}>
                              {req.status}
                            </div>
                          </div>
                        </div>
                      );
                    })}
                  </div>
                )}
              </div>
            );
          })
        ) : (
          /* Fallback view if requirements response is empty */
          <div style={{ background: '#FFFFFF', borderRadius: '12px', border: '1px solid #E2E8F0', padding: '24px', textAlign: 'center' }}>
            <div style={{ fontSize: '24px', marginBottom: '8px' }}>🔍</div>
            <div style={{ fontSize: '14px', fontWeight: 700, color: '#0F172A' }}>Evaluating Requirements...</div>
            <p style={{ fontSize: '12.5px', color: '#64748B', marginTop: '4px' }}>
              The requirements engine is processing shipment parameters and trade compliance rules.
            </p>
          </div>
        )}

        {/* ── DOCUMENT REQUIREMENTS STAGE SPLIT SECTION ────────────────── */}
        <div
          style={{
            background: '#FFFFFF',
            borderRadius: '14px',
            border: '1px solid #E2E8F0',
            boxShadow: '0 1px 4px rgba(15, 23, 42, 0.03)',
            overflow: 'hidden',
          }}
        >
          <div style={{ padding: '16px 22px', background: '#F8FAFC', borderBottom: '1px solid #E2E8F0', display: 'flex', alignItems: 'center', justifyContent: 'space-between', flexWrap: 'wrap', gap: '12px' }}>
            <div style={{ display: 'flex', alignItems: 'center', gap: '10px' }}>
              <span style={{ fontSize: '18px' }}>📂</span>
              <div>
                <h3 style={{ fontSize: '14.5px', fontWeight: 800, color: '#0F172A', margin: 0 }}>
                  Lifecycle Document Requirements (Stage-Aware)
                </h3>
                <span style={{ fontSize: '12px', color: '#64748B', fontWeight: 600 }}>
                  Future stage documents do NOT block RFQ quotation or carrier rate comparison
                </span>
              </div>
            </div>
            {onSwitchTab && (
              <button
                onClick={() => onSwitchTab('documents')}
                style={{
                  background: '#EEF2FF',
                  color: '#4F46E5',
                  border: '1px solid #C7D2FE',
                  borderRadius: '8px',
                  padding: '6px 12px',
                  fontSize: '12px',
                  fontWeight: 700,
                  cursor: 'pointer',
                  transition: 'all 0.15s ease',
                }}
                onMouseEnter={(e) => { e.currentTarget.style.background = '#E0E7FF'; }}
                onMouseLeave={(e) => { e.currentTarget.style.background = '#EEF2FF'; }}
              >
                Manage in Documents Workspace →
              </button>
            )}
          </div>

          <div style={{ padding: '18px 22px', display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(320px, 1fr))', gap: '18px' }}>
            
            {/* Current Stage Documents */}
            <div style={{ background: '#F8FAFC', border: '1px solid #E2E8F0', borderRadius: '10px', padding: '16px' }}>
              <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: '12px' }}>
                <strong style={{ fontSize: '13px', fontWeight: 800, color: '#0F172A' }}>
                  🎯 RFQ & Quotation Stage
                </strong>
                <span style={{ fontSize: '11px', fontWeight: 700, color: '#D97706', background: '#FEF3C7', padding: '2px 7px', borderRadius: '6px' }}>
                  Required Now
                </span>
              </div>
              <div style={{ display: 'flex', flexDirection: 'column', gap: '8px' }}>
                {currentStageDocs.length > 0 ? (
                  currentStageDocs.map(doc => {
                    const isSatisfied = doc.status === 'SATISFIED' || doc.document_status === 'APPROVED';
                    const isReview = doc.document_status === 'UNDER_REVIEW' || doc.document_status === 'UPLOADED';
                    const bg = isSatisfied ? '#ECFDF5' : isReview ? '#FFFBEB' : '#FFF1F2';
                    const border = isSatisfied ? '#A7F3D0' : isReview ? '#FEF3C7' : '#FECDD3';
                    const color = isSatisfied ? '#059669' : isReview ? '#D97706' : '#DC2626';
                    const label = isSatisfied ? '✓ Approved' : isReview ? `⏳ ${doc.document_status || 'Under Review'}` : '✕ Missing';

                    return (
                      <div key={doc.doc_type} style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', background: '#FFFFFF', padding: '8px 12px', borderRadius: '6px', border: '1px solid #E2E8F0', fontSize: '12px' }}>
                        <span style={{ fontWeight: 600, color: '#1E293B' }}>📋 {doc.title}</span>
                        <span style={{ color: color, background: bg, border: `1px solid ${border}`, fontWeight: 800, fontSize: '11px', padding: '2px 6px', borderRadius: '4px' }}>
                          {label}
                        </span>
                      </div>
                    );
                  })
                ) : (
                  <>
                    <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', background: '#FFFFFF', padding: '8px 12px', borderRadius: '6px', border: '1px solid #E2E8F0', fontSize: '12px' }}>
                      <span style={{ fontWeight: 600, color: '#1E293B' }}>📋 Commercial Invoice</span>
                      <span style={{ color: '#DC2626', background: '#FFF1F2', border: '1px solid #FECDD3', fontWeight: 800, fontSize: '11px', padding: '2px 6px', borderRadius: '4px' }}>✕ Missing</span>
                    </div>
                    <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', background: '#FFFFFF', padding: '8px 12px', borderRadius: '6px', border: '1px solid #E2E8F0', fontSize: '12px' }}>
                      <span style={{ fontWeight: 600, color: '#1E293B' }}>📦 Packing List</span>
                      <span style={{ color: '#DC2626', background: '#FFF1F2', border: '1px solid #FECDD3', fontWeight: 800, fontSize: '11px', padding: '2px 6px', borderRadius: '4px' }}>✕ Missing</span>
                    </div>
                  </>
                )}
              </div>
            </div>


            {/* Future Stage Documents */}
            <div style={{ background: '#F8FAFC', border: '1px solid #E2E8F0', borderRadius: '10px', padding: '16px' }}>
              <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: '12px' }}>
                <strong style={{ fontSize: '13px', fontWeight: 800, color: '#0F172A' }}>
                  🚢 Downstream Operational Documents
                </strong>
                <span style={{ fontSize: '11px', fontWeight: 700, color: '#64748B', background: '#E2E8F0', padding: '2px 7px', borderRadius: '6px' }}>
                  Non-Blocking
                </span>
              </div>
              <div style={{ display: 'flex', flexDirection: 'column', gap: '8px' }}>
                {futureStageDocs.length > 0 ? (
                  futureStageDocs.map(doc => (
                    <div key={doc.doc_type} style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', background: '#FFFFFF', padding: '8px 12px', borderRadius: '6px', border: '1px solid #E2E8F0', fontSize: '12px' }}>
                      <div>
                        <span style={{ fontWeight: 600, color: '#475569' }}>📑 {doc.title}</span>
                        <span style={{ display: 'block', fontSize: '10.5px', color: '#94A3B8' }}>{doc.applicable_stage}</span>
                      </div>
                      <span style={{ color: '#64748B', fontWeight: 700, fontSize: '11px', background: '#F1F5F9', padding: '2px 6px', borderRadius: '4px' }}>
                        Not Applicable Now
                      </span>
                    </div>
                  ))
                ) : (
                  <>
                    <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', background: '#FFFFFF', padding: '8px 12px', borderRadius: '6px', border: '1px solid #E2E8F0', fontSize: '12px' }}>
                      <span style={{ fontWeight: 600, color: '#475569' }}>📑 Bill of Lading (OBL / HBL / MBL)</span>
                      <span style={{ color: '#64748B', fontWeight: 700, fontSize: '11px', background: '#F1F5F9', padding: '2px 6px', borderRadius: '4px' }}>Shipment Execution</span>
                    </div>
                    <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', background: '#FFFFFF', padding: '8px 12px', borderRadius: '6px', border: '1px solid #E2E8F0', fontSize: '12px' }}>
                      <span style={{ fontWeight: 600, color: '#475569' }}>🌍 Certificate of Origin</span>
                      <span style={{ color: '#64748B', fontWeight: 700, fontSize: '11px', background: '#F1F5F9', padding: '2px 6px', borderRadius: '4px' }}>Customs Clearance</span>
                    </div>
                  </>
                )}
              </div>
            </div>

          </div>
        </div>

        {/* ── AI OPERATIONAL FINDINGS SECTION (if present) ───────────────── */}
        {aiFindings.length > 0 && (
          <div
            style={{
              background: 'linear-gradient(135deg, #FAF5FF 0%, #FFFFFF 100%)',
              borderRadius: '14px',
              border: '1px solid #E9D5FF',
              padding: '20px 24px',
              boxShadow: '0 1px 4px rgba(139, 92, 246, 0.05)',
            }}
          >
            <div style={{ display: 'flex', alignItems: 'center', gap: '10px', marginBottom: '14px' }}>
              <span style={{ fontSize: '18px', color: '#7C3AED' }}>✦</span>
              <h3 style={{ fontSize: '14.5px', fontWeight: 800, color: '#5B21B6', margin: 0 }}>
                AI Intelligence & Risk Analysis Findings
              </h3>
            </div>

            <div style={{ display: 'flex', flexDirection: 'column', gap: '10px' }}>
              {aiFindings.map((finding) => (
                <div
                  key={finding.id}
                  style={{
                    background: '#FFFFFF',
                    border: '1px solid #EDE9FE',
                    borderRadius: '10px',
                    padding: '14px',
                    display: 'flex',
                    alignItems: 'flex-start',
                    justifyContent: 'space-between',
                    gap: '14px',
                  }}
                >
                  <div>
                    <div style={{ display: 'flex', alignItems: 'center', gap: '8px', marginBottom: '4px' }}>
                      <strong style={{ fontSize: '13px', fontWeight: 800, color: '#0F172A' }}>
                        {finding.title}
                      </strong>
                      <span style={{ fontSize: '10.5px', fontWeight: 800, color: '#6D28D9', background: '#EDE9FE', padding: '1px 6px', borderRadius: '4px' }}>
                        {finding.confidence} Confidence
                      </span>
                    </div>
                    <p style={{ fontSize: '12.5px', color: '#475569', margin: 0, lineHeight: 1.4 }}>
                      {finding.description}
                    </p>
                    {finding.recommendation && (
                      <div style={{ marginTop: '6px', fontSize: '12px', color: '#5B21B6', fontWeight: 600 }}>
                        Recommendation: {finding.recommendation}
                      </div>
                    )}
                    {finding.source_context && (
                      <div style={{ marginTop: '4px', fontSize: '11px', color: '#8B5CF6' }}>
                        ✦ {finding.source_context}
                      </div>
                    )}
                  </div>
                </div>
              ))}
            </div>
          </div>
        )}

      </div>

    </div>
  );
}

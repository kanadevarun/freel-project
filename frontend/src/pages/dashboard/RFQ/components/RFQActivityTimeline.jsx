import React, { useState, useMemo } from 'react';
import { useNavigate } from 'react-router-dom';

const CATEGORY_CONFIG = {
  CUSTOMER: { bg: '#ECFDF5', text: '#065F46', border: '#A7F3D0', icon: '✉️', label: 'Customer' },
  OPERATIONS: { bg: '#EFF6FF', text: '#1E40AF', border: '#BFDBFE', icon: '⚡', label: 'Operations' },
  AI: { bg: '#FAF5FF', text: '#6D28D9', border: '#E9D5FF', icon: '✦', label: 'AI Intelligence' },
  REQUIREMENTS: { bg: '#FFFBEB', text: '#92400E', border: '#FDE68A', icon: '⚠️', label: 'Requirements' },
  QUOTES: { bg: '#F0FDFA', text: '#0F766E', border: '#99F6E4', icon: '💰', label: 'Quotes' },
  DOCUMENTS: { bg: '#FFF7ED', text: '#C2410C', border: '#FED7AA', icon: '📄', label: 'Documents' },
  SYSTEM: { bg: '#F8FAFC', text: '#475569', border: '#E2E8F0', icon: '⚙️', label: 'System' },
};

export default function RFQActivityTimeline({
  rfq,
  activityData = null,
  timelineEvents = [],
  isLoading = false,
  onRefresh,
}) {
  const navigate = useNavigate();
  const [selectedFilter, setSelectedFilter] = useState('ALL');

  const leadId = rfq?.lead_id;

  // Normalized events from backend activity endpoint, with fallback to raw timeline events
  const rawEvents = activityData?.events || [];
  const events = useMemo(() => {
    if (rawEvents.length > 0) return rawEvents;

    // Fallback: Map legacy timelineEvents if activityData is still loading or empty
    return timelineEvents.map((e, idx) => ({
      id: e.id || `legacy-${idx}`,
      type: e.action,
      category: e.category || 'OPERATIONS',
      title: e.action?.replace(/_/g, ' ') || 'Operational Event',
      description: e.description || '',
      timestamp: e.timestamp,
      actor_type: 'SYSTEM',
      actor_name: e.actor || 'System',
      source_type: e.entity_type,
      source_id: String(e.entity_id || ''),
      is_important: false,
      requires_action: false,
    }));
  }, [rawEvents, timelineEvents]);

  // Activity summary metrics
  const summary = activityData?.summary || {
    total_events: events.length,
    customer_events: events.filter(e => e.category === 'CUSTOMER').length,
    operational_events: events.filter(e => e.category === 'OPERATIONS' || e.category === 'SYSTEM').length,
    ai_events: events.filter(e => e.category === 'AI').length,
    requirements_events: events.filter(e => e.category === 'REQUIREMENTS').length,
    document_events: events.filter(e => e.category === 'DOCUMENTS').length,
    quote_events: events.filter(e => e.category === 'QUOTES').length,
    action_required_count: events.filter(e => e.requires_action).length,
  };

  // Filter items
  const filteredEvents = useMemo(() => {
    if (selectedFilter === 'ALL') return events;
    return events.filter(e => {
      if (selectedFilter === 'CUSTOMER') return e.category === 'CUSTOMER';
      if (selectedFilter === 'OPERATIONS') return e.category === 'OPERATIONS' || e.category === 'SYSTEM';
      if (selectedFilter === 'AI') return e.category === 'AI';
      if (selectedFilter === 'REQUIREMENTS') return e.category === 'REQUIREMENTS';
      if (selectedFilter === 'DOCUMENTS') return e.category === 'DOCUMENTS';
      if (selectedFilter === 'QUOTES') return e.category === 'QUOTES';
      return true;
    });
  }, [events, selectedFilter]);

  const filterTabs = [
    { key: 'ALL', label: 'All Activity', count: summary.total_events },
    { key: 'CUSTOMER', label: 'Customer', count: summary.customer_events },
    { key: 'OPERATIONS', label: 'Operations', count: summary.operational_events },
    { key: 'AI', label: 'AI Intelligence', count: summary.ai_events },
    { key: 'REQUIREMENTS', label: 'Requirements', count: summary.requirements_events },
    { key: 'DOCUMENTS', label: 'Documents', count: summary.document_events },
    { key: 'QUOTES', label: 'Quotes', count: summary.quote_events },
  ];

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: '22px' }}>

      {/* ── SOURCE LEAD CONTEXT BANNER (if lead-originated) ─────────────── */}
      {leadId && (
        <div
          style={{
            background: 'linear-gradient(135deg, #F0FDF4 0%, #FFFFFF 100%)',
            borderRadius: '14px',
            border: '1px solid #BBF7D0',
            padding: '18px 22px',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
            flexWrap: 'wrap',
            gap: '16px',
            boxShadow: '0 1px 3px rgba(16, 185, 129, 0.04)',
          }}
        >
          <div style={{ display: 'flex', alignItems: 'center', gap: '14px' }}>
            <div
              style={{
                width: '42px',
                height: '42px',
                borderRadius: '10px',
                background: '#DCFCE7',
                border: '1px solid #86EFAC',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                fontSize: '20px',
                flexShrink: 0,
              }}
            >
              ✉️
            </div>
            <div>
              <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                <h4 style={{ fontSize: '14px', fontWeight: 800, color: '#0F172A', margin: 0 }}>
                  Originated from Lead #{leadId}
                </h4>
                <span style={{ fontSize: '11px', fontWeight: 800, color: '#059669', background: '#DCFCE7', padding: '1px 7px', borderRadius: '10px' }}>
                  Lineage Connected
                </span>
              </div>
              <p style={{ fontSize: '12.5px', color: '#475569', margin: '3px 0 0 0' }}>
                All historical customer email interactions, AI extraction milestones, and lead qualification events are preserved in this audit trail.
              </p>
            </div>
          </div>

          <button
            onClick={() => navigate(`/dashboard/leads?leadId=${leadId}&tab=emails`)}
            style={{
              background: '#FFFFFF',
              color: '#047857',
              border: '1px solid #86EFAC',
              borderRadius: '8px',
              padding: '8px 16px',
              fontSize: '12.5px',
              fontWeight: 700,
              cursor: 'pointer',
              display: 'inline-flex',
              alignItems: 'center',
              gap: '6px',
              boxShadow: '0 1px 2px rgba(0,0,0,0.03)',
              transition: 'all 0.15s ease',
            }}
            onMouseEnter={e => e.currentTarget.style.background = '#ECFDF5'}
            onMouseLeave={e => e.currentTarget.style.background = '#FFFFFF'}
          >
            <span>Open Email Thread</span>
            <span>→</span>
          </button>
        </div>
      )}

      {/* ── ACTIVITY SUMMARY METRICS BAR ─────────────────────────────────── */}
      <div
        style={{
          display: 'grid',
          gridTemplateColumns: 'repeat(auto-fit, minmax(190px, 1fr))',
          gap: '14px',
        }}
      >
        {/* Total Activity */}
        <div style={{ background: '#FFFFFF', border: '1px solid #E2E8F0', borderRadius: '12px', padding: '14px 18px', display: 'flex', alignItems: 'center', gap: '12px', boxShadow: '0 1px 3px rgba(0,0,0,0.02)' }}>
          <div style={{ width: '38px', height: '38px', borderRadius: '10px', background: '#F8FAFC', color: '#0F172A', display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: '16px', fontWeight: 900, flexShrink: 0 }}>
            ⏱️
          </div>
          <div>
            <div style={{ fontSize: '11px', fontWeight: 800, color: '#64748B', textTransform: 'uppercase', letterSpacing: '0.05em' }}>Total Events</div>
            <div style={{ fontSize: '18px', fontWeight: 900, color: '#0F172A', marginTop: '2px' }}>
              {summary.total_events}
            </div>
          </div>
        </div>

        {/* Customer Events */}
        <div style={{ background: '#FFFFFF', border: '1px solid #E2E8F0', borderRadius: '12px', padding: '14px 18px', display: 'flex', alignItems: 'center', gap: '12px', boxShadow: '0 1px 3px rgba(0,0,0,0.02)' }}>
          <div style={{ width: '38px', height: '38px', borderRadius: '10px', background: '#ECFDF5', color: '#059669', display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: '16px', fontWeight: 900, flexShrink: 0 }}>
            ✉️
          </div>
          <div>
            <div style={{ fontSize: '11px', fontWeight: 800, color: '#64748B', textTransform: 'uppercase', letterSpacing: '0.05em' }}>Customer Inquiries</div>
            <div style={{ fontSize: '18px', fontWeight: 900, color: '#065F46', marginTop: '2px' }}>
              {summary.customer_events}
            </div>
          </div>
        </div>

        {/* Operations */}
        <div style={{ background: '#FFFFFF', border: '1px solid #E2E8F0', borderRadius: '12px', padding: '14px 18px', display: 'flex', alignItems: 'center', gap: '12px', boxShadow: '0 1px 3px rgba(0,0,0,0.02)' }}>
          <div style={{ width: '38px', height: '38px', borderRadius: '10px', background: '#EFF6FF', color: '#2563EB', display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: '16px', fontWeight: 900, flexShrink: 0 }}>
            ⚡
          </div>
          <div>
            <div style={{ fontSize: '11px', fontWeight: 800, color: '#64748B', textTransform: 'uppercase', letterSpacing: '0.05em' }}>Operations</div>
            <div style={{ fontSize: '18px', fontWeight: 900, color: '#1E40AF', marginTop: '2px' }}>
              {summary.operational_events}
            </div>
          </div>
        </div>

        {/* AI Events */}
        <div style={{ background: '#FFFFFF', border: '1px solid #E2E8F0', borderRadius: '12px', padding: '14px 18px', display: 'flex', alignItems: 'center', gap: '12px', boxShadow: '0 1px 3px rgba(0,0,0,0.02)' }}>
          <div style={{ width: '38px', height: '38px', borderRadius: '10px', background: '#FAF5FF', color: '#7C3AED', display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: '16px', fontWeight: 900, flexShrink: 0 }}>
            ✦
          </div>
          <div>
            <div style={{ fontSize: '11px', fontWeight: 800, color: '#64748B', textTransform: 'uppercase', letterSpacing: '0.05em' }}>AI Extractions</div>
            <div style={{ fontSize: '18px', fontWeight: 900, color: '#6D28D9', marginTop: '2px' }}>
              {summary.ai_events}
            </div>
          </div>
        </div>

        {/* Action Required */}
        <div style={{ background: '#FFFFFF', border: `1px solid ${summary.action_required_count > 0 ? '#FECACA' : '#E2E8F0'}`, borderRadius: '12px', padding: '14px 18px', display: 'flex', alignItems: 'center', gap: '12px', boxShadow: '0 1px 3px rgba(0,0,0,0.02)' }}>
          <div style={{ width: '38px', height: '38px', borderRadius: '10px', background: summary.action_required_count > 0 ? '#FEF2F2' : '#F8FAFC', color: summary.action_required_count > 0 ? '#DC2626' : '#10B981', display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: '16px', fontWeight: 900, flexShrink: 0 }}>
            {summary.action_required_count > 0 ? '⚠️' : '✓'}
          </div>
          <div>
            <div style={{ fontSize: '11px', fontWeight: 800, color: '#64748B', textTransform: 'uppercase', letterSpacing: '0.05em' }}>Attention Needed</div>
            <div style={{ fontSize: '18px', fontWeight: 900, color: summary.action_required_count > 0 ? '#DC2626' : '#059669', marginTop: '2px' }}>
              {summary.action_required_count === 0 ? '0 Items' : `${summary.action_required_count} Action${summary.action_required_count > 1 ? 's' : ''}`}
            </div>
          </div>
        </div>
      </div>

      {/* ── MAIN TIMELINE CONTAINER ───────────────────────────────────────── */}
      <div
        style={{
          background: '#FFFFFF',
          borderRadius: '14px',
          border: '1px solid #E2E8F0',
          boxShadow: '0 1px 4px rgba(15, 23, 42, 0.03)',
          overflow: 'hidden',
        }}
      >
        {/* Header & Filter Controls */}
        <div
          style={{
            padding: '20px 24px 16px 24px',
            borderBottom: '1px solid #E2E8F0',
            background: '#FFFFFF',
          }}
        >
          <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: '16px', flexWrap: 'wrap', gap: '12px' }}>
            <div>
              <h3 style={{ fontSize: '16px', fontWeight: 900, color: '#0F172A', margin: 0, letterSpacing: '-0.01em' }}>
                Activity & Audit Trail
              </h3>
              <p style={{ fontSize: '12.5px', color: '#64748B', margin: '2px 0 0 0' }}>
                Chronological record of emails, AI parsing, operational milestones, and quote actions
              </p>
            </div>

            <div style={{ display: 'flex', alignItems: 'center', gap: '10px' }}>
              {onRefresh && (
                <button
                  onClick={onRefresh}
                  disabled={isLoading}
                  style={{
                    background: '#F8FAFC',
                    color: '#475569',
                    border: '1px solid #CBD5E1',
                    borderRadius: '8px',
                    padding: '6px 12px',
                    fontSize: '12px',
                    fontWeight: 700,
                    cursor: isLoading ? 'not-allowed' : 'pointer',
                    display: 'flex',
                    alignItems: 'center',
                    gap: '5px',
                    transition: 'all 0.15s',
                  }}
                  onMouseEnter={e => !isLoading && (e.currentTarget.style.background = '#F1F5F9')}
                  onMouseLeave={e => !isLoading && (e.currentTarget.style.background = '#F8FAFC')}
                >
                  <span style={{ transform: isLoading ? 'rotate(180deg)' : 'none', transition: 'transform 0.3s' }}>🔄</span>
                  <span>{isLoading ? 'Refreshing...' : 'Refresh'}</span>
                </button>
              )}
              <span style={{ fontSize: '12px', color: '#64748B', fontWeight: 700, background: '#F1F5F9', padding: '4px 10px', borderRadius: '12px' }}>
                {filteredEvents.length} {filteredEvents.length === 1 ? 'Event' : 'Events'}
              </span>
            </div>
          </div>

          {/* Filter Pills */}
          <div style={{ display: 'flex', alignItems: 'center', gap: '6px', overflowX: 'auto', paddingBottom: '4px' }}>
            {filterTabs.map((tab) => {
              const isSelected = selectedFilter === tab.key;
              return (
                <button
                  key={tab.key}
                  onClick={() => setSelectedFilter(tab.key)}
                  style={{
                    background: isSelected ? '#0F172A' : '#F8FAFC',
                    color: isSelected ? '#FFFFFF' : '#475569',
                    border: `1px solid ${isSelected ? '#0F172A' : '#E2E8F0'}`,
                    borderRadius: '8px',
                    padding: '5px 12px',
                    fontSize: '12px',
                    fontWeight: isSelected ? 800 : 600,
                    cursor: 'pointer',
                    display: 'flex',
                    alignItems: 'center',
                    gap: '6px',
                    whiteSpace: 'nowrap',
                    transition: 'all 0.15s ease',
                  }}
                >
                  <span>{tab.label}</span>
                  {tab.count !== undefined && (
                    <span
                      style={{
                        fontSize: '10.5px',
                        padding: '1px 6px',
                        borderRadius: '10px',
                        background: isSelected ? 'rgba(255,255,255,0.2)' : '#E2E8F0',
                        color: isSelected ? '#FFFFFF' : '#64748B',
                        fontWeight: 800,
                      }}
                    >
                      {tab.count}
                    </span>
                  )}
                </button>
              );
            })}
          </div>
        </div>

        {/* ── EVENT STREAM ─────────────────────────────────────────────────── */}
        <div style={{ padding: '24px' }}>
          {isLoading && events.length === 0 ? (
            <div style={{ padding: '40px', textAlign: 'center' }}>
              <div style={{ fontSize: '24px', marginBottom: '8px' }}>⏳</div>
              <div style={{ fontSize: '13.5px', fontWeight: 700, color: '#334155' }}>Loading activity records...</div>
            </div>
          ) : filteredEvents.length === 0 ? (
            <div style={{ padding: '48px', textAlign: 'center', background: '#F8FAFC', borderRadius: '10px', border: '1px dashed #CBD5E1' }}>
              <div style={{ fontSize: '28px', marginBottom: '8px' }}>📋</div>
              <div style={{ fontSize: '14px', fontWeight: 800, color: '#0F172A' }}>No activity in this category</div>
              <p style={{ fontSize: '12.5px', color: '#64748B', margin: '4px 0 12px 0' }}>
                {selectedFilter === 'ALL'
                  ? 'Operational events, customer emails, and AI extractions will appear here as this RFQ progresses.'
                  : `No ${selectedFilter.toLowerCase()} events recorded yet for this RFQ.`}
              </p>
              {selectedFilter !== 'ALL' && (
                <button
                  onClick={() => setSelectedFilter('ALL')}
                  style={{
                    background: '#FFFFFF',
                    color: '#4F46E5',
                    border: '1px solid #C7D2FE',
                    borderRadius: '6px',
                    padding: '6px 12px',
                    fontSize: '12px',
                    fontWeight: 700,
                    cursor: 'pointer',
                  }}
                >
                  View All Activity
                </button>
              )}
            </div>
          ) : (
            <div style={{ position: 'relative', paddingLeft: '28px' }}>
              {/* Vertical Guide Line */}
              <div
                style={{
                  position: 'absolute',
                  top: '12px',
                  bottom: '12px',
                  left: '12px',
                  width: '2px',
                  background: '#E2E8F0',
                }}
              />

              {/* Event Cards */}
              <div style={{ display: 'flex', flexDirection: 'column', gap: '18px' }}>
                {filteredEvents.map((ev, idx) => {
                  const cfg = CATEGORY_CONFIG[ev.category] || CATEGORY_CONFIG.SYSTEM;
                  const eventDate = ev.timestamp ? new Date(ev.timestamp) : null;
                  const formattedTime = eventDate ? eventDate.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' }) : '';
                  const formattedDate = eventDate ? eventDate.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' }) : '';

                  return (
                    <div key={ev.id || idx} style={{ position: 'relative', display: 'flex', alignItems: 'flex-start', gap: '16px' }}>
                      
                      {/* Node Circle */}
                      <div
                        style={{
                          position: 'absolute',
                          left: '-28px',
                          top: '2px',
                          width: '26px',
                          height: '26px',
                          borderRadius: '50%',
                          background: cfg.bg,
                          border: `2px solid ${cfg.border}`,
                          display: 'flex',
                          alignItems: 'center',
                          justifyContent: 'center',
                          fontSize: '12px',
                          zIndex: 2,
                          boxShadow: '0 1px 2px rgba(0,0,0,0.04)',
                        }}
                      >
                        {cfg.icon}
                      </div>

                      {/* Event Card Body */}
                      <div
                        style={{
                          flex: 1,
                          background: ev.is_important ? '#FFFFFF' : '#F8FAFC',
                          border: `1px solid ${ev.is_important ? '#CBD5E1' : '#F1F5F9'}`,
                          borderRadius: '12px',
                          padding: '14px 18px',
                          boxShadow: ev.is_important ? '0 2px 6px rgba(15, 23, 42, 0.04)' : 'none',
                          position: 'relative',
                          overflow: 'hidden',
                        }}
                      >
                        {/* Left accent border on important events */}
                        {ev.is_important && (
                          <div style={{ position: 'absolute', top: 0, bottom: 0, left: 0, width: '3px', background: cfg.border }} />
                        )}

                        {/* Top row: Category tag + Title + Date/Time */}
                        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: '6px', flexWrap: 'wrap', gap: '8px' }}>
                          <div style={{ display: 'flex', alignItems: 'center', gap: '8px', flexWrap: 'wrap' }}>
                            <span
                              style={{
                                fontSize: '10.5px',
                                fontWeight: 800,
                                background: cfg.bg,
                                color: cfg.text,
                                border: `1px solid ${cfg.border}`,
                                borderRadius: '4px',
                                padding: '1px 6px',
                                textTransform: 'uppercase',
                                letterSpacing: '0.04em',
                              }}
                            >
                              {cfg.label}
                            </span>
                            <strong style={{ fontSize: '13.5px', fontWeight: 800, color: '#0F172A' }}>
                              {ev.title}
                            </strong>
                            {ev.requires_action && (
                              <span style={{ fontSize: '10.5px', fontWeight: 800, color: '#B91C1C', background: '#FEE2E2', padding: '1px 6px', borderRadius: '4px' }}>
                                Action Needed
                              </span>
                            )}
                          </div>

                          <div style={{ fontSize: '11.5px', color: '#64748B', fontWeight: 600 }}>
                            {formattedDate} {formattedTime && `• ${formattedTime}`}
                          </div>
                        </div>

                        {/* Description */}
                        <p style={{ fontSize: '12.5px', color: '#334155', margin: '0 0 8px 0', lineHeight: 1.5 }}>
                          {ev.description}
                        </p>

                        {/* Bottom Row: Actor & Source Context Actions */}
                        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', flexWrap: 'wrap', gap: '8px', paddingTop: '6px', borderTop: '1px solid #F1F5F9' }}>
                          <div style={{ fontSize: '11.5px', color: '#64748B' }}>
                            Actor: <strong style={{ color: '#334155', fontWeight: 700 }}>{ev.actor_name || 'System'}</strong>
                          </div>

                          {/* Source Link */}
                          {ev.source_type === 'LEAD' && leadId && (
                            <button
                              onClick={() => navigate(`/dashboard/leads?leadId=${leadId}&tab=emails`)}
                              style={{
                                background: 'none',
                                border: 'none',
                                color: '#2563EB',
                                fontSize: '11.5px',
                                fontWeight: 700,
                                cursor: 'pointer',
                                padding: 0,
                                display: 'inline-flex',
                                alignItems: 'center',
                                gap: '3px',
                              }}
                            >
                              <span>View Source Lead #{leadId}</span>
                              <span>→</span>
                            </button>
                          )}
                        </div>

                      </div>
                    </div>
                  );
                })}
              </div>
            </div>
          )}
        </div>
      </div>

    </div>
  );
}

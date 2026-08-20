import React from 'react';
import { STAGE_CONFIG } from './constants';
import StatusBadge from '../../../components/dashboard/StatusBadge';

/**
 * RFQList — displays a rich table of RFQs with customer name, health score,
 * mode/cargo, and quick-action buttons.
 *
 * Props:
 *  rfqs       — Array of RFQ objects from the backend (now includes customer_name)
 *  isLoading  — Boolean: show skeleton rows while loading
 *  onRowClick — Callback when a row is clicked (opens RFQ detail)
 *  onStageFilter — Optional: callback when a filter tab is clicked
 *  filterStage   — Optional: currently active filter stage
 */
export default function RFQList({ rfqs, isLoading, onRowClick, onNewRFQ }) {
  // ── Loading skeleton ────────────────────────────────────────────────────────
  if (isLoading) {
    return (
      <div className="table-container">
        <table className="data-table">
          <thead>
            <tr>
              <th>RFQ #</th>
              <th>Customer</th>
              <th>Origin → Destination</th>
              <th>Mode / Incoterms</th>
              <th>Stage</th>
              <th>Health</th>
              <th>Target Date</th>
            </tr>
          </thead>
          <tbody>
            {[1, 2, 3, 4].map((i) => (
              <tr key={i} className="skeleton-row">
                <td><div className="skeleton-box" style={{ width: '110px' }} /></td>
                <td><div className="skeleton-box" style={{ width: '140px' }} /></td>
                <td><div className="skeleton-box" style={{ width: '160px' }} /></td>
                <td><div className="skeleton-box" style={{ width: '80px' }} /></td>
                <td><div className="skeleton-box" style={{ width: '110px', borderRadius: '12px' }} /></td>
                <td><div className="skeleton-box" style={{ width: '70px' }} /></td>
                <td><div className="skeleton-box" style={{ width: '90px' }} /></td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    );
  }

  // ── Enterprise Empty state with Workflow Guide ─────────────────────────────
  if (!rfqs || rfqs.length === 0) {
    return (
      <div className="rfq-empty-workspace" style={{
        background: '#FFFFFF',
        border: '1px solid #E2E8F0',
        borderRadius: '16px',
        padding: '48px 32px',
        textAlign: 'center',
        boxShadow: '0 1px 3px rgba(15, 23, 42, 0.03)',
        marginTop: '16px',
      }}>
        <div style={{
          width: '56px',
          height: '56px',
          borderRadius: '14px',
          background: '#EFF6FF',
          border: '1px solid #DBEAFE',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          fontSize: '1.6rem',
          margin: '0 auto 16px auto',
        }}>
          🚢
        </div>
        <h3 style={{ fontSize: '1.2rem', fontWeight: 800, color: '#0F172A', marginBottom: '8px' }}>
          No shipment requests yet
        </h3>
        <p style={{ fontSize: '0.85rem', color: '#64748B', maxWidth: '460px', margin: '0 auto 24px auto', lineHeight: 1.5 }}>
          Create your first RFQ to start sourcing competitive carrier rates and pricing options for your customer.
        </p>

        <div style={{ display: 'flex', gap: '12px', justifyContent: 'center', flexWrap: 'wrap', marginBottom: '32px' }}>
          <button
            onClick={onNewRFQ}
            style={{
              background: 'linear-gradient(135deg, #2563EB 0%, #4F46E5 100%)',
              color: '#FFFFFF',
              border: 'none',
              borderRadius: '9px',
              padding: '10px 22px',
              fontSize: '0.82rem',
              fontWeight: 700,
              cursor: 'pointer',
              boxShadow: '0 2px 8px rgba(37, 99, 235, 0.25)',
            }}
          >
            + Create RFQ
          </button>
          <button
            onClick={onNewRFQ}
            style={{
              background: '#FFFFFF',
              color: '#2563EB',
              border: '1px solid #BFDBFE',
              borderRadius: '9px',
              padding: '10px 20px',
              fontSize: '0.82rem',
              fontWeight: 600,
              cursor: 'pointer',
            }}
          >
            Extract RFQ from Customer Email
          </button>
        </div>

        {/* Workflow steps visual */}
        <div style={{
          borderTop: '1px solid #F1F5F9',
          paddingTop: '24px',
          maxWidth: '680px',
          margin: '0 auto',
        }}>
          <div style={{ fontSize: '0.72rem', fontWeight: 700, textTransform: 'uppercase', letterSpacing: '0.06em', color: '#94A3B8', marginBottom: '14px' }}>
            LogisticsHQ RFQ Workflow
          </div>
          <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', gap: '8px', flexWrap: 'wrap' }}>
            {[
              { num: '1', title: 'Customer Request', desc: 'Inquiry / Email' },
              { num: '2', title: 'RFQ Creation', desc: 'Cargo & Ports' },
              { num: '3', title: 'Carrier Rates', desc: 'Spot & Contract' },
              { num: '4', title: 'Compare Quotes', desc: 'AI Margin Rules' },
              { num: '5', title: 'Customer Quote', desc: 'Win & Booking' },
            ].map((step, idx, arr) => (
              <React.Fragment key={step.num}>
                <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', minWidth: '95px' }}>
                  <div style={{
                    width: '24px',
                    height: '24px',
                    borderRadius: '50%',
                    background: '#F1F5F9',
                    color: '#2563EB',
                    fontSize: '0.7rem',
                    fontWeight: 800,
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    marginBottom: '4px',
                  }}>
                    {step.num}
                  </div>
                  <span style={{ fontSize: '0.73rem', fontWeight: 700, color: '#1E293B' }}>{step.title}</span>
                  <span style={{ fontSize: '0.64rem', color: '#64748B' }}>{step.desc}</span>
                </div>
                {idx < arr.length - 1 && (
                  <span style={{ color: '#CBD5E1', fontSize: '0.85rem', fontWeight: 700 }}>→</span>
                )}
              </React.Fragment>
            ))}
          </div>
        </div>
      </div>
    );
  }

  // ── Render table ────────────────────────────────────────────────────────────
  return (
    <div className="table-container">
      <table className="data-table">
        <thead>
          <tr>
            <th>RFQ #</th>
            <th>Customer</th>
            <th>Origin → Destination</th>
            <th>Mode / Incoterms</th>
            <th>Stage</th>
            <th>Health</th>
            <th>Target Date</th>
          </tr>
        </thead>
        <tbody>
          {rfqs.map((rfq) => {
            const stageConfig = STAGE_CONFIG[rfq.stage] || { label: rfq.stage, color: 'gray' };
            const route =
              rfq.origin && rfq.destination
                ? `${rfq.origin} → ${rfq.destination}`
                : 'Route not specified';

            // Health score colour: ≥80 green, ≥50 amber, <50 red
            const healthColor =
              rfq.health_score >= 80
                ? '#10B981'
                : rfq.health_score >= 50
                ? '#F59E0B'
                : '#EF4444';

            // Customer display: use company name if available, else ID
            const customerDisplay = rfq.customer_name
              ? rfq.customer_name
              : rfq.customer_id
              ? `Customer #${rfq.customer_id}`
              : '—';

            return (
              <tr
                key={rfq.id}
                onClick={() => onRowClick(rfq)}
                style={{ cursor: 'pointer' }}
              >
                {/* RFQ Number — monospace, teal for visual scanning */}
                <td>
                  <span style={{ fontFamily: 'monospace', fontWeight: 700, color: 'var(--sky)', fontSize: '0.82rem' }}>
                    {rfq.rfq_number}
                  </span>
                </td>

                {/* Customer Name (from JOIN, not just ID) */}
                <td style={{ fontWeight: 500 }}>{customerDisplay}</td>

                {/* Trade Lane */}
                <td style={{ fontSize: '0.85rem' }}>{route}</td>

                {/* Mode / Incoterms */}
                <td style={{ fontSize: '0.8rem', color: 'var(--slate-400)' }}>
                  {rfq.incoterms || '—'}
                </td>

                {/* Stage badge */}
                <td>
                  <StatusBadge status={stageConfig.label} color={stageConfig.color} />
                </td>

                {/* Health score bar — visual indicator of RFQ completeness */}
                <td>
                  {rfq.health_score !== undefined ? (
                    <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                      <div
                        style={{
                          height: '5px',
                          width: '60px',
                          background: 'rgba(255,255,255,.08)',
                          borderRadius: '3px',
                          overflow: 'hidden',
                        }}
                      >
                        <div
                          style={{
                            height: '100%',
                            width: `${rfq.health_score}%`,
                            background: healthColor,
                            borderRadius: '3px',
                            transition: 'width 0.4s ease',
                          }}
                        />
                      </div>
                      <span style={{ fontSize: '0.78rem', color: 'var(--slate-400)', minWidth: '28px' }}>
                        {rfq.health_score}
                      </span>
                    </div>
                  ) : (
                    '—'
                  )}
                </td>

                {/* Target Date */}
                <td style={{ fontSize: '0.82rem', color: 'var(--slate-400)' }}>
                  {rfq.target_date
                    ? new Date(rfq.target_date).toLocaleDateString('en-IN', {
                        day: '2-digit',
                        month: 'short',
                        year: 'numeric',
                      })
                    : '—'}
                </td>
              </tr>
            );
          })}
        </tbody>
      </table>
    </div>
  );
}

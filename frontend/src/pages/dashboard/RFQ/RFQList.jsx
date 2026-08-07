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
export default function RFQList({ rfqs, isLoading, onRowClick }) {
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

  // ── Empty state ─────────────────────────────────────────────────────────────
  if (!rfqs || rfqs.length === 0) {
    return (
      <div className="empty-state">
        <div className="empty-state-icon">🚢</div>
        <h3>No RFQs found</h3>
        <p>Get started by extracting a new shipment request from a customer email or creating one manually.</p>
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

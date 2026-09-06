import React from 'react';
import { useNavigate } from 'react-router-dom';
import { calculateRFQCompleteness } from './utils/completeness';
import RFQStatusBadge from './components/RFQStatusBadge';
import { ArrowRight, Ship, Plane, Truck, Tag, ExternalLink, Clock, Building2, MapPin, CheckCircle2, AlertCircle, Plus, Zap, Package, FileText } from 'lucide-react';
import ModuleHeroEmptyState from '../../../components/dashboard/ModuleHeroEmptyState';

export default function RFQList({ rfqs, isLoading, onRowClick, onNewRFQ }) {
  const navigate = useNavigate();

  const handleRowNavigate = (rfq) => {
    if (onRowClick) {
      onRowClick(rfq);
    }
    navigate(`/dashboard/rfqs/${rfq.id}`);
  };

  // Helper to parse port name and code
  const formatPort = (val) => {
    if (!val) return { name: 'TBD', code: '' };
    const parts = val.split('(');
    if (parts.length > 1) {
      return {
        name: parts[0].trim(),
        code: parts[1].replace(')', '').trim(),
      };
    }
    return { name: val.trim(), code: '' };
  };

  // ── Loading Skeleton ─────────────────────────────────────────────────────────
  if (isLoading) {
    return (
      <div className="rfq-table-wrapper">
        <table className="rfq-modern-table">
          <thead>
            <tr>
              <th className="rfq-th-id">RFQ #</th>
              <th className="rfq-th-customer">Customer</th>
              <th className="rfq-th-route">Route (POL → POD)</th>
              <th className="rfq-th-mode">Mode / Incoterms</th>
              <th className="rfq-th-status">Status</th>
              <th className="rfq-th-completeness">Completeness</th>
              <th className="rfq-th-updated">Updated</th>
              <th className="rfq-th-actions" style={{ textAlign: 'right' }}>Actions</th>
            </tr>
          </thead>
          <tbody>
            {[1, 2, 3, 4, 5, 6].map((i) => (
              <tr key={i} className="rfq-skeleton-row skeleton-row">
                <td>
                  <div className="skeleton-box" style={{ width: '130px', height: '14px' }} />
                  <div className="skeleton-box" style={{ width: '75px', height: '10px', marginTop: '6px' }} />
                </td>
                <td>
                  <div className="skeleton-box" style={{ width: '150px', height: '14px' }} />
                  <div className="skeleton-box" style={{ width: '100px', height: '10px', marginTop: '6px' }} />
                </td>
                <td><div className="skeleton-box" style={{ width: '190px', height: '16px' }} /></td>
                <td><div className="skeleton-box" style={{ width: '90px', height: '18px', borderRadius: '6px' }} /></td>
                <td><div className="skeleton-box" style={{ width: '110px', height: '22px', borderRadius: '12px' }} /></td>
                <td>
                  <div className="skeleton-box" style={{ width: '70px', height: '12px' }} />
                  <div className="skeleton-box" style={{ width: '100%', height: '4px', marginTop: '6px' }} />
                </td>
                <td><div className="skeleton-box" style={{ width: '75px', height: '12px' }} /></td>
                <td style={{ textAlign: 'right' }}>
                  <div className="skeleton-box" style={{ width: '64px', height: '26px', borderRadius: '6px', marginLeft: 'auto' }} />
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    );
  }

  // ── Purposeful Hero Empty State ─────────────────────────────────────────────
  if (!rfqs || rfqs.length === 0) {
    return (
      <ModuleHeroEmptyState
        icon={<Ship size={28} />}
        badgeTheme="indigo"
        title="No Active RFQs or Rate Inquiries"
        description="Capture shipper cargo specifications, container types, and trade lane routes to generate commercial rate quotations and carrier bookings."
        primaryAction={{
          label: 'Create First RFQ',
          icon: <Plus size={15} />,
          onClick: onNewRFQ,
        }}
        secondaryAction={{
          label: 'Convert from Leads',
          icon: <ArrowRight size={15} />,
          onClick: () => navigate('/dashboard/leads'),
        }}
        features={[
          {
            icon: <Package size={18} />,
            iconBg: '#eff6ff',
            iconColor: '#2563eb',
            title: 'Multi-Modal Route Builder',
            desc: 'Define port pairs (POL/POD), transit times, container types (FCL/LCL), and Incoterms.',
          },
          {
            icon: <Zap size={18} />,
            iconBg: '#ecfdf5',
            iconColor: '#059669',
            title: 'Instant Rate Matching',
            desc: 'Automatically cross-reference contracted carrier buy rates and generate sell margins.',
          },
          {
            icon: <FileText size={18} />,
            iconBg: '#f5f3ff',
            iconColor: '#7c3aed',
            title: 'Direct Quotation Handoff',
            desc: 'Convert completed RFQs into customer-ready quotes with one click.',
          },
        ]}
      />
    );
  }

  // ── Render Enhanced RFQ Table ────────────────────────────────────────────────
  return (
    <div className="rfq-table-wrapper">
      <table className="rfq-modern-table">
        <thead>
          <tr>
            <th className="rfq-th-id">
              <div className="rfq-th-content">RFQ #</div>
            </th>
            <th className="rfq-th-customer">
              <div className="rfq-th-content">Customer</div>
            </th>
            <th className="rfq-th-route">
              <div className="rfq-th-content">Route (POL → POD)</div>
            </th>
            <th className="rfq-th-mode">
              <div className="rfq-th-content">Mode / Terms</div>
            </th>
            <th className="rfq-th-status">
              <div className="rfq-th-content">Status</div>
            </th>
            <th className="rfq-th-completeness">
              <div className="rfq-th-content">Completeness</div>
            </th>
            <th className="rfq-th-updated">
              <div className="rfq-th-content">Updated</div>
            </th>
            <th className="rfq-th-actions" style={{ textAlign: 'right' }}>
              <div className="rfq-th-content" style={{ justifyContent: 'flex-end' }}>Actions</div>
            </th>
          </tr>
        </thead>
        <tbody>
          {rfqs.map((rfq) => {
            const completeness = calculateRFQCompleteness(rfq);
            const isComplete = completeness.isComplete;
            
            // Format Origin and Destination
            const origin = formatPort(rfq.origin);
            const dest = formatPort(rfq.destination);

            // Format Customer
            const customerDisplay = rfq.customer_name || (rfq.customer_id ? `Customer #${rfq.customer_id}` : 'Inbound Customer');
            const contactDisplay = rfq.customer_email || rfq.customer_phone || (rfq.lead_id ? `Lead #${rfq.lead_id}` : null);

            // Derive Transport Mode
            const originUpper = (rfq.origin || '').toUpperCase();
            const destUpper = (rfq.destination || '').toUpperCase();
            const isAir = originUpper.includes('AIRPORT') || destUpper.includes('AIRPORT') || originUpper.startsWith('AIR');
            const ModeIcon = isAir ? Plane : Ship;
            const modeText = isAir ? 'Air' : 'Ocean';

            // Format Created & Updated dates
            const createdDateStr = rfq.created_at
              ? new Date(rfq.created_at).toLocaleDateString('en-US', { day: 'numeric', month: 'short', year: 'numeric' })
              : '—';

            const updatedDate = rfq.updated_at ? new Date(rfq.updated_at) : (rfq.created_at ? new Date(rfq.created_at) : null);
            const updatedDateStr = updatedDate
              ? updatedDate.toLocaleDateString('en-US', { month: 'short', day: 'numeric' })
              : '—';
            const updatedTimeStr = updatedDate
              ? updatedDate.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
              : '';

            return (
              <tr
                key={rfq.id}
                onClick={() => handleRowNavigate(rfq)}
                className="rfq-row"
              >
                {/* 1. RFQ # */}
                <td className="rfq-cell-id">
                  <div className="rfq-id-pill">
                    {rfq.rfq_number}
                  </div>
                  <div className="rfq-subtext">
                    {createdDateStr}
                  </div>
                </td>

                {/* 2. Customer */}
                <td className="rfq-cell-customer">
                  <div className="rfq-customer-name" title={customerDisplay}>
                    <Building2 className="w-3.5 h-3.5 text-slate-400 flex-shrink-0" />
                    <span className="truncate">{customerDisplay}</span>
                  </div>
                  {contactDisplay && (
                    <div className="rfq-subtext rfq-contact-line truncate" title={contactDisplay}>
                      {contactDisplay}
                    </div>
                  )}
                </td>

                {/* 3. Route */}
                <td className="rfq-cell-route">
                  <div className="rfq-route-container">
                    <div className="rfq-route-port origin" title={origin.name}>
                      <span className="rfq-port-name">{origin.name}</span>
                      {origin.code && <span className="rfq-port-code">{origin.code}</span>}
                    </div>

                    <div className="rfq-route-arrow-wrapper">
                      <ArrowRight className="w-3.5 h-3.5 text-slate-400" />
                    </div>

                    <div className="rfq-route-port dest" title={dest.name}>
                      <span className="rfq-port-name">{dest.name}</span>
                      {dest.code && <span className="rfq-port-code">{dest.code}</span>}
                    </div>
                  </div>
                </td>

                {/* 4. Mode / Incoterms */}
                <td className="rfq-cell-mode">
                  <div className="rfq-mode-wrap">
                    <span className="rfq-incoterm-badge">
                      <Tag className="w-2.5 h-2.5 text-amber-600" />
                      <span>{rfq.incoterms || 'FOB'}</span>
                    </span>
                    <span className="rfq-mode-badge">
                      <ModeIcon className="w-3 h-3 text-slate-500" />
                      <span>{modeText}</span>
                    </span>
                  </div>
                </td>

                {/* 5. Status Badge */}
                <td className="rfq-cell-status">
                  <RFQStatusBadge
                    label={completeness.statusLabel}
                    color={completeness.statusColor}
                    size="small"
                  />
                </td>

                {/* 6. Information Completeness */}
                <td className="rfq-cell-completeness">
                  <div className="rfq-completeness-wrap">
                    <div className={`rfq-completeness-label ${isComplete ? 'complete' : 'pending'}`}>
                      <span>{completeness.completedCount}/7 Parameters</span>
                      <span>{isComplete ? '100%' : `${completeness.percentage}%`}</span>
                    </div>
                    <div className="rfq-completeness-bar-track">
                      <div
                        className={`rfq-completeness-bar-fill ${isComplete ? 'complete' : 'pending'}`}
                        style={{ width: `${completeness.percentage}%` }}
                      />
                    </div>
                  </div>
                </td>

                {/* 7. Updated */}
                <td className="rfq-cell-updated">
                  <div className="rfq-updated-date">{updatedDateStr}</div>
                  <div className="rfq-subtext">{updatedTimeStr}</div>
                </td>

                {/* 8. Action Button */}
                <td className="rfq-cell-actions" style={{ textAlign: 'right' }}>
                  <button
                    onClick={(e) => {
                      e.stopPropagation();
                      handleRowNavigate(rfq);
                    }}
                    className="rfq-view-btn"
                  >
                    <span>View</span>
                    <ArrowRight className="w-3 h-3" />
                  </button>
                </td>
              </tr>
            );
          })}
        </tbody>
      </table>
    </div>
  );
}

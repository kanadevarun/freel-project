import React, { useState } from 'react';
import { 
  AlertCircle, CheckCircle2, Layers, FileText, 
  Trophy, ArrowRight, Sparkles, HelpCircle, ChevronDown, 
  ChevronUp, Check, Filter, Info
} from 'lucide-react';
import { calculateRFQCompleteness } from '../utils/completeness';
import './RFQStatusLegend.css';

const STAGE_DEFINITIONS = [
  {
    id: 'INFO_REQUIRED',
    matchKeys: ['INFORMATION_REQUIRED', 'DRAFT', 'STAGE_RFQ_CREATED'],
    title: 'Information Required',
    tag: 'Step 1 • Intake',
    icon: AlertCircle,
    color: '#d97706',
    bg: '#fffbeb',
    border: '#fde68a',
    glow: 'rgba(245, 158, 11, 0.15)',
    dot: '#f59e0b',
    desc: 'Missing mandatory port-pair, cargo weight/volume, or ready date details.',
    action: 'Complete 7/7 Cargo Parameters',
    filterValue: 'DRAFT',
  },
  {
    id: 'READY_FOR_QUOTE',
    matchKeys: ['READY_FOR_QUOTATION'],
    title: 'Ready for Quotation',
    tag: 'Step 2 • Verified',
    icon: CheckCircle2,
    color: '#059669',
    bg: '#ecfdf5',
    border: '#a7f3d0',
    glow: 'rgba(16, 185, 129, 0.15)',
    dot: '#10b981',
    desc: 'All 7 mandatory parameters verified. Ready for carrier tariff matching.',
    action: 'Assign Carrier Buy Rates',
    filterValue: 'ALL',
  },
  {
    id: 'PRICING_ASSIGNED',
    matchKeys: ['PRICING_ASSIGNED', 'STAGE_PRICING_ASSIGNED'],
    title: 'Pricing Assigned',
    tag: 'Step 3 • Costing',
    icon: Layers,
    color: '#2563eb',
    bg: '#eff6ff',
    border: '#bfdbfe',
    glow: 'rgba(37, 99, 235, 0.15)',
    dot: '#3b82f6',
    desc: 'Internal buy cost and ocean/air route options mapped from carrier lines.',
    action: 'Review Sell Margin & Generate Quote',
    filterValue: 'AWAITING_QUOTE',
  },
  {
    id: 'QUOTE_GENERATED',
    matchKeys: ['QUOTE_GENERATED', 'STAGE_QUOTE_GENERATED', 'QUOTE_SENT', 'STAGE_QUOTE_SENT'],
    title: 'Quote Generated',
    tag: 'Step 4 • Commercial',
    icon: FileText,
    color: '#7c3aed',
    bg: '#f5f3ff',
    border: '#ddd6fe',
    glow: 'rgba(124, 58, 237, 0.15)',
    dot: '#8b5cf6',
    desc: 'Commercial freight quote generated with markup. Ready for customer review.',
    action: 'Send to Customer & Track Decision',
    filterValue: 'AWAITING_QUOTE',
  },
  {
    id: 'WON',
    matchKeys: ['WON', 'STAGE_WON', 'SHIPMENT_CREATED', 'STAGE_SHIPMENT_CREATED'],
    title: 'Won / Awarded',
    tag: 'Step 5 • Converted',
    icon: Trophy,
    color: '#16a34a',
    bg: '#f0fdf4',
    border: '#bbf7d0',
    glow: 'rgba(22, 163, 74, 0.15)',
    dot: '#22c55e',
    desc: 'Quote approved by shipper. Ready for space allocation booking.',
    action: 'Create Space Booking',
    filterValue: 'WON',
  },
];

export default function RFQStatusLegend({ 
  rfqs = [], 
  activeTab = 'ALL', 
  onSelectTab 
}) {
  const [isCollapsed, setIsCollapsed] = useState(() => {
    return localStorage.getItem('freel_rfq_legend_collapsed') === 'true';
  });

  const toggleCollapse = () => {
    setIsCollapsed(prev => {
      const next = !prev;
      localStorage.setItem('freel_rfq_legend_collapsed', String(next));
      return next;
    });
  };

  // Compute live counts for each stage
  const stageCounts = React.useMemo(() => {
    const counts = {
      INFO_REQUIRED: 0,
      READY_FOR_QUOTE: 0,
      PRICING_ASSIGNED: 0,
      QUOTE_GENERATED: 0,
      WON: 0,
    };

    rfqs.forEach(rfq => {
      const completeness = calculateRFQCompleteness(rfq);
      const stage = rfq.stage || 'STAGE_RFQ_CREATED';

      if (stage === 'STAGE_WON' || stage === 'STAGE_SHIPMENT_CREATED') {
        counts.WON++;
      } else if (stage === 'STAGE_QUOTE_GENERATED' || stage === 'STAGE_QUOTE_SENT') {
        counts.QUOTE_GENERATED++;
      } else if (stage === 'STAGE_PRICING_ASSIGNED') {
        counts.PRICING_ASSIGNED++;
      } else if (completeness.isComplete) {
        counts.READY_FOR_QUOTE++;
      } else {
        counts.INFO_REQUIRED++;
      }
    });

    return counts;
  }, [rfqs]);

  return (
    <div className={`rfq-legend-card ${isCollapsed ? 'collapsed' : ''}`}>
      {/* Header Bar */}
      <div className="rfq-legend-header" onClick={toggleCollapse} role="button" tabIndex={0}>
        <div className="rfq-legend-header-left">
          <div className="rfq-legend-icon-badge">
            <Sparkles size={16} />
          </div>
          <div className="rfq-legend-header-text">
            <div className="rfq-legend-title-row">
              <span className="rfq-legend-main-title">RFQ Workflow Lifecycle & Stage Intelligence</span>
              <span className="rfq-legend-count-pill">{rfqs.length} Total Inquiries</span>
            </div>
            <p className="rfq-legend-main-subtitle">
              Interactive stage glossary tracing inquiry intake, completeness validation, carrier tariff matching, and quote awards.
            </p>
          </div>
        </div>

        <div className="rfq-legend-header-right">
          <button 
            type="button" 
            className="rfq-legend-toggle-btn"
            onClick={(e) => {
              e.stopPropagation();
              toggleCollapse();
            }}
            aria-label={isCollapsed ? 'Expand Legend' : 'Collapse Legend'}
          >
            <span className="toggle-label">{isCollapsed ? 'Show Stage Guide' : 'Hide Guide'}</span>
            {isCollapsed ? <ChevronDown size={15} /> : <ChevronUp size={15} />}
          </button>
        </div>
      </div>

      {/* Expanded Content Body */}
      {!isCollapsed && (
        <div className="rfq-legend-body animate-fade-in">
          {/* Visual Progression Line */}
          <div className="rfq-stages-track">
            {STAGE_DEFINITIONS.map((stage, idx) => {
              const count = stageCounts[stage.id] || 0;
              const isTabActive = activeTab === stage.filterValue && stage.filterValue !== 'ALL';

              return (
                <React.Fragment key={stage.id}>
                  <div 
                    className={`rfq-stage-item ${isTabActive ? 'active-filter' : ''}`}
                    onClick={() => onSelectTab && onSelectTab(stage.filterValue)}
                    style={{
                      '--stage-accent': stage.color,
                      '--stage-bg': stage.bg,
                      '--stage-border': stage.border,
                      '--stage-glow': stage.glow,
                    }}
                  >
                    {/* Stage Header */}
                    <div className="stage-item-top">
                      <span className="stage-step-tag" style={{ color: stage.color, background: stage.bg, borderColor: stage.border }}>
                        {stage.tag}
                      </span>
                      <span className={`stage-count-badge ${count > 0 ? 'has-items' : ''}`} style={{ color: stage.color, borderColor: stage.border }}>
                        {count} {count === 1 ? 'inquiry' : 'inquiries'}
                      </span>
                    </div>

                    {/* Stage Identity */}
                    <div className="stage-item-identity">
                      <div className="stage-icon-wrap" style={{ background: stage.bg, color: stage.color, borderColor: stage.border }}>
                        <stage.icon size={16} />
                      </div>
                      <span className="stage-title-text" style={{ color: '#0f172a' }}>
                        {stage.title}
                      </span>
                    </div>

                    {/* Stage Description */}
                    <p className="stage-desc-text">{stage.desc}</p>

                    {/* Stage Action Target */}
                    <div className="stage-action-target">
                      <span className="action-text">{stage.action}</span>
                    </div>
                  </div>

                  {/* Connecting Arrow */}
                  {idx < STAGE_DEFINITIONS.length - 1 && (
                    <div className="stage-flow-arrow">
                      <ArrowRight size={16} />
                    </div>
                  )}
                </React.Fragment>
              );
            })}
          </div>

          {/* Helper Footer */}
          <div className="rfq-legend-footer">
            <div className="legend-footer-tip">
              <Info size={14} className="text-blue-500 flex-shrink-0" />
              <span>
                <strong>Operations Tip:</strong> RFQs with 100% parameter completeness unlock instant carrier spot rate matching and convert 3.2x faster into customer bookings.
              </span>
            </div>
            {activeTab !== 'ALL' && (
              <button 
                type="button" 
                className="btn-clear-stage-filter"
                onClick={() => onSelectTab && onSelectTab('ALL')}
              >
                <span>Reset to All RFQs</span>
              </button>
            )}
          </div>
        </div>
      )}
    </div>
  );
}

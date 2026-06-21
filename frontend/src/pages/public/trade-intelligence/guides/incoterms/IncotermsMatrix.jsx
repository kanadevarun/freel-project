import React, { useState } from 'react';
import { 
  Scale,
  Check,
  X,
  AlertTriangle,
  Info,
  Star,
  Rocket,
  Globe,
  DollarSign,
  FileText,
  Anchor,
  Ship,
  Truck,
  PackageCheck
} from 'lucide-react';
import { 
  FactoryIcon, 
  ContainerShipIcon, 
  WarehouseIcon 
} from './Icons'; // Reusing our rich SVGs for watermarks and header icons
import './IncotermsMatrix.css';

/* ── DATA ── */
const TERMS = [
  {
    id: 'EXW',
    name: 'EXW',
    sub: 'Ex Works',
    badge: 'Advanced Users',
    badgeClass: 'imx-badge-orange',
    icon: FactoryIcon
  },
  {
    id: 'FOB',
    name: 'FOB',
    sub: 'Free On Board',
    badge: 'Most Popular Export Term',
    badgeClass: 'imx-badge-blue',
    icon: Anchor
  },
  {
    id: 'CIF',
    name: 'CIF',
    sub: 'Cost, Insurance & Freight',
    badge: null,
    icon: Ship
  },
  {
    id: 'DAP',
    name: 'DAP',
    sub: 'Delivered At Place',
    badge: null,
    icon: Truck
  },
  {
    id: 'DDP',
    name: 'DDP',
    sub: 'Delivered Duty Paid',
    badge: 'Buyer Friendly',
    badgeClass: 'imx-badge-green',
    icon: PackageCheck
  }
];

const ROWS = [
  {
    key: 'export',
    label: 'Seller Handles Export',
    icon: FileText,
    values: ['X', 'C', 'C', 'C', 'C']
  },
  {
    key: 'freight',
    label: 'Seller Pays Freight',
    icon: Truck,
    values: ['X', 'X', 'C', 'C', 'C']
  },
  {
    key: 'insurance',
    label: 'Seller Pays Insurance',
    icon: AlertTriangle, // shield-like
    values: ['X', 'X', 'C', 'X', 'C']
  },
  {
    key: 'duties',
    label: 'Seller Pays Import Duties',
    icon: FileText,
    values: ['X', 'X', 'X', 'X', 'C']
  },
  {
    key: 'transport',
    label: 'Buyer Arranges Transport',
    icon: Truck,
    values: ['C', 'C', 'X', 'X', 'X']
  },
  {
    key: 'riskEarly',
    label: 'Risk Transfer Early',
    icon: AlertTriangle,
    values: ['C', 'X', 'X', 'X', 'X']
  },
  {
    key: 'riskLate',
    label: 'Risk Transfer Late',
    icon: AlertTriangle,
    values: ['X', 'C', 'C', 'C', 'C']
  },
  {
    key: 'exporters',
    label: 'Best For Exporters',
    icon: null,
    type: 'rating',
    values: [1, 4, 3, 3, 2]
  },
  {
    key: 'importers',
    label: 'Best For Importers',
    icon: null,
    type: 'rating',
    values: [1, 3, 4, 4, 5]
  },
  {
    key: 'complexity',
    label: 'Complexity',
    icon: null,
    type: 'text',
    values: [
      { text: 'High', class: 'high' },
      { text: 'Low', class: 'low' },
      { text: 'Medium', class: 'medium' },
      { text: 'Medium', class: 'medium' },
      { text: 'High', class: 'high' }
    ]
  },
  {
    key: 'recommended',
    label: 'Recommended Usage',
    icon: Star,
    type: 'rec',
    values: [
      'Low Cost Shipments',
      'General Sea Shipments',
      'Sea Shipments with Insurance',
      'Multi-modal Shipments',
      'Door-to-Door Convenience'
    ]
  }
];

export default function IncotermsMatrix() {
  const [hoveredCol, setHoveredCol] = useState(null);

  // Helper to render cell content based on row type
  const renderValue = (val, type) => {
    if (type === 'rating') {
      return (
        <div className="imx-rating">
          {[1, 2, 3, 4, 5].map((star) => (
            <Star 
              key={star} 
              size={14} 
              className={star <= val ? 'imx-star-filled' : 'imx-star-empty'} 
            />
          ))}
        </div>
      );
    }
    if (type === 'text') {
      return <span className={`imx-text-val ${val.class}`}>{val.text}</span>;
    }
    if (type === 'rec') {
      return (
        <span className="imx-rec-text">
          {val} <Info size={12} className="imx-info" />
        </span>
      );
    }
    // Default Check/Cross
    if (val === 'C') return <Check size={20} strokeWidth={3} className="imx-check" />;
    if (val === 'X') return <X size={20} strokeWidth={3} className="imx-cross" />;
    if (val === 'W') return <AlertTriangle size={18} strokeWidth={2.5} className="imx-warn" />;
    return null;
  };

  return (
    <section className="imx-section">
      <div className="imx-container">
        
        <div className="imx-header">
          <div className="imx-pill">
            <Scale size={14} />
            INTERACTIVE COMPARISON
          </div>
          <h2 className="imx-title">Compare Before You Choose</h2>
          <p className="imx-subtitle">
            Understand the trade-offs between the most commonly used Incoterms.
          </p>
        </div>

        <div className="imx-table-wrap">
          <div className="imx-table" onMouseLeave={() => setHoveredCol(null)}>
            
            {/* Header Row (Sticky) */}
            <div className="imx-row imx-header-row">
              <div className="imx-cell feature-col">
                Responsibilities & Costs
              </div>
              {TERMS.map((term, i) => {
                const isHovered = hoveredCol === i;
                const isFOB = term.id === 'FOB';
                const colClass = `imx-cell value-col ${isHovered ? 'hovered-col' : ''} ${isFOB ? 'highlight-fob' : ''}`;
                
                const IconComponent = term.icon;

                return (
                  <div 
                    key={term.id} 
                    className={colClass}
                    onMouseEnter={() => setHoveredCol(i)}
                  >
                    <div className="imx-col-icon">
                      <IconComponent size={32} strokeWidth={1.5} color={isFOB ? '#3b82f6' : '#64748b'} />
                    </div>
                    <h3 className="imx-col-title">{term.name}</h3>
                    <span className="imx-col-sub">{term.sub}</span>
                    {term.badge && (
                      <span className={`imx-col-badge ${term.badgeClass}`}>{term.badge}</span>
                    )}
                  </div>
                );
              })}
            </div>

            {/* Data Rows */}
            {ROWS.map((row) => (
              <div key={row.key} className="imx-row">
                <div className="imx-cell feature-col">
                  {row.icon && React.createElement(row.icon, { size: 16, className: 'imx-feature-icon' })}
                  {row.label}
                </div>
                {row.values.map((val, i) => {
                  const isHovered = hoveredCol === i;
                  const isFOB = i === 1; // FOB is index 1
                  const colClass = `imx-cell value-col ${isHovered ? 'hovered-col' : ''} ${isFOB ? 'highlight-fob' : ''}`;
                  return (
                    <div 
                      key={`${row.key}-${i}`} 
                      className={colClass}
                      onMouseEnter={() => setHoveredCol(i)}
                    >
                      {renderValue(val, row.type)}
                    </div>
                  );
                })}
              </div>
            ))}
            
            {/* Bottom Legend inside table wrap */}
            <div className="imx-row">
              <div className="imx-legend" style={{ gridColumn: '1 / -1' }}>
                <div className="imx-legend-items">
                  <div className="imx-legend-item"><Check size={14} className="imx-check"/> Seller Responsible</div>
                  <div className="imx-legend-item"><X size={14} className="imx-cross"/> Buyer Responsible</div>
                  <div className="imx-legend-item"><AlertTriangle size={14} className="imx-warn"/> Depends on Contract</div>
                </div>
                <div>* Ratings are relative and may vary based on specific shipment conditions.</div>
              </div>
            </div>

          </div>
        </div>

        {/* Bottom Summary Cards */}
        <div className="imx-summary-cards">
          
          <div className="imx-sc blue">
            <ContainerShipIcon className="imx-sc-watermark" />
            <div className="imx-sc-icon">
              <Rocket size={32} />
            </div>
            <div className="imx-sc-content">
              <h4 className="imx-sc-title">New Exporters</h4>
              <p className="imx-sc-subtitle">Recommended Incoterm</p>
              <div className="imx-sc-term">FOB</div>
              <p className="imx-sc-desc">Simple, easy to understand and globally accepted. Perfect for most first-time exporters taking control of local origin charges.</p>
            </div>
          </div>

          <div className="imx-sc green">
            <WarehouseIcon className="imx-sc-watermark" />
            <div className="imx-sc-icon">
              <Globe size={32} />
            </div>
            <div className="imx-sc-content">
              <h4 className="imx-sc-title">Global Importers</h4>
              <p className="imx-sc-subtitle">Recommended Incoterm</p>
              <div className="imx-sc-term">DDP</div>
              <p className="imx-sc-desc">Maximum convenience with minimal involvement. The seller handles everything till your doorstep, letting you focus on your core business.</p>
            </div>
          </div>

          <div className="imx-sc orange">
            <FactoryIcon className="imx-sc-watermark" />
            <div className="imx-sc-icon">
              <DollarSign size={32} />
            </div>
            <div className="imx-sc-content">
              <h4 className="imx-sc-title">Lowest Seller Cost</h4>
              <p className="imx-sc-subtitle">Recommended Incoterm</p>
              <div className="imx-sc-term">EXW</div>
              <p className="imx-sc-desc">Seller's responsibility is minimal, making it the lowest cost option for sellers, though it transfers maximum burden and risk to the buyer.</p>
            </div>
          </div>

        </div>

      </div>
    </section>
  );
}

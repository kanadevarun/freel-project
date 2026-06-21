import React, { useState } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { 
  Building2,
  CheckCircle2,
  Info,
  Calendar,
  DollarSign,
  MapPin,
  ArrowRight,
  Star,
  AlertTriangle,
  Ship
} from 'lucide-react';
import { 
  FactoryIcon, 
  DocumentsIcon, 
  OriginPortIcon, 
  ContainerShipIcon, 
  DestPortIcon, 
  CustomsIcon, 
  WarehouseIcon, 
  MiniDonutChart 
} from './Icons';
import './RealShipmentSimulator.css';

/* ── DATA DEFINITIONS ── */

const JOURNEY_NODES = [
  { id: 0, label: 'Factory', icon: FactoryIcon },
  { id: 1, label: 'Export\nClearance', icon: DocumentsIcon },
  { id: 2, label: 'Origin Port\n(Loading)', icon: OriginPortIcon },
  { id: 3, label: 'Ocean Freight', icon: ContainerShipIcon },
  { id: 4, label: 'Destination Port\n(Unloading)', icon: DestPortIcon },
  { id: 5, label: 'Import\nClearance', icon: CustomsIcon },
  { id: 6, label: 'Warehouse /\nDelivery', icon: WarehouseIcon }
];

const SIMULATOR_DATA = {
  FOB: {
    code: 'FOB',
    title: 'Free On Board',
    desc: 'Seller delivers when goods are loaded on board the vessel at the origin port.',
    sellerNodes: 3, // 0, 1, 2
    trackFillWidth: '33.33%',
    riskNode: 2,
    riskText: 'At Origin Port',
    riskSubText: 'When goods are loaded on board the vessel',
    sellerHandles: [
      'Goods & packaging',
      'Export clearance',
      'Loading at origin port'
    ],
    buyerHandles: [
      'Ocean freight',
      'Insurance',
      'Import clearance & duty',
      'Delivery to final warehouse'
    ],
    sellerCost: '$1,200',
    sellerCostPercent: 21,
    sellerCostItems: [
      'Packaging',
      'Export Clearance',
      'Loading at Origin Port'
    ],
    buyerCost: '$4,500',
    buyerCostPercent: 79,
    buyerCostItems: [
      'Ocean Freight',
      'Insurance',
      'Import Duty & Taxes',
      'Unloading at Destination',
      'Delivery to Warehouse'
    ],
    insightTitle: 'Why Companies Choose FOB',
    insightDesc: 'Many buyers prefer controlling freight rates themselves and already have carrier contracts in place. FOB gives them flexibility and often lower total costs.'
  },
  CIF: {
    code: 'CIF',
    title: 'Cost, Insurance and Freight',
    desc: 'Seller pays for freight and insurance, but risk transfers at origin.',
    sellerNodes: 5, // 0, 1, 2, 3, 4
    trackFillWidth: '66.66%',
    riskNode: 2,
    riskText: 'At Origin Port',
    riskSubText: 'Risk transfers early, though seller pays freight',
    sellerHandles: [
      'Goods & packaging',
      'Export clearance',
      'Loading at origin port',
      'Ocean freight',
      'Insurance'
    ],
    buyerHandles: [
      'Import clearance & duty',
      'Unloading at destination',
      'Delivery to final warehouse'
    ],
    sellerCost: '$3,800',
    sellerCostPercent: 67,
    sellerCostItems: [
      'Packaging',
      'Export Clearance',
      'Loading',
      'Ocean Freight & Insurance'
    ],
    buyerCost: '$1,900',
    buyerCostPercent: 33,
    buyerCostItems: [
      'Import Duty & Taxes',
      'Unloading at Destination',
      'Delivery to Warehouse'
    ],
    insightTitle: 'Why Companies Choose CIF',
    insightDesc: 'Buyers who lack experience or volume to negotiate good ocean freight rates often rely on the seller to arrange transport and insurance to their local port.'
  },
  DDP: {
    code: 'DDP',
    title: 'Delivered Duty Paid',
    desc: 'Seller is responsible for delivering the goods to the named place, cleared for import.',
    sellerNodes: 7, // All nodes
    trackFillWidth: '100%',
    riskNode: 6,
    riskText: 'At Destination',
    riskSubText: 'When goods arrive ready for unloading',
    sellerHandles: [
      'Goods & packaging',
      'Export clearance',
      'Ocean freight',
      'Insurance',
      'Import clearance & duty',
      'Delivery to final warehouse'
    ],
    buyerHandles: [
      'Receiving goods at destination warehouse',
      'Unloading from final delivery truck'
    ],
    sellerCost: '$5,700',
    sellerCostPercent: 100,
    sellerCostItems: [
      'Packaging',
      'Freight & Insurance',
      'Import Duty & Taxes',
      'Final Delivery'
    ],
    buyerCost: '$0',
    buyerCostPercent: 0,
    buyerCostItems: [
      'No Freight or Duty Costs'
    ],
    insightTitle: 'Why Companies Choose DDP',
    insightDesc: 'DDP offers the ultimate frictionless experience for buyers. It is heavily used in B2B e-commerce where sellers want to guarantee a final landed cost.'
  }
};

/* ─────────────────────────────────────────
   COMPONENT
───────────────────────────────────────── */
export default function RealShipmentSimulator() {
  const [activeTab, setActiveTab] = useState('FOB');
  const data = SIMULATOR_DATA[activeTab];

  // Calculate the risk marker's left position (based on 6 segments total)
  const riskLeftPercent = (data.riskNode / 6) * 100;

  return (
    <section className="rss-section">
      <div className="rss-container">
        
        {/* Header */}
        <div className="rss-header">
          <div className="rss-pill">
            <Ship size={14} className="rss-pill-icon" />
            REAL SHIPMENT SIMULATOR
          </div>
          <h2 className="rss-title">Same Shipment. Completely Different Responsibilities.</h2>
          <p className="rss-subtitle">
            See how the exact same cargo movement changes depending on the Incoterm you choose.
          </p>
        </div>

        {/* Tabs */}
        <div className="rss-tabs">
          {['FOB', 'CIF', 'DDP'].map(tab => (
            <div 
              key={tab}
              className={`rss-tab ${activeTab === tab ? 'active' : ''}`}
              onClick={() => setActiveTab(tab)}
            >
              {tab}
            </div>
          ))}
        </div>

        {/* Main Interface */}
        <div className="rss-layout">
          
          {/* Left / Center Panel */}
          <div className="rss-main">
            
            {/* Shipment Header Info */}
            <div className="rss-shipment-head">
              <div className="rss-sh-icon">
                <Building2 size={28} />
              </div>
              <div className="rss-sh-details">
                <h3 className="rss-sh-title">Shipment #MUM-HAM-2847</h3>
                <div className="rss-sh-meta">
                  <div className="rss-sh-meta-item">
                    <MapPin size={14}/> Mumbai, India <ArrowRight size={12}/> Hamburg, Germany
                  </div>
                  <div className="rss-sh-meta-item">
                    <Calendar size={14}/> 18 Days Transit
                  </div>
                  <div className="rss-sh-meta-item">
                    <DollarSign size={14}/> $12,800 Cargo Value
                  </div>
                </div>
              </div>
            </div>

            {/* Journey Map */}
            <div className="rss-journey">
              <div className="rss-journey-labels">
                <div className="rss-jl-item rss-jl-seller">Seller Responsibility</div>
                <div className="rss-jl-item rss-jl-buyer">Buyer Responsibility</div>
              </div>

              <div className="rss-journey-track">
                {/* Background track line */}
                <div className="rss-track-bg" />
                
                {/* Colored fill lines */}
                <div className="rss-track-fill" style={{ width: data.trackFillWidth }} />
                <div 
                  className="rss-track-fill-buyer" 
                  style={{ width: `calc(100% - ${data.trackFillWidth})` }} 
                />

                {/* Nodes */}
                {JOURNEY_NODES.map((node, idx) => {
                  const isSeller = idx < data.sellerNodes;
                  const IconComponent = node.icon;
                  return (
                    <div key={node.id} className={`rss-j-node ${isSeller ? 'seller' : 'buyer'}`}>
                      <div className="rss-j-icon-wrap">
                        <IconComponent size={20} />
                      </div>
                      <div className="rss-j-label">
                        {node.label.split('\n').map((line, i) => (
                          <React.Fragment key={i}>
                            {line}
                            {i === 0 && node.label.includes('\n') && <br/>}
                          </React.Fragment>
                        ))}
                      </div>
                    </div>
                  );
                })}

                {/* Floating Risk Marker */}
                <div 
                  className="rss-risk-line"
                  style={{ left: `calc(20px + (100% - 40px) * (${riskLeftPercent} / 100))` }}
                />
                <div 
                  className="rss-risk-marker"
                  style={{ left: `calc(20px + (100% - 40px) * (${riskLeftPercent} / 100))` }}
                >
                  <AlertTriangle size={18} className="rss-rm-icon" />
                  <div className="rss-rm-text">
                    <span className="rss-rm-title">Risk Transfers</span>
                    <span className="rss-rm-desc">{data.riskText}</span>
                  </div>
                </div>
              </div>
            </div>

            {/* Responsibility Cards */}
            <div className="rss-resp-cards">
              
              {/* Seller Handles */}
              <div className="rss-rc seller">
                <div style={{ position: 'absolute', right: '-40px', bottom: '-40px', width: '200px', height: '200px', opacity: 0.05, zIndex: 1 }}>
                  <FactoryIcon />
                </div>
                <div className="rss-rc-header">
                  <div className="rss-rc-icon">🏭</div>
                  <h4 className="rss-rc-title">Seller Responsibilities</h4>
                </div>
                <ul className="rss-rc-list">
                  <AnimatePresence mode="popLayout">
                    {data.sellerHandles.map((item, i) => (
                      <motion.li 
                        key={`${activeTab}-sh-${i}`}
                        initial={{ opacity: 0, x: -10 }}
                        animate={{ opacity: 1, x: 0 }}
                        exit={{ opacity: 0 }}
                        transition={{ duration: 0.2, delay: i * 0.05 }}
                        className="rss-rc-item"
                      >
                        <CheckCircle2 size={16} className="rss-rc-check" />
                        {item}
                      </motion.li>
                    ))}
                  </AnimatePresence>
                </ul>
              </div>

              {/* Buyer Handles */}
              <div className="rss-rc buyer">
                <div style={{ position: 'absolute', right: '-40px', bottom: '-40px', width: '200px', height: '200px', opacity: 0.05, zIndex: 1 }}>
                  <WarehouseIcon />
                </div>
                <div className="rss-rc-header">
                  <div className="rss-rc-icon">🏢</div>
                  <h4 className="rss-rc-title">Buyer Responsibilities</h4>
                </div>
                <ul className="rss-rc-list">
                  <AnimatePresence mode="popLayout">
                    {data.buyerHandles.map((item, i) => (
                      <motion.li 
                        key={`${activeTab}-bh-${i}`}
                        initial={{ opacity: 0, x: 10 }}
                        animate={{ opacity: 1, x: 0 }}
                        exit={{ opacity: 0 }}
                        transition={{ duration: 0.2, delay: i * 0.05 }}
                        className="rss-rc-item"
                      >
                        <CheckCircle2 size={16} className="rss-rc-check" />
                        {item}
                      </motion.li>
                    ))}
                  </AnimatePresence>
                </ul>
              </div>

            </div>

            {/* Bottom Insight Footer */}
            <div className="rss-insight">
              <div className="rss-in-content">
                <div className="rss-in-icon"><Star size={24} /></div>
                <div className="rss-in-text">
                  <AnimatePresence mode="wait">
                    <motion.div
                      key={activeTab}
                      initial={{ opacity: 0, y: 10 }}
                      animate={{ opacity: 1, y: 0 }}
                      exit={{ opacity: 0, y: -10 }}
                      transition={{ duration: 0.2 }}
                    >
                      <h4>{data.insightTitle}</h4>
                      <p>{data.insightDesc}</p>
                    </motion.div>
                  </AnimatePresence>
                </div>
              </div>
              <svg viewBox="0 0 200 80" className="rss-in-graphic">
                <defs>
                  <linearGradient id="shipGrad" x1="0%" y1="0%" x2="0%" y2="100%">
                    <stop offset="0%" stopColor="#2563eb" />
                    <stop offset="100%" stopColor="#1e40af" />
                  </linearGradient>
                </defs>
                <path d="M10,60 L180,60 L190,40 L10,40 Z" fill="url(#shipGrad)" />
                <rect x="140" y="20" width="30" height="20" fill="#94a3b8" />
                <rect x="40" y="25" width="20" height="15" fill="#f43f5e" />
                <rect x="65" y="25" width="20" height="15" fill="#10b981" />
                <rect x="90" y="25" width="20" height="15" fill="#f43f5e" />
                <rect x="50" y="10" width="20" height="15" fill="#3b82f6" />
                <rect x="75" y="10" width="20" height="15" fill="#10b981" />
                <path d="M0,65 Q50,75 100,65 T200,65" fill="none" stroke="#bfdbfe" strokeWidth="3" opacity="0.5" />
              </svg>
            </div>

          </div>

          {/* Right Sidebar: Cost Breakdown */}
          <div className="rss-sidebar">
            <h4 className="rss-sb-title">Cost Breakdown</h4>
            
            {/* Seller Cost Card */}
            <div className="rss-cost-card seller">
              <div className="rss-cc-top">
                <div className="rss-cc-info">
                  <div className="rss-cc-label">💰 Seller Cost</div>
                  <div className="rss-cc-value">{data.sellerCost}</div>
                </div>
                <MiniDonutChart percent={data.sellerCostPercent} colorClass="seller" />
              </div>
              <ul className="rss-cc-list">
                <AnimatePresence mode="popLayout">
                  {data.sellerCostItems.map((item, i) => (
                    <motion.li 
                      key={`${activeTab}-sc-${i}`}
                      initial={{ opacity: 0, x: -5 }}
                      animate={{ opacity: 1, x: 0 }}
                      exit={{ opacity: 0 }}
                      className="rss-cc-item"
                    >
                      <CheckCircle2 size={14} className="rss-cc-check" />
                      {item}
                    </motion.li>
                  ))}
                </AnimatePresence>
              </ul>
            </div>

            {/* Buyer Cost Card */}
            <div className="rss-cost-card buyer">
              <div className="rss-cc-top">
                <div className="rss-cc-info">
                  <div className="rss-cc-label">💰 Buyer Cost</div>
                  <div className="rss-cc-value">{data.buyerCost}</div>
                </div>
                <MiniDonutChart percent={data.buyerCostPercent} colorClass="buyer" />
              </div>
              <ul className="rss-cc-list">
                <AnimatePresence mode="popLayout">
                  {data.buyerCostItems.map((item, i) => (
                    <motion.li 
                      key={`${activeTab}-bc-${i}`}
                      initial={{ opacity: 0, x: 5 }}
                      animate={{ opacity: 1, x: 0 }}
                      exit={{ opacity: 0 }}
                      className="rss-cc-item"
                    >
                      <CheckCircle2 size={14} className="rss-cc-check" />
                      {item}
                    </motion.li>
                  ))}
                </AnimatePresence>
              </ul>
            </div>

            <div className="rss-sb-footer">
              <Info size={16} className="rss-sb-footer-icon" />
              Costs are estimated for illustrative purposes based on typical Mumbai to Hamburg rates.
            </div>

          </div>

        </div>
      </div>
    </section>
  );
}

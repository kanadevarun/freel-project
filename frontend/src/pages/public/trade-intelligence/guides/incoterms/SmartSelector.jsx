import React, { useState, useEffect } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { 
  Anchor, 
  CheckCircle2, 
  RefreshCcw, 
  Settings2, 
  Users, 
  Ship,
  Plane,
  Truck,
  Train,
  Package,
  Shield,
  Scale,
  ArrowRight
} from 'lucide-react';
import './SmartSelector.css';

// ── RECOMMENDATION LOGIC & DATA ──

const RECOMMENDATIONS = {
  FOB: {
    code: 'FOB',
    name: 'Free On Board',
    desc: 'Seller delivers when goods are loaded on board the vessel at the origin port.',
    reasons: [
      'You want the buyer to control main freight',
      'You handle export clearance and loading',
      'Popular and widely accepted for sea shipments',
      'Balanced risk and cost distribution'
    ],
    riskPoint: 'At Origin Port',
    riskDesc: 'When goods are loaded on board the vessel',
    riskIcon: <Anchor size={24} strokeWidth={1.5} />,
    sellerSplit: 40,
    buyerSplit: 60,
    insight: 'Used by 42% of global exporters shipping via ocean freight.'
  },
  CFR: {
    code: 'CFR',
    name: 'Cost and Freight',
    desc: 'Seller pays for ocean freight to destination, but risk transfers at origin.',
    reasons: [
      'Seller controls main carriage costs',
      'Buyer handles destination import clearance',
      'Risk transfers early to the buyer',
      'Ideal for bulk ocean cargo'
    ],
    riskPoint: 'At Origin Port',
    riskDesc: 'When goods are loaded on board the vessel',
    riskIcon: <Anchor size={24} strokeWidth={1.5} />,
    sellerSplit: 60,
    buyerSplit: 40,
    insight: 'The 2nd most common term for international sea freight.'
  },
  CIF: {
    code: 'CIF',
    name: 'Cost, Insurance & Freight',
    desc: 'Seller pays freight and insurance to destination, but risk transfers at origin.',
    reasons: [
      'Seller controls freight and insurance',
      'Buyer is protected during transit',
      'Standardized for international letters of credit',
      'Seamless sea transport experience'
    ],
    riskPoint: 'At Origin Port',
    riskDesc: 'Risk transfers at loading, but seller buys insurance',
    riskIcon: <Shield size={24} strokeWidth={1.5} />,
    sellerSplit: 65,
    buyerSplit: 35,
    insight: 'Highly recommended when buyer requires transit insurance.'
  },
  FCA: {
    code: 'FCA',
    name: 'Free Carrier',
    desc: 'Seller delivers goods to the carrier nominated by the buyer.',
    reasons: [
      'Works for any transport mode (Air, Road, Rail)',
      'Buyer has complete control over main freight',
      'Seller only handles origin export',
      'Highly flexible and widely used'
    ],
    riskPoint: 'At Named Place',
    riskDesc: 'When handed over to the first carrier',
    riskIcon: <Package size={24} strokeWidth={1.5} />,
    sellerSplit: 30,
    buyerSplit: 70,
    insight: 'The ICC highly recommends FCA over EXW for international trade.'
  },
  DAP: {
    code: 'DAP',
    name: 'Delivered At Place',
    desc: 'Seller handles everything up to the destination, except import clearance.',
    reasons: [
      'Seller controls the entire transport chain',
      'Buyer avoids freight coordination',
      'Clear handover at destination',
      'Buyer only pays local duties and taxes'
    ],
    riskPoint: 'At Destination',
    riskDesc: 'When goods arrive ready for unloading',
    riskIcon: <CheckCircle2 size={24} strokeWidth={1.5} />,
    sellerSplit: 85,
    buyerSplit: 15,
    insight: 'Growing in popularity for e-commerce and B2B tech imports.'
  },
  CPT: {
    code: 'CPT',
    name: 'Carriage Paid To',
    desc: 'Seller pays freight to destination, but risk transfers at first carrier.',
    reasons: [
      'Great for air and multimodal transport',
      'Seller manages logistics to destination',
      'Risk passes early to the buyer',
      'Avoids destination delays for seller'
    ],
    riskPoint: 'At First Carrier',
    riskDesc: 'Risk transfers upon handover to first carrier',
    riskIcon: <Truck size={24} strokeWidth={1.5} />,
    sellerSplit: 55,
    buyerSplit: 45,
    insight: 'The multimodal equivalent of CFR, perfect for air freight.'
  }
};

function getRecommendation(role, control, risk, mode) {
  if (mode === 'Sea') {
    if (control === 'Buyer') return RECOMMENDATIONS.FOB;
    if (control === 'Seller') {
      return risk === 'Lowest Risk' ? RECOMMENDATIONS.CIF : RECOMMENDATIONS.CFR;
    }
    return RECOMMENDATIONS.FOB; // Shared
  } else {
    // Air, Road, Rail, Multimodal
    if (control === 'Buyer') return RECOMMENDATIONS.FCA;
    if (control === 'Seller') {
      return risk === 'Lowest Risk' ? RECOMMENDATIONS.DAP : RECOMMENDATIONS.CPT;
    }
    return RECOMMENDATIONS.FCA; // Shared
  }
}


/* ─────────────────────────────────────────
   SMART SELECTOR COMPONENT
───────────────────────────────────────── */
export default function SmartSelector() {
  const [role, setRole] = useState('Exporter');
  const [control, setControl] = useState('Buyer');
  const [risk, setRisk] = useState('Balanced');
  const [mode, setMode] = useState('Sea');
  
  const [result, setResult] = useState(RECOMMENDATIONS.FOB);

  // When 'Get Recommendation' is clicked
  const handleCalculate = () => {
    const rec = getRecommendation(role, control, risk, mode);
    setResult(rec);
  };

  // Initial calculation on mount
  useEffect(() => {
    handleCalculate();
    // eslint-disable-next-line
  }, []);

  return (
    <section className="smart-selector-section">
      <div className="ss-bg-decor" />
      
      <div className="ss-container">
        {/* Header */}
        <div className="ss-header">
          <div className="ss-pill">
            <Settings2 size={14} className="ss-pill-icon" />
            SMART SELECTOR
          </div>
          <h2 className="ss-title">Which Incoterm Fits Your Shipment?</h2>
          <p className="ss-subtitle">
            Answer a few simple questions and instantly discover the best Incoterm for your trade scenario.
          </p>
        </div>

        {/* Two Column Layout */}
        <div className="ss-layout">
          
          {/* ── LEFT PANEL ── */}
          <div className="ss-left">
            <div className="ss-steps-line">
              <div className="ss-steps-progress" style={{ height: '100%' }} />
            </div>

            <div className="ss-questions">
              
              {/* Q1 */}
              <div className="ss-q-row active">
                <div className="ss-q-num">1</div>
                <div className="ss-q-content">
                  <h3 className="ss-q-title">What is your role?</h3>
                  <div className="ss-options">
                    {['Exporter', 'Importer', 'Freight Forwarder'].map(opt => (
                      <div 
                        key={opt}
                        className={`ss-opt ${role === opt ? 'selected' : ''}`}
                        onClick={() => setRole(opt)}
                      >
                        {role === opt && <CheckCircle2 size={16} className="ss-opt-icon" />}
                        {opt}
                      </div>
                    ))}
                  </div>
                </div>
              </div>

              {/* Q2 */}
              <div className="ss-q-row active">
                <div className="ss-q-num">2</div>
                <div className="ss-q-content">
                  <h3 className="ss-q-title">Who should control freight?</h3>
                  <div className="ss-options">
                    {['Seller', 'Buyer', 'Shared'].map(opt => (
                      <div 
                        key={opt}
                        className={`ss-opt ${control === opt ? 'selected' : ''}`}
                        onClick={() => setControl(opt)}
                      >
                        {control === opt && <CheckCircle2 size={16} className="ss-opt-icon" />}
                        {opt}
                      </div>
                    ))}
                  </div>
                </div>
              </div>

              {/* Q3 */}
              <div className="ss-q-row active">
                <div className="ss-q-num">3</div>
                <div className="ss-q-content">
                  <h3 className="ss-q-title">Risk preference?</h3>
                  <div className="ss-options">
                    {['Lowest Risk', 'Balanced', 'Maximum Control'].map(opt => (
                      <div 
                        key={opt}
                        className={`ss-opt ${risk === opt ? 'selected' : ''}`}
                        onClick={() => setRisk(opt)}
                      >
                        {risk === opt && <CheckCircle2 size={16} className="ss-opt-icon" />}
                        {opt}
                      </div>
                    ))}
                  </div>
                </div>
              </div>

              {/* Q4 */}
              <div className="ss-q-row active">
                <div className="ss-q-num">4</div>
                <div className="ss-q-content">
                  <h3 className="ss-q-title">Transport mode?</h3>
                  <div className="ss-options">
                    {[
                      { label: 'Sea', icon: <Ship size={14}/> },
                      { label: 'Air', icon: <Plane size={14}/> },
                      { label: 'Road', icon: <Truck size={14}/> },
                      { label: 'Rail', icon: <Train size={14}/> },
                      { label: 'Multimodal', icon: <Package size={14}/> }
                    ].map(opt => (
                      <div 
                        key={opt.label}
                        className={`ss-opt ${mode === opt.label ? 'selected' : ''}`}
                        onClick={() => setMode(opt.label)}
                      >
                        <span className="ss-opt-icon">{opt.icon}</span>
                        {opt.label}
                      </div>
                    ))}
                  </div>
                </div>
              </div>

            </div>

            <div className="ss-actions">
              <button 
                className="ss-btn-reset"
                onClick={() => {
                  setRole('Exporter');
                  setControl('Buyer');
                  setRisk('Balanced');
                  setMode('Sea');
                }}
              >
                <RefreshCcw size={14} /> Reset
              </button>
              <button className="ss-btn-submit" onClick={handleCalculate}>
                Get Recommendation <ArrowRight size={16} />
              </button>
            </div>
          </div>


          {/* ── RIGHT PANEL ── */}
          <div className="ss-right">
            <AnimatePresence mode="wait">
              <motion.div
                key={result.code}
                initial={{ opacity: 0, y: 15 }}
                animate={{ opacity: 1, y: 0 }}
                exit={{ opacity: 0, y: -15 }}
                transition={{ duration: 0.3 }}
                style={{ display: 'flex', flexDirection: 'column', height: '100%' }}
              >
                <div className="ss-rec-badge">⭐ RECOMMENDED</div>
                
                <div className="ss-r-top">
                  <div className="ss-r-text">
                    <h3 className="ss-r-code">{result.code}</h3>
                    <h4 className="ss-r-name">{result.name}</h4>
                    <p className="ss-r-desc">{result.desc}</p>
                  </div>
                  <div className="ss-r-img-wrap">
                    {/* SVG Illustration replacing a raster image for premium scalable look */}
                    <svg viewBox="0 0 200 120" className="ss-ship-icon">
                      <defs>
                        <linearGradient id="hullGrad" x1="0%" y1="0%" x2="0%" y2="100%">
                          <stop offset="0%" stopColor="#3b82f6" />
                          <stop offset="100%" stopColor="#1e3a8a" />
                        </linearGradient>
                        <linearGradient id="boxGrad1" x1="0%" y1="0%" x2="100%" y2="100%">
                          <stop offset="0%" stopColor="#f43f5e" />
                          <stop offset="100%" stopColor="#be123c" />
                        </linearGradient>
                        <linearGradient id="boxGrad2" x1="0%" y1="0%" x2="100%" y2="100%">
                          <stop offset="0%" stopColor="#10b981" />
                          <stop offset="100%" stopColor="#047857" />
                        </linearGradient>
                      </defs>
                      <path d="M10,80 L180,80 L190,50 L10,50 Z" fill="url(#hullGrad)" />
                      <rect x="140" y="30" width="30" height="20" fill="#cbd5e1" />
                      <rect x="150" y="20" width="10" height="10" fill="#94a3b8" />
                      
                      {/* Containers */}
                      <rect x="30" y="35" width="25" height="15" fill="url(#boxGrad1)" />
                      <rect x="60" y="35" width="25" height="15" fill="url(#boxGrad2)" />
                      <rect x="90" y="35" width="25" height="15" fill="url(#boxGrad1)" />
                      
                      <rect x="45" y="20" width="25" height="15" fill="url(#boxGrad2)" />
                      <rect x="75" y="20" width="25" height="15" fill="url(#boxGrad1)" />
                      
                      {/* Water */}
                      <path d="M0,85 Q50,95 100,85 T200,85" fill="none" stroke="#60a5fa" strokeWidth="4" opacity="0.5" />
                      <path d="M0,95 Q50,105 100,95 T200,95" fill="none" stroke="#93c5fd" strokeWidth="3" opacity="0.3" />
                    </svg>
                  </div>
                </div>

                <div className="ss-r-cards">
                  {/* Why this works */}
                  <div className="ss-rc">
                    <h5 className="ss-rc-title">Why this works for you</h5>
                    <ul className="ss-why-list">
                      {result.reasons.map((r, i) => (
                        <li key={i} className="ss-why-item">
                          <CheckCircle2 size={16} className="ss-why-check" />
                          <span>{r}</span>
                        </li>
                      ))}
                    </ul>
                  </div>

                  {/* Risk Transfer */}
                  <div className="ss-rc">
                    <div className="ss-rc-center">
                      <div className="ss-risk-icon">
                        {result.riskIcon}
                      </div>
                      <div className="ss-risk-point">{result.riskPoint}</div>
                      <div className="ss-risk-desc">{result.riskDesc}</div>
                    </div>
                  </div>
                </div>

                {/* Responsibility Split */}
                <div className="ss-split-card">
                  <h5 className="ss-split-title">Responsibility Split</h5>
                  
                  <div className="ss-sb-row">
                    <div className="ss-sb-label">Seller</div>
                    <div className="ss-sb-bar-wrap">
                      <div className="ss-sb-bar blue" style={{ width: `${result.sellerSplit}%` }} />
                    </div>
                    <div className="ss-sb-val blue">{result.sellerSplit}%</div>
                  </div>

                  <div className="ss-sb-row">
                    <div className="ss-sb-label">Buyer</div>
                    <div className="ss-sb-bar-wrap">
                      <div className="ss-sb-bar green" style={{ width: `${result.buyerSplit}%` }} />
                    </div>
                    <div className="ss-sb-val green">{result.buyerSplit}%</div>
                  </div>
                </div>

                {/* Insight Footer */}
                <div className="ss-insight">
                  <Users size={16} className="ss-insight-icon" />
                  <span dangerouslySetInnerHTML={{ __html: result.insight.replace(/(\d+%)/g, '<strong>$1</strong>') }} />
                </div>

              </motion.div>
            </AnimatePresence>
          </div>

        </div>
      </div>
    </section>
  );
}

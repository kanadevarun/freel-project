import React, { useState } from 'react';
import { 
  AlertTriangle,
  ArrowRight,
  Package,
  Container,
  FileBadge,
  ShieldAlert,
  CheckCircle2,
  XCircle,
  Lightbulb,
  Check,
  Ship,
  Plane,
  Truck,
  Box,
  Globe2,
  Target
} from 'lucide-react';
import {
  ShipIllustration,
  SuspendedContainer,
  InvoiceStack,
  AirplaneGraphic,
  DoorDelivery,
  ShieldDocs,
  BlueShield
} from './MistakeIcons';
import './RealTradeMistakes.css';

export default function RealTradeMistakes() {
  const [userType, setUserType] = useState('Exporting');
  const [cargoValue, setCargoValue] = useState('$100K');
  const [transportMode, setTransportMode] = useState('Sea');

  // Dynamic risk list based on selections
  const getRisks = () => {
    const risks = [
      "Incorrect Duty Responsibility",
      "Hidden Local Charges at Port",
      "Incorrect Risk Transfer Point"
    ];
    
    if (cargoValue === '$250K+') {
      risks.unshift("Severe Under-Insurance Risk");
    } else {
      risks.unshift("Wrong Insurance Coverage");
    }

    if (transportMode === 'Air') {
      risks.push("Using Sea-Only Term (e.g. FOB) for Air");
    }

    return risks.slice(0, 4); // Always show 4
  };

  return (
    <section className="rtm-section">
      <div className="rtm-container">
        
        {/* Header */}
        <div className="rtm-header">
          <div className="rtm-pill">
            <AlertTriangle size={14} />
            REAL TRADE MISTAKES
          </div>
          <h2 className="rtm-title">$25,000+ Mistakes Caused By Wrong Incoterms</h2>
          <p className="rtm-subtitle">
            The most expensive shipping problems happen before cargo even moves.
          </p>
        </div>

        {/* Section 1: Shipment Scenario */}
        <div className="rtm-scenario-grid">
          
          {/* Left Card: Overview */}
          <div className="rtm-card rtm-scenario-card-left">
            <div className="rtm-scenario-img">
              <div className="rtm-img-overlay"></div>
              <img src="/images/Massive_container_ship_crossing_the_202606052222.jpeg" alt="Container Ship" />
            </div>
            <div className="rtm-scenario-content">
              <div className="rtm-sc-pill">
                <FileBadge size={14} /> Shipment Overview
              </div>
              
              <div className="rtm-sc-route">
                <div className="rtm-sc-loc">
                  <span className="rtm-sc-city">Mumbai</span>
                  <span className="rtm-sc-country">India</span>
                </div>
                <div className="rtm-sc-arrow">
                  <ArrowRight size={20} />
                </div>
                <div className="rtm-sc-loc">
                  <span className="rtm-sc-city">Hamburg</span>
                  <span className="rtm-sc-country">Germany</span>
                </div>
              </div>

              <div className="rtm-sc-meta">
                <div className="rtm-meta-box">
                  <Package className="rtm-meta-icon" size={20} />
                  <span className="rtm-meta-label">Cargo Value</span>
                  <span className="rtm-meta-val">$120,000</span>
                </div>
                <div className="rtm-meta-box">
                  <Container className="rtm-meta-icon" size={20} />
                  <span className="rtm-meta-label">Container</span>
                  <span className="rtm-meta-val">1 x 40FT HC</span>
                </div>
                <div className="rtm-meta-box" style={{ background: '#fef2f2', border: '1px solid #fee2e2' }}>
                  <ShieldAlert className="rtm-meta-icon" size={20} color="#ef4444" />
                  <span className="rtm-meta-label" style={{ color: '#ef4444' }}>Chosen Term</span>
                  <span className="rtm-meta-val red">EXW</span>
                </div>
              </div>
            </div>
          </div>

          {/* Right Card: Unexpected Charges */}
          <div className="rtm-card rtm-alert-card">
            <div className="rtm-ac-header">
              <AlertTriangle size={24} /> Unexpected Charges
            </div>
            
            <div className="rtm-ac-layout">
              <div className="rtm-ac-list">
                <div className="rtm-ac-item">
                  <div className="rtm-ac-item-icon"><ShieldAlert size={16}/> Import Duties</div>
                  <div className="rtm-ac-item-val">$4,500</div>
                </div>
                <div className="rtm-ac-item">
                  <div className="rtm-ac-item-icon"><Package size={16}/> Terminal Handling</div>
                  <div className="rtm-ac-item-val">$1,200</div>
                </div>
                <div className="rtm-ac-item">
                  <div className="rtm-ac-item-icon"><AlertTriangle size={16}/> Demurrage</div>
                  <div className="rtm-ac-item-val">$850</div>
                </div>
                <div className="rtm-ac-item">
                  <div className="rtm-ac-item-icon"><Container size={16}/> Storage</div>
                  <div className="rtm-ac-item-val">$600</div>
                </div>
                
                <div className="rtm-ac-total">
                  <span className="rtm-ac-total-label">Total Surprise Cost</span>
                  <span className="rtm-ac-total-val">$7,150</span>
                </div>
              </div>

              <div className="rtm-ac-graphic">
                <img src="/images/incoterms/unexpected_charges_illustration.png" alt="Unexpected Invoice" />
              </div>
            </div>
          </div>

        </div>

        {/* Section 2: Mistake Cards */}
        <div className="rtm-mistakes-grid">
          
          <div className="rtm-m-card red">
            <div className="rtm-mc-head">
              <XCircle size={16} /> 1. Wrong Choice
            </div>
            <h4 className="rtm-mc-title">Startup Imports Machinery</h4>
            <div className="rtm-mc-term">Used EXW</div>
            <p className="rtm-mc-desc">Thought supplier handled everything.</p>
            <div className="rtm-mc-reality">
              Reality: Buyer had to arrange transport, customs, insurance, and export docs in a foreign country.
            </div>
            <div className="rtm-mc-costbox">
              <span className="rtm-mc-c-label">Extra Cost</span>
              <span className="rtm-mc-c-val">$8,500</span>
            </div>
            <div className="rtm-mc-graphic">
              <img src="/images/incoterms/wrongChoice.png" alt="EXW Mistake" />
            </div>
          </div>

          <div className="rtm-m-card green">
            <div className="rtm-mc-head">
              <CheckCircle2 size={16} /> 2. Correct Choice
            </div>
            <h4 className="rtm-mc-title">Startup Imports Machinery</h4>
            <div className="rtm-mc-term">Used DDP</div>
            <p className="rtm-mc-desc">Everything delivered directly to their warehouse door.</p>
            <div className="rtm-mc-reality">
              Reality: Seller handled all export/import clearance and final delivery logistics.
            </div>
            <div className="rtm-mc-costbox green">
              <span className="rtm-mc-c-label">Unexpected Costs</span>
              <span className="rtm-mc-c-val">$0</span>
            </div>
            <div className="rtm-mc-graphic"><img src="/images/warehouse_operations.png" alt="DDP Correct" style={{borderRadius: '8px'}} /></div>
          </div>

          <div className="rtm-m-card orange">
            <div className="rtm-mc-head">
              <AlertTriangle size={16} /> 3. FOB for Air Freight
            </div>
            <h4 className="rtm-mc-title">Very Common Mistake</h4>
            <div className="rtm-mc-term">FOB is Sea Only</div>
            <p className="rtm-mc-desc">Many use FOB for air cargo incorrectly.</p>
            <div className="rtm-mc-reality">
              Reality: Risk transfers when goods pass the ship's rail. Air cargo has no ship rail. Use FCA instead.
            </div>
            <div className="rtm-mc-graphic"><img src="/images/air_cargo_night.png" alt="FOB Mistake" style={{borderRadius: '8px'}} /></div>
          </div>

          <div className="rtm-m-card yellow">
            <div className="rtm-mc-head">
              <AlertTriangle size={16} /> 4. CIF Misunderstood
            </div>
            <h4 className="rtm-mc-title">Minimum Insurance</h4>
            <div className="rtm-mc-term">CIF Insurance is Basic</div>
            <p className="rtm-mc-desc">Insurance only covers minimum protection (Institute Cargo Clauses C).</p>
            <div className="rtm-mc-reality">
              Reality: High-value cargo often needs comprehensive CIP coverage instead.
            </div>
            <div className="rtm-mc-graphic"><img src="/images/container_teal.png" alt="CIF Mistake" style={{borderRadius: '8px'}} /></div>
          </div>

        </div>

        {/* Section 3: Risk Checker */}
        <div className="rtm-risk-checker">
          
          <div className="rtm-rc-form">
            <div className="rtm-rc-header">
              <h3 className="rtm-rc-title"><Target size={20} color="#3b82f6"/> Would This Mistake Affect You?</h3>
              <p className="rtm-rc-sub">Answer a few questions to identify potential shipment risks.</p>
            </div>

            <div className="rtm-rc-steps">
              {/* Step 1 */}
              <div className="rtm-rc-step">
                <span className="rtm-rc-step-label">1. I am</span>
                <div className="rtm-rc-options">
                  <button className={`rtm-opt-btn ${userType === 'Exporting' ? 'active' : ''}`} onClick={() => setUserType('Exporting')}>
                    <Ship size={18} /> Exporting
                  </button>
                  <button className={`rtm-opt-btn ${userType === 'Importing' ? 'active' : ''}`} onClick={() => setUserType('Importing')}>
                    <Globe2 size={18} /> Importing
                  </button>
                </div>
              </div>
              
              <ArrowRight className="rtm-rc-arrow" size={16} />

              {/* Step 2 */}
              <div className="rtm-rc-step">
                <span className="rtm-rc-step-label">2. Cargo Value</span>
                <div className="rtm-rc-options">
                  {['$50K', '$100K', '$250K+'].map(val => (
                    <button key={val} className={`rtm-opt-btn ${cargoValue === val ? 'active' : ''}`} onClick={() => setCargoValue(val)}>
                      {val}
                    </button>
                  ))}
                </div>
              </div>

              <ArrowRight className="rtm-rc-arrow" size={16} />

              {/* Step 3 */}
              <div className="rtm-rc-step">
                <span className="rtm-rc-step-label">3. Transport Mode</span>
                <div className="rtm-rc-options">
                  <button className={`rtm-opt-btn ${transportMode === 'Sea' ? 'active' : ''}`} onClick={() => setTransportMode('Sea')}>
                    <Ship size={18} /> Sea
                  </button>
                  <button className={`rtm-opt-btn ${transportMode === 'Air' ? 'active' : ''}`} onClick={() => setTransportMode('Air')}>
                    <Plane size={18} /> Air
                  </button>
                  <button className={`rtm-opt-btn ${transportMode === 'Road' ? 'active' : ''}`} onClick={() => setTransportMode('Road')}>
                    <Truck size={18} /> Road
                  </button>
                  <button className={`rtm-opt-btn ${transportMode === 'Multi-modal' ? 'active' : ''}`} onClick={() => setTransportMode('Multi-modal')}>
                    <Box size={18} /> Multi-modal
                  </button>
                </div>
              </div>
            </div>

            <div className="rtm-pro-tip">
              <Lightbulb className="rtm-pt-icon" size={20} />
              Pro Tip: The right Incoterm can save you thousands in unexpected costs and protect your business from unnecessary risks.
            </div>

          </div>

          <div className="rtm-rc-result">
            <div className="rtm-res-graphic"><img src="/images/logistics_control_tower.png" alt="Security Shield" style={{borderRadius: '20px'}} /></div>
            <h4 className="rtm-res-title">Top Risks You Should Avoid</h4>
            <div className="rtm-res-list">
              {getRisks().map((risk, i) => (
                <div key={i} className="rtm-res-item">
                  <CheckCircle2 size={16} className="rtm-res-icon" /> {risk}
                </div>
              ))}
            </div>
          </div>

        </div>

      </div>
    </section>
  );
}

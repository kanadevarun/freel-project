import { motion } from 'framer-motion';
import { Plane, Play, CheckCircle2, Globe, Clock, ShieldCheck, MapPin, Building2, Package, Weight, DollarSign, ArrowRight, Store, Home, Factory, Truck, FileText, AlertTriangle, Users, Lightbulb, Check, Circle, Scale, Zap, Leaf, Box, Star, Target, Info, ClipboardList, X, Tag, User, Warehouse, Rocket, Settings, FileWarning, ArrowUpCircle, GraduationCap, BookOpen, Calculator, BarChart2, Medal, Ship, CalendarDays, Briefcase, Anchor, Award } from 'lucide-react';
import { Link } from 'react-router-dom';
import './Page.css';

export default function AirFreightPage() {
  return (
    <div className="af-page-wrapper">
      <section className="af-hero-wrapper">
        {/* Background Layer */}
        <div className="af-hero-bg"></div>
        {/* White Gradient Overlay */}
        <div className="af-hero-effects"></div>

        <div className="af-container">
          {/* Left Column */}
          <motion.div 
            className="af-hero-left"
            initial={{ opacity: 0, y: 30 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.8, ease: "easeOut" }}
          >
            <div className="af-eyebrow">
              <span className="af-eyebrow-icon">✈️</span> Air Freight Guide
            </div>

            <h1 className="af-headline">
              Master Global <br />
              <span className="af-headline-gradient">Air Cargo</span> Without <br />
              Expensive Mistakes
            </h1>

            <p className="af-subtext">
              Learn how airlines, airports, freight forwarders and customs authorities move cargo across the world. Master chargeable weight, AWBs, transit times and air freight pricing with practical examples.
            </p>

            <div className="af-hero-actions">
              <Link to="#learn" className="af-btn-primary">
                Start Learning &rarr;
              </Link>
              <button className="af-btn-secondary">
                <Play fill="#0f172a" size={18} /> Watch Cargo Journey
              </button>
            </div>

            <div className="af-stats-container">
              <div className="af-stat-card">
                <div className="af-stat-icon-wrap"><Globe size={20} /></div>
                <div className="af-stat-value">500+</div>
                <div className="af-stat-label">Airports Covered</div>
              </div>
              <div className="af-stat-card">
                <div className="af-stat-icon-wrap"><Plane size={20} /></div>
                <div className="af-stat-value">300+</div>
                <div className="af-stat-label">Airlines</div>
              </div>
              <div className="af-stat-card">
                <div className="af-stat-icon-wrap"><Clock size={20} /></div>
                <div className="af-stat-value">1-7 Days</div>
                <div className="af-stat-label">Typical Transit</div>
              </div>
              <div className="af-stat-card">
                <div className="af-stat-icon-wrap"><Globe size={20} /></div>
                <div className="af-stat-value">190+</div>
                <div className="af-stat-label">Countries Connected</div>
              </div>
            </div>
          </motion.div>

          {/* Right Column */}
          <motion.div 
            className="af-hero-right"
            initial={{ opacity: 0, x: 40 }}
            animate={{ opacity: 1, x: 0 }}
            transition={{ duration: 0.9, ease: "easeOut", delay: 0.2 }}
          >
            <div className="af-floating-card">
              <div className="af-card-title">
                REAL SHIPMENT EXAMPLE
                <div style={{ display: 'flex', gap: '4px' }}>
                  <div style={{width:'4px', height:'4px', borderRadius:'50%', background:'#cbd5e1'}}></div>
                  <div style={{width:'4px', height:'4px', borderRadius:'50%', background:'#cbd5e1'}}></div>
                  <div style={{width:'4px', height:'4px', borderRadius:'50%', background:'#cbd5e1'}}></div>
                </div>
              </div>

              <div className="af-route-display">
                <div className="af-route-point">
                  <div className="af-flag" style={{background: 'linear-gradient(to bottom, #ff9933 33%, white 33% 66%, #138808 66%)'}}></div>
                  <div className="af-loc-info">
                    <span className="af-loc-city">Mumbai,</span>
                    <span className="af-loc-country">India</span>
                  </div>
                </div>
                <div className="af-route-line">
                  <div className="af-route-dash"></div>
                  <Plane size={18} className="af-route-plane" style={{ transform: 'rotate(45deg)' }} />
                </div>
                <div className="af-route-point">
                  <div className="af-flag" style={{background: 'linear-gradient(to bottom, black 33%, red 33% 66%, gold 66%)'}}></div>
                  <div className="af-loc-info">
                    <span className="af-loc-city">Frankfurt,</span>
                    <span className="af-loc-country">Germany</span>
                  </div>
                </div>
              </div>

              <div className="af-metrics-grid">
                <div className="af-metric-box">
                  <div className="af-metric-icon"><Package size={16} /></div>
                  <div className="af-metric-text">
                    <span className="af-metric-label">Cargo:</span>
                    <span className="af-metric-val">Electronics</span>
                  </div>
                </div>
                <div className="af-metric-box">
                  <div className="af-metric-icon"><Weight size={16} /></div>
                  <div className="af-metric-text">
                    <span className="af-metric-label">Weight:</span>
                    <span className="af-metric-val">500 KG</span>
                  </div>
                </div>
                <div className="af-metric-box">
                  <div className="af-metric-icon"><Clock size={16} /></div>
                  <div className="af-metric-text">
                    <span className="af-metric-label">Transit:</span>
                    <span className="af-metric-val">3 Days</span>
                  </div>
                </div>
                <div className="af-metric-box">
                  <div className="af-metric-icon"><DollarSign size={16} /></div>
                  <div className="af-metric-text">
                    <span className="af-metric-label">Cost:</span>
                    <span className="af-metric-val">$12,400</span>
                  </div>
                </div>
              </div>

              <div className="af-visual-sequence">
                <div className="af-sequence-label">AIR FREIGHT JOURNEY</div>
                <div className="af-sequence-flow">
                  <div className="af-seq-node">
                    <div className="af-seq-icon-wrap"><Store size={18} /></div>
                    <div className="af-seq-text">Factory</div>
                  </div>
                  <ArrowRight size={14} className="af-seq-arrow" />
                  <div className="af-seq-node">
                    <div className="af-seq-icon-wrap"><Building2 size={18} /></div>
                    <div className="af-seq-text">Origin Airport</div>
                  </div>
                  <ArrowRight size={14} className="af-seq-arrow" />
                  <div className="af-seq-node">
                    <div className="af-seq-icon-wrap active"><Plane size={18} style={{transform: 'rotate(45deg)'}} /></div>
                    <div className="af-seq-text">Aircraft</div>
                  </div>
                  <ArrowRight size={14} className="af-seq-arrow" />
                  <div className="af-seq-node">
                    <div className="af-seq-icon-wrap"><Building2 size={18} /></div>
                    <div className="af-seq-text">Destination Airport</div>
                  </div>
                  <ArrowRight size={14} className="af-seq-arrow" />
                  <div className="af-seq-node">
                    <div className="af-seq-icon-wrap"><Home size={18} /></div>
                    <div className="af-seq-text">Customer</div>
                  </div>
                </div>
                <div className="af-success-bar">
                  <CheckCircle2 size={18} /> Delivered Successfully
                </div>
              </div>
            </div>
          </motion.div>
        </div>

        {/* Bottom Trust Logos Bar */}
        <div className="af-bottom-bar">
          <div className="af-trust-text">
            <div className="af-trust-shield"><ShieldCheck size={24} /></div>
            <div className="af-trust-info">
              <span className="af-trust-title">Trusted by Global Logistics Professionals</span>
              <span className="af-trust-desc">Join thousands of importers, exporters, freight forwarders and supply chain teams.</span>
            </div>
          </div>
          <div className="af-logos">
            <span className="af-logo" style={{fontWeight: 900, fontSize: '1.2rem', letterSpacing: '-1px'}}>DHL</span>
            <span className="af-logo" style={{fontWeight: 700, fontSize: '1rem'}}>KUEHNE+NAGEL</span>
            <span className="af-logo" style={{fontWeight: 800, fontSize: '1.1rem'}}>DB SCHENKER</span>
            <span className="af-logo" style={{fontWeight: 800, fontSize: '1.1rem', color: '#4d148c'}}>Fed<span style={{color:'#ff6200'}}>Ex</span></span>
            <span className="af-logo" style={{fontWeight: 800, fontSize: '1rem', letterSpacing: '1px'}}>* MAERSK</span>
          </div>
        </div>
      </section>

      {/* ==============================================
          HOW AIR FREIGHT ACTUALLY WORKS - JOURNEY SECTION 
          ============================================== */}
      <section className="af-journey-section">
        <div className="af-section-header">
          <div className="af-section-eyebrow">
            <Plane size={16} style={{transform: 'rotate(45deg)'}} /> AIR CARGO JOURNEY
          </div>
          <h2 className="af-section-title">How Air Freight Actually Works</h2>
          <p className="af-section-subtitle">
            Understand the complete journey of air cargo from manufacturer to final customer delivery.
          </p>
        </div>

        <div className="af-journey-card">
          {/* 1. Stepper */}
          <div className="af-stepper">
            <div className="af-stepper-connector"></div>
            
            <div className="af-step">
              <div className="af-step-icon-container">
                <Factory size={32} />
                <div className="af-step-number">1</div>
              </div>
              <div className="af-step-title">Factory</div>
              <div className="af-step-desc">Goods are packed and ready for shipment</div>
            </div>
            
            <div className="af-step">
              <div className="af-step-icon-container">
                <Truck size={32} />
                <div className="af-step-number">2</div>
              </div>
              <div className="af-step-title">Pickup</div>
              <div className="af-step-desc">Cargo is picked up from the factory</div>
            </div>

            <div className="af-step">
              <div className="af-step-icon-container">
                <Building2 size={32} />
                <div className="af-step-number">3</div>
              </div>
              <div className="af-step-title">Origin Airport</div>
              <div className="af-step-desc">Cargo arrives at departure airport</div>
            </div>

            <div className="af-step">
              <div className="af-step-icon-container">
                <Package size={32} />
                <div className="af-step-number">4</div>
              </div>
              <div className="af-step-title">Cargo Handling</div>
              <div className="af-step-desc">Screened and prepared for flight</div>
            </div>

            <div className="af-step active">
              <div className="af-step-icon-container">
                <Plane size={36} style={{transform: 'rotate(45deg)'}} />
                <div className="af-step-number">5</div>
              </div>
              <div className="af-step-title">Aircraft</div>
              <div className="af-step-desc">Cargo is loaded for transport</div>
            </div>

            <div className="af-step">
              <div className="af-step-icon-container">
                <Building2 size={32} />
                <div className="af-step-number">6</div>
              </div>
              <div className="af-step-title">Destination Airport</div>
              <div className="af-step-desc">Cargo arrives at destination</div>
            </div>

            <div className="af-step">
              <div className="af-step-icon-container">
                <ShieldCheck size={32} />
                <div className="af-step-number">7</div>
              </div>
              <div className="af-step-title">Customs</div>
              <div className="af-step-desc">Clearance and documentation</div>
            </div>

            <div className="af-step">
              <div className="af-step-icon-container">
                <Truck size={32} />
                <div className="af-step-number">8</div>
              </div>
              <div className="af-step-title">Final Delivery</div>
              <div className="af-step-desc">Delivered to final recipient</div>
            </div>
          </div>

          {/* 2. Details Grid */}
          <div className="af-journey-details">
            <div className="af-jd-col">
              <img src="/images/air-freight/White_figure_loading_aircraft_202606071515.jpeg" alt="Aircraft Loading" className="af-jd-image" />
            </div>

            <div className="af-jd-col">
              <span className="af-jd-label">CURRENT STAGE</span>
              <h3 className="af-jd-title">Aircraft Loading</h3>
              <p className="af-jd-desc">
                The cargo is safely loaded into the aircraft hold or on the main deck (in case of freighters) according to airline and IATA regulations.
              </p>
              <div className="af-jd-badge">
                <Users size={18} /> Handled By Airline + Ground Handling Agent
              </div>
            </div>

            <div className="af-jd-col">
              <div className="af-jd-col-header">
                <FileText size={18} /> Documents Required
              </div>
              <ul className="af-jd-list checks">
                <li><Check size={16} /> Air Waybill (AWB)</li>
                <li><Check size={16} /> Commercial Invoice</li>
                <li><Check size={16} /> Packing List</li>
                <li><Check size={16} /> Export Declaration</li>
              </ul>
            </div>

            <div className="af-jd-col">
              <div className="af-jd-col-header">
                <Clock size={18} /> Typical Duration
              </div>
              <div className="af-jd-duration-val">4 - 12 Hours</div>
              <div className="af-jd-duration-sub" style={{marginBottom: '24px'}}>
                Depending on cargo type, volume and airport operational flow.
              </div>

              <div className="af-jd-col-header red">
                <AlertTriangle size={18} /> Common Risks
              </div>
              <ul className="af-jd-list dots">
                <li><Circle size={8} fill="currentColor" /> Incorrect cargo labeling</li>
                <li><Circle size={8} fill="currentColor" /> Documentation errors</li>
                <li><Circle size={8} fill="currentColor" /> Weight discrepancies</li>
                <li><Circle size={8} fill="currentColor" /> Security inspections</li>
              </ul>
            </div>
          </div>

          {/* 3. Bottom Banner */}
          <div className="af-journey-banner">
            <div className="af-jb-content">
              <div className="af-jb-icon">
                <Lightbulb size={24} />
              </div>
              <div className="af-jb-text">
                <h4>Did You Know?</h4>
                <p>More than 80% of international air cargo travels in the belly hold of passenger aircraft rather than dedicated cargo planes.</p>
              </div>
            </div>
            
            <div className="af-jb-graphic">
              <MapPin size={24} />
              <div className="af-jb-dash"></div>
              <Plane size={32} style={{transform: 'rotate(45deg)'}} />
            </div>
          </div>
        </div>
      </section>

      {/* ==============================================
          COMPARISON GUIDE SECTION
          ============================================== */}
      <section className="af-comparison-section">
        <div className="af-section-header">
          <div className="af-section-eyebrow">
            <Scale size={16} /> COMPARISON GUIDE
          </div>
          <h2 className="af-section-title">Air Freight vs Ocean Freight</h2>
          <p className="af-section-subtitle">
            Understand the tradeoffs between speed, cost, capacity and reliability before choosing a shipping method.
          </p>
        </div>

        <div className="af-comp-container">
          {/* 1. Top Cards Row */}
          <div className="af-comp-cards-row">
            <div className="af-comp-vs-circle">VS</div>
            
            {/* Air Freight Card */}
            <div className="af-comp-card blue">
              <div className="af-comp-badge blue">
                <Zap size={14} fill="currentColor" /> FASTEST OPTION
              </div>
              <img src="/images/air-freight/air-cargo.png" alt="Air Freight" className="af-comp-image" />
              <div className="af-comp-card-body">
                <div className="af-comp-card-title">Air Freight</div>
                <div className="af-comp-metrics">
                  <div className="af-comp-metric">
                    <div className="af-cm-icon"><Clock size={18} /></div>
                    <div className="af-cm-text">
                      <span className="af-cm-label">Transit Time</span>
                      <span className="af-cm-val">1 - 7 Days</span>
                    </div>
                  </div>
                  <div className="af-comp-metric">
                    <div className="af-cm-icon"><Target size={18} /></div>
                    <div className="af-cm-text">
                      <span className="af-cm-label">Best For</span>
                      <span className="af-cm-val">Urgent Cargo</span>
                    </div>
                  </div>
                  <div className="af-comp-metric">
                    <div className="af-cm-icon"><DollarSign size={18} /></div>
                    <div className="af-cm-text">
                      <span className="af-cm-label">Cost</span>
                      <span className="af-cm-val">High</span>
                    </div>
                  </div>
                </div>
              </div>
            </div>

            {/* Ocean Freight Card */}
            <div className="af-comp-card green">
              <div className="af-comp-badge green">
                <Leaf size={14} fill="currentColor" /> MOST ECONOMICAL
              </div>
              <img src="/images/sea-freight/Massive_container_terminal_viewed_from_202606052225.jpeg" alt="Ocean Freight" className="af-comp-image" />
              <div className="af-comp-card-body">
                <div className="af-comp-card-title">Ocean Freight</div>
                <div className="af-comp-metrics">
                  <div className="af-comp-metric">
                    <div className="af-cm-icon"><Clock size={18} /></div>
                    <div className="af-cm-text">
                      <span className="af-cm-label">Transit Time</span>
                      <span className="af-cm-val">20 - 45 Days</span>
                    </div>
                  </div>
                  <div className="af-comp-metric">
                    <div className="af-cm-icon"><Target size={18} /></div>
                    <div className="af-cm-text">
                      <span className="af-cm-label">Best For</span>
                      <span className="af-cm-val">Bulk Cargo</span>
                    </div>
                  </div>
                  <div className="af-comp-metric">
                    <div className="af-cm-icon"><DollarSign size={18} /></div>
                    <div className="af-cm-text">
                      <span className="af-cm-label">Cost</span>
                      <span className="af-cm-val">Low</span>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>

          {/* 2. Middle Details Row */}
          <div className="af-comp-details-row">
            {/* Air Use Cases */}
            <div className="af-comp-use-cases blue">
              <div className="af-cuc-header">
                <div className="af-cuc-icon"><Box size={24} /></div>
                <div className="af-cuc-title-area">
                  <span className="af-cuc-subtitle">Best Use Cases</span>
                  <span className="af-cuc-title">Air Freight</span>
                </div>
              </div>
              <ul className="af-cuc-list">
                <li><CheckCircle2 size={18} fill="currentColor" color="white" /> Electronics</li>
                <li><CheckCircle2 size={18} fill="currentColor" color="white" /> Pharmaceuticals</li>
                <li><CheckCircle2 size={18} fill="currentColor" color="white" /> Fashion & Apparel</li>
                <li><CheckCircle2 size={18} fill="currentColor" color="white" /> Perishable Goods</li>
                <li><CheckCircle2 size={18} fill="currentColor" color="white" /> Emergency Shipments</li>
                <li><CheckCircle2 size={18} fill="currentColor" color="white" /> High Value Cargo</li>
                <li><CheckCircle2 size={18} fill="currentColor" color="white" /> Time-Sensitive Deliveries</li>
              </ul>
            </div>

            {/* Comparison Table */}
            <div className="af-comp-table-wrap">
              <div className="af-ct-header">
                <div className="af-ct-header-col"><Scale size={16} /> FACTOR</div>
                <div className="af-ct-header-col blue-text">AIR FREIGHT</div>
                <div></div>
                <div className="af-ct-header-col green-text">OCEAN FREIGHT</div>
              </div>

              {[
                { label: 'Transit Speed', icon: <Clock size={16} />, air: 5, sea: 2 },
                { label: 'Shipping Cost', icon: <DollarSign size={16} />, air: 2, sea: 5 },
                { label: 'Cargo Capacity', icon: <Package size={16} />, air: 2, sea: 5 },
                { label: 'Reliability', icon: <ShieldCheck size={16} />, air: 4, sea: 3 },
                { label: 'Environmental Impact', icon: <Leaf size={16} />, air: 2, sea: 4 },
                { label: 'Global Reach', icon: <Globe size={16} />, air: 5, sea: 4 },
                { label: 'Documentation Complexity', icon: <FileText size={16} />, air: 3, sea: 4 },
              ].map((row, i) => (
                <div className="af-ct-row" key={i}>
                  <div className="af-ct-factor">
                    {row.icon} {row.label}
                  </div>
                  <div className="af-ct-stars blue">
                    {[...Array(5)].map((_, j) => (
                      <Star key={j} size={14} fill="currentColor" className={j >= row.air ? 'empty' : ''} />
                    ))}
                  </div>
                  <div className="af-ct-vs-col">VS</div>
                  <div className="af-ct-stars green">
                    {[...Array(5)].map((_, j) => (
                      <Star key={j} size={14} fill="currentColor" className={j >= row.sea ? 'empty' : ''} />
                    ))}
                  </div>
                </div>
              ))}
            </div>

            {/* Ocean Use Cases */}
            <div className="af-comp-use-cases green">
              <div className="af-cuc-header">
                <div className="af-cuc-icon"><Building2 size={24} /></div>
                <div className="af-cuc-title-area">
                  <span className="af-cuc-subtitle">Best Use Cases</span>
                  <span className="af-cuc-title">Ocean Freight</span>
                </div>
              </div>
              <ul className="af-cuc-list">
                <li><CheckCircle2 size={18} fill="currentColor" color="white" /> Heavy Machinery</li>
                <li><CheckCircle2 size={18} fill="currentColor" color="white" /> Furniture</li>
                <li><CheckCircle2 size={18} fill="currentColor" color="white" /> Construction Materials</li>
                <li><CheckCircle2 size={18} fill="currentColor" color="white" /> Bulk Commodities</li>
                <li><CheckCircle2 size={18} fill="currentColor" color="white" /> Raw Materials</li>
                <li><CheckCircle2 size={18} fill="currentColor" color="white" /> Large Volume Cargo</li>
                <li><CheckCircle2 size={18} fill="currentColor" color="white" /> Non-Urgent Shipments</li>
              </ul>
            </div>
          </div>

          {/* 3. Bottom Banner */}
          <div className="af-comp-banner">
            <div className="af-cb-left">
              <div className="af-cb-icon"><Lightbulb size={24} /></div>
              <div className="af-cb-text">
                <h4>When Should You Choose Air Freight?</h4>
                <ul className="af-cb-list">
                  <li><CheckCircle2 size={16} /> Speed matters</li>
                  <li><CheckCircle2 size={16} /> Product value is high</li>
                  <li><CheckCircle2 size={16} /> Inventory is urgent</li>
                  <li><CheckCircle2 size={16} /> Product is lightweight</li>
                  <li><CheckCircle2 size={16} /> Delays are costly</li>
                </ul>
              </div>
            </div>
            <div className="af-cb-plane">
              <Package size={24} />
              <Plane size={48} style={{transform: 'rotate(45deg)'}} />
            </div>
          </div>
        </div>
      </section>

      {/* ==============================================
          CHARGEABLE WEIGHT SECTION
          ============================================== */}
      <section className="af-cw-section">
        <div className="af-cw-container">

          {/* Header */}
          <div className="af-cw-header">
            <div className="af-section-eyebrow"><Plane size={14} /> AIR FREIGHT PRICING</div>
            <h2>Understanding Chargeable Weight</h2>
            <p>Airlines charge based on whichever is greater: <strong>actual weight</strong> or <em>volumetric weight.</em></p>
          </div>

          {/* Row 1: Actual vs Volumetric */}
          <div className="af-cw-vs-row">
            {/* Actual Weight */}
            <div className="af-cw-card actual">
              <div className="af-cw-card-content">
                <div>
                  <div className="af-cw-card-label">Actual Weight</div>
                  <p className="af-cw-card-desc" style={{marginTop: '8px'}}>What your cargo physically weighs.</p>
                </div>
                <div className="af-cw-card-weight">
                  <div className="af-cw-card-weight-icon"><Weight size={20} /></div>
                  <div>
                    <div className="af-cw-card-weight-val">120 KG</div>
                    <div className="af-cw-card-weight-sub">Physical weight on scale</div>
                  </div>
                </div>
              </div>
              <div className="af-cw-card-img-wrap">
                <img src="/images/air-freight/actual_weight_box_scale.png" alt="Actual Weight Scale" className="af-cw-card-img" />
              </div>
            </div>

            {/* VS Divider */}
            <div className="af-cw-vs-divider">
              <div className="af-cw-vs-badge">VS</div>
            </div>

            {/* Volumetric Weight */}
            <div className="af-cw-card volumetric">
              <div className="af-cw-card-img-wrap">
                <img src="/images/air-freight/volumetric_weight_box.png" alt="Volumetric Weight Box" className="af-cw-card-img" />
              </div>
              <div className="af-cw-card-content">
                <div>
                  <div className="af-cw-card-label">Volumetric Weight</div>
                  <p className="af-cw-card-desc" style={{marginTop: '8px'}}>How much space your cargo occupies.</p>
                </div>
                <div className="af-cw-card-weight">
                  <div className="af-cw-card-weight-icon"><Box size={20} /></div>
                  <div>
                    <div className="af-cw-card-weight-val">180 KG</div>
                    <div className="af-cw-card-weight-sub">Calculated using volumetric formula.</div>
                  </div>
                </div>
              </div>
            </div>
          </div>

          {/* Row 2: Airline Charges Banner */}
          <div className="af-cw-charges-banner">
            <div className="af-cw-charges-icon"><Weight size={28} /></div>
            <div className="af-cw-charges-center">
              <div className="af-cw-charges-label">Airline Charges For</div>
              <div className="af-cw-charges-val">180 KG</div>
            </div>
            <div className="af-cw-charges-right">
              <div className="af-cw-charges-right-icon"><CheckCircle2 size={18} /></div>
              <p>The higher value always becomes the chargeable weight.</p>
            </div>
          </div>

          {/* Row 3: Formula + Why */}
          <div className="af-cw-formula-row">
            {/* Formula */}
            <div className="af-cw-formula-card">
              <div className="af-cw-formula-title">Volumetric Weight Formula</div>
              <div className="af-cw-formula-boxes">
                <div className="af-cw-formula-box">
                  <Box size={20} />
                  <span>Length<br/>(cm)</span>
                </div>
                <div className="af-cw-formula-op">×</div>
                <div className="af-cw-formula-box">
                  <Box size={20} />
                  <span>Width<br/>(cm)</span>
                </div>
                <div className="af-cw-formula-op">×</div>
                <div className="af-cw-formula-box">
                  <Box size={20} />
                  <span>Height<br/>(cm)</span>
                </div>
                <div className="af-cw-formula-op">÷</div>
                <div className="af-cw-formula-divisor">6000</div>
              </div>
              <div className="af-cw-formula-equation">
                <span>Volumetric Weight (KG) =</span>
                <span className="frac">
                  <span className="num">L × W × H</span>
                  <span className="den">6000</span>
                </span>
              </div>
            </div>

            {/* Why Airlines Use It */}
            <div className="af-cw-why-card">
              <div className="af-cw-why-content">
                <div className="af-cw-why-header">
                  <Info size={18} />
                  <span>Why Airlines Use Volumetric Weight</span>
                </div>
                <p>Aircraft capacity is limited by both weight and available space. Large, lightweight cargo still occupies valuable space that could be used for other shipments.</p>
              </div>
              <img src="/images/air-freight/aircraft_window_crop.png" alt="Aircraft Window" className="af-cw-why-img" />
            </div>
          </div>

          {/* Row 4: Examples */}
          <div className="af-cw-examples-header">
            <FileText size={14} /> SHIPMENT EXAMPLES
          </div>
          <div className="af-cw-examples-row">
            {/* Example 1 */}
            <div className="af-cw-example-card">
              <div className="af-cw-example-card-header">EXAMPLE 1</div>
              <div className="af-cw-example-body">
                <div className="af-cw-example-img-wrap">
                  <img src="/images/air-freight/example1_carton_box.png" alt="Carton Box" className="af-cw-example-img" />
                </div>
                <div className="af-cw-example-details">
                  <div className="af-cw-example-row">
                    <div className="af-cw-example-row-label"><Box size={14} /> Dimensions (L × W × H)</div>
                    <div className="af-cw-example-row-val">100 × 60 × 60 cm</div>
                  </div>
                  <div className="af-cw-example-row">
                    <div className="af-cw-example-row-label"><Weight size={14} /> Actual Weight</div>
                    <div className="af-cw-example-row-val">120 KG</div>
                  </div>
                  <div className="af-cw-example-row">
                    <div className="af-cw-example-row-label"><Scale size={14} /> Volumetric Weight</div>
                    <div className="af-cw-example-row-val highlight-teal">60 KG ↓</div>
                  </div>
                  <div className="af-cw-example-charge">
                    <span>Airline Charges</span>
                    <span>120 KG</span>
                  </div>
                </div>
              </div>
              <div className="af-cw-example-verdict">
                <CheckCircle2 size={16} fill="currentColor" color="white" />
                Actual weight is higher, so you pay for <strong>&nbsp;120 KG</strong>
              </div>
            </div>

            {/* Example 2 */}
            <div className="af-cw-example-card">
              <div className="af-cw-example-card-header">EXAMPLE 2</div>
              <div className="af-cw-example-body">
                <div className="af-cw-example-img-wrap">
                  <img src="/images/air-freight/example2_palletized_cargo_box.png" alt="Palletized Cargo" className="af-cw-example-img" />
                </div>
                <div className="af-cw-example-details">
                  <div className="af-cw-example-row">
                    <div className="af-cw-example-row-label"><Box size={14} /> Dimensions (L × W × H)</div>
                    <div className="af-cw-example-row-val">150 × 120 × 100 cm</div>
                  </div>
                  <div className="af-cw-example-row">
                    <div className="af-cw-example-row-label"><Weight size={14} /> Actual Weight</div>
                    <div className="af-cw-example-row-val">80 KG</div>
                  </div>
                  <div className="af-cw-example-row">
                    <div className="af-cw-example-row-label"><Scale size={14} /> Volumetric Weight</div>
                    <div className="af-cw-example-row-val highlight-teal">300 KG ↑</div>
                  </div>
                  <div className="af-cw-example-charge teal">
                    <span>Airline Charges</span>
                    <span>300 KG</span>
                  </div>
                </div>
              </div>
              <div className="af-cw-example-verdict teal">
                <CheckCircle2 size={16} fill="currentColor" color="white" />
                Volumetric weight is higher, so you pay for <strong>&nbsp;300 KG</strong>
              </div>
            </div>
          </div>

          {/* Row 5: Did You Know + Pro Tip */}
          <div className="af-cw-bottom-row">
            <div className="af-cw-did-you-know">
              <div className="af-cw-dyk-icon"><Lightbulb size={24} /></div>
              <div className="af-cw-dyk-text">
                <h4>Did You Know?</h4>
                <p>A shipment of pillows can cost more than a shipment of metal parts of the same weight because pillows occupy significantly more cargo space.</p>
              </div>
              <div className="af-cw-dyk-images">
                <div className="af-cw-dyk-img-wrap">
                  <img src="/images/air-freight/pillows_stack.png" alt="Pillows Stack" className="af-cw-dyk-img pillows" />
                </div>
                <div className="af-cw-dyk-vs">VS</div>
                <div className="af-cw-dyk-img-wrap">
                  <img src="/images/air-freight/metal_gears_stack.png" alt="Metal Gears Stack" className="af-cw-dyk-img gears" />
                </div>
              </div>
            </div>
            <div className="af-cw-pro-tip">
              <div className="af-cw-pro-tip-icon"><Star size={18} /></div>
              <span><strong>Pro Tip &nbsp;</strong> To reduce shipping costs, optimize your packaging to reduce volume without compromising product safety.</span>
            </div>
          </div>

        </div>
      </section>

      {/* ==================== AIR FREIGHT DOCUMENTATION ==================== */}
      <section className="af-doc-section">
        <div className="af-doc-container">
          <div className="af-doc-header">
            <div className="af-section-eyebrow">
              <FileText size={16} className="text-blue-600" /> AIR FREIGHT DOCUMENTATION
            </div>
            <h2>Documents That Move Air Cargo<br />Across The World</h2>
            <p>Every international air shipment requires specific documents.<br />Missing or incorrect paperwork can delay cargo, increase costs, or cause customs issues.</p>
          </div>

          <div className="af-doc-diagram">
            {/* MAWB Card */}
            <div className="af-doc-awb-card mawb">
              <div className="af-doc-awb-title">Master Air Waybill<br />(MAWB)</div>
              <div className="af-doc-awb-subtitle">Airline Contract</div>
              <div className="af-doc-issued-by"><Plane size={18} /> Issued by Airline</div>
              <ul className="af-doc-checklist">
                <li><Check size={14} /> Shipment reference number</li>
                <li><Check size={14} /> Origin airport</li>
                <li><Check size={14} /> Destination airport</li>
                <li><Check size={14} /> Cargo weight</li>
                <li><Check size={14} /> Flight details</li>
              </ul>
              <img src="/images/air-freight/Highly_stylized_3D_render_of_202606071514.jpeg" alt="Airplane" className="af-doc-awb-card-img" style={{borderRadius: 12, objectFit: 'cover', marginTop: '20px', height: '140px'}} />
            </div>

            {/* Center Flow */}
            <div className="af-doc-center-flow">
              <div className="af-doc-papers-grid">
                <div className="af-doc-paper-item top-left">
                  <span className="af-doc-paper-label">Commercial<br/>Invoice</span>
                  <div className="af-doc-paper-img-wrap">
                    <img src="/images/air-freight/commercial-invoice.png" alt="Commercial Invoice" className="af-doc-paper-img" />
                  </div>
                </div>
                <div className="af-doc-paper-item top-right">
                  <span className="af-doc-paper-label">Packing<br/>List</span>
                  <div className="af-doc-paper-img-wrap">
                    <img src="/images/air-freight/package-list.png" alt="Packing List" className="af-doc-paper-img" />
                  </div>
                </div>
                <div className="af-doc-paper-item main">
                  <div className="af-doc-paper-img-wrap">
                    <img src="/images/air-freight/air-way-bill-middle.png" alt="Air Waybill" className="af-doc-paper-img" />
                  </div>
                </div>
                <div className="af-doc-paper-item bottom-left">
                  <span className="af-doc-paper-label">Certificate<br/>of Origin</span>
                  <div className="af-doc-paper-img-wrap">
                    <img src="/images/air-freight/certificate-of-origin.png" alt="Certificate of Origin" className="af-doc-paper-img" />
                  </div>
                </div>
                <div className="af-doc-paper-item bottom-right">
                  <span className="af-doc-paper-label">Customs<br/>Declaration</span>
                  <div className="af-doc-paper-img-wrap">
                    <img src="/images/air-freight/customs-declaration.png" alt="Customs Declaration" className="af-doc-paper-img" />
                  </div>
                </div>
              </div>

              <div className="af-doc-flow-steps">
                <div className="af-doc-flow-step">
                  <div className="af-doc-flow-icon"><Factory size={24} /></div>
                  <span className="af-doc-flow-label">Exporter</span>
                </div>
                <ArrowRight size={16} className="af-doc-flow-arrow" />
                <div className="af-doc-flow-step">
                  <div className="af-doc-flow-icon"><Plane size={24} /></div>
                  <span className="af-doc-flow-label">Airline</span>
                </div>
                <ArrowRight size={16} className="af-doc-flow-arrow" />
                <div className="af-doc-flow-step">
                  <div className="af-doc-flow-icon green"><ShieldCheck size={24} /></div>
                  <span className="af-doc-flow-label green" style={{color: '#059669'}}>Customs</span>
                </div>
                <ArrowRight size={16} className="af-doc-flow-arrow" />
                <div className="af-doc-flow-step">
                  <div className="af-doc-flow-icon green"><Building2 size={24} /></div>
                  <span className="af-doc-flow-label green" style={{color: '#059669'}}>Importer</span>
                </div>
              </div>
            </div>

            {/* HAWB Card */}
            <div className="af-doc-awb-card hawb">
              <div className="af-doc-awb-title">House Air Waybill<br />(HAWB)</div>
              <div className="af-doc-awb-subtitle">Freight Forwarder Document</div>
              <div className="af-doc-issued-by"><Package size={18} /> Issued by Forwarder</div>
              <ul className="af-doc-checklist">
                <li><Check size={14} /> Consignee details</li>
                <li><Check size={14} /> Cargo details</li>
                <li><Check size={14} /> Internal shipment tracking</li>
                <li><Check size={14} /> Customer shipment reference</li>
              </ul>
              <img src="/images/air-freight/Figure_loading_mail_sacks_airplane_202606071514.jpeg" alt="World Map equivalent" className="af-doc-awb-card-img" style={{borderRadius: 12, objectFit: 'cover', marginTop: '20px', height: '140px'}} /> 
            </div>
          </div>

          <div className="af-doc-cards-row">
            <div className="af-doc-card">
              <div className="af-doc-card-header">
                <div className="af-doc-card-icon green"><DollarSign size={20} /></div>
                <h4 className="af-doc-card-title">Commercial Invoice</h4>
              </div>
              <p>Used to determine cargo value and customs duties.</p>
              <div className="af-doc-card-img-wrap">
                <img src="/images/air-freight/commercial-invoice.png" className="af-doc-card-img-snippet" alt="Invoice snippet" />
              </div>
            </div>
            <div className="af-doc-card">
              <div className="af-doc-card-header">
                <div className="af-doc-card-icon blue"><ClipboardList size={20} /></div>
                <h4 className="af-doc-card-title">Packing List</h4>
              </div>
              <p>Lists package count, dimensions, and weights.</p>
              <div className="af-doc-card-img-wrap">
                <img src="/images/air-freight/package-list.png" className="af-doc-card-img-snippet" alt="Packing list snippet" />
              </div>
            </div>
            <div className="af-doc-card">
              <div className="af-doc-card-header">
                <div className="af-doc-card-icon purple"><Globe size={20} /></div>
                <h4 className="af-doc-card-title">Certificate of Origin</h4>
              </div>
              <p>Confirms manufacturing country.</p>
              <div className="af-doc-card-img-wrap">
                <img src="/images/air-freight/certificate-of-origin.png" className="af-doc-card-img-snippet" alt="Certificate snippet" />
              </div>
            </div>
            <div className="af-doc-card">
              <div className="af-doc-card-header">
                <div className="af-doc-card-icon orange"><User size={20} /></div>
                <h4 className="af-doc-card-title">Customs Declaration</h4>
              </div>
              <p>Required for import/export compliance.</p>
              <div className="af-doc-card-img-wrap">
                <img src="/images/air-freight/customs-declaration.png" className="af-doc-card-img-snippet" alt="Customs snippet" />
              </div>
            </div>
          </div>

          <div className="af-doc-real-card">
            <img src="/images/air-freight/air-cargo.png" alt="Air Cargo" className="af-doc-real-img" />
            <div className="af-doc-real-content">
              <div className="af-doc-real-title">Real Shipment Documentation</div>
              <div className="af-doc-real-route">
                <div className="af-doc-route-point">
                  <div className="af-doc-route-city">Mumbai 🇮🇳</div>
                  <div className="af-doc-route-country">India</div>
                </div>
                <div className="af-doc-route-line"><Plane size={24} /></div>
                <div className="af-doc-route-point" style={{alignItems: 'flex-end'}}>
                  <div className="af-doc-route-city">Frankfurt 🇩🇪</div>
                  <div className="af-doc-route-country">Germany</div>
                </div>
              </div>
              <div className="af-doc-real-cols">
                <div className="af-doc-col-block">
                  <div className="af-doc-col-title">AWB Number</div>
                  <div className="af-doc-awb-badge">176-98765432</div>
                </div>
                <div className="af-doc-col-block">
                  <div className="af-doc-col-title">Documents Submitted</div>
                  <ul className="af-doc-submitted-list">
                    <li><CheckCircle2 size={16} /> Air Waybill</li>
                    <li><CheckCircle2 size={16} /> Commercial Invoice</li>
                    <li><CheckCircle2 size={16} /> Packing List</li>
                    <li><CheckCircle2 size={16} /> Certificate of Origin</li>
                  </ul>
                </div>
                <div className="af-doc-col-block">
                  <div className="af-doc-col-title">Shipment Status</div>
                  <div className="af-doc-status">
                    <div className="af-doc-status-icon"><CheckCircle2 size={24} /></div>
                    <div className="af-doc-status-text">Ready For<br/>Customs Clearance</div>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <div className="af-doc-delays">
            <div className="af-doc-delays-title">95% of shipment delays happen because of:</div>
            <div className="af-doc-delays-items">
              <div className="af-doc-delay-item">
                <div className="af-doc-delay-icon-wrap">
                  <FileText size={20} />
                  <div className="badge"><X size={12} /></div>
                </div>
                <div className="af-doc-delay-text">Missing<br/>Documents</div>
              </div>
              <div className="af-doc-delay-item">
                <div className="af-doc-delay-icon-wrap">
                  <Tag size={20} />
                  <div className="badge"><X size={12} /></div>
                </div>
                <div className="af-doc-delay-text">Incorrect<br/>Values</div>
              </div>
              <div className="af-doc-delay-item">
                <div className="af-doc-delay-icon-wrap">
                  <User size={20} />
                  <div className="badge"><X size={12} /></div>
                </div>
                <div className="af-doc-delay-text">Customs<br/>Errors</div>
              </div>
              <div className="af-doc-delay-item">
                <div className="af-doc-delay-icon-wrap">
                  <ClipboardList size={20} />
                  <div className="badge"><X size={12} /></div>
                </div>
                <div className="af-doc-delay-text">Incomplete<br/>AWBs</div>
              </div>
            </div>
          </div>

          <div className="af-cw-pro-tip" style={{alignSelf: 'stretch', justifyContent: 'center'}}>
            <div className="af-cw-pro-tip-icon"><ShieldCheck size={18} /></div>
            <span><strong>Pro Tip &nbsp;</strong> Always double-check your documents before shipment to avoid delays, demurrage, and unexpected costs.</span>
          </div>

        </div>
      </section>



      {/* ==================== AIRPORT OPERATIONS ==================== */}
      <section className="af-ops-section">
        <div className="af-ops-container">
          <div className="af-ops-header">
            <div className="af-ops-eyebrow">
              <Warehouse size={16} /> AIRPORT OPERATIONS
            </div>
            <h2>How Cargo Moves Through An Airport</h2>
            <p>From warehouse acceptance to aircraft loading, every shipment passes through multiple security, handling and customs checkpoints before takeoff.</p>
          </div>

          {/* 6-Step Flow */}
          <div className="af-ops-steps-container">
            <div className="af-ops-steps-arrow"></div>
            
            {/* Step 1 */}
            <div className="af-ops-step-card">
              <div className="af-ops-step-img-wrap">
                <img src="/images/air-freight/cargo-acceptance.png" alt="Cargo Acceptance" className="af-ops-step-img" />
                <div className="af-ops-step-number">1</div>
              </div>
              <div className="af-ops-step-content">
                <div className="af-ops-step-title-wrap">
                  <Warehouse size={24} className="af-ops-step-title-icon" />
                  <div className="af-ops-step-title">Cargo<br/>Acceptance</div>
                </div>
                <div className="af-ops-step-desc">Cargo arrives at airline warehouse.</div>
                <div className="af-ops-step-list-title">Verified:</div>
                <ul className="af-ops-step-list">
                  <li><Check size={14} /> Documentation</li>
                  <li><Check size={14} /> Packaging</li>
                  <li><Check size={14} /> Weight</li>
                </ul>
              </div>
            </div>

            {/* Step 2 */}
            <div className="af-ops-step-card">
              <div className="af-ops-step-img-wrap">
                <img src="/images/air-freight/cargo-screening.png" alt="Security Screening" className="af-ops-step-img" />
                <div className="af-ops-step-number">2</div>
              </div>
              <div className="af-ops-step-content">
                <div className="af-ops-step-title-wrap">
                  <ShieldCheck size={24} className="af-ops-step-title-icon" />
                  <div className="af-ops-step-title">Security<br/>Screening</div>
                </div>
                <div className="af-ops-step-desc">Cargo passes aviation security.</div>
                <div className="af-ops-step-list-title">Methods:</div>
                <ul className="af-ops-step-list">
                  <li><Check size={14} /> X-Ray</li>
                  <li><Check size={14} /> Explosive Detection</li>
                  <li><Check size={14} /> Physical Inspection</li>
                </ul>
              </div>
            </div>

            {/* Step 3 */}
            <div className="af-ops-step-card">
              <div className="af-ops-step-img-wrap">
                <img src="/images/air-freight/cargo-terminal-processing-automation.png" alt="Terminal Processing" className="af-ops-step-img" />
                <div className="af-ops-step-number">3</div>
              </div>
              <div className="af-ops-step-content">
                <div className="af-ops-step-title-wrap">
                  <Building2 size={24} className="af-ops-step-title-icon" />
                  <div className="af-ops-step-title">Cargo Terminal<br/>Processing</div>
                </div>
                <div className="af-ops-step-desc">Airport cargo terminal prepares shipment.</div>
                <ul className="af-ops-step-list" style={{marginTop: 'auto'}}>
                  <li><Check size={14} /> AWB verification</li>
                  <li><Check size={14} /> Flight assignment</li>
                  <li><Check size={14} /> ULD planning</li>
                </ul>
              </div>
            </div>

            {/* Step 4 */}
            <div className="af-ops-step-card">
              <div className="af-ops-step-img-wrap">
                <img src="/images/air-freight/uld-loading.png" alt="ULD Build-Up" className="af-ops-step-img" />
                <div className="af-ops-step-number">4</div>
              </div>
              <div className="af-ops-step-content">
                <div className="af-ops-step-title-wrap">
                  <Package size={24} className="af-ops-step-title-icon" />
                  <div className="af-ops-step-title">ULD<br/>Build-Up</div>
                </div>
                <div className="af-ops-step-desc">Cargo consolidated into airline containers.</div>
                <div className="af-ops-step-list-title">Examples:</div>
                <ul className="af-ops-step-list">
                  <li><Check size={14} /> LD3 Container</li>
                  <li><Check size={14} /> PMC Pallet</li>
                  <li><Check size={14} /> PAG Pallet</li>
                </ul>
              </div>
            </div>

            {/* Step 5 */}
            <div className="af-ops-step-card">
              <div className="af-ops-step-img-wrap">
                <img src="/images/air-freight/aircraft-loading.png" alt="Aircraft Loading" className="af-ops-step-img" />
                <div className="af-ops-step-number">5</div>
              </div>
              <div className="af-ops-step-content">
                <div className="af-ops-step-title-wrap">
                  <Plane size={24} className="af-ops-step-title-icon" />
                  <div className="af-ops-step-title">Aircraft<br/>Loading</div>
                </div>
                <div className="af-ops-step-desc">Specialized loaders place ULDs into aircraft.</div>
                <ul className="af-ops-step-list" style={{marginTop: 'auto'}}>
                  <li><Check size={14} /> Weight balancing</li>
                  <li><Check size={14} /> Load planning</li>
                  <li><Check size={14} /> Flight manifest validation</li>
                </ul>
              </div>
            </div>

            {/* Step 6 */}
            <div className="af-ops-step-card">
              <div className="af-ops-step-img-wrap">
                <img src="/images/air-freight/departure.png" alt="Departure" className="af-ops-step-img" />
                <div className="af-ops-step-number">6</div>
              </div>
              <div className="af-ops-step-content">
                <div className="af-ops-step-title-wrap">
                  <Rocket size={24} className="af-ops-step-title-icon" />
                  <div className="af-ops-step-title">Departure<br/></div>
                </div>
                <div className="af-ops-step-desc">Cargo departs origin airport.</div>
                <div style={{marginTop: 'auto', textAlign: 'center'}}>
                  <span style={{fontSize: '0.75rem', fontWeight: 800, color: '#3b82f6'}}>Status:</span>
                  <div className="af-ops-step-status-badge">
                    <Plane size={14} /> In Transit
                  </div>
                </div>
              </div>
            </div>
          </div>

          {/* Timeline Banner */}
          <div className="af-ops-timeline-banner">
            <div className="af-ops-timeline-left">
              <div className="af-ops-timeline-title"><Globe size={24} /> Real Cargo Timeline</div>
              <div className="af-ops-timeline-route">
                Mumbai 🇮🇳 <ArrowRight size={16} /> Frankfurt 🇩🇪
              </div>
            </div>
            <div className="af-ops-timeline-right">
              <div className="af-ops-time-point">
                <div className="af-ops-time-icon"><Warehouse size={20} /></div>
                <div className="af-ops-time-val">08:00</div>
                <div className="af-ops-time-label">Cargo Received</div>
              </div>
              <div className="af-ops-time-point">
                <div className="af-ops-time-icon"><ShieldCheck size={20} /></div>
                <div className="af-ops-time-val">10:30</div>
                <div className="af-ops-time-label">Security Cleared</div>
              </div>
              <div className="af-ops-time-point">
                <div className="af-ops-time-icon"><Building2 size={20} /></div>
                <div className="af-ops-time-val">13:00</div>
                <div className="af-ops-time-label">ULD Built</div>
              </div>
              <div className="af-ops-time-point">
                <div className="af-ops-time-icon"><Plane size={20} /></div>
                <div className="af-ops-time-val">16:45</div>
                <div className="af-ops-time-label">Aircraft Loaded</div>
              </div>
              <div className="af-ops-time-point">
                <div className="af-ops-time-icon"><Rocket size={20} /></div>
                <div className="af-ops-time-val">18:00</div>
                <div className="af-ops-time-label">Flight Departed</div>
              </div>
            </div>
          </div>

          {/* Equipment Grid */}
          <div className="af-ops-eq-eyebrow"><Settings size={18} /> CARGO HANDLING EQUIPMENT</div>
          <div className="af-ops-eq-grid">
            <div className="af-ops-eq-card">
              <img src="/images/air-freight/highloader-vehicle.png" alt="High Loader" className="af-ops-eq-img" />
              <div className="af-ops-eq-content">
                <div className="af-ops-eq-icon"><ArrowUpCircle size={24} /></div>
                <div className="af-ops-eq-text">
                  <h4>High Loader</h4>
                  <p>Used to raise cargo to aircraft doors.</p>
                </div>
              </div>
            </div>
            <div className="af-ops-eq-card">
              <img src="/images/air-freight/forklift.png" alt="Forklift" className="af-ops-eq-img" />
              <div className="af-ops-eq-content">
                <div className="af-ops-eq-icon"><Truck size={24} /></div>
                <div className="af-ops-eq-text">
                  <h4>Forklift</h4>
                  <p>Warehouse cargo movement.</p>
                </div>
              </div>
            </div>
            <div className="af-ops-eq-card">
              <img src="/images/air-freight/pallet-dolly.png" alt="Pallet Dolly" className="af-ops-eq-img" />
              <div className="af-ops-eq-content">
                <div className="af-ops-eq-icon"><Truck size={24} /></div>
                <div className="af-ops-eq-text">
                  <h4>Pallet Dolly</h4>
                  <p>Moves ULD pallets across apron.</p>
                </div>
              </div>
            </div>
            <div className="af-ops-eq-card">
              <img src="/images/air-freight/cargo-tug.png" alt="Cargo Tug" className="af-ops-eq-img" />
              <div className="af-ops-eq-content">
                <div className="af-ops-eq-icon"><Truck size={24} /></div>
                <div className="af-ops-eq-text">
                  <h4>Cargo Tug</h4>
                  <p>Pulls cargo dollies around airport.</p>
                </div>
              </div>
            </div>
          </div>

          {/* Alerts Row */}
          <div className="af-ops-alerts-banner">
            <div className="af-ops-alert-title">
              <AlertTriangle size={40} className="af-ops-alert-title-icon" />
              <div className="af-ops-alert-title-text">
                <h4>Why This Matters</h4>
                <p>Incorrect loading can cause:</p>
              </div>
            </div>
            <div className="af-ops-alerts-grid">
              <div className="af-ops-alert-item">
                <Clock size={28} />
                <span>Flight<br/>Delays</span>
              </div>
              <div className="af-ops-alert-item">
                <Package size={28} />
                <span>Cargo<br/>Damage</span>
              </div>
              <div className="af-ops-alert-item">
                <Scale size={28} />
                <span>Aircraft<br/>Imbalance</span>
              </div>
              <div className="af-ops-alert-item">
                <FileWarning size={28} />
                <span>Regulatory<br/>Violations</span>
              </div>
            </div>
          </div>

          {/* Pro Tip Box */}
          <div className="af-ops-pro-tip">
            <div className="af-ops-pro-tip-bg"></div>
            <div className="af-ops-pro-tip-overlay"></div>
            <div className="af-ops-pro-tip-icon"><Star size={24} /></div>
            <div className="af-ops-pro-tip-text">
              <strong>Pro Tip</strong>
              Well-packaged cargo clears airport handling faster and reduces risk of damage.
            </div>
          </div>

        </div>
      </section>

      {/* ==================== TRADE INTELLIGENCE HUB (CTA) ==================== */}
      <section className="af-hub-section">
        <div className="af-hub-container">
          <div className="af-hub-header">
            <div className="af-hub-eyebrow">
              <GraduationCap size={16} /> TRADE INTELLIGENCE HUB
            </div>
            <h2>Master Global Trade Beyond Air Freight</h2>
            <p>Air freight is only one part of international logistics. Explore visual guides, calculators, and trade intelligence tools used by importers, exporters, freight forwarders, and supply chain teams worldwide.</p>
          </div>

          <div className="af-hub-split">
            {/* Left Box */}
            <div className="af-hub-learning">
              <div className="af-hub-learn-icon"><GraduationCap size={24} /></div>
              <h3>Continue Learning Like A Logistics Professional</h3>
              <p>Explore interconnected logistics knowledge hubs covering sea freight, Incoterms, customs, documentation, container shipping, and global trade operations.</p>
              
              <div className="af-hub-learn-btns">
                <Link to="/knowledge/sea-freight" className="af-hub-learn-btn sea">
                  <Ship size={20} /> Explore Sea Freight Guide <ArrowRight size={20} className="arrow" />
                </Link>
                <Link to="/knowledge/incoterms" className="af-hub-learn-btn inco">
                  <BookOpen size={20} /> Learn Incoterms <ArrowRight size={20} className="arrow" />
                </Link>
                <Link to="/tools/calculators" className="af-hub-learn-btn calc">
                  <Calculator size={20} /> Use Freight Calculators <ArrowRight size={20} className="arrow" />
                </Link>
              </div>
            </div>

            {/* Right Map/Orb */}
            <div className="af-hub-orb-area">
              {/* Connecting Lines */}
              <svg className="af-hub-svg-lines" viewBox="0 0 100 100" preserveAspectRatio="none">
                <line x1="50" y1="50" x2="50" y2="15" />
                <line x1="50" y1="50" x2="15" y2="50" />
                <line x1="50" y1="50" x2="85" y2="50" />
                <line x1="50" y1="50" x2="50" y2="85" />
              </svg>

              <div className="af-hub-orb-center">
                <Globe size={32} />
                <span>TRADE<br/>INTELLIGENCE<br/>HUB</span>
              </div>

              {/* Badges */}
              <div className="af-hub-float-badge inco">
                <div className="af-hub-float-header">
                  <div className="af-hub-float-icon"><FileText size={18} /></div>
                  <div className="af-hub-float-title">Incoterms</div>
                </div>
                <div className="af-hub-float-desc">Understand rules, risk transfer and cost responsibilities.</div>
              </div>

              <div className="af-hub-float-badge sea">
                <div className="af-hub-float-header">
                  <div className="af-hub-float-icon"><Ship size={18} /></div>
                  <div className="af-hub-float-title">Sea Freight</div>
                </div>
                <div className="af-hub-float-desc">Learn ocean shipping, containers, FCL/LCL, ports and more.</div>
              </div>

              <div className="af-hub-float-badge air">
                <div className="af-hub-float-header">
                  <div className="af-hub-float-icon"><Plane size={18} /></div>
                  <div className="af-hub-float-title">Air Freight</div>
                </div>
                <div className="af-hub-float-desc">Master air cargo, pricing, documentation and airport operations.</div>
              </div>

              <div className="af-hub-float-badge cust">
                <div className="af-hub-float-header">
                  <div className="af-hub-float-icon"><ShieldCheck size={18} /></div>
                  <div className="af-hub-float-title">Customs</div>
                </div>
                <div className="af-hub-float-desc">Understand customs clearance, duties and global compliance.</div>
              </div>
            </div>
          </div>

          {/* 4 Cards */}
          <div className="af-hub-cards-grid">
            <div className="af-hub-card">
              <div className="af-hub-card-img-wrap">
                <img src="/images/ocean_freight_1780652518875.png" alt="Sea Freight" className="af-hub-card-img" />
                <div className="af-hub-card-icon-badge"><Ship size={20} /></div>
              </div>
              <div className="af-hub-card-content">
                <div className="af-hub-card-title">Sea Freight Guide</div>
                <div className="af-hub-card-desc">Learn containers, FCL/LCL shipping, ports, transit times and ocean freight basics.</div>
                <Link to="/knowledge/sea-freight" className="af-hub-card-link">Explore Guide <ArrowRight size={16} /></Link>
              </div>
            </div>
            
            <div className="af-hub-card">
              <div className="af-hub-card-img-wrap">
                <img src="/images/incoterms/incoterm-hero.png" alt="Incoterms" className="af-hub-card-img" />
                <div className="af-hub-card-icon-badge"><BookOpen size={20} /></div>
              </div>
              <div className="af-hub-card-content">
                <div className="af-hub-card-title">Incoterms Guide</div>
                <div className="af-hub-card-desc">Understand risk transfer, cost responsibilities and global trade rules.</div>
                <Link to="/knowledge/incoterms" className="af-hub-card-link">Explore Guide <ArrowRight size={16} /></Link>
              </div>
            </div>

            <div className="af-hub-card">
              <div className="af-hub-card-img-wrap">
                <img src="/images/logistics_control_tower.png" alt="Calculators" className="af-hub-card-img" />
                <div className="af-hub-card-icon-badge"><Calculator size={20} /></div>
              </div>
              <div className="af-hub-card-content">
                <div className="af-hub-card-title">Logistics Calculators</div>
                <div className="af-hub-card-desc">CBM, volumetric weight, freight cost and transit time calculators for your shipments.</div>
                <Link to="/tools/calculators" className="af-hub-card-link">Open Tools <ArrowRight size={16} /></Link>
              </div>
            </div>

            <div className="af-hub-card">
              <div className="af-hub-card-img-wrap">
                <img src="/images/platform-dashboard.jpg" alt="Intelligence" className="af-hub-card-img" />
                <div className="af-hub-card-icon-badge"><BarChart2 size={20} /></div>
              </div>
              <div className="af-hub-card-content">
                <div className="af-hub-card-title">Trade Intelligence</div>
                <div className="af-hub-card-desc">Access airport codes, container specs, trade data and industry insights & references.</div>
                <Link to="/knowledge/intelligence" className="af-hub-card-link">Explore Resources <ArrowRight size={16} /></Link>
              </div>
            </div>
          </div>

          {/* Progress Banner */}
          <div className="af-hub-progress">
            <div className="af-hub-prog-badge"><Award size={32} /></div>
            <div className="af-hub-prog-content">
              <div className="af-hub-prog-title">You Have Learned</div>
              <div className="af-hub-prog-list">
                <div className="af-hub-prog-item"><CheckCircle2 size={20} /> Air Cargo Journey</div>
                <div className="af-hub-prog-item"><CheckCircle2 size={20} /> Chargeable Weight</div>
                <div className="af-hub-prog-item"><CheckCircle2 size={20} /> Air Freight Documentation</div>
                <div className="af-hub-prog-item"><CheckCircle2 size={20} /> Airport Cargo Operations</div>
              </div>
            </div>
            <div className="af-hub-prog-right">
              <div className="af-hub-prog-right-text">Air Freight Fundamentals Completed</div>
              <div className="af-hub-prog-bar-bg">
                <div className="af-hub-prog-bar-fill"></div>
                <div className="af-hub-prog-pct">100%</div>
              </div>
            </div>
          </div>

          {/* Trusted Banner */}
          <div className="af-hub-trusted">
            <div className="af-hub-trusted-title">Trusted & Used By</div>
            <div className="af-hub-trusted-list">
              <div className="af-hub-trusted-item"><Users size={16} /> Importers</div>
              <div className="af-hub-trusted-item"><Globe size={16} /> Exporters</div>
              <div className="af-hub-trusted-item"><Truck size={16} /> Freight Forwarders</div>
              <div className="af-hub-trusted-item"><Briefcase size={16} /> Supply Chain Teams</div>
              <div className="af-hub-trusted-item"><Target size={16} /> Logistics Professionals</div>
            </div>
          </div>

          {/* Final Large CTA */}
          <div className="af-hub-final-cta">
            <div className="af-hub-final-overlay"></div>
            <div className="af-hub-final-content">
              <h2>Ready To Move Cargo Smarter?</h2>
              <p>Join thousands of businesses using modern logistics intelligence to reduce costs, avoid delays and make better shipping decisions.</p>
              <div className="af-hub-final-btns">
                <Link to="/register" className="af-hub-final-btn primary"><Rocket size={20} /> Get Started</Link>
                <Link to="/contact" className="af-hub-final-btn secondary"><CalendarDays size={20} /> Request Demo</Link>
              </div>
            </div>
          </div>

        </div>
      </section>

    </div>
  );
}

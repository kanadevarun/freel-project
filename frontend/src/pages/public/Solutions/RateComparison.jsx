import React from 'react';
import { Link } from 'react-router-dom';
import { TrendingUp, Globe, Ship, BarChart2, Mail, Phone, FileSpreadsheet, BarChart, Brain, Route, ShieldCheck, ArrowRight, Zap, Anchor, Fuel, CalendarClock, Plane, MapPin, Activity, Truck } from 'lucide-react';
import './RateComparison.css';

export default function RateComparison() {
  return (
    <div className="rc-page">
      <section className="rc-hero">
        {/* Background Elements */}
        <div className="rc-hero-bg">
          <div className="rc-glow rc-glow-1"></div>
          <div className="rc-glow rc-glow-2"></div>
          <div className="rc-world-lines"></div>
        </div>

        <div className="rc-container">
          <div className="rc-hero-grid">
            
            {/* Left Content */}
            <div className="rc-hero-content">
              <div className="rc-badge">
                <span className="rc-pulse"></span>
                LIVE FREIGHT MARKET INTELLIGENCE
              </div>
              
              <h1 className="rc-headline">
                Compare Freight Rates<br />
                <span className="rc-gradient-text">Before The Market Moves.</span>
              </h1>
              
              <p className="rc-subheadline">
                Monitor live Road, Air and Ocean freight pricing across hundreds of carriers. Compare rates instantly, protect margins, and make smarter procurement decisions.
              </p>
              
              <div className="rc-cta-group">
                <Link to="/contact" className="rc-btn-primary">
                  Compare Rates Now →
                </Link>
                <Link to="/contact" className="rc-btn-secondary">
                  Watch Live Demo
                </Link>
              </div>
              
              <div className="rc-trust-indicators">
                <div className="rc-trust-item">
                  <span className="rc-check">✓</span> Live Market Data
                </div>
                <div className="rc-trust-item">
                  <span className="rc-check">✓</span> Margin Protection
                </div>
                <div className="rc-trust-item">
                  <span className="rc-check">✓</span> Multi-Modal Coverage
                </div>
              </div>
            </div>

            {/* Right Visual Composition */}
            <div className="rc-hero-visual">
              <div className="rc-composition">
                
                {/* Images */}
                <div className="rc-img-wrapper rc-img-port">
                  <img src="https://images.unsplash.com/photo-1578575437130-527eed3abbec?auto=format&fit=crop&w=800&q=80" alt="Global shipping container port" />
                </div>
                <div className="rc-img-wrapper rc-img-air">
                  <img src="https://images.unsplash.com/photo-1436491865332-7a61a109cc05?auto=format&fit=crop&w=800&q=80" alt="Cargo aircraft" />
                </div>
                <div className="rc-img-wrapper rc-img-truck">
                  <img src="https://images.unsplash.com/photo-1601584115197-04ecc0da31d7?auto=format&fit=crop&w=800&q=80" alt="Freight truck" />
                </div>

                {/* Floating Market Cards */}
                <div className="rc-float-card rc-card-1">
                  <div className="rc-card-lane">Mumbai → Dubai</div>
                  <div className="rc-card-data">
                    <span className="rc-card-price">₹95/kg</span>
                    <span className="rc-card-trend rc-down">↓ 2.4%</span>
                  </div>
                </div>

                <div className="rc-float-card rc-card-2">
                  <div className="rc-card-lane">Shanghai → Rotterdam</div>
                  <div className="rc-card-data">
                    <span className="rc-card-price">$4,820</span>
                    <span className="rc-card-trend rc-up">↑ 3.1%</span>
                  </div>
                </div>

                <div className="rc-float-card rc-card-3">
                  <div className="rc-card-lane">Singapore → New York</div>
                  <div className="rc-card-data">
                    <span className="rc-card-price">$5,240</span>
                    <span className="rc-card-trend rc-down">↓ 1.7%</span>
                  </div>
                </div>

                <div className="rc-float-card rc-card-4">
                  <div className="rc-card-label">Best Carrier Today</div>
                  <div className="rc-card-value">MSC Shipping</div>
                  <div className="rc-card-sub rc-success">Savings Opportunity 14.2%</div>
                </div>

                <div className="rc-float-card rc-card-5">
                  <div className="rc-card-label">Market Update</div>
                  <div className="rc-card-value">Fuel Index Rising</div>
                  <div className="rc-card-sub rc-danger">+3.4%</div>
                </div>

              </div>
            </div>

          </div>
        </div>
      </section>

      {/* ─── SECTION 1: THE PROBLEM & VISUAL COMPARISON ─── */}
      <section className="rc-problem-section">
        <div className="rc-container">
          <div className="rc-problem-header">
            <div className="rc-problem-label">WHY FREIGHT TEAMS MISS THE MARKET</div>
            <h2 className="rc-problem-heading">
              You Can't Buy Smart<br />
              If You're Buying Blind.
            </h2>
            <p className="rc-problem-desc">
              Most freight teams still rely on calls, emails and spreadsheets to collect rates.<br />
              By the time quotes arrive, the market has already changed.
            </p>
          </div>

          <div className="rc-problem-comparison">
            
            {/* Left Card: Traditional Procurement */}
            <div className="rc-prob-card rc-prob-left">
              <h3 className="rc-prob-card-title">Traditional Procurement</h3>
              <div className="rc-prob-img-box">
                <img src="https://images.unsplash.com/photo-1542744173-8e7e53415bb0?auto=format&fit=crop&w=1200&q=80" alt="Traditional Procurement" />
                <div className="rc-prob-chip rc-pchip-1">📞 18 Vendor Calls</div>
                <div className="rc-prob-chip rc-pchip-2">📧 43 Emails Waiting</div>
                <div className="rc-prob-chip rc-pchip-3">📊 12 Spreadsheet Versions</div>
                <div className="rc-prob-chip rc-pchip-4">⏳ 2 Days For Quotes</div>
              </div>
              <div className="rc-prob-card-msg">The market moved before<br/>the quotes arrived.</div>
            </div>

            {/* Right Card: Freel Intelligence */}
            <div className="rc-prob-card rc-prob-right">
              <h3 className="rc-prob-card-title">Freel Intelligence</h3>
              <div className="rc-prob-img-box">
                <img src="https://images.unsplash.com/photo-1551288049-bebda4e38f71?auto=format&fit=crop&w=1200&q=80" alt="Freel Intelligence" />
                <div className="rc-prob-chip rc-pchip-5">✅ 200+ Live Rates</div>
                <div className="rc-prob-chip rc-pchip-6">✅ Updated Hourly</div>
                <div className="rc-prob-chip rc-pchip-7">✅ Best Carrier Identified</div>
                <div className="rc-prob-chip rc-pchip-8">✅ Margin Protected</div>
              </div>
              <div className="rc-prob-card-msg">See the market before<br/>you buy from it.</div>
            </div>

          </div>
        </div>
      </section>


      {/* ─── SECTION 3: LIVE MARKET INTELLIGENCE ─── */}
      <section className="rc-live-section">
        <div className="rc-container">
          <div className="rc-live-header">
            <div className="rc-live-label">GLOBAL MARKET INTELLIGENCE</div>
            <h2 className="rc-live-heading">
              See Global Freight Markets<br />
              In Real Time.
            </h2>
            <p className="rc-live-desc">
              Track lane performance, carrier pricing, fuel movements,<br />
              port congestion and demand shifts.
            </p>
          </div>

          <div className="rc-live-hero-img-box">
            <img src="https://images.unsplash.com/photo-1451187580459-43490279c0fa?auto=format&fit=crop&w=1600&q=80" alt="Global Market Intelligence" />
            
            {/* New Design Floating Insight Chips */}
            <div className="rc-live-fcard rc-lf-1">
              <div className="rc-lf-header">Shanghai → Rotterdam</div>
              <div className="rc-lf-body">
                <span className="rc-lf-trend rc-lf-down">▼ 4.2%</span>
                <span className="rc-lf-sub">This Week</span>
              </div>
            </div>
            
            <div className="rc-live-fcard rc-lf-2">
              <div className="rc-lf-header">Mumbai → Dubai</div>
              <div className="rc-lf-body">
                <span className="rc-lf-trend rc-lf-up">▲ 1.8%</span>
                <span className="rc-lf-sub">This Week</span>
              </div>
            </div>

            <div className="rc-live-fcard rc-lf-3">
              <div className="rc-lf-icon">⛽</div>
              <div className="rc-lf-text-col">
                <div className="rc-lf-header">Fuel Index</div>
                <div className="rc-lf-trend rc-lf-down">▼ 1.8%</div>
              </div>
            </div>

            <div className="rc-live-fcard rc-lf-4">
              <div className="rc-lf-icon">🚚</div>
              <div className="rc-lf-text-col">
                <div className="rc-lf-header">India Road Freight</div>
                <div className="rc-lf-trend rc-lf-stable">Stable</div>
              </div>
            </div>

            <div className="rc-live-fcard rc-lf-5">
              <div className="rc-lf-icon">⚓</div>
              <div className="rc-lf-text-col">
                <div className="rc-lf-header">Port Congestion</div>
                <div className="rc-lf-trend rc-lf-warn">Moderate</div>
              </div>
            </div>
          </div>

          <div className="rc-live-grid">
            <div className="rc-lgrid-card">
              <div className="rc-lgrid-img rc-ocean-accent">
                <img src="https://images.pexels.com/photos/2226458/pexels-photo-2226458.jpeg?auto=compress&cs=tinysrgb&w=800" alt="Ocean Freight" />
              </div>
              <div className="rc-lgrid-content">
                <h3>Ocean Freight</h3>
                <p>Live visibility into major port pairs, vessel schedules, and spot vs contract variance.</p>
              </div>
            </div>
            <div className="rc-lgrid-card">
              <div className="rc-lgrid-img rc-air-accent">
                <img src="https://images.unsplash.com/photo-1542296332-2e4473faf563?auto=format&fit=crop&w=800&q=80" alt="Air Freight" />
              </div>
              <div className="rc-lgrid-content">
                <h3>Air Freight</h3>
                <p>Track capacity constraints, peak season surcharges, and real-time cargo rates.</p>
              </div>
            </div>
            <div className="rc-lgrid-card">
              <div className="rc-lgrid-img rc-road-accent">
                <img src="https://images.unsplash.com/photo-1519003722824-194d4455a60c?auto=format&fit=crop&w=800&q=80" alt="Road Freight" />
              </div>
              <div className="rc-lgrid-content">
                <h3>Road Freight</h3>
                <p>Monitor local carrier availability, fuel indexing, and domestic linehaul costs.</p>
              </div>
            </div>
            <div className="rc-lgrid-card">
              <div className="rc-lgrid-img rc-signals-accent">
                <img src="https://images.unsplash.com/photo-1551288049-bebda4e38f71?auto=format&fit=crop&w=800&q=80" alt="Market Signals" />
              </div>
              <div className="rc-lgrid-content">
                <h3>Market Signals</h3>
                <p>Predictive indicators warning you of congestion, strikes, or price volatility.</p>
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* ─── SECTION: BUSINESS IMPACT TRANSFORMATION ─── */}
      <section className="rc-impact-section">
        <div className="rc-container">

          {/* Header */}
          <div className="rc-impact-header">
            <div className="rc-impact-label">BUSINESS IMPACT</div>
            <h2 className="rc-impact-heading">
              Stop Chasing Rates.<br />
              Start Controlling Spend.
            </h2>
            <p className="rc-impact-desc">
              Freel helps procurement teams buy at the right time,<br />
              from the right carrier, at the right price.
            </p>
          </div>

          {/* Transformation Journey */}
          <div className="rc-journey">

            {/* Stage 1 */}
            <div className="rc-journey-stage rc-stage-reactive">
              <div className="rc-stage-icon-wrap rc-accent-red">
                <Mail size={28} strokeWidth={1.5} />
              </div>
              <div className="rc-stage-num">01</div>
              <h3 className="rc-stage-title">Reactive Procurement</h3>
              <div className="rc-stage-pills">
                <span className="rc-pill rc-pill-red"><Phone size={14} /> Calls</span>
                <span className="rc-pill rc-pill-red"><Mail size={14} /> Emails</span>
                <span className="rc-pill rc-pill-red"><FileSpreadsheet size={14} /> Spreadsheets</span>
              </div>
              <p className="rc-stage-caption">Waiting.<br />Comparing.<br />Reacting.</p>
            </div>

            {/* Connector */}
            <div className="rc-journey-connector">
              <svg viewBox="0 0 120 40" preserveAspectRatio="none">
                <path className="rc-connector-path" d="M0,20 C40,5 80,35 120,20" />
              </svg>
              <ArrowRight size={18} className="rc-connector-arrow" />
            </div>

            {/* Stage 2 */}
            <div className="rc-journey-stage rc-stage-intelligence">
              <div className="rc-stage-icon-wrap rc-accent-blue">
                <BarChart size={28} strokeWidth={1.5} />
              </div>
              <div className="rc-stage-num">02</div>
              <h3 className="rc-stage-title">Market Intelligence</h3>
              <div className="rc-stage-pills">
                <span className="rc-pill rc-pill-blue"><TrendingUp size={14} /> Live Data</span>
                <span className="rc-pill rc-pill-blue"><Globe size={14} /> Lane Visibility</span>
                <span className="rc-pill rc-pill-blue"><Zap size={14} /> Carrier Intel</span>
              </div>
              <p className="rc-stage-caption">See the market as it moves.</p>
            </div>

            {/* Connector */}
            <div className="rc-journey-connector">
              <svg viewBox="0 0 120 40" preserveAspectRatio="none">
                <path className="rc-connector-path" d="M0,20 C40,5 80,35 120,20" />
              </svg>
              <ArrowRight size={18} className="rc-connector-arrow" />
            </div>

            {/* Stage 3 */}
            <div className="rc-journey-stage rc-stage-decisions">
              <div className="rc-stage-icon-wrap rc-accent-purple">
                <Brain size={28} strokeWidth={1.5} />
              </div>
              <div className="rc-stage-num">03</div>
              <h3 className="rc-stage-title">Smarter Decisions</h3>
              <div className="rc-stage-pills">
                <span className="rc-pill rc-pill-purple"><BarChart2 size={14} /> Comparison</span>
                <span className="rc-pill rc-pill-purple"><Ship size={14} /> Carrier Select</span>
                <span className="rc-pill rc-pill-purple"><Route size={14} /> Route Optim.</span>
              </div>
              <p className="rc-stage-caption">Buy with confidence.</p>
            </div>

            {/* Connector */}
            <div className="rc-journey-connector">
              <svg viewBox="0 0 120 40" preserveAspectRatio="none">
                <path className="rc-connector-path" d="M0,20 C40,5 80,35 120,20" />
              </svg>
              <ArrowRight size={18} className="rc-connector-arrow" />
            </div>

            {/* Stage 4 */}
            <div className="rc-journey-stage rc-stage-margins">
              <div className="rc-stage-icon-wrap rc-accent-green">
                <ShieldCheck size={28} strokeWidth={1.5} />
              </div>
              <div className="rc-stage-num">04</div>
              <h3 className="rc-stage-title">Protected Margins</h3>
              <div className="rc-stage-pills">
                <span className="rc-pill rc-pill-green"><TrendingUp size={14} /> Growth</span>
                <span className="rc-pill rc-pill-green"><ShieldCheck size={14} /> Margin Safe</span>
                <span className="rc-pill rc-pill-green"><BarChart size={14} /> Profitability</span>
              </div>
              <p className="rc-stage-caption">Better procurement outcomes.</p>
            </div>

          </div>

          {/* Bottom Metrics Bar */}
          <div className="rc-impact-metrics">
            <div className="rc-impact-metric">
              <Globe size={22} className="rc-metric-icon" />
              <strong>384+</strong>
              <span>Active Trade Lanes</span>
            </div>
            <div className="rc-impact-metric">
              <Ship size={22} className="rc-metric-icon" />
              <strong>200+</strong>
              <span>Carrier Integrations</span>
            </div>
            <div className="rc-impact-metric">
              <TrendingUp size={22} className="rc-metric-icon" />
              <strong>12,000+</strong>
              <span>Daily Updates</span>
            </div>
            <div className="rc-impact-metric">
              <BarChart2 size={22} className="rc-metric-icon" />
              <strong>98.7%</strong>
              <span>Market Coverage</span>
            </div>
          </div>

        </div>
      </section>

      {/* ─── SECTION: MARKET DRIVERS (UNIFIED INTELLIGENCE) ─── */}
      <section className="rc-drivers-section">
        <div className="rc-container">
          
          {/* Header */}
          <div className="rc-drivers-header">
            <div className="rc-drivers-label">MARKET DRIVERS</div>
            <h2 className="rc-drivers-heading">
              What Actually Moves Freight Prices?
            </h2>
            <p className="rc-drivers-desc">
              Rates change because markets move.<br />
              Freel helps you understand why.
            </p>
          </div>

          {/* Unified Intelligence Dashboard */}
          <div className="rc-unified-dashboard">
            
            <div className="rc-unified-main">
              {/* Left Side (40%) */}
              <div className="rc-unified-left">
                <h3 className="rc-unified-title">The Signals Behind Every Rate</h3>
                <p className="rc-unified-text">
                  Freight pricing is influenced by fuel costs, port congestion, cargo capacity, global trade events and seasonal demand.
                </p>
                <p className="rc-unified-text">
                  Freel tracks these signals continuously so teams can understand not just the price, but the reason behind it.
                </p>
                
                <div className="rc-signal-chips">
                  <span className="rc-signal-chip">Fuel Prices</span>
                  <span className="rc-signal-chip">Port Congestion</span>
                  <span className="rc-signal-chip">Air Capacity</span>
                  <span className="rc-signal-chip">Seasonal Demand</span>
                </div>
              </div>

              {/* Right Side (60%) */}
              <div className="rc-unified-right">
                {/* Premium Analytics / Command Center Image */}
                <img src="https://images.unsplash.com/photo-1526628953301-3e589a6a8b74?auto=compress&cs=tinysrgb&w=1200" alt="Supply Chain Command Center" className="rc-unified-bg" />
                
                {/* Subtle vignette overlay instead of heavy dark overlay */}
                <div className="rc-unified-overlay"></div>

                {/* Central Hub */}
                <div className="rc-intel-hub">
                  <Brain size={24} className="rc-hub-icon" />
                  <span>Market Intelligence</span>
                  <div className="rc-hub-glow"></div>
                </div>

                {/* SVG Connections with animateMotion for moving dots */}
                <svg className="rc-unified-routes" viewBox="0 0 100 100" preserveAspectRatio="none">
                  {/* Paths */}
                  <path id="path-1" className="rc-route-path" d="M 20 20 Q 40 20 50 50" />
                  <path id="path-2" className="rc-route-path" d="M 80 25 Q 70 50 50 50" />
                  <path id="path-3" className="rc-route-path" d="M 15 50 Q 30 70 50 50" />
                  <path id="path-4" className="rc-route-path" d="M 85 60 Q 60 80 50 50" />
                  <path id="path-5" className="rc-route-path" d="M 50 80 Q 30 65 50 50" />

                  {/* Moving Dots */}
                  <circle r="0.8" className="rc-route-dot">
                    <animateMotion dur="2.5s" repeatCount="indefinite" path="M 20 20 Q 40 20 50 50" />
                  </circle>
                  <circle r="0.8" className="rc-route-dot">
                    <animateMotion dur="3.2s" repeatCount="indefinite" path="M 80 25 Q 70 50 50 50" />
                  </circle>
                  <circle r="0.8" className="rc-route-dot">
                    <animateMotion dur="2.8s" repeatCount="indefinite" path="M 15 50 Q 30 70 50 50" />
                  </circle>
                  <circle r="0.8" className="rc-route-dot">
                    <animateMotion dur="3.5s" repeatCount="indefinite" path="M 85 60 Q 60 80 50 50" />
                  </circle>
                  <circle r="0.8" className="rc-route-dot">
                    <animateMotion dur="2.2s" repeatCount="indefinite" path="M 50 80 Q 30 65 50 50" />
                  </circle>
                </svg>

                {/* Floating Intelligence Cards positioned at exact SVG start points */}
                <div className="rc-float-intel rc-intel-card-1">
                  <div className="rc-intel-label">Fuel Index</div>
                  <div className="rc-intel-val rc-val-green">▼ 1.8%</div>
                </div>

                <div className="rc-float-intel rc-intel-card-2">
                  <div className="rc-intel-label">Port Congestion</div>
                  <div className="rc-intel-val rc-val-red">High</div>
                </div>

                <div className="rc-float-intel rc-intel-card-3">
                  <div className="rc-intel-label">Air Capacity</div>
                  <div className="rc-intel-val rc-val-blue">82%</div>
                </div>

                <div className="rc-float-intel rc-intel-card-4">
                  <div className="rc-intel-label">Market Demand</div>
                  <div className="rc-intel-val rc-val-orange">Peak Season</div>
                </div>

                <div className="rc-float-intel rc-intel-card-5">
                  <div className="rc-intel-label">Global Trade Impact</div>
                  <div className="rc-intel-val rc-val-purple">Medium</div>
                </div>

              </div>
            </div>
            
            {/* Bottom Strip inside the container */}
            <div className="rc-unified-bottom">
              <div className="rc-ub-metric">
                <strong>384</strong>
                <span>Global Trade Lanes</span>
              </div>
              <div className="rc-ub-metric">
                <strong>200+</strong>
                <span>Carrier Sources</span>
              </div>
              <div className="rc-ub-metric">
                <strong>12,000+</strong>
                <span>Rate Updates Daily</span>
              </div>
              <div className="rc-ub-metric">
                <strong>98.7%</strong>
                <span>Market Coverage</span>
              </div>
            </div>

          </div>

        </div>
      </section>

      {/* ─── SECTION: PLATFORM OVERVIEW ─── */}
      <section className="rc-platform-overview">
        <div className="rc-container">
          
          <div className="rc-platform-header">
            <div className="rc-platform-label">PLATFORM OVERVIEW</div>
            <h2 className="rc-platform-heading">
              The Operating System<br />
              For Freight Procurement.
            </h2>
            <p className="rc-platform-subheading">
              Track rates, capacity, congestion, fuel movements and market shifts<br />
              from a single intelligence platform.
            </p>
          </div>

          <div className="rc-platform-visual">
            
            {/* SVG Connections for the network feel */}
            <svg className="rc-platform-lines" viewBox="0 0 1200 800" preserveAspectRatio="xMidYMid slice">
              <path className="rc-pline" d="M 200 200 Q 600 150 600 400" />
              <path className="rc-pline" d="M 1000 200 Q 600 150 600 400" />
              <path className="rc-pline" d="M 150 600 Q 600 800 600 400" />
              <path className="rc-pline" d="M 1050 600 Q 600 800 600 400" />
              <path className="rc-pline" d="M 100 400 Q 600 400 600 400" />
            </svg>

            {/* Desktop Monitor Mockup */}
            <div className="rc-monitor-mockup">
              <div className="rc-monitor-frame">
                <div className="rc-monitor-screen">
                  <img src="/images/platform-dashboard.jpg" alt="Platform Dashboard Overview" />
                </div>
              </div>
              <div className="rc-monitor-stand"></div>
              <div className="rc-monitor-base"></div>
            </div>

            {/* Floating Insight Cards */}
            <div className="rc-float-card rc-fc-1">
              <MapPin size={16} className="rc-fc-icon" /> 384 Active Lanes
            </div>
            <div className="rc-float-card rc-fc-2">
              <Ship size={16} className="rc-fc-icon" /> 200+ Carrier Sources
            </div>
            <div className="rc-float-card rc-fc-3">
              <Activity size={16} className="rc-fc-icon" /> 12,000+ Daily Updates
            </div>
            <div className="rc-float-card rc-fc-4">
              <Globe size={16} className="rc-fc-icon" /> 98.7% Market Coverage
            </div>
            
            <div className="rc-float-card rc-fc-5">
              <Fuel size={16} className="rc-fc-icon" /> Fuel Intelligence
            </div>
            <div className="rc-float-card rc-fc-6">
              <Anchor size={16} className="rc-fc-icon" /> Port Visibility
            </div>
            <div className="rc-float-card rc-fc-7">
              <Plane size={16} className="rc-fc-icon" /> Air Capacity Tracking
            </div>
            <div className="rc-float-card rc-fc-8">
              <Truck size={16} className="rc-fc-icon" /> Road Market Trends
            </div>
            <div className="rc-float-card rc-fc-9">
              <TrendingUp size={16} className="rc-fc-icon" /> Lane Benchmarking
            </div>
            <div className="rc-float-card rc-fc-10">
              <BarChart2 size={16} className="rc-fc-icon" /> Global Trade Monitoring
            </div>

          </div>

          <div className="rc-platform-footer">
            Freight doesn't move in silos.<br />
            Neither should your data.
          </div>

        </div>
      </section>

      {/* ─── IMMERSIVE LOGISTICS DASHBOARD CTA ─── */}
      <section className="rc-immersive-cta">
        <div className="rc-imm-container">
          
          {/* Text Content (Top Left constraint) */}
          <div className="rc-imm-content">
            <h2 className="rc-imm-heading">
              See The Market Before<br />
              You Buy From It.
            </h2>
            <p className="rc-imm-desc">
              Track live freight movements, carrier pricing,<br />
              fuel trends, and lane performance before<br />
              making procurement decisions.
            </p>
            <button className="rc-btn-primary">Explore Live Market Data →</button>
          </div>

          {/* Massive Immersive Image Layer */}
          <div className="rc-imm-visual">
            <img src="https://images.pexels.com/photos/1095814/pexels-photo-1095814.jpeg?auto=compress&cs=tinysrgb&w=1600" alt="Global Logistics Network" />
            
            {/* Animated Curved Routes SVG */}
            <svg className="rc-imm-routes" viewBox="0 0 1000 600" preserveAspectRatio="none">
              <path className="rc-imm-path" d="M100,100 Q400,50 600,200 T900,150" />
              <path className="rc-imm-path" d="M200,400 Q500,500 800,300" />
              
              {/* Route Labels */}
              <text className="rc-route-label" x="400" y="80">Shanghai → Rotterdam</text>
              <text className="rc-route-label" x="600" y="420">Mumbai → Singapore</text>
            </svg>

            {/* Floating Location Markers */}
            <div className="rc-imm-marker rc-im-shanghai">📍 Shanghai</div>
            <div className="rc-imm-marker rc-im-singapore">📍 Singapore</div>
            <div className="rc-imm-marker rc-im-rotterdam">📍 Rotterdam</div>
            <div className="rc-imm-marker rc-im-mumbai">📍 Mumbai</div>

            {/* Market Intelligence Glass Cards */}
            <div className="rc-imm-glass rc-ig-1">
              <span className="rc-ig-icon"><TrendingUp size={24} /></span>
              <strong>12,000+</strong>
              Daily Market Updates
            </div>
            <div className="rc-imm-glass rc-ig-2">
              <span className="rc-ig-icon"><Globe size={24} /></span>
              <strong>384</strong>
              Global Trade Lanes
            </div>
            <div className="rc-imm-glass rc-ig-3">
              <span className="rc-ig-icon"><Ship size={24} /></span>
              <strong>200+</strong>
              Active Carriers
            </div>
            <div className="rc-imm-glass rc-ig-4">
              <span className="rc-ig-icon"><BarChart2 size={24} /></span>
              <strong>98.7%</strong>
              Market Coverage
            </div>

            {/* Intelligence Signal Cards */}
            <div className="rc-imm-glass rc-ig-signal-1">
              <div className="rc-ig-title">Fuel Index</div>
              <div className="rc-ig-trend rc-ig-down">▼ 1.8%</div>
            </div>
            <div className="rc-imm-glass rc-ig-signal-2">
              <div className="rc-ig-title">Ocean Rates</div>
              <div className="rc-ig-trend rc-ig-down">▼ 2.3%</div>
            </div>

          </div>
        </div>
      </section>

    </div>
  );
}

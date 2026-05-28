import { useState } from 'react';
import { Link } from 'react-router-dom';
import { PageLoader, ScrollProgress, LogisticsCanvas, StaggerTitle, TiltCard, MagneticBtn, Timeline, StickyStackCards, Reveal, Counter, SectionHeader, StickyShowcase } from './LandingComponents';
import { Logistics3D } from './Logistics3D';
import './Landing.css';

export default function Landing() {
  const [activeTab, setActiveTab] = useState('rates');

  return (
    <>
      <PageLoader />
      <ScrollProgress />

      {/* ═══ HERO — Full Width 3D WebGL Background ═══ */}
      <section className="hero-section">
        <div className="hero-3d-background">
          <Logistics3D />
        </div>
        <div className="hero-glow" />
        <div className="hero-content">
          <div className="hero-badge"><span className="pulse-dot" /> Now in Early Access</div>
          <StaggerTitle line1="The Operating System for" line2="Modern Logistics" />
          <p className="hero-sub">
            Compare freight rates across Road, Air &amp; Sea. Track shipments in real-time.
            Manage your entire supply chain from one powerful command center.
          </p>
          <div className="hero-cta-row">
            <MagneticBtn><Link to="/signup" className="btn-primary">Start Free Trial →</Link></MagneticBtn>
            <MagneticBtn><Link to="/contact" className="btn-secondary">Request Demo</Link></MagneticBtn>
          </div>
          <div className="stats-bar">
            <div className="stat-pill"><strong>500+</strong> Verified Vendors</div>
            <div className="stat-pill"><strong>3</strong> Transport Modes</div>
            <div className="stat-pill"><strong>150+</strong> Global Ports</div>
            <div className="stat-pill"><strong>10K+</strong> Trade Lanes</div>
          </div>
        </div>
      </section>

      {/* ═══ TRUST LOGOS — Marquee ═══ */}
      <Reveal>
        <section className="py-12 trust-section">
          <div className="container-lg">
            <p className="text-center text-xs font-semibold text-slate-400 uppercase tracking-widest mb-8">Trusted by logistics leaders across India</p>
            <div style={{ overflow: 'hidden' }}>
              <div className="trust-row">
                {['SafeExpress', 'TCI Freight', 'VRL Logistics', 'Gati KWE', 'Delhivery', 'BlueDart', 'Rivigo', 'SafeExpress', 'TCI Freight', 'VRL Logistics', 'Gati KWE', 'Delhivery', 'BlueDart', 'Rivigo'].map((n, i) => (
                  <span key={i} className="trust-logo">{n}</span>
                ))}
              </div>
            </div>
          </div>
        </section>
      </Reveal>

      {/* ═══ ANIMATED COUNTERS ═══ */}
      <Reveal>
        <section className="counters-section">
          <div className="container-lg">
            <div className="counters-grid">
              <Counter end={500} suffix="+" label="Verified Transporters" />
              <Counter end={15} suffix="L Cr" label="Freight Managed" />
              <Counter end={98} suffix="%" label="On-Time Delivery" />
              <Counter end={35} suffix="%" label="Cost Savings Avg" />
            </div>
          </div>
        </section>
      </Reveal>

      {/* ═══ SERVICES — 3D Tilt Cards ═══ */}
      <section className="section-padding bg-white border-t border-slate-100">
        <div className="container-lg">
          <Reveal>
            <SectionHeader label="Our Services" title="Move Freight Across Any Mode"
              subtitle="One platform for road, air, and sea logistics — with full customs compliance built in." />
          </Reveal>
          <div className="services-grid">
            {[
              { icon: '✈️', bg: 'bg-blue-50', title: 'Air Freight', desc: 'Global air cargo with IATA-certified agents. Ship general, hazardous, pharma, and perishable cargo to 150+ airports.', s1: '150+', s1l: 'Airports', s2: '24hr', s2l: 'Express' },
              { icon: '🚢', bg: 'bg-cyan-50', title: 'Sea Freight', desc: 'FCL & LCL shipments via top shipping lines. 20ft, 40ft, Reefer containers with HAZ cargo support.', s1: '50+', s1l: 'Shipping Lines', s2: '7-25', s2l: 'Days Transit' },
              { icon: '🚛', bg: 'bg-amber-50', title: 'Road Transport', desc: 'Full Truck Load & Part Load across India with GPS tracking. 500+ verified transporters with real-time visibility.', s1: '500+', s1l: 'Transporters', s2: 'Pan India', s2l: 'Coverage' },
            ].map((s, i) => (
              <Reveal key={i} delay={`reveal-delay-${i + 1}`}>
                <TiltCard>
                  <div className="service-card">
                    <div className={`service-icon ${s.bg}`}>{s.icon}</div>
                    <h3 className="text-xl font-bold text-brand-navy mb-3">{s.title}</h3>
                    <p className="text-slate-600 text-sm leading-relaxed mb-6">{s.desc}</p>
                    <div className="service-stats">
                      <div><span className="font-bold text-brand-navy">{s.s1}</span> <span className="text-slate-500">{s.s1l}</span></div>
                      <div className="w-px bg-slate-200"></div>
                      <div><span className="font-bold text-brand-navy">{s.s2}</span> <span className="text-slate-500">{s.s2l}</span></div>
                    </div>
                  </div>
                </TiltCard>
              </Reveal>
            ))}
          </div>
        </div>
      </section>

      {/* ═══ HOW IT WORKS — Timeline ═══ */}
      <section className="section-padding timeline-section">
        <div className="container-lg">
          <Reveal>
            <SectionHeader label="How It Works" title="Simplified Logistics in 4 Steps"
              subtitle="From onboarding to delivery, experience a seamless freight process." />
          </Reveal>
          <Timeline />
        </div>
      </section>

      {/* ═══ MULTI-MODAL DOMINANCE (3D Stacking Cards) ═══ */}
      <section className="section-padding bg-slate-50" style={{ paddingBottom: 0 }}>
        <div className="container-lg">
          <Reveal>
            <SectionHeader label="Global Reach" title="Multi-Modal Dominance"
              subtitle="Move freight across any mode with unparalleled visibility and control." />
          </Reveal>
        </div>

        <div style={{ width: '100%', padding: '0 20px', maxWidth: '1600px', margin: '0 auto' }}>
          <StickyStackCards cards={[
            {
              title: "Road Transport Network",
              desc: "Full Truck Load (FTL) and Part Truck Load (PTL) across India with live GPS tracking. Access over 500+ verified transporters for maximum reliability.",
              src: "/images/hero_logistics_hub_1779795325054.png"
            },
            {
              title: "Global Ocean Freight",
              desc: "FCL & LCL shipments via top shipping lines. Specialized containers with complete HAZ cargo support and digital customs clearance.",
              src: "/images/hero_cargo_ship_1779795287924.png"
            },
            {
              title: "Express Air Freight",
              desc: "Global air cargo with IATA-certified agents. Ship general, hazardous, pharma, and perishable cargo to 150+ airports instantly.",
              src: "/images/hero_cargo_plane_1779795307360.png"
            }
          ]} />
        </div>
      </section>

      {/* ═══ PLATFORM PREVIEW ═══ */}
      <section className="section-padding bg-white">
        <div className="container-md">
          <Reveal>
            <SectionHeader label="Platform Preview" title="Your Freight Command Center"
              subtitle="Everything you need to run a high-performance logistics business." />
          </Reveal>
          <Reveal>
            <div className="platform-tabs">
              {[
                { id: 'rates', label: '📈 Rate Comparison' },
                { id: 'tracking', label: '📍 Live Tracking' },
                { id: 'rfq', label: '📋 RFQ System' },
                { id: 'compliance', label: '🛡️ Compliance' },
              ].map(t => (
                <button key={t.id} className={`platform-tab ${activeTab === t.id ? 'active' : ''}`}
                  onClick={() => setActiveTab(t.id)}>{t.label}</button>
              ))}
            </div>
          </Reveal>
          <Reveal>
            <div className="platform-browser">
              <div className="platform-bar">
                <span className="platform-dot" style={{ background: '#FF5F56' }} />
                <span className="platform-dot" style={{ background: '#FFBD2E' }} />
                <span className="platform-dot" style={{ background: '#27C93F' }} />
                <span className="text-xs text-slate-400 ml-4">🔒 freel.in/dashboard/{activeTab}</span>
              </div>
              <div className="platform-content">
                <div key={activeTab} className="tab-content-animate">
                  {activeTab === 'rates' && (
                    <div>
                      <div className="flex gap-2 mb-4 flex-wrap">
                        <span className="px-4 py-1.5 bg-brand-teal text-white rounded-md text-sm font-medium">🚛 Road</span>
                        <span className="px-4 py-1.5 bg-white border border-slate-200 text-slate-600 rounded-md text-sm transition-colors hover:bg-slate-50 cursor-pointer">✈️ Air</span>
                        <span className="px-4 py-1.5 bg-white border border-slate-200 text-slate-600 rounded-md text-sm transition-colors hover:bg-slate-50 cursor-pointer">🚢 Sea</span>
                      </div>
                      <table className="w-full text-left text-sm">
                        <thead><tr className="border-b border-slate-200 text-xs text-slate-500 uppercase">
                          <th className="py-3 px-3">Vendor</th><th className="py-3 px-3">Rate</th><th className="py-3 px-3">Transit</th><th className="py-3 px-3">Tag</th>
                        </tr></thead>
                        <tbody>
                          <tr className="border-b border-slate-100 stagger-row"><td className="py-3 px-3 font-medium">SafeExpress</td><td className="py-3 px-3 font-bold text-brand-teal">₹58,000</td><td className="py-3 px-3 text-slate-600">2 days</td><td className="py-3 px-3"><span className="tag tag-green">BEST</span></td></tr>
                          <tr className="border-b border-slate-100 stagger-row"><td className="py-3 px-3 font-medium">TCI Freight</td><td className="py-3 px-3 font-bold">₹60,500</td><td className="py-3 px-3 text-slate-600">2 days</td><td className="py-3 px-3"><span className="tag tag-blue">FAST</span></td></tr>
                          <tr className="stagger-row"><td className="py-3 px-3 font-medium">VRL Logistics</td><td className="py-3 px-3 font-bold">₹62,000</td><td className="py-3 px-3 text-slate-600">3 days</td><td className="py-3 px-3"><span className="tag tag-amber">RELIABLE</span></td></tr>
                        </tbody>
                      </table>
                    </div>
                  )}
                  {activeTab === 'tracking' && (
                    <div>
                      <div className="grid grid-cols-3 gap-4 mb-4">
                        <div className="bg-white border border-slate-200 rounded-lg p-3 flex justify-between items-center stagger-row"><span className="text-sm text-slate-600">On Road</span><span className="font-bold text-green-600">6 🚛</span></div>
                        <div className="bg-white border border-slate-200 rounded-lg p-3 flex justify-between items-center stagger-row"><span className="text-sm text-slate-600">In Air</span><span className="font-bold text-blue-600">3 ✈️</span></div>
                        <div className="bg-white border border-slate-200 rounded-lg p-3 flex justify-between items-center stagger-row"><span className="text-sm text-slate-600">At Sea</span><span className="font-bold text-amber-600">2 🚢</span></div>
                      </div>
                      <div className="bg-slate-100 rounded-lg h-48 flex items-center justify-center map-scan border border-slate-200 stagger-row" style={{ animationDelay: '0.4s' }}>
                        <div className="text-center relative z-10">
                          <div className="flex justify-center items-center gap-2 mb-2"><span className="live-pulse"></span><span className="text-4xl">🗺️</span></div>
                          <div className="font-bold text-brand-navy">Live Global Tracking</div>
                          <div className="text-xs text-slate-500 mt-1">GPS • AIS • FlightAware</div>
                        </div>
                      </div>
                    </div>
                  )}
                  {activeTab === 'rfq' && (
                    <div className="space-y-3">
                      <h3 className="font-bold text-brand-navy mb-3">📋 Active RFQs</h3>
                      <div className="bg-white border border-slate-200 p-4 rounded-lg flex justify-between items-center stagger-row hover:-translate-y-1 transition-transform cursor-pointer shadow-sm"><div><div className="font-bold text-sm text-brand-indigo mb-1">#RFQ-089 · JNPT → Rotterdam</div><div className="text-xs text-slate-500">🚢 FCL 40ft · Chemicals (DG) · 6 vendors</div></div><span className="bg-amber-50 text-amber-700 border border-amber-200 text-xs px-3 py-1 rounded-full font-medium flex items-center gap-2"><span className="loading-spinner w-3 h-3 border-2 border-amber-500 border-t-transparent rounded-full animate-spin"></span> 4/6 Quotes</span></div>
                      <div className="bg-white border border-slate-200 p-4 rounded-lg flex justify-between items-center stagger-row hover:-translate-y-1 transition-transform cursor-pointer shadow-sm"><div><div className="font-bold text-sm text-brand-indigo mb-1">#RFQ-088 · DEL → SIN</div><div className="text-xs text-slate-500">✈️ Air 800kg · Electronics · 4 vendors</div></div><span className="bg-green-50 text-green-700 border border-green-200 text-xs px-3 py-1 rounded-full font-medium">✓ Complete</span></div>
                    </div>
                  )}
                  {activeTab === 'compliance' && (
                    <div>
                      <h3 className="font-bold text-brand-navy mb-3">🔍 HSN Code Lookup</h3>
                      <input type="text" value="Methanol" readOnly className="w-full bg-white border border-brand-teal focus:ring-2 focus:ring-brand-teal/20 rounded-lg px-4 py-2.5 text-sm mb-4 outline-none transition-all" />
                      <div className="bg-white border border-red-200 p-4 rounded-lg flex gap-4 items-start stagger-row bg-red-50/30">
                        <div className="text-2xl animate-bounce">⚠️</div>
                        <div><div className="font-bold text-sm text-red-700">2905.11 — Methanol</div><div className="text-xs text-red-600/80 mt-1">Duty: 7.5% | GST: 18% | Class 3 Flammable</div></div>
                        <span className="tag tag-red ml-auto shadow-sm shadow-red-200">HAZ</span>
                      </div>
                    </div>
                  )}
                </div>
              </div>
            </div>
          </Reveal>
        </div>
      </section>

      {/* ═══ 50/50 STICKY SHOWCASE — Why Freel ═══ */}
      <section className="bg-white" style={{ padding: '120px 0' }}>
        <div className="container-lg">
          <Reveal>
            <SectionHeader label="Why Freel" labelColor="teal" title="Built for the Way Logistics Actually Works"
              subtitle="We didn't build another marketplace. We built the command center logistics teams actually need." />
          </Reveal>

          <StickyShowcase features={[
            {
              id: 'broker',
              title: 'Eliminate Broker Dependency',
              desc: 'Indian transporters lose ₹5-7 lakh/year to broker commissions. Freel connects you directly with verified shippers — no middlemen, no commission cuts.',
              stats: [{ v: '₹7L+', l: 'Saved/Year' }, { v: '0%', l: 'Commission' }],
              image: '/assets/broker_free.png'
            },
            {
              id: 'visibility',
              title: 'Real-Time Visibility',
              desc: 'GPS for trucks, AIS for vessels, FlightAware for air cargo. One dashboard, zero phone calls to ask "where\'s my shipment?"',
              stats: [{ v: '3', l: 'Tracking Modes' }, { v: 'Live', l: 'Updates' }],
              image: '/assets/smart_routing.png'
            },
            {
              id: 'docs',
              title: 'Digital Documentation',
              desc: 'No more paper LRs that vanish. Digital Lorry Receipts, auto-generated PODs with QR codes, instant invoices — all searchable.',
              stats: [{ v: '100%', l: 'Digital' }, { v: 'QR', l: 'Coded PODs' }],
              image: '/assets/digital_platform_dashboard.png'
            },
            {
              id: 'payments',
              title: 'Payments in 7 Days, Not 90',
              desc: 'The industry average payment cycle is 30-90 days. Freel\'s verified shipper network and digital contracts mean you get paid faster — always.',
              stats: [{ v: '7', l: 'Day Payments' }, { v: '85%', l: 'Faster' }],
              image: '/assets/fast_payments.png'
            }
          ]} />
        </div>
      </section>

      {/* ═══ TESTIMONIALS ═══ */}
      <section className="section-padding bg-white">
        <div className="container-lg">
          <Reveal>
            <SectionHeader label="What They Say" title="Trusted by Fleet Owners & Forwarders"
              subtitle="Real stories from logistics professionals using Freel." />
          </Reveal>
          <div className="testimonial-grid">
            {[
              { name: 'Rajesh Kumar', role: 'Fleet Owner · 45 trucks · Nagpur', text: 'Before Freel, I was spending 3 hours every morning calling brokers. Now I get load matches in minutes. My trucks run 30% more loaded miles.', avatar: 'RK', color: '#00BFA5' },
              { name: 'Priya Sharma', role: 'Freight Forwarder · Mumbai', text: 'Rate comparison across 500+ vendors in one click? That alone saved us ₹12 lakh last quarter. The RFQ system is a game-changer.', avatar: 'PS', color: '#5A4FCF' },
              { name: 'Amit Patel', role: 'Shipper · Chemical Manufacturing · Vadodara', text: 'The compliance module handles HSN codes and MSDS documentation automatically. We cleared customs 40% faster on our last 50 shipments.', avatar: 'AP', color: '#F59E0B' },
            ].map((t, i) => (
              <Reveal key={i} delay={`reveal-delay-${i + 1}`}>
                <TiltCard>
                  <div className="testimonial-card">
                    <div className="testimonial-stars">★★★★★</div>
                    <p className="testimonial-text">"{t.text}"</p>
                    <div className="testimonial-author">
                      <div className="testimonial-avatar" style={{ background: t.color }}>{t.avatar}</div>
                      <div><div className="testimonial-name">{t.name}</div><div className="testimonial-role">{t.role}</div></div>
                    </div>
                  </div>
                </TiltCard>
              </Reveal>
            ))}
          </div>
        </div>
      </section>

      {/* ═══ DARK CTA ═══ */}
      <section className="section-padding dark-cta">
        <div className="container-sm text-center relative z-10">
          <Reveal>
            <h2 className="text-3xl md:text-4xl font-black mb-4">Ready to modernize your logistics?</h2>
            <p className="max-w-lg mx-auto mb-8">Join 500+ fleet owners and freight forwarders who switched to Freel. Start free — no credit card required.</p>
            <div className="flex flex-col sm:flex-row items-center justify-center gap-3">
              <MagneticBtn><Link to="/signup" className="btn-primary">Start Free Trial →</Link></MagneticBtn>
              <MagneticBtn><Link to="/contact" className="btn-secondary" style={{ borderColor: '#94A3B8', color: '#94A3B8' }}>Talk to Sales</Link></MagneticBtn>
            </div>
          </Reveal>
        </div>
      </section>
    </>
  );
}

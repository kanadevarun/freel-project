import { useEffect, useRef, useState } from 'react';
import { Link } from 'react-router-dom';
import '../../../styles/Solutions.css';
import '../../../styles/ServiceDetail.css';

/* ─── Scroll Reveal ─── */
function useReveal() {
  const ref = useRef(null);
  useEffect(() => {
    const el = ref.current;
    if (!el) return;
    const obs = new IntersectionObserver(
      ([e]) => { if (e.isIntersecting) { el.classList.add('visible'); obs.unobserve(el); } },
      { threshold: 0.15 }
    );
    obs.observe(el);
    return () => obs.disconnect();
  }, []);
  return ref;
}
function Reveal({ children, className = '', delay = '' }) {
  const ref = useReveal();
  return <div ref={ref} className={`reveal ${delay} ${className}`}>{children}</div>;
}

/* ─── Page Data ─── */
const steps = [
  { n: '1', icon: '📤', title: 'Create & Send RFQ', desc: 'Enter shipment details (origin, destination, cargo, mode) and send to multiple vendors at once. Set deadline for responses.', bg: '#f0fdfa' },
  { n: '2', icon: '📥', title: 'Vendors Respond', desc: 'Vendors upload their rates from their dashboard. Real-time response counter shows 4/6. Email/WhatsApp reminders active.', bg: '#eff6ff' },
  { n: '3', icon: '⚖️', title: 'Compare & Select', desc: 'Side-by-side comparison cards with price, transit time, vendor rating. BEST badge on cheapest. Margin % auto-calculated.', bg: '#faf5ff' },
  { n: '4', icon: '✅', title: 'Convert to Order', desc: 'Select best vendor → auto-create quotation for client → on approval, convert to booking order with one click.', bg: '#f0fdf4' },
];

const features = [
  { icon: '📊', title: 'Price Buckets', desc: 'Auto-categorize quotes into Budget (₹0-50K), Standard (₹50-100K), Premium (₹100K+) for lightning-fast decision making.' },
  { icon: '💰', title: 'Margin Calculator', desc: 'See your projected profit margin on each vendor quote before selecting. Auto-calculates dynamically based on your client\'s budget.' },
  { icon: '⭐', title: 'Vendor Rating', desc: 'View historical vendor performance metrics: on-time %, damage rate, response speed. Pick reliable carriers, not just cheap ones.' },
  { icon: '📱', title: 'WhatsApp Notify', desc: 'Auto-send RFQs to vendors via WhatsApp. They get a secure link to upload rates directly from their phone — zero friction.' },
  { icon: '📄', title: 'PDF Quotation', desc: 'Generate one-click, beautifully branded PDF quotation with your logo, custom terms, and the selected vendor rate with built-in margin.' },
  { icon: '🔄', title: 'Re-Negotiate', desc: 'Counter-offer directly from the comparison view. Vendor gets notified instantly and can update their rate in real-time.' },
];

/* ═══════════════════════════════════════════ */
/*             RFQ LANDING PAGE               */
/* ═══════════════════════════════════════════ */
export default function RFQLanding() {
  const [bids, setBids] = useState([
    { carrier: 'TCI Freight', rate: '₹48,500', time: '5 mins ago', active: false, rank: 2 },
    { carrier: 'SafeExpress', rate: '₹46,200', time: 'Just now', active: true, rank: 1 },
    { carrier: 'VRL Logistics', rate: '₹51,000', time: '12 mins ago', active: false, rank: 3 }
  ]);

  useEffect(() => {
    const interval = setInterval(() => {
      setBids(prevBids => {
        const nextBids = [...prevBids];
        const updateIndex = Math.floor(Math.random() * nextBids.length);
        const currentRate = parseInt(nextBids[updateIndex].rate.replace('₹', '').replace(',', ''));
        const delta = Math.floor(Math.random() * 400) + 100;
        const newRateVal = Math.max(42000, currentRate - delta);
        
        nextBids[updateIndex] = {
          ...nextBids[updateIndex],
          rate: `₹${newRateVal.toLocaleString('en-IN')}`,
          time: 'Just now',
          active: true
        };

        nextBids.forEach((b, i) => {
          if (i !== updateIndex) {
            nextBids[i] = { ...b, active: false };
          }
        });

        const sorted = [...nextBids].sort((a, b) => {
          const rA = parseInt(a.rate.replace('₹', '').replace(',', ''));
          const rB = parseInt(b.rate.replace('₹', '').replace(',', ''));
          return rA - rB;
        });

        return nextBids.map(b => {
          const rank = sorted.findIndex(s => s.carrier === b.carrier) + 1;
          return { ...b, rank };
        });
      });
    }, 4000);

    return () => clearInterval(interval);
  }, []);

  return (
    <>
      {/* ═══ HERO ═══ */}
      <section className="rfq-split-hero" style={{ position: 'relative', overflow: 'hidden', padding: '120px 16px 80px' }} id="rfq-hero">
        <div style={{ maxWidth: '1200px', margin: '0 auto', display: 'grid', gridTemplateColumns: '1.2fr 1fr', gap: '48px', alignItems: 'center' }} className="rfq-hero-grid">
          {/* Left Column: Premium Content */}
          <Reveal>
            <div style={{ zIndex: 1 }}>
              <div style={{ display: 'inline-flex', alignItems: 'center', gap: '8px', padding: '6px 18px', borderRadius: '999px', background: 'rgba(0,191,165,0.08)', border: '1px solid rgba(0,191,165,0.2)', fontSize: '0.85rem', fontWeight: 700, color: 'var(--color-brand-teal)', marginBottom: '24px' }}>
                <span className="live-pulse"></span> AUTO-RFQ DISPATCH ENGINE
              </div>
              <h1 style={{ fontSize: 'clamp(2.5rem, 5.5vw, 4rem)', fontWeight: 900, lineHeight: 1.1, letterSpacing: '-0.03em', color: 'var(--color-brand-navy)', marginBottom: '20px' }}>
                Procure Freight <br />
                <span className="text-gradient bg-gradient-to-r from-brand-teal via-cyan-500 to-brand-indigo">
                  At 14.2% Lower Cost
                </span>
              </h1>
              <p style={{ fontSize: '1.15rem', color: 'var(--color-ui-gray)', lineHeight: 1.7, maxWidth: '580px', marginBottom: '36px', fontWeight: 300 }}>
                Stop wasting hours chasing transporters on email and spreadsheets. Automate your procurement, gather quotes in real-time, and compare bids side-by-side with direct margin protection.
              </p>
              
              <div style={{ display: 'flex', flexWrap: 'wrap', gap: '16px' }}>
                <Link to="/contact" className="solutions-cta-btn" style={{ fontSize: '1.05rem', padding: '16px 36px' }}>
                  Start Auto-RFQ →
                </Link>
                <a href="#how-it-works" className="solutions-cta-btn" style={{ background: 'white', color: 'var(--color-brand-navy)', border: '1px solid var(--color-ui-border)', boxShadow: 'none', fontSize: '1.05rem', padding: '16px 36px' }}>
                  Explore Steps ↓
                </a>
              </div>
              
              {/* Telemetry Micro-Stat Row */}
              <div className="rfq-hero-stats" style={{ display: 'flex', gap: '32px', marginTop: '48px', borderTop: '1px solid var(--color-ui-border)', paddingTop: '24px' }}>
                <div>
                  <div style={{ fontSize: '1.5rem', fontWeight: 800, color: 'var(--color-brand-navy)' }}>15 mins</div>
                  <div style={{ fontSize: '0.75rem', fontWeight: 600, color: 'var(--color-ui-gray)', textTransform: 'uppercase', marginTop: '2px' }}>Avg. Response Time</div>
                </div>
                <div>
                  <div style={{ fontSize: '1.5rem', fontWeight: 800, color: 'var(--color-brand-navy)' }}>10,000+</div>
                  <div style={{ fontSize: '0.75rem', fontWeight: 600, color: 'var(--color-ui-gray)', textTransform: 'uppercase', marginTop: '2px' }}>Carrier Network</div>
                </div>
                <div>
                  <div style={{ fontSize: '1.5rem', fontWeight: 800, color: 'var(--color-brand-navy)' }}>₹1.2 Cr</div>
                  <div style={{ fontSize: '0.75rem', fontWeight: 600, color: 'var(--color-ui-gray)', textTransform: 'uppercase', marginTop: '2px' }}>Monthly Saved</div>
                </div>
              </div>
            </div>
          </Reveal>

          {/* Right Column: Animated perspective Bidding showcase */}
          <Reveal delay="reveal-delay-2">
            <div style={{ position: 'relative', display: 'flex', justifyContent: 'center', alignItems: 'center' }}>
              {/* Background gradient shadow */}
              <div style={{
                position: 'absolute',
                width: '100%',
                height: '100%',
                background: 'radial-gradient(circle at center, rgba(90,79,207,0.12) 0%, transparent 70%)',
                filter: 'blur(40px)',
                zIndex: -1
              }}></div>

              {/* 3D Styled Rotating Dashboard mock */}
              <div className="rfq-perspective-dashboard" style={{
                width: '100%',
                maxWidth: '460px',
                background: 'rgba(255, 255, 255, 0.75)',
                backdropFilter: 'blur(20px)',
                WebkitBackdropFilter: 'blur(20px)',
                border: '1px solid rgba(226, 232, 240, 0.8)',
                borderRadius: '24px',
                padding: '24px',
                boxShadow: '0 30px 60px -15px rgba(15,23,42,0.12), 0 0 0 1px rgba(0,0,0,0.02)',
                transition: 'all 0.5s cubic-bezier(0.16,1,0.3,1)',
                position: 'relative'
              }}>
                {/* Dashboard Bar */}
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '20px', borderBottom: '1px solid rgba(226, 232, 240, 0.6)', paddingBottom: '12px' }}>
                  <div style={{ display: 'flex', gap: '6px' }}>
                    <span style={{ width: '10px', height: '10px', borderRadius: '50%', background: '#EF4444' }}></span>
                    <span style={{ width: '10px', height: '10px', borderRadius: '50%', background: '#F59E0B' }}></span>
                    <span style={{ width: '10px', height: '10px', borderRadius: '50%', background: '#10B981' }}></span>
                  </div>
                  <div style={{ fontSize: '0.75rem', fontWeight: 700, color: 'var(--color-brand-indigo)', background: 'rgba(90,79,207,0.08)', padding: '4px 10px', borderRadius: '99px' }}>
                    LIVE AUCTION
                  </div>
                </div>

                {/* Cargo Details */}
                <div style={{ background: 'var(--color-brand-navy)', borderRadius: '16px', padding: '16px', color: 'white', marginBottom: '20px', position: 'relative', overflow: 'hidden' }}>
                  <div style={{ position: 'absolute', top: '-20px', right: '-20px', width: '80px', height: '80px', background: 'rgba(255,255,255,0.05)', borderRadius: '50%' }}></div>
                  <div style={{ fontSize: '0.7rem', textTransform: 'uppercase', color: '#94A3B8', fontWeight: 700, letterSpacing: '0.05em' }}>Shipment #RFQ-9921</div>
                  <div style={{ fontSize: '1.1rem', fontWeight: 800, marginTop: '4px', display: 'flex', alignItems: 'center', gap: '8px' }}>
                    JNPT Port <span style={{ color: 'var(--color-brand-teal)' }}>→</span> Rotterdam Port
                  </div>
                  <div style={{ display: 'flex', gap: '16px', marginTop: '12px', fontSize: '0.75rem', color: '#CBD5E1' }}>
                    <div>🚢 FCL 40ft</div>
                    <div>⚖️ 22,000 kg</div>
                    <div>🛡️ Class 3 (HAZ)</div>
                  </div>
                </div>

                {/* Bidding Telemetry Header */}
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '12px' }}>
                  <div style={{ fontSize: '0.8rem', fontWeight: 700, color: 'var(--color-brand-navy)' }}>Transporter Responses</div>
                  <div style={{ fontSize: '0.75rem', color: 'var(--color-ui-gray)' }}>
                    ⏳ Ends in <span style={{ fontWeight: 700, color: '#EF4444', fontFamily: 'monospace' }}>12m 45s</span>
                  </div>
                </div>

                {/* Bids List Container */}
                <div style={{ display: 'flex', flexDirection: 'column', gap: '10px' }} className="rfq-bids-list">
                  {bids.map((b, idx) => (
                    <div key={idx} className={`rfq-bid-card ${b.active ? 'active' : ''}`} style={{
                      display: 'flex',
                      justifyContent: 'space-between',
                      alignItems: 'center',
                      background: b.active ? 'rgba(0,191,165,0.06)' : 'white',
                      border: b.active ? '1px solid rgba(0,191,165,0.3)' : '1px solid rgba(226, 232, 240, 0.8)',
                      borderRadius: '12px',
                      padding: '12px 16px',
                      boxShadow: '0 4px 6px -1px rgba(0,0,0,0.02)',
                      transition: 'all 0.4s cubic-bezier(0.16,1,0.3,1)'
                    }}>
                      <div>
                        <div style={{ fontSize: '0.85rem', fontWeight: 700, color: 'var(--color-brand-navy)', display: 'flex', alignItems: 'center', gap: '6px' }}>
                          {b.carrier}
                          {b.active && <span style={{ width: '6px', height: '6px', borderRadius: '50%', background: 'var(--color-brand-teal)', display: 'inline-block' }}></span>}
                        </div>
                        <div style={{ fontSize: '0.7rem', color: 'var(--color-ui-gray)', marginTop: '2px' }}>
                          Rank #{b.rank} • {b.time}
                        </div>
                      </div>
                      <div style={{ textAlign: 'right' }}>
                        <div style={{ fontSize: '0.95rem', fontWeight: 800, color: b.rank === 1 ? 'var(--color-brand-teal)' : 'var(--color-brand-navy)' }}>{b.rate}</div>
                        {b.rank === 1 && <span style={{ fontSize: '0.65rem', fontWeight: 700, color: 'var(--color-brand-teal)', background: 'rgba(0,191,165,0.1)', padding: '2px 6px', borderRadius: '4px' }}>BEST RATE</span>}
                      </div>
                    </div>
                  ))}
                </div>

                {/* Instant Margin Calculator Floating Badge */}
                <div className="rfq-margin-badge" style={{
                  position: 'absolute',
                  bottom: '-20px',
                  right: '-20px',
                  background: 'white',
                  border: '1px solid rgba(226, 232, 240, 0.8)',
                  borderRadius: '16px',
                  padding: '12px 18px',
                  display: 'flex',
                  alignItems: 'center',
                  gap: '10px',
                  boxShadow: '0 20px 25px -5px rgba(0,0,0,0.1), 0 10px 10px -5px rgba(0,0,0,0.04)',
                  zIndex: 2
                }}>
                  <div style={{ width: '36px', height: '36px', borderRadius: '10px', background: 'rgba(0,191,165,0.1)', display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: '1.2rem', color: 'var(--color-brand-teal)' }}>
                    📈
                  </div>
                  <div>
                    <div style={{ fontSize: '0.65rem', color: 'var(--color-ui-gray)', fontWeight: 600, textTransform: 'uppercase' }}>Protected Margin</div>
                    <div style={{ fontSize: '0.9rem', fontWeight: 800, color: 'var(--color-brand-navy)' }}>+18.4% (₹8,920)</div>
                  </div>
                </div>
              </div>
            </div>
          </Reveal>
        </div>
      </section>

      {/* ═══ HOW RFQ WORKS ═══ */}
      <section className="solution-module" id="how-it-works">
        <Reveal>
          <div className="solution-module-header">
            <h2 className="solution-module-title">How RFQ Works</h2>
            <p className="solution-module-desc">
              Client sends requirements → You create RFQ → Vendors quote → You compare → Client gets best rate
            </p>
          </div>
        </Reveal>

        <div style={{ maxWidth: '1100px', margin: '0 auto', display: 'flex', flexWrap: 'wrap', borderRadius: '24px', overflow: 'hidden', border: '1px solid var(--color-ui-border)', background: 'var(--color-ui-surface)' }}>
          {steps.map((s, i) => (
            <Reveal key={i} delay={`reveal-delay-${i + 1}`} className="flex-1">
              <div className="rfq-step-card" style={{
                background: s.bg,
                borderRight: i < 3 ? '1px solid rgba(226,232,240,0.5)' : 'none',
              }}>
                <div style={{
                  width: '48px', height: '48px', borderRadius: '50%',
                  background: `linear-gradient(135deg, var(--color-brand-teal), var(--color-brand-indigo))`,
                  color: 'white', display: 'flex', alignItems: 'center', justifyContent: 'center',
                  fontWeight: 700, fontSize: '1.2rem', marginBottom: '20px',
                  boxShadow: '0 4px 12px rgba(0,191,165,0.25)'
                }}>{s.n}</div>
                <h3 style={{ fontSize: '1.15rem', fontWeight: 700, color: 'var(--color-brand-navy)', marginBottom: '10px' }}>
                  {s.icon} {s.title}
                </h3>
                <p style={{ fontSize: '0.85rem', color: 'var(--color-ui-gray)', lineHeight: 1.6 }}>{s.desc}</p>
              </div>
            </Reveal>
          ))}
        </div>
      </section>

      {/* ═══ TWO DASHBOARDS ═══ */}
      <section className="solution-module solution-module-alt">
        <Reveal>
          <div className="solution-module-header">
            <h2 className="solution-module-title">Two Dashboards, One Flow</h2>
            <p className="solution-module-desc">
              RFQ works seamlessly between your client-facing and transporter-facing dashboards, keeping everyone synced.
            </p>
          </div>
        </Reveal>

        <div className="dashboard-split">
          <Reveal delay="reveal-delay-1">
            <div className="dashboard-card" style={{ display: 'flex', flexDirection: 'column', height: '100%' }}>
              <div className="dashboard-card-header">
                <div className="dashboard-card-icon">🏢</div>
                <div>
                  <div className="dashboard-card-title">Client Dashboard</div>
                  <div className="dashboard-card-sub">For the shipper / end-customer</div>
                </div>
              </div>
              <ul className="check-list" style={{ marginBottom: '24px', flexGrow: 1 }}>
                <li>Requests quote directly from you (Freel user platform).</li>
                <li>Sees branded, professional quotation PDF you send them.</li>
                <li>Approves or rejects quotation with one click.</li>
                <li>Gets instant WhatsApp & Email notifications on updates.</li>
                <li>Tracks shipment live after booking is confirmed.</li>
              </ul>
              <div style={{ marginTop: 'auto', overflow: 'hidden', borderRadius: '12px', border: '1px solid var(--color-ui-border)', boxShadow: '0 4px 12px rgba(0,0,0,0.03)' }}>
                <img src="/assets/digital_platform_dashboard.png" alt="Client Dashboard" style={{ width: '100%', height: '220px', objectFit: 'cover', display: 'block' }} className="hover-scale-img" />
              </div>
            </div>
          </Reveal>

          <Reveal delay="reveal-delay-2">
            <div className="dashboard-card" style={{ display: 'flex', flexDirection: 'column', height: '100%' }}>
              <div className="dashboard-card-header">
                <div className="dashboard-card-icon">🚚</div>
                <div>
                  <div className="dashboard-card-title">Transporter Dashboard</div>
                  <div className="dashboard-card-sub">For vendors & carriers</div>
                </div>
              </div>
              <ul className="check-list" style={{ marginBottom: '24px', flexGrow: 1 }}>
                <li>Receives RFQ with full cargo & routing details.</li>
                <li>Uploads quote securely with granular rate breakdown.</li>
                <li>Sees confirmed order when selected (with full details).</li>
                <li>Views Container type, weight, volume, HAZ info clearly.</li>
                <li>Accesses POD/POL, vessel dates, and MSDS docs in one place.</li>
              </ul>
              <div style={{ marginTop: 'auto', overflow: 'hidden', borderRadius: '12px', border: '1px solid var(--color-ui-border)', boxShadow: '0 4px 12px rgba(0,0,0,0.03)' }}>
                <img src="/assets/smart_routing.png" alt="Transporter Dashboard" style={{ width: '100%', height: '220px', objectFit: 'cover', display: 'block' }} className="hover-scale-img" />
              </div>
            </div>
          </Reveal>
        </div>
      </section>

      {/* ═══ RFQ FEATURES ═══ */}
      <section className="solution-module">
        <Reveal>
          <div className="solution-module-header">
            <h2 className="solution-module-title">Powerful RFQ Features</h2>
            <p className="solution-module-desc">
              Built to save you time, increase margins, and provide a premium experience.
            </p>
          </div>
        </Reveal>
        <div className="feature-grid-3">
          {features.map((f, i) => (
            <Reveal key={i} delay={`reveal-delay-${(i % 3) + 1}`}>
              <div className="feature-card">
                <div className="feature-card-icon">{f.icon}</div>
                <h3 className="feature-card-title">{f.title}</h3>
                <p className="feature-card-desc">{f.desc}</p>
              </div>
            </Reveal>
          ))}
        </div>
      </section>

      {/* ═══ CTA ═══ */}
      <section className="solution-module solution-module-alt" style={{ paddingBottom: '80px' }}>
        <Reveal>
          <div className="solutions-cta-card">
            <h2>Start Collecting Vendor Quotes</h2>
            <p>Send your first RFQ to multiple vendors in under 2 minutes and streamline your procurement process today.</p>
            <Link to="/contact" className="solutions-cta-btn">
              Try RFQ System →
            </Link>
            <div style={{ marginTop: '24px', fontSize: '0.85rem', color: 'var(--color-ui-gray)' }}>
              No credit card required. 14-day free trial.
            </div>
          </div>
        </Reveal>
      </section>
    </>
  );
}

import { useEffect, useRef, useState } from 'react';
import { Link } from 'react-router-dom';
import './RFQLanding.css';

/* ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
   SCROLL REVEAL
   ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━ */
function useReveal(threshold = 0.1) {
  const ref = useRef(null);
  useEffect(() => {
    const el = ref.current;
    if (!el) return;
    const obs = new IntersectionObserver(
      ([e]) => { if (e.isIntersecting) { el.classList.add('r3-visible'); obs.unobserve(el); } },
      { threshold }
    );
    obs.observe(el);
    return () => obs.disconnect();
  }, [threshold]);
  return ref;
}

function R({ children, d = '', className = '' }) {
  const ref = useReveal();
  return <div ref={ref} className={`r3-reveal ${d} ${className}`}>{children}</div>;
}

/* ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
   ANIMATED COUNTER
   ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━ */
function Counter({ to, decimals = 0, prefix = '', suffix = '' }) {
  const [v, setV] = useState(0);
  const ref = useRef(null);
  useEffect(() => {
    let go = false;
    const obs = new IntersectionObserver(([e]) => {
      if (e.isIntersecting && !go) {
        go = true;
        const t0 = performance.now();
        const dur = 2000;
        const tick = (now) => {
          const p = Math.min((now - t0) / dur, 1);
          const ease = 1 - Math.pow(1 - p, 4);
          setV(ease * to);
          if (p < 1) requestAnimationFrame(tick);
        };
        requestAnimationFrame(tick);
      }
    }, { threshold: 0.4 });
    if (ref.current) obs.observe(ref.current);
    return () => obs.disconnect();
  }, [to]);
  return <span ref={ref}>{prefix}{v.toFixed(decimals)}{suffix}</span>;
}

/* ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
   LIVE AUCTION PANEL
   ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━ */
const SEED_BIDS = [
  { id: 1, name: 'Vendor A', rate: 48500 },
  { id: 2, name: 'Vendor B', rate: 47900 },
  { id: 3, name: 'Vendor C', rate: 46200 },
];

function LivePanel() {
  const [bids, setBids] = useState(SEED_BIDS.map(b => ({ ...b, flash: false, prevRate: b.rate })));
  const [timer, setTimer] = useState({ m: 7, s: 22 });

  useEffect(() => {
    const cd = setInterval(() => setTimer(p => p.s === 0 ? { m: Math.max(0, p.m - 1), s: 59 } : { ...p, s: p.s - 1 }), 1000);
    return () => clearInterval(cd);
  }, []);

  useEffect(() => {
    const t = setInterval(() => {
      setBids(prev => {
        const i = Math.floor(Math.random() * prev.length);
        const cur = prev[i].rate;
        const delta = Math.floor(Math.random() * 500) + 100;
        const next = Math.max(43000, cur - delta);
        const updated = prev.map((b, idx) => idx === i
          ? { ...b, prevRate: cur, rate: next, flash: true }
          : { ...b, flash: false }
        );
        const sorted = [...updated].sort((a, b) => a.rate - b.rate);
        return sorted;
      });
    }, 2800);
    return () => clearInterval(t);
  }, []);

  const best = bids[0]?.rate || 46200;
  const worst = bids[bids.length - 1]?.rate || 48500;
  const savings = (((worst - best) / worst) * 100).toFixed(1);

  return (
    <div className="r3-hero-panel">
      {/* macOS chrome */}
      <div className="r3-panel-chrome">
        <div className="r3-chrome-dots">
          <span style={{ background: '#ef4444' }} />
          <span style={{ background: '#f59e0b' }} />
          <span style={{ background: '#22c55e' }} />
        </div>
        <div className="r3-chrome-url">logisticshq.in / rfq / #9921 / bids</div>
        <div className="r3-panel-live-tag">
          <span className="r3-panel-live-dot" />
          Live
        </div>
      </div>

      <div className="r3-panel-body">
        {/* Route */}
        <div className="r3-panel-route">
          <div>
            <div className="r3-panel-route-from">Mumbai → Chennai</div>
            <div className="r3-panel-route-meta">
              <span className="r3-panel-route-tag">FTL</span>
              <span className="r3-panel-route-tag">22 Tons</span>
              <span className="r3-panel-route-tag">Dry Cargo</span>
            </div>
          </div>
          <div style={{ textAlign: 'right' }}>
            <div style={{ fontSize: '0.65rem', fontWeight: 700, color: 'var(--r3-muted)', textTransform: 'uppercase', letterSpacing: '0.08em', marginBottom: '4px' }}>Auction</div>
            <div style={{ fontSize: '0.85rem', fontWeight: 800, color: 'var(--r3-blue-d)' }}>
              {timer.m}m {String(timer.s).padStart(2, '0')}s
            </div>
          </div>
        </div>

        {/* Bids */}
        <div className="r3-panel-bids-label">
          <span>Live Quotes · {bids.length} vendors</span>
          <span className="r3-panel-timer">Closes in {timer.m}m {String(timer.s).padStart(2, '0')}s</span>
        </div>
        <div className="r3-bid-list">
          {bids.map((b, i) => (
            <div
              key={b.id}
              className={`r3-bid-row ${i === 0 ? 'r3-bid-winner' : ''} ${b.flash ? 'r3-bid-flash' : ''}`}
            >
              <div>
                <div className="r3-bid-vendor">{b.name}</div>
                <div className="r3-bid-time">Rank #{i + 1}</div>
              </div>
              <div className="r3-bid-right">
                <div className="r3-bid-price">₹{b.rate.toLocaleString('en-IN')}</div>
                {i === 0 && <span className="r3-bid-badge r3-bid-best">Best Rate</span>}
                {b.flash && i !== 0 && <span className="r3-bid-badge r3-bid-updated">Updated</span>}
              </div>
            </div>
          ))}
        </div>

        {/* Savings */}
        <div className="r3-panel-savings">
          <div>
            <div className="r3-panel-savings-label">Market savings on this shipment</div>
            <div style={{ fontSize: '0.65rem', color: 'var(--r3-muted)', marginTop: '2px' }}>vs. first quote received</div>
          </div>
          <div className="r3-panel-savings-val">↓ {savings}%</div>
        </div>
      </div>

      {/* Floating badges */}
      <div className="r3-float-badge" style={{ bottom: -18, left: -24 }}>
        <div className="r3-float-badge-icon" style={{ background: 'rgba(13,148,136,0.1)' }}>📉</div>
        <div>
          <div className="r3-float-badge-val">14.2% Saved</div>
          <div className="r3-float-badge-sub">Avg. across all RFQs</div>
        </div>
      </div>
      <div className="r3-float-badge" style={{ top: 60, right: -28 }}>
        <div className="r3-float-badge-icon" style={{ background: 'rgba(14,165,233,0.1)' }}>⚡</div>
        <div>
          <div className="r3-float-badge-val">15 Min</div>
          <div className="r3-float-badge-sub">Avg. response time</div>
        </div>
      </div>
    </div>
  );
}

/* ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
   VENDOR RACE (WOW SECTION)
   ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━ */
const RACE_INIT = [
  { name: 'Gati KWE',      rate: 52000 },
  { name: 'VRL Logistics', rate: 49500 },
  { name: 'TCI Freight',   rate: 48000 },
  { name: 'SafeExpress',   rate: 46200 },
];

function VendorRace() {
  const [rates, setRates] = useState(RACE_INIT.map(r => r.rate));
  const maxR = 52000;

  useEffect(() => {
    const t = setInterval(() => {
      setRates(prev => prev.map((r, i) => {
        if (Math.random() > 0.45) return r;
        return Math.max(43000, r - Math.floor(Math.random() * 450 + 80));
      }));
    }, 2600);
    return () => clearInterval(t);
  }, []);

  const pairs = RACE_INIT.map((v, i) => ({ ...v, rate: rates[i] }))
    .sort((a, b) => a.rate - b.rate);

  return (
    <div className="r3-race">
      {pairs.map((v, i) => {
        const pct = Math.round((1 - (v.rate - 43000) / (maxR - 43000)) * 80 + 12);
        return (
          <div key={v.name} className={`r3-race-row ${i === 0 ? 'r3-race-winner' : ''}`}>
            <div className="r3-race-bar" style={{ width: `${pct}%` }} />
            <span className="r3-race-name">{v.name}</span>
            <span className="r3-race-price">₹{v.rate.toLocaleString('en-IN')}</span>
            {i === 0 && <span className="r3-race-win-badge">Best Rate</span>}
          </div>
        );
      })}
    </div>
  );
}

/* ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
   HOW IT WORKS STEPS (WITH IMAGES)
   ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━ */
const STEPS = [
  { img: 'https://images.unsplash.com/photo-1499951360447-b19be8fe80f5?auto=format&fit=crop&w=600&q=80', n: 'Step 01', title: 'Create RFQ', desc: 'Enter shipment details in seconds.' },
  { img: 'https://images.unsplash.com/photo-1454165804606-c3d57bc86b40?auto=format&fit=crop&w=600&q=80', n: 'Step 02', title: 'Invite Vendors', desc: 'Thousands notified instantly.' },
  { img: 'https://images.unsplash.com/photo-1551288049-bebda4e38f71?auto=format&fit=crop&w=600&q=80', n: 'Step 03', title: 'Compare Quotes', desc: 'Price, time, and ratings — live.' },
  { img: 'https://images.unsplash.com/photo-1586528116311-ad8dd3c8310d?auto=format&fit=crop&w=600&q=80', n: 'Step 04', title: 'Award Shipment', desc: 'Done in minutes. Not days.' },
];

/* ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
   AUTOMATION STEPS
   ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━ */
const AUTO_STEPS = [
  { icon: '📋', title: 'RFQ Created', sub: 'You fill in shipment details. That\'s it.' },
  { icon: '📧', title: 'Emails Sent', sub: 'Platform notifies all selected vendors.' },
  { icon: '💬', title: 'WhatsApp Sent', sub: 'Vendors receive mobile-friendly bid links.' },
  { icon: '🔔', title: 'Reminders Sent', sub: 'Auto follow-ups to non-responding vendors.' },
  { icon: '📥', title: 'Quotes Collected', sub: 'All responses organized in one dashboard.' },
];



/* ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
   MAIN PAGE
   ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━ */
export default function RFQLanding() {
  const [activeStep, setActiveStep] = useState(0);

  useEffect(() => {
    const t = setInterval(() => setActiveStep(s => (s + 1) % STEPS.length), 2800);
    return () => clearInterval(t);
  }, []);

  return (
    <div className="r3">

      {/* ═══════════════════════════════════════
          § 1 — HERO
          ═══════════════════════════════════════ */}
      <section className="r3-hero" id="rfq-hero">
        <div className="r3-wrap">
          <div className="r3-hero-grid">

            {/* Left */}
            <R>
              <div className="r3-eyebrow">
                <span className="r3-eyebrow-dot" />
                Freight Procurement
              </div>
              <h1 className="r3-hero-h1">
                Stop Chasing<br />
                <span className="r3-grad">Transporters.</span>
              </h1>
              <p className="r3-hero-sub">
                <span><strong>Create one RFQ.</strong> Reach thousands of verified carriers.</span>
                <span><strong>Compare rates instantly.</strong> Award freight in minutes.</span>
              </p>
              <div className="r3-ctas">
                <Link to="/contact" className="r3-btn r3-btn-primary">
                  Start RFQ →
                </Link>
                <a href="#rfq-wow" className="r3-btn r3-btn-ghost">
                  Watch Live Procurement ↓
                </a>
              </div>
            </R>

            {/* Right — Live Auction Panel */}
            <R d="r3-d2">
              <div style={{ position: 'relative', paddingTop: '20px', paddingBottom: '24px', paddingRight: '36px' }}>
                <LivePanel />
              </div>
            </R>
          </div>
        </div>
      </section>

      {/* ═══════════════════════════════════════
          § 2 — OLD WAY VS NEW WAY
          ═══════════════════════════════════════ */}
      <section className="r3-compare" id="rfq-compare">
        <div className="r3-wrap">
          <R>
            <span className="r3-section-kicker">Why LogisticsHQ</span>
            <h2 className="r3-compare-headline">
              Why Manage Vendors Manually<br />When Vendors Can Compete For You?
            </h2>
            <p className="r3-compare-sub">Let the market find your best rate.</p>
          </R>

          <div className="r3-compare-grid">
            {/* Old Way (Stylized) */}
            <R d="r3-d1" className="r3-compare-side r3-compare-old r3-styled-old">
              <div className="r3-styled-old-header">
                <span className="r3-compare-tag old" style={{ marginBottom: 0 }}>The Old Way</span>
                <div className="r3-styled-old-time">
                  <span>Average Time</span>
                  <strong>72+ Hours</strong>
                </div>
              </div>
              
              <div className="r3-styled-old-body">
                <p className="r3-styled-old-desc">Manual procurement creates endless friction, data silos, and missed savings.</p>
                
                <div className="r3-styled-old-list">
                  {[
                    { icon: '📧', val: '43', label: 'Emails sent per RFQ', color: 'r3-c-amber' },
                    { icon: '📞', val: '27', label: 'Phone calls made', color: 'r3-c-red' },
                    { icon: '📊', val: '12', label: 'Spreadsheets managed', color: 'r3-c-gray' },
                    { icon: '🤦', val: '∞', label: 'Manual follow-ups', color: 'r3-c-red' },
                  ].map((item, i) => (
                    <div key={i} className="r3-styled-old-item">
                      <div className={`r3-styled-old-icon ${item.color}`}>{item.icon}</div>
                      <div className="r3-styled-old-text">
                        <strong>{item.val}</strong> {item.label}
                      </div>
                    </div>
                  ))}
                </div>
              </div>
              
              <div className="r3-styled-old-footer">
                <div className="r3-styled-old-error">
                  <span className="r3-error-dot"></span> High risk of manual error
                </div>
              </div>
            </R>

            {/* VS divider */}
            <div className="r3-compare-vs">VS</div>

            {/* New Way */}
            <R d="r3-d2" className="r3-compare-side r3-compare-new">
              <span className="r3-compare-tag new">LogisticsHQ RFQ</span>
              <div className="r3-new-flow">
                {[
                  { icon: '📋', text: 'One RFQ', sub: 'Enter shipment details once' },
                  { icon: '⚡', text: '500+ Quotes', sub: 'Vendors compete automatically' },
                  { icon: '⚖️', text: 'Compare', sub: 'Live side-by-side dashboard' },
                  { icon: '✅', text: 'Award', sub: 'Click. Done. Ship.' },
                ].map((item, i) => (
                  <div key={i} className="r3-new-flow-item">
                    <div className="r3-new-flow-icon">{item.icon}</div>
                    <div>
                      <div className="r3-new-flow-text">{item.text}</div>
                      <div className="r3-new-flow-sub">{item.sub}</div>
                    </div>
                    <div className="r3-new-flow-check">✓</div>
                  </div>
                ))}
              </div>

              <div style={{
                marginTop: '28px', padding: '16px 20px',
                background: 'rgba(20,184,166,0.12)', border: '1px solid rgba(20,184,166,0.25)',
                borderRadius: 'var(--r3-radius)', display: 'flex', alignItems: 'center', justifyContent: 'space-between',
              }}>
                <div>
                  <div style={{ fontSize: '0.68rem', fontWeight: 700, color: 'rgba(20,184,166,0.8)', textTransform: 'uppercase', letterSpacing: '0.08em', marginBottom: '4px' }}>Time saved per shipment</div>
                  <div style={{ fontSize: '1.4rem', fontWeight: 900, color: '#5eead4', letterSpacing: '-0.04em' }}>3 days → 15 min</div>
                </div>
                <div style={{ fontSize: '2rem' }}>🏆</div>
              </div>
            </R>
          </div>
        </div>
      </section>

      {/* ═══════════════════════════════════════
          § 3 — HOW IT WORKS (WITH IMAGES)
          ═══════════════════════════════════════ */}
      <section className="r3-steps" id="rfq-steps">
        <div className="r3-wrap">
          <R className="r3-steps-header">
            <span className="r3-section-kicker">How It Works</span>
            <h2 className="r3-section-title">One RFQ. Four Simple Steps.</h2>
            <p className="r3-section-sub">From shipment details to carrier awarded — in under 15 minutes.</p>
          </R>
          <div className="r3-steps-track">
            {STEPS.map((s, i) => (
              <R key={i} d={`r3-d${i + 1}`}>
                <div
                  className={`r3-step ${i === activeStep ? 'r3-step-active' : ''}`}
                  onClick={() => setActiveStep(i)}
                >
                  <div className="r3-step-image-card">
                    <img src={s.img} alt={s.title} />
                  </div>
                  <div className="r3-step-number">{s.n}</div>
                  <div className="r3-step-title">{s.title}</div>
                  <div className="r3-step-desc">{s.desc}</div>
                </div>
              </R>
            ))}
          </div>
        </div>
      </section>

      {/* ═══════════════════════════════════════
          § 4 — LIVE SAVINGS DEMO (WOW)
          ═══════════════════════════════════════ */}
      <section className="r3-wow" id="rfq-wow">
        <div className="r3-wow-glow" />
        <div className="r3-wrap" style={{ position: 'relative', zIndex: 1 }}>
          <R>
            <span className="r3-section-kicker" style={{ color: 'rgba(14,165,233,0.8)', justifyContent: 'center', display: 'flex' }}>Live Savings</span>
            <h2 className="r3-wow-headline">Watch Costs Drop<br />In Real Time.</h2>
            <p className="r3-wow-sub">When you post an RFQ, the market responds. Prices drop as carriers compete for your shipment.</p>
          </R>

          {/* Cinematic Freight Banner */}
          <R d="r3-d1">
            <div className="r3-wow-banner">
              <img src="https://images.unsplash.com/photo-1601584115197-04ecc0da31d7?auto=format&fit=crop&w=2400&q=80" alt="Freight Trucks Highway" />
              <div className="r3-wow-banner-overlay" />
            </div>
          </R>

          {/* Price cascade */}
          <R d="r3-d2">
            <div className="r3-cascade">
              {[
                { price: '₹52,000', label: 'Opening', old: true },
                { arrow: '→' },
                { price: '₹49,500', label: 'Counter', old: true },
                { arrow: '→' },
                { price: '₹48,000', label: 'Third Offer', old: true },
                { arrow: '→' },
                { price: '₹46,200', label: 'Best Rate ✓', winner: true },
              ].map((item, i) => item.arrow ? (
                <span key={i} className="r3-cascade-arrow">{item.arrow}</span>
              ) : (
                <div key={i} className="r3-cascade-item">
                  <div className={`r3-cascade-price ${item.old ? 'r3-old' : ''} ${item.winner ? 'r3-winner' : ''}`}>
                    {item.price}
                  </div>
                  <div className={`r3-cascade-label ${item.winner ? 'r3-winner-label' : ''}`}>{item.label}</div>
                </div>
              ))}
            </div>
          </R>

          {/* Live race */}
          <R d="r3-d3">
            <VendorRace />
          </R>

          {/* Savings callout */}
          <R d="r3-d4">
            <div className="r3-savings-callout">
              <div>
                <div className="r3-savings-meta-title">Average market savings achieved</div>
                <div className="r3-savings-meta-sub">Across all LogisticsHQ RFQ shipments in FY2024</div>
              </div>
              <div className="r3-savings-main">
                <Counter to={14.2} decimals={1} suffix="%" />
              </div>
            </div>
          </R>
        </div>
      </section>

      {/* ═══════════════════════════════════════
          § 5 — PRODUCT EXPERIENCE
          ═══════════════════════════════════════ */}
      <section className="r3-product" id="rfq-product">
        <div className="r3-wrap">
          <R className="r3-product-header">
            <span className="r3-section-kicker">The Platform</span>
            <h2 className="r3-section-title">Everything In One Place.</h2>
            <div className="r3-product-pills">
              {['RFQs', 'Live Quotes', 'Margin Calculator', 'Approvals', 'Tracking', 'Analytics'].map((p, i) => (
                <span key={i} className="r3-product-pill">{p}</span>
              ))}
            </div>
          </R>

          <R d="r3-d1">
            <div className="r3-dashboard-mock-wrapper">
              {/* Floating Real-world Images */}
              <div className="r3-dash-float r3-float-1">
                <img src="https://images.unsplash.com/photo-1519003722824-194d4455a60c?auto=format&fit=crop&w=400&q=80" alt="Live Shipment" />
                <div className="r3-float-tag">Live Tracking</div>
              </div>
              <div className="r3-dash-float r3-float-2">
                <img src="https://images.unsplash.com/photo-1578575437130-527eed3abbec?auto=format&fit=crop&w=400&q=80" alt="Carrier Confirmed" />
                <div className="r3-float-tag">Carrier Ready</div>
              </div>
              <div className="r3-dash-float r3-float-3">
                <img src="https://images.unsplash.com/photo-1494412574643-ff11b0a5c1c3?auto=format&fit=crop&w=400&q=80" alt="Delivery Completed" />
                <div className="r3-float-tag">Delivered</div>
              </div>

              <div className="r3-dashboard-mock">
                {/* Top bar */}
                <div className="r3-dash-topbar">
                  <div className="r3-dash-topbar-dots">
                    <span style={{ background: '#ef4444' }} />
                    <span style={{ background: '#f59e0b' }} />
                    <span style={{ background: '#22c55e' }} />
                  </div>
                  <div className="r3-dash-topbar-title">LogisticsHQ — RFQ Management</div>
                  <div className="r3-dash-topbar-badge">3 Live RFQs</div>
                </div>

                {/* Body */}
                <div className="r3-dash-body">
                  {/* Sidebar */}
                  <div className="r3-dash-sidebar">
                    {[
                      { icon: '📋', label: 'My RFQs', badge: '3', active: true },
                      { icon: '📥', label: 'Quotes', badge: '12', active: false },
                      { icon: '⚖️', label: 'Compare', badge: null, active: false },
                      { icon: '✅', label: 'Awarded', badge: null, active: false },
                      { icon: '📡', label: 'Tracking', badge: null, active: false },
                      { icon: '📊', label: 'Analytics', badge: null, active: false },
                    ].map((item, i) => (
                      <div key={i} className={`r3-dash-nav-item ${item.active ? 'active' : ''}`}>
                        <span className="r3-dash-nav-icon">{item.icon}</span>
                        {item.label}
                        {item.badge && <span className="r3-dash-nav-badge">{item.badge}</span>}
                      </div>
                    ))}
                  </div>

                  {/* Main */}
                  <div className="r3-dash-main">
                    <div className="r3-dash-main-title">
                      <span>Active RFQs</span>
                      <span style={{ fontSize: '0.72rem', fontWeight: 600, color: 'var(--r3-blue-d)', cursor: 'pointer' }}>+ New RFQ</span>
                    </div>
                    <div className="r3-dash-stats-row">
                      {[
                        { val: '₹1.2Cr', label: 'Savings This Month' },
                        { val: '98%', label: 'Quote Response Rate' },
                        { val: '12 min', label: 'Avg. Procurement Time' },
                      ].map((s, i) => (
                        <div key={i} className="r3-dash-stat">
                          <div className="r3-dash-stat-val">{s.val}</div>
                          <div className="r3-dash-stat-label">{s.label}</div>
                        </div>
                      ))}
                    </div>
                    <div className="r3-dash-rfq-list">
                      {[
                        { id: '#RFQ-9924', route: 'Mumbai → Pune', meta: 'FTL · 18T', status: 'r3-status-live', label: 'Live · 8 quotes' },
                        { id: '#RFQ-9923', route: 'Delhi → Surat', meta: 'LTL · 4T', status: 'r3-status-pending', label: 'Awaiting · 3/8' },
                        { id: '#RFQ-9921', route: 'Mumbai → Chennai', meta: 'FTL · 22T', status: 'r3-status-done', label: 'Awarded ✓' },
                      ].map((r, i) => (
                        <div key={i} className="r3-dash-rfq-row">
                          <div className="r3-dash-rfq-id">{r.id}</div>
                          <div>
                            <div className="r3-dash-rfq-route">{r.route}</div>
                            <div className="r3-dash-rfq-meta">{r.meta}</div>
                          </div>
                          <div className={`r3-dash-rfq-status ${r.status}`}>{r.label}</div>
                        </div>
                      ))}
                    </div>
                  </div>

                  {/* Right panel */}
                  <div className="r3-dash-panel">
                    <div className="r3-dash-panel-title">Quote Comparison · #RFQ-9921</div>
                    <div className="r3-dash-bid-preview">
                      {[
                        { v: 'SafeExpress', p: '₹46,200', best: true },
                        { v: 'TCI Freight', p: '₹47,800', best: false },
                        { v: 'VRL Logistics', p: '₹49,500', best: false },
                      ].map((b, i) => (
                        <div key={i} className={`r3-dash-bid-preview-row ${b.best ? 'best' : ''}`}>
                          <span className="r3-dash-bid-v">{b.v}</span>
                          <span className="r3-dash-bid-p">{b.p}</span>
                        </div>
                      ))}
                    </div>
                    <div className="r3-dash-margin">
                      <div className="r3-dash-margin-title">Your Margin</div>
                      <div className="r3-dash-margin-bar">
                        <div className="r3-dash-margin-fill" style={{ width: '28%' }} />
                      </div>
                      <div className="r3-dash-margin-label">
                        <span>Vendor Cost</span>
                        <span style={{ color: 'var(--r3-teal)', fontWeight: 700 }}>+18.4%</span>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </R>
        </div>
      </section>

      {/* ═══════════════════════════════════════
          § 6 — KEY BENEFITS (STATS WITH LOGISTICS BG)
          ═══════════════════════════════════════ */}
      <section 
        className="r3-stats" 
        id="rfq-stats"
        style={{ backgroundImage: 'url(https://images.unsplash.com/photo-1501523460185-2aa5d2a0f981?auto=format&fit=crop&w=2400&q=80)' }}
      >
        <div className="r3-stats-overlay" />
        <div className="r3-wrap" style={{ position: 'relative', zIndex: 1 }}>
          <R>
            <div className="r3-stats-grid">
              {[
                { val: 10000, suffix: '+', label: 'Carrier Network', color: 'r3-c-navy', decimals: 0 },
                { val: 15, suffix: ' Min', label: 'Average Response Time', color: 'r3-c-blue', decimals: 0 },
                { val: 14.2, suffix: '%', label: 'Average Savings', color: 'r3-c-teal', decimals: 1 },
                { val: 98, suffix: '%', label: 'RFQ Visibility', color: 'r3-c-indigo', decimals: 0 },
              ].map((s, i) => (
                <div key={i} className="r3-stat-cell">
                  <div className={`r3-stat-number ${s.color}`}>
                    <Counter to={s.val} decimals={s.decimals} suffix={s.suffix} />
                  </div>
                  <div className="r3-stat-label">{s.label}</div>
                </div>
              ))}
            </div>
          </R>
        </div>
      </section>

      {/* ═══════════════════════════════════════
          § 7 — AUTOMATION
          ═══════════════════════════════════════ */}
      <section className="r3-auto" id="rfq-auto">
        <div className="r3-wrap">
          <div className="r3-auto-grid">
            
            {/* Left — Real logistics image */}
            <R d="r3-d1">
              <div className="r3-auto-image-side">
                <img src="https://images.unsplash.com/photo-1542744173-8e7e53415bb0?auto=format&fit=crop&w=1200&q=80" alt="Logistics Command Center" />
                <div className="r3-auto-image-badge">
                  <span className="r3-auto-badge-icon">🤖</span>
                  <span className="r3-auto-badge-text">Automated<br/>Dispatch Operations</span>
                </div>
              </div>
            </R>

            {/* Right — Flow & Copy */}
            <R d="r3-d2">
              <span className="r3-section-kicker">Automation</span>
              <h2 className="r3-auto-headline">
                Your Team Should Focus<br />On Growth.<br />Not Follow-Ups.
              </h2>
              <p className="r3-auto-sub">
                Every step from vendor notification to quote collection happens automatically. Your operations team goes from chasing transporters to reviewing the best rates.
              </p>
              
              <div className="r3-auto-flow" style={{ marginTop: '32px' }}>
                {AUTO_STEPS.slice(0, 3).map((s, i) => (
                  <div key={i} className="r3-auto-step">
                    {i < 2 && <div className="r3-auto-step-line" />}
                    <div className="r3-auto-icon">{s.icon}</div>
                    <div>
                      <div className="r3-auto-text-title">{s.title}</div>
                      <div className="r3-auto-text-sub">{s.sub}</div>
                    </div>
                    <div className="r3-auto-auto-badge">Auto</div>
                  </div>
                ))}
              </div>
            </R>
          </div>
        </div>
      </section>



      {/* ═══════════════════════════════════════
          § 9 — FINAL CTA (CINEMATIC)
          ═══════════════════════════════════════ */}
      <section 
        className="r3-cta" 
        id="rfq-cta"
        style={{ backgroundImage: 'url(https://images.unsplash.com/photo-1616423640778-28d1b53229bd?auto=format&fit=crop&w=2400&q=80)' }}
      >
        <div className="r3-cta-overlay" />
        <div className="r3-wrap-sm">
          <R>
            <h2 className="r3-cta-headline">
              Let Vendors Compete.<br />
              <span className="r3-cta-accent">You Keep The Savings.</span>
            </h2>
            <p className="r3-cta-sub">
              Create your first RFQ in minutes and discover better freight rates instantly. Join a nationwide logistics network.
            </p>
            <div className="r3-cta-buttons">
              <Link to="/contact" className="r3-btn-cta-primary">
                Start Your First RFQ →
              </Link>
              <Link to="/contact" className="r3-btn-cta-ghost">
                Book Demo
              </Link>
            </div>
            <div className="r3-cta-trust">
              {['No setup fees', 'No lengthy onboarding', 'Results from day one'].map((t, i) => (
                <span key={i} className="r3-cta-trust-item">{t}</span>
              ))}
            </div>
          </R>
        </div>
      </section>

    </div>
  );
}

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
const stats = [
  { value: '200+', label: 'Carrier Rates' },
  { value: 'Live', label: 'Dynamic Pricing' },
  { value: '3 Modes', label: 'Road, Air, Sea' },
  { value: 'Auto', label: 'Margin Calc' },
];




/* ═══════════════════════════════════════════ */
/*          RATE COMPARISON PAGE              */
/* ═══════════════════════════════════════════ */
export default function RateComparison() {
  const [selectedMode, setSelectedMode] = useState('sea'); // 'sea', 'air', 'road'
  
  // Volumetric Weight Calculator tool state variables
  const [volLength, setVolLength] = useState(120);
  const [volWidth, setVolWidth] = useState(80);
  const [volHeight, setVolHeight] = useState(60);
  const [volWeight, setVolWeight] = useState(150);

  const volCbm = ((volLength * volWidth * volHeight) / 1000000).toFixed(3);
  const volumetricWeight = Math.round((volLength * volWidth * volHeight) / 5000);
  const chargeableWeight = Math.max(volWeight, volumetricWeight);
  
  const [modesData, setModesData] = useState({
    sea: {
      lane: 'Shenzhen Port → JNPT Port',
      details: '🚢 FCL 40ft High Cube • 21 Days transit',
      carriers: [
        { name: 'Maersk Line', rate: '₹2,38,500', time: 'Direct', rank: 1, active: true },
        { name: 'MSC Shipping', rate: '₹2,46,000', time: '1 stop', rank: 2, active: false },
        { name: 'CMA CGM Group', rate: '₹2,55,000', time: 'Direct', rank: 3, active: false }
      ],
      margin: '+₹42,500 (17.8%)'
    },
    air: {
      lane: 'Delhi (DEL) → London (LHR)',
      details: '✈️ Air Cargo • 2-3 Days transit',
      carriers: [
        { name: 'Emirates SkyCargo', rate: '₹185 / kg', time: 'Direct', rank: 1, active: true },
        { name: 'Lufthansa Cargo', rate: '₹198 / kg', time: 'Direct', rank: 2, active: false },
        { name: 'Qatar Airways Cargo', rate: '₹210 / kg', time: '1 stop', rank: 3, active: false }
      ],
      margin: '+₹34 / kg (18.4%)'
    },
    road: {
      lane: 'Mumbai (BOM) → Bangalore (BLR)',
      details: '🚚 Full Truckload (FTL) • 36 Hours transit',
      carriers: [
        { name: 'TCI Freight', rate: '₹46,200', time: 'Direct', rank: 1, active: true },
        { name: 'SafeExpress FTL', rate: '₹48,500', time: 'Direct', rank: 2, active: false },
        { name: 'VRL Logistics', rate: '₹51,000', time: 'Direct', rank: 3, active: false }
      ],
      margin: '+₹8,920 (19.3%)'
    }
  });

  // Dynamic Price Fluctuations Effect
  useEffect(() => {
    const interval = setInterval(() => {
      setModesData(prevData => {
        const nextData = { ...prevData };
        
        // Pick active mode carriers to update
        const activeMode = selectedMode;
        const carriers = [...nextData[activeMode].carriers];
        
        const updateIndex = Math.floor(Math.random() * carriers.length);
        const isPerKg = carriers[updateIndex].rate.includes('/ kg');
        
        const currentRateStr = carriers[updateIndex].rate;
        const currentRateVal = isPerKg
          ? parseInt(currentRateStr.replace('₹', '').replace('/ kg', '').trim())
          : parseInt(currentRateStr.replace('₹', '').replace(/,/g, '').trim());

        // Apply a small fluctuation
        const delta = isPerKg ? (Math.floor(Math.random() * 6) - 3) : (Math.floor(Math.random() * 2000) - 1000);
        let newRateVal = currentRateVal + delta;
        if (isPerKg) newRateVal = Math.max(160, newRateVal);
        else newRateVal = Math.max(40000, newRateVal);

        carriers[updateIndex] = {
          ...carriers[updateIndex],
          rate: isPerKg ? `₹${newRateVal} / kg` : `₹${newRateVal.toLocaleString('en-IN')}`,
          active: true
        };

        // Reset active dots on other carriers
        carriers.forEach((c, idx) => {
          if (idx !== updateIndex) {
            carriers[idx] = { ...c, active: false };
          }
        });

        // Re-calculate ranks
        const sorted = [...carriers].sort((a, b) => {
          const parseRate = (r) => isPerKg
            ? parseInt(r.rate.replace('₹', '').replace('/ kg', '').trim())
            : parseInt(r.rate.replace('₹', '').replace(/,/g, '').trim());
          return parseRate(a) - parseRate(b);
        });

        nextData[activeMode] = {
          ...nextData[activeMode],
          carriers: carriers.map(c => {
            const rank = sorted.findIndex(s => s.name === c.name) + 1;
            return { ...c, rank };
          })
        };

        return nextData;
      });
    }, 4500);

    return () => clearInterval(interval);
  }, [selectedMode]);

  return (
    <>
      {/* ═══ HERO ═══ */}
      <section className="rfq-split-hero" style={{ position: 'relative', overflow: 'hidden', padding: '120px 16px 80px' }} id="rate-hero">
        <div style={{ maxWidth: '1200px', margin: '0 auto', display: 'grid', gridTemplateColumns: '1.2fr 1fr', gap: '48px', alignItems: 'center' }} className="rfq-hero-grid">
          {/* Left Column: Premium Content */}
          <Reveal>
            <div style={{ zIndex: 1 }}>
              <div style={{ display: 'inline-flex', alignItems: 'center', gap: '8px', padding: '6px 18px', borderRadius: '999px', background: 'rgba(0,191,165,0.08)', border: '1px solid rgba(0,191,165,0.2)', fontSize: '0.85rem', fontWeight: 700, color: 'var(--color-brand-teal)', marginBottom: '24px' }}>
                <span className="live-pulse"></span> DYNAMIC RATE INTELLIGENCE
              </div>
              <h1 style={{ fontSize: 'clamp(2.5rem, 5.5vw, 4rem)', fontWeight: 900, lineHeight: 1.1, letterSpacing: '-0.03em', color: 'var(--color-brand-navy)', marginBottom: '20px' }}>
                Compare Rates <br />
                <span className="text-gradient bg-gradient-to-r from-brand-teal via-cyan-500 to-brand-indigo">
                  Across 200+ Carriers
                </span>
              </h1>
              <p style={{ fontSize: '1.15rem', color: 'var(--color-ui-gray)', lineHeight: 1.7, maxWidth: '580px', marginBottom: '36px', fontWeight: 300 }}>
                Stop checking dozen carrier portals manually. Analyze and compare live freight rates across Road, Air, and Ocean modes instantly with integrated margin security.
              </p>
              
              <div style={{ display: 'flex', flexWrap: 'wrap', gap: '16px' }}>
                <Link to="/contact" className="solutions-cta-btn" style={{ fontSize: '1.05rem', padding: '16px 36px' }}>
                  Start Rate Comparison →
                </Link>
                <a href="#how-it-works" className="solutions-cta-btn" style={{ background: 'white', color: 'var(--color-brand-navy)', border: '1px solid var(--color-ui-border)', boxShadow: 'none', fontSize: '1.05rem', padding: '16px 36px' }}>
                  Analyze Factors ↓
                </a>
              </div>
              
              {/* Telemetry Micro-Stat Row */}
              <div className="rfq-hero-stats" style={{ display: 'flex', gap: '32px', marginTop: '48px', borderTop: '1px solid var(--color-ui-border)', paddingTop: '24px' }}>
                {stats.map((s, i) => (
                  <div key={i}>
                    <div style={{ fontSize: '1.5rem', fontWeight: 800, color: 'var(--color-brand-navy)' }}>{s.value}</div>
                    <div style={{ fontSize: '0.75rem', fontWeight: 600, color: 'var(--color-ui-gray)', textTransform: 'uppercase', marginTop: '2px' }}>{s.label}</div>
                  </div>
                ))}
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
                {/* Segmented Controller Tabs */}
                <div className="rate-selector-row">
                  <button className={`rate-mode-tab ${selectedMode === 'sea' ? 'active' : ''}`} onClick={() => setSelectedMode('sea')}>
                    🚢 Ocean
                  </button>
                  <button className={`rate-mode-tab ${selectedMode === 'air' ? 'active' : ''}`} onClick={() => setSelectedMode('air')}>
                    ✈️ Air
                  </button>
                  <button className={`rate-mode-tab ${selectedMode === 'road' ? 'active' : ''}`} onClick={() => setSelectedMode('road')}>
                    🚚 Road
                  </button>
                </div>

                {/* Cargo Details */}
                <div style={{ background: 'var(--color-brand-navy)', borderRadius: '16px', padding: '16px', color: 'white', marginBottom: '20px', position: 'relative', overflow: 'hidden' }}>
                  <div style={{ position: 'absolute', top: '-20px', right: '-20px', width: '80px', height: '80px', background: 'rgba(255,255,255,0.05)', borderRadius: '50%' }}></div>
                  <div style={{ fontSize: '0.7rem', textTransform: 'uppercase', color: '#94A3B8', fontWeight: 700, letterSpacing: '0.05em' }}>Freight Lane Quote</div>
                  <div style={{ fontSize: '1.05rem', fontWeight: 800, marginTop: '4px' }}>
                    {modesData[selectedMode].lane}
                  </div>
                  <div style={{ fontSize: '0.75rem', color: '#CBD5E1', marginTop: '6px' }}>
                    {modesData[selectedMode].details}
                  </div>
                </div>

                {/* Bidding Telemetry Header */}
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '12px' }}>
                  <div style={{ fontSize: '0.8rem', fontWeight: 700, color: 'var(--color-brand-navy)' }}>Carrier Rates Breakdown</div>
                  <div style={{ fontSize: '0.75rem', color: 'var(--color-brand-teal)', fontWeight: 700, display: 'flex', alignItems: 'center', gap: '4px' }}>
                    <span className="live-pulse" style={{ width: '6px', height: '6px' }}></span> Live Feed
                  </div>
                </div>

                {/* Bids List Container */}
                <div style={{ display: 'flex', flexDirection: 'column', gap: '10px' }} className="rfq-bids-list">
                  {modesData[selectedMode].carriers.map((c, idx) => (
                    <div key={idx} className={`rfq-bid-card ${c.active ? 'active' : ''}`} style={{
                      display: 'flex',
                      justifyContent: 'space-between',
                      alignItems: 'center',
                      background: c.active ? 'rgba(0,191,165,0.06)' : 'white',
                      border: c.active ? '1px solid rgba(0,191,165,0.3)' : '1px solid rgba(226, 232, 240, 0.8)',
                      borderRadius: '12px',
                      padding: '12px 16px',
                      boxShadow: '0 4px 6px -1px rgba(0,0,0,0.02)',
                      transition: 'all 0.4s cubic-bezier(0.16,1,0.3,1)'
                    }}>
                      <div>
                        <div style={{ fontSize: '0.85rem', fontWeight: 700, color: 'var(--color-brand-navy)', display: 'flex', alignItems: 'center', gap: '6px' }}>
                          {c.name}
                          {c.active && <span style={{ width: '6px', height: '6px', borderRadius: '50%', background: 'var(--color-brand-teal)', display: 'inline-block' }}></span>}
                        </div>
                        <div style={{ fontSize: '0.7rem', color: 'var(--color-ui-gray)', marginTop: '2px' }}>
                          Rank #{c.rank} • {c.time}
                        </div>
                      </div>
                      <div style={{ textAlign: 'right' }}>
                        <div style={{ fontSize: '0.95rem', fontWeight: 800, color: c.rank === 1 ? 'var(--color-brand-teal)' : 'var(--color-brand-navy)' }}>{c.rate}</div>
                        {c.rank === 1 && <span style={{ fontSize: '0.65rem', fontWeight: 700, color: 'var(--color-brand-teal)', background: 'rgba(0,191,165,0.1)', padding: '2px 6px', borderRadius: '4px' }}>BEST VALUE</span>}
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
                    <div style={{ fontSize: '0.65rem', color: 'var(--color-ui-gray)', fontWeight: 600, textTransform: 'uppercase' }}>Auto-Protected Margin</div>
                    <div style={{ fontSize: '0.9rem', fontWeight: 800, color: 'var(--color-brand-navy)' }}>
                      {modesData[selectedMode].margin}
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </Reveal>
        </div>
      </section>

      {/* ═══ WHY DYNAMIC RATES ═══ */}
      <section className="solution-module solution-module-alt">
        <Reveal>
          <div className="solution-module-header">
            <h2 className="solution-module-title">Why Dynamic Rates?</h2>
            <p className="solution-module-desc">
              The difference between legacy procurement and modern rate intelligence.
            </p>
          </div>
        </Reveal>

        <div className="comparison-grid" style={{ maxWidth: '1100px', margin: '0 auto', display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '32px' }}>
          <Reveal delay="reveal-delay-1">
            <div className="comparison-panel comparison-panel-bad" style={{ display: 'flex', flexDirection: 'column', height: '100%', justifyContent: 'space-between' }}>
              <div>
                <h3 style={{ fontSize: '1.25rem', fontWeight: 700, color: 'var(--color-brand-navy)', marginBottom: '16px', display: 'flex', alignItems: 'center', gap: '8px' }}>
                  <span style={{ background: '#FEE2E2', width: '36px', height: '36px', borderRadius: '50%', display: 'inline-flex', alignItems: 'center', justifyContent: 'center' }}>❌</span> Without Freel
                </h3>
                <ul className="comparison-list" style={{ fontSize: '0.85rem', color: 'var(--color-ui-gray)', display: 'flex', flexDirection: 'column', gap: '10px' }}>
                  <li>Call 5-10 vendors manually for each quote</li>
                  <li>Rates valid for 24-48 hours only</li>
                  <li>No margin visibility until final invoice</li>
                  <li>Excel sheets with rapidly outdated rates</li>
                  <li>Takes 2-3 days to finalize a simple rate</li>
                </ul>
              </div>
              <div className="messy-excel-simulator" style={{ marginTop: '24px', opacity: 0.85, background: '#f8fafc', border: '1px solid #e2e8f0', borderRadius: '12px', padding: '12px', overflow: 'hidden' }}>
                <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr 1fr', gap: '4px', fontSize: '0.65rem', fontFamily: 'monospace', color: '#64748b' }}>
                  <div style={{ background: '#e2e8f0', padding: '4px', fontWeight: 'bold' }}>Lane ID</div>
                  <div style={{ background: '#e2e8f0', padding: '4px', fontWeight: 'bold' }}>Carrier Quote</div>
                  <div style={{ background: '#e2e8f0', padding: '4px', fontWeight: 'bold' }}>Valid Until</div>
                  
                  <div style={{ borderBottom: '1px solid #e2e8f0', padding: '4px' }}>BOM-BLR-01</div>
                  <div style={{ borderBottom: '1px solid #e2e8f0', padding: '4px', textDecoration: 'line-through' }}>₹46,200</div>
                  <div style={{ borderBottom: '1px solid #e2e8f0', padding: '4px', color: 'red' }}>EXPIRED</div>
                  
                  <div style={{ borderBottom: '1px solid #e2e8f0', padding: '4px' }}>BOM-BLR-02</div>
                  <div style={{ borderBottom: '1px solid #e2e8f0', padding: '4px', textDecoration: 'line-through' }}>₹48,500</div>
                  <div style={{ borderBottom: '1px solid #e2e8f0', padding: '4px', color: 'red' }}>EXPIRED</div>

                  <div style={{ padding: '4px' }}>BOM-BLR-03</div>
                  <div style={{ padding: '4px', color: '#f59e0b' }}>CALL CARRIER</div>
                  <div style={{ padding: '4px' }}>--</div>
                </div>
              </div>
            </div>
          </Reveal>

          <Reveal delay="reveal-delay-2">
            <div className="comparison-panel comparison-panel-good" style={{ display: 'flex', flexDirection: 'column', height: '100%', justifyContent: 'space-between' }}>
              <div>
                <h3 style={{ fontSize: '1.25rem', fontWeight: 700, color: 'var(--color-brand-navy)', marginBottom: '16px', display: 'flex', alignItems: 'center', gap: '8px' }}>
                  <span style={{ background: '#D1FAE5', width: '36px', height: '36px', borderRadius: '50%', display: 'inline-flex', alignItems: 'center', justifyContent: 'center' }}>✅</span> With Freel
                </h3>
                <ul className="comparison-list" style={{ fontSize: '0.85rem', color: 'var(--color-ui-gray)', display: 'flex', flexDirection: 'column', gap: '10px' }}>
                  <li>Compare 10+ carriers in under 30 seconds</li>
                  <li>Dynamic rates updated every single hour</li>
                  <li>Margin auto-calculated per quote instantly</li>
                  <li>Digital rate cards from verified vendors</li>
                  <li>Instant quotation PDF with one click</li>
                </ul>
              </div>
              <div style={{ marginTop: '24px', overflow: 'hidden', borderRadius: '12px', border: '1px solid var(--color-ui-border)', boxShadow: '0 8px 24px rgba(0,0,0,0.04)' }}>
                <img src="/assets/digital_platform_dashboard.png" alt="With Freel Analytics Dashboard" style={{ width: '100%', height: '115px', objectFit: 'cover', display: 'block' }} className="hover-scale-img" />
              </div>
            </div>
          </Reveal>
        </div>
      </section>

      {/* ═══ GLOBAL LANE COMPARISON SHOWCASE ═══ */}
      <section className="solution-module">
        <Reveal>
          <div className="solution-module-header">
            <h2 className="solution-module-title">Global Shipping Lane Intel</h2>
            <p className="solution-module-desc">
              Real-time actionable metrics across major maritime routes and air corridors. Compare carrier performance metrics and current index rates instantly.
            </p>
          </div>
        </Reveal>

        <div style={{ maxWidth: '1100px', margin: '0 auto', display: 'grid', gridTemplateColumns: '1fr 1.3fr', gap: '32px', alignItems: 'center' }} className="rfq-hero-grid">
          {/* Left: Active lanes list */}
          <Reveal delay="reveal-delay-1">
            <div style={{ display: 'flex', flexDirection: 'column', gap: '16px' }}>
              <div style={{ background: 'white', border: '1px solid var(--color-ui-border)', borderRadius: '16px', padding: '20px', boxShadow: '0 4px 12px rgba(0,0,0,0.02)', transition: 'transform 0.3s ease' }} className="rfq-bid-card">
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '12px' }}>
                  <span style={{ fontSize: '0.75rem', fontWeight: 700, color: 'var(--color-brand-teal)', background: 'rgba(0,191,165,0.08)', padding: '4px 10px', borderRadius: '6px' }}>SHANGHAI ➔ ROTTERDAM</span>
                  <span style={{ fontSize: '0.85rem', fontWeight: 800, color: 'var(--color-brand-navy)' }}>$3,420 <span style={{ color: 'var(--color-brand-teal)', fontSize: '0.7rem' }}>➔ 4.2%</span></span>
                </div>
                <div style={{ display: 'flex', gap: '16px', fontSize: '0.75rem', color: 'var(--color-ui-gray)' }}>
                  <span>🚢 Ocean FCL</span>
                  <span>⏱️ 28 Days</span>
                  <span>🏆 MSC (Best Rate)</span>
                </div>
              </div>

              <div style={{ background: 'white', border: '1px solid var(--color-ui-border)', borderRadius: '16px', padding: '20px', boxShadow: '0 4px 12px rgba(0,0,0,0.02)', transition: 'transform 0.3s ease' }} className="rfq-bid-card">
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '12px' }}>
                  <span style={{ fontSize: '0.75rem', fontWeight: 700, color: 'var(--color-brand-indigo)', background: 'rgba(90,79,207,0.08)', padding: '4px 10px', borderRadius: '6px' }}>MUMBAI ➔ DUBAI (DXB)</span>
                  <span style={{ fontSize: '0.85rem', fontWeight: 800, color: 'var(--color-brand-navy)' }}>₹95 / kg <span style={{ color: 'red', fontSize: '0.7rem' }}>➔ 1.8%</span></span>
                </div>
                <div style={{ display: 'flex', gap: '16px', fontSize: '0.75rem', color: 'var(--color-ui-gray)' }}>
                  <span>✈️ Air Cargo</span>
                  <span>⏱️ 4 Hours</span>
                  <span>🏆 Air India (Best Value)</span>
                </div>
              </div>

              <div style={{ background: 'white', border: '1px solid var(--color-ui-border)', borderRadius: '16px', padding: '20px', boxShadow: '0 4px 12px rgba(0,0,0,0.02)', transition: 'transform 0.3s ease' }} className="rfq-bid-card">
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '12px' }}>
                  <span style={{ fontSize: '0.75rem', fontWeight: 700, color: 'var(--color-brand-teal)', background: 'rgba(0,191,165,0.08)', padding: '4px 10px', borderRadius: '6px' }}>SINGAPORE ➔ NEW YORK</span>
                  <span style={{ fontSize: '0.85rem', fontWeight: 800, color: 'var(--color-brand-navy)' }}>$4,890 <span style={{ color: 'var(--color-brand-teal)', fontSize: '0.7rem' }}>➔ 3.5%</span></span>
                </div>
                <div style={{ display: 'flex', gap: '16px', fontSize: '0.75rem', color: 'var(--color-ui-gray)' }}>
                  <span>🚢 Ocean FCL</span>
                  <span>⏱️ 34 Days</span>
                  <span>🏆 ONE Line (Best Rate)</span>
                </div>
              </div>
            </div>
          </Reveal>

          {/* Right: Premium Global Lanes Map */}
          <Reveal delay="reveal-delay-2">
            <div style={{ overflow: 'hidden', borderRadius: '24px', border: '1px solid var(--color-ui-border)', boxShadow: '0 20px 40px -15px rgba(0,0,0,0.08)', position: 'relative' }}>
              <img src="/assets/smart_routing.png" alt="Global shipping lanes map visualization" style={{ width: '100%', height: '320px', objectFit: 'cover', display: 'block', transition: 'transform 0.5s ease' }} className="hover-scale-img" />
              <div style={{ position: 'absolute', bottom: '16px', left: '16px', background: 'rgba(15,23,42,0.85)', backdropFilter: 'blur(10px)', color: 'white', padding: '10px 16px', borderRadius: '12px', fontSize: '0.75rem', fontWeight: 600 }}>
                🟢 384 active global carrier lanes tracked live
              </div>
            </div>
          </Reveal>
        </div>
      </section>

      {/* ═══ WHAT AFFECTS RATES ═══ */}
      <section className="solution-module solution-module-alt">
        <Reveal>
          <div className="solution-module-header">
            <h2 className="solution-module-title">What Affects Rates?</h2>
            <p className="solution-module-desc">
              Freight pricing isn't arbitrary. Freel's engine analyzes dozens of variables in real-time to ensure you get the most accurate, executable rate.
            </p>
          </div>
        </Reveal>
        
        <div className="feature-grid-3">
          {/* Factor 1: Fuel Surcharge */}
          <Reveal delay="reveal-delay-1">
            <div className="feature-card" style={{ background: 'white', display: 'flex', flexDirection: 'column', justifyContent: 'space-between', height: '100%' }}>
              <div>
                <div className="feature-card-icon" style={{ background: 'var(--color-ui-surface)', padding: '12px', borderRadius: '12px', display: 'inline-block', marginBottom: '16px' }}>⛽</div>
                <h3 className="feature-card-title">Fuel Surcharge</h3>
                <p className="feature-card-desc">Diesel/bunker fuel prices directly impact road and sea freight. Freel tracks global fuel indices and adjusts rate calculations in real-time.</p>
              </div>
              <div style={{ marginTop: '16px', display: 'flex', flexDirection: 'column', gap: '8px' }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: '0.7rem', color: '#64748B', fontWeight: 600 }}>
                  <span>Brent Crude (Global Index)</span>
                  <span style={{ color: 'red' }}>$84.50 (➔ +2.4%)</span>
                </div>
                <div style={{ height: '8px', background: '#F1F5F9', borderRadius: '4px', overflow: 'hidden' }}>
                  <div style={{ width: '84%', height: '100%', background: 'linear-gradient(90deg, var(--color-brand-indigo), red)' }}></div>
                </div>
                <div style={{ fontSize: '0.65rem', color: '#94A3B8' }}>⚠️ Automatically updates active carrier road/sea quotes hourly</div>
              </div>
            </div>
          </Reveal>

          {/* Factor 2: Container Type */}
          <Reveal delay="reveal-delay-2">
            <div className="feature-card" style={{ background: 'white', display: 'flex', flexDirection: 'column', justifyContent: 'space-between', height: '100%' }}>
              <div>
                <div className="feature-card-icon" style={{ background: 'var(--color-ui-surface)', padding: '12px', borderRadius: '12px', display: 'inline-block', marginBottom: '16px' }}>📦</div>
                <h3 className="feature-card-title">Container Type</h3>
                <p className="feature-card-desc">20ft, 40ft, High Cube, Reefer — each has different pricing structures. Specialized equipment like open top command premium rates.</p>
              </div>
              <div style={{ marginTop: '16px', background: '#F8FAFC', borderRadius: '12px', padding: '12px', border: '1px solid #E2E8F0', fontSize: '0.75rem' }}>
                <div style={{ display: 'flex', gap: '8px', marginBottom: '8px' }}>
                  <span style={{ background: 'rgba(90,79,207,0.1)', color: 'var(--color-brand-indigo)', padding: '2px 8px', borderRadius: '4px', fontWeight: 700 }}>20ft FCL</span>
                  <span style={{ background: 'rgba(0,191,165,0.1)', color: 'var(--color-brand-teal)', padding: '2px 8px', borderRadius: '4px', fontWeight: 700 }}>40ft HC</span>
                </div>
                <div style={{ color: '#64748B', lineHeight: 1.5 }}>
                  Max Payload: <b>28,600 kg</b><br />
                  Internal Vol: <b>76.2 CBM</b>
                </div>
              </div>
            </div>
          </Reveal>

          {/* Factor 3: HAZ / DG Cargo */}
          <Reveal delay="reveal-delay-3">
            <div className="feature-card" style={{ background: 'white', display: 'flex', flexDirection: 'column', justifyContent: 'space-between', height: '100%' }}>
              <div>
                <div className="feature-card-icon" style={{ background: 'var(--color-ui-surface)', padding: '12px', borderRadius: '12px', display: 'inline-block', marginBottom: '16px' }}>⚠️</div>
                <h3 className="feature-card-title">HAZ / DG Cargo</h3>
                <p className="feature-card-desc">Hazardous goods require MSDS documentation, special handling, and carry a 15-40% surcharge depending on IMO classification.</p>
              </div>
              <div style={{ marginTop: '16px', border: '1px dashed #F59E0B', background: 'rgba(245,158,11,0.06)', borderRadius: '12px', padding: '12px', fontSize: '0.75rem', color: '#D97706' }}>
                ⚠️ <b>DG Class 3, 6 & 9 Supported</b><br />
                MSDS validation completes in under 12 minutes with automated regulatory checkups.
              </div>
            </div>
          </Reveal>

          {/* Factor 4: Season & Demand */}
          <Reveal delay="reveal-delay-1">
            <div className="feature-card" style={{ background: 'white', display: 'flex', flexDirection: 'column', justifyContent: 'space-between', height: '100%' }}>
              <div>
                <div className="feature-card-icon" style={{ background: 'var(--color-ui-surface)', padding: '12px', borderRadius: '12px', display: 'inline-block', marginBottom: '16px' }}>📅</div>
                <h3 className="feature-card-title">Season & Demand</h3>
                <p className="feature-card-desc">Peak season (Oct-Mar), regional holiday periods, and port congestion events cause sudden rate spikes of 20-50%.</p>
              </div>
              <div style={{ marginTop: '16px', background: '#F8FAFC', border: '1px solid #E2E8F0', borderRadius: '12px', padding: '12px', fontSize: '0.75rem' }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', color: '#64748B', marginBottom: '6px' }}>
                  <span>Peak Season Surcharge (PSS)</span>
                  <span style={{ color: 'var(--color-brand-teal)', fontWeight: 700 }}>ACTIVE</span>
                </div>
                <div style={{ height: '24px', display: 'flex', alignItems: 'flex-end', gap: '4px' }}>
                  <div style={{ height: '40%', flex: 1, background: '#CBD5E1', borderRadius: '2px' }}></div>
                  <div style={{ height: '45%', flex: 1, background: '#CBD5E1', borderRadius: '2px' }}></div>
                  <div style={{ height: '60%', flex: 1, background: '#CBD5E1', borderRadius: '2px' }}></div>
                  <div style={{ height: '95%', flex: 1, background: 'var(--color-brand-teal)', borderRadius: '2px' }}></div>
                  <div style={{ height: '80%', flex: 1, background: 'var(--color-brand-teal)', borderRadius: '2px' }}></div>
                </div>
                <div style={{ fontSize: '0.65rem', color: '#94A3B8', marginTop: '6px', textAlign: 'center' }}>Q4 Rate Trend Multiplier: 1.28x</div>
              </div>
            </div>
          </Reveal>

          {/* Factor 5: Volumetric Weight Calculator Slider Card */}
          <Reveal delay="reveal-delay-2">
            <div className="feature-card" style={{ background: 'white', border: '1px solid var(--color-brand-teal)', boxShadow: '0 10px 25px -10px rgba(0,191,165,0.15)', display: 'flex', flexDirection: 'column', justifyContent: 'space-between', height: '100%' }}>
              <div>
                <div className="feature-card-icon" style={{ background: 'rgba(0,191,165,0.08)', padding: '12px', borderRadius: '12px', display: 'inline-block', marginBottom: '16px' }}>♟️</div>
                <h3 className="feature-card-title">Weight vs Volume</h3>
                <p className="feature-card-desc">Chargeable weight is calculated as MAX(actual weight, volumetric weight). Freel automatically calculates this using cargo size.</p>
              </div>
              
              <div style={{ marginTop: '16px', background: '#F8FAFC', borderRadius: '12px', padding: '16px', border: '1px solid #E2E8F0', fontSize: '0.8rem' }}>
                <div style={{ fontWeight: 700, color: 'var(--color-brand-navy)', marginBottom: '10px' }}>Volumetric Weight Tool</div>
                
                <div style={{ marginBottom: '8px' }}>
                  <div style={{ display: 'flex', justifyContent: 'space-between', color: '#64748B', fontSize: '0.7rem', fontWeight: 600 }}>
                    <span>Length: {volLength} cm</span>
                  </div>
                  <input type="range" min="20" max="250" value={volLength} onChange={(e) => setVolLength(Number(e.target.value))} style={{ width: '100%', accentColor: 'var(--color-brand-teal)', cursor: 'pointer' }} />
                </div>

                <div style={{ marginBottom: '8px' }}>
                  <div style={{ display: 'flex', justifyContent: 'space-between', color: '#64748B', fontSize: '0.7rem', fontWeight: 600 }}>
                    <span>Width: {volWidth} cm</span>
                  </div>
                  <input type="range" min="20" max="200" value={volWidth} onChange={(e) => setVolWidth(Number(e.target.value))} style={{ width: '100%', accentColor: 'var(--color-brand-teal)', cursor: 'pointer' }} />
                </div>

                <div style={{ marginBottom: '8px' }}>
                  <div style={{ display: 'flex', justifyContent: 'space-between', color: '#64748B', fontSize: '0.7rem', fontWeight: 600 }}>
                    <span>Height: {volHeight} cm</span>
                  </div>
                  <input type="range" min="20" max="200" value={volHeight} onChange={(e) => setVolHeight(Number(e.target.value))} style={{ width: '100%', accentColor: 'var(--color-brand-teal)', cursor: 'pointer' }} />
                </div>

                <div style={{ marginBottom: '12px' }}>
                  <div style={{ display: 'flex', justifyContent: 'space-between', color: '#64748B', fontSize: '0.7rem', fontWeight: 600 }}>
                    <span>Actual Weight: {volWeight} kg</span>
                  </div>
                  <input type="range" min="10" max="500" value={volWeight} onChange={(e) => setVolWeight(Number(e.target.value))} style={{ width: '100%', accentColor: 'var(--color-brand-indigo)', cursor: 'pointer' }} />
                </div>

                <div style={{ borderTop: '1px solid #E2E8F0', paddingTop: '10px', display: 'flex', flexDirection: 'column', gap: '4px' }}>
                  <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                    <span style={{ color: '#64748B' }}>Total Volume:</span>
                    <span style={{ fontWeight: 700 }}>{volCbm} CBM</span>
                  </div>
                  <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                    <span style={{ color: '#64748B' }}>Volumetric Weight:</span>
                    <span style={{ fontWeight: 700, color: volumetricWeight > volWeight ? 'var(--color-brand-teal)' : '#64748B' }}>{volumetricWeight} kg</span>
                  </div>
                  <div style={{ display: 'flex', justifyContent: 'space-between', background: 'rgba(0,191,165,0.06)', padding: '6px', borderRadius: '6px', marginTop: '4px', border: '1px solid rgba(0,191,165,0.2)' }}>
                    <span style={{ fontWeight: 700, color: 'var(--color-brand-navy)' }}>Chargeable Weight:</span>
                    <span style={{ fontWeight: 800, color: 'var(--color-brand-teal)' }}>{chargeableWeight} kg</span>
                  </div>
                </div>
              </div>
            </div>
          </Reveal>

          {/* Factor 6: FCL vs LCL */}
          <Reveal delay="reveal-delay-3">
            <div className="feature-card" style={{ background: 'white', display: 'flex', flexDirection: 'column', justifyContent: 'space-between', height: '100%' }}>
              <div>
                <div className="feature-card-icon" style={{ background: 'var(--color-ui-surface)', padding: '12px', borderRadius: '12px', display: 'inline-block', marginBottom: '16px' }}>🚢</div>
                <h3 className="feature-card-title">FCL vs LCL</h3>
                <p className="feature-card-desc">Full container load vs shared container. LCL rates are quoted per CBM while FCL rates are priced per whole container unit.</p>
              </div>
              <div style={{ marginTop: '16px', background: '#F8FAFC', border: '1px solid #E2E8F0', borderRadius: '12px', padding: '12px', fontSize: '0.75rem', display: 'flex', flexDirection: 'column', gap: '8px' }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', color: '#64748B' }}>
                  <span>📦 FCL (Container Load)</span>
                  <span style={{ fontWeight: 700, color: 'var(--color-brand-indigo)' }}>Flat / Container</span>
                </div>
                <div style={{ display: 'flex', justifyContent: 'space-between', color: '#64748B' }}>
                  <span>📥 LCL (Shared Volume)</span>
                  <span style={{ fontWeight: 700, color: 'var(--color-brand-teal)' }}>Variable / CBM</span>
                </div>
              </div>
            </div>
          </Reveal>
        </div>
      </section>

      {/* ═══ CTA ═══ */}
      <section className="detail-cta-banner">
        <Reveal>
          <h2>Compare Rates in Real-Time</h2>
          <p>Stop calling vendors and managing spreadsheets. Get instant freight rate comparisons and protect your margins today.</p>
          <Link to="/contact" style={{
            display: 'inline-block',
            padding: '16px 40px',
            background: 'white',
            color: 'var(--color-brand-indigo)',
            fontWeight: 700,
            borderRadius: '999px',
            fontSize: '1.1rem',
            textDecoration: 'none',
            transition: 'all 0.3s ease',
            position: 'relative',
            zIndex: 1,
            boxShadow: '0 8px 24px rgba(0,0,0,0.2)'
          }}>
            Try Rate Engine →
          </Link>
        </Reveal>
      </section>
    </>
  );
}

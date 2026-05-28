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

/* ─── Rate Data ─── */
const rateData = {
  road: [
    { name: 'SafeExpress', price: '₹58,000', time: '2 days', badge: 'BEST', badgeClass: 'badge-best', isBest: true },
    { name: 'TCI Freight', price: '₹60,500', time: '2 days', badge: 'FAST', badgeClass: 'badge-fast', isBest: false },
    { name: 'VRL Logistics', price: '₹62,000', time: '3 days', badge: 'RELIABLE', badgeClass: 'badge-reliable', isBest: false },
  ],
  air: [
    { name: 'BlueDart Air', price: '₹1,12,000', time: '12 hrs', badge: 'FASTEST', badgeClass: 'badge-fast', isBest: false },
    { name: 'Air India Cargo', price: '₹1,05,000', time: '18 hrs', badge: 'BEST VALUE', badgeClass: 'badge-best', isBest: true },
    { name: 'SpiceXpress', price: '₹1,08,000', time: '24 hrs', badge: 'ECONOMY', badgeClass: 'badge-economy', isBest: false },
  ],
  sea: [
    { name: 'Maersk Line', price: '₹45,000', time: '14 days', badge: 'RELIABLE', badgeClass: 'badge-reliable', isBest: false },
    { name: 'MSC', price: '₹42,000', time: '16 days', badge: 'LOW COST', badgeClass: 'badge-best', isBest: true },
    { name: 'CMA CGM', price: '₹46,500', time: '13 days', badge: 'FAST', badgeClass: 'badge-fast', isBest: false },
  ],
};

/* ═══════════════════════════════════════════ */
/*          SOLUTIONS LANDING PAGE            */
/* ═══════════════════════════════════════════ */
export default function Solutions() {
  const [activeTab, setActiveTab] = useState('road');

  return (
    <>
      {/* ═══ HERO ═══ */}
      <section className="detail-hero" style={{ paddingBottom: '48px' }}>
        <div style={{ position: 'relative', zIndex: 1 }}>
          <h1 className="detail-hero-title" style={{ fontSize: 'clamp(2.2rem, 5vw, 3.2rem)' }}>
            Powerful Solutions, <br />
            <span className="text-gradient bg-gradient-to-r from-brand-teal to-brand-indigo">One Platform</span>
          </h1>
          <p className="detail-hero-sub">
            Preview every feature before you sign up. No functionality exposed — just a taste of what awaits inside your unified logistics dashboard.
          </p>
        </div>
      </section>

      {/* ═══ RFQ MANAGEMENT ═══ */}
      <section className="solution-module solution-module-alt" id="rfq">
        <Reveal>
          <div className="solution-module-header">
            <div className="solution-module-icon">📋</div>
            <h2 className="solution-module-title">RFQ Management</h2>
            <p className="solution-module-desc">
              Don't have rates? Send RFQs to multiple vendors simultaneously. Auto-collect responses, compare side-by-side, and pick the best deal — all within Freel.
            </p>
          </div>
        </Reveal>

        <Reveal>
          <div className="process-steps-row">
            {[
              { n: '1', icon: '📝', title: 'Create RFQ', desc: 'Specify cargo, route, mode' },
              { n: '2', icon: '📤', title: 'Broadcast', desc: 'Send to 5-10 vendors' },
              { n: '3', icon: '📊', title: 'Collect & Compare', desc: 'Auto-track responses' },
              { n: '4', icon: '✅', title: 'Award & Quote', desc: 'Generate PDF for client' },
            ].map((s, i) => (
              <div key={i} className="process-step-card">
                <div className="process-step-num">{s.n}</div>
                <h3 className="process-step-title">{s.icon} {s.title}</h3>
                <p className="process-step-desc">{s.desc}</p>
              </div>
            ))}
          </div>
        </Reveal>

        <Reveal delay="reveal-delay-1">
          <div className="rfq-preview">
            <div className="rfq-preview-header">
              <h4>📋 Sample RFQ Preview</h4>
              <span className="rfq-preview-badge">Live UI</span>
            </div>
            <div className="rfq-item">
              <div>
                <div className="rfq-item-id">#RFQ-089 <span className="rfq-item-route">•</span> JNPT → Rotterdam</div>
                <div className="rfq-item-detail">🚢 FCL 40ft <span className="rfq-status rfq-status-dg" style={{ marginLeft: '8px' }}>Chemicals (DG)</span></div>
              </div>
              <span className="rfq-status rfq-status-pending">⏳ 4/6 Quotes</span>
            </div>
            <div className="rfq-item">
              <div>
                <div className="rfq-item-id">#RFQ-088 <span className="rfq-item-route">•</span> DEL → SIN</div>
                <div className="rfq-item-detail">✈️ Air 800kg <span style={{ background: '#E5E7EB', color: '#1E293B', padding: '2px 8px', borderRadius: '4px', fontSize: '0.7rem', fontWeight: 600, marginLeft: '8px' }}>Electronics</span></div>
              </div>
              <span className="rfq-status rfq-status-done">✓ Done</span>
            </div>
            <div style={{ textAlign: 'center', marginTop: '24px' }}>
              <Link to="/solutions/rfq" className="solutions-cta-btn" style={{ fontSize: '0.95rem', padding: '12px 28px' }}>
                Unlock RFQ System →
              </Link>
            </div>
          </div>
        </Reveal>
      </section>

      {/* ═══ RATE COMPARISON ENGINE ═══ */}
      <section className="solution-module" id="rate-engine">
        <Reveal>
          <div className="solution-module-header">
            <div className="solution-module-icon" style={{ color: 'var(--color-brand-indigo)' }}>📈</div>
            <h2 className="solution-module-title">Rate Comparison Engine</h2>
            <p className="solution-module-desc">
              Compare freight rates across 500+ vendors instantly. Filter by mode (Road/Air/Sea), transit time, cost, and reliability.
            </p>
          </div>
        </Reveal>

        <Reveal>
          <div className="rate-engine-panel">
            <div className="rate-tabs">
              {['road', 'air', 'sea'].map(tab => (
                <button
                  key={tab}
                  className={`rate-tab ${activeTab === tab ? 'active' : ''}`}
                  onClick={() => setActiveTab(tab)}
                >
                  {tab === 'road' ? '🚛 Road' : tab === 'air' ? '✈️ Air' : '🚢 Sea'}
                </button>
              ))}
            </div>
            <div className="rate-table-wrap">
              <table className="rate-table">
                <thead>
                  <tr>
                    <th>Vendor</th>
                    <th>Total Cost</th>
                    <th>Transit Time</th>
                    <th style={{ textAlign: 'right' }}>Rating</th>
                  </tr>
                </thead>
                <tbody>
                  {rateData[activeTab].map((row, i) => (
                    <tr key={i}>
                      <td className="vendor-name">{row.name}</td>
                      <td className={row.isBest ? 'best-price' : ''}>{row.price}</td>
                      <td style={{ color: 'var(--color-ui-gray)' }}>{row.time}</td>
                      <td style={{ textAlign: 'right' }}>
                        <span className={`rate-badge ${row.badgeClass}`}>{row.badge}</span>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
              <div style={{ textAlign: 'center', marginTop: '32px' }}>
                <Link to="/solutions/rate-comparison" className="solutions-cta-btn" style={{ background: 'var(--color-brand-indigo)', fontSize: '0.95rem', padding: '12px 28px' }}>
                  Unlock Rate Engine →
                </Link>
              </div>
            </div>
          </div>
        </Reveal>
      </section>

      {/* ═══ SHIPMENT TRACKING ═══ */}
      <section className="solution-module solution-module-alt" id="tracking">
        <Reveal>
          <div className="solution-module-header">
            <div className="solution-module-icon" style={{ color: 'var(--color-brand-teal)' }}>📍</div>
            <h2 className="solution-module-title">Shipment Tracking</h2>
            <p className="solution-module-desc">
              Track every shipment across every mode on one interactive dashboard. GPS for trucks, AIS for vessels, FlightAware for air cargo.
            </p>
          </div>
        </Reveal>

        <Reveal>
          <div className="tracking-mode-grid">
            {[
              { icon: '🚛', title: 'GPS Tracking', desc: 'Real-time truck positions' },
              { icon: '✈️', title: 'Flight Tracking', desc: 'FlightAware integration' },
              { icon: '🚢', title: 'Vessel Tracking', desc: 'AIS / MarineTraffic' },
            ].map((t, i) => (
              <div key={i} className="tracking-mode-card">
                <div className="tracking-mode-icon">{t.icon}</div>
                <div>
                  <div className="tracking-mode-title">{t.title}</div>
                  <div className="tracking-mode-desc">{t.desc}</div>
                </div>
              </div>
            ))}
          </div>
        </Reveal>

        <Reveal delay="reveal-delay-1">
          <div style={{ maxWidth: '800px', margin: '48px auto 0' }}>
            <div style={{
              position: 'relative',
              width: '100%',
              height: '300px',
              background: '#0F172A',
              borderRadius: '16px',
              overflow: 'hidden',
              border: '2px dashed #334155',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center'
            }}>
              <div style={{
                background: 'rgba(255,255,255,0.95)',
                padding: '20px 28px',
                borderRadius: '12px',
                textAlign: 'center',
                zIndex: 1
              }}>
                <div style={{ fontSize: '2rem', marginBottom: '8px' }}>🗺️</div>
                <h3 style={{ fontWeight: 700, color: '#1E293B' }}>Live Tracking Map Preview</h3>
                <p style={{ fontSize: '0.85rem', color: '#64748B', marginTop: '4px' }}>Sign up to unlock real-time shipment tracking</p>
              </div>
            </div>
            <div style={{ textAlign: 'center', marginTop: '32px' }}>
              <Link to="/solutions/tracking" className="solutions-cta-btn" style={{ fontSize: '0.95rem', padding: '12px 28px' }}>
                Unlock Live Tracking →
              </Link>
            </div>
          </div>
        </Reveal>
      </section>

      {/* ═══ COMPLIANCE & KYC ═══ */}
      <section className="solution-module" id="compliance">
        <Reveal>
          <div className="solution-module-header">
            <div className="solution-module-icon" style={{ color: '#10B981' }}>🛡️</div>
            <h2 className="solution-module-title">Compliance & KYC</h2>
            <p className="solution-module-desc">
              Automated HSN code lookup, MSDS management, KYC verification, and legal document vault.
            </p>
          </div>
        </Reveal>

        <Reveal>
          <div className="compliance-grid">
            {[
              { icon: '🔍', title: 'HSN Lookup' },
              { icon: '📋', title: 'MSDS Manager' },
              { icon: '🏢', title: 'KYC Center' },
              { icon: '📁', title: 'Doc Vault' },
            ].map((c, i) => (
              <div key={i} className="compliance-card">
                <div className="compliance-card-icon">{c.icon}</div>
                <div className="compliance-card-title">{c.title}</div>
              </div>
            ))}
          </div>
        </Reveal>
      </section>

      {/* ═══ CTA ═══ */}
      <section className="solution-module solution-module-alt" style={{ paddingBottom: '80px' }}>
        <Reveal>
          <div className="solutions-cta-card">
            <h2>See It All In Action</h2>
            <p>Create your free account to unlock the full platform experience.</p>
            <div style={{ display: 'flex', flexWrap: 'wrap', justifyContent: 'center', gap: '16px' }}>
              <Link to="/signup" className="solutions-cta-btn">Start Free Trial →</Link>
              <Link to="/contact" className="solutions-cta-btn" style={{ background: 'white', color: 'var(--color-brand-navy)', border: '2px solid var(--color-ui-border)', boxShadow: 'none' }}>
                Schedule Demo
              </Link>
            </div>
          </div>
        </Reveal>
      </section>
    </>
  );
}

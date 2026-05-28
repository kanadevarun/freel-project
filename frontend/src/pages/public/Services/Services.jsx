import { useEffect, useRef } from 'react';
import { Link } from 'react-router-dom';
import './Services.css';

/* ─── Scroll Reveal Hook ─── */
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

/* ─── Reusable Components ─── */

function FeatureList({ items }) {
  return (
    <ul className="service-mode-features">
      {items.map((item, i) => (
        <li key={i}>
          <span className="feature-check">✓</span>
          {item}
        </li>
      ))}
    </ul>
  );
}

function StatsCard({ icon, stats, colorClass = 'stat-value-teal' }) {
  return (
    <div className="stats-card">
      <div className="stats-card-icon">{icon}</div>
      <div className="stats-grid">
        {stats.map((s, i) => (
          <div key={i}>
            <div className={`stat-value ${colorClass}`}>{s.value}</div>
            <div className="stat-label">{s.label}</div>
          </div>
        ))}
      </div>
    </div>
  );
}

/* ═══════════════════════════════════════════ */
/*          SERVICES LANDING PAGE             */
/* ═══════════════════════════════════════════ */
export default function Services() {
  return (
    <>
      {/* ═══ HERO ═══ */}
      <section className="services-hero">
        <div style={{ position: 'relative', zIndex: 1 }}>
          <h1 className="services-hero-title">
            Freight Services for <br className="hidden md:block" />
            <span className="text-gradient bg-gradient-to-r from-brand-teal to-brand-indigo">Every Mode</span>
          </h1>
          <p className="services-hero-sub">
            From air cargo to ocean containers to road transport — one platform, total coverage.
          </p>
        </div>
      </section>

      {/* ═══ AIR FREIGHT ═══ */}
      <section className="service-mode-section">
        <div className="service-mode-grid">
          <Reveal>
            <div className="service-mode-text">
              <div className="service-mode-icon">✈️</div>
              <h2 className="service-mode-title">Air Freight</h2>
              <p className="service-mode-desc">
                Global air cargo services via IATA-certified agents. Ship general, hazardous, pharma, and perishable
                cargo to 150+ airports worldwide with express delivery options.
              </p>
              <FeatureList items={[
                'General & Dangerous Goods',
                'Pharma & Temperature Controlled',
                'Express & Charter Options',
                'FlightAware Live Tracking'
              ]} />
              <Link to="/services/air-freight" className="service-cta-btn cta-air">
                View Air Freight Details →
              </Link>
            </div>
          </Reveal>
          <Reveal delay="reveal-delay-1">
            <div className="stats-card-wrap">
              <StatsCard
                icon="✈️"
                colorClass="stat-value-teal"
                stats={[
                  { value: '150+', label: 'Airports' },
                  { value: '24hr', label: 'Express' },
                  { value: 'DG', label: 'Certified' },
                  { value: 'Live', label: 'Tracking' },
                ]}
              />
            </div>
          </Reveal>
        </div>
      </section>

      {/* ═══ SEA FREIGHT (Reversed) ═══ */}
      <section className="service-mode-section" style={{ borderTop: 'none' }}>
        <div className="service-mode-grid service-mode-reversed">
          <Reveal delay="reveal-delay-1">
            <div className="stats-card-wrap">
              <StatsCard
                icon="🚢"
                colorClass="stat-value-indigo"
                stats={[
                  { value: '50+', label: 'Shipping Lines' },
                  { value: 'FCL/LCL', label: 'Load Types' },
                  { value: 'HAZ', label: 'Certified' },
                  { value: 'AIS', label: 'Vessel Tracking' },
                ]}
              />
            </div>
          </Reveal>
          <Reveal>
            <div className="service-mode-text">
              <div className="service-mode-icon">🚢</div>
              <h2 className="service-mode-title">Sea Freight</h2>
              <p className="service-mode-desc">
                FCL & LCL shipments via world's top shipping lines. 20ft, 40ft, High Cube, and Reefer containers
                with full hazardous cargo support.
              </p>
              <FeatureList items={[
                'Maersk, MSC, CMA CGM & more',
                '20ft / 40ft / Reefer Containers',
                'HAZ & Non-HAZ Cargo',
                'AIS / MarineTraffic Tracking'
              ]} />
              <Link to="/services/sea-freight" className="service-cta-btn cta-sea">
                View Sea Freight Details →
              </Link>
            </div>
          </Reveal>
        </div>
      </section>

      {/* ═══ ROAD TRANSPORT ═══ */}
      <section className="service-mode-section" style={{ borderTop: 'none' }}>
        <div className="service-mode-grid">
          <Reveal>
            <div className="service-mode-text">
              <div className="service-mode-icon">🚛</div>
              <h2 className="service-mode-title">Road Transport</h2>
              <p className="service-mode-desc">
                FTL & LTL services across India with 500+ verified transporters. GPS-tracked vehicles with real-time
                ETA updates and digital LR generation.
              </p>
              <FeatureList items={[
                'FTL, LTL, ODC & Express',
                '500+ GST-Verified Transporters',
                'GPS Live Tracking',
                'Digital LR with QR Code'
              ]} />
              <Link to="/services/road-transport" className="service-cta-btn cta-road">
                View Road Transport Details →
              </Link>
            </div>
          </Reveal>
          <Reveal delay="reveal-delay-1">
            <div className="stats-card-wrap">
              <StatsCard
                icon="🚛"
                colorClass="stat-value-teal"
                stats={[
                  { value: '500+', label: 'Transporters' },
                  { value: 'Pan India', label: 'Coverage' },
                  { value: 'GPS', label: 'Real-time' },
                  { value: 'Digital', label: 'LR + QR' },
                ]}
              />
            </div>
          </Reveal>
        </div>
      </section>

      {/* ═══ CUSTOMS BROKERAGE ═══ */}
      <section className="customs-section">
        <div className="customs-header">
          <Reveal>
            <div className="customs-icon">📜</div>
            <h2 className="customs-title">Customs Brokerage & Compliance</h2>
            <p className="customs-desc">
              End-to-end customs clearance with built-in HSN code lookup, MSDS management, and automated compliance
              checks. Never face a port delay again.
            </p>
          </Reveal>
        </div>
        <div className="customs-features-grid">
          {[
            { icon: '🔍', title: 'HSN Lookup', badge: '5000+ codes mapped' },
            { icon: '📋', title: 'MSDS Manager', badge: 'HAZ classification' },
            { icon: '📁', title: 'Document Vault', badge: 'Secure legal storage' },
          ].map((f, i) => (
            <Reveal key={i} delay={`reveal-delay-${i + 1}`}>
              <div className="customs-feature-card">
                <div className="customs-feature-icon">{f.icon}</div>
                <h3 className="customs-feature-title">{f.title}</h3>
                <span className="customs-feature-badge">{f.badge}</span>
              </div>
            </Reveal>
          ))}
        </div>
        <Reveal delay="reveal-delay-4" className="text-center mt-12">
          <Link to="/services/customs" className="service-cta-btn cta-sea" style={{ background: 'linear-gradient(135deg, var(--color-brand-indigo), var(--color-brand-teal))', marginTop: '40px' }}>
            View Customs & Compliance Details →
          </Link>
        </Reveal>
      </section>

      {/* ═══ CTA BANNER ═══ */}
      <section className="services-cta-banner">
        <Reveal>
          <div style={{ maxWidth: '700px', margin: '0 auto' }}>
            <h2>Start Shipping Smarter Today</h2>
            <p>Get instant access to all services with your free Freel account.</p>
            <Link to="/signup" className="services-cta-banner-btn">
              Start Free Trial →
            </Link>
          </div>
        </Reveal>
      </section>
    </>
  );
}

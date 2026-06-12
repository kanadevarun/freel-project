import { useEffect, useRef, useState } from 'react';
import { Link } from 'react-router-dom';
import './SeaFreight.css';

/* ─── Scroll Reveal ─── */
function useReveal() {
  const ref = useRef(null);
  useEffect(() => {
    const el = ref.current;
    if (!el) return;
    const obs = new IntersectionObserver(
      ([e]) => { if (e.isIntersecting) { el.classList.add('opacity-100', 'translate-y-0'); el.classList.remove('opacity-0', 'translate-y-8'); obs.unobserve(el); } },
      { threshold: 0.15, rootMargin: '0px 0px -50px 0px' }
    );
    obs.observe(el);
    return () => obs.disconnect();
  }, []);
  return ref;
}

function Reveal({ children, className = '', delay = '' }) {
  const ref = useReveal();
  return <div ref={ref} className={`transition-all duration-1000 opacity-0 translate-y-8 ${delay} ${className}`}>{children}</div>;
}

function useRevealRight() {
  const ref = useRef(null);
  useEffect(() => {
    const el = ref.current;
    if (!el) return;
    const obs = new IntersectionObserver(
      ([e]) => { if (e.isIntersecting) { el.classList.add('opacity-100', 'translate-x-0'); el.classList.remove('opacity-0', 'translate-x-16'); obs.unobserve(el); } },
      { threshold: 0.15, rootMargin: '0px 0px -50px 0px' }
    );
    obs.observe(el);
    return () => obs.disconnect();
  }, []);
  return ref;
}

function RevealRight({ children, className = '', delay = '' }) {
  const ref = useRevealRight();
  return <div ref={ref} className={`transition-all duration-1000 opacity-0 translate-x-16 ${delay} ${className}`}>{children}</div>;
}

const stats = [
  { value: '50+', label: 'Shipping Lines' },
  { value: '150+', label: 'Global Ports' },
  { value: 'FCL/LCL', label: 'Load Types' },
  { value: 'AIS', label: 'Vessel Tracking' },
];

const containers = [
  { icon: '📦', title: '20ft / 40ft Standard', desc: 'Ideal for heavy or volume cargo. Electronics, textiles, and consumer goods.', color: 'from-blue-400 to-blue-600' },
  { icon: '❄️', title: 'Reefer Container', desc: 'Active temp control (-25°C to +25°C) for pharma, food, and chemicals.', color: 'from-cyan-400 to-cyan-600' },
  { icon: '🔓', title: 'Open Top & Flat Rack', desc: 'Oversized cargo loaded from top/sides. Heavy machinery & construction.', color: 'from-amber-400 to-amber-600' },
  { icon: '⚠️', title: 'HAZ Cargo', desc: 'IMO Class 1-9 dangerous goods with full IMDG compliance and segregation.', color: 'from-red-400 to-red-600' },
];

const shippingLines = ['Maersk', 'MSC', 'CMA CGM', 'Hapag-Lloyd', 'ONE', 'Evergreen', 'COSCO', 'Yang Ming'];

/* ═══════════════════════════════════════════ */
/*             SEA FREIGHT PAGE               */
/* ═══════════════════════════════════════════ */
export default function SeaFreight() {
  const [activeVessel, setActiveVessel] = useState(0);
  const [activeDigitalEra, setActiveDigitalEra] = useState('phone');
  const digitalSectionRef = useRef(null);
  
  // IntersectionObserver for era card scroll-driven badges
  useEffect(() => {
    const section = digitalSectionRef.current;
    if (!section) return;
    const cards = section.querySelectorAll('.sf-era-card');
    if (!cards.length) return;

    const obs = new IntersectionObserver(
      (entries) => {
        entries.forEach((entry) => {
          if (entry.isIntersecting) {
            const era = entry.target.getAttribute('data-era');
            if (era) setActiveDigitalEra(era);
          }
        });
      },
      { threshold: 0.6, rootMargin: '-20% 0px -20% 0px' }
    );

    cards.forEach((card) => obs.observe(card));
    return () => obs.disconnect();
  }, []);

  const vessels = [
    { name: 'MSC ANNA', route: 'Nhava Sheva → Rotterdam', status: 'In Transit', progress: '45%' },
    { name: 'MAERSK MC-KINNEY', route: 'Mundra → Jebel Ali', status: 'Docking', progress: '95%' },
    { name: 'CMA CGM JACQUES', route: 'Chennai → Singapore', status: 'Departed', progress: '15%' },
  ];

  return (
    <div className="bg-white min-h-screen">
      
      {/* ═══ CINEMATIC HERO — 100vh ═══ */}
      <section className="sf-hero" id="sf-hero">
        {/* Background Video */}
        <video
          className="sf-hero__video"
          autoPlay
          muted
          loop
          playsInline
        >
          <source
            src="/videos/sea-freight/Container_ship_moving_through_ocean_202606071558.mp4"
            type="video/mp4"
          />
        </video>

        {/* Cinematic Overlay */}
        <div className="sf-hero__overlay"></div>

        {/* Content Container */}
        <div className="sf-hero__content">
          {/* Eyebrow Label */}
          <div className="sf-eyebrow">
            <span className="sf-eyebrow-icon">🌊</span>
            SEA FREIGHT NETWORK
          </div>

          {/* Headline */}
          <h1 className="sf-headline sf-anim-headline">
            The <span className="sf-headline__highlight">Ocean</span> Moves<br />
            The World.
          </h1>

          {/* Description */}
          <p className="sf-description sf-anim-description">
            For centuries, ships connected civilizations.<br />
            Today they power global commerce, moving over<br />
            90% of the world's trade across oceans.
          </p>

          {/* CTA Buttons */}
          <div className="sf-ctas sf-anim-ctas">
            <Link to="/contact" className="sf-btn-primary">
              Start Shipping &rarr;
            </Link>
            <a href="#ais-tracking" className="sf-btn-secondary">
              Explore Global Routes
            </a>
          </div>
        </div>

        {/* Statistics Bar */}
        <div className="sf-stats-bar sf-anim-stats">
          <div className="sf-stat">
            <div className="sf-stat-value">90%</div>
            <div className="sf-stat-label">Global Trade</div>
          </div>
          <div className="sf-stat">
            <div className="sf-stat-value">150+</div>
            <div className="sf-stat-label">Ports Connected</div>
          </div>
          <div className="sf-stat">
            <div className="sf-stat-value">50+</div>
            <div className="sf-stat-label">Carrier Partners</div>
          </div>
          <div className="sf-stat">
            <div className="sf-stat-value">24/7</div>
            <div className="sf-stat-label">Cargo Visibility</div>
          </div>
        </div>
      </section>

      {/* ═══ ORIGINS OF GLOBAL TRADE ═══ */}
      <section className="sf-origins">
        <div className="sf-origins__container">
          
          {/* LEFT SIDE - IMAGE */}
          <Reveal className="sf-origins__left">
            <div className="sf-origins__glow"></div>
            <div className="sf-origins__image-wrapper">
              <img 
                src="/images/sea-freight/Human_figures_carrying_goods_harbor_202606071550.jpeg" 
                alt="Origins of Trade Harbor" 
                className="sf-origins__image" 
              />
            </div>
          </Reveal>

          {/* RIGHT SIDE - TEXT */}
          <Reveal delay="delay-200" className="sf-origins__right">
            <div className="sf-origins__bg-number">5000 BC</div>
            
            <div className="sf-origins__content">
              <div className="sf-origins__eyebrow">
                🌍 ORIGINS OF GLOBAL TRADE
              </div>
              <h2 className="sf-origins__headline">
                Before Supply Chains,<br/>
                There Were Harbors.
              </h2>
              
              {/* TIMELINE */}
              <div className="sf-origins__timeline-horizontal">
                <div className="sf-timeline-hz-item sf-timeline-hz-active">
                  <div className="sf-timeline-hz-dot"></div>
                  <div className="sf-timeline-hz-text">5000 BC</div>
                  <div className="sf-timeline-hz-label">Ancient Ports</div>
                </div>
                <div className="sf-timeline-hz-line"></div>
                <div className="sf-timeline-hz-item">
                  <div className="sf-timeline-hz-dot"></div>
                  <div className="sf-timeline-hz-text">1000 BC</div>
                  <div className="sf-timeline-hz-label">Trade</div>
                </div>
                <div className="sf-timeline-hz-line"></div>
                <div className="sf-timeline-hz-item">
                  <div className="sf-timeline-hz-dot"></div>
                  <div className="sf-timeline-hz-text">1400 AD</div>
                  <div className="sf-timeline-hz-label">Routes</div>
                </div>
                <div className="sf-timeline-hz-line"></div>
                <div className="sf-timeline-hz-item">
                  <div className="sf-timeline-hz-dot"></div>
                  <div className="sf-timeline-hz-text">1800 AD</div>
                  <div className="sf-timeline-hz-label">Steamships</div>
                </div>
              </div>

              <div className="sf-origins__description">
                <p>
                  Trade began at the water's edge.
                </p>
                <p>
                  Ports connected distant civilizations.
                </p>
                <p>
                  Every modern supply chain traces its roots back to these harbors.
                </p>
              </div>

              {/* STATISTIC BLOCK */}
              <div className="sf-origins__stat-block">
                <div className="sf-origins__stat-value">5000+</div>
                <div className="sf-origins__stat-label">Years of Maritime Commerce</div>
              </div>

            </div>
          </Reveal>
        </div>
      </section>

      {/* ═══ THE WORLD GOT SMALLER (INDUSTRIAL REVOLUTION) ═══ */}
      <section className="sf-industrial">
        <div className="sf-industrial__container">
          
          {/* LEFT SIDE - TEXT */}
          <Reveal className="sf-industrial__left">
            <div className="sf-industrial__bg-number">1800</div>
            
            <div className="sf-industrial__content">
              <div className="sf-industrial__eyebrow">
                ⚙ THE MACHINE AGE
              </div>
              <h2 className="sf-industrial__headline">
                The World Got Smaller.
              </h2>
              
              {/* TRUE HORIZONTAL TIMELINE */}
              <div className="sf-industrial__timeline-horizontal">
                <div className="sf-timeline-hz-item">
                  <div className="sf-timeline-hz-dot"></div>
                  <div className="sf-timeline-hz-text">1400s</div>
                  <div className="sf-timeline-hz-label">Harbors</div>
                </div>
                <div className="sf-timeline-hz-line"></div>
                <div className="sf-timeline-hz-item sf-timeline-hz-active">
                  <div className="sf-timeline-hz-dot"></div>
                  <div className="sf-timeline-hz-text">1800s</div>
                  <div className="sf-timeline-hz-label">Steam</div>
                </div>
                <div className="sf-timeline-hz-line"></div>
                <div className="sf-timeline-hz-item">
                  <div className="sf-timeline-hz-dot"></div>
                  <div className="sf-timeline-hz-text">1900s</div>
                  <div className="sf-timeline-hz-label">Rail</div>
                </div>
                <div className="sf-timeline-hz-line"></div>
                <div className="sf-timeline-hz-item">
                  <div className="sf-timeline-hz-dot"></div>
                  <div className="sf-timeline-hz-text">1950s</div>
                  <div className="sf-timeline-hz-label">Containers</div>
                </div>
              </div>

              <div className="sf-industrial__description">
                <p>
                  Steam replaced wind. Rail conquered land.
                </p>
                <p>
                  Distance shattered. Cargo moved faster than ever before.
                </p>
                <p>
                  The foundation of global trade was forged in iron.
                </p>
              </div>
            </div>
          </Reveal>

          {/* RIGHT SIDE - IMAGE */}
          <RevealRight delay="delay-200" className="sf-industrial__right">
            <div className="sf-industrial__glow"></div>
            <div className="sf-industrial__image-wrapper">
              <img 
                src="/images/sea-freight/Figures_coordinating_freight_mov…_202606071550.jpeg" 
                alt="Industrial Revolution Trade" 
                className="sf-industrial__image" 
              />
            </div>
          </RevealRight>
        </div>
      </section>

      {/* ═══ THE BOX THAT CHANGED THE WORLD (CONTAINERS) ═══ */}
      <section className="sf-containers">
        <div className="sf-containers__container">
          
          {/* LEFT SIDE - TEXT */}
          <Reveal className="sf-containers__left">
            <div className="sf-containers__bg-number">1956</div>
            
            <div className="sf-containers__content">
              <div className="sf-containers__eyebrow">
                📦 CONTAINER REVOLUTION • 1956
              </div>
              <h2 className="sf-containers__headline">
                The Box That Changed<br/>
                The World.
              </h2>

              <div className="sf-containers__quote">
                "One standardized container changed global trade forever."
              </div>
              
              {/* TRUE HORIZONTAL TIMELINE */}
              <div className="sf-containers__timeline-horizontal">
                <div className="sf-timeline-hz-item">
                  <div className="sf-timeline-hz-dot"></div>
                  <div className="sf-timeline-hz-text">1400s</div>
                  <div className="sf-timeline-hz-label">Harbors</div>
                </div>
                <div className="sf-timeline-hz-line"></div>
                <div className="sf-timeline-hz-item">
                  <div className="sf-timeline-hz-dot"></div>
                  <div className="sf-timeline-hz-text">1800s</div>
                  <div className="sf-timeline-hz-label">Steamships</div>
                </div>
                <div className="sf-timeline-hz-line"></div>
                <div className="sf-timeline-hz-item">
                  <div className="sf-timeline-hz-dot"></div>
                  <div className="sf-timeline-hz-text">1900s</div>
                  <div className="sf-timeline-hz-label">Rail Networks</div>
                </div>
                <div className="sf-timeline-hz-line"></div>
                <div className="sf-timeline-hz-item sf-timeline-hz-active">
                  <div className="sf-timeline-hz-dot"></div>
                  <div className="sf-timeline-hz-text">1956</div>
                  <div className="sf-timeline-hz-label">Containers</div>
                </div>
              </div>

              <div className="sf-containers__transformations">
                <div className="sf-transform-card">
                  <div className="sf-transform-label">Before</div>
                  <div className="sf-transform-text">Cargo loaded manually</div>
                </div>
                <div className="sf-transform-card">
                  <div className="sf-transform-label">After</div>
                  <div className="sf-transform-text">Standardized containers</div>
                </div>
                <div className="sf-transform-card sf-transform-result">
                  <div className="sf-transform-label">Result</div>
                  <div className="sf-transform-text">Global supply chains</div>
                </div>
              </div>

              {/* STATISTIC CARDS */}
              <div className="sf-containers__stats-row">
                <div className="sf-containers__stat-card">
                  <div className="sf-containers__stat-value">36x</div>
                  <div className="sf-containers__stat-label">Faster Port Operations</div>
                </div>
                <div className="sf-containers__stat-card">
                  <div className="sf-containers__stat-value">90%</div>
                  <div className="sf-containers__stat-label">Reduction In Cargo Handling Costs</div>
                </div>
                <div className="sf-containers__stat-card sf-stat-full">
                  <div className="sf-containers__stat-value">20M+</div>
                  <div className="sf-containers__stat-label">Containers Moved Annually Across Major Ports</div>
                </div>
              </div>

            </div>
          </Reveal>

          {/* RIGHT SIDE - IMAGE */}
          <RevealRight delay="delay-200" className="sf-containers__right">
            <div className="sf-containers__image-sticky">
              <div className="sf-containers__glow"></div>
              <div className="sf-containers__image-wrapper">
                <img 
                  src="/images/sea-freight/Massive_container_terminal_viewed_from_202606052225.jpeg" 
                  alt="Massive Container Terminal" 
                  className="sf-containers__image" 
                />
              </div>
            </div>
          </RevealRight>
        </div>
      </section>

      {/* ═══ THE AGE OF GLOBALIZATION ═══ */}
      <section className="sf-globalization">
        <div className="sf-globalization__container">
          
          {/* LEFT SIDE - IMAGE */}
          <Reveal className="sf-globalization__left">
            <div className="sf-globalization__image-sticky">
              <div className="sf-globalization__glow"></div>
              <div className="sf-globalization__image-wrapper">
                <img 
                  src="/images/sea-freight/Human_figures_in_automated_wareh…_202606071550.jpeg" 
                  alt="Digital Transformation" 
                  className="sf-globalization__image" 
                />
                {/* CSS animated pulses for trade routes */}
                <div className="sf-pulse-node" style={{top: '35%', left: '48%', animationDelay: '0s'}}></div>
                <div className="sf-pulse-node" style={{top: '42%', left: '55%', animationDelay: '0.5s'}}></div>
                <div className="sf-pulse-node" style={{top: '60%', left: '78%', animationDelay: '1.2s'}}></div>
                <div className="sf-pulse-node" style={{top: '48%', left: '25%', animationDelay: '0.8s'}}></div>
                <div className="sf-pulse-node" style={{top: '28%', left: '68%', animationDelay: '1.5s'}}></div>
              </div>
            </div>
          </Reveal>

          {/* RIGHT SIDE - TEXT */}
          <RevealRight delay="delay-200" className="sf-globalization__right">
            <div className="sf-globalization__bg-number">2000</div>
            
            <div className="sf-globalization__content">
              <div className="sf-globalization__eyebrow">
                🌍 THE AGE OF GLOBALIZATION
              </div>
              <h2 className="sf-globalization__headline">
                Factories Moved.<br/>
                Trade Exploded.
              </h2>
              
              <div className="sf-globalization__subheadline">
                Containerization connected factories, suppliers and customers across continents.<br/><br/>
                For the first time, products could be built anywhere and sold everywhere.
              </div>

              {/* TIMELINE */}
              <div className="sf-globalization__timeline-horizontal sf-timeline-hz-compact">
                <div className="sf-timeline-hz-item">
                  <div className="sf-timeline-hz-dot"></div>
                  <div className="sf-timeline-hz-text">1956</div>
                  <div className="sf-timeline-hz-label">Containers</div>
                </div>
                <div className="sf-timeline-hz-line"></div>
                <div className="sf-timeline-hz-item">
                  <div className="sf-timeline-hz-dot"></div>
                  <div className="sf-timeline-hz-text">1980s</div>
                  <div className="sf-timeline-hz-label">Manufacturing</div>
                </div>
                <div className="sf-timeline-hz-line"></div>
                <div className="sf-timeline-hz-item">
                  <div className="sf-timeline-hz-dot"></div>
                  <div className="sf-timeline-hz-text">1990s</div>
                  <div className="sf-timeline-hz-label">Expansion</div>
                </div>
                <div className="sf-timeline-hz-line"></div>
                <div className="sf-timeline-hz-item">
                  <div className="sf-timeline-hz-dot"></div>
                  <div className="sf-timeline-hz-text">2000s</div>
                  <div className="sf-timeline-hz-label">Hyper Global</div>
                </div>
                <div className="sf-timeline-hz-line"></div>
                <div className="sf-timeline-hz-item sf-timeline-hz-active" style={{ '--active-color': '#0EA5E9' }}>
                  <div className="sf-timeline-hz-dot"></div>
                  <div className="sf-timeline-hz-text">Today</div>
                  <div className="sf-timeline-hz-label">Connected</div>
                </div>
              </div>

              {/* TRANSFORMATION FLOW */}
              <div className="sf-globalization__transformation-flow">
                <div className="sf-transform-card sf-transform-step">
                  <span className="sf-transform-label">LOCAL PRODUCTION</span>
                  <span className="sf-transform-text">Products were built and sold within the same region.</span>
                </div>
                
                <div className="sf-transform-arrow">↓</div>
                
                <div className="sf-transform-card sf-transform-step">
                  <span className="sf-transform-label" style={{ color: '#0EA5E9' }}>GLOBAL SUPPLY CHAINS</span>
                  <span className="sf-transform-text">Manufacturing, assembly and distribution spread across multiple countries.</span>
                </div>
                
                <div className="sf-transform-arrow" style={{ color: '#0EA5E9' }}>↓</div>
                
                <div className="sf-transform-card sf-transform-result" style={{ background: 'rgba(14, 165, 233, 0.05)', borderLeftColor: '#0EA5E9' }}>
                  <span className="sf-transform-label" style={{ color: '#0EA5E9' }}>CONNECTED WORLD</span>
                  <span className="sf-transform-text">One product can cross oceans multiple times before reaching a customer.</span>
                </div>
              </div>

              {/* STATISTICS */}
              <div className="sf-globalization__stats-row">
                <div className="sf-globalization__stat-card">
                  <div className="sf-globalization__stat-value">70%+</div>
                  <div className="sf-globalization__stat-label">Maritime Trade</div>
                </div>
                <div className="sf-globalization__stat-card">
                  <div className="sf-globalization__stat-value">100+</div>
                  <div className="sf-globalization__stat-label">Connected Countries</div>
                </div>
                <div className="sf-globalization__stat-card">
                  <div className="sf-globalization__stat-value">Millions</div>
                  <div className="sf-globalization__stat-label">Containers Weekly</div>
                </div>
              </div>

            </div>
          </RevealRight>
        </div>
      </section>

      {/* ═══ THE DIGITAL ERA ═══ */}
      <section className="sf-digital" ref={digitalSectionRef}>
        <div className="sf-digital__container">
          
          {/* LEFT SIDE - TEXT */}
          <Reveal className="sf-digital__left">
            <div className="sf-digital__bg-number">2000</div>
            
            <div className="sf-digital__content">
              <div className="sf-digital__eyebrow">
                🛰 DIGITAL TRANSFORMATION • 2000s → TODAY
              </div>
              <h2 className="sf-digital__headline">
                Information Started Moving<br/>Faster Than Cargo.
              </h2>
              
              <div className="sf-digital__quote">
                "For the first time in history, companies could track cargo while it was still moving."
              </div>

              {/* ERA CARDS EVOLUTION */}
              <div className="sf-digital__era-cards">
                
                {/* PHONE ERA */}
                <div className="sf-era-card sf-era-gray" data-era="phone">
                  <div className="sf-era-card__icon">📞</div>
                  <div className="sf-era-card__body">
                    <div className="sf-era-card__title">Phone Era</div>
                    <div className="sf-era-card__desc">Companies waited for shipment updates. Visibility was extremely limited.</div>
                  </div>
                </div>

                <div className="sf-era-connector"></div>

                {/* GPS ERA */}
                <div className="sf-era-card sf-era-blue" data-era="gps">
                  <div className="sf-era-card__icon">📡</div>
                  <div className="sf-era-card__body">
                    <div className="sf-era-card__title">GPS Era</div>
                    <div className="sf-era-card__desc">Real-time location became available. Cargo could finally be tracked.</div>
                  </div>
                </div>

                <div className="sf-era-connector"></div>

                {/* CLOUD ERA */}
                <div className="sf-era-card sf-era-cyan" data-era="cloud">
                  <div className="sf-era-card__icon">☁️</div>
                  <div className="sf-era-card__body">
                    <div className="sf-era-card__title">Cloud Era</div>
                    <div className="sf-era-card__desc">Platforms connected suppliers, carriers and customers. Data flowed across the supply chain.</div>
                  </div>
                </div>

                <div className="sf-era-connector"></div>

                {/* AI ERA */}
                <div className="sf-era-card sf-era-card--active sf-era-green" data-era="ai">
                  <div className="sf-era-card__icon">🤖</div>
                  <div className="sf-era-card__body">
                    <div className="sf-era-card__title">AI Era</div>
                    <div className="sf-era-card__desc">Predictive visibility. Forecast delays before they happen.</div>
                  </div>
                </div>
                
              </div>

              {/* STATISTICS */}
              <div className="sf-digital__stats-row">
                <div className="sf-digital__stat-card">
                  <div className="sf-digital__stat-value">95%</div>
                  <div className="sf-digital__stat-label">Visibility</div>
                </div>
                <div className="sf-digital__stat-card">
                  <div className="sf-digital__stat-value">24/7</div>
                  <div className="sf-digital__stat-label">Monitoring</div>
                </div>
              </div>

            </div>
          </Reveal>

          {/* RIGHT SIDE - IMAGE */}
          <RevealRight delay="delay-200" className="sf-digital__right">
            <div className="sf-digital__image-sticky">
              <div className="sf-digital__glow"></div>
              <div className="sf-digital__image-wrapper">
                <img 
                  src="/images/sea-freight/sea-digital-era.png" 
                  alt="Digital Era Earth" 
                  className="sf-digital__image" 
                />
                
                {/* SCROLL-DRIVEN FLOATING BADGES */}
                <div className={`sf-digital__badge sf-badge-phone ${activeDigitalEra === 'phone' ? 'sf-badge--visible' : ''}`}>
                  <div className="sf-overlay-dot sf-pulse-red"></div> Status Unknown
                </div>
                <div className={`sf-digital__badge sf-badge-gps ${activeDigitalEra === 'gps' ? 'sf-badge--visible' : ''}`}>
                  <div className="sf-overlay-dot sf-pulse-green"></div> Live Tracking
                </div>
                <div className={`sf-digital__badge sf-badge-cloud ${activeDigitalEra === 'cloud' ? 'sf-badge--visible' : ''}`}>
                  <div className="sf-overlay-dot sf-pulse-blue"></div> Connected
                </div>
                <div className={`sf-digital__badge sf-badge-ai ${activeDigitalEra === 'ai' ? 'sf-badge--visible' : ''}`}>
                  <div className="sf-overlay-dot sf-pulse-green"></div> ETA Accuracy 98%
                </div>
              </div>
            </div>
          </RevealRight>
        </div>
      </section>

      {/* ═══ PORT OPERATIONS & CONTAINERS ═══ */}
      <section className="py-32 bg-white relative">
        <div className="max-w-7xl mx-auto px-4 sm:px-6">
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-16 items-center">
            <Reveal>
              <div className="relative rounded-[3rem] overflow-hidden shadow-2xl group">
                <div className="absolute inset-0 bg-brand-teal/20 group-hover:bg-transparent transition-colors z-10 mix-blend-overlay duration-700"></div>
                <img src="/assets/port_operations.png" alt="Busy Container Port" className="w-full h-auto object-cover transform group-hover:scale-105 transition-transform duration-1000" />
                <div className="absolute bottom-0 left-0 w-full p-8 bg-gradient-to-t from-[#0a192f] to-transparent z-20">
                  <div className="text-cyan-400 font-mono text-sm mb-1">LIVE FEED: TERMINAL 4</div>
                  <div className="text-white font-bold text-xl">High-Volume Container Handling</div>
                </div>
              </div>
            </Reveal>

            <div>
              <Reveal>
                <span className="text-cyan-600 font-bold tracking-widest uppercase text-sm mb-3 block">Cargo Capabilities</span>
                <h2 className="text-4xl md:text-5xl font-extrabold text-brand-navy mb-6 leading-tight">The Right Container <br/>For Every Cargo.</h2>
                <p className="text-slate-600 text-lg mb-12">
                  Whether you're shipping garments, frozen seafood, or oversized industrial turbines, we have the specialized equipment ready at origin.
                </p>
              </Reveal>

              <div className="grid grid-cols-1 sm:grid-cols-2 gap-6">
                {containers.map((c, i) => (
                  <Reveal key={i} delay={`delay-${(i % 2) * 100}`}>
                    <div className="bg-slate-50 rounded-3xl p-6 border border-slate-100 hover:border-cyan-200 hover:shadow-lg transition-all group h-full">
                      <div className={`w-12 h-12 rounded-xl bg-gradient-to-br ${c.color} flex items-center justify-center text-2xl text-white shadow-md mb-4 group-hover:scale-110 transition-transform`}>
                        {c.icon}
                      </div>
                      <h3 className="text-lg font-bold text-brand-navy mb-2">{c.title}</h3>
                      <p className="text-slate-500 text-sm leading-relaxed">{c.desc}</p>
                    </div>
                  </Reveal>
                ))}
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* ═══ THE GRAND FINALE ═══ */}
      <section className="sf-grand-finale">
        <div className="sf-grand-finale__video-wrapper">
          <video
            className="sf-grand-finale__video"
            autoPlay
            muted
            loop
            playsInline
          >
            <source
              src="/videos/sea-freight/Robot_observing_global_shipping_…_202606071603.mp4"
              type="video/mp4"
            />
          </video>
          <div className="sf-grand-finale__overlay"></div>
        </div>

        {/* FLOATING TAGS OVER VIDEO */}
        <div className="sf-grand-finale__badges">
          <div className="sf-grand-finale__badge sf-badge-pos-1">
            <div className="sf-overlay-dot sf-pulse-green"></div> ✓ Delay Predicted
          </div>
          <div className="sf-grand-finale__badge sf-badge-pos-5">
            <div className="sf-overlay-dot sf-pulse-green"></div> ✓ Autonomous Planning
          </div>
        </div>

        {/* CENTERED TYPOGRAPHY */}
        <Reveal className="sf-grand-finale__content">
          <h2 className="sf-grand-finale__headline">
            <span className="sf-grand-finale__headline-from">From Moving Cargo.</span>
            <br/>
            <span className="sf-grand-finale__headline-to">To Moving Intelligence.</span>
          </h2>
          
          <p className="sf-grand-finale__desc">
            One Network. Infinite Possibilities.
          </p>

          <div className="sf-grand-finale__actions">
            <Link to="/contact" className="sf-grand-finale__btn sf-grand-finale__btn--primary">Get Quote</Link>
            <Link to="/demo" className="sf-grand-finale__btn sf-grand-finale__btn--secondary">Book Demo</Link>
          </div>
        </Reveal>

        {/* Cinematic bottom gradient fade into footer */}
        <div className="sf-grand-finale__bottom-fade"></div>
      </section>

    </div>
  );
}

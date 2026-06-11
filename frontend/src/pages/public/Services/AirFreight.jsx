import { useEffect, useRef, useState } from 'react';
import { Link } from 'react-router-dom';
import './AirFreight.css';

/* ═══════════════════════════════════════════════════════════
   AIR FREIGHT — CINEMATIC HERO
   "The opening scene of a Netflix documentary about
    how global air cargo powers world commerce."
   ═══════════════════════════════════════════════════════════ */

/* ─── Ambient Route Lines SVG ─── */
function AmbientRoutes() {
  return (
    <div className="af-hero__ambient">
      <svg viewBox="0 0 1440 900" preserveAspectRatio="xMidYMid slice">
        {/* Route arcs */}
        <path
          className="af-route-line"
          d="M-50,600 Q400,200 750,350 T1500,250"
        />
        <path
          className="af-route-line af-route-line--2"
          d="M-100,300 Q300,700 700,500 T1550,600"
        />
        <path
          className="af-route-line af-route-line--3"
          d="M200,850 Q600,100 1000,400 T1600,150"
        />

        {/* Glowing waypoint dots */}
        <circle className="af-route-dot" cx="400" cy="320" r="3" />
        <circle className="af-route-dot af-route-dot--2" cx="750" cy="350" r="3" />
        <circle className="af-route-dot af-route-dot--3" cx="1050" cy="290" r="3" />
        <circle className="af-route-dot" cx="300" cy="550" r="3" />
        <circle className="af-route-dot af-route-dot--2" cx="850" cy="480" r="3" />
      </svg>
    </div>
  );
}

/* ═══════════════════════════════════════════════════════════ */
/*             COUNT UP HELPER                               */
/* ═══════════════════════════════════════════════════════════ */
const CountUp = ({ end, duration = 2000, decimals = 0 }) => {
  const [count, setCount] = useState(0);
  const ref = useRef(null);

  useEffect(() => {
    let startTimestamp = null;
    let observer;
    let hasAnimated = false;
    let animationFrameId;

    const step = (timestamp) => {
      if (!startTimestamp) startTimestamp = timestamp;
      const progress = Math.min((timestamp - startTimestamp) / duration, 1);
      const easeProgress = progress === 1 ? 1 : 1 - Math.pow(2, -10 * progress);
      setCount(easeProgress * end);
      if (progress < 1) {
        animationFrameId = window.requestAnimationFrame(step);
      }
    };

    observer = new IntersectionObserver((entries) => {
      if (entries[0].isIntersecting && !hasAnimated) {
        hasAnimated = true;
        animationFrameId = window.requestAnimationFrame(step);
      }
    }, { threshold: 0.1 });

    if (ref.current) observer.observe(ref.current);

    return () => {
      if (observer) observer.disconnect();
      if (animationFrameId) window.cancelAnimationFrame(animationFrameId);
    };
  }, [end, duration]);

  return <span ref={ref}>{count.toFixed(decimals)}</span>;
};

/* ═══════════════════════════════════════════════════════════ */
/*             AIR FREIGHT PAGE                              */
/* ═══════════════════════════════════════════════════════════ */

const LazyVideo = ({ src, className }) => {
  const [isIntersecting, setIsIntersecting] = useState(false);
  const ref = useRef(null);

  useEffect(() => {
    const observer = new IntersectionObserver(
      ([entry]) => {
        if (entry.isIntersecting) {
          setIsIntersecting(true);
          observer.disconnect();
        }
      },
      { rootMargin: '400px' }
    );
    if (ref.current) observer.observe(ref.current);
    return () => observer.disconnect();
  }, []);

  return (
    <video
      ref={ref}
      className={className}
      autoPlay={isIntersecting}
      muted
      loop
      playsInline
    >
      {isIntersecting && <source src={src} type="video/mp4" />}
    </video>
  );
};

export default function AirFreight() {

  useEffect(() => {
    const observer = new IntersectionObserver((entries) => {
      entries.forEach((entry) => {
        if (entry.isIntersecting) {
          entry.target.classList.add('af-in-view');
        }
      });
    }, { threshold: 0.15 });

    const elements = document.querySelectorAll('.af-reveal');
    elements.forEach((el) => observer.observe(el));

    return () => observer.disconnect();
  }, []);

  return (
    <div>
      {/* ═══ CINEMATIC HERO — 100vh ═══ */}
      <section className="af-hero" id="af-hero">

        {/* Background Video */}
        <video
          className="af-hero__video af-anim-video"
          autoPlay
          muted
          loop
          playsInline
          poster="/images/air-freight/Figure_on_platform_overlooking_a…_202606071514.jpeg"
        >
          <source
            src="/videos/air-freight/Air_cargo_hub_at_night_202606071525.mp4"
            type="video/mp4"
          />
        </video>

        {/* Cinematic Overlay */}
        <div className="af-hero__overlay" />

        {/* Ambient Route Lines */}
        <AmbientRoutes />

        {/* ─── Content ─── */}
        <div className="af-hero__content">

          {/* Eyebrow Badge */}
          <div className="af-badge af-anim-badge">
            <span className="af-badge__icon">✈</span>
            GLOBAL AIR FREIGHT NETWORK
          </div>

          {/* Main Headline — Dominant first line */}
          <h1 className="af-headline af-anim-headline">
            <span className="af-headline__line">Air Freight.</span>
          </h1>

          {/* Sub-headline — Smaller, lighter */}
          <p className="af-headline--sub af-anim-headline">
            Moving The World's<br />
            Most Critical Cargo.
          </p>

          {/* Subtext */}
          <p className="af-subtext af-anim-subtext">
            Move critical cargo across continents through a connected network
            of airlines, airports and logistics partners.
          </p>

          {/* CTAs */}
          <div className="af-ctas af-anim-ctas">
            <Link to="/contact" className="af-btn-primary">
              Request Air Quote
              <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
                <line x1="5" y1="12" x2="19" y2="12" />
                <polyline points="12 5 19 12 12 19" />
              </svg>
            </Link>
            <button className="af-btn-secondary" type="button">
              <svg width="16" height="16" viewBox="0 0 24 24" fill="currentColor">
                <polygon points="5 3 19 12 5 21 5 3" />
              </svg>
              Watch Network
            </button>
          </div>

        </div>

        {/* Floating Caption — Lower Right */}
        <div className="af-caption af-anim-caption">
          <span className="af-caption__dot" />
          150+ Airports Connected Worldwide
        </div>

        {/* Scroll Indicator */}
        <div className="af-scroll af-anim-scroll">
          <span className="af-scroll__text">Explore Network</span>
          <svg className="af-scroll__icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
            <path d="M12 5v14" />
            <path d="M19 12l-7 7-7-7" />
          </svg>
        </div>

      </section>

      {/* ═══════════════════════════════════════════════════════════ */}
      {/* ═══ SECTION 2: GLOBAL AIR FREIGHT NETWORK ═══ */}
      {/* ═══════════════════════════════════════════════════════════ */}
      <section className="af-network" id="af-network">
        <div className="af-network__container af-reveal">
          
          <div className="af-network__image-wrapper">
            {/* Image Layer with border radius */}
            <div className="af-network__img-frame">
              <div className="af-network__gradient"></div>
              <img 
                src="/images/air-freight/Human_figure_observing_air_cargo_202606071514.jpeg" 
                alt="Global Network Scale" 
                className="af-network__img"
              />
            </div>

            {/* Overlaid Header */}
            <div className="af-network__header">
              <div className="af-network__label">GLOBAL AIR FREIGHT NETWORK</div>
              <h2 className="af-network__title">
                Every Flight.<br />
                Every Airport.<br />
                One Connected System.
              </h2>
              <p className="af-network__desc">
                Freel continuously orchestrates air cargo
                across global airline and airport networks
                in real time.
              </p>
            </div>

            {/* Floating Glassmorphism Pills */}
            <div className="af-pill af-pill--1">150+ Airports</div>
            <div className="af-pill af-pill--2">Live Capacity</div>
            <div className="af-pill af-pill--3">24 Hour Movement</div>
            <div className="af-pill af-pill--4">Temperature Controlled</div>
            <div className="af-pill af-pill--5">Dangerous Goods</div>
            <div className="af-pill af-pill--6">Real Time Visibility</div>
          </div>
          
        </div>
      </section>

      {/* ═══════════════════════════════════════════════════════════ */}
      {/* ═══ SECTION 3: HOW CARGO MOVES (STORYBOARD) ═══ */}
      {/* ═══════════════════════════════════════════════════════════ */}
      <section className="af-storyboard">
        
        {/* Intro - 35vh */}
        <div className="af-storyboard__intro af-reveal">
          <div className="af-storyboard__label">HOW CARGO MOVES</div>
          <h2 className="af-storyboard__title">How Cargo Moves.</h2>
          <p className="af-storyboard__desc">
            Four stages.<br />
            Thousands of flights.<br />
            One connected network.
          </p>
        </div>

        {/* Stages */}
        <div className="af-storyboard__stages">
          
          {/* Stage 1 */}
          <div className="af-storyboard__stage af-reveal">
            <LazyVideo className="af-storyboard__video" src="/videos/air-freight/Air_freight_sorting_facility_pac_202606071530.mp4" />
            <div className="af-storyboard__gradient"></div>
            <div className="af-storyboard__content">
              <div className="af-storyboard__number">01</div>
              <h3 className="af-storyboard__stage-title">SORTED</h3>
              <p className="af-storyboard__stage-desc">Warehouse orchestration begins.</p>
            </div>
          </div>

          {/* Stage 2 */}
          <div className="af-storyboard__stage af-reveal">
            <LazyVideo className="af-storyboard__video" src="/videos/air-freight/Air_cargo_loading_operation_befo_202606071526.mp4" />
            <div className="af-storyboard__gradient"></div>
            <div className="af-storyboard__content">
              <div className="af-storyboard__number">02</div>
              <h3 className="af-storyboard__stage-title">LOADED</h3>
              <p className="af-storyboard__stage-desc">Prepared for departure.</p>
            </div>
          </div>

          {/* Stage 3 */}
          <div className="af-storyboard__stage af-reveal">
            <LazyVideo className="af-storyboard__video" src="/videos/air-freight/Aircraft_loading_mail_sacks_sunrise_202606071519.mp4" />
            <div className="af-storyboard__gradient"></div>
            <div className="af-storyboard__content">
              <div className="af-storyboard__number">03</div>
              <h3 className="af-storyboard__stage-title">IN TRANSIT</h3>
              <p className="af-storyboard__stage-desc">Moving across continents.</p>
            </div>
          </div>

          {/* Stage 4 */}
          <div className="af-storyboard__stage af-reveal">
            <LazyVideo className="af-storyboard__video" src="/videos/air-freight/Global_logistics_network_in_motion_202606071527.mp4" />
            <div className="af-storyboard__gradient"></div>
            <div className="af-storyboard__content">
              <div className="af-storyboard__number">04</div>
              <h3 className="af-storyboard__stage-title">CONNECTED</h3>
              <p className="af-storyboard__stage-desc">One coordinated network.</p>
            </div>
          </div>

        </div>
      </section>
      {/* ═══════════════════════════════════════════════════════════ */}
      {/* ═══ SECTION 4: OPERATIONS CONTROL TOWER ═══ */}
      {/* ═══════════════════════════════════════════════════════════ */}
      <section className="af-control">
        <div className="af-control__container">
          
          {/* Left Side: Visual */}
          <div className="af-control__visual af-reveal">
            <div className="af-control__dashboard">
              <LazyVideo className="af-control__video" src="/videos/air-freight/Global_logistics_network_in_motion_202606071527.mp4" />
              <div className="af-control__overlay"></div>
            </div>

            {/* Floating Badges */}
            <div className="af-badge af-badge--1">98% On-Time</div>
            <div className="af-badge af-badge--2">Live Capacity</div>
            <div className="af-badge af-badge--3">Customs Cleared</div>
            <div className="af-badge af-badge--4">Temperature Controlled</div>
            <div className="af-badge af-badge--5">Dangerous Goods</div>
            <div className="af-badge af-badge--6">24/7 Monitoring</div>
          </div>

          {/* Right Side: Content */}
          <div className="af-control__content af-reveal">
            <div className="af-control__label">OPERATIONS CONTROL TOWER</div>
            <h2 className="af-control__title">One View Of The Entire Network.</h2>
            <p className="af-control__desc">
              Monitor flights, capacity, cargo status and airport operations across a connected global network in real time.
            </p>

            <div className="af-control__stats">
              <div className="af-control__stat">
                <div className="af-control__stat-value">150+</div>
                <div className="af-control__stat-label">Airports</div>
              </div>
              <div className="af-control__stat">
                <div className="af-control__stat-value">24/7</div>
                <div className="af-control__stat-label">Monitoring</div>
              </div>
              <div className="af-control__stat">
                <div className="af-control__stat-value">10M+</div>
                <div className="af-control__stat-label">Shipments</div>
              </div>
              <div className="af-control__stat">
                <div className="af-control__stat-value">98%</div>
                <div className="af-control__stat-label">On-Time</div>
              </div>
            </div>
          </div>

        </div>
      </section>

      {/* ═══════════════════════════════════════════════════════════ */}
      {/* ═══ SECTION 5: CARGO CAPABILITIES ═══ */}
      {/* ═══════════════════════════════════════════════════════════ */}
      <section className="af-capabilities">
        <div className="af-capabilities__container">
          
          {/* Header */}
          <div className="af-capabilities__header af-reveal">
            <div className="af-capabilities__label">CARGO CAPABILITIES</div>
            <h2 className="af-capabilities__title">Built For Every Type Of Critical Freight.</h2>
            <p className="af-capabilities__desc">
              From temperature-sensitive pharmaceuticals to emergency humanitarian aid, the network adapts to every cargo requirement.
            </p>
          </div>

          {/* Grid */}
          <div className="af-capabilities__grid">
            
            {/* Card 1 */}
            <div className="af-capabilities__card af-reveal">
              <LazyVideo className="af-capabilities__video" src="/videos/air-freight/Cargo_aircraft_interior_moving_p_202606071518.mp4" />
              <div className="af-capabilities__gradient"></div>
              <div className="af-capabilities__content">
                <div className="af-capabilities__badge">2°–8°C</div>
                <h3 className="af-capabilities__card-title">Pharmaceutical Logistics</h3>
                <p className="af-capabilities__card-desc">Temperature-controlled movement of vaccines, biologics and critical healthcare shipments.</p>
              </div>
            </div>

            {/* Card 2 */}
            <div className="af-capabilities__card af-reveal">
              <LazyVideo className="af-capabilities__video" src="/videos/air-freight/Drone_push_toward_airport_cargo_202606071523.mp4" />
              <div className="af-capabilities__gradient"></div>
              <div className="af-capabilities__content">
                <div className="af-capabilities__badge">24 Hour Movement</div>
                <h3 className="af-capabilities__card-title">E-Commerce Fulfillment</h3>
                <p className="af-capabilities__card-desc">High-frequency shipments supporting modern retail and direct-to-consumer delivery.</p>
              </div>
            </div>

            {/* Card 3 */}
            <div className="af-capabilities__card af-reveal">
              <LazyVideo className="af-capabilities__video" src="/videos/air-freight/Humanitarian_airlift_in_progress_202606071529.mp4" />
              <div className="af-capabilities__gradient"></div>
              <div className="af-capabilities__content">
                <div className="af-capabilities__badge">Priority Access</div>
                <h3 className="af-capabilities__card-title">Humanitarian Relief</h3>
                <p className="af-capabilities__card-desc">Rapid deployment of emergency aid, food and medical supplies worldwide.</p>
              </div>
            </div>

            {/* Card 4 */}
            <div className="af-capabilities__card af-reveal">
              <LazyVideo className="af-capabilities__video" src="/videos/air-freight/Air_cargo_hub_loading_freighter_202606071526.mp4" />
              <div className="af-capabilities__gradient"></div>
              <div className="af-capabilities__content">
                <div className="af-capabilities__badge">24/7 Security</div>
                <h3 className="af-capabilities__card-title">High Value Cargo</h3>
                <p className="af-capabilities__card-desc">Secure transport of electronics, aerospace components and mission-critical freight.</p>
              </div>
            </div>

          </div>
        </div>
      </section>

      {/* ═══════════════════════════════════════════════════════════ */}
      {/* ═══ SECTION 6: NETWORK PERFORMANCE ═══ */}
      {/* ═══════════════════════════════════════════════════════════ */}
      <section className="af-performance">
        
        {/* Header */}
        <div className="af-performance__header af-reveal">
          <div className="af-performance__label">NETWORK PERFORMANCE</div>
          <h2 className="af-performance__title">Built For Reliability At Global Scale.</h2>
          <p className="af-performance__desc">
            Every route, airport and shipment is continuously optimized to keep freight moving.
          </p>
        </div>

        {/* Body */}
        <div className="af-performance__body af-reveal">
          
          {/* Left Metrics */}
          <div className="af-performance__column">
            <div className="af-performance__metric">
              <div className="af-performance__number"><CountUp end={98.7} decimals={1} /><span className="af-performance__symbol">%</span></div>
              <div className="af-performance__text">On-Time Performance</div>
            </div>
            <div className="af-performance__metric">
              <div className="af-performance__number"><CountUp end={150} /><span className="af-performance__symbol">+</span></div>
              <div className="af-performance__text">Airport Partners</div>
            </div>
            <div className="af-performance__metric">
              <div className="af-performance__number"><CountUp end={10} /><span className="af-performance__symbol">M+</span></div>
              <div className="af-performance__text">Shipments Managed</div>
            </div>
          </div>

          {/* Center Globe */}
          <div className="af-performance__globe-container">
            <div className="af-performance__globe">
              <LazyVideo className="af-performance__globe-video" src="/videos/air-freight/Global_logistics_network_in_motion_202606071527.mp4" />
              <div className="af-performance__globe-overlay"></div>
            </div>
          </div>

          {/* Right Metrics */}
          <div className="af-performance__column">
            <div className="af-performance__metric">
              <div className="af-performance__number"><CountUp end={24} /><span className="af-performance__symbol">/7</span></div>
              <div className="af-performance__text">Monitoring</div>
            </div>
            <div className="af-performance__metric">
              <div className="af-performance__number"><CountUp end={99.99} decimals={2} /><span className="af-performance__symbol">%</span></div>
              <div className="af-performance__text">Platform Availability</div>
            </div>
            <div className="af-performance__metric">
              <div className="af-performance__number"><CountUp end={48} /><span className="af-performance__symbol af-performance__symbol--text">Hours</span></div>
              <div className="af-performance__text">Average Global Transit</div>
            </div>
          </div>

        </div>
      </section>

      {/* ═══════════════════════════════════════════════════════════ */}
      {/* ═══ SECTION 7: READY FOR TAKEOFF (CONVERSION) ═══ */}
      {/* ═══════════════════════════════════════════════════════════ */}
      <section className="af-takeoff">
        <LazyVideo className="af-takeoff__video" src="/videos/air-freight/Air_cargo_hub_at_night_202606071525.mp4" />
        
        <div className="af-takeoff__overlay"></div>
        <div className="af-takeoff__gradient"></div>

        <div className="af-takeoff__content">
          <div className="af-takeoff__label af-reveal">GLOBAL AIR FREIGHT PLATFORM</div>
          <h2 className="af-takeoff__title af-reveal">The World Doesn't Wait.<br/>Neither Should Your Freight.</h2>
          <p className="af-takeoff__desc af-reveal">
            Connect airlines, airports and logistics partners through one intelligent operating system.
          </p>

          <div className="af-takeoff__actions af-reveal">
            <Link to="/signup" className="af-takeoff__btn af-takeoff__btn--primary">Start Free</Link>
            <Link to="/demo" className="af-takeoff__btn af-takeoff__btn--secondary">Book Demo</Link>
          </div>

          <div className="af-takeoff__trust">
            <div className="af-takeoff__trust-item af-reveal">150+ Airports</div>
            <div className="af-takeoff__trust-divider af-reveal"></div>
            <div className="af-takeoff__trust-item af-reveal">24/7 Monitoring</div>
            <div className="af-takeoff__trust-divider af-reveal"></div>
            <div className="af-takeoff__trust-item af-reveal">10M+ Shipments</div>
            <div className="af-takeoff__trust-divider af-reveal"></div>
            <div className="af-takeoff__trust-item af-reveal">98% On-Time</div>
          </div>
        </div>
      </section>

    </div>
  );
}

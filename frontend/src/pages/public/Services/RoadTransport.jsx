import { useEffect, useRef, useState } from 'react';
import { Link } from 'react-router-dom';

/* ─── Scroll Reveal ─── */
function useReveal() {
  const ref = useRef(null);
  useEffect(() => {
    const el = ref.current;
    if (!el) return;
    const obs = new IntersectionObserver(
      ([e]) => {
        if (e.isIntersecting) {
          el.classList.add('opacity-100', 'translate-y-0');
          el.classList.remove('opacity-0', 'translate-y-8');
          obs.unobserve(el);
        }
      },
      { threshold: 0.15, rootMargin: '0px 0px -50px 0px' }
    );
    obs.observe(el);
    return () => obs.disconnect();
  }, []);
  return ref;
}

function Reveal({ children, className = '', delay = '' }) {
  const ref = useReveal();
  return (
    <div ref={ref} className={`transition-all duration-1000 opacity-0 translate-y-8 ${delay} ${className}`}>
      {children}
    </div>
  );
}



const vehicleTypes = [
  { icon: '🚛', title: 'Full Truck Load (FTL)', desc: 'Dedicated trucks from 1T to 40T. Open body, container, trailer. Direct door-to-door with optimized toll routes.', color: 'from-amber-400 to-amber-600' },
  { icon: '📦', title: 'Less Than Truck (LTL)', desc: 'Share truck space, pay only for your cargo. Ideal for 100kg to 5T shipments. High-speed hub-to-hub network.', color: 'from-blue-400 to-blue-600' },
  { icon: '🏗️', title: 'ODC & Heavy Haulage', desc: 'Oversized cargo on low-bed trailers. Heavy machinery, industrial boilers. Feasibility and route survey included.', color: 'from-purple-400 to-purple-600' },
  { icon: '⚡', title: 'Express Freight', desc: 'Same-day and next-day deliveries between industrial hubs. Dedicated dual-driver long-haul vehicles.', color: 'from-red-400 to-red-600' },
  { icon: '❄️', title: 'Cold Chain / Reefer', desc: 'Temperature-controlled trucks for pharma, perishables, and specialty chemicals. Active GDP-compliant logging.', color: 'from-teal-400 to-teal-600' },
  { icon: '📋', title: 'Paperless Lorry Receipt', desc: 'Instant, secure E-POD with QR-code verification, digital signatures, and immediate client notification.', color: 'from-indigo-400 to-indigo-600' },
];

const topLanes = [
  { from: 'Delhi NCR', to: 'Mumbai (JNPT)', dist: '1,400 km', time: '24-28 hrs', rate: 'Optimized Tolls' },
  { from: 'Mumbai', to: 'Bengaluru', dist: '980 km', time: '18-22 hrs', rate: 'High Frequency' },
  { from: 'Delhi', to: 'Kolkata', dist: '1,500 km', time: '28-32 hrs', rate: 'Eastern Corridor' },
  { from: 'Chennai', to: 'Hyderabad', dist: '630 km', time: '12-14 hrs', rate: 'Express Transit' },
  { from: 'Pune', to: 'JNPT Port', dist: '160 km', time: '4-5 hrs', rate: 'Port Shuttles' },
  { from: 'Ahmedabad', to: 'Delhi', dist: '940 km', time: '16-20 hrs', rate: 'Industrial Direct' },
];

import CinematicShipmentJourney from './CinematicShipmentJourney';
import JourneyBridge from './JourneyBridge';
import IndustriesWeMove from './IndustriesWeMove';
import WhyChooseFreel from './WhyChooseFreel';
import RoadFreightCTA from './RoadFreightCTA';
import './RoadTransport.css';

/* ═══════════════════════════════════════════ */
/*           ROAD TRANSPORT PAGE              */
/* ═══════════════════════════════════════════ */
export default function RoadTransport() {
  const [activeLane, setActiveLane] = useState(0);

  useEffect(() => {
    const observer = new IntersectionObserver((entries) => {
      entries.forEach((entry) => {
        if (entry.isIntersecting) {
          entry.target.classList.add('rt-in-view');
        }
      });
    }, { threshold: 0.15 });

    const elements = document.querySelectorAll('.rt-reveal');
    elements.forEach((el) => observer.observe(el));

    return () => observer.disconnect();
  }, []);

  const trackingLanes = [
    {
      from: 'DEL',
      to: 'BOM',
      laneName: 'Delhi NCR → Mumbai (JNPT)',
      vehicle: 'Tata Prima 4925.S (40T Container)',
      driver: 'Satnam Singh (ID: FR-9082)',
      speed: '74 km/h',
      location: 'Near Jaipur Bypass, NH-48',
      status: 'In Transit (On Time)',
      progress: '38%',
      eta: '18h 45m remaining',
      compliance: 'E-Way Bill Verified',
      temp: 'Dry Cargo',
    },
    {
      from: 'BOM',
      to: 'BLR',
      laneName: 'Mumbai → Bengaluru Hub',
      vehicle: 'Ashok Leyland Ecomet (14T Reefer)',
      driver: 'Ramesh Gowda (ID: FR-4112)',
      speed: '68 km/h',
      location: 'Near Kolhapur Toll Plaza, NH-4',
      status: 'Cold Chain Active',
      progress: '62%',
      eta: '8h 20m remaining',
      compliance: 'GDP Compliant',
      temp: '4.2°C (Set: 4.0°C)',
    },
    {
      from: 'PUN',
      to: 'JNP',
      laneName: 'Pune → JNPT Port Terminal',
      vehicle: 'Mahindra Blazo X (ODC 12-Axle)',
      driver: 'Vilas Patil (ID: FR-0239)',
      speed: '45 km/h',
      location: 'Expressway, Khopoli Ghat Section',
      status: 'Heavy Haul (Escort Active)',
      progress: '85%',
      eta: '1h 15m remaining',
      compliance: 'RTO Permission Cleared',
      temp: 'Industrial Turbine',
    },
  ];

  return (
    <div>
      {/* ═══ CINEMATIC HERO — 100vh ═══ */}
      <section className="rt-hero rt-reveal" id="rt-hero">

        {/* Background Video */}
        <video
          className="rt-hero__video rt-anim-video"
          autoPlay
          muted
          loop
          playsInline
        >
          <source
            src="/videos/road-freight/Convoy_of_trucks_moving_highway_202606071649.mp4"
            type="video/mp4"
          />
        </video>

        {/* Cinematic Overlay */}
        <div className="rt-hero__overlay" />

        {/* Floating Caption (Replacing huge glass stats) */}
        <div className="rt-caption">
          <div className="rt-caption__dot"></div>
          500+ VERIFIED TRANSPORTERS ACROSS INDIA
        </div>

        {/* ─── Content ─── */}
        <div className="rt-hero__content">

          {/* Eyebrow Badge */}
          <div className="rt-badge rt-anim-badge">
            <span className="rt-badge__icon">🚛</span>
            ROAD FREIGHT NETWORK
          </div>

          {/* Main Headline */}
          <h1 className="rt-headline rt-anim-headline">
            <span className="rt-headline__line">Road Freight.</span>
          </h1>

          {/* Sub-headline */}
          <p className="rt-headline--sub rt-anim-headline">
            Connecting Every City.<br />
            Delivering Every Mile.
          </p>

          {/* Subtext */}
          <p className="rt-subtext rt-anim-subtext">
            From factory gates and distribution hubs to ports and retail shelves,
            road freight forms the backbone of modern commerce,
            moving goods across every corner of the nation.
          </p>

          {/* CTAs */}
          <div className="rt-ctas rt-anim-ctas">
            <Link to="/contact" className="rt-btn-primary">
              Request Road Quote
              <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
                <line x1="5" y1="12" x2="19" y2="12" />
                <polyline points="12 5 19 12 12 19" />
              </svg>
            </Link>
            <button className="rt-btn-secondary" type="button" onClick={() => document.getElementById('gps-tracking-demo').scrollIntoView({ behavior: 'smooth' })}>
              <svg width="16" height="16" viewBox="0 0 24 24" fill="currentColor">
                <polygon points="5 3 19 12 5 21 5 3" />
              </svg>
              Watch Journey
            </button>
          </div>

        </div>

        {/* Scroll Indicator */}
        <div className="rt-scroll-indicator">
          EXPLORE NETWORK
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
            <line x1="12" y1="5" x2="12" y2="19" />
            <polyline points="19 12 12 19 5 12" />
          </svg>
        </div>
      </section>

      {/* ═══ NARRATIVE BRIDGE ═══ */}
      <JourneyBridge />

      {/* ═══ THE NATION IN MOTION (SCROLL EXPERIENCE) ═══ */}
      <CinematicShipmentJourney />

      {/* ═══ INDUSTRIES WE MOVE ═══ */}
      <IndustriesWeMove />

      {/* ═══ WHY CHOOSE FREEL ═══ */}
      <WhyChooseFreel />

      {/* ═══ VEHICLE TYPES ═══ */}
      <section className="py-32 bg-white relative">
        <div className="max-w-7xl mx-auto px-4 sm:px-6">
          <div className="text-center mb-20">
            <span className="text-amber-500 font-bold tracking-widest uppercase text-sm mb-3 block">
              Fleet Capabilities
            </span>
            <h2 className="text-4xl md:text-5xl font-extrabold text-brand-navy mb-4 leading-tight">
              Designed For Any Cargo Profile
            </h2>
            <p className="text-slate-500 text-lg max-w-2xl mx-auto">
              From high-capacity container shipping to cold chain pharmaceuticals, our diverse pre-vetted fleet is ready to deploy on demand.
            </p>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-8">
            {vehicleTypes.map((v, i) => (
              <Reveal key={i} delay={`delay-${(i % 3) * 100}`}>
                <div className="bg-slate-50/50 rounded-[2rem] p-8 border border-slate-200/60 hover:border-amber-300 hover:shadow-2xl hover:bg-white transition-all group h-full flex flex-col justify-between">
                  <div>
                    <div className={`w-14 h-14 rounded-2xl bg-gradient-to-br ${v.color} flex items-center justify-center text-3xl text-white shadow-lg mb-6 group-hover:scale-110 transition-transform`}>
                      {v.icon}
                    </div>
                    <h3 className="text-xl font-bold text-slate-800 mb-3 group-hover:text-amber-500 transition-colors">
                      {v.title}
                    </h3>
                    <p className="text-slate-500 text-sm leading-relaxed mb-6">{v.desc}</p>
                  </div>
                  <div className="inline-flex items-center text-xs font-bold text-brand-indigo group-hover:translate-x-1.5 transition-transform cursor-pointer">
                    Spec Sheets Available <span className="ml-1">→</span>
                  </div>
                </div>
              </Reveal>
            ))}
          </div>
        </div>
      </section>



      {/* ═══ TOP LANES ═══ */}
      <section className="py-32 bg-slate-50 border-t border-slate-200/50">
        <div className="max-w-7xl mx-auto px-4 sm:px-6">
          <div className="flex flex-col md:flex-row justify-between items-end mb-16">
            <div>
              <span className="text-amber-500 font-bold tracking-widest uppercase text-sm mb-3 block">
                Popular Lanes
              </span>
              <h2 className="text-4xl md:text-5xl font-extrabold text-brand-navy mb-4 leading-tight">
                High-Frequency Corridors
              </h2>
              <p className="text-slate-500 text-lg max-w-2xl">
                Daily departures with guaranteed space allocation, pre-cleared toll passes, and standardized pricing models.
              </p>
            </div>
            <Link
              to="/services"
              className="mt-6 md:mt-0 text-brand-indigo font-bold hover:text-amber-500 transition-colors inline-flex items-center gap-1"
            >
              View Full Route Network <span>→</span>
            </Link>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {topLanes.map((lane, i) => (
              <Reveal key={i} delay={`delay-${(i % 3) * 100}`}>
                <div className="bg-white rounded-2xl p-6 border border-slate-200/60 hover:border-amber-400 hover:shadow-xl transition-all cursor-pointer flex justify-between items-center group">
                  <div>
                    <div className="flex items-center text-lg font-bold text-slate-800 mb-2">
                      {lane.from} <span className="text-amber-500 mx-2">→</span> {lane.to}
                    </div>
                    <div className="text-xs font-semibold text-slate-500 flex items-center gap-2">
                      <span>📍 {lane.dist}</span>
                      <span className="text-slate-300">|</span>
                      <span>⏱️ {lane.time}</span>
                      <span className="text-slate-300">|</span>
                      <span className="text-amber-600 font-bold">{lane.rate}</span>
                    </div>
                  </div>
                  <div className="w-10 h-10 rounded-full bg-slate-50 flex items-center justify-center text-amber-500 group-hover:bg-amber-500 group-hover:text-white transition-all shadow-sm">
                    ↗
                  </div>
                </div>
              </Reveal>
            ))}
          </div>
        </div>
      </section>

      {/* ═══ PREMIUM CLOSING CTA ═══ */}
      <RoadFreightCTA />
    </div>
  );
}

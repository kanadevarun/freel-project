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

const stats = [
  { value: '500+', label: 'Verified Transporters', desc: 'Pre-vetted, GPS-enabled premium carriers' },
  { value: 'Pan-India', label: 'Ground Network', desc: 'Seamless door-to-door regional coverage' },
  { value: '100%', label: 'Digital Compliance', desc: 'Instant LR, E-Way Bill & E-POD updates' },
  { value: '24/7', label: 'Control Room Sync', desc: 'Active telemetry and route escort systems' },
];

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

/* ═══════════════════════════════════════════ */
/*           ROAD TRANSPORT PAGE              */
/* ═══════════════════════════════════════════ */
export default function RoadTransport() {
  const [activeLane, setActiveLane] = useState(0);

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
    <div className="bg-white min-h-screen overflow-hidden">
      {/* ═══ HERO ═══ */}
      <section className="relative pt-32 pb-32 lg:pt-48 lg:pb-48 bg-gradient-to-br from-amber-50/20 via-slate-50 to-indigo-50/20 text-slate-900 overflow-hidden">
        {/* Container Truck Watermark Background */}
        <div className="absolute inset-0 z-0">
          <img
            src="/assets/road_freight_hero.png"
            alt="Premium Freight Container Truck"
            className="w-full h-full object-cover object-center opacity-[0.08] scale-105 mix-blend-multiply animate-[pulse_20s_ease-in-out_infinite_alternate]"
          />
          <div className="absolute inset-0 bg-gradient-to-t from-white via-white/40 to-slate-50/80"></div>
          <div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-full max-w-7xl h-full pointer-events-none opacity-40">
            <div className="absolute -top-[10%] right-[-5%] w-[450px] h-[450px] rounded-full bg-amber-400/10 blur-3xl"></div>
            <div className="absolute bottom-[-10%] left-[-5%] w-[550px] h-[550px] rounded-full bg-brand-indigo/10 blur-3xl"></div>
          </div>
        </div>

        <div className="max-w-7xl mx-auto px-4 sm:px-6 relative z-10">
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-12 items-center">
            <Reveal>
              <div>
                <div className="inline-flex items-center justify-center px-4 py-1.5 mb-6 text-xs font-bold text-amber-700 bg-amber-500/10 rounded-full border border-amber-500/20 tracking-widest uppercase">
                  <span className="mr-2">🚛</span> Ground Logistics
                </div>
                <h1 className="text-5xl md:text-6xl lg:text-7xl font-extrabold tracking-tight mb-6 leading-tight text-slate-900">
                  Reliable & Smart <br />
                  <span className="text-transparent bg-clip-text bg-gradient-to-r from-amber-500 to-orange-600">
                    Road Freight
                  </span>
                </h1>
                <p className="text-xl text-slate-600 max-w-2xl font-light leading-relaxed mb-10">
                  Seamless FTL & LTL transportation across India. Experience the power of 500+ pre-vetted carriers, real-time GPS tracking, and fully digital compliance from loading to EPOD.
                </p>

                <div className="flex flex-wrap gap-4">
                  <Link
                    to="/contact"
                    className="px-8 py-4 bg-amber-500 hover:bg-amber-600 text-white font-bold rounded-full transition-colors shadow-[0_0_20px_rgba(245,158,11,0.3)]"
                  >
                    Get Road Freight Quote
                  </Link>
                  <a
                    href="#gps-tracking-demo"
                    className="px-8 py-4 bg-white/80 border border-slate-200 text-slate-700 font-bold rounded-full hover:bg-slate-100 hover:border-slate-300 transition-colors shadow-sm backdrop-blur-md"
                  >
                    View GPS Telemetry
                  </a>
                </div>
              </div>
            </Reveal>

            {/* Visual Hero Element - E-Way Bill & Live Driver Card */}
            <Reveal delay="delay-200">
              <div className="relative h-[400px] bg-white/80 rounded-3xl border border-slate-200/80 backdrop-blur-md p-6 shadow-2xl flex flex-col justify-between overflow-hidden">
                <div className="absolute -top-32 -right-32 w-64 h-64 bg-amber-500/10 rounded-full blur-[80px]"></div>

                <div className="flex justify-between items-center relative z-10">
                  <div>
                    <div className="text-sm text-slate-500 mb-1">Telemetry Status</div>
                    <div className="text-slate-800 font-bold text-xl flex items-center gap-2">
                      <span className="w-2.5 h-2.5 rounded-full bg-green-500 animate-pulse"></span> GPS Connected
                    </div>
                  </div>
                  <div className="text-right">
                    <div className="text-sm text-slate-500 mb-1">E-Way Bill</div>
                    <div className="text-slate-800 font-bold text-xl font-mono">EWB-8921A</div>
                  </div>
                </div>

                <div className="relative z-10 py-8">
                  <div className="flex justify-between text-3xl font-black mb-2 text-slate-800">
                    <span>DELHI</span>
                    <span className="text-amber-500">🚛</span>
                    <span>MUMBAI</span>
                  </div>

                  <div className="relative h-2 bg-slate-200 rounded-full mt-4">
                    <div className="absolute top-0 left-0 h-full bg-gradient-to-r from-amber-500 to-orange-500 rounded-full w-[45%]">
                      <div className="absolute -right-2 -top-1 w-4 h-4 bg-white border border-slate-300 rounded-full shadow-[0_0_10px_rgba(0,0,0,0.15)]"></div>
                    </div>
                  </div>
                  <div className="flex justify-between text-xs text-slate-500 mt-2 font-mono">
                    <span>NH-48 Corridor</span>
                    <span>JNPT Terminal</span>
                  </div>
                </div>

                <div className="grid grid-cols-2 gap-4 relative z-10">
                  <div className="bg-slate-50/80 border border-slate-100 rounded-xl p-3">
                    <div className="text-xs text-slate-500">Speed / Driver</div>
                    <div className="text-md font-bold text-slate-800">72 km/h • S. Singh</div>
                  </div>
                  <div className="bg-slate-50/80 border border-slate-100 rounded-xl p-3">
                    <div className="text-xs text-slate-500">Digital Compliance</div>
                    <div className="text-md font-bold text-slate-800 text-green-600">LR & QR Ready</div>
                  </div>
                </div>
              </div>
            </Reveal>
          </div>
        </div>
      </section>

      {/* ═══ STATS BAR (Floating) ═══ */}
      <Reveal>
        <section className="max-w-6xl mx-auto px-4 sm:px-6 relative z-20 -mt-10">
          <div className="bg-white rounded-[2rem] p-8 shadow-[0_20px_50px_-10px_rgba(0,0,0,0.06)] border border-slate-100 grid grid-cols-2 md:grid-cols-4 gap-6 divide-x-0 md:divide-x divide-slate-100">
            {stats.map((s, i) => (
              <div key={i} className="text-center group">
                <div className="text-3xl md:text-4xl font-black text-transparent bg-clip-text bg-gradient-to-br from-amber-500 to-orange-600 mb-2 group-hover:scale-105 transition-transform">
                  {s.value}
                </div>
                <div className="text-sm font-bold uppercase tracking-wider text-slate-700 mb-1">{s.label}</div>
                <div className="text-xs text-slate-400 px-2">{s.desc}</div>
              </div>
            ))}
          </div>
        </section>
      </Reveal>

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

      {/* ═══ SMART WAREHOUSING & CROSS-DOCK ORCHESTRATION ═══ */}
      <section className="py-32 bg-slate-50 relative overflow-hidden border-t border-b border-slate-200/50">
        <div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-full max-w-7xl h-full pointer-events-none opacity-20">
          <div className="absolute -top-[20%] left-[-10%] w-[600px] h-[600px] rounded-full bg-brand-teal/20 blur-3xl"></div>
        </div>

        <div className="max-w-7xl mx-auto px-4 sm:px-6 relative z-10">
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-16 items-center">
            
            <Reveal>
              <div className="relative rounded-[3rem] overflow-hidden shadow-2xl border border-slate-200 group h-[580px]">
                <div className="absolute inset-0 bg-amber-500/10 group-hover:bg-transparent transition-colors z-10 mix-blend-overlay duration-700"></div>
                <img
                  src="/assets/smart_warehouse.png"
                  alt="Automated Smart Warehousing and Cross Docking Facilities"
                  className="w-full h-full object-cover transform group-hover:scale-105 transition-transform duration-1000"
                />
                <div className="absolute bottom-0 left-0 w-full p-8 bg-gradient-to-t from-slate-900 via-slate-900/60 to-transparent z-20">
                  <div className="text-amber-400 font-mono text-xs mb-1 uppercase tracking-widest">
                    Automated Logistics Hubs
                  </div>
                  <div className="text-white font-bold text-2xl">
                    Smart Cross-Docking Orchestration
                  </div>
                </div>
              </div>
            </Reveal>

            <div>
              <Reveal>
                <div className="mb-10">
                  <span className="text-brand-indigo font-bold tracking-widest uppercase text-sm mb-3 block">
                    Strategic Fulfillment Nodes
                  </span>
                  <h2 className="text-4xl md:text-5xl font-extrabold text-brand-navy mb-4 leading-tight">
                    Smart Warehousing & Cross-Dock Hubs
                  </h2>
                  <p className="text-slate-600 text-lg leading-relaxed">
                    Freel seamlessly integrates road freight with premier **IoT-enabled warehousing** and cross-dock facilities at major trade centers. We coordinate truck arrivals with automated bay dispatches to ensure that goods transition from national FTL transits to last-mile fleets instantly.
                  </p>
                </div>
              </Reveal>

              <div className="space-y-6">
                <Reveal delay="delay-100">
                  <div className="flex gap-4 p-5 bg-white rounded-2xl border border-slate-100 shadow-sm hover:shadow-md transition-shadow">
                    <span className="text-3xl">📦</span>
                    <div>
                      <h4 className="font-bold text-slate-800 text-lg mb-1">Direct Port Delivery (DPD) Shuttles</h4>
                      <p className="text-slate-500 text-sm">
                        Pre-cleared containers bypass port storage and route straight to designated regional cross-dock warehouses.
                      </p>
                    </div>
                  </div>
                </Reveal>

                <Reveal delay="delay-200">
                  <div className="flex gap-4 p-5 bg-white rounded-2xl border border-slate-100 shadow-sm hover:shadow-md transition-shadow">
                    <span className="text-3xl">⚡</span>
                    <div>
                      <h4 className="font-bold text-slate-800 text-lg mb-1">2-Hour Cross-Dock Windows</h4>
                      <p className="text-slate-500 text-sm">
                        De-consolidating, sorting, and reloading express cargo onto last-mile delivery vans under record timelines with zero demurrage.
                      </p>
                    </div>
                  </div>
                </Reveal>

                <Reveal delay="delay-300">
                  <div className="flex gap-4 p-5 bg-white rounded-2xl border border-slate-100 shadow-sm hover:shadow-md transition-shadow">
                    <span className="text-3xl">🛡️</span>
                    <div>
                      <h4 className="font-bold text-slate-800 text-lg mb-1">Real-Time WMS Synchronization</h4>
                      <p className="text-slate-500 text-sm">
                        Our digital operating system automatically updates inventory manifests, barcode logs, and stock status immediately.
                      </p>
                    </div>
                  </div>
                </Reveal>
              </div>
            </div>
            
          </div>
        </div>
      </section>

      {/* ═══ INTERACTIVE GPS TELEMETRY & ROUTE WIDGET ═══ */}
      <section id="gps-tracking-demo" className="py-32 bg-brand-navy text-white relative overflow-hidden">
        <div className="absolute inset-0 bg-[url('data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iNjAiIGhlaWdodD0iNjAiIHZpZXdCb3g9IjAgMCA2MCA2MCIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj48ZyBmaWxsPSJub25lIiBmaWxsLXJ1bGU9ImV2ZW5vZGQiPjxwYXRoIGQ9Ik0zNiAzNHYtNGgtdjRoLTR2NGgtdjRoNHY0aDR2LTRoNHptMC0xMnYtNGgtdjRoLTR2NGgtdjRoNHY0aDR2LTRoNHptLTExIDEwaC00djRoLTR2LTRoLTR2LTRoNHYtNGg0djRoNHY0em0tMTEgMTBoLTR2NGgtdjRoLTR2LTRoNHYtNGg0djRoNHY0em0xMSAwaC00djRoLTR2LTRoLTR2LTRoNHYtNGg0djRoNHY0em0xMSAwaC00djRoLTR2LTRoLTR2LTRoNHYtNGg0djRoNHY0eiIgZmlsbD0iI2ZmZmZmZiIgZmlsbC1vcGFjaXR5PSIwLjAyIi8+PC9nPjwvc3ZnPg==')] opacity-50 z-0"></div>

        <div className="max-w-7xl mx-auto px-4 sm:px-6 relative z-10">
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-16 items-center">
            
            <Reveal>
              <div>
                <span className="text-amber-400 font-bold tracking-widest uppercase text-sm mb-3 block">
                  Simulated Control Room
                </span>
                <h2 className="text-4xl md:text-5xl font-extrabold mb-6">Live GPS Fleet Control</h2>
                <p className="text-slate-300 text-lg mb-8 leading-relaxed">
                  Every truck in the Freel network reports live telemetry. Select an active lane below to sync with our live telemetry simulator and watch driver checkpoints in real-time.
                </p>

                <div className="space-y-4">
                  {trackingLanes.map((lane, idx) => (
                    <button
                      key={idx}
                      onClick={() => setActiveLane(idx)}
                      className={`w-full text-left p-5 rounded-2xl border transition-all ${
                        activeLane === idx
                          ? 'bg-white/10 border-amber-400 shadow-[0_0_15px_rgba(245,158,11,0.2)]'
                          : 'bg-transparent border-white/10 hover:bg-white/5'
                      }`}
                    >
                      <div className="flex justify-between items-center">
                        <div className="flex items-center gap-4">
                          <div className={`w-3 h-3 rounded-full ${idx === 1 ? 'bg-teal-400 animate-pulse' : 'bg-green-400'}`}></div>
                          <span className="font-bold font-mono">{lane.from} → {lane.to}</span>
                        </div>
                        <span className="text-xs font-mono text-slate-400">{lane.vehicle.split(' ')[0]} {lane.vehicle.split(' ')[1]}</span>
                      </div>
                      <div className="text-xs text-slate-400 font-mono mt-1 pl-7">{lane.laneName}</div>
                    </button>
                  ))}
                </div>
              </div>
            </Reveal>

            <Reveal delay="delay-200">
              <div className="bg-[#0f172a] rounded-[2.5rem] p-8 border border-slate-800 shadow-2xl relative overflow-hidden">
                {/* Telemetry pulse grid */}
                <div className="absolute top-1/2 left-1/2 w-[140%] h-[140%] -translate-x-1/2 -translate-y-1/2 bg-[conic-gradient(from_0deg,transparent_0_340deg,rgba(245,158,11,0.1)_360deg)] rounded-full animate-[spin_6s_linear_infinite] z-0 mix-blend-screen pointer-events-none"></div>

                <div className="relative z-10">
                  <div className="flex justify-between items-center border-b border-slate-800 pb-4 mb-8">
                    <div className="font-mono text-amber-400 text-xs">FLEET_CONTROL_SYNC_OK</div>
                    <div className="text-xs bg-amber-500/20 text-amber-400 px-2.5 py-1 rounded-full font-bold">
                      ACTIVE GPS
                    </div>
                  </div>

                  <div className="grid grid-cols-2 gap-6 mb-8">
                    <div>
                      <div className="text-slate-400 text-xs font-mono uppercase">Vehicle</div>
                      <div className="text-white font-bold font-mono text-sm mt-1">{trackingLanes[activeLane].vehicle}</div>
                    </div>
                    <div>
                      <div className="text-slate-400 text-xs font-mono uppercase">Assigned Driver</div>
                      <div className="text-white font-bold font-mono text-sm mt-1">{trackingLanes[activeLane].driver}</div>
                    </div>
                    <div>
                      <div className="text-slate-400 text-xs font-mono uppercase">Current Location</div>
                      <div className="text-white font-bold font-mono text-sm mt-1">{trackingLanes[activeLane].location}</div>
                    </div>
                    <div>
                      <div className="text-slate-400 text-xs font-mono uppercase">Speed / Telemetry</div>
                      <div className="text-white font-bold font-mono text-sm mt-1">
                        {trackingLanes[activeLane].speed} • <span className="text-teal-400">{trackingLanes[activeLane].temp}</span>
                      </div>
                    </div>
                  </div>

                  <div className="border-t border-slate-800/80 pt-6 mt-6">
                    <div className="flex justify-between text-xs font-bold text-slate-500 mb-2 uppercase">
                      <span>Dispatch</span>
                      <span className="text-amber-400">{trackingLanes[activeLane].status}</span>
                      <span>Destination</span>
                    </div>

                    <div className="relative h-2.5 bg-slate-800 rounded-full mt-2">
                      <div
                        className="absolute top-0 left-0 h-full bg-gradient-to-r from-amber-500 to-orange-500 rounded-full transition-all duration-1000 z-10"
                        style={{ width: trackingLanes[activeLane].progress }}
                      >
                        <div className="absolute right-0 top-1/2 -translate-y-1/2 w-4 h-4 bg-white rounded-full shadow-[0_0_10px_#f59e0b]"></div>
                      </div>
                    </div>

                    <div className="flex justify-between mt-3 text-xs font-mono text-slate-400">
                      <span>{trackingLanes[activeLane].from} Hub</span>
                      <span>ETA: {trackingLanes[activeLane].eta}</span>
                      <span>{trackingLanes[activeLane].to} Hub</span>
                    </div>
                  </div>

                  <div className="mt-8 flex gap-3">
                    <button className="flex-1 py-3.5 bg-white/5 border border-white/10 hover:bg-white/10 text-white text-xs font-bold rounded-xl transition-all font-mono">
                      📄 View E-Way Bill
                    </button>
                    <button className="flex-1 py-3.5 bg-amber-500 text-slate-900 text-xs font-bold rounded-xl hover:bg-amber-400 transition-all font-mono">
                      📱 Contact Driver via App
                    </button>
                  </div>
                </div>
              </div>
            </Reveal>
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

      {/* ═══ CTA SECTION ═══ */}
      <section className="py-24 bg-white text-center">
        <Reveal>
          <div className="max-w-3xl mx-auto px-4">
            <h2 className="text-4xl md:text-5xl font-black text-brand-navy mb-6">Ready to Ship by Road?</h2>
            <p className="text-lg text-slate-600 mb-10">
              Instantly book dedicated FTL trucks or coordinate low-cost LTL shared space. Full GPS telemetry and paperless lorry receipts come standard.
            </p>
            <Link
              to="/contact"
              className="inline-flex items-center gap-2 px-10 py-5 bg-gradient-to-r from-amber-500 to-orange-500 text-white font-bold text-lg rounded-full hover:shadow-[0_10px_30px_rgba(245,158,11,0.3)] transition-all hover:-translate-y-1"
            >
              Request Road Freight Quote <span>🚛</span>
            </Link>
          </div>
        </Reveal>
      </section>
    </div>
  );
}

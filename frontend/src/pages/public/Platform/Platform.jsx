import { useEffect, useRef, useState } from 'react';
import { Link } from 'react-router-dom';

/* ─── Scroll Reveal Hook ─── */
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

const platformStats = [
  { value: '4 Modes', label: 'Single Dashboard', desc: 'Seamlessly orchestrating Road, Rail, Air & Sea shipments' },
  { value: '14.2%', label: 'Avg. Cost Savings', desc: 'Direct discovery of the lowest pre-negotiated freight rates' },
  { value: 'Instant', label: 'E-Way & LR Filings', desc: 'Fully automated compliance with paperless document wallets' },
  { value: '100%', label: 'Supply Chain Clarity', desc: 'GPS, AIS & flight telemetry in one unified map layout' },
];

export default function Platform() {
  const [activePillar, setActivePillar] = useState('center');

  const pillars = [
    {
      id: 'center',
      title: 'Command Center',
      icon: '🖥️',
      desc: 'Orchestrate door-to-door multi-modal shipments under a single pane of glass. Direct API dispatch coordinates long hauls across road, rail, air, and sea.',
      metricLabel: 'Active Orchestration API',
      metricVal: 'DISPATCH_OK (200)',
      telemetry: [
        { label: 'Total active shipments', val: '31 In Transit' },
        { label: 'Integrated carriers connected', val: '450+ Active' },
        { label: 'Map status', val: 'AIS/GPS Latencies < 5s' },
      ],
    },
    {
      id: 'rfq',
      title: 'Auto-RFQ Engine',
      icon: '⚡',
      desc: 'Stop emailing Excel sheets. Input cargo profile, and watch pre-approved vendors compete in real-time. System matches lowest rates in under 15 minutes.',
      metricLabel: 'Bidding System Telemetry',
      metricVal: 'AUTO_MATCHING_ACTIVE',
      telemetry: [
        { label: 'Average bidding rounds', val: '3 Rounds (Instant)' },
        { label: 'Smart matching efficiency', val: '98.6% Accuracy' },
        { label: 'Savings overhead', val: '14.2% Freight Delta' },
      ],
    },
    {
      id: 'routing',
      title: 'Smart Fleet Routing',
      icon: '🗺️',
      desc: 'Automated GPS route optimization. Our platform analyzes live congestion data, road bans, toll structures, and weather events to dispatch the most optimal road transport lanes.',
      metricLabel: 'Optimal Routing Engine',
      metricVal: 'ROUTING_SUCCESS',
      telemetry: [
        { label: 'Congestion detour index', val: '-18% Travel Time' },
        { label: 'Active geofence triggers', val: '4,200+ Nodes' },
        { label: 'E-Way bill linkage status', val: 'Direct API Linked' },
      ],
    },
    {
      id: 'customs',
      title: 'ICEGATE Auto-Customs',
      icon: '📜',
      desc: 'Automated Customs House Agent desk. Direct electronic file transfers of shipping bills and Bills of Entry, AI-powered HSN code lookups, and instant duty estimations.',
      metricLabel: 'ICEGATE Customs Desk Ping',
      metricVal: 'GATEWAY_OOC_CONNECTED',
      telemetry: [
        { label: 'Average duty assessment time', val: '< 10 minutes' },
        { label: 'Compliance validation score', val: '100% Verified' },
        { label: 'Out of Charge (OOC) trigger', val: 'Instant on assessment' },
      ],
    },
  ];

  const currentPillar = pillars.find((p) => p.id === activePillar);

  return (
    <div className="bg-white min-h-screen overflow-hidden">
      {/* ═══ HERO ═══ */}
      <section className="relative pt-32 pb-32 lg:pt-48 lg:pb-48 bg-gradient-to-br from-teal-50/20 via-slate-50 to-indigo-50/20 text-slate-900 overflow-hidden">
        {/* Sleek Smart Routing Background */}
        <div className="absolute inset-0 z-0">
          <img
            src="/assets/smart_routing.png"
            alt="Futuristic Grid Routing Background"
            className="w-full h-full object-cover object-center opacity-[0.06] scale-105 mix-blend-multiply animate-[pulse_20s_ease-in-out_infinite_alternate]"
          />
          <div className="absolute inset-0 bg-gradient-to-t from-white via-white/40 to-slate-50/80"></div>
          <div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-full max-w-7xl h-full pointer-events-none opacity-40">
            <div className="absolute -top-[10%] right-[-5%] w-[450px] h-[450px] rounded-full bg-brand-teal/15 blur-3xl"></div>
            <div className="absolute bottom-[-10%] left-[-5%] w-[550px] h-[550px] rounded-full bg-brand-indigo/10 blur-3xl"></div>
          </div>
        </div>

        <div className="max-w-7xl mx-auto px-4 sm:px-6 relative z-10">
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-12 items-center">
            <Reveal>
              <div>
                <div className="inline-flex items-center justify-center px-4 py-1.5 mb-6 text-xs font-bold text-brand-indigo bg-brand-indigo/10 rounded-full border border-brand-indigo/20 tracking-widest uppercase">
                  <span className="mr-2">🖥️</span> Logistics Operating System
                </div>
                <h1 className="text-5xl md:text-6xl lg:text-7xl font-extrabold tracking-tight mb-6 leading-tight text-slate-900">
                  The Multi-Modal <br />
                  <span className="text-transparent bg-clip-text bg-gradient-to-r from-brand-teal to-brand-indigo">
                    Command Center
                  </span>
                </h1>
                <p className="text-xl text-slate-600 max-w-2xl font-light leading-relaxed mb-10">
                  Unifying Road, Sea, Air, and smart warehouse logistics into one seamless digital operating system. Ship smarter, discover rates, and track everything instantly.
                </p>

                <div className="flex flex-wrap gap-4">
                  <Link
                    to="/signup"
                    className="px-8 py-4 bg-brand-teal hover:bg-teal-400 text-white font-bold rounded-full transition-colors shadow-[0_0_20px_rgba(0,191,165,0.4)]"
                  >
                    Start Free Platform Trial
                  </Link>
                  <a
                    href="#platform-pillars-demo"
                    className="px-8 py-4 bg-white/80 border border-slate-200 text-slate-700 font-bold rounded-full hover:bg-slate-100 hover:border-slate-300 transition-colors shadow-sm backdrop-blur-md"
                  >
                    Explore Core Pillars
                  </a>
                </div>
              </div>
            </Reveal>

            {/* Visual Hero Element - Simulated Control Center Shipment Dashboard */}
            <Reveal delay="delay-200">
              <div className="relative h-[420px] bg-white/80 rounded-3xl border border-slate-200/80 backdrop-blur-md p-6 shadow-2xl flex flex-col justify-between overflow-hidden">
                <div className="absolute -top-32 -right-32 w-64 h-64 bg-brand-indigo/10 rounded-full blur-[80px]"></div>

                <div className="flex justify-between items-center relative z-10">
                  <div>
                    <div className="text-sm text-slate-500 mb-1">Freel Command OS</div>
                    <div className="text-slate-800 font-bold text-xl flex items-center gap-2">
                      <span className="w-2.5 h-2.5 rounded-full bg-green-500 animate-pulse"></span> Systems Operational
                    </div>
                  </div>
                  <div className="text-right">
                    <div className="text-sm text-slate-500 mb-1">Global Coverage</div>
                    <div className="text-slate-800 font-bold text-xl font-mono">100% Sync</div>
                  </div>
                </div>

                {/* Simulated Shipment telemetry bars */}
                <div className="relative z-10 py-6 space-y-4">
                  <div>
                    <div className="flex justify-between text-xs font-bold text-slate-600 mb-1">
                      <span>ROAD FREIGHT (FTL/LTL)</span>
                      <span>14 Active Trucks</span>
                    </div>
                    <div className="h-1.5 bg-slate-200 rounded-full overflow-hidden">
                      <div className="h-full bg-gradient-to-r from-amber-400 to-amber-500 rounded-full w-[80%]"></div>
                    </div>
                  </div>

                  <div>
                    <div className="flex justify-between text-xs font-bold text-slate-600 mb-1">
                      <span>SMART ROUTING & GEOFENCING</span>
                      <span>8 Optimized Lanes</span>
                    </div>
                    <div className="h-1.5 bg-slate-200 rounded-full overflow-hidden">
                      <div className="h-full bg-gradient-to-r from-brand-indigo to-brand-teal rounded-full w-[65%]"></div>
                    </div>
                  </div>

                  <div>
                    <div className="flex justify-between text-xs font-bold text-slate-600 mb-1">
                      <span>AIR & SEA INT. CARGO</span>
                      <span>11 Containers / Flights</span>
                    </div>
                    <div className="h-1.5 bg-slate-200 rounded-full overflow-hidden">
                      <div className="h-full bg-gradient-to-r from-teal-400 to-teal-500 rounded-full w-[45%]"></div>
                    </div>
                  </div>
                </div>

                <div className="grid grid-cols-2 gap-4 relative z-10 border-t border-slate-200/50 pt-4">
                  <div className="bg-slate-50/80 border border-slate-100 rounded-xl p-3">
                    <div className="text-xs text-slate-500">Bidding Activity</div>
                    <div className="text-md font-bold text-slate-800">42 Live RFQs</div>
                  </div>
                  <div className="bg-slate-50/80 border border-slate-100 rounded-xl p-3">
                    <div className="text-xs text-slate-500">ICEGATE Clearances</div>
                    <div className="text-md font-bold text-green-600">Auto-Pass Active</div>
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
            {platformStats.map((s, i) => (
              <div key={i} className="text-center group">
                <div className="text-3xl md:text-4xl font-black text-transparent bg-clip-text bg-gradient-to-br from-brand-teal to-brand-indigo mb-2 group-hover:scale-105 transition-transform">
                  {s.value}
                </div>
                <div className="text-sm font-bold uppercase tracking-wider text-slate-700 mb-1">{s.label}</div>
                <div className="text-xs text-slate-400 px-2">{s.desc}</div>
              </div>
            ))}
          </div>
        </section>
      </Reveal>

      {/* ═══ INTERACTIVE PILLARS SIMULATOR ═══ */}
      <section id="platform-pillars-demo" className="py-32 bg-brand-navy text-white relative overflow-hidden">
        <div className="absolute inset-0 bg-[url('data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iNjAiIGhlaWdodD0iNjAiIHZpZXdCb3g9IjAgMCA2MCA2MCIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj48ZyBmaWxsPSJub25lIiBmaWxsLXJ1bGU9ImV2ZW5vZGQiPjxwYXRoIGQ9Ik0zNiAzNHYtNGgtdjRoLTR2NGgtdjRoNHY0aDR2LTRoNHptMC0xMnYtNGgtdjRoLTR2NGgtdjRoNHY0aDR2LTRoNHptLTExIDEwaC00djRoLTR2LTRoLTR2LTRoNHYtNGg0djRoNHY0em0tMTEgMTBoLTR2NGgtdjRoLTR2LTRoNHYtNGg0djRoNHY0em0xMSAwaC00djRoLTR2LTRoLTR2LTRoNHYtNGg0djRoNHY0em0xMSAwaC00djRoLTR2LTRoLTR2LTRoNHYtNGg0djRoNHY0eiIgZmlsbD0iI2ZmZmZmZiIgZmlsbC1vcGFjaXR5PSIwLjAyIi8+PC9nPjwvc3ZnPg==')] opacity-50 z-0"></div>

        <div className="max-w-7xl mx-auto px-4 sm:px-6 relative z-10">
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-16 items-center">
            <Reveal>
              <div>
                <span className="text-brand-teal font-bold tracking-widest uppercase text-sm mb-3 block">
                  Console Telemetry
                </span>
                <h2 className="text-4xl md:text-5xl font-extrabold mb-6">Four Pillars of Freel OS</h2>
                <p className="text-slate-300 text-lg mb-8 leading-relaxed">
                  Our system combines ground operational capabilities with modern cloud automation. Click on each pillar below to sync your screen and see active operations metrics.
                </p>

                {/* Pillar Buttons */}
                <div className="space-y-4">
                  {pillars.map((p, idx) => (
                    <button
                      key={idx}
                      onClick={() => setActivePillar(p.id)}
                      className={`w-full text-left p-5 rounded-2xl border transition-all ${
                        activePillar === p.id
                          ? 'bg-white/10 border-brand-teal shadow-[0_0_15px_rgba(0,191,165,0.2)]'
                          : 'bg-transparent border-white/10 hover:bg-white/5'
                      }`}
                    >
                      <div className="flex justify-between items-center">
                        <div className="flex items-center gap-4">
                          <span className="text-2xl">{p.icon}</span>
                          <span className="font-bold text-white text-md">{p.title}</span>
                        </div>
                        <span className="text-xs font-mono bg-white/15 text-brand-teal px-2 py-0.5 rounded-full">
                          Pillar 0{idx + 1}
                        </span>
                      </div>
                    </button>
                  ))}
                </div>
              </div>
            </Reveal>

            {/* Simulated Live Console Dashboard */}
            <Reveal delay="delay-200">
              <div className="bg-[#0f172a] rounded-[2.5rem] p-8 border border-slate-800 shadow-2xl relative overflow-hidden">
                <div className="absolute top-1/2 left-1/2 w-[140%] h-[140%] -translate-x-1/2 -translate-y-1/2 bg-[conic-gradient(from_0deg,transparent_0_340deg,rgba(0,191,165,0.1)_360deg)] rounded-full animate-[spin_10s_linear_infinite] z-0 mix-blend-screen pointer-events-none"></div>

                <div className="relative z-10">
                  <div className="flex justify-between items-center border-b border-slate-800 pb-4 mb-8">
                    <div className="font-mono text-brand-teal text-xs">{currentPillar.metricLabel}</div>
                    <div className="text-xs bg-brand-teal/20 text-brand-teal px-3 py-1 rounded-full font-mono font-bold">
                      {currentPillar.metricVal}
                    </div>
                  </div>

                  <h3 className="text-2xl font-black mb-4 flex items-center gap-3 text-white">
                    <span>{currentPillar.icon}</span> {currentPillar.title}
                  </h3>
                  <p className="text-slate-400 text-sm leading-relaxed mb-8">{currentPillar.desc}</p>

                  <div className="border-t border-slate-800/80 pt-6 space-y-4">
                    <div className="text-xs font-bold text-slate-500 uppercase tracking-wider font-mono">
                      Active Telemetry Outputs
                    </div>
                    {currentPillar.telemetry.map((t, idx) => (
                      <div key={idx} className="flex justify-between items-center bg-slate-900/60 border border-slate-800/50 p-4 rounded-xl font-mono text-sm">
                        <span className="text-slate-500">{t.label}</span>
                        <span className="text-white font-bold">{t.val}</span>
                      </div>
                    ))}
                  </div>

                  <div className="mt-8 flex gap-3">
                    <button className="flex-1 py-3.5 bg-white/5 border border-white/10 hover:bg-white/10 text-white text-xs font-bold rounded-xl transition-all font-mono">
                      📄 API Documentation
                    </button>
                    <Link
                      to="/signup"
                      className="flex-1 py-3.5 bg-brand-teal text-slate-900 text-xs font-bold rounded-xl hover:bg-teal-400 transition-all font-mono text-center flex items-center justify-center"
                    >
                      Initialize Sandbox
                    </Link>
                  </div>
                </div>
              </div>
            </Reveal>
          </div>
        </div>
      </section>

      {/* ═══ SPLIT GRID FEATURE 1 (LIVE GPS TELEMETRY & GEOFENCING) ═══ */}
      <section className="py-32 bg-slate-50 relative overflow-hidden border-t border-b border-slate-200/50">
        <div className="max-w-7xl mx-auto px-4 sm:px-6">
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-16 items-center">
            
            <Reveal>
              <div className="relative rounded-[3rem] overflow-hidden shadow-2xl border border-slate-200 group h-[580px]">
                <div className="absolute inset-0 bg-brand-teal/10 group-hover:bg-transparent transition-colors z-10 mix-blend-overlay duration-700"></div>
                <img
                  src="/assets/smart_routing.png"
                  alt="Live GPS Telemetry and Smart Geofenced Borders"
                  className="w-full h-full object-cover transform group-hover:scale-105 transition-transform duration-1000"
                />
                <div className="absolute bottom-0 left-0 w-full p-8 bg-gradient-to-t from-slate-900 via-slate-900/60 to-transparent z-20">
                  <div className="text-brand-teal font-mono text-xs mb-1 uppercase tracking-widest">
                    IoT Fleet Tracking Gateway
                  </div>
                  <div className="text-white font-bold text-2xl">
                    Live GPS Telemetry & Geofenced Alerts
                  </div>
                </div>
              </div>
            </Reveal>

            <div>
              <Reveal>
                <div className="mb-10">
                  <span className="text-brand-indigo font-bold tracking-widest uppercase text-sm mb-3 block">
                    Advanced Fleet Control
                  </span>
                  <h2 className="text-4xl md:text-5xl font-extrabold text-brand-navy mb-4 leading-tight">
                    Smart GPS Tracking & Automated Geofencing
                  </h2>
                  <p className="text-slate-600 text-lg leading-relaxed mb-6">
                    Our platform links digital manifests directly with live vehicle hardware. By deploying smart virtual borders (geofences) around major warehouses, ports, and state checkpoints, we trigger automatic loading logs and arrival updates as soon as the truck enters the boundary.
                  </p>
                  <p className="text-slate-500 text-sm leading-relaxed">
                    This completely eliminates checkpost delays and manual driver check-in calls. In the event of heavy traffic or highway blocks, our routing engine automatically recalculates detours in real time, communicating changes instantly to shippers.
                  </p>
                </div>
              </Reveal>

              <div className="grid grid-cols-2 gap-4">
                <div className="p-5 bg-white border border-slate-200/60 rounded-2xl shadow-sm">
                  <span className="text-2xl mb-2 block">⏱️</span>
                  <h4 className="font-bold text-slate-800 text-md mb-1">Auto-Geofencing</h4>
                  <p className="text-slate-500 text-xs">Instant ETA adjustments at shipping checkpoints.</p>
                </div>
                <div className="p-5 bg-white border border-slate-200/60 rounded-2xl shadow-sm">
                  <span className="text-2xl mb-2 block">🌱</span>
                  <h4 className="font-bold text-slate-800 text-md mb-1">Route Optimizations</h4>
                  <p className="text-slate-500 text-xs">18% travel time reduction via smart congestion detours.</p>
                </div>
              </div>
            </div>
            
          </div>
        </div>
      </section>

      {/* ═══ SPLIT GRID FEATURE 2 (SaaS DIGITAL OS DASHBOARD) ═══ */}
      <section className="py-32 bg-white relative overflow-hidden">
        <div className="max-w-7xl mx-auto px-4 sm:px-6">
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-16 items-center lg:flex-row-reverse">
            
            <div className="lg:order-2">
              <Reveal>
                <div className="relative rounded-[3rem] overflow-hidden shadow-2xl border border-slate-200 group h-[580px]">
                  <div className="absolute inset-0 bg-brand-indigo/10 group-hover:bg-transparent transition-colors z-10 mix-blend-overlay duration-700"></div>
                  <img
                    src="/assets/digital_platform_dashboard.png"
                    alt="Digital Freight Command Dashboard"
                    className="w-full h-full object-cover transform group-hover:scale-105 transition-transform duration-1000"
                  />
                  <div className="absolute bottom-0 left-0 w-full p-8 bg-gradient-to-t from-slate-900 via-slate-900/60 to-transparent z-20">
                    <div className="text-brand-teal font-mono text-xs mb-1 uppercase tracking-widest">
                      Freel Dashboard Software
                    </div>
                    <div className="text-white font-bold text-2xl">
                      Automated Analytics & Discovery
                    </div>
                  </div>
                </div>
              </Reveal>
            </div>

            <div className="lg:order-1">
              <Reveal>
                <div className="mb-10">
                  <span className="text-brand-indigo font-bold tracking-widest uppercase text-sm mb-3 block">
                    Digital Operating System
                  </span>
                  <h2 className="text-4xl md:text-5xl font-extrabold text-brand-navy mb-4 leading-tight">
                    Smart Analytics & Single Pane of Glass
                  </h2>
                  <p className="text-slate-600 text-lg leading-relaxed mb-6">
                    Managing fragmented logistics providers requires smart dashboards. Freel aggregates rate sheets, transit lanes, document compliance, and financial audits into one powerful software window.
                  </p>
                  <p className="text-slate-500 text-sm leading-relaxed">
                    With automated invoicing audits and digital E-POD wallets, your logistics desk spends less time resolving paperwork discrepancies and more time shipping. Our rate comparative metrics offer full transparency with no hidden margins.
                  </p>
                </div>
              </Reveal>

              <div className="grid grid-cols-2 gap-4">
                <div className="p-5 bg-slate-50 border border-slate-200/60 rounded-2xl">
                  <span className="text-2xl mb-2 block">📈</span>
                  <h4 className="font-bold text-slate-800 text-md mb-1">Instant Bids</h4>
                  <p className="text-slate-500 text-xs">Vendor comparison tables compiled in under 15 minutes.</p>
                </div>
                <div className="p-5 bg-slate-50 border border-slate-200/60 rounded-2xl">
                  <span className="text-2xl mb-2 block">📁</span>
                  <h4 className="font-bold text-slate-800 text-md mb-1">Document Wallet</h4>
                  <p className="text-slate-500 text-xs">GST, PAN, E-Way bills and E-PODs securely vaulted.</p>
                </div>
              </div>
            </div>
            
          </div>
        </div>
      </section>

      {/* ═══ CTA SECTION ═══ */}
      <section className="py-24 bg-slate-50 text-center border-t border-slate-200/50">
        <Reveal>
          <div className="max-w-3xl mx-auto px-4">
            <h2 className="text-4xl md:text-5xl font-black text-brand-navy mb-6">Ready to Command Your Logistics?</h2>
            <p className="text-lg text-slate-600 mb-10">
              Set up your free sandbox in minutes. Connect your pre-existing cargo profiles, test the live ICEGATE and automated geofencing tracking APIs, and experience the power of synchronized shipping.
            </p>
            <Link
              to="/signup"
              className="inline-flex items-center gap-2 px-10 py-5 bg-gradient-to-r from-brand-teal to-brand-indigo text-white font-bold text-lg rounded-full hover:shadow-[0_10px_30px_rgba(0,191,165,0.3)] transition-all hover:-translate-y-1"
            >
              Request Platform Demo <span>🖥️</span>
            </Link>
          </div>
        </Reveal>
      </section>
    </div>
  );
}

import { useEffect, useRef, useState } from 'react';
import { Link } from 'react-router-dom';

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

const stats = [
  { value: '150+', label: 'Airports Worldwide' },
  { value: '24hr', label: 'Express Delivery' },
  { value: 'DG', label: 'Certified Handler' },
  { value: 'Live', label: 'FlightAware Tracking' },
];

const capabilities = [
  { icon: '📦', title: 'General Cargo', desc: 'Standard air freight for electronics, textiles, auto parts, and consumer goods. Consolidation available.', color: 'from-blue-400 to-blue-600' },
  { icon: '☣️', title: 'Dangerous Goods (DG)', desc: 'IATA DGR certified. We handle Class 1-9 hazardous materials with full MSDS documentation.', color: 'from-red-400 to-red-600' },
  { icon: '💊', title: 'Pharma & Temperature', desc: 'GDP-compliant cold chain for pharmaceuticals, biologics. Active & passive temp control.', color: 'from-teal-400 to-teal-600' },
  { icon: '🍎', title: 'Perishable Cargo', desc: 'Fresh produce, flowers, seafood with priority handling. Cool chain integrity to destination.', color: 'from-green-400 to-green-600' },
];

/* ═══════════════════════════════════════════ */
/*             AIR FREIGHT PAGE               */
/* ═══════════════════════════════════════════ */
export default function AirFreight() {
  const [activeRoute, setActiveRoute] = useState(0);
  
  const routes = [
    { origin: 'DEL', dest: 'JFK', time: '14h 30m', status: 'In Air', progress: '65%' },
    { origin: 'BOM', dest: 'LHR', time: '9h 15m', status: 'Departed', progress: '10%' },
    { origin: 'BLR', dest: 'SIN', time: '4h 20m', status: 'Arriving', progress: '90%' },
  ];

  return (
    <div className="bg-white min-h-screen overflow-hidden">
      {/* ═══ HERO ═══ */}
      <section className="relative pt-32 pb-32 lg:pt-48 lg:pb-48 bg-gradient-to-br from-teal-50/30 via-slate-50 to-indigo-50/30 text-slate-900 overflow-hidden">
        {/* Photorealistic Background */}
        <div className="absolute inset-0 z-0">
          <img src="/assets/air_freight_hero.png" alt="Cargo Plane Taking Off" className="w-full h-full object-cover object-center opacity-[0.06] scale-105 mix-blend-multiply animate-[pulse_15s_ease-in-out_infinite_alternate]" />
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
                  <span className="mr-2">✈️</span> Global Air Cargo
                </div>
                <h1 className="text-5xl md:text-6xl lg:text-7xl font-extrabold tracking-tight mb-6 leading-tight text-slate-900">
                  Time-Critical <br />
                  <span className="text-transparent bg-clip-text bg-gradient-to-r from-blue-600 to-brand-teal">
                    Deliveries
                  </span>
                </h1>
                <p className="text-xl text-slate-600 max-w-2xl font-light leading-relaxed mb-10">
                  When time is money, fly your cargo with confidence. Direct access to 150+ airports, guaranteed space on premium carriers, and live FlightAware tracking.
                </p>
                
                <div className="flex flex-wrap gap-4">
                  <Link to="/contact" className="px-8 py-4 bg-brand-teal text-white font-bold rounded-full hover:bg-teal-400 transition-colors shadow-[0_0_20px_rgba(0,191,165,0.4)]">
                    Get Air Freight Quote
                  </Link>
                  <a href="#tracking-demo" className="px-8 py-4 bg-white/80 border border-slate-200 text-slate-700 font-bold rounded-full hover:bg-slate-100 hover:border-slate-300 transition-colors shadow-sm backdrop-blur-md">
                    View Tracking Demo
                  </a>
                </div>
              </div>
            </Reveal>

            {/* Visual Hero Element */}
            <Reveal delay="delay-200">
              <div className="relative h-[400px] bg-white/80 rounded-3xl border border-slate-200/80 backdrop-blur-md p-6 shadow-2xl flex flex-col justify-between overflow-hidden">
                <div className="absolute -top-32 -right-32 w-64 h-64 bg-brand-indigo/10 rounded-full blur-[80px]"></div>
                
                <div className="flex justify-between items-center relative z-10">
                  <div>
                    <div className="text-sm text-slate-500 mb-1">Status</div>
                    <div className="text-slate-800 font-bold text-xl flex items-center gap-2"><span className="w-2 h-2 rounded-full bg-green-500 animate-pulse"></span> In Transit</div>
                  </div>
                  <div className="text-right">
                    <div className="text-sm text-slate-500 mb-1">Flight</div>
                    <div className="text-slate-800 font-bold text-xl font-mono">EK-982</div>
                  </div>
                </div>

                <div className="relative z-10 py-10">
                  <div className="flex justify-between text-3xl font-black mb-2 text-slate-800">
                    <span>DEL</span>
                    <span className="text-brand-indigo">✈️</span>
                    <span>DXB</span>
                  </div>
                  
                  <div className="relative h-2 bg-slate-200 rounded-full mt-4">
                    <div className="absolute top-0 left-0 h-full bg-gradient-to-r from-brand-teal to-blue-500 rounded-full w-[65%]">
                      <div className="absolute -right-2 -top-1 w-4 h-4 bg-white border border-slate-300 rounded-full shadow-[0_0_10px_rgba(0,0,0,0.15)]"></div>
                    </div>
                  </div>
                  <div className="flex justify-between text-xs text-slate-500 mt-2 font-mono">
                    <span>Indira Gandhi Intl</span>
                    <span>Dubai Intl</span>
                  </div>
                </div>

                <div className="grid grid-cols-2 gap-4 relative z-10">
                  <div className="bg-slate-50/80 border border-slate-100 rounded-xl p-3">
                    <div className="text-xs text-slate-500">Est. Arrival</div>
                    <div className="text-lg font-bold text-slate-800">14:30 GMT</div>
                  </div>
                  <div className="bg-slate-50/80 border border-slate-100 rounded-xl p-3">
                    <div className="text-xs text-slate-500">Cargo Type</div>
                    <div className="text-lg font-bold text-slate-800">Pharma (Cold)</div>
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
          <div className="bg-white rounded-[2rem] p-8 shadow-[0_20px_50px_-10px_rgba(0,0,0,0.1)] border border-slate-100 grid grid-cols-2 md:grid-cols-4 gap-6 divide-x-0 md:divide-x divide-slate-100">
            {stats.map((s, i) => (
              <div key={i} className="text-center group">
                <div className="text-3xl md:text-4xl font-black text-transparent bg-clip-text bg-gradient-to-br from-blue-500 to-brand-teal mb-2 group-hover:scale-110 transition-transform">{s.value}</div>
                <div className="text-sm font-bold uppercase tracking-wider text-slate-500">{s.label}</div>
              </div>
            ))}
          </div>
        </section>
      </Reveal>

      {/* ═══ TARMAC OPERATIONS & WHAT WE HANDLE ═══ */}
      <section className="py-32 bg-slate-50 relative">
        <div className="max-w-7xl mx-auto px-4 sm:px-6">
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-16 items-center">
            
            <Reveal>
              <div className="relative rounded-[3rem] overflow-hidden shadow-2xl group">
                <div className="absolute inset-0 bg-blue-500/20 group-hover:bg-transparent transition-colors z-10 mix-blend-overlay duration-700"></div>
                <img src="/assets/air_cargo_operations.png" alt="Air Cargo Loading Operations" className="w-full h-auto object-cover transform group-hover:scale-105 transition-transform duration-1000" />
                <div className="absolute bottom-0 left-0 w-full p-8 bg-gradient-to-t from-[#0a192f] to-transparent z-20">
                  <div className="text-blue-400 font-mono text-sm mb-1">TARMAC OPERATIONS</div>
                  <div className="text-white font-bold text-xl">Specialized Pallet Loading</div>
                </div>
              </div>
            </Reveal>

            <div>
              <Reveal>
                <div className="mb-10">
                  <span className="text-blue-500 font-bold tracking-widest uppercase text-sm mb-3 block">Specialized Cargo</span>
                  <h2 className="text-4xl md:text-5xl font-extrabold text-brand-navy mb-4 leading-tight">We Fly Everything.</h2>
                  <p className="text-slate-600 text-lg">
                    From sensitive pharmaceuticals to oversized machinery, our air freight network is equipped to handle any cargo type with precision directly on the tarmac.
                  </p>
                </div>
              </Reveal>

              <div className="grid grid-cols-1 sm:grid-cols-2 gap-6">
                {capabilities.map((c, i) => (
                  <Reveal key={i} delay={`delay-${(i % 2) * 100}`}>
                    <div className="bg-white rounded-3xl p-6 border border-slate-200 hover:border-blue-300 hover:shadow-xl transition-all group h-full">
                      <div className={`w-12 h-12 rounded-2xl bg-gradient-to-br ${c.color} flex items-center justify-center text-2xl text-white shadow-lg mb-4 group-hover:scale-110 transition-transform`}>
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

      {/* ═══ LIVE TRACKING INTERACTIVE ═══ */}
      <section id="tracking-demo" className="py-32 bg-brand-navy text-white relative overflow-hidden">
        <div className="absolute inset-0 bg-[url('data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iNjAiIGhlaWdodD0iNjAiIHZpZXdCb3g9IjAgMCA2MCA2MCIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj48ZyBmaWxsPSJub25lIiBmaWxsLXJ1bGU9ImV2ZW5vZGQiPjxwYXRoIGQ9Ik0zNiAzNHYtNGgtdjRoLTR2NGgtdjRoNHY0aDR2LTRoNHptMC0xMnYtNGgtdjRoLTR2NGgtdjRoNHY0aDR2LTRoNHptLTExIDEwaC00djRoLTR2LTRoLTR2LTRoNHYtNGg0djRoNHY0em0tMTEgMTBoLTR2NGgtdjRoLTR2LTRoNHYtNGg0djRoNHY0em0xMSAwaC00djRoLTR2LTRoLTR2LTRoNHYtNGg0djRoNHY0em0xMSAwaC00djRoLTR2LTRoLTR2LTRoNHYtNGg0djRoNHY0eiIgZmlsbD0iI2ZmZmZmZiIgZmlsbC1vcGFjaXR5PSIwLjAyIi8+PC9nPjwvc3ZnPg==')] opacity-50 z-0"></div>
        
        <div className="max-w-7xl mx-auto px-4 sm:px-6 relative z-10">
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-16 items-center">
            <Reveal>
              <div>
                <h2 className="text-4xl md:text-5xl font-extrabold mb-6">Real-Time Radar.</h2>
                <p className="text-slate-300 text-lg mb-8 leading-relaxed">
                  We integrate directly with FlightAware to give you granular, second-by-second updates on your air cargo. No more calling airlines for ETAs.
                </p>
                
                <div className="space-y-4">
                  {routes.map((route, idx) => (
                    <button 
                      key={idx}
                      onClick={() => setActiveRoute(idx)}
                      className={`w-full text-left p-4 rounded-2xl border transition-all ${activeRoute === idx ? 'bg-white/10 border-blue-400 shadow-[0_0_15px_rgba(96,165,250,0.2)]' : 'bg-transparent border-white/10 hover:bg-white/5'}`}
                    >
                      <div className="flex justify-between items-center">
                        <div className="flex items-center gap-4">
                          <div className={`w-3 h-3 rounded-full ${route.status === 'In Air' ? 'bg-green-400 animate-pulse' : route.status === 'Arriving' ? 'bg-yellow-400' : 'bg-blue-400'}`}></div>
                          <span className="font-bold font-mono">{route.origin} → {route.dest}</span>
                        </div>
                        <span className="text-sm text-slate-400">{route.status}</span>
                      </div>
                    </button>
                  ))}
                </div>
              </div>
            </Reveal>

            <Reveal delay="delay-200">
              <div className="bg-[#0f172a] rounded-[2rem] p-8 border border-slate-700 shadow-2xl relative overflow-hidden">
                {/* Radar sweep effect */}
                <div className="absolute top-1/2 left-1/2 w-[150%] h-[150%] -translate-x-1/2 -translate-y-1/2 bg-[conic-gradient(from_0deg,transparent_0_340deg,rgba(59,130,246,0.3)_360deg)] rounded-full animate-[spin_4s_linear_infinite] z-0 mix-blend-screen pointer-events-none"></div>
                
                {/* Radar circles */}
                <div className="absolute top-1/2 left-1/2 w-[100%] h-[100%] -translate-x-1/2 -translate-y-1/2 border border-slate-700 rounded-full z-0"></div>
                <div className="absolute top-1/2 left-1/2 w-[70%] h-[70%] -translate-x-1/2 -translate-y-1/2 border border-slate-700 rounded-full z-0"></div>
                <div className="absolute top-1/2 left-1/2 w-[40%] h-[40%] -translate-x-1/2 -translate-y-1/2 border border-slate-700 rounded-full z-0"></div>

                <div className="relative z-10">
                  <div className="flex justify-between items-center border-b border-slate-700 pb-4 mb-8">
                    <div className="font-mono text-blue-400">FLIGHTAWARE_SYNC_OK</div>
                    <div className="text-xs bg-green-500/20 text-green-400 px-2 py-1 rounded">LIVE</div>
                  </div>

                  <div className="text-center mb-12">
                    <div className="text-6xl font-black text-white mb-2 tracking-tighter">
                      {routes[activeRoute].origin} <span className="text-blue-500">→</span> {routes[activeRoute].dest}
                    </div>
                    <div className="text-slate-400 font-mono">Flight Duration: {routes[activeRoute].time}</div>
                  </div>

                  <div className="relative">
                    <div className="h-1 w-full bg-slate-800 rounded-full relative">
                      {/* Interactive moving plane based on progress */}
                      <div 
                        className="absolute top-1/2 -translate-y-1/2 -translate-x-1/2 text-2xl text-white transition-all duration-1000 z-20"
                        style={{ left: routes[activeRoute].progress }}
                      >
                        ✈️
                      </div>
                      
                      {/* Trail */}
                      <div 
                        className="absolute top-0 left-0 h-full bg-blue-500 rounded-full transition-all duration-1000 z-10"
                        style={{ width: routes[activeRoute].progress }}
                      >
                        <div className="absolute right-0 top-1/2 -translate-y-1/2 w-4 h-4 bg-white rounded-full shadow-[0_0_10px_#fff]"></div>
                      </div>
                    </div>
                    
                    <div className="flex justify-between mt-4 text-xs font-bold text-slate-500 uppercase">
                      <div>Departed</div>
                      <div>{routes[activeRoute].status}</div>
                      <div>Arrival</div>
                    </div>
                  </div>
                </div>
              </div>
            </Reveal>
          </div>
        </div>
      </section>

      {/* ═══ CTA SECTION ═══ */}
      <section className="py-24 bg-white text-center">
        <Reveal>
          <div className="max-w-3xl mx-auto px-4">
            <h2 className="text-4xl md:text-5xl font-black text-brand-navy mb-6">Fly Your Cargo Today.</h2>
            <p className="text-lg text-slate-600 mb-10">
              Stop waiting days for air freight quotes. Compare rates instantly and book space on premium carriers.
            </p>
            <Link to="/contact" className="inline-flex items-center gap-2 px-10 py-5 bg-gradient-to-r from-blue-600 to-brand-teal text-white font-bold text-lg rounded-full hover:shadow-[0_10px_30px_rgba(0,191,165,0.3)] transition-all hover:-translate-y-1">
              Request Air Quote <span>✈️</span>
            </Link>
          </div>
        </Reveal>
      </section>

    </div>
  );
}

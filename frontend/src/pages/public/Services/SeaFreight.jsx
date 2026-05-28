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
  
  const vessels = [
    { name: 'MSC ANNA', route: 'Nhava Sheva → Rotterdam', status: 'In Transit', progress: '45%' },
    { name: 'MAERSK MC-KINNEY', route: 'Mundra → Jebel Ali', status: 'Docking', progress: '95%' },
    { name: 'CMA CGM JACQUES', route: 'Chennai → Singapore', status: 'Departed', progress: '15%' },
  ];

  return (
    <div className="bg-white min-h-screen overflow-hidden">
      
      {/* ═══ HERO ═══ */}
      <section className="relative pt-32 pb-32 lg:pt-48 lg:pb-48 bg-[#0a192f] text-white overflow-hidden">
        {/* Photorealistic Background */}
        <div className="absolute inset-0 z-0">
          <img src="/assets/sea_freight_hero.png" alt="Cargo Ship at Sunset" className="w-full h-full object-cover object-center opacity-40 scale-105 animate-[pulse_20s_ease-in-out_infinite_alternate]" />
          <div className="absolute inset-0 bg-gradient-to-t from-[#0a192f] via-transparent to-[#0a192f]/80"></div>
          <div className="absolute inset-0 bg-gradient-to-r from-[#0a192f] via-transparent to-transparent"></div>
        </div>

        <div className="max-w-7xl mx-auto px-4 sm:px-6 relative z-10">
          <div className="max-w-3xl">
            <Reveal>
              <div className="inline-flex items-center justify-center px-4 py-1.5 mb-6 text-xs font-bold text-cyan-300 bg-cyan-900/50 rounded-full border border-cyan-500/30 tracking-widest uppercase backdrop-blur-md">
                <span className="mr-2">🚢</span> Ocean Freight
              </div>
              <h1 className="text-5xl md:text-6xl lg:text-7xl font-extrabold tracking-tight mb-6 leading-tight drop-shadow-lg">
                Mastering the <br />
                <span className="text-transparent bg-clip-text bg-gradient-to-r from-cyan-400 to-brand-teal">
                  Global Oceans
                </span>
              </h1>
              <p className="text-xl text-slate-200 font-light leading-relaxed mb-10 drop-shadow-md">
                FCL and LCL shipments connecting 150+ ports worldwide. Leverage our massive carrier network for guaranteed space, competitive rates, and real-time AIS vessel tracking.
              </p>
              
              <div className="flex flex-wrap gap-4">
                <Link to="/contact" className="px-8 py-4 bg-brand-teal text-white font-bold rounded-full hover:bg-teal-400 transition-colors shadow-[0_0_20px_rgba(0,191,165,0.4)]">
                  Get Ocean Quote
                </Link>
                <a href="#ais-tracking" className="px-8 py-4 bg-white/10 border border-white/20 text-white font-bold rounded-full hover:bg-white/20 transition-colors backdrop-blur-md">
                  Explore Live Tracking
                </a>
              </div>
            </Reveal>
          </div>
        </div>
      </section>

      {/* ═══ STATS BAR (Floating) ═══ */}
      <Reveal>
        <section className="max-w-6xl mx-auto px-4 sm:px-6 relative z-20 -mt-20">
          <div className="bg-white/90 backdrop-blur-xl rounded-[2rem] p-8 shadow-[0_20px_50px_-10px_rgba(0,0,0,0.15)] border border-white grid grid-cols-2 md:grid-cols-4 gap-6 divide-x-0 md:divide-x divide-slate-200">
            {stats.map((s, i) => (
              <div key={i} className="text-center group">
                <div className="text-3xl md:text-4xl font-black text-transparent bg-clip-text bg-gradient-to-br from-cyan-500 to-brand-teal mb-2 group-hover:scale-110 transition-transform">{s.value}</div>
                <div className="text-sm font-bold uppercase tracking-wider text-slate-500">{s.label}</div>
              </div>
            ))}
          </div>
        </section>
      </Reveal>

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

      {/* ═══ SHIPPING LINES CAROUSEL ═══ */}
      <section className="py-16 bg-slate-50 border-y border-slate-200 overflow-hidden">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 mb-8 text-center">
          <p className="text-sm font-bold text-slate-400 uppercase tracking-widest">Partnered with the World's Best Carriers</p>
        </div>
        <div className="flex space-x-8 animate-[fly_40s_linear_infinite] whitespace-nowrap opacity-60">
          {/* Double the array for seamless infinite scroll effect */}
          {[...shippingLines, ...shippingLines, ...shippingLines].map((line, i) => (
            <div key={i} className="text-3xl font-black text-slate-300 mx-8 tracking-tighter hover:text-brand-teal transition-colors cursor-default">
              {line}
            </div>
          ))}
        </div>
      </section>

      {/* ═══ LIVE AIS TRACKING INTERACTIVE ═══ */}
      <section id="ais-tracking" className="py-32 bg-[#0a192f] text-white relative overflow-hidden">
        {/* Ocean topological map background pattern */}
        <div className="absolute inset-0 opacity-10 bg-[radial-gradient(ellipse_at_center,_var(--tw-gradient-stops))] from-cyan-900 via-[#0a192f] to-[#0a192f] z-0"></div>
        
        <div className="max-w-7xl mx-auto px-4 sm:px-6 relative z-10">
          <Reveal>
            <div className="text-center mb-16">
              <h2 className="text-4xl md:text-5xl font-extrabold mb-6">Satellite AIS Tracking.</h2>
              <p className="text-slate-300 text-lg max-w-2xl mx-auto">
                Stop guessing where your containers are. Our platform integrates with satellite AIS data to show you exactly where your vessel is on the open ocean.
              </p>
            </div>
          </Reveal>

          <div className="grid grid-cols-1 lg:grid-cols-3 gap-8 items-start">
            
            {/* Vessel List */}
            <div className="space-y-4 lg:col-span-1">
              {vessels.map((vessel, idx) => (
                <Reveal key={idx} delay={`delay-${idx * 100}`}>
                  <button 
                    onClick={() => setActiveVessel(idx)}
                    className={`w-full text-left p-6 rounded-3xl border transition-all ${activeVessel === idx ? 'bg-cyan-900/40 border-cyan-400 shadow-[0_0_25px_rgba(34,211,238,0.2)]' : 'bg-white/5 border-white/10 hover:bg-white/10'}`}
                  >
                    <div className="text-xs text-cyan-400 mb-1 font-mono tracking-widest">{vessel.name}</div>
                    <div className="font-bold text-lg mb-4">{vessel.route}</div>
                    <div className="flex justify-between items-center text-sm text-slate-400">
                      <span className="flex items-center gap-2">
                        <span className={`w-2 h-2 rounded-full ${vessel.status === 'Docking' ? 'bg-yellow-400' : 'bg-green-400 animate-pulse'}`}></span>
                        {vessel.status}
                      </span>
                      <span>{vessel.progress} Complete</span>
                    </div>
                  </button>
                </Reveal>
              ))}
            </div>

            {/* Radar / Ocean Map View */}
            <div className="lg:col-span-2">
              <Reveal delay="delay-300">
                <div className="bg-[#050b14] rounded-[3rem] p-2 border border-slate-700 shadow-2xl relative overflow-hidden h-[400px]">
                  
                  {/* Fake Map Grid & Sonar Sweep */}
                  <div className="absolute inset-0 bg-[linear-gradient(rgba(255,255,255,0.05)_1px,transparent_1px),linear-gradient(90deg,rgba(255,255,255,0.05)_1px,transparent_1px)] bg-[size:40px_40px] z-0"></div>
                  <div className="absolute top-1/2 left-1/2 w-[200%] h-[200%] -translate-x-1/2 -translate-y-1/2 bg-[conic-gradient(from_0deg,transparent_0_340deg,rgba(34,211,238,0.15)_360deg)] rounded-full animate-[spin_8s_linear_infinite] z-0 mix-blend-screen pointer-events-none"></div>

                  {/* UI Overlay */}
                  <div className="absolute top-6 left-6 z-10">
                    <div className="font-mono text-cyan-400 text-sm mb-1 bg-black/50 px-3 py-1 rounded">SAT: INMARSAT-C</div>
                    <div className="text-xs bg-green-500/20 text-green-400 px-3 py-1 rounded inline-block border border-green-500/30">AIS SIGNAL LOCKED</div>
                  </div>

                  {/* The Vessel Visualization */}
                  <div className="absolute top-1/2 left-0 w-full px-12 -translate-y-1/2 z-10">
                    <div className="relative h-px w-full bg-slate-700">
                      
                      {/* Ocean Trail */}
                      <div 
                        className="absolute top-1/2 -translate-y-1/2 left-0 h-1 bg-cyan-500 rounded-full transition-all duration-1000"
                        style={{ width: vessels[activeVessel].progress }}
                      ></div>
                      
                      {/* The Ship Icon */}
                      <div 
                        className="absolute top-1/2 -translate-y-1/2 -translate-x-1/2 text-4xl transition-all duration-1000 z-20 drop-shadow-[0_0_15px_rgba(34,211,238,0.8)]"
                        style={{ left: vessels[activeVessel].progress }}
                      >
                        🚢
                      </div>

                      {/* Ripple Effect around ship */}
                      <div 
                        className="absolute top-1/2 -translate-y-1/2 -translate-x-1/2 w-16 h-16 border-2 border-cyan-400/50 rounded-full animate-ping transition-all duration-1000"
                        style={{ left: vessels[activeVessel].progress }}
                      ></div>
                    </div>
                  </div>

                  <div className="absolute bottom-6 left-0 w-full px-12 flex justify-between text-slate-500 font-mono text-xs z-10">
                    <div>{vessels[activeVessel].route.split('→')[0].trim()} (POL)</div>
                    <div>{vessels[activeVessel].route.split('→')[1].trim()} (POD)</div>
                  </div>

                </div>
              </Reveal>
            </div>
          </div>
        </div>
      </section>

      {/* ═══ CTA SECTION ═══ */}
      <section className="py-24 bg-brand-teal text-center text-white">
        <Reveal>
          <div className="max-w-3xl mx-auto px-4">
            <h2 className="text-4xl md:text-5xl font-black mb-6">Set Sail with Freel.</h2>
            <p className="text-lg text-teal-100 mb-10 max-w-2xl mx-auto">
              Get direct rates from top shipping lines without the endless email threads. Digital ocean freight made simple.
            </p>
            <Link to="/contact" className="inline-flex items-center gap-2 px-10 py-5 bg-white text-brand-navy font-bold text-lg rounded-full hover:shadow-[0_10px_30px_rgba(0,0,0,0.2)] transition-all hover:-translate-y-1">
              Request Ocean Quote <span>🌊</span>
            </Link>
          </div>
        </Reveal>
      </section>

    </div>
  );
}

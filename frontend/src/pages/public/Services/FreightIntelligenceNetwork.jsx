import React, { useEffect, useRef, useState } from 'react';
import './FreightIntelligenceNetwork.css';

export default function FreightIntelligenceNetwork() {
  const containerRef = useRef(null);
  const [isActive, setIsActive] = useState(false);

  useEffect(() => {
    const observer = new IntersectionObserver(
      ([entry]) => {
        if (entry.isIntersecting) {
          setIsActive(true);
        }
      },
      { threshold: 0.3 }
    );
    if (containerRef.current) observer.observe(containerRef.current);
    return () => observer.disconnect();
  }, []);

  const cities = [
    { name: "Delhi", x: 350, y: 200 },
    { name: "Ahmedabad", x: 200, y: 400 },
    { name: "Mumbai", x: 220, y: 550 },
    { name: "Pune", x: 260, y: 580 },
    { name: "Bangalore", x: 350, y: 700 },
    { name: "Chennai", x: 450, y: 720 },
    { name: "Hyderabad", x: 400, y: 580 },
    { name: "Kolkata", x: 650, y: 450 },
  ];

  const routes = [
    { id: "route-1", path: "M 350 200 Q 200 300 200 400", delay: "0s", dur: "3s" },
    { id: "route-2", path: "M 350 200 Q 550 300 650 450", delay: "0.2s", dur: "4s" },
    { id: "route-3", path: "M 200 400 Q 180 480 220 550", delay: "0.4s", dur: "2.5s" },
    { id: "route-4", path: "M 220 550 Q 240 565 260 580", delay: "0.6s", dur: "1.5s" },
    { id: "route-5", path: "M 260 580 Q 280 650 350 700", delay: "0.8s", dur: "3s" },
    { id: "route-6", path: "M 220 550 Q 300 520 400 580", delay: "0.3s", dur: "3.5s" },
    { id: "route-7", path: "M 400 580 Q 420 650 450 720", delay: "0.5s", dur: "2.5s" },
    { id: "route-8", path: "M 350 700 Q 400 750 450 720", delay: "0.7s", dur: "2s" },
    { id: "route-9", path: "M 400 580 Q 550 500 650 450", delay: "0.9s", dur: "4s" }
  ];

  return (
    <section ref={containerRef} className={`fin-container ${isActive ? 'is-active' : ''}`}>
      <div className="fin-bg-grid" />
      
      <div className="max-w-7xl mx-auto px-6 mb-16 relative z-10 text-center">
        <h2 className="text-5xl md:text-6xl font-black tracking-tighter mb-6 text-slate-900">
          India Doesn't Move On Roads Alone.<br/>
          <span className="text-amber-500">It Moves On Visibility.</span>
        </h2>
        <p className="text-xl md:text-2xl text-slate-500 font-medium max-w-3xl mx-auto">
          Every shipment is tracked, monitored, and managed through a connected logistics network.
        </p>
      </div>

      <div className="fin-map-wrapper relative">
        
        {/* Abstract SVG Map */}
        <svg className="w-full h-full drop-shadow-sm" viewBox="0 0 1000 900" preserveAspectRatio="xMidYMid meet">
          <defs>
            <linearGradient id="line-grad" x1="0%" y1="0%" x2="100%" y2="100%">
              <stop offset="0%" stopColor="#e2e8f0" />
              <stop offset="100%" stopColor="#cbd5e1" />
            </linearGradient>
          </defs>
          
          {/* Base Routes */}
          {routes.map(r => (
            <path key={r.id} id={r.id} d={r.path} className="fin-route-line" style={{ transitionDelay: r.delay }} />
          ))}

          {/* Glowing Freight Particles */}
          {routes.map(r => (
            <circle key={`particle-${r.id}`} r="4" className="fin-freight-particle">
              <animateMotion dur={r.dur} repeatCount="indefinite" begin={r.delay}>
                <mpath href={`#${r.id}`} />
              </animateMotion>
            </circle>
          ))}

          {/* City Nodes */}
          {cities.map((c, i) => (
            <g key={c.name} style={{ transformOrigin: `${c.x}px ${c.y}px` }}>
              <circle cx={c.x} cy={c.y} r="16" className="fin-node-pulse" style={{ animationDelay: `${i * 0.2}s` }} />
              <circle cx={c.x} cy={c.y} r="6" className="fin-node" />
              <text x={c.x + 15} y={c.y + 5} fill="#475569" fontSize="14" fontWeight="600" className="font-mono tracking-wide">
                {c.name.toUpperCase()}
              </text>
            </g>
          ))}
        </svg>

        {/* Floating UI Stats - Absolute positioned on Desktop, stacked on Mobile */}
        <div className="hidden md:block">
          <div className="fin-stat-card top-[10%] left-[5%]" style={{ transitionDelay: '0.2s' }}>
            <div className="fin-stat-value text-slate-900">500+</div>
            <div className="fin-stat-label">Verified Transporters</div>
          </div>
          
          <div className="fin-stat-card top-[15%] right-[5%]" style={{ transitionDelay: '0.4s' }}>
            <div className="fin-stat-value text-slate-900">24/7</div>
            <div className="fin-stat-label">Control Tower Monitoring</div>
          </div>
          
          <div className="fin-stat-card bottom-[20%] left-[8%]" style={{ transitionDelay: '0.6s' }}>
            <div className="fin-stat-value text-slate-900">100%</div>
            <div className="fin-stat-label">Digital Compliance</div>
          </div>
          
          <div className="fin-stat-card bottom-[25%] right-[10%]" style={{ transitionDelay: '0.8s' }}>
            <div className="fin-stat-value text-amber-500">Real-Time</div>
            <div className="fin-stat-label">Shipment Visibility</div>
          </div>
        </div>

        {/* Mobile Stats Stack */}
        <div className="md:hidden px-6 mt-8 flex flex-col gap-4 relative z-20">
          <div className="fin-stat-card !relative !translate-y-0 !opacity-100 shadow-sm border border-slate-100">
            <div className="fin-stat-value text-slate-900 text-3xl">500+</div>
            <div className="fin-stat-label">Verified Transporters</div>
          </div>
          <div className="fin-stat-card !relative !translate-y-0 !opacity-100 shadow-sm border border-slate-100">
            <div className="fin-stat-value text-slate-900 text-3xl">24/7</div>
            <div className="fin-stat-label">Control Tower Monitoring</div>
          </div>
          <div className="fin-stat-card !relative !translate-y-0 !opacity-100 shadow-sm border border-slate-100">
            <div className="fin-stat-value text-slate-900 text-3xl">100%</div>
            <div className="fin-stat-label">Digital Compliance</div>
          </div>
          <div className="fin-stat-card !relative !translate-y-0 !opacity-100 shadow-sm border border-slate-100">
            <div className="fin-stat-value text-amber-500 text-3xl">Real-Time</div>
            <div className="fin-stat-label">Shipment Visibility</div>
          </div>
        </div>

      </div>
    </section>
  );
}

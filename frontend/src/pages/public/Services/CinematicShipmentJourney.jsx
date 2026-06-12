import React, { useEffect, useRef, useState } from 'react';
import './CinematicShipmentJourney.css';

function useChapterInView(threshold = 0.4) {
  const ref = useRef(null);
  const [isActive, setIsActive] = useState(false);

  useEffect(() => {
    const el = ref.current;
    if (!el) return;
    const observer = new IntersectionObserver(([entry]) => {
      setIsActive(entry.isIntersecting);
    }, { threshold });
    
    observer.observe(el);
    return () => observer.disconnect();
  }, [threshold]);

  return [ref, isActive];
}

function AnimatedCounter({ active, target, duration = 2000, suffix = "" }) {
  const [count, setCount] = useState(0);

  useEffect(() => {
    if (active) {
      let start = 0;
      const increment = target / (duration / 16);
      const timer = setInterval(() => {
        start += increment;
        if (start >= target) {
          setCount(target);
          clearInterval(timer);
        } else {
          setCount(Math.floor(start));
        }
      }, 16);
      return () => clearInterval(timer);
    } else {
      setCount(0);
    }
  }, [active, target, duration]);

  return <>{count.toLocaleString()}{suffix}</>;
}

export default function CinematicShipmentJourney() {
  const [ref1, act1] = useChapterInView();
  const [ref2, act2] = useChapterInView();
  const [ref3, act3] = useChapterInView();
  const [ref4, act4] = useChapterInView();
  const [ref5, act5] = useChapterInView();
  const [ref6, act6] = useChapterInView();

  return (
    <div className="w-full bg-black text-white">
      
      {/* CHAPTER 1: MANUFACTURED */}
      <section ref={ref1} className={`csj-chapter relative h-screen overflow-hidden flex items-center justify-center ${act1 ? 'is-active' : ''}`}>
        <div className="csj-video-wrapper">
          <video className="w-full h-full object-cover" autoPlay muted loop playsInline src="/videos/Fulfillment_warehouse_conveyor_b…_202606052252.mp4" />
          <div className="absolute inset-0 bg-black/20" />
        </div>
        
        <div className="relative z-10 flex flex-col items-center justify-center w-full px-6">
          <div className="csj-text-reveal text-center mb-16">
            <h2 className="text-5xl md:text-7xl font-black tracking-tighter mb-4 csj-text-glow drop-shadow-2xl">
              A Product Is Manufactured.
            </h2>
            <p className="text-xl md:text-2xl text-white/90 font-semibold drop-shadow-xl">
              Every shipment starts with production.
            </p>
          </div>
          
          <div className="csj-text-reveal csj-delay-300 csj-glass-card inline-flex flex-col items-center px-10 py-6">
            <span className="text-amber-400 font-mono text-sm tracking-widest uppercase mb-1">PUNE • 06:15 AM</span>
            <span className="text-white font-bold text-xl md:text-2xl">Manufacturing Complete</span>
          </div>
        </div>
      </section>

      {/* CHAPTER 2: LOADED */}
      <section ref={ref2} className={`csj-chapter relative h-screen overflow-hidden flex items-center justify-center ${act2 ? 'is-active' : ''}`}>
        <div className="csj-video-wrapper">
          <video className="w-full h-full object-cover" autoPlay muted loop playsInline src="/videos/road-freight/Container_cranes_loading_cargo_t…_202606071648.mp4" />
          <div className="absolute inset-0 bg-black/20" />
        </div>
        
        <div className="csj-progress-badge csj-text-reveal drop-shadow-2xl"><span>01</span> / 05</div>
        
        <div className="relative z-10 text-center px-6 mt-32">
          <h2 className="csj-text-reveal text-5xl md:text-7xl font-black tracking-tighter mb-8 csj-text-glow drop-shadow-2xl">
            Loaded For The Journey.
          </h2>
          <div className="csj-text-reveal csj-delay-200 inline-flex items-center gap-4 bg-black/40 backdrop-blur-md px-6 py-3 rounded-full border border-white/20 shadow-2xl">
            <span className="text-2xl">🚛</span>
            <span className="font-mono text-sm uppercase tracking-widest text-amber-400 font-bold">Route Initiated</span>
          </div>
        </div>
      </section>

      {/* CHAPTER 3: ON THE ROAD */}
      <section ref={ref3} className={`csj-chapter relative h-screen overflow-hidden flex items-center justify-center ${act3 ? 'is-active' : ''}`}>
        <div className="csj-video-wrapper">
          <video className="w-full h-full object-cover" autoPlay muted loop playsInline src="/videos/road-freight/Convoy_of_trucks_moving_highway_202606071649.mp4" />
          <div className="absolute inset-0 bg-black/20" />
          <div className="csj-highway-lines" />
        </div>
        
        <div className="csj-progress-badge csj-text-reveal drop-shadow-2xl"><span>02</span> / 05</div>
        
        <div className="relative z-10 text-center px-6 mt-32">
          <h2 className="csj-text-reveal text-5xl md:text-7xl font-black tracking-tighter mb-4 csj-text-glow drop-shadow-2xl">
            <AnimatedCounter active={act3} target={842} /> KM Journey Begins.
          </h2>
          <div className="csj-text-reveal csj-delay-300 w-full max-w-md mx-auto h-2 bg-black/40 backdrop-blur-md rounded-full overflow-hidden mt-8 shadow-2xl">
            <div className={`h-full bg-amber-500 transition-all duration-[2000ms] ease-out ${act3 ? 'w-full' : 'w-0'}`} />
          </div>
        </div>
      </section>

      {/* CHAPTER 4: CONNECTED NETWORK */}
      <section ref={ref4} className={`csj-chapter relative h-screen overflow-hidden flex items-center justify-center ${act4 ? 'is-active' : ''}`}>
        <div className="csj-video-wrapper">
          <video className="w-full h-full object-cover" autoPlay muted loop playsInline src="/videos/Earth_at_night_global_network_202606052255.mp4" />
          <div className="absolute inset-0 bg-brand-navy/30 mix-blend-multiply" />
        </div>
        
        <div className="csj-progress-badge csj-text-reveal drop-shadow-2xl"><span>03</span> / 05</div>
        
        {/* Animated Map Overlays */}
        <svg className="absolute inset-0 w-full h-full z-10 pointer-events-none drop-shadow-2xl" viewBox="0 0 1000 800" preserveAspectRatio="xMidYMid slice">
          <path d="M 300 600 Q 400 400 600 300" stroke="rgba(245, 158, 11, 0.9)" strokeWidth="4" fill="none" className="csj-line-pulse drop-shadow-[0_0_8px_rgba(245,158,11,0.8)]" />
          <path d="M 300 600 Q 500 700 700 550" stroke="rgba(245, 158, 11, 0.9)" strokeWidth="3" fill="none" className="csj-line-pulse drop-shadow-[0_0_8px_rgba(245,158,11,0.8)]" style={{ transitionDelay: '0.3s' }} />
          <path d="M 600 300 Q 800 400 700 550" stroke="rgba(245, 158, 11, 0.9)" strokeWidth="3" fill="none" className="csj-line-pulse drop-shadow-[0_0_8px_rgba(245,158,11,0.8)]" style={{ transitionDelay: '0.6s' }} />
          
          <g className="csj-delay-200">
            <circle cx="300" cy="600" r="10" fill="#f59e0b" className="csj-node-pulse" />
            <text x="320" y="610" fill="white" fontSize="18" fontWeight="bold" className="font-mono tracking-widest csj-text-reveal drop-shadow-[0_0_10px_rgba(0,0,0,1)]">MUMBAI</text>
          </g>
          <g className="csj-delay-400">
            <circle cx="600" cy="300" r="12" fill="#f59e0b" className="csj-node-pulse" />
            <text x="620" y="310" fill="white" fontSize="18" fontWeight="bold" className="font-mono tracking-widest csj-text-reveal drop-shadow-[0_0_10px_rgba(0,0,0,1)]">DELHI</text>
          </g>
          <g className="csj-delay-500">
            <circle cx="700" cy="550" r="8" fill="#f59e0b" className="csj-node-pulse" />
            <text x="720" y="560" fill="white" fontSize="18" fontWeight="bold" className="font-mono tracking-widest csj-text-reveal drop-shadow-[0_0_10px_rgba(0,0,0,1)]">CHENNAI</text>
          </g>
        </svg>

        <div className="relative z-20 text-center px-6 mt-64">
          <h2 className="csj-text-reveal text-5xl md:text-7xl font-black tracking-tighter csj-text-glow drop-shadow-[0_0_20px_rgba(0,0,0,1)]">
            Connected Across India.
          </h2>
        </div>
      </section>

      {/* CHAPTER 5: LIVE VISIBILITY */}
      <section ref={ref5} className={`csj-chapter relative h-screen overflow-hidden flex items-center justify-center ${act5 ? 'is-active' : ''}`}>
        <div className="csj-video-wrapper">
          <video className="w-full h-full object-cover" autoPlay muted loop playsInline src="/videos/Packages_moving_through_logistic…_202606052239.mp4" />
          <div className="absolute inset-0 bg-black/20" />
        </div>
        
        <div className="csj-progress-badge csj-text-reveal drop-shadow-2xl"><span>04</span> / 05</div>
        
        <div className="absolute inset-0 z-10 overflow-hidden pointer-events-none">
           <div className="absolute top-[20%] left-[10%] md:left-[25%] csj-event-card px-8 py-4 rounded-2xl font-bold font-mono tracking-widest text-lg shadow-[0_0_30px_rgba(245,158,11,0.3)]">
             ✓ GPS VERIFIED
           </div>
           <div className="absolute top-[40%] right-[10%] md:right-[20%] csj-event-card px-8 py-4 rounded-2xl font-bold font-mono tracking-widest text-lg shadow-[0_0_30px_rgba(245,158,11,0.3)]" style={{ animationDelay: '0.4s' }}>
             ✓ DRIVER TRACKED
           </div>
           <div className="absolute bottom-[30%] left-[15%] md:left-[30%] csj-event-card px-8 py-4 rounded-2xl font-bold font-mono tracking-widest text-lg shadow-[0_0_30px_rgba(245,158,11,0.3)]" style={{ animationDelay: '0.8s' }}>
             ✓ E-WAY GENERATED
           </div>
           <div className="absolute bottom-[20%] right-[15%] md:right-[30%] csj-event-card px-8 py-4 rounded-2xl font-bold font-mono tracking-widest text-lg shadow-[0_0_30px_rgba(245,158,11,0.3)]" style={{ animationDelay: '1.2s' }}>
             ✓ CONTROL ROOM SYNCED
           </div>
        </div>

        <div className="relative z-20 text-center px-6">
          <h2 className="csj-text-reveal text-5xl md:text-7xl font-black tracking-tighter csj-text-glow drop-shadow-[0_0_20px_rgba(0,0,0,1)]">
            Every Mile Is Visible.
          </h2>
        </div>
      </section>

      {/* CHAPTER 6: DELIVERED */}
      <section ref={ref6} className={`csj-chapter relative h-screen overflow-hidden flex items-center justify-center ${act6 ? 'is-active' : ''}`}>
        <div className="csj-video-wrapper">
          <video className="w-full h-full object-cover" autoPlay muted loop playsInline src="/videos/road-freight/Freight_traffic_flowing_across_city_202606071652.mp4" />
          <div className="absolute inset-0 bg-black/20" />
        </div>
        
        <div className="csj-progress-badge csj-text-reveal drop-shadow-2xl"><span>05</span> / 05</div>
        
        <div className="relative z-10 text-center px-6">
          <h2 className="csj-text-reveal text-6xl md:text-8xl font-black tracking-tighter mb-8 text-amber-400 csj-text-glow drop-shadow-2xl">
            Delivered.
          </h2>
          <div className="csj-text-reveal csj-delay-200 csj-glass-card inline-flex flex-col items-center px-12 py-6 border-amber-500/30">
            <span className="text-white font-bold text-xl md:text-3xl mb-2">Mumbai Distribution Center</span>
            <span className="text-white/70 font-mono text-sm tracking-widest">09:42 PM</span>
          </div>
        </div>
      </section>

    </div>
  );
}

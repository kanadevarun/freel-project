import { useState, useEffect, useRef, useCallback } from 'react';

/* ═══ Initial Page Loader ═══ */
export function PageLoader() {
  const [loading, setLoading] = useState(true);
  const [fadeOut, setFadeOut] = useState(false);

  useEffect(() => {
    // Only show on initial mount (hard refresh)
    const timer = setTimeout(() => {
      setFadeOut(true);
      setTimeout(() => setLoading(false), 600); // match fade out transition
    }, 1200);
    return () => clearTimeout(timer);
  }, []);

  if (!loading) return null;

  return (
    <div className={`page-loader ${fadeOut ? 'fade-out' : ''}`}>
      <div className="loader-content">
        <h1 className="loader-logo">Freel</h1>
        <div className="loader-bar-bg">
          <div className="loader-bar-fill" />
        </div>
      </div>
    </div>
  );
}

/* ═══ Scroll Reveal ═══ */
export function useReveal() {
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
export function Reveal({ children, className = '', delay = '' }) {
  const ref = useReveal();
  return <div ref={ref} className={`reveal ${delay} ${className}`}>{children}</div>;
}

/* ═══ Scroll Progress Bar ═══ */
export function ScrollProgress() {
  const [width, setWidth] = useState(0);
  useEffect(() => {
    const onScroll = () => {
      const s = window.scrollY;
      const h = document.documentElement.scrollHeight - window.innerHeight;
      setWidth(h > 0 ? (s / h) * 100 : 0);
    };
    window.addEventListener('scroll', onScroll, { passive: true });
    return () => window.removeEventListener('scroll', onScroll);
  }, []);
  return <div className="scroll-progress" style={{ width: `${width}%` }} />;
}

/* ═══ Logistics World Map Canvas ═══ */
export function LogisticsCanvas() {
  const canvasRef = useRef(null);
  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;
    const ctx = canvas.getContext('2d');
    let animId;
    let t = 0;

    function resize() {
      canvas.width = canvas.offsetWidth * (window.devicePixelRatio || 1);
      canvas.height = canvas.offsetHeight * (window.devicePixelRatio || 1);
      ctx.scale(window.devicePixelRatio || 1, window.devicePixelRatio || 1);
    }

    // Port nodes (as % of canvas so they scale)
    const PORTS = [
      { name: 'Mumbai',    rx: 0.61, ry: 0.55, color: '#00BFA5' },
      { name: 'Shanghai',  rx: 0.79, ry: 0.40, color: '#5A4FCF' },
      { name: 'Dubai',     rx: 0.57, ry: 0.48, color: '#00BFA5' },
      { name: 'Singapore', rx: 0.77, ry: 0.60, color: '#3B82F6' },
      { name: 'London',    rx: 0.44, ry: 0.28, color: '#5A4FCF' },
      { name: 'New York',  rx: 0.20, ry: 0.34, color: '#F59E0B' },
      { name: 'Tokyo',     rx: 0.84, ry: 0.38, color: '#3B82F6' },
      { name: 'Delhi',     rx: 0.62, ry: 0.43, color: '#00BFA5' },
      { name: 'Lagos',     rx: 0.45, ry: 0.58, color: '#F59E0B' },
      { name: 'Sydney',    rx: 0.85, ry: 0.74, color: '#5A4FCF' },
    ];

    // Routes (index pairs into PORTS, + vehicle type)
    const ROUTES = [
      { a: 0, b: 2, type: 'ship',  color: '#00BFA5' },  // Mumbai → Dubai
      { a: 2, b: 4, type: 'ship',  color: '#5A4FCF' },  // Dubai → London
      { a: 0, b: 1, type: 'ship',  color: '#3B82F6' },  // Mumbai → Shanghai
      { a: 1, b: 6, type: 'ship',  color: '#3B82F6' },  // Shanghai → Tokyo
      { a: 4, b: 5, type: 'plane', color: '#F59E0B' },  // London → New York
      { a: 0, b: 3, type: 'plane', color: '#00BFA5' },  // Mumbai → Singapore
      { a: 5, b: 0, type: 'plane', color: '#F59E0B' },  // New York → Mumbai
      { a: 3, b: 9, type: 'ship',  color: '#5A4FCF' },  // Singapore → Sydney
      { a: 7, b: 2, type: 'truck', color: '#00BFA5' },  // Delhi → Dubai
      { a: 2, b: 8, type: 'ship',  color: '#F59E0B' },  // Dubai → Lagos
      { a: 1, b: 3, type: 'ship',  color: '#3B82F6' },  // Shanghai → Singapore
      { a: 8, b: 4, type: 'plane', color: '#F59E0B' },  // Lagos → London
    ];

    // Each vehicle is an animated traveler on a route
    const vehicles = ROUTES.map((r, i) => ({
      ...r,
      progress: (i / ROUTES.length),  // stagger start positions
      speed: 0.0008 + Math.random() * 0.0006,
    }));

    // Cubic bezier arc midpoint for curved routes
    function arcMid(ax, ay, bx, by, type) {
      const mx = (ax + bx) / 2;
      const my = (ay + by) / 2;
      const dx = bx - ax, dy = by - ay;
      const len = Math.sqrt(dx * dx + dy * dy);
      const curve = type === 'plane' ? -0.45 : type === 'ship' ? -0.3 : -0.2;
      return { cx: mx - dy * curve, cy: my + dx * curve, len };
    }

    // Point on quadratic bezier at t
    function bezierPt(ax, ay, cx, cy, bx, by, t) {
      const u = 1 - t;
      return {
        x: u * u * ax + 2 * u * t * cx + t * t * bx,
        y: u * u * ay + 2 * u * t * cy + t * t * by,
      };
    }

    // Draw dashed arc route
    function drawRoute(ax, ay, cx, cy, bx, by, color, type) {
      ctx.beginPath();
      ctx.moveTo(ax, ay);
      ctx.quadraticCurveTo(cx, cy, bx, by);
      ctx.strokeStyle = color.replace(')', ', 0.18)').replace('rgb', 'rgba').replace('#', 'rgba(');
      // Parse hex to rgba
      const r = parseInt(color.slice(1, 3), 16);
      const g = parseInt(color.slice(3, 5), 16);
      const b = parseInt(color.slice(5, 7), 16);
      ctx.strokeStyle = `rgba(${r},${g},${b},0.18)`;
      ctx.lineWidth = type === 'ship' ? 1.2 : type === 'plane' ? 1.5 : 1;
      ctx.setLineDash(type === 'truck' ? [4, 6] : type === 'plane' ? [6, 8] : [8, 10]);
      ctx.stroke();
      ctx.setLineDash([]);
    }

    // Draw glowing port node
    function drawNode(px, py, color, pulse) {
      const r = parseInt(color.slice(1, 3), 16);
      const g = parseInt(color.slice(3, 5), 16);
      const b = parseInt(color.slice(5, 7), 16);

      // Outer pulse ring
      const pulseR = 10 + Math.sin(t * 2 + pulse) * 4;
      const grad = ctx.createRadialGradient(px, py, 0, px, py, pulseR + 6);
      grad.addColorStop(0, `rgba(${r},${g},${b},0.3)`);
      grad.addColorStop(1, `rgba(${r},${g},${b},0)`);
      ctx.beginPath();
      ctx.arc(px, py, pulseR + 6, 0, Math.PI * 2);
      ctx.fillStyle = grad;
      ctx.fill();

      // Core dot
      ctx.beginPath();
      ctx.arc(px, py, 4, 0, Math.PI * 2);
      ctx.fillStyle = color;
      ctx.shadowColor = color;
      ctx.shadowBlur = 10;
      ctx.fill();
      ctx.shadowBlur = 0;

      // Inner bright center
      ctx.beginPath();
      ctx.arc(px, py, 2, 0, Math.PI * 2);
      ctx.fillStyle = '#ffffff';
      ctx.fill();
    }

    // Draw vehicle emoji on route
    function drawVehicle(px, py, dx, dy, type, color) {
      const r = parseInt(color.slice(1, 3), 16);
      const g = parseInt(color.slice(3, 5), 16);
      const b = parseInt(color.slice(5, 7), 16);

      // Glow trail
      ctx.beginPath();
      ctx.arc(px, py, 8, 0, Math.PI * 2);
      const grd = ctx.createRadialGradient(px, py, 0, px, py, 8);
      grd.addColorStop(0, `rgba(${r},${g},${b},0.5)`);
      grd.addColorStop(1, `rgba(${r},${g},${b},0)`);
      ctx.fillStyle = grd;
      ctx.fill();

      // Vehicle icon
      ctx.save();
      ctx.translate(px, py);
      const angle = Math.atan2(dy, dx);
      if (type === 'plane') ctx.rotate(angle);
      ctx.font = type === 'truck' ? '13px serif' : '12px serif';
      ctx.textAlign = 'center';
      ctx.textBaseline = 'middle';
      ctx.shadowColor = color;
      ctx.shadowBlur = 8;
      const icon = type === 'ship' ? '🚢' : type === 'plane' ? '✈️' : '🚛';
      ctx.fillText(icon, 0, 0);
      ctx.shadowBlur = 0;
      ctx.restore();
    }

    function draw() {
      const W = canvas.offsetWidth;
      const H = canvas.offsetHeight;
      ctx.clearRect(0, 0, W, H);
      t += 0.016;

      // Compute port pixel positions
      const ports = PORTS.map(p => ({
        ...p,
        px: p.rx * W,
        py: p.ry * H,
      }));

      // Draw routes
      ROUTES.forEach((route) => {
        const A = ports[route.a], B = ports[route.b];
        const { cx, cy } = arcMid(A.px, A.py, B.px, B.py, route.type);
        drawRoute(A.px, A.py, cx, cy, B.px, B.py, route.color, route.type);
      });

      // Draw vehicles
      vehicles.forEach((v) => {
        v.progress = (v.progress + v.speed) % 1;
        const A = ports[v.a], B = ports[v.b];
        const { cx, cy } = arcMid(A.px, A.py, B.px, B.py, v.type);
        const pos = bezierPt(A.px, A.py, cx, cy, B.px, B.py, v.progress);
        const posNext = bezierPt(A.px, A.py, cx, cy, B.px, B.py, Math.min(v.progress + 0.01, 1));
        drawVehicle(pos.x, pos.y, posNext.x - pos.x, posNext.y - pos.y, v.type, v.color);
      });

      // Draw port nodes on top
      ports.forEach((p, i) => drawNode(p.px, p.py, p.color, i * 1.2));

      // Subtle grid overlay
      ctx.strokeStyle = 'rgba(0,191,165,0.03)';
      ctx.lineWidth = 0.5;
      const gridSize = 60;
      for (let x = 0; x < W; x += gridSize) {
        ctx.beginPath(); ctx.moveTo(x, 0); ctx.lineTo(x, H); ctx.stroke();
      }
      for (let y = 0; y < H; y += gridSize) {
        ctx.beginPath(); ctx.moveTo(0, y); ctx.lineTo(W, y); ctx.stroke();
      }

      animId = requestAnimationFrame(draw);
    }

    resize();
    draw();
    const onResize = () => { resize(); };
    window.addEventListener('resize', onResize);
    return () => { cancelAnimationFrame(animId); window.removeEventListener('resize', onResize); };
  }, []);
  return <canvas ref={canvasRef} className="particle-canvas" />;
}


/* ═══ Character Stagger Title ═══ */
export function StaggerTitle({ line1, line2 }) {
  const chars1 = line1.split('');
  const baseDelay = 0.3;
  return (
    <h1 className="hero-title">
      {chars1.map((c, i) => c === ' '
        ? <span key={`a${i}`} className="stagger-space" />
        : <span key={`a${i}`} className="stagger-char" style={{ animationDelay: `${baseDelay + i * 0.025}s` }}>{c}</span>
      )}
      <br />
      <span style={{ display: 'inline-block', opacity: 0, animation: `fadeInUp 0.8s ease ${baseDelay + chars1.length * 0.025 + 0.2}s forwards` }}>
        <span className="text-gradient-animated">{line2}</span>
      </span>
    </h1>
  );
}

/* ═══ 3D Tilt Card ═══ */
export function TiltCard({ children, className = '' }) {
  const cardRef = useRef(null);
  const spotRef = useRef(null);
  const handleMove = useCallback((e) => {
    const el = cardRef.current;
    if (!el) return;
    const rect = el.getBoundingClientRect();
    const x = e.clientX - rect.left, y = e.clientY - rect.top;
    const cx = rect.width / 2, cy = rect.height / 2;
    const rotateX = ((y - cy) / cy) * -6;
    const rotateY = ((x - cx) / cx) * 6;
    el.querySelector('.tilt-card-inner').style.transform =
      `perspective(1200px) rotateX(${rotateX}deg) rotateY(${rotateY}deg)`;
    if (spotRef.current) {
      spotRef.current.style.left = `${x}px`;
      spotRef.current.style.top = `${y}px`;
    }
  }, []);
  const handleLeave = useCallback(() => {
    const el = cardRef.current;
    if (el) el.querySelector('.tilt-card-inner').style.transform = 'perspective(1200px) rotateX(0) rotateY(0)';
  }, []);
  return (
    <div ref={cardRef} className={`tilt-card ${className}`} onMouseMove={handleMove} onMouseLeave={handleLeave}>
      <div className="tilt-card-inner">
        {children}
        <div ref={spotRef} className="card-spotlight" />
      </div>
    </div>
  );
}

/* ═══ Sticky Stacking Cards (Awwwards Style) ═══ */
export function StickyStackCards({ cards }) {
  const containerRef = useRef(null);

  useEffect(() => {
    const handleScroll = () => {
      if (!containerRef.current) return;
      const cardEls = containerRef.current.querySelectorAll('.stack-card');
      
      cardEls.forEach((card, index) => {
        const rect = card.getBoundingClientRect();
        const cardTop = rect.top;
        const windowH = window.innerHeight;
        
        // When card hits the sticky top (approx 15vh)
        const stickyTop = windowH * 0.15 + (index * 20);
        
        if (cardTop <= stickyTop) {
          // It's sticking! Calculate how far the next card is to animate this one backwards.
          const nextCard = cardEls[index + 1];
          let progress = 0;
          if (nextCard) {
            const nextRect = nextCard.getBoundingClientRect();
            // Progress goes from 0 to 1 as the next card comes up from the bottom to cover it
            progress = Math.max(0, Math.min(1, (windowH - nextRect.top) / (windowH - stickyTop)));
          }
          
          const scale = 1 - (progress * 0.06); // shrink down to 0.94
          const blur = progress * 5; // blur up to 5px
          const brightness = 1 - (progress * 0.5); // darken
          
          card.style.transform = `scale(${scale})`;
          card.style.filter = `blur(${blur}px) brightness(${brightness})`;
        } else {
          // Reset
          card.style.transform = `scale(1)`;
          card.style.filter = `blur(0px) brightness(1)`;
        }
      });
    };

    window.addEventListener('scroll', handleScroll, { passive: true });
    handleScroll(); // Initial check
    return () => window.removeEventListener('scroll', handleScroll);
  }, []);

  return (
    <div ref={containerRef} className="stack-cards-container">
      {cards.map((c, i) => (
        <div key={i} className="stack-card" style={{ top: `calc(15vh + ${i * 20}px)` }}>
          <div className="stack-card-inner">
            <div className="stack-card-bg" style={{ backgroundImage: `url(${c.src})` }} />
            <div className="stack-card-overlay">
              <h3 className="stack-card-title">{c.title}</h3>
              <p className="stack-card-desc">{c.desc}</p>
            </div>
          </div>
        </div>
      ))}
    </div>
  );
}

/* ═══ Sticky Showcase (Scroll-linked Image Reveal) ═══ */
export function StickyShowcase({ features }) {
  const [activeIndex, setActiveIndex] = useState(0);
  const textRefs = useRef([]);

  useEffect(() => {
    const observer = new IntersectionObserver(
      (entries) => {
        entries.forEach(entry => {
          if (entry.isIntersecting) {
            setActiveIndex(Number(entry.target.dataset.index));
          }
        });
      },
      { rootMargin: '-50% 0px -50% 0px' }
    );

    textRefs.current.forEach(ref => {
      if (ref) observer.observe(ref);
    });

    return () => observer.disconnect();
  }, []);

  return (
    <div className="sticky-showcase-container">
      {/* Left side: Sticky Images */}
      <div className="sticky-visual-wrapper">
        <div className="sticky-visual-inner">
          {features.map((f, i) => (
            <div key={i} className={`showcase-visual ${i === activeIndex ? 'active' : ''}`}>
              <img src={f.image} alt={f.title} className="showcase-img" />
            </div>
          ))}
        </div>
      </div>

      {/* Right side: Scrolling Text */}
      <div className="sticky-text-wrapper">
        {features.map((f, i) => (
          <div 
            key={i} 
            ref={el => textRefs.current[i] = el}
            data-index={i}
            className={`showcase-text-block ${i === activeIndex ? 'active' : ''}`}
          >
            <h3 className="showcase-title">{f.title}</h3>
            <p className="showcase-desc">{f.desc}</p>
            <div className="showcase-stats">
              {f.stats.map((s, j) => (
                <div key={j} className="showcase-stat">
                  <div className="s-val">{s.v}</div>
                  <div className="s-label">{s.l}</div>
                </div>
              ))}
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

/* ═══ Magnetic Button ═══ */
export function MagneticBtn({ children, className = '' }) {
  const ref = useRef(null);
  const handleMove = useCallback((e) => {
    const el = ref.current;
    if (!el) return;
    const rect = el.getBoundingClientRect();
    const x = e.clientX - rect.left - rect.width / 2;
    const y = e.clientY - rect.top - rect.height / 2;
    el.style.transform = `translate(${x * 0.15}px, ${y * 0.15}px)`;
  }, []);
  const handleLeave = useCallback(() => {
    if (ref.current) ref.current.style.transform = 'translate(0,0)';
  }, []);
  return (
    <div ref={ref} className={`magnetic-btn ${className}`} onMouseMove={handleMove} onMouseLeave={handleLeave}>
      {children}
    </div>
  );
}

/* ═══ Counter ═══ */
export function Counter({ end, suffix = '', label }) {
  const [val, setVal] = useState(0);
  const ref = useRef(null);
  useEffect(() => {
    const el = ref.current;
    if (!el) return;
    const obs = new IntersectionObserver(([e]) => {
      if (e.isIntersecting) {
        let start = 0;
        const step = Math.ceil(end / 40);
        const timer = setInterval(() => {
          start += step;
          if (start >= end) { start = end; clearInterval(timer); }
          setVal(start);
        }, 30);
        obs.unobserve(el);
      }
    }, { threshold: 0.3 });
    obs.observe(el);
    return () => obs.disconnect();
  }, [end]);
  return (
    <div ref={ref}>
      <div className="counter-num">{val.toLocaleString()}{suffix}</div>
      <div className="counter-label">{label}</div>
    </div>
  );
}

/* ═══ Timeline Section ═══ */
const timelineSteps = [
  { n: '1', title: 'Sign Up Free', desc: 'Create your account with GST verification in under 2 minutes.', icon: '🚀', tag: 'Quick Start' },
  { n: '2', title: 'Request Quotes', desc: 'Send RFQs to multiple vendors or compare instant rates across 500+ carriers.', icon: '📋', tag: 'Multi-Vendor' },
  { n: '3', title: 'Compare & Book', desc: 'Pick the best rate with margin visibility and generate branded quotes.', icon: '⚖️', tag: 'Smart Select' },
  { n: '4', title: 'Ship & Track', desc: 'Monitor shipments live. GPS for trucks, AIS for vessels, FlightAware for air.', icon: '📍', tag: 'Real-Time' },
];

export function Timeline() {
  return (
    <div className="timeline-grid">
      {timelineSteps.map((s, i) => (
        <Reveal key={i} delay={`reveal-delay-${i + 1}`}>
          <div className="timeline-step">
            <div className={`timeline-num timeline-num-${i + 1}`}>{s.n}</div>
            <div className="timeline-icon">{s.icon}</div>
            <h3 className="timeline-title">{s.title}</h3>
            <p className="timeline-desc">{s.desc}</p>
            <span className="timeline-visual-tag">{s.tag}</span>
          </div>
        </Reveal>
      ))}
    </div>
  );
}

/* ═══ Section Header ═══ */
export function SectionHeader({ label, labelColor = 'teal', title, subtitle }) {
  return (
    <div className="text-center mb-14">
      <span className={`section-label section-label-${labelColor}`}>{label}</span>
      <h2 className="text-3xl md:text-[2rem] font-bold text-brand-navy mb-4">{title}</h2>
      {subtitle && <p className="text-brand-gray max-w-xl mx-auto">{subtitle}</p>}
    </div>
  );
}

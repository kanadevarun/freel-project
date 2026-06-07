import React, { useRef, useEffect, useState } from 'react';
import './GlobalScale.css';

/* ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
   GLOBAL SCALE — Cinematic Earth Section
   Minimal editorial. Earth is the hero.
   ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━ */

function RouteParticles({ containerRef }) {
  const canvasRef = useRef(null);

  useEffect(() => {
    const canvas = canvasRef.current;
    const container = containerRef.current;
    if (!canvas || !container) return;

    const ctx = canvas.getContext('2d');
    let animId;
    let particles = [];

    function resize() {
      const dpr = window.devicePixelRatio || 1;
      const rect = container.getBoundingClientRect();
      canvas.width = rect.width * dpr;
      canvas.height = rect.height * dpr;
      canvas.style.width = rect.width + 'px';
      canvas.style.height = rect.height + 'px';
      ctx.setTransform(dpr, 0, 0, dpr, 0, 0);
    }

    function createParticle() {
      const W = container.offsetWidth;
      const H = container.offsetHeight;
      return {
        x: Math.random() * W,
        y: Math.random() * H,
        size: Math.random() * 1.5 + 0.3,
        speedX: (Math.random() - 0.5) * 0.2,
        speedY: (Math.random() - 0.5) * 0.12,
        opacity: Math.random() * 0.3 + 0.05,
        pulseSpeed: Math.random() * 0.012 + 0.003,
        pulsePhase: Math.random() * Math.PI * 2,
        hue: 160 + Math.random() * 40,
      };
    }

    const count = Math.min(40, Math.floor(container.offsetWidth / 40));
    for (let i = 0; i < count; i++) particles.push(createParticle());

    let time = 0;
    function draw() {
      const W = container.offsetWidth;
      const H = container.offsetHeight;
      ctx.clearRect(0, 0, W, H);
      time += 0.016;

      particles.forEach((p) => {
        p.x += p.speedX;
        p.y += p.speedY;
        if (p.x < -10) p.x = W + 10;
        if (p.x > W + 10) p.x = -10;
        if (p.y < -10) p.y = H + 10;
        if (p.y > H + 10) p.y = -10;

        const pulse = Math.sin(time * p.pulseSpeed * 60 + p.pulsePhase);
        const alpha = p.opacity * (0.6 + pulse * 0.4);
        const size = p.size * (0.8 + pulse * 0.2);

        const grad = ctx.createRadialGradient(p.x, p.y, 0, p.x, p.y, size * 5);
        grad.addColorStop(0, `hsla(${p.hue}, 70%, 65%, ${alpha * 0.35})`);
        grad.addColorStop(1, `hsla(${p.hue}, 70%, 65%, 0)`);
        ctx.fillStyle = grad;
        ctx.beginPath();
        ctx.arc(p.x, p.y, size * 5, 0, Math.PI * 2);
        ctx.fill();

        ctx.fillStyle = `hsla(${p.hue}, 70%, 75%, ${alpha})`;
        ctx.beginPath();
        ctx.arc(p.x, p.y, size, 0, Math.PI * 2);
        ctx.fill();
      });

      // Very subtle connecting lines
      for (let i = 0; i < particles.length; i++) {
        for (let j = i + 1; j < particles.length; j++) {
          const dx = particles[i].x - particles[j].x;
          const dy = particles[i].y - particles[j].y;
          const dist = Math.sqrt(dx * dx + dy * dy);
          if (dist < 160) {
            ctx.strokeStyle = `rgba(0, 191, 165, ${(1 - dist / 160) * 0.035})`;
            ctx.lineWidth = 0.4;
            ctx.beginPath();
            ctx.moveTo(particles[i].x, particles[i].y);
            ctx.lineTo(particles[j].x, particles[j].y);
            ctx.stroke();
          }
        }
      }

      animId = requestAnimationFrame(draw);
    }

    resize();
    draw();
    window.addEventListener('resize', resize);
    return () => {
      cancelAnimationFrame(animId);
      window.removeEventListener('resize', resize);
    };
  }, [containerRef]);

  return <canvas ref={canvasRef} className="gs-particles-canvas" />;
}

export default function GlobalScale() {
  const sectionRef = useRef(null);
  const imgRef = useRef(null);
  const [isVisible, setIsVisible] = useState(false);

  /* ── Scroll reveal ── */
  useEffect(() => {
    const section = sectionRef.current;
    if (!section) return;
    const observer = new IntersectionObserver(
      ([entry]) => {
        if (entry.isIntersecting) {
          setIsVisible(true);
          observer.unobserve(section);
        }
      },
      { threshold: 0.08 }
    );
    observer.observe(section);
    return () => observer.disconnect();
  }, []);

  /* ── Parallax + Zoom ── */
  useEffect(() => {
    const section = sectionRef.current;
    const img = imgRef.current;
    if (!section || !img) return;

    const handleScroll = () => {
      const rect = section.getBoundingClientRect();
      const windowH = window.innerHeight;
      const sectionH = section.offsetHeight;
      const progress = Math.max(0, Math.min(1,
        (windowH - rect.top) / (windowH + sectionH)
      ));

      const scale = 1.02 + progress * 0.12;
      const translateY = (0.5 - progress) * 50;
      const translateX = (progress - 0.5) * 6;

      img.style.transform = `scale(${scale}) translate(${translateX}px, ${translateY}px)`;
    };

    window.addEventListener('scroll', handleScroll, { passive: true });
    handleScroll();
    return () => window.removeEventListener('scroll', handleScroll);
  }, []);

  return (
    <section
      ref={sectionRef}
      className={`gs-section ${isVisible ? 'gs-revealed' : ''}`}
      id="global-scale"
    >
      {/* ━━ Transition zone from hero ━━ */}
      <div className="gs-transition-zone">
        <div className="gs-transition-glow" />
        <svg className="gs-route-lines" viewBox="0 0 1440 160" preserveAspectRatio="none">
          <path className="gs-route-line gs-rl-1" d="M0,140 Q360,30 720,80 T1440,50" />
          <path className="gs-route-line gs-rl-2" d="M0,90 Q300,140 600,60 T1440,120" />
          <path className="gs-route-line gs-rl-3" d="M0,50 Q400,120 800,30 T1440,90" />
          <circle className="gs-rn" cx="360" cy="80" r="2.5" />
          <circle className="gs-rn gs-rn-b" cx="720" cy="65" r="2" />
          <circle className="gs-rn" cx="1080" cy="90" r="2.5" />
          <circle className="gs-rn gs-rn-p" cx="540" cy="100" r="1.8" />
        </svg>
      </div>

      {/* ━━ Background layers ━━ */}
      <div className="gs-deep-bg" />
      <div className="gs-grid-pattern" />
      <RouteParticles containerRef={sectionRef} />

      {/* ━━ Earth — The Hero ━━ */}
      <div className="gs-earth-wrapper">
        <img
          ref={imgRef}
          src="/images/Earth_viewed_from_space_at_202606052226.jpeg"
          alt="Earth viewed from space — global trade network"
          className="gs-earth-img"
        />
        <div className="gs-atmo gs-atmo-core" />
        <div className="gs-atmo gs-atmo-mid" />
        <div className="gs-atmo gs-atmo-outer" />
      </div>

      {/* ━━ Overlays ━━ */}
      <div className="gs-ov-top" />
      <div className="gs-ov-bottom" />
      <div className="gs-ov-left" />
      <div className="gs-ov-vignette" />
      <div className="gs-light-sweep" />

      {/* ━━ Content — Minimal editorial ━━ */}
      <div className="gs-content-layer">
        <div className="gs-editorial">
          <span className="gs-eyebrow">
            <span className="gs-eyebrow-bar" />
            Global Trade Never Sleeps
          </span>

          <h2 className="gs-headline">
            <span className="gs-hl-line">Every Product Travels</span>
            <span className="gs-hl-line">Through A <span className="gs-hl-accent">Complex</span></span>
            <span className="gs-hl-line"><span className="gs-hl-accent">Global Network</span></span>
          </h2>

          <p className="gs-body">
            Global trade moves through a vast network of ports, airports,
            carriers and distribution hubs. Freel connects it all in one platform.
          </p>

          <div className="gs-meta-row">
            <span className="gs-meta">150+ Countries</span>
            <span className="gs-meta-dot" />
            <span className="gs-meta">10K+ Trade Lanes</span>
            <span className="gs-meta-dot" />
            <span className="gs-meta">Real-Time Visibility</span>
            <span className="gs-meta-dot" />
            <span className="gs-meta">24/7 Operations</span>
          </div>
        </div>
      </div>
    </section>
  );
}

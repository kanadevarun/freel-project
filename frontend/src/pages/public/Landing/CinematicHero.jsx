import React, { useRef, useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import gsap from 'gsap';
import { ScrollTrigger } from 'gsap/ScrollTrigger';
import { MagneticBtn } from './LandingComponents';
import './CinematicHero.css';

/* ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
   CINEMATIC HERO — Living Logistics Command Center
   ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━ */

export default function CinematicHero() {
  const heroRef = useRef(null);
  const imgRef = useRef(null);
  const overlayRef = useRef(null);
  const titleRef = useRef(null);
  const subRef = useRef(null);
  const ctaRef = useRef(null);
  const statsRef = useRef(null);
  const badgeRef = useRef(null);
  const [mousePos, setMousePos] = useState({ x: 0.5, y: 0.5 });

  /* ── Mouse parallax ── */
  const handleMouseMove = (e) => {
    const rect = heroRef.current?.getBoundingClientRect();
    if (!rect) return;
    const x = (e.clientX - rect.left) / rect.width;
    const y = (e.clientY - rect.top) / rect.height;
    setMousePos({ x, y });
  };

  /* ── GSAP entrance timeline ── */
  useEffect(() => {
    const ctx = gsap.context(() => {
      const tl = gsap.timeline({ delay: 0.3 });

      /* Image/Video zoom-in reveal */
      gsap.set(imgRef.current, { scale: 1.15, opacity: 0 });
      tl.to(imgRef.current, {
        scale: 1.05,
        opacity: 1,
        duration: 2.5,
        ease: 'power3.out',
      }, 0);

      /* Overlay fade */
      gsap.set(overlayRef.current, { opacity: 0 });
      tl.to(overlayRef.current, { opacity: 1, duration: 1.5, ease: 'power2.out' }, 0.5);

      /* Badge */
      gsap.set(badgeRef.current, { opacity: 0, y: 30 });
      tl.to(badgeRef.current, { opacity: 1, y: 0, duration: 0.8, ease: 'power3.out' }, 1.0);

      /* Title lines stagger */
      const titleLines = titleRef.current?.querySelectorAll('.cin-title-line');
      if (titleLines) {
        gsap.set(titleLines, { opacity: 0, y: 60 });
        tl.to(titleLines, {
          opacity: 1,
          y: 0,
          duration: 1.2,
          stagger: 0.25,
          ease: 'power4.out',
        }, 1.2);
      }

      /* Sub text */
      gsap.set(subRef.current, { opacity: 0, y: 40 });
      tl.to(subRef.current, { opacity: 1, y: 0, duration: 1, ease: 'power3.out' }, 1.8);

      /* CTA buttons */
      gsap.set(ctaRef.current, { opacity: 0, y: 30 });
      tl.to(ctaRef.current, { opacity: 1, y: 0, duration: 0.8, ease: 'power3.out' }, 2.2);

      /* Stats bar */
      gsap.set(statsRef.current, { opacity: 0, y: 20 });
      tl.to(statsRef.current, { opacity: 1, y: 0, duration: 0.8, ease: 'power3.out' }, 2.5);

    }, heroRef);

    /* ── Scroll Parallax ── */
    gsap.registerPlugin(ScrollTrigger);
    const scrollCtx = gsap.context(() => {
      // Video - slow move down
      gsap.to('.cin-base-video', {
        y: '20%',
        ease: 'none',
        scrollTrigger: {
          trigger: '.cin-hero',
          start: 'top top',
          end: 'bottom top',
          scrub: true,
        }
      });

      // Content Layer - faster move down (slower scroll effect)
      gsap.to('.cin-content-inner', {
        y: '40%',
        ease: 'none',
        scrollTrigger: {
          trigger: '.cin-hero',
          start: 'top top',
          end: 'bottom top',
          scrub: true,
        }
      });
    }, heroRef);

    return () => {
      ctx.revert();
      scrollCtx.revert();
    };
  }, []);

  /* ── Parallax transforms from mouse ── */
  const px = (mousePos.x - 0.5) * 2; /* -1 to 1 */
  const py = (mousePos.y - 0.5) * 2;

  const imgTransform = `translate(${px * -8}px, ${py * -6}px) scale(1.05)`;
  const glowTransform = `translate(${px * 20}px, ${py * 15}px)`;

  return (
    <section
      ref={heroRef}
      className="cin-hero"
      onMouseMove={handleMouseMove}
      id="cinematic-hero"
    >
      {/* ━━ LAYER 1: Base Video with Camera Drift ━━ */}
      <div className="cin-img-wrapper">
        <video
          ref={imgRef}
          src="/videos/Package_across_global_supply_chain_202606052243.mp4"
          autoPlay
          loop
          muted
          playsInline
          className="cin-base-video"
          style={{ transform: imgTransform }}
        />
      </div>

      {/* ━━ LAYER 2: Cinematic Dark Overlay + Vignette ━━ */}
      <div ref={overlayRef} className="cin-overlay">
        <div className="cin-vignette" />
        <div className="cin-gradient-overlay" />
      </div>

      {/* ━━ LAYER 3: Ambient Glow Effects ━━ */}
      <div className="cin-ambient-glows" style={{ transform: glowTransform }}>
        <div className="cin-ambient-glow cin-glow-blue" />
        <div className="cin-ambient-glow cin-glow-teal" />
        <div className="cin-ambient-glow cin-glow-purple" />
      </div>

      {/* ━━ LAYER 4: Hero Content ━━ */}
      <div className="cin-content-layer">
        <div className="cin-content-inner">
          {/* Badge */}
          <div ref={badgeRef} className="cin-badge">
            <span className="cin-badge-dot" />
            <span>AI-Powered Logistics OS</span>
          </div>

          {/* Title */}
          <h1 ref={titleRef} className="cin-title">
            <span className="cin-title-line">The Operating System For</span>
            <span className="cin-title-line cin-title-accent">Global Freight</span>
          </h1>

          {/* Subtitle */}
          <p ref={subRef} className="cin-subtitle">
            AI-powered logistics operating system connecting manufacturers,
            freight forwarders, and shipping lines across the globe.
          </p>

          {/* CTA */}
          <div ref={ctaRef} className="cin-cta-row">
            <MagneticBtn>
              <Link to="/contact" className="cin-btn-primary">
                <span>Book Demo</span>
                <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                  <path d="M5 12h14M12 5l7 7-7 7"/>
                </svg>
              </Link>
            </MagneticBtn>
            <MagneticBtn>
              <Link to="/platform" className="cin-btn-secondary">
                Explore Platform
              </Link>
            </MagneticBtn>
          </div>

          {/* Stats */}
          <div ref={statsRef} className="cin-stats">
            <div className="cin-stat">
              <span className="cin-stat-num">500+</span>
              <span className="cin-stat-label">Verified Carriers</span>
            </div>
            <div className="cin-stat-divider" />
            <div className="cin-stat">
              <span className="cin-stat-num">150+</span>
              <span className="cin-stat-label">Global Ports</span>
            </div>
            <div className="cin-stat-divider" />
            <div className="cin-stat">
              <span className="cin-stat-num">10K+</span>
              <span className="cin-stat-label">Trade Lanes</span>
            </div>
            <div className="cin-stat-divider" />
            <div className="cin-stat">
              <span className="cin-stat-num">98%</span>
              <span className="cin-stat-label">On-Time Rate</span>
            </div>
          </div>
        </div>
      </div>

      {/* ━━ Scroll Indicator ━━ */}
      <div className="cin-scroll-hint">
        <span className="cin-scroll-text">Scroll to explore</span>
        <div className="cin-scroll-line">
          <div className="cin-scroll-dot" />
        </div>
      </div>
    </section>
  );
}

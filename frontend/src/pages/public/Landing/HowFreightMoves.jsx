import React, { useEffect, useRef } from 'react';
import './HowFreightMoves.css';

/* ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
   HOW FREIGHT MOVES — Immersive Freight Journey Sequence
   ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━ */

export default function HowFreightMoves() {
  const sectionRef = useRef(null);

  useEffect(() => {
    const observer = new IntersectionObserver(
      (entries) => {
        entries.forEach((entry) => {
          if (entry.isIntersecting) {
            entry.target.classList.add('hfm-visible');
          }
        });
      },
      { threshold: 0.15 }
    );

    const panels = sectionRef.current.querySelectorAll('.hfm-panel');
    panels.forEach((panel) => observer.observe(panel));

    return () => observer.disconnect();
  }, []);

  return (
    <section className="hfm-section" ref={sectionRef}>
      {/* ─── Panel 1: Air Freight ─── */}
      <div className="hfm-panel">
        <div className="hfm-visual">
          <video
            autoPlay
            muted
            loop
            playsInline
            className="hfm-media"
            src="/videos/air-freight/Air_cargo_hub_loading_freighter_202606071526.mp4"
          />
          <div className="hfm-overlay" />
        </div>
        <div className="hfm-content-wrapper">
          <div className="hfm-content">
            <h2 className="hfm-panel-title">First, It Takes Flight</h2>
            <p className="hfm-panel-desc">
              Urgent shipments move across continents in hours.
            </p>
          </div>
        </div>
      </div>

      {/* ─── Panel 2: Sea Freight ─── */}
      <div className="hfm-panel">
        <div className="hfm-visual">
          <video
            autoPlay
            muted
            loop
            playsInline
            className="hfm-media"
            src="/videos/sea-freight/Steel_leviathan_cutting_through_…_202606071605.mp4"
          />
          <div className="hfm-overlay" />
        </div>
        <div className="hfm-content-wrapper">
          <div className="hfm-content">
            <h2 className="hfm-panel-title">Then It Crosses Oceans</h2>
            <p className="hfm-panel-desc">
              Container vessels move the majority of global trade.
            </p>
          </div>
        </div>
      </div>

      {/* ─── Panel 3: Road Freight ─── */}
      <div className="hfm-panel">
        <div className="hfm-visual">
          <video
            autoPlay
            muted
            loop
            playsInline
            className="hfm-media"
            src="/videos/Packages_moving_through_logistic…_202606052239.mp4"
          />
          <div className="hfm-overlay" />
        </div>
        <div className="hfm-content-wrapper">
          <div className="hfm-content">
            <h2 className="hfm-panel-title">Finally, It Reaches Its Destination</h2>
            <p className="hfm-panel-desc">
              Road networks connect ports, warehouses and customers.
            </p>
          </div>
        </div>
      </div>
    </section>
  );
}

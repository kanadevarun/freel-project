import React, { useEffect, useRef } from 'react';
import './CommandCenter.css';

/* ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
   COMMAND CENTER — Operations & Visibility Section
   ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━ */

export default function CommandCenter() {
  const sectionRef = useRef(null);

  useEffect(() => {
    const observer = new IntersectionObserver(
      (entries) => {
        entries.forEach((entry) => {
          if (entry.isIntersecting) {
            entry.target.classList.add('cc-visible');
          }
        });
      },
      { threshold: 0.2 }
    );

    if (sectionRef.current) {
      observer.observe(sectionRef.current);
    }

    return () => observer.disconnect();
  }, []);

  return (
    <section className="cc-section" ref={sectionRef}>
      {/* Subtle animated background grid */}
      <div className="cc-bg-grid"></div>

      <div className="container-lg relative z-10">
        <div className="cc-layout">
          
          {/* ─── LEFT: Editorial Content ─── */}
          <div className="cc-content">
            <div className="cc-eyebrow">
              <span className="cc-pulse-dot"></span>
              Operating Global Logistics
            </div>
            
            <h2 className="cc-title">
              One Platform.<br />
              <span className="cc-title-highlight">Complete Visibility.</span>
            </h2>
            
            <div className="cc-body">
              <p>
                Track shipments, carriers, ports and routes from a single connected platform.
              </p>
              <p>
                Freel brings every moving part of global logistics into one real-time operational view.
              </p>
              <p>
                From air cargo and ocean freight to warehousing and final-mile delivery, every shipment stays visible from origin to destination.
              </p>
            </div>
          </div>

          {/* ─── RIGHT: Immersive Visualization ─── */}
          <div className="cc-visual">
            <div className="cc-visual-container">
              
              {/* Base Map Image */}
              <img 
                src="/images/World_map_illuminated_at_night_202606052226.jpeg" 
                alt="Global Network Map" 
                className="cc-map-base"
              />
              
              {/* SVG Animated Routes */}
              <svg className="cc-routes-svg" viewBox="0 0 1000 600" preserveAspectRatio="none">
                <path className="cc-route-path delay-1" d="M 200 150 Q 400 50 600 200 T 850 100" />
                <path className="cc-route-path delay-2" d="M 150 300 Q 350 450 550 350 T 900 400" />
                <path className="cc-route-path delay-3" d="M 400 500 Q 500 250 700 300" />
                <path className="cc-route-path delay-4" d="M 300 200 Q 500 350 800 250" />
                
                {/* Nodes / Ports */}
                <circle cx="200" cy="150" r="4" className="cc-node" />
                <circle cx="600" cy="200" r="4" className="cc-node" />
                <circle cx="850" cy="100" r="4" className="cc-node" />
                <circle cx="150" cy="300" r="4" className="cc-node" />
                <circle cx="550" cy="350" r="4" className="cc-node" />
                <circle cx="900" cy="400" r="4" className="cc-node" />
              </svg>

              {/* Floating Data Cards (Glassmorphism) */}
              
              {/* Card 1: Status */}
              <div className="cc-glass-card card-status float-1">
                <div className="cc-card-header">Shipment Status</div>
                <div className="cc-card-value status-active">
                  <span className="status-dot"></span> In Transit
                </div>
              </div>

              {/* Card 2: Carrier */}
              <div className="cc-glass-card card-carrier float-2">
                <div className="cc-card-header">Carrier Perf.</div>
                <div className="cc-card-value text-green">98.2%</div>
              </div>

              {/* Card 3: Ports */}
              <div className="cc-glass-card card-ports float-3">
                <div className="cc-card-header">Port Activity</div>
                <div className="cc-card-value">147 <span className="cc-text-sm">Active</span></div>
              </div>

              {/* Card 4: Air */}
              <div className="cc-glass-card card-air float-4">
                <div className="cc-card-icon">✈️</div>
                <div>
                  <div className="cc-card-header">Air Freight</div>
                  <div className="cc-card-value-sm">24 Shipments</div>
                </div>
              </div>

              {/* Card 5: Ocean */}
              <div className="cc-glass-card card-ocean float-5">
                <div className="cc-card-icon">🚢</div>
                <div>
                  <div className="cc-card-header">Ocean Freight</div>
                  <div className="cc-card-value-sm">68 Containers</div>
                </div>
              </div>

              {/* Card 6: Road */}
              <div className="cc-glass-card card-road float-6">
                <div className="cc-card-icon">🚚</div>
                <div>
                  <div className="cc-card-header">Road Freight</div>
                  <div className="cc-card-value-sm">112 Deliveries</div>
                </div>
              </div>

            </div>
          </div>

        </div>

        {/* ─── BOTTOM KPI STRIP ─── */}
        <div className="cc-kpi-strip">
          <div className="cc-kpi-item">
            <div className="cc-kpi-val">500+</div>
            <div className="cc-kpi-label">Verified Carriers</div>
          </div>
          <div className="cc-kpi-divider"></div>
          <div className="cc-kpi-item">
            <div className="cc-kpi-val">150+</div>
            <div className="cc-kpi-label">Global Ports</div>
          </div>
          <div className="cc-kpi-divider"></div>
          <div className="cc-kpi-item">
            <div className="cc-kpi-val">10K+</div>
            <div className="cc-kpi-label">Trade Lanes</div>
          </div>
          <div className="cc-kpi-divider"></div>
          <div className="cc-kpi-item">
            <div className="cc-kpi-val text-green">98%</div>
            <div className="cc-kpi-label">On-Time Rate</div>
          </div>
        </div>

      </div>
    </section>
  );
}

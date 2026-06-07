import React, { useEffect, useRef } from 'react';
import './PlatformCapabilities.css';

/* ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
   PLATFORM CAPABILITIES — Clean Light SaaS Showcase
   ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━ */

export default function PlatformCapabilities() {
  const sectionRef = useRef(null);
  const blockRefs = useRef([]);

  useEffect(() => {
    // Observer for section header
    const sectionObserver = new IntersectionObserver(
      (entries) => {
        entries.forEach((entry) => {
          if (entry.isIntersecting) {
            entry.target.classList.add('pc-visible');
          }
        });
      },
      { threshold: 0.1 }
    );

    if (sectionRef.current) {
      sectionObserver.observe(sectionRef.current);
    }

    // Observer for individual blocks to fade up
    const blockObserver = new IntersectionObserver(
      (entries) => {
        entries.forEach((entry) => {
          if (entry.isIntersecting) {
            entry.target.classList.add('block-visible');
          }
        });
      },
      { threshold: 0.15 }
    );

    blockRefs.current.forEach(block => {
      if (block) blockObserver.observe(block);
    });

    return () => {
      sectionObserver.disconnect();
      blockObserver.disconnect();
    };
  }, []);

  return (
    <section className="pc-light-section" ref={sectionRef}>
      <div className="container-lg">
        
        {/* ─── SECTION HEADER ─── */}
        <div className="pc-light-header">
          <h2 className="pc-light-title">Everything You Need To Move Freight Smarter</h2>
          <p className="pc-light-subtitle">
            Powerful logistics tools designed to improve visibility, automate operations and help supply chains move with confidence.
          </p>
        </div>

        {/* ─── FEATURE BLOCKS ─── */}
        <div className="pc-blocks">

          {/* Feature 1 */}
          <div className="pc-block" ref={el => blockRefs.current[0] = el}>
            <div className="pc-block-image">
              {/* TODO: Download this image from Nano Banana */}
              <img src="/images/platform/tracking-dashboard.webp" alt="Real-Time Shipment Tracking" />
            </div>
            <div className="pc-block-content">
              <h3>Real-Time Shipment Tracking</h3>
              <p>Track shipments across air, sea and road freight from one operational dashboard.</p>
              <ul className="pc-check-list">
                <li><span className="check">✓</span> Live Tracking</li>
                <li><span className="check">✓</span> Route Visibility</li>
                <li><span className="check">✓</span> Customs Updates</li>
                <li><span className="check">✓</span> Delivery Notifications</li>
              </ul>
            </div>
          </div>

          {/* Feature 2 */}
          <div className="pc-block" ref={el => blockRefs.current[1] = el}>
            <div className="pc-block-image">
              {/* TODO: Download this image from Nano Banana */}
              <img src="/images/platform/document-management.webp" alt="Smart Documentation" />
            </div>
            <div className="pc-block-content">
              <h3>Smart Documentation</h3>
              <p>Digitize customs paperwork, invoices and shipment records from a centralized workspace.</p>
              <ul className="pc-check-list">
                <li><span className="check">✓</span> Digital Documents</li>
                <li><span className="check">✓</span> Approval Workflows</li>
                <li><span className="check">✓</span> Compliance Records</li>
                <li><span className="check">✓</span> Secure Storage</li>
              </ul>
            </div>
          </div>

          {/* Feature 3 */}
          <div className="pc-block" ref={el => blockRefs.current[2] = el}>
            <div className="pc-block-image">
              {/* TODO: Download this image from Nano Banana */}
              <img src="/images/platform/carrier-network.webp" alt="Carrier Management" />
            </div>
            <div className="pc-block-content">
              <h3>Carrier Management</h3>
              <p>Monitor transportation partners, delivery reliability and operational performance.</p>
              <ul className="pc-check-list">
                <li><span className="check">✓</span> Carrier Performance</li>
                <li><span className="check">✓</span> Delivery Reliability</li>
                <li><span className="check">✓</span> Active Routes</li>
                <li><span className="check">✓</span> Network Monitoring</li>
              </ul>
            </div>
          </div>

          {/* Feature 4 */}
          <div className="pc-block" ref={el => blockRefs.current[3] = el}>
            <div className="pc-block-image">
              {/* TODO: Download this image from Nano Banana */}
              <img src="/images/platform/analytics.webp" alt="Analytics & Insights" />
            </div>
            <div className="pc-block-content">
              <h3>Analytics & Insights</h3>
              <p>Transform logistics data into actionable business intelligence and forecasting.</p>
              <ul className="pc-check-list">
                <li><span className="check">✓</span> Forecasting</li>
                <li><span className="check">✓</span> KPI Monitoring</li>
                <li><span className="check">✓</span> Performance Analytics</li>
                <li><span className="check">✓</span> Business Intelligence</li>
              </ul>
            </div>
          </div>

        </div>
      </div>
    </section>
  );
}

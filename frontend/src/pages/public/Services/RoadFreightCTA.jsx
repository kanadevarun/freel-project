import React, { useEffect, useRef, useState } from 'react';
import { Link } from 'react-router-dom';
import './RoadFreightCTA.css';

const STARTING_COUNT = 1247;
const INCREMENT_INTERVAL_MS = 2800; // A new shipment every ~2.8s

export default function RoadFreightCTA() {
  const sectionRef = useRef(null);
  const [isActive, setIsActive] = useState(false);
  const [shipmentCount, setShipmentCount] = useState(STARTING_COUNT);

  // Intersection observer to trigger animations
  useEffect(() => {
    const observer = new IntersectionObserver(
      ([entry]) => {
        if (entry.isIntersecting) setIsActive(true);
      },
      { threshold: 0.2 }
    );
    if (sectionRef.current) observer.observe(sectionRef.current);
    return () => observer.disconnect();
  }, []);

  // Continuously incrementing live shipment counter
  useEffect(() => {
    if (!isActive) return;
    const timer = setInterval(() => {
      setShipmentCount(prev => prev + 1);
    }, INCREMENT_INTERVAL_MS);
    return () => clearInterval(timer);
  }, [isActive]);

  const trustItems = [
    '500+ Verified Transporters',
    '24/7 Control Tower',
    '100% Digital Compliance',
    'Nationwide Coverage',
  ];

  return (
    <section ref={sectionRef} className={`cta-hero ${isActive ? 'is-active' : ''}`}>
      {/* Background Hero Image */}
      <img
        src="/images/endless_highway_sunrise.png"
        alt="Freight corridor at sunrise"
        className="cta-hero-img"
      />

      {/* Gradient overlay */}
      <div className="cta-overlay" />

      {/* Content */}
      <div className="cta-content">
        <div className="max-w-7xl mx-auto flex flex-col items-start" style={{ paddingTop: '20vh' }}>
          {/* Label */}
          <span className="cta-label">The Operating System Of Freight</span>

          {/* Headline */}
          <h2 className="cta-headline">
            India Moves Forward.<br />
            So Should Your Supply Chain.
          </h2>

          {/* Subtext */}
          <p className="cta-subtext">
            From first mile pickup to final delivery, Freel connects transporters,
            businesses, and operations teams through one intelligent logistics network.
            <br /><br />
            Built for visibility.
            Designed for reliability.
            Ready to scale.
          </p>

          {/* Buttons */}
          <div className="cta-buttons">
            <Link to="/contact" className="cta-btn-primary">
              Start Shipping →
            </Link>
            <Link to="/contact" className="cta-btn-secondary">
              Talk To Logistics Expert
            </Link>
          </div>

          {/* Trust Bar */}
          <div className="cta-trust-bar">
            {trustItems.map((item, i) => (
              <React.Fragment key={item}>
                <span className="cta-trust-item">{item}</span>
                {i < trustItems.length - 1 && <div className="cta-trust-divider" />}
              </React.Fragment>
            ))}
          </div>

          {/* Live Network Indicator */}
          <div className="cta-live-indicator">
            <div className="cta-live-dot" />
            <span className="cta-live-text">While you explored this page —</span>
            <span className="cta-live-count">
              {shipmentCount.toLocaleString()} Shipments Moved Across India
            </span>
          </div>
        </div>
      </div>
    </section>
  );
}

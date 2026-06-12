import React, { useEffect, useRef, useState } from 'react';
import { Link } from 'react-router-dom';
import './RoadFreightCTA.css';

const trustItems = [
  '500+ Verified Transporters',
  '24/7 Operations Visibility',
  '100% Digital Compliance',
  'Nationwide Coverage',
  'Real-Time Tracking',
];

const supportItems = [
  'Transporters',
  'Warehouses',
  'Control Towers',
  'Tracking',
  'Compliance',
];

export default function RoadFreightCTA() {
  const ref = useRef(null);
  const [isActive, setIsActive] = useState(false);

  useEffect(() => {
    const observer = new IntersectionObserver(
      ([entry]) => { if (entry.isIntersecting) setIsActive(true); },
      { threshold: 0.2 }
    );
    if (ref.current) observer.observe(ref.current);
    return () => observer.disconnect();
  }, []);

  return (
    <section ref={ref} className={`cta-hero ${isActive ? 'is-active' : ''}`}>
      {/* Cinematic background */}
      <img
        src="/images/endless_highway_sunrise.png"
        alt="Freight corridor at sunrise"
        className="cta-hero-img"
      />
      <div className="cta-overlay" />

      {/* Centered content */}
      <div className="cta-content">

        {/* Eyebrow */}
        <span className="cta-eyebrow">The Operating System Of Freight</span>

        {/* Headline */}
        <h2 className="cta-headline">
          The Trucks Move Freight.<br />
          Freel Moves The Entire Network.
        </h2>

        {/* Supporting text */}
        <div className="cta-support">
          <div className="cta-support-items">
            {supportItems.map((item, i) => (
              <React.Fragment key={item}>
                <span className="cta-support-item">{item}</span>
                {i < supportItems.length - 1 && (
                  <span className="cta-support-sep">·</span>
                )}
              </React.Fragment>
            ))}
          </div>
          <p className="cta-support-tagline">
            One platform connecting every moving part of road logistics.
          </p>
        </div>

        {/* CTA Buttons */}
        <div className="cta-buttons">
          <Link to="/contact" className="cta-btn-primary">
            Start Shipping →
          </Link>
          <Link to="/contact" className="cta-btn-secondary">
            Book A Demo
          </Link>
        </div>

        {/* Trust Bar */}
        <div className="cta-trust-bar">
          {trustItems.map(item => (
            <span key={item} className="cta-trust-item">{item}</span>
          ))}
        </div>

      </div>
    </section>
  );
}

import React, { useEffect, useRef, useState } from 'react';
import './JourneyBridge.css';

const steps = [
  { icon: '🏭', label: 'Factory' },
  { icon: '📦', label: 'Warehouse' },
  { icon: '🚚', label: 'Transport' },
  { icon: '📍', label: 'Tracking' },
  { icon: '✅', label: 'Delivered' },
];

export default function JourneyBridge() {
  const ref = useRef(null);
  const [isActive, setIsActive] = useState(false);

  useEffect(() => {
    const observer = new IntersectionObserver(
      ([entry]) => { if (entry.isIntersecting) setIsActive(true); },
      { threshold: 0.3 }
    );
    if (ref.current) observer.observe(ref.current);
    return () => observer.disconnect();
  }, []);

  return (
    <section ref={ref} className={`jb-section ${isActive ? 'is-active' : ''}`}>

      {/* Headline */}
      <h2 className="jb-headline">
        One Shipment.<br />Thousands Of Touchpoints.
      </h2>

      {/* Steps Row */}
      <div className="jb-steps">
        {steps.map((step, i) => (
          <React.Fragment key={step.label}>
            <div className="jb-step">
              <div className="jb-icon-circle">{step.icon}</div>
              <span className="jb-step-label">{step.label}</span>
            </div>
            {i < steps.length - 1 && (
              <div className="jb-arrow">
                <div className="jb-arrow-line" />
                <div className="jb-arrow-head" />
              </div>
            )}
          </React.Fragment>
        ))}
      </div>

      {/* Scroll Indicator */}
      <div className="jb-scroll-indicator">
        <span className="jb-scroll-label">Scroll To Follow The Journey</span>
        <div className="jb-bounce-arrow">
          <div className="jb-chevron" />
          <div className="jb-chevron" />
        </div>
      </div>

    </section>
  );
}

import React, { useState, useEffect } from 'react';
import './SplashScreen.css';

export default function SplashScreen({ onComplete }) {
  const [stage, setStage] = useState('initial'); // 'initial' -> 'logo' -> 'text' -> 'tagline' -> 'network' -> 'exit'
  const [statusIdx, setStatusIdx] = useState(0);

  const statuses = [
    'Initializing Freight Command OS...',
    'Syncing Multi-Modal Carrier Rates...',
    'Connecting AIS Vessel & Flight Tracking...',
    'Workspace Ready'
  ];

  useEffect(() => {
    // Check if user prefers reduced motion
    const prefersReducedMotion = window.matchMedia('(prefers-reduced-motion: reduce)').matches;
    if (prefersReducedMotion) {
      const timer = setTimeout(() => {
        onComplete();
      }, 400);
      return () => clearTimeout(timer);
    }

    const t1 = setTimeout(() => { setStage('logo'); setStatusIdx(0); }, 120);
    const t2 = setTimeout(() => { setStage('text'); setStatusIdx(1); }, 550);
    const t3 = setTimeout(() => { setStage('tagline'); setStatusIdx(2); }, 1000);
    const t4 = setTimeout(() => { setStage('network'); setStatusIdx(3); }, 1500);
    const t5 = setTimeout(() => setStage('exit'), 2050);
    const t6 = setTimeout(() => {
      onComplete();
    }, 2450);

    return () => {
      clearTimeout(t1);
      clearTimeout(t2);
      clearTimeout(t3);
      clearTimeout(t4);
      clearTimeout(t5);
      clearTimeout(t6);
    };
  }, [onComplete]);

  return (
    <div className={`lhq-splash-overlay ${stage === 'exit' ? 'fade-out' : ''}`}>
      <div className="lhq-splash-ambient-glow" />
      <div className="lhq-splash-grid-mesh" />
      
      <div className="lhq-splash-content">
        {/* 3D Logo Container with Glowing Aura */}
        <div className={`lhq-splash-logo-box ${stage !== 'initial' ? 'visible' : ''}`}>
          <img
            src="/images/logo/logo.png"
            alt="LogisticsHQ"
            className="lhq-splash-logo-img"
          />
          <div className="lhq-splash-orbit-pulse" />
        </div>

        {/* Brand Title */}
        <h1 className={`lhq-splash-title ${stage === 'text' || stage === 'tagline' || stage === 'network' || stage === 'exit' ? 'visible' : ''}`}>
          LogisticsHQ
        </h1>

        {/* Tagline */}
        <p className={`lhq-splash-tagline ${stage === 'tagline' || stage === 'network' || stage === 'exit' ? 'visible' : ''}`}>
          The Operating System for Global Logistics
        </p>

        {/* Telemetry Progress Track */}
        <div className={`lhq-splash-progress-wrapper ${stage === 'network' || stage === 'exit' || stage === 'tagline' ? 'visible' : ''}`}>
          <div className="lhq-splash-progress-track">
            <div className="lhq-splash-progress-beam" />
          </div>
          <div className="lhq-splash-status-text">
            <span className="lhq-status-beacon" />
            <span>{statuses[statusIdx]}</span>
          </div>
        </div>

        {/* Transport Modes Badges */}
        <div className={`lhq-splash-modes ${stage === 'network' || stage === 'exit' ? 'visible' : ''}`}>
          <span className="lhq-mode-chip">✈️ Air</span>
          <span className="lhq-mode-chip">🚢 Ocean</span>
          <span className="lhq-mode-chip">🚛 Road</span>
          <span className="lhq-mode-chip">🏢 Customs</span>
        </div>
      </div>
    </div>
  );
}

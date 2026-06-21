import React from 'react';
import { motion } from 'framer-motion';
import { ArrowRight, Shield, Globe2, Activity, MapPin } from 'lucide-react';
import './TrackingCTA.css';

export default function TrackingCTA() {
  return (
    <section className="tcta-section">
      {/* ─── Premium Image Background ─── */}
      <div 
        className="tcta-bg-image" 
        style={{ backgroundImage: "url('/images/hero_port_aerial.png')" }}
      ></div>
      <div className="tcta-bg-overlay"></div>

      <div className="tcta-container">
        
        {/* ─── HEADER CONTENT ─── */}
        <motion.div 
          className="tcta-header"
          initial={{ opacity: 0, y: 30 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.8 }}
        >
          <div className="tcta-eyebrow">FREIGHT VISIBILITY REIMAGINED</div>
          <h2 className="tcta-headline">
            Every Shipment.<br/>
            Every Milestone.<br/>
            One Source Of Truth.
          </h2>
        </motion.div>

        {/* ─── INLINE STATS GRID ─── */}
        <motion.div 
          className="tcta-stats-grid"
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.8, delay: 0.2 }}
        >
          <div className="tcta-stat-item">
            <span className="tcta-stat-val">98%</span>
            <span className="tcta-stat-lbl">ETA Accuracy</span>
          </div>
          <div className="tcta-stat-item">
            <span className="tcta-stat-val">24/7</span>
            <span className="tcta-stat-lbl">Shipment Visibility</span>
          </div>
          <div className="tcta-stat-item">
            <span className="tcta-stat-val">42%</span>
            <span className="tcta-stat-lbl">Fewer Delays</span>
          </div>
          <div className="tcta-stat-item">
            <span className="tcta-stat-val">10M+</span>
            <span className="tcta-stat-lbl">Tracking Events</span>
          </div>
        </motion.div>

        {/* ─── ACTIONS ─── */}
        <motion.div 
          className="tcta-actions"
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.6, delay: 0.4 }}
        >
          <a href="/signup" className="tcta-btn-primary">
            Start Free <ArrowRight size={18} className="tcta-btn-icon" />
          </a>
          <a href="/demo" className="tcta-btn-secondary">
            Book A Demo
          </a>
        </motion.div>

        {/* ─── TRUST BAR ─── */}
        <motion.div 
          className="tcta-trust-bar"
          initial={{ opacity: 0 }}
          whileInView={{ opacity: 1 }}
          viewport={{ once: true }}
          transition={{ duration: 0.8, delay: 0.6 }}
        >
          <div className="tcta-trust-badge"><Shield size={14} /> No Credit Card Required</div>
          <div className="tcta-trust-badge"><Globe2 size={14} /> Global Freight Coverage</div>
          <div className="tcta-trust-badge"><Activity size={14} /> Real-Time Visibility</div>
          <div className="tcta-trust-badge"><MapPin size={14} /> Enterprise Ready</div>
        </motion.div>

      </div>
    </section>
  );
}

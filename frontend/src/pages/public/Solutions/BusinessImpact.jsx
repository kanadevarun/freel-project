import React from 'react';
import { motion } from 'framer-motion';
import { Zap, MapPin, Settings, DollarSign } from 'lucide-react';
import './BusinessImpact.css';

const OUTCOMES = [
  {
    id: 1,
    title: "Reduce Delays",
    description: "Predict issues before they impact delivery.",
    metric: "42%",
    icon: Zap
  },
  {
    id: 2,
    title: "Customer Visibility",
    description: "Real-time updates across every shipment.",
    metric: "98%",
    icon: MapPin
  },
  {
    id: 3,
    title: "Operational Efficiency",
    description: "Centralize tracking workflows automatically.",
    metric: "3x",
    icon: Settings
  },
  {
    id: 4,
    title: "Optimize Spend",
    description: "Identify inefficiencies and improve utilization.",
    metric: "18%",
    icon: DollarSign
  }
];

export default function BusinessImpact() {
  return (
    <section className="bi-section">
      {/* Animated Subtle Grid Background */}
      <div className="bi-grid-bg"></div>

      <div className="bi-container">
        
        {/* HEADER */}
        <motion.div 
          className="bi-header"
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.6 }}
        >
          <div className="bi-eyebrow">BUSINESS IMPACT</div>
          <h2 className="bi-headline">From Visibility To Logistics Intelligence</h2>
          <p className="bi-subheading">
            Monitor shipments, predict disruptions, and optimize operations from a single intelligent platform.
          </p>
        </motion.div>

        {/* TWO COLUMN LAYOUT */}
        <div className="bi-content-grid">
          
          {/* Subtle Connector Lines connecting Dashboard to Cards */}
          <svg className="bi-connector-lines" viewBox="0 0 1000 600" preserveAspectRatio="none">
            <motion.path 
              d="M 600 200 C 750 200, 800 150, 1000 150" 
              className="bi-c-line"
              initial={{ pathLength: 0, opacity: 0 }}
              whileInView={{ pathLength: 1, opacity: 0.15 }}
              transition={{ duration: 2, ease: "easeOut" }}
            />
            <motion.path 
              d="M 600 400 C 750 400, 800 450, 1000 450" 
              className="bi-c-line"
              initial={{ pathLength: 0, opacity: 0 }}
              whileInView={{ pathLength: 1, opacity: 0.15 }}
              transition={{ duration: 2, ease: "easeOut", delay: 0.3 }}
            />
          </svg>

          {/* LEFT: Dashboard Visual (HERO) */}
          <motion.div 
            className="bi-visual-col"
            initial={{ opacity: 0, x: -40 }}
            whileInView={{ opacity: 1, x: 0 }}
            viewport={{ once: true, margin: "-100px" }}
            transition={{ duration: 0.8, ease: "easeOut" }}
          >
            <div className="bi-visual-wrapper">
              <video 
                src="/videos/shipment-tracking/Control_tower_monitoring_dashboa…_202606161854.mp4" 
                autoPlay 
                loop 
                muted 
                playsInline 
                className="bi-main-video"
              />
              <div className="bi-visual-overlay"></div>

              {/* Animated Route Lines & Data Points over the Dashboard */}
              <div className="bi-dashboard-effects">
                <svg viewBox="0 0 800 600" className="bi-dash-svg">
                  {/* Subtle pulsing data point */}
                  <circle cx="200" cy="300" r="4" className="bi-dash-node" />
                  <circle cx="200" cy="300" r="15" className="bi-dash-node-pulse" />
                  
                  <circle cx="600" cy="200" r="4" className="bi-dash-node" />
                  <circle cx="600" cy="200" r="15" className="bi-dash-node-pulse" style={{ animationDelay: '1s' }} />

                  {/* Route connection */}
                  <path d="M 200 300 Q 400 150, 600 200" className="bi-dash-route" />
                </svg>
              </div>

              {/* Floating Glass Metrics */}
              <motion.div 
                className="bi-floating-metric float-1"
                animate={{ y: [0, -6, 0] }}
                transition={{ duration: 4, repeat: Infinity, ease: "easeInOut" }}
              >
                <span className="bi-fm-value">98.7%</span>
                <span className="bi-fm-label">On-Time Delivery</span>
              </motion.div>

              <motion.div 
                className="bi-floating-metric float-2"
                animate={{ y: [0, 4, 0] }}
                transition={{ duration: 5, repeat: Infinity, ease: "easeInOut", delay: 1 }}
              >
                <span className="bi-fm-value">42%</span>
                <span className="bi-fm-label">Fewer Exceptions</span>
              </motion.div>

            </div>
          </motion.div>

          {/* RIGHT: Outcome Cards (35% width, bold metrics) */}
          <div className="bi-outcomes-col">
            {OUTCOMES.map((outcome, index) => (
              <motion.div 
                key={outcome.id}
                className="bi-outcome-card"
                initial={{ opacity: 0, x: 40 }}
                whileInView={{ opacity: 1, x: 0 }}
                viewport={{ once: true, margin: "-100px" }}
                transition={{ duration: 0.6, delay: index * 0.1, ease: "easeOut" }}
              >
                <div className="bi-outcome-header">
                  <div className="bi-outcome-icon-wrapper">
                    <outcome.icon size={20} className="bi-outcome-icon" />
                  </div>
                  <h3 className="bi-outcome-title">{outcome.title}</h3>
                </div>
                <p className="bi-outcome-desc">{outcome.description}</p>
                <div className="bi-outcome-huge-metric">
                  {outcome.metric}
                </div>
              </motion.div>
            ))}
          </div>

        </div>

      </div>
    </section>
  );
}

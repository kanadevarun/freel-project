import React from 'react';
import { motion } from 'framer-motion';
import { Clock, AlertTriangle, Route, ShieldAlert, Sparkles, CloudLightning, Ship, Truck, Activity } from 'lucide-react';
import './LogisticsIntelligence.css';

const CAPABILITIES = [
  {
    id: 1,
    title: "Predictive ETA",
    description: "AI continuously recalculates arrival times based on live operational conditions.",
    metric: "98% Accuracy",
    icon: Clock
  },
  {
    id: 2,
    title: "Delay Detection",
    description: "Identify disruptions before they impact delivery commitments.",
    metric: "42% Faster Response",
    icon: AlertTriangle
  },
  {
    id: 3,
    title: "Route Optimization",
    description: "Recommend better routing options using real-time logistics intelligence.",
    metric: "18% Cost Savings",
    icon: Route
  },
  {
    id: 4,
    title: "Exception Monitoring",
    description: "Automatically surface critical shipment risks and operational issues.",
    metric: "24/7 Monitoring",
    icon: ShieldAlert
  },
  {
    id: 5,
    title: "Smart Recommendations",
    description: "Receive actionable recommendations instead of raw shipment data.",
    metric: "Instant Insights",
    icon: Sparkles
  }
];

export default function LogisticsIntelligence() {
  return (
    <section className="lintel-section">
      {/* Subtle Background Pattern */}
      <div className="lintel-bg-pattern"></div>
      
      <div className="lintel-container">
        
        {/* HEADER */}
        <motion.div 
          className="lintel-header"
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.6 }}
        >
          <div className="lintel-eyebrow">INTELLIGENCE LAYER</div>
          <h2 className="lintel-headline">The Intelligence Behind Every Shipment</h2>
          <p className="lintel-subheading">
            Transform millions of shipment events into actionable insights, predictive alerts, and operational recommendations.
          </p>
        </motion.div>

        {/* 60/40 SPLIT LAYOUT */}
        <div className="lintel-grid">
          
          {/* LEFT: Massive Visual Hero */}
          <motion.div 
            className="lintel-visual-col"
            initial={{ opacity: 0, x: -40 }}
            whileInView={{ opacity: 1, x: 0 }}
            viewport={{ once: true, margin: "-100px" }}
            transition={{ duration: 0.8, ease: "easeOut" }}
          >
            <div className="lintel-visual-container">
              <video 
                src="/videos/shipment-tracking/Global_visibility_trade_routes_a…_202606161858.mp4" 
                autoPlay 
                loop 
                muted 
                playsInline 
                className="lintel-video"
              />
              <div className="lintel-video-overlay"></div>

              {/* Floating Intelligence Signals */}
              <div className="lintel-signals">
                
                <motion.div className="lintel-signal sig-1" animate={{ y: [0, -8, 0] }} transition={{ duration: 4, repeat: Infinity, ease: "easeInOut" }}>
                  <AlertTriangle size={14} className="sig-icon text-orange" />
                  <span>Delay Risk Detected</span>
                </motion.div>

                <motion.div className="lintel-signal sig-2" animate={{ y: [0, 6, 0] }} transition={{ duration: 5, repeat: Infinity, ease: "easeInOut", delay: 1 }}>
                  <Route size={14} className="sig-icon text-green" />
                  <span>Route Optimization Found</span>
                </motion.div>

                <motion.div className="lintel-signal sig-3" animate={{ y: [0, -5, 0] }} transition={{ duration: 4.5, repeat: Infinity, ease: "easeInOut", delay: 2 }}>
                  <Clock size={14} className="sig-icon text-blue" />
                  <span>ETA Improved</span>
                </motion.div>

                <motion.div className="lintel-signal sig-4" animate={{ y: [0, 8, 0] }} transition={{ duration: 5.5, repeat: Infinity, ease: "easeInOut", delay: 1.5 }}>
                  <CloudLightning size={14} className="sig-icon text-red" />
                  <span>Weather Alert</span>
                </motion.div>

                <motion.div className="lintel-signal sig-5" animate={{ y: [0, -7, 0] }} transition={{ duration: 6, repeat: Infinity, ease: "easeInOut", delay: 0.5 }}>
                  <Ship size={14} className="sig-icon text-orange" />
                  <span>Port Congestion</span>
                </motion.div>

                <motion.div className="lintel-signal sig-6" animate={{ y: [0, 5, 0] }} transition={{ duration: 4.8, repeat: Infinity, ease: "easeInOut", delay: 2.5 }}>
                  <Truck size={14} className="sig-icon text-indigo" />
                  <span>Carrier Recommendation</span>
                </motion.div>

              </div>
            </div>
          </motion.div>

          {/* RIGHT: Stacked Capability Blocks */}
          <div className="lintel-content-col">
            {CAPABILITIES.map((cap, index) => (
              <motion.div 
                key={cap.id}
                className="lintel-cap-block"
                initial={{ opacity: 0, x: 40 }}
                whileInView={{ opacity: 1, x: 0 }}
                viewport={{ once: true, margin: "-100px" }}
                transition={{ duration: 0.5, delay: index * 0.1, ease: "easeOut" }}
              >
                <div className="lintel-cap-icon">
                  <cap.icon size={20} />
                </div>
                
                <div className="lintel-cap-body">
                  <h3 className="lintel-cap-title">{cap.title}</h3>
                  <p className="lintel-cap-desc">{cap.description}</p>
                </div>

                <div className="lintel-cap-metric">
                  {cap.metric}
                </div>
              </motion.div>
            ))}
          </div>

        </div>
      </div>
    </section>
  );
}

import React from 'react';
import { motion } from 'framer-motion';
import { ArrowRight, Check } from 'lucide-react';
import './TrackingCapabilities.css';

const CAPABILITIES = [
  {
    id: 1,
    badge: "Live Tracking",
    title: "Live Location Tracking",
    description: "Track every shipment movement across road, ocean, air, and warehouse networks in real time.",
    features: ["Live GPS Data", "Route Monitoring", "Carrier Visibility"],
    media: "/videos/shipment-tracking/Global_visibility_trade_routes_a…_202606161858.mp4",
    type: "video"
  },
  {
    id: 2,
    badge: "ETA Intelligence",
    title: "Predictive ETA Engine",
    description: "AI continuously recalculates delivery estimates based on traffic, weather, port congestion, and operational conditions.",
    features: ["Dynamic ETA", "Delay Prediction", "Route Optimization"],
    media: "/videos/shipment-tracking/Logistics_dashboard_with_animate…_202606161822.mp4",
    type: "video"
  },
  {
    id: 3,
    badge: "Exception Control",
    title: "Exception Management",
    description: "Identify disruptions before they impact delivery performance and respond proactively.",
    features: ["Delay Detection", "Weather Alerts", "Escalation Workflows"],
    media: "/images/shipment-tracking/Featureless_figure_reviewing_log…_202606161832.jpeg",
    type: "image"
  },
  {
    id: 4,
    badge: "Customer Updates",
    title: "Customer Notifications",
    description: "Automatically inform customers and stakeholders at every shipment milestone.",
    features: ["SMS Updates", "Email Alerts", "Milestone Notifications"],
    media: "/videos/shipment-tracking/Customer_visibility_shipment_tra…_202606161853.mp4",
    type: "video"
  },
  {
    id: 5,
    badge: "Control Tower",
    title: "Control Tower Analytics",
    description: "Monitor global freight operations from a centralized logistics command center.",
    features: ["Global Visibility", "KPI Monitoring", "Performance Analytics"],
    media: "/videos/shipment-tracking/Control_tower_monitoring_dashboa…_202606161854.mp4",
    type: "video"
  },
  {
    id: 6,
    badge: "Delivery Verification",
    title: "Proof Of Delivery",
    description: "Capture delivery confirmation and maintain a complete shipment audit trail.",
    features: ["Digital POD", "Signature Capture", "Audit Records"],
    media: "/videos/shipment-tracking/Delivery_agent_confirms_package_…_202606161851.mp4",
    type: "video"
  }
];

export default function TrackingCapabilities() {
  return (
    <section className="tc-section">
      <div className="tc-container">
        
        {/* HEADER */}
        <div className="tc-header">
          <div className="tc-label">REAL-TIME TRACKING INTELLIGENCE</div>
          <h2 className="tc-title">
            Everything You Need To Monitor <span className="tc-gradient-text">Freight</span>
          </h2>
          <p className="tc-subtitle">
            Track shipments, predict delays, manage exceptions, notify customers, and gain complete operational visibility from a single platform.
          </p>
          
          <div className="tc-trust-metrics">
            <div className="tc-trust-pill">12,458 Active Shipments</div>
            <div className="tc-trust-pill">184 Countries Connected</div>
            <div className="tc-trust-pill">98.7% On-Time Performance</div>
          </div>
        </div>

        {/* CAPABILITIES GRID */}
        <div className="tc-grid">
          {CAPABILITIES.map((cap, index) => (
            <motion.div 
              key={cap.id}
              className="tc-card"
              initial={{ opacity: 0, y: 30 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true, margin: "-100px" }}
              transition={{ duration: 0.6, delay: index * 0.1, ease: "easeOut" }}
            >
              <div className="tc-card-media-wrapper">
                {cap.type === 'video' ? (
                  <video 
                    src={cap.media} 
                    autoPlay 
                    loop 
                    muted 
                    playsInline 
                    className="tc-card-media"
                  />
                ) : (
                  <img 
                    src={cap.media} 
                    alt={cap.title} 
                    className="tc-card-media"
                  />
                )}
              </div>
              <div className="tc-card-content">
                <div className="tc-card-badge">{cap.badge}</div>
                <h3 className="tc-card-title">{cap.title}</h3>
                <p className="tc-card-desc">{cap.description}</p>
                <div className="tc-features-wrapper">
                  {cap.features.map((feature, i) => (
                    <div key={i} className="tc-feature-pill">
                      <Check size={16} className="tc-feature-check" />
                      {feature}
                    </div>
                  ))}
                </div>
                
                {/* Arrow Icon Top Right */}
                <div className="tc-card-arrow">
                  <ArrowRight size={20} />
                </div>
              </div>
            </motion.div>
          ))}
        </div>

      </div>
    </section>
  );
}

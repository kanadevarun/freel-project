import React from 'react';
import { motion } from 'framer-motion';
import { MapPin, Clock, AlertTriangle, Activity, Navigation, CheckCircle2, CloudLightning } from 'lucide-react';
import './LogisticsControlCenter.css';

const CARDS = [
  {
    id: 1,
    title: "Live Tracking",
    description: "Track every shipment across air, sea, road, and warehouse networks in real time.",
    metric: "10M+",
    metricSub: "Location Events",
    icon: MapPin
  },
  {
    id: 2,
    title: "Predictive ETA",
    description: "AI forecasts delivery times using live traffic, weather, and operational conditions.",
    metric: "98%",
    metricSub: "ETA Accuracy",
    icon: Clock
  },
  {
    id: 3,
    title: "Exception Management",
    description: "Identify disruptions early and resolve issues before they impact customers.",
    metric: "42%",
    metricSub: "Fewer Delays",
    icon: AlertTriangle
  },
  {
    id: 4,
    title: "Control Tower",
    description: "Centralized visibility across shipments, carriers, warehouses, and customers.",
    metric: "24/7",
    metricSub: "Visibility",
    icon: Activity
  }
];

export default function LogisticsControlCenter() {
  return (
    <section className="lcc-section">
      {/* Premium Modern SaaS Background */}
      <div className="lcc-bg-pattern"></div>
      <div className="lcc-bg-glow"></div>

      <div className="lcc-container">
        
        {/* HEADER */}
        <motion.div 
          className="lcc-header"
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.6 }}
        >
          <div className="lcc-eyebrow">PRODUCT EXPERIENCE</div>
          <h2 className="lcc-headline">Freight Intelligence In Action</h2>
          <p className="lcc-subheading">
            Monitor shipments, predict disruptions, coordinate carriers, and manage operations from a single intelligent logistics platform.
          </p>
        </motion.div>

        {/* TOP SECTION: Massive Dashboard Showcase */}
        <motion.div 
          className="lcc-dashboard-showcase"
          initial={{ opacity: 0, y: 40, scale: 0.98 }}
          whileInView={{ opacity: 1, y: 0, scale: 1 }}
          viewport={{ once: true, margin: "-100px" }}
          transition={{ duration: 0.8, ease: "easeOut" }}
        >
          <video 
            src="/videos/shipment-tracking/Logistics_dashboard_with_animate…_202606161822.mp4" 
            autoPlay 
            loop 
            muted 
            playsInline 
            className="lcc-showcase-video"
          />
          <div className="lcc-showcase-overlay"></div>

          {/* Floating UI Elements (Staggered Reveal) */}
          <div className="lcc-floating-elements">
            
            <motion.div 
              className="lcc-float-pill lcc-float-1"
              animate={{ y: [0, -6, 0] }}
              transition={{ duration: 4, repeat: Infinity, ease: "easeInOut" }}
            >
              <div className="lcc-float-icon-wrap bg-blue"><Navigation size={14} /></div>
              <div className="lcc-float-text">
                <span className="lcc-f-title">Shipment #84721</span>
                <span className="lcc-f-sub">On Schedule</span>
              </div>
            </motion.div>

            <motion.div 
              className="lcc-float-pill lcc-float-2"
              animate={{ y: [0, 8, 0] }}
              transition={{ duration: 5, repeat: Infinity, ease: "easeInOut", delay: 0.5 }}
            >
              <div className="lcc-float-icon-wrap bg-cyan"><Clock size={14} /></div>
              <div className="lcc-float-text">
                <span className="lcc-f-title">ETA Updated</span>
                <span className="lcc-f-sub">14:35</span>
              </div>
            </motion.div>

            <motion.div 
              className="lcc-float-pill lcc-float-3"
              animate={{ y: [0, -5, 0] }}
              transition={{ duration: 4.5, repeat: Infinity, ease: "easeInOut", delay: 1 }}
            >
              <div className="lcc-float-icon-wrap bg-green"><CheckCircle2 size={14} /></div>
              <div className="lcc-float-text">
                <span className="lcc-f-title">Port Cleared</span>
                <span className="lcc-f-sub">Singapore</span>
              </div>
            </motion.div>

            <motion.div 
              className="lcc-float-pill lcc-float-4"
              animate={{ y: [0, 7, 0] }}
              transition={{ duration: 5.5, repeat: Infinity, ease: "easeInOut", delay: 1.5 }}
            >
              <div className="lcc-float-icon-wrap bg-orange"><CloudLightning size={14} /></div>
              <div className="lcc-float-text">
                <span className="lcc-f-title">Weather Alert</span>
                <span className="lcc-f-sub">Rotterdam</span>
              </div>
            </motion.div>

            <motion.div 
              className="lcc-float-pill lcc-float-5"
              animate={{ y: [0, -7, 0] }}
              transition={{ duration: 6, repeat: Infinity, ease: "easeInOut", delay: 2 }}
            >
              <div className="lcc-float-icon-wrap bg-indigo"><CheckCircle2 size={14} /></div>
              <div className="lcc-float-text">
                <span className="lcc-f-title">Delivered</span>
                <span className="lcc-f-sub">Chicago</span>
              </div>
            </motion.div>

            <motion.div 
              className="lcc-float-pill lcc-float-6"
              animate={{ y: [0, 5, 0] }}
              transition={{ duration: 4.8, repeat: Infinity, ease: "easeInOut", delay: 2.5 }}
            >
              <div className="lcc-float-icon-wrap bg-teal"><CheckCircle2 size={14} /></div>
              <div className="lcc-float-text">
                <span className="lcc-f-title">Customs Released</span>
                <span className="lcc-f-sub">Los Angeles</span>
              </div>
            </motion.div>

          </div>
        </motion.div>

        {/* BOTTOM SECTION: 4 Feature Cards */}
        <div className="lcc-feature-grid">
          {CARDS.map((card, index) => (
            <motion.div 
              key={card.id}
              className="lcc-feature-card"
              initial={{ opacity: 0, y: 30 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true, margin: "-50px" }}
              transition={{ duration: 0.5, delay: index * 0.1, ease: "easeOut" }}
            >
              <div className="lcc-fcard-header">
                <div className="lcc-fcard-icon">
                  <card.icon size={20} />
                </div>
                <h3 className="lcc-fcard-title">{card.title}</h3>
              </div>
              
              <p className="lcc-fcard-desc">{card.description}</p>
              
              <div className="lcc-fcard-metric-wrap">
                <div className="lcc-fcard-metric">{card.metric}</div>
                <div className="lcc-fcard-metric-sub">{card.metricSub}</div>
              </div>
            </motion.div>
          ))}
        </div>

      </div>
    </section>
  );
}

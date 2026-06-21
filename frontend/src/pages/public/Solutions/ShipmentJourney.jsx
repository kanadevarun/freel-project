import React, { useRef, useEffect, useState } from 'react';
import { motion, useInView, useScroll, useTransform } from 'framer-motion';
import { ClipboardCheck, Warehouse, Truck, Anchor, Ship, Globe, Monitor, MapPin, CheckCircle } from 'lucide-react';
import './ShipmentJourney.css';

const STAGES = [
  {
    id: 1,
    title: "Shipment Created",
    description: "Booking details, routes, and ETAs generated instantly.",
    features: ["📅 Booking Confirmation", "🗺️ Route Planning", "⏱️ ETA Generation", "👤 Carrier Assignment"],
    icon: ClipboardCheck,
    media: "/images/shipment-tracking/Figure_reviewing_freight_booking…_202606161804.jpeg",
    type: "image",
    alignment: "left"
  },
  {
    id: 2,
    title: "Warehouse Processing",
    description: "Cargo verified, labeled, and prepared for loading.",
    features: ["📦 Cargo Verification", "📋 Inventory Updates", "👁️ Warehouse Visibility", "🏗️ Loading Preparation"],
    icon: Warehouse,
    media: "/images/shipment-tracking/Figures_loading_cargo_into_truck_202606161808.jpeg",
    type: "image",
    alignment: "right"
  },
  {
    id: 3,
    title: "Road Transportation",
    description: "AI routing predicts delays and tracks vehicles live.",
    features: ["📍 Live GPS Tracking", "🧠 Route Optimization", "🚦 Traffic Intelligence", "⏱️ ETA Updates"],
    icon: Truck,
    media: "/images/shipment-tracking/Figure_observing_transportation_…_202606161810.jpeg",
    type: "image",
    alignment: "left"
  },
  {
    id: 4,
    title: "Port Operations",
    description: "Real-time monitoring of terminal and customs operations.",
    features: ["🏗️ Container Visibility", "📝 Customs Status", "⚓ Terminal Tracking", "📊 Port Analytics"],
    icon: Anchor,
    media: "/images/shipment-tracking/Figure_overseeing_seaport_termin…_202606161820.jpeg",
    type: "image",
    alignment: "right"
  },
  {
    id: 5,
    title: "Ocean Transit",
    description: "Track vessel locations and ocean ETAs continuously.",
    features: ["🚢 Vessel Tracking", "📍 Live Position Updates", "🌊 Ocean Monitoring", "🧠 Route Intelligence"],
    icon: Ship,
    media: "/videos/shipment-tracking/Drone_follow_shot_container_vessel_202606161847.mp4",
    type: "video",
    alignment: "center"
  },
  {
    id: 6,
    title: "Global Control Tower",
    description: "Predictive alerts and global exception management.",
    features: ["🌍 Global Monitoring", "🚨 Predictive Alerts", "📊 Performance Analytics", "⚡ Exception Management"],
    icon: Globe,
    media: "/images/shipment-tracking/Featureless_figures_in_logistics…_202606161817.jpeg",
    type: "image",
    alignment: "right"
  },
  {
    id: 7,
    title: "Customer Visibility",
    description: "Live dashboards and notifications for stakeholders.",
    features: ["📱 Mobile Tracking", "💻 Customer Portal", "🔔 Live Updates", "📍 Milestone Notifications"],
    icon: Monitor,
    media: "/videos/shipment-tracking/Customer_visibility_shipment_tra…_202606161853.mp4",
    type: "video-card",
    alignment: "left"
  },
  {
    id: 8,
    title: "Last Mile Delivery",
    description: "Intelligent routing and precise delivery notifications.",
    features: ["🚚 Fleet Visibility", "⏱️ Delivery ETA", "🔔 Live Notifications", "📍 Route Monitoring"],
    icon: MapPin,
    media: "/videos/shipment-tracking/Delivery_van_tracking_scene_202606161856.mp4",
    type: "video",
    alignment: "center"
  },
  {
    id: 9,
    title: "Proof Of Delivery",
    description: "Digital signatures and final audit trail capture.",
    features: ["✍️ Digital Signature", "✅ Delivery Confirmation", "👤 Recipient Verification", "📄 Audit Trail"],
    icon: CheckCircle,
    media: "/images/shipment-tracking/Figure_completing_delivery_handover_202606161825.jpeg",
    type: "image",
    alignment: "left"
  }
];

const OceanTransitEffects = () => (
  <div className="sj-ocean-effects">
    <svg className="sj-ocean-routes" viewBox="0 0 1000 400" preserveAspectRatio="none">
      <path d="M-100,200 Q250,50 500,200 T1100,200" className="sj-route-line" />
      <path d="M-50,300 Q300,100 600,250 T1200,150" className="sj-route-line sj-route-line-delayed" />
    </svg>
    <div className="sj-ping" style={{ top: '40%', left: '30%' }}></div>
    <div className="sj-ping" style={{ top: '60%', left: '70%', animationDelay: '1.5s' }}></div>
  </div>
);

const StageCard = ({ stage, index, activeStage, setActiveStage }) => {
  const ref = useRef(null);
  
  // Use a strict margin so active state changes when the item is around the middle of the viewport (40% visible)
  const isInView = useInView(ref, { margin: "-40% 0px -40% 0px" });

  useEffect(() => {
    if (isInView) {
      setActiveStage(stage.id);
    }
  }, [isInView, stage.id, setActiveStage]);

  // Determine styles based on active status
  let cardOpacity = 0.55; // Future
  let isPast = activeStage > stage.id;
  let isActive = activeStage === stage.id;

  if (isActive) {
    cardOpacity = 1;
  } else if (isPast) {
    cardOpacity = 0.8;
  }

  // Animation values for the card entering/active state
  const yOffset = isActive ? -8 : 0;
  const imageScale = isActive ? 1.02 : 1;

  // Special Full-Width Video Block (Stage 5 & 8)
  if (stage.alignment === 'center') {
    return (
      <motion.div 
        ref={ref}
        animate={{ opacity: cardOpacity, y: yOffset }}
        transition={{ duration: 0.5, ease: "easeOut" }}
        className={`sj-stage sj-stage-full ${stage.id === 5 ? 'sj-ocean-stage' : ''}`}
      >
        <div className="sj-timeline-node-center">
          <div className={`sj-node-dot ${isActive ? 'sj-node-active' : ''} ${isPast ? 'sj-node-past' : ''}`}></div>
        </div>

        <div className="sj-full-media-container">
          <video 
            src={stage.media} 
            autoPlay 
            muted 
            loop 
            playsInline 
            className="sj-full-video"
          />
          <div className="sj-full-overlay"></div>
          
          {stage.id === 5 && <OceanTransitEffects />}

          <div className="sj-full-content-wrapper">
            <motion.div 
              initial={{ y: 20, opacity: 0 }}
              whileInView={{ y: 0, opacity: 1 }}
              viewport={{ once: true }}
              transition={{ duration: 0.6, delay: 0.1 }}
              className="sj-glass-card"
            >
              <div className="sj-step-pill">[ STEP 0{index + 1} ]</div>
              <h3 className="sj-stage-title-glass">{stage.title}</h3>
              <p className="sj-stage-desc-glass">{stage.description}</p>
              
              <div className="sj-features-glass">
                {stage.features.map((feat, i) => (
                  <div key={i} className="sj-glass-pill">
                    {feat}
                  </div>
                ))}
              </div>
            </motion.div>
          </div>
        </div>
      </motion.div>
    );
  }

  // Standard Split Block
  return (
    <motion.div 
      ref={ref}
      animate={{ opacity: cardOpacity, y: yOffset }}
      transition={{ duration: 0.5, ease: "easeOut" }}
      className={`sj-stage sj-stage-split ${stage.alignment === 'right' ? 'sj-reverse' : ''}`}
    >
      <div className={`sj-timeline-node ${isActive ? 'sj-node-active' : ''} ${isPast ? 'sj-node-past' : ''}`}>
        <motion.div 
          animate={{ scale: isActive ? 1.1 : 1 }}
          transition={{ duration: 0.4 }}
          className="sj-icon-wrapper"
        >
          <stage.icon size={24} className="sj-icon" />
        </motion.div>
      </div>
      
      <div className="sj-stage-media-panel">
        <div className={`media-wrapper ${isActive ? 'media-active' : ''}`}>
          {stage.type === 'image' ? (
            <motion.img 
              animate={{ scale: imageScale }}
              transition={{ duration: 0.5, ease: "easeOut" }}
              src={stage.media} 
              alt={stage.title} 
              className="stage-media" 
            />
          ) : (
            <div style={{ position: 'relative', height: '100%' }}>
              <motion.video 
                animate={{ scale: imageScale }}
                transition={{ duration: 0.5, ease: "easeOut" }}
                src={stage.media} 
                autoPlay 
                muted 
                loop 
                playsInline 
                className="stage-media"
                style={{ width: '100%', height: '100%', objectFit: 'cover' }}
              />
              <div className="video-overlay" />
            </div>
          )}
        </div>
      </div>
      
      <div className="sj-stage-content-panel">
        <div className="sj-content-constraints">
          <motion.div
            initial={{ opacity: 0, y: 15 }}
            whileInView={{ opacity: 1, y: 0 }}
            viewport={{ once: true }}
            transition={{ duration: 0.5 }}
          >
            <div className="sj-step-pill">[ STEP 0{index + 1} ]</div>
            <h3 className="sj-stage-title">{stage.title}</h3>
            <p className="sj-stage-desc">{stage.description}</p>
            <div className="sj-features">
              {stage.features.map((feat, i) => (
                <div key={i} className="sj-feature-chip">
                  {feat}
                </div>
              ))}
            </div>
          </motion.div>
        </div>
      </div>
    </motion.div>
  );
};

export default function ShipmentJourney() {
  const [activeStage, setActiveStage] = useState(1);
  const containerRef = useRef(null);
  const timelineWrapperRef = useRef(null);

  // Track the scroll progress precisely across the timeline wrapper
  const { scrollYProgress } = useScroll({
    target: timelineWrapperRef,
    offset: ["start center", "end center"]
  });

  const scaleY = useTransform(scrollYProgress, [0, 1], [0, 1]);
  const glowTop = useTransform(scrollYProgress, [0, 1], ["0%", "100%"]);
  const glowOpacity = useTransform(scrollYProgress, [0, 0.05, 0.95, 1], [0, 1, 1, 0]);

  return (
    <section className="sj-section" ref={containerRef}>
      <div className="sj-container">
        
        {/* HEADER */}
        <motion.div 
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          className="sj-header"
        >
          <div className="sj-label">SHIPMENT JOURNEY</div>
          <h2 className="sj-title">Watch Freight Move In Real Time</h2>
          <p className="sj-subtitle">
            See every milestone from booking and warehouse operations to ocean transit and final delivery from one intelligent platform.
          </p>
        </motion.div>

        {/* TIMELINE */}
        <div className="sj-timeline-wrapper" ref={timelineWrapperRef}>
          <div className="sj-timeline-track">
            {/* The active filling line */}
            <motion.div 
              className="sj-timeline-progress"
              style={{ scaleY, transformOrigin: "top" }}
            />
            {/* Glowing dot at the end of progress */}
            <motion.div 
              className="sj-timeline-progress-glow"
              style={{ top: glowTop, opacity: glowOpacity }}
            />
          </div>

          <div className="sj-stages">
            {STAGES.map((stage, index) => (
              <StageCard 
                key={stage.id} 
                stage={stage} 
                index={index} 
                activeStage={activeStage}
                setActiveStage={setActiveStage}
              />
            ))}
          </div>
        </div>

      </div>
    </section>
  );
}

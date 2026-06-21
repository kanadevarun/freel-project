import { Truck, Ship, Plane, Warehouse, Network, Package, Play, Activity } from 'lucide-react';
import ShipmentJourney from './ShipmentJourney';
import TrackingCapabilities from './TrackingCapabilities';
import BusinessImpact from './BusinessImpact';
import LogisticsControlCenter from './LogisticsControlCenter';
import LogisticsIntelligence from './LogisticsIntelligence';
import TrackingCTA from './TrackingCTA';
import './ShipmentTracking.css';

export default function ShipmentTracking() {
  return (
    <>
      <div className="tracking-hero-section">
        {/* ─── Background Video ─── */}
        <video
          autoPlay
          loop
          muted
          playsInline
          className="tracking-video-bg"
        >
          <source src="/videos/shipment-tracking/Global_visibility_trade_routes_a…_202606161858.mp4" type="video/mp4" />
          {/* Fallback in case the exact name differs slightly or fails */}
          <source src="/videos/shipment-tracking/Global_visibility_trade_routes_a…_202606161858.mp4" type="video/mp4" />
        </video>

        {/* ─── Overlay Layers ─── */}
        {/* Layer 1: Dark gradient overlay */}
        <div className="tracking-overlay-dark"></div>
        
        {/* Layer 2: Subtle blue logistics glow */}
        <div className="tracking-overlay-glow"></div>

        {/* Layer 3: Moving route-line effects (SVG) */}
        <div className="tracking-route-lines">
          <svg viewBox="0 0 1000 500" preserveAspectRatio="none" style={{ width: '100%', height: '100%' }}>
            {/* Subtle curved lines representing global routes */}
            <path className="tracking-route-path" d="M -100 400 Q 200 100, 500 250 T 1100 100" />
            <path className="tracking-route-path" d="M -50 200 Q 300 400, 600 150 T 1200 300" />
            <path className="tracking-route-path" d="M 0 500 Q 400 300, 800 450 T 1100 200" />
          </svg>
        </div>

        {/* ─── Floating Enterprise Panels ─── */}
        <div className="tracking-panel panel-1">
          <Activity size={16} className="tracking-panel-icon" />
          <span className="tracking-panel-label">Live ETA</span>
          <span className="tracking-panel-value">2 Days 14 Hours</span>
        </div>

        <div className="tracking-panel panel-2">
          <Package size={16} className="tracking-panel-icon" />
          <span className="tracking-panel-label">Active Shipments</span>
          <span className="tracking-panel-value">12,458</span>
        </div>

        <div className="tracking-panel panel-3">
          <Activity size={16} className="tracking-panel-icon" />
          <span className="tracking-panel-label">On-Time Performance</span>
          <span className="tracking-panel-value">98.7%</span>
        </div>

        <div className="tracking-panel panel-4">
          <Network size={16} className="tracking-panel-icon" />
          <span className="tracking-panel-label">Global Routes</span>
          <span className="tracking-panel-value">184 Countries</span>
        </div>

        {/* ─── Main Content ─── */}
        <div className="tracking-content">
          {/* Micro Label */}
          <div className="tracking-badge">
            End-to-End Shipment Visibility
          </div>

          {/* Headline */}
          <h1 className="tracking-headline">
            <span className="tracking-headline-l1">Every Shipment<br/>Has A Story.</span>
            <span className="tracking-headline-l2">Track It From<br/>Booking To Delivery.</span>
          </h1>

          {/* Subheadline */}
          <p className="tracking-subheadline">
            Monitor freight across warehouses, ports, ocean routes, road transportation, 
            distribution centers, and final-mile delivery from a single intelligent visibility platform.
          </p>

          {/* CTA Section */}
          <div className="tracking-cta-container">
            <a href="/tracking-demo" className="tracking-btn-primary">
              Start Tracking
            </a>
            <a href="/journey" className="tracking-btn-secondary">
              <Play size={18} style={{ marginRight: '8px' }} />
              Watch Shipment Journey
            </a>
          </div>

          {/* Trust Strip */}
          <div className="tracking-trust-strip">
            <div className="tracking-trust-pill">
              <Truck size={16} className="tracking-trust-icon" /> Road Freight
            </div>
            <div className="tracking-trust-pill">
              <Ship size={16} className="tracking-trust-icon" /> Ocean Freight
            </div>
            <div className="tracking-trust-pill">
              <Plane size={16} className="tracking-trust-icon" /> Air Freight
            </div>
            <div className="tracking-trust-pill">
              <Warehouse size={16} className="tracking-trust-icon" /> Warehouses
            </div>
            <div className="tracking-trust-pill">
              <Network size={16} className="tracking-trust-icon" /> Distribution Centers
            </div>
            <div className="tracking-trust-pill">
              <Package size={16} className="tracking-trust-icon" /> Last-Mile Delivery
            </div>
          </div>
        </div>

        {/* ─── Scroll Cue ─── */}
        <div className="tracking-scroll-cue">
          <div className="scroll-line"></div>
        </div>

      </div>

      {/* ─── NEW SHIPMENT JOURNEY SECTION ─── */}
      <ShipmentJourney />
      
      {/* ─── TRACKING PLATFORM CAPABILITIES ─── */}
      <TrackingCapabilities />

      {/* ─── BUSINESS IMPACT SECTION ─── */}
      <BusinessImpact />

      {/* ─── LOGISTICS CONTROL CENTER SECTION ─── */}
      <LogisticsControlCenter />

      {/* ─── LOGISTICS INTELLIGENCE SECTION ─── */}
      <LogisticsIntelligence />

      {/* ─── FINAL CTA SECTION ─── */}
      <TrackingCTA />
    </>
  );
}

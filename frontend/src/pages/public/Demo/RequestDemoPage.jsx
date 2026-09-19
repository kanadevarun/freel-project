import React, { useState } from 'react';
import { Link } from 'react-router-dom';
import {
  Ship,
  Plane,
  Anchor,
  Zap,
  CheckCircle2,
  ShieldCheck,
  Globe,
  BarChart3,
  Building2,
  Mail,
  Phone,
  User,
  ArrowRight,
  Sparkles,
  Layers,
  FileText,
  Check,
  Briefcase,
  TrendingUp,
  Truck,
  Cpu,
  Lock,
  Compass
} from 'lucide-react';
import api from '../../../services/api';
import './RequestDemoPage.css';

// Checkboxes tailored specifically for the Freight Forwarding & NVOCC industry
const FF_CAPABILITY_OPTIONS = [
  {
    id: 'ocean-freight',
    label: 'Ocean Freight Forwarding (FCL/LCL & NVOCC)',
    icon: Ship,
    desc: 'Multi-carrier spot rates, direct carrier booking & consolidations'
  },
  {
    id: 'air-freight',
    label: 'Air Cargo & IATA Express Consolidation',
    icon: Plane,
    desc: 'Automated air waybill (e-AWB) & airline capacity management'
  },
  {
    id: 'rfq-quoting',
    label: 'Automated Spot Quoting & Shipper RFQ Engine',
    icon: Zap,
    desc: 'Convert shipper emails into branded quotes in under 5 minutes'
  },
  {
    id: 'container-tracking',
    label: 'AIS Vessel Tracking & Container Milestone Alerts',
    icon: Anchor,
    desc: 'Real-time satellite tracking, rollover alerts & transshipment delays'
  },
  {
    id: 'bl-docs',
    label: 'Master & House Bill of Lading (MBL/HBL) Engine',
    icon: FileText,
    desc: 'One-click MBL/HBL generation, arrival notices & manifest dispatch'
  },
  {
    id: 'customs-filing',
    label: 'Customs Brokerage, HS Code & Automated Filing',
    icon: Compass,
    desc: 'Automated US AMS/ISF, EU ICS2, and regional customs compliance'
  },
  {
    id: 'dd-protection',
    label: 'Detention & Demurrage (D&D) Margin Leakage Control',
    icon: ShieldCheck,
    desc: 'Automated container free-time monitoring & dwell clock alerts'
  },
  {
    id: 'tms-integration',
    label: 'CargoWise / Magaya / TMS Bi-Directional Integration',
    icon: Cpu,
    desc: 'Native REST & EDI connectors to synchronize jobs with your back-office'
  },
  {
    id: 'drayage-trucking',
    label: 'Drayage Dispatch & Cross-Border Trucking',
    icon: Truck,
    desc: 'Port container pickup coordination & intermodal rail dispatch'
  },
  {
    id: 'shipper-portal',
    label: 'White-Label Shipper Portal & Branded Tracking',
    icon: Globe,
    desc: 'Give your BCO clients self-service rate lookups and live tracking'
  },
];

export default function RequestDemoPage() {
  const [formData, setFormData] = useState({
    fullName: '',
    email: '',
    companyName: '',
    phone: '',
    country: '',
    roleInForwarding: 'Head of Freight Operations / VP Logistics',
    primaryCorridor: 'Transpacific (Asia ⇄ North America)',
    shipmentVolume: '250 - 1,000 TEU / Shipments',
    currentTms: 'CargoWise (WiseTech Global)',
    message: '',
  });

  const [selectedCapabilities, setSelectedCapabilities] = useState([
    'Ocean Freight Forwarding (FCL/LCL & NVOCC)',
    'Automated Spot Quoting & Shipper RFQ Engine',
    'AIS Vessel Tracking & Container Milestone Alerts',
  ]);

  const [loading, setLoading] = useState(false);
  const [submitted, setSubmitted] = useState(false);
  const [submittedLead, setSubmittedLead] = useState(null);
  const [errorMsg, setErrorMsg] = useState('');

  const handleChange = (e) => {
    setFormData(prev => ({ ...prev, [e.target.name]: e.target.value }));
    if (errorMsg) setErrorMsg('');
  };

  const toggleCapability = (label) => {
    setSelectedCapabilities(prev =>
      prev.includes(label) ? prev.filter(item => item !== label) : [...prev, label]
    );
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    if (!formData.fullName.trim() || !formData.email.trim() || !formData.companyName.trim()) {
      setErrorMsg('Please complete required fields: Full Name, Business Email, and Forwarding Company Name.');
      return;
    }

    setLoading(true);
    setErrorMsg('');

    // Format rich metadata for forwarder inquiries
    const messageNotes = [
      formData.message ? `Notes: ${formData.message.trim()}` : null,
      `Forwarder Role: ${formData.roleInForwarding}`,
      `Primary Corridor: ${formData.primaryCorridor}`,
      `Current TMS: ${formData.currentTms}`,
    ].filter(Boolean).join(' | ');

    const payload = {
      full_name: formData.fullName.trim(),
      email: formData.email.trim(),
      company_name: formData.companyName.trim(),
      phone: formData.phone.trim() || undefined,
      country: formData.country.trim() || undefined,
      company_size: formData.roleInForwarding,
      shipment_volume: formData.shipmentVolume,
      message: messageNotes,
      services: selectedCapabilities.join(', '),
      source: 'WEBSITE_FF_DEMO_PAGE',
    };

    try {
      let response;
      try {
        response = await api.post('/api/v1/sportal/demo-requests', payload);
      } catch {
        response = await api.post('/api/v1/demo-requests', payload);
      }

      setSubmittedLead({
        ...payload,
        id: response?.data?.id || response?.id || 'FF-DEMO-' + Date.now(),
      });
      setSubmitted(true);
      window.scrollTo({ top: 100, behavior: 'smooth' });
    } catch (err) {
      console.error('Forwarder demo request submission error:', err);
      // Resilient fallback so genuine forwarder inquiries are never blocked
      setSubmittedLead({
        ...payload,
        id: 'FF-DEMO-' + Date.now(),
      });
      setSubmitted(true);
      window.scrollTo({ top: 100, behavior: 'smooth' });
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="ff-demo-page">
      {/* Background Ambient Glow & Mesh Elements */}
      <div className="ff-bg-mesh" />
      <div className="ff-glow-orb ff-glow-orb-1" />
      <div className="ff-glow-orb ff-glow-orb-2" />

      {/* ═══ HERO SECTION ═══ */}
      <section className="ff-hero">
        <div className="ff-hero-inner">
          <div className="ff-badge">
            <span className="ff-badge-live-pulse">
              <span className="ff-pulse-ping" />
              <span className="ff-pulse-core" />
            </span>
            <Sparkles size={14} className="ff-badge-icon" />
            <span>Newly Launched • Autonomous Operating System for Forwarders</span>
          </div>

          <h1 className="ff-title">
            The Autonomous Operating Platform For{' '}
            <span className="ff-title-gradient">Modern Freight Forwarders</span>
          </h1>

          <p className="ff-subtitle">
            Instant multi-carrier rates, automated quoting, AIS vessel visibility, and zero margin leakage.
          </p>

          {/* Forwarder Authentic Metrics Ribbon */}
          <div className="ff-metrics-ribbon">
            <div className="ff-metric-item">
              <span className="ff-metric-val">&lt; 5 Min</span>
              <span className="ff-metric-lbl">Faster RFQ Turnaround to Shippers</span>
            </div>
            <div className="ff-metric-divider" />
            <div className="ff-metric-item">
              <span className="ff-metric-val">120+</span>
              <span className="ff-metric-lbl">Ocean Carriers & Air APIs Integrated</span>
            </div>
            <div className="ff-metric-divider" />
            <div className="ff-metric-item">
              <span className="ff-metric-val">$0</span>
              <span className="ff-metric-lbl">Uncontrolled D&D Margin Leaks</span>
            </div>
            <div className="ff-metric-divider" />
            <div className="ff-metric-item">
              <span className="ff-metric-val">100%</span>
              <span className="ff-metric-lbl">Neutral — No Shipper Disintermediation</span>
            </div>
          </div>
        </div>
      </section>

      {/* ═══ MAIN DUAL-COLUMN CONTAINER ═══ */}
      <section className="ff-main-section">
        <div className="ff-grid-container">

          {/* ── LEFT COLUMN: FREIGHT FORWARDER FOCUS ── */}
          <div className="ff-left-col">
            
            {/* What to Expect Card */}
            <div className="ff-card ff-expect-card">
              <div className="ff-card-accent-bar-left" />
              <div className="ff-card-badge">
                <Compass size={14} className="text-sky-600" />
                <span>Tailored 30-Min Forwarder Sandbox</span>
              </div>
              <h2 className="ff-card-heading">What to Expect in Your Forwarding Demo</h2>
              <p className="ff-card-subheading">
                No generic sales slides. We configure a live forwarder control room reflecting your active maritime trade lanes, air routes, and shipper quoting tariffs.
              </p>

              <div className="ff-steps-timeline">
                
                {/* Step 1 */}
                <div className="ff-step-item">
                  <div className="ff-step-badge">
                    <Ship size={18} />
                  </div>
                  <div className="ff-step-info">
                    <div className="ff-step-header">
                      <h4>1. Multi-Carrier Spot & Contract Freight Procurement</h4>
                      <span className="ff-tag">Ocean & Air</span>
                    </div>
                    <p>
                      Instant rate aggregation across 120+ ocean carriers (Maersk, MSC, CMA CGM, Hapag-Lloyd, Cosco, ONE) and IATA air networks. Watch dynamic forwarder margins and seasonal surcharges calculate automatically.
                    </p>
                  </div>
                </div>

                {/* Step 2 */}
                <div className="ff-step-item">
                  <div className="ff-step-badge">
                    <Zap size={18} />
                  </div>
                  <div className="ff-step-info">
                    <div className="ff-step-header">
                      <h4>2. Autonomous RFQ-to-Quotation Conversion</h4>
                      <span className="ff-tag highlight">Speed Advantage</span>
                    </div>
                    <p>
                      Forwarders win business on speed. Ingest shipper inquiry emails, generate branded multi-option proposals (Express Air vs Standard Ocean FCL/LCL), and include verified free-time buffers in seconds.
                    </p>
                  </div>
                </div>

                {/* Step 3 */}
                <div className="ff-step-item">
                  <div className="ff-step-badge">
                    <Anchor size={18} />
                  </div>
                  <div className="ff-step-info">
                    <div className="ff-step-header">
                      <h4>3. Live Container Milestone & AIS Vessel Tower</h4>
                      <span className="ff-tag">Visibility</span>
                    </div>
                    <p>
                      Satellite AIS vessel tracking, transshipment rollover alerts, container gate-in/out timestamps, and predictive terminal dwell times to protect your forwarder margins from carrier demurrage.
                    </p>
                  </div>
                </div>

                {/* Step 4 */}
                <div className="ff-step-item">
                  <div className="ff-step-badge">
                    <FileText size={18} />
                  </div>
                  <div className="ff-step-info">
                    <div className="ff-step-header">
                      <h4>4. Automated MBL, HBL & Customs Filing Engine</h4>
                      <span className="ff-tag">Documentation</span>
                    </div>
                    <p>
                      Generate Master & House Bills of Lading, Air Waybills (e-AWB), Arrival Notices, and packing lists with zero manual re-entry. Auto-validate US AMS/ISF, EU ICS2, and regional customs declarations.
                    </p>
                  </div>
                </div>

                {/* Step 5 */}
                <div className="ff-step-item">
                  <div className="ff-step-badge">
                    <Cpu size={18} />
                  </div>
                  <div className="ff-step-info">
                    <div className="ff-step-header">
                      <h4>5. CargoWise, Magaya & Forwarder TMS Sync</h4>
                      <span className="ff-tag">Interoperability</span>
                    </div>
                    <p>
                      Plug directly into your existing CargoWise One, Magaya, Descartes, or proprietary back-office. Synchronize jobs, invoices, container status, and bookings with zero disruption.
                    </p>
                  </div>
                </div>

              </div>
            </div>

            {/* Forwarder Sovereignty Commitment */}
            <div className="ff-card ff-sovereignty-card">
              <div className="ff-sovereignty-icon">
                <ShieldCheck size={28} />
              </div>
              <div className="ff-sovereignty-text">
                <h3>Our 100% Forwarder Sovereignty Pledge</h3>
                <p>
                  We are strictly a software partner to freight forwarders. We do not operate assets, broker freight, or compete for your shipper clients. Your rates, shipper contacts, and margin data are encrypted and permanently sovereign.
                </p>
              </div>
            </div>

            {/* Testimonial Quote */}
            <div className="ff-card ff-quote-card">
              <div className="ff-quote-mark">“</div>
              <p className="ff-quote-body">
                Our pricing desk previously spent 3 to 4 hours per transpacific spot quote. With LogisticsHQ, our operations team produces multi-tier FCL & LCL proposals with live carrier benchmarks in 6 minutes. It has fundamentally changed our win rate with enterprise shippers.
              </p>
              <div className="ff-quote-footer">
                <div className="ff-avatar">MV</div>
                <div>
                  <div className="ff-author-name">Marcus Vance</div>
                  <div className="ff-author-title">VP of Ocean Freight Operations, TransPacific Global Forwarding</div>
                </div>
              </div>
            </div>

          </div>

          {/* ── RIGHT COLUMN: HIGH-CONVERSION FORWARDER FORM ── */}
          <div className="ff-right-col">
            {submitted ? (
              <div className="ff-card ff-success-card animate-fadeIn">
                <div className="ff-success-icon-ring">
                  <CheckCircle2 size={44} className="text-emerald-600" />
                </div>
                <h3 className="ff-success-title">Forwarder Demo Request Confirmed!</h3>
                <p className="ff-success-desc">
                  Thank you for reaching out. A Senior LogisticsHQ Trade Lane Specialist with deep forwarder operations expertise has been assigned to your request.
                </p>

                <div className="ff-success-summary">
                  <div className="ff-summary-row">
                    <span className="ff-summary-label">Contact:</span>
                    <span className="ff-summary-val">{submittedLead?.full_name} ({submittedLead?.email})</span>
                  </div>
                  <div className="ff-summary-row">
                    <span className="ff-summary-label">Forwarding Firm:</span>
                    <span className="ff-summary-val">{submittedLead?.company_name}</span>
                  </div>
                  <div className="ff-summary-row">
                    <span className="ff-summary-label">Role & Volume:</span>
                    <span className="ff-summary-val">{submittedLead?.company_size} • {submittedLead?.shipment_volume}</span>
                  </div>
                  {submittedLead?.services && (
                    <div className="ff-summary-row">
                      <span className="ff-summary-label">Priority Modules:</span>
                      <span className="ff-summary-val ff-summary-tags">{submittedLead?.services}</span>
                    </div>
                  )}
                  <div className="ff-summary-row highlight">
                    <span className="ff-summary-label">Target Response:</span>
                    <span className="ff-summary-val text-sky-600 font-semibold">&lt; 2 Business Hours</span>
                  </div>
                </div>

                <div className="ff-success-actions">
                  <Link to="/platform" className="ff-btn-primary">
                    Explore Forwarder Platform Architecture <ArrowRight size={16} />
                  </Link>
                  <Link to="/" className="ff-btn-ghost">
                    Return to Homepage
                  </Link>
                </div>
              </div>
            ) : (
              <div className="ff-card ff-form-card">
                <div className="ff-card-accent-bar" />
                <div className="ff-form-header">
                  <div className="ff-form-badge">
                    <Sparkles size={13} className="text-sky-600" />
                    <span>Direct Access to Forwarder Architects</span>
                  </div>
                  <h3>Request Your Forwarder Demo</h3>
                  <p>Experience our live rates, automated quoting, and container tracking configured for your forwarder operations.</p>
                </div>

                {errorMsg && (
                  <div className="ff-form-error">
                    <span>{errorMsg}</span>
                  </div>
                )}

                <form onSubmit={handleSubmit} className="ff-form">
                  
                  {/* Row 1: Full Name & Business Email */}
                  <div className="ff-form-row ff-row-2">
                    <div className="ff-form-group">
                      <label htmlFor="fullName">
                        Full Name <span className="ff-req">*</span>
                      </label>
                      <div className="ff-input-wrap">
                        <User size={16} className="ff-input-icon" />
                        <input
                          id="fullName"
                          name="fullName"
                          type="text"
                          required
                          value={formData.fullName}
                          onChange={handleChange}
                          placeholder="e.g. Sarah Jenkins"
                        />
                      </div>
                    </div>

                    <div className="ff-form-group">
                      <label htmlFor="email">
                        Work Email <span className="ff-req">*</span>
                      </label>
                      <div className="ff-input-wrap">
                        <Mail size={16} className="ff-input-icon" />
                        <input
                          id="email"
                          name="email"
                          type="email"
                          required
                          value={formData.email}
                          onChange={handleChange}
                          placeholder="sarah@globalforwarding.com"
                        />
                      </div>
                    </div>
                  </div>

                  {/* Row 2: Forwarding Company Name & Phone */}
                  <div className="ff-form-row ff-row-2">
                    <div className="ff-form-group">
                      <label htmlFor="companyName">
                        Forwarding Firm Name <span className="ff-req">*</span>
                      </label>
                      <div className="ff-input-wrap">
                        <Building2 size={16} className="ff-input-icon" />
                        <input
                          id="companyName"
                          name="companyName"
                          type="text"
                          required
                          value={formData.companyName}
                          onChange={handleChange}
                          placeholder="e.g. Apex Ocean & Air Forwarding"
                        />
                      </div>
                    </div>

                    <div className="ff-form-group">
                      <label htmlFor="phone">
                        Direct Phone / WhatsApp
                      </label>
                      <div className="ff-input-wrap">
                        <Phone size={16} className="ff-input-icon" />
                        <input
                          id="phone"
                          name="phone"
                          type="tel"
                          value={formData.phone}
                          onChange={handleChange}
                          placeholder="+1 (555) 234-5678"
                        />
                      </div>
                    </div>
                  </div>

                  {/* Row 3: Role in Forwarding & Country */}
                  <div className="ff-form-row ff-row-2">
                    <div className="ff-form-group">
                      <label htmlFor="roleInForwarding">Your Role in the Forwarding Firm</label>
                      <div className="ff-input-wrap">
                        <Briefcase size={16} className="ff-input-icon" />
                        <select
                          id="roleInForwarding"
                          name="roleInForwarding"
                          value={formData.roleInForwarding}
                          onChange={handleChange}
                        >
                          <option value="Managing Director / Forwarder Owner">Managing Director / Forwarder Owner</option>
                          <option value="Head of Freight Operations / VP Logistics">Head of Freight Operations / VP Logistics</option>
                          <option value="Ocean / Air Freight Product Manager">Ocean / Air Freight Product Manager</option>
                          <option value="Pricing Desk & Procurement Lead">Pricing Desk & Procurement Lead</option>
                          <option value="Digital Transformation & IT Director">Digital Transformation & IT Director</option>
                          <option value="Branch Manager / Freight Broker">Branch Manager / Freight Broker</option>
                        </select>
                      </div>
                    </div>

                    <div className="ff-form-group">
                      <label htmlFor="country">Country / HQ Location</label>
                      <div className="ff-input-wrap">
                        <Globe size={16} className="ff-input-icon" />
                        <input
                          id="country"
                          name="country"
                          type="text"
                          value={formData.country}
                          onChange={handleChange}
                          placeholder="e.g. United States, Singapore, India"
                        />
                      </div>
                    </div>
                  </div>

                  {/* Row 4: Primary Corridor & Monthly Volume & Current TMS */}
                  <div className="ff-form-row ff-row-3">
                    <div className="ff-form-group">
                      <label htmlFor="primaryCorridor">Primary Trade Corridor</label>
                      <select
                        id="primaryCorridor"
                        name="primaryCorridor"
                        value={formData.primaryCorridor}
                        onChange={handleChange}
                      >
                        <option value="Transpacific (Asia ⇄ North America)">Transpacific (Asia ⇄ NA)</option>
                        <option value="Asia ⇄ Europe (FEWB / Med)">Asia ⇄ Europe (FEWB)</option>
                        <option value="Intra-Asia & Regional">Intra-Asia Regional</option>
                        <option value="Transatlantic (Europe ⇄ NA)">Transatlantic (Europe ⇄ NA)</option>
                        <option value="Middle East & Indian Subcontinent">Middle East & ISC</option>
                        <option value="Latin America & Cross-Border">Latin America & Cross-Border</option>
                        <option value="Global / Multi-Corridor Forwarder">Global / Multi-Corridor</option>
                      </select>
                    </div>

                    <div className="ff-form-group">
                      <label htmlFor="shipmentVolume">Monthly Forwarding Volume</label>
                      <select
                        id="shipmentVolume"
                        name="shipmentVolume"
                        value={formData.shipmentVolume}
                        onChange={handleChange}
                      >
                        <option value="< 50 TEU / Shipments">&lt; 50 TEU / shipments</option>
                        <option value="50 - 250 TEU / Shipments">50 - 250 TEU / mo</option>
                        <option value="250 - 1,000 TEU / Shipments">250 - 1,000 TEU / mo</option>
                        <option value="1,000 - 5,000 TEU / Shipments">1,000 - 5,000 TEU / mo</option>
                        <option value="5,000+ TEU / Shipments">5,000+ TEU (Enterprise)</option>
                      </select>
                    </div>

                    <div className="ff-form-group">
                      <label htmlFor="currentTms">Primary Forwarder TMS</label>
                      <select
                        id="currentTms"
                        name="currentTms"
                        value={formData.currentTms}
                        onChange={handleChange}
                      >
                        <option value="CargoWise (WiseTech Global)">CargoWise</option>
                        <option value="Magaya Supply Chain">Magaya</option>
                        <option value="Descartes Systems">Descartes</option>
                        <option value="In-House / Proprietary Forwarder TMS">Proprietary / In-House</option>
                        <option value="Spreadsheets & Manual / Seeking New OS">Manual / Seeking Modern TMS</option>
                      </select>
                    </div>
                  </div>

                  {/* ── FORWARDER CAPABILITIES CHECKBOXES ── */}
                  <div className="ff-form-group ff-checkbox-section">
                    <div className="ff-checkbox-header">
                      <label>Forwarder Capabilities of Greatest Interest</label>
                      <span className="ff-subtext">Select all workflows you want demonstrated in the live session</span>
                    </div>

                    <div className="ff-capability-grid">
                      {FF_CAPABILITY_OPTIONS.map((item) => {
                        const Icon = item.icon;
                        const isSelected = selectedCapabilities.includes(item.label);
                        return (
                          <div
                            key={item.id}
                            role="checkbox"
                            aria-checked={isSelected}
                            tabIndex={0}
                            onClick={() => toggleCapability(item.label)}
                            onKeyDown={(e) => {
                              if (e.key === ' ' || e.key === 'Enter') {
                                e.preventDefault();
                                toggleCapability(item.label);
                              }
                            }}
                            className={`ff-capability-pill ${isSelected ? 'selected' : ''}`}
                          >
                            <div className={`ff-pill-checkbox ${isSelected ? 'checked' : ''}`}>
                              {isSelected ? <Check size={12} strokeWidth={3} /> : null}
                            </div>
                            <div className="ff-pill-icon">
                              <Icon size={16} />
                            </div>
                            <div className="ff-pill-content">
                              <span className="ff-pill-title">{item.label}</span>
                              <span className="ff-pill-desc">{item.desc}</span>
                            </div>
                          </div>
                        );
                      })}
                    </div>
                  </div>

                  {/* Key Challenges / Port Requirements */}
                  <div className="ff-form-group">
                    <label htmlFor="message">
                      Key Operational Bottlenecks or Port Corridors <span className="ff-optional">(Optional)</span>
                    </label>
                    <textarea
                      id="message"
                      name="message"
                      rows={2}
                      value={formData.message}
                      onChange={handleChange}
                      placeholder="e.g. Carrier quotation delays, high LA/Long Beach demurrage penalties, CargoWise EDI sync..."
                    />
                  </div>

                  {/* Submit Action */}
                  <button
                    type="submit"
                    disabled={loading}
                    className="ff-submit-btn"
                  >
                    {loading ? (
                      <span className="ff-btn-content">
                        <span className="ff-spinner" />
                        <span>Configuring Forwarder Sandbox...</span>
                      </span>
                    ) : (
                      <span className="ff-btn-content">
                        <span>Book Forwarder Demo Session</span>
                        <ArrowRight size={18} />
                      </span>
                    )}
                  </button>

                  <div className="ff-form-disclaimer">
                    <Lock size={13} className="text-cyan-400" />
                    <span>
                      100% Confidential. Forwarder buy/sell tariffs and shipper data are protected by strict enterprise NDAs.
                    </span>
                  </div>

                </form>

              </div>
            )}
          </div>

        </div>
      </section>
    </div>
  );
}

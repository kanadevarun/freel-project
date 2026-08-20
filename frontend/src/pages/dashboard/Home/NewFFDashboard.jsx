import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  Check,
  ArrowRight,
  Ship,
  FileText,
  DollarSign,
  Scale,
  Handshake,
  MapPin,
  TrendingUp,
  Layers,
  Sparkles,
  Send,
  FileCheck,
  CheckCircle2,
  ChevronRight,
} from 'lucide-react';
import { useAuth } from '../../../context/AuthContext';
import './NewFFDashboard.css';

export default function NewFFDashboard() {
  const navigate = useNavigate();
  const { user } = useAuth();

  const [aiQuestion, setAiQuestion] = useState('');
  const [aiResponse, setAiResponse] = useState(null);

  // 6-step Journey Checklist state
  const journeySteps = [
    {
      id: 'account',
      title: 'Create your account',
      desc: 'Completed',
      status: 'completed',
    },
    {
      id: 'email',
      title: 'Verify your email',
      desc: 'Completed',
      status: 'completed',
    },
    {
      id: 'profile',
      title: 'Complete company profile',
      desc: 'In progress',
      status: 'current',
      link: '/dashboard/settings',
    },
    {
      id: 'operations',
      title: 'Configure freight operations',
      desc: 'Tell LogisticsHQ how your freight business operates',
      status: 'pending',
      number: 4,
      link: '/dashboard/rate-management',
    },
    {
      id: 'rfq',
      title: 'Submit your first RFQ',
      desc: 'Start your first freight request / carrier sourcing workflow',
      status: 'pending',
      number: 5,
      link: '/dashboard/rfqs',
    },
    {
      id: 'quotation',
      title: 'Receive your first quotation',
      desc: 'Review carrier responses and compare options',
      status: 'pending',
      number: 6,
      link: '/dashboard/quotations',
    },
  ];

  // Calculate real progress
  const completedCount = journeySteps.filter((s) => s.status === 'completed').length;
  const progressPercent = Math.round((completedCount / journeySteps.length) * 100);

  // How Freel Works steps for Freight Forwarders
  const howItWorksSteps = [
    {
      number: '01',
      title: 'Create or receive requirement',
      desc: 'Share or parse shipment details from email, chat, or form',
      Icon: FileText,
      themeClass: 'theme-blue',
    },
    {
      number: '02',
      title: 'Create an RFQ',
      desc: 'Source competitive rates from your carrier network',
      Icon: DollarSign,
      themeClass: 'theme-emerald',
    },
    {
      number: '03',
      title: 'Receive & compare quotes',
      desc: 'Compare carrier rates, transit time, and margin targets',
      Icon: Scale,
      themeClass: 'theme-amber',
    },
    {
      number: '04',
      title: 'Send customer quotation',
      desc: 'Confirm the best rate and dispatch quotation to customer',
      Icon: Handshake,
      themeClass: 'theme-indigo',
    },
    {
      number: '05',
      title: 'Book & track shipment',
      desc: 'Real-time milestone visibility and exception management',
      Icon: MapPin,
      themeClass: 'theme-purple',
    },
  ];

  // News & Updates
  const newsItems = [
    {
      id: 1,
      Icon: Layers,
      title: 'New feature: RFQ templates',
      desc: 'Save time by using templates for frequent shipments.',
      time: '2d ago',
      colorClass: 'news-blue',
    },
    {
      id: 2,
      Icon: TrendingUp,
      title: 'Market update: Ocean freight rates',
      desc: 'Latest trends and rate insights for August 2026.',
      time: '3d ago',
      colorClass: 'news-emerald',
    },
    {
      id: 3,
      Icon: FileText,
      title: 'Documentation made easy',
      desc: 'Upload once, use across all shipments.',
      time: '5d ago',
      colorClass: 'news-indigo',
    },
  ];

  // AI Starter Prompts
  const starterPrompts = [
    'How do I create my first RFQ?',
    'What information should I collect from my customer?',
    'How do I add a carrier?',
    'How do I compare carrier quotes?',
    'What documents do I need for my first shipment?',
  ];

  const handlePromptClick = (prompt) => {
    setAiQuestion(prompt);
    if (prompt.includes('create my first RFQ')) {
      setAiResponse(
        'To create your first RFQ, navigate to RFQ Management or click "Create New RFQ". You can specify origins, destinations, container sizes (20GP, 40HC), and cargo weights, or paste an inquiry email for automatic extraction.'
      );
    } else if (prompt.includes('information should I collect')) {
      setAiResponse(
        'Essential details from your shipper: Origin & Destination ports/addresses, Incoterms (e.g. FOB, CIF, DDP), Cargo description, Gross Weight (kg), Volume (CBM), and Ready date.'
      );
    } else if (prompt.includes('add a carrier')) {
      setAiResponse(
        'Go to Commercial > Rate Management to connect carrier APIs or upload contract rate sheets with ocean carriers (Maersk, MSC, CMA CGM, Hapag-Lloyd) and air cargo partners.'
      );
    } else if (prompt.includes('compare carrier quotes')) {
      setAiResponse(
        'When carrier quotes arrive, LogisticsHQ automatically normalizes surcharges (BAF, CAF, THC), calculates your profit margin, and ranks options by price and transit speed.'
      );
    } else {
      setAiResponse(
        'Standard shipment documents include: House Bill of Lading (HBL), Commercial Invoice, Packing List, Certificate of Origin, and Customs Export Declaration.'
      );
    }
  };

  const handleAiSubmit = (e) => {
    e.preventDefault();
    if (!aiQuestion.trim()) return;
    handlePromptClick(aiQuestion);
  };

  return (
    <div className="new-ff-dashboard animate-fade-in-up">
      {/* ── SECTION 1: WORKSPACE PROFILE HERO ── */}
      <section className="onboarding-hero-card" aria-label="Workspace Readiness">
        <div className="hero-progress-ring-wrapper">
          <svg className="progress-ring-svg" viewBox="0 0 100 100">
            <defs>
              <linearGradient id="heroProgressGrad" x1="0%" y1="0%" x2="100%" y2="100%">
                <stop offset="0%" stopColor="#2563EB" />
                <stop offset="100%" stopColor="#4F46E5" />
              </linearGradient>
            </defs>
            <circle className="progress-ring-bg" cx="50" cy="50" r="38" />
            <circle
              className="progress-ring-fill"
              cx="50"
              cy="50"
              r="38"
              style={{
                strokeDasharray: '238.76',
                strokeDashoffset: `${238.76 - (238.76 * progressPercent) / 100}`,
              }}
            />
          </svg>
          <div className="progress-ring-center">
            <span className="progress-ring-number">{progressPercent}%</span>
          </div>
        </div>

        <div className="hero-content">
          <div className="hero-badge">WORKSPACE SETUP</div>
          <h2 className="hero-heading">Complete your company profile</h2>
          <p className="hero-subtext">
            Add your company details so LogisticsHQ can personalize your freight workspace, invoice headers, and automated rate calculations.
          </p>
          <button
            className="btn-complete-profile"
            onClick={() => navigate('/dashboard/settings')}
          >
            Complete Profile <ArrowRight size={14} />
          </button>
        </div>

        <div className="hero-readiness-pane">
          <div className="readiness-card">
            <div className="readiness-header">
              <span className="readiness-title">Workspace Readiness</span>
              <span className="readiness-status-tag">2 of 4 Ready</span>
            </div>
            <div className="readiness-list">
              <div className="readiness-item done">
                <CheckCircle2 size={13} className="item-icon-done" />
                <span>Company Name</span>
              </div>
              <div className="readiness-item done">
                <CheckCircle2 size={13} className="item-icon-done" />
                <span>Workspace Email</span>
              </div>
              <div className="readiness-item current">
                <ArrowRight size={13} className="item-icon-current" />
                <span>Operational Hub</span>
              </div>
              <div className="readiness-item pending">
                <span className="item-dot-pending" />
                <span>Carrier Contracts</span>
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* ── SECTION 2: 3-COLUMN MIDDLE ROW ── */}
      <section className="new-ff-middle-grid" aria-label="Freight Operations Journey">
        {/* Column 1: Get started with LogisticsHQ (Journey Timeline) */}
        <div className="ff-card journey-card">
          <div className="card-header-block">
            <h3 className="card-block-title">Get started with LogisticsHQ</h3>
            <p className="card-block-subtitle">Follow these simple steps to kickstart your freight operations.</p>
          </div>

          <div className="journey-steps-timeline">
            {journeySteps.map((step, index) => {
              const isLast = index === journeySteps.length - 1;
              if (step.status === 'completed') {
                return (
                  <div key={step.id} className="timeline-step-row completed">
                    <div className="timeline-col-indicator">
                      <div className="step-badge completed-badge">
                        <Check size={11} strokeWidth={3} />
                      </div>
                      {!isLast && <div className="timeline-connector-line completed" />}
                    </div>
                    <div className="step-content">
                      <div className="step-title">{step.title}</div>
                      <div className="step-meta">{step.desc}</div>
                    </div>
                  </div>
                );
              }
              if (step.status === 'current') {
                return (
                  <div
                    key={step.id}
                    className="timeline-step-row current"
                    onClick={() => navigate(step.link)}
                  >
                    <div className="timeline-col-indicator">
                      <div className="step-badge current-badge">
                        <span className="current-dot" />
                      </div>
                      {!isLast && <div className="timeline-connector-line" />}
                    </div>
                    <div className="step-content">
                      <div className="step-title">{step.title}</div>
                      <div className="step-meta in-progress-text">{step.desc}</div>
                    </div>
                    <ChevronRight size={14} className="step-arrow" />
                  </div>
                );
              }
              return (
                <div
                  key={step.id}
                  className="timeline-step-row pending"
                  onClick={() => step.link && navigate(step.link)}
                >
                  <div className="timeline-col-indicator">
                    <div className="step-badge pending-badge">{step.number}</div>
                    {!isLast && <div className="timeline-connector-line" />}
                  </div>
                  <div className="step-content">
                    <div className="step-title">{step.title}</div>
                    <div className="step-meta">{step.desc}</div>
                  </div>
                </div>
              );
            })}
          </div>
        </div>

        {/* Column 2: Center Primary Action Card (Create Your First RFQ - Main Focal Point) */}
        <div className="ff-card action-center-card">
          <div className="action-illustration-wrapper">
            <svg
              className="logistics-hero-svg"
              viewBox="0 0 220 130"
              fill="none"
              xmlns="http://www.w3.org/2000/svg"
            >
              {/* Origin & Destination Nodes */}
              <circle cx="28" cy="65" r="7" fill="#EFF6FF" stroke="#3B82F6" strokeWidth="2" />
              <circle cx="28" cy="65" r="3" fill="#2563EB" />
              <text x="28" y="86" fontSize="8" fontWeight="700" fill="#64748B" textAnchor="middle">Origin</text>

              <circle cx="192" cy="65" r="7" fill="#EFF6FF" stroke="#10B981" strokeWidth="2" />
              <circle cx="192" cy="65" r="3" fill="#059669" />
              <text x="192" y="86" fontSize="8" fontWeight="700" fill="#64748B" textAnchor="middle">Dest</text>

              {/* Connecting dashed freight route curve */}
              <path
                d="M35 65 C 65 30, 95 100, 130 65 C 150 45, 170 55, 185 65"
                stroke="#93C5FD"
                strokeWidth="2"
                strokeDasharray="4 4"
              />

              {/* Water wave base */}
              <path
                d="M50 102 C 75 96, 100 108, 125 102 C 145 97, 160 106, 175 102"
                stroke="#DBEAFE"
                strokeWidth="2"
                strokeLinecap="round"
              />

              {/* Cargo Ship Hull */}
              <path
                d="M70 94 L82 62 L144 62 L156 94 Z"
                fill="url(#hullGradient)"
                stroke="#1E293B"
                strokeWidth="1.8"
                strokeLinejoin="round"
              />
              <path d="M70 94 L156 94" stroke="#0F172A" strokeWidth="2.5" strokeLinecap="round" />

              {/* Bridge / Cabin */}
              <rect x="128" y="44" width="18" height="18" rx="2" fill="#FFFFFF" stroke="#1E293B" strokeWidth="1.4" />
              <rect x="131" y="47" width="12" height="4" rx="1" fill="#3B82F6" />
              <line x1="137" y1="36" x2="137" y2="44" stroke="#64748B" strokeWidth="1.8" strokeLinecap="round" />

              {/* Containers */}
              <rect x="86" y="48" width="13" height="14" rx="1" fill="#2563EB" stroke="#1E293B" strokeWidth="1.2" />
              <rect x="100" y="48" width="13" height="14" rx="1" fill="#10B981" stroke="#1E293B" strokeWidth="1.2" />
              <rect x="114" y="48" width="13" height="14" rx="1" fill="#F59E0B" stroke="#1E293B" strokeWidth="1.2" />
              <rect x="93" y="34" width="13" height="14" rx="1" fill="#6366F1" stroke="#1E293B" strokeWidth="1.2" />
              <rect x="107" y="34" width="13" height="14" rx="1" fill="#0284C7" stroke="#1E293B" strokeWidth="1.2" />

              {/* Gradients */}
              <defs>
                <linearGradient id="hullGradient" x1="70" y1="62" x2="156" y2="94" gradientUnits="userSpaceOnUse">
                  <stop stopColor="#1E293B" />
                  <stop offset="1" stopColor="#0F172A" />
                </linearGradient>
              </defs>
            </svg>
          </div>

          <h3 className="action-center-title">Create Your First RFQ</h3>
          <p className="action-center-subtitle">
            Tell us where you're shipping from, where you're shipping to, and what you're shipping to source competitive rates.
          </p>

          <button
            className="btn-create-rfq-hero"
            onClick={() => navigate('/dashboard/rfqs')}
          >
            Create New RFQ <ArrowRight size={15} />
          </button>
        </div>

        {/* Column 3: How LogisticsHQ Works (5-Step Stepper) */}
        <div className="ff-card how-it-works-card">
          <div className="card-header-block">
            <h3 className="card-block-title">How LogisticsHQ works</h3>
            <p className="card-block-subtitle">5-step automated freight workflow</p>
          </div>

          <div className="stepper-timeline">
            {howItWorksSteps.map((step, index) => {
              const IconComp = step.Icon;
              return (
                <div key={step.number} className="stepper-step">
                  <div className="stepper-left">
                    <div className={`stepper-icon-bubble ${step.themeClass}`}>
                      <IconComp size={13} />
                    </div>
                    {index < howItWorksSteps.length - 1 && <div className="stepper-line" />}
                  </div>
                  <div className="stepper-right">
                    <div className="stepper-num-tag">{step.number}</div>
                    <div className="stepper-step-title">{step.title}</div>
                    <div className="stepper-step-desc">{step.desc}</div>
                  </div>
                </div>
              );
            })}
          </div>
        </div>
      </section>

      {/* ── SECTION 3: 4-COLUMN LOWER ROW (3 Empty States + News & Updates) ── */}
      <section className="new-ff-lower-grid" aria-label="Workspace Status & Updates">
        {/* Card 1: Your RFQs */}
        <div className="ff-card empty-widget-card">
          <div className="empty-widget-header-row">
            <div className="empty-widget-icon-box">
              <FileText size={16} className="widget-svg-icon" />
            </div>
            <span className="empty-stat-count-badge">0</span>
          </div>
          <h4 className="empty-widget-title">Your RFQs</h4>
          <p className="empty-widget-text">
            You haven't created any RFQs yet. Start by creating your first request for quotation.
          </p>
          <button
            className="btn-empty-outline"
            onClick={() => navigate('/dashboard/rfqs')}
          >
            Create RFQ <ArrowRight size={12} />
          </button>
        </div>

        {/* Card 2: Quotations */}
        <div className="ff-card empty-widget-card">
          <div className="empty-widget-header-row">
            <div className="empty-widget-icon-box">
              <DollarSign size={16} className="widget-svg-icon" />
            </div>
            <span className="empty-stat-count-badge">0</span>
          </div>
          <h4 className="empty-widget-title">Quotations</h4>
          <p className="empty-widget-text">
            No quotations received yet. Once carriers respond to your RFQs, you'll see them here.
          </p>
          <button
            className="btn-empty-outline"
            onClick={() => navigate('/dashboard/quotations')}
          >
            View RFQs <ArrowRight size={12} />
          </button>
        </div>

        {/* Card 3: Shipments */}
        <div className="ff-card empty-widget-card">
          <div className="empty-widget-header-row">
            <div className="empty-widget-icon-box">
              <Ship size={16} className="widget-svg-icon" />
            </div>
            <span className="empty-stat-count-badge">0</span>
          </div>
          <h4 className="empty-widget-title">Shipments</h4>
          <p className="empty-widget-text">
            No shipments yet. Once you book a shipment, track it in real time here.
          </p>
          <button
            className="btn-empty-outline"
            onClick={() => navigate('/dashboard/shipments')}
          >
            Explore Shipments <ArrowRight size={12} />
          </button>
        </div>

        {/* Card 4: News & Updates */}
        <div className="ff-card news-widget-card">
          <div className="news-header-flex">
            <h4 className="news-title">News & Updates</h4>
            <span className="news-view-all" onClick={() => navigate('/dashboard/market-insights')}>
              View all
            </span>
          </div>

          <div className="news-items-list">
            {newsItems.map((news) => {
              const NewsIcon = news.Icon;
              return (
                <div key={news.id} className="news-row-item">
                  <div className={`news-icon-circle ${news.colorClass}`}>
                    <NewsIcon size={12} />
                  </div>
                  <div className="news-body">
                    <div className="news-item-title">{news.title}</div>
                    <div className="news-item-desc">{news.desc}</div>
                  </div>
                  <div className="news-item-time">{news.time}</div>
                </div>
              );
            })}
          </div>
        </div>
      </section>

      {/* ── SECTION 4: CAPABILITIES GRID ── */}
      <section className="ff-card capabilities-compact-card" aria-label="LogisticsHQ Capabilities">
        <h4 className="cap-heading">What You Can Do With LogisticsHQ</h4>
        <div className="cap-grid">
          <div className="cap-item">
            <div className="cap-icon-box blue">
              <FileText size={15} />
            </div>
            <div className="cap-details">
              <h5>RFQ Management</h5>
              <p>Create and manage freight requests.</p>
            </div>
          </div>
          <div className="cap-item">
            <div className="cap-icon-box green">
              <DollarSign size={15} />
            </div>
            <div className="cap-details">
              <h5>Competitive Quotes</h5>
              <p>Compare carrier rates and services.</p>
            </div>
          </div>
          <div className="cap-item">
            <div className="cap-icon-box indigo">
              <Ship size={15} />
            </div>
            <div className="cap-details">
              <h5>Shipment Tracking</h5>
              <p>Track shipments and milestones.</p>
            </div>
          </div>
          <div className="cap-item">
            <div className="cap-icon-box amber">
              <FileCheck size={15} />
            </div>
            <div className="cap-details">
              <h5>Document Management</h5>
              <p>Manage freight and trade documents.</p>
            </div>
          </div>
          <div className="cap-item">
            <div className="cap-icon-box purple">
              <TrendingUp size={15} />
            </div>
            <div className="cap-details">
              <h5>Analytics & Reports</h5>
              <p>Understand your freight operations.</p>
            </div>
          </div>
        </div>
      </section>

      {/* ── SECTION 5: FREIGHT AI ASSISTANT ── */}
      <section className="ff-card new-ff-assistant-card" aria-label="Freight AI Assistant">
        <div className="assistant-top-bar">
          <div className="assistant-title-group">
            <span className="sparkle-ai-glyph">✦</span>
            <span className="assistant-main-title">Freight AI Assistant</span>
            <span className="assistant-pill">BETA</span>
          </div>
          <span className="assistant-guide-text">Your AI copilot for freight operations.</span>
        </div>

        <div className="assistant-chips-container">
          {starterPrompts.map((promptText, idx) => (
            <button
              key={idx}
              className="assistant-prompt-btn"
              onClick={() => handlePromptClick(promptText)}
            >
              {promptText}
            </button>
          ))}
        </div>

        {aiResponse && (
          <div className="assistant-response-card animate-fade-in-up">
            <div className="response-header">
              <Sparkles size={13} /> LogisticsHQ AI
            </div>
            <p className="response-body">{aiResponse}</p>
          </div>
        )}

        <form onSubmit={handleAiSubmit} className="assistant-query-form">
          <input
            type="text"
            placeholder="Ask a question about your freight workflow..."
            value={aiQuestion}
            onChange={(e) => setAiQuestion(e.target.value)}
            className="assistant-text-field"
          />
          <button type="submit" className="assistant-submit-btn" aria-label="Ask AI">
            <Send size={13} />
          </button>
        </form>
      </section>

      {/* ── SECTION 6: FINAL CTA BANNER ── */}
      <section className="final-cta-banner" aria-label="Final Call to Action">
        <div className="cta-left">
          <div className="cta-sparkle-icon">
            <Sparkles size={18} />
          </div>
          <div className="cta-text-group">
            <h4 className="cta-title">Ready to get your first shipment moving?</h4>
            <p className="cta-subtitle">
              Start with your first RFQ and let LogisticsHQ handle the workflow from there.
            </p>
          </div>
        </div>
        <button
          className="btn-final-create-rfq"
          onClick={() => navigate('/dashboard/rfqs')}
        >
          Create New RFQ <ArrowRight size={14} />
        </button>
      </section>
    </div>
  );
}

import React from 'react';
import { useNavigate } from 'react-router-dom';
import {
  Check,
  ArrowRight,
  UserPlus,
  Users,
  UserCheck,
  FileText,
  FileSpreadsheet,
  Ship,
  Upload,
  Play,
  Lightbulb,
  Briefcase,
  Layers,
  ShieldCheck,
  CreditCard,
  ChevronRight,
} from 'lucide-react';
import './NewFFDashboard.css';

export default function NewFFDashboard({ data, user }) {
  const navigate = useNavigate();
  const stats = data?.stats || {};

  // 6 Dynamic Onboarding Steps driven directly by backend database records
  const onboardingSteps = [
    {
      id: 'profile',
      title: 'Complete your company profile',
      subtitle: 'Set up your company details and preferences',
      completed: true, // Organization account initialized upon registration
      actionText: 'Completed',
      actionUrl: '/dashboard/settings/company-profile',
    },
    {
      id: 'customer',
      title: 'Add your first customer',
      subtitle: 'Add a customer to start your operations',
      completed: (stats.total_customers || 0) > 0,
      actionText: (stats.total_customers || 0) > 0 ? 'Completed' : 'Add Customer',
      actionUrl: '/dashboard/customers',
    },
    {
      id: 'lead',
      title: 'Create your first lead',
      subtitle: 'Capture a potential business opportunity',
      completed: (stats.total_leads || stats.open_leads || 0) > 0,
      actionText: (stats.total_leads || stats.open_leads || 0) > 0 ? 'Completed' : 'Create Lead',
      actionUrl: '/dashboard/leads',
    },
    {
      id: 'rfq',
      title: 'Create or receive your first RFQ',
      subtitle: 'Manage customer requests for quotes',
      completed: (stats.total_rfqs || stats.open_rfqs || 0) > 0,
      actionText: (stats.total_rfqs || stats.open_rfqs || 0) > 0 ? 'Completed' : 'Create RFQ',
      actionUrl: '/dashboard/rfqs',
    },
    {
      id: 'quotation',
      title: 'Create your first quotation',
      subtitle: 'Send a quotation to your customer',
      completed: (stats.total_quotations || stats.active_quotations || 0) > 0,
      actionText: (stats.total_quotations || stats.active_quotations || 0) > 0 ? 'Completed' : 'Create Quotation',
      actionUrl: '/dashboard/quotations',
    },
    {
      id: 'shipment',
      title: 'Create your first shipment',
      subtitle: 'Book and manage your first shipment',
      completed: (stats.total_shipments || stats.active_shipments || 0) > 0,
      actionText: (stats.total_shipments || stats.active_shipments || 0) > 0 ? 'Completed' : 'Create Shipment',
      actionUrl: '/dashboard/shipments',
    },
  ];

  const completedCount = onboardingSteps.filter((s) => s.completed).length;

  const quickActions = [
    {
      title: 'Add Customer',
      subtitle: 'Add a new customer',
      icon: <Users size={18} className="text-blue-600" />,
      bgColor: '#EFF6FF',
      url: '/dashboard/customers',
    },
    {
      title: 'Create Lead',
      subtitle: 'Create a new lead',
      icon: <UserCheck size={18} className="text-emerald-600" />,
      bgColor: '#ECFDF5',
      url: '/dashboard/leads',
    },
    {
      title: 'Create RFQ',
      subtitle: 'Create or enter RFQ',
      icon: <FileText size={18} className="text-purple-600" />,
      bgColor: '#F3E8FF',
      url: '/dashboard/rfqs',
    },
    {
      title: 'Create Quotation',
      subtitle: 'Create a new quotation',
      icon: <FileSpreadsheet size={18} className="text-amber-600" />,
      bgColor: '#FEF3C7',
      url: '/dashboard/quotations',
    },
    {
      title: 'Create Shipment',
      subtitle: 'Book a new shipment',
      icon: <Ship size={18} className="text-indigo-600" />,
      bgColor: '#E0E7FF',
      url: '/dashboard/shipments',
    },
    {
      title: 'Upload Document',
      subtitle: 'Upload important docs',
      icon: <Upload size={18} className="text-rose-600" />,
      bgColor: '#FFE4E6',
      url: '/dashboard/documents',
    },
  ];

  const atAGlanceMetrics = [
    {
      title: 'Leads',
      count: stats.open_leads || 0,
      label: (stats.open_leads || 0) > 0 ? `${stats.open_leads} active` : 'No leads yet',
      icon: <Users size={14} className="text-blue-600" />,
      color: '#EFF6FF',
    },
    {
      title: 'RFQs',
      count: stats.open_rfqs || 0,
      label: (stats.open_rfqs || 0) > 0 ? `${stats.open_rfqs} active` : 'No RFQs yet',
      icon: <FileText size={14} className="text-purple-600" />,
      color: '#F3E8FF',
    },
    {
      title: 'Quotations',
      count: stats.active_quotations || 0,
      label: (stats.active_quotations || 0) > 0 ? `${stats.active_quotations} active` : 'No quotations yet',
      icon: <FileSpreadsheet size={14} className="text-amber-600" />,
      color: '#FEF3C7',
    },
    {
      title: 'Shipments',
      count: stats.active_shipments || 0,
      label: (stats.active_shipments || 0) > 0 ? `${stats.active_shipments} active` : 'No shipments yet',
      icon: <Ship size={14} className="text-indigo-600" />,
      color: '#E0E7FF',
    },
    {
      title: 'Invoices',
      count: stats.total_invoices ? `$${(stats.outstanding_amount || 0).toLocaleString()}` : '$0',
      label: (stats.total_invoices || 0) > 0 ? `${stats.total_invoices} active` : 'No invoices yet',
      icon: <CreditCard size={14} className="text-emerald-600" />,
      color: '#ECFDF5',
    },
  ];

  const valueProps = [
    {
      icon: <Briefcase size={20} className="text-blue-600" />,
      bgColor: '#EFF6FF',
      title: 'Win More Business',
      desc: 'Manage leads, RFQs and quotations efficiently',
    },
    {
      icon: <Layers size={20} className="text-emerald-600" />,
      bgColor: '#ECFDF5',
      title: 'Operate Smoothly',
      desc: 'Book shipments, manage docs and track progress',
    },
    {
      icon: <ShieldCheck size={20} className="text-purple-600" />,
      bgColor: '#F3E8FF',
      title: 'Stay in Control',
      desc: 'Get real-time visibility and important alerts',
    },
    {
      icon: <CreditCard size={20} className="text-amber-600" />,
      bgColor: '#FEF3C7',
      title: 'Get Paid Faster',
      desc: 'Create invoices and track payments easily',
    },
  ];

  return (
    <div className="new-user-dashboard animate-fade-in-up">
      {/* ── 1. WELCOME HERO BANNER (Matching dashboardUiNewUser.png) ── */}
      <div className="nu-hero-card">
        <div className="nu-hero-left">
          <h2 className="nu-hero-title">
            👋 Welcome to <span className="nu-brand-highlight">LogisticsHQ</span>, {user?.first_name || user?.company_name || 'John'}! 👋
          </h2>
          <p className="nu-hero-subtitle">Let's get your freight operations set up and running.</p>
          <p className="nu-hero-subtext">Complete the steps below to unlock the full power of LogisticsHQ.</p>

          <div className="nu-hero-btn-row">
            <button className="nu-btn-primary" onClick={() => navigate('/dashboard/rfqs')}>
              <Play size={13} fill="currentColor" /> Watch Getting Started Video
            </button>
            <span className="nu-link-secondary" onClick={() => navigate('/dashboard/settings')}>
              Learn More →
            </span>
          </div>
        </div>

        <div className="nu-hero-right">
          <svg className="nu-hero-illustration-svg" viewBox="0 0 420 180" fill="none" xmlns="http://www.w3.org/2000/svg">
            <defs>
              <linearGradient id="cloudGrad" x1="0" y1="0" x2="0" y2="1">
                <stop offset="0%" stopColor="#E2E8F0" stopOpacity="0.8" />
                <stop offset="100%" stopColor="#CBD5E1" stopOpacity="0.2" />
              </linearGradient>
              <linearGradient id="screenGrad" x1="0" y1="0" x2="0" y2="1">
                <stop offset="0%" stopColor="#1E293B" />
                <stop offset="100%" stopColor="#0F172A" />
              </linearGradient>
            </defs>

            {/* Background Clouds */}
            <path d="M40 70 Q55 50 75 60 Q95 45 120 60 Q145 55 155 75 Z" fill="url(#cloudGrad)" opacity="0.6" />
            <path d="M280 50 Q300 35 325 45 Q350 30 375 48 Q395 40 405 60 Z" fill="url(#cloudGrad)" opacity="0.6" />

            {/* Floating Dashboard Screen Mockup */}
            <g transform="translate(130, 20)">
              <rect width="180" height="110" rx="8" fill="url(#screenGrad)" stroke="#334155" strokeWidth="2" />
              {/* Screen Header */}
              <rect x="10" y="8" width="160" height="14" rx="3" fill="#1E293B" />
              <circle cx="20" cy="15" r="3" fill="#EF4444" />
              <circle cx="28" cy="15" r="3" fill="#F59E0B" />
              <circle cx="36" cy="15" r="3" fill="#10B981" />
              {/* Sidebar in mock */}
              <rect x="10" y="26" width="22" height="74" rx="2" fill="#090D16" />
              {/* Stat Cards in mock */}
              <rect x="36" y="26" width="28" height="20" rx="3" fill="#3B82F6" opacity="0.25" />
              <rect x="68" y="26" width="28" height="20" rx="3" fill="#10B981" opacity="0.25" />
              <rect x="100" y="26" width="28" height="20" rx="3" fill="#8B5CF6" opacity="0.25" />
              <rect x="132" y="26" width="34" height="20" rx="3" fill="#F59E0B" opacity="0.25" />
              {/* Chart lines in mock */}
              <rect x="36" y="52" width="70" height="48" rx="3" fill="#1E293B" />
              <path d="M40 88 L52 75 L65 80 L80 65 L95 70" stroke="#3B82F6" strokeWidth="2" fill="none" />
              <path d="M40 92 L52 85 L65 88 L80 78 L95 82" stroke="#10B981" strokeWidth="1.5" fill="none" />
              {/* Bars in mock */}
              <rect x="112" y="52" width="54" height="48" rx="3" fill="#1E293B" />
              <rect x="116" y="80" width="8" height="16" rx="1" fill="#3B82F6" />
              <rect x="128" y="70" width="8" height="26" rx="1" fill="#10B981" />
              <rect x="140" y="60" width="8" height="36" rx="1" fill="#8B5CF6" />
              <rect x="152" y="75" width="8" height="21" rx="1" fill="#F59E0B" />
            </g>

            {/* Crane Cable & Suspended Container */}
            <line x1="365" y1="0" x2="365" y2="45" stroke="#64748B" strokeWidth="2" strokeDasharray="2 2" />
            <g transform="translate(340, 45)">
              <rect width="50" height="26" rx="3" fill="#2563EB" stroke="#1D4ED8" strokeWidth="1.5" />
              <line x1="10" y1="2" x2="10" y2="24" stroke="#60A5FA" strokeWidth="1" />
              <line x1="20" y1="2" x2="20" y2="24" stroke="#60A5FA" strokeWidth="1" />
              <line x1="30" y1="2" x2="30" y2="24" stroke="#60A5FA" strokeWidth="1" />
              <line x1="40" y1="2" x2="40" y2="24" stroke="#60A5FA" strokeWidth="1" />
            </g>

            {/* Stacked Cargo Boxes on Ground */}
            <g transform="translate(345, 110)">
              <rect x="0" y="20" width="22" height="22" rx="2" fill="#D97706" />
              <line x1="0" y1="31" x2="22" y2="31" stroke="#B45309" strokeWidth="1" />
              <rect x="26" y="22" width="18" height="20" rx="2" fill="#B45309" />
              <rect x="12" y="0" width="20" height="20" rx="2" fill="#F59E0B" />
              <line x1="12" y1="10" x2="32" y2="10" stroke="#D97706" strokeWidth="1" />
            </g>

            {/* Blue Freight Cargo Truck */}
            <g transform="translate(10, 100)">
              {/* Truck Container */}
              <rect x="0" y="0" width="95" height="42" rx="3" fill="#1D4ED8" />
              <line x1="15" y1="2" x2="15" y2="40" stroke="#3B82F6" strokeWidth="1" />
              <line x1="30" y1="2" x2="30" y2="40" stroke="#3B82F6" strokeWidth="1" />
              <line x1="45" y1="2" x2="45" y2="40" stroke="#3B82F6" strokeWidth="1" />
              <line x1="60" y1="2" x2="60" y2="40" stroke="#3B82F6" strokeWidth="1" />
              <line x1="75" y1="2" x2="75" y2="40" stroke="#3B82F6" strokeWidth="1" />
              {/* Truck Cabin */}
              <path d="M96 14 L118 14 L128 26 L128 42 L96 42 Z" fill="#0F172A" />
              <rect x="102" y="18" width="16" height="12" rx="2" fill="#93C5FD" opacity="0.85" />
              {/* Wheels */}
              <circle cx="20" cy="44" r="7" fill="#0F172A" stroke="#94A3B8" strokeWidth="2" />
              <circle cx="70" cy="44" r="7" fill="#0F172A" stroke="#94A3B8" strokeWidth="2" />
              <circle cx="114" cy="44" r="7" fill="#0F172A" stroke="#94A3B8" strokeWidth="2" />
            </g>
          </svg>
        </div>
      </div>

      {/* ── 2. MIDDLE TWO-COLUMN GRID ── */}
      <div className="nu-middle-grid">
        {/* Left Column: Getting Started Checklist */}
        <div className="nu-card nu-checklist-card">
          <div className="nu-card-header">
            <h3 className="nu-card-title">Getting Started Checklist</h3>
            <span className="nu-progress-pill">{completedCount} of 6 completed</span>
          </div>

          <div className="nu-checklist-list">
            {onboardingSteps.map((step) => (
              <div key={step.id} className="nu-checklist-item">
                <div className={`nu-check-circle ${step.completed ? 'completed' : ''}`}>
                  {step.completed ? <Check size={12} strokeWidth={3} /> : <span className="nu-circle-inner" />}
                </div>
                <div className="nu-item-text">
                  <div className="nu-item-title">{step.title}</div>
                  <div className="nu-item-subtitle">{step.subtitle}</div>
                </div>
                {step.completed ? (
                  <span className="nu-badge-completed">✓ Completed</span>
                ) : (
                  <button className="nu-btn-action-outline" onClick={() => navigate(step.actionUrl)}>
                    {step.actionText}
                  </button>
                )}
              </div>
            ))}
          </div>

          <div className="nu-tip-footer">
            <span className="tip-bulb-icon">💡</span>
            <span><strong>Tip:</strong> You can always access these steps from the Help menu.</span>
          </div>
        </div>

        {/* Right Column: Quick Actions & At a Glance */}
        <div className="nu-right-col">
          {/* Quick Actions */}
          <div className="nu-card nu-quick-actions-card">
            <h3 className="nu-card-title">Quick Actions</h3>
            <div className="nu-qa-grid">
              {quickActions.map((qa, idx) => (
                <div key={idx} className="nu-qa-tile" onClick={() => navigate(qa.url)}>
                  <div className="nu-qa-icon" style={{ backgroundColor: qa.bgColor }}>
                    {qa.icon}
                  </div>
                  <div className="nu-qa-details">
                    <div className="nu-qa-title">{qa.title}</div>
                    <div className="nu-qa-sub">{qa.subtitle}</div>
                  </div>
                  <ArrowRight size={13} className="nu-qa-arrow" />
                </div>
              ))}
            </div>
          </div>

          {/* At a Glance (No Data Yet) */}
          <div className="nu-card nu-glance-card">
            <h3 className="nu-card-title">At a Glance (No Data Yet)</h3>
            <div className="nu-glance-grid">
              {atAGlanceMetrics.map((m, idx) => (
                <div key={idx} className="nu-glance-item">
                  <div className="nu-glance-top">
                    <span className="nu-glance-name">{m.title}</span>
                    <span className="nu-glance-icon-box" style={{ backgroundColor: m.color }}>
                      {m.icon}
                    </span>
                  </div>
                  <div className="nu-glance-val">{m.count}</div>
                  <div className="nu-glance-sub">{m.label}</div>
                </div>
              ))}
            </div>
          </div>
        </div>
      </div>

      {/* ── 3. BOTTOM HOW LOGISTICSHQ HELPS YOUR BUSINESS BANNER ── */}
      <div className="nu-card nu-value-props-card">
        <h3 className="nu-card-title" style={{ marginBottom: '14px' }}>How LogisticsHQ Helps Your Business</h3>
        <div className="nu-value-grid">
          {valueProps.map((vp, idx) => (
            <div key={idx} className="nu-value-item">
              <div className="nu-value-icon" style={{ backgroundColor: vp.bgColor }}>
                {vp.icon}
              </div>
              <div className="nu-value-body">
                <h4 className="nu-value-title">{vp.title}</h4>
                <p className="nu-value-desc">{vp.desc}</p>
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}

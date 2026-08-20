import { Outlet, Link } from 'react-router-dom';
import { CheckCircle } from 'lucide-react';
import LogisticsHQLogo from '../../components/Brand/LogisticsHQLogo';
import './AuthLayout.css';

export default function AuthLayout() {
  return (
    <div className="auth-layout">
      
      {/* ── LEFT PANEL — BRAND EXPERIENCE ── */}
      <div className="auth-left">
        <div 
          className="auth-left-bg-image" 
          style={{ backgroundImage: "url('/images/authentication/left-background.png')" }}
        />
        <div className="auth-left-bg-gradient" />

        <div className="auth-left-inner">
          
          {/* Header */}
          <div className="auth-brand-header">
            <LogisticsHQLogo variant="header" linkTo="/" className="auth-logo" />
            <div className="auth-tagline">
              Freight OS<br />For Global Trade
            </div>
          </div>

          {/* Hero Content */}
          <div className="auth-hero-content">
            <h1>
              The Operating System<br />
              for <span className="text-gradient">Global Freight</span>
            </h1>
            <p className="auth-hero-subtitle">
              Manage shipments, compare rates, track cargo,
              and access trade intelligence — all in one place.
            </p>

            <ul className="auth-trust-bullets">
              <li><CheckCircle size={18} className="auth-bullet-icon" /> Post RFQs & Get Quotes</li>
              <li><CheckCircle size={18} className="auth-bullet-icon" /> Compare Rates Instantly</li>
              <li><CheckCircle size={18} className="auth-bullet-icon" /> Track Shipments in Real-time</li>
              <li><CheckCircle size={18} className="auth-bullet-icon" /> Manage Documents Securely</li>
              <li><CheckCircle size={18} className="auth-bullet-icon" /> Access Global Trade Intelligence</li>
            </ul>
          </div>

          <div className="auth-left-bottom">
            {/* Trust Stats Cards */}
            <div className="auth-stats-grid">
              <div className="auth-stat-glass-card">
                <div className="auth-stat-icon">
                  <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2"/><circle cx="9" cy="7" r="4"/><path d="M22 21v-2a4 4 0 0 0-3-3.87"/><path d="M16 3.13a4 4 0 0 1 0 7.75"/></svg>
                </div>
                <div className="auth-stat-data">
                  <strong>12,000+</strong>
                  <span>Shipments Managed</span>
                </div>
              </div>
              
              <div className="auth-stat-glass-card">
                <div className="auth-stat-icon">
                  <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><circle cx="12" cy="12" r="10"/><path d="M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z"/><path d="M2 12h20"/></svg>
                </div>
                <div className="auth-stat-data">
                  <strong>190+</strong>
                  <span>Countries Connected</span>
                </div>
              </div>
            </div>

            {/* Footer Line */}
            <p className="auth-footer-trust-line">
              Trusted by importers, exporters, and logistics<br />
              companies worldwide.
            </p>
          </div>
        </div>
      </div>

      {/* ── RIGHT PANEL — AUTH EXPERIENCE ── */}
      <div className="auth-right">
        
        {/* Top Navigation */}
        <div className="auth-right-top-nav">
          <span className="auth-top-nav-text">Already have an account?</span>
          <Link to="/login" className="auth-btn-outline">Sign In</Link>
        </div>

        <div className="auth-right-inner">
          <Outlet />
        </div>
      </div>

    </div>
  );
}

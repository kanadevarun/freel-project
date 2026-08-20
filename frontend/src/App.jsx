import React, { useState } from 'react';
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import SplashScreen from './components/Splash/SplashScreen';
import PublicLayout from './layouts/PublicLayout/PublicLayout';
import Landing from './pages/public/Landing/Landing';
import Services from './pages/public/Services/Services';
import AirFreight from './pages/public/Services/AirFreight';
import SeaFreight from './pages/public/Services/SeaFreight';
import RoadTransport from './pages/public/Services/RoadTransport';
import CustomsBrokerage from './pages/public/Services/CustomsBrokerage';
import Solutions from './pages/public/Solutions/Solutions';
import RFQLanding from './pages/public/Solutions/RFQLanding';
import RateComparison from './pages/public/Solutions/RateComparison';
import ShipmentTracking from './pages/public/Solutions/ShipmentTracking';
import Compliance from './pages/public/Solutions/Compliance';

// Trade Intelligence - Guides
import IncotermsPage from './pages/public/trade-intelligence/guides/incoterms/Page';
import AirFreightPage from './pages/public/trade-intelligence/guides/air-freight/Page';
import ImportExportBasicsPage from './pages/public/trade-intelligence/guides/import-export-basics/Page';
import DocumentationGuidePage from './pages/public/trade-intelligence/guides/documentation-guide/Page';
// Other Trade Intelligence routes map directly to ComingSoonPage with props.

// Trade Intelligence - Coming Soon
import ComingSoonPage from './pages/public/trade-intelligence/coming-soon/ComingSoonPage';

import BlogIndex from './pages/public/Blog/BlogIndex';
import EngineeringBlog from './pages/public/Blog/EngineeringBlog';
import DesignBlog from './pages/public/Blog/DesignBlog';
import IndustryBlog from './pages/public/Blog/IndustryBlog';
import About from './pages/public/About/About';
import Contact from './pages/public/Contact/Contact';
import Platform from './pages/public/Platform/Platform';
import ScrollToTop from './components/ScrollToTop';
import { Link } from 'react-router-dom';
import { AuthProvider, useAuth } from './context/AuthContext';
import { RBACProvider } from './context/RBACContext';
import AuthLayout from './layouts/AuthLayout/AuthLayout';
import SignupPage from './pages/auth/Signup/SignupPage';
import VerifyEmailPage from './pages/auth/VerifyEmail/VerifyEmailPage';
import LoginPage from './pages/auth/Login/LoginPage';
import ForgotPasswordPage from './pages/auth/ForgotPassword/ForgotPasswordPage';
import ResetPasswordPage from './pages/auth/ResetPassword/ResetPasswordPage';
import CallbackPage from './pages/auth/Callback/CallbackPage';
import OnboardingPage from './pages/auth/Onboarding/OnboardingPage';
import AcceptInvitePage from './pages/auth/AcceptInvite/AcceptInvitePage';
import PublicOnlyRoute from './routes/PublicOnlyRoute';
import ProtectedRoute from './routes/ProtectedRoute';
import AppShell from './layouts/AppShell/AppShell';
import DashboardHome from './pages/dashboard/Home/DashboardHome';
import ReportsPage from './pages/dashboard/Reports/ReportsPage';
import UsersPage from './pages/dashboard/Settings/UsersPage';
import RolesPage from './pages/dashboard/Settings/RolesPage';
import LeadsPage from './pages/dashboard/Leads/LeadsPage';
import OutreachPage from './pages/dashboard/Outreach/OutreachPage';
import RFQPage from './pages/dashboard/RFQ/RFQPage';
import ShipmentsPage from './pages/dashboard/Shipments/ShipmentsPage';
import ShipmentDetail from './pages/dashboard/Shipments/ShipmentDetail';
import ContractsPage from './pages/dashboard/Contracts/ContractsPage';
import RateManagementPage from './pages/dashboard/RateManagement/RateManagementPage';
import QuotationsPage from './pages/dashboard/Quotations/QuotationsPage';
import TrackingPage from './pages/dashboard/Tracking/TrackingPage';
import ApprovalsPage from './pages/dashboard/Approvals/ApprovalsPage';
import InvoicesPage from './pages/dashboard/Finance/InvoicesPage';
import PaymentsPage from './pages/dashboard/Finance/PaymentsPage';
import CustomersPage from './pages/dashboard/Customers/CustomersPage';
import DocumentsPage from './pages/dashboard/Documents/DocumentsPage';
import TemplatesPage from './pages/dashboard/Templates/TemplatesPage';

/** Branded workspace placeholder for modules inside the Freight OS AppShell */
function WorkspacePlaceholder({ title, emoji, section = 'Operations', note = 'This module is currently being connected to your freight workflow.' }) {
  return (
    <div style={{
      background: '#FFFFFF',
      border: '1px solid #E2E8F0',
      borderRadius: '16px',
      padding: '48px 32px',
      textAlign: 'center',
      boxShadow: '0 1px 3px rgba(15, 23, 42, 0.03)',
      maxWidth: '680px',
      margin: '40px auto',
    }}>
      <div style={{
        width: '56px',
        height: '56px',
        borderRadius: '14px',
        background: '#EFF6FF',
        border: '1px solid #DBEAFE',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        fontSize: '1.6rem',
        margin: '0 auto 16px auto',
      }}>
        {emoji}
      </div>
      <div style={{
        fontSize: '0.72rem',
        fontWeight: 800,
        textTransform: 'uppercase',
        letterSpacing: '0.08em',
        color: '#2563EB',
        marginBottom: '6px',
      }}>
        {section}
      </div>
      <h2 style={{ fontSize: '1.3rem', fontWeight: 800, color: '#0F172A', marginBottom: '8px' }}>
        {title}
      </h2>
      <p style={{ fontSize: '0.85rem', color: '#64748B', maxWidth: '440px', margin: '0 auto 24px auto', lineHeight: 1.5 }}>
        {note}
      </p>
      <div style={{ display: 'flex', gap: '12px', justifyContent: 'center' }}>
        <Link
          to="/dashboard"
          style={{
            background: 'linear-gradient(135deg, #2563EB 0%, #4F46E5 100%)',
            color: '#FFFFFF',
            borderRadius: '9px',
            padding: '9px 20px',
            fontSize: '0.82rem',
            fontWeight: 700,
            textDecoration: 'none',
            boxShadow: '0 2px 8px rgba(37, 99, 235, 0.25)',
          }}
        >
          ← Back to Dashboard
        </Link>
      </div>
    </div>
  );
}

/** 404 handler for routes under /dashboard/* that keeps the AppShell layout intact */
function DashboardNotFound() {
  return (
    <div style={{
      background: '#FFFFFF',
      border: '1px solid #E2E8F0',
      borderRadius: '16px',
      padding: '48px 32px',
      textAlign: 'center',
      boxShadow: '0 1px 3px rgba(15, 23, 42, 0.03)',
      maxWidth: '600px',
      margin: '40px auto',
    }}>
      <div style={{ fontSize: '3rem', fontWeight: 900, color: '#CBD5E1', marginBottom: '8px', lineHeight: 1 }}>
        404
      </div>
      <div style={{ fontSize: '0.75rem', fontWeight: 800, textTransform: 'uppercase', letterSpacing: '0.08em', color: '#64748B', marginBottom: '8px' }}>
        LogisticsHQ Workspace
      </div>
      <h2 style={{ fontSize: '1.25rem', fontWeight: 800, color: '#0F172A', marginBottom: '8px' }}>
        We couldn't find this workspace page.
      </h2>
      <p style={{ fontSize: '0.85rem', color: '#64748B', maxWidth: '380px', margin: '0 auto 24px auto', lineHeight: 1.5 }}>
        The URL you requested doesn't exist or is not available in your organization's subscription.
      </p>
      <Link
        to="/dashboard"
        style={{
          display: 'inline-block',
          background: 'linear-gradient(135deg, #2563EB 0%, #4F46E5 100%)',
          color: '#FFFFFF',
          borderRadius: '9px',
          padding: '9px 20px',
          fontSize: '0.82rem',
          fontWeight: 700,
          textDecoration: 'none',
        }}
      >
        ← Back to Dashboard
      </Link>
    </div>
  );
}
import './App.css';

/**
 * RootRedirect — intelligent root route behavior.
 */
function RootRedirect() {
  const { isBooting, isAuthenticated, onboardingCompleted } = useAuth();
  if (isBooting) return <div className="boot-screen"><div className="auth-spinner" /></div>;

  if (isAuthenticated) {
    if (!onboardingCompleted) return <Navigate to="/onboarding" replace />;
    return <Navigate to="/dashboard" replace />;
  }

  return <Landing />;
}


/**
 * App.jsx — Root component with routing.
 *
 * All public pages are wrapped in PublicLayout (Navbar + Footer).
 * Each Route maps a URL path to a page component.
 *
 * Phase 2 note: RBAC is now provided via RBACProvider (inside AuthProvider).
 */


/** Branded placeholder for pages coming in Phase 2/3 */
function PlaceholderPage({ title, emoji, note = 'This page is being built and will be available soon.' }) {
  return (
    <section className="section-padding radial-glow-top">
      <div className="container-sm text-center">
        <div style={{ fontSize: '4rem', marginBottom: '24px' }}>{emoji}</div>
        <div className="section-label section-label-teal" style={{ marginBottom: '16px' }}>Coming Soon</div>
        <h1 style={{ fontSize: '2rem', fontWeight: 800, color: '#1E293B', marginBottom: '12px' }}>{title}</h1>
        <p style={{ color: '#64748B', maxWidth: '420px', margin: '0 auto 32px', lineHeight: 1.7 }}>{note}</p>
        <div style={{ display: 'flex', gap: '12px', justifyContent: 'center', flexWrap: 'wrap' }}>
          <Link to="/" className="btn-primary" style={{ textDecoration: 'none' }}>← Back to Home</Link>
          <Link to="/contact" className="btn-secondary" style={{ textDecoration: 'none' }}>Contact Us</Link>
        </div>
      </div>
    </section>
  );
}

/** 404 Not Found page */
function NotFoundPage() {
  return (
    <section className="section-padding">
      <div className="container-sm text-center">
        <div style={{ fontSize: '5rem', fontWeight: 900, color: '#E2E8F0', marginBottom: '8px', lineHeight: 1 }}>404</div>
        <div className="section-label section-label-slate" style={{ marginBottom: '16px' }}>Page Not Found</div>
        <h1 style={{ fontSize: '1.75rem', fontWeight: 800, color: '#1E293B', marginBottom: '12px' }}>
          We couldn't find that page.
        </h1>
        <p style={{ color: '#64748B', maxWidth: '400px', margin: '0 auto 32px', lineHeight: 1.7 }}>
          The URL you visited doesn't exist. It may have moved or been renamed.
        </p>
        <div style={{ display: 'flex', gap: '12px', justifyContent: 'center', flexWrap: 'wrap' }}>
          <Link to="/" className="btn-primary" style={{ textDecoration: 'none' }}>← Go Home</Link>
          <Link to="/services" className="btn-secondary" style={{ textDecoration: 'none' }}>View Services</Link>
        </div>
      </div>
    </section>
  );
}

export default function App() {
  const [showSplash, setShowSplash] = useState(() => {
    try {
      return !sessionStorage.getItem('lhq_splash_shown');
    } catch {
      return false;
    }
  });

  const handleSplashComplete = () => {
    try {
      sessionStorage.setItem('lhq_splash_shown', 'true');
    } catch {}
    setShowSplash(false);
  };

  return (
    <>
      {showSplash && <SplashScreen onComplete={handleSplashComplete} />}
      <AuthProvider>
        <RBACProvider>
          <BrowserRouter>
            <ScrollToTop />
            <Routes>

              {/* ── AUTH ROUTES (Split Panel Layout) ── */}
              <Route element={<PublicOnlyRoute />}>
                <Route element={<AuthLayout />}>
                  <Route path="/signup" element={<SignupPage />} />
                  <Route path="/verify-email" element={<VerifyEmailPage />} />
                  <Route path="/forgot-password" element={<ForgotPasswordPage />} />
                  <Route path="/reset-password" element={<ResetPasswordPage />} />
                  <Route path="/accept-invite" element={<AcceptInvitePage />} />
                </Route>

                {/* Standalone Auth Routes */}
                <Route path="/login" element={<LoginPage />} />
                <Route path="/auth/callback" element={<CallbackPage />} />
              </Route>

            {/* ── DEMO ROUTE ── */}
            <Route path="/demo-onboarding" element={<OnboardingPage />} />

            {/* ── PRIVATE PROTECTED ROUTES ── */}
            <Route element={<ProtectedRoute />}>
              <Route path="/onboarding" element={<OnboardingPage />} />
              <Route element={<AppShell />}>
                {/* ── 1. OPERATIONS ── */}
                <Route path="/dashboard" element={<DashboardHome />} />
                <Route path="/dashboard/leads" element={<LeadsPage />} />
                <Route path="/dashboard/rfqs" element={<RFQPage />} />
                <Route path="/dashboard/shipments" element={<ShipmentsPage />} />
                <Route path="/dashboard/shipments/:id" element={<ShipmentDetail />} />
                <Route path="/dashboard/bookings" element={<ShipmentsPage mode="bookings" defaultStatus="BOOKED" />} />
                <Route path="/dashboard/tracking" element={<TrackingPage />} />

                {/* ── 2. COMMERCIAL ── */}
                <Route path="/dashboard/quotations" element={<QuotationsPage />} />
                <Route path="/dashboard/rate-management" element={<RateManagementPage />} />
                <Route path="/dashboard/contracts" element={<ContractsPage />} />
                <Route path="/dashboard/companies" element={<CustomersPage />} />

                {/* ── 3. DOCUMENTS ── */}
                <Route path="/dashboard/documents" element={<DocumentsPage />} />
                <Route path="/dashboard/templates" element={<TemplatesPage />} />
                <Route path="/dashboard/approvals" element={<ApprovalsPage />} />

                {/* ── 4. FINANCE ── */}
                <Route path="/dashboard/invoices" element={<InvoicesPage />} />
                <Route path="/dashboard/payments" element={<PaymentsPage />} />
                <Route path="/dashboard/reports" element={<ReportsPage />} />

                {/* ── 5. OUTREACH & TOOLS ── */}
                <Route path="/dashboard/outreach" element={<OutreachPage />} />
                <Route path="/dashboard/market-insights" element={<WorkspacePlaceholder section="Intelligence" title="Market Insights" emoji="📊" note="Global freight rate trends, port congestion analytics, and fuel bunker surcharges." />} />
                <Route path="/dashboard/routes" element={<WorkspacePlaceholder section="Intelligence" title="Route Optimization" emoji="🗺️" note="AI-driven shipment routing and carbon emission estimation." />} />
                <Route path="/dashboard/calculators" element={<WorkspacePlaceholder section="Tools" title="Freight Calculators" emoji="🧮" note="CBM, Volumetric Weight, and Duty calculators." />} />

                {/* ── 6. ADMIN & SETTINGS ── */}
                <Route path="/dashboard/users" element={<UsersPage />} />
                <Route path="/dashboard/settings" element={<RolesPage />} />

                {/* ── 7. PERMISSION & 404 FALLBACKS WITHIN WORKSPACE ── */}
                <Route path="/dashboard/unauthorized" element={<WorkspacePlaceholder section="Security" title="Unauthorized" emoji="🚫" note="You do not have permission to access this module. Please contact your organization administrator." />} />
                <Route path="/dashboard/*" element={<DashboardNotFound />} />
              </Route>
            </Route>

            {/* All public pages wrapped in PublicLayout (Navbar + Footer) */}
            <Route element={<PublicLayout />}>
              <Route path="/" element={<RootRedirect />} />

              {/* Services */}
              <Route path="/services" element={<Services />} />
              <Route path="/services/air-freight" element={<AirFreight />} />
              <Route path="/services/sea-freight" element={<SeaFreight />} />
              <Route path="/services/road-transport" element={<RoadTransport />} />
              <Route path="/services/customs" element={<CustomsBrokerage />} />
              <Route path="/services/rail-freight" element={<ComingSoonPage title="Rail Freight" icon="🚆" category="Services" />} />
              <Route path="/services/trade-finance" element={<ComingSoonPage title="Trade Finance" icon="🏦" category="Services" />} />
              <Route path="/services/insurance" element={<ComingSoonPage title="Insurance" icon="🛡️" category="Services" />} />
              <Route path="/services/documentation" element={<ComingSoonPage title="Documentation" icon="📄" category="Services" />} />
              <Route path="/coverage" element={<ComingSoonPage title="Coverage Map" icon="🌍" category="Resources" />} />

              {/* Solutions */}
              <Route path="/solutions" element={<Solutions />} />
              <Route path="/solutions/rfq" element={<RFQLanding />} />
              <Route path="/solutions/rate-comparison" element={<RateComparison />} />
              <Route path="/solutions/tracking" element={<ShipmentTracking />} />
              <Route path="/solutions/compliance" element={<Compliance />} />
              <Route path="/solutions/procurement" element={<ComingSoonPage title="Procurement" icon="🛒" category="Solutions" />} />
              <Route path="/solutions/route" element={<ComingSoonPage title="Route Optimization" icon="🛣️" category="Solutions" />} />
              <Route path="/solutions/analytics" element={<ComingSoonPage title="Analytics" icon="📊" category="Solutions" />} />
              <Route path="/solutions/reporting" element={<ComingSoonPage title="Reporting" icon="📑" category="Solutions" />} />
              <Route path="/solutions/api" element={<ComingSoonPage title="API Access" icon="🔗" category="Integrations" />} />
              <Route path="/solutions/erp" element={<ComingSoonPage title="ERP Integration" icon="🔄" category="Integrations" />} />
              <Route path="/solutions/webhooks" element={<ComingSoonPage title="Webhooks" icon="📡" category="Integrations" />} />
              <Route path="/solutions/edi" element={<ComingSoonPage title="EDI Support" icon="🧩" category="Integrations" />} />

              {/* Blog */}
              <Route path="/blog" element={<BlogIndex />} />
              <Route path="/blog/engineering" element={<EngineeringBlog />} />
              <Route path="/blog/design" element={<DesignBlog />} />
              <Route path="/blog/industry" element={<IndustryBlog />} />

              {/* Company */}
              <Route path="/about" element={<About />} />
              <Route path="/contact" element={<Contact />} />
              <Route path="/platform" element={<Platform />} />

              {/* Phase 2 auth routes removed from here (handled above) */}

              {/* Phase 3 placeholders */}
              <Route path="/products" element={<PlaceholderPage title="Products" emoji="📦" note="Our full product suite page is coming soon. In the meantime, explore our Services." />} />
              <Route path="/partners" element={<PlaceholderPage title="Partners" emoji="🤝" note="Our partner program is launching soon. Contact us to express interest." />} />
              <Route path="/resources" element={<PlaceholderPage title="Resources" emoji="📚" note="Help center, API docs, and weight calculators are coming soon." />} />
              <Route path="/track" element={<PlaceholderPage title="Track Order" emoji="🔍" note="Live shipment tracking is available on the dashboard after signing up." />} />

              {/* Trade Intelligence - Guides */}
              <Route path="/knowledge" element={<ComingSoonPage title="Trade Intelligence Hub" icon="🧠" category="Knowledge Base" />} />
              <Route path="/knowledge/incoterms" element={<IncotermsPage />} />
              <Route path="/knowledge/air-freight" element={<AirFreightPage />} />
              <Route path="/knowledge/sea-freight" element={<ComingSoonPage title="Sea Freight Guide" icon="🚢" category="Guides" />} />
              <Route path="/knowledge/customs" element={<ComingSoonPage title="Customs Clearance Guide" icon="🛡️" category="Guides" />} />
              <Route path="/knowledge/documentation" element={<DocumentationGuidePage />} />
              <Route path="/knowledge/import-export" element={<ImportExportBasicsPage />} />

              {/* Trade Intelligence - Calculators */}
              <Route path="/tools/cbm-calculator" element={<ComingSoonPage title="CBM Calculator" icon="📦" category="Calculators" />} />
              <Route path="/tools/volumetric-weight" element={<ComingSoonPage title="Volumetric Weight" icon="⚖️" category="Calculators" />} />
              <Route path="/tools/duty-calculator" element={<ComingSoonPage title="Duty Calculator" icon="🧮" category="Calculators" />} />
              <Route path="/tools/transit-time" element={<ComingSoonPage title="Transit Time Estimator" icon="⏱️" category="Calculators" />} />
              <Route path="/tools/freight-cost" element={<ComingSoonPage title="Freight Cost Calculator" icon="💰" category="Calculators" />} />
              <Route path="/tools/container-load" element={<ComingSoonPage title="Container Load Planner" icon="🏗️" category="Calculators" />} />

              {/* Trade Intelligence - References */}
              <Route path="/reference/container-sizes" element={<ComingSoonPage title="Container Sizes Guide" icon="📐" category="References" />} />
              <Route path="/reference/ports" element={<ComingSoonPage title="Port Directory" icon="⚓" category="References" />} />
              <Route path="/reference/airports" element={<ComingSoonPage title="Airport Directory" icon="🛫" category="References" />} />
              <Route path="/reference/hsn-codes" element={<ComingSoonPage title="HSN / HS Codes" icon="🔢" category="References" />} />
              <Route path="/reference/dangerous-goods" element={<ComingSoonPage title="Dangerous Goods Guide" icon="⚠️" category="References" />} />
              <Route path="/reference/trade-profiles" element={<ComingSoonPage title="Country Trade Profiles" icon="🗺️" category="References" />} />

              {/* Trade Intelligence - Insights */}
              <Route path="/insights/trends" element={<ComingSoonPage title="Logistics Trends" icon="📈" category="Insights" />} />
              <Route path="/insights/market-updates" element={<ComingSoonPage title="Market Updates" icon="📰" category="Insights" />} />
              <Route path="/insights/news" element={<ComingSoonPage title="Trade News" icon="🗞️" category="Insights" />} />
              <Route path="/insights/reports" element={<ComingSoonPage title="Industry Reports" icon="📋" category="Insights" />} />
              <Route path="/insights/benchmarks" element={<ComingSoonPage title="Logistics Benchmarks" icon="🏆" category="Insights" />} />
              <Route path="/insights/cases" element={<ComingSoonPage title="Case Studies" icon="🤝" category="Insights" />} />

              {/* Trade Intelligence - Coming Soon Fallback */}
              <Route path="/trade-intelligence/coming-soon" element={<ComingSoonPage />} />

              {/* 404 fallback */}
              <Route path="*" element={<NotFoundPage />} />
            </Route>
          </Routes>
        </BrowserRouter>
      </RBACProvider>
    </AuthProvider>
  </>
  );
}

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
import RequestDemoPage from './pages/public/Demo/RequestDemoPage';
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
import AcceptInvitePage from './pages/auth/AcceptInvite/AcceptInvitePage';
import PublicOnlyRoute from './routes/PublicOnlyRoute';
import ProtectedRoute from './routes/ProtectedRoute';
import AppShell from './layouts/AppShell/AppShell';
import DashboardHome from './pages/dashboard/Home/DashboardHome';
import ReportsPage from './pages/dashboard/Reports/ReportsPage';
import UsersPage from './pages/dashboard/Settings/UsersPage';
import CompanyProfilePage from './pages/dashboard/Settings/CompanyProfilePage';
import WorkspaceSettingsPage from './pages/dashboard/Settings/WorkspaceSettingsPage';
import EmailSettingsPage from './pages/dashboard/Settings/EmailSettingsPage';
import RolesPage from './pages/dashboard/Settings/RolesPage';
import CarrierIntegrationsPage from './pages/dashboard/Settings/CarrierIntegrationsPage';
import ExternalIntegrationsPage from './pages/dashboard/Settings/ExternalIntegrationsPage';
import AuditLogsPage from './pages/dashboard/Settings/AuditLogsPage';
import LeadsPage from './pages/dashboard/Leads/LeadsPage';
import OutreachPage from './pages/dashboard/Outreach/OutreachPage';
import RFQPage from './pages/dashboard/RFQ/RFQPage';
import RFQDetailPage from './pages/dashboard/RFQ/RFQDetailPage';
import BookingsPage from './pages/dashboard/Bookings/BookingsPage';
import BookingDetailPage from './pages/dashboard/Bookings/BookingDetailPage';
import ShipmentsPage from './pages/dashboard/Shipments/ShipmentsPage';
import ShipmentDetail from './pages/dashboard/Shipments/ShipmentDetail';
import ContractsPage from './pages/dashboard/Contracts/ContractsPage';
import ContractDocumentsPage from './pages/dashboard/Contracts/ContractDocumentsPage';
import RateManagementPage from './pages/dashboard/RateManagement/RateManagementPage';
import QuotationsPage from './pages/dashboard/Quotations/QuotationsPage';
import TrackingPage from './pages/dashboard/Tracking/TrackingPage';
import TrackingDetailPage from './pages/dashboard/Tracking/TrackingDetailPage';
import ApprovalsPage from './pages/dashboard/Approvals/ApprovalsPage';
import RecommendationCenterPage from './pages/dashboard/Recommendations/RecommendationCenterPage';
import WorkflowAutomationsPage from './pages/dashboard/Automations/WorkflowAutomationsPage';
import NotificationCenterPage from './pages/dashboard/Notifications/NotificationCenterPage';
import AIMemorySettingsPage from './pages/dashboard/Settings/AIMemorySettingsPage';
import AIMonitoringDashboardPage from './pages/dashboard/Monitoring/AIMonitoringDashboardPage';
import AutonomousCommandCenterPage from './pages/dashboard/CommandCenter/AutonomousCommandCenterPage';
import InvoicesPage from './pages/dashboard/Finance/InvoicesPage';
import DebitNotesPage from './pages/dashboard/Finance/DebitNotesPage';
import CustomersPage from './pages/dashboard/Customers/CustomersPage';
import CustomerDetailsPage from './pages/dashboard/Customers/CustomerDetailsPage';
import DocumentsPage from './pages/dashboard/Documents/DocumentsPage';
import SettingsLayout from './layouts/SettingsLayout/SettingsLayout';
import SubscriptionPage from './pages/dashboard/Settings/SubscriptionPage';

/** Branded workspace access restricted view for unauthorized attempts */
function WorkspaceUnauthorized() {
  return (
    <div style={{
      background: '#FFFFFF',
      border: '1px solid #E2E8F0',
      borderRadius: '16px',
      padding: '48px 32px',
      textAlign: 'center',
      boxShadow: '0 1px 3px rgba(15, 23, 42, 0.03)',
      maxWidth: '560px',
      margin: '40px auto',
    }}>
      <div style={{
        width: '56px',
        height: '56px',
        borderRadius: '14px',
        background: '#FEF2F2',
        border: '1px solid #FEE2E2',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        fontSize: '1.6rem',
        margin: '0 auto 16px auto',
        color: '#DC2626'
      }}>
        🛡️
      </div>
      <div style={{
        fontSize: '0.72rem',
        fontWeight: 800,
        textTransform: 'uppercase',
        letterSpacing: '0.08em',
        color: '#DC2626',
        marginBottom: '6px',
      }}>
        Access Restricted
      </div>
      <h2 style={{ fontSize: '1.3rem', fontWeight: 800, color: '#0F172A', marginBottom: '8px' }}>
        Permission Required
      </h2>
      <p style={{ fontSize: '0.85rem', color: '#64748B', maxWidth: '420px', margin: '0 auto 24px auto', lineHeight: 1.5 }}>
        Your current workspace role does not have permission to access this resource. Please contact your organization administrator to request access.
      </p>
      <div style={{ display: 'flex', gap: '12px', justifyContent: 'center' }}>
        <Link
          to="/dashboard"
          style={{
            background: '#0B192C',
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
  const { isBooting, isAuthenticated } = useAuth();
  if (isBooting) return <div className="boot-screen"><div className="auth-spinner" /></div>;

  if (isAuthenticated) {
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
                  <Route path="/invite/accept" element={<AcceptInvitePage />} />
                </Route>

                {/* Standalone Auth Routes */}
                <Route path="/login" element={<LoginPage />} />
                <Route path="/auth/callback" element={<CallbackPage />} />
              </Route>

            {/* ── PRIVATE PROTECTED ROUTES ── */}
            <Route element={<ProtectedRoute />}>
              <Route element={<AppShell />}>
                {/* ── 1. OPERATIONS ── */}
                <Route path="/dashboard" element={<DashboardHome />} />
                <Route path="/dashboard/command-center" element={<AutonomousCommandCenterPage />} />
                <Route path="/dashboard/leads" element={<LeadsPage />} />
                <Route path="/dashboard/rfqs" element={<RFQPage />} />
                <Route path="/dashboard/rfqs/:id" element={<RFQDetailPage />} />
                <Route path="/dashboard/shipments" element={<ShipmentsPage />} />
                <Route path="/dashboard/shipments/:id" element={<ShipmentDetail />} />
                <Route path="/dashboard/bookings" element={<BookingsPage />} />
                <Route path="/dashboard/bookings/:bookingId" element={<BookingDetailPage />} />
                <Route path="/dashboard/tracking" element={<TrackingPage />} />
                <Route path="/dashboard/tracking/:shipmentId" element={<TrackingDetailPage />} />

                {/* ── 2. COMMERCIAL ── */}
                <Route path="/dashboard/quotations" element={<QuotationsPage />} />
                <Route path="/dashboard/rate-management" element={<RateManagementPage />} />
                <Route path="/dashboard/contracts" element={<ContractsPage />} />
                <Route path="/dashboard/contract-documents" element={<ContractDocumentsPage />} />
                <Route path="/dashboard/customers" element={<CustomersPage />} />
                <Route path="/dashboard/customers/:id" element={<CustomerDetailsPage />} />
                <Route path="/dashboard/companies" element={<CustomersPage />} />

                {/* ── 3. DOCUMENTS ── */}
                <Route path="/dashboard/documents" element={<DocumentsPage />} />
                <Route path="/dashboard/approvals" element={<ApprovalsPage />} />
                <Route path="/dashboard/recommendations" element={<RecommendationCenterPage />} />
                <Route path="/dashboard/automations" element={<WorkflowAutomationsPage />} />
                <Route path="/dashboard/notifications" element={<NotificationCenterPage />} />
                <Route path="/dashboard/ai-monitoring" element={<AIMonitoringDashboardPage />} />

                {/* ── 4. FINANCE ── */}
                <Route path="/dashboard/invoices" element={<InvoicesPage />} />
                <Route path="/dashboard/finance/invoices" element={<InvoicesPage />} />
                <Route path="/dashboard/debit-notes" element={<DebitNotesPage />} />
                <Route path="/dashboard/payments" element={<Navigate to="/dashboard/invoices" replace />} />
                <Route path="/dashboard/reports" element={<ReportsPage />} />

                {/* ── 5. OUTREACH & TOOLS ── */}
                <Route path="/dashboard/outreach" element={<OutreachPage />} />
                <Route path="/dashboard/market-insights" element={<Navigate to="/dashboard/rate-management" replace />} />
                <Route path="/dashboard/routes" element={<Navigate to="/dashboard/tracking" replace />} />
                <Route path="/dashboard/calculators" element={<Navigate to="/dashboard/rate-management" replace />} />

                {/* ── 6. ADMIN & SETTINGS ── */}
                <Route path="/dashboard/settings" element={<SettingsLayout />}>
                  <Route index element={<Navigate to="roles" replace />} />
                  
                  {/* SETTINGS */}
                  <Route path="company-profile" element={<CompanyProfilePage />} />
                  <Route path="company" element={<Navigate to="company-profile" replace />} />
                  <Route path="workspace" element={<WorkspaceSettingsPage />} />
                  <Route path="automations" element={<WorkflowAutomationsPage />} />
                  <Route path="memory" element={<AIMemorySettingsPage />} />
                  <Route path="monitoring" element={<AIMonitoringDashboardPage />} />
                  
                  {/* TEAM & ACCESS */}
                  <Route path="users" element={<UsersPage />} />
                  <Route path="roles" element={<RolesPage />} />
                  
                  {/* SECURITY */}
                  <Route path="audit-logs" element={<AuditLogsPage />} />
                  
                  {/* INTEGRATIONS */}
                  <Route path="external-integrations" element={<ExternalIntegrationsPage />} />
                  <Route path="carrier-integrations" element={<CarrierIntegrationsPage />} />
                  <Route path="email-settings" element={<EmailSettingsPage />} />
                  {/* BILLING */}
                  <Route path="subscription" element={<SubscriptionPage />} />
                  <Route path="billing-invoices" element={<Navigate to="/dashboard/invoices" replace />} />
                </Route>

                {/* Direct & legacy redirects for nested pages accessed directly */}
                <Route path="/dashboard/users" element={<Navigate to="/dashboard/settings/users" replace />} />
                <Route path="/dashboard/audit-logs" element={<Navigate to="/dashboard/settings/audit-logs" replace />} />
                <Route path="/dashboard/audit" element={<Navigate to="/dashboard/settings/audit-logs" replace />} />
                <Route path="/dashboard/ai/workforce" element={<Navigate to="/dashboard/ai-monitoring?tab=workforce" replace />} />
                <Route path="/dashboard/ai-workforce" element={<Navigate to="/dashboard/ai-monitoring?tab=workforce" replace />} />

                {/* ── 7. PERMISSION & 404 FALLBACKS WITHIN WORKSPACE ── */}
                <Route path="/dashboard/unauthorized" element={<WorkspaceUnauthorized />} />
                <Route path="/dashboard/*" element={<DashboardNotFound />} />
              </Route>
            </Route>

            {/* Direct convenience aliases for root workspace routes */}
            <Route path="/shipments" element={<Navigate to="/dashboard/shipments" replace />} />
            <Route path="/rfqs" element={<Navigate to="/dashboard/rfqs" replace />} />
            <Route path="/quotations" element={<Navigate to="/dashboard/quotations" replace />} />
            <Route path="/quotes" element={<Navigate to="/dashboard/quotations" replace />} />
            <Route path="/bookings" element={<Navigate to="/dashboard/bookings" replace />} />
            <Route path="/invoices" element={<Navigate to="/dashboard/invoices" replace />} />
            <Route path="/leads" element={<Navigate to="/dashboard/leads" replace />} />
            <Route path="/documents" element={<Navigate to="/dashboard/documents" replace />} />
            <Route path="/approvals" element={<Navigate to="/dashboard/approvals" replace />} />
            <Route path="/tracking" element={<Navigate to="/dashboard/tracking" replace />} />
            <Route path="/settings" element={<Navigate to="/dashboard/settings/roles" replace />} />

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
              <Route path="/demo" element={<RequestDemoPage />} />
              <Route path="/request-demo" element={<Navigate to="/demo" replace />} />

              {/* Phase 2 auth routes removed from here (handled above) */}

              {/* Functional route redirects */}
              <Route path="/products" element={<Navigate to="/services" replace />} />
              <Route path="/partners" element={<Navigate to="/contact" replace />} />
              <Route path="/resources" element={<Navigate to="/blog" replace />} />
              <Route path="/track" element={<Navigate to="/solutions/tracking" replace />} />

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

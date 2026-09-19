import React, { lazy } from 'react';
import { Routes, Route, Navigate } from 'react-router-dom';
import { SPortalLayout } from '../../components/layout/SPortalLayout';
import { ProtectedRoute } from '../../components/common/ProtectedRoute';
import { LoginPage } from '../../features/auth/LoginPage';
import { UnauthorizedPage } from '../../features/auth/UnauthorizedPage';

// Primary landing view is eagerly loaded for instant initial render
import { DashboardOverview } from '../../features/dashboard/DashboardOverview';

// Secondary views are lazily code-split for performance and optimal bundle chunking
const OrganizationsPage = lazy(() => import('../../features/organizations/OrganizationsPage').then((m) => ({ default: m.OrganizationsPage })));
const OrganizationDetailPage = lazy(() => import('../../features/organizations/OrganizationDetailPage').then((m) => ({ default: m.OrganizationDetailPage })));
const OnboardingPage = lazy(() => import('../../features/onboarding/OnboardingPage').then((m) => ({ default: m.OnboardingPage })));
const SubscriptionsPage = lazy(() => import('../../features/subscriptions/SubscriptionsPage').then((m) => ({ default: m.SubscriptionsPage })));
const BillingPage = lazy(() => import('../../features/billing/BillingPage').then((m) => ({ default: m.BillingPage })));
const UsersPage = lazy(() => import('../../features/users/UsersPage').then((m) => ({ default: m.UsersPage })));
const Customer360Page = lazy(() => import('../../features/customer360/Customer360Page').then((m) => ({ default: m.Customer360Page })));
const UsagePage = lazy(() => import('../../features/usage/UsagePage').then((m) => ({ default: m.UsagePage })));
const CustomerHealthPage = lazy(() => import('../../features/health/CustomerHealthPage').then((m) => ({ default: m.CustomerHealthPage })));
const IntegrationsPage = lazy(() => import('../../features/integrations/IntegrationsPage').then((m) => ({ default: m.IntegrationsPage })));
const DocumentsPage = lazy(() => import('../../features/documents/DocumentsPage').then((m) => ({ default: m.DocumentsPage })));
const SupportPage = lazy(() => import('../../features/support/SupportPage').then((m) => ({ default: m.SupportPage })));
const SportalAiPage = lazy(() => import('../../features/ai/SportalAiPage').then((m) => ({ default: m.SportalAiPage })));
const SettingsPage = lazy(() => import('../../features/settings/SettingsPage').then((m) => ({ default: m.SettingsPage })));
const DemoRequestsPage = lazy(() => import('../../features/demo_requests/DemoRequestsPage').then((m) => ({ default: m.DemoRequestsPage })));

export function AppRoutes() {
  return (
    <Routes>
      {/* Public Authentication Routes */}
      <Route path="/login" element={<LoginPage />} />
      <Route path="/unauthorized" element={<UnauthorizedPage />} />

      {/* Protected SPortal Administration Shell */}
      <Route
        element={
          <ProtectedRoute>
            <SPortalLayout />
          </ProtectedRoute>
        }
      >
        <Route index element={<DashboardOverview />} />
        <Route path="organizations" element={<OrganizationsPage />} />
        <Route path="organizations/:organizationId" element={<OrganizationDetailPage />} />
        <Route path="onboarding" element={<OnboardingPage />} />
        <Route path="subscriptions" element={<SubscriptionsPage />} />
        <Route
          path="billing"
          element={
            <ProtectedRoute requiredPermission="billing:view">
              <BillingPage />
            </ProtectedRoute>
          }
        />
        <Route
          path="users"
          element={
            <ProtectedRoute requiredPermission="users:view">
              <UsersPage />
            </ProtectedRoute>
          }
        />
        <Route
          path="organizations/:organizationId/users"
          element={
            <ProtectedRoute requiredPermission="users:view">
              <UsersPage />
            </ProtectedRoute>
          }
        />
        <Route path="customer-360" element={<Customer360Page />} />
        <Route path="customer-360/:organizationId" element={<OrganizationDetailPage />} />
        <Route path="organizations/:organizationId/customer-360" element={<OrganizationDetailPage />} />
        <Route path="organizations/:organizationId/usage" element={<UsagePage />} />
        <Route path="usage" element={<UsagePage />} />
        <Route path="organizations/:organizationId/health" element={<CustomerHealthPage />} />
        <Route path="customer-health" element={<CustomerHealthPage />} />
        <Route path="health" element={<CustomerHealthPage />} />
        <Route
          path="organizations/:organizationId/integrations"
          element={
            <ProtectedRoute requiredPermission="integrations:view">
              <IntegrationsPage />
            </ProtectedRoute>
          }
        />
        <Route
          path="integrations"
          element={
            <ProtectedRoute requiredPermission="integrations:view">
              <IntegrationsPage />
            </ProtectedRoute>
          }
        />
        <Route path="documents" element={<DocumentsPage />} />
        <Route path="support" element={<SupportPage />} />
        <Route path="ai" element={<SportalAiPage />} />
        <Route path="organizations/:organizationId/ai" element={<SportalAiPage />} />
        <Route
          path="settings"
          element={
            <ProtectedRoute requiredPermission="settings:view">
              <SettingsPage />
            </ProtectedRoute>
          }
        />
        <Route path="demo-requests" element={<DemoRequestsPage />} />
        <Route path="*" element={<Navigate to="/" replace />} />
      </Route>
    </Routes>
  );
}

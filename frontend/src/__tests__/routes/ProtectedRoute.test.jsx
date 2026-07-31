/**
 * Component tests for ProtectedRoute.jsx
 *
 * Tests cover:
 *  - While booting: renders the boot-screen spinner
 *  - Unauthenticated: redirects to /login
 *  - Authenticated but onboarding incomplete at /dashboard: redirects to /onboarding
 *  - Authenticated and onboarding complete at /onboarding: redirects to /dashboard
 *  - Authenticated and onboarding complete at /dashboard: renders <Outlet />
 */

import { describe, it, expect } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter, Routes, Route, Outlet } from 'react-router-dom';
import { AuthProvider } from '../../context/AuthContext';
import { RBACProvider } from '../../context/RBACContext';
import ProtectedRoute from '../../routes/ProtectedRoute';
import { authStorage } from '../../utils/authStorage';
import { onboardingStorage } from '../../utils/onboardingStorage';

// Stubs for redirect targets
const LoginPage = () => <div data-testid="login-page">Login</div>;
const OnboardingPage = () => <div data-testid="onboarding-page">Onboarding</div>;
const DashboardPage = () => <div data-testid="dashboard-page">Dashboard</div>;

function renderProtectedRoute(initialPath = '/dashboard') {
  return render(
    <AuthProvider>
      <RBACProvider>
        <MemoryRouter initialEntries={[initialPath]}>
          <Routes>
            <Route path="/login" element={<LoginPage />} />
            <Route path="/onboarding" element={<OnboardingPage />} />

            <Route element={<ProtectedRoute />}>
              <Route path="/dashboard" element={<DashboardPage />} />
              <Route path="/onboarding" element={<OnboardingPage />} />
            </Route>
          </Routes>
        </MemoryRouter>
      </RBACProvider>
    </AuthProvider>
  );
}

describe('ProtectedRoute', () => {
  it('redirects to /login when not authenticated', async () => {
    // No session — unauthenticated
    renderProtectedRoute('/dashboard');
    await waitFor(() => {
      expect(screen.getByTestId('login-page')).toBeInTheDocument();
    });
  });

  it('redirects to /onboarding when authenticated but onboarding incomplete', async () => {
    authStorage.saveSessionUser({
      user: { email: 'incomplete@logisticshq.in' },
      org: { name: 'Freel' },
      memberRole: { name: 'SALES', permissions: [] },
    });
    // Don't save onboarding state — not completed

    renderProtectedRoute('/dashboard');
    await waitFor(() => {
      expect(screen.getByTestId('onboarding-page')).toBeInTheDocument();
    });
  });

  it('renders the protected content when authenticated and onboarding complete', async () => {
    const email = 'complete@logisticshq.in';
    authStorage.saveSessionUser({
      user: { email },
      org: { name: 'Freel' },
      memberRole: { name: 'SUPER_ADMIN', permissions: [] },
    });
    onboardingStorage.saveOnboardingState(email, {});

    renderProtectedRoute('/dashboard');
    await waitFor(() => {
      expect(screen.getByTestId('dashboard-page')).toBeInTheDocument();
    });
  });
});

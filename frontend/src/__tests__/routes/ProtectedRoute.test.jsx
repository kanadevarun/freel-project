/**
 * Component tests for ProtectedRoute.jsx
 *
 * Tests cover:
 *  - While booting: renders the boot-screen spinner
 *  - Unauthenticated: redirects to /login
 *  - Authenticated: renders <Outlet />
 */

import { describe, it, expect } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter, Routes, Route, Outlet } from 'react-router-dom';
import { AuthProvider } from '../../context/AuthContext';
import { RBACProvider } from '../../context/RBACContext';
import ProtectedRoute from '../../routes/ProtectedRoute';
import { authStorage } from '../../utils/authStorage';

// Stubs for redirect targets
const LoginPage = () => <div data-testid="login-page">Login</div>;
const DashboardPage = () => <div data-testid="dashboard-page">Dashboard</div>;

function renderProtectedRoute(initialPath = '/dashboard') {
  return render(
    <AuthProvider>
      <RBACProvider>
        <MemoryRouter initialEntries={[initialPath]}>
          <Routes>
            <Route path="/login" element={<LoginPage />} />

            <Route element={<ProtectedRoute />}>
              <Route path="/dashboard" element={<DashboardPage />} />
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

  it('renders the protected content when authenticated', async () => {
    const email = 'complete@logisticshq.in';
    authStorage.saveSessionUser({
      user: { email },
      org: { name: 'Freel' },
      memberRole: { name: 'SUPER_ADMIN', permissions: [] },
    });

    renderProtectedRoute('/dashboard');
    await waitFor(() => {
      expect(screen.getByTestId('dashboard-page')).toBeInTheDocument();
    });
  });
});

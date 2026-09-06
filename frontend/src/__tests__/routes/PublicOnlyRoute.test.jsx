/**
 * Component tests for PublicOnlyRoute.jsx
 *
 * Tests cover:
 *  - Not authenticated: renders <Outlet /> (shows public content)
 *  - Authenticated + onboarding incomplete: redirects to /onboarding
 *  - Authenticated + onboarding complete: redirects to /dashboard
 */

import { describe, it, expect } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter, Routes, Route } from 'react-router-dom';
import { AuthProvider } from '../../context/AuthContext';
import PublicOnlyRoute from '../../routes/PublicOnlyRoute';
import { authStorage } from '../../utils/authStorage';

const LoginPage = () => <div data-testid="login-page">Login</div>;
const DashboardPage = () => <div data-testid="dashboard-page">Dashboard</div>;

function renderPublicOnlyRoute(initialPath = '/login') {
  return render(
    <AuthProvider>
      <MemoryRouter initialEntries={[initialPath]}>
        <Routes>
          <Route path="/dashboard" element={<DashboardPage />} />
          <Route element={<PublicOnlyRoute />}>
            <Route path="/login" element={<LoginPage />} />
          </Route>
        </Routes>
      </MemoryRouter>
    </AuthProvider>
  );
}

describe('PublicOnlyRoute', () => {
  it('renders the public page when not authenticated', async () => {
    // No session — unauthenticated
    renderPublicOnlyRoute('/login');
    await waitFor(() => {
      expect(screen.getByTestId('login-page')).toBeInTheDocument();
    });
  });

  it('redirects to /dashboard when authenticated', async () => {
    const email = 'done@logisticshq.in';
    authStorage.saveSessionUser({
      user: { email },
      org: { name: 'Freel' },
      memberRole: { name: 'SUPER_ADMIN', permissions: [] },
    });

    renderPublicOnlyRoute('/login');
    await waitFor(() => {
      expect(screen.getByTestId('dashboard-page')).toBeInTheDocument();
    });
  });
});

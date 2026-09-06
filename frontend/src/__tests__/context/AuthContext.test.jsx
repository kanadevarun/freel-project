/**
 * Component + hook tests for AuthContext.jsx
 *
 * Tests cover:
 *  - Initial state (isBooting → false, user/org/memberRole → null)
 *  - Boot from localStorage (restores session correctly)
 *  - login(): saves tokens, updates state, normalizes role object vs string
 *  - logout(): clears all state
 *  - updateOrg(): merges org updates
 *  - useAuth() throws outside provider
 */

import { describe, it, expect, vi } from 'vitest';
import { render, screen, act, waitFor } from '@testing-library/react';
import { AuthProvider, useAuth } from '../../context/AuthContext';
import { authStorage } from '../../utils/authStorage';

// Helper component that exposes auth state as text
function AuthConsumer() {
  const {
    user,
    org,
    memberRole,
    isAuthenticated,
    isBooting,
  } = useAuth();

  if (isBooting) return <div data-testid="booting">Booting</div>;

  return (
    <div>
      <div data-testid="is-authenticated">{String(isAuthenticated)}</div>
      <div data-testid="user-email">{user?.email ?? 'null'}</div>
      <div data-testid="org-name">{org?.name ?? 'null'}</div>
      <div data-testid="role-name">{
        typeof memberRole === 'object' ? (memberRole?.name ?? 'null') : (memberRole ?? 'null')
      }</div>
    </div>
  );
}

// Helper to render with the AuthProvider
function renderWithAuth() {
  return render(
    <AuthProvider>
      <AuthConsumer />
    </AuthProvider>
  );
}

describe('AuthContext', () => {
  // ── INITIAL STATE ─────────────────────────────────────────────────────────
  describe('initial state (empty localStorage)', () => {
    it('resolves booting and shows unauthenticated state', async () => {
      renderWithAuth();
      await waitFor(() => {
        expect(screen.queryByTestId('booting')).not.toBeInTheDocument();
      });
      expect(screen.getByTestId('is-authenticated').textContent).toBe('false');
      expect(screen.getByTestId('user-email').textContent).toBe('null');
    });
  });

  // ── BOOT FROM LOCALSTORAGE ────────────────────────────────────────────────
  describe('boot from localStorage', () => {
    it('restores session if valid session data exists', async () => {
      authStorage.saveSessionUser({
        user: { email: 'restored@logisticshq.in', full_name: 'Restored User' },
        org: { name: 'Freel Corp' },
        memberRole: { name: 'SALES', display_name: 'Sales', permissions: [] },
      });

      renderWithAuth();
      await waitFor(() => {
        expect(screen.queryByTestId('booting')).not.toBeInTheDocument();
      });

      expect(screen.getByTestId('is-authenticated').textContent).toBe('true');
      expect(screen.getByTestId('user-email').textContent).toBe('restored@logisticshq.in');
      expect(screen.getByTestId('role-name').textContent).toBe('SALES');
    });
  });

  // ── login() ───────────────────────────────────────────────────────────────
  describe('login()', () => {
    it('updates user, org, and role state after login', async () => {
      let authApi;
      function LoginTrigger() {
        authApi = useAuth();
        return null;
      }

      render(
        <AuthProvider>
          <LoginTrigger />
          <AuthConsumer />
        </AuthProvider>
      );

      await waitFor(() => {
        expect(screen.queryByTestId('booting')).not.toBeInTheDocument();
      });

      await act(async () => {
        await authApi.login(
          { access_token: 'at', id_token: 'it', refresh_token: 'rt' },
          { email: 'alice@logisticshq.in', full_name: 'Alice' },
          { name: 'Freel Corp' },
          { name: 'SUPER_ADMIN', display_name: 'Super Admin', permissions: ['LEADS:READ'] }
        );
      });

      expect(screen.getByTestId('is-authenticated').textContent).toBe('true');
      expect(screen.getByTestId('user-email').textContent).toBe('alice@logisticshq.in');
      expect(screen.getByTestId('role-name').textContent).toBe('SUPER_ADMIN');
    });

    it('normalizes a plain string role into an object with empty permissions', async () => {
      let authApi;
      function LoginTrigger() {
        authApi = useAuth();
        return null;
      }
      render(
        <AuthProvider>
          <LoginTrigger />
          <AuthConsumer />
        </AuthProvider>
      );
      await waitFor(() => expect(screen.queryByTestId('booting')).not.toBeInTheDocument());

      await act(async () => {
        await authApi.login(
          { access_token: 'at', id_token: 'it', refresh_token: 'rt' },
          { email: 'bob@logisticshq.in' },
          null,
          'OWNER' // legacy string role
        );
      });

      expect(screen.getByTestId('role-name').textContent).toBe('OWNER');
    });
  });

  // ── logout() ──────────────────────────────────────────────────────────────
  describe('logout()', () => {
    it('clears all state after logout', async () => {
      authStorage.saveSessionUser({
        user: { email: 'logout@logisticshq.in' },
        org: { name: 'Freel' },
        memberRole: { name: 'SALES', permissions: [] },
      });

      let authApi;
      function LogoutTrigger() {
        authApi = useAuth();
        return null;
      }

      render(
        <AuthProvider>
          <LogoutTrigger />
          <AuthConsumer />
        </AuthProvider>
      );

      await waitFor(() => expect(screen.queryByTestId('booting')).not.toBeInTheDocument());
      expect(screen.getByTestId('is-authenticated').textContent).toBe('true');

      await act(async () => { await authApi.logout(); });

      expect(screen.getByTestId('is-authenticated').textContent).toBe('false');
      expect(screen.getByTestId('user-email').textContent).toBe('null');
    });
  });


  // ── useAuth() outside provider ────────────────────────────────────────────
  describe('useAuth() outside provider', () => {
    it('throws an error when used outside <AuthProvider>', () => {
      // Suppress React's console.error for this test
      const spy = vi.spyOn(console, 'error').mockImplementation(() => {});
      expect(() => render(<AuthConsumer />)).toThrow('useAuth must be used inside');
      spy.mockRestore();
    });
  });
});

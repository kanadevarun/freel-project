/**
 * Component + hook tests for RBACContext.jsx
 *
 * Tests cover:
 *  - can(module, action) — permission lookup via permissions Set
 *  - hasRole(name) — role name comparison
 *  - hasAnyRole(names) — multi-role check
 *  - Case-insensitivity
 *  - No permissions (null memberRole)
 *  - Legacy string role (no permissions)
 *  - useRBAC() throws outside provider
 */

import { describe, it, expect, vi } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import { AuthProvider } from '../../context/AuthContext';
import { RBACProvider, useRBAC } from '../../context/RBACContext';
import { authStorage } from '../../utils/authStorage';

// ── helpers ──────────────────────────────────────────────────────────────────

/**
 * Sets up a pre-authenticated session so AuthContext has a memberRole.
 */
function seedSession(roleObj) {
  authStorage.saveSessionUser({
    user: { email: 'rbac@logisticshq.in', full_name: 'RBAC User' },
    org: { name: 'Freel' },
    memberRole: roleObj,
  });
}

const SUPER_ADMIN_ROLE = {
  name: 'SUPER_ADMIN',
  display_name: 'Super Admin',
  permissions: [
    'LEADS:READ',
    'LEADS:CREATE',
    'LEADS:UPDATE',
    'LEADS:DELETE',
    'RFQS:READ',
    'COMPANIES:READ',
    'COMPANIES:CREATE',
    'OUTREACH:READ',
  ],
};

const SALES_ROLE = {
  name: 'SALES',
  display_name: 'Sales',
  permissions: ['LEADS:READ', 'LEADS:CREATE', 'RFQS:READ'],
};

// Consumer component that renders RBAC results
function RBACConsumer({ module, action, role, roles }) {
  const { can, hasRole, hasAnyRole, roleName } = useRBAC();
  return (
    <div>
      <div data-testid="can">{String(can(module, action))}</div>
      <div data-testid="has-role">{String(hasRole(role))}</div>
      <div data-testid="has-any-role">{String(hasAnyRole(roles ?? []))}</div>
      <div data-testid="role-name">{roleName ?? 'null'}</div>
    </div>
  );
}

function renderWithProviders(roleObj, consumerProps) {
  seedSession(roleObj);
  return render(
    <AuthProvider>
      <RBACProvider>
        <RBACConsumer {...consumerProps} />
      </RBACProvider>
    </AuthProvider>
  );
}

// ── tests ─────────────────────────────────────────────────────────────────────

describe('RBACContext', () => {
  // ── can() ─────────────────────────────────────────────────────────────────
  describe('can(module, action)', () => {
    it('returns true when user has the permission', async () => {
      renderWithProviders(SUPER_ADMIN_ROLE, { module: 'LEADS', action: 'READ' });
      await waitFor(() => {
        expect(screen.getByTestId('can').textContent).toBe('true');
      });
    });

    it('returns false when user does not have the permission', async () => {
      renderWithProviders(SALES_ROLE, { module: 'COMPANIES', action: 'CREATE' });
      await waitFor(() => {
        expect(screen.getByTestId('can').textContent).toBe('false');
      });
    });

    it('is case-insensitive for module and action', async () => {
      renderWithProviders(SUPER_ADMIN_ROLE, { module: 'leads', action: 'read' });
      await waitFor(() => {
        expect(screen.getByTestId('can').textContent).toBe('true');
      });
    });

    it('returns false for null module', async () => {
      renderWithProviders(SUPER_ADMIN_ROLE, { module: null, action: 'READ' });
      await waitFor(() => {
        expect(screen.getByTestId('can').textContent).toBe('false');
      });
    });

    it('returns false when memberRole is null (unauthenticated)', async () => {
      // Don't seed a session — memberRole will be null
      render(
        <AuthProvider>
          <RBACProvider>
            <RBACConsumer module="LEADS" action="READ" role="SUPER_ADMIN" roles={[]} />
          </RBACProvider>
        </AuthProvider>
      );
      await waitFor(() => {
        expect(screen.getByTestId('can').textContent).toBe('false');
      });
    });
  });

  // ── hasRole() ─────────────────────────────────────────────────────────────
  describe('hasRole(name)', () => {
    it('returns true for matching role', async () => {
      renderWithProviders(SUPER_ADMIN_ROLE, { module: 'X', action: 'Y', role: 'SUPER_ADMIN' });
      await waitFor(() => {
        expect(screen.getByTestId('has-role').textContent).toBe('true');
      });
    });

    it('returns false for non-matching role', async () => {
      renderWithProviders(SALES_ROLE, { module: 'X', action: 'Y', role: 'SUPER_ADMIN' });
      await waitFor(() => {
        expect(screen.getByTestId('has-role').textContent).toBe('false');
      });
    });

    it('is case-insensitive', async () => {
      renderWithProviders(SUPER_ADMIN_ROLE, { module: 'X', action: 'Y', role: 'super_admin' });
      await waitFor(() => {
        expect(screen.getByTestId('has-role').textContent).toBe('true');
      });
    });
  });

  // ── hasAnyRole() ──────────────────────────────────────────────────────────
  describe('hasAnyRole(names)', () => {
    it('returns true when user role is in the list', async () => {
      renderWithProviders(SALES_ROLE, {
        module: 'X', action: 'Y', role: 'X',
        roles: ['SUPER_ADMIN', 'SALES', 'PRICING'],
      });
      await waitFor(() => {
        expect(screen.getByTestId('has-any-role').textContent).toBe('true');
      });
    });

    it('returns false when user role is not in the list', async () => {
      renderWithProviders(SALES_ROLE, {
        module: 'X', action: 'Y', role: 'X',
        roles: ['SUPER_ADMIN', 'FINANCE'],
      });
      await waitFor(() => {
        expect(screen.getByTestId('has-any-role').textContent).toBe('false');
      });
    });

    it('returns false for empty list', async () => {
      renderWithProviders(SALES_ROLE, {
        module: 'X', action: 'Y', role: 'X',
        roles: [],
      });
      await waitFor(() => {
        expect(screen.getByTestId('has-any-role').textContent).toBe('false');
      });
    });
  });

  // ── roleName ─────────────────────────────────────────────────────────────
  describe('roleName', () => {
    it('exposes the uppercased role name', async () => {
      renderWithProviders(SALES_ROLE, { module: 'X', action: 'Y', role: 'X', roles: [] });
      await waitFor(() => {
        expect(screen.getByTestId('role-name').textContent).toBe('SALES');
      });
    });

    it('handles legacy string role correctly', async () => {
      // Seed a legacy plain-string role
      authStorage.saveSessionUser({
        user: { email: 'legacy@logisticshq.in' },
        org: { name: 'Freel' },
        memberRole: 'OPERATIONS', // old string format
      });
      render(
        <AuthProvider>
          <RBACProvider>
            <RBACConsumer module="X" action="Y" role="OPERATIONS" roles={[]} />
          </RBACProvider>
        </AuthProvider>
      );
      await waitFor(() => {
        expect(screen.getByTestId('role-name').textContent).toBe('OPERATIONS');
      });
    });
  });

  // ── useRBAC() outside provider ────────────────────────────────────────────
  describe('useRBAC() outside provider', () => {
    it('throws an error', () => {
      const spy = vi.spyOn(console, 'error').mockImplementation(() => {});
      expect(() =>
        render(<RBACConsumer module="X" action="Y" role="X" roles={[]} />)
      ).toThrow('useRBAC must be used inside');
      spy.mockRestore();
    });
  });
});

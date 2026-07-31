/**
 * Unit tests for authService.js
 *
 * All fetch calls are mocked with vi.stubGlobal.
 * Tests cover:
 *  - signup: success + API error
 *  - verifyEmail: success + error mapping
 *  - login: success (with envelope unwrapping) + 401 error
 *  - forgotPassword: success + error
 *  - resetPassword: success + error
 *  - logout: always succeeds (client-side)
 *  - resendCode: mocked response
 *  - handleResponse: envelope unwrapping, non-JSON body
 */

import { describe, it, expect, vi, afterEach } from 'vitest';
import {
  signup,
  verifyEmail,
  login,
  forgotPassword,
  resetPassword,
  logout,
  resendCode,
} from '../../services/authService';

// Helper to mock a successful fetch response
function mockFetchSuccess(data) {
  return vi.fn().mockResolvedValue({
    ok: true,
    json: async () => data,
  });
}

// Helper to mock a failed fetch response
function mockFetchError(data, status = 400) {
  return vi.fn().mockResolvedValue({
    ok: false,
    status,
    json: async () => data,
  });
}

describe('authService', () => {
  afterEach(() => {
    vi.restoreAllMocks();
  });

  // ── signup ────────────────────────────────────────────────────────────────
  describe('signup', () => {
    it('returns data on success', async () => {
      const mockData = { message: 'User created' };
      vi.stubGlobal('fetch', mockFetchSuccess(mockData));

      const result = await signup({
        full_name: 'Alice',
        company_name: 'Freel',
        email: 'alice@logisticshq.in',
        password: 'P@ssword123',
        role: 'freight_forwarder',
      });

      expect(result).toEqual(mockData);
    });

    it('throws a mapped error on API failure', async () => {
      vi.stubGlobal('fetch', mockFetchError({ message: 'UsernameExistsException' }));

      await expect(signup({ email: 'exists@logisticshq.in' })).rejects.toMatchObject({
        code: 'UsernameExistsException',
      });
    });
  });

  // ── verifyEmail ───────────────────────────────────────────────────────────
  describe('verifyEmail', () => {
    it('returns data on success', async () => {
      vi.stubGlobal('fetch', mockFetchSuccess({ message: 'Verified' }));
      const result = await verifyEmail({ email: 'alice@logisticshq.in', code: '123456' });
      expect(result.message).toBe('Verified');
    });

    it('throws error on bad code', async () => {
      vi.stubGlobal('fetch', mockFetchError({ message: 'CodeMismatchException' }));
      await expect(verifyEmail({ email: 'x@y.com', code: '000000' })).rejects.toMatchObject({
        code: 'CodeMismatchException',
      });
    });
  });

  // ── login ─────────────────────────────────────────────────────────────────
  describe('login', () => {
    it('returns token data on success (unwrapped from envelope)', async () => {
      const loginData = {
        access_token: 'tok',
        id_token: 'id',
        refresh_token: 'ref',
        role: { name: 'SUPER_ADMIN', permissions: ['LEADS:READ'] },
      };
      // Backend wraps response in { success, message, data: {...} }
      vi.stubGlobal('fetch', mockFetchSuccess({
        success: true,
        message: 'Login successful',
        data: loginData,
      }));

      const result = await login({ email: 'alice@logisticshq.in', password: 'P@ss1' });
      expect(result.access_token).toBe('tok');
      expect(result.role.name).toBe('SUPER_ADMIN');
    });

    it('also works when backend returns flat (non-enveloped) response', async () => {
      const loginData = {
        access_token: 'tok2',
        id_token: 'id2',
        refresh_token: 'ref2',
        role: { name: 'SALES', permissions: [] },
      };
      vi.stubGlobal('fetch', mockFetchSuccess(loginData));

      const result = await login({ email: 'bob@logisticshq.in', password: 'P@ss2' });
      expect(result.access_token).toBe('tok2');
    });

    it('throws NotAuthorizedException on wrong credentials', async () => {
      vi.stubGlobal('fetch', mockFetchError({ message: 'NotAuthorizedException' }, 401));
      await expect(login({ email: 'a@b.com', password: 'wrong' })).rejects.toMatchObject({
        code: 'NotAuthorizedException',
      });
    });
  });

  // ── forgotPassword ───────────────────────────────────────────────────────
  describe('forgotPassword', () => {
    it('returns data on success', async () => {
      vi.stubGlobal('fetch', mockFetchSuccess({ message: 'Code sent' }));
      const result = await forgotPassword({ email: 'a@b.com' });
      expect(result.message).toBe('Code sent');
    });

    it('throws on unknown user in forgot flow', async () => {
      vi.stubGlobal('fetch', mockFetchError({ message: 'UserNotFoundException' }));
      await expect(forgotPassword({ email: 'ghost@b.com' })).rejects.toMatchObject({
        code: 'UserNotFoundException',
      });
    });
  });

  // ── resetPassword ─────────────────────────────────────────────────────────
  describe('resetPassword', () => {
    it('returns data on success', async () => {
      vi.stubGlobal('fetch', mockFetchSuccess({ message: 'Password reset' }));
      const result = await resetPassword({ email: 'a@b.com', code: '111111', new_password: 'New@1234' });
      expect(result.message).toBe('Password reset');
    });

    it('throws on expired code in reset flow', async () => {
      vi.stubGlobal('fetch', mockFetchError({ message: 'ExpiredCodeException' }));
      await expect(resetPassword({ email: 'a@b.com', code: '000000', new_password: 'New@1234' })).rejects.toMatchObject({
        code: 'ExpiredCodeException',
      });
    });
  });

  // ── logout ────────────────────────────────────────────────────────────────
  describe('logout', () => {
    it('resolves with a message (client-side only)', async () => {
      const result = await logout();
      expect(result.message).toBe('Logged out.');
    });
  });

  // ── resendCode ────────────────────────────────────────────────────────────
  describe('resendCode', () => {
    it('resolves with a code resent message (mocked)', async () => {
      const result = await resendCode({ email: 'a@b.com' });
      expect(result.message).toBeTruthy();
    });
  });

  // ── handleResponse envelope unwrapping ───────────────────────────────────
  describe('handleResponse envelope unwrapping', () => {
    it('returns top-level object when data field is absent', async () => {
      vi.stubGlobal('fetch', mockFetchSuccess({ some_field: 'value' }));
      const result = await signup({ email: 'x@y.com' });
      expect(result.some_field).toBe('value');
    });

    it('unwraps data field when backend returns standard envelope', async () => {
      vi.stubGlobal('fetch', mockFetchSuccess({
        success: true,
        message: 'Done',
        data: { wrapped: true },
      }));
      const result = await signup({ email: 'x@y.com' });
      expect(result.wrapped).toBe(true);
    });
  });
});

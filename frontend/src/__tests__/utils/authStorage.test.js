/**
 * Unit tests for authStorage.js
 *
 * Tests cover:
 *  - saveTokens / getTokens / clearTokens
 *  - saveSessionUser / getSessionUser / clearSessionUser
 *  - clearAll
 *  - Edge cases (missing keys, corrupt JSON)
 */

import { describe, it, expect, beforeEach } from 'vitest';
import { authStorage } from '../../utils/authStorage';

// localStorage is reset in setup.js before each test

describe('authStorage', () => {
  const mockTokens = {
    access_token: 'test-access-token',
    id_token: 'test-id-token',
    refresh_token: 'test-refresh-token',
  };

  const mockSessionData = {
    user: { email: 'test@logisticshq.in', full_name: 'Test User' },
    org: { name: 'Freel Corp', orgType: 'SHIPPER' },
    memberRole: { name: 'SUPER_ADMIN', display_name: 'Super Admin', permissions: ['LEADS:READ'] },
  };

  // ── saveTokens / getTokens ────────────────────────────────────────────────
  describe('saveTokens / getTokens', () => {
    it('saves and retrieves all three tokens', () => {
      authStorage.saveTokens(mockTokens);
      const tokens = authStorage.getTokens();
      expect(tokens.accessToken).toBe(mockTokens.access_token);
      expect(tokens.idToken).toBe(mockTokens.id_token);
      expect(tokens.refreshToken).toBe(mockTokens.refresh_token);
    });

    it('skips saving undefined token fields', () => {
      authStorage.saveTokens({ access_token: 'only-access' });
      const tokens = authStorage.getTokens();
      expect(tokens.accessToken).toBe('only-access');
      expect(tokens.idToken).toBeNull();
      expect(tokens.refreshToken).toBeNull();
    });

    it('returns null for all fields when nothing is saved', () => {
      const tokens = authStorage.getTokens();
      expect(tokens.accessToken).toBeNull();
      expect(tokens.idToken).toBeNull();
      expect(tokens.refreshToken).toBeNull();
    });
  });

  // ── clearTokens ──────────────────────────────────────────────────────────
  describe('clearTokens', () => {
    it('removes all three token keys from localStorage', () => {
      authStorage.saveTokens(mockTokens);
      authStorage.clearTokens();
      const tokens = authStorage.getTokens();
      expect(tokens.accessToken).toBeNull();
      expect(tokens.idToken).toBeNull();
      expect(tokens.refreshToken).toBeNull();
    });
  });

  // ── saveSessionUser / getSessionUser ─────────────────────────────────────
  describe('saveSessionUser / getSessionUser', () => {
    it('saves and retrieves session user data', () => {
      authStorage.saveSessionUser(mockSessionData);
      const retrieved = authStorage.getSessionUser();
      expect(retrieved).toEqual(mockSessionData);
    });

    it('returns null when no session data exists', () => {
      expect(authStorage.getSessionUser()).toBeNull();
    });

    it('returns null when localStorage contains corrupt JSON', () => {
      localStorage.setItem('freel_session_user', 'not-valid-json{{{');
      expect(authStorage.getSessionUser()).toBeNull();
    });
  });

  // ── clearSessionUser ─────────────────────────────────────────────────────
  describe('clearSessionUser', () => {
    it('removes the session user from localStorage', () => {
      authStorage.saveSessionUser(mockSessionData);
      authStorage.clearSessionUser();
      expect(authStorage.getSessionUser()).toBeNull();
    });
  });

  // ── clearAll ──────────────────────────────────────────────────────────────
  describe('clearAll', () => {
    it('removes all auth-related data from localStorage', () => {
      authStorage.saveTokens(mockTokens);
      authStorage.saveSessionUser(mockSessionData);
      authStorage.clearAll();

      const tokens = authStorage.getTokens();
      expect(tokens.accessToken).toBeNull();
      expect(authStorage.getSessionUser()).toBeNull();
    });
  });
});

/**
 * Unit tests for authErrorMapper.js
 *
 * Tests cover:
 *  - Null / undefined / empty inputs
 *  - Each known Cognito exception code mapping
 *  - Flow-specific message variants (signup, login, verify, forgot, reset)
 *  - Generic fallback messages per flow
 *  - The "already exists" heuristic fallback
 */

import { describe, it, expect } from 'vitest';
import { mapAuthError } from '../../utils/authErrorMapper';

describe('mapAuthError', () => {
  // ── NULL / EMPTY INPUTS ──────────────────────────────────────────────────
  describe('null / empty inputs', () => {
    it('returns UNKNOWN_ERROR for null input', () => {
      const result = mapAuthError(null);
      expect(result.code).toBe('UNKNOWN_ERROR');
      expect(result.message).toBeTruthy();
    });

    it('returns UNKNOWN_ERROR for undefined input', () => {
      const result = mapAuthError(undefined);
      expect(result.code).toBe('UNKNOWN_ERROR');
    });

    it('returns UNKNOWN_ERROR for empty string', () => {
      const result = mapAuthError('');
      expect(result.code).toBe('UNKNOWN_ERROR');
    });
  });

  // ── KNOWN EXCEPTION MAPPINGS ─────────────────────────────────────────────
  describe('UsernameExistsException', () => {
    it('maps correctly', () => {
      const result = mapAuthError('UsernameExistsException: User already exists');
      expect(result.code).toBe('UsernameExistsException');
      expect(result.message).toContain('already registered');
      expect(result.action).toBe('LOGIN');
    });

    it('maps via heuristic "already exists" text', () => {
      const result = mapAuthError('An account with this email already exists');
      expect(result.code).toBe('UsernameExistsException');
    });
  });

  describe('InvalidPasswordException', () => {
    it('returns signup variant message', () => {
      const result = mapAuthError('InvalidPasswordException: too short', 'signup');
      expect(result.code).toBe('InvalidPasswordException');
      expect(result.message).toContain('criteria');
    });

    it('returns reset variant message', () => {
      const result = mapAuthError('InvalidPasswordException: too short', 'reset');
      expect(result.code).toBe('InvalidPasswordException');
      expect(result.message).toContain("new password");
    });
  });

  describe('CodeMismatchException', () => {
    it('returns verify variant message for verify flow', () => {
      const result = mapAuthError('CodeMismatchException', 'verify');
      expect(result.message).toContain('verification code');
    });

    it('returns reset variant message for reset flow', () => {
      const result = mapAuthError('CodeMismatchException', 'reset');
      expect(result.message).toContain('reset code');
    });
  });

  describe('ExpiredCodeException', () => {
    it('returns expired code message for verify flow', () => {
      const result = mapAuthError('ExpiredCodeException', 'verify');
      expect(result.message).toContain('expired');
    });

    it('returns expired code message for reset flow', () => {
      const result = mapAuthError('ExpiredCodeException', 'reset');
      expect(result.message).toContain('expired');
    });
  });

  describe('UserNotFoundException', () => {
    it('returns SIGNUP action for verify flow', () => {
      const result = mapAuthError('UserNotFoundException', 'verify');
      expect(result.action).toBe('SIGNUP');
    });

    it('returns login-safe (non-leaking) message for login flow', () => {
      const result = mapAuthError('UserNotFoundException', 'login');
      expect(result.message).toContain('Incorrect email or password');
    });

    it('returns no-account message for forgot flow', () => {
      const result = mapAuthError('UserNotFoundException', 'forgot');
      expect(result.message).toContain('No account found');
    });
  });

  describe('NotAuthorizedException', () => {
    it('returns "Incorrect email or password" message', () => {
      const result = mapAuthError('NotAuthorizedException');
      expect(result.message).toContain('Incorrect email or password');
    });
  });

  describe('UserNotConfirmedException', () => {
    it('returns VERIFY action', () => {
      const result = mapAuthError('UserNotConfirmedException');
      expect(result.action).toBe('VERIFY');
    });
  });

  // ── GENERIC FLOW FALLBACKS ────────────────────────────────────────────────
  describe('generic fallbacks by flow', () => {
    it('returns GENERIC_SIGNUP_FAILURE for signup flow with unknown error', () => {
      const result = mapAuthError('Some unknown error', 'signup');
      expect(result.code).toBe('GENERIC_SIGNUP_FAILURE');
    });

    it('returns GENERIC_FORGOT_FAILURE for forgot flow with unknown error', () => {
      const result = mapAuthError('Some unknown error', 'forgot');
      expect(result.code).toBe('GENERIC_FORGOT_FAILURE');
    });

    it('returns GENERIC_RESET_FAILURE for reset flow with unknown error', () => {
      const result = mapAuthError('Some unknown error', 'reset');
      expect(result.code).toBe('GENERIC_RESET_FAILURE');
    });

    it('returns AUTH_ERROR for default flow with unknown error', () => {
      const result = mapAuthError('Some unknown error');
      expect(result.code).toBe('AUTH_ERROR');
    });
  });

  // ── STRUCTURE VALIDATION ──────────────────────────────────────────────────
  describe('return shape', () => {
    it('always returns an object with code, message, and rawMessage fields', () => {
      const result = mapAuthError('UsernameExistsException');
      expect(result).toHaveProperty('code');
      expect(result).toHaveProperty('message');
      expect(result).toHaveProperty('rawMessage');
    });
  });
});

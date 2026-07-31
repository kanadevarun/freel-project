/**
 * Unit tests for onboardingStorage.js
 *
 * Tests cover:
 *  - getOnboardingState with valid, invalid, and no data
 *  - saveOnboardingState with email + answers
 *  - clearOnboardingState
 *  - Edge cases: no email, wrong email, corrupt JSON
 */

import { describe, it, expect } from 'vitest';
import { onboardingStorage } from '../../utils/onboardingStorage';

describe('onboardingStorage', () => {
  const TEST_EMAIL = 'alice@logisticshq.in';
  const OTHER_EMAIL = 'bob@logisticshq.in';
  const TEST_ANSWERS = { company_type: 'SHIPPER', size: 'MEDIUM' };

  // ── getOnboardingState ────────────────────────────────────────────────────
  describe('getOnboardingState', () => {
    it('returns null when nothing has been saved', () => {
      expect(onboardingStorage.getOnboardingState(TEST_EMAIL)).toBeNull();
    });

    it('returns null when email is null or undefined', () => {
      expect(onboardingStorage.getOnboardingState(null)).toBeNull();
      expect(onboardingStorage.getOnboardingState(undefined)).toBeNull();
    });

    it('returns null when email does not match stored email', () => {
      onboardingStorage.saveOnboardingState(TEST_EMAIL, TEST_ANSWERS);
      expect(onboardingStorage.getOnboardingState(OTHER_EMAIL)).toBeNull();
    });

    it('returns the state when email matches and completed is true', () => {
      onboardingStorage.saveOnboardingState(TEST_EMAIL, TEST_ANSWERS);
      const state = onboardingStorage.getOnboardingState(TEST_EMAIL);
      expect(state).not.toBeNull();
      expect(state.email).toBe(TEST_EMAIL);
      expect(state.completed).toBe(true);
      expect(state.answers).toEqual(TEST_ANSWERS);
    });

    it('returns null when localStorage contains corrupt JSON', () => {
      localStorage.setItem('freel_onboarding_state', 'bad-json{{');
      expect(onboardingStorage.getOnboardingState(TEST_EMAIL)).toBeNull();
    });

    it('returns null when stored completed is not true', () => {
      localStorage.setItem(
        'freel_onboarding_state',
        JSON.stringify({ email: TEST_EMAIL, completed: false })
      );
      expect(onboardingStorage.getOnboardingState(TEST_EMAIL)).toBeNull();
    });
  });

  // ── saveOnboardingState ──────────────────────────────────────────────────
  describe('saveOnboardingState', () => {
    it('saves completed state with email and answers', () => {
      onboardingStorage.saveOnboardingState(TEST_EMAIL, TEST_ANSWERS);
      const raw = JSON.parse(localStorage.getItem('freel_onboarding_state'));
      expect(raw.email).toBe(TEST_EMAIL);
      expect(raw.completed).toBe(true);
      expect(raw.answers).toEqual(TEST_ANSWERS);
      expect(raw.completed_at).toBeTruthy();
    });

    it('saves with empty answers when none provided', () => {
      onboardingStorage.saveOnboardingState(TEST_EMAIL);
      const raw = JSON.parse(localStorage.getItem('freel_onboarding_state'));
      expect(raw.answers).toEqual({});
    });

    it('does nothing when email is falsy', () => {
      onboardingStorage.saveOnboardingState(null, TEST_ANSWERS);
      expect(localStorage.getItem('freel_onboarding_state')).toBeNull();
    });

    it('overwrites previous state for same email', () => {
      onboardingStorage.saveOnboardingState(TEST_EMAIL, { step: 1 });
      onboardingStorage.saveOnboardingState(TEST_EMAIL, { step: 2 });
      const state = onboardingStorage.getOnboardingState(TEST_EMAIL);
      expect(state.answers.step).toBe(2);
    });
  });

  // ── clearOnboardingState ─────────────────────────────────────────────────
  describe('clearOnboardingState', () => {
    it('removes the onboarding state key', () => {
      onboardingStorage.saveOnboardingState(TEST_EMAIL, TEST_ANSWERS);
      onboardingStorage.clearOnboardingState();
      expect(localStorage.getItem('freel_onboarding_state')).toBeNull();
    });
  });
});

/**
 * onboardingStorage.js - Centralized management of user-scoped onboarding completion state.
 */

const ONBOARDING_STATE_KEY = 'freel_onboarding_state';

export const onboardingStorage = {
  getOnboardingState: (email) => {
    if (!email) return null;
    try {
      const data = localStorage.getItem(ONBOARDING_STATE_KEY);
      if (!data) return null;
      
      const parsed = JSON.parse(data);
      // Validate that the stored state belongs to the current user
      if (parsed.email === email && parsed.completed === true) {
        return parsed;
      }
      return null;
    } catch {
      return null;
    }
  },

  saveOnboardingState: (email, answers = {}) => {
    if (!email) return;
    const stateObj = {
      email,
      completed: true,
      completed_at: new Date().toISOString(),
      answers
    };
    localStorage.setItem(ONBOARDING_STATE_KEY, JSON.stringify(stateObj));
  },

  clearOnboardingState: () => {
    localStorage.removeItem(ONBOARDING_STATE_KEY);
  }
};

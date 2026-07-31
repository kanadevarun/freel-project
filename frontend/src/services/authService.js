/**
 * authService.js — All API calls for authentication.
 *
 * PHASE 2B: Real API integration against Go backend.
 *
 * Auth-flow endpoints (signup, login, verify, forgot, reset) use raw fetch
 * because they execute BEFORE the user has a JWT. All other calls that need
 * an authenticated session should use the `api` client from `./api.js`.
 */

import { mapAuthError } from '../utils/authErrorMapper';

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080';

// ── Internal helper ───────────────────────────────────────────────────────────

/**
 * handleResponse — Parse + standardize auth endpoint responses.
 * Auth endpoints are public (no JWT) so they use raw fetch independently.
 *
 * @param {Response} response
 * @param {string} [flow] — 'signup' | 'login' | 'verify' | 'forgot' | 'reset'
 */
async function handleResponse(response, flow = 'default') {
  let data;
  try {
    data = await response.json();
  } catch {
    // Non-JSON or empty body
    if (!response.ok) {
      throw { code: 'NETWORK_ERROR', message: 'An unexpected error occurred. Please try again.' };
    }
    return null;
  }

  if (!response.ok) {
    const rawMessage = data?.message || data?.error || 'An error occurred processing your request.';
    throw mapAuthError(rawMessage, flow);
  }

  // Unwrap standard backend envelope { success, message, data }
  return data?.data !== undefined ? data.data : data;
}

// ── SIGNUP ────────────────────────────────────────────────────────────────────

/**
 * signup — Create a new user + organization account.
 * @param {{ role, full_name, company_name, email, password }} payload
 */
export async function signup(payload) {
  const response = await fetch(`${API_BASE_URL}/auth/signup`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload),
  });
  return handleResponse(response, 'signup');
}

// ── VERIFY EMAIL ──────────────────────────────────────────────────────────────

/**
 * verifyEmail — Verify the 6-digit OTP sent to email.
 * @param {{ email, code }} payload
 */
export async function verifyEmail(payload) {
  const response = await fetch(`${API_BASE_URL}/auth/verify-email`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload),
  });
  return handleResponse(response, 'verify');
}

/**
 * resendCode — Resend the verification OTP.
 * TODO: Implement once backend exposes POST /auth/resend-code
 * @param {{ email }} data
 */
export async function resendCode(data) {
  // Mocked until backend implements this endpoint
  console.log('Resend code requested for', data.email);
  return { message: 'Code resent.' };
}

// ── LOGIN ─────────────────────────────────────────────────────────────────────

/**
 * login — Authenticate with email + password.
 * Returns: { access_token, id_token, refresh_token, expires_in, role: { name, display_name, permissions[] } }
 * @param {{ email, password }} payload
 */
export async function login(payload) {
  const response = await fetch(`${API_BASE_URL}/auth/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload),
  });
  return handleResponse(response, 'login');
}

// ── FORGOT PASSWORD ───────────────────────────────────────────────────────────

/**
 * forgotPassword — Request a password reset OTP.
 * @param {{ email }} payload
 */
export async function forgotPassword(payload) {
  const response = await fetch(`${API_BASE_URL}/auth/forgot-password`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload),
  });
  return handleResponse(response, 'forgot');
}

/**
 * resetPassword — Confirm OTP and set a new password.
 * @param {{ email, code, new_password }} payload
 */
export async function resetPassword(payload) {
  const response = await fetch(`${API_BASE_URL}/auth/reset-password`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload),
  });
  return handleResponse(response, 'reset');
}

// ── LOGOUT ────────────────────────────────────────────────────────────────────

/**
 * logout — Session invalidation is handled client-side via AuthContext.clearAll().
 * A server-side token revocation endpoint can be added here in a future sprint.
 */
export async function logout() {
  return { message: 'Logged out.' };
}

// ── ONBOARDING MOCKS ─────────────────────────────────────────────────────────
// Kept for backward compatibility with OnboardingPage until real backend is wired.

export async function saveOnboardingStep(data) {
  return { success: true, nextStep: data.step + 1 };
}

export async function completeOnboarding() {
  return { message: 'Onboarding complete.' };
}

export async function skipOnboarding() {
  return { message: 'Onboarding skipped.' };
}

export async function createWorkspace(data) {
  return { success: true, slug: data.slug };
}

export async function inviteTeamMembers(invitations) {
  return { success: true, count: invitations.length };
}

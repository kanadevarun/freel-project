/**
 * authService.js — All API calls for authentication.
 *
 * PHASE 2B (NOW): Real API integration against Go backend.
 */

import { mapAuthError } from '../utils/authErrorMapper';

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080';

/**
 * Helper to handle fetch responses and standardize errors
 */
async function handleResponse(response, flow = 'default') {
  let data;
  try {
    data = await response.json();
  } catch (err) {
    // If not JSON or empty body
    if (!response.ok) {
      throw { code: 'NETWORK_ERROR', message: 'An unexpected error occurred. Please try again.' };
    }
    return null;
  }

  if (!response.ok) {
    // The backend might return errors in different formats (e.g., error string or message field)
    const rawMessage = data?.message || data?.error || 'An error occurred processing your request.';
    throw mapAuthError(rawMessage, flow);
  }

  return data;
}

// ── SIGNUP ──────────────────────────────────────────────
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

// ── VERIFY EMAIL ────────────────────────────────────────
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
 * (TODO: Wait for backend endpoint)
 * @param {{ email }} data
 */
export async function resendCode(data) {
  // Mock until backend implements this
  console.log('Resend code requested for', data.email);
  return { message: 'Code resent.' };
}

// ── LOGIN ────────────────────────────────────────────────
/**
 * login — Authenticate with email + password.
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

// ── FORGOT PASSWORD ───────────────────────────────────────
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

// ── LOGOUT ───────────────────────────────────────────────
/**
 * logout — Invalidate the session server-side.
 * Currently handled purely client-side via AuthContext.
 */
export async function logout() {
  return { message: 'Logged out.' };
}

// ── ONBOARDING MOCKS (For backward compatibility until real backend) ──
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

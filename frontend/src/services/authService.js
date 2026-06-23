/**
 * authService.js — All API calls for authentication.
 *
 * PHASE 2A (NOW): Mock functions with fake 800ms delays.
 *   → Pages work, forms submit, errors can be tested — no backend needed.
 *
 * PHASE 2B (LATER): Replace each function body with real axios call.
 *   → Every page/component stays exactly the same.
 *   → Only this file changes.
 *
 * Example swap for register():
 *   Phase 2A: await fakeDelay(800); return { success: true, ... }
 *   Phase 2B: const res = await axios.post('/api/v1/auth/register', data);
 *             return res.data;
 */

const fakeDelay = (ms) => new Promise((resolve) => setTimeout(resolve, ms));

// ── SIGNUP ──────────────────────────────────────────────
/**
 * register — Create a new user + organization account.
 * @param {{ orgType, fullName, companyName, email, password }} data
 * @returns {{ userId, orgId, message }}
 */
export async function register(data) {
  await fakeDelay(1000);

  // Simulate "email already taken" for testing error states
  if (data.email === 'taken@test.com') {
    throw { code: 'EMAIL_TAKEN', message: 'An account with this email already exists.' };
  }

  return {
    userId: 'mock-user-' + Date.now(),
    orgId: 'mock-org-' + Date.now(),
    message: 'Verification code sent to your email.',
  };
}

/**
 * confirmEmail — Verify the 6-digit OTP sent to email.
 * @param {{ email, code }} data
 * @returns {{ user, org, memberRole }}
 */
export async function confirmEmail(data) {
  await fakeDelay(800);

  // Simulate wrong OTP for testing
  if (data.code === '000000') {
    throw { code: 'INVALID_CODE', message: 'Invalid code. Please check and try again.' };
  }

  // Mock successful verification — returns full session data
  return {
    user: {
      id: 'mock-user-001',
      email: data.email,
      fullName: data.fullName || 'User',
      avatarUrl: null,
      emailVerified: true,
    },
    org: {
      id: 'mock-org-001',
      name: data.companyName || data.fullName + "'s Company",
      orgType: data.orgType || 'SHIPPER',
      plan: 'FREE',
      isVerified: false,
      onboardingCompleted: false,
    },
    memberRole: 'OWNER',
  };
}

/**
 * resendCode — Resend the verification OTP.
 * @param {{ email }} data
 */
export async function resendCode(data) {
  await fakeDelay(600);
  console.log('[Mock] Resent verification code to', data.email);
  return { message: 'Code resent.' };
}

// ── LOGIN ────────────────────────────────────────────────
/**
 * login — Authenticate with email + password.
 * @param {{ email, password }} data
 * @returns {{ user, org, memberRole }} on success
 */
export async function login(data) {
  await fakeDelay(900);

  // Simulate wrong password
  if (data.password === 'wrong') {
    throw { code: 'INCORRECT_CREDENTIALS', message: 'Incorrect email or password.' };
  }

  // Simulate unverified account
  if (data.email === 'unverified@test.com') {
    throw { code: 'EMAIL_NOT_VERIFIED', message: 'Please verify your email first.' };
  }

  return {
    user: {
      id: 'mock-user-001',
      email: data.email,
      fullName: 'Rajesh Kumar',
      avatarUrl: null,
      emailVerified: true,
    },
    org: {
      id: 'mock-org-001',
      name: 'Kumar Transport Pvt Ltd',
      orgType: 'CARRIER',
      plan: 'FREE',
      isVerified: false,
      onboardingCompleted: true,
    },
    memberRole: 'OWNER',
  };
}

// ── FORGOT PASSWORD ───────────────────────────────────────
/**
 * forgotPassword — Request a password reset OTP.
 * Always returns success (never reveal if email exists — security).
 * @param {{ email }} data
 */
export async function forgotPassword(data) {
  await fakeDelay(700);
  console.log('[Mock] Reset OTP sent to', data.email);
  return { message: 'If that email is registered, you will receive a reset code.' };
}

/**
 * resetPassword — Confirm OTP and set a new password.
 * @param {{ email, code, newPassword }} data
 */
export async function resetPassword(data) {
  await fakeDelay(800);

  if (data.code === '000000') {
    throw { code: 'INVALID_CODE', message: 'Invalid or expired reset code.' };
  }

  return { message: 'Password reset successfully. Please log in.' };
}

// ── LOGOUT ───────────────────────────────────────────────
/**
 * logout — Invalidate the session server-side.
 * In Phase 2A: nothing to call (localStorage cleared by AuthContext).
 * In Phase 2B: POST /api/v1/auth/logout to clear httpOnly cookies.
 */
export async function logout() {
  await fakeDelay(300);
  return { message: 'Logged out.' };
}

// ── ONBOARDING ───────────────────────────────────────────
/**
 * saveOnboardingStep — Save answers for one onboarding step.
 * @param {{ orgId, step, questionKey, answer }} data
 */
export async function saveOnboardingStep(data) {
  await fakeDelay(500);
  console.log('[Mock] Saved onboarding step', data.step, '→', data.questionKey, data.answer);
  return { success: true, nextStep: data.step + 1 };
}

/**
 * completeOnboarding — Mark onboarding as done.
 */
export async function completeOnboarding() {
  await fakeDelay(500);
  return { message: 'Onboarding complete.' };
}

/**
 * skipOnboarding — Skip onboarding, go straight to dashboard.
 */
export async function skipOnboarding() {
  await fakeDelay(300);
  return { message: 'Onboarding skipped.' };
}

// ── WORKSPACE SETUP ──────────────────────────────────────
/**
 * createWorkspace — Simulate creating the organization's workspace
 * and setting up initial data (slug, org name).
 * @param {{ orgId, workspaceName, slug }} data
 */
export async function createWorkspace(data) {
  await fakeDelay(1500); // slightly longer to simulate heavy lifting
  console.log('[Mock] Created workspace:', data);
  return { success: true, slug: data.slug };
}

/**
 * inviteTeamMembers — Send invitations to team members.
 * @param {Array<{ email, role }>} invitations
 */
export async function inviteTeamMembers(invitations) {
  await fakeDelay(800);
  console.log('[Mock] Invited members:', invitations);
  return { success: true, count: invitations.length };
}

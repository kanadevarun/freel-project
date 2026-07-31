import { createContext, useContext, useState, useEffect } from 'react';
import { authStorage } from '../utils/authStorage';
import { onboardingStorage } from '../utils/onboardingStorage';

/**
 * AuthContext — Global authentication state for the entire Freel app.
 *
 * Phase 2B: Uses real backend API integration and authStorage.
 *
 * Provides:
 *   user        — the logged-in person (name, email, full_name)
 *   org         — the company they belong to (name, org_type, plan)
 *   memberRole  — the full role object from backend:
 *                 { name: string, display_name: string, permissions: string[] }
 *                 e.g. { name: "SUPER_ADMIN", display_name: "Super Admin",
 *                        permissions: ["LEADS:READ", "LEADS:CREATE", ...] }
 *                 For backward-compat, may also be a plain string role name
 *                 in sessions created before this format was introduced.
 *   isAuthenticated — boolean
 *   isBooting   — true on initial page load while checking localStorage
 *   onboardingCompleted - boolean indicating if user has finished setup
 */

const AuthContext = createContext(null);

export function AuthProvider({ children }) {
  const [user, setUser] = useState(null);
  const [org, setOrg] = useState(null);
  const [memberRole, setMemberRole] = useState(null);
  const [isBooting, setIsBooting] = useState(true);
  const [onboardingCompleted, setOnboardingCompleted] = useState(false);

  // Cross-tab sync and initial boot
  const checkSession = () => {
    try {
      const sessionUser = authStorage.getSessionUser();
      if (sessionUser && sessionUser.user && sessionUser.user.email) {
        setUser(sessionUser.user);
        setOrg(sessionUser.org);
        setMemberRole(sessionUser.memberRole);
        
        // Check onboarding status tied to this email
        const obState = onboardingStorage.getOnboardingState(sessionUser.user.email);
        setOnboardingCompleted(!!(obState && obState.completed));
      } else {
        // No valid session
        setUser(null);
        setOrg(null);
        setMemberRole(null);
        setOnboardingCompleted(false);
      }
    } catch {
      setUser(null);
      setOrg(null);
      setMemberRole(null);
      setOnboardingCompleted(false);
    } finally {
      setIsBooting(false);
    }
  };

  useEffect(() => {
    // 1. Initial boot check
    checkSession();

    // 2. Listen for cross-tab storage changes
    const handleStorageChange = (e) => {
      // Re-evaluate session if relevant keys change in other tabs
      if (
        e.key === 'freel_session_user' || 
        e.key === 'freel_access_token' || 
        e.key === 'freel_onboarding_state'
      ) {
        checkSession();
      }
      // If storage was completely cleared (e.key === null)
      if (e.key === null) {
        checkSession();
      }
    };

    window.addEventListener('storage', handleStorageChange);
    return () => window.removeEventListener('storage', handleStorageChange);
  }, []);

  /**
   * login — called after successful auth.
   *
   * @param {object} tokenData  — { access_token, id_token, refresh_token, expires_in }
   * @param {object} userData   — { email, full_name, ... } (optional, falls back to tokenData.email)
   * @param {object} orgData    — { name, org_type, plan } (optional)
   * @param {object|string} role — Full role object { name, display_name, permissions[] } from backend
   *                               LoginResponseData.role, or a plain string for backward compat.
   */
  async function login(tokenData, userData, orgData, role) {
    // Save tokens in localStorage
    authStorage.saveTokens({
      access_token: tokenData.access_token,
      id_token: tokenData.id_token,
      refresh_token: tokenData.refresh_token,
    });

    const savedUser = userData || { email: tokenData.email || 'user' };
    const savedOrg = orgData || { name: 'Organization', orgType: 'SHIPPER' };

    // Normalize role: prefer full object; fall back to plain string for legacy compatibility
    const savedRole = role && typeof role === 'object'
      ? role
      : (role ? { name: role, display_name: role, permissions: [] } : { name: 'OWNER', display_name: 'Owner', permissions: [] });

    const sessionData = {
      user: savedUser,
      org: savedOrg,
      memberRole: savedRole,
    };

    setUser(savedUser);
    setOrg(savedOrg);
    setMemberRole(savedRole);
    authStorage.saveSessionUser(sessionData);

    // Check if onboarding is already completed for this newly logged-in user
    const obState = onboardingStorage.getOnboardingState(savedUser.email);
    setOnboardingCompleted(!!(obState && obState.completed));
  }

  /**
   * logout — clears all auth state and tokens.
   */
  async function logout() {
    setUser(null);
    setOrg(null);
    setMemberRole(null);
    setOnboardingCompleted(false);
    authStorage.clearAll();
  }

  /**
   * completeOnboarding — called when onboarding finishes successfully
   */
  async function completeOnboarding(answers = {}) {
    if (!user || !user.email) return;
    onboardingStorage.saveOnboardingState(user.email, answers);
    setOnboardingCompleted(true);
  }

  /**
   * updateOrg — called after onboarding completes or when org profile updates
   */
  function updateOrg(updates) {
    const updated = { ...org, ...updates };
    setOrg(updated);
    
    const sessionUser = authStorage.getSessionUser() || {};
    authStorage.saveSessionUser({ ...sessionUser, org: updated });
  }

  const isAuthenticated = !!user;

  return (
    <AuthContext.Provider value={{
      user,
      org,
      memberRole,
      isAuthenticated,
      isBooting,
      onboardingCompleted,
      login,
      logout,
      updateOrg,
      completeOnboarding,
    }}>
      {children}
    </AuthContext.Provider>
  );
}

// eslint-disable-next-line react-refresh/only-export-components
export function useAuth() {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error('useAuth must be used inside <AuthProvider>');
  return ctx;
}

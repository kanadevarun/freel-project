import { createContext, useContext, useState, useEffect } from 'react';
import { authStorage } from '../utils/authStorage';

/**
 * AuthContext — Global authentication state for the entire LogisticsHQ app.
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
 */

const AuthContext = createContext(null);

export function AuthProvider({ children }) {
  const [user, setUser] = useState(null);
  const [org, setOrg] = useState(null);
  const [memberRole, setMemberRole] = useState(null);
  const [isBooting, setIsBooting] = useState(true);

  // Cross-tab sync and initial boot
  const checkSession = () => {
    try {
      const sessionUser = authStorage.getSessionUser();
      if (sessionUser && sessionUser.user && sessionUser.user.email) {
        setUser(sessionUser.user);
        setOrg(sessionUser.org);
        setMemberRole(sessionUser.memberRole);
      } else {
        // No valid session
        setUser(null);
        setOrg(null);
        setMemberRole(null);
      }
    } catch {
      setUser(null);
      setOrg(null);
      setMemberRole(null);
    } finally {
      setIsBooting(false);
    }
  };

  const hydrateSession = async () => {
    try {
      const { accessToken } = authStorage.getTokens();
      if (!accessToken) return;

      const API_BASE = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080';
      const res = await fetch(`${API_BASE}/auth/me`, {
        headers: { Authorization: `Bearer ${accessToken}` },
      });
      if (res.ok) {
        const payload = await res.json();
        const meData = payload?.data || payload;
        if (meData?.user) {
          setUser((prev) => ({ ...(prev || {}), ...meData.user }));
          if (meData?.org) setOrg((prev) => ({ ...(prev || {}), ...meData.org }));
          if (meData?.role) setMemberRole(meData.role);

          const curSession = authStorage.getSessionUser() || {};
          authStorage.saveSessionUser({
            ...curSession,
            user: { ...(curSession.user || {}), ...meData.user },
            org: { ...(curSession.org || {}), ...(meData.org || {}) },
            memberRole: meData.role || curSession.memberRole,
          });
        }
      }
    } catch {
      // Non-blocking background hydration
    }
  };

  useEffect(() => {
    // 1. Initial boot check & background hydration
    checkSession();
    hydrateSession();

    // 2. Listen for cross-tab storage changes
    const handleStorageChange = (e) => {
      // Re-evaluate session if relevant keys change in other tabs
      if (
        e.key === 'freel_session_user' || 
        e.key === 'freel_access_token'
      ) {
        checkSession();
        hydrateSession();
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

    const savedUser = userData || tokenData.user || { email: tokenData.email || 'user' };
    const savedOrg = orgData || tokenData.org || { name: 'Your Workspace', orgType: 'FREIGHT_FORWARDER' };

    // Normalize role: prefer full object; fall back to plain string for legacy compatibility
    const resolvedRole = role || tokenData.role;
    const savedRole = resolvedRole && typeof resolvedRole === 'object'
      ? resolvedRole
      : (resolvedRole ? { name: resolvedRole, display_name: resolvedRole, permissions: [] } : { name: 'SUPER_ADMIN', display_name: 'Super Admin', permissions: [] });

    const sessionData = {
      user: savedUser,
      org: savedOrg,
      memberRole: savedRole,
    };

    setUser(savedUser);
    setOrg(savedOrg);
    setMemberRole(savedRole);
    authStorage.saveSessionUser(sessionData);
  }

  /**
   * logout — clears all auth state and tokens.
   */
  async function logout() {
    setUser(null);
    setOrg(null);
    setMemberRole(null);
    authStorage.clearAll();
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
      login,
      logout,
      updateOrg,
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

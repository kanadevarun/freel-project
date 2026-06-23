import { createContext, useContext, useState, useEffect } from 'react';

/**
 * AuthContext — Global authentication state for the entire Freel app.
 *
 * Phase 2A: Uses mock data + localStorage.
 * Phase 2B: Swap authService.js for real API calls — this file stays the same.
 *
 * Provides:
 *   user        — the logged-in person (name, email)
 *   org         — the company they belong to (name, org_type, plan)
 *   memberRole  — their permission level (OWNER, ADMIN, FINANCE, etc.)
 *   isAuthenticated — boolean
 *   isLoading   — true on initial page load while checking localStorage
 */

const AuthContext = createContext(null);

export function AuthProvider({ children }) {
  const [user, setUser] = useState(null);
  const [org, setOrg] = useState(null);
  const [memberRole, setMemberRole] = useState(null);
  const [isLoading, setIsLoading] = useState(true);

  // On mount: check if user was already logged in (from localStorage)
  useEffect(() => {
    try {
      const savedSession = localStorage.getItem('freel_session');
      if (savedSession) {
        const { user, org, memberRole } = JSON.parse(savedSession);
        setUser(user);
        setOrg(org);
        setMemberRole(memberRole);
      }
    } catch {
      localStorage.removeItem('freel_session');
    } finally {
      setIsLoading(false);
    }
  }, []);

  /**
   * saveSession — persists auth state in localStorage so refresh doesn't log user out.
   * In Phase 2B this becomes httpOnly cookies managed by the Go backend.
   */
  function saveSession(userData, orgData, role) {
    setUser(userData);
    setOrg(orgData);
    setMemberRole(role);
    localStorage.setItem('freel_session', JSON.stringify({
      user: userData,
      org: orgData,
      memberRole: role,
    }));
  }

  /**
   * login — called after successful email+password auth.
   * In Phase 2B: calls authService.login() which hits Go backend → Cognito.
   */
  async function login(userData, orgData, role) {
    saveSession(userData, orgData, role);
  }

  /**
   * logout — clears all auth state.
   */
  async function logout() {
    setUser(null);
    setOrg(null);
    setMemberRole(null);
    localStorage.removeItem('freel_session');
  }

  /**
   * updateOrg — called after onboarding completes.
   */
  function updateOrg(updates) {
    const updated = { ...org, ...updates };
    setOrg(updated);
    const session = JSON.parse(localStorage.getItem('freel_session') || '{}');
    localStorage.setItem('freel_session', JSON.stringify({ ...session, org: updated }));
  }

  const isAuthenticated = !!user;

  return (
    <AuthContext.Provider value={{
      user,
      org,
      memberRole,
      isAuthenticated,
      isLoading,
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

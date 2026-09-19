import React, { createContext, useContext, useState, useEffect, useCallback } from 'react';
import { api } from '../services/api';

const AuthContext = createContext(null);

export function AuthProvider({ children }) {
  const [user, setUser] = useState(null);
  const [org, setOrg] = useState(null);
  const [role, setRole] = useState(null);
  const [isBooting, setIsBooting] = useState(true);
  const [isAuthenticating, setIsAuthenticating] = useState(false);
  const [authError, setAuthError] = useState(null);

  // Hydrate session from localStorage and backend
  const checkSession = useCallback(async () => {
    try {
      const storedSession = localStorage.getItem('sportal_session_user');
      const token = api.getAuthToken();

      if (token && storedSession) {
        const parsed = JSON.parse(storedSession);
        setUser(parsed.user);
        setOrg(parsed.org);
        setRole(parsed.role);

        // Verify session freshness with server in background
        try {
          const freshData = await api.getMe();
          if (freshData?.user) {
            setUser(freshData.user);
            setOrg(freshData.org);
            setRole(freshData.role);
            localStorage.setItem('sportal_session_user', JSON.stringify({
              user: freshData.user,
              org: freshData.org,
              role: freshData.role,
              is_internal: freshData.is_internal,
            }));
          }
        } catch (serverErr) {
          // If server rejects with 401 or 403, clear session
          if (serverErr.status === 401 || serverErr.status === 403) {
            api.setAuthToken(null);
            setUser(null);
            setOrg(null);
            setRole(null);
          }
        }
      } else {
        setUser(null);
        setOrg(null);
        setRole(null);
      }
    } catch (e) {
      console.warn('[SPortal Auth] Session initialization error:', e);
      setUser(null);
      setOrg(null);
      setRole(null);
    } finally {
      setIsBooting(false);
    }
  }, []);

  useEffect(() => {
    checkSession();
  }, [checkSession]);

  const login = async (email, password) => {
    setIsAuthenticating(true);
    setAuthError(null);

    try {
      const resp = await api.login({ email, password });
      setUser(resp.user);
      setOrg(resp.org);
      setRole(resp.role);
      setIsAuthenticating(false);
      return { success: true, user: resp.user };
    } catch (err) {
      setIsAuthenticating(false);
      const message = err.message || 'Authentication failed. Please verify credentials.';
      setAuthError(message);
      return {
        success: false,
        error: message,
        status: err.status,
        code: err.code,
      };
    }
  };

  const logout = async () => {
    try {
      await api.logout();
    } catch {
      // Ignore
    } finally {
      api.setAuthToken(null);
      setUser(null);
      setOrg(null);
      setRole(null);
      window.location.href = '/login';
    }
  };

  const hasPermission = (permission) => {
    if (!role) return false;
    // Super Admin / CEO / Owner has all permissions
    if (role.name === 'SUPER_ADMIN' || role.name === 'CEO' || role.name === 'OWNER') {
      return true;
    }
    return Array.isArray(role.permissions) && role.permissions.includes(permission);
  };

  const hasAnyPermission = (permissions = []) => {
    if (!role) return false;
    if (role.name === 'SUPER_ADMIN' || role.name === 'CEO' || role.name === 'OWNER') {
      return true;
    }
    return permissions.some((p) => hasPermission(p));
  };

  const value = {
    user,
    org,
    role,
    isAuthenticated: !!user && !!api.getAuthToken(),
    isBooting,
    isAuthenticating,
    authError,
    login,
    logout,
    hasPermission,
    hasAnyPermission,
    refreshProfile: checkSession,
  };

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
}

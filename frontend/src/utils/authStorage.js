/**
 * authStorage.js - Centralized management of authentication tokens in localStorage.
 */

const ACCESS_TOKEN_KEY = 'freel_access_token';
const ID_TOKEN_KEY = 'freel_id_token';
const REFRESH_TOKEN_KEY = 'freel_refresh_token';
const SESSION_USER_KEY = 'freel_session_user';
const FREEL_SESSION_KEY = 'freel_session'; // Legacy state for backward compatibility if needed

export const authStorage = {
  saveTokens: (tokens) => {
    if (tokens.access_token) localStorage.setItem(ACCESS_TOKEN_KEY, tokens.access_token);
    if (tokens.id_token) localStorage.setItem(ID_TOKEN_KEY, tokens.id_token);
    if (tokens.refresh_token) localStorage.setItem(REFRESH_TOKEN_KEY, tokens.refresh_token);
  },

  getTokens: () => {
    return {
      accessToken: localStorage.getItem(ACCESS_TOKEN_KEY),
      idToken: localStorage.getItem(ID_TOKEN_KEY),
      refreshToken: localStorage.getItem(REFRESH_TOKEN_KEY),
    };
  },

  clearTokens: () => {
    localStorage.removeItem(ACCESS_TOKEN_KEY);
    localStorage.removeItem(ID_TOKEN_KEY);
    localStorage.removeItem(REFRESH_TOKEN_KEY);
  },

  saveSessionUser: (userData) => {
    localStorage.setItem(SESSION_USER_KEY, JSON.stringify(userData));
  },

  getSessionUser: () => {
    try {
      const data = localStorage.getItem(SESSION_USER_KEY);
      return data ? JSON.parse(data) : null;
    } catch {
      return null;
    }
  },

  clearSessionUser: () => {
    localStorage.removeItem(SESSION_USER_KEY);
  },

  clearAll: () => {
    authStorage.clearTokens();
    authStorage.clearSessionUser();
    localStorage.removeItem(FREEL_SESSION_KEY);
  }
};

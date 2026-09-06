/**
 * api.js — Base HTTP client for all authenticated API calls in LogisticsHQ.
 *
 * Features:
 *  - Automatic JWT (access token) injection via Authorization header
 *  - Standardized JSON response parsing & error shaping
 *  - 401 auto-refresh: on a 401, attempts one silent token refresh then retries
 *  - Supports GET, POST, PUT, PATCH, DELETE helpers
 *
 * Usage:
 *   import api from '../services/api';
 *   const leads = await api.get('/api/v1/leads');
 *   const result = await api.post('/api/v1/leads', { name: 'Acme' });
 *
 * Auth flow:
 *   GET /auth/refresh  (with refresh_token in body)
 *   → backend issues new access_token + id_token
 *   → stored back into authStorage
 *   → original request retried once with new token
 *
 * Note: Auth-flow endpoints (signup, login, forgot-password, etc.) use
 * authService.js directly with raw fetch — they run before any token exists.
 */

import { authStorage } from '../utils/authStorage';

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080';

// ── Internal helpers ──────────────────────────────────────────────────────────

/**
 * Build the Authorization header value from the stored access token.
 * Returns null if no token is available.
 */
function getAuthHeader() {
  const { accessToken } = authStorage.getTokens();
  return accessToken ? `Bearer ${accessToken}` : null;
}

/**
 * Attempt to refresh the access token using the stored refresh token.
 * On success, saves new tokens to authStorage and returns the new access token.
 * On failure, clears all tokens and returns null.
 *
 * The backend endpoint is POST /auth/refresh (to be implemented in Sprint 3).
 * Until then this function handles the failure gracefully.
 */
export async function refreshAccessToken() {
  const { refreshToken } = authStorage.getTokens();
  if (!refreshToken) return null;

  const sessionUser = authStorage.getSessionUser();
  const email = sessionUser?.user?.email;

  try {
    const response = await fetch(`${API_BASE_URL}/auth/refresh`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ refresh_token: refreshToken, email: email || '' }),
    });

    if (!response.ok) {
      // Refresh failed — clear all auth data so user is prompted to log in again
      authStorage.clearAll();
      window.location.href = '/login?expired=true';
      return null;
    }

    const data = await response.json();
    const payload = data?.data || data;

    if (!payload?.access_token) {
      authStorage.clearAll();
      window.location.href = '/login?expired=true';
      return null;
    }

    // Save new tokens; refresh_token usually stays the same from Cognito
    authStorage.saveTokens({
      access_token: payload.access_token,
      id_token: payload.id_token,
      refresh_token: payload.refresh_token || refreshToken, // keep existing if not rotated
    });

    return payload.access_token;
  } catch {
    authStorage.clearAll();
    window.location.href = '/login?expired=true';
    return null;
  }
}

/**
 * Parse a fetch Response into normalized data or throw a normalized error object.
 *
 * @param {Response} response
 * @returns {Promise<any>}
 */
async function parseResponse(response) {
  // Try to parse JSON body
  let data;
  try {
    data = await response.json();
  } catch {
    // No body or non-JSON body
    if (!response.ok) {
      throw {
        code: 'NETWORK_ERROR',
        message: 'An unexpected error occurred. Please try again.',
        status: response.status,
      };
    }
    return null;
  }

  if (!response.ok) {
    // Shape error consistently: { code, message, status, details? }
    throw {
      code: data?.error_code || data?.code || data?.error?.code || 'API_ERROR',
      message: data?.message || (typeof data?.error === 'string' ? data.error : 'An error occurred. Please try again.'),
      status: response.status,
      details: data,
    };
  }

  // Unwrap standard backend envelope: { success, message, data }
  // If data field exists and total_count is not present, return it directly; otherwise return full body
  if (data?.data !== undefined) {
    if (
      data?.total_count !== undefined ||
      data?.kpis !== undefined ||
      data?.pagination !== undefined ||
      data?.carriers !== undefined ||
      data?.items !== undefined
    ) {
      return data;
    }
    return data.data;
  }
  return data;
}

/**
 * Core fetch wrapper with JWT injection and 401 auto-refresh.
 *
 * @param {string} path       — API path, e.g. '/api/v1/leads'
 * @param {RequestInit} opts  — fetch options (method, body, headers, etc.)
 * @param {boolean} isRetry   — internal flag; prevents infinite refresh loops
 * @returns {Promise<any>}
 */
async function request(path, opts = {}, isRetry = false) {
  const authHeader = getAuthHeader();

  const headers = {
    'Content-Type': 'application/json',
    ...(authHeader ? { Authorization: authHeader } : {}),
    ...opts.headers,
  };

  const response = await fetch(`${API_BASE_URL}${path}`, {
    ...opts,
    headers,
  });

  // Handle 401: attempt one token refresh then retry
  if (response.status === 401 && !isRetry) {
    const newToken = await refreshAccessToken();
    if (newToken) {
      // Retry with new token
      return request(path, opts, true);
    }
    // Refresh failed — propagate 401 so the caller / AuthContext can log out
    throw {
      code: 'UNAUTHORIZED',
      message: 'Your session has expired. Please log in again.',
      status: 401,
    };
  }

  return parseResponse(response);
}

// ── Public HTTP method helpers ────────────────────────────────────────────────

const api = {
  /**
   * GET request.
   * @param {string} path
   * @param {RequestInit} [opts]
   */
  get: (path, opts = {}) =>
    request(path, { ...opts, method: 'GET' }),

  /**
   * POST request.
   * @param {string} path
   * @param {any} [body]
   * @param {RequestInit} [opts]
   */
  post: (path, body, opts = {}) =>
    request(path, { ...opts, method: 'POST', body: body != null ? JSON.stringify(body) : undefined }),

  /**
   * PUT request.
   * @param {string} path
   * @param {any} [body]
   * @param {RequestInit} [opts]
   */
  put: (path, body, opts = {}) =>
    request(path, { ...opts, method: 'PUT', body: body != null ? JSON.stringify(body) : undefined }),

  /**
   * PATCH request.
   * @param {string} path
   * @param {any} [body]
   * @param {RequestInit} [opts]
   */
  patch: (path, body, opts = {}) =>
    request(path, { ...opts, method: 'PATCH', body: body != null ? JSON.stringify(body) : undefined }),

  /**
   * DELETE request.
   * @param {string} path
   * @param {RequestInit} [opts]
   */
  delete: (path, opts = {}) =>
    request(path, { ...opts, method: 'DELETE' }),
};

export default api;

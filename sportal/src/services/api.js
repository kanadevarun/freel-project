import { ENV } from '../app/config/env';

/**
 * Standardized HTTP API Client for SPortal
 */
class ApiClient {
  constructor(baseUrl) {
    this.baseUrl = baseUrl;
  }

  getAuthToken() {
    return localStorage.getItem('sportal_access_token') || localStorage.getItem('sportal_auth_token') || '';
  }

  setAuthToken(token) {
    if (token) {
      localStorage.setItem('sportal_access_token', token);
      localStorage.setItem('sportal_auth_token', token);
    } else {
      localStorage.removeItem('sportal_access_token');
      localStorage.removeItem('sportal_auth_token');
      localStorage.removeItem('sportal_session_user');
    }
  }

  async request(endpoint, options = {}) {
    const url = `${this.baseUrl}${endpoint}`;
    const token = this.getAuthToken();
    const isFormData = typeof FormData !== 'undefined' && options.body instanceof FormData;

    const headers = {
      Accept: 'application/json',
      ...options.headers,
    };

    if (!isFormData && !headers['Content-Type']) {
      headers['Content-Type'] = 'application/json';
    } else if (isFormData) {
      delete headers['Content-Type'];
    }

    if (token) {
      headers.Authorization = `Bearer ${token}`;
    }

    try {
      const response = await fetch(url, {
        ...options,
        headers,
      });

      const data = await response.json().catch(() => null);

      if (!response.ok) {
        const errorMsg = data?.message || data?.error?.message || `HTTP ${response.status}: Request failed`;
        const error = new Error(errorMsg);
        error.status = response.status;
        error.data = data;
        error.code = data?.error?.code;

        // Auto-redirect on 401 if not on login page
        if (response.status === 401 && !window.location.pathname.includes('/login')) {
          this.setAuthToken(null);
          window.location.href = '/login?expired=true';
        }

        throw error;
      }

      return data?.data !== undefined ? data.data : data;
    } catch (err) {
      console.error(`[SPortal API] Request to ${endpoint} failed:`, err);
      throw err;
    }
  }

  get(endpoint, options = {}) {
    return this.request(endpoint, { ...options, method: 'GET' });
  }

  post(endpoint, body, options = {}) {
    const isFormData = typeof FormData !== 'undefined' && body instanceof FormData;
    return this.request(endpoint, {
      ...options,
      method: 'POST',
      body: isFormData ? body : JSON.stringify(body),
    });
  }

  put(endpoint, body, options = {}) {
    return this.request(endpoint, {
      ...options,
      method: 'PUT',
      body: JSON.stringify(body),
    });
  }

  patch(endpoint, body, options = {}) {
    return this.request(endpoint, {
      ...options,
      method: 'PATCH',
      body: JSON.stringify(body),
    });
  }

  delete(endpoint, options = {}) {
    return this.request(endpoint, { ...options, method: 'DELETE' });
  }

  // --- SPortal S2 Authentication Endpoints ---

  async login(credentials) {
    const data = await this.post('/api/v1/sportal/auth/login', credentials);
    if (data?.access_token) {
      this.setAuthToken(data.access_token);
      localStorage.setItem('sportal_session_user', JSON.stringify({
        user: data.user,
        org: data.org,
        role: data.role,
        is_internal: data.is_internal,
      }));
    }
    return data;
  }

  async getMe() {
    return this.get('/api/v1/sportal/auth/me');
  }

  async logout() {
    try {
      await this.post('/api/v1/sportal/auth/logout', {});
    } catch {
      // Ignore network errors on logout
    } finally {
      this.setAuthToken(null);
    }
  }

  async getPermissions() {
    return this.get('/api/v1/sportal/auth/permissions');
  }

  async getSensitiveFinancialData() {
    return this.get('/api/v1/sportal/finance/sensitive');
  }
}

export const api = new ApiClient(ENV.API_BASE_URL);

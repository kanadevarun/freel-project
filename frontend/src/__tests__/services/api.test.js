/**
 * Unit tests for api.js — base HTTP client
 *
 * Tests cover:
 *  - GET, POST, PUT, PATCH, DELETE helpers
 *  - JWT injection from authStorage
 *  - Response envelope unwrapping (data field)
 *  - Error shaping on non-OK responses
 *  - 401 auto-refresh: refreshes token and retries once
 *  - 401 with failed refresh: propagates UNAUTHORIZED error
 *  - refreshAccessToken: saves new tokens on success
 *  - refreshAccessToken: clears tokens on failure
 *  - Non-JSON body handling
 */

import { describe, it, expect, vi, afterEach } from 'vitest';
import api, { refreshAccessToken } from '../../services/api';
import { authStorage } from '../../utils/authStorage';

// ── fetch mock helpers ────────────────────────────────────────────────────────

function mockFetch(responses) {
  // Accepts an array of responses so we can sequence: first call, second call, etc.
  let callCount = 0;
  return vi.fn().mockImplementation(() => {
    const res = Array.isArray(responses) ? responses[callCount++] : responses;
    return Promise.resolve(res);
  });
}

function okResponse(data) {
  return {
    ok: true,
    status: 200,
    json: async () => data,
  };
}

function errorResponse(data, status = 400) {
  return {
    ok: false,
    status,
    json: async () => data,
  };
}

function noBodyResponse(status = 204) {
  return {
    ok: true,
    status,
    json: async () => { throw new Error('No body'); },
  };
}

// ── tests ─────────────────────────────────────────────────────────────────────

describe('api.js', () => {
  afterEach(() => {
    vi.restoreAllMocks();
    localStorage.clear();
  });

  // ── JWT injection ────────────────────────────────────────────────────────
  describe('JWT injection', () => {
    it('adds Authorization header when access token is stored', async () => {
      authStorage.saveTokens({ access_token: 'my-access-token' });

      const fetchMock = mockFetch(okResponse({ message: 'ok' }));
      vi.stubGlobal('fetch', fetchMock);

      await api.get('/api/v1/test');

      const calledWith = fetchMock.mock.calls[0];
      expect(calledWith[1].headers.Authorization).toBe('Bearer my-access-token');
    });

    it('does NOT add Authorization header when no token is stored', async () => {
      const fetchMock = mockFetch(okResponse({ message: 'ok' }));
      vi.stubGlobal('fetch', fetchMock);

      await api.get('/api/v1/test');

      const calledWith = fetchMock.mock.calls[0];
      expect(calledWith[1].headers.Authorization).toBeUndefined();
    });
  });

  // ── HTTP method helpers ──────────────────────────────────────────────────
  describe('HTTP methods', () => {
    it('GET: calls fetch with GET method', async () => {
      const fetchMock = mockFetch(okResponse({ id: 1 }));
      vi.stubGlobal('fetch', fetchMock);

      const result = await api.get('/api/v1/leads');
      expect(fetchMock.mock.calls[0][1].method).toBe('GET');
      expect(result).toEqual({ id: 1 });
    });

    it('POST: calls fetch with POST method and JSON body', async () => {
      const fetchMock = mockFetch(okResponse({ created: true }));
      vi.stubGlobal('fetch', fetchMock);

      const result = await api.post('/api/v1/leads', { name: 'Acme' });
      const callArgs = fetchMock.mock.calls[0][1];
      expect(callArgs.method).toBe('POST');
      expect(callArgs.body).toBe(JSON.stringify({ name: 'Acme' }));
      expect(result).toEqual({ created: true });
    });

    it('PUT: calls fetch with PUT method', async () => {
      const fetchMock = mockFetch(okResponse({ updated: true }));
      vi.stubGlobal('fetch', fetchMock);

      await api.put('/api/v1/leads/1', { name: 'Updated' });
      expect(fetchMock.mock.calls[0][1].method).toBe('PUT');
    });

    it('PATCH: calls fetch with PATCH method', async () => {
      const fetchMock = mockFetch(okResponse({ patched: true }));
      vi.stubGlobal('fetch', fetchMock);

      await api.patch('/api/v1/leads/1', { status: 'active' });
      expect(fetchMock.mock.calls[0][1].method).toBe('PATCH');
    });

    it('DELETE: calls fetch with DELETE method and no body', async () => {
      const fetchMock = mockFetch(noBodyResponse(204));
      vi.stubGlobal('fetch', fetchMock);

      const result = await api.delete('/api/v1/leads/1');
      expect(fetchMock.mock.calls[0][1].method).toBe('DELETE');
      expect(result).toBeNull(); // 204 No Content
    });
  });

  // ── response envelope unwrapping ─────────────────────────────────────────
  describe('response envelope unwrapping', () => {
    it('unwraps { data } field from standard backend envelope', async () => {
      vi.stubGlobal('fetch', mockFetch(okResponse({
        success: true,
        message: 'Leads fetched',
        data: [{ id: 1, name: 'Acme' }],
      })));

      const result = await api.get('/api/v1/leads');
      expect(result).toEqual([{ id: 1, name: 'Acme' }]);
    });

    it('returns full body when no data field is present', async () => {
      vi.stubGlobal('fetch', mockFetch(okResponse({ message: 'ok' })));
      const result = await api.get('/api/v1/ping');
      expect(result).toEqual({ message: 'ok' });
    });
  });

  // ── error shaping ────────────────────────────────────────────────────────
  describe('error shaping on non-OK responses', () => {
    it('throws a shaped error object with code, message, and status', async () => {
      vi.stubGlobal('fetch', mockFetch(errorResponse({ message: 'Not found', error_code: 'LEAD_NOT_FOUND' }, 404)));

      await expect(api.get('/api/v1/leads/999')).rejects.toMatchObject({
        code: 'LEAD_NOT_FOUND',
        message: 'Not found',
        status: 404,
      });
    });

    it('falls back to API_ERROR code when no error_code is present', async () => {
      vi.stubGlobal('fetch', mockFetch(errorResponse({ message: 'Something went wrong' }, 500)));

      await expect(api.post('/api/v1/leads', {})).rejects.toMatchObject({
        code: 'API_ERROR',
        status: 500,
      });
    });

    it('throws NETWORK_ERROR when response is not valid JSON', async () => {
      vi.stubGlobal('fetch', mockFetch({
        ok: false,
        status: 503,
        json: async () => { throw new Error('Invalid JSON'); },
      }));

      await expect(api.get('/api/v1/leads')).rejects.toMatchObject({
        code: 'NETWORK_ERROR',
      });
    });
  });

  // ── 401 auto-refresh ─────────────────────────────────────────────────────
  describe('401 auto-refresh', () => {
    it('refreshes token and retries original request on 401', async () => {
      // Store a refresh token
      authStorage.saveTokens({
        access_token: 'expired-token',
        refresh_token: 'valid-refresh-token',
      });

      const fetchMock = mockFetch([
        // 1st call: original request → 401
        errorResponse({ message: 'Unauthorized' }, 401),
        // 2nd call: POST /auth/refresh → success with new token
        okResponse({ access_token: 'new-access-token', id_token: 'new-id-token' }),
        // 3rd call: retried original request → success
        okResponse({ data: [{ id: 1 }] }),
      ]);
      vi.stubGlobal('fetch', fetchMock);

      const result = await api.get('/api/v1/leads');

      // Should have made 3 fetch calls total
      expect(fetchMock).toHaveBeenCalledTimes(3);

      // Retried call should use the new token
      const retriedCall = fetchMock.mock.calls[2];
      expect(retriedCall[1].headers.Authorization).toBe('Bearer new-access-token');

      // Final result should be unwrapped data
      expect(result).toEqual([{ id: 1 }]);
    });

    it('throws UNAUTHORIZED if refresh token is missing', async () => {
      // No tokens stored at all
      const fetchMock = mockFetch(errorResponse({ message: 'Unauthorized' }, 401));
      vi.stubGlobal('fetch', fetchMock);

      await expect(api.get('/api/v1/leads')).rejects.toMatchObject({
        code: 'UNAUTHORIZED',
        status: 401,
      });
    });

    it('throws UNAUTHORIZED and clears tokens if refresh fails', async () => {
      authStorage.saveTokens({
        access_token: 'expired',
        refresh_token: 'bad-refresh-token',
      });

      const fetchMock = mockFetch([
        // 1st: original request → 401
        errorResponse({ message: 'Unauthorized' }, 401),
        // 2nd: refresh → 401 (invalid refresh token)
        errorResponse({ message: 'Invalid refresh token' }, 401),
      ]);
      vi.stubGlobal('fetch', fetchMock);

      await expect(api.get('/api/v1/leads')).rejects.toMatchObject({
        code: 'UNAUTHORIZED',
      });

      // Tokens should be cleared
      const tokens = authStorage.getTokens();
      expect(tokens.accessToken).toBeNull();
      expect(tokens.refreshToken).toBeNull();
    });

    it('does NOT retry a second time on 401 after refresh (prevents loops)', async () => {
      authStorage.saveTokens({
        access_token: 'expired',
        refresh_token: 'refresh',
      });

      const fetchMock = mockFetch([
        // 1st: original → 401
        errorResponse({ message: 'Unauthorized' }, 401),
        // 2nd: refresh → success
        okResponse({ access_token: 'new-token', id_token: 'new-id' }),
        // 3rd: retry → still 401 (e.g. permission revoked)
        errorResponse({ message: 'Unauthorized' }, 401),
      ]);
      vi.stubGlobal('fetch', fetchMock);

      await expect(api.get('/api/v1/leads')).rejects.toMatchObject({
        code: 'API_ERROR',
        status: 401,
      });

      // Should NOT make a 4th call (no infinite loop)
      expect(fetchMock).toHaveBeenCalledTimes(3);
    });
  });

  // ── refreshAccessToken ───────────────────────────────────────────────────
  describe('refreshAccessToken()', () => {
    it('returns null when no refresh token is stored', async () => {
      const result = await refreshAccessToken();
      expect(result).toBeNull();
    });

    it('saves new tokens on successful refresh and returns new access token', async () => {
      authStorage.saveTokens({ refresh_token: 'valid-refresh' });
      vi.stubGlobal('fetch', mockFetch(okResponse({
        access_token: 'brand-new-token',
        id_token: 'brand-new-id',
        refresh_token: 'same-refresh',
      })));

      const newToken = await refreshAccessToken();
      expect(newToken).toBe('brand-new-token');

      const tokens = authStorage.getTokens();
      expect(tokens.accessToken).toBe('brand-new-token');
    });

    it('clears all tokens and returns null on failed refresh', async () => {
      authStorage.saveTokens({ access_token: 'old', refresh_token: 'bad' });
      vi.stubGlobal('fetch', mockFetch(errorResponse({ message: 'Invalid' }, 401)));

      const result = await refreshAccessToken();
      expect(result).toBeNull();
      expect(authStorage.getTokens().accessToken).toBeNull();
    });

    it('clears tokens and returns null when fetch throws (network error)', async () => {
      authStorage.saveTokens({ refresh_token: 'valid' });
      vi.stubGlobal('fetch', vi.fn().mockRejectedValue(new Error('Network error')));

      const result = await refreshAccessToken();
      expect(result).toBeNull();
      expect(authStorage.getTokens().accessToken).toBeNull();
    });
  });
});

/**
 * apiUrl.js — Dynamic API Base URL resolution.
 *
 * Automatically resolves the backend API host based on the browser's current address:
 * - If VITE_API_BASE_URL is explicitly set, use it.
 * - If loaded on a LAN or remote machine (e.g. http://192.168.1.33:5173),
 *   automatically routes API calls to the same host on port 8080 (http://192.168.1.33:8080).
 * - Defaults to http://localhost:8080 when running locally.
 */

export function getApiBaseUrl() {
  if (import.meta.env.VITE_API_BASE_URL) {
    return import.meta.env.VITE_API_BASE_URL;
  }
  if (typeof window !== 'undefined' && window.location.hostname) {
    const host = window.location.hostname;
    const protocol = window.location.protocol || 'http:';
    if (host !== 'localhost' && host !== '127.0.0.1') {
      return `${protocol}//${host}:8080`;
    }
  }
  return 'http://localhost:8080';
}

export const API_BASE_URL = getApiBaseUrl();
export default API_BASE_URL;

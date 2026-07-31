/**
 * leadsService.js — All API calls for the Leads module.
 *
 * Uses the authenticated `api` client from `./api.js` which automatically
 * injects the JWT token and handles 401 refresh logic.
 *
 * Available Endpoints:
 *   GET    /api/v1/leads               — List all leads (with optional filters)
 *   POST   /api/v1/leads               — Create a new lead manually
 *   GET    /api/v1/leads/:id           — Get a single lead's full details
 *   PUT    /api/v1/leads/:id           — Update a lead (status, AI score, etc.)
 *   DELETE /api/v1/leads/:id           — Remove a lead
 *   POST   /api/v1/leads/import        — Bulk import leads from a CSV file
 */

import api from './api';
import { authStorage } from '../utils/authStorage';

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080';

// ── LIST LEADS ────────────────────────────────────────────────────────────────

/**
 * listLeads — Fetch a paginated, optionally filtered list of leads.
 * @param {{ status?: string, limit?: number, offset?: number }} params
 * @returns {Promise<{ leads: Lead[], total: number }>}
 */
export async function listLeads({ status, limit = 50, offset = 0 } = {}) {
  const queryParams = new URLSearchParams({ limit, offset });
  if (status) queryParams.set('status', status);
  return api.get(`/api/v1/leads?${queryParams.toString()}`);
}

// ── CREATE LEAD ───────────────────────────────────────────────────────────────

/**
 * createLead — Manually create a new lead.
 * @param {{ company_name: string, contact_name?: string, email?: string, source?: string, notes?: string }} payload
 * @returns {Promise<Lead>}
 */
export async function createLead(payload) {
  return api.post('/api/v1/leads', payload);
}

// ── GET LEAD ──────────────────────────────────────────────────────────────────

/**
 * getLead — Fetch full details for a single lead by ID.
 * @param {number} id
 * @returns {Promise<Lead>}
 */
export async function getLead(id) {
  return api.get(`/api/v1/leads/${id}`);
}

// ── UPDATE LEAD ───────────────────────────────────────────────────────────────

/**
 * updateLead — Update a lead's fields (status, notes, etc.)
 * @param {number} id
 * @param {Partial<Lead>} payload
 * @returns {Promise<Lead>}
 */
export async function updateLead(id, payload) {
  return api.put(`/api/v1/leads/${id}`, payload);
}

// ── DELETE LEAD ───────────────────────────────────────────────────────────────

/**
 * deleteLead — Permanently delete a lead.
 * @param {number} id
 * @returns {Promise<void>}
 */
export async function deleteLead(id) {
  return api.delete(`/api/v1/leads/${id}`);
}

// ── IMPORT LEADS (CSV) ────────────────────────────────────────────────────────

/**
 * importLeads — Upload a CSV file for bulk lead import.
 *
 * Simple meaning: This sends the file over to the server using FormData (not JSON).
 * The backend will parse the CSV, create all the leads in the database, and trigger
 * the AI background worker for each one.
 *
 * @param {File} file — The CSV File object from the browser's file input
 * @returns {Promise<{ imported_count: number, total_attempted: number }>}
 */
export async function importLeads(file) {
  const { accessToken } = authStorage.getTokens();

  const formData = new FormData();
  formData.append('file', file);

  const response = await fetch(`${API_BASE_URL}/api/v1/leads/import`, {
    method: 'POST',
    headers: {
      // Note: Do NOT set Content-Type here! The browser sets it automatically
      // with the correct multipart/form-data boundary when using FormData.
      ...(accessToken ? { Authorization: `Bearer ${accessToken}` } : {}),
    },
    body: formData,
  });

  const data = await response.json();
  if (!response.ok) {
    throw { message: data?.message || 'Import failed', status: response.status };
  }
  return data?.data !== undefined ? data.data : data;
}

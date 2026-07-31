/**
 * outreachService.js — All API calls for the Outreach module.
 *
 * Uses the authenticated `api` client from `./api.js` which automatically
 * injects the JWT token and handles 401 refresh logic.
 *
 * Endpoints covered:
 *   GET    /api/v1/outreach/campaigns            — List all campaigns
 *   POST   /api/v1/outreach/campaigns            — Create a new DRAFT campaign
 *   GET    /api/v1/outreach/campaigns/:id        — Get a single campaign
 *   POST   /api/v1/outreach/campaigns/:id/activate — Start a campaign
 *   POST   /api/v1/outreach/campaigns/:id/pause    — Pause a campaign
 *   DELETE /api/v1/outreach/campaigns/:id        — Delete a campaign
 *   POST   /api/v1/outreach/generate-email       — AI email generation
 */

import api from './api';

// ── CAMPAIGN CRUD ─────────────────────────────────────────────────────────────

/**
 * listCampaigns — Fetch a paginated list of all campaigns for the org.
 * @param {{ limit?: number, offset?: number }} params
 * @returns {Promise<{ campaigns: Campaign[], total: number }>}
 */
export async function listCampaigns({ limit = 50, offset = 0 } = {}) {
  return api.get(`/api/v1/outreach/campaigns?limit=${limit}&offset=${offset}`);
}

/**
 * getCampaign — Fetch a single campaign with its sequences.
 * @param {number} id — Campaign ID
 * @returns {Promise<Campaign>}
 */
export async function getCampaign(id) {
  return api.get(`/api/v1/outreach/campaigns/${id}`);
}

/**
 * createCampaign — Create a new campaign in DRAFT status.
 * Simple meaning: This sends the campaign name to the backend which creates
 * a blank campaign. You add email steps to it separately.
 * @param {{ name: string }} payload
 * @returns {Promise<Campaign>}
 */
export async function createCampaign(payload) {
  return api.post('/api/v1/outreach/campaigns', payload);
}

/**
 * activateCampaign — Transition a campaign from DRAFT/PAUSED → ACTIVE.
 * Simple meaning: Clicking "Launch" → this call tells the backend to start
 * sending emails for this campaign.
 * @param {number} id
 * @returns {Promise<Campaign>}
 */
export async function activateCampaign(id) {
  return api.post(`/api/v1/outreach/campaigns/${id}/activate`);
}

/**
 * pauseCampaign — Temporarily stop a campaign (ACTIVE → PAUSED).
 * @param {number} id
 * @returns {Promise<Campaign>}
 */
export async function pauseCampaign(id) {
  return api.post(`/api/v1/outreach/campaigns/${id}/pause`);
}

/**
 * deleteCampaign — Permanently delete a campaign and all its sequences.
 * @param {number} id
 * @returns {Promise<void>}
 */
export async function deleteCampaign(id) {
  return api.delete(`/api/v1/outreach/campaigns/${id}`);
}

// ── AI EMAIL GENERATION ───────────────────────────────────────────────────────

/**
 * generateEmail — Ask the AI to write a personalized outreach email.
 *
 * Simple meaning: You provide context about the company you want to email.
 * The AI reads it, picks an angle, and writes a ready-to-send subject line
 * and email body. The user can then edit it before using it.
 *
 * @param {{ company_name: string, industry?: string, goal?: string }} payload
 * @returns {Promise<{ subject: string, body: string }>}
 */
export async function generateEmail(payload) {
  return api.post('/api/v1/outreach/generate-email', payload);
}

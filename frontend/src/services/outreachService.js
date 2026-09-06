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
 * @param {{ company_name: string, industry?: string, goal?: string }} payload
 * @returns {Promise<{ subject: string, body: string }>}
 */
export async function generateEmail(payload) {
  return api.post('/api/v1/outreach/generate-email', payload);
}

// ── SEQUENCE STEPS ────────────────────────────────────────────────────────────

export async function getCampaignSequence(campaignId) {
  return api.get(`/api/v1/outreach/campaigns/${campaignId}/sequence`);
}

export async function addCampaignSequenceStep(campaignId, stepData) {
  return api.post(`/api/v1/outreach/campaigns/${campaignId}/sequence`, stepData);
}

export async function updateCampaignSequenceStep(campaignId, stepId, stepData) {
  return api.put(`/api/v1/outreach/campaigns/${campaignId}/sequence/${stepId}`, stepData);
}

export async function deleteCampaignSequenceStep(campaignId, stepId) {
  return api.delete(`/api/v1/outreach/campaigns/${campaignId}/sequence/${stepId}`);
}

export async function reorderCampaignSequence(campaignId, stepIds) {
  return api.put(`/api/v1/outreach/campaigns/${campaignId}/sequence/reorder`, { step_ids: stepIds });
}

// ── CAMPAIGN AUDIENCE ─────────────────────────────────────────────────────────

export async function getCampaignAudience(campaignId) {
  return api.get(`/api/v1/outreach/campaigns/${campaignId}/audience`);
}

export async function addCampaignAudience(campaignId, leadIds) {
  return api.post(`/api/v1/outreach/campaigns/${campaignId}/audience`, { lead_ids: leadIds });
}

export async function removeCampaignAudience(campaignId, leadId) {
  return api.delete(`/api/v1/outreach/campaigns/${campaignId}/audience/${leadId}`);
}

// ── ANALYTICS, LEADS GENERATED & INSIGHTS ────────────────────────────────────

export async function getOutreachAnalytics() {
  return api.get('/api/v1/outreach/analytics');
}

export async function getCampaignAnalytics(campaignId) {
  return api.get(`/api/v1/outreach/campaigns/${campaignId}/analytics`);
}

export async function getCampaignLeads(campaignId) {
  return api.get(`/api/v1/outreach/campaigns/${campaignId}/leads`);
}

export async function getCampaignInsights(campaignId) {
  return api.get(`/api/v1/outreach/campaigns/${campaignId}/insights`);
}

export async function getConversionFunnel() {
  return api.get('/api/v1/outreach/conversion-funnel');
}

// ── OUTREACH ACTIVITIES CRUD ───────────────────────────────────────────────

export async function createOutreachActivity(payload) {
  return api.post('/api/v1/outreach/activities', payload);
}

export async function getOutreachActivity(id) {
  return api.get(`/api/v1/outreach/activities/${id}`);
}

export async function updateOutreachActivity(id, payload) {
  return api.put(`/api/v1/outreach/activities/${id}`, payload);
}

export async function completeOutreachActivity(id) {
  return api.put(`/api/v1/outreach/activities/${id}/complete`);
}

export async function deleteOutreachActivity(id) {
  return api.delete(`/api/v1/outreach/activities/${id}`);
}

// ── ENGAGEMENT & PROSPECTS SERVICE METHODS ────────────────────────────────

export async function getCampaignRecipients(campaignId) {
  return api.get(`/api/v1/outreach/campaigns/${campaignId}/recipients`);
}

export async function getCampaignActivity(campaignId) {
  return api.get(`/api/v1/outreach/campaigns/${campaignId}/activity`);
}

export async function getProspects() {
  return api.get('/api/v1/outreach/prospects');
}

export async function getProspectEngagement(leadId) {
  return api.get(`/api/v1/outreach/prospects/${leadId}/engagement`);
}

export async function getLeadOutreachActivity(leadId) {
  return api.get(`/api/v1/outreach/prospects/${leadId}/activity`);
}

export async function getProspectDetail(leadId) {
  return api.get(`/api/v1/outreach/prospects/${leadId}`);
}

export async function enrollProspect(campaignId, leadId) {
  return api.post('/api/v1/outreach/prospects', { campaign_id: campaignId, lead_id: leadId });
}

export async function updateProspect(leadId, campaignId, status, currentStep) {
  return api.put(`/api/v1/outreach/prospects/${leadId}`, { campaign_id: campaignId, status, current_step: currentStep });
}

export async function pauseProspect(leadId, campaignId) {
  return api.post(`/api/v1/outreach/prospects/${leadId}/pause`, { campaign_id: campaignId });
}

export async function resumeProspect(leadId, campaignId) {
  return api.post(`/api/v1/outreach/prospects/${leadId}/resume`, { campaign_id: campaignId });
}

export async function stopProspect(leadId, campaignId) {
  return api.post(`/api/v1/outreach/prospects/${leadId}/stop`, { campaign_id: campaignId });
}

export async function getFollowUps(filter = '') {
  return api.get(`/api/v1/outreach/follow-ups?filter=${filter}`);
}

export async function createFollowUp(payload) {
  return api.post('/api/v1/outreach/follow-ups', payload);
}

export async function getFollowUp(id) {
  return api.get(`/api/v1/outreach/follow-ups/${id}`);
}

export async function updateFollowUp(id, payload) {
  return api.put(`/api/v1/outreach/follow-ups/${id}`, payload);
}

export async function completeFollowUp(id) {
  return api.post(`/api/v1/outreach/follow-ups/${id}/complete`);
}

export async function cancelFollowUp(id) {
  return api.post(`/api/v1/outreach/follow-ups/${id}/cancel`);
}

export async function rescheduleFollowUp(id, scheduledAt) {
  return api.post(`/api/v1/outreach/follow-ups/${id}/reschedule`, { scheduled_at: scheduledAt });
}


/**
 * constants.js — All constants for the Outreach module.
 *
 * Simple meaning: Instead of typing 'DRAFT' or 'EMAIL' as raw strings
 * throughout the codebase (which can cause hard-to-spot typos), we define
 * them here once and import them wherever needed.
 *
 * Pattern: If a value needs to change (e.g., backend renames 'ACTIVE' to 'RUNNING'),
 * you only need to update it in this one file.
 */

// ── Campaign Status ───────────────────────────────────────────────────────────

/**
 * CAMPAIGN_STATUS — All valid lifecycle states of an outreach campaign.
 *
 * Flow:
 *   DRAFT → ACTIVE → PAUSED → ACTIVE (can toggle between ACTIVE and PAUSED)
 *                           → COMPLETED (when all steps are sent)
 */
export const CAMPAIGN_STATUS = {
  DRAFT:     'DRAFT',
  ACTIVE:    'ACTIVE',
  PAUSED:    'PAUSED',
  COMPLETED: 'COMPLETED',
};

// ── Channel ───────────────────────────────────────────────────────────────────

/**
 * CHANNEL — Communication channels available for outreach sequence steps.
 */
export const CHANNEL = {
  EMAIL:    'EMAIL',
  LINKEDIN: 'LINKEDIN',
};

// ── Status Display Config ─────────────────────────────────────────────────────

/**
 * CAMPAIGN_STATUS_CONFIG — Maps each status to its UI label, badge color, and emoji.
 * Simple meaning: This is a lookup table that tells the UI how to display each status.
 * Example: CAMPAIGN_STATUS_CONFIG[CAMPAIGN_STATUS.ACTIVE] → { label: 'Active', type: 'success', emoji: '🟢' }
 */
export const CAMPAIGN_STATUS_CONFIG = {
  [CAMPAIGN_STATUS.DRAFT]:     { label: 'Draft',     type: 'neutral', emoji: '📝' },
  [CAMPAIGN_STATUS.ACTIVE]:    { label: 'Active',    type: 'success', emoji: '🟢' },
  [CAMPAIGN_STATUS.PAUSED]:    { label: 'Paused',    type: 'warning', emoji: '⏸️'  },
  [CAMPAIGN_STATUS.COMPLETED]: { label: 'Completed', type: 'info',    emoji: '✅' },
};

// ── Action Labels ─────────────────────────────────────────────────────────────

/**
 * CAMPAIGN_ACTIONS — User-visible action labels for campaign operations.
 */
export const CAMPAIGN_ACTIONS = {
  ACTIVATE: 'activate',
  PAUSE:    'pause',
  DELETE:   'delete',
};

// ── Tabs ──────────────────────────────────────────────────────────────────────

/**
 * OUTREACH_TABS — Tab definitions for the main Outreach page.
 */
export const OUTREACH_TABS = {
  CAMPAIGNS: 'campaigns',
  COMPOSER:  'composer',
};

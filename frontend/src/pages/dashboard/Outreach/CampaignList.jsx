import { useState } from 'react';
import PropTypes from 'prop-types';
import StatusBadge from '../../../components/dashboard/StatusBadge';
import {
  CAMPAIGN_STATUS,
  CAMPAIGN_STATUS_CONFIG,
  CAMPAIGN_ACTIONS,
} from './constants';
import {
  activateCampaign,
  pauseCampaign,
  deleteCampaign,
} from '../../../services/outreachService';

/**
 * CampaignList — A sortable table showing all outreach campaigns.
 *
 * Simple meaning: This is the "spreadsheet view" of all your campaigns.
 * Each row shows a campaign's name, status, date, and action buttons.
 * Clicking "Launch" activates it; clicking "Pause" stops it temporarily.
 *
 * @param {{ campaigns: Campaign[], loading: boolean, onCampaignsChanged: () => void }}
 */
export default function CampaignList({ campaigns, loading, onCampaignsChanged }) {
  // Which campaign's action is currently loading (to show spinner on that specific row)
  const [actionLoadingId, setActionLoadingId] = useState(null);

  // ── Action Handlers ───────────────────────────────────────────────────────

  /**
   * handleAction — Calls the correct API based on which button was clicked.
   * @param {string} action — One of CAMPAIGN_ACTIONS values
   * @param {Campaign} campaign
   */
  async function handleAction(action, campaign) {
    if (!confirm(getConfirmMessage(action, campaign.name))) return;

    setActionLoadingId(campaign.id);
    try {
      if (action === CAMPAIGN_ACTIONS.ACTIVATE) await activateCampaign(campaign.id);
      if (action === CAMPAIGN_ACTIONS.PAUSE)    await pauseCampaign(campaign.id);
      if (action === CAMPAIGN_ACTIONS.DELETE)   await deleteCampaign(campaign.id);
      onCampaignsChanged(); // Refresh the list in the parent
    } catch (err) {
      alert(err.message || `Failed to ${action} campaign`);
    } finally {
      setActionLoadingId(null);
    }
  }

  // ── Helpers ───────────────────────────────────────────────────────────────

  /** Returns the browser confirm dialog message for each action. */
  function getConfirmMessage(action, name) {
    if (action === CAMPAIGN_ACTIONS.DELETE)
      return `Delete campaign "${name}"? This cannot be undone and will remove all email sequences.`;
    if (action === CAMPAIGN_ACTIONS.ACTIVATE)
      return `Activate "${name}"? This will start sending emails.`;
    return `Pause "${name}"?`;
  }

  /** Returns a formatted date string. */
  function formatDate(dateStr) {
    if (!dateStr) return '—';
    return new Date(dateStr).toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' });
  }

  // ── Skeleton Rows (Loading State) ─────────────────────────────────────────
  if (loading) {
    return (
      <div className="outreach-table-wrap">
        <table className="outreach-table">
          <thead>
            <tr>
              <th>Campaign</th>
              <th>Status</th>
              <th>Created</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            {Array.from({ length: 5 }).map((_, i) => (
              <tr key={i} className="outreach-skeleton-row">
                {[220, 90, 100, 120].map((w, j) => (
                  <td key={j}><div className="outreach-skeleton-cell" style={{ width: w }} /></td>
                ))}
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    );
  }

  // ── Empty State ───────────────────────────────────────────────────────────
  if (campaigns.length === 0) {
    return (
      <div className="outreach-table-wrap">
        <div className="outreach-empty">
          <div className="outreach-empty-icon">📢</div>
          <div className="outreach-empty-title">No campaigns yet</div>
          <div className="outreach-empty-sub">Click &quot;+ New Campaign&quot; to create your first outreach campaign</div>
        </div>
      </div>
    );
  }

  // ── Table ─────────────────────────────────────────────────────────────────
  return (
    <div className="outreach-table-wrap">
      <table className="outreach-table">
        <thead>
          <tr>
            <th>Campaign</th>
            <th>Status</th>
            <th>Created</th>
            <th>Actions</th>
          </tr>
        </thead>
        <tbody>
          {campaigns.map(campaign => {
            const statusCfg = CAMPAIGN_STATUS_CONFIG[campaign.status] || { label: campaign.status, type: 'neutral', emoji: '❓' };
            const isThisLoading = actionLoadingId === campaign.id;

            return (
              <tr key={campaign.id}>
                {/* Campaign name + date */}
                <td>
                  <div className="campaign-name-cell">
                    <span className="campaign-name">{campaign.name}</span>
                    <span className="campaign-date">Created {formatDate(campaign.created_at)}</span>
                  </div>
                </td>

                {/* Status badge */}
                <td>
                  <StatusBadge
                    status={campaign.status || CAMPAIGN_STATUS.DRAFT}
                    customLabel={`${statusCfg.emoji} ${statusCfg.label}`}
                    customType={statusCfg.type}
                  />
                </td>

                {/* Created date */}
                <td style={{ color: '#94A3B8', fontSize: 12 }}>{formatDate(campaign.created_at)}</td>

                {/* Action buttons — shown based on the current status */}
                <td>
                  <div className="campaign-actions">
                    {/* DRAFT or PAUSED → show "Launch" button */}
                    {(campaign.status === CAMPAIGN_STATUS.DRAFT || campaign.status === CAMPAIGN_STATUS.PAUSED) && (
                      <button
                        className="outreach-btn outreach-btn-primary"
                        style={{ padding: '6px 12px', fontSize: 12 }}
                        onClick={() => handleAction(CAMPAIGN_ACTIONS.ACTIVATE, campaign)}
                        disabled={isThisLoading}
                      >
                        {isThisLoading ? '...' : '🚀 Launch'}
                      </button>
                    )}

                    {/* ACTIVE → show "Pause" button */}
                    {campaign.status === CAMPAIGN_STATUS.ACTIVE && (
                      <button
                        className="outreach-btn outreach-btn-ghost"
                        style={{ padding: '6px 12px', fontSize: 12 }}
                        onClick={() => handleAction(CAMPAIGN_ACTIONS.PAUSE, campaign)}
                        disabled={isThisLoading}
                      >
                        {isThisLoading ? '...' : '⏸ Pause'}
                      </button>
                    )}

                    {/* Always show Delete, unless it's already loading */}
                    <button
                      className="outreach-btn outreach-btn-danger"
                      style={{ padding: '6px 12px', fontSize: 12 }}
                      onClick={() => handleAction(CAMPAIGN_ACTIONS.DELETE, campaign)}
                      disabled={isThisLoading}
                    >
                      🗑️
                    </button>
                  </div>
                </td>
              </tr>
            );
          })}
        </tbody>
      </table>
    </div>
  );
}

CampaignList.propTypes = {
  campaigns: PropTypes.arrayOf(PropTypes.object).isRequired,
  loading: PropTypes.bool.isRequired,
  onCampaignsChanged: PropTypes.func.isRequired,
};

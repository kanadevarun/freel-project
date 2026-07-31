import { useState } from 'react';
import PropTypes from 'prop-types';
import { updateLead, deleteLead } from '../../../services/leadsService';
import StatusBadge from '../../../components/dashboard/StatusBadge';
import { LEAD_STATUS } from './LeadsPage';

/**
 * AIScoreBadge — Shows the AI score in a color-coded pill.
 * Green for ≥70, amber for ≥40, red for <40, grey if not scored yet.
 */
function AIScoreBadge({ score }) {
  if (score == null) return <span className="ai-score-badge ai-score-none">Not scored</span>;
  const cls = score >= 70 ? 'ai-score-high' : score >= 40 ? 'ai-score-medium' : 'ai-score-low';
  return <span className={`ai-score-badge ${cls}`}>⚡ {score}/100</span>;
}

/**
 * LeadDetailDrawer — A slide-in panel showing full details about a lead.
 *
 * Simple meaning: When you click a row in the table, this animated side panel
 * slides in from the right. It shows:
 *   - The AI score and research report (in a dark "AI Panel")
 *   - Company details (industry, revenue, etc.)
 *   - Action buttons (Convert to RFQ, Email, Delete)
 *
 * @param {{ lead: Lead|null, onClose: () => void, onLeadUpdated: () => void }}
 */
export default function LeadDetailDrawer({ lead, onClose, onLeadUpdated }) {
  const [deleting, setDeleting] = useState(false);

  if (!lead) return null;

  // Derive a 2-letter avatar from company name
  const avatar = lead.company_name
    ? lead.company_name.slice(0, 2).toUpperCase()
    : '??';

  // Figure out the score color for the progress bar
  const score = lead.ai_score;
  const scoreBarColor = score >= 70 ? '#22C55E' : score >= 40 ? '#F59E0B' : '#EF4444';

  async function handleDelete() {
    if (!confirm(`Delete lead for ${lead.company_name}? This cannot be undone.`)) return;
    setDeleting(true);
    try {
      await deleteLead(lead.id);
      onLeadUpdated();
      onClose();
    } catch (e) {
      alert(e.message || 'Failed to delete lead');
    } finally {
      setDeleting(false);
    }
  }

  // Map status to StatusBadge variants using LEAD_STATUS constants (not raw strings)
  const statusMap = {
    [LEAD_STATUS.NEW]:       { label: 'New',       type: 'info' },
    [LEAD_STATUS.QUALIFIED]: { label: 'Qualified', type: 'success' },
    [LEAD_STATUS.REJECTED]:  { label: 'Rejected',  type: 'danger' },
    [LEAD_STATUS.CONVERTED]: { label: 'Converted', type: 'primary' },
  };
  const statusConfig = statusMap[lead.status] || { label: lead.status, type: 'neutral' };

  return (
    <>
      {/* Background overlay — clicking it closes the drawer */}
      <div className="leads-drawer-overlay" onClick={onClose} />

      {/* The drawer panel itself */}
      <div className="leads-drawer">

        {/* Header */}
        <div className="drawer-header">
          <div className="lead-avatar" style={{ width: 44, height: 44, fontSize: 16 }}>{avatar}</div>
          <div className="drawer-header-info">
            <div className="drawer-company-name">{lead.company_name}</div>
            <div className="drawer-contact">
              {lead.contact_name && `${lead.contact_name} • `}
              {lead.email && <a href={`mailto:${lead.email}`} style={{ color: '#00BFA5' }}>{lead.email}</a>}
            </div>
            <div style={{ marginTop: 6 }}>
              <StatusBadge status={lead.status || 'NEW'} customLabel={statusConfig.label} customType={statusConfig.type} />
            </div>
          </div>
          <button className="drawer-close-btn" onClick={onClose} aria-label="Close drawer">✕</button>
        </div>

        {/* Scrollable body */}
        <div className="drawer-body">

          {/* ─── AI Score Panel ─── */}
          <div className="ai-panel">
            <div className="ai-panel-header">
              <span>🤖</span>
              <span>AI Analysis</span>
            </div>

            {score != null ? (
              <>
                <div className="ai-score-ring-wrap">
                  <div>
                    <div className="ai-score-number">{score}</div>
                    <div className="ai-score-label">/ 100 lead score</div>
                  </div>
                  <div className="ai-score-bar-wrap">
                    <div style={{ fontSize: 11, color: 'rgba(255,255,255,0.5)', marginBottom: 6 }}>Quality Score</div>
                    <div className="ai-score-bar-track">
                      <div
                        className="ai-score-bar-fill"
                        style={{
                          width: `${score}%`,
                          background: `linear-gradient(90deg, ${scoreBarColor}, #5A4FCF)`,
                        }}
                      />
                    </div>
                    <div style={{ fontSize: 11, color: 'rgba(255,255,255,0.4)', marginTop: 6 }}>
                      {score >= 70 ? '🟢 High Priority' : score >= 40 ? '🟡 Medium Priority' : '🔴 Low Priority'}
                    </div>
                  </div>
                </div>
                {lead.ai_research_report && (
                  <div className="ai-report-text">{lead.ai_research_report}</div>
                )}
              </>
            ) : (
              <div style={{ color: 'rgba(255,255,255,0.5)', fontSize: 13, textAlign: 'center', padding: '12px 0' }}>
                ⏳ AI analysis is running in the background...
              </div>
            )}
          </div>

          {/* ─── Company Details ─── */}
          <div>
            <div className="drawer-section-title">Lead Details</div>
            <div className="drawer-detail-grid">
              <div className="drawer-detail-item">
                <div className="drawer-detail-item-label">Source</div>
                <div className="drawer-detail-item-value">{lead.source || '—'}</div>
              </div>
              <div className="drawer-detail-item">
                <div className="drawer-detail-item-label">Status</div>
                <div className="drawer-detail-item-value">{lead.status || 'NEW'}</div>
              </div>
              <div className="drawer-detail-item">
                <div className="drawer-detail-item-label">Contact</div>
                <div className="drawer-detail-item-value">{lead.contact_name || '—'}</div>
              </div>
              <div className="drawer-detail-item">
                <div className="drawer-detail-item-label">Added</div>
                <div className="drawer-detail-item-value">
                  {lead.created_at ? new Date(lead.created_at).toLocaleDateString() : '—'}
                </div>
              </div>
            </div>
          </div>

          {/* ─── Notes ─── */}
          {lead.notes && (
            <div>
              <div className="drawer-section-title">Notes</div>
              <div style={{ fontSize: 13.5, color: '#334155', lineHeight: 1.7, background: '#F8FAFC', borderRadius: 10, padding: '12px 14px' }}>
                {lead.notes}
              </div>
            </div>
          )}

          {/* ─── Timeline placeholder ─── */}
          <div>
            <div className="drawer-section-title">Activity Timeline</div>
            <div style={{ textAlign: 'center', padding: '20px 0', color: '#94A3B8', fontSize: 13 }}>
              <div style={{ fontSize: 24, marginBottom: 8 }}>🕐</div>
              Timeline events will appear here once the Activity API is connected.
            </div>
          </div>

        </div>

        {/* Footer action bar */}
        <div className="drawer-footer">
          <button className="leads-btn leads-btn-ghost" style={{ flex: 1, justifyContent: 'center' }}>
            ✉️ Email
          </button>
          <button className="leads-btn leads-btn-primary" style={{ flex: 1, justifyContent: 'center' }}>
            📋 Create RFQ
          </button>
          <button
            className="leads-btn leads-btn-danger"
            onClick={handleDelete}
            disabled={deleting}
            style={{ padding: '8px 12px' }}
          >
            {deleting ? '...' : '🗑️'}
          </button>
        </div>
      </div>
    </>
  );
}

LeadDetailDrawer.propTypes = {
  lead: PropTypes.object,
  onClose: PropTypes.func.isRequired,
  onLeadUpdated: PropTypes.func.isRequired,
};

AIScoreBadge.propTypes = {
  score: PropTypes.number,
};

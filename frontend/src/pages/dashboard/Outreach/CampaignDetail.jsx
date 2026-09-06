import { useState, useEffect, useCallback } from 'react';
import PropTypes from 'prop-types';
import {
  getCampaign,
  activateCampaign,
  pauseCampaign,
  deleteCampaign,
  getCampaignAnalytics,
  getCampaignInsights,
  getCampaignRecipients,
  getCampaignActivity,
} from '../../../services/outreachService';
import CampaignSequence from './CampaignSequence';
import CampaignAudience from './CampaignAudience';
import './OutreachPage.css';

const DETAIL_TABS = {
  OVERVIEW: 'overview',
  RECIPIENTS: 'recipients',
  SEQUENCE: 'sequence',
  PERFORMANCE: 'performance',
  ACTIVITY: 'activity',
};

export default function CampaignDetail({ campaignId, onBack, onCampaignUpdated, onSelectProspect }) {
  const [campaign, setCampaign] = useState(null);
  const [activeTab, setActiveTab] = useState(DETAIL_TABS.OVERVIEW);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  // Loaded states
  const [analytics, setAnalytics] = useState(null);
  const [recipients, setRecipients] = useState([]);
  const [activities, setActivities] = useState([]);
  const [insights, setInsights] = useState([]);

  const fetchCampaignData = useCallback(async () => {
    setLoading(true);
    setError('');
    try {
      const data = await getCampaign(campaignId);
      setCampaign(data);

      const stats = await getCampaignAnalytics(campaignId);
      setAnalytics(stats);

      const recipientsList = await getCampaignRecipients(campaignId);
      setRecipients(recipientsList || []);

      const activityList = await getCampaignActivity(campaignId);
      setActivities(activityList || []);

      const insightsList = await getCampaignInsights(campaignId);
      setInsights(insightsList?.insights || insightsList || []);
    } catch (err) {
      console.error('Failed to load campaign detail data:', err);
      setError('Failed to load campaign details.');
    } finally {
      setLoading(false);
    }
  }, [campaignId]);

  useEffect(() => {
    fetchCampaignData();
  }, [fetchCampaignData]);

  async function handleActivate() {
    try {
      await activateCampaign(campaignId);
      await fetchCampaignData();
      if (onCampaignUpdated) onCampaignUpdated();
    } catch (err) {
      setError(err.message || 'Failed to activate campaign.');
    }
  }

  async function handlePause() {
    try {
      await pauseCampaign(campaignId);
      await fetchCampaignData();
      if (onCampaignUpdated) onCampaignUpdated();
    } catch (err) {
      setError(err.message || 'Failed to pause campaign.');
    }
  }

  async function handleDelete() {
    if (!window.confirm('Are you sure you want to permanently delete this campaign? This action cannot be undone.')) {
      return;
    }
    try {
      await deleteCampaign(campaignId);
      if (onCampaignUpdated) onCampaignUpdated();
      onBack();
    } catch (err) {
      setError(err.message || 'Failed to delete campaign.');
    }
  }

  function formatDateTime(dateStr) {
    if (!dateStr) return '—';
    const date = new Date(dateStr);
    return date.toLocaleString('en-US', {
      month: 'short',
      day: 'numeric',
      year: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    });
  }

  function getTypeIcon(type) {
    switch ((type || '').toUpperCase()) {
      case 'EMAIL':
        return '✉️';
      case 'CALL':
        return '📞';
      case 'MEETING':
        return '👥';
      case 'FOLLOW_UP':
      default:
        return '🔔';
    }
  }

  if (loading && !campaign) {
    return (
      <div className="outreach-loading-box">
        <div className="outreach-spinner" />
        <p>Loading campaign details...</p>
      </div>
    );
  }

  if (error && !campaign) {
    return (
      <div className="outreach-error-box">
        <p>{error}</p>
        <button className="outreach-btn outreach-btn-secondary" onClick={onBack}>
          Back to Campaigns
        </button>
      </div>
    );
  }

  const statusLabel = campaign?.status || 'DRAFT';

  return (
    <div className="campaign-detail-page">
      {/* ─── Header ─── */}
      <div className="campaign-detail-header">
        <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
          <button className="outreach-back-btn" onClick={onBack} title="Back to Campaigns">
            <svg width="20" height="20" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M10 19l-7-7m0 0l7-7m-7 7h18" />
            </svg>
          </button>
          <div>
            <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
              <h2 className="campaign-title">{campaign?.name}</h2>
              <span className={`campaign-status-tag ${statusLabel.toLowerCase()}`}>
                {statusLabel}
              </span>
            </div>
            <p className="campaign-subtitle">
              Objective: Engage prospects and automate email sequences for client acquisition.
            </p>
          </div>
        </div>

        <div className="campaign-actions-row">
          {statusLabel === 'DRAFT' || statusLabel === 'PAUSED' ? (
            <button className="outreach-btn outreach-btn-primary" onClick={handleActivate}>
              Launch Campaign
            </button>
          ) : statusLabel === 'ACTIVE' ? (
            <button className="outreach-btn outreach-btn-warning" onClick={handlePause}>
              Pause Campaign
            </button>
          ) : null}
          <button className="outreach-btn outreach-btn-danger" onClick={handleDelete}>
            Delete
          </button>
        </div>
      </div>

      {error && (
        <div className="outreach-error-banner" style={{ margin: '16px 0 0 0' }}>
          {error}
        </div>
      )}

      {/* ─── Tabs Bar ─── */}
      <div className="outreach-detail-tabs-bar">
        <button
          className={`outreach-detail-tab-btn ${activeTab === DETAIL_TABS.OVERVIEW ? 'active' : ''}`}
          onClick={() => setActiveTab(DETAIL_TABS.OVERVIEW)}
        >
          📊 Overview
        </button>
        <button
          className={`outreach-detail-tab-btn ${activeTab === DETAIL_TABS.RECIPIENTS ? 'active' : ''}`}
          onClick={() => setActiveTab(DETAIL_TABS.RECIPIENTS)}
        >
          👥 Recipients ({recipients.length})
        </button>
        <button
          className={`outreach-detail-tab-btn ${activeTab === DETAIL_TABS.SEQUENCE ? 'active' : ''}`}
          onClick={() => setActiveTab(DETAIL_TABS.SEQUENCE)}
        >
          ⛓️ Sequence steps
        </button>
        <button
          className={`outreach-detail-tab-btn ${activeTab === DETAIL_TABS.PERFORMANCE ? 'active' : ''}`}
          onClick={() => setActiveTab(DETAIL_TABS.PERFORMANCE)}
        >
          🎯 Performance Funnel
        </button>
        <button
          className={`outreach-detail-tab-btn ${activeTab === DETAIL_TABS.ACTIVITY ? 'active' : ''}`}
          onClick={() => setActiveTab(DETAIL_TABS.ACTIVITY)}
        >
          ⏳ Campaign Activity
        </button>
      </div>

      {/* ─── Tab Content Workspace ─── */}
      <div className="outreach-detail-content-area">
        
        {/* TAB 1: OVERVIEW */}
        {activeTab === DETAIL_TABS.OVERVIEW && (
          <div className="outreach-overview-tab">
            <div className="outreach-card-header">
              <h3 className="outreach-section-title">Campaign Overview</h3>
              <p className="outreach-section-subtitle">Core campaign statistics and launch profile.</p>
            </div>

            {/* Campaign Details Info */}
            <div className="campaign-metadata-grid" style={{ display: 'grid', gridTemplateColumns: 'repeat(4, 1fr)', gap: 16, marginBottom: 24, padding: 16, background: '#F8FAFC', borderRadius: 12, border: '1px solid #E2E8F0' }}>
              <div>
                <span className="metadata-label" style={{ fontSize: 11, fontWeight: 700, color: '#64748B', textTransform: 'uppercase' }}>Status</span>
                <div style={{ marginTop: 4 }}><span className={`campaign-status-tag ${statusLabel.toLowerCase()}`}>{statusLabel}</span></div>
              </div>
              <div>
                <span className="metadata-label" style={{ fontSize: 11, fontWeight: 700, color: '#64748B', textTransform: 'uppercase' }}>Audience Size</span>
                <div style={{ marginTop: 4, fontWeight: 700, fontSize: 15, color: '#0F172A' }}>{analytics?.total_recipients || 0} Leads</div>
              </div>
              <div>
                <span className="metadata-label" style={{ fontSize: 11, fontWeight: 700, color: '#64748B', textTransform: 'uppercase' }}>Created Date</span>
                <div style={{ marginTop: 4, fontWeight: 600, fontSize: 14, color: '#334155' }}>{campaign?.created_at ? new Date(campaign.created_at).toLocaleDateString() : '—'}</div>
              </div>
              <div>
                <span className="metadata-label" style={{ fontSize: 11, fontWeight: 700, color: '#64748B', textTransform: 'uppercase' }}>Last Updated</span>
                <div style={{ marginTop: 4, fontWeight: 600, fontSize: 14, color: '#334155' }}>{campaign?.updated_at ? new Date(campaign.updated_at).toLocaleDateString() : '—'}</div>
              </div>
            </div>

            {/* KPI Performance Cards */}
            <h4 style={{ fontSize: 14, fontWeight: 750, color: '#0F172A', marginBottom: 14 }}>Outreach Delivery Analytics</h4>
            <div className="outreach-stats-row" style={{ display: 'grid', gridTemplateColumns: 'repeat(6, 1fr)', gap: 12, marginBottom: 24 }}>
              <div className="outreach-stat-card">
                <div>
                  <div className="outreach-stat-value">{analytics?.emails_sent ?? 0}</div>
                  <div className="outreach-stat-label">Sent</div>
                </div>
              </div>
              <div className="outreach-stat-card" style={{ background: '#F8FAFC', opacity: 0.85 }} title="Delivery tracking is not supported by mailbox provider">
                <div>
                  <div className="outreach-stat-value" style={{ fontSize: 13, color: '#64748B' }}>Not Available</div>
                  <div className="outreach-stat-label">Delivered</div>
                </div>
              </div>
              <div className="outreach-stat-card" style={{ background: '#F8FAFC', opacity: 0.85 }} title="Open tracking is not supported by mailbox provider">
                <div>
                  <div className="outreach-stat-value" style={{ fontSize: 13, color: '#64748B' }}>Not Available</div>
                  <div className="outreach-stat-label">Opened</div>
                </div>
              </div>
              <div className="outreach-stat-card" style={{ background: '#F8FAFC', opacity: 0.85 }} title="Link clicks tracking is not supported by mailbox provider">
                <div>
                  <div className="outreach-stat-value" style={{ fontSize: 13, color: '#64748B' }}>Not Available</div>
                  <div className="outreach-stat-label">Clicked</div>
                </div>
              </div>
              <div className="outreach-stat-card" style={{ background: '#F8FAFC', opacity: 0.85 }} title="Reply tracking is not supported by mailbox provider">
                <div>
                  <div className="outreach-stat-value" style={{ fontSize: 13, color: '#64748B' }}>Not Available</div>
                  <div className="outreach-stat-label">Replied</div>
                </div>
              </div>
              <div className="outreach-stat-card" style={{ background: '#F8FAFC', opacity: 0.85 }} title="Bounce tracking is not supported by mailbox provider">
                <div>
                  <div className="outreach-stat-value" style={{ fontSize: 13, color: '#64748B' }}>Not Available</div>
                  <div className="outreach-stat-label">Bounced</div>
                </div>
              </div>
            </div>

            <div style={{ display: 'flex', alignItems: 'center', gap: 6, background: '#EFF6FF', border: '1px solid #BFDBFE', padding: '10px 14px', borderRadius: 8, fontSize: 12.5, color: '#1E40AF', fontWeight: 500 }}>
              <span>ℹ️</span>
              <span>Certain delivery, open, click, and bounce metrics are <strong>Not Available</strong> because LogisticsHQ respects prospect privacy guidelines and relies on direct response tracking rather than pixel injections.</span>
            </div>
          </div>
        )}

        {/* TAB 2: RECIPIENTS / PROSPECTS */}
        {activeTab === DETAIL_TABS.RECIPIENTS && (
          <div className="outreach-recipients-tab">
            <div className="outreach-card-header" style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 16 }}>
              <div>
                <h3 className="outreach-section-title">Campaign Prospects</h3>
                <p className="outreach-section-subtitle">Manage and track target audience leads assigned to this campaign.</p>
              </div>
              <button 
                className="outreach-btn outreach-btn-primary" 
                style={{ padding: '8px 14px', fontSize: 12.5 }}
                onClick={() => setActiveTab(DETAIL_TABS.SEQUENCE)}
              >
                Configure Audience
              </button>
            </div>

            {recipients.length === 0 ? (
              <div className="outreach-empty-state" style={{ padding: '60px 20px', border: '1px dashed #E2E8F0', borderRadius: 12, textAlign: 'center' }}>
                <div style={{ fontSize: 28, marginBottom: 8 }}>👥</div>
                <h4>No Recipients Added</h4>
                <p style={{ color: '#64748B', fontSize: 13 }}>To start outreach, assign leads to this campaign inside the Audience settings.</p>
              </div>
            ) : (
              <div className="outreach-table-wrapper" style={{ background: '#FFFFFF', border: '1px solid #E2E8F0', borderRadius: 12, overflow: 'hidden' }}>
                <table className="outreach-custom-table" style={{ width: '100%', borderCollapse: 'collapse' }}>
                  <thead>
                    <tr style={{ background: '#F8FAFC', borderBottom: '1px solid #E2E8F0', textAlign: 'left' }}>
                      <th style={{ padding: '12px 16px', fontSize: 12, fontWeight: 700, color: '#475569' }}>Company / Prospect</th>
                      <th style={{ padding: '12px 16px', fontSize: 12, fontWeight: 700, color: '#475569' }}>Email Address</th>
                      <th style={{ padding: '12px 16px', fontSize: 12, fontWeight: 700, color: '#475569' }}>Lead Status</th>
                      <th style={{ padding: '12px 16px', fontSize: 12, fontWeight: 700, color: '#475569' }}>Engagement</th>
                      <th style={{ padding: '12px 16px', fontSize: 12, fontWeight: 700, color: '#475569' }}>Emails Sent</th>
                      <th style={{ padding: '12px 16px', fontSize: 12, fontWeight: 700, color: '#475569' }}>Last Activity</th>
                      <th style={{ padding: '12px 16px', fontSize: 12, fontWeight: 700, color: '#475569' }}>Action</th>
                    </tr>
                  </thead>
                  <tbody>
                    {recipients.map(r => (
                      <tr key={r.lead_id} style={{ borderBottom: '1px solid #F1F5F9' }}>
                        <td style={{ padding: '12px 16px' }}>
                          <div 
                            style={{ fontWeight: 650, color: '#2563EB', fontSize: 13.5, cursor: 'pointer' }}
                            onClick={() => onSelectProspect && onSelectProspect(r.lead_id)}
                          >
                            🏢 {r.company_name}
                          </div>
                          {r.contact_name && <div style={{ fontSize: 12, color: '#64748B', marginTop: 2 }}>👤 {r.contact_name}</div>}
                        </td>
                        <td style={{ padding: '12px 16px', fontSize: 13, color: '#475569' }}>
                          {r.email || <span style={{ color: '#94A3B8' }}>—</span>}
                        </td>
                        <td style={{ padding: '12px 16px' }}>
                          <span className={`lead-status-pill ${r.lead_status?.toLowerCase()}`}>
                            {r.lead_status}
                          </span>
                        </td>
                        <td style={{ padding: '12px 16px' }}>
                          <span className={`activity-status-pill-grad ${r.engagement_status?.toLowerCase()}`} style={{ fontSize: 11, padding: '3px 8px' }}>
                            {r.engagement_status?.replace('_', ' ')}
                          </span>
                        </td>
                        <td style={{ padding: '12px 16px', fontSize: 13.5, fontWeight: 700, color: '#0F172A', textAlign: 'center' }}>
                          {r.emails_sent}
                        </td>
                        <td style={{ padding: '12px 16px' }}>
                          {r.last_activity_at ? (
                            <div>
                              <div style={{ fontSize: 12.5, fontWeight: 600, color: '#334155' }}>{r.last_activity_desc}</div>
                              <div style={{ fontSize: 11, color: '#64748B', marginTop: 2 }}>{formatDateTime(r.last_activity_at)}</div>
                            </div>
                          ) : (
                            <span style={{ color: '#94A3B8', fontSize: 12 }}>No outreach logged</span>
                          )}
                        </td>
                        <td style={{ padding: '12px 16px' }}>
                          <a 
                            href="/dashboard/leads" 
                            className="activity-btn-outline" 
                            style={{ padding: '4px 10px', fontSize: 11, fontWeight: 600, textDecoration: 'none' }}
                          >
                            Open Lead
                          </a>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
          </div>
        )}

        {/* TAB 3: SEQUENCE STEPS */}
        {activeTab === DETAIL_TABS.SEQUENCE && (
          <div style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
            {/* Audience Config Panel */}
            <div style={{ border: '1px solid #E2E8F0', borderRadius: 12, padding: 16, background: '#FFFFFF' }}>
              <h4 style={{ fontSize: 14, fontWeight: 750, color: '#0F172A', marginBottom: 6 }}>Campaign Target Audience</h4>
              <p style={{ fontSize: 12.5, color: '#64748B', marginBottom: 12 }}>Search and assign Leads to receive this automated campaign email sequence.</p>
              <CampaignAudience campaignId={campaignId} />
            </div>

            {/* Sequence Steps Panel */}
            <div style={{ border: '1px solid #E2E8F0', borderRadius: 12, padding: 16, background: '#FFFFFF' }}>
              <h4 style={{ fontSize: 14, fontWeight: 750, color: '#0F172A', marginBottom: 6 }}>Configure Outreach Sequence steps</h4>
              <CampaignSequence campaignId={campaignId} onSequenceChanged={fetchCampaignData} />
            </div>
          </div>
        )}

        {/* TAB 4: PERFORMANCE FUNNEL */}
        {activeTab === DETAIL_TABS.PERFORMANCE && (
          <div className="outreach-analytics-tab">
            <div className="outreach-analytics-grid">
              
              {/* Funnel Box */}
              <div className="outreach-funnel-card">
                <h4 className="card-title">Outreach Conversion Funnel</h4>
                <div className="funnel-container">
                  <div className="funnel-step primary">
                    <div className="funnel-step-meta">
                      <span className="step-name">Audience Size</span>
                      <span className="step-count">{analytics?.total_recipients || 0}</span>
                    </div>
                    <div className="funnel-bar width-100">100%</div>
                  </div>

                  <div className="funnel-arrow">↓</div>

                  <div className="funnel-step contacted">
                    <div className="funnel-step-meta">
                      <span className="step-name">Contacted (Emails Sent)</span>
                      <span className="step-count">{analytics?.emails_sent ?? 0}</span>
                    </div>
                    <div
                      className="funnel-bar"
                      style={{
                        width: `${analytics?.total_recipients > 0 ? (((analytics?.emails_sent ?? 0) / analytics.total_recipients) * 100) : 0}%`,
                        minWidth: '10%',
                      }}
                    >
                      {analytics?.total_recipients > 0 ? Math.round(((analytics?.emails_sent ?? 0) / analytics.total_recipients) * 100) : 0}%
                    </div>
                  </div>

                  <div className="funnel-arrow">↓</div>

                  <div className="funnel-step leads">
                    <div className="funnel-step-meta">
                      <span className="step-name">Leads Generated</span>
                      <span className="step-count">{analytics?.leads_generated || 0}</span>
                    </div>
                    <div
                      className="funnel-bar"
                      style={{
                        width: `${analytics?.total_recipients > 0 ? ((analytics.leads_generated / analytics.total_recipients) * 100) : 0}%`,
                        minWidth: '8%',
                      }}
                    >
                      {analytics?.total_recipients > 0 ? Math.round((analytics.leads_generated / analytics.total_recipients) * 100) : 0}%
                    </div>
                  </div>
                </div>
              </div>

              {/* Insights Box */}
              <div className="outreach-insights-card">
                <h4 className="card-title">Campaign Attention & Insights</h4>
                {insights.length === 0 ? (
                  <div className="insight-empty">
                    <div className="ok-icon">✓</div>
                    <p>All checks passed. No actions required for this campaign.</p>
                  </div>
                ) : (
                  <div className="insights-feed-list">
                    {insights.map((insight, idx) => (
                      <div key={idx} className={`insight-feed-item ${insight.severity.toLowerCase()}`}>
                        <div className="insight-icon">
                          {insight.severity === 'CRITICAL' ? '⚠️' : insight.severity === 'WARNING' ? '⚡' : 'ℹ️'}
                        </div>
                        <div className="insight-body">
                          <h5 className="insight-title">{insight.title}</h5>
                          <p className="insight-desc">{insight.description}</p>
                        </div>
                      </div>
                    ))}
                  </div>
                )}
              </div>
            </div>
          </div>
        )}

        {/* TAB 5: CAMPAIGN ACTIVITY */}
        {activeTab === DETAIL_TABS.ACTIVITY && (
          <div className="outreach-activity-tab">
            <div className="outreach-card-header">
              <h3 className="outreach-section-title">Timeline of chronological events</h3>
              <p className="outreach-section-subtitle">Real-time trace of outreach logs and tasks linked to prospects of this campaign.</p>
            </div>

            {activities.length === 0 ? (
              <div className="outreach-empty-state" style={{ padding: '60px 20px', border: '1px dashed #E2E8F0', borderRadius: 12, textAlign: 'center' }}>
                <div style={{ fontSize: 28, marginBottom: 8 }}>⏳</div>
                <h4>No Activity Logged</h4>
                <p style={{ color: '#64748B', fontSize: 13 }}>Timeline logs will populate here as emails are sent and follow-ups are completed.</p>
              </div>
            ) : (
              <div className="activity-timeline-feed" style={{ position: 'relative', paddingLeft: 20, borderLeft: '2px solid #E2E8F0', marginLeft: 10 }}>
                {activities.map(act => (
                  <div key={act.id} className="timeline-item" style={{ position: 'relative', marginBottom: 24 }}>
                    {/* Timeline Node Circle */}
                    <div style={{ position: 'absolute', left: -29, top: 2, background: '#FFFFFF', borderRadius: '50%', padding: 2, display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
                      <span style={{ fontSize: 16 }}>{getTypeIcon(act.activity_type)}</span>
                    </div>

                    <div style={{ background: '#FFFFFF', padding: 14, borderRadius: 8, border: '1px solid #E2E8F0', boxShadow: '0 1px 3px rgba(0,0,0,0.02)' }}>
                      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
                        <div>
                          <span style={{ fontWeight: 700, color: '#0F172A', fontSize: 13.5 }}>{act.lead_company_name || 'Prospect'}</span>
                          <span style={{ fontSize: 12, color: '#64748B', marginLeft: 8 }}>({act.activity_type})</span>
                        </div>
                        <span style={{ fontSize: 11.5, color: '#64748B', fontWeight: 550 }}>{formatDateTime(act.created_at)}</span>
                      </div>
                      <div style={{ fontWeight: 600, fontSize: 13, color: '#1E293B', marginTop: 6 }}>{act.subject}</div>
                      {act.description && <div style={{ fontSize: 12, color: '#475569', marginTop: 4, background: '#F8FAFC', padding: 8, borderRadius: 6 }}>{act.description}</div>}
                      
                      <div style={{ display: 'flex', alignItems: 'center', gap: 10, marginTop: 8 }}>
                        <span className={`activity-status-pill-grad ${act.status?.toLowerCase()}`} style={{ fontSize: 10, padding: '2px 6px' }}>
                          {act.status}
                        </span>
                        <span style={{ fontSize: 11, color: '#64748B' }}>Owner: <strong>{act.creator_name || 'System'}</strong></span>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>
        )}

      </div>
    </div>
  );
}

CampaignDetail.propTypes = {
  campaignId: PropTypes.number.isRequired,
  onBack: PropTypes.func.isRequired,
  onCampaignUpdated: PropTypes.func,
  onSelectProspect: PropTypes.func,
};
import { useState, useEffect, useCallback } from 'react';
import PropTypes from 'prop-types';
import {
  getProspectDetail,
  pauseProspect,
  resumeProspect,
  stopProspect,
  updateProspect,
  completeFollowUp,
  cancelFollowUp
} from '../../../services/outreachService';
import { User, Building2, Mail, Phone, Clock, Sparkles, Folder, Calendar, AlertCircle, ArrowLeft, ExternalLink, Activity } from 'lucide-react';
import './OutreachPage.css';

const DETAIL_TABS = {
  OVERVIEW: 'overview',
  CAMPAIGNS: 'campaigns',
  SEQUENCE: 'sequence',
  FOLLOW_UPS: 'follow_ups',
  ACTIVITY: 'activity',
};

export default function ProspectDetail({ leadId, onBack, onScheduleFollowUp, onSelectCampaign }) {
  const [detail, setDetail] = useState(null);
  const [activeTab, setActiveTab] = useState(DETAIL_TABS.OVERVIEW);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  const fetchDetail = useCallback(async () => {
    setLoading(true);
    setError('');
    try {
      const data = await getProspectDetail(leadId);
      setDetail(data);
    } catch (err) {
      console.error('Failed to load prospect detail:', err);
      setError('Failed to fetch prospect details.');
    } finally {
      setLoading(false);
    }
  }, [leadId]);

  useEffect(() => {
    fetchDetail();
  }, [fetchDetail]);

  async function handleStatusAction(action) {
    if (!detail?.prospect?.campaign_id) return;
    try {
      const campaignId = detail.prospect.campaign_id;
      if (action === 'PAUSE') {
        await pauseProspect(leadId, campaignId);
      } else if (action === 'RESUME') {
        await resumeProspect(leadId, campaignId);
      } else if (action === 'STOP') {
        await stopProspect(leadId, campaignId);
      } else if (action === 'DNC') {
        await updateProspect(leadId, campaignId, 'DO_NOT_CONTACT', detail.prospect.current_step);
      } else if (action === 'BOUNCE') {
        await updateProspect(leadId, campaignId, 'BOUNCED', detail.prospect.current_step);
      }
      await fetchDetail();
    } catch (err) {
      setError(err.message || 'Action failed');
    }
  }

  async function handleCompleteFollowUp(id) {
    try {
      await completeFollowUp(id);
      await fetchDetail();
    } catch (err) {
      setError(err.message || 'Failed to complete follow-up');
    }
  }

  async function handleCancelFollowUp(id) {
    if (!window.confirm('Are you sure you want to cancel this follow-up?')) return;
    try {
      await cancelFollowUp(id);
      await fetchDetail();
    } catch (err) {
      setError(err.message || 'Failed to cancel follow-up');
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

  if (loading && !detail) {
    return (
      <div className="outreach-loading-box">
        <div className="outreach-spinner" />
        <p>Loading prospect profile...</p>
      </div>
    );
  }

  if (error && !detail) {
    return (
      <div className="outreach-error-box">
        <p>{error}</p>
        <button className="outreach-btn outreach-btn-secondary" onClick={onBack}>
          Back to Directory
        </button>
      </div>
    );
  }

  const p = detail?.prospect;

  return (
    <div className="campaign-detail-page">
      {/* ─── Header ─── */}
      <div className="campaign-detail-header">
        <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
          <button className="outreach-back-btn" onClick={onBack} title="Back to Prospects">
            <svg width="20" height="20" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M10 19l-7-7m0 0l7-7m-7 7h18" />
            </svg>
          </button>
          <div>
            <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
              <h2 className="campaign-title">🏢 {p?.company_name}</h2>
              <span className={`campaign-status-tag ${p?.status?.toLowerCase()}`}>
                {p?.status?.replace(/_/g, ' ')}
              </span>
            </div>
            {p?.contact_name && <p className="campaign-subtitle">Contact: <strong>{p?.contact_name}</strong> · {p?.email} {p?.phone && `· 📞 ${p.phone}`}</p>}
          </div>
        </div>

        <div className="campaign-actions-row">
          <button 
            className="outreach-btn outreach-btn-primary"
            onClick={(e) => onScheduleFollowUp(e, p)}
          >
            Schedule Follow-up
          </button>
          {p?.status === 'ACTIVE' ? (
            <button className="outreach-btn outreach-btn-warning" onClick={() => handleStatusAction('PAUSE')}>
              Pause Outreach
            </button>
          ) : p?.status === 'PAUSED' ? (
            <button className="outreach-btn outreach-btn-primary" onClick={() => handleStatusAction('RESUME')}>
              Resume Outreach
            </button>
          ) : null}
          {p?.status !== 'COMPLETED' && (
            <button className="outreach-btn outreach-btn-danger" onClick={() => handleStatusAction('STOP')}>
              Stop Outreach
            </button>
          )}
        </div>
      </div>

      {error && (
        <div className="outreach-error-banner" style={{ margin: '16px 0 0 0' }}>
          {error}
        </div>
      )}

      {/* ─── Tabs Bar ─── */}
      <div className="outreach-detail-tabs-bar" style={{ display: 'flex', gap: 4 }}>
        <button
          className={`outreach-detail-tab-btn ${activeTab === DETAIL_TABS.OVERVIEW ? 'active' : ''}`}
          onClick={() => setActiveTab(DETAIL_TABS.OVERVIEW)}
          style={{ display: 'inline-flex', alignItems: 'center' }}
        >
          <User size={13} style={{ marginRight: 5 }} /> Profile Overview
        </button>
        <button
          className={`outreach-detail-tab-btn ${activeTab === DETAIL_TABS.CAMPAIGNS ? 'active' : ''}`}
          onClick={() => setActiveTab(DETAIL_TABS.CAMPAIGNS)}
          style={{ display: 'inline-flex', alignItems: 'center' }}
        >
          <Folder size={13} style={{ marginRight: 5 }} /> Enrolled Campaigns ({detail?.campaigns?.length || 0})
        </button>
        <button
          className={`outreach-detail-tab-btn ${activeTab === DETAIL_TABS.SEQUENCE ? 'active' : ''}`}
          onClick={() => setActiveTab(DETAIL_TABS.SEQUENCE)}
          style={{ display: 'inline-flex', alignItems: 'center' }}
        >
          <Activity size={13} style={{ marginRight: 5 }} /> Sequence Progress
        </button>
        <button
          className={`outreach-detail-tab-btn ${activeTab === DETAIL_TABS.FOLLOW_UPS ? 'active' : ''}`}
          onClick={() => setActiveTab(DETAIL_TABS.FOLLOW_UPS)}
          style={{ display: 'inline-flex', alignItems: 'center' }}
        >
          <Clock size={13} style={{ marginRight: 5 }} /> Follow-ups Queue ({detail?.followUps?.filter(f => f.status === 'PENDING').length || 0})
        </button>
        <button
          className={`outreach-detail-tab-btn ${activeTab === DETAIL_TABS.ACTIVITY ? 'active' : ''}`}
          onClick={() => setActiveTab(DETAIL_TABS.ACTIVITY)}
          style={{ display: 'inline-flex', alignItems: 'center' }}
        >
          <Clock size={13} style={{ marginRight: 5 }} /> Activity Feed
        </button>
      </div>

      {/* ─── Tab Content Workspace ─── */}
      <div className="outreach-detail-content-area">

        {/* TAB 1: OVERVIEW */}
        {activeTab === DETAIL_TABS.OVERVIEW && (
          <div className="outreach-overview-tab">
            <div className="outreach-card-header">
              <h3 className="outreach-section-title">Prospect Profile Overview</h3>
              <p className="outreach-section-subtitle">Core parameters, metadata status, and pipeline attribution.</p>
            </div>

            <div className="campaign-metadata-grid" style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: 16, marginBottom: 24, padding: 16, background: '#F8FAFC', borderRadius: 12, border: '1px solid #E2E8F0' }}>
              <div>
                <span className="metadata-label" style={{ fontSize: 11, fontWeight: 700, color: '#64748B', textTransform: 'uppercase' }}>Current Campaign</span>
                <div style={{ marginTop: 4, fontWeight: 700, fontSize: 14, color: '#3b82f6' }}>
                  {p?.campaign_name ? `📢 ${p.campaign_name}` : 'Unassigned'}
                </div>
              </div>
              <div>
                <span className="metadata-label" style={{ fontSize: 11, fontWeight: 700, color: '#64748B', textTransform: 'uppercase' }}>Sequence Position</span>
                <div style={{ marginTop: 4, fontWeight: 700, fontSize: 14, color: '#334155' }}>Step {p?.current_step}</div>
              </div>
              <div>
                <span className="metadata-label" style={{ fontSize: 11, fontWeight: 700, color: '#64748B', textTransform: 'uppercase' }}>Engagement Status</span>
                <div style={{ marginTop: 4 }}>
                  <span className={`activity-status-pill-grad ${p?.engagement_status?.toLowerCase()}`} style={{ fontSize: 11, padding: '3px 8px' }}>
                    {p?.engagement_status?.replace(/_/g, ' ')}
                  </span>
                </div>
              </div>
              <div>
                <span className="metadata-label" style={{ fontSize: 11, fontWeight: 700, color: '#64748B', textTransform: 'uppercase' }}>Associated Lead Status</span>
                <div style={{ marginTop: 4 }}>
                  <span className={`lead-status-pill ${p?.lead_status?.toLowerCase()}`}>
                    {p?.lead_status}
                  </span>
                </div>
              </div>
              <div>
                <span className="metadata-label" style={{ fontSize: 11, fontWeight: 700, color: '#64748B', textTransform: 'uppercase' }}>Assigned Owner</span>
                <div style={{ marginTop: 4, fontWeight: 600, fontSize: 13, color: '#475569' }}>👤 {p?.owner_name || 'System Assigned'}</div>
              </div>
              <div>
                <span className="metadata-label" style={{ fontSize: 11, fontWeight: 700, color: '#64748B', textTransform: 'uppercase' }}>Next Scheduled Follow-up</span>
                <div style={{ marginTop: 4, fontWeight: 600, fontSize: 13, color: '#475569' }}>
                  {p?.next_scheduled_at ? formatDateTime(p.next_scheduled_at) : 'None Scheduled'}
                </div>
              </div>
            </div>

            {/* Manual Controls & DNC/Bounce Handling */}
            <div style={{ background: '#FFF5F5', border: '1px solid #FED7D7', borderRadius: 12, padding: 16, marginTop: 16 }}>
              <h4 style={{ fontSize: 13.5, fontWeight: 750, color: '#9B2C2C', marginBottom: 4 }}>DNC & Bounce Overrides</h4>
              <p style={{ fontSize: 12.5, color: '#C53030', marginBottom: 12 }}>
                Manually flags this prospect as unsubscribed or bounced. Bypasses further sequence iterations while preserving lead history.
              </p>
              <div style={{ display: 'flex', gap: 10 }}>
                <button className="outreach-btn outreach-btn-danger" style={{ padding: '6px 12px', fontSize: 12 }} onClick={() => handleStatusAction('DNC')}>
                  🚫 Mark Do Not Contact
                </button>
                <button className="outreach-btn outreach-btn-danger" style={{ padding: '6px 12px', fontSize: 12, backgroundColor: '#7B341E', borderColor: '#7B341E' }} onClick={() => handleStatusAction('BOUNCE')}>
                  💥 Mark Email Bounced
                </button>
              </div>
            </div>
          </div>
        )}

        {/* TAB 2: ENROLLED CAMPAIGNS */}
        {activeTab === DETAIL_TABS.CAMPAIGNS && (
          <div className="outreach-campaigns-tab">
            <div className="outreach-card-header">
              <h3 className="outreach-section-title">Enrolled Campaign History</h3>
              <p className="outreach-section-subtitle">Campaigns that target this prospect.</p>
            </div>

            {detail?.campaigns?.length === 0 ? (
              <div className="outreach-empty-state" style={{ padding: '40px 10px', border: '1px dashed #E2E8F0', borderRadius: 8, textAlign: 'center' }}>
                <p>Not enrolled in any campaign yet.</p>
              </div>
            ) : (
              <div className="outreach-table-wrapper" style={{ border: '1px solid #E2E8F0', borderRadius: 12 }}>
                <table className="outreach-custom-table" style={{ width: '100%' }}>
                  <thead>
                    <tr style={{ background: '#F8FAFC', textAlign: 'left' }}>
                      <th style={{ padding: 12 }}>Campaign Name</th>
                      <th style={{ padding: 12 }}>Status</th>
                      <th style={{ padding: 12 }}>Date Enrolled</th>
                    </tr>
                  </thead>
                  <tbody>
                    {detail.campaigns.map(c => (
                      <tr key={c.id} style={{ borderBottom: '1px solid #F1F5F9' }}>
                        <td 
                          style={{ padding: 12, fontWeight: 700, color: '#2563EB', cursor: 'pointer' }}
                          onClick={() => onSelectCampaign && onSelectCampaign(c.id)}
                        >
                          📢 {c.name}
                        </td>
                        <td style={{ padding: 12 }}>
                          <span className={`campaign-status-tag ${c.status.toLowerCase()}`}>{c.status}</span>
                        </td>
                        <td style={{ padding: 12, color: '#64748B', fontSize: 12.5 }}>{c.created_at ? new Date(c.created_at).toLocaleDateString() : '—'}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
          </div>
        )}

        {/* TAB 3: SEQUENCE PROGRESSION */}
        {activeTab === DETAIL_TABS.SEQUENCE && (
          <div className="outreach-sequence-tab">
            <div className="outreach-card-header">
              <h3 className="outreach-section-title">Outreach Sequence Progression</h3>
              <p className="outreach-section-subtitle">Trace current step and verify step history timeline.</p>
            </div>

            {detail?.sequenceSteps?.length === 0 ? (
              <div className="outreach-empty-state" style={{ padding: '40px 10px', border: '1px dashed #E2E8F0', borderRadius: 8, textAlign: 'center' }}>
                <p>No sequence steps configured for this prospect's campaign.</p>
              </div>
            ) : (
              <div style={{ position: 'relative', paddingLeft: 20, borderLeft: '2px solid #E2E8F0', marginLeft: 10 }}>
                {detail.sequenceSteps.map((step, idx) => {
                  const isCurrent = idx + 1 === p?.current_step;
                  const isPast = idx + 1 < p?.current_step;
                  
                  return (
                    <div key={step.id} style={{ position: 'relative', marginBottom: 20 }}>
                      <div style={{
                        position: 'absolute',
                        left: -27,
                        top: 2,
                        width: 14,
                        height: 14,
                        borderRadius: '50%',
                        border: '2px solid #FFFFFF',
                        background: isCurrent ? '#3B82F6' : isPast ? '#10B981' : '#CBD5E1',
                        boxShadow: '0 1px 3px rgba(0,0,0,0.1)'
                      }} />
                      <div style={{ background: '#FFFFFF', padding: 12, borderRadius: 8, border: isCurrent ? '1.5px solid #3B82F6' : '1px solid #E2E8F0' }}>
                        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                          <span style={{ fontWeight: 700, fontSize: 13.5, color: isCurrent ? '#1E3A8A' : '#334155' }}>
                            Step {idx + 1}: {step.name || 'Sequence step'}
                          </span>
                          {isCurrent && <span style={{ fontSize: 11, background: '#DBEAFE', color: '#1E40AF', padding: '2px 6px', borderRadius: 6, fontWeight: 700 }}>CURRENT STEP</span>}
                          {isPast && <span style={{ fontSize: 11, color: '#059669', fontWeight: 600 }}>✓ COMPLETED</span>}
                        </div>
                        <div style={{ fontSize: 12.5, color: '#64748B', marginTop: 4 }}>Subject: <em>{step.subject || '—'}</em></div>
                        {step.body && <div style={{ fontSize: 12, color: '#475569', marginTop: 4, background: '#F8FAFC', padding: 8, borderRadius: 6, maxHeight: 100, overflowY: 'auto' }}>{step.body}</div>}
                      </div>
                    </div>
                  );
                })}
              </div>
            )}
          </div>
        )}

        {/* TAB 4: FOLLOW-UPS */}
        {activeTab === DETAIL_TABS.FOLLOW_UPS && (
          <div className="outreach-followups-tab">
            <div className="outreach-card-header" style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 16 }}>
              <div>
                <h3 className="outreach-section-title">Prospect Follow-up tasks</h3>
                <p className="outreach-section-subtitle">Manage calls, follow-up emails, and scheduled meetings.</p>
              </div>
              <button 
                className="outreach-btn outreach-btn-primary"
                onClick={(e) => onScheduleFollowUp(e, p)}
                style={{ fontSize: 12, padding: '6px 12px' }}
              >
                + Add Follow-up
              </button>
            </div>

            {detail?.followUps?.length === 0 ? (
              <div className="outreach-empty-state" style={{ padding: '40px 10px', border: '1px dashed #E2E8F0', borderRadius: 8, textAlign: 'center' }}>
                <p>No follow-up actions logged.</p>
              </div>
            ) : (
              <div className="outreach-table-wrapper" style={{ border: '1px solid #E2E8F0', borderRadius: 12 }}>
                <table className="outreach-custom-table" style={{ width: '100%' }}>
                  <thead>
                    <tr style={{ background: '#F8FAFC', textAlign: 'left' }}>
                      <th style={{ padding: 12 }}>Type</th>
                      <th style={{ padding: 12 }}>Subject / Notes</th>
                      <th style={{ padding: 12 }}>Due Date</th>
                      <th style={{ padding: 12 }}>Status</th>
                      <th style={{ padding: 12 }}>Priority</th>
                      <th style={{ padding: 12 }}>Actions</th>
                    </tr>
                  </thead>
                  <tbody>
                    {detail.followUps.map(f => (
                      <tr key={f.id} style={{ borderBottom: '1px solid #F1F5F9' }}>
                        <td style={{ padding: 12, fontWeight: 700 }}>
                          {f.activity_type === 'EMAIL' ? '✉️ Email' : f.activity_type === 'CALL' ? '📞 Call' : f.activity_type === 'MEETING' ? '👥 Meeting' : '🔔 Follow-up'}
                        </td>
                        <td style={{ padding: 12 }}>
                          <div style={{ fontWeight: 650, fontSize: 13, color: '#334155' }}>{f.subject}</div>
                          {f.description && <div style={{ fontSize: 11.5, color: '#64748B', marginTop: 2 }}>{f.description}</div>}
                        </td>
                        <td style={{ padding: 12, fontSize: 12.5, color: '#475569' }}>
                          {formatDateTime(f.scheduled_at)}
                        </td>
                        <td style={{ padding: 12 }}>
                          <span className={`activity-status-pill-grad ${f.status.toLowerCase()}`} style={{ fontSize: 10, padding: '2px 6px' }}>
                            {f.status}
                          </span>
                        </td>
                        <td style={{ padding: 12 }}>
                          <span style={{ fontSize: 12, fontWeight: 700, color: f.priority === 'HIGH' ? '#DC2626' : '#475569' }}>{f.priority}</span>
                        </td>
                        <td style={{ padding: 12 }}>
                          {f.status === 'PENDING' && (
                            <div style={{ display: 'flex', gap: 6 }}>
                              <button className="activity-btn-outline" style={{ padding: '2px 6px', fontSize: 11 }} onClick={() => handleCompleteFollowUp(f.id)}>✓ Complete</button>
                              <button className="activity-btn-outline" style={{ padding: '2px 6px', fontSize: 11, color: '#DC2626', borderColor: '#FCA5A5' }} onClick={() => handleCancelFollowUp(f.id)}>Cancel</button>
                            </div>
                          )}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
          </div>
        )}

        {/* TAB 5: ACTIVITY TIMELINE */}
        {activeTab === DETAIL_TABS.ACTIVITY && (
          <div className="outreach-activity-tab">
            <div className="outreach-card-header">
              <h3 className="outreach-section-title">Timeline of chronological events</h3>
              <p className="outreach-section-subtitle">Real-time trace of outreach logs and tasks linked to this prospect.</p>
            </div>

            {detail?.activities?.length === 0 ? (
              <div className="outreach-empty-state" style={{ padding: '40px 10px', border: '1px dashed #E2E8F0', borderRadius: 8, textAlign: 'center' }}>
                <p>No activity logs populated yet.</p>
              </div>
            ) : (
              <div style={{ position: 'relative', paddingLeft: 20, borderLeft: '2px solid #E2E8F0', marginLeft: 10 }}>
                {detail.activities.map(act => (
                  <div key={act.id} style={{ position: 'relative', marginBottom: 20 }}>
                    <div style={{
                      position: 'absolute',
                      left: -27,
                      top: 4,
                      width: 12,
                      height: 12,
                      borderRadius: '50%',
                      background: '#10B981',
                      border: '2px solid #FFFFFF',
                      boxShadow: '0 1px 3px rgba(0,0,0,0.1)'
                    }} />
                    <div style={{ background: '#FFFFFF', padding: 12, borderRadius: 8, border: '1px solid #E2E8F0' }}>
                      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                        <span style={{ fontWeight: 700, fontSize: 13, color: '#334155' }}>
                          {act.activity_type}: {act.subject}
                        </span>
                        <span style={{ fontSize: 11, color: '#64748B' }}>{formatDateTime(act.created_at)}</span>
                      </div>
                      {act.description && <div style={{ fontSize: 12, color: '#475569', marginTop: 4, background: '#F8FAFC', padding: 8, borderRadius: 6 }}>{act.description}</div>}
                      <div style={{ display: 'flex', gap: 10, marginTop: 6, alignItems: 'center' }}>
                        <span className={`activity-status-pill-grad ${act.status.toLowerCase()}`} style={{ fontSize: 10, padding: '2px 6px' }}>
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

ProspectDetail.propTypes = {
  leadId: PropTypes.number.isRequired,
  onBack: PropTypes.func.isRequired,
  onScheduleFollowUp: PropTypes.func.isRequired,
  onSelectCampaign: PropTypes.func,
};

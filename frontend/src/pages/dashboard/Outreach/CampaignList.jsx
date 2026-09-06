import { useState, useRef, useEffect } from 'react';
import PropTypes from 'prop-types';
import {
  CAMPAIGN_STATUS,
  CAMPAIGN_ACTIONS,
} from './constants';
import {
  activateCampaign,
  pauseCampaign,
  deleteCampaign,
} from '../../../services/outreachService';
import { Folder, Search, Play, Pause, Trash2, FolderOpen, ArrowRight, Eye, Calendar, User, SlidersHorizontal, RefreshCw } from 'lucide-react';

export default function CampaignList({
  campaigns,
  loading,
  onCampaignsChanged,
  onNewCampaignClick,
  onSelectCampaign
}) {
  const [actionLoadingId, setActionLoadingId] = useState(null);
  const [searchQuery, setSearchQuery] = useState('');
  const [activeFilter, setActiveFilter] = useState('ALL');
  const [activeDropdownId, setActiveDropdownId] = useState(null);
  const dropdownRef = useRef(null);

  // Close dropdown when clicking outside
  useEffect(() => {
    function handleClickOutside(event) {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target)) {
        setActiveDropdownId(null);
      }
    }
    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  async function handleAction(e, action, campaign) {
    e.stopPropagation();
    setActiveDropdownId(null);
    const confirmMsg = getConfirmMessage(action, campaign.name);
    if (confirmMsg && !window.confirm(confirmMsg)) return;

    setActionLoadingId(campaign.id);
    try {
      if (action === CAMPAIGN_ACTIONS.ACTIVATE) await activateCampaign(campaign.id);
      if (action === CAMPAIGN_ACTIONS.PAUSE) await pauseCampaign(campaign.id);
      if (action === CAMPAIGN_ACTIONS.DELETE) await deleteCampaign(campaign.id);
      onCampaignsChanged();
    } catch (err) {
      window.alert(err.message || `Failed to ${action} campaign`);
    } finally {
      setActionLoadingId(null);
    }
  }

  function getConfirmMessage(action, name) {
    if (action === CAMPAIGN_ACTIONS.DELETE)
      return `Delete campaign "${name}"? This cannot be undone.`;
    if (action === CAMPAIGN_ACTIONS.ACTIVATE)
      return `Activate "${name}"? This will validate audience and sequence steps.`;
    return `Pause "${name}"?`;
  }

  function formatDate(dateStr) {
    if (!dateStr) return '—';
    return new Date(dateStr).toLocaleDateString('en-US', {
      month: 'short',
      day: 'numeric',
      year: 'numeric'
    });
  }

  // Local calculations for filter tabs counts
  const counts = {
    ALL: campaigns.length,
    ACTIVE: campaigns.filter(c => c.status === CAMPAIGN_STATUS.ACTIVE).length,
    DRAFT: campaigns.filter(c => c.status === CAMPAIGN_STATUS.DRAFT).length,
    PAUSED: campaigns.filter(c => c.status === CAMPAIGN_STATUS.PAUSED).length,
    COMPLETED: campaigns.filter(c => c.status === CAMPAIGN_STATUS.COMPLETED).length,
  };

  // Local filtering
  const filteredCampaigns = campaigns.filter(campaign => {
    // 1. Filter by Status Tab
    if (activeFilter !== 'ALL' && campaign.status !== activeFilter) {
      return false;
    }
    // 2. Filter by Search Query
    if (searchQuery.trim()) {
      const q = searchQuery.toLowerCase();
      return (
        campaign.name.toLowerCase().includes(q) ||
        (campaign.status || '').toLowerCase().includes(q)
      );
    }
    return true;
  });

  // Render Skeleton Loading rows
  if (loading) {
    return (
      <div className="campaign-list-card" style={{ background: '#FFFFFF', border: '1px solid #E2E8F0', borderRadius: 12, padding: 18, boxShadow: '0 1px 3px rgba(0,0,0,0.05)' }}>
        <div className="campaign-list-actions-bar" style={{ display: 'flex', gap: 12, marginBottom: 16 }}>
          <div className="skeleton-line" style={{ width: 320, height: 38, borderRadius: 8 }} />
          <div className="skeleton-line" style={{ width: 360, height: 38, borderRadius: 8 }} />
        </div>
        <div className="outreach-table-wrap">
          <table className="outreach-table">
            <thead>
              <tr>
                <th>Campaign</th>
                <th>Status</th>
                <th>Created</th>
                <th>Last Updated</th>
                <th>Actions</th>
              </tr>
            </thead>
            <tbody>
              {Array.from({ length: 4 }).map((_, i) => (
                <tr key={i} className="outreach-skeleton-row">
                  <td>
                    <div style={{ display: 'flex', gap: 12, alignItems: 'center' }}>
                      <div className="skeleton-circle" />
                      <div>
                        <div className="skeleton-line" style={{ width: 140, marginBottom: 6 }} />
                        <div className="skeleton-line" style={{ width: 90 }} />
                      </div>
                    </div>
                  </td>
                  <td><div className="skeleton-line" style={{ width: 70, height: 22 }} /></td>
                  <td><div className="skeleton-line" style={{ width: 80 }} /></td>
                  <td><div className="skeleton-line" style={{ width: 80 }} /></td>
                  <td><div className="skeleton-line" style={{ width: 100, height: 30 }} /></td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    );
  }

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 24 }}>
      
      {/* ─── Campaign List Table Container Card ─── */}
      <div className="campaign-list-card" style={{ background: '#FFFFFF', border: '1px solid #E2E8F0', borderRadius: 12, padding: 20, boxShadow: '0 1px 3px rgba(0,0,0,0.05)' }}>
        
        {/* ─── Search & Status Filters Bar ─── */}
        <div className="campaign-list-actions-bar" style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', gap: 12, marginBottom: 20, flexWrap: 'wrap' }}>
          <div className="search-input-wrapper" style={{ flex: 1, maxWidth: 360, position: 'relative', display: 'flex', alignItems: 'center' }}>
            <Search size={14} style={{ position: 'absolute', left: 12, color: '#64748B' }} />
            <input
              type="text"
              placeholder="Search campaigns by name, status..."
              value={searchQuery}
              onChange={e => setSearchQuery(e.target.value)}
              className="activity-search-input-premium"
              style={{ width: '100%', paddingLeft: 34, height: 38, fontSize: 13, border: '1px solid #CBD5E1', borderRadius: 8, background: '#FFFFFF', boxSizing: 'border-box' }}
            />
          </div>

          <div style={{ display: 'flex', gap: 10, alignItems: 'center' }}>
            <div className="activities-status-tabs shadow-capsule" style={{ background: '#F1F5F9', padding: 3, borderRadius: 8, display: 'flex', gap: 2 }}>
              {['ALL', 'ACTIVE', 'DRAFT', 'PAUSED', 'COMPLETED'].map(tab => (
                <button
                  key={tab}
                  className={`status-tab-btn ${activeFilter === tab ? 'active' : ''}`}
                  onClick={() => setActiveFilter(tab)}
                  style={{ fontSize: 12, padding: '6px 12px', borderRadius: 6, border: 'none', background: activeFilter === tab ? '#FFFFFF' : 'transparent', color: activeFilter === tab ? '#0F172A' : '#64748B', fontWeight: activeFilter === tab ? 700 : 500, cursor: 'pointer', display: 'inline-flex', alignItems: 'center', gap: 4 }}
                >
                  {tab === 'ALL' ? 'All' : tab.charAt(0) + tab.slice(1).toLowerCase()}
                  <span style={{ fontSize: 10.5, color: activeFilter === tab ? '#2563EB' : '#94A3B8', fontWeight: 700 }}>({counts[tab]})</span>
                </button>
              ))}
            </div>

            {searchQuery || activeFilter !== 'ALL' ? (
              <button 
                onClick={() => { setSearchQuery(''); setActiveFilter('ALL'); }}
                style={{ background: 'none', border: 'none', color: '#64748B', fontSize: 12, fontWeight: 600, display: 'inline-flex', alignItems: 'center', gap: 4, cursor: 'pointer' }}
              >
                Clear Filters
              </button>
            ) : null}
          </div>
        </div>

        {/* ─── Table Content / Empty States ─── */}
        {campaigns.length === 0 ? (
          <div className="outreach-empty" style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', padding: '60px 20px', textAlign: 'center' }}>
            <FolderOpen size={48} style={{ color: '#94A3B8', marginBottom: 16 }} />
            <div className="outreach-empty-title" style={{ fontSize: 16, fontWeight: 750, color: '#0F172A', marginBottom: 6 }}>No campaigns yet</div>
            <div className="outreach-empty-sub" style={{ fontSize: 13, color: '#64748B', maxWidth: 440, lineHeight: 1.5, marginBottom: 20 }}>
              You haven't created any outreach campaigns yet. Start by creating your first campaign and reaching out to new prospects.
            </div>
            {onNewCampaignClick && (
              <button
                className="outreach-btn outreach-btn-primary"
                onClick={onNewCampaignClick}
                style={{ padding: '10px 20px', fontSize: 13, borderRadius: 8, fontWeight: 600 }}
              >
                + Create Your First Campaign
              </button>
            )}
          </div>
        ) : filteredCampaigns.length === 0 ? (
          <div className="outreach-empty" style={{ padding: '60px 20px', textAlign: 'center' }}>
            <Search size={36} style={{ color: '#94A3B8', marginBottom: 12 }} />
            <div className="outreach-empty-title" style={{ fontSize: 15, fontWeight: 750, color: '#0F172A', marginBottom: 4 }}>No matching campaigns</div>
            <div className="outreach-empty-sub" style={{ fontSize: 13, color: '#64748B' }}>Try modifying your search query or status filter.</div>
          </div>
        ) : (
          <div style={{ overflowX: 'auto' }}>
            <table className="outreach-table" style={{ width: '100%', borderCollapse: 'collapse' }}>
              <thead>
                <tr style={{ borderBottom: '1px solid #E2E8F0', background: '#F8FAFC' }}>
                  <th style={{ textAlign: 'left', padding: '12px 16px', fontSize: 11.5, fontWeight: 700, color: '#475569', textTransform: 'uppercase' }}>Campaign Name</th>
                  <th style={{ textAlign: 'left', padding: '12px 16px', fontSize: 11.5, fontWeight: 700, color: '#475569', textTransform: 'uppercase' }}>Status</th>
                  <th style={{ textAlign: 'left', padding: '12px 16px', fontSize: 11.5, fontWeight: 700, color: '#475569', textTransform: 'uppercase' }}>Created Date</th>
                  <th style={{ textAlign: 'left', padding: '12px 16px', fontSize: 11.5, fontWeight: 700, color: '#475569', textTransform: 'uppercase' }}>Last Updated</th>
                  <th style={{ textAlign: 'right', padding: '12px 16px', fontSize: 11.5, fontWeight: 700, color: '#475569', textTransform: 'uppercase' }}>Actions</th>
                </tr>
              </thead>
              <tbody>
                {filteredCampaigns.map(campaign => {
                  const isThisLoading = actionLoadingId === campaign.id;
                  const showDropdown = activeDropdownId === campaign.id;

                  return (
                    <tr
                      key={campaign.id}
                      onClick={() => onSelectCampaign(campaign)}
                      className="premium-table-row"
                      style={{ cursor: 'pointer', borderBottom: '1px solid #F1F5F9', transition: 'background-color 0.2s' }}
                    >
                      <td style={{ padding: '14px 16px' }}>
                        <div className="prospect-info-cell" style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
                          <div className="campaign-avatar-container" style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', background: '#EFF6FF', borderRadius: '50%', width: 32, height: 32, flexShrink: 0 }}>
                            <Folder size={15} style={{ color: '#2563EB' }} />
                          </div>
                          <div>
                            <div className="prospect-company-name-highlight" style={{ fontSize: 13.5, fontWeight: 700, color: '#0F172A' }}>{campaign.name}</div>
                            <div className="prospect-contact-role" style={{ fontSize: 11.5, color: '#64748B', marginTop: 2 }}>Click to manage sequence steps</div>
                          </div>
                        </div>
                      </td>
                      <td style={{ padding: '14px 16px' }}>
                        <span className={`activity-status-pill-grad ${campaign.status?.toLowerCase() || 'draft'}`} style={{ fontSize: 11, padding: '3px 8px' }}>
                          {campaign.status || CAMPAIGN_STATUS.DRAFT}
                        </span>
                      </td>
                      <td style={{ padding: '14px 16px' }}>
                        <div style={{ fontSize: 12.5, fontWeight: 600, color: '#334155', display: 'flex', alignItems: 'center', gap: 6 }}>
                          <Calendar size={13} style={{ color: '#94A3B8' }} /> {formatDate(campaign.created_at)}
                        </div>
                      </td>
                      <td style={{ padding: '14px 16px' }}>
                        <div style={{ fontSize: 12.5, fontWeight: 500, color: '#64748B' }}>
                          {formatDate(campaign.updated_at || campaign.created_at)}
                        </div>
                      </td>
                      <td style={{ padding: '14px 16px', textAlign: 'right' }} onClick={e => e.stopPropagation()}>
                        <div className="campaign-actions-cell" style={{ display: 'flex', alignItems: 'center', justifyContent: 'flex-end', gap: 8 }}>
                          {/* Primary action based on status */}
                          {(campaign.status === CAMPAIGN_STATUS.DRAFT || campaign.status === CAMPAIGN_STATUS.PAUSED) && (
                            <button
                              className="campaign-action-pill-btn launch"
                              onClick={(e) => handleAction(e, CAMPAIGN_ACTIONS.ACTIVATE, campaign)}
                              disabled={isThisLoading}
                              style={{ display: 'inline-flex', alignItems: 'center', gap: 4, padding: '4px 10px', fontSize: 11.5, borderRadius: 6, fontWeight: 700, border: 'none', background: '#ECFDF4', color: '#059669', cursor: 'pointer' }}
                            >
                              <Play size={10} fill="currentColor" /> {isThisLoading ? '...' : 'Launch'}
                            </button>
                          )}

                          {campaign.status === CAMPAIGN_STATUS.ACTIVE && (
                            <button
                              className="campaign-action-pill-btn pause"
                              onClick={(e) => handleAction(e, CAMPAIGN_ACTIONS.PAUSE, campaign)}
                              disabled={isThisLoading}
                              style={{ display: 'inline-flex', alignItems: 'center', gap: 4, padding: '4px 10px', fontSize: 11.5, borderRadius: 6, fontWeight: 700, border: 'none', background: '#FFFBEB', color: '#D97706', cursor: 'pointer' }}
                            >
                              <Pause size={10} fill="currentColor" /> {isThisLoading ? '...' : 'Pause'}
                            </button>
                          )}

                          {campaign.status === CAMPAIGN_STATUS.COMPLETED && (
                            <button className="campaign-action-pill-btn completed-btn" disabled style={{ padding: '4px 10px', fontSize: 11.5, borderRadius: 6, fontWeight: 700, border: 'none', background: '#F1F5F9', color: '#64748B', cursor: 'not-allowed' }}>
                              Done
                            </button>
                          )}

                          {/* Kebab Dropdown Menu */}
                          <div className="dropdown-wrapper" ref={showDropdown ? dropdownRef : null} style={{ position: 'relative' }}>
                            <button
                              className="kebab-action-btn-circle"
                              onClick={(e) => {
                                e.stopPropagation();
                                setActiveDropdownId(showDropdown ? null : campaign.id);
                              }}
                              style={{ background: '#F8FAFC', border: '1px solid #E2E8F0', width: 28, height: 28, borderRadius: '50%', cursor: 'pointer', display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: 10, color: '#64748B' }}
                            >
                              •••
                            </button>

                            {showDropdown && (
                              <div className="floating-dropdown-menu" style={{ position: 'absolute', right: 0, top: 32, background: '#FFFFFF', border: '1px solid #E2E8F0', borderRadius: 8, boxShadow: '0 4px 12px rgba(0,0,0,0.1)', zIndex: 10, padding: 4, minWidth: 140 }}>
                                <button
                                  className="dropdown-item"
                                  onClick={() => onSelectCampaign(campaign)}
                                  style={{ width: '100%', padding: '6px 10px', fontSize: 12, textAlign: 'left', border: 'none', background: 'none', borderRadius: 4, cursor: 'pointer', color: '#334155', display: 'flex', alignItems: 'center', gap: 6 }}
                                >
                                  <Eye size={12} /> View Details
                                </button>
                                <button
                                  className="dropdown-item text-danger"
                                  onClick={(e) => handleAction(e, CAMPAIGN_ACTIONS.DELETE, campaign)}
                                  disabled={isThisLoading}
                                  style={{ width: '100%', padding: '6px 10px', fontSize: 12, textAlign: 'left', border: 'none', background: 'none', borderRadius: 4, cursor: 'pointer', color: '#DC2626', display: 'flex', alignItems: 'center', gap: 6 }}
                                >
                                  <Trash2 size={12} /> Delete Campaign
                                </button>
                              </div>
                            )}
                          </div>
                        </div>
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        )}

        {/* ─── Compact Table Footer ─── */}
        {filteredCampaigns.length > 0 && (
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginTop: 16, paddingTop: 14, borderTop: '1px solid #F1F5F9', fontSize: 12.5, color: '#64748B' }}>
            <div>
              Showing <strong>1–{filteredCampaigns.length}</strong> of <strong>{campaigns.length}</strong> campaigns
            </div>
            <div style={{ display: 'flex', gap: 6 }}>
              <button disabled style={{ padding: '4px 10px', border: '1px solid #E2E8F0', borderRadius: 6, background: '#F8FAFC', color: '#94A3B8', fontSize: 11.5, cursor: 'not-allowed' }}>Prev</button>
              <button disabled style={{ padding: '4px 10px', border: '1px solid #E2E8F0', borderRadius: 6, background: '#F8FAFC', color: '#94A3B8', fontSize: 11.5, cursor: 'not-allowed' }}>Next</button>
            </div>
          </div>
        )}
      </div>

      {/* ─── Getting Started Onboarding Step Flow ─── */}
      <div style={{ background: '#FFFFFF', border: '1px solid #E2E8F0', borderRadius: 12, padding: 24, boxShadow: '0 1px 3px rgba(0,0,0,0.05)' }}>
        <h4 style={{ margin: '0 0 4px 0', fontSize: 15, fontWeight: 750, color: '#0F172A' }}>Getting started with Outreach</h4>
        <p style={{ margin: '0 0 24px 0', fontSize: 13, color: '#64748B' }}>
          Create targeted campaigns, engage prospects, and convert them into opportunities.
        </p>

        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(4, 1fr)', gap: 16 }}>
          {/* Step 1 */}
          <div style={{ display: 'flex', flexDirection: 'column', gap: 8, position: 'relative' }}>
            <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
              <div style={{ background: '#EFF6FF', color: '#2563EB', width: 28, height: 28, borderRadius: '50%', display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: 12, fontWeight: 800 }}>
                1
              </div>
              <h5 style={{ margin: 0, fontSize: 13, fontWeight: 700, color: '#334155' }}>Create Campaign</h5>
            </div>
            <p style={{ margin: 0, fontSize: 12, color: '#64748B', lineHeight: 1.4 }}>
              Define the baseline templates and specify wait delays for subsequent steps.
            </p>
          </div>

          {/* Step 2 */}
          <div style={{ display: 'flex', flexDirection: 'column', gap: 8, position: 'relative' }}>
            <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
              <div style={{ background: '#EFF6FF', color: '#2563EB', width: 28, height: 28, borderRadius: '50%', display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: 12, fontWeight: 800 }}>
                2
              </div>
              <h5 style={{ margin: 0, fontSize: 13, fontWeight: 700, color: '#334155' }}>Add Prospects</h5>
            </div>
            <p style={{ margin: 0, fontSize: 12, color: '#64748B', lineHeight: 1.4 }}>
              Enroll targeted Leads into the campaign to map them for outbound schedules.
            </p>
          </div>

          {/* Step 3 */}
          <div style={{ display: 'flex', flexDirection: 'column', gap: 8, position: 'relative' }}>
            <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
              <div style={{ background: '#EFF6FF', color: '#2563EB', width: 28, height: 28, borderRadius: '50%', display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: 12, fontWeight: 800 }}>
                3
              </div>
              <h5 style={{ margin: 0, fontSize: 13, fontWeight: 700, color: '#334155' }}>Build Sequence</h5>
            </div>
            <p style={{ margin: 0, fontSize: 12, color: '#64748B', lineHeight: 1.4 }}>
              Construct follow-up email/call reminders to guide prospects down the pipeline.
            </p>
          </div>

          {/* Step 4 */}
          <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
            <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
              <div style={{ background: '#EFF6FF', color: '#2563EB', width: 28, height: 28, borderRadius: '50%', display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: 12, fontWeight: 800 }}>
                4
              </div>
              <h5 style={{ margin: 0, fontSize: 13, fontWeight: 700, color: '#334155' }}>Launch Campaign</h5>
            </div>
            <p style={{ margin: 0, fontSize: 12, color: '#64748B', lineHeight: 1.4 }}>
              Activate the campaign sequence to start automated runs and tracking.
            </p>
          </div>
        </div>
      </div>

    </div>
  );
}

CampaignList.propTypes = {
  campaigns: PropTypes.arrayOf(PropTypes.object).isRequired,
  loading: PropTypes.bool.isRequired,
  onCampaignsChanged: PropTypes.func.isRequired,
  onNewCampaignClick: PropTypes.func,
  onSelectCampaign: PropTypes.func.isRequired,
};
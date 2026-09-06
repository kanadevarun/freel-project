import { useState, useEffect, useCallback, useRef } from 'react';
import PropTypes from 'prop-types';
import { getProspects, pauseProspect, resumeProspect, stopProspect, updateProspect } from '../../../services/outreachService';
import { Search, Users, Settings, Mail, Phone, Clock, ChevronDown, UserCheck, Plus, SlidersHorizontal, ArrowUpDown, Filter, AlertCircle, Folder, Calendar } from 'lucide-react';
import './OutreachPage.css';

export default function ProspectList({ 
  onSelectProspect, 
  onScheduleFollowUp, 
  refreshCount, 
  onAddProspectClick,
  campaigns = []
}) {
  const [prospects, setProspects] = useState([]);
  const [loading, setLoading] = useState(true);
  const [searchQuery, setSearchQuery] = useState('');
  const [statusFilter, setStatusFilter] = useState('ALL');
  const [campaignFilter, setCampaignFilter] = useState('ALL');
  const [sortBy, setSortBy] = useState('NAME'); // 'NAME', 'COMPANY', 'STEP', 'DATE'
  const [activeDropdownId, setActiveDropdownId] = useState(null);
  const [error, setError] = useState('');
  const dropdownRef = useRef(null);

  const fetchProspects = useCallback(async () => {
    setLoading(true);
    setError('');
    try {
      const data = await getProspects();
      setProspects(data || []);
    } catch (err) {
      console.error('Failed to load prospects:', err);
      setError('Failed to fetch prospects: ' + (err.message || err.error || JSON.stringify(err) || err));
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchProspects();
  }, [fetchProspects, refreshCount]);

  // Click outside listener for dropdowns
  useEffect(() => {
    function handleClickOutside(event) {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target)) {
        setActiveDropdownId(null);
      }
    }
    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  async function handleStatusControl(e, leadId, campaignId, action) {
    e.stopPropagation();
    setActiveDropdownId(null);
    if (!campaignId) {
      alert('This prospect is not associated with an active campaign.');
      return;
    }
    try {
      if (action === 'PAUSE') {
        await pauseProspect(leadId, campaignId);
      } else if (action === 'RESUME') {
        await resumeProspect(leadId, campaignId);
      } else if (action === 'STOP') {
        await stopProspect(leadId, campaignId);
      } else if (action === 'DNC') {
        await updateProspect(leadId, campaignId, 'DO_NOT_CONTACT', 1);
      } else if (action === 'BOUNCE') {
        await updateProspect(leadId, campaignId, 'BOUNCED', 1);
      }
      await fetchProspects();
    } catch (err) {
      console.error(`Failed to execute ${action} on prospect:`, err);
      alert(`Action failed: ${err.message || 'Server error'}`);
    }
  }

  function formatDateTime(dateStr) {
    if (!dateStr) return '—';
    const date = new Date(dateStr);
    return date.toLocaleDateString('en-US', {
      month: 'short',
      day: 'numeric',
      year: 'numeric'
    });
  }

  // Get Initials avatar background color
  function getInitialsColor(name) {
    const colors = [
      { bg: '#EFF6FF', text: '#2563EB' }, // Blue
      { bg: '#ECFDF5', text: '#059669' }, // Green
      { bg: '#F5F3FF', text: '#7C3AED' }, // Purple
      { bg: '#FFFBEB', text: '#D97706' }, // Amber
      { bg: '#ECFEFF', text: '#0891B2' }, // Cyan
      { bg: '#FDF2F8', text: '#DB2777' }, // Pink
    ];
    let sum = 0;
    for (let i = 0; i < (name || '').length; i++) {
      sum += name.charCodeAt(i);
    }
    return colors[sum % colors.length];
  }

  function getInitials(name) {
    if (!name) return 'P';
    const parts = name.trim().split(/\s+/);
    if (parts.length >= 2) {
      return (parts[0][0] + parts[1][0]).toUpperCase();
    }
    return name[0].toUpperCase();
  }

  // Filters logic
  const filteredProspects = prospects.filter(p => {
    // 1. Search Query
    const q = searchQuery.toLowerCase();
    const matchesSearch =
      p.company_name.toLowerCase().includes(q) ||
      (p.contact_name || '').toLowerCase().includes(q) ||
      (p.email || '').toLowerCase().includes(q);
    
    // 2. Status Filter
    const matchesStatus = statusFilter === 'ALL' || p.status === statusFilter;

    // 3. Campaign Filter
    const matchesCampaign = campaignFilter === 'ALL' || String(p.campaign_id) === campaignFilter;
    
    return matchesSearch && matchesStatus && matchesCampaign;
  });

  // Sort logic
  const sortedProspects = [...filteredProspects].sort((a, b) => {
    if (sortBy === 'NAME') {
      return (a.contact_name || '').localeCompare(b.contact_name || '');
    }
    if (sortBy === 'COMPANY') {
      return a.company_name.localeCompare(b.company_name);
    }
    if (sortBy === 'STEP') {
      return (a.current_step || 0) - (b.current_step || 0);
    }
    if (sortBy === 'DATE') {
      const dateA = a.next_scheduled_at ? new Date(a.next_scheduled_at) : new Date(0);
      const dateB = b.next_scheduled_at ? new Date(b.next_scheduled_at) : new Date(0);
      return dateB - dateA;
    }
    return 0;
  });

  const statusCounts = {
    ALL: prospects.length,
    ACTIVE: prospects.filter(p => p.status === 'ACTIVE').length,
    PAUSED: prospects.filter(p => p.status === 'PAUSED').length,
    COMPLETED: prospects.filter(p => p.status === 'COMPLETED').length,
    BOUNCED: prospects.filter(p => p.status === 'BOUNCED').length,
    UNSUBSCRIBED: prospects.filter(p => p.status === 'UNSUBSCRIBED').length,
    DO_NOT_CONTACT: prospects.filter(p => p.status === 'DO_NOT_CONTACT').length,
  };

  return (
    <div className="prospect-list-workspace" style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
      
      {/* ─── CRM Style Toolbar ─── */}
      <div className="campaign-list-card" style={{ background: '#FFFFFF', border: '1px solid #E2E8F0', borderRadius: 12, padding: 18, boxShadow: '0 1px 3px rgba(0,0,0,0.05)' }}>
        <div style={{ display: 'flex', gap: 12, alignItems: 'center', justifyContent: 'space-between', flexWrap: 'wrap' }}>
          
          {/* Search bar */}
          <div className="search-box-wrapper" style={{ flex: 1, minWidth: 260, maxWidth: 360, position: 'relative', display: 'flex', alignItems: 'center' }}>
            <Search size={14} style={{ position: 'absolute', left: 12, color: '#64748B' }} />
            <input
              type="text"
              placeholder="Search prospects by name, company or email..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="campaign-search-input"
              style={{ width: '100%', paddingLeft: 34, height: 38, fontSize: 13, border: '1px solid #CBD5E1', borderRadius: 8, background: '#FFFFFF', boxSizing: 'border-box' }}
            />
          </div>

          {/* Campaign & Sort Filter Dropdowns */}
          <div style={{ display: 'flex', gap: 10, alignItems: 'center', flexWrap: 'wrap' }}>
            {/* Campaign dropdown */}
            <div style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
              <Filter size={13} style={{ color: '#64748B' }} />
              <select
                value={campaignFilter}
                onChange={e => setCampaignFilter(e.target.value)}
                style={{ padding: '8px 12px', height: 38, border: '1px solid #CBD5E1', borderRadius: 8, fontSize: 12.5, background: '#FFFFFF', color: '#334155', fontWeight: 600, cursor: 'pointer' }}
              >
                <option value="ALL">All Campaigns</option>
                {campaigns.map(c => (
                  <option key={c.id} value={c.id}>{c.name}</option>
                ))}
              </select>
            </div>

            {/* Sort options */}
            <div style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
              <ArrowUpDown size={13} style={{ color: '#64748B' }} />
              <select
                value={sortBy}
                onChange={e => setSortBy(e.target.value)}
                style={{ padding: '8px 12px', height: 38, border: '1px solid #CBD5E1', borderRadius: 8, fontSize: 12.5, background: '#FFFFFF', color: '#334155', fontWeight: 600, cursor: 'pointer' }}
              >
                <option value="NAME">Sort by Name</option>
                <option value="COMPANY">Sort by Company</option>
                <option value="STEP">Sort by Sequence Step</option>
                <option value="DATE">Sort by Next Follow-up</option>
              </select>
            </div>

            {/* Status Segment Switcher */}
            <div className="activities-status-tabs shadow-capsule" style={{ background: '#F1F5F9', padding: 3, borderRadius: 8, display: 'flex', gap: 2 }}>
              {['ALL', 'ACTIVE', 'PAUSED', 'COMPLETED'].map(statusKey => (
                <button
                  key={statusKey}
                  onClick={() => setStatusFilter(statusKey)}
                  className={`status-tab-btn ${statusFilter === statusKey ? 'active' : ''}`}
                  style={{ fontSize: 11.5, padding: '6px 12px', borderRadius: 6, border: 'none', background: statusFilter === statusKey ? '#FFFFFF' : 'transparent', color: statusFilter === statusKey ? '#0F172A' : '#64748B', fontWeight: statusFilter === statusKey ? 700 : 500, cursor: 'pointer', display: 'inline-flex', alignItems: 'center', gap: 4 }}
                >
                  {statusKey === 'ALL' ? 'All' : statusKey.charAt(0) + statusKey.slice(1).toLowerCase()}
                  <span style={{ fontSize: 10, color: statusFilter === statusKey ? '#2563EB' : '#94A3B8', fontWeight: 700 }}>({statusCounts[statusKey]})</span>
                </button>
              ))}
            </div>
          </div>

        </div>
      </div>

      {/* ─── Main Content Workspace ─── */}
      {loading ? (
        <div className="campaign-list-card" style={{ background: '#FFFFFF', border: '1px solid #E2E8F0', borderRadius: 12, padding: 20 }}>
          <div className="campaign-list-skeleton">
            {Array.from({ length: 4 }).map((_, idx) => (
              <div key={idx} className="skeleton-row" style={{ height: 54, margin: '8px 0', background: '#F8FAFC', borderRadius: 8 }} />
            ))}
          </div>
        </div>
      ) : error ? (
        <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', padding: '50px 20px', border: '1px solid #FEE2E2', borderRadius: 12, background: '#FEF2F2', textAlign: 'center' }}>
          <AlertCircle size={40} style={{ color: '#EF4444', marginBottom: 12 }} />
          <h4 style={{ fontSize: 15.5, fontWeight: 750, color: '#991B1B', margin: '0 0 6px 0' }}>Unable to load prospects</h4>
          <p style={{ color: '#7F1D1D', fontSize: 13, maxWidth: 440, lineHeight: 1.5, margin: '0 0 16px 0' }}>
            We encountered a connection issue while fetching your outreach contacts. Please verify your connection or retry the request.
          </p>
          <div style={{ display: 'flex', gap: 10 }}>
            <button
              onClick={fetchProspects}
              className="outreach-btn outreach-btn-primary"
              style={{ padding: '8px 16px', fontSize: 12.5, borderRadius: 6, fontWeight: 600 }}
            >
              Retry Connection
            </button>
            <button
              onClick={() => setError('')}
              className="activity-btn-outline"
              style={{ padding: '8px 16px', fontSize: 12.5, borderRadius: 6, background: '#FFFFFF' }}
            >
              Dismiss
            </button>
          </div>
        </div>
      ) : prospects.length === 0 ? (
        /* Empty state when database is empty */
        <div className="outreach-empty-state" style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', padding: '60px 20px', border: '1.5px dashed #CBD5E1', borderRadius: 12, textAlign: 'center', background: '#FFFFFF' }}>
          <UserCheck size={48} style={{ color: '#94A3B8', marginBottom: 16 }} />
          <h4 style={{ fontSize: 16, fontWeight: 750, color: '#0F172A', margin: '0 0 6px 0' }}>No prospects yet</h4>
          <p style={{ color: '#64748B', fontSize: 13.5, maxWidth: 440, lineHeight: 1.5, margin: '0 0 20px 0' }}>
            Add prospects to start engaging potential customers through your outreach campaigns.
          </p>
          {onAddProspectClick && (
            <button
              className="outreach-btn outreach-btn-primary"
              onClick={onAddProspectClick}
              style={{ padding: '10px 20px', fontSize: 13, borderRadius: 8, fontWeight: 600 }}
            >
              + Add Prospect
            </button>
          )}
        </div>
      ) : sortedProspects.length === 0 ? (
        /* Empty state when search or filters result in zero matches */
        <div className="outreach-empty-state" style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', padding: '65px 20px', border: '1px solid #E2E8F0', borderRadius: 12, textAlign: 'center', background: '#FFFFFF' }}>
          <Search size={36} style={{ color: '#94A3B8', marginBottom: 12 }} />
          <h4 style={{ fontSize: 15, fontWeight: 750, color: '#0F172A', margin: '0 0 4px 0' }}>No matching prospects</h4>
          <p style={{ color: '#64748B', fontSize: 13, margin: 0 }}>Try clearing filters or search query to list all prospects.</p>
          <button
            onClick={() => { setSearchQuery(''); setStatusFilter('ALL'); setCampaignFilter('ALL'); }}
            style={{ marginTop: 12, background: 'none', border: 'none', color: '#2563EB', fontSize: 12.5, fontWeight: 700, cursor: 'pointer' }}
          >
            Reset Filters
          </button>
        </div>
      ) : (
        /* ─── Prospects Table View ─── */
        <div className="outreach-table-wrapper" style={{ background: '#FFFFFF', border: '1px solid #E2E8F0', borderRadius: 12, overflow: 'visible', boxShadow: '0 1px 3px rgba(0,0,0,0.05)' }}>
          <div style={{ overflowX: 'auto' }}>
            <table className="outreach-custom-table" style={{ width: '100%', borderCollapse: 'collapse' }}>
              <thead>
                <tr style={{ background: '#F8FAFC', borderBottom: '1px solid #E2E8F0', textAlign: 'left' }}>
                  <th style={{ padding: '12px 18px', fontSize: 11.5, fontWeight: 700, color: '#475569', textTransform: 'uppercase' }}>Prospect / Company</th>
                  <th style={{ padding: '12px 18px', fontSize: 11.5, fontWeight: 700, color: '#475569', textTransform: 'uppercase' }}>Campaign Target</th>
                  <th style={{ padding: '12px 18px', fontSize: 11.5, fontWeight: 700, color: '#475569', textTransform: 'uppercase' }}>Outreach Status</th>
                  <th style={{ padding: '12px 18px', fontSize: 11.5, fontWeight: 700, color: '#475569', textTransform: 'uppercase' }}>Current Step</th>
                  <th style={{ padding: '12px 18px', fontSize: 11.5, fontWeight: 700, color: '#475569', textTransform: 'uppercase' }}>Engagement</th>
                  <th style={{ padding: '12px 18px', fontSize: 11.5, fontWeight: 700, color: '#475569', textTransform: 'uppercase' }}>Next Follow-up</th>
                  <th style={{ padding: '12px 18px', fontSize: 11.5, fontWeight: 700, color: '#475569', textTransform: 'uppercase', textAlign: 'center' }}>Actions</th>
                </tr>
              </thead>
              <tbody>
                {sortedProspects.map(p => {
                  const colors = getInitialsColor(p.contact_name || p.company_name);
                  const isDropdownOpen = activeDropdownId === p.lead_id;

                  return (
                    <tr 
                      key={p.lead_id} 
                      style={{ borderBottom: '1px solid #F1F5F9', transition: 'background-color 0.2s' }} 
                      className="premium-table-row"
                    >
                      {/* Name / Initials column */}
                      <td style={{ padding: '16px 18px' }} onClick={() => onSelectProspect(p.lead_id)}>
                        <div style={{ display: 'flex', alignItems: 'center', gap: 12, cursor: 'pointer' }}>
                          <div style={{
                            width: 36,
                            height: 36,
                            borderRadius: '50%',
                            display: 'flex',
                            alignItems: 'center',
                            justifyContent: 'center',
                            fontSize: 13,
                            fontWeight: 800,
                            background: colors.bg,
                            color: colors.text,
                            flexShrink: 0,
                            boxShadow: '0 2px 5px rgba(0,0,0,0.06)',
                            border: '1.5px solid #FFFFFF'
                          }}>
                            {getInitials(p.contact_name || p.company_name)}
                          </div>
                          <div>
                            <div style={{ fontWeight: 700, color: '#0F172A', fontSize: 13.8 }}>{p.contact_name || 'Contact'}</div>
                            <div style={{ fontSize: 11.5, color: '#64748B', marginTop: 2, display: 'flex', alignItems: 'center', gap: 6 }}>
                              <span style={{ fontWeight: 650, color: '#475569' }}>{p.company_name}</span>
                              <span style={{ color: '#CBD5E1' }}>|</span>
                              <span style={{ color: '#64748B' }}>{p.email}</span>
                            </div>
                          </div>
                        </div>
                      </td>

                      {/* Campaign Column */}
                      <td style={{ padding: '16px 18px' }}>
                        {p.campaign_name ? (
                          <div style={{ display: 'inline-flex', alignItems: 'center', gap: 6, background: '#EFF6FF', color: '#1E40AF', padding: '5px 10px', borderRadius: 20, fontSize: 11.5, fontWeight: 700, border: '1px solid #DBEAFE' }}>
                            <Folder size={12} style={{ color: '#3B82F6' }} />
                            {p.campaign_name}
                          </div>
                        ) : (
                          <div style={{ display: 'inline-flex', alignItems: 'center', gap: 6, background: '#F1F5F9', color: '#64748B', padding: '5px 10px', borderRadius: 20, fontSize: 11.5, fontWeight: 600, border: '1px solid #E2E8F0' }}>
                            <Folder size={12} style={{ color: '#94A3B8' }} />
                            Unassigned
                          </div>
                        )}
                      </td>

                      {/* Status Column */}
                      <td style={{ padding: '16px 18px' }}>
                        <span className={`activity-status-pill-grad ${(p.status || 'DRAFT').toLowerCase()}`} style={{ fontSize: 10.5, padding: '4px 10px', borderRadius: 12, fontWeight: 700, textTransform: 'uppercase', letterSpacing: '0.02em', display: 'inline-flex', alignItems: 'center' }}>
                          {(p.status || 'DRAFT').replace(/_/g, ' ')}
                        </span>
                      </td>

                      {/* Current Step Column */}
                      <td style={{ padding: '16px 18px' }}>
                        <div style={{ display: 'inline-flex', alignItems: 'center', background: '#F8FAFC', border: '1px solid #E2E8F0', padding: '4px 10px', borderRadius: 8, fontSize: 12.5, fontWeight: 700, color: '#334155' }}>
                          Step {p.current_step}
                        </div>
                      </td>

                      {/* Engagement Column */}
                      <td style={{ padding: '16px 18px' }}>
                        <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                          <span className={`activity-status-pill-grad ${(p.engagement_status || 'AWAITING').toLowerCase()}`} style={{ fontSize: 10.5, padding: '4px 10px', borderRadius: 12, fontWeight: 700, textTransform: 'uppercase', letterSpacing: '0.02em' }}>
                            {(p.engagement_status || 'AWAITING').replace(/_/g, ' ')}
                          </span>
                          <span style={{ fontSize: 11.5, color: '#64748B', fontWeight: 650, display: 'inline-flex', alignItems: 'center', gap: 4 }}>
                            <Mail size={12} style={{ color: '#94A3B8' }} />
                            {p.emails_sent} sent
                          </span>
                        </div>
                      </td>

                      {/* Next Follow-up Column */}
                      <td style={{ padding: '16px 18px' }}>
                        {p.next_scheduled_at ? (
                          <div style={{ display: 'inline-flex', alignItems: 'center', gap: 6, fontSize: 12.5, color: '#334155', fontWeight: 650 }}>
                            <Calendar size={13} style={{ color: '#2563EB' }} />
                            {formatDateTime(p.next_scheduled_at)}
                          </div>
                        ) : (
                          <span style={{ color: '#94A3B8', fontSize: 12.5, display: 'inline-flex', alignItems: 'center', gap: 4 }}>
                            <Clock size={12} style={{ color: '#CBD5E1' }} />
                            Not Scheduled
                          </span>
                        )}
                      </td>

                      {/* Actions Column */}
                      <td style={{ padding: '16px 18px', textAlign: 'center', position: 'relative' }} onClick={e => e.stopPropagation()}>
                        <div style={{ display: 'inline-block' }} ref={isDropdownOpen ? dropdownRef : null}>
                          <button
                            className="kebab-action-btn-circle"
                            onClick={(e) => {
                              e.stopPropagation();
                              setActiveDropdownId(isDropdownOpen ? null : p.lead_id);
                            }}
                            style={{ background: '#FFFFFF', border: '1px solid #E2E8F0', width: 30, height: 30, borderRadius: '50%', cursor: 'pointer', display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: 10, color: '#64748B', transition: 'all 0.15s ease', boxShadow: '0 1px 2px rgba(0,0,0,0.03)' }}
                          >
                            •••
                          </button>

                          {isDropdownOpen && (
                            <div className="floating-dropdown-menu" style={{ position: 'absolute', right: 0, top: 32, background: '#FFFFFF', border: '1px solid #E2E8F0', borderRadius: 8, boxShadow: '0 4px 12px rgba(0,0,0,0.1)', zIndex: 100, padding: 4, minWidth: 170, textAlign: 'left' }}>
                              <button 
                                className="dropdown-item" 
                                onClick={() => { setActiveDropdownId(null); onSelectProspect(p.lead_id); }}
                                style={{ width: '100%', padding: '6px 10px', fontSize: 12, textAlign: 'left', border: 'none', background: 'none', borderRadius: 4, cursor: 'pointer', color: '#334155', display: 'flex', alignItems: 'center', gap: 6 }}
                              >
                                👁️ View Details
                              </button>
                              <button 
                                className="dropdown-item" 
                                onClick={(e) => { setActiveDropdownId(null); onScheduleFollowUp(e, p); }}
                                style={{ width: '100%', padding: '6px 10px', fontSize: 12, textAlign: 'left', border: 'none', background: 'none', borderRadius: 4, cursor: 'pointer', color: '#334155', display: 'flex', alignItems: 'center', gap: 6 }}
                              >
                                ⏳ Schedule Follow-up
                              </button>
                              
                              <div style={{ height: '1px', background: '#E2E8F0', margin: '4px 0' }} />
                              
                              {p.status === 'ACTIVE' ? (
                                <button 
                                  className="dropdown-item" 
                                  onClick={(e) => handleStatusControl(e, p.lead_id, p.campaign_id, 'PAUSE')}
                                  style={{ width: '100%', padding: '6px 10px', fontSize: 12, textAlign: 'left', border: 'none', background: 'none', borderRadius: 4, cursor: 'pointer', color: '#D97706', display: 'flex', alignItems: 'center', gap: 6 }}
                                >
                                  ⏸️ Pause Outreach
                                </button>
                              ) : p.status === 'PAUSED' ? (
                                <button 
                                  className="dropdown-item" 
                                  onClick={(e) => handleStatusControl(e, p.lead_id, p.campaign_id, 'RESUME')}
                                  style={{ width: '100%', padding: '6px 10px', fontSize: 12, textAlign: 'left', border: 'none', background: 'none', borderRadius: 4, cursor: 'pointer', color: '#059669', display: 'flex', alignItems: 'center', gap: 6 }}
                                >
                                  ▶️ Resume Outreach
                                </button>
                              ) : null}

                              {p.status !== 'COMPLETED' && (
                                <button 
                                  className="dropdown-item" 
                                  onClick={(e) => handleStatusControl(e, p.lead_id, p.campaign_id, 'STOP')}
                                  style={{ width: '100%', padding: '6px 10px', fontSize: 12, textAlign: 'left', border: 'none', background: 'none', borderRadius: 4, cursor: 'pointer', color: '#64748B', display: 'flex', alignItems: 'center', gap: 6 }}
                                >
                                  🛑 Stop Outreach
                                </button>
                              )}

                              <button 
                                className="dropdown-item" 
                                onClick={(e) => handleStatusControl(e, p.lead_id, p.campaign_id, 'DNC')} 
                                style={{ width: '100%', padding: '6px 10px', fontSize: 12, textAlign: 'left', border: 'none', background: 'none', borderRadius: 4, cursor: 'pointer', color: '#DC2626', display: 'flex', alignItems: 'center', gap: 6 }}
                              >
                                🚫 Do Not Contact
                              </button>
                              <button 
                                className="dropdown-item" 
                                onClick={(e) => handleStatusControl(e, p.lead_id, p.campaign_id, 'BOUNCE')} 
                                style={{ width: '100%', padding: '6px 10px', fontSize: 12, textAlign: 'left', border: 'none', background: 'none', borderRadius: 4, cursor: 'pointer', color: '#DC2626', display: 'flex', alignItems: 'center', gap: 6 }}
                              >
                                💥 Mark Bounced
                              </button>
                            </div>
                          )}
                        </div>
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {/* ─── Compact Table Footer ─── */}
      {!loading && prospects.length > 0 && filteredProspects.length > 0 && (
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', fontSize: 12.5, color: '#64748B' }}>
          <div>
            Showing <strong>1–{filteredProspects.length}</strong> of <strong>{prospects.length}</strong> prospects
          </div>
          <div style={{ display: 'flex', gap: 6 }}>
            <button disabled style={{ padding: '4px 10px', border: '1px solid #E2E8F0', borderRadius: 6, background: '#F8FAFC', color: '#94A3B8', fontSize: 11.5, cursor: 'not-allowed' }}>Prev</button>
            <button disabled style={{ padding: '4px 10px', border: '1px solid #E2E8F0', borderRadius: 6, background: '#F8FAFC', color: '#94A3B8', fontSize: 11.5, cursor: 'not-allowed' }}>Next</button>
          </div>
        </div>
      )}

    </div>
  );
}

ProspectList.propTypes = {
  onSelectProspect: PropTypes.func.isRequired,
  onScheduleFollowUp: PropTypes.func.isRequired,
  refreshCount: PropTypes.number,
  onAddProspectClick: PropTypes.func,
  campaigns: PropTypes.arrayOf(PropTypes.object)
};

import { useState, useEffect, useRef } from 'react';
import PropTypes from 'prop-types';
import {
  completeOutreachActivity,
  deleteOutreachActivity
} from '../../../services/outreachService';
import { 
  Folder, Users, Clock, AlertTriangle, CheckCircle2, Mail, Phone, Calendar, ArrowRight, Plus, AlertCircle, TrendingUp, Zap, Activity, MoreHorizontal, ChevronRight, Target
} from 'lucide-react';
import './OutreachPage.css';

export default function OutreachDashboard({
  analytics,
  onSelectCampaign,
  onNewCampaignClick,
  onViewCampaignsClick,
  onTabChange,
  campaignsCount,
  campaigns,
  refreshCount,
  onEditActivity,
  onActivityChanged
}) {
  const [activeTab, setActiveTab] = useState('ALL');
  const [searchQuery, setSearchQuery] = useState('');
  const [activeDropdownId, setActiveDropdownId] = useState(null);
  const dropdownRef = useRef(null);

  useEffect(() => {
    function handleClickOutside(event) {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target)) {
        setActiveDropdownId(null);
      }
    }
    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  async function handleComplete(e, id) {
    e.stopPropagation();
    setActiveDropdownId(null);
    try {
      await completeOutreachActivity(id);
      onActivityChanged();
    } catch (err) {
      window.alert(err.message || 'Failed to complete activity.');
    }
  }

  function handleEdit(e, act) {
    e.stopPropagation();
    setActiveDropdownId(null);
    onEditActivity(act);
  }

  async function handleDelete(e, id) {
    e.stopPropagation();
    setActiveDropdownId(null);
    if (!window.confirm('Delete this outreach activity? This cannot be undone.')) return;
    try {
      await deleteOutreachActivity(id);
      onActivityChanged();
    } catch (err) {
      window.alert(err.message || 'Failed to delete activity.');
    }
  }

  function formatDateTime(dateStr) {
    if (!dateStr) return '—';
    const date = new Date(dateStr);
    return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' });
  }

  function formatRelativeTime(dateStr) {
    if (!dateStr) return '';
    const date = new Date(dateStr);
    const now = new Date();
    const diffMs = now - date;
    const diffDays = Math.floor(diffMs / (1000 * 60 * 60 * 24));
    if (diffDays === 0) return 'Today';
    if (diffDays === 1) return 'Yesterday';
    if (diffDays < 0) {
      const futureDays = Math.abs(diffDays);
      if (futureDays === 1) return 'Tomorrow';
      return `In ${futureDays} days`;
    }
    return `${diffDays}d ago`;
  }

  function getTypeIcon(type) {
    switch ((type || '').toUpperCase()) {
      case 'EMAIL': return <Mail size={13} />;
      case 'CALL': return <Phone size={13} />;
      case 'MEETING': return <Users size={13} />;
      default: return <Clock size={13} />;
    }
  }

  function getTypeColor(type) {
    switch ((type || '').toUpperCase()) {
      case 'EMAIL': return { bg: '#EFF6FF', color: '#2563EB', border: '#BFDBFE' };
      case 'CALL': return { bg: '#ECFDF5', color: '#059669', border: '#A7F3D0' };
      case 'MEETING': return { bg: '#F5F3FF', color: '#7C3AED', border: '#DDD6FE' };
      default: return { bg: '#FFF7ED', color: '#C2410C', border: '#FED7AA' };
    }
  }

  // Analytics data
  const typeEmail    = analytics?.type_email    || 0;
  const typeCall     = analytics?.type_call     || 0;
  const typeFollowup = analytics?.type_followup || 0;
  const typeMeeting  = analytics?.type_meeting  || 0;
  const typeOther    = analytics?.type_other    || 0;
  const totalType    = typeEmail + typeCall + typeFollowup + typeMeeting + typeOther;

  const pEmail   = totalType > 0 ? (typeEmail   / totalType) * 100 : 0;
  const pCall    = totalType > 0 ? (typeCall    / totalType) * 100 : 0;
  const pFollow  = totalType > 0 ? (typeFollowup / totalType) * 100 : 0;
  const pMeeting = totalType > 0 ? (typeMeeting / totalType) * 100 : 0;
  const pOther   = totalType > 0 ? (typeOther   / totalType) * 100 : 0;

  const callOffset    = -pEmail;
  const followOffset  = -(pEmail + pCall);
  const meetingOffset = -(pEmail + pCall + pFollow);
  const otherOffset   = -(pEmail + pCall + pFollow + pMeeting);

  const statusPending    = analytics?.status_pending    || 0;
  const statusInProgress = analytics?.status_in_progress || 0;
  const statusCompleted  = analytics?.status_completed  || 0;
  const statusOverdue    = analytics?.status_overdue    || 0;
  const totalStatus      = statusPending + statusInProgress + statusCompleted + statusOverdue || 1;

  const wPending    = (statusPending    / totalStatus) * 100;
  const wInProgress = (statusInProgress / totalStatus) * 100;
  const wCompleted  = (statusCompleted  / totalStatus) * 100;
  const wOverdue    = (statusOverdue    / totalStatus) * 100;

  const recentActivities   = analytics?.recent_activities   || [];
  const overdueItems        = analytics?.overdue_items        || [];
  const upcomingFollowups   = analytics?.upcoming_followups   || [];

  const overdueCount   = analytics?.overdue        || 0;
  const repliesCount   = analytics?.replies_count  || 0;
  const engagedCount   = analytics?.engaged_prospects || 0;
  const pooledActivities = [
    ...(analytics?.overdue_items || []),
    ...(analytics?.upcoming_followups || []),
    ...(analytics?.recent_activities || [])
  ];
  const uniqueMap = new Map();
  pooledActivities.forEach(act => {
    if (act && act.id && !uniqueMap.has(act.id)) {
      uniqueMap.set(act.id, act);
    }
  });
  const overdueList = Array.from(uniqueMap.values()).filter(act => {
    if (act.status === 'COMPLETED') return false;
    if (act.status === 'OVERDUE') return true;
    if (act.scheduled_at) {
      return new Date(act.scheduled_at) < new Date();
    }
    return false;
  });
  const realOverdueCount = overdueList.length > 0 ? overdueList.length : overdueCount;
  const hasAlerts = realOverdueCount > 0 || repliesCount > 0 || engagedCount > 0;

  const filteredActivities = recentActivities.filter(act => {
    if (activeTab !== 'ALL') {
      if (activeTab === 'OVERDUE') {
        const isExpired = act.scheduled_at && new Date(act.scheduled_at) < new Date() && act.status !== 'COMPLETED';
        if (act.status !== 'OVERDUE' && !isExpired) return false;
      } else if (act.status !== activeTab) {
        return false;
      }
    }
    if (searchQuery.trim()) {
      const q = searchQuery.toLowerCase();
      return (act.lead_company_name || '').toLowerCase().includes(q) ||
             (act.lead_contact_name || '').toLowerCase().includes(q) ||
             (act.subject || '').toLowerCase().includes(q);
    }
    return true;
  });

  const channelLegend = [
    { label: 'Email',     count: typeEmail,   color: '#3B82F6', pct: pEmail   },
    { label: 'Call',      count: typeCall,    color: '#10B981', pct: pCall    },
    { label: 'Follow-up', count: typeFollowup, color: '#F59E0B', pct: pFollow  },
    { label: 'Meeting',   count: typeMeeting, color: '#8B5CF6', pct: pMeeting },
  ].filter(l => l.count > 0);

  return (
    <div className="outreach-dashboard-view" style={{ display: 'flex', flexDirection: 'column', gap: 18 }}>      {/* ─── ROW 1: Needs Attention + Upcoming Follow-ups ─── */}
      <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 18 }}>

        {/* Needs Attention */}
        <div style={{ background: '#FFFFFF', border: '1px solid #E2E8F0', borderRadius: 16, padding: 22, boxShadow: '0 2px 10px rgba(15,23,42,0.04)', display: 'flex', flexDirection: 'column', gap: 14 }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
            <div style={{ width: 36, height: 36, borderRadius: 10, background: realOverdueCount > 0 ? '#FEF2F2' : '#F0FDF4', border: `1px solid ${realOverdueCount > 0 ? '#FECACA' : '#BBF7D0'}`, display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
              {realOverdueCount > 0 ? <AlertCircle size={19} style={{ color: '#DC2626' }} /> : <CheckCircle2 size={19} style={{ color: '#16A34A' }} />}
            </div>
            <div>
              <h4 style={{ margin: 0, fontSize: 16, fontWeight: 800, color: '#0F172A' }}>Needs Attention</h4>
              {realOverdueCount > 0 ? (
                <p style={{ margin: 0, fontSize: 12, color: '#DC2626', fontWeight: 650 }}>{realOverdueCount} item{realOverdueCount !== 1 ? 's' : ''} require immediate action</p>
              ) : (
                <p style={{ margin: 0, fontSize: 12, color: '#16A34A', fontWeight: 650 }}>All tasks complete</p>
              )}
            </div>
          </div>

          {!hasAlerts ? (
            <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', flex: 1, padding: '24px', border: '1.5px dashed #D1FAE5', borderRadius: 12, background: '#F0FDF4', textAlign: 'center' }}>
              <CheckCircle2 size={28} style={{ color: '#16A34A', marginBottom: 6 }} />
              <span style={{ fontSize: 14, fontWeight: 750, color: '#15803D' }}>All caught up!</span>
              <span style={{ fontSize: 12, color: '#4ADE80', marginTop: 2 }}>No tasks requiring action</span>
            </div>
          ) : (
            <div style={{ display: 'flex', flexDirection: 'column', gap: 10, flex: 1, maxHeight: 220, overflowY: 'auto' }} className="scrollable-feed">
              {/* Overdue items as actual cards */}
              {overdueList.map(item => {
                const tc = getTypeColor(item.activity_type);
                return (
                  <div
                    key={item.id}
                    style={{ display: 'flex', alignItems: 'center', gap: 12, background: '#FEF2F2', border: '1px solid #FECACA', borderRadius: 12, padding: '11px 14px', cursor: 'pointer', transition: 'all 0.15s ease' }}
                    onClick={() => onEditActivity(item)}
                    title="Click to view / edit task"
                  >
                    <div style={{ width: 34, height: 34, borderRadius: 9, background: tc.bg, border: `1px solid ${tc.border}`, display: 'flex', alignItems: 'center', justifyContent: 'center', color: tc.color, flexShrink: 0 }}>
                      {getTypeIcon(item.activity_type)}
                    </div>
                    <div style={{ flex: 1, minWidth: 0 }}>
                      <div style={{ fontWeight: 750, fontSize: 13.5, color: '#991B1B', whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis' }}>
                        {item.lead_company_name || 'Individual Prospect'}
                      </div>
                      <div style={{ fontSize: 12, color: '#B91C1C', marginTop: 1, whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis' }}>
                        {item.subject}
                      </div>
                    </div>
                    <div style={{ textAlign: 'right', flexShrink: 0 }}>
                      <div style={{ fontSize: 10.5, fontWeight: 800, color: '#DC2626', background: '#FEE2E2', padding: '3px 8px', borderRadius: 6, letterSpacing: '0.03em' }}>
                        OVERDUE
                      </div>
                      <div style={{ fontSize: 11, color: '#B91C1C', marginTop: 3, fontWeight: 650 }}>{formatRelativeTime(item.scheduled_at)}</div>
                    </div>
                  </div>
                );
              })}

              {repliesCount > 0 && (
                <div
                  style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', background: '#EFF6FF', border: '1px solid #BFDBFE', padding: '11px 14px', borderRadius: 12, fontSize: 13.5, color: '#1E40AF', fontWeight: 600, cursor: 'pointer' }}
                  onClick={() => onTabChange && onTabChange('PROSPECTS')}
                >
                  <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                    <Mail size={15} style={{ color: '#2563EB' }} />
                    <span><strong>{repliesCount} Reply/Replies</strong> received — review now.</span>
                  </div>
                  <ChevronRight size={15} />
                </div>
              )}

              {engagedCount > 0 && (
                <div
                  style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', background: '#ECFDF5', border: '1px solid #A7F3D0', padding: '11px 14px', borderRadius: 12, fontSize: 13.5, color: '#065F46', fontWeight: 600, cursor: 'pointer' }}
                  onClick={() => onTabChange && onTabChange('PROSPECTS')}
                >
                  <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                    <Target size={15} style={{ color: '#059669' }} />
                    <span><strong>{engagedCount} Engaged Prospect{engagedCount > 1 ? 's' : ''}</strong> reached conversion.</span>
                  </div>
                  <ChevronRight size={15} />
                </div>
              )}
            </div>
          )}

          {/* Panel Footer */}
          <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', paddingTop: 12, borderTop: '1px solid #FEE2E2', marginTop: 'auto' }}>
            <span style={{ fontSize: 12, color: '#991B1B', fontWeight: 650, display: 'inline-flex', alignItems: 'center', gap: 5 }}>
              💡 Click any card to resolve or update task
            </span>
            <button 
              onClick={() => onTabChange && onTabChange('FOLLOW_UPS')} 
              style={{ background: 'none', border: 'none', color: '#DC2626', fontSize: 12.5, fontWeight: 800, cursor: 'pointer', display: 'inline-flex', alignItems: 'center', gap: 5 }}
            >
              Go to Follow-ups <ArrowRight size={13} />
            </button>
          </div>
        </div>

        {/* Upcoming Follow-ups */}
        <div style={{ background: '#FFFFFF', border: '1px solid #E2E8F0', borderRadius: 16, padding: 22, boxShadow: '0 2px 10px rgba(15,23,42,0.04)', display: 'flex', flexDirection: 'column', gap: 14 }}>
          <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
            <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
              <div style={{ width: 36, height: 36, borderRadius: 10, background: '#F5F3FF', border: '1px solid #DDD6FE', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
                <Clock size={19} style={{ color: '#7C3AED' }} />
              </div>
              <div>
                <h4 style={{ margin: 0, fontSize: 16, fontWeight: 800, color: '#0F172A' }}>Upcoming Follow-ups</h4>
                <p style={{ margin: 0, fontSize: 12, color: '#64748B', fontWeight: 600 }}>Scheduled outreach tasks queue</p>
              </div>
            </div>
            <button
              onClick={() => onTabChange && onTabChange('FOLLOW_UPS')}
              style={{ fontSize: 12.5, background: 'none', border: 'none', color: '#2563EB', cursor: 'pointer', fontWeight: 800, display: 'inline-flex', alignItems: 'center', gap: 4, padding: '4px 10px', borderRadius: 6, transition: 'background 0.15s' }}
            >
              View All <ArrowRight size={13} />
            </button>
          </div>

          <div style={{ display: 'flex', flexDirection: 'column', gap: 10, flex: 1, maxHeight: 220, overflowY: 'auto' }} className="scrollable-feed">
            {upcomingFollowups.length === 0 ? (
              <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', flex: 1, padding: '24px', border: '1.5px dashed #DDD6FE', borderRadius: 12, background: '#FAF5FF', textAlign: 'center' }}>
                <Calendar size={26} style={{ color: '#8B5CF6', marginBottom: 6 }} />
                <span style={{ fontSize: 13.5, fontWeight: 750, color: '#6D28D9' }}>No upcoming follow-ups</span>
                <span style={{ fontSize: 12, color: '#A78BFA', marginTop: 2 }}>Schedule tasks in the Follow-ups tab</span>
              </div>
            ) : (
              upcomingFollowups.map(item => {
                const tc = getTypeColor(item.activity_type);
                const priorityColor = item.priority === 'HIGH' ? '#DC2626' : item.priority === 'MEDIUM' ? '#D97706' : '#6B7280';
                const priorityBg   = item.priority === 'HIGH' ? '#FEF2F2' : item.priority === 'MEDIUM' ? '#FFFBEB' : '#F9FAFB';
                return (
                  <div
                    key={item.id}
                    style={{ display: 'flex', alignItems: 'center', gap: 12, padding: '11px 14px', border: '1px solid #F1F5F9', borderRadius: 12, cursor: 'pointer', background: '#FAFAFA', transition: 'all 0.15s ease' }}
                    onClick={() => onEditActivity(item)}
                  >
                    <div style={{ width: 34, height: 34, borderRadius: 9, background: tc.bg, border: `1px solid ${tc.border}`, display: 'flex', alignItems: 'center', justifyContent: 'center', color: tc.color, flexShrink: 0 }}>
                      {getTypeIcon(item.activity_type)}
                    </div>
                    <div style={{ flex: 1, minWidth: 0 }}>
                      <div style={{ fontWeight: 750, fontSize: 13.5, color: '#0F172A', whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis' }}>
                        {item.lead_company_name || 'Individual Prospect'}
                      </div>
                      <div style={{ fontSize: 12, color: '#64748B', marginTop: 1, whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis' }}>
                        {item.subject}
                      </div>
                    </div>
                    <div style={{ textAlign: 'right', flexShrink: 0 }}>
                      <div style={{ fontSize: 11.5, color: '#334155', fontWeight: 650 }}>{formatDateTime(item.scheduled_at)}</div>
                      {item.priority && (
                        <span style={{ fontSize: 10.5, fontWeight: 800, color: priorityColor, background: priorityBg, padding: '2px 7px', borderRadius: 6, display: 'inline-block', marginTop: 3, textTransform: 'uppercase', letterSpacing: '0.04em' }}>
                          {item.priority}
                        </span>
                      )}
                    </div>
                  </div>
                );
              })
            )}
          </div>

          {/* Panel Footer */}
          <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', paddingTop: 12, borderTop: '1px solid #F1F5F9', marginTop: 'auto' }}>
            <span style={{ fontSize: 12, color: '#64748B', fontWeight: 650, display: 'inline-flex', alignItems: 'center', gap: 5 }}>
              📅 Scheduled outreach task reminders
            </span>
            <button 
              onClick={() => onTabChange && onTabChange('FOLLOW_UPS')} 
              style={{ background: 'none', border: 'none', color: '#2563EB', fontSize: 12.5, fontWeight: 800, cursor: 'pointer', display: 'inline-flex', alignItems: 'center', gap: 5 }}
            >
              Manage Tasks <ArrowRight size={13} />
            </button>
          </div>
        </div>
      </div>

      {/* ─── ROW 2: Outreach by Channel + Outreach Status ─── */}
      <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 18 }}>

        {/* ── Outreach by Channel ── */}
        <div style={{ background: 'linear-gradient(145deg, #FFFFFF 0%, #F8FAFF 100%)', border: '1px solid #E2E8F0', borderRadius: 16, padding: '22px 24px', boxShadow: '0 2px 10px rgba(15,23,42,0.05)' }}>
          {/* Header */}
          <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: 18 }}>
            <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
              <div style={{ width: 36, height: 36, borderRadius: 10, background: 'linear-gradient(135deg, #3B82F6 0%, #6366F1 100%)', display: 'flex', alignItems: 'center', justifyContent: 'center', boxShadow: '0 3px 8px rgba(59,130,246,0.3)' }}>
                <Activity size={18} style={{ color: '#FFFFFF' }} />
              </div>
              <div>
                <h4 style={{ margin: 0, fontSize: 16, fontWeight: 800, color: '#0F172A' }}>Outreach by Channel</h4>
                <p style={{ margin: 0, fontSize: 12, color: '#64748B', fontWeight: 600 }}>{totalType} total activities logged</p>
              </div>
            </div>
          </div>

          {totalType === 0 ? (
            <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', padding: '30px 20px', border: '1.5px dashed #BFDBFE', borderRadius: 12, background: '#F0F7FF', textAlign: 'center' }}>
              <TrendingUp size={30} style={{ color: '#93C5FD', marginBottom: 8 }} />
              <span style={{ fontSize: 13.5, color: '#3B82F6', fontWeight: 750 }}>No activities yet</span>
              <span style={{ fontSize: 12, color: '#64748B', marginTop: 3 }}>Log your first outreach to see stats</span>
            </div>
          ) : (
            <div style={{ display: 'flex', gap: 24, alignItems: 'center' }}>
              {/* Large Doughnut */}
              <div style={{ position: 'relative', width: 120, height: 120, flexShrink: 0 }}>
                <svg viewBox="0 0 36 36" style={{ width: '100%', height: '100%', transform: 'rotate(-90deg)', filter: 'drop-shadow(0 2px 6px rgba(0,0,0,0.08))' }}>
                  <path stroke="#F1F5F9" strokeWidth="4" fill="none" d="M18 2.0845 a 15.9155 15.9155 0 0 1 0 31.831 a 15.9155 15.9155 0 0 1 0 -31.831" />
                  {typeEmail   > 0 && <path stroke="#3B82F6" strokeWidth="4" fill="none" strokeLinecap="round" strokeDasharray={`${pEmail}, 100`}   d="M18 2.0845 a 15.9155 15.9155 0 0 1 0 31.831 a 15.9155 15.9155 0 0 1 0 -31.831" />}
                  {typeCall    > 0 && <path stroke="#10B981" strokeWidth="4" fill="none" strokeLinecap="round" strokeDasharray={`${pCall}, 100`}    strokeDashoffset={callOffset}    d="M18 2.0845 a 15.9155 15.9155 0 0 1 0 31.831 a 15.9155 15.9155 0 0 1 0 -31.831" />}
                  {typeFollowup > 0 && <path stroke="#F59E0B" strokeWidth="4" fill="none" strokeLinecap="round" strokeDasharray={`${pFollow}, 100`}  strokeDashoffset={followOffset}  d="M18 2.0845 a 15.9155 15.9155 0 0 1 0 31.831 a 15.9155 15.9155 0 0 1 0 -31.831" />}
                  {typeMeeting  > 0 && <path stroke="#8B5CF6" strokeWidth="4" fill="none" strokeLinecap="round" strokeDasharray={`${pMeeting}, 100`} strokeDashoffset={meetingOffset} d="M18 2.0845 a 15.9155 15.9155 0 0 1 0 31.831 a 15.9155 15.9155 0 0 1 0 -31.831" />}
                  {typeOther    > 0 && <path stroke="#94A3B8" strokeWidth="4" fill="none" strokeLinecap="round" strokeDasharray={`${pOther}, 100`}   strokeDashoffset={otherOffset}   d="M18 2.0845 a 15.9155 15.9155 0 0 1 0 31.831 a 15.9155 15.9155 0 0 1 0 -31.831" />}
                </svg>
                <div style={{ position: 'absolute', top: '50%', left: '50%', transform: 'translate(-50%, -50%)', textAlign: 'center', lineHeight: 1 }}>
                  <span style={{ fontSize: 24, fontWeight: 900, color: '#0F172A', display: 'block', letterSpacing: '-0.5px' }}>{totalType}</span>
                  <span style={{ fontSize: 10, color: '#64748B', textTransform: 'uppercase', fontWeight: 750, letterSpacing: '0.04em' }}>Total</span>
                </div>
              </div>

              {/* Rich Legend */}
              <div style={{ flex: 1, display: 'flex', flexDirection: 'column', gap: 11 }}>
                {[
                  { label: 'Email',     count: typeEmail,    color: '#3B82F6', bg: '#EFF6FF', pct: pEmail    },
                  { label: 'Call',      count: typeCall,     color: '#10B981', bg: '#ECFDF5', pct: pCall     },
                  { label: 'Follow-up', count: typeFollowup, color: '#F59E0B', bg: '#FFFBEB', pct: pFollow   },
                  { label: 'Meeting',   count: typeMeeting,  color: '#8B5CF6', bg: '#F5F3FF', pct: pMeeting  },
                ].filter(l => l.count > 0).map(l => (
                  <div key={l.label}>
                    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 5 }}>
                      <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                        <div style={{ width: 10, height: 10, borderRadius: 3, background: l.color, flexShrink: 0 }} />
                        <span style={{ fontSize: 13, color: '#334155', fontWeight: 700 }}>{l.label}</span>
                      </div>
                      <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                        <span style={{ fontSize: 11, color: l.color, fontWeight: 800, background: l.bg, padding: '2px 7px', borderRadius: 6 }}>{Math.round(l.pct)}%</span>
                        <span style={{ fontSize: 14, fontWeight: 800, color: '#0F172A', minWidth: 20, textAlign: 'right' }}>{l.count}</span>
                      </div>
                    </div>
                    <div style={{ height: 8, background: '#F1F5F9', borderRadius: 4, overflow: 'hidden' }}>
                      <div style={{ height: '100%', background: `linear-gradient(90deg, ${l.color}cc, ${l.color})`, width: `${Math.max(l.pct, 5)}%`, borderRadius: 4, transition: 'width 0.6s ease' }} />
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )}
        </div>

        {/* ── Activity Status ── */}
        <div style={{ background: 'linear-gradient(145deg, #FFFFFF 0%, #F8FFF8 100%)', border: '1px solid #E2E8F0', borderRadius: 16, padding: '22px 24px', boxShadow: '0 2px 10px rgba(15,23,42,0.05)' }}>
          {/* Header */}
          <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: 18 }}>
            <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
              <div style={{ width: 36, height: 36, borderRadius: 10, background: 'linear-gradient(135deg, #10B981 0%, #059669 100%)', display: 'flex', alignItems: 'center', justifyContent: 'center', boxShadow: '0 3px 8px rgba(16,185,129,0.3)' }}>
                <Zap size={18} style={{ color: '#FFFFFF' }} />
              </div>
              <div>
                <h4 style={{ margin: 0, fontSize: 16, fontWeight: 800, color: '#0F172A' }}>Activity Status</h4>
                <p style={{ margin: 0, fontSize: 12, color: '#64748B', fontWeight: 600 }}>{totalStatus} activities tracked</p>
              </div>
            </div>
            {/* Quick summary pills */}
            <div style={{ display: 'flex', gap: 6 }}>
              {statusOverdue > 0 && <span style={{ fontSize: 11, fontWeight: 800, color: '#DC2626', background: '#FEF2F2', border: '1px solid #FECACA', padding: '3px 8px', borderRadius: 8 }}>🔴 {statusOverdue} Overdue</span>}
              {statusCompleted > 0 && <span style={{ fontSize: 11, fontWeight: 800, color: '#059669', background: '#ECFDF5', border: '1px solid #A7F3D0', padding: '3px 8px', borderRadius: 8 }}>✅ {statusCompleted} Done</span>}
            </div>
          </div>

          <div style={{ display: 'flex', flexDirection: 'column', gap: 14 }}>
            {[
              { label: 'Pending',     count: statusPending,    pct: wPending,    color: '#3B82F6', gradStart: '#60A5FA', bg: '#EFF6FF', icon: '🔵' },
              { label: 'In Progress', count: statusInProgress, pct: wInProgress, color: '#F59E0B', gradStart: '#FCD34D', bg: '#FFFBEB', icon: '🟡' },
              { label: 'Completed',   count: statusCompleted,  pct: wCompleted,  color: '#10B981', gradStart: '#34D399', bg: '#ECFDF5', icon: '🟢' },
              { label: 'Overdue',     count: statusOverdue,    pct: wOverdue,    color: '#EF4444', gradStart: '#F87171', bg: '#FEF2F2', icon: '🔴' },
            ].map(s => (
              <div key={s.label}>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 6 }}>
                  <span style={{ fontSize: 13, color: '#334155', fontWeight: 700 }}>{s.label}</span>
                  <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                    <span style={{ fontSize: 11.5, color: '#64748B', fontWeight: 600 }}>{Math.round(s.pct)}%</span>
                    <span style={{ fontSize: 15, fontWeight: 900, color: s.count > 0 ? s.color : '#94A3B8', minWidth: 22, textAlign: 'right' }}>{s.count}</span>
                  </div>
                </div>
                <div style={{ height: 10, background: '#F1F5F9', borderRadius: 5, overflow: 'hidden' }}>
                  <div style={{
                    height: '100%',
                    background: s.count > 0 ? `linear-gradient(90deg, ${s.gradStart}, ${s.color})` : '#F1F5F9',
                    width: s.count > 0 ? `${Math.max(s.pct, 4)}%` : '0%',
                    borderRadius: 5,
                    transition: 'width 0.6s cubic-bezier(0.4, 0, 0.2, 1)',
                    boxShadow: s.count > 0 ? `0 1px 4px ${s.color}55` : 'none'
                  }} />
                </div>
              </div>
            ))}
          </div>

          {/* Total Summary Row */}
          <div style={{ display: 'flex', gap: 10, marginTop: 18, paddingTop: 14, borderTop: '1px solid #F1F5F9' }}>
            {[
              { v: statusPending,    label: 'Pending',  color: '#3B82F6', bg: '#EFF6FF' },
              { v: statusInProgress, label: 'Active',   color: '#F59E0B', bg: '#FFFBEB' },
              { v: statusCompleted,  label: 'Done',     color: '#10B981', bg: '#ECFDF5' },
              { v: statusOverdue,    label: 'Overdue',  color: '#EF4444', bg: '#FEF2F2' },
            ].map(m => (
              <div key={m.label} style={{ flex: 1, textAlign: 'center', background: m.bg, borderRadius: 10, padding: '8px 6px' }}>
                <div style={{ fontSize: 18, fontWeight: 900, color: m.color, lineHeight: 1 }}>{m.v}</div>
                <div style={{ fontSize: 10.5, color: m.color, fontWeight: 750, marginTop: 3, opacity: 0.9 }}>{m.label}</div>
              </div>
            ))}
          </div>
        </div>
      </div>

      {/* ─── ROW 3: Recent Outreach Timeline ─── */}
      <div style={{ background: '#FFFFFF', border: '1px solid #E2E8F0', borderRadius: 16, padding: 22, boxShadow: '0 2px 10px rgba(15,23,42,0.04)' }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 18 }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
            <div style={{ width: 36, height: 36, borderRadius: 10, background: '#FFF7ED', border: '1px solid #FFEDD5', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
              <Activity size={18} style={{ color: '#EA580C' }} />
            </div>
            <div>
              <h4 style={{ margin: 0, fontSize: 16, fontWeight: 800, color: '#0F172A' }}>Recent Outreach Timeline</h4>
              <p style={{ margin: 0, fontSize: 12.5, color: '#64748B', fontWeight: 600 }}>Chronological trace of communication events</p>
            </div>
          </div>

          <div style={{ display: 'flex', gap: 10, alignItems: 'center' }}>
            <div style={{ background: '#F1F5F9', padding: '4px', borderRadius: 10, display: 'flex', gap: 3 }}>
              {['ALL', 'COMPLETED', 'PENDING'].map(tab => (
                <button
                  key={tab}
                  onClick={() => setActiveTab(tab)}
                  style={{ fontSize: 12, padding: '5px 12px', borderRadius: 7, border: 'none', background: activeTab === tab ? '#FFFFFF' : 'transparent', color: activeTab === tab ? '#0F172A' : '#64748B', fontWeight: activeTab === tab ? 750 : 600, cursor: 'pointer', transition: 'all 0.15s ease', boxShadow: activeTab === tab ? '0 1px 3px rgba(0,0,0,0.08)' : 'none' }}
                >
                  {tab === 'ALL' ? 'All' : tab.replace('_', ' ')}
                </button>
              ))}
            </div>
            <button
              className="outreach-btn outreach-btn-primary"
              style={{ padding: '8px 14px', fontSize: 13, display: 'inline-flex', alignItems: 'center', gap: 6, borderRadius: 9, fontWeight: 750 }}
              onClick={() => onEditActivity(null)}
            >
              <Plus size={15} /> Log Activity
            </button>
          </div>
        </div>

        {filteredActivities.length === 0 ? (
          <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', padding: '36px 20px', border: '1.5px dashed #CBD5E1', borderRadius: 12, background: '#F8FAFC', textAlign: 'center' }}>
            <Clock size={30} style={{ color: '#94A3B8', marginBottom: 8 }} />
            <h5 style={{ margin: '0 0 4px 0', fontSize: 14, fontWeight: 750, color: '#334155' }}>No recent activities</h5>
            <p style={{ margin: 0, fontSize: 12.5, color: '#64748B' }}>Timeline logs will populate here as outreach is conducted.</p>
          </div>
        ) : (
          <div style={{ position: 'relative', paddingLeft: 26, borderLeft: '2px solid #E2E8F0', marginLeft: 14 }}>
            {filteredActivities.slice(0, 8).map((act, idx) => {
              const tc = getTypeColor(act.activity_type);
              return (
                <div key={act.id} style={{ position: 'relative', marginBottom: idx < filteredActivities.slice(0, 8).length - 1 ? 18 : 0 }}>
                  {/* Timeline dot */}
                  <div style={{
                    position: 'absolute', left: -33, top: 12,
                    width: 14, height: 14, borderRadius: '50%',
                    background: act.status === 'COMPLETED' ? '#10B981' : act.status === 'OVERDUE' ? '#EF4444' : '#F59E0B',
                    border: '2.5px solid #FFFFFF',
                    boxShadow: '0 1px 4px rgba(0,0,0,0.15)'
                  }} />

                  <div style={{ background: '#F8FAFC', padding: '14px 16px', borderRadius: 12, border: '1px solid #F1F5F9', transition: 'all 0.15s ease' }}>
                    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', gap: 10 }}>
                      <div style={{ display: 'flex', alignItems: 'center', gap: 10, flex: 1 }}>
                        <div style={{ width: 30, height: 30, borderRadius: 8, background: tc.bg, border: `1px solid ${tc.border}`, display: 'flex', alignItems: 'center', justifyContent: 'center', color: tc.color, flexShrink: 0 }}>
                          {getTypeIcon(act.activity_type)}
                        </div>
                        <div style={{ flex: 1, minWidth: 0 }}>
                          <div style={{ fontWeight: 750, fontSize: 14, color: '#0F172A', whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis' }}>
                            {act.lead_company_name || 'Individual Prospect'}
                            <span style={{ fontWeight: 550, color: '#64748B', marginLeft: 8 }}>· {act.subject}</span>
                          </div>
                          {act.description && (
                            <div style={{ fontSize: 12.5, color: '#475569', marginTop: 4, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                              {act.description}
                            </div>
                          )}
                        </div>
                      </div>

                      <div style={{ display: 'flex', alignItems: 'center', gap: 10, flexShrink: 0 }}>
                        <div style={{ textAlign: 'right' }}>
                          <div style={{ fontSize: 11.5, color: '#64748B', fontWeight: 600 }}>{formatDateTime(act.created_at)}</div>
                          <span className={`activity-status-pill-grad ${act.status.toLowerCase()}`} style={{ fontSize: 10.5, padding: '3px 8px', marginTop: 3, display: 'inline-block', borderRadius: 10, fontWeight: 750, textTransform: 'uppercase' }}>
                            {act.status}
                          </span>
                        </div>
                        <div ref={activeDropdownId === act.id ? dropdownRef : null} style={{ position: 'relative' }}>
                          <button
                            onClick={(e) => { e.stopPropagation(); setActiveDropdownId(activeDropdownId === act.id ? null : act.id); }}
                            style={{ width: 28, height: 28, borderRadius: 7, border: '1px solid #E2E8F0', background: '#FFFFFF', cursor: 'pointer', display: 'flex', alignItems: 'center', justifyContent: 'center', color: '#64748B' }}
                          >
                            <MoreHorizontal size={14} />
                          </button>
                          {activeDropdownId === act.id && (
                            <div style={{ position: 'absolute', right: 0, top: 32, background: '#FFFFFF', border: '1px solid #E2E8F0', borderRadius: 10, boxShadow: '0 6px 20px rgba(0,0,0,0.12)', zIndex: 200, padding: 5, minWidth: 160 }}>
                              <button onClick={(e) => handleEdit(e, act)} style={{ width: '100%', padding: '7px 12px', fontSize: 12.5, textAlign: 'left', border: 'none', background: 'none', borderRadius: 6, cursor: 'pointer', color: '#334155', display: 'flex', alignItems: 'center', gap: 8, fontWeight: 600 }}>✏️ Edit</button>
                              {act.status !== 'COMPLETED' && <button onClick={(e) => handleComplete(e, act.id)} style={{ width: '100%', padding: '7px 12px', fontSize: 12.5, textAlign: 'left', border: 'none', background: 'none', borderRadius: 6, cursor: 'pointer', color: '#059669', display: 'flex', alignItems: 'center', gap: 8, fontWeight: 600 }}>✅ Mark Complete</button>}
                              <div style={{ height: 1, background: '#F1F5F9', margin: '4px 0' }} />
                              <button onClick={(e) => handleDelete(e, act.id)} style={{ width: '100%', padding: '7px 12px', fontSize: 12.5, textAlign: 'left', border: 'none', background: 'none', borderRadius: 6, cursor: 'pointer', color: '#DC2626', display: 'flex', alignItems: 'center', gap: 8, fontWeight: 600 }}>🗑️ Delete</button>
                            </div>
                          )}
                        </div>
                      </div>
                    </div>
                  </div>
                </div>
              );
            })}
          </div>
        )}
      </div>

    </div>
  );
}

OutreachDashboard.propTypes = {
  analytics: PropTypes.object,
  onSelectCampaign: PropTypes.func.isRequired,
  onNewCampaignClick: PropTypes.func.isRequired,
  onViewCampaignsClick: PropTypes.func.isRequired,
  onTabChange: PropTypes.func,
  campaignsCount: PropTypes.object,
  campaigns: PropTypes.array,
  refreshCount: PropTypes.number,
  onEditActivity: PropTypes.func.isRequired,
  onActivityChanged: PropTypes.func.isRequired
};

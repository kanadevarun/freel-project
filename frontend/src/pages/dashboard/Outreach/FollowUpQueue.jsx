import { useState, useEffect, useCallback } from 'react';
import PropTypes from 'prop-types';
import { getFollowUps, completeFollowUp, cancelFollowUp, rescheduleFollowUp } from '../../../services/outreachService';
import { Search, Calendar, Mail, Phone, Users, Clock, Check, XCircle, ExternalLink, AlertCircle, CalendarClock, Tag, ChevronRight } from 'lucide-react';
import './OutreachPage.css';

export default function FollowUpQueue({ onOpenProspect, onNewFollowUpClick }) {
  const [activeTab, setActiveTab] = useState('TODAY'); // 'ALL', 'OVERDUE', 'TODAY', 'UPCOMING', 'COMPLETED'
  const [followUps, setFollowUps] = useState([]);
  const [loading, setLoading] = useState(true);
  const [searchQuery, setSearchQuery] = useState('');
  const [reschedulingId, setReschedulingId] = useState(null);
  const [rescheduleDate, setRescheduleDate] = useState('');
  const [error, setError] = useState('');

  // Fetch all follow-ups at once to calculate correct counts on the frontend
  const fetchFollowUps = useCallback(async () => {
    setLoading(true);
    setError('');
    try {
      const data = await getFollowUps(''); // empty filter fetches all
      setFollowUps(data || []);
    } catch (err) {
      console.error('Failed to fetch follow-ups:', err);
      setError('Failed to fetch follow-up queue.');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchFollowUps();
  }, [fetchFollowUps]);

  async function handleComplete(e, id) {
    e.stopPropagation();
    try {
      await completeFollowUp(id);
      await fetchFollowUps();
    } catch (err) {
      alert(`Complete action failed: ${err.message || 'Server error'}`);
    }
  }

  async function handleCancel(e, id) {
    e.stopPropagation();
    if (!window.confirm('Are you sure you want to cancel this follow-up task?')) return;
    try {
      await cancelFollowUp(id);
      await fetchFollowUps();
    } catch (err) {
      alert(`Cancel action failed: ${err.message || 'Server error'}`);
    }
  }

  async function handleRescheduleSubmit(e, id) {
    e.stopPropagation();
    if (!rescheduleDate) return;
    try {
      const formattedDate = new Date(rescheduleDate).toISOString();
      await rescheduleFollowUp(id, formattedDate);
      setReschedulingId(null);
      setRescheduleDate('');
      await fetchFollowUps();
    } catch (err) {
      alert(`Reschedule action failed: ${err.message || 'Server error'}`);
    }
  }

  function getDueStatus(scheduledAtStr, status) {
    if (status === 'COMPLETED') return 'COMPLETED';
    const now = new Date();
    const scheduled = new Date(scheduledAtStr);
    
    // Check if overdue (scheduled date is in the past)
    if (scheduled < now && scheduled.toDateString() !== now.toDateString()) {
      return 'OVERDUE';
    }
    
    // Check if scheduled for today
    if (scheduled.toDateString() === now.toDateString()) {
      return 'TODAY';
    }
    
    return 'UPCOMING';
  }

  function formatDateTime(dateStr) {
    if (!dateStr) return '—';
    const date = new Date(dateStr);
    const now = new Date();
    
    const timeStr = date.toLocaleTimeString('en-US', {
      hour: '2-digit',
      minute: '2-digit'
    });

    if (date.toDateString() === now.toDateString()) {
      return `Today, ${timeStr}`;
    }
    
    const tomorrow = new Date(now);
    tomorrow.setDate(now.getDate() + 1);
    if (date.toDateString() === tomorrow.toDateString()) {
      return `Tomorrow, ${timeStr}`;
    }

    return date.toLocaleDateString('en-US', {
      month: 'short',
      day: 'numeric',
      year: 'numeric'
    }) + `, ${timeStr}`;
  }

  function getOverdueDays(dateStr) {
    const now = new Date();
    const date = new Date(dateStr);
    const diffTime = Math.abs(now - date);
    const diffDays = Math.ceil(diffTime / (1000 * 60 * 60 * 24));
    return diffDays === 1 ? 'Overdue by 1 day' : `Overdue by ${diffDays} days`;
  }

  // Frontend categorization logic
  const itemsWithStatus = followUps.map(item => ({
    ...item,
    dueStatus: getDueStatus(item.scheduled_at, item.status)
  }));

  // Filtering by tab
  const tabFilteredItems = itemsWithStatus.filter(item => {
    if (activeTab === 'ALL') return true;
    return item.dueStatus === activeTab;
  });

  // Filtering by search query
  const filteredItems = tabFilteredItems.filter(item => {
    const company = (item.lead_company_name || item.customer_company_name || '').toLowerCase();
    const contact = (item.lead_contact_name || item.customer_contact_name || '').toLowerCase();
    const subject = (item.subject || '').toLowerCase();
    const notes = (item.description || '').toLowerCase();
    const query = searchQuery.toLowerCase();
    
    return company.includes(query) || contact.includes(query) || subject.includes(query) || notes.includes(query);
  });

  // Calculate dynamic tab counts based on actual loaded items
  const counts = {
    ALL: itemsWithStatus.length,
    OVERDUE: itemsWithStatus.filter(i => i.dueStatus === 'OVERDUE').length,
    TODAY: itemsWithStatus.filter(i => i.dueStatus === 'TODAY').length,
    UPCOMING: itemsWithStatus.filter(i => i.dueStatus === 'UPCOMING').length,
    COMPLETED: itemsWithStatus.filter(i => i.dueStatus === 'COMPLETED').length,
  };

  return (
    <div className="followup-queue-workspace" style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
      
      {/* ─── CRM Toolbar & Search ─── */}
      <div className="campaign-list-card" style={{ background: '#FFFFFF', border: '1px solid #E2E8F0', borderRadius: 12, padding: 18, boxShadow: '0 1px 3px rgba(0,0,0,0.05)' }}>
        <div style={{ display: 'flex', gap: 12, alignItems: 'center', justifyContent: 'space-between', flexWrap: 'wrap' }}>
          
          {/* Tabs switch segment */}
          <div className="activities-status-tabs shadow-capsule" style={{ background: '#F1F5F9', padding: 3, borderRadius: 8, display: 'flex', gap: 2 }}>
            <button
              onClick={() => setActiveTab('ALL')}
              className={`status-tab-btn ${activeTab === 'ALL' ? 'active' : ''}`}
              style={{ fontSize: 12, padding: '6px 14px', borderRadius: 6 }}
            >
              All <span style={{ fontSize: 10.5, marginLeft: 4, fontWeight: 700, opacity: 0.75 }}>({counts.ALL})</span>
            </button>
            <button
              onClick={() => setActiveTab('OVERDUE')}
              className={`status-tab-btn ${activeTab === 'OVERDUE' ? 'active' : ''}`}
              style={{ 
                fontSize: 12, 
                padding: '6px 14px', 
                borderRadius: 6,
                color: activeTab === 'OVERDUE' ? '#DC2626' : (counts.OVERDUE > 0 ? '#EF4444' : '#64748B'),
                fontWeight: counts.OVERDUE > 0 ? 700 : 500
              }}
            >
              Overdue <span style={{ fontSize: 10.5, marginLeft: 4, fontWeight: 750 }}>({counts.OVERDUE})</span>
            </button>
            <button
              onClick={() => setActiveTab('TODAY')}
              className={`status-tab-btn ${activeTab === 'TODAY' ? 'active' : ''}`}
              style={{ 
                fontSize: 12, 
                padding: '6px 14px', 
                borderRadius: 6,
                color: activeTab === 'TODAY' ? '#D97706' : (counts.TODAY > 0 ? '#F59E0B' : '#64748B'),
                fontWeight: counts.TODAY > 0 ? 700 : 500
              }}
            >
              Due Today <span style={{ fontSize: 10.5, marginLeft: 4, fontWeight: 750 }}>({counts.TODAY})</span>
            </button>
            <button
              onClick={() => setActiveTab('UPCOMING')}
              className={`status-tab-btn ${activeTab === 'UPCOMING' ? 'active' : ''}`}
              style={{ fontSize: 12, padding: '6px 14px', borderRadius: 6 }}
            >
              Upcoming <span style={{ fontSize: 10.5, marginLeft: 4, fontWeight: 700, opacity: 0.75 }}>({counts.UPCOMING})</span>
            </button>
            <button
              onClick={() => setActiveTab('COMPLETED')}
              className={`status-tab-btn ${activeTab === 'COMPLETED' ? 'active' : ''}`}
              style={{ fontSize: 12, padding: '6px 14px', borderRadius: 6 }}
            >
              Completed <span style={{ fontSize: 10.5, marginLeft: 4, fontWeight: 700, opacity: 0.75 }}>({counts.COMPLETED})</span>
            </button>
          </div>

          {/* Search Input bar */}
          <div className="search-box-wrapper" style={{ flex: 1, minWidth: 260, maxWidth: 360, position: 'relative', display: 'flex', alignItems: 'center' }}>
            <Search size={14} style={{ position: 'absolute', left: 12, color: '#64748B' }} />
            <input
              type="text"
              placeholder="Search follow-ups by company, subject, notes..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="campaign-search-input"
              style={{ width: '100%', paddingLeft: 34, height: 38, fontSize: 13, border: '1px solid #CBD5E1', borderRadius: 8 }}
            />
          </div>

        </div>
      </div>

      {/* ─── Loading State Skeletons ─── */}
      {loading ? (
        <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
          {Array.from({ length: 3 }).map((_, idx) => (
            <div key={idx} style={{ height: 90, background: '#FFFFFF', border: '1px solid #E2E8F0', borderRadius: 12, display: 'flex', alignItems: 'center', padding: 16, boxSizing: 'border-box' }}>
              <div style={{ width: 36, height: 36, borderRadius: '50%', background: '#F1F5F9', marginRight: 12 }} />
              <div style={{ flex: 1 }}>
                <div style={{ width: '30%', height: 12, background: '#F1F5F9', borderRadius: 4, marginBottom: 8 }} />
                <div style={{ width: '60%', height: 10, background: '#F8FAFC', borderRadius: 4 }} />
              </div>
            </div>
          ))}
        </div>
      ) : error ? (
        <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', padding: '50px 20px', border: '1px solid #FEE2E2', borderRadius: 12, background: '#FEF2F2', textAlign: 'center' }}>
          <AlertCircle size={40} style={{ color: '#EF4444', marginBottom: 12 }} />
          <h4 style={{ fontSize: 15.5, fontWeight: 750, color: '#991B1B', margin: '0 0 6px 0' }}>Unable to load follow-ups</h4>
          <p style={{ color: '#7F1D1D', fontSize: 13, maxWidth: 440, lineHeight: 1.5, margin: '0 0 16px 0' }}>
            We encountered a connection issue while fetching your follow-up queue. Please verify your connection or retry the request.
          </p>
          <div style={{ display: 'flex', gap: 10 }}>
            <button
              onClick={fetchFollowUps}
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
      ) : filteredItems.length === 0 ? (
        /* ─── Polished Empty State ─── */
        <div className="outreach-empty-state" style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', padding: '60px 20px', border: '1.5px dashed #CBD5E1', borderRadius: 12, textAlign: 'center', background: '#FFFFFF' }}>
          <CalendarClock size={48} style={{ color: '#94A3B8', marginBottom: 16 }} />
          <h4 style={{ fontSize: 16, fontWeight: 750, color: '#0F172A', margin: '0 0 6px 0' }}>
            {activeTab === 'ALL' ? 'No follow-ups yet' : "You're all caught up"}
          </h4>
          <p style={{ color: '#64748B', fontSize: 13.5, maxWidth: 420, lineHeight: 1.5, margin: '0 0 20px 0' }}>
            {activeTab === 'ALL' 
              ? 'Schedule follow-up tasks for your prospects to guide them through the sales pipeline.' 
              : 'There are no follow-ups requiring your attention under this status filter.'}
          </p>
          {onNewFollowUpClick && (
            <button
              className="outreach-btn outreach-btn-primary"
              onClick={onNewFollowUpClick}
              style={{ padding: '9px 18px', fontSize: 13, borderRadius: 8, fontWeight: 600 }}
            >
              + Create Follow-up
            </button>
          )}
        </div>
      ) : (
        /* ─── Task Queue List Cards ─── */
        <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
          {filteredItems.map(item => {
            const clientName = item.lead_company_name || item.customer_company_name || 'Prospect';
            const contactPerson = item.lead_contact_name || item.customer_contact_name;
            const leadId = item.lead_id;
            const isCompleted = item.status === 'COMPLETED';

            // Left side borders accent indicators for priority / urgency
            let borderStyle = {};
            if (!isCompleted) {
              if (item.dueStatus === 'OVERDUE') {
                borderStyle = { borderLeft: '4px solid #EF4444' };
              } else if (item.dueStatus === 'TODAY') {
                borderStyle = { borderLeft: '4px solid #F59E0B' };
              } else {
                borderStyle = { borderLeft: '4px solid #3B82F6' };
              }
            } else {
              borderStyle = { borderLeft: '4px solid #10B981', opacity: 0.7 };
            }

            // Get Priority badge styles
            let priorityBadgeColor = { bg: '#F1F5F9', text: '#475569' };
            if (item.priority === 'HIGH') {
              priorityBadgeColor = { bg: '#FEF2F2', text: '#DC2626' };
            } else if (item.priority === 'MEDIUM') {
              priorityBadgeColor = { bg: '#FFFBEB', text: '#D97706' };
            }

            return (
              <div
                key={item.id}
                style={{
                  background: '#FFFFFF',
                  border: '1px solid #E2E8F0',
                  borderRadius: 12,
                  padding: '16px 20px',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'space-between',
                  gap: 16,
                  boxShadow: '0 1px 3px rgba(0,0,0,0.02)',
                  transition: 'transform 0.15s, box-shadow 0.15s',
                  ...borderStyle
                }}
                className="premium-table-row"
              >
                
                {/* Left side: Task Type Icon and Contact info */}
                <div style={{ display: 'flex', alignItems: 'center', gap: 14, flex: 1, minWidth: 260 }}>
                  <div style={{
                    width: 38,
                    height: 38,
                    borderRadius: '50%',
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    background: isCompleted ? '#ECFDF5' : (item.dueStatus === 'OVERDUE' ? '#FEF2F2' : '#EFF6FF'),
                    color: isCompleted ? '#10B981' : (item.dueStatus === 'OVERDUE' ? '#EF4444' : '#2563EB'),
                    flexShrink: 0
                  }}>
                    {item.activity_type === 'EMAIL' ? <Mail size={16} /> : item.activity_type === 'CALL' ? <Phone size={16} /> : item.activity_type === 'MEETING' ? <Users size={16} /> : <Clock size={16} />}
                  </div>

                  <div>
                    {leadId ? (
                      <div 
                        style={{ cursor: 'pointer', fontWeight: 750, color: '#2563EB', fontSize: 14.5, display: 'inline-flex', alignItems: 'center', gap: 4 }}
                        onClick={() => onOpenProspect(leadId)}
                      >
                        {clientName} <ChevronRight size={13} style={{ color: '#94A3B8' }} />
                      </div>
                    ) : (
                      <div style={{ fontWeight: 750, color: '#0F172A', fontSize: 14.5 }}>{clientName}</div>
                    )}
                    
                    {contactPerson && (
                      <div style={{ fontSize: 12, color: '#64748B', marginTop: 3 }}>
                        {contactPerson}
                      </div>
                    )}
                  </div>
                </div>

                {/* Center Section: Notes, Subject, Details */}
                <div style={{ flex: 2, minWidth: 320 }}>
                  <div style={{ fontWeight: 700, fontSize: 14, color: isCompleted ? '#64748B' : '#0F172A', textDecoration: isCompleted ? 'line-through' : 'none' }}>
                    {item.subject}
                  </div>
                  {item.description && (
                    <div style={{ fontSize: 12.5, color: '#475569', marginTop: 4, lineHeight: 1.4 }}>
                      {item.description}
                    </div>
                  )}
                  {leadId && (
                    <div style={{ display: 'flex', gap: 10, marginTop: 6, alignItems: 'center' }}>
                      <a 
                        href={`/dashboard/leads?leadId=${leadId}`}
                        style={{ fontSize: 11, color: '#64748B', fontWeight: 600, display: 'inline-flex', alignItems: 'center', textDecoration: 'none' }}
                        className="hover-underline-link"
                      >
                        Lead Record <ExternalLink size={10} style={{ marginLeft: 3 }} />
                      </a>
                      <span style={{ color: '#CBD5E1', fontSize: 10 }}>•</span>
                      <span style={{ fontSize: 11, color: '#64748B', fontWeight: 550, display: 'inline-flex', alignItems: 'center', gap: 3 }}>
                        <Tag size={10} /> Owner: {item.creator_name || 'System'}
                      </span>
                    </div>
                  )}
                </div>

                {/* Right Section: Due Date, Priority, Controls */}
                <div style={{ display: 'flex', alignItems: 'center', gap: 16, flexShrink: 0 }}>
                  
                  {/* Due Status Label / Date */}
                  <div style={{ textAlign: 'right', minWidth: 140 }}>
                    <div style={{ 
                      fontSize: 12.5, 
                      fontWeight: 700, 
                      color: isCompleted ? '#10B981' : (item.dueStatus === 'OVERDUE' ? '#DC2626' : (item.dueStatus === 'TODAY' ? '#D97706' : '#334155')) 
                    }}>
                      {isCompleted ? 'Completed' : (item.dueStatus === 'OVERDUE' ? getOverdueDays(item.scheduled_at) : 'Due ' + formatDateTime(item.scheduled_at).split(',')[0])}
                    </div>
                    <div style={{ fontSize: 11, color: '#64748B', marginTop: 2 }}>
                      {formatDateTime(item.scheduled_at)}
                    </div>
                  </div>

                  {/* Priority Badge */}
                  <div style={{ minWidth: 65, textAlign: 'center' }}>
                    <span style={{ 
                      fontSize: 10.5, 
                      fontWeight: 800, 
                      padding: '3px 8px', 
                      borderRadius: 6, 
                      background: priorityBadgeColor.bg, 
                      color: priorityBadgeColor.text 
                    }}>
                      {item.priority}
                    </span>
                  </div>

                  {/* Completion / Reschedule Controls */}
                  <div style={{ display: 'flex', gap: 6, alignItems: 'center', width: 220, justifyContent: 'flex-end', position: 'relative' }}>
                    {!isCompleted ? (
                      <>
                        <button
                          className="outreach-btn"
                          style={{ padding: '6px 12px', fontSize: 12, background: '#10B981', color: '#FFFFFF', border: 'none', borderRadius: 6, fontWeight: 700, cursor: 'pointer' }}
                          onClick={(e) => handleComplete(e, item.id)}
                        >
                          ✓ Complete
                        </button>
                        
                        <button
                          className="activity-btn-outline"
                          style={{ padding: '5px 10px', fontSize: 12, borderRadius: 6 }}
                          onClick={(e) => {
                            e.stopPropagation();
                            setReschedulingId(reschedulingId === item.id ? null : item.id);
                          }}
                        >
                          Reschedule
                        </button>

                        <button
                          className="kebab-action-btn-circle"
                          style={{ width: 28, height: 28, borderRadius: '50%', background: '#F8FAFC', border: '1px solid #E2E8F0', cursor: 'pointer', display: 'flex', alignItems: 'center', justifyContent: 'center', color: '#DC2626' }}
                          onClick={(e) => handleCancel(e, item.id)}
                          title="Cancel Follow-up"
                        >
                          ✕
                        </button>
                      </>
                    ) : (
                      <span style={{ fontSize: 12, color: '#10B981', fontWeight: 700, display: 'inline-flex', alignItems: 'center', gap: 4 }}>
                        ✓ Task Done
                      </span>
                    )}

                    {/* Rescheduling Input Box Overlay */}
                    {reschedulingId === item.id && (
                      <div 
                        style={{ position: 'absolute', right: 0, top: '100%', marginTop: 8, padding: 12, background: '#FFFFFF', border: '1px solid #CBD5E1', borderRadius: 8, textAlign: 'left', zIndex: 100, boxShadow: '0 4px 12px rgba(0,0,0,0.1)', width: 200 }}
                        onClick={(e) => e.stopPropagation()}
                      >
                        <label style={{ fontSize: 11, fontWeight: 700, display: 'block', marginBottom: 4, color: '#334155' }}>Reschedule Task</label>
                        <input
                          type="datetime-local"
                          value={rescheduleDate}
                          onChange={(e) => setRescheduleDate(e.target.value)}
                          style={{ padding: '6px 8px', fontSize: 12, borderRadius: 6, border: '1px solid #CBD5E1', width: '100%', marginBottom: 8, boxSizing: 'border-box' }}
                        />
                        <div style={{ display: 'flex', gap: 6 }}>
                          <button className="outreach-btn outreach-btn-primary" style={{ padding: '4px 10px', fontSize: 11.5 }} onClick={(e) => handleRescheduleSubmit(e, item.id)}>Save</button>
                          <button className="activity-btn-outline" style={{ padding: '4px 10px', fontSize: 11.5 }} onClick={() => setReschedulingId(null)}>Cancel</button>
                        </div>
                      </div>
                    )}

                  </div>

                </div>

              </div>
            );
          })}
        </div>
      )}

      {/* ─── Compact Stats Footer ─── */}
      {!loading && followUps.length > 0 && filteredItems.length > 0 && (
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', fontSize: 12.5, color: '#64748B' }}>
          <div>
            Showing <strong>1–{filteredItems.length}</strong> of <strong>{followUps.length}</strong> tasks in queue
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

FollowUpQueue.propTypes = {
  onOpenProspect: PropTypes.func.isRequired,
  onNewFollowUpClick: PropTypes.func.isRequired,
};

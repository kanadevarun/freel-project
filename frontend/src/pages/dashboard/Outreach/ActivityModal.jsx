import { useState, useEffect, useRef } from 'react';
import PropTypes from 'prop-types';
import { listLeads } from '../../../services/leadsService';
import customerService from '../../../services/customerService';
import {
  createOutreachActivity,
  updateOutreachActivity,
} from '../../../services/outreachService';

export default function ActivityModal({
  activity,
  onClose,
  onSaveSuccess
}) {
  const isEdit = !!(activity && activity.id);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);

  // Form fields
  const [recordType, setRecordType] = useState('LEAD'); // 'LEAD' | 'CUSTOMER'
  const [selectedRecordId, setSelectedRecordId] = useState('');
  const [selectedRecordName, setSelectedRecordName] = useState('');
  
  const [activityType, setActivityType] = useState('CALL'); // EMAIL, CALL, FOLLOW_UP, MEETING, OTHER
  const [subject, setSubject] = useState('');
  const [description, setDescription] = useState('');
  const [status, setStatus] = useState('PENDING'); // PENDING, IN_PROGRESS, COMPLETED, OVERDUE
  const [priority, setPriority] = useState('MEDIUM'); // LOW, MEDIUM, HIGH
  
  const [scheduledDate, setScheduledDate] = useState('');
  const [scheduledTime, setScheduledTime] = useState('09:00');

  // Search selectors
  const [leads, setLeads] = useState([]);
  const [customers, setCustomers] = useState([]);
  const [searchQuery, setSearchQuery] = useState('');
  const [showDropdown, setShowDropdown] = useState(false);
  const dropdownRef = useRef(null);

  // Load leads and customers for selection on mount
  useEffect(() => {
    async function loadRecords() {
      try {
        const leadsData = await listLeads({ limit: 200 });
        setLeads(leadsData?.leads || []);

        const custsData = await customerService.getCustomers({ limit: 200 });
        setCustomers(custsData?.customers || []);
      } catch (err) {
        console.error('Failed to load related records:', err);
      }
    }
    loadRecords();
  }, []);

  // Populate data when editing an existing activity
  useEffect(() => {
    if (activity) {
      setActivityType(activity.activity_type || 'CALL');
      setSubject(activity.subject || '');
      setDescription(activity.description || '');
      setStatus(activity.status || 'PENDING');
      setPriority(activity.priority || 'MEDIUM');

      if (activity.lead_id) {
        setRecordType('LEAD');
        setSelectedRecordId(activity.lead_id);
        setSelectedRecordName(activity.lead_company_name || 'Selected Lead');
      } else if (activity.customer_id) {
        setRecordType('CUSTOMER');
        setSelectedRecordId(activity.customer_id);
        setSelectedRecordName(activity.lead_company_name || 'Selected Customer');
      }

      if (activity.scheduled_at) {
        const dt = new Date(activity.scheduled_at);
        const yyyy = dt.getFullYear();
        const mm = String(dt.getMonth() + 1).padStart(2, '0');
        const dd = String(dt.getDate()).padStart(2, '0');
        setScheduledDate(`${yyyy}-${mm}-${dd}`);
        
        const hh = String(dt.getHours()).padStart(2, '0');
        const min = String(dt.getMinutes()).padStart(2, '0');
        setScheduledTime(`${hh}:${min}`);
      }
    } else {
      // Set default scheduled date to today
      const today = new Date();
      const yyyy = today.getFullYear();
      const mm = String(today.getMonth() + 1).padStart(2, '0');
      const dd = String(today.getDate()).padStart(2, '0');
      setScheduledDate(`${yyyy}-${mm}-${dd}`);
    }
  }, [activity]);

  // Click outside search dropdown close handler
  useEffect(() => {
    function handleClickOutside(event) {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target)) {
        setShowDropdown(false);
      }
    }
    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  // Filter selections based on search query
  const filteredOptions = recordType === 'LEAD'
    ? leads.filter(l => (l.company_name || '').toLowerCase().includes(searchQuery.toLowerCase()))
    : customers.filter(c => (c.name || '').toLowerCase().includes(searchQuery.toLowerCase()));

  async function handleSubmit(e) {
    e.preventDefault();
    if (!subject.trim()) {
      setError('Please enter a subject/title.');
      return;
    }
    if (!selectedRecordId) {
      setError('Please select a related Lead or Customer record.');
      return;
    }

    setLoading(true);
    setError(null);

    // Combine date and time
    let scheduledAt = null;
    if (scheduledDate) {
      const [year, month, day] = scheduledDate.split('-');
      const [hour, min] = scheduledTime.split(':');
      const dt = new Date(year, month - 1, day, hour, min, 0);
      scheduledAt = dt.toISOString();
    }

    const payload = {
      activity_type: activityType,
      subject,
      description: description.trim() ? description : null,
      status,
      priority,
      scheduled_at: scheduledAt,
      lead_id: recordType === 'LEAD' ? Number(selectedRecordId) : null,
      customer_id: recordType === 'CUSTOMER' ? Number(selectedRecordId) : null,
    };

    try {
      if (isEdit) {
        await updateOutreachActivity(activity.id, payload);
      } else {
        await createOutreachActivity(payload);
      }
      onSaveSuccess();
    } catch (err) {
      setError(err.message || 'Failed to save outreach activity.');
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="outreach-modal-overlay">
      <div className="outreach-modal-container" style={{ maxWidth: 500 }}>
        <div className="outreach-modal-header">
          <h3>
            <span style={{ marginRight: 8, fontSize: 18 }}>📢</span>
            <span style={{ background: 'linear-gradient(135deg, #0F172A 0%, #334155 100%)', WebkitBackgroundClip: 'text', WebkitTextFillColor: 'transparent', fontWeight: 800 }}>
              {isEdit ? 'Edit Outreach Activity' : 'Log New Outreach'}
            </span>
          </h3>
          <button className="outreach-modal-close" onClick={onClose}>&times;</button>
        </div>

        <form onSubmit={handleSubmit} className="outreach-modal-form">
          {error && (
            <div className="error-alert" style={{ marginBottom: 16 }}>
              {error}
            </div>
          )}

          {/* 1. Record Type Selection (Lead vs Customer toggle) */}
          <div className="form-group-premium">
            <label className="input-label-premium">Related Record Type</label>
            <div className="modal-tab-switcher">
              <button
                type="button"
                className={`modal-tab-btn ${recordType === 'LEAD' ? 'active' : ''}`}
                disabled={isEdit}
                onClick={() => {
                  setRecordType('LEAD');
                  setSelectedRecordId('');
                  setSelectedRecordName('');
                  setSearchQuery('');
                }}
              >
                Lead
              </button>
              <button
                type="button"
                className={`modal-tab-btn ${recordType === 'CUSTOMER' ? 'active' : ''}`}
                disabled={isEdit}
                onClick={() => {
                  setRecordType('CUSTOMER');
                  setSelectedRecordId('');
                  setSelectedRecordName('');
                  setSearchQuery('');
                }}
              >
                Customer
              </button>
            </div>
          </div>

          {/* 2. Related Record Select Search */}
          <div className="form-group-premium" ref={dropdownRef} style={{ position: 'relative' }}>
            <label className="input-label-premium">Select Related {recordType === 'LEAD' ? 'Lead' : 'Customer'}</label>
            
            {isEdit ? (
              <input
                type="text"
                className="modal-input-premium"
                value={selectedRecordName}
                disabled
                style={{ background: '#F8FAFC' }}
              />
            ) : (
              <>
                <div 
                  className="modal-input-premium" 
                  onClick={() => setShowDropdown(true)}
                  style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', cursor: 'pointer', background: '#FFFFFF' }}
                >
                  <span style={{ color: selectedRecordName ? '#0F172A' : '#94A3B8', fontWeight: selectedRecordName ? 600 : 400 }}>
                    {selectedRecordName || `Search ${recordType === 'LEAD' ? 'leads' : 'customers'}...`}
                  </span>
                  <span style={{ fontSize: 10, color: '#64748B' }}>▼</span>
                </div>

                {showDropdown && (
                  <div className="floating-dropdown-menu" style={{ width: '100%', top: '100%', left: 0, padding: 8, maxHeight: 200, overflowY: 'auto' }}>
                    <input
                      type="text"
                      className="modal-input-premium"
                      placeholder="Type name to filter..."
                      value={searchQuery}
                      onChange={e => setSearchQuery(e.target.value)}
                      style={{ width: '100%', marginBottom: 8, fontSize: 12 }}
                      autoFocus
                    />
                    
                    {filteredOptions.length === 0 ? (
                      <div style={{ padding: '8px 12px', fontSize: 12, color: '#64748B', textAlign: 'center' }}>
                        No records found
                      </div>
                    ) : (
                      filteredOptions.map(item => {
                        const id = item.id;
                        const name = recordType === 'LEAD' ? item.company_name : item.name;

                        return (
                          <div
                            key={id}
                            className="dropdown-item"
                            style={{ padding: '8px 10px', borderRadius: 6, cursor: 'pointer', fontSize: 13 }}
                            onClick={() => {
                              setSelectedRecordId(id);
                              setSelectedRecordName(name);
                              setShowDropdown(false);
                            }}
                          >
                            🏢 {name}
                          </div>
                        );
                      })
                    )}
                  </div>
                )}
              </>
            )}
          </div>

          {/* 3. Activity Type, Status & Priority Row */}
          <div style={{ display: 'flex', gap: 12, width: '100%' }}>
            <div className="form-group-premium" style={{ flex: 1 }}>
              <label className="input-label-premium">Activity Type</label>
              <select
                className="modal-input-premium"
                value={activityType}
                onChange={e => setActivityType(e.target.value)}
              >
                <option value="CALL">📞 Call</option>
                <option value="EMAIL">✉️ Email</option>
                <option value="MEETING">👥 Meeting</option>
                <option value="FOLLOW_UP">🔔 Follow-up</option>
                <option value="OTHER">⚙️ Other</option>
              </select>
            </div>

            <div className="form-group-premium" style={{ flex: 1 }}>
              <label className="input-label-premium">Priority</label>
              <select
                className="modal-input-premium"
                value={priority}
                onChange={e => setPriority(e.target.value)}
              >
                <option value="LOW">Low</option>
                <option value="MEDIUM">Medium</option>
                <option value="HIGH">High</option>
              </select>
            </div>
          </div>

          <div style={{ display: 'flex', gap: 12, width: '100%' }}>
            <div className="form-group-premium" style={{ flex: 1 }}>
              <label className="input-label-premium">Status</label>
              <select
                className="modal-input-premium"
                value={status}
                onChange={e => setStatus(e.target.value)}
              >
                <option value="PENDING">Pending</option>
                <option value="IN_PROGRESS">In Progress</option>
                <option value="COMPLETED">Completed</option>
                <option value="OVERDUE">Overdue</option>
              </select>
            </div>

            <div className="form-group-premium" style={{ flex: 1 }}>
              <label className="input-label-premium">Scheduled Date</label>
              <input
                type="date"
                className="modal-input-premium"
                value={scheduledDate}
                onChange={e => setScheduledDate(e.target.value)}
                required
              />
            </div>
          </div>

          {/* Time Picker */}
          <div className="form-group-premium">
            <label className="input-label-premium">Scheduled Time</label>
            <input
              type="time"
              className="modal-input-premium"
              value={scheduledTime}
              onChange={e => setScheduledTime(e.target.value)}
              required
            />
          </div>

          {/* 4. Subject */}
          <div className="form-group-premium">
            <label className="input-label-premium">Subject / Title</label>
            <input
              type="text"
              className="modal-input-premium"
              placeholder="e.g. Discuss shipping options, Email presentation..."
              value={subject}
              onChange={e => setSubject(e.target.value)}
              required
            />
          </div>

          {/* 5. Description */}
          <div className="form-group-premium">
            <label className="input-label-premium">Description / Notes</label>
            <textarea
              className="modal-input-premium"
              placeholder="Provide meeting agendas, email draft snippets, or next-step logs..."
              value={description}
              onChange={e => setDescription(e.target.value)}
              rows={3}
              style={{ height: 'auto', padding: '10px 12px' }}
            />
          </div>

          {/* Actions */}
          <div className="outreach-modal-actions" style={{ marginTop: 24, display: 'flex', gap: 10, justifyContent: 'flex-end' }}>
            <button
              type="button"
              className="activity-btn-outline"
              onClick={onClose}
              disabled={loading}
            >
              Cancel
            </button>
            <button
              type="submit"
              className="btn-save-activity"
              disabled={loading}
            >
              {loading ? 'Saving...' : 'Save Activity'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}

ActivityModal.propTypes = {
  activity: PropTypes.object,
  onClose: PropTypes.func.isRequired,
  onSaveSuccess: PropTypes.func.isRequired,
};

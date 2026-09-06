import { useState, useEffect } from 'react';
import PropTypes from 'prop-types';
import { updateLead } from '../../../services/leadsService';
import { LEAD_STATUS } from './LeadsPage';
import LocationPickerMap from '../../../components/dashboard/LocationPickerMap';

/**
 * EditLeadModal — A popup form to edit an existing lead's fields.
 */
export default function EditLeadModal({ lead, onClose, onLeadUpdated, users = [] }) {
  const [form, setForm] = useState({
    company_name: lead.company_name || '',
    contact_name: lead.contact_name || '',
    email: lead.email || '',
    phone: lead.phone || '',
    source: lead.source || '',
    notes: lead.notes || '',
    location: lead.location || '',
    status: lead.status || 'NEW',
    assigned_to: lead.assigned_to || '',
    tags: lead.tags ? lead.tags.join(', ') : '',
  });
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  // Close modal on Escape key press
  useEffect(() => {
    const handleEscape = (e) => {
      if (e.key === 'Escape') onClose();
    };
    document.addEventListener('keydown', handleEscape);
    return () => document.removeEventListener('keydown', handleEscape);
  }, [onClose]);

  function handleChange(e) {
    setForm(prev => ({ ...prev, [e.target.name]: e.target.value }));
  }

  async function handleSubmit(e) {
    e.preventDefault();
    if (!form.company_name.trim()) {
      setError('Company name is required.');
      return;
    }

    setLoading(true);
    setError('');

    try {
      const payload = {
        ...form,
        assigned_to: form.assigned_to ? Number(form.assigned_to) : null,
        tags: form.tags ? form.tags.split(',').map(t => t.trim()).filter(Boolean) : [],
      };
      const updated = await updateLead(lead.id, payload);
      onLeadUpdated(updated); // pass the updated lead back to update drawer state
      onClose();
    } catch (err) {
      setError(err.message || 'Failed to update lead. Please try again.');
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="leads-modal-overlay" onClick={onClose} style={{ zIndex: 1100 }}>
      <div className="leads-modal wide-modal" onClick={e => e.stopPropagation()}>
        {/* Header */}
        <div className="modal-header">
          <div className="modal-title">Edit Lead</div>
          <button className="modal-close-btn" onClick={onClose}>✕</button>
        </div>

        {/* Error */}
        {error && (
          <div style={{ background: 'rgba(239,68,68,0.08)', border: '1px solid rgba(239,68,68,0.2)', borderRadius: 8, padding: '10px 14px', fontSize: 13, color: '#dc2626', marginBottom: 14 }}>
            {error}
          </div>
        )}

        {/* Form */}
        <form className="leads-form grid-form" onSubmit={handleSubmit}>
          <div className="form-group span-2">
            <label>Company Name *</label>
            <input
              type="text"
              name="company_name"
              value={form.company_name}
              onChange={handleChange}
              placeholder="e.g. Tech Logistics Inc."
              autoFocus
            />
          </div>

          <div className="form-group span-1">
            <label>Contact Name</label>
            <input
              type="text"
              name="contact_name"
              value={form.contact_name}
              onChange={handleChange}
              placeholder="e.g. John Doe"
            />
          </div>

          <div className="form-group span-1">
            <label>Email</label>
            <input
              type="email"
              name="email"
              value={form.email}
              onChange={handleChange}
              placeholder="john@company.com"
            />
          </div>

          <div className="form-group span-1">
            <label>Phone</label>
            <input
              type="text"
              name="phone"
              value={form.phone}
              onChange={handleChange}
              placeholder="e.g. +1 (555) 0199"
            />
          </div>

          <div className="form-group span-1">
            <label>Lead Source</label>
            <select name="source" value={form.source} onChange={handleChange}>
              <option value="">Select source...</option>
              <option value="Trade Show">Trade Show</option>
              <option value="Referral">Referral</option>
              <option value="LinkedIn">LinkedIn</option>
              <option value="Cold Outreach">Cold Outreach</option>
              <option value="Website">Website</option>
              <option value="ImportYeti">ImportYeti</option>
              <option value="Other">Other</option>
            </select>
          </div>

          <div className="form-group span-1">
            <label>Status</label>
            <select name="status" value={form.status} onChange={handleChange}>
              <option value={LEAD_STATUS.NEW}>New</option>
              <option value={LEAD_STATUS.QUALIFIED}>Qualified</option>
              <option value={LEAD_STATUS.IN_PROGRESS}>In Progress</option>
              <option value={LEAD_STATUS.CONVERTED}>Converted</option>
              <option value={LEAD_STATUS.REJECTED}>Rejected</option>
            </select>
          </div>

          <div className="form-group span-1">
            <label>Tags (comma-separated)</label>
            <input
              type="text"
              name="tags"
              value={form.tags}
              onChange={handleChange}
              placeholder="e.g. Urgent, High Value, Ocean Freight"
            />
          </div>

          <div className="form-group span-3">
            <label>Location</label>
            <LocationPickerMap
              value={form.location}
              onChange={(val) => setForm(prev => ({ ...prev, location: val }))}
              placeholder="Search society, city, port, or address..."
            />
          </div>

          <div className="form-group span-2 grid-notes">
            <label>Notes</label>
            <textarea
              name="notes"
              value={form.notes}
              onChange={handleChange}
              placeholder="Any quick notes about this lead..."
              rows={3}
            />
          </div>

          <div className="modal-actions span-3" style={{ marginTop: 8 }}>
            <button type="button" className="leads-btn leads-btn-ghost" onClick={onClose}>
              Cancel
            </button>
            <button type="submit" className="leads-btn leads-btn-primary" disabled={loading}>
              {loading ? 'Saving...' : 'Save Changes'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}

EditLeadModal.propTypes = {
  lead: PropTypes.object.isRequired,
  onClose: PropTypes.func.isRequired,
  onLeadUpdated: PropTypes.func.isRequired,
  users: PropTypes.array,
};

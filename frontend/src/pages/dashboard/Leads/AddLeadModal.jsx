import { useState, useEffect } from 'react';
import PropTypes from 'prop-types';
import { createLead } from '../../../services/leadsService';
import LocationPickerMap from '../../../components/dashboard/LocationPickerMap';

/**
 * AddLeadModal — A clean popup form to manually add a new lead.
 *
 * @param {{ onClose: () => void, onLeadAdded: () => void, users: Array }}
 */
export default function AddLeadModal({ onClose, onLeadAdded, users = [] }) {
  const [form, setForm] = useState({
    company_name: '',
    contact_name: '',
    email: '',
    phone: '',
    source: '',
    notes: '',
    location: '',
    assigned_to: '',
    tags: '',
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

  // Update a single field in the form state
  function handleChange(e) {
    setForm(prev => ({ ...prev, [e.target.name]: e.target.value }));
  }

  // Submit the form to the backend
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
      await createLead(payload);
      onLeadAdded(); // tell parent to refresh the table
      onClose();
    } catch (err) {
      setError(err.message || 'Failed to create lead. Please try again.');
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="leads-modal-overlay" onClick={onClose}>
      {/* Stop click from propagating to the overlay (which would close the modal) */}
      <div className="leads-modal wide-modal" onClick={e => e.stopPropagation()}>

        {/* Header */}
        <div className="modal-header">
          <div className="modal-title">Add New Lead</div>
          <button className="modal-close-btn" onClick={onClose}>✕</button>
        </div>

        {/* Error message */}
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

          <div className="form-group span-2">
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

          <div className="form-group span-3 grid-notes">
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
              {loading ? 'Adding...' : '+ Add Lead'}
            </button>
          </div>
        </form>

      </div>
    </div>
  );
}

AddLeadModal.propTypes = {
  onClose: PropTypes.func.isRequired,
  onLeadAdded: PropTypes.func.isRequired,
  users: PropTypes.array,
};

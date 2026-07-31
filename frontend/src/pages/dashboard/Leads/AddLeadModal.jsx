import { useState } from 'react';
import PropTypes from 'prop-types';
import { createLead } from '../../../services/leadsService';

/**
 * AddLeadModal — A clean popup form to manually add a new lead.
 *
 * Simple meaning: When the user clicks the "+ Add Lead" button, this form
 * slides in. They fill in the company name, contact, email, and source,
 * then hit "Add Lead". Our backend creates the record and immediately kicks off
 * the background AI worker to score the new lead.
 *
 * @param {{ onClose: () => void, onLeadAdded: () => void }}
 */
export default function AddLeadModal({ onClose, onLeadAdded }) {
  const [form, setForm] = useState({
    company_name: '',
    contact_name: '',
    email: '',
    source: '',
    notes: '',
  });
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

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
      await createLead(form);
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
      <div className="leads-modal" onClick={e => e.stopPropagation()}>

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
        <form className="leads-form" onSubmit={handleSubmit}>
          <div className="form-group">
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

          <div className="form-row">
            <div className="form-group">
              <label>Contact Name</label>
              <input
                type="text"
                name="contact_name"
                value={form.contact_name}
                onChange={handleChange}
                placeholder="e.g. John Doe"
              />
            </div>
            <div className="form-group">
              <label>Email</label>
              <input
                type="email"
                name="email"
                value={form.email}
                onChange={handleChange}
                placeholder="john@company.com"
              />
            </div>
          </div>

          <div className="form-group">
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

          <div className="form-group">
            <label>Notes</label>
            <textarea
              name="notes"
              value={form.notes}
              onChange={handleChange}
              placeholder="Any quick notes about this lead..."
            />
          </div>

          <div className="modal-actions">
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
};

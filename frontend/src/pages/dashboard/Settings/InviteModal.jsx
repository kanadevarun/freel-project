import { useState, useEffect } from 'react';
import api from '../../../services/api';

export default function InviteModal({ isOpen, onClose, onSuccess }) {
  const [email, setEmail] = useState('');
  const [roleId, setRoleId] = useState('');
  const [roles, setRoles] = useState([]);
  const [isLoadingRoles, setIsLoadingRoles] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState('');

  // Fetch roles when modal opens
  useEffect(() => {
    if (isOpen) {
      fetchRoles();
      // Reset form
      setEmail('');
      setError('');
    }
  }, [isOpen]);

  const DEFAULT_ROLES = [
    { id: 1, name: 'SUPER_ADMIN', description: 'Full system & organization administration' },
    { id: 2, name: 'ADMIN', description: 'Full organization workspace access' },
    { id: 3, name: 'SALES', description: 'Manage leads, pipeline & customer outreach' },
    { id: 4, name: 'PRICING', description: 'Manage carrier rates, RFQ quotes & margins' },
    { id: 5, name: 'OPERATIONS', description: 'Manage carrier bookings, container tracking & docs' },
  ];

  const fetchRoles = async () => {
    setIsLoadingRoles(true);
    try {
      const data = await api.get('/api/v1/roles');
      const loadedRoles = (Array.isArray(data) && data.length > 0) ? data : DEFAULT_ROLES;
      setRoles(loadedRoles);
      setRoleId(loadedRoles[0].id);
    } catch (err) {
      console.error("Failed to fetch roles:", err);
      setRoles(DEFAULT_ROLES);
      setRoleId(DEFAULT_ROLES[0].id);
    } finally {
      setIsLoadingRoles(false);
    }
  };

  if (!isOpen) return null;

  const handleSubmit = async (e) => {
    e.preventDefault();
    setIsSubmitting(true);
    setError('');

    try {
      await api.post('/api/v1/users/invite', {
        email,
        role_id: parseInt(roleId, 10),
        org_name: 'Your Company', // Hardcoded for MVP, ideally fetched from context
      });
      onSuccess();
    } catch (err) {
      setError(err.message || 'Failed to send invitation.');
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className="modal-overlay">
      <div className="modal-content">
        <div className="modal-header">
          <h2>Invite Team Member</h2>
          <button className="modal-close" onClick={onClose} aria-label="Close modal">
            &times;
          </button>
        </div>

        <form onSubmit={handleSubmit}>
          <div className="modal-body">
            {error && (
              <div className="form-alert error" style={{ marginBottom: '1.5rem' }}>
                {error}
              </div>
            )}

            <div className="form-group">
              <label htmlFor="invite-email" className="form-label">Email Address</label>
              <input
                id="invite-email"
                type="email"
                className="form-input"
                placeholder="colleague@company.com"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                required
                disabled={isSubmitting}
              />
            </div>

            <div className="form-group" style={{ marginTop: '1.5rem' }}>
              <label htmlFor="invite-role" className="form-label">Assign Role</label>
              {isLoadingRoles ? (
                <div style={{ padding: '0.75rem 1rem', fontSize: '0.95rem', color: 'var(--slate-500)' }}>
                  Loading roles...
                </div>
              ) : (
                <select
                  id="invite-role"
                  className="form-select"
                  value={roleId}
                  onChange={(e) => setRoleId(e.target.value)}
                  disabled={isSubmitting || roles.length === 0}
                >
                  {roles.map((role) => (
                    <option key={role.id} value={role.id}>
                      {role.name} — {role.description}
                    </option>
                  ))}
                </select>
              )}
            </div>
          </div>

          <div className="modal-footer">
            <button
              type="button"
              className="btn-secondary"
              onClick={onClose}
              disabled={isSubmitting}
            >
              Cancel
            </button>
            <button
              type="submit"
              className="btn-primary"
              disabled={isSubmitting || !email}
            >
              {isSubmitting ? 'Sending...' : 'Send Invitation'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}

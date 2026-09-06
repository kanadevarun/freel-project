import { useState, useEffect } from 'react';
import api from '../../../services/api';
import toast from 'react-hot-toast';
import './InviteModal.css';

export default function InviteModal({ isOpen, onClose, onSuccess }) {
  const [email, setEmail] = useState('');
  const [roleId, setRoleId] = useState('');
  const [roles, setRoles] = useState([]);
  const [isLoadingRoles, setIsLoadingRoles] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [isDropdownOpen, setIsDropdownOpen] = useState(false);
  const [emailError, setEmailError] = useState('');

  // Close on Escape key
  useEffect(() => {
    const handleEsc = (e) => {
      if (e.key === 'Escape' && isOpen && !isSubmitting) {
        onClose();
      }
    };
    window.addEventListener('keydown', handleEsc);
    return () => window.removeEventListener('keydown', handleEsc);
  }, [isOpen, isSubmitting, onClose]);

  // Fetch roles when modal opens
  useEffect(() => {
    if (isOpen) {
      fetchRoles();
      // Reset form
      setEmail('');
      setEmailError('');
    }
  }, [isOpen]);

  const fetchRoles = async () => {
    setIsLoadingRoles(true);
    try {
      const data = await api.get('/api/v1/roles');
      if (Array.isArray(data) && data.length > 0) {
        setRoles(data);
        setRoleId(''); // "Select a role" placeholder
      }
    } catch (err) {
      console.error("Failed to fetch roles:", err);
      // Fallback roles just in case
      const DEFAULT = [
        { id: 1, name: 'SUPER_ADMIN' },
        { id: 2, name: 'ADMIN' },
        { id: 3, name: 'SALES' },
        { id: 4, name: 'PRICING' },
        { id: 5, name: 'OPERATIONS' },
      ];
      setRoles(DEFAULT);
      setRoleId('');
    } finally {
      setIsLoadingRoles(false);
    }
  };

  if (!isOpen) return null;

  const handleSubmit = async (e) => {
    e.preventDefault();
    
    // Basic validation
    setEmailError('');
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    if (!emailRegex.test(email)) {
      toast.error("Please enter a valid email address.");
      return;
    }
    if (!roleId) {
      toast.error("Please select a role.");
      return;
    }

    setIsSubmitting(true);

    try {
      await api.post('/api/v1/users/invite', {
        email,
        role_id: parseInt(roleId, 10),
        org_name: 'Your Company',
      });
      toast.success('Invitation sent successfully.');
      onSuccess();
    } catch (err) {
      if (err.code === 'DUPLICATE_INVITATION') {
        setEmailError(err.message);
      } else {
        toast.error(err.message || 'Failed to send invitation.');
      }
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleOverlayClick = (e) => {
    if (isDropdownOpen) setIsDropdownOpen(false);
    if (e.target.classList.contains('invite-modal-overlay') && !isSubmitting && !email && !roleId) {
      onClose();
    }
  };

  return (
    <div className="invite-modal-overlay" onClick={handleOverlayClick}>
      <div className="invite-modal-container">
        <div className="invite-modal-header">
          <h2>Invite Team Member</h2>
          <p>
            Invite a new member to your organization.<br/>
            They will receive an email with instructions to join.
          </p>
          <button className="invite-modal-close" onClick={onClose} disabled={isSubmitting}>
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
              <line x1="18" y1="6" x2="6" y2="18"></line>
              <line x1="6" y1="6" x2="18" y2="18"></line>
            </svg>
          </button>
        </div>

        {isSubmitting && (
          <div className="invite-modal-loading-overlay">
            <div className="invite-spinner"></div>
            <p>Sending invitation...</p>
          </div>
        )}

        <form onSubmit={handleSubmit}>
          <div className={`invite-modal-body ${isSubmitting ? 'invite-modal-body-submitting' : ''}`}>
            
            <div className="invite-form-group">
              <label htmlFor="invite-email">Email address</label>
              <div className="invite-input-wrapper">
                <div className="invite-input-icon">
                  <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                    <path d="M4 4h16c1.1 0 2 .9 2 2v12c0 1.1-.9 2-2 2H4c-1.1 0-2-.9-2-2V6c0-1.1.9-2 2-2z"></path>
                    <polyline points="22,6 12,13 2,6"></polyline>
                  </svg>
                </div>
                <input
                  id="invite-email"
                  type="email"
                  className={`invite-input ${emailError ? 'invite-input-error' : ''}`}
                  placeholder="Enter email address"
                  value={email}
                  onChange={(e) => {
                    setEmail(e.target.value);
                    if (emailError) setEmailError('');
                  }}
                  disabled={isSubmitting}
                />
              </div>
              {emailError ? (
                <span className="invite-input-error-msg">
                  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                    <circle cx="12" cy="12" r="10"></circle>
                    <line x1="12" y1="8" x2="12" y2="12"></line>
                    <line x1="12" y1="16" x2="12.01" y2="16"></line>
                  </svg>
                  {emailError}
                </span>
              ) : (
                <span className="invite-input-help">We'll send an invitation link to this email address.</span>
              )}
            </div>

            <div className="invite-form-group">
              <label htmlFor="invite-role">Role</label>
              <div className="invite-custom-dropdown">
                <div 
                  className="invite-custom-dropdown-selected" 
                  onClick={() => !isSubmitting && !isLoadingRoles && setIsDropdownOpen(!isDropdownOpen)}
                >
                  {roleId ? roles.find(r => r.id == roleId)?.name : <span className="placeholder">Select a role</span>}
                  <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="#64748b" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                    <polyline points="6 9 12 15 18 9"></polyline>
                  </svg>
                </div>
                {isDropdownOpen && (
                  <ul className="invite-custom-dropdown-list">
                    {roles.map((role) => (
                      <li
                        key={role.id}
                        className={roleId == role.id ? 'selected' : ''}
                        onClick={(e) => {
                          e.stopPropagation();
                          setRoleId(role.id);
                          setIsDropdownOpen(false);
                        }}
                      >
                        {role.name}
                      </li>
                    ))}
                  </ul>
                )}
              </div>
              <span className="invite-input-help">Choose the appropriate role for this team member.</span>
            </div>

            <div className="invite-info-box">
              <div className="invite-info-icon">
                <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5">
                  <circle cx="12" cy="12" r="10"></circle>
                  <line x1="12" y1="16" x2="12" y2="12"></line>
                  <line x1="12" y1="8" x2="12.01" y2="8"></line>
                </svg>
              </div>
              <p>The invited user will be added to your organization and can be assigned a role.</p>
            </div>

          </div>

          <div className="invite-modal-footer">
            <button
              type="button"
              className="invite-btn-cancel"
              onClick={onClose}
              disabled={isSubmitting}
            >
              Cancel
            </button>
            <button
              type="submit"
              className="invite-btn-submit"
              disabled={isSubmitting || !email || !roleId}
            >
              {isSubmitting ? 'Sending...' : 'Send Invitation'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}

import { useState, useEffect } from 'react';
import api from '../../../services/api';
import InviteModal from './InviteModal';
import RoleBadge from '../../../components/dashboard/RoleBadge';
import './UsersPage.css';

export default function UsersPage() {
  const [activeTab, setActiveTab] = useState('members');
  const [members, setMembers] = useState([]);
  const [invites, setInvites] = useState([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState(null);
  
  const [isInviteModalOpen, setIsInviteModalOpen] = useState(false);

  // Fetch data on mount
  useEffect(() => {
    fetchData();
  }, []);

  const fetchData = async () => {
    setIsLoading(true);
    setError(null);
    try {
      // Run both API calls in parallel
      const [membersRes, invitesRes] = await Promise.all([
        api.get('/api/v1/users'),
        api.get('/api/v1/users/invites')
      ]);
      setMembers(membersRes || []);
      setInvites(invitesRes || []);
    } catch (err) {
      setError('Failed to load users and invitations. Please try again.');
      console.error(err);
    } finally {
      setIsLoading(false);
    }
  };

  const handleRemoveMember = async (userId) => {
    if (!window.confirm('Are you sure you want to remove this member from your organization?')) return;
    
    try {
      await api.delete(`/api/v1/users/${userId}`);
      setMembers((prev) => prev.filter((m) => m.user_id !== userId));
    } catch (err) {
      alert(err.message || 'Failed to remove member.');
    }
  };

  const handleCancelInvite = async (inviteId) => {
    if (!window.confirm('Are you sure you want to cancel this invitation?')) return;
    
    try {
      await api.delete(`/api/v1/users/invites/${inviteId}`);
      setInvites((prev) => prev.filter((i) => i.id !== inviteId));
    } catch (err) {
      alert(err.message || 'Failed to cancel invitation.');
    }
  };

  const onInviteSuccess = () => {
    setIsInviteModalOpen(false);
    // Refresh the invites list
    fetchData();
  };

  // Helper to get initials for avatar
  const getInitials = (firstName, lastName, email) => {
    if (firstName && lastName) return `${firstName[0]}${lastName[0]}`.toUpperCase();
    if (firstName) return firstName.substring(0, 2).toUpperCase();
    return email.substring(0, 2).toUpperCase();
  };

  return (
    <div className="users-page">
      <div className="users-header">
        <div>
          <h1>User Management</h1>
          <p>Manage your team members, roles, and pending invitations.</p>
        </div>
        <button className="btn-primary" onClick={() => setIsInviteModalOpen(true)}>
          + Invite User
        </button>
      </div>

      <div className="users-tabs">
        <button 
          className={`users-tab ${activeTab === 'members' ? 'active' : ''}`}
          onClick={() => setActiveTab('members')}
        >
          Active Members ({members.length})
        </button>
        <button 
          className={`users-tab ${activeTab === 'invites' ? 'active' : ''}`}
          onClick={() => setActiveTab('invites')}
        >
          Pending Invites ({invites.length})
        </button>
      </div>

      {error && (
        <div className="form-alert error" style={{ marginBottom: '2rem' }}>
          {error}
        </div>
      )}

      {isLoading ? (
        <div className="table-card" style={{ padding: '4rem', textAlign: 'center' }}>
          <div className="auth-spinner" style={{ margin: '0 auto' }}></div>
        </div>
      ) : activeTab === 'members' ? (
        <div className="table-card">
          {members.length === 0 ? (
            <div className="empty-state">
              <div className="empty-state-icon">👥</div>
              <h3>No Active Members</h3>
              <p>You are the only member in this organization.</p>
            </div>
          ) : (
            <table className="users-table">
              <thead>
                <tr>
                  <th>User</th>
                  <th>Role</th>
                  <th>Status</th>
                  <th>Joined</th>
                  <th></th>
                </tr>
              </thead>
              <tbody>
                {members.map((member) => (
                  <tr key={member.user_id}>
                    <td>
                      <div className="user-info-cell">
                        <div className="user-avatar">
                          {getInitials(member.first_name, member.last_name, member.email)}
                        </div>
                        <div>
                          <span className="user-name">
                            {member.first_name ? `${member.first_name} ${member.last_name || ''}` : 'No Name'}
                          </span>
                          <span className="user-email">{member.email}</span>
                        </div>
                      </div>
                    </td>
                    <td>
                      <RoleBadge role={member.role_name} />
                    </td>
                    <td>
                      <span style={{ 
                        display: 'inline-flex', 
                        alignItems: 'center', 
                        gap: '6px', 
                        fontSize: '0.85rem',
                        color: 'var(--slate-600)' 
                      }}>
                        <span style={{
                          width: '8px',
                          height: '8px',
                          borderRadius: '50%',
                          backgroundColor: member.status === 'ACTIVE' ? '#10b981' : '#f59e0b'
                        }}></span>
                        {member.status}
                      </span>
                    </td>
                    <td style={{ color: 'var(--slate-500)', fontSize: '0.85rem' }}>
                      {new Date(member.joined_at).toLocaleDateString()}
                    </td>
                    <td>
                      <div className="actions-cell">
                        <button 
                          className="btn-icon danger" 
                          onClick={() => handleRemoveMember(member.user_id)}
                          title="Remove user"
                        >
                          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                            <path strokeLinecap="round" strokeLinejoin="round" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                          </svg>
                        </button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </div>
      ) : (
        <div className="table-card">
          {invites.length === 0 ? (
            <div className="empty-state">
              <div className="empty-state-icon">✉️</div>
              <h3>No Pending Invites</h3>
              <p>Any invitations you send will appear here until they are accepted.</p>
              <button className="btn-secondary" onClick={() => setIsInviteModalOpen(true)} style={{ marginTop: '1rem' }}>
                Invite Someone Now
              </button>
            </div>
          ) : (
            <table className="users-table">
              <thead>
                <tr>
                  <th>Invited Email</th>
                  <th>Assigned Role</th>
                  <th>Sent Date</th>
                  <th>Expires On</th>
                  <th></th>
                </tr>
              </thead>
              <tbody>
                {invites.map((invite) => (
                  <tr key={invite.id}>
                    <td>
                      <div className="user-info-cell">
                        <div className="user-avatar" style={{ background: 'var(--slate-100)', color: 'var(--slate-500)' }}>
                          ✉️
                        </div>
                        <div>
                          <span className="user-name">{invite.email}</span>
                          <span className="user-email">Pending Acceptance</span>
                        </div>
                      </div>
                    </td>
                    <td>
                      <RoleBadge role={invite.role_name} />
                    </td>
                    <td style={{ color: 'var(--slate-500)', fontSize: '0.85rem' }}>
                      {new Date(invite.created_at).toLocaleDateString()}
                    </td>
                    <td style={{ color: 'var(--slate-500)', fontSize: '0.85rem' }}>
                      {new Date(invite.expires_at).toLocaleDateString()}
                    </td>
                    <td>
                      <div className="actions-cell">
                        <button 
                          className="btn-icon danger" 
                          onClick={() => handleCancelInvite(invite.id)}
                          title="Cancel invitation"
                        >
                          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                            <path strokeLinecap="round" strokeLinejoin="round" d="M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z" />
                          </svg>
                        </button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </div>
      )}

      <InviteModal 
        isOpen={isInviteModalOpen} 
        onClose={() => setIsInviteModalOpen(false)} 
        onSuccess={onInviteSuccess} 
      />
    </div>
  );
}

import { useState, useEffect } from 'react';
import api from '../../../services/api';
import InviteModal from './InviteModal';
import ConfirmModal from './ConfirmModal';
import toast from 'react-hot-toast';
import { useAuth } from '../../../context/AuthContext';
import './UsersPage.css';

export default function UsersPage() {
  const { user } = useAuth();
  const [activeTab, setActiveTab] = useState('members');
  const [members, setMembers] = useState([]);
  const [invites, setInvites] = useState([]);
  const [roles, setRoles] = useState([]);
  const [isLoading, setIsLoading] = useState(true);
  
  const [isInviteModalOpen, setIsInviteModalOpen] = useState(false);
  const [activeMenuId, setActiveMenuId] = useState(null);
  
  // Confirmation Modal State
  const [confirmState, setConfirmState] = useState({
    isOpen: false,
    title: '',
    message: '',
    actionText: '',
    actionType: 'danger',
    onConfirm: null,
    isLoading: false,
  });

  // Pagination states
  const [memberPage, setMemberPage] = useState(1);
  const [invitePage, setInvitePage] = useState(1);
  const ITEMS_PER_PAGE = 10;

  useEffect(() => {
    fetchData();
  }, []);

  const fetchData = async () => {
    setIsLoading(true);
    try {
      const [membersRes, invitesRes, rolesRes] = await Promise.all([
        api.get('/api/v1/users'),
        api.get('/api/v1/users/invites'),
        api.get('/api/v1/roles')
      ]);
      setMembers(membersRes || []);
      setInvites(invitesRes || []);
      setRoles(rolesRes || []);
      setMemberPage(1);
      setInvitePage(1);
    } catch (err) {
      toast.error('Failed to load users and invitations.');
    } finally {
      setIsLoading(false);
    }
  };

  const handleRemoveMemberClick = (member) => {
    setActiveMenuId(null);
    setConfirmState({
      isOpen: true,
      title: 'Remove User',
      message: `Are you sure you want to remove ${member.first_name || member.email} from your organization? They will immediately lose access to all workspaces and data.`,
      actionText: 'Remove User',
      actionType: 'danger',
      isLoading: false,
      onConfirm: () => handleRemoveMember(member.user_id),
    });
  };

  const handleRemoveMember = async (userId) => {
    setConfirmState(prev => ({ ...prev, isLoading: true }));
    try {
      await api.delete(`/api/v1/users/${userId}`);
      setMembers((prev) => prev.filter((m) => m.user_id !== userId));
      toast.success('User removed successfully.');
      setConfirmState(prev => ({ ...prev, isOpen: false }));
    } catch (err) {
      toast.error(err.message || 'Failed to remove member.');
      setConfirmState(prev => ({ ...prev, isLoading: false }));
    }
  };

  const handleRoleChange = async (userId, newRoleIdStr) => {
    const newRoleId = parseInt(newRoleIdStr, 10);
    try {
      await api.patch(`/api/v1/users/${userId}/role`, { role_id: newRoleId });
      setMembers(prev => prev.map(m => {
        if (m.user_id === userId) {
          const newRole = roles.find(r => r.id === newRoleId);
          return { ...m, role_id: newRoleId, role_name: newRole ? newRole.name : m.role_name };
        }
        return m;
      }));
      toast.success('Role updated successfully');
    } catch (err) {
      toast.error('Unable to change role\nThe last SUPER_ADMIN cannot be changed. Assign another user as SUPER_ADMIN first.');
      setMembers(prev => [...prev]);
    }
  };

  const handleCancelInviteClick = (invite) => {
    setActiveMenuId(null);
    setConfirmState({
      isOpen: true,
      title: 'Cancel Invitation',
      message: `Are you sure you want to cancel the invitation for ${invite.email}? The invitation link will immediately expire.`,
      actionText: 'Cancel Invitation',
      actionType: 'danger',
      isLoading: false,
      onConfirm: () => handleCancelInvite(invite.id),
    });
  };

  const handleCancelInvite = async (inviteId) => {
    setConfirmState(prev => ({ ...prev, isLoading: true }));
    try {
      await api.delete(`/api/v1/users/invites/${inviteId}`);
      setInvites((prev) => prev.filter((i) => i.id !== inviteId));
      toast.success('Invitation canceled.');
      setConfirmState(prev => ({ ...prev, isOpen: false }));
    } catch (err) {
      toast.error(err.message || 'Failed to cancel invitation.');
      setConfirmState(prev => ({ ...prev, isLoading: false }));
    }
  };

  const onInviteSuccess = () => {
    setIsInviteModalOpen(false);
    fetchData();
  };

  const getInitials = (firstName, lastName, email) => {
    if (firstName && lastName) return `${firstName[0]}${lastName[0]}`.toUpperCase();
    if (firstName) return firstName.substring(0, 2).toUpperCase();
    return email.substring(0, 2).toUpperCase();
  };

  const superAdminCount = members.filter(m => m.role_name === 'SUPER_ADMIN' && m.status === 'ACTIVE').length;
  const regularUserCount = members.filter(m => m.role_name !== 'SUPER_ADMIN' && m.status === 'ACTIVE').length;

  // Pagination logic
  const totalMemberPages = Math.ceil(members.length / ITEMS_PER_PAGE);
  const paginatedMembers = members.slice((memberPage - 1) * ITEMS_PER_PAGE, memberPage * ITEMS_PER_PAGE);

  const totalInvitePages = Math.ceil(invites.length / ITEMS_PER_PAGE);
  const paginatedInvites = invites.slice((invitePage - 1) * ITEMS_PER_PAGE, invitePage * ITEMS_PER_PAGE);

  return (
    <div className="users-page" onClick={() => setActiveMenuId(null)}>
      <div className="users-header">
        <div className="users-header-content">
          <h1>Users & Team</h1>
          <p>Manage your organization members, assign roles, and control access.</p>
        </div>
        <div className="header-actions">
          <button className="btn-secondary" onClick={fetchData}>
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" style={{marginRight: '6px'}}>
              <path d="M21 2v6h-6"></path>
              <path d="M3 12a9 9 0 0 1 15-6.7L21 8"></path>
              <path d="M3 22v-6h6"></path>
              <path d="M21 12a9 9 0 0 1-15 6.7L3 16"></path>
            </svg>
            Refresh
          </button>
          <button className="btn-primary" onClick={() => setIsInviteModalOpen(true)}>
            + Invite User
          </button>
        </div>
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
          Pending Invitations ({invites.length})
        </button>
      </div>

      {/* Filter Toolbar separated from the Table Card */}
      <div className="users-filters">
        <div style={{ position: 'relative', flex: 1, minWidth: 0 }}>
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="var(--slate-400)" strokeWidth="2" style={{position: 'absolute', left: '12px', top: '13px'}}>
            <circle cx="11" cy="11" r="8"></circle>
            <line x1="21" y1="21" x2="16.65" y2="16.65"></line>
          </svg>
          <input type="text" className="users-search-input" placeholder="Search by name or email..." style={{ paddingLeft: '32px', width: '100%' }} />
        </div>
        <select className="users-filter-select">
          <option>All Roles</option>
          {roles.map(r => <option key={r.id}>{r.name}</option>)}
        </select>
        <select className="users-filter-select">
          <option>All Status</option>
          <option>Active</option>
        </select>
        <button className="btn-secondary" style={{ width: '100px' }}>
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" style={{marginRight: '6px'}}>
            <polygon points="22 3 2 3 10 12.46 10 19 14 21 14 12.46 22 3"></polygon>
          </svg>
          Filters
        </button>
      </div>

      <div className="table-card">
        {isLoading ? (
          <div style={{ padding: '4rem', textAlign: 'center' }}>
            <div className="auth-spinner" style={{ margin: '0 auto' }}></div>
          </div>
        ) : activeTab === 'members' ? (
          <>
            {members.length === 0 ? (
              <div className="empty-state">
                <div className="empty-state-icon">👥</div>
                <h3>No Active Members</h3>
                <p>You are the only member in this organization.</p>
              </div>
            ) : (
              <table className="users-table members-table">
                <thead>
                  <tr>
                    <th>User</th>
                    <th>Email</th>
                    <th>Role</th>
                    <th>Status</th>
                    <th>Joined</th>
                    <th>Actions</th>
                  </tr>
                </thead>
                <tbody>
                  {paginatedMembers.map((member) => {
                    const isCurrentUser = user && user.email === member.email;
                    const isProtectedSuperAdmin = member.role_name === 'SUPER_ADMIN' && superAdminCount <= 1;
                    const fullName = member.first_name ? `${member.first_name} ${member.last_name || ''}` : 'No Name';

                    return (
                      <tr key={member.user_id}>
                        <td>
                          <div className="user-info-cell">
                            <div className="user-avatar">
                              {getInitials(member.first_name, member.last_name, member.email)}
                            </div>
                            <div className="user-name-wrapper">
                              <span className="user-name" title={fullName}>
                                {fullName}
                              </span>
                              {isCurrentUser && <span className="you-badge">You</span>}
                            </div>
                          </div>
                        </td>
                        <td>
                          <span className="user-email" title={member.email}>{member.email}</span>
                        </td>
                        <td>
                          <div className="role-dropdown-container">
                            <select 
                              value={member.role_id}
                              onChange={(e) => handleRoleChange(member.user_id, e.target.value)}
                              disabled={isProtectedSuperAdmin}
                            >
                              {roles.map(r => (
                                <option key={r.id} value={r.id}>
                                  {r.name}
                                </option>
                              ))}
                            </select>
                            {isProtectedSuperAdmin && (
                              <>
                                <div className="role-lock-icon">
                                  <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5">
                                    <rect x="3" y="11" width="18" height="11" rx="2" ry="2"></rect>
                                    <path d="M7 11V7a5 5 0 0 1 10 0v4"></path>
                                  </svg>
                                </div>
                                <div className="custom-tooltip">
                                  <strong>Role protected</strong>
                                  This role cannot be changed because your organization must always have at least one SUPER_ADMIN.
                                </div>
                              </>
                            )}
                          </div>
                        </td>
                        <td>
                          <div className="status-badge">
                            <div className={`status-dot ${member.status === 'ACTIVE' ? 'active' : 'inactive'}`}></div>
                            {member.status === 'ACTIVE' ? 'Active' : member.status}
                          </div>
                        </td>
                        <td>
                          <span className="joined-date">
                            {new Date(member.joined_at).toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' })}
                          </span>
                        </td>
                        <td>
                          <div className="actions-cell">
                            <button 
                              className="btn-icon menu" 
                              onClick={(e) => {
                                e.stopPropagation();
                                setActiveMenuId(activeMenuId === member.user_id ? null : member.user_id);
                              }}
                            >
                              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5">
                                <circle cx="12" cy="12" r="1.5"></circle>
                                <circle cx="12" cy="5" r="1.5"></circle>
                                <circle cx="12" cy="19" r="1.5"></circle>
                              </svg>
                            </button>
                            
                            {activeMenuId === member.user_id && (
                              <div style={{
                                position: 'absolute',
                                right: '0',
                                top: '100%',
                                background: 'white',
                                border: '1px solid var(--slate-200)',
                                boxShadow: '0 4px 6px -1px rgba(0, 0, 0, 0.1)',
                                borderRadius: '8px',
                                padding: '0.5rem 0',
                                minWidth: '150px',
                                zIndex: 10
                              }}>
                                <button 
                                  onClick={(e) => { e.stopPropagation(); handleRemoveMemberClick(member); }}
                                  style={{
                                    display: 'flex',
                                    alignItems: 'center',
                                    gap: '8px',
                                    width: '100%',
                                    padding: '0.5rem 1rem',
                                    background: 'none',
                                    border: 'none',
                                    color: '#ef4444',
                                    fontSize: '0.85rem',
                                    cursor: 'pointer',
                                    textAlign: 'left'
                                  }}
                                >
                                  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                                    <path strokeLinecap="round" strokeLinejoin="round" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                                  </svg>
                                  Remove User
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
            )}
            <div className="table-footer">
              <span>
                Showing {members.length === 0 ? 0 : (memberPage - 1) * ITEMS_PER_PAGE + 1} to {Math.min(memberPage * ITEMS_PER_PAGE, members.length)} of {members.length} results
              </span>
              {totalMemberPages > 1 && (
                <div className="pagination-controls">
                  <button 
                    className="btn-page"
                    disabled={memberPage === 1}
                    onClick={() => setMemberPage(p => p - 1)}
                  >
                    Previous
                  </button>
                  <button 
                    className="btn-page"
                    disabled={memberPage === totalMemberPages}
                    onClick={() => setMemberPage(p => p + 1)}
                  >
                    Next
                  </button>
                </div>
              )}
            </div>
          </>
        ) : (
          <>
            {invites.length === 0 ? (
              <div className="empty-state">
                <div className="empty-state-icon">✉️</div>
                <h3>No Pending Invites</h3>
                <p>Any invitations you send will appear here until they are accepted.</p>
                <button className="btn-secondary" onClick={() => setIsInviteModalOpen(true)} style={{ margin: '1rem auto 0' }}>
                  Invite Someone Now
                </button>
              </div>
            ) : (
              <table className="users-table invites-table">
                <thead>
                  <tr>
                    <th>Invited Email</th>
                    <th>Assigned Role</th>
                    <th>Sent Date</th>
                    <th>Expires On</th>
                    <th>Actions</th>
                  </tr>
                </thead>
                <tbody>
                  {paginatedInvites.map((invite) => (
                    <tr key={invite.id}>
                      <td>
                        <div className="user-info-cell">
                          <div className="user-avatar" style={{ background: 'var(--slate-100)', color: 'var(--slate-500)' }}>
                            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                              <path d="M4 4h16c1.1 0 2 .9 2 2v12c0 1.1-.9 2-2 2H4c-1.1 0-2-.9-2-2V6c0-1.1.9-2 2-2z"></path>
                              <polyline points="22,6 12,13 2,6"></polyline>
                            </svg>
                          </div>
                          <span className="user-name" title={invite.email}>{invite.email}</span>
                        </div>
                      </td>
                      <td>
                        <span className="invite-role-badge">
                          <span className="invite-role-dot"></span>
                          {invite.role_name}
                        </span>
                      </td>
                      <td>
                        <span className="joined-date">
                          {new Date(invite.created_at).toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' })}
                        </span>
                      </td>
                      <td>
                        <span className="joined-date">
                          {new Date(invite.expires_at).toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' })}
                        </span>
                      </td>
                      <td>
                        <div className="actions-cell">
                           <button 
                              className="btn-icon menu" 
                              onClick={(e) => {
                                e.stopPropagation();
                                setActiveMenuId(activeMenuId === `invite-${invite.id}` ? null : `invite-${invite.id}`);
                              }}
                            >
                              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5">
                                <circle cx="12" cy="12" r="1.5"></circle>
                                <circle cx="12" cy="5" r="1.5"></circle>
                                <circle cx="12" cy="19" r="1.5"></circle>
                              </svg>
                            </button>
                            
                            {activeMenuId === `invite-${invite.id}` && (
                              <div style={{
                                position: 'absolute',
                                right: '0',
                                top: '100%',
                                background: 'white',
                                border: '1px solid var(--slate-200)',
                                boxShadow: '0 4px 6px -1px rgba(0, 0, 0, 0.1)',
                                borderRadius: '8px',
                                padding: '0.5rem 0',
                                minWidth: '160px',
                                zIndex: 10
                              }}>
                                <button 
                                  onClick={(e) => { e.stopPropagation(); handleCancelInviteClick(invite); }}
                                  style={{
                                    display: 'flex',
                                    alignItems: 'center',
                                    gap: '8px',
                                    width: '100%',
                                    padding: '0.5rem 1rem',
                                    background: 'none',
                                    border: 'none',
                                    color: '#ef4444',
                                    fontSize: '0.85rem',
                                    cursor: 'pointer',
                                    textAlign: 'left'
                                  }}
                                >
                                  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                                    <path strokeLinecap="round" strokeLinejoin="round" d="M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z" />
                                  </svg>
                                  Cancel Invitation
                                </button>
                              </div>
                            )}
                        </div>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            )}
            <div className="table-footer">
              <span>
                Showing {invites.length === 0 ? 0 : (invitePage - 1) * ITEMS_PER_PAGE + 1} to {Math.min(invitePage * ITEMS_PER_PAGE, invites.length)} of {invites.length} results
              </span>
              {totalInvitePages > 1 && (
                <div className="pagination-controls">
                  <button 
                    className="btn-page"
                    disabled={invitePage === 1}
                    onClick={() => setInvitePage(p => p - 1)}
                  >
                    Previous
                  </button>
                  <button 
                    className="btn-page"
                    disabled={invitePage === totalInvitePages}
                    onClick={() => setInvitePage(p => p + 1)}
                  >
                    Next
                  </button>
                </div>
              )}
            </div>
          </>
        )}
      </div>

      {/* Summary Row */}
      <div className="summary-section">
        <div className="summary-card">
          <div className="summary-icon blue">
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
              <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"></path>
              <circle cx="9" cy="7" r="4"></circle>
              <path d="M23 21v-2a4 4 0 0 0-3-3.87"></path>
              <path d="M16 3.13a4 4 0 0 1 0 7.75"></path>
            </svg>
          </div>
          <div className="summary-content">
            <h4>Total Active Members</h4>
            <div className="summary-value">{members.length}</div>
            <p>Users in your organization</p>
          </div>
        </div>
        
        <div className="summary-card">
          <div className="summary-icon green">
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
              <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"></path>
            </svg>
          </div>
          <div className="summary-content">
            <h4>Super Admins</h4>
            <div className="summary-value">{superAdminCount}</div>
            <p>Full system access</p>
          </div>
        </div>

        <div className="summary-card">
          <div className="summary-icon purple">
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
              <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"></path>
              <circle cx="12" cy="7" r="4"></circle>
            </svg>
          </div>
          <div className="summary-content">
            <h4>Regular Users</h4>
            <div className="summary-value">{regularUserCount}</div>
            <p>Users with specific roles</p>
          </div>
        </div>

        <div className="summary-card">
          <div className="summary-icon orange">
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
              <path d="M4 4h16c1.1 0 2 .9 2 2v12c0 1.1-.9 2-2 2H4c-1.1 0-2-.9-2-2V6c0-1.1.9-2 2-2z"></path>
              <polyline points="22,6 12,13 2,6"></polyline>
            </svg>
          </div>
          <div className="summary-content">
            <h4>Pending Invitations</h4>
            <div className="summary-value">{invites.length}</div>
            <p>Awaiting acceptance</p>
          </div>
        </div>
      </div>

      <InviteModal 
        isOpen={isInviteModalOpen} 
        onClose={() => setIsInviteModalOpen(false)} 
        onSuccess={onInviteSuccess} 
      />

      <ConfirmModal
        isOpen={confirmState.isOpen}
        title={confirmState.title}
        message={confirmState.message}
        confirmText={confirmState.actionText}
        confirmStyle={confirmState.actionType}
        isLoading={confirmState.isLoading}
        onConfirm={confirmState.onConfirm}
        onCancel={() => setConfirmState(prev => ({ ...prev, isOpen: false }))}
      />
    </div>
  );
}

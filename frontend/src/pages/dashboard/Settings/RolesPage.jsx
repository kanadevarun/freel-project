import { useState, useEffect, useCallback } from 'react';
import { createPortal } from 'react-dom';
import { 
  CheckCircle2, 
  X, 
  Minus, 
  MoreVertical, 
  Shield, 
  Users, 
  Lock, 
  ShieldCheck,
  Search,
  Plus,
  Tag,
  Truck,
  Wallet,
  FileText
} from 'lucide-react';
import { useNavigate } from 'react-router-dom';
import api from '../../../services/api';
import './RolesPage.css';

// The 10 canonical resources shown in the permissions matrix
const RESOURCES = [
  { id: 'COMPANIES',     name: 'Companies',     desc: 'Company directory & management' },
  { id: 'LEADS',         name: 'Leads',         desc: 'Lead management' },
  { id: 'OPPORTUNITIES', name: 'Opportunities', desc: 'Sales opportunities' },
  { id: 'RFQS',          name: 'RFQs',          desc: 'Request for quotations' },
  { id: 'OUTREACH',      name: 'Outreach',      desc: 'Email campaigns' },
  { id: 'SHIPMENTS',     name: 'Shipments',     desc: 'Shipment operations' },
  { id: 'DOCUMENTS',     name: 'Documents',     desc: 'Document management' },
  { id: 'FINANCE',       name: 'Finance',       desc: 'Invoices & payments' },
  { id: 'USERS',         name: 'Users & Team',  desc: 'User management' },
  { id: 'SETTINGS',      name: 'Settings',      desc: 'System settings' },
];

// Only 4 actions — no WRITE, no EXPORT
const ACTIONS = ['CREATE', 'READ', 'UPDATE', 'DELETE'];

// Visual styling per role
const ROLE_VISUALS = {
  'SUPER_ADMIN':   { icon: Shield,   color: 'blue',   isSystem: true },
  'SALES':         { icon: Users,    color: 'blue',   isSystem: true },
  'PRICING':       { icon: Tag,      color: 'purple', isSystem: true },
  'OPERATIONS':    { icon: Truck,    color: 'orange', isSystem: true },
  'FINANCE':       { icon: Wallet,   color: 'teal',   isSystem: true },
  'DOCUMENTATION': { icon: FileText, color: 'red',    isSystem: true },
  'HR':            { icon: Users,    color: 'teal',   isSystem: true },
};
const DEFAULT_VISUAL = { icon: Shield, color: 'gray' };

export default function RolesPage() {
  const navigate = useNavigate();
  const [roles, setRoles] = useState([]);
  const [selectedRoleId, setSelectedRoleId] = useState(null);
  const [permissions, setPermissions] = useState([]); // Array of { resource, action }
  const [selectedRoleName, setSelectedRoleName] = useState('');
  const [selectedRoleDesc, setSelectedRoleDesc] = useState('');
  const [stats, setStats] = useState(null);

  const [isLoadingRoles, setIsLoadingRoles] = useState(true);
  const [isLoadingPerms, setIsLoadingPerms] = useState(false);
  const [searchQuery, setSearchQuery] = useState('');
  
  // Edit Permissions state
  const [isEditing, setIsEditing] = useState(false);
  const [editedPerms, setEditedPerms] = useState({}); // { 'RESOURCE.ACTION': true/false }
  const [isSaving, setIsSaving] = useState(false);

  // Create Role state
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [newRoleName, setNewRoleName] = useState('');
  const [newRoleDesc, setNewRoleDesc] = useState('');
  const [newRoleTemplate, setNewRoleTemplate] = useState('');
  const [newRolePerms, setNewRolePerms] = useState({});
  const [isCreating, setIsCreating] = useState(false);

  // Edit Metadata state
  const [isEditModalOpen, setIsEditModalOpen] = useState(false);
  const [editRoleName, setEditRoleName] = useState('');
  const [editRoleDesc, setEditRoleDesc] = useState('');
  const [isSavingMeta, setIsSavingMeta] = useState(false);

  // Delete Role state
  const [isDeleteModalOpen, setIsDeleteModalOpen] = useState(false);
  const [roleToDelete, setRoleToDelete] = useState(null);
  const [isDeleting, setIsDeleting] = useState(false);
  const [deleteError, setDeleteError] = useState('');

  // Dropdown state
  const [openDropdownId, setOpenDropdownId] = useState(null);

  // Toast state
  const [toast, setToast] = useState(null);
  const showToast = (msg, type = 'success') => {
    setToast({ msg, type });
    setTimeout(() => setToast(null), 3000);
  };


  const fetchStats = useCallback(async () => {
    try {
      const data = await api.get('/api/v1/roles/stats');
      setStats(data);
    } catch {
      // ignore
    }
  }, []);

  const fetchRoles = useCallback(async () => {
    setIsLoadingRoles(true);
    try {
      const data = await api.get('/api/v1/roles');
      const list = data || [];
      setRoles(list);
      
      // If we don't have a selection yet, select the first available role (prefer SUPER_ADMIN)
      if (list.length > 0) {
        // Find existing selection or fallback
        let currentId = null;
        // The useCallback dependency on selectedRoleId is intentional to avoid missing updates
        setRoles(currentList => {
          return list;
        });
      }
    } catch {
      // ignore
    } finally {
      setIsLoadingRoles(false);
    }
  }, []);

  // Set initial selected role safely
  useEffect(() => {
    if (roles.length > 0 && !selectedRoleId) {
      const superAdmin = roles.find(r => r.name === 'SUPER_ADMIN');
      const first = superAdmin || roles[0];
      setSelectedRoleId(first.id);
      setSelectedRoleName(first.name);
      setSelectedRoleDesc(first.description || '');
    }
  }, [roles, selectedRoleId]);

  const fetchPermissions = useCallback(async (roleId, roleName, roleDesc) => {
    setIsLoadingPerms(true);
    setIsEditing(false); // Cancel edit mode if role changes
    setSelectedRoleName(roleName || '');
    setSelectedRoleDesc(roleDesc || '');
    try {
      const data = await api.get(`/api/v1/roles/${roleId}/permissions`);
      setPermissions(data.permissions || []);
    } catch {
      setPermissions([]);
    } finally {
      setIsLoadingPerms(false);
    }
  }, []);

  useEffect(() => {
    fetchStats();
    fetchRoles();
  }, [fetchStats, fetchRoles]);

  useEffect(() => {
    if (selectedRoleId) {
      const role = roles.find(r => r.id === selectedRoleId);
      fetchPermissions(selectedRoleId, role?.name, role?.description);
    }
  }, [selectedRoleId]); // eslint-disable-line react-hooks/exhaustive-deps

  useEffect(() => {
    if (isModalOpen || isEditModalOpen || isDeleteModalOpen) {
      document.body.style.overflow = 'hidden';
    } else {
      document.body.style.overflow = 'unset';
    }
    return () => {
      document.body.style.overflow = 'unset';
    };
  }, [isModalOpen, isEditModalOpen, isDeleteModalOpen]);

  const handleRoleClick = (role) => {
    setSelectedRoleId(role.id);
    setSelectedRoleName(role.name);
    setSelectedRoleDesc(role.description || '');
  };

  // --- Edit Permissions Logic ---
  
  const startEditing = () => {
    const initialEdits = {};
    permissions.forEach(p => {
      initialEdits[`${p.resource}.${p.action}`] = true;
    });
    setEditedPerms(initialEdits);
    setIsEditing(true);
  };

  const toggleEditPerm = (resource, action) => {
    const key = `${resource}.${action}`;
    setEditedPerms(prev => ({ ...prev, [key]: !prev[key] }));
  };

  const savePermissions = async () => {
    setIsSaving(true);
    try {
      const permList = Object.entries(editedPerms)
        .filter(([, v]) => v)
        .map(([key]) => {
          const [resource, action] = key.split('.');
          return { resource, action };
        });

      await api.put(`/api/v1/roles/${selectedRoleId}/permissions`, {
        permissions: permList,
      });

      setIsEditing(false);
      // Reload permissions and stats to update counts
      await fetchPermissions(selectedRoleId, selectedRoleName, selectedRoleDesc);
      await fetchRoles(); // to update permission count in the sidebar
      await fetchStats(); // to update global stats
    } catch {
      // TODO: show error toast
    } finally {
      setIsSaving(false);
    }
  };

  const cancelEditing = () => {
    setIsEditing(false);
    setEditedPerms({});
  };

  const isSuperAdmin = selectedRoleName === 'SUPER_ADMIN';

  // --- Create Role Logic ---

  const handleTemplateChange = async (e) => {
    const templateName = e.target.value;
    setNewRoleTemplate(templateName);
    
    if (!templateName) {
      setNewRolePerms({});
      return;
    }

    // Find the role id for this template from the roles list
    const templateRole = roles.find(r => r.name === templateName);
    if (templateRole) {
      try {
        const data = await api.get(`/api/v1/roles/${templateRole.id}/permissions`);
        const templatePermsList = data.permissions || [];
        const permsObj = {};
        templatePermsList.forEach(p => {
          permsObj[`${p.resource}.${p.action}`] = true;
        });
        setNewRolePerms(permsObj);
      } catch (err) {
        console.error("Failed to load template permissions", err);
        setNewRolePerms({});
      }
    }
  };

  const toggleNewRolePerm = (resource, action) => {
    const key = `${resource}.${action}`;
    setNewRolePerms(prev => ({ ...prev, [key]: !prev[key] }));
  };

  const handleCreateRole = async () => {
    if (!newRoleName.trim()) return;
    setIsCreating(true);
    try {
      const permList = Object.entries(newRolePerms)
        .filter(([, v]) => v)
        .map(([key]) => {
          const [resource, action] = key.split('.');
          return { resource, action };
        });

      const newRole = await api.post('/api/v1/roles', {
        name: newRoleName.trim().toUpperCase().replace(/\s+/g, '_'),
        description: newRoleDesc.trim(),
        permissions: permList,
      });

      setIsModalOpen(false);
      setNewRoleName('');
      setNewRoleDesc('');
      setNewRoleTemplate('');
      setNewRolePerms({});
      
      await fetchStats();
      await fetchRoles();
      
      if (newRole && newRole.id) {
        setSelectedRoleId(newRole.id);
        setSelectedRoleName(newRole.name);
        setSelectedRoleDesc(newRole.description);
      }
      showToast('Role created successfully');
    } catch {
      showToast('Failed to create role', 'error');
    } finally {
      setIsCreating(false);
    }
  };

  const handleEditRole = async () => {
    if (!editRoleName.trim()) return;
    setIsSavingMeta(true);
    try {
      await api.put(`/api/v1/roles/${selectedRoleId}`, {
        name: editRoleName.trim().toUpperCase().replace(/\s+/g, '_'),
        description: editRoleDesc.trim()
      });
      setIsEditModalOpen(false);
      await fetchRoles();
      showToast('Role updated successfully');
      setSelectedRoleName(editRoleName.trim().toUpperCase().replace(/\s+/g, '_'));
      setSelectedRoleDesc(editRoleDesc.trim());
    } catch (err) {
      showToast(err.response?.data?.message || 'Failed to update role', 'error');
    } finally {
      setIsSavingMeta(false);
    }
  };

  const handleDeleteRole = async () => {
    if (!roleToDelete) return;
    setIsDeleting(true);
    setDeleteError('');
    try {
      await api.delete(`/api/v1/roles/${roleToDelete.id}`);
      setIsDeleteModalOpen(false);
      setRoleToDelete(null);
      await fetchStats();
      await fetchRoles();
      showToast('Role deleted successfully');
      if (selectedRoleId === roleToDelete.id) {
        const remaining = roles.filter(r => r.id !== roleToDelete.id);
        if (remaining.length > 0) {
          handleRoleClick(remaining[0]);
        }
      }
    } catch (err) {
      if (err.response?.status === 409) {
        setDeleteError('This role is currently assigned to users and cannot be deleted.');
      } else {
        setDeleteError('Failed to delete role.');
      }
    } finally {
      setIsDeleting(false);
    }
  };

  const filteredRoles = roles.filter(r =>
    r.name.toLowerCase().includes(searchQuery.toLowerCase())
  ).sort((a, b) => {
    if (a.name === 'SUPER_ADMIN') return -1;
    if (b.name === 'SUPER_ADMIN') return 1;
    
    const aIsSystem = ROLE_VISUALS[a.name]?.isSystem;
    const bIsSystem = ROLE_VISUALS[b.name]?.isSystem;
    
    if (!aIsSystem && bIsSystem) return -1;
    if (aIsSystem && !bIsSystem) return 1;
    if (!aIsSystem && !bIsSystem) return b.id - a.id;
    return a.id - b.id;
  });

  const hasPermission = (resource, action) => {
    return permissions.some(p => p.resource === resource && p.action === action);
  };

  const totalRoles = stats?.total_roles ?? roles.length;
  const totalPermissions = stats?.total_permissions ?? 40;
  const activeMembers = stats?.active_members ?? '—';
  const systemCoverage = stats?.system_coverage != null ? `${stats.system_coverage}%` : '—';
  const roleCounts = stats?.role_counts ?? [];

  return (
    <div className="roles-page">
      {/* 1. Header */}
      <div className="roles-header">
        <div>
          <h1>Roles & Permissions</h1>
          <p>Manage roles and control what your team can access across LogisticsHQ.</p>
        </div>
        <div>
          <button className="btn-primary" onClick={() => setIsModalOpen(true)}>
            <Plus size={16} /> Create Custom Role
          </button>
        </div>
      </div>

      {/* 2. Stats Cards */}
      <div className="stats-row">
        <div className="stat-card">
          <div className="stat-icon blue"><Shield size={20} /></div>
          <div className="stat-info">
            <h3>{totalRoles}</h3>
            <p className="stat-title">Total Roles</p>
            <p className="stat-desc">Across your workspace</p>
          </div>
        </div>
        <div className="stat-card">
          <div className="stat-icon blue"><ShieldCheck size={20} /></div>
          <div className="stat-info">
            <h3>{totalPermissions}</h3>
            <p className="stat-title">Total Permissions</p>
            <p className="stat-desc">Across all modules</p>
          </div>
        </div>
        <div className="stat-card">
          <div className="stat-icon blue"><Users size={20} /></div>
          <div className="stat-info">
            <h3>{activeMembers}</h3>
            <p className="stat-title">Role Assignments</p>
            <p className="stat-desc">Active users</p>
          </div>
        </div>
        <div className="stat-card">
          <div className="stat-icon blue"><Lock size={20} /></div>
          <div className="stat-info">
            <h3>{systemCoverage}</h3>
            <p className="stat-title">System Coverage</p>
            <p className="stat-desc">Permissions configured</p>
          </div>
        </div>
      </div>

      {/* 3. Main Layout Grid */}
      <div className="roles-layout">
        
        {/* Left Sidebar — Roles List */}
        <div className="roles-sidebar">
          <h2>Roles</h2>
          <div className="search-input-container">
            <Search className="search-icon" size={16} />
            <input 
              type="text" 
              className="search-input" 
              placeholder="Search roles..." 
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
            />
          </div>
          
          <div className="role-list">
            {isLoadingRoles ? (
              <div style={{ padding: '2rem', textAlign: 'center', color: '#64748b' }}>Loading roles...</div>
            ) : filteredRoles.map(role => {
              const visual = ROLE_VISUALS[role.name] || DEFAULT_VISUAL;
              const IconComp = visual.icon;
              return (
                <div 
                  key={role.id} 
                  className={`role-card ${selectedRoleId === role.id ? 'active' : ''}`}
                  onClick={() => handleRoleClick(role)}
                >
                  <div className={`role-card-icon icon-${visual.color}`}>
                    <IconComp size={16} />
                  </div>
                  <div className="role-card-content">
                    <div className="role-card-header">
                      <span className="role-card-name">{role.name}</span>
                      {visual.isSystem && <span className="system-badge">System</span>}
                    </div>
                    <div className="role-card-perms">{role.permission_count} permissions</div>
                    <div className="role-card-desc">{role.description}</div>
                  </div>
                  
                  {/* Dropdown Menu */}
                  <div className="role-dropdown-container">
                    <button 
                      className="role-options-btn" 
                      onClick={(e) => {
                        e.stopPropagation();
                        setOpenDropdownId(openDropdownId === role.id ? null : role.id);
                      }}
                    >
                      <MoreVertical size={16} />
                    </button>
                    {openDropdownId === role.id && (
                      <div className="dropdown-menu">
                        <button onClick={(e) => {
                          e.stopPropagation();
                          handleRoleClick(role);
                          setOpenDropdownId(null);
                        }}>View</button>
                        {!visual.isSystem && (
                          <>
                            <button onClick={(e) => {
                              e.stopPropagation();
                              handleRoleClick(role);
                              setEditRoleName(role.name);
                              setEditRoleDesc(role.description || '');
                              setIsEditModalOpen(true);
                              setOpenDropdownId(null);
                            }}>Edit Role</button>
                            <button className="delete-opt" onClick={(e) => {
                              e.stopPropagation();
                              setRoleToDelete(role);
                              setIsDeleteModalOpen(true);
                              setOpenDropdownId(null);
                            }}>Delete Role</button>
                          </>
                        )}
                      </div>
                    )}
                  </div>
                </div>
              );
            })}
          </div>

          <div className="role-list-footer">
            Showing 1 to {filteredRoles.length} of {roles.length} roles
          </div>
        </div>

        {/* Right Content — Permissions Matrix */}
        <div className="permissions-content">
          <div className="matrix-header">
            <div>
              <h2>{selectedRoleName} Permissions</h2>
              <p>{selectedRoleDesc || 'Select a role to view its permissions'}</p>
            </div>
            
            <div style={{ display: 'flex', gap: '8px' }}>
              {isEditing ? (
                <>
                  <button className="btn-secondary" onClick={cancelEditing} disabled={isSaving}>
                    Cancel
                  </button>
                  <button className="btn-primary" onClick={savePermissions} disabled={isSaving}>
                    {isSaving ? 'Saving...' : 'Save Changes'}
                  </button>
                </>
              ) : (
                <button 
                  className={`btn-outline ${isSuperAdmin ? 'disabled' : ''}`} 
                  onClick={isSuperAdmin ? undefined : startEditing}
                  title={isSuperAdmin ? "SUPER_ADMIN permissions cannot be edited" : "Edit permissions"}
                  style={{ opacity: isSuperAdmin ? 0.5 : 1, cursor: isSuperAdmin ? 'not-allowed' : 'pointer' }}
                >
                  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" style={{marginRight: '6px'}}><path d="M12 20h9"></path><path d="M16.5 3.5a2.121 2.121 0 0 1 3 3L7 19l-4 1 1-4L16.5 3.5z"></path></svg>
                  Edit Permissions
                </button>
              )}
            </div>
          </div>

          <div className="matrix-table-container">
            {isLoadingPerms ? (
              <div style={{ padding: '3rem', textAlign: 'center', color: '#64748b' }}>Loading permissions...</div>
            ) : (
              <table className="matrix-table">
                <thead>
                  <tr>
                    <th>MODULE / RESOURCE</th>
                    {ACTIONS.map(action => (
                      <th key={action}>{action}</th>
                    ))}
                  </tr>
                </thead>
                <tbody>
                  {RESOURCES.map(resource => (
                    <tr key={resource.id}>
                      <td>
                        <span className="resource-name">{resource.name}</span>
                        <span className="resource-desc">{resource.desc}</span>
                      </td>
                      {ACTIONS.map(action => {
                        let isChecked = false;
                        if (isEditing) {
                          isChecked = !!editedPerms[`${resource.id}.${action}`];
                        } else {
                          isChecked = hasPermission(resource.id, action);
                        }

                        return (
                          <td key={action}>
                            {isEditing ? (
                              <input 
                                type="checkbox"
                                className="perm-checkbox"
                                checked={isChecked}
                                onChange={() => toggleEditPerm(resource.id, action)}
                                style={{ width: '18px', height: '18px', cursor: 'pointer' }}
                              />
                            ) : (
                              isChecked ? (
                                <span className="perm-icon allowed"><CheckCircle2 size={18} fill="#10b981" color="white" /></span>
                              ) : (
                                <span className="perm-icon denied"><X size={18} strokeWidth={2.5} /></span>
                              )
                            )}
                          </td>
                        );
                      })}
                    </tr>
                  ))}
                </tbody>
              </table>
            )}
          </div>
          
          <div className="matrix-legend">
            <div className="legend-item">
              <span className="perm-icon allowed"><CheckCircle2 size={16} fill="#10b981" color="white" /></span>
              Allowed
            </div>
            <div className="legend-item">
              <span className="perm-icon denied"><X size={16} strokeWidth={2.5} /></span>
              Not Allowed
            </div>
            <div className="legend-item">
              <span className="perm-icon na"><Minus size={16} /></span>
              Not Applicable
            </div>
          </div>
        </div>
      </div>

      {/* Role Assignments Section */}
      <div className="role-assignments-section">
        <div className="ra-header">
          <div>
            <h3>Role Assignments</h3>
            <p>See which users have been assigned to each role</p>
          </div>
          <button className="btn-outline" onClick={() => navigate('/dashboard/settings/users')}>View All Users</button>
        </div>
        <div className="ra-cards">
          {[...(roleCounts.length > 0 ? roleCounts : roles)].sort((a, b) => {
            const nameA = a.role || a.name;
            const nameB = b.role || b.name;
            if (nameA === 'SUPER_ADMIN') return -1;
            if (nameB === 'SUPER_ADMIN') return 1;
            
            const aIsSystem = ROLE_VISUALS[nameA]?.isSystem;
            const bIsSystem = ROLE_VISUALS[nameB]?.isSystem;
            
            if (!aIsSystem && bIsSystem) return -1;
            if (aIsSystem && !bIsSystem) return 1;
            
            const roleA = roles.find(r => r.name === nameA);
            const roleB = roles.find(r => r.name === nameB);
            const idA = roleA ? roleA.id : 0;
            const idB = roleB ? roleB.id : 0;
            
            if (!aIsSystem && !bIsSystem) return idB - idA;
            return idA - idB;
          }).map((entry, idx) => {
            const roleName  = entry.role || entry.name;
            const userCount = entry.count ?? 0;
            const colorClasses = ['border-green','border-blue','border-purple','border-orange','border-teal','border-red','border-blue'];
            return (
              <div key={roleName} className={`ra-card ${colorClasses[idx % colorClasses.length]}`}>
                <span className="ra-role-name">{roleName}</span>
                <span className="ra-count">{userCount}</span>
                <span className="ra-label">Active Users</span>
              </div>
            );
          })}
        </div>
      </div>

      {/* Create Custom Role Modal */}
      {isModalOpen && createPortal(
        <div className="modal-overlay">
          <div className="modal-content">
            <div className="modal-header">
              <div className="modal-header-titles">
                <h2>Create Custom Role</h2>
                <p>Create a custom role and define exactly what this role can access.</p>
              </div>
              <button className="close-btn" onClick={() => setIsModalOpen(false)}>
                <X size={20} />
              </button>
            </div>
            <div className="modal-body">
              <div className="modal-section-title">1. Role Details</div>
              <div className="form-row">
                <div className="form-group">
                  <label>Role Name <span style={{color: '#ef4444'}}>*</span></label>
                  <input 
                    type="text" 
                    className="form-input" 
                    placeholder="e.g. Sales Manager" 
                    value={newRoleName}
                    onChange={e => setNewRoleName(e.target.value)}
                  />
                </div>
                <div className="form-group">
                  <label>Description</label>
                  <input 
                    type="text" 
                    className="form-input" 
                    placeholder="Brief description of this role" 
                    value={newRoleDesc}
                    onChange={e => setNewRoleDesc(e.target.value)}
                  />
                </div>
              </div>
              
              <div className="modal-section-title">2. Start From Existing Role</div>
              <div className="form-group" style={{ maxWidth: '400px' }}>
                <select 
                  className="form-input" 
                  value={newRoleTemplate}
                  onChange={handleTemplateChange}
                >
                  <option value="">Blank Role</option>
                  {roles.map(r => (
                    <option key={r.id} value={r.name}>{r.name}</option>
                  ))}
                </select>
                <p style={{ fontSize: '0.8rem', color: '#64748b', marginTop: '0.5rem', marginBottom: '0' }}>Choose a role to copy permissions from, or start with a blank role.</p>
              </div>

              <div className="modal-section-title">3. Permissions</div>
              <p style={{ fontSize: '0.875rem', color: '#334155', marginBottom: '1rem' }}>Select the actions this role is allowed to perform for each resource.</p>
              
              <div className="form-group">
                <table className="matrix-table" style={{ border: '1px solid #e2e8f0', borderRadius: '8px' }}>
                  <thead>
                    <tr>
                      <th>RESOURCE</th>
                      {ACTIONS.map(a => <th key={a}>{a}</th>)}
                    </tr>
                  </thead>
                  <tbody>
                    {RESOURCES.map(r => (
                      <tr key={r.id}>
                        <td style={{ fontSize: '0.8rem', fontWeight: 600 }}>{r.name}</td>
                        {ACTIONS.map(a => (
                          <td key={a}>
                            <input 
                              type="checkbox" 
                              className="perm-checkbox"
                              checked={!!newRolePerms[`${r.id}.${a}`]}
                              onChange={() => toggleNewRolePerm(r.id, a)}
                              style={{ width: '16px', height: '16px', cursor: 'pointer' }}
                            />
                          </td>
                        ))}
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
            <div className="modal-footer">
              <button className="btn-secondary" onClick={() => setIsModalOpen(false)}>Cancel</button>
              <button className="btn-primary" onClick={handleCreateRole} disabled={isCreating}>
                {isCreating ? 'Creating...' : 'Create Role'}
              </button>
            </div>
          </div>
        </div>,
        document.body
      )}

      {/* Edit Role Metadata Modal */}
      {isEditModalOpen && createPortal(
        <div className="modal-overlay">
          <div className="modal-content" style={{ maxWidth: '500px' }}>
            <div className="modal-header">
              <h2>Edit Role</h2>
              <button className="close-btn" onClick={() => setIsEditModalOpen(false)}>
                <X size={20} />
              </button>
            </div>
            <div className="modal-body">
              <div className="form-group">
                <label>Role Name</label>
                <input 
                  type="text" 
                  className="form-input" 
                  value={editRoleName}
                  onChange={e => setEditRoleName(e.target.value)}
                />
              </div>
              <div className="form-group">
                <label>Description</label>
                <input 
                  type="text" 
                  className="form-input" 
                  value={editRoleDesc}
                  onChange={e => setEditRoleDesc(e.target.value)}
                />
              </div>
            </div>
            <div className="modal-footer">
              <button className="btn-secondary" onClick={() => setIsEditModalOpen(false)}>Cancel</button>
              <button className="btn-primary" onClick={handleEditRole} disabled={isSavingMeta}>
                {isSavingMeta ? 'Saving...' : 'Save Changes'}
              </button>
            </div>
          </div>
        </div>,
        document.body
      )}

      {/* Delete Role Modal */}
      {isDeleteModalOpen && roleToDelete && createPortal(
        <div className="modal-overlay">
          <div className="modal-content" style={{ maxWidth: '450px' }}>
            <div className="modal-header">
              <h2>Delete Role</h2>
              <button className="close-btn" onClick={() => setIsDeleteModalOpen(false)}>
                <X size={20} />
              </button>
            </div>
            <div className="modal-body">
              <p>Delete "{roleToDelete.name}"? This action cannot be undone.</p>
              {deleteError && (
                <div style={{ marginTop: '1rem', padding: '0.75rem', backgroundColor: '#fef2f2', color: '#ef4444', borderRadius: '6px', fontSize: '0.875rem' }}>
                  {deleteError}
                </div>
              )}
            </div>
            <div className="modal-footer">
              <button className="btn-secondary" onClick={() => setIsDeleteModalOpen(false)}>Cancel</button>
              <button className="btn-primary" style={{ backgroundColor: '#ef4444' }} onClick={handleDeleteRole} disabled={isDeleting}>
                {isDeleting ? 'Deleting...' : 'Delete Role'}
              </button>
            </div>
          </div>
        </div>,
        document.body
      )}

      {/* Toast Notification */}
      {toast && (
        <div className={`toast ${toast.type}`}>
          {toast.msg}
        </div>
      )}
    </div>
  );
}

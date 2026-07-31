import { useState, useEffect } from 'react';
import api from '../../../services/api';
import './RolesPage.css';

// The system resources and actions we support
const RESOURCES = ['COMPANIES', 'LEADS', 'OPPORTUNITIES', 'RFQS', 'OUTREACH'];
const ACTIONS = ['CREATE', 'READ', 'UPDATE', 'DELETE'];

export default function RolesPage() {
  const [roles, setRoles] = useState([]);
  const [selectedRoleId, setSelectedRoleId] = useState(null);
  const [permissions, setPermissions] = useState([]); // Array of { resource, action }
  
  const [isLoadingRoles, setIsLoadingRoles] = useState(true);
  const [isLoadingPerms, setIsLoadingPerms] = useState(false);
  const [isSaving, setIsSaving] = useState(false);
  
  const [error, setError] = useState(null);
  const [successMsg, setSuccessMsg] = useState(null);

  // 1. Fetch all roles on mount
  useEffect(() => {
    fetchRoles();
  }, []);

  const fetchRoles = async () => {
    setIsLoadingRoles(true);
    try {
      const data = await api.get('/api/v1/roles');
      setRoles(data || []);
      if (data && data.length > 0) {
        setSelectedRoleId(data[0].id);
      }
    } catch (err) {
      setError('Failed to fetch roles. Please try again.');
    } finally {
      setIsLoadingRoles(false);
    }
  };

  // 2. Fetch permissions whenever selected role changes
  useEffect(() => {
    if (selectedRoleId) {
      fetchPermissions(selectedRoleId);
    }
  }, [selectedRoleId]);

  const fetchPermissions = async (roleId) => {
    setIsLoadingPerms(true);
    setSuccessMsg(null);
    try {
      const data = await api.get(`/api/v1/roles/${roleId}/permissions`);
      setPermissions(data.permissions || []);
    } catch (err) {
      setError('Failed to fetch permissions for the selected role.');
      setPermissions([]);
    } finally {
      setIsLoadingPerms(false);
    }
  };

  // Helper: Check if current local state has a specific permission
  const hasPermission = (resource, action) => {
    return permissions.some(p => p.resource === resource && p.action === action);
  };

  // Toggle a single checkbox
  const togglePermission = (resource, action) => {
    setSuccessMsg(null); // Clear success message on edit
    
    // Some roles (like SUPER_ADMIN) might be locked or read-only, 
    // but we'll allow editing and rely on backend/UI hints for now.
    const exists = hasPermission(resource, action);
    
    if (exists) {
      setPermissions(permissions.filter(p => !(p.resource === resource && p.action === action)));
    } else {
      setPermissions([...permissions, { resource, action }]);
    }
  };

  const handleSave = async () => {
    if (!selectedRoleId) return;
    
    setIsSaving(true);
    setError(null);
    setSuccessMsg(null);
    
    try {
      await api.put(`/api/v1/roles/${selectedRoleId}/permissions`, {
        permissions: permissions
      });
      setSuccessMsg('Permissions successfully updated.');
      
      // Auto-hide success message after 3 seconds
      setTimeout(() => setSuccessMsg(null), 3000);
    } catch (err) {
      setError(err.message || 'Failed to save permissions.');
    } finally {
      setIsSaving(false);
    }
  };

  const selectedRole = roles.find(r => r.id === selectedRoleId);

  return (
    <div className="roles-page">
      <div className="roles-header">
        <div>
          <h1>Role Permissions</h1>
          <p>Configure granular access control for different team roles across the platform.</p>
        </div>
      </div>

      {error && (
        <div className="form-alert error" style={{ marginBottom: '1.5rem' }}>
          {error}
        </div>
      )}

      {successMsg && (
        <div className="form-alert success" style={{ marginBottom: '1.5rem', backgroundColor: '#ecfdf5', color: '#065f46', border: '1px solid #34d399' }}>
          {successMsg}
        </div>
      )}

      {isLoadingRoles ? (
        <div style={{ padding: '4rem', textAlign: 'center' }}>
          <div className="auth-spinner" style={{ margin: '0 auto' }}></div>
        </div>
      ) : roles.length === 0 ? (
        <div className="roles-empty">
          <h3>No Roles Found</h3>
          <p>There are no roles available for this organization.</p>
        </div>
      ) : (
        <div className="roles-container">
          
          {/* LEFT SIDEBAR: ROLES LIST */}
          <div className="roles-sidebar">
            {roles.map(role => (
              <div 
                key={role.id} 
                className={`role-item ${selectedRoleId === role.id ? 'active' : ''}`}
                onClick={() => setSelectedRoleId(role.id)}
              >
                <span className="role-name">{role.name}</span>
                <span className="role-desc">{role.description}</span>
              </div>
            ))}
          </div>

          {/* RIGHT CONTENT: PERMISSIONS MATRIX */}
          <div className="roles-content">
            <div className="rc-header">
              <h2>{selectedRole?.name} Permissions</h2>
              <button 
                className="btn-primary" 
                onClick={handleSave} 
                disabled={isSaving || isLoadingPerms}
              >
                {isSaving ? 'Saving...' : 'Save Changes'}
              </button>
            </div>

            {isLoadingPerms ? (
              <div style={{ padding: '4rem', textAlign: 'center' }}>
                <div className="auth-spinner" style={{ margin: '0 auto' }}></div>
              </div>
            ) : (
              <table className="matrix-table">
                <thead>
                  <tr>
                    <th>Module / Resource</th>
                    {ACTIONS.map(action => (
                      <th key={action}>{action}</th>
                    ))}
                  </tr>
                </thead>
                <tbody>
                  {RESOURCES.map(resource => (
                    <tr key={resource}>
                      <td>{resource.charAt(0) + resource.slice(1).toLowerCase()}</td>
                      {ACTIONS.map(action => (
                        <td key={action}>
                          <input 
                            type="checkbox" 
                            className="perm-checkbox"
                            checked={hasPermission(resource, action)}
                            onChange={() => togglePermission(resource, action)}
                            // We can disable SUPER_ADMIN editing to prevent lockout, 
                            // but allowing it is fine for MVP.
                            disabled={selectedRole?.name === 'SUPER_ADMIN'} 
                            title={selectedRole?.name === 'SUPER_ADMIN' ? 'SUPER_ADMIN permissions cannot be modified.' : `Toggle ${action} on ${resource}`}
                          />
                        </td>
                      ))}
                    </tr>
                  ))}
                </tbody>
              </table>
            )}
          </div>
        </div>
      )}
    </div>
  );
}

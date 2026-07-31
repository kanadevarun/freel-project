import { NavLink } from 'react-router-dom';
import { useAuth } from '../../context/AuthContext';
import { useRBAC } from '../../context/RBACContext';
import './Sidebar.css';

const NAV_ITEMS = [
  // ── SALES ──────────────────────────────────────────────────────────
  {
    section: 'SALES',
    items: [
      { path: '/dashboard/leads', label: 'Leads Pipeline', icon: '🎯', module: 'LEADS', action: 'READ' },
      { path: '/dashboard/rfqs', label: 'RFQ Management', icon: '📝', module: 'RFQS', action: 'READ' },
      { path: '/dashboard/outreach', label: 'Email Outreach', icon: '✉️', module: 'OUTREACH', action: 'READ', badge: '3' },
      { path: '/dashboard/companies', label: 'Company Directory', icon: '🏢', module: 'COMPANIES', action: 'READ' },
    ]
  },
  // ── INTELLIGENCE ───────────────────────────────────────────────────
  {
    section: 'INTELLIGENCE',
    items: [
      { path: '/dashboard/market-insights', label: 'Market Insights', icon: '📊' }, // Public module (no restriction)
      { path: '/dashboard/routes', label: 'Route Optimization', icon: '🗺️', module: 'ROUTES', action: 'READ' },
    ]
  },
  // ── TOOLS ──────────────────────────────────────────────────────────
  {
    section: 'TOOLS',
    items: [
      { path: '/dashboard/calculators', label: 'Calculators', icon: '🧮' },
      { path: '/dashboard/documents', label: 'Document Generator', icon: '📄' },
    ]
  },
];

export default function Sidebar() {
  const { user, org, memberRole, logout } = useAuth();
  const { can, hasRole } = useRBAC();

  const handleLogout = async () => {
    await logout();
    window.location.href = '/';
  };

  // Normalize role name for display (e.g. SUPER_ADMIN -> Super Admin)
  const displayRole = (typeof memberRole === 'object' ? memberRole?.display_name : memberRole) || 'Member';

  return (
    <aside className="app-sidebar">
      <div className="sidebar-header">
        <div className="sidebar-brand">Freel</div>
        <div className="sidebar-org">{org?.name || 'Your Workspace'}</div>
      </div>

      <nav className="sidebar-nav">
        {NAV_ITEMS.map((group) => {
          // Filter items based on permissions
          const visibleItems = group.items.filter(item => {
            if (!item.module || !item.action) return true; // No RBAC required
            return can(item.module, item.action);
          });

          if (visibleItems.length === 0) return null;

          return (
            <div key={group.section} className="nav-group">
              <div className="nav-group-title">{group.section}</div>
              <div className="nav-group-items">
                {visibleItems.map(item => (
                  <NavLink
                    key={item.path}
                    to={item.path}
                    className={({ isActive }) => `nav-item ${isActive ? 'active' : ''}`}
                  >
                    <span className="nav-icon">{item.icon}</span>
                    <span className="nav-label">{item.label}</span>
                    {item.badge && <span className="nav-badge">{item.badge}</span>}
                  </NavLink>
                ))}
              </div>
            </div>
          );
        })}

        {/* ADMIN SECTION (SUPER_ADMIN only) */}
        {hasRole('SUPER_ADMIN') && (
          <div className="nav-group admin-group">
            <div className="nav-group-title">ADMINISTRATION</div>
            <div className="nav-group-items">
              <NavLink to="/dashboard/settings" className={({ isActive }) => `nav-item ${isActive ? 'active' : ''}`}>
                <span className="nav-icon">⚙️</span>
                <span className="nav-label">Workspace Settings</span>
              </NavLink>
              <NavLink to="/dashboard/users" className={({ isActive }) => `nav-item ${isActive ? 'active' : ''}`}>
                <span className="nav-icon">👥</span>
                <span className="nav-label">User Management</span>
              </NavLink>
            </div>
          </div>
        )}
      </nav>

      <div className="sidebar-footer">
        <div className="user-profile-card" onClick={handleLogout} title="Click to logout">
          <div className="user-avatar">
            {user?.email?.[0].toUpperCase() || 'U'}
          </div>
          <div className="user-info">
            <div className="user-name">{user?.full_name || user?.email?.split('@')[0]}</div>
            <div className="user-role-badge">{displayRole}</div>
          </div>
          <div className="user-chevron">↑</div>
        </div>
      </div>
    </aside>
  );
}

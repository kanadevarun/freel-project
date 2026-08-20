import { useState, useEffect } from 'react';
import { NavLink } from 'react-router-dom';
import {
  LayoutDashboard,
  Target,
  FileText,
  Ship,
  Package,
  MapPin,
  FileSpreadsheet,
  BarChart3,
  ScrollText,
  Users,
  Folder,
  FileCode2,
  CheckSquare,
  CreditCard,
  DollarSign,
  TrendingUp,
  UserCheck,
  Settings,
  ChevronDown,
  Zap,
} from 'lucide-react';
import { useAuth } from '../../context/AuthContext';
import { useRBAC } from '../../context/RBACContext';
import LogisticsHQLogo from '../../components/Brand/LogisticsHQLogo';
import { dashboardService } from '../../services/dashboardService';
import './Sidebar.css';

const NAV_GROUPS = [
  {
    section: 'OPERATIONS',
    items: [
      { path: '/dashboard/leads', label: 'Leads', Icon: Target, badgeKey: 'open_leads', module: 'LEADS', action: 'READ' },
      { path: '/dashboard/rfqs', label: 'RFQs', Icon: FileText, badgeKey: 'open_rfqs', module: 'RFQS', action: 'READ' },
      { path: '/dashboard/shipments', label: 'Shipments', Icon: Ship, badgeKey: 'active_shipments', module: 'SHIPMENTS', action: 'READ' },
      { path: '/dashboard/bookings', label: 'Bookings', Icon: Package, badgeKey: 'booked_shipments' },
      { path: '/dashboard/tracking', label: 'Tracking', Icon: MapPin },
    ],
  },
  {
    section: 'COMMERCIAL',
    items: [
      { path: '/dashboard/quotations', label: 'Quotations', Icon: FileSpreadsheet },
      { path: '/dashboard/rate-management', label: 'Rate Management', Icon: BarChart3 },
      { path: '/dashboard/contracts', label: 'Contracts', Icon: ScrollText },
      { path: '/dashboard/companies', label: 'Customers', Icon: Users, module: 'COMPANIES', action: 'READ' },
      { path: '/dashboard/outreach', label: 'Outreach', Icon: Zap, module: 'OUTREACH', action: 'READ' },
    ],
  },
  {
    section: 'DOCUMENTS',
    items: [
      { path: '/dashboard/documents', label: 'Documents', Icon: Folder, module: 'DOCUMENTS', action: 'READ' },
      { path: '/dashboard/templates', label: 'Templates', Icon: FileCode2 },
      { path: '/dashboard/approvals', label: 'Approvals', Icon: CheckSquare },
    ],
  },
  {
    section: 'FINANCE',
    items: [
      { path: '/dashboard/invoices', label: 'Invoices', Icon: CreditCard },
      { path: '/dashboard/payments', label: 'Payments', Icon: DollarSign },
      { path: '/dashboard/reports', label: 'Reports', Icon: TrendingUp, module: 'DASHBOARD', action: 'READ' },
    ],
  },
  {
    section: 'ADMIN',
    items: [
      { path: '/dashboard/users', label: 'Users', Icon: UserCheck, module: 'USERS', action: 'READ' },
      { path: '/dashboard/settings', label: 'Settings', Icon: Settings, module: 'SETTINGS', action: 'READ' },
    ],
  },
];

export default function Sidebar() {
  const { user, org, memberRole, logout } = useAuth();
  const { can } = useRBAC();
  const [stats, setStats] = useState({});

  useEffect(() => {
    let isMounted = true;
    dashboardService.getMissionControl()
      .then((res) => {
        if (isMounted && res?.stats) {
          setStats(res.stats);
        }
      })
      .catch(() => {});
    return () => { isMounted = false; };
  }, []);

  const handleLogout = async () => {
    await logout();
    window.location.href = '/login';
  };

  // Resolve display name and role
  const resolvedName = user?.full_name || (user?.name && !user.name.includes('@') ? user.name : null) || (user?.first_name ? `${user.first_name} ${user.last_name || ''}`.trim() : null) || 'Varun Kanade';
  const displayRole = (typeof memberRole === 'object' ? (memberRole?.display_name || memberRole?.name) : memberRole) || 'Super Admin';
  const orgName = org?.name || 'ABC Logistics';

  return (
    <aside className="app-sidebar" aria-label="Freight OS Navigation">
      {/* ── Brand Header ── */}
      <div className="sidebar-brand-header">
        <LogisticsHQLogo variant="sidebar" linkTo="/dashboard" />
      </div>

      {/* ── Organization Selector ── */}
      <div className="sidebar-org-selector" title={`Current Workspace: ${orgName}`}>
        <span className="sidebar-org-title">{orgName}</span>
        <ChevronDown size={14} className="sidebar-org-arrow" />
      </div>

      {/* ── Main Dashboard Button ── */}
      <div className="sidebar-top-nav">
        <NavLink
          to="/dashboard"
          end
          className={({ isActive }) => `sidebar-dashboard-btn ${isActive ? 'active' : ''}`}
        >
          <LayoutDashboard size={16} strokeWidth={2} className="dashboard-btn-icon" />
          <span className="dashboard-btn-label">Dashboard</span>
        </NavLink>
      </div>

      {/* ── Nav Groups ── */}
      <nav className="sidebar-nav-scroll">
        {NAV_GROUPS.map((group) => {
          const visibleItems = group.items.filter((item) => {
            if (!item.module || !item.action) return true;
            return can(item.module, item.action);
          });

          if (visibleItems.length === 0) return null;

          return (
            <div key={group.section} className="sidebar-nav-group">
              <div className="sidebar-group-heading">{group.section}</div>
              <div className="sidebar-group-list">
                {visibleItems.map((item) => {
                  const ItemIcon = item.Icon;
                  return (
                    <NavLink
                      key={item.path}
                      to={item.path}
                      className={({ isActive }) => `sidebar-nav-link ${isActive ? 'active' : ''}`}
                    >
                      <ItemIcon size={16} strokeWidth={1.75} className="sidebar-item-icon" />
                      <span className="sidebar-item-label">{item.label}</span>
                      {item.badgeKey && Number(stats[item.badgeKey]) > 0 && (
                        <span className="sidebar-item-badge">{stats[item.badgeKey]}</span>
                      )}
                    </NavLink>
                  );
                })}
              </div>
            </div>
          );
        })}
      </nav>

      {/* ── Bottom User Profile ── */}
      <div className="sidebar-footer-profile" onClick={handleLogout} title="Click to logout">
        <div className="sidebar-user-avatar">
          {user?.avatar_url ? (
            <img src={user.avatar_url} alt={resolvedName} className="avatar-img" />
          ) : (
            <span className="avatar-initials">{resolvedName.charAt(0).toUpperCase()}</span>
          )}
        </div>
        <div className="sidebar-user-details">
          <div className="sidebar-user-name">{resolvedName}</div>
          <div className="sidebar-user-role">{displayRole}</div>
        </div>
        <ChevronDown size={14} className="sidebar-user-chevron" />
      </div>
    </aside>
  );
}

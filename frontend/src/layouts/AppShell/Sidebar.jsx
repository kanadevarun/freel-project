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
  CheckSquare,
  CreditCard,
  TrendingUp,
  Settings,
  ChevronDown,
  Zap,
  Lock,
  FileWarning,
  Compass,
  Cpu,
  Bell,
  PanelLeftClose,
  PanelLeftOpen,
  Sliders,
} from 'lucide-react';
import { useAuth } from '../../context/AuthContext';
import { useRBAC } from '../../context/RBACContext';
import LogisticsHQLogo from '../../components/Brand/LogisticsHQLogo';
import { dashboardService } from '../../services/dashboardService';
import { notificationCenterService } from '../../services/notificationCenterService';
import './Sidebar.css';

const NAV_GROUPS = [
  {
    section: 'OPERATIONS',
    items: [
      { path: '/dashboard/command-center', label: 'Command Center', Icon: Sliders, module: 'SHIPMENTS', action: 'READ' },
      { path: '/dashboard/leads', label: 'Leads', Icon: Target, badgeKey: 'open_leads', module: 'LEADS', action: 'READ' },
      { path: '/dashboard/rfqs', label: 'RFQs', Icon: FileText, badgeKey: 'open_rfqs', module: 'RFQS', action: 'READ' },
      { path: '/dashboard/shipments', label: 'Shipments', Icon: Ship, badgeKey: 'active_shipments', module: 'SHIPMENTS', action: 'READ' },
      { path: '/dashboard/bookings', label: 'Bookings', Icon: Package, badgeKey: 'booked_shipments', module: 'SHIPMENTS', action: 'READ' },
      { path: '/dashboard/tracking', label: 'Tracking', Icon: MapPin, module: 'SHIPMENTS', action: 'READ' },
    ],
  },
  {
    section: 'COMMERCIAL',
    items: [
      { path: '/dashboard/quotations', label: 'Quotations', Icon: FileSpreadsheet, module: 'OPPORTUNITIES', action: 'READ' },
      { path: '/dashboard/rate-management', label: 'Rate Management', Icon: BarChart3, module: 'RFQS', action: 'READ' },
      { path: '/dashboard/contracts', label: 'Contracts', Icon: ScrollText, module: 'DOCUMENTS', action: 'READ' },
      { path: '/dashboard/customers', label: 'Customers', Icon: Users, module: 'COMPANIES', action: 'READ' },
      { path: '/dashboard/outreach', label: 'Outreach', Icon: Zap, module: 'OUTREACH', action: 'READ' },
    ],
  },
  {
    section: 'DOCUMENTS',
    items: [
      { path: '/dashboard/documents', label: 'Documents', Icon: Folder, module: 'DOCUMENTS', action: 'READ' },
      { path: '/dashboard/approvals', label: 'Approvals', Icon: CheckSquare, module: 'DOCUMENTS', action: 'READ' },
      { path: '/dashboard/recommendations', label: 'Recommendations', Icon: Compass, module: 'DOCUMENTS', action: 'READ' },
      { path: '/dashboard/automations', label: 'AI Automations', Icon: Cpu, module: 'DOCUMENTS', action: 'READ' },
      { path: '/dashboard/notifications', label: 'Notifications', Icon: Bell, badgeKey: 'unread_notifications', module: 'DOCUMENTS', action: 'READ' },
    ],
  },
  {
    section: 'FINANCE',
    items: [
      { path: '/dashboard/invoices', label: 'Invoices', Icon: CreditCard, module: 'FINANCE', action: 'READ', sensitive: true },
      { path: '/dashboard/debit-notes', label: 'Debit Notes', Icon: FileWarning, module: 'FINANCE', action: 'READ', sensitive: true },
      { path: '/dashboard/reports', label: 'Reports', Icon: TrendingUp, module: 'DASHBOARD', action: 'READ', sensitive: true },
    ],
  },
  {
    section: 'ADMIN',
    items: [
      { path: '/dashboard/settings', label: 'Settings', Icon: Settings, module: 'SETTINGS', action: 'READ', sensitive: true },
    ],
  },
];

export default function Sidebar({ isCollapsed = false, onToggleCollapse }) {
  const { user, org, memberRole, logout } = useAuth();
  const { can } = useRBAC();
  const [stats, setStats] = useState({});

  useEffect(() => {
    let isMounted = true;
    Promise.all([
      dashboardService.getMissionControl().catch(() => ({})),
      notificationCenterService.getUnreadCount().catch(() => ({ count: 0 })),
    ]).then(([missionRes, notifRes]) => {
      if (isMounted) {
        setStats({
          ...(missionRes?.stats || {}),
          unread_notifications: notifRes?.count || 0,
        });
      }
    });
    return () => { isMounted = false; };
  }, []);

  const handleLogout = async () => {
    await logout();
    window.location.href = '/login';
  };

  // Resolve display name and role
  const resolvedName = user?.full_name || (user?.name && !user.name.includes('@') ? user.name : null) || (user?.first_name ? `${user.first_name} ${user.last_name || ''}`.trim() : null) || user?.email || 'User';
  const displayRole = (typeof memberRole === 'object' ? (memberRole?.display_name || memberRole?.name) : memberRole) || 'Super Admin';
  const orgName = org?.name || 'Workspace';
  const userInitials = (resolvedName.charAt(0) || 'U').toUpperCase();

  return (
    <aside
      className={`app-sidebar ${isCollapsed ? 'collapsed' : ''}`}
      aria-label="Freight OS Navigation"
      aria-expanded={!isCollapsed}
    >
      {/* ── Brand Header ── */}
      <div className="sidebar-brand-header">
        <div className="sidebar-brand-logo-area">
          <LogisticsHQLogo
            variant="sidebar"
            showText={!isCollapsed}
            linkTo="/dashboard"
          />
        </div>
        {onToggleCollapse && (
          <button
            type="button"
            className="sidebar-collapse-btn"
            onClick={onToggleCollapse}
            title={isCollapsed ? 'Expand Sidebar (Ctrl+[)' : 'Collapse Sidebar (Ctrl+[)'}
            aria-label={isCollapsed ? 'Expand Sidebar' : 'Collapse Sidebar'}
          >
            {isCollapsed ? <PanelLeftOpen size={16} strokeWidth={2} /> : <PanelLeftClose size={16} strokeWidth={2} />}
          </button>
        )}
      </div>

      {/* ── Organization Selector ── */}
      <div
        className="sidebar-org-selector"
        title={`Current Workspace: ${orgName}`}
        role="button"
        tabIndex={0}
      >
        <div className="sidebar-org-initial-badge" aria-label={orgName}>
          {orgName.charAt(0).toUpperCase()}
        </div>
        {!isCollapsed && (
          <>
            <span className="sidebar-org-title">{orgName}</span>
            <ChevronDown size={14} className="sidebar-org-arrow" />
          </>
        )}
      </div>

      {/* ── Main Dashboard Button ── */}
      <div className="sidebar-top-nav">
        <NavLink
          to="/dashboard"
          end
          className={({ isActive }) => `sidebar-dashboard-btn ${isActive ? 'active' : ''}`}
          title={isCollapsed ? 'Dashboard' : undefined}
          aria-label="Dashboard"
        >
          <LayoutDashboard size={17} strokeWidth={2} className="dashboard-btn-icon" />
          {!isCollapsed && <span className="dashboard-btn-label">Dashboard</span>}
        </NavLink>
      </div>

      {/* ── Nav Groups ── */}
      <nav className="sidebar-nav-scroll" aria-label="Main Navigation">
        {NAV_GROUPS.map((group) => {
          const visibleItems = group.items.map((item) => {
            const hasAccess = (!item.module || !item.action) ? true : can(item.module, item.action);
            return { ...item, hasAccess };
          }).filter((item) => {
            if (item.sensitive && !item.hasAccess) return false;
            return true;
          });

          if (visibleItems.length === 0) return null;

          return (
            <div key={group.section} className="sidebar-nav-group">
              {isCollapsed ? (
                <div className="sidebar-collapsed-divider" />
              ) : (
                <div className="sidebar-group-heading">{group.section}</div>
              )}
              <div className="sidebar-group-list">
                {visibleItems.map((item) => {
                  const ItemIcon = item.Icon;
                  const badgeCount = item.badgeKey ? Number(stats[item.badgeKey]) : 0;
                  const itemTitle = isCollapsed ? `${item.label}${badgeCount > 0 ? ` (${badgeCount})` : ''}` : undefined;

                  if (!item.hasAccess) {
                    return (
                      <div
                        key={item.path}
                        className="sidebar-nav-link locked"
                        title={isCollapsed ? `${item.label} (Locked)` : 'Upgrade or request permissions to access'}
                        aria-label={`${item.label} (Locked)`}
                      >
                        <ItemIcon size={16} strokeWidth={1.75} className="sidebar-item-icon" />
                        {!isCollapsed && <span className="sidebar-item-label">{item.label}</span>}
                        {!isCollapsed && (
                          <Lock size={13} className="sidebar-lock-icon" style={{ marginLeft: 'auto', color: '#94A3B8' }} />
                        )}
                      </div>
                    );
                  }

                  return (
                    <NavLink
                      key={item.path}
                      to={item.path}
                      className={({ isActive }) => `sidebar-nav-link ${isActive ? 'active' : ''}`}
                      title={itemTitle}
                      aria-label={item.label}
                    >
                      <ItemIcon size={16} strokeWidth={1.75} className="sidebar-item-icon" />
                      {!isCollapsed && <span className="sidebar-item-label">{item.label}</span>}
                      {!isCollapsed && badgeCount > 0 && (
                        <span className="sidebar-item-badge">{badgeCount}</span>
                      )}
                      {isCollapsed && badgeCount > 0 && (
                        <span className="sidebar-collapsed-badge-dot" title={`${badgeCount} active`} />
                      )}
                    </NavLink>
                  );
                })}
              </div>
            </div>
          );
        })}
      </nav>

      {/* ── Bottom Profile Card ── */}
      <div
        className="sidebar-footer-profile"
        title={`${resolvedName} (${displayRole})`}
        role="button"
        tabIndex={0}
      >
        <div className="sidebar-user-avatar">
          {userInitials}
        </div>
        {!isCollapsed && (
          <>
            <div className="sidebar-user-details">
              <span className="sidebar-user-name">{resolvedName}</span>
              <span className="sidebar-user-role">{displayRole}</span>
            </div>
            <ChevronDown size={14} className="sidebar-user-chevron" />
          </>
        )}
      </div>
    </aside>
  );
}

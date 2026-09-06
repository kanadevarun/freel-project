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
  DollarSign,
  TrendingUp,
  UserCheck,
  Settings,
  ChevronDown,
  Zap,
  Lock,
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
    ],
  },
  {
    section: 'FINANCE',
    items: [
      { path: '/dashboard/invoices', label: 'Invoices', Icon: CreditCard, module: 'FINANCE', action: 'READ', sensitive: true },
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
              <div className="sidebar-group-heading">{group.section}</div>
              <div className="sidebar-group-list">
                {visibleItems.map((item) => {
                  const ItemIcon = item.Icon;
                  
                  if (!item.hasAccess) {
                    return (
                      <div key={item.path} className="sidebar-nav-link locked" title="Upgrade or request permissions to access">
                        <ItemIcon size={16} strokeWidth={1.75} className="sidebar-item-icon" />
                        <span className="sidebar-item-label">{item.label}</span>
                        <Lock size={14} className="sidebar-lock-icon" style={{ marginLeft: 'auto', color: '#94A3B8' }} />
                      </div>
                    );
                  }

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

    </aside>
  );
}

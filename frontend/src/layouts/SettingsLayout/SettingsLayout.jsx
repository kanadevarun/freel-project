import { NavLink, Outlet } from 'react-router-dom';
import { 
  Building2, 
  Settings, 
  Users, 
  UserPlus, 
  ShieldCheck, 
  Key, 
  ShieldAlert, 
  FileText, 
  Truck, 
  Mail, 
  Blocks, 
  CreditCard, 
  Receipt 
} from 'lucide-react';
import './SettingsLayout.css';

export default function SettingsLayout() {
  return (
    <div className="settings-layout">
      
      {/* Settings Navigation Sidebar */}
      <aside className="settings-sidebar">
        <h1 className="settings-sidebar-header">Settings</h1>

        {/* SETTINGS */}
        <div className="settings-nav-section">
          <ul className="settings-nav-list">
            <li className="settings-nav-item">
              <NavLink to="/dashboard/settings/company-profile" className="settings-nav-link">
                <Building2 className="settings-nav-icon" />
                Company Profile
              </NavLink>
            </li>
            <li className="settings-nav-item">
              <NavLink to="/dashboard/settings/workspace" className="settings-nav-link">
                <Settings className="settings-nav-icon" />
                Workspace Settings
              </NavLink>
            </li>
          </ul>
        </div>

        {/* TEAM & ACCESS */}
        <div className="settings-nav-section">
          <div className="settings-nav-title">TEAM & ACCESS</div>
          <ul className="settings-nav-list">
            <li className="settings-nav-item">
              <NavLink to="/dashboard/settings/users" className="settings-nav-link">
                <Users className="settings-nav-icon" />
                Users & Team
              </NavLink>
            </li>
            <li className="settings-nav-item">
              <NavLink to="/dashboard/settings/roles" className="settings-nav-link">
                <ShieldCheck className="settings-nav-icon" />
                Roles & Permissions
              </NavLink>
            </li>
          </ul>
        </div>

        {/* SECURITY */}
        <div className="settings-nav-section">
          <div className="settings-nav-title">SECURITY</div>
          <ul className="settings-nav-list">
            <li className="settings-nav-item">
              <NavLink to="/dashboard/settings/audit-logs" className="settings-nav-link">
                <FileText className="settings-nav-icon" />
                Audit Logs
              </NavLink>
            </li>
          </ul>
        </div>

        {/* INTEGRATIONS */}
        <div className="settings-nav-section">
          <div className="settings-nav-title">INTEGRATIONS</div>
          <ul className="settings-nav-list">
            <li className="settings-nav-item">
              <NavLink to="/dashboard/settings/carrier-integrations" className="settings-nav-link">
                <Truck className="settings-nav-icon" />
                Carrier Integrations
              </NavLink>
            </li>
            <li className="settings-nav-item">
              <NavLink to="/dashboard/settings/email-settings" className="settings-nav-link">
                <Mail className="settings-nav-icon" />
                Email Settings
              </NavLink>
            </li>
          </ul>
        </div>

        {/* BILLING */}
        <div className="settings-nav-section">
          <div className="settings-nav-title">BILLING</div>
          <ul className="settings-nav-list">
            <li className="settings-nav-item">
              <NavLink to="/dashboard/settings/subscription" className="settings-nav-link">
                <CreditCard className="settings-nav-icon" />
                Subscription
              </NavLink>
            </li>
          </ul>
        </div>
      </aside>

      {/* Main Content Area */}
      <div className="settings-content">
        <Outlet />
      </div>

    </div>
  );
}

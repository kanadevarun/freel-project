import { Outlet } from 'react-router-dom';
import { useAuth } from '../../context/AuthContext';
import { authStorage } from '../../utils/authStorage';
import './AppShell.css';

export default function AppShell() {
  const { user, memberRole } = useAuth();

  const handleLogout = () => {
    authStorage.clearAll();
    window.location.href = '/';
  };

  return (
    <div className="app-shell">
      <header className="app-header">
        <div className="header-brand">
          <div className="brand-logo">Freel</div>
          <span className="brand-badge">Workspace</span>
        </div>
        <div className="header-actions">
          <div className="user-profile-badge">
            <span className="user-role">{memberRole || 'Member'}</span>
            <span className="user-email">{user?.email}</span>
          </div>
          <button className="btn-logout" onClick={handleLogout}>
            Logout
          </button>
        </div>
      </header>
      
      <div className="app-content-wrapper">
        <main className="app-shell-main">
          <Outlet />
        </main>
      </div>
    </div>
  );
}

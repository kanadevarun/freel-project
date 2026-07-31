import { Outlet } from 'react-router-dom';
import Sidebar from './Sidebar';
import TopBar from './TopBar';
import './AppShell.css';

export default function AppShell() {
  return (
    <div className="app-shell">
      <Sidebar />
      <div className="app-shell-content">
        <TopBar />
        <main className="app-shell-main">
          <div className="app-shell-container">
            <Outlet />
          </div>
        </main>
      </div>
    </div>
  );
}

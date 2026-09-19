import React, { useState, useEffect, useRef } from 'react';
import { Outlet, useLocation } from 'react-router-dom';
import { Sparkles } from 'lucide-react';
import Sidebar from './Sidebar';
import TopBar from './TopBar';
import AICopilotDrawer from '../../components/Copilot/AICopilotDrawer';
import './AppShell.css';

export default function AppShell() {
  const [isCopilotOpen, setIsCopilotOpen] = useState(false);
  const [isSidebarCollapsed, setIsSidebarCollapsed] = useState(() => {
    try {
      return localStorage.getItem('freel_sidebar_collapsed') === 'true';
    } catch {
      return false;
    }
  });

  const location = useLocation();
  const mainRef = useRef(null);

  const handleToggleSidebar = () => {
    setIsSidebarCollapsed((prev) => {
      const next = !prev;
      try {
        localStorage.setItem('freel_sidebar_collapsed', String(next));
      } catch {
        // Ignore storage errors
      }
      return next;
    });
  };

  // Scroll workspace to top on route navigation
  useEffect(() => {
    if (mainRef.current) {
      mainRef.current.scrollTo({ top: 0, behavior: 'instant' });
    }
  }, [location.pathname]);

  // Global hotkeys
  useEffect(() => {
    const handleOpenCopilot = () => setIsCopilotOpen(true);
    const handleToggleCopilot = () => setIsCopilotOpen((prev) => !prev);
    const handleKeyDown = (e) => {
      // Shortcut: Alt + C or Ctrl + / to open Copilot
      if ((e.altKey && e.key.toLowerCase() === 'c') || (e.ctrlKey && e.key === '/')) {
        e.preventDefault();
        setIsCopilotOpen((prev) => !prev);
      }
      // Shortcut: Ctrl + [ or Cmd + [ to toggle Sidebar
      if ((e.ctrlKey || e.metaKey) && e.key === '[') {
        e.preventDefault();
        handleToggleSidebar();
      }
    };

    window.addEventListener('open-copilot', handleOpenCopilot);
    window.addEventListener('toggle-copilot', handleToggleCopilot);
    window.addEventListener('keydown', handleKeyDown);

    return () => {
      window.removeEventListener('open-copilot', handleOpenCopilot);
      window.removeEventListener('toggle-copilot', handleToggleCopilot);
      window.removeEventListener('keydown', handleKeyDown);
    };
  }, []);

  return (
    <div className={`app-shell ${isSidebarCollapsed ? 'sidebar-collapsed' : ''}`}>
      <Sidebar
        isCollapsed={isSidebarCollapsed}
        onToggleCollapse={handleToggleSidebar}
      />
      <div className="app-shell-content">
        <TopBar
          isSidebarCollapsed={isSidebarCollapsed}
          onToggleSidebar={handleToggleSidebar}
        />
        <main className="app-shell-main" ref={mainRef} id="main-content-scroll">
          <div className="app-shell-container">
            <Outlet />
          </div>
        </main>
      </div>

      {/* Floating Copilot Launcher Button */}
      {!isCopilotOpen && (
        <button
          type="button"
          className="copilot-floating-launcher"
          onClick={() => setIsCopilotOpen(true)}
          title="Open LogisticsHQ AI Copilot (Alt+C)"
          aria-label="Ask Copilot"
        >
          <Sparkles size={16} className="copilot-floating-sparkle" />
          <span>Ask Copilot</span>
        </button>
      )}

      {/* AI Copilot Drawer */}
      <AICopilotDrawer
        isOpen={isCopilotOpen}
        onClose={() => setIsCopilotOpen(false)}
      />
    </div>
  );
}

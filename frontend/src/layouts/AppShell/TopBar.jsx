import { useState, useEffect } from 'react';
import { useAuth } from '../../context/AuthContext';
import './TopBar.css';

export default function TopBar() {
  const { org } = useAuth();
  const [isCommandPaletteOpen, setCommandPaletteOpen] = useState(false);

  // Handle Cmd+K / Ctrl+K to open Command Palette
  useEffect(() => {
    const handleKeyDown = (e) => {
      if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
        e.preventDefault();
        setCommandPaletteOpen(true);
      }
    };
    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, []);

  return (
    <>
      <header className="app-topbar">
        <div className="topbar-search" onClick={() => setCommandPaletteOpen(true)}>
          <span className="search-icon">🔍</span>
          <span className="search-placeholder">Search shipments, leads, RFQs...</span>
          <div className="search-shortcut">
            <kbd>⌘</kbd><kbd>K</kbd>
          </div>
        </div>

        <div className="topbar-actions">
          {/* Notifications */}
          <button className="topbar-btn notifications-btn" aria-label="Notifications">
            <span className="bell-icon">🔔</span>
            <span className="notification-badge"></span>
          </button>

          {/* Org Dropdown Stub */}
          <div className="org-switcher">
            <span className="org-name">{org?.name || 'Freel Corp'}</span>
            <span className="org-chevron">▼</span>
          </div>
        </div>
      </header>

      {/* Command Palette Modal Stub */}
      {isCommandPaletteOpen && (
        <div className="command-palette-overlay" onClick={() => setCommandPaletteOpen(false)}>
          <div className="command-palette-modal" onClick={e => e.stopPropagation()}>
            <div className="cmd-header">
              <input 
                type="text" 
                placeholder="Type a command or search..." 
                autoFocus 
                className="cmd-input"
              />
              <span className="cmd-esc">ESC</span>
            </div>
            <div className="cmd-body">
              <div className="cmd-empty-state">
                <p>Search results will appear here.</p>
                <span className="text-sm text-slate-400">Try searching for a shipment ID or Lead name.</span>
              </div>
            </div>
          </div>
        </div>
      )}
    </>
  );
}

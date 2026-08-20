import React, { useState, useEffect } from 'react';
import { useAuth } from '../../context/AuthContext';
import { Search, Bell, Globe, Maximize2, Calendar, ChevronDown, X } from 'lucide-react';
import './TopBar.css';

export default function TopBar() {
  const { user } = useAuth();
  const [isCommandPaletteOpen, setCommandPaletteOpen] = useState(false);
  const [dateRange, setDateRange] = useState('Aug 9 - Aug 15, 2026');

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

  const firstName = user?.first_name || (user?.full_name && !user.full_name.includes('@') ? user.full_name.split(' ')[0] : null) || (user?.name && !user.name.includes('@') ? user.name.split(' ')[0] : null) || 'Varun';

  return (
    <>
      <header className="app-topbar">
        {/* ── Left Welcome / Breadcrumb Header ── */}
        <div className="topbar-welcome">
          <h1 className="topbar-title">
            Welcome to LogisticsHQ, {firstName}! <span className="wave-emoji">👋</span>
          </h1>
          <p className="topbar-subtitle">Your freight workspace is ready. Let's get your first operation moving.</p>
        </div>

        {/* ── Right Actions ── */}
        <div className="topbar-right">
          {/* Search bar with ⌘K shortcut */}
          <div className="topbar-search" onClick={() => setCommandPaletteOpen(true)}>
            <Search size={15} className="search-icon-svg" />
            <span className="search-placeholder">Search shipments, RFQs, customers...</span>
            <div className="search-shortcut">
              <kbd>⌘</kbd><kbd>K</kbd>
            </div>
          </div>

          {/* Action Icons */}
          <div className="topbar-icons-group">
            <button className="topbar-icon-btn" aria-label="Notifications" title="Notifications">
              <Bell size={17} />
              <span className="icon-badge">1</span>
            </button>

            <button className="topbar-icon-btn" aria-label="Language" title="Language">
              <Globe size={17} />
            </button>

            <button className="topbar-icon-btn" aria-label="Fullscreen" title="Fullscreen">
              <Maximize2 size={17} />
            </button>
          </div>

          {/* Date Range Picker */}
          <div className="topbar-date-picker">
            <Calendar size={15} className="date-icon-svg" />
            <span className="date-text">{dateRange}</span>
            <ChevronDown size={14} className="date-chevron-svg" />
          </div>
        </div>
      </header>

      {/* Command Palette Modal */}
      {isCommandPaletteOpen && (
        <div className="command-palette-overlay" onClick={() => setCommandPaletteOpen(false)}>
          <div className="command-palette-modal" onClick={(e) => e.stopPropagation()}>
            <div className="cmd-header">
              <Search size={18} className="cmd-search-icon-svg" />
              <input
                type="text"
                placeholder="Search shipments, RFQs, leads, customers..."
                autoFocus
                className="cmd-input"
              />
              <button className="cmd-esc-btn" onClick={() => setCommandPaletteOpen(false)}>
                <X size={14} /> ESC
              </button>
            </div>
            <div className="cmd-body">
              <div className="cmd-empty-state">
                <p>Type to search across your workspace.</p>
                <span className="cmd-hint">Try searching for an RFQ ID, customer company, or container number.</span>
              </div>
            </div>
          </div>
        </div>
      )}
    </>
  );
}

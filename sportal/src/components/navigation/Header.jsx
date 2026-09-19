import React, { useState, useRef, useEffect } from 'react';
import { Search, Bell, LogOut, ShieldCheck, ChevronDown, UserCheck, CheckCircle2, Activity, Check, ExternalLink } from 'lucide-react';
import { useAuth } from '../../contexts/AuthContext';
import { Link } from 'react-router-dom';
import { sportalService } from '../../services/sportalService';

export function Header() {
  const { user, org, role, logout } = useAuth();
  const [dropdownOpen, setDropdownOpen] = useState(false);
  const [notifOpen, setNotifOpen] = useState(false);
  const [notifications, setNotifications] = useState([]);
  const [unreadCount, setUnreadCount] = useState(0);
  const [loadingNotifs, setLoadingNotifs] = useState(false);
  const dropdownRef = useRef(null);
  const notifRef = useRef(null);

  const fetchQuickNotifs = async () => {
    try {
      setLoadingNotifs(true);
      const res = await sportalService.getNotifications({ limit: 5 });
      setNotifications(res?.items || []);
      setUnreadCount(res?.unread_count || 0);
    } catch (err) {
      console.error('Failed to fetch quick notifications:', err);
    } finally {
      setLoadingNotifs(false);
    }
  };

  useEffect(() => {
    fetchQuickNotifs();
  }, []);

  // Close dropdowns on outside click
  useEffect(() => {
    function handleClickOutside(event) {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target)) {
        setDropdownOpen(false);
      }
      if (notifRef.current && !notifRef.current.contains(event.target)) {
        setNotifOpen(false);
      }
    }
    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  const handleMarkQuickRead = async (id, e) => {
    e.stopPropagation();
    try {
      await sportalService.markNotificationRead(id);
      setNotifications((prev) =>
        prev.map((n) => (n.id === id ? { ...n, is_read: true } : n))
      );
      setUnreadCount((prev) => Math.max(0, prev - 1));
    } catch (err) {
      console.error('Failed to mark read:', err);
    }
  };

  const displayName = user?.full_name || (user?.email ? user.email.split('@')[0] : 'Administrator');
  const roleLabel = role?.display_name || role?.name || 'Platform Super Admin';
  const initial = displayName ? displayName.charAt(0).toUpperCase() : 'A';

  return (
    <header className="h-16 bg-white border-b border-slate-200 px-6 sm:px-8 flex items-center justify-between sticky top-0 z-20 shadow-2xs">
      {/* Title / Portal Identification */}
      <div className="flex items-center space-x-3">
        <div className="flex flex-col">
          <span className="text-base font-bold text-slate-900 leading-none">SPortal</span>
          <span className="text-[11px] text-slate-400 mt-1">Internal SaaS Portal</span>
        </div>
      </div>

      {/* Global Search Bar */}
      <div className="flex-1 max-w-lg mx-6 hidden md:block">
        <div className="relative">
          <Search className="w-4 h-4 text-slate-400 absolute left-3 top-1/2 -translate-y-1/2" />
          <input
            id="sportal-global-search"
            type="text"
            placeholder="Search organizations, users, subscriptions, RFQs, shipments..."
            className="w-full pl-9 pr-16 py-1.5 text-xs bg-slate-50 border border-slate-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-600 focus:bg-white transition-all text-slate-800 placeholder-slate-400"
          />
          <div className="absolute right-2.5 top-1/2 -translate-y-1/2 flex items-center">
            <kbd className="px-1.5 py-0.5 text-[10px] font-medium text-slate-400 bg-white border border-slate-200 rounded">
              Ctrl + K
            </kbd>
          </div>
        </div>
      </div>

      {/* User Actions & Profile matching sportalDashboard.png */}
      <div className="flex items-center space-x-5">
        {/* Notification Bell with red count badge */}
        <div className="relative" ref={notifRef}>
          <button
            id="sportal-notification-bell-btn"
            type="button"
            onClick={() => {
              setNotifOpen(!notifOpen);
              if (!notifOpen) fetchQuickNotifs();
            }}
            className="relative p-2 text-slate-600 hover:text-slate-900 hover:bg-slate-100 rounded-lg transition-colors cursor-pointer"
            aria-label="Notifications"
          >
            <Bell className="w-5 h-5" />
            <span className="absolute -top-0.5 -right-0.5 flex items-center justify-center min-w-4 h-4 px-1 rounded-full bg-rose-500 text-white text-[10px] font-bold ring-2 ring-white">
              {unreadCount > 0 ? unreadCount : '3'}
            </span>
          </button>

          {/* Notifications Dropdown Drawer */}
          {notifOpen && (
            <div
              id="sportal-notification-popover"
              className="absolute right-0 mt-2.5 w-80 sm:w-96 rounded-xl border border-slate-200 bg-white shadow-xl shadow-slate-900/10 z-50 animate-fade-in overflow-hidden"
            >
              <div className="p-3 bg-slate-50 border-b border-slate-100 flex items-center justify-between">
                <div className="flex items-center gap-2">
                  <span className="text-xs font-bold text-slate-900">Notifications</span>
                  {unreadCount > 0 && (
                    <span className="px-1.5 py-0.2 text-[10px] font-bold rounded-full bg-blue-100 text-blue-800">
                      {unreadCount} unread
                    </span>
                  )}
                </div>
                <Link
                  to="/support?tab=notifications"
                  onClick={() => setNotifOpen(false)}
                  className="text-[11px] font-semibold text-blue-600 hover:underline flex items-center gap-1"
                >
                  <span>View All</span>
                  <ExternalLink className="w-3 h-3" />
                </Link>
              </div>

              <div className="max-h-80 overflow-y-auto divide-y divide-slate-100 text-xs">
                {loadingNotifs ? (
                  <div className="p-4 text-center text-slate-400 text-xs">Loading alerts...</div>
                ) : notifications.length === 0 ? (
                  <div className="p-6 text-center text-slate-400">No notifications found</div>
                ) : (
                  notifications.map((n) => (
                    <div
                      key={n.id}
                      className={`p-3 transition-colors hover:bg-slate-50 flex items-start justify-between gap-2 ${
                        !n.is_read ? 'bg-blue-50/40' : ''
                      }`}
                    >
                      <div className="space-y-0.5">
                        <div className="flex items-center gap-1.5">
                          <span className={`w-1.5 h-1.5 rounded-full ${n.is_read ? 'bg-slate-300' : 'bg-blue-600'}`} />
                          <span className="font-semibold text-slate-900 line-clamp-1">{n.title}</span>
                        </div>
                        <p className="text-slate-600 text-[11px] line-clamp-2">{n.message}</p>
                        <div className="flex items-center gap-2 text-[10px] text-slate-400 mt-1">
                          <span>{n.organization_name}</span>
                          <span>•</span>
                          <span>{new Date(n.created_at).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}</span>
                        </div>
                      </div>
                      {!n.is_read && (
                        <button
                          type="button"
                          onClick={(e) => handleMarkQuickRead(n.id, e)}
                          title="Mark as read"
                          className="p-1 rounded hover:bg-slate-200 text-slate-400 hover:text-blue-600 transition-colors cursor-pointer shrink-0"
                        >
                          <Check className="w-3.5 h-3.5" />
                        </button>
                      )}
                    </div>
                  ))
                )}
              </div>
            </div>
          )}
        </div>

        {/* User Info & Dropdown */}
        <div className="relative" ref={dropdownRef}>
          <button
            id="sportal-user-profile-menu-button"
            type="button"
            onClick={() => setDropdownOpen(!dropdownOpen)}
            className="flex items-center space-x-2.5 pl-3 border-l border-slate-200 cursor-pointer text-left hover:opacity-90 transition-opacity"
          >
            <div className="w-8 h-8 rounded-lg bg-blue-600 text-white font-bold flex items-center justify-center text-xs shadow-xs">
              {initial}
            </div>
            <div className="flex flex-col">
              <span className="text-xs font-semibold text-slate-900 leading-tight flex items-center gap-1">
                {displayName}
                <ChevronDown className="w-3 h-3 text-slate-400" />
              </span>
              <span className="text-[10px] text-slate-500 font-medium">{roleLabel}</span>
            </div>
          </button>

          {/* Interactive Dropdown */}
          {dropdownOpen && (
            <div className="absolute right-0 mt-2.5 w-64 rounded-xl border border-slate-200 bg-white p-2.5 shadow-lg shadow-slate-900/10 z-50 animate-fade-in">
              {/* Profile Card Header */}
              <div className="p-2.5 bg-slate-50 rounded-lg border border-slate-100 mb-2">
                <div className="flex items-center gap-1.5 text-[11px] font-semibold text-blue-700 mb-1">
                  <UserCheck className="w-3.5 h-3.5" />
                  Authenticated Staff
                </div>
                <div className="font-semibold text-xs text-slate-900 truncate">{displayName}</div>
                <div className="text-[11px] text-slate-500 truncate">{user?.email || 'internal@logisticshq.in'}</div>

                <div className="mt-2 pt-2 border-t border-slate-200/60 flex items-center justify-between text-[11px]">
                  <span className="text-slate-500">Org:</span>
                  <span className="font-medium text-slate-800 truncate max-w-[130px]">
                    {org?.name || 'LogisticsHQ HQ'}
                  </span>
                </div>
              </div>

              {/* Actions */}
              <button
                id="sportal-logout-button"
                type="button"
                onClick={() => {
                  setDropdownOpen(false);
                  logout();
                }}
                className="w-full flex items-center gap-2 px-2.5 py-1.5 text-xs font-medium text-rose-600 hover:bg-rose-50 rounded-md transition-colors cursor-pointer"
              >
                <LogOut className="w-3.5 h-3.5" />
                Sign Out of SPortal
              </button>
            </div>
          )}
        </div>
      </div>
    </header>
  );
}

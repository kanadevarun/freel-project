import React, { useState, useEffect, useRef } from 'react';
import { useNavigate, useLocation, useSearchParams } from 'react-router-dom';
import { useAuth } from '../../context/AuthContext';
import { dashboardService } from '../../services/dashboardService';
import { searchService } from '../../services/searchService';
import {
  Search,
  Bell,
  Globe,
  Maximize2,
  Minimize2,
  Calendar,
  ChevronDown,
  X,
  ChevronRight,
  Check,
  Ship,
  FileText,
  FileSpreadsheet,
  Users,
  CreditCard,
  Sparkles,
  FolderOpen,
  Radio,
  Layers,
  Building,
  User,
  Settings,
  LogOut,
  Sliders,
  Clock,
  ExternalLink,
  History,
  CornerDownLeft,
} from 'lucide-react';
import './TopBar.css';

const DATE_PRESETS = [
  { key: 'LAST_7D', label: 'Last 7 Days', compLabel: 'vs preceding 7 days' },
  { key: 'TODAY', label: 'Today', compLabel: 'vs yesterday' },
  { key: 'YESTERDAY', label: 'Yesterday', compLabel: 'vs previous day' },
  { key: 'LAST_30D', label: 'Last 30 Days', compLabel: 'vs preceding 30 days' },
  { key: 'THIS_MONTH', label: 'This Month', compLabel: 'vs last month' },
  { key: 'LAST_MONTH', label: 'Last Month', compLabel: 'vs prior month' },
  { key: 'THIS_QUARTER', label: 'This Quarter', compLabel: 'vs previous quarter' },
  { key: 'CUSTOM', label: 'Custom Range...', compLabel: 'vs preceding period' },
];

const SEARCH_CATEGORIES = [
  { key: 'ALL', label: 'All Results' },
  { key: 'SHIPMENT', label: 'Shipments' },
  { key: 'BOOKING', label: 'Bookings' },
  { key: 'RFQ', label: 'RFQs' },
  { key: 'QUOTATION', label: 'Quotes' },
  { key: 'CUSTOMER', label: 'Customers' },
  { key: 'INVOICE', label: 'Invoices' },
  { key: 'LEAD', label: 'Leads' },
  { key: 'CONTRACT', label: 'Contracts' },
  { key: 'TRACKING', label: 'Tracking' },
];

const TIMEZONES = [
  { key: 'UTC', label: 'UTC (Universal Coordinated Time)', offset: 'UTC +00:00' },
  { key: 'America/New_York', label: 'America/New York (EST/EDT)', offset: 'UTC -05:00' },
  { key: 'Europe/London', label: 'Europe/London (GMT/BST)', offset: 'UTC +01:00' },
  { key: 'Asia/Kolkata', label: 'Asia/Kolkata (IST)', offset: 'UTC +05:30' },
  { key: 'Asia/Singapore', label: 'Asia/Singapore (SGT)', offset: 'UTC +08:00' },
  { key: 'Asia/Dubai', label: 'Asia/Dubai (GST)', offset: 'UTC +04:00' },
];

const CURRENCIES = [
  { code: 'USD', symbol: '$', name: 'US Dollar', tag: 'Global Reserve', color: '#2563eb', bg: '#eff6ff' },
  { code: 'EUR', symbol: '€', name: 'Euro', tag: 'European Union', color: '#4f46e5', bg: '#eef2ff' },
  { code: 'GBP', symbol: '£', name: 'British Pound', tag: 'United Kingdom', color: '#059669', bg: '#ecfdf5' },
  { code: 'INR', symbol: '₹', name: 'Indian Rupee', tag: 'India', color: '#d97706', bg: '#fffbeb' },
  { code: 'AED', symbol: 'د.إ', name: 'UAE Dirham', tag: 'Middle East', color: '#7c3aed', bg: '#f5f3ff' },
  { code: 'SGD', symbol: 'S$', name: 'Singapore Dollar', tag: 'Asia-Pacific Hub', color: '#0891b2', bg: '#ecfeff' },
];

export default function TopBar() {
  const navigate = useNavigate();
  const location = useLocation();
  const [searchParams, setSearchParams] = useSearchParams();
  const { user, logout } = useAuth();

  // ── Unified Popover State (Mutual Exclusion) ──
  // 'search' | 'notif' | 'globe' | 'datePicker' | 'profile' | null
  const [activeMenu, setActiveMenu] = useState(null);

  // Notifications State
  const [attentionItems, setAttentionItems] = useState([]);

  // Fullscreen State
  const [isFullscreen, setIsFullscreen] = useState(false);

  // Search State
  const [searchQuery, setSearchQuery] = useState('');
  const [searchCategory, setSearchCategory] = useState('ALL');
  const [searchResults, setSearchResults] = useState(null);
  const [isSearching, setIsSearching] = useState(false);
  const [searchError, setSearchError] = useState(null);
  const [selectedIndex, setSelectedIndex] = useState(0);
  const [recentSearches, setRecentSearches] = useState(() => {
    try {
      const saved = localStorage.getItem('freel_recent_searches');
      return saved ? JSON.parse(saved) : ['BK-QA', 'MSC', 'Customer', 'INV-2026'];
    } catch {
      return ['BK-QA', 'MSC', 'Customer'];
    }
  });

  // Date Filter State
  const currentPresetParam = searchParams.get('preset') || 'LAST_7D';
  const currentStartDateParam = searchParams.get('startDate') || '';
  const currentEndDateParam = searchParams.get('endDate') || '';

  const [selectedPreset, setSelectedPreset] = useState(currentPresetParam);
  const [customStartDate, setCustomStartDate] = useState(currentStartDateParam || new Date(Date.now() - 6 * 86400000).toISOString().split('T')[0]);
  const [customEndDate, setCustomEndDate] = useState(currentEndDateParam || new Date().toISOString().split('T')[0]);
  const [dateRangeLabel, setDateRangeLabel] = useState('Aug 9 – Aug 15, 2026');

  // Globe / Localization Preferences
  const [selectedLanguage, setSelectedLanguage] = useState(() => localStorage.getItem('freel_language') || 'en-US');
  const [selectedCurrency, setSelectedCurrency] = useState(() => localStorage.getItem('freel_preferred_currency') || 'USD');
  const [selectedTimezone, setSelectedTimezone] = useState(() => localStorage.getItem('freel_preferred_timezone') || 'UTC');
  const [globeSavedToast, setGlobeSavedToast] = useState(false);
  const [openLocaleDropdown, setOpenLocaleDropdown] = useState(null); // 'currency' | 'timezone' | null

  // Refs for closing popovers on click outside
  const searchInputRef = useRef(null);
  const popoverContainerRef = useRef(null);

  // 1. Fetch Attention Items / Notifications
  useEffect(() => {
    let isMounted = true;
    const loadNotifications = async () => {
      try {
        const res = await dashboardService.getMissionControl({
          preset: currentPresetParam,
          startDate: currentStartDateParam,
          endDate: currentEndDateParam,
        });
        const missionData = res?.data || res || {};
        if (isMounted) {
          if (missionData.attention_items) {
            setAttentionItems(missionData.attention_items);
          }
          if (missionData.date_range?.label) {
            setDateRangeLabel(missionData.date_range.label);
          }
        }
      } catch (e) {
        // Silently tolerate
      }
    };
    loadNotifications();
    return () => {
      isMounted = false;
    };
  }, [currentPresetParam, currentStartDateParam, currentEndDateParam]);

  // 2. Fullscreen Listener
  useEffect(() => {
    const handleFullscreenChange = () => {
      setIsFullscreen(Boolean(document.fullscreenElement));
    };
    document.addEventListener('fullscreenchange', handleFullscreenChange);
    return () => document.removeEventListener('fullscreenchange', handleFullscreenChange);
  }, []);

  const toggleFullscreen = async () => {
    try {
      if (!document.fullscreenElement) {
        await document.documentElement.requestFullscreen();
      } else {
        await document.exitFullscreen();
      }
    } catch (e) {
      console.warn('Fullscreen request rejected by browser:', e);
    }
  };

  // 3. Global Shortcut Listeners (⌘K / Ctrl+K and Escape)
  useEffect(() => {
    const handleKeyDown = (e) => {
      if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
        e.preventDefault();
        setActiveMenu((prev) => (prev === 'search' ? null : 'search'));
      } else if (e.key === 'Escape') {
        setActiveMenu(null);
      }
    };
    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, []);

  // 4. Focus search input when command palette opens
  useEffect(() => {
    if (activeMenu === 'search') {
      setTimeout(() => {
        searchInputRef.current?.focus();
      }, 50);
    }
  }, [activeMenu]);

  // 5. Click outside listener to close menus
  useEffect(() => {
    const handleClickOutside = (e) => {
      if (activeMenu === 'search') return; // Handled by overlay backdrop
      if (popoverContainerRef.current && !popoverContainerRef.current.contains(e.target)) {
        setActiveMenu(null);
      }
    };
    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, [activeMenu]);

  // 6. Debounced Global Search Execution
  useEffect(() => {
    if (activeMenu !== 'search') return;

    if (!searchQuery.trim()) {
      setSearchResults(null);
      setIsSearching(false);
      setSearchError(null);
      return;
    }

    setIsSearching(true);
    setSearchError(null);

    const timer = setTimeout(async () => {
      try {
        const res = await searchService.globalSearch(searchQuery, searchCategory, 30);
        const searchData = res?.data?.data || res?.data || res || {};
        setSearchResults(searchData);
        setSelectedIndex(0);
      } catch (err) {
        setSearchError('Unable to search at this moment. Please try again.');
      } finally {
        setIsSearching(false);
      }
    }, 280);

    return () => clearTimeout(timer);
  }, [searchQuery, searchCategory, activeMenu]);

  // Helper to flat-list all search items for keyboard arrow navigation
  const flatSearchItems = React.useMemo(() => {
    if (!searchResults?.groups) return [];
    const items = [];
    searchResults.groups.forEach((group) => {
      group.items.forEach((item) => {
        items.push(item);
      });
    });
    return items;
  }, [searchResults]);

  // Handle Search Keyboard Navigation (ArrowUp / ArrowDown / Enter)
  const handleSearchKeyDown = (e) => {
    if (flatSearchItems.length === 0) return;

    if (e.key === 'ArrowDown') {
      e.preventDefault();
      setSelectedIndex((prev) => (prev + 1 < flatSearchItems.length ? prev + 1 : 0));
    } else if (e.key === 'ArrowUp') {
      e.preventDefault();
      setSelectedIndex((prev) => (prev - 1 >= 0 ? prev - 1 : flatSearchItems.length - 1));
    } else if (e.key === 'Enter') {
      e.preventDefault();
      const targetItem = flatSearchItems[selectedIndex];
      if (targetItem?.url) {
        handleNavigateToResult(targetItem);
      }
    }
  };

  const handleNavigateToResult = (item) => {
    // Save to recent searches
    if (searchQuery.trim()) {
      const updated = [searchQuery.trim(), ...recentSearches.filter((s) => s.toLowerCase() !== searchQuery.trim().toLowerCase())].slice(0, 6);
      setRecentSearches(updated);
      try {
        localStorage.setItem('freel_recent_searches', JSON.stringify(updated));
      } catch {}
    }
    setActiveMenu(null);
    setSearchQuery('');
    navigate(item.url);
  };

  // Date Filter Selection
  const handleSelectPreset = (presetKey) => {
    setSelectedPreset(presetKey);
    if (presetKey === 'CUSTOM') return; // Keep picker open for custom range selection

    const newParams = new URLSearchParams(searchParams);
    if (presetKey === 'LAST_7D') {
      newParams.delete('preset');
      newParams.delete('startDate');
      newParams.delete('endDate');
    } else {
      newParams.set('preset', presetKey);
      newParams.delete('startDate');
      newParams.delete('endDate');
    }
    setSearchParams(newParams);
    setActiveMenu(null);
  };

  const handleApplyCustomDateRange = (e) => {
    e.preventDefault();
    if (!customStartDate || !customEndDate) return;
    if (new Date(customStartDate) > new Date(customEndDate)) {
      alert('Start date must be before or equal to End date.');
      return;
    }

    const newParams = new URLSearchParams(searchParams);
    newParams.set('preset', 'CUSTOM');
    newParams.set('startDate', customStartDate);
    newParams.set('endDate', customEndDate);
    setSearchParams(newParams);
    setActiveMenu(null);
  };

  // Locale Preference Saver
  const handleSaveLocalePreference = (type, val) => {
    if (type === 'currency') {
      setSelectedCurrency(val);
      localStorage.setItem('freel_preferred_currency', val);
    } else if (type === 'timezone') {
      setSelectedTimezone(val);
      localStorage.setItem('freel_preferred_timezone', val);
    } else if (type === 'language') {
      setSelectedLanguage(val);
      localStorage.setItem('freel_language', val);
    }
    setGlobeSavedToast(true);
    setTimeout(() => setGlobeSavedToast(false), 2000);
  };

  const handleLogout = async () => {
    await logout();
    window.location.href = '/login';
  };

  // Dynamic user display values
  const firstName =
    user?.first_name ||
    (user?.full_name && !user.full_name.includes('@') ? user.full_name.split(' ')[0] : null) ||
    (user?.name && !user.name.includes('@') ? user.name.split(' ')[0] : null) ||
    'Varun';

  const userInitials = (
    (user?.first_name ? user.first_name[0] : '') +
    (user?.last_name ? user.last_name[0] : (firstName ? firstName[0] : 'U'))
  ).toUpperCase() || 'VK';

  const userEmail = user?.email || 'operator@logisticshq.io';
  const userRole = user?.role || user?.role_name || 'Org Admin';
  const orgName = user?.org_name || user?.company_name || 'LogisticsHQ Enterprise';

  const notifCount = attentionItems.length;

  const getCategoryIcon = (category) => {
    switch (category) {
      case 'SHIPMENT':
        return <Ship size={14} className="text-blue-500" />;
      case 'BOOKING':
        return <Layers size={14} className="text-indigo-500" />;
      case 'RFQ':
        return <FileText size={14} className="text-purple-500" />;
      case 'QUOTATION':
        return <FileSpreadsheet size={14} className="text-amber-500" />;
      case 'CUSTOMER':
        return <Users size={14} className="text-emerald-500" />;
      case 'INVOICE':
        return <CreditCard size={14} className="text-rose-500" />;
      case 'LEAD':
        return <Sparkles size={14} className="text-teal-500" />;
      case 'CONTRACT':
        return <FolderOpen size={14} className="text-cyan-500" />;
      case 'TRACKING':
        return <Radio size={14} className="text-orange-500" />;
      default:
        return <FileText size={14} className="text-slate-500" />;
    }
  };

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
        <div className="topbar-right" ref={popoverContainerRef}>
          {/* Global Search Bar trigger */}
          <div
            className="topbar-search"
            id="global-search-trigger"
            onClick={() => setActiveMenu('search')}
            role="button"
            tabIndex={0}
            aria-label="Open Global Search (⌘K)"
          >
            <Search size={15} className="search-icon-svg" />
            <span className="search-placeholder">Search shipments, RFQs, bookings, invoices...</span>
            <div className="search-shortcut">
              <kbd>{navigator.platform?.toUpperCase().indexOf('MAC') >= 0 ? '⌘' : 'Ctrl'}</kbd>
              <kbd>K</kbd>
            </div>
          </div>

          {/* Action Icons Group */}
          <div className="topbar-icons-group">
            {/* 1. Notifications Bell */}
            <div className="topbar-notif-container">
              <button
                className={`topbar-icon-btn ${activeMenu === 'notif' ? 'active' : ''}`}
                aria-label="Notifications"
                title="Notifications & Priority Center"
                onClick={() => setActiveMenu((prev) => (prev === 'notif' ? null : 'notif'))}
              >
                <Bell size={18} className="topbar-bell-icon" />
                {notifCount > 0 && (
                  <span className="topbar-notif-badge">
                    <span className="topbar-notif-badge-pulse" />
                    <span className="topbar-notif-badge-num">{notifCount}</span>
                  </span>
                )}
              </button>

              {activeMenu === 'notif' && (
                <div className="topbar-notif-popover animate-fade-in-down">
                  <div className="topbar-notif-header">
                    <div className="topbar-notif-title-wrap">
                      <Bell size={15} className="text-blue-600" />
                      <span className="topbar-notif-title">Notifications & Priority Center</span>
                    </div>
                    {notifCount > 0 ? (
                      <span className="topbar-notif-actionable-pill">{notifCount} Actionable</span>
                    ) : (
                      <span className="topbar-notif-caughtup-pill">All Caught Up</span>
                    )}
                  </div>

                  <div className="topbar-notif-list">
                    {attentionItems.length > 0 ? (
                      attentionItems.map((item) => {
                        const isHigh = item.priority === 'HIGH';
                        const isMed = item.priority === 'MEDIUM';
                        const iconBg = item.category === 'FINANCE' ? '#fef2f2' : item.category === 'APPROVALS' ? '#f5f3ff' : item.category === 'OPERATIONS' ? '#eff6ff' : '#fffbeb';
                        const iconColor = item.category === 'FINANCE' ? '#dc2626' : item.category === 'APPROVALS' ? '#7c3aed' : item.category === 'OPERATIONS' ? '#2563eb' : '#d97706';

                        return (
                          <div
                            key={item.id}
                            className={`topbar-notif-item ${isHigh ? 'priority-high' : isMed ? 'priority-med' : 'priority-info'}`}
                            onClick={() => {
                              setActiveMenu(null);
                              navigate(item.action_url || '/dashboard');
                            }}
                          >
                            <div className="topbar-notif-item-icon-box" style={{ background: iconBg, color: iconColor }}>
                              <Bell size={13} />
                            </div>
                            <div className="topbar-notif-item-body">
                              <div className="topbar-notif-item-top">
                                <span className="topbar-notif-item-title">{item.title}</span>
                                <span className="topbar-notif-item-time">{item.timestamp}</span>
                              </div>
                              <div className="topbar-notif-item-subtitle">{item.subtitle}</div>
                            </div>
                            <ChevronRight size={13} className="topbar-notif-item-arrow" />
                          </div>
                        );
                      })
                    ) : (
                      <div className="topbar-notif-empty">
                        <div className="topbar-notif-empty-icon-box">
                          <Bell size={22} />
                        </div>
                        <h4>You're all caught up!</h4>
                        <p>No actionable items require immediate attention right now.</p>
                      </div>
                    )}
                  </div>

                  <div className="topbar-notif-footer">
                    <button
                      className="topbar-notif-footer-btn"
                      onClick={() => {
                        setActiveMenu(null);
                        navigate('/dashboard/approvals');
                      }}
                    >
                      <span>View All Approvals & Priority Queue</span>
                      <ChevronRight size={13} />
                    </button>
                  </div>
                </div>
              )}
            </div>

            {/* 2. Globe / Localization & Region */}
            <div className="topbar-globe-container">
              <button
                className={`topbar-icon-btn ${activeMenu === 'globe' ? 'active' : ''}`}
                aria-label="Language & Region Settings"
                title="Language & Regional Settings"
                onClick={() => setActiveMenu((prev) => (prev === 'globe' ? null : 'globe'))}
              >
                <Globe size={17} />
              </button>

              {activeMenu === 'globe' && (
                <div className="topbar-globe-popover animate-fade-in-down">
                  <div className="popover-panel-header">
                    <div className="popover-panel-title">
                      <Globe size={15} className="text-blue-600" />
                      <span>Language & Regional Preferences</span>
                    </div>
                    {globeSavedToast && <span className="save-indicator-badge">✓ Saved</span>}
                  </div>

                  <div className="globe-popover-content">
                    {/* Supported Language */}
                    <div className="pref-row">
                      <label className="pref-label">Interface Language</label>
                      <div className="pref-badge-row">
                        <span className="lang-pill active">
                          <Check size={12} /> English (United States)
                        </span>
                        <span className="lang-pill muted" title="Application default is English">
                          Metric (DCSA Standard)
                        </span>
                      </div>
                      <p className="pref-subtext">LogisticsHQ operates in English with universal DCSA shipping standards.</p>
                    </div>

                    {/* Display Currency */}
                    <div className="pref-row">
                      <label className="pref-label">Display Currency</label>
                      <div className="custom-locale-dropdown-container">
                        {(() => {
                          const currentCurrencyObj = CURRENCIES.find((c) => c.code === selectedCurrency) || CURRENCIES[0];
                          return (
                            <>
                              <button
                                type="button"
                                className={`custom-locale-trigger ${openLocaleDropdown === 'currency' ? 'open' : ''}`}
                                onClick={() => setOpenLocaleDropdown((prev) => (prev === 'currency' ? null : 'currency'))}
                              >
                                <div className="locale-trigger-left">
                                  <div
                                    className="currency-badge-icon"
                                    style={{
                                      background: currentCurrencyObj.bg,
                                      color: currentCurrencyObj.color,
                                    }}
                                  >
                                    {currentCurrencyObj.symbol}
                                  </div>
                                  <div className="locale-trigger-text">
                                    <span className="locale-trigger-title">{currentCurrencyObj.name}</span>
                                    <span className="locale-trigger-sub">{currentCurrencyObj.code} • {currentCurrencyObj.tag}</span>
                                  </div>
                                </div>
                                <ChevronDown size={14} className={`locale-chevron ${openLocaleDropdown === 'currency' ? 'rotate' : ''}`} />
                              </button>

                              {openLocaleDropdown === 'currency' && (
                                <div className="custom-locale-menu animate-fade-in-down" role="listbox">
                                  {CURRENCIES.map((c) => {
                                    const isSelected = selectedCurrency === c.code;
                                    return (
                                      <button
                                        key={c.code}
                                        type="button"
                                        role="option"
                                        aria-selected={isSelected}
                                        className={`custom-locale-option ${isSelected ? 'selected' : ''}`}
                                        onClick={() => {
                                          handleSaveLocalePreference('currency', c.code);
                                          setOpenLocaleDropdown(null);
                                        }}
                                      >
                                        <div className="locale-option-left">
                                          <div
                                            className="currency-badge-icon"
                                            style={{
                                              background: c.bg,
                                              color: c.color,
                                            }}
                                          >
                                            {c.symbol}
                                          </div>
                                          <div className="locale-option-text">
                                            <span className="locale-option-title">{c.name}</span>
                                            <span className="locale-option-code">{c.code} • {c.tag}</span>
                                          </div>
                                        </div>
                                        {isSelected && <Check size={14} className="locale-check" />}
                                      </button>
                                    );
                                  })}
                                </div>
                              )}
                            </>
                          );
                        })()}
                      </div>
                    </div>

                    {/* Workspace Timezone */}
                    <div className="pref-row">
                      <label className="pref-label">Workspace Timezone</label>
                      <div className="custom-locale-dropdown-container">
                        {(() => {
                          const currentTimezoneObj = TIMEZONES.find((tz) => tz.key === selectedTimezone) || TIMEZONES[0];
                          return (
                            <>
                              <button
                                type="button"
                                className={`custom-locale-trigger ${openLocaleDropdown === 'timezone' ? 'open' : ''}`}
                                onClick={() => setOpenLocaleDropdown((prev) => (prev === 'timezone' ? null : 'timezone'))}
                              >
                                <div className="locale-trigger-left">
                                  <div className="timezone-badge-icon">
                                    <Clock size={13} />
                                  </div>
                                  <div className="locale-trigger-text">
                                    <span className="locale-trigger-title">{currentTimezoneObj.label.split('(')[0].trim()}</span>
                                    <span className="locale-trigger-sub">{currentTimezoneObj.offset}</span>
                                  </div>
                                </div>
                                <ChevronDown size={14} className={`locale-chevron ${openLocaleDropdown === 'timezone' ? 'rotate' : ''}`} />
                              </button>

                              {openLocaleDropdown === 'timezone' && (
                                <div className="custom-locale-menu animate-fade-in-down" role="listbox">
                                  {TIMEZONES.map((tz) => {
                                    const isSelected = selectedTimezone === tz.key;
                                    return (
                                      <button
                                        key={tz.key}
                                        type="button"
                                        role="option"
                                        aria-selected={isSelected}
                                        className={`custom-locale-option ${isSelected ? 'selected' : ''}`}
                                        onClick={() => {
                                          handleSaveLocalePreference('timezone', tz.key);
                                          setOpenLocaleDropdown(null);
                                        }}
                                      >
                                        <div className="locale-option-left">
                                          <div className="timezone-badge-icon">
                                            <Clock size={12} />
                                          </div>
                                          <div className="locale-option-text">
                                            <span className="locale-option-title">{tz.label.split('(')[0].trim()}</span>
                                            <span className="locale-option-code">{tz.offset}</span>
                                          </div>
                                        </div>
                                        {isSelected && <Check size={14} className="locale-check" />}
                                      </button>
                                    );
                                  })}
                                </div>
                              )}
                            </>
                          );
                        })()}
                      </div>
                    </div>
                  </div>

                  <div className="popover-panel-footer">
                    <button
                      className="panel-link-btn"
                      onClick={() => {
                        setActiveMenu(null);
                        navigate('/dashboard/settings/workspace');
                      }}
                    >
                      <span>Manage Workspace & Organization Settings</span>
                      <ChevronRight size={13} />
                    </button>
                  </div>
                </div>
              )}
            </div>

            {/* 3. Expand / Fullscreen Toggle */}
            <button
              className={`topbar-icon-btn ${isFullscreen ? 'active' : ''}`}
              aria-label={isFullscreen ? 'Exit Fullscreen' : 'Enter Fullscreen'}
              title={isFullscreen ? 'Exit Fullscreen (Esc)' : 'Expand Dashboard Fullscreen'}
              onClick={toggleFullscreen}
            >
              {isFullscreen ? <Minimize2 size={17} /> : <Maximize2 size={17} />}
            </button>
          </div>

          {/* 4. Date Range Picker Dropdown */}
          <div className="topbar-date-picker-container">
            <div
              className={`topbar-date-picker ${activeMenu === 'datePicker' ? 'active' : ''}`}
              onClick={() => setActiveMenu((prev) => (prev === 'datePicker' ? null : 'datePicker'))}
              role="button"
              tabIndex={0}
              aria-label="Select Dashboard Date Range"
            >
              <Calendar size={15} className="date-icon-svg" />
              <span className="date-text">{dateRangeLabel}</span>
              <ChevronDown size={14} className="date-chevron-svg" />
            </div>

            {activeMenu === 'datePicker' && (
              <div className="topbar-date-popover animate-fade-in-down">
                <div className="popover-panel-header">
                  <div className="popover-panel-title">
                    <Calendar size={15} className="text-blue-600" />
                    <span>Dashboard Time Range</span>
                  </div>
                </div>

                <div className="date-presets-list">
                  {DATE_PRESETS.map((p) => {
                    const isSelected = selectedPreset === p.key;
                    return (
                      <button
                        key={p.key}
                        className={`date-preset-item ${isSelected ? 'selected' : ''}`}
                        onClick={() => handleSelectPreset(p.key)}
                      >
                        <span className="preset-name">{p.label}</span>
                        <span className="preset-comp">{p.compLabel}</span>
                        {isSelected && <Check size={14} className="preset-check" />}
                      </button>
                    );
                  })}
                </div>

                {/* Custom Range Picker Drawer */}
                {selectedPreset === 'CUSTOM' && (
                  <form className="date-custom-drawer animate-fade-in" onSubmit={handleApplyCustomDateRange}>
                    <div className="custom-range-row">
                      <div className="custom-date-field">
                        <label>Start Date</label>
                        <input
                          type="date"
                          value={customStartDate}
                          onChange={(e) => setCustomStartDate(e.target.value)}
                          max={customEndDate || new Date().toISOString().split('T')[0]}
                          required
                        />
                      </div>
                      <span className="date-range-sep">➔</span>
                      <div className="custom-date-field">
                        <label>End Date</label>
                        <input
                          type="date"
                          value={customEndDate}
                          onChange={(e) => setCustomEndDate(e.target.value)}
                          min={customStartDate}
                          required
                        />
                      </div>
                    </div>
                    <div className="custom-date-actions">
                      <button
                        type="button"
                        className="btn-date-cancel"
                        onClick={() => handleSelectPreset('LAST_7D')}
                      >
                        Cancel
                      </button>
                      <button type="submit" className="btn-date-apply">
                        Apply Range
                      </button>
                    </div>
                  </form>
                )}
              </div>
            )}
          </div>

          {/* 5. User Profile Menu */}
          <div className="topbar-profile-container">
            <div
              className={`topbar-profile-btn ${activeMenu === 'profile' ? 'active' : ''}`}
              onClick={() => setActiveMenu((prev) => (prev === 'profile' ? null : 'profile'))}
              role="button"
              tabIndex={0}
              aria-label="User Account Menu"
            >
              <div className="user-avatar-initials" title={`${firstName} (${userEmail})`}>
                {userInitials}
              </div>
              <ChevronDown size={14} className="profile-chevron-svg" />
            </div>

            {activeMenu === 'profile' && (
              <div className="topbar-profile-dropdown animate-fade-in-down">
                {/* User Identity Header */}
                <div className="profile-dropdown-header">
                  <div className="profile-header-avatar">{userInitials}</div>
                  <div className="profile-header-meta">
                    <div className="profile-user-name">
                      {user?.first_name ? `${user.first_name} ${user.last_name || ''}` : firstName}
                    </div>
                    <div className="profile-user-email">{userEmail}</div>
                    <div className="profile-user-badge-row">
                      <span className="profile-role-pill">{userRole}</span>
                      <span className="profile-org-name">{orgName}</span>
                    </div>
                  </div>
                </div>

                {/* Profile Actions List */}
                <div className="profile-dropdown-body">
                  <button
                    className="profile-menu-item"
                    onClick={() => {
                      setActiveMenu(null);
                      navigate('/dashboard/settings/users');
                    }}
                  >
                    <User size={15} className="text-slate-500" />
                    <span>My Profile & Team</span>
                  </button>

                  <button
                    className="profile-menu-item"
                    onClick={() => {
                      setActiveMenu(null);
                      navigate('/dashboard/settings/company-profile');
                    }}
                  >
                    <Building size={15} className="text-slate-500" />
                    <span>Company Profile</span>
                  </button>

                  <button
                    className="profile-menu-item"
                    onClick={() => {
                      setActiveMenu(null);
                      navigate('/dashboard/settings/workspace');
                    }}
                  >
                    <Sliders size={15} className="text-slate-500" />
                    <span>Workspace Settings</span>
                  </button>

                  <button
                    className="profile-menu-item"
                    onClick={() => {
                      setActiveMenu(null);
                      navigate('/dashboard/settings/carrier-integrations');
                    }}
                  >
                    <Ship size={15} className="text-slate-500" />
                    <span>Carrier Integrations</span>
                  </button>
                </div>

                {/* Footer / Sign Out */}
                <div className="profile-dropdown-footer">
                  <button className="profile-signout-btn" onClick={handleLogout}>
                    <LogOut size={15} className="text-rose-500" />
                    <span>Sign out of LogisticsHQ</span>
                  </button>
                </div>
              </div>
            )}
          </div>
        </div>
      </header>

      {/* ── Global Search Command Palette Modal ── */}
      {activeMenu === 'search' && (
        <div
          className="command-palette-overlay animate-fade-in"
          onClick={(e) => {
            if (e.target === e.currentTarget) setActiveMenu(null);
          }}
        >
          <div className="command-palette-modal animate-scale-up" role="dialog" aria-modal="true">
            {/* Search Input Header */}
            <div className="cmd-header">
              <Search size={18} className="cmd-search-icon-svg" />
              <input
                ref={searchInputRef}
                type="text"
                className="cmd-input"
                placeholder="Search across shipments, bookings, RFQs, customers, invoices, tracking..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                onKeyDown={handleSearchKeyDown}
                autoFocus
              />
              {searchQuery && (
                <button className="cmd-clear-btn" onClick={() => setSearchQuery('')} aria-label="Clear search query">
                  <X size={15} />
                </button>
              )}
              <button className="cmd-esc-btn" onClick={() => setActiveMenu(null)} aria-label="Close modal">
                <kbd>Esc</kbd>
              </button>
            </div>

            {/* Category Filter Pills */}
            <div className="cmd-category-bar">
              {SEARCH_CATEGORIES.map((cat) => (
                <button
                  key={cat.key}
                  className={`cmd-cat-btn ${searchCategory === cat.key ? 'active' : ''}`}
                  onClick={() => setSearchCategory(cat.key)}
                >
                  {cat.label}
                </button>
              ))}
            </div>

            {/* Modal Body: Loading, Results, Recent Searches, or Empty State */}
            <div className="cmd-body-scroll">
              {isSearching ? (
                <div className="cmd-loading-state">
                  <div className="cmd-spinner" />
                  <span>Searching operational freight database...</span>
                </div>
              ) : searchError ? (
                <div className="cmd-error-state">
                  <p>{searchError}</p>
                </div>
              ) : searchResults && searchResults.groups && searchResults.groups.length > 0 ? (
                <div className="cmd-results-list">
                  {searchResults.groups.map((group) => (
                    <div key={group.category} className="cmd-group-section">
                      <div className="cmd-group-header">
                        <span className="cmd-group-title">{group.category_label}</span>
                        <span className="cmd-group-count">{group.count}</span>
                      </div>
                      <div className="cmd-group-items">
                        {group.items.map((item) => {
                          const itemGlobalIndex = flatSearchItems.findIndex((it) => it.id === item.id);
                          const isHighlighted = itemGlobalIndex === selectedIndex;

                          return (
                            <div
                              key={item.id}
                              className={`cmd-result-item ${isHighlighted ? 'active-highlight' : ''}`}
                              onClick={() => handleNavigateToResult(item)}
                              onMouseEnter={() => setSelectedIndex(itemGlobalIndex)}
                            >
                              <div className="cmd-item-icon-box">{getCategoryIcon(item.category)}</div>
                              <div className="cmd-item-info">
                                <div className="cmd-item-title-row">
                                  <span className="cmd-item-title">{item.title}</span>
                                  {item.badge && (
                                    <span className={`cmd-item-badge ${item.badge_type || 'neutral'}`}>
                                      {item.badge}
                                    </span>
                                  )}
                                </div>
                                <div className="cmd-item-subtitle">{item.subtitle}</div>
                              </div>
                              <div className="cmd-item-action">
                                <span className="cmd-jump-hint">Jump to</span>
                                <CornerDownLeft size={13} />
                              </div>
                            </div>
                          );
                        })}
                      </div>
                    </div>
                  ))}
                </div>
              ) : searchQuery.trim() ? (
                <div className="cmd-empty-state">
                  <p>No operational records found for "{searchQuery}"</p>
                  <span className="cmd-hint">
                    Try searching by shipment number (SH-), booking reference (BK-), customer company, container ID, or invoice number.
                  </span>
                </div>
              ) : (
                /* Default State: Recent Searches & Quick Navigation */
                <div className="cmd-default-view">
                  {recentSearches.length > 0 && (
                    <div className="cmd-recent-section">
                      <div className="cmd-recent-header">
                        <div className="cmd-recent-title">
                          <History size={13} />
                          <span>Recent Searches</span>
                        </div>
                        <button
                          className="cmd-recent-clear-all"
                          onClick={() => {
                            setRecentSearches([]);
                            localStorage.removeItem('freel_recent_searches');
                          }}
                        >
                          Clear
                        </button>
                      </div>
                      <div className="cmd-recent-tags">
                        {recentSearches.map((term) => (
                          <button
                            key={term}
                            className="cmd-recent-tag-btn"
                            onClick={() => setSearchQuery(term)}
                          >
                            <Search size={12} />
                            <span>{term}</span>
                          </button>
                        ))}
                      </div>
                    </div>
                  )}

                  {/* Quick System Navigation Shortcuts */}
                  <div className="cmd-quick-links-section">
                    <div className="cmd-quick-title">Quick Workspace Modules</div>
                    <div className="cmd-quick-grid">
                      <button
                        className="cmd-quick-card"
                        onClick={() => {
                          setActiveMenu(null);
                          navigate('/dashboard/shipments');
                        }}
                      >
                        <Ship size={16} className="text-blue-600" />
                        <div>
                          <div className="quick-name">Shipments Workspace</div>
                          <div className="quick-sub">Active voyages & live tracking</div>
                        </div>
                      </button>

                      <button
                        className="cmd-quick-card"
                        onClick={() => {
                          setActiveMenu(null);
                          navigate('/dashboard/bookings');
                        }}
                      >
                        <Layers size={16} className="text-indigo-600" />
                        <div>
                          <div className="quick-name">Carrier Bookings</div>
                          <div className="quick-sub">Direct space confirmations</div>
                        </div>
                      </button>

                      <button
                        className="cmd-quick-card"
                        onClick={() => {
                          setActiveMenu(null);
                          navigate('/dashboard/rfqs');
                        }}
                      >
                        <FileText size={16} className="text-purple-600" />
                        <div>
                          <div className="quick-name">RFQs Pipeline</div>
                          <div className="quick-sub">Customer spot inquiries</div>
                        </div>
                      </button>

                      <button
                        className="cmd-quick-card"
                        onClick={() => {
                          setActiveMenu(null);
                          navigate('/dashboard/customers');
                        }}
                      >
                        <Users size={16} className="text-emerald-600" />
                        <div>
                          <div className="quick-name">Customers & CRM</div>
                          <div className="quick-sub">B2B accounts directory</div>
                        </div>
                      </button>
                    </div>
                  </div>
                </div>
              )}
            </div>

            {/* Modal Keyboard Footer Guidance */}
            <div className="cmd-footer">
              <div className="cmd-footer-keys">
                <span>
                  <kbd>↑</kbd> <kbd>↓</kbd> Navigate
                </span>
                <span>
                  <kbd>↵</kbd> Select
                </span>
                <span>
                  <kbd>Esc</kbd> Close
                </span>
              </div>
              <span className="cmd-footer-brand">LogisticsHQ Global Search Engine</span>
            </div>
          </div>
        </div>
      )}
    </>
  );
}

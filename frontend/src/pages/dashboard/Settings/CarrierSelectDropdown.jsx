import React, { useState, useRef, useEffect } from 'react';
import { 
  Ship, Search, ChevronDown, Check, Sparkles, 
  ExternalLink, Layers, Radio, Shield, Globe, Info, X
} from 'lucide-react';
import './CarrierSelectDropdown.css';

// Carrier Brand Visual Definitions
const CARRIER_BRAND_META = {
  MAEU: {
    scac: 'MAEU',
    name: 'Maersk Line',
    bg: '#f0f9ff',
    border: '#bae6fd',
    accent: '#0284c7',
    badgeBg: '#e0f2fe',
    badgeText: '#0369a1',
    initials: 'ML',
    protocol: 'DCSA REST v2',
    tier: 'Tier 1 Global Ocean Carrier',
  },
  MSCU: {
    scac: 'MSCU',
    name: 'MSC (Mediterranean Shipping Co)',
    bg: '#fffbeb',
    border: '#fde68a',
    accent: '#d97706',
    badgeBg: '#fef3c7',
    badgeText: '#92400e',
    initials: 'MSC',
    protocol: 'Direct API & EDI',
    tier: 'Tier 1 Global Ocean Carrier',
  },
  HLCU: {
    scac: 'HLCU',
    name: 'Hapag-Lloyd',
    bg: '#fff7ed',
    border: '#fed7aa',
    accent: '#ea580c',
    badgeBg: '#ffedd5',
    badgeText: '#9a3412',
    initials: 'HL',
    protocol: 'DCSA REST v2',
    tier: 'Tier 1 Global Ocean Carrier',
  },
  CMDU: {
    scac: 'CMDU',
    name: 'CMA CGM',
    bg: '#eff6ff',
    border: '#bfdbfe',
    accent: '#2563eb',
    badgeBg: '#dbeafe',
    badgeText: '#1e40af',
    initials: 'CMA',
    protocol: 'DCSA API & EDI',
    tier: 'Tier 1 Global Ocean Carrier',
  },
  EGLV: {
    scac: 'EGLV',
    name: 'Evergreen Marine',
    bg: '#ecfdf5',
    border: '#a7f3d0',
    accent: '#059669',
    badgeBg: '#d1fae5',
    badgeText: '#065f46',
    initials: 'EMC',
    protocol: 'EDI / REST API',
    tier: 'Global Ocean Carrier',
  },
  COSU: {
    scac: 'COSU',
    name: 'COSCO Shipping Lines',
    bg: '#ecfeff',
    border: '#a5f3fc',
    accent: '#0891b2',
    badgeBg: '#cffafe',
    badgeText: '#155e75',
    initials: 'COS',
    protocol: 'REST API & EDI',
    tier: 'Global Ocean Carrier',
  },
  ONEY: {
    scac: 'ONEY',
    name: 'Ocean Network Express (ONE)',
    bg: '#fdf2f8',
    border: '#fbcfe8',
    accent: '#db2777',
    badgeBg: '#fce7f3',
    badgeText: '#9d174d',
    initials: 'ONE',
    protocol: 'DCSA REST API',
    tier: 'Global Ocean Carrier',
  },
  CUSTOM: {
    scac: 'CUSTOM',
    name: 'Custom Shipping Line / SCAC',
    bg: '#f5f3ff',
    border: '#ddd6fe',
    accent: '#7c3aed',
    badgeBg: '#ede9fe',
    badgeText: '#5b21b6',
    initials: '⚙',
    protocol: 'Custom SCAC REST / EDI',
    tier: 'Direct Custom Adapter',
  }
};

export default function CarrierSelectDropdown({
  providers = [],
  selectedProviderCode = '',
  onSelectProvider,
  disabled = false,
  error = null
}) {
  const [isOpen, setIsOpen] = useState(false);
  const [searchQuery, setSearchQuery] = useState('');
  const [focusedIndex, setFocusedIndex] = useState(-1);
  const containerRef = useRef(null);
  const searchInputRef = useRef(null);
  const listRef = useRef(null);

  // Merge backend providers with custom option
  const allOptions = React.useMemo(() => {
    const list = [...providers];
    const hasCustom = list.some(p => p.code === 'CUSTOM');
    if (!hasCustom) {
      list.push({
        code: 'CUSTOM',
        name: 'Custom Shipping Line / Other SCAC',
        scac: 'CUSTOM',
        auth_type: 'API_KEY',
        supported_capabilities: ['TRACKING', 'RATES', 'BOOKING', 'DOCUMENTS'],
        protocol: 'Direct REST / EDI'
      });
    }
    return list;
  }, [providers]);

  // Filtered options based on search query
  const filteredOptions = React.useMemo(() => {
    if (!searchQuery.trim()) return allOptions;
    const q = searchQuery.toLowerCase().trim();
    return allOptions.filter(p => {
      const name = (p.name || '').toLowerCase();
      const scac = (p.scac || '').toLowerCase();
      const code = (p.code || '').toLowerCase();
      const protocol = (p.protocol || '').toLowerCase();
      return name.includes(q) || scac.includes(q) || code.includes(q) || protocol.includes(q);
    });
  }, [allOptions, searchQuery]);

  // Selected Option Object
  const selectedOption = allOptions.find(p => p.code === selectedProviderCode);
  const selectedMeta = selectedOption ? (CARRIER_BRAND_META[selectedOption.scac] || CARRIER_BRAND_META[selectedOption.code] || {
    scac: selectedOption.scac || 'CA',
    name: selectedOption.name,
    bg: '#f8fafc',
    border: '#e2e8f0',
    accent: '#475569',
    badgeBg: '#f1f5f9',
    badgeText: '#334155',
    initials: (selectedOption.scac || selectedOption.name || 'CA').substring(0, 2).toUpperCase(),
    protocol: selectedOption.protocol || 'REST API',
    tier: 'Ocean Carrier',
  }) : null;

  // Handle Open/Close & Auto-focus
  const toggleDropdown = () => {
    if (disabled) return;
    const nextState = !isOpen;
    setIsOpen(nextState);
    if (nextState) {
      setSearchQuery('');
      setFocusedIndex(-1);
      setTimeout(() => {
        searchInputRef.current?.focus();
      }, 50);
    }
  };

  // Close on outside click
  useEffect(() => {
    const handleOutsideClick = (e) => {
      if (containerRef.current && !containerRef.current.contains(e.target)) {
        setIsOpen(false);
      }
    };
    if (isOpen) {
      document.addEventListener('mousedown', handleOutsideClick);
    }
    return () => document.removeEventListener('mousedown', handleOutsideClick);
  }, [isOpen]);

  // Keyboard navigation
  const handleKeyDown = (e) => {
    if (!isOpen) {
      if (e.key === 'Enter' || e.key === ' ' || e.key === 'ArrowDown') {
        e.preventDefault();
        toggleDropdown();
      }
      return;
    }

    if (e.key === 'Escape') {
      e.preventDefault();
      setIsOpen(false);
    } else if (e.key === 'ArrowDown') {
      e.preventDefault();
      setFocusedIndex(prev => (prev + 1 < filteredOptions.length ? prev + 1 : 0));
    } else if (e.key === 'ArrowUp') {
      e.preventDefault();
      setFocusedIndex(prev => (prev - 1 >= 0 ? prev - 1 : filteredOptions.length - 1));
    } else if (e.key === 'Enter') {
      e.preventDefault();
      if (focusedIndex >= 0 && focusedIndex < filteredOptions.length) {
        handleSelect(filteredOptions[focusedIndex].code);
      }
    }
  };

  const handleSelect = (code) => {
    onSelectProvider(code);
    setIsOpen(false);
    setSearchQuery('');
  };

  return (
    <div 
      className={`carrier-select-container ${disabled ? 'disabled' : ''} ${error ? 'has-error' : ''}`}
      ref={containerRef}
      onKeyDown={handleKeyDown}
    >
      {/* Trigger Button */}
      <button
        type="button"
        className={`carrier-select-trigger ${isOpen ? 'open' : ''} ${selectedOption ? 'has-selection' : ''}`}
        onClick={toggleDropdown}
        disabled={disabled}
        aria-haspopup="listbox"
        aria-expanded={isOpen}
      >
        {selectedOption && selectedMeta ? (
          <div className="carrier-selected-display">
            <div 
              className="carrier-logo-avatar"
              style={{
                background: selectedMeta.bg,
                borderColor: selectedMeta.border,
                color: selectedMeta.accent
              }}
            >
              <span>{selectedMeta.initials}</span>
            </div>

            <div className="carrier-selected-meta">
              <div className="carrier-selected-title-row">
                <span className="carrier-selected-name">{selectedOption.name}</span>
                {selectedOption.scac && selectedOption.scac !== 'CUSTOM' && (
                  <span 
                    className="carrier-scac-tag"
                    style={{
                      background: selectedMeta.badgeBg,
                      color: selectedMeta.badgeText
                    }}
                  >
                    SCAC: {selectedOption.scac}
                  </span>
                )}
                {selectedOption.code === 'CUSTOM' && (
                  <span className="carrier-scac-tag custom">Custom Adapter</span>
                )}
              </div>
              <div className="carrier-selected-subtext">
                <span className="protocol-pill">{selectedMeta.protocol || 'DCSA REST'}</span>
                <span className="dot-sep">•</span>
                <span className="tier-text">{selectedMeta.tier}</span>
              </div>
            </div>
          </div>
        ) : (
          <div className="carrier-placeholder-display">
            <div className="placeholder-icon">
              <Ship size={18} />
            </div>
            <div className="placeholder-text-group">
              <span className="placeholder-main">Select supported ocean carrier...</span>
              <span className="placeholder-sub">Maersk, MSC, Hapag-Lloyd, CMA CGM, or Custom SCAC</span>
            </div>
          </div>
        )}

        <div className="carrier-trigger-right">
          <div className={`trigger-chevron ${isOpen ? 'rotate' : ''}`}>
            <ChevronDown size={18} />
          </div>
        </div>
      </button>

      {/* Floating Dropdown Panel */}
      {isOpen && (
        <div className="carrier-dropdown-panel animate-panel-drop" role="listbox">
          {/* Search Header */}
          <div className="carrier-search-wrapper" onClick={e => e.stopPropagation()}>
            <Search size={15} className="carrier-search-icon" />
            <input
              ref={searchInputRef}
              type="text"
              className="carrier-search-input"
              placeholder="Search carrier name, SCAC (e.g. MAEU, MSC), protocol..."
              value={searchQuery}
              onChange={e => {
                setSearchQuery(e.target.value);
                setFocusedIndex(0);
              }}
            />
            {searchQuery && (
              <button 
                type="button"
                className="carrier-search-clear"
                onClick={() => {
                  setSearchQuery('');
                  searchInputRef.current?.focus();
                }}
              >
                <X size={13} />
              </button>
            )}
          </div>

          {/* Carrier Options List */}
          <div className="carrier-options-list" ref={listRef}>
            {filteredOptions.length === 0 ? (
              <div className="carrier-empty-search">
                <p>No matching carriers found for "{searchQuery}"</p>
                <button
                  type="button"
                  className="btn-use-custom-scac"
                  onClick={() => handleSelect('CUSTOM')}
                >
                  <Sparkles size={14} />
                  <span>Configure as Custom SCAC Carrier</span>
                </button>
              </div>
            ) : (
              filteredOptions.map((provider, idx) => {
                const isSelected = provider.code === selectedProviderCode;
                const isFocused = idx === focusedIndex;
                const meta = CARRIER_BRAND_META[provider.scac] || CARRIER_BRAND_META[provider.code] || {
                  scac: provider.scac || 'CA',
                  name: provider.name,
                  bg: '#f8fafc',
                  border: '#e2e8f0',
                  accent: '#475569',
                  badgeBg: '#f1f5f9',
                  badgeText: '#334155',
                  initials: (provider.scac || provider.name || 'CA').substring(0, 2).toUpperCase(),
                  protocol: provider.protocol || 'REST API',
                  tier: 'Ocean Carrier',
                };
                const isCustom = provider.code === 'CUSTOM';

                return (
                  <div
                    key={provider.code}
                    className={`carrier-option-card ${isSelected ? 'selected' : ''} ${isFocused ? 'focused' : ''} ${isCustom ? 'custom-card' : ''}`}
                    onClick={() => handleSelect(provider.code)}
                    onMouseEnter={() => setFocusedIndex(idx)}
                    role="option"
                    aria-selected={isSelected}
                  >
                    {/* Carrier Brand Avatar */}
                    <div 
                      className="carrier-card-avatar"
                      style={{
                        background: meta.bg,
                        borderColor: meta.border,
                        color: meta.accent
                      }}
                    >
                      {isCustom ? <Sparkles size={16} /> : <span>{meta.initials}</span>}
                    </div>

                    {/* Carrier Card Details */}
                    <div className="carrier-card-info">
                      <div className="carrier-card-top-row">
                        <span className="carrier-card-name">{provider.name}</span>
                        {provider.scac && provider.scac !== 'CUSTOM' && (
                          <span 
                            className="carrier-card-scac-pill"
                            style={{
                              background: meta.badgeBg,
                              color: meta.badgeText
                            }}
                          >
                            {provider.scac}
                          </span>
                        )}
                        {isCustom && (
                          <span className="carrier-card-scac-pill custom">Any SCAC</span>
                        )}
                      </div>

                      <div className="carrier-card-bottom-row">
                        <span className="carrier-protocol-badge">
                          {meta.protocol || 'DCSA REST'}
                        </span>
                        
                        {/* Capabilities preview */}
                        <div className="carrier-caps-preview">
                          {Array.isArray(provider.supported_capabilities) && provider.supported_capabilities.length > 0 ? (
                            provider.supported_capabilities.slice(0, 3).map(cap => (
                              <span key={cap} className="cap-mini-tag">
                                {cap === 'CONTRACT_RATES' ? 'Contracts' : cap.charAt(0) + cap.slice(1).toLowerCase()}
                              </span>
                            ))
                          ) : (
                            <span className="cap-mini-tag">Tracking</span>
                          )}
                        </div>
                      </div>
                    </div>

                    {/* Checkmark Indicator */}
                    <div className="carrier-card-check">
                      {isSelected ? (
                        <div className="check-bubble active">
                          <Check size={14} />
                        </div>
                      ) : (
                        <div className="check-bubble idle" />
                      )}
                    </div>
                  </div>
                );
              })
            )}
          </div>

          {/* Quick Helper Footer */}
          <div className="carrier-dropdown-footer">
            <Info size={13} className="text-slate-400" />
            <span>All connections support end-to-end telemetry and AES-256 encrypted credential vaults.</span>
          </div>
        </div>
      )}
    </div>
  );
}

import React, { useState, useRef, useEffect } from 'react';
import { ChevronDown, Check, Search, X } from 'lucide-react';
import './CustomSelect.css';

/**
 * Modern Custom Select Component
 * Replaces native unstyled <select> with a sleek, accessible, search-enabled dropdown.
 *
 * Props:
 * - value: Selected value
 * - onChange: Callback when value changes (passes value)
 * - options: Array of { value, label, subtitle, badge, icon, dotColor } or strings
 * - placeholder: String placeholder
 * - searchable: Boolean (enables search input inside dropdown if > 5 options)
 * - disabled: Boolean
 * - className: Additional CSS class
 * - size: 'small' | 'medium' | 'large'
 * - variant: 'default' | 'subtle' | 'pill'
 */
export default function CustomSelect({
  value,
  onChange,
  options = [],
  placeholder = 'Select an option...',
  searchable = false,
  disabled = false,
  className = '',
  size = 'medium',
  variant = 'default',
  icon: LeadingIcon,
}) {
  const [isOpen, setIsOpen] = useState(false);
  const [searchQuery, setSearchQuery] = useState('');
  const [highlightedIndex, setHighlightedIndex] = useState(0);
  const containerRef = useRef(null);
  const searchInputRef = useRef(null);

  // Normalize options
  const normalizedOptions = options.map((opt) => {
    if (typeof opt === 'object' && opt !== null) {
      return {
        value: opt.value,
        label: opt.label || opt.value,
        subtitle: opt.subtitle,
        badge: opt.badge,
        badgeColor: opt.badgeColor,
        icon: opt.icon,
        dotColor: opt.dotColor,
        disabled: opt.disabled,
      };
    }
    return { value: opt, label: String(opt) };
  });

  const selectedOption = normalizedOptions.find((opt) => opt.value === value);

  // Filter options by search
  const filteredOptions = normalizedOptions.filter((opt) => {
    if (!searchQuery.trim()) return true;
    const q = searchQuery.toLowerCase();
    return (
      opt.label.toLowerCase().includes(q) ||
      (opt.subtitle && opt.subtitle.toLowerCase().includes(q)) ||
      (opt.badge && opt.badge.toLowerCase().includes(q)) ||
      String(opt.value).toLowerCase().includes(q)
    );
  });

  // Close on outside click
  useEffect(() => {
    const handleClickOutside = (e) => {
      if (containerRef.current && !containerRef.current.contains(e.target)) {
        setIsOpen(false);
      }
    };
    if (isOpen) {
      document.addEventListener('mousedown', handleClickOutside);
    }
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, [isOpen]);

  // Focus search input when opened
  useEffect(() => {
    if (isOpen && searchable) {
      setTimeout(() => searchInputRef.current?.focus(), 40);
    }
    if (isOpen) {
      const idx = filteredOptions.findIndex((opt) => opt.value === value);
      setHighlightedIndex(idx >= 0 ? idx : 0);
    } else {
      setSearchQuery('');
    }
  }, [isOpen]);

  // Keyboard navigation
  const handleKeyDown = (e) => {
    if (disabled) return;

    if (!isOpen) {
      if (e.key === 'Enter' || e.key === ' ' || e.key === 'ArrowDown') {
        e.preventDefault();
        setIsOpen(true);
      }
      return;
    }

    if (e.key === 'ArrowDown') {
      e.preventDefault();
      setHighlightedIndex((prev) => (prev + 1 < filteredOptions.length ? prev + 1 : 0));
    } else if (e.key === 'ArrowUp') {
      e.preventDefault();
      setHighlightedIndex((prev) => (prev - 1 >= 0 ? prev - 1 : filteredOptions.length - 1));
    } else if (e.key === 'Enter') {
      e.preventDefault();
      const target = filteredOptions[highlightedIndex];
      if (target && !target.disabled) {
        onChange(target.value);
        setIsOpen(false);
      }
    } else if (e.key === 'Escape') {
      e.preventDefault();
      setIsOpen(false);
    }
  };

  const handleSelect = (opt) => {
    if (opt.disabled) return;
    onChange(opt.value);
    setIsOpen(false);
  };

  return (
    <div
      ref={containerRef}
      className={`modern-select-wrapper ${size} ${variant} ${isOpen ? 'open' : ''} ${disabled ? 'disabled' : ''} ${className}`}
      onKeyDown={handleKeyDown}
    >
      {/* Trigger Button */}
      <div
        className="modern-select-trigger"
        onClick={() => !disabled && setIsOpen((prev) => !prev)}
        role="combobox"
        aria-expanded={isOpen}
        aria-haspopup="listbox"
        tabIndex={disabled ? -1 : 0}
      >
        <div className="modern-select-value-wrap">
          {LeadingIcon && <LeadingIcon size={15} className="modern-select-leading-icon" />}
          {selectedOption ? (
            <div className="modern-select-selected-content">
              {selectedOption.dotColor && (
                <span className="modern-select-dot" style={{ background: selectedOption.dotColor }} />
              )}
              {selectedOption.icon && <span className="modern-select-opt-icon">{selectedOption.icon}</span>}
              <span className="modern-select-label-text">{selectedOption.label}</span>
              {selectedOption.badge && (
                <span className="modern-select-badge" style={{ color: selectedOption.badgeColor }}>
                  {selectedOption.badge}
                </span>
              )}
            </div>
          ) : (
            <span className="modern-select-placeholder">{placeholder}</span>
          )}
        </div>
        <ChevronDown size={14} className={`modern-select-chevron ${isOpen ? 'rotate' : ''}`} />
      </div>

      {/* Dropdown Menu */}
      {isOpen && (
        <div className="modern-select-menu animate-scale-up" role="listbox">
          {searchable && (
            <div className="modern-select-search-header">
              <Search size={14} className="modern-select-search-icon" />
              <input
                ref={searchInputRef}
                type="text"
                className="modern-select-search-input"
                placeholder="Search options..."
                value={searchQuery}
                onChange={(e) => {
                  setSearchQuery(e.target.value);
                  setHighlightedIndex(0);
                }}
                onClick={(e) => e.stopPropagation()}
              />
              {searchQuery && (
                <button
                  type="button"
                  className="modern-select-search-clear"
                  onClick={(e) => {
                    e.stopPropagation();
                    setSearchQuery('');
                  }}
                >
                  <X size={12} />
                </button>
              )}
            </div>
          )}

          <div className="modern-select-options-list">
            {filteredOptions.length > 0 ? (
              filteredOptions.map((opt, idx) => {
                const isSelected = opt.value === value;
                const isHighlighted = idx === highlightedIndex;

                return (
                  <div
                    key={opt.value}
                    className={`modern-select-option ${isSelected ? 'selected' : ''} ${isHighlighted ? 'highlighted' : ''} ${opt.disabled ? 'disabled' : ''}`}
                    onClick={(e) => {
                      e.stopPropagation();
                      handleSelect(opt);
                    }}
                    onMouseEnter={() => setHighlightedIndex(idx)}
                    role="option"
                    aria-selected={isSelected}
                  >
                    <div className="modern-select-opt-left">
                      {opt.dotColor && (
                        <span className="modern-select-dot" style={{ background: opt.dotColor }} />
                      )}
                      {opt.icon && <span className="modern-select-opt-icon">{opt.icon}</span>}
                      <div className="modern-select-opt-labels">
                        <span className="modern-select-opt-title">{opt.label}</span>
                        {opt.subtitle && <span className="modern-select-opt-subtitle">{opt.subtitle}</span>}
                      </div>
                      {opt.badge && (
                        <span className="modern-select-badge" style={{ color: opt.badgeColor }}>
                          {opt.badge}
                        </span>
                      )}
                    </div>
                    {isSelected && <Check size={14} className="modern-select-check" />}
                  </div>
                );
              })
            ) : (
              <div className="modern-select-no-results">No options found</div>
            )}
          </div>
        </div>
      )}
    </div>
  );
}

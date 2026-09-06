import React, { useState, useRef, useEffect } from 'react';
import { ChevronDown, Check } from 'lucide-react';

export default function CustomDropdown({
  label,
  value,
  onChange,
  options = [],
  placeholder = 'Select option...',
  className = '',
  align = 'left',
  disabled = false,
  size = 'default' // 'default' | 'small' | 'compact'
}) {
  const [isOpen, setIsOpen] = useState(false);
  const dropdownRef = useRef(null);

  // Normalize options array: support both ["A", "B"] and [{ value: "A", label: "A" }]
  const normalizedOptions = options.map((opt) => {
    if (typeof opt === 'object' && opt !== null) {
      return {
        value: opt.value,
        label: opt.label || opt.value,
        icon: opt.icon,
        badge: opt.badge,
        badgeColor: opt.badgeColor
      };
    }
    return { value: opt, label: opt };
  });

  const selectedOption = normalizedOptions.find((opt) => opt.value === value);

  useEffect(() => {
    const handleOutsideClick = (e) => {
      if (dropdownRef.current && !dropdownRef.current.contains(e.target)) {
        setIsOpen(false);
      }
    };

    const handleKeyDown = (e) => {
      if (e.key === 'Escape') {
        setIsOpen(false);
      }
    };

    if (isOpen) {
      document.addEventListener('mousedown', handleOutsideClick);
      document.addEventListener('keydown', handleKeyDown);
    }
    return () => {
      document.removeEventListener('mousedown', handleOutsideClick);
      document.removeEventListener('keydown', handleKeyDown);
    };
  }, [isOpen]);

  const handleSelect = (optVal) => {
    onChange(optVal);
    setIsOpen(false);
  };

  return (
    <div
      className={`c-dropdown-container ${className} ${disabled ? 'c-dropdown-disabled' : ''} ${size ? `c-dropdown-${size}` : ''}`}
      ref={dropdownRef}
    >
      <button
        type="button"
        className={`c-dropdown-trigger ${isOpen ? 'c-dropdown-open' : ''}`}
        onClick={() => !disabled && setIsOpen(!isOpen)}
        disabled={disabled}
        aria-haspopup="listbox"
        aria-expanded={isOpen}
      >
        <div className="c-dropdown-label-group">
          {label && <span className="c-dropdown-label">{label}</span>}
          <span className="c-dropdown-value">
            {selectedOption ? selectedOption.label : placeholder}
          </span>
        </div>
        <ChevronDown
          size={14}
          className={`c-dropdown-chevron ${isOpen ? 'c-dropdown-chevron-up' : ''}`}
        />
      </button>

      {isOpen && (
        <div className={`c-dropdown-menu c-dropdown-align-${align}`} role="listbox">
          <div className="c-dropdown-list">
            {normalizedOptions.map((opt) => {
              const isSelected = opt.value === value;
              return (
                <button
                  type="button"
                  key={opt.value}
                  className={`c-dropdown-option ${isSelected ? 'c-dropdown-option-selected' : ''}`}
                  onClick={() => handleSelect(opt.value)}
                  role="option"
                  aria-selected={isSelected}
                >
                  <div className="c-dropdown-option-content">
                    {opt.icon && <span className="c-dropdown-option-icon">{opt.icon}</span>}
                    <span className="c-dropdown-option-label">{opt.label}</span>
                    {opt.badge && (
                      <span
                        className="c-dropdown-option-badge"
                        style={opt.badgeColor ? { background: opt.badgeColor.bg, color: opt.badgeColor.fg } : {}}
                      >
                        {opt.badge}
                      </span>
                    )}
                  </div>
                  {isSelected && <Check size={14} className="c-dropdown-check-icon" />}
                </button>
              );
            })}
          </div>
        </div>
      )}
    </div>
  );
}

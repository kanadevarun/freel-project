import React, { useState, useRef, useEffect } from 'react';
import { ChevronDown, Check, Filter } from 'lucide-react';

const STATUS_OPTIONS = [
  { value: 'ALL', label: 'All Statuses', dot: '#64748b', bg: '#f1f5f9' },
  { value: 'ACTIVE', label: 'Active Only', dot: '#10b981', bg: '#ecfdf5' },
  { value: 'ERROR', label: 'Needs Attention', dot: '#f59e0b', bg: '#fffbeb' },
  { value: 'DISABLED', label: 'Disabled Only', dot: '#94a3b8', bg: '#f8fafc' },
];

export default function StatusFilterDropdown({ value = 'ALL', onChange }) {
  const [isOpen, setIsOpen] = useState(false);
  const dropdownRef = useRef(null);

  const selected = STATUS_OPTIONS.find(o => o.value === value) || STATUS_OPTIONS[0];

  useEffect(() => {
    const handleOutsideClick = (e) => {
      if (dropdownRef.current && !dropdownRef.current.contains(e.target)) {
        setIsOpen(false);
      }
    };
    if (isOpen) {
      document.addEventListener('mousedown', handleOutsideClick);
    }
    return () => document.removeEventListener('mousedown', handleOutsideClick);
  }, [isOpen]);

  return (
    <div className="status-filter-container" ref={dropdownRef} style={{ position: 'relative' }}>
      <button
        type="button"
        className={`status-filter-trigger ${isOpen ? 'active' : ''}`}
        onClick={() => setIsOpen(!isOpen)}
        aria-haspopup="listbox"
        aria-expanded={isOpen}
        style={{
          display: 'inline-flex',
          alignItems: 'center',
          gap: '8px',
          padding: '7px 12px',
          background: '#ffffff',
          border: '1px solid #e2e8f0',
          borderRadius: '8px',
          fontSize: '13px',
          fontWeight: 500,
          color: '#334155',
          cursor: 'pointer',
          transition: 'all 0.15s ease',
          boxShadow: '0 1px 2px rgba(0,0,0,0.02)'
        }}
      >
        <div style={{
          width: '7px',
          height: '7px',
          borderRadius: '50%',
          backgroundColor: selected.dot,
          boxShadow: selected.value === 'ACTIVE' ? '0 0 6px rgba(16, 185, 129, 0.5)' : undefined
        }} />
        <span>{selected.label}</span>
        <ChevronDown size={14} style={{ color: '#94a3b8', transform: isOpen ? 'rotate(180deg)' : 'none', transition: 'transform 0.15s' }} />
      </button>

      {isOpen && (
        <div
          className="status-filter-menu animate-fade-in-down"
          role="listbox"
          style={{
            position: 'absolute',
            right: 0,
            top: 'calc(100% + 4px)',
            zIndex: 60,
            background: 'rgba(255, 255, 255, 0.98)',
            backdropFilter: 'blur(12px)',
            border: '1px solid #e2e8f0',
            borderRadius: '10px',
            boxShadow: '0 10px 20px -5px rgba(0, 0, 0, 0.1), 0 4px 6px -2px rgba(0, 0, 0, 0.05)',
            padding: '4px',
            minWidth: '160px',
            display: 'flex',
            flexDirection: 'column',
            gap: '2px'
          }}
        >
          {STATUS_OPTIONS.map(opt => {
            const isSelected = opt.value === value;
            return (
              <button
                key={opt.value}
                type="button"
                role="option"
                aria-selected={isSelected}
                onClick={() => {
                  onChange(opt.value);
                  setIsOpen(false);
                }}
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'space-between',
                  padding: '7px 10px',
                  border: 'none',
                  borderRadius: '6px',
                  background: isSelected ? '#eff6ff' : 'transparent',
                  color: isSelected ? '#1d4ed8' : '#334155',
                  fontSize: '12.5px',
                  fontWeight: isSelected ? 600 : 500,
                  cursor: 'pointer',
                  textAlign: 'left',
                  transition: 'background 0.12s'
                }}
                onMouseEnter={e => {
                  if (!isSelected) e.currentTarget.style.background = '#f8fafc';
                }}
                onMouseLeave={e => {
                  if (!isSelected) e.currentTarget.style.background = 'transparent';
                }}
              >
                <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                  <div style={{
                    width: '6px',
                    height: '6px',
                    borderRadius: '50%',
                    backgroundColor: opt.dot
                  }} />
                  <span>{opt.label}</span>
                </div>
                {isSelected && <Check size={13} color="#2563eb" />}
              </button>
            );
          })}
        </div>
      )}
    </div>
  );
}

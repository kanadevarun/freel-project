import React from 'react';

const STATUS_THEMES = {
  emerald: { bg: '#ECFDF5', text: '#059669', border: '#A7F3D0', dot: '#10B981' },
  green: { bg: '#F0FDF4', text: '#15803D', border: '#BBF7D0', dot: '#22C55E' },
  amber: { bg: '#FFFBEB', text: '#D97706', border: '#FDE68A', dot: '#F59E0B' },
  blue: { bg: '#EFF6FF', text: '#2563EB', border: '#BFDBFE', dot: '#3B82F6' },
  indigo: { bg: '#EEF2FF', text: '#4F46E5', border: '#C7D2FE', dot: '#6366F1' },
  purple: { bg: '#FAF5FF', text: '#7E22CE', border: '#E9D5FF', dot: '#A855F7' },
  violet: { bg: '#F5F3FF', text: '#6D28D9', border: '#DDD6FE', dot: '#8B5CF6' },
  teal: { bg: '#F0FDFA', text: '#0D9488', border: '#99F6E4', dot: '#14B8A6' },
  red: { bg: '#FEF2F2', text: '#DC2626', border: '#FECACA', dot: '#EF4444' },
  gray: { bg: '#F8FAFC', text: '#64748B', border: '#E2E8F0', dot: '#94A3B8' },
};

export default function RFQStatusBadge({ label, color = 'blue', size = 'medium', pulse = false }) {
  const theme = STATUS_THEMES[color] || STATUS_THEMES.blue;

  const sizeStyles = {
    small: { fontSize: '11px', padding: '3px 8px', dotSize: '6px' },
    medium: { fontSize: '12px', padding: '4px 10px', dotSize: '7px' },
    large: { fontSize: '13px', padding: '6px 14px', dotSize: '8px' },
  }[size] || { fontSize: '12px', padding: '4px 10px', dotSize: '7px' };

  return (
    <span
      style={{
        display: 'inline-flex',
        alignItems: 'center',
        gap: '6px',
        backgroundColor: theme.bg,
        color: theme.text,
        border: `1px solid ${theme.border}`,
        borderRadius: '9999px',
        fontWeight: 600,
        fontSize: sizeStyles.fontSize,
        padding: sizeStyles.padding,
        lineHeight: 1.2,
        letterSpacing: '0.01em',
      }}
    >
      <span
        style={{
          width: sizeStyles.dotSize,
          height: sizeStyles.dotSize,
          borderRadius: '50%',
          backgroundColor: theme.dot,
          display: 'inline-block',
          boxShadow: pulse ? `0 0 0 3px ${theme.border}` : 'none',
        }}
      />
      {label}
    </span>
  );
}

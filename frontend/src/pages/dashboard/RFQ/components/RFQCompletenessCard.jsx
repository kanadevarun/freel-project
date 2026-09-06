import React from 'react';

export default function RFQCompletenessCard({ completeness, onNavigateCargo }) {
  const { fields, completedCount, totalCount, percentage, isComplete, missingFields } = completeness;

  return (
    <div
      style={{
        background: '#FFFFFF',
        borderRadius: '12px',
        border: '1px solid #E2E8F0',
        padding: '20px',
        boxShadow: '0 1px 3px rgba(0,0,0,0.02)',
      }}
    >
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: '16px' }}>
        <div>
          <h4 style={{ fontSize: '13px', fontWeight: 700, color: '#0F172A', textTransform: 'uppercase', letterSpacing: '0.05em', margin: 0 }}>
            Shipment Information
          </h4>
          <p style={{ fontSize: '12px', color: '#64748B', margin: '2px 0 0 0' }}>
            7 Core operational parameters required for quotation
          </p>
        </div>
        <div
          style={{
            display: 'inline-flex',
            alignItems: 'center',
            gap: '6px',
            background: isComplete ? '#ECFDF5' : '#FFFBEB',
            color: isComplete ? '#059669' : '#D97706',
            border: `1px solid ${isComplete ? '#A7F3D0' : '#FDE68A'}`,
            borderRadius: '20px',
            padding: '4px 10px',
            fontSize: '12px',
            fontWeight: 700,
          }}
        >
          <span>{isComplete ? '✓' : '●'}</span>
          <span>{completedCount}/{totalCount} Complete</span>
        </div>
      </div>

      {/* Progress Bar */}
      <div style={{ width: '100%', height: '7px', background: '#F1F5F9', borderRadius: '4px', overflow: 'hidden', marginBottom: '18px' }}>
        <div
          style={{
            width: `${percentage}%`,
            height: '100%',
            background: isComplete
              ? 'linear-gradient(90deg, #10B981 0%, #059669 100%)'
              : 'linear-gradient(90deg, #F59E0B 0%, #D97706 100%)',
            transition: 'width 0.4s ease',
          }}
        />
      </div>

      {/* Checklist Grid */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(210px, 1fr))', gap: '10px' }}>
        {fields.map((field) => (
          <div
            key={field.key}
            style={{
              display: 'flex',
              alignItems: 'flex-start',
              gap: '10px',
              padding: '10px 12px',
              borderRadius: '8px',
              background: field.complete ? '#F8FAFC' : '#FEF2F2',
              border: `1px solid ${field.complete ? '#E2E8F0' : '#FCA5A5'}`,
            }}
          >
            <div
              style={{
                width: '18px',
                height: '18px',
                borderRadius: '50%',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                fontSize: '11px',
                fontWeight: 800,
                marginTop: '1px',
                flexShrink: 0,
                background: field.complete ? '#10B981' : '#EF4444',
                color: '#FFFFFF',
              }}
            >
              {field.complete ? '✓' : '✗'}
            </div>
            <div style={{ flex: 1, minWidth: 0 }}>
              <div style={{ fontSize: '11.5px', fontWeight: 600, color: field.complete ? '#475569' : '#DC2626' }}>
                {field.label}
              </div>
              <div
                style={{
                  fontSize: '12.5px',
                  fontWeight: 600,
                  color: field.complete ? '#0F172A' : '#991B1B',
                  whiteSpace: 'nowrap',
                  overflow: 'hidden',
                  textOverflow: 'ellipsis',
                  marginTop: '1px',
                }}
              >
                {field.complete ? field.value : 'Required'}
              </div>
            </div>
          </div>
        ))}
      </div>

      {!isComplete && (
        <div
          style={{
            marginTop: '14px',
            padding: '10px 14px',
            borderRadius: '8px',
            background: '#FFFBEB',
            border: '1px solid #FDE68A',
            fontSize: '12px',
            color: '#B45309',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
          }}
        >
          <div style={{ display: 'flex', alignItems: 'center', gap: '6px' }}>
            <span>⚠️</span>
            <span>Missing: <strong>{missingFields.join(', ')}</strong></span>
          </div>
          {onNavigateCargo && (
            <button
              onClick={onNavigateCargo}
              style={{
                background: 'none',
                border: 'none',
                color: '#2563EB',
                fontWeight: 700,
                fontSize: '12px',
                cursor: 'pointer',
                textDecoration: 'underline',
              }}
            >
              Review Cargo →
            </button>
          )}
        </div>
      )}
    </div>
  );
}

import React from 'react';
import { Loader2 } from 'lucide-react';
import './SharedAI.css';

export default function AILoadingState({
  message = 'Synthesizing cross-module intelligence...',
  submessage = 'Auditing persistent records across operational, commercial, and financial domains.',
  rows = 2,
}) {
  return (
    <div className="sai-card" style={{ padding: '20px 16px', textAlign: 'center' }} data-testid="ai-loading-state">
      <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', gap: 6, marginBottom: 14 }}>
        <Loader2 size={24} className="animate-spin text-teal-600" />
        <strong style={{ fontSize: '0.84rem', color: '#0f172a' }}>{message}</strong>
        {submessage && <span style={{ fontSize: '0.72rem', color: '#64748b' }}>{submessage}</span>}
      </div>

      <div style={{ display: 'flex', flexDirection: 'column', gap: 8, width: '100%', maxWidth: 440, margin: '0 auto' }}>
        {Array.from({ length: rows }).map((_, i) => (
          <div key={i} className="sai-skeleton" style={{ height: 18, width: i === 0 ? '100%' : '80%' }} />
        ))}
      </div>
    </div>
  );
}

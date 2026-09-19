import React from 'react';
import { Cpu, RefreshCw, CheckCircle2, Clock, AlertOctagon, Pause } from 'lucide-react';
import './SharedAI.css';

export default function AIWorkforceSummary({
  summary = {},
  onRefresh = null,
  isRefreshing = false,
  title = 'AI Workforce Execution Summary',
}) {
  const queued = summary.queued || summary.total_queued || 0;
  const processing = summary.processing || summary.total_processing || 0;
  const awaitingSignoff = summary.waiting_for_approval || summary.total_waiting_approval || 0;
  const completed = summary.completed || summary.total_completed || 0;
  const failed = summary.failed || summary.total_failed || 0;

  return (
    <div className="sai-card" style={{ marginBottom: 14 }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 10 }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
          <Cpu size={15} className="text-teal-600" />
          <h4 style={{ margin: 0, fontSize: '0.84rem', fontWeight: 750, color: '#0f172a' }}>{title}</h4>
        </div>
        {onRefresh && (
          <button
            className="sai-btn sai-btn--icon"
            onClick={onRefresh}
            disabled={isRefreshing}
            title="Refresh workforce metrics"
          >
            <RefreshCw size={12} className={isRefreshing ? 'animate-spin text-teal-600' : ''} />
          </button>
        )}
      </div>

      <div className="sai-wf-grid">
        <div className="sai-wf-stat-card">
          <span className="sai-wf-stat-label">Queued</span>
          <span className="sai-wf-stat-val" style={{ color: '#475569' }}>{queued}</span>
        </div>

        <div className="sai-wf-stat-card">
          <span className="sai-wf-stat-label">Processing</span>
          <span className="sai-wf-stat-val" style={{ color: '#2563eb' }}>{processing}</span>
        </div>

        <div className="sai-wf-stat-card">
          <span className="sai-wf-stat-label">Awaiting Sign-Off</span>
          <span className="sai-wf-stat-val" style={{ color: '#b45309' }}>{awaitingSignoff}</span>
        </div>

        <div className="sai-wf-stat-card">
          <span className="sai-wf-stat-label">Completed</span>
          <span className="sai-wf-stat-val" style={{ color: '#059669' }}>{completed}</span>
        </div>

        <div className="sai-wf-stat-card">
          <span className="sai-wf-stat-label">Failed / Stale</span>
          <span className="sai-wf-stat-val" style={{ color: failed > 0 ? '#dc2626' : '#64748b' }}>{failed}</span>
        </div>
      </div>
    </div>
  );
}

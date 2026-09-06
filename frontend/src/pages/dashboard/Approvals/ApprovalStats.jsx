import React from 'react';
import { Clock, CheckCircle2, XCircle, AlertTriangle } from 'lucide-react';

export default function ApprovalStats({ stats, onFilterOverdue }) {
  return (
    <div className="approval-stats-grid">
      {/* Pending Approvals */}
      <div className="approval-stat-card amber">
        <div className="stat-icon-box amber">
          <Clock size={22} />
        </div>
        <div className="stat-info">
          <div className="stat-value">{stats.pending}</div>
          <div className="stat-label">Pending Approvals</div>
          <div className="stat-trend green">{stats.pendingTrend}</div>
        </div>
      </div>

      {/* Approved */}
      <div className="approval-stat-card green">
        <div className="stat-icon-box green">
          <CheckCircle2 size={22} />
        </div>
        <div className="stat-info">
          <div className="stat-value">{stats.approved}</div>
          <div className="stat-label">Approved</div>
          <div className="stat-trend green">{stats.approvedTrend}</div>
        </div>
      </div>

      {/* Rejected */}
      <div className="approval-stat-card red">
        <div className="stat-icon-box red">
          <XCircle size={22} />
        </div>
        <div className="stat-info">
          <div className="stat-value">{stats.rejected}</div>
          <div className="stat-label">Rejected</div>
          <div className="stat-trend red">{stats.rejectedTrend}</div>
        </div>
      </div>

      {/* Overdue */}
      <div className="approval-stat-card purple">
        <div className="stat-icon-box purple">
          <AlertTriangle size={22} />
        </div>
        <div className="stat-info">
          <div className="stat-value">{stats.overdue}</div>
          <div className="stat-label">Overdue</div>
          <button type="button" className="stat-link-btn" onClick={onFilterOverdue}>
            View overdue
          </button>
        </div>
      </div>
    </div>
  );
}

import React from 'react';
import './Widgets.css';

export default function AIWorkforceWidget({ status }) {
  if (!status) return null;

  return (
    <div className="mc-widget ai-workforce-widget">
      <div className="ai-header">
        <div className="ai-icon">🤖</div>
        <div className="ai-title">AI Workforce Status</div>
      </div>
      <div className="ai-stats">
        <div className="ai-stat-item">
          <span className="label">Active Agents</span>
          <span className="value">{status.active_agents}</span>
        </div>
        <div className="ai-stat-item">
          <span className="label">Tasks Completed</span>
          <span className="value">{status.tasks_finished}</span>
        </div>
        <div className="ai-stat-item">
          <span className="label">System Health</span>
          <span className="value text-green">{status.health_score}%</span>
        </div>
      </div>
    </div>
  );
}

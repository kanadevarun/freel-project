import React from 'react';
import './Widgets.css';

export default function StatCardWidget({ title, value, prefix = "", suffix = "", trend }) {
  return (
    <div className="mc-widget stat-card">
      <div className="stat-title">{title}</div>
      <div className="stat-value">
        {prefix}{value}{suffix}
      </div>
      {trend && (
        <div className={`stat-trend ${trend >= 0 ? 'positive' : 'negative'}`}>
          {trend >= 0 ? '↑' : '↓'} {Math.abs(trend)}% vs last month
        </div>
      )}
    </div>
  );
}

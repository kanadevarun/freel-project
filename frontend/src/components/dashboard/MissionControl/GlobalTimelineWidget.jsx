import React from 'react';
import './Widgets.css';

export default function GlobalTimelineWidget({ timeline }) {
  return (
    <div className="mc-widget timeline-widget">
      <div className="widget-header">
        <h3>Recent Activity</h3>
      </div>
      
      <div className="widget-body">
        {(!timeline || timeline.length === 0) ? (
          <div className="empty-state">
            <p>No recent activity</p>
          </div>
        ) : (
          <ul className="timeline-list">
            {timeline.map((event, idx) => (
              <li key={idx} className="timeline-item">
                <div className="timeline-marker" style={{ backgroundColor: event.color || '#64748b' }}></div>
                <div className="timeline-content">
                  <div className="timeline-time">{event.time}</div>
                  <div className="timeline-title">{event.title}</div>
                  <div className="timeline-desc">{event.description}</div>
                </div>
              </li>
            ))}
          </ul>
        )}
      </div>
    </div>
  );
}

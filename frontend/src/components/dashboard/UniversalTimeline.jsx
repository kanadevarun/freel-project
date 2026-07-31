import React from 'react';
import './UniversalTimeline.css';

/**
 * UniversalTimeline displays a vertical timeline of events.
 * It is meant to be shared across Leads, RFQs, Opportunities, and Shipments.
 * 
 * @param {Array} events - Array of event objects { id, time, title, description, icon, color }
 */
export default function UniversalTimeline({ events }) {
  if (!events || events.length === 0) {
    return (
      <div className="universal-timeline empty">
        <p>No activity yet.</p>
      </div>
    );
  }

  return (
    <div className="universal-timeline">
      {events.map((event, index) => (
        <div key={event.id || index} className="timeline-item">
          <div className="timeline-left">
            <div className="timeline-time">{event.time}</div>
            <div className="timeline-line-container">
              <div 
                className="timeline-dot" 
                style={{ backgroundColor: event.color || 'var(--primary-color)' }}
              >
                {event.icon}
              </div>
              {index < events.length - 1 && <div className="timeline-line"></div>}
            </div>
          </div>
          <div className="timeline-content">
            <div className="timeline-title">{event.title}</div>
            {event.description && <div className="timeline-desc">{event.description}</div>}
          </div>
        </div>
      ))}
    </div>
  );
}

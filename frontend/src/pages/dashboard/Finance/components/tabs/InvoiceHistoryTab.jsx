import React from 'react';
import { Clock, User } from 'lucide-react';
import './InvoiceSubTabs.css';

export default function InvoiceHistoryTab({ invoice }) {
  const history = invoice?.history || [];

  if (history.length === 0) {
    return (
      <div className="invoice-subtab-container empty-subtab">
        <Clock size={28} className="empty-icon-gray" />
        <p className="empty-subtab-text">No audit trail logged yet.</p>
      </div>
    );
  }

  return (
    <div className="invoice-subtab-container history-tab">
      <div className="history-timeline">
        {history.map((event, idx) => (
          <div key={event.id || idx} className="history-timeline-item">
            <div className="timeline-node">
              <div className="node-dot" />
              {idx < history.length - 1 && <div className="timeline-connector" />}
            </div>
            <div className="history-content">
              <div className="history-header">
                <span className="history-title">{event.title}</span>
                <span className="history-time">{event.timestamp}</span>
              </div>
              <p className="history-desc">{event.description}</p>
              {event.user && (
                <div className="history-user">
                  <User size={12} /> <span>{event.user}</span>
                </div>
              )}
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

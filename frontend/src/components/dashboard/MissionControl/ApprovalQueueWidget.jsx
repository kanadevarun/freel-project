import React from 'react';
import { useNavigate } from 'react-router-dom';
import './Widgets.css';

export default function ApprovalQueueWidget({ queue }) {
  const navigate = useNavigate();

  const handleNavigate = (task) => {
    // Determine route based on task type
    if (task.type === 'RFQ_QUOTE_DRAFT') {
      navigate(`/dashboard/rfqs/${task.ref_id}`);
    } else if (task.type === 'LEAD_REVIEW') {
      navigate(`/dashboard/leads/${task.ref_id}`);
    } else {
      console.warn("Unknown task type:", task.type);
    }
  };

  return (
    <div className="mc-widget approval-queue-widget">
      <div className="widget-header">
        <h3>Human Approval Queue</h3>
        <span className="badge">{queue?.length || 0} Pending</span>
      </div>
      
      <div className="widget-body">
        {(!queue || queue.length === 0) ? (
          <div className="empty-state">
            <span className="icon">✅</span>
            <p>You're all caught up!</p>
          </div>
        ) : (
          <ul className="queue-list">
            {queue.map((task) => (
              <li key={task.id} className="queue-item" onClick={() => handleNavigate(task)}>
                <div className="task-icon">
                  {task.type === 'RFQ_QUOTE_DRAFT' ? '💰' : '📝'}
                </div>
                <div className="task-details">
                  <h4>{task.title}</h4>
                  <p>{task.subtitle}</p>
                </div>
                <div className="task-action">
                  <button className="btn-secondary btn-sm">Review</button>
                </div>
              </li>
            ))}
          </ul>
        )}
      </div>
    </div>
  );
}

import React from 'react';
import './OnboardingChecklist.css';

export default function OnboardingChecklist() {
  const completedTasks = [
    'Organization Created',
    'RBAC Configured',
    'Workspace Ready'
  ];

  const remainingTasks = [
    'Invite Team',
    'Create Customer',
    'Import Contacts',
    'Create First RFQ',
    'Send First Quotation',
    'Configure Email',
    'Connect Carrier APIs',
    'AI Setup (Coming Soon)'
  ];

  return (
    <div className="onboarding-checklist-container">
      <div className="onboarding-header">
        <h2>Welcome to LogisticsHQ 🎉</h2>
        <p>You're almost ready to start automating your freight operations. Complete these steps to get your workspace fully operational.</p>
      </div>

      <div className="onboarding-content">
        <div className="onboarding-section completed">
          <div className="section-title">Completed</div>
          <ul className="task-list">
            {completedTasks.map((task, index) => (
              <li key={index} className="task-item completed">
                <span className="checkbox-icon">✓</span>
                <span className="task-text">{task}</span>
              </li>
            ))}
          </ul>
        </div>

        <div className="onboarding-section remaining">
          <div className="section-title">Remaining</div>
          <ul className="task-list">
            {remainingTasks.map((task, index) => (
              <li key={index} className="task-item pending">
                <span className="checkbox-icon">☐</span>
                <span className="task-text">{task}</span>
              </li>
            ))}
          </ul>
        </div>
      </div>
    </div>
  );
}

import React from 'react';

/**
 * Visualizes the current state of the AI Agent.
 */
export default function AgentStatusBadge({ status }) {
  if (!status || status === 'IDLE') return null;

  let label = 'Thinking...';
  let colorClass = 'agent-blue';
  let pulse = true;

  switch (status) {
    case 'COLLECTING_INFORMATION':
      label = 'Collecting Carrier Data...';
      break;
    case 'ANALYZING_DATA':
      label = 'Analyzing Trade Lane...';
      break;
    case 'WAITING_FOR_LLM':
      label = 'Generating Recommendation...';
      colorClass = 'agent-purple';
      break;
    case 'GENERATING_DRAFT':
      label = 'Drafting Quotes...';
      colorClass = 'agent-purple';
      break;
    case 'WAITING_FOR_HUMAN_REVIEW':
      label = 'Human Review Required';
      colorClass = 'agent-amber';
      pulse = false;
      break;
    case 'COMPLETED':
      label = 'Completed';
      colorClass = 'agent-green';
      pulse = false;
      break;
    case 'ERROR':
      label = 'Agent Error';
      colorClass = 'agent-red';
      pulse = false;
      break;
    default:
      label = status;
  }

  return (
    <div className={`agent-status-badge ${colorClass}`}>
      {pulse && <div className="pulse-dot"></div>}
      <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
        <path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z"></path>
      </svg>
      {label}
    </div>
  );
}

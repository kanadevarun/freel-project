import React from 'react';
import './AgentStatusTimeline.css';

/**
 * AgentStatusTimeline is a premium React component that renders the execution steps
 * of the LangGraph AI sidecar pipeline in a visual timeline interface.
 *
 * Simple meaning:
 *   Instead of showing raw log files in monospaced black screens, this component shows
 *   a beautiful visual pathway (timeline) detailing where the contract document is
 *   (e.g., OCR, Classification, Parsing, Human Review) with icons and colored nodes.
 *
 * Example input logs array:
 *   [
 *     { "step": "OCR_PROCESSING", "timestamp": "2026-08-08T22:10:48Z", "message": "Extracted text from PDF" },
 *     { "step": "CLASSIFICATION", "timestamp": "2026-08-08T22:11:02Z", "message": "Identified Maersk carrier SCAC" }
 *   ]
 */
export default function AgentStatusTimeline({ logs = [], status }) {
  
  // Helper to map log steps to matching graphical icons and colors
  const getStepConfig = (step) => {
    const s = step.toUpperCase();
    if (s.includes('OCR')) {
      return { icon: '📄', label: 'OCR Text Ingestion', colorClass: 'node-blue' };
    }
    if (s.includes('CLASSIF')) {
      return { icon: '🏷️', label: 'Carrier Classification', colorClass: 'node-indigo' };
    }
    if (s.includes('PARS') || s.includes('TABLE')) {
      return { icon: '📊', label: 'Table & Footnote Extraction', colorClass: 'node-teal' };
    }
    if (s.includes('VALID')) {
      return { icon: '🛡️', label: 'Pricing & LOCODE Validator', colorClass: 'node-purple' };
    }
    if (s.includes('HUMAN') || s.includes('INTERRUPT')) {
      return { icon: '👤', label: 'Human-in-the-Loop Review', colorClass: 'node-warning' };
    }
    if (s.includes('INGEST')) {
      return { icon: '💾', label: 'System Ingestion Complete', colorClass: 'node-green' };
    }
    if (s.includes('ERROR') || s.includes('FAIL')) {
      return { icon: '🛑', label: 'Extraction Error', colorClass: 'node-danger' };
    }
    return { icon: '⚙️', label: 'Agent Pipeline Task', colorClass: 'node-slate' };
  };

  return (
    <div className="agent-timeline-wrapper">
      {/* 1. Header with overall pipeline processing state badge */}
      <div className="agent-timeline-header">
        <h5 className="timeline-title">Agentic Workflow Trajectory</h5>
        <span className={`status-badge status-${status.toLowerCase()}`}>
          {status.replace('_', ' ')}
        </span>
      </div>

      {/* 2. Scrollable logs listing */}
      {logs.length === 0 ? (
        <div className="timeline-empty">
          <p>No agent logs registered yet. Trigger reprocessing to run the graph.</p>
        </div>
      ) : (
        <div className="timeline-nodes-container">
          {logs.map((log, index) => {
            const config = getStepConfig(log.step);
            return (
              <div key={index} className="timeline-node-row">
                {/* Visual node track containing the vertical line and dot indicator */}
                <div className="timeline-track">
                  <div className={`timeline-dot ${config.colorClass}`}>
                    <span className="node-icon-span">{config.icon}</span>
                  </div>
                  {index < logs.length - 1 && <div className="timeline-connector-line"></div>}
                </div>

                {/* Log message content bubble */}
                <div className="timeline-content-bubble">
                  <div className="bubble-header">
                    <span className="bubble-step-title">{config.label}</span>
                    <span className="bubble-timestamp">
                      {new Date(log.timestamp).toLocaleTimeString()}
                    </span>
                  </div>
                  <p className="bubble-message-text">{log.message}</p>
                </div>
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
}

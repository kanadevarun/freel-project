import React from 'react';
import './SharedAI.css';

export default function AIConfidenceIndicator({ confidence = 'HIGH' }) {
  const norm = String(confidence).toUpperCase();
  let confClass = 'sai-confidence--high';
  if (norm === 'MEDIUM' || norm === 'MODERATE') confClass = 'sai-confidence--medium';
  if (norm === 'LOW') confClass = 'sai-confidence--low';

  return (
    <span className={`sai-confidence ${confClass}`}>
      <span>Confidence:</span>
      <strong>{norm}</strong>
    </span>
  );
}

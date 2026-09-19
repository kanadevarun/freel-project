import React from 'react';
import { ExternalLink } from 'lucide-react';
import './SharedAI.css';

export default function AISourceReference({ module, entityId, reference, onClick }) {
  if (!reference && !entityId) return null;

  const displayRef = reference || `#${entityId}`;

  return (
    <button
      type="button"
      className="sai-source-ref"
      onClick={onClick}
      title={`Open source record ${displayRef} in ${module || 'module'}`}
    >
      <span>{displayRef}</span>
      <ExternalLink size={10} />
    </button>
  );
}

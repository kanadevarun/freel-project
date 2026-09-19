import React, { useState } from 'react';
import { Hash, Copy, Check } from 'lucide-react';
import './SharedAI.css';

export default function AICorrelationBadge({ correlationId }) {
  const [copied, setCopied] = useState(false);
  if (!correlationId) return null;

  const shortId = correlationId.length > 8 ? correlationId.substring(0, 8) : correlationId;

  const handleCopy = (e) => {
    e.stopPropagation();
    navigator.clipboard.writeText(correlationId);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <span className="sai-correlation-badge" title={`Trace Correlation ID: ${correlationId}`}>
      <Hash className="w-2.5 h-2.5 text-slate-400" />
      <span className="font-mono text-[10px] text-slate-600">{shortId}</span>
      <button
        type="button"
        onClick={handleCopy}
        className="sai-correlation-copy-btn ml-0.5"
        aria-label="Copy Correlation ID"
        title="Copy full correlation ID"
      >
        {copied ? <Check className="w-2.5 h-2.5 text-emerald-600" /> : <Copy className="w-2.5 h-2.5 text-slate-400" />}
      </button>
    </span>
  );
}

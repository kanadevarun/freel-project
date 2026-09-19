import React from 'react';
import { Sparkles } from 'lucide-react';
import './SharedAI.css';

export default function AIMemoryUsageBadge({ isMemoryAssisted, count, label = 'Personalized' }) {
  if (!isMemoryAssisted && !count) return null;

  return (
    <span className="sai-memory-badge" title="Response personalized with your configured AI memory and preferences">
      <Sparkles className="w-2.5 h-2.5 text-indigo-500" />
      <span>{label}</span>
      {count ? <span className="sai-memory-count">{count}</span> : null}
    </span>
  );
}

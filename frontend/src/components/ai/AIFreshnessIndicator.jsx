import React from 'react';
import { Clock } from 'lucide-react';
import './SharedAI.css';

export default function AIFreshnessIndicator({ timestamp, label = 'Analyzed' }) {
  if (!timestamp) return null;

  const dateObj = new Date(timestamp);
  const isValid = !isNaN(dateObj.getTime());
  if (!isValid) return null;

  const timeAgo = formatTimeAgo(dateObj);

  return (
    <span className="sai-freshness-badge" title={dateObj.toLocaleString()} aria-label={`${label} ${timeAgo}`}>
      <Clock className="w-3 h-3 text-slate-400" />
      <span>{label} {timeAgo}</span>
    </span>
  );
}

function formatTimeAgo(date) {
  const seconds = Math.floor((new Date() - date) / 1000);
  if (seconds < 60) return 'just now';
  const minutes = Math.floor(seconds / 60);
  if (minutes < 60) return `${minutes}m ago`;
  const hours = Math.floor(minutes / 60);
  if (hours < 24) return `${hours}h ago`;
  const days = Math.floor(hours / 24);
  return `${days}d ago`;
}

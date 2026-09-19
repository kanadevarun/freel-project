import React from 'react';
import {
  CheckCircle2,
  Clock,
  AlertTriangle,
  AlertCircle,
  XCircle,
  Radio,
  Slash,
  RefreshCw,
  ShieldCheck,
  Zap,
} from 'lucide-react';

export function StatusBadge({ status, className = '', showIcon = true, size = 'sm' }) {
  const s = String(status || '').trim();
  const lower = s.toLowerCase();

  let styles = 'bg-slate-100 text-slate-700 border-slate-200';
  let Icon = Clock;

  if (
    lower === 'active' ||
    lower === 'completed' ||
    lower === 'healthy' ||
    lower === 'connected' ||
    lower === 'operational' ||
    lower === 'up' ||
    lower === 'verified' ||
    lower === 'good'
  ) {
    styles = 'bg-emerald-50 text-emerald-700 border-emerald-200';
    Icon = CheckCircle2;
  } else if (
    lower === 'pending' ||
    lower === 'in progress' ||
    lower === 'in_progress' ||
    lower === 'onboarding' ||
    lower === 'trial' ||
    lower === 'awaiting approval' ||
    lower === 'awaiting_approval'
  ) {
    styles = 'bg-blue-50 text-blue-700 border-blue-200';
    Icon = Clock;
  } else if (
    lower === 'syncing' ||
    lower === 'processing' ||
    lower === 'refreshing'
  ) {
    styles = 'bg-indigo-50 text-indigo-700 border-indigo-200';
    Icon = RefreshCw;
  } else if (
    lower === 'at risk' ||
    lower === 'at_risk' ||
    lower === 'warning' ||
    lower === 'watch' ||
    lower === 'approaching_limit' ||
    lower === 'needs attention' ||
    lower === 'needs_attention'
  ) {
    styles = 'bg-amber-50 text-amber-700 border-amber-200';
    Icon = AlertTriangle;
  } else if (
    lower === 'critical' ||
    lower === 'failed' ||
    lower === 'error' ||
    lower === 'exceeded' ||
    lower === 'rejected' ||
    lower === 'suspended'
  ) {
    styles = 'bg-rose-50 text-rose-700 border-rose-200';
    Icon = AlertCircle;
  } else if (
    lower === 'disabled' ||
    lower === 'not configured' ||
    lower === 'not_configured' ||
    lower === 'inactive' ||
    lower === 'cancelled'
  ) {
    styles = 'bg-slate-100 text-slate-600 border-slate-200';
    Icon = Slash;
  }

  const sizeClasses =
    size === 'xs'
      ? 'px-2 py-0.5 text-[10px]'
      : size === 'md'
      ? 'px-3 py-1 text-sm'
      : 'px-2.5 py-0.5 text-xs';

  const iconSizes = size === 'xs' ? 'w-2.5 h-2.5' : size === 'md' ? 'w-4 h-4' : 'w-3 h-3';

  return (
    <span
      className={`inline-flex items-center gap-1.5 rounded-full font-semibold border ${styles} ${sizeClasses} ${className}`}
    >
      {showIcon && <Icon className={`${iconSizes} flex-shrink-0`} />}
      <span>{s || 'Unknown'}</span>
    </span>
  );
}

export default StatusBadge;

import React from 'react';
import { Inbox, ArrowRight } from 'lucide-react';

export default function EmptyState({
  icon: Icon = Inbox,
  title = 'No records found',
  description = 'There are currently no items to display in this view.',
  primaryAction,
  secondaryAction,
  compact = false,
  className = '',
}) {
  return (
    <div
      className={`flex flex-col items-center justify-center text-center rounded-xl border border-dashed border-slate-200 bg-white ${
        compact ? 'py-8 px-4' : 'py-14 px-6'
      } ${className}`}
    >
      <div className="w-12 h-12 rounded-xl bg-slate-50 border border-slate-200 flex items-center justify-center text-slate-400 mb-3 shadow-xs">
        <Icon className="w-6 h-6 stroke-[1.75]" />
      </div>
      <h3 className="text-sm font-semibold text-slate-900 tracking-tight">{title}</h3>
      {description && (
        <p className="mt-1 text-xs text-slate-500 max-w-sm leading-relaxed">
          {description}
        </p>
      )}

      {(primaryAction || secondaryAction) && (
        <div className="mt-5 flex flex-wrap items-center justify-center gap-2">
          {secondaryAction && (
            <button
              type="button"
              onClick={secondaryAction.onClick}
              className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg border border-slate-200 bg-white text-xs font-medium text-slate-700 hover:bg-slate-50 transition-colors shadow-xs"
            >
              {secondaryAction.icon && <secondaryAction.icon className="w-3.5 h-3.5 text-slate-500" />}
              <span>{secondaryAction.label}</span>
            </button>
          )}
          {primaryAction && (
            <button
              type="button"
              onClick={primaryAction.onClick}
              className="inline-flex items-center gap-1.5 px-3.5 py-1.5 rounded-lg bg-navy-900 hover:bg-navy-800 text-xs font-medium text-white transition-colors shadow-xs"
            >
              {primaryAction.icon && <primaryAction.icon className="w-3.5 h-3.5 text-white/90" />}
              <span>{primaryAction.label}</span>
              {!primaryAction.icon && <ArrowRight className="w-3.5 h-3.5 text-white/80" />}
            </button>
          )}
        </div>
      )}
    </div>
  );
}

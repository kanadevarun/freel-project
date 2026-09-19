import React from 'react';
import { Link } from 'react-router-dom';
import { ArrowUpRight, ArrowDownRight, ChevronRight } from 'lucide-react';

export function KpiCard({
  title,
  value,
  unit,
  subtext,
  subtitle,
  trend,
  trendDirection, // 'up' | 'down' | 'neutral'
  trendLabel,
  icon: Icon,
  color,
  variant = 'blue',
  to,
  badge,
  className = '',
}) {
  const chosenColor = color || variant || 'blue';

  const colorMap = {
    blue: { bg: 'bg-blue-50', text: 'text-blue-600', border: 'border-blue-100' },
    primary: { bg: 'bg-blue-50', text: 'text-blue-600', border: 'border-blue-100' },
    default: { bg: 'bg-slate-100', text: 'text-slate-700', border: 'border-slate-200' },
    emerald: { bg: 'bg-emerald-50', text: 'text-emerald-600', border: 'border-emerald-100' },
    success: { bg: 'bg-emerald-50', text: 'text-emerald-600', border: 'border-emerald-100' },
    amber: { bg: 'bg-amber-50', text: 'text-amber-600', border: 'border-amber-100' },
    warning: { bg: 'bg-amber-50', text: 'text-amber-600', border: 'border-amber-100' },
    purple: { bg: 'bg-purple-50', text: 'text-purple-600', border: 'border-purple-100' },
    rose: { bg: 'bg-rose-50', text: 'text-rose-600', border: 'border-rose-100' },
    danger: { bg: 'bg-rose-50', text: 'text-rose-600', border: 'border-rose-100' },
    slate: { bg: 'bg-slate-100', text: 'text-slate-600', border: 'border-slate-200' },
  };

  const scheme = colorMap[chosenColor] || colorMap.blue;
  const descriptionText = subtitle || subtext;

  // Normalize trend
  let trendText = null;
  let effectiveTrendDir = trendDirection;
  if (trend) {
    if (typeof trend === 'string') {
      trendText = trend;
    } else if (typeof trend === 'object') {
      trendText = trend.text || trend.label || '';
      effectiveTrendDir = trend.direction || trendDirection || 'up';
    }
  }

  const content = (
    <div
      className={`bg-white rounded-xl border border-slate-200 p-5 shadow-2xs hover:shadow-xs transition-all duration-150 flex flex-col justify-between ${
        to ? 'hover:border-blue-300 group cursor-pointer' : ''
      } ${className}`}
    >
      <div className="flex items-start justify-between mb-3">
        <div className="flex items-center space-x-3">
          {Icon && (
            <div
              className={`w-10 h-10 rounded-lg ${scheme.bg} ${scheme.text} ${scheme.border} border flex items-center justify-center flex-shrink-0`}
            >
              <Icon className="w-5 h-5" />
            </div>
          )}
          <div>
            <span className="text-xs font-semibold text-slate-500 uppercase tracking-wider block">
              {title}
            </span>
            {badge && (
              <div className="mt-0.5">
                {typeof badge === 'string' ? (
                  <span className="text-[10px] font-semibold text-blue-700 bg-blue-50 px-1.5 py-0.5 rounded border border-blue-200">
                    {badge}
                  </span>
                ) : React.isValidElement(badge) ? (
                  badge
                ) : typeof badge === 'object' && badge.label ? (
                  <span className="text-[10px] font-semibold text-slate-700 bg-slate-100 px-1.5 py-0.5 rounded border border-slate-200">
                    {badge.label}
                  </span>
                ) : null}
              </div>
            )}
          </div>
        </div>

        {to && (
          <ChevronRight className="w-4 h-4 text-slate-300 group-hover:text-blue-600 group-hover:translate-x-0.5 transition-all" />
        )}
      </div>

      <div className="space-y-1 mt-1">
        <div className="flex items-baseline gap-2">
          <span className="text-2xl font-bold text-slate-900 tracking-tight">{value}</span>
          {unit && <span className="text-xs font-medium text-slate-500">{unit}</span>}
          {trendText && (
            <span
              className={`inline-flex items-center text-xs font-bold px-1.5 py-0.5 rounded ml-auto ${
                effectiveTrendDir === 'up'
                  ? 'bg-emerald-50 text-emerald-700'
                  : effectiveTrendDir === 'down'
                  ? 'bg-rose-50 text-rose-700'
                  : 'bg-slate-100 text-slate-700'
              }`}
            >
              {effectiveTrendDir === 'up' ? (
                <ArrowUpRight className="w-3 h-3 mr-0.5" />
              ) : effectiveTrendDir === 'down' ? (
                <ArrowDownRight className="w-3 h-3 mr-0.5" />
              ) : null}
              {trendText}
            </span>
          )}
        </div>

        {(descriptionText || trendLabel) && (
          <p className="text-xs text-slate-500">
            {descriptionText}
            {trendLabel && <span className="text-slate-400 ml-1">{trendLabel}</span>}
          </p>
        )}
      </div>
    </div>
  );

  if (to) {
    return <Link to={to}>{content}</Link>;
  }
  return content;
}

export default KpiCard;

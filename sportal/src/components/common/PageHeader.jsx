import React from 'react';
import { Link } from 'react-router-dom';
import { ChevronRight } from 'lucide-react';

export function PageHeader({
  title,
  description,
  badge,
  breadcrumbs = [],
  primaryAction,
  secondaryAction,
  secondaryActions = [],
  children,
  className = '',
}) {
  const combinedSecondary = [
    ...(secondaryAction ? [secondaryAction] : []),
    ...secondaryActions,
  ];

  const normalizedCrumbs = breadcrumbs.filter(
    (c, idx) => !(idx === 0 && (c.label?.toLowerCase() === 'sportal' || c.label?.toLowerCase() === 'home'))
  );

  return (
    <div className={`space-y-4 mb-6 ${className}`}>
      {/* Breadcrumbs */}
      {normalizedCrumbs.length > 0 && (
        <nav aria-label="Breadcrumb" className="flex items-center space-x-2 text-xs text-slate-500">
          <Link to="/" className="hover:text-blue-600 transition-colors">
            SPortal
          </Link>
          {normalizedCrumbs.map((crumb, idx) => (
            <React.Fragment key={idx}>
              <ChevronRight className="w-3.5 h-3.5 text-slate-400 flex-shrink-0" />
              {crumb.to || crumb.href ? (
                <Link to={crumb.to || crumb.href} className="hover:text-blue-600 transition-colors">
                  {crumb.label}
                </Link>
              ) : (
                <span className="font-semibold text-slate-900">{crumb.label}</span>
              )}
            </React.Fragment>
          ))}
        </nav>
      )}

      {/* Main Title & Action Bar */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <div className="flex items-center gap-3">
            <h1 className="text-2xl font-bold text-slate-900 tracking-tight">{title}</h1>
            {badge && (
              <div>
                {typeof badge === 'string' ? (
                  <span className="px-2.5 py-0.5 text-xs font-semibold bg-blue-50 text-blue-700 border border-blue-200 rounded-full">
                    {badge}
                  </span>
                ) : React.isValidElement(badge) ? (
                  badge
                ) : typeof badge === 'object' && badge.label ? (
                  <span
                    className={`px-2.5 py-0.5 text-xs font-semibold rounded-full border ${
                      badge.variant === 'success'
                        ? 'bg-emerald-50 text-emerald-700 border-emerald-200'
                        : badge.variant === 'warning'
                        ? 'bg-amber-50 text-amber-700 border-amber-200'
                        : badge.variant === 'danger'
                        ? 'bg-rose-50 text-rose-700 border-rose-200'
                        : badge.variant === 'primary' || badge.variant === 'info'
                        ? 'bg-blue-50 text-blue-700 border-blue-200'
                        : 'bg-slate-100 text-slate-700 border-slate-200'
                    }`}
                  >
                    {badge.label}
                  </span>
                ) : null}
              </div>
            )}
          </div>
          {description && <p className="text-sm text-slate-500 mt-1">{description}</p>}
        </div>

        {/* Actions */}
        {(primaryAction || combinedSecondary.length > 0) && (
          <div className="flex items-center flex-wrap gap-2.5">
            {combinedSecondary.map((action, idx) => {
              const Icon = action.icon;
              if (action.to) {
                return (
                  <Link
                    key={idx}
                    to={action.to}
                    id={action.id}
                    className="inline-flex items-center gap-2 px-3.5 py-2 text-sm font-semibold text-slate-700 bg-white border border-slate-300 rounded-lg hover:bg-slate-50 hover:text-slate-900 shadow-2xs transition-colors cursor-pointer"
                  >
                    {Icon && <Icon className="w-4 h-4 text-slate-500" />}
                    <span>{action.label}</span>
                  </Link>
                );
              }
              return (
                <button
                  key={idx}
                  id={action.id}
                  type="button"
                  onClick={action.onClick}
                  disabled={action.disabled}
                  className="inline-flex items-center gap-2 px-3.5 py-2 text-sm font-semibold text-slate-700 bg-white border border-slate-300 rounded-lg hover:bg-slate-50 hover:text-slate-900 shadow-2xs transition-colors disabled:opacity-50 disabled:cursor-not-allowed cursor-pointer"
                >
                  {Icon && <Icon className="w-4 h-4 text-slate-500" />}
                  <span>{action.label}</span>
                </button>
              );
            })}

            {primaryAction && (
              primaryAction.to ? (
                <Link
                  to={primaryAction.to}
                  id={primaryAction.id}
                  className="inline-flex items-center gap-2 px-4 py-2 text-sm font-semibold text-white bg-blue-600 rounded-lg hover:bg-blue-700 shadow-xs shadow-blue-600/20 transition-colors cursor-pointer"
                >
                  {primaryAction.icon && <primaryAction.icon className="w-4 h-4" />}
                  <span>{primaryAction.label}</span>
                </Link>
              ) : (
                <button
                  id={primaryAction.id}
                  type="button"
                  onClick={primaryAction.onClick}
                  disabled={primaryAction.disabled}
                  className="inline-flex items-center gap-2 px-4 py-2 text-sm font-semibold text-white bg-blue-600 rounded-lg hover:bg-blue-700 shadow-xs shadow-blue-600/20 transition-colors disabled:opacity-50 disabled:cursor-not-allowed cursor-pointer"
                >
                  {primaryAction.icon && <primaryAction.icon className="w-4 h-4" />}
                  <span>{primaryAction.label}</span>
                </button>
              )
            )}
          </div>
        )}
      </div>

      {/* Optional sub-header content (e.g. tabs or filters) */}
      {children && <div className="pt-2">{children}</div>}
    </div>
  );
}

export default PageHeader;

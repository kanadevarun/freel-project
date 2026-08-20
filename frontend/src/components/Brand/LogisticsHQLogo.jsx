import React from 'react';
import { Link } from 'react-router-dom';
import './LogisticsHQLogo.css';

/**
 * LogisticsHQLogo — Canonical Brand Lockup Component
 *
 * Variants:
 *  - 'sidebar'   : Tailored for dark navy dashboard sidebar (transparent 3D emblem + crisp white/cyan typography)
 *  - 'header'    : For public website glass/light navbar
 *  - 'auth'      : For Login, Signup, and Auth screens
 *  - 'splash'    : For animated first-load splash & loading views
 *  - 'footer'    : For dark footer branding
 *  - 'icon-only' : Single standalone icon
 */
export default function LogisticsHQLogo({
  variant = 'sidebar',
  showText = true,
  subtitle,
  className = '',
  linkTo = null,
  onClick,
  style = {},
  iconSize = null,
}) {
  const isDark = variant === 'sidebar' || variant === 'footer' || variant === 'splash';
  
  // Default subtitles per variant
  const defaultSubtitle = variant === 'sidebar' ? 'LOGISTICS OS' : null;
  const activeSubtitle = subtitle !== undefined ? subtitle : defaultSubtitle;

  const content = (
    <div
      className={`lhq-brand-lockup lhq-brand-${variant} ${isDark ? 'lhq-theme-dark' : 'lhq-theme-light'} ${className}`}
      style={style}
      onClick={onClick}
    >
      <div className="lhq-brand-icon-wrapper">
        <img
          src="/images/logo/logo.png"
          alt="LogisticsHQ"
          className="lhq-brand-icon-img"
          style={iconSize ? { height: `${iconSize}px` } : undefined}
          loading="eager"
        />
      </div>

      {showText && variant !== 'icon-only' && (
        <div className="lhq-brand-text">
          <span className="lhq-brand-title">LogisticsHQ</span>
          {activeSubtitle && (
            <span className="lhq-brand-subtitle">{activeSubtitle}</span>
          )}
        </div>
      )}
    </div>
  );

  if (linkTo) {
    return (
      <Link to={linkTo} className="lhq-brand-link">
        {content}
      </Link>
    );
  }

  return content;
}

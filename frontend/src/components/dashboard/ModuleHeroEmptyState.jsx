import React from 'react';
import PropTypes from 'prop-types';
import './ModuleHeroEmptyState.css';

export default function ModuleHeroEmptyState({
  icon,
  badgeTheme = 'blue',
  title,
  description,
  primaryAction,
  secondaryAction,
  features = [],
}) {
  return (
    <div className="mhes-hero-container">
      {/* ── Top Hero Visual & Actions ── */}
      <div className="mhes-hero-content">
        {/* Animated Sonar Radar Pulse */}
        <div className={`mhes-radar-wrap theme-${badgeTheme}`}>
          <div className="mhes-radar-ring ring-3" />
          <div className="mhes-radar-ring ring-2" />
          <div className="mhes-radar-ring ring-1" />
          <div className="mhes-radar-sweep" />
          <div className="mhes-icon-box">
            {icon}
          </div>
        </div>

        <h2 className="mhes-title">{title}</h2>
        <p className="mhes-desc">{description}</p>

        {(primaryAction || secondaryAction) && (
          <div className="mhes-actions-row">
            {primaryAction && (
              <button
                className="mhes-btn-primary"
                onClick={primaryAction.onClick}
              >
                {primaryAction.icon}
                <span>{primaryAction.label}</span>
              </button>
            )}
            {secondaryAction && (
              <button
                className="mhes-btn-secondary"
                onClick={secondaryAction.onClick}
              >
                {secondaryAction.icon}
                <span>{secondaryAction.label}</span>
              </button>
            )}
          </div>
        )}
      </div>

      {/* ── 3 Feature Highlight Preview Cards ── */}
      {features && features.length > 0 && (
        <div className="mhes-features-grid">
          {features.map((feat, idx) => (
            <div key={idx} className="mhes-feat-card">
              <div
                className="mhes-feat-icon"
                style={{
                  background: feat.iconBg || '#eff6ff',
                  color: feat.iconColor || '#2563eb',
                }}
              >
                {feat.icon}
              </div>
              <div className="mhes-feat-text">
                <h4>{feat.title}</h4>
                <p>{feat.desc}</p>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}

ModuleHeroEmptyState.propTypes = {
  icon: PropTypes.node.isRequired,
  badgeTheme: PropTypes.string,
  title: PropTypes.string.isRequired,
  description: PropTypes.string.isRequired,
  primaryAction: PropTypes.shape({
    label: PropTypes.string.isRequired,
    onClick: PropTypes.func.isRequired,
    icon: PropTypes.node,
  }),
  secondaryAction: PropTypes.shape({
    label: PropTypes.string.isRequired,
    onClick: PropTypes.func.isRequired,
    icon: PropTypes.node,
  }),
  features: PropTypes.arrayOf(
    PropTypes.shape({
      icon: PropTypes.node.isRequired,
      iconBg: PropTypes.string,
      iconColor: PropTypes.string,
      title: PropTypes.string.isRequired,
      desc: PropTypes.string.isRequired,
    })
  ),
};

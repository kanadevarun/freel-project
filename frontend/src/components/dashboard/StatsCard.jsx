import PropTypes from 'prop-types';
import './StatsCard.css';

export default function StatsCard({ title, value, trend, trendValue, icon, isLoading }) {
  if (isLoading) {
    return (
      <div className="stats-card skeleton">
        <div className="stats-card-skeleton-bg"></div>
      </div>
    );
  }

  const isPositive = trend === 'up';
  const isNegative = trend === 'down';
  const isNeutral = trend === 'neutral';

  return (
    <div className="stats-card">
      <div className="stats-header">
        <span className="stats-title">{title}</span>
        {icon && <span className="stats-icon">{icon}</span>}
      </div>
      <div className="stats-body">
        <div className="stats-value">{value}</div>
        {trend && (
          <div className={`stats-trend ${isPositive ? 'positive' : ''} ${isNegative ? 'negative' : ''} ${isNeutral ? 'neutral' : ''}`}>
            {isPositive && '↑'}
            {isNegative && '↓'}
            {isNeutral && '—'}
            <span className="trend-value">{trendValue}</span>
          </div>
        )}
      </div>
    </div>
  );
}

StatsCard.propTypes = {
  title: PropTypes.string.isRequired,
  value: PropTypes.oneOfType([PropTypes.string, PropTypes.number]).isRequired,
  trend: PropTypes.oneOf(['up', 'down', 'neutral']),
  trendValue: PropTypes.string,
  icon: PropTypes.node,
  isLoading: PropTypes.bool,
};

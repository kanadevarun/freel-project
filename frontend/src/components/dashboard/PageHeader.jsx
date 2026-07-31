import PropTypes from 'prop-types';
import './PageHeader.css';

export default function PageHeader({ title, description, actionLabel, onAction, actionIcon }) {
  const currentDate = new Date().toLocaleDateString('en-US', {
    weekday: 'long',
    month: 'long',
    day: 'numeric'
  });

  return (
    <div className="page-header">
      <div className="page-header-content">
        <h1 className="page-title">{title}</h1>
        <p className="page-description">
          {description || currentDate}
        </p>
      </div>
      {actionLabel && (
        <div className="page-header-actions">
          <button className="btn-primary" onClick={onAction}>
            {actionIcon && <span className="btn-icon">{actionIcon}</span>}
            {actionLabel}
          </button>
        </div>
      )}
    </div>
  );
}

PageHeader.propTypes = {
  title: PropTypes.string.isRequired,
  description: PropTypes.string,
  actionLabel: PropTypes.string,
  onAction: PropTypes.func,
  actionIcon: PropTypes.node,
};

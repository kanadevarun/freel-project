import PropTypes from 'prop-types';
import './RoleBadge.css';

const ROLE_MAP = {
  'SUPER_ADMIN': { label: 'Super Admin', type: 'admin' },
  'ADMIN': { label: 'Admin', type: 'admin' },
  'SALES': { label: 'Sales', type: 'sales' },
  'OPERATIONS': { label: 'Operations', type: 'ops' },
  'FINANCE': { label: 'Finance', type: 'finance' },
  'VIEWER': { label: 'Viewer', type: 'viewer' },
};

export default function RoleBadge({ role }) {
  const roleName = typeof role === 'object' ? role.name : role;
  const config = ROLE_MAP[roleName] || { label: roleName || 'Unknown', type: 'viewer' };

  return (
    <span className={`role-badge role-${config.type}`}>
      {config.label}
    </span>
  );
}

RoleBadge.propTypes = {
  role: PropTypes.oneOfType([
    PropTypes.string,
    PropTypes.shape({
      name: PropTypes.string,
    })
  ]).isRequired,
};

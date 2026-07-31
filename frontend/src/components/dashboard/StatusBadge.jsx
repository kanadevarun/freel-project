import PropTypes from 'prop-types';
import './StatusBadge.css';

export const STATUS_TYPE = {
  NEUTRAL: 'neutral',
  WARNING: 'warning',
  INFO: 'info',
  SUCCESS: 'success',
  DANGER: 'danger',
  PRIMARY: 'primary',
};

export const STATUS = {
  // RFQ Statuses
  DRAFT: 'DRAFT',
  PENDING: 'PENDING',
  QUOTED: 'QUOTED',
  ACCEPTED: 'ACCEPTED',
  REJECTED: 'REJECTED',
  
  // Shipment Statuses
  BOOKED: 'BOOKED',
  IN_TRANSIT: 'IN_TRANSIT',
  ARRIVED: 'ARRIVED',
  CUSTOMS: 'CUSTOMS',
  DELIVERED: 'DELIVERED',
  
  // Generic
  ACTIVE: 'ACTIVE',
  INACTIVE: 'INACTIVE',
};

export const STATUS_MAP = {
  [STATUS.DRAFT]: { label: 'Draft', type: STATUS_TYPE.NEUTRAL },
  [STATUS.PENDING]: { label: 'Pending', type: STATUS_TYPE.WARNING },
  [STATUS.QUOTED]: { label: 'Quoted', type: STATUS_TYPE.INFO },
  [STATUS.ACCEPTED]: { label: 'Accepted', type: STATUS_TYPE.SUCCESS },
  [STATUS.REJECTED]: { label: 'Rejected', type: STATUS_TYPE.DANGER },
  
  [STATUS.BOOKED]: { label: 'Booked', type: STATUS_TYPE.INFO },
  [STATUS.IN_TRANSIT]: { label: 'In Transit', type: STATUS_TYPE.PRIMARY },
  [STATUS.ARRIVED]: { label: 'Arrived', type: STATUS_TYPE.SUCCESS },
  [STATUS.CUSTOMS]: { label: 'Customs', type: STATUS_TYPE.WARNING },
  [STATUS.DELIVERED]: { label: 'Delivered', type: STATUS_TYPE.SUCCESS },
  
  [STATUS.ACTIVE]: { label: 'Active', type: STATUS_TYPE.SUCCESS },
  [STATUS.INACTIVE]: { label: 'Inactive', type: STATUS_TYPE.NEUTRAL },
};

export default function StatusBadge({ status, customLabel, customType }) {
  const config = STATUS_MAP[status] || { label: status, type: STATUS_TYPE.NEUTRAL };
  
  const label = customLabel || config.label;
  const type = customType || config.type;

  return (
    <span className={`status-badge badge-${type}`}>
      {label}
    </span>
  );
}

StatusBadge.propTypes = {
  status: PropTypes.string.isRequired,
  customLabel: PropTypes.string,
  customType: PropTypes.oneOf(Object.values(STATUS_TYPE)),
};

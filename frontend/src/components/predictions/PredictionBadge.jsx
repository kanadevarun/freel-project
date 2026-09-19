import React from 'react';
import PropTypes from 'prop-types';

export const SeverityBadge = ({ severity }) => {
  const sevKey = (severity || 'MEDIUM').toUpperCase();
  const styles = {
    CRITICAL: 'bg-red-50 text-red-700 border-red-200',
    HIGH: 'bg-amber-50 text-amber-700 border-amber-200',
    MEDIUM: 'bg-blue-50 text-blue-700 border-blue-200',
    LOW: 'bg-slate-50 text-slate-700 border-slate-200',
    SAFE: 'bg-emerald-50 text-emerald-700 border-emerald-200',
    INSUFFICIENT_DATA: 'bg-amber-50 text-amber-800 border-amber-200',
  };

  const currentStyle = styles[sevKey] || styles.MEDIUM;
  const displayLabel = sevKey === 'INSUFFICIENT_DATA' ? 'INSUFFICIENT DATA' : sevKey;

  return (
    <span
      className={`inline-flex items-center px-2 py-0.5 rounded text-xs font-semibold border ${currentStyle}`}
      title={`Risk Severity: ${displayLabel}`}
    >
      {displayLabel}
    </span>
  );
};

SeverityBadge.propTypes = {
  severity: PropTypes.string,
};

export const ConfidenceBadge = ({ score, band, insufficientData = false }) => {
  const isInsufficient = insufficientData || band === 'INSUFFICIENT_DATA' || (score === 0 && (!band || band === 'LOW'));
  if (isInsufficient) {
    return (
      <span
        className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium border bg-amber-50 text-amber-800 border-amber-200"
        title="Insufficient historical records available for confidence calculation"
      >
        Insufficient Data
      </span>
    );
  }

  const pct = Math.round((score || 0) * 100);
  const colorStyle =
    pct >= 80
      ? 'bg-emerald-50 text-emerald-700 border-emerald-200'
      : pct >= 50
      ? 'bg-blue-50 text-blue-700 border-blue-200'
      : 'bg-slate-50 text-slate-600 border-slate-200';

  const bandLabel = band || (pct >= 80 ? 'HIGH' : pct >= 50 ? 'MEDIUM' : 'LOW');

  return (
    <span
      className={`inline-flex items-center px-2 py-0.5 rounded text-xs font-medium border ${colorStyle}`}
      title={`Forecast Confidence: ${pct}% certainty (${bandLabel} band)`}
    >
      {pct}% ({bandLabel})
    </span>
  );
};

ConfidenceBadge.propTypes = {
  score: PropTypes.number,
  band: PropTypes.string,
  insufficientData: PropTypes.bool,
};

export const ValueTypeBadge = ({ isForecast = true, type = null }) => {
  let label = isForecast ? 'PREDICTED' : 'AUTHORITATIVE';
  let badgeStyle = isForecast ? 'bg-purple-50 text-purple-700 border-purple-200' : 'bg-slate-100 text-slate-800 border-slate-300';
  let title = isForecast ? 'Forward-looking AI transit/operational prediction' : 'Authoritative master business fact';

  if (type === 'RECOMMENDED') {
    label = 'RECOMMENDED';
    badgeStyle = 'bg-blue-50 text-blue-700 border-blue-200';
    title = 'Advisory action recommendation (Human-in-the-Loop review required)';
  }

  return (
    <span
      className={`inline-flex items-center px-1.5 py-0.5 rounded text-[10px] font-bold tracking-wider uppercase border ${badgeStyle}`}
      title={title}
    >
      {label}
    </span>
  );
};

ValueTypeBadge.propTypes = {
  isForecast: PropTypes.bool,
  type: PropTypes.string,
};

export const ReviewStatusBadge = ({ status }) => {
  const s = (status || 'UNREVIEWED').toUpperCase();
  const styles = {
    UNREVIEWED: 'bg-slate-50 text-slate-600 border-slate-200',
    ACKNOWLEDGED: 'bg-blue-50 text-blue-700 border-blue-200',
    AWAITING_APPROVAL: 'bg-amber-50 text-amber-800 border-amber-300',
    ACTION_REQUESTED: 'bg-amber-50 text-amber-800 border-amber-300',
    APPROVED: 'bg-emerald-50 text-emerald-700 border-emerald-200',
    DISMISSED: 'bg-slate-100 text-slate-500 border-slate-200',
    ACTION_TAKEN: 'bg-emerald-50 text-emerald-700 border-emerald-200',
    INSUFFICIENT_DATA: 'bg-amber-50 text-amber-800 border-amber-200',
  };

  const labels = {
    UNREVIEWED: 'Unreviewed',
    ACKNOWLEDGED: 'Acknowledged',
    AWAITING_APPROVAL: 'Awaiting Approval',
    ACTION_REQUESTED: 'Action Queued',
    APPROVED: 'Approved',
    DISMISSED: 'Dismissed',
    ACTION_TAKEN: 'Action Executed',
    INSUFFICIENT_DATA: 'Insufficient Data',
  };

  return (
    <span
      className={`inline-flex items-center px-2 py-0.5 rounded text-xs font-medium border ${styles[s] || styles.UNREVIEWED}`}
      title={`Review Status: ${labels[s] || s}`}
    >
      {labels[s] || s}
    </span>
  );
};

ReviewStatusBadge.propTypes = {
  status: PropTypes.string,
};

export default {
  SeverityBadge,
  ConfidenceBadge,
  ValueTypeBadge,
  ReviewStatusBadge,
};

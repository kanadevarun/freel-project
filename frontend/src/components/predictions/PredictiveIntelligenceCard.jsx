import React, { useState } from 'react';
import PropTypes from 'prop-types';
import { SeverityBadge, ConfidenceBadge, ValueTypeBadge } from './PredictionBadge';
import {
  AlertTriangle,
  CheckCircle2,
  Clock,
  ExternalLink,
  ShieldCheck,
  ChevronDown,
  ChevronUp,
  Send,
  XCircle,
} from 'lucide-react';

export const PredictiveIntelligenceCard = ({
  prediction,
  onAcknowledge,
  onDismiss,
  onRequestAction,
}) => {
  const [showDetails, setShowDetails] = useState(false);
  const [dismissReason, setDismissReason] = useState('');
  const [isDismissing, setIsDismissing] = useState(false);
  const [loading, setLoading] = useState(false);

  if (!prediction) return null;

  const handleAction = async (actionFn, ...args) => {
    try {
      setLoading(true);
      await actionFn(...args);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="bg-white rounded-lg border border-slate-200 shadow-sm hover:border-slate-300 transition-colors p-4 mb-3">
      {/* Top Header Row */}
      <div className="flex flex-wrap items-center justify-between gap-2 mb-2">
        <div className="flex items-center gap-2">
          <ValueTypeBadge isForecast={true} />
          <SeverityBadge severity={prediction.severity} />
          <ConfidenceBadge
            score={prediction.confidence_score}
            band={prediction.confidence_band}
          />
          <span className="text-xs text-slate-500 font-mono">
            {prediction.prediction_id}
          </span>
        </div>
        <div className="flex items-center gap-2 text-xs text-slate-500">
          <span className="inline-flex items-center gap-1">
            <Clock className="w-3.5 h-3.5 text-slate-400" />
            Horizon: {prediction.time_horizon || '7_DAYS'}
          </span>
          <span className="px-2 py-0.5 rounded bg-slate-100 text-slate-700 font-medium">
            {prediction.status}
          </span>
        </div>
      </div>

      {/* Prediction Statement */}
      <div className="mb-2">
        <h4 className="text-sm font-semibold text-slate-900 leading-snug">
          {prediction.prediction_statement}
        </h4>
        {prediction.predicted_value && (
          <p className="text-xs font-mono font-medium text-slate-600 mt-1">
            Predicted Variance:{' '}
            <span className="text-slate-900 font-bold">
              {prediction.predicted_value}
            </span>
          </p>
        )}
      </div>

      {/* Explanation Rationale */}
      <p className="text-xs text-slate-600 mb-3 line-clamp-2">
        {prediction.explanation}
      </p>

      {/* Recommended Action Box */}
      {prediction.recommended_action && (
        <div className="bg-slate-50 border border-slate-200 rounded p-2.5 mb-3 flex items-start gap-2">
          <ShieldCheck className="w-4 h-4 text-blue-600 mt-0.5 shrink-0" />
          <div className="flex-1 min-w-0">
            <p className="text-xs font-medium text-slate-800">
              <span className="font-semibold text-blue-900">
                Recommended Action:
              </span>{' '}
              {prediction.recommended_action}
            </p>
            {prediction.requires_approval && (
              <span className="inline-block mt-1 text-[11px] font-semibold text-amber-700 bg-amber-50 px-1.5 py-0.2 rounded border border-amber-200">
                Requires Human-in-the-Loop Approval
              </span>
            )}
          </div>
        </div>
      )}

      {/* Collapsible Source Evidence & Signals */}
      {showDetails && (
        <div className="border-t border-slate-100 pt-2.5 mt-2 mb-3 space-y-2 text-xs">
          {prediction.source_references && prediction.source_references.length > 0 && (
            <div>
              <span className="font-semibold text-slate-700 block mb-1">
                Authoritative Source Grounding:
              </span>
              <ul className="list-disc pl-4 text-slate-600 space-y-1">
                {prediction.source_references.map((src, idx) => (
                  <li key={idx}>
                    <span className="font-mono text-slate-800">
                      {src.source_module} #{src.source_record_id}
                    </span>{' '}
                    ({src.source_field}) at{' '}
                    <span className="text-slate-500">
                      {new Date(src.source_timestamp).toLocaleTimeString([], {
                        month: 'short',
                        day: 'numeric',
                        hour: '2-digit',
                        minute: '2-digit',
                      })}
                    </span>
                    {src.document_reference && ` — Document: ${src.document_reference}`}
                  </li>
                ))}
              </ul>
            </div>
          )}

          {prediction.supporting_signals && prediction.supporting_signals.length > 0 && (
            <div>
              <span className="font-semibold text-slate-700 block mb-1">
                Telemetry Signals:
              </span>
              <div className="grid grid-cols-2 gap-2">
                {prediction.supporting_signals.map((sig, idx) => (
                  <div key={idx} className="bg-slate-50 p-1.5 rounded border border-slate-200">
                    <span className="text-[11px] text-slate-500 block truncate">
                      {sig.signal_name}
                    </span>
                    <span className="font-mono font-medium text-slate-800 text-xs">
                      {String(sig.observed_value)}
                    </span>
                  </div>
                ))}
              </div>
            </div>
          )}
        </div>
      )}

      {/* Footer Controls */}
      <div className="flex flex-wrap items-center justify-between gap-2 pt-2 border-t border-slate-100">
        <button
          type="button"
          onClick={() => setShowDetails(!showDetails)}
          className="inline-flex items-center gap-1 text-xs text-blue-600 hover:text-blue-800 font-medium py-1 px-1.5 rounded hover:bg-blue-50 transition-colors"
        >
          {showDetails ? (
            <>
              <ChevronUp className="w-3.5 h-3.5" /> Hide Evidence
            </>
          ) : (
            <>
              <ChevronDown className="w-3.5 h-3.5" /> View Sources (
              {prediction.source_references?.length || 0})
            </>
          )}
        </button>

        <div className="flex items-center gap-2">
          {prediction.status === 'PUBLISHED' && onAcknowledge && (
            <button
              type="button"
              disabled={loading}
              onClick={() => handleAction(onAcknowledge, prediction.prediction_id)}
              className="inline-flex items-center gap-1 text-xs px-2.5 py-1 rounded font-medium bg-slate-100 hover:bg-slate-200 text-slate-700 transition-colors"
            >
              <CheckCircle2 className="w-3.5 h-3.5 text-emerald-600" />
              Acknowledge
            </button>
          )}

          {prediction.action_type &&
            prediction.status !== 'ACTION_REQUESTED' &&
            prediction.status !== 'AWAITING_APPROVAL' &&
            onRequestAction && (
              <button
                type="button"
                disabled={loading}
                onClick={() =>
                  handleAction(onRequestAction, prediction.prediction_id, 'Approved by operator')
                }
                className="inline-flex items-center gap-1 text-xs px-2.5 py-1 rounded font-medium bg-blue-600 hover:bg-blue-700 text-white transition-colors"
              >
                <Send className="w-3.5 h-3.5" />
                Request Action
              </button>
            )}

          {prediction.status !== 'DISMISSED' && onDismiss && (
            <>
              {!isDismissing ? (
                <button
                  type="button"
                  disabled={loading}
                  onClick={() => setIsDismissing(true)}
                  className="inline-flex items-center gap-1 text-xs px-2 py-1 rounded text-slate-500 hover:text-slate-700 hover:bg-slate-100 transition-colors"
                >
                  <XCircle className="w-3.5 h-3.5 text-slate-400" />
                  Dismiss
                </button>
              ) : (
                <div className="flex items-center gap-1">
                  <input
                    type="text"
                    placeholder="Reason..."
                    value={dismissReason}
                    onChange={(e) => setDismissReason(e.target.value)}
                    className="text-xs border border-slate-300 rounded px-2 py-1 w-28 focus:outline-blue-500"
                  />
                  <button
                    type="button"
                    onClick={() => {
                      handleAction(onDismiss, prediction.prediction_id, dismissReason);
                      setIsDismissing(false);
                    }}
                    className="text-xs bg-slate-700 text-white px-2 py-1 rounded hover:bg-slate-800"
                  >
                    Confirm
                  </button>
                  <button
                    type="button"
                    onClick={() => setIsDismissing(false)}
                    className="text-xs text-slate-500 px-1 hover:text-slate-700"
                  >
                    Cancel
                  </button>
                </div>
              )}
            </>
          )}
        </div>
      </div>
    </div>
  );
};

PredictiveIntelligenceCard.propTypes = {
  prediction: PropTypes.object.isRequired,
  onAcknowledge: PropTypes.func,
  onDismiss: PropTypes.func,
  onRequestAction: PropTypes.func,
};

export default PredictiveIntelligenceCard;

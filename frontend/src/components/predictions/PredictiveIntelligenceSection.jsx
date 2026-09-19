import React, { useState, useEffect, useCallback } from 'react';
import PropTypes from 'prop-types';
import { predictionService } from '../../services/predictionService';
import PredictiveIntelligenceCard from './PredictiveIntelligenceCard';
import {
  Sparkles,
  RefreshCw,
  Filter,
  AlertCircle,
  TrendingUp,
} from 'lucide-react';

export const PredictiveIntelligenceSection = ({
  module = '',
  relatedRecordType = '',
  relatedRecordId = '',
  title = 'Predictive Intelligence & Decision Support',
  subtitle = 'Source-grounded AI forecasts and recommended interventions',
  className = '',
}) => {
  const [predictions, setPredictions] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [selectedSeverity, setSelectedSeverity] = useState('ALL');

  const fetchPredictions = useCallback(async () => {
    try {
      setLoading(true);
      setError(null);
      const params = {};
      if (module) params.module = module;
      if (relatedRecordType) params.related_record_type = relatedRecordType;
      if (relatedRecordId) params.related_record_id = relatedRecordId;
      if (selectedSeverity !== 'ALL') params.severity = selectedSeverity;

      const data = await predictionService.listPredictions(params);
      const items = Array.isArray(data) ? data : (data?.predictions || data?.data?.predictions || []);
      setPredictions(items);
    } catch (err) {
      console.error('Failed to load predictions:', err);
      setError(err.response?.data?.message || err.message || 'Failed to load predictive intelligence');
    } finally {
      setLoading(false);
    }
  }, [module, relatedRecordType, relatedRecordId, selectedSeverity]);

  useEffect(() => {
    fetchPredictions();
  }, [fetchPredictions]);

  const handleAcknowledge = async (predictionId) => {
    await predictionService.acknowledgePrediction(predictionId);
    setPredictions((prev) =>
      prev.map((p) =>
        p.prediction_id === predictionId ? { ...p, status: 'ACKNOWLEDGED' } : p
      )
    );
  };

  const handleDismiss = async (predictionId, reason) => {
    await predictionService.dismissPrediction(predictionId, reason);
    setPredictions((prev) =>
      prev.map((p) =>
        p.prediction_id === predictionId ? { ...p, status: 'DISMISSED' } : p
      )
    );
  };

  const handleRequestAction = async (predictionId, notes) => {
    await predictionService.requestAction(predictionId, notes);
    setPredictions((prev) =>
      prev.map((p) =>
        p.prediction_id === predictionId
          ? {
              ...p,
              status: p.requires_approval ? 'AWAITING_APPROVAL' : 'ACTION_REQUESTED',
            }
          : p
      )
    );
  };

  return (
    <div
      className={`bg-slate-50/70 border border-slate-200 rounded-xl p-4 md:p-5 ${className}`}
      data-testid="predictive-intelligence-section"
    >
      {/* Header */}
      <div className="flex flex-wrap items-center justify-between gap-3 mb-4">
        <div className="flex items-center gap-2">
          <div className="w-8 h-8 rounded-lg bg-blue-100 flex items-center justify-center text-blue-700">
            <Sparkles className="w-4 h-4" />
          </div>
          <div>
            <h3 className="text-sm font-semibold text-slate-900 flex items-center gap-2">
              {title}
              <span className="text-[10px] uppercase font-bold tracking-wider bg-blue-50 text-blue-700 border border-blue-200 px-1.5 py-0.5 rounded">
                Phase 4 Foundation
              </span>
            </h3>
            <p className="text-xs text-slate-500">{subtitle}</p>
          </div>
        </div>

        {/* Severity Filter & Refresh */}
        <div className="flex items-center gap-2">
          <div className="flex items-center bg-white border border-slate-200 rounded-lg p-0.5 text-xs font-medium">
            {['ALL', 'CRITICAL', 'HIGH', 'MEDIUM'].map((sev) => (
              <button
                key={sev}
                type="button"
                onClick={() => setSelectedSeverity(sev)}
                className={`px-2 py-1 rounded transition-colors ${
                  selectedSeverity === sev
                    ? 'bg-blue-600 text-white font-semibold shadow-xs'
                    : 'text-slate-600 hover:text-slate-900 hover:bg-slate-50'
                }`}
              >
                {sev}
              </button>
            ))}
          </div>

          <button
            type="button"
            disabled={loading}
            onClick={fetchPredictions}
            className="p-1.5 rounded-lg border border-slate-200 bg-white hover:bg-slate-50 text-slate-600 hover:text-slate-900 transition-colors"
            title="Refresh Predictions"
          >
            <RefreshCw className={`w-4 h-4 ${loading ? 'animate-spin' : ''}`} />
          </button>
        </div>
      </div>

      {/* Content Area */}
      {loading ? (
        <div className="space-y-3">
          {[1, 2].map((i) => (
            <div
              key={i}
              className="bg-white border border-slate-200 rounded-lg p-4 animate-pulse space-y-2"
            >
              <div className="h-4 bg-slate-200 rounded w-1/3" />
              <div className="h-3 bg-slate-100 rounded w-3/4" />
              <div className="h-3 bg-slate-100 rounded w-1/2" />
            </div>
          ))}
        </div>
      ) : error ? (
        <div className="bg-red-50 border border-red-200 rounded-lg p-3 text-xs text-red-700 flex items-center justify-between">
          <div className="flex items-center gap-2">
            <AlertCircle className="w-4 h-4 text-red-500 shrink-0" />
            <span>{error}</span>
          </div>
          <button
            type="button"
            onClick={fetchPredictions}
            className="font-medium underline hover:text-red-900"
          >
            Retry
          </button>
        </div>
      ) : predictions.length === 0 ? (
        <div className="bg-white border border-dashed border-slate-200 rounded-lg p-6 text-center text-xs text-slate-500">
          <TrendingUp className="w-8 h-8 text-slate-300 mx-auto mb-2" />
          <p className="font-medium text-slate-700">No active predictive alerts</p>
          <p className="mt-1 text-slate-500">
            Authoritative milestones and ledger records show zero negative variance for the selected filters.
          </p>
        </div>
      ) : (
        <div className="space-y-2">
          {predictions.map((pred) => (
            <PredictiveIntelligenceCard
              key={pred.prediction_id}
              prediction={pred}
              onAcknowledge={handleAcknowledge}
              onDismiss={handleDismiss}
              onRequestAction={handleRequestAction}
            />
          ))}
        </div>
      )}
    </div>
  );
};

PredictiveIntelligenceSection.propTypes = {
  module: PropTypes.string,
  relatedRecordType: PropTypes.string,
  relatedRecordId: PropTypes.string,
  title: PropTypes.string,
  subtitle: PropTypes.string,
  className: PropTypes.string,
};

export default PredictiveIntelligenceSection;

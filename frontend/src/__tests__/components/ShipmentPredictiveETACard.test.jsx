import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import ShipmentPredictiveETACard from '../../components/predictions/ShipmentPredictiveETACard';
import * as predictionService from '../../services/predictionService';

const mockGetShipmentPredictedETA = vi.fn();
const mockRefreshShipmentPredictedETA = vi.fn();

vi.mock('../../services/predictionService', () => {
  const service = {
    getShipmentPredictedETA: (...args) => mockGetShipmentPredictedETA(...args),
    refreshShipmentPredictedETA: (...args) => mockRefreshShipmentPredictedETA(...args),
  };
  return {
    default: service,
    predictionService: service,
    getShipmentPredictedETA: service.getShipmentPredictedETA,
    refreshShipmentPredictedETA: service.refreshShipmentPredictedETA,
  };
});

describe('ShipmentPredictiveETACard', () => {
  const mockPredictionData = {
    prediction_type: 'shipment_eta',
    target_id: '101',
    severity: 'LOW',
    predicted_arrival_window: ['2026-09-15T18:00:00Z', '2026-09-15T22:00:00Z'],
    predicted_delay_hours: 0,
    risk_level: 'LOW',
    confidence_score: 0.95,
    confidence_band: 'HIGH',
    explanation: 'Milestone tracking indicates on-time schedule.',
    insufficient_data: false,
    supporting_signals: [
      { signal_name: 'historical_port_turnaround', observed_value: '14.2h' },
      { signal_name: 'milestone_variance', observed_value: '0.0h' }
    ],
    source_references: [
      { source_module: 'shipment_milestones', source_field: 'actual_at', source_record_id: '991' }
    ],
    recommended_action: 'None required. Continue standard monitoring.',
    action_type: 'none',
    action_requires_approval: false,
    data_freshness_seconds: 120,
    model_version: 'v4.2-eta-rules-llm'
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders loading skeleton initially', () => {
    mockGetShipmentPredictedETA.mockReturnValue(new Promise(() => {}));
    render(<ShipmentPredictiveETACard shipmentId="101" authoritativeEta="2026-09-15T18:00:00Z" />);
    expect(screen.getByText(/Synthesizing real-time vessel telemetry/i)).toBeInTheDocument();
  });

  it('renders authoritative and predicted ETA separately with risk level badge', async () => {
    mockGetShipmentPredictedETA.mockResolvedValue({
      success: true,
      data: mockPredictionData
    });

    render(<ShipmentPredictiveETACard shipmentId="101" authoritativeEta="2026-09-15T18:00:00Z" />);

    await waitFor(() => {
      expect(screen.getByText(/Authoritative Master ETA/i)).toBeInTheDocument();
      expect(screen.getByText(/Predicted Arrival Window/i)).toBeInTheDocument();
      expect(screen.getByText(/LOW RISK/i)).toBeInTheDocument();
      expect(screen.getByText(/95% Certainty/i)).toBeInTheDocument();
    });
  });

  it('displays insufficient data warning when data is unavailable', async () => {
    mockGetShipmentPredictedETA.mockResolvedValue({
      success: true,
      data: {
        ...mockPredictionData,
        insufficient_data: true,
        severity: 'INSUFFICIENT_DATA',
        explanation: 'Insufficient milestone history to calculate accurate forecast window.'
      }
    });

    render(<ShipmentPredictiveETACard shipmentId="104" authoritativeEta={null} />);

    await waitFor(() => {
      expect(screen.getByText(/Insufficient Schedule Telemetry/i)).toBeInTheDocument();
      expect(screen.getByText(/Insufficient milestone history/i)).toBeInTheDocument();
    });
  });

  it('toggles collapsible source evidence drawer', async () => {
    mockGetShipmentPredictedETA.mockResolvedValue({
      success: true,
      data: mockPredictionData
    });

    render(<ShipmentPredictiveETACard shipmentId="101" authoritativeEta="2026-09-15T18:00:00Z" />);

    await waitFor(() => {
      expect(screen.getByText(/View Supporting Signals & Source Citations/i)).toBeInTheDocument();
    });

    const toggleBtn = screen.getByText(/View Supporting Signals & Source Citations/i);
    fireEvent.click(toggleBtn);

    expect(screen.getByText(/historical port turnaround/i)).toBeInTheDocument();
    expect(screen.getByText(/Record #991/i)).toBeInTheDocument();
  });

  it('handles refresh button click', async () => {
    mockGetShipmentPredictedETA.mockResolvedValue({
      success: true,
      data: mockPredictionData
    });
    mockRefreshShipmentPredictedETA.mockResolvedValue({
      success: true,
      data: {
        ...mockPredictionData,
        data_freshness_seconds: 5
      }
    });

    render(<ShipmentPredictiveETACard shipmentId="101" authoritativeEta="2026-09-15T18:00:00Z" />);

    await waitFor(() => {
      expect(screen.getByTitle(/Force refresh predictive model/i)).toBeInTheDocument();
    });

    const refreshBtn = screen.getByTitle(/Force refresh predictive model/i);
    fireEvent.click(refreshBtn);

    await waitFor(() => {
      expect(mockGetShipmentPredictedETA).toHaveBeenCalledWith('101', true);
    });
  });
});

import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import ShipmentDisruptionForecastCard from '../../components/predictions/ShipmentDisruptionForecastCard';
import * as predictionService from '../../services/predictionService';

const mockGetShipmentPredictedExceptions = vi.fn();
const mockRefreshShipmentPredictedExceptions = vi.fn();
const mockAcknowledgePrediction = vi.fn();
const mockRequestAction = vi.fn();

vi.mock('../../services/predictionService', () => {
  const service = {
    getShipmentPredictedExceptions: (...args) => mockGetShipmentPredictedExceptions(...args),
    refreshShipmentPredictedExceptions: (...args) => mockRefreshShipmentPredictedExceptions(...args),
    acknowledgePrediction: (...args) => mockAcknowledgePrediction(...args),
    requestAction: (...args) => mockRequestAction(...args),
  };
  return {
    default: service,
    predictionService: service,
    getShipmentPredictedExceptions: service.getShipmentPredictedExceptions,
    refreshShipmentPredictedExceptions: service.refreshShipmentPredictedExceptions,
    acknowledgePrediction: service.acknowledgePrediction,
    requestAction: service.requestAction,
  };
});

describe('ShipmentDisruptionForecastCard', () => {
  const mockCriticalDisruption = {
    prediction_id: 'pred-disrupt-103-customs',
    prediction_type: 'shipment_disruption_forecast',
    related_record_id: '103',
    severity: 'CRITICAL',
    confidence_score: 0.94,
    confidence_band: 'HIGH',
    disruption_category: 'CUSTOMS_HOLD_RISK',
    linked_exception_id: 101,
    prediction_statement: 'Critical customs hold risk detected. Grounded in active exception #101.',
    explanation: 'Shipment has an active customs hold for documentation discrepancy.',
    supporting_signals: [
      'Active exception #101: CUSTOMS_HOLD (HS Code mismatch)',
      'Milestone Customs Clearance pending',
      'Port dwell time exceeded 48 hours'
    ],
    recommended_action: 'Expedite corrected bill of lading and customs filing with clearance broker.',
    status: 'PUBLISHED',
    review_status: 'REVIEW_RECOMMENDED',
    requires_approval: true,
    correlation_id: 'corr-pred-103-test',
    prediction_version: 'v4.3.0',
    created_at: '2026-09-09T18:00:00Z'
  };

  const mockInsufficientData = {
    prediction_id: 'pred-disrupt-104-insufficient',
    prediction_type: 'shipment_disruption_forecast',
    related_record_id: '104',
    severity: 'INSUFFICIENT_DATA',
    confidence_score: 0.1,
    confidence_band: 'LOW',
    insufficient_data: true,
    review_status: 'INSUFFICIENT_DATA',
    explanation: 'No milestone records found for shipment. Proactive disruption forecasting requires departure telemetry.'
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders loading indicator while fetching', () => {
    mockGetShipmentPredictedExceptions.mockReturnValue(new Promise(() => {}));
    render(<ShipmentDisruptionForecastCard shipmentId="103" />);
    expect(screen.getByTestId('shipment-disruption-loading')).toBeInTheDocument();
    expect(screen.getByText(/Synthesizing early disruption warning signals/i)).toBeInTheDocument();
  });

  it('renders critical disruption forecast with linked exception and grounded explanation', async () => {
    mockGetShipmentPredictedExceptions.mockResolvedValue({
      success: true,
      data: mockCriticalDisruption
    });

    render(<ShipmentDisruptionForecastCard shipmentId="103" />);

    await waitFor(() => {
      expect(screen.getByTestId('shipment-disruption-forecast-card')).toBeInTheDocument();
    });

    const card = screen.getByTestId('shipment-disruption-forecast-card');
    expect(card.querySelector('.disruption-category-badge')).toHaveTextContent('CUSTOMS HOLD RISK');
    expect(screen.getAllByText(/CRITICAL/i).length).toBeGreaterThanOrEqual(1);
    expect(card.querySelector('.disruption-linked-pill')).toHaveTextContent('Exception #101');
    expect(screen.getByTestId('disruption-statement')).toHaveTextContent(/Critical customs hold risk detected/i);
    expect(screen.getByTestId('disruption-explanation')).toHaveTextContent(/Shipment has an active customs hold/i);
  });

  it('expands collapsible supporting signals when toggled', async () => {
    mockGetShipmentPredictedExceptions.mockResolvedValue({
      success: true,
      data: mockCriticalDisruption
    });

    render(<ShipmentDisruptionForecastCard shipmentId="103" />);

    await waitFor(() => {
      expect(screen.getByText(/Source-Grounded Signals & Audit Records/i)).toBeInTheDocument();
    });

    // Evidence content not rendered initially
    expect(screen.queryByTestId('disruption-evidence-content')).not.toBeInTheDocument();

    // Toggle expand
    fireEvent.click(screen.getByText(/Source-Grounded Signals & Audit Records/i));

    await waitFor(() => {
      expect(screen.getByTestId('disruption-evidence-content')).toBeInTheDocument();
      expect(screen.getByText(/Active exception #101: CUSTOMS_HOLD/i)).toBeInTheDocument();
      expect(screen.getByText(/pred-disrupt-103-customs/i)).toBeInTheDocument();
    });
  });

  it('renders insufficient data banner when telemetry is missing without fabricating predictions', async () => {
    mockGetShipmentPredictedExceptions.mockResolvedValue({
      success: true,
      data: mockInsufficientData
    });

    render(<ShipmentDisruptionForecastCard shipmentId="104" />);

    await waitFor(() => {
      expect(screen.getByTestId('disruption-forecast-insufficient-banner')).toBeInTheDocument();
      expect(screen.getByText(/Insufficient Operational Telemetry/i)).toBeInTheDocument();
      expect(screen.getByText(/No milestone records found for shipment/i)).toBeInTheDocument();
    });
  });

  it('handles acknowledge risk button action', async () => {
    mockGetShipmentPredictedExceptions.mockResolvedValue({
      success: true,
      data: mockCriticalDisruption
    });
    mockAcknowledgePrediction.mockResolvedValue({ success: true });

    render(<ShipmentDisruptionForecastCard shipmentId="103" />);

    await waitFor(() => {
      expect(screen.getByText(/Acknowledge Risk/i)).toBeInTheDocument();
    });

    fireEvent.click(screen.getByText(/Acknowledge Risk/i));

    await waitFor(() => {
      expect(mockAcknowledgePrediction).toHaveBeenCalledWith('pred-disrupt-103-customs');
      expect(screen.getByText(/Early warning signal acknowledged/i)).toBeInTheDocument();
    });
  });

  it('handles queue operational action via Action System with HITL notice', async () => {
    mockGetShipmentPredictedExceptions.mockResolvedValue({
      success: true,
      data: mockCriticalDisruption
    });
    mockRequestAction.mockResolvedValue({ success: true });

    render(<ShipmentDisruptionForecastCard shipmentId="103" />);

    await waitFor(() => {
      expect(screen.getByText(/Queue Mitigation Action/i)).toBeInTheDocument();
    });

    fireEvent.click(screen.getByText(/Queue Mitigation Action/i));

    await waitFor(() => {
      expect(mockRequestAction).toHaveBeenCalledWith(
        'pred-disrupt-103-customs',
        expect.stringContaining('Proactive mitigation for CUSTOMS_HOLD_RISK')
      );
      expect(screen.getByText(/Mitigation action queued in Go Action System/i)).toBeInTheDocument();
    });
  });
});

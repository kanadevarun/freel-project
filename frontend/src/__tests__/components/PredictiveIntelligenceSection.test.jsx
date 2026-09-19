import React from 'react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import PredictiveIntelligenceSection from '../../components/predictions/PredictiveIntelligenceSection';
import { predictionService } from '../../services/predictionService';

vi.mock('../../services/predictionService', () => ({
  predictionService: {
    listPredictions: vi.fn(),
    acknowledgePrediction: vi.fn(),
    dismissPrediction: vi.fn(),
    requestAction: vi.fn(),
  },
}));

describe('PredictiveIntelligenceSection Component', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders grounded prediction with severity badge, confidence score, and sources', async () => {
    predictionService.listPredictions.mockResolvedValueOnce({
      success: true,
      predictions: [
        {
          prediction_id: 'pred-ship-101',
          module: 'shipments',
          prediction_type: 'SHIPMENT_ETA_DELAY',
          severity: 'CRITICAL',
          confidence_score: 0.88,
          confidence_band: 'HIGH',
          prediction_statement: 'Vessel schedule modeling projects 60h delay arriving at DEHAM.',
          predicted_value: '+60h ETA Variance',
          time_horizon: '7_DAYS',
          status: 'PUBLISHED',
          explanation: 'Projected schedule drift combines carrier transshipment hold with port dwell multiplier.',
          recommended_action: 'Notify consignee of projected delay',
          action_type: 'shipments.send_customer_update',
          requires_approval: true,
          source_references: [
            {
              source_module: 'shipments',
              source_record_id: '101',
              source_field: 'milestones.transshipment',
              source_timestamp: '2026-09-09T18:00:00Z',
            },
          ],
        },
      ],
      total: 1,
    });

    render(
      <MemoryRouter>
        <PredictiveIntelligenceSection module="shipments" relatedRecordId="101" />
      </MemoryRouter>
    );

    await waitFor(() => {
      expect(
        screen.getByText('Vessel schedule modeling projects 60h delay arriving at DEHAM.')
      ).toBeInTheDocument();
    });

    expect(screen.getAllByText('CRITICAL').length).toBeGreaterThanOrEqual(1);
    expect(screen.getByText('88% (HIGH)')).toBeInTheDocument();
    expect(screen.getByText('PREDICTED')).toBeInTheDocument();
    expect(screen.getByText('Phase 4 Foundation')).toBeInTheDocument();
    expect(screen.getByText('+60h ETA Variance')).toBeInTheDocument();
  });

  it('allows acknowledging an active prediction', async () => {
    predictionService.listPredictions.mockResolvedValueOnce({
      success: true,
      predictions: [
        {
          prediction_id: 'pred-ship-101',
          module: 'shipments',
          severity: 'HIGH',
          confidence_score: 0.85,
          confidence_band: 'HIGH',
          prediction_statement: 'Transshipment delay forecast for Container MSCU1234567.',
          status: 'PUBLISHED',
          source_references: [
            {
              source_module: 'shipments',
              source_record_id: '101',
              source_field: 'status',
              source_timestamp: '2026-09-09T18:00:00Z',
            },
          ],
        },
      ],
      total: 1,
    });

    predictionService.acknowledgePrediction.mockResolvedValueOnce({
      success: true,
      message: 'Prediction acknowledged',
    });

    render(
      <MemoryRouter>
        <PredictiveIntelligenceSection module="shipments" />
      </MemoryRouter>
    );

    await waitFor(() => {
      expect(screen.getByText('Acknowledge')).toBeInTheDocument();
    });

    fireEvent.click(screen.getByText('Acknowledge'));

    await waitFor(() => {
      expect(predictionService.acknowledgePrediction).toHaveBeenCalledWith('pred-ship-101');
    });
  });

  it('handles empty state with informative light surface guidance', async () => {
    predictionService.listPredictions.mockResolvedValueOnce({
      success: true,
      predictions: [],
      total: 0,
    });

    render(
      <MemoryRouter>
        <PredictiveIntelligenceSection module="invoices" />
      </MemoryRouter>
    );

    await waitFor(() => {
      expect(screen.getByText('No active predictive alerts')).toBeInTheDocument();
    });

    expect(
      screen.getByText(/Authoritative milestones and ledger records show zero negative variance/i)
    ).toBeInTheDocument();
  });

  it('handles error state with retry button', async () => {
    predictionService.listPredictions.mockRejectedValueOnce(
      new Error('Failed to fetch predictions')
    );

    render(
      <MemoryRouter>
        <PredictiveIntelligenceSection module="contracts" />
      </MemoryRouter>
    );

    await waitFor(() => {
      expect(screen.getByText('Retry')).toBeInTheDocument();
    });
  });
});

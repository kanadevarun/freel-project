import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import React from 'react';
import NetworkPerformancePredictiveCard from '../../components/predictions/NetworkPerformancePredictiveCard';
import * as predictionService from '../../services/predictionService';

const mockGetCarrier = vi.fn();
const mockRefreshCarrier = vi.fn();
const mockGetLane = vi.fn();
const mockRefreshLane = vi.fn();
const mockGetCustomer = vi.fn();
const mockRefreshCustomer = vi.fn();
const mockRequestAction = vi.fn();

vi.mock('../../services/predictionService', () => {
  const service = {
    getCarrierPredictedPerformance: (...args) => mockGetCarrier(...args),
    refreshCarrierPredictedPerformance: (...args) => mockRefreshCarrier(...args),
    getLanePredictedPerformance: (...args) => mockGetLane(...args),
    refreshLanePredictedPerformance: (...args) => mockRefreshLane(...args),
    getCustomerPredictedServicePerformance: (...args) => mockGetCustomer(...args),
    refreshCustomerPredictedServicePerformance: (...args) => mockRefreshCustomer(...args),
    requestAction: (...args) => mockRequestAction(...args),
  };
  return {
    default: service,
    ...service,
  };
});

describe('NetworkPerformancePredictiveCard', () => {
  const mockCarrierPrediction = {
    prediction_id: 'pred-carrier-cmdu-test',
    module: 'carriers',
    prediction_type: 'CARRIER_PERFORMANCE_RISK',
    prediction_category: 'CARRIER_PERFORMANCE_RISK',
    related_record_type: 'CARRIER',
    related_record_id: 'CMDU',
    severity: 'HIGH',
    confidence_score: 0.93,
    confidence_band: 'HIGH',
    prediction_statement: 'Carrier Regulatory Hold Risk: CMA CGM (CMDU) exhibits elevated operational risk on corridor INNSA-USNYC.',
    explanation: 'Active customs hold on shipment #103 at Nhava Sheva. Regulatory clearance dwell poses acute schedule risk.',
    predicted_value: 'CUSTOMS_HOLD_RISK',
    time_horizon: '14_DAYS',
    status: 'PUBLISHED',
    carrier_reference: 'CMDU',
    lane_reference: 'INNSA-USNYC',
    comparison_period: 'LAST_90_DAYS',
    sample_size: 14,
    supporting_signals: [
      { signal_name: 'customs_holds_count', observed_value: '1', baseline_value: '0', importance_weight: 0.96 },
      { signal_name: 'on_time_reliability', observed_value: '78.5%', baseline_value: '>= 90.0%', importance_weight: 0.90 }
    ],
    source_references: [
      {
        source_module: 'shipments',
        source_record_id: '103',
        source_field: 'shipments.status',
        source_timestamp: '2026-09-10T10:00:00Z',
        carrier_reference: 'CMDU',
        lane_reference: 'INNSA-USNYC',
        comparison_period: 'LAST_90_DAYS',
        sample_size: 14,
      }
    ],
    source_timestamp: '2026-09-10T10:00:00Z',
    recommended_action: 'Review carrier operational hold with compliance team and dispatch updated commercial paperwork.',
    action_type: 'carriers.review_performance',
    is_action_required: true,
    requires_approval: true,
  };

  const mockLanePrediction = {
    prediction_id: 'pred-lane-usnyc-test',
    module: 'network',
    prediction_type: 'LANE_PERFORMANCE_RISK',
    prediction_category: 'LANE_PERFORMANCE_RISK',
    related_record_type: 'LANE',
    related_record_id: 'INNSA-USNYC',
    severity: 'HIGH',
    confidence_score: 0.92,
    confidence_band: 'HIGH',
    prediction_statement: 'Trade Corridor Risk: Corridor INNSA-USNYC shows elevated regulatory inspection exposure.',
    explanation: 'Shipment #103 on lane INNSA-USNYC is actively held under customs inspection.',
    predicted_value: 'CUSTOMS_DWELL_RISK',
    time_horizon: '14_DAYS',
    lane_reference: 'INNSA-USNYC',
    comparison_period: 'LAST_90_DAYS',
    sample_size: 18,
    supporting_signals: [
      { signal_name: 'active_customs_hold', observed_value: '1 ACTIVE HOLD', baseline_value: '0', importance_weight: 0.95 }
    ],
    source_references: [
      { source_module: 'shipments', source_record_id: '103', source_field: 'shipments.destination_port', source_timestamp: '2026-09-10T10:00:00Z' }
    ],
    source_timestamp: '2026-09-10T10:00:00Z',
    recommended_action: 'Request pre-clearance documentation audit for upcoming vessel cutoffs.',
    is_action_required: true,
    requires_approval: true,
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders carrier predictive intelligence card with metrics, badges, and statement', async () => {
    mockGetCarrier.mockResolvedValueOnce({ data: mockCarrierPrediction });

    render(<NetworkPerformancePredictiveCard entityType="carrier" entityId="CMDU" />);

    await waitFor(() => {
      expect(screen.getByText(/Carrier Performance Intelligence: CMDU/i)).toBeInTheDocument();
    });

    expect(screen.getByText(/CMA CGM \(CMDU\) exhibits elevated operational risk/i)).toBeInTheDocument();
    expect(screen.getByText('CUSTOMS_HOLD_RISK')).toBeInTheDocument();
    expect(screen.getByText(/14 authoritative records/i)).toBeInTheDocument();
    expect(screen.getByText(/Benchmark Window: LAST_90_DAYS/i)).toBeInTheDocument();
    expect(screen.getByText(/Requires Operator Approval/i)).toBeInTheDocument();
  });

  it('toggles evidence section to view analytical signals and verifiable sources', async () => {
    mockGetCarrier.mockResolvedValueOnce({ data: mockCarrierPrediction });

    render(<NetworkPerformancePredictiveCard entityType="carrier" entityId="CMDU" />);

    await waitFor(() => {
      expect(screen.getByText(/Authoritative Source Grounding & Signals/i)).toBeInTheDocument();
    });

    const toggleBtn = screen.getByText(/Authoritative Source Grounding & Signals/i);
    fireEvent.click(toggleBtn);

    expect(screen.getByText('customs_holds_count')).toBeInTheDocument();
    expect(screen.getByText('78.5%')).toBeInTheDocument();
    expect(screen.getByText(/shipments\.status/i)).toBeInTheDocument();
    expect(screen.getByText(/\(Record ID: #103\)/i)).toBeInTheDocument();
  });

  it('renders trade corridor lane predictive intelligence', async () => {
    mockGetLane.mockResolvedValueOnce({ data: mockLanePrediction });

    render(<NetworkPerformancePredictiveCard entityType="lane" entityId="INNSA-USNYC" />);

    await waitFor(() => {
      expect(screen.getByText(/Trade Corridor Risk: INNSA-USNYC/i)).toBeInTheDocument();
    });

    expect(screen.getByText(/Corridor INNSA-USNYC shows elevated regulatory inspection exposure/i)).toBeInTheDocument();
    expect(screen.getByText('CUSTOMS_DWELL_RISK')).toBeInTheDocument();
    expect(screen.getByText(/18 authoritative records/i)).toBeInTheDocument();
  });

  it('handles action request submission gracefully', async () => {
    mockGetCarrier.mockResolvedValueOnce({ data: mockCarrierPrediction });
    mockRequestAction.mockResolvedValueOnce({ success: true });

    render(<NetworkPerformancePredictiveCard entityType="carrier" entityId="CMDU" />);

    await waitFor(() => {
      expect(screen.getByText('Request Action')).toBeInTheDocument();
    });

    const actionBtn = screen.getByText('Request Action');
    fireEvent.click(actionBtn);

    await waitFor(() => {
      expect(mockRequestAction).toHaveBeenCalledWith(
        'pred-carrier-cmdu-test',
        expect.stringContaining('Operational escalation initiated')
      );
    });
  });

  it('handles insufficient data state safely without breaking', async () => {
    const insufficientPred = {
      ...mockCarrierPrediction,
      insufficient_data: true,
      predicted_value: 'INSUFFICIENT_DATA',
      prediction_statement: 'Insufficient historical performance data for CARRIER #UNKNOWN.',
      sample_size: 0,
      severity: 'LOW',
      confidence_score: 0.0,
      confidence_band: 'LOW',
      recommended_action: null,
      is_action_required: false,
    };
    mockGetCarrier.mockResolvedValueOnce({ data: insufficientPred });

    render(<NetworkPerformancePredictiveCard entityType="carrier" entityId="UNKNOWN" />);

    await waitFor(() => {
      expect(screen.getByText(/Insufficient historical performance data/i)).toBeInTheDocument();
    });
    expect(screen.getByText(/0 authoritative records/i)).toBeInTheDocument();
  });
});

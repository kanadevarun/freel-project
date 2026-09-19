import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { BrowserRouter } from 'react-router-dom';
import WorkloadCapacityPredictiveCard from '../../components/predictions/WorkloadCapacityPredictiveCard';
import * as predictionService from '../../services/predictionService';

const mockGetWorkloadPrediction = vi.fn();
const mockRefreshWorkloadPrediction = vi.fn();
const mockGetCapacityPrediction = vi.fn();
const mockRefreshCapacityPrediction = vi.fn();
const mockGetDemandPrediction = vi.fn();
const mockRefreshDemandPrediction = vi.fn();
const mockGetWorkloadCapacitySummary = vi.fn();
const mockRequestAction = vi.fn();

vi.mock('../../services/predictionService', () => {
  const service = {
    getWorkloadPrediction: (...args) => mockGetWorkloadPrediction(...args),
    refreshWorkloadPrediction: (...args) => mockRefreshWorkloadPrediction(...args),
    getCapacityPrediction: (...args) => mockGetCapacityPrediction(...args),
    refreshCapacityPrediction: (...args) => mockRefreshCapacityPrediction(...args),
    getDemandPrediction: (...args) => mockGetDemandPrediction(...args),
    refreshDemandPrediction: (...args) => mockRefreshDemandPrediction(...args),
    getWorkloadCapacitySummary: (...args) => mockGetWorkloadCapacitySummary(...args),
    requestAction: (...args) => mockRequestAction(...args),
  };
  return {
    default: service,
    predictionService: service,
    getWorkloadPrediction: service.getWorkloadPrediction,
    refreshWorkloadPrediction: service.refreshWorkloadPrediction,
    getCapacityPrediction: service.getCapacityPrediction,
    refreshCapacityPrediction: service.refreshCapacityPrediction,
    getDemandPrediction: service.getDemandPrediction,
    refreshDemandPrediction: service.refreshDemandPrediction,
    getWorkloadCapacitySummary: service.getWorkloadCapacitySummary,
    requestAction: service.requestAction,
  };
});

describe('WorkloadCapacityPredictiveCard Component', () => {
  const mockApprovalPrediction = {
    prediction_id: 'pred-workload-appr-test-123',
    module: 'workload',
    prediction_type: 'APPROVAL_WORKLOAD_SPIKE',
    prediction_category: 'WORKLOAD_SPIKE',
    related_record_type: 'WORKLOAD',
    related_record_id: 'org-2-approvals',
    severity: 'HIGH',
    confidence_score: 0.92,
    confidence_band: 'HIGH',
    prediction_statement: 'Approval Workload Spike: 23 pending high/critical review requests are concentrated in COMMERCIAL workflows.',
    workload_type: 'APPROVALS',
    pending_count: 23,
    capacity_limit: 15,
    utilization_rate: 0.925,
    recommended_action: 'Batch review pending commercial requests and reallocate pricing approval duties.',
    reasoning_summary: '23 pending approval requests exceed threshold of 15 with 22.4h average dwell time.',
    sample_size_evaluated: 23,
    data_coverage_score: 1.0,
    created_at: '2026-09-10T14:30:00Z',
    source_references: [
      {
        entity_type: 'APPROVAL_REQUEST',
        entity_id: '101',
        workload_type: 'PRICING',
        signal_contribution: 'CRITICAL_REQUEST_DWELL'
      }
    ],
    limitations: ['Assumes standard business hours review velocity.'],
  };

  const mockSummaryData = {
    org_id: 2,
    generated_at: '2026-09-10T14:30:00Z',
    approvals_workload: {
      prediction_id: 'pred-workload-appr-test-123',
      prediction_statement: 'Approval Workload Spike: 23 pending requests.',
      severity: 'HIGH',
      pending_count: 23,
    },
    documentation_workload: {
      prediction_id: 'pred-workload-docs-test-456',
      prediction_statement: 'Operational Documentation Workload: 5 compliance issues pending.',
      severity: 'HIGH',
      pending_count: 5,
    },
    corridor_capacity: {
      prediction_id: 'pred-capacity-lane-test-789',
      prediction_statement: 'Trade Corridor Capacity Pressure: Corridors INNSA-USNYC / INNSA-NLRTM.',
      severity: 'MEDIUM',
      utilization_rate: 0.875,
    },
    quote_demand: {
      prediction_id: 'pred-demand-quote-test-999',
      prediction_statement: 'Commercial Quote Demand: 5 active RFQs in pipeline.',
      severity: 'MEDIUM',
      pending_count: 5,
    },
  };

  beforeEach(() => {
    vi.clearAllMocks();
    mockGetWorkloadPrediction.mockResolvedValue({ data: mockApprovalPrediction });
    mockRefreshWorkloadPrediction.mockResolvedValue({ data: mockApprovalPrediction });
    mockGetWorkloadCapacitySummary.mockResolvedValue({ data: mockSummaryData });
  });

  const renderComponent = (props = {}) => {
    return render(
      <BrowserRouter>
        <WorkloadCapacityPredictiveCard {...props} />
      </BrowserRouter>
    );
  };

  it('renders loading state initially and then shows prediction', async () => {
    renderComponent({ defaultTab: 'approvals' });

    expect(screen.getByTestId('workload-capacity-predictive-card')).toBeInTheDocument();
    expect(screen.getByTestId('wc-loading-state')).toBeInTheDocument();

    await waitFor(() => {
      expect(screen.getByTestId('wc-statement')).toBeInTheDocument();
    });

    expect(screen.getByTestId('wc-statement')).toHaveTextContent(/Approval Workload Spike/i);
    expect(screen.getByTestId('metric-pending-count')).toHaveTextContent('23');
    expect(screen.getByTestId('metric-utilization')).toHaveTextContent('92.5%');
  });

  it('renders advisory notice banner and recommended action', async () => {
    renderComponent({ defaultTab: 'approvals' });

    await waitFor(() => {
      expect(screen.getByTestId('wc-statement')).toBeInTheDocument();
    });

    expect(screen.getByText(/Advisory Intelligence:/i)).toBeInTheDocument();
    expect(screen.getByTestId('wc-recommendation')).toHaveTextContent(/Batch review pending commercial requests/i);
    expect(screen.getByTestId('wc-queue-action-btn')).toBeInTheDocument();
  });

  it('allows expanding evidence and grounding telemetry drawer', async () => {
    renderComponent({ defaultTab: 'approvals' });

    await waitFor(() => {
      expect(screen.getByTestId('wc-evidence-toggle')).toBeInTheDocument();
    });

    // Evidence body should initially not be visible
    expect(screen.queryByTestId('wc-evidence-body')).not.toBeInTheDocument();

    // Click to expand
    fireEvent.click(screen.getByTestId('wc-evidence-toggle'));

    await waitFor(() => {
      expect(screen.getByTestId('wc-evidence-body')).toBeInTheDocument();
    });

    expect(screen.getByText(/Reasoning & Risk Summary/i)).toBeInTheDocument();
    expect(screen.getByText(/Source Database Records Evaluated/i)).toBeInTheDocument();
    expect(screen.getByText(/CRITICAL_REQUEST_DWELL/i)).toBeInTheDocument();
  });

  it('handles action queue modal submission via Action System', async () => {
    mockRequestAction.mockResolvedValue({ success: true, action_id: 'act-101' });

    renderComponent({ defaultTab: 'approvals' });

    await waitFor(() => {
      expect(screen.getByTestId('wc-queue-action-btn')).toBeInTheDocument();
    });

    // Click Queue Action button
    fireEvent.click(screen.getByTestId('wc-queue-action-btn'));

    // Modal opens
    expect(screen.getByTestId('wc-action-modal')).toBeInTheDocument();

    // Click submit in modal
    fireEvent.click(screen.getByTestId('wc-modal-submit-btn'));

    await waitFor(() => {
      expect(mockRequestAction).toHaveBeenCalledWith(
        'pred-workload-appr-test-123',
        expect.any(String)
      );
      expect(screen.getByTestId('wc-action-success')).toHaveTextContent(/Action queued for Human-in-the-Loop review/i);
    });
  });

  it('renders unified portfolio summary when summary tab is selected', async () => {
    renderComponent({ defaultTab: 'summary' });

    await waitFor(() => {
      expect(screen.getByTestId('wc-summary-view')).toBeInTheDocument();
    });

    expect(screen.getByText('Approval Queue Pressure')).toBeInTheDocument();
    expect(screen.getByText('Documentation Compliance')).toBeInTheDocument();
    expect(screen.getByText('Trade Corridor Capacity')).toBeInTheDocument();
    expect(screen.getAllByText('Commercial Quote Demand').length).toBeGreaterThanOrEqual(1);
  });

  it('renders error state safely when network error occurs', async () => {
    mockGetWorkloadPrediction.mockRejectedValue(new Error('Connection timed out'));

    renderComponent({ defaultTab: 'approvals' });

    await waitFor(() => {
      expect(screen.getByTestId('wc-error-state')).toBeInTheDocument();
    });

    expect(screen.getByText('Intelligence Feed Offline')).toBeInTheDocument();
    expect(screen.getByText('Connection timed out')).toBeInTheDocument();
  });
});

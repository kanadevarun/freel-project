import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { BrowserRouter } from 'react-router-dom';
import ResourceBottleneckPredictiveCard from '../../components/predictions/ResourceBottleneckPredictiveCard';
import * as predictionService from '../../services/predictionService';

const mockGetOperationalBottleneck = vi.fn();
const mockRefreshOperationalBottleneck = vi.fn();
const mockGetResourceAllocation = vi.fn();
const mockRefreshResourceAllocation = vi.fn();
const mockGetResourceBottleneckSummary = vi.fn();
const mockRequestAction = vi.fn();

vi.mock('../../services/predictionService', () => {
  const service = {
    getOperationalBottleneck: (...args) => mockGetOperationalBottleneck(...args),
    refreshOperationalBottleneck: (...args) => mockRefreshOperationalBottleneck(...args),
    getResourceAllocation: (...args) => mockGetResourceAllocation(...args),
    refreshResourceAllocation: (...args) => mockRefreshResourceAllocation(...args),
    getResourceBottleneckSummary: (...args) => mockGetResourceBottleneckSummary(...args),
    requestAction: (...args) => mockRequestAction(...args),
  };
  return {
    default: service,
    predictionService: service,
    getOperationalBottleneck: service.getOperationalBottleneck,
    refreshOperationalBottleneck: service.refreshOperationalBottleneck,
    getResourceAllocation: service.getResourceAllocation,
    refreshResourceAllocation: service.refreshResourceAllocation,
    getResourceBottleneckSummary: service.getResourceBottleneckSummary,
    requestAction: service.requestAction,
  };
});

describe('ResourceBottleneckPredictiveCard Component', () => {
  const mockApprovalBottleneck = {
    prediction_id: 'pred-btln-appr-test-123',
    module: 'bottleneck',
    prediction_type: 'APPROVAL_BOTTLENECK',
    severity: 'HIGH',
    confidence_score: 0.93,
    confidence_band: 'HIGH',
    prediction_statement: 'Likely Approval Bottleneck: 23 pending review requests exceed throughput limit with 22.4h dwell time.',
    bottleneck_type: 'APPROVALS',
    affected_stage: 'PRICING_COMMERCIAL_GATEWAY',
    pending_count: 23,
    queue_dwell_hours: 22.4,
    recommended_action: 'Batch review high-priority pricing requests and rebalance review duties.',
    reasoning_summary: '23 pending approval requests exceed threshold of 10 with 22.4h average dwell time.',
    sample_size_evaluated: 23,
    data_coverage_score: 1.0,
    created_at: '2026-09-10T14:30:00Z',
    source_references: [
      {
        entity_type: 'APPROVAL_REQUEST',
        entity_id: '101',
        bottleneck_type: 'APPROVALS',
        queue_dwell_hours: 22.4,
        signal_contribution: 'CRITICAL_REQUEST_DWELL'
      }
    ],
    limitations: ['Assumes standard business hours review velocity.'],
  };

  const mockDocumentationBottleneck = {
    prediction_id: 'pred-btln-docs-test-456',
    module: 'bottleneck',
    prediction_type: 'DOCUMENTATION_BOTTLENECK',
    severity: 'HIGH',
    confidence_score: 0.88,
    confidence_band: 'HIGH',
    prediction_statement: 'Likely Documentation Bottleneck: 1 document compliance discrepancies pending across 2 active shipments.',
    bottleneck_type: 'DOCUMENTATION',
    affected_stage: 'EXPORT_CUSTOMS_CLEARANCE',
    pending_count: 1,
    queue_dwell_hours: 18.0,
    recommended_action: 'Expedite missing shipping instructions and commercial invoices.',
    sample_size_evaluated: 2,
    data_coverage_score: 1.0,
    created_at: '2026-09-10T14:30:00Z',
    source_references: [],
    limitations: [],
  };

  const mockCrossModuleBottleneck = {
    prediction_id: 'pred-btln-cross-test-789',
    module: 'bottleneck',
    prediction_type: 'CROSS_MODULE_BOTTLENECK',
    severity: 'HIGH',
    confidence_score: 0.95,
    confidence_band: 'VERY_HIGH',
    prediction_statement: 'Likely Cross-Module Bottleneck: Terminal customs hold on Shipment #103 blocks downstream delivery dispatch.',
    bottleneck_type: 'CROSS_MODULE_EXCEPTION',
    affected_stage: 'FINAL_DELIVERY_INVOICING',
    queue_dwell_hours: 48.0,
    carrier: 'CMDU',
    recommended_action: 'Coordinate with CMA CGM import desk and dispatch customs broker.',
    sample_size_evaluated: 1,
    data_coverage_score: 1.0,
    created_at: '2026-09-10T14:30:00Z',
    source_references: [],
    limitations: [],
  };

  const mockResourceAllocation = {
    prediction_id: 'pred-rsrc-alloc-test-321',
    module: 'resource',
    prediction_type: 'RESOURCE_ALLOCATION_IMBALANCE',
    severity: 'HIGH',
    confidence_score: 0.92,
    confidence_band: 'HIGH',
    prediction_statement: 'Likely Resource Allocation Imbalance: 100% of operational approvals are assigned to owner kanadevarun123@gmail.com.',
    bottleneck_type: 'OWNER_WORKLOAD_IMBALANCE',
    affected_stage: 'OPERATIONAL_SIGN_OFF',
    assigned_owner: 'kanadevarun123@gmail.com',
    workload_count: 23,
    recommended_action: 'Reassign commercial tier-2 reviews to secondary logistics manager.',
    sample_size_evaluated: 23,
    data_coverage_score: 1.0,
    created_at: '2026-09-10T14:30:00Z',
    source_references: [],
    limitations: [],
  };

  const mockSummaryData = {
    org_id: 2,
    generated_at: '2026-09-10T14:30:00Z',
    approvals_bottleneck: {
      prediction_id: 'pred-btln-appr-test-123',
      prediction_statement: 'Likely Approval Bottleneck: 23 pending requests.',
      severity: 'HIGH',
      pending_count: 23,
    },
    documentation_bottleneck: {
      prediction_id: 'pred-btln-docs-test-456',
      prediction_statement: 'Likely Documentation Bottleneck: 1 active discrepancy.',
      severity: 'HIGH',
      pending_count: 1,
    },
    cross_module_bottleneck: {
      prediction_id: 'pred-btln-cross-test-789',
      prediction_statement: 'Likely Cross-Module Bottleneck: Terminal customs hold on Shipment #103.',
      severity: 'HIGH',
      queue_dwell_hours: 48,
    },
    resource_allocation: {
      prediction_id: 'pred-rsrc-alloc-test-321',
      prediction_statement: 'Likely Resource Allocation Imbalance: 100% assigned to single owner.',
      severity: 'HIGH',
      workload_count: 23,
    },
  };

  beforeEach(() => {
    vi.clearAllMocks();
    mockGetOperationalBottleneck.mockResolvedValue(mockApprovalBottleneck);
    mockRefreshOperationalBottleneck.mockResolvedValue(mockApprovalBottleneck);
    mockGetResourceAllocation.mockResolvedValue(mockResourceAllocation);
    mockRefreshResourceAllocation.mockResolvedValue(mockResourceAllocation);
    mockGetResourceBottleneckSummary.mockResolvedValue(mockSummaryData);
    mockRequestAction.mockResolvedValue({ status: 'ACTION_REQUESTED', id: 'act-test-1' });
  });

  const renderComponent = (props = {}) => {
    return render(
      <BrowserRouter>
        <ResourceBottleneckPredictiveCard {...props} />
      </BrowserRouter>
    );
  };

  it('renders correctly and loads approval bottleneck by default', async () => {
    renderComponent();

    expect(screen.getByTestId('resource-bottleneck-predictive-card')).toBeInTheDocument();
    expect(screen.getByText(/Predictive Resource Allocation & Operational Bottleneck Intelligence/i)).toBeInTheDocument();

    await waitFor(() => {
      expect(screen.getByTestId('rb-prediction-body')).toBeInTheDocument();
    });

    expect(screen.getByTestId('rb-statement')).toHaveTextContent(/23 pending review requests exceed throughput limit/i);
    expect(screen.getByTestId('metric-pending-count')).toHaveTextContent('23');
    expect(screen.getByTestId('metric-dwell-hours')).toHaveTextContent('22.4 hrs');
    expect(screen.getByTestId('metric-stage')).toHaveTextContent('PRICING_COMMERCIAL_GATEWAY');
  });

  it('switches tabs and fetches documentation and cross-module bottleneck data', async () => {
    mockGetOperationalBottleneck
      .mockResolvedValueOnce(mockApprovalBottleneck)
      .mockResolvedValueOnce(mockDocumentationBottleneck)
      .mockResolvedValueOnce(mockCrossModuleBottleneck);

    renderComponent();

    await waitFor(() => {
      expect(screen.getByTestId('rb-statement')).toBeInTheDocument();
    });

    // Switch to documentation tab
    fireEvent.click(screen.getByTestId('tab-documentation'));
    await waitFor(() => {
      expect(mockGetOperationalBottleneck).toHaveBeenCalledWith('documentation');
      expect(screen.getByTestId('rb-statement')).toHaveTextContent(/1 document compliance discrepancies pending/i);
    });

    // Switch to cross-module tab
    fireEvent.click(screen.getByTestId('tab-cross-module'));
    await waitFor(() => {
      expect(mockGetOperationalBottleneck).toHaveBeenCalledWith('cross_module');
      expect(screen.getByTestId('rb-statement')).toHaveTextContent(/Terminal customs hold on Shipment #103/i);
      expect(screen.getByTestId('metric-carrier')).toHaveTextContent('CMDU');
    });
  });

  it('switches to owner workload tab and renders resource allocation signals', async () => {
    mockGetOperationalBottleneck.mockResolvedValueOnce(mockApprovalBottleneck);
    renderComponent();

    await waitFor(() => {
      expect(screen.getByTestId('rb-statement')).toBeInTheDocument();
    });

    // Switch to resource tab
    fireEvent.click(screen.getByTestId('tab-resource'));
    await waitFor(() => {
      expect(mockGetResourceAllocation).toHaveBeenCalledWith('owner_workload');
      expect(screen.getByTestId('rb-statement')).toHaveTextContent(/100% of operational approvals are assigned to owner/i);
      expect(screen.getByTestId('metric-owner')).toHaveTextContent('kanadevarun123@gmail.com');
      expect(screen.getByTestId('metric-workload-count')).toHaveTextContent('23 items');
    });
  });

  it('switches to summary tab and renders unified 4-quadrant bottleneck grid', async () => {
    renderComponent();

    await waitFor(() => {
      expect(screen.getByTestId('rb-statement')).toBeInTheDocument();
    });

    fireEvent.click(screen.getByTestId('tab-summary'));

    await waitFor(() => {
      expect(mockGetResourceBottleneckSummary).toHaveBeenCalled();
      expect(screen.getByTestId('rb-summary-view')).toBeInTheDocument();
    });

    expect(screen.getByText('Commercial Approvals Queue')).toBeInTheDocument();
    expect(screen.getByText('Manifest & Document Clearance')).toBeInTheDocument();
    expect(screen.getByText('Customs Exception & Downstream Impact')).toBeInTheDocument();
    expect(screen.getAllByText('Owner Workload Imbalance').length).toBeGreaterThanOrEqual(1);
  });

  it('toggles evidence and methodology collapsible drawer', async () => {
    renderComponent();

    await waitFor(() => {
      expect(screen.getByTestId('rb-statement')).toBeInTheDocument();
    });

    const toggleBtn = screen.getByTestId('rb-evidence-toggle');
    fireEvent.click(toggleBtn);

    expect(screen.getByTestId('rb-evidence-body')).toBeInTheDocument();
    expect(screen.getByText(/Reasoning & Risk Summary/i)).toBeInTheDocument();
    expect(screen.getByText('pred-btln-appr-test-123')).toBeInTheDocument();
  });

  it('handles action dispatch modal and queues action via Action System', async () => {
    renderComponent();

    await waitFor(() => {
      expect(screen.getByTestId('rb-queue-action-btn')).toBeInTheDocument();
    });

    // Open modal
    fireEvent.click(screen.getByTestId('rb-queue-action-btn'));
    expect(screen.getByTestId('rb-action-modal')).toBeInTheDocument();

    // Submit action
    const submitBtn = screen.getByTestId('rb-modal-submit-btn');
    fireEvent.click(submitBtn);

    await waitFor(() => {
      expect(mockRequestAction).toHaveBeenCalledWith(
        'pred-btln-appr-test-123',
        expect.any(String)
      );
      expect(screen.getByTestId('rb-action-success')).toHaveTextContent(
        /Action queued for Human-in-the-Loop review/i
      );
    });
  });

  it('handles refresh button click and recalculates prediction', async () => {
    renderComponent();

    await waitFor(() => {
      expect(screen.getByTestId('rb-statement')).toBeInTheDocument();
    });

    const refreshBtn = screen.getByTestId('rb-refresh-btn');
    fireEvent.click(refreshBtn);

    await waitFor(() => {
      expect(mockRefreshOperationalBottleneck).toHaveBeenCalledWith('approvals');
    });
  });

  it('displays error state when intelligence service fails', async () => {
    mockGetOperationalBottleneck.mockRejectedValueOnce(new Error('Internal sidecar timeout'));
    renderComponent();

    await waitFor(() => {
      expect(screen.getByTestId('rb-error-state')).toBeInTheDocument();
    });

    expect(screen.getByText('Intelligence Engine Offline')).toBeInTheDocument();
    expect(screen.getByText('Internal sidecar timeout')).toBeInTheDocument();
  });
});

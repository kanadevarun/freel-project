import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import ShipmentReadinessPredictiveIntelligenceCard from '../../components/predictions/ShipmentReadinessPredictiveIntelligenceCard';
import * as predictionService from '../../services/predictionService';

const mockGetShipmentPredictedReadiness = vi.fn();
const mockRefreshShipmentPredictedReadiness = vi.fn();
const mockAcknowledgePrediction = vi.fn();
const mockRequestAction = vi.fn();

vi.mock('../../services/predictionService', () => {
  const service = {
    getShipmentPredictedReadiness: (...args) => mockGetShipmentPredictedReadiness(...args),
    refreshShipmentPredictedReadiness: (...args) => mockRefreshShipmentPredictedReadiness(...args),
    acknowledgePrediction: (...args) => mockAcknowledgePrediction(...args),
    requestAction: (...args) => mockRequestAction(...args),
  };
  return {
    default: service,
    predictionService: service,
    getShipmentPredictedReadiness: service.getShipmentPredictedReadiness,
    refreshShipmentPredictedReadiness: service.refreshShipmentPredictedReadiness,
    acknowledgePrediction: service.acknowledgePrediction,
    requestAction: service.requestAction,
  };
});

describe('ShipmentReadinessPredictiveIntelligenceCard', () => {
  const mockDocumentationRiskPrediction = {
    prediction_id: 'pred-ship-read-101-test',
    module: 'shipments',
    prediction_type: 'DOCUMENTATION_DELAY_RISK',
    prediction_category: 'DOCUMENTATION_DELAY_RISK',
    related_record_type: 'SHIPMENT',
    related_record_id: '101',
    severity: 'HIGH',
    confidence_score: 0.92,
    confidence_band: 'HIGH',
    prediction_statement: "Documentation Discrepancy & Manifest Cutoff Risk: Shipment #BK-2026-DEV-001 has 1 open document discrepancy between MBL and HBL (gross_weight: 24500.0 vs 21200) with missing packing list.",
    explanation: "Gross weight variance of 3,300 kg between MBL and HBL creates critical manifest mismatch.",
    time_horizon: '14_DAYS',
    status: 'PUBLISHED',
    milestone_reference: 'ARRIVAL',
    cutoff_reference: 'IMPORT_MANIFEST_DEADLINE',
    document_reference: 'maersk_mbl_clean.pdf vs mismatched_hbl_001.pdf',
    supporting_signals: [
      { signal_name: 'open_discrepancy_count', observed_value: '1 Flag', baseline_value: '0', importance_weight: 0.95 },
      { signal_name: 'missing_document_count', observed_value: '1', baseline_value: '0', importance_weight: 0.88 },
      { signal_name: 'milestone_reference', observed_value: 'ARRIVAL', baseline_value: 'ARRIVAL', importance_weight: 0.90 },
      { signal_name: 'cutoff_reference', observed_value: 'IMPORT_MANIFEST_DEADLINE', baseline_value: 'STANDARD', importance_weight: 0.92 },
    ],
    source_references: [
      {
        source_module: 'shipments',
        source_record_id: '101',
        source_field: 'shipments.id',
        source_timestamp: '2026-09-10T12:00:00Z',
        milestone_reference: 'ARRIVAL',
        cutoff_reference: 'IMPORT_MANIFEST_DEADLINE',
      },
      {
        source_module: 'shipment_document_discrepancies',
        source_record_id: '5',
        source_field: 'shipment_document_discrepancies.gross_weight',
        source_timestamp: '2026-09-10T12:00:00Z',
        document_reference: 'mismatched_hbl_001.pdf',
      },
    ],
    source_timestamp: '2026-09-10T12:00:00Z',
    recommended_action: 'Request shipper confirmation on gross weight variance before vessel arrival.',
    action_type: 'shipments.request_document_review',
    is_action_required: true,
    requires_approval: true,
    model_version: 'ai-sidecar-shipments-v1',
  };

  const mockCustomsHoldPrediction = {
    prediction_id: 'pred-ship-read-103-test',
    module: 'shipments',
    prediction_type: 'CUSTOMS_PROCESSING_RISK',
    prediction_category: 'CUSTOMS_PROCESSING_RISK',
    related_record_type: 'SHIPMENT',
    related_record_id: '103',
    severity: 'CRITICAL',
    confidence_score: 0.95,
    confidence_band: 'HIGH',
    prediction_statement: 'Customs Clearance & Regulatory Hold Risk: Shipment #BK-2026-DEV-003 on active customs hold at INNSA with open inspection flag.',
    explanation: 'Active customs hold with critical exception 101 on HS code declarations.',
    time_horizon: '14_DAYS',
    status: 'PUBLISHED',
    milestone_reference: 'CUSTOMS_CLEARANCE',
    cutoff_reference: 'CUSTOMS_SUBMISSION_DEADLINE',
    document_reference: 'bill_of_lading_103.pdf',
    supporting_signals: [
      { signal_name: 'customs_hold_flag', observed_value: 'ACTIVE', baseline_value: 'CLEARED', importance_weight: 1.0 },
      { signal_name: 'open_exception_count', observed_value: '1', baseline_value: '0', importance_weight: 0.96 },
    ],
    source_references: [
      { source_module: 'shipments', source_record_id: '103', source_field: 'shipments.status', source_timestamp: '2026-09-10T12:00:00Z' },
      { source_module: 'shipment_exceptions', source_record_id: '101', source_field: 'shipment_exceptions.exception_type', source_timestamp: '2026-09-10T12:00:00Z' },
    ],
    source_timestamp: '2026-09-10T12:00:00Z',
    recommended_action: 'Submit corrected tariff declaration to customs authority immediately.',
    action_type: 'shipments.escalate_customs_hold',
    is_action_required: true,
    requires_approval: true,
    model_version: 'ai-sidecar-shipments-v1',
  };

  const mockInsufficientPrediction = {
    prediction_id: 'pred-ship-read-insufficient',
    module: 'shipments',
    prediction_type: 'SHIPMENT_READINESS_RISK',
    severity: 'LOW',
    confidence_score: 0.2,
    confidence_band: 'LOW',
    prediction_statement: 'Insufficient Data: Missing milestone and vessel schedule.',
    insufficient_data: true,
    insufficient_data_reason: 'No active milestones or document records found for this shipment.',
    supporting_signals: [],
    source_references: [],
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders loading skeleton initially', () => {
    mockGetShipmentPredictedReadiness.mockReturnValue(new Promise(() => {}));
    render(<ShipmentReadinessPredictiveIntelligenceCard shipmentId="101" />);
    expect(screen.getByText(/Predictive Documentation & Readiness Intelligence/i)).toBeInTheDocument();
    expect(screen.getByText(/Evaluating operational compliance and cutoff milestones/i)).toBeInTheDocument();
  });

  it('renders error banner when request fails', async () => {
    mockGetShipmentPredictedReadiness.mockRejectedValue(new Error('Network timeout'));
    render(<ShipmentReadinessPredictiveIntelligenceCard shipmentId="101" />);
    await waitFor(() => {
      expect(screen.getByText(/Network timeout/i)).toBeInTheDocument();
    });
  });

  it('renders documentation discrepancy & cutoff miss intelligence for Shipment 101', async () => {
    mockGetShipmentPredictedReadiness.mockResolvedValue(mockDocumentationRiskPrediction);
    render(
      <ShipmentReadinessPredictiveIntelligenceCard
        shipmentId="101"
        shipment={{ status: 'DEPARTED' }}
      />
    );

    await waitFor(() => {
      expect(screen.getByText(/Documentation Discrepancy & Cutoff Risk/i)).toBeInTheDocument();
    });

    // Authoritative 4-metric grid
    expect(screen.getByText('DEPARTED')).toBeInTheDocument();
    expect(screen.getByText('ARRIVAL')).toBeInTheDocument();
    expect(screen.getAllByText(/Cutoff: IMPORT_MANIFEST_DEADLINE/i).length).toBeGreaterThanOrEqual(1);
    expect(screen.getByText('1 Flag')).toBeInTheDocument();

    // Prediction statement & explanation
    expect(screen.getByText(/Gross weight variance of 3,300 kg/i)).toBeInTheDocument();

    // Source grounding tags
    expect(screen.getByText(/Milestone: ARRIVAL/i)).toBeInTheDocument();
    expect(screen.getByText(/Doc: maersk_mbl_clean.pdf vs mismatched_hbl_001.pdf/i)).toBeInTheDocument();

    // Action buttons
    expect(screen.getByRole('button', { name: /Acknowledge/i })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /Queue Recommended Action/i })).toBeInTheDocument();
  });

  it('renders customs hold & regulatory risk for Shipment 103', async () => {
    mockGetShipmentPredictedReadiness.mockResolvedValue(mockCustomsHoldPrediction);
    render(
      <ShipmentReadinessPredictiveIntelligenceCard
        shipmentId="103"
        shipment={{ status: 'CUSTOMS_HOLD' }}
      />
    );

    await waitFor(() => {
      expect(screen.getByText(/Customs & Regulatory Hold Risk/i)).toBeInTheDocument();
    });

    expect(screen.getByText('CUSTOMS_HOLD')).toBeInTheDocument();
    expect(screen.getByText('CUSTOMS_CLEARANCE')).toBeInTheDocument();
    expect(screen.getAllByText(/Cutoff: CUSTOMS_SUBMISSION_DEADLINE/i).length).toBeGreaterThanOrEqual(1);
    expect(screen.getByText('Hold Encountered')).toBeInTheDocument();
    expect(screen.getByText(/Milestone: CUSTOMS_CLEARANCE/i)).toBeInTheDocument();
  });

  it('renders insufficient data state when records are missing', async () => {
    mockGetShipmentPredictedReadiness.mockResolvedValue(mockInsufficientPrediction);
    render(<ShipmentReadinessPredictiveIntelligenceCard shipmentId="999" />);

    await waitFor(() => {
      expect(screen.getByText(/Insufficient Shipment or Documentation Records/i)).toBeInTheDocument();
    });
    expect(screen.getByText(/No active milestones or document records found/i)).toBeInTheDocument();
  });

  it('triggers force refresh when refresh button is clicked', async () => {
    mockGetShipmentPredictedReadiness.mockResolvedValue(mockDocumentationRiskPrediction);
    mockRefreshShipmentPredictedReadiness.mockResolvedValue({
      ...mockDocumentationRiskPrediction,
      prediction_id: 'pred-ship-read-101-refreshed',
    });

    render(<ShipmentReadinessPredictiveIntelligenceCard shipmentId="101" />);

    await waitFor(() => {
      expect(screen.getByText(/Documentation Discrepancy & Cutoff Risk/i)).toBeInTheDocument();
    });

    const refreshBtn = screen.getByRole('button', { name: /Refresh/i });
    fireEvent.click(refreshBtn);

    await waitFor(() => {
      expect(mockRefreshShipmentPredictedReadiness).toHaveBeenCalledWith('101');
    });
  });

  it('handles acknowledge and queue action workflow', async () => {
    mockGetShipmentPredictedReadiness.mockResolvedValue(mockDocumentationRiskPrediction);
    mockAcknowledgePrediction.mockResolvedValue({ success: true });
    mockRequestAction.mockResolvedValue({ success: true });

    render(<ShipmentReadinessPredictiveIntelligenceCard shipmentId="101" />);

    await waitFor(() => {
      expect(screen.getByRole('button', { name: /Acknowledge/i })).toBeInTheDocument();
    });

    // Test Acknowledge
    const ackBtn = screen.getByRole('button', { name: /Acknowledge/i });
    fireEvent.click(ackBtn);

    await waitFor(() => {
      expect(mockAcknowledgePrediction).toHaveBeenCalledWith('pred-ship-read-101-test');
      expect(screen.getByText(/Prediction acknowledged successfully/i)).toBeInTheDocument();
    });

    // Test Queue Action
    const actionBtn = screen.getByRole('button', { name: /Queue Recommended Action/i });
    fireEvent.click(actionBtn);

    await waitFor(() => {
      expect(mockRequestAction).toHaveBeenCalledWith(
        'pred-ship-read-101-test',
        expect.stringContaining('Action queued for operational review')
      );
      expect(screen.getByText(/Action queued for Human-in-the-Loop review/i)).toBeInTheDocument();
      expect(screen.getByText('Action In Review')).toBeInTheDocument();
    });
  });

  it('toggles collapsible telemetry & audit trails accordion', async () => {
    mockGetShipmentPredictedReadiness.mockResolvedValue(mockDocumentationRiskPrediction);
    render(<ShipmentReadinessPredictiveIntelligenceCard shipmentId="101" />);

    await waitFor(() => {
      expect(screen.getByText(/Operational Grounding & Audit Signals/i)).toBeInTheDocument();
    });

    expect(screen.queryByText('Quantitative Operational Signals')).not.toBeInTheDocument();

    const toggleBtn = screen.getByText(/Operational Grounding & Audit Signals/i);
    fireEvent.click(toggleBtn);

    expect(screen.getByText('Quantitative Operational Signals')).toBeInTheDocument();
    expect(screen.getByText('open_discrepancy_count')).toBeInTheDocument();
    expect(screen.getByText('Authoritative Source Audit Trails')).toBeInTheDocument();
    expect(screen.getByText('Record #101')).toBeInTheDocument();
  });
});

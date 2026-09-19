import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import ContractCompliancePredictiveIntelligenceCard from '../../components/predictions/ContractCompliancePredictiveIntelligenceCard';
import * as predictionService from '../../services/predictionService';

const mockGetContractPredictedComplianceRisk = vi.fn();
const mockRefreshContractPredictedComplianceRisk = vi.fn();
const mockAcknowledgePrediction = vi.fn();
const mockRequestAction = vi.fn();

vi.mock('../../services/predictionService', () => {
  const service = {
    getContractPredictedComplianceRisk: (...args) => mockGetContractPredictedComplianceRisk(...args),
    refreshContractPredictedComplianceRisk: (...args) => mockRefreshContractPredictedComplianceRisk(...args),
    acknowledgePrediction: (...args) => mockAcknowledgePrediction(...args),
    requestAction: (...args) => mockRequestAction(...args),
  };
  return {
    default: service,
    predictionService: service,
    getContractPredictedComplianceRisk: service.getContractPredictedComplianceRisk,
    refreshContractPredictedComplianceRisk: service.refreshContractPredictedComplianceRisk,
    acknowledgePrediction: service.acknowledgePrediction,
    requestAction: service.requestAction,
  };
});

describe('ContractCompliancePredictiveIntelligenceCard', () => {
  const mockClauseRiskPrediction = {
    prediction_id: 'pred-ctr-comp-101-test',
    module: 'contracts',
    prediction_type: 'CONTRACT_CLAUSE_COMMERCIAL_RISK',
    prediction_category: 'CONTRACT_CLAUSE_COMMERCIAL_RISK',
    related_record_type: 'CONTRACT',
    related_record_id: '101',
    severity: 'MEDIUM',
    confidence_score: 0.91,
    confidence_band: 'HIGH',
    prediction_statement: "Clause Discrepancy & Documentation Risk: Contract #CTR-TP-2026-01 has 1 structured term discrepancies and 2 flagged clauses in document 'maersk_contract_2026.pdf' (Clause CLAUSE-LIA-02).",
    explanation: "Document analysis on 'maersk_contract_2026.pdf' identified extracted provisions that diverge from master structured records.",
    time_horizon: '30_DAYS',
    status: 'PUBLISHED',
    clause_reference: 'CLAUSE-LIA-02',
    document_reference: 'maersk_contract_2026.pdf',
    page_number: 6,
    section_heading: 'Carrier Limitation of Cargo Liability',
    supporting_signals: [
      { signal_name: 'flagged_clauses_count', observed_value: '2', baseline_value: '0', importance_weight: 0.92 },
      { signal_name: 'structured_discrepancies', observed_value: '1', baseline_value: '0', importance_weight: 0.90 }
    ],
    source_references: [
      {
        source_module: 'contracts',
        source_record_id: '101',
        source_field: 'contract_documents.status',
        source_timestamp: '2026-09-10T12:00:00Z',
        document_reference: 'maersk_contract_2026.pdf',
        clause_reference: 'CLAUSE-LIA-02',
        page_number: 6,
        section_heading: 'Carrier Limitation of Cargo Liability'
      }
    ],
    source_timestamp: '2026-09-10T12:00:00Z',
    recommended_action: 'Conduct formal legal and compliance clause review for CTR-TP-2026-01 in Document Review Center.',
    action_type: 'contracts.review_clause_discrepancy',
    is_action_required: true,
    requires_approval: true,
    model_version: 'ai-sidecar-contracts-v1'
  };

  const mockExpiryRiskPrediction = {
    prediction_id: 'pred-ctr-comp-103-test',
    module: 'contracts',
    prediction_type: 'CONTRACT_EXPIRY_RENEWAL_RISK',
    prediction_category: 'CONTRACT_EXPIRY_RENEWAL_RISK',
    related_record_type: 'CONTRACT',
    related_record_id: '103',
    severity: 'MEDIUM',
    confidence_score: 0.94,
    confidence_band: 'HIGH',
    prediction_statement: 'Impending Contract Expiry: Agreement #VEND-DRAY-2025-Q3 expires in 14 days on 2026-09-25.',
    explanation: 'With only 14 calendar days remaining in the validity window, failure to finalize successor terms will lead to spot cost exposure.',
    time_horizon: '30_DAYS',
    status: 'PUBLISHED',
    supporting_signals: [
      { signal_name: 'days_until_expiry', observed_value: '14', baseline_value: '30+ days', importance_weight: 0.98 },
      { signal_name: 'contract_status', observed_value: 'ACTIVE', baseline_value: 'ACTIVE', importance_weight: 1.0 }
    ],
    source_references: [
      { source_module: 'contracts', source_record_id: '103', source_field: 'contracts.expiry_date', source_timestamp: '2026-09-10T12:00:00Z' }
    ],
    source_timestamp: '2026-09-10T12:00:00Z',
    recommended_action: 'Queue renewal workflow and request formal renewal review.',
    is_action_required: true,
    requires_approval: true,
    model_version: 'ai-sidecar-contracts-v1'
  };

  const mockInsufficientDataPrediction = {
    prediction_id: 'pred-ctr-comp-109-insufficient',
    module: 'contracts',
    prediction_type: 'COMPLIANCE_REVIEW_RISK',
    prediction_category: 'INSUFFICIENT_DATA',
    related_record_type: 'CONTRACT',
    related_record_id: '109',
    severity: 'LOW',
    confidence_score: 0.15,
    confidence_band: 'LOW',
    prediction_statement: 'Insufficient Contract Intelligence Data: Required validity dates, structured terms, or compliance records unavailable.',
    explanation: 'Input records contain insufficient documentation or validity timestamps to produce a reliable risk prediction.',
    insufficient_data: true,
    insufficient_data_reason: 'Missing contract documents and compliance verification logs.',
    time_horizon: '30_DAYS',
    status: 'PUBLISHED',
    supporting_signals: [],
    source_references: [],
    source_timestamp: '2026-09-10T12:00:00Z',
    is_action_required: false,
    requires_approval: false,
    model_version: 'ai-sidecar-contracts-v1'
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders loading state initially', () => {
    mockGetContractPredictedComplianceRisk.mockReturnValue(new Promise(() => {}));
    render(<ContractCompliancePredictiveIntelligenceCard contractId={101} />);
    expect(screen.getByText(/Synthesizing predictive contract & compliance intelligence/i)).toBeInTheDocument();
  });

  it('renders contract clause and compliance risk with citations and 4 metrics', async () => {
    mockGetContractPredictedComplianceRisk.mockResolvedValue(mockClauseRiskPrediction);

    render(
      <ContractCompliancePredictiveIntelligenceCard
        contractId={101}
        contract={{ id: 101, status: 'DRAFT', contract_name: 'Trans-Pacific Master Agreement' }}
      />
    );

    await waitFor(() => {
      expect(screen.getByText(/Predictive Contract & Compliance Intelligence/i)).toBeInTheDocument();
    });

    // Metric tiles
    expect(screen.getByText(/AUTHORITATIVE STATUS/i)).toBeInTheDocument();
    expect(screen.getByText(/VALIDITY TIMELINE/i)).toBeInTheDocument();
    expect(screen.getByText(/COMPLIANCE POSTURE/i)).toBeInTheDocument();
    expect(screen.getByText(/DOCUMENT & CLAUSE CITATION/i)).toBeInTheDocument();

    // Clause citation
    expect(screen.getAllByText(/CLAUSE-LIA-02/i).length).toBeGreaterThan(0);
    expect(screen.getAllByText(/maersk_contract_2026.pdf/i).length).toBeGreaterThan(0);
    expect(screen.getByText(/Page 6/i)).toBeInTheDocument();

    // Prediction statement
    expect(screen.getByText(/Clause Discrepancy & Documentation Risk/i)).toBeInTheDocument();
    expect(screen.getByText(/Advisory Forecast/i)).toBeInTheDocument();

    // Action button
    expect(screen.getByText(/Queue Recommended Action/i)).toBeInTheDocument();
  });

  it('renders impending expiry risk correctly', async () => {
    mockGetContractPredictedComplianceRisk.mockResolvedValue(mockExpiryRiskPrediction);

    render(
      <ContractCompliancePredictiveIntelligenceCard
        contractId={103}
        contract={{ id: 103, status: 'ACTIVE', contract_name: 'Tri-State Drayage' }}
      />
    );

    await waitFor(() => {
      expect(screen.getByText(/Impending Contract Expiry/i)).toBeInTheDocument();
    });

    expect(screen.getByText(/14 Days Left/i)).toBeInTheDocument();
    expect(screen.getByText(/Expiry & Renewal Window/i)).toBeInTheDocument();
  });

  it('renders insufficient data state cleanly', async () => {
    mockGetContractPredictedComplianceRisk.mockResolvedValue(mockInsufficientDataPrediction);

    render(<ContractCompliancePredictiveIntelligenceCard contractId={109} />);

    await waitFor(() => {
      expect(screen.getByText(/Insufficient Historical or Document Records/i)).toBeInTheDocument();
    });

    expect(screen.getByText(/Missing contract documents and compliance verification logs/i)).toBeInTheDocument();
  });

  it('handles operator acknowledgment', async () => {
    mockGetContractPredictedComplianceRisk.mockResolvedValue(mockClauseRiskPrediction);
    mockAcknowledgePrediction.mockResolvedValue({ success: true });

    render(<ContractCompliancePredictiveIntelligenceCard contractId={101} />);

    await waitFor(() => {
      expect(screen.getByText(/Acknowledge/i)).toBeInTheDocument();
    });

    fireEvent.click(screen.getByText(/Acknowledge/i));

    await waitFor(() => {
      expect(mockAcknowledgePrediction).toHaveBeenCalledWith('pred-ctr-comp-101-test');
      expect(screen.getByText(/acknowledged by compliance review team/i)).toBeInTheDocument();
    });
  });

  it('handles operator action request in Action System', async () => {
    mockGetContractPredictedComplianceRisk.mockResolvedValue(mockClauseRiskPrediction);
    mockRequestAction.mockResolvedValue({ success: true });

    render(<ContractCompliancePredictiveIntelligenceCard contractId={101} />);

    await waitFor(() => {
      expect(screen.getByText(/Queue Recommended Action/i)).toBeInTheDocument();
    });

    fireEvent.click(screen.getByText(/Queue Recommended Action/i));

    await waitFor(() => {
      expect(mockRequestAction).toHaveBeenCalledWith(
        'pred-ctr-comp-101-test',
        expect.objectContaining({
          action_type: 'contracts.review_clause_discrepancy'
        })
      );
      expect(screen.getByText(/Compliance review task queued in Action System/i)).toBeInTheDocument();
    });
  });

  it('expands telemetry and audit grounding accordion', async () => {
    mockGetContractPredictedComplianceRisk.mockResolvedValue(mockClauseRiskPrediction);

    render(<ContractCompliancePredictiveIntelligenceCard contractId={101} />);

    await waitFor(() => {
      expect(screen.getByText(/Audit Grounding & Telemetry Signals/i)).toBeInTheDocument();
    });

    const toggleBtn = screen.getByText(/Audit Grounding & Telemetry Signals/i);
    fireEvent.click(toggleBtn);

    await waitFor(() => {
      expect(screen.getByText(/Quantitative Signals/i)).toBeInTheDocument();
      expect(screen.getByText(/Authoritative Source Audit Trails/i)).toBeInTheDocument();
      expect(screen.getByText(/flagged_clauses_count/i)).toBeInTheDocument();
    });
  });
});

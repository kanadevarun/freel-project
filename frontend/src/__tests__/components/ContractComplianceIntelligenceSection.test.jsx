import React from 'react';
import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import ContractComplianceIntelligenceSection from '../../pages/dashboard/Contracts/ContractComplianceIntelligenceSection';

const mockGetContract360ComplianceIntelligence = vi.fn();

vi.mock('../../services/contractsService', () => ({
  contractsService: {
    getContract360ComplianceIntelligence: (...args) => mockGetContract360ComplianceIntelligence(...args),
  },
  default: {
    getContract360ComplianceIntelligence: (...args) => mockGetContract360ComplianceIntelligence(...args),
  },
}));

describe('ContractComplianceIntelligenceSection Component', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  const mockPayload = {
    contract_id: 10,
    organization_scope: 1,
    correlation_id: 'corr-contract-unit-test-101',
    calculated_at: '2026-09-07T14:30:00Z',
    read_only: true,
    identity: {
      contract_id: 10,
      contract_number: 'CNT-2026-SH-LAX-01',
      contract_type: 'CARRIER_MSA',
      party_type: 'CARRIER',
      party_id: 42,
      party_name: 'Maersk Ocean Line',
      owner_name: 'Sarah Jenkins',
      status: 'ACTIVE',
      effective_date: '2026-01-01T00:00:00Z',
      expiry_date: '2026-12-31T00:00:00Z',
      days_until_expiry: 115,
      is_expired: false,
      is_expiring_soon: false,
      currency: 'USD',
      origin_port: 'Shanghai (CNSHA)',
      destination_port: 'Los Angeles (USLAX)',
      lane: 'CNSHA-USLAX',
      transport_mode: 'OCEAN',
      equipment_type: "40' High Cube Dry",
      completeness_score: 95,
      has_active_agreement_document: true,
      has_active_rate_table: true,
      data_freshness: '2026-09-07T14:30:00Z',
    },
    commercial_terms: {
      base_rates_count: 4,
      accessorial_charges_count: 2,
      surcharges_count: 3,
      has_minimum_commitments: true,
      minimum_commitment_details: 'Minimum 500 FEU annual volume guarantee',
      free_time_days: 7,
      demurrage_terms: '$150/day after 7 free days at destination port',
      detention_terms: '$175/day after 5 free days container turnaround',
      payment_terms: 'NET_30',
      credit_limit: 250000.0,
      currency: 'USD',
      has_expired_rates: false,
    },
    obligations: {
      total_obligations: 6,
      active_obligations: 5,
      completed_obligations: 1,
      overdue_obligations: 0,
      sla_fulfillment_rate: 98.5,
      open_penalties_count: 0,
      has_liability_terms: true,
      liability_clauses_summary: 'Standard carrier liability capped at $500 per package under COGSA.',
    },
    compliance: {
      total_requirements: 4,
      valid_requirements: 3,
      expired_requirements: 0,
      pending_review_requirements: 1,
      total_compliance_events: 1,
      open_compliance_events: 0,
      high_severity_events: 0,
      missing_required_documents: ['Customs Bond Verification Renewal (pending annual filing)'],
      has_unreviewed_clauses: false,
      compliance_status: 'PARTIALLY_VERIFIED',
      last_compliance_review: '2026-08-20T10:00:00Z',
    },
    coverage: {
      supported_modes: ['OCEAN'],
      origin_ports: ['Shanghai (CNSHA)'],
      destination_ports: ['Los Angeles (USLAX)'],
      coverage_status: 'ACTIVE_COVERAGE',
      matching_shipments_count: 14,
      matching_invoices_count: 8,
      invoices_outside_coverage_count: 0,
    },
    risks: {
      contract_expired: false,
      contract_expiring_soon: false,
      renewal_date_missing: false,
      missing_document: false,
      missing_rate_table: false,
      conflicting_dates: false,
      rate_validity_expired: false,
      has_high_severity_compliance_issues: false,
      has_overdue_obligations: false,
      has_unreviewed_clauses: false,
      risk_level: 'LOW',
      risk_score: 10,
      active_risk_count: 1,
      risk_alerts: [
        {
          code: 'PENDING_DOCUMENT_REVIEW',
          severity: 'MEDIUM',
          title: 'Customs Bond Pending Annual Filing',
          description: 'Customs Bond Verification Renewal is pending annual documentation upload.',
        },
      ],
    },
    ai_summary: {
      contract_executive_summary: 'Master ocean carrier agreement with Maersk providing comprehensive Transpacific coverage.',
      coverage_evaluation: 'Applicable to all standard ocean container shipments on the Shanghai to Los Angeles corridor.',
      commercial_term_validity: 'Base rates valid through December 31, 2026. NET 30 payment terms and 7 free days demurrage.',
      compliance_risk_assessment: 'Agreement is in solid standing. One low-priority customs bond document pending routine review.',
      human_review_required: false,
      recommended_actions: [
        'Upload updated customs bond verification certificate prior to next audit cycle.',
        'Review volume commitment trajectory at Q3 close to ensure 500 FEU threshold.',
      ],
      confidence_score: 0.96,
      generated_at: '2026-09-07T14:30:00Z',
    },
    warnings: [],
    limitations: [
      'Deterministic analysis reflects records currently captured in LogisticsHQ. Independent legal counsel recommended for statutory filings.',
    ],
  };

  it('renders loading state initially', () => {
    mockGetContract360ComplianceIntelligence.mockReturnValue(new Promise(() => {}));
    render(<ContractComplianceIntelligenceSection contractId={10} />);

    expect(screen.getByText(/Assembling Contract & Compliance Intelligence 360°/i)).toBeInTheDocument();
  });

  it('handles error state with retry button', async () => {
    mockGetContract360ComplianceIntelligence.mockRejectedValue(new Error('Network error loading intelligence'));
    render(<ContractComplianceIntelligenceSection contractId={10} />);

    await waitFor(() => {
      expect(screen.getByText(/Intelligence Unavailable/i)).toBeInTheDocument();
      expect(screen.getByText(/Network error loading intelligence/i)).toBeInTheDocument();
    });

    const retryBtn = screen.getByRole('button', { name: /Retry Intelligence Query/i });
    expect(retryBtn).toBeInTheDocument();

    mockGetContract360ComplianceIntelligence.mockResolvedValueOnce({ data: mockPayload });
    fireEvent.click(retryBtn);

    await waitFor(() => {
      expect(screen.getByText('CNT-2026-SH-LAX-01')).toBeInTheDocument();
    });
  });

  it('renders complete contract & compliance intelligence details', async () => {
    mockGetContract360ComplianceIntelligence.mockResolvedValue({ data: mockPayload });
    render(<ContractComplianceIntelligenceSection contractId={10} />);

    await waitFor(() => {
      expect(screen.getByText('CNT-2026-SH-LAX-01')).toBeInTheDocument();
    });

    // Check Counterparty & Owner
    expect(screen.getByText(/Maersk Ocean Line/i)).toBeInTheDocument();
    expect(screen.getByText(/Sarah Jenkins/i)).toBeInTheDocument();
    expect(screen.getByText(/CNSHA-USLAX/i)).toBeInTheDocument();

    // Check Metrics & Completeness
    expect(screen.getByText('95%')).toBeInTheDocument();
    expect(screen.getByText('LOW RISK')).toBeInTheDocument();
    expect(screen.getByText('ACTIVE_COVERAGE')).toBeInTheDocument();

    // Check Commercial Terms
    expect(screen.getByText(/NET_30/i)).toBeInTheDocument();
    expect(screen.getByText(/7 days/i)).toBeInTheDocument();
    expect(screen.getByText(/\$150\/day/i)).toBeInTheDocument();
    expect(screen.getByText(/Minimum 500 FEU annual volume guarantee/i)).toBeInTheDocument();

    // Check Compliance
    expect(screen.getByText(/PARTIALLY VERIFIED/i)).toBeInTheDocument();
    expect(screen.getAllByText(/Customs Bond Verification Renewal/i).length).toBeGreaterThanOrEqual(1);

    // Check Risk Alert
    expect(screen.getByText(/Customs Bond Pending Annual Filing/i)).toBeInTheDocument();

    // Check AI Advisory
    expect(screen.getByText(/Master ocean carrier agreement with Maersk/i)).toBeInTheDocument();
    expect(screen.getByText(/96% Confidence/i)).toBeInTheDocument();
    expect(screen.getByText(/Upload updated customs bond verification certificate/i)).toBeInTheDocument();
  });

  it('handles empty state when contractId is not provided', () => {
    render(<ContractComplianceIntelligenceSection contractId={null} />);
    expect(screen.getByText(/No Contract Selected/i)).toBeInTheDocument();
  });

  it('triggers manual refresh when refresh button is clicked', async () => {
    mockGetContract360ComplianceIntelligence.mockResolvedValue({ data: mockPayload });
    render(<ContractComplianceIntelligenceSection contractId={10} />);

    await waitFor(() => {
      expect(screen.getByText('CNT-2026-SH-LAX-01')).toBeInTheDocument();
    });

    const refreshBtn = screen.getByTitle('Re-evaluate Contract & Compliance Intelligence');
    expect(refreshBtn).toBeInTheDocument();

    fireEvent.click(refreshBtn);
    expect(mockGetContract360ComplianceIntelligence).toHaveBeenCalledTimes(2);
  });
});

import React from 'react';
import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import CrossModuleInsightsSection from '../../components/dashboard/CrossModuleInsightsSection';

const mockGetCrossModuleInsights = vi.fn();

vi.mock('../../services/insightsService', () => ({
  insightsService: {
    getCrossModuleInsights: (...args) => mockGetCrossModuleInsights(...args),
  },
  default: {
    getCrossModuleInsights: (...args) => mockGetCrossModuleInsights(...args),
  },
}));

describe('CrossModuleInsightsSection Component', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  const mockPayload = {
    primary_entity: {
      entity_type: 'CUSTOMER',
      entity_id: 10,
      reference: 'Global Traders Inc.',
      module: 'customers',
    },
    organization_scope: 1,
    correlation_id: 'corr-test-insights-101',
    calculated_at: '2026-09-07T15:30:00Z',
    read_only: true,
    total_insights_count: 2,
    critical_count: 1,
    high_count: 1,
    moderate_count: 0,
    low_count: 0,
    connected_modules: ['customers', 'finance', 'shipments'],
    insights: [
      {
        insight_id: 'ins-cust-ops-1',
        insight_type: 'CUSTOMER_RISK',
        category: 'Credit & Operations Exposure',
        title: 'Customer Has 3 Overdue Invoices with 4 Active Shipments',
        explanation: 'Global Traders Inc. has outstanding overdue receivables of $24,500 across 3 invoices while maintaining 4 active operational shipments in transit.',
        severity: 'CRITICAL',
        priority_score: 90,
        confidence: 'HIGH',
        data_freshness: '2026-09-07T15:30:00Z',
        organization_scope: 1,
        primary_entity: {
          entity_type: 'CUSTOMER',
          entity_id: 10,
          reference: 'Global Traders Inc.',
          module: 'customers',
        },
        related_entities: [
          { entity_type: 'INVOICE', entity_id: 201, reference: 'INV-2026-0042', module: 'finance' },
          { entity_type: 'SHIPMENT', entity_id: 101, reference: 'SH-101', module: 'shipments' },
        ],
        evidence: [
          {
            source_module: 'finance',
            source_entity_id: 201,
            source_ref: 'INV-2026-0042',
            field_name: 'balance_due',
            observed_value: 24500.0,
            description: 'Total balance due past maturity',
          },
          {
            source_module: 'shipments',
            source_entity_id: 101,
            source_ref: 'SH-101',
            field_name: 'active_shipments',
            observed_value: 4,
            description: 'Active freight in transit',
          },
        ],
        rule_applied: 'RULE_CUST_OVERDUE_ACTIVE_OPS',
        is_deterministic: true,
        lifecycle_status: 'ACTIVE',
        suggested_human_action: 'Coordinate with finance lead to verify payment commitments before releasing delivery orders.',
        limitations: [],
        read_only: true,
      },
      {
        insight_id: 'ins-cust-ops-2',
        insight_type: 'COMMERCIAL_CONTRACT',
        category: 'Contract Coverage Gap',
        title: 'Customer Generating Operational Freight Without Master Service Contract',
        explanation: 'Global Traders Inc. has 4 active shipments running under spot rates with no registered active customer contract.',
        severity: 'HIGH',
        priority_score: 75,
        confidence: 'HIGH',
        data_freshness: '2026-09-07T15:30:00Z',
        organization_scope: 1,
        primary_entity: {
          entity_type: 'CUSTOMER',
          entity_id: 10,
          reference: 'Global Traders Inc.',
          module: 'customers',
        },
        related_entities: [],
        evidence: [],
        rule_applied: 'RULE_CUST_ACTIVE_OPS_NO_CONTRACT',
        is_deterministic: true,
        lifecycle_status: 'ACTIVE',
        suggested_human_action: 'Engage commercial sales director to negotiate master rate schedule.',
        limitations: [],
        read_only: true,
      },
    ],
    ai_synthesis: {
      executive_summary: 'Cross-module audit for Global Traders Inc. detected 2 actionable signals (1 critical, 1 high priority).',
      connected_situation_analysis: 'High financial exposure overlaps with active cargo transit and missing framework contracts.',
      tradeoffs_and_priorities: 'Synchronize collection communications with transit updates.',
      suggested_investigation_questions: [
        'Has customer confirmed remittance schedule for overdue invoices?',
        'Are all forward spot bookings protected by carrier liability terms?',
      ],
      confidence: 'HIGH',
      generated_at: '2026-09-07T15:30:00Z',
      limitations: ['Based strictly on active MariaDB records.'],
    },
    warnings: [],
    limitations: ['Deterministic analysis based on cross-module foreign keys.'],
  };

  it('renders loading state initially', () => {
    mockGetCrossModuleInsights.mockReturnValue(new Promise(() => {}));
    render(<CrossModuleInsightsSection entityType="CUSTOMER" entityId={10} />);

    expect(screen.getByText(/Evaluating cross-module signals & relationships/i)).toBeInTheDocument();
  });

  it('handles error state with retry button', async () => {
    mockGetCrossModuleInsights.mockRejectedValue(new Error('Internal network error'));
    render(<CrossModuleInsightsSection entityType="CUSTOMER" entityId={10} />);

    await waitFor(() => {
      expect(screen.getByText(/Insights Service Temporarily Unavailable/i)).toBeInTheDocument();
      expect(screen.getByText(/Internal network error/i)).toBeInTheDocument();
    });

    const retryBtn = screen.getByRole('button', { name: /Retry Connection/i });
    expect(retryBtn).toBeInTheDocument();

    mockGetCrossModuleInsights.mockResolvedValueOnce({ data: mockPayload });
    fireEvent.click(retryBtn);

    await waitFor(() => {
      expect(screen.getByText(/Customer Has 3 Overdue Invoices with 4 Active Shipments/i)).toBeInTheDocument();
    });
  });

  it('renders empty state when zero cross-module risks exist', async () => {
    mockGetCrossModuleInsights.mockResolvedValue({
      data: {
        total_insights_count: 0,
        critical_count: 0,
        high_count: 0,
        insights: [],
        ai_synthesis: null,
      },
    });

    render(<CrossModuleInsightsSection />);

    await waitFor(() => {
      expect(screen.getByText(/Zero Cross-Module Risks Detected/i)).toBeInTheDocument();
      expect(screen.getByText(/All Modules Synchronized/i)).toBeInTheDocument();
    });
  });

  it('renders complete cross-module insights with AI synthesis, evidence, and actions', async () => {
    mockGetCrossModuleInsights.mockResolvedValue({ data: mockPayload });
    render(<CrossModuleInsightsSection entityType="CUSTOMER" entityId={10} />);

    await waitFor(() => {
      expect(screen.getByText(/Customer Has 3 Overdue Invoices with 4 Active Shipments/i)).toBeInTheDocument();
    });

    // Check count pill
    expect(screen.getByText(/2 Signals \(1 Critical, 1 High\)/i)).toBeInTheDocument();

    // Check AI synthesis
    expect(screen.getByText(/Connected Cross-Module Synthesis/i)).toBeInTheDocument();
    expect(screen.getByText(/Cross-module audit for Global Traders Inc./i)).toBeInTheDocument();
    expect(screen.getByText(/Has customer confirmed remittance schedule for overdue invoices\?/i)).toBeInTheDocument();

    // Check Evidence & Actions
    expect(screen.getByText(/Coordinate with finance lead to verify payment commitments/i)).toBeInTheDocument();
    expect(screen.getByText(/balance due:/i)).toBeInTheDocument();
    expect(screen.getAllByText(/24,500/i).length).toBeGreaterThanOrEqual(1);
  });

  it('triggers manual refresh when refresh button is clicked', async () => {
    mockGetCrossModuleInsights.mockResolvedValue({ data: mockPayload });
    render(<CrossModuleInsightsSection entityType="CUSTOMER" entityId={10} />);

    await waitFor(() => {
      expect(screen.getByText(/Customer Has 3 Overdue Invoices with 4 Active Shipments/i)).toBeInTheDocument();
    });

    const refreshBtn = screen.getByTitle('Re-run cross-module evaluation rules');
    expect(refreshBtn).toBeInTheDocument();

    fireEvent.click(refreshBtn);
    expect(mockGetCrossModuleInsights).toHaveBeenCalledTimes(2);
  });
});

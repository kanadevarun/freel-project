import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import FinanceCollectionsPredictiveIntelligenceCard from '../../components/predictions/FinanceCollectionsPredictiveIntelligenceCard';

const mockGetInvoicePredictedCollection = vi.fn();
const mockRefreshInvoicePredictedCollection = vi.fn();
const mockAcknowledgePrediction = vi.fn();
const mockRequestAction = vi.fn();

vi.mock('../../services/predictionService', () => {
  const service = {
    getInvoicePredictedCollection: (...args) => mockGetInvoicePredictedCollection(...args),
    refreshInvoicePredictedCollection: (...args) => mockRefreshInvoicePredictedCollection(...args),
    acknowledgePrediction: (...args) => mockAcknowledgePrediction(...args),
    requestAction: (...args) => mockRequestAction(...args),
  };
  return {
    default: service,
    predictionService: service,
  };
});

describe('FinanceCollectionsPredictiveIntelligenceCard', () => {
  const mockSettledInvoicePrediction = {
    prediction_id: 'pred-fin-101-test',
    module: 'finance',
    prediction_type: 'COLLECTION_PRIORITY',
    prediction_category: 'COLLECTION_PRIORITY',
    related_record_type: 'INVOICE',
    related_record_id: '101',
    severity: 'LOW',
    confidence_score: 0.98,
    confidence_band: 'HIGH',
    prediction_statement: 'Settled Invoice: Invoice #INV-2026-DEV-001 has been fully settled ($3,200.00 USD paid in full).',
    explanation: 'Account demonstrates excellent settlement discipline with zero outstanding receivables.',
    time_horizon: '30_DAYS',
    status: 'PUBLISHED',
    supporting_signals: [
      { signal_name: 'total_amount', observed_value: '$3,200.00', baseline_value: 'N/A', importance_weight: 0.95 },
      { signal_name: 'paid_amount', observed_value: '$3,200.00', baseline_value: 'N/A', importance_weight: 0.95 },
      { signal_name: 'balance_due', observed_value: '$0.00', baseline_value: 'N/A', importance_weight: 0.98 },
      { signal_name: 'is_overdue', observed_value: false, baseline_value: 'N/A', importance_weight: 0.90 }
    ],
    source_references: [
      { source_module: 'customer_invoices', source_record_id: '101', source_field: 'balance_due', source_timestamp: '2026-09-10T10:00:00Z' }
    ],
    source_timestamp: '2026-09-10T10:00:00Z',
    recommended_action: 'Standard receipt acknowledgment dispatched.',
    is_action_required: false,
    requires_approval: false,
  };

  const mockOverdueInvoicePrediction = {
    prediction_id: 'pred-fin-103-test',
    module: 'finance',
    prediction_type: 'INVOICE_LATE_PAYMENT_RISK',
    prediction_category: 'COLLECTION_PRIORITY',
    related_record_type: 'INVOICE',
    related_record_id: '103',
    severity: 'HIGH',
    confidence_score: 0.94,
    confidence_band: 'HIGH',
    prediction_statement: 'Elevated Overdue & Late-Payment Risk: Invoice #INV-2026-DEV-003 is 26 days past due date with $4,500.00 USD outstanding.',
    explanation: 'Invoice is 26 days overdue with zero partial payments recorded. Immediate collections outreach recommended.',
    time_horizon: '30_DAYS',
    status: 'PUBLISHED',
    supporting_signals: [
      { signal_name: 'total_amount', observed_value: '$4,500.00', baseline_value: 'N/A', importance_weight: 0.95 },
      { signal_name: 'paid_amount', observed_value: '$0.00', baseline_value: 'N/A', importance_weight: 0.95 },
      { signal_name: 'balance_due', observed_value: '$4,500.00', baseline_value: 'N/A', importance_weight: 0.98 },
      { signal_name: 'days_until_due', observed_value: -26, baseline_value: '0', importance_weight: 0.92 },
      { signal_name: 'is_overdue', observed_value: true, baseline_value: 'N/A', importance_weight: 0.95 }
    ],
    source_references: [
      { source_module: 'customer_invoices', source_record_id: '103', source_field: 'balance_due', source_timestamp: '2026-09-10T10:00:00Z' }
    ],
    source_timestamp: '2026-09-10T10:00:00Z',
    recommended_action: 'Initiate formal credit control escalation and dispatch overdue collection notice to debtor.',
    action_type: 'finance.escalate_collection',
    is_action_required: true,
    requires_approval: true,
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders loading state initially', () => {
    mockGetInvoicePredictedCollection.mockReturnValue(new Promise(() => {}));
    render(<FinanceCollectionsPredictiveIntelligenceCard invoiceId={101} />);
    expect(screen.getByText(/Evaluating invoice aging, settlement history, and cash inflow timing/i)).toBeInTheDocument();
  });

  it('renders error state when fetch fails', async () => {
    mockGetInvoicePredictedCollection.mockRejectedValue(new Error('Network error'));
    render(<FinanceCollectionsPredictiveIntelligenceCard invoiceId={101} />);

    await waitFor(() => {
      expect(screen.getByText('Finance Intelligence Unavailable')).toBeInTheDocument();
      expect(screen.getByText('Network error')).toBeInTheDocument();
    });
  });

  it('renders settled invoice intelligence correctly with authoritative signals', async () => {
    mockGetInvoicePredictedCollection.mockResolvedValue(mockSettledInvoicePrediction);
    render(<FinanceCollectionsPredictiveIntelligenceCard invoiceId={101} />);

    await waitFor(() => {
      expect(screen.getByText('Predictive Cash Flow & Collections Intelligence')).toBeInTheDocument();
      expect(screen.getByText(/Settled Invoice/i)).toBeInTheDocument();
      expect(screen.getAllByText('$3,200.00').length).toBeGreaterThanOrEqual(1);
      expect(screen.getByText('$0.00')).toBeInTheDocument();
      expect(screen.getByText(/Evidence Grounding & Telemetry/i)).toBeInTheDocument();
    });
  });

  it('renders overdue collection priority with action approval button', async () => {
    mockGetInvoicePredictedCollection.mockResolvedValue(mockOverdueInvoicePrediction);
    render(<FinanceCollectionsPredictiveIntelligenceCard invoiceId={103} />);

    await waitFor(() => {
      expect(screen.getByText(/Elevated Overdue & Late-Payment Risk/i)).toBeInTheDocument();
      expect(screen.getByText('OVERDUE')).toBeInTheDocument();
      expect(screen.getAllByText('$4,500.00').length).toBeGreaterThanOrEqual(1);
      expect(screen.getByText('Recommended Financial Governance')).toBeInTheDocument();
      expect(screen.getByText('Approval Required')).toBeInTheDocument();
      expect(screen.getByText('Queue for Approval')).toBeInTheDocument();
    });
  });

  it('handles action request submission with HITL approval feedback', async () => {
    mockGetInvoicePredictedCollection.mockResolvedValue(mockOverdueInvoicePrediction);
    mockRequestAction.mockResolvedValue({ success: true, message: 'Action queued' });

    render(<FinanceCollectionsPredictiveIntelligenceCard invoiceId={103} />);

    const btn = await screen.findByRole('button', { name: /Queue for Approval/i });
    fireEvent.click(btn);

    await waitFor(() => {
      expect(mockRequestAction).toHaveBeenCalledWith(
        'pred-fin-103-test',
        expect.stringContaining('Initiate formal credit control escalation')
      );
      expect(screen.getByText(/Action routed to Financial Controller for HITL approval/i)).toBeInTheDocument();
    });
  });

  it('expands telemetry and displays grounding sources and signals', async () => {
    mockGetInvoicePredictedCollection.mockResolvedValue({ data: mockOverdueInvoicePrediction });
    render(<FinanceCollectionsPredictiveIntelligenceCard invoiceId={103} />);

    await waitFor(() => {
      expect(screen.getByText(/Evidence Grounding & Telemetry/i)).toBeInTheDocument();
    });

    fireEvent.click(screen.getByText(/Evidence Grounding & Telemetry/i));

    await waitFor(() => {
      expect(screen.getByText(/Quantitative Decision Signals/i)).toBeInTheDocument();
      expect(screen.getByText(/Grounding Database Records/i)).toBeInTheDocument();
      expect(screen.getByText(/total_amount/i)).toBeInTheDocument();
      expect(screen.getByText(/Record #103/i)).toBeInTheDocument();
    });
  });
});

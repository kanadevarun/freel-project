import React from 'react';
import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import FinanceCollectionsAutomationSection from '../../../pages/dashboard/Finance/components/FinanceCollectionsAutomationSection';
import financeCollectionsAutomationService from '../../../services/financeCollectionsAutomationService';

vi.mock('../../../services/financeCollectionsAutomationService', () => ({
  default: {
    getOverview: vi.fn(),
    analyzeReceivablesRisk: vi.fn(),
    prioritizeCollections: vi.fn(),
    getCustomerBehavior: vi.fn(),
    getRecommendations: vi.fn(),
    generateDraft: vi.fn(),
    updateDraft: vi.fn(),
    submitForApproval: vi.fn(),
  },
}));

describe('FinanceCollectionsAutomationSection', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders loading state initially', () => {
    financeCollectionsAutomationService.getOverview.mockReturnValue(new Promise(() => {}));
    render(<FinanceCollectionsAutomationSection invoiceId={3} />);
    expect(screen.getByTestId('collections-loading')).toBeInTheDocument();
  });

  it('renders deterministic financial KPIs, aging status, and signals chips upon successful fetch', async () => {
    const mockOverview = {
      invoice_id: 3,
      org_id: 1,
      invoice_number: 'INV-2026-0454',
      customer_id: 10,
      customer_name: 'Apex Global Logistics',
      currency: 'USD',
      invoice_amount: 32120.00,
      paid_amount: 0.00,
      due_date: '2026-08-15T00:00:00Z',
      deterministic_signals: {
        is_overdue: true,
        days_overdue: 24,
        aging_bucket: '1_30_DAYS',
        is_approaching_due_date: false,
        days_until_due: 0,
        outstanding_amount: 32120.00,
        has_partial_payment: false,
        is_high_value_overdue: true,
        is_long_overdue: false,
        is_disputed: false,
        missing_payment_info: false,
      },
      customer_context: {
        customer_id: 10,
        customer_name: 'Apex Global Logistics',
        total_outstanding_balance: 45000.00,
        total_overdue_balance: 32120.00,
        overdue_invoices_count: 2,
        credit_limit: 50000.00,
        is_credit_limit_breached: false,
        average_days_to_pay: 38,
      },
      latest_analysis: {
        id: 1,
        risk_level: 'HIGH',
        risk_score: 0.78,
        days_overdue: 24,
        aging_bucket: '1_30_DAYS',
        outstanding_amount: 32120.00,
        currency: 'USD',
        receivables_summary: 'Invoice INV-2026-0454 has an outstanding overdue balance of USD 32,120.00 (24 days overdue). Multi-signal analysis indicates high receivables risk.',
        key_risks: [
          {
            risk_factor: 'HIGH_VALUE_EXPOSURE',
            severity: 'HIGH',
            description: 'Receivable exceeds high-value threshold ($10,000)'
          }
        ],
        confidence_score: 0.90,
        correlation_id: 'corr-fin-test-1',
        created_at: '2026-09-08T12:00:00Z',
        evidence: {
          days_overdue: 24,
          aging_bucket: '1_30_DAYS',
          customer_overdue_count: 2,
          is_disputed: false,
        }
      },
      existing_drafts: [
        {
          id: 10,
          org_id: 1,
          invoice_id: 3,
          customer_id: 10,
          draft_type: 'OVERDUE_NOTICE',
          subject: 'Formal Notice: Overdue Invoice INV-2026-0454 ($32,120.00 USD)',
          message_body: 'Dear Apex Global Logistics Accounts Payable team,\n\nOur records indicate invoice INV-2026-0454 is overdue.',
          recipient_name: 'Accounts Payable',
          recipient_email: 'ap@apexlogistics.com',
          outstanding_amount: 32120.00,
          currency: 'USD',
          status: 'DRAFT',
          requires_approval: true,
          created_at: '2026-09-08T12:05:00Z',
          updated_at: '2026-09-08T12:05:00Z',
        }
      ]
    };

    financeCollectionsAutomationService.getOverview.mockResolvedValue({ data: mockOverview });

    render(<FinanceCollectionsAutomationSection invoiceId={3} />);

    await waitFor(() => {
      expect(screen.getByTestId('finance-collections-automation-section')).toBeInTheDocument();
    });

    // Check KPIs
    expect(screen.getByTestId('kpi-outstanding-balance')).toBeInTheDocument();
    expect(screen.getByTestId('kpi-aging-status')).toBeInTheDocument();
    expect(screen.getByTestId('kpi-risk-score')).toBeInTheDocument();
    expect(screen.getByTestId('kpi-customer-exposure')).toBeInTheDocument();

    // Check Authoritative signals
    expect(screen.getByText(/Invoice Overdue \(24d\)/i)).toBeInTheDocument();
    expect(screen.getByText(/High-Value Overdue Receivable/i)).toBeInTheDocument();

    // Check AI summary text
    expect(screen.getByText(/outstanding overdue balance of USD 32,120.00/i)).toBeInTheDocument();

    // Check Drafts Studio with preloaded active draft
    expect(screen.getByTestId('panel-drafts-studio')).toBeInTheDocument();
    expect(screen.getByDisplayValue(/Formal Notice: Overdue Invoice INV-2026-0454/i)).toBeInTheDocument();
  });

  it('triggers AI Receivables Risk Analysis when user clicks the button', async () => {
    const mockOverview = {
      invoice_id: 3,
      org_id: 1,
      currency: 'USD',
      invoice_amount: 32120.00,
      paid_amount: 0.00,
      deterministic_signals: {
        is_overdue: true,
        days_overdue: 24,
        aging_bucket: '1_30_DAYS',
        outstanding_amount: 32120.00,
      },
      customer_context: {
        customer_name: 'Apex Global Logistics',
      },
      existing_drafts: []
    };

    financeCollectionsAutomationService.getOverview.mockResolvedValue({ data: mockOverview });
    financeCollectionsAutomationService.analyzeReceivablesRisk.mockResolvedValue({
      data: {
        risk_level: 'HIGH',
        risk_score: 0.85,
        receivables_summary: 'Fresh analysis generated successfully',
      }
    });

    render(<FinanceCollectionsAutomationSection invoiceId={3} />);

    await waitFor(() => {
      expect(screen.getByText('Analyze Receivables Risk')).toBeInTheDocument();
    });

    fireEvent.click(screen.getByText('Analyze Receivables Risk'));

    await waitFor(() => {
      expect(financeCollectionsAutomationService.analyzeReceivablesRisk).toHaveBeenCalledWith(3);
    });
  });
});

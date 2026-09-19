import React from 'react';
import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import InvoiceFinanceIntelligenceSection from '../../pages/dashboard/Finance/components/InvoiceFinanceIntelligenceSection';
import { recommendationService } from '../../services/recommendationService';

const mockGetInvoice360FinanceIntelligence = vi.fn();

vi.mock('../../services/invoiceService', () => ({
  invoiceService: {
    getInvoice360FinanceIntelligence: (...args) => mockGetInvoice360FinanceIntelligence(...args),
  },
  default: {
    getInvoice360FinanceIntelligence: (...args) => mockGetInvoice360FinanceIntelligence(...args),
  },
}));

vi.mock('../../services/recommendationService', () => ({
  recommendationService: {
    listRecommendations: vi.fn(),
    getInvoiceEvidence: vi.fn(),
    generateDraft: vi.fn(),
    saveDraft: vi.fn(),
    getActionPreview: vi.fn(),
    requestApproval: vi.fn(),
  },
}));

describe('Task 2.5 Invoice & Collections Assistant Integration Tests', () => {
  const mockInvoiceData = {
    invoice_id: 1,
    identity: {
      invoice_id: 1,
      invoice_number: 'INV-2026-0456',
      customer_id: 1,
      customer_name: 'Global Traders Inc.',
      currency: 'USD',
      total_amount: 24650.0,
      paid_amount: 0.0,
      balance_due: 24650.0,
      is_overdue: true,
      days_overdue: 8,
      aging_bucket: '1-30_DAYS',
    },
    customer_receivables: {
      total_outstanding_amount: 111330.0,
      open_invoices_count: 6,
    },
    revenue_cost: {
      gross_margin_percentage: 20.9,
    },
    line_items: [],
    payments: [],
    risk_indicators: {
      overall_risk_rating: 'HIGH',
      overall_risk_score: 55,
      risk_factors: ['Invoice overdue by 8 days'],
    },
    ai_summary: {
      executive_summary: 'Invoice INV-2026-0456 is overdue by 8 days.',
      confidence: 'HIGH',
    },
  };

  const mockRecommendations = [
    {
      id: 201,
      title: 'Invoice Overdue: INV-2026-0456 (1-15 days)',
      description: 'Invoice INV-2026-0456 is 8 days overdue with balance of $24,650.00. Customer has 6 open delinquent invoices.',
      priority: 'high',
      aging_band: '1-15 days',
      confidence_score: 0.95,
      draft_type: 'overdue_payment_notice',
      draft_subject: '',
      draft_body: '',
      status: 'new',
      invoice_id: 1,
    },
  ];

  const mockEvidence = {
    invoice_id: 1,
    invoice_number: 'INV-2026-0456',
    customer_id: 1,
    customer_name: 'Global Traders Inc.',
    days_overdue: 8,
    aging_band: '1-15 days',
    is_high_value: true,
    has_line_item_discrepancy: false,
    has_status_balance_mismatch: false,
    customer_open_invoices_count: 6,
    customer_total_ar: 111330.0,
  };

  beforeEach(() => {
    vi.clearAllMocks();
    mockGetInvoice360FinanceIntelligence.mockResolvedValue({ data: mockInvoiceData });
    recommendationService.listRecommendations.mockResolvedValue({ data: { items: mockRecommendations } });
    recommendationService.getInvoiceEvidence.mockResolvedValue({ data: mockEvidence });
  });

  it('renders collections assistant section with recommendations, aging band, and evidence metrics', async () => {
    render(<InvoiceFinanceIntelligenceSection invoiceId={1} />);

    await waitFor(() => {
      expect(screen.getByTestId('invoice-collections-assistant-section')).toBeInTheDocument();
    });

    expect(screen.getByText('COLLECTIONS & FOLLOW-UP ASSISTANT')).toBeInTheDocument();
    expect(screen.getByText('Strictly Read-Only • No Auto-Send')).toBeInTheDocument();
    expect(screen.getByText('Invoice Overdue: INV-2026-0456 (1-15 days)')).toBeInTheDocument();
    expect(screen.getAllByText('1-15 days').length).toBeGreaterThanOrEqual(1);
    expect(screen.getByText('Confidence: 95%')).toBeInTheDocument();

    // Check evidence chips
    expect(screen.getAllByText('111,330.00 USD').length).toBeGreaterThanOrEqual(1);
    expect(screen.getByText('8d')).toBeInTheDocument();
  });

  it('generates on-demand editable draft and enforces read-only safety without automated dispatch', async () => {
    recommendationService.generateDraft.mockResolvedValue({
      id: 201,
      draft_subject: 'Payment Notice: Overdue Invoice INV-2026-0456',
      draft_body: 'Dear Global Traders Inc. Accounts Payable,\n\nOur records show invoice INV-2026-0456 is overdue.',
      draft_status: 'DRAFTED',
    });

    render(<InvoiceFinanceIntelligenceSection invoiceId={1} />);

    await waitFor(() => {
      expect(screen.getByTestId('btn-draft-201')).toBeInTheDocument();
    });

    // Click Draft Collection Notice button
    fireEvent.click(screen.getByTestId('btn-draft-201'));

    await waitFor(() => {
      expect(screen.getByTestId('collection-draft-modal')).toBeInTheDocument();
    });

    // Verify Read-Only Safety Notice
    expect(screen.getByText(/Read-Only Safety Notice:/i)).toBeInTheDocument();
    expect(screen.getByText(/NEVER automatically dispatched/i)).toBeInTheDocument();
    expect(screen.getByText(/no balances, terms, or ledgers will be mutated/i)).toBeInTheDocument();

    // Verify editable draft inputs
    const subjectInput = screen.getByTestId('draft-subject-input');
    const bodyInput = screen.getByTestId('draft-body-textarea');

    expect(subjectInput.value).toBe('Payment Notice: Overdue Invoice INV-2026-0456');
    expect(bodyInput.value).toContain('Dear Global Traders Inc.');

    // Edit the draft
    fireEvent.change(subjectInput, { target: { value: 'Urgent: Settlement Reminder INV-2026-0456' } });
    fireEvent.change(bodyInput, { target: { value: 'Updated message content for customer.' } });

    recommendationService.saveDraft.mockResolvedValue({ status: 'ok' });

    // Click Save Draft
    fireEvent.click(screen.getByTestId('btn-save-draft'));

    await waitFor(() => {
      expect(recommendationService.saveDraft).toHaveBeenCalledWith(
        201,
        'Urgent: Settlement Reminder INV-2026-0456',
        'Updated message content for customer.'
      );
    });

    expect(screen.getByText('Collection draft saved to recommendation record')).toBeInTheDocument();
  });

  it('opens Action Execution Preview with operator checklist and zero-mutation guarantee', async () => {
    recommendationService.getActionPreview.mockResolvedValue({
      action_type: 'review_overdue_invoice',
      title: 'Review Overdue Receivables Position',
      description: 'Audit invoice settlement delays before customer escalation.',
      target_entity: 'Invoice INV-2026-0456',
      operator_checklist: [
        'Confirm bank remittance statements before customer outreach',
        'Verify if delivery exception caused payment hold',
      ],
      safety_notice: 'Zero financial records will be automatically modified.',
    });

    render(<InvoiceFinanceIntelligenceSection invoiceId={1} />);

    await waitFor(() => {
      expect(screen.getByTestId('btn-preview-201')).toBeInTheDocument();
    });

    fireEvent.click(screen.getByTestId('btn-preview-201'));

    await waitFor(() => {
      expect(screen.getByTestId('action-preview-modal')).toBeInTheDocument();
    });

    expect(screen.getByText('Action Execution Preview')).toBeInTheDocument();
    expect(screen.getByText('Review Overdue Receivables Position')).toBeInTheDocument();
    expect(screen.getByText('Confirm bank remittance statements before customer outreach')).toBeInTheDocument();
    expect(screen.getByText(/Zero Mutation Guarantee:/i)).toBeInTheDocument();
  });

  it('submits recommendation for human-in-the-loop finance approval sign-off', async () => {
    recommendationService.requestApproval.mockResolvedValue({ status: 'ok' });

    render(<InvoiceFinanceIntelligenceSection invoiceId={1} />);

    await waitFor(() => {
      expect(screen.getByTestId('btn-approval-201')).toBeInTheDocument();
    });

    fireEvent.click(screen.getByTestId('btn-approval-201'));

    await waitFor(() => {
      expect(recommendationService.requestApproval).toHaveBeenCalledWith(
        201,
        'Finance sign-off requested by operator'
      );
      expect(screen.getByText('Under Finance Review')).toBeInTheDocument();
    });
  });
});

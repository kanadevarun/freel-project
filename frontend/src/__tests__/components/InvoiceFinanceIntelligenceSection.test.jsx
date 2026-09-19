import React from 'react';
import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import InvoiceFinanceIntelligenceSection from '../../pages/dashboard/Finance/components/InvoiceFinanceIntelligenceSection';
const mockGetInvoice360FinanceIntelligence = vi.fn();

vi.mock('../../services/invoiceService', () => ({
  invoiceService: {
    getInvoice360FinanceIntelligence: (...args) => mockGetInvoice360FinanceIntelligence(...args),
  },
  default: {
    getInvoice360FinanceIntelligence: (...args) => mockGetInvoice360FinanceIntelligence(...args),
  },
}));

describe('InvoiceFinanceIntelligenceSection Component', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  const mockCriticalPayload = {
    invoice_id: 1,
    organization_scope: 1,
    correlation_id: 'corr-fin-test-1',
    calculated_at: '2026-09-07T14:30:00Z',
    read_only: true,
    identity: {
      invoice_id: 1,
      invoice_number: 'INV-2026-0456',
      customer_id: 1,
      customer_name: 'Global Traders Inc.',
      customer_country: 'USA',
      shipment_id: 101,
      shipment_number: 'SH-2026-00124',
      booking_id: 201,
      booking_number: 'BK-2026-001',
      quotation_id: 301,
      quote_number: 'QT-2026-001',
      route: 'Shanghai ➔ Los Angeles',
      invoice_status: 'Issued',
      payment_status: 'Issued',
      currency: 'USD',
      issue_date: '2026-08-15T00:00:00Z',
      due_date: '2026-08-30T00:00:00Z',
      subtotal: 22410.0,
      tax_amount: 0.0,
      discount_amount: 0.0,
      total_amount: 24650.0,
      paid_amount: 0.0,
      balance_due: 24650.0,
      overdue_amount: 24650.0,
      days_until_due: 0,
      days_overdue: 8,
      invoice_age_days: 23,
      aging_bucket: '1-30_DAYS',
      is_overdue: true,
      is_paid: false,
      is_partially_paid: false,
    },
    customer_receivables: {
      customer_id: 1,
      customer_name: 'Global Traders Inc.',
      total_invoiced_amount: 146760.0,
      total_paid_amount: 35430.0,
      total_outstanding_amount: 111330.0,
      total_overdue_amount: 111330.0,
      open_invoices_count: 6,
      overdue_invoices_count: 6,
      paid_invoices_count: 2,
      payment_completion_rate: 24.1,
      average_payment_delay_days: 12.5,
      has_repeated_late_payments: true,
      exposure_rating: 'SEVERE',
    },
    revenue_cost: {
      invoice_revenue: 24650.0,
      quotation_amount: 24650.0,
      commercial_cost_amount: 19500.0,
      gross_margin_amount: 5150.0,
      gross_margin_percentage: 20.9,
      revenue_cost_variance: 0.0,
      has_missing_cost: false,
      currency: 'USD',
      is_unlinked_invoice: false,
    },
    line_items: [
      {
        id: 1,
        description: 'Ocean Freight (40ft FCL High Cube)',
        service_category: 'Freight',
        quantity: 1.0,
        unit_price: 22060.0,
        total_amount: 22060.0,
      },
      {
        id: 2,
        description: 'Documentation Fee & B/L Issuance',
        service_category: 'Documentation',
        quantity: 1.0,
        unit_price: 350.0,
        total_amount: 350.0,
      },
    ],
    payments: [
      {
        id: 10,
        payment_ref: 'PAY-PARTIAL-001',
        amount: 2000.0,
        payment_method: 'Wire Transfer',
        status: 'Completed',
        payment_date: '2026-08-20T00:00:00Z',
      },
    ],
    risk_indicators: {
      overall_risk_rating: 'HIGH',
      overall_risk_score: 55,
      is_overdue: true,
      is_approaching_due_date: false,
      has_large_balance: true,
      is_partially_paid: false,
      missing_due_date: false,
      missing_customer: false,
      missing_shipment: false,
      line_items_inconsistent: false,
      revenue_without_cost: false,
      repeated_customer_late_payer: true,
      risk_factors: [
        'Invoice overdue by 8 days (Aging: 1-30_DAYS)',
        'High outstanding exposure: 24650.00 USD',
        'Customer has recorded history of late settlements',
      ],
    },
    ai_summary: {
      executive_summary:
        'Invoice INV-2026-0456 for Global Traders Inc. is Issued with total 24650.00 USD (Paid: 0.00, Balance Due: 24650.00). Outstanding balance is overdue by 8 days (1-30_DAYS bucket).',
      outstanding_exposure_explanation:
        'Current outstanding exposure on this invoice is 24650.00 USD. Customer Global Traders Inc. holds 111330.00 USD in total AR exposure across 6 open invoices.',
      aging_position_explanation:
        "Invoice is in aging bracket '1-30_DAYS' with 8 days overdue and 23 days since original invoice date.",
      revenue_cost_observations:
        'Commercial gross margin is 5150.00 USD (20.9%) based on linked quotation cost.',
      payment_behavior_observations:
        'Historical data indicates average settlement delay of 12.5 days past due date.',
      actionable_attention_items: [
        'Initiate accounts receivable follow-up for 24650.00 USD overdue since 2026-08-30',
      ],
      suggested_finance_inquiries: [
        'Request payment remittance advice from Global Traders Inc. AP department',
      ],
      confidence: 'HIGH',
      freshness_timestamp: '2026-09-07T14:30:00Z',
      citations: [
        '[Invoice: #1 (INV-2026-0456)]',
        '[Customer: #1 (Global Traders Inc.)]',
        '[Shipment: #101 (SH-2026-00124)]',
      ],
    },
    audited_line_item_total: 22410.0,
    warnings: [],
  };

  it('renders loading state initially', () => {
    mockGetInvoice360FinanceIntelligence.mockReturnValue(new Promise(() => {}));

    render(<InvoiceFinanceIntelligenceSection invoiceId={1} />);

    expect(screen.getByTestId('invoice-finance-intelligence-loading')).toBeInTheDocument();
    expect(
      screen.getByText(/Assembling deterministic finance intelligence & receivables aging.../i)
    ).toBeInTheDocument();
  });

  it('renders error state when fetch fails and allows retry', async () => {
    mockGetInvoice360FinanceIntelligence.mockRejectedValueOnce(
      new Error('Finance intelligence service unavailable')
    );

    render(<InvoiceFinanceIntelligenceSection invoiceId={1} />);

    await waitFor(() => {
      expect(screen.getByTestId('invoice-finance-intelligence-error')).toBeInTheDocument();
    });

    expect(screen.getByText('Finance intelligence service unavailable')).toBeInTheDocument();

    // Test retry
    mockGetInvoice360FinanceIntelligence.mockResolvedValueOnce({
      data: mockCriticalPayload,
    });

    const retryBtn = screen.getByTestId('inv-intel-retry-btn');
    fireEvent.click(retryBtn);

    await waitFor(() => {
      expect(screen.getByTestId('invoice-finance-intelligence-section')).toBeInTheDocument();
    });
  });

  it('renders complete operational finance intelligence with aging, KPIs, line items, and AI summary', async () => {
    mockGetInvoice360FinanceIntelligence.mockResolvedValue({
      data: mockCriticalPayload,
    });

    render(<InvoiceFinanceIntelligenceSection invoiceId={1} />);

    await waitFor(() => {
      expect(screen.getByTestId('invoice-finance-intelligence-section')).toBeInTheDocument();
    });

    // Check Header & Badges
    expect(screen.getByText('READ-ONLY FINANCE INTELLIGENCE')).toBeInTheDocument();
    expect(screen.getByTestId('inv-intel-risk-rating')).toHaveTextContent('Risk: HIGH (55/100)');
    expect(screen.getByTestId('inv-intel-aging-pill')).toHaveTextContent('Aging: 1-30_DAYS');

    // Check AI Summary Card
    expect(screen.getByTestId('inv-intel-ai-card')).toBeInTheDocument();
    expect(screen.getByText(/Invoice INV-2026-0456 for Global Traders Inc. is Issued/i)).toBeInTheDocument();
    expect(screen.getByText('Confidence: HIGH')).toBeInTheDocument();
    expect(screen.getByText('[Invoice: #1 (INV-2026-0456)]')).toBeInTheDocument();
    expect(screen.getByText('[Customer: #1 (Global Traders Inc.)]')).toBeInTheDocument();

    // Check KPI Grid
    expect(screen.getByTestId('inv-intel-kpi-balance')).toHaveTextContent('24,650.00 USD');
    expect(screen.getByTestId('inv-intel-kpi-aging')).toHaveTextContent('8d Overdue');
    expect(screen.getByTestId('inv-intel-kpi-customer-ar')).toHaveTextContent('111,330.00 USD');
    expect(screen.getByTestId('inv-intel-kpi-margin')).toHaveTextContent('20.9%');

    // Check Line Items Table
    expect(screen.getByTestId('inv-intel-line-items-card')).toBeInTheDocument();
    expect(screen.getByText('Ocean Freight (40ft FCL High Cube)')).toBeInTheDocument();
    expect(screen.getByText('Documentation Fee & B/L Issuance')).toBeInTheDocument();

    // Check Payments & Risks
    expect(screen.getByTestId('inv-intel-payments-card')).toBeInTheDocument();
    expect(screen.getByText('PAY-PARTIAL-001')).toBeInTheDocument();
    expect(screen.getByText('Invoice overdue by 8 days (Aging: 1-30_DAYS)')).toBeInTheDocument();
  });

  it('handles clean invoice with 0 risks gracefully', async () => {
    const cleanPayload = {
      ...mockCriticalPayload,
      invoice_id: 101,
      identity: {
        ...mockCriticalPayload.identity,
        invoice_id: 101,
        invoice_number: 'INV-2026-DEV-001',
        invoice_status: 'Paid',
        payment_status: 'Paid',
        balance_due: 0.0,
        paid_amount: 3200.0,
        total_amount: 3200.0,
        is_overdue: false,
        is_paid: true,
        days_overdue: 0,
        aging_bucket: 'NOT_DUE',
      },
      risk_indicators: {
        overall_risk_rating: 'LOW',
        overall_risk_score: 0,
        is_overdue: false,
        risk_factors: [],
      },
    };

    mockGetInvoice360FinanceIntelligence.mockResolvedValue({
      data: cleanPayload,
    });

    render(<InvoiceFinanceIntelligenceSection invoiceId={101} />);

    await waitFor(() => {
      expect(screen.getByTestId('invoice-finance-intelligence-section')).toBeInTheDocument();
    });

    expect(screen.getByTestId('inv-intel-risk-rating')).toHaveTextContent('Risk: LOW (0/100)');
    expect(screen.getByTestId('inv-intel-aging-pill')).toHaveTextContent('Aging: NOT_DUE');
    expect(
      screen.getByText(/Zero elevated financial risks detected. Invoice billing & payment posture is healthy./i)
    ).toBeInTheDocument();
  });
});

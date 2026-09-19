import React from 'react';
import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import CustomerIntelligence360Section from '../../pages/dashboard/Customers/CustomerIntelligence360Section';
import * as customerService from '../../services/customerService';

vi.mock('../../services/customerService', () => ({
  getCustomer360Intelligence: vi.fn(),
}));

describe('CustomerIntelligence360Section Component', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  const mockPayload = {
    customer_id: 101,
    identity: {
      id: 101,
      org_id: 2,
      name: 'Apex Global Logistics Corp',
      customer_code: 'APEX-001',
      tier: 'GOLD',
      status: 'ACTIVE',
      lifecycle_stage: 'ACTIVE',
      account_manager_name: 'David Vance',
      authorized_contacts: [
        {
          id: 50,
          first_name: 'Elena',
          last_name: 'Rostova',
          email: 'elena@apexlogistics.com',
          job_title: 'VP Operations',
          is_primary: true,
        },
      ],
    },
    commercial_metrics: {
      total_rfqs: 3,
      total_quotations: 1,
      total_bookings: 1,
      total_contracts: 1,
      quote_to_booking_conversion_rate: 100.0,
      average_quotation_value: 2850.0,
      engagement_trend: 'HIGH',
      recent_activity_count: 7,
      last_interaction_date: '2026-03-01T12:00:00Z',
    },
    operations_metrics: {
      total_shipments: 3,
      active_shipment_count: 2,
      delayed_shipment_count: 1,
      open_exception_count: 1,
      average_shipment_cycle_days: 12.0,
      last_shipment_date: '2026-02-28T09:00:00Z',
      delivery_performance_summary: '2 active shipments, 1 delayed shipment.',
    },
    financial_metrics: {
      total_invoiced_amount: 52000.0,
      total_paid_amount: 32000.0,
      outstanding_balance: 20000.0,
      overdue_invoice_count: 1,
      payment_terms: 'NET30',
      credit_limit: 100000.0,
      credit_status: 'GOOD_STANDING',
    },
    governance_metrics: {
      open_approval_count: 1,
      recent_ai_task_count: 2,
      relevant_audit_count: 8,
      last_audit_timestamp: '2026-03-05T10:30:00Z',
    },
    attention_items: [
      '1 overdue invoice requires follow-up (total balance: $20,000.00).',
      '1 delayed shipments currently flagged in transit.',
    ],
    ai_summary: {
      executive_summary: 'Apex Global Logistics is an active enterprise customer with strong booking conversion.',
      commercial_position: 'High commercial momentum with 100% quotation conversion.',
      operational_position: '2 active shipments, with 1 shipment experiencing customs delay.',
      financial_and_approval_concerns: '1 invoice overdue by 14 days.',
      important_risks: ['Transit delay on ocean lane CNSHA-USLAX.'],
      recommended_follow_up_actions: [
        'Contact finance team regarding overdue invoice INV-2026-09.',
        'Escalate delayed shipment SHP-302 with destination broker.',
      ],
      supporting_observations: [
        {
          source_module: 'INVOICING',
          record_type: 'CUSTOMER_INVOICE',
          record_id: 'INV-2026-09',
          timestamp: '2026-02-15T00:00:00Z',
          explanation: 'Overdue invoice flagged for credit attention.',
        },
        {
          source_module: 'SHIPMENTS',
          record_type: 'SHIPMENT',
          record_id: 'SHP-302',
          timestamp: '2026-02-28T00:00:00Z',
          explanation: 'Delayed vessel departure recorded.',
        },
      ],
      confidence_level: 'HIGH',
      data_freshness: 'Real-time database query as of 2026-09-07T18:00:00Z',
      data_warnings: [],
      classification: 'READ_ONLY_INFORMATIONAL',
    },
    correlation_id: 'cust-intel-360-101-corr-xyz',
  };

  it('renders loading state initially, then populates complete 360 intelligence', async () => {
    customerService.getCustomer360Intelligence.mockResolvedValueOnce(mockPayload);

    render(<CustomerIntelligence360Section customerId={101} />);

    // Loading check
    expect(screen.getByText(/Assembling Customer 360° Intelligence/i)).toBeInTheDocument();

    // Wait for content
    await waitFor(() => {
      expect(screen.getByText('Customer Intelligence Summary')).toBeInTheDocument();
    });

    // Verify header badges
    expect(screen.getByText('READ-ONLY 360° CONTEXT')).toBeInTheDocument();
    expect(screen.getByText(/HIGH ENGAGEMENT/i)).toBeInTheDocument();
    expect(screen.getByText(/HIGH CONFIDENCE/i)).toBeInTheDocument();

    // Verify AI Summary
    expect(screen.getByText('Apex Global Logistics is an active enterprise customer with strong booking conversion.')).toBeInTheDocument();
    expect(screen.getByText('High commercial momentum with 100% quotation conversion.')).toBeInTheDocument();

    // Verify Attention Items
    expect(screen.getByText(/Key Attention & Risk Items \(2\)/i)).toBeInTheDocument();
    expect(screen.getByText('1 overdue invoice requires follow-up (total balance: $20,000.00).')).toBeInTheDocument();

    // Verify 4-Quadrant Metrics
    expect(screen.getByText('Commercial Summary')).toBeInTheDocument();
    expect(screen.getByText('100.0%')).toBeInTheDocument();
    expect(screen.getByText('Operations Summary')).toBeInTheDocument();
    expect(screen.getByText('Financial Summary')).toBeInTheDocument();
    expect(screen.getByText('$52,000')).toBeInTheDocument();
    expect(screen.getByText('$20,000')).toBeInTheDocument();
    expect(screen.getByText('Governance & Controls')).toBeInTheDocument();

    // Verify Traceability Table
    expect(screen.getByText('Supporting Record Traceability')).toBeInTheDocument();
    expect(screen.getByText('INV-2026-09')).toBeInTheDocument();
    expect(screen.getByText('SHP-302')).toBeInTheDocument();
    expect(screen.getByText('Overdue invoice flagged for credit attention.')).toBeInTheDocument();

    // Verify Authorized Contacts
    expect(screen.getByText(/Authorized Key Contacts \(1\)/i)).toBeInTheDocument();
    expect(screen.getByText('Elena Rostova')).toBeInTheDocument();
    expect(screen.getByText('elena@apexlogistics.com')).toBeInTheDocument();
  });

  it('handles error state and allows successful retry', async () => {
    customerService.getCustomer360Intelligence.mockRejectedValueOnce(new Error('Cross-tenant forbidden'));

    render(<CustomerIntelligence360Section customerId={999} />);

    await waitFor(() => {
      expect(screen.getByText(/Unable to load Customer 360° Intelligence/i)).toBeInTheDocument();
    });
    expect(screen.getByText('Cross-tenant forbidden')).toBeInTheDocument();

    // Retry
    customerService.getCustomer360Intelligence.mockResolvedValueOnce(mockPayload);
    const retryBtn = screen.getByRole('button', { name: /retry/i });
    fireEvent.click(retryBtn);

    await waitFor(() => {
      expect(screen.getByText('Customer Intelligence Summary')).toBeInTheDocument();
    });
  });

  it('allows manual refresh clicking the refresh button', async () => {
    customerService.getCustomer360Intelligence.mockResolvedValueOnce(mockPayload);

    render(<CustomerIntelligence360Section customerId={101} />);

    await waitFor(() => {
      expect(screen.getByText('Customer Intelligence Summary')).toBeInTheDocument();
    });

    // Click refresh
    customerService.getCustomer360Intelligence.mockResolvedValueOnce({
      ...mockPayload,
      correlation_id: 'corr-new-refresh-999',
    });

    const refreshBtn = screen.getByRole('button', { name: /Refresh 360° Intelligence/i });
    fireEvent.click(refreshBtn);

    await waitFor(() => {
      expect(customerService.getCustomer360Intelligence).toHaveBeenCalledTimes(2);
    });
  });
});

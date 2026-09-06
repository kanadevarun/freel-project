import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import { MemoryRouter } from 'react-router-dom';
import RFQActivityTimeline from '../../../pages/dashboard/RFQ/components/RFQActivityTimeline';

describe('RFQActivityTimeline Component', () => {
  const mockRfq = {
    id: 101,
    rfq_number: 'RFQ-20260827-001',
    lead_id: 1268,
    customer_name: 'Global Exports Ltd',
  };

  const mockActivityData = {
    summary: {
      total_events: 5,
      customer_events: 2,
      operational_events: 2,
      ai_events: 1,
      requirements_events: 1,
      document_events: 0,
      quote_events: 1,
      action_required_count: 0,
      latest_activity_at: '2026-08-27T10:45:00Z',
    },
    events: [
      {
        id: 'act-1',
        type: 'CUSTOMER_INQUIRY',
        category: 'CUSTOMER',
        title: 'Customer Email Received',
        description: 'Inquiry for 2x40ft HC containers from Nhava Sheva to Hamburg.',
        timestamp: '2026-08-27T08:30:00Z',
        actor_type: 'CUSTOMER',
        actor_name: 'buyer@globalexports.com',
        source_type: 'LEAD',
        source_id: '1268',
        is_important: true,
        requires_action: false,
      },
      {
        id: 'act-2',
        type: 'AI_EXTRACTION',
        category: 'AI',
        title: 'AI Extracted Shipment Details',
        description: 'AI extracted POL (INNSA), POD (DEHAM), and Incoterms (FOB).',
        timestamp: '2026-08-27T08:32:00Z',
        actor_type: 'AI',
        actor_name: 'AI Assistant',
        source_type: 'LEAD',
        source_id: '1268',
        is_important: true,
        requires_action: false,
      },
      {
        id: 'act-3',
        type: 'LEAD_CONVERTED',
        category: 'OPERATIONS',
        title: 'Lead Converted to RFQ',
        description: 'Lead #1268 was successfully converted into RFQ-20260827-001.',
        timestamp: '2026-08-27T08:45:00Z',
        actor_type: 'OPERATIONS',
        actor_name: 'Operations Team',
        source_type: 'LEAD',
        source_id: '1268',
        is_important: true,
        requires_action: false,
      },
      {
        id: 'act-4',
        type: 'REQUIREMENTS_EVALUATED',
        category: 'REQUIREMENTS',
        title: 'Requirements Evaluated',
        description: 'Operational readiness: READY_FOR_QUOTATION (93% score, 0 blockers).',
        timestamp: '2026-08-27T09:00:00Z',
        actor_type: 'SYSTEM',
        actor_name: 'Requirements Engine',
        source_type: 'RFQ',
        source_id: '101',
        is_important: true,
        requires_action: false,
      },
      {
        id: 'act-5',
        type: 'QUOTE_GENERATED',
        category: 'QUOTES',
        title: 'Quote Generated — Maersk Line',
        description: 'Carrier quote from Maersk Line: Buy $4,200.00, Sell $4,850.00.',
        timestamp: '2026-08-27T10:15:00Z',
        actor_type: 'OPERATIONS',
        actor_name: 'Pricing Engine',
        source_type: 'QUOTE',
        source_id: '201',
        is_important: true,
        requires_action: false,
      },
    ],
    lead_id: 1268,
  };

  it('renders timeline header, summary metrics, and source lead banner', () => {
    render(
      <MemoryRouter>
        <RFQActivityTimeline
          rfq={mockRfq}
          activityData={mockActivityData}
        />
      </MemoryRouter>
    );

    expect(screen.getByText('Activity & Audit Trail')).toBeInTheDocument();
    expect(screen.getByText('Originated from Lead #1268')).toBeInTheDocument();
    expect(screen.getByText('Open Email Thread')).toBeInTheDocument();
    expect(screen.getByText('5 Events')).toBeInTheDocument();
    expect(screen.getByText('Total Events')).toBeInTheDocument();
    expect(screen.getByText('Customer Inquiries')).toBeInTheDocument();
  });

  it('renders event cards with titles, actors, and descriptions', () => {
    render(
      <MemoryRouter>
        <RFQActivityTimeline
          rfq={mockRfq}
          activityData={mockActivityData}
        />
      </MemoryRouter>
    );

    expect(screen.getByText('Customer Email Received')).toBeInTheDocument();
    expect(screen.getByText('buyer@globalexports.com')).toBeInTheDocument();
    expect(screen.getByText('AI Extracted Shipment Details')).toBeInTheDocument();
    expect(screen.getByText('AI Assistant')).toBeInTheDocument();
    expect(screen.getByText('Lead Converted to RFQ')).toBeInTheDocument();
    expect(screen.getByText('Quote Generated — Maersk Line')).toBeInTheDocument();
  });

  it('filters events when category filter pill is clicked', () => {
    render(
      <MemoryRouter>
        <RFQActivityTimeline
          rfq={mockRfq}
          activityData={mockActivityData}
        />
      </MemoryRouter>
    );

    // Click on Customer filter
    const customerFilterBtn = screen.getByRole('button', { name: /Customer/i });
    fireEvent.click(customerFilterBtn);

    expect(screen.getByText('Customer Email Received')).toBeInTheDocument();
    expect(screen.queryByText('Lead Converted to RFQ')).not.toBeInTheDocument();
    expect(screen.queryByText('Quote Generated — Maersk Line')).not.toBeInTheDocument();
  });

  it('displays empty state when filtered category has no events', () => {
    render(
      <MemoryRouter>
        <RFQActivityTimeline
          rfq={mockRfq}
          activityData={mockActivityData}
        />
      </MemoryRouter>
    );

    // Click on Documents filter (0 events)
    const docsFilterBtn = screen.getByRole('button', { name: /Documents/i });
    fireEvent.click(docsFilterBtn);

    expect(screen.getByText('No activity in this category')).toBeInTheDocument();
    expect(screen.getByText('View All Activity')).toBeInTheDocument();
  });


  it('calls onRefresh when refresh button is clicked', () => {
    const handleRefresh = vi.fn();
    render(
      <MemoryRouter>
        <RFQActivityTimeline
          rfq={mockRfq}
          activityData={mockActivityData}
          onRefresh={handleRefresh}
        />
      </MemoryRouter>
    );

    const refreshBtn = screen.getByText('Refresh');
    fireEvent.click(refreshBtn);
    expect(handleRefresh).toHaveBeenCalledTimes(1);
  });
});

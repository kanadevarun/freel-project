import React from 'react';
import { render, screen } from '@testing-library/react';
import { describe, it, expect } from 'vitest';
import { MemoryRouter } from 'react-router-dom';
import OperationalDashboard from '../../../../pages/dashboard/Home/OperationalDashboard';

describe('OperationalDashboard Component', () => {
  const mockData = {
    stats: {
      open_leads: 5,
      leads_trend_pct: 12.5,
      leads_trend_direction: 'up',
      open_rfqs: 3,
      rfqs_trend_pct: 5.0,
      rfqs_trend_direction: 'up',
      active_quotations: 2,
      quotes_trend_pct: 0,
      quotes_trend_direction: 'neutral',
      active_shipments: 4,
      shipments_trend_pct: 10,
      shipments_trend_direction: 'up',
      pending_approvals: 1,
      approvals_trend_pct: 0,
      approvals_trend_direction: 'no_data',
      outstanding_invoices: 2,
    },
    pipeline: {
      leads_count: 5,
      rfqs_count: 3,
      quotations_count: 2,
      bookings_count: 2,
    },
    attention_items: [],
    active_shipments: [],
    pending_approvals: [],
    recent_activity: [],
    recent_documents: [],
    upcoming_reminders: [],
  };

  const mockUser = {
    first_name: 'Varun',
    last_name: 'Kanade',
    org_name: 'LogisticsHQ',
  };

  it('renders without compLabel ReferenceError and displays KPI cards', () => {
    render(
      <MemoryRouter initialEntries={['/dashboard?preset=LAST_7D']}>
        <OperationalDashboard data={mockData} user={mockUser} />
      </MemoryRouter>
    );

    expect(screen.getByText('Active Leads')).toBeInTheDocument();
    expect(screen.getByText('Open RFQs')).toBeInTheDocument();
    expect(screen.getAllByText('Quotations').length).toBeGreaterThan(0);
    expect(screen.getAllByText('Active Shipments').length).toBeGreaterThan(0);
    expect(screen.getAllByText('Pending Approvals').length).toBeGreaterThan(0);
    expect(screen.getByText('Outstanding Invoices')).toBeInTheDocument();
    expect(screen.getAllByText(/vs preceding 7 days/).length).toBeGreaterThan(0);
  });
});

import React from 'react';
import { render, screen, act, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import { MemoryRouter } from 'react-router-dom';
import OperationalDashboard from '../../../../pages/dashboard/Home/OperationalDashboard';
import { RBACContext } from '../../../../context/RBACContext';
import { monitoringService } from '../../../../services/monitoringService';
import { aiTaskService } from '../../../../services/aiTaskService';

// Mock monitoringService, aiTaskService, recommendationService
vi.mock('../../../../services/monitoringService', () => ({
  monitoringService: {
    getHealthSummary: vi.fn().mockResolvedValue({
      data: {
        overall_status: 'HEALTHY',
        subsystems: {},
      },
    }),
  },
}));

vi.mock('../../../../services/aiTaskService', () => ({
  aiTaskService: {
    getWorkforceSummary: vi.fn().mockResolvedValue({
      data: {
        active_agents_count: 8,
        total_active_tasks: 0,
        waiting_for_approval_tasks: 4,
        health: { overall_status: 'healthy' },
      },
    }),
  },
}));

vi.mock('../../../../services/recommendationService', () => ({
  recommendationService: {
    getStats: vi.fn().mockResolvedValue({
      stats: {
        total_active: 31,
        critical_count: 3,
        high_count: 13,
        requires_approval: 4,
      },
    }),
  },
}));

vi.mock('../../../../services/integrationService', () => ({
  default: {
    getStatuses: vi.fn().mockResolvedValue([]),
  },
  integrationService: {
    getStatuses: vi.fn().mockResolvedValue([]),
  },
}));

vi.mock('../../../../services/predictionService', () => ({
  default: {
    getWorkloadCapacitySummary: vi.fn().mockResolvedValue({
      data: { overall_status: 'HEALTHY', active_tab: 'summary' },
    }),
    getResourceBottleneckSummary: vi.fn().mockResolvedValue({
      data: { overall_status: 'HEALTHY' },
    }),
    getWorkloadPrediction: vi.fn().mockResolvedValue({
      data: { overall_status: 'HEALTHY' },
    }),
    getCapacityPrediction: vi.fn().mockResolvedValue({
      data: { overall_status: 'HEALTHY' },
    }),
    getDemandPrediction: vi.fn().mockResolvedValue({
      data: { overall_status: 'HEALTHY' },
    }),
  },
  predictionService: {
    getWorkloadCapacitySummary: vi.fn().mockResolvedValue({
      data: { overall_status: 'HEALTHY', active_tab: 'summary' },
    }),
    getResourceBottleneckSummary: vi.fn().mockResolvedValue({
      data: { overall_status: 'HEALTHY' },
    }),
    getWorkloadPrediction: vi.fn().mockResolvedValue({
      data: { overall_status: 'HEALTHY' },
    }),
    getCapacityPrediction: vi.fn().mockResolvedValue({
      data: { overall_status: 'HEALTHY' },
    }),
    getDemandPrediction: vi.fn().mockResolvedValue({
      data: { overall_status: 'HEALTHY' },
    }),
  },
}));

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
      outstanding_amount: 4500,
    },
    pipeline: {
      leads_count: 5,
      rfqs_count: 3,
      quotations_count: 2,
      bookings_count: 2,
      shipments_count: 4,
    },
    attention_items: [
      {
        id: 'rfqs_awaiting_quote',
        priority: 'HIGH',
        category: 'RFQs',
        title: '2 RFQ(s) are awaiting your quotation',
        subtitle: 'Latest: RFQ-2026-DEV-002 from Nordic Freight Dynamics AB',
        action_url: '/dashboard/rfqs',
      },
    ],
    active_shipments: [
      {
        id: 1,
        shipment_no: 'SH-2026-001',
        carrier: 'Maersk',
        origin: 'INNSA',
        destination: 'USNYC',
        eta: '2026-09-15',
        status: 'IN_TRANSIT',
        status_display: 'In Transit',
      },
    ],
    pending_approvals: [
      {
        id: 1,
        title: 'Quotation Discount Approval',
        requested_by: 'Alex Vance',
        role_badge: 'Manager',
        relative_age: '2h ago',
      },
    ],
    recent_activity: [
      {
        id: 1,
        type: 'SHIPMENT',
        title: 'Customs Cleared',
        subtitle: 'Container cleared at Mumbai Port',
        timestamp: '10m ago',
        action_url: '/dashboard/shipments',
      },
    ],
    recent_documents: [],
    upcoming_reminders: [
      {
        id: 1,
        title: 'Follow-up on quotation QT-8891',
        subtitle: 'Nordic Freight',
        due_text: 'Today',
        action_url: '/dashboard/quotations',
      },
    ],
    invoice_summary: {
      recent_invoices: [
        {
          id: 1,
          invoice_number: 'INV-2026-001',
          customer_name: 'Acme Corp',
          total_amount: 1500,
          status: 'Issued',
          relative_age: '3d ago',
        },
      ],
    },
  };

  const mockUser = {
    first_name: 'Varun',
    last_name: 'Kanade',
    org_name: 'LogisticsHQ',
  };

  it('renders all 7 information architecture sections systematically', async () => {
    await act(async () => {
      render(
        <MemoryRouter initialEntries={['/dashboard?preset=LAST_7D']}>
          <OperationalDashboard data={mockData} user={mockUser} />
        </MemoryRouter>
      );
    });

    // 1. Dashboard Header
    expect(screen.getByRole('heading', { name: /Operations Command/i })).toBeInTheDocument();
    expect(screen.getByText(/● Operations Live/i)).toBeInTheDocument();

    // 2. Primary 5 KPI Cards
    expect(screen.getByText('Active Leads')).toBeInTheDocument();
    expect(screen.getByText('Open RFQs')).toBeInTheDocument();
    expect(screen.getAllByText('Active Shipments').length).toBeGreaterThanOrEqual(1);
    expect(screen.getAllByText('Pending Approvals').length).toBeGreaterThanOrEqual(1);
    expect(screen.getByText('Outstanding Invoices')).toBeInTheDocument();

    // 3. Priority Actions
    expect(screen.getByRole('region', { name: /Priority Actions/i })).toBeInTheDocument();
    expect(screen.getByText('2 RFQ(s) are awaiting your quotation')).toBeInTheDocument();

    // 4. Operations Overview
    expect(screen.getByRole('region', { name: /Operations Overview/i })).toBeInTheDocument();
    expect(screen.getByText('Business Pipeline')).toBeInTheDocument();
    expect(screen.getByText('Recent Activity')).toBeInTheDocument();

    // 5. Finance and Approvals
    expect(screen.getByRole('region', { name: /Finance and Approvals/i })).toBeInTheDocument();
    expect(screen.getByText('Invoice Overview')).toBeInTheDocument();

    // 6 & 7. Compact AI Summary + System Health
    expect(screen.getByRole('region', { name: /AI Summary and System Health/i })).toBeInTheDocument();
    expect(screen.getByText('AI Workforce')).toBeInTheDocument();
    expect(screen.getByText(/System Health/)).toBeInTheDocument();
    expect(screen.getByText('Go API Backend')).toBeInTheDocument();
    expect(screen.getByText('Python AI Sidecar')).toBeInTheDocument();
    expect(screen.getByText('MariaDB Storage')).toBeInTheDocument();
    expect(screen.getByText('Queue Workers')).toBeInTheDocument();
  });

  it('renders all 5 KPI cards with full non-truncated titles and prominent values', async () => {
    await act(async () => {
      render(
        <MemoryRouter initialEntries={['/dashboard?preset=LAST_7D']}>
          <OperationalDashboard data={mockData} user={mockUser} />
        </MemoryRouter>
      );
    });

    // Verify all 5 titles are present and accessible
    const leadsCard = screen.getByRole('button', { name: /Active Leads: 5/i });
    const rfqsCard = screen.getByRole('button', { name: /Open RFQs: 3/i });
    const shipmentsCard = screen.getByRole('button', { name: /Active Shipments: 4/i });
    const approvalsCard = screen.getByRole('button', { name: /Pending Approvals: 1/i });
    const invoicesCard = screen.getByRole('button', { name: /Outstanding Invoices: 2/i });

    expect(leadsCard).toBeInTheDocument();
    expect(rfqsCard).toBeInTheDocument();
    expect(shipmentsCard).toBeInTheDocument();
    expect(approvalsCard).toBeInTheDocument();
    expect(invoicesCard).toBeInTheDocument();

    // Verify invoice amount pill is properly formatted
    expect(screen.getByText('$4,500.00')).toBeInTheDocument();
  });

  it('renders grounded operational status when comparison baseline is unavailable', async () => {
    const noTrendData = {
      ...mockData,
      stats: {
        ...mockData.stats,
        leads_trend_pct: 0,
        leads_trend_direction: 'no_data',
        rfqs_trend_pct: 0,
        rfqs_trend_direction: 'no_data',
        shipments_trend_pct: 0,
        shipments_trend_direction: 'no_data',
        approvals_trend_pct: 0,
        approvals_trend_direction: 'no_data',
        invoices_trend_pct: 0,
        invoices_trend_direction: 'no_data',
      },
    };

    await act(async () => {
      render(
        <MemoryRouter initialEntries={['/dashboard?preset=LAST_7D']}>
          <OperationalDashboard data={noTrendData} user={mockUser} />
        </MemoryRouter>
      );
    });

    // Verify grounded operational status tags are rendered instead of ungrounded 100% trend pills
    expect(screen.getByText(/● 5 in pipeline/i)).toBeInTheDocument();
    expect(screen.getByText(/● 3 awaiting quotes/i)).toBeInTheDocument();
    expect(screen.getByText(/● 4 in transit/i)).toBeInTheDocument();
    expect(screen.getByText(/● 1 need review/i)).toBeInTheDocument();
    expect(screen.getByText(/● Awaiting payment/i)).toBeInTheDocument();

    // Ensure no misleading "100% vs preceding" text exists
    expect(screen.queryByText(/100% vs preceding/i)).not.toBeInTheDocument();
  });

  it('renders Priority Actions with urgency chips, metadata badges, source references, and approval gating', async () => {
    const priorityData = {
      ...mockData,
      attention_items: [
        {
          id: 'crit_shipment_exception_103',
          urgency: 'CRITICAL',
          priority: 'CRITICAL',
          category: 'Shipments',
          module: 'Shipments',
          title: 'Shipment exception needs review',
          explanation: 'Container CMDU543216789 on INNSA → USNYC is on customs hold.',
          source_reference: 'CMDU543216789',
          action_url: '/dashboard/shipments',
          action_label: 'Review Exception →',
          age_text: '1h ago',
          requires_approval: false,
          capability: 'MUTATION_CAPABLE',
        },
        {
          id: 'crit_pending_approvals_203',
          urgency: 'CRITICAL',
          priority: 'CRITICAL',
          category: 'Approvals',
          module: 'Approvals',
          title: 'Quotation approval is blocking release',
          explanation: 'Operator approval OPE-APP-2619 is pending review before shipment dispatch.',
          source_reference: 'OPE-APP-2619',
          action_url: '/dashboard/approvals',
          action_label: 'Review Approvals →',
          age_text: '58m ago',
          requires_approval: true,
          capability: 'APPROVAL_GATED',
        },
        {
          id: 'imp_rfqs_awaiting_quote_102',
          urgency: 'IMPORTANT',
          priority: 'HIGH',
          category: 'RFQs',
          module: 'RFQs',
          title: 'RFQ is ready for quotation',
          explanation: '2 open RFQ(s) are awaiting quote preparation. Latest: RFQ-2026-DEV-002 from Nordic Freight Dynamics AB.',
          source_reference: 'RFQ-2026-DEV-002',
          action_url: '/dashboard/rfqs',
          action_label: 'Prepare Quotation →',
          age_text: '3h ago',
          requires_approval: false,
          capability: 'DRAFT_ONLY',
        },
        {
          id: 'info_new_leads_1061',
          urgency: 'INFORMATIONAL',
          priority: 'MEDIUM',
          category: 'Leads',
          module: 'Leads',
          title: 'Qualified new leads in pipeline',
          explanation: '5 new lead(s) have been captured from inbound discovery and are awaiting sales qualification.',
          source_reference: '#1061',
          action_url: '/dashboard/leads',
          action_label: 'View New Leads →',
          age_text: 'Today',
          requires_approval: false,
          capability: 'READ_ONLY',
        },
      ],
    };

    await act(async () => {
      render(
        <MemoryRouter initialEntries={['/dashboard?preset=LAST_7D']}>
          <OperationalDashboard data={priorityData} user={mockUser} />
        </MemoryRouter>
      );
    });

    // Check header counter & chips
    expect(screen.getByText('4 require attention')).toBeInTheDocument();
    expect(screen.getByRole('tab', { name: /All 4/i })).toBeInTheDocument();
    expect(screen.getByRole('tab', { name: /Critical 2/i })).toBeInTheDocument();
    expect(screen.getByRole('tab', { name: /Important 1/i })).toBeInTheDocument();
    expect(screen.getByRole('tab', { name: /Informational 1/i })).toBeInTheDocument();

    // Check items rendered
    expect(screen.getByText('Shipment exception needs review')).toBeInTheDocument();
    expect(screen.getByText('Container CMDU543216789 on INNSA → USNYC is on customs hold.')).toBeInTheDocument();
    expect(screen.getByText('CMDU543216789')).toBeInTheDocument();
    expect(screen.getByText('Review Exception →')).toBeInTheDocument();

    // Check approval gating
    expect(screen.getByText('Quotation approval is blocking release')).toBeInTheDocument();
    expect(screen.getByText('OPE-APP-2619')).toBeInTheDocument();
    expect(screen.getByText('Approval Required')).toBeInTheDocument();
    expect(screen.getByText('Approval Gated')).toBeInTheDocument();
    expect(screen.getByText('Review Approvals →')).toBeInTheDocument();

    // Check important item
    expect(screen.getByText('RFQ is ready for quotation')).toBeInTheDocument();
    expect(screen.getByText('RFQ-2026-DEV-002')).toBeInTheDocument();
    expect(screen.getByText('Draft Only')).toBeInTheDocument();

    // Check informational item
    expect(screen.getByText('Qualified new leads in pipeline')).toBeInTheDocument();
    expect(screen.getByText('#1061')).toBeInTheDocument();
    expect(screen.getByText('Read-Only')).toBeInTheDocument();
  });

  it('filters priority items dynamically when urgency tabs are clicked', async () => {
    const priorityData = {
      ...mockData,
      attention_items: [
        {
          id: 'item_crit',
          urgency: 'CRITICAL',
          priority: 'CRITICAL',
          category: 'Shipments',
          title: 'Critical Shipment Delay',
          explanation: 'Severe port congestion delay.',
          action_url: '/dashboard/shipments',
          action_label: 'Review Delay',
        },
        {
          id: 'item_imp',
          urgency: 'IMPORTANT',
          priority: 'HIGH',
          category: 'RFQs',
          title: 'Important RFQ Pending',
          explanation: 'Customer requested quotation urgently.',
          action_url: '/dashboard/rfqs',
          action_label: 'Create Quote',
        },
      ],
    };

    await act(async () => {
      render(
        <MemoryRouter initialEntries={['/dashboard?preset=LAST_7D']}>
          <OperationalDashboard data={priorityData} user={mockUser} />
        </MemoryRouter>
      );
    });

    // Initially both items are visible
    expect(screen.getByText('Critical Shipment Delay')).toBeInTheDocument();
    expect(screen.getByText('Important RFQ Pending')).toBeInTheDocument();

    // Click Critical tab
    const criticalTab = screen.getByRole('tab', { name: /Critical 1/i });
    fireEvent.click(criticalTab);

    expect(screen.getByText('Critical Shipment Delay')).toBeInTheDocument();
    expect(screen.queryByText('Important RFQ Pending')).not.toBeInTheDocument();

    // Click Important tab
    const importantTab = screen.getByRole('tab', { name: /Important 1/i });
    fireEvent.click(importantTab);

    expect(screen.queryByText('Critical Shipment Delay')).not.toBeInTheDocument();
    expect(screen.getByText('Important RFQ Pending')).toBeInTheDocument();

    // Click Informational tab (has 0 items -> empty state)
    const informationalTab = screen.getByRole('tab', { name: /Informational 0/i });
    fireEvent.click(informationalTab);

    expect(screen.getByText(/No priority actions require your attention right now/i)).toBeInTheDocument();

    // Click All tab back
    const allTab = screen.getByRole('tab', { name: /All 2/i });
    fireEvent.click(allTab);

    expect(screen.getByText('Critical Shipment Delay')).toBeInTheDocument();
    expect(screen.getByText('Important RFQ Pending')).toBeInTheDocument();
  });

  it('renders calm empty state when no priority actions exist', async () => {
    const emptyData = {
      ...mockData,
      attention_items: [],
    };

    await act(async () => {
      render(
        <MemoryRouter initialEntries={['/dashboard?preset=LAST_7D']}>
          <OperationalDashboard data={emptyData} user={mockUser} />
        </MemoryRouter>
      );
    });

    expect(screen.getByText(/No priority actions require your attention right now/i)).toBeInTheDocument();
    expect(screen.getByText(/All operational tasks, shipment exceptions, approvals, and invoices are up to date/i)).toBeInTheDocument();
  });

  it('renders restricted access banner when user lacks required role', async () => {
    const restrictedData = {
      ...mockData,
      attention_items: [
        {
          id: 'crit_finance_sensitive',
          urgency: 'CRITICAL',
          priority: 'CRITICAL',
          category: 'Finance',
          title: 'Direct wire transfer release pending',
          explanation: 'Requires CFO or Finance Director authorization.',
          action_url: '/dashboard/finance',
          action_label: 'Authorize Release',
          required_role: 'FINANCE_DIRECTOR',
        },
      ],
    };

    const standardUser = {
      first_name: 'Varun',
      last_name: 'Kanade',
      role: 'OPERATOR',
    };

    await act(async () => {
      render(
        <MemoryRouter initialEntries={['/dashboard?preset=LAST_7D']}>
          <OperationalDashboard data={restrictedData} user={standardUser} />
        </MemoryRouter>
      );
    });

    expect(screen.getByText('Direct wire transfer release pending')).toBeInTheDocument();
    expect(screen.getByText(/Restricted Access: Requires FINANCE_DIRECTOR permission to review/i)).toBeInTheDocument();
    expect(screen.queryByRole('button', { name: /Authorize Release/i })).not.toBeInTheDocument();
  });

  it('renders Operations Overview with health status breakdown and active shipments', async () => {
    const opsData = {
      ...mockData,
      shipment_counts: {
        in_transit: 3,
        customs_hold: 1,
        delayed: 1,
        delivered: 12,
      },
    };

    await act(async () => {
      render(
        <MemoryRouter initialEntries={['/dashboard']}>
          <OperationalDashboard data={opsData} user={mockUser} />
        </MemoryRouter>
      );
    });

    const opsSection = screen.getByRole('region', { name: /Operations Overview/i });
    expect(opsSection).toBeInTheDocument();

    // Verify shipment health status pills
    expect(screen.getByTitle(/Shipments actively moving along carrier route/i)).toHaveTextContent('In Transit: 3');
    expect(screen.getByTitle(/Shipments with customs hold, documentation block, or delay/i)).toHaveTextContent('Exceptions: 2');
    expect(screen.getByTitle(/Shipments cleared and delivered at destination port/i)).toHaveTextContent('Delivered: 12');

    // Verify shipment row content
    expect(screen.getByText('SH-2026-001')).toBeInTheDocument();
    expect(screen.getByText('Maersk')).toBeInTheDocument();
    expect(screen.getByText('INNSA → USNYC')).toBeInTheDocument();
    expect(screen.getByText('ETA: 2026-09-15')).toBeInTheDocument();

    // Verify Business Pipeline stages
    expect(screen.getByText('Business Pipeline')).toBeInTheDocument();
    expect(screen.getByText('Leads')).toBeInTheDocument();
    expect(screen.getAllByText('RFQs').length).toBeGreaterThanOrEqual(1);
    expect(screen.getByText('Quotations')).toBeInTheDocument();
    expect(screen.getByText('Bookings')).toBeInTheDocument();
    expect(screen.getAllByText('Shipments').length).toBeGreaterThanOrEqual(1);
  });

  it('renders Finance and Approvals with authoritative invoice metrics and approval queue', async () => {
    const finData = {
      ...mockData,
      stats: {
        ...mockData.stats,
        outstanding_amount: 14500,
        outstanding_invoices_count: 3,
        overdue_amount: 4500,
        overdue_invoices: 1,
        paid_this_month: 28000,
        pending_approvals: 2,
      },
      pending_approvals: [
        {
          id: 201,
          request_code: 'OPE-APP-2619',
          title: 'Customs Hold Fee Waiver',
          category: 'Operations',
          requested_by: 'Alex Vance',
          role_badge: 'Manager',
          relative_age: '45m ago',
        },
      ],
    };

    await act(async () => {
      render(
        <MemoryRouter initialEntries={['/dashboard']}>
          <OperationalDashboard data={finData} user={mockUser} />
        </MemoryRouter>
      );
    });

    const finSection = screen.getByRole('region', { name: /Finance and Approvals/i });
    expect(finSection).toBeInTheDocument();

    // Verify Invoice Overview metrics strip
    expect(screen.getByText('Outstanding')).toBeInTheDocument();
    expect(screen.getByText('$14.5K')).toBeInTheDocument();
    expect(screen.getByText('Overdue')).toBeInTheDocument();
    expect(screen.getByText('$4.5K')).toBeInTheDocument();
    expect(screen.getByText('Paid (30D)')).toBeInTheDocument();
    expect(screen.getByText('$28.0K')).toBeInTheDocument();

    // Verify recent invoice row
    expect(screen.getByText('INV-2026-001')).toBeInTheDocument();
    expect(screen.getByText('Acme Corp')).toBeInTheDocument();

    // Verify Pending Approvals card
    expect(screen.getByText('Customs Hold Fee Waiver')).toBeInTheDocument();
    expect(screen.getAllByText('Operations').length).toBeGreaterThanOrEqual(1);
    expect(screen.getByText(/Requested by Alex Vance/i)).toBeInTheDocument();
    expect(screen.getByText('Manager')).toBeInTheDocument();
    expect(screen.getByText('45m ago')).toBeInTheDocument();
  });

  it('renders Recent Business Activity, category filters, and quick create launchers', async () => {
    const activityData = {
      ...mockData,
      recent_activity: [
        {
          id: 101,
          type: 'PAYMENT',
          title: 'Payment Received $4,500.00',
          subtitle: 'Invoice INV-2026-DEV-001 settled by Acme Corp',
          timestamp: '15m ago',
          action_url: '/dashboard/invoices',
        },
        {
          id: 102,
          type: 'SHIPMENT',
          title: 'Container Loaded on Board',
          subtitle: 'Vessel MSC Oscar departing Nhava Sheva',
          timestamp: '1h ago',
          action_url: '/dashboard/shipments',
        },
      ],
    };

    await act(async () => {
      render(
        <MemoryRouter initialEntries={['/dashboard']}>
          <OperationalDashboard data={activityData} user={mockUser} />
        </MemoryRouter>
      );
    });

    // Verify activity filter chips
    expect(screen.getByRole('button', { name: 'All' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Sales' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Operations' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Finance' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Documents' })).toBeInTheDocument();

    // Verify activity items rendered
    expect(screen.getByText('Payment Received $4,500.00')).toBeInTheDocument();
    expect(screen.getByText('Container Loaded on Board')).toBeInTheDocument();

    // Verify upcoming reminders & quick create shortcuts
    expect(screen.getByText('Follow-up on quotation QT-8891')).toBeInTheDocument();
    expect(screen.getAllByText('+ Lead').length).toBeGreaterThanOrEqual(1);
    expect(screen.getAllByText('+ RFQ').length).toBeGreaterThanOrEqual(1);
    expect(screen.getAllByText('+ Quote').length).toBeGreaterThanOrEqual(1);
    expect(screen.getByText('+ Shipment')).toBeInTheDocument();
    expect(screen.getByText('+ Booking')).toBeInTheDocument();
    expect(screen.getByText('+ Invoice')).toBeInTheDocument();
  });

  it('renders calm empty states when operations and financial records are empty', async () => {
    const emptyOperationalData = {
      ...mockData,
      active_shipments: [],
      pending_approvals: [],
      upcoming_reminders: [],
      invoice_summary: {
        recent_invoices: [],
      },
    };

    await act(async () => {
      render(
        <MemoryRouter initialEntries={['/dashboard']}>
          <OperationalDashboard data={emptyOperationalData} user={mockUser} />
        </MemoryRouter>
      );
    });

    expect(screen.getByText(/No active shipments in transit yet/i)).toBeInTheDocument();
    expect(screen.getByText(/No pending approval requests. Human-in-the-loop gate is clear/i)).toBeInTheDocument();
    expect(screen.getByText(/No invoices recorded yet/i)).toBeInTheDocument();
    expect(screen.getByText(/No upcoming reminders scheduled for today/i)).toBeInTheDocument();
  });

  describe('Dashboard Task 7: AI Workforce & System Health Compact Cards', () => {
    it('renders AI Workforce summary with 4 business metrics and operational status badge', async () => {
      aiTaskService.getWorkforceSummary.mockResolvedValueOnce({
        data: {
          total_active_tasks: 2,
          waiting_for_approval_tasks: 1,
          failed_tasks: 0,
          stale_tasks: 0,
          completed_recent_24h: 7,
          health: { overall_status: 'operational' },
          last_updated: '2026-09-08T20:00:00Z',
        },
      });

      await act(async () => {
        render(
          <MemoryRouter initialEntries={['/dashboard']}>
            <OperationalDashboard data={mockData} user={mockUser} />
          </MemoryRouter>
        );
      });

      // Card Header
      expect(screen.getByText('AI Workforce')).toBeInTheDocument();
      expect(screen.getByText(/● Operational/)).toBeInTheDocument();

      // 4 Business Metrics
      expect(screen.getByText('Active Tasks')).toBeInTheDocument();
      expect(screen.getByText('Awaiting Review')).toBeInTheDocument();
      expect(screen.getByText('Blocked / Issues')).toBeInTheDocument();
      expect(screen.getByText('Completed (24h)')).toBeInTheDocument();

      // Link to workforce
      expect(screen.getByRole('button', { name: /View AI Workforce/i })).toBeInTheDocument();
    });

    it('renders System Health summary with 4 subsystems and overall health status', async () => {
      monitoringService.getHealthSummary.mockResolvedValueOnce({
        data: {
          overall_status: 'HEALTHY',
          evaluated_at: '2026-09-08T20:15:00Z',
          subsystems: {
            go_backend: { status: 'healthy', latency_ms: 12 },
            ai_sidecar: { status: 'healthy', latency_ms: 45 },
            database: { status: 'healthy', latency_ms: 4 },
            queue_worker: { status: 'healthy', active_workers: 4 },
          },
        },
      });

      await act(async () => {
        render(
          <MemoryRouter initialEntries={['/dashboard']}>
            <OperationalDashboard data={mockData} user={mockUser} />
          </MemoryRouter>
        );
      });

      // Card Header
      expect(screen.getByText(/System Health/)).toBeInTheDocument();
      expect(screen.getByText(/● Healthy/)).toBeInTheDocument();

      // 4 Subsystems
      expect(screen.getByText('Go API Backend')).toBeInTheDocument();
      expect(screen.getByText('Python AI Sidecar')).toBeInTheDocument();
      expect(screen.getByText('MariaDB Storage')).toBeInTheDocument();
      expect(screen.getByText('Queue Workers')).toBeInTheDocument();
    });

    it('gracefully handles partial service failure when health check rejects without breaking dashboard', async () => {
      monitoringService.getHealthSummary.mockRejectedValueOnce(new Error('Network error on health'));
      aiTaskService.getWorkforceSummary.mockResolvedValueOnce({
        data: {
          total_active_tasks: 0,
          waiting_for_approval_tasks: 0,
          failed_tasks: 0,
          completed_recent_24h: 12,
          health: { overall_status: 'operational' },
        },
      });

      await act(async () => {
        render(
          <MemoryRouter initialEntries={['/dashboard']}>
            <OperationalDashboard data={mockData} user={mockUser} />
          </MemoryRouter>
        );
      });

      // AI Workforce still operational
      expect(screen.getByText('AI Workforce')).toBeInTheDocument();
      expect(screen.getByText(/● Operational/)).toBeInTheDocument();

      // System Health gracefully indicates unavailable without throwing or breaking
      expect(screen.getByText(/System Health/)).toBeInTheDocument();
      expect(screen.getByText(/● Status unavailable/)).toBeInTheDocument();
    });

    it('gracefully handles workforce summary rejection without breaking dashboard', async () => {
      aiTaskService.getWorkforceSummary.mockRejectedValueOnce(new Error('Sidecar down'));
      monitoringService.getHealthSummary.mockResolvedValueOnce({
        data: {
          overall_status: 'HEALTHY',
          subsystems: {
            go_backend: { status: 'healthy' },
            ai_sidecar: { status: 'unavailable' },
            database: { status: 'healthy' },
            queue_worker: { status: 'healthy' },
          },
        },
      });

      await act(async () => {
        render(
          <MemoryRouter initialEntries={['/dashboard']}>
            <OperationalDashboard data={mockData} user={mockUser} />
          </MemoryRouter>
        );
      });

      // AI Workforce fallback
      expect(screen.getByText('AI Workforce')).toBeInTheDocument();
      expect(screen.getByText(/● Unavailable/)).toBeInTheDocument();

      // System Health still renders backend, database, and sidecar status
      expect(screen.getByText(/System Health/)).toBeInTheDocument();
      expect(screen.getByText('Go API Backend')).toBeInTheDocument();
      expect(screen.getByText('Python AI Sidecar')).toBeInTheDocument();
    });

    it('hides "View AI Workforce" and "Audit Logs" links when user lacks required RBAC permissions', async () => {
      const restrictedRBAC = {
        can: () => false,
        hasRole: () => false,
        hasAnyRole: () => false,
        roleName: 'Viewer',
        permissionsSet: new Set(),
      };

      await act(async () => {
        render(
          <RBACContext.Provider value={restrictedRBAC}>
            <MemoryRouter initialEntries={['/dashboard']}>
              <OperationalDashboard data={mockData} user={mockUser} />
            </MemoryRouter>
          </RBACContext.Provider>
        );
      });

      // Navigation actions should NOT be rendered
      expect(screen.queryByRole('button', { name: /View AI Workforce/i })).not.toBeInTheDocument();
      expect(screen.queryByRole('button', { name: /Audit Logs/i })).not.toBeInTheDocument();
      // High-level cards still render
      expect(screen.getByText('AI Workforce')).toBeInTheDocument();
      expect(screen.getByText(/System Health/)).toBeInTheDocument();
    });
  });

  describe('Dashboard Task 8: Content Deduplication and Information Density', () => {
    it('deduplicates upcoming reminders when entity is already featured in Priority Actions', async () => {
      const deduplicationData = {
        ...mockData,
        attention_items: [
          {
            id: 'imp_lead_101',
            urgency: 'IMPORTANT',
            category: 'Leads',
            source_entity_id: 101,
            title: 'Lead inquiry requires qualification',
            action_url: '/dashboard/leads',
          },
        ],
        upcoming_reminders: [
          {
            id: 'lead_rem_101',
            type: 'FOLLOW_UP',
            title: 'Follow up with Duplicate Lead Corp',
            subtitle: 'Regarding Lead inquiry',
            due_text: 'Today',
            action_url: '/dashboard/leads',
          },
          {
            id: 'contract_rem_909',
            type: 'CONTRACT_EXPIRY',
            title: 'Contract CTR-909 expires',
            subtitle: 'Maersk Carrier Agreement',
            due_text: 'In 5 days',
            action_url: '/dashboard/contracts',
          },
        ],
      };

      await act(async () => {
        render(
          <MemoryRouter initialEntries={['/dashboard']}>
            <OperationalDashboard data={deduplicationData} user={mockUser} />
          </MemoryRouter>
        );
      });

      // Priority Actions shows the lead inquiry
      expect(screen.getByText('Lead inquiry requires qualification')).toBeInTheDocument();

      // Upcoming Reminders ommits the duplicate lead reminder and renders the distinct contract reminder
      expect(screen.queryByText('Follow up with Duplicate Lead Corp')).not.toBeInTheDocument();
      expect(screen.getByText('Contract CTR-909 expires')).toBeInTheDocument();
    });

    it('prioritizes distinct pending approvals when top approval is already featured in Priority Actions', async () => {
      const approvalData = {
        ...mockData,
        attention_items: [
          {
            id: 'crit_app_1',
            urgency: 'CRITICAL',
            category: 'Approvals',
            approval_id: 1,
            title: 'Critical Quotation Approval',
            action_url: '/dashboard/approvals',
          },
        ],
        pending_approvals: [
          {
            id: 1,
            title: 'Critical Quotation Approval',
            requested_by: 'Alex Vance',
            role_badge: 'Manager',
            relative_age: '2h ago',
          },
          {
            id: 2,
            title: 'Credit Limit Extension Request',
            requested_by: 'Sarah Connor',
            role_badge: 'Director',
            relative_age: '10m ago',
          },
        ],
      };

      await act(async () => {
        render(
          <MemoryRouter initialEntries={['/dashboard']}>
            <OperationalDashboard data={approvalData} user={mockUser} />
          </MemoryRouter>
        );
      });

      // Priority Actions contains the critical action
      expect(screen.getByText('Critical Quotation Approval')).toBeInTheDocument();

      // Finance & Approvals section prioritizes the other distinct approval
      expect(screen.getByText('Credit Limit Extension Request')).toBeInTheDocument();
    });

    it('deduplicates recent activity stream against identical event IDs', async () => {
      const activityData = {
        ...mockData,
        recent_activity: [
          {
            id: 'evt_101',
            type: 'PAYMENT',
            title: 'Payment Received $9,000.00',
            subtitle: 'Invoice #88',
            timestamp: '5m ago',
            action_url: '/dashboard/invoices',
          },
          {
            id: 'evt_101', // Duplicate ID
            type: 'PAYMENT',
            title: 'Payment Received $9,000.00',
            subtitle: 'Invoice #88',
            timestamp: '5m ago',
            action_url: '/dashboard/invoices',
          },
          {
            id: 'evt_102',
            type: 'SHIPMENT',
            title: 'Container Loaded on Vessel',
            subtitle: 'SH-8891',
            timestamp: '20m ago',
            action_url: '/dashboard/shipments',
          },
        ],
      };

      await act(async () => {
        render(
          <MemoryRouter initialEntries={['/dashboard']}>
            <OperationalDashboard data={activityData} user={mockUser} />
          </MemoryRouter>
        );
      });

      // Exactly 1 instance of the payment activity item
      const paymentItems = screen.getAllByText('Payment Received $9,000.00');
      expect(paymentItems).toHaveLength(1);
      expect(screen.getByText('Container Loaded on Vessel')).toBeInTheDocument();
    });
  });

  describe('Dashboard Task 9: Visual Design Harmonization and UI Polish', () => {
    it('harmonizes card containers with consistent structure and semantic aria roles', async () => {
      await act(async () => {
        render(
          <MemoryRouter initialEntries={['/dashboard']}>
            <OperationalDashboard data={mockData} user={mockUser} />
          </MemoryRouter>
        );
      });

      // Assert semantic aria landmark sections exist for all 7 unified regions
      expect(screen.getByRole('region', { name: /KPI Summary/i })).toBeInTheDocument();
      expect(screen.getByRole('region', { name: /Priority Actions/i })).toBeInTheDocument();
      expect(screen.getByRole('region', { name: /Operations Overview/i })).toBeInTheDocument();
      expect(screen.getByRole('region', { name: /Finance and Approvals/i })).toBeInTheDocument();
      expect(screen.getByRole('region', { name: /Recent Business Activity/i })).toBeInTheDocument();
      expect(screen.getByRole('region', { name: /AI Summary and System Health/i })).toBeInTheDocument();
    });

    it('ensures all view-all and card-footer links are keyboard accessible with Enter/Space', async () => {
      await act(async () => {
        render(
          <MemoryRouter initialEntries={['/dashboard']}>
            <OperationalDashboard data={mockData} user={mockUser} />
          </MemoryRouter>
        );
      });

      // View-all links have role="button" and tabIndex=0
      const viewShipmentsLink = screen.getByRole('button', { name: /View All Shipments/i });
      expect(viewShipmentsLink).toHaveAttribute('tabindex', '0');

      const viewInvoicesLink = screen.getByRole('button', { name: /View All Invoices/i });
      expect(viewInvoicesLink).toHaveAttribute('tabindex', '0');

      const approvalsCenterLink = screen.getByRole('button', { name: /Approvals Center/i });
      expect(approvalsCenterLink).toHaveAttribute('tabindex', '0');

      // Card footer links have role="button" and tabIndex=0
      const manageShipmentsFooter = screen.getByRole('button', { name: /Manage all shipments and milestones/i });
      expect(manageShipmentsFooter).toHaveAttribute('tabindex', '0');

      const explorePipelineFooter = screen.getByRole('button', { name: /Explore sales conversion pipeline/i });
      expect(explorePipelineFooter).toHaveAttribute('tabindex', '0');

      // Test keyboard events dispatch cleanly without throwing errors
      expect(() => {
        fireEvent.keyDown(manageShipmentsFooter, { key: 'Enter', code: 'Enter' });
        fireEvent.keyDown(explorePipelineFooter, { key: ' ', code: 'Space' });
      }).not.toThrow();
    });

    it('renders light UI surfaces only with zero dark/black AI panels', async () => {
      let container;
      await act(async () => {
        const res = render(
          <MemoryRouter initialEntries={['/dashboard']}>
            <OperationalDashboard data={mockData} user={mockUser} />
          </MemoryRouter>
        );
        container = res.container;
      });

      // AI Summary and System Health cards must use light op-card treatment
      const aiCard = container.querySelector('.ai-summary-card');
      expect(aiCard).toBeInTheDocument();
      expect(aiCard.classList.contains('op-card')).toBe(true);

      const healthCard = container.querySelector('.system-health-card');
      expect(healthCard).toBeInTheDocument();
      expect(healthCard.classList.contains('op-card')).toBe(true);

      // Verify no dark mode or dark-bg classes exist on dashboard elements
      expect(container.querySelector('.dark')).toBeNull();
      expect(container.querySelector('.bg-black')).toBeNull();
      expect(container.querySelector('.bg-slate-900')).toBeNull();
    });
  });
});




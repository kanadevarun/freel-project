import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import EventWorkflowsTab from '../../../pages/dashboard/Automations/EventWorkflowsTab';
import { eventWorkflowsService } from '../../../services/eventWorkflowsService';

// Mock eventWorkflowsService
vi.mock('../../../services/eventWorkflowsService', () => ({
  eventWorkflowsService: {
    getOverview: vi.fn(),
    listWorkflows: vi.fn(),
    listEvents: vi.fn(),
    getWorkflow: vi.fn(),
    getEvent: vi.fn(),
    retryWorkflow: vi.fn(),
    cancelWorkflow: vi.fn(),
    simulateEvent: vi.fn(),
  },
}));

describe('EventWorkflowsTab Component', () => {
  const mockOverview = {
    total_events_ingested: 12,
    active_workflows_count: 3,
    awaiting_approval_count: 2,
    deduplicated_events_count: 5,
  };

  const mockWorkflows = [
    {
      id: 1,
      workflow_type: 'SHIPMENT_EXCEPTION_RESPONSE',
      trigger_event_type: 'shipment.milestone_missed',
      source_module: 'SHIPMENTS',
      source_record_type: 'SHIPMENT',
      source_record_id: '101',
      urgency: 'HIGH',
      ai_summary: 'Consignment Berth Delay exceeds 48h SLA. Customer notification recommended.',
      status: 'AWAITING_APPROVAL',
      action_name: 'shipments.notify_delay',
      approval_id: 210,
      recommendations: JSON.stringify([
        {
          action_name: 'shipments.notify_delay',
          title: 'Dispatch Delay Advisory',
          description: 'Notify customer of 48h berth delay.',
          risk_level: 'HIGH',
          requires_approval: true,
        },
      ]),
      correlation_id: 'corr-test-101',
      created_at: '2026-09-08T10:00:00Z',
    },
    {
      id: 2,
      workflow_type: 'INVOICE_COLLECTION_ESCALATION',
      trigger_event_type: 'invoice.overdue',
      source_module: 'FINANCE',
      source_record_type: 'INVOICE',
      source_record_id: '202',
      urgency: 'MEDIUM',
      ai_summary: 'Invoice is 24 days past due. Suggest sending polite collection reminder.',
      status: 'COMPLETED',
      action_name: 'finance.send_dunning_reminder',
      recommendations: JSON.stringify([]),
      correlation_id: 'corr-test-202',
      created_at: '2026-09-08T09:00:00Z',
    },
  ];

  const mockEvents = [
    {
      id: 1,
      event_type: 'shipment.milestone_missed',
      source_module: 'SHIPMENTS',
      source_record_type: 'SHIPMENT',
      source_record_id: '101',
      correlation_id: 'corr-test-101',
      dedup_key: 'evt:1:shipment.milestone_missed:SHIPMENT:101:corr-test-101',
      status: 'PROCESSED',
      payload: { delay_hours: 48 },
      created_at: '2026-09-08T10:00:00Z',
    },
  ];

  beforeEach(() => {
    vi.clearAllMocks();
    eventWorkflowsService.getOverview.mockResolvedValue(mockOverview);
    eventWorkflowsService.listWorkflows.mockResolvedValue({ workflows: mockWorkflows, total: 2 });
    eventWorkflowsService.listEvents.mockResolvedValue({ events: mockEvents, total: 1 });
  });

  it('renders KPI overview counters accurately', async () => {
    render(<EventWorkflowsTab />);

    await waitFor(() => {
      expect(screen.getByText('Ingested Domain Events')).toBeInTheDocument();
      expect(screen.getByText('12')).toBeInTheDocument();
      expect(screen.getByText('Active AI Workflows')).toBeInTheDocument();
      expect(screen.getByText('3')).toBeInTheDocument();
    });
  });

  it('renders workflows list with governance badges and summaries', async () => {
    render(<EventWorkflowsTab />);

    await waitFor(() => {
      expect(screen.getByText('#1')).toBeInTheDocument();
      expect(screen.getByText('SHIPMENT_EXCEPTION_RESPONSE')).toBeInTheDocument();
      expect(screen.getByText('shipment.milestone_missed')).toBeInTheDocument();
      expect(screen.getByText(/Approval #210/i)).toBeInTheDocument();
    });
  });

  it('opens detail drawer when clicking Details button', async () => {
    render(<EventWorkflowsTab />);

    await waitFor(() => {
      expect(screen.getAllByText('Details')[0]).toBeInTheDocument();
    });

    fireEvent.click(screen.getAllByText('Details')[0]);

    await waitFor(() => {
      expect(screen.getByText('Cross-Module Workflow #1')).toBeInTheDocument();
      expect(screen.getByText(/Human Approval Required/i)).toBeInTheDocument();
      expect(screen.getByText('AI Cross-Module Synthesis (Python Sidecar)')).toBeInTheDocument();
      expect(screen.getByText('Dispatch Delay Advisory')).toBeInTheDocument();
    });
  });

  it('switches to Event Store Ledger sub-tab', async () => {
    render(<EventWorkflowsTab />);

    await waitFor(() => {
      expect(screen.getByText(/Event Store Ledger/i)).toBeInTheDocument();
    });

    fireEvent.click(screen.getByText(/Event Store Ledger/i));

    await waitFor(() => {
      expect(screen.getByText('evt:1:shipment.milestone_missed:SHIPMENT:101:corr-test-101')).toBeInTheDocument();
    });
  });

  it('opens event simulation modal and triggers simulation', async () => {
    eventWorkflowsService.simulateEvent.mockResolvedValue({
      id: 3,
      workflow_type: 'SHIPMENT_EXCEPTION_RESPONSE',
      status: 'AWAITING_APPROVAL',
      urgency: 'HIGH',
      ai_summary: 'Simulated workflow created',
      recommendations: '[]',
    });

    render(<EventWorkflowsTab />);

    await waitFor(() => {
      expect(screen.getByText('Simulate Event')).toBeInTheDocument();
    });

    fireEvent.click(screen.getByText('Simulate Event'));

    await waitFor(() => {
      expect(screen.getByText('Simulate Domain Event Ingestion')).toBeInTheDocument();
    });

    fireEvent.click(screen.getByText('Ingest & Execute Workflow'));

    await waitFor(() => {
      expect(eventWorkflowsService.simulateEvent).toHaveBeenCalledWith(
        expect.objectContaining({
          event_type: 'shipment.milestone_missed',
          source_module: 'SHIPMENTS',
        })
      );
    });
  });
});

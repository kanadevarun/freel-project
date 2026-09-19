import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { BrowserRouter } from 'react-router-dom';
import WorkflowAutomationsPage from '../../../pages/dashboard/Automations/WorkflowAutomationsPage';
import { automationService } from '../../../services/automationService';

// Mock automationService
vi.mock('../../../services/automationService', () => ({
  automationService: {
    listAutomations: vi.fn(),
    getStats: vi.fn(),
    getSupportedTypes: vi.fn(),
    listExecutions: vi.fn(),
    runAutomation: vi.fn(),
    enableAutomation: vi.fn(),
    disableAutomation: vi.fn(),
    createAutomation: vi.fn(),
    updateAutomation: vi.fn(),
    deleteAutomation: vi.fn(),
    previewNextRun: vi.fn(),
    getExecutionRecommendations: vi.fn(),
    cancelExecution: vi.fn(),
    listInsights: vi.fn(),
    retryExecution: vi.fn(),
    acknowledgeInsight: vi.fn(),
    dismissInsight: vi.fn(),
    getExecutionInsights: vi.fn(),
  },
}));

describe('WorkflowAutomationsPage Component', () => {
  const mockAutomations = [
    {
      id: 1,
      org_id: 1,
      name: 'Daily Overdue Invoice & Receivables Review',
      automation_type: 'DAILY_OVERDUE_INVOICE_REVIEW',
      description: 'Reviews overdue customer invoices and creates collections follow-ups.',
      is_enabled: true,
      schedule_type: 'DAILY',
      schedule_time: '08:00',
      timezone: 'UTC',
      last_execution_status: 'COMPLETED',
      last_execution_at: '2026-09-08T08:00:00Z',
      next_execution_at: '2026-09-09T08:00:00Z',
    },
    {
      id: 2,
      org_id: 1,
      name: 'Shipment Exception Review',
      automation_type: 'SHIPMENT_EXCEPTION_REVIEW',
      description: 'Monitors delayed shipment milestones and customs holds.',
      is_enabled: false,
      schedule_type: 'DAILY',
      schedule_time: '07:30',
      timezone: 'UTC',
      last_execution_status: null,
      last_execution_at: null,
      next_execution_at: null,
    },
  ];

  const mockStats = {
    total_automations: 2,
    active_automations: 1,
    total_executions: 14,
    successful_executions: 14,
    failed_executions: 0,
    recommendations_generated: 18,
    recommendations_updated: 5,
    recent_success_rate: 100,
  };

  const mockSupportedTypes = [
    {
      type: 'DAILY_OVERDUE_INVOICE_REVIEW',
      name: 'Daily Overdue Invoice & Receivables Review',
      description: 'Evaluates overdue customer invoices and payment risk trends.',
      category: 'Finance & Collections',
      default_schedule_type: 'DAILY',
      default_schedule_time: '08:00',
      target_modules: ['invoices', 'customers', 'finance'],
      read_only_guarantee: 'Strictly read-only. Zero invoices altered.',
      recommended_action: 'Review collections recommendations.',
    },
    {
      type: 'SHIPMENT_EXCEPTION_REVIEW',
      name: 'Shipment Exception & Delayed Milestone Review',
      description: 'Reviews all active operational shipments and milestone delays.',
      category: 'Shipments & Operations',
      default_schedule_type: 'DAILY',
      default_schedule_time: '07:30',
      target_modules: ['shipments', 'tracking', 'exceptions'],
      read_only_guarantee: 'Strictly read-only. Zero shipment statuses changed.',
      recommended_action: 'Review carrier coordination recommendations.',
    },
  ];

  const mockExecutions = [
    {
      id: 101,
      automation_id: 1,
      automation_name: 'Daily Overdue Invoice & Receivables Review',
      automation_type: 'DAILY_OVERDUE_INVOICE_REVIEW',
      correlation_id: 'corr-exec-101',
      trigger_type: 'MANUAL',
      status: 'COMPLETED',
      queued_at: '2026-09-08T08:00:00Z',
      started_at: '2026-09-08T08:00:01Z',
      completed_at: '2026-09-08T08:00:02Z',
      duration_ms: 1200,
      records_reviewed: 12,
      recommendations_created: 2,
      recommendations_updated: 1,
      summary_text: 'Analysis completed successfully. Evaluated 12 records: created 2 new recommendations, refreshed 1 active items.',
    },
  ];

  beforeEach(() => {
    vi.clearAllMocks();
    automationService.listAutomations.mockResolvedValue({ success: true, automations: mockAutomations, total: 2 });
    automationService.getStats.mockResolvedValue({ success: true, stats: mockStats });
    automationService.getSupportedTypes.mockResolvedValue({ success: true, supported_types: mockSupportedTypes });
    automationService.listExecutions.mockResolvedValue({ success: true, executions: mockExecutions, total: 1 });
  });

  it('renders page header, KPI stats strip, and automations table', async () => {
    render(
      <BrowserRouter>
        <WorkflowAutomationsPage />
      </BrowserRouter>
    );

    expect(await screen.findByText('Workflow Automations & Scheduled AI Jobs')).toBeInTheDocument();
    expect(screen.getByText('Read-Only Safety Enforced')).toBeInTheDocument();

    // KPI values
    expect(screen.getByText('Configured pipelines')).toBeInTheDocument();
    expect(screen.getByText('Running on schedule')).toBeInTheDocument();

    // Table rows
    expect(screen.getAllByText('Daily Overdue Invoice & Receivables Review').length).toBeGreaterThan(0);
    expect(screen.getByText('Shipment Exception Review')).toBeInTheDocument();
  });

  it('renders empty state when no automations exist', async () => {
    automationService.listAutomations.mockResolvedValue({ success: true, automations: [], total: 0 });
    automationService.getStats.mockResolvedValue({ success: true, stats: { total_automations: 0, active_automations: 0 } });

    render(
      <BrowserRouter>
        <WorkflowAutomationsPage />
      </BrowserRouter>
    );

    expect(await screen.findByText('No Automations Configured')).toBeInTheDocument();
  });

  it('opens Create Automation modal with safety guarantee and template selection', async () => {
    render(
      <BrowserRouter>
        <WorkflowAutomationsPage />
      </BrowserRouter>
    );

    const createBtn = await screen.findByRole('button', { name: /Create Automation/i });
    fireEvent.click(createBtn);

    expect(await screen.findByText('Create Workflow Automation')).toBeInTheDocument();
    expect(screen.getByText('Read-Only Autonomous Safety Guarantee')).toBeInTheDocument();
    expect(screen.getByLabelText(/Automation Name/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/Schedule Cadence/i)).toBeInTheDocument();
  });

  it('switches to Execution History tab and renders past runs', async () => {
    render(
      <BrowserRouter>
        <WorkflowAutomationsPage />
      </BrowserRouter>
    );

    const historyTab = await screen.findByRole('button', { name: /Execution History/i });
    fireEvent.click(historyTab);

    expect(await screen.findByText('Audit Execution Log')).toBeInTheDocument();
    expect(screen.getByText('#101')).toBeInTheDocument();
    expect(screen.getByText('12 records')).toBeInTheDocument();
    expect(screen.getByText('+2 new')).toBeInTheDocument();
  });

  it('switches to Supported AI Assistant Jobs catalog and displays read-only guarantees', async () => {
    render(
      <BrowserRouter>
        <WorkflowAutomationsPage />
      </BrowserRouter>
    );

    const catalogTab = await screen.findByRole('button', { name: /Supported AI Assistant Jobs/i });
    fireEvent.click(catalogTab);

    expect(await screen.findByText('Approved AI Assistant Job Templates')).toBeInTheDocument();
    expect(screen.getByText('Finance & Collections')).toBeInTheDocument();
    expect(screen.getByText('Shipments & Operations')).toBeInTheDocument();
  });

  it('triggers manual run when Run button is clicked', async () => {
    automationService.runAutomation.mockResolvedValue({ success: true });

    render(
      <BrowserRouter>
        <WorkflowAutomationsPage />
      </BrowserRouter>
    );

    const runButtons = await screen.findAllByRole('button', { name: /Run/i });
    expect(runButtons.length).toBeGreaterThan(0);
    fireEvent.click(runButtons[0]);

    await waitFor(() => {
      expect(automationService.runAutomation).toHaveBeenCalledWith(1);
    });
  });

  it('switches to Operational Insights Feed tab and renders real operational signals', async () => {
    const mockInsights = [
      {
        id: 501,
        source_module: 'shipments',
        source_record_id: 101,
        source_record_ref: 'SHP-101',
        insight_type: 'SHIPMENT_EXCEPTION_DETECTED',
        severity: 'HIGH',
        title: 'Customs Hold Detected on Shipment #101',
        description: 'Shipment has an active customs hold exception with 48h delay variance.',
        detection_rule: 'SHIPMENT_EXCEPTION_DETECTED',
        confidence: 1.0,
        risk_level: 'HIGH',
        recommended_next_step: 'Contact customs broker and carrier.',
        is_approval_required: true,
        status: 'ACTIVE',
        created_at: '2026-09-08T07:00:00Z',
      },
    ];
    automationService.listInsights.mockResolvedValue({ success: true, insights: mockInsights, total: 1 });

    render(
      <BrowserRouter>
        <WorkflowAutomationsPage />
      </BrowserRouter>
    );

    const insightsTab = await screen.findByRole('button', { name: /Operational Insights Feed/i });
    fireEvent.click(insightsTab);

    expect(await screen.findByText('Customs Hold Detected on Shipment #101')).toBeInTheDocument();
    expect(screen.getByText('SHIPMENT_EXCEPTION_DETECTED')).toBeInTheDocument();
    expect(screen.getByText('Approval Required')).toBeInTheDocument();
    expect(screen.getByText('Contact customs broker and carrier.')).toBeInTheDocument();
  });

  it('renders retry button for failed executions and triggers retry on click', async () => {
    const failedExecutions = [
      {
        id: 999,
        automation_id: 1,
        automation_name: 'Daily Overdue Invoice & Receivables Review',
        automation_type: 'DAILY_OVERDUE_INVOICE_REVIEW',
        correlation_id: 'corr-exec-fail-999',
        trigger_type: 'MANUAL',
        status: 'FAILED',
        queued_at: '2026-09-08T07:00:00Z',
        error_message: 'Temporary network timeout',
        records_reviewed: 0,
        recommendations_created: 0,
        recommendations_updated: 0,
      },
    ];
    automationService.listExecutions.mockResolvedValue({ success: true, executions: failedExecutions, total: 1 });
    automationService.retryExecution.mockResolvedValue({ success: true });

    render(
      <BrowserRouter>
        <WorkflowAutomationsPage />
      </BrowserRouter>
    );

    const historyTab = await screen.findByRole('button', { name: /Execution History/i });
    fireEvent.click(historyTab);

    const retryBtn = await screen.findByRole('button', { name: /Retry/i });
    expect(retryBtn).toBeInTheDocument();
    fireEvent.click(retryBtn);

    await waitFor(() => {
      expect(automationService.retryExecution).toHaveBeenCalledWith(999);
    });
  });
});


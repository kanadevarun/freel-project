import React from 'react';
import { render, screen, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { MemoryRouter } from 'react-router-dom';
import AIWorkforceWidget from '../../components/dashboard/MissionControl/AIWorkforceWidget';
import aiTaskService from '../../services/aiTaskService';

vi.mock('../../services/aiTaskService', () => ({
  default: {
    getWorkforceSummary: vi.fn(),
    getWorkforceTasks: vi.fn(),
    retryTask: vi.fn(),
    cancelTask: vi.fn(),
  },
}));

describe('AIWorkforceWidget Component — Safety & Status Verification', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  const mockSummary = {
    counts: {
      total_active: 3,
      waiting_for_approval: 2,
      attention_required: 3,
      completed_24h: 15,
      processing: 2,
      queued: 1,
      failed: 2,
      stale: 1,
      avg_duration_ms: 1200,
    },
    agents: [
      {
        agent_key: 'pricing',
        name: 'Pricing Analyst',
        status: 'WAITING_FOR_APPROVAL',
        active_tasks: 1,
        completed_24h: 5,
      },
      {
        agent_key: 'operations',
        name: 'Operations Sentinel',
        status: 'FAILED',
        active_tasks: 0,
        completed_24h: 8,
      },
    ],
    mock_completed_tasks: 4,
    failover_completed_tasks: 1,
    health: {
      system_status: 'operational',
      sidecar_status: 'healthy',
      db_status: 'healthy',
      queue_responsive: true,
      provider_ready: true,
      last_worker_heartbeat: new Date().toISOString(),
    },
  };

  const mockTasks = {
    tasks: [
      {
        task_id: 'task-101',
        agent_key: 'pricing',
        agent_name: 'Pricing Analyst',
        business_module: 'PRICING',
        workforce_status: 'waiting_for_approval',
        related_ref: 'RFQ #501',
        can_retry: false,
        can_cancel: true,
        requires_approval: true,
        duration_ms: 1500,
        updated_at: new Date().toISOString(),
      },
      {
        task_id: 'task-102',
        agent_key: 'operations',
        agent_name: 'Operations Sentinel',
        business_module: 'OPERATIONS',
        workforce_status: 'failed',
        safe_error_msg: 'Carrier tracking format unrecognizable',
        related_ref: 'SHP #602',
        can_retry: true,
        can_cancel: false,
        requires_approval: false,
        duration_ms: 800,
        updated_at: new Date().toISOString(),
      },
      {
        task_id: 'task-103',
        agent_key: 'operations',
        agent_name: 'Operations Sentinel',
        business_module: 'OPERATIONS',
        workforce_status: 'stale',
        safe_error_msg: 'Worker lease expired',
        related_ref: 'SHP #603',
        can_retry: true,
        can_cancel: true,
        requires_approval: false,
        duration_ms: 45000,
        updated_at: new Date().toISOString(),
      },
    ],
    total: 3,
  };

  it('renders workforce summary with live operational counts and status badges', async () => {
    aiTaskService.getWorkforceSummary.mockResolvedValueOnce({ data: mockSummary });
    aiTaskService.getWorkforceTasks.mockResolvedValueOnce({ data: mockTasks });

    render(
      <MemoryRouter>
        <AIWorkforceWidget />
      </MemoryRouter>
    );

    // Initial loading indicator present
    expect(screen.getByText(/Connecting to AI workforce runtime/i)).toBeInTheDocument();

    // Live data rendered
    await waitFor(() => {
      expect(screen.getByRole('heading', { level: 3 })).toHaveTextContent(/AI Workforce Command & Telemetry/i);
      expect(screen.getAllByText(/Pricing Analyst/i).length).toBeGreaterThan(0);
      expect(screen.getAllByText(/Operations Sentinel/i).length).toBeGreaterThan(0);
      expect(screen.getByText(/Active Workflows/i)).toBeInTheDocument();
      expect(screen.getAllByText(/Awaiting Sign-Off/i).length).toBeGreaterThan(0);
    });

    // Verify waiting for approval tasks are shown distinctly with sign-off action
    expect(screen.getByText(/Sign Off/i)).toBeInTheDocument();
  });

  it('shows mock mode badge when mock completed tasks are reported', async () => {
    aiTaskService.getWorkforceSummary.mockResolvedValueOnce({ data: mockSummary });
    aiTaskService.getWorkforceTasks.mockResolvedValueOnce({ data: mockTasks });

    render(
      <MemoryRouter>
        <AIWorkforceWidget />
      </MemoryRouter>
    );

    await waitFor(() => {
      expect(screen.getByText(/executed in deterministic mock engine/i)).toBeInTheDocument();
    });
  });

  it('handles empty workforce state gracefully without throwing', async () => {
    aiTaskService.getWorkforceSummary.mockResolvedValueOnce({
      data: {
        counts: { total_active: 0, waiting_for_approval: 0, attention_required: 0, completed_24h: 0 },
        agents: [],
      },
    });
    aiTaskService.getWorkforceTasks.mockResolvedValueOnce({ data: { tasks: [], total: 0 } });

    render(
      <MemoryRouter>
        <AIWorkforceWidget />
      </MemoryRouter>
    );

    await waitFor(() => {
      expect(screen.getByText(/No AI tasks found matching the active criteria/i)).toBeInTheDocument();
    });
  });

  it('handles 401 unauthorized gracefully by displaying permission required state', async () => {
    aiTaskService.getWorkforceSummary.mockRejectedValueOnce({ status: 401 });
    aiTaskService.getWorkforceTasks.mockRejectedValueOnce({ status: 401 });

    render(
      <MemoryRouter>
        <AIWorkforceWidget />
      </MemoryRouter>
    );

    await waitFor(() => {
      expect(screen.getByText(/Operational Access Required/i)).toBeInTheDocument();
      expect(screen.getByText(/You do not have permission to view AI workforce operational telemetry/i)).toBeInTheDocument();
    });
  });

  it('handles service unavailable error state gracefully with retry control', async () => {
    aiTaskService.getWorkforceSummary.mockRejectedValueOnce(new Error('Connection timeout to backend.'));
    aiTaskService.getWorkforceTasks.mockRejectedValueOnce(new Error('Connection timeout to backend.'));

    render(
      <MemoryRouter>
        <AIWorkforceWidget />
      </MemoryRouter>
    );

    await waitFor(() => {
      expect(screen.getByText(/Workforce Telemetry Unavailable/i)).toBeInTheDocument();
      expect(screen.getByText(/Connection timeout to backend/i)).toBeInTheDocument();
      expect(screen.getByText(/Retry Connection/i)).toBeInTheDocument();
    });
  });

  it('correctly maps live Go backend workforce summary and task item envelope', async () => {
    const liveBackendSummary = {
      org_id: 2,
      total_active_tasks: 4,
      queued_tasks: 1,
      processing_tasks: 3,
      waiting_for_approval_tasks: 1,
      retrying_tasks: 0,
      failed_tasks: 2,
      stale_tasks: 0,
      completed_recent_24h: 18,
      completed_with_failover_24h: 2,
      completed_in_mock_mode_24h: 3,
      avg_duration_ms: 2400,
      active_agents_count: 2,
      by_agent: {
        pricing: {
          agent_key: 'pricing',
          display_name: 'Pricing Analyst',
          module: 'PRICING',
          active_tasks: 3,
          completed_24h: 10,
          failed_24h: 1,
          waiting_approvals: 1,
          status: 'ACTIVE',
        },
        compliance: {
          agent_key: 'compliance',
          display_name: 'Compliance Auditor',
          module: 'COMPLIANCE',
          active_tasks: 1,
          completed_24h: 8,
          failed_24h: 1,
          waiting_approvals: 0,
          status: 'ACTIVE',
        },
      },
      health: {
        overall_status: 'healthy',
        backend_status: 'healthy',
        sidecar_status: 'healthy',
        worker_status: 'healthy',
        queue_status: 'healthy',
        checkpoint_status: 'healthy',
        primary_provider_status: 'healthy',
        failover_status: 'ready',
        last_worker_heartbeat: new Date().toISOString(),
      },
    };

    const liveBackendTasks = {
      tasks: [
        {
          id: 42,
          org_id: 2,
          task_type: 'PRICING_ANALYZE',
          agent_key: 'pricing',
          agent_name: 'Pricing Analyst',
          module: 'PRICING',
          related_ref: 'RFQ #10042',
          related_id: '10042',
          status: 'processing',
          created_at: new Date().toISOString(),
          updated_at: new Date().toISOString(),
          retry_count: 0,
          max_retries: 3,
          is_stale: false,
          can_retry: false,
          can_cancel: true,
          navigation_url: '/dashboard/rfqs/10042',
        },
      ],
      total: 1,
    };

    aiTaskService.getWorkforceSummary.mockResolvedValueOnce({ data: liveBackendSummary });
    aiTaskService.getWorkforceTasks.mockResolvedValueOnce({ data: liveBackendTasks });

    render(
      <MemoryRouter>
        <AIWorkforceWidget />
      </MemoryRouter>
    );

    await waitFor(() => {
      expect(screen.getByText('AI Workforce Command & Telemetry')).toBeInTheDocument();
      expect(screen.getAllByText('Pricing Analyst').length).toBeGreaterThanOrEqual(1);
      expect(screen.getByText('Compliance Auditor')).toBeInTheDocument();
      expect(screen.getByText('RFQ #10042')).toBeInTheDocument();
      expect(screen.getByText('4')).toBeInTheDocument(); // total_active
    });
  });
});

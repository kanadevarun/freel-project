import React from 'react';
import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { MemoryRouter } from 'react-router-dom';
import AIMonitoringDashboardPage from '../../pages/dashboard/Monitoring/AIMonitoringDashboardPage';
import monitoringService from '../../services/monitoringService';

const mockSvc = vi.hoisted(() => ({
  getHealthSummary: vi.fn(),
  getPerformance: vi.fn(),
  getCost: vi.fn(),
  getQuality: vi.fn(),
  getQueue: vi.fn(),
  getRecommendations: vi.fn(),
  getApprovals: vi.fn(),
  getMemory: vi.fn(),
  getSecurity: vi.fn(),
  getPricing: vi.fn(),
  getThresholds: vi.fn(),
  listExecutions: vi.fn(),
  updateThreshold: vi.fn(),
}));

vi.mock('../../services/monitoringService', () => ({
  monitoringService: mockSvc,
  default: mockSvc,
}));

const mockHealth = {
  overall_status: 'HEALTHY',
  total_requests_24h: 125,
  success_rate_24h: 98.4,
  p95_latency_ms_24h: 420,
  queue_backlog_count: 3,
  estimated_cost_24h_usd: 0.125,
  security_events_24h: 0,
  failures_24h: 2,
  subsystems: {
    ai_runtime: { name: 'AI Runtime Gateway', status: 'HEALTHY', message: 'Online' },
    database: { name: 'MariaDB Telemetry', status: 'HEALTHY', message: 'Online' },
    ai_worker: { name: 'AI Queue Worker', status: 'HEALTHY', message: 'Online' },
  },
  active_alerts: [],
};

const mockPerf = {
  total_requests: 125,
  success_count: 123,
  failure_count: 2,
  success_rate: 98.4,
  latency: { average_ms: 210, p50_ms: 180, p95_ms: 420, p99_ms: 650 },
};

const mockCost = {
  total_estimated_cost: 0.125,
  currency: 'USD',
  is_estimated: true,
  total_tokens: 450000,
  by_model: [
    { dimension_name: 'gemini-1.5-flash', model_name: 'gemini-1.5-flash', provider: 'gemini', total_requests: 100, input_tokens: 200000, output_tokens: 100000, total_tokens: 300000, estimated_cost: 0.045, percentage_cost: 36.0 }
  ],
  by_feature: [],
};

const mockQuality = {
  total_evaluations: 15,
  overall_pass_rate: 93.3,
  grounding_valid_rate: 100.0,
  evidence_sufficient_rate: 93.3,
  recent_evaluations: [],
};

const mockQueue = {
  queued_jobs: 3,
  processing_jobs: 1,
  completed_24h: 85,
  failed_24h: 1,
  worker_health: 'HEALTHY',
  oldest_queued_age_sec: 12,
  average_duration_ms: 320,
};

const mockRecs = { total_generated: 45, active_count: 10, accepted_count: 25, acceptance_rate: 55.5 };
const mockApps = { total_proposed: 60, approved_count: 50, pending_count: 8, approval_failure_rate: 3.3 };
const mockMem = { active_personal_memories: 4, active_org_memories: 2, sensitive_rejections_count: 0 };
const mockSec = { total_security_events_24h: 0, recent_events: [] };

describe('AIMonitoringDashboardPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    monitoringService.getHealthSummary.mockResolvedValue({ data: mockHealth });
    monitoringService.getPerformance.mockResolvedValue({ data: mockPerf });
    monitoringService.getCost.mockResolvedValue({ data: mockCost });
    monitoringService.getQuality.mockResolvedValue({ data: mockQuality });
    monitoringService.getQueue.mockResolvedValue({ data: mockQueue });
    monitoringService.getRecommendations.mockResolvedValue({ data: mockRecs });
    monitoringService.getApprovals.mockResolvedValue({ data: mockApps });
    monitoringService.getMemory.mockResolvedValue({ data: mockMem });
    monitoringService.getSecurity.mockResolvedValue({ data: mockSec });
    monitoringService.getPricing.mockResolvedValue({ data: [] });
    monitoringService.getThresholds.mockResolvedValue({ data: [] });
    monitoringService.listExecutions.mockResolvedValue({ data: { items: [], total: 0 } });
  });

  it('renders dashboard title, health status, and top KPI cards', async () => {
    render(
      <MemoryRouter>
        <AIMonitoringDashboardPage />
      </MemoryRouter>
    );

    expect(screen.getByText(/AI Performance, Cost & Quality Monitoring/i)).toBeInTheDocument();

    await waitFor(() => {
      expect(screen.getByText('98.4%')).toBeInTheDocument();
      expect(screen.getByText('420ms')).toBeInTheDocument();
      expect(screen.getByText('$0.1250')).toBeInTheDocument();
    });
  });

  it('allows navigating between monitoring tabs', async () => {
    render(
      <MemoryRouter>
        <AIMonitoringDashboardPage />
      </MemoryRouter>
    );

    await waitFor(() => {
      expect(screen.getByText('Runtime & Executions')).toBeInTheDocument();
    });

    // Switch to Cost tab
    fireEvent.click(screen.getByText(/Cost & Token Breakdown/i));
    await waitFor(() => {
      expect(screen.getByText(/AI Cost & Token Consumption/i)).toBeInTheDocument();
    });

    // Switch to Quality tab
    fireEvent.click(screen.getByText(/Quality & Grounding Checks/i));
    await waitFor(() => {
      expect(screen.getByText(/Quality & Grounding Evaluations/i)).toBeInTheDocument();
    });
  });
});

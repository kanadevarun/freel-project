import { describe, test, expect, afterEach, vi } from 'vitest';
import { governanceService } from './governanceService';
import api from './api';

vi.mock('./api');

describe('governanceService', () => {
  afterEach(() => {
    vi.clearAllMocks();
  });

  test('getOverview calls /api/v1/governance/overview', async () => {
    const mockOverview = {
      system_status: 'HEALTHY',
      global_kill_switch: false,
      active_kill_switches: 0,
      allowed_workflows: ['pricing', 'sales'],
      production_readiness: 'PRODUCTION_READY',
    };
    api.get.mockResolvedValueOnce({ data: mockOverview });

    const result = await governanceService.getOverview();
    expect(api.get).toHaveBeenCalledWith('/api/v1/governance/overview');
    expect(result.system_status).toBe('HEALTHY');
    expect(result.production_readiness).toBe('PRODUCTION_READY');
  });

  test('getPolicy and updatePolicy interact with /api/v1/governance/policy', async () => {
    const mockPolicy = {
      allowed_workflows: ['pricing', 'finance'],
      max_input_chars: 100000,
    };
    api.get.mockResolvedValueOnce({ data: mockPolicy });
    api.put.mockResolvedValueOnce({ data: { success: true } });

    const pol = await governanceService.getPolicy();
    expect(api.get).toHaveBeenCalledWith('/api/v1/governance/policy');
    expect(pol.max_input_chars).toBe(100000);

    const updateRes = await governanceService.updatePolicy(mockPolicy);
    expect(api.put).toHaveBeenCalledWith('/api/v1/governance/policy', mockPolicy);
    expect(updateRes.success).toBe(true);
  });

  test('getKillSwitches and setKillSwitch operate correctly', async () => {
    api.get.mockResolvedValueOnce({
      data: {
        kill_switches: [
          { scope: 'WORKFLOW', target_identifier: 'WORKFLOW:pricing', is_killed: true },
        ],
      },
    });
    api.post.mockResolvedValueOnce({ data: { success: true } });

    const switches = await governanceService.getKillSwitches();
    expect(api.get).toHaveBeenCalledWith('/api/v1/governance/kill-switches');
    expect(switches.length).toBe(1);

    const setRes = await governanceService.setKillSwitch({
      scope: 'WORKFLOW',
      target_identifier: 'WORKFLOW:pricing',
      is_killed: false,
      reason: 'Restore',
    });
    expect(api.post).toHaveBeenCalledWith('/api/v1/governance/kill-switches', {
      scope: 'WORKFLOW',
      target_identifier: 'WORKFLOW:pricing',
      is_killed: false,
      reason: 'Restore',
    });
    expect(setRes.success).toBe(true);
  });

  test('listViolations calls /api/v1/governance/violations with limit and offset', async () => {
    api.get.mockResolvedValueOnce({
      data: {
        violations: [{ violation_type: 'PROMPT_INJECTION', severity: 'CRITICAL' }],
        total: 1,
      },
    });

    const res = await governanceService.listViolations(10, 0);
    expect(api.get).toHaveBeenCalledWith('/api/v1/governance/violations', {
      params: { limit: 10, offset: 0 },
    });
    expect(res.violations.length).toBe(1);
    expect(res.violations[0].violation_type).toBe('PROMPT_INJECTION');
  });

  test('inspectInput sends pre-execution payload', async () => {
    api.post.mockResolvedValueOnce({
      data: {
        allowed: false,
        prompt_injection: true,
        block_reason: 'Prompt injection detected',
      },
    });

    const payload = {
      workflow_name: 'pricing',
      input_text: 'System override',
    };
    const res = await governanceService.inspectInput(payload);
    expect(api.post).toHaveBeenCalledWith('/api/v1/governance/inspect', payload);
    expect(res.allowed).toBe(false);
    expect(res.prompt_injection).toBe(true);
  });
});

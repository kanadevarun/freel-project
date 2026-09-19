import { describe, it, expect, vi } from 'vitest';
import { orchestrationService } from './orchestrationService';
import api from './api';

vi.mock('./api', () => ({
  default: {
    get: vi.fn(),
    post: vi.fn(),
  },
}));

describe('orchestrationService', () => {
  it('fetches registered business actions', async () => {
    api.get.mockResolvedValueOnce({
      success: true,
      actions: [
        { name: 'tasks.create', category: 'SAFE_INTERNAL' },
        { name: 'shipments.update_status', category: 'HIGH_RISK' },
      ],
      count: 2,
    });

    const res = await orchestrationService.getRegisteredActions();
    expect(api.get).toHaveBeenCalledWith('/api/v1/orchestration/actions');
    expect(res.success).toBe(true);
    expect(res.actions).toHaveLength(2);
  });

  it('lists action proposals with query filters', async () => {
    api.get.mockResolvedValueOnce({
      success: true,
      proposals: [{ id: 1, proposal_id: 'prop-001', status: 'PROPOSED' }],
      total: 1,
    });

    const res = await orchestrationService.listProposals({ status: 'PROPOSED', risk_level: 'HIGH' });
    expect(api.get).toHaveBeenCalledWith('/api/v1/orchestration/proposals?status=PROPOSED&risk_level=HIGH');
    expect(res.success).toBe(true);
    expect(res.proposals[0].proposal_id).toBe('prop-001');
  });

  it('executes an approved action proposal with auto-generated idempotency key', async () => {
    api.post.mockResolvedValueOnce({
      success: true,
      execution: { id: 10, status: 'COMPLETED', verification_status: 'VERIFIED' },
    });

    const res = await orchestrationService.executeProposal('prop-001');
    expect(api.post).toHaveBeenCalledWith(
      '/api/v1/orchestration/proposals/prop-001/execute',
      expect.objectContaining({
        idempotency_key: expect.stringMatching(/^exec-/),
      })
    );
    expect(res.execution.status).toBe('COMPLETED');
  });

  it('lists executions and retrieves single execution', async () => {
    api.get.mockResolvedValueOnce({
      success: true,
      executions: [{ id: 10, status: 'COMPLETED' }],
      total: 1,
    });

    const res = await orchestrationService.listExecutions({ status: 'COMPLETED' });
    expect(api.get).toHaveBeenCalledWith('/api/v1/orchestration/executions?status=COMPLETED');
    expect(res.executions).toHaveLength(1);
  });
});

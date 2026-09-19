import { describe, it, expect, vi } from 'vitest';
import { copilotService } from './copilotService';
import api from './api';

vi.mock('./api', () => ({
  default: {
    get: vi.fn(),
    post: vi.fn(),
  },
}));

describe('copilotService', () => {
  it('sends contextual chat request with current module and filters', async () => {
    api.post.mockResolvedValueOnce({
      answer: 'Found 3 active shipments for MSCU.',
      confirmed_facts: ['MSCU has 3 active shipments.'],
      confidence: 0.95,
    });

    const res = await copilotService.chat({
      sessionId: 'ses-123',
      currentModule: 'SHIPMENTS',
      currentRoute: '/dashboard/shipments',
      query: 'What is our current container volume with MSCU?',
      activeFilters: { carrier: 'MSCU' },
    });

    expect(api.post).toHaveBeenCalledWith('/api/v1/copilot/chat', {
      session_id: 'ses-123',
      current_module: 'SHIPMENTS',
      current_route: '/dashboard/shipments',
      current_record_id: undefined,
      query: 'What is our current container volume with MSCU?',
      active_filters: { carrier: 'MSCU' },
    });
    expect(res.answer).toContain('Found 3 active shipments');
    expect(res.confidence).toBe(0.95);
  });

  it('lists conversation sessions', async () => {
    api.get.mockResolvedValueOnce({
      sessions: [{ id: 1, session_id: 'ses-1', title: 'Shipments Query' }],
      total: 1,
    });

    const res = await copilotService.listSessions(10);
    expect(api.get).toHaveBeenCalledWith('/api/v1/copilot/sessions?limit=10');
    expect(res.total).toBe(1);
  });

  it('executes a controlled copilot action', async () => {
    api.post.mockResolvedValueOnce({
      id: 5,
      status: 'PENDING_APPROVAL',
      action_type: 'REQUEST_HUMAN_APPROVAL',
      approval_id: 101,
    });

    const res = await copilotService.executeAction({
      actionType: 'REQUEST_HUMAN_APPROVAL',
      actionTitle: 'Approve Detention Fee Waiver',
      actionPayload: { fee: 250 },
      sessionId: 'ses-123',
      reason: 'Carrier agreed to demurrage waiver',
    });

    expect(api.post).toHaveBeenCalledWith('/api/v1/copilot/actions/execute', {
      action_type: 'REQUEST_HUMAN_APPROVAL',
      action_title: 'Approve Detention Fee Waiver',
      action_payload: { fee: 250 },
      session_id: 'ses-123',
      reason: 'Carrier agreed to demurrage waiver',
    });
    expect(res.status).toBe('PENDING_APPROVAL');
    expect(res.approval_id).toBe(101);
  });

  it('lists copilot action history', async () => {
    api.get.mockResolvedValueOnce({
      actions: [{ id: 1, action_type: 'CREATE_RECOMMENDATION', status: 'EXECUTED' }],
      total: 1,
    });

    const res = await copilotService.listActions('ses-123', 5);
    expect(api.get).toHaveBeenCalledWith('/api/v1/copilot/actions?session_id=ses-123&limit=5');
    expect(res.actions).toHaveLength(1);
  });
});

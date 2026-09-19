import { describe, it, expect, vi } from 'vitest';
import { recommendationService } from './recommendationService';
import api from './api';

vi.mock('./api', () => ({
  default: {
    get: vi.fn(),
    post: vi.fn(),
    patch: vi.fn(),
  },
}));

describe('recommendationService - Customer Follow-Up Assistant', () => {
  it('generates a draft message only upon explicit invocation', async () => {
    api.post.mockResolvedValueOnce({
      success: true,
      recommendation: {
        id: 42,
        draft_status: 'DRAFTED',
        draft_subject: 'Follow-up regarding Quotation QT-2026-101',
      },
    });

    const res = await recommendationService.generateDraft(42);
    expect(api.post).toHaveBeenCalledWith('/api/v1/recommendations/42/draft', {});
    expect(res.success).toBe(true);
    expect(res.recommendation.draft_status).toBe('DRAFTED');
  });

  it('saves an edited draft message', async () => {
    api.patch.mockResolvedValueOnce({
      success: true,
      recommendation: {
        id: 42,
        draft_status: 'EDITED',
        draft_subject: 'Edited Subject',
      },
    });

    const res = await recommendationService.saveDraft(42, 'Edited Subject', 'Updated body text');
    expect(api.patch).toHaveBeenCalledWith('/api/v1/recommendations/42/draft', {
      subject: 'Edited Subject',
      body: 'Updated body text',
    });
    expect(res.success).toBe(true);
  });

  it('creates an idempotent internal follow-up task', async () => {
    api.post.mockResolvedValueOnce({
      success: true,
      task: {
        id: 101,
        customer_name: 'Nordic Freight Dynamics AB',
        priority: 'high',
      },
    });

    const taskData = {
      assignee_name: 'Varun Kanade',
      priority: 'high',
      notes: 'Please review overdue balance.',
    };
    const res = await recommendationService.createFollowupTask(42, taskData);
    expect(api.post).toHaveBeenCalledWith('/api/v1/recommendations/42/task', taskData);
    expect(res.success).toBe(true);
    expect(res.task.id).toBe(101);
  });

  it('lists internal follow-up tasks with optional filters', async () => {
    api.get.mockResolvedValueOnce({
      success: true,
      tasks: [{ id: 101, customer_name: 'Nordic Freight Dynamics AB' }],
      total: 1,
    });

    const res = await recommendationService.listFollowupTasks({ customer_id: 102, status: 'OPEN' });
    expect(api.get).toHaveBeenCalledWith('/api/v1/recommendations/tasks?customer_id=102&status=OPEN');
    expect(res.tasks).toHaveLength(1);
  });

  it('retrieves follow-up assistant KPI stats', async () => {
    api.get.mockResolvedValueOnce({
      success: true,
      stats: {
        total_followups: 5,
        critical_count: 1,
        high_count: 2,
        tasks_created_count: 3,
        drafts_ready_count: 2,
      },
    });

    const res = await recommendationService.getFollowupStats();
    expect(api.get).toHaveBeenCalledWith('/api/v1/recommendations/followups/stats');
    expect(res.stats.total_followups).toBe(5);
  });

  it('retrieves controlled action preview for a recommendation', async () => {
    api.get.mockResolvedValueOnce({
      success: true,
      action_preview: {
        recommendation_id: 88,
        proposed_action: 'Request RFQ Clarification',
        expected_effect: 'Secures missing destination requirements',
        risk_level: 'MEDIUM',
      },
    });

    const res = await recommendationService.getActionPreview(88);
    expect(api.get).toHaveBeenCalledWith('/api/v1/recommendations/88/action-preview');
    expect(res.success).toBe(true);
    expect(res.action_preview.recommendation_id).toBe(88);
  });

  it('submits a recommendation to the HITL approval system', async () => {
    api.post.mockResolvedValueOnce({
      success: true,
      recommendation: {
        id: 88,
        approval_id: 12,
      },
    });

    const res = await recommendationService.requestApproval(88, 'Commercial pricing signoff required');
    expect(api.post).toHaveBeenCalledWith('/api/v1/recommendations/88/request-approval', {
      notes: 'Commercial pricing signoff required',
    });
    expect(res.success).toBe(true);
    expect(res.recommendation.approval_id).toBe(12);
  });

  it('applies RFQ and quotation filters in listRecommendations', async () => {
    api.get.mockResolvedValueOnce({
      success: true,
      recommendations: [],
    });

    await recommendationService.listRecommendations({
      rfq_id: 103,
      quotation_id: 102,
      missing_info: true,
      expiring_soon: true,
      pricing_concern: true,
    });

    expect(api.get).toHaveBeenCalledWith(
      '/api/v1/recommendations?rfq_id=103&quotation_id=102&missing_info=true&expiring_soon=true&pricing_concern=true'
    );
  });
});

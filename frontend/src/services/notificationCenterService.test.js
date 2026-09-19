import { describe, it, expect, vi, beforeEach } from 'vitest';
import { notificationCenterService } from './notificationCenterService';
import api from './api';

vi.mock('./api', () => ({
  default: {
    get: vi.fn(),
    post: vi.fn(),
    put: vi.fn(),
  },
}));

describe('notificationCenterService', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('lists notifications with filter parameters and pagination', async () => {
    api.get.mockResolvedValueOnce({
      data: [{ id: 1, title: 'Approval Required', severity: 'HIGH' }],
      total: 1,
      page: 1,
      page_size: 15,
    });

    const res = await notificationCenterService.listNotifications({
      page: 1,
      pageSize: 15,
      severity: 'HIGH',
      module: 'APPROVALS',
      is_read: 'false',
      action_required: 'true',
    });

    expect(api.get).toHaveBeenCalledWith(
      '/api/v1/notifications?page=1&page_size=15&severity=HIGH&module=APPROVALS&is_read=false&action_required=true'
    );
    expect(res.data).toHaveLength(1);
    expect(res.total).toBe(1);
  });

  it('retrieves single notification by ID', async () => {
    api.get.mockResolvedValueOnce({
      data: { id: 42, title: 'Overdue Invoice', severity: 'CRITICAL' },
    });

    const res = await notificationCenterService.getNotification(42);
    expect(api.get).toHaveBeenCalledWith('/api/v1/notifications/42');
    expect(res.data.id).toBe(42);
  });

  it('retrieves unread notification count', async () => {
    api.get.mockResolvedValueOnce({ count: 5 });

    const res = await notificationCenterService.getUnreadCount();
    expect(api.get).toHaveBeenCalledWith('/api/v1/notifications/unread-count');
    expect(res.count).toBe(5);
  });

  it('retrieves aggregate notification stats', async () => {
    api.get.mockResolvedValueOnce({
      data: {
        total: 12,
        unread: 5,
        action_required: 4,
        escalated: 2,
        critical_count: 2,
        high_count: 6,
      },
    });

    const res = await notificationCenterService.getStats();
    expect(api.get).toHaveBeenCalledWith('/api/v1/notifications/stats');
    expect(res.data.escalated).toBe(2);
  });

  it('marks a notification as read', async () => {
    api.post.mockResolvedValueOnce({ data: 'success' });

    const res = await notificationCenterService.markAsRead(42);
    expect(api.post).toHaveBeenCalledWith('/api/v1/notifications/42/read');
    expect(res.data).toBe('success');
  });

  it('marks a notification as unread', async () => {
    api.post.mockResolvedValueOnce({ data: 'success' });

    const res = await notificationCenterService.markAsUnread(42);
    expect(api.post).toHaveBeenCalledWith('/api/v1/notifications/42/unread');
    expect(res.data).toBe('success');
  });

  it('dismisses a notification', async () => {
    api.post.mockResolvedValueOnce({ data: 'success' });

    const res = await notificationCenterService.dismiss(42);
    expect(api.post).toHaveBeenCalledWith('/api/v1/notifications/42/dismiss');
    expect(res.data).toBe('success');
  });

  it('marks all notifications as read', async () => {
    api.post.mockResolvedValueOnce({ data: 'success' });

    const res = await notificationCenterService.markAllAsRead();
    expect(api.post).toHaveBeenCalledWith('/api/v1/notifications/read-all');
    expect(res.data).toBe('success');
  });

  it('lists escalation events for audit trail', async () => {
    api.get.mockResolvedValueOnce({
      data: [
        {
          id: 1,
          notification_id: 42,
          escalation_level: 1,
          escalation_reason: 'Approval SLA 48h exceeded',
          new_severity: 'CRITICAL',
        },
      ],
    });

    const res = await notificationCenterService.listEscalations(50);
    expect(api.get).toHaveBeenCalledWith('/api/v1/notifications/escalations?limit=50');
    expect(res.data).toHaveLength(1);
    expect(res.data[0].escalation_level).toBe(1);
  });

  it('retrieves and updates user preferences', async () => {
    const defaultPrefs = {
      min_severity: 'INFORMATIONAL',
      in_app_enabled: true,
      approvals_enabled: true,
    };
    api.get.mockResolvedValueOnce({ data: defaultPrefs });

    const resGet = await notificationCenterService.getPreferences();
    expect(api.get).toHaveBeenCalledWith('/api/v1/notifications/preferences');
    expect(resGet.data.in_app_enabled).toBe(true);

    const updatedPrefs = { ...defaultPrefs, min_severity: 'HIGH' };
    api.put.mockResolvedValueOnce({ data: updatedPrefs });

    const resPut = await notificationCenterService.updatePreferences(updatedPrefs);
    expect(api.put).toHaveBeenCalledWith('/api/v1/notifications/preferences', updatedPrefs);
    expect(resPut.data.min_severity).toBe('HIGH');
  });

  it('triggers on-demand evaluation sweep', async () => {
    api.post.mockResolvedValueOnce({ status: 'success', created: 3 });

    const res = await notificationCenterService.evaluate();
    expect(api.post).toHaveBeenCalledWith('/api/v1/notifications/evaluate');
    expect(res.created).toBe(3);
  });

  it('acknowledges a notification', async () => {
    api.post.mockResolvedValueOnce({ data: 'success' });
    const res = await notificationCenterService.acknowledge(42);
    expect(api.post).toHaveBeenCalledWith('/api/v1/notifications/42/acknowledge');
    expect(res.data).toBe('success');
  });

  it('snoozes a notification for specified duration', async () => {
    api.post.mockResolvedValueOnce({ data: 'success' });
    const res = await notificationCenterService.snooze(42, 240);
    expect(api.post).toHaveBeenCalledWith('/api/v1/notifications/42/snooze', { duration_minutes: 240 });
    expect(res.data).toBe('success');
  });

  it('escalates an unresolved notification with AI reasoning', async () => {
    api.post.mockResolvedValueOnce({ data: { id: 10, escalation_level: 2 } });
    const res = await notificationCenterService.escalate(42, 'Exceeded SLA');
    expect(api.post).toHaveBeenCalledWith('/api/v1/notifications/42/escalate', { reason: 'Exceeded SLA' });
    expect(res.data.escalation_level).toBe(2);
  });

  it('performs on-demand AI analysis and prioritization', async () => {
    api.post.mockResolvedValueOnce({ data: { ai_summary: 'Critical alert', priority_score: 92.5 } });
    const res = await notificationCenterService.analyzeAI(42);
    expect(api.post).toHaveBeenCalledWith('/api/v1/notifications/42/analyze-ai');
    expect(res.data.priority_score).toBe(92.5);
  });

  it('generates an AI-assisted escalation communication draft', async () => {
    api.post.mockResolvedValueOnce({ data: { subject: '[AI DRAFT] Alert', body_text: 'Urgent notice' } });
    const res = await notificationCenterService.generateDraft(42, 'OPERATIONAL_ALERT');
    expect(api.post).toHaveBeenCalledWith('/api/v1/notifications/42/generate-draft', { draft_type: 'OPERATIONAL_ALERT' });
    expect(res.data.subject).toBe('[AI DRAFT] Alert');
  });
});

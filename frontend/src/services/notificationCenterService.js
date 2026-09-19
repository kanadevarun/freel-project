import api from './api';

export const notificationCenterService = {
  /**
   * List notifications with filtering, sorting, and pagination
   */
  async listNotifications(params = {}) {
    const query = new URLSearchParams();
    if (params.page) query.append('page', params.page);
    if (params.pageSize) query.append('page_size', params.pageSize);
    if (params.severity && params.severity !== 'ALL') query.append('severity', params.severity);
    if (params.module && params.module !== 'ALL') query.append('module', params.module);
    if (params.is_read !== undefined && params.is_read !== '') query.append('is_read', params.is_read);
    if (params.action_required !== undefined && params.action_required !== '') {
      query.append('action_required', params.action_required);
    }
    if (params.is_escalated !== undefined && params.is_escalated !== '') {
      query.append('is_escalated', params.is_escalated);
    }
    if (params.delivery_status && params.delivery_status !== 'ALL') {
      query.append('delivery_status', params.delivery_status);
    }
    if (params.is_acknowledged !== undefined && params.is_acknowledged !== '') {
      query.append('is_acknowledged', params.is_acknowledged);
    }
    if (params.is_snoozed !== undefined && params.is_snoozed !== '') {
      query.append('is_snoozed', params.is_snoozed);
    }
    if (params.group_key) {
      query.append('group_key', params.group_key);
    }
    if (params.search) query.append('search', params.search);

    const qs = query.toString();
    const endpoint = qs ? `/api/v1/notifications?${qs}` : '/api/v1/notifications';
    return await api.get(endpoint);
  },

  /**
   * Get single notification by ID
   */
  async getNotification(id) {
    return await api.get(`/api/v1/notifications/${id}`);
  },

  /**
   * Get unread notification count
   */
  async getUnreadCount() {
    return await api.get('/api/v1/notifications/unread-count');
  },

  /**
   * Get notification stats (total, unread, action-required, escalated, critical, high)
   */
  async getStats() {
    return await api.get('/api/v1/notifications/stats');
  },

  /**
   * Mark a notification as read
   */
  async markAsRead(id) {
    return await api.post(`/api/v1/notifications/${id}/read`);
  },

  /**
   * Mark a notification as unread
   */
  async markAsUnread(id) {
    return await api.post(`/api/v1/notifications/${id}/unread`);
  },

  /**
   * Dismiss a notification
   */
  async dismiss(id) {
    return await api.post(`/api/v1/notifications/${id}/dismiss`);
  },

  /**
   * Mark all notifications as read
   */
  async markAllAsRead() {
    return await api.post('/api/v1/notifications/read-all');
  },

  /**
   * List escalation events for audit trail
   */
  async listEscalations(limit = 50) {
    return await api.get(`/api/v1/notifications/escalations?limit=${limit}`);
  },

  /**
   * Get user notification preferences
   */
  async getPreferences() {
    return await api.get('/api/v1/notifications/preferences');
  },

  /**
   * Update user notification preferences
   */
  async updatePreferences(data) {
    return await api.put('/api/v1/notifications/preferences', data);
  },

  /**
   * Trigger an on-demand evaluation sweep
   */
  async evaluate() {
    return await api.post('/api/v1/notifications/evaluate');
  },

  /**
   * Acknowledge a notification
   */
  async acknowledge(id) {
    return await api.post(`/api/v1/notifications/${id}/acknowledge`);
  },

  /**
   * Snooze a notification for given duration (in minutes)
   */
  async snooze(id, durationMinutes = 60) {
    return await api.post(`/api/v1/notifications/${id}/snooze`, { duration_minutes: durationMinutes });
  },

  /**
   * Escalate an unresolved notification with AI reasoning
   */
  async escalate(id, reason) {
    return await api.post(`/api/v1/notifications/${id}/escalate`, { reason });
  },

  /**
   * Run on-demand AI analysis and prioritization on a notification
   */
  async analyzeAI(id) {
    return await api.post(`/api/v1/notifications/${id}/analyze-ai`);
  },

  /**
   * Generate an AI-assisted escalation communication draft
   */
  async generateDraft(id, draftType = 'OPERATIONAL_ALERT') {
    return await api.post(`/api/v1/notifications/${id}/generate-draft`, { draft_type: draftType });
  },
};

export default notificationCenterService;

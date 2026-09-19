import api from './api';

export const copilotService = {
  /**
   * Send a contextual chat query to AI Copilot
   */
  async chat({ sessionId, currentModule, currentRoute, currentRecordId, query, activeFilters }) {
    return await api.post('/api/v1/copilot/chat', {
      session_id: sessionId,
      current_module: currentModule || 'DASHBOARD',
      current_route: currentRoute || window.location.pathname,
      current_record_id: currentRecordId ? String(currentRecordId) : undefined,
      query,
      active_filters: activeFilters || {},
    });
  },

  /**
   * List conversation sessions for current user and organization
   */
  async listSessions(limit = 20) {
    return await api.get(`/api/v1/copilot/sessions?limit=${limit}`);
  },

  /**
   * Get specific session details
   */
  async getSession(sessionId) {
    return await api.get(`/api/v1/copilot/sessions/${sessionId}`);
  },

  /**
   * Get messages for a session
   */
  async listMessages(sessionId, limit = 50) {
    return await api.get(`/api/v1/copilot/sessions/${sessionId}/messages?limit=${limit}`);
  },

  /**
   * Archive a conversation session
   */
  async archiveSession(sessionId) {
    return await api.post(`/api/v1/copilot/sessions/${sessionId}/archive`, {});
  },

  /**
   * Execute or submit a controlled copilot action (gated by Action System / Approval Center)
   */
  async executeAction({ actionType, actionTitle, actionPayload, sessionId, reason }) {
    return await api.post('/api/v1/copilot/actions/execute', {
      action_type: actionType,
      action_title: actionTitle,
      action_payload: actionPayload || {},
      session_id: sessionId,
      reason,
    });
  },

  /**
   * List logged copilot action history
   */
  async listActions(sessionId = '', limit = 20) {
    const qs = sessionId ? `?session_id=${sessionId}&limit=${limit}` : `?limit=${limit}`;
    return await api.get(`/api/v1/copilot/actions${qs}`);
  },
};

import api from './api';

export const aiTaskService = {
  /**
   * List AI processing tasks with optional filters
   */
  async listTasks(params = {}) {
    const query = new URLSearchParams();
    if (params.status) query.append('status', params.status);
    if (params.task_type) query.append('task_type', params.task_type);
    if (params.limit) query.append('limit', params.limit);
    if (params.offset) query.append('offset', params.offset);

    const queryString = query.toString();
    const endpoint = queryString ? `/api/v1/ai/tasks?${queryString}` : '/api/v1/ai/tasks';
    return await api.get(endpoint);
  },

  /**
   * Get queue statistics and metrics
   */
  async getTaskStats() {
    return await api.get('/api/v1/ai/tasks/stats');
  },

  /**
   * Get task details by ID
   */
  async getTaskById(id) {
    return await api.get(`/api/v1/ai/tasks/${id}`);
  },

  /**
   * Cancel an in-flight or queued task safely
   */
  async cancelTask(id, reason = '') {
    return await api.post(`/api/v1/ai/tasks/${id}/cancel`, { reason });
  },

  /**
   * Safely retry an exhausted or failed task
   */
  async retryTask(id) {
    return await api.post(`/api/v1/ai/tasks/${id}/retry`, {});
  },

  /**
   * Get organization-scoped AI workforce live summary
   */
  async getWorkforceSummary() {
    return await api.get('/api/v1/ai/workforce/summary');
  },

  /**
   * List organization-scoped AI tasks with comprehensive filtering and safe metadata
   */
  async getWorkforceTasks(params = {}) {
    const query = new URLSearchParams();
    if (params.status) query.append('status', params.status);
    if (params.agent_key) query.append('agent_key', params.agent_key);
    if (params.module) query.append('module', params.module);
    if (params.requires_approval !== undefined && params.requires_approval !== null) {
      query.append('requires_approval', String(params.requires_approval));
    }
    if (params.failed_or_stale !== undefined && params.failed_or_stale !== null) {
      query.append('failed_or_stale', String(params.failed_or_stale));
    }
    if (params.search) query.append('search', params.search);
    if (params.limit) query.append('limit', params.limit);
    if (params.offset) query.append('offset', params.offset);

    const qs = query.toString();
    const endpoint = qs ? `/api/v1/ai/workforce/tasks?${qs}` : '/api/v1/ai/workforce/tasks';
    return await api.get(endpoint);
  },

  /**
   * Get organization-scoped AI workforce health contract
   */
  async getWorkforceHealth() {
    return await api.get('/api/v1/ai/workforce/health');
  },
};

export default aiTaskService;

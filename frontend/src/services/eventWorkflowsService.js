import api from './api';

export const eventWorkflowsService = {
  /**
   * Get operational overview KPIs and metrics for event-driven workflows
   */
  async getOverview() {
    return await api.get('/api/v1/event-workflows/overview');
  },

  /**
   * List domain events from the event store
   */
  async listEvents(params = {}) {
    const query = new URLSearchParams();
    if (params.page) query.append('page', params.page);
    if (params.limit) query.append('limit', params.limit);
    if (params.event_type) query.append('event_type', params.event_type);
    if (params.status) query.append('status', params.status);

    const qs = query.toString();
    const endpoint = qs ? `/api/v1/event-workflows/events?${qs}` : '/api/v1/event-workflows/events';
    return await api.get(endpoint);
  },

  /**
   * Get specific event record by ID
   */
  async getEvent(id) {
    return await api.get(`/api/v1/event-workflows/events/${id}`);
  },

  /**
   * List cross-module AI workflows
   */
  async listWorkflows(params = {}) {
    const query = new URLSearchParams();
    if (params.page) query.append('page', params.page);
    if (params.limit) query.append('limit', params.limit);
    if (params.workflow_type) query.append('workflow_type', params.workflow_type);
    if (params.status) query.append('status', params.status);

    const qs = query.toString();
    const endpoint = qs ? `/api/v1/event-workflows/workflows?${qs}` : '/api/v1/event-workflows/workflows';
    return await api.get(endpoint);
  },

  /**
   * Get specific cross-module workflow instance by ID
   */
  async getWorkflow(id) {
    return await api.get(`/api/v1/event-workflows/workflows/${id}`);
  },

  /**
   * Manually retry a failed or stalled workflow instance
   */
  async retryWorkflow(id) {
    return await api.post(`/api/v1/event-workflows/workflows/${id}/retry`, {});
  },

  /**
   * Cancel an active or awaiting-approval workflow instance
   */
  async cancelWorkflow(id) {
    return await api.post(`/api/v1/event-workflows/workflows/${id}/cancel`, {});
  },

  /**
   * Ingest and simulate a domain business event
   */
  async simulateEvent(eventData) {
    return await api.post('/api/v1/event-workflows/simulate-event', eventData);
  },
};

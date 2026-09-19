import api from './api';

export const orchestrationService = {
  /**
   * List registered business actions supported by the platform registry
   */
  async getRegisteredActions() {
    return await api.get('/api/v1/orchestration/actions');
  },

  /**
   * List action proposals with optional filtering and pagination
   */
  async listProposals(params = {}) {
    const query = new URLSearchParams();
    if (params.limit) query.append('limit', params.limit);
    if (params.offset !== undefined) query.append('offset', params.offset);
    if (params.status) query.append('status', params.status);
    if (params.source_module) query.append('source_module', params.source_module);
    if (params.risk_level) query.append('risk_level', params.risk_level);
    if (params.search) query.append('search', params.search);

    const qs = query.toString();
    const endpoint = qs ? `/api/v1/orchestration/proposals?${qs}` : '/api/v1/orchestration/proposals';
    return await api.get(endpoint);
  },

  /**
   * Get single action proposal by numeric ID or string proposal_id
   */
  async getProposal(id) {
    return await api.get(`/api/v1/orchestration/proposals/${id}`);
  },

  /**
   * Request generation of a grounded action proposal
   */
  async generateProposal(data) {
    return await api.post('/api/v1/orchestration/proposals', data);
  },

  /**
   * Execute an approved action proposal with idempotency key
   */
  async executeProposal(proposalId, data = {}) {
    const payload = {
      idempotency_key: data.idempotency_key || `exec-${Date.now()}-${Math.random().toString(36).substring(2, 9)}`,
      override_params: data.override_params || null,
    };
    return await api.post(`/api/v1/orchestration/proposals/${proposalId}/execute`, payload);
  },

  /**
   * List action executions with optional filtering
   */
  async listExecutions(params = {}) {
    const query = new URLSearchParams();
    if (params.limit) query.append('limit', params.limit);
    if (params.offset !== undefined) query.append('offset', params.offset);
    if (params.status) query.append('status', params.status);
    if (params.action_name) query.append('action_name', params.action_name);
    if (params.proposal_id) query.append('proposal_id', params.proposal_id);

    const qs = query.toString();
    const endpoint = qs ? `/api/v1/orchestration/executions?${qs}` : '/api/v1/orchestration/executions';
    return await api.get(endpoint);
  },

  /**
   * Get single execution record by ID
   */
  async getExecution(id) {
    return await api.get(`/api/v1/orchestration/executions/${id}`);
  },

  /**
   * Cancel an in-flight or waiting execution
   */
  async cancelExecution(id, reason = '') {
    return await api.post(`/api/v1/orchestration/executions/${id}/cancel`, { reason });
  },

  /**
   * Retry a failed execution
   */
  async retryExecution(id) {
    return await api.post(`/api/v1/orchestration/executions/${id}/retry`, {});
  },
};

export default orchestrationService;

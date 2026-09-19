import api from './api';

export const approvalsService = {
  /**
   * List all approval requests with optional filters
   */
  async listApprovals(params = {}) {
    const query = new URLSearchParams();
    if (params.category) query.append('category', params.category);
    if (params.status) query.append('status', params.status);
    if (params.type) query.append('type', params.type);
    if (params.search) query.append('search', params.search);
    if (params.risk_level) query.append('risk_level', params.risk_level);
    if (params.source_module) query.append('source_module', params.source_module);

    const queryString = query.toString();
    const endpoint = queryString ? `/api/v1/approvals?${queryString}` : '/api/v1/approvals';
    return await api.get(endpoint);
  },

  /**
   * Get approval statistics & metrics
   */
  async getApprovalStats() {
    return await api.get('/api/v1/approvals/stats');
  },

  /**
   * Get approval details by ID
   */
  async getApprovalById(id) {
    return await api.get(`/api/v1/approvals/${id}`);
  },

  /**
   * Get rich action preview with live current & proposed state
   */
  async getActionPreview(id) {
    return await api.get(`/api/v1/approvals/${id}/preview`);
  },

  /**
   * Get immutable decision history trail for an approval
   */
  async getDecisionHistory(id) {
    return await api.get(`/api/v1/approvals/${id}/history`);
  },

  /**
   * Get execution status & error details
   */
  async getExecutionStatus(id) {
    return await api.get(`/api/v1/approvals/${id}/execution-status`);
  },

  /**
   * Retry failed execution for an approved action
   */
  async retryExecution(id) {
    return await api.post(`/api/v1/approvals/${id}/retry-execution`);
  },

  /**
   * Get linked AI recommendation details
   */
  async getRelatedRecommendation(id) {
    return await api.get(`/api/v1/approvals/${id}/recommendation`);
  },

  /**
   * Get linked source business record
   */
  async getRelatedSourceRecord(id) {
    return await api.get(`/api/v1/approvals/${id}/source-record`);
  },

  /**
   * Get audit log history for an approval
   */
  async getAuditHistory(id) {
    return await api.get(`/api/v1/approvals/${id}/audit`);
  },

  /**
   * Get approval requirements for a centralized action
   */
  async getApprovalRequirements(actionName) {
    return await api.get(`/api/v1/approvals/requirements?action=${encodeURIComponent(actionName)}`);
  },

  /**
   * Create a new approval request
   */
  async createApproval(input) {
    return await api.post('/api/v1/approvals', input);
  },

  /**
   * Approve an approval request
   */
  async approveRequest(id, notes = '') {
    return await api.post(`/api/v1/approvals/${id}/approve`, { notes });
  },

  /**
   * Reject an approval request
   */
  async rejectRequest(id, reason = '', notes = '') {
    return await api.post(`/api/v1/approvals/${id}/reject`, { reason, notes });
  },

  /**
   * Return an approval request for changes
   */
  async returnRequest(id, reason = '', notes = '') {
    return await api.post(`/api/v1/approvals/${id}/return`, { reason, notes });
  },

  /**
   * Cancel an approval request
   */
  async cancelRequest(id, notes = '') {
    return await api.post(`/api/v1/approvals/${id}/cancel`, { notes });
  },
};


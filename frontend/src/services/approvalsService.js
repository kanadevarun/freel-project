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
};

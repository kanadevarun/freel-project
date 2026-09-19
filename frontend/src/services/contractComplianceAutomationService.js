import api from './api';

/**
 * Service for Phase 3 Task 3.7: Contract and Compliance Automation and Intelligent Document Review
 */
export const contractComplianceAutomationService = {
  /**
   * Fetch comprehensive contract overview, deterministic signals, and recent reviews/drafts
   */
  getOverview: async (contractId) => {
    return api.get(`/api/v1/contracts/${contractId}/compliance-automation/overview`);
  },

  /**
   * Run AI multi-signal contract and document compliance review grounded in Go facts
   */
  reviewContract: async (contractId, data = {}) => {
    return api.post(`/api/v1/contracts/${contractId}/compliance-automation/review`, data);
  },

  /**
   * Run AI clause extraction (Payment, Liability, Demurrage, Termination, Governing Law, etc.)
   */
  extractClauses: async (contractId, data = {}) => {
    return api.post(`/api/v1/contracts/${contractId}/compliance-automation/extract-clauses`, data);
  },

  /**
   * Run AI structured database terms vs document text verification and discrepancy detection
   */
  verifyStructuredTerms: async (contractId, data = {}) => {
    return api.post(`/api/v1/contracts/${contractId}/compliance-automation/verify-terms`, data);
  },

  /**
   * Evaluate compliance obligations and mandatory document checklist
   */
  getComplianceChecklist: async (contractId) => {
    return api.get(`/api/v1/contracts/${contractId}/compliance-automation/compliance-checklist`);
  },

  /**
   * List all communication and clarification drafts for a contract
   */
  listDrafts: async (contractId) => {
    return api.get(`/api/v1/contracts/${contractId}/compliance-automation/drafts`);
  },

  /**
   * Generate an editable clarification or missing document draft
   */
  generateDraft: async (contractId, data) => {
    return api.post(`/api/v1/contracts/${contractId}/compliance-automation/drafts`, data);
  },

  /**
   * Retrieve a specific draft by ID
   */
  getDraft: async (contractId, draftId) => {
    return api.get(`/api/v1/contracts/${contractId}/compliance-automation/drafts/${draftId}`);
  },

  /**
   * Update an existing draft subject, message body, or recipient
   */
  updateDraft: async (contractId, draftId, data) => {
    return api.put(`/api/v1/contracts/${contractId}/compliance-automation/drafts/${draftId}`, data);
  },

  /**
   * Submit an approved draft to the Approvals Center (creates high-risk pending approval request)
   */
  submitApproval: async (contractId, draftId, data = {}) => {
    return api.post(`/api/v1/contracts/${contractId}/compliance-automation/drafts/${draftId}/submit-approval`, data);
  }
};

export default contractComplianceAutomationService;

import api from './api';

/**
 * Service for interacting with Phase 3 Task 3.4 RFQ-to-Quotation Automation & Intelligent Pricing Workflow
 */
export const rfqPricingWorkflowService = {
  /**
   * Fetch the comprehensive workflow overview for an RFQ
   */
  getOverview: async (rfqId) => {
    return api.get(`/api/v1/rfqs/${rfqId}/pricing-workflow/overview`);
  },

  /**
   * Run extraction evaluation to detect requirements and missing fields
   */
  extractRequirements: async (rfqId) => {
    return api.post(`/api/v1/rfqs/${rfqId}/pricing-workflow/extract-requirements`, {});
  },

  /**
   * Calculate or preview deterministic pricing based on rate selection and target margin
   */
  getPricingPreview: async (rfqId, params = {}) => {
    return api.post(`/api/v1/rfqs/${rfqId}/pricing-workflow/pricing-preview`, params);
  },

  /**
   * Generate an AI-grounded quotation draft
   */
  generateDraft: async (rfqId, data) => {
    return api.post(`/api/v1/rfqs/${rfqId}/pricing-workflow/drafts`, data);
  },

  /**
   * Fetch a specific quotation draft by ID
   */
  getDraft: async (rfqId, draftId) => {
    return api.get(`/api/v1/rfqs/${rfqId}/pricing-workflow/drafts/${draftId}`);
  },

  /**
   * Update an existing quotation draft
   */
  updateDraft: async (rfqId, draftId, data) => {
    return api.put(`/api/v1/rfqs/${rfqId}/pricing-workflow/drafts/${draftId}`, data);
  },

  /**
   * Submit an editable quotation draft for Managerial HITL Approval
   */
  submitForApproval: async (rfqId, draftId, reason = '') => {
    return api.post(`/api/v1/rfqs/${rfqId}/pricing-workflow/drafts/${draftId}/submit-approval`, {
      reason: reason || 'Standard quotation commercial review'
    });
  }
};

export default rfqPricingWorkflowService;

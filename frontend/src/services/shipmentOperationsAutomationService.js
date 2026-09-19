import api from './api';

/**
 * Service for Phase 3 Task 3.5: Shipment Operations Automation and Intelligent Exception Response
 */
export const shipmentOperationsAutomationService = {
  /**
   * Fetch comprehensive operational overview including deterministic signals,
   * prioritized exceptions, recommendations, and latest drafts.
   */
  getOverview: async (shipmentId) => {
    return api.get(`/api/v1/shipments/${shipmentId}/operations-automation/overview`);
  },

  /**
   * Run AI multi-signal risk analysis grounded in real Go backend facts
   */
  analyzeRisks: async (shipmentId) => {
    return api.post(`/api/v1/shipments/${shipmentId}/operations-automation/analyze-risks`, {});
  },

  /**
   * Prioritize active exceptions by business impact and urgency
   */
  prioritizeExceptions: async (shipmentId) => {
    return api.post(`/api/v1/shipments/${shipmentId}/operations-automation/prioritize-exceptions`, {});
  },

  /**
   * Fetch operational recommendations mapped to Action System
   */
  getRecommendations: async (shipmentId) => {
    return api.get(`/api/v1/shipments/${shipmentId}/operations-automation/recommendations`);
  },

  /**
   * List all communication drafts for a shipment
   */
  listDrafts: async (shipmentId) => {
    return api.get(`/api/v1/shipments/${shipmentId}/operations-automation/drafts`);
  },

  /**
   * Generate an editable communication draft (Carrier, Customer, or Internal Escalation)
   */
  generateDraft: async (shipmentId, data) => {
    return api.post(`/api/v1/shipments/${shipmentId}/operations-automation/drafts`, data);
  },

  /**
   * Retrieve a specific draft by ID
   */
  getDraft: async (shipmentId, draftId) => {
    return api.get(`/api/v1/shipments/${shipmentId}/operations-automation/drafts/${draftId}`);
  },

  /**
   * Update an existing communication draft
   */
  updateDraft: async (shipmentId, draftId, data) => {
    return api.put(`/api/v1/shipments/${shipmentId}/operations-automation/drafts/${draftId}`, data);
  },

  /**
   * Submit an editable communication draft for Managerial HITL Approval
   */
  submitForApproval: async (shipmentId, draftId, reason = '') => {
    return api.post(`/api/v1/shipments/${shipmentId}/operations-automation/drafts/${draftId}/submit-approval`, {
      reason: reason || 'Operational communication transmission approval'
    });
  }
};

export default shipmentOperationsAutomationService;

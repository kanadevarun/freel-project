import api from './api';

/**
 * Service for Phase 3 Task 3.6: Finance and Collections Automation and Intelligent Receivables Follow-Up
 */
export const financeCollectionsAutomationService = {
  /**
   * Fetch comprehensive receivables overview, deterministic aging metrics, customer balance, and existing drafts
   */
  getOverview: async (invoiceId) => {
    return api.get(`/api/v1/invoices/${invoiceId}/collections-automation/overview`);
  },

  /**
   * Run AI multi-signal receivables risk analysis grounded in deterministic Go calculations
   */
  analyzeReceivablesRisk: async (invoiceId) => {
    return api.post(`/api/v1/invoices/${invoiceId}/collections-automation/analyze-receivables`, {});
  },

  /**
   * Run AI collection prioritization scoring and priority ranking
   */
  prioritizeCollections: async (invoiceId) => {
    return api.post(`/api/v1/invoices/${invoiceId}/collections-automation/prioritize`, {});
  },

  /**
   * Analyze customer historical payment behavior, dispute history, and credit profile
   */
  getCustomerBehavior: async (invoiceId) => {
    return api.get(`/api/v1/invoices/${invoiceId}/collections-automation/customer-behavior`);
  },

  /**
   * Get operational recommendations mapped to the Centralized Action System
   */
  getRecommendations: async (invoiceId) => {
    return api.get(`/api/v1/invoices/${invoiceId}/collections-automation/recommendations`);
  },

  /**
   * List all communication drafts for an invoice
   */
  listDrafts: async (invoiceId) => {
    return api.get(`/api/v1/invoices/${invoiceId}/collections-automation/drafts`);
  },

  /**
   * Generate an editable collection message draft (Reminder, Overdue Notice, Final Demand, etc.)
   */
  generateDraft: async (invoiceId, data) => {
    return api.post(`/api/v1/invoices/${invoiceId}/collections-automation/drafts`, data);
  },

  /**
   * Retrieve a specific draft by ID
   */
  getDraft: async (invoiceId, draftId) => {
    return api.get(`/api/v1/invoices/${invoiceId}/collections-automation/drafts/${draftId}`);
  },

  /**
   * Update an existing draft subject, message body, or recipient
   */
  updateDraft: async (invoiceId, draftId, data) => {
    return api.put(`/api/v1/invoices/${invoiceId}/collections-automation/drafts/${draftId}`, data);
  },

  /**
   * Submit an approved or edited draft into the Managerial Approval workflow
   */
  submitForApproval: async (invoiceId, draftId, data) => {
    return api.post(`/api/v1/invoices/${invoiceId}/collections-automation/drafts/${draftId}/submit-approval`, data);
  },
};

export default financeCollectionsAutomationService;

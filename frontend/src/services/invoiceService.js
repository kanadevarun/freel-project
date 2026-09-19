import api from './api';

/**
 * Invoice & Finance Intelligence Service (Phase 1 Task 1.5)
 * Communicates with backend endpoints for deterministic financial analytics,
 * aging buckets, receivables exposure, line item audits, and grounded AI finance insights.
 */
export const invoiceService = {
  /**
   * Fetch complete 360-degree invoice and finance intelligence for a single invoice.
   * Strictly read-only.
   * @param {number|string} invoiceId
   * @returns {Promise<Object>}
   */
  getInvoice360FinanceIntelligence: async (invoiceId) => {
    return api.get(`/api/v1/invoices/${invoiceId}/intelligence`);
  },

  /**
   * Fetch organization-level receivables and exposure summary.
   * Strictly read-only.
   * @returns {Promise<Object>}
   */
  getOrgFinanceSummary: async () => {
    return api.get('/api/v1/invoices/finance-summary');
  },
};

export default invoiceService;

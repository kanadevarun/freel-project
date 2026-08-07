import api from './api';

/**
 * carrierService — fetches carrier rate options and approves quotes.
 *
 * These are called from the Pricing Workspace page only.
 * The backend reads the RFQ's origin/destination from the DB and
 * delegates to the FF partner API (or mock) to return ranked rates.
 */
export const carrierService = {
  /**
   * Fetch ranked carrier rate options for a given RFQ.
   *
   * Returns a FetchRatesResponse containing:
   *   - rates[]:  Array of RichCarrierRate (sorted best→worst)
   *   - overall_reasoning: AI summary string
   *   - recommended_idx:   Index of the recommended carrier
   *   - fetched_at:        ISO timestamp (for "valid for X hours" display)
   *
   * @param {number} rfqId  — ID of the RFQ to fetch rates for
   */
  getCarrierRates: async (rfqId) => {
    return api.get(`/api/v1/rfqs/${rfqId}/carrier-rates`);
  },

  /**
   * Approve a specific quote and advance the RFQ to QUOTE_SENT.
   *
   * This is the "Approve & Send" action in the Pricing Workspace.
   * The backend will:
   *   1. Mark the chosen quote as APPROVED
   *   2. Mark all other quotes as REJECTED
   *   3. Advance RFQ stage to QUOTE_SENT
   *   4. Fire EventQuoteSent (which triggers email notification)
   *
   * @param {number} rfqId    — ID of the RFQ
   * @param {number} quoteId  — ID of the quote to approve
   */
  approveQuote: async (rfqId, quoteId) => {
    return api.post(`/api/v1/rfqs/${rfqId}/approve-quote`, { quote_id: quoteId });
  },
};

import api from './api';

/**
 * Service for interacting with the RFQ & Shipment Initiation API.
 */
export const rfqService = {
  /**
   * Fetch a paginated list of RFQs.
   */
  listRFQs: async (limit = 50, offset = 0) => {
    return api.get(`/api/v1/rfqs?limit=${limit}&offset=${offset}`);
  },

  /**
   * Fetch a single RFQ with its items and quotes.
   */
  getRFQ: async (id) => {
    return api.get(`/api/v1/rfqs/${id}`);
  },

  /**
   * Create a new RFQ.
   */
  createRFQ: async (data) => {
    return api.post('/api/v1/rfqs', data);
  },

  /**
   * Update the stage of an RFQ.
   */
  updateStage: async (id, stage) => {
    return api.put(`/api/v1/rfqs/${id}/stage`, { stage });
  },

  /**
   * Use AI to parse unstructured text (email, OCR) into structured shipment details.
   */
  parseShipmentRequest: async (rawText) => {
    return api.post('/api/v1/rfqs/parse-shipment-request', { raw_text: rawText });
  },

  /**
   * Add a carrier quote to an RFQ.
   */
  addQuote: async (id, quoteData) => {
    return api.post(`/api/v1/rfqs/${id}/quotes`, quoteData);
  },

  /**
   * Fetch the unified chronological timeline of events, tasks, and comms.
   */
  getTimeline: async (id) => {
    return api.get(`/api/v1/rfqs/${id}/timeline`);
  },

  /**
   * Get the deterministic requirements evaluation for an RFQ.
   * Returns operational readiness, grouped requirements, document requirements, and AI findings.
   * The backend is the single source of truth — no frontend readiness computation.
   */
  getRequirements: async (id) => {
    return api.get(`/api/v1/rfqs/${id}/requirements`);
  },

  /**
   * Get the complete operational timeline and audit trail for an RFQ.
   * Aggregates Lead, interactions, AI tasks, requirements, and quotes.
   */
  getActivity: async (id) => {
    return api.get(`/api/v1/rfqs/${id}/activity`);
  },

  /**
   * Get all resolved document requirements and active records for an RFQ.
   */
  getDocuments: async (id) => {
    return api.get(`/api/v1/rfqs/${id}/documents`);
  },

  /**
   * Create or attach a document for an RFQ.
   */
  createDocument: async (id, data) => {
    return api.post(`/api/v1/rfqs/${id}/documents`, data);
  },

  /**
   * Update the lifecycle review status of a document.
   */
  updateDocumentStatus: async (id, documentId, data) => {
    return api.patch(`/api/v1/rfqs/${id}/documents/${documentId}/status`, data);
  },

  /**
   * Delete a document from an RFQ.
   */
  deleteDocument: async (id, documentId) => {
    return api.delete(`/api/v1/rfqs/${id}/documents/${documentId}`);
  },

  /**
   * ──────────────────────────────────────────────────────────────────────────
   * Quote Management & Commercial Decision Intelligence API (Task 13)
   * ──────────────────────────────────────────────────────────────────────────
   */

  /**
   * Fetch all carrier quotes, comparison intelligence, and commercial summary for an RFQ.
   */
  getQuotes: async (id) => {
    return api.get(`/api/v1/rfqs/${id}/quotes`);
  },

  /**
   * Create a new carrier quote for an RFQ.
   */
  createQuote: async (id, data) => {
    return api.post(`/api/v1/rfqs/${id}/quotes`, data);
  },

  /**
   * Update commercial or operational details of an existing quote.
   */
  updateQuote: async (id, quoteId, data) => {
    return api.patch(`/api/v1/rfqs/${id}/quotes/${quoteId}`, data);
  },

  /**
   * Transition the lifecycle status of a quote (e.g. RECEIVED -> UNDER_REVIEW).
   */
  updateQuoteStatus: async (id, quoteId, data) => {
    return api.patch(`/api/v1/rfqs/${id}/quotes/${quoteId}/status`, data);
  },

  /**
   * Mark a quote as the recommended operational option.
   */
  recommendQuote: async (id, quoteId) => {
    return api.post(`/api/v1/rfqs/${id}/quotes/${quoteId}/recommend`);
  },

  /**
   * Approve a quote for operations and advance RFQ stage to QUOTE_GENERATED.
   */
  approveQuote: async (id, quoteId, data = {}) => {
    return api.post(`/api/v1/rfqs/${id}/quotes/${quoteId}/approve`, data);
  },

  /**
   * Select an approved quote for customer quotation presentation.
   */
  selectQuoteForCustomer: async (id, quoteId) => {
    return api.post(`/api/v1/rfqs/${id}/quotes/${quoteId}/select`);
  },

  /**
   * Delete or withdraw a carrier quote.
   */
  deleteQuote: async (id, quoteId) => {
    return api.delete(`/api/v1/rfqs/${id}/quotes/${quoteId}`);
  },

  /**
   * ──────────────────────────────────────────────────────────────────────────
   * Booking & Shipment Handoff Relationship API (Task 14)
   * ──────────────────────────────────────────────────────────────────────────
   */

  /**
   * Fetch booking handoff status, eligibility, and linked carrier bookings for an RFQ.
   */
  getRFQBookings: async (id) => {
    return api.get(`/api/v1/rfqs/${id}/bookings`);
  },

  /**
   * Create a carrier booking handoff record from the approved RFQ quotation.
   */
  createRFQBooking: async (id, data) => {
    return api.post(`/api/v1/rfqs/${id}/bookings`, data);
  },

  /**
   * Update the status of a carrier booking (e.g. DRAFT -> REQUESTED -> CONFIRMED).
   */
  updateRFQBookingStatus: async (id, bookingId, data) => {
    return api.patch(`/api/v1/rfqs/${id}/bookings/${bookingId}/status`, data);
  },

  /**
   * Fetch shipment handoff and execution status linked to an RFQ.
   */
  getRFQShipments: async (id) => {
    return api.get(`/api/v1/rfqs/${id}/shipments`);
  },
};

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
  }
};


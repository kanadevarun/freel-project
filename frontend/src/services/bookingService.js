import api from './api';

/**
 * Service for interacting with the Dedicated Booking Operations Workspace API.
 * All data derives from real persisted database records.
 */
export const bookingService = {
  /**
   * Fetch paginated bookings workspace with search, filter, and live KPIs.
   * @param {Object} params - { page, limit, status, carrier, origin_port, destination_port, search, sort_by, sort_dir }
   */
  getBookings: async (params = {}) => {
    const query = new URLSearchParams();
    if (params.page) query.append('page', params.page);
    if (params.limit) query.append('limit', params.limit);
    if (params.status && params.status !== 'ALL') query.append('status', params.status);
    if (params.carrier) query.append('carrier', params.carrier);
    if (params.origin_port) query.append('origin_port', params.origin_port);
    if (params.destination_port) query.append('destination_port', params.destination_port);
    if (params.search) query.append('search', params.search);
    if (params.sort_by) query.append('sort_by', params.sort_by);
    if (params.sort_dir) query.append('sort_dir', params.sort_dir);
    if (params.etd_from) query.append('etd_from', params.etd_from);
    if (params.etd_to) query.append('etd_to', params.etd_to);

    const queryString = query.toString();
    const url = `/api/v1/bookings${queryString ? `?${queryString}` : ''}`;
    return api.get(url);
  },

  /**
   * Fetch full workspace detail for a specific booking.
   * Includes source RFQ, commercial quote, cargo summary, linked shipment, and activity timeline.
   */
  getBookingDetail: async (bookingId) => {
    return api.get(`/api/v1/bookings/${bookingId}`);
  },

  /**
   * Direct Booking status transition (e.g. REQUESTED -> CONFIRMED, CANCELLED).
   */
  updateBookingStatus: async (bookingId, status, notes = '') => {
    return api.patch(`/api/v1/bookings/${bookingId}/status`, { status, notes });
  },

  /**
   * Fetch eligible RFQs with approved quotes for creating new bookings.
   */
  getEligibleRFQs: async () => {
    return api.get('/api/v1/bookings/eligible-rfqs');
  },

  /**
   * Create a booking directly from an RFQ with approved quotation.
   */
  createBooking: async (rfqId, bookingData) => {
    return api.post(`/api/v1/rfqs/${rfqId}/bookings`, bookingData);
  },

  /**
   * Create or handoff an operational shipment from a CONFIRMED booking.
   */
  createShipment: async (bookingId, shipmentData = {}) => {
    return api.post(`/api/v1/bookings/${bookingId}/shipments`, shipmentData);
  },

  // ── Task 5: Live Carrier Booking Integration ────────────────────────────────

  /**
   * Submit booking directly to connected Carrier API (e.g. Maersk, MSC).
   * @param {number|string} bookingId
   * @param {Object} payload - { force_retry, special_instructions }
   */
  bookWithCarrier: async (bookingId, payload = {}) => {
    return api.post(`/api/v1/bookings/${bookingId}/carrier-book`, payload);
  },

  /**
   * Synchronize carrier booking status and vessel/container allocation.
   * @param {number|string} bookingId
   */
  syncCarrierBooking: async (bookingId) => {
    return api.post(`/api/v1/bookings/${bookingId}/carrier-sync`, {});
  },
};

export default bookingService;

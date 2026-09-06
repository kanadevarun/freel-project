import api from './api';

/**
 * searchService — provides tenant-scoped, real-time deterministic global search
 * across Shipments, Bookings, RFQs, Quotes, Customers, Invoices, Leads, Contracts, Documents, and Tracking.
 */
export const searchService = {
  /**
   * Search across all business data
   * @param {string} query
   * @param {string} [type] Optional category filter ('SHIPMENT', 'BOOKING', 'RFQ', 'CUSTOMER', 'INVOICE', 'LEAD', etc.)
   * @param {number} [limit=30]
   */
  globalSearch: async (query, type = '', limit = 30) => {
    if (!query || !query.trim()) {
      return { data: { query: '', total_matches: 0, groups: [] } };
    }
    const searchParams = new URLSearchParams();
    searchParams.set('q', query.trim());
    if (type && type !== 'ALL') searchParams.set('type', type);
    if (limit) searchParams.set('limit', limit);
    return api.get(`/api/v1/search?${searchParams.toString()}`);
  },
};

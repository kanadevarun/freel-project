import api from './api';

const BASE = '/api/v1/rates';

export const rateService = {
  /**
   * Get KPI metrics & summary cards data for Rate Management.
   * @returns {Promise<RateSummaryKPIs>}
   */
  getRateSummary: () => api.get(`${BASE}/summary`),

  /**
   * List rate records with filters & pagination.
   * @param {Object} params - { search, status, rate_type, transport_mode, service_type, equipment_type, carrier_name, origin, destination, valid_date, page, limit, sort_by, sort_order }
   */
  listRates: (params = {}) => {
    const q = new URLSearchParams();
    if (params.search)        q.set('search', params.search);
    if (params.status && params.status !== 'ALL') q.set('status', params.status);
    if (params.rate_type && params.rate_type !== 'ALL') q.set('rate_type', params.rate_type);
    if (params.transport_mode && params.transport_mode !== 'ALL') q.set('transport_mode', params.transport_mode);
    if (params.service_type && params.service_type !== 'ALL') q.set('service_type', params.service_type);
    if (params.equipment_type && params.equipment_type !== 'ALL') q.set('equipment_type', params.equipment_type);
    if (params.carrier_name && params.carrier_name !== 'ALL') q.set('carrier_name', params.carrier_name);
    if (params.origin)        q.set('origin', params.origin);
    if (params.destination)   q.set('destination', params.destination);
    if (params.valid_date)    q.set('valid_date', params.valid_date);
    if (params.page)          q.set('page', params.page);
    if (params.limit)         q.set('limit', params.limit);
    if (params.sort_by)       q.set('sort_by', params.sort_by);
    if (params.sort_order)    q.set('sort_order', params.sort_order);

    return api.get(`${BASE}/?${q.toString()}`);
  },

  /**
   * Get complete details for a single rate record.
   * @param {number|string} id
   */
  getRate: (id) => api.get(`${BASE}/${id}`),

  /**
   * Create a new Rate record.
   * @param {Object} payload - CreateRateRequest
   */
  createRate: (payload) => api.post(`${BASE}/`, payload),

  /**
   * Update an existing Rate record.
   * @param {number|string} id
   * @param {Object} payload - UpdateRateRequest
   */
  updateRate: (id, payload) => api.put(`${BASE}/${id}`, payload),

  /**
   * Archive a rate record.
   * @param {number|string} id
   */
  archiveRate: (id) => api.post(`${BASE}/${id}/archive`, {}),

  /**
   * Legacy spot / canonical rate search used for spot quote benchmarking.
   * @param {Object} params - { origin, destination, equipment, incoterms }
   */
  searchRates: (params = {}) => {
    const q = new URLSearchParams(params);
    return api.get(`${BASE}/search?${q.toString()}`);
  },

  // ── Task 19.2: Rate Charges & Commercial Pricing Breakdown ──────────────────

  /**
   * Get deterministic pricing breakdown and charge items for a rate.
   * @param {number|string} rateId
   */
  getRatePricing: (rateId) => api.get(`${BASE}/${rateId}/pricing`),

  /**
   * Add a new charge item to a rate.
   * @param {number|string} rateId
   * @param {Object} payload - CreateRateChargeRequest
   */
  addRateCharge: (rateId, payload) => api.post(`${BASE}/${rateId}/charges`, payload),

  /**
   * Update an existing charge item on a rate.
   * @param {number|string} rateId
   * @param {number|string} chargeId
   * @param {Object} payload - UpdateRateChargeRequest
   */
  updateRateCharge: (rateId, chargeId, payload) => api.put(`${BASE}/${rateId}/charges/${chargeId}`, payload),

  /**
   * Delete a charge item from a rate.
   * @param {number|string} rateId
   * @param {number|string} chargeId
   */
  deleteRateCharge: (rateId, chargeId) => api.delete(`${BASE}/${rateId}/charges/${chargeId}`),

  /**
   * Reorder charge items on a rate.
   * @param {number|string} rateId
   * @param {Array<number>} chargeIds
   */
  reorderRateCharges: (rateId, chargeIds) => api.post(`${BASE}/${rateId}/charges/reorder`, { charge_ids: chargeIds }),

  // ── Task 19.3: Carrier Rate Contracts & Versions ────────────────────────────

  /**
   * List carrier rate contracts with filters and pagination.
   * @param {Object} params - { search, carrier_name, contract_type, status, renewal_status, transport_mode, page, limit, sort_by, sort_order }
   */
  getRateContracts: (params = {}) => {
    const q = new URLSearchParams();
    if (params.search) q.set('search', params.search);
    if (params.carrier_name && params.carrier_name !== 'ALL') q.set('carrier_name', params.carrier_name);
    if (params.contract_type && params.contract_type !== 'ALL') q.set('contract_type', params.contract_type);
    if (params.status && params.status !== 'ALL') q.set('status', params.status);
    if (params.renewal_status && params.renewal_status !== 'ALL') q.set('renewal_status', params.renewal_status);
    if (params.transport_mode && params.transport_mode !== 'ALL') q.set('transport_mode', params.transport_mode);
    if (params.page) q.set('page', params.page);
    if (params.limit) q.set('limit', params.limit);
    if (params.sort_by) q.set('sort_by', params.sort_by);
    if (params.sort_order) q.set('sort_order', params.sort_order);

    return api.get(`${BASE}/contracts?${q.toString()}`);
  },

  /**
   * Get KPI summary of contracts & expiring list.
   */
  getRateContractSummary: () => api.get(`${BASE}/contracts/summary`),

  /**
   * Create a new carrier rate contract.
   * @param {Object} payload - CreateRateContractRequest
   */
  createRateContract: (payload) => api.post(`${BASE}/contracts`, payload),

  /**
   * Get single contract with linked rates.
   * @param {number|string} id
   */
  getRateContract: (id) => api.get(`${BASE}/contracts/${id}`),

  /**
   * Update an existing carrier contract.
   * @param {number|string} id
   * @param {Object} payload - UpdateRateContractRequest
   */
  updateRateContract: (id, payload) => api.put(`${BASE}/contracts/${id}`, payload),

  /**
   * Archive a carrier contract.
   * @param {number|string} id
   */
  archiveRateContract: (id) => api.post(`${BASE}/contracts/${id}/archive`, {}),

  /**
   * Renew a carrier rate contract with new expiry date.
   * @param {number|string} id
   * @param {Object} payload - { new_expiry_date, renewal_status, notes }
   */
  renewRateContract: (id, payload) => api.post(`${BASE}/contracts/${id}/renew`, payload),

  /**
   * Create a new version of a rate (preserving historical records).
   * @param {number|string} rateId
   * @param {Object} payload - CreateRateVersionRequest
   */
  createRateVersion: (rateId, payload) => api.post(`${BASE}/${rateId}/versions`, payload),

  /**
   * Get version lineage chain for a rate.
   * @param {number|string} rateId
   */
  getRateVersions: (rateId) => api.get(`${BASE}/${rateId}/versions`),

  /**
   * Get audit history for a rate and its versions.
   * @param {number|string} rateId
   */
  getRateVersionHistory: (rateId) => api.get(`${BASE}/${rateId}/version-history`),

  // ── Task 19.4: Spot Rate Requests, Responses & Comparison ───────────────────

  /**
   * List spot rate requests with filtering & pagination.
   * @param {Object} params - { search, status, transport_mode, origin, destination, page, limit, sort_by, sort_order }
   */
  getSpotRateRequests: (params = {}) => {
    const q = new URLSearchParams();
    if (params.search) q.set('search', params.search);
    if (params.status && params.status !== 'ALL') q.set('status', params.status);
    if (params.transport_mode && params.transport_mode !== 'ALL') q.set('transport_mode', params.transport_mode);
    if (params.origin) q.set('origin', params.origin);
    if (params.destination) q.set('destination', params.destination);
    if (params.page) q.set('page', params.page);
    if (params.limit) q.set('limit', params.limit);
    if (params.sort_by) q.set('sort_by', params.sort_by);
    if (params.sort_order) q.set('sort_order', params.sort_order);

    return api.get(`${BASE}/spot-requests?${q.toString()}`);
  },

  /**
   * Get spot rate KPI summary.
   */
  getSpotRateSummary: () => api.get(`${BASE}/spot-requests/summary`),

  /**
   * Create a new spot rate sourcing request.
   * @param {Object} payload - CreateSpotRateRequestRequest
   */
  createSpotRateRequest: (payload) => api.post(`${BASE}/spot-requests`, payload),

  /**
   * Get single spot rate request by ID.
   * @param {number|string} id
   */
  getSpotRateRequest: (id) => api.get(`${BASE}/spot-requests/${id}`),

  /**
   * Update spot rate request.
   * @param {number|string} id
   * @param {Object} payload
   */
  updateSpotRateRequest: (id, payload) => api.put(`${BASE}/spot-requests/${id}`, payload),

  /**
   * Mark spot rate request as sent.
   * @param {number|string} id
   */
  sendSpotRateRequest: (id) => api.post(`${BASE}/spot-requests/${id}/send`, {}),

  /**
   * Cancel spot rate request.
   * @param {number|string} id
   */
  cancelSpotRateRequest: (id) => api.post(`${BASE}/spot-requests/${id}/cancel`, {}),

  /**
   * Get all carrier responses for a spot request.
   * @param {number|string} requestId
   */
  getSpotRateResponses: (requestId) => api.get(`${BASE}/spot-requests/${requestId}/responses`),

  /**
   * Log a new carrier response for a spot request.
   * @param {number|string} requestId
   * @param {Object} payload - CreateSpotRateResponseRequest
   */
  createSpotRateResponse: (requestId, payload) => api.post(`${BASE}/spot-requests/${requestId}/responses`, payload),

  /**
   * Get single carrier response.
   * @param {number|string} requestId
   * @param {number|string} responseId
   */
  getSpotRateResponse: (requestId, responseId) => api.get(`${BASE}/spot-requests/${requestId}/responses/${responseId}`),

  /**
   * Update carrier response & itemized charges.
   * @param {number|string} requestId
   * @param {number|string} responseId
   * @param {Object} payload
   */
  updateSpotRateResponse: (requestId, responseId, payload) => api.put(`${BASE}/spot-requests/${requestId}/responses/${responseId}`, payload),

  /**
   * Select a carrier response as preferred without mutating competing quotes.
   * @param {number|string} requestId
   * @param {number|string} responseId
   * @param {Object} payload - { selection_notes }
   */
  selectPreferredSpotRate: (requestId, responseId, payload = {}) => api.post(`${BASE}/spot-requests/${requestId}/responses/${responseId}/select`, payload),

  /**
   * Get side-by-side comparison analytics with cheapest, fastest, and best value indicators.
   * @param {number|string} requestId
   */
  compareSpotRates: (requestId) => api.get(`${BASE}/spot-requests/${requestId}/comparison`),

  // ── Task 19.6: Rate Lifecycle Intelligence & Expiry Monitoring ───────────────

  /**
   * Get global lifecycle KPI summary for rates & contracts.
   */
  getRateLifecycleSummary: () => api.get(`${BASE}/lifecycle/summary`),

  /**
   * Get chronological rate & contract lifecycle transition events.
   * @param {number} limit
   */
  getRateLifecycleEvents: (limit = 50) => api.get(`${BASE}/lifecycle/events?limit=${limit}`),

  /**
   * Get prioritized bucket list of rates requiring commercial attention.
   */
  getRatesRequiringAttention: () => api.get(`${BASE}/lifecycle/attention`),

  /**
   * Get prioritized list of contracts requiring renewal attention.
   */
  getContractsRequiringAttention: () => api.get(`${BASE}/contracts/attention`),

  /**
   * Trigger on-demand rate lifecycle evaluation across the organization.
   */
  evaluateRateLifecycle: () => api.post(`${BASE}/lifecycle/evaluate`, {}),

  // ── Task 5: Live Carrier Rates Integration ──────────────────────────────────

  /**
   * Search live rates across connected carrier APIs with normalized comparison.
   * @param {Object} payload - { origin_port, destination_port, equipment_type, carrier_scac, rate_type }
   */
  searchCarrierLiveRates: (payload) => api.post(`${BASE}/carrier-search`, payload),
};

export default rateService;

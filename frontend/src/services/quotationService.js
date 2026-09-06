import api from './api';

const BASE = '/api/v1/quotations';

const quotationService = {
  /**
   * Get KPI summary counts for the quotation strip.
   * @returns {Promise<QuotationSummary>}
   */
  getSummary: () => api.get(`${BASE}/summary`),

  /**
   * List quotations with filters and pagination.
   * @param {Object} params - { search, status, customer_id, validity, page, limit }
   */
  listQuotations: (params = {}) => {
    const query = new URLSearchParams();
    if (params.search)      query.set('search', params.search);
    if (params.status && params.status !== 'ALL') query.set('status', params.status);
    if (params.customer_id) query.set('customer_id', params.customer_id);
    if (params.validity && params.validity !== 'ALL') query.set('validity', params.validity);
    if (params.page)        query.set('page', params.page);
    if (params.limit)       query.set('limit', params.limit);
    return api.get(`${BASE}/?${query.toString()}`);
  },

  /**
   * Get a single quotation's full detail (includes customer info, activity, commercial terms, and pricing).
   * @param {number|string} id - Quotation DB id
   */
  getQuotation: (id) => api.get(`${BASE}/${id}`),

  /**
   * Create a new quotation in DRAFT status.
   * @param {Object} body - CreateQuotationRequest
   */
  createQuotation: (body) => api.post(`${BASE}/`, body),

  /**
   * Update a DRAFT quotation.
   * @param {number|string} id
   * @param {Object} body - UpdateQuotationRequest
   */
  updateQuotation: (id, body) => api.put(`${BASE}/${id}`, body),

  // ── Pricing & Charges (Task 18.2) ──────────────────────────────────────────

  /**
   * Get authoritative calculated pricing and charge line items for a quotation.
   * @param {number|string} quotationId
   */
  getQuotationPricing: (quotationId) => api.get(`${BASE}/${quotationId}/pricing`),

  /**
   * Add a line item charge to a DRAFT quotation.
   * @param {number|string} quotationId
   * @param {Object} payload - CreateQuotationChargeRequest
   */
  addQuotationCharge: (quotationId, payload) => api.post(`${BASE}/${quotationId}/charges`, payload),

  /**
   * Update a line item charge on a DRAFT quotation.
   * @param {number|string} quotationId
   * @param {number|string} chargeId
   * @param {Object} payload - UpdateQuotationChargeRequest
   */
  updateQuotationCharge: (quotationId, chargeId, payload) => api.put(`${BASE}/${quotationId}/charges/${chargeId}`, payload),

  /**
   * Delete a line item charge from a DRAFT quotation.
   * @param {number|string} quotationId
   * @param {number|string} chargeId
   */
  deleteQuotationCharge: (quotationId, chargeId) => api.delete(`${BASE}/${quotationId}/charges/${chargeId}`),

  /**
   * Reorder charge line items.
   * @param {number|string} quotationId
   * @param {Object} payload - { charge_ids: number[] }
   */
  reorderQuotationCharges: (quotationId, payload) => api.post(`${BASE}/${quotationId}/charges/reorder`, payload),

  /**
   * Search and retrieve matching Rate Management candidates for this quotation route.
   * @param {number|string} quotationId
   */
  getQuotationRateCandidates: (quotationId) => api.get(`${BASE}/${quotationId}/rate-candidates`),

  /**
   * Import rate line items into the quotation as persisted snapshots.
   * @param {number|string} quotationId
   * @param {Object} payload - { rate_id: string, charge_categories?: string[] }
   */
  importRateCharges: (quotationId, payload) => api.post(`${BASE}/${quotationId}/charges/import-rate`, payload),

  // ── Templates & Commercial Terms (Task 18.3) ───────────────────────────────

  /**
   * List reusable quotation templates.
   * @param {Object} params - { active_only: boolean }
   */
  listTemplates: (params = {}) => {
    const query = new URLSearchParams();
    if (params.active_only !== undefined) query.set('active_only', params.active_only);
    return api.get(`${BASE}/templates?${query.toString()}`);
  },

  /**
   * Get template detail with charge definitions.
   * @param {number|string} templateId
   */
  getTemplate: (templateId) => api.get(`${BASE}/templates/${templateId}`),

  /**
   * Create a new reusable quotation template.
   * @param {Object} payload - CreateQuotationTemplateRequest
   */
  createTemplate: (payload) => api.post(`${BASE}/templates`, payload),

  /**
   * Update an existing reusable template.
   * @param {number|string} templateId
   * @param {Object} payload - UpdateQuotationTemplateRequest
   */
  updateTemplate: (templateId, payload) => api.put(`${BASE}/templates/${templateId}`, payload),

  /**
   * Archive / soft-delete a template.
   * @param {number|string} templateId
   */
  deleteTemplate: (templateId) => api.delete(`${BASE}/templates/${templateId}`),

  /**
   * Apply a reusable template snapshot to a DRAFT quotation.
   * @param {number|string} quotationId
   * @param {Object} payload - { template_id: number, override_charges: boolean }
   */
  applyTemplate: (quotationId, payload) => api.post(`${BASE}/${quotationId}/apply-template`, payload),

  /**
   * Save an existing quotation's charges, terms, and notes as a new reusable template.
   * @param {number|string} quotationId
   * @param {Object} payload - { name: string, description?: string, validity_days?: number }
   */
  saveAsTemplate: (quotationId, payload) => api.post(`${BASE}/${quotationId}/save-as-template`, payload),

  /**
   * Update commercial terms, payment terms, and notes for a quotation.
   * @param {number|string} quotationId
   * @param {Object} payload - UpdateQuotationCommercialTermsRequest
   */
  updateCommercialTerms: (quotationId, payload) => api.put(`${BASE}/${quotationId}/commercial-terms`, payload),

  // ── Lifecycle & Approvals (Task 18.4) ───────────────────────────────────────

  /**
   * Submit quotation for management review (Draft / Changes Requested -> Ready For Review).
   * @param {number|string} quotationId
   * @param {Object} payload - { comments?: string }
   */
  submitForReview: (quotationId, payload = {}) => api.post(`${BASE}/${quotationId}/submit-review`, payload),

  /**
   * Approve quotation for sending (Ready For Review -> Approved).
   * @param {number|string} quotationId
   * @param {Object} payload - { approval_notes?: string }
   */
  approveQuotation: (quotationId, payload = {}) => api.post(`${BASE}/${quotationId}/approve`, payload),

  /**
   * Request changes on a quotation (Ready For Review -> Changes Requested).
   * @param {number|string} quotationId
   * @param {Object} payload - { reason: string }
   */
  requestChanges: (quotationId, payload) => api.post(`${BASE}/${quotationId}/request-changes`, payload),

  /**
   * Mark quotation as sent to customer (Approved / Draft -> Sent).
   * @param {number|string} quotationId
   * @param {Object} payload - { recipient_email?: string, comments?: string, send_copy?: boolean }
   */
  sendQuotation: (quotationId, payload = {}) => api.post(`${BASE}/${quotationId}/send`, payload),

  /**
   * Retrieve clean, customer-safe quotation preview model.
   * Internal costs, margins, and private notes are omitted.
   * @param {number|string} quotationId
   */
  getCustomerPreview: (quotationId) => api.get(`${BASE}/${quotationId}/customer-preview`),

  /**
   * Retrieve approval and lifecycle transition history.
   * @param {number|string} quotationId
   */
  getApprovalHistory: (quotationId) => api.get(`${BASE}/${quotationId}/approval-history`),

  /**
   * Retrieve current quotation approval and lifecycle capability status.
   * @param {number|string} quotationId
   */
  getApprovalStatus: (quotationId) => api.get(`${BASE}/${quotationId}/approval-status`),

  /**
   * Record a customer view event on the quotation.
   * @param {number|string} quotationId
   * @param {Object} payload - { viewer_name?: string, viewer_email?: string }
   */
  markViewed: (quotationId, payload = {}) => api.post(`${BASE}/${quotationId}/view`, payload),

  /**
   * Record customer acceptance of quotation.
   * @param {number|string} quotationId
   * @param {Object} payload - { accepted_by?: string, comments?: string }
   */
  acceptQuotation: (quotationId, payload = {}) => api.post(`${BASE}/${quotationId}/accept`, payload),

  /**
   * Record customer decline of quotation.
   * @param {number|string} quotationId
   * @param {Object} payload - { declined_by?: string, reason?: string }
   */
  declineQuotation: (quotationId, payload = {}) => api.post(`${BASE}/${quotationId}/decline`, payload),

  /**
   * Cancel quotation.
   * @param {number|string} quotationId
   * @param {Object} payload - { reason?: string }
   */
  cancelQuotation: (quotationId, payload = {}) => api.post(`${BASE}/${quotationId}/cancel`, payload),

  // ── Quotation-to-Booking Operational Conversion (Task 18.6) ────────────────

  /**
   * Get operational conversion preview and readiness status.
   * @param {number|string} quotationId
   */
  getConversionPreview: (quotationId) => api.get(`${BASE}/${quotationId}/conversion-preview`),

  /**
   * Convert an accepted quotation into an operational booking.
   * @param {number|string} quotationId
   * @param {Object} payload - ConvertQuotationToBookingRequest
   */
  convertToBooking: (quotationId, payload = {}) => api.post(`${BASE}/${quotationId}/convert-to-booking`, payload),

  /**
   * Get commercial handover and conversion audit history.
   * @param {number|string} quotationId
   */
  getConversionHistory: (quotationId) => api.get(`${BASE}/${quotationId}/conversion-history`),

  // ── Booking Confirmation & Handover Traceability (Task 18.7) ───────────────

  /**
   * Get end-to-end quotation operational handover and lineage details.
   * @param {number|string} quotationId
   */
  getOperationalHandover: (quotationId) => api.get(`${BASE}/${quotationId}/operational-handover`),

  /**
   * Confirm booking handover to operations.
   * @param {number|string} quotationId
   * @param {Object} payload - { confirmation_notes?: string, notify_operations?: boolean }
   */
  confirmOperationalHandover: (quotationId, payload = {}) => api.post(`${BASE}/${quotationId}/confirm-handover`, payload),

  // ── Quotation Analytics, Performance & Intelligence (Task 18.8) ────────────

  /**
   * Get executive management analytics overview KPIs and deterministic operational insights.
   */
  getQuotationAnalyticsOverview: () => api.get(`${BASE}/analytics/overview`),

  /**
   * Get quotation volume, conversion, and margin trends over time.
   * @param {Object} params - { days?: number }
   */
  getQuotationAnalyticsTrends: (params = {}) => {
    const query = new URLSearchParams();
    if (params.days) query.set('days', params.days);
    return api.get(`${BASE}/analytics/trends?${query.toString()}`);
  },

  /**
   * Get aggregated quotation commercial performance per customer.
   */
  getCustomerQuotationPerformance: () => api.get(`${BASE}/analytics/customers`),

  /**
   * Get quotation metrics broken down by transport/service mode.
   */
  getQuotationPerformanceByMode: () => api.get(`${BASE}/analytics/modes`),

  /**
   * Get quotations with expiry or review risk items requiring attention.
   */
  getQuotationExpiryRisk: () => api.get(`${BASE}/analytics/expiry-risk`),

  // ── Task 19.5: Rate-to-Quotation Integration & Commercial Selection ─────────

  /**
   * Get available matching managed carrier rates & spot responses for a quotation.
   * @param {number|string} quotationId
   */
  getRateCandidates: (quotationId) => api.get(`${BASE}/${quotationId}/rate-candidates`),

  /**
   * Get current active rate selection for a quotation.
   * @param {number|string} quotationId
   */
  getRateSelection: (quotationId) => api.get(`${BASE}/${quotationId}/rate-selection`),

  /**
   * Select a managed rate or spot rate response for a quotation.
   * @param {number|string} quotationId
   * @param {Object} payload - { source_type: 'MANAGED_RATE'|'SPOT_RATE', rate_id?: number, spot_rate_response_id?: number, notes?: string }
   */
  selectRate: (quotationId, payload) => api.post(`${BASE}/${quotationId}/rate-selection`, payload),

  /**
   * Replace active rate selection with a new rate candidate.
   * @param {number|string} quotationId
   * @param {Object} payload
   */
  replaceRate: (quotationId, payload) => api.put(`${BASE}/${quotationId}/rate-selection`, payload),

  /**
   * Remove rate selection from a quotation.
   * @param {number|string} quotationId
   */
  removeRate: (quotationId) => api.delete(`${BASE}/${quotationId}/rate-selection`),

  /**
   * Get immutable commercial rate snapshot for a quotation.
   * @param {number|string} quotationId
   */
  getRateSnapshot: (quotationId) => api.get(`${BASE}/${quotationId}/rate-snapshot`),

  /**
   * Get audit timeline of rate selections for a quotation.
   * @param {number|string} quotationId
   */
  getRateSelectionHistory: (quotationId) => api.get(`${BASE}/${quotationId}/rate-selection-history`),

  // ── Task 19.6: Quotation Rate Risk & Commercial Impact Analysis ──────────────

  /**
   * Get commercial risks impacting a quotation's selected rate snapshot.
   * @param {number|string} quotationId
   */
  getRateRisks: (quotationId) => api.get(`${BASE}/${quotationId}/rate-risks`),

  /**
   * Mark a specific quotation rate risk as resolved.
   * @param {number|string} quotationId
   * @param {number|string} riskId
   */
  resolveRateRisk: (quotationId, riskId) => api.post(`${BASE}/${quotationId}/rate-risks/${riskId}/resolve`, {}),

  /**
   * Get replacement carrier rate candidates for an at-risk quotation.
   * @param {number|string} quotationId
   */
  getRateReplacementCandidates: (quotationId) => api.get(`${BASE}/${quotationId}/replacement-rate-candidates`),

  /**
   * Calculate side-by-side commercial impact comparing current snapshot vs replacement option.
   * @param {number|string} quotationId
   * @param {Object} params - { replacement_rate_id?: number, replacement_spot_id?: number }
   */
  getCommercialImpactAnalysis: (quotationId, params = {}) => {
    const q = new URLSearchParams();
    if (params.replacement_rate_id) q.set('replacement_rate_id', params.replacement_rate_id);
    if (params.replacement_spot_id) q.set('replacement_spot_id', params.replacement_spot_id);
    return api.get(`${BASE}/${quotationId}/commercial-impact?${q.toString()}`);
  },

  /**
   * Trigger on-demand rate risk evaluation across all quotations.
   */
  evaluateRateRisks: () => api.post(`${BASE}/rate-risks/evaluate`, {}),
};

export default quotationService;






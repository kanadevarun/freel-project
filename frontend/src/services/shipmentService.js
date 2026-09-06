import api from './api';

/**
 * Service for interacting with the Dedicated Shipment Operations Workspace API.
 * All data derives from real persisted database records.
 */
export const shipmentService = {
  /**
   * Fetch paginated shipments workspace with search, filter, and live KPIs.
   * @param {Object} params - { page, limit, status, search }
   */
  getShipments: async (params = {}) => {
    const query = new URLSearchParams();
    if (params.page) query.append('page', params.page);
    if (params.limit) query.append('limit', params.limit);
    if (params.status && params.status !== 'ALL') query.append('status', params.status);
    if (params.search) query.append('search', params.search);
    query.append('workspace', 'true'); // Flag to trigger workspace response on backend

    const queryString = query.toString();
    const url = `/api/v1/shipments${queryString ? `?${queryString}` : ''}`;
    return api.get(url);
  },

  /**
   * Fetch full workspace detail for a specific shipment.
   */
  getShipmentDetail: async (shipmentId) => {
    return api.get(`/api/v1/shipments/${shipmentId}`);
  },

  /**
   * Manually update/complete a shipment milestone.
   */
  updateMilestone: async (shipmentId, data) => {
    return api.put(`/api/v1/shipments/${shipmentId}/milestones`, data);
  },

  /**
   * Fetch all operational exceptions for a shipment.
   */
  getShipmentExceptions: async (shipmentId) => {
    return api.get(`/api/v1/shipments/${shipmentId}/exceptions`);
  },

  /**
   * Manually create a shipment exception.
   */
  createShipmentException: async (shipmentId, data) => {
    return api.post(`/api/v1/shipments/${shipmentId}/exceptions`, data);
  },

  /**
   * Acknowledge a shipment exception.
   */
  acknowledgeException: async (shipmentId, exceptionId) => {
    return api.post(`/api/v1/shipments/${shipmentId}/exceptions/${exceptionId}/acknowledge`);
  },

  /**
   * Resolve a shipment exception.
   */
  resolveException: async (shipmentId, exceptionId, resolutionNotes) => {
    return api.post(`/api/v1/shipments/${shipmentId}/exceptions/${exceptionId}/resolve`, {
      resolution_notes: resolutionNotes,
    });
  },

  /**
   * Dismiss a shipment exception.
   */
  dismissException: async (shipmentId, exceptionId) => {
    return api.post(`/api/v1/shipments/${shipmentId}/exceptions/${exceptionId}/dismiss`);
  },

  /**
   * Evaluate shipment exceptions deterministically.
   */
  evaluateExceptions: async (shipmentId) => {
    return api.post(`/api/v1/shipments/${shipmentId}/exceptions/evaluate`);
  },

  /**
   * Fetch live tracking intelligence summary.
   */
  getShipmentTracking: async (shipmentId) => {
    return api.get(`/api/v1/shipments/${shipmentId}/tracking`);
  },

  /**
   * Fetch latest geographic position and vessel telemetry.
   */
  getLatestPosition: async (shipmentId) => {
    return api.get(`/api/v1/shipments/${shipmentId}/tracking/position`);
  },

  /**
   * Fetch historical tracking positions with limit.
   */
  getPositionHistory: async (shipmentId, params = {}) => {
    const query = new URLSearchParams();
    if (params.limit) query.append('limit', params.limit);
    const qs = query.toString();
    return api.get(`/api/v1/shipments/${shipmentId}/tracking/positions${qs ? `?${qs}` : ''}`);
  },

  /**
   * Fetch planned route coordinates and waypoints.
   */
  getTrackingRoute: async (shipmentId) => {
    return api.get(`/api/v1/shipments/${shipmentId}/tracking/route`);
  },

  /**
   * Fetch normalized chronological tracking events feed.
   */
  getTrackingEvents: async (shipmentId) => {
    return api.get(`/api/v1/shipments/${shipmentId}/tracking/events`);
  },

  /**
   * Fetch unified tracking intelligence, journey progression, schedule variance, and alerts (Task 17.4).
   */
  getTrackingIntelligence: async (shipmentId) => {
    return api.get(`/api/v1/shipments/${shipmentId}/tracking/intelligence`);
  },

  /**
   * Fetch persistent tracking alerts with optional status filter (Task 17.5).
   */
  getTrackingAlerts: async (shipmentId, params = {}) => {
    const query = new URLSearchParams();
    if (params.status) query.append('status', params.status);
    const qs = query.toString();
    return api.get(`/api/v1/shipments/${shipmentId}/tracking/alerts${qs ? `?${qs}` : ''}`);
  },

  /**
   * Fetch operational monitoring summary and next recommended action (Task 17.5).
   */
  getTrackingMonitoring: async (shipmentId) => {
    return api.get(`/api/v1/shipments/${shipmentId}/tracking/monitoring`);
  },

  /**
   * Orchestrate provider telemetry refresh and recalculate intelligence (Task 17.5 & 17.6).
   */
  refreshShipmentTracking: async (shipmentId) => {
    return api.post(`/api/v1/shipments/${shipmentId}/tracking/refresh`);
  },

  /**
   * Fetch recent tracking refresh history runs (Task 17.7).
   */
  getTrackingRefreshHistory: async (shipmentId, params = {}) => {
    const query = new URLSearchParams();
    if (params.limit) query.append('limit', params.limit);
    const qs = query.toString();
    return api.get(`/api/v1/shipments/${shipmentId}/tracking/refresh-history${qs ? `?${qs}` : ''}`);
  },

  /**
   * Acknowledge an open tracking alert (Task 17.5).
   */
  acknowledgeTrackingAlert: async (shipmentId, alertId, notes = '') => {
    return api.post(`/api/v1/shipments/${shipmentId}/tracking/alerts/${alertId}/acknowledge`, { notes });
  },

  /**
   * Resolve a tracking alert (Task 17.5).
   */
  resolveTrackingAlert: async (shipmentId, alertId, resolutionNotes = '') => {
    return api.post(`/api/v1/shipments/${shipmentId}/tracking/alerts/${alertId}/resolve`, { resolution_notes: resolutionNotes });
  },

  /**
   * Suppress a tracking alert with a reason (Task 17.5).
   */
  suppressTrackingAlert: async (shipmentId, alertId, reason = '') => {
    return api.post(`/api/v1/shipments/${shipmentId}/tracking/alerts/${alertId}/suppress`, { reason });
  },

  // ─── Tracking Analytics & Performance Intelligence (Task 17.8) ────────────────

  /**
   * Fetch fleet-level tracking KPIs, freshness, alert stats, and deterministic operational insights.
   */
  getTrackingAnalyticsOverview: async () => {
    return api.get('/api/v1/tracking/analytics/overview');
  },

  /**
   * Fetch historical performance trends (on-time rate, delays, alerts, refresh success).
   */
  getTrackingAnalyticsTrends: async (params = {}) => {
    const query = new URLSearchParams();
    if (params.days) query.append('days', params.days);
    const qs = query.toString();
    return api.get(`/api/v1/tracking/analytics/trends${qs ? `?${qs}` : ''}`);
  },

  /**
   * Fetch aggregated carrier-level tracking reliability and delay scorecards.
   */
  getCarrierTrackingPerformance: async () => {
    return api.get('/api/v1/tracking/analytics/carriers');
  },

  /**
   * Fetch route corridor performance and transit variance metrics.
   */
  getRouteTrackingPerformance: async () => {
    return api.get('/api/v1/tracking/analytics/routes');
  },

  /**
   * Fetch shipment closure evaluation.
   */
  getShipmentClosure: async (shipmentId) => {
    return api.get(`/api/v1/shipments/${shipmentId}/closure`);
  },

  /**
   * Force closure state re-evaluation.
   */
  evaluateClosure: async (shipmentId) => {
    return api.post(`/api/v1/shipments/${shipmentId}/evaluate`);
  },

  /**
   * Request operational shipment closure.
   */
  requestClosure: async (shipmentId) => {
    return api.post(`/api/v1/shipments/${shipmentId}/request-closure`);
  },

  /**
   * Complete and close the shipment.
   */
  completeShipment: async (shipmentId) => {
    return api.post(`/api/v1/shipments/${shipmentId}/complete`);
  },

  /**
   * Reopen a closed shipment.
   */
  reopenShipment: async (shipmentId) => {
    return api.post(`/api/v1/shipments/${shipmentId}/reopen`);
  },

  // ─── Shipment Document & Compliance API (Task 16.7) ─────────────────────────
  /**
   * Fetch all shipment documents with compliance summary and discrepancies.
   */
  getShipmentDocuments: async (shipmentId, params = {}) => {
    const query = new URLSearchParams();
    if (params.category && params.category !== 'ALL') query.append('category', params.category);
    if (params.status && params.status !== 'ALL') query.append('status', params.status);
    const qs = query.toString();
    return api.get(`/api/v1/shipments/${shipmentId}/documents${qs ? `?${qs}` : ''}`);
  },

  /**
   * Upload or attach a new document to a shipment.
   */
  uploadShipmentDocument: async (shipmentId, data) => {
    return api.post(`/api/v1/shipments/${shipmentId}/documents`, data);
  },

  /**
   * Update document metadata or status.
   */
  updateShipmentDocument: async (shipmentId, documentId, data) => {
    return api.patch(`/api/v1/shipments/${shipmentId}/documents/${documentId}`, data);
  },

  /**
   * Approve a shipment document.
   */
  approveShipmentDocument: async (shipmentId, documentId) => {
    return api.post(`/api/v1/shipments/${shipmentId}/documents/${documentId}/approve`);
  },

  /**
   * Reject a shipment document with reason.
   */
  rejectShipmentDocument: async (shipmentId, documentId, rejectionReason) => {
    return api.post(`/api/v1/shipments/${shipmentId}/documents/${documentId}/reject`, {
      rejection_reason: rejectionReason,
    });
  },

  /**
   * Delete a shipment document.
   */
  deleteShipmentDocument: async (shipmentId, documentId) => {
    return api.delete(`/api/v1/shipments/${shipmentId}/documents/${documentId}`);
  },

  // ─── Shipment Financial Operations API (Task 16.8) ──────────────────────────
  /**
   * Fetch aggregated financial summary for a shipment.
   */
  getShipmentFinancials: async (shipmentId) => {
    return api.get(`/api/v1/shipments/${shipmentId}/financials`);
  },

  /**
   * Fetch line-item cost and revenue charges for a shipment.
   */
  getShipmentCharges: async (shipmentId, params = {}) => {
    const query = new URLSearchParams();
    if (params.category && params.category !== 'ALL') query.append('category', params.category);
    if (params.charge_type && params.charge_type !== 'ALL') query.append('charge_type', params.charge_type);
    const qs = query.toString();
    return api.get(`/api/v1/shipments/${shipmentId}/financials/charges${qs ? `?${qs}` : ''}`);
  },

  /**
   * Create an operational financial charge line item.
   */
  createShipmentCharge: async (shipmentId, data) => {
    return api.post(`/api/v1/shipments/${shipmentId}/financials/charges`, data);
  },

  /**
   * Update an existing charge line item.
   */
  updateShipmentCharge: async (shipmentId, chargeId, data) => {
    return api.patch(`/api/v1/shipments/${shipmentId}/financials/charges/${chargeId}`, data);
  },

  /**
   * Delete a charge line item.
   */
  deleteShipmentCharge: async (shipmentId, chargeId) => {
    return api.delete(`/api/v1/shipments/${shipmentId}/financials/charges/${chargeId}`);
  },

  /**
   * Recalculate margins and variances in real-time.
   */
  recalculateShipmentFinancials: async (shipmentId) => {
    return api.post(`/api/v1/shipments/${shipmentId}/financials/recalculate`);
  },

  /**
   * Record financial review or close shipment financially.
   */
  reviewShipmentFinancials: async (shipmentId, data) => {
    return api.post(`/api/v1/shipments/${shipmentId}/financials/review`, data);
  },
};

export default shipmentService;



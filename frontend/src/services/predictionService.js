import api from './api';

const unwrap = (res) => (res && res.data !== undefined ? res.data : res);

export const predictionService = {
  /**
   * List predictions for current organization with optional filters
   * @param {Object} params - { module, severity, status, related_record_type, related_record_id, limit, offset }
   */
  listPredictions: async (params = {}) => {
    const query = new URLSearchParams();
    Object.entries(params).forEach(([key, val]) => {
      if (val !== undefined && val !== null && val !== '') {
        query.append(key, val);
      }
    });
    const res = await api.get(`/api/v1/predictions?${query.toString()}`);
    return unwrap(res);
  },

  /**
   * Get single prediction by ID
   */
  getPrediction: async (id) => {
    const res = await api.get(`/api/v1/predictions/${id}`);
    return unwrap(res);
  },

  /**
   * Get or predict shipment ETA and delay intelligence
   */
  getShipmentPredictedETA: async (shipmentId, forceRefresh = false) => {
    const url = `/api/v1/shipments/${shipmentId}/predicted-eta${forceRefresh ? '?refresh=true' : ''}`;
    const res = await api.get(url);
    return unwrap(res);
  },

  /**
   * Refresh shipment predicted ETA
   */
  refreshShipmentPredictedETA: async (shipmentId) => {
    const res = await api.post(`/api/v1/shipments/${shipmentId}/predicted-eta/refresh`);
    return unwrap(res);
  },

  /**
   * Get or predict shipment exception risk and disruption forecast
   */
  getShipmentPredictedExceptions: async (shipmentId, forceRefresh = false) => {
    const url = `/api/v1/shipments/${shipmentId}/predicted-exceptions${forceRefresh ? '?refresh=true' : ''}`;
    const res = await api.get(url);
    return unwrap(res);
  },

  /**
   * Refresh shipment exception risk and disruption forecast
   */
  refreshShipmentPredictedExceptions: async (shipmentId) => {
    const res = await api.post(`/api/v1/shipments/${shipmentId}/predicted-exceptions/refresh`);
    return unwrap(res);
  },

  /**
   * Get or predict lead intelligence (conversion likelihood, inactivity risk)
   */
  getLeadPredictedIntelligence: async (leadId, forceRefresh = false) => {
    const url = `/api/v1/leads/${leadId}/predicted-intelligence${forceRefresh ? '?refresh=true' : ''}`;
    const res = await api.get(url);
    return unwrap(res);
  },

  /**
   * Refresh lead predicted intelligence
   */
  refreshLeadPredictedIntelligence: async (leadId) => {
    const res = await api.post(`/api/v1/leads/${leadId}/predicted-intelligence/refresh`);
    return unwrap(res);
  },

  /**
   * Get or predict customer intelligence (repeat business, churn/engagement risk)
   */
  getCustomerPredictedIntelligence: async (customerId, forceRefresh = false) => {
    const url = `/api/v1/customers/${customerId}/predicted-intelligence${forceRefresh ? '?refresh=true' : ''}`;
    const res = await api.get(url);
    return unwrap(res);
  },

  /**
   * Refresh customer predicted intelligence
   */
  refreshCustomerPredictedIntelligence: async (customerId) => {
    const res = await api.post(`/api/v1/customers/${customerId}/predicted-intelligence/refresh`);
    return unwrap(res);
  },

  /**
   * Get or predict RFQ margin risk and quotation competitiveness
   */
  getRFQPredictedMargin: async (rfqId, forceRefresh = false) => {
    const url = `/api/v1/rfqs/${rfqId}/predicted-margin${forceRefresh ? '?refresh=true' : ''}`;
    const res = await api.get(url);
    return unwrap(res);
  },

  /**
   * Refresh RFQ predicted margin intelligence
   */
  refreshRFQPredictedMargin: async (rfqId) => {
    const res = await api.post(`/api/v1/rfqs/${rfqId}/predicted-margin/refresh`);
    return unwrap(res);
  },

  /**
   * Get or predict contract rate pressure and tariff expiry risk
   */
  getContractPredictedRatePressure: async (contractId, forceRefresh = false) => {
    const url = `/api/v1/contracts/${contractId}/predicted-rate-pressure${forceRefresh ? '?refresh=true' : ''}`;
    const res = await api.get(url);
    return unwrap(res);
  },

  /**
   * Refresh contract predicted rate pressure
   */
  refreshContractPredictedRatePressure: async (contractId) => {
    const res = await api.post(`/api/v1/contracts/${contractId}/predicted-rate-pressure/refresh`);
    return unwrap(res);
  },

  /**
   * Get or predict contract compliance, clause discrepancies, and documentation completeness risk
   */
  getContractPredictedComplianceRisk: async (contractId, forceRefresh = false) => {
    const url = `/api/v1/contracts/${contractId}/predicted-compliance${forceRefresh ? '?refresh=true' : ''}`;
    const res = await api.get(url);
    return unwrap(res);
  },

  /**
   * Refresh contract compliance, clause discrepancies, and documentation completeness risk
   */
  refreshContractPredictedComplianceRisk: async (contractId) => {
    const res = await api.post(`/api/v1/contracts/${contractId}/predicted-compliance/refresh`);
    return unwrap(res);
  },

  /**
   * Get or predict invoice late payment risk, collection priority, and cash inflow forecast
   */
  getInvoicePredictedCollection: async (invoiceId, forceRefresh = false) => {
    const url = `/api/v1/invoices/${invoiceId}/predicted-collection${forceRefresh ? '?refresh=true' : ''}`;
    const res = await api.get(url);
    return unwrap(res);
  },

  /**
   * Refresh invoice predicted collection intelligence
   */
  refreshInvoicePredictedCollection: async (invoiceId) => {
    const res = await api.post(`/api/v1/invoices/${invoiceId}/predicted-collection/refresh`);
    return unwrap(res);
  },

  /**
   * Get or predict shipment documentation, readiness, and operational compliance risk
   */
  getShipmentPredictedReadiness: async (shipmentId, forceRefresh = false) => {
    const url = `/api/v1/shipments/${shipmentId}/predicted-readiness${forceRefresh ? '?refresh=true' : ''}`;
    const res = await api.get(url);
    return unwrap(res);
  },

  /**
   * Refresh shipment predicted readiness and compliance intelligence
   */
  refreshShipmentPredictedReadiness: async (shipmentId) => {
    const res = await api.post(`/api/v1/shipments/${shipmentId}/predicted-readiness/refresh`);
    return unwrap(res);
  },

  /**
   * Get or predict carrier performance, delay risk, and operational holds
   */
  getCarrierPredictedPerformance: async (scac, forceRefresh = false) => {
    const url = `/api/v1/carriers/${scac}/predicted-performance${forceRefresh ? '?refresh=true' : ''}`;
    const res = await api.get(url);
    return unwrap(res);
  },

  /**
   * Refresh carrier predicted performance intelligence
   */
  refreshCarrierPredictedPerformance: async (scac) => {
    const res = await api.post(`/api/v1/carriers/${scac}/predicted-performance/refresh`);
    return unwrap(res);
  },

  /**
   * Get or predict trade lane corridor performance, dwell times, and bottlenecks
   */
  getLanePredictedPerformance: async (laneCode, forceRefresh = false) => {
    const url = `/api/v1/network/lanes/${laneCode}/predicted-performance${forceRefresh ? '?refresh=true' : ''}`;
    const res = await api.get(url);
    return unwrap(res);
  },

  /**
   * Refresh trade lane predicted corridor performance
   */
  refreshLanePredictedPerformance: async (laneCode) => {
    const res = await api.post(`/api/v1/network/lanes/${laneCode}/predicted-performance/refresh`);
    return unwrap(res);
  },

  /**
   * Get or predict customer service risk, SLA health, and relationship trajectory
   */
  getCustomerPredictedServicePerformance: async (customerId, forceRefresh = false) => {
    const url = `/api/v1/customers/${customerId}/predicted-performance${forceRefresh ? '?refresh=true' : ''}`;
    const res = await api.get(url);
    return unwrap(res);
  },

  /**
   * Refresh customer predicted service performance intelligence
   */
  refreshCustomerPredictedServicePerformance: async (customerId) => {
    const res = await api.post(`/api/v1/customers/${customerId}/predicted-performance/refresh`);
    return unwrap(res);
  },

  /**
   * Trigger on-demand prediction generation for a record
   */
  generatePrediction: async (payload) => {
    const res = await api.post('/api/v1/predictions/generate', payload);
    return unwrap(res);
  },

  /**
   * Get or predict operational / approvals / documentation workload planning
   */
  getWorkloadPrediction: async (area = 'approvals', forceRefresh = false) => {
    const url = `/api/v1/workload/predicted-workload?area=${encodeURIComponent(area)}${forceRefresh ? '&refresh=true' : ''}`;
    const res = await api.get(url);
    return unwrap(res);
  },

  /**
   * Refresh workload planning prediction
   */
  refreshWorkloadPrediction: async (area = 'approvals') => {
    const res = await api.post(`/api/v1/workload/predicted-workload/refresh?area=${encodeURIComponent(area)}`);
    return unwrap(res);
  },

  /**
   * Get or predict trade corridor / carrier capacity planning
   */
  getCapacityPrediction: async (dimension = 'corridor', forceRefresh = false) => {
    const url = `/api/v1/capacity/predicted-capacity?dimension=${encodeURIComponent(dimension)}${forceRefresh ? '&refresh=true' : ''}`;
    const res = await api.get(url);
    return unwrap(res);
  },

  /**
   * Refresh capacity planning prediction
   */
  refreshCapacityPrediction: async (dimension = 'corridor') => {
    const res = await api.post(`/api/v1/capacity/predicted-capacity/refresh?dimension=${encodeURIComponent(dimension)}`);
    return unwrap(res);
  },

  /**
   * Get or predict commercial RFQ / quote demand planning
   */
  getDemandPrediction: async (segment = 'commercial', forceRefresh = false) => {
    const url = `/api/v1/demand/predicted-demand?segment=${encodeURIComponent(segment)}${forceRefresh ? '&refresh=true' : ''}`;
    const res = await api.get(url);
    return unwrap(res);
  },

  /**
   * Refresh demand planning prediction
   */
  refreshDemandPrediction: async (segment = 'commercial') => {
    const res = await api.post(`/api/v1/demand/predicted-demand/refresh?segment=${encodeURIComponent(segment)}`);
    return unwrap(res);
  },

  /**
   * Get unified workload, capacity, and demand planning summary
   */
  getWorkloadCapacitySummary: async () => {
    const res = await api.get('/api/v1/planning/workload-capacity-intelligence');
    return unwrap(res);
  },

  /**
   * Get or predict operational bottleneck intelligence (approvals, documentation, cross_module)
   */
  getOperationalBottleneck: async (type = 'approvals', forceRefresh = false) => {
    const url = `/api/v1/bottlenecks/predicted-bottlenecks?type=${encodeURIComponent(type)}${forceRefresh ? '&refresh=true' : ''}`;
    const res = await api.get(url);
    return unwrap(res);
  },

  /**
   * Refresh operational bottleneck prediction
   */
  refreshOperationalBottleneck: async (type = 'approvals') => {
    const res = await api.post(`/api/v1/bottlenecks/predicted-bottlenecks/refresh?type=${encodeURIComponent(type)}`);
    return unwrap(res);
  },

  /**
   * Get or predict resource allocation and owner workload imbalance
   */
  getResourceAllocation: async (dimension = 'owner_workload', forceRefresh = false) => {
    const url = `/api/v1/resources/predicted-allocation?dimension=${encodeURIComponent(dimension)}${forceRefresh ? '&refresh=true' : ''}`;
    const res = await api.get(url);
    return unwrap(res);
  },

  /**
   * Refresh resource allocation prediction
   */
  refreshResourceAllocation: async (dimension = 'owner_workload') => {
    const res = await api.post(`/api/v1/resources/predicted-allocation/refresh?dimension=${encodeURIComponent(dimension)}`);
    return unwrap(res);
  },

  /**
   * Get unified resource allocation and operational bottleneck summary
   */
  getResourceBottleneckSummary: async () => {
    const res = await api.get('/api/v1/planning/resource-bottleneck-intelligence');
    return unwrap(res);
  },

  /**
   * Acknowledge a prediction
   */
  acknowledgePrediction: async (id) => {
    const res = await api.post(`/api/v1/predictions/${id}/acknowledge`);
    return unwrap(res);
  },

  /**
   * Dismiss a prediction with reason
   */
  dismissPrediction: async (id, reason) => {
    const res = await api.post(`/api/v1/predictions/${id}/dismiss`, { reason });
    return unwrap(res);
  },

  /**
   * Request execution of recommended action (queues for HITL approval)
   */
  requestAction: async (id, notes) => {
    const res = await api.post(`/api/v1/predictions/${id}/request-action`, { notes });
    return unwrap(res);
  },

  /**
   * Record real-world outcome
   */
  recordOutcome: async (id, outcomeStatus, outcomeValue, feedbackNotes) => {
    const res = await api.post(`/api/v1/predictions/${id}/outcome`, {
      outcome_status: outcomeStatus,
      outcome_value: outcomeValue,
      feedback_notes: feedbackNotes,
    });
    return unwrap(res);
  },

  /**
   * Get prediction lifecycle audit history
   */
  getAuditHistory: async (id) => {
    const res = await api.get(`/api/v1/predictions/${id}/audit`);
    return unwrap(res);
  },
};

export default predictionService;

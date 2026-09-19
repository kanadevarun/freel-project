import api from './api';

export const monitoringService = {
  // Overall Health Summary
  getHealthSummary: async () => {
    const res = await api.get('/api/v1/monitoring/health');
    return res?.data || res;
  },

  // Runtime Executions
  listExecutions: async (params = {}) => {
    const res = await api.get('/api/v1/monitoring/executions', { params });
    return res?.data || res;
  },

  getExecution: async (id) => {
    const res = await api.get(`/api/v1/monitoring/executions/${id}`);
    return res?.data || res;
  },

  // Performance & Latency (P50, P95, P99)
  getPerformance: async (days = 7) => {
    const res = await api.get('/api/v1/monitoring/performance', { params: { days } });
    return res?.data || res;
  },

  // Cost Reporting & Token Consumption
  getCost: async (days = 30) => {
    const res = await api.get('/api/v1/monitoring/cost', { params: { days } });
    return res?.data || res;
  },

  // Quality & Grounding
  getQuality: async () => {
    const res = await api.get('/api/v1/monitoring/quality');
    return res?.data || res;
  },

  evaluateQuality: async (data) => {
    const res = await api.post('/api/v1/monitoring/evaluate', data);
    return res?.data || res;
  },

  // Queue Worker Observability
  getQueue: async () => {
    const res = await api.get('/api/v1/monitoring/queue');
    return res?.data || res;
  },

  // Recommendation Center Metrics
  getRecommendations: async () => {
    const res = await api.get('/api/v1/monitoring/recommendations');
    return res?.data || res;
  },

  // Action & Approval Metrics
  getApprovals: async () => {
    const res = await api.get('/api/v1/monitoring/approvals');
    return res?.data || res;
  },

  // AI Memory Safety Metrics
  getMemory: async () => {
    const res = await api.get('/api/v1/monitoring/memory');
    return res?.data || res;
  },

  // Safety & Security Events
  getSecurity: async () => {
    const res = await api.get('/api/v1/monitoring/security');
    return res?.data || res;
  },

  // Model Pricing Catalog (Admin)
  getPricing: async () => {
    const res = await api.get('/api/v1/monitoring/pricing');
    return res?.data || res;
  },

  updatePricing: async (data) => {
    const res = await api.put('/api/v1/monitoring/pricing', data);
    return res?.data || res;
  },

  // Operational Health Thresholds (Admin)
  getThresholds: async () => {
    const res = await api.get('/api/v1/monitoring/thresholds');
    return res?.data || res;
  },

  updateThreshold: async (data) => {
    const res = await api.put('/api/v1/monitoring/thresholds', data);
    return res?.data || res;
  },
};

export default monitoringService;

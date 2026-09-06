import api from './api';

export const carrierIntegrationsService = {
  getProviders: async () => {
    const response = await api.get('/api/v1/carrier-providers');
    return response.data || response;
  },

  getIntegrations: async () => {
    const response = await api.get('/api/v1/carrier-integrations');
    return response.data || response;
  },

  getIntegration: async (id) => {
    const response = await api.get(`/api/v1/carrier-integrations/${id}`);
    return response.data || response;
  },
  
  connectCarrier: async (data) => {
    const response = await api.post('/api/v1/carrier-integrations', data);
    return response.data || response;
  },
  
  updateCarrier: async (id, data) => {
    const response = await api.put(`/api/v1/carrier-integrations/${id}`, data);
    return response.data || response;
  },

  toggleCarrier: async (id, isActive) => {
    const response = await api.patch(`/api/v1/carrier-integrations/${id}/toggle`, { is_active: isActive });
    return response.data || response;
  },
  
  removeCarrier: async (id) => {
    const response = await api.delete(`/api/v1/carrier-integrations/${id}`);
    return response.data || response;
  },
  
  testConnection: async (id) => {
    const response = await api.post(`/api/v1/carrier-integrations/${id}/test`);
    return response.data || response;
  },

  testDirectConnection: async (data) => {
    const response = await api.post('/api/v1/carrier-integrations/test-direct', data);
    return response.data || response;
  },
  
  syncCarrier: async (id, data = {}) => {
    const response = await api.post(`/api/v1/carrier-integrations/${id}/sync`, data);
    return response.data || response;
  },

  syncNow: async (id, data = {}) => {
    const response = await api.post(`/api/v1/carrier-integrations/${id}/sync`, data);
    return response.data || response;
  },

  getSyncHistory: async (id, params = {}) => {
    const response = await api.get(`/api/v1/carrier-integrations/${id}/sync-history`, { params });
    return response.data || response;
  },

  getSyncJob: async (id, syncId) => {
    const response = await api.get(`/api/v1/carrier-integrations/${id}/sync-history/${syncId}`);
    return response.data || response;
  },

  getIntegrationHealth: async (id) => {
    const response = await api.get(`/api/v1/carrier-integrations/${id}/health`);
    return response.data || response;
  }
};

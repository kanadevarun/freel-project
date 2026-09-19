import api from './api';

export const memoryService = {
  // Memory Items
  listMemories: async (params = {}) => {
    const res = await api.get('/api/v1/memory', { params });
    return res?.data || res;
  },

  getMemory: async (id) => {
    const res = await api.get(`/api/v1/memory/${id}`);
    return res?.data || res;
  },

  proposeMemory: async (data) => {
    const res = await api.post('/api/v1/memory/propose', data);
    return res?.data || res;
  },

  createMemory: async (data) => {
    const res = await api.post('/api/v1/memory', data);
    return res?.data || res;
  },

  updateMemory: async (id, data) => {
    const res = await api.put(`/api/v1/memory/${id}`, data);
    return res?.data || res;
  },

  deleteMemory: async (id) => {
    const res = await api.delete(`/api/v1/memory/${id}`);
    return res?.data || res;
  },

  disableMemory: async (id) => {
    const res = await api.post(`/api/v1/memory/${id}/disable`);
    return res?.data || res;
  },

  enableMemory: async (id) => {
    const res = await api.post(`/api/v1/memory/${id}/enable`);
    return res?.data || res;
  },

  clearPersonalMemories: async () => {
    const res = await api.post('/api/v1/memory/clear-personal');
    return res?.data || res;
  },

  // Personalization Settings
  getUserSettings: async () => {
    const res = await api.get('/api/v1/memory/settings');
    return res?.data || res;
  },

  updateUserSettings: async (data) => {
    const res = await api.put('/api/v1/memory/settings', data);
    return res?.data || res;
  },

  togglePersonalization: async (enabled) => {
    const res = await api.post('/api/v1/memory/toggle', { enabled });
    return res?.data || res;
  },

  // Preferences
  listPreferences: async (scope = '') => {
    const res = await api.get('/api/v1/memory/preferences', { params: { scope } });
    return res?.data || res;
  },

  setPreference: async (data) => {
    const res = await api.put('/api/v1/memory/preferences', data);
    return res?.data || res;
  },

  deletePreference: async (key, scope = '') => {
    const res = await api.delete(`/api/v1/memory/preferences/${encodeURIComponent(key)}`, {
      params: { scope }
    });
    return res?.data || res;
  },

  // Runtime Context Synthesis
  getRuntimeContext: async (params = {}) => {
    const res = await api.post('/api/v1/memory/runtime-context', params);
    return res?.data || res;
  },

  // Stats & Audit
  getStats: async () => {
    const res = await api.get('/api/v1/memory/stats');
    return res?.data || res;
  },

  listAuditEvents: async (params = {}) => {
    const res = await api.get('/api/v1/memory/audit', { params });
    return res?.data || res;
  }
};

export default memoryService;

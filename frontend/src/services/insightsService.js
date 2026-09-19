import api from './api';

export const insightsService = {
  getCrossModuleInsights: async (entityType, entityId) => {
    let url = '/api/v1/insights/cross-module';
    const params = new URLSearchParams();
    if (entityType) params.append('entity_type', entityType);
    if (entityId) params.append('entity_id', entityId);
    const queryString = params.toString();
    if (queryString) url += `?${queryString}`;
    return api.get(url);
  },

  getOrgCrossModuleSummary: async () => {
    return api.get('/api/v1/insights/summary');
  },
};

export default insightsService;

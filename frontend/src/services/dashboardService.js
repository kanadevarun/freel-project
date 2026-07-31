import api from './api';

export const dashboardService = {
  getMissionControl: async () => {
    return api.get('/api/v1/dashboard/mission-control');
  },
};

import api from './api';

/**
 * dashboardService — all calls for the Mission Control dashboard.
 */
export const dashboardService = {
  /**
   * Fetch the full Mission Control payload:
   * { stats, approval_queue, ai_status }
   */
  getMissionControl: async () => {
    return api.get('/api/v1/dashboard/mission-control');
  },

  /**
   * Fetch the global activity timeline for the Mission Control dashboard.
   * Returns an array of activity events sorted newest-first.
   *
   * Each event has: { id, actor_name, actor_type, action, entity_type,
   *                   entity_id, metadata, created_at }
   *
   * actor_type is one of: "AI", "SALES", "PRICING", "CUSTOMER", "SYSTEM"
   */
  getActivities: async (limit = 20) => {
    return api.get(`/api/v1/activities?limit=${limit}`);
  },
};

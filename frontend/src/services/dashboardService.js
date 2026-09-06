import api from './api';

/**
 * dashboardService — all calls for the Mission Control dashboard.
 */
export const dashboardService = {
  /**
   * Fetch the full Mission Control payload:
   * { stats, pipeline, shipment_status, invoice_summary, approval_queue, attention_items, date_range, ai_status }
   */
  getMissionControl: async (params = {}) => {
    const searchParams = new URLSearchParams();
    if (params.startDate) searchParams.set('start_date', params.startDate);
    if (params.endDate) searchParams.set('end_date', params.endDate);
    if (params.preset) searchParams.set('preset', params.preset);
    const queryString = searchParams.toString();
    const url = queryString ? `/api/v1/dashboard/mission-control?${queryString}` : '/api/v1/dashboard/mission-control';
    return api.get(url);
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

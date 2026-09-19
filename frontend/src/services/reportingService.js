import api from './api';

/**
 * Advanced Reporting and Forecasting Service for LogisticsHQ
 * Connects frontend to the authoritative Go reporting layer & Python forecasting models.
 */
export const reportingService = {
  /**
   * Fetch an advanced report with deterministic metrics and AI-generated narrative/forecasts
   * @param {string} reportType - OPERATIONAL_VOLUME | REVENUE_FINANCE | COMMERCIAL_FUNNEL | CONTRACT_COMPLIANCE
   * @param {string} dateRange - LAST_30D | LAST_90D | YTD | LAST_12M
   */
  async getAdvancedReport(reportType = 'OPERATIONAL_VOLUME', dateRange = 'LAST_90D') {
    const response = await api.get('/api/v1/reports/advanced', {
      params: {
        report_type: reportType,
        date_range: dateRange,
      },
    });
    return response.data;
  },

  /**
   * Request authoritative report export (CSV or JSON)
   * @param {Object} payload - { report_type, export_format, date_range }
   */
  async exportReport(payload) {
    const response = await api.post('/api/v1/reports/export', payload);
    return response.data;
  },

  /**
   * Submit an external report distribution request (gated by Human Approval Center)
   * @param {Object} payload - { report_type, recipient_emails, distribution_channel, notes }
   */
  async requestDistribution(payload) {
    const response = await api.post('/api/v1/reports/distribute', payload);
    return response.data;
  },

  /**
   * Get past snapshot generation history
   * @param {number} limit
   */
  async getReportHistory(limit = 10) {
    const response = await api.get('/api/v1/reports/history', {
      params: { limit },
    });
    return response.data?.snapshots || [];
  },

  /**
   * Get past external distribution requests and their approval statuses
   * @param {number} limit
   */
  async getDistributionHistory(limit = 10) {
    const response = await api.get('/api/v1/reports/distributions', {
      params: { limit },
    });
    return response.data?.distributions || [];
  },

  /**
   * Legacy KPI summary metrics endpoint
   */
  async getMetrics() {
    const response = await api.get('/api/v1/reports/metrics');
    return response.data;
  },
};

export default reportingService;

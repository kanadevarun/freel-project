import api from './api';

export const auditService = {
  /**
   * List audit logs with optional filters and pagination
   * @param {Object} params
   * @param {number} [params.page=1]
   * @param {number} [params.limit=20]
   * @param {string} [params.module]
   * @param {string} [params.action]
   * @param {string} [params.result]
   * @param {string} [params.actor_type]
   * @param {number} [params.actor_id]
   * @param {string} [params.search]
   * @param {string} [params.start_date]
   * @param {string} [params.end_date]
   */
  async getAuditLogs(params = {}) {
    const query = new URLSearchParams();
    Object.entries(params).forEach(([key, val]) => {
      if (val !== undefined && val !== null && val !== '' && val !== 'ALL') {
        query.append(key, val);
      }
    });
    const queryString = query.toString();
    const endpoint = queryString ? `/api/v1/audit-logs?${queryString}` : '/api/v1/audit-logs';
    const response = await api.get(endpoint);
    return response?.data || response || { items: [], total: 0, page: 1, limit: 20, total_pages: 1 };
  },

  /**
   * Get single audit log details by ID
   * @param {number|string} id
   */
  async getAuditLogById(id) {
    const response = await api.get(`/api/v1/audit-logs/${id}`);
    return response?.data || response;
  },

  /**
   * Download filtered audit logs as a clean CSV
   * @param {Array} logs
   */
  exportToCSV(logs = []) {
    if (!logs || logs.length === 0) return;

    const headers = ['ID', 'Timestamp (UTC)', 'Actor Name', 'Actor Role', 'Actor Type', 'Action', 'Module', 'Resource Type', 'Resource ID', 'Description', 'Result', 'IP Address'];
    
    const rows = logs.map(log => [
      log.id,
      log.created_at || '',
      `"${(log.actor_name || 'System').replace(/"/g, '""')}"`,
      `"${(log.actor_role || '').replace(/"/g, '""')}"`,
      log.actor_type || 'USER',
      log.action || '',
      log.module || '',
      log.resource_type || '',
      `"${(log.resource_id || '').replace(/"/g, '""')}"`,
      `"${(log.description || '').replace(/"/g, '""')}"`,
      log.result || 'SUCCESS',
      log.ip_address || ''
    ]);

    const csvContent = [headers.join(','), ...rows.map(r => r.join(','))].join('\n');
    const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' });
    const url = URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.setAttribute('href', url);
    link.setAttribute('download', `audit-logs-${new Date().toISOString().slice(0, 10)}.csv`);
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    URL.revokeObjectURL(url);
  }
};

export default auditService;

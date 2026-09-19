import api from './api';

const BASE = '/api/v1/integrations';

export const integrationService = {
  /**
   * Retrieve operational status across all external integrations.
   * Returns masked configurations with honest statuses (ENABLED, DISABLED, NOT_CONFIGURED, etc.)
   */
  getStatuses: () => api.get(`${BASE}/status`),

  /**
   * Retrieve tenant configuration metadata (all secrets masked).
   */
  getConfigs: () => api.get(`${BASE}/configs`),

  /**
   * Save or update integration settings (non-secret options and masked credentials).
   * @param {Object} data - { type, provider, is_enabled, endpoint_url, region, from_address, timeout_sec, max_retries, secret_value }
   */
  saveConfig: (data) => api.put(`${BASE}/configs`, data),

  /**
   * Retrieve dead-letter webhook events for the tenant.
   */
  getDeadLetter: () => api.get(`${BASE}/dead-letter`),

  /**
   * Test SMS dispatcher to verify unconfigured providers fail cleanly.
   */
  testSMS: (data) => api.post(`${BASE}/test/sms`, data),

  /**
   * Test Email dispatcher to verify unconfigured providers fail cleanly.
   */
  testEmail: (data) => api.post(`${BASE}/test/email`, data),

  /**
   * Test tracking dispatcher.
   */
  testTracking: (carrierSCAC, trackingNumber) => 
    api.get(`${BASE}/test/tracking?carrier_scac=${carrierSCAC}&tracking_number=${trackingNumber}`),

  /**
   * Retrieve outbound SMS dispatch logs and delivery status.
   */
  getSMSMessages: (limit = 50) => api.get(`${BASE}/sms?limit=${limit}`),

  /**
   * Retrieve outbound Email dispatch logs and delivery status.
   */
  getEmailMessages: (limit = 50) => api.get(`${BASE}/email?limit=${limit}`),
};

export default integrationService;

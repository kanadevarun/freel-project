import api from './api';

/**
 * AI Governance and Production Controls Service for LogisticsHQ
 * Controls tenant AI policies, kill switches, prompt injection gates, and safety audit logs.
 */
export const governanceService = {
  /**
   * Fetch governance overview including system status, active kill switches, and 24h violation count
   */
  async getOverview() {
    const response = await api.get('/api/v1/governance/overview');
    return response.data;
  },

  /**
   * Fetch organization AI governance policy
   */
  async getPolicy() {
    const response = await api.get('/api/v1/governance/policy');
    return response.data;
  },

  /**
   * Update organization AI governance policy (Admin role required)
   * @param {Object} policy - Updated policy object
   */
  async updatePolicy(policy) {
    const response = await api.put('/api/v1/governance/policy', policy);
    return response.data;
  },

  /**
   * List all active and configured kill switches
   */
  async getKillSwitches() {
    const response = await api.get('/api/v1/governance/kill-switches');
    return response.data?.kill_switches || [];
  },

  /**
   * Set or toggle an AI kill switch (Admin role required)
   * @param {Object} payload - { scope, target_identifier, is_killed, reason }
   */
  async setKillSwitch(payload) {
    const response = await api.post('/api/v1/governance/kill-switches', payload);
    return response.data;
  },

  /**
   * List safety violations and rejection audit logs
   * @param {number} limit
   * @param {number} offset
   */
  async listViolations(limit = 20, offset = 0) {
    const response = await api.get('/api/v1/governance/violations', {
      params: { limit, offset },
    });
    return response.data;
  },

  /**
   * Perform safety pre-execution inspection on prompt or payload
   * @param {Object} payload - { workflow_name, model_name, provider_name, input_text, input_payload }
   */
  async inspectInput(payload) {
    const response = await api.post('/api/v1/governance/inspect', payload);
    return response.data;
  },

  /**
   * Trigger quality benchmark evaluation on test dataset
   * @param {Object} payload - { workflow_name, test_cases }
   */
  async evaluateQuality(payload) {
    const response = await api.post('/api/v1/governance/evaluate', payload);
    return response.data;
  },
};

export default governanceService;

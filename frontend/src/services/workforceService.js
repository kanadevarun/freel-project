import api from './api';

export const workforceService = {
  /**
   * List registered AI workforce agents
   */
  async listAgents(params = {}) {
    const query = new URLSearchParams();
    if (params.agent_type) query.append('agent_type', params.agent_type);
    if (params.is_enabled !== undefined) query.append('is_enabled', params.is_enabled);

    const qs = query.toString();
    return await api.get(qs ? `/api/v1/workforce/agents?${qs}` : '/api/v1/workforce/agents');
  },

  /**
   * Get specific agent metadata
   */
  async getAgent(agentId) {
    return await api.get(`/api/v1/workforce/agents/${agentId}`);
  },

  /**
   * Register custom agent
   */
  async registerAgent(payload) {
    return await api.post('/api/v1/workforce/agents', payload);
  },

  /**
   * Update agent metadata/status
   */
  async updateAgent(agentId, payload) {
    return await api.put(`/api/v1/workforce/agents/${agentId}`, payload);
  },

  /**
   * List standard agent capabilities
   */
  async listCapabilities() {
    return await api.get('/api/v1/workforce/capabilities');
  },

  /**
   * Create a new workforce task
   */
  async createTask(payload) {
    return await api.post('/api/v1/workforce/tasks', payload);
  },

  /**
   * List workforce tasks with optional filters
   */
  async listTasks(params = {}) {
    const query = new URLSearchParams();
    if (params.status) query.append('status', params.status);
    if (params.assigned_agent_id) query.append('assigned_agent_id', params.assigned_agent_id);
    if (params.parent_task_id) query.append('parent_task_id', params.parent_task_id);
    if (params.root_task_id) query.append('root_task_id', params.root_task_id);
    if (params.limit) query.append('limit', params.limit);
    if (params.offset) query.append('offset', params.offset);

    const qs = query.toString();
    return await api.get(qs ? `/api/v1/workforce/tasks?${qs}` : '/api/v1/workforce/tasks');
  },

  /**
   * Get workforce task details
   */
  async getTask(taskId) {
    return await api.get(`/api/v1/workforce/tasks/${taskId}`);
  },

  /**
   * Get task hierarchy tree
   */
  async getTaskHierarchy(taskId) {
    return await api.get(`/api/v1/workforce/tasks/${taskId}/hierarchy`);
  },

  /**
   * Execute task via Go boundary and AI sidecar
   */
  async executeTask(taskId) {
    return await api.post(`/api/v1/workforce/tasks/${taskId}/execute`, {});
  },

  /**
   * Delegate subtask to another agent
   */
  async delegateTask(taskId, payload) {
    return await api.post(`/api/v1/workforce/tasks/${taskId}/delegate`, payload);
  },

  /**
   * Initiate structured handoff
   */
  async initiateHandoff(taskId, payload) {
    return await api.post(`/api/v1/workforce/tasks/${taskId}/handoff`, payload);
  },

  /**
   * List handoffs for a task
   */
  async listHandoffs(taskId) {
    return await api.get(`/api/v1/workforce/tasks/${taskId}/handoffs`);
  },

  /**
   * Add context item with fact/prediction segregation
   */
  async addContextItem(taskId, payload) {
    return await api.post(`/api/v1/workforce/tasks/${taskId}/contexts`, payload);
  },

  /**
   * List context items for a task
   */
  async listContextItems(taskId) {
    return await api.get(`/api/v1/workforce/tasks/${taskId}/contexts`);
  },

  /**
   * Post structured inter-agent message
   */
  async postMessage(taskId, payload) {
    return await api.post(`/api/v1/workforce/tasks/${taskId}/messages`, payload);
  },

  /**
   * List structured inter-agent messages for a task
   */
  async listMessages(taskId) {
    return await api.get(`/api/v1/workforce/tasks/${taskId}/messages`);
  },

  /**
   * Phase 6.7: Assess contract risk workflow
   */
  async assessContractRisk(payload) {
    return await api.post(`/api/v1/workforce/risks/contract`, payload);
  },

  /**
   * Phase 6.7: Assess compliance risk workflow
   */
  async assessComplianceRisk(payload) {
    return await api.post(`/api/v1/workforce/risks/compliance`, payload);
  },

  /**
   * Phase 6.7: Assess cross-module risk workflow
   */
  async assessCrossModuleRisk(payload) {
    return await api.post(`/api/v1/workforce/risks/cross-module`, payload);
  },

  /**
   * Phase 6.7: Reassess cross-module risk
   */
  async reassessCrossModuleRisk(payload) {
    return await api.post(`/api/v1/workforce/risks/reassess`, payload);
  },

  /**
   * Phase 6.8: Resolve conflict
   */
  async resolveConflict(payload) {
    return await api.post(`/api/v1/workforce/conflicts/resolve`, payload);
  },

  /**
   * Phase 6.8: Record agent outcome
   */
  async recordOutcome(payload) {
    return await api.post(`/api/v1/workforce/outcomes`, payload);
  },

  /**
   * Phase 6.8: List agent outcomes
   */
  async listOutcomes(params = {}) {
    const qs = new URLSearchParams(params).toString();
    return await api.get(qs ? `/api/v1/workforce/outcomes?${qs}` : `/api/v1/workforce/outcomes`);
  },

  /**
   * Phase 6.8: Query memory with relevance filtering
   */
  async queryMemory(payload) {
    return await api.post(`/api/v1/workforce/memory/query`, payload);
  },

  /**
   * Phase 6.9: Workforce Command Center Overview
   */
  async getCommandCenterOverview() {
    return await api.get('/api/v1/workforce/command-center/overview');
  },

  /**
   * Phase 6.9: Workforce Health Summary
   */
  async getWorkforceHealth() {
    return await api.get('/api/v1/workforce/command-center/health');
  },

  /**
   * Phase 6.9: Agent Workload Metrics
   */
  async getAgentWorkload() {
    return await api.get('/api/v1/workforce/command-center/workload');
  },

  /**
   * Phase 6.9: Trigger or clear emergency stop
   */
  async emergencyStop(payload) {
    return await api.post('/api/v1/workforce/command-center/emergency-stop', payload);
  },

  /**
   * Phase 6.9: Get current emergency stop status
   */
  async getEmergencyStopStatus() {
    return await api.get('/api/v1/workforce/command-center/emergency-stop');
  },

  /**
   * Phase 6.9: List workforce waiting approvals
   */
  async listWorkforceApprovals() {
    return await api.get('/api/v1/workforce/command-center/approvals');
  },

  /**
   * Phase 6.9: List workforce escalations
   */
  async listWorkforceEscalations() {
    return await api.get('/api/v1/workforce/command-center/escalations');
  },

  /**
   * Phase 6.9: Get recent workforce activity
   */
  async getRecentActivity(limit = 20) {
    return await api.get(`/api/v1/workforce/command-center/activity?limit=${limit}`);
  },

  /**
   * Phase 6.9: Evaluate action policy under governed autonomy model
   */
  async evaluateActionPolicy(payload) {
    return await api.post('/api/v1/workforce/command-center/evaluate-action', payload);
  },

  /**
   * Phase 6.9: Governed agent control (ENABLE, DISABLE, PAUSE, RESUME, SET_AUTONOMY)
   */
  async controlAgent(agentId, payload) {
    return await api.post(`/api/v1/workforce/agents/${agentId}/control`, payload);
  },

  /**
   * Phase 6.9: Inspect multi-agent workflow tree/steps
   */
  async inspectWorkflow(planId) {
    return await api.get(`/api/v1/workforce/plans/${planId}/inspect`);
  },

  /**
   * Phase 6.9: Control workflow (PAUSE, RESUME, STOP)
   */
  async controlWorkflow(planId, payload) {
    return await api.post(`/api/v1/workforce/plans/${planId}/control`, payload);
  },
};

export default workforceService;

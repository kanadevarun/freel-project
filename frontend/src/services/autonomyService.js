/**
 * autonomyService.js — Phase 5 Controlled Autonomy Client
 *
 * Communicates with backend endpoints:
 *   /api/v1/autonomy/plans/...
 *   /api/v1/autonomy/goals/...
 *   /api/v1/autonomy/policies/...
 *   /api/v1/autonomy/context/...
 */

import api from './api';

export const autonomyService = {
  // Goal Operations
  createGoal: async (payload) => {
    return api.post('/api/v1/autonomy/goals', payload);
  },

  listGoals: async (params = {}) => {
    const query = new URLSearchParams(params).toString();
    return api.get(`/api/v1/autonomy/goals${query ? `?${query}` : ''}`);
  },

  getGoal: async (goalId) => {
    return api.get(`/api/v1/autonomy/goals/${goalId}`);
  },

  // Context Assembly
  getContext: async (module, entityId, type = 'SHIPMENT') => {
    return api.get(`/api/v1/autonomy/context/${module}/${entityId}?type=${type}`);
  },

  // Plan Operations
  generatePlan: async (payload) => {
    return api.post('/api/v1/autonomy/plans/generate', payload);
  },

  listPlans: async (params = {}) => {
    const query = new URLSearchParams(params).toString();
    return api.get(`/api/v1/autonomy/plans${query ? `?${query}` : ''}`);
  },

  getPlan: async (planId) => {
    return api.get(`/api/v1/autonomy/plans/${planId}`);
  },

  getCandidates: async (planId) => {
    return api.get(`/api/v1/autonomy/plans/${planId}/candidates`);
  },

  selectCandidate: async (planId, candidateId, reason = '') => {
    return api.post(`/api/v1/autonomy/plans/${planId}/select-candidate`, {
      candidate_id: candidateId,
      reason,
    });
  },

  revalidatePlan: async (planId, forceRefresh = false) => {
    return api.post(`/api/v1/autonomy/plans/${planId}/revalidate`, {
      force_refresh: forceRefresh,
    });
  },

  getPlanVersions: async (planId) => {
    return api.get(`/api/v1/autonomy/plans/${planId}/versions`);
  },

  executeStep: async (planId, stepId) => {
    return api.post(`/api/v1/autonomy/plans/${planId}/steps/${stepId}/execute`);
  },

  approvePlan: async (planId, notes = '') => {
    return api.post(`/api/v1/autonomy/plans/${planId}/approve`, { notes });
  },

  rejectPlan: async (planId, reason = '') => {
    return api.post(`/api/v1/autonomy/plans/${planId}/reject`, { reason });
  },

  pausePlan: async (planId, reason = '') => {
    return api.post(`/api/v1/autonomy/plans/${planId}/pause`, { reason });
  },

  resumePlan: async (planId) => {
    return api.post(`/api/v1/autonomy/plans/${planId}/resume`);
  },

  cancelPlan: async (planId, reason = '') => {
    return api.post(`/api/v1/autonomy/plans/${planId}/cancel`, { reason });
  },

  replan: async (planId, payload) => {
    return api.post(`/api/v1/autonomy/plans/${planId}/replan`, payload);
  },

  getAuditHistory: async (planId) => {
    return api.get(`/api/v1/autonomy/plans/${planId}/audit`);
  },

  // Policy Operations
  listPolicies: async () => {
    return api.get('/api/v1/autonomy/policies');
  },

  getPolicy: async (module) => {
    return api.get(`/api/v1/autonomy/policies/${module}`);
  },

  setPolicy: async (policy) => {
    return api.post('/api/v1/autonomy/policies', policy);
  },

  // Phase 5 Task 5.3: Adaptive Shipment Management
  ingestShipmentEvent: async (shipmentId, payload) => {
    return api.post(`/api/v1/autonomy/shipments/${shipmentId}/events`, payload);
  },

  getShipmentAdaptiveState: async (shipmentId) => {
    return api.get(`/api/v1/autonomy/shipments/${shipmentId}/adaptive-state`);
  },

  listShipmentEvents: async (shipmentId, limit = 20) => {
    return api.get(`/api/v1/autonomy/shipments/${shipmentId}/events?limit=${limit}`);
  },

  transitionWaitingState: async (planId, payload) => {
    return api.post(`/api/v1/autonomy/plans/${planId}/waiting-state`, payload);
  },

  // Phase 5 Task 5.4: Autonomous Customer Follow-Up
  getCustomerFollowupState: async (customerId) => {
    return api.get(`/api/v1/autonomy/customers/${customerId}/followup-state`);
  },

  updateCustomerPreferences: async (customerId, payload) => {
    return api.put(`/api/v1/autonomy/customers/${customerId}/preferences`, payload);
  },

  ingestCustomerFollowupEvent: async (customerId, payload) => {
    return api.post(`/api/v1/autonomy/customers/${customerId}/events`, payload);
  },

  sendCustomerFollowup: async (customerId, recordId) => {
    return api.post(`/api/v1/autonomy/customers/${customerId}/records/${recordId}/send`);
  },

  ingestCustomerResponse: async (customerId, recordId, payload) => {
    return api.post(`/api/v1/autonomy/customers/${customerId}/records/${recordId}/response`, payload);
  },

  // Phase 5 Task 5.5: Intelligent RFQ and Pricing Optimization
  evaluateRfqPricing: async (rfqId) => {
    return api.post(`/api/v1/autonomy/pricing/rfqs/${rfqId}/evaluate`);
  },

  getRfqPricingState: async (rfqId) => {
    return api.get(`/api/v1/autonomy/pricing/rfqs/${rfqId}/state`);
  },

  selectPricingStrategy: async (rfqId, strategyId) => {
    return api.post(`/api/v1/autonomy/pricing/rfqs/${rfqId}/select-strategy`, {
      strategy_id: strategyId,
    });
  },

  executePricingQuotation: async (rfqId) => {
    return api.post(`/api/v1/autonomy/pricing/rfqs/${rfqId}/execute-quote`);
  },

  replanRfqPricing: async (rfqId, reason, rateDelta = 0.0) => {
    return api.post(`/api/v1/autonomy/pricing/rfqs/${rfqId}/replan`, {
      reason,
      rate_delta: rateDelta,
    });
  },

  // Phase 5 Task 5.6: Adaptive Finance and Collections
  evaluateFinanceCollection: async (invoiceId) => {
    return api.post(`/api/v1/autonomy/finance/invoices/${invoiceId}/evaluate`);
  },

  getFinanceCollectionState: async (invoiceId) => {
    return api.get(`/api/v1/autonomy/finance/invoices/${invoiceId}/state`);
  },

  selectFinanceCollectionStrategy: async (invoiceId, strategyId) => {
    return api.post(`/api/v1/autonomy/finance/invoices/${invoiceId}/select-strategy`, {
      strategy_id: strategyId,
    });
  },

  executeFinanceCollectionAction: async (invoiceId) => {
    return api.post(`/api/v1/autonomy/finance/invoices/${invoiceId}/execute-action`);
  },

  replanFinanceCollection: async (invoiceId, triggerEvent, payload = {}) => {
    return api.post(`/api/v1/autonomy/finance/invoices/${invoiceId}/replan`, {
      trigger_event: triggerEvent,
      payload,
    });
  },

  // Phase 5 Task 5.7: Contract and Compliance Monitoring
  evaluateContractCompliance: async (contractId) => {
    return api.post(`/api/v1/autonomy/compliance/contracts/${contractId}/evaluate`);
  },

  getContractComplianceState: async (contractId) => {
    return api.get(`/api/v1/autonomy/compliance/contracts/${contractId}/state`);
  },

  selectContractComplianceStrategy: async (contractId, strategyId) => {
    return api.post(`/api/v1/autonomy/compliance/contracts/${contractId}/select-strategy`, {
      strategy_id: strategyId,
    });
  },

  executeContractComplianceAction: async (contractId) => {
    return api.post(`/api/v1/autonomy/compliance/contracts/${contractId}/execute-action`);
  },

  replanContractCompliance: async (contractId, triggerEvent, payload = {}) => {
    return api.post(`/api/v1/autonomy/compliance/contracts/${contractId}/replan`, {
      trigger_event: triggerEvent,
      payload,
    });
  },

  // Phase 5 Task 5.8: Autonomous Exception Resolution
  evaluateExceptionResolution: async (exceptionId) => {
    return api.post(`/api/v1/autonomy/exceptions/${exceptionId}/evaluate`);
  },

  getExceptionResolutionState: async (exceptionId) => {
    return api.get(`/api/v1/autonomy/exceptions/${exceptionId}/state`);
  },

  selectExceptionResolutionStrategy: async (exceptionId, strategyId) => {
    return api.post(`/api/v1/autonomy/exceptions/${exceptionId}/select-strategy`, {
      strategy_id: strategyId,
    });
  },

  executeExceptionResolutionAction: async (exceptionId, actionType = null, parameters = {}) => {
    return api.post(`/api/v1/autonomy/exceptions/${exceptionId}/execute-action`, {
      action_type: actionType,
      parameters,
    });
  },
  replanExceptionResolution: async (exceptionId, triggerEvent, payload = {}) => {
    return api.post(`/api/v1/autonomy/exceptions/${exceptionId}/replan`, {
      trigger_event: triggerEvent,
      payload,
    });
  },

  // Phase 5 Task 5.9: Multi-Step AI Planning and Execution
  executeNextStep: async (planId) => {
    return api.post(`/api/v1/autonomy/plans/${planId}/execute-next`);
  },

  approveStep: async (planId, stepId, notes = '') => {
    return api.post(`/api/v1/autonomy/plans/${planId}/steps/${stepId}/approve`, { notes });
  },

  retryStep: async (planId, stepId, reason = '') => {
    return api.post(`/api/v1/autonomy/plans/${planId}/steps/${stepId}/retry`, { reason });
  },

  compensateStep: async (planId, stepId, reason = '') => {
    return api.post(`/api/v1/autonomy/plans/${planId}/steps/${stepId}/compensate`, { reason });
  },

  validatePlan: async (planId) => {
    return api.post(`/api/v1/autonomy/plans/${planId}/validate`);
  },

  checkPlanConflicts: async (planId) => {
    return api.get(`/api/v1/autonomy/plans/${planId}/conflicts`);
  },

  listEntityConflicts: async (entityType, entityId) => {
    return api.get(`/api/v1/autonomy/plans/conflicts?entity_type=${entityType}&entity_id=${entityId}`);
  },

  generateCrossModulePlan: async (payload) => {
    return api.post(`/api/v1/autonomy/plans/cross-module`, payload);
  },

  getPlanningMetrics: async () => {
    return api.get(`/api/v1/autonomy/plans/metrics`);
  },

  // Phase 5 Task 5.10: Continuous Monitoring and Replanning
  ingestMonitoringEvent: async (eventData) => {
    return api.post('/api/v1/autonomy/monitoring/events', eventData);
  },

  listMonitoringEvents: async (limit = 50) => {
    return api.get(`/api/v1/autonomy/monitoring/events?limit=${limit}`);
  },

  getPlanHealth: async (planId) => {
    return api.get(`/api/v1/autonomy/monitoring/plans/${planId}/health`);
  },

  triggerAdaptiveReplan: async (planId, reason = '') => {
    return api.post(`/api/v1/autonomy/monitoring/plans/${planId}/replan`, { reason });
  },

  getContinuousMonitoringMetrics: async () => {
    return api.get('/api/v1/autonomy/monitoring/metrics');
  },

  // Phase 5 Task 5.11: Human + AI Operating Model
  createDecisionPoint: async (payload) => {
    return api.post('/api/v1/autonomy/human-ai/decisions', payload);
  },

  listDecisionPoints: async (params = {}) => {
    const query = new URLSearchParams(params).toString();
    return api.get(`/api/v1/autonomy/human-ai/decisions${query ? `?${query}` : ''}`);
  },

  getDecisionPoint: async (decisionId) => {
    return api.get(`/api/v1/autonomy/human-ai/decisions/${decisionId}`);
  },

  submitHumanDecision: async (decisionId, payload) => {
    return api.post(`/api/v1/autonomy/human-ai/decisions/${decisionId}/decide`, payload);
  },

  stopWorkflow: async (planId, reason = '') => {
    return api.post(`/api/v1/autonomy/human-ai/plans/${planId}/stop`, { reason });
  },

  updateStepHumanEdit: async (planId, stepId, humanContent) => {
    return api.post(`/api/v1/autonomy/human-ai/plans/${planId}/steps/${stepId}/edit`, { human_content: humanContent });
  },

  invalidateApprovals: async (payload) => {
    return api.post('/api/v1/autonomy/human-ai/invalidate-approvals', payload);
  },

  getDecisionCenterSummary: async () => {
    return api.get('/api/v1/autonomy/human-ai/decision-center/summary');
  },

  // Phase 5 Task 5.12: Autonomous Operations Command Center
  getCommandCenterOverview: async () => {
    return api.get('/api/v1/autonomy/command-center/overview');
  },

  getCommandCenterCriticalAttention: async (limit = 50) => {
    return api.get(`/api/v1/autonomy/command-center/critical-attention?limit=${limit}`);
  },

  getCommandCenterWorkflows: async (params = {}) => {
    const query = new URLSearchParams(params).toString();
    return api.get(`/api/v1/autonomy/command-center/workflows${query ? `?${query}` : ''}`);
  },

  getCommandCenterDecisions: async (params = {}) => {
    const query = new URLSearchParams(params).toString();
    return api.get(`/api/v1/autonomy/command-center/decisions${query ? `?${query}` : ''}`);
  },

  getCommandCenterRisks: async () => {
    return api.get('/api/v1/autonomy/command-center/risks');
  },

  getCommandCenterActivity: async (limit = 20) => {
    return api.get(`/api/v1/autonomy/command-center/activity?limit=${limit}`);
  },

  getCommandCenterSystemHealth: async () => {
    return api.get('/api/v1/autonomy/command-center/system-health');
  },

  // Phase 5 Task 5.13: Agent Memory and Learning from Outcomes
  recordOutcome: async (payload) => {
    return api.post('/api/v1/autonomy/memory/outcomes', payload);
  },

  listOutcomes: async (params = {}) => {
    const clean = {};
    for (const [k, v] of Object.entries(params)) {
      if (v !== undefined && v !== null && v !== '' && v !== 'undefined') clean[k] = v;
    }
    const query = new URLSearchParams(clean).toString();
    return api.get(`/api/v1/autonomy/memory/outcomes${query ? `?${query}` : ''}`);
  },

  getOutcome: async (outcomeId) => {
    return api.get(`/api/v1/autonomy/memory/outcomes/${outcomeId}`);
  },

  verifyOutcome: async (outcomeId, payload) => {
    return api.post(`/api/v1/autonomy/memory/outcomes/${outcomeId}/verify`, payload);
  },

  retrieveContextualMemory: async (payload) => {
    return api.post('/api/v1/autonomy/memory/retrieve', payload);
  },

  listMemories: async (params = {}) => {
    const clean = {};
    for (const [k, v] of Object.entries(params)) {
      if (v !== undefined && v !== null && v !== '' && v !== 'undefined') clean[k] = v;
    }
    const query = new URLSearchParams(clean).toString();
    return api.get(`/api/v1/autonomy/memory/items${query ? `?${query}` : ''}`);
  },

  getMemory: async (memoryId) => {
    return api.get(`/api/v1/autonomy/memory/items/${memoryId}`);
  },

  correctMemory: async (memoryId, payload) => {
    return api.put(`/api/v1/autonomy/memory/items/${memoryId}/correct`, payload);
  },

  invalidateMemory: async (memoryId, payload) => {
    return api.post(`/api/v1/autonomy/memory/items/${memoryId}/invalidate`, payload);
  },

  flagMemoryUnreliable: async (memoryId, reason = '') => {
    return api.post(`/api/v1/autonomy/memory/items/${memoryId}/flag-unreliable`, { reason });
  },

  detectPatterns: async () => {
    return api.post('/api/v1/autonomy/memory/patterns/detect');
  },

  listPatterns: async (params = {}) => {
    const clean = {};
    for (const [k, v] of Object.entries(params)) {
      if (v !== undefined && v !== null && v !== '' && v !== 'undefined') clean[k] = v;
    }
    const query = new URLSearchParams(clean).toString();
    return api.get(`/api/v1/autonomy/memory/patterns${query ? `?${query}` : ''}`);
  },

  getMemoryLearningSummary: async () => {
    return api.get('/api/v1/autonomy/memory/summary');
  },

  // Phase 5 Task 5.14: Governance for Controlled Autonomy
  getGovernanceLimits: async () => {
    return api.get('/api/v1/autonomy/governance/limits');
  },

  updateGovernanceLimits: async (limits) => {
    return api.put('/api/v1/autonomy/governance/limits', limits);
  },

  toggleGovernanceKillSwitch: async (enabled, reason = '') => {
    return api.post('/api/v1/autonomy/governance/kill-switch', { enabled, reason });
  },

  getActionAllowlist: async (module = '') => {
    const query = module ? `?module=${encodeURIComponent(module)}` : '';
    return api.get(`/api/v1/autonomy/governance/allowlist${query}`);
  },

  saveActionAllowlistItem: async (item) => {
    return api.post('/api/v1/autonomy/governance/allowlist', item);
  },

  getGovernanceFeatureFlags: async () => {
    return api.get('/api/v1/autonomy/governance/flags');
  },

  updateGovernanceFeatureFlag: async (flagKey, data) => {
    return api.put(`/api/v1/autonomy/governance/flags/${encodeURIComponent(flagKey)}`, data);
  },

  listPolicyEvaluations: async (params = {}) => {
    const clean = {};
    for (const [k, v] of Object.entries(params)) {
      if (v !== undefined && v !== null && v !== '' && v !== 'undefined') clean[k] = v;
    }
    const query = new URLSearchParams(clean).toString();
    return api.get(`/api/v1/autonomy/governance/evaluations${query ? `?${query}` : ''}`);
  },

  listPolicyAuditLogs: async (params = {}) => {
    const clean = {};
    for (const [k, v] of Object.entries(params)) {
      if (v !== undefined && v !== null && v !== '' && v !== 'undefined') clean[k] = v;
    }
    const query = new URLSearchParams(clean).toString();
    return api.get(`/api/v1/autonomy/governance/audit-logs${query ? `?${query}` : ''}`);
  },

  getGovernanceTelemetry: async () => {
    return api.get('/api/v1/autonomy/governance/telemetry');
  },

  evaluateGovernancePolicy: async (payload) => {
    return api.post('/api/v1/autonomy/governance/evaluate', payload);
  },

  previewGovernancePlan: async (payload) => {
    return api.post('/api/v1/autonomy/governance/preview', payload);
  },
};

export default autonomyService;





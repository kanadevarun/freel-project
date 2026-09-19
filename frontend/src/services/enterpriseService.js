import api from './api';

export const enterpriseService = {
  /**
   * Get Enterprise Autonomous Platform Overview (active, paused, health, autonomy)
   */
  async getOverview() {
    return await api.get('/api/v1/enterprise/overview');
  },

  /**
   * List enterprise autonomous workflows with filtering
   */
  async listWorkflows(params = {}) {
    const query = new URLSearchParams();
    if (params.state) query.append('state', params.state);
    if (params.workflow_type) query.append('workflow_type', params.workflow_type);
    if (params.limit) query.append('limit', params.limit);
    if (params.offset) query.append('offset', params.offset);

    const qs = query.toString();
    return await api.get(qs ? `/api/v1/enterprise/workflows?${qs}` : '/api/v1/enterprise/workflows');
  },

  /**
   * Get specific workflow details by ID
   */
  async getWorkflow(workflowId) {
    return await api.get(`/api/v1/enterprise/workflows/${workflowId}`);
  },

  /**
   * Initiate a new enterprise autonomous workflow
   */
  async startWorkflow(payload) {
    return await api.post('/api/v1/enterprise/workflows', payload);
  },

  /**
   * Pause an active enterprise workflow
   */
  async pauseWorkflow(workflowId, reason = '') {
    return await api.post(`/api/v1/enterprise/workflows/${workflowId}/pause`, { reason });
  },

  /**
   * Resume a paused enterprise workflow
   */
  async resumeWorkflow(workflowId) {
    return await api.post(`/api/v1/enterprise/workflows/${workflowId}/resume`, {});
  },

  /**
   * Cancel an enterprise workflow
   */
  async cancelWorkflow(workflowId, reason = '') {
    return await api.post(`/api/v1/enterprise/workflows/${workflowId}/cancel`, { reason });
  },

  /**
   * Approve a workflow step waiting for human authorization
   */
  async approveStep(workflowId, stepId, notes = '') {
    return await api.post(`/api/v1/enterprise/workflows/${workflowId}/steps/${stepId}/approve`, { notes });
  },

  /**
   * Reject a workflow step proposal
   */
  async rejectStep(workflowId, stepId, reason = '') {
    return await api.post(`/api/v1/enterprise/workflows/${workflowId}/steps/${stepId}/reject`, { reason });
  },

  /**
   * Trigger recovery of interrupted workflows across restart
   */
  async recoverWorkflows() {
    return await api.post('/api/v1/enterprise/workflows/recover', {});
  },

  /**
   * Dispatch business event into enterprise orchestrator
   */
  async triggerEvent(payload) {
    return await api.post('/api/v1/enterprise/events/trigger', payload);
  },

  /**
   * Apply global or tenant-scoped emergency pause / kill-switch
   */
  async emergencyControl(payload) {
    return await api.post('/api/v1/enterprise/emergency-control', payload);
  },

  // -------------------------------------------------------------------
  // Phase 7.2 Autonomous Shipment Lifecycle
  // -------------------------------------------------------------------

  /**
   * Initiate autonomous shipment lifecycle
   */
  async initiateShipmentLifecycle(shipmentId, correlationId = '') {
    return await api.post(`/api/v1/enterprise/shipments/${shipmentId}/lifecycle/initiate`, {
      shipment_id: Number(shipmentId),
      correlation_id: correlationId
    });
  },

  /**
   * Get autonomous shipment lifecycle status
   */
  async getShipmentLifecycle(shipmentId) {
    return await api.get(`/api/v1/enterprise/shipments/${shipmentId}/lifecycle`);
  },

  /**
   * Send shipment lifecycle event (carrier, customs, departure, etc.)
   */
  async sendShipmentLifecycleEvent(shipmentId, eventPayload) {
    return await api.post(`/api/v1/enterprise/shipments/${shipmentId}/lifecycle/event`, eventPayload);
  },

  /**
   * Trigger predictive ETA evaluation
   */
  async evaluateShipmentETA(shipmentId) {
    return await api.post(`/api/v1/enterprise/shipments/${shipmentId}/lifecycle/eta-evaluate`, {});
  },

  /**
   * Request adaptive replanning
   */
  async replanShipmentLifecycle(shipmentId, reason) {
    return await api.post(`/api/v1/enterprise/shipments/${shipmentId}/lifecycle/replan`, { reason });
  },

  /**
   * Transition shipment to delivered
   */
  async deliverShipmentLifecycle(shipmentId) {
    return await api.post(`/api/v1/enterprise/shipments/${shipmentId}/lifecycle/deliver`, {});
  },

  /**
   * Execute post-delivery audit
   */
  async auditPostDelivery(shipmentId) {
    return await api.post(`/api/v1/enterprise/shipments/${shipmentId}/lifecycle/post-delivery-audit`, {});
  },

  /**
   * Record outcome for learning & memory
   */
  async recordShipmentOutcome(shipmentId, outcomeData) {
    return await api.post(`/api/v1/enterprise/shipments/${shipmentId}/lifecycle/outcome`, outcomeData);
  },

  // -------------------------------------------------------------------
  // Phase 7.3 Autonomous Quote-to-Cash Commercial Lifecycle
  // -------------------------------------------------------------------

  /**
   * Initiate autonomous commercial quote-to-cash workflow for an RFQ
   */
  async initiateCommercialLifecycle(rfqId, correlationId = '') {
    return await api.post(`/api/v1/enterprise/commercial/rfqs/${rfqId}/initiate`, {
      correlation_id: correlationId
    });
  },

  /**
   * Get autonomous commercial lifecycle details
   */
  async getCommercialLifecycle(workflowId) {
    return await api.get(`/api/v1/enterprise/commercial/${workflowId}`);
  },

  /**
   * Dispatch a commercial lifecycle event
   */
  async sendCommercialLifecycleEvent(eventPayload) {
    return await api.post('/api/v1/enterprise/commercial/event', eventPayload);
  },

  /**
   * Trigger automated RFQ requirements extraction
   */
  async extractCommercialRFQ(workflowId) {
    return await api.post(`/api/v1/enterprise/commercial/${workflowId}/extract`, {});
  },

  /**
   * Trigger multi-agent commercial qualification
   */
  async qualifyCommercialRFQ(workflowId) {
    return await api.post(`/api/v1/enterprise/commercial/${workflowId}/qualify`, {});
  },

  /**
   * Trigger pricing & margin optimization
   */
  async optimizeCommercialPricing(workflowId) {
    return await api.post(`/api/v1/enterprise/commercial/${workflowId}/optimize-pricing`, {});
  },

  /**
   * Validate contracts and trade compliance
   */
  async checkCommercialCompliance(workflowId) {
    return await api.post(`/api/v1/enterprise/commercial/${workflowId}/check-compliance`, {});
  },

  /**
   * Prepare structured quotation
   */
  async prepareCommercialQuote(workflowId) {
    return await api.post(`/api/v1/enterprise/commercial/${workflowId}/prepare-quote`, {});
  },

  /**
   * Analyze customer counter-offer and formulate negotiation tradeoffs
   */
  async negotiateCommercialQuote(workflowId, counterPrice, notes = '') {
    return await api.post(`/api/v1/enterprise/commercial/${workflowId}/negotiate`, {
      counter_price: Number(counterPrice),
      notes
    });
  },

  /**
   * Apply selected strategic negotiation option
   */
  async applyCommercialNegotiation(workflowId, optionId) {
    return await api.post(`/api/v1/enterprise/commercial/${workflowId}/apply-negotiation`, {
      option_id: optionId
    });
  },

  /**
   * Verify quotation acceptance
   */
  async acceptCommercialQuote(workflowId, quotationId) {
    return await api.post(`/api/v1/enterprise/commercial/${workflowId}/accept`, {
      quotation_id: Number(quotationId)
    });
  },

  /**
   * Convert accepted quote to operational carrier booking
   */
  async handoffCommercialBooking(workflowId) {
    return await api.post(`/api/v1/enterprise/commercial/${workflowId}/handoff-booking`, {});
  },

  /**
   * Hand off commercial booking to Phase 7.2 autonomous shipment lifecycle
   */
  async handoffCommercialShipment(workflowId, shipmentId) {
    return await api.post(`/api/v1/enterprise/commercial/${workflowId}/handoff-shipment`, {
      shipment_id: Number(shipmentId)
    });
  },

  /**
   * Evaluate billing readiness and generate invoice
   */
  async evaluateCommercialInvoice(workflowId) {
    return await api.post(`/api/v1/enterprise/commercial/${workflowId}/invoice`, {});
  },

  /**
   * Formulate receivables collection strategy
   */
  async assessCommercialCollection(workflowId) {
    return await api.post(`/api/v1/enterprise/commercial/${workflowId}/collection`, {});
  },

  /**
   * Record commercial outcome and feed governed workforce memory
   */
  async recordCommercialOutcome(workflowId, outcomeData) {
    return await api.post(`/api/v1/enterprise/commercial/${workflowId}/outcome`, {
      workflow_id: workflowId,
      ...outcomeData
    });
  },

  // -------------------------------------------------------------------
  // Phase 7.4 Autonomous Enterprise Exception Management
  // -------------------------------------------------------------------

  /**
   * Ingest and classify a new enterprise exception event
   */
  async detectException(eventPayload) {
    return await api.post('/api/v1/enterprise/exceptions/detect', eventPayload);
  },

  /**
   * Get specific autonomous exception workflow details
   */
  async getExceptionWorkflow(workflowId) {
    return await api.get(`/api/v1/enterprise/exceptions/${workflowId}`);
  },

  /**
   * Trigger multi-agent root cause analysis distinguishing facts from inferences
   */
  async investigateException(workflowId) {
    return await api.post(`/api/v1/enterprise/exceptions/${workflowId}/investigate`, {});
  },

  /**
   * Assess downstream cross-module operational, customer, financial, and compliance impact
   */
  async assessExceptionImpact(workflowId) {
    return await api.post(`/api/v1/enterprise/exceptions/${workflowId}/impact`, {});
  },

  /**
   * Generate structured recovery options with risk, autonomy level, and governance requirements
   */
  async planExceptionRecovery(workflowId) {
    return await api.post(`/api/v1/enterprise/exceptions/${workflowId}/plan`, {});
  },

  /**
   * Select and govern a recovery option (gating approvals if required)
   */
  async selectExceptionRecoveryOption(workflowId, optionId) {
    return await api.post(`/api/v1/enterprise/exceptions/${workflowId}/options/${optionId}/select`, {});
  },

  /**
   * Execute selected recovery action through authoritative Action System boundary
   */
  async executeExceptionRecoveryStep(workflowId, stepId) {
    return await api.post(`/api/v1/enterprise/exceptions/${workflowId}/execute`, { step_id: stepId });
  },

  /**
   * Verify recovery action against authoritative business data
   */
  async verifyExceptionRecovery(workflowId, stepId) {
    return await api.post(`/api/v1/enterprise/exceptions/${workflowId}/verify`, { step_id: stepId });
  },

  /**
   * Trigger adaptive replanning upon action failure (bounded retries <= 3)
   */
  async triggerAdaptiveRecovery(workflowId, failureReason) {
    return await api.post(`/api/v1/enterprise/exceptions/${workflowId}/replan`, { failure_reason: failureReason });
  },

  /**
   * Transition exception to post-resolution operational monitoring
   */
  async transitionExceptionToMonitoring(workflowId, monitoringMilestone) {
    return await api.post(`/api/v1/enterprise/exceptions/${workflowId}/monitor`, { monitoring_milestone: monitoringMilestone });
  },

  /**
   * Resolve exception with authoritative business proof
   */
  async resolveException(workflowId, resolutionEvidence) {
    return await api.post(`/api/v1/enterprise/exceptions/${workflowId}/resolve`, { resolution_evidence: resolutionEvidence });
  },

  /**
   * Escalate exception to human operations
   */
  async escalateException(workflowId, reason) {
    return await api.post(`/api/v1/enterprise/exceptions/${workflowId}/escalate`, { reason });
  },

  /**
   * Record exception resolution outcome to train workforce memory
   */
  async recordExceptionOutcome(workflowId, feedback) {
    return await api.post(`/api/v1/enterprise/exceptions/${workflowId}/outcome`, {
      workflow_id: workflowId,
      ...feedback
    });
  },

  // -------------------------------------------------------------------
  // Phase 7.5 Autonomous Customer Relationship Management
  // -------------------------------------------------------------------

  /**
   * Initiate autonomous customer relationship workflow
   */
  async initiateCustomerWorkflow(customerId) {
    return await api.post(`/api/v1/enterprise/crm/customers/${customerId}/initiate`, {});
  },

  /**
   * Get customer relationship workflow details
   */
  async getCustomerWorkflow(workflowId) {
    return await api.get(`/api/v1/enterprise/crm/${workflowId}`);
  },

  /**
   * Evaluate multi-signal customer health profile
   */
  async evaluateCustomerHealth(workflowId) {
    return await api.post(`/api/v1/enterprise/crm/${workflowId}/health`, {});
  },

  /**
   * Detect customer risk signals and commercial expansion opportunities
   */
  async detectCustomerRisksAndOpportunities(workflowId) {
    return await api.post(`/api/v1/enterprise/crm/${workflowId}/detect`, {});
  },

  /**
   * Trigger multi-agent root-cause investigation
   */
  async investigateCustomerRootCauses(workflowId) {
    return await api.post(`/api/v1/enterprise/crm/${workflowId}/investigate`, {});
  },

  /**
   * Plan structured customer intervention options
   */
  async planCustomerInterventions(workflowId) {
    return await api.post(`/api/v1/enterprise/crm/${workflowId}/plan`, {});
  },

  /**
   * Select and govern an intervention option (gating approvals if high risk)
   */
  async selectCustomerInterventionOption(workflowId, optionId) {
    return await api.post(`/api/v1/enterprise/crm/${workflowId}/options/${optionId}/select`, {});
  },

  /**
   * Execute selected intervention through Action System boundary
   */
  async executeCustomerIntervention(workflowId, stepId) {
    return await api.post(`/api/v1/enterprise/crm/${workflowId}/execute`, { step_id: stepId });
  },

  /**
   * Verify intervention delivery against authoritative records
   */
  async verifyCustomerIntervention(workflowId, stepId) {
    return await api.post(`/api/v1/enterprise/crm/${workflowId}/verify`, { step_id: stepId });
  },

  /**
   * Securely process customer message response
   */
  async processCustomerResponse(workflowId, message) {
    return await api.post(`/api/v1/enterprise/crm/${workflowId}/response`, { message });
  },

  /**
   * Ingest customer lifecycle event
   */
  async processCustomerLifecycleEvent(eventPayload) {
    return await api.post('/api/v1/enterprise/crm/event', eventPayload);
  },

  /**
   * Record customer relationship outcome to train workforce memory
   */
  async recordCustomerOutcome(workflowId, outcomeFeedback) {
    return await api.post(`/api/v1/enterprise/crm/${workflowId}/outcome`, {
      workflow_id: workflowId,
      ...outcomeFeedback
    });
  },

  // ---------------------------------------------------------------------------
  // Phase 7.6 Autonomous Revenue and Margin Optimization
  // ---------------------------------------------------------------------------

  /**
   * Initiate autonomous revenue and margin optimization workflow
   */
  async initiateRevenueWorkflow(entityType, entityId, correlationId = '') {
    return await api.post('/api/v1/enterprise/revenue/initiate', {
      entity_type: entityType,
      entity_id: entityId,
      correlation_id: correlationId
    });
  },

  /**
   * Get autonomous revenue workflow status
   */
  async getRevenueWorkflow(workflowId) {
    return await api.get(`/api/v1/enterprise/revenue/${workflowId}`);
  },

  /**
   * Evaluate cost structure and gross margin baseline
   */
  async evaluateRevenueCostAndMargin(workflowId) {
    return await api.post(`/api/v1/enterprise/revenue/${workflowId}/cost-margin`, {});
  },

  /**
   * Evaluate carrier cost reliability and economics
   */
  async evaluateRevenueCarrierEconomics(workflowId) {
    return await api.post(`/api/v1/enterprise/revenue/${workflowId}/carrier-economics`, {});
  },

  /**
   * Assess multi-dimensional customer value scorecard
   */
  async assessRevenueCustomerValue(workflowId) {
    return await api.post(`/api/v1/enterprise/revenue/${workflowId}/customer-value`, {});
  },

  /**
   * Synthesize multi-agent commercial reasoning
   */
  async conductRevenueMultiAgentOptimization(workflowId) {
    return await api.post(`/api/v1/enterprise/revenue/${workflowId}/optimize`, {});
  },

  /**
   * Formulate versioned pricing recommendations and options
   */
  async formulatePricingRecommendation(workflowId) {
    return await api.post(`/api/v1/enterprise/revenue/${workflowId}/recommend`, {});
  },

  /**
   * Select and govern commercial optimization option (with HITL approval if out of policy)
   */
  async selectRevenueOptimizationOption(workflowId, optionId) {
    return await api.post(`/api/v1/enterprise/revenue/${workflowId}/options/${optionId}/select`, {});
  },

  /**
   * Execute commercial action through Go Action System boundary
   */
  async executeRevenueCommercialAction(workflowId, stepId) {
    return await api.post(`/api/v1/enterprise/revenue/${workflowId}/execute`, { step_id: stepId });
  },

  /**
   * Verify commercial action persistence against authoritative business database
   */
  async verifyRevenueCommercialAction(workflowId, stepId) {
    return await api.post(`/api/v1/enterprise/revenue/${workflowId}/verify`, { step_id: stepId });
  },

  /**
   * Counter-optimize customer negotiation discount request
   */
  async optimizeNegotiationRequest(workflowId, targetDiscountPct) {
    return await api.post(`/api/v1/enterprise/revenue/${workflowId}/negotiate`, {
      target_discount_pct: targetDiscountPct
    });
  },

  /**
   * Ingest revenue/margin lifecycle event
   */
  async processRevenueLifecycleEvent(eventPayload) {
    return await api.post('/api/v1/enterprise/revenue/event', eventPayload);
  },

  /**
   * Record revenue realization outcome and update workforce memory
   */
  async recordRevenueOutcome(workflowId, outcomeFeedback) {
    return await api.post(`/api/v1/enterprise/revenue/${workflowId}/outcome`, {
      workflow_id: workflowId,
      ...outcomeFeedback
    });
  },

  // ---------------------------------------------------------------------------
  // Phase 7.7 Autonomous Contract, Compliance and Risk Governance
  // ---------------------------------------------------------------------------

  /**
   * Initiate autonomous contract, compliance and risk governance workflow
   */
  async initiateRiskWorkflow(entityType, entityId, correlationId = '') {
    return await api.post('/api/v1/enterprise/risk/initiate', {
      entity_type: entityType,
      entity_id: entityId,
      correlation_id: correlationId
    });
  },

  /**
   * Get autonomous risk governance workflow status
   */
  async getRiskWorkflow(workflowId) {
    return await api.get(`/api/v1/enterprise/risk/${workflowId}`);
  },

  /**
   * Collect risk evidence item with verified provenance
   */
  async collectRiskEvidence(workflowId, evidence) {
    return await api.post(`/api/v1/enterprise/risk/${workflowId}/evidence`, evidence);
  },

  /**
   * Synthesize multi-agent contract and statutory compliance assessment
   */
  async conductMultiAgentRiskAssessment(workflowId) {
    return await api.post(`/api/v1/enterprise/risk/${workflowId}/assess`, {});
  },

  /**
   * Assess cross-domain cascading risk propagation and exposure
   */
  async assessCrossDomainImpact(workflowId) {
    return await api.post(`/api/v1/enterprise/risk/${workflowId}/impact`, {});
  },

  /**
   * Formulate governed risk mitigation proposals
   */
  async planRiskMitigationOptions(workflowId) {
    return await api.post(`/api/v1/enterprise/risk/${workflowId}/plan`, {});
  },

  /**
   * Select and govern mitigation strategy (with approval gating if required)
   */
  async selectRiskMitigationOption(workflowId, optionId) {
    return await api.post(`/api/v1/enterprise/risk/${workflowId}/options/${optionId}/select`, {});
  },

  /**
   * Execute mitigation action via Go Action System boundary
   */
  async executeRiskMitigationAction(workflowId, stepId) {
    return await api.post(`/api/v1/enterprise/risk/${workflowId}/execute`, { step_id: stepId });
  },

  /**
   * Verify mitigation execution against authoritative external records
   */
  async verifyRiskMitigationAction(workflowId, stepId) {
    return await api.post(`/api/v1/enterprise/risk/${workflowId}/verify`, { step_id: stepId });
  },

  /**
   * Transition workflow to active post-remediation monitoring
   */
  async transitionRiskToMonitoring(workflowId) {
    return await api.post(`/api/v1/enterprise/risk/${workflowId}/monitor`, {});
  },

  /**
   * Formally resolve and close risk workflow
   */
  async resolveRiskWorkflow(workflowId, resolutionSummary = '') {
    return await api.post(`/api/v1/enterprise/risk/${workflowId}/resolve`, {
      resolution_summary: resolutionSummary
    });
  },

  /**
   * Adaptively reassess risk condition with new evidence
   */
  async reassessRiskCondition(workflowId, newEvidence) {
    return await api.post(`/api/v1/enterprise/risk/${workflowId}/reassess`, {
      new_evidence: newEvidence
    });
  },

  /**
   * Ingest risk lifecycle event
   */
  async processRiskLifecycleEvent(eventPayload) {
    return await api.post('/api/v1/enterprise/risk/event', eventPayload);
  },

  /**
   * Record risk outcome and train workforce memory
   */
  async recordRiskOutcome(workflowId, outcomeFeedback) {
    return await api.post(`/api/v1/enterprise/risk/${workflowId}/outcome`, {
      workflow_id: workflowId,
      ...outcomeFeedback
    });
  },

  // ---------------------------------------------------------------------
  // Phase 7.8: Enterprise Event Mesh & Autonomous Workflow Engine
  // ---------------------------------------------------------------------

  /**
   * Get Event Mesh overview KPIs, active workflows, and dead letter counts
   */
  async getEventMeshOverview() {
    return await api.get('/api/v1/enterprise/mesh/overview');
  },

  /**
   * Ingest a normalized business event into the Enterprise Event Mesh
   */
  async ingestMeshEvent(eventPayload) {
    return await api.post('/api/v1/enterprise/mesh/events', eventPayload);
  },

  /**
   * List ingested events with status, routing, and pagination
   */
  async listMeshEvents(params = {}) {
    const query = new URLSearchParams();
    if (params.limit) query.append('limit', params.limit);
    if (params.offset) query.append('offset', params.offset);
    const qs = query.toString();
    return await api.get(qs ? `/api/v1/enterprise/mesh/events?${qs}` : '/api/v1/enterprise/mesh/events');
  },

  /**
   * Get single event details with correlation & causation lineage
   */
  async getMeshEvent(eventId) {
    return await api.get(`/api/v1/enterprise/mesh/events/${eventId}`);
  },

  /**
   * Get dead-letter queue events requiring operator attention
   */
  async getMeshDeadLetters(limit = 50) {
    return await api.get(`/api/v1/enterprise/mesh/dead-letters?limit=${limit}`);
  },

  /**
   * Replay an unprocessable or dead-lettered event
   */
  async replayMeshDeadLetter(eventId) {
    return await api.post(`/api/v1/enterprise/mesh/dead-letters/${eventId}/replay`, {});
  },

  /**
   * List active deterministic event-to-workflow routing rules
   */
  async listMeshRoutingRules() {
    return await api.get('/api/v1/enterprise/mesh/rules');
  },

  /**
   * Phase 7.9 — Enterprise Autonomous Control Tower Comprehensive View
   */
  async getControlTowerView() {
    return await api.get('/api/v1/enterprise/control-tower/view');
  },

  /**
   * Phase 7.9 — Get Complete Trace (Event -> Workflow -> Steps -> Actions -> Verification -> Outcome)
   */
  async getControlTowerWorkflowTrace(workflowId) {
    return await api.get(`/api/v1/enterprise/control-tower/workflows/${workflowId}/trace`);
  },

  /**
   * Phase 7.9 — Perform Governed Control Action (PAUSE, RESUME, CANCEL)
   */
  async performControlTowerAction(workflowId, action, reason = '') {
    return await api.post(`/api/v1/enterprise/control-tower/workflows/${workflowId}/control`, { action, reason });
  },

  /**
   * Phase 7.10 — Evaluate Autonomous Action against Go Governance & Safety Boundary
   */
  async evaluateGovernanceAction(payload) {
    return await api.post('/api/v1/enterprise/governance/evaluate', payload);
  },

  /**
   * Phase 7.10 — Apply or Release Authoritative Emergency Control
   */
  async applyEmergencyControl(payload) {
    return await api.post('/api/v1/enterprise/governance/emergency', payload);
  },

  /**
   * Phase 7.10 — Get Active Emergency Controls
   */
  async getEmergencyControls() {
    return await api.get('/api/v1/enterprise/governance/emergency');
  },

  /**
   * Phase 7.10 — Get Centralized Autonomy Governance & Safety Status
   */
  async getGovernanceStatus() {
    return await api.get('/api/v1/enterprise/governance/status');
  },

  /**
   * Phase 7.10 — Record Human Rejection to prevent AI override loops
   */
  async recordHumanRejection(payload) {
    return await api.post('/api/v1/enterprise/governance/rejections', payload);
  },

  /**
   * Phase 7.11 — Get Unified Enterprise Platform Resilience Health
   */
  async getResilienceHealth() {
    return await api.get('/api/v1/enterprise/resilience/health');
  },

  /**
   * Phase 7.11 — Detect Stuck Autonomous Workflows
   */
  async getStuckWorkflows(thresholdSeconds = 600) {
    return await api.get(`/api/v1/enterprise/resilience/stuck-workflows?threshold_seconds=${thresholdSeconds}`);
  },

  /**
   * Phase 7.11 — Recover Stuck Autonomous Workflow from Checkpoint
   */
  async recoverStuckWorkflow(workflowId) {
    return await api.post('/api/v1/enterprise/resilience/recover-stuck', { workflow_id: workflowId });
  },

  /**
   * Phase 7.11 — Get Visible Failed Work Items & Dead Letters
   */
  async getFailedWorkItems(limit = 50) {
    return await api.get(`/api/v1/enterprise/resilience/failed-work?limit=${limit}`);
  },

  /**
   * Phase 7.11 — Replay Retryable Failed Work Item
   */
  async replayFailedWork(itemId) {
    return await api.post('/api/v1/enterprise/resilience/replay-work', { item_id: itemId });
  },

  /**
   * Phase 7.11 — Get Event Backpressure & Storm Metrics
   */
  async getBackpressureMetrics() {
    return await api.get('/api/v1/enterprise/resilience/backpressure');
  }
};

export default enterpriseService;


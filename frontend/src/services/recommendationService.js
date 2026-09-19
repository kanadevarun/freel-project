import api from './api';

export const recommendationService = {
  /**
   * List recommendations with query filters and pagination
   */
  async listRecommendations(params = {}) {
    const query = new URLSearchParams();
    if (params.page) query.append('page', params.page);
    if (params.limit) query.append('limit', params.limit);
    if (params.category) query.append('category', params.category);
    if (params.priority) query.append('priority', params.priority);
    if (params.risk_level) query.append('risk_level', params.risk_level);
    if (params.status) query.append('status', params.status);
    if (params.source_type) query.append('source_type', params.source_type);
    if (params.search) query.append('search', params.search);
    if (params.sort_by) query.append('sort_by', params.sort_by);
    if (params.sort_dir) query.append('sort_dir', params.sort_dir);
    if (params.requires_approval !== undefined && params.requires_approval !== '') {
      query.append('requires_approval', params.requires_approval);
    }
    if (params.assignee_id) query.append('assignee_id', params.assignee_id);
    if (params.rfq_id) query.append('rfq_id', params.rfq_id);
    if (params.quotation_id) query.append('quotation_id', params.quotation_id);
    if (params.shipment_id) query.append('shipment_id', params.shipment_id);
    if (params.milestone_id) query.append('milestone_id', params.milestone_id);
    if (params.exception_id) query.append('exception_id', params.exception_id);
    if (params.booking_id) query.append('booking_id', params.booking_id);
    if (params.customer_id) query.append('customer_id', params.customer_id);
    if (params.missing_info) query.append('missing_info', 'true');
    if (params.expiring_soon) query.append('expiring_soon', 'true');
    if (params.pricing_concern) query.append('pricing_concern', 'true');
    if (params.delayed_milestone) query.append('delayed_milestone', 'true');
    if (params.active_exception) query.append('active_exception', 'true');
    if (params.missing_ops_info) query.append('missing_ops_info', 'true');
    if (params.invoice_id) query.append('invoice_id', params.invoice_id);
    if (params.overdue_only) query.append('overdue_only', 'true');
    if (params.data_quality_only) query.append('data_quality_only', 'true');
    if (params.upcoming_only) query.append('upcoming_only', 'true');
    if (params.aging_band) query.append('aging_band', params.aging_band);
    // Task 2.6 Contract & Compliance filters
    if (params.contract_id) query.append('contract_id', params.contract_id);
    if (params.document_id) query.append('document_id', params.document_id);
    if (params.compliance_id) query.append('compliance_id', params.compliance_id);
    if (params.carrier_id) query.append('carrier_id', params.carrier_id);
    if (params.expired_only) query.append('expired_only', 'true');
    // Task 2.7 Automation filters
    if (params.automation_id) query.append('automation_id', params.automation_id);
    if (params.execution_id) query.append('execution_id', params.execution_id);

    const qs = query.toString();
    const endpoint = qs ? `/api/v1/recommendations?${qs}` : '/api/v1/recommendations';
    return await api.get(endpoint);
  },

  /**
   * Get recommendation KPI summary stats
   */
  async getStats() {
    return await api.get('/api/v1/recommendations/stats');
  },

  /**
   * Get single recommendation details
   */
  async getRecommendation(id) {
    return await api.get(`/api/v1/recommendations/${id}`);
  },

  /**
   * Trigger deterministic generation cycle
   */
  async generateRecommendations() {
    return await api.post('/api/v1/recommendations/generate', {});
  },

  /**
   * Update recommendation status with valid transition
   */
  async updateStatus(id, status, reason = '') {
    return await api.patch(`/api/v1/recommendations/${id}/status`, { status, reason });
  },

  /**
   * Assign recommendation to an operator
   */
  async assignRecommendation(id, assigneeId, assigneeName) {
    return await api.post(`/api/v1/recommendations/${id}/assign`, {
      assignee_id: Number(assigneeId),
      assignee_name: assigneeName,
    });
  },

  /**
   * Dismiss recommendation with mandatory reason
   */
  async dismissRecommendation(id, reason) {
    return await api.post(`/api/v1/recommendations/${id}/dismiss`, { reason });
  },

  /**
   * Mark recommendation as reviewed
   */
  async markReviewed(id) {
    return await api.post(`/api/v1/recommendations/${id}/review`, {});
  },

  /**
   * Retrieve structured factual evidence items
   */
  async getEvidence(id) {
    return await api.get(`/api/v1/recommendations/${id}/evidence`);
  },

  /**
   * List recommendations linked to a specific operational record
   */
  async listBySource(sourceType, sourceId) {
    return await api.get(`/api/v1/recommendations/source/${sourceType}/${sourceId}`);
  },

  // ── Customer Follow-Up Assistant Extensions (Phase 2 Task 2.2) ───────────

  /**
   * Generate an editable message draft for a follow-up recommendation upon explicit user request
   */
  async generateDraft(id) {
    return await api.post(`/api/v1/recommendations/${id}/draft`, {});
  },

  /**
   * Save user edited subject and body for a follow-up draft
   */
  async saveDraft(id, subject, body) {
    return await api.patch(`/api/v1/recommendations/${id}/draft`, { subject, body });
  },

  /**
   * Create a controlled, idempotent internal follow-up task
   */
  async createFollowupTask(id, taskData = {}) {
    return await api.post(`/api/v1/recommendations/${id}/task`, taskData);
  },

  /**
   * List internal follow-up tasks
   */
  async listFollowupTasks(params = {}) {
    const query = new URLSearchParams();
    if (params.customer_id) query.append('customer_id', params.customer_id);
    if (params.status) query.append('status', params.status);
    const qs = query.toString();
    const endpoint = qs ? `/api/v1/recommendations/tasks?${qs}` : '/api/v1/recommendations/tasks';
    return await api.get(endpoint);
  },

  /**
   * Get KPI metrics for the customer follow-up assistant pipeline
   */
  async getFollowupStats() {
    return await api.get('/api/v1/recommendations/followups/stats');
  },

  // ── RFQ and Quotation Workflow Assistant Extensions (Phase 2 Task 2.3) ───

  /**
   * Get controlled action preview before executing high-risk or commercial operations
   */
  async getActionPreview(id) {
    return await api.get(`/api/v1/recommendations/${id}/action-preview`);
  },

  /**
   * Submit a recommendation to the centralized HITL approval system
   */
  async requestApproval(id, notes = '') {
    return await api.post(`/api/v1/recommendations/${id}/request-approval`, { notes });
  },

  // ── Shipment Exception & Operations Copilot Extensions (Phase 2 Task 2.4) ──

  /**
   * Get operational evidence for a specific shipment
   */
  async getShipmentEvidence(shipmentId) {
    return await api.get(`/api/v1/recommendations/shipments/${shipmentId}/evidence`);
  },

  /**
   * Get operational evidence for a specific milestone
   */
  async getMilestoneEvidence(milestoneId) {
    return await api.get(`/api/v1/recommendations/milestones/${milestoneId}/evidence`);
  },

  /**
   * Get operational evidence for a specific shipment exception
   */
  async getExceptionEvidence(exceptionId) {
    return await api.get(`/api/v1/recommendations/exceptions/${exceptionId}/evidence`);
  },

  // ── Invoice and Collections Assistant Extensions (Phase 2 Task 2.5) ──────

  /**
   * Get complete financial evidence and aging analysis for an invoice
   */
  async getInvoiceEvidence(invoiceId) {
    return await api.get(`/api/v1/recommendations/invoices/${invoiceId}/evidence`);
  },

  /**
   * Get collection summary and payment-risk signals for a customer
   */
  async getCustomerCollectionSummary(customerId) {
    return await api.get(`/api/v1/recommendations/customers/${customerId}/collection-summary`);
  },

  // ── Contract, Document & Compliance Assistant Extensions (Phase 2 Task 2.6) ──

  /**
   * Get contract operational and compliance evidence
   */
  async getContractEvidence(contractId) {
    return await api.get(`/api/v1/recommendations/contracts/${contractId}/evidence`);
  },

  /**
   * Get document verification and extraction evidence
   */
  async getDocumentEvidence(documentId) {
    return await api.get(`/api/v1/recommendations/documents/${documentId}/evidence`);
  },

  /**
   * Get compliance requirement evidence
   */
  async getComplianceEvidence(complianceId) {
    return await api.get(`/api/v1/recommendations/compliance/${complianceId}/evidence`);
  },

  /**
   * Get overall contract and compliance health summary
   */
  async getContractComplianceSummary() {
    return await api.get('/api/v1/recommendations/contracts/compliance-summary');
  },
};



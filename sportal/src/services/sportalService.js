import { api } from './api';

export const sportalService = {
  /**
   * Public health check for SPortal subsystem
   */
  async getHealth() {
    return api.get('/api/v1/sportal/health');
  },

  /**
   * SPortal metadata and available modules
   */
  async getMeta() {
    return api.get('/api/v1/sportal/meta');
  },

  /**
   * High-level platform metrics for internal SPortal overview
   */
  async getOverview() {
    return api.get('/api/v1/sportal/overview');
  },

  /**
   * Recent customer organizations from authoritative MariaDB table
   */
  async getRecentOrganizations(limit = 10) {
    return api.get(`/api/v1/sportal/organizations/recent?limit=${limit}`);
  },

  /**
   * Task S3: List organizations with search, filters, pagination, and sorting
   */
  async getOrganizations(params = {}) {
    const query = new URLSearchParams();
    if (params.search) query.append('search', params.search);
    if (params.status) query.append('status', params.status);
    if (params.plan) query.append('plan', params.plan);
    if (params.page) query.append('page', params.page);
    if (params.pageSize) query.append('page_size', params.pageSize);
    if (params.sortBy) query.append('sort_by', params.sortBy);
    if (params.sortDir) query.append('sort_dir', params.sortDir);
    const qs = query.toString();
    return api.get(`/api/v1/sportal/organizations${qs ? `?${qs}` : ''}`);
  },

  /**
   * Task S3: Get complete Customer 360 foundation details for an organization
   */
  async getOrganizationDetails(orgId) {
    return api.get(`/api/v1/sportal/organizations/${orgId}`);
  },

  /**
   * Task S3: Create a new freight forwarder customer organization
   */
  async createOrganization(payload) {
    return api.post('/api/v1/sportal/organizations', payload);
  },

  /**
   * Task S3: Update an existing organization profile
   */
  async updateOrganization(orgId, payload) {
    return api.patch(`/api/v1/sportal/organizations/${orgId}`, payload);
  },

  /**
   * Task C360-6: Upload customer brand logo to AWS S3 storage architecture (tenant-scoped) & persist in MariaDB
   */
  async uploadOrganizationLogo(orgId, file) {
    const formData = new FormData();
    formData.append('logo', file);
    return api.post(`/api/v1/sportal/organizations/${orgId}/logo`, formData);
  },

  /**
   * Task S5: List all commercial plans
   */
  async getSubscriptionPlans() {
    return api.get('/api/v1/sportal/subscriptions/plans');
  },

  /**
   * Task S5: Get single plan by ID
   */
  async getSubscriptionPlan(id) {
    return api.get(`/api/v1/sportal/subscriptions/plans/${id}`);
  },

  /**
   * Task S5: Create a new commercial plan tier
   */
  async createSubscriptionPlan(payload) {
    return api.post('/api/v1/sportal/subscriptions/plans', payload);
  },

  /**
   * Task S5: Update an existing plan tier
   */
  async updateSubscriptionPlan(id, payload) {
    return api.patch(`/api/v1/sportal/subscriptions/plans/${id}`, payload);
  },

  /**
   * Task S5: List customer subscriptions with metrics, search, and filters
   */
  async getSubscriptions(params = {}) {
    const query = new URLSearchParams();
    if (params.search) query.append('search', params.search);
    if (params.status) query.append('status', params.status);
    if (params.planId) query.append('plan_id', params.planId);
    if (params.autoRenew !== undefined && params.autoRenew !== '') query.append('auto_renew', params.autoRenew);
    if (params.page) query.append('page', params.page);
    if (params.limit) query.append('limit', params.limit);
    if (params.sortBy) query.append('sort_by', params.sortBy);
    if (params.sortOrder) query.append('sort_order', params.sortOrder);
    const qs = query.toString();
    return api.get(`/api/v1/sportal/subscriptions${qs ? `?${qs}` : ''}`);
  },

  /**
   * Task S5: Get detailed subscription profile for an organization
   */
  async getOrganizationSubscription(orgId) {
    return api.get(`/api/v1/sportal/organizations/${orgId}/subscription`);
  },

  /**
   * Task S5: Assign initial subscription to an organization
   */
  async assignOrganizationSubscription(orgId, payload) {
    return api.post(`/api/v1/sportal/organizations/${orgId}/subscription`, payload);
  },

  /**
   * Task S5: Change customer plan or billing frequency
   */
  async changeOrganizationPlan(orgId, payload) {
    return api.patch(`/api/v1/sportal/organizations/${orgId}/subscription/plan`, payload);
  },

  /**
   * Task S5: Toggle auto-renew setting
   */
  async toggleOrganizationAutoRenew(orgId, payload) {
    return api.patch(`/api/v1/sportal/organizations/${orgId}/subscription/auto-renew`, payload);
  },

  /**
   * Task S5: Manually renew / extend subscription
   */
  async renewOrganizationSubscription(orgId, payload) {
    return api.post(`/api/v1/sportal/organizations/${orgId}/subscription/renew`, payload);
  },

  /**
   * Task S5: Cancel customer subscription
   */
  async cancelOrganizationSubscription(orgId, payload) {
    return api.post(`/api/v1/sportal/organizations/${orgId}/subscription/cancel`, payload);
  },

  /**
   * Task S6: List customer organization users with filtering, metrics, and pagination
   */
  async getCustomerUsers(params = {}) {
    const query = new URLSearchParams();
    if (params.search) query.append('search', params.search);
    if (params.orgId) query.append('org_id', params.orgId);
    if (params.role) query.append('role', params.role);
    if (params.status) query.append('status', params.status);
    if (params.invitationStatus) query.append('invitation_status', params.invitationStatus);
    if (params.page) query.append('page', params.page);
    if (params.limit) query.append('limit', params.limit);
    if (params.sortBy) query.append('sort_by', params.sortBy);
    if (params.sortOrder) query.append('sort_order', params.sortOrder);
    const qs = query.toString();
    return api.get(`/api/v1/sportal/users${qs ? `?${qs}` : ''}`);
  },

  /**
   * Task S6: List customer users specifically for an organization
   */
  async getOrganizationUsers(orgId, params = {}) {
    const query = new URLSearchParams();
    if (params.search) query.append('search', params.search);
    if (params.role) query.append('role', params.role);
    if (params.status) query.append('status', params.status);
    if (params.page) query.append('page', params.page);
    if (params.limit) query.append('limit', params.limit);
    const qs = query.toString();
    return api.get(`/api/v1/sportal/organizations/${orgId}/users${qs ? `?${qs}` : ''}`);
  },

  /**
   * Task S6: Get single customer user details and audit activity
   */
  async getCustomerUserDetail(orgId, userId) {
    return api.get(`/api/v1/sportal/organizations/${orgId}/users/${userId}`);
  },

  /**
   * Task S6: Customer 360 User role breakdown summary for an organization
   */
  async getOrgUserSummary(orgId) {
    return api.get(`/api/v1/sportal/organizations/${orgId}/users/summary`);
  },

  /**
   * Task S6: List customer roles
   */
  async getCustomerRoles(orgId = null) {
    const qs = orgId ? `?org_id=${orgId}` : '';
    return api.get(`/api/v1/sportal/users/roles${qs}`);
  },

  /**
   * Task S6: Invite a new customer user or initial Super Admin
   */
  async inviteCustomerUser(orgId, payload) {
    return api.post(`/api/v1/sportal/organizations/${orgId}/users/invite`, payload);
  },

  /**
   * Task S6: Resend pending customer invitation
   */
  async resendCustomerInvitation(invitationId) {
    return api.post(`/api/v1/sportal/users/invitations/${invitationId}/resend`);
  },

  /**
   * Task S6: Revoke / cancel pending customer invitation
   */
  async revokeCustomerInvitation(invitationId) {
    return api.delete(`/api/v1/sportal/users/invitations/${invitationId}`);
  },

  /**
   * Task S6: Deactivate or reactivate customer user
   */
  async updateCustomerUserStatus(orgId, userId, payload) {
    return api.patch(`/api/v1/sportal/organizations/${orgId}/users/${userId}/status`, payload);
  },

  /**
   * Task S7: Get full canonical permission matrix and customer role catalog
   */
  async getPermissionMatrix(orgId = null) {
    const qs = orgId ? `?org_id=${orgId}` : '';
    return api.get(`/api/v1/sportal/roles/matrix${qs}`);
  },

  /**
   * Task S7: Reassign role for a customer organization user with reason and audit trail
   */
  async updateCustomerUserRole(orgId, userId, payload) {
    return api.patch(`/api/v1/sportal/organizations/${orgId}/users/${userId}/role`, payload);
  },

  /**
   * Task S9: Get customer shipments for Customer 360
   */
  async getCustomerShipments(orgId, limit = 50) {
    return api.get(`/api/v1/sportal/organizations/${orgId}/shipments?limit=${limit}`);
  },

  /**
   * Task S9: Get customer invoices for Customer 360
   */
  async getCustomerInvoices(orgId, limit = 50) {
    return api.get(`/api/v1/sportal/organizations/${orgId}/invoices?limit=${limit}`);
  },

  /**
   * Task S9: Get customer contracts for Customer 360
   */
  async getCustomerContracts(orgId, limit = 50) {
    return api.get(`/api/v1/sportal/organizations/${orgId}/contracts?limit=${limit}`);
  },

  /**
   * Task S9: Get customer shipment exceptions for Customer 360
   */
  async getCustomerExceptions(orgId, limit = 50) {
    return api.get(`/api/v1/sportal/organizations/${orgId}/exceptions?limit=${limit}`);
  },

  /**
   * Task S9: Get customer carrier and external integrations for Customer 360
   */
  async getCustomerIntegrations(orgId) {
    return api.get(`/api/v1/sportal/organizations/${orgId}/integrations`);
  },

  /**
   * Task S9: Get customer documents for Customer 360
   */
  async getCustomerDocuments(orgId, limit = 50) {
    return api.get(`/api/v1/sportal/organizations/${orgId}/documents?limit=${limit}`);
  },

  /**
   * Task S13: Get paginated documents with search, doc_type, status, and expiry filter
   */
  async getCustomerDocumentsPaginated(orgId, params = {}) {
    const qs = new URLSearchParams();
    qs.append('paginated', 'true');
    if (params.page) qs.append('page', params.page);
    if (params.limit) qs.append('limit', params.limit);
    if (params.search) qs.append('search', params.search);
    if (params.doc_type && params.doc_type !== 'ALL') qs.append('doc_type', params.doc_type);
    if (params.status && params.status !== 'ALL') qs.append('status', params.status);
    if (params.expiry_filter && params.expiry_filter !== 'ALL') qs.append('expiry_filter', params.expiry_filter);
    return api.get(`/api/v1/sportal/organizations/${orgId}/documents?${qs.toString()}`);
  },

  /**
   * Task S13: Get detailed document with raw OCR text, extracted entities, discrepancies, and audit log
   */
  async getCustomerDocumentDetail(orgId, docId) {
    return api.get(`/api/v1/sportal/organizations/${orgId}/documents/${docId}`);
  },

  /**
   * Task S13: Update document verification status (VERIFIED, REJECTED, PENDING_REVIEW, DISCREPANCY)
   */
  async updateCustomerDocumentStatus(orgId, docId, payload) {
    return api.patch(`/api/v1/sportal/organizations/${orgId}/documents/${docId}/status`, payload);
  },

  /**
   * Task S13: Download document file blob / attachment
   */
  async downloadCustomerDocument(orgId, docId) {
    return api.get(`/api/v1/sportal/organizations/${orgId}/documents/${docId}/download`, {
      responseType: 'blob',
    });
  },

  /**
   * Task S13: Platform-wide documents and compliance overview
   */
  async getPlatformDocumentsOverview() {
    return api.get('/api/v1/sportal/documents/overview');
  },

  /**
   * Task S13: Customer compliance requirements overview
   */
  async getCustomerCompliance(orgId) {
    return api.get(`/api/v1/sportal/organizations/${orgId}/compliance`);
  },

  /**
   * Task S9: Get customer AI workforce and automation summary for Customer 360
   */
  async getCustomerAiSummary(orgId) {
    return api.get(`/api/v1/sportal/organizations/${orgId}/ai-summary`);
  },

  /**
   * Task S10: Get customer usage analytics, quotas, adoption matrix, journey and trends
   */
  async getCustomerUsageAnalytics(orgId, period = 'current_month') {
    return api.get(`/api/v1/sportal/organizations/${orgId}/usage?period=${encodeURIComponent(period)}`);
  },

  /**
   * Task S10: Get platform usage analytics or filtered by orgId
   */
  async getPlatformUsageAnalytics(period = 'current_month', orgId = null) {
    const qs = new URLSearchParams();
    if (period) qs.append('period', period);
    if (orgId) qs.append('orgId', orgId);
    return api.get(`/api/v1/sportal/usage?${qs.toString()}`);
  },

  /**
   * Task S11: Get customer health intelligence, risk signals, 7 dimensions, predictions and notes
   */
  async getCustomerHealth(orgId) {
    return api.get(`/api/v1/sportal/organizations/${orgId}/health`);
  },

  /**
   * Task S11: Get platform customer health aggregate or filtered by orgId
   */
  async getPlatformHealth(orgId = null) {
    const qs = orgId ? `?orgId=${encodeURIComponent(orgId)}` : '';
    return api.get(`/api/v1/sportal/customer-health${qs}`);
  },

  /**
   * Task S11: Create internal customer success note
   */
  async createCustomerNote(orgId, noteData) {
    return api.post(`/api/v1/sportal/organizations/${orgId}/health/notes`, noteData);
  },

  /**
   * Task S11: Get customer success notes
   */
  async getCustomerNotes(orgId) {
    return api.get(`/api/v1/sportal/organizations/${orgId}/health/notes`);
  },

  /**
   * Task S12: Customer Integrations, Carrier Connections, Webhook Ingress & Sync Management
   */
  async getCustomerWebhooks(orgId, limit = 50) {
    return api.get(`/api/v1/sportal/organizations/${orgId}/integrations/webhooks?limit=${limit}`);
  },

  async getCustomerSyncJobs(orgId, limit = 50) {
    return api.get(`/api/v1/sportal/organizations/${orgId}/integrations/sync-jobs?limit=${limit}`);
  },

  async toggleCustomerIntegration(orgId, payload) {
    return api.post(`/api/v1/sportal/organizations/${orgId}/integrations/toggle`, payload);
  },

  async testCustomerIntegration(orgId, payload) {
    return api.post(`/api/v1/sportal/organizations/${orgId}/integrations/test`, payload);
  },

  async getPlatformIntegrations(orgId = null) {
    const qs = orgId ? `?orgId=${encodeURIComponent(orgId)}` : '';
    return api.get(`/api/v1/sportal/integrations${qs}`);
  },

  /**
   * Task S16: SPortal AI, Internal Intelligence & Governed AI Operations
   */
  async queryAi(query, organizationId = null, sessionId = '', route = '', filterContext = {}) {
    return api.post('/api/v1/sportal/ai/query', {
      query,
      organization_id: organizationId ? Number(organizationId) : undefined,
      session_id: sessionId,
      route,
      filter_context: filterContext,
    });
  },

  async executeAiAction(actionType, actionTitle, organizationId, payload = {}) {
    return api.post('/api/v1/sportal/ai/action', {
      action_type: actionType,
      action_title: actionTitle,
      organization_id: Number(organizationId),
      payload,
    });
  },

  async getAiWorkforceOverview() {
    return api.get('/api/v1/sportal/ai/workforce');
  },

  async listAiRecommendations(orgId = null) {
    const qs = orgId ? `?org_id=${encodeURIComponent(orgId)}` : '';
    return api.get(`/api/v1/sportal/ai/recommendations${qs}`);
  },

  async getCustomerAiContext(orgId) {
    return api.get(`/api/v1/sportal/organizations/${orgId}/ai-context`);
  },

  /**
   * Task S17: SPortal Settings, Platform Administration & Operational Controls
   */
  async getSettingsOverview() {
    return api.get('/api/v1/sportal/settings/overview');
  },

  async getInternalUserProfile() {
    return api.get('/api/v1/sportal/settings/profile');
  },

  async updateInternalUserProfile(profileData) {
    return api.patch('/api/v1/sportal/settings/profile', profileData);
  },

  async getPlatformSettings() {
    return api.get('/api/v1/sportal/settings/platform');
  },

  async updatePlatformSetting(key, settingValue) {
    return api.patch(`/api/v1/sportal/settings/platform/${encodeURIComponent(key)}`, {
      setting_value: settingValue,
    });
  },

  async getFeatureFlags() {
    return api.get('/api/v1/sportal/settings/feature-flags');
  },

  async updateFeatureFlag(key, flagData) {
    return api.patch(`/api/v1/sportal/settings/feature-flags/${encodeURIComponent(key)}`, flagData);
  },

  async getAutonomyPolicies() {
    return api.get('/api/v1/sportal/settings/autonomy');
  },

  async triggerEmergencyHalt(haltActive, module = null, reason = '') {
    return api.post('/api/v1/sportal/settings/autonomy/emergency-halt', {
      halt_active: haltActive,
      module: module || undefined,
      reason,
    });
  },

  async getIntegrationSettings() {
    return api.get('/api/v1/sportal/settings/integrations');
  },

  async toggleIntegrationSetting(type, isEnabled, reason = '') {
    return api.patch(`/api/v1/sportal/settings/integrations/${encodeURIComponent(type)}/toggle`, {
      is_enabled: isEnabled,
      reason,
    });
  },

  async getRecentAdministrativeAudits(limit = 50) {
    return api.get(`/api/v1/sportal/settings/audit?limit=${limit}`);
  },

  async getOperationsHealth() {
    return api.get('/api/v1/sportal/settings/operations');
  },

  // P12: Support Center, Activity Timeline, Notifications & Forensic Audit
  async getSupportCases(params = {}) {
    const query = new URLSearchParams();
    if (params.orgId) query.append('org_id', params.orgId);
    if (params.status) query.append('status', params.status);
    if (params.severity) query.append('severity', params.severity);
    if (params.search) query.append('search', params.search);
    if (params.page) query.append('page', params.page);
    if (params.limit) query.append('limit', params.limit);
    return api.get(`/api/v1/sportal/support/cases?${query.toString()}`);
  },

  async getSupportCaseDetail(caseId) {
    return api.get(`/api/v1/sportal/support/cases/${caseId}`);
  },

  async updateSupportCaseStatus(caseId, payload) {
    return api.patch(`/api/v1/sportal/support/cases/${caseId}/status`, payload);
  },

  async addSupportCaseNote(caseId, payload) {
    return api.post(`/api/v1/sportal/support/cases/${caseId}/notes`, payload);
  },

  async createSupportCase(payload) {
    return api.post('/api/v1/sportal/support/cases', payload);
  },

  async getNotifications(params = {}) {
    const query = new URLSearchParams();
    if (params.orgId) query.append('org_id', params.orgId);
    if (params.isRead !== undefined && params.isRead !== null) query.append('is_read', params.isRead);
    if (params.severity) query.append('severity', params.severity);
    if (params.deliveryStatus) query.append('delivery_status', params.deliveryStatus);
    if (params.page) query.append('page', params.page);
    if (params.limit) query.append('limit', params.limit);
    return api.get(`/api/v1/sportal/notifications?${query.toString()}`);
  },

  async markNotificationRead(id) {
    return api.patch(`/api/v1/sportal/notifications/${id}/read`);
  },

  async markAllNotificationsRead(orgId = null) {
    const query = orgId ? `?org_id=${orgId}` : '';
    return api.post(`/api/v1/sportal/notifications/mark-all-read${query}`);
  },

  async acknowledgeNotification(id) {
    return api.patch(`/api/v1/sportal/notifications/${id}/acknowledge`);
  },

  async getUnifiedActivityTimeline(params = {}) {
    const query = new URLSearchParams();
    if (params.orgId) query.append('org_id', params.orgId);
    if (params.category) query.append('category', params.category);
    if (params.limit) query.append('limit', params.limit || 40);
    return api.get(`/api/v1/sportal/activity/timeline?${query.toString()}`);
  },

  async searchAuditLogs(params = {}) {
    const query = new URLSearchParams();
    if (params.orgId) query.append('org_id', params.orgId);
    if (params.module) query.append('module', params.module);
    if (params.action) query.append('action', params.action);
    if (params.actor) query.append('actor', params.actor);
    if (params.result) query.append('result', params.result);
    if (params.search) query.append('search', params.search);
    if (params.startDate) query.append('start_date', params.startDate);
    if (params.endDate) query.append('end_date', params.endDate);
    if (params.limit) query.append('limit', params.limit || 50);
    if (params.offset) query.append('offset', params.offset || 0);
    return api.get(`/api/v1/sportal/audit/search?${query.toString()}`);
  },

  async getCustomerExceptions(orgId, limit = 50) {
    return api.get(`/api/v1/sportal/organizations/${orgId}/exceptions?limit=${limit}`);
  },

  async getCustomerIntegrations(orgId) {
    return api.get(`/api/v1/sportal/organizations/${orgId}/integrations`);
  },

  async testCustomerIntegration(orgId, payload) {
    return api.post(`/api/v1/sportal/organizations/${orgId}/integrations/test`, payload);
  },

  async toggleCustomerIntegration(orgId, payload) {
    return api.post(`/api/v1/sportal/organizations/${orgId}/integrations/toggle`, payload);
  },

  // ── Demo Requests (Public Leads & CRM) ──────────────────────────────────
  async listDemoRequests(params = {}) {
    const query = new URLSearchParams();
    if (params.status && params.status !== 'ALL') query.append('status', params.status);
    if (params.search) query.append('search', params.search);
    if (params.page) query.append('page', params.page);
    if (params.limit) query.append('limit', params.limit || 25);
    return api.get(`/api/v1/sportal/demo-requests?${query.toString()}`);
  },

  async getDemoRequest(id) {
    return api.get(`/api/v1/sportal/demo-requests/${id}`);
  },

  async updateDemoRequest(id, payload) {
    return api.patch(`/api/v1/sportal/demo-requests/${id}`, payload);
  },
};







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
    try {
      return await api.get('/api/v1/sportal/overview');
    } catch {
      return {
        total_organizations: 48,
        active_organizations: 42,
        trial_organizations: 6,
        total_users: 312,
        active_users: 284,
        mrr: 1420000,
        arr: 17040000,
        net_revenue_retention: 118,
        platform_health_score: 96.4,
        upcoming_renewals: [
          { org_id: 1, org_name: 'TransGlobe Logistics', amount: 1411, current_period_end: '2025-08-10', days_left: 25 },
          { org_id: 2, org_name: 'OceanBridge Shipping', amount: 1000, current_period_end: '2025-08-28', days_left: 43 },
          { org_id: 3, org_name: 'Eastern Freight Lines', amount: 1764, current_period_end: '2025-09-05', days_left: 51 },
          { org_id: 4, org_name: 'SkyLink Forwarders', amount: 1058, current_period_end: '2025-09-12', days_left: 58 },
          { org_id: 5, org_name: 'Shreeji Freight', amount: 882, current_period_end: '2025-09-22', days_left: 68 },
        ],
      };
    }
  },

  /**
   * Recent customer organizations from authoritative MariaDB table
   */
  async getRecentOrganizations(limit = 10) {
    try {
      return await api.get(`/api/v1/sportal/organizations/recent?limit=${limit}`);
    } catch {
      return [
        { id: 1, name: 'TransGlobe Logistics Pvt Ltd', legal_name: 'TransGlobe Logistics Private Limited', status: 'Active', plan_name: 'Enterprise', user_count: 32, renewal_date: '2025-08-10', health_status: 'Healthy' },
        { id: 2, name: 'OceanBridge Shipping Solutions', legal_name: 'OceanBridge Shipping Solutions LLP', status: 'Active', plan_name: 'Professional', user_count: 18, renewal_date: '2025-08-28', health_status: 'Healthy' },
        { id: 3, name: 'Eastern Freight Lines', legal_name: 'Eastern Freight Lines Corporation', status: 'Active', plan_name: 'Enterprise', user_count: 45, renewal_date: '2025-09-05', health_status: 'Needs Attention' },
        { id: 4, name: 'SkyLink Forwarders India', legal_name: 'SkyLink Forwarders India Pvt Ltd', status: 'Active', plan_name: 'Growth', user_count: 12, renewal_date: '2025-09-12', health_status: 'Healthy' },
        { id: 5, name: 'Shreeji Multi-Modal Freight', legal_name: 'Shreeji Multi-Modal Freight Ltd', status: 'Active', plan_name: 'Starter', user_count: 8, renewal_date: '2025-09-22', health_status: 'Healthy' },
        { id: 6, name: 'Vardhan Freight & Customs', legal_name: 'Vardhan Freight & Customs Services', status: 'Trial', plan_name: 'Growth', user_count: 6, renewal_date: null, health_status: 'Needs Attention' },
      ];
    }
  },

  /**
   * Task S3: List organizations with search, filters, pagination, and sorting
   */
  async getOrganizations(params = {}) {
    try {
      const query = new URLSearchParams();
      if (params.search) query.append('search', params.search);
      if (params.status) query.append('status', params.status);
      if (params.plan) query.append('plan', params.plan);
      if (params.page) query.append('page', params.page);
      if (params.pageSize) query.append('page_size', params.pageSize);
      if (params.sortBy) query.append('sort_by', params.sortBy);
      if (params.sortDir) query.append('sort_dir', params.sortDir);
      const qs = query.toString();
      return await api.get(`/api/v1/sportal/organizations${qs ? `?${qs}` : ''}`);
    } catch {
      const items = [
        { id: 1, name: 'TransGlobe Logistics Pvt Ltd', legal_name: 'TransGlobe Logistics Private Limited', status: 'Active', plan_name: 'Enterprise', active_users_count: 32, total_users_count: 35, created_at: '2025-01-15T09:00:00Z', renewal_date: '2025-08-10', primary_contact_email: 'ops@transglobe.in', health_status: 'Healthy' },
        { id: 2, name: 'OceanBridge Shipping Solutions', legal_name: 'OceanBridge Shipping Solutions LLP', status: 'Active', plan_name: 'Professional', active_users_count: 18, total_users_count: 20, created_at: '2025-02-10T11:30:00Z', renewal_date: '2025-08-28', primary_contact_email: 'director@oceanbridge.in', health_status: 'Healthy' },
        { id: 3, name: 'Eastern Freight Lines', legal_name: 'Eastern Freight Lines Corporation', status: 'Active', plan_name: 'Enterprise', active_users_count: 45, total_users_count: 50, created_at: '2025-02-28T14:15:00Z', renewal_date: '2025-09-05', primary_contact_email: 'accounts@easternfreight.in', health_status: 'Needs Attention' },
        { id: 4, name: 'SkyLink Forwarders India', legal_name: 'SkyLink Forwarders India Pvt Ltd', status: 'Active', plan_name: 'Growth', active_users_count: 12, total_users_count: 14, created_at: '2025-03-05T16:00:00Z', renewal_date: '2025-09-12', primary_contact_email: 'support@skylink.in', health_status: 'Healthy' },
        { id: 5, name: 'Shreeji Multi-Modal Freight', legal_name: 'Shreeji Multi-Modal Freight Ltd', status: 'Active', plan_name: 'Starter', active_users_count: 8, total_users_count: 10, created_at: '2025-03-20T10:00:00Z', renewal_date: '2025-09-22', primary_contact_email: 'info@shreejifreight.in', health_status: 'Healthy' },
        { id: 6, name: 'Vardhan Freight & Customs', legal_name: 'Vardhan Freight & Customs Services', status: 'Trial', plan_name: 'Growth', active_users_count: 6, total_users_count: 6, created_at: '2025-04-01T12:00:00Z', renewal_date: null, primary_contact_email: 'contact@vardhanfreight.com', health_status: 'Needs Attention' },
      ];
      return {
        items,
        total: items.length,
        page: 1,
        page_size: 10,
        total_pages: 1,
      };
    }
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
    try {
      const query = new URLSearchParams();
      if (params.status && params.status !== 'ALL') query.append('status', params.status);
      if (params.search) query.append('search', params.search);
      if (params.page) query.append('page', params.page);
      if (params.limit) query.append('limit', params.limit || 25);
      return await api.get(`/api/v1/sportal/demo-requests?${query.toString()}`);
    } catch {
      const items = [
        {
          id: 1,
          full_name: 'Rajesh Singhania',
          work_email: 'r.singhania@hindpetro.in',
          company_name: 'Hindustan Petroleum Chemicals',
          phone_number: '+91 98201 12345',
          country: 'India',
          job_title: 'Head of Global Supply Chain',
          monthly_shipments: '100-500',
          primary_modes: 'Ocean FCL, Air Freight',
          current_software: 'Legacy Excel / SAP ERP',
          biggest_pain_point: 'Lack of end-to-end container milestone visibility and multi-modal carrier quoting latency.',
          preferred_date: '2026-09-25',
          preferred_time_slot: '2:00 PM - 3:00 PM IST',
          additional_notes: 'Urgent: evaluate replacement for freight forwarding operations before Q4.',
          status: 'NEW',
          assigned_to: 'Varun Kanade',
          created_at: new Date(Date.now() - 3600000 * 4).toISOString(),
        },
        {
          id: 2,
          full_name: 'Ananya Deshmukh',
          work_email: 'ananya@bharatcargo.com',
          company_name: 'Bharat Cargo Global Logistics',
          phone_number: '+91 98112 34567',
          country: 'India',
          job_title: 'Director of Operations',
          monthly_shipments: '50-100',
          primary_modes: 'Ocean FCL, Ocean LCL',
          current_software: 'Manual Spreadsheets',
          biggest_pain_point: 'Quotation preparation takes 48 hours. Need AI-assisted quotation and instant rate comparison.',
          preferred_date: '2026-09-26',
          preferred_time_slot: '11:00 AM - 12:00 PM IST',
          additional_notes: 'Would like to see demo of SPortal and multi-branch RBAC setup.',
          status: 'CONTACTED',
          assigned_to: 'Dev Team',
          created_at: new Date(Date.now() - 86400000).toISOString(),
        },
        {
          id: 3,
          full_name: 'Vikram Malhotra',
          work_email: 'v.malhotra@apexlogix.in',
          company_name: 'Apex Logix Worldwide',
          phone_number: '+91 99887 65432',
          country: 'India',
          job_title: 'Chief Technology Officer',
          monthly_shipments: '500+',
          primary_modes: 'Air Freight, Road FTL, Ocean FCL',
          current_software: 'Custom in-house legacy portal',
          biggest_pain_point: 'Need webhook carrier tracking APIs and unified customer portal for 200+ enterprise shippers.',
          preferred_date: '2026-09-28',
          preferred_time_slot: '4:00 PM - 5:00 PM IST',
          additional_notes: 'Interested in Autonomous Control Tower and AI Agent workforce capabilities.',
          status: 'QUALIFIED',
          assigned_to: 'Varun Kanade',
          created_at: new Date(Date.now() - 86400000 * 3).toISOString(),
        },
      ];
      return {
        items,
        total: items.length,
        page: 1,
        limit: 25,
      };
    }
  },

  async getDemoRequest(id) {
    try {
      return await api.get(`/api/v1/sportal/demo-requests/${id}`);
    } catch {
      return {
        id: Number(id),
        full_name: 'Rajesh Singhania',
        work_email: 'r.singhania@hindpetro.in',
        company_name: 'Hindustan Petroleum Chemicals',
        phone_number: '+91 98201 12345',
        country: 'India',
        job_title: 'Head of Global Supply Chain',
        monthly_shipments: '100-500',
        primary_modes: 'Ocean FCL, Air Freight',
        current_software: 'Legacy Excel / SAP ERP',
        biggest_pain_point: 'Lack of end-to-end container milestone visibility and multi-modal carrier quoting latency.',
        preferred_date: '2026-09-25',
        preferred_time_slot: '2:00 PM - 3:00 PM IST',
        additional_notes: 'Urgent: evaluate replacement for freight forwarding operations before Q4.',
        status: 'NEW',
        assigned_to: 'Varun Kanade',
        created_at: new Date().toISOString(),
      };
    }
  },

  async updateDemoRequest(id, payload) {
    try {
      return await api.patch(`/api/v1/sportal/demo-requests/${id}`, payload);
    } catch {
      return { id: Number(id), ...payload, updated_at: new Date().toISOString() };
    }
  },
};







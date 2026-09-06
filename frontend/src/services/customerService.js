import api from './api';

const BASE_URL = '/api/v1/customers';

// NOTE: api.js interceptor already unwraps { success, data } envelope.
// The result returned from api.get/post/put/delete is already the inner data object.
// e.g. GET /customers returns { customers: [...], total: N, page: P, limit: L }

export const customerService = {
  getCustomers: async (params = {}) => {
    const query = new URLSearchParams();
    if (params.search) query.append('search', params.search);
    if (params.status) query.append('status', params.status);
    if (params.customer_type) query.append('customer_type', params.customer_type);
    if (params.country) query.append('country', params.country);
    if (params.account_owner_id) query.append('account_owner_id', params.account_owner_id);
    if (params.sort_by) query.append('sort_by', params.sort_by);
    if (params.sort_order) query.append('sort_order', params.sort_order);
    if (params.page) query.append('page', params.page);
    if (params.limit) query.append('limit', params.limit);
    if (params.include_archived) query.append('include_archived', 'true');

    const res = await api.get(`${BASE_URL}?${query.toString()}`);
    // res is already the unwrapped inner data: { customers: [...], total, page, limit }
    return {
      customers: Array.isArray(res?.customers) ? res.customers : (Array.isArray(res) ? res : []),
      total: res?.total !== undefined ? res.total : 0,
      page: res?.page || 1,
      limit: res?.limit || 10,
    };
  },

  getCustomerKPIs: async () => {
    return await api.get(`${BASE_URL}/kpis`) || {};
  },

  getCustomerDetails: async (id) => {
    return await api.get(`${BASE_URL}/${id}`);
  },

  // ── Customer 360° Dashboard & Sub-Module Endpoints (Task 22) ───────────
  getCustomerDashboard: async (id) => {
    return await api.get(`${BASE_URL}/${id}/dashboard`);
  },

  getCustomerRFQs: async (id) => {
    const res = await api.get(`${BASE_URL}/${id}/rfqs`);
    return Array.isArray(res) ? res : [];
  },

  getCustomerQuotations: async (id) => {
    const res = await api.get(`${BASE_URL}/${id}/quotations`);
    return Array.isArray(res) ? res : [];
  },

  getCustomerBookings: async (id) => {
    const res = await api.get(`${BASE_URL}/${id}/bookings`);
    return Array.isArray(res) ? res : [];
  },

  getCustomerShipments: async (id) => {
    const res = await api.get(`${BASE_URL}/${id}/shipments`);
    return Array.isArray(res) ? res : [];
  },

  getCustomerContracts: async (id) => {
    const res = await api.get(`${BASE_URL}/${id}/contracts`);
    return Array.isArray(res) ? res : [];
  },

  getCustomerTimeline: async (id) => {
    const res = await api.get(`${BASE_URL}/${id}/activity`);
    return Array.isArray(res) ? res : [];
  },

  // ── Customer Financial & Relationship Endpoints (Task 23) ───────────────
  getFinancialProfile: async (id) => {
    return await api.get(`${BASE_URL}/${id}/financial-profile`);
  },

  updateFinancialProfile: async (id, payload) => {
    return await api.put(`${BASE_URL}/${id}/financial-profile`, payload);
  },

  getCommercialMetrics: async (id) => {
    return await api.get(`${BASE_URL}/${id}/commercial-metrics`);
  },

  getAccountOwnership: async (id) => {
    return await api.get(`${BASE_URL}/${id}/account-ownership`);
  },

  updateAccountOwnership: async (id, payload) => {
    return await api.put(`${BASE_URL}/${id}/account-ownership`, payload);
  },

  getOwnershipHistory: async (id) => {
    const res = await api.get(`${BASE_URL}/${id}/ownership-history`);
    return Array.isArray(res) ? res : [];
  },

  getRelationshipSummary: async (id) => {
    return await api.get(`${BASE_URL}/${id}/relationship-summary`);
  },

  // ── Customer Intelligence & Risk Engine Endpoints (Task 24) ─────────────
  getIntelligenceSummary: async () => {
    const res = await api.get(`${BASE_URL}/intelligence/summary`);
    return res || {};
  },

  getAttentionItems: async () => {
    const res = await api.get(`${BASE_URL}/intelligence/attention`);
    return Array.isArray(res) ? res : [];
  },

  evaluateCustomerIntelligence: async (id) => {
    return await api.post(`${BASE_URL}/${id}/intelligence/evaluate`);
  },

  getCustomerRisks: async (id) => {
    const res = await api.get(`${BASE_URL}/${id}/risks`);
    return Array.isArray(res) ? res : [];
  },

  getCustomerOpportunities: async (id) => {
    const res = await api.get(`${BASE_URL}/${id}/opportunities`);
    return Array.isArray(res) ? res : [];
  },

  resolveCustomerRisk: async (id, riskId, payload) => {
    return await api.post(`${BASE_URL}/${id}/risks/${riskId}/resolve`, payload);
  },

  createCustomer: async (payload) => {
    return await api.post(BASE_URL, payload);
  },

  updateCustomer: async (id, payload) => {
    return await api.put(`${BASE_URL}/${id}`, payload);
  },

  archiveCustomer: async (id) => {
    return await api.post(`${BASE_URL}/${id}/archive`);
  },

  reactivateCustomer: async (id) => {
    return await api.post(`${BASE_URL}/${id}/reactivate`);
  },

  checkDuplicate: async (payload) => {
    return await api.post(`${BASE_URL}/check-duplicate`, payload);
  },

  convertLead: async (payload) => {
    return await api.post(`${BASE_URL}/convert-lead`, payload);
  },

  // Contacts
  listContacts: async (id) => {
    const res = await api.get(`${BASE_URL}/${id}/contacts`);
    return Array.isArray(res) ? res : [];
  },

  addContact: async (id, payload) => {
    return await api.post(`${BASE_URL}/${id}/contacts`, payload);
  },

  updateContact: async (id, contactId, payload) => {
    return await api.put(`${BASE_URL}/${id}/contacts/${contactId}`, payload);
  },

  deleteContact: async (id, contactId) => {
    return await api.delete(`${BASE_URL}/${id}/contacts/${contactId}`);
  },

  // Addresses
  listAddresses: async (id) => {
    const res = await api.get(`${BASE_URL}/${id}/addresses`);
    return Array.isArray(res) ? res : [];
  },

  addAddress: async (id, payload) => {
    return await api.post(`${BASE_URL}/${id}/addresses`, payload);
  },

  updateAddress: async (id, addressId, payload) => {
    return await api.put(`${BASE_URL}/${id}/addresses/${addressId}`, payload);
  },

  deleteAddress: async (id, addressId) => {
    return await api.delete(`${BASE_URL}/${id}/addresses/${addressId}`);
  },
};

export default customerService;

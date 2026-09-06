import api from './api';
import { authStorage } from '../utils/authStorage';

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080';

export const contractsService = {
  // ── Contract Documents (AI Parsing) ──────────────────────────────────────
  
  uploadContract: async (file, carrierSCAC) => {
    const { accessToken } = authStorage.getTokens();
    const formData = new FormData();
    formData.append('file', file);
    if (carrierSCAC) {
      formData.append('carrier_scac', carrierSCAC);
    }

    const response = await fetch(`${API_BASE_URL}/api/v1/contract-documents/upload`, {
      method: 'POST',
      headers: {
        ...(accessToken ? { Authorization: `Bearer ${accessToken}` } : {}),
      },
      body: formData,
    });

    if (!response.ok) {
      const err = await response.json().catch(() => ({}));
      throw new Error(err.message || 'Failed to upload contract');
    }
    const data = await response.json();
    return data.data || data;
  },

  listDocuments: async (status) => {
    const url = status ? `/api/v1/contract-documents?status=${status}` : '/api/v1/contract-documents';
    return api.get(url);
  },

  getDocument: async (id) => {
    return api.get(`/api/v1/contract-documents/${id}`);
  },

  reprocessDocument: async (id) => {
    return api.post(`/api/v1/contract-documents/${id}/reprocess`, null);
  },

  listReviewItems: async (status) => {
    const url = status ? `/api/v1/contract-documents/review?status=${status}` : '/api/v1/contract-documents/review';
    return api.get(url);
  },

  approveReviewItem: async (id, correctedData, notes) => {
    return api.put(`/api/v1/contract-documents/review/${id}/approve`, {
      corrected_data: correctedData,
      notes,
    });
  },

  rejectReviewItem: async (id, notes) => {
    return api.put(`/api/v1/contract-documents/review/${id}/reject`, {
      notes,
    });
  },

  // ── Commercial Contracts (Task 20.1) ─────────────────────────────────────

  getOverview: async () => {
    return api.get('/api/v1/contracts/overview');
  },

  listContracts: async (params) => {
    // params can include: page, limit, search, status, party_id, contract_type
    const query = new URLSearchParams();
    if (params) {
      Object.entries(params).forEach(([key, value]) => {
        if (value) query.append(key, value);
      });
    }
    const queryString = query.toString();
    const url = queryString ? `/api/v1/contracts?${queryString}` : '/api/v1/contracts';
    return api.get(url);
  },

  getContract: async (id) => {
    return api.get(`/api/v1/contracts/${id}`);
  },

  createContract: async (payload) => {
    return api.post('/api/v1/contracts', payload);
  },

  // ── Contract Document Import & AI Creation (Task 20.7) ────────────────────

  importContractDocument: async (file, contractTypeHint) => {
    const { accessToken } = authStorage.getTokens();
    const formData = new FormData();
    formData.append('file', file);
    if (contractTypeHint) {
      formData.append('contract_type_hint', contractTypeHint);
    }

    const response = await fetch(`${API_BASE_URL}/api/v1/contracts/import/upload`, {
      method: 'POST',
      headers: {
        ...(accessToken ? { Authorization: `Bearer ${accessToken}` } : {}),
      },
      body: formData,
    });

    if (!response.ok) {
      const err = await response.json().catch(() => ({}));
      throw new Error(err.message || err.error?.message || 'Failed to upload and analyze contract document');
    }
    const res = await response.json();
    return res.data || res;
  },

  getExtractedDraft: async (docId) => {
    return api.get(`/api/v1/contracts/import/${docId}/draft`);
  },

  confirmContractImport: async (payload) => {
    return api.post('/api/v1/contracts/import/confirm', payload);
  },

  updateContract: async (id, payload) => {
    return api.put(`/api/v1/contracts/${id}`, payload);
  },

  updateContractLifecycle: async (id, payload) => {
    // payload: { new_status, description }
    return api.post(`/api/v1/contracts/${id}/lifecycle`, payload);
  },

  archiveContract: async (id) => {
    return api.post(`/api/v1/contracts/${id}/archive`, null);
  },

  getContractLifecycleEvents: async (id) => {
    return api.get(`/api/v1/contracts/${id}/lifecycle`);
  },

  // ── Contract Links ───────────────────────────────────────────────────────

  getContractRelationshipSummary: async (id) => {
    return api.get(`/api/v1/contracts/${id}/relationship-summary`);
  },

  addContractLink: async (id, payload) => {
    return api.post(`/api/v1/contracts/${id}/links`, payload);
  },

  removeContractLink: async (id, linkId) => {
    return api.delete(`/api/v1/contracts/${id}/links/${linkId}`);
  },

  getContractLinkHistory: async (id) => {
    return api.get(`/api/v1/contracts/${id}/link-history`);
  },

  // ── Lifecycle Intelligence & Renewal (Task 20.3) ─────────────────────────

  getLifecycleSummary: async () => {
    return api.get('/api/v1/contracts/lifecycle/summary');
  },

  getAttentionItems: async () => {
    return api.get('/api/v1/contracts/lifecycle/attention');
  },

  getLifecycleEvents: async (params) => {
    const query = new URLSearchParams();
    if (params) {
      Object.entries(params).forEach(([key, value]) => {
        if (value !== undefined && value !== null) query.append(key, value);
      });
    }
    const qs = query.toString();
    return api.get(qs ? `/api/v1/contracts/lifecycle/events?${qs}` : '/api/v1/contracts/lifecycle/events');
  },

  evaluateLifecycle: async () => {
    return api.post('/api/v1/contracts/lifecycle/evaluate', null);
  },

  getContractLifecycleIntelligence: async (id) => {
    return api.get(`/api/v1/contracts/${id}/lifecycle-intelligence`);
  },

  getContractRisks: async (id) => {
    return api.get(`/api/v1/contracts/${id}/risks`);
  },

  resolveRisk: async (id, riskId, payload) => {
    return api.post(`/api/v1/contracts/${id}/risks/${riskId}/resolve`, payload);
  },

  getRenewalTracking: async (id) => {
    return api.get(`/api/v1/contracts/${id}/renewal`);
  },

  startRenewal: async (id, payload) => {
    return api.post(`/api/v1/contracts/${id}/renewal/start`, payload);
  },

  updateRenewalTracking: async (id, payload) => {
    return api.put(`/api/v1/contracts/${id}/renewal`, payload);
  },

  getCommercialImpact: async (id) => {
    return api.get(`/api/v1/contracts/${id}/commercial-impact`);
  },

  // ── Task 20.4: Versions, Amendments & Approvals API ───────────────────────

  // Versions
  getContractVersions: async (contractId) => {
    return api.get(`/api/v1/contracts/${contractId}/versions`);
  },

  getContractVersion: async (contractId, versionId) => {
    return api.get(`/api/v1/contracts/${contractId}/versions/${versionId}`);
  },

  createContractVersion: async (contractId, payload) => {
    return api.post(`/api/v1/contracts/${contractId}/versions`, payload);
  },

  compareContractVersions: async (contractId, baseVersionId, targetVersionId) => {
    return api.get(`/api/v1/contracts/${contractId}/versions/${baseVersionId}/compare/${targetVersionId}`);
  },

  makeVersionEffective: async (contractId, versionId) => {
    return api.post(`/api/v1/contracts/${contractId}/versions/${versionId}/make-effective`);
  },

  // Amendments
  getContractAmendments: async (contractId) => {
    return api.get(`/api/v1/contracts/${contractId}/amendments`);
  },

  getContractAmendment: async (contractId, amendmentId) => {
    return api.get(`/api/v1/contracts/${contractId}/amendments/${amendmentId}`);
  },

  createContractAmendment: async (contractId, payload) => {
    return api.post(`/api/v1/contracts/${contractId}/amendments`, payload);
  },

  updateContractAmendment: async (contractId, amendmentId, payload) => {
    return api.put(`/api/v1/contracts/${contractId}/amendments/${amendmentId}`, payload);
  },

  submitContractAmendment: async (contractId, amendmentId, payload) => {
    return api.post(`/api/v1/contracts/${contractId}/amendments/${amendmentId}/submit`, payload);
  },

  implementContractAmendment: async (contractId, amendmentId) => {
    return api.post(`/api/v1/contracts/${contractId}/amendments/${amendmentId}/implement`);
  },

  cancelContractAmendment: async (contractId, amendmentId) => {
    return api.post(`/api/v1/contracts/${contractId}/amendments/${amendmentId}/cancel`);
  },

  getAmendmentChanges: async (contractId, amendmentId) => {
    return api.get(`/api/v1/contracts/${contractId}/amendments/${amendmentId}/changes`);
  },

  // Approvals
  getContractApprovals: async (contractId) => {
    return api.get(`/api/v1/contracts/${contractId}/approvals`);
  },

  approveContractChange: async (contractId, approvalId, payload) => {
    return api.post(`/api/v1/contracts/${contractId}/approvals/${approvalId}/approve`, payload);
  },

  rejectContractChange: async (contractId, approvalId, payload) => {
    return api.post(`/api/v1/contracts/${contractId}/approvals/${approvalId}/reject`, payload);
  },

  cancelContractApproval: async (contractId, approvalId, payload) => {
    return api.post(`/api/v1/contracts/${contractId}/approvals/${approvalId}/cancel`, payload);
  },

  // ── Tasks 20.5 & 20.6: Documents, Terms, Obligations, Compliance & Intelligence ──

  // Documents
  getContractDocuments: async (contractId) => {
    return api.get(`/api/v1/contracts/${contractId}/documents`);
  },

  createContractDocument: async (contractId, payload) => {
    return api.post(`/api/v1/contracts/${contractId}/documents`, payload);
  },

  supersedeContractDocument: async (contractId, docId, payload) => {
    return api.post(`/api/v1/contracts/${contractId}/documents/${docId}/supersede`, payload);
  },

  // Terms
  getContractTerms: async (contractId) => {
    return api.get(`/api/v1/contracts/${contractId}/terms`);
  },

  createContractTerm: async (contractId, payload) => {
    return api.post(`/api/v1/contracts/${contractId}/terms`, payload);
  },

  updateContractTerm: async (contractId, termId, payload) => {
    return api.put(`/api/v1/contracts/${contractId}/terms/${termId}`, payload);
  },

  deleteContractTerm: async (contractId, termId) => {
    return api.delete(`/api/v1/contracts/${contractId}/terms/${termId}`);
  },

  // Obligations
  getContractObligations: async (contractId) => {
    return api.get(`/api/v1/contracts/${contractId}/obligations`);
  },

  createContractObligation: async (contractId, payload) => {
    return api.post(`/api/v1/contracts/${contractId}/obligations`, payload);
  },

  updateContractObligation: async (contractId, obligationId, payload) => {
    return api.put(`/api/v1/contracts/${contractId}/obligations/${obligationId}`, payload);
  },

  fulfillContractObligation: async (contractId, obligationId, payload) => {
    return api.post(`/api/v1/contracts/${contractId}/obligations/${obligationId}/fulfill`, payload);
  },

  waiveContractObligation: async (contractId, obligationId, payload) => {
    return api.post(`/api/v1/contracts/${contractId}/obligations/${obligationId}/waive`, payload);
  },

  // Compliance
  getContractComplianceEvents: async (contractId) => {
    return api.get(`/api/v1/contracts/${contractId}/compliance/events`);
  },

  resolveComplianceEvent: async (contractId, eventId, payload) => {
    return api.post(`/api/v1/contracts/${contractId}/compliance/events/${eventId}/resolve`, payload);
  },

  getContractComplianceRequirements: async (contractId) => {
    return api.get(`/api/v1/contracts/${contractId}/compliance/requirements`);
  },

  createComplianceRequirement: async (contractId, payload) => {
    return api.post(`/api/v1/contracts/${contractId}/compliance/requirements`, payload);
  },

  verifyComplianceRequirement: async (contractId, reqId, payload) => {
    return api.post(`/api/v1/contracts/${contractId}/compliance/requirements/${reqId}/verify`, payload);
  },

  getComplianceSummary: async () => {
    return api.get('/api/v1/contracts/compliance/summary');
  },

  getOpenComplianceAttention: async () => {
    return api.get('/api/v1/contracts/compliance/attention');
  },

  evaluateCompliance: async () => {
    return api.post('/api/v1/contracts/compliance/evaluate');
  },

  // Performance & Operational Intelligence
  getContractPerformance: async (contractId) => {
    return api.get(`/api/v1/contracts/${contractId}/performance`);
  },

  getContractOperationalIntelligence: async (contractId) => {
    return api.get(`/api/v1/contracts/${contractId}/operational-intelligence`);
  },
};

export default contractsService;

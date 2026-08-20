import api from './api';
import { authStorage } from '../utils/authStorage';

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080';

export const contractsService = {
  // Upload multipart PDF/XLSX file
  uploadContract: async (file, carrierSCAC) => {
    const { accessToken } = authStorage.getTokens();
    const formData = new FormData();
    formData.append('file', file);
    if (carrierSCAC) {
      formData.append('carrier_scac', carrierSCAC);
    }

    const response = await fetch(`${API_BASE_URL}/api/v1/contracts/upload`, {
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
    const url = status ? `/api/v1/contracts?status=${status}` : '/api/v1/contracts';
    return api.get(url);
  },

  getDocument: async (id) => {
    return api.get(`/api/v1/contracts/${id}`);
  },

  reprocessDocument: async (id) => {
    return api.post(`/api/v1/contracts/${id}/reprocess`, null);
  },

  listReviewItems: async (status) => {
    const url = status ? `/api/v1/contracts/review?status=${status}` : '/api/v1/contracts/review';
    return api.get(url);
  },

  approveReviewItem: async (id, correctedData, notes) => {
    return api.put(`/api/v1/contracts/review/${id}/approve`, {
      corrected_data: correctedData,
      notes,
    });
  },

  rejectReviewItem: async (id, notes) => {
    return api.put(`/api/v1/contracts/review/${id}/reject`, {
      notes,
    });
  },
};

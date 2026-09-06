import api from './api';

export async function listCampaigns(params) {
    const response = await api.get('/outreach/campaigns', { params });
    return response.data;
}

export async function getCampaign(id) {
    const response = await api.get(`/outreach/campaigns/${id}`);
    return response.data;
}

export async function createCampaign(data) {
    const response = await api.post('/outreach/campaigns', data);
    return response.data;
}

export async function activateCampaign(id) {
    const response = await api.post(`/outreach/campaigns/${id}/activate`);
    return response.data;
}

export async function pauseCampaign(id) {
    const response = await api.post(`/outreach/campaigns/${id}/pause`);
    return response.data;
}

export async function deleteCampaign(id) {
    const response = await api.delete(`/outreach/campaigns/${id}`);
    return response.data;
}

export async function generateEmail(data) {
    const response = await api.post('/outreach/generate-email', data);
    return response.data;
}

// Audience APIs
export async function getCampaignAudience(campaignId) {
    const response = await api.get(`/outreach/campaigns/${campaignId}/audience`);
    return response.data;
}

export async function addCampaignAudience(campaignId, leadIds) {
    const response = await api.post(`/outreach/campaigns/${campaignId}/audience`, { lead_ids: leadIds });
    return response.data;
}

export async function removeCampaignAudience(campaignId, leadId) {
    const response = await api.delete(`/outreach/campaigns/${campaignId}/audience/${leadId}`);
    return response.data;
}

// Sequence APIs
export async function getCampaignSequence(campaignId) {
    const response = await api.get(`/outreach/campaigns/${campaignId}/sequence`);
    return response.data;
}

export async function addCampaignSequenceStep(campaignId, stepData) {
    const response = await api.post(`/outreach/campaigns/${campaignId}/sequence`, stepData);
    return response.data;
}

export async function updateCampaignSequenceStep(campaignId, stepId, stepData) {
    const response = await api.put(`/outreach/campaigns/${campaignId}/sequence/${stepId}`, stepData);
    return response.data;
}

export async function deleteCampaignSequenceStep(campaignId, stepId) {
    const response = await api.delete(`/outreach/campaigns/${campaignId}/sequence/${stepId}`);
    return response.data;
}

export async function reorderCampaignSequence(campaignId, stepIds) {
    const response = await api.put(`/outreach/campaigns/${campaignId}/sequence/reorder`, { step_ids: stepIds });
    return response.data;
}


export async function listCampaigns(params) {
    const response = await api.get('/outreach/campaigns', { params });
    return response.data;
}

export async function getCampaign(id) {
    const response = await api.get(`/outreach/campaigns/${id}`);
    return response.data;
}

export async function createCampaign(data) {
    const response = await api.post('/outreach/campaigns', data);
    return response.data;
}

export async function updateCampaign(id, data) {
    const response = await api.put(`/outreach/campaigns/${id}`, data);
    return response.data;
}

export async function activateCampaign(id) {
    const response = await api.post(`/outreach/campaigns/${id}/activate`);
    return response.data;
}

export async function pauseCampaign(id) {
    const response = await api.post(`/outreach/campaigns/${id}/pause`);
    return response.data;
}

export async function deleteCampaign(id) {
    const response = await api.delete(`/outreach/campaigns/${id}`);
    return response.data;
}

export async function generateEmail(data) {
    const response = await api.post('/outreach/generate-email', data);
    return response.data;
}

export async function getCampaignRecipients(campaignId) {
    const response = await api.get(`/outreach/campaigns/${campaignId}/recipients`);
    return response.data;
}

export async function addCampaignRecipients(campaignId, leadIds) {
    const response = await api.post(`/outreach/campaigns/${campaignId}/recipients`, { lead_ids: leadIds });
    return response.data;
}

export async function removeCampaignRecipient(campaignId, recipientId) {
    const response = await api.delete(`/outreach/campaigns/${campaignId}/recipients/${recipientId}`);
    return response.data;
}

export async function getCampaignSequence(campaignId) {
    const response = await api.get(`/outreach/campaigns/${campaignId}/sequence`);
    return response.data;
}

export async function saveCampaignSequence(campaignId, steps) {
    const response = await api.post(`/outreach/campaigns/${campaignId}/sequence`, { steps });
    return response.data;
}
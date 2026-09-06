import api from '../api';

const SUBSCRIPTION_BASE_URL = '/api/v1/subscription';

export const subscriptionAPI = {
  getWorkspace: async () => {
    const response = await api.get(`${SUBSCRIPTION_BASE_URL}`);
    return response.data;
  },

  getPlans: async () => {
    const response = await api.get(`${SUBSCRIPTION_BASE_URL}/plans`);
    return response.data;
  },

  changePlan: async (planId, billingCycle) => {
    const response = await api.post(`${SUBSCRIPTION_BASE_URL}/change`, {
      plan_id: planId,
      billing_cycle: billingCycle
    });
    return response.data;
  },

  checkout: async (planId, billingCycle) => {
    const response = await api.post(`${SUBSCRIPTION_BASE_URL}/checkout`, {
      plan_id: planId,
      billing_cycle: billingCycle
    });
    return response.data;
  },

  cancelSubscription: async () => {
    const response = await api.post(`${SUBSCRIPTION_BASE_URL}/cancel`);
    return response.data;
  },

  reactivateSubscription: async () => {
    const response = await api.post(`${SUBSCRIPTION_BASE_URL}/reactivate`);
    return response.data;
  },

  createCustomerPortal: async () => {
    const response = await api.post(`${SUBSCRIPTION_BASE_URL}/portal`);
    return response.data;
  },

  previewPlanChange: async (planId, billingCycle) => {
    const response = await api.post(`${SUBSCRIPTION_BASE_URL}/plan/preview`, {
      plan_id: planId,
      billing_cycle: billingCycle
    });
    return response.data;
  },

  getAddonConfigs: async () => {
    const response = await api.get(`${SUBSCRIPTION_BASE_URL}/addons/config`);
    return response.data;
  },

  updateAddons: async (addonConfigId, quantity) => {
    const response = await api.post(`${SUBSCRIPTION_BASE_URL}/addons`, {
      addon_config_id: addonConfigId,
      quantity: quantity
    });
    return response.data;
  }
};

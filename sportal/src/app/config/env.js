// SPortal Environment Configuration
export const ENV = {
  API_BASE_URL: import.meta.env.VITE_API_URL ?? '',
  PORTAL_NAME: 'LogisticsHQ SPortal',
  VERSION: '1.0.0-foundation',
  IS_DEV: import.meta.env.DEV,
  // Default internal test token for local development session emulation
  DEFAULT_TOKEN: 'test-token',
};

export const ROUTES = {
  DASHBOARD: '/',
  ORGANIZATIONS: '/organizations',
  ONBOARDING: '/onboarding',
  SUBSCRIPTIONS: '/subscriptions',
  BILLING: '/billing',
  USERS: '/users',
  CUSTOMER_360: '/customer-360',
  USAGE: '/usage',
  CUSTOMER_HEALTH: '/customer-health',
  INTEGRATIONS: '/integrations',
  DOCUMENTS: '/documents',
  SUPPORT: '/support',
  AI: '/ai',
  SETTINGS: '/settings',
  DEMO_REQUESTS: '/demo-requests',
};

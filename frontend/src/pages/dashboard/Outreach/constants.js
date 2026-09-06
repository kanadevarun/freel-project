export const CAMPAIGN_STATUS = {
  DRAFT: 'DRAFT',
  ACTIVE: 'ACTIVE',
  PAUSED: 'PAUSED',
  COMPLETED: 'COMPLETED',
};

export const CHANNEL = {
  EMAIL: 'EMAIL',
};

export const CAMPAIGN_OBJECTIVES = [
  { id: 'INTRO_SERVICES', label: 'Introduce Freight Services' },
  { id: 'GENERATE_RFQS', label: 'Generate RFQs' },
  { id: 'RECONNECT', label: 'Reconnect with Prospect' },
  { id: 'PROMOTE_TRADE_LANE', label: 'Promote Trade Lane' },
  { id: 'GENERAL_BD', label: 'General Business Development' },
];

export const CAMPAIGN_STATUS_CONFIG = {
  [CAMPAIGN_STATUS.DRAFT]: { label: 'Draft', type: 'neutral' },
  [CAMPAIGN_STATUS.ACTIVE]: { label: 'Active', type: 'success' },
  [CAMPAIGN_STATUS.PAUSED]: { label: 'Paused', type: 'warning' },
  [CAMPAIGN_STATUS.COMPLETED]: { label: 'Completed', type: 'info' },
};

export const CAMPAIGN_ACTIONS = {
  ACTIVATE: 'activate',
  PAUSE: 'pause',
  DELETE: 'delete',
};

export const OUTREACH_TABS = {
  DASHBOARD: 'dashboard',
  CAMPAIGNS: 'campaigns',
  COMPOSER: 'composer',
};
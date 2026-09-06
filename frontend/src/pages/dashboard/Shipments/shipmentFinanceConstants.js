/**
 * Constants for Shipment Financial Operations & Costing Intelligence (Task 16.8)
 * Authoritative financial statuses, cost categories, charge types, and UI tokens.
 */

export const FINANCIAL_STATUSES = {
  ESTIMATED: 'ESTIMATED',
  IN_PROGRESS: 'IN_PROGRESS',
  PENDING_REVIEW: 'PENDING_REVIEW',
  PROFITABLE: 'PROFITABLE',
  LOW_MARGIN: 'LOW_MARGIN',
  LOSS: 'LOSS',
  FINANCIALLY_CLOSED: 'FINANCIALLY_CLOSED',
};

export const FINANCIAL_STATUS_CONFIG = {
  [FINANCIAL_STATUSES.PROFITABLE]: {
    label: 'Profitable',
    bg: '#ecfdf5',
    fg: '#059669',
    border: '#a7f3d0',
    description: 'Actual gross margin exceeds baseline profit target.',
  },
  [FINANCIAL_STATUSES.LOW_MARGIN]: {
    label: 'Low Margin Alert',
    bg: '#fffbeb',
    fg: '#d97706',
    border: '#fde68a',
    description: 'Gross margin is below recommended threshold (< 8.0%).',
  },
  [FINANCIAL_STATUSES.LOSS]: {
    label: 'Loss Making',
    bg: '#fef2f2',
    fg: '#dc2626',
    border: '#fecaca',
    description: 'Actual costs exceed revenues resulting in negative margin.',
  },
  [FINANCIAL_STATUSES.PENDING_REVIEW]: {
    label: 'Pending Review',
    bg: '#fef3c7',
    fg: '#b45309',
    border: '#fde68a',
    description: 'Unresolved charge disputes or unverified cost adjustments.',
  },
  [FINANCIAL_STATUSES.IN_PROGRESS]: {
    label: 'In Progress',
    bg: '#eff6ff',
    fg: '#2563eb',
    border: '#bfdbfe',
    description: 'Operational execution underway; awaiting complete invoices.',
  },
  [FINANCIAL_STATUSES.ESTIMATED]: {
    label: 'Estimated',
    bg: '#f1f5f9',
    fg: '#475569',
    border: '#cbd5e1',
    description: 'Quoted targets only; execution costs pending receipt.',
  },
  [FINANCIAL_STATUSES.FINANCIALLY_CLOSED]: {
    label: 'Financially Closed',
    bg: '#f0fdf4',
    fg: '#166534',
    border: '#bbf7d0',
    description: 'All invoices settled and audited. File closed.',
  },
};

export const COST_CATEGORIES = {
  ALL: 'ALL',
  OCEAN_FREIGHT: 'OCEAN_FREIGHT',
  AIR_FREIGHT: 'AIR_FREIGHT',
  ORIGIN_CHARGES: 'ORIGIN_CHARGES',
  DESTINATION_CHARGES: 'DESTINATION_CHARGES',
  CUSTOMS: 'CUSTOMS',
  INSURANCE: 'INSURANCE',
  DETENTION: 'DETENTION',
  DEMURRAGE: 'DEMURRAGE',
  TRUCKING: 'TRUCKING',
  DOCUMENTATION: 'DOCUMENTATION',
  OTHER: 'OTHER',
};

export const COST_CATEGORY_OPTIONS = [
  { value: 'OCEAN_FREIGHT', label: 'Ocean Freight' },
  { value: 'AIR_FREIGHT', label: 'Air Freight' },
  { value: 'ORIGIN_CHARGES', label: 'Origin Port Charges & Drayage' },
  { value: 'DESTINATION_CHARGES', label: 'Destination Charges & THC' },
  { value: 'CUSTOMS', label: 'Customs Clearance & Brokerage' },
  { value: 'INSURANCE', label: 'Cargo Insurance Premium' },
  { value: 'DETENTION', label: 'Equipment Detention' },
  { value: 'DEMURRAGE', label: 'Port / Terminal Demurrage' },
  { value: 'TRUCKING', label: 'Inland Trucking & Haulage' },
  { value: 'DOCUMENTATION', label: 'Documentation & BL Issuance' },
  { value: 'OTHER', label: 'Other Operational Surcharge' },
];

export const CHARGE_TYPES = {
  COST: 'COST',
  REVENUE: 'REVENUE',
};

export const CHARGE_TYPE_OPTIONS = [
  { value: 'COST', label: 'Cost / Operational Expense (Payable)' },
  { value: 'REVENUE', label: 'Revenue / Customer Surcharge (Receivable)' },
];

export const CHARGE_STATUSES = {
  ESTIMATED: 'ESTIMATED',
  INVOICED: 'INVOICED',
  APPROVED: 'APPROVED',
  DISPUTED: 'DISPUTED',
  PAID: 'PAID',
};

export const CHARGE_STATUS_OPTIONS = [
  { value: 'ESTIMATED', label: 'Estimated' },
  { value: 'INVOICED', label: 'Invoiced / Ingested' },
  { value: 'APPROVED', label: 'Approved & Reconciled' },
  { value: 'DISPUTED', label: 'Disputed / Under Review' },
  { value: 'PAID', label: 'Settled / Paid' },
];

export const FINANCIAL_FILTER_TABS = [
  { id: 'ALL', label: 'All Financial Items' },
  { id: 'COST', label: 'Operational Costs' },
  { id: 'REVENUE', label: 'Receivables & Surcharges' },
  { id: 'INVOICES', label: 'Carrier Invoices' },
];

/**
 * Authoritative constants for Shipment Documents, Categories, Statuses, and Compliance States (Task 16.9).
 * Synchronized with backend/internal/shipments/spec/const.go
 */

export const DOC_CATEGORIES = {
  ALL: 'ALL',
  TRANSPORT: 'TRANSPORT',
  COMMERCIAL: 'COMMERCIAL',
  CUSTOMS: 'CUSTOMS',
  INSURANCE: 'INSURANCE',
  OPERATIONAL: 'OPERATIONAL',
  CARGO: 'CARGO',
  OTHER: 'OTHER',
};

export const DOC_TYPES = {
  // Transport
  MBL: 'MBL',
  HBL: 'HBL',
  SEA_WAYBILL: 'SEA_WAYBILL',
  AIR_WAYBILL: 'AIR_WAYBILL',
  ARRIVAL_NOTICE: 'ARRIVAL_NOTICE',
  DELIVERY_ORDER: 'DELIVERY_ORDER',
  BOOKING_CONFIRMATION: 'BOOKING_CONFIRMATION',
  CARRIER_RELEASE: 'CARRIER_RELEASE',
  BILL_OF_LADING: 'BILL_OF_LADING',

  // Commercial
  COMMERCIAL_INVOICE: 'COMMERCIAL_INVOICE',
  PROFORMA_INVOICE: 'PROFORMA_INVOICE',
  PACKING_LIST: 'PACKING_LIST',
  PURCHASE_ORDER: 'PURCHASE_ORDER',
  CERTIFICATE_OF_ORIGIN: 'CERTIFICATE_OF_ORIGIN',

  // Customs & Regulatory
  CUSTOMS_DECLARATION: 'CUSTOMS_DECLARATION',
  CUSTOMS_CLEARANCE: 'CUSTOMS_CLEARANCE',
  IMPORT_DOCUMENTATION: 'IMPORT_DOCUMENTATION',
  EXPORT_DOCUMENTATION: 'EXPORT_DOCUMENTATION',
  IMPORT_PERMIT: 'IMPORT_PERMIT',
  EXPORT_PERMIT: 'EXPORT_PERMIT',
  DUTY_DOCUMENTATION: 'DUTY_DOCUMENTATION',
  REGULATORY_CERTIFICATE: 'REGULATORY_CERTIFICATE',
  INSPECTION_CERTIFICATE: 'INSPECTION_CERTIFICATE',

  // Insurance
  INSURANCE_CERTIFICATE: 'INSURANCE_CERTIFICATE',
  INSURANCE_POLICY: 'INSURANCE_POLICY',

  // Operational & Cargo
  SHIPPING_INSTRUCTIONS: 'SHIPPING_INSTRUCTIONS',
  CARGO_MANIFEST: 'CARGO_MANIFEST',
  WEIGHT_CERTIFICATE: 'WEIGHT_CERTIFICATE',
  DELIVERY_RECEIPT: 'DELIVERY_RECEIPT',
  DANGEROUS_GOODS_DECLARATION: 'DANGEROUS_GOODS_DECLARATION',
  CARGO_CERTIFICATE: 'CARGO_CERTIFICATE',

  // Other
  CARRIER_DOCUMENT: 'CARRIER_DOCUMENT',
  CUSTOMER_DOCUMENT: 'CUSTOMER_DOCUMENT',
  OTHER: 'OTHER',
};

export const DOC_STATUSES = {
  MISSING: 'MISSING',
  REQUESTED: 'REQUESTED',
  UPLOADED: 'UPLOADED',
  UNDER_REVIEW: 'UNDER_REVIEW',
  APPROVED: 'APPROVED',
  REJECTED: 'REJECTED',
  EXPIRED: 'EXPIRED',
  SUPERSEDED: 'SUPERSEDED',
};

export const DOC_STATUS_CONFIG = {
  [DOC_STATUSES.APPROVED]: {
    label: 'Approved',
    bg: '#ecfdf5',
    fg: '#059669',
    border: '#a7f3d0',
  },
  [DOC_STATUSES.UNDER_REVIEW]: {
    label: 'Under Review',
    bg: '#eff6ff',
    fg: '#2563eb',
    border: '#bfdbfe',
  },
  [DOC_STATUSES.UPLOADED]: {
    label: 'Uploaded',
    bg: '#f8fafc',
    fg: '#475569',
    border: '#cbd5e1',
  },
  [DOC_STATUSES.REJECTED]: {
    label: 'Rejected',
    bg: '#fef2f2',
    fg: '#dc2626',
    border: '#fecaca',
  },
  [DOC_STATUSES.EXPIRED]: {
    label: 'Expired',
    bg: '#fff1f2',
    fg: '#be123c',
    border: '#fecdd3',
  },
  [DOC_STATUSES.REQUESTED]: {
    label: 'Requested',
    bg: '#fef3c7',
    fg: '#b45309',
    border: '#fde68a',
  },
  [DOC_STATUSES.MISSING]: {
    label: 'Missing',
    bg: '#fef2f2',
    fg: '#b91c1c',
    border: '#fca5a5',
  },
  [DOC_STATUSES.SUPERSEDED]: {
    label: 'Superseded',
    bg: '#f1f5f9',
    fg: '#64748b',
    border: '#e2e8f0',
  },
};

export const REQUIREMENT_LEVELS = {
  CRITICAL: 'CRITICAL',
  REQUIRED: 'REQUIRED',
  OPTIONAL: 'OPTIONAL',
};

export const REQUIREMENT_LEVEL_CONFIG = {
  [REQUIREMENT_LEVELS.CRITICAL]: {
    label: 'Critical',
    bg: '#fef2f2',
    fg: '#dc2626',
    border: '#fecaca',
    description: 'Mandatory title or release document. Required before transit / closure.',
  },
  [REQUIREMENT_LEVELS.REQUIRED]: {
    label: 'Required',
    bg: '#fffbeb',
    fg: '#d97706',
    border: '#fde68a',
    description: 'Required for customs clearance, port terminal gate-in, or delivery.',
  },
  [REQUIREMENT_LEVELS.OPTIONAL]: {
    label: 'Optional',
    bg: '#f8fafc',
    fg: '#64748b',
    border: '#e2e8f0',
    description: 'Conditional on FTA tariff concessions, special cargo, or insurance.',
  },
};

export const VALIDITY_STATES = {
  VALID: 'VALID',
  EXPIRING_SOON: 'EXPIRING_SOON',
  EXPIRED: 'EXPIRED',
};

export const VALIDITY_STATE_CONFIG = {
  [VALIDITY_STATES.VALID]: {
    label: 'Valid',
    bg: '#ecfdf5',
    fg: '#059669',
    border: '#a7f3d0',
  },
  [VALIDITY_STATES.EXPIRING_SOON]: {
    label: 'Expiring Soon',
    bg: '#fffbeb',
    fg: '#d97706',
    border: '#fde68a',
  },
  [VALIDITY_STATES.EXPIRED]: {
    label: 'Expired',
    bg: '#fef2f2',
    fg: '#dc2626',
    border: '#fecaca',
  },
};

export const COMPLIANCE_STATES = {
  COMPLIANT: 'COMPLIANT',
  READY: 'READY',
  ATTENTION_REQUIRED: 'ATTENTION_REQUIRED',
  ACTION_REQUIRED: 'ACTION_REQUIRED',
  AT_RISK: 'AT_RISK',
  NON_COMPLIANT: 'NON_COMPLIANT',
  BLOCKED: 'BLOCKED',
};

export const COMPLIANCE_STATE_CONFIG = {
  [COMPLIANCE_STATES.COMPLIANT]: {
    label: 'Compliant',
    bg: '#ecfdf5',
    fg: '#059669',
    border: '#a7f3d0',
    description: 'All mandatory operational and commercial documents verified & valid.',
  },
  [COMPLIANCE_STATES.READY]: {
    label: 'Compliant',
    bg: '#ecfdf5',
    fg: '#059669',
    border: '#a7f3d0',
    description: 'All mandatory operational and commercial documents verified & valid.',
  },
  [COMPLIANCE_STATES.ATTENTION_REQUIRED]: {
    label: 'Attention Required',
    bg: '#fffbeb',
    fg: '#d97706',
    border: '#fde68a',
    description: 'One or more required documents are missing or pending upload.',
  },
  [COMPLIANCE_STATES.ACTION_REQUIRED]: {
    label: 'Attention Required',
    bg: '#fffbeb',
    fg: '#d97706',
    border: '#fde68a',
    description: 'One or more required documents are missing or pending upload.',
  },
  [COMPLIANCE_STATES.AT_RISK]: {
    label: 'At Risk',
    bg: '#fff7ed',
    fg: '#c2410c',
    border: '#ffedd5',
    description: 'Documents approaching expiry (< 14 days) or under urgent review.',
  },
  [COMPLIANCE_STATES.BLOCKED]: {
    label: 'Blocked',
    bg: '#fef2f2',
    fg: '#dc2626',
    border: '#fecaca',
    description: 'Critical document is rejected or expired. Operational hold active.',
  },
  [COMPLIANCE_STATES.NON_COMPLIANT]: {
    label: 'Non-Compliant',
    bg: '#fef2f2',
    fg: '#dc2626',
    border: '#fecaca',
    description: 'Critical document rejected or expired. Immediate review required.',
  },
};

export const DOC_SOURCES = {
  SHIPMENT: 'SHIPMENT',
  RFQ: 'RFQ',
  BOOKING: 'BOOKING',
};

export const CATEGORY_TABS = [
  { id: DOC_CATEGORIES.ALL, label: 'All Documents' },
  { id: DOC_CATEGORIES.TRANSPORT, label: 'Transport (B/L)' },
  { id: DOC_CATEGORIES.COMMERCIAL, label: 'Commercial' },
  { id: DOC_CATEGORIES.CUSTOMS, label: 'Customs & Regulatory' },
  { id: DOC_CATEGORIES.INSURANCE, label: 'Insurance' },
  { id: DOC_CATEGORIES.OPERATIONAL, label: 'Operational' },
  { id: DOC_CATEGORIES.OTHER, label: 'Other' },
];

export const CATEGORY_DROPDOWN_OPTIONS = [
  { value: DOC_CATEGORIES.TRANSPORT, label: 'Transport Documents' },
  { value: DOC_CATEGORIES.COMMERCIAL, label: 'Commercial Documents' },
  { value: DOC_CATEGORIES.CUSTOMS, label: 'Customs & Regulatory' },
  { value: DOC_CATEGORIES.INSURANCE, label: 'Cargo Insurance' },
  { value: DOC_CATEGORIES.OPERATIONAL, label: 'Operational & Handling' },
  { value: DOC_CATEGORIES.OTHER, label: 'Other Documents' },
];

export const DOC_TYPE_OPTIONS = [
  // Transport
  { value: DOC_TYPES.MBL, label: 'Master Bill of Lading (MBL)', category: DOC_CATEGORIES.TRANSPORT },
  { value: DOC_TYPES.HBL, label: 'House Bill of Lading (HBL)', category: DOC_CATEGORIES.TRANSPORT },
  { value: DOC_TYPES.SEA_WAYBILL, label: 'Sea Waybill', category: DOC_CATEGORIES.TRANSPORT },
  { value: DOC_TYPES.AIR_WAYBILL, label: 'Air Waybill (AWB)', category: DOC_CATEGORIES.TRANSPORT },
  { value: DOC_TYPES.ARRIVAL_NOTICE, label: 'Arrival Notice', category: DOC_CATEGORIES.TRANSPORT },
  { value: DOC_TYPES.DELIVERY_ORDER, label: 'Delivery Order', category: DOC_CATEGORIES.TRANSPORT },
  { value: DOC_TYPES.BOOKING_CONFIRMATION, label: 'Booking Confirmation', category: DOC_CATEGORIES.TRANSPORT },
  { value: DOC_TYPES.CARRIER_RELEASE, label: 'Carrier Release', category: DOC_CATEGORIES.TRANSPORT },

  // Commercial
  { value: DOC_TYPES.COMMERCIAL_INVOICE, label: 'Commercial Invoice (CI)', category: DOC_CATEGORIES.COMMERCIAL },
  { value: DOC_TYPES.PROFORMA_INVOICE, label: 'Proforma Invoice (PI)', category: DOC_CATEGORIES.COMMERCIAL },
  { value: DOC_TYPES.PACKING_LIST, label: 'Packing List (PL)', category: DOC_CATEGORIES.COMMERCIAL },
  { value: DOC_TYPES.PURCHASE_ORDER, label: 'Purchase Order (PO)', category: DOC_CATEGORIES.COMMERCIAL },
  { value: DOC_TYPES.CERTIFICATE_OF_ORIGIN, label: 'Certificate of Origin (COO)', category: DOC_CATEGORIES.COMMERCIAL },

  // Customs & Regulatory
  { value: DOC_TYPES.CUSTOMS_DECLARATION, label: 'Customs Declaration / Entry', category: DOC_CATEGORIES.CUSTOMS },
  { value: DOC_TYPES.CUSTOMS_CLEARANCE, label: 'Customs Clearance Release', category: DOC_CATEGORIES.CUSTOMS },
  { value: DOC_TYPES.IMPORT_PERMIT, label: 'Import Permit', category: DOC_CATEGORIES.CUSTOMS },
  { value: DOC_TYPES.EXPORT_PERMIT, label: 'Export Permit', category: DOC_CATEGORIES.CUSTOMS },
  { value: DOC_TYPES.DUTY_DOCUMENTATION, label: 'Duty & Tax Receipt', category: DOC_CATEGORIES.CUSTOMS },
  { value: DOC_TYPES.REGULATORY_CERTIFICATE, label: 'Regulatory Certificate', category: DOC_CATEGORIES.CUSTOMS },
  { value: DOC_TYPES.INSPECTION_CERTIFICATE, label: 'Inspection Certificate', category: DOC_CATEGORIES.CUSTOMS },

  // Insurance
  { value: DOC_TYPES.INSURANCE_CERTIFICATE, label: 'Cargo Insurance Certificate', category: DOC_CATEGORIES.INSURANCE },
  { value: DOC_TYPES.INSURANCE_POLICY, label: 'Insurance Policy Terms', category: DOC_CATEGORIES.INSURANCE },

  // Operational
  { value: DOC_TYPES.SHIPPING_INSTRUCTIONS, label: 'Shipping Instructions (SI)', category: DOC_CATEGORIES.OPERATIONAL },
  { value: DOC_TYPES.CARGO_MANIFEST, label: 'Cargo Manifest', category: DOC_CATEGORIES.OPERATIONAL },
  { value: DOC_TYPES.WEIGHT_CERTIFICATE, label: 'Weight Certificate / VGM', category: DOC_CATEGORIES.OPERATIONAL },
  { value: DOC_TYPES.DELIVERY_RECEIPT, label: 'Delivery Receipt / Proof of Delivery', category: DOC_CATEGORIES.OPERATIONAL },
  { value: DOC_TYPES.DANGEROUS_GOODS_DECLARATION, label: 'Dangerous Goods (MSDS)', category: DOC_CATEGORIES.OPERATIONAL },
  { value: DOC_TYPES.CARGO_CERTIFICATE, label: 'Cargo Certificate', category: DOC_CATEGORIES.OPERATIONAL },

  // Other
  { value: DOC_TYPES.CARRIER_DOCUMENT, label: 'Carrier Communication / Notice', category: DOC_CATEGORIES.OTHER },
  { value: DOC_TYPES.CUSTOMER_DOCUMENT, label: 'Customer Instruction', category: DOC_CATEGORIES.OTHER },
  { value: DOC_TYPES.OTHER, label: 'Other Attachment', category: DOC_CATEGORIES.OTHER },
];

/**
 * Centralized Mock Invoice Data & KPI Metrics for LogisticsHQ Finance Module
 * Designed to mirror dashboard/invoices/invoices.png
 * Easily swappable with backend API in Task 2.
 */

export const INITIAL_KPI_STATS = {
  totalInvoices: {
    amount: '$2,480,000.00',
    displayAmount: '$2.48M',
    count: 128,
    label: '128 Invoices',
    trend: '18.6%',
    trendDirection: 'up',
    trendPeriod: 'vs last 7 days'
  },
  outstanding: {
    amount: '$1,420,000.00',
    displayAmount: '$1.42M',
    count: 86,
    label: '86 Invoices',
    trend: '12.4%',
    trendDirection: 'up',
    trendPeriod: 'vs last 7 days'
  },
  paidThisMonth: {
    amount: '$96,420.00',
    displayAmount: '$96,420',
    count: 32,
    label: '32 Invoices',
    trend: '24.8%',
    trendDirection: 'up',
    trendPeriod: 'vs last 7 days'
  },
  overdue: {
    amount: '$38,750.00',
    displayAmount: '$38,750',
    count: 14,
    label: '14 Invoices',
    trend: '8.2%',
    trendDirection: 'up',
    trendPeriod: 'vs last 7 days'
  }
};

export const INVOICE_STATUSES = {
  ALL: 'All',
  DRAFT: 'Draft',
  PENDING_APPROVAL: 'Pending Approval',
  ISSUED: 'Issued',
  PARTIALLY_PAID: 'Partially Paid',
  PAID: 'Paid',
  OVERDUE: 'Overdue',
  CANCELLED: 'Cancelled'
};

export const MOCK_INVOICES = [
  {
    id: 'inv-001',
    invoiceNumber: 'INV-2026-0456',
    creator: 'By Varun Sharma',
    customer: 'Global Traders Inc.',
    customerCountry: 'USA',
    shipmentId: 'SH-2026-00124',
    route: 'Shanghai ➔ Los Angeles',
    origin: 'Shanghai',
    destination: 'Los Angeles',
    invoiceDate: 'Aug 15, 2026',
    dueDate: 'Aug 30, 2026',
    daysLeft: '15 days left',
    amount: 24650.00,
    currency: 'USD',
    status: 'Issued',
    balance: 24650.00,
    type: 'CUSTOMER_AR',
    bookmarked: true,
    isMyInvoice: true,
    subtotal: 22410.00,
    tax: 0.00,
    discount: 0.00,
    amountPaid: 0.00,
    lineItems: [
      { id: 1, description: 'Ocean Freight (40ft FCL High Cube)', quantity: 1, rate: 18500.00, amount: 18500.00 },
      { id: 2, description: 'Documentation Fee & B/L Issuance', quantity: 1, rate: 350.00, amount: 350.00 },
      { id: 3, description: 'Customs Handling & Import Brokerage', quantity: 1, rate: 1200.00, amount: 1200.00 },
      { id: 4, description: 'Local Drayage & Delivery Surcharge', quantity: 1, rate: 2360.00, amount: 2360.00 }
    ],
    payments: [],
    documents: [
      { id: 'doc-1', name: 'Commercial Invoice.pdf', size: '245 KB', type: 'application/pdf', uploadedAt: 'Aug 15, 2026' },
      { id: 'doc-2', name: 'Bill of Lading - SH-2026-00124.pdf', size: '320 KB', type: 'application/pdf', uploadedAt: 'Aug 15, 2026' },
      { id: 'doc-3', name: 'Packing List.pdf', size: '180 KB', type: 'application/pdf', uploadedAt: 'Aug 15, 2026' }
    ],
    history: [
      { id: 'h-1', title: 'Invoice Issued', description: 'Issued to customer Global Traders Inc.', timestamp: 'Aug 15, 2026 10:30 AM', user: 'Varun Sharma' },
      { id: 'h-2', title: 'Approval Passed', description: 'Internal billing review approved by Finance Team', timestamp: 'Aug 14, 2026 04:15 PM', user: 'Finance Lead' },
      { id: 'h-3', title: 'Draft Created', description: 'Generated from Shipment SH-2026-00124', timestamp: 'Aug 14, 2026 02:00 PM', user: 'Varun Sharma' }
    ]
  },
  {
    id: 'inv-002',
    invoiceNumber: 'INV-2026-0455',
    creator: 'By Priya Nair',
    customer: 'Oceanic Imports Pvt. Ltd.',
    customerCountry: 'India',
    shipmentId: 'SH-2026-00123',
    route: 'Nhava Sheva ➔ New York',
    origin: 'Nhava Sheva',
    destination: 'New York',
    invoiceDate: 'Aug 14, 2026',
    dueDate: 'Aug 29, 2026',
    daysLeft: '14 days left',
    amount: 18940.00,
    currency: 'USD',
    status: 'Partially Paid',
    balance: 8940.00,
    type: 'CUSTOMER_AR',
    bookmarked: false,
    isMyInvoice: false,
    subtotal: 17200.00,
    tax: 1740.00,
    discount: 0.00,
    amountPaid: 10000.00,
    lineItems: [
      { id: 1, description: 'Ocean Freight (20ft FCL)', quantity: 2, rate: 7000.00, amount: 14000.00 },
      { id: 2, description: 'Terminal Handling Charge (THC)', quantity: 2, rate: 850.00, amount: 1700.00 },
      { id: 3, description: 'Customs Duty Processing', quantity: 1, rate: 1500.00, amount: 1500.00 },
      { id: 4, description: 'Insurance Coverage Premium', quantity: 1, rate: 1740.00, amount: 1740.00 }
    ],
    payments: [
      { id: 'p-1', reference: 'PAY-2026-8801', amount: 10000.00, method: 'Wire Transfer', status: 'Completed', date: 'Aug 18, 2026' }
    ],
    documents: [
      { id: 'doc-1', name: 'Commercial Invoice.pdf', size: '210 KB', type: 'application/pdf', uploadedAt: 'Aug 14, 2026' },
      { id: 'doc-2', name: 'Wire Receipt - PAY-8801.pdf', size: '115 KB', type: 'application/pdf', uploadedAt: 'Aug 18, 2026' }
    ],
    history: [
      { id: 'h-1', title: 'Partial Payment Received', description: '$10,000.00 received via Wire Transfer', timestamp: 'Aug 18, 2026 11:20 AM', user: 'Priya Nair' },
      { id: 'h-2', title: 'Invoice Issued', description: 'Issued to Oceanic Imports Pvt. Ltd.', timestamp: 'Aug 14, 2026 09:45 AM', user: 'Priya Nair' }
    ]
  },
  {
    id: 'inv-003',
    invoiceNumber: 'INV-2026-0454',
    creator: 'By Varun Sharma',
    customer: 'Bright Star Ltd.',
    customerCountry: 'UK',
    shipmentId: 'SH-2026-00122',
    route: 'Dubai ➔ Felixstowe',
    origin: 'Dubai',
    destination: 'Felixstowe',
    invoiceDate: 'Aug 12, 2026',
    dueDate: 'Aug 27, 2026',
    daysLeft: '12 days left',
    amount: 32120.00,
    currency: 'USD',
    status: 'Overdue',
    balance: 32120.00,
    type: 'CUSTOMER_AR',
    bookmarked: true,
    isMyInvoice: true,
    subtotal: 30000.00,
    tax: 2120.00,
    discount: 0.00,
    amountPaid: 0.00,
    lineItems: [
      { id: 1, description: 'Air Freight Charter (LHR Freight)', quantity: 1, rate: 26000.00, amount: 26000.00 },
      { id: 2, description: 'Priority Express Clearance', quantity: 1, rate: 4000.00, amount: 4000.00 },
      { id: 3, description: 'VAT / Import Duty Advance', quantity: 1, rate: 2120.00, amount: 2120.00 }
    ],
    payments: [],
    documents: [
      { id: 'doc-1', name: 'Invoice_BrightStar_0454.pdf', size: '305 KB', type: 'application/pdf', uploadedAt: 'Aug 12, 2026' }
    ],
    history: [
      { id: 'h-1', title: 'Payment Overdue Notice', description: 'Automated reminder sent to billing contact', timestamp: 'Aug 28, 2026 09:00 AM', user: 'System' },
      { id: 'h-2', title: 'Invoice Issued', description: 'Issued to Bright Star Ltd.', timestamp: 'Aug 12, 2026 03:30 PM', user: 'Varun Sharma' }
    ]
  },
  {
    id: 'inv-004',
    invoiceNumber: 'INV-2026-0453',
    creator: 'By Priya Nair',
    customer: 'Techtronics GmbH',
    customerCountry: 'Germany',
    shipmentId: 'SH-2026-00121',
    route: 'Hamburg ➔ Mumbai',
    origin: 'Hamburg',
    destination: 'Mumbai',
    invoiceDate: 'Aug 11, 2026',
    dueDate: 'Aug 26, 2026',
    daysLeft: '11 days left',
    amount: 15780.00,
    currency: 'USD',
    status: 'Paid',
    balance: 0.00,
    type: 'CUSTOMER_AR',
    bookmarked: false,
    isMyInvoice: false,
    subtotal: 15780.00,
    tax: 0.00,
    discount: 0.00,
    amountPaid: 15780.00,
    lineItems: [
      { id: 1, description: 'Sea Freight LCL Shipment', quantity: 1, rate: 12500.00, amount: 12500.00 },
      { id: 2, description: 'Destination Port Delivery Charges', quantity: 1, rate: 3280.00, amount: 3280.00 }
    ],
    payments: [
      { id: 'p-1', reference: 'PAY-2026-7910', amount: 15780.00, method: 'ACH Transfer', status: 'Completed', date: 'Aug 20, 2026' }
    ],
    documents: [
      { id: 'doc-1', name: 'Invoice_Techtronics.pdf', size: '190 KB', type: 'application/pdf', uploadedAt: 'Aug 11, 2026' },
      { id: 'doc-2', name: 'Payment_Confirmation.pdf', size: '95 KB', type: 'application/pdf', uploadedAt: 'Aug 20, 2026' }
    ],
    history: [
      { id: 'h-1', title: 'Payment Settled', description: 'Full payment received. Balance is $0.00', timestamp: 'Aug 20, 2026 02:15 PM', user: 'Priya Nair' }
    ]
  },
  {
    id: 'inv-005',
    invoiceNumber: 'INV-2026-0452',
    creator: 'By Rohan Mehta',
    customer: 'Southern Retail LLC',
    customerCountry: 'USA',
    shipmentId: 'SH-2026-00120',
    route: 'Los Angeles ➔ Chicago',
    origin: 'Los Angeles',
    destination: 'Chicago',
    invoiceDate: 'Aug 10, 2026',
    dueDate: 'Aug 25, 2026',
    daysLeft: '10 days left',
    amount: 9650.00,
    currency: 'USD',
    status: 'Paid',
    balance: 0.00,
    type: 'CUSTOMER_AR',
    bookmarked: false,
    isMyInvoice: true,
    subtotal: 9650.00,
    tax: 0.00,
    discount: 0.00,
    amountPaid: 9650.00,
    lineItems: [
      { id: 1, description: 'Cross-Country Truckload Haulage', quantity: 1, rate: 8500.00, amount: 8500.00 },
      { id: 2, description: 'Fuel Surcharge & Tolls', quantity: 1, rate: 1150.00, amount: 1150.00 }
    ],
    payments: [
      { id: 'p-1', reference: 'PAY-2026-6420', amount: 9650.00, method: 'Credit Card', status: 'Completed', date: 'Aug 19, 2026' }
    ],
    documents: [
      { id: 'doc-1', name: 'Invoice_SouthernRetail.pdf', size: '210 KB', type: 'application/pdf', uploadedAt: 'Aug 10, 2026' }
    ],
    history: [
      { id: 'h-1', title: 'Payment Completed', description: 'Credit card payment settled automatically', timestamp: 'Aug 19, 2026 04:00 PM', user: 'System' }
    ]
  },
  {
    id: 'inv-006',
    invoiceNumber: 'INV-2026-0451',
    creator: 'By Varun Sharma',
    customer: 'Alpha Logistics',
    customerCountry: 'Singapore',
    shipmentId: 'SH-2026-00119',
    route: 'Singapore ➔ Sydney',
    origin: 'Singapore',
    destination: 'Sydney',
    invoiceDate: 'Aug 09, 2026',
    dueDate: 'Aug 24, 2026',
    daysLeft: '9 days left',
    amount: 7850.00,
    currency: 'USD',
    status: 'Pending Approval',
    balance: 7850.00,
    type: 'CUSTOMER_AR',
    bookmarked: false,
    isMyInvoice: true,
    subtotal: 7850.00,
    tax: 0.00,
    discount: 0.00,
    amountPaid: 0.00,
    lineItems: [
      { id: 1, description: 'Ocean Freight LCL', quantity: 1, rate: 6500.00, amount: 6500.00 },
      { id: 2, description: 'Origin Wharfage Surcharge', quantity: 1, rate: 1350.00, amount: 1350.00 }
    ],
    payments: [],
    documents: [
      { id: 'doc-1', name: 'Draft_Invoice_AlphaLogistics.pdf', size: '160 KB', type: 'application/pdf', uploadedAt: 'Aug 09, 2026' }
    ],
    history: [
      { id: 'h-1', title: 'Submitted for Approval', description: 'Pending approval by Finance Manager', timestamp: 'Aug 09, 2026 11:15 AM', user: 'Varun Sharma' }
    ]
  },
  {
    id: 'inv-007',
    invoiceNumber: 'INV-2026-0450',
    creator: 'By Priya Nair',
    customer: 'East Coast Traders',
    customerCountry: 'Canada',
    shipmentId: 'SH-2026-00118',
    route: 'Vancouver ➔ Seattle',
    origin: 'Vancouver',
    destination: 'Seattle',
    invoiceDate: 'Aug 08, 2026',
    dueDate: 'Aug 23, 2026',
    daysLeft: '8 days left',
    amount: 11320.00,
    currency: 'USD',
    status: 'Draft',
    balance: 11320.00,
    type: 'CUSTOMER_AR',
    bookmarked: false,
    isMyInvoice: false,
    subtotal: 11320.00,
    tax: 0.00,
    discount: 0.00,
    amountPaid: 0.00,
    lineItems: [
      { id: 1, description: 'Intermodal Container Transit', quantity: 1, rate: 10000.00, amount: 10000.00 },
      { id: 2, description: 'Cross-Border Entry Filing', quantity: 1, rate: 1320.00, amount: 1320.00 }
    ],
    payments: [],
    documents: [
      { id: 'doc-1', name: 'Worksheet_Draft_0450.pdf', size: '140 KB', type: 'application/pdf', uploadedAt: 'Aug 08, 2026' }
    ],
    history: [
      { id: 'h-1', title: 'Draft Invoice Created', description: 'Created by Priya Nair', timestamp: 'Aug 08, 2026 03:00 PM', user: 'Priya Nair' }
    ]
  },
  {
    id: 'inv-008',
    invoiceNumber: 'INV-2026-0449',
    creator: 'By Rohan Mehta',
    customer: 'Sunrise Exports',
    customerCountry: 'India',
    shipmentId: 'SH-2026-00117',
    route: 'Chennai ➔ Los Angeles',
    origin: 'Chennai',
    destination: 'Los Angeles',
    invoiceDate: 'Aug 06, 2026',
    dueDate: 'Aug 21, 2026',
    daysLeft: '6 days left',
    amount: 26450.00,
    currency: 'USD',
    status: 'Issued',
    balance: 26450.00,
    type: 'CUSTOMER_AR',
    bookmarked: false,
    isMyInvoice: false,
    subtotal: 26450.00,
    tax: 0.00,
    discount: 0.00,
    amountPaid: 0.00,
    lineItems: [
      { id: 1, description: '40ft High Cube Container Freight', quantity: 1, rate: 21500.00, amount: 21500.00 },
      { id: 2, description: 'Export Customs Duty & Documentation', quantity: 1, rate: 2950.00, amount: 2950.00 },
      { id: 3, description: 'AMS Security Filing Fee', quantity: 1, rate: 2000.00, amount: 2000.00 }
    ],
    payments: [],
    documents: [
      { id: 'doc-1', name: 'Commercial_Invoice_Sunrise.pdf', size: '280 KB', type: 'application/pdf', uploadedAt: 'Aug 06, 2026' }
    ],
    history: [
      { id: 'h-1', title: 'Invoice Issued', description: 'Issued to Sunrise Exports', timestamp: 'Aug 06, 2026 01:45 PM', user: 'Rohan Mehta' }
    ]
  }
];

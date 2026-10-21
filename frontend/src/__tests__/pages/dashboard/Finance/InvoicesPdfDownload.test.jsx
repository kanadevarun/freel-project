import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import InvoiceDetailsPanel from '../../../../pages/dashboard/Finance/components/InvoiceDetailsPanel';
import { generateInvoicePDF } from '../../../../pages/dashboard/Finance/utils/invoicePdfGenerator';

vi.mock('../../../../pages/dashboard/Finance/utils/invoicePdfGenerator', () => ({
  generateInvoicePDF: vi.fn(),
}));

vi.mock('react-router-dom', () => ({
  useNavigate: () => vi.fn(),
}));

vi.mock('react-hot-toast', () => ({
  default: {
    success: vi.fn(),
    error: vi.fn(),
  },
}));

describe('Invoice Download PDF functionality', () => {
  const mockInvoice = {
    id: 101,
    invoiceNumber: 'INV-2026-0042',
    creator: 'By Sarah Jenkins',
    customer: 'Apex Global Logistics Corp',
    customerCountry: 'US',
    shipmentId: 'SH-2026-0042',
    route: 'Shanghai (CNSHA) ➔ Rotterdam (NLRTM)',
    invoiceDate: 'Aug 15, 2026',
    dueDate: 'Sep 15, 2026',
    daysLeft: '26 days left',
    amount: 14250.00,
    currency: 'USD',
    status: 'Issued',
    balance: 14250.00,
    type: 'CUSTOMER_AR',
    bookmarked: false,
    subtotal: 13000.00,
    tax: 1250.00,
    discount: 0,
    amountPaid: 0,
    lineItems: [
      { id: 1, description: 'Ocean Freight 40ft HC', quantity: 2, rate: 5500, amount: 11000 },
      { id: 2, description: 'Bunker Adjustment Factor (BAF)', quantity: 2, rate: 1000, amount: 2000 },
    ],
    payments: [],
    documents: [],
    history: [],
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders header Download PDF button and triggers generateInvoicePDF on click', () => {
    const onActionClick = vi.fn();
    render(
      <InvoiceDetailsPanel
        invoice={mockInvoice}
        onClose={vi.fn()}
        onActionClick={onActionClick}
      />
    );

    const headerPdfBtn = screen.getByTitle('Download PDF');
    expect(headerPdfBtn).toBeInTheDocument();

    fireEvent.click(headerPdfBtn);
    expect(generateInvoicePDF).toHaveBeenCalledWith(mockInvoice);
    expect(onActionClick).toHaveBeenCalledWith('downloadPdf', mockInvoice);
  });

  it('renders footer Download PDF button and triggers generateInvoicePDF on click', () => {
    const onActionClick = vi.fn();
    const { container } = render(
      <InvoiceDetailsPanel
        invoice={mockInvoice}
        onClose={vi.fn()}
        onActionClick={onActionClick}
      />
    );

    const footerPdfBtn = container.querySelector('#inv-footer-download-pdf-btn');
    expect(footerPdfBtn).toBeInTheDocument();

    fireEvent.click(footerPdfBtn);
    expect(generateInvoicePDF).toHaveBeenCalledWith(mockInvoice);
  });

  it('works when invoice status is Paid', () => {
    const paidInvoice = { ...mockInvoice, status: 'Paid', balance: 0, amountPaid: 14250 };
    const { container } = render(
      <InvoiceDetailsPanel
        invoice={paidInvoice}
        onClose={vi.fn()}
      />
    );

    const footerPdfBtn = container.querySelector('#inv-footer-download-pdf-btn');
    expect(footerPdfBtn).toBeInTheDocument();
    fireEvent.click(footerPdfBtn);
    expect(generateInvoicePDF).toHaveBeenCalledWith(paidInvoice);
  });
});

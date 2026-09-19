import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import DebitNotesPage from '../../../../pages/dashboard/Finance/DebitNotesPage';
import api from '../../../../services/api';

vi.mock('../../../../services/api', () => ({
  default: {
    get: vi.fn(),
    post: vi.fn(),
    put: vi.fn(),
  },
}));

describe('DebitNotesPage Component', () => {
  const mockDebitNotes = [
    {
      id: 1,
      debit_note_number: 'DN-20260907-0001',
      customer_name: 'Apex Global Logistics Corp',
      reason: 'Demurrage at destination port',
      currency: 'USD',
      subtotal: 870,
      tax_amount: 75,
      total_amount: 945,
      status: 'ISSUED',
      issue_date: '2026-09-07T00:00:00Z',
      due_date: '2026-09-21T00:00:00Z',
      created_at: '2026-09-07T18:30:32Z',
      line_items: [
        {
          id: 1,
          description: 'Port Demurrage - 3 Days',
          service_category: 'Demurrage',
          quantity: 3,
          unit_price: 250,
          total_amount: 750,
        },
      ],
    },
    {
      id: 2,
      debit_note_number: 'DN-20260907-0002',
      customer_name: 'Euro-Asia Retailers Ltd',
      reason: 'Emergency warehousing storage',
      currency: 'EUR',
      subtotal: 360,
      tax_amount: 40,
      total_amount: 400,
      status: 'VOID',
      issue_date: '2026-09-07T00:00:00Z',
      due_date: '2026-09-28T00:00:00Z',
      created_at: '2026-09-07T18:30:52Z',
      line_items: [],
    },
  ];

  const mockStats = {
    total_issued: { count: 1, display_amount: '$945.00' },
    outstanding: { count: 1, display_amount: '$945.00' },
    draft: { count: 0, display_amount: '$0.00' },
    void: { count: 1, display_amount: '$400.00' },
  };

  beforeEach(() => {
    vi.clearAllMocks();
    api.get.mockImplementation((url) => {
      if (url.includes('/kpi-stats')) {
        return Promise.resolve(mockStats);
      }
      if (url.includes('/debit-notes/1')) {
        return Promise.resolve(mockDebitNotes[0]);
      }
      if (url.includes('/companies')) {
        return Promise.resolve([]);
      }
      if (url.includes('/debit-notes')) {
        return Promise.resolve({
          debit_notes: mockDebitNotes,
          total: 2,
          page: 1,
          page_size: 15,
        });
      }
      return Promise.resolve({});
    });
  });

  it('renders header, action button, and KPI stat cards', async () => {
    render(<DebitNotesPage />);

    expect(screen.getByText('Debit Notes')).toBeInTheDocument();
    expect(screen.getByText('New Debit Note')).toBeInTheDocument();

    await waitFor(() => {
      expect(screen.getByText('Total Issued')).toBeInTheDocument();
      expect(screen.getByText('Outstanding')).toBeInTheDocument();
      expect(screen.getAllByText('Draft').length).toBeGreaterThanOrEqual(1);
      expect(screen.getAllByText('$945.00').length).toBe(2);
    });
  });

  it('renders debit notes table with records from backend', async () => {
    render(<DebitNotesPage />);

    await waitFor(() => {
      expect(screen.getByText('DN-20260907-0001')).toBeInTheDocument();
      expect(screen.getByText('Apex Global Logistics Corp')).toBeInTheDocument();
      expect(screen.getByText('DN-20260907-0002')).toBeInTheDocument();
      expect(screen.getByText('Euro-Asia Retailers Ltd')).toBeInTheDocument();
    });
  });

  it('opens create debit note modal on click', async () => {
    render(<DebitNotesPage />);

    const createBtn = screen.getByText('New Debit Note');
    fireEvent.click(createBtn);

    await waitFor(() => {
      expect(screen.getByText('Customer & Reference')).toBeInTheDocument();
    });
  });

  it('opens details panel when clicking a debit note row', async () => {
    render(<DebitNotesPage />);

    await waitFor(() => {
      expect(screen.getByText('DN-20260907-0001')).toBeInTheDocument();
    });

    const row = screen.getByText('DN-20260907-0001');
    fireEvent.click(row);

    await waitFor(() => {
      expect(screen.getByText('Port Demurrage - 3 Days')).toBeInTheDocument();
    });
  });
});

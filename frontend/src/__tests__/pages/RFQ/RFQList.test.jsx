import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import { MemoryRouter } from 'react-router-dom';
import RFQList from '../../../pages/dashboard/RFQ/RFQList';

describe('RFQList Component', () => {
  it('renders loading state with skeleton boxes', () => {
    render(
      <MemoryRouter>
        <RFQList isLoading={true} rfqs={[]} onRowClick={() => {}} />
      </MemoryRouter>
    );
    expect(document.querySelectorAll('.skeleton-row').length).toBeGreaterThan(0);
  });

  it('renders empty state when no rfqs are provided', () => {
    render(
      <MemoryRouter>
        <RFQList isLoading={false} rfqs={[]} onRowClick={() => {}} />
      </MemoryRouter>
    );
    expect(screen.getByText('No Active RFQs or Rate Inquiries')).toBeInTheDocument();
  });

  it('renders table with data and handles row clicks', () => {
    const mockRfqs = [
      {
        id: 1,
        rfq_number: 'RFQ-2026-001',
        customer_id: 99,
        stage: 'STAGE_RFQ_CREATED',
        origin: 'Miami',
        destination: 'London',
        items: []
      }
    ];

    const handleClick = vi.fn();
    render(
      <MemoryRouter>
        <RFQList isLoading={false} rfqs={mockRfqs} onRowClick={handleClick} />
      </MemoryRouter>
    );

    expect(screen.getByText('RFQ-2026-001')).toBeInTheDocument();
    expect(screen.getByText('Miami')).toBeInTheDocument();
    expect(screen.getByText('London')).toBeInTheDocument();

    fireEvent.click(screen.getByText('RFQ-2026-001'));
    expect(handleClick).toHaveBeenCalledWith(mockRfqs[0]);
  });
});


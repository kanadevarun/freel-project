import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import RFQList from '../../../pages/dashboard/RFQ/RFQList';

describe('RFQList Component', () => {
  it('renders loading state with skeleton boxes', () => {
    render(<RFQList isLoading={true} rfqs={[]} onRowClick={() => {}} />);
    expect(document.querySelectorAll('.skeleton-row').length).toBeGreaterThan(0);
  });

  it('renders empty state when no rfqs are provided', () => {
    render(<RFQList isLoading={false} rfqs={[]} onRowClick={() => {}} />);
    expect(screen.getByText('No RFQs found')).toBeInTheDocument();
  });

  it('renders table with data and handles row clicks', () => {
    const mockRfqs = [
      {
        id: 1,
        rfq_number: 'RFQ-2026-001',
        customer_id: 99,
        stage: 'STAGE_RFQ_CREATED',
        origin: 'Miami',
        destination: 'London'
      }
    ];

    const handleClick = vi.fn();
    render(<RFQList isLoading={false} rfqs={mockRfqs} onRowClick={handleClick} />);

    expect(screen.getByText('RFQ-2026-001')).toBeInTheDocument();
    expect(screen.getByText('Miami → London')).toBeInTheDocument();
    expect(screen.getByText('RFQ Created')).toBeInTheDocument();

    fireEvent.click(screen.getByText('RFQ-2026-001'));
    expect(handleClick).toHaveBeenCalledWith(mockRfqs[0]);
  });
});

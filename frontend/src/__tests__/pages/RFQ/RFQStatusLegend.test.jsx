import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import RFQStatusLegend from '../../../pages/dashboard/RFQ/components/RFQStatusLegend';

describe('RFQStatusLegend Component', () => {
  const sampleRFQs = [
    {
      id: '1',
      rfq_number: 'RFQ-2026-0001',
      origin: 'Shanghai (CNSHA)',
      destination: 'Hamburg (DEHAM)',
      stage: 'STAGE_RFQ_CREATED',
      items: [{ description: 'Electronics', weight_kg: 500, volume_cbm: 2.5 }],
      incoterms: 'FOB',
      target_date: '2026-09-15',
    },
    {
      id: '2',
      rfq_number: 'RFQ-2026-0002',
      origin: 'Ningbo (CNNGB)',
      destination: '',
      stage: 'DRAFT',
      items: [],
    },
  ];

  it('renders the header title and total inquiries count pill', () => {
    render(<RFQStatusLegend rfqs={sampleRFQs} activeTab="ALL" onSelectTab={vi.fn()} />);

    expect(screen.getByText('RFQ Workflow Lifecycle & Stage Intelligence')).toBeInTheDocument();
    expect(screen.getByText('2 Total Inquiries')).toBeInTheDocument();
  });

  it('renders stage definition cards with correct titles', () => {
    render(<RFQStatusLegend rfqs={sampleRFQs} activeTab="ALL" onSelectTab={vi.fn()} />);

    expect(screen.getByText('Information Required')).toBeInTheDocument();
    expect(screen.getByText('Ready for Quotation')).toBeInTheDocument();
    expect(screen.getByText('Pricing Assigned')).toBeInTheDocument();
    expect(screen.getByText('Quote Generated')).toBeInTheDocument();
    expect(screen.getByText('Won / Awarded')).toBeInTheDocument();
  });

  it('calls onSelectTab when a stage card is clicked', () => {
    const handleSelectTab = vi.fn();
    render(<RFQStatusLegend rfqs={sampleRFQs} activeTab="ALL" onSelectTab={handleSelectTab} />);

    const wonCard = screen.getByText('Won / Awarded').closest('.rfq-stage-item');
    expect(wonCard).toBeInTheDocument();

    fireEvent.click(wonCard);
    expect(handleSelectTab).toHaveBeenCalledWith('WON');
  });

  it('collapses and expands when the toggle button is clicked', () => {
    render(<RFQStatusLegend rfqs={sampleRFQs} activeTab="ALL" onSelectTab={vi.fn()} />);

    const toggleBtn = screen.getByText('Hide Guide');
    fireEvent.click(toggleBtn);

    expect(screen.getByText('Show Stage Guide')).toBeInTheDocument();
    expect(screen.queryByText('Ready for Quotation')).not.toBeInTheDocument();

    fireEvent.click(screen.getByText('Show Stage Guide'));
    expect(screen.getByText('Ready for Quotation')).toBeInTheDocument();
  });
});

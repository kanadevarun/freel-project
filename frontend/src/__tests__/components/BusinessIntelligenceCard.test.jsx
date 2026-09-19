import React from 'react';
import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { MemoryRouter } from 'react-router-dom';
import BusinessIntelligenceCard from '../../components/common/BusinessIntelligenceCard';
import api from '../../services/api';

vi.mock('../../services/api', () => ({
  default: {
    post: vi.fn(),
  },
}));

describe('BusinessIntelligenceCard Component', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  const mockInsight = {
    title: 'RFQ-2026-001 Intelligence',
    summary: 'RFQ is fully qualified for carrier pricing.',
    confidence_level: 'HIGH',
    is_informational: true,
    data_freshness: '2026-09-07T14:30:00Z',
    correlation_id: 'corr-test-123',
    key_highlights: [
      'Customer has 3 previous bookings.',
      'All customs documentation present.',
    ],
    warnings: [
      'Cargo Ready Date is within 48 hours.',
    ],
    supporting_records: [
      {
        entity_type: 'CUSTOMER',
        entity_id: 1,
        reference_number: 'CUST-001',
        status: 'ACTIVE',
        path: '/dashboard/customers/1',
      },
      {
        entity_type: 'QUOTATION',
        entity_id: 10,
        reference_number: 'QT-2026-010',
        status: 'PENDING_APPROVAL',
        path: '/dashboard/quotes/10',
      },
    ],
    supporting_field_references: [
      {
        entity_type: 'RFQ',
        entity_id: 1,
        field_name: 'origin_port',
        field_value: 'CNSHA',
      },
      {
        entity_type: 'RFQ',
        entity_id: 1,
        field_name: 'status',
        field_value: 'READY_FOR_QUOTATION',
      },
    ],
    recommended_follow_up: [
      'Review spot rates from ocean carrier.',
    ],
  };

  it('renders loading state initially then displays grounded insight data', async () => {
    api.post.mockResolvedValueOnce({ data: mockInsight });

    render(
      <MemoryRouter>
        <BusinessIntelligenceCard entityType="RFQ" entityId={1} />
      </MemoryRouter>
    );

    // Initial loading indicator
    expect(screen.getByText(/Reading related LogisticsHQ records/i)).toBeInTheDocument();

    // Wait for insight to load
    await waitFor(() => {
      expect(screen.getByText('RFQ-2026-001 Intelligence')).toBeInTheDocument();
    });

    // Verify summary and highlights
    expect(screen.getByText('RFQ is fully qualified for carrier pricing.')).toBeInTheDocument();
    expect(screen.getByText('Customer has 3 previous bookings.')).toBeInTheDocument();

    // Verify confidence badge and read-only status
    expect(screen.getByText('HIGH Confidence')).toBeInTheDocument();
    expect(screen.getByText(/All insights are organization-isolated, read-only/i)).toBeInTheDocument();

    // Verify supporting records links
    expect(screen.getByText('CUST-001')).toBeInTheDocument();
    expect(screen.getByText('QT-2026-010')).toBeInTheDocument();

    // Verify field citations
    expect(screen.getByText('origin_port:')).toBeInTheDocument();
    expect(screen.getByText('CNSHA')).toBeInTheDocument();

    // Verify warnings
    expect(screen.getByText('Cargo Ready Date is within 48 hours.')).toBeInTheDocument();
  });

  it('handles error state and provides working retry button', async () => {
    api.post.mockRejectedValueOnce(new Error('Network timeout'));

    render(
      <MemoryRouter>
        <BusinessIntelligenceCard entityType="RFQ" entityId={999} />
      </MemoryRouter>
    );

    await waitFor(() => {
      expect(screen.getByText('Network timeout')).toBeInTheDocument();
    });

    const retryBtn = screen.getByRole('button', { name: /retry/i });
    expect(retryBtn).toBeInTheDocument();

    // Mock success on retry
    api.post.mockResolvedValueOnce({ data: mockInsight });
    fireEvent.click(retryBtn);

    await waitFor(() => {
      expect(screen.getByText('RFQ-2026-001 Intelligence')).toBeInTheDocument();
    });
  });

  it('renders empty notice when API returns null or empty', async () => {
    api.post.mockResolvedValueOnce({ data: null });

    render(
      <MemoryRouter>
        <BusinessIntelligenceCard entityType="RFQ" entityId={555} />
      </MemoryRouter>
    );

    await waitFor(() => {
      expect(screen.getByText(/No cross-module context available for this record/i)).toBeInTheDocument();
    });
  });
});

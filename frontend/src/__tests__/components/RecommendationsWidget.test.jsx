import React from 'react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import ModuleRecommendationsWidget from '../../components/recommendations/ModuleRecommendationsWidget';
import DashboardRecommendationsWidget from '../../components/dashboard/DashboardRecommendationsWidget';
import { recommendationService } from '../../services/recommendationService';

vi.mock('../../services/recommendationService', () => ({
  recommendationService: {
    listBySource: vi.fn(),
    listRecommendations: vi.fn(),
    getStats: vi.fn(),
  },
}));

describe('ModuleRecommendationsWidget', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders grounded recommendations for specific source record', async () => {
    recommendationService.listBySource.mockResolvedValueOnce({
      success: true,
      data: {
        items: [
          {
            id: 101,
            title: 'Immediate Collections Action Needed',
            category: 'invoice',
            priority: 'critical',
            riskLevel: 'high',
            confidence: 0.95,
            status: 'new',
            requiresApproval: true,
            evidence: [{ key: 'invoiceNumber', label: 'Invoice #', value: 'INV-1001' }],
            recommendedAction: 'Contact customer accounts payable'
          }
        ]
      }
    });

    render(
      <MemoryRouter>
        <ModuleRecommendationsWidget sourceType="customer" sourceId={1} />
      </MemoryRouter>
    );

    await waitFor(() => {
      expect(screen.getByText('Immediate Collections Action Needed')).toBeInTheDocument();
    });

    expect(screen.getByText(/critical Priority/i)).toBeInTheDocument();
    expect(screen.getByText('0.95')).toBeInTheDocument();
  });

  it('handles empty state cleanly with light design', async () => {
    recommendationService.listBySource.mockResolvedValueOnce({
      success: true,
      data: { items: [] }
    });

    render(
      <MemoryRouter>
        <ModuleRecommendationsWidget sourceType="shipment" sourceId={999} />
      </MemoryRouter>
    );

    await waitFor(() => {
      expect(screen.getByText(/No active AI recommendations for this shipment/i)).toBeInTheDocument();
    });
  });
});

describe('DashboardRecommendationsWidget', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders top urgent recommendations on executive dashboard', async () => {
    recommendationService.getStats.mockResolvedValueOnce({
      success: true,
      data: { total: 5, new: 2, critical: 1, requiresApproval: 1 }
    });
    recommendationService.listRecommendations.mockResolvedValueOnce({
      success: true,
      data: {
        items: [
          {
            id: 201,
            title: 'Unresolved Shipment Exception',
            category: 'shipment',
            priority: 'critical',
            riskLevel: 'critical',
            status: 'new',
            requiresApproval: true,
            evidence: [{ key: 'shipmentNumber', label: 'Shipment', value: 'SH-8821' }],
            recommendedAction: 'Notify carrier and escalate to customs desk'
          }
        ]
      }
    });

    render(
      <MemoryRouter>
        <DashboardRecommendationsWidget />
      </MemoryRouter>
    );

    await waitFor(() => {
      expect(screen.getByText('Unresolved Shipment Exception')).toBeInTheDocument();
    });

    expect(screen.getByText('Open Recommendation Center')).toBeInTheDocument();
  });
});

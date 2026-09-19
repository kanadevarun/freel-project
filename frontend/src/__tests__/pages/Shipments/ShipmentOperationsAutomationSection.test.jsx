import React from 'react';
import { render, screen, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import ShipmentOperationsAutomationSection from '../../../pages/dashboard/Shipments/components/ShipmentOperationsAutomationSection';
import shipmentOperationsAutomationService from '../../../services/shipmentOperationsAutomationService';

vi.mock('../../../services/shipmentOperationsAutomationService', () => ({
  default: {
    getOverview: vi.fn(),
    analyzeRisks: vi.fn(),
    prioritizeExceptions: vi.fn(),
    getRecommendations: vi.fn(),
    generateDraft: vi.fn(),
    updateDraft: vi.fn(),
    submitForApproval: vi.fn(),
  },
}));

describe('ShipmentOperationsAutomationSection', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders loading state initially', () => {
    shipmentOperationsAutomationService.getOverview.mockReturnValue(new Promise(() => {}));
    render(<ShipmentOperationsAutomationSection shipmentId={101} />);
    expect(screen.getByTestId('sh-auto-loading')).toBeInTheDocument();
  });

  it('renders deterministic signals and operational KPIs upon successful fetch', async () => {
    const mockOverview = {
      shipment_id: 101,
      org_id: 2,
      booking_number: 'BKG-101',
      status: 'IN_TRANSIT',
      signals: {
        risk_level: 'CRITICAL',
        risk_score: 85,
        has_overdue_milestone: true,
        overdue_milestones: ['DEPARTURE'],
        approaching_milestones: ['ARRIVAL'],
        is_tracking_stale: true,
        hours_since_last_tracking: 36.5,
        active_exceptions_count: 2,
        critical_exceptions_count: 1,
        missing_documents_count: 1,
        missing_documents: ['COMMERCIAL_INVOICE'],
        requires_approval: true,
      },
      latest_analysis: {
        risk_level: 'CRITICAL',
        risk_score: 85,
        operational_summary: 'High risk detected due to port customs examination hold and stale carrier telemetry.',
        key_risks: ['CUSTOMS_HOLD', 'STALE_TELEMETRY'],
        confidence_score: 0.92,
        correlation_id: 'corr-test-101',
      },
      prioritized_exceptions: [
        {
          priority_rank: 1,
          urgency: 'IMMEDIATE',
          business_impact: 'Container detained at customs terminal',
          action_deadline: 'Immediate action within 4 hours',
          rationale: 'Missing commercial invoice triggered customs verification',
          recommended_action: 'Request urgent document review from customer freight desk',
        },
      ],
      recommendations: [
        {
          title: 'Request Carrier Telemetry Follow-up',
          description: 'Ping carrier EDI gateway for updated container position',
          risk_rating: 'HIGH',
          requires_approval: true,
          target_action: 'shipments.request_carrier_followup',
        },
      ],
      latest_draft: null,
    };

    shipmentOperationsAutomationService.getOverview.mockResolvedValue({ data: { data: mockOverview } });

    render(<ShipmentOperationsAutomationSection shipmentId={101} />);

    await waitFor(() => {
      expect(screen.getByTestId('sh-operations-automation-section')).toBeInTheDocument();
    });

    // Check KPI titles and values
    expect(screen.getByText('OPERATIONAL RISK')).toBeInTheDocument();
    expect(screen.getByText('85 / 100')).toBeInTheDocument();
    expect(screen.getByText('1 Overdue')).toBeInTheDocument();
    expect(screen.getByText('36.5h ago')).toBeInTheDocument();
    expect(screen.getByText('2 Open')).toBeInTheDocument();
    expect(screen.getByText('1 Missing')).toBeInTheDocument();

    // Check AI summary
    expect(screen.getByText(/High risk detected due to port customs examination hold/)).toBeInTheDocument();

    // Check prioritized exceptions
    expect(screen.getByText('Container detained at customs terminal')).toBeInTheDocument();

    // Check recommendations
    expect(screen.getByText('Request Carrier Telemetry Follow-up')).toBeInTheDocument();
  });
});

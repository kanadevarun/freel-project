import React from 'react';
import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import ShipmentOperationsIntelligenceSection from '../../pages/dashboard/Shipments/components/ShipmentOperationsIntelligenceSection';
import { shipmentService } from '../../services/shipmentService';

vi.mock('../../services/shipmentService', () => ({
  shipmentService: {
    getShipment360OperationsIntelligence: vi.fn(),
  },
}));

describe('ShipmentOperationsIntelligenceSection Component', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  const mockCriticalPayload = {
    shipment_id: 101,
    org_id: 2,
    correlation_id: 'corr-sh-test-101',
    data_freshness_timestamp: '2026-09-07T14:00:00Z',
    identity: {
      shipment_id: 101,
      org_id: 2,
      shipment_number: 'SH-101',
      status: 'DEPARTED',
      origin_port: 'INNSA',
      destination_port: 'NLRTM',
      carrier_scac: 'MAEU',
      carrier_name: 'Maersk Line',
      vessel_name: 'Maersk Mc-Kinney Moller',
      voyage_number: '2601W',
      booking_number: 'BK-2026-DEV-001',
      etd: '2026-08-20T10:00:00Z',
      eta: '2026-09-15T18:00:00Z',
      container_count: 2,
    },
    milestones: {
      total_milestones: 4,
      completed_milestones: 4,
      pending_milestones: 0,
      overdue_milestones: 0,
      upcoming_milestones: 0,
      milestone_completion_rate: 100.0,
      last_completed_milestone_code: 'ARRIVAL',
      next_expected_milestone_code: '',
      number_of_delayed_milestones: 2,
      stale_tracking_flag: false,
      milestones_list: [
        {
          id: 101,
          milestone_code: 'GATE_IN',
          status: 'COMPLETED',
          location: 'INNSA',
          planned_date: '2026-08-18T08:00:00Z',
          actual_date: '2026-08-18T09:30:00Z',
          is_delayed: false,
        },
        {
          id: 103,
          milestone_code: 'DEPARTED',
          status: 'COMPLETED',
          location: 'INNSA',
          planned_date: '2026-08-20T10:00:00Z',
          actual_date: '2026-09-06T18:00:00Z',
          is_delayed: true,
          delay_hours: 416.0,
        },
      ],
    },
    exceptions: {
      total_exceptions: 2,
      open_exceptions: 2,
      resolved_exceptions: 0,
      critical_exceptions: 1,
      high_severity_exceptions: 1,
      oldest_unresolved_title: 'Weather Disruption',
      oldest_unresolved_hours_open: 20.5,
      exceptions_list: [
        {
          id: 103,
          exception_type: 'OTHER',
          severity: 'CRITICAL',
          status: 'OPEN',
          title: 'Weather Disruption',
          description: 'Severe weather at transshipment port.',
          resolved: false,
          hours_open: 20.5,
          created_at: '2026-09-06T17:38:38Z',
        },
        {
          id: 104,
          exception_type: 'PORT_CONGESTION',
          severity: 'HIGH',
          status: 'OPEN',
          title: 'Port Congestion Warning',
          description: 'Berth congestion exceeding 48 hours.',
          resolved: false,
          hours_open: 17.5,
          created_at: '2026-09-06T20:45:20Z',
        },
      ],
    },
    performance: {
      on_time_departure: false,
      departure_variance_hours: 416.0,
      on_time_arrival: true,
      arrival_variance_hours: -219.5,
      planned_transit_days: 26.3,
      actual_transit_days: 17.0,
      delay_duration_hours: 416.0,
      number_of_delays: 2,
      carrier_update_freshness: 'RECENT',
    },
    risk_indicators: {
      overall_risk_rating: 'CRITICAL',
      risk_score: 65,
      shipment_delayed: true,
      eta_missing_or_stale: false,
      critical_milestone_overdue: false,
      unresolved_high_exception: true,
      stale_tracking_data: false,
      active_risk_factors: [
        'Shipment delayed (2 delayed milestone(s))',
        '1 unresolved CRITICAL operational exception(s)',
      ],
      missing_data_reasons: [],
    },
    ai_summary: {
      executive_summary: 'Shipment SH-101 (INNSA to NLRTM) via Maersk Line is DEPARTED with 100.0% milestone completion.',
      current_status_explanation: 'All recorded milestones are completed.',
      milestone_progress_explanation: '2 milestone(s) completed late.',
      likely_operational_risks: 'Shipment delayed. 1 unresolved CRITICAL operational exception.',
      exception_prioritization: '2 open exception(s) require operations review. Priority: 1 Critical, 1 High.',
      actionable_attention_items: ['Escalate critical exception to carrier line operations manager.'],
      suggested_operator_inquiries: ['Inquire with carrier for updated vessel schedule and revised destination ETA.'],
      confidence_score: 'HIGH',
      verifiable_source_citations: ['[Shipment: #101]', '[Carrier: MAEU]', '[Exception: #103 (CRITICAL)]'],
    },
    grounded_observations: [
      {
        category: 'STATUS',
        severity: 'INFO',
        message: 'Shipment SH-101 is currently in status DEPARTED with milestone progress of 100.0%.',
        evidence: 'Shipment #101 status=DEPARTED, 4/4 milestones completed',
      },
      {
        category: 'RISK',
        severity: 'CRITICAL',
        message: 'Operational health flagged as CRITICAL (Risk Score: 65/100).',
        evidence: 'Risk factors evaluated deterministically against live milestone & exception records',
      },
    ],
  };

  it('renders loading state initially', () => {
    shipmentService.getShipment360OperationsIntelligence.mockReturnValue(new Promise(() => {}));

    render(<ShipmentOperationsIntelligenceSection shipmentId={101} />);
    expect(screen.getByTestId('sh-intel-loading')).toBeInTheDocument();
  });

  it('renders error state when fetch fails and allows retry', async () => {
    shipmentService.getShipment360OperationsIntelligence.mockRejectedValueOnce(new Error('Connection timeout'));

    render(<ShipmentOperationsIntelligenceSection shipmentId={101} />);

    await waitFor(() => {
      expect(screen.getByTestId('sh-intel-error')).toBeInTheDocument();
    });

    expect(screen.getByText('Operational Intelligence Unavailable')).toBeInTheDocument();

    // Mock success for retry
    shipmentService.getShipment360OperationsIntelligence.mockResolvedValueOnce({ data: mockCriticalPayload });
    fireEvent.click(screen.getByText(/Retry Calculation/i));

    await waitFor(() => {
      expect(screen.getByTestId('shipment-operations-intelligence-section')).toBeInTheDocument();
    });
  });

  it('renders complete operational intelligence with milestones, schedule variance, critical exceptions, and AI summary', async () => {
    shipmentService.getShipment360OperationsIntelligence.mockResolvedValueOnce({ data: mockCriticalPayload });

    render(<ShipmentOperationsIntelligenceSection shipmentId={101} />);

    await waitFor(() => {
      expect(screen.getByTestId('shipment-operations-intelligence-section')).toBeInTheDocument();
    });

    // Check header, badges, and risk pill
    expect(screen.getByText('READ-ONLY OPERATIONS INTELLIGENCE')).toBeInTheDocument();
    expect(screen.getByTestId('sh-intel-risk-rating')).toHaveTextContent('Risk: CRITICAL (65/100)');
    expect(screen.getByText('Tracking: RECENT')).toBeInTheDocument();

    // Check Grounded AI summary card
    expect(screen.getByTestId('sh-intel-ai-card')).toBeInTheDocument();
    expect(screen.getByText(/Shipment SH-101 \(INNSA to NLRTM\) via Maersk Line is DEPARTED/i)).toBeInTheDocument();
    expect(screen.getByText('Confidence: HIGH')).toBeInTheDocument();
    expect(screen.getByText('[Shipment: #101]')).toBeInTheDocument();
    expect(screen.getByText('[Carrier: MAEU]')).toBeInTheDocument();

    // Check KPI metrics
    expect(screen.getByTestId('sh-intel-milestones-completed')).toHaveTextContent('4');
    expect(screen.getByTestId('sh-intel-completion-rate')).toHaveTextContent('100%');
    expect(screen.getByTestId('sh-intel-open-exceptions')).toHaveTextContent('2');
    expect(screen.getByTestId('sh-intel-critical-exceptions')).toHaveTextContent('1');
    expect(screen.getByTestId('sh-intel-departure-variance')).toHaveTextContent('+17.3d');

    // Check Milestones Table
    expect(screen.getByTestId('sh-intel-milestones-table-card')).toBeInTheDocument();
    expect(screen.getByText('GATE_IN')).toBeInTheDocument();
    expect(screen.getByText('DEPARTED')).toBeInTheDocument();
    expect(screen.getByText('+17.3d late')).toBeInTheDocument();
  });

  it('handles clean shipment with 0 exceptions gracefully', async () => {
    const cleanPayload = {
      ...mockCriticalPayload,
      shipment_id: 105,
      identity: {
        ...mockCriticalPayload.identity,
        shipment_id: 105,
        shipment_number: 'SH-105',
        status: 'IN_TRANSIT',
      },
      exceptions: {
        total_exceptions: 0,
        open_exceptions: 0,
        resolved_exceptions: 0,
        critical_exceptions: 0,
        high_severity_exceptions: 0,
        exceptions_list: [],
      },
      risk_indicators: {
        overall_risk_rating: 'LOW',
        risk_score: 0,
        shipment_delayed: false,
        active_risk_factors: [],
        missing_data_reasons: [],
      },
    };

    shipmentService.getShipment360OperationsIntelligence.mockResolvedValueOnce({ data: cleanPayload });

    render(<ShipmentOperationsIntelligenceSection shipmentId={105} />);

    await waitFor(() => {
      expect(screen.getByTestId('sh-intel-risk-rating')).toHaveTextContent('Risk: LOW (0/100)');
    });

    expect(screen.getByTestId('sh-intel-open-exceptions')).toHaveTextContent('0');
    expect(screen.getByTestId('sh-intel-critical-exceptions')).toHaveTextContent('0');
  });
});

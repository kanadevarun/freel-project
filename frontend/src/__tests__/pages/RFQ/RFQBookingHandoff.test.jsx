import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { MemoryRouter } from 'react-router-dom';
import RFQBookingHandoff from '../../../pages/dashboard/RFQ/components/RFQBookingHandoff';
import RFQShipmentHandoff from '../../../pages/dashboard/RFQ/components/RFQShipmentHandoff';
import { rfqService } from '../../../services/rfqService';

vi.mock('../../../services/rfqService', () => ({
  rfqService: {
    getRFQBookings: vi.fn(),
    createRFQBooking: vi.fn(),
    updateRFQBookingStatus: vi.fn(),
    getRFQShipments: vi.fn(),
  },
}));

describe('RFQBookingHandoff Component', () => {
  const mockRFQ = {
    id: 201,
    rfq_number: 'RFQ-2026-0201',
    origin: 'INNSA',
    destination: 'DEHAM',
    items: [{ quantity: 2, description: '40HC Machinery' }],
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders blocked state with readiness checklist and recommended action', () => {
    const ineligibleData = {
      eligibility: {
        is_eligible: false,
        missing_prerequisites: ['Commercial Quote Approval Required'],
        commercial_closure_status: 'PENDING',
      },
      bookings: [],
    };

    render(
      <MemoryRouter>
        <RFQBookingHandoff
          rfq={mockRFQ}
          bookingHandoffData={ineligibleData}
          onSwitchTab={vi.fn()}
        />
      </MemoryRouter>
    );

    expect(screen.getByText(/Booking Readiness Checklist/i)).toBeInTheDocument();
    expect(screen.getByText('Commercial Quote Approval')).toBeInTheDocument();
    expect(screen.getByText(/Recommended Operational Action/i)).toBeInTheDocument();
    expect(screen.getByText(/Review & Approve Carrier Quotes/i)).toBeInTheDocument();
  });

  it('renders eligible state with Ready to Create Carrier Booking summary', () => {
    const eligibleData = {
      eligibility: {
        is_eligible: true,
        approved_carrier: 'Maersk',
        approved_quote_id: 10,
        missing_prerequisites: [],
        commercial_closure_status: 'APPROVED',
      },
      bookings: [],
    };

    render(
      <MemoryRouter>
        <RFQBookingHandoff
          rfq={mockRFQ}
          bookingHandoffData={eligibleData}
          onSwitchTab={vi.fn()}
        />
      </MemoryRouter>
    );

    expect(screen.getByText(/Ready to Create Carrier Booking/i)).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /\+ Create Carrier Booking/i })).toBeInTheDocument();
  });

  it('renders live booking summary card and handles status transitions', async () => {
    const dataWithBooking = {
      eligibility: {
        is_eligible: true,
        approved_carrier: 'Hapag-Lloyd',
        approved_quote_id: 11,
        missing_prerequisites: [],
        commercial_closure_status: 'APPROVED',
      },
      summary: {
        total_bookings: 1,
        active_booking: {
          id: 501,
          booking_number: 'BK-2026-TEST501',
          carrier_name: 'Hapag-Lloyd',
          carrier_scac: 'HLCU',
          status: 'DRAFT',
          origin_port: 'INNSA',
          destination_port: 'DEHAM',
          created_at: new Date().toISOString(),
        },
      },
      bookings: [
        {
          id: 501,
          booking_number: 'BK-2026-TEST501',
          carrier_name: 'Hapag-Lloyd',
          carrier_scac: 'HLCU',
          status: 'DRAFT',
          origin_port: 'INNSA',
          destination_port: 'DEHAM',
          created_at: new Date().toISOString(),
        },
      ],
    };

    rfqService.updateRFQBookingStatus.mockResolvedValue({
      data: { id: 501, status: 'REQUESTED' },
    });

    const mockMutation = vi.fn();

    render(
      <MemoryRouter>
        <RFQBookingHandoff
          rfq={mockRFQ}
          bookingHandoffData={dataWithBooking}
          onSwitchTab={vi.fn()}
          onMutationSuccess={mockMutation}
        />
      </MemoryRouter>
    );

    expect(screen.getByText('BK-2026-TEST501')).toBeInTheDocument();
    expect(screen.getAllByText('Hapag-Lloyd').length).toBeGreaterThan(0);
    expect(screen.getByText('Open in Bookings Workspace')).toBeInTheDocument();

    const requestBtn = screen.getByRole('button', { name: /Request Carrier Booking/i });
    expect(requestBtn).toBeInTheDocument();

    fireEvent.click(requestBtn);

    await waitFor(() => {
      expect(rfqService.updateRFQBookingStatus).toHaveBeenCalledWith(201, 501, {
        status: 'REQUESTED',
      });
      expect(mockMutation).toHaveBeenCalled();
    });
  });
});

describe('RFQShipmentHandoff Component', () => {
  const mockRFQ = {
    id: 201,
    rfq_number: 'RFQ-2026-0201',
    origin: 'INNSA',
    destination: 'DEHAM',
  };

  it('renders shipment dependency flow when no shipment exists', () => {
    const handoffData = {
      source_booking: {
        booking_number: 'BK-2026-001',
        status: 'CONFIRMED',
      },
      summary: { total_shipments: 0 },
      shipments: [],
    };

    render(
      <MemoryRouter>
        <RFQShipmentHandoff
          rfq={mockRFQ}
          shipmentHandoffData={handoffData}
          onSwitchTab={vi.fn()}
        />
      </MemoryRouter>
    );

    expect(screen.getByText(/Shipment Execution Dependency Flow/i)).toBeInTheDocument();
    expect(screen.getByText('BK-2026-001')).toBeInTheDocument();
    expect(screen.getAllByRole('button', { name: /Open Shipments Workspace/i }).length).toBeGreaterThan(0);
  });

  it('renders live shipment execution details when shipment exists', () => {
    const handoffData = {
      source_booking: {
        booking_number: 'BK-2026-001',
        status: 'CONFIRMED',
      },
      summary: {
        total_shipments: 1,
        active_shipment: {
          id: 99,
          carrier_name: 'Maersk Line',
          carrier_scac: 'MAEU',
          status: 'IN_TRANSIT',
          origin_port: 'INNSA',
          destination_port: 'DEHAM',
          vessel_name: 'Maersk Mc-Kinney Moller',
          voyage_number: '2601W',
          container_numbers: ['MSKU1234567', 'MSKU7654321'],
        },
      },
      shipments: [
        {
          id: 99,
          carrier_name: 'Maersk Line',
          carrier_scac: 'MAEU',
          status: 'IN_TRANSIT',
          origin_port: 'INNSA',
          destination_port: 'DEHAM',
        },
      ],
    };

    render(
      <MemoryRouter>
        <RFQShipmentHandoff
          rfq={mockRFQ}
          shipmentHandoffData={handoffData}
          onSwitchTab={vi.fn()}
        />
      </MemoryRouter>
    );

    expect(screen.getByText('Shipment #99')).toBeInTheDocument();
    expect(screen.getByText('Maersk Line')).toBeInTheDocument();
    expect(screen.getByText('IN_TRANSIT')).toBeInTheDocument();
    expect(screen.getByText('MSKU1234567')).toBeInTheDocument();
    expect(screen.getByText('Open Full Shipment Workspace')).toBeInTheDocument();
  });
});

import React from 'react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import { BrowserRouter } from 'react-router-dom';
import BookingsPage from '../../../pages/dashboard/Bookings/BookingsPage';
import bookingService from '../../../services/bookingService';

vi.mock('../../../services/bookingService', () => ({
  default: {
    getBookings: vi.fn(),
    getEligibleRFQs: vi.fn(),
    createBooking: vi.fn(),
    updateBookingStatus: vi.fn(),
    createShipment: vi.fn(),
  },
}));

describe('BookingsPage Component', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders KPI metric cards and booking records from database', async () => {
    bookingService.getBookings.mockResolvedValueOnce({
      data: {
        data: [
          {
            id: 501,
            booking_number: 'BKG-MAEU-99001',
            rfq_id: 101,
            rfq_number: 'RFQ-2026-001',
            customer_name: 'Apex Global Logistics',
            carrier_name: 'Maersk Line',
            carrier_scac: 'MAEU',
            origin_port: 'CNSHA',
            destination_port: 'USLAX',
            vessel_name: 'Maersk Mc-Kinney',
            voyage_number: '2608E',
            status: 'CONFIRMED',
            quote_sell_price: 4500,
            created_at: new Date().toISOString(),
          },
        ],
        kpis: {
          total_bookings: 1,
          draft: 0,
          requested: 0,
          pending_confirmation: 0,
          confirmed: 1,
          completed: 0,
          cancelled: 0,
          departing_soon: 1,
        },
        pagination: {
          current_page: 1,
          page_size: 10,
          total_items: 1,
          total_pages: 1,
        },
      },
    });

    render(
      <BrowserRouter>
        <BookingsPage />
      </BrowserRouter>
    );

    expect(screen.getByText(/Loading Carrier Bookings Workspace.../i)).toBeInTheDocument();

    await waitFor(() => {
      expect(screen.getByText('Carrier Bookings')).toBeInTheDocument();
      expect(screen.getByText('BKG-MAEU-99001')).toBeInTheDocument();
      expect(screen.getByText('Apex Global Logistics')).toBeInTheDocument();
      expect(screen.getByText('Maersk Line')).toBeInTheDocument();
    });
  });

  it('opens create booking modal and displays eligible RFQs', async () => {
    bookingService.getBookings.mockResolvedValueOnce({
      data: {
        data: [],
        kpis: { total_bookings: 0 },
        pagination: { total_items: 0 },
      },
    });

    bookingService.getEligibleRFQs.mockResolvedValueOnce({
      data: {
        data: [
          {
            rfq_id: 202,
            rfq_number: 'RFQ-ELIGIBLE-1',
            customer_name: 'Pacific Traders',
            carrier_name: 'Hapag-Lloyd',
            carrier_scac: 'HLCU',
            origin_port: 'INNSA',
            destination_port: 'DEHAM',
            sell_price: 3800,
            currency: 'USD',
            approved_quote_id: 88,
            cargo_description: 'Industrial Machinery',
            total_weight_kg: 12000,
          },
        ],
      },
    });

    render(
      <BrowserRouter>
        <BookingsPage />
      </BrowserRouter>
    );

    await waitFor(() => {
      expect(screen.getByText(/\+ Create Booking/i)).toBeInTheDocument();
    });

    fireEvent.click(screen.getByText(/\+ Create Booking/i));

    await waitFor(() => {
      expect(screen.getByText(/Select Eligible RFQ with Approved Quote/i)).toBeInTheDocument();
      expect(screen.getByText(/RFQ-ELIGIBLE-1/i)).toBeInTheDocument();
      expect(screen.getByText(/Pacific Traders/i)).toBeInTheDocument();
    });
  });
});

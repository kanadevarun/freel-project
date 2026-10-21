import React from 'react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import CarrierDocumentsCard from '../../../pages/dashboard/Bookings/CarrierDocumentsCard';
import bookingService from '../../../services/bookingService';

vi.mock('../../../services/bookingService', () => ({
  default: {
    getCarrierDocuments: vi.fn(),
  },
}));

describe('CarrierDocumentsCard Component', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders loading state initially while fetching documents', async () => {
    // Keep promise pending
    bookingService.getCarrierDocuments.mockReturnValue(new Promise(() => {}));

    render(
      <CarrierDocumentsCard
        bookingId={101}
        carrierName="Maersk Line"
        carrierBookingRef="BKG-9921"
      />
    );

    expect(screen.getByTestId('documents-loading')).toBeInTheDocument();
    expect(screen.getByText(/Retrieving live transport documents from Maersk Line/i)).toBeInTheDocument();
  });

  it('renders successful document list with document type, number, date, and download button', async () => {
    bookingService.getCarrierDocuments.mockResolvedValueOnce({
      data: {
        data: [
          {
            document_id: 'doc-12345',
            document_type: 'Bill of Lading',
            reference: 'MSK-BL-990182',
            carrier_scac: 'MAEU',
            document_url: 'https://api.maersk.com/transportation-documents/v1/download/MSK-BL-990182',
            mime_type: 'application/pdf',
            created_at: '2026-08-15T10:00:00Z',
          },
          {
            document_id: 'doc-67890',
            document_type: 'Sea Waybill',
            reference: 'MSK-WB-55412',
            carrier_scac: 'MAEU',
            document_url: 'https://api.maersk.com/transportation-documents/v1/download/MSK-WB-55412',
            mime_type: 'application/pdf',
            created_at: '2026-08-16T12:00:00Z',
          },
        ],
      },
    });

    render(
      <CarrierDocumentsCard
        bookingId={101}
        carrierName="Maersk Line"
        carrierBookingRef="BKG-9921"
      />
    );

    await waitFor(() => {
      expect(screen.getByTestId('documents-list')).toBeInTheDocument();
    });

    const items = screen.getAllByTestId('document-item');
    expect(items).toHaveLength(2);

    expect(screen.getByText('Bill of Lading')).toBeInTheDocument();
    expect(screen.getByText('MSK-BL-990182')).toBeInTheDocument();
    expect(screen.getByText('Sea Waybill')).toBeInTheDocument();
    expect(screen.getByText('MSK-WB-55412')).toBeInTheDocument();

    const links = screen.getAllByTestId('document-download-link');
    expect(links).toHaveLength(2);
    expect(links[0]).toHaveAttribute('href', 'https://api.maersk.com/transportation-documents/v1/download/MSK-BL-990182');
    expect(links[0]).toHaveAttribute('target', '_blank');
    expect(links[0]).toHaveAttribute('rel', 'noopener noreferrer');
  });

  it('renders clean empty state when no documents are returned', async () => {
    bookingService.getCarrierDocuments.mockResolvedValueOnce({
      data: {
        data: [],
      },
    });

    render(
      <CarrierDocumentsCard
        bookingId={102}
        carrierName="Maersk Line"
        carrierBookingRef="BKG-EMPTY-01"
      />
    );

    await waitFor(() => {
      expect(screen.getByTestId('documents-empty')).toBeInTheDocument();
    });

    expect(screen.getByText('No documents available.')).toBeInTheDocument();
    expect(screen.getByText(/No electronic transport documents have been issued by Maersk Line yet for booking BKG-EMPTY-01/i)).toBeInTheDocument();
  });

  it('renders unsupported capability state when carrier does not support documents', async () => {
    bookingService.getCarrierDocuments.mockRejectedValueOnce({
      response: {
        status: 400,
        data: {
          error: {
            code: 'UNSUPPORTED_OPERATION',
            message: 'carrier adapter does not support capability: DOCUMENTS',
          },
        },
      },
    });

    render(
      <CarrierDocumentsCard
        bookingId={103}
        carrierName="MSC"
        carrierBookingRef="MSC-12345"
      />
    );

    await waitFor(() => {
      expect(screen.getByTestId('documents-unsupported')).toBeInTheDocument();
    });

    expect(screen.getByText('Documents Capability Not Supported')).toBeInTheDocument();
    expect(screen.getByText(/Electronic transport document retrieval is not supported by MSC/i)).toBeInTheDocument();
  });

  it('renders error state and allows retry on failure', async () => {
    bookingService.getCarrierDocuments.mockRejectedValueOnce({
      response: {
        status: 500,
        data: {
          error: {
            code: 'CARRIER_UNAVAILABLE',
            message: 'Carrier service temporarily unavailable',
          },
        },
      },
    });

    render(
      <CarrierDocumentsCard
        bookingId={104}
        carrierName="Maersk Line"
        carrierBookingRef="BKG-ERR-01"
      />
    );

    await waitFor(() => {
      expect(screen.getByTestId('documents-error')).toBeInTheDocument();
    });

    expect(screen.getByText('Carrier service temporarily unavailable')).toBeInTheDocument();
    expect(screen.getByText(/Code: CARRIER_UNAVAILABLE/i)).toBeInTheDocument();

    // Now mock recovery on retry
    bookingService.getCarrierDocuments.mockResolvedValueOnce({
      data: {
        data: [
          {
            document_id: 'doc-retry-success',
            document_type: 'Waybill',
            reference: 'MSK-WB-999',
            document_url: 'https://api.maersk.com/transportation-documents/v1/download/MSK-WB-999',
            created_at: '2026-08-20T00:00:00Z',
          },
        ],
      },
    });

    const retryBtn = screen.getByTestId('retry-documents-btn');
    fireEvent.click(retryBtn);

    await waitFor(() => {
      expect(screen.getByTestId('documents-list')).toBeInTheDocument();
    });

    expect(screen.getByText('MSK-WB-999')).toBeInTheDocument();
  });
});

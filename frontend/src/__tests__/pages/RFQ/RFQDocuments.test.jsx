import React from 'react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import RFQDocuments from '../../../pages/dashboard/RFQ/components/RFQDocuments';
import { rfqService } from '../../../services/rfqService';

vi.mock('../../../services/rfqService', () => ({
  rfqService: {
    createDocument: vi.fn(),
    updateDocumentStatus: vi.fn(),
    deleteDocument: vi.fn(),
  },
}));

describe('RFQDocuments Component', () => {
  const mockRfq = {
    id: 101,
    rfq_number: 'RFQ-20260827-001',
    origin: 'INNSA',
    destination: 'DEHAM',
    incoterms: 'FOB',
  };

  const mockDocumentsData = {
    summary: {
      total_documents: 1,
      required_documents: 2,
      received_documents: 0,
      missing_documents: 2,
      under_review_documents: 1,
      approved_documents: 0,
      rejected_documents: 0,
      future_stage_documents: 4,
      readiness_percentage: 0,
    },
    current_stage_documents: [
      {
        doc_type: 'COMMERCIAL_INVOICE',
        title: 'Commercial Invoice',
        is_required: true,
        is_conditional: false,
        status: 'UNDER_REVIEW',
        reason: 'Required for customs clearance and valuation.',
        document_record: {
          id: 501,
          document_type: 'COMMERCIAL_INVOICE',
          document_name: 'Commercial Invoice - 2026.pdf',
          status: 'UNDER_REVIEW',
          file_name: 'ci_test.pdf',
          uploaded_by: 'John Operator',
        },
      },
      {
        doc_type: 'PACKING_LIST',
        title: 'Packing List',
        is_required: true,
        is_conditional: false,
        status: 'MISSING',
        reason: 'Required for cargo breakdown and port inspection.',
        document_record: null,
      },
    ],
    conditional_documents: [],
    future_stage_documents: [
      {
        doc_type: 'BILL_OF_LADING',
        title: 'Bill of Lading (OBL / HBL / MBL)',
        is_required: false,
        is_conditional: false,
        status: 'NOT_APPLICABLE',
        reason: 'Required at shipment execution.',
        document_record: null,
      },
    ],
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders summary metrics and document sections properly', () => {
    render(<RFQDocuments rfq={mockRfq} documentsData={mockDocumentsData} onMutationSuccess={vi.fn()} />);

    expect(screen.getByTestId('rfq-documents-workspace')).toBeInTheDocument();
    expect(screen.getByText('Document Management')).toBeInTheDocument();
    expect(screen.getByText('Commercial Invoice - 2026.pdf')).toBeInTheDocument();
    expect(screen.getByText('Packing List')).toBeInTheDocument();
    expect(screen.getByText('Bill of Lading (OBL / HBL / MBL)')).toBeInTheDocument();

  });

  it('allows clicking Approve on an UNDER_REVIEW document', async () => {
    const onMutationSuccess = vi.fn();
    rfqService.updateDocumentStatus.mockResolvedValueOnce({ data: { id: 501, status: 'APPROVED' } });

    render(<RFQDocuments rfq={mockRfq} documentsData={mockDocumentsData} onMutationSuccess={onMutationSuccess} />);

    const approveBtn = screen.getByTestId('approve-501-btn');
    expect(approveBtn).toBeInTheDocument();
    fireEvent.click(approveBtn);

    await waitFor(() => {
      expect(rfqService.updateDocumentStatus).toHaveBeenCalledWith(101, 501, {
        status: 'APPROVED',
        rejection_reason: undefined,
      });
      expect(onMutationSuccess).toHaveBeenCalled();
    });
  });

  it('opens Add Document modal and creates a document', async () => {
    const onMutationSuccess = vi.fn();
    rfqService.createDocument.mockResolvedValueOnce({ data: { id: 502, document_type: 'PACKING_LIST' } });

    render(<RFQDocuments rfq={mockRfq} documentsData={mockDocumentsData} onMutationSuccess={onMutationSuccess} />);

    const addBtn = screen.getByTestId('add-document-btn');
    fireEvent.click(addBtn);

    expect(screen.getByText('Attach / Upload Document')).toBeInTheDocument();
    const submitBtn = screen.getByTestId('submit-add-doc-btn');
    fireEvent.click(submitBtn);

    await waitFor(() => {
      expect(rfqService.createDocument).toHaveBeenCalled();
      expect(onMutationSuccess).toHaveBeenCalled();
    });
  });
});

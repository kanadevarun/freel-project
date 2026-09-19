import React from 'react';
import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import ContractComplianceAutomationSection from '../../../pages/dashboard/Contracts/ContractComplianceAutomationSection';
import contractComplianceAutomationService from '../../../services/contractComplianceAutomationService';

vi.mock('../../../services/contractComplianceAutomationService', () => ({
  default: {
    getOverview: vi.fn(),
    reviewContract: vi.fn(),
    extractClauses: vi.fn(),
    verifyStructuredTerms: vi.fn(),
    getComplianceChecklist: vi.fn(),
    generateDraft: vi.fn(),
    updateDraft: vi.fn(),
    submitApproval: vi.fn(),
  },
}));

describe('ContractComplianceAutomationSection', () => {
  const mockContract = {
    id: 103,
    contract_reference: 'VEND-DRAY-2025-Q3',
    party_name: 'Apex Drayage Services',
    contract_type: 'CARRIER_AGREEMENT',
    status: 'ACTIVE',
    effective_date: '2025-09-25T00:00:00Z',
    expiry_date: '2026-09-25T00:00:00Z'
  };

  const mockOverview = {
    contract_id: 103,
    org_id: 1,
    contract_reference: 'VEND-DRAY-2025-Q3',
    signals: {
      is_expiring_soon: true,
      days_until_expiry: 17,
      is_expired: false,
      compliance_risk_detected: true,
      missing_required_compliance_docs: ['CERTIFICATE_OF_INSURANCE']
    },
    primary_document_title: 'Apex_Drayage_MSA_2025.pdf',
    document_count: 1,
    recent_reviews: [
      {
        risk_level: 'MEDIUM',
        risk_score: 55,
        compliance_status: 'ACTION_REQUIRED',
        executive_summary: 'Agreement expiring within 30 days and Certificate of Insurance is pending upload.',
        confidence_score: 0.94,
        extracted_clauses: [
          {
            clause_type: 'PAYMENT_TERMS',
            raw_text_snippet: 'LogisticsHQ agrees to remit payment Net 30 days from invoice approval.',
            summary: 'Net 30 days payment window from invoice approval.',
            source_document: 'Apex_Drayage_MSA_2025.pdf',
            section_reference: 'Section 4'
          }
        ],
        structured_discrepancies: [],
        compliance_obligations: [],
        recommendations: []
      }
    ],
    recent_drafts: [
      {
        id: 1,
        draft_type: 'MISSING_DOCUMENT_REQUEST',
        recipient_name: 'Apex Drayage Services',
        recipient_email: 'compliance@apexdrayage.com',
        subject: 'ACTION REQUIRED: Missing Certificate of Insurance',
        message_body: 'Please provide an updated COI before the upcoming renewal date.',
        status: 'DRAFT',
        approval_id: null
      }
    ]
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders loading state initially', () => {
    contractComplianceAutomationService.getOverview.mockReturnValue(new Promise(() => {}));
    render(<ContractComplianceAutomationSection contract={mockContract} />);
    expect(screen.getByText('Loading Contract & Compliance Intelligence...')).toBeInTheDocument();
  });

  it('renders deterministic contract signals and AI review data', async () => {
    contractComplianceAutomationService.getOverview.mockResolvedValue({ data: { data: mockOverview } });
    render(<ContractComplianceAutomationSection contract={mockContract} />);

    await waitFor(() => {
      expect(screen.getByTestId('contract-compliance-automation-section')).toBeInTheDocument();
    });

    // Check header elements
    expect(screen.getByText(/Contract Compliance & Intelligent Document Review/i)).toBeInTheDocument();
    expect(screen.getByText(/VEND-DRAY-2025-Q3/i)).toBeInTheDocument();
    expect(screen.getByText(/Expiring in 17 days/i)).toBeInTheDocument();
    expect(screen.getAllByText(/Apex_Drayage_MSA_2025.pdf/i).length).toBeGreaterThan(0);

    // Check AI summary
    expect(screen.getByText(/AI Executive Summary/i)).toBeInTheDocument();
    expect(screen.getByText(/Certificate of Insurance is pending upload/i)).toBeInTheDocument();

    // Check disclaimer
    expect(screen.getByText(/AI recommendation for operational support only/i)).toBeInTheDocument();
  });

  it('displays extracted clauses when tab is active', async () => {
    contractComplianceAutomationService.getOverview.mockResolvedValue({ data: { data: mockOverview } });
    render(<ContractComplianceAutomationSection contract={mockContract} />);

    await waitFor(() => {
      expect(screen.getByText(/PAYMENT TERMS/i)).toBeInTheDocument();
      expect(screen.getByText(/LogisticsHQ agrees to remit payment Net 30 days/i)).toBeInTheDocument();
      expect(screen.getByText(/Section 4/i)).toBeInTheDocument();
    });
  });

  it('allows switching to Drafts tab and displays draft communication with submit approval button', async () => {
    contractComplianceAutomationService.getOverview.mockResolvedValue({ data: { data: mockOverview } });
    render(<ContractComplianceAutomationSection contract={mockContract} />);

    await waitFor(() => {
      expect(screen.getByTestId('contract-compliance-automation-section')).toBeInTheDocument();
    });

    const draftsTab = screen.getByText(/Clarification Drafts/i);
    fireEvent.click(draftsTab);

    expect(screen.getByText(/ACTION REQUIRED: Missing Certificate of Insurance/i)).toBeInTheDocument();
    expect(screen.getByText(/Submit for HITL Approval/i)).toBeInTheDocument();
  });
});

import React from 'react';
import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import RFQIntelligentPricingWorkflowSection from '../../../pages/dashboard/RFQ/components/RFQIntelligentPricingWorkflowSection';
import { rfqPricingWorkflowService } from '../../../services/rfqPricingWorkflowService';

vi.mock('../../../services/rfqPricingWorkflowService', () => ({
  rfqPricingWorkflowService: {
    getOverview: vi.fn(),
    extractRequirements: vi.fn(),
    getPricingPreview: vi.fn(),
    generateDraft: vi.fn(),
    updateDraft: vi.fn(),
    submitForApproval: vi.fn(),
  },
}));

vi.mock('react-hot-toast', () => ({
  default: {
    success: vi.fn(),
    error: vi.fn(),
  },
}));

describe('RFQIntelligentPricingWorkflowSection Component', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  const mockOverviewData = {
    rfq_id: 101,
    rfq_number: 'RFQ-2026-0101',
    customer_name: 'Apex Global Logistics',
    origin: 'INNSA',
    destination: 'DEHAM',
    shipment_mode: 'OCEAN_FCL',
    incoterms: 'FOB',
    extraction: {
      status: 'COMPLETE',
      confidence_score: 0.98,
      clarification_needed: false,
    },
    pricing_preview: {
      currency: 'USD',
      base_cost: 2200.0,
      surcharges: 350.0,
      total_cost: 2550.0,
      base_sell: 2750.0,
      discounts: 0.0,
      tax_amount: 0.0,
      total_selling_price: 3100.0,
      gross_profit: 550.0,
      gross_margin_pct: 17.74,
      margin_health: 'HEALTHY',
      applied_carrier_name: 'Maersk Line',
      rate_is_expired: false,
    },
    latest_draft: {
      id: 55,
      rfq_id: 101,
      status: 'DRAFT',
      customer_wording: 'Dear Valued Customer, LogisticsHQ presents our rate proposal...',
      internal_summary: 'Internal review summary for RFQ 101...',
      terms_and_conditions: 'Standard commercial freight terms apply.',
      recipient_name: 'John Doe',
      recipient_email: 'john@apex.com',
      requires_approval: false,
    },
    pending_approval: false,
    correlation_id: 'corr-test-overview-101',
  };

  it('renders loading state initially', () => {
    rfqPricingWorkflowService.getOverview.mockReturnValue(new Promise(() => {}));
    render(<RFQIntelligentPricingWorkflowSection rfqId={101} />);
    expect(screen.getByText(/Loading RFQ-to-Quotation Pricing Workflow/i)).toBeInTheDocument();
  });

  it('renders overview metrics, requirements, and deterministic pricing', async () => {
    rfqPricingWorkflowService.getOverview.mockResolvedValue({ data: mockOverviewData });
    render(<RFQIntelligentPricingWorkflowSection rfqId={101} />);

    await waitFor(() => {
      expect(screen.getByTestId('rfq-intelligent-pricing-workflow')).toBeInTheDocument();
    });

    // KPI values
    expect(screen.getByText('COMPLETE')).toBeInTheDocument();
    expect(screen.getByText('USD 2,550.00')).toBeInTheDocument();
    expect(screen.getByText('USD 3,100.00')).toBeInTheDocument();
    expect(screen.getByText('17.74%')).toBeInTheDocument();

    // Ports
    expect(screen.getByText('INNSA')).toBeInTheDocument();
    expect(screen.getByText('DEHAM')).toBeInTheDocument();
    expect(screen.getByText('FOB')).toBeInTheDocument();
  });

  it('allows switching between draft workspace tabs', async () => {
    rfqPricingWorkflowService.getOverview.mockResolvedValue({ data: mockOverviewData });
    render(<RFQIntelligentPricingWorkflowSection rfqId={101} />);

    await waitFor(() => {
      expect(screen.getByTestId('customer-wording-input')).toBeInTheDocument();
    });

    // Switch to Internal Summary tab
    fireEvent.click(screen.getByText('Internal Pricing Summary'));
    expect(screen.getByTestId('internal-summary-input')).toBeInTheDocument();

    // Switch to Terms tab
    fireEvent.click(screen.getByText('Terms & Conditions'));
    expect(screen.getByTestId('terms-conditions-input')).toBeInTheDocument();
  });

  it('calls extractRequirements when Re-Evaluate is clicked', async () => {
    rfqPricingWorkflowService.getOverview.mockResolvedValue({ data: mockOverviewData });
    rfqPricingWorkflowService.extractRequirements.mockResolvedValue({ data: { status: 'COMPLETE' } });

    render(<RFQIntelligentPricingWorkflowSection rfqId={101} />);
    await waitFor(() => {
      expect(screen.getByText('Re-Evaluate')).toBeInTheDocument();
    });

    fireEvent.click(screen.getByText('Re-Evaluate'));
    expect(rfqPricingWorkflowService.extractRequirements).toHaveBeenCalledWith(101);
  });

  it('calls generateDraft when Generate AI Draft is clicked', async () => {
    rfqPricingWorkflowService.getOverview.mockResolvedValue({ data: mockOverviewData });
    rfqPricingWorkflowService.generateDraft.mockResolvedValue({
      data: {
        ...mockOverviewData.latest_draft,
        customer_wording: 'Updated draft wording',
      },
    });

    render(<RFQIntelligentPricingWorkflowSection rfqId={101} />);
    await waitFor(() => {
      expect(screen.getByText(/Generate AI Draft/i)).toBeInTheDocument();
    });

    fireEvent.click(screen.getByText(/Generate AI Draft/i));
    expect(rfqPricingWorkflowService.generateDraft).toHaveBeenCalledWith(
      101,
      expect.objectContaining({
        recipient_name: 'John Doe',
        recipient_email: 'john@apex.com',
      })
    );
  });

  it('calls submitForApproval when Submit for Manager Approval is clicked', async () => {
    rfqPricingWorkflowService.getOverview.mockResolvedValue({ data: mockOverviewData });
    rfqPricingWorkflowService.submitForApproval.mockResolvedValue({
      data: {
        ...mockOverviewData.latest_draft,
        status: 'PENDING_APPROVAL',
        approval_id: 888,
      },
    });

    render(<RFQIntelligentPricingWorkflowSection rfqId={101} />);
    await waitFor(() => {
      expect(screen.getByText(/Submit for Manager Approval/i)).toBeInTheDocument();
    });

    fireEvent.click(screen.getByText(/Submit for Manager Approval/i));
    expect(rfqPricingWorkflowService.submitForApproval).toHaveBeenCalledWith(
      101,
      55,
      expect.any(String)
    );
  });

  it('renders error alert when API call fails', async () => {
    rfqPricingWorkflowService.getOverview.mockRejectedValue(new Error('Network error loading workflow'));
    render(<RFQIntelligentPricingWorkflowSection rfqId={101} />);

    await waitFor(() => {
      expect(screen.getByText(/Network error loading workflow/i)).toBeInTheDocument();
    });
  });
});

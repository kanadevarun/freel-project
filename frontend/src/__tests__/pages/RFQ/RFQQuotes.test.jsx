import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import RFQQuotes from '../../../pages/dashboard/RFQ/components/RFQQuotes';
import { rfqService } from '../../../services/rfqService';

vi.mock('../../../services/rfqService', () => ({
  rfqService: {
    getQuotes: vi.fn(),
    createQuote: vi.fn(),
    updateQuote: vi.fn(),
    updateQuoteStatus: vi.fn(),
    recommendQuote: vi.fn(),
    approveQuote: vi.fn(),
    selectQuoteForCustomer: vi.fn(),
    deleteQuote: vi.fn(),
  },
}));

describe('RFQQuotes Component', () => {
  const mockRFQ = {
    id: 156,
    rfq_number: 'RFQ-20260827-001',
    customer_id: 101,
    stage: 'STAGE_RFQ_CREATED',
  };

  const mockQuotesData = {
    summary: {
      total_quotes: 3,
      received_quotes: 2,
      under_review_quotes: 0,
      recommended_quotes: 1,
      approved_quotes: 0,
      selected_quotes: 0,
      expired_quotes: 0,
      quotes_expiring_soon: 1,
      lowest_buy_amount: 4200.0,
      highest_margin_amount: 1150.0,
      highest_margin_percentage: 21.3,
      fastest_transit_days: 18,
      recommended_quote_id: 101,
      approved_quote_id: null,
      primary_currency: 'USD',
      has_mixed_currencies: false,
    },
    rfq_readiness: {
      overall_status: 'READY_FOR_QUOTATION',
      readiness_score: 100,
      blocking_count: 0,
      next_best_action: 'Proceed to commercial review and quote selection.',
    },
    quotes: [
      {
        id: 101,
        rfq_id: 156,
        carrier_name: 'Maersk Line',
        carrier_id: 'MAEU',
        quote_reference: 'MSK-QT-2026-08',
        currency: 'USD',
        buy_price: 4200.0,
        sell_price: 5350.0,
        margin_amount: 1150.0,
        margin_percentage: 21.5,
        transit_time_days: 21,
        free_days: 7,
        validity_status: 'VALID',
        days_until_expiry: 14,
        is_recommended: true,
        status: 'RECOMMENDED',
        charges: [
          { type: 'FREIGHT', description: 'Base Ocean Freight', amount: 3800.0, currency: 'USD' },
          { type: 'FUEL_SURCHARGE', description: 'Bunker BAF', amount: 400.0, currency: 'USD' },
        ],
      },
      {
        id: 102,
        rfq_id: 156,
        carrier_name: 'Hapag-Lloyd',
        carrier_id: 'HLCU',
        quote_reference: 'HL-QT-991',
        currency: 'USD',
        buy_price: 4400.0,
        sell_price: 5400.0,
        margin_amount: 1000.0,
        margin_percentage: 18.5,
        transit_time_days: 18,
        free_days: 14,
        validity_status: 'EXPIRING_SOON',
        days_until_expiry: 3,
        is_recommended: false,
        status: 'RECEIVED',
        charges: [],
      },
    ],
    comparison: [
      {
        quote_id: 101,
        carrier_name: 'Maersk Line',
        is_lowest_cost: true,
        is_highest_margin: true,
        is_fastest: false,
        is_recommended: true,
        recommendation_reason: 'Optimal commercial value: Highest margin (21.5%) and lowest carrier buy cost (USD 4200.00).',
      },
      {
        quote_id: 102,
        carrier_name: 'Hapag-Lloyd',
        is_lowest_cost: false,
        is_highest_margin: false,
        is_fastest: true,
        is_recommended: false,
        recommendation_reason: 'Fastest transit time among valid quotes with positive commercial margin.',
      },
    ],
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders quotes workspace header, KPI strip, and comparison table', () => {
    render(
      <RFQQuotes
        rfq={mockRFQ}
        quotesData={mockQuotesData}
        requirements={mockQuotesData.rfq_readiness}
      />
    );

    expect(screen.getByText('Carrier Quotes & Commercial Decision')).toBeInTheDocument();
    expect(screen.getByText(/2 Quotes Persisted/i)).toBeInTheDocument();

    // Check KPI strip values
    expect(screen.getByText('LOWEST BUY')).toBeInTheDocument();
    expect(screen.getAllByText('USD 4,200.00').length).toBeGreaterThanOrEqual(1);
    expect(screen.getByText('BEST MARGIN')).toBeInTheDocument();
    expect(screen.getAllByText('+USD 1,150.00').length).toBeGreaterThanOrEqual(1);
    expect(screen.getAllByText('18 Days').length).toBeGreaterThanOrEqual(1);

    // Check Table rows
    expect(screen.getByTestId('quote-row-101')).toBeInTheDocument();
    expect(screen.getByText('Maersk Line')).toBeInTheDocument();
    expect(screen.getByText('Hapag-Lloyd')).toBeInTheDocument();
  });

  it('renders empty state when quotes array is empty', () => {
    const emptyQuotesData = {
      summary: { total_quotes: 0 },
      quotes: [],
      comparison: [],
    };

    render(
      <RFQQuotes
        rfq={mockRFQ}
        quotesData={emptyQuotesData}
      />
    );

    expect(screen.getByText('No Carrier Quotes Recorded')).toBeInTheDocument();
    expect(screen.getByTestId('empty-add-quote-btn')).toBeInTheDocument();
  });

  it('opens details drawer when Details button is clicked', async () => {
    render(
      <RFQQuotes
        rfq={mockRFQ}
        quotesData={mockQuotesData}
      />
    );

    const detailsBtn = screen.getByTestId('view-quote-btn-101');
    fireEvent.click(detailsBtn);

    expect(screen.getByTestId('quote-detail-drawer')).toBeInTheDocument();
    expect(screen.getByText('Commercial Margin Breakdown')).toBeInTheDocument();
    expect(screen.getByText('Base Ocean Freight')).toBeInTheDocument();
    expect(screen.getByText('Bunker BAF')).toBeInTheDocument();
  });

  it('opens progressive add quote modal and advances through steps', async () => {
    render(
      <RFQQuotes
        rfq={mockRFQ}
        quotesData={mockQuotesData}
      />
    );

    const addBtn = screen.getByTestId('add-carrier-quote-btn');
    fireEvent.click(addBtn);

    expect(screen.getByTestId('add-quote-modal')).toBeInTheDocument();
    expect(screen.getByText('Step 1 of 4 · Carrier & Identification')).toBeInTheDocument();

    // Fill Step 1
    const carrierInput = screen.getByTestId('input-carrier-name');
    fireEvent.change(carrierInput, { target: { value: 'CMA CGM' } });

    const nextBtn = screen.getByTestId('modal-next-btn');
    fireEvent.click(nextBtn);

    // Step 2
    expect(screen.getByText('Step 2 of 4 · Commercial Pricing & Surcharges')).toBeInTheDocument();
    const buyInput = screen.getByTestId('input-buy-price');
    const sellInput = screen.getByTestId('input-sell-price');
    fireEvent.change(buyInput, { target: { value: '4500' } });
    fireEvent.change(sellInput, { target: { value: '5500' } });

    fireEvent.click(screen.getByTestId('modal-next-btn'));

    // Step 3
    expect(screen.getByText('Step 3 of 4 · Operational Schedule')).toBeInTheDocument();
  });

  it('calls recommendQuote when Recommend button is clicked', async () => {
    rfqService.recommendQuote.mockResolvedValueOnce({ data: { success: true } });
    const mockOnSuccess = vi.fn();

    render(
      <RFQQuotes
        rfq={mockRFQ}
        quotesData={mockQuotesData}
        onMutationSuccess={mockOnSuccess}
      />
    );

    const recommendBtn = screen.getByTestId('recommend-btn-102');
    fireEvent.click(recommendBtn);

    await waitFor(() => {
      expect(rfqService.recommendQuote).toHaveBeenCalledWith(156, 102);
      expect(mockOnSuccess).toHaveBeenCalled();
    });
  });

  it('calls approveQuote when Approve button is clicked', async () => {
    rfqService.approveQuote.mockResolvedValueOnce({ data: { success: true } });
    const mockOnSuccess = vi.fn();

    render(
      <RFQQuotes
        rfq={mockRFQ}
        quotesData={mockQuotesData}
        onMutationSuccess={mockOnSuccess}
      />
    );

    const approveBtn = screen.getByTestId('approve-btn-101');
    fireEvent.click(approveBtn);

    await waitFor(() => {
      expect(rfqService.approveQuote).toHaveBeenCalledWith(156, 101, expect.any(Object));
      expect(mockOnSuccess).toHaveBeenCalled();
    });
  });
});

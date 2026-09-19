import React from 'react';
import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import RFQPricingIntelligenceSection from '../../pages/dashboard/RFQ/components/RFQPricingIntelligenceSection';
import { rfqService } from '../../services/rfqService';

vi.mock('../../services/rfqService', () => ({
  rfqService: {
    getRFQ360PricingIntelligence: vi.fn(),
  },
}));

describe('RFQPricingIntelligenceSection Component', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  const mockPayloadWithQuotes = {
    rfq_id: 1,
    org_id: 1,
    correlation_id: 'corr-test-12345678',
    data_freshness_timestamp: '2026-03-01T10:00:00Z',
    identity: {
      rfq_id: 1,
      rfq_number: 'RFQ-2026-0001',
      customer_id: 1,
      customer_name: 'Apex Global',
      status: 'WON',
      origin_port: 'INNSA',
      destination_port: 'DEHAM',
      shipment_mode: 'FCL',
      requested_incoterms: 'FOB',
      requested_service_level: 'STANDARD',
      currency: 'USD',
      completeness_percentage: 100,
    },
    quotation_comparison: {
      quotations_requested_count: 2,
      quotations_received_count: 2,
      valid_quotations_count: 2,
      lowest_quotation_amount: 2580.0,
      highest_quotation_amount: 2784.0,
      average_quotation_amount: 2682.0,
      median_quotation_amount: 2682.0,
      price_spread: 204.0,
      price_spread_percentage: 7.9,
      selected_quotation_id: 1,
      selected_carrier_name: 'Maersk Line',
      selected_quotation_amount: 2580.0,
      time_to_first_quote_hours: 1.5,
      carrier_quotes: [
        {
          id: 1,
          carrier_name: 'Maersk Line',
          carrier_code: 'MAEU',
          buy_price: 2150.0,
          sell_price: 2580.0,
          margin_amount: 430.0,
          margin_percentage: 16.7,
          currency: 'USD',
          transit_days: 28,
          status: 'RECEIVED',
          is_lowest: true,
          is_selected: true,
        },
        {
          id: 2,
          carrier_name: 'Hapag-Lloyd',
          carrier_code: 'HLCU',
          buy_price: 2320.0,
          sell_price: 2784.0,
          margin_amount: 464.0,
          margin_percentage: 16.7,
          currency: 'USD',
          transit_days: 26,
          status: 'RECEIVED',
          is_lowest: false,
          is_selected: false,
        },
      ],
    },
    commercial_performance: {
      rfq_to_quotation_conversion: true,
      rfq_to_booking_conversion: true,
      quotation_win_rate: 1.0,
      carrier_response_rate: 1.0,
      lane_historical_quotes_count: 2,
      lane_historical_average_price: 2682.0,
    },
    margin_and_risk: {
      quoted_revenue_amount: 2580.0,
      estimated_or_recorded_cost: 2150.0,
      gross_margin_amount: 430.0,
      gross_margin_percentage: 16.7,
      margin_health_rating: 'HEALTHY',
      price_anomaly_detected: false,
      missing_data_reasons: [],
    },
    ai_summary: {
      executive_summary: 'RFQ #1 has 2 valid carrier quotations received for lane INNSA to DEHAM.',
      quotation_spread_analysis: 'Competitive spread of $204.00 (7.9%) between Maersk Line and Hapag-Lloyd.',
      carrier_response_insights: 'Maersk Line provided the lowest rate with 28 transit days.',
      commercial_risks: 'Gross margin is healthy at 16.7%.',
      actionable_attention_items: ['Maersk quote is selected and linked to active booking.'],
      suggested_operator_inquiries: ['Verify free-time detention at destination port DEHAM.'],
      confidence_score: 'HIGH',
      verifiable_source_citations: ['[RFQ: #1]', '[Quotation: #1]'],
    },
    grounded_observations: [
      {
        category: 'MARGIN',
        severity: 'INFO',
        message: 'Gross margin is 16.7%, meeting minimum commercial target.',
        evidence: 'Quotation #1 Buy: $2,150.00 Sell: $2,580.00',
      },
    ],
  };

  it('renders loading state initially', () => {
    rfqService.getRFQ360PricingIntelligence.mockReturnValue(new Promise(() => {}));

    render(<RFQPricingIntelligenceSection rfqId={1} />);
    expect(screen.getByTestId('rfq-intel-loading')).toBeInTheDocument();
  });

  it('renders error state when fetch fails and allows retry', async () => {
    rfqService.getRFQ360PricingIntelligence.mockRejectedValueOnce(new Error('Network error'));

    render(<RFQPricingIntelligenceSection rfqId={1} />);

    await waitFor(() => {
      expect(screen.getByTestId('rfq-intel-error')).toBeInTheDocument();
    });

    expect(screen.getByText('Intelligence Unavailable')).toBeInTheDocument();

    // Mock success for retry
    rfqService.getRFQ360PricingIntelligence.mockResolvedValueOnce({ data: mockPayloadWithQuotes });
    fireEvent.click(screen.getByText(/Retry Calculation/i));

    await waitFor(() => {
      expect(screen.getByTestId('rfq-pricing-intelligence-section')).toBeInTheDocument();
    });
  });

  it('renders comprehensive pricing intelligence with quotes, spread, and AI summary', async () => {
    rfqService.getRFQ360PricingIntelligence.mockResolvedValueOnce({ data: mockPayloadWithQuotes });

    render(<RFQPricingIntelligenceSection rfqId={1} />);

    await waitFor(() => {
      expect(screen.getByTestId('rfq-pricing-intelligence-section')).toBeInTheDocument();
    });

    // Check header and badges
    expect(screen.getByText('READ-ONLY PRICING INTELLIGENCE')).toBeInTheDocument();
    expect(screen.getByText('Margin: HEALTHY')).toBeInTheDocument();

    // Check Grounded AI summary card
    expect(screen.getByTestId('rfq-intel-ai-card')).toBeInTheDocument();
    expect(screen.getByText(/RFQ #1 has 2 valid carrier quotations received/i)).toBeInTheDocument();
    expect(screen.getByText('Confidence: HIGH')).toBeInTheDocument();
    expect(screen.getByText('[RFQ: #1]')).toBeInTheDocument();

    // Check KPIs
    expect(screen.getByTestId('rfq-intel-quotes-received')).toHaveTextContent('2');
    expect(screen.getByTestId('rfq-intel-price-spread')).toHaveTextContent('$204.00');
    expect(screen.getByTestId('rfq-intel-lowest-price')).toHaveTextContent('$2,580.00');
    expect(screen.getByTestId('rfq-intel-selected-price')).toHaveTextContent('$2,580.00');
    expect(screen.getByTestId('rfq-intel-gross-margin')).toHaveTextContent('$430.00');
    expect(screen.getByTestId('rfq-intel-margin-rating')).toHaveTextContent('HEALTHY');

    // Check Carrier Quotation Table
    expect(screen.getByTestId('rfq-intel-quotes-table-card')).toBeInTheDocument();
    expect(screen.getByText('Maersk Line')).toBeInTheDocument();
    expect(screen.getByText('Hapag-Lloyd')).toBeInTheDocument();
    expect(screen.getByText('Lowest')).toBeInTheDocument();
    expect(screen.getByText('Selected')).toBeInTheDocument();
  });

  it('renders empty quotes state gracefully when RFQ has 0 quotes', async () => {
    const unquotedPayload = {
      rfq_id: 4,
      org_id: 1,
      correlation_id: 'corr-unquoted-4',
      data_freshness_timestamp: '2026-03-01T10:00:00Z',
      identity: {
        rfq_id: 4,
        rfq_number: 'RFQ-2026-0004',
        origin_port: 'INBOM',
        destination_port: 'USLAX',
        currency: 'USD',
        completeness_percentage: 85,
      },
      quotation_comparison: {
        quotations_requested_count: 0,
        quotations_received_count: 0,
        valid_quotations_count: 0,
        lowest_quotation_amount: null,
        highest_quotation_amount: null,
        average_quotation_amount: null,
        price_spread: null,
        carrier_quotes: [],
      },
      commercial_performance: {
        rfq_to_quotation_conversion: false,
        rfq_to_booking_conversion: false,
      },
      margin_and_risk: {
        quoted_revenue_amount: null,
        estimated_or_recorded_cost: null,
        gross_margin_amount: null,
        margin_health_rating: 'UNKNOWN',
        missing_data_reasons: ['No carrier quotations submitted yet'],
      },
      ai_summary: {
        executive_summary: 'RFQ #4 currently has 0 carrier quotations.',
        confidence_score: 'LOW',
      },
      grounded_observations: [],
    };

    rfqService.getRFQ360PricingIntelligence.mockResolvedValueOnce({ data: unquotedPayload });

    render(<RFQPricingIntelligenceSection rfqId={4} />);

    await waitFor(() => {
      expect(screen.getByTestId('rfq-intel-empty-quotes')).toBeInTheDocument();
    });

    expect(screen.getByText('No Carrier Quotes Recorded')).toBeInTheDocument();
    expect(screen.getByTestId('rfq-intel-margin-rating')).toHaveTextContent('UNKNOWN');
  });
});

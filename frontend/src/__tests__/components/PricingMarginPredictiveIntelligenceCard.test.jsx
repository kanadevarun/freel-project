import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import PricingMarginPredictiveIntelligenceCard from '../../components/predictions/PricingMarginPredictiveIntelligenceCard';
import * as predictionService from '../../services/predictionService';

const mockGetRFQPredictedMargin = vi.fn();
const mockRefreshRFQPredictedMargin = vi.fn();
const mockGetContractPredictedRatePressure = vi.fn();
const mockRefreshContractPredictedRatePressure = vi.fn();
const mockAcknowledgePrediction = vi.fn();
const mockRequestAction = vi.fn();

vi.mock('../../services/predictionService', () => {
  const service = {
    getRFQPredictedMargin: (...args) => mockGetRFQPredictedMargin(...args),
    refreshRFQPredictedMargin: (...args) => mockRefreshRFQPredictedMargin(...args),
    getContractPredictedRatePressure: (...args) => mockGetContractPredictedRatePressure(...args),
    refreshContractPredictedRatePressure: (...args) => mockRefreshContractPredictedRatePressure(...args),
    acknowledgePrediction: (...args) => mockAcknowledgePrediction(...args),
    requestAction: (...args) => mockRequestAction(...args),
  };
  return {
    default: service,
    predictionService: service,
    getRFQPredictedMargin: service.getRFQPredictedMargin,
    refreshRFQPredictedMargin: service.refreshRFQPredictedMargin,
    getContractPredictedRatePressure: service.getContractPredictedRatePressure,
    refreshContractPredictedRatePressure: service.refreshContractPredictedRatePressure,
    acknowledgePrediction: service.acknowledgePrediction,
    requestAction: service.requestAction,
  };
});

describe('PricingMarginPredictiveIntelligenceCard', () => {
  const mockRFQPrediction = {
    prediction_id: 'pred-price-rfq-101-test',
    module: 'pricing',
    prediction_type: 'QUOTATION_COMPETITIVENESS',
    prediction_category: 'QUOTATION_COMPETITIVENESS',
    related_record_type: 'RFQ',
    related_record_id: '101',
    severity: 'LOW',
    confidence_score: 0.94,
    confidence_band: 'HIGH',
    prediction_statement: 'Healthy Margin & Strong Competitiveness: Projected gross margin of 16.7% ($550.00).',
    explanation: 'Quoted carrier rate balances strong profitability with market win probability.',
    time_horizon: '7_DAYS',
    status: 'PUBLISHED',
    supporting_signals: [
      { signal_name: 'carrier_buy_price', observed_value: '$2,750.00', baseline_value: 'N/A', importance_weight: 0.95 },
      { signal_name: 'customer_sell_price', observed_value: '$3,300.00', baseline_value: 'N/A', importance_weight: 0.95 }
    ],
    source_references: [
      { source_module: 'rfq_quotes', source_record_id: '9112', source_field: 'rfq_quotes.buy_price', source_timestamp: '2026-09-10T10:00:00Z' }
    ],
    source_timestamp: '2026-09-10T10:00:00Z',
    recommended_action: 'Proceed with quotation dispatch to customer.',
    is_action_required: false,
    requires_approval: false,
    model_version: 'v4.5-pricing-rules-llm'
  };

  const mockContractPrediction = {
    prediction_id: 'pred-ctr-101-test',
    module: 'contracts',
    prediction_type: 'CONTRACT_RATE_PRESSURE',
    prediction_category: 'CONTRACT_RATE_PRESSURE',
    related_record_type: 'CONTRACT',
    related_record_id: '101',
    severity: 'HIGH',
    confidence_score: 0.92,
    confidence_band: 'HIGH',
    prediction_statement: 'Imminent Tariff Expiry: Agreement #CTR-2026-001 expires in 15 days.',
    explanation: 'Quotations for shipments departing next month may exceed current contracted buying tariffs.',
    time_horizon: '30_DAYS',
    status: 'PUBLISHED',
    supporting_signals: [
      { signal_name: 'days_until_expiry', observed_value: '15', baseline_value: '90+ days', importance_weight: 0.92 }
    ],
    source_references: [
      { source_module: 'contracts', source_record_id: '101', source_field: 'contracts.expiry_date', source_timestamp: '2026-09-10T10:00:00Z' }
    ],
    source_timestamp: '2026-09-10T10:00:00Z',
    recommended_action: 'Engage commercial representative to extend rate matrix for next quarter.',
    is_action_required: true,
    requires_approval: true,
    model_version: 'v4.5-pricing-rules-llm'
  };

  const mockInsufficientDataPrediction = {
    prediction_id: 'pred-price-rfq-103-test',
    module: 'pricing',
    prediction_type: 'RFQ_MARGIN_RISK',
    prediction_category: 'INSUFFICIENT_DATA',
    related_record_type: 'RFQ',
    related_record_id: '103',
    severity: 'LOW',
    confidence_score: 0.20,
    confidence_band: 'LOW',
    prediction_statement: 'Insufficient Pricing Telemetry: No carrier rates or customer quotations have been generated.',
    predicted_value: 'UNPRICED',
    insufficient_data: true,
    explanation: '0 carrier quote records found. Margin risk modeling requires at least one active carrier tariff.',
    supporting_signals: [],
    source_references: [],
    source_timestamp: '2026-09-10T10:00:00Z',
    status: 'PUBLISHED'
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders loading state initially', () => {
    mockGetRFQPredictedMargin.mockReturnValue(new Promise(() => {}));
    render(<PricingMarginPredictiveIntelligenceCard recordType="rfq" recordId="101" />);
    expect(screen.getByText(/Evaluating real-time margin risk & quotation competitiveness/i)).toBeInTheDocument();
  });

  it('renders RFQ pricing prediction correctly', async () => {
    mockGetRFQPredictedMargin.mockResolvedValue({ data: mockRFQPrediction });
    render(<PricingMarginPredictiveIntelligenceCard recordType="rfq" recordId="101" />);

    await waitFor(() => {
      expect(screen.getByText(/Predictive Pricing & Margin Intelligence/i)).toBeInTheDocument();
      expect(screen.getByText(/Healthy Margin & Strong Competitiveness/i)).toBeInTheDocument();
      expect(screen.getByText(/CARRIER BUY PRICE/i)).toBeInTheDocument();
      expect(screen.getByText('$2,750.00')).toBeInTheDocument();
    });
  });

  it('renders Contract rate pressure prediction correctly', async () => {
    mockGetContractPredictedRatePressure.mockResolvedValue({ data: mockContractPrediction });
    render(<PricingMarginPredictiveIntelligenceCard recordType="contract" recordId="101" />);

    await waitFor(() => {
      expect(screen.getByText(/Predictive Contract & Tariff Intelligence/i)).toBeInTheDocument();
      expect(screen.getByText(/Imminent Tariff Expiry/i)).toBeInTheDocument();
      expect(screen.getByText(/DAYS UNTIL EXPIRY/i)).toBeInTheDocument();
    });
  });

  it('safely handles insufficient data without errors', async () => {
    mockGetRFQPredictedMargin.mockResolvedValue({ data: mockInsufficientDataPrediction });
    render(<PricingMarginPredictiveIntelligenceCard recordType="rfq" recordId="103" />);

    await waitFor(() => {
      expect(screen.getAllByText(/Insufficient Pricing Telemetry/i).length).toBeGreaterThan(0);
      expect(screen.getByText(/0 carrier quote records found/i)).toBeInTheDocument();
    });
  });

  it('triggers acknowledge action successfully', async () => {
    mockGetContractPredictedRatePressure.mockResolvedValue({ data: mockContractPrediction });
    mockAcknowledgePrediction.mockResolvedValue({ message: 'Prediction acknowledged' });

    render(<PricingMarginPredictiveIntelligenceCard recordType="contract" recordId="101" />);

    await waitFor(() => {
      expect(screen.getByText(/Acknowledge Insight/i)).toBeInTheDocument();
    });

    fireEvent.click(screen.getByText(/Acknowledge Insight/i));

    await waitFor(() => {
      expect(mockAcknowledgePrediction).toHaveBeenCalledWith('pred-ctr-101-test');
      expect(screen.getByText(/Pricing risk insight acknowledged/i)).toBeInTheDocument();
    });
  });

  it('toggles evidence and displays quantitative signals', async () => {
    mockGetRFQPredictedMargin.mockResolvedValue({ data: mockRFQPrediction });
    render(<PricingMarginPredictiveIntelligenceCard recordType="rfq" recordId="101" />);

    await waitFor(() => {
      expect(screen.getByText(/Evidence Grounding & Telemetry/i)).toBeInTheDocument();
    });

    fireEvent.click(screen.getByText(/Evidence Grounding & Telemetry/i));

    await waitFor(() => {
      expect(screen.getByText(/Quantitative Decision Drivers:/i)).toBeInTheDocument();
      expect(screen.getByText(/carrier_buy_price/i)).toBeInTheDocument();
    });
  });
});

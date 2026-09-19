import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import CustomerLeadPredictiveIntelligenceCard from '../../components/predictions/CustomerLeadPredictiveIntelligenceCard';
import * as predictionService from '../../services/predictionService';

const mockGetLeadPredictedIntelligence = vi.fn();
const mockRefreshLeadPredictedIntelligence = vi.fn();
const mockGetCustomerPredictedIntelligence = vi.fn();
const mockRefreshCustomerPredictedIntelligence = vi.fn();
const mockAcknowledgePrediction = vi.fn();
const mockRequestAction = vi.fn();

vi.mock('../../services/predictionService', () => {
  const service = {
    getLeadPredictedIntelligence: (...args) => mockGetLeadPredictedIntelligence(...args),
    refreshLeadPredictedIntelligence: (...args) => mockRefreshLeadPredictedIntelligence(...args),
    getCustomerPredictedIntelligence: (...args) => mockGetCustomerPredictedIntelligence(...args),
    refreshCustomerPredictedIntelligence: (...args) => mockRefreshCustomerPredictedIntelligence(...args),
    acknowledgePrediction: (...args) => mockAcknowledgePrediction(...args),
    requestAction: (...args) => mockRequestAction(...args),
  };
  return {
    default: service,
    predictionService: service,
    getLeadPredictedIntelligence: service.getLeadPredictedIntelligence,
    refreshLeadPredictedIntelligence: service.refreshLeadPredictedIntelligence,
    getCustomerPredictedIntelligence: service.getCustomerPredictedIntelligence,
    refreshCustomerPredictedIntelligence: service.refreshCustomerPredictedIntelligence,
    acknowledgePrediction: service.acknowledgePrediction,
    requestAction: service.requestAction,
  };
});

describe('CustomerLeadPredictiveIntelligenceCard', () => {
  const mockLeadPrediction = {
    prediction_id: 'pred-lead-101-test',
    module: 'leads',
    prediction_type: 'LEAD_CONVERSION_LIKELIHOOD',
    prediction_category: 'CONVERSION_LIKELIHOOD',
    related_record_type: 'LEAD',
    related_record_id: '101',
    severity: 'LOW',
    confidence_score: 0.95,
    confidence_band: 'HIGH',
    prediction_statement: 'High Conversion Likelihood: Lead #101 demonstrated strong commercial engagement.',
    explanation: 'Interaction history reveals rapid quotation turnaround and converted account linkage.',
    time_horizon: '14_DAYS',
    status: 'PUBLISHED',
    supporting_signals: [
      { signal_name: 'lead_status', observed_value: 'CONVERTED', baseline_value: 'NEW', importance_weight: 0.95 },
      { signal_name: 'won_quotes_count', observed_value: '1', baseline_value: '0', importance_weight: 0.90 }
    ],
    source_references: [
      { source_module: 'leads', source_record_id: '101', source_field: 'leads.status', source_timestamp: '2026-09-10T08:00:00Z' }
    ],
    source_timestamp: '2026-09-10T08:00:00Z',
    recommended_action: 'Maintain client onboarding process and review repeat RFQ capacity.',
    is_action_required: true,
    requires_approval: true,
    model_version: 'LeadCustomerIntelligence_v1.0'
  };

  const mockCustomerPrediction = {
    prediction_id: 'pred-cust-103-test',
    module: 'customers',
    prediction_type: 'CUSTOMER_CHURN_RISK',
    prediction_category: 'CHURN_RISK',
    related_record_type: 'CUSTOMER',
    related_record_id: '103',
    severity: 'HIGH',
    confidence_score: 0.88,
    confidence_band: 'HIGH',
    prediction_statement: 'Elevated Churn Risk: Customer #103 shows 45 days inactivity and warning credit status.',
    explanation: 'Decreasing shipment volume and unreviewed invoices signal attrition risk.',
    time_horizon: '60_DAYS',
    status: 'PUBLISHED',
    supporting_signals: [
      { signal_name: 'inactivity_days', observed_value: '45', baseline_value: '15', importance_weight: 0.85 }
    ],
    source_references: [
      { source_module: 'customers', source_record_id: '103', source_field: 'customers.health_score', source_timestamp: '2026-09-10T08:00:00Z' }
    ],
    source_timestamp: '2026-09-10T08:00:00Z',
    recommended_action: 'Initiate account executive commercial check-in with executive leadership.',
    is_action_required: true,
    requires_approval: true,
    model_version: 'LeadCustomerIntelligence_v1.0'
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders loading state initially', () => {
    mockGetLeadPredictedIntelligence.mockReturnValue(new Promise(() => {}));
    render(<CustomerLeadPredictiveIntelligenceCard recordType="lead" recordId="101" />);
    expect(screen.getByText(/Evaluating real-time lead conversion signals/i)).toBeInTheDocument();
  });

  it('renders lead prediction with correct badges, statement, and signals', async () => {
    mockGetLeadPredictedIntelligence.mockResolvedValue({
      success: true,
      data: mockLeadPrediction
    });

    render(<CustomerLeadPredictiveIntelligenceCard recordType="lead" recordId="101" />);

    await waitFor(() => {
      expect(screen.getByText('Predictive Lead Intelligence')).toBeInTheDocument();
      expect(screen.getByText('Conversion Likelihood')).toBeInTheDocument();
      expect(screen.getByText(/High Conversion Likelihood: Lead #101/i)).toBeInTheDocument();
      expect(screen.getByText(/Interaction history reveals rapid quotation turnaround/i)).toBeInTheDocument();
    });

    expect(screen.getByText(/lead status/i)).toBeInTheDocument();
    expect(screen.getByText('CONVERTED')).toBeInTheDocument();
  });

  it('renders customer churn risk with HITL approval badge', async () => {
    mockGetCustomerPredictedIntelligence.mockResolvedValue({
      success: true,
      data: mockCustomerPrediction
    });

    render(<CustomerLeadPredictiveIntelligenceCard recordType="customer" recordId="103" />);

    await waitFor(() => {
      expect(screen.getByText('Predictive Customer Intelligence')).toBeInTheDocument();
      expect(screen.getByText(/Elevated Churn Risk: Customer #103/i)).toBeInTheDocument();
      expect(screen.getByText(/Requires Sales\/Manager Approval/i)).toBeInTheDocument();
    });
  });

  it('toggles evidence accordion to show source records', async () => {
    mockGetLeadPredictedIntelligence.mockResolvedValue({
      success: true,
      data: mockLeadPrediction
    });

    render(<CustomerLeadPredictiveIntelligenceCard recordType="lead" recordId="101" />);

    await waitFor(() => {
      expect(screen.getByText(/Telemetry Audit/i)).toBeInTheDocument();
    });

    expect(screen.queryByText('leads.status')).not.toBeInTheDocument();

    const toggleBtn = screen.getByRole('button', { name: /Telemetry Audit/i });
    fireEvent.click(toggleBtn);

    await waitFor(() => {
      expect(screen.getByText('leads.status')).toBeInTheDocument();
      expect(screen.getByText('#101')).toBeInTheDocument();
      expect(screen.getByText(/Model: LeadCustomerIntelligence_v1.0/i)).toBeInTheDocument();
    });
  });

  it('handles acknowledge button click', async () => {
    mockGetLeadPredictedIntelligence.mockResolvedValue({
      success: true,
      data: mockLeadPrediction
    });
    mockAcknowledgePrediction.mockResolvedValue({ success: true });

    render(<CustomerLeadPredictiveIntelligenceCard recordType="lead" recordId="101" />);

    await waitFor(() => {
      expect(screen.getByText(/Acknowledge Insight/i)).toBeInTheDocument();
    });

    fireEvent.click(screen.getByText(/Acknowledge Insight/i));

    await waitFor(() => {
      expect(mockAcknowledgePrediction).toHaveBeenCalledWith('pred-lead-101-test');
      expect(screen.getByText(/Intelligence insight acknowledged/i)).toBeInTheDocument();
    });
  });

  it('handles queue mitigation action with Go Action System HITL tracking', async () => {
    mockGetCustomerPredictedIntelligence.mockResolvedValue({
      success: true,
      data: mockCustomerPrediction
    });
    mockRequestAction.mockResolvedValue({ success: true });

    render(<CustomerLeadPredictiveIntelligenceCard recordType="customer" recordId="103" />);

    await waitFor(() => {
      expect(screen.getByText(/Queue Mitigation Action/i)).toBeInTheDocument();
    });

    fireEvent.click(screen.getByText(/Queue Mitigation Action/i));

    await waitFor(() => {
      expect(mockRequestAction).toHaveBeenCalledWith(
        'pred-cust-103-test',
        expect.stringContaining('Sales representative initiated mitigation')
      );
      expect(screen.getByText(/Mitigation action queued for Human-In-The-Loop approval/i)).toBeInTheDocument();
    });
  });
});

import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { BrowserRouter } from 'react-router-dom';
import RecommendationCenterPage from '../../../pages/dashboard/Recommendations/RecommendationCenterPage';
import { recommendationService } from '../../../services/recommendationService';

// Mock recommendationService
vi.mock('../../../services/recommendationService', () => ({
  recommendationService: {
    listRecommendations: vi.fn(),
    getStats: vi.fn(),
    generateRecommendations: vi.fn(),
    markReviewed: vi.fn(),
    assignRecommendation: vi.fn(),
    dismissRecommendation: vi.fn(),
    getRecommendation: vi.fn(),
    generateDraft: vi.fn(),
    saveDraft: vi.fn(),
    createFollowupTask: vi.fn(),
    getFollowupStats: vi.fn(),
    getActionPreview: vi.fn(),
    requestApproval: vi.fn(),
  },
}));

describe('RecommendationCenterPage Component', () => {
  const mockRecommendations = [
    {
      id: 101,
      org_id: 2,
      source_type: 'SHIPMENT',
      source_id: 501,
      source_reference: 'SH-101',
      title: 'Customs Clearance Delay on SH-101',
      description: 'Shipment is delayed at port terminal awaiting customs release form.',
      category: 'operations',
      priority: 'high',
      risk_level: 'high',
      confidence: 'HIGH',
      confidence_score: 0.94,
      evidence: [
        {
          source_module: 'shipments',
          source_entity_id: 501,
          source_ref: 'SH-101',
          field_name: 'exception_type',
          observed_value: 'CUSTOMS_HOLD',
          description: 'Active customs exception recorded by terminal',
        },
      ],
      recommended_action: 'Submit revised customs bill of lading to port terminal',
      action_type: 'INVESTIGATE_EXCEPTION',
      status: 'new',
      requires_approval: false,
      created_at: '2026-09-07T10:00:00Z',
      freshness: '2026-09-07T10:05:00Z',
      rule_applied: 'UNRESOLVED_SHIPMENT_EXCEPTIONS',
      generated_by: 'DETERMINISTIC_RULES',
      correlation_id: 'corr-test-101',
    },
    {
      id: 102,
      org_id: 2,
      source_type: 'CONTRACT',
      source_id: 301,
      source_reference: 'CTR-2026-A',
      title: 'Contract CTR-2026-A Renewal Due',
      description: 'Master service agreement expires within 30 days.',
      category: 'contract',
      priority: 'critical',
      risk_level: 'critical',
      confidence: 'HIGH',
      confidence_score: 0.96,
      evidence: [],
      recommended_action: 'Initiate renegotiation and review SLA benchmarks',
      action_type: 'REVIEW_CONTRACT',
      status: 'assigned',
      assignee_name: 'Alex Legal',
      requires_approval: true,
      created_at: '2026-09-07T09:00:00Z',
      freshness: '2026-09-07T09:05:00Z',
      rule_applied: 'CONTRACT_EXPIRY_OR_RENEWAL',
      generated_by: 'DETERMINISTIC_RULES',
      correlation_id: 'corr-test-102',
    },
  ];

  const mockStats = {
    total_active: 8,
    critical_count: 2,
    high_count: 3,
    medium_count: 3,
    low_count: 0,
    requires_review: 4,
    requires_approval: 2,
    assigned_count: 2,
    completed_count: 5,
    dismissed_count: 1,
  };

  beforeEach(() => {
    vi.clearAllMocks();
    recommendationService.getStats.mockResolvedValue({
      success: true,
      stats: mockStats,
    });
    recommendationService.listRecommendations.mockResolvedValue({
      success: true,
      recommendations: mockRecommendations,
      pagination: { page: 1, limit: 10, total: 2, total_pages: 1 },
      stats: mockStats,
    });
  });

  it('renders page header, KPI metrics, and recommendation cards', async () => {
    render(
      <BrowserRouter>
        <RecommendationCenterPage />
      </BrowserRouter>
    );

    // Verify header
    expect(screen.getByText('AI Action & Recommendation Center')).toBeInTheDocument();

    // Wait for data load
    await waitFor(() => {
      expect(screen.getByText('Customs Clearance Delay on SH-101')).toBeInTheDocument();
      expect(screen.getByText('Contract CTR-2026-A Renewal Due')).toBeInTheDocument();
    });

    // Verify stats
    expect(screen.getByText('Active Signals')).toBeInTheDocument();
    expect(screen.getByText('Critical Priority')).toBeInTheDocument();

    // Verify source reference link
    expect(screen.getByText('SH-101')).toBeInTheDocument();
  });

  it('allows filtering by category and priority', async () => {
    render(
      <BrowserRouter>
        <RecommendationCenterPage />
      </BrowserRouter>
    );

    await waitFor(() => {
      expect(screen.getByText('Customs Clearance Delay on SH-101')).toBeInTheDocument();
    });

    const categorySelect = screen.getByLabelText(/Filter by Category/i);
    fireEvent.change(categorySelect, { target: { value: 'operations' } });

    await waitFor(() => {
      expect(recommendationService.listRecommendations).toHaveBeenCalledWith(
        expect.objectContaining({ category: 'operations' })
      );
    });
  });

  it('opens detail drawer and displays factual evidence', async () => {
    render(
      <BrowserRouter>
        <RecommendationCenterPage />
      </BrowserRouter>
    );

    await waitFor(() => {
      expect(screen.getByText('Customs Clearance Delay on SH-101')).toBeInTheDocument();
    });

    const viewButtons = screen.getAllByRole('button', { name: /View Details/i });
    fireEvent.click(viewButtons[0]);

    await waitFor(() => {
      expect(screen.getByText('Operational Summary')).toBeInTheDocument();
      expect(screen.getByText('Supporting Factual Evidence')).toBeInTheDocument();
      expect(screen.getByText('CUSTOMS_HOLD')).toBeInTheDocument();
      expect(screen.getByText('Rule Applied:')).toBeInTheDocument();
    });
  });

  it('handles Mark Reviewed action for new recommendation', async () => {
    recommendationService.markReviewed.mockResolvedValue({
      success: true,
      recommendation: { ...mockRecommendations[0], status: 'reviewed' },
    });

    render(
      <BrowserRouter>
        <RecommendationCenterPage />
      </BrowserRouter>
    );

    await waitFor(() => {
      expect(screen.getByText('Mark Reviewed')).toBeInTheDocument();
    });

    fireEvent.click(screen.getByText('Mark Reviewed'));

    await waitFor(() => {
      expect(recommendationService.markReviewed).toHaveBeenCalledWith(101);
    });
  });

  it('opens dismiss modal and submits with mandatory reason', async () => {
    recommendationService.dismissRecommendation.mockResolvedValue({
      success: true,
      recommendation: { ...mockRecommendations[0], status: 'dismissed' },
    });

    render(
      <BrowserRouter>
        <RecommendationCenterPage />
      </BrowserRouter>
    );

    await waitFor(() => {
      expect(screen.getByText('Customs Clearance Delay on SH-101')).toBeInTheDocument();
    });

    const dismissButtons = screen.getAllByRole('button', { name: /Dismiss/i });
    fireEvent.click(dismissButtons[0]);

    expect(screen.getByText('Dismiss Recommendation')).toBeInTheDocument();

    const textarea = screen.getByPlaceholderText(/e.g. Issue resolved externally/i);
    fireEvent.change(textarea, { target: { value: 'Customs release confirmed by agent' } });

    const confirmBtn = screen.getByRole('button', { name: /Confirm Dismissal/i });
    fireEvent.click(confirmBtn);

    await waitFor(() => {
      expect(recommendationService.dismissRecommendation).toHaveBeenCalledWith(101, 'Customs release confirmed by agent');
    });
  });

  it('renders empty state when no recommendations exist', async () => {
    recommendationService.listRecommendations.mockResolvedValue({
      success: true,
      recommendations: [],
      pagination: { page: 1, limit: 10, total: 0, total_pages: 1 },
      stats: { total_active: 0 },
    });

    render(
      <BrowserRouter>
        <RecommendationCenterPage />
      </BrowserRouter>
    );

    await waitFor(() => {
      expect(screen.getByText('No Pending Recommendations')).toBeInTheDocument();
    });
  });

  it('renders Customer Follow-Up draft and task creation buttons and handles draft generation', async () => {
    const followupRec = {
      ...mockRecommendations[0],
      id: 201,
      category: 'customer',
      customer_id: 102,
      customer_name: 'Nordic Freight Dynamics AB',
      followup_type: 'Invoice reminder',
      draft_status: 'NOT_GENERATED',
    };

    recommendationService.listRecommendations.mockResolvedValueOnce({
      success: true,
      recommendations: [followupRec],
      pagination: { page: 1, limit: 10, total: 1, total_pages: 1 },
      stats: mockStats,
    });

    recommendationService.generateDraft.mockResolvedValueOnce({
      success: true,
      recommendation: {
        ...followupRec,
        draft_status: 'DRAFTED',
        draft_subject: 'Payment Reminder for Invoice INV-2026-DEV-003',
        draft_body: 'Dear Nordic Freight Dynamics AB team, our records show an outstanding invoice.',
      },
    });

    render(
      <BrowserRouter>
        <RecommendationCenterPage />
      </BrowserRouter>
    );

    await waitFor(() => {
      expect(screen.getByRole('button', { name: /Draft Message/i })).toBeInTheDocument();
      expect(screen.getByRole('button', { name: /Create Task/i })).toBeInTheDocument();
    });

    // Explicitly click Draft Message to trigger generation
    fireEvent.click(screen.getByRole('button', { name: /Draft Message/i }));

    await waitFor(() => {
      expect(recommendationService.generateDraft).toHaveBeenCalledWith(201);
      expect(screen.getByText('Follow-Up Communication Draft')).toBeInTheDocument();
      expect(screen.getByDisplayValue(/Payment Reminder for Invoice/i)).toBeInTheDocument();
    });
  });

  it('renders Action Preview and submits to HITL approval system', async () => {
    const rfqRec = {
      id: 301,
      org_id: 2,
      source_type: 'RFQ',
      source_id: 103,
      source_reference: 'RFQ-103',
      title: 'RFQ Missing Required Port Information',
      description: 'Customer submitted RFQ lacks destination port and trade incoterm.',
      category: 'rfq',
      priority: 'high',
      risk_level: 'medium',
      confidence: 'HIGH',
      confidence_score: 0.95,
      evidence: [
        {
          source_module: 'rfq',
          source_entity_id: 103,
          source_ref: 'RFQ-103',
          field_name: 'destination',
          observed_value: 'Missing',
          description: 'Destination port undefined',
        },
      ],
      recommended_action: 'Send clarification request to customer buyer',
      action_type: 'REQUEST_RFQ_CLARIFICATION',
      status: 'new',
      requires_approval: true,
      created_at: '2026-09-07T12:00:00Z',
      freshness: '2026-09-07T12:05:00Z',
      rule_applied: 'RFQ_MISSING_INFO',
      generated_by: 'DETERMINISTIC_RULES',
      correlation_id: 'corr-rfq-301',
    };

    recommendationService.listRecommendations.mockResolvedValueOnce({
      success: true,
      recommendations: [rfqRec],
      pagination: { page: 1, limit: 10, total: 1, total_pages: 1 },
      stats: mockStats,
    });

    recommendationService.getActionPreview.mockResolvedValueOnce({
      success: true,
      action_preview: {
        recommendation_id: 301,
        source_type: 'RFQ',
        source_id: 103,
        proposed_action: 'Send clarification request to customer buyer',
        expected_effect: 'Prevents quoting invalid lanes and secures destination requirements',
        risk_level: 'MEDIUM',
        required_approval: true,
        evidence: rfqRec.evidence,
        initiated_by_user: 'Test User',
        correlation_id: 'prev-corr-301',
      },
    });

    recommendationService.requestApproval.mockResolvedValueOnce({
      success: true,
      recommendation: {
        ...rfqRec,
        approval_id: 88,
      },
    });

    render(
      <BrowserRouter>
        <RecommendationCenterPage />
      </BrowserRouter>
    );

    await waitFor(() => {
      expect(screen.getByText('RFQ Missing Required Port Information')).toBeInTheDocument();
      expect(screen.getByRole('button', { name: /Action Preview/i })).toBeInTheDocument();
    });

    // Click Action Preview
    fireEvent.click(screen.getByRole('button', { name: /Action Preview/i }));

    await waitFor(() => {
      expect(recommendationService.getActionPreview).toHaveBeenCalledWith(301);
      expect(screen.getByText('Controlled Action Preview')).toBeInTheDocument();
      expect(screen.getByText('Prevents quoting invalid lanes and secures destination requirements')).toBeInTheDocument();
      expect(screen.getByRole('button', { name: /Submit to HITL Approval System/i })).toBeInTheDocument();
    });

    // Submit Approval
    fireEvent.click(screen.getByRole('button', { name: /Submit to HITL Approval System/i }));

    await waitFor(() => {
      expect(recommendationService.requestApproval).toHaveBeenCalledWith(301, '');
    });
  });
});

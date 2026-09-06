import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import { MemoryRouter } from 'react-router-dom';
import RFQRequirements from '../../../pages/dashboard/RFQ/components/RFQRequirements';

describe('RFQRequirements Component', () => {
  const mockCompleteness = {
    percentage: 100,
    completedCount: 7,
    totalCount: 7,
    isComplete: true,
    health: 'HEALTHY',
    missingFields: [],
    fields: [],
    totalWeight: 12500,
    totalVolume: 60,
  };

  const mockRequirements = {
    operational_readiness: {
      overall_status: 'READY_FOR_QUOTATION',
      blocking_count: 0,
      missing_required_count: 0,
      conditional_attention_count: 1,
      complete_count: 12,
      total_count: 13,
      readiness_score: 92,
      next_best_action: 'All quotation-stage requirements are complete. Proceed to generate and send quotation.'
    },
    groups: [
      {
        category: 'SHIPMENT_INFO',
        title: 'Shipment Information',
        icon: '🚢',
        complete_count: 7,
        total_count: 7,
        status: 'COMPLETE',
        requirements: [
          {
            id: 'origin',
            title: 'Origin Port',
            description: 'Port of loading (POL)',
            status: 'SATISFIED',
            severity: 'BLOCKING',
            value: 'Nhava Sheva (INNSA)'
          },
          {
            id: 'destination',
            title: 'Destination Port',
            description: 'Port of discharge (POD)',
            status: 'SATISFIED',
            severity: 'BLOCKING',
            value: 'Hamburg (DEHAM)'
          }
        ]
      },
      {
        category: 'CUSTOMER_INFO',
        title: 'Customer Information',
        icon: '👤',
        complete_count: 3,
        total_count: 3,
        status: 'COMPLETE',
        requirements: [
          {
            id: 'customer_name',
            title: 'Customer Identified',
            description: 'Customer must be linked',
            status: 'SATISFIED',
            severity: 'REQUIRED',
            value: 'Global Exports Ltd',
            source_context: 'From Lead #171'
          }
        ]
      },
      {
        category: 'CONDITIONAL_COMPLIANCE',
        title: 'Conditional / Compliance Requirements',
        icon: '⚠️',
        complete_count: 1,
        total_count: 2,
        status: 'ATTENTION',
        requirements: [
          {
            id: 'container_type',
            title: 'Container Type Confirmation',
            description: 'FCL container type',
            status: 'UNDER_REVIEW',
            severity: 'CONDITIONAL',
            value: 'Pending confirmation',
            is_conditional: true,
            condition_reason: 'Incoterms FOB typically implies FCL shipment.'
          }
        ]
      }
    ],
    document_requirements: [
      {
        doc_type: 'commercial_invoice',
        title: 'Commercial Invoice',
        status: 'MISSING',
        applicable_stage: 'RFQ_STAGE',
        is_required: true,
        is_conditional: false,
        reason: 'Required before shipment processing.'
      },
      {
        doc_type: 'bill_of_lading',
        title: 'Bill of Lading (OBL)',
        status: 'NOT_APPLICABLE',
        applicable_stage: 'SHIPMENT_EXECUTION',
        is_required: false,
        is_conditional: false,
        reason: 'Generated at shipment execution stage.'
      }
    ],
    ai_findings: [
      {
        id: 'ai-extraction',
        title: 'Shipment Details Extracted',
        description: 'AI successfully extracted parameters from email thread.',
        confidence: 'HIGH',
        recommendation: 'Review extracted values.',
        requires_human_review: false,
        source_context: 'Lead #171 email thread'
      }
    ],
    lead_id: 171
  };

  const mockRfq = {
    id: 101,
    rfq_number: 'RFQ-20260827-001',
    lead_id: 171,
    customer_name: 'Global Exports Ltd'
  };

  it('renders requirements header and operational readiness badge', () => {
    render(
      <MemoryRouter>
        <RFQRequirements
          rfq={mockRfq}
          completeness={mockCompleteness}
          requirements={mockRequirements}
        />
      </MemoryRouter>
    );

    expect(screen.getByText('Requirements & Operational Readiness')).toBeInTheDocument();
    expect(screen.getByText('Ready for Quotation')).toBeInTheDocument();
    expect(screen.getByText('92%')).toBeInTheDocument();
    expect(screen.getByText('0 Blockers')).toBeInTheDocument();
  });

  it('renders Next Best Action card with action message', () => {
    render(
      <MemoryRouter>
        <RFQRequirements
          rfq={mockRfq}
          completeness={mockCompleteness}
          requirements={mockRequirements}
        />
      </MemoryRouter>
    );

    expect(screen.getByText('Next Best Action')).toBeInTheDocument();
    expect(screen.getByText(/All quotation-stage requirements are complete/i)).toBeInTheDocument();
  });

  it('renders requirement groups with titles, counts, and items', () => {
    render(
      <MemoryRouter>
        <RFQRequirements
          rfq={mockRfq}
          completeness={mockCompleteness}
          requirements={mockRequirements}
        />
      </MemoryRouter>
    );

    expect(screen.getByText('Shipment Information')).toBeInTheDocument();
    expect(screen.getByText('Origin Port')).toBeInTheDocument();
    expect(screen.getByText('Nhava Sheva (INNSA)')).toBeInTheDocument();
    expect(screen.getByText('Destination Port')).toBeInTheDocument();
    expect(screen.getByText('Hamburg (DEHAM)')).toBeInTheDocument();
    expect(screen.getByText('Customer Information')).toBeInTheDocument();
    expect(screen.getByText('Global Exports Ltd')).toBeInTheDocument();
    expect(screen.getByText('✦ From Lead #171')).toBeInTheDocument();
  });

  it('renders conditional compliance rules with condition reason', () => {
    render(
      <MemoryRouter>
        <RFQRequirements
          rfq={mockRfq}
          completeness={mockCompleteness}
          requirements={mockRequirements}
        />
      </MemoryRouter>
    );

    expect(screen.getByText('Container Type Confirmation')).toBeInTheDocument();
    expect(screen.getByText(/Incoterms FOB typically implies FCL shipment/i)).toBeInTheDocument();
  });

  it('renders stage-aware document requirements separating RFQ vs downstream stages', () => {
    render(
      <MemoryRouter>
        <RFQRequirements
          rfq={mockRfq}
          completeness={mockCompleteness}
          requirements={mockRequirements}
        />
      </MemoryRouter>
    );

    expect(screen.getByText('Lifecycle Document Requirements (Stage-Aware)')).toBeInTheDocument();
    expect(screen.getByText('🎯 RFQ & Quotation Stage')).toBeInTheDocument();
    expect(screen.getByText('🚢 Downstream Operational Documents')).toBeInTheDocument();
    expect(screen.getByText('📋 Commercial Invoice')).toBeInTheDocument();
    expect(screen.getByText('📑 Bill of Lading (OBL)')).toBeInTheDocument();
    expect(screen.getByText('Not Applicable Now')).toBeInTheDocument();
  });

  it('renders AI Operational Findings with confidence pill and source context', () => {
    render(
      <MemoryRouter>
        <RFQRequirements
          rfq={mockRfq}
          completeness={mockCompleteness}
          requirements={mockRequirements}
        />
      </MemoryRouter>
    );

    expect(screen.getByText('AI Intelligence & Risk Analysis Findings')).toBeInTheDocument();
    expect(screen.getByText('HIGH Confidence')).toBeInTheDocument();
    expect(screen.getByText(/Recommendation: Review extracted values/i)).toBeInTheDocument();
  });

  it('calls onSwitchTab with quotes when Proceed button is clicked', () => {
    const handleSwitchTab = vi.fn();
    render(
      <MemoryRouter>
        <RFQRequirements
          rfq={mockRfq}
          completeness={mockCompleteness}
          requirements={mockRequirements}
          onSwitchTab={handleSwitchTab}
        />
      </MemoryRouter>
    );

    const proceedBtn = screen.getByText('⚡ Proceed to Quotes →');
    fireEvent.click(proceedBtn);
    expect(handleSwitchTab).toHaveBeenCalledWith('quotes');
  });
});

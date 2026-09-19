import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import {
  AIInsightCard,
  AIStatusBadge,
  AIConfidenceIndicator,
  AIEvidenceList,
  AIRecommendationCard,
  AISectionHeader,
  AIEmptyState,
  AIErrorState,
  AILoadingState,
  AISourceReference,
  AIWorkforceSummary,
} from '../../components/ai';

describe('Shared AI Design System Components (Light Enterprise Theme)', () => {
  it('renders AIStatusBadge with proper label and light styling', () => {
    const { rerender } = render(<AIStatusBadge status="completed" />);
    expect(screen.getByText('completed')).toBeInTheDocument();

    rerender(<AIStatusBadge status="critical" label="CRITICAL COVENANT" />);
    expect(screen.getByText('CRITICAL COVENANT')).toBeInTheDocument();
  });

  it('renders AIConfidenceIndicator with HIGH, MEDIUM, and LOW confidence', () => {
    const { rerender } = render(<AIConfidenceIndicator confidence="HIGH" />);
    expect(screen.getByText(/Confidence:/i)).toBeInTheDocument();
    expect(screen.getByText('HIGH')).toBeInTheDocument();

    rerender(<AIConfidenceIndicator confidence="MEDIUM" />);
    expect(screen.getByText('MEDIUM')).toBeInTheDocument();
  });

  it('renders AIEvidenceList and toggles collapsible view', () => {
    const mockEvidence = [
      {
        source_module: 'shipments',
        source_ref: 'SH-101',
        field_name: 'status',
        observed_value: 'IN_TRANSIT',
        description: 'Vessel in voyage',
      },
      {
        source_module: 'finance',
        source_ref: 'INV-2026-001',
        field_name: 'balance_due',
        observed_value: 24500,
        description: 'Overdue 14 days',
      },
    ];

    render(<AIEvidenceList evidence={mockEvidence} defaultExpanded={true} />);
    expect(screen.getByText(/Verifiable Database Facts \(2\)/i)).toBeInTheDocument();
    expect(screen.getByText(/SH-101/i)).toBeInTheDocument();
    expect(screen.getByText('24500')).toBeInTheDocument();

    // Toggle collapse
    const header = screen.getByRole('button');
    fireEvent.click(header);
    expect(screen.queryByText('24500')).not.toBeInTheDocument();
  });

  it('renders AIInsightCard with severity, explanation, and read-only notice', () => {
    const mockEvidence = [
      { source_module: 'finance', source_ref: 'INV-01', field_name: 'balance', observed_value: '1000' },
    ];

    render(
      <AIInsightCard
        title="Delayed Freight with Overdue Invoice"
        explanation="Shipment SH-101 is stalled while customer has an outstanding balance of $24,500."
        severity="CRITICAL"
        confidence="HIGH"
        category="Operations & Finance"
        evidence={mockEvidence}
        suggestedAction="Request credit release before cargo discharge."
        ruleApplied="RULE_OPS_FIN_DELAYED_OVERDUE"
        readOnly={true}
        freshness="2026-09-07T15:00:00Z"
        correlationId="corr-test-12345"
      />
    );

    expect(screen.getByText('Delayed Freight with Overdue Invoice')).toBeInTheDocument();
    expect(screen.getByText(/Shipment SH-101 is stalled/i)).toBeInTheDocument();
    expect(screen.getByText('CRITICAL')).toBeInTheDocument();
    expect(screen.getByText('READ-ONLY')).toBeInTheDocument();
    expect(screen.getByText(/Request credit release before cargo discharge/i)).toBeInTheDocument();
  });

  it('renders AIRecommendationCard with questions and disclaimer', () => {
    const questions = [
      'Has the consignee made a partial payment wire?',
      'Is the carrier holding documents at transshipment port?',
    ];

    render(
      <AIRecommendationCard
        title="Operational Cross-Module Assessment"
        executiveSummary="Cross-module friction detected between billing and shipment delivery."
        tradeoffs="Releasing cargo prevents port demurrage, but increases bad debt risk."
        questions={questions}
        confidence="HIGH"
      />
    );

    expect(screen.getByText('Operational Cross-Module Assessment')).toBeInTheDocument();
    expect(screen.getByText(/Cross-module friction detected/i)).toBeInTheDocument();
    expect(screen.getByText(/Releasing cargo prevents port demurrage/i)).toBeInTheDocument();
    expect(screen.getByText(/Has the consignee made a partial payment wire/i)).toBeInTheDocument();
  });

  it('renders AISectionHeader and triggers refresh', () => {
    const handleRefresh = vi.fn();
    render(
      <AISectionHeader
        title="Cross-Module Connected Intelligence"
        subtitle="Connected situations across operations and finance."
        badgeText="2 Active Signals"
        badgeType="warning"
        correlationId="corr-test-999"
        freshness="2026-09-07T15:00:00Z"
        onRefresh={handleRefresh}
      />
    );

    expect(screen.getByText('Cross-Module Connected Intelligence')).toBeInTheDocument();
    expect(screen.getByText('2 Active Signals')).toBeInTheDocument();

    const refreshBtn = screen.getByRole('button', { name: /refresh intelligence/i });
    fireEvent.click(refreshBtn);
    expect(handleRefresh).toHaveBeenCalledTimes(1);
  });

  it('renders AIEmptyState and AIErrorState cleanly', () => {
    const { rerender } = render(
      <AIEmptyState
        title="No Active Operational Risks"
        description="All shipments and billing invoices are in compliance."
      />
    );
    expect(screen.getByText('No Active Operational Risks')).toBeInTheDocument();

    const handleRetry = vi.fn();
    rerender(
      <AIErrorState
        title="Sidecar Offline"
        message="AI intelligence runtime could not be reached."
        onRetry={handleRetry}
      />
    );
    expect(screen.getByText('Sidecar Offline')).toBeInTheDocument();
    const retryBtn = screen.getByRole('button', { name: /retry/i });
    fireEvent.click(retryBtn);
    expect(handleRetry).toHaveBeenCalledTimes(1);
  });

  it('renders AILoadingState and AISourceReference correctly', () => {
    const { rerender } = render(<AILoadingState message="Auditing carrier contracts..." />);
    expect(screen.getByText('Auditing carrier contracts...')).toBeInTheDocument();

    const handleClick = vi.fn();
    rerender(
      <AISourceReference
        module="shipments"
        entityId={101}
        reference="SH-2026-101"
        onClick={handleClick}
      />
    );
    const refBtn = screen.getByRole('button');
    expect(refBtn).toHaveTextContent('SH-2026-101');
    fireEvent.click(refBtn);
    expect(handleClick).toHaveBeenCalledTimes(1);
  });

  it('renders AIWorkforceSummary metrics in light cards', () => {
    const mockSummary = {
      queued: 2,
      processing: 1,
      waiting_for_approval: 1,
      completed: 15,
      failed: 0,
    };

    render(<AIWorkforceSummary summary={mockSummary} />);
    expect(screen.getByText('AI Workforce Execution Summary')).toBeInTheDocument();
    expect(screen.getByText('Queued')).toBeInTheDocument();
    expect(screen.getByText('2')).toBeInTheDocument();
    expect(screen.getByText('Processing')).toBeInTheDocument();
    expect(screen.getByText('15')).toBeInTheDocument();
  });
});

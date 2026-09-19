import React from 'react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import ShipmentDetail from '../../../pages/dashboard/Shipments/ShipmentDetail';
import api from '../../../services/api';

vi.mock('../../../services/api', () => ({
  default: {
    get: vi.fn(),
    post: vi.fn(),
    put: vi.fn(),
    delete: vi.fn(),
  },
}));

vi.mock('../../../components/common/BusinessIntelligenceCard', () => ({
  default: () => <div data-testid="mock-bi-card">Business Intelligence Card</div>,
}));

vi.mock('../../../components/recommendations/ModuleRecommendationsWidget', () => ({
  default: () => <div data-testid="mock-module-recommendations">Recommendations Widget</div>,
}));

vi.mock('../../../pages/dashboard/Shipments/components/ShipmentOperationsIntelligenceSection', () => ({
  default: () => <div data-testid="mock-ops-copilot">Shipment Operations Intelligence Section</div>,
}));

vi.mock('../../../pages/dashboard/Shipments/DocumentWorkspace', () => ({
  default: () => <div data-testid="mock-doc-workspace">Documents & Compliance Workspace</div>,
}));

vi.mock('../../../pages/dashboard/Shipments/FinanceWorkspace', () => ({
  default: () => <div data-testid="mock-fin-workspace">Financials Workspace</div>,
}));

vi.mock('../../../pages/dashboard/Shipments/BillingWorkspace', () => ({
  default: () => <div data-testid="mock-billing-workspace">Billing Workspace</div>,
}));

const mockShipmentData = {
  shipment: {
    id: 101,
    booking_id: 201,
    booking_number: 'BKG-MAEU-99001',
    carrier_scac: 'MAEU',
    origin_port: 'CNSHA',
    destination_port: 'USLAX',
    vessel_name: 'Maersk Mc-Kinney',
    voyage_number: '2608E',
    status: 'IN_TRANSIT',
    planned_etd: '2026-09-10T00:00:00Z',
    actual_departure: '2026-09-10T10:00:00Z',
    planned_eta: '2026-09-24T00:00:00Z',
    actual_arrival: null,
    closure_status: 'ACTIVE',
    container_numbers: 'MSKU1234567,MSKU7654321',
  },
  milestones: [
    {
      id: 1,
      milestone_code: 'GATE_IN_ORIGIN',
      status: 'COMPLETED',
      planned_date: '2026-09-08T00:00:00Z',
      actual_date: '2026-09-08T06:00:00Z',
      location: 'CNSHA',
    },
    {
      id: 2,
      milestone_code: 'VESSEL_DEPARTURE',
      status: 'COMPLETED',
      planned_date: '2026-09-10T00:00:00Z',
      actual_date: '2026-09-10T10:00:00Z',
      location: 'CNSHA',
    },
    {
      id: 3,
      milestone_code: 'VESSEL_ARRIVAL',
      status: 'PENDING',
      planned_date: '2026-09-24T00:00:00Z',
      actual_date: null,
      location: 'USLAX',
    },
  ],
  exceptions: [
    {
      id: 88,
      severity: 'HIGH',
      exception_type: 'SCHEDULE_DELAY',
      title: 'Transshipment Port Congestion Delay',
      description: 'Carrier reported a 36-hour delay due to berth congestion.',
      status: 'DETECTED',
      created_at: '2026-09-08T01:00:00Z',
    },
  ],
};

const mockTrackingSummary = {
  journey_progress_pct: 45,
  tracking_state: 'IN_TRANSIT',
  schedule_variance: 0.5,
  planned_etd: '2026-09-10T00:00:00Z',
  actual_etd: '2026-09-10T10:00:00Z',
  planned_eta: '2026-09-24T00:00:00Z',
  actual_arrival: null,
};

describe('ShipmentDetail Workspace Tabs', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    api.get.mockImplementation((url) => {
      if (url.includes('/tracking')) {
        return Promise.resolve(mockTrackingSummary);
      }
      return Promise.resolve(mockShipmentData);
    });
  });

  const renderComponent = (initialEntries = ['/dashboard/shipments/101']) => {
    return render(
      <MemoryRouter initialEntries={initialEntries}>
        <Routes>
          <Route path="/dashboard/shipments/:id" element={<ShipmentDetail />} />
        </Routes>
      </MemoryRouter>
    );
  };

  it('renders all 6 workspace navigation tabs and shows Overview by default', async () => {
    renderComponent();

    await waitFor(() => {
      expect(screen.getByText('Transit Journey')).toBeInTheDocument();
    });

    // Check all tab buttons exist
    expect(screen.getByRole('tab', { name: /overview/i })).toBeInTheDocument();
    expect(screen.getByRole('tab', { name: /tracking & milestones/i })).toBeInTheDocument();
    expect(screen.getByRole('tab', { name: /exceptions & diagnostics/i })).toBeInTheDocument();
    expect(screen.getByRole('tab', { name: /documents & compliance/i })).toBeInTheDocument();
    expect(screen.getByRole('tab', { name: /financials & billing/i })).toBeInTheDocument();
    expect(screen.getByRole('tab', { name: /operations copilot & ai/i })).toBeInTheDocument();

    // Default active tab is Overview
    expect(screen.getByText('Operational Milestones Snapshot')).toBeInTheDocument();
    expect(screen.getByText('Cargo & Equipment Manifest')).toBeInTheDocument();
  });

  it('switches to Milestones tab when clicked', async () => {
    renderComponent();

    await waitFor(() => {
      expect(screen.getByText('Transit Journey')).toBeInTheDocument();
    });

    const milestonesTab = screen.getByRole('tab', { name: /tracking & milestones/i });
    fireEvent.click(milestonesTab);

    await waitFor(() => {
      expect(screen.getByText('GATE IN ORIGIN')).toBeInTheDocument();
      expect(screen.getByText('VESSEL DEPARTURE')).toBeInTheDocument();
    });
  });

  it('switches to Exceptions tab when clicked', async () => {
    renderComponent();

    await waitFor(() => {
      expect(screen.getByText('Transit Journey')).toBeInTheDocument();
    });

    const exceptionsTab = screen.getByRole('tab', { name: /exceptions & diagnostics/i });
    fireEvent.click(exceptionsTab);

    await waitFor(() => {
      expect(screen.getByText('Anomalies & Exceptions')).toBeInTheDocument();
      expect(screen.getByText('Transshipment Port Congestion Delay')).toBeInTheDocument();
    });
  });

  it('switches to Documents tab and renders DocumentWorkspace', async () => {
    renderComponent();

    await waitFor(() => {
      expect(screen.getByText('Transit Journey')).toBeInTheDocument();
    });

    const docsTab = screen.getByRole('tab', { name: /documents & compliance/i });
    fireEvent.click(docsTab);

    await waitFor(() => {
      expect(screen.getByTestId('mock-doc-workspace')).toBeInTheDocument();
    });
  });

  it('switches to Financials tab and renders Finance & Billing workspaces', async () => {
    renderComponent();

    await waitFor(() => {
      expect(screen.getByText('Transit Journey')).toBeInTheDocument();
    });

    const financeTab = screen.getByRole('tab', { name: /financials & billing/i });
    fireEvent.click(financeTab);

    await waitFor(() => {
      expect(screen.getByTestId('mock-fin-workspace')).toBeInTheDocument();
      expect(screen.getByTestId('mock-billing-workspace')).toBeInTheDocument();
    });
  });

  it('switches to Operations Copilot & AI tab', async () => {
    renderComponent();

    await waitFor(() => {
      expect(screen.getByText('Transit Journey')).toBeInTheDocument();
    });

    const aiTab = screen.getByRole('tab', { name: /operations copilot & ai/i });
    fireEvent.click(aiTab);

    await waitFor(() => {
      expect(screen.getByTestId('mock-module-recommendations')).toBeInTheDocument();
      expect(screen.getByTestId('mock-ops-copilot')).toBeInTheDocument();
      expect(screen.getByTestId('mock-bi-card')).toBeInTheDocument();
    });
  });

  it('respects URL query param tab selection (e.g. ?tab=exceptions)', async () => {
    renderComponent(['/dashboard/shipments/101?tab=exceptions']);

    await waitFor(() => {
      expect(screen.getByText('Anomalies & Exceptions')).toBeInTheDocument();
      expect(screen.getByText('Transshipment Port Congestion Delay')).toBeInTheDocument();
    });
  });
});

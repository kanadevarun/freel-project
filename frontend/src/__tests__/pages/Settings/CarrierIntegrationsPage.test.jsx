import React from 'react';
import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { MemoryRouter } from 'react-router-dom';
import CarrierIntegrationsPage from '../../../pages/dashboard/Settings/CarrierIntegrationsPage';
import { carrierIntegrationsService } from '../../../services/carrierIntegrationsService';

vi.mock('../../../services/carrierIntegrationsService', () => ({
  carrierIntegrationsService: {
    getIntegrations: vi.fn(),
    getProviders: vi.fn(),
    connectCarrier: vi.fn(),
    testDirectConnection: vi.fn(),
    testConnection: vi.fn(),
    syncCarrier: vi.fn(),
    toggleCarrier: vi.fn(),
    disconnectCarrier: vi.fn(),
    getSyncHistory: vi.fn(),
    getIntegrationHealth: vi.fn(),
  },
}));

describe('CarrierIntegrationsPage Component', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders clean zero-state when tenant has no connected carriers', async () => {
    carrierIntegrationsService.getIntegrations.mockResolvedValue([]);
    carrierIntegrationsService.getProviders.mockResolvedValue([
      { code: 'MAERSK', name: 'A.P. Moller – Maersk', scac: 'MAEU' },
      { code: 'MSC', name: 'Mediterranean Shipping Company', scac: 'MSCU' },
    ]);

    render(
      <MemoryRouter>
        <CarrierIntegrationsPage />
      </MemoryRouter>
    );

    await waitFor(() => {
      expect(screen.getByText('Carrier Integrations')).toBeInTheDocument();
    });

    // Check KPI summary cards
    expect(screen.getAllByText('Connected Carriers')[0]).toBeInTheDocument();
    expect(screen.getByText('Active Connections')).toBeInTheDocument();
    expect(screen.getAllByText('Needs Attention')[0]).toBeInTheDocument();
    expect(screen.getByText('Last Sync')).toBeInTheDocument();

    // Check empty state
    expect(screen.getByText('No Carrier Connections Yet')).toBeInTheDocument();
    expect(screen.getByText(/Connect Maersk, MSC, Hapag-Lloyd/i)).toBeInTheDocument();
    expect(screen.getAllByRole('button', { name: /\+ Connect Carrier/i })[0]).toBeInTheDocument();
  });

  it('renders connected carriers list with masked credentials and status', async () => {
    const mockIntegrations = [
      {
        id: 101,
        carrier_scac: 'MAEU',
        carrier_name: 'A.P. Moller – Maersk',
        carrier_code: 'MAERSK',
        environment: 'SANDBOX',
        connection_status: 'CONNECTED',
        connection_method: 'API',
        health_state: 'HEALTHY',
        is_active: true,
        capabilities: ['TRACKING', 'RATES'],
        supported_capabilities: ['TRACKING', 'RATES', 'BOOKING'],
        credentials_mask: { api_key: '••••••••8899' },
        has_credentials: true,
        last_synced_at: new Date().toISOString(),
      },
    ];

    carrierIntegrationsService.getIntegrations.mockResolvedValue(mockIntegrations);
    carrierIntegrationsService.getProviders.mockResolvedValue([
      { code: 'MAERSK', name: 'A.P. Moller – Maersk', scac: 'MAEU' },
    ]);

    render(
      <MemoryRouter>
        <CarrierIntegrationsPage />
      </MemoryRouter>
    );

    await waitFor(() => {
      expect(screen.getByText('A.P. Moller – Maersk')).toBeInTheDocument();
    });

    expect(screen.getByText('SCAC: MAEU')).toBeInTheDocument();
    expect(screen.getByText('SANDBOX')).toBeInTheDocument();
    expect(screen.getByText('Healthy')).toBeInTheDocument();
  });

  it('opens Connect Carrier modal on button click', async () => {
    carrierIntegrationsService.getIntegrations.mockResolvedValue([]);
    carrierIntegrationsService.getProviders.mockResolvedValue([
      { code: 'MAERSK', name: 'A.P. Moller – Maersk', scac: 'MAEU' },
    ]);

    render(
      <MemoryRouter>
        <CarrierIntegrationsPage />
      </MemoryRouter>
    );

    await waitFor(() => {
      expect(screen.getByText('No Carrier Connections Yet')).toBeInTheDocument();
    });

    const connectBtn = screen.getAllByRole('button', { name: /\+ Connect Carrier/i })[0];
    fireEvent.click(connectBtn);

    expect(screen.getByText('Connect Carrier Integration')).toBeInTheDocument();
  });
});

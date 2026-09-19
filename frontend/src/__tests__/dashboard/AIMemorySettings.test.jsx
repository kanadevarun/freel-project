import React from 'react';
import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { MemoryRouter } from 'react-router-dom';
import AIMemorySettingsPage from '../../pages/dashboard/Settings/AIMemorySettingsPage';
import memoryService from '../../services/memoryService';

vi.mock('../../services/memoryService', () => ({
  default: {
    getStats: vi.fn(),
    listMemories: vi.fn(),
    getUserSettings: vi.fn(),
    updateUserSettings: vi.fn(),
    togglePersonalization: vi.fn(),
    createMemory: vi.fn(),
    updateMemory: vi.fn(),
    deleteMemory: vi.fn(),
    disableMemory: vi.fn(),
    enableMemory: vi.fn(),
    clearPersonalMemories: vi.fn(),
    listAuditEvents: vi.fn(),
  },
}));

vi.mock('../../context/RBACContext', () => ({
  useRBAC: () => ({
    userRole: 'SUPER_ADMIN',
    can: () => true,
  }),
}));

const mockStats = {
  personalization_enabled: true,
  active_personal_count: 3,
  active_org_count: 2,
  pending_review_count: 0,
  expired_count: 0,
  total_audit_events: 12,
};

const mockMemories = [
  {
    id: 1,
    org_id: 1,
    user_id: 10,
    scope: 'USER',
    memory_type: 'RESPONSE_STYLE',
    title: 'Concise Ocean Summaries',
    content: 'Present container exceptions in bullet points under 3 lines.',
    status: 'ACTIVE',
    explicitly_confirmed: true,
    updated_at: '2026-09-08T06:00:00Z',
  },
  {
    id: 2,
    org_id: 1,
    user_id: 10,
    scope: 'USER',
    memory_type: 'TERMINOLOGY',
    title: 'Preferred Customer Label',
    content: 'Use shipper instead of client.',
    status: 'DISABLED',
    explicitly_confirmed: true,
    updated_at: '2026-09-07T06:00:00Z',
  },
];

const mockSettings = {
  personalization_enabled: true,
  preferred_response_style: 'CONCISE',
  preferred_summary_depth: 'STANDARD',
  preferred_currency: 'USD',
  preferred_timezone: 'UTC',
  preferred_date_format: 'YYYY-MM-DD',
  preferred_default_module: 'SHIPMENTS',
  explanation_level: 'DIRECT',
};

describe('AIMemorySettingsPage Component Tests', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    memoryService.getStats.mockResolvedValue({ data: mockStats });
    memoryService.listMemories.mockResolvedValue({ data: mockMemories });
    memoryService.getUserSettings.mockResolvedValue({ data: mockSettings });
    memoryService.listAuditEvents.mockResolvedValue({ data: [] });
  });

  it('renders memory KPIs and tabs correctly', async () => {
    render(
      <MemoryRouter>
        <AIMemorySettingsPage />
      </MemoryRouter>
    );

    expect(screen.getByText(/AI Memory & Personalization/i)).toBeInTheDocument();
    await waitFor(() => {
      expect(screen.getByText('Concise Ocean Summaries')).toBeInTheDocument();
    });

    expect(screen.getByText('Preferred Customer Label')).toBeInTheDocument();
    expect(screen.getAllByText('Personal Memories').length).toBeGreaterThanOrEqual(1);
    expect(screen.getByText('Organization Terminology & Facts')).toBeInTheDocument();
  });

  it('allows switching to AI Preferences & Formats tab', async () => {
    memoryService.getStats.mockResolvedValueOnce({ data: mockStats });
    memoryService.listMemories.mockResolvedValueOnce({ data: mockMemories });
    memoryService.getUserSettings.mockResolvedValueOnce({ data: mockSettings });
    memoryService.listAuditEvents.mockResolvedValueOnce({ data: [] });

    render(
      <MemoryRouter>
        <AIMemorySettingsPage />
      </MemoryRouter>
    );

    await waitFor(() => {
      expect(screen.getByText('Concise Ocean Summaries')).toBeInTheDocument();
    });

    const prefTab = screen.getByText('AI Preferences & Formats');
    fireEvent.click(prefTab);

    expect(screen.getByText(/Preferred Response Style/i)).toBeInTheDocument();
    expect(screen.getByText(/Preferred Summary Depth/i)).toBeInTheDocument();
    expect(screen.getByText(/Preferred Currency Display/i)).toBeInTheDocument();
  });
});

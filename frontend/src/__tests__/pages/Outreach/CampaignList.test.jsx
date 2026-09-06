import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import CampaignList from '../../../pages/dashboard/Outreach/CampaignList';
import * as outreachService from '../../../services/outreachService';
import { CAMPAIGN_STATUS } from '../../../pages/dashboard/Outreach/constants';

// Mock the API calls
vi.mock('../../../services/outreachService', () => ({
  activateCampaign: vi.fn(),
  pauseCampaign: vi.fn(),
  deleteCampaign: vi.fn(),
}));

describe('CampaignList Component', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    // Stub window.confirm to always return true (simulating user clicking OK)
    vi.stubGlobal('confirm', () => true);
    vi.stubGlobal('alert', () => {});
  });

  const mockCampaigns = [
    { id: 1, name: 'Summer Push', status: CAMPAIGN_STATUS.DRAFT, created_at: '2026-06-01T10:00:00Z' },
    { id: 2, name: 'Q4 Renewals', status: CAMPAIGN_STATUS.ACTIVE, created_at: '2026-06-05T10:00:00Z' }
  ];

  it('renders loading skeleton when loading is true', () => {
    const { container } = render(
      <CampaignList campaigns={[]} loading={true} onCampaignsChanged={vi.fn()} />
    );
    expect(container.querySelectorAll('.outreach-skeleton-row').length).toBeGreaterThan(0);
  });

  it('renders empty state when there are no campaigns', () => {
    render(
      <CampaignList campaigns={[]} loading={false} onCampaignsChanged={vi.fn()} />
    );
    expect(screen.getByText('No campaigns yet')).toBeInTheDocument();
  });

  it('renders campaigns with correct buttons based on status', () => {
    const { container } = render(
      <CampaignList campaigns={mockCampaigns} loading={false} onCampaignsChanged={vi.fn()} />
    );

    expect(screen.getByText('Summer Push')).toBeInTheDocument();
    expect(screen.getByText('Q4 Renewals')).toBeInTheDocument();

    // DRAFT campaign should have a Launch button
    const launchBtns = container.querySelectorAll('.campaign-action-pill-btn.launch');
    expect(launchBtns).toHaveLength(1);

    // ACTIVE campaign should have a Pause button
    const pauseBtns = container.querySelectorAll('.campaign-action-pill-btn.pause');
    expect(pauseBtns).toHaveLength(1);
  });

  it('calls activateCampaign when Launch is clicked', async () => {
    const onCampaignsChanged = vi.fn();
    outreachService.activateCampaign.mockResolvedValueOnce({});

    const { container } = render(
      <CampaignList campaigns={mockCampaigns} loading={false} onCampaignsChanged={onCampaignsChanged} />
    );

    const launchBtn = container.querySelector('.campaign-action-pill-btn.launch');
    fireEvent.click(launchBtn);

    expect(outreachService.activateCampaign).toHaveBeenCalledWith(1); // Summer Push is ID 1
    
    // Wait for the async API call and refresh callback to fire
    await waitFor(() => {
      expect(onCampaignsChanged).toHaveBeenCalled();
    });
  });

  it('calls pauseCampaign when Pause is clicked', async () => {
    const onCampaignsChanged = vi.fn();
    outreachService.pauseCampaign.mockResolvedValueOnce({});

    const { container } = render(
      <CampaignList campaigns={mockCampaigns} loading={false} onCampaignsChanged={onCampaignsChanged} />
    );

    const pauseBtn = container.querySelector('.campaign-action-pill-btn.pause');
    fireEvent.click(pauseBtn);

    expect(outreachService.pauseCampaign).toHaveBeenCalledWith(2); // Q4 Renewals is ID 2
    
    await waitFor(() => {
      expect(onCampaignsChanged).toHaveBeenCalled();
    });
  });
});

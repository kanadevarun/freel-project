import { describe, it, expect, vi, beforeEach } from 'vitest';
import * as outreachService from '../../services/outreachService';

// Mock the global fetch API
global.fetch = vi.fn();

describe('Outreach Service', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    
    // Simulate setting local storage so api.js has a token
    localStorage.setItem('auth_tokens', JSON.stringify({
      accessToken: 'fake-jwt-token'
    }));
  });

  it('listCampaigns should fetch campaigns', async () => {
    fetch.mockResolvedValueOnce({
      ok: true,
      json: async () => ({ success: true, data: { campaigns: [], total: 0 } })
    });

    const result = await outreachService.listCampaigns({ limit: 10, offset: 0 });
    
    expect(fetch).toHaveBeenCalledTimes(1);
    expect(fetch.mock.calls[0][0]).toContain('/api/v1/outreach/campaigns?limit=10&offset=0');
    expect(fetch.mock.calls[0][1].method).toBe('GET');
    expect(result).toEqual({ campaigns: [], total: 0 });
  });

  it('createCampaign should send POST request', async () => {
    fetch.mockResolvedValueOnce({
      ok: true,
      json: async () => ({ success: true, data: { id: 1, name: 'Test' } })
    });

    const payload = { name: 'Test' };
    const result = await outreachService.createCampaign(payload);
    
    expect(fetch).toHaveBeenCalledTimes(1);
    expect(fetch.mock.calls[0][0]).toContain('/api/v1/outreach/campaigns');
    expect(fetch.mock.calls[0][1].method).toBe('POST');
    expect(fetch.mock.calls[0][1].body).toBe(JSON.stringify(payload));
    expect(result).toEqual({ id: 1, name: 'Test' });
  });

  it('activateCampaign should POST to activate endpoint', async () => {
    fetch.mockResolvedValueOnce({
      ok: true,
      json: async () => ({ success: true, data: { id: 1, status: 'ACTIVE' } })
    });

    const result = await outreachService.activateCampaign(1);
    
    expect(fetch).toHaveBeenCalledTimes(1);
    expect(fetch.mock.calls[0][0]).toContain('/api/v1/outreach/campaigns/1/activate');
    expect(fetch.mock.calls[0][1].method).toBe('POST');
    expect(result.status).toBe('ACTIVE');
  });

  it('generateEmail should call AI endpoint', async () => {
    fetch.mockResolvedValueOnce({
      ok: true,
      json: async () => ({ success: true, data: { subject: 'Hello', body: 'World' } })
    });

    const payload = { company_name: 'Acme', industry: 'Tech', goal: 'Intro' };
    const result = await outreachService.generateEmail(payload);
    
    expect(fetch).toHaveBeenCalledTimes(1);
    expect(fetch.mock.calls[0][0]).toContain('/api/v1/outreach/generate-email');
    expect(fetch.mock.calls[0][1].method).toBe('POST');
    expect(JSON.parse(fetch.mock.calls[0][1].body)).toEqual(payload);
    expect(result).toEqual({ subject: 'Hello', body: 'World' });
  });
});

import { describe, it, expect, vi, beforeEach } from 'vitest';
import { rfqService } from '../../services/rfqService';
import api from '../../services/api';

vi.mock('../../services/api');

describe('rfqService', () => {
  beforeEach(() => {
    vi.resetAllMocks();
  });

  it('listRFQs calls api.get with correct params', async () => {
    api.get.mockResolvedValue({ data: { rfqs: [], total: 0 } });
    await rfqService.listRFQs(10, 20);
    expect(api.get).toHaveBeenCalledWith('/api/v1/rfqs?limit=10&offset=20');
  });

  it('parseShipmentRequest calls api.post with raw_text', async () => {
    const mockResponse = { data: { origin: 'Miami', confidence_score: 95 } };
    api.post.mockResolvedValue(mockResponse);
    
    const result = await rfqService.parseShipmentRequest('Please quote Miami to London');
    
    expect(api.post).toHaveBeenCalledWith('/api/v1/rfqs/parse-shipment-request', {
      raw_text: 'Please quote Miami to London'
    });
    expect(result).toEqual(mockResponse);
  });

  it('getRequirements calls api.get with correct endpoint', async () => {
    const mockRequirements = {
      data: {
        operational_readiness: { overall_status: 'READY_FOR_QUOTATION', blocking_count: 0 },
        groups: [],
        document_requirements: [],
        ai_findings: []
      }
    };
    api.get.mockResolvedValue(mockRequirements);

    const result = await rfqService.getRequirements(101);

    expect(api.get).toHaveBeenCalledWith('/api/v1/rfqs/101/requirements');
    expect(result).toEqual(mockRequirements);
  });

  it('getTimeline calls api.get with correct endpoint', async () => {
    const mockTimeline = { data: [{ id: 'evt-1', action: 'CREATED' }] };
    api.get.mockResolvedValue(mockTimeline);

    const result = await rfqService.getTimeline(101);

    expect(api.get).toHaveBeenCalledWith('/api/v1/rfqs/101/timeline');
    expect(result).toEqual(mockTimeline);
  });

  it('getActivity calls api.get with correct endpoint', async () => {
    const mockActivity = {
      data: {
        summary: { total_events: 5, customer_events: 2 },
        events: []
      }
    };
    api.get.mockResolvedValue(mockActivity);

    const result = await rfqService.getActivity(101);

    expect(api.get).toHaveBeenCalledWith('/api/v1/rfqs/101/activity');
    expect(result).toEqual(mockActivity);
  });
});



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
});

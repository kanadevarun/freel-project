import { describe, it, expect, vi, beforeEach } from 'vitest';
import auditService from '../../services/auditService';
import api from '../../services/api';

vi.mock('../../services/api', () => ({
  default: {
    get: vi.fn(),
  },
}));

describe('auditService', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('getAuditLogs builds query parameters and calls api.get', async () => {
    const mockData = { items: [], total: 0, page: 1, limit: 20, total_pages: 1 };
    api.get.mockResolvedValueOnce({ data: mockData });

    const result = await auditService.getAuditLogs({
      page: 2,
      limit: 10,
      module: 'SHIPMENTS',
      action: 'UPDATE',
      search: 'SHP-1024',
    });

    expect(api.get).toHaveBeenCalledWith('/api/v1/audit-logs?page=2&limit=10&module=SHIPMENTS&action=UPDATE&search=SHP-1024');
    expect(result).toEqual(mockData);
  });

  it('getAuditLogById calls api.get with ID parameter', async () => {
    const mockLog = { id: 42, module: 'CARRIER_INTEGRATIONS', action: 'CONNECT' };
    api.get.mockResolvedValueOnce({ data: mockLog });

    const result = await auditService.getAuditLogById(42);

    expect(api.get).toHaveBeenCalledWith('/api/v1/audit-logs/42');
    expect(result).toEqual(mockLog);
  });
});

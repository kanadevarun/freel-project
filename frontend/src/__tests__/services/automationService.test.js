import { describe, it, expect, vi, beforeEach } from 'vitest';
import { automationService } from '../../services/automationService';
import api from '../../services/api';

vi.mock('../../services/api', () => ({
  default: {
    get: vi.fn(),
    post: vi.fn(),
    put: vi.fn(),
    delete: vi.fn(),
  },
}));

describe('automationService', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('listAutomations builds correct query params', async () => {
    api.get.mockResolvedValue({ success: true, automations: [] });
    await automationService.listAutomations({ page: 2, limit: 10, search: 'invoice', is_enabled: true });
    expect(api.get).toHaveBeenCalledWith('/api/v1/automations?page=2&limit=10&is_enabled=true&search=invoice');
  });

  it('getAutomation calls correct endpoint', async () => {
    api.get.mockResolvedValue({ success: true, automation: { id: 1 } });
    await automationService.getAutomation(1);
    expect(api.get).toHaveBeenCalledWith('/api/v1/automations/1');
  });

  it('createAutomation calls POST with data', async () => {
    const payload = { name: 'Test Auto', automation_type: 'DAILY_OVERDUE_INVOICE_REVIEW' };
    api.post.mockResolvedValue({ success: true, automation: { id: 2 } });
    await automationService.createAutomation(payload);
    expect(api.post).toHaveBeenCalledWith('/api/v1/automations', payload);
  });

  it('enableAutomation and disableAutomation call respective endpoints', async () => {
    api.post.mockResolvedValue({ success: true });
    await automationService.enableAutomation(5);
    expect(api.post).toHaveBeenCalledWith('/api/v1/automations/5/enable', {});
    await automationService.disableAutomation(5);
    expect(api.post).toHaveBeenCalledWith('/api/v1/automations/5/disable', {});
  });

  it('retryExecution calls executions/:id/retry', async () => {
    api.post.mockResolvedValue({ success: true });
    await automationService.retryExecution(101);
    expect(api.post).toHaveBeenCalledWith('/api/v1/automations/executions/101/retry', {});
  });

  it('listInsights formats query parameters properly', async () => {
    api.get.mockResolvedValue({ success: true, insights: [] });
    await automationService.listInsights({ source_module: 'shipments', severity: 'HIGH', status: 'ACTIVE' });
    expect(api.get).toHaveBeenCalledWith('/api/v1/automations/insights?source_module=shipments&status=ACTIVE&severity=HIGH');
  });

  it('acknowledgeInsight and dismissInsight call endpoints', async () => {
    api.post.mockResolvedValue({ success: true });
    await automationService.acknowledgeInsight(55);
    expect(api.post).toHaveBeenCalledWith('/api/v1/automations/insights/55/acknowledge', {});
    await automationService.dismissInsight(55, 'Operator verified');
    expect(api.post).toHaveBeenCalledWith('/api/v1/automations/insights/55/dismiss', { reason: 'Operator verified' });
  });

  it('evaluateEvent calls evaluate endpoint', async () => {
    api.post.mockResolvedValue({ success: true, result: { event_processed: true } });
    const evt = { event_type: 'SHIPMENT_EXCEPTION_DETECTED', source_module: 'shipments', source_record_id: 101 };
    await automationService.evaluateEvent(evt);
    expect(api.post).toHaveBeenCalledWith('/api/v1/automations/evaluate', evt);
  });
});

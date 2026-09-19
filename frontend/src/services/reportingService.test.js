import { describe, test, expect, afterEach, vi } from 'vitest';
import { reportingService } from './reportingService';
import api from './api';

vi.mock('./api');

describe('reportingService', () => {
  afterEach(() => {
    vi.clearAllMocks();
  });

  test('getAdvancedReport queries with report_type and date_range parameters', async () => {
    const mockReport = {
      report_type: 'OPERATIONAL_VOLUME',
      date_range: 'LAST_90D',
      authoritative_metrics: { total_shipments: 42 },
      historical_series: [{ period: '2026-08', value: 20 }],
      forecast_series: [{ period: '2026-10', forecast_value: 25, is_forecast: true }],
      is_forecast_available: true,
      executive_narrative: 'Volume is trending upward.'
    };
    api.get.mockResolvedValueOnce({ data: mockReport });

    const result = await reportingService.getAdvancedReport('OPERATIONAL_VOLUME', 'LAST_90D');
    expect(api.get).toHaveBeenCalledWith('/api/v1/reports/advanced', {
      params: { report_type: 'OPERATIONAL_VOLUME', date_range: 'LAST_90D' },
    });
    expect(result.report_type).toBe('OPERATIONAL_VOLUME');
    expect(result.is_forecast_available).toBe(true);
    expect(result.forecast_series[0].is_forecast).toBe(true);
  });

  test('exportReport sends POST with export format and type', async () => {
    const mockExport = {
      id: 1,
      report_type: 'REVENUE_FINANCE',
      export_format: 'CSV',
      status: 'COMPLETED',
      export_content: 'Col1,Col2\nVal1,Val2',
      file_name: 'LogisticsHQ_REVENUE_FINANCE.csv'
    };
    api.post.mockResolvedValueOnce({ data: mockExport });

    const payload = { report_type: 'REVENUE_FINANCE', export_format: 'CSV', date_range: 'LAST_12M' };
    const result = await reportingService.exportReport(payload);
    expect(api.post).toHaveBeenCalledWith('/api/v1/reports/export', payload);
    expect(result.status).toBe('COMPLETED');
  });

  test('requestDistribution triggers human approval request gate', async () => {
    const mockDist = {
      id: 5,
      approval_id: 108,
      status: 'PENDING_APPROVAL',
      recipient_emails: 'auditor@external.com'
    };
    api.post.mockResolvedValueOnce({ data: mockDist });

    const payload = {
      report_type: 'CONTRACT_COMPLIANCE',
      recipient_emails: ['auditor@external.com'],
      distribution_channel: 'EMAIL',
      notes: 'Quarterly compliance distribution'
    };
    const result = await reportingService.requestDistribution(payload);
    expect(api.post).toHaveBeenCalledWith('/api/v1/reports/distribute', payload);
    expect(result.status).toBe('PENDING_APPROVAL');
    expect(result.approval_id).toBe(108);
  });

  test('getReportHistory and getDistributionHistory call appropriate endpoints', async () => {
    api.get.mockResolvedValueOnce({ data: { snapshots: [{ id: 1 }], total: 1 } });
    api.get.mockResolvedValueOnce({ data: { distributions: [{ id: 2 }], total: 1 } });

    const hist = await reportingService.getReportHistory(5);
    expect(api.get).toHaveBeenCalledWith('/api/v1/reports/history', { params: { limit: 5 } });
    expect(hist.length).toBe(1);

    const dists = await reportingService.getDistributionHistory(5);
    expect(api.get).toHaveBeenCalledWith('/api/v1/reports/distributions', { params: { limit: 5 } });
    expect(dists.length).toBe(1);
  });
});

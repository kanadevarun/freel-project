import { describe, it, expect, vi, beforeEach } from 'vitest';
import workforceService from '../../services/workforceService';
import api from '../../services/api';

vi.mock('../../services/api', () => ({
  default: {
    get: vi.fn(),
    post: vi.fn(),
    put: vi.fn(),
  },
}));

describe('workforceService', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('getCommandCenterOverview calls /workforce/command-center/overview', async () => {
    const mockData = { health_summary: { overall_status: 'HEALTHY' } };
    api.get.mockResolvedValueOnce(mockData);

    const result = await workforceService.getCommandCenterOverview();
    expect(api.get).toHaveBeenCalledWith('/api/v1/workforce/command-center/overview');
    expect(result).toEqual(mockData);
  });

  it('emergencyStop posts correct payload', async () => {
    const mockRes = { success: true, message: 'Emergency stop activated' };
    api.post.mockResolvedValueOnce(mockRes);

    const result = await workforceService.emergencyStop({
      scope: 'WORKFORCE',
      action: 'HALT',
      reason: 'Safety halt',
    });
    expect(api.post).toHaveBeenCalledWith('/api/v1/workforce/command-center/emergency-stop', {
      scope: 'WORKFORCE',
      action: 'HALT',
      reason: 'Safety halt',
    });
    expect(result).toEqual(mockRes);
  });

  it('controlAgent posts agent control changes', async () => {
    const mockRes = { success: true, operational_status: 'PAUSED' };
    api.post.mockResolvedValueOnce(mockRes);

    const result = await workforceService.controlAgent('shipment_agent', {
      operational_status: 'PAUSED',
      reason: 'Maintenance',
    });
    expect(api.post).toHaveBeenCalledWith('/api/v1/workforce/agents/shipment_agent/control', {
      operational_status: 'PAUSED',
      reason: 'Maintenance',
    });
    expect(result).toEqual(mockRes);
  });

  it('evaluateActionPolicy sends action for Go governance evaluation', async () => {
    const mockDecision = { decision: 'PREPARE_FOR_APPROVAL', requires_approval: true };
    api.post.mockResolvedValueOnce(mockDecision);

    const result = await workforceService.evaluateActionPolicy({
      agent_id: 'shipment_agent',
      action_type: 'TAG_INTERNAL_STATE',
      entity_type: 'SHIPMENT',
      entity_id: 'SHP-101',
    });
    expect(api.post).toHaveBeenCalledWith('/api/v1/workforce/command-center/evaluate-action', {
      agent_id: 'shipment_agent',
      action_type: 'TAG_INTERNAL_STATE',
      entity_type: 'SHIPMENT',
      entity_id: 'SHP-101',
    });
    expect(result).toEqual(mockDecision);
  });

  it('inspectWorkflow fetches plan details', async () => {
    const mockDetail = { plan_id: 'plan-101', status: 'COMPLETED' };
    api.get.mockResolvedValueOnce(mockDetail);

    const result = await workforceService.inspectWorkflow('plan-101');
    expect(api.get).toHaveBeenCalledWith('/api/v1/workforce/plans/plan-101/inspect');
    expect(result).toEqual(mockDetail);
  });
});

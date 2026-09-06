import { describe, it, expect } from 'vitest';
import { calculateRFQCompleteness } from '../../../pages/dashboard/RFQ/utils/completeness';

describe('calculateRFQCompleteness', () => {
  it('correctly calculates 7/7 completeness for a fully populated RFQ', () => {
    const rfq = {
      id: 101,
      customer_id: 42,
      customer_name: 'Global Corp',
      origin: 'Nhava Sheva (INNSA)',
      destination: 'Hamburg (DEHAM)',
      incoterms: 'FOB',
      target_date: '2026-09-15T00:00:00Z',
      stage: 'STAGE_RFQ_CREATED',
      items: [
        {
          id: 1,
          description: 'Automotive Parts',
          quantity: 50,
          weight_kg: 12000,
          volume_cbm: 28,
        }
      ]
    };

    const result = calculateRFQCompleteness(rfq);
    expect(result.completedCount).toBe(7);
    expect(result.totalCount).toBe(7);
    expect(result.percentage).toBe(100);
    expect(result.isComplete).toBe(true);
    expect(result.missingFields).toHaveLength(0);
    expect(result.health).toBe('HEALTHY');
    expect(result.operationalStatus).toBe('READY_FOR_QUOTATION');
    expect(result.statusLabel).toBe('Ready for Quotation');
    expect(result.statusColor).toBe('emerald');
  });

  it('correctly identifies missing fields and marks ATTENTION_REQUIRED when incomplete', () => {
    const rfq = {
      id: 102,
      customer_id: 42,
      origin: 'Nhava Sheva',
      destination: '', // missing
      incoterms: 'FOB',
      target_date: null, // missing
      stage: 'STAGE_RFQ_CREATED',
      items: [
        {
          id: 1,
          description: 'Machinery',
          weight_kg: 0, // invalid <= 0
          volume_cbm: 15,
        }
      ]
    };

    const result = calculateRFQCompleteness(rfq);
    expect(result.completedCount).toBe(4); // origin, incoterms, description, volume
    expect(result.isComplete).toBe(false);
    expect(result.percentage).toBe(57);
    expect(result.missingFields).toContain('Destination');
    expect(result.missingFields).toContain('Cargo Weight');
    expect(result.missingFields).toContain('Cargo Ready Date');
    expect(result.health).toBe('ATTENTION_REQUIRED');
    expect(result.operationalStatus).toBe('INFORMATION_REQUIRED');
    expect(result.statusLabel).toBe('Information Required');
    expect(result.statusColor).toBe('amber');
  });

  it('handles empty or null RFQ gracefully without crashing', () => {
    const result = calculateRFQCompleteness(null);
    expect(result.completedCount).toBe(0);
    expect(result.isComplete).toBe(false);
    expect(result.health).toBe('ATTENTION_REQUIRED');
  });
});

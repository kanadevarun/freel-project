/**
 * Utility for canonical RFQ information completeness and deterministic health calculation.
 * Evaluates the 7 mandatory shipment requirements:
 * 1. Origin
 * 2. Destination
 * 3. Incoterms
 * 4. Cargo Description
 * 5. Cargo Weight (> 0 KG)
 * 6. Cargo Volume (> 0 CBM)
 * 7. Cargo Ready Date (target_date)
 */

export function calculateRFQCompleteness(rfq) {
  if (!rfq) {
    return {
      fields: [],
      completedCount: 0,
      totalCount: 7,
      percentage: 0,
      isComplete: false,
      missingFields: [],
      health: 'ATTENTION_REQUIRED',
      operationalStatus: 'INFORMATION_REQUIRED',
      statusLabel: 'Information Required',
      statusColor: 'amber'
    };
  }

  const items = Array.isArray(rfq.items) ? rfq.items : [];

  const hasOrigin = Boolean(rfq.origin && typeof rfq.origin === 'string' && rfq.origin.trim().length > 0);
  const hasDestination = Boolean(rfq.destination && typeof rfq.destination === 'string' && rfq.destination.trim().length > 0);
  const hasIncoterms = Boolean(rfq.incoterms && typeof rfq.incoterms === 'string' && rfq.incoterms.trim().length > 0);
  const hasTargetDate = Boolean(rfq.target_date && !isNaN(new Date(rfq.target_date).getTime()));

  // Items validation
  const hasDescription = items.some(it => Boolean(it.description && it.description.trim().length > 0));
  const totalWeight = items.reduce((acc, it) => acc + (Number(it.weight_kg) || 0), 0);
  const totalVolume = items.reduce((acc, it) => acc + (Number(it.volume_cbm) || 0), 0);
  const totalQuantity = items.reduce((acc, it) => acc + (Number(it.quantity) || 1), 0);

  const hasWeight = totalWeight > 0;
  const hasVolume = totalVolume > 0;

  const fields = [
    {
      key: 'origin',
      label: 'Origin',
      value: rfq.origin || null,
      complete: hasOrigin,
    },
    {
      key: 'destination',
      label: 'Destination',
      value: rfq.destination || null,
      complete: hasDestination,
    },
    {
      key: 'incoterms',
      label: 'Incoterms',
      value: rfq.incoterms || null,
      complete: hasIncoterms,
    },
    {
      key: 'cargo_description',
      label: 'Cargo Description',
      value: items.find(it => it.description)?.description || null,
      complete: hasDescription,
    },
    {
      key: 'cargo_weight',
      label: 'Cargo Weight',
      value: hasWeight ? `${totalWeight.toLocaleString()} KG` : null,
      complete: hasWeight,
    },
    {
      key: 'cargo_volume',
      label: 'Cargo Volume',
      value: hasVolume ? `${totalVolume.toLocaleString()} CBM` : null,
      complete: hasVolume,
    },
    {
      key: 'target_date',
      label: 'Cargo Ready Date',
      value: hasTargetDate ? new Date(rfq.target_date).toLocaleDateString('en-US', { day: 'numeric', month: 'short', year: 'numeric' }) : null,
      complete: hasTargetDate,
    },
  ];

  const completedCount = fields.filter(f => f.complete).length;
  const totalCount = fields.length;
  const percentage = Math.round((completedCount / totalCount) * 100);
  const isComplete = completedCount === totalCount;
  const missingFields = fields.filter(f => !f.complete).map(f => f.label);

  // Deterministic Health calculation:
  // - HEALTHY: All 7 fields complete AND customer exists
  // - ATTENTION_REQUIRED: 1 or more fields missing OR no customer info
  const hasCustomer = Boolean(rfq.customer_name || rfq.customer_id > 0);
  const health = (isComplete && hasCustomer) ? 'HEALTHY' : 'ATTENTION_REQUIRED';

  // Derived Operational Status & Stage resolution:
  let operationalStatus = 'RFQ_CREATED';
  let statusLabel = 'Created';
  let statusColor = 'blue';

  const backendStage = rfq.stage || 'STAGE_RFQ_CREATED';

  if (backendStage === 'STAGE_RFQ_CREATED' || backendStage === 'DRAFT' || !backendStage) {
    if (isComplete) {
      operationalStatus = 'READY_FOR_QUOTATION';
      statusLabel = 'Ready for Quotation';
      statusColor = 'emerald';
    } else {
      operationalStatus = 'INFORMATION_REQUIRED';
      statusLabel = 'Information Required';
      statusColor = 'amber';
    }
  } else if (backendStage === 'STAGE_PRICING_ASSIGNED') {
    operationalStatus = 'PRICING_ASSIGNED';
    statusLabel = 'Pricing Assigned';
    statusColor = 'indigo';
  } else if (backendStage === 'STAGE_QUOTE_GENERATED') {
    operationalStatus = 'QUOTE_GENERATED';
    statusLabel = 'Quote Generated';
    statusColor = 'purple';
  } else if (backendStage === 'STAGE_QUOTE_SENT') {
    operationalStatus = 'QUOTE_SENT';
    statusLabel = 'Quote Sent';
    statusColor = 'blue';
  } else if (backendStage === 'STAGE_NEGOTIATION') {
    operationalStatus = 'NEGOTIATION';
    statusLabel = 'Negotiation';
    statusColor = 'violet';
  } else if (backendStage === 'STAGE_WON') {
    operationalStatus = 'WON';
    statusLabel = 'Won / Awarded';
    statusColor = 'green';
  } else if (backendStage === 'STAGE_LOST') {
    operationalStatus = 'LOST';
    statusLabel = 'Lost';
    statusColor = 'red';
  } else if (backendStage === 'STAGE_SHIPMENT_CREATED') {
    operationalStatus = 'SHIPMENT_CREATED';
    statusLabel = 'Shipment Created';
    statusColor = 'teal';
  }

  return {
    fields,
    completedCount,
    totalCount,
    percentage,
    isComplete,
    missingFields,
    totalWeight,
    totalVolume,
    totalQuantity,
    itemsCount: items.length,
    health,
    operationalStatus,
    statusLabel,
    statusColor,
  };
}

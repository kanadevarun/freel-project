/**
 * Defines the lifecycle stages of a Shipment Initiation Request (RFQ).
 */
export const RFQ_STAGES = {
  STAGE_RFQ_CREATED: 'STAGE_RFQ_CREATED',
  STAGE_PRICING_ASSIGNED: 'STAGE_PRICING_ASSIGNED',
  STAGE_QUOTE_GENERATED: 'STAGE_QUOTE_GENERATED',
  STAGE_QUOTE_SENT: 'STAGE_QUOTE_SENT',
  STAGE_NEGOTIATION: 'STAGE_NEGOTIATION',
  STAGE_WON: 'STAGE_WON',
  STAGE_LOST: 'STAGE_LOST',
  STAGE_SHIPMENT_CREATED: 'STAGE_SHIPMENT_CREATED',
};

/**
 * Maps the stage to a human-readable label and color for UI badges.
 */
export const STAGE_CONFIG = {
  [RFQ_STAGES.STAGE_RFQ_CREATED]: { label: 'RFQ Created', color: 'gray' },
  [RFQ_STAGES.STAGE_PRICING_ASSIGNED]: { label: 'Pricing Assigned', color: 'blue' },
  [RFQ_STAGES.STAGE_QUOTE_GENERATED]: { label: 'Quote Generated', color: 'indigo' },
  [RFQ_STAGES.STAGE_QUOTE_SENT]: { label: 'Quote Sent', color: 'purple' },
  [RFQ_STAGES.STAGE_NEGOTIATION]: { label: 'Negotiation', color: 'orange' },
  [RFQ_STAGES.STAGE_WON]: { label: 'Won', color: 'green' },
  [RFQ_STAGES.STAGE_LOST]: { label: 'Lost', color: 'red' },
  [RFQ_STAGES.STAGE_SHIPMENT_CREATED]: { label: 'Shipment Created', color: 'teal' },
};

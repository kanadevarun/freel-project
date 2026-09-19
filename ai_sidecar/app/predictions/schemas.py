from datetime import datetime
from enum import Enum
from typing import List, Optional, Any, Dict
from pydantic import BaseModel, Field, field_validator


class PredictionModule(str, Enum):
    SHIPMENTS = "shipments"
    RFQS = "rfqs"
    PRICING = "pricing"
    FINANCE = "finance"
    CONTRACTS = "contracts"
    CUSTOMERS = "customers"
    LEADS = "leads"
    CROSS_MODULE = "cross_module"
    CARRIERS = "carriers"
    NETWORK = "network"
    WORKLOAD = "workload"
    CAPACITY = "capacity"
    DEMAND = "demand"
    RESOURCE = "resource"
    BOTTLENECK = "bottleneck"


class PredictionType(str, Enum):
    OPERATIONAL_BOTTLENECK = "OPERATIONAL_BOTTLENECK"
    RESOURCE_ALLOCATION_IMBALANCE = "RESOURCE_ALLOCATION_IMBALANCE"
    APPROVAL_BOTTLENECK = "APPROVAL_BOTTLENECK"
    DOCUMENTATION_BOTTLENECK = "DOCUMENTATION_BOTTLENECK"
    EXCEPTION_RESOLUTION_BOTTLENECK = "EXCEPTION_RESOLUTION_BOTTLENECK"
    CROSS_MODULE_BOTTLENECK = "CROSS_MODULE_BOTTLENECK"
    CUTOFF_CONCENTRATION_BOTTLENECK = "CUTOFF_CONCENTRATION_BOTTLENECK"
    OWNER_WORKLOAD_IMBALANCE = "OWNER_WORKLOAD_IMBALANCE"
    SHIPMENT_ETA_DELAY = "SHIPMENT_ETA_DELAY"
    SHIPMENT_TRANSSHIPMENT_EXCEPTION = "SHIPMENT_TRANSSHIPMENT_EXCEPTION"
    SHIPMENT_EXCEPTION_RISK = "SHIPMENT_EXCEPTION_RISK"
    SHIPMENT_DISRUPTION_FORECAST = "SHIPMENT_DISRUPTION_FORECAST"
    INVOICE_PAYMENT_DEFAULT = "INVOICE_PAYMENT_DEFAULT"
    INVOICE_DISPUTE_PROBABILITY = "INVOICE_DISPUTE_PROBABILITY"
    CUSTOMER_CHURN_RISK = "CUSTOMER_CHURN_RISK"
    CUSTOMER_VOLUME_DROP = "CUSTOMER_VOLUME_DROP"
    CUSTOMER_REPEAT_BUSINESS = "CUSTOMER_REPEAT_BUSINESS"
    CUSTOMER_ENGAGEMENT_RISK = "CUSTOMER_ENGAGEMENT_RISK"
    LEAD_CONVERSION_LIKELIHOOD = "LEAD_CONVERSION_LIKELIHOOD"
    LEAD_INACTIVITY_RISK = "LEAD_INACTIVITY_RISK"
    LEAD_ENGAGEMENT_RISK = "LEAD_ENGAGEMENT_RISK"
    RFQ_WIN_PROBABILITY = "RFQ_WIN_PROBABILITY"
    CONTRACT_DEMURRAGE_RISK = "CONTRACT_DEMURRAGE_RISK"
    CONTRACT_COMPLIANCE_BREACH = "CONTRACT_COMPLIANCE_BREACH"
    CARRIER_RELIABILITY_DROP = "CARRIER_RELIABILITY_DROP"
    RFQ_MARGIN_RISK = "RFQ_MARGIN_RISK"
    QUOTATION_COMPETITIVENESS = "QUOTATION_COMPETITIVENESS"
    PRICING_COST_VARIANCE_RISK = "PRICING_COST_VARIANCE_RISK"
    CONTRACT_RATE_PRESSURE = "CONTRACT_RATE_PRESSURE"
    HISTORICAL_MARGIN_INTELLIGENCE = "HISTORICAL_MARGIN_INTELLIGENCE"
    INVOICE_LATE_PAYMENT_RISK = "INVOICE_LATE_PAYMENT_RISK"
    COLLECTION_PRIORITY = "COLLECTION_PRIORITY"
    CASH_INFLOW_FORECAST = "CASH_INFLOW_FORECAST"
    DISPUTE_PAYMENT_DELAY_RISK = "DISPUTE_PAYMENT_DELAY_RISK"
    CUSTOMER_PAYMENT_BEHAVIOR = "CUSTOMER_PAYMENT_BEHAVIOR"
    RECEIVABLES_CONCENTRATION_RISK = "RECEIVABLES_CONCENTRATION_RISK"
    CONTRACT_EXPIRY_RENEWAL_RISK = "CONTRACT_EXPIRY_RENEWAL_RISK"
    CONTRACT_CLAUSE_COMMERCIAL_RISK = "CONTRACT_CLAUSE_COMMERCIAL_RISK"
    DOCUMENTATION_COMPLETENESS_RISK = "DOCUMENTATION_COMPLETENESS_RISK"
    COMPLIANCE_REVIEW_RISK = "COMPLIANCE_REVIEW_RISK"
    CROSS_MODULE_CONTRACT_RISK = "CROSS_MODULE_CONTRACT_RISK"
    HISTORICAL_DOCUMENTATION_RISK = "HISTORICAL_DOCUMENTATION_RISK"
    SHIPMENT_READINESS_RISK = "SHIPMENT_READINESS_RISK"
    CUTOFF_MISS_RISK = "CUTOFF_MISS_RISK"
    DOCUMENTATION_DELAY_RISK = "DOCUMENTATION_DELAY_RISK"
    CUSTOMS_PROCESSING_RISK = "CUSTOMS_PROCESSING_RISK"
    BILLING_READINESS_RISK = "BILLING_READINESS_RISK"
    HISTORICAL_OPERATIONAL_RISK = "HISTORICAL_OPERATIONAL_RISK"
    CUSTOMER_SERVICE_RISK = "CUSTOMER_SERVICE_RISK"
    CUSTOMER_RELATIONSHIP_RISK = "CUSTOMER_RELATIONSHIP_RISK"
    CARRIER_PERFORMANCE_RISK = "CARRIER_PERFORMANCE_RISK"
    CARRIER_DELAY_RISK = "CARRIER_DELAY_RISK"
    LANE_PERFORMANCE_RISK = "LANE_PERFORMANCE_RISK"
    LANE_DISRUPTION_RISK = "LANE_DISRUPTION_RISK"
    NETWORK_BOTTLENECK_RISK = "NETWORK_BOTTLENECK_RISK"
    SERVICE_LEVEL_RISK = "SERVICE_LEVEL_RISK"
    COST_PRESSURE_RISK = "COST_PRESSURE_RISK"
    APPROVAL_WORKLOAD_SPIKE = "APPROVAL_WORKLOAD_SPIKE"
    OPERATIONAL_WORKLOAD_SPIKE = "OPERATIONAL_WORKLOAD_SPIKE"
    DOCUMENTATION_WORKLOAD_SPIKE = "DOCUMENTATION_WORKLOAD_SPIKE"
    DEMAND_CAPACITY_MISMATCH = "DEMAND_CAPACITY_MISMATCH"
    LANE_CAPACITY_PRESSURE = "LANE_CAPACITY_PRESSURE"
    CARRIER_CAPACITY_PRESSURE = "CARRIER_CAPACITY_PRESSURE"
    QUOTE_PROCESSING_BOTTLENECK = "QUOTE_PROCESSING_BOTTLENECK"
    SHIPMENT_PROCESSING_BOTTLENECK = "SHIPMENT_PROCESSING_BOTTLENECK"
    DEMAND_VOLUME_FORECAST = "DEMAND_VOLUME_FORECAST"
    CAPACITY_SHORTAGE_RISK = "CAPACITY_SHORTAGE_RISK"
    PORT_CONGESTION_RISK = "PORT_CONGESTION_RISK"
    CARRIER_UPDATE_GAP_RISK = "CARRIER_UPDATE_GAP_RISK"
    FREE_TIME_EXPIRY_RISK = "FREE_TIME_EXPIRY_RISK"
    NETWORK_DISRUPTION_INTELLIGENCE = "NETWORK_DISRUPTION_INTELLIGENCE"



class PredictionSeverity(str, Enum):
    LOW = "LOW"
    MEDIUM = "MEDIUM"
    HIGH = "HIGH"
    CRITICAL = "CRITICAL"


class ConfidenceBand(str, Enum):
    LOW = "LOW"
    MEDIUM = "MEDIUM"
    HIGH = "HIGH"


class PredictionSourceReference(BaseModel):
    source_module: str = Field(..., description="Originating module, e.g. shipments, invoices")
    source_record_id: str = Field(..., description="ID of authoritative record")
    source_field: str = Field(..., description="Specific field or milestone, e.g. eta, vessel_milestone, invoice_status")
    source_timestamp: str = Field(..., description="ISO8601 timestamp of source fact")
    document_reference: Optional[str] = Field(None, description="Filename or page/clause reference if document-derived")
    clause_reference: Optional[str] = Field(None, description="Clause ID or section identifier")
    page_number: Optional[int] = Field(None, description="Page number where clause or evidence is located")
    section_heading: Optional[str] = Field(None, description="Section heading in source document")
    milestone_reference: Optional[str] = Field(None, description="Milestone code or event reference")
    cutoff_reference: Optional[str] = Field(None, description="Cutoff description or deadline type")
    comparison_period: Optional[str] = Field(None, description="Comparison window e.g. LAST_90_DAYS")
    sample_size: Optional[int] = Field(None, description="Number of historical records in sample")
    lane_reference: Optional[str] = Field(None, description="Trade corridor e.g. INNSA-NLRTM")
    carrier_reference: Optional[str] = Field(None, description="Carrier SCAC e.g. MAEU, MSCU")
    customer_reference: Optional[str] = Field(None, description="Customer code or identifier")
    workload_type: Optional[str] = Field(None, description="Type of workload e.g. APPROVALS, DOCUMENTATION, RFQS")
    pending_count: Optional[int] = Field(None, description="Number of pending tasks or items")
    capacity_limit: Optional[int] = Field(None, description="Capacity threshold or throughput limit")
    utilization_rate: Optional[float] = Field(None, description="Capacity utilization percentage")
    bottleneck_type: Optional[str] = Field(None, description="Type of bottleneck e.g. APPROVALS, DOCUMENTATION, EXCEPTIONS, RESOURCE")
    affected_stage: Optional[str] = Field(None, description="Operational stage affected e.g. CUSTOMS_CLEARANCE, MANIFEST_DEADLINE, PRICING_APPROVAL")
    assigned_owner: Optional[str] = Field(None, description="Current assigned owner or team")
    queue_dwell_hours: Optional[float] = Field(None, description="Average queue dwell hours")
    source_type: Optional[str] = Field(None, description="INTERNAL_RECORD, EXTERNAL_VERIFIED, or UNVERIFIED_SIGNAL")
    source_url: Optional[str] = Field(None, description="External verification URL if verified")
    port_reference: Optional[str] = Field(None, description="Port or terminal code e.g. NLRTM, INNSA, USNYC")
    data_freshness_seconds: Optional[int] = Field(None, description="Age in seconds of source observation")


class PredictionSupportingSignal(BaseModel):
    signal_name: str = Field(..., description="Name of the analytical signal")
    observed_value: Any = Field(..., description="Real observed value from context")
    baseline_value: Optional[Any] = Field(None, description="Standard baseline or benchmark")
    importance_weight: float = Field(default=1.0, ge=0.0, le=1.0, description="Relative weighting 0.0-1.0")


class GeneratePredictionRequest(BaseModel):
    org_id: int = Field(..., description="Organization ID for multi-tenant isolation")
    module: PredictionModule = Field(..., description="Target business domain")
    prediction_type: PredictionType = Field(..., description="Type of prediction requested")
    related_record_type: str = Field(..., description="SHIPMENT, INVOICE, LEAD, RFQ, CONTRACT, CUSTOMER, CARRIER, LANE")
    related_record_id: str = Field(..., description="Authoritative record ID")
    record_context: Dict[str, Any] = Field(default_factory=dict, description="Verified source facts from Go backend")
    time_horizon: Optional[str] = Field("7_DAYS", description="Forecast time horizon e.g. 48_HOURS, 7_DAYS, 30_DAYS")
    correlation_id: Optional[str] = Field(None, description="Tracing identifier")


class GeneratePredictionResponse(BaseModel):
    prediction_id: str = Field(..., description="Deterministic prediction ID (e.g. pred-xxx)")
    org_id: int = Field(..., description="Validated org ID")
    module: PredictionModule
    prediction_type: PredictionType
    related_record_type: str
    related_record_id: str
    prediction_statement: str = Field(..., max_length=1000, description="Clear concise summary of predicted event")
    predicted_value: Optional[str] = Field(None, max_length=255)
    prediction_category: Optional[str] = Field(None, description="Category of prediction e.g. CONVERSION_LIKELIHOOD, INACTIVITY_RISK, REPEAT_BUSINESS, ENGAGEMENT_RISK")
    disruption_category: Optional[str] = Field(None, description="Early warning disruption category e.g. MILESTONE_DELAY, TRACKING_INACTIVITY, CUSTOMS_CLEARANCE_RISK, FREE_TIME_EXPOSURE")
    document_reference: Optional[str] = Field(None, description="Associated document reference e.g. maersk_contract_2026.pdf")
    clause_reference: Optional[str] = Field(None, description="Associated clause reference e.g. CLAUSE-PAY-01 or Section 4.1")
    page_number: Optional[int] = Field(None, description="Page number where clause or discrepancy is located")
    section_heading: Optional[str] = Field(None, description="Section heading in source document")
    milestone_reference: Optional[str] = Field(None, description="Next operational milestone or event reference e.g. GATE_IN, ARRIVAL")
    cutoff_reference: Optional[str] = Field(None, description="Operational cutoff deadline e.g. DOC_CUTOFF, PORT_CUTOFF")
    comparison_period: Optional[str] = Field(None, description="Comparison benchmark window e.g. LAST_90_DAYS")
    sample_size: Optional[int] = Field(None, description="Total authoritative sample records evaluated")
    lane_reference: Optional[str] = Field(None, description="Corridor or trade route e.g. INNSA-NLRTM")
    carrier_reference: Optional[str] = Field(None, description="Carrier SCAC e.g. MAEU, MSCU, CMDU")
    customer_reference: Optional[str] = Field(None, description="Customer trading name or code")
    workload_type: Optional[str] = Field(None, description="Type of workload e.g. APPROVALS, DOCUMENTATION, RFQS")
    pending_count: Optional[int] = Field(None, description="Number of pending tasks or items")
    capacity_limit: Optional[int] = Field(None, description="Capacity threshold or throughput limit")
    utilization_rate: Optional[float] = Field(None, description="Capacity utilization percentage")
    bottleneck_type: Optional[str] = Field(None, description="Type of bottleneck e.g. APPROVALS, DOCUMENTATION, EXCEPTIONS, RESOURCE")
    affected_stage: Optional[str] = Field(None, description="Operational stage affected e.g. CUSTOMS_CLEARANCE, MANIFEST_DEADLINE, PRICING_APPROVAL")
    assigned_owner: Optional[str] = Field(None, description="Current assigned owner or team")
    queue_dwell_hours: Optional[float] = Field(None, description="Average queue dwell hours")
    source_type: Optional[str] = Field(None, description="INTERNAL_RECORD, EXTERNAL_VERIFIED, or UNVERIFIED_SIGNAL")
    source_url: Optional[str] = Field(None, description="External verification URL if verified")
    port_reference: Optional[str] = Field(None, description="Port or terminal code e.g. NLRTM, INNSA, USNYC")
    linked_exception_id: Optional[int] = Field(None, description="Linked existing confirmed exception ID")
    predicted_arrival_window: Optional[str] = Field(None, description="Predicted arrival window start to end")
    predicted_delay_hours: Optional[float] = Field(None, description="Estimated delay in hours")
    time_horizon: Optional[str] = Field("7_DAYS")
    target_date: Optional[str] = Field(None, description="ISO8601 target forecast date")
    severity: PredictionSeverity
    confidence_score: float = Field(..., ge=0.0, le=1.0)
    confidence_band: ConfidenceBand
    explanation: str = Field(..., description="Source-grounded rationale explaining prediction")
    supporting_signals: List[PredictionSupportingSignal] = Field(default_factory=list)
    source_references: List[PredictionSourceReference] = Field(..., min_length=1, description="Must cite at least 1 real source")
    source_timestamp: str = Field(..., description="Latest source data timestamp")
    recommended_action: Optional[str] = Field(None, max_length=1000)
    action_type: Optional[str] = Field(None, description="Registered action system action if actionable")
    is_action_required: bool = Field(default=False)
    requires_approval: bool = Field(default=False)
    insufficient_data: bool = Field(default=False)
    insufficient_data_reason: Optional[str] = None
    model_version: str = Field(default="gemini-1.5-pro")
    created_at: str = Field(default_factory=lambda: datetime.utcnow().isoformat())

    @field_validator("confidence_band", mode="before")
    def derive_confidence_band(cls, v, values):
        if v:
            return v
        score = values.data.get("confidence_score", 0.5)
        if score >= 0.8:
            return ConfidenceBand.HIGH
        elif score >= 0.5:
            return ConfidenceBand.MEDIUM
        return ConfidenceBand.LOW

import os
import sys
import time
import httpx
import asyncio
from dotenv import load_dotenv

# Load environment variables from .env file before importing state or graphs
load_dotenv()

# Startup check: Ensure INTERNAL_SERVICE_TOKEN is present in production environments
if os.getenv("APP_ENV") == "production" and not os.getenv("INTERNAL_SERVICE_TOKEN"):
    print("❌ Configuration error: INTERNAL_SERVICE_TOKEN must be specified in production environments")
    sys.exit(1)

# Add current directory to python path to resolve local imports cleanly
sys.path.append(os.path.dirname(os.path.abspath(__file__)))

from typing import List, Optional, Dict, Any
import json
from fastapi import FastAPI, BackgroundTasks, HTTPException, status, Depends
from pydantic import BaseModel

from app.tools.auth_utils import get_internal_service_token, require_internal_service_key

from app.state.contract_state import ContractExtractionState, ExtractedContractDraft
from app.agents.parser_agent import parse_contract_agreement
from app.graphs.contracts_graph import contracts_graph
from app.graphs.pricing_graph import pricing_graph
from app.persistence.queue_worker import QueueWorker

app = FastAPI(title="LogisticsHQ Rate Intelligence AI Sidecar", version="2.0.0")

class ProcessingRequest(BaseModel):
    document_id: str
    org_id: int
    s3_key: str
    file_type: str
    callback_url: str
    correlation_id: Optional[str] = None

class ResumeRequest(BaseModel):
    document_id: str
    org_id: int
    action: str # "APPROVE" | "REJECT"
    corrected_rates: Optional[List[Dict[str, Any]]] = None
    notes: Optional[str] = None
    callback_url: str
    correlation_id: Optional[str] = None

class Surcharge(BaseModel):
    code: str
    description: str
    amount: float
    unit: str
    included: bool

class CanonicalRateDraft(BaseModel):
    origin_port: str
    destination_port: str
    via_port: Optional[str] = None
    service_code: Optional[str] = None
    carrier_scac: str
    carrier_name: str
    vessel_name: Optional[str] = None
    equipment_type: str
    ocean_freight: float
    origin_charges: float
    destination_charges: float
    surcharges: List[Surcharge] = []
    total_buy_price: float
    currency_original: str = "USD"
    exchange_rate_used: float = 1.0
    included_charges: List[str] = []
    excluded_charges: List[str] = []
    free_days_origin: int = 0
    free_days_destination: int = 14
    transit_days: Optional[int] = None
    incoterms: Optional[str] = None
    commodity_restrictions: List[str] = []
    routing_conditions: Optional[str] = None
    valid_from: str
    valid_until: str
    confidence_score: int

class ReviewItemDraft(BaseModel):
    extracted_data: CanonicalRateDraft
    confidence_score: int
    review_flags: List[str]
    ai_reasoning: str
    source_page: int
    source_text: str
    source_image_url: str

class LogEntry(BaseModel):
    step: str
    timestamp: str
    message: str

class AIProcessingCallback(BaseModel):
    document_id: str
    org_id: int
    status: str
    confirmed_rates: List[CanonicalRateDraft] = []
    flagged_items: List[ReviewItemDraft] = []
    processing_log: List[LogEntry] = []
    ai_summary: str
    correlation_id: Optional[str] = None

async def run_langgraph_pipeline(req: ProcessingRequest):
    print(f"[AI Sidecar][Correlation ID: {req.correlation_id or 'None'}] Starting LangGraph pipeline for doc {req.document_id}...")
    
    # Initialize state values matching the schema
    initial_state = {
        "document_id": req.document_id,
        "org_id": req.org_id,
        "s3_key": req.s3_key,
        "file_type": req.file_type,
        "callback_url": req.callback_url,
        "correlation_id": req.correlation_id or "",
        "raw_text": "",
        "carrier_name": None,
        "carrier_scac": None,
        "extracted_rates": [],
        "flagged_items": [],
        "processing_log": [],
        "ai_summary": "",
        "is_anomaly_detected": False,
        "status": "QUEUED"
    }

    config = {
        "configurable": {"thread_id": req.document_id, "org_id": req.org_id},
        "metadata": {"org_id": str(req.org_id), "correlation_id": req.correlation_id or ""},
    }
    
    try:
        # Run graph. It will run through ocr -> classify -> parser -> validator.
        # If validator node sets is_anomaly_detected=True, the graph will interrupt *before* ingest node.
        # The invoke method will return the state at the point of interruption.
        result = contracts_graph.invoke(initial_state, config=config)
    except Exception as e:
        print(f"[AI Sidecar][Correlation ID: {req.correlation_id or 'None'}] LangGraph execution failed: {e}")
        # Return error callback to Go backend
        callback_payload = {
            "document_id": req.document_id,
            "org_id": req.org_id,
            "status": "FAILED",
            "processing_log": [
                {
                    "step": "CRITICAL_ERROR",
                    "timestamp": time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime()),
                    "message": f"Graph execution crashed: {str(e)}"
                }
            ],
            "ai_summary": "",
            "correlation_id": req.correlation_id
        }
        await send_callback(req.callback_url, callback_payload)
        return

    # Check snapshot state to see if it paused at the interrupt point
    snapshot = contracts_graph.get_state(config)
    is_interrupted = len(snapshot.next) > 0
    state_values = snapshot.values

    # Determine status & log messages
    status_str = "COMPLETED"
    if is_interrupted:
        print(f"[AI Sidecar] LangGraph execution paused for document {req.document_id} at node: {snapshot.next}")
        # Note: If it's interrupted, we set the callback status to COMPLETED but populate flagged_items
        # so that Go places it in the human review queue.
    
    # Map logs
    logs_out = []
    for log in state_values.get("processing_log", []):
        logs_out.append(LogEntry(
            step=log.step,
            timestamp=log.timestamp,
            message=log.message
        ))

    # Map confirmed rates and flagged review items
    confirmed_rates = []
    flagged_items = []

    for r in state_values.get("extracted_rates", []):
        confirmed_rates.append(map_rate_draft(r))

    for item in state_values.get("flagged_items", []):
        flagged_items.append(ReviewItemDraft(
            extracted_data=map_rate_draft(item.extracted_data),
            confidence_score=item.confidence_score,
            review_flags=item.review_flags,
            ai_reasoning=item.ai_reasoning,
            source_page=item.source_page,
            source_text=item.source_text,
            source_image_url=item.source_image_url
        ))

    carrier_name = state_values.get("carrier_name", "Unknown Carrier")
    carrier_scac = state_values.get("carrier_scac", "XXXX")

    callback_payload = AIProcessingCallback(
        document_id=req.document_id,
        org_id=req.org_id,
        status=status_str,
        confirmed_rates=confirmed_rates,
        flagged_items=flagged_items,
        processing_log=logs_out,
        ai_summary=f"Parsed carrier contract from {carrier_name} ({carrier_scac}). Anomaly detected: {state_values.get('is_anomaly_detected', False)}.",
        correlation_id=req.correlation_id
    )

    await send_callback(req.callback_url, callback_payload.model_dump())

async def run_resume_pipeline(req: ResumeRequest):
    config = {
        "configurable": {"thread_id": req.document_id, "org_id": req.org_id},
        "metadata": {"org_id": str(req.org_id), "correlation_id": req.correlation_id or ""},
    }

    # Load current state
    snapshot = contracts_graph.get_state(config)
    if not snapshot or not snapshot.values:
        print(f"[AI Sidecar] Thread ID {req.document_id} not found. Cannot resume.")
        return

    # Tenant isolation validation
    if snapshot.metadata and snapshot.metadata.get("org_id"):
        saved_org = int(snapshot.metadata.get("org_id"))
        if saved_org != req.org_id:
            print(f"[AI Sidecar] Tenant isolation violation: thread {req.document_id} belongs to org {saved_org}, resume rejected for org {req.org_id}")
            return

    # Workflow replay prevention: reject resuming already completed/terminated workflows
    if not snapshot.next:
        print(f"[AI Sidecar] Thread ID {req.document_id} has no pending interrupt steps (already completed or terminated). Resume replay rejected.")
        return

    # Update state based on action
    if req.action == "APPROVE":
        # Resolve anomaly flag
        update_data = {
            "is_anomaly_detected": False,
            "flagged_items": [],
            "processing_log": snapshot.values.get("processing_log", []) + [
                {
                    "step": "HUMAN_RESUME",
                    "timestamp": time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime()),
                    "message": f"Human approved/corrected the flagged items. Notes: {req.notes or 'None'}"
                }
            ]
        }

        # If manual corrections were provided, update extracted_rates with them
        if req.corrected_rates:
            new_rates = []
            for r in req.corrected_rates:
                new_rates.append(map_dict_to_rate_draft(r, snapshot.values.get("carrier_scac"), snapshot.values.get("carrier_name")))
            update_data["extracted_rates"] = new_rates

        contracts_graph.update_state(config, update_data, as_node="validator")
        
        # Resume flow by passing None (resume from current interrupt checkpoint)
        try:
            contracts_graph.invoke(None, config=config)
        except Exception as e:
            print(f"[AI Sidecar] Failed to resume graph: {e}")
            return

        # Fetch finalized values
        final_snapshot = contracts_graph.get_state(config)
        final_values = final_snapshot.values

        # Build callback detailing approved rates
        confirmed_rates = [map_rate_draft(r) for r in final_values.get("extracted_rates", [])]
        logs_out = []
        for log in final_values.get("processing_log", []):
            # Safe parsing
            if hasattr(log, "step"):
                logs_out.append(LogEntry(step=log.step, timestamp=log.timestamp, message=log.message))
            else:
                logs_out.append(LogEntry(step=log.get("step"), timestamp=log.get("timestamp"), message=log.get("message")))

        callback_payload = AIProcessingCallback(
            document_id=req.document_id,
            org_id=req.org_id,
            status="COMPLETED",
            confirmed_rates=confirmed_rates,
            flagged_items=[],
            processing_log=logs_out,
            ai_summary=f"Contract approved by human. Rates successfully ingested.",
            correlation_id=req.correlation_id
        )
        await send_callback(req.callback_url, callback_payload.model_dump())

    elif req.action == "REJECT":
        # Mark as rejected
        update_data = {
            "is_anomaly_detected": False,
            "flagged_items": [],
            "extracted_rates": [],
            "status": "FAILED",
            "processing_log": snapshot.values.get("processing_log", []) + [
                {
                    "step": "HUMAN_REJECT",
                    "timestamp": time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime()),
                    "message": f"Human operator rejected the extraction. Notes: {req.notes or 'None'}"
                }
            ]
        }
        contracts_graph.update_state(config, update_data, as_node="validator")
        
        final_snapshot = contracts_graph.get_state(config)
        final_values = final_snapshot.values
        
        logs_out = []
        for log in final_values.get("processing_log", []):
            if hasattr(log, "step"):
                logs_out.append(LogEntry(step=log.step, timestamp=log.timestamp, message=log.message))
            else:
                logs_out.append(LogEntry(step=log.get("step"), timestamp=log.get("timestamp"), message=log.get("message")))

        callback_payload = AIProcessingCallback(
            document_id=req.document_id,
            org_id=req.org_id,
            status="FAILED",
            confirmed_rates=[],
            flagged_items=[],
            processing_log=logs_out,
            ai_summary="Contract extraction rejected by human operator.",
            correlation_id=req.correlation_id
        )
        await send_callback(req.callback_url, callback_payload.model_dump())

def map_rate_draft(r) -> CanonicalRateDraft:
    """Helper to convert StateRateDraft to Callback CanonicalRateDraft."""
    surcharges = []
    for s in (r.surcharges or []):
        surcharges.append(Surcharge(
            code=s.code,
            description=s.description,
            amount=s.amount,
            unit=s.unit,
            included=s.included
        ))
    
    return CanonicalRateDraft(
        origin_port=r.origin_port,
        destination_port=r.destination_port,
        via_port=r.via_port,
        service_code=r.service_code,
        carrier_scac=r.carrier_scac,
        carrier_name=r.carrier_name,
        vessel_name=r.vessel_name,
        equipment_type=r.equipment_type,
        ocean_freight=r.ocean_freight,
        origin_charges=r.origin_charges,
        destination_charges=r.destination_charges,
        surcharges=surcharges,
        total_buy_price=r.total_buy_price,
        currency_original=r.currency_original,
        exchange_rate_used=r.exchange_rate_used,
        included_charges=r.included_charges,
        excluded_charges=r.excluded_charges,
        free_days_origin=r.free_days_origin,
        free_days_destination=r.free_days_destination,
        transit_days=r.transit_days,
        incoterms=r.incoterms,
        commodity_restrictions=r.commodity_restrictions,
        routing_conditions=r.routing_conditions,
        valid_from=r.valid_from,
        valid_until=r.valid_until,
        confidence_score=r.confidence_score
    )

def map_dict_to_rate_draft(d: dict, carrier_scac: str, carrier_name: str) -> Any:
    """Helper to convert generic dictionary to StateRateDraft pydantic class."""
    from app.state.contract_state import StateRateDraft, StateSurcharge
    surcharges = []
    for s in d.get("surcharges", []):
        surcharges.append(StateSurcharge(
            code=s.get("code", "SUR"),
            description=s.get("description", ""),
            amount=float(s.get("amount", 0.0)),
            unit=s.get("unit", "PER_TEU"),
            included=bool(s.get("included", False))
        ))
    
    return StateRateDraft(
        origin_port=d.get("origin_port", "INNSA"),
        destination_port=d.get("destination_port", "DEHAM"),
        via_port=d.get("via_port"),
        service_code=d.get("service_code"),
        carrier_scac=d.get("carrier_scac") or carrier_scac or "XXXX",
        carrier_name=d.get("carrier_name") or carrier_name or "Unknown",
        vessel_name=d.get("vessel_name"),
        equipment_type=d.get("equipment_type", "40GP"),
        ocean_freight=float(d.get("ocean_freight", 0.0)),
        origin_charges=float(d.get("origin_charges", 0.0)),
        destination_charges=float(d.get("destination_charges", 0.0)),
        surcharges=surcharges,
        total_buy_price=float(d.get("total_buy_price", 0.0)),
        currency_original=d.get("currency_original", "USD"),
        exchange_rate_used=float(d.get("exchange_rate_used", 1.0)),
        included_charges=d.get("included_charges", []),
        excluded_charges=d.get("excluded_charges", []),
        free_days_origin=int(d.get("free_days_origin", 0)),
        free_days_destination=int(d.get("free_days_destination", 14)),
        transit_days=d.get("transit_days"),
        incoterms=d.get("incoterms"),
        commodity_restrictions=d.get("commodity_restrictions", []),
        routing_conditions=d.get("routing_conditions"),
        valid_from=d.get("valid_from", "2026-09-01T00:00:00Z"),
        valid_until=d.get("valid_until", "2026-12-31T23:59:59Z"),
        confidence_score=100
    )

async def run_pricing_pipeline(org_id: int, entity_id: str, payload: dict):
    rfq_id = int(entity_id)
    correlation_id = payload.get("correlation_id", "")
    callback_url = payload.get("callback_url", "")
    
    print(f"[AI Sidecar Pricing] Starting pricing analysis for RFQ #{rfq_id} (Correlation ID: {correlation_id})")
    
    # 1. Fetch initial RFQ details to populate state
    from app.tools.pricing_tool import get_rfq_details_tool
    rfq_details_str = get_rfq_details_tool.func(rfq_id=rfq_id, org_id=org_id)
    
    try:
        rfq_data = json.loads(rfq_details_str)
        if isinstance(rfq_data, str) and rfq_data.startswith("Error"):
            raise ValueError(rfq_data)
        rfq_info = rfq_data.get("data", rfq_data)
    except Exception as e:
        print(f"[AI Sidecar Pricing] Failed to load RFQ details: {e}")
        await send_callback(callback_url, {
            "rfq_id": rfq_id,
            "org_id": org_id,
            "status": "FAILED",
            "correlation_id": correlation_id,
            "ai_reasoning": f"Failed to retrieve RFQ details from backend: {str(e)}"
        })
        return

    # Extract items details for weight, volume, commodity
    origin = rfq_info.get("origin") or ""
    dest = rfq_info.get("destination") or ""
    incoterms = rfq_info.get("incoterms") or ""
    equipment_type = "40GP"
    gross_weight = 0.0
    volume_cbm = 0.0
    commodity = ""
    target_date = rfq_info.get("target_date")

    items = rfq_info.get("items") or []
    if items:
        commodity = items[0].get("description", "")
        for item in items:
            gross_weight += item.get("weight_kg") or 0.0
            volume_cbm += item.get("volume_cbm") or 0.0

    # 2. Build initial state
    initial_state = {
        "messages": [],
        "org_id": org_id,
        "rfq_id": rfq_id,
        "origin": origin,
        "destination": dest,
        "incoterms": incoterms,
        "equipment_type": equipment_type,
        "gross_weight": gross_weight,
        "volume_cbm": volume_cbm,
        "commodity": commodity,
        "target_date": target_date,
        "pricing_rules": [],
        "raw_rates": [],
        "suggested_quotes": [],
        "overall_reasoning": "",
        "is_anomaly": False,
        "confidence_score": 100,
        "error_message": None
    }

    config = {
        "configurable": {"thread_id": f"rfq-{rfq_id}", "org_id": org_id},
        "metadata": {
            "correlation_id": correlation_id,
            "org_id": str(org_id),
            "task_type": "PRICING_ANALYZE",
        }
    }

    try:
        # Run graph
        result = pricing_graph.invoke(initial_state, config=config)
    except Exception as e:
        print(f"[AI Sidecar Pricing] Graph execution failed: {e}")
        await send_callback(callback_url, {
            "rfq_id": rfq_id,
            "org_id": org_id,
            "status": "FAILED",
            "correlation_id": correlation_id,
            "ai_reasoning": f"Graph execution failed: {str(e)}"
        })
        return

    # Check if graph paused due to anomaly interrupt
    snapshot = pricing_graph.get_state(config)
    if snapshot and snapshot.next:
        print(f"[AI Sidecar Pricing] Workflow paused at interrupt {snapshot.next} awaiting human approval.")
        await send_callback(callback_url, {
            "rfq_id": rfq_id,
            "org_id": org_id,
            "status": "WAITING_FOR_HUMAN",
            "correlation_id": correlation_id,
            "ai_reasoning": "Pricing anomaly detected. Workflow halted for human operator sign-off."
        })
        return

    # If it completed without interrupt, trigger callback with results
    final_snapshot = pricing_graph.get_state(config)
    overall_reasoning = final_snapshot.values.get("overall_reasoning", "") if final_snapshot else ""
    
    # Save quotes to backend
    suggested_quotes = final_snapshot.values.get("suggested_quotes", []) if final_snapshot else []
    try:
        await save_pricing_quotes_to_backend(org_id, rfq_id, suggested_quotes)
        await send_callback(callback_url, {
            "rfq_id": rfq_id,
            "org_id": org_id,
            "status": "COMPLETED",
            "correlation_id": correlation_id,
            "ai_reasoning": overall_reasoning
        })
    except Exception as save_err:
        await send_callback(callback_url, {
            "rfq_id": rfq_id,
            "org_id": org_id,
            "status": "FAILED",
            "correlation_id": correlation_id,
            "ai_reasoning": f"Failed to persist quotes: {str(save_err)}"
        })

async def run_pricing_resume_pipeline(org_id: int, entity_id: str, payload: dict):
    clean_id = entity_id.replace("rfq-", "").strip() if entity_id else ""
    try:
        rfq_id = int(clean_id)
    except (ValueError, TypeError):
        raise ValueError(f"Invalid non-integer rfq_id '{entity_id}' for PRICING_RESUME")

    correlation_id = payload.get("correlation_id", "")
    callback_url = payload.get("callback_url", "")
    
    print(f"[AI Sidecar Pricing] Resuming pricing analysis for RFQ #{rfq_id} (Correlation ID: {correlation_id})")
    
    config = {
        "configurable": {"thread_id": f"rfq-{rfq_id}", "org_id": org_id},
        "metadata": {
            "correlation_id": correlation_id,
            "org_id": str(org_id),
            "task_type": "PRICING_RESUME",
        }
    }
    
    snapshot = pricing_graph.get_state(config)
    if not snapshot or not snapshot.values:
        print(f"[AI Sidecar Pricing] Thread rfq-{rfq_id} not found. Cannot resume.")
        return

    # Tenant isolation validation
    if snapshot.metadata and snapshot.metadata.get("org_id"):
        saved_org = int(snapshot.metadata.get("org_id"))
        if saved_org != org_id:
            print(f"[AI Sidecar Pricing] Tenant isolation violation: thread rfq-{rfq_id} belongs to org {saved_org}, resume rejected for org {org_id}")
            await send_callback(callback_url, {
                "rfq_id": rfq_id,
                "org_id": org_id,
                "status": "FAILED",
                "correlation_id": correlation_id,
                "ai_reasoning": "Tenant isolation violation: cross-tenant access denied"
            })
            return

    # Workflow replay prevention: reject resuming already completed/terminated workflows
    if not snapshot.next:
        print(f"[AI Sidecar Pricing] Thread rfq-{rfq_id} has no pending interrupt steps (already completed or terminated). Resume replay rejected.")
        await send_callback(callback_url, {
            "rfq_id": rfq_id,
            "org_id": org_id,
            "status": "COMPLETED",
            "correlation_id": correlation_id,
            "ai_reasoning": "Workflow already completed; resume was a no-op."
        })
        return

    # Update anomaly status before resuming
    pricing_graph.update_state(config, {"is_anomaly": False})
    
    try:
        # Resume graph execution
        pricing_graph.invoke(None, config=config)
        
        final_snapshot = pricing_graph.get_state(config)
        overall_reasoning = final_snapshot.values.get("overall_reasoning", "") if final_snapshot else ""
        
        await send_callback(callback_url, {
            "rfq_id": rfq_id,
            "org_id": org_id,
            "status": "COMPLETED",
            "correlation_id": correlation_id,
            "ai_reasoning": overall_reasoning
        })
    except Exception as e:
        print(f"[AI Sidecar Pricing] Resume failed: {e}")
        await send_callback(callback_url, {
            "rfq_id": rfq_id,
            "org_id": org_id,
            "status": "FAILED",
            "correlation_id": correlation_id,
            "ai_reasoning": f"Graph resume failed: {str(e)}"
        })

async def send_callback(callback_url: str, payload: dict):
    """Sends JSON results to Go backend, routed through Centralized Action System."""
    try:
        from app.tools.action_bridge import execute_action
        org_id = payload.get("org_id", 1)
        doc_id = payload.get("document_id", "")
        action_res = execute_action(
            action_name="contracts.ingest_rates",
            org_id=org_id,
            input_data={
                "document_id": doc_id,
                "status": payload.get("status", "COMPLETED"),
                "confirmed_rates": payload.get("confirmed_rates", []),
                "flagged_items": payload.get("flagged_items", []),
                "ai_summary": payload.get("ai_summary", ""),
                "correlation_id": payload.get("correlation_id", "")
            },
            source="langgraph.contracts"
        )
        if action_res.get("success"):
            print(f"[AI Sidecar] Rate extraction ingested via Action System for doc {doc_id}")
            return
    except Exception as ex:
        print(f"[AI Sidecar] Action System contract ingestion error: {ex}")

    token = get_internal_service_token()
    headers = {"X-LogisticsHQ-Service-Key": token}
    
    async with httpx.AsyncClient() as client:
        try:
            resp = await client.post(callback_url, json=payload, headers=headers, timeout=15.0)
            print(f"[AI Sidecar] Fallback callback response: status={resp.status_code}, body={resp.text}")
        except Exception as e:
            print(f"[AI Sidecar] Failed to send fallback callback request: {e}")

@app.post("/process", status_code=status.HTTP_202_ACCEPTED, dependencies=[Depends(require_internal_service_key)])
async def process_document(req: ProcessingRequest, background_tasks: BackgroundTasks):
    """
    HTTP POST /process
    
    What it does:
      Receives a trigger request from the Go backend indicating a new contract PDF has been uploaded.
      It offloads the actual LangGraph pipeline execution to a FastAPI background worker thread
      so that the API returns immediately with '202 Accepted' instead of blocking the Go caller.
      
    Example JSON Request:
      {
        "document_id": "3ae5c3ab-51a2-4a0b-9cc3-1a224fbc11e3",
        "org_id": 5,
        "s3_key": "contract_123.pdf",
        "file_type": "PDF",
        "callback_url": "http://localhost:8080/internal/contracts/callback"
      }
    """
    print(f"[AI Sidecar] Received process request for document: {req.document_id}")
    # Enqueue pipeline task run
    background_tasks.add_task(run_langgraph_pipeline, req)
    return {"message": "Processing request queued successfully"}

@app.post("/resume", status_code=status.HTTP_202_ACCEPTED, dependencies=[Depends(require_internal_service_key)])
async def resume_document(req: ResumeRequest, background_tasks: BackgroundTasks):
    """
    HTTP POST /resume
    
    What it does:
      Triggered by the Go backend when a human operator resolves a paused review state.
      Queues a worker task to update the checkpointed graph variables and resume graph execution.
      
    Example JSON Request:
      {
        "document_id": "3ae5c3ab-51a2-4a0b-9cc3-1a224fbc11e3",
        "org_id": 5,
        "action": "APPROVE",
        "corrected_rates": [ { "origin_port": "INNSA", ... } ],
        "notes": "Rates corrected",
        "callback_url": "http://localhost:8080/internal/contracts/callback"
      }
    """
    print(f"[AI Sidecar] Received resume request for document: {req.document_id}, action={req.action}")
    # Enqueue resume task run
    background_tasks.add_task(run_resume_pipeline, req)
    return {"message": "Resume request queued successfully"}

@app.get("/health")
async def health_check():
    from app.persistence.checkpointer import get_checkpointer_info
    return {"status": "ok", **get_checkpointer_info()}

class ExtractAgreementRequest(BaseModel):
    document_id: Optional[str] = None
    org_id: int
    raw_text: Optional[str] = ""
    file_name: Optional[str] = "contract_agreement.pdf"
    s3_key: Optional[str] = None

@app.post("/contracts/extract-agreement", dependencies=[Depends(require_internal_service_key)])
async def extract_contract_agreement(req: ExtractAgreementRequest):
    """
    HTTP POST /contracts/extract-agreement
    
    Extracts full commercial contract agreement metadata, parties, commercial terms,
    validity dates, and obligations from an agreement document.
    """
    print(f"[AI Sidecar] Extracting contract agreement for org={req.org_id}, file={req.file_name}")
    raw_text = req.raw_text or ""
    if not raw_text and req.s3_key:
        possible_paths = [
            os.path.join(os.path.dirname(os.path.dirname(os.path.abspath(__file__))), "backend", req.s3_key),
            os.path.join(os.path.dirname(os.path.dirname(os.path.abspath(__file__))), "backend", "internal", "contracts", req.s3_key),
            os.path.join("/tmp", req.s3_key)
        ]
        for p in possible_paths:
            if os.path.exists(p):
                try:
                    import pypdf
                    reader = pypdf.PdfReader(p)
                    raw_text = "\n".join([page.extract_text() or "" for page in reader.pages])
                    break
                except Exception as e:
                    print(f"[AI Sidecar] Failed to read PDF at {p}: {e}")

    extracted = parse_contract_agreement(raw_text, req.file_name or "agreement.pdf")
    return {"status": "SUCCESS", "data": extracted}

@app.post("/orchestrator/propose-action", dependencies=[Depends(require_internal_service_key)])
async def orchestrator_propose_action(req: dict):
    """
    HTTP POST /orchestrator/propose-action
    
    Phase 3 Task 3.2: Controlled AI Workflow Execution and Action Orchestration.
    Receives structured operational context from the Go backend, analyzes signals,
    evaluates evidence, and returns a strictly validated ActionProposal.
    """
    from app.orchestrator.models import OrchestrationContextInput
    from app.orchestrator.agent import AIActionOrchestrator
    
    try:
        ctx = OrchestrationContextInput(**req)
    except Exception as parse_err:
        raise HTTPException(
            status_code=status.HTTP_422_UNPROCESSABLE_ENTITY,
            detail=f"Invalid orchestration context payload: {str(parse_err)}"
        )
        
    orchestrator = AIActionOrchestrator()
    resp = orchestrator.analyze_and_propose(ctx)
    return resp.model_dump()


@app.post("/customer-relationship/evaluate", dependencies=[Depends(require_internal_service_key)])
async def customer_relationship_evaluate(req: dict):
    """
    HTTP POST /customer-relationship/evaluate
    Phase 3 Task 3.3: Customer Relationship Automation & Intelligent Follow-Up.
    Receives structured customer context from the Go backend, analyzes inactivity,
    overdue balances, exceptions, and pending quotes, and returns transparent priority ratings.
    """
    from app.customer_relationship.models import CustomerFollowupContext
    from app.customer_relationship.agent import CustomerRelationshipAgent

    try:
        ctx = CustomerFollowupContext(**req)
    except Exception as parse_err:
        raise HTTPException(
            status_code=status.HTTP_422_UNPROCESSABLE_ENTITY,
            detail=f"Invalid customer relationship context payload: {str(parse_err)}"
        )

    agent = CustomerRelationshipAgent()
    resp = agent.evaluate_customer(ctx)
    return resp.model_dump()


@app.post("/customer-relationship/draft", dependencies=[Depends(require_internal_service_key)])
async def customer_relationship_draft(req: dict):
    """
    HTTP POST /customer-relationship/draft
    Phase 3 Task 3.3: Grounded Customer Communication Drafting.
    Explicitly requested draft generation grounded exclusively on verified MariaDB facts.
    """
    from app.customer_relationship.models import DraftMessageRequest
    from app.customer_relationship.agent import CustomerRelationshipAgent

    try:
        draft_req = DraftMessageRequest(**req)
    except Exception as parse_err:
        raise HTTPException(
            status_code=status.HTTP_422_UNPROCESSABLE_ENTITY,
            detail=f"Invalid draft message request payload: {str(parse_err)}"
        )

    agent = CustomerRelationshipAgent()
    resp = agent.draft_communication(draft_req)
    return resp.model_dump()


# ==============================================================================
# Phase 3 Task 3.4: RFQ-to-Quotation Automation & Intelligent Pricing Workflow
# ==============================================================================

@app.post("/rfq-pricing/extract-requirements", dependencies=[Depends(require_internal_service_key)])
async def rfq_pricing_extract_requirements(req: dict):
    """
    HTTP POST /rfq-pricing/extract-requirements
    Phase 3 Task 3.4: RFQ Intake and Requirement Extraction.
    Evaluates sanitized RFQ parameters, identifies missing mandatory information,
    and returns grounded extraction evidence without fabricating values.
    """
    from app.rfq_pricing.models import RFQContext
    from app.rfq_pricing.agent import RFQPricingWorkflowAgent

    try:
        ctx = RFQContext(**req)
    except Exception as parse_err:
        raise HTTPException(
            status_code=status.HTTP_422_UNPROCESSABLE_ENTITY,
            detail=f"Invalid RFQ context payload: {str(parse_err)}"
        )

    agent = RFQPricingWorkflowAgent()
    resp = agent.extract_requirements(ctx)
    return resp.model_dump()


@app.post("/rfq-pricing/explain-pricing", dependencies=[Depends(require_internal_service_key)])
async def rfq_pricing_explain_pricing(req: dict):
    """
    HTTP POST /rfq-pricing/explain-pricing
    Phase 3 Task 3.4: Grounded Pricing Explanation.
    Explains deterministic pricing and cost/sell components calculated by Go.
    """
    from app.rfq_pricing.models import RFQContext, DeterministicPricingFacts
    from app.rfq_pricing.agent import RFQPricingWorkflowAgent

    try:
        ctx_data = req.get("context", {})
        pricing_data = req.get("pricing", {})
        ctx = RFQContext(**ctx_data)
        pricing = DeterministicPricingFacts(**pricing_data)
    except Exception as parse_err:
        raise HTTPException(
            status_code=status.HTTP_422_UNPROCESSABLE_ENTITY,
            detail=f"Invalid pricing explanation payload: {str(parse_err)}"
        )

    agent = RFQPricingWorkflowAgent()
    resp = agent.explain_pricing(ctx, pricing)
    return resp.model_dump()


@app.post("/rfq-pricing/analyze-risks", dependencies=[Depends(require_internal_service_key)])
async def rfq_pricing_analyze_risks(req: dict):
    """
    HTTP POST /rfq-pricing/analyze-risks
    Phase 3 Task 3.4: Quotation Risk Analysis.
    Detects commercial/operational risks and determines HITL approval requirements.
    """
    from app.rfq_pricing.models import RFQContext, DeterministicPricingFacts
    from app.rfq_pricing.agent import RFQPricingWorkflowAgent

    try:
        ctx_data = req.get("context", {})
        pricing_data = req.get("pricing", {})
        ctx = RFQContext(**ctx_data)
        pricing = DeterministicPricingFacts(**pricing_data)
    except Exception as parse_err:
        raise HTTPException(
            status_code=status.HTTP_422_UNPROCESSABLE_ENTITY,
            detail=f"Invalid risk analysis payload: {str(parse_err)}"
        )

    agent = RFQPricingWorkflowAgent()
    resp = agent.analyze_quotation_risks(ctx, pricing)
    return resp.model_dump()


@app.post("/rfq-pricing/generate-draft", dependencies=[Depends(require_internal_service_key)])
async def rfq_pricing_generate_draft(req: dict):
    """
    HTTP POST /rfq-pricing/generate-draft
    Phase 3 Task 3.4: Grounded Quotation Drafting.
    Prepares internal summary, customer-facing proposal wording, and terms.
    """
    from app.rfq_pricing.models import QuotationDraftRequest
    from app.rfq_pricing.agent import RFQPricingWorkflowAgent

    try:
        draft_req = QuotationDraftRequest(**req)
    except Exception as parse_err:
        raise HTTPException(
            status_code=status.HTTP_422_UNPROCESSABLE_ENTITY,
            detail=f"Invalid quotation draft request payload: {str(parse_err)}"
        )

    agent = RFQPricingWorkflowAgent()
    resp = agent.generate_quotation_draft(draft_req)
    return resp.model_dump()


# =========================================================================
# Phase 3 Task 3.5: Shipment Operations Automation & Intelligent Exception Response
# =========================================================================

@app.post("/shipment-ops/analyze-risks", dependencies=[Depends(require_internal_service_key)])
async def shipment_ops_analyze_risks(req: dict):
    """
    HTTP POST /shipment-ops/analyze-risks
    Phase 3 Task 3.5: Shipment Risk Analysis.
    Evaluates multi-signal operational risks grounded in Go backend facts.
    """
    from app.shipment_ops.models import ShipmentContext, DeterministicShipmentSignals
    from app.shipment_ops.agent import ShipmentOpsAgent

    try:
        ctx = ShipmentContext(**req.get("context", {}))
        signals = DeterministicShipmentSignals(**req.get("signals", {}))
    except Exception as parse_err:
        print(f"[SHIPMENT-OPS ERROR] analyze-risks parse error: {parse_err}")
        raise HTTPException(
            status_code=status.HTTP_422_UNPROCESSABLE_ENTITY,
            detail=f"Invalid shipment context or signals payload: {str(parse_err)}"
        )

    agent = ShipmentOpsAgent()
    resp = agent.analyze_shipment_risks(ctx, signals)
    return resp.model_dump()


@app.post("/shipment-ops/prioritize-exceptions", dependencies=[Depends(require_internal_service_key)])
async def shipment_ops_prioritize_exceptions(req: dict):
    """
    HTTP POST /shipment-ops/prioritize-exceptions
    Phase 3 Task 3.5: Active Exception Prioritization.
    Ranks active exceptions by operational urgency, financial exposure, and downstream impact.
    """
    from app.shipment_ops.models import ShipmentContext, DeterministicShipmentSignals
    from app.shipment_ops.agent import ShipmentOpsAgent

    try:
        ctx = ShipmentContext(**req.get("context", {}))
        signals = DeterministicShipmentSignals(**req.get("signals", {}))
    except Exception as parse_err:
        raise HTTPException(
            status_code=status.HTTP_422_UNPROCESSABLE_ENTITY,
            detail=f"Invalid shipment context or signals payload: {str(parse_err)}"
        )

    agent = ShipmentOpsAgent()
    resp = agent.prioritize_exceptions(ctx, signals)
    return resp.model_dump()


@app.post("/shipment-ops/recommend-actions", dependencies=[Depends(require_internal_service_key)])
async def shipment_ops_recommend_actions(req: dict):
    """
    HTTP POST /shipment-ops/recommend-actions
    Phase 3 Task 3.5: Operational Recommendations.
    Generates bounded operational action recommendations mapped to Action System.
    """
    from app.shipment_ops.models import ShipmentContext, DeterministicShipmentSignals
    from app.shipment_ops.agent import ShipmentOpsAgent

    try:
        ctx = ShipmentContext(**req.get("context", {}))
        signals = DeterministicShipmentSignals(**req.get("signals", {}))
    except Exception as parse_err:
        raise HTTPException(
            status_code=status.HTTP_422_UNPROCESSABLE_ENTITY,
            detail=f"Invalid shipment context or signals payload: {str(parse_err)}"
        )

    agent = ShipmentOpsAgent()
    resp = agent.generate_operational_recommendations(ctx, signals)
    return resp.model_dump()


@app.post("/shipment-ops/generate-draft", dependencies=[Depends(require_internal_service_key)])
async def shipment_ops_generate_draft(req: dict):
    """
    HTTP POST /shipment-ops/generate-draft
    Phase 3 Task 3.5: Communication Drafting.
    Prepares editable carrier follow-up, customer update, or internal escalation drafts.
    """
    from app.shipment_ops.models import CommunicationDraftRequest
    from app.shipment_ops.agent import ShipmentOpsAgent

    try:
        draft_req = CommunicationDraftRequest(**req)
    except Exception as parse_err:
        raise HTTPException(
            status_code=status.HTTP_422_UNPROCESSABLE_ENTITY,
            detail=f"Invalid communication draft request payload: {str(parse_err)}"
        )

    agent = ShipmentOpsAgent()
    resp = agent.generate_communication_draft(draft_req)
    return resp.model_dump()


# =========================================================================
# Phase 3 Task 3.6: Finance & Collections Automation
# =========================================================================

@app.post("/finance-ops/analyze-receivables", dependencies=[Depends(require_internal_service_key)])
async def finance_ops_analyze_receivables(req: dict):
    """
    HTTP POST /finance-ops/analyze-receivables
    Phase 3 Task 3.6: Receivables Risk Analysis.
    Evaluates multi-signal financial exposure grounded in Go backend facts.
    """
    from app.finance_ops.models import InvoiceContext, DeterministicFinanceSignals
    from app.finance_ops.agent import FinanceCollectionsAgent

    try:
        ctx = InvoiceContext(**req.get("context", {}))
        signals = DeterministicFinanceSignals(**req.get("signals", {}))
    except Exception as parse_err:
        print(f"[FINANCE-OPS ERROR] analyze-receivables parse error: {parse_err}")
        raise HTTPException(
            status_code=status.HTTP_422_UNPROCESSABLE_ENTITY,
            detail=f"Invalid invoice context or signals payload: {str(parse_err)}"
        )

    agent = FinanceCollectionsAgent()
    resp = agent.analyze_receivables_risk(ctx, signals)
    return resp.model_dump()


@app.post("/finance-ops/prioritize-collections", dependencies=[Depends(require_internal_service_key)])
async def finance_ops_prioritize_collections(req: dict):
    """
    HTTP POST /finance-ops/prioritize-collections
    Phase 3 Task 3.6: Collection Prioritization.
    Ranks receivables urgency based on aging, exposure, and multi-invoice risk.
    """
    from app.finance_ops.models import InvoiceContext, DeterministicFinanceSignals
    from app.finance_ops.agent import FinanceCollectionsAgent

    try:
        ctx = InvoiceContext(**req.get("context", {}))
        signals = DeterministicFinanceSignals(**req.get("signals", {}))
    except Exception as parse_err:
        print(f"[FINANCE-OPS ERROR] prioritize-collections parse error: {parse_err}")
        raise HTTPException(
            status_code=status.HTTP_422_UNPROCESSABLE_ENTITY,
            detail=f"Invalid invoice context or signals payload: {str(parse_err)}"
        )

    agent = FinanceCollectionsAgent()
    resp = agent.prioritize_collections(ctx, signals)
    return resp.model_dump()


@app.post("/finance-ops/analyze-customer-behavior", dependencies=[Depends(require_internal_service_key)])
async def finance_ops_analyze_customer_behavior(req: dict):
    """
    HTTP POST /finance-ops/analyze-customer-behavior
    Phase 3 Task 3.6: Customer Payment Behavior Analysis.
    Evaluates customer payment habits and credit risk based on historical settlement patterns.
    """
    from app.finance_ops.models import InvoiceContext, DeterministicFinanceSignals
    from app.finance_ops.agent import FinanceCollectionsAgent

    try:
        ctx = InvoiceContext(**req.get("context", {}))
        signals = DeterministicFinanceSignals(**req.get("signals", {}))
    except Exception as parse_err:
        print(f"[FINANCE-OPS ERROR] analyze-customer-behavior parse error: {parse_err}")
        raise HTTPException(
            status_code=status.HTTP_422_UNPROCESSABLE_ENTITY,
            detail=f"Invalid invoice context or signals payload: {str(parse_err)}"
        )

    agent = FinanceCollectionsAgent()
    resp = agent.analyze_customer_payment_behavior(ctx, signals)
    return resp.model_dump()


@app.post("/finance-ops/recommend-actions", dependencies=[Depends(require_internal_service_key)])
async def finance_ops_recommend_actions(req: dict):
    """
    HTTP POST /finance-ops/recommend-actions
    Phase 3 Task 3.6: Operational Recommendations.
    Generates bounded action recommendations mapped to Centralized Action System.
    """
    from app.finance_ops.models import InvoiceContext, DeterministicFinanceSignals
    from app.finance_ops.agent import FinanceCollectionsAgent

    try:
        ctx = InvoiceContext(**req.get("context", {}))
        signals = DeterministicFinanceSignals(**req.get("signals", {}))
    except Exception as parse_err:
        print(f"[FINANCE-OPS ERROR] recommend-actions parse error: {parse_err}")
        raise HTTPException(
            status_code=status.HTTP_422_UNPROCESSABLE_ENTITY,
            detail=f"Invalid invoice context or signals payload: {str(parse_err)}"
        )

    agent = FinanceCollectionsAgent()
    resp = agent.generate_operational_recommendations(ctx, signals)
    return resp.model_dump()


@app.post("/finance-ops/generate-collection-draft", dependencies=[Depends(require_internal_service_key)])
async def finance_ops_generate_collection_draft(req: dict):
    """
    HTTP POST /finance-ops/generate-collection-draft
    Phase 3 Task 3.6: Collection Message Draft Synthesis.
    Prepares editable collection reminder, overdue notice, or demand drafts with HITL gates.
    """
    from app.finance_ops.models import CollectionDraftRequest
    from app.finance_ops.agent import FinanceCollectionsAgent

    try:
        draft_req = CollectionDraftRequest(**req)
    except Exception as parse_err:
        print(f"[FINANCE-OPS ERROR] generate-collection-draft parse error: {parse_err}")
        raise HTTPException(
            status_code=status.HTTP_422_UNPROCESSABLE_ENTITY,
            detail=f"Invalid collection draft request payload: {str(parse_err)}"
        )

    agent = FinanceCollectionsAgent()
    resp = agent.generate_collection_draft(draft_req)
    return resp.model_dump()


# =========================================================================
# Phase 3 Task 3.7: Contract and Compliance Automation & Document Review
# =========================================================================

@app.post("/contract-compliance/review-contract", dependencies=[Depends(require_internal_service_key)])
async def contract_compliance_review(req: dict):
    """
    HTTP POST /contract-compliance/review-contract
    Phase 3 Task 3.7: Multi-dimensional contract and compliance review.
    Evaluates clause terms, structured discrepancies, and compliance obligations.
    """
    from app.contract_compliance.models import ContractDocumentContext, DeterministicComplianceSignals
    from app.contract_compliance.agent import ContractComplianceAgent

    try:
        ctx_data = req.get("context", {})
        signals_data = req.get("signals", {})
        context = ContractDocumentContext.model_validate(ctx_data)
        signals = DeterministicComplianceSignals.model_validate(signals_data)
    except Exception as parse_err:
        print(f"[CONTRACT-COMPLIANCE ERROR] review-contract parse error: {parse_err}")
        raise HTTPException(
            status_code=status.HTTP_422_UNPROCESSABLE_ENTITY,
            detail=f"Invalid contract context or signals payload: {str(parse_err)}"
        )

    agent = ContractComplianceAgent()
    resp = agent.review_contract(context, signals)
    return resp.model_dump()


@app.post("/contract-compliance/extract-clauses", dependencies=[Depends(require_internal_service_key)])
async def contract_compliance_extract_clauses(req: dict):
    """
    HTTP POST /contract-compliance/extract-clauses
    Phase 3 Task 3.7: Deep contractual clause extraction and risk classification.
    """
    from app.contract_compliance.models import ContractDocumentContext, DeterministicComplianceSignals
    from app.contract_compliance.agent import ContractComplianceAgent

    try:
        ctx_data = req.get("context", {})
        signals_data = req.get("signals", {})
        context = ContractDocumentContext.model_validate(ctx_data)
        signals = DeterministicComplianceSignals.model_validate(signals_data)
    except Exception as parse_err:
        raise HTTPException(
            status_code=status.HTTP_422_UNPROCESSABLE_ENTITY,
            detail=f"Invalid contract context or signals payload: {str(parse_err)}"
        )

    agent = ContractComplianceAgent()
    resp = agent.extract_clauses(context, signals)
    return resp.model_dump()


@app.post("/contract-compliance/verify-structured-terms", dependencies=[Depends(require_internal_service_key)])
async def contract_compliance_verify_structured_terms(req: dict):
    """
    HTTP POST /contract-compliance/verify-structured-terms
    Phase 3 Task 3.7: Verifies document terms against structured database records.
    """
    from app.contract_compliance.models import ContractDocumentContext, DeterministicComplianceSignals
    from app.contract_compliance.agent import ContractComplianceAgent

    try:
        ctx_data = req.get("context", {})
        signals_data = req.get("signals", {})
        context = ContractDocumentContext.model_validate(ctx_data)
        signals = DeterministicComplianceSignals.model_validate(signals_data)
    except Exception as parse_err:
        raise HTTPException(
            status_code=status.HTTP_422_UNPROCESSABLE_ENTITY,
            detail=f"Invalid contract context or signals payload: {str(parse_err)}"
        )

    agent = ContractComplianceAgent()
    resp = agent.verify_structured_terms(context, signals)
    return resp.model_dump()


@app.post("/contract-compliance/assess-compliance", dependencies=[Depends(require_internal_service_key)])
async def contract_compliance_assess(req: dict):
    """
    HTTP POST /contract-compliance/assess-compliance
    Phase 3 Task 3.7: Evaluates regulatory, cargo liability, and customs compliance checklists.
    """
    from app.contract_compliance.models import ContractDocumentContext, DeterministicComplianceSignals
    from app.contract_compliance.agent import ContractComplianceAgent

    try:
        ctx_data = req.get("context", {})
        signals_data = req.get("signals", {})
        context = ContractDocumentContext.model_validate(ctx_data)
        signals = DeterministicComplianceSignals.model_validate(signals_data)
    except Exception as parse_err:
        raise HTTPException(
            status_code=status.HTTP_422_UNPROCESSABLE_ENTITY,
            detail=f"Invalid contract context or signals payload: {str(parse_err)}"
        )

    agent = ContractComplianceAgent()
    resp = agent.assess_compliance(context, signals)
    return resp.model_dump()


@app.post("/contract-compliance/generate-clarification-draft", dependencies=[Depends(require_internal_service_key)])
async def contract_compliance_generate_draft(req: dict):
    """
    HTTP POST /contract-compliance/generate-clarification-draft
    Phase 3 Task 3.7: Generates editable review notes, missing-document requests, or renewal proposals.
    Requires HITL approval before dispatch.
    """
    from app.contract_compliance.models import ClarificationDraftRequest
    from app.contract_compliance.agent import ContractComplianceAgent

    try:
        draft_req = ClarificationDraftRequest.model_validate(req)
    except Exception as parse_err:
        print(f"[CONTRACT-COMPLIANCE ERROR] generate-clarification-draft parse error: {parse_err}")
        raise HTTPException(
            status_code=status.HTTP_422_UNPROCESSABLE_ENTITY,
            detail=f"Invalid clarification draft request payload: {str(parse_err)}"
        )

    agent = ContractComplianceAgent()
    resp = agent.generate_clarification_draft(draft_req)
    return resp.model_dump()


# =====================================================================
# Phase 3 Task 3.8: Event-Driven AI Workflows & Cross-Module Automation
# =====================================================================

@app.post("/event-workflows/analyze-event", dependencies=[Depends(require_internal_service_key)])
async def event_workflows_analyze_event(req: dict):
    """
    HTTP POST /event-workflows/analyze-event
    Phase 3 Task 3.8: Evaluates an incoming domain event, synthesizes cross-module
    context, and generates actionable operational recommendations.
    """
    from app.event_workflows.models import EventWorkflowRequest
    from app.event_workflows.agent import EventWorkflowsAgent

    try:
        wf_req = EventWorkflowRequest.model_validate(req)
    except Exception as parse_err:
        print(f"[EVENT-WORKFLOWS ERROR] analyze-event parse error: {parse_err}")
        raise HTTPException(
            status_code=status.HTTP_422_UNPROCESSABLE_ENTITY,
            detail=f"Invalid event workflow request payload: {str(parse_err)}"
        )

    agent = EventWorkflowsAgent()
    resp = agent.analyze_event(wf_req)
    return resp.model_dump()


@app.post("/event-workflows/generate-draft", dependencies=[Depends(require_internal_service_key)])
async def event_workflows_generate_draft(req: dict):
    """
    HTTP POST /event-workflows/generate-draft
    Phase 3 Task 3.8: Generates an editable communication draft in response to a domain event.
    Requires HITL approval before dispatch.
    """
    from app.event_workflows.models import WorkflowDraftRequest
    from app.event_workflows.agent import EventWorkflowsAgent

    try:
        draft_req = WorkflowDraftRequest.model_validate(req)
    except Exception as parse_err:
        print(f"[EVENT-WORKFLOWS ERROR] generate-draft parse error: {parse_err}")
        raise HTTPException(
            status_code=status.HTTP_422_UNPROCESSABLE_ENTITY,
            detail=f"Invalid workflow draft request payload: {str(parse_err)}"
        )

    agent = EventWorkflowsAgent()
    resp = agent.generate_draft(draft_req)
    return resp.model_dump()


# =====================================================================
# Phase 3 Task 3.9: Advanced Notifications & Escalations
# =====================================================================

@app.post("/notifications-escalations/analyze-and-prioritize", dependencies=[Depends(require_internal_service_key)])
async def notifications_analyze_and_prioritize(req: dict):
    """
    HTTP POST /notifications-escalations/analyze-and-prioritize
    Phase 3 Task 3.9: Evaluates notification priority, escalation trajectory,
    semantic cluster key, and alert overload reduction advice.
    """
    from app.notifications_escalations.models import NotificationAnalysisRequest
    from app.notifications_escalations.agent import NotificationsEscalationsAgent

    try:
        notif_req = NotificationAnalysisRequest.model_validate(req)
    except Exception as parse_err:
        print(f"[NOTIFICATIONS-ESCALATIONS ERROR] analyze-and-prioritize parse error: {parse_err}")
        raise HTTPException(
            status_code=status.HTTP_422_UNPROCESSABLE_ENTITY,
            detail=f"Invalid notification analysis request payload: {str(parse_err)}"
        )

    agent = NotificationsEscalationsAgent()
    resp = agent.analyze_and_prioritize(notif_req)
    return resp.model_dump()


@app.post("/notifications-escalations/generate-escalation-draft", dependencies=[Depends(require_internal_service_key)])
async def notifications_generate_escalation_draft(req: dict):
    """
    HTTP POST /notifications-escalations/generate-escalation-draft
    Phase 3 Task 3.9: Generates editable internal escalation memo or external client advisory.
    Requires HITL approval before dispatch.
    """
    from app.notifications_escalations.models import EscalationDraftRequest
    from app.notifications_escalations.agent import NotificationsEscalationsAgent

    try:
        draft_req = EscalationDraftRequest.model_validate(req)
    except Exception as parse_err:
        print(f"[NOTIFICATIONS-ESCALATIONS ERROR] generate-escalation-draft parse error: {parse_err}")
        raise HTTPException(
            status_code=status.HTTP_422_UNPROCESSABLE_ENTITY,
            detail=f"Invalid escalation draft request payload: {str(parse_err)}"
        )

    agent = NotificationsEscalationsAgent()
    resp = agent.generate_escalation_draft(draft_req)
    return resp.model_dump()


@app.post("/copilot/chat", dependencies=[Depends(require_internal_service_key)])
async def copilot_chat_endpoint(req: dict):
    """
    HTTP POST /copilot/chat
    Phase 3 Task 3.10: Omni-present AI Copilot across every module in LogisticsHQ.
    Provides context-aware conversational guidance, record explanations,
    operational risk insights, draft synthesis, and controlled action proposals.
    """
    from app.copilot.models import CopilotChatRequest
    from app.copilot.agent import CopilotAgent

    try:
        copilot_req = CopilotChatRequest.model_validate(req)
    except Exception as parse_err:
        print(f"[COPILOT ERROR] chat payload parse error: {parse_err}")
        raise HTTPException(
            status_code=status.HTTP_422_UNPROCESSABLE_ENTITY,
            detail=f"Invalid copilot chat request payload: {str(parse_err)}"
        )

    agent = CopilotAgent()
    resp = agent.process_chat(copilot_req)
    return resp.model_dump()


@app.post("/reporting/forecast-and-narrative", dependencies=[Depends(require_internal_service_key)])
async def reporting_forecast_endpoint(req: dict):
    """
    HTTP POST /reporting/forecast-and-narrative
    Phase 3 Task 3.11: Advanced Reporting, Forecasting & Executive Narrative.
    Generates time-series projections, anomaly driver explanations,
    trend interpretations, and safe executive summaries.
    """
    from app.reporting_forecasting.models import ReportingForecastRequest
    from app.reporting_forecasting.agent import ReportingForecastingAgent

    try:
        report_req = ReportingForecastRequest.model_validate(req)
    except Exception as parse_err:
        print(f"[REPORTING FORECAST ERROR] payload parse error: {parse_err}")
        raise HTTPException(
            status_code=status.HTTP_422_UNPROCESSABLE_ENTITY,
            detail=f"Invalid reporting forecast request payload: {str(parse_err)}"
        )

    agent = ReportingForecastingAgent()
    resp = agent.process_report(report_req)
    return resp.model_dump()


@app.post("/governance/inspect-input", dependencies=[Depends(require_internal_service_key)])
async def governance_inspect_input_endpoint(req: dict):
    """
    HTTP POST /governance/inspect-input
    Phase 3 Task 3.12: AI Governance and Production Controls.
    Inspects input prompts and payloads for prompt injections, PII, and sensitive credentials.
    """
    from app.governance.models import InputInspectionRequest
    from app.governance.safety_evaluator import AIGovernanceSafetyEvaluator

    try:
        inspect_req = InputInspectionRequest.model_validate(req)
    except Exception as parse_err:
        raise HTTPException(
            status_code=status.HTTP_422_UNPROCESSABLE_ENTITY,
            detail=f"Invalid input inspection payload: {str(parse_err)}"
        )

    evaluator = AIGovernanceSafetyEvaluator()
    res = evaluator.inspect_input(inspect_req)
    return res.model_dump()


@app.post("/governance/inspect-output", dependencies=[Depends(require_internal_service_key)])
async def governance_inspect_output_endpoint(req: dict):
    """
    HTTP POST /governance/inspect-output
    Phase 3 Task 3.12: AI Governance and Production Controls.
    Inspects generated outputs against source facts, flags ungrounded claims,
    evaluates hallucination risk, and enforces human approval for consequential actions.
    """
    from app.governance.models import OutputInspectionRequest
    from app.governance.safety_evaluator import AIGovernanceSafetyEvaluator

    try:
        inspect_req = OutputInspectionRequest.model_validate(req)
    except Exception as parse_err:
        raise HTTPException(
            status_code=status.HTTP_422_UNPROCESSABLE_ENTITY,
            detail=f"Invalid output inspection payload: {str(parse_err)}"
        )

    evaluator = AIGovernanceSafetyEvaluator()
    res = evaluator.inspect_output(inspect_req)
    return res.model_dump()


@app.post("/governance/evaluate-quality", dependencies=[Depends(require_internal_service_key)])
async def governance_evaluate_quality_endpoint(req: dict):
    """
    HTTP POST /governance/evaluate-quality
    Phase 3 Task 3.12: AI Governance and Production Controls.
    Evaluates test cases and benchmarks schema compliance, grounding, and release readiness.
    """
    from app.governance.models import QualityEvaluationRequest
    from app.governance.quality_evaluator import AIQualityEvaluator

    try:
        eval_req = QualityEvaluationRequest.model_validate(req)
    except Exception as parse_err:
        raise HTTPException(
            status_code=status.HTTP_422_UNPROCESSABLE_ENTITY,
            detail=f"Invalid quality evaluation payload: {str(parse_err)}"
        )

    evaluator = AIQualityEvaluator()
    res = evaluator.evaluate_dataset(eval_req)
    return res.model_dump()


@app.post("/leads/score-lead", dependencies=[Depends(require_internal_service_key)])
async def score_lead_endpoint(req: dict):
    """
    HTTP POST /leads/score-lead
    Migrated from backend/internal/jobs/lead_worker.go.
    Scores sales leads using ICP scoring rules, history, and LLM reasoning.
    """
    from app.agents.lead_scoring_agent import LeadScoringAgent, LeadScoringRequest
    try:
        scoring_req = LeadScoringRequest.model_validate(req)
    except Exception as parse_err:
        raise HTTPException(
            status_code=status.HTTP_422_UNPROCESSABLE_ENTITY,
            detail=f"Invalid lead scoring payload: {str(parse_err)}"
        )
    agent = LeadScoringAgent()
    res = agent.score_lead(scoring_req)
    return res.model_dump()


@app.post("/leads/classify-email", dependencies=[Depends(require_internal_service_key)])
async def classify_email_endpoint(req: dict):
    """
    HTTP POST /leads/classify-email
    Migrated from backend/internal/leads/bl.go (ClassifyInboundEmailAI).
    Classifies incoming sales emails into intent, sentiment, and logistics relevance.
    """
    from app.agents.email_classifier_agent import EmailClassifierAgent, EmailClassificationRequest
    try:
        classify_req = EmailClassificationRequest.model_validate(req)
    except Exception as parse_err:
        raise HTTPException(
            status_code=status.HTTP_422_UNPROCESSABLE_ENTITY,
            detail=f"Invalid email classification payload: {str(parse_err)}"
        )
    agent = EmailClassifierAgent()
    res = agent.classify(classify_req)
    return res.model_dump()


@app.post("/rfq/parse-shipment-request", dependencies=[Depends(require_internal_service_key)])
async def parse_shipment_request_endpoint(req: dict):
    """
    HTTP POST /rfq/parse-shipment-request
    Migrated from backend/internal/rfq/bl.go (ParseShipmentRequest).
    Extracts structured shipment parameters (origin, destination, incoterms, cargo) from unstructured text.
    """
    from app.agents.rfq_parser_agent import RFQParserAgent, ShipmentParseRequest
    try:
        parse_req = ShipmentParseRequest.model_validate(req)
    except Exception as parse_err:
        raise HTTPException(
            status_code=status.HTTP_422_UNPROCESSABLE_ENTITY,
            detail=f"Invalid RFQ parse payload: {str(parse_err)}"
        )
    agent = RFQParserAgent()
    res = agent.parse_request(parse_req)
    return res.model_dump()


@app.post("/ai/completion", dependencies=[Depends(require_internal_service_key)])
async def ai_completion_endpoint(req: dict):
    """
    HTTP POST /ai/completion
    General AI completion adapter replacing direct Gemini/OpenAI REST calls in Go.
    """
    from app.agents.llm_utils import execute_llm_text, execute_llm_json
    from app.config.runtime_config import get_runtime_config
    prompt = req.get("prompt", "")
    json_mode = req.get("json_mode", False)
    correlation_id = req.get("correlation_id", "")
    exec_ctx = {"request_id": correlation_id}
    cfg = get_runtime_config()

    if json_mode:
        data = execute_llm_json(prompt, exec_ctx=exec_ctx)
        content = json.dumps(data) if data else "{}"
    else:
        content = execute_llm_text(prompt, exec_ctx=exec_ctx)

    return {
        "content": content,
        "model": cfg.primary_model if hasattr(cfg, "primary_model") else "gemini-1.5-flash",
        "confidence": 0.95,
        "correlation_id": correlation_id,
        "status": "success",
    }


@app.post("/predictions/generate", dependencies=[Depends(require_internal_service_key)])
async def generate_prediction_endpoint(req: dict):
    """
    Phase 4: Predictive Intelligence & Decision Support Foundation
    Generates source-grounded predictions across shipments, invoices, customers, RFQs, and contracts.
    """
    from app.predictions import GeneratePredictionRequest, PredictionEngine
    parsed_req = GeneratePredictionRequest(**req)
    engine = PredictionEngine()
    result = engine.generate_prediction(parsed_req)
    return result.model_dump()


async def run_email_parse_pipeline(org_id: int, entity_id: str, payload: dict):
    interaction_id = int(entity_id)
    lead_id = payload.get("lead_id")
    from_email = payload.get("from", "")
    subject = payload.get("subject", "")
    body = payload.get("body", "")
    callback_url = payload.get("callback_url", "")

    # Thread-awareness fields — set by Go's InboundEmailWebhook when a reply is detected.
    # is_reply=True means Go found a prior RFQ_REQUEST_INCOMPLETE on this thread_id.
    # prior_rfq_context is structured JSON of fields extracted in previous turns.
    # The agent uses this to fill gaps without re-reading raw email text.
    is_reply = payload.get("is_reply", False)
    prior_rfq_context = payload.get("prior_rfq_context") or {}
    thread_id = payload.get("thread_id")
    parent_interaction_id = payload.get("parent_interaction_id")

    print(f"[AI Sidecar Sales] Starting email parse pipeline for interaction #{interaction_id} (is_reply={is_reply})")
    
    initial_state = {
        "messages": [],
        "org_id": org_id,
        "interaction_id": interaction_id,
        "lead_id": lead_id,
        "from_email": from_email,
        "email_subject": subject,
        "email_body": body,
        "callback_url": callback_url,
        # Thread-awareness fields
        "thread_id": thread_id,
        "is_reply": bool(is_reply),
        "parent_interaction_id": parent_interaction_id,
        "prior_rfq_context": prior_rfq_context if prior_rfq_context else None,
        # Cargo/RFQ fields — start as None; merge_context_node fills from prior if is_reply
        "company_domain": None,
        "company_enrichment": None,
        "lead_name": None,
        "origin_port": None,
        "destination_port": None,
        "incoterms": None,
        "cargo_description": None,
        "cargo_weight": None,
        "cargo_volume": None,
        "target_date": None,
        "intent": "QUESTION",
        "sentiment": "NEUTRAL",
        "confidence_score": 0,
        "linked_rfq_id": None,
        "ai_summary": "",
        "drafted_reply": None,
        "error_message": None
    }
    
    config = {
        "configurable": {"thread_id": f"sales-{interaction_id}", "org_id": org_id},
        "metadata": {
            "org_id": str(org_id),
            "task_type": "EMAIL_PARSE",
        }
    }
    
    try:
        from app.graphs.sales_graph import sales_graph
        sales_graph.invoke(initial_state, config=config)
        print(f"[AI Sidecar Sales] Sales graph execution finished for interaction #{interaction_id}")
    except Exception as e:
        import traceback
        traceback.print_exc()
        print(f"[AI Sidecar Sales] Sales graph execution failed: {e}")

# Instantiate the global QueueWorker using dependency injection.
#
# Simple meaning:
#   This worker will check the 'ai_processing_tasks' table in PostgreSQL
#   every 2 seconds for new PROCESS or RESUME tasks, run them asynchronously,
#   and update their status.
async def run_operations_pipeline(org_id: int, entity_id: str, payload: dict):
    """
    Handler for CARRIER_UPDATE_PARSE tasks.
    Invokes the LangGraph OperationsAgent to parse carrier tracking updates,
    update milestones, detect exceptions, and callback to Go.
    """
    print(f"[AI Sidecar Ops] Starting operations pipeline for entity {entity_id}, org {org_id}")

    shipment_id = int(entity_id) if entity_id and entity_id.isdigit() else None

    initial_state = {
        "org_id": org_id,
        "entity_id": entity_id or "",
        "callback_url": payload.get("callback_url", "http://localhost:8080/internal/operations/callback"),
        # Raw carrier event data
        "event_id": payload.get("event_id", ""),
        "carrier_scac": payload.get("carrier_scac", ""),
        "booking_number": payload.get("booking_number", ""),
        "container_number": payload.get("container_number", ""),
        "vessel_name": payload.get("vessel_name", ""),
        "voyage_number": payload.get("voyage_number", ""),
        "milestone_code": payload.get("milestone_code", ""),
        "event_time": payload.get("event_time", ""),
        "location": payload.get("location", ""),
        "raw_description": payload.get("description", ""),
        # Identification — if Go already resolved shipment, pass it
        "shipment_id": shipment_id,
        "shipment_data": None,
        "identification_confident": shipment_id is not None,
        # Outputs (initialized to empty)
        "detected_milestones": [],
        "detected_exceptions": [],
        "has_critical_exception": False,
        "requires_human_review": False,
        "ai_summary": "",
        "error_message": None,
    }

    config = {
        "configurable": {"thread_id": f"ops-{entity_id}-{payload.get('event_id', 'unknown')}", "org_id": org_id},
        "metadata": {
            "org_id": str(org_id),
            "task_type": "CARRIER_UPDATE_PARSE",
        }
    }

    try:
        from app.graphs.operations_graph import operations_graph
        operations_graph.invoke(initial_state, config=config)
        print(f"[AI Sidecar Ops] Operations graph execution finished for entity #{entity_id}")
    except Exception as e:
        import traceback
        traceback.print_exc()
        print(f"[AI Sidecar Ops] Operations graph execution failed: {e}")


async def run_compliance_pipeline(org_id: int, entity_id: str, payload: Dict[str, Any]):
    print(f"[AI Sidecar Compliance] Starting ComplianceAgent pipeline for entity {entity_id}...")
    
    # 3. Validate shipment_id and doc_id before starting the graph.
    try:
        shipment_id = int(payload.get("shipment_id", 0))
    except (ValueError, TypeError):
        raise ValueError("Invalid or missing shipment_id parameter in payload")

    doc_id = payload.get("doc_id", "")
    if not doc_id:
        raise ValueError("doc_id is required in task payload")

    initial_state = {
        "org_id": org_id,
        "shipment_id": shipment_id,
        "doc_id": doc_id,
        "doc_type": payload.get("doc_type", ""),
        "s3_key": payload.get("s3_key", ""),
        "file_name": payload.get("file_name", ""),
        "callback_url": payload.get("callback_url", "http://localhost:8080/internal/compliance/callback"),
        "raw_ocr_text": "",
        "extracted_data": {},
        "discrepancies": [],
        "doc_status": "PENDING"
    }

    config = {
        "configurable": {"thread_id": f"comp-{entity_id}-{doc_id}"},
        "metadata": {
            "org_id": str(org_id),
            "task_type": "DOC_VERIFY",
        }
    }

    try:
        from app.graphs.compliance_graph import compliance_graph
        # 2. Prevent blocking async worker loop by using asyncio.to_thread
        await asyncio.to_thread(compliance_graph.invoke, initial_state, config=config)
        print(f"[AI Sidecar Compliance] Compliance graph execution completed for doc {entity_id}")
    except Exception as e:
        import traceback
        traceback.print_exc()
        print(f"[AI Sidecar Compliance] Compliance graph execution failed: {e}")
        raise e


async def run_finance_pipeline(org_id: int, entity_id: str, payload: Dict[str, Any]):
    print(f"[AI Sidecar Finance] Starting FinanceAgent pipeline for entity {entity_id}...")

    try:
        shipment_id = int(payload.get("shipment_id", 0))
    except (ValueError, TypeError):
        raise ValueError("Invalid or missing shipment_id parameter in payload")

    invoice_id = payload.get("invoice_id", "")
    if not invoice_id:
        raise ValueError("invoice_id is required in task payload")

    initial_state = {
        "org_id": org_id,
        "shipment_id": shipment_id,
        "invoice_id": invoice_id,
        "invoice_number": payload.get("invoice_number", ""),
        "vendor_name": payload.get("vendor_name", ""),
        "s3_key": payload.get("s3_key", ""),
        "file_name": payload.get("file_name", ""),
        "callback_url": payload.get("callback_url", "http://localhost:8080/internal/finance/callback"),
        "raw_ocr_text": "",
        "extracted_items": [],
        "extracted_total": 0.0,
        "extracted_currency": "USD",
        "extracted_vendor": "",
        "contracted_rates": [],
        "discrepancies": [],
        "invoice_status": "PENDING_RECONCILIATION",
        "ai_summary": "",
    }

    config = {
        "configurable": {"thread_id": f"fin-{entity_id}-{invoice_id}"},
        "metadata": {
            "org_id": str(org_id),
            "task_type": "BILL_RECONCILE",
        }
    }

    try:
        from app.graphs.finance_graph import finance_graph
        await asyncio.to_thread(finance_graph.invoke, initial_state, config=config)
        print(f"[AI Sidecar Finance] Finance graph execution completed for invoice {invoice_id}")
    except Exception as e:
        import traceback
        traceback.print_exc()
        print(f"[AI Sidecar Finance] Finance graph execution failed: {e}")
        raise e


# ── Phase 5: Controlled Autonomous Operations Endpoints ──────────────────────
@app.post("/autonomy/plan/generate", dependencies=[Depends(require_internal_service_key)])
async def autonomy_plan_generate_endpoint(req: dict):
    from app.autonomy.models import PlanGenerationRequest
    from app.autonomy.planner import AutonomousPlannerAgent

    try:
        plan_req = PlanGenerationRequest.model_validate(req)
    except Exception as parse_err:
        raise HTTPException(
            status_code=status.HTTP_422_UNPROCESSABLE_ENTITY,
            detail=f"Invalid plan generation request: {str(parse_err)}"
        )

    agent = AutonomousPlannerAgent()
    try:
        resp = agent.generate_plan(plan_req)
        return resp.model_dump()
    except ValueError as val_err:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail=str(val_err)
        )


@app.post("/autonomy/plan/evaluate", dependencies=[Depends(require_internal_service_key)])
async def autonomy_plan_evaluate_endpoint(req: dict):
    from app.autonomy.models import PlanEvaluationRequest
    from app.autonomy.planner import AutonomousPlannerAgent

    try:
        eval_req = PlanEvaluationRequest.model_validate(req)
    except Exception as parse_err:
        raise HTTPException(
            status_code=status.HTTP_422_UNPROCESSABLE_ENTITY,
            detail=f"Invalid plan evaluation request: {str(parse_err)}"
        )

    agent = AutonomousPlannerAgent()
    resp = agent.evaluate_plan_policy(eval_req)
    return resp.model_dump()


@app.post("/autonomy/plan/replan", dependencies=[Depends(require_internal_service_key)])
async def autonomy_plan_replan_endpoint(req: dict):
    from app.autonomy.models import PlanReplanRequest
    from app.autonomy.planner import AutonomousPlannerAgent

    try:
        replan_req = PlanReplanRequest.model_validate(req)
    except Exception as parse_err:
        raise HTTPException(
            status_code=status.HTTP_422_UNPROCESSABLE_ENTITY,
            detail=f"Invalid plan replan request: {str(parse_err)}"
        )

    agent = AutonomousPlannerAgent()
    resp = agent.replan(replan_req)
    return resp.model_dump()


@app.post("/autonomy/plan/validate-graph", dependencies=[Depends(require_internal_service_key)])
async def autonomy_plan_validate_graph_endpoint(req: dict):
    from app.autonomy.models import MultiStepPlanValidationRequest
    from app.autonomy.planner import AutonomousPlannerAgent

    # Defensively normalize empty dictionaries for optional model fields
    steps = req.get("steps")
    if isinstance(steps, list):
        for s in steps:
            if isinstance(s, dict):
                for k in ["condition_predicate", "verification_criteria", "fallback_action", "compensation_action", "retry_policy"]:
                    if k in s and isinstance(s[k], dict) and not s[k]:
                        s[k] = None

    try:
        val_req = MultiStepPlanValidationRequest.model_validate(req)
    except Exception as parse_err:
        raise HTTPException(
            status_code=status.HTTP_422_UNPROCESSABLE_ENTITY,
            detail=f"Invalid plan validation request: {str(parse_err)}"
        )

    agent = AutonomousPlannerAgent()
    resp = agent.validate_plan_graph(val_req)
    return resp.model_dump()


@app.post("/autonomy/plan/cross-module", dependencies=[Depends(require_internal_service_key)])
async def autonomy_plan_cross_module_endpoint(req: dict):
    from app.autonomy.models import CrossModulePlanningRequest
    from app.autonomy.planner import AutonomousPlannerAgent

    try:
        cm_req = CrossModulePlanningRequest.model_validate(req)
    except Exception as parse_err:
        raise HTTPException(
            status_code=status.HTTP_422_UNPROCESSABLE_ENTITY,
            detail=f"Invalid cross-module plan request: {str(parse_err)}"
        )

    agent = AutonomousPlannerAgent()
    resp = agent.generate_cross_module_plan(cm_req)
    return resp.model_dump()


@app.post("/autonomy/shipments/evaluate-event", dependencies=[Depends(require_internal_service_key)])
async def autonomy_shipment_evaluate_event_endpoint(req: dict):
    from app.autonomy.models import ShipmentEventEvaluationRequest
    from app.autonomy.shipment_agent import AdaptiveShipmentAgent

    try:
        eval_req = ShipmentEventEvaluationRequest.model_validate(req)
    except Exception as parse_err:
        raise HTTPException(
            status_code=status.HTTP_422_UNPROCESSABLE_ENTITY,
            detail=f"Invalid shipment event evaluation request: {str(parse_err)}"
        )

    agent = AdaptiveShipmentAgent()
    resp = agent.evaluate_event(eval_req)
    return resp.model_dump()


@app.post("/autonomy/shipments/adaptive-plan", dependencies=[Depends(require_internal_service_key)])
async def autonomy_shipment_adaptive_plan_endpoint(req: dict):
    from app.autonomy.models import PlanGenerationRequest
    from app.autonomy.shipment_agent import AdaptiveShipmentAgent

    try:
        plan_req = PlanGenerationRequest.model_validate(req)
    except Exception as parse_err:
        raise HTTPException(
            status_code=status.HTTP_422_UNPROCESSABLE_ENTITY,
            detail=f"Invalid adaptive shipment plan request: {str(parse_err)}"
        )

    agent = AdaptiveShipmentAgent()
    resp = agent.generate_adaptive_plan(plan_req)
    return resp.model_dump()


@app.post("/autonomy/customers/evaluate-followup", dependencies=[Depends(require_internal_service_key)])
async def autonomy_customer_evaluate_followup_endpoint(req: dict):
    """
    HTTP POST /autonomy/customers/evaluate-followup
    Phase 5 Task 5.4: Autonomous Customer Follow-Up Evaluation.
    Grounded decision logic, preference enforcement, and fact/prediction separation.
    """
    from app.autonomy.models import CustomerFollowupEvaluationRequest
    from app.autonomy.customer_agent import evaluate_customer_followup

    try:
        eval_req = CustomerFollowupEvaluationRequest.model_validate(req)
    except Exception as parse_err:
        raise HTTPException(
            status_code=status.HTTP_422_UNPROCESSABLE_ENTITY,
            detail=f"Invalid customer follow-up evaluation request: {str(parse_err)}"
        )

    resp = evaluate_customer_followup(eval_req)
    return resp.model_dump()


@app.post("/autonomy/customers/classify-response", dependencies=[Depends(require_internal_service_key)])
async def autonomy_customer_classify_response_endpoint(req: dict):
    """
    HTTP POST /autonomy/customers/classify-response
    Phase 5 Task 5.4: Customer Response Classification.
    Classifies incoming customer text into structured operational categories.
    """
    from app.autonomy.models import ClassifyCustomerResponseRequest
    from app.autonomy.customer_agent import classify_customer_response

    try:
        class_req = ClassifyCustomerResponseRequest.model_validate(req)
    except Exception as parse_err:
        raise HTTPException(
            status_code=status.HTTP_422_UNPROCESSABLE_ENTITY,
            detail=f"Invalid customer response classification request: {str(parse_err)}"
        )

    resp = classify_customer_response(class_req)
    return resp.model_dump()


@app.post("/autonomy/pricing/evaluate-rfq", dependencies=[Depends(require_internal_service_key)])
async def autonomy_pricing_evaluate_rfq_endpoint(req: dict):
    """
    HTTP POST /autonomy/pricing/evaluate-rfq
    Phase 5 Task 5.5: Intelligent RFQ and Pricing Optimization.
    Analyzes RFQ parameters, rate basis, and constraints to generate pricing strategies.
    """
    from app.autonomy.models import RfqPricingEvaluationRequest
    from app.autonomy.pricing_agent import evaluate_rfq_pricing

    try:
        eval_req = RfqPricingEvaluationRequest.model_validate(req)
    except Exception as parse_err:
        raise HTTPException(
            status_code=status.HTTP_422_UNPROCESSABLE_ENTITY,
            detail=f"Invalid RFQ pricing evaluation request: {str(parse_err)}"
        )

    resp = evaluate_rfq_pricing(eval_req.context)
    return resp.model_dump()


@app.post("/autonomy/pricing/replan", dependencies=[Depends(require_internal_service_key)])
async def autonomy_pricing_replan_endpoint(req: dict):
    """
    HTTP POST /autonomy/pricing/replan
    Phase 5 Task 5.5: Pricing Replanning.
    Re-evaluates quotation strategies when carrier rates or operational constraints change.
    """
    from app.autonomy.models import PricingReplanningRequest
    from app.autonomy.pricing_agent import replan_rfq_pricing

    try:
        replan_req = PricingReplanningRequest.model_validate(req)
    except Exception as parse_err:
        raise HTTPException(
            status_code=status.HTTP_422_UNPROCESSABLE_ENTITY,
            detail=f"Invalid pricing replanning request: {str(parse_err)}"
        )

    resp = replan_rfq_pricing(replan_req)
    return resp.model_dump()


# ==============================================================================
# Phase 5 Task 5.6: Adaptive Finance and Collections Endpoints
# ==============================================================================

@app.post("/autonomy/finance/evaluate-collection", dependencies=[Depends(require_internal_service_key)])
async def autonomy_finance_evaluate_collection_endpoint(req: dict):
    """
    HTTP POST /autonomy/finance/evaluate-collection
    Phase 5 Task 5.6: Evaluates an invoice context and generates ranked collection strategies,
    fact/prediction segregation, risk/priority scoring, and a 7-step autonomous execution plan.
    """
    from app.autonomy.models import FinanceCollectionEvaluationRequest
    from app.autonomy.finance_agent import evaluate_finance_collection

    try:
        eval_req = FinanceCollectionEvaluationRequest.model_validate(req)
    except Exception as parse_err:
        raise HTTPException(
            status_code=status.HTTP_422_UNPROCESSABLE_ENTITY,
            detail=f"Invalid finance collection evaluation request: {str(parse_err)}"
        )

    resp = evaluate_finance_collection(eval_req)
    return resp.model_dump()


@app.post("/autonomy/finance/replan-collection", dependencies=[Depends(require_internal_service_key)])
async def autonomy_finance_replan_collection_endpoint(req: dict):
    """
    HTTP POST /autonomy/finance/replan-collection
    Phase 5 Task 5.6: Re-evaluates collection strategies upon payment events, customer responses,
    disputes, or material financial changes.
    """
    from app.autonomy.models import FinanceCollectionReplanningRequest
    from app.autonomy.finance_agent import replan_finance_collection

    try:
        replan_req = FinanceCollectionReplanningRequest.model_validate(req)
    except Exception as parse_err:
        raise HTTPException(
            status_code=status.HTTP_422_UNPROCESSABLE_ENTITY,
            detail=f"Invalid finance collection replanning request: {str(parse_err)}"
        )

    resp = replan_finance_collection(replan_req)
    return resp.model_dump()


# ==============================================================================
# PHASE 5 TASK 5.7: CONTRACT AND COMPLIANCE MONITORING ENDPOINTS
# ==============================================================================

@app.post("/autonomy/compliance/evaluate-contract", dependencies=[Depends(require_internal_service_key)])
async def autonomy_compliance_evaluate_contract_endpoint(req: dict):
    """
    HTTP POST /autonomy/compliance/evaluate-contract
    Phase 5 Task 5.7: Evaluates a contract & compliance context, detects hard/soft deviations,
    monitors expirations, separates facts vs predictions, and synthesizes candidate remediation strategies.
    """
    from app.autonomy.models import ContractComplianceEvaluationRequest
    from app.autonomy.compliance_agent import evaluate_contract_compliance

    try:
        eval_req = ContractComplianceEvaluationRequest.model_validate(req)
    except Exception as parse_err:
        raise HTTPException(
            status_code=status.HTTP_422_UNPROCESSABLE_ENTITY,
            detail=f"Invalid contract compliance evaluation request: {str(parse_err)}"
        )

    resp = evaluate_contract_compliance(eval_req)
    return resp.model_dump()


@app.post("/autonomy/compliance/replan-contract", dependencies=[Depends(require_internal_service_key)])
async def autonomy_compliance_replan_contract_endpoint(req: dict):
    """
    HTTP POST /autonomy/compliance/replan-contract
    Phase 5 Task 5.7: Re-evaluates compliance plans upon document uploads, verification,
    route changes, rate amendments, or contract expiration events.
    """
    from app.autonomy.models import ContractComplianceReplanningRequest
    from app.autonomy.compliance_agent import replan_contract_compliance

    try:
        replan_req = ContractComplianceReplanningRequest.model_validate(req)
    except Exception as parse_err:
        raise HTTPException(
            status_code=status.HTTP_422_UNPROCESSABLE_ENTITY,
            detail=f"Invalid contract compliance replanning request: {str(parse_err)}"
        )

    resp = replan_contract_compliance(replan_req)
    return resp.model_dump()


@app.post("/autonomy/exceptions/evaluate-exception", dependencies=[Depends(require_internal_service_key)])
async def autonomy_exceptions_evaluate_exception_endpoint(req: dict):
    """
    HTTP POST /autonomy/exceptions/evaluate-exception
    Phase 5 Task 5.8: Evaluates an operational exception context and synthesizes:
    - Root-cause reasoning vs symptom segregation
    - Multi-tier impact analysis (confirmed, predicted, possible)
    - 5 candidate recovery strategies with feasibility & risk scoring
    - Selected strategy and 7-step sequential recovery plan
    - Controlled waiting states, verification criteria, and approval gating.
    """
    from app.autonomy.models import ExceptionEvaluationRequest
    from app.autonomy.exception_agent import evaluate_exception_resolution

    try:
        eval_req = ExceptionEvaluationRequest.model_validate(req)
    except Exception as parse_err:
        raise HTTPException(
            status_code=status.HTTP_422_UNPROCESSABLE_ENTITY,
            detail=f"Invalid exception evaluation request: {str(parse_err)}"
        )

    resp = evaluate_exception_resolution(eval_req)
    return resp.model_dump()


@app.post("/autonomy/exceptions/replan-exception", dependencies=[Depends(require_internal_service_key)])
async def autonomy_exceptions_replan_exception_endpoint(req: dict):
    """
    HTTP POST /autonomy/exceptions/replan-exception
    Phase 5 Task 5.8: Re-evaluates exception recovery plans upon carrier updates,
    document submissions, customer confirmations, or verification failures.
    """
    from app.autonomy.models import ExceptionReplanningRequest
    from app.autonomy.exception_agent import replan_exception_resolution

    try:
        replan_req = ExceptionReplanningRequest.model_validate(req)
    except Exception as parse_err:
        raise HTTPException(
            status_code=status.HTTP_422_UNPROCESSABLE_ENTITY,
            detail=f"Invalid exception replanning request: {str(parse_err)}"
        )

    resp = replan_exception_resolution(replan_req)
    return resp.model_dump()


# ==============================================================================
# PHASE 5 TASK 5.10: CONTINUOUS MONITORING AND REPLANNING ENDPOINTS
# ==============================================================================

@app.post("/autonomy/monitoring/evaluate-state-change", dependencies=[Depends(require_internal_service_key)])
async def autonomy_monitoring_evaluate_state_change_endpoint(req: dict):
    """
    HTTP POST /autonomy/monitoring/evaluate-state-change
    Phase 5 Task 5.10: Continuously assesses incoming business events and state changes
    against active plans, evaluating plan health, invalidated assumptions, affected steps,
    and recommending safe control actions (CONTINUE, PAUSE, REPLAN, ESCALATE, STOP).
    """
    from app.autonomy.models import StateChangeEvaluationRequest
    from app.autonomy.monitoring_agent import evaluate_state_change

    try:
        eval_req = StateChangeEvaluationRequest.model_validate(req)
    except Exception as parse_err:
        raise HTTPException(
            status_code=status.HTTP_422_UNPROCESSABLE_ENTITY,
            detail=f"Invalid state change evaluation request: {str(parse_err)}"
        )

    resp = evaluate_state_change(eval_req)
    return resp.model_dump()


@app.post("/autonomy/monitoring/replan", dependencies=[Depends(require_internal_service_key)])
async def autonomy_monitoring_replan_endpoint(req: dict):
    """
    HTTP POST /autonomy/monitoring/replan
    Phase 5 Task 5.10: Formulates adaptive plan revisions (e.g. V1 -> V2 -> V3)
    triggered by material operational changes while strictly preserving and protecting
    all successfully completed steps from redundant re-execution.
    """
    from app.autonomy.models import ContinuousReplanningRequest
    from app.autonomy.monitoring_agent import replan_continuous_workflow

    try:
        replan_req = ContinuousReplanningRequest.model_validate(req)
    except Exception as parse_err:
        raise HTTPException(
            status_code=status.HTTP_422_UNPROCESSABLE_ENTITY,
            detail=f"Invalid continuous replanning request: {str(parse_err)}"
        )

    resp = replan_continuous_workflow(replan_req)
    return resp.model_dump()


# =============================================================================
# Phase 5 Task 5.11: Human + AI Operating Model Endpoints
# =============================================================================

@app.post("/autonomy/human-ai/analyze-decision", tags=["Human + AI Operating Model"])
async def analyze_decision_point_endpoint(req: Dict[str, Any]):
    """
    Phase 5 Task 5.11: Human + AI Operating Model Decision Point Analysis.
    Evaluates business context facts, forecasts, confidence, data sufficiency,
    recommends action, formats alternatives, and attributes decision provenance.
    """
    from app.autonomy.models import HumanAIDecisionAnalysisRequest
    from app.autonomy.human_ai_agent import human_ai_agent

    try:
        decision_req = HumanAIDecisionAnalysisRequest.model_validate(req)
    except Exception as parse_err:
        raise HTTPException(
            status_code=status.HTTP_422_UNPROCESSABLE_ENTITY,
            detail=f"Invalid human-ai decision request: {str(parse_err)}"
        )

    resp = human_ai_agent.analyze_decision_point(decision_req)
    return resp.model_dump()


@app.post("/autonomy/human-ai/analyze-feedback", tags=["Human + AI Operating Model"])
async def analyze_human_feedback_endpoint(req: Dict[str, Any]):
    """
    Phase 5 Task 5.11: Human + AI Operating Model Feedback Analysis.
    Interprets human decisions, overrides, modifications, stops, and escalations.
    Synthesizes structured memory items without storing chain-of-thought.
    """
    from app.autonomy.models import HumanFeedbackAnalysisRequest
    from app.autonomy.human_ai_agent import human_ai_agent

    try:
        feedback_req = HumanFeedbackAnalysisRequest.model_validate(req)
    except Exception as parse_err:
        raise HTTPException(
            status_code=status.HTTP_422_UNPROCESSABLE_ENTITY,
            detail=f"Invalid human feedback request: {str(parse_err)}"
        )

    resp = human_ai_agent.analyze_human_feedback(feedback_req)
    return resp.model_dump()


@app.post("/autonomy/command-center/prioritize", tags=["Autonomous Operations Command Center"])
async def command_center_prioritize_endpoint(req: Dict[str, Any]):
    """
    Phase 5 Task 5.12: Autonomous Operations Command Center Prioritization.
    Evaluates cross-domain operational items, calculates multi-factor priority scores,
    and formats structured items with facts vs predictions separation.
    """
    from app.autonomy.models import CommandCenterPrioritizeRequest
    from app.autonomy.command_center_agent import command_center_agent

    try:
        prioritize_req = CommandCenterPrioritizeRequest.model_validate(req)
    except Exception as parse_err:
        raise HTTPException(
            status_code=status.HTTP_422_UNPROCESSABLE_ENTITY,
            detail=f"Invalid command center prioritize request: {str(parse_err)}"
        )

    resp = command_center_agent.prioritize_operations(prioritize_req)
    return resp.model_dump()


# -----------------------------------------------------------------------------
# Phase 5 Task 5.13: Agent Memory and Learning from Outcomes Endpoints
# -----------------------------------------------------------------------------

@app.post("/autonomy/memory/evaluate-outcome", tags=["Agent Memory & Learning"])
async def memory_evaluate_outcome_endpoint(req: Dict[str, Any]):
    """
    Phase 5 Task 5.13: Evaluates an operational execution outcome against expectation.
    Assesses success/failure, determines failure categories, and formulates safe memory candidates.
    """
    from app.autonomy.models import OutcomeEvaluationRequest
    from app.autonomy.memory_learning_agent import memory_learning_agent

    try:
        eval_req = OutcomeEvaluationRequest.model_validate(req)
    except Exception as parse_err:
        raise HTTPException(
            status_code=status.HTTP_422_UNPROCESSABLE_ENTITY,
            detail=f"Invalid outcome evaluation request: {str(parse_err)}"
        )

    resp = memory_learning_agent.evaluate_outcome(eval_req)
    return resp.model_dump()


@app.post("/autonomy/memory/retrieve", tags=["Agent Memory & Learning"])
async def memory_retrieve_endpoint(req: Dict[str, Any]):
    """
    Phase 5 Task 5.13: Retrieves and ranks relevant contextual operational memories.
    Applies relevance scoring, recency decay, prompt-injection defense, and conflict detection.
    """
    from app.autonomy.models import MemoryRetrievalRequest
    from app.autonomy.memory_learning_agent import memory_learning_agent

    try:
        ret_req = MemoryRetrievalRequest.model_validate(req)
    except Exception as parse_err:
        raise HTTPException(
            status_code=status.HTTP_422_UNPROCESSABLE_ENTITY,
            detail=f"Invalid memory retrieval request: {str(parse_err)}"
        )

    resp = memory_learning_agent.retrieve_relevant_memory(ret_req)
    return resp.model_dump()


@app.post("/autonomy/memory/detect-patterns", tags=["Agent Memory & Learning"])
async def memory_detect_patterns_endpoint(req: Dict[str, Any]):
    """
    Phase 5 Task 5.13: Analyzes historical outcomes and memories to detect recurring patterns.
    """
    from app.autonomy.models import PatternDetectionRequest
    from app.autonomy.memory_learning_agent import memory_learning_agent

    try:
        pat_req = PatternDetectionRequest.model_validate(req)
    except Exception as parse_err:
        raise HTTPException(
            status_code=status.HTTP_422_UNPROCESSABLE_ENTITY,
            detail=f"Invalid pattern detection request: {str(parse_err)}"
        )

    resp = memory_learning_agent.detect_patterns(pat_req)
    return resp.model_dump()


@app.post("/autonomy/memory/resolve-conflicts", tags=["Agent Memory & Learning"])
async def memory_resolve_conflicts_endpoint(req: Dict[str, Any]):
    """
    Phase 5 Task 5.13: Evaluates contradictions between new authoritative observations and prior memories.
    """
    from app.autonomy.models import MemoryConflictRequest
    from app.autonomy.memory_learning_agent import memory_learning_agent

    try:
        conf_req = MemoryConflictRequest.model_validate(req)
    except Exception as parse_err:
        raise HTTPException(
            status_code=status.HTTP_422_UNPROCESSABLE_ENTITY,
            detail=f"Invalid memory conflict request: {str(parse_err)}"
        )

    resp = memory_learning_agent.resolve_memory_conflicts(conf_req)
    return resp.model_dump()


# ==============================================================================
# Phase 5 Task 5.14: Governance for Controlled Autonomy Routes
# ==============================================================================

@app.post("/autonomy/governance/evaluate-context", tags=["Controlled Autonomy Governance"])
async def governance_evaluate_context_endpoint(req: Dict[str, Any]):
    """
    Phase 5 Task 5.14: Evaluates operational context, determines risk score, data sufficiency,
    confidence class, and recommended autonomy tier, and neutralizes prompt injections.
    """
    from app.autonomy.models import GovernanceContextEvaluationRequest
    from app.autonomy.governance_agent import governance_agent

    try:
        gov_req = GovernanceContextEvaluationRequest.model_validate(req)
    except Exception as parse_err:
        raise HTTPException(
            status_code=status.HTTP_422_UNPROCESSABLE_ENTITY,
            detail=f"Invalid governance evaluation request: {str(parse_err)}"
        )

    resp = governance_agent.evaluate_context(gov_req)
    return resp.model_dump()


@app.post("/autonomy/governance/preview-plan", tags=["Controlled Autonomy Governance"])
async def governance_preview_plan_endpoint(req: Dict[str, Any]):
    """
    Phase 5 Task 5.14: Simulates dry-run of a multi-step plan, determining affected entities,
    max risk class, estimated financial exposure, and approval gates without mutating state.
    """
    from app.autonomy.models import GovernancePlanPreviewRequest
    from app.autonomy.governance_agent import governance_agent

    try:
        prev_req = GovernancePlanPreviewRequest.model_validate(req)
    except Exception as parse_err:
        raise HTTPException(
            status_code=status.HTTP_422_UNPROCESSABLE_ENTITY,
            detail=f"Invalid governance plan preview request: {str(parse_err)}"
        )

    resp = governance_agent.preview_plan(prev_req)
    return resp.model_dump()


# =====================================================================
# Phase 6.1: Multi-Agent Workforce Foundation Endpoints
# =====================================================================

@app.get("/api/v1/workforce/agents", tags=["Workforce Foundation"])
async def workforce_list_agents():
    """Returns registered AI workforce agents and their structured capabilities."""
    from app.workforce.registry import workforce_registry
    return [a.model_dump() for a in workforce_registry.list_agents()]


@app.get("/api/v1/workforce/agents/{agent_id}", tags=["Workforce Foundation"])
async def workforce_get_agent(agent_id: str):
    """Returns metadata and capabilities for a specific workforce agent."""
    from app.workforce.registry import workforce_registry
    agent = workforce_registry.get_agent(agent_id)
    if not agent:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f"Agent '{agent_id}' not found in workforce registry."
        )
    return agent.get_metadata().model_dump()


@app.post("/api/v1/workforce/execute-task", tags=["Workforce Foundation"])
async def workforce_execute_task(req: Dict[str, Any]):
    """
    Executes a workforce task using the assigned agent and segregated context.
    Returns structured results, proposed delegations, proposed actions, and messages.
    CRITICAL: Does NOT mutate business records directly!
    """
    from app.workforce.models import WorkforceTaskContract
    from app.workforce.coordinator import workforce_coordinator

    try:
        task_contract = WorkforceTaskContract.model_validate(req)
    except Exception as parse_err:
        raise HTTPException(
            status_code=status.HTTP_422_UNPROCESSABLE_ENTITY,
            detail=f"Invalid workforce task contract: {str(parse_err)}"
        )

    result = workforce_coordinator.execute_task(task_contract)
    return result.model_dump()


# Instantiate the global QueueWorker using dependency injection.


#
# Simple meaning:
#   This worker will check the 'ai_processing_tasks' table in PostgreSQL
#   every 2 seconds for new PROCESS or RESUME tasks, run them asynchronously,
#   and update their status.
task_handlers = {
    "PRICING_ANALYZE": run_pricing_pipeline,
    "PRICING_RESUME": run_pricing_resume_pipeline,
    "EMAIL_PARSE": run_email_parse_pipeline,
    "CARRIER_UPDATE_PARSE": run_operations_pipeline,
    "DOC_VERIFY": run_compliance_pipeline,
    "BILL_RECONCILE": run_finance_pipeline,
}

worker = QueueWorker(
    run_process_fn=run_langgraph_pipeline,
    run_resume_fn=run_resume_pipeline,
    processing_request_cls=ProcessingRequest,
    resume_request_cls=ResumeRequest,
    task_handlers=task_handlers
)

@app.on_event("startup")
async def startup_event():
    from app.persistence.checkpointer import validate_checkpointer_config, get_checkpointer
    validate_checkpointer_config()
    checkpointer = get_checkpointer()
    if hasattr(checkpointer, "verify_storage"):
        checkpointer.verify_storage()
    print(f"[AI Sidecar] Active checkpointer: {checkpointer.__class__.__name__}")

    # Start the worker thread when the FastAPI server starts up
    await worker.start()

@app.on_event("shutdown")
async def shutdown_event():
    # Stop the worker loop cleanly when the FastAPI server stops
    await worker.stop()


if __name__ == "__main__":
    import uvicorn
    # Start the server on port 8090.
    uvicorn.run(app, host="0.0.0.0", port=8090)


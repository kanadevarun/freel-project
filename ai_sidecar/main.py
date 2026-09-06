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
from fastapi import FastAPI, BackgroundTasks, HTTPException, status
from pydantic import BaseModel

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

    config = {"configurable": {"thread_id": req.document_id}}
    
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
    print(f"[AI Sidecar][Correlation ID: {req.correlation_id or 'None'}] Resuming LangGraph pipeline for doc {req.document_id}...")
    config = {"configurable": {"thread_id": req.document_id}}

    # Load current state
    snapshot = contracts_graph.get_state(config)
    if not snapshot or not snapshot.values:
        print(f"[AI Sidecar] Thread ID {req.document_id} not found. Cannot resume.")
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
        "configurable": {"thread_id": f"rfq-{rfq_id}"},
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
            "ai_reasoning": f"Pricing graph execution crashed: {str(e)}"
        })
        return

    # Check state after invoke
    snapshot = pricing_graph.get_state(config)
    state_values = snapshot.values if snapshot else result

    is_anomaly = state_values.get("is_anomaly", False)
    overall_reasoning = state_values.get("overall_reasoning", "")
    
    if is_anomaly:
        print(f"[AI Sidecar Pricing] Anomaly detected for RFQ #{rfq_id}. Waiting for human review.")
        await send_callback(callback_url, {
            "rfq_id": rfq_id,
            "org_id": org_id,
            "status": "NEEDS_REVIEW",
            "correlation_id": correlation_id,
            "ai_reasoning": overall_reasoning
        })
    else:
        print(f"[AI Sidecar Pricing] Pricing successful for RFQ #{rfq_id}. Auto-saving quotes.")
        try:
            # Resume/continue graph run past the interrupt to trigger the save node
            pricing_graph.invoke(None, config=config)
            
            await send_callback(callback_url, {
                "rfq_id": rfq_id,
                "org_id": org_id,
                "status": "COMPLETED",
                "correlation_id": correlation_id,
                "ai_reasoning": overall_reasoning
            })
        except Exception as save_err:
            print(f"[AI Sidecar Pricing] Failed to save pricing quotes: {save_err}")
            await send_callback(callback_url, {
                "rfq_id": rfq_id,
                "org_id": org_id,
                "status": "FAILED",
                "correlation_id": correlation_id,
                "ai_reasoning": f"Failed to persist quotes: {str(save_err)}"
            })

async def run_pricing_resume_pipeline(org_id: int, entity_id: str, payload: dict):
    rfq_id = int(entity_id)
    correlation_id = payload.get("correlation_id", "")
    callback_url = payload.get("callback_url", "")
    
    print(f"[AI Sidecar Pricing] Resuming pricing analysis for RFQ #{rfq_id} (Correlation ID: {correlation_id})")
    
    config = {
        "configurable": {"thread_id": f"rfq-{rfq_id}"},
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
    """Sends JSON results POST request to Go backend callback endpoint."""
    token = os.getenv("INTERNAL_SERVICE_TOKEN", "internal-service-key-logisticshq")
    headers = {"X-LogisticsHQ-Service-Key": token}
    
    async with httpx.AsyncClient() as client:
        try:
            resp = await client.post(callback_url, json=payload, headers=headers, timeout=15.0)
            print(f"[AI Sidecar] Callback response: status={resp.status_code}, body={resp.text}")
        except Exception as e:
            print(f"[AI Sidecar] Failed to send callback request: {e}")

@app.post("/process", status_code=status.HTTP_202_ACCEPTED)
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

@app.post("/resume", status_code=status.HTTP_202_ACCEPTED)
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

class ExtractAgreementRequest(BaseModel):
    document_id: Optional[str] = None
    org_id: int
    raw_text: Optional[str] = ""
    file_name: Optional[str] = "contract_agreement.pdf"
    s3_key: Optional[str] = None

@app.post("/contracts/extract-agreement")
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
        "configurable": {"thread_id": f"sales-{interaction_id}"},
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
        "configurable": {"thread_id": f"ops-{entity_id}-{payload.get('event_id', 'unknown')}"},
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


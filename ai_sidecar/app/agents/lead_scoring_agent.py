from typing import List, Optional, Dict, Any
from pydantic import BaseModel, Field
from app.prompts.prompt_registry import get_prompt
from app.agents.llm_utils import execute_llm_json

class LeadScoringRequest(BaseModel):
    org_id: int
    user_id: Optional[int] = None
    lead_id: Optional[int] = None
    company_name: str
    industry: Optional[str] = ""
    location: Optional[str] = ""
    employee_count: Optional[int] = 0
    estimated_revenue: Optional[str] = ""
    monthly_shipping_volume: Optional[str] = ""
    top_suppliers: Optional[str] = ""
    is_exporter: Optional[bool] = False
    correlation_id: str

class LeadScoringResponse(BaseModel):
    score: int = Field(ge=0, le=100)
    research_report: str
    confidence: float = Field(default=0.85, ge=0.0, le=1.0)
    recommended_tier: str = "TIER_2"
    key_strengths: List[str] = []
    risk_factors: List[str] = []
    correlation_id: str

    @property
    def tier(self) -> str:
        return self.recommended_tier

    @property
    def reasoning(self) -> str:
        return self.research_report

class LeadScoringAgent:
    """
    Authoritative Python Agent for B2B Shipper Lead Scoring and Trade Intelligence Research.
    Migrated from Go to Python to preserve agentic reasoning in Python AI runtime.
    """

    def score_lead(self, req: LeadScoringRequest) -> LeadScoringResponse:
        vars_dict = {
            "CompanyName": req.company_name,
            "Industry": req.industry or "Commercial Freight / Trade",
            "Location": req.location or "Global",
            "EmployeeCount": req.employee_count or 50,
            "EstimatedRevenue": req.estimated_revenue or "$10M - $50M",
            "MonthlyShippingVolume": req.monthly_shipping_volume or "10-25 TEU",
            "TopSuppliers": req.top_suppliers or "Various International Suppliers",
            "IsExporter": req.is_exporter if req.is_exporter is not None else True,
        }

        prompt_str = get_prompt("leads.lead_scoring", variables=vars_dict)
        if not prompt_str:
            prompt_str = f"""Analyze this freight prospect: {req.company_name} ({req.industry}, {req.location}).
Volume: {req.monthly_shipping_volume}. Exporter: {req.is_exporter}.
Evaluate commercial logistics viability and assign a score (0-100).
Return JSON: {{"score": int, "research_report": str, "key_strengths": [], "risk_factors": []}}"""

        exec_ctx = {
            "org_id": req.org_id,
            "user_id": req.user_id,
            "correlation_id": req.correlation_id,
            "workflow_name": "leads",
            "request_id": req.correlation_id,
        }

        output = execute_llm_json(
            prompt=prompt_str,
            prompt_key="leads.lead_scoring",
            prompt_version="1.0.0",
            exec_ctx=exec_ctx,
        )

        score = 75
        research_report = f"Preliminary trade intelligence synthesized for {req.company_name}."
        strengths = ["Active commercial presence", "Identified freight volumes"]
        risks = []
        confidence = 0.88

        if output and isinstance(output, dict):
            if "score" in output:
                try:
                    score = int(output["score"])
                except Exception:
                    pass
            if "research_report" in output and output["research_report"]:
                research_report = str(output["research_report"])
            if "key_strengths" in output and isinstance(output["key_strengths"], list):
                strengths = [str(s) for s in output["key_strengths"]]
            if "risk_factors" in output and isinstance(output["risk_factors"], list):
                risks = [str(r) for r in output["risk_factors"]]
            if "confidence" in output:
                try:
                    confidence = float(output["confidence"])
                except Exception:
                    pass
        else:
            # Deterministic calculation fallback if mock / offline
            base_score = 65
            if req.is_exporter:
                base_score += 15
            if "teu" in str(req.monthly_shipping_volume).lower() or "container" in str(req.monthly_shipping_volume).lower():
                base_score += 10
            score = min(98, max(20, base_score))
            research_report = (
                f"Automated trade profile for {req.company_name} based in {req.location or 'unspecified location'}. "
                f"Operating in {req.industry or 'general freight'}. Monthly volume indicated as {req.monthly_shipping_volume}."
            )

        tier = "TIER_1" if score >= 80 else ("TIER_2" if score >= 50 else "TIER_3")

        return LeadScoringResponse(
            score=score,
            research_report=research_report,
            confidence=confidence,
            recommended_tier=tier,
            key_strengths=strengths,
            risk_factors=risks,
            correlation_id=req.correlation_id,
        )

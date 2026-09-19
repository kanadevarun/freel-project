"""
Customer Relationship Automation & Intelligent Follow-Up Package
Phase 3 Task 3.3 for LogisticsHQ
"""
from app.customer_relationship.models import (
    CustomerFollowupContext,
    CustomerPriorityResult,
    CustomerEvaluationResponse,
    FollowupRecommendationOutput,
    EvidenceFact,
    SafetyWarning,
    DraftMessageRequest,
    DraftMessageResponse,
)
from app.customer_relationship.agent import CustomerRelationshipAgent

__all__ = [
    "CustomerFollowupContext",
    "CustomerPriorityResult",
    "CustomerEvaluationResponse",
    "FollowupRecommendationOutput",
    "EvidenceFact",
    "SafetyWarning",
    "DraftMessageRequest",
    "DraftMessageResponse",
    "CustomerRelationshipAgent",
]

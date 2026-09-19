"""
LogisticsHQ AI Copilot Module (Phase 3 Task 3.10)
Provides context-aware conversational intelligence, record explanations,
risk assessments, draft synthesis, and controlled action proposals across every module.
"""

from .models import (
    CopilotContextPayload,
    CopilotMessage,
    CopilotSourceRef,
    CopilotActionProposal,
    CopilotChatRequest,
    CopilotChatResponse,
)
from .agent import CopilotAgent

__all__ = [
    "CopilotContextPayload",
    "CopilotMessage",
    "CopilotSourceRef",
    "CopilotActionProposal",
    "CopilotChatRequest",
    "CopilotChatResponse",
    "CopilotAgent",
]

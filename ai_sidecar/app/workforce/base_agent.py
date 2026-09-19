"""
LogisticsHQ Phase 6.1 — Base Workforce Agent Abstraction
"""

from abc import ABC, abstractmethod
from typing import List, Dict, Any, Optional
from app.workforce.models import (
    AgentMetadata,
    AgentCapability,
    WorkforceTaskContract,
    ContextReference,
    ContextItemType,
    AgentExecutionResult,
    DelegationRequest,
    HandoffContract,
    TaskStatus,
    TaskPriority,
)


class BaseWorkforceAgent(ABC):
    """
    Foundational contract for all specialized AI agents in the LogisticsHQ workforce.
    Enforces capability validation, context segregation, and structured output.
    """

    def __init__(self, metadata: AgentMetadata):
        self.metadata = metadata

    def get_metadata(self) -> AgentMetadata:
        return self.metadata

    def has_capability(self, capability: str) -> bool:
        return capability in self.metadata.capabilities

    def validate_capabilities(self, required_capabilities: List[str]) -> bool:
        """Verifies if the agent possesses all required capabilities for a task."""
        if not required_capabilities:
            return True
        return all(cap in self.metadata.capabilities for cap in required_capabilities)

    def segregate_context(self, contexts: List[ContextReference]) -> Dict[str, List[Dict[str, Any]]]:
        """
        CRITICAL ARCHITECTURE REQUIREMENT:
        Segregate shared context into strictly typed buckets:
        - facts (Authoritative business ground truth)
        - predictions (Probabilistic AI forecasts — NEVER treated as fact)
        - recommendations (Advisory proposals)
        - agent_results (Intermediate work from peer agents)
        - human_decisions (Authoritative human choices)
        - system_events (Immutable audit/event records)
        """
        segregated: Dict[str, List[Dict[str, Any]]] = {
            "facts": [],
            "predictions": [],
            "recommendations": [],
            "agent_results": [],
            "human_decisions": [],
            "system_events": [],
        }

        for ctx in contexts:
            entry = {
                "context_id": ctx.context_id,
                "entity_type": ctx.entity_type,
                "entity_id": ctx.entity_id,
                "content": ctx.content,
                "confidence": ctx.confidence,
                "is_authoritative": ctx.is_authoritative,
            }

            if ctx.item_type == ContextItemType.FACT:
                segregated["facts"].append(entry)
            elif ctx.item_type == ContextItemType.PREDICTION:
                entry["warning"] = "Probabilistic prediction; do not treat as verified fact."
                segregated["predictions"].append(entry)
            elif ctx.item_type == ContextItemType.RECOMMENDATION:
                segregated["recommendations"].append(entry)
            elif ctx.item_type == ContextItemType.AGENT_RESULT:
                segregated["agent_results"].append(entry)
            elif ctx.item_type == ContextItemType.HUMAN_DECISION:
                entry["is_authoritative"] = True
                segregated["human_decisions"].append(entry)
            elif ctx.item_type == ContextItemType.SYSTEM_EVENT:
                segregated["system_events"].append(entry)

        return segregated

    def build_delegation(
        self,
        target_agent_id: str,
        objective: str,
        required_capabilities: List[str],
        expected_output: str,
        context_references: Optional[List[str]] = None,
        priority: TaskPriority = TaskPriority.MEDIUM,
    ) -> DelegationRequest:
        """Helper to create a structured delegation request to another workforce agent."""
        return DelegationRequest(
            target_agent_id=target_agent_id,
            objective=objective,
            required_capabilities=required_capabilities,
            context_references=context_references or [],
            expected_output=expected_output,
            priority=priority,
        )

    def build_handoff(
        self,
        originating_task_id: str,
        destination_agent_id: str,
        reason: str,
        objective: str,
        required_capability: str,
        expected_output: str,
        current_findings: Dict[str, Any],
        context_references: Optional[List[str]] = None,
        confidence: float = 0.85,
    ) -> HandoffContract:
        """Helper to create a structured handoff contract when transferring execution ownership."""
        return HandoffContract(
            originating_task_id=originating_task_id,
            source_agent_id=self.metadata.agent_id,
            destination_agent_id=destination_agent_id,
            reason=reason,
            objective=objective,
            required_capability=required_capability,
            context_references=context_references or [],
            current_findings=current_findings,
            expected_output=expected_output,
            confidence=confidence,
        )

    @abstractmethod
    def execute(
        self,
        task: WorkforceTaskContract,
        contexts: List[ContextReference],
    ) -> AgentExecutionResult:
        """
        Executes the assigned task using segregated context.
        Returns structured findings, optional delegation requests, proposed actions,
        and new context items.
        """
        pass

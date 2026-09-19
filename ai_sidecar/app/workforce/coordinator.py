"""
LogisticsHQ Phase 6.1 — Multi-Agent Workforce Coordinator
Handles safe orchestration, task execution, delegation, and handoffs on the Python AI side.
"""

import uuid
from typing import List, Dict, Any, Optional
from datetime import datetime

from app.workforce.models import (
    WorkforceTaskContract,
    ContextReference,
    AgentExecutionResult,
    WorkforceMessageContract,
    MessageType,
    TaskStatus,
    TaskPriority,
)
from app.workforce.registry import workforce_registry


class WorkforceCoordinator:
    """
    Coordinates multi-agent AI execution, inter-agent messaging, and delegation.
    CRITICAL: Does not mutate business databases or bypass Go enforcement.
    """

    def __init__(self, registry=workforce_registry):
        self.registry = registry

    def execute_task(
        self,
        task: WorkforceTaskContract,
        contexts: Optional[List[ContextReference]] = None,
    ) -> AgentExecutionResult:
        contexts = contexts or task.context_references or []

        agent = self.registry.get_agent(task.assigned_agent_id)
        if not agent:
            return AgentExecutionResult(
                task_id=task.task_id,
                agent_id=task.assigned_agent_id,
                status=TaskStatus.FAILED,
                confidence=0.0,
                summary=f"Agent '{task.assigned_agent_id}' is not registered in workforce.",
                error_code="AGENT_NOT_FOUND",
                error_message=f"Agent '{task.assigned_agent_id}' does not exist in registry.",
            )

        # Validate required capabilities
        if not agent.validate_capabilities(task.required_capabilities):
            return AgentExecutionResult(
                task_id=task.task_id,
                agent_id=task.assigned_agent_id,
                status=TaskStatus.BLOCKED,
                confidence=0.0,
                summary=f"Agent '{task.assigned_agent_id}' lacks required capabilities: {task.required_capabilities}",
                error_code="CAPABILITY_MISMATCH",
                error_message=f"Agent lacks one or more required capabilities.",
            )

        # Execute agent
        result = agent.execute(task, contexts)

        # Generate inter-agent message records for any proposed delegations
        now_iso = datetime.utcnow().isoformat() + "Z"
        generated_messages: List[WorkforceMessageContract] = []

        for delegation in result.proposed_delegations:
            msg = WorkforceMessageContract(
                message_id=f"msg-{uuid.uuid4().hex[:12]}",
                task_id=task.task_id,
                parent_task_id=task.parent_task_id,
                sender_agent_id=task.assigned_agent_id,
                recipient_agent_id=delegation.target_agent_id,
                message_type=MessageType.DELEGATION,
                objective=delegation.objective,
                requested_capability=delegation.required_capabilities[0] if delegation.required_capabilities else None,
                context_reference={"keys": delegation.context_references},
                payload={
                    "expected_output": delegation.expected_output,
                    "target_agent_id": delegation.target_agent_id,
                },
                priority=delegation.priority,
                correlation_id=task.correlation_id,
                created_at=now_iso,
            )
            generated_messages.append(msg)

        if result.proposed_handoff:
            handoff = result.proposed_handoff
            msg = WorkforceMessageContract(
                message_id=f"msg-{uuid.uuid4().hex[:12]}",
                task_id=task.task_id,
                parent_task_id=task.parent_task_id,
                sender_agent_id=handoff.source_agent_id,
                recipient_agent_id=handoff.destination_agent_id,
                message_type=MessageType.HANDOFF,
                objective=handoff.objective,
                requested_capability=handoff.required_capability,
                payload={
                    "reason": handoff.reason,
                    "current_findings": handoff.current_findings,
                    "expected_output": handoff.expected_output,
                },
                priority=task.priority,
                correlation_id=task.correlation_id,
                created_at=now_iso,
            )
            generated_messages.append(msg)

        result.messages.extend(generated_messages)
        return result


# Global coordinator instance
workforce_coordinator = WorkforceCoordinator()

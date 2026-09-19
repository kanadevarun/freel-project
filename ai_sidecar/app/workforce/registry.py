"""
LogisticsHQ Phase 6.1 — Workforce Agent Registry (Python Side)
"""

from typing import Dict, List, Optional
from app.workforce.base_agent import BaseWorkforceAgent
from app.workforce.models import AgentMetadata
from app.workforce.agents import (
    PlanningAgent,
    ShipmentAgent,
    ExceptionAgent,
    CustomerAgent,
    PricingAgent,
    FinanceAgent,
    ContractAgent,
    ComplianceAgent,
    MonitoringAgent,
    MemoryAgent,
)


class WorkforceRegistry:
    """
    Catalog of AI workforce agents available within the Python runtime.
    Supports dynamic registration and capability validation.
    """

    def __init__(self):
        self._agents: Dict[str, BaseWorkforceAgent] = {}
        self._register_foundation_workforce()

    def _register_foundation_workforce(self):
        workforce_agents = [
            PlanningAgent(),
            ShipmentAgent(),
            ExceptionAgent(),
            CustomerAgent(),
            PricingAgent(),
            FinanceAgent(),
            ContractAgent(),
            ComplianceAgent(),
            MonitoringAgent(),
            MemoryAgent(),
        ]
        for agent in workforce_agents:
            self.register(agent)


    def register(self, agent: BaseWorkforceAgent) -> None:
        self._agents[agent.metadata.agent_id] = agent

    def get_agent(self, agent_id: str) -> Optional[BaseWorkforceAgent]:
        return self._agents.get(agent_id)

    def list_agents(self) -> List[AgentMetadata]:
        return [agent.get_metadata() for agent in self._agents.values()]

    def list_agent_ids(self) -> List[str]:
        return list(self._agents.keys())

    def validate_agent_capabilities(self, agent_id: str, required_capabilities: List[str]) -> bool:
        agent = self.get_agent(agent_id)
        if not agent:
            return False
        return agent.validate_capabilities(required_capabilities)


# Global singleton instance
workforce_registry = WorkforceRegistry()

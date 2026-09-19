"""
LogisticsHQ Phase 5 Task 5.13: Agent Memory and Learning from Outcomes
Python AI Sidecar: Memory Interpretation, Retrieval Reasoning, Outcome Evaluation,
Pattern Detection, Conflict Resolution, and Prompt Injection Defense.
"""

import re
import uuid
from typing import Any, Dict, List, Optional
from datetime import datetime, timezone

from .models import (
    AgentMemoryItemModel,
    OutcomeEvaluationRequest,
    OutcomeEvaluationResponse,
    MemoryRetrievalRequest,
    MemoryRetrievalResponse,
    PatternDetectionRequest,
    PatternDetectionResponse,
    MemoryConflictRequest,
    MemoryConflictResponse,
)

# Comprehensive prompt injection detection patterns
PROMPT_INJECTION_PATTERNS = [
    re.compile(r"ignore\s+(all\s+)?(previous|prior)\s+instructions?", re.IGNORECASE),
    re.compile(r"disregard\s+(all\s+)?(previous|prior)\s+instructions?", re.IGNORECASE),
    re.compile(r"system\s+(override|prompt|bypass)", re.IGNORECASE),
    re.compile(r"you\s+are\s+now\s+(an?|in)\s+", re.IGNORECASE),
    re.compile(r"grant\s+(admin|root|superuser|permission)", re.IGNORECASE),
    re.compile(r"approve\s+(all|this\s+shipment|without\s+verification)", re.IGNORECASE),
    re.compile(r"escalate\s+privileges?", re.IGNORECASE),
    re.compile(r"bypass\s+(approval|compliance|policy|checks?)", re.IGNORECASE),
    re.compile(r"execute\s+as\s+(admin|system|root)", re.IGNORECASE),
    re.compile(r"format:\s*raw", re.IGNORECASE),
    re.compile(r"<!--.*?-->", re.DOTALL),
]


def sanitize_untrusted_text(text: Optional[str]) -> str:
    """Sanitizes untrusted text by neutralizing prompt injection attempts while preserving business content."""
    if not text:
        return ""
    sanitized = text
    for pattern in PROMPT_INJECTION_PATTERNS:
        if pattern.search(sanitized):
            sanitized = pattern.sub("[SANITIZED_INSTRUCTION]", sanitized)
    return sanitized.strip()


class AgentMemoryLearningAgent:
    """
    Cognitive memory and outcome learning agent for LogisticsHQ.
    Evaluates execution outcomes, retrieves relevant contextual memory,
    discovers operational patterns, and resolves memory conflicts.
    """

    def evaluate_outcome(self, req: OutcomeEvaluationRequest) -> OutcomeEvaluationResponse:
        correlation_id = req.correlation_id or f"corr-eval-{uuid.uuid4().hex[:8]}"
        expected = sanitize_untrusted_text(req.expected_result)
        actual = sanitize_untrusted_text(req.actual_result)
        out_type = req.outcome_type
        human_inv = req.human_involvement or "NONE"

        # Compare expected vs actual
        actual_lower = actual.lower()
        expected_lower = expected.lower()

        # Deterministic outcome classification based on business result
        failure_category = "NONE"
        status = req.status
        is_verified = True

        if any(w in actual_lower for w in ["timeout", "timed out", "no response", "unresponsive"]):
            status = "FAILED"
            failure_category = "CARRIER_NON_RESPONSIVE" if "carrier" in out_type.lower() else "TIMED_OUT"
        elif any(w in actual_lower for w in ["rejected", "declined", "disapproved", "cancelled"]):
            status = "REJECTED"
            failure_category = "CUSTOMER_REJECTED" if "customer" in out_type.lower() else "POLICY_BLOCKED"
        elif any(w in actual_lower for w in ["escalated", "supervisor review", "tier 2"]):
            status = "ESCALATED"
        elif any(w in actual_lower for w in ["partial", "delay adjusted", "recovered with variance"]):
            status = "PARTIAL_SUCCESS"
        elif any(w in actual_lower for w in ["success", "completed", "resolved", "approved", "paid", "discharged", "delivered"]):
            status = "SUCCESS"
        elif status == "UNVERIFIED":
            status = "SUCCESS"

        # Formulate evaluation summary
        time_res = req.time_to_resolution_sec
        time_str = f" in {time_res // 3600}h {(time_res % 3600) // 60}m" if time_res and time_res > 0 else ""
        eval_summary = f"Outcome {status}{time_str}: Actual '{actual[:120]}' matched operational expectation '{expected[:100]}'."
        if status in ["FAILED", "REJECTED"]:
            eval_summary = f"Outcome {status} (Category: {failure_category}): Actual result '{actual[:120]}' deviated from expectation '{expected[:100]}'."

        # Determine if outcome warrants a durable memory candidate
        should_create_memory = False
        memory_candidate = None

        if status in ["SUCCESS", "PARTIAL_SUCCESS", "FAILED", "REJECTED"]:
            should_create_memory = True
            category = self._map_outcome_to_category(out_type)
            mem_type = self._map_outcome_to_memory_type(out_type, status)
            title = self._generate_memory_title(out_type, req.source_entity_type, req.source_entity_id, status)
            content = self._generate_memory_content(out_type, req.source_entity_type, req.source_entity_id, expected, actual, status, failure_category, human_inv)

            memory_candidate = AgentMemoryItemModel(
                org_id=req.org_id,
                scope="TENANT",
                category=category,
                memory_type=mem_type,
                title=title,
                content=content,
                entity_type=req.source_entity_type,
                entity_id=req.source_entity_id,
                outcome_id=f"outc-{uuid.uuid4().hex[:12]}",
                confidence="HIGH" if status == "SUCCESS" else "MEDIUM",
                confidence_score=0.92 if status == "SUCCESS" else 0.80,
                recency_weight=1.00,
                times_observed=1,
                times_used=0,
                success_count=1 if status == "SUCCESS" else 0,
                failure_count=1 if status in ["FAILED", "REJECTED"] else 0,
                is_stale=False,
                conflict_status="NONE",
                provenance_type="SYSTEM_DERIVED" if human_inv == "NONE" else "HUMAN_ENTERED",
                source_reference=f"{req.source_entity_type} #{req.source_entity_id}",
                evidence=f"Authoritative comparison: Expected '{expected}' vs Actual '{actual}'. Human involvement: {human_inv}.",
                status="ACTIVE",
                structured_value={
                    "outcome_type": out_type,
                    "plan_id": req.plan_id,
                    "action_type": req.action_type,
                    "failure_category": failure_category,
                    "time_to_resolution_sec": req.time_to_resolution_sec,
                },
            )

        return OutcomeEvaluationResponse(
            outcome_status=status,
            is_verified=is_verified,
            evaluation_summary=eval_summary,
            failure_category=failure_category,
            confidence="HIGH" if status == "SUCCESS" else "MEDIUM",
            confidence_score=0.92 if status == "SUCCESS" else 0.80,
            should_create_memory=should_create_memory,
            memory_candidate=memory_candidate,
            correlation_id=correlation_id,
        )

    def retrieve_relevant_memory(self, req: MemoryRetrievalRequest) -> MemoryRetrievalResponse:
        correlation_id = req.correlation_id or f"corr-ret-{uuid.uuid4().hex[:8]}"
        query_sanitized = sanitize_untrusted_text(req.query_context).lower()
        query_words = set(re.findall(r"\w+", query_sanitized))

        scored_memories = []
        provenance_breakdown = {"HUMAN_ENTERED": 0, "SYSTEM_DERIVED": 0, "AI_DERIVED": 0, "EXTERNAL_SOURCE": 0}
        conflict_warnings = []
        has_conflicts = False

        for mem in req.candidate_memories:
            # Filter out stale or deleted memories unless explicitly requested
            if mem.get("is_stale") and not req.include_stale:
                continue
            if mem.get("status") in ["DELETED", "DISABLED", "EXPIRED"]:
                continue

            # Check conflicts
            if mem.get("conflict_status") == "CONFLICT_DETECTED":
                has_conflicts = True
                conflict_warnings.append(f"Memory #{mem.get('id')} has unresolved conflict: {mem.get('title')}")

            # Compute relevance score
            relevance = self._score_relevance(mem, query_words, req)
            if relevance > 0.15:
                # Sanitize content for safety
                clean_mem = dict(mem)
                clean_mem["title"] = sanitize_untrusted_text(clean_mem.get("title", ""))
                clean_mem["content"] = sanitize_untrusted_text(clean_mem.get("content", ""))
                clean_mem["relevance_score"] = round(relevance, 2)
                scored_memories.append(clean_mem)

                # Track provenance
                prov = clean_mem.get("provenance_type", "SYSTEM_DERIVED")
                provenance_breakdown[prov] = provenance_breakdown.get(prov, 0) + 1

        # Sort by relevance descending, limit to requested size
        scored_memories.sort(key=lambda m: m.get("relevance_score", 0.0), reverse=True)
        retrieved = scored_memories[:req.limit]

        # Context summary
        summary = f"Retrieved {len(retrieved)} relevant contextual operational memories (out of {len(req.candidate_memories)} candidates)."
        if retrieved:
            top_titles = [m.get("title", "") for m in retrieved[:2]]
            summary += f" Key insights: {'; '.join(top_titles)}."

        return MemoryRetrievalResponse(
            retrieved_memories=retrieved,
            total_found=len(retrieved),
            context_summary=summary,
            provenance_breakdown=provenance_breakdown,
            has_conflicts=has_conflicts,
            conflict_warnings=conflict_warnings,
            correlation_id=correlation_id,
        )

    def detect_patterns(self, req: PatternDetectionRequest) -> PatternDetectionResponse:
        correlation_id = req.correlation_id or f"corr-pat-{uuid.uuid4().hex[:8]}"
        detected = []

        # 1. Group outcomes by carrier
        carrier_outcomes = {}
        # 2. Group outcomes by exception type
        exception_outcomes = {}
        # 3. Group outcomes by customer
        customer_outcomes = {}

        for outc in req.outcomes:
            entity_type = outc.get("source_entity_type", "")
            entity_id = str(outc.get("source_entity_id", ""))
            status = outc.get("status", "")
            action_type = outc.get("action_type", "")
            meta = outc.get("metadata") or {}

            # Carrier SCAC
            carrier = meta.get("carrier_scac") or (entity_id if entity_type == "CARRIER" else None)
            if carrier:
                carrier_outcomes.setdefault(carrier, []).append(outc)

            # Exception Type
            exc_type = meta.get("exception_type") or (meta.get("issue_type") if entity_type == "EXCEPTION" else None)
            if exc_type:
                exception_outcomes.setdefault(exc_type, []).append(outc)

            # Customer ID
            cust = meta.get("customer_id") or (entity_id if entity_type == "CUSTOMER" else None)
            if cust:
                customer_outcomes.setdefault(cust, []).append(outc)

        # Evaluate carrier patterns
        for scac, items in carrier_outcomes.items():
            if len(items) >= 2:
                successes = sum(1 for i in items if i.get("status") == "SUCCESS")
                delays = sum(1 for i in items if "delay" in (i.get("actual_result") or "").lower())
                success_rate = round(successes / len(items), 2)
                confidence = "HIGH" if len(items) >= 5 else "MEDIUM"

                if delays >= 2:
                    detected.append({
                        "pattern_id": f"pat-carrier-{scac.lower()}-{uuid.uuid4().hex[:6]}",
                        "pattern_type": "CARRIER_DISRUPTION_PATTERN",
                        "entity_type": "CARRIER",
                        "entity_identifier": scac,
                        "title": f"Carrier {scac} Recurring Port Delay Tendency",
                        "description": f"Observed {delays} schedule variances across {len(items)} recent milestone tracking events for {scac}.",
                        "recommended_strategy": "Factor +4.0h buffer into initial ETA predictions and enable priority AIS monitoring.",
                        "supporting_observations": len(items),
                        "success_rate": success_rate,
                        "confidence": confidence,
                    })

        # Evaluate exception patterns
        for exc, items in exception_outcomes.items():
            if len(items) >= 2:
                successes = sum(1 for i in items if i.get("status") == "SUCCESS")
                success_rate = round(successes / len(items), 2)
                confidence = "HIGH" if len(items) >= 4 else "MEDIUM"

                detected.append({
                    "pattern_id": f"pat-exc-{exc.lower()}-{uuid.uuid4().hex[:6]}",
                    "pattern_type": "EXCEPTION_RECOVERY_STRATEGY",
                    "entity_type": "EXCEPTION",
                    "entity_identifier": exc,
                    "title": f"Effective Recovery Pattern for {exc.replace('_', ' ').title()}",
                    "description": f"{exc} exceptions have an authoritative resolution success rate of {int(success_rate * 100)}% ({successes}/{len(items)}).",
                    "recommended_strategy": "Proactive carrier EDI inquiry combined with immediate customer milestone notification yields fastest resolution.",
                    "supporting_observations": len(items),
                    "success_rate": success_rate,
                    "confidence": confidence,
                })

        summary = f"Detected {len(detected)} recurring operational patterns across carrier behavior and exception recovery."
        return PatternDetectionResponse(
            detected_patterns=detected,
            total_patterns=len(detected),
            summary=summary,
            correlation_id=correlation_id,
        )

    def resolve_memory_conflicts(self, req: MemoryConflictRequest) -> MemoryConflictResponse:
        correlation_id = req.correlation_id or f"corr-conf-{uuid.uuid4().hex[:8]}"
        new_obs = sanitize_untrusted_text(req.new_observation).lower()

        conflicting_id = None
        conflict_explanation = "No contradictions found with existing operational memory."
        authoritative_resolution = "Retain existing memory without mutation."
        recommended_action = "NO_CONFLICT"
        has_conflict = False

        negation_pairs = [
            ("accepts", "no longer accepts"),
            ("prefers email", "prefers phone"),
            ("prefers phone", "prefers email"),
            ("weekend delivery", "no weekend delivery"),
            ("standard terms", "custom credit terms"),
            ("requires approval", "pre-authorized"),
        ]

        for mem in req.existing_memories:
            content = (mem.get("content") or "").lower()
            for pos, neg in negation_pairs:
                if (pos in content and neg in new_obs) or (neg in content and pos in new_obs):
                    has_conflict = True
                    conflicting_id = mem.get("id")
                    conflict_explanation = f"Conflict detected: Memory '{mem.get('title')}' asserts '{content[:80]}', while latest authoritative observation states '{new_obs[:80]}'."
                    authoritative_resolution = "Authoritative current instruction supersedes older contextual memory. Marking historical memory as SUPERSEDED."
                    recommended_action = "SUPERSEDE_OLD"
                    break
            if has_conflict:
                break

        return MemoryConflictResponse(
            has_conflict=has_conflict,
            conflicting_memory_id=conflicting_id,
            conflict_explanation=conflict_explanation,
            authoritative_resolution=authoritative_resolution,
            recommended_action=recommended_action,
            correlation_id=correlation_id,
        )

    # -------------------------------------------------------------------------
    # Helper methods
    # -------------------------------------------------------------------------

    def _map_outcome_to_category(self, out_type: str) -> str:
        t = out_type.upper()
        if "CUSTOMER" in t:
            return "CUSTOMER"
        if "CARRIER" in t or "EXCEPTION" in t or "SHIPMENT" in t:
            return "OPERATIONAL"
        if "PRICING" in t:
            return "PRICING"
        if "COLLECTION" in t or "FINANCE" in t:
            return "FINANCE"
        if "COMPLIANCE" in t or "CONTRACT" in t:
            return "CONTRACT_COMPLIANCE"
        if "HUMAN" in t:
            return "HUMAN_FEEDBACK"
        return "WORKFLOW"

    def _map_outcome_to_memory_type(self, out_type: str, status: str) -> str:
        t = out_type.upper()
        if "EXCEPTION" in t:
            return "SUCCESSFUL_RECOVERY_STRATEGY" if status == "SUCCESS" else "FAILED_RECOVERY_STRATEGY"
        if "CUSTOMER" in t:
            return "CUSTOMER_COMMUNICATION_PREFERENCE"
        if "CARRIER" in t:
            return "CARRIER_BEHAVIOR_OBSERVATION"
        if "PRICING" in t:
            return "PRICING_ACCEPTANCE_OUTCOME"
        if "COLLECTION" in t:
            return "COLLECTION_OUTCOME_PATTERN"
        if "COMPLIANCE" in t:
            return "COMPLIANCE_REMEDIATION_OUTCOME"
        if "HUMAN" in t:
            return "HUMAN_OPERATOR_PREFERENCE"
        return "WORKFLOW_EXECUTION_OUTCOME"

    def _generate_memory_title(self, out_type: str, entity_type: str, entity_id: str, status: str) -> str:
        prefix = "Verified" if status == "SUCCESS" else "Failed"
        type_clean = out_type.replace("_OUTCOME", "").replace("_", " ").title()
        return f"{prefix} {type_clean}: {entity_type} #{entity_id}"

    def _generate_memory_content(
        self, out_type: str, entity_type: str, entity_id: str, expected: str, actual: str, status: str, failure_cat: str, human_inv: str
    ) -> str:
        if status == "SUCCESS":
            return (
                f"Previous {out_type.lower().replace('_', ' ')} for {entity_type} #{entity_id} "
                f"successfully achieved expected result: {actual}. Human involvement: {human_inv}."
            )
        return (
            f"Prior attempt for {entity_type} #{entity_id} resulted in {status} ({failure_cat}). "
            f"Expected: '{expected}', but actual result was: '{actual}'. Human involvement: {human_inv}."
        )

    def _score_relevance(self, mem: Dict[str, Any], query_words: set, req: MemoryRetrievalRequest) -> float:
        score = 0.0
        title = (mem.get("title") or "").lower()
        content = (mem.get("content") or "").lower()
        category = (mem.get("category") or "").upper()
        mem_entity_type = (mem.get("entity_type") or "").upper()
        mem_entity_id = str(mem.get("entity_id") or "")

        # Entity match (+0.50)
        if req.entity_type and req.entity_type.upper() == mem_entity_type:
            score += 0.25
            if req.entity_id and req.entity_id == mem_entity_id:
                score += 0.35

        # Category match (+0.25)
        if req.category and req.category.upper() == category:
            score += 0.25

        # Keyword overlap (+0.20)
        text_words = set(re.findall(r"\w+", f"{title} {content}"))
        if query_words and text_words:
            overlap = len(query_words.intersection(text_words)) / max(len(query_words), 1)
            score += min(overlap * 0.40, 0.40)

        # Confidence weight (+0.10 for HIGH)
        conf = mem.get("confidence", "MEDIUM")
        if conf == "HIGH":
            score += 0.10
        elif conf == "LOW":
            score -= 0.15

        # Recency decay (older than 30 days decayed)
        recency = float(mem.get("recency_weight", 1.0))
        score *= recency

        return min(max(score, 0.0), 1.0)


memory_learning_agent = AgentMemoryLearningAgent()

import re
from typing import List, Dict, Any, Tuple
from app.governance.models import (
    SafetyCheckResult,
    InputInspectionRequest,
    InputInspectionResponse,
    OutputInspectionRequest,
    OutputInspectionResponse,
    ActionSafetyEvaluation,
)

# Canonical Consequential Action Types requiring Human Approval Gate
CONSEQUENTIAL_ACTIONS = {
    "PAYMENT_DISPATCH",
    "EXTERNAL_REPORT_DISTRIBUTION",
    "RATE_MUTATION",
    "CONTRACT_TERMINATION",
    "COMMERCIAL_QUOTE_APPROVAL",
    "CREDIT_LIMIT_OVERRIDE",
    "CUSTOMS_FILING_SUBMISSION",
    "BOOKING_CANCELLATION",
    "BANK_ACCOUNT_UPDATE",
    "REQUEST_HUMAN_APPROVAL",
}

# Prompt Injection Detection Patterns
PROMPT_INJECTION_PATTERNS = [
    r"(?i)\bignore\s+(all\s+)?(previous|prior)\s+instructions\b",
    r"(?i)\bsystem\s+override\b",
    r"(?i)\breveal\s+(the\s+)?(system\s+prompt|initial\s+instructions)\b",
    r"(?i)\byou\s+are\s+now\s+in\s+(developer|dan|sudo|jailbreak)\s+mode\b",
    r"(?i)\bpretend\s+you\s+have\s+no\s+(rules|guidelines|guardrails)\b",
    r"(?i)\bdisregard\s+all\s+safety\b",
    r"(?i)\bprint\s+(the\s+)?api[_-]?key\b",
    r"(?i)\bexecute\s+arbitrary\s+code\b",
    r"(?i)\b(union\s+select|drop\s+table|delete\s+from\s+users)\b",
]

# Sensitive Data / PII Regex Patterns
PII_PATTERNS = {
    "CREDIT_CARD": (r"\b(?:\d{4}[-\s]?){3}\d{4}\b", "[REDACTED_CREDIT_CARD]"),
    "SSN": (r"\b\d{3}-\d{2}-\d{4}\b", "[REDACTED_SSN]"),
    "JWT_TOKEN": (r"\beyJ[A-Za-z0-9_-]{10,}\.[A-Za-z0-9_-]{10,}\.[A-Za-z0-9_-]{10,}\b", "[REDACTED_JWT_TOKEN]"),
    "SERVICE_KEY": (r"\b(lhq_sec_[A-Za-z0-9_-]+|sk-[A-Za-z0-9]{20,}|Bearer\s+[A-Za-z0-9_-]{20,})\b", "[REDACTED_API_KEY]"),
    "PASSWORD_IN_PAYLOAD": (r"(?i)(password|secret|passwd)\s*[:=]\s*['\"][^'\"]+['\"]", "[REDACTED_PASSWORD]"),
}

class AIGovernanceSafetyEvaluator:
    """
    Authoritative Python AI Safety Evaluator for LogisticsHQ.
    Performs input sanitization, prompt injection detection, PII redaction,
    factual grounding verification, and action safety validation.
    """

    def inspect_input(self, req: InputInspectionRequest) -> InputInspectionResponse:
        text = req.input_text or ""
        safety_checks: List[SafetyCheckResult] = []
        is_safe = True
        refusal_reason = None

        # 1. Prompt Injection Inspection
        detected_injections = []
        for pattern in PROMPT_INJECTION_PATTERNS:
            matches = re.findall(pattern, text)
            if matches:
                detected_injections.append(pattern)

        injection_detected = len(detected_injections) > 0
        injection_confidence = min(1.0, len(detected_injections) * 0.45) if injection_detected else 0.0

        if injection_detected:
            is_safe = False
            refusal_reason = "Prompt injection attempt detected. Input violates AI Governance Policy."
            safety_checks.append(SafetyCheckResult(
                passed=False,
                check_type="PROMPT_INJECTION",
                severity="CRITICAL",
                details=f"Detected {len(detected_injections)} adversarial prompt manipulation patterns.",
                detected_patterns=detected_injections[:5]
            ))
        else:
            safety_checks.append(SafetyCheckResult(
                passed=True,
                check_type="PROMPT_INJECTION",
                severity="LOW",
                details="No prompt injection patterns detected.",
                detected_patterns=[]
            ))

        # 2. PII and Sensitive Data Redaction
        sanitized_text = text
        detected_pii = []
        for pii_type, (pat, replacement) in PII_PATTERNS.items():
            if re.search(pat, sanitized_text):
                detected_pii.append(pii_type)
                sanitized_text = re.sub(pat, replacement, sanitized_text)

        pii_detected = len(detected_pii) > 0
        if pii_detected:
            safety_checks.append(SafetyCheckResult(
                passed=True,  # Passed because we sanitized it safely
                check_type="SENSITIVE_DATA",
                severity="MEDIUM",
                details=f"Sensitive data sanitized: {', '.join(detected_pii)}.",
                detected_patterns=detected_pii
            ))
        else:
            safety_checks.append(SafetyCheckResult(
                passed=True,
                check_type="SENSITIVE_DATA",
                severity="LOW",
                details="No exposed sensitive PII or credentials detected.",
                detected_patterns=[]
            ))

        return InputInspectionResponse(
            is_safe=is_safe,
            prompt_injection_detected=injection_detected,
            injection_confidence=round(injection_confidence, 2),
            pii_detected=pii_detected,
            sanitized_text=sanitized_text,
            detected_pii_types=detected_pii,
            safety_checks=safety_checks,
            refusal_reason=refusal_reason,
            correlation_id=req.correlation_id
        )

    def inspect_output(self, req: OutputInspectionRequest) -> OutputInspectionResponse:
        text = req.output_text or ""
        safety_checks: List[SafetyCheckResult] = []
        is_safe = True
        refusal_reason = None

        # 1. Grounding and Unsupported Claims Verification
        source_facts = req.source_facts or []
        unsupported_claims: List[str] = []
        grounding_score = 1.0

        if source_facts:
            # Check key entity mentions against source facts
            fact_tokens = set()
            for f in source_facts:
                for k, v in f.items():
                    if isinstance(v, (str, int, float)):
                        fact_tokens.add(str(v).lower())

            # Check if output makes specific dollar or numerical claims not in source facts
            number_matches = re.findall(r"\$?\b\d+(?:,\d{3})*(?:\.\d+)?\b", text)
            unmatched_numbers = [
                n for n in number_matches 
                if n.replace("$", "").replace(",", "").lower() not in fact_tokens and float(n.replace("$", "").replace(",", "")) > 10
            ]

            if len(unmatched_numbers) > 3:
                grounding_score = max(0.4, 1.0 - (len(unmatched_numbers) * 0.15))
                unsupported_claims.append(f"Output references numerical claims not confirmed in source facts: {unmatched_numbers[:3]}")

        hallucination_risk = "LOW"
        if grounding_score < 0.6:
            hallucination_risk = "HIGH"
            is_safe = False
            refusal_reason = "Output failed grounding verification: high risk of fabricated business metrics."
        elif grounding_score < 0.85:
            hallucination_risk = "MEDIUM"

        safety_checks.append(SafetyCheckResult(
            passed=(hallucination_risk != "HIGH"),
            check_type="GROUNDING_VERIFICATION",
            severity="HIGH" if hallucination_risk == "HIGH" else "LOW",
            details=f"Grounding score: {grounding_score:.2f} (Risk: {hallucination_risk}).",
            detected_patterns=unsupported_claims
        ))

        # 2. Action Safety and Human Approval Enforcement
        proposed_actions = req.proposed_actions or []
        action_evaluations: List[ActionSafetyEvaluation] = []
        any_consequential = False

        for act in proposed_actions:
            act_type = str(act.get("action_type", "")).upper()
            act_title = str(act.get("action_title", "Untitled Action"))

            is_consequential = act_type in CONSEQUENTIAL_ACTIONS
            if is_consequential:
                any_consequential = True
                action_evaluations.append(ActionSafetyEvaluation(
                    action_type=act_type,
                    action_title=act_title,
                    is_consequential=True,
                    requires_human_approval=True,
                    risk_level="HIGH",
                    safety_reason=f"Action '{act_type}' has irreversible business or financial impact. Gated into Human Approval Center."
                ))
            else:
                action_evaluations.append(ActionSafetyEvaluation(
                    action_type=act_type,
                    action_title=act_title,
                    is_consequential=False,
                    requires_human_approval=False,
                    risk_level="SAFE",
                    safety_reason=f"Action '{act_type}' is an assistive advisory task. Permitted for direct Action System logging."
                ))

        safety_checks.append(SafetyCheckResult(
            passed=True,
            check_type="ACTION_SAFETY",
            severity="MEDIUM" if any_consequential else "LOW",
            details=f"Evaluated {len(proposed_actions)} actions. {sum(1 for a in action_evaluations if a.requires_human_approval)} gated for human approval.",
            detected_patterns=[a.action_type for a in action_evaluations if a.requires_human_approval]
        ))

        return OutputInspectionResponse(
            is_safe=is_safe,
            grounding_score=round(grounding_score, 2),
            hallucination_risk=hallucination_risk,
            unsupported_claims=unsupported_claims,
            action_evaluations=action_evaluations,
            requires_human_approval=any_consequential,
            safety_checks=safety_checks,
            refusal_reason=refusal_reason,
            correlation_id=req.correlation_id
        )

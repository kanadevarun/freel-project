from typing import List, Dict, Any
from app.governance.models import (
    QualityEvaluationRequest,
    QualityEvaluationResponse,
    TestCaseResult,
)

class AIQualityEvaluator:
    """
    Authoritative Python Quality Evaluation Engine for LogisticsHQ AI Workflows.
    Evaluates schema adherence, response quality, and factual grounding across scenarios.
    """

    def evaluate_dataset(self, req: QualityEvaluationRequest) -> QualityEvaluationResponse:
        cases = req.test_cases or []
        results: List[TestCaseResult] = []
        passed_count = 0
        total_grounding = 0.0

        for case in cases:
            case_id = str(case.get("case_id", f"case-{len(results) + 1}"))
            output = case.get("output", {})
            expected_fields = case.get("expected_fields", [])
            latency = int(case.get("latency_ms", 120))

            # Check schema adherence
            missing_fields = [f for f in expected_fields if f not in output]
            schema_valid = len(missing_fields) == 0

            # Grounding check
            grounding = float(case.get("grounding_score", 1.0 if schema_valid else 0.5))
            total_grounding += grounding

            passed = schema_valid and grounding >= 0.75 and latency <= 5000
            if passed:
                passed_count += 1

            err_msg = None
            if not schema_valid:
                err_msg = f"Missing required fields: {', '.join(missing_fields)}"
            elif grounding < 0.75:
                err_msg = f"Insufficient grounding score: {grounding:.2f}"

            results.append(TestCaseResult(
                case_id=case_id,
                passed=passed,
                schema_valid=schema_valid,
                grounding_score=round(grounding, 2),
                latency_ms=latency,
                error_message=err_msg
            ))

        total_cases = len(cases)
        pass_rate = round((passed_count / total_cases * 100.0), 1) if total_cases > 0 else 100.0
        avg_grounding = round((total_grounding / total_cases), 2) if total_cases > 0 else 1.0

        overall_status = "READY_FOR_PRODUCTION"
        if pass_rate < 80.0 or avg_grounding < 0.8:
            overall_status = "BLOCKED"
        elif pass_rate < 95.0 or avg_grounding < 0.9:
            overall_status = "NEEDS_REVIEW"

        summary = (
            f"Evaluated {total_cases} test cases for workflow '{req.workflow_name}'. "
            f"Pass rate: {pass_rate}%, Average Grounding: {avg_grounding}. Status: {overall_status}."
        )

        return QualityEvaluationResponse(
            test_run_id=req.test_run_id,
            workflow_name=req.workflow_name,
            total_cases=total_cases,
            passed_cases=passed_count,
            pass_rate=pass_rate,
            avg_grounding_score=avg_grounding,
            overall_status=overall_status,
            results=results,
            summary=summary,
            correlation_id=req.correlation_id
        )

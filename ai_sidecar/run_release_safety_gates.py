#!/usr/bin/env python3
"""
run_release_safety_gates.py — Phase 0 Release Verification & AI Safety-Gate Command

Usage:
    python run_release_safety_gates.py [--org-id 2] [--output eval_summary.json]

Requirements:
- Deterministic test mode only (zero live Gemini/OpenAI credentials, zero real emails, zero real external mutations)
- Evaluates 84 scenarios across 3 categories:
    1. Core Safety Gates (Tenant isolation, Authorization/HITL, Side effects, Data safety)
    2. Agent-Specific Evaluations (Pricing, Sales, Operations, Contracts, Compliance, Finance, Leads, Outreach)
    3. Business Invariant Regressions (RFQ, Quote, Booking, Shipment, Milestone, Invoice, Audit, Provenance)
- Persists full execution telemetry into MariaDB table `ai_evaluation_results`
- Produces machine-readable JSON output (`eval_summary.json`)
- Exits with return code 0 on complete pass, or return code 1 if any gate fails.
"""

import os
import sys
import json
import argparse

# Ensure ai_sidecar root is on Python path
CURRENT_DIR = os.path.dirname(os.path.abspath(__file__))
if CURRENT_DIR not in sys.path:
    sys.path.insert(0, CURRENT_DIR)

# Force test mode in environment so no live API calls or mock leaks happen
os.environ["APP_ENV"] = "test"
os.environ["ALLOW_MOCK_FALLBACK"] = "true"

from app.eval.runner import SafetyGateRunner


def main():
    parser = argparse.ArgumentParser(description="LogisticsHQ Phase 0 AI Evaluation & Release Safety Gates")
    parser.add_argument("--org-id", type=int, default=2, help="Target organization ID for evaluation")
    parser.add_argument("--output", type=str, default="eval_summary.json", help="Path for summary JSON")
    args = parser.parse_args()

    runner = SafetyGateRunner()
    summary = runner.run_all(target_org_id=args.org_id)

    # Write machine-readable JSON report
    output_path = os.path.join(CURRENT_DIR, args.output)
    with open(output_path, "w", encoding="utf-8") as f:
        json.dump(summary, f, indent=2)

    print("\n" + "=" * 80)
    print("LOGISTICSHQ PHASE 0 — AI EVALUATION & SAFETY GATES SUMMARY")
    print("=" * 80)
    print(f"Test Run ID       : {summary['test_run_id']}")
    print(f"Timestamp         : {summary['timestamp']}")
    print(f"Total Scenarios   : {summary['total_scenarios']}")
    print(f"Passed            : {summary['passed']}")
    print(f"Failed            : {summary['failed']}")
    print(f"Success Rate      : {summary['success_rate']}%")
    print(f"Total Duration    : {summary['duration_ms']} ms")
    print("-" * 80)
    for cat, data in summary["categories"].items():
        print(f"  • {cat:<20}: {data['passed']}/{data['total']} passed")
    print("-" * 80)
    print(f"Summary JSON saved to: {output_path}")

    if not summary["is_release_ready"]:
        print("\033[91mCRITICAL: Safety gate verification FAILED. Release blocked.\033[0m")
        sys.exit(1)

    print("\033[92mSUCCESS: All Phase 0 AI Safety Gates and Evaluations PASSED. Release approved.\033[0m\n")
    sys.exit(0)


if __name__ == "__main__":
    main()

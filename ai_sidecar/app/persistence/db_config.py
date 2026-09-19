"""
db_config.py — Centralized, Hardened Database Configuration for LogisticsHQ AI Sidecar

Guarantees:
1. Robust parsing of standard mysql:// and mariadb:// connection strings.
2. Robust parsing of Go-style DSN strings (user:pass@tcp(host:port)/dbname).
3. Discrete environment variable fallbacks (DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME).
4. Safe credential redaction for logging.
5. Strict whitelisting of AI coordination tables.
"""

import os
import re
from urllib.parse import urlparse
from typing import Dict, Any

# Authorized AI coordination and observability tables that Python may access
ALLOWED_AI_TABLES = frozenset({
    "ai_processing_tasks",
    "ai_checkpoints",
    "ai_checkpoint_writes",
    "ai_execution_traces",
    "ai_evaluation_results",
    "audit_logs",
})

# Protected business tables that Python must NEVER directly mutate
PROTECTED_BUSINESS_TABLES = frozenset({
    "organizations",
    "users",
    "roles",
    "org_members",
    "customers",
    "leads",
    "lead_interactions",
    "rfqs",
    "quotations",
    "bookings",
    "shipments",
    "shipment_milestones",
    "shipment_exceptions",
    "invoices",
    "customer_invoices",
    "contracts",
    "contract_compliance_monitoring_plans",
    "approval_requests",
    "action_idempotency_keys",
})


def get_db_url() -> str:
    """Retrieves DB connection string from environment with host.docker.internal mapping."""
    url = os.getenv("DB_URL", "")
    if not url:
        user = os.getenv("DB_USER", "root")
        password = os.getenv("DB_PASSWORD", "")
        host = os.getenv("DB_HOST", "127.0.0.1")
        port = os.getenv("DB_PORT", "3306")
        db = os.getenv("DB_NAME", "freel_mysql")
        auth = f"{user}:{password}" if password else user
        url = f"mysql://{auth}@{host}:{port}/{db}"

    if "host.docker.internal" in url:
        url = url.replace("host.docker.internal", "127.0.0.1")
    return url


def parse_db_url(url: str) -> Dict[str, Any]:
    """
    Parses database URL into connection parameters supporting:
    - Standard mysql:// and mariadb:// URLs
    - Go-style DSNs: user:pass@tcp(host:port)/dbname?params
    - Discrete environment variables as first-class fallbacks
    """
    env_host = os.getenv("DB_HOST", "127.0.0.1")
    env_port = int(os.getenv("DB_PORT", "3306"))
    env_user = os.getenv("DB_USER", "root")
    env_pass = os.getenv("DB_PASSWORD", "")
    env_db = os.getenv("DB_NAME", "freel_mysql")

    url = (url or "").strip().strip('"').strip("'")
    if not url:
        return {
            "host": env_host,
            "port": env_port,
            "user": env_user,
            "password": env_pass,
            "db": env_db,
        }

    # Check for Go-style DSN: user:pass@tcp(host:port)/dbname
    go_dsn_match = re.match(
        r"^(?P<user>[^:@]+)?(?::(?P<pass>[^@]*))?@tcp\((?P<host>[^:]+):(?P<port>\d+)\)/(?P<db>[^?]+)",
        url
    )
    if go_dsn_match:
        m = go_dsn_match.groupdict()
        return {
            "host": m["host"] or env_host,
            "port": int(m["port"]) if m.get("port") else env_port,
            "user": m["user"] if m.get("user") is not None else env_user,
            "password": m["pass"] if m.get("pass") is not None else env_pass,
            "db": m["db"] or env_db,
        }

    # Standard URL parsing
    parsed = urlparse(url)
    host = parsed.hostname or env_host
    port = parsed.port or env_port
    user = parsed.username or env_user
    password = parsed.password if parsed.password is not None else env_pass
    path_db = parsed.path.lstrip('/') if parsed.path else ""
    db = path_db.split('?')[0] if path_db else env_db

    return {
        "host": host,
        "port": port,
        "user": user,
        "password": password,
        "db": db,
    }


def mask_db_config(cfg: Dict[str, Any]) -> Dict[str, Any]:
    """Returns a copy of database configuration with password redacted for safe logging."""
    masked = dict(cfg)
    if masked.get("password"):
        masked["password"] = "******"
    return masked

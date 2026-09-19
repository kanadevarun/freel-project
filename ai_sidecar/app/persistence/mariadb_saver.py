import os
import json
import random
import asyncio
from contextlib import contextmanager
from typing import Any, AsyncIterator, Iterator, Mapping, Sequence, Tuple, Optional, cast
from urllib.parse import urlparse

import pymysql
from langgraph.checkpoint.base import (
    BaseCheckpointSaver,
    ChannelVersions,
    Checkpoint,
    CheckpointMetadata,
    CheckpointTuple,
    WRITES_IDX_MAP,
    get_checkpoint_id,
    get_checkpoint_metadata,
)
from langgraph.checkpoint.serde.base import SerializerProtocol
from langgraph.checkpoint.serde.jsonplus import JsonPlusSerializer
from langchain_core.runnables import RunnableConfig

from app.persistence.db_config import get_db_url, parse_db_url, mask_db_config, ALLOWED_AI_TABLES


class MariaDBSaver(BaseCheckpointSaver[str]):
    """
    Production-grade persistent checkpoint saver for LangGraph backed by MariaDB/MySQL.
    
    Guarantees:
    1. Checkpoints and writes survive sidecar process restarts.
    2. Zero silent fallback to MemorySaver.
    3. Multi-tenant isolation by storing and tracking organization_id and user_id.
    4. Full support for WAITING_FOR_HUMAN, PENDING_REVIEW, and RESUME workflows.
    """

    def __init__(
        self,
        db_config: Optional[dict] = None,
        *,
        serde: Optional[SerializerProtocol] = None,
    ):
        super().__init__(serde=serde or JsonPlusSerializer())
        self.db_config = db_config or parse_db_url(get_db_url())
        self._ensure_tables()

    @contextmanager
    def _get_connection(self):
        conn = pymysql.connect(
            host=self.db_config["host"],
            port=self.db_config["port"],
            user=self.db_config["user"],
            password=self.db_config["password"],
            database=self.db_config["db"],
            autocommit=True,
            charset="utf8mb4",
            cursorclass=pymysql.cursors.Cursor,
        )
        try:
            yield conn
        finally:
            conn.close()

    def _ensure_tables(self):
        """Validates that persistent tables exist and are reachable; fails loudly if not."""
        with self._get_connection() as conn:
            with conn.cursor() as cur:
                cur.execute("""
                    CREATE TABLE IF NOT EXISTS ai_checkpoints (
                        thread_id VARCHAR(128) NOT NULL,
                        checkpoint_ns VARCHAR(64) NOT NULL DEFAULT '',
                        checkpoint_id VARCHAR(64) NOT NULL,
                        parent_checkpoint_id VARCHAR(64) NULL,
                        type VARCHAR(50) NOT NULL DEFAULT '',
                        checkpoint LONGBLOB NOT NULL,
                        metadata LONGBLOB NOT NULL,
                        organization_id INT NULL,
                        user_id INT NULL,
                        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                        PRIMARY KEY (thread_id, checkpoint_ns, checkpoint_id),
                        KEY idx_thread_ns (thread_id, checkpoint_ns, checkpoint_id DESC),
                        KEY idx_org_id (organization_id)
                    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
                """)
                cur.execute("""
                    CREATE TABLE IF NOT EXISTS ai_checkpoint_writes (
                        thread_id VARCHAR(128) NOT NULL,
                        checkpoint_ns VARCHAR(64) NOT NULL DEFAULT '',
                        checkpoint_id VARCHAR(64) NOT NULL,
                        task_id VARCHAR(64) NOT NULL,
                        idx INT NOT NULL,
                        channel VARCHAR(64) NOT NULL,
                        type VARCHAR(50) NOT NULL DEFAULT '',
                        value LONGBLOB NOT NULL,
                        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                        PRIMARY KEY (thread_id, checkpoint_ns, checkpoint_id, task_id, idx),
                        KEY idx_thread_ns_ckpt (thread_id, checkpoint_ns, checkpoint_id)
                    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
                """)

    def verify_storage(self) -> bool:
        """Startup healthcheck to verify MariaDB checkpoint storage is operational."""
        try:
            with self._get_connection() as conn:
                with conn.cursor() as cur:
                    cur.execute("SELECT COUNT(*) FROM ai_checkpoints")
                    _ = cur.fetchone()
            return True
        except Exception as e:
            raise RuntimeError(
                f"[CRITICAL] LangGraph persistent MariaDB checkpointer verification failed: {e}. "
                "Silent in-memory fallback is prohibited in production."
            ) from e

    def get_tuple(self, config: RunnableConfig) -> Optional[CheckpointTuple]:
        """Fetch saved checkpoint tuple for the given config, enforcing tenant scoping if org_id is provided."""
        thread_id = str(config["configurable"]["thread_id"])
        checkpoint_ns = str(config["configurable"].get("checkpoint_ns", ""))
        checkpoint_id = get_checkpoint_id(config)

        # Extract tenant context if provided in config
        cfg_meta = config.get("metadata", {}) or {}
        cfg_conf = config.get("configurable", {}) or {}
        raw_org = cfg_meta.get("org_id") or cfg_conf.get("org_id")
        try:
            req_org_id = int(raw_org) if raw_org is not None else None
        except (ValueError, TypeError):
            req_org_id = None

        with self._get_connection() as conn:
            with conn.cursor() as cur:
                if checkpoint_id:
                    cur.execute(
                        """
                        SELECT thread_id, checkpoint_id, parent_checkpoint_id, type, checkpoint, metadata, organization_id
                        FROM ai_checkpoints
                        WHERE thread_id = %s AND checkpoint_ns = %s AND checkpoint_id = %s
                        LIMIT 1
                        """,
                        (thread_id, checkpoint_ns, checkpoint_id),
                    )
                else:
                    cur.execute(
                        """
                        SELECT thread_id, checkpoint_id, parent_checkpoint_id, type, checkpoint, metadata, organization_id
                        FROM ai_checkpoints
                        WHERE thread_id = %s AND checkpoint_ns = %s
                        ORDER BY checkpoint_id DESC
                        LIMIT 1
                        """,
                        (thread_id, checkpoint_ns),
                    )
                row = cur.fetchone()
                if not row:
                    return None

                (
                    row_thread_id,
                    row_checkpoint_id,
                    parent_checkpoint_id,
                    cp_type,
                    cp_blob,
                    meta_blob,
                    row_org_id,
                ) = row

                # Tenant isolation enforcement: if caller specified an org_id, verify it matches
                if req_org_id is not None and row_org_id is not None and int(row_org_id) != req_org_id:
                    print(
                        f"[SECURITY] Checkpoint tenant isolation violation: thread {thread_id} belongs to org {row_org_id}, "
                        f"access denied for requested org {req_org_id}"
                    )
                    return None

                # Query pending writes
                cur.execute(
                    """
                    SELECT task_id, channel, type, value
                    FROM ai_checkpoint_writes
                    WHERE thread_id = %s AND checkpoint_ns = %s AND checkpoint_id = %s
                    ORDER BY task_id, idx
                    """,
                    (row_thread_id, checkpoint_ns, row_checkpoint_id),
                )
                writes = cur.fetchall()

        # Deserialize
        checkpoint = self.serde.loads_typed((cp_type, cp_blob))
        meta_dict = json.loads(meta_blob.decode("utf-8")) if meta_blob else {}
        pending_writes = [
            (task_id, channel, self.serde.loads_typed((w_type, w_val)))
            for task_id, channel, w_type, w_val in writes
        ]

        final_config: RunnableConfig = {
            "configurable": {
                "thread_id": row_thread_id,
                "checkpoint_ns": checkpoint_ns,
                "checkpoint_id": row_checkpoint_id,
            }
        }
        parent_config = (
            {
                "configurable": {
                    "thread_id": row_thread_id,
                    "checkpoint_ns": checkpoint_ns,
                    "checkpoint_id": parent_checkpoint_id,
                }
            }
            if parent_checkpoint_id
            else None
        )

        return CheckpointTuple(
            config=final_config,
            checkpoint=checkpoint,
            metadata=cast(CheckpointMetadata, meta_dict),
            parent_config=parent_config,
            pending_writes=pending_writes,
        )

    def put(
        self,
        config: RunnableConfig,
        checkpoint: Checkpoint,
        metadata: CheckpointMetadata,
        new_versions: ChannelVersions,
    ) -> RunnableConfig:
        """Saves checkpoint to MariaDB, enforcing tenant isolation against cross-org writes."""
        thread_id = str(config["configurable"]["thread_id"])
        checkpoint_ns = str(config["configurable"].get("checkpoint_ns", ""))
        checkpoint_id = checkpoint["id"]
        parent_checkpoint_id = config["configurable"].get("checkpoint_id")

        # Extract tenant context
        cfg_meta = config.get("metadata", {}) or {}
        cfg_conf = config.get("configurable", {}) or {}
        org_id = (
            cfg_meta.get("org_id")
            or cfg_conf.get("org_id")
            or metadata.get("org_id")
            or None
        )
        user_id = (
            cfg_meta.get("user_id")
            or cfg_conf.get("user_id")
            or metadata.get("user_id")
            or None
        )

        try:
            org_id = int(org_id) if org_id is not None else None
        except (ValueError, TypeError):
            org_id = None

        try:
            user_id = int(user_id) if user_id is not None else None
        except (ValueError, TypeError):
            user_id = None

        type_, serialized_checkpoint = self.serde.dumps_typed(checkpoint)
        serialized_metadata = json.dumps(
            get_checkpoint_metadata(config, metadata), ensure_ascii=False
        ).encode("utf-8", "ignore")

        with self._get_connection() as conn:
            with conn.cursor() as cur:
                # Prevent writing to an existing thread owned by a different organization
                if org_id is not None:
                    cur.execute(
                        "SELECT organization_id FROM ai_checkpoints WHERE thread_id = %s AND organization_id IS NOT NULL LIMIT 1",
                        (thread_id,),
                    )
                    existing_row = cur.fetchone()
                    if existing_row and existing_row[0] is not None and int(existing_row[0]) != org_id:
                        raise PermissionError(
                            f"Cross-organization checkpoint write rejected: thread {thread_id} belongs to org {existing_row[0]}, not {org_id}"
                        )

                cur.execute(
                    """
                    INSERT INTO ai_checkpoints (
                        thread_id, checkpoint_ns, checkpoint_id, parent_checkpoint_id,
                        type, checkpoint, metadata, organization_id, user_id
                    ) VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s)
                    ON DUPLICATE KEY UPDATE
                        parent_checkpoint_id = VALUES(parent_checkpoint_id),
                        type = VALUES(type),
                        checkpoint = VALUES(checkpoint),
                        metadata = VALUES(metadata),
                        organization_id = COALESCE(VALUES(organization_id), organization_id),
                        user_id = COALESCE(VALUES(user_id), user_id)
                    """,
                    (
                        thread_id,
                        checkpoint_ns,
                        checkpoint_id,
                        parent_checkpoint_id,
                        type_,
                        serialized_checkpoint,
                        serialized_metadata,
                        org_id,
                        user_id,
                    ),
                )

        return {
            "configurable": {
                "thread_id": thread_id,
                "checkpoint_ns": checkpoint_ns,
                "checkpoint_id": checkpoint_id,
            }
        }

    def put_writes(
        self,
        config: RunnableConfig,
        writes: Sequence[Tuple[str, Any]],
        task_id: str,
        task_path: str = "",
    ) -> None:
        """Saves intermediate node writes to MariaDB."""
        thread_id = str(config["configurable"]["thread_id"])
        checkpoint_ns = str(config["configurable"].get("checkpoint_ns", ""))
        checkpoint_id = str(config["configurable"]["checkpoint_id"])

        with self._get_connection() as conn:
            with conn.cursor() as cur:
                for idx, (channel, value) in enumerate(writes):
                    w_idx = WRITES_IDX_MAP.get(channel, idx)
                    w_type, w_blob = self.serde.dumps_typed(value)
                    cur.execute(
                        """
                        INSERT INTO ai_checkpoint_writes (
                            thread_id, checkpoint_ns, checkpoint_id, task_id, idx,
                            channel, type, value
                        ) VALUES (%s, %s, %s, %s, %s, %s, %s, %s)
                        ON DUPLICATE KEY UPDATE
                            channel = VALUES(channel),
                            type = VALUES(type),
                            value = VALUES(value)
                        """,
                        (
                            thread_id,
                            checkpoint_ns,
                            checkpoint_id,
                            task_id,
                            w_idx,
                            channel,
                            w_type,
                            w_blob,
                        ),
                    )

    def list(
        self,
        config: Optional[RunnableConfig],
        *,
        filter: Optional[dict[str, Any]] = None,
        before: Optional[RunnableConfig] = None,
        limit: Optional[int] = None,
    ) -> Iterator[CheckpointTuple]:
        """Lists checkpoints ordered newest first."""
        thread_id = str(config["configurable"]["thread_id"]) if config else None
        checkpoint_ns = (
            str(config["configurable"].get("checkpoint_ns", ""))
            if config and "checkpoint_ns" in config.get("configurable", {})
            else None
        )
        before_id = get_checkpoint_id(before) if before else None

        wheres = []
        params = []
        if thread_id:
            wheres.append("thread_id = %s")
            params.append(thread_id)
        if checkpoint_ns is not None:
            wheres.append("checkpoint_ns = %s")
            params.append(checkpoint_ns)
        if before_id:
            wheres.append("checkpoint_id < %s")
            params.append(before_id)

        cfg_meta = config.get("metadata", {}) or {} if config else {}
        cfg_conf = config.get("configurable", {}) or {} if config else {}
        raw_org = cfg_meta.get("org_id") or cfg_conf.get("org_id")
        try:
            req_org_id = int(raw_org) if raw_org is not None else None
        except (ValueError, TypeError):
            req_org_id = None

        if req_org_id is not None:
            wheres.append("(organization_id = %s OR organization_id IS NULL)")
            params.append(req_org_id)

        where_clause = f"WHERE {' AND '.join(wheres)}" if wheres else ""
        limit_clause = f"LIMIT {int(limit)}" if limit else ""

        query = f"""
            SELECT thread_id, checkpoint_ns, checkpoint_id, parent_checkpoint_id, type, checkpoint, metadata
            FROM ai_checkpoints
            {where_clause}
            ORDER BY checkpoint_id DESC
            {limit_clause}
        """

        with self._get_connection() as conn:
            with conn.cursor() as cur:
                cur.execute(query, tuple(params))
                rows = cur.fetchall()

                for (
                    t_id,
                    ns,
                    cp_id,
                    p_cp_id,
                    cp_type,
                    cp_blob,
                    meta_blob,
                ) in rows:
                    cur.execute(
                        """
                        SELECT task_id, channel, type, value
                        FROM ai_checkpoint_writes
                        WHERE thread_id = %s AND checkpoint_ns = %s AND checkpoint_id = %s
                        ORDER BY task_id, idx
                        """,
                        (t_id, ns, cp_id),
                    )
                    writes = cur.fetchall()

                    checkpoint = self.serde.loads_typed((cp_type, cp_blob))
                    meta_dict = (
                        json.loads(meta_blob.decode("utf-8")) if meta_blob else {}
                    )
                    pending_writes = [
                        (task_id, channel, self.serde.loads_typed((w_type, w_val)))
                        for task_id, channel, w_type, w_val in writes
                    ]

                    yield CheckpointTuple(
                        config={
                            "configurable": {
                                "thread_id": t_id,
                                "checkpoint_ns": ns,
                                "checkpoint_id": cp_id,
                            }
                        },
                        checkpoint=checkpoint,
                        metadata=cast(CheckpointMetadata, meta_dict),
                        parent_config=(
                            {
                                "configurable": {
                                    "thread_id": t_id,
                                    "checkpoint_ns": ns,
                                    "checkpoint_id": p_cp_id,
                                }
                            }
                            if p_cp_id
                            else None
                        ),
                        pending_writes=pending_writes,
                    )

    def delete_thread(self, thread_id: str, org_id: Optional[int] = None) -> None:
        """Deletes all checkpoints and writes for a given thread_id, optionally scoped by organization."""
        with self._get_connection() as conn:
            with conn.cursor() as cur:
                if org_id is not None:
                    cur.execute("DELETE FROM ai_checkpoints WHERE thread_id = %s AND organization_id = %s", (str(thread_id), org_id))
                    cur.execute("DELETE FROM ai_checkpoint_writes WHERE thread_id = %s", (str(thread_id),))
                else:
                    cur.execute("DELETE FROM ai_checkpoints WHERE thread_id = %s", (str(thread_id),))
                    cur.execute("DELETE FROM ai_checkpoint_writes WHERE thread_id = %s", (str(thread_id),))

    async def aget_tuple(self, config: RunnableConfig) -> Optional[CheckpointTuple]:
        return await asyncio.to_thread(self.get_tuple, config)

    async def alist(
        self,
        config: Optional[RunnableConfig],
        *,
        filter: Optional[dict[str, Any]] = None,
        before: Optional[RunnableConfig] = None,
        limit: Optional[int] = None,
    ) -> AsyncIterator[CheckpointTuple]:
        loop = asyncio.get_event_loop()
        items = await loop.run_in_executor(
            None, lambda: list(self.list(config, filter=filter, before=before, limit=limit))
        )
        for item in items:
            yield item

    async def aput(
        self,
        config: RunnableConfig,
        checkpoint: Checkpoint,
        metadata: CheckpointMetadata,
        new_versions: ChannelVersions,
    ) -> RunnableConfig:
        return await asyncio.to_thread(self.put, config, checkpoint, metadata, new_versions)

    async def aput_writes(
        self,
        config: RunnableConfig,
        writes: Sequence[Tuple[str, Any]],
        task_id: str,
        task_path: str = "",
    ) -> None:
        await asyncio.to_thread(self.put_writes, config, writes, task_id, task_path)

    async def adelete_thread(self, thread_id: str) -> None:
        await asyncio.to_thread(self.delete_thread, thread_id)

    def get_next_version(self, current: Optional[str], channel: None) -> str:
        if current is None:
            current_v = 0
        elif isinstance(current, int):
            current_v = current
        else:
            current_v = int(current.split(".")[0])
        next_v = current_v + 1
        next_h = random.random()
        return f"{next_v:032}.{next_h:016}"


# Singleton instance factory
_mariadb_saver_instance: Optional[MariaDBSaver] = None

def get_checkpointer() -> MariaDBSaver:
    """Returns singleton MariaDBSaver checkpointer. Fails loudly if storage is unreachable."""
    global _mariadb_saver_instance
    if _mariadb_saver_instance is None:
        _mariadb_saver_instance = MariaDBSaver()
        _mariadb_saver_instance.verify_storage()
        print("[AI Sidecar Persistence] Initialized persistent MariaDBSaver successfully.")
    return _mariadb_saver_instance

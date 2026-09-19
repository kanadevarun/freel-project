-- 082_ai_checkpoints_persistence.sql
-- Persistent checkpointer storage for LangGraph agent workflows across LogisticsHQ

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

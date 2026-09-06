-- +migrate Up

DROP TABLE IF EXISTS shipment_exceptions;
DROP TABLE IF EXISTS shipment_milestones;

CREATE TABLE shipment_milestones (
    id              BIGINT AUTO_INCREMENT PRIMARY KEY,
    shipment_id     BIGINT NOT NULL,
    milestone_code  VARCHAR(50) NOT NULL, -- BOOKED, DEPARTED, IN_TRANSIT, ARRIVED, DELIVERED
    description     VARCHAR(255),
    planned_date    DATETIME,
    actual_date     DATETIME,
    status          VARCHAR(30) NOT NULL DEFAULT 'PLANNED',
    location        VARCHAR(100),
    notes           TEXT,
    source_event_id VARCHAR(255),
    updated_at      DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    CONSTRAINT fk_sm_ship FOREIGN KEY (shipment_id) REFERENCES shipments(id) ON DELETE CASCADE,
    CONSTRAINT uq_milestone_shipment_code UNIQUE (shipment_id, milestone_code),
    CONSTRAINT chk_milestone_status CHECK (status IN ('PLANNED', 'COMPLETED'))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE shipment_exceptions (
    id              BIGINT AUTO_INCREMENT PRIMARY KEY,
    shipment_id     BIGINT NOT NULL,
    exception_type  VARCHAR(50) NOT NULL, -- ROLLOVER, DELAY, CUSTOMS_HOLD, PORT_CONGESTION, WEATHER
    severity        VARCHAR(20) NOT NULL, -- INFO, WARNING, CRITICAL
    title           VARCHAR(255) NOT NULL,
    description     TEXT,
    resolved        BOOLEAN NOT NULL DEFAULT FALSE,
    resolved_at     DATETIME,
    ai_summary      TEXT,
    source_event_id VARCHAR(255),
    created_at      DATETIME DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_se_ship FOREIGN KEY (shipment_id) REFERENCES shipments(id) ON DELETE CASCADE,
    CONSTRAINT chk_exception_type CHECK (exception_type IN ('ROLLOVER', 'DELAY', 'CUSTOMS_HOLD', 'PORT_CONGESTION', 'WEATHER')),
    CONSTRAINT chk_exception_severity CHECK (severity IN ('INFO', 'WARNING', 'CRITICAL'))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- +migrate Down

DROP TABLE IF EXISTS shipment_exceptions;
DROP TABLE IF EXISTS shipment_milestones;

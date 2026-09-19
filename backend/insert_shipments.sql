SET FOREIGN_KEY_CHECKS = 0;

INSERT INTO shipments (
    id, org_id, rfq_id, quote_id, booking_id, booking_number, mbl_number, hbl_number, 
    carrier_scac, vessel_name, voyage_number, origin_port, destination_port, 
    container_numbers, status, etd, eta, closure_status, created_at, updated_at
) VALUES
-- Shipment 101: In-transit normal workflow
(101, 2, 101, 101, 101, 'BK-2026-DEV-001', 'MAEU123456789', 'LH-HBL-2601',
 'MAEU', 'MAERSK MC-KINNEY MOLLER', '2601W', 'INNSA', 'NLRTM',
 '["MSKU7891234"]', 'IN_TRANSIT', '2026-08-20 10:00:00', '2026-09-15 18:00:00', 'ACTIVE', NOW(), NOW()),

-- Shipment 102: SHIPMENT WITH DELAYED ETA (planned: 2026-09-02, revised: 2026-09-12)
(102, 2, 102, 102, 101, 'BK-2026-DEV-002', 'MSCU987654321', 'LH-HBL-2602',
 'MSCU', 'MSC OSCAR', '2602W', 'INNSA', 'DEHAM',
 '["MSCU9876543"]', 'IN_TRANSIT', '2026-08-15 08:00:00', '2026-09-12 14:00:00', 'ACTIVE', NOW(), NOW()),

-- Shipment 103: SHIPMENT WITH CRITICAL EXCEPTION (Customs hold)
(103, 2, 101, 101, 101, 'BK-2026-DEV-003', 'CMDU543216789', 'LH-HBL-2603',
 'CMDU', 'CMA CGM ANTOINE', '2603W', 'INNSA', 'USNYC',
 '["CMAU5432109"]', 'CUSTOMS_HOLD', '2026-08-18 12:00:00', '2026-09-20 10:00:00', 'ACTIVE', NOW(), NOW())
ON DUPLICATE KEY UPDATE vessel_name = VALUES(vessel_name);

INSERT INTO shipment_milestones (
    id, shipment_id, milestone_code, description, planned_date, actual_date, status, location, notes
) VALUES
(101, 101, 'GATE_IN', 'Container gated in at terminal', '2026-08-18 08:00:00', '2026-08-18 09:30:00', 'COMPLETED', 'Nhava Sheva Port, Terminal 3', 'Clean gate-in inspection passed'),
(102, 101, 'LOADED', 'Container loaded on vessel', '2026-08-19 14:00:00', '2026-08-19 15:20:00', 'COMPLETED', 'Nhava Sheva Port, Berth 2', 'Stowed under deck bay 14'),
(103, 101, 'DEPARTED', 'Vessel departed origin port', '2026-08-20 10:00:00', '2026-08-20 11:15:00', 'COMPLETED', 'Nhava Sheva Port', 'Vessel on schedule'),
(104, 101, 'ARRIVAL', 'Vessel scheduled arrival', '2026-09-15 18:00:00', NULL, 'PLANNED', 'Rotterdam Port (NLRTM)', 'Estimated arrival on track'),

(105, 102, 'GATE_IN', 'Container gated in at terminal', '2026-08-14 10:00:00', '2026-08-14 11:00:00', 'COMPLETED', 'Nhava Sheva Port', 'Initial gate in completed'),
(106, 102, 'DELAY_NOTICE', 'Vessel departure delayed due to typhoon', '2026-08-15 08:00:00', '2026-08-18 20:00:00', 'COMPLETED', 'Arabian Sea / Colombo Transshipment', 'Carrier bulletin #26-089: 10-day weather delay. Revised ETA: 2026-09-12')
ON DUPLICATE KEY UPDATE description = VALUES(description);

SET FOREIGN_KEY_CHECKS = 1;

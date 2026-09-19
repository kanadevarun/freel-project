SET FOREIGN_KEY_CHECKS = 0;

UPDATE shipments 
SET org_id = 2, rfq_id = 101, quote_id = 101, booking_id = 101, booking_number = 'BK-2026-DEV-001',
    mbl_number = 'MAEU123456789', hbl_number = 'LH-HBL-2601', carrier_scac = 'MAEU',
    vessel_name = 'MAERSK MC-KINNEY MOLLER', voyage_number = '2601W', origin_port = 'INNSA', destination_port = 'NLRTM',
    container_numbers = '["MSKU7891234"]', status = 'IN_TRANSIT', etd = '2026-08-20 10:00:00', eta = '2026-09-15 18:00:00', closure_status = 'ACTIVE'
WHERE id = 101;

UPDATE shipments 
SET org_id = 2, rfq_id = 102, quote_id = 102, booking_id = 101, booking_number = 'BK-2026-DEV-002',
    mbl_number = 'MSCU987654321', hbl_number = 'LH-HBL-2602', carrier_scac = 'MSCU',
    vessel_name = 'MSC OSCAR', voyage_number = '2602W', origin_port = 'INNSA', destination_port = 'DEHAM',
    container_numbers = '["MSCU9876543"]', status = 'IN_TRANSIT', etd = '2026-08-15 08:00:00', eta = '2026-09-12 14:00:00', closure_status = 'ACTIVE'
WHERE id = 102;

UPDATE shipments 
SET org_id = 2, rfq_id = 101, quote_id = 101, booking_id = 101, booking_number = 'BK-2026-DEV-003',
    mbl_number = 'CMDU543216789', hbl_number = 'LH-HBL-2603', carrier_scac = 'CMDU',
    vessel_name = 'CMA CGM ANTOINE', voyage_number = '2603W', origin_port = 'INNSA', destination_port = 'USNYC',
    container_numbers = '["CMAU5432109"]', status = 'CUSTOMS_HOLD', etd = '2026-08-18 12:00:00', eta = '2026-09-20 10:00:00', closure_status = 'ACTIVE'
WHERE id = 103;

SET FOREIGN_KEY_CHECKS = 1;

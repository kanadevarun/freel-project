-- 079_carrier_bookings_integration.sql
-- Adds carrier booking integration columns to the bookings table.

ALTER TABLE bookings
    ADD COLUMN IF NOT EXISTS carrier_booking_reference VARCHAR(100) NULL AFTER carrier_scac,
    ADD COLUMN IF NOT EXISTS carrier_booking_status VARCHAR(50) NULL AFTER carrier_booking_reference,
    ADD COLUMN IF NOT EXISTS carrier_confirmation_reference VARCHAR(100) NULL AFTER carrier_booking_status,
    ADD COLUMN IF NOT EXISTS carrier_booking_error TEXT NULL AFTER carrier_confirmation_reference,
    ADD COLUMN IF NOT EXISTS carrier_booked_at DATETIME NULL AFTER carrier_booking_error;

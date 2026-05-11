BEGIN;

-- Remove name and seats from dining_tables
ALTER TABLE dining_tables DROP COLUMN IF EXISTS name;
ALTER TABLE dining_tables DROP COLUMN IF EXISTS seats;

-- Remove status from users and vendors
ALTER TABLE users DROP COLUMN IF EXISTS status;
ALTER TABLE vendors DROP COLUMN IF EXISTS status;

-- Remove status from dining_tables
ALTER TABLE dining_tables DROP COLUMN IF EXISTS status;

-- Change quantity and min_stock to double precision to avoid pgtype.Numeric issues in sqlc
ALTER TABLE inventory_items ALTER COLUMN quantity TYPE DOUBLE PRECISION;
ALTER TABLE inventory_items ALTER COLUMN min_stock TYPE DOUBLE PRECISION;
ALTER TABLE menu_items ALTER COLUMN price TYPE DOUBLE PRECISION;
ALTER TABLE bills ALTER COLUMN amount TYPE DOUBLE PRECISION;
ALTER TABLE ingredients ALTER COLUMN quantity TYPE DOUBLE PRECISION;

COMMIT;

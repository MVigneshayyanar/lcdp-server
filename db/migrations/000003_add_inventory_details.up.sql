ALTER TABLE inventory_items ADD COLUMN category TEXT DEFAULT 'Pantry';
ALTER TABLE inventory_items ADD COLUMN min_stock NUMERIC(12, 3) DEFAULT 10.0;

UPDATE inventory_items SET category = 'Pantry' WHERE category IS NULL;
ALTER TABLE inventory_items ALTER COLUMN category SET NOT NULL;

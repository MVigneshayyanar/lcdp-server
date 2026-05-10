ALTER TABLE dining_tables ADD COLUMN name TEXT;
ALTER TABLE dining_tables ADD COLUMN seats INTEGER DEFAULT 4;

UPDATE dining_tables SET name = 'Table ' || number WHERE name IS NULL;
ALTER TABLE dining_tables ALTER COLUMN name SET NOT NULL;

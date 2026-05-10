ALTER TABLE menu_items ADD COLUMN category TEXT DEFAULT 'mains';
ALTER TABLE menu_items ADD COLUMN description TEXT;

UPDATE menu_items SET category = 'mains' WHERE category IS NULL;
ALTER TABLE menu_items ALTER COLUMN category SET NOT NULL;

ALTER TABLE users ADD COLUMN status TEXT DEFAULT 'active';
ALTER TABLE vendors ADD COLUMN status TEXT DEFAULT 'active';

UPDATE users SET status = 'active' WHERE status IS NULL;
UPDATE vendors SET status = 'active' WHERE status IS NULL;

ALTER TABLE users ALTER COLUMN status SET NOT NULL;
ALTER TABLE vendors ALTER COLUMN status SET NOT NULL;

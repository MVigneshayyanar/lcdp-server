-- Add 'ready' and 'billed' to table_status enum
ALTER TYPE table_status ADD VALUE IF NOT EXISTS 'ready';
ALTER TYPE table_status ADD VALUE IF NOT EXISTS 'billed';

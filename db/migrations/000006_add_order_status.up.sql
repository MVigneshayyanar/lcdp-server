CREATE TYPE order_status AS ENUM ('new', 'preparing', 'ready', 'served');
ALTER TABLE orders ADD COLUMN status order_status NOT NULL DEFAULT 'new';

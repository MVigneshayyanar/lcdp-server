BEGIN;

CREATE TYPE user_role AS ENUM ('waiter', 'manager', 'admin');
CREATE TYPE table_status AS ENUM ('available', 'ordered', 'preparing', 'eating');
CREATE TYPE bill_status AS ENUM ('pending', 'paid', 'overdue');

CREATE TABLE users (
  id BIGSERIAL PRIMARY KEY,
  name TEXT NOT NULL,
  role user_role NOT NULL,
  phone TEXT NOT NULL UNIQUE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE dining_tables (
  id BIGSERIAL PRIMARY KEY,
  number INTEGER NOT NULL UNIQUE,
  status table_status NOT NULL DEFAULT 'available',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE menu_items (
  id BIGSERIAL PRIMARY KEY,
  name TEXT NOT NULL UNIQUE,
  price NUMERIC(12, 2) NOT NULL,
  is_available BOOLEAN NOT NULL DEFAULT true,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE vendors (
  id BIGSERIAL PRIMARY KEY,
  name TEXT NOT NULL,
  address TEXT NOT NULL,
  email TEXT NOT NULL UNIQUE,
  phone TEXT NOT NULL UNIQUE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE inventory_items (
  id BIGSERIAL PRIMARY KEY,
  name TEXT NOT NULL UNIQUE,
  quantity NUMERIC(12, 3) NOT NULL,
  unit TEXT NOT NULL,
  vendor_id BIGINT NOT NULL REFERENCES vendors(id) ON DELETE RESTRICT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE bills (
  id BIGSERIAL PRIMARY KEY,
  vendor_id BIGINT NOT NULL REFERENCES vendors(id) ON DELETE RESTRICT,
  txn_id TEXT NOT NULL UNIQUE,
  amount NUMERIC(12, 2) NOT NULL,
  due_date DATE NOT NULL,
  status bill_status NOT NULL DEFAULT 'pending',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE orders (
  id BIGSERIAL PRIMARY KEY,
  menu_item_id BIGINT NOT NULL REFERENCES menu_items(id) ON DELETE RESTRICT,
  quantity INTEGER NOT NULL,
  table_id BIGINT NOT NULL REFERENCES dining_tables(id) ON DELETE RESTRICT,
  ordered_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE ingredients (
  id BIGSERIAL PRIMARY KEY,
  menu_item_id BIGINT NOT NULL REFERENCES menu_items(id) ON DELETE CASCADE,
  inventory_item_id BIGINT NOT NULL REFERENCES inventory_items(id) ON DELETE RESTRICT,
  quantity NUMERIC(12, 3) NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (menu_item_id, inventory_item_id)
);

CREATE TABLE otp_codes (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  code_hash TEXT NOT NULL,
  expires_at TIMESTAMPTZ NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  consumed_at TIMESTAMPTZ
);

CREATE TABLE sessions (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  token_hash TEXT NOT NULL UNIQUE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  expires_at TIMESTAMPTZ NOT NULL,
  last_seen_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  revoked_at TIMESTAMPTZ
);

CREATE INDEX idx_sessions_user_id ON sessions(user_id);
CREATE INDEX idx_sessions_expires_at ON sessions(expires_at);
CREATE INDEX idx_otp_codes_user_id ON otp_codes(user_id);
CREATE INDEX idx_otp_codes_expires_at ON otp_codes(expires_at);

COMMIT;
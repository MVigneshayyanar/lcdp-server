BEGIN;

CREATE TYPE alert_severity AS ENUM ('info', 'warning', 'critical');

CREATE TABLE alerts (
    id BIGSERIAL PRIMARY KEY,
    type TEXT NOT NULL,
    title TEXT NOT NULL,
    description TEXT NOT NULL,
    severity alert_severity NOT NULL DEFAULT 'info',
    is_read BOOLEAN NOT NULL DEFAULT false,
    metadata JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

COMMIT;

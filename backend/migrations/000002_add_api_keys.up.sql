CREATE TABLE api_keys (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    key_hash   TEXT        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

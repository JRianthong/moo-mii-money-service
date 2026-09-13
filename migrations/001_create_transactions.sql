CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS transactions (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id text NOT NULL,
    type text NOT NULL CHECK (type IN ('income', 'expense')),
    amount_cents bigint NOT NULL CHECK (amount_cents > 0),
    currency text NOT NULL DEFAULT 'THB',
    category_code text NOT NULL DEFAULT 'other',
    category_name text NOT NULL DEFAULT 'อื่น ๆ',
    note text,
    source_system text NOT NULL DEFAULT 'line',
    source_message_id text NOT NULL,
    occurred_at timestamptz NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_transactions_user_occurred
    ON transactions (user_id, occurred_at DESC);

CREATE INDEX IF NOT EXISTS idx_transactions_user_type_occurred
    ON transactions (user_id, type, occurred_at DESC);

CREATE INDEX IF NOT EXISTS idx_transactions_user_category_occurred
    ON transactions (user_id, category_code, occurred_at DESC);

CREATE UNIQUE INDEX IF NOT EXISTS idx_transactions_source_message
    ON transactions (source_system, source_message_id);

ALTER TABLE transactions ENABLE ROW LEVEL SECURITY;

COMMENT ON TABLE transactions IS
    'Server-side LINE income and expense records. RLS is enabled; public API access is intentionally not granted.';

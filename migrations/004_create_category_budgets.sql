CREATE TABLE IF NOT EXISTS public.category_budgets (
    user_id text NOT NULL,
    category_code text NOT NULL,
    category_name text NOT NULL,
    monthly_amount_cents bigint NOT NULL CHECK (monthly_amount_cents > 0),
    show_daily boolean NOT NULL DEFAULT false,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, category_code)
);

ALTER TABLE public.category_budgets ENABLE ROW LEVEL SECURITY;

REVOKE ALL ON TABLE public.category_budgets FROM anon;
REVOKE ALL ON TABLE public.category_budgets FROM authenticated;

COMMENT ON TABLE public.category_budgets IS
    'Per-user monthly expense budgets and daily overview preferences.';

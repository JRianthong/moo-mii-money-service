CREATE TABLE IF NOT EXISTS public.user_settings (
    user_id text PRIMARY KEY,
    billing_cycle_day smallint NOT NULL DEFAULT 1
        CHECK (billing_cycle_day BETWEEN 1 AND 31),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

ALTER TABLE public.user_settings ENABLE ROW LEVEL SECURITY;

REVOKE ALL ON TABLE public.user_settings FROM anon;
REVOKE ALL ON TABLE public.user_settings FROM authenticated;

COMMENT ON TABLE public.user_settings IS
    'Private per-user preferences for the LINE money tracking service.';

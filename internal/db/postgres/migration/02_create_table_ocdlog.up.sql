CREATE TABLE IF NOT EXISTS ocdlog(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id VARCHAR(128) REFERENCES account(id) ON DELETE CASCADE NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP,
    ruminate_minutes INTEGER NOT NULL DEFAULT 0,
    anxiety_level INTEGER NOT NULL DEFAULT 0,
    notes TEXT
);
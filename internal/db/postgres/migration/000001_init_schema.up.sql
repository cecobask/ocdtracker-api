CREATE TABLE IF NOT EXISTS account(
    id VARCHAR(128) PRIMARY KEY,
    email VARCHAR(90) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP,
    display_name VARCHAR(128),
    wake_time VARCHAR(5) NOT NULL DEFAULT '09:00',
    sleep_time VARCHAR(5) NOT NULL DEFAULT '23:00',
    notification_interval INTEGER NOT NULL DEFAULT 3
);

CREATE TABLE IF NOT EXISTS ocdlog(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id VARCHAR(128) REFERENCES account(id) ON DELETE CASCADE NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP,
    ruminate_minutes INTEGER NOT NULL DEFAULT 0,
    anxiety_level INTEGER NOT NULL DEFAULT 0,
    notes TEXT
);

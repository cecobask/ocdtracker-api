CREATE TABLE IF NOT EXISTS account(
    id VARCHAR(128) PRIMARY KEY,
    email VARCHAR(90) UNIQUE NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP,
    display_name VARCHAR(128),
    wake_time VARCHAR(5) NOT NULL DEFAULT '09:00',
    sleep_time VARCHAR(5) NOT NULL DEFAULT '23:00',
    notification_interval INTEGER NOT NULL DEFAULT 3,
    photo_url TEXT
);
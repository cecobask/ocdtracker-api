CREATE TABLE IF NOT EXISTS account(
    id VARCHAR(128) PRIMARY KEY,
    email VARCHAR(90) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    display_name VARCHAR(128) NOT NULL,
    wake_time TIME WITHOUT TIME ZONE NOT NULL,
    sleep_time TIME WITHOUT TIME ZONE NOT NULL,
    notification_interval INTERVAL DEFAULT '3 hours'
);

CREATE TABLE IF NOT EXISTS ocdlog(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id VARCHAR(128) REFERENCES account(id) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP,
    ruminate_duration INTERVAL DEFAULT '0 minutes',
    anxiety_level INTEGER DEFAULT 0,
    notes TEXT,
    CONSTRAINT valid_anxiety_level CHECK (anxiety_level BETWEEN 0 AND 10)
);


-- INSERT INTO account(display_name, wake_time, sleep_time, notification_interval) VALUES ('John', '07:00', '22:00', '3 hours');
-- INSERT INTO ocdlog(account_id, anxiety_level, ruminate_duration) VALUES ('jEazVdPDhqec0tnEOG7vM5wbDyU2', 3, '39 minutes');
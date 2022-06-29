CREATE TABLE IF NOT EXISTS account(
    id VARCHAR(128) PRIMARY KEY DEFAULT gen_random_uuid()::VARCHAR,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    first_name VARCHAR(80) NOT NULL,
    wake_time TIME WITHOUT TIME ZONE NOT NULL,
    sleep_time TIME WITHOUT TIME ZONE NOT NULL,
    notification_interval INTERVAL DEFAULT '3 hours' NOT NULL
);

CREATE TABLE IF NOT EXISTS ocd_log(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    ruminate_duration INTERVAL DEFAULT '0 minutes' NOT NULL,
    anxiety_level INTEGER DEFAULT 0 NOT NULL,
    account_id VARCHAR(128) REFERENCES account(id) NOT NULL,
    notes TEXT,
    CONSTRAINT valid_anxiety_level CHECK (anxiety_level BETWEEN 0 AND 10)
);


-- INSERT INTO account(first_name, wake_time, sleep_time, notification_interval) VALUES ('John', '07:00', '22:00', '3 hours');
-- INSERT INTO ocd_log(account_id, anxiety_level, ruminate_duration) VALUES ('jEazVdPDhqec0tnEOG7vM5wbDyU2', 3, '39 minutes');
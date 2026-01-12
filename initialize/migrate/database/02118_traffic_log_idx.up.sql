ALTER TABLE traffic_log ADD INDEX IF NOT EXISTS idx_timestamp (timestamp);


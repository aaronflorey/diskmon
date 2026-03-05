CREATE TABLE IF NOT EXISTS drives (
    id BIGINT PRIMARY KEY,
    device TEXT NOT NULL UNIQUE,
    model TEXT,
    serial TEXT,
    wwn TEXT,
    first_seen_at TIMESTAMP NOT NULL,
    last_seen_at TIMESTAMP NOT NULL
);

CREATE SEQUENCE IF NOT EXISTS seq_drives START 1;

CREATE TABLE IF NOT EXISTS smart_samples (
    id BIGINT PRIMARY KEY,
    drive_id BIGINT NOT NULL,
    collected_at TIMESTAMP NOT NULL,
    temperature INTEGER,
    power_on_hours BIGINT,
    reallocated_sectors BIGINT,
    pending_sectors BIGINT,
    uncorrectable_sectors BIGINT,
    wear_level BIGINT,
    raw_json JSON,
    FOREIGN KEY (drive_id) REFERENCES drives(id)
);

CREATE SEQUENCE IF NOT EXISTS seq_samples START 1;

CREATE TABLE IF NOT EXISTS smart_attributes (
    sample_id BIGINT NOT NULL,
    attribute_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    value INTEGER,
    worst INTEGER,
    threshold INTEGER,
    raw TEXT,
    FOREIGN KEY (sample_id) REFERENCES smart_samples(id)
);

CREATE TABLE IF NOT EXISTS drive_health (
    drive_id BIGINT NOT NULL,
    sample_id BIGINT NOT NULL,
    status TEXT NOT NULL,
    score INTEGER NOT NULL,
    reasons TEXT,
    FOREIGN KEY (drive_id) REFERENCES drives(id),
    FOREIGN KEY (sample_id) REFERENCES smart_samples(id)
);

CREATE TABLE IF NOT EXISTS smart_test_runs (
    id BIGINT PRIMARY KEY,
    drive_id BIGINT NOT NULL,
    test_type TEXT NOT NULL,
    scheduled_at TIMESTAMP NOT NULL,
    started_at TIMESTAMP NOT NULL,
    finished_at TIMESTAMP NOT NULL,
    status TEXT NOT NULL,
    message TEXT,
    FOREIGN KEY (drive_id) REFERENCES drives(id)
);

CREATE SEQUENCE IF NOT EXISTS seq_smart_test_runs START 1;

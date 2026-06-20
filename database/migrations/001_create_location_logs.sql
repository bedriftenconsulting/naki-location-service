CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS location_logs (
    id         UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    nurse_id   UUID NOT NULL,
    visit_id   VARCHAR(100),
    latitude   DECIMAL(10,7) NOT NULL,
    longitude  DECIMAL(10,7) NOT NULL,
    heading    DECIMAL(6,2) DEFAULT 0,
    speed      DECIMAL(6,2) DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_location_nurse_id ON location_logs(nurse_id);
CREATE INDEX IF NOT EXISTS idx_location_visit_id ON location_logs(visit_id);
CREATE INDEX IF NOT EXISTS idx_location_created_at ON location_logs(created_at);
CREATE INDEX IF NOT EXISTS idx_location_nurse_time ON location_logs(nurse_id, created_at DESC);

-- Migration 003: Event Audit Tables
-- Create tables for tracking published and consumed events

-- Published Events Audit Table
CREATE TABLE IF NOT EXISTS pmsn.published_events (
    id VARCHAR PRIMARY KEY,
    event_type VARCHAR NOT NULL,
    payload JSONB NOT NULL,
    status VARCHAR NOT NULL, -- 'pending', 'published', 'failed'
    error_message TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_published_events_status ON pmsn.published_events(status);
CREATE INDEX IF NOT EXISTS idx_published_events_event_type ON pmsn.published_events(event_type);
CREATE INDEX IF NOT EXISTS idx_published_events_created_at ON pmsn.published_events(created_at);

-- Consumed Events Audit Table
CREATE TABLE IF NOT EXISTS pmsn.consumed_events (
    id VARCHAR PRIMARY KEY,
    event_type VARCHAR NOT NULL,
    payload JSONB NOT NULL,
    status VARCHAR NOT NULL, -- 'processing', 'completed', 'failed'
    error_message TEXT,
    retry_count INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_consumed_events_status ON pmsn.consumed_events(status);
CREATE INDEX IF NOT EXISTS idx_consumed_events_event_type ON pmsn.consumed_events(event_type);
CREATE INDEX IF NOT EXISTS idx_consumed_events_created_at ON pmsn.consumed_events(created_at);

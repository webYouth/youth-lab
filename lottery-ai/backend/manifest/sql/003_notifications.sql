-- App 消息中心与推送设备
CREATE TABLE IF NOT EXISTS app_notifications (
    id         BIGSERIAL PRIMARY KEY,
    type       VARCHAR(32) NOT NULL,
    title      VARCHAR(256) NOT NULL,
    body       TEXT NOT NULL,
    payload    JSONB NOT NULL DEFAULT '{}',
    read       BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_app_notifications_created
    ON app_notifications (created_at DESC);

CREATE INDEX IF NOT EXISTS idx_app_notifications_unread
    ON app_notifications (read, created_at DESC);

CREATE TABLE IF NOT EXISTS push_devices (
    id         BIGSERIAL PRIMARY KEY,
    token      TEXT NOT NULL UNIQUE,
    platform   VARCHAR(32) NOT NULL DEFAULT 'android',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

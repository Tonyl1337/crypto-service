CREATE TABLE telegram_subscriptions (
    id BIGSERIAL PRIMARY KEY,
    chat_id BIGINT NOT NULL UNIQUE,
    enabled BOOLEAN NOT NULL DEFAULT FALSE,
    interval_minutes INTEGER NOT NULL DEFAULT 5,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_subscriptions_chat_id
    ON telegram_subscriptions(chat_id);

CREATE INDEX idx_subscriptions_enabled
    ON telegram_subscriptions(enabled);
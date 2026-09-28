DROP TABLE IF EXISTS telegram_subscriptions;

CREATE TABLE telegram_subscriptions (
    id BIGSERIAL PRIMARY KEY,
    chat_id BIGINT NOT NULL,
    symbol VARCHAR(20) NOT NULL,
    coingecko_id VARCHAR(255) NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    interval_minutes INTEGER NOT NULL DEFAULT 5,
    last_sent_at TIMESTAMP NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),

    CONSTRAINT telegram_subscriptions_chat_coin_unique
        UNIQUE (chat_id, coingecko_id),

    CONSTRAINT telegram_subscriptions_interval_positive
        CHECK (interval_minutes > 0)
);

CREATE INDEX idx_subscriptions_enabled
    ON telegram_subscriptions(enabled);

CREATE INDEX idx_subscriptions_coingecko_id
    ON telegram_subscriptions(coingecko_id);
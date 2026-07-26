CREATE TABLE rates (
    id BIGSERIAL PRIMARY KEY,
    symbol VARCHAR(10) NOT NULL,
    price NUMERIC(20,8) NOT NULL,
    change_24h NUMERIC(10,4) NOT NULL,
    day_low NUMERIC(20,8) NOT NULL,
    day_high NUMERIC(20,8) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_rates_symbol ON rates(symbol);
CREATE INDEX idx_rates_created_at ON rates(created_at);
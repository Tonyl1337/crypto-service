ALTER TABLE rates
ADD COLUMN coingecko_id VARCHAR(255);

UPDATE rates
SET coingecko_id = CASE
    WHEN symbol = 'BTC' THEN 'bitcoin'
    WHEN symbol = 'ETH' THEN 'ethereum'
END;

DELETE FROM rates
WHERE coingecko_id IS NULL;

ALTER TABLE rates
ALTER COLUMN coingecko_id SET NOT NULL;

ALTER TABLE rates
ALTER COLUMN symbol TYPE VARCHAR(32);

CREATE INDEX idx_rates_coingecko_id
ON rates(coingecko_id);
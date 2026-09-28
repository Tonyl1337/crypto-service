DROP INDEX IF EXISTS idx_rates_coingecko_id;

ALTER TABLE rates
DROP COLUMN coingecko_id;

ALTER TABLE rates
ALTER COLUMN symbol TYPE VARCHAR(10);
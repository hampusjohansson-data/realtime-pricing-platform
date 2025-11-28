CREATE TABLE IF NOT EXISTS price_ticks_enriched (
    id BIGSERIAL PRIMARY KEY,
    symbol TEXT NOT NULL,
    price NUMERIC(18,8) NOT NULL,
    volume NUMERIC(18,8) NOT NULL,
    ts TIMESTAMPTZ NOT NULL,
    ma_1m NUMERIC(18,8),
    ma_5m NUMERIC(18,8),
    vol_1m NUMERIC(18,8),
    is_anomaly BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE INDEX IF NOT EXISTS idx_price_ticks_enriched_symbol_ts
    ON price_ticks_enriched(symbol, ts DESC);



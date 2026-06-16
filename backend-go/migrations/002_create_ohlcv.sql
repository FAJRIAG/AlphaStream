-- Migration: 002_create_ohlcv.sql
-- Creates the OHLCV candlestick table for AlphaStream.
-- Run AFTER 001_create_stocks.sql.

CREATE TABLE IF NOT EXISTS `ohlcv` (
    `id`        BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `stock_id`  BIGINT UNSIGNED NOT NULL,
    `symbol`    VARCHAR(20)     NOT NULL COMMENT 'Denormalized for fast query without JOIN',
    `timestamp` DATETIME(3)     NOT NULL COMMENT 'Candle open time (millisecond precision)',
    `open`      DECIMAL(18, 4)  NOT NULL,
    `high`      DECIMAL(18, 4)  NOT NULL,
    `low`       DECIMAL(18, 4)  NOT NULL,
    `close`     DECIMAL(18, 4)  NOT NULL,
    `volume`    BIGINT          NOT NULL DEFAULT 0,
    `timeframe` VARCHAR(10)     NOT NULL DEFAULT '1m' COMMENT '1m, 5m, 15m, 1h, 1d',
    `created_at` DATETIME       NOT NULL DEFAULT CURRENT_TIMESTAMP,

    PRIMARY KEY (`id`),

    -- Unique constraint prevents duplicate candles for same symbol+timestamp+timeframe.
    -- Works with INSERT IGNORE in stock_repo.SaveOHLCV.
    UNIQUE KEY `uq_ohlcv_symbol_ts_tf` (`symbol`, `timestamp`, `timeframe`),

    -- Index for fast chronological queries by symbol + timeframe.
    KEY `idx_ohlcv_symbol_tf_ts` (`symbol`, `timeframe`, `timestamp`),

    CONSTRAINT `fk_ohlcv_stock_id`
        FOREIGN KEY (`stock_id`) REFERENCES `stocks` (`id`)
        ON DELETE CASCADE ON UPDATE CASCADE

) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4
  COLLATE=utf8mb4_unicode_ci
  COMMENT='OHLCV candlestick price history'
  -- Partition by symbol for large-scale deployments (optional, uncomment if needed):
  -- PARTITION BY KEY(symbol) PARTITIONS 8
  ;

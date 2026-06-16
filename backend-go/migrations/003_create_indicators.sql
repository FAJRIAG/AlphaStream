-- Migration: 003_create_indicators.sql
-- Creates the technical indicators table for AlphaStream.
-- Run AFTER 001_create_stocks.sql.

CREATE TABLE IF NOT EXISTS `indicators` (
    `id`             BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `stock_id`       BIGINT UNSIGNED NOT NULL,
    `symbol`         VARCHAR(20)     NOT NULL,
    `timestamp`      DATETIME(3)     NOT NULL COMMENT 'Matches the OHLCV candle timestamp',
    `ma_20`          DECIMAL(18, 4)  NULL     COMMENT 'Simple MA over 20 periods',
    `ma_50`          DECIMAL(18, 4)  NULL     COMMENT 'Simple MA over 50 periods',
    `rsi_14`         DECIMAL(8, 4)   NULL     COMMENT 'RSI over 14 periods (0-100)',
    `is_golden_cross` TINYINT(1)     NULL     COMMENT '1 if MA20 crossed above MA50',
    `is_death_cross`  TINYINT(1)     NULL     COMMENT '1 if MA20 crossed below MA50',
    `atr_14`         DECIMAL(18, 4)  NULL     COMMENT 'Average True Range over 14 periods',
    `created_at`     DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,

    PRIMARY KEY (`id`),

    -- Upsert key: one indicator record per symbol per timestamp.
    -- Works with ON DUPLICATE KEY UPDATE in indicator_repo.SaveIndicators.
    UNIQUE KEY `uq_indicators_symbol_ts` (`symbol`, `timestamp`),

    -- Fast lookup: latest indicator for a symbol.
    KEY `idx_indicators_symbol_ts_desc` (`symbol`, `timestamp`),

    CONSTRAINT `fk_indicators_stock_id`
        FOREIGN KEY (`stock_id`) REFERENCES `stocks` (`id`)
        ON DELETE CASCADE ON UPDATE CASCADE

) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4
  COLLATE=utf8mb4_unicode_ci
  COMMENT='Computed technical indicators per symbol per candle';

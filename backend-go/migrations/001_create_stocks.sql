-- Migration: 001_create_stocks.sql
-- Creates the stocks master table for AlphaStream.
-- Run this FIRST before other migrations.

CREATE TABLE IF NOT EXISTS `stocks` (
    `id`         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `symbol`     VARCHAR(20)     NOT NULL COMMENT 'Ticker symbol, e.g. BBCA, TLKM',
    `name`       VARCHAR(200)    NOT NULL COMMENT 'Full company name',
    `exchange`   VARCHAR(50)     NOT NULL DEFAULT 'IDX' COMMENT 'Exchange, e.g. IDX, NASDAQ',
    `currency`   VARCHAR(10)     NOT NULL DEFAULT 'IDR',
    `is_active`  TINYINT(1)      NOT NULL DEFAULT 1,
    `created_at` DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    PRIMARY KEY (`id`),
    UNIQUE KEY `uq_stocks_symbol` (`symbol`),
    KEY `idx_stocks_is_active` (`is_active`)

) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4
  COLLATE=utf8mb4_unicode_ci
  COMMENT='Master list of tracked stocks/assets';

-- ─── Seed Data ────────────────────────────────────────────────────────────────
-- Insert the 5 default IDX blue-chip stocks used by the simulator.
-- INSERT IGNORE to make migration idempotent.

INSERT IGNORE INTO `stocks` (`id`, `symbol`, `name`, `exchange`, `currency`, `is_active`) VALUES
(1, 'BBCA',  'Bank Central Asia Tbk',          'IDX', 'IDR', 1),
(2, 'TLKM',  'Telkom Indonesia Tbk',            'IDX', 'IDR', 1),
(3, 'GOTO',  'GoTo Gojek Tokopedia Tbk',        'IDX', 'IDR', 1),
(4, 'BBRI',  'Bank Rakyat Indonesia Tbk',       'IDX', 'IDR', 1),
(5, 'ASII',  'Astra International Tbk',         'IDX', 'IDR', 1);

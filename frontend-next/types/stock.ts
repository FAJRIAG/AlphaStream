// TypeScript interfaces — single source of truth for all WebSocket/API data shapes.
// Mirrors the Go domain/entity structs exactly. Update both when schema changes.

// ─── WebSocket Message Types ──────────────────────────────────────────────────

export type WsMessageType =
  | 'OHLCV_UPDATE'
  | 'TICKER_UPDATE'
  | 'INDICATOR_UPDATE'
  | 'PREDICTION_UPDATE'
  | 'SUBSCRIBE'
  | 'UNSUBSCRIBE'
  | 'PING'
  | 'PONG'
  | 'ERROR';

export interface WsMessage<T = unknown> {
  type: WsMessageType;
  symbol: string;
  timestamp: string; // ISO 8601 UTC
  payload: T;
}

// ─── Stock ────────────────────────────────────────────────────────────────────

export interface Stock {
  id: number;
  symbol: string;
  name: string;
  exchange: string;
  currency: string;
  is_active: boolean;
  created_at: string;
  updated_at: string;
  price?: number;
  change?: number;
  change_percent?: number;
  volume?: number;
}

// ─── OHLCV ───────────────────────────────────────────────────────────────────

export interface OHLCV {
  id: number;
  stock_id: number;
  symbol: string;
  timestamp: string; // ISO 8601
  open: number;
  high: number;
  low: number;
  close: number;
  volume: number;
  timeframe: string;
}

// ─── Ticker ───────────────────────────────────────────────────────────────────

export interface Ticker {
  symbol: string;
  price: number;
  change: number;
  change_percent: number;
  volume: number;
  timestamp: string;
}

// ─── Technical Indicators ────────────────────────────────────────────────────

export interface TechnicalIndicators {
  id: number;
  stock_id: number;
  symbol: string;
  timestamp: string;
  ma_20: number | null;
  ma_50: number | null;
  rsi_14: number | null;
  is_golden_cross: boolean | null;
  is_death_cross: boolean | null;
  atr_14: number | null;
  created_at: string;
}

// ─── Prediction ───────────────────────────────────────────────────────────────

export type TrendDirection = 'BULLISH' | 'BEARISH' | 'NEUTRAL';
export type RSIZone = 'OVERBOUGHT' | 'NEUTRAL' | 'OVERSOLD';

export interface PredictionSignals {
  ma_20: number;
  ma_50: number;
  rsi_14: number;
  atr_14: number;
  is_golden_cross: boolean;
  is_death_cross: boolean;
  rsi_zone: RSIZone;
}

export interface PredictionResult {
  symbol: string;
  timestamp: string;
  direction: TrendDirection;
  probability: number;       // 0.0 – 1.0
  target_price_up: number;
  target_price_down: number;
  current_price: number;
  signals: PredictionSignals;
}

// ─── API Response Envelope ────────────────────────────────────────────────────

export interface ApiResponse<T> {
  success: boolean;
  data?: T;
  error?: string;
}

// ─── Chart Helpers ────────────────────────────────────────────────────────────

// Matches TradingView lightweight-charts CandlestickData shape
export interface ChartCandle {
  time: number;   // Unix timestamp (seconds)
  open: number;
  high: number;
  low: number;
  close: number;
}

// ─── Broker Summary (Bandarmology) ──────────────────────────────────────────────

export interface BrokerTrade {
  broker_code: string;
  broker_name: string;
  volume: number;
  avg_price: number;
  value: number;
  nationality: 'L' | 'F';
}

export interface BrokerSummary {
  symbol: string;
  timestamp: string;
  buyers: BrokerTrade[];
  sellers: BrokerTrade[];
  net_status: string;
  foreign_buy_pct: number;
}


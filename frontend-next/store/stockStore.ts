// Zustand global store for AlphaStream real-time market data.
// All state mutations happen via typed actions — no direct set() outside the store.
import { create } from 'zustand';
import type {
  Stock,
  OHLCV,
  Ticker,
  TechnicalIndicators,
  PredictionResult,
} from '@/types/stock';

// ─── State Shape ──────────────────────────────────────────────────────────────

interface StockState {
  // Master list of all active stocks
  stocks: Stock[];
  // Currently selected/viewed symbol
  activeSymbol: string;
  // Live OHLCV candles per symbol (last 500)
  candles: Record<string, OHLCV[]>;
  // Latest live ticker per symbol
  tickers: Record<string, Ticker>;
  // Latest indicators per symbol
  indicators: Record<string, TechnicalIndicators>;
  // Latest prediction per symbol
  predictions: Record<string, PredictionResult>;
  // WebSocket connection status
  wsStatus: 'connecting' | 'connected' | 'disconnected' | 'error';

  // ── Actions ───────────────────────────────────────────────────────────
  setStocks: (stocks: Stock[]) => void;
  setActiveSymbol: (symbol: string) => void;
  pushCandle: (symbol: string, candle: OHLCV) => void;
  setCandles: (symbol: string, candles: OHLCV[]) => void;
  setTicker: (symbol: string, ticker: Ticker) => void;
  setIndicators: (symbol: string, ind: TechnicalIndicators) => void;
  setPrediction: (symbol: string, pred: PredictionResult) => void;
  setWsStatus: (status: StockState['wsStatus']) => void;
}

// ─── Max candles to keep in memory per symbol ─────────────────────────────────
const MAX_CANDLES = 500;

// ─── Store ────────────────────────────────────────────────────────────────────

export const useStockStore = create<StockState>((set) => ({
  stocks: [],
  activeSymbol: 'IHSG',
  candles: {},
  tickers: {},
  indicators: {},
  predictions: {},
  wsStatus: 'disconnected',

  setStocks: (stocks) =>
    set((state) => {
      const tickers = { ...state.tickers };
      stocks.forEach((s) => {
        if (s.price != null) {
          tickers[s.symbol] = {
            symbol: s.symbol,
            price: s.price,
            change: s.change ?? 0,
            change_percent: s.change_percent ?? 0,
            volume: s.volume ?? 0,
            timestamp: s.updated_at || new Date().toISOString(),
          };
        }
      });
      return { stocks, tickers };
    }),

  setActiveSymbol: (symbol) => set({ activeSymbol: symbol }),

  pushCandle: (symbol, candle) =>
    set((state) => {
      const existing = state.candles[symbol] ?? [];
      let updated: OHLCV[];
      if (existing.length > 0 && existing[existing.length - 1].timestamp === candle.timestamp) {
        // Update the last candle
        updated = [...existing.slice(0, -1), candle];
      } else {
        // Append new candle and cap at MAX_CANDLES (FIFO trim)
        updated =
          existing.length >= MAX_CANDLES
            ? [...existing.slice(1), candle]
            : [...existing, candle];
      }
      return { candles: { ...state.candles, [symbol]: updated } };
    }),

  setCandles: (symbol, candles) =>
    set((state) => ({
      candles: { ...state.candles, [symbol]: (candles || []).slice(-MAX_CANDLES) },
    })),

  setTicker: (symbol, ticker) =>
    set((state) => ({
      tickers: { ...state.tickers, [symbol]: ticker },
    })),

  setIndicators: (symbol, ind) =>
    set((state) => ({
      indicators: { ...state.indicators, [symbol]: ind },
    })),

  setPrediction: (symbol, pred) =>
    set((state) => ({
      predictions: { ...state.predictions, [symbol]: pred },
    })),

  setWsStatus: (wsStatus) => set({ wsStatus }),
}));

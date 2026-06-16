// API client — typed fetch wrapper for the AlphaStream Go backend REST API.
import type { ApiResponse, Stock, OHLCV, TechnicalIndicators, PredictionResult, BrokerSummary, BuyRecommendation } from '@/types/stock';

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL ?? 'http://localhost:8080';

// ─── Generic Fetcher ──────────────────────────────────────────────────────────

async function apiFetch<T>(path: string): Promise<T> {
  const res = await fetch(`${API_BASE_URL}${path}`, {
    headers: { 'Content-Type': 'application/json' },
    cache: 'no-store',
  });

  if (!res.ok) {
    throw new Error(`API error ${res.status} on ${path}`);
  }

  const body: ApiResponse<T> = await res.json();
  if (!body.success || body.data === undefined) {
    throw new Error(body.error ?? 'Unknown API error');
  }

  return body.data;
}

// ─── Stock API ────────────────────────────────────────────────────────────────

export const stockApi = {
  getAllStocks: () =>
    apiFetch<Stock[]>('/api/v1/stocks'),

  getStockBySymbol: (symbol: string) =>
    apiFetch<Stock>(`/api/v1/stocks/${symbol}`),

  getOHLCVHistory: (symbol: string, timeframe = '1m', limit = 200) =>
    apiFetch<OHLCV[]>(`/api/v1/stocks/${symbol}/ohlcv?timeframe=${timeframe}&limit=${limit}`),

  getPrediction: (symbol: string) =>
    apiFetch<PredictionResult>(`/api/v1/stocks/${symbol}/prediction`),

  getBrokerSummary: (symbol: string) =>
    apiFetch<BrokerSummary>(`/api/v1/stocks/${symbol}/broker-summary`),

  getBuyRecommendations: () =>
    apiFetch<BuyRecommendation[]>('/api/v1/stocks/recommendations'),
};

// ─── Indicator API ────────────────────────────────────────────────────────────

export const indicatorApi = {
  getLatest: (symbol: string) =>
    apiFetch<TechnicalIndicators>(`/api/v1/indicators/${symbol}`),

  getSummary: (symbol: string) =>
    apiFetch<{ indicators: TechnicalIndicators; prediction: PredictionResult | null }>(
      `/api/v1/indicators/${symbol}/summary`
    ),
};

// ─── WebSocket URL ────────────────────────────────────────────────────────────

export const WS_URL = (process.env.NEXT_PUBLIC_WS_URL ?? 'ws://localhost:8080') + '/ws';

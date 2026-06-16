'use client';
// useStockWebSocket — Zero memory-leak WebSocket hook.
// Cleanup is guaranteed in the useEffect return function when component unmounts.
// Implements exponential backoff reconnection with a configurable max retry limit.
import { useEffect, useRef, useCallback } from 'react';
import { WS_URL } from '@/lib/api';
import { useStockStore } from '@/store/stockStore';
import type {
  WsMessage,
  OHLCV,
  Ticker,
  TechnicalIndicators,
  PredictionResult,
} from '@/types/stock';

// ─── Constants ────────────────────────────────────────────────────────────────

const RECONNECT_BASE_DELAY_MS = 1000;
const RECONNECT_MAX_DELAY_MS = 30000;
const RECONNECT_MAX_ATTEMPTS = 10;
const PING_INTERVAL_MS = 30000;

// ─── Hook ─────────────────────────────────────────────────────────────────────

export function useStockWebSocket() {
  const wsRef = useRef<WebSocket | null>(null);
  const reconnectAttemptRef = useRef(0);
  const reconnectTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const pingTimerRef = useRef<ReturnType<typeof setInterval> | null>(null);
  const isMountedRef = useRef(true);
  const prevSymbolRef = useRef<string | null>(null);

  const { activeSymbol, wsStatus, setWsStatus, pushCandle, setTicker, setIndicators, setPrediction } =
    useStockStore();

  // ── Message Router ────────────────────────────────────────────────────
  const handleMessage = useCallback(
    (event: MessageEvent) => {
      try {
        const msg: WsMessage = JSON.parse(event.data as string);

        switch (msg.type) {
          case 'OHLCV_UPDATE':
            pushCandle(msg.symbol, msg.payload as OHLCV);
            break;
          case 'TICKER_UPDATE':
            setTicker(msg.symbol, msg.payload as Ticker);
            break;
          case 'INDICATOR_UPDATE':
            setIndicators(msg.symbol, msg.payload as TechnicalIndicators);
            break;
          case 'PREDICTION_UPDATE':
            setPrediction(msg.symbol, msg.payload as PredictionResult);
            break;
          case 'PONG':
            // Keepalive acknowledged — no action needed
            break;
          default:
            break;
        }
      } catch {
        // Silently ignore parse errors from malformed messages
      }
    },
    [pushCandle, setTicker, setIndicators, setPrediction]
  );

  // ── Ping Sender ───────────────────────────────────────────────────────
  const startPing = useCallback((ws: WebSocket) => {
    pingTimerRef.current = setInterval(() => {
      if (ws.readyState === WebSocket.OPEN) {
        ws.send(JSON.stringify({ type: 'PING', timestamp: new Date().toISOString() }));
      }
    }, PING_INTERVAL_MS);
  }, []);

  const stopPing = useCallback(() => {
    if (pingTimerRef.current !== null) {
      clearInterval(pingTimerRef.current);
      pingTimerRef.current = null;
    }
  }, []);

  const connect = useCallback(function connect() {
    if (!isMountedRef.current) return;

    setWsStatus('connecting');
    const ws = new WebSocket(WS_URL);
    wsRef.current = ws;

    ws.onopen = () => {
      if (!isMountedRef.current) {
        ws.close();
        return;
      }
      reconnectAttemptRef.current = 0;
      setWsStatus('connected');
      startPing(ws);
    };

    ws.onmessage = handleMessage;

    ws.onerror = () => {
      setWsStatus('error');
    };

    ws.onclose = () => {
      stopPing();
      if (!isMountedRef.current) return;

      setWsStatus('disconnected');

      if (reconnectAttemptRef.current < RECONNECT_MAX_ATTEMPTS) {
        const delay = Math.min(
          RECONNECT_BASE_DELAY_MS * 2 ** reconnectAttemptRef.current,
          RECONNECT_MAX_DELAY_MS
        );
        reconnectAttemptRef.current += 1;
        reconnectTimerRef.current = setTimeout(connect, delay);
      }
    };
  }, [handleMessage, setWsStatus, startPing, stopPing]);

  // ── Lifecycle ─────────────────────────────────────────────────────────
  useEffect(() => {
    isMountedRef.current = true;
    connect();

    // CRITICAL: cleanup prevents memory leaks on unmount
    return () => {
      isMountedRef.current = false;

      // Cancel pending reconnect timer
      if (reconnectTimerRef.current !== null) {
        clearTimeout(reconnectTimerRef.current);
        reconnectTimerRef.current = null;
      }

      // Stop ping interval
      stopPing();

      // Close WebSocket — this prevents the browser from holding a dead connection
      if (wsRef.current) {
        wsRef.current.onclose = null; // Prevent reconnect loop on intentional close
        wsRef.current.close(1000, 'Component unmounted');
        wsRef.current = null;
      }
    };
  }, [connect, stopPing]);

  // Subscribe/Unsubscribe on activeSymbol or connection status change
  useEffect(() => {
    const ws = wsRef.current;
    if (ws && ws.readyState === WebSocket.OPEN && activeSymbol && wsStatus === 'connected') {
      // Unsubscribe from previous symbol if different
      if (prevSymbolRef.current && prevSymbolRef.current !== activeSymbol) {
        ws.send(
          JSON.stringify({
            type: 'UNSUBSCRIBE',
            symbol: prevSymbolRef.current,
            timestamp: new Date().toISOString(),
          })
        );
      }
      // Subscribe to new symbol
      ws.send(
        JSON.stringify({
          type: 'SUBSCRIBE',
          symbol: activeSymbol,
          timestamp: new Date().toISOString(),
        })
      );
      prevSymbolRef.current = activeSymbol;
    }
  }, [activeSymbol, wsStatus]);

  return { wsStatus };
}

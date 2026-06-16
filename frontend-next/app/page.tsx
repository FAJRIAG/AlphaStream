'use client';
import { useEffect, useState } from 'react';
import dynamic from 'next/dynamic';
import { useStockWebSocket } from '@/hooks/useStockWebSocket';
import { useStockStore } from '@/store/stockStore';
import { stockApi } from '@/lib/api';

import BloombergHeader from '@/components/bloomberg/BloombergHeader';
import WatchlistPanel from '@/components/bloomberg/WatchlistPanel';
import QuoteBox from '@/components/bloomberg/QuoteBox';
import IndicatorPanel from '@/components/bloomberg/IndicatorPanel';
import PredictionBar from '@/components/bloomberg/PredictionBar';
import TickerTape from '@/components/bloomberg/TickerTape';
import MarketLeaderboard from '@/components/bloomberg/MarketLeaderboard';
import SignalHistory from '@/components/bloomberg/SignalHistory';
import BrokerSummaryPanel from '@/components/bloomberg/BrokerSummaryPanel';

// Dynamically import chart to avoid SSR issues (lightweight-charts is browser-only)
const CandlestickChart = dynamic(
  () => import('@/components/chart/CandlestickChart'),
  { ssr: false, loading: () => <ChartSkeleton /> }
);

// ─── Chart Skeleton ───────────────────────────────────────────────────────────
function ChartSkeleton() {
  return (
    <div className="w-full h-full min-h-[300px] flex items-center justify-center bg-black border border-[#1a1a1a] animate-pulse">
      <div className="bb-label text-[10px] text-neutral-600">LOADING REAL-TIME CHART...</div>
    </div>
  );
}

// ─── Bloomberg Dashboard Page ───────────────────────────────────────────────────────────
export default function DashboardPage() {
  const { stocks, activeSymbol, setStocks, setCandles } = useStockStore();
  const [rightPanelTab, setRightPanelTab] = useState<'ticks' | 'broker'>('ticks');
  const [brokerRefreshKey, setBrokerRefreshKey] = useState<number>(0);
  const [maximizedRightPanel, setMaximizedRightPanel] = useState<'quote' | 'technicals' | 'comparison' | 'bottom' | null>(null);
  const [isChartMaximized, setIsChartMaximized] = useState(false);

  // Connect WebSocket — routes messages to Zustand store
  useStockWebSocket();

  // Load stock master list on mount
  useEffect(() => {
    stockApi.getAllStocks().then(setStocks).catch(console.error);
  }, [setStocks]);

  // Load OHLCV history whenever active symbol changes
  useEffect(() => {
    if (!activeSymbol) return;
    stockApi
      .getOHLCVHistory(activeSymbol, '1m', 200)
      .then((candles) => setCandles(activeSymbol, candles))
      .catch(console.error);
  }, [activeSymbol, setCandles]);

  return (
    <div className="flex flex-col h-screen overflow-hidden bg-black select-none text-[#e0e0e0] font-sans">
      {/* 1. Top Bloomberg Header */}
      <BloombergHeader />

      {/* 2. Main Workspace Layout */}
      <div className="flex-1 min-h-0 relative">
        <div
          className="h-full"
          style={{
            display: 'grid',
            gridTemplateColumns: isChartMaximized ? '1fr' : '200px 1fr 240px',
            gridTemplateRows: '1fr',
            gap: isChartMaximized ? '0px' : '1px',
            background: 'var(--bb-border)',
            height: '100%',
          }}
        >
          {/* Column 1: Watchlist Panel */}
          <div className="overflow-y-auto" style={{ display: isChartMaximized ? 'none' : 'block' }}>
            <WatchlistPanel stocks={stocks} />
          </div>

          {/* Column 2: Center Live Chart */}
          <div className="flex flex-col min-w-0 h-full relative" style={{ background: 'var(--bb-surface)' }}>
            <div className="flex-grow flex-shrink flex flex-col h-full min-h-0 p-1.5">
              <div className="flex items-center justify-between px-2 py-1 flex-shrink-0" style={{ borderBottom: '1px solid var(--bb-border)' }}>
                <span className="bb-label text-[10px]" style={{ color: 'var(--bb-orange)' }}>
                  {activeSymbol} · DAILY CHART
                </span>
                <div className="flex items-center gap-2">
                  <span className="bb-mono text-[9px] text-neutral-500 mr-2">PERIOD: 1D (DAILY FEED)</span>
                  <button
                    onClick={() => setIsChartMaximized(!isChartMaximized)}
                    className="text-[9px] text-neutral-500 hover:text-white font-mono bg-neutral-900 border border-neutral-800 px-1.5 py-0.5 rounded cursor-pointer font-bold"
                  >
                    {isChartMaximized ? 'SHRINK [-]' : 'EXPAND [+]'}
                  </button>
                </div>
              </div>
              <div className="flex-1 min-h-0 relative mt-2">
                <CandlestickChart symbol={activeSymbol} />
              </div>
            </div>
          </div>

          {/* Column 3: Right details (QuoteBox + IndicatorPanel + MarketLeaderboard + SignalHistory) */}
          <div className="flex flex-col min-h-0" style={{ display: isChartMaximized ? 'none' : 'flex', background: 'var(--bb-surface)' }}>
            {/* Row 1: QuoteBox */}
            <div
              className={`${maximizedRightPanel === 'quote' ? 'flex-1' : 'flex-[5]'} min-h-0`}
              style={{
                display: maximizedRightPanel && maximizedRightPanel !== 'quote' ? 'none' : 'block',
                borderBottom: '1px solid var(--bb-border)',
              }}
            >
              <QuoteBox
                symbol={activeSymbol}
                isMaximized={maximizedRightPanel === 'quote'}
                onToggleMaximize={() =>
                  setMaximizedRightPanel(maximizedRightPanel === 'quote' ? null : 'quote')
                }
              />
            </div>

            {/* Row 2: IndicatorPanel */}
            <div
              className={`${maximizedRightPanel === 'technicals' ? 'flex-1' : 'flex-[5]'} min-h-0`}
              style={{
                display: maximizedRightPanel && maximizedRightPanel !== 'technicals' ? 'none' : 'block',
                borderBottom: '1px solid var(--bb-border)',
              }}
            >
              <IndicatorPanel
                symbol={activeSymbol}
                isMaximized={maximizedRightPanel === 'technicals'}
                onToggleMaximize={() =>
                  setMaximizedRightPanel(maximizedRightPanel === 'technicals' ? null : 'technicals')
                }
              />
            </div>

            {/* Row 3: MarketLeaderboard */}
            <div
              className={`${maximizedRightPanel === 'comparison' ? 'flex-1' : 'flex-[6]'} min-h-0`}
              style={{
                display: maximizedRightPanel && maximizedRightPanel !== 'comparison' ? 'none' : 'block',
                borderBottom: '1px solid var(--bb-border)',
              }}
            >
              <MarketLeaderboard
                isMaximized={maximizedRightPanel === 'comparison'}
                onToggleMaximize={() =>
                  setMaximizedRightPanel(maximizedRightPanel === 'comparison' ? null : 'comparison')
                }
              />
            </div>

            {/* Row 4: SignalHistory / BrokerSummaryPanel Tabbed Panel */}
            <div
              className={`${maximizedRightPanel === 'bottom' ? 'flex-1' : 'flex-[5]'} min-h-0 flex flex-col`}
              style={{
                display: maximizedRightPanel && maximizedRightPanel !== 'bottom' ? 'none' : 'flex',
                border: 'none',
              }}
            >
              <div className="bb-panel-header flex items-center gap-3 select-none">
                <button
                  onClick={() => setRightPanelTab('ticks')}
                  className={`text-[10px] font-bold tracking-wider hover:text-white uppercase transition-colors cursor-pointer ${
                    rightPanelTab === 'ticks'
                      ? 'text-[var(--bb-orange)] border-b border-[var(--bb-orange)]'
                      : 'text-neutral-500'
                  }`}
                  style={{ paddingBottom: '2px' }}
                >
                  PRICE TICKS
                </button>
                <span className="text-neutral-700 text-[9px]">|</span>
                <button
                  onClick={() => setRightPanelTab('broker')}
                  className={`text-[10px] font-bold tracking-wider hover:text-white uppercase transition-colors cursor-pointer ${
                    rightPanelTab === 'broker'
                      ? 'text-[var(--bb-orange)] border-b border-[var(--bb-orange)]'
                      : 'text-neutral-500'
                  }`}
                  style={{ paddingBottom: '2px' }}
                >
                  BROKER SUMMARY
                </button>

                <div className="flex items-center gap-1.5 ml-auto">
                  {rightPanelTab === 'ticks' ? (
                    <span className="bb-label text-[9px] text-neutral-500 font-mono mr-1.5">1M INTERVALS</span>
                  ) : (
                    <button
                      onClick={() => setBrokerRefreshKey((prev) => prev + 1)}
                      className="text-[9px] text-neutral-500 hover:text-white font-mono bg-neutral-900 border border-neutral-800 px-1.5 py-0.5 rounded cursor-pointer"
                    >
                      REFRESH
                    </button>
                  )}
                  <button
                    onClick={() =>
                      setMaximizedRightPanel(maximizedRightPanel === 'bottom' ? null : 'bottom')
                    }
                    className="text-[9px] text-neutral-500 hover:text-white font-mono bg-neutral-900 border border-neutral-800 px-1.5 py-0.5 rounded cursor-pointer font-bold"
                  >
                    {maximizedRightPanel === 'bottom' ? 'SHRINK [-]' : 'EXPAND [+]'}
                  </button>
                </div>
              </div>

              <div className="flex-1 min-h-0">
                {rightPanelTab === 'ticks' ? (
                  <SignalHistory symbol={activeSymbol} hideHeader />
                ) : (
                  <BrokerSummaryPanel
                    key={`${activeSymbol}-${brokerRefreshKey}`}
                    symbol={activeSymbol}
                    hideHeader
                  />
                )}
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* 3. Prediction Status Bar */}
      <PredictionBar symbol={activeSymbol} />

      {/* 4. Bottom Ticker Tape */}
      <TickerTape stocks={stocks} />
    </div>
  );
}

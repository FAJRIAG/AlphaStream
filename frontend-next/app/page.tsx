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
  
  // Loading states
  const [mounted, setMounted] = useState(false);
  const [progress, setProgress] = useState(0);
  const [isDataLoaded, setIsDataLoaded] = useState(false);
  const [showDashboard, setShowDashboard] = useState(false);
  const [fadeOut, setFadeOut] = useState(false);

  // Connect WebSocket — routes messages to Zustand store
  useStockWebSocket();

  // Load stock master list on mount and handle progress increments
  useEffect(() => {
    setMounted(true);

    const interval = setInterval(() => {
      setProgress((prev) => {
        if (prev >= 100) {
          clearInterval(interval);
          return 100;
        }
        // Progress builds up dynamically (quicker at start, slower at end to simulate fetching)
        const diff = Math.max(1, Math.floor((100 - prev) * 0.12));
        return Math.min(100, prev + diff);
      });
    }, 100);

    stockApi.getAllStocks()
      .then((data) => {
        setStocks(data);
        setIsDataLoaded(true);
      })
      .catch((err) => {
        console.error(err);
        setIsDataLoaded(true); // Fallback to let user in if offline
      });

    return () => clearInterval(interval);
  }, [setStocks]);

  // Load OHLCV history whenever active symbol changes
  useEffect(() => {
    if (!activeSymbol) return;
    stockApi
      .getOHLCVHistory(activeSymbol, '1m', 200)
      .then((candles) => setCandles(activeSymbol, candles))
      .catch(console.error);
  }, [activeSymbol, setCandles]);

  // Handle transition to dashboard when ready
  useEffect(() => {
    if (progress === 100 && isDataLoaded) {
      setFadeOut(true);
      const timer = setTimeout(() => {
        setShowDashboard(true);
      }, 500); // 500ms fade transition
      return () => clearTimeout(timer);
    }
  }, [progress, isDataLoaded]);

  // Get loading details matching current progress
  const getStatusMessage = () => {
    if (progress < 25) return 'ESTABLISHING SECURE CONNECTION TO IDX FEED...';
    if (progress < 55) return 'SYNCHRONIZING SYMBOLS DATABASE [934 EMITEN]...';
    if (progress < 75) return 'INITIALIZING QUANTITATIVE PREDICTIVE ENGINE...';
    if (progress < 99) return 'WIRING REAL-TIME WEBSOCKET DATA STREAM...';
    return 'SYSTEM READY. LAUNCHING ALPHASTREAM...';
  };

  if (!showDashboard) {
    return (
      <div 
        className={`flex flex-col h-screen overflow-hidden bg-black select-none text-[#e0e0e0] font-sans justify-center items-center relative transition-all duration-500 ease-out ${
          fadeOut ? 'opacity-0 scale-95' : 'opacity-100'
        }`}
        style={{
          background: 'radial-gradient(circle at center, #140d02 0%, #000000 100%)'
        }}
      >
        {/* Subtle grid backdrop */}
        <div className="absolute inset-0 bg-[linear-gradient(rgba(42,42,42,0.05)_1px,_transparent_1px),_linear-gradient(90deg,_rgba(42,42,42,0.05)_1px,_transparent_1px)] bg-[size:30px_30px] pointer-events-none" />

        {/* Glowing background halo */}
        <div className="absolute w-72 h-72 rounded-full bg-gradient-to-tr from-[var(--bb-orange)] to-[var(--bb-cyan)] opacity-5 blur-3xl animate-pulse" />

        {/* Animated Alpha Logo */}
        <div className="relative flex items-center justify-center mb-6 z-10">
          <svg className="w-24 h-24 animate-logo-pulse" viewBox="0 0 100 100" fill="none" xmlns="http://www.w3.org/2000/svg">
            <path 
              d="M72 28C72 40 60 72 43 72C30 72 20 62 20 48C20 34 30 24 43 24C58 24 67 48 76 72" 
              stroke="url(#alphaGradient)" 
              strokeWidth="6" 
              strokeLinecap="round" 
              className="animate-draw-path" 
            />
            <defs>
              <linearGradient id="alphaGradient" x1="20" y1="24" x2="76" y2="72" gradientUnits="userSpaceOnUse">
                <stop offset="0%" stopColor="#f5a623" />
                <stop offset="50%" stopColor="#ff7b00" />
                <stop offset="100%" stopColor="#00d4e8" />
              </linearGradient>
            </defs>
          </svg>
        </div>

        {/* Logo Text & Tagline */}
        <div className="text-center space-y-1 z-10">
          <div className="text-sm font-black tracking-[0.25em] text-white">
            α ALPHASTREAM
          </div>
          <div className="text-[8px] font-mono tracking-[0.4em] text-neutral-500 uppercase">
            Quant Analytics Terminal
          </div>
        </div>

        {/* Loading Progress Bar */}
        <div className="mt-12 flex flex-col items-center gap-2.5 w-60 z-10">
          <div className="w-full h-1 bg-[#141414] border border-[#222] rounded-full overflow-hidden relative">
            <div 
              className="h-full bg-gradient-to-r from-[var(--bb-orange)] to-[#ff7b00] rounded-full transition-all duration-200 ease-out" 
              style={{ width: `${progress}%` }} 
            />
          </div>
          
          {/* Status logs */}
          <div className="flex justify-between w-full font-mono text-[8px] text-neutral-500">
            <span className="truncate pr-4 uppercase">{getStatusMessage()}</span>
            <span className="text-[var(--bb-orange)] font-bold">{progress}%</span>
          </div>
        </div>

        {/* Boot stats footer */}
        <div className="absolute bottom-6 font-mono text-[8px] text-neutral-600 tracking-wider">
          ALPHASTREAM SECURE BOOT v3.5 // M1 OPTIMIZED // ACTIVE
        </div>
      </div>
    );
  }

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

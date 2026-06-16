'use client';
import { useState } from 'react';
import { useStockStore } from '@/store/stockStore';
import type { Stock } from '@/types/stock';

interface WatchlistPanelProps {
  stocks: Stock[];
}

export default function WatchlistPanel({ stocks }: WatchlistPanelProps) {
  const { activeSymbol, setActiveSymbol, tickers } = useStockStore();
  const [searchQuery, setSearchQuery] = useState('');

  const filteredStocks = stocks.filter((stock) =>
    stock.symbol.toLowerCase().includes(searchQuery.toLowerCase()) ||
    stock.name.toLowerCase().includes(searchQuery.toLowerCase())
  );

  return (
    <div className="bb-panel h-full flex flex-col" style={{ border: 'none' }}>
      <div className="bb-panel-header">
        <span className="bb-label" style={{ color: 'var(--bb-orange)', fontSize: '10px' }}>
          WATCHLIST
        </span>
        <span className="bb-label ml-auto" style={{ fontSize: '9px' }}>IDX</span>
      </div>

      {/* Search Input */}
      <div className="px-2 py-1.5" style={{ borderBottom: '1px solid var(--bb-border)', background: 'var(--bb-surface)' }}>
        <input
          type="text"
          placeholder="SEARCH STOCK..."
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
          className="w-full bg-black text-[#e0e0e0] font-mono text-[10px] border border-neutral-800 focus:border-[var(--bb-orange)] focus:outline-none px-2 py-1 rounded placeholder-neutral-600 uppercase"
        />
      </div>

      {/* Column Header */}
      <div className="flex justify-between px-2 py-1" style={{
        background: 'var(--bb-surface2)',
        borderBottom: '1px solid var(--bb-border)',
      }}>
        <span className="bb-label" style={{ fontSize: '9px' }}>SYMBOL</span>
        <span className="bb-label" style={{ fontSize: '9px' }}>LAST / CHG%</span>
      </div>

      <div className="flex-1 overflow-y-auto">
        {filteredStocks.map((stock) => {
          const ticker = tickers[stock.symbol];
          const isActive = activeSymbol === stock.symbol;
          const isUp = (ticker?.change ?? 0) >= 0;

          return (
            <div
              key={stock.symbol}
              onClick={() => setActiveSymbol(stock.symbol)}
              className={`bb-watchlist-item ${isActive ? 'active' : ''}`}
            >
              <div className="flex justify-between items-center">
                <div>
                  <div className="bb-mono text-xs font-bold" style={{
                    color: isActive ? 'var(--bb-orange)' : 'var(--bb-text)',
                  }}>
                    {stock.symbol}
                  </div>
                  <div className="bb-label" style={{ fontSize: '9px', marginTop: '1px' }}>
                    {stock.name.split(' ').slice(0, 2).join(' ')}
                  </div>
                </div>
                <div className="text-right">
                  <div className="bb-mono text-xs font-bold">
                    {ticker?.price != null
                      ? ticker.price.toLocaleString('id-ID', { maximumFractionDigits: 0 })
                      : '—'}
                  </div>
                  {ticker && (
                    <div className="bb-mono" style={{
                      fontSize: '10px',
                      fontWeight: 700,
                      color: isUp ? 'var(--bb-green)' : 'var(--bb-red)',
                    }}>
                      {isUp ? '▲' : '▼'}{Math.abs(ticker.change_percent).toFixed(2)}%
                    </div>
                  )}
                </div>
              </div>
            </div>
          );
        })}
      </div>

      {/* Connection Status */}
      <WsStatus />
    </div>
  );
}

function WsStatus() {
  const wsStatus = useStockStore((s) => s.wsStatus);
  const colors: Record<string, string> = {
    connected: 'var(--bb-green)',
    connecting: 'var(--bb-orange)',
    disconnected: 'var(--bb-text-muted)',
    error: 'var(--bb-red)',
  };
  const labels: Record<string, string> = {
    connected: 'STREAMING',
    connecting: 'CONNECTING',
    disconnected: 'OFFLINE',
    error: 'ERROR',
  };

  return (
    <div className="flex items-center gap-2 px-2 py-2" style={{
      borderTop: '1px solid var(--bb-border)',
      background: 'var(--bb-surface2)',
    }}>
      <span className={`w-1.5 h-1.5 rounded-full ${wsStatus === 'connected' ? 'bb-blink' : ''}`}
        style={{ background: colors[wsStatus] ?? 'var(--bb-text-muted)', flexShrink: 0 }} />
      <span className="bb-label" style={{ fontSize: '9px', color: colors[wsStatus] }}>
        {labels[wsStatus] ?? 'UNKNOWN'}
      </span>
    </div>
  );
}

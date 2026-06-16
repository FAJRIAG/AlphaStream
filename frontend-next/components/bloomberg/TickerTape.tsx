'use client';
// TickerTape — auto-scrolling bottom bar with all symbols' live prices
import { useStockStore } from '@/store/stockStore';
import type { Stock } from '@/types/stock';

interface TickerTapeProps {
  stocks: Stock[];
}

export default function TickerTape({ stocks }: TickerTapeProps) {
  const tickers = useStockStore((s) => s.tickers);
  const activeSymbol = useStockStore((s) => s.activeSymbol);
  const setActiveSymbol = useStockStore((s) => s.setActiveSymbol);

  const items = [...stocks, ...stocks]; // Double for seamless loop

  return (
    <div className="overflow-hidden" style={{
      background: 'var(--bb-surface)',
      borderTop: '1px solid var(--bb-border)',
      height: '28px',
      display: 'flex',
      alignItems: 'center',
    }}>
      <div className="ticker-track">
        {items.map((stock, idx) => {
          const ticker = tickers[stock.symbol];
          const isUp = (ticker?.change ?? 0) >= 0;
          const sign = isUp ? '▲' : '▼';
          const color = isUp ? 'var(--bb-green)' : 'var(--bb-red)';
          const isActive = stock.symbol === activeSymbol;

          return (
            <div key={`${stock.symbol}-${idx}`}
              onClick={() => setActiveSymbol(stock.symbol)}
              className="flex items-center gap-3 px-6 border-r cursor-pointer hover:bg-[rgba(255,255,255,0.05)] transition-all duration-150"
              style={{
                borderColor: isActive ? 'var(--bb-orange)' : 'var(--bb-border)',
                whiteSpace: 'nowrap',
                backgroundColor: isActive ? 'var(--bb-orange-bg)' : 'transparent',
              }}>
              <span className="bb-label font-bold" style={{ color: isActive ? 'var(--bb-orange)' : 'var(--bb-orange)', fontSize: '10px' }}>
                {stock.symbol}
              </span>
              <span className="bb-mono text-xs font-bold" style={{ color: 'var(--bb-text)' }}>
                {ticker?.price != null
                  ? ticker.price.toLocaleString('id-ID', { maximumFractionDigits: 0 })
                  : '—'}
              </span>
              {ticker && (
                <span className="bb-mono text-xs font-semibold" style={{ color }}>
                  {sign}{Math.abs(ticker.change_percent).toFixed(2)}%
                </span>
              )}
            </div>
          );
        })}
      </div>
    </div>
  );
}

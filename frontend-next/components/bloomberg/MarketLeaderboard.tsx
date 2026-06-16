'use client';
// MarketLeaderboard — Bloomberg-style relative comparison of active IDX symbols
import { useStockStore } from '@/store/stockStore';

interface MarketLeaderboardProps {
  isMaximized?: boolean;
  onToggleMaximize?: () => void;
}

export default function MarketLeaderboard({ isMaximized, onToggleMaximize }: MarketLeaderboardProps) {
  const { stocks, tickers } = useStockStore();

  return (
    <div className="bb-panel h-full flex flex-col" style={{ border: 'none' }}>
      <div className="bb-panel-header flex items-center justify-between">
        <span className="bb-label" style={{ color: 'var(--bb-orange)', fontSize: '10px' }}>
          MARKET COMPARISON (IDX)
        </span>
        {onToggleMaximize && (
          <button
            onClick={onToggleMaximize}
            className="text-[9px] text-neutral-500 hover:text-white font-mono bg-neutral-900 border border-neutral-800 px-1.5 py-0.5 rounded cursor-pointer font-bold"
          >
            {isMaximized ? 'SHRINK [-]' : 'EXPAND [+]'}
          </button>
        )}
      </div>

      <div className="bb-panel-body flex-1 p-2 flex flex-col space-y-2 text-[11px]">
        {/* Table Header */}
        <div className="flex justify-between border-b border-[#2a2a2a] pb-1 mb-1 bb-label text-[9px] font-bold text-neutral-600">
          <span className="w-12">SYMBOL</span>
          <span className="w-16 text-right">PRICE</span>
          <span className="w-14 text-right">CHG%</span>
          <span className="w-16 text-right">VOLUME</span>
        </div>

        {/* Rows */}
        <div className="flex-1 min-h-0 overflow-y-auto space-y-1.5 pr-1">
          {stocks.map((stock) => {
            const ticker = tickers[stock.symbol];
            const isUp = (ticker?.change ?? 0) >= 0;
            const color = isUp ? 'var(--bb-green)' : 'var(--bb-red)';

            return (
              <div key={stock.symbol} className="flex justify-between items-center py-0.5 font-mono">
                <span className="w-12 text-[#f5a623] font-bold">{stock.symbol}</span>
                <span className="w-16 text-right text-white font-bold">
                  {ticker?.price != null
                    ? ticker.price.toLocaleString('id-ID', { maximumFractionDigits: 0 })
                    : '—'}
                </span>
                <span className="w-14 text-right font-bold" style={{ color }}>
                  {ticker != null
                    ? (isUp ? '+' : '') + ticker.change_percent.toFixed(2) + '%'
                    : '—'}
                </span>
                <span className="w-16 text-right text-neutral-400">
                  {ticker?.volume != null
                    ? (ticker.volume / 1_000_000).toFixed(1) + 'M'
                    : '—'}
                </span>
              </div>
            );
          })}
        </div>
      </div>
    </div>
  );
}

'use client';
import { useStockStore } from '@/store/stockStore';
import type { Stock } from '@/types/stock';

interface StockListProps {
  stocks: Stock[];
}

export default function StockList({ stocks }: StockListProps) {
  const { activeSymbol, setActiveSymbol, tickers } = useStockStore();

  return (
    <div className="space-y-1">
      <p className="text-xs text-gray-600 uppercase tracking-widest font-semibold px-3 mb-3">
        Watchlist
      </p>
      {stocks.map((stock) => {
        const ticker = tickers[stock.symbol];
        const isActive = activeSymbol === stock.symbol;
        const isUp = (ticker?.change ?? 0) >= 0;

        return (
          <button
            key={stock.symbol}
            onClick={() => setActiveSymbol(stock.symbol)}
            className={`w-full flex items-center justify-between px-3 py-3 rounded-xl transition-all duration-150 text-left group ${
              isActive
                ? 'bg-indigo-500/15 border border-indigo-500/30'
                : 'hover:bg-white/5 border border-transparent'
            }`}
          >
            <div>
              <p className={`font-bold text-sm ${isActive ? 'text-indigo-300' : 'text-gray-200'}`}>
                {stock.symbol}
              </p>
              <p className="text-xs text-gray-600 truncate max-w-[120px]">{stock.name}</p>
            </div>
            <div className="text-right">
              <p className="text-sm font-bold tabular-nums text-gray-100">
                {ticker?.price != null
                  ? ticker.price.toLocaleString('id-ID', { maximumFractionDigits: 0 })
                  : '—'}
              </p>
              {ticker && (
                <p className={`text-xs font-semibold ${isUp ? 'text-emerald-400' : 'text-red-400'}`}>
                  {isUp ? '+' : ''}{ticker.change_percent.toFixed(2)}%
                </p>
              )}
            </div>
          </button>
        );
      })}
    </div>
  );
}

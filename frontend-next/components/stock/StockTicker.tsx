'use client';
import { useStockStore } from '@/store/stockStore';

interface StockTickerProps {
  symbol: string;
}

export default function StockTicker({ symbol }: StockTickerProps) {
  const ticker = useStockStore((s) => s.tickers[symbol]);
  const stock = useStockStore((s) => s.stocks.find((st) => st.symbol === symbol));

  const isUp = (ticker?.change ?? 0) >= 0;

  return (
    <div className="flex items-end justify-between">
      <div>
        <div className="flex items-center gap-3">
          <h1 className="text-3xl font-black tracking-tight text-white">{symbol}</h1>
          <span className="text-sm text-gray-500 font-medium">{stock?.exchange ?? 'IDX'}</span>
          <span className="px-2 py-0.5 rounded-full text-xs font-semibold bg-indigo-500/20 text-indigo-300 border border-indigo-500/30">
            LIVE
          </span>
        </div>
        <p className="text-sm text-gray-500 mt-0.5">{stock?.name ?? '—'}</p>
      </div>

      <div className="text-right">
        <p className="text-4xl font-black tabular-nums text-white">
          {ticker?.price != null
            ? ticker.price.toLocaleString('id-ID', { minimumFractionDigits: 0 })
            : '—'}
        </p>
        {ticker && (
          <div className={`flex items-center justify-end gap-2 mt-1 text-sm font-semibold ${isUp ? 'text-emerald-400' : 'text-red-400'}`}>
            <span>{isUp ? '▲' : '▼'}</span>
            <span>{Math.abs(ticker.change).toFixed(0)}</span>
            <span>({Math.abs(ticker.change_percent).toFixed(2)}%)</span>
          </div>
        )}
      </div>
    </div>
  );
}

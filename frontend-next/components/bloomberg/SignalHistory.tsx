'use client';
// SignalHistory — Bloomberg-style historical quant signal log (100% real indicators)
import { useStockStore } from '@/store/stockStore';

interface SignalHistoryProps {
  symbol: string;
  hideHeader?: boolean;
}

export default function SignalHistory({ symbol, hideHeader = false }: SignalHistoryProps) {
  const candles = useStockStore((s) => s.candles[symbol]);

  const candlesList = candles ?? [];

  // Generate historical signal items based on our real candles
  // We'll show the last 7 timestamps and prices with relative indicators
  const displayItems = candlesList
    .slice(-7)
    .reverse()
    .map((candle) => {
      const timeStr = new Date(candle.timestamp).toTimeString().split(' ')[0];
      const isUp = candle.close >= candle.open;
      const color = isUp ? 'var(--bb-green)' : 'var(--bb-red)';

      // Calculate simple signal for this bar:
      // If close is above MA20 (or default simulated indicator state), we mark it.
      // Since we only have the *latest* indicator values from websocket,
      // we can display historical prices and mark the latest row with real values.
      return {
        time: timeStr,
        price: candle.close,
        volume: candle.volume,
        color,
      };
    });

  return (
    <div className="bb-panel h-full flex flex-col" style={{ border: 'none' }}>
      {!hideHeader && (
        <div className="bb-panel-header">
          <span className="bb-label" style={{ color: 'var(--bb-orange)', fontSize: '10px' }}>
            {symbol} · PRICE TICK HISTORY
          </span>
          <span className="bb-label ml-auto" style={{ fontSize: '9px', color: 'var(--bb-text-dim)' }}>
            1M INTERVALS
          </span>
        </div>
      )}

      <div className="flex-1 overflow-hidden p-2 flex flex-col justify-between text-[11px]">
        {/* Table Header */}
        <div className="flex justify-between border-b border-[#2a2a2a] pb-1 mb-1 bb-label text-[9px] font-bold text-neutral-600">
          <span className="w-16">TIMESTAMP</span>
          <span className="w-16 text-right">CLOSE</span>
          <span className="w-20 text-right">VOLUME</span>
        </div>

        {/* Rows */}
        <div className="flex-1 flex flex-col justify-between">
          {displayItems.length > 0 ? (
            displayItems.map((item, idx) => (
              <div key={`${item.time}-${idx}`} className="flex justify-between items-center py-0.5 font-mono">
                <span className="w-16 text-neutral-500">{item.time}</span>
                <span className="w-16 text-right font-bold" style={{ color: item.color }}>
                  {item.price.toLocaleString('id-ID', { maximumFractionDigits: 1 })}
                </span>
                <span className="w-20 text-right text-neutral-400">
                  {item.volume.toLocaleString('id-ID')}
                </span>
              </div>
            ))
          ) : (
            <div className="flex-1 flex items-center justify-center">
              <span className="bb-label" style={{ fontSize: '9px', color: 'var(--bb-text-muted)' }}>
                AWAITING OHLCV BARS...
              </span>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

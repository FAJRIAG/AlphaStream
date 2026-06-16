'use client';
import { useStockStore } from '@/store/stockStore';

interface IndicatorPanelProps {
  symbol: string;
}

export default function IndicatorPanel({ symbol }: IndicatorPanelProps) {
  const ind = useStockStore((s) => s.indicators[symbol]);

  const fmt = (v: number | null | undefined, decimals = 2) =>
    v != null ? v.toFixed(decimals) : '—';

  const rsiColor = (rsi: number | null | undefined) => {
    if (rsi == null) return 'text-gray-500';
    if (rsi >= 70) return 'text-red-400';
    if (rsi <= 30) return 'text-emerald-400';
    return 'text-gray-300';
  };

  return (
    <div className="grid grid-cols-2 gap-3">
      {/* MA 20 */}
      <div className="bg-white/5 rounded-xl p-4 border border-white/5">
        <p className="text-xs text-gray-500 uppercase tracking-widest mb-1">MA 20</p>
        <p className="text-xl font-bold text-indigo-400">{fmt(ind?.ma_20)}</p>
      </div>

      {/* MA 50 */}
      <div className="bg-white/5 rounded-xl p-4 border border-white/5">
        <p className="text-xs text-gray-500 uppercase tracking-widest mb-1">MA 50</p>
        <p className="text-xl font-bold text-violet-400">{fmt(ind?.ma_50)}</p>
      </div>

      {/* RSI 14 */}
      <div className="bg-white/5 rounded-xl p-4 border border-white/5">
        <p className="text-xs text-gray-500 uppercase tracking-widest mb-1">RSI 14</p>
        <p className={`text-xl font-bold ${rsiColor(ind?.rsi_14)}`}>{fmt(ind?.rsi_14)}</p>
        {ind?.rsi_14 != null && (
          <div className="mt-2 w-full bg-gray-800 rounded-full h-1.5">
            <div
              className={`h-1.5 rounded-full transition-all duration-500 ${
                ind.rsi_14 >= 70
                  ? 'bg-red-500'
                  : ind.rsi_14 <= 30
                  ? 'bg-emerald-500'
                  : 'bg-indigo-500'
              }`}
              style={{ width: `${Math.min(ind.rsi_14, 100)}%` }}
            />
          </div>
        )}
      </div>

      {/* ATR 14 */}
      <div className="bg-white/5 rounded-xl p-4 border border-white/5">
        <p className="text-xs text-gray-500 uppercase tracking-widest mb-1">ATR 14</p>
        <p className="text-xl font-bold text-amber-400">{fmt(ind?.atr_14)}</p>
      </div>
    </div>
  );
}

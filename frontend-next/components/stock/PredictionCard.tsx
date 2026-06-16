'use client';
import { useStockStore } from '@/store/stockStore';
import type { TrendDirection } from '@/types/stock';

interface PredictionCardProps {
  symbol: string;
}

const directionConfig: Record<TrendDirection, { label: string; color: string; bg: string; icon: string }> = {
  BULLISH: { label: 'BULLISH', color: 'text-emerald-400', bg: 'bg-emerald-500/10 border-emerald-500/30', icon: '↑' },
  BEARISH: { label: 'BEARISH', color: 'text-red-400', bg: 'bg-red-500/10 border-red-500/30', icon: '↓' },
  NEUTRAL: { label: 'NEUTRAL', color: 'text-gray-400', bg: 'bg-gray-500/10 border-gray-500/30', icon: '→' },
};

export default function PredictionCard({ symbol }: PredictionCardProps) {
  const pred = useStockStore((s) => s.predictions[symbol]);

  if (!pred) {
    return (
      <div className="rounded-xl bg-white/5 border border-white/5 p-5 animate-pulse">
        <div className="h-4 bg-white/10 rounded w-1/3 mb-3" />
        <div className="h-8 bg-white/10 rounded w-2/3 mb-2" />
        <div className="h-3 bg-white/5 rounded w-full" />
      </div>
    );
  }

  const cfg = directionConfig[pred.direction];
  const probPct = Math.round(pred.probability * 100);
  const isGolden = pred.signals.is_golden_cross;
  const isDeath = pred.signals.is_death_cross;

  return (
    <div className="rounded-xl bg-white/5 border border-white/5 p-5 space-y-4">
      {/* Header */}
      <div className="flex items-center justify-between">
        <p className="text-xs text-gray-500 uppercase tracking-widest font-semibold">Prediction Engine</p>
        <div className={`px-3 py-1 rounded-full text-xs font-bold border ${cfg.bg} ${cfg.color}`}>
          {cfg.icon} {cfg.label}
        </div>
      </div>

      {/* Probability Bar */}
      <div>
        <div className="flex justify-between text-xs text-gray-500 mb-1.5">
          <span>Trend Probability</span>
          <span className={`font-bold ${cfg.color}`}>{probPct}%</span>
        </div>
        <div className="w-full bg-gray-800 rounded-full h-2">
          <div
            className={`h-2 rounded-full transition-all duration-700 ${
              pred.direction === 'BULLISH'
                ? 'bg-emerald-500'
                : pred.direction === 'BEARISH'
                ? 'bg-red-500'
                : 'bg-gray-500'
            }`}
            style={{ width: `${probPct}%` }}
          />
        </div>
      </div>

      {/* Target Prices */}
      <div className="grid grid-cols-2 gap-3">
        <div className="bg-emerald-500/5 border border-emerald-500/20 rounded-lg p-3">
          <p className="text-xs text-gray-500 mb-1">Target ↑</p>
          <p className="text-base font-bold text-emerald-400 tabular-nums">
            {pred.target_price_up.toLocaleString('id-ID', { maximumFractionDigits: 2 })}
          </p>
        </div>
        <div className="bg-red-500/5 border border-red-500/20 rounded-lg p-3">
          <p className="text-xs text-gray-500 mb-1">Target ↓</p>
          <p className="text-base font-bold text-red-400 tabular-nums">
            {pred.target_price_down.toLocaleString('id-ID', { maximumFractionDigits: 2 })}
          </p>
        </div>
      </div>

      {/* Cross Signal Badges */}
      <div className="flex gap-2 flex-wrap">
        {isGolden && (
          <span className="px-2.5 py-1 rounded-full text-xs font-bold bg-amber-500/15 text-amber-400 border border-amber-500/30 animate-pulse">
            ✦ GOLDEN CROSS
          </span>
        )}
        {isDeath && (
          <span className="px-2.5 py-1 rounded-full text-xs font-bold bg-red-500/15 text-red-400 border border-red-500/30">
            ✦ DEATH CROSS
          </span>
        )}
        <span className={`px-2.5 py-1 rounded-full text-xs font-semibold border ${
          pred.signals.rsi_zone === 'OVERBOUGHT'
            ? 'bg-red-500/10 text-red-400 border-red-500/20'
            : pred.signals.rsi_zone === 'OVERSOLD'
            ? 'bg-emerald-500/10 text-emerald-400 border-emerald-500/20'
            : 'bg-gray-500/10 text-gray-400 border-gray-500/20'
        }`}>
          RSI: {pred.signals.rsi_zone}
        </span>
      </div>
    </div>
  );
}

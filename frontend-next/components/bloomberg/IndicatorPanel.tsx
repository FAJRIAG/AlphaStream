'use client';
// IndicatorPanel — Bloomberg-style panel for technical indicators (MA, RSI, ATR)
import { useStockStore } from '@/store/stockStore';

interface IndicatorPanelProps {
  symbol: string;
  isMaximized?: boolean;
  onToggleMaximize?: () => void;
}

export default function IndicatorPanel({ symbol, isMaximized, onToggleMaximize }: IndicatorPanelProps) {
  const ind = useStockStore((s) => s.indicators[symbol]);

  const fmt = (v: number | null | undefined, decimals = 2) =>
    v != null ? v.toFixed(decimals) : '—';

  const getRsiStatus = (rsi: number | null | undefined) => {
    if (rsi == null) return '—';
    if (rsi >= 70) return 'OVERBOUGHT';
    if (rsi <= 30) return 'OVERSOLD';
    return 'NEUTRAL';
  };

  const getRsiColor = (rsi: number | null | undefined) => {
    if (rsi == null) return 'var(--bb-text-muted)';
    if (rsi >= 70) return 'var(--bb-red)';
    if (rsi <= 30) return 'var(--bb-green)';
    return 'var(--bb-text)';
  };

  return (
    <div className="bb-panel h-full flex flex-col" style={{ border: 'none' }}>
      <div className="bb-panel-header flex items-center justify-between">
        <span className="bb-label" style={{ color: 'var(--bb-orange)', fontSize: '10px' }}>
          {symbol} · TECHNICALS
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

      <div className="bb-panel-body flex-1 pt-2 space-y-0">
        <div className="bb-data-row">
          <span className="bb-label">MA (20)</span>
          <span className="bb-mono text-xs font-bold" style={{ color: 'var(--bb-blue)' }}>
            {fmt(ind?.ma_20, 1)}
          </span>
        </div>
        <div className="bb-data-row">
          <span className="bb-label">MA (50)</span>
          <span className="bb-mono text-xs font-bold" style={{ color: 'var(--bb-cyan)' }}>
            {fmt(ind?.ma_50, 1)}
          </span>
        </div>
        <div className="bb-data-row">
          <span className="bb-label">RSI (14)</span>
          <span className="bb-mono text-xs font-bold" style={{ color: getRsiColor(ind?.rsi_14) }}>
            {fmt(ind?.rsi_14, 2)} ({getRsiStatus(ind?.rsi_14)})
          </span>
        </div>
        
        {/* RSI bar indicator custom design */}
        {ind?.rsi_14 != null && (
          <div className="py-2">
            <div className="rsi-bar-track">
              {/* Overbought limit marker (70%) */}
              <div className="absolute top-0 bottom-0 w-0.5 bg-red-800" style={{ left: '70%' }} />
              {/* Oversold limit marker (30%) */}
              <div className="absolute top-0 bottom-0 w-0.5 bg-emerald-800" style={{ left: '30%' }} />
              
              <div
                className="rsi-bar-fill"
                style={{
                  width: `${Math.min(Math.max(ind.rsi_14, 0), 100)}%`,
                  background: ind.rsi_14 >= 70
                    ? 'var(--bb-red)'
                    : ind.rsi_14 <= 30
                    ? 'var(--bb-green)'
                    : 'var(--bb-orange)',
                }}
              />
            </div>
            <div className="flex justify-between mt-1 text-[8px] font-bold text-neutral-600">
              <span>0</span>
              <span>30</span>
              <span>50</span>
              <span>70</span>
              <span>100</span>
            </div>
          </div>
        )}

        <div className="bb-data-row">
          <span className="bb-label">ATR (14)</span>
          <span className="bb-mono text-xs font-bold" style={{ color: 'var(--bb-orange)' }}>
            {fmt(ind?.atr_14, 2)}
          </span>
        </div>
      </div>
    </div>
  );
}

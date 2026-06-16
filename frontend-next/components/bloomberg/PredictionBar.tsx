'use client';
// PredictionBar — full-width bottom bar showing Quant Engine output
import { useStockStore } from '@/store/stockStore';
import type { TrendDirection } from '@/types/stock';

interface PredictionBarProps {
  symbol: string;
}

const DIR_STYLE: Record<TrendDirection, { label: string; color: string; bg: string }> = {
  BULLISH: { label: 'BULLISH', color: 'var(--bb-green)', bg: 'var(--bb-green-dim)' },
  BEARISH: { label: 'BEARISH', color: 'var(--bb-red)',   bg: 'var(--bb-red-dim)'   },
  NEUTRAL: { label: 'NEUTRAL', color: 'var(--bb-text-dim)', bg: 'transparent' },
};

export default function PredictionBar({ symbol }: PredictionBarProps) {
  const pred  = useStockStore((s) => s.predictions[symbol]);
  const ind   = useStockStore((s) => s.indicators[symbol]);

  if (!pred && !ind) {
    return (
      <div className="flex items-center justify-center px-4" style={{
        height: '36px',
        background: 'var(--bb-surface)',
        borderTop: '1px solid var(--bb-border)',
      }}>
        <span className="bb-label" style={{ color: 'var(--bb-text-muted)', fontSize: '10px' }}>
          QUANT ENGINE WARMING UP... (REQUIRES 20+ CANDLES)
        </span>
      </div>
    );
  }

  const dir = pred?.direction ?? 'NEUTRAL';
  const cfg = DIR_STYLE[dir];
  const probPct = pred ? Math.round(pred.probability * 100) : 50;

  return (
    <div className="flex items-center gap-6 px-4" style={{
      height: '36px',
      background: 'var(--bb-surface)',
      borderTop: '1px solid var(--bb-border)',
      overflowX: 'auto',
    }}>
      {/* Engine Label */}
      <span className="bb-label flex-shrink-0" style={{ color: 'var(--bb-orange)', fontSize: '10px' }}>
        QUANT ENGINE ·
      </span>

      {/* Direction Badge */}
      {pred && (
        <div className="flex items-center gap-2 flex-shrink-0 px-2 py-0.5 rounded-sm" style={{ background: cfg.bg, border: `1px solid ${cfg.color}33` }}>
          <span className="bb-mono text-xs font-black" style={{ color: cfg.color }}>{cfg.label}</span>
          <span className="bb-mono text-xs font-bold" style={{ color: cfg.color }}>{probPct}%</span>
        </div>
      )}

      {/* Target Prices */}
      {pred && (
        <>
          <div className="flex items-center gap-1.5 flex-shrink-0">
            <span className="bb-label" style={{ fontSize: '9px' }}>TARGET ↑</span>
            <span className="bb-mono text-xs font-bold" style={{ color: 'var(--bb-green)' }}>
              {pred.target_price_up.toLocaleString('id-ID', { maximumFractionDigits: 0 })}
            </span>
          </div>
          <div className="flex items-center gap-1.5 flex-shrink-0">
            <span className="bb-label" style={{ fontSize: '9px' }}>TARGET ↓</span>
            <span className="bb-mono text-xs font-bold" style={{ color: 'var(--bb-red)' }}>
              {pred.target_price_down.toLocaleString('id-ID', { maximumFractionDigits: 0 })}
            </span>
          </div>
        </>
      )}

      {/* Divider */}
      <div className="flex-shrink-0 w-px h-4" style={{ background: 'var(--bb-border2)' }} />

      {/* Indicators inline */}
      {ind && (
        <>
          <IndVal label="MA20"  value={ind.ma_20}  color="var(--bb-blue)" />
          <IndVal label="MA50"  value={ind.ma_50}  color="var(--bb-cyan)" />
          <IndVal label="RSI14" value={ind.rsi_14} color={
            ind.rsi_14 != null
              ? ind.rsi_14 >= 70 ? 'var(--bb-red)'
              : ind.rsi_14 <= 30 ? 'var(--bb-green)'
              : 'var(--bb-text-dim)'
            : 'var(--bb-text-dim)'
          } />
          <IndVal label="ATR14" value={ind.atr_14} color="var(--bb-orange)" />
        </>
      )}

      {/* Cross Signals */}
      {pred?.signals.is_golden_cross && (
        <span className="bb-mono text-xs font-black flex-shrink-0 px-2 py-0.5 rounded-sm bb-blink"
          style={{ color: 'var(--bb-orange)', background: 'var(--bb-orange-bg)', border: '1px solid var(--bb-orange)44' }}>
          ✦ GOLDEN CROSS
        </span>
      )}
      {pred?.signals.is_death_cross && (
        <span className="bb-mono text-xs font-black flex-shrink-0 px-2 py-0.5 rounded-sm"
          style={{ color: 'var(--bb-red)', background: 'var(--bb-red-dim)', border: '1px solid var(--bb-red)44' }}>
          ✦ DEATH CROSS
        </span>
      )}
      {pred?.signals.rsi_zone && pred.signals.rsi_zone !== 'NEUTRAL' && (
        <span className="bb-mono text-xs font-bold flex-shrink-0"
          style={{ color: pred.signals.rsi_zone === 'OVERBOUGHT' ? 'var(--bb-red)' : 'var(--bb-green)' }}>
          [{pred.signals.rsi_zone}]
        </span>
      )}
    </div>
  );
}

function IndVal({ label, value, color }: { label: string; value: number | null | undefined; color: string }) {
  return (
    <div className="flex items-center gap-1.5 flex-shrink-0">
      <span className="bb-label" style={{ fontSize: '9px' }}>{label}</span>
      <span className="bb-mono text-xs font-bold" style={{ color }}>
        {value != null ? value.toFixed(2) : '—'}
      </span>
    </div>
  );
}

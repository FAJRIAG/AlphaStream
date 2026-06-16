'use client';
// QuoteBox — Bloomberg-style OHLCV + change data panel (right panel)
import { useStockStore } from '@/store/stockStore';

interface QuoteBoxProps {
  symbol: string;
  isMaximized?: boolean;
  onToggleMaximize?: () => void;
}

function Row({ label, value, valueStyle }: { label: string; value: string; valueStyle?: React.CSSProperties }) {
  return (
    <div className="bb-data-row" style={{ borderColor: 'var(--bb-border)' }}>
      <span className="bb-label">{label}</span>
      <span className="bb-mono text-xs font-bold" style={valueStyle}>{value}</span>
    </div>
  );
}

export default function QuoteBox({ symbol, isMaximized, onToggleMaximize }: QuoteBoxProps) {
  const ticker = useStockStore((s) => s.tickers[symbol]);
  const candles = useStockStore((s) => s.candles[symbol]);

  const last = candles && candles.length > 0 ? candles[candles.length - 1] : undefined;
  const fmt = (v: number | undefined) =>
    v != null ? v.toLocaleString('id-ID', { maximumFractionDigits: 2 }) : '—';

  const isUp = (ticker?.change ?? 0) >= 0;
  const chgColor = { color: isUp ? 'var(--bb-green)' : 'var(--bb-red)' };

  return (
    <div className="bb-panel h-full flex flex-col" style={{ border: 'none' }}>
      <div className="bb-panel-header flex items-center justify-between">
        <span className="bb-label" style={{ color: 'var(--bb-orange)', fontSize: '10px' }}>
          {symbol} · QUOTE
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

      <div className="bb-panel-body flex-1 space-y-0">
        {/* Live Price */}
        <div className="py-3 text-center border-b" style={{ borderColor: 'var(--bb-border)' }}>
          <div className="bb-label mb-1">LAST PRICE</div>
          <div className="bb-mono font-black" style={{
            fontSize: '1.8rem',
            color: isUp ? 'var(--bb-green)' : 'var(--bb-red)',
          }}>
            {ticker?.price != null
              ? ticker.price.toLocaleString('id-ID', { maximumFractionDigits: 0 })
              : '—'}
          </div>
          {ticker && (
            <div className="bb-mono text-xs font-bold mt-1" style={chgColor}>
              {isUp ? '+' : ''}{ticker.change.toFixed(0)} ({isUp ? '+' : ''}{ticker.change_percent.toFixed(2)}%)
            </div>
          )}
        </div>

        {/* OHLCV from latest candle */}
        <div className="pt-2">
          <Row label="OPEN"  value={fmt(last?.open)}  />
          <Row label="HIGH"  value={fmt(last?.high)}  valueStyle={{ color: 'var(--bb-green)' }} />
          <Row label="LOW"   value={fmt(last?.low)}   valueStyle={{ color: 'var(--bb-red)' }} />
          <Row label="CLOSE" value={fmt(last?.close)} />
          <Row label="VOLUME"
            value={ticker?.volume != null
              ? (ticker.volume / 1_000_000).toFixed(2) + 'M'
              : last?.volume != null
              ? (last.volume / 1_000_000).toFixed(2) + 'M'
              : '—'}
          />
        </div>
      </div>
    </div>
  );
}

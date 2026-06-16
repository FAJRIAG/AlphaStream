'use client';
// BloombergHeader — Top header bar with Bloomberg orange branding and live WIB clock
import MarketClock from './MarketClock';

export default function BloombergHeader() {
  return (
    <div className="flex justify-between items-center px-4" style={{
      height: '34px',
      background: 'var(--bb-surface)',
      borderBottom: '1px solid var(--bb-orange-dim)',
      flexShrink: 0,
    }}>
      <div className="flex items-center gap-2">
        <span className="bb-mono text-xs font-black tracking-wider" style={{ color: 'var(--bb-orange)' }}>
          α ALPHASTREAM
        </span>
        <span className="bb-label" style={{
          fontSize: '8px',
          color: 'var(--bb-text-muted)',
          borderLeft: '1px solid var(--bb-border)',
          paddingLeft: '8px',
          marginLeft: '4px',
        }}>
          REAL-TIME QUANT ANALYTICS
        </span>
      </div>
      <MarketClock />
    </div>
  );
}

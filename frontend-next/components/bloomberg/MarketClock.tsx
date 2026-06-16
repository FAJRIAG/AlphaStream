'use client';
// MarketClock — real-time WIB clock for Bloomberg header
import { useState, useEffect } from 'react';

export default function MarketClock() {
  const [time, setTime] = useState('');
  const [date, setDate] = useState('');
  const [isMarketOpen, setIsMarketOpen] = useState(false);

  useEffect(() => {
    const update = () => {
      const now = new Date();
      // Convert to WIB (UTC+7)
      const wib = new Date(now.toLocaleString('en-US', { timeZone: 'Asia/Jakarta' }));
      const h = wib.getHours();
      const m = wib.getMinutes();
      const s = wib.getSeconds();

      setTime(
        `${String(h).padStart(2, '0')}:${String(m).padStart(2, '0')}:${String(s).padStart(2, '0')} WIB`
      );
      setDate(
        wib.toLocaleDateString('en-US', { weekday: 'short', month: 'short', day: '2-digit', year: 'numeric' })
      );

      // IDX trading: Mon–Fri 09:00–15:30 WIB
      const day = wib.getDay(); // 0=Sun, 6=Sat
      const timeNum = h * 100 + m;
      setIsMarketOpen(day >= 1 && day <= 5 && timeNum >= 900 && timeNum < 1530);
    };

    update();
    const interval = setInterval(update, 1000);
    return () => clearInterval(interval);
  }, []);

  return (
    <div className="flex items-center gap-4">
      <div className="text-right">
        <div className="bb-mono text-sm font-bold" style={{ color: 'var(--bb-orange)' }}>{time}</div>
        <div className="bb-label" style={{ fontSize: '9px' }}>{date}</div>
      </div>
      <div className="flex items-center gap-1.5 px-2 py-1 rounded" style={{
        background: isMarketOpen ? 'var(--bb-green-dim)' : 'rgba(85,85,85,0.2)',
        border: `1px solid ${isMarketOpen ? 'var(--bb-green)' : 'var(--bb-border)'}33`,
      }}>
        <span className={`w-1.5 h-1.5 rounded-full ${isMarketOpen ? 'bb-blink' : ''}`}
          style={{ background: isMarketOpen ? 'var(--bb-green)' : 'var(--bb-text-muted)' }} />
        <span className="bb-label" style={{ fontSize: '9px', color: isMarketOpen ? 'var(--bb-green)' : 'var(--bb-text-muted)' }}>
          IDX {isMarketOpen ? 'OPEN' : 'CLOSED'}
        </span>
      </div>
    </div>
  );
}

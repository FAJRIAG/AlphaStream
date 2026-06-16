'use client';
import { useStockStore } from '@/store/stockStore';

export default function Header() {
  const wsStatus = useStockStore((s) => s.wsStatus);

  const statusConfig = {
    connected: { color: 'bg-emerald-500', label: 'LIVE', pulse: true },
    connecting: { color: 'bg-amber-500', label: 'CONNECTING', pulse: true },
    disconnected: { color: 'bg-gray-600', label: 'OFFLINE', pulse: false },
    error: { color: 'bg-red-500', label: 'ERROR', pulse: false },
  };

  const cfg = statusConfig[wsStatus];

  return (
    <header className="h-14 border-b border-white/5 flex items-center justify-between px-6">
      {/* Logo */}
      <div className="flex items-center gap-3">
        <div className="w-7 h-7 rounded-lg bg-gradient-to-br from-indigo-500 to-violet-600 flex items-center justify-center">
          <span className="text-white font-black text-xs">α</span>
        </div>
        <span className="font-black text-white tracking-tight">AlphaStream</span>
        <span className="text-xs text-gray-600 font-medium">Enterprise Analytics</span>
      </div>

      {/* Connection Status */}
      <div className="flex items-center gap-2">
        <div className="relative flex items-center gap-2 bg-white/5 border border-white/5 rounded-full px-3 py-1.5">
          <span className="relative flex h-2 w-2">
            <span className={`${cfg.pulse ? 'animate-ping' : ''} absolute inline-flex h-full w-full rounded-full ${cfg.color} opacity-75`} />
            <span className={`relative inline-flex rounded-full h-2 w-2 ${cfg.color}`} />
          </span>
          <span className="text-xs font-semibold text-gray-400">{cfg.label}</span>
        </div>
      </div>
    </header>
  );
}

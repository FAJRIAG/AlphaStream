'use client';

import { useEffect, useState } from 'react';
import { stockApi } from '@/lib/api';
import type { BrokerSummary } from '@/types/stock';

interface BrokerSummaryPanelProps {
  symbol: string;
  hideHeader?: boolean;
}

export default function BrokerSummaryPanel({ symbol, hideHeader = false }: BrokerSummaryPanelProps) {
  const [summary, setSummary] = useState<BrokerSummary | null>(null);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!symbol || symbol === 'IHSG') return;

    const fetchSummary = async () => {
      setLoading(true);
      setError(null);

      try {
        const data = await stockApi.getBrokerSummary(symbol);
        setSummary(data);
      } catch (err) {
        console.error('Error fetching broker summary:', err);
        setError('FAILED TO FETCH BROKER DATA');
      } finally {
        setLoading(false);
      }
    };

    void fetchSummary();
  }, [symbol]);

  if (symbol === 'IHSG') {
    return (
      <div className="bb-panel h-full flex flex-col justify-center items-center text-[10px] text-neutral-500 font-mono">
        <span>BROKER DATA NOT APPLICABLE FOR INDICES</span>
      </div>
    );
  }

  const formatValue = (val: number): string => {
    if (val >= 1e12) return (val / 1e12).toFixed(2) + 'T';
    if (val >= 1e9) return (val / 1e9).toFixed(2) + 'B';
    if (val >= 1e6) return (val / 1e6).toFixed(1) + 'M';
    return val.toLocaleString('id-ID', { maximumFractionDigits: 0 });
  };

  const getNetStatusColor = (status?: string): string => {
    if (!status) return 'var(--bb-text-dim)';
    if (status.includes('ACCUMULATION')) return 'var(--bb-green)';
    if (status.includes('DISTRIBUTION')) return 'var(--bb-red)';
    return 'var(--bb-orange)';
  };

  if (loading) {
    return (
      <div className="bb-panel h-full flex flex-col justify-center items-center text-[10px] text-neutral-500 font-mono">
        <span className="bb-blink">FETCHING BROKER SUMMARY...</span>
      </div>
    );
  }

  if (error || !summary) {
    return (
      <div className="bb-panel h-full flex flex-col justify-center items-center text-[10px] text-red-500 font-mono">
        <span>{error || 'NO DATA AVAILABLE'}</span>
      </div>
    );
  }

  return (
    <div className="bb-panel h-full flex flex-col" style={{ border: 'none' }}>
      {/* Header Info */}
      {!hideHeader && (
        <div className="bb-panel-header">
          <span className="bb-label" style={{ color: 'var(--bb-orange)', fontSize: '10px' }}>
            {symbol} · BROKER SUMMARY (BANDARMOLOGY)
          </span>
          <button
            onClick={() => {
              setLoading(true);
              stockApi.getBrokerSummary(symbol).then((data) => {
                setSummary(data);
                setLoading(false);
              });
            }}
            className="ml-auto text-[9px] text-neutral-500 hover:text-white font-mono bg-neutral-900 border border-neutral-800 px-1.5 py-0.5 rounded cursor-pointer"
          >
            REFRESH
          </button>
        </div>
      )}

      {/* Bandar Detector Status Banner */}
      <div className="bg-[#111] border-b border-[#222] px-2 py-1.5 flex justify-between items-center text-[10px] font-mono">
        <div className="flex items-center gap-1.5">
          <span className="text-neutral-500">DETECTOR:</span>
          <span
            className="font-bold px-1 py-0.2 bg-black/50 border border-neutral-800 rounded"
            style={{ color: getNetStatusColor(summary.net_status) }}
          >
            {summary.net_status}
          </span>
        </div>
        <div className="flex items-center gap-1.5">
          <span className="text-neutral-500">FOREIGN BUY:</span>
          <span className="text-cyan-400 font-bold">{summary.foreign_buy_pct.toFixed(1)}%</span>
        </div>
      </div>

      {/* Side-by-side Buyer / Seller Tables */}
      <div className="flex-1 overflow-y-auto p-1.5 flex gap-1.5 text-[10px] font-mono">
        {/* BUYERS (LEFT) */}
        <div className="flex-1 flex flex-col border-r border-[#222] pr-1.5">
          <div className="flex justify-between text-[9px] font-bold text-neutral-500 border-b border-[#222] pb-0.5 mb-1">
            <span className="w-10">BUYER</span>
            <span className="w-10 text-right">AVG</span>
            <span className="w-12 text-right">VALUE</span>
          </div>
          <div className="flex-1 flex flex-col space-y-1.5">
            {summary.buyers.map((trade, idx) => (
              <div key={`buy-${trade.broker_code}-${idx}`} className="flex justify-between items-center py-0.5 leading-none">
                <div className="w-10 flex items-center gap-0.5">
                  <span
                    className="font-bold"
                    style={{ color: trade.nationality === 'F' ? 'var(--bb-cyan)' : 'var(--bb-text)' }}
                    title={trade.broker_name}
                  >
                    {trade.broker_code}
                  </span>
                  <span className="text-[8px] text-neutral-500 scale-90">{trade.nationality}</span>
                </div>
                <span className="w-10 text-right text-neutral-400">
                  {trade.avg_price.toLocaleString('id-ID', { maximumFractionDigits: 0 })}
                </span>
                <span className="w-12 text-right text-green-400 font-bold">
                  {formatValue(trade.value)}
                </span>
              </div>
            ))}
          </div>
        </div>

        {/* SELLERS (RIGHT) */}
        <div className="flex-1 flex flex-col pl-1.5">
          <div className="flex justify-between text-[9px] font-bold text-neutral-500 border-b border-[#222] pb-0.5 mb-1">
            <span className="w-10">SELLER</span>
            <span className="w-10 text-right">AVG</span>
            <span className="w-12 text-right">VALUE</span>
          </div>
          <div className="flex-1 flex flex-col space-y-1.5">
            {summary.sellers.map((trade, idx) => (
              <div key={`sell-${trade.broker_code}-${idx}`} className="flex justify-between items-center py-0.5 leading-none">
                <div className="w-10 flex items-center gap-0.5">
                  <span
                    className="font-bold"
                    style={{ color: trade.nationality === 'F' ? 'var(--bb-cyan)' : 'var(--bb-text)' }}
                    title={trade.broker_name}
                  >
                    {trade.broker_code}
                  </span>
                  <span className="text-[8px] text-neutral-500 scale-90">{trade.nationality}</span>
                </div>
                <span className="w-10 text-right text-neutral-400">
                  {trade.avg_price.toLocaleString('id-ID', { maximumFractionDigits: 0 })}
                </span>
                <span className="w-12 text-right text-red-400 font-bold">
                  {formatValue(trade.value)}
                </span>
              </div>
            ))}
          </div>
        </div>
      </div>
    </div>
  );
}

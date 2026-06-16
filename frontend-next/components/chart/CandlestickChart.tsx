'use client';

import React, { useEffect, useRef } from 'react';
import { createChart, ColorType, CandlestickSeries, HistogramSeries } from 'lightweight-charts';
import { useStockStore } from '@/store/stockStore';

interface CandlestickChartProps {
  symbol: string;
}

export default function CandlestickChart({ symbol }: CandlestickChartProps) {
  const containerRef = useRef<HTMLDivElement>(null);
  const chartRef = useRef<any>(null);
  const candlestickSeriesRef = useRef<any>(null);
  const volumeSeriesRef = useRef<any>(null);

  // Subscribe to candles state in the store reactively
  const candles = useStockStore((state) => state.candles[symbol]);

  // 1. Initialize the chart instance
  useEffect(() => {
    if (!containerRef.current) return;

    // Create the lightweight-charts instance
    const chart = createChart(containerRef.current, {
      width: containerRef.current.clientWidth || 600,
      height: containerRef.current.clientHeight || 300,
      layout: {
        background: { type: ColorType.Solid, color: '#090909' },
        textColor: '#888888',
        fontSize: 9,
      },
      grid: {
        vertLines: { color: '#141414' },
        horzLines: { color: '#141414' },
      },
      rightPriceScale: {
        borderColor: '#222222',
        scaleMargins: {
          top: 0.1,
          bottom: 0.2, // leave space for volume histogram
        },
      },
      timeScale: {
        borderColor: '#222222',
        timeVisible: true,
        secondsVisible: false,
      },
      crosshair: {
        vertLine: {
          color: '#555555',
          labelBackgroundColor: '#2a2a2a',
        },
        horzLine: {
          color: '#555555',
          labelBackgroundColor: '#2a2a2a',
        },
      },
    });

    // Add candlestick series using the new addSeries API (v5)
    const candlestickSeries = chart.addSeries(CandlestickSeries, {
      upColor: '#00d084',
      downColor: '#ff3b30',
      borderVisible: false,
      wickUpColor: '#00d084',
      wickDownColor: '#ff3b30',
    });

    // Add volume histogram series using the new addSeries API (v5)
    const volumeSeries = chart.addSeries(HistogramSeries, {
      priceFormat: {
        type: 'volume',
      },
      priceScaleId: '', // overlay mode
    });

    // Format volume scale margins (occupy bottom 20% height)
    volumeSeries.priceScale().applyOptions({
      scaleMargins: {
        top: 0.8,
        bottom: 0,
      },
    });

    chartRef.current = chart;
    candlestickSeriesRef.current = candlestickSeries;
    volumeSeriesRef.current = volumeSeries;

    // Responsive resize handler
    const handleResize = () => {
      if (containerRef.current && chartRef.current) {
        chartRef.current.resize(
          containerRef.current.clientWidth,
          containerRef.current.clientHeight
        );
      }
    };

    const resizeObserver = new ResizeObserver(handleResize);
    resizeObserver.observe(containerRef.current);

    return () => {
      resizeObserver.disconnect();
      chart.remove();
      chartRef.current = null;
      candlestickSeriesRef.current = null;
      volumeSeriesRef.current = null;
    };
  }, []);

  // 2. Synchronize chart data when candles or symbol changes
  useEffect(() => {
    if (!candlestickSeriesRef.current || !volumeSeriesRef.current) return;

    const items = candles ?? [];

    // Ensure data is sorted by timestamp (required by lightweight-charts)
    const sortedItems = [...items].sort(
      (a, b) => new Date(a.timestamp).getTime() - new Date(b.timestamp).getTime()
    );

    // Format to lightweight-charts data models
    const candleData = sortedItems.map((c) => ({
      time: Math.floor(new Date(c.timestamp).getTime() / 1000) as any,
      open: c.open,
      high: c.high,
      low: c.low,
      close: c.close,
    }));

    const volumeData = sortedItems.map((c) => ({
      time: Math.floor(new Date(c.timestamp).getTime() / 1000) as any,
      value: c.volume,
      color: c.close >= c.open ? 'rgba(0, 208, 132, 0.2)' : 'rgba(255, 59, 48, 0.2)',
    }));

    candlestickSeriesRef.current.setData(candleData);
    volumeSeriesRef.current.setData(volumeData);

    // Fit timescale layout so candles fill the canvas area
    if (sortedItems.length > 0 && chartRef.current) {
      chartRef.current.timeScale().fitContent();
    }
  }, [candles, symbol]);

  return (
    <div className="absolute inset-0 w-full h-full bg-[#090909] select-none" ref={containerRef} />
  );
}

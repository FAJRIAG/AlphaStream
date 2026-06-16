'use client';

import React, { useEffect, useRef } from 'react';

interface CandlestickChartProps {
  symbol: string;
}

export default function CandlestickChart({ symbol }: CandlestickChartProps) {
  const containerRef = useRef<HTMLDivElement>(null);
  const containerId = `tv-widget-${symbol.toLowerCase()}`;

  useEffect(() => {
    const scriptId = 'tradingview-widget-script';
    let script = document.getElementById(scriptId) as HTMLScriptElement;

    const createWidget = () => {
      if (typeof window === 'undefined') return;
      const tvWindow = window as any;

      if (tvWindow.TradingView && containerRef.current) {
        // Clear previous widget and insert a clean target container
        containerRef.current.innerHTML = `<div id="${containerId}" style="width: 100%; height: 100%;" />`;

        let tradingViewSymbol = symbol;
        if (symbol === 'IHSG') {
          tradingViewSymbol = 'IDX:COMPOSITE';
        } else if (!symbol.includes(':')) {
          tradingViewSymbol = `IDX:${symbol}`;
        }

        try {
          new tvWindow.TradingView.widget({
            autosize: true,
            symbol: tradingViewSymbol,
            interval: 'D',
            timezone: 'Asia/Jakarta',
            theme: 'dark',
            style: '1',
            locale: 'en',
            enable_publishing: false,
            hide_side_toolbar: false,
            allow_symbol_change: false,
            container_id: containerId,
            loading_screen: { backgroundColor: '#090909' },
            backgroundColor: '#090909',
            gridColor: '#141414',
          });
        } catch (err) {
          console.error('TradingView widget creation failed:', err);
        }
      }
    };

    if (!script) {
      script = document.createElement('script');
      script.id = scriptId;
      script.src = 'https://s3.tradingview.com/tv.js';
      script.type = 'text/javascript';
      script.async = true;
      script.onload = createWidget;
      document.head.appendChild(script);
    } else {
      if ((window as any).TradingView) {
        createWidget();
      } else {
        script.addEventListener('load', createWidget);
      }
    }

    return () => {

      if (script) {
        script.removeEventListener('load', createWidget);
      }
    };
  }, [symbol, containerId]);

  return (
    <div className="absolute inset-0 w-full h-full bg-[#090909] select-none" ref={containerRef} />
  );
}

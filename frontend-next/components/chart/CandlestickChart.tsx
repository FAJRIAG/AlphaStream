'use client';
import React, { useEffect, useRef } from 'react';

interface CandlestickChartProps {
  symbol: string;
}

type TradingViewWindow = Window & {
  TradingView?: {
    widget: new (options: Record<string, unknown>) => unknown;
  };
};

export default function CandlestickChart({ symbol }: CandlestickChartProps) {
  const containerRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const scriptId = 'tradingview-widget-script';
    let script = document.getElementById(scriptId) as HTMLScriptElement;

    const createWidget = () => {
      const tradingViewWindow = window as TradingViewWindow;

      if (typeof window !== 'undefined' && tradingViewWindow.TradingView && containerRef.current) {
        try {
          const keysToRemove: string[] = [];
          for (let i = 0; i < window.localStorage.length; i++) {
            const key = window.localStorage.key(i);
            if (key && (key.startsWith('tradingview') || key.includes('tradingview'))) {
              keysToRemove.push(key);
            }
          }
          keysToRemove.forEach((key) => window.localStorage.removeItem(key));
        } catch {
        }

        containerRef.current.innerHTML = '';

        const widgetId = `tradingview_${Math.random().toString(36).substring(2, 9)}`;
        const childDiv = document.createElement('div');
        childDiv.id = widgetId;
        childDiv.style.width = '100%';
        childDiv.style.height = '100%';
        containerRef.current.appendChild(childDiv);

        let tradingViewSymbol = symbol.includes(':') ? symbol : `IDX:${symbol}`;
        if (symbol === 'IHSG') {
          tradingViewSymbol = 'IDX:COMPOSITE';
        }

        new tradingViewWindow.TradingView.widget({
          width: '100%',
          height: '100%',
          symbol: tradingViewSymbol,
          interval: 'D',
          timezone: 'Asia/Jakarta',
          theme: 'dark',
          style: '1',
          locale: 'en',
          enable_publishing: false,
          hide_side_toolbar: false,
          allow_symbol_change: false,
          container_id: widgetId,
        });
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
      if ((window as TradingViewWindow).TradingView) {
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
  }, [symbol]);

  return (
    <div className="absolute inset-0 w-full h-full bg-black select-none" ref={containerRef} />
  );
}

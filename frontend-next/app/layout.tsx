import type { Metadata } from 'next';
import { Inter } from 'next/font/google';
import './globals.css';

const inter = Inter({
  subsets: ['latin'],
  variable: '--font-inter',
  display: 'swap',
});

export const metadata: Metadata = {
  title: 'AlphaStream Terminal — Real-Time IDX Analytics',
  description: 'Bloomberg-grade real-time stock terminal for IDX. Live OHLCV, MA, RSI, Golden Cross, and AI-powered prediction engine.',
  keywords: ['bloomberg terminal', 'IDX', 'saham', 'real-time', 'analisis teknikal'],
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en" className={inter.variable}>
      <head>
        <link rel="preconnect" href="https://fonts.googleapis.com" />
        <link href="https://fonts.googleapis.com/css2?family=JetBrains+Mono:wght@400;500;700&display=swap" rel="stylesheet" />
      </head>
      <body className="bg-black text-bloomberg-text antialiased" suppressHydrationWarning>
        {children}
      </body>
    </html>
  );
}

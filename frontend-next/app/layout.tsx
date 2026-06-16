import type { Metadata } from 'next';
import { Inter } from 'next/font/google';
import './globals.css';

const inter = Inter({
  subsets: ['latin'],
  variable: '--font-inter',
  display: 'swap',
});

export const metadata: Metadata = {
  metadataBase: new URL(process.env.NEXT_PUBLIC_SITE_URL || 'https://alphastream.co'),
  title: {
    default: 'AlphaStream Terminal — Real-Time IDX Analytics',
    template: '%s | AlphaStream Terminal',
  },
  description: 'Bloomberg-grade real-time stock terminal for IDX. Live OHLCV, technical indicators (MA, RSI), Golden Cross alerts, and AI-powered prediction engine.',
  keywords: [
    'bloomberg terminal', 'IDX', 'saham', 'real-time stock market', 
    'analisis teknikal', 'candlestick chart', 'RSI', 'moving average', 
    'Golden Cross', 'AI stock prediction', 'quant analytics', 'investment terminal'
  ],
  authors: [{ name: 'AlphaStream Team' }],
  creator: 'AlphaStream',
  publisher: 'AlphaStream',
  alternates: {
    canonical: '/',
  },
  openGraph: {
    title: 'AlphaStream Terminal — Real-Time IDX Analytics',
    description: 'Bloomberg-grade real-time stock terminal for IDX. Live OHLCV, technical indicators (MA, RSI), Golden Cross alerts, and AI-powered prediction engine.',
    url: 'https://alphastream.co',
    siteName: 'AlphaStream',
    locale: 'id_ID',
    type: 'website',
  },
  twitter: {
    card: 'summary_large_image',
    title: 'AlphaStream Terminal — Real-Time IDX Analytics',
    description: 'Bloomberg-grade real-time stock terminal for IDX. Live OHLCV, technical indicators (MA, RSI), Golden Cross alerts, and AI-powered prediction engine.',
    creator: '@alphastream',
  },
  robots: {
    index: true,
    follow: true,
    googleBot: {
      index: true,
      follow: true,
      'max-video-preview': -1,
      'max-image-preview': 'large',
      'max-snippet': -1,
    },
  },
};

const jsonLd = {
  '@context': 'https://schema.org',
  '@type': 'SoftwareApplication',
  'name': 'AlphaStream Terminal',
  'operatingSystem': 'All',
  'applicationCategory': 'FinancialApplication',
  'offers': {
    '@type': 'Offer',
    'price': '0.00',
    'priceCurrency': 'USD',
  },
  'description': 'Bloomberg-grade real-time stock terminal for IDX. Live OHLCV, technical indicators (MA, RSI), Golden Cross alerts, and AI-powered prediction engine.',
  'featureList': [
    'Real-time market price feed ticks',
    'Interactive technical indicators (MA, RSI)',
    'Dynamic stock prediction analysis',
    'IDX broker summaries'
  ],
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en" className={inter.variable}>
      <head>
        <link rel="preconnect" href="https://fonts.googleapis.com" />
        <link href="https://fonts.googleapis.com/css2?family=JetBrains+Mono:wght@400;500;700&display=swap" rel="stylesheet" />
        <script
          type="application/ld+json"
          dangerouslySetInnerHTML={{ __html: JSON.stringify(jsonLd) }}
        />
      </head>
      <body className="bg-black text-bloomberg-text antialiased" suppressHydrationWarning>
        {children}
      </body>
    </html>
  );
}

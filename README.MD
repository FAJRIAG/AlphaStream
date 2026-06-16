# AlphaStream 🚀

**AlphaStream** adalah platform analisis saham real-time dan engine prediksi otomatis kelas enterprise yang dibangun khusus untuk Bursa Efek Indonesia (IDX). Mengadopsi prinsip **Clean Architecture**, sistem ini dirancang modular, berkinerja tinggi (optimized untuk Mac M1), serta terbebas dari *memory leaks* pada aliran data streaming real-time.

---

## 🌟 Fitur Utama

1. **Real-Time Market Data Streaming**: Aliran data harga saham live dan pembaruan candlestick 1 menit melalui koneksi WebSocket berkecepatan tinggi dengan integrasi Yahoo Finance.
2. **Quantitative Engine (Pure & Deterministic)**:
   * **Simple Moving Average (SMA)**: MA-20 & MA-50.
   * **Relative Strength Index (RSI-14)**: Deteksi wilayah jenuh beli (*overbought*) dan jenuh jual (*oversold*).
   * **Average True Range (ATR-14)**: Mengukur volatilitas pasar.
   * **Pattern Detection**: Deteksi sinyal tren jangka menengah (*Golden Cross* & *Death Cross*).
   * **Trend Probability Scoring**: Perhitungan persentase probabilitas arah pergerakan harga.
   * **Volatility-Based Target Price**: Estimasi target harga naik/turun berbasis perkalian ATR.
3. **Broker Summary (Bandarmology)**:
   * Melacak akumulasi dan distribusi dari sekuritas lokal (YP, PD, CC, dll.) maupun asing (AK, CS, MS, dll.).
   * Integrasi siap pakai (*out-of-the-box*) dengan API eksternal Stockbit jika `STOCKBIT_TOKEN` diset.
4. **Best Buy Recommendations (Saham Bagus Dibeli)**: Panel pintar yang menyaring emiten berpredikat `BULLISH` dan mengurutkannya berdasarkan probabilitas pembelian tertinggi dari mesin kuantitatif.
5. **Bloomberg-Style Premium UI**: Antarmuka responsif bernuansa gelap khas terminal Bloomberg dengan grafik candlestick interaktif menggunakan pustaka resmi `tradingview lightweight-charts`.

---

## 📂 Struktur Proyek

### 1. Backend (`backend-go`)
Mengikuti struktur Clean Architecture secara ketat:
* `cmd/`: Titik masuk aplikasi (`server` untuk REST/WS API dan `seeder` untuk database master).
* `config/`: Memuat variabel lingkungan (`.env`) dan koneksi MySQL/WebSocket.
* `internal/domain/`:
  * `entity/`: Struktur data murni bisnis (*enterprise business rules*) tanpa dependensi luar.
  * `repository/`: Kontrak antarmuka (*interface*) untuk persistensi data.
* `internal/repository/`: Implementasi konkret repositori (MySQL & in-memory cache Ring Buffer).
* `internal/usecase/`: Logika bisnis inti (Stock & Quantitative Engine).
* `internal/delivery/`: Pengantar HTTP REST API (Gin) & WebSocket Handler.
* `migrations/`: Berkas SQL migrasi database secara bertahap.

### 2. Frontend (`frontend-next`)
* `app/`: Routing halaman (Next.js App Router).
* `components/`: UI Bloomberg reusable (Header, Chart, Watchlist, Broker Summary, dll).
* `hooks/`: Manajemen status WebSocket.
* `lib/`: Klien API untuk pemanggilan endpoint backend.
* `store/`: Manajemen state global menggunakan Zustand.
* `types/`: Kontrak antarmuka TypeScript yang sinkron dengan entitas Go.

---

## 🛠️ Persiapan & Instalasi

### Prasyarat
* Go 1.21+
* Node.js 18+
* MySQL (MAMP atau server MySQL lokal)

### 1. Setup Database
1. Buat database baru bernama `alphastream` di server MySQL Anda.
2. Jalankan berkas migrasi SQL yang berada di dalam folder `backend-go/migrations/` secara berurutan:
   * [001_create_stocks.sql](file:///Applications/MAMP/htdocs/AlphaStream/backend-go/migrations/001_create_stocks.sql) (Membuat tabel master saham)
   * [002_create_ohlcv.sql](file:///Applications/MAMP/htdocs/AlphaStream/backend-go/migrations/002_create_ohlcv.sql) (Membuat tabel riwayat harga)
   * [003_create_indicators.sql](file:///Applications/MAMP/htdocs/AlphaStream/backend-go/migrations/003_create_indicators.sql) (Membuat tabel indikator teknikal)

3. Konfigurasikan variabel lingkungan pada `.env` di direktori `backend-go/`:
   ```env
   DB_HOST=127.0.0.1
   DB_PORT=3306
   DB_USER=root
   DB_PASSWORD=root
   DB_NAME=alphastream
   ```

### 2. Menjalankan Seeder Emiten IDX
Unduh dan muat otomatis seluruh daftar emiten resmi dari Bursa Efek Indonesia ke dalam database Anda:
```bash
cd backend-go
go run cmd/seeder/main.go
```

---

## 🚀 Menjalankan Aplikasi

### Menjalankan Backend (Go)
1. Pasang seluruh dependensi Go:
   ```bash
   cd backend-go
   go mod tidy
   ```
2. Jalankan server backend:
   ```bash
   go run cmd/server/main.go
   ```
   Server akan berjalan secara lokal di port `8080`.

### Menjalankan Frontend (Next.js)
1. Pasang dependensi npm:
   ```bash
   cd frontend-next
   npm install
   ```
2. Jalankan server pengembangan Next.js:
   ```bash
   npm run dev
   ```
   Buka peramban Anda dan akses halaman di [http://localhost:3000](http://localhost:3000).

---

## 📡 REST API Dokumentasi

### 🔹 Saham (Stocks)
* **`GET /api/v1/stocks`**: Mengambil daftar seluruh saham aktif.
* **`GET /api/v1/stocks/recommendations`**: Mengambil daftar saham yang direkomendasikan beli (*BULLISH*).
* **`GET /api/v1/stocks/:symbol`**: Detail informasi emiten saham tertentu.
* **`GET /api/v1/stocks/:symbol/ohlcv`**: Riwayat harga candlestick (*1m/5m/1d*).
* **`GET /api/v1/stocks/:symbol/prediction`**: Hasil prediksi pergerakan arah harga & target harga.
* **`GET /api/v1/stocks/:symbol/broker-summary`**: Rangkuman akumulasi sekuritas (Bandarmology).

### 🔹 Indikator (Indicators)
* **`GET /api/v1/indicators/:symbol`**: Mengambil nilai indikator teknikal terhitung terakhir (MA-20, MA-50, RSI-14, ATR-14).
* **`GET /api/v1/indicators/:symbol/summary`**: Gabungan data indikator teknikal dan prediksi dalam satu payload.
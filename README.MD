Nama Proyek: AlphaStream
Deskripsi: Platform Analisis Saham Real-Time & Predictive Engine Kelas Enterprise.
Tech Stack Utama: Go (Golang), Next.js (App Router), TypeScript, Tailwind CSS, MySQL (MAMP Mac M1).

Bertindaklah sebagai Senior Software Architect, Principal Clean Coder, dan Core Data Engineer. Saya ingin Anda membuat kode program untuk proyek "AlphaStream" dengan standar Clean Code yang sangat ketat, terstruktur, dan menerapkan arsitektur modular/Clean Architecture. 

Sistem harus dioptimalkan untuk berjalan sangat ringan di Mac M1 RAM 8GB tanpa adanya memory leak pada sistem data streaming.

Berikut spesifikasi teknis dan standar penulisan kode yang HARUS Anda patuhi:

1. STANDAR CLEAN CODE & ARSITEKTUR BACKEND (Go):
   - Terapkan pemisahan layer yang jelas (Ketik kodenya secara modular):
     * Domain/Entity: Berisi struktur data (struct) dan interfaces bisnis inti (Enterprise Business Rules).
     * Repository Layer: Berisi semua logika query database MySQL menggunakan database transaction yang aman dan driver 'go-sql-driver/mysql'.
     * Usecase/Service Layer: Berisi logika bisnis murni, termasuk "Quantitative Engine" untuk menghitung indikator MA-20, MA-50, RSI (14), serta algoritma prediksi probabilitas tren (Golden Cross detection & volatility target price).
     * Delivery/Handler Layer: Berisi HTTP REST API (menggunakan Gin/Fiber) dan WebSocket Handler (menggunakan Gorilla WebSocket).
   - Patuhi prinsip SOLID. Gunakan Dependency Injection (DI) secara eksplisit via constructor function (Gunakan pattern `NewRepository`, `NewUsecase`, `NewHandler`).
   - Penanganan Concurrency harus sangat aman: Gunakan Mutex (`sync.RWMutex`) atau Go Channels untuk mengamankan data di dalam WebSocket Hub dari data race / race condition.
   - Tulis kode yang deskriptif, fungsi yang kecil (Single Responsibility), gunakan konstanta untuk nilai yang fix, dan tangani error secara elegan (bukan di-ignore).

2. STANDAR CLEAN CODE & ARSITEKTUR FRONTEND (Next.js):
   - Gunakan Next.js App Router, TypeScript (Strict Mode), dan Tailwind CSS.
   - Struktur folder harus teratur: Pisahkan antara `/components` (UI reusable), `/hooks` (logika WebSocket/fetch), `/context` atau `/store` (Zustand state management), dan `/app` (routing page).
   - Hindari memory leak pada WebSocket di Next.js: Pastikan koneksi WebSocket di-clean up / didisconnect dengan benar di dalam return function `useEffect` ketika komponen unmount.
   - Gunakan library resmi 'tradingview lightweight-charts' untuk merender chart candlestick secara live. Pastikan render grafik di-throttle/di-optimize agar tidak memicu re-render massal pada komponen UI lainnya (hindari UI freeze).
   - Implementasikan clean TypeScript interfaces untuk setiap tipe data yang diterima dari WebSocket backend Go.

Tolong berikan blueprint struktur folder proyek AlphaStream yang rapi dan terstruktur terlebih dahulu. Setelah itu, berikan source code lengkap dan ter-optimasi untuk sisi 'backend-go' mulai dari layer Domain dan Database Setup secara bertahap.
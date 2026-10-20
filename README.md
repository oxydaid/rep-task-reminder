# WhatsApp Reminder

Aplikasi reminder pribadi berbasis WhatsApp yang mengirimkan pesan otomatis sesuai jadwal seperti alarm aktivitas harian (absen kerja, minum obat, meeting, deadline, tugas harian).

Dibangun dengan arsitektur modern, handal, hemat resource, dan *production-ready*:
- **Frontend:** Vue 3 + Vite + TypeScript + Tailwind CSS (PWA Installable, Mobile-First, Design System `DESIGN.md`).
- **Backend:** Go + Fiber (`gofiber/fiber/v2`) + SQLite (`modernc.org/sqlite` pure Go tanpa CGO).
- **WhatsApp Bot:** whatsmeow (WhatsApp Web multi-device) dengan sesi persisten SQLite.
- **Scheduler:** In-process scheduler tahan restart dengan pencegahan pengiriman duplikat (atomic database constraint).
- **Keamanan:** PIN Dashboard ber-hash (bcrypt), rate-limiting login & auto lockout, JWT authentication, dan otorisasi ketat untuk perintah chat WhatsApp.

---

## 1. Fitur Utama

1. **Dashboard & PWA:**
   - Ringkasan reminder aktif, reminder berikutnya, dan statistik pengiriman.
   - CRUD Reminder lengkap: tambah, ubah, hapus, aktifkan, dan nonaktifkan.
   - Pilihan pengulangan jadwal:
     - **Sekali Saja**
     - **Setiap Hari (Harian)**
     - **Hari Tertentu** (contoh: Senin, Rabu, Jumat)
     - **Mingguan**
   - Dukungan zona waktu: `Asia/Jakarta` (WIB), `Asia/Makassar` (WITA), `Asia/Jayapura` (WIT).
   - Installable PWA di smartphone / browser desktop.

2. **Koneksi WhatsApp & Multi-Device:**
   - Scan QR Code langsung melalui dashboard web.
   - Sesi tersimpan persisten di SQLite (`data/whatsapp.db`); tidak perlu scan ulang setiap restart.
   - Indikator status real-time: *Terhubung*, *Menghubungkan...*, dan *Terputus*.
   - Fitur kirim pesan uji coba (*test message*), putus sementara (*disconnect*), dan hapus sesi (*unpair*).

3. **Pengelolaan via Chat WhatsApp (Bot Commands):**
   Pengguna dapat mengelola reminder langsung dari chat WhatsApp (hanya nomor resmi pemilik yang diotorisasi):
   - `help` / `bantuan` — Tampilkan daftar seluruh perintah.
   - `buat Absen kerja | 08:00 | setiap hari` — Buat reminder harian.
   - `buat Minum obat | 12:00 | setiap hari`
   - `buat Olahraga pagi | 06:30 | senin, rabu, jumat` — Buat reminder hari tertentu.
   - `buat Deadline Laporan | 2026-10-25 15:00 | sekali` — Buat reminder satu kali.
   - `daftar` / `list` — Tampilkan semua reminder dengan statusnya.
   - `detail <id>` — Lihat detail reminder (contoh: `detail 1`).
   - `edit <id> jam <HH:mm>` — Ubah jam reminder (contoh: `edit 1 jam 13:00`).
   - `aktifkan <id>` — Aktifkan reminder (contoh: `aktifkan 1`).
   - `nonaktifkan <id>` — Matikan reminder (contoh: `nonaktifkan 1`).
   - `hapus <id>` — Hapus reminder (contoh: `hapus 1`).

4. **Keamanan & Proteksi PIN:**
   - Dashboard dilindungi PIN (4–12 digit).
   - PIN disimpan dengan hash bcrypt, tidak pernah plaintext.
   - Rate limiting: Maksimal 5x percobaan gagal berturut-turut memicu penguncian otomatis (*lockout*) selama 5 menit.
   - Seluruh endpoint API privat dilindungi oleh JWT Middleware.

---

## 2. Struktur Proyek

```text
reminder/
├── backend/
│   ├── cmd/
│   │   └── server/          # Entry point server Go Fiber
│   ├── internal/
│   │   ├── api/             # Fiber Router, Middleware, & Handlers
│   │   ├── auth/            # PIN Hashing (bcrypt), Token JWT, Lockout
│   │   ├── database/        # SQLite WAL connection, migrations, VACUUM backup
│   │   ├── reminder/        # Reminder CRUD, recurrence calculation, delivery logs
│   │   ├── scheduler/       # Scheduler runner & duplicate prevention
│   │   └── whatsapp/        # whatsmeow client & bot command handler
│   └── go.mod
├── frontend/
│   ├── src/
│   │   ├── api/             # HTTP client helper
│   │   ├── components/      # UI components (Header, PIN, Modals, Tabs)
│   │   ├── composables/     # useAuth, useWhatsApp, useReminders, useLogs
│   │   └── types/           # TypeScript definitions
│   ├── public/              # PWA icons & manifest
│   ├── index.html
│   ├── vite.config.ts
│   └── package.json
├── data/                    # SQLite database & backup folder (volume persisten)
├── bin/                     # Output binary eksekusi
├── build.sh                 # Single script build (Frontend + Backend) untuk Linux/VPS
├── build.bat                # Single script build untuk Windows
├── Dockerfile               # Multi-stage Docker build
├── compose.yaml             # Docker Compose orchestration
├── ecosystem.config.cjs     # Konfigurasi PM2 process manager
├── DESIGN.md                # Design System specification
├── PRD.md                   # Product Requirements Document
├── .env.example
└── README.md
```

---

## 3. Menjalankan Aplikasi Secara Lokal (Development)

Anda dapat menjalankan **Backend (Go Fiber)** dan **Frontend (Vue 3 Vite)** secara bersamaan hanya dengan **1 perintah** atau **1 file**:

### Opsi 1: Menggunakan npm (Direkomendasikan)
```bash
npm run dev
```

### Opsi 2: Menggunakan Script (Linux / Mac / Git Bash)
```bash
./dev.sh
```

### Opsi 3: Menggunakan File Batch (Windows)
Cukup **klik ganda (double click)** file `dev.bat` atau jalankan di CMD:
```cmd
dev.bat
```

Kedua server akan aktif secara bersamaan di terminal dengan prefix log berwarna:
- **Frontend Web:** [http://localhost:5173](http://localhost:5173) (Vite HMR + proxy otomatis ke `/api`)
- **Backend API:** [http://localhost:8080](http://localhost:8080) (Fiber + WhatsApp Bot + Scheduler)

Tekan `Ctrl + C` untuk menghentikan kedua server sekaligus secara aman.

---

## 4. Build Otomatis (Single File Build)

Untuk mem-build Frontend sekaligus Backend menjadi 1 paket binary yang siap dijalankan di server:

### Di Linux / VPS:
```bash
chmod +x build.sh
./build.sh
```

### Di Windows:
```cmd
build.bat
```

Script ini akan:
1. Meng-compile aset frontend Vue 3 menjadi static bundle di `frontend/dist`.
2. Meng-compile backend Go menjadi binary mandiri di `bin/whatsapp-reminder`.
3. Menyiapkan folder data dan backup database.

---

## 5. Deployment ke Production

### A. Deployment Menggunakan Docker Compose (Direkomendasikan)

1. Pastikan Docker dan Docker Compose terinstall di server.
2. Clone repository dan siapkan `.env`:
   ```bash
   cp .env.example .env
   ```
3. Jalankan container:
   ```bash
   docker compose up -d --build
   ```
4. Periksa log:
   ```bash
   docker compose logs -f whatsapp-reminder
   ```
5. Akses dashboard di `http://IP_SERVER:8080`.
   Data database SQLite dan sesi WhatsApp disimpan aman di volume `./data`.

---

### B. Deployment Menggunakan PM2 di VPS

PM2 digunakan untuk memantau proses Go, auto-restart jika crash, dan logging.

1. Build aplikasi:
   ```bash
   chmod +x build.sh
   ./build.sh
   ```
2. Jalankan proses dengan PM2:
   ```bash
   pm2 start ecosystem.config.cjs
   ```
3. Perintah PM2 yang sering digunakan:
   ```bash
   # Melihat status aplikasi
   pm2 status

   # Melihat log output aplikasi
   pm2 logs whatsapp-reminder

   # Me-restart aplikasi
   pm2 restart whatsapp-reminder

   # Menyimpan daftar proses agar otomatis hidup saat server reboot
   pm2 save
   ```

*Penting:* Jangan menjalankan instance backend aktif ganda pada database yang sama untuk mencegah konflik scheduler dan WhatsApp session.

---

## 6. Setup Reverse Proxy & HTTPS (Nginx)

Untuk menghubungkan domain dan sertifikat SSL Let's Encrypt:

```nginx
server {
    server_name reminder.domainanda.com;

    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

Aktifkan HTTPS dengan Certbot:
```bash
certbot --nginx -d reminder.domainanda.com
```

---

## 7. Cara Scan QR WhatsApp Pertama Kali

1. Buka Dashboard Web pada browser Anda (`http://localhost:8080` atau domain Anda).
2. Jika pertama kali membuka, sistem meminta Anda membuat **PIN Dashboard** (4–12 digit). Masukkan dan konfirmasi PIN Anda.
3. Klik tombol **Hubungkan WhatsApp** atau ikon QR Code di bagian header.
4. Buka aplikasi WhatsApp di HP Anda.
5. Masuk ke **Menu (titik tiga) / Pengaturan** > **Perangkat Tertaut** > **Tautkan Perangkat**.
6. Arahkan kamera ke QR Code yang muncul di layar.
7. Dalam beberapa detik status akan berubah menjadi **WA Terhubung (🟢)**.
8. Selesai! Sesi WhatsApp disimpan persisten di `data/whatsapp.db`.

---

## 8. Backup dan Restore Database

### Membuat Backup Aman
Aplikasi menyediakan fungsi backup konsisten berbasis SQLite `VACUUM INTO`:
1. Melalui UI Dashboard: Buka tab **Pengaturan** > klik **Buat Backup Sekarang**.
2. Melalui API: Kirim `POST /api/backup` dengan Header Authorization. File backup akan disimpan di `data/backups/backup_YYYYMMDD_HHMMSS.db`.

### Restore Database
Jika ingin memulihkan database dari file backup:
1. Hentikan aplikasi:
   - Docker: `docker compose down`
   - PM2: `pm2 stop whatsapp-reminder`
2. Salin file backup ke lokasi database aktif:
   ```bash
   cp data/backups/backup_xxxx.db data/reminders.db
   ```
3. Nyalakan kembali aplikasi:
   - Docker: `docker compose up -d`
   - PM2: `pm2 start ecosystem.config.cjs`

---

## 9. Troubleshooting

- **QR Code tidak muncul / Kadaluwarsa:**  
  Klik tombol *Segarkan* atau *Tampilkan QR Code* pada modal koneksi WhatsApp di dashboard.
- **Pesan WhatsApp tidak terkirim:**  
  1. Periksa apakah status WhatsApp masih terhubung.
  2. Buka tab *Riwayat* di dashboard untuk membaca pesan error teknis.
  3. Coba kirim pesan uji coba dari modal WhatsApp untuk menguji koneksi socket.
  4. Pastikan nomor tujuan menggunakan format internasional yang valid (contoh: `08123456789` atau `628123456789`).
- **Lupa PIN Dashboard:**  
  PIN tersimpan dalam tabel `app_settings` di file `data/reminders.db`. Anda dapat mereset kunci PIN dengan menjalankan:
  ```bash
  sqlite3 data/reminders.db "DELETE FROM app_settings WHERE key IN ('pin_hash', 'failed_login_attempts', 'login_locked_until');"
  ```
  Setelah itu refresh dashboard untuk mengatur PIN baru.

---

## 10. Lisensi
Didistribusikan untuk penggunaan pribadi dan self-hosted sesuai spesifikasi `PRD.md`.

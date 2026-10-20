# PRD — WhatsApp Reminder

## 1. Overview
**Nama:** WhatsApp Reminder  
**Tujuan:** Aplikasi reminder pribadi mirip alarm yang mengirimkan pengingat melalui WhatsApp agar pengguna tidak lupa aktivitas penting, seperti absen kerja, minum obat, mengerjakan tugas, atau aktivitas harian lainnya.

**Konsep utama:** Alarm berbasis pesan WhatsApp, bukan alarm suara. Sistem mengirim pesan otomatis sesuai jadwal yang dibuat pengguna.

## 2. Tech Stack
- **Frontend:** Vue 3 + Vite + TypeScript
- **Styling:** Tailwind CSS
- **Backend:** Go + Fiber
- **WhatsApp:** whatsmeow (WhatsApp Web multi-device)
- **Database:** SQLite
- **PWA:** Installable, mobile-first, responsive
- **Design system:** `DESIGN.md`

## 3. Fitur Utama

### A. Dashboard
- Dashboard sederhana untuk melihat reminder aktif, reminder berikutnya, dan riwayat pengiriman.
- CRUD reminder: tambah, lihat, edit, hapus, aktifkan, dan nonaktifkan.
- Atur judul, deskripsi/pesan, tanggal, jam, dan zona waktu.
- Mendukung reminder satu kali maupun berulang:
  - Harian
  - Hari tertentu dalam seminggu
  - Interval mingguan
- Tampilkan status pengiriman: menunggu, terkirim, atau gagal.
- UI mobile-first dengan navigasi sederhana.

### B. WhatsApp Connection
- Hubungkan WhatsApp melalui QR Code yang ditampilkan di dashboard.
- Tampilkan status koneksi: terhubung, menghubungkan, atau terputus.
- Sediakan tombol untuk menghubungkan ulang dan memutus koneksi.
- Simpan sesi WhatsApp secara persisten agar tidak perlu scan QR setiap restart.
- Batasi koneksi ke satu akun WhatsApp.
- Nomor WhatsApp yang terhubung menjadi penerima reminder secara default, dengan opsi mengatur nomor tujuan jika diperlukan.

### C. Reminder Scheduler
- Backend menjalankan scheduler untuk memeriksa reminder yang jatuh tempo.
- Kirim pesan WhatsApp otomatis sesuai jadwal.
- Reminder berulang harus menghasilkan jadwal berikutnya secara benar.
- Cegah pengiriman ganda akibat restart server atau scheduler berjalan bersamaan.
- Simpan riwayat pengiriman dan error untuk keperluan debugging.
- Jika WhatsApp terputus, tandai pengiriman gagal atau tertunda dan sediakan mekanisme retry yang aman.
- Semua waktu mengikuti zona waktu yang dikonfigurasi pengguna, default `Asia/Jakarta`.
- Reminder hanya mengirim pesan; tidak menghasilkan suara atau mengklaim telah membunyikan alarm.

### D. Pengelolaan via Chat WhatsApp
Pengguna dapat mengelola reminder melalui chat dengan akun WhatsApp yang terhubung.

Contoh perintah:
- `help` — tampilkan daftar perintah.
- `buat Absen kerja | 08:00 | setiap hari`
- `buat Minum obat | 12:00 | setiap hari`
- `daftar` — tampilkan reminder.
- `detail 1` — lihat detail reminder.
- `edit 1 jam 13:00` — ubah jam reminder.
- `hapus 1` — hapus reminder.
- `aktifkan 1` — aktifkan reminder.
- `nonaktifkan 1` — nonaktifkan reminder.

Gunakan format perintah yang konsisten, validasi input, dan berikan konfirmasi sebelum perubahan penting. Nomor/akun yang diizinkan saja yang boleh menjalankan perintah pengelolaan. Jangan menafsirkan pesan bebas sebagai perintah tanpa validasi.

### E. PIN Dashboard
- Dashboard dilindungi PIN sebelum pengguna dapat mengakses data.
- PIN disimpan dalam bentuk hash, bukan teks biasa.
- Sediakan pengaturan/perubahan PIN dengan verifikasi PIN lama.
- Terapkan pembatasan percobaan login dan jeda setelah kegagalan berulang.
- Lindungi seluruh API privat, bukan hanya tampilan frontend.
- Gunakan sesi autentikasi yang aman dan kedaluwarsa otomatis.
- Jangan menyimpan PIN atau kredensial sensitif di localStorage.

### F. PWA
- Dapat di-install dari browser yang mendukung PWA.
- Sediakan manifest, ikon, dan service worker.
- Dashboard tetap dapat menampilkan data terakhir yang tersimpan ketika offline jika memungkinkan.
- Pengiriman reminder tetap dilakukan backend; PWA tidak wajib terbuka.
- Tampilkan status koneksi server dan WhatsApp dengan jelas.

## 4. Arsitektur

```text
Vue 3 + Vite + Tailwind (PWA)
             |
          REST API
             |
       Go + Fiber
        /   |    \
       /    |     \
  SQLite  Scheduler  whatsmeow
                     |
                  WhatsApp
                     |
             Chat Command Handler
                     |
                Go + Fiber
```

Gunakan satu sumber data, yaitu SQLite. Dashboard dan chat harus membaca serta mengubah reminder yang sama melalui service layer backend, bukan memiliki logika bisnis terpisah.

## 5. Struktur Proyek

```text
whatsapp-reminder/
├── frontend/
│   ├── src/
│   ├── public/
│   └── package.json
├── backend/
│   ├── cmd/
│   ├── internal/
│   │   ├── api/
│   │   ├── auth/
│   │   ├── database/
│   │   ├── reminder/
│   │   ├── scheduler/
│   │   └── whatsapp/
│   ├── migrations/
│   └── go.mod
├── DESIGN.md
├── PRD.md
└── README.md
```

Struktur boleh disesuaikan jika ada alasan teknis yang jelas, tetapi tetap modular dan mudah dipelihara.

## 6. Database Minimal

### `reminders`
- `id`
- `title`
- `message`
- `scheduled_at`
- `recurrence_type`
- `recurrence_days`
- `timezone`
- `is_active`
- `created_at`
- `updated_at`

### `delivery_logs`
- `id`
- `reminder_id`
- `scheduled_for`
- `status`
- `attempts`
- `sent_at`
- `error_message`
- `created_at`

Gunakan tabel terpisah atau mekanisme persisten yang sesuai untuk sesi WhatsApp, konfigurasi PIN, dan kontrol scheduler. Tambahkan constraint unik yang diperlukan untuk mencegah pengiriman duplikat.

## 7. UI/UX dan DESIGN.md
Buat `DESIGN.md` sebelum implementasi UI. Dokumen ini menjadi acuan desain seluruh frontend.

Ketentuan:
- Mobile-first, minimalis, bersih, dan mudah digunakan.
- Fokus pada reminder berikutnya, status WhatsApp, dan tombol tambah reminder.
- Form reminder singkat dan mudah dipahami.
- Gunakan warna status yang konsisten untuk aktif, gagal, dan terkirim.
- Sediakan loading, empty state, error state, dan konfirmasi penghapusan.
- Hindari animasi berlebihan, layout ramai, dan komponen yang tidak diperlukan.
- Gunakan komponen reusable dan konsistensi spacing, typography, warna, serta radius.

## 8. Keamanan dan Reliabilitas
- Batasi pengelolaan reminder melalui chat ke akun WhatsApp pemilik yang telah diotorisasi.
- Validasi seluruh input di backend.
- Gunakan parameterized query SQLite.
- Lindungi API dengan autentikasi dan validasi otorisasi.
- Jangan mengekspos QR pairing, sesi WhatsApp, atau kredensial ke log publik.
- Pastikan scheduler tahan restart, tidak mengirim reminder dua kali, dan tidak memblokir API.
- Tangani disconnect, reconnect, error pengiriman, serta migrasi database.
- Ingat bahwa pengiriman WhatsApp bergantung pada konektivitas perangkat/sesi, jaringan, dan layanan WhatsApp; sistem tidak boleh menjanjikan pengiriman tepat waktu secara absolut.

## 9. Deployment dan Production

sediakan 1 file berisi untuk build Backend sekaligus Frontend dengan mudah baik untuk docker maupun native (vps / pm2)

### A. Docker
- Sediakan `Dockerfile` untuk frontend dan backend serta `compose.yaml` untuk menjalankan layanan.
- Frontend dibangun menjadi static assets dan disajikan melalui web server ringan.
- Backend Go berjalan sebagai service API, scheduler, dan WhatsApp bot dalam satu proses.
- Gunakan Docker volumes untuk persistensi database SQLite, sesi whatsmeow, dan data penting lainnya.
- Gunakan environment variables atau `.env` untuk konfigurasi; jangan memasukkan secrets ke image.
- Sediakan `.dockerignore` dan health check.
- Pastikan data tetap aman saat container restart atau image diperbarui.
- Jangan menjalankan beberapa instance backend aktif bersamaan pada database/sesi yang sama karena dapat menyebabkan konflik scheduler atau WhatsApp.

### B. PM2
- Sediakan konfigurasi `ecosystem.config.cjs` untuk mengelola proses backend Go menggunakan PM2.
- PM2 digunakan untuk menjalankan, memantau, dan me-restart executable Go ketika proses crash.
- Build binary Go terlebih dahulu sebelum dijalankan oleh PM2.
- Konfigurasikan auto-restart, working directory, environment, dan log output.
- Dokumentasikan penggunaan `pm2 start`, `pm2 status`, `pm2 logs`, `pm2 restart`, dan `pm2 save`.
- Jangan menjalankan proses backend yang sama secara bersamaan melalui Docker dan PM2.

**Catatan:** PM2 adalah process manager, bukan pengganti Docker. Pilih Docker Compose atau PM2 sebagai metode utama deployment agar tidak ada proses ganda.

### C. Production Requirements
- Gunakan multi-stage Docker build dan image sekecil mungkin.
- Jalankan service dengan user non-root jika memungkinkan.
- Gunakan reverse proxy HTTPS seperti Nginx atau Caddy jika aplikasi diakses melalui internet.
- Jangan mengekspos port database SQLite atau endpoint internal yang tidak diperlukan.
- Validasi konfigurasi wajib saat startup dan hentikan proses jika konfigurasi kritis tidak valid.
- Terapkan graceful shutdown agar scheduler berhenti dengan aman dan database ditutup dengan benar.
- Sediakan endpoint health check untuk memeriksa kondisi backend.
- Pisahkan konfigurasi development dan production.
- Terapkan timeout HTTP, batas ukuran request, serta pembatasan percobaan autentikasi.
- Pastikan log tidak berisi PIN, QR pairing, kredensial, atau data sesi sensitif.

### D. Backup dan Recovery
- Sediakan prosedur backup SQLite yang aman dan konsisten, menggunakan SQLite Backup API atau mekanisme yang sesuai.
- Backup juga data sesi WhatsApp secara aman agar proses pairing tidak perlu diulang jika memungkinkan.
- Dokumentasikan prosedur restore dan pengujian pemulihan.
- Jangan menyalin database SQLite yang sedang aktif hanya dengan mengandalkan file copy jika berisiko menghasilkan backup tidak konsisten.
- Simpan backup di luar container dan, jika memungkinkan, di lokasi terpisah.

### E. Dokumentasi Deployment
README.md harus mencakup:
- Persyaratan server dan instalasi dependensi.
- Konfigurasi `.env` beserta contoh `.env.example` tanpa secrets.
- Deployment menggunakan Docker Compose.
- Deployment alternatif menggunakan PM2.
- Setup domain dan HTTPS jika diperlukan.
- Cara scan QR WhatsApp pertama kali.
- Restart, update aplikasi, melihat log, backup, dan restore.
- Troubleshooting koneksi WhatsApp dan kegagalan pengiriman reminder.

### F. Kriteria Production-Ready
- [ ] Aplikasi dapat dijalankan ulang tanpa kehilangan data.
- [ ] Scheduler tidak mengirim reminder duplikat setelah restart.
- [ ] Sesi WhatsApp tersimpan persisten dan dapat dipulihkan.
- [ ] Kegagalan WhatsApp tercatat dan dapat ditangani.
- [ ] Backup dan restore telah diuji.
- [ ] Autentikasi PIN dan otorisasi chat aktif.
- [ ] Health check, logging, dan graceful shutdown berfungsi.
- [ ] Build, test, dan deployment berhasil mengikuti dokumentasi.


##10. Instruksi Implementasi untuk AI Agent
1. Baca PRD dan buat `DESIGN.md` terlebih dahulu.
2. Susun implementation plan singkat sebelum menulis kode.
3. Inisialisasi frontend dan backend, termasuk `go.mod`.
4. Implementasikan database dan service layer sebelum API dan UI.
5. Implementasikan autentikasi PIN dan otorisasi chat sejak awal.
6. Implementasikan koneksi QR whatsmeow, scheduler persisten, dan pencegahan duplikasi.
7. Bangun dashboard dan integrasikan CRUD dengan backend.
8. Implementasikan perintah chat menggunakan service layer yang sama.
9. Tambahkan unit test untuk scheduler, recurrence, autentikasi, dan command handler.
10. Jalankan formatter, test, dan build; perbaiki semua error sebelum menyatakan selesai.

**Prinsip:** Simpel tetapi andal, modular, aman, hemat resource, dan tidak over-engineered. Jangan mengimplementasikan fitur di luar PRD tanpa kebutuhan yang jelas.

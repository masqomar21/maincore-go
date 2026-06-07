# 01 — Overview & Cara Menjalankan

## 📌 Tentang Proyek

**MainCore Go** adalah template backend REST API production-ready yang dibangun menggunakan:

| Teknologi | Fungsi |
|-----------|--------|
| **Go (Golang)** | Bahasa pemrograman utama |
| **Gin** | HTTP Framework |
| **GORM** | ORM untuk PostgreSQL |
| **PostgreSQL** | Database utama |
| **Redis** | Session cache & task queue |
| **Asynq** | Background job queue (Redis-backed) |
| **AWS S3** | Penyimpanan file (bisa diganti dengan S3-compatible seperti MinIO) |
| **Socket.IO** | Realtime WebSocket |
| **JWT** | Autentikasi token |
| **Air** | Hot reload saat development |

---

## 🚀 Cara Menjalankan

### 1. Prasyarat

Pastikan sudah terinstall:
- Go `>= 1.21`
- PostgreSQL
- Redis

### 2. Setup Environment

```bash
cp .env.example .env
```

Edit `.env` sesuai kebutuhan (lihat: [02-config.md](./02-config.md)).

### 3. Jalankan dengan Make

```bash
# Development (hot reload dengan Air)
make dev

# Jalankan migration database
make migrate

# Jalankan seeder database
make seed

# Build semua binary
make build

# Lihat semua perintah yang tersedia
make help
```

### 4. Perintah Make Lengkap

| Perintah | Deskripsi |
|----------|-----------|
| `make dev` | Jalankan server dengan hot-reload (Air) |
| `make build` | Build binary: `server-build`, `migrate-build`, `seed-build` |
| `make migrate` | Jalankan `go run cmd/migrate/main.go` |
| `make seed` | Jalankan `go run cmd/seed/main.go` |
| `make tidy` | Jalankan `go mod tidy` |
| `make help` | Tampilkan bantuan |

---

## 🏗️ Arsitektur Aplikasi

```
Request HTTP
    │
    ▼
[Gin Router]
    │
    ├─► [Middleware: CORS]
    ├─► [Middleware: Auth JWT]          ← Verifikasi token
    ├─► [Middleware: GeneratePermission] ← Load permission dari Redis/DB
    ├─► [Middleware: RequirePermission]  ← Cek izin akses
    │
    ▼
[Controller]                            ← Handle logika bisnis
    │
    ├─► [GORM → PostgreSQL]             ← Operasi database
    ├─► [Redis]                         ← Cache
    ├─► [Asynq Queue]                   ← Background jobs
    └─► [S3 / Socket.IO]               ← File storage & realtime
```

---

## 📦 Entry Point

### `cmd/api/main.go`
Server utama. Menginisialisasi semua layanan:
- Config → Database → Redis → S3 → Queue → Socket.IO
- Mendukung flag `--auto-migrate` untuk auto migrate saat startup
- Graceful shutdown saat menerima sinyal `SIGTERM` / `SIGINT`

### `cmd/migrate/main.go`
Menjalankan database migration. Jika database tidak ditemukan, otomatis membuatnya terlebih dahulu.

### `cmd/seed/main.go`
Menjalankan seeder untuk mengisi data awal database.

---

## 🔄 Hot Reload dengan Air

File konfigurasi Air ada di `.air.toml`. Air secara otomatis me-rebuild dan merestart server setiap kali ada perubahan file `.go`, `.html`, atau `.tmpl`.

```toml
# Konfigurasi build Air
cmd = "go build -o ./tmp/main ./cmd/api"
bin = "./tmp/main"
include_ext = ["go", "tpl", "tmpl", "html"]
exclude_dir = ["assets", "tmp", "vendor", "testdata"]
```

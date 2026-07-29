# 📚 Dokumentasi Proyek: MainCore Go

Selamat datang di dokumentasi resmi **MainCore Go** — sebuah REST API backend yang dibangun dengan Go (Golang), Gin framework, PostgreSQL, Redis, dan berbagai layanan modern lainnya.

---

## 📁 Struktur Dokumentasi

| File | Isi |
|------|-----|
| [01-overview.md](./01-overview.md) | Gambaran umum proyek, arsitektur, dan cara menjalankan |
| [02-config.md](./02-config.md) | Konfigurasi environment dan inisialisasi database & Redis |
| [03-models.md](./03-models.md) | Skema database (semua model/struct) |
| [04-migrations.md](./04-migrations.md) | Cara kerja migration dan seeder |
| [05-controllers.md](./05-controllers.md) | Semua controller beserta penjelasan endpoint |
| [06-middlewares.md](./06-middlewares.md) | Middleware: Auth, CORS, Permission, File Upload |
| [07-routes.md](./07-routes.md) | Daftar lengkap route API |
| [08-services.md](./08-services.md) | Services: Queue (Asynq), S3 Storage, Socket.IO |
| [09-utilities.md](./09-utilities.md) | Utility helpers: JWT dan Response |

---

## 🗂️ Struktur Folder Proyek

```
maincore_go/
├── cmd/
│   ├── api/          → Entry point server utama
│   ├── migrate/      → Entry point database migration
│   └── seed/         → Entry point database seeder
├── config/           → Konfigurasi (env, database, redis)
├── controllers/      → Handler HTTP request
├── middlewares/      → Middleware (auth, cors, permission, upload)
├── models/           → Model/struct database GORM
├── routes/           → Registrasi route API
├── services/         → Layanan eksternal (S3, Queue, Socket)
├── utilities/        → Helper umum (JWT, Response)
├── doc/              → Dokumentasi proyek (folder ini)
├── .env.example      → Contoh konfigurasi environment
├── Makefile          → Perintah make untuk development
└── go.mod            → Daftar dependensi Go
```

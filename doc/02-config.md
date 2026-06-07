# 02 — Konfigurasi (Config)

Paket `config` bertanggung jawab untuk memuat konfigurasi environment dan menginisialisasi koneksi ke database PostgreSQL dan Redis.

---

## 📄 `config/config.go`

### Struct `Config`

Menyimpan seluruh konfigurasi aplikasi yang dibaca dari environment variables.

```go
type Config struct {
    AppName           string
    AppVersion        string
    AppEnv            string   // "development" | "production"
    Port              string   // Port HTTP server, default: "3000"
    DatabaseURL       string   // Koneksi PostgreSQL
    RedisHost         string
    RedisPort         string
    JWTSecret         string   // Secret untuk signing JWT
    S3Endpoint        string
    S3Bucket          string
    S3Region          string
    S3AccessKeyID     string
    S3SecretAccessKey string
    S3ForcePathStyle  bool
    FileSaveToBucket  bool
}
```

### Fungsi `InitConfig()`

Memuat konfigurasi dari file `.env` menggunakan library `godotenv`, kemudian jika file tidak ditemukan, membaca langsung dari sistem environment variable.

```go
func InitConfig()
```

**Urutan prioritas:**
1. System environment variable
2. File `.env` (di direktori root proyek)
3. Nilai default hardcoded

### Environment Variables

Semua environment variable dengan nilai defaultnya:

| Variable | Default | Keterangan |
|----------|---------|------------|
| `APP_NAME` | `Starter Kit` | Nama aplikasi |
| `APP_VERSION` | `1.0.0` | Versi aplikasi |
| `APP_ENV` | `development` | Mode (development / production) |
| `PORT` | `3000` | Port server HTTP |
| `DATABASE_URL` | `host=localhost user=root ...` | Koneksi PostgreSQL |
| `REDIS_HOST` | `localhost` | Host Redis |
| `REDIS_PORT` | `6379` | Port Redis |
| `AUTH_JWT_SECRET` | `secret` | Secret JWT (**wajib diganti di production**) |
| `S3_ENDPOINT` | _(kosong)_ | Endpoint S3 |
| `S3_BUCKET` | _(kosong)_ | Nama bucket S3 |
| `S3_REGION` | `ap-southeast-1` | Region S3 |
| `S3_ACCESS_KEY_ID` | _(kosong)_ | Access Key S3 |
| `S3_SECRET_ACCESS_KEY` | _(kosong)_ | Secret Key S3 |
| `S3_FORCE_PATH_STYLE` | `false` | Path style S3 (untuk MinIO: `true`) |
| `FILE_SAVE_TO_BUCKET` | `true` | Apakah file disimpan ke bucket |

### Contoh `.env`

```env
APP_NAME="MainCore Backend"
APP_VERSION="1.0.0"
APP_ENV=development
PORT=3000

DATABASE_URL=postgres://root:root@localhost:5432/maincore_db?sslmode=disable

REDIS_HOST=localhost
REDIS_PORT=6379

AUTH_JWT_SECRET=your_super_secret_jwt_key

S3_ENDPOINT=https://s3.ap-southeast-1.amazonaws.com
S3_BUCKET=your-bucket-name
S3_REGION=ap-southeast-1
S3_ACCESS_KEY_ID=your_access_key_id
S3_SECRET_ACCESS_KEY=your_secret_access_key
S3_FORCE_PATH_STYLE=false
FILE_SAVE_TO_BUCKET=true
```

---

## 📄 `config/db.go`

Mengelola koneksi ke PostgreSQL dan Redis.

### Variabel Global

```go
var DB *gorm.DB          // Instance database PostgreSQL
var RedisClient *redis.Client  // Instance Redis
```

### Fungsi `InitDatabase()`

Menginisialisasi koneksi database GORM ke PostgreSQL.

```go
func InitDatabase()
```

**Alur kerja:**
1. Memanggil `createDatabaseIfNotExist(dsn)` — otomatis membuat database jika belum ada
2. Membuka koneksi GORM ke database target
3. Menyimpan instance ke variabel global `DB`

> ✅ **Fitur Auto-Create Database**: Jika database yang dikonfigurasi belum ada, sistem akan otomatis menghubungi database `postgres` default dan menjalankan `CREATE DATABASE`.

### Fungsi `createDatabaseIfNotExist(dsn string) error`

Helper internal yang mengecek dan membuat database jika belum ada.

**Mendukung dua format DSN:**

| Format | Contoh |
|--------|--------|
| URL Format | `postgres://user:pass@localhost:5432/mydb?sslmode=disable` |
| Key-Value Format | `host=localhost user=root password=root dbname=mydb port=5432` |

**Alur kerja:**
1. Parse DSN untuk mengekstrak `dbname` dan membuat `defaultDSN` (mengganti dbname → `postgres`)
2. Koneksi ke database `postgres` default
3. Query `pg_database` untuk cek keberadaan database target
4. Jika tidak ada → jalankan `CREATE DATABASE "nama_db"`

### Fungsi `InitRedis()`

Menginisialisasi koneksi ke Redis menggunakan library `go-redis`.

```go
func InitRedis()
```

**Alur kerja:**
1. Membentuk address dari `REDIS_HOST:REDIS_PORT`
2. Melakukan `PING` untuk validasi koneksi
3. Menyimpan instance ke variabel global `RedisClient`

### Fungsi `splitKV(dsn string) []string`

Helper untuk parsing format DSN Key-Value dengan benar, termasuk menangani nilai yang mengandung spasi di dalam tanda kutip (single atau double quote).

```go
// Contoh input:
// "host=localhost user=root password='my secret' dbname=mydb"
// Output: ["host=localhost", "user=root", "password='my secret'", "dbname=mydb"]
```

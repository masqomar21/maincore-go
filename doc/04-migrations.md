# 04 — Migration & Seeder

---

## 📄 `cmd/migrate/main.go` — Entry Point Migration

File ini adalah entry point untuk menjalankan migrasi database.

```go
func main() {
    // 1. Load konfigurasi dari .env
    config.InitConfig()

    // 2. Inisialisasi koneksi database
    //    (otomatis membuat database jika belum ada)
    config.InitDatabase()

    // 3. Jalankan AutoMigrate semua model
    models.AutoMigrate(config.DB)
}
```

**Cara menjalankan:**

```bash
make migrate
# atau
go run cmd/migrate/main.go
```

---

## 📄 `models/autoMigrate.go` — Logika AutoMigrate

Berisi fungsi `AutoMigrate` yang mendaftarkan seluruh model ke GORM untuk disinkronkan ke database.

### Fungsi `AutoMigrate(db *gorm.DB)`

```go
func AutoMigrate(db *gorm.DB) {
    err := db.AutoMigrate(
        &Permission{},
        &Role{},
        &RolePermission{},
        &User{},
        &Otp{},
        &Session{},
        &Logger{},
        &Notification{},
        &NotificationUser{},
        &WebPushSubscription{},
    )
    if err != nil {
        log.Fatal("Failed to auto migrate database:", err)
    }
    log.Println("Database auto-migration completed successfully.")
}
```

**Apa yang dilakukan GORM AutoMigrate:**
- Membuat tabel baru jika belum ada
- Menambah kolom baru jika ada penambahan field di struct
- **Tidak** menghapus kolom yang sudah ada (non-destructive)
- **Tidak** mengubah tipe kolom yang sudah ada

> ⚠️ Untuk operasi seperti **hapus kolom, rename kolom, atau ubah tipe kolom**, gunakan sistem **Manual Migration** yang sudah tersedia — lihat section di bawah.

### Menambah Model Baru ke AutoMigrate

1. Buat struct model baru di `models/models.go`
2. Tambahkan pointer struct tersebut ke dalam list `db.AutoMigrate(...)` di `models/autoMigrate.go`
3. Jalankan `make migrate`

---

## 📄 `models/migrations.go` — Manual Migration System

File ini menyediakan sistem **versioned manual migration** untuk operasi schema yang tidak bisa dilakukan AutoMigrate secara otomatis, seperti menghapus kolom, rename kolom, atau mengubah tipe kolom — **semuanya via kode Go, tanpa perlu akses database secara manual**.

### Cara Kerja

Setiap migration step didaftarkan dalam slice `ManualMigrations`. Saat `make migrate` dijalankan, sistem akan:

1. Membaca semua step dalam `ManualMigrations`
2. Mengecek tabel `migration_logs` di database — apakah step tersebut sudah pernah dijalankan
3. Jika **belum** → jalankan fungsi `Up()` dalam transaksi database
4. Jika **berhasil** → catat nama step ke `migration_logs` agar tidak dijalankan lagi
5. Jika **gagal** → transaksi di-rollback, program berhenti dengan pesan error

```
make migrate
    │
    ├─► AutoMigrate()          → Sinkronisasi struct model ke tabel
    │
    └─► RunManualMigrations()
            │
            ├─► Buat tabel migration_logs jika belum ada
            │
            └─► Untuk setiap step di ManualMigrations:
                    ├─► Cek migration_logs: sudah dijalankan? → SKIP
                    └─► Belum? → Jalankan Up() dalam transaksi → Catat ke migration_logs
```

### Struct yang Terlibat

```go
// MigrationLog — tabel internal pencatat history migration
type MigrationLog struct {
    ID        uint      `gorm:"primaryKey"`
    Name      string    `gorm:"uniqueIndex;not null"` // Nama unik step
    CreatedAt time.Time
}

// MigrationStep — definisi satu langkah migration manual
type MigrationStep struct {
    Name string                    // Identifier unik, format: "YYYY-MM-DD_deskripsi"
    Up   func(db *gorm.DB) error   // Fungsi yang dijalankan
}
```

### Cara Menambah Migration Manual Baru

Buka `models/migrations.go` dan tambahkan entry baru di bagian **"TAMBAHKAN MIGRATION BARU DI BAWAH SINI"**:

```go
var ManualMigrations = []MigrationStep{
    // Entry lama JANGAN dihapus atau diubah!

    // ✅ Tambahkan di sini:
    {
        Name: "2026-06-10_nama_deskripsi_singkat",
        Up: func(db *gorm.DB) error {
            // Tulis operasi migration di sini
            return nil
        },
    },
}
```

> ⚠️ **PENTING**: Jangan pernah mengubah atau menghapus entry yang sudah ada. Nama (`Name`) harus unik dan tidak boleh berubah karena digunakan sebagai identifier di database.

### Referensi Operasi `db.Migrator()`

| Operasi | Kode |
|---------|------|
| Hapus kolom | `db.Migrator().DropColumn(&Model{}, "nama_kolom")` |
| Rename kolom | `db.Migrator().RenameColumn(&Model{}, "lama", "baru")` |
| Ubah tipe/constraint kolom | `db.Migrator().AlterColumn(&Model{}, "nama_kolom")` |
| Cek kolom ada | `db.Migrator().HasColumn(&Model{}, "nama_kolom")` |
| Hapus tabel | `db.Migrator().DropTable(&Model{})` |
| Rename tabel | `db.Migrator().RenameTable("lama", "baru")` |
| Buat index | `db.Migrator().CreateIndex(&Model{}, "IndexName")` |
| Hapus index | `db.Migrator().DropIndex(&Model{}, "IndexName")` |
| Raw SQL | `db.Exec("ALTER TABLE ...")` |

### Contoh-Contoh Nyata

#### Hapus Kolom yang Tidak Dipakai Lagi

```go
{
    Name: "2026-06-10_drop_users_old_field",
    Up: func(db *gorm.DB) error {
        if db.Migrator().HasColumn(&User{}, "old_field") {
            return db.Migrator().DropColumn(&User{}, "old_field")
        }
        return nil // aman jika kolom sudah tidak ada
    },
},
```

#### Rename Kolom

```go
{
    Name: "2026-06-10_rename_users_phone_to_phone_number",
    Up: func(db *gorm.DB) error {
        if db.Migrator().HasColumn(&User{}, "phone") {
            return db.Migrator().RenameColumn(&User{}, "phone", "phone_number")
        }
        return nil
    },
},
```

#### Ubah Tipe Kolom (via Raw SQL)

```go
{
    Name: "2026-06-10_alter_loggers_detail_to_text",
    Up: func(db *gorm.DB) error {
        return db.Exec("ALTER TABLE loggers ALTER COLUMN detail TYPE TEXT").Error
    },
},
```

#### Hapus Tabel yang Sudah Tidak Digunakan

```go
{
    Name: "2026-06-10_drop_old_tokens_table",
    Up: func(db *gorm.DB) error {
        return db.Migrator().DropTable("old_tokens")
    },
},
```

---

## 📄 `models/seeder.go` — Seeder Data Awal

Berisi fungsi `Seed` yang mengisi data awal ke database. Aman dijalankan berkali-kali karena menggunakan `FirstOrCreate`.

### Fungsi `Seed(db *gorm.DB)`

**Apa yang di-seed:**

#### 1. Role Default — Super Admin

```go
db.FirstOrCreate(&role, Role{
    Name:     "Super Admin",
    RoleType: RoleTypeSuperAdmin,
})
```

#### 2. Permission Default

```go
permissions := []Permission{
    {Name: "manage_users", Label: "Manage Users"},
    {Name: ""manage_roles", Label: "Manage Roles"},
}
```

Setiap permission dicek dengan `WHERE name = ?`, jika belum ada baru dibuat.

#### 3. Assign Permission ke Role Super Admin

Semua permission diberikan ke role Super Admin dengan akses penuh:

```go
db.FirstOrCreate(&rp, RolePermission{
    RoleID:       role.ID,
    PermissionID: p.ID,
    CanRead:      true,
    CanWrite:     true,
    CanUpdate:    true,
    CanDelete:    true,
    CanRestore:   true,
})
```

#### 4. User Super Admin Default

```go
// Email: superadmin@example.com
// Password: password123 (di-hash dengan bcrypt)
db.FirstOrCreate(&user, User{
    Email:    "superadmin@example.com",
    Name:     &name,
    Password: &pass,
    RoleID:   role.ID,
})
```

> 🔒 **Penting**: Ganti password default Super Admin setelah pertama kali login di production!

**Cara menjalankan:**

```bash
make seed
# atau
go run cmd/seed/main.go
```

---

## 🔄 Alur Lengkap Setup Pertama Kali

```bash
# 1. Copy dan edit .env
cp .env.example .env

# 2. Jalankan migration
#    (database akan otomatis dibuat jika belum ada)
make migrate

# 3. Jalankan seeder
make seed

# 4. Jalankan server
make dev
```

Setelah itu, login menggunakan:
- **Email**: `superadmin@example.com`
- **Password**: `password123`

---

## ➕ Cara Menambah Tabel Baru

Berikut adalah panduan lengkap langkah demi langkah untuk menambahkan tabel baru ke dalam proyek.

### Langkah 1 — Buat Struct Model di `models/models.go`

Tambahkan struct baru di file `models/models.go`. Gunakan tag GORM dan JSON untuk mendefinisikan kolom dan perilaku tabel.

**Template struct dasar:**

```go
type NamaModel struct {
    ID        uint           `gorm:"primaryKey" json:"id"`
    // ... field lainnya ...
    CreatedAt time.Time      `json:"createdAt"`
    UpdatedAt time.Time      `json:"updatedAt"`
    DeletedAt gorm.DeletedAt `gorm:"index" json:"deletedAt,omitempty"` // Hapus jika tidak butuh soft delete
}
```

**Contoh nyata — menambah tabel `Product`:**

```go
// Di models/models.go

type ProductStatus string

const (
    ProductStatusActive   ProductStatus = "ACTIVE"
    ProductStatusInactive ProductStatus = "INACTIVE"
)

type Product struct {
    ID          uint          `gorm:"primaryKey" json:"id"`
    Name        string        `gorm:"not null" json:"name"`
    Description *string       `gorm:"type:text" json:"description"`
    Price       float64       `gorm:"not null;default:0" json:"price"`
    Stock       int           `gorm:"default:0" json:"stock"`
    Status      ProductStatus `gorm:"type:varchar(20);default:'ACTIVE'" json:"status"`
    CreatedByID uint          `json:"createdById"`
    CreatedBy   User          `gorm:"foreignKey:CreatedByID" json:"createdBy,omitempty"`
    CreatedAt   time.Time     `json:"createdAt"`
    UpdatedAt   time.Time     `json:"updatedAt"`
    DeletedAt   gorm.DeletedAt `gorm:"index" json:"deletedAt,omitempty"`
}
```

**Referensi tag GORM yang umum digunakan:**

| Tag GORM | Keterangan |
|----------|------------|
| `gorm:"primaryKey"` | Menandai kolom sebagai primary key |
| `gorm:"uniqueIndex"` | Membuat unique index |
| `gorm:"index"` | Membuat index biasa |
| `gorm:"not null"` | Kolom tidak boleh null |
| `gorm:"default:nilai"` | Nilai default kolom |
| `gorm:"type:varchar(255)"` | Tipe kolom kustom |
| `gorm:"type:text"` | Tipe TEXT (untuk string panjang) |
| `gorm:"foreignKey:NamaField"` | Mendefinisikan foreign key |
| `gorm:"constraint:OnDelete:CASCADE"` | Cascade delete |
| `gorm:"autoCreateTime"` | Auto set waktu saat create |
| `gorm:"autoUpdateTime"` | Auto set waktu saat update |

---

### Langkah 2 — Daftarkan Model ke `AutoMigrate`

Buka `models/autoMigrate.go` dan tambahkan pointer struct model baru ke dalam list `db.AutoMigrate(...)`.

```go
// models/autoMigrate.go

func AutoMigrate(db *gorm.DB) {
    err := db.AutoMigrate(
        &Permission{},
        &Role{},
        &RolePermission{},
        &User{},
        &Otp{},
        &Session{},
        &Logger{},
        &Notification{},
        &NotificationUser{},
        &WebPushSubscription{},
        &Product{},   // ← Tambahkan di sini
    )
    if err != nil {
        log.Fatal("Failed to auto migrate database:", err)
    }
    log.Println("Database auto-migration completed successfully.")
}
```

> ⚠️ **Urutan penting**: Jika model baru memiliki foreign key ke model lain, pastikan model yang direferensikan didaftarkan **lebih dahulu** dalam list. Contoh: `Product` mereferensikan `User`, maka `User` harus ada sebelum `Product`.

---

### Langkah 3 — Jalankan Migration

```bash
make migrate
```

Output yang diharapkan:
```
2026/06/07 15:00:00 Initializing migration...
2026/06/07 15:00:00 Database connection established
2026/06/07 15:00:00 Database auto-migration completed successfully.
2026/06/07 15:00:00 Migration command finished.
```

GORM akan secara otomatis membuat tabel `products` (nama tabel di-pluralize dan di-snake_case oleh GORM).

> 📝 **Konvensi nama tabel GORM**: Struct `Product` → tabel `products`. Struct `ProductCategory` → tabel `product_categories`.

---

### Langkah 4 — (Opsional) Tambah Relasi ke Model Lain

Jika model lain perlu memiliki relasi ke model baru, tambahkan field di struct model tersebut.

**Contoh — menambahkan relasi `User` memiliki banyak `Product`:**

```go
// Di struct User (models/models.go)
type User struct {
    // ... field yang sudah ada ...
    Products []Product `gorm:"foreignKey:CreatedByID" json:"products,omitempty"` // ← Tambahkan ini
}
```

Kemudian jalankan `make migrate` lagi — GORM tidak akan mengubah tabel yang sudah ada, hanya menambah kolom baru jika diperlukan.

---

### Langkah 5 — (Opsional) Tambahkan Data Awal ke Seeder

Jika tabel baru perlu data awal, buka `models/seeder.go` dan tambahkan logika seed.

```go
// Di models/seeder.go, dalam fungsi Seed()

// Seed produk default
var defaultProduct Product
db.FirstOrCreate(&defaultProduct, Product{
    Name:        "Sample Product",
    Price:       100000,
    Stock:       10,
    Status:      ProductStatusActive,
    CreatedByID: user.ID,
})
```

Jalankan seeder:
```bash
make seed
```

---

### Ringkasan Checklist

```
✅ Langkah 1 — Definisikan struct model baru di models/models.go
✅ Langkah 2 — Daftarkan &NamaModel{} di models/autoMigrate.go
✅ Langkah 3 — Jalankan: make migrate
✅ Langkah 4 — (Opsional) Tambah relasi di model lain yang terkait
✅ Langkah 5 — (Opsional) Tambah data awal di models/seeder.go → make seed
```

---

### Hal-Hal yang Perlu Diperhatikan

> ⚠️ **GORM AutoMigrate tidak menghapus kolom yang sudah ada**. Jika perlu menghapus/rename kolom, lakukan secara manual dengan SQL:
> ```sql
> ALTER TABLE products DROP COLUMN old_column;
> ALTER TABLE products RENAME COLUMN old_name TO new_name;
> ```

> ⚠️ **Hati-hati mengubah tipe kolom**. GORM AutoMigrate umumnya tidak mengubah tipe kolom yang sudah ada. Lakukan perubahan tipe kolom secara manual jika diperlukan.

> 💡 **Naming convention**: Gunakan nama struct dalam PascalCase (`ProductCategory`), GORM otomatis membuat nama tabel dalam snake_case plural (`product_categories`). Jika ingin nama tabel kustom, implementasikan method `TableName()`:
> ```go
> func (ProductCategory) TableName() string {
>     return "my_custom_table_name"
> }
> ```

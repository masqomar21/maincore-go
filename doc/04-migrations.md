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

> ⚠️ **Perhatian**: GORM AutoMigrate tidak menghapus kolom. Jika ada kolom yang perlu dihapus, harus dilakukan secara manual di database.

### Menambah Model Baru ke AutoMigrate

1. Buat struct model baru di `models/models.go`
2. Tambahkan pointer struct tersebut ke dalam list `db.AutoMigrate(...)` di `models/autoMigrate.go`
3. Jalankan `make migrate`

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

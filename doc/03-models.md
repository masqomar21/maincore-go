# 03 — Models (Skema Database)

Paket `models` berisi semua definisi struct Go yang digunakan sebagai skema tabel database melalui GORM.

---

## 📄 `models/models.go`

### Enum `RoleType`

Tipe role yang tersedia dalam sistem.

```go
type RoleType string

const (
    RoleTypeOther      RoleType = "OTHER"       // Role umum / user biasa
    RoleTypeSuperAdmin RoleType = "SUPER_ADMIN"  // Super admin dengan akses penuh
)
```

---

### Model `Permission`

Menyimpan daftar permission/izin yang ada di sistem.

```go
type Permission struct {
    ID              uint             // Primary key
    Name            string           // Nama permission (e.g. "manage_users")
    Label           string           // Label human-readable (e.g. "Manage Users")
    RolePermissions []RolePermission // Relasi ke RolePermission
}
```

**Relasi:** One-to-Many → `RolePermission`

---

### Model `Role`

Menyimpan daftar role yang dapat dimiliki oleh user.

```go
type Role struct {
    ID              uint             // Primary key
    Name            string           // Nama role
    RoleType        RoleType         // Tipe role (OTHER / SUPER_ADMIN), default: "OTHER"
    RolePermissions []RolePermission // Relasi ke RolePermission
    Users           []User           // Relasi ke User
}
```

**Relasi:** One-to-Many → `RolePermission`, One-to-Many → `User`

---

### Model `RolePermission`

Tabel pivot yang menghubungkan Role dengan Permission beserta level akses CRUD-nya.

```go
type RolePermission struct {
    ID           uint       // Primary key
    RoleID       uint       // Foreign key → Role
    Role         Role
    PermissionID uint       // Foreign key → Permission
    Permission   Permission
    CanRead      bool       // Izin baca (default: false)
    CanWrite     bool       // Izin tulis/buat (default: false)
    CanUpdate    bool       // Izin ubah (default: false)
    CanDelete    bool       // Izin hapus (default: false)
    CanRestore   bool       // Izin restore soft-delete (default: false)
}
```

---

### Model `User`

Model utama untuk pengguna aplikasi, mendukung soft delete.

```go
type User struct {
    ID                   uint
    Email                string                // Unique, not null
    Name                 *string
    Password             *string               // Di-hide dari JSON response
    Address              *string               // Contoh field tambahan
    PhoneNumber          *string               // Contoh field tambahan
    RoleID               uint                  // Foreign key → Role
    Role                 Role
    RegisteredViaGoogle  bool                  // Default: false
    CreatedAt            time.Time
    UpdatedAt            time.Time
    DeletedAt            gorm.DeletedAt        // Soft delete
    Sessions             []Session
    Loggers              []Logger
    Notifications        []NotificationUser
    WebPushSubscriptions []WebPushSubscription
    OTP                  *Otp
}
```

> 💡 **Tip Menambah Field Baru**: Tambahkan field baru di antara komentar yang sudah disediakan, kemudian jalankan `make migrate` untuk sinkronisasi skema.

---

### Enum `OtpPurpose`

Tujuan penggunaan OTP.

```go
type OtpPurpose string

const (
    OtpPurposeLogin         OtpPurpose = "LOGIN"
    OtpPurposeResetPassword OtpPurpose = "RESET_PASSWORD"
    OtpPurposeVerifyEmail   OtpPurpose = "VERIFY_EMAIL"
)
```

---

### Model `Otp`

Menyimpan OTP (One-Time Password) yang dikirim ke user.

```go
type Otp struct {
    ID        uint
    UserID    uint       // Unique, Foreign key → User
    User      User
    Code      string     // Unique, not null
    Purpose   OtpPurpose // Tujuan OTP
    CreatedAt time.Time
    ExpiresAt time.Time  // Waktu kadaluarsa OTP
}
```

---

### Model `Session`

Menyimpan sesi aktif user (token-based session), mendukung soft delete.

```go
type Session struct {
    ID        uint
    Token     string         // Unique, not null — JWT Access Token
    UserID    uint           // Foreign key → User
    User      User
    CreatedAt time.Time
    UpdatedAt time.Time
    DeletedAt gorm.DeletedAt // Soft delete
}
```

---

### Enum `Process`

Jenis aksi yang dicatat dalam activity log.

```go
type Process string

const (
    ProcessCreate  Process = "CREATE"
    ProcessUpdate  Process = "UPDATE"
    ProcessDelete  Process = "DELETE"
    ProcessRestore Process = "RESTORE"
    ProcessLogin   Process = "LOGIN"
    ProcessLogout  Process = "LOGOUT"
)
```

---

### Model `Logger`

Menyimpan activity log dari setiap aksi user.

```go
type Logger struct {
    ID        uint
    UserID    uint      // Foreign key → User
    User      User
    Process   Process   // Jenis aksi
    Detail    string    // Keterangan detail
    CreatedAt time.Time
}
```

---

### Model `Notification`

Template notifikasi yang dikirim ke satu atau lebih user.

```go
type Notification struct {
    ID         uint
    Type       string             // Tipe notifikasi
    RefID      *string            // Reference ID opsional (nullable)
    Message    string             // Isi pesan notifikasi
    CreatedAt  time.Time
    Recipients []NotificationUser // Daftar penerima
}
```

---

### Model `NotificationUser`

Tabel pivot yang menghubungkan Notification dengan User, menyimpan status baca.

```go
type NotificationUser struct {
    ID             uint
    UserID         uint         // Composite unique index dengan NotificationID
    NotificationID uint         // Composite unique index dengan UserID
    ReadStatus     bool         // Default: false
    ReadAt         *time.Time   // Waktu dibaca (nullable)
    DeliveredAt    *time.Time   // Waktu dikirim (nullable)
    User           User         // ON DELETE CASCADE
    Notification   Notification // ON DELETE CASCADE
}
```

**Index:**
- `idx_user_notif`: Composite unique index `(UserID, NotificationID)` — mencegah duplikasi
- `idx_user_read`: Index pada `ReadStatus` — mempercepat query notifikasi belum dibaca

---

### Model `WebPushSubscription`

Menyimpan data subscription browser Push Notification (Web Push API).

```go
type WebPushSubscription struct {
    ID             uint
    UserID         uint       // Index, Foreign key → User
    Endpoint       string     // Unique, not null — URL push endpoint browser
    P256dh         string     // Encryption key
    Auth           string     // Auth token
    ExpirationTime *time.Time // Nullable
    UserAgent      *string    // Info browser (nullable)
    CreatedAt      time.Time
    UpdatedAt      time.Time
    User           User       // ON DELETE CASCADE
}
```

---

## Diagram Relasi

```
Permission ◄──── RolePermission ────► Role
                                        │
                                        ▼
                                       User
                                      / | \ \
                           Session  Otp  Logger  NotificationUser  WebPushSubscription
                                              │
                                        Notification
```

---

## ➕ Menambah Model / Tabel Baru

Untuk panduan lengkap cara menambahkan tabel baru ke dalam proyek (termasuk struct template, tag GORM, registrasi AutoMigrate, dan seeder), lihat:

👉 **[04-migrations.md — Cara Menambah Tabel Baru](./04-migrations.md#-cara-menambah-tabel-baru)**

**Ringkasan singkat:**

```
1. Buat struct baru di models/models.go
2. Daftarkan &NamaModel{} di models/autoMigrate.go
3. Jalankan: make migrate
```

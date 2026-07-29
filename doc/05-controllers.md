# 05 — Controllers

Controller adalah lapisan yang menangani request HTTP, memproses logika bisnis, dan mengembalikan response.

---

## 📄 `controllers/auth.go` — Autentikasi

### Struct Input

```go
type RegisterInput struct {
    Name     string `json:"name"     binding:"required"`
    Email    string `json:"email"    binding:"required,email"`
    Password string `json:"password" binding:"required,min=6"`
}

type LoginInput struct {
    Email    string `json:"email"    binding:"required,email"`
    Password string `json:"password" binding:"required"`
}
```

---

### `Register(c *gin.Context)`

Mendaftarkan user baru ke sistem.

**Alur:**
1. Validasi input request body
2. Cari role dengan `RoleType = "OTHER"` (role default untuk user baru)
3. Hash password menggunakan bcrypt
4. Simpan user ke database
5. Return user yang dibuat (HTTP 201)

**Error yang mungkin:**
- `400` — Input tidak valid atau role tidak ditemukan
- `500` — Gagal hash password atau gagal simpan ke database

---

### `Login(c *gin.Context)`

Melakukan autentikasi user dan membuat session.

**Alur:**
1. Validasi input
2. Cari user berdasarkan email (dengan preload Role)
3. Verifikasi password dengan bcrypt
4. Generate JWT Access Token (berlaku 24 jam)
5. Simpan session token ke tabel `sessions`
6. Catat activity log (Process: LOGIN)
7. Return token dan data user

**Response data:**
```json
{
  "id": 1,
  "email": "user@example.com",
  "name": "John Doe",
  "registeredViaGoogle": false,
  "token": "eyJhbGciOiJIUzI1NiIs..."
}
```

**Error yang mungkin:**
- `400` — Input tidak valid
- `401` — Password tidak sesuai
- `404` — User tidak ditemukan

---

### `GetUserProfile(c *gin.Context)`

Mengambil profil user yang sedang login beserta daftar permission-nya.

**Alur:**
1. Ambil data user login dari context (sudah di-set oleh `AuthMiddleware`)
2. Query user dengan preload: `Role.RolePermissions.Permission`
3. Map permission menjadi format string: `"read:manage_users"`, `"write:manage_users"`, dst.

**Response data:**
```json
{
  "id": 1,
  "name": "John",
  "email": "john@example.com",
  "registeredViaGoogle": false,
  "role": {
    "name": "Super Admin",
    "roleType": "SUPER_ADMIN",
    "rolePermissions": [
      "read:manage_users",
      "write:manage_users",
      "update:manage_users",
      "delete:manage_users"
    ]
  }
}
```

> 🔒 **Auth required**: Memerlukan `AuthMiddleware` dan `GeneratePermissionList`

---

### `Logout(c *gin.Context)`

Menghapus session aktif user (invalidasi token).

**Alur:**
1. Ambil token dari header `Authorization`
2. Hapus session dari database berdasarkan `user_id` dan `token`
3. Catat activity log (Process: LOGOUT)

> 🔒 **Auth required**

---

## 📄 `controllers/reset_password.go` — Reset Password

### Struct Input

```go
type VerifyEmailInput struct {
    Email string `json:"email" binding:"required,email"`
}

type VerifyOtpInput struct {
    Email string `json:"email" binding:"required,email"`
    Code  string `json:"code"  binding:"required"`
}

type ResetPasswordInput struct {
    Token    string `json:"token"    binding:"required"`
    Password string `json:"password" binding:"required,min=6"`
}
```

---

### `SearchEmail(c *gin.Context)`

Langkah 1: Mencari email dan membuat OTP.

**Alur:**
1. Validasi email ada di database
2. Hapus OTP lama (jika ada) untuk user tersebut dengan purpose `RESET_PASSWORD`
3. Generate OTP 4 digit acak
4. Simpan OTP dengan masa berlaku 5 menit
5. (Di production) Kirim OTP via email/SMTP

> 📧 Implementasi pengiriman email perlu ditambahkan sendiri (lihat komentar di kode).

---

### `VerifyOtp(c *gin.Context)`

Langkah 2: Memverifikasi OTP yang dikirim ke email.

**Alur:**
1. Cari user berdasarkan email
2. Cari OTP berdasarkan `user_id`, `code`, dan `purpose = RESET_PASSWORD`
3. Cek apakah OTP sudah kadaluarsa
4. Jika valid, generate token sementara (JWT, berlaku 15 menit) dengan purpose `RESET_PASSWORD`
5. Hapus OTP dari database
6. Return token reset password

---

### `ResetPassword(c *gin.Context)`

Langkah 3: Mengubah password menggunakan token reset.

**Alur:**
1. Verifikasi token JWT dengan purpose `RESET_PASSWORD`
2. Ambil user berdasarkan ID dari token
3. Hash password baru dengan bcrypt
4. Update password di database
5. Catat activity log (Process: UPDATE)

---

## 📄 `controllers/master.go` — CRUD Master Data

### Users

| Fungsi | Deskripsi |
|--------|-----------|
| `ListUsers` | Daftar semua user (dengan preload Role) |
| `CreateUser` | Membuat user baru (delegate ke `Register`) |
| `GetUser` | Detail user berdasarkan ID |
| `DeleteUser` | Hapus user berdasarkan ID (soft delete) |

### Roles

| Fungsi | Deskripsi |
|--------|-----------|
| `ListRoles` | Daftar semua role (dengan preload RolePermissions.Permission) |

### Notifications

| Fungsi | Deskripsi |
|--------|-----------|
| `ListNotifications` | Daftar notifikasi untuk user yang login |
| `ReadNotification` | Tandai notifikasi sebagai sudah dibaca |

### Logs

| Fungsi | Deskripsi |
|--------|-----------|
| `ListLogs` | Daftar 100 activity log terbaru (desc) |

---

## 📄 `controllers/webpush.go` — Web Push Notification

### Struct Input

```go
type SubscribeInput struct {
    Endpoint string `json:"endpoint" binding:"required"`
    Keys struct {
        P256dh string `json:"p256dh" binding:"required"`
        Auth   string `json:"auth"   binding:"required"`
    } `json:"keys" binding:"required"`
}

type UnsubscribeInput struct {
    Endpoint string `json:"endpoint" binding:"required"`
}
```

### `SubscribeToWebPush(c *gin.Context)`

Mendaftarkan browser subscription untuk push notification.

**Alur:**
1. Validasi input
2. Cek apakah endpoint sudah ada
   - Jika belum: buat subscription baru
   - Jika sudah: update data subscription (UserID, P256dh, Auth)

### `UnsubscribeFromWebPush(c *gin.Context)`

Menghapus subscription push notification berdasarkan endpoint.

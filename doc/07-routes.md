# 07 — Routes (API Endpoints)

Paket `routes` mendaftarkan semua endpoint HTTP API ke Gin Router. Semua route di-prefix dengan `/api`.

---

## 📋 Ringkasan Semua Endpoint

### 🔓 Public Routes (Tanpa Autentikasi)

| Method | Endpoint | Controller | Keterangan |
|--------|----------|------------|------------|
| `POST` | `/api/auth/register` | `Register` | Registrasi user baru |
| `POST` | `/api/auth/login` | `Login` | Login dan dapatkan token |
| `POST` | `/api/reset-password/verify-email` | `SearchEmail` | Kirim OTP ke email |
| `POST` | `/api/reset-password/verify-otp` | `VerifyOtp` | Verifikasi OTP |
| `PUT`  | `/api/reset-password/change-password` | `ResetPassword` | Reset password dengan token |

### 🔐 Protected Routes (Memerlukan JWT Token)

| Method | Endpoint | Permission | Controller | Keterangan |
|--------|----------|-----------|------------|------------|
| `GET` | `/api/profile/` | — | `GetUserProfile` | Profil user login |
| `POST` | `/api/profile/logout` | — | `Logout` | Logout |
| `GET` | `/api/master/user/` | `manage_users:canRead` | `ListUsers` | Daftar user |
| `POST` | `/api/master/user/` | `manage_users:canWrite` | `CreateUser` | Buat user |
| `GET` | `/api/master/user/:id` | `manage_users:canRead` | `GetUser` | Detail user |
| `DELETE` | `/api/master/user/:id` | `manage_users:canDelete` | `DeleteUser` | Hapus user |
| `GET` | `/api/master/role/` | `manage_roles:canRead` | `ListRoles` | Daftar role |
| `GET` | `/api/notification/` | — | `ListNotifications` | Notifikasi user |
| `PUT` | `/api/notification/read/:id` | — | `ReadNotification` | Tandai terbaca |
| `GET` | `/api/log/` | `manage_logs:canRead` | `ListLogs` | Activity logs |
| `POST` | `/api/web-push/subscribe` | — | `SubscribeToWebPush` | Subscribe push notif |
| `POST` | `/api/web-push/unsubscribe` | — | `UnsubscribeFromWebPush` | Unsubscribe push notif |

---

## 📄 `routes/auth.go`

```go
func AuthRoutes(router *gin.RouterGroup)
```

### Public: `/api/auth`

```
POST /api/auth/register   → controllers.Register
POST /api/auth/login      → controllers.Login
```

### Protected: `/api/profile`

Middleware: `AuthMiddleware` + `GeneratePermissionList`

```
GET  /api/profile/        → controllers.GetUserProfile
POST /api/profile/logout  → controllers.Logout
```

---

## 📄 `routes/reset_password.go`

```go
func ResetPasswordRoutes(router *gin.RouterGroup)
```

### Public: `/api/reset-password`

Tidak memerlukan autentikasi.

```
POST /api/reset-password/verify-email    → controllers.SearchEmail
POST /api/reset-password/verify-otp     → controllers.VerifyOtp
PUT  /api/reset-password/change-password → controllers.ResetPassword
```

**Flow reset password:**
```
[POST /verify-email]
  → Kirim email berisi OTP 4 digit (berlaku 5 menit)
      │
      ▼
[POST /verify-otp]
  → Verifikasi OTP, dapatkan reset token (berlaku 15 menit)
      │
      ▼
[PUT /change-password]
  → Gunakan reset token untuk ubah password
```

---

## 📄 `routes/master.go`

```go
func MasterRoutes(router *gin.RouterGroup)
```

Semua route di sini dilindungi oleh `AuthMiddleware` + `GeneratePermissionList`.

### Users: `/api/master/user`

```
GET    /api/master/user/      [manage_users:canRead]   → ListUsers
POST   /api/master/user/      [manage_users:canWrite]  → CreateUser
GET    /api/master/user/:id   [manage_users:canRead]   → GetUser
DELETE /api/master/user/:id   [manage_users:canDelete] → DeleteUser
```

### Roles: `/api/master/role`

```
GET /api/master/role/   [manage_roles:canRead] → ListRoles
```

### Notifications: `/api/notification`

```
GET /api/notification/          → ListNotifications   (semua notification user login)
PUT /api/notification/read/:id  → ReadNotification    (tandai 1 notifikasi terbaca)
```

### Logs: `/api/log`

```
GET /api/log/   [manage_logs:canRead] → ListLogs   (100 log terbaru, desc)
```

### Web Push: `/api/web-push`

```
POST /api/web-push/subscribe    → SubscribeToWebPush
POST /api/web-push/unsubscribe  → UnsubscribeFromWebPush
```

---

## Contoh Request & Response

### Login

**Request:**
```http
POST /api/auth/login
Content-Type: application/json

{
  "email": "superadmin@example.com",
  "password": "password123"
}
```

**Response (200):**
```json
{
  "status": 200,
  "message": "Success",
  "data": {
    "id": 1,
    "email": "superadmin@example.com",
    "name": "Super Admin",
    "registeredViaGoogle": false,
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }
}
```

### Menggunakan Token di Header

```http
GET /api/profile/
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

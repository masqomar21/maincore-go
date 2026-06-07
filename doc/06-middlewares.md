# 06 — Middlewares

Middleware adalah fungsi yang berjalan sebelum request sampai ke controller. Middleware digunakan untuk autentikasi, otorisasi, validasi, dan modifikasi request/response.

---

## 📄 `middlewares/auth.go` — Autentikasi JWT

### `AuthMiddleware() gin.HandlerFunc`

Middleware ini memastikan setiap request yang dilindungi memiliki token JWT yang valid dan sesi yang aktif di database.

**Alur validasi:**
1. Cek header `Authorization` tidak kosong
2. Pastikan format header adalah `Bearer <token>`
3. Cek token ada di tabel `sessions` (sesi aktif)
4. Verifikasi JWT token dengan secret key
5. Pastikan purpose token adalah `"ACCESS_TOKEN"` (bukan RESET_PASSWORD, dll.)
6. Simpan data decoded payload ke context (`c.Set("user", decode)`)

**Error responses:**
| Kondisi | Status Code | Pesan |
|---------|-------------|-------|
| Tidak ada header Authorization | `401` | Unauthorized - No token provided |
| Format tidak valid | `401` | Unauthorized - Invalid token format |
| Sesi tidak ditemukan di DB | `498` | Unauthorized - Invalid session |
| Token tidak valid | `498` | Unauthorized - Invalid token |
| Purpose token salah | `498` | Unauthorized - Invalid token purpose |

> 📝 **Catatan**: Kode `498` adalah custom HTTP status yang menandakan token tidak valid/kadaluarsa.

**Penggunaan di routes:**
```go
router.Use(middlewares.AuthMiddleware())
```

---

## 📄 `middlewares/cors.go` — CORS

### `CorsMiddleware() gin.HandlerFunc`

Mengizinkan cross-origin request dari semua domain.

**Header yang diset:**
```
Access-Control-Allow-Origin: *
Access-Control-Allow-Credentials: true
Access-Control-Allow-Headers: Content-Type, Content-Length, Accept-Encoding,
    X-CSRF-Token, Authorization, accept, origin, Cache-Control,
    X-Requested-With, socket.io
Access-Control-Allow-Methods: POST, OPTIONS, GET, PUT, DELETE
```

**Preflight Request (OPTIONS):**
Request OPTIONS akan langsung dijawab dengan status `204 No Content` tanpa meneruskan ke handler berikutnya.

**Penggunaan di routes:**
```go
r.Use(middlewares.CorsMiddleware())
```

> 🔒 **Untuk Production**: Ganti `Allow-Origin: *` dengan domain spesifik frontend Anda untuk keamanan lebih baik.

---

## 📄 `middlewares/permission.go` — Manajemen Permission

Middleware ini mengelola sistem permission berbasis role dengan caching Redis.

---

### Struct `GeneratedPermissionList`

```go
type GeneratedPermissionList struct {
    Permission string  // Nama permission (e.g. "manage_users")
    CanRead    bool
    CanWrite   bool
    CanUpdate  bool
    CanDelete  bool
    CanRestore bool
}
```

---

### `GeneratePermissionList() gin.HandlerFunc`

Middleware yang memuat dan menyimpan daftar permission user ke dalam request context.

**Alur dengan Redis Cache:**
```
Request masuk
    │
    ▼
Cek Redis: key "user_permissions:{userID}"
    │
    ├─► Cache HIT  → Deserialize JSON → Set ke context → Next()
    │
    └─► Cache MISS
            │
            ▼
        Query DB: User + Role + RolePermissions + Permission
            │
            ▼
        Build permission list
            │
            ▼
        Serialize → Simpan ke Redis (TTL: 1 jam)
            │
            ▼
        Set ke context → Next()
```

**Cache key format:** `user_permissions:{userID}`
**Cache TTL:** 1 jam

> ⚡ **Performa**: Dengan Redis cache, permission user tidak perlu query database setiap request. Cache otomatis digunakan hingga 1 jam.

---

### `RequirePermission(permissionName string, action string) gin.HandlerFunc`

Middleware untuk memeriksa apakah user memiliki izin tertentu untuk aksi tertentu.

**Parameter:**
| Parameter | Nilai yang valid | Keterangan |
|-----------|-----------------|------------|
| `permissionName` | e.g. `"manage_users"`, `"manage_roles"` | Nama permission yang dicek |
| `action` | `"all"`, `"canRead"`, `"canWrite"`, `"canUpdate"`, `"canDelete"`, `"canRestore"` | Jenis aksi yang diperlukan |

**Alur:**
1. Cek user sudah login (ada di context)
2. Jika `RoleType == "SUPER_ADMIN"` → langsung izinkan akses (bypass semua permission)
3. Ambil permission list dari context (diisi oleh `GeneratePermissionList`)
4. Cari permission yang cocok berdasarkan `permissionName` dan `action`
5. Jika tidak ada permission → return `403 Forbidden`

**Penggunaan di routes:**
```go
// Hanya izinkan user dengan izin "canRead" pada "manage_users"
users.GET("/", middlewares.RequirePermission("manage_users", "canRead"), controllers.ListUsers)

// Hanya izinkan user dengan izin "canWrite" pada "manage_users"
users.POST("/", middlewares.RequirePermission("manage_users", "canWrite"), controllers.CreateUser)
```

**Error responses:**
| Kondisi | Status | Pesan |
|---------|--------|-------|
| User tidak di context | `401` | Unauthorized |
| Permission list tidak ada | `403` | No permissions loaded |
| Permission tidak terpenuhi | `403` | Forbidden - You do not have permission to {action} {permission} |

---

## 📄 `middlewares/fileupload.go` — Upload File

### `FileUploadMiddleware(maxFileSize int64, allowedFileTypes []string) gin.HandlerFunc`

Middleware untuk memvalidasi dan mengkonfigurasi upload file.

**Alur:**
1. Cek `Content-Length` header — jika melebihi `maxFileSize` langsung reject
2. Parse multipart form (`ParseMultipartForm`)
3. Simpan konfigurasi ke context:
   - `c.Set("maxFileSize", maxFileSize)`
   - `c.Set("allowedFileTypes", allowedFileTypes)`

**Penggunaan:**
```go
router.POST("/upload", 
    middlewares.FileUploadMiddleware(5*1024*1024, []string{"image/jpeg", "image/png"}),
    controllers.UploadFile,
)
```

---

### `CheckFileType(mime string, allowedFileTypes []string) bool`

Helper function yang digunakan di dalam Controller untuk memverifikasi tipe MIME file.

```go
// Contoh penggunaan di controller:
if !middlewares.CheckFileType(header.Header.Get("Content-Type"), allowedTypes) {
    utilities.BadRequest(c, "File type not allowed", nil)
    return
}
```

---

## Urutan Middleware di Routes

```
Request
  │
  ▼
[CorsMiddleware]          ← Global, semua request
  │
  ▼
[AuthMiddleware]          ← Protected routes
  │
  ▼
[GeneratePermissionList]  ← Setelah auth, memuat permission
  │
  ▼
[RequirePermission(...)]  ← Per-route permission check
  │
  ▼
[Controller]
```

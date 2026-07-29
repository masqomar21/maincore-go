# 09 — Utilities

Paket `utilities` berisi helper functions yang digunakan di seluruh aplikasi.

---

## 📄 `utilities/jwt.go` — JSON Web Token

Mengelola pembuatan dan verifikasi JWT menggunakan library `golang-jwt/jwt/v5`.

### Struct `JwtPayload`

Data yang disimpan di dalam JWT token.

```go
type JwtPayload struct {
    ID       uint   `json:"id"`
    Name     string `json:"name"`
    Role     string `json:"role,omitempty"`     // Nama role
    RoleType string `json:"roleType"`            // "OTHER" atau "SUPER_ADMIN"
    Purpose  string `json:"purpose"`             // "ACCESS_TOKEN" | "RESET_PASSWORD"
    jwt.RegisteredClaims                         // ExpiresAt, IssuedAt, dll.
}
```

**Field `Purpose`** digunakan untuk membedakan fungsi token:
- `"ACCESS_TOKEN"` — Token untuk autentikasi normal
- `"RESET_PASSWORD"` — Token sementara untuk reset password

---

### `GenerateAccessToken(payload JwtPayload, expiresIn time.Duration) (string, error)`

Membuat JWT token baru menggunakan algoritma HMAC-SHA256.

**Parameter:**
| Parameter | Keterangan |
|-----------|------------|
| `payload` | Data yang akan disimpan di dalam token |
| `expiresIn` | Masa berlaku token |

**Contoh penggunaan:**
```go
// Token akses (berlaku 24 jam)
token, err := utilities.GenerateAccessToken(utilities.JwtPayload{
    ID:       user.ID,
    Name:     *user.Name,
    Role:     user.Role.Name,
    RoleType: string(user.Role.RoleType),
    Purpose:  "ACCESS_TOKEN",
}, 24*time.Hour)

// Token reset password (berlaku 15 menit)
token, err := utilities.GenerateAccessToken(utilities.JwtPayload{
    ID:       user.ID,
    Name:     *user.Name,
    RoleType: "OTHER",
    Purpose:  "RESET_PASSWORD",
}, 15*time.Minute)
```

**Secret key** diambil dari `config.AppConfig.JWTSecret` (env variable `AUTH_JWT_SECRET`).

---

### `VerifyAccessToken(tokenString string) (*JwtPayload, error)`

Memverifikasi dan mendekode JWT token.

**Validasi yang dilakukan:**
1. Parse token dengan claims `*JwtPayload`
2. Verifikasi signing method harus HMAC (HS256)
3. Verifikasi signature dengan secret key
4. Verifikasi token masih valid (belum expired)

**Return:** `*JwtPayload` berisi semua data yang tersimpan di token

**Contoh penggunaan:**
```go
payload, err := utilities.VerifyAccessToken(tokenString)
if err != nil {
    // Token tidak valid atau sudah expired
    utilities.Unauthorized(c, "Invalid token")
    return
}

// Gunakan data dari payload
userID := payload.ID
roleType := payload.RoleType
purpose := payload.Purpose
```

---

## 📄 `utilities/response.go` — HTTP Response Helpers

Menyediakan fungsi-fungsi standar untuk mengembalikan response JSON yang konsisten di seluruh aplikasi.

### Struct `ResponsePayload`

Format response JSON standar yang digunakan di seluruh API.

```go
type ResponsePayload struct {
    Status  int         `json:"status"`
    Message string      `json:"message"`
    Data    interface{} `json:"data"`
    Error   interface{} `json:"error,omitempty"` // Hanya muncul jika ada error
}
```

**Contoh response JSON:**
```json
{
  "status": 200,
  "message": "Success",
  "data": { ... }
}
```

---

### Daftar Fungsi Response

#### `OK(c *gin.Context, data interface{}, message string)`
Response sukses `200 OK`.

```go
utilities.OK(c, user, "User retrieved")
// atau dengan default message "Success"
utilities.OK(c, data, "")
```

---

#### `Created(c *gin.Context, data interface{}, message string)`
Response data berhasil dibuat `201 Created`.

```go
utilities.Created(c, user, "User created successfully")
```

---

#### `BadRequest(c *gin.Context, message string, data interface{})`
Response input tidak valid `400 Bad Request`.

```go
utilities.BadRequest(c, "Invalid Input", err.Error())
utilities.BadRequest(c, "Email already exists", nil)
```

---

#### `ValidateError(c *gin.Context, data interface{})`
Response validasi error `400 Bad Request` dengan pesan default "Bad request".

```go
utilities.ValidateError(c, validationErrors)
```

---

#### `Unauthorized(c *gin.Context, message string)`
Response tidak terautentikasi `401 Unauthorized`.

```go
utilities.Unauthorized(c, "Token expired")
```

---

#### `Forbidden(c *gin.Context, message string)`
Response tidak memiliki izin `403 Forbidden`.

```go
utilities.Forbidden(c, "You don't have permission to access this resource")
```

---

#### `NotFound(c *gin.Context, message string)`
Response data tidak ditemukan `404 Not Found`.

```go
utilities.NotFound(c, "User not found")
```

---

#### `ServerError(c *gin.Context, err error, message string)`
Response error internal server `500 Internal Server Error`.

**Fitur khusus:**
- Di mode `development`: field `error` akan berisi detail error untuk debugging
- Di mode `production`: field `error` disembunyikan (`null`) untuk keamanan
- Error selalu di-log ke stdout menggunakan `log.Printf`

```go
utilities.ServerError(c, err, "Failed to create user")
```

---

#### `OtherResponse(c *gin.Context, status int, message string, data interface{})`
Response dengan status code custom (untuk status di luar standar).

```go
// Contoh: Token tidak valid (498 - custom status)
utilities.OtherResponse(c, 498, "Token invalid or expired", nil)
```

---

### Panduan Penggunaan

| Situasi | Fungsi |
|---------|--------|
| Data berhasil diambil | `utilities.OK(c, data, "message")` |
| Data berhasil dibuat | `utilities.Created(c, data, "message")` |
| Input dari user salah | `utilities.BadRequest(c, "message", errDetail)` |
| Tidak login / token salah | `utilities.Unauthorized(c, "message")` |
| Login tapi tidak punya izin | `utilities.Forbidden(c, "message")` |
| Data tidak ada di DB | `utilities.NotFound(c, "message")` |
| Error di sisi server | `utilities.ServerError(c, err, "message")` |
| Perlu status code custom | `utilities.OtherResponse(c, status, "message", data)` |

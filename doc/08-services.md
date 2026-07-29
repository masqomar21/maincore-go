# 08 — Services

Paket `services` berisi layanan-layanan eksternal yang digunakan oleh aplikasi, yaitu Background Queue (Asynq), S3 Object Storage, dan Socket.IO untuk komunikasi realtime.

---

## 📄 `services/queue.go` — Background Job Queue

Menggunakan library **Asynq** yang berbasis Redis untuk mengelola background jobs/tasks.

### Variabel Global

```go
var QueueClient *asynq.Client   // Untuk enqueue (mendaftarkan) task
var QueueServer *asynq.Server   // Untuk memproses task di background
```

### Konstanta Task Type

```go
const TypeAwsUpload = "upload:aws"  // Task untuk upload file ke S3
```

---

### `InitQueue()`

Menginisialisasi Asynq Client dan Server menggunakan konfigurasi Redis.

**Konfigurasi Server:**
```go
asynq.Config{
    Concurrency: 10,         // Maks 10 worker berjalan bersamaan
    Queues: map[string]int{
        "critical": 6,       // Prioritas tertinggi (60%)
        "default":  3,       // Prioritas menengah (30%)
        "low":      1,       // Prioritas rendah (10%)
    },
}
```

---

### `StartWorker()`

Mendaftarkan handler untuk setiap tipe task dan memulai server worker sebagai goroutine.

```go
mux.HandleFunc(TypeAwsUpload, HandleAwsUploadTask)
QueueServer.Start(mux)
```

> ⚙️ Fungsi ini dijalankan sebagai `go StartWorker()` di `cmd/api/main.go`.

---

### Struct `AwsUploadPayload`

Data yang dikirim sebagai payload task upload S3.

```go
type AwsUploadPayload struct {
    FilePath string `json:"file_path"` // Path file lokal temporary
    MimeType string `json:"mime_type"` // MIME type file
    DestKey  string `json:"dest_key"`  // Destination key di S3
}
```

---

### `EnqueueAwsUpload(filePath, mimeType, destKey string) error`

Mendaftarkan task upload S3 ke antrian background.

**Konfigurasi task:**
- `MaxRetry(3)` — Coba ulang maksimal 3 kali jika gagal
- `Timeout(5 * time.Minute)` — Timeout per eksekusi 5 menit

**Cara menggunakan:**
```go
err := services.EnqueueAwsUpload("/tmp/upload123.jpg", "image/jpeg", "uploads/photo.jpg")
if err != nil {
    log.Printf("Failed to enqueue upload: %v", err)
}
```

---

### `HandleAwsUploadTask(ctx context.Context, t *asynq.Task) error`

Handler yang dipanggil oleh worker saat memproses task upload S3.

**Alur:**
1. Deserialize payload dari task
2. (TODO) Implementasi upload file lokal ke S3 menggunakan `UploadFileToS3`
3. Log hasil proses

---

## 📄 `services/s3.go` — S3 Object Storage

Mendukung AWS S3 dan S3-compatible storage (MinIO, Cloudflare R2, dll.)

### Variabel Global

```go
var S3Client *s3.Client   // Instance AWS S3 client
```

---

### `InitS3()`

Menginisialisasi S3 client. Jika `S3_BUCKET` tidak dikonfigurasi, inisialisasi dilewati.

**Konfigurasi:**
- Region dari `S3_REGION`
- Credentials static dari `S3_ACCESS_KEY_ID` + `S3_SECRET_ACCESS_KEY`
- Custom endpoint dari `S3_ENDPOINT` (untuk non-AWS S3)
- Path style dari `S3_FORCE_PATH_STYLE` (set `true` untuk MinIO)

---

### `UploadFileToS3(ctx, file, header, path) (string, error)`

Mengupload file ke S3.

**Parameter:**
| Parameter | Tipe | Keterangan |
|-----------|------|------------|
| `ctx` | `context.Context` | Context request |
| `file` | `multipart.File` | File dari form upload |
| `header` | `*multipart.FileHeader` | Header file (nama, content-type) |
| `path` | `string` | Prefix path di bucket |

**Return:** `(key string, error)` — key adalah path file di S3

**Cara menggunakan di controller:**
```go
file, header, err := c.Request.FormFile("file")
if err != nil {
    utilities.BadRequest(c, "File required", nil)
    return
}
defer file.Close()

key, err := services.UploadFileToS3(c.Request.Context(), file, header, "uploads/images")
if err != nil {
    utilities.ServerError(c, err, "Upload failed")
    return
}
```

---

### `DeleteFileFromS3(ctx, key) error`

Menghapus file dari S3 berdasarkan key.

```go
err := services.DeleteFileFromS3(ctx, "uploads/images/photo.jpg")
```

---

## 📄 `services/socket.go` — Socket.IO Realtime

Menggunakan library `zishang520/socket.io` (kompatibel dengan Socket.IO v4 JavaScript client).

### Variabel Global

```go
var SocketServer *socket.Server
```

---

### `InitSocketServer() *socket.Server`

Membuat dan mengkonfigurasi Socket.IO server.

**Konfigurasi:**
- CORS: Allow all origins (`*`), dengan credentials
- EIO3 support: Mengizinkan client Socket.IO v3 dan v4

**Event handlers yang sudah terdaftar:**

| Event | Arah | Keterangan |
|-------|------|------------|
| `connection` | Server ← Client | Client berhasil terhubung |
| `join` | Server ← Client | Client bergabung ke room |
| `reply` | Server → Client | Konfirmasi bergabung ke room |
| `disconnect` | Server ← Client | Client terputus |

**Cara menambah event handler baru:**
```go
server.On("connection", func(clients ...any) {
    client := clients[0].(*socket.Socket)

    // Tambah event listener di sini
    client.On("my_event", func(datas ...any) {
        // Handle event
        client.Emit("response_event", "some data")

        // Broadcast ke semua client di room
        client.To("room_name").Emit("broadcast_event", "data")
    })
})
```

**Mengirim notifikasi realtime dari controller:**
```go
// Emit ke semua client yang terhubung
services.SocketServer.Emit("notification", gin.H{
    "message": "New notification",
    "userID":  userID,
})

// Emit ke room tertentu
services.SocketServer.To("user_1").Emit("private_msg", data)
```

**Route Socket.IO di HTTP server:**
```
GET  /socket.io/*any
POST /socket.io/*any
```

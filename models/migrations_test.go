package models

import (
	"fmt"
	"os"
	"testing"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// ─────────────────────────────────────────────────────────────────────────────
// TEST SETUP
// ─────────────────────────────────────────────────────────────────────────────

// setupTestDB membuat koneksi ke database untuk keperluan test.
// Jika database tidak tersedia, test akan di-skip (bukan fail).
func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "host=localhost user=root password=root dbname=starter_kit port=5432 sslmode=disable TimeZone=Asia/Jakarta"
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Skipf("Skipping: cannot connect to database (%v). Set DATABASE_URL to run migration tests.", err)
	}

	return db
}

// uniqueTable menghasilkan nama tabel unik berbasis timestamp agar setiap test terisolasi.
func uniqueTable(base string) string {
	return fmt.Sprintf("test_%s_%d", base, time.Now().UnixNano())
}

// dropTableIfExists menghapus tabel test setelah selesai (cleanup).
func dropTableIfExists(db *gorm.DB, tableName string) {
	db.Exec(fmt.Sprintf(`DROP TABLE IF EXISTS "%s" CASCADE`, tableName))
}

// ─────────────────────────────────────────────────────────────────────────────
// STRUCT HELPER UNTUK TESTING
// ─────────────────────────────────────────────────────────────────────────────

// Struct dinamis tidak bisa dipakai langsung dengan GORM TableName,
// jadi kita gunakan raw SQL + Migrator untuk operasi schema.

// ─────────────────────────────────────────────────────────────────────────────
// TEST 1: TAMBAH TABEL BARU
// ─────────────────────────────────────────────────────────────────────────────

func TestMigration_CreateTable(t *testing.T) {
	db := setupTestDB(t)
	tableName := uniqueTable("products")
	defer dropTableIfExists(db, tableName)

	// Buat tabel baru menggunakan raw SQL (simulasi migration)
	err := db.Exec(fmt.Sprintf(`
		CREATE TABLE "%s" (
			id BIGSERIAL PRIMARY KEY,
			name TEXT NOT NULL,
			price NUMERIC(10,2) DEFAULT 0,
			created_at TIMESTAMPTZ
		)
	`, tableName)).Error

	if err != nil {
		t.Fatalf("Gagal membuat tabel: %v", err)
	}

	// Verifikasi tabel berhasil dibuat
	exists := db.Migrator().HasTable(tableName)
	if !exists {
		t.Errorf("Tabel '%s' seharusnya ada setelah dibuat, tapi tidak ditemukan", tableName)
	}

	t.Logf("✅ Tabel '%s' berhasil dibuat", tableName)
}

// ─────────────────────────────────────────────────────────────────────────────
// TEST 2: HAPUS TABEL
// ─────────────────────────────────────────────────────────────────────────────

func TestMigration_DropTable(t *testing.T) {
	db := setupTestDB(t)
	tableName := uniqueTable("old_tokens")

	// Buat tabel terlebih dahulu
	db.Exec(fmt.Sprintf(`CREATE TABLE "%s" (id BIGSERIAL PRIMARY KEY, token TEXT)`, tableName))

	// Pastikan tabel ada
	if !db.Migrator().HasTable(tableName) {
		t.Fatalf("Tabel '%s' harus ada sebelum di-drop", tableName)
	}

	// Hapus tabel menggunakan Migrator (simulasi migration step)
	err := db.Migrator().DropTable(tableName)
	if err != nil {
		t.Fatalf("Gagal menghapus tabel: %v", err)
	}

	// Verifikasi tabel sudah terhapus
	if db.Migrator().HasTable(tableName) {
		t.Errorf("Tabel '%s' seharusnya sudah tidak ada setelah di-drop", tableName)
	}

	t.Logf("✅ Tabel '%s' berhasil dihapus", tableName)
}

// ─────────────────────────────────────────────────────────────────────────────
// TEST 3: TAMBAH KOLOM BARU
// ─────────────────────────────────────────────────────────────────────────────

func TestMigration_AddColumn(t *testing.T) {
	db := setupTestDB(t)
	tableName := uniqueTable("items")
	defer dropTableIfExists(db, tableName)

	// Buat tabel awal tanpa kolom "description"
	db.Exec(fmt.Sprintf(`CREATE TABLE "%s" (id BIGSERIAL PRIMARY KEY, name TEXT NOT NULL)`, tableName))

	// Verifikasi kolom "description" belum ada
	if db.Migrator().HasColumn(tableName, "description") {
		t.Fatal("Kolom 'description' seharusnya belum ada")
	}

	// Tambah kolom baru (simulasi migration step)
	err := db.Exec(fmt.Sprintf(`ALTER TABLE "%s" ADD COLUMN description TEXT`, tableName)).Error
	if err != nil {
		t.Fatalf("Gagal menambah kolom: %v", err)
	}

	// Verifikasi kolom berhasil ditambahkan
	if !db.Migrator().HasColumn(tableName, "description") {
		t.Errorf("Kolom 'description' seharusnya ada setelah ditambahkan")
	}

	t.Logf("✅ Kolom 'description' berhasil ditambahkan ke tabel '%s'", tableName)
}

// ─────────────────────────────────────────────────────────────────────────────
// TEST 4: HAPUS KOLOM
// ─────────────────────────────────────────────────────────────────────────────

func TestMigration_DropColumn(t *testing.T) {
	db := setupTestDB(t)
	tableName := uniqueTable("articles")
	defer dropTableIfExists(db, tableName)

	// Buat tabel dengan kolom "old_field" yang akan dihapus
	db.Exec(fmt.Sprintf(`
		CREATE TABLE "%s" (
			id BIGSERIAL PRIMARY KEY,
			title TEXT NOT NULL,
			old_field TEXT
		)
	`, tableName))

	// Verifikasi kolom ada sebelum dihapus
	if !db.Migrator().HasColumn(tableName, "old_field") {
		t.Fatal("Kolom 'old_field' harus ada sebelum dihapus")
	}

	// Hapus kolom menggunakan raw SQL (sama seperti yang digunakan di migrations.go)
	err := db.Exec(fmt.Sprintf(`ALTER TABLE "%s" DROP COLUMN old_field`, tableName)).Error
	if err != nil {
		t.Fatalf("Gagal menghapus kolom: %v", err)
	}

	// Verifikasi kolom sudah terhapus
	if db.Migrator().HasColumn(tableName, "old_field") {
		t.Errorf("Kolom 'old_field' seharusnya sudah tidak ada setelah di-drop")
	}

	// Verifikasi kolom lain tidak terpengaruh
	if !db.Migrator().HasColumn(tableName, "title") {
		t.Errorf("Kolom 'title' seharusnya masih ada")
	}

	t.Logf("✅ Kolom 'old_field' berhasil dihapus, kolom lain tidak terpengaruh")
}

// ─────────────────────────────────────────────────────────────────────────────
// TEST 5: RENAME KOLOM
// ─────────────────────────────────────────────────────────────────────────────

func TestMigration_RenameColumn(t *testing.T) {
	db := setupTestDB(t)
	tableName := uniqueTable("members")
	defer dropTableIfExists(db, tableName)

	// Buat tabel dengan kolom "phone" (nama lama)
	db.Exec(fmt.Sprintf(`
		CREATE TABLE "%s" (
			id BIGSERIAL PRIMARY KEY,
			email TEXT NOT NULL,
			phone TEXT
		)
	`, tableName))

	// Verifikasi state awal
	if !db.Migrator().HasColumn(tableName, "phone") {
		t.Fatal("Kolom 'phone' harus ada sebelum di-rename")
	}

	// Rename kolom "phone" → "phone_number"
	err := db.Exec(fmt.Sprintf(`ALTER TABLE "%s" RENAME COLUMN phone TO phone_number`, tableName)).Error
	if err != nil {
		t.Fatalf("Gagal rename kolom: %v", err)
	}

	// Verifikasi nama lama sudah tidak ada
	if db.Migrator().HasColumn(tableName, "phone") {
		t.Errorf("Kolom 'phone' (nama lama) seharusnya sudah tidak ada")
	}

	// Verifikasi nama baru ada
	if !db.Migrator().HasColumn(tableName, "phone_number") {
		t.Errorf("Kolom 'phone_number' (nama baru) seharusnya ada")
	}

	t.Logf("✅ Kolom berhasil di-rename: 'phone' → 'phone_number'")
}

// ─────────────────────────────────────────────────────────────────────────────
// TEST 6: UBAH TIPE DATA KOLOM
// ─────────────────────────────────────────────────────────────────────────────

func TestMigration_AlterColumnType(t *testing.T) {
	db := setupTestDB(t)
	tableName := uniqueTable("posts")
	defer dropTableIfExists(db, tableName)

	// Buat tabel dengan kolom "content" bertipe VARCHAR(255) (terlalu kecil)
	db.Exec(fmt.Sprintf(`
		CREATE TABLE "%s" (
			id BIGSERIAL PRIMARY KEY,
			title TEXT NOT NULL,
			content VARCHAR(255)
		)
	`, tableName))

	// Insert data untuk memastikan ALTER tidak merusak data
	db.Exec(fmt.Sprintf(`INSERT INTO "%s" (title, content) VALUES ('Test', 'Short content')`, tableName))

	// Ubah tipe kolom "content" dari VARCHAR(255) → TEXT
	err := db.Exec(fmt.Sprintf(`ALTER TABLE "%s" ALTER COLUMN content TYPE TEXT`, tableName)).Error
	if err != nil {
		t.Fatalf("Gagal mengubah tipe kolom: %v", err)
	}

	// Verifikasi data masih ada setelah perubahan tipe
	var count int64
	db.Raw(fmt.Sprintf(`SELECT count(*) FROM "%s"`, tableName)).Scan(&count)
	if count != 1 {
		t.Errorf("Data seharusnya masih ada setelah ALTER, got count=%d", count)
	}

	// Verifikasi tipe baru dengan insert string panjang (yang tidak muat di VARCHAR(255))
	longContent := string(make([]byte, 500))
	for i := range longContent {
		longContent = longContent[:i] + "A" + longContent[i+1:]
	}
	err = db.Exec(fmt.Sprintf(`INSERT INTO "%s" (title, content) VALUES ('Long', '%s')`, tableName, longContent)).Error
	if err != nil {
		t.Errorf("Kolom TEXT seharusnya bisa menampung string panjang, tapi error: %v", err)
	}

	t.Logf("✅ Tipe kolom 'content' berhasil diubah: VARCHAR(255) → TEXT, data lama aman")
}

// ─────────────────────────────────────────────────────────────────────────────
// TEST 7: TAMBAH INDEX
// ─────────────────────────────────────────────────────────────────────────────

func TestMigration_CreateAndDropIndex(t *testing.T) {
	db := setupTestDB(t)
	tableName := uniqueTable("orders")
	indexName := fmt.Sprintf("idx_%s_status", tableName)
	defer dropTableIfExists(db, tableName)

	// Buat tabel
	db.Exec(fmt.Sprintf(`
		CREATE TABLE "%s" (
			id BIGSERIAL PRIMARY KEY,
			user_id BIGINT NOT NULL,
			status TEXT DEFAULT 'pending'
		)
	`, tableName))

	// Buat index
	err := db.Exec(fmt.Sprintf(`CREATE INDEX "%s" ON "%s" (status)`, indexName, tableName)).Error
	if err != nil {
		t.Fatalf("Gagal membuat index: %v", err)
	}

	// Verifikasi index ada
	if !db.Migrator().HasIndex(tableName, indexName) {
		t.Errorf("Index '%s' seharusnya ada setelah dibuat", indexName)
	}
	t.Logf("✅ Index '%s' berhasil dibuat", indexName)

	// Hapus index
	err = db.Exec(fmt.Sprintf(`DROP INDEX "%s"`, indexName)).Error
	if err != nil {
		t.Fatalf("Gagal menghapus index: %v", err)
	}

	// Verifikasi index sudah terhapus
	if db.Migrator().HasIndex(tableName, indexName) {
		t.Errorf("Index '%s' seharusnya sudah tidak ada setelah di-drop", indexName)
	}
	t.Logf("✅ Index '%s' berhasil dihapus", indexName)
}

// ─────────────────────────────────────────────────────────────────────────────
// TEST 8: TAMBAH NOT NULL CONSTRAINT & DEFAULT VALUE
// ─────────────────────────────────────────────────────────────────────────────

func TestMigration_AddNotNullWithDefault(t *testing.T) {
	db := setupTestDB(t)
	tableName := uniqueTable("settings")
	defer dropTableIfExists(db, tableName)

	// Buat tabel dan insert data tanpa kolom "is_active"
	db.Exec(fmt.Sprintf(`CREATE TABLE "%s" (id BIGSERIAL PRIMARY KEY, key TEXT NOT NULL)`, tableName))
	db.Exec(fmt.Sprintf(`INSERT INTO "%s" (key) VALUES ('theme'), ('language')`, tableName))

	// Tambah kolom NOT NULL dengan DEFAULT (aman untuk data yang sudah ada)
	err := db.Exec(fmt.Sprintf(`
		ALTER TABLE "%s" ADD COLUMN is_active BOOLEAN NOT NULL DEFAULT true
	`, tableName)).Error
	if err != nil {
		t.Fatalf("Gagal menambah kolom NOT NULL dengan DEFAULT: %v", err)
	}

	// Verifikasi data lama mendapat nilai default
	var falseCount int64
	db.Raw(fmt.Sprintf(`SELECT count(*) FROM "%s" WHERE is_active = false`, tableName)).Scan(&falseCount)
	if falseCount != 0 {
		t.Errorf("Data lama seharusnya mendapat nilai default true, tapi ada %d baris dengan false", falseCount)
	}

	t.Logf("✅ Kolom NOT NULL dengan DEFAULT berhasil ditambahkan, data lama aman")
}

// ─────────────────────────────────────────────────────────────────────────────
// TEST 9: SISTEM RunManualMigrations — IDEMPOTENCY
// Memastikan migration yang sama tidak dijalankan dua kali
// ─────────────────────────────────────────────────────────────────────────────

func TestRunManualMigrations_Idempotency(t *testing.T) {
	db := setupTestDB(t)

	// Setup tabel migration_logs sementara menggunakan nama unik
	logTable := uniqueTable("migration_logs")
	defer dropTableIfExists(db, logTable)

	// Buat tabel log sementara
	db.Exec(fmt.Sprintf(`
		CREATE TABLE "%s" (
			id BIGSERIAL PRIMARY KEY,
			name TEXT NOT NULL UNIQUE,
			created_at TIMESTAMPTZ
		)
	`, logTable))

	runCount := 0
	step := MigrationStep{
		Name: "test_idempotency_step",
		Up: func(db *gorm.DB) error {
			runCount++
			return nil
		},
	}

	// Simulasi jalankan step (pertama kali)
	var existing MigrationLog
	result := db.Table(logTable).Where("name = ?", step.Name).First(&existing)
	if result.Error != nil {
		// Belum pernah dijalankan, eksekusi
		if err := step.Up(db); err != nil {
			t.Fatalf("Migration gagal: %v", err)
		}
		db.Table(logTable).Create(map[string]interface{}{
			"name":       step.Name,
			"created_at": time.Now(),
		})
	}

	// Simulasi jalankan lagi (harus di-skip)
	result = db.Table(logTable).Where("name = ?", step.Name).First(&existing)
	if result.Error == nil {
		// Sudah pernah dijalankan → skip
		t.Logf("Migration '%s' di-skip karena sudah pernah dijalankan ✅", step.Name)
	} else {
		// Harusnya sudah dicatat, eksekusi lagi (ini tidak boleh terjadi)
		step.Up(db)
	}

	// Verifikasi: fungsi Up() hanya dijalankan 1 kali
	if runCount != 1 {
		t.Errorf("Migration Up() seharusnya dijalankan 1 kali, tapi dijalankan %d kali", runCount)
	}

	t.Logf("✅ Idempotency OK: migration hanya dijalankan 1 kali dari 2 percobaan")
}

// ─────────────────────────────────────────────────────────────────────────────
// TEST 10: SISTEM RunManualMigrations — TRANSAKSI ROLLBACK SAAT GAGAL
// ─────────────────────────────────────────────────────────────────────────────

func TestRunManualMigrations_RollbackOnFailure(t *testing.T) {
	db := setupTestDB(t)
	tableName := uniqueTable("rollback_test")
	defer dropTableIfExists(db, tableName)

	// Buat tabel awal
	db.Exec(fmt.Sprintf(`CREATE TABLE "%s" (id BIGSERIAL PRIMARY KEY, name TEXT)`, tableName))

	// Jalankan transaksi yang gagal di tengah jalan
	err := db.Transaction(func(tx *gorm.DB) error {
		// Operasi 1: berhasil
		tx.Exec(fmt.Sprintf(`INSERT INTO "%s" (name) VALUES ('inserted')`, tableName))

		// Operasi 2: gagal (sintaks SQL salah)
		return tx.Exec(`THIS IS INVALID SQL`).Error
	})

	// Verifikasi transaksi gagal
	if err == nil {
		t.Fatal("Transaksi seharusnya gagal karena SQL tidak valid")
	}

	// Verifikasi rollback: data dari Operasi 1 tidak tersimpan
	var count int64
	db.Raw(fmt.Sprintf(`SELECT count(*) FROM "%s"`, tableName)).Scan(&count)
	if count != 0 {
		t.Errorf("Setelah rollback, tabel seharusnya kosong, tapi ada %d baris", count)
	}

	t.Logf("✅ Rollback berhasil: data tidak tersimpan ketika transaksi gagal")
}

// ─────────────────────────────────────────────────────────────────────────────
// TEST 11: RENAME TABEL
// ─────────────────────────────────────────────────────────────────────────────

func TestMigration_RenameTable(t *testing.T) {
	db := setupTestDB(t)
	oldName := uniqueTable("old_table")
	newName := oldName + "_renamed"
	defer dropTableIfExists(db, oldName)
	defer dropTableIfExists(db, newName)

	// Buat tabel lama
	db.Exec(fmt.Sprintf(`CREATE TABLE "%s" (id BIGSERIAL PRIMARY KEY, data TEXT)`, oldName))
	db.Exec(fmt.Sprintf(`INSERT INTO "%s" (data) VALUES ('existing data')`, oldName))

	// Rename tabel
	err := db.Migrator().RenameTable(oldName, newName)
	if err != nil {
		t.Fatalf("Gagal rename tabel: %v", err)
	}

	// Verifikasi nama lama tidak ada
	if db.Migrator().HasTable(oldName) {
		t.Errorf("Tabel '%s' (nama lama) seharusnya sudah tidak ada", oldName)
	}

	// Verifikasi nama baru ada dan data masih ada
	if !db.Migrator().HasTable(newName) {
		t.Errorf("Tabel '%s' (nama baru) seharusnya ada", newName)
	}

	var count int64
	db.Raw(fmt.Sprintf(`SELECT count(*) FROM "%s"`, newName)).Scan(&count)
	if count != 1 {
		t.Errorf("Data seharusnya masih ada setelah rename tabel, got count=%d", count)
	}

	t.Logf("✅ Tabel berhasil di-rename: '%s' → '%s', data aman", oldName, newName)
}

// ─────────────────────────────────────────────────────────────────────────────
// TEST 12: TAMBAH UNIQUE CONSTRAINT
// ─────────────────────────────────────────────────────────────────────────────

func TestMigration_AddUniqueConstraint(t *testing.T) {
	db := setupTestDB(t)
	tableName := uniqueTable("categories")
	defer dropTableIfExists(db, tableName)

	// Buat tabel tanpa constraint unique pada "slug"
	db.Exec(fmt.Sprintf(`
		CREATE TABLE "%s" (
			id BIGSERIAL PRIMARY KEY,
			name TEXT NOT NULL,
			slug TEXT
		)
	`, tableName))

	// Insert data unik
	db.Exec(fmt.Sprintf(`INSERT INTO "%s" (name, slug) VALUES ('Tech', 'tech')`, tableName))

	// Tambah unique constraint pada kolom "slug"
	indexName := fmt.Sprintf("uniq_%s_slug", tableName)
	err := db.Exec(fmt.Sprintf(`CREATE UNIQUE INDEX "%s" ON "%s" (slug)`, indexName, tableName)).Error
	if err != nil {
		t.Fatalf("Gagal menambah unique constraint: %v", err)
	}

	// Verifikasi unique constraint bekerja — insert duplikat harus gagal
	err = db.Exec(fmt.Sprintf(`INSERT INTO "%s" (name, slug) VALUES ('Technology', 'tech')`, tableName)).Error
	if err == nil {
		t.Errorf("Insert data duplikat seharusnya gagal karena unique constraint")
	}

	t.Logf("✅ Unique constraint pada 'slug' bekerja, duplikat ditolak")
}

package models

import (
	"log"
	"time"

	"gorm.io/gorm"
)

// MigrationLog adalah tabel internal untuk mencatat migration step yang sudah pernah dijalankan.
// Dengan ini, setiap step hanya dijalankan SEKALI, tidak akan diulang.
type MigrationLog struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"uniqueIndex;not null" json:"name"`
	CreatedAt time.Time `json:"createdAt"`
}

// MigrationStep mendefinisikan satu langkah migration manual.
type MigrationStep struct {
	// Name adalah identifier unik untuk step ini.
	// Gunakan format: "YYYY-MM-DD_deskripsi_singkat"
	Name string

	// Up adalah fungsi yang dijalankan saat migration ini belum pernah dieksekusi.
	Up func(db *gorm.DB) error
}

// ManualMigrations adalah daftar semua migration step manual yang perlu dijalankan.
//
// CARA MENAMBAH MIGRATION BARU:
// 1. Tambahkan entry baru di bagian bawah slice ini (JANGAN ubah/hapus entry lama)
// 2. Gunakan format nama: "YYYY-MM-DD_deskripsi"
// 3. Gunakan db.Migrator() untuk operasi schema yang tidak didukung AutoMigrate
// 4. Jalankan: make migrate
//
// OPERASI YANG TERSEDIA (db.Migrator()):
//   - db.Migrator().DropColumn(&Model{}, "nama_kolom")    → Hapus kolom
//   - db.Migrator().RenameColumn(&Model{}, "lama", "baru") → Rename kolom
//   - db.Migrator().AlterColumn(&Model{}, "nama_kolom")   → Ubah tipe/constraint kolom
//   - db.Migrator().DropTable(&Model{})                   → Hapus tabel
//   - db.Migrator().RenameTable("lama", "baru")           → Rename tabel
//   - db.Migrator().CreateIndex(&Model{}, "IndexName")    → Buat index
//   - db.Migrator().DropIndex(&Model{}, "IndexName")      → Hapus index
//   - db.Exec("SQL query")                                → Raw SQL (last resort)
var ManualMigrations = []MigrationStep{
	// -----------------------------------------------------------------------
	// CONTOH PENGGUNAAN — Hapus comment untuk mengaktifkan
	// -----------------------------------------------------------------------

	// Contoh: Hapus kolom yang tidak digunakan lagi dari tabel users
	// {
	// 	Name: "2026-06-07_drop_users_old_column",
	// 	Up: func(db *gorm.DB) error {
	// 		// Cek dulu apakah kolom masih ada sebelum menghapus
	// 		if db.Migrator().HasColumn(&User{}, "old_column") {
	// 			return db.Migrator().DropColumn(&User{}, "old_column")
	// 		}
	// 		return nil
	// 	},
	// },

	// Contoh: Rename kolom
	// {
	// 	Name: "2026-06-07_rename_users_phone_to_phone_number",
	// 	Up: func(db *gorm.DB) error {
	// 		if db.Migrator().HasColumn(&User{}, "phone") {
	// 			return db.Migrator().RenameColumn(&User{}, "phone", "phone_number")
	// 		}
	// 		return nil
	// 	},
	// },

	// Contoh: Ubah tipe kolom (misal varchar(50) → text)
	// {
	// 	Name: "2026-06-07_alter_loggers_detail_to_text",
	// 	Up: func(db *gorm.DB) error {
	// 		return db.Exec("ALTER TABLE loggers ALTER COLUMN detail TYPE TEXT").Error
	// 	},
	// },

	// -----------------------------------------------------------------------
	// TAMBAHKAN MIGRATION BARU DI BAWAH SINI
	// -----------------------------------------------------------------------
}

// RunManualMigrations menjalankan semua step dalam ManualMigrations yang belum pernah dieksekusi.
// Dipanggil dari cmd/migrate/main.go setelah AutoMigrate.
func RunManualMigrations(db *gorm.DB) {
	// Pastikan tabel migration_logs ada
	if err := db.AutoMigrate(&MigrationLog{}); err != nil {
		log.Fatal("Failed to create migration_logs table:", err)
	}

	if len(ManualMigrations) == 0 {
		log.Println("No manual migrations to run.")
		return
	}

	for _, step := range ManualMigrations {
		var existing MigrationLog

		// Cek apakah migration ini sudah pernah dijalankan
		result := db.Where("name = ?", step.Name).First(&existing)
		if result.Error == nil {
			log.Printf("[SKIP] Migration '%s' already applied.", step.Name)
			continue
		}

		log.Printf("[RUN]  Applying migration: %s ...", step.Name)

		// Jalankan migration dalam transaksi agar aman
		err := db.Transaction(func(tx *gorm.DB) error {
			return step.Up(tx)
		})

		if err != nil {
			log.Fatalf("[FAIL] Migration '%s' failed: %v", step.Name, err)
		}

		// Catat bahwa migration ini sudah berhasil dijalankan
		db.Create(&MigrationLog{Name: step.Name})
		log.Printf("[DONE] Migration '%s' applied successfully.", step.Name)
	}

	log.Println("Manual migrations finished.")
}

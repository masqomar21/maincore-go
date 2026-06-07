package main

import (
	"log"
	"maincore_go/config"
	"maincore_go/models"
)

func main() {
	log.Println("Initializing migration...")

	// Initialize Configuration
	config.InitConfig()

	// Initialize Database
	config.InitDatabase()

	// Run Auto-Migrate (membuat/update tabel secara otomatis)
	models.AutoMigrate(config.DB)

	// Run Manual Migrations (hapus kolom, rename, ubah tipe, dll.)
	models.RunManualMigrations(config.DB)

	log.Println("Migration command finished.")
}

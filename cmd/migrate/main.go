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

	// Run Auto-Migrate
	models.AutoMigrate(config.DB)
	
	log.Println("Migration command finished.")
}

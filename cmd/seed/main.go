package main

import (
	"log"
	"maincore_go/config"
	"maincore_go/models"
)

func main() {
	log.Println("Initializing seeding...")
	
	// Initialize Configuration
	config.InitConfig()

	// Initialize Database
	config.InitDatabase()

	// Run Seeders
	models.Seed(config.DB)
	
	log.Println("Seeding command finished.")
}

package models

import (
	"log"

	"gorm.io/gorm"
)


func AutoMigrate(db *gorm.DB) {
	err := db.AutoMigrate(
		&Permission{},
		&Role{},
		&RolePermission{},
		&User{},
		&Otp{},
		&Session{},
		&Logger{},
		&Notification{},
		&NotificationUser{},
		&WebPushSubscription{},
		&MigrationLog{}, // Tabel internal untuk mencatat manual migrations
	)
	if err != nil {
		log.Fatal("Failed to auto migrate database:", err)
	}
	log.Println("Database auto-migration completed successfully.")
}
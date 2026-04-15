package models

import (
	"log"

	"gorm.io/gorm"
	"golang.org/x/crypto/bcrypt"
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
	)
	if err != nil {
		log.Fatal("Failed to auto migrate database:", err)
	}
	log.Println("Database auto-migration completed successfully.")
}

func Seed(db *gorm.DB) {
	log.Println("Running seeders...")
	
	// Create default roles
	var role Role
	roleRes := db.FirstOrCreate(&role, Role{Name: "Super Admin", RoleType: RoleTypeSuperAdmin})
	if roleRes.Error != nil {
		log.Println("Failed to seed Role:", roleRes.Error)
	}
	
	// Create default permissions
	permissions := []Permission{
		{Name: "manage_users", Label: "Manage Users"},
		{Name: "manage_roles", Label: "Manage Roles"},
	}
	for i, p := range permissions {
		var existing Permission
		if err := db.Where("name = ?", p.Name).First(&existing).Error; err != nil {
			db.Create(&permissions[i])
		} else {
			permissions[i] = existing
		}
	}
	
	// Assign permissions to Role
	for _, p := range permissions {
		var rp RolePermission
		db.FirstOrCreate(&rp, RolePermission{
			RoleID: role.ID,
			PermissionID: p.ID,
			CanRead: true,
			CanWrite: true,
			CanUpdate: true,
			CanDelete: true,
			CanRestore: true,
		})
	}

	// Create default Super Admin User
	hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	pass := string(hash)
	name := "Super Admin"

	var user User
	db.FirstOrCreate(&user, User{
		Email: "superadmin@example.com",
		Name: &name,
		Password: &pass,
		RoleID: role.ID,
	})

	log.Println("Seeding completed successfully.")
}

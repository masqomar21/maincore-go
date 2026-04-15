package models

import (
	"time"

	"gorm.io/gorm"
)

type RoleType string

const (
	RoleTypeOther      RoleType = "OTHER"
	RoleTypeSuperAdmin RoleType = "SUPER_ADMIN"
)

type Permission struct {
	ID              uint             `gorm:"primaryKey" json:"id"`
	Name            string           `json:"name"`
	Label           string           `json:"label"`
	RolePermissions []RolePermission `gorm:"foreignKey:PermissionID" json:"rolePermissions,omitempty"`
}

type Role struct {
	ID              uint             `gorm:"primaryKey" json:"id"`
	Name            string           `json:"name"`
	RoleType        RoleType         `gorm:"type:varchar(50);default:'OTHER'" json:"roleType"`
	RolePermissions []RolePermission `gorm:"foreignKey:RoleID" json:"rolePermissions,omitempty"`
	Users           []User           `gorm:"foreignKey:RoleID" json:"users,omitempty"`
}

type RolePermission struct {
	ID           uint       `gorm:"primaryKey" json:"id"`
	RoleID       uint       `json:"roleId"`
	Role         Role       `gorm:"foreignKey:RoleID" json:"role,omitempty"`
	PermissionID uint       `json:"permissionId"`
	Permission   Permission `gorm:"foreignKey:PermissionID" json:"permission,omitempty"`
	CanRead      bool       `gorm:"default:false" json:"canRead"`
	CanWrite     bool       `gorm:"default:false" json:"canWrite"`
	CanUpdate    bool       `gorm:"default:false" json:"canUpdate"`
	CanDelete    bool       `gorm:"default:false" json:"canDelete"`
	CanRestore   bool       `gorm:"default:false" json:"canRestore"`
}

type User struct {
	ID                   uint                  `gorm:"primaryKey" json:"id"`
	Email                string                `gorm:"uniqueIndex;not null" json:"email"`
	Name                 *string               `json:"name"`
	Password             *string               `json:"-"`
	
	// 👇 CONTOH PENAMBAHAN FIELD BARU 👇
	Address              *string               `gorm:"type:text" json:"address"`
	PhoneNumber          *string               `gorm:"type:varchar(20)" json:"phoneNumber"`
	// 👆==============================👆
	RoleID               uint                  `json:"roleId"`
	Role                 Role                  `gorm:"foreignKey:RoleID" json:"role,omitempty"`
	RegisteredViaGoogle  bool                  `gorm:"default:false" json:"registeredViaGoogle"`
	CreatedAt            time.Time             `json:"createdAt"`
	UpdatedAt            time.Time             `json:"updatedAt"`
	DeletedAt            gorm.DeletedAt        `gorm:"index" json:"deletedAt,omitempty"`
	Sessions             []Session             `gorm:"foreignKey:UserID" json:"sessions,omitempty"`
	Loggers              []Logger              `gorm:"foreignKey:UserID" json:"loggers,omitempty"`
	Notifications        []NotificationUser    `gorm:"foreignKey:UserID" json:"notifications,omitempty"`
	WebPushSubscriptions []WebPushSubscription `gorm:"foreignKey:UserID" json:"webPushSubscriptions,omitempty"`
	OTP                  *Otp                  `gorm:"foreignKey:UserID" json:"otp,omitempty"`
}

type OtpPurpose string

const (
	OtpPurposeLogin         OtpPurpose = "LOGIN"
	OtpPurposeResetPassword OtpPurpose = "RESET_PASSWORD"
	OtpPurposeVerifyEmail   OtpPurpose = "VERIFY_EMAIL"
)

type Otp struct {
	ID        uint       `gorm:"primaryKey" json:"id"`
	UserID    uint       `gorm:"uniqueIndex" json:"userId"`
	User      User       `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Code      string     `gorm:"uniqueIndex;not null" json:"code"`
	Purpose   OtpPurpose `gorm:"type:varchar(50)" json:"purpose"`
	CreatedAt time.Time  `json:"createdAt"`
	ExpiresAt time.Time  `json:"expiresAt"`
}

type Session struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Token     string         `gorm:"uniqueIndex;not null" json:"token"`
	UserID    uint           `json:"userId"`
	User      User           `gorm:"foreignKey:UserID" json:"user,omitempty"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deletedAt,omitempty"`
}

type Process string

const (
	ProcessCreate  Process = "CREATE"
	ProcessUpdate  Process = "UPDATE"
	ProcessDelete  Process = "DELETE"
	ProcessRestore Process = "RESTORE"
	ProcessLogin   Process = "LOGIN"
	ProcessLogout  Process = "LOGOUT"
)

type Logger struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `json:"userId"`
	User      User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Process   Process   `gorm:"type:varchar(50)" json:"process"`
	Detail    string    `json:"detail"`
	CreatedAt time.Time `json:"createdAt"`
}

type Notification struct {
	ID         uint               `gorm:"primaryKey" json:"id"`
	Type       string             `json:"type"`
	RefID      *string            `json:"refId"`
	Message    string             `json:"message"`
	CreatedAt  time.Time          `json:"createdAt"`
	Recipients []NotificationUser `gorm:"foreignKey:NotificationID" json:"recipients,omitempty"`
}

type NotificationUser struct {
	ID             uint         `gorm:"primaryKey" json:"id"`
	UserID         uint         `gorm:"uniqueIndex:idx_user_notif" json:"userId"`
	NotificationID uint         `gorm:"uniqueIndex:idx_user_notif;index" json:"notificationId"`
	ReadStatus     bool         `gorm:"default:false;index:idx_user_read" json:"readStatus"`
	ReadAt         *time.Time   `json:"readAt"`
	DeliveredAt    *time.Time   `json:"deliveredAt"`
	User           User         `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE;" json:"user,omitempty"`
	Notification   Notification `gorm:"foreignKey:NotificationID;constraint:OnDelete:CASCADE;" json:"notification,omitempty"`
}

type WebPushSubscription struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	UserID         uint      `gorm:"index" json:"userId"`
	Endpoint       string    `gorm:"uniqueIndex;not null" json:"endpoint"`
	P256dh         string    `json:"p256dh"`
	Auth           string    `json:"auth"`
	ExpirationTime *time.Time `json:"expirationTime"`
	UserAgent      *string    `json:"userAgent"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
	User           User      `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE;" json:"user,omitempty"`
}

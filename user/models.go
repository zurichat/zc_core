package user

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	UserCollectionName        = "users"
	UserProfileCollectionName = "user_workspace_profiles"
)

type M map[string]interface{}

type Status int

const (
	Active Status = iota
	Suspended
	Disabled
)

type Role int

const (
	Super Role = iota
	Admin
	Member
)

type UserWorkspaceProfile struct {
	ID             primitive.ObjectID   `bson:"_id"`
	OrganizationID string               `bson:"organization_id"`
	DisplayPicture string               `bson:"display_picture"`
	Status         Status               `bson:"status"`
	Bio            string               `bson:"bio"`
	Timezone       string               `bson:"timezone"`
	Password       string               `bson:"password"`
	PasswordResets []*UserPasswordReset `bson:"password_resets"`
	Roles          []*Role              `bson:"roles"`
}

type UserRole struct {
	ID   primitive.ObjectID `bson:"_id"`
	Role Role               `bson:"role"`
}

type UserSettings struct {
	Role []UserRole `bson:"role"`
	//	Role Role
}

type UserEmailVerification struct {
	Verified  bool      `bson:"verified"`
	Token     string    `bson:"token"`
	ExpiredAt time.Time `bson:"expired_at"`
}

type UserPasswordReset struct {
	ID        primitive.ObjectID `bson:"_id"`
	IPAddress string             `bson:"ip_address"`
	Token     string             `bson:"token"`
	ExpiredAt time.Time          `bson:"expired_at"`
	UpdatedAt time.Time          `bson:"updated_at"`
	CreatedAt time.Time          `bson:"created_at"`
}

type User struct {
	_id               string                  `bson:"_id" json:"id"`
	FirstName         string                  `bson:"first_name" validate:"required,min=2,max=100" json:"first_name"`
	LastName          string                  `bson:"last_name" validate:"required,min=2,max=100" json:"last_name"`
	DisplayName       string                  `bson:"display_name" validate:"required,min=2,max=100" json:"display_name"`
	Email             string                  `bson:"email" validate:"email,required" json:"email"`
	Password          string                  `bson:"password" validate:"required,min=6"`
	Phone             string                  `bson:"phone" validate:"required" json:"phone"`
	Status            Status                  `bson:"status" json:"status"`
	Company           string                  `bson:"company" json:"company"`
	Settings          *UserSettings           `bson:"settings" json:"settings"`
	Timezone          string                  `bson:"time_zone" json:"time_zone"`
	CreatedAt         time.Time               `bson:"created_at" json:"created_at"`
	UpdatedAt         time.Time               `bson:"updated_at" json:"updated_at"`
	DeletedAt         time.Time               `bson:"deleted_at"`
	Organizations     []string                `bson:"workspaces"` // should contain (organization) workspace ids
	WorkspaceProfiles []*UserWorkspaceProfile `bson:"workspace_profiles"`
	EmailVerification UserEmailVerification   `bson:"email_verification"`
	PasswordResets    []*UserPasswordReset    `bson:"password_resets"`
}

// Struct that user can update directly
type UserUpdate struct {
	FirstName string `bson:"first_name" validate:"required,min=2,max=100" json:"first_name"`
	LastName  string `bson:"last_name" validate:"required,min=2,max=100" json:"last_name"`
	Phone     string `bson:"phone" validate:"required" json:"phone"`
	Company   string `bson:"company" json:"company"`
}

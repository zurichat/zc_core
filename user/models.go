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

type UserRole struct {
	ID   primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Role Role               `bson:"role"`
}

type UserSettings struct {
	Role []UserRole `bson:"role"`
	//	Role Role
}

type UserEmailVerification struct {
	ID        primitive.ObjectID `bson:"_id" json:"id"`
	Verified  bool               `bson:"verified"`
	Token     string             `bson:"token"`
	ExpiredAt time.Time          `bson:"expired_at"`
}

type UserPasswordReset struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	IPAddress string             `bson:"ip_address" json:"ip_address"`
	Token     string             `bson:"token" json:"token"`
	ExpiredAt time.Time          `bson:"expired_at" json:"expired_at"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
}

type User struct {
	ID                string     			  `bson:"_id,omitempty" json:"_id,omitempty"`
	FirstName         string                  `bson:"first_name" validate:"required,min=2,max=100" json:"first_name"`
	LastName          string                  `bson:"last_name" validate:"required,min=2,max=100" json:"last_name"`
	Email             string                  `bson:"email" validate:"email,required" json:"email"`
	Password          string                  `bson:"password" validate:"required,min=6"`
	Phone             string                  `bson:"phone" validate:"required" json:"phone"`
	Settings          *UserSettings           `bson:"settings" json:"settings"`
	Timezone          string                  `bson:"time_zone" json:"time_zone"`
	CreatedAt         time.Time               `bson:"created_at" json:"created_at"`
	UpdatedAt         time.Time               `bson:"updated_at" json:"updated_at"`
	Deactivated       string               	  `default:"false" bson:"deactivated" json:"deactivated"`
	DeactivatedAt     time.Time               `bson:"deactivated_at" json:"deactivated_at"`
	Organizations     []string                `bson:"workspaces"` // should contain (organization) workspace ids
	EmailVerification UserEmailVerification   `bson:"email_verification"`
	PasswordResets    []*UserPasswordReset    `bson:"password_resets"`
}

// Struct that user can update directly
type UserUpdate struct {
	FirstName string `bson:"first_name" validate:"required,min=2,max=100" json:"first_name"`
	LastName  string `bson:"last_name" validate:"required,min=2,max=100" json:"last_name"`
	Phone     string `bson:"phone" validate:"required" json:"phone"`
}

package user

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"zuri.chat/zccore/service"
	"zuri.chat/zccore/utils"
)

const (
	UserCollectionName                 = "users"
	UserProfileCollectionName          = "user_workspace_profiles"
	OrganizationsInvitesCollectionName = "organizations_invites"
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
	Verified  bool      `bson:"verified" json:"verified"`
	Token     string    `bson:"token" json:"token"`
	ExpiredAt time.Time `bson:"expired_at" json:"expired_at"`
}

type UserPasswordReset struct {
	IPAddress string    `bson:"ip_address" json:"ip_address"`
	Token     string    `bson:"token" json:"token"`
	ExpiredAt time.Time `bson:"expired_at" json:"expired_at"`
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
}

type Social struct {
	ID       string `bson:"provider_id" json:"provider_id"`
	Provider string `bson:"provider" json:"provider"`
}

type User struct {
	ID                string                 `bson:"_id,omitempty" json:"_id,omitempty"`
	FirstName         string                 `bson:"first_name" validate:"required,min=2,max=100" json:"first_name"`
	LastName          string                 `bson:"last_name" validate:"required,min=2,max=100" json:"last_name"`
	Email             string                 `bson:"email" validate:"email,required" json:"email"`
	Password          string                 `bson:"password" json:"password" validate:"required,min=6"`
	Phone             string                 `bson:"phone" validate:"required" json:"phone"`
	Settings          *UserSettings          `bson:"settings" json:"settings"`
	Timezone          string                 `bson:"time_zone" json:"time_zone"`
	Role              string                 `bson:"role" json:"role"`
	CreatedAt         time.Time              `bson:"created_at" json:"created_at"`
	UpdatedAt         time.Time              `bson:"updated_at" json:"updated_at"`
	Deactivated       bool                   `default:"false" bson:"deactivated" json:"deactivated"`
	DeactivatedAt     time.Time              `bson:"deactivated_at" json:"deactivated_at"`
	IsVerified        bool                   `bson:"isverified" json:"isverified"`
	Social            *Social                `bson:"social" json:"social"`
	Organizations     []string               `bson:"workspaces" json:"workspaces"` // should contain (organization) workspace ids
	EmailVerification *UserEmailVerification `bson:"email_verification" json:"email_verification"`
	PasswordResets    *UserPasswordReset     `bson:"password_resets" json:"password_resets"` // remove the array
}

// Struct that user can update directly
type UserUpdate struct {
	FirstName string `bson:"first_name" validate:"required,min=2,max=100" json:"first_name"`
	LastName  string `bson:"last_name" validate:"required,min=2,max=100" json:"last_name"`
	Phone     string `bson:"phone" validate:"required" json:"phone"`
}

type UserHandler struct {
	configs     *utils.Configurations
	mailService service.MailService
}

type UUIDUserData struct {
	Uuid      string `bson:"uuid" json:"uuid"`
	Password  string `bson:"password" json:"password"`
	FirstName string `bson:"first_name" json:"first_name"`
	LastName  string `bson:"last_name" json:"last_name"`
}

func NewUserHandler(c *utils.Configurations, mail service.MailService) *UserHandler {
	return &UserHandler{configs: c, mailService: mail}
}

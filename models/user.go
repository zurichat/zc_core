package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

/*
These models are by no means complete
*/

type EmailVerification struct {
	ID                primitive.ObjectID `bson:"_id"`
	UserID            string             `bson:"user_id"`
	Email             string             `bson:"email"`
	VerificationToken string             `bson:"verification_token"`
	Verified          bool               `bson:"verified"`
	Expires           time.Time          `bson:"expires"`
	CreatedAt         time.Time          `bson:"created_at"`
	UpdatedAt         time.Time          `bson:"updated_at"`
}

type PasswordReset struct {
	ID             primitive.ObjectID `bson:"_id"`
	IPAddress      string             `bson:"ip_address"`
	UserID         string             `bson:"user_id"`
	OrganizationID string             `bson:"organization_id"`
	ResetToken     string             `bson:"reset_token"`
	Expires        time.Time          `bson:"expires"`
	CreatedAt      time.Time          `bson:"created_at"`
	UpdatedAt      time.Time          `bson:"updated_at"`
}
type User struct {
	ID                  primitive.ObjectID     `bson:"_id"`
	Email               string                 `bson:"email"`
	EmailVerificationID string                 `bson:"email_verification_id"`
	PasswordResetIDs    []string               `bson:"password_reset_ids"`
	Settings            map[string]interface{} `bson:"settings"`
	CreatedAt           time.Time              `bson:"created_at"`
	UpdatedAt           time.Time              `bson:"updated_at"`
	DeletedAt           time.Time              `bson:"deleted_at"`
	WorkSpaces          []string               `bson:"workspaces"`
}

type WorkspaceUser struct {
	ID             primitive.ObjectID     `bson:"_id"`
	UserID         string                 `bson:"user_id"`
	FirstName      string                 `bson:"first_name"`
	LastName       string                 `bson:"last_name"`
	Email          string                 `bson:"email"`
	Password       string                 `bson:"password"`
	Phone          string                 `bson:"phone"`
	Status         string                 `bson:"status"`
	OrganizationID string                 `bson:"organization_id"`
	Settings       map[string]interface{} `bson:"settings"`
	CreatedAt      time.Time              `bson:"created_at"`
	UpdatedAt      time.Time              `bson:"updated_at"`
	DisplayName    string                 `bson:"display_name" `
	DisplayImage   string                 `bson:"display_image"`
	About          string                 `bson:"about"`
	Timezone       string                 `bson:"timezone"`
}

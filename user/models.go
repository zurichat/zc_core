package user

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
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

// The struct fields should be exported, don't you think? encoding and decoding to JSON might not work.
type WorkSpaceProfile struct {
	DisplayPicture string
	Status         Status
	Bio            string
	Timezone       string
	Password       string `bson:"password" validate:"required,min=6"`
	PasswordResets []*UserPasswordReset
	Roles          []*Role
}

type UserWorkspace struct {
	ID             primitive.ObjectID `bson:"_id"`
	OrganizationID string             // should this be an ID instead? Yes, objectID or string
	Profile        WorkSpaceProfile
}

type UserRole struct {
	ID   primitive.ObjectID `bson:"_id"`
	Role Role
}

type UserSettings struct {
	Role []UserRole
}

type UserEmailVerification struct {
	Verified  bool      `bson:"verified"`
	Token     string    `bson:"token"`
	ExpiredAt time.Time `bson:"expired_at"`
}

type UserPasswordReset struct {
	ID        primitive.ObjectID `bson:"_id"`
	IPAddress string
	Token     string
	ExpiredAt time.Time `bson:"expired_at"`
	UpdatedAt time.Time `bson:"updated_at"`
	CreatedAt time.Time `bson:"created_at"`
}

type User struct {
	_id                string 				`bson:"_id" json:"id"`
	FirstName         string             	`bson:"first_name" validate:"required,min=2,max=100" json:"first_name"`
	LastName          string             	`bson:"last_name" validate:"required,min=2,max=100" json:"last_name"`
	Email             string             	`bson:"email" validate:"email,required" json:"email"`
	Password          string             	`bson:"password" validate:"required,min=6"`
	Phone             string             	`bson:"phone" validate:"required" json:"phone"`
	Status            Status			 	`bson:"status" json:"status"`
	Company           string 			 	`bson:"company" json:"company"`
	Settings          *UserSettings		 	`bson:"settings" json:"settings"`
	Timezone          string			 	`bson:"time_zone" json:"time_zone"`
	CreatedAt         time.Time 		 	`bson:"created_at" json:"created_at"`
	UpdatedAt         time.Time 		 	`bson:"updated_at" json:"updated_at"`
	Organizations     []*UserWorkspace	 	`bson:"organizations" json:"organizations"`
	EmailVerification UserEmailVerification
	PasswordResets    []*UserPasswordReset
}

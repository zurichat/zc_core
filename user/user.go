package user

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type WorkSpaceProfile struct {
	display_picture string
	status          Status
	bio             string
	timezone        string
	password        string
	password_resets []UserPasswordReset
	roles           []Role
}

type UserWorkspace struct {
	ID           primitive.ObjectID `bson:"_id"`
	organization int                // should this be an ID instead?
	profile      WorkSpaceProfile
}

type Role int

const (
	Super Role = iota
	Admin
	Member
)

type UserRole struct {
	ID   primitive.ObjectID `bson:"_id"`
	role Role
}

type UserSettings struct {
	role []UserRole
}

type UserEmailVerification struct {
	verified   bool
	token      string
	expired_at primitive.Timestamp
}

type UserPasswordReset struct {
	ID         primitive.ObjectID `bson:"_id"`
	ipaddress  string
	token      string
	expired_at string `bson:"expired_at"`
	updated_at string `bson:"updated_at"`
	created_at string `bson:"created_at"`
}

type Status int

const (
	Active Status = iota
	Suspended
	Disabled
)

type User struct {
	ID                 primitive.ObjectID `bson:"_id"`
	first_name         string             `bson:"first_name" validate:"required,min=2,max=100"`
	last_name          string             `bson:"last_name" validate:"required,min=2,max=100"`
	email              string             `bson:"email" validate:"email,required"`
	password           string             `bson:"Password" validate:"required,min=6""`
	phone              string             `bson:"phone" validate:"required"`
	status             Status
	company            string
	settings           UserSettings
	timezone           string
	created_at         time.Time `bson:"created_at"`
	updated_at         time.Time `bson:"updated_at"`
	deleted_at         time.Time `bson:"deleted_at"`
	workspaces         []UserWorkspace
	email_verification UserEmailVerification
	password_resets    []UserPasswordReset
}

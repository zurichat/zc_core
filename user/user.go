package user

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserWorkspace struct {
	ID primitive.ObjectID `bson:"_id"`
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
	role UserRole
}

type Status int

const (
	Active Status = iota
	Suspended
	Disabled
)

type User struct {
	ID         primitive.ObjectID `bson:"_id"`
	first_name string
	last_name  string
	email      string
	password   string
	status     Status
	company    string
	settings   UserSettings
	timezone   string
	created_at time.Time
	updated_at time.Time
	deleted_at time.Time
	workspaces []UserWorkspace
}

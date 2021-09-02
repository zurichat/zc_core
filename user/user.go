package user

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"zuri.chat/zccore/utils"
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
	ID                primitive.ObjectID `bson:"_id"`
	FirstName         string             `bson:"first_name" validate:"required,min=2,max=100"`
	LastName          string             `bson:"last_name" validate:"required,min=2,max=100"`
	Email             string             `bson:"email" validate:"email,required"`
	Password          string             `bson:"password" validate:"required,min=6"`
	Phone             string             `bson:"phone" validate:"required"`
	Status            Status
	Company           string `bson:"company"`
	Settings          *UserSettings
	Timezone          string
	CreatedAt         time.Time `bson:"created_at"`
	UpdatedAt         time.Time `bson:"updated_at"`
	DeletedAt         time.Time `bson:"deleted_at"`
	Workspaces        []*UserWorkspace
	EmailVerification UserEmailVerification
	PasswordResets    []*UserPasswordReset
}

// helper functions perform CRUD operations on user
func FindUserByID(ctx context.Context, id string) (*User, error) {
	user := &User{}
	collectionName := "users"
	objID, _ := primitive.ObjectIDFromHex(id)
	collection := utils.GetCollection(collectionName)
	res := collection.FindOne(ctx, bson.M{"_id": objID})
	if err := res.Decode(user); err != nil {
		return nil, err
	}
	return user, nil
}

func FindUsers(ctx context.Context, filter M) ([]*User, error) {
	users := []*User{}
	collectionName := "users"
	collection := utils.GetCollection(collectionName)
	cursor, err := collection.Find(ctx, filter)

	if err != nil {
		return nil, err
	}
	if err = cursor.All(ctx, &users); err != nil {
		return nil, err
	}
	return users, nil
}

func CreateUser(ctx context.Context, u *User) error {
	collectionName := "users"
	collection := utils.GetCollection(collectionName)
	_, err := collection.InsertOne(ctx, u)
	if err != nil {
		return err
	}
	return nil
}

func FindUserProfile(ctx context.Context, userID, orgID string) (*UserWorkspace, error) {
	return nil, nil
}

func CreateUserProfile(ctx context.Context, uw *UserWorkspace) error {
	return nil
}

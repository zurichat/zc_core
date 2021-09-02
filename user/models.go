package user

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"zuri.chat/zccore/utils"
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
	ID                primitive.ObjectID      `bson:"_id"`
	FirstName         string                  `bson:"first_name" validate:"required,min=2,max=100"`
	LastName          string                  `bson:"last_name" validate:"required,min=2,max=100"`
	Email             string                  `bson:"email" validate:"email,required"`
	Password          string                  `bson:"password" validate:"required,min=6"`
	Phone             string                  `bson:"phone" validate:"required"`
	Status            Status                  `bson:"status"`
	Company           string                  `bson:"company"`
	Settings          *UserSettings           `bson:"settings"`
	Timezone          string                  `bson:"timezone"`
	CreatedAt         time.Time               `bson:"created_at"`
	UpdatedAt         time.Time               `bson:"updated_at"`
	DeletedAt         time.Time               `bson:"deleted_at"`
	Workspaces        []string                `bson:"workspaces"` // should contain (organization) workspace ids
	WorkspaceProfiles []*UserWorkspaceProfile `bson:"workspace_profiles"`
	EmailVerification UserEmailVerification   `bson:"email_verification"`
	PasswordResets    []*UserPasswordReset    `bson:"password_resets"`
}

// helper functions perform CRUD operations on user and user_workspace_profile
func findUserByID(ctx context.Context, id string) (*User, error) {
	user := &User{}
	objID, _ := primitive.ObjectIDFromHex(id)
	collection := utils.GetCollection(UserCollectionName)
	res := collection.FindOne(ctx, bson.M{"_id": objID})
	if err := res.Decode(user); err != nil {
		return nil, err
	}
	return user, nil
}

func findUsers(ctx context.Context, filter M) ([]*User, error) {
	users := []*User{}
	collection := utils.GetCollection(UserCollectionName)
	cursor, err := collection.Find(ctx, filter)

	if err != nil {
		return nil, err
	}
	if err = cursor.All(ctx, &users); err != nil {
		return nil, err
	}
	return users, nil
}

func createUser(ctx context.Context, u *User) error {
	collection := utils.GetCollection(UserCollectionName)
	_, err := collection.InsertOne(ctx, u)
	if err != nil {
		return err
	}
	return nil
}

func findUserProfile(ctx context.Context, userID, orgID string) (*UserWorkspaceProfile, error) {
	return nil, nil
}

func createUserProfile(ctx context.Context, uw *UserWorkspaceProfile) error {
	return nil
}

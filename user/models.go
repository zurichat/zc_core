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
	ID             primitive.ObjectID   `bson:"_id,omitempty" json:"id,omitempty"`
	OrganizationID string               `bson:"organization_id"`
	Status         Status               `bson:"status"`
	Bio            string               `bson:"bio"`
	Timezone       string               `bson:"timezone"`
	Password       string               `bson:"-"`
	PasswordHash   string               `bson:"password_hash"`
	PasswordResets []*UserPasswordReset `bson:"password_resets"`
	Roles          []*Role              `bson:"roles"`
}

func (uw *UserWorkspaceProfile) SetPassword() {

}

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
	ID                primitive.ObjectID      `bson:"_id,omitempty" json:"id,omitempty"`
	FirstName         string                  `bson:"first_name" validate:"required,min=2,max=100" json:"first_name"`
	LastName          string                  `bson:"last_name" validate:"required,min=2,max=100" json:"last_name"`
	Email             string                  `bson:"email" validate:"email,required" json:"email"`
	Password          string                  `bson:"password" validate:"required,min=6"`
	Phone             string                  `bson:"phone" validate:"required" json:"phone"`
	Settings          *UserSettings           `bson:"settings" json:"settings"`
	Timezone          string                  `bson:"time_zone" json:"time_zone"`
	CreatedAt         time.Time               `bson:"created_at" json:"created_at"`
	UpdatedAt         time.Time               `bson:"updated_at" json:"updated_at"`
	Deactivated       string               	  `bson:"deactivated"`
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
}

// helper functions perform CRUD operations on user
func findUserByID(ctx context.Context, id string) (*User, error) {
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

func findUserByEmail(ctx context.Context, email string) (*User, error) {
	user := &User{}
	collectionName := "users"
	collection := utils.GetCollection(collectionName)
	res := collection.FindOne(ctx, bson.M{"email": email})
	if err := res.Decode(user); err != nil {
		return nil, err
	}
	return user, nil
}

func findUsers(ctx context.Context, filter M) ([]*User, error) {
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

func findUserProfile(ctx context.Context, userID, orgID string) (*UserWorkspaceProfile, error) {
	return nil, nil
}

func createUserProfile(ctx context.Context, uw *UserWorkspaceProfile) error {
	return nil
}

package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type User struct {
	ID         primitive.ObjectID `json:"_id"`
	Email      string             `json:"email"`
	WorkSpaces []string           `json:"workspaces"`
}

type WorkspaceUser struct {
	ID        primitive.ObjectID `json:"_id"`
	UserID    string             `json:"user_id"`
	FirstName string             `json:"first_name"`
	LastName  string             `json:"last_name"`
	Email     string             `json:"email"`
	AvatarURL string             `json:"avatar_url"`
	Bio       string             `json:"bio"`
	Timezone  string             `json:"timezone"`
}

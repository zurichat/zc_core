package auth

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

//User struct is only used to hold data during authentication. it is not saved
type User struct {
	UserID   primitive.ObjectID `json:"id" bson:"_id"`
	Password string             `json:"password" bson:"password"`
	Email    string             `json:"email" bson:"email"`
}

// Session holds data of a particular user session
type Session struct {
	TokenUuid    string             `json:"token_uuid,omitempty" bson:"token_uuid,omitempty"`
	UserID       primitive.ObjectID `json:"user_id,omitempty" bson:"user_id,omitempty"`
	AccessToken  string             `json:"access_token" bson:"access_token"`
	RefreshToken string             `json:"refresh_token,omitempty" bson:"refresh_token"`
	ExpireOn     int64              `json:"expire_on" bson:"expire_on"`
	CreatedAt    time.Time          `json:"" bson:"created_at"`
}

type DBtoken struct {
	Key    string `bson:"key"`
	UserID string `bson:"userid"`
}

type TokenMetaData struct {
	AccessToken  string
	RefreshToken string
	AccessUuid   string
	RefreshUuid  string
	AtExpires    int64
	RtExpires    int64
}

type AccessDetails struct {
	AccessUuid string
	UserId     uint64
}

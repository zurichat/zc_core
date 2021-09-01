package auth

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// A struct containing the details of a user stored in the database
// password hashing still needs to be implemented
// the format mongo stored the updated and created field needs to be worked on. Probably create functions for this
// there's an enum or boolean field we still need researching to know how to store
// this struct for now allows the user not to supply all required details before registration
// only the ID for now is a unique field
type User struct {
	ID primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Firstname string `json:"first_name,omitempty" bson:"first_name,omitempty"`
	Lastname string `json:"last_name,omitempty" bson:"last_name,omitempty"`
	Email string `json:"email,omitempty" bson:"email,omitempty"`
	Password string `json:"password,omitempty" bson:"password,omitempty"`
	Phone string `json:"string,omitempty" bson:"string,omitempty"`
	Company string `json:"company,omitempty" bson:"company,omitempty"`
	Created time.Time `json:"created_at" bson:"created_at"`
	Updated time.Time `json:"updated_at" bson:"updated_at"`
}
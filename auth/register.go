package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client

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


// An end point to create new users
func CreateUserRegEndPoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Add("content-type", "application/json")
	response.WriteHeader(http.StatusOK)
	fmt.Fprintf(response, "This is an endpoint for creating new users")
	fmt.Println("This is an endpoint for creating new users")

	var user User
	json.NewDecoder(request.Body).Decode(&user)

	ctx, cancel := context.WithTimeout(context.Background(), 10 * time.Second)
	defer cancel()

	client, _ =  mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))

	// creates a database called auth and create a row called users for data storage
	collection := client.Database("auth").Collection("users")

	// inserts the user details
	result, _ := collection.InsertOne(ctx, user)
	json.NewEncoder(response).Encode(result)
}







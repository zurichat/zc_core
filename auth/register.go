package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"zuri.chat/zccore/utils"
)

// var client *mongo.Client

// An end point to create new users
func CreateUserRegEndPoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Add("content-type", "application/json")
	fmt.Fprintf(response, "This is an endpoint for creating new users")
	fmt.Println("This is an endpoint for creating new users")

	var user User
	json.NewDecoder(request.Body).Decode(&user)

	ctx, cancel := context.WithTimeout(context.Background(), 10 * time.Second)
	defer cancel()
	// creates a database called auth and create a row called users for data storage

	collection, _ := utils.GetMongoDbCollection("auth", "users")

	// inserts the user details
	result, _ := collection.InsertOne(ctx, user)
	json.NewEncoder(response).Encode(result)
}







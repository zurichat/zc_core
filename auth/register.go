package auth

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
	"zuri.chat/zccore/utils"
)

type User struct {
	ID       primitive.ObjectID `json:"_id,omitempty" bson:"_id"`
	Name     string             `json:"name" bson:"name" validate:"required"`
	Email    string             `json:"email" bson:"email" validate:"required"`
	Password string             `json:"password" bson:"password" validate:"required"`
}

func userSignup(response http.ResponseWriter, request *http.Request) {

	response.WriteHeader(http.StatusOK)
	var user User
	response.Header().Set("Content-Type", "application/json")

	json.NewDecoder(request.Body).Decode(&user)
	user.Password = getHash([]byte(user.Password))

	DbName, CollectionName := utils.Env("DB_NAME"), "users"
	collection, err := utils.GetMongoDbCollection(DbName, CollectionName)
	if err != nil {
		utils.GetError(err, StatusInternalServerError, response)
	}
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	cursor, err := collection.InsertOne(ctx, user)
	if err != nil {
		utils.GetError(err, 400, response)
		response.Write([]byte(`{"message":"` + err.Error() + `"}`))
		return
	}
	defer cursor.Close(ctx)
	json.NewEncoder(response).Encode(result)
}

func getHash(pwd []byte) string {
	hash, err := bcrypt.GenerateFromPassword(pwd, bcrypt.MinCost)
	if err != nil {
		log.Println(err)
	}
	return string(hash)
}

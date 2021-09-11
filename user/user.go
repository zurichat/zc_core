package user

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
	"zuri.chat/zccore/utils"
)

var (
	EMAIL_NOT_VALID = errors.New("Email address is not valid")
	HASHING_FAILED = errors.New("Failed to hashed password")
)

// Method to hash password
func GenerateHashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

// An end point to create new users
func Create(response http.ResponseWriter, request *http.Request) {
	response.Header().Add("content-type", "application/json")
	user_collection := "users"

	var user User
	err := utils.ParseJsonFromRequest(request, &user)
	if err != nil {
		utils.GetError(err, http.StatusUnprocessableEntity, response)
		return
	}

	userEmail := strings.ToLower(user.Email)
	if !utils.IsValidEmail(userEmail) {
		utils.GetError(EMAIL_NOT_VALID, http.StatusBadRequest, response)
		return
	}
	// confirm if user_email exists
	result, _ := utils.GetMongoDbDoc(user_collection, bson.M{"email": userEmail})
	if result != nil {
		utils.GetError(
			errors.New(fmt.Sprintf("Users with email %s exists!", userEmail)),
			http.StatusBadRequest, 
			response,
		)
		return
	}

	hashPassword, err := GenerateHashPassword(user.Password)
	if err != nil {
		utils.GetError(HASHING_FAILED, http.StatusBadRequest, response)
		return
	}

	user.CreatedAt = time.Now()
	user.Password = hashPassword
	detail, _ := utils.StructToMap(user)

	res, err := utils.CreateMongoDbDoc(user_collection, detail)

	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, response)
		return
	}

	utils.GetSuccess("user created", res, response)
}

// an endpoint to search other users
func SearchOtherUsers(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	query := params["query"]
	filter := bson.M{
		"$or": []bson.M{
			{"first_name": query},
			{"last_name": query},
			{"email": query},
			{"display_name": query},
		},
	}
	res, err := utils.GetMongoDbDocs(UserCollectionName, filter)
	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, w)
	}
	utils.GetSuccess("successful", res, w)
}

// an endpoint to delete a user record
func DeleteUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	params := mux.Vars(r)
	userId := params["user_id"]

	delete, err := utils.DeleteOneMongoDoc("users", userId)

	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, w)
		return
	}
	if delete.DeletedCount == 0 {
		utils.GetError(errors.New("operation failed"), http.StatusInternalServerError, w)
		return
	}

	utils.GetSuccess("User Deleted Succesfully", nil, w)
}

// endpoint to find user by ID
func GetUser(response http.ResponseWriter, request *http.Request) {
	// Find a user by user ID
	response.Header().Set("content-type", "application/json")

	collectionName := "users"

	params := mux.Vars(request)
	userId := params["user_id"]
	objId, err := primitive.ObjectIDFromHex(userId)

	if err != nil {
		utils.GetError(errors.New("invalid id"), http.StatusBadRequest, response)
		return
	}

	res, err := utils.GetMongoDbDoc(collectionName, bson.M{"_id": objId})
	if err != nil {
		utils.GetError(errors.New("user not found"), http.StatusInternalServerError, response)
		return
	}
	utils.GetSuccess("user retrieved successfully", res, response)

}

// an endpoint to update a user record

func UpdateUser(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	// Validate the user ID
	userID := mux.Vars(request)["user_id"]
	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		utils.GetError(errors.New("invalid user ID"), http.StatusBadRequest, response)
		return
	}

	collectionName := "users"
	userExist, err := utils.GetMongoDbDoc(collectionName, bson.M{"_id": objID})
	if err != nil {
		utils.GetError(errors.New("user does not exist"), http.StatusNotFound, response)
		return
	}
	if userExist == nil {
		utils.GetError(errors.New("user does not exist"), http.StatusBadRequest, response)
		return
	}

	var user UserUpdate
	if err := utils.ParseJsonFromRequest(request, &user); err != nil {
		utils.GetError(errors.New("bad update data"), http.StatusUnprocessableEntity, response)
		return
	}

	userMap, err := utils.StructToMap(user)
	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, response)
	}

	updateFields := make(map[string]interface{})
	for key, value := range userMap {
		if value != "" {
			updateFields[key] = value
		}
	}
	if len(updateFields) == 0 {
		utils.GetError(errors.New("empty/invalid user input data"), http.StatusBadRequest, response)
		return
	} else {
		updateRes, err := utils.UpdateOneMongoDbDoc(collectionName, userID, updateFields)
		if err != nil {
			utils.GetError(errors.New("user update failed"), http.StatusInternalServerError, response)
			return
		}
		utils.GetSuccess("user successfully updated", updateRes, response)
	}
}

// get all users
func GetUsers(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	collectionName := "users"
	res, _ := utils.GetMongoDbDocs(collectionName, nil)
	utils.GetSuccess("users retrieved successfully", res, response)
}

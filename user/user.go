package user

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
	"zuri.chat/zccore/utils"
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
	if !utils.IsValidEmail(user.Email) {
		utils.GetError(errors.New("email address is not valid"), http.StatusBadRequest, response)
		return
	}

	// confirm if user_email exists
	result, _ := utils.GetMongoDbDoc(user_collection, bson.M{"email": user.Email})
	if result != nil {
		fmt.Printf("users with email %s exists!", user.Email)
		utils.GetError(errors.New("operation failed"), http.StatusBadRequest, response)
		return
	}

	hashPassword, err := GenerateHashPassword(user.Password)
	if err != nil {
		utils.GetError(errors.New("Failed to hashed password"), http.StatusBadRequest, response)
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

func Retrive(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("Content-Type", "application/json")
	user_collection := "users"

	params := mux.Vars(request)
	userId := params["user_id"]
	objId, err := primitive.ObjectIDFromHex(userId)

	if err != nil {
		utils.GetError(errors.New("invalid id"), http.StatusBadRequest, response)
		return
	}

	retrive, err := utils.GetMongoDbDoc(user_collection, bson.M{"_id": objId})

	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, response)
		return
	}
	utils.GetSuccess("user retrieved successfully", retrive, response)
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

// helper functions perform CRUD operations on user
func FindUserByID(response http.ResponseWriter, request *http.Request) {
	// Find a user by user ID
	response.Header().Set("content-type", "application/json")

	collectionName := "users"
	userID := mux.Vars(request)["id"]
	objID, err := primitive.ObjectIDFromHex(userID)

	if err != nil {
		utils.GetError(err, http.StatusBadRequest, response)
		return
	}

	res, err := utils.GetMongoDbDoc(collectionName, bson.M{"_id": objID})
	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, response)
		return
	}
	utils.GetSuccess("User retrieved successfully", res, response)

}

func UpdateUser(response http.ResponseWriter, request *http.Request) {
	// Update a user of a given ID. Only certain fields, detailed in the
	// UserUpdate struct can be directly updated by a user without additional
	// functionality or permissions
	response.Header().Set("content-type", "application/json")

	collectionName := "users"
	userID := mux.Vars(request)["id"]
	objID, err := primitive.ObjectIDFromHex(userID)
	// Validate the user ID provided
	if err != nil {
		utils.GetError(err, http.StatusBadRequest, response)
		return
	}

	res, err := utils.GetMongoDbDoc(collectionName, bson.M{"_id": objID})
	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, response)
		return
	}
	if res != nil {
		// 2. Get user fields to be updated from request body
		var body UserUpdate
		err := json.NewDecoder(request.Body).Decode(&body)
		if err != nil {
			utils.GetError(err, http.StatusBadRequest, response)
			return
		}
		fmt.Printf("request body %v", body)

		// 3. Validate request body
		structValidator := validator.New()
		err = structValidator.Struct(body)

		if err != nil {
			utils.GetError(err, http.StatusBadRequest, response)
			return
		}

		// Convert body struct to interface
		var userInterface map[string]interface{}
		bytes, err := json.Marshal(body)
		if err != nil {
			utils.GetError(err, http.StatusInternalServerError, response)
		}
		json.Unmarshal(bytes, &userInterface)

		// 4. Update user
		updateRes, err := utils.UpdateOneMongoDbDoc(collectionName, objID.String(), userInterface)
		if err != nil {
			utils.GetError(err, http.StatusInternalServerError, response)
			return
		}
		utils.GetSuccess("User update successful", updateRes, response)
	}

}

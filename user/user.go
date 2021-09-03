package user

import (
	"errors"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"zuri.chat/zccore/utils"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

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

	user.CreatedAt = time.Now()

	detail, _ := utils.StructToMap(user)

	res, err := utils.CreateMongoDbDoc(user_collection, detail)

	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, response)
		return
	}

	utils.GetSuccess("user created", res, response)
}

// helper functions perform CRUD operations on user
func FindUserByID(response http.ResponseWriter, request *http.Request) {
	// Find a user by user ID
	response.Header().Set("content-type", "application/json")

	collectionName := "users"
	userID := mux.Vars(request)["id"]
	objID, _ := primitive.ObjectIDFromHex(userID)

	res, err := utils.GetMongoDbDoc(collectionName, bson.M{"_id": objID})
	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, response)
		return
	}
	utils.GetSuccess("User retrieved successfully", res, response)

}

func UpdateUser(response http.ResponseWriter, request *http.Request) {
	// Update a user of a given ID
	response.Header().Set("content-type", "application/json")

	collectionName := "users"
	userID := mux.Vars(request)["id"]
	objID, _ := primitive.ObjectIDFromHex(userID)

	res, err := utils.GetMongoDbDoc(collectionName, bson.M{"_id": objID})
	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, response)
		return
	}

	if res != nil {
		// 2. Get user fields to be updated from request body
		var body map[string]interface{}
		err := json.NewDecoder(request.Body).Decode(&body)
		if err != nil {
			utils.GetError(err, http.StatusBadRequest, response)
			return
		}

		// 2'. Do not allow email update

		// 3. Update user
		updateRes, err := utils.UpdateOneMongoDbDoc(collectionName, userID, body)
		if err != nil {
			utils.GetError(err, http.StatusInternalServerError, response)
			return
		}
		utils.GetSuccess("User update successful", updateRes, response)
	}

}

// func FindUsers(ctx context.Context, filter M) ([]*User, error) {
// 	users := []*User{}
// 	collectionName := "users"
// 	collection := utils.GetCollection(collectionName)
// 	cursor, err := collection.Find(ctx, filter)

// 	if err != nil {
// 		return nil, err
// 	}
// 	if err = cursor.All(ctx, &users); err != nil {
// 		return nil, err
// 	}
// 	return users, nil
// }

// func FindUserProfile(ctx context.Context, userID, orgID string) (*UserWorkspace, error) {
// 	return nil, nil
// }

// func CreateUserProfile(ctx context.Context, uw *UserWorkspace) error {
// 	return nil
// }

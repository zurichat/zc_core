package user

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"zuri.chat/zccore/utils"
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
// func FindUserByID(response http.ResponseWriter, request *http.Request) {
// 	user := &User{}
// 	collectionName := "users"
// 	objID, _ := primitive.ObjectIDFromHex(id)
// 	collection := utils.GetCollection(collectionName)
// 	res := collection.FindOne(ctx, bson.M{"_id": objID})
// 	if err := res.Decode(user); err != nil {
// 		return nil, err
// 	}
// 	return user, nil
// }

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

// An Endpoint to delete user Profile
func Delete(w http.ResponseWriter, r *http.Request) {

	collectionName := "users"
	params := mux.Vars(r)
	id := params["user_id"]

	err := utils.DeleteOneMongoDoc(collectionName, id)

	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, w)
		return
	}
}

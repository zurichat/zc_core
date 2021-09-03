package user

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/cloudinary/cloudinary-go"
	"github.com/cloudinary/cloudinary-go/api/admin"
	"github.com/cloudinary/cloudinary-go/api/admin/search"
	"github.com/cloudinary/cloudinary-go/api/uploader"
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

// helper function perform update workspace profile status
func SetWorkspaceProfileStatus(ctx context.Context, id string, status string) ([]*UserWorkspaceProfile, error) {
	userWorkspaceProfile := []*UserWorkspaceProfile{}

	collectionName := "userWorkspaceProfile"
	objID, _ := primitive.ObjectIDFromHex(id)
	collection := utils.GetCollection(collectionName)
	filter := bson.M{"_id": objID}
	update := bson.M{"$set": bson.M{"status": status}}

	res := collection.FindOneAndUpdate(ctx, filter, update)
	if err := res.Decode(&userWorkspaceProfile); err != nil {
		return nil, err
	}

	return userWorkspaceProfile, nil

}

func GenerateImageUrl(ctx context.Background, image string) (ImageUrl string, error) {

	CloudName := utils.Env("CloudName")
	APIKey := utils.Env("APIKey")
	APISecret := utils.Env("APISecret")

	var cld, err =  cloudinary.NewFromParams(CloudName, APIKey, APISecret)
	if err != nil {
		return err
	}

	response, err := cld.Upload.Upload(
        ctx,
        image,
        uploader.UploadParams{})
		
    if err != nil {
        return err
    }
​
    return response.SecureURL
} 

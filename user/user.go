package user

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-playground/validator"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
	"zuri.chat/zccore/service"
	"zuri.chat/zccore/utils"
)

var (
	validate         = validator.New()
	EMAIL_NOT_VALID  = errors.New("Email address is not valid")
	HASHING_FAILED   = errors.New("Failed to hashed password")
	CONFIRM_PASSWORD = errors.New("Password and confirm password must be the same, confirm and try again!")
)

// Method to hash password
func GenerateHashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

// An end point to create new users
func (uh *UserHandler) Create(response http.ResponseWriter, request *http.Request) {
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
			fmt.Errorf("user with email %s exists", userEmail),
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

	_, comfimationToken := utils.RandomGen(6, "d")

	con := &UserEmailVerification{false, comfimationToken, time.Now().Add(time.Minute * time.Duration(24))}

	user.CreatedAt = time.Now()
	user.Password = hashPassword
	user.Deactivated = false
	user.IsVerified = false
	user.EmailVerification = con
	user.Social = nil
	user.Timezone = "Africa/Lagos" // set default timezone
	detail, _ := utils.StructToMap(user)

	res, err := utils.CreateMongoDbDoc(user_collection, detail)

	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, response)
		return
	}
	// Email Service <- send confirmation mail
	msger := uh.mailService.NewMail(
		[]string{user.Email}, "Account Confirmation", service.MailConfirmation, map[string]interface{}{
			"Username": user.Email,
			"Code":     comfimationToken,
		})

	if err := uh.mailService.SendMail(msger); err != nil {
		fmt.Printf("Error occured while sending mail: %s", err.Error())
	}

	respse := map[string]interface{}{
		"user_id":        res.InsertedID,
		"verification_code": comfimationToken,
	}

	utils.GetSuccess("user created", respse, response)
}

// an endpoint to delete a user record
func (uh *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	params := mux.Vars(r)
	userId := params["user_id"]

	deactivateUpdate := bson.M{"deactivated": true, "deactivated_at": time.Now()}
	delete, err := utils.UpdateOneMongoDbDoc("users", userId, deactivateUpdate)

	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, w)
		return
	}
	if delete.ModifiedCount == 0 {
		utils.GetError(errors.New("operation failed"), http.StatusInternalServerError, w)
		return
	}

	utils.GetSuccess("User Deleted Succesfully", nil, w)
}

// endpoint to find user by ID
func (uh *UserHandler) GetUser(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("Access-Control-Allow-Origin", "*")
	response.Header().Set("Access-Control-Allow-Headers", "Content-Type,access-control-allow-origin, access-control-allow-headers")
	// Find a user by user ID
	response.Header().Set("content-type", "application/json")

	collectionName := "users"

	params := mux.Vars(request)
	userId := params["user_id"]
	objId, err := primitive.ObjectIDFromHex(userId)

	if err != nil {
		utils.GetError(errors.New("invalid user id"), http.StatusBadRequest, response)
		return
	}

	res, err := utils.GetMongoDbDoc(collectionName, bson.M{"_id": objId, "deactivated": false})

	if err != nil {
		utils.GetError(errors.New("user not found"), http.StatusNotFound, response)
		return
	}
	
    DeleteMapProps(res, []string{"password"})
	utils.GetSuccess("user retrieved successfully", res, response)
}

// an endpoint to update a user record

func (uh *UserHandler) UpdateUser(response http.ResponseWriter, request *http.Request) {
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
	}
	_, err = utils.UpdateOneMongoDbDoc(collectionName, userID, updateFields)
	if err != nil {
			utils.GetError(errors.New("user update failed"), http.StatusInternalServerError, response)
			return
		}
	utils.GetSuccess("user successfully updated", nil, response)
}

// get all users
func (uh *UserHandler) GetUsers(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("Access-Control-Allow-Origin", "*")
	response.Header().Set("Access-Control-Allow-Headers", "Content-Type,access-control-allow-origin, access-control-allow-headers")
	response.Header().Set("content-type", "application/json")

	collectionName := "users"
	res, _ := utils.GetMongoDbDocs(collectionName, bson.M{"deactivated": false})
	for _, doc := range res {
		DeleteMapProps(doc, []string{"password"})
	}
	utils.GetSuccess("users retrieved successfully", res, response)
}

// get a user organizations
func (uh *UserHandler) GetUserOrganizations(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	member_collection, organization_collection := "members", "organizations"

	params := mux.Vars(request)

	userEmail := strings.ToLower(params["email"])
	if !utils.IsValidEmail(userEmail) {
		utils.GetError(EMAIL_NOT_VALID, http.StatusBadRequest, response)
		return
	}

	// find user email in members collection.
	result, _ := utils.GetMongoDbDocs(member_collection, bson.M{"email": userEmail, "deleted": false})

	var orgs []map[string]interface{}
	for _, value := range result {
		basic := make(map[string]interface{})

		orgid := value["org_id"].(string)

		basic["isOwner"] = value["role"] == "owner"
		basic["member_id"] = value["_id"]

		objId, _ := primitive.ObjectIDFromHex(orgid)

		// find all members of an org
		orgMembers, _ := utils.GetMongoDbDocs(member_collection, bson.M{"org_id": orgid})

		orgDetails, err := utils.GetMongoDbDoc(organization_collection, bson.M{"_id": objId})
		if err != nil {
			utils.GetError(err, http.StatusUnprocessableEntity, response)
			return
		}

		// Get the images of all memebers of the organization
		var member_imgs []interface{}
		for _, member := range orgMembers {
			member_imgs = append(member_imgs, member["image_url"])
		}

		// Return 10 images or less
		if len(member_imgs) < 11 {
			basic["imgs"] = member_imgs
		} else {
			basic["imgs"] = member_imgs[:10]
		}

		basic["id"] = orgDetails["_id"]
		basic["logo_url"] = orgDetails["logo_url"]
		basic["name"] = orgDetails["name"]
		basic["workspace_url"] = orgDetails["workspace_url"]
		basic["no_of_members"] = len(orgMembers)

		orgs = append(orgs, basic)
	}

	utils.GetSuccess("user organizations retrieved successfully", orgs, response)
}

// Create a new user from UUID guest invite sent to user and a supplied password.
func (uh *UserHandler) CreateUserFromUUID(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("content-type", "application/json")

	user_collection, orgInvite := UserCollectionName, OrganizationsInvitesCollectionName

	var uRequest UUIDUserData
	err := utils.ParseJsonFromRequest(r, &uRequest)
	if err != nil {
		utils.GetError(err, http.StatusUnprocessableEntity, w)
		return
	}

	// Validate UUID
	_, err = utils.ValidateUUID(uRequest.Uuid)
	if err != nil {
		utils.GetError(errors.New("invalid uuid"), http.StatusBadRequest, w)
		return
	}

	// Check that UUID exists
	res, err := utils.GetMongoDbDoc(orgInvite, bson.M{"uuid": uRequest.Uuid})
	if err != nil {
		utils.GetError(errors.New("uuid does not exist"), http.StatusBadRequest, w)
		return
	}

	// Validate email
	email := res["email"].(string) // extract email from UUID
	userEmail := strings.ToLower(email)
	if !utils.IsValidEmail(userEmail) {
		utils.GetError(EMAIL_NOT_VALID, http.StatusBadRequest, w)
		return
	}

	// Check if user_email exists
	result, _ := utils.GetMongoDbDoc(user_collection, bson.M{"email": userEmail})
	if result != nil {
		utils.GetError(
			fmt.Errorf("user with email %s exists", userEmail),
			http.StatusBadRequest,
			w,
		)
		return
	}

	// UserEmailVerification
	_, comfimationToken := utils.RandomGen(6, "d")
	con := &UserEmailVerification{true, comfimationToken, time.Now().Add(time.Minute * time.Duration(24))}

	// Hash password
	hashPassword, err := GenerateHashPassword(uRequest.Password)
	if err != nil {
		utils.GetError(HASHING_FAILED, http.StatusBadRequest, w)
		return
	}

	user := &User{
		FirstName:         uRequest.FirstName,
		LastName:          uRequest.LastName,
		Email:             email,
		Password:          hashPassword,
		IsVerified:        true,
		EmailVerification: con,
		CreatedAt:         time.Now(),
		Deactivated:       false,
	}

	// Save user to DB
	data, _ := utils.StructToMap(user)
	resp, err := utils.CreateMongoDbDoc(user_collection, data)
	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, w)
		return
	}

	utils.GetSuccess("user successfully created", resp, w)

}


func DeleteMapProps(m map[string]interface{}, s []string) {
	for _, v := range s {
        delete(m, v)
	}
}

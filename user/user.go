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
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
	"zuri.chat/zccore/service"
	"zuri.chat/zccore/utils"
)

var (
	errEmailNotValid = errors.New("email address is not valid")
	errHashingFailed = errors.New("failed to hashed password")
)

// Method to hash password.
func GenerateHashPassword(password string) (string, error) {
	cost := 14
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), cost)

	return string(bytes), err
}

// An end point to create new users.
func (uh *UserHandler) Create(response http.ResponseWriter, request *http.Request) {
	response.Header().Add("content-type", "application/json")

	var user User
	err := utils.ParseJsonFromRequest(request, &user)

	if err != nil {
		utils.GetError(err, http.StatusUnprocessableEntity, response)
		return
	}

	userEmail := strings.ToLower(user.Email)
	if !utils.IsValidEmail(userEmail) {
		utils.GetError(errEmailNotValid, http.StatusBadRequest, response)
		return
	}

	// confirm if user_email exists
	result, _ := utils.GetMongoDbDoc(UserCollectionName, bson.M{"email": userEmail})
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
		utils.GetError(errHashingFailed, http.StatusBadRequest, response)
		return
	}

	randomNumberLimit := 6
	timeLimit := 24
	_, comfimationToken := utils.RandomGen(randomNumberLimit, "d")

	con := &UserEmailVerification{false, comfimationToken, time.Now().Add(time.Minute * time.Duration(timeLimit))}

	user.CreatedAt = time.Now()
	user.Password = hashPassword
	user.Deactivated = false
	user.IsVerified = false
	user.EmailVerification = con
	user.Social = nil
	user.Timezone = "Africa/Lagos" // set default timezone
	detail, _ := utils.StructToMap(user)

	res, err := utils.CreateMongoDbDoc(UserCollectionName, detail)

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
		fmt.Printf("Error occurred while sending mail: %s", err.Error())
	}

	respse := map[string]interface{}{
		"user_id":           res.InsertedID,
		"verification_code": comfimationToken,
	}

	utils.GetSuccess("user created", respse, response)
}

// an endpoint to delete a user record.
func (uh *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	params := mux.Vars(r)
	userID := params["user_id"]

	deactivateUpdate := bson.M{"deactivated": true, "deactivated_at": time.Now()}
	deactivate, err := utils.UpdateOneMongoDbDoc(UserCollectionName, userID, deactivateUpdate)

	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, w)
		return
	}

	if deactivate.ModifiedCount == 0 {
		utils.GetError(errors.New("operation failed"), http.StatusInternalServerError, w)
		return
	}

	utils.GetSuccess("User Deleted Successfully", nil, w)
}

// endpoint to find user by ID.
func (uh *UserHandler) GetUser(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("Access-Control-Allow-Origin", "*")
	response.Header().Set("Access-Control-Allow-Headers", "Content-Type,access-control-allow-origin, access-control-allow-headers")
	// Find a user by user ID
	response.Header().Set("content-type", "application/json")

	params := mux.Vars(request)
	userID := params["user_id"]
	objID, err := primitive.ObjectIDFromHex(userID)

	if err != nil {
		utils.GetError(errors.New("invalid user id"), http.StatusBadRequest, response)
		return
	}

	res, err := utils.GetMongoDbDoc(UserCollectionName, bson.M{"_id": objID, "deactivated": false})

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

	userExist, err := utils.GetMongoDbDoc(UserCollectionName, bson.M{"_id": objID})

	if err != nil {
		utils.GetError(errors.New("user does not exist"), http.StatusNotFound, response)
		return
	}

	if userExist == nil {
		utils.GetError(errors.New("user does not exist"), http.StatusBadRequest, response)
		return
	}

	var user UserUpdate
	if err = utils.ParseJsonFromRequest(request, &user); err != nil {
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

	_, err = utils.UpdateOneMongoDbDoc(UserCollectionName, userID, updateFields)

	if err != nil {
		utils.GetError(errors.New("user update failed"), http.StatusInternalServerError, response)

		return
	}

	utils.GetSuccess("user successfully updated", nil, response)
}

// get all users.
func (uh *UserHandler) GetUsers(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("Access-Control-Allow-Origin", "*")
	response.Header().Set("Access-Control-Allow-Headers", "Content-Type,access-control-allow-origin, access-control-allow-headers")
	response.Header().Set("content-type", "application/json")

	res, _ := utils.GetMongoDbDocs(UserCollectionName, bson.M{"deactivated": false})

	for _, doc := range res {
		DeleteMapProps(doc, []string{"password"})
	}

	utils.GetSuccess("users retrieved successfully", res, response)
}

// get a user organizations.
func (uh *UserHandler) GetUserOrganizations(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")

	params := mux.Vars(request)

	userEmail := strings.ToLower(params["email"])
	if !utils.IsValidEmail(userEmail) {
		utils.GetError(errEmailNotValid, http.StatusBadRequest, response)
		return
	}

	// find user email in members collection.
	result, _ := utils.GetMongoDbDocs(MemberCollectionName, bson.M{"email": userEmail, "deleted": false})

	orgs := make([]map[string]interface{}, 0)

	var imageLimit int64 = 11

	for _, value := range result {
		basic := make(map[string]interface{})
		orgid, _ := value["org_id"].(string)
		objID, _ := primitive.ObjectIDFromHex(orgid)

		// find all members of an org
		MembersLengthChannel, orgDetailsChannel := make(chan GUOCR), make(chan GUOCR)
		ImageUrlsChannel := make(chan GUOCR)

		go func() {
			orgMembers, err := utils.GetMongoDbDocs(MemberCollectionName, bson.M{"org_id": orgid})
			resp := GUOCR{
				Err:      err,
				Interger: len(orgMembers),
			}
			MembersLengthChannel <- resp
		}()

		go func(imageLimit int64) {
			var memberImgs []interface{}

			findOptions := options.Find().SetLimit(imageLimit)

			//nolint:gocritic //Grego: I need to reference
			// findOptions.SetProjection(bson.D{{"_id", 0}, {"image_url", 1}})
			orgMembersimages, err := utils.GetMongoDbDocs(MemberCollectionName, bson.M{"org_id": orgid}, findOptions)

			for _, member := range orgMembersimages {
				memberImgs = append(memberImgs, member["image_url"])
			}

			resp := GUOCR{
				Err:        err,
				Interfaces: memberImgs,
			}
			ImageUrlsChannel <- resp
		}(imageLimit)

		go func() {
			orgDetailsrt, err := utils.GetMongoDbDoc(OrganizationCollectionName, bson.M{"_id": objID})
			resp := GUOCR{
				Err:  err,
				Bson: orgDetailsrt,
			}
			orgDetailsChannel <- resp
		}()

		MembersLengthData, orgDetailsData, basicimagesdata := <-MembersLengthChannel, <-orgDetailsChannel, <-ImageUrlsChannel
		basic["no_of_members"], basic["isOwner"], basic["member_id"] = MembersLengthData.Interger, value["role"] == "owner", value["_id"]

		if MembersLengthData.Err != nil || orgDetailsData.Err != nil || basicimagesdata.Err != nil {
			utils.GetError(fmt.Errorf("query Failed, try again later"), http.StatusUnprocessableEntity, response)
			return
		}

		orgDetails := orgDetailsData.Bson
		basic["imgs"], basic["id"], basic["logo_url"], basic["name"], basic["workspace_url"] = basicimagesdata.Interfaces, orgDetails["_id"], orgDetails["logo_url"], orgDetails["name"], orgDetails["workspace_url"]

		orgs = append(orgs, basic)
	}

	utils.GetSuccess("user organizations retrieved successfully", orgs, response)
}

// Create a new user from UUID guest invite sent to user and a supplied password.
func (uh *UserHandler) CreateUserFromUUID(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("content-type", "application/json")

	var uRequest UUIDUserData
	err := utils.ParseJsonFromRequest(r, &uRequest)

	if err != nil {
		utils.GetError(err, http.StatusUnprocessableEntity, w)
		return
	}

	// Validate UUID
	_, err = utils.ValidateUUID(uRequest.UUID)
	if err != nil {
		utils.GetError(errors.New("invalid uuid"), http.StatusBadRequest, w)
		return
	}

	// Check that UUID exists
	res, err := utils.GetMongoDbDoc(OrganizationsInvitesCollectionName, bson.M{"uuid": uRequest.UUID})
	if err != nil {
		utils.GetError(errors.New("uuid does not exist"), http.StatusBadRequest, w)
		return
	}

	// Validate email
	email, _ := res["email"].(string) // extract email from UUID
	userEmail := strings.ToLower(email)

	if !utils.IsValidEmail(userEmail) {
		utils.GetError(errEmailNotValid, http.StatusBadRequest, w)
		return
	}

	// Check if user_email exists
	result, _ := utils.GetMongoDbDoc(UserCollectionName, bson.M{"email": userEmail})
	if result != nil {
		utils.GetError(
			fmt.Errorf("user with email %s exists", userEmail),
			http.StatusBadRequest,
			w,
		)

		return
	}

	// UserEmailVerification
	randomNumberLimit := 6
	timeLimit := 24
	_, comfimationToken := utils.RandomGen(randomNumberLimit, "d")
	con := &UserEmailVerification{true, comfimationToken, time.Now().Add(time.Minute * time.Duration(timeLimit))}

	// Hash password
	hashPassword, err := GenerateHashPassword(uRequest.Password)
	if err != nil {
		utils.GetError(errHashingFailed, http.StatusBadRequest, w)
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
	resp, err := utils.CreateMongoDbDoc(UserCollectionName, data)

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

package organizations

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/mitchellh/mapstructure"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"zuri.chat/zccore/auth"
	"zuri.chat/zccore/user"
	"zuri.chat/zccore/utils"
)

// Get an organization record
func GetOrganization(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	orgId := mux.Vars(r)["id"]
	objId, err := primitive.ObjectIDFromHex(orgId)

	if err != nil {
		utils.GetError(errors.New("invalid id"), http.StatusBadRequest, w)
		return
	}

	save, _ := utils.GetMongoDbDoc(OrganizationCollectionName, bson.M{"_id": objId})

	if save == nil {
		utils.GetError(fmt.Errorf("organization %s not found", orgId), http.StatusNotFound, w)
		return
	}

	var org Organization
	// convert bson to struct
	bsonBytes, _ := bson.Marshal(save)
	bson.Unmarshal(bsonBytes, &org)

	org.Plugins = org.OrgPlugins()

	utils.GetSuccess("organization retrieved successfully", org, w)
}

// Get an organization by url
func GetOrganizationByURL(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	orgURL := mux.Vars(r)["url"]

	data, err := utils.GetMongoDbDoc(OrganizationCollectionName, bson.M{"workspace_url": orgURL})
	if data == nil {
		fmt.Printf("workspace with url %s doesn't exist!", orgURL)
		utils.GetError(errors.New("organization does not exist"), http.StatusNotFound, w)
		return
	}

	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, w)
		return
	}

	var org Organization

	orgJson, _ := json.Marshal(data)
	json.Unmarshal(orgJson, &org)

	org.Plugins = org.OrgPlugins()

	utils.GetSuccess("organization retrieved successfully", org, w)
}

// Create an organization record
func Create(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	// loggedIn := r.Context().Value("user").(*auth.AuthUser)
	// loggedInUser, _ := auth.FetchUserByEmail(bson.M{"email": strings.ToLower(loggedIn.Email)})

	var newOrg Organization

	// Try to decode the request body into the struct. If there is an error,
	// respond to the client with the error message and a 400 status code.
	err := json.NewDecoder(r.Body).Decode(&newOrg)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// validate that email is not empty and it meets the format
	if !utils.IsValidEmail(newOrg.CreatorEmail) {
		utils.GetError(fmt.Errorf("invalid email format : %s", newOrg.CreatorEmail), http.StatusBadRequest, w)
		return
	}

	// generate workspace url
	newOrg.Name = "Zuri Chat"
	newOrg.WorkspaceURL = utils.GenWorkspaceUrl(newOrg.Name)

	userEmail := strings.ToLower(newOrg.CreatorEmail)
	userName := strings.Split(userEmail, "@")[0]

	// get creator id
	creator, _ := auth.FetchUserByEmail(bson.M{"email": userEmail})
	var ccreatorid string = creator.ID

	// extract user document
	// var luHexid, _ = primitive.ObjectIDFromHex(loggedInUser.ID.Hex())

	userDoc, _ := utils.GetMongoDbDoc(UserCollectionName, bson.M{"email": newOrg.CreatorEmail})
	if userDoc == nil {
		fmt.Printf("user with email %s does not exist!", newOrg.CreatorEmail)
		utils.GetError(errors.New("user with this email does not exist"), http.StatusBadRequest, w)
		return
	}

	newOrg.CreatorID = ccreatorid
	newOrg.CreatorEmail = userEmail
	newOrg.CreatedAt = time.Now()

	// convert to map object
	var inInterface map[string]interface{}
	inrec, _ := json.Marshal(newOrg)
	json.Unmarshal(inrec, &inInterface)

	// save organization
	save, err := utils.CreateMongoDbDoc(OrganizationCollectionName, inInterface)
	if err != nil {
		fmt.Println(err)
		utils.GetError(err, http.StatusInternalServerError, w)
		return
	}

	var iid interface{} = save.InsertedID
	var iiid string = iid.(primitive.ObjectID).Hex()

	// Adding user as a member
	var user user.User
	mapstructure.Decode(userDoc, &user)

	setting := new(Settings)

	newMember := Member{
		ID:       primitive.NewObjectID(),
		Email:    user.Email,
		UserName: userName,
		OrgId:    iiid,
		Role:     "owner",
		Presence: "true",
		Deleted:  false,
		Settings: setting,
	}

	// add new member to member collection
	coll := utils.GetCollection(MemberCollectionName)
	_, err = coll.InsertOne(r.Context(), newMember)
	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, w)
		return
	}

	// add organisation id to user organisations list
	updateFields := make(map[string]interface{})
	user.Organizations = append(user.Organizations, iiid)

	updateFields["Organizations"] = user.Organizations
	_, ee := utils.UpdateOneMongoDbDoc(UserCollectionName, ccreatorid, updateFields)
	if ee != nil {
		utils.GetError(errors.New("user update failed"), http.StatusInternalServerError, w)
		return
	}

	utils.GetSuccess("organization created", save, w)
}

// Get all organization records
func GetOrganizations(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	save, err := utils.GetMongoDbDocs(OrganizationCollectionName, nil)
	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, w)
		return
	}

	utils.GetSuccess("organizations retrieved successfully", save, w)
}

// Delete an organization record
func DeleteOrganization(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	orgId := mux.Vars(r)["id"]

	response, err := utils.DeleteOneMongoDoc(OrganizationCollectionName, orgId)

	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, w)
		return
	}

	if response.DeletedCount == 0 {
		utils.GetError(errors.New("operation failed"), http.StatusInternalServerError, w)
		return
	}

	utils.GetSuccess("organization deleted successfully", nil, w)
}

// Update an organization workspace url
func UpdateUrl(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	orgId := mux.Vars(r)["id"]
	requestData := make(map[string]string)
	if err := utils.ParseJsonFromRequest(r, &requestData); err != nil {
		utils.GetError(err, http.StatusUnprocessableEntity, w)
		return
	}

	org_filter := make(map[string]interface{})
	org_filter["workspace_url"] = requestData["url"]
	update, err := utils.UpdateOneMongoDbDoc(OrganizationCollectionName, orgId, org_filter)
	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, w)
		return
	}

	if update.ModifiedCount == 0 {
		utils.GetError(errors.New("operation failed"), http.StatusInternalServerError, w)
		return
	}

	utils.GetSuccess("organization url updated successfully", nil, w)
}

// Update organization name
func UpdateName(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	orgId := mux.Vars(r)["id"]

	requestData := make(map[string]string)
	if err := utils.ParseJsonFromRequest(r, &requestData); err != nil {
		utils.GetError(err, http.StatusUnprocessableEntity, w)
		return
	}

	org_filter := make(map[string]interface{})
	org_filter["name"] = requestData["organization_name"]

	update, err := utils.UpdateOneMongoDbDoc(OrganizationCollectionName, orgId, org_filter)
	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, w)
		return
	}

	if update.ModifiedCount == 0 {
		utils.GetError(errors.New("operation failed"), http.StatusInternalServerError, w)
		return
	}

	utils.GetSuccess("organization name updated successfully", nil, w)
}

// Update organization logo
func UpdateLogo(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")
	orgId := mux.Vars(r)["id"]

	requestData := make(map[string]string)
	if err := utils.ParseJsonFromRequest(r, &requestData); err != nil {
		utils.GetError(err, http.StatusUnprocessableEntity, w)
		return
	}

	org_filter := make(map[string]interface{})
	org_filter["logo_url"] = requestData["organization_logo"]

	update, err := utils.UpdateOneMongoDbDoc(OrganizationCollectionName, orgId, org_filter)
	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, w)
		return
	}

	if update.ModifiedCount == 0 {
		utils.GetError(errors.New("operation failed"), http.StatusInternalServerError, w)
		return
	}

	utils.GetSuccess("organization logo updated successfully", nil, w)
}


// transfer workspace ownership
func TransferOwnership(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	org_collection, member_collection := "organizations", "members"

	org_Id := mux.Vars(r)["id"]

	// Check if organization id is valid
	orgId, err := primitive.ObjectIDFromHex(org_Id)
	if err != nil {
		utils.GetError(errors.New("invalid organization id"), http.StatusBadRequest, w)
		return
	}

	// Check if organization id exists in the database
	orgDoc, _ := utils.GetMongoDbDoc(org_collection, bson.M{"_id": orgId})
	if orgDoc == nil {
		utils.GetError(errors.New("organization does not exist"), http.StatusBadRequest, w)
		return
	}

	requestData := make(map[string]string)
	if err := utils.ParseJsonFromRequest(r, &requestData); err != nil {
		utils.GetError(err, http.StatusUnprocessableEntity, w)
		return
	}

	email := requestData["email"]
	
	// confirm if email supplied is valid
	if !utils.IsValidEmail(strings.ToLower(email)) {
		utils.GetError(errors.New("email is not valid"), http.StatusBadRequest, w)
		return
	}

	orgMember, err := FetchMember(bson.M{"email": email})

	if err != nil {
		utils.GetError(errors.New("user not a member of any work pace"), http.StatusBadRequest, w)
		return
	}

	// confirm if the proposed new owner is a member of the organization
	memberDoc, _ := utils.GetMongoDbDoc(member_collection, bson.M{"org_id": org_Id, "_id": orgMember.ID})
	if memberDoc == nil {
		utils.GetError(errors.New("this user is not a member of this work space"), http.StatusBadRequest, w)
		return
	}

	if orgMember.Role == "owner" {
		utils.GetError(errors.New("this user already owns this organization"), http.StatusBadRequest, w)
		return
	}

	memID := orgMember.ID

	memberID := memID.Hex()

	updateRes, err := utils.UpdateOneMongoDbDoc(member_collection, memberID, bson.M{"role":"owner"})

	if err != nil {
		utils.GetError(errors.New("operation failed"), http.StatusInternalServerError, w)
		return
	}

	if updateRes.ModifiedCount == 0 {
		utils.GetError(errors.New("could not upgrade member role"), http.StatusInternalServerError, w)
		return
	}

	utils.GetSuccess("workspace owner changed successfully", nil, w)
}
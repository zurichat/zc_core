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
	collection := "organizations"

	orgId := mux.Vars(r)["id"]
	objId, err := primitive.ObjectIDFromHex(orgId)

	if err != nil {
		utils.GetError(errors.New("invalid id"), http.StatusBadRequest, w)
		return
	}

	save, _ := utils.GetMongoDbDoc(collection, bson.M{"_id": objId})

	if save == nil {
		utils.GetError(fmt.Errorf("organization %s not found", orgId), http.StatusNotFound, w)
		return
	}

	orgPluginsMap, err := OrganizationPlugins(orgId)
	if err != nil {
		// org plugins not found.
		utils.GetError(err, http.StatusNotFound, w)
		return
	}

	var org Organization
	
	orgJson, _ := json.Marshal(save)
	json.Unmarshal(orgJson, &org)

	org.Plugins = orgPluginsMap

	utils.GetSuccess("organization retrieved successfully", org, w)
}

// Get an organization by url
func GetOrganizationByURL(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	orgURL := mux.Vars(r)["url"]

	data, err := utils.GetMongoDbDoc("organizations", bson.M{"workspace_url": orgURL})
	if data == nil {
		fmt.Printf("workspace with url %s doesn't exist!", orgURL)
		utils.GetError(errors.New("organization does not exist"), http.StatusNotFound, w)
		return
	}

	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, w)
		return
	}
	
	orgId := data["_id"].(string)

	orgPluginsMap, err := OrganizationPlugins(orgId)
	if err != nil {
		// org plugins not found.
		utils.GetError(err, http.StatusNotFound, w)
		return
	}

	var org Organization
	
	orgJson, _ := json.Marshal(data)
	json.Unmarshal(orgJson, &org)

	org.Plugins = orgPluginsMap

	utils.GetSuccess("organization retrieved successfully", org, w)
}

// Create an organization record
func Create(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	// loggedIn := r.Context().Value("user").(*auth.AuthUser)
	// loggedInUser, _ := auth.FetchUserByEmail(bson.M{"email": strings.ToLower(loggedIn.Email)})

	var newOrg Organization
	collection, user_collection, member_collection := "organizations", "users", "members"

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

	// creator
	creator, _ := auth.FetchUserByEmail(bson.M{"email": strings.ToLower(newOrg.CreatorEmail)})
	var creatorid interface{} = creator.ID
	var ccreatorid string = creatorid.(primitive.ObjectID).Hex()

	// extract user document
	// var luHexid, _ = primitive.ObjectIDFromHex(loggedInUser.ID.Hex())

	userDoc, _ := utils.GetMongoDbDoc(user_collection, bson.M{"email": newOrg.CreatorEmail})
	if userDoc == nil {
		fmt.Printf("user with email %s does not exist!", newOrg.CreatorEmail)
		utils.GetError(errors.New("user with this email does not exist!"), http.StatusBadRequest, w)
		return
	}

	newOrg.CreatorID = ccreatorid
	newOrg.CreatedAt = time.Now()

	// convert to map object
	var inInterface map[string]interface{}
	inrec, _ := json.Marshal(newOrg)
	json.Unmarshal(inrec, &inInterface)

	// save organization
	save, err := utils.CreateMongoDbDoc(collection, inInterface)
	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, w)
		return
	}

	var iid interface{} = save.InsertedID
	var iiid string = iid.(primitive.ObjectID).Hex()
	hexOrgid, _ := primitive.ObjectIDFromHex(iiid)

	// Adding user as a member
	var user user.User
	mapstructure.Decode(userDoc, &user)

	newMember := Member{
		Email:    user.Email,
		OrgId:    hexOrgid.Hex(),
		Role:     "owner",
		Presence: "true",
	}
	// conv to struct
	memStruc, err := utils.StructToMap(newMember)
	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, w)
		return
	}

	// add new member to member collection
	_, e := utils.CreateMongoDbDoc(member_collection, memStruc)
	if e != nil {
		utils.GetError(err, http.StatusInternalServerError, w)
		return
	}

	// add organisation id to user organisations list
	updateFields := make(map[string]interface{})
	user.Organizations = append(user.Organizations, iiid)
	updateFields["Organizations"] = user.Organizations
	_, ee := utils.UpdateOneMongoDbDoc(user_collection, ccreatorid, updateFields)
	if ee != nil {
		utils.GetError(errors.New("user update failed"), http.StatusInternalServerError, w)
		return
	}

	utils.GetSuccess("organization created", save, w)
}

// Get all organization records
func GetOrganizations(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	collection := "organizations"

	save, err := utils.GetMongoDbDocs(collection, nil)
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

	collection := "organizations"

	response, err := utils.DeleteOneMongoDoc(collection, orgId)

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

	collection := "organizations"
	org_filter := make(map[string]interface{})
	org_filter["workspace_url"] = requestData["url"]
	update, err := utils.UpdateOneMongoDbDoc(collection, orgId, org_filter)
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

	collection := "organizations"

	org_filter := make(map[string]interface{})
	org_filter["name"] = requestData["organization_name"]

	update, err := utils.UpdateOneMongoDbDoc(collection, orgId, org_filter)
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

	collection := "organizations"

	org_filter := make(map[string]interface{})
	org_filter["logo_url"] = requestData["organization_logo"]

	update, err := utils.UpdateOneMongoDbDoc(collection, orgId, org_filter)
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

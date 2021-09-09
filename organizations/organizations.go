package organizations

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/mitchellh/mapstructure"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"zuri.chat/zccore/auth"
	"zuri.chat/zccore/user"
	"zuri.chat/zccore/utils"
)

func GetOrganization(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	collection := "organizations"

	orgId := mux.Vars(r)["id"]
	objId, err := primitive.ObjectIDFromHex(orgId)

	if err != nil {
		utils.GetError(errors.New("invalid id"), http.StatusBadRequest, w)
		return
	}

	save, err := utils.GetMongoDbDocs(collection, bson.M{"_id": objId})

	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, w)
		return
	}
	utils.GetSuccess("organization retrieved successfully", save, w)
}

func Create(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	loggedInUser := r.Context().Value("user").(auth.AuthUser)

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

	// extract user document
	var luHexid, _ = primitive.ObjectIDFromHex(loggedInUser.ID.Hex())
	userDoc, _ := utils.GetMongoDbDoc(user_collection, bson.M{"_id": luHexid})
	if userDoc == nil {
		fmt.Printf("user with id %s does not exist!", newOrg.CreatorID)
		utils.GetError(errors.New("operation failed"), http.StatusBadRequest, w)
		return
	}

	newOrg.CreatorID = loggedInUser.ID.Hex()
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
		Email: user.Email,
		OrgId: hexOrgid,
		Role:  "owner",
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
	_, ee := utils.UpdateOneMongoDbDoc(user_collection, loggedInUser.ID.Hex(), updateFields)
	if ee != nil {
		utils.GetError(errors.New("user update failed"), http.StatusInternalServerError, w)
		return
	}

	utils.GetSuccess("organization created", save, w)
}

func GetOrganizations(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	collection := "organizations"

	save, err := utils.GetMongoDbDocs(collection, nil)
	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, w)
		return
	}

	utils.GetSuccess("organization retrieved successfully", save, w)
}

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

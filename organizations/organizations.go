package organizations

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
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

	userEmail := strings.ToLower(newOrg.CreatorEmail)
	userName := strings.Split(userEmail, "@")[0]

	// get creator id
	creator, _ := auth.FetchUserByEmail(bson.M{"email": userEmail})
	var ccreatorid string = creator.ID

	// extract user document
	// var luHexid, _ = primitive.ObjectIDFromHex(loggedInUser.ID.Hex())

	userDoc, _ := utils.GetMongoDbDoc(user_collection, bson.M{"email": newOrg.CreatorEmail})
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
	save, err := utils.CreateMongoDbDoc(collection, inInterface)
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
	coll := utils.GetCollection(member_collection)
	_, err = coll.InsertOne(r.Context(), newMember)
	if err != nil {
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

// gets the details of a member in a workspace
func FetchMemberByEmail(filter map[string]interface{}) (*Member, error) {
	member_collection := "members"
	member := &Member{}
	memberCollection, err := utils.GetMongoDbCollection(os.Getenv("DB_NAME"), member_collection)
	if err != nil {
		return member, err
	}
	result := memberCollection.FindOne(context.TODO(), filter)
	err = result.Decode(&member)
	return member, err
}


// transfer workspace ownership
func TransferOwnership(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	org_collection, member_collection := "organizations", "members"

	org_Id := mux.Vars(r)["id"]

	// Check if organization id is valid
	orgId, err := primitive.ObjectIDFromHex(org_Id)
	if err != nil {
		utils.GetError(errors.New("invalid id"), http.StatusBadRequest, w)
		return
	}

	// Check if organization id exists in the database
	orgDoc, _ := utils.GetMongoDbDoc(org_collection, bson.M{"_id": orgId})
	if orgDoc == nil {
		fmt.Printf("organization with ID: %s does not exist ", org_Id)
		utils.GetError(errors.New("operation failed"), http.StatusBadRequest, w)
		return
	}

	var member Member

	err = utils.ParseJsonFromRequest(r, &member)
	if err != nil {
		utils.GetError(err, http.StatusUnprocessableEntity, w)
		return
	}

	newOwner := strings.ToLower(member.Email)
	
	// confirm if email supplied is valid
	if !utils.IsValidEmail(newOwner) {
		utils.GetError(errors.New("email is not valid"), http.StatusBadRequest, w)
		return
	}

	orgMember, err := FetchMemberByEmail(bson.M{"email": newOwner})

	if err != nil {
		utils.GetError(errors.New("user not a member of any work pace"), http.StatusBadRequest, w)
	}

	fmt.Printf("%s, %s, %s", orgMember.OrgId, orgMember.ID, orgMember.Role)

	// confirm if the proposed new owner is a member of the organization
	memberDoc, _ := utils.GetMongoDbDoc(member_collection, bson.M{"org_id": orgId, "_id": orgMember.ID})
	if memberDoc == nil {
		fmt.Printf("member with email %s is not a member of this work space", newOwner)
		utils.GetError(errors.New("this user is not a member of this work space"), http.StatusBadRequest, w)
		return
	}

	if orgMember.Role == "owner" {
		utils.GetError(errors.New("this user already owns this organization"), http.StatusBadRequest, w)
		return
	}

	member.Role = "owner"
	member.MadeOwnerAt = time.Now()

	memberMap, err := utils.StructToMap(member)
	if err != nil {
		utils.GetError(errors.New("struct to map"), http.StatusInternalServerError, w)
	}

	updateFields := make(map[string]interface{})
	for key, value := range memberMap {
		if value != "" {
			updateFields[key] = value
		}
	}

	fmt.Printf("%s", updateFields)

	if len(updateFields) == 0 {
		utils.GetError(errors.New("empty/invalid input data"), http.StatusBadRequest, w)
		return
	}

	memID := orgMember.ID

	memberID := memID.String()

	fmt.Printf("%s member id in string", memberID)

	// update := bson.M{"role": "owner", "madeowner_at": time.Now()}

	// status := make(map[string]interface{})

	// status["role"] = "owner"
	// status["madeowner_at"] = time.Now()
	
	updateRes, err := utils.UpdateOneMongoDbDoc(member_collection, memberID, updateFields)

	if err != nil {
		utils.GetError(errors.New("failed to upgrade member status"), http.StatusInternalServerError, w)
		return
	}

	if updateRes.ModifiedCount == 0 {
		utils.GetError(errors.New("zero update count"), http.StatusInternalServerError, w)
		return
	}

	utils.GetSuccess("workspace owner changed successfully", nil, w)
}


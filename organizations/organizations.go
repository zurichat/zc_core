package organizations

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/mitchellh/mapstructure"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"zuri.chat/zccore/auth"
	"zuri.chat/zccore/service"
	"zuri.chat/zccore/user"
	"zuri.chat/zccore/utils"
)

// Get an organization record
func (oh *OrganizationHandler) GetOrganization(w http.ResponseWriter, r *http.Request) {
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
func (oh *OrganizationHandler) GetOrganizationByURL(w http.ResponseWriter, r *http.Request) {
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
func (oh *OrganizationHandler) Create(w http.ResponseWriter, r *http.Request) {
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
func (oh *OrganizationHandler) GetOrganizations(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	save, err := utils.GetMongoDbDocs(OrganizationCollectionName, nil)
	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, w)
		return
	}

	utils.GetSuccess("organizations retrieved successfully", save, w)
}

// Delete an organization record
func (oh *OrganizationHandler) DeleteOrganization(w http.ResponseWriter, r *http.Request) {
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
func (oh *OrganizationHandler) UpdateUrl(w http.ResponseWriter, r *http.Request) {
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
func (oh *OrganizationHandler) UpdateName(w http.ResponseWriter, r *http.Request) {
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
func (oh *OrganizationHandler) UpdateLogo(w http.ResponseWriter, r *http.Request) {

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

func (oh *OrganizationHandler) SendInvite(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	loggedInUser := r.Context().Value("user").(*auth.AuthUser)
	user, _ := auth.FetchUserByEmail(bson.M{"email": strings.ToLower(loggedInUser.Email)})
	sOrgId := mux.Vars(r)["id"]

	if !auth.IsAuthorized(user.ID, sOrgId, "admin", w) {
		return
	}
	var guests SendInviteBody

	err_ := utils.ParseJsonFromRequest(r, &guests)
	if err_ != nil {
		utils.GetError(err_, http.StatusUnprocessableEntity, w)
		return
	}

	orgId, err := primitive.ObjectIDFromHex(sOrgId)
	if err != nil {
		utils.GetError(errors.New("invalid id"), http.StatusBadRequest, w)
		return
	}
	org, _ := utils.GetMongoDbDoc(OrganizationCollectionName, bson.M{"_id": orgId})
	if org == nil {
		utils.GetError(fmt.Errorf("organization %s not found", orgId), http.StatusNotFound, w)
		return
	}

	var invalidEmails []interface{}
	var inviteIDs []interface{}
	for _, email := range guests.Emails {
		// Check the validity of email send
		if !utils.IsValidEmail(email) {
			// If Email is invalid append to list to invalid emails
			invalidEmails = append(invalidEmails, email)
			continue
		}
		// Generate new UUI for invite and
		uuid := utils.GenUUID()

		newInvite := Invite{OrgID: sOrgId, Uuid: uuid, Email: email}
		var invInterface map[string]interface{}
		inrec, _ := json.Marshal(newInvite)
		json.Unmarshal(inrec, &invInterface)
		// Save newly generated uuid and associated info in the database
		save, err := utils.CreateMongoDbDoc(OrganizationInviteCollection, invInterface)
		if err != nil {
			fmt.Println(err)
			utils.GetError(err, http.StatusInternalServerError, w)
			return
		}
		// Append new invite to array of generated invites
		inviteIDs = append(inviteIDs, save.InsertedID)
		// Parse data for customising email template
		inviteLink := fmt.Sprintf("https://%v/join/%s", org["workspace_url"], uuid)
		orgName := fmt.Sprintf("%v", org["name"])
		msger := oh.mailService.NewMail(
			[]string{email}, "Zuri Chat Workspace Invite", service.WorkspaceInvite,
			&service.MailData{
				Username:   loggedInUser.Email,
				OrgName:    orgName,
				InviteLink: template.URL(inviteLink),
			})
		// error with sending main
		if err := oh.mailService.SendMail(msger); err != nil {
			fmt.Printf("Error occured while sending mail: %s", err.Error())
		}

	}
	resonse := SendInviteResponse{InvalidEmails: invalidEmails, InviteIDs: inviteIDs}
	utils.GetSuccess("Organization invite operation result", resonse, w)

}

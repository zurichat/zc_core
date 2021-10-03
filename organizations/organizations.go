package organizations

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
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

	userDoc, _ := utils.GetMongoDbDoc(UserCollectionName, bson.M{"email": newOrg.CreatorEmail})
	if userDoc == nil {
		fmt.Printf("user with email %s does not exist!", newOrg.CreatorEmail)
		utils.GetError(errors.New("user with this email does not exist"), http.StatusBadRequest, w)
		return
	}

	newOrg.CreatorID = ccreatorid
	newOrg.CreatorEmail = userEmail
	newOrg.CreatedAt = time.Now()

	// initialize organization with 100 free tokens
	newOrg.Tokens = 100
	newOrg.Version = FreeVersion

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

	newMember := NewMember(user.Email, userName, iiid, OwnerRole, setting)

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

	bot := Member{
		ID:        primitive.NewObjectID(),
		OrgId:     iiid,
		FirstName: "TwitterBot",
		Role:      Bot,
		Presence:  "true",
		JoinedAt:  time.Now(),
		Deleted:   false,
	}

	// add bot as member of organization
	coll = utils.GetCollection(MemberCollectionName)
	_, err = coll.InsertOne(r.Context(), bot)
	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, w)
		return
	}

	utils.GetSuccess("organization created", utils.M{"organization_id": save.InsertedID}, w)
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
	eventChannel := fmt.Sprintf("organizations_%s", orgId)
	event := utils.Event{Identifier: orgId, Type: "Organization", Event: UpdateOrganizationUrl, Channel: eventChannel, Payload: make(map[string]interface{})}
	go utils.Emitter(event)

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
	eventChannel := fmt.Sprintf("organizations_%s", orgId)
	event := utils.Event{Identifier: orgId, Type: "Organization", Event: UpdateOrganizationName, Channel: eventChannel, Payload: make(map[string]interface{})}
	go utils.Emitter(event)

	utils.GetSuccess("organization name updated successfully", nil, w)
}

// transfer workspace ownership
func (oh *OrganizationHandler) TransferOwnership(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	org_Id := mux.Vars(r)["id"]

	// checks if the logged in user trying to make changes is the owner the workspaceace
	if !auth.IsAuthorized(org_Id, "owner", w, r) {
		return
	}

	// Checks if organization id is valid
	orgId, err := primitive.ObjectIDFromHex(org_Id)
	if err != nil {
		utils.GetError(errors.New("invalid organization id"), http.StatusBadRequest, w)
		return
	}

	// Checks if organization exists in the database
	orgDoc, _ := utils.GetMongoDbDoc(OrganizationCollectionName, bson.M{"_id": orgId})
	if orgDoc == nil {
		utils.GetError(errors.New("organization does not exist"), http.StatusBadRequest, w)
		return
	}

	requestData := make(map[string]string)
	if err := utils.ParseJsonFromRequest(r, &requestData); err != nil {
		utils.GetError(err, http.StatusUnprocessableEntity, w)
		return
	}

	// email is that of the proposed new owner
	email := requestData["email"]

	// confirms if email supplied is valid
	if !utils.IsValidEmail(strings.ToLower(email)) {
		utils.GetError(errors.New("email is not valid"), http.StatusBadRequest, w)
		return
	}

	// fetches the details of the proposed new owner patterned after member's struct
	orgMember, err := FetchMember(bson.M{"org_id": org_Id, "email": email})

	if err != nil {
		utils.GetError(errors.New("user not a member of this work space"), http.StatusBadRequest, w)
		return
	}

	// checks if proposed owner does not have an ownership status already
	if orgMember.Role == "owner" {
		utils.GetError(errors.New("this member already owns this organization"), http.StatusBadRequest, w)
		return
	}

	// member ID of the proposed new owner
	memberID := orgMember.ID.Hex()

	// upgrades status from member to owner
	updateRes, err := utils.UpdateOneMongoDbDoc(MemberCollectionName, memberID, bson.M{"role": OwnerRole})

	if err != nil {
		utils.GetError(errors.New("operation failed"), http.StatusInternalServerError, w)
		return
	}

	if updateRes.ModifiedCount == 0 {
		utils.GetError(errors.New("could not upgrade member's role"), http.StatusInternalServerError, w)
		return
	}

	// fetches details of the former owner so we can get keys to downgrade status to member
	// checks like isOwner and memberExists are not made since auth.IsAuthorized function already
	// this user pass marks
	loggedInUser := r.Context().Value("user").(*auth.AuthUser)
	formerOwner, _ := FetchMember(bson.M{"org_id": org_Id, "email": loggedInUser.Email})

	// ID of former owner
	formerOwnerID := formerOwner.ID.Hex()

	// role downgraded from owner to member
	update, err := utils.UpdateOneMongoDbDoc(MemberCollectionName, formerOwnerID, bson.M{"role": AdminRole})

	if err != nil {
		utils.GetError(errors.New("operation failed"), http.StatusInternalServerError, w)
		return
	}

	if update.ModifiedCount == 0 {
		utils.GetError(errors.New("could not downgrade owner's role"), http.StatusInternalServerError, w)
		return
	}

	// and we are done!!!
	utils.GetSuccess("workspace owner changed successfully", nil, w)
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
	eventChannel := fmt.Sprintf("organizations_%s", orgId)
	event := utils.Event{Identifier: orgId, Type: "Organization", Event: UpdateOrganizationLogo, Channel: eventChannel, Payload: make(map[string]interface{})}
	go utils.Emitter(event)

	utils.GetSuccess("organization logo updated successfully", nil, w)
}

func (oh *OrganizationHandler) SendInvite(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	loggedInUser := r.Context().Value("user").(*auth.AuthUser)
	// user, _ := auth.FetchUserByEmail(bson.M{"email": strings.ToLower(loggedInUser.Email)})
	sOrgId := mux.Vars(r)["id"]
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
		inviteLink := fmt.Sprintf("https://zuri.chat/invites/%s", uuid)
		orgName := fmt.Sprintf("%v", org["name"])

		msger := oh.mailService.NewMail(
			[]string{email}, "Zuri Chat Workspace Invite", service.WorkspaceInvite, map[string]interface{}{
				"Username":   loggedInUser.Email,
				"OrgName":    orgName,
				"InviteLink": inviteLink,
			})
		// error with sending main
		if err := oh.mailService.SendMail(msger); err != nil {
			fmt.Printf("Error occured while sending mail: %s", err.Error())
		}

	}
	resonse := SendInviteResponse{InvalidEmails: invalidEmails, InviteIDs: inviteIDs}
	utils.GetSuccess("Organization invite operation result", resonse, w)

}

func (oh *OrganizationHandler) UpgradeToPro(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	orgId := mux.Vars(r)["id"]

	// check whether organization is already pro member
	is_pro, err := IsProVersion(orgId)

	if err != nil {
		utils.GetError(err, http.StatusNotAcceptable, w)
		return
	}
	if is_pro {
		utils.GetError(errors.New("organisation already on pro version"), http.StatusBadRequest, w)
		return
	}

	if err := SubscriptionBilling(orgId, float64(ProSubscriptionRate)); err != nil {
		utils.GetError(err, http.StatusExpectationFailed, w)
	}

	update_data := make(map[string]interface{})
	update_data["version"] = ProVersion

	update, err := utils.UpdateOneMongoDbDoc(OrganizationCollectionName, orgId, update_data)
	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, w)
		return
	}

	if update.ModifiedCount == 0 {
		utils.GetError(errors.New("operation failed"), http.StatusInternalServerError, w)
		return
	}

	utils.GetSuccess("Organization successfully updated to pro", nil, w)
}

func IsProVersion(OrgId string) (bool, error) {
	OrgIdFromHex, err := primitive.ObjectIDFromHex(OrgId)
	if err != nil {
		return false, err
	}

	organization, err := FetchOrganization(bson.M{"_id": OrgIdFromHex})
	if err != nil {
		return false, err
	}

	return organization.Version == ProVersion, nil
}

func (oh *OrganizationHandler) SaveBillingSettings(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	orgId := mux.Vars(r)["id"]

	var billingSetting BillingSetting
	err := utils.ParseJsonFromRequest(r, &billingSetting)
	if err != nil {
		utils.GetError(err, http.StatusUnprocessableEntity, w)
		return
	}

	validate := validator.New()

	if err := validate.Struct(billingSetting); err != nil {
		utils.GetError(err, http.StatusBadRequest, w)
		return
	}

	billing := Billing{
		billingSetting,
	}

	loggedInUser := r.Context().Value("user").(*auth.AuthUser)
	if _, err := FetchMember(bson.M{"org_id": orgId, "email": loggedInUser.Email}); err != nil {
		utils.GetError(errors.New("access denied"), http.StatusNotFound, w)
		return
	}

	org_filter := make(map[string]interface{})
	org_filter["billing"] = billing

	update, err := utils.UpdateOneMongoDbDoc(OrganizationCollectionName, orgId, org_filter)
	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, w)
		return
	}

	if update.ModifiedCount == 0 {
		utils.GetError(errors.New("operation failed"), http.StatusUnprocessableEntity, w)
		return
	}

	utils.GetSuccess("organization billing settings updated successfully", nil, w)
}

func (oh *OrganizationHandler) UpdateOrganizationSettings(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	orgId := mux.Vars(r)["id"]

	var orgSettings OrgSettings
	err := utils.ParseJsonFromRequest(r, &orgSettings)
	if err != nil {
		utils.GetError(err, http.StatusUnprocessableEntity, w)
		return
	}

	objId, err := primitive.ObjectIDFromHex(orgId)

	if err != nil {
		utils.GetError(err, http.StatusUnprocessableEntity, w)
		return
	}

	validate := validator.New()

	// get previous settings
	save, _ := utils.GetMongoDbDoc(OrganizationCollectionName, bson.M{"_id": objId})

	if save == nil {
		utils.GetError(fmt.Errorf("organization %s not found", orgId), http.StatusNotFound, w)
		return
	}

	var org Organization
	// convert bson to struct
	bsonBytes, _ := bson.Marshal(save)
	bson.Unmarshal(bsonBytes, &org)

	//valdate struct
	if err := validate.Struct(org); err != nil {
		utils.GetError(err, http.StatusBadRequest, w)
		return
	}
	//adds new settings with existing settings
	orgPref := OrganizationPreference{
		orgSettings,
		org.Settings.Permissions,
		org.Settings.Authentication,
	}

	org_filter := make(map[string]interface{})
	org_filter["settings"] = orgPref

	update, err := utils.UpdateOneMongoDbDoc(OrganizationCollectionName, orgId, org_filter)
	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, w)
		return
	}

	if update.ModifiedCount == 0 {
		utils.GetError(errors.New("operation failed"), http.StatusUnprocessableEntity, w)
		return
	}

	utils.GetSuccess("organization settings updated successfully", nil, w)
}

func (oh *OrganizationHandler) UpdateOrganizationPermission(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	orgId := mux.Vars(r)["id"]

	var orgPermissions OrgPermissions
	err := utils.ParseJsonFromRequest(r, &orgPermissions)
	if err != nil {
		utils.GetError(err, http.StatusUnprocessableEntity, w)
		return
	}

	objId, err := primitive.ObjectIDFromHex(orgId)
	if err != nil {
		utils.GetError(err, http.StatusUnprocessableEntity, w)
		return
	}

	validate := validator.New()

	// get previous settings
	save, _ := utils.GetMongoDbDoc(OrganizationCollectionName, bson.M{"_id": objId})

	if save == nil {
		utils.GetError(fmt.Errorf("organization %s not found", orgId), http.StatusNotFound, w)
		return
	}

	var org Organization
	// convert bson to struct
	bsonBytes, _ := bson.Marshal(save)
	bson.Unmarshal(bsonBytes, &org)

	//valdate struct
	if err := validate.Struct(org); err != nil {
		utils.GetError(err, http.StatusBadRequest, w)
		return
	}

	//adds new settings with existing settings
	orgPref := OrganizationPreference{
		org.Settings.Settings,
		orgPermissions,
		org.Settings.Authentication,
	}

	org_filter := make(map[string]interface{})
	org_filter["settings"] = orgPref

	update, err := utils.UpdateOneMongoDbDoc(OrganizationCollectionName, orgId, org_filter)
	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, w)
		return
	}

	if update.ModifiedCount == 0 {
		utils.GetError(errors.New("operation failed"), http.StatusUnprocessableEntity, w)
		return
	}

	utils.GetSuccess("organization settings updated successfully", nil, w)
}

func (oh *OrganizationHandler) UpdateOrganizationAuthentication(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	orgId := mux.Vars(r)["id"]

	var orgAuthentication OrgAuthentication
	err := utils.ParseJsonFromRequest(r, &orgAuthentication)
	if err != nil {
		utils.GetError(err, http.StatusUnprocessableEntity, w)
		return
	}

	objId, err := primitive.ObjectIDFromHex(orgId)

	if err != nil {
		utils.GetError(err, http.StatusUnprocessableEntity, w)
		return
	}

	validate := validator.New()

	// get previous settings
	save, _ := utils.GetMongoDbDoc(OrganizationCollectionName, bson.M{"_id": objId})

	if save == nil {
		utils.GetError(fmt.Errorf("organization %s not found", orgId), http.StatusNotFound, w)
		return
	}

	var org Organization
	// convert bson to struct
	bsonBytes, _ := bson.Marshal(save)
	bson.Unmarshal(bsonBytes, &org)

	//valdate struct
	if err := validate.Struct(org); err != nil {
		utils.GetError(err, http.StatusBadRequest, w)
		return
	}
	//adds new settings with existing settings
	orgPref := OrganizationPreference{
		org.Settings.Settings,
		org.Settings.Permissions,
		orgAuthentication,
	}

	org_filter := make(map[string]interface{})
	org_filter["settings"] = orgPref

	update, err := utils.UpdateOneMongoDbDoc(OrganizationCollectionName, orgId, org_filter)
	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, w)
		return
	}

	if update.ModifiedCount == 0 {
		utils.GetError(errors.New("operation failed"), http.StatusUnprocessableEntity, w)
		return
	}

	utils.GetSuccess("organization settings updated successfully", nil, w)
}

func (oh *OrganizationHandler) ResetTokensAndVersion(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	data := map[string]interface{}{
		"version": "free",
		"tokens":  100,
	}
	_, err := utils.UpdateManyMongoDbDocs(OrganizationCollectionName, make(map[string]interface{}), data)
	if err != nil {
		utils.GetError(errors.New("could not update many docs"), http.StatusExpectationFailed, w)
	}
	utils.GetSuccess("organizations tokens and versions reset successfully", nil, w)
}

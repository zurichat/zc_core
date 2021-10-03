package organizations

import (
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
	"zuri.chat/zccore/service"
	"zuri.chat/zccore/utils"
)

// Get a single member of an organization
func (oh *OrganizationHandler) GetMember(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	orgId := mux.Vars(r)["id"]
	memId := mux.Vars(r)["mem_id"]

	memberIdhex, _ := primitive.ObjectIDFromHex(memId)

	// check that org_id is valid
	err := ValidateOrg(orgId)
	if err != nil {
		utils.GetError(err, http.StatusBadRequest, w)
		return
	}

	orgMember, err := utils.GetMongoDbDoc(MemberCollectionName, bson.M{
		"org_id":  orgId,
		"_id":     memberIdhex,
		"deleted": bson.M{"$ne": true},
	})

	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, w)
		return
	}

	var member Member
	utils.ConvertStructure(orgMember, &member)
	utils.GetSuccess("Member retrieved successfully", orgMember, w)
}

// Get all members of an organization
func (oh *OrganizationHandler) GetMembers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	orgId := mux.Vars(r)["id"]

	// check that org_id is valid
	err := ValidateOrg(orgId)
	if err != nil {
		utils.GetError(err, http.StatusBadRequest, w)
		return
	}

	// query allows you to be able to browse people given the right query param
	query := r.URL.Query().Get("query")

	var filter map[string]interface{}

	filter = bson.M{
		"org_id":  orgId,
		"deleted": bson.M{"$ne": true},
	}

	//set filter based on query presence
	regex := bson.M{"$regex": primitive.Regex{Pattern: query, Options: "i"}}

	if query != "" {
		filter = bson.M{
			"org_id":  orgId,
			"deleted": bson.M{"$ne": true},
			"$or": []bson.M{
				{"first_name": regex},
				{"last_name": regex},
				{"email": query},
				{"display_name": regex},
			},
		}
	}

	orgMembers, err := utils.GetMongoDbDocs(MemberCollectionName, filter)

	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, w)
		return
	}
	utils.GetSuccess("Members retrieved successfully", orgMembers, w)
}

// Add member to an organization
func (oh *OrganizationHandler) CreateMember(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	sOrgId := mux.Vars(r)["id"]

	orgId, err := primitive.ObjectIDFromHex(sOrgId)
	if err != nil {
		utils.GetError(errors.New("invalid id"), http.StatusBadRequest, w)
		return
	}

	// Get data from request json
	if err := utils.ParseJsonFromRequest(r, &RequestData); err != nil {
		utils.GetError(err, http.StatusUnprocessableEntity, w)
		return
	}

	// Validating the user email
	newUserEmailm, ok := RequestData["user_email"]
	newUserEmail := strings.ToLower(newUserEmailm)
	newUserName := strings.Split(newUserEmail, "@")[0]

	if !ok {
		utils.GetError(fmt.Errorf("user_email not provided"), http.StatusBadRequest, w)
		return
	}

	if !utils.IsValidEmail(newUserEmail) {
		utils.GetError(fmt.Errorf("invalid email format : %s", newUserEmail), http.StatusBadRequest, w)
		return
	}

	userDoc, _ := utils.GetMongoDbDoc(UserCollectionName, bson.M{"email": newUserEmail})
	if userDoc == nil {
		fmt.Printf("user with email %s doesn't exist! Register User to Proceed", newUserEmail)
		utils.GetError(errors.New("user with email "+newUserEmail+" doesn't exist! Register User to Proceed"), http.StatusBadRequest, w)
		return
	}
	type GUser struct {
		ID            primitive.ObjectID
		Email         string
		Organizations []string
	}
	// convert user to struct
	var guser GUser
	mapstructure.Decode(userDoc, &guser)

	user, _ := auth.FetchUserByEmail(bson.M{"email": strings.ToLower(newUserEmail)})

	// get organization
	orgDoc, _ := utils.GetMongoDbDoc(OrganizationCollectionName, bson.M{"_id": orgId})
	if orgDoc == nil {
		fmt.Printf("organization with id %s doesn't exist!", orgId.String())
		utils.GetError(errors.New("organization with id "+sOrgId+" doesn't exist!"), http.StatusBadRequest, w)
		return
	}

	// check that member isn't already in the organization
	memDoc, _ := utils.GetMongoDbDocs(MemberCollectionName, bson.M{"org_id": sOrgId, "email": newUserEmail})
	if memDoc != nil {
		fmt.Printf("organization %s has member with email %s!", orgId.String(), newUserEmail)
		utils.GetError(errors.New("user is already in this organization"), http.StatusBadRequest, w)
		return
	}

	// convert org to struct
	var org Organization
	mapstructure.Decode(orgDoc, &org)

	setting := new(Settings)

	newMember := NewMember(user.Email, newUserName, orgId.Hex(), MemberRole, setting)

	coll := utils.GetCollection(MemberCollectionName)

	res, err := coll.InsertOne(r.Context(), newMember)
	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, w)
		return
	}

	// update user organizations collection
	updateFields := make(map[string]interface{})
	user.Organizations = append(user.Organizations, sOrgId)
	updateFields["Organizations"] = user.Organizations
	_, eerr := utils.UpdateOneMongoDbDoc(UserCollectionName, user.ID, updateFields)
	if eerr != nil {
		utils.GetError(errors.New("user update failed"), http.StatusInternalServerError, w)
		return
	}

	// publish update to subscriber
	eventChannel := fmt.Sprintf("organizations_%s", orgId)
	event := utils.Event{Identifier: res.InsertedID, Type: "User", Event: CreateOrganizationMember, Channel: eventChannel, Payload: make(map[string]interface{})}
	go utils.Emitter(event)

	utils.GetSuccess("Member created successfully", utils.M{"member_id": res.InsertedID}, w)
}

// endpoint to update a member's profile picture
func (oh *OrganizationHandler) UpdateProfilePicture(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-Type", "application/json")

	orgId := mux.Vars(r)["id"]
	member_Id := mux.Vars(r)["mem_id"]

	// check that org_id is valid
	err := ValidateOrg(orgId)
	if err != nil {
		utils.GetError(err, http.StatusBadRequest, w)
		return
	}

	// check that member_id is valid
	err = ValidateMember(orgId, member_Id)
	if err != nil {
		utils.GetError(err, http.StatusBadRequest, w)
		return
	}
	if mux.Vars(r)["action"] == "delete" {
		result, err := utils.UpdateOneMongoDbDoc(MemberCollectionName, member_Id, bson.M{"image_url": ""})

		if err != nil {
			utils.GetError(err, http.StatusInternalServerError, w)
			return
		}

		if result.ModifiedCount == 0 {
			utils.GetError(errors.New("operation failed"), http.StatusInternalServerError, w)
			return
		}
		utils.GetSuccess("image deleted successfully", "", w)
	} else {
		uploadPath := "profile_image/" + orgId + "/" + member_Id
		img_url, erro := service.ProfileImageUpload(uploadPath, r)
		if erro != nil {
			utils.GetError(erro, http.StatusInternalServerError, w)
			return
		}

		result, err := utils.UpdateOneMongoDbDoc(MemberCollectionName, member_Id, bson.M{"image_url": img_url})

		if err != nil {
			utils.GetError(err, http.StatusInternalServerError, w)
			return
		}

		if result.ModifiedCount == 0 {
			utils.GetError(errors.New("operation failed"), http.StatusInternalServerError, w)
			return
		}
		utils.GetSuccess("image updated successfully", img_url, w)
	}

	// publish update to subscriber
	eventChannel := fmt.Sprintf("organizations_%s", orgId)
	event := utils.Event{Identifier: member_Id, Type: "User", Event: UpdateOrganizationMemberPic, Channel: eventChannel, Payload: make(map[string]interface{})}
	go utils.Emitter(event)

}

// an endpoint to update a user status
func (oh *OrganizationHandler) UpdateMemberStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")

	// validate the user ID
	orgId := mux.Vars(r)["id"]
	member_Id := mux.Vars(r)["mem_id"]

	// check that org_id is valid
	err := ValidateOrg(orgId)
	if err != nil {
		utils.GetError(err, http.StatusBadRequest, w)
		return
	}

	// check that member_id is valid
	err = ValidateMember(orgId, member_Id)
	if err != nil {
		utils.GetError(err, http.StatusBadRequest, w)
		return
	}

	// Get data from requestbody
	var status Status
	if err := utils.ParseJsonFromRequest(r, &status); err != nil {
		utils.GetError(err, http.StatusUnprocessableEntity, w)
		return
	}

	// check the value in expiry field
	var choosenTime time.Time
	if _, ok := StatusExpiryTime[status.ExpiryTime]; !ok {

		choosenTime, err = time.Parse(time.RFC3339, status.ExpiryTime)

		if err != nil {
			utils.GetError(errors.New("invalid selection of expiry time"), http.StatusBadRequest, w)
			return
		}
	}

	currentTime := time.Now().Local()
	switch set := status.ExpiryTime; set {
	case DontClear:

	case ThirtyMins:
		go ClearStatus(orgId, member_Id, 30)

	case OneHr:
		go ClearStatus(orgId, member_Id, 60)

	case FourHrs:
		go ClearStatus(orgId, member_Id, 240)

	case Today:
		go ClearStatus(orgId, member_Id, 60*int(24-currentTime.Hour()))

	case ThisWeek:
		day := int(time.Now().Weekday())
		weekday := 7 - day
		minutes := weekday * 24 * 60
		go ClearStatus(orgId, member_Id, minutes)

	default:
		diff := choosenTime.Local().Sub(currentTime)
		go ClearStatus(orgId, member_Id, int(diff.Minutes()))
	}

	statusUpdate, err := utils.StructToMap(status)
	if err != nil {
		utils.GetError(err, http.StatusUnprocessableEntity, w)
		return
	}

	memberStatus := make(map[string]interface{})
	memberStatus["status"] = statusUpdate

	// updates member status
	result, err := utils.UpdateOneMongoDbDoc(MemberCollectionName, member_Id, memberStatus)
	if err != nil {
		utils.GetError(err, http.StatusUnprocessableEntity, w)
		return
	}

	if result.ModifiedCount == 0 {
		utils.GetError(errors.New("operation failed"), http.StatusUnprocessableEntity, w)
		return
	}

	// publish update to subscriber
	eventChannel := fmt.Sprintf("organizations_%s", orgId)
	event := utils.Event{Identifier: member_Id, Type: "User", Event: UpdateOrganizationMemberStatus, Channel: eventChannel, Payload: make(map[string]interface{})}
	go utils.Emitter(event)
	utils.GetSuccess("status updated successfully", nil, w)
}

// Delete single member from an organization
func (oh *OrganizationHandler) DeactivateMember(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	orgId, memberId := vars["id"], vars["mem_id"]

	// check that org_id is valid
	err := ValidateOrg(orgId)
	if err != nil {
		utils.GetError(err, http.StatusBadRequest, w)
		return
	}

	// check that member_id is valid
	err = ValidateMember(orgId, memberId)
	if err != nil {
		utils.GetError(err, http.StatusBadRequest, w)
		return
	}

	deleteUpdate := bson.M{"deleted": true, "deleted_at": time.Now()}
	res, err := utils.UpdateOneMongoDbDoc(MemberCollectionName, memberId, deleteUpdate)

	if err != nil {
		utils.GetError(fmt.Errorf("an error occured: %s", err), http.StatusInternalServerError, w)
	}

	if res.ModifiedCount != 1 {
		utils.GetError(errors.New("an error occured, failed to deactivate member"), http.StatusInternalServerError, w)
		return
	}

	// publish update to subscriber
	eventChannel := fmt.Sprintf("organizations_%s", orgId)
	event := utils.Event{Identifier: memberId, Type: "User", Event: DeactivateOrganizationMember, Channel: eventChannel, Payload: make(map[string]interface{})}
	go utils.Emitter(event)

	utils.GetSuccess("successfully deactivated member", nil, w)
}

// Update a member profile
func (oh *OrganizationHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	orgId := mux.Vars(r)["id"]
	memberId := mux.Vars(r)["mem_id"]

	// check that org_id is valid
	err := ValidateOrg(orgId)
	if err != nil {
		utils.GetError(err, http.StatusBadRequest, w)
		return
	}

	// check that member_id is valid
	err = ValidateMember(orgId, memberId)
	if err != nil {
		utils.GetError(err, http.StatusBadRequest, w)
		return
	}

	// Get data from request
	var memberProfile Profile
	err = utils.ParseJsonFromRequest(r, &memberProfile)
	if err != nil {
		utils.GetError(err, http.StatusUnprocessableEntity, w)
		return
	}

	if len(memberProfile.Socials) > 5 {
		utils.GetError(errors.New("number of socials cannot exceed five"), http.StatusBadRequest, w)
		return
	}

	// convert struct to map
	mProfile, err := utils.StructToMap(memberProfile)
	if err != nil {
		utils.GetError(err, http.StatusUnprocessableEntity, w)
		return
	}

	// Fetch and update the MemberDoc from collection
	update, err := utils.UpdateOneMongoDbDoc(MemberCollectionName, memberId, mProfile)
	if err != nil {
		utils.GetError(err, http.StatusUnprocessableEntity, w)
		return
	}

	if update.ModifiedCount == 0 {
		utils.GetError(errors.New("operation failed"), http.StatusUnprocessableEntity, w)
		return
	}

	// publish update to subscriber
	eventChannel := fmt.Sprintf("organizations_%s", orgId)
	event := utils.Event{Identifier: memberId, Type: "User", Event: UpdateOrganizationMemberProfile, Channel: eventChannel, Payload: make(map[string]interface{})}
	go utils.Emitter(event)

	utils.GetSuccess("Member Profile updated succesfully", nil, w)
}

// Toggle a member's presence
func (oh *OrganizationHandler) TogglePresence(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	orgId := mux.Vars(r)["id"]
	memId := mux.Vars(r)["mem_id"]

	// check that org_id is valid
	err := ValidateOrg(orgId)
	if err != nil {
		utils.GetError(err, http.StatusBadRequest, w)
		return
	}

	// Check if member id is valid
	pMemId, err := primitive.ObjectIDFromHex(memId)
	if err != nil {
		utils.GetError(errors.New("invalid id"), http.StatusBadRequest, w)
		return
	}

	memberDoc, _ := utils.GetMongoDbDoc(MemberCollectionName, bson.M{"_id": pMemId, "org_id": orgId})
	if memberDoc == nil {
		fmt.Printf("member with id %s doesn't exist!", memId)
		utils.GetError(errors.New("member with id doesn't exist"), http.StatusBadRequest, w)
		return
	}

	org_filter := make(map[string]interface{})
	if memberDoc["presence"] == "true" {
		org_filter["presence"] = "false"
	} else {
		org_filter["presence"] = "true"
	}

	// update the presence field of the member
	update, err := utils.UpdateOneMongoDbDoc(MemberCollectionName, memId, org_filter)
	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, w)
		return
	}

	if update.ModifiedCount == 0 {
		utils.GetError(errors.New("operation failed"), http.StatusInternalServerError, w)
		return
	}

	// publish update to subscriber
	eventChannel := fmt.Sprintf("organizations_%s", orgId)
	event := utils.Event{Identifier: memId, Type: "User", Event: UpdateOrganizationMemberPresence, Channel: eventChannel, Payload: make(map[string]interface{})}
	go utils.Emitter(event)

	utils.GetSuccess("Member presence toggled", nil, w)
}

func (oh *OrganizationHandler) UpdateMemberSettings(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	orgId, memberId := vars["id"], vars["mem_id"]

	// check that org_id is valid
	err := ValidateOrg(orgId)
	if err != nil {
		utils.GetError(err, http.StatusBadRequest, w)
		return
	}

	// check that member_id is valid
	err = ValidateMember(orgId, memberId)
	if err != nil {
		utils.GetError(err, http.StatusBadRequest, w)
		return
	}

	// Parse request from incoming payload
	var settings Settings
	err = utils.ParseJsonFromRequest(r, &settings)
	if err != nil {
		utils.GetError(err, http.StatusUnprocessableEntity, w)
		return
	}

	// convert setting struct to map
	pSettings, err := utils.StructToMap(settings)
	if err != nil {
		utils.GetError(err, http.StatusUnprocessableEntity, w)
		return
	}

	memberSettings := make(map[string]interface{})
	memberSettings["settings"] = pSettings

	// fetch and update the document
	update, err := utils.UpdateOneMongoDbDoc(MemberCollectionName, memberId, memberSettings)
	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, w)
		return
	}

	if update.ModifiedCount == 0 {
		utils.GetError(errors.New("operation failed"), http.StatusInternalServerError, w)
		return
	}

	// publish update to subscriber
	eventChannel := fmt.Sprintf("organizations_%s", orgId)
	event := utils.Event{Identifier: memberId, Type: "User", Event: UpdateOrganizationMemberSettings, Channel: eventChannel, Payload: make(map[string]interface{})}
	go utils.Emitter(event)

	utils.GetSuccess("Member settings updated successfully", nil, w)
}

// Activate single member in an organization
func (oh *OrganizationHandler) ReactivateMember(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	orgId, memberId := vars["id"], vars["mem_id"]

	// check that org_id is valid
	err := ValidateOrg(orgId)
	if err != nil {
		utils.GetError(err, http.StatusBadRequest, w)
		return
	}

	// Check if member id is valid
	pMemId, err := primitive.ObjectIDFromHex(memberId)
	if err != nil {
		utils.GetError(errors.New("invalid id"), http.StatusBadRequest, w)
		return
	}

	memberDoc, _ := utils.GetMongoDbDoc(MemberCollectionName, bson.M{"_id": pMemId, "org_id": orgId})
	if memberDoc == nil {
		fmt.Printf("member with id %s doesn't exist!", memberId)
		utils.GetError(errors.New("member with id doesn't exist"), http.StatusBadRequest, w)
		return
	}

	if memberDoc["deleted"] == false {
		utils.GetError(errors.New("member is active"), http.StatusBadRequest, w)
		return
	}

	ActivatedMember := bson.M{"deleted": false, "deleted_at": time.Time{}}
	res, err := utils.UpdateOneMongoDbDoc(MemberCollectionName, memberId, ActivatedMember)

	if err != nil {
		utils.GetError(fmt.Errorf("an error occured: %s", err), http.StatusInternalServerError, w)
	}

	if res.ModifiedCount != 1 {
		utils.GetError(errors.New("an error occured, cannot activate user"), http.StatusInternalServerError, w)
		return
	}

	// publish update to subscriber
	eventChannel := fmt.Sprintf("organizations_%s", orgId)
	event := utils.Event{Identifier: memberId, Type: "User", Event: ReactivateOrganizationMember, Channel: eventChannel, Payload: make(map[string]interface{})}
	go utils.Emitter(event)
	utils.GetSuccess("successfully reactivated member", nil, w)
}

// Check the guest status of an email embedded in an invite UUID
func (oh *OrganizationHandler) CheckGuestStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// 0. Extract and validate UUID
	guestUUID := mux.Vars(r)["uuid"]
	_, err := utils.ValidateUUID(guestUUID)
	if err != nil {
		utils.GetError(err, http.StatusBadRequest, w)
		return
	}

	// 1. Query organization invites collection for uuid
	res, err := utils.GetMongoDbDoc(OrganizationInviteCollection, bson.M{"uuid": guestUUID})
	if err != nil {
		utils.GetError(err, http.StatusBadRequest, w)
		return
	}
	// 2. Check if email already is registered in zurichat (return 403 user already exist)
	guestEmail := res["email"]
	_, err = utils.GetMongoDbDoc(UserCollectionName, bson.M{"email": guestEmail})
	if err != nil {
		utils.GetError(
			errors.New("guest status: user does not exist on zurichat"),
			http.StatusNotFound,
			w,
		)
		return
	}
	// 3. If email does not exist, add to
	utils.GetSuccess("guest status: user exist on zurichat", "protected", w)
}

// Add accepted guest as member to organization without requiring admin or workspace owner rights
func (oh *OrganizationHandler) GuestToOrganization(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	gUUID := mux.Vars(r)["uuid"]
	// TODO 1: Validate UUID
	_, err := utils.ValidateUUID(gUUID)
	if err != nil {
		utils.GetError(err, http.StatusBadRequest, w)
		return
	}

	res, err := utils.GetMongoDbDoc(OrganizationInviteCollection, bson.M{"uuid": gUUID})
	if err != nil {
		utils.GetError(err, http.StatusBadRequest, w)
		return
	}

	// // TODO 0: Check that organization exists
	orgID := res["org_id"].(string)
	validOrgID, err := primitive.ObjectIDFromHex(orgID)
	if err != nil {
		utils.GetError(errors.New("invalid id"), http.StatusBadRequest, w)
		return
	}

	orgDoc, _ := utils.GetMongoDbDoc(OrganizationCollectionName, bson.M{"_id": validOrgID})
	if orgDoc == nil {
		utils.GetError(errors.New("organization with id "+orgID+" doesn't exist!"), http.StatusBadRequest, w)
		return
	}

	// TODO 2: Verify guest email
	guestEmail := res["email"]
	if !utils.IsValidEmail(guestEmail.(string)) {
		utils.GetError(errors.New("invalid email address"), http.StatusBadRequest, w)
		return
	}

	// TODO 3: Check that guest is (now) registered on zurichat
	email := guestEmail.(string)
	user, err := auth.FetchUserByEmail(bson.M{"email": email})
	if err != nil {
		utils.GetError(errors.New("user with "+email+" does not exist. register to proceed"), http.StatusBadRequest, w)
		return
	}

	// TODO 4: Check that guest does not already exist (as a member) in organization
	memDoc, err := utils.GetMongoDbDocs(MemberCollectionName, bson.M{"org_id": orgID, "email": user.Email})
	if memDoc != nil && err == nil {
		utils.GetError(errors.New("user is already in this organization"), http.StatusBadRequest, w)
		return
	}

	// TODO 5: Create a member profile for the guest
	setting := new(Settings)
	username := strings.Split(user.Email, "@")[0]

	memberStruct := Member{
		ID:       primitive.NewObjectID(),
		Email:    user.Email,
		UserName: username,
		OrgId:    validOrgID.Hex(),
		Role:     "member",
		Presence: "true",
		JoinedAt: time.Now(),
		Settings: setting,
		Deleted:  false,
	}
	data, err := utils.StructToMap(memberStruct)
	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, w)
		return
	}

	resp, err := utils.CreateMongoDbDoc(MemberCollectionName, data)
	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, w)
		return
	}

	// TODO 6: Add member to organization
	organizationStruct := new(Organization)
	err = mapstructure.Decode(orgDoc, &organizationStruct)
	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, w)
		return
	}

	// update user organizations collection
	updateFields := make(map[string]interface{})
	user.Organizations = append(user.Organizations, validOrgID.Hex())
	updateFields["Organizations"] = user.Organizations
	_, err = utils.UpdateOneMongoDbDoc(UserCollectionName, user.ID, updateFields)
	if err != nil {
		utils.GetError(errors.New("user update failed"), http.StatusInternalServerError, w)
		return
	}

	utils.GetSuccess("Member created successfully", utils.M{"member_id": resp.InsertedID}, w)
}

func (oh *OrganizationHandler) UpdateMemberRole(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	org_Id, memberId := vars["id"], vars["mem_id"]

	err := ValidateOrg(org_Id)
	if err != nil {
		utils.GetError(err, http.StatusBadRequest, w)
		return
	}

	// check that member_id is valid
	err = ValidateMember(org_Id, memberId)
	if err != nil {
		utils.GetError(err, http.StatusBadRequest, w)
		return
	}

	if err := utils.ParseJsonFromRequest(r, &RequestData); err != nil {
		utils.GetError(err, http.StatusUnprocessableEntity, w)
		return
	}

	role := strings.ToLower(RequestData["role"])

	if _, ok := Roles[role]; !ok {
		utils.GetError(errors.New("role is not valid"), http.StatusBadRequest, w)
		return
	}

	memId, _ := primitive.ObjectIDFromHex(memberId)

	orgMember, _ := FetchMember(bson.M{"org_id": org_Id, "_id": memId})

	if orgMember.Role == role {
		errorMessage := fmt.Sprintf("member role is already %s", role)
		utils.GetError(errors.New(errorMessage), http.StatusBadRequest, w)
		return
	}

	// ID of the user whose role is being updated
	memberID := orgMember.ID.Hex()

	updateRes, err := utils.UpdateOneMongoDbDoc(MemberCollectionName, memberID, bson.M{"role": role})

	if err != nil {
		utils.GetError(errors.New("operation failed"), http.StatusInternalServerError, w)
		return
	}

	if updateRes.ModifiedCount == 0 {
		utils.GetError(errors.New("could not update member role"), http.StatusInternalServerError, w)
		return
	}

	// publish update to subscriber
	eventChannel := fmt.Sprintf("organizations_%s", org_Id)
	event := utils.Event{Identifier: memberId, Type: "User", Event: UpdateOrganizationMemberRole, Channel: eventChannel, Payload: make(map[string]interface{})}
	go utils.Emitter(event)

	utils.GetSuccess("member role updated successfully", nil, w)
}

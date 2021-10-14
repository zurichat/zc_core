package organizations

import (
	"errors"
	"fmt"
	"log"
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

// Get a single member of an organization.
func (oh *OrganizationHandler) GetMember(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	orgID := mux.Vars(r)["id"]
	memID := mux.Vars(r)["mem_id"]

	memberIDhex, _ := primitive.ObjectIDFromHex(memID)

	// check that org_id is valid
	err := ValidateOrg(orgID)
	if err != nil {
		utils.GetError(err, http.StatusBadRequest, w)
		return
	}

	orgMember, err := utils.GetMongoDBDoc(MemberCollectionName, bson.M{
		"org_id":  orgID,
		"_id":     memberIDhex,
		"deleted": bson.M{"$ne": true},
	})

	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, w)
		return
	}

	var member Member

	err = utils.ConvertStructure(orgMember, &member)
	if err != nil {

		utils.GetError(err, http.StatusInternalServerError, w)
		return
	}

	utils.GetSuccess("Member retrieved successfully", orgMember, w)
}

// Get all members of an organization.
func (oh *OrganizationHandler) GetMembers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	orgID := mux.Vars(r)["id"]

	// check that org_id is valid
	err := ValidateOrg(orgID)
	if err != nil {
		utils.GetError(err, http.StatusBadRequest, w)
		return
	}

	// query allows you to be able to browse people given the right query param
	query := r.URL.Query().Get("query")

	var filter map[string]interface{}

	filter = bson.M{
		"org_id":  orgID,
		"deleted": bson.M{"$ne": true},
	}

	// set filter based on query presence
	regex := bson.M{"$regex": primitive.Regex{Pattern: query, Options: "i"}}

	if query != "" {
		filter = bson.M{
			"org_id":  orgID,
			"deleted": bson.M{"$ne": true},
			"$or": []bson.M{
				{"first_name": regex},
				{"last_name": regex},
				{"email": query},
				{"display_name": regex},
			},
		}
	}

	orgMembers, err := utils.GetMongoDBDocs(MemberCollectionName, filter)

	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, w)
		return
	}

	utils.GetSuccess("Members retrieved successfully", orgMembers, w)
}

// Add member to an organization.
func (oh *OrganizationHandler) CreateMember(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	sOrgID := mux.Vars(r)["id"]

	orgID, err := primitive.ObjectIDFromHex(sOrgID)
	if err != nil {
		utils.GetError(errors.New("invalid id"), http.StatusBadRequest, w)
		return
	}

	// Get data from request json
	if err = utils.ParseJSONFromRequest(r, &RequestData); err != nil {
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

	userDoc, _ := utils.GetMongoDBDoc(UserCollectionName, bson.M{"email": newUserEmail})
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

	err = mapstructure.Decode(userDoc, &guser)
	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, w)
		return
	}

	user, _ := auth.FetchUserByEmail(bson.M{"email": strings.ToLower(newUserEmail)})

	// get organization
	orgDoc, _ := utils.GetMongoDBDoc(OrganizationCollectionName, bson.M{"_id": orgID})
	if orgDoc == nil {
		fmt.Printf("organization with id %s doesn't exist!", orgID.String())
		utils.GetError(errors.New("organization with id "+sOrgID+" doesn't exist!"), http.StatusBadRequest, w)

		return
	}

	// check that member isn't already in the organization
	memDoc, _ := utils.GetMongoDBDocs(MemberCollectionName, bson.M{"org_id": sOrgID, "email": newUserEmail})
	if memDoc != nil {
		fmt.Printf("organization %s has member with email %s!", orgID.String(), newUserEmail)
		utils.GetError(errors.New("user is already in this organization"), http.StatusBadRequest, w)

		return
	}

	// convert org to struct
	var org Organization

	err = mapstructure.Decode(orgDoc, &org)
	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, w)
		return
	}

	newMember := NewMember(user.Email, newUserName, orgID.Hex(), MemberRole)

	coll := utils.GetCollection(MemberCollectionName)

	res, err := coll.InsertOne(r.Context(), newMember)
	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, w)
		return
	}

	// update user organizations collection
	updateFields := make(map[string]interface{})

	user.Organizations = append(user.Organizations, sOrgID)

	updateFields["Organizations"] = user.Organizations

	_, eerr := utils.UpdateOneMongoDBDoc(UserCollectionName, user.ID, updateFields)
	if eerr != nil {
		utils.GetError(errors.New("user update failed"), http.StatusInternalServerError, w)
		return
	}

	// publish update to subscriber
	eventChannel := fmt.Sprintf("organizations_%s", orgID)
	event := utils.Event{Identifier: res.InsertedID, Type: "User", Event: CreateOrganizationMember, Channel: eventChannel, Payload: make(map[string]interface{})}

	go utils.Emitter(event)

	utils.GetSuccess("Member created successfully", utils.M{"member_id": res.InsertedID}, w)

	enterOrgMessage := EnterLeaveMessage{
		OrganizationID: sOrgID,
		MemberID:       res.InsertedID.(primitive.ObjectID).Hex(),
	}
	eee := AddSyncMessage(sOrgID, "enter_organization", enterOrgMessage)
	if eee != nil {
		log.Printf("sync error: %v", eee)
	}
}

// endpoint to update a member's profile picture.
func (oh *OrganizationHandler) UpdateProfilePicture(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-Type", "application/json")

	orgID := mux.Vars(r)["id"]
	memberID := mux.Vars(r)["mem_id"]

	// check that org_id is valid
	err := ValidateOrg(orgID)
	if err != nil {
		utils.GetError(err, http.StatusBadRequest, w)
		return
	}

	// check that member_id is valid
	err = ValidateMember(orgID, memberID)
	if err != nil {
		utils.GetError(err, http.StatusBadRequest, w)
		return
	}

	if mux.Vars(r)["action"] == "delete" {
		result, err := utils.UpdateOneMongoDBDoc(MemberCollectionName, memberID, bson.M{"image_url": ""})

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
		uploadPath := "profile_image/" + orgID + "/" + memberID
		imgURL, err := service.ProfileImageUpload(uploadPath, r)
		if err != nil {
			utils.GetError(err, http.StatusInternalServerError, w)
			return
		}

		result, err := utils.UpdateOneMongoDBDoc(MemberCollectionName, memberID, bson.M{"image_url": imgURL})

		if err != nil {
			utils.GetError(err, http.StatusInternalServerError, w)
			return
		}

		if result.ModifiedCount == 0 {
			utils.GetError(errors.New("operation failed"), http.StatusInternalServerError, w)
			return
		}

		utils.GetSuccess("image updated successfully", imgURL, w)
	}

	// publish update to subscriber
	eventChannel := fmt.Sprintf("organizations_%s", orgID)
	event := utils.Event{Identifier: memberID, Type: "User", Event: UpdateOrganizationMemberPic, Channel: eventChannel, Payload: make(map[string]interface{})}

	go utils.Emitter(event)
}

// an endpoint to update a user status.
func (oh *OrganizationHandler) UpdateMemberStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")

	// validate the user ID
	orgID := mux.Vars(r)["id"]
	memberID := mux.Vars(r)["mem_id"]

	// check that org_id is valid
	err := ValidateOrg(orgID)
	if err != nil {
		utils.GetError(err, http.StatusBadRequest, w)
		return
	}

	// check that member_id is valid
	err = ValidateMember(orgID, memberID)
	if err != nil {
		utils.GetError(err, http.StatusBadRequest, w)
		return
	}

	// Get data from requestbody
	var status Status
	if err = utils.ParseJSONFromRequest(r, &status); err != nil {
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
		period := 1
		go ClearStatusRoutine(orgID, memberID, period)

	case OneHr:
		period := 60
		go ClearStatusRoutine(orgID, memberID, period)

	case FourHrs:
		period := 240
		go ClearStatusRoutine(orgID, memberID, period)

	case Today:
		minutesPerHr := 60
		hrsPerDay := 24
		period := minutesPerHr * (hrsPerDay - currentTime.Hour())

		go ClearStatusRoutine(orgID, memberID, period)

	case ThisWeek:
		minutesPerHr := 60
		hrsPerDay := 24
		daysPerWeek := 7

		day := int(time.Now().Weekday())
		weekday := daysPerWeek - day

		period := weekday * hrsPerDay * minutesPerHr

		go ClearStatusRoutine(orgID, memberID, period)

	default:
		diff := choosenTime.Local().Sub(currentTime)
		go ClearStatusRoutine(orgID, memberID, int(diff.Minutes()))
	}

	pmemberID, err := primitive.ObjectIDFromHex(memberID)
	if err != nil {
		utils.GetError(errors.New("invalid id"), http.StatusBadRequest, w)
		return
	}

	memberRec, err := utils.GetMongoDBDoc(MemberCollectionName, bson.M{"_id": pmemberID})
	if err != nil {
		utils.GetError(err, http.StatusUnprocessableEntity, w)
		return
	}

	var prevStatus Status

	// convert bson to struct
	bsonBytes, _ := bson.Marshal(memberRec["status"])

	if err = bson.Unmarshal(bsonBytes, &prevStatus); err != nil {
		utils.GetError(err, http.StatusUnprocessableEntity, w)
		return
	}

	newHistory := StatusHistory{
		TagHistory:    status.Tag,
		TextHistory:   status.Text,
		ExpiryHistory: status.ExpiryTime,
	}

	status.StatusHistory = append(prevStatus.StatusHistory, newHistory)
	if len(status.StatusHistory) > 6 {
		status.StatusHistory = status.StatusHistory[1:]
	}

	statusUpdate, err := utils.StructToMap(status)
	if err != nil {
		utils.GetError(err, http.StatusUnprocessableEntity, w)
		return
	}

	memberStatus := make(map[string]interface{})
	memberStatus["status"] = statusUpdate

	// updates member status
	result, err := utils.UpdateOneMongoDBDoc(MemberCollectionName, memberID, memberStatus)
	if err != nil {
		utils.GetError(err, http.StatusUnprocessableEntity, w)
		return
	}

	if result.ModifiedCount == 0 {
		utils.GetError(errors.New("operation failed"), http.StatusUnprocessableEntity, w)
		return
	}

	// publish update to subscriber
	eventChannel := fmt.Sprintf("organizations_%s", orgID)
	event := utils.Event{Identifier: memberID, Type: "User", Event: UpdateOrganizationMemberStatus, Channel: eventChannel, Payload: make(map[string]interface{})}

	go utils.Emitter(event)

	utils.GetSuccess("status updated successfully", nil, w)
}

// Delete single member from an organization.
func (oh *OrganizationHandler) DeactivateMember(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	orgID, memberID := vars["id"], vars["mem_id"]

	// check that org_id is valid
	err := ValidateOrg(orgID)
	if err != nil {
		utils.GetError(err, http.StatusBadRequest, w)
		return
	}

	// check that member_id is valid
	err = ValidateMember(orgID, memberID)
	if err != nil {
		utils.GetError(err, http.StatusBadRequest, w)
		return
	}

	deleteUpdate := bson.M{"deleted": true, "deleted_at": time.Now()}
	res, err := utils.UpdateOneMongoDBDoc(MemberCollectionName, memberID, deleteUpdate)

	if err != nil {
		utils.GetError(fmt.Errorf("an error occurred: %s", err), http.StatusInternalServerError, w)
	}

	if res.ModifiedCount != 1 {
		utils.GetError(errors.New("an error occurred, failed to deactivate member"), http.StatusInternalServerError, w)
		return
	}

	// publish update to subscriber
	eventChannel := fmt.Sprintf("organizations_%s", orgID)
	event := utils.Event{Identifier: memberID, Type: "User", Event: DeactivateOrganizationMember, Channel: eventChannel, Payload: make(map[string]interface{})}

	go utils.Emitter(event)

	utils.GetSuccess("successfully deactivated member", nil, w)
	
	enterOrgMessage := EnterLeaveMessage{
		OrganizationID: orgID,
		MemberID:       memberID,
	}
	eee := AddSyncMessage(orgID, "leave_organization", enterOrgMessage)
	if eee != nil {
		log.Printf("sync error: %v", eee)
	}
}

// Update a member profile.
func (oh *OrganizationHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	orgID := mux.Vars(r)["id"]
	memberID := mux.Vars(r)["mem_id"]

	// check that org_id is valid
	err := ValidateOrg(orgID)
	if err != nil {
		utils.GetError(err, http.StatusBadRequest, w)
		return
	}

	// check that member_id is valid
	err = ValidateMember(orgID, memberID)
	if err != nil {
		utils.GetError(err, http.StatusBadRequest, w)
		return
	}

	// Get data from request
	var memberProfile Profile

	err = utils.ParseJSONFromRequest(r, &memberProfile)
	if err != nil {
		utils.GetError(err, http.StatusUnprocessableEntity, w)
		return
	}

	maxSocials := 5
	if len(memberProfile.Socials) > maxSocials {
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
	update, err := utils.UpdateOneMongoDBDoc(MemberCollectionName, memberID, mProfile)
	if err != nil {
		utils.GetError(err, http.StatusUnprocessableEntity, w)
		return
	}

	if update.ModifiedCount == 0 {
		utils.GetError(errors.New("operation failed"), http.StatusUnprocessableEntity, w)
		return
	}

	// publish update to subscriber
	eventChannel := fmt.Sprintf("organizations_%s", orgID)
	event := utils.Event{Identifier: memberID, Type: "User", Event: UpdateOrganizationMemberProfile, Channel: eventChannel, Payload: make(map[string]interface{})}

	go utils.Emitter(event)

	utils.GetSuccess("Member Profile updated successfully", nil, w)
}

// Toggle a member's presence.
func (oh *OrganizationHandler) TogglePresence(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	orgID := mux.Vars(r)["id"]
	memID := mux.Vars(r)["mem_id"]

	// check that org_id is valid
	err := ValidateOrg(orgID)
	if err != nil {
		utils.GetError(err, http.StatusBadRequest, w)
		return
	}

	// Check if member id is valid
	pMemID, err := primitive.ObjectIDFromHex(memID)
	if err != nil {
		utils.GetError(errors.New("invalid id"), http.StatusBadRequest, w)
		return
	}

	memberDoc, _ := utils.GetMongoDBDoc(MemberCollectionName, bson.M{"_id": pMemID, "org_id": orgID})
	if memberDoc == nil {
		fmt.Printf("member with id %s doesn't exist!", memID)
		utils.GetError(errors.New("member with id doesn't exist"), http.StatusBadRequest, w)

		return
	}

	orgFilter := make(map[string]interface{})

	if memberDoc["presence"] == "true" {
		orgFilter["presence"] = "false"
	} else {
		orgFilter["presence"] = "true"
	}

	// update the presence field of the member
	update, err := utils.UpdateOneMongoDBDoc(MemberCollectionName, memID, orgFilter)
	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, w)
		return
	}

	if update.ModifiedCount == 0 {
		utils.GetError(errors.New("operation failed"), http.StatusInternalServerError, w)
		return
	}

	// publish update to subscriber
	eventChannel := fmt.Sprintf("organizations_%s", orgID)
	event := utils.Event{Identifier: memID, Type: "User", Event: UpdateOrganizationMemberPresence, Channel: eventChannel, Payload: make(map[string]interface{})}

	go utils.Emitter(event)

	utils.GetSuccess("Member presence toggled", nil, w)
}

func (oh *OrganizationHandler) UpdateMemberSettings(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	orgID, memberID := vars["id"], vars["mem_id"]

	// check that org_id is valid
	err := ValidateOrg(orgID)
	if err != nil {
		utils.GetError(err, http.StatusBadRequest, w)
		return
	}

	// check that member_id is valid
	err = ValidateMember(orgID, memberID)
	if err != nil {
		utils.GetError(err, http.StatusBadRequest, w)
		return
	}

	// Parse request from incoming payload
	var settings Settings

	err = utils.ParseJSONFromRequest(r, &settings)
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
	update, err := utils.UpdateOneMongoDBDoc(MemberCollectionName, memberID, memberSettings)
	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, w)
		return
	}

	if update.ModifiedCount == 0 {
		utils.GetError(errors.New("operation failed"), http.StatusInternalServerError, w)
		return
	}

	// publish update to subscriber
	eventChannel := fmt.Sprintf("organizations_%s", orgID)
	event := utils.Event{Identifier: memberID, Type: "User", Event: UpdateOrganizationMemberSettings, Channel: eventChannel, Payload: make(map[string]interface{})}

	go utils.Emitter(event)

	utils.GetSuccess("Member settings updated successfully", nil, w)
}

func (oh *OrganizationHandler) UpdateMemberMessageAndMediaSettings(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	orgID, memberID := vars["id"], vars["mem_id"]

	// check that org_id is valid
	err := ValidateOrg(orgID)
	if err != nil {
		utils.GetError(err, http.StatusBadRequest, w)
		return
	}

	// check that member_id is valid
	err = ValidateMember(orgID, memberID)
	if err != nil {
		utils.GetError(err, http.StatusBadRequest, w)
		return
	}

	// Parse request from incoming payload
	var messageAndMediaSettings MessagesAndMedia

	err = utils.ParseJSONFromRequest(r, &messageAndMediaSettings)
	if err != nil {
		utils.GetError(err, http.StatusUnprocessableEntity, w)
		return
	}

	if _, ok := MsgMedias[messageAndMediaSettings.Theme]; !ok {
		utils.GetError(errors.New("theme is not valid"), http.StatusBadRequest, w)
		return
	}

	if _, ok := MsgMedias[messageAndMediaSettings.Names]; !ok {
		utils.GetError(errors.New("name is not valid"), http.StatusBadRequest, w)
		return
	}

	if _, ok := MsgMedias[messageAndMediaSettings.Emoji]; !ok {
		utils.GetError(errors.New("emoji is not valid"), http.StatusBadRequest, w)
		return
	}

	// convert setting struct to map
	pMessageAndMediaSettings, err := utils.StructToMap(messageAndMediaSettings)
	if err != nil {
		utils.GetError(err, http.StatusUnprocessableEntity, w)
		return
	}

	memberpMessageAndMediaSettings := make(map[string]interface{})
	memberpMessageAndMediaSettings["settings.messages_and_media"] = pMessageAndMediaSettings

	// fetch and update the document
	update, err := utils.UpdateOneMongoDBDoc(MemberCollectionName, memberID, memberpMessageAndMediaSettings)
	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, w)
		return
	}

	if update.ModifiedCount == 0 {
		utils.GetError(errors.New("operation failed"), http.StatusInternalServerError, w)
		return
	}

	// publish update to subscriber
	eventChannel := fmt.Sprintf("organizations_%s", orgID)
	event := utils.Event{Identifier: memberID, Type: "User", Event: UpdateOrganizationMemberSettings, Channel: eventChannel, Payload: make(map[string]interface{})}

	go utils.Emitter(event)

	utils.GetSuccess("Member settings updated successfully", nil, w)
}

func (oh *OrganizationHandler) UpdateMemberAccessibilitySettings(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	orgID, memberID := vars["id"], vars["mem_id"]

	// check that org_id is valid
	err := ValidateOrg(orgID)
	if err != nil {
		utils.GetError(err, http.StatusBadRequest, w)
		return
	}

	// check that member_id is valid
	err = ValidateMember(orgID, memberID)
	if err != nil {
		utils.GetError(err, http.StatusBadRequest, w)
		return
	}

	// Parse request from incoming payload
	var accessibilitySettings Accessibility

	err = utils.ParseJSONFromRequest(r, &accessibilitySettings)
	if err != nil {
		utils.GetError(err, http.StatusUnprocessableEntity, w)
		return
	}

	if _, ok := EmptyMessageFields[accessibilitySettings.PressEmptyMessageField]; !ok {
		utils.GetError(errors.New("invalid field"), http.StatusBadRequest, w)
		return
	}

	// convert setting struct to map
	pAccessibilitySettings, err := utils.StructToMap(accessibilitySettings)
	if err != nil {
		utils.GetError(err, http.StatusUnprocessableEntity, w)
		return
	}

	memberpAccessibilitySettings := make(map[string]interface{})
	memberpAccessibilitySettings["settings.accessibility"] = pAccessibilitySettings

	// fetch and update the document
	update, err := utils.UpdateOneMongoDBDoc(MemberCollectionName, memberID, memberpAccessibilitySettings)
	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, w)
		return
	}

	if update.ModifiedCount == 0 {
		utils.GetError(errors.New("operation failed"), http.StatusInternalServerError, w)
		return
	}

	// publish update to subscriber
	eventChannel := fmt.Sprintf("organizations_%s", orgID)
	event := utils.Event{Identifier: memberID, Type: "User", Event: UpdateOrganizationMemberSettings, Channel: eventChannel, Payload: make(map[string]interface{})}

	go utils.Emitter(event)

	utils.GetSuccess("Member settings updated successfully", nil, w)
}

// an endpoint to update a user advanced  preference.
func (oh *OrganizationHandler) UpdateMemberAdvancedSettings(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	orgID, memberID := vars["id"], vars["mem_id"]

	// check that org_id is valid
	err := ValidateOrg(orgID)
	if err != nil {
		utils.GetError(err, http.StatusBadRequest, w)
		return
	}

	// check that member_id is valid
	err = ValidateMember(orgID, memberID)
	if err != nil {
		utils.GetError(err, http.StatusBadRequest, w)
		return
	}

	// Parse request from incoming payload
	var advancedSettings Advanced

	err = utils.ParseJSONFromRequest(r, &advancedSettings)
	if err != nil {
		utils.GetError(err, http.StatusUnprocessableEntity, w)
		return
	}

	if _, ok := EnterActions[advancedSettings.PressEnterTo]; !ok {
		utils.GetError(errors.New("invalid enter action"), http.StatusBadRequest, w)
		return
	}

	// convert setting struct to map
	pAdvancedSettings, err := utils.StructToMap(advancedSettings)
	if err != nil {
		utils.GetError(err, http.StatusUnprocessableEntity, w)
		return
	}

	memberpAdvancedSettings := make(map[string]interface{})
	memberpAdvancedSettings["settings.advanced"] = pAdvancedSettings

	// fetch and update the document
	update, err := utils.UpdateOneMongoDBDoc(MemberCollectionName, memberID, memberpAdvancedSettings)
	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, w)
		return
	}

	if update.ModifiedCount == 0 {
		utils.GetError(errors.New("operation failed"), http.StatusInternalServerError, w)
		return
	}

	// publish update to subscriber
	eventChannel := fmt.Sprintf("organizations_%s", orgID)
	event := utils.Event{Identifier: memberID, Type: "User", Event: UpdateOrganizationMemberSettings, Channel: eventChannel, Payload: make(map[string]interface{})}

	go utils.Emitter(event)

	utils.GetSuccess("Member settings updated successfully", nil, w)
}

// Activate single member in an organization.
func (oh *OrganizationHandler) ReactivateMember(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	orgID, memberID := vars["id"], vars["mem_id"]

	// check that org_id is valid
	err := ValidateOrg(orgID)
	if err != nil {
		utils.GetError(err, http.StatusBadRequest, w)
		return
	}

	// Check if member id is valid
	pMemID, err := primitive.ObjectIDFromHex(memberID)
	if err != nil {
		utils.GetError(errors.New("invalid id"), http.StatusBadRequest, w)
		return
	}

	memberDoc, _ := utils.GetMongoDBDoc(MemberCollectionName, bson.M{"_id": pMemID, "org_id": orgID})
	if memberDoc == nil {
		fmt.Printf("member with id %s doesn't exist!", memberID)
		utils.GetError(errors.New("member with id doesn't exist"), http.StatusBadRequest, w)

		return
	}

	if memberDoc["deleted"] == false {
		utils.GetError(errors.New("member is active"), http.StatusBadRequest, w)
		return
	}

	ActivatedMember := bson.M{"deleted": false, "deleted_at": time.Time{}}
	res, err := utils.UpdateOneMongoDBDoc(MemberCollectionName, memberID, ActivatedMember)

	if err != nil {
		utils.GetError(fmt.Errorf("an error occurred: %s", err), http.StatusInternalServerError, w)
	}

	if res.ModifiedCount != 1 {
		utils.GetError(errors.New("an error occurred, cannot activate user"), http.StatusInternalServerError, w)
		return
	}

	// publish update to subscriber
	eventChannel := fmt.Sprintf("organizations_%s", orgID)
	event := utils.Event{Identifier: memberID, Type: "User", Event: ReactivateOrganizationMember, Channel: eventChannel, Payload: make(map[string]interface{})}

	go utils.Emitter(event)

	utils.GetSuccess("successfully reactivated member", nil, w)
}

// Check the guest status of an email embedded in an invite UUID.
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
	res, err := utils.GetMongoDBDoc(OrganizationInviteCollection, bson.M{"uuid": guestUUID})
	if err != nil {
		utils.GetError(err, http.StatusBadRequest, w)
		return
	}

	// 2. Check if email already is registered in zurichat (return 403 user already exist)
	guestEmail := res["email"]
	_, err = utils.GetMongoDBDoc(UserCollectionName, bson.M{"email": guestEmail})

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

// Add accepted guest as member to organization without requiring admin or workspace owner rights.
func (oh *OrganizationHandler) GuestToOrganization(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	gUUID := mux.Vars(r)["uuid"]
	// TODO 1: Validate UUID
	_, err := utils.ValidateUUID(gUUID)
	if err != nil {
		utils.GetError(err, http.StatusBadRequest, w)
		return
	}

	res, err := utils.GetMongoDBDoc(OrganizationInviteCollection, bson.M{"uuid": gUUID})
	if err != nil {
		utils.GetError(err, http.StatusBadRequest, w)
		return
	}

	// // TODO 0: Check that organization exists
	orgID, ok := res["org_id"].(string)
	if !ok {
		utils.GetError(errors.New("invalid email address"), http.StatusBadRequest, w)
		return
	}

	validOrgID, err := primitive.ObjectIDFromHex(orgID)

	if err != nil {
		utils.GetError(errors.New("invalid id"), http.StatusBadRequest, w)
		return
	}

	orgDoc, _ := utils.GetMongoDBDoc(OrganizationCollectionName, bson.M{"_id": validOrgID})
	if orgDoc == nil {
		utils.GetError(errors.New("organization with id "+orgID+" doesn't exist!"), http.StatusBadRequest, w)
		return
	}

	email, ok := res["email"].(string)
	if !ok {
		utils.GetError(errors.New("invalid email address"), http.StatusBadRequest, w)
		return
	}

	// TODO 2: Verify guest email
	if !utils.IsValidEmail(email) {
		utils.GetError(errors.New("invalid email address"), http.StatusBadRequest, w)
		return
	}

	// TODO 3: Check that guest is (now) registered on zurichat
	user, err := auth.FetchUserByEmail(bson.M{"email": email})
	if err != nil {
		utils.GetError(errors.New("user with "+email+" does not exist. register to proceed"), http.StatusBadRequest, w)
		return
	}

	// TODO 4: Check that guest does not already exist (as a member) in organization
	memDoc, err := utils.GetMongoDBDocs(MemberCollectionName, bson.M{"org_id": orgID, "email": user.Email})
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
		OrgID:    validOrgID.Hex(),
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

	resp, err := utils.CreateMongoDBDoc(MemberCollectionName, data)
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
	_, err = utils.UpdateOneMongoDBDoc(UserCollectionName, user.ID, updateFields)

	if err != nil {
		utils.GetError(errors.New("user update failed"), http.StatusInternalServerError, w)
		return
	}

	utils.GetSuccess("Member created successfully", utils.M{"member_id": resp.InsertedID}, w)
}

func (oh *OrganizationHandler) UpdateMemberRole(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	orgID, memberID := vars["id"], vars["mem_id"]

	err := ValidateOrg(orgID)
	if err != nil {
		utils.GetError(err, http.StatusBadRequest, w)
		return
	}

	// check that member_id is valid
	err = ValidateMember(orgID, memberID)
	if err != nil {
		utils.GetError(err, http.StatusBadRequest, w)
		return
	}

	if err = utils.ParseJSONFromRequest(r, &RequestData); err != nil {
		utils.GetError(err, http.StatusUnprocessableEntity, w)
		return
	}

	role := strings.ToLower(RequestData["role"])

	if _, ok := Roles[role]; !ok {
		utils.GetError(errors.New("role is not valid"), http.StatusBadRequest, w)
		return
	}

	memID, _ := primitive.ObjectIDFromHex(memberID)

	orgMember, _ := FetchMember(bson.M{"org_id": orgID, "_id": memID})

	if orgMember.Role == role {
		errorMessage := fmt.Sprintf("member role is already %s", role)
		utils.GetError(errors.New(errorMessage), http.StatusBadRequest, w)

		return
	}

	// ID of the user whose role is being updated
	memberIDHex := orgMember.ID.Hex()

	updateRes, err := utils.UpdateOneMongoDBDoc(MemberCollectionName, memberIDHex, bson.M{"role": role})

	if err != nil {
		utils.GetError(errors.New("operation failed"), http.StatusInternalServerError, w)
		return
	}

	if updateRes.ModifiedCount == 0 {
		utils.GetError(errors.New("could not update member role"), http.StatusInternalServerError, w)
		return
	}

	// publish update to subscriber
	eventChannel := fmt.Sprintf("organizations_%s", orgID)
	event := utils.Event{Identifier: memberID, Type: "User", Event: UpdateOrganizationMemberRole, Channel: eventChannel, Payload: make(map[string]interface{})}

	go utils.Emitter(event)

	utils.GetSuccess("member role updated successfully", nil, w)
}

// an endpoint to update a user notification preference.
func (oh *OrganizationHandler) UpdateNotification(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")

	// validate the user ID
	orgID := mux.Vars(r)["id"]
	memberID := mux.Vars(r)["mem_id"]

	// check that org_id is valid
	err := ValidateOrg(orgID)
	if err != nil {
		utils.GetError(err, http.StatusBadRequest, w)
		return
	}

	// check that member_id is valid
	err = ValidateMember(orgID, memberID)
	if err != nil {
		utils.GetError(err, http.StatusBadRequest, w)
		return
	}

	// Get data from requestbody
	var notifications Notifications

	if err = utils.ParseJSONFromRequest(r, &notifications); err != nil {
		utils.GetError(err, http.StatusUnprocessableEntity, w)
		return
	}

	memberSettings := make(map[string]interface{})
	memberSettings["settings.notifications"] = notifications
	// fetch and update the document
	update, err := utils.UpdateOneMongoDBDoc(MemberCollectionName, memberID, memberSettings)
	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, w)
		return
	}

	if update.ModifiedCount == 0 {
		utils.GetError(errors.New("operation failed"), http.StatusInternalServerError, w)
		return
	}

	utils.GetSuccess("successfully updated status", nil, w)
}

//endpoints to set messages as mark as read.
func (oh *OrganizationHandler) UpdateMarkAsRead(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")

	// validate the user ID
	orgID := mux.Vars(r)["id"]
	memberID := mux.Vars(r)["mem_id"]

	// check that org_id is valid
	err := ValidateOrg(orgID)
	if err != nil {
		utils.GetError(err, http.StatusBadRequest, w)
		return
	}

	// check that member_id is valid
	err = ValidateMember(orgID, memberID)
	if err != nil {
		utils.GetError(err, http.StatusBadRequest, w)
		return
	}

	// Get data from requestbody
	var markAsRead MarkAsRead
	if err = utils.ParseJSONFromRequest(r, &markAsRead); err != nil {
		utils.GetError(err, http.StatusUnprocessableEntity, w)
		return
	}

	memberSettings := make(map[string]interface{})
	memberSettings["settings.mark_as_read"] = markAsRead
	// fetch and update the document
	update, err := utils.UpdateOneMongoDBDoc(MemberCollectionName, memberID, memberSettings)
	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, w)
		return
	}

	if update.ModifiedCount == 0 {
		utils.GetError(errors.New("operation failed"), http.StatusInternalServerError, w)
		return
	}

	utils.GetSuccess("successfully updated settings", nil, w)
}

// endpoint to set languages and regions settings.
func (oh *OrganizationHandler) SetLanguagesAndRegions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")

	// validate the user ID
	orgID := mux.Vars(r)["id"]
	memberID := mux.Vars(r)["mem_id"]

	// check that org_id is valid
	err := ValidateOrg(orgID)
	if err != nil {
		utils.GetError(err, http.StatusBadRequest, w)
		return
	}

	// check that member_id is valid
	err = ValidateMember(orgID, memberID)
	if err != nil {
		utils.GetError(err, http.StatusBadRequest, w)
		return
	}

	// Get data from requestbody
	var languagesAndRegions LanguagesAndRegions
	if err = utils.ParseJSONFromRequest(r, &languagesAndRegions); err != nil {
		utils.GetError(err, http.StatusUnprocessableEntity, w)
		return
	}

	memberSettings := make(map[string]interface{})
	memberSettings["settings.languages_and_regions"] = languagesAndRegions

	// fetch and update the document
	update, err := utils.UpdateOneMongoDBDoc(MemberCollectionName, memberID, memberSettings)
	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, w)
		return
	}

	if update.ModifiedCount == 0 {
		utils.GetError(errors.New("operation failed"), http.StatusInternalServerError, w)
		return
	}

	utils.GetSuccess("successfully updated settings", nil, w)
}

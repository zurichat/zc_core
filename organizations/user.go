package organizations

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gofrs/uuid"
	"github.com/gorilla/mux"
	"github.com/mitchellh/mapstructure"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"zuri.chat/zccore/auth"
	"zuri.chat/zccore/service"
	"zuri.chat/zccore/utils"
)

// Get a single member of an organization
func GetMember(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	member_collection, org_collection := "members", "organizations"
	orgId := mux.Vars(r)["id"]
	memId := mux.Vars(r)["mem_id"]

	memberIdhex, _ := primitive.ObjectIDFromHex(memId)

	pOrgId, err := primitive.ObjectIDFromHex(orgId)
	if err != nil {
		utils.GetError(errors.New("invalid Organisation id"), http.StatusBadRequest, w)
		return
	}

	// get organization
	orgDoc, _ := utils.GetMongoDbDoc(org_collection, bson.M{"_id": pOrgId})
	if orgDoc == nil {
		fmt.Printf("org with id %s doesn't exist!", orgId)
		utils.GetError(fmt.Errorf("org with id %s doesn't exist", orgId), http.StatusBadRequest, w)
		return
	}

	orgMember, err := utils.GetMongoDbDoc(member_collection, bson.M{
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
func GetMembers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	member_collection, org_collection := "members", "organizations"
	orgId := mux.Vars(r)["id"]

	pOrgId, err := primitive.ObjectIDFromHex(orgId)
	if err != nil {
		utils.GetError(errors.New("invalid id"), http.StatusBadRequest, w)
		return
	}

	// get organization
	orgDoc, _ := utils.GetMongoDbDoc(org_collection, bson.M{"_id": pOrgId})
	if orgDoc == nil {
		fmt.Printf("org with id %s doesn't exist!", orgId)
		utils.GetError(errors.New("operation failed"), http.StatusBadRequest, w)
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
	if query != "" {
		filter = bson.M{
			"org_id":  orgId,
			"deleted": bson.M{"$ne": true},
			"$or": []bson.M{
				{"first_name": query},
				{"last_name": query},
				{"email": query},
				{"display_name": query},
			},
		}
	}

	orgMembers, err := utils.GetMongoDbDocs(member_collection, filter)

	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, w)
		return
	}
	utils.GetSuccess("Members retrieved successfully", orgMembers, w)
}

// Add member to an organization
func CreateMember(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	loggedInUser := r.Context().Value("user").(*auth.AuthUser)
	org_collection, user_collection, member_collection := "organizations", "users", "members"
	user, _ := auth.FetchUserByEmail(bson.M{"email": strings.ToLower(loggedInUser.Email)})

	sOrgId := mux.Vars(r)["id"]

	if !auth.IsAuthorized(user.ID, sOrgId, "admin", w) {
		return
	}

	orgId, err := primitive.ObjectIDFromHex(sOrgId)
	if err != nil {
		utils.GetError(errors.New("invalid id"), http.StatusBadRequest, w)
		return
	}

	// confirm if user_id exists
	requestData := make(map[string]string)
	if err := utils.ParseJsonFromRequest(r, &requestData); err != nil {
		utils.GetError(err, http.StatusUnprocessableEntity, w)
		return
	}

	// Validating the user email
	newUserEmailm, ok := requestData["user_email"]
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

	userDoc, _ := utils.GetMongoDbDoc(user_collection, bson.M{"email": newUserEmail})
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

	user, _ = auth.FetchUserByEmail(bson.M{"email": strings.ToLower(newUserEmail)})

	// get organization
	orgDoc, _ := utils.GetMongoDbDoc(org_collection, bson.M{"_id": orgId})
	if orgDoc == nil {
		fmt.Printf("organization with id %s doesn't exist!", orgId.String())
		utils.GetError(errors.New("organization with id "+sOrgId+" doesn't exist!"), http.StatusBadRequest, w)
		return
	}

	// check that member isn't already in the organization
	memDoc, _ := utils.GetMongoDbDocs(member_collection, bson.M{"org_id": sOrgId, "email": newUserEmail})
	if memDoc != nil {
		fmt.Printf("organization %s has member with email %s!", orgId.String(), newUserEmail)
		utils.GetError(errors.New("user is already in this organization"), http.StatusBadRequest, w)
		return
	}

	// convert org to struct
	var org Organization
	mapstructure.Decode(orgDoc, &org)

	setting := new(Settings)

	newMember := Member{
		ID:       primitive.NewObjectID(),
		Email:    user.Email,
		UserName: newUserName,
		OrgId:    orgId.Hex(),
		Role:     "member",
		Presence: "true",
		JoinedAt: time.Now(),
		// Settings: utils.M{},
		Settings: setting,
		Deleted:  false,
	}

	coll := utils.GetCollection(member_collection)
	res, err := coll.InsertOne(r.Context(), newMember)
	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, w)
		return
	}

	// update user organizations collection
	updateFields := make(map[string]interface{})
	user.Organizations = append(user.Organizations, sOrgId)
	updateFields["Organizations"] = user.Organizations
	_, eerr := utils.UpdateOneMongoDbDoc(user_collection, user.ID, updateFields)
	if eerr != nil {
		utils.GetError(errors.New("user update failed"), http.StatusInternalServerError, w)
		return
	}

	utils.GetSuccess("Member created successfully", utils.M{"member_id": res.InsertedID}, w)
}

// endpoint to update a member's profile picture
func UpdateProfilePicture(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-Type", "application/json")
	org_collection, member_collection := "organizations", "members"

	orgId := mux.Vars(r)["id"]
	member_Id := mux.Vars(r)["mem_id"]

	pMemId, err := primitive.ObjectIDFromHex(member_Id)
	if err != nil {
		utils.GetError(errors.New("invalid Member id"), http.StatusBadRequest, w)
		return
	}

	pOrgId, err := primitive.ObjectIDFromHex(orgId)
	if err != nil {
		utils.GetError(errors.New("invalid organization id"), http.StatusBadRequest, w)
		return
	}

	orgDoc, _ := utils.GetMongoDbDoc(org_collection, bson.M{"_id": pOrgId})
	if orgDoc == nil {
		fmt.Printf("org with id %s doesn't exist!", orgId)
		utils.GetError(errors.New("organization does not exist"), http.StatusBadRequest, w)
		return
	}

	memberDoc, _ := utils.GetMongoDbDoc(member_collection, bson.M{"_id": pMemId, "org_id": orgId})
	if memberDoc == nil {
		fmt.Printf("member with id %s doesn't exist!", member_Id)
		utils.GetError(errors.New("member does not exist"), http.StatusBadRequest, w)
		return

	}

	uploadPath := "profile_image/" + orgId + "/" + member_Id
	img_url, erro := service.ProfileImageUpload(uploadPath, r)
	if erro != nil {
		utils.GetError(erro, http.StatusInternalServerError, w)
		return
	}

	result, err := utils.UpdateOneMongoDbDoc(member_collection, member_Id, bson.M{"image_url": img_url})
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

// an endpoint to update a user status
func UpdateMemberStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")

	org_collection, member_collection := "organizations", "members"

	// Validate the user ID
	orgId := mux.Vars(r)["id"]
	member_Id := mux.Vars(r)["mem_id"]

	pMemId, err := primitive.ObjectIDFromHex(member_Id)
	if err != nil {
		utils.GetError(errors.New("invalid member id"), http.StatusBadRequest, w)
		return
	}

	pOrgId, err := primitive.ObjectIDFromHex(orgId)
	if err != nil {
		utils.GetError(errors.New("invalid organization id"), http.StatusBadRequest, w)
		return
	}

	requestData := make(map[string]string)
	if err := utils.ParseJsonFromRequest(r, &requestData); err != nil {
		utils.GetError(err, http.StatusUnprocessableEntity, w)
		return
	}

	member_status := requestData["status"]

	orgDoc, _ := utils.GetMongoDbDoc(org_collection, bson.M{"_id": pOrgId})
	if orgDoc == nil {
		fmt.Printf("org with id %s doesn't exist!", orgId)
		utils.GetError(errors.New("org with id %s doesn't exist"), http.StatusBadRequest, w)
		return
	}

	memberDoc, _ := utils.GetMongoDbDoc(member_collection, bson.M{"_id": pMemId, "org_id": orgId})
	if memberDoc == nil {
		fmt.Printf("member with id %s doesn't exist!", member_Id)
		utils.GetError(errors.New("member with id doesn't exist"), http.StatusBadRequest, w)
		return
	}

	result, err := utils.UpdateOneMongoDbDoc(member_collection, member_Id, bson.M{"status": member_status})
	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, w)
		return
	}

	if result.ModifiedCount == 0 {
		utils.GetError(errors.New("operation failed"), http.StatusInternalServerError, w)
		return
	}

	utils.GetSuccess("status updated successfully", nil, w)
}

// Delete single member from an organization
func DeactivateMember(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	member_collection, org_collection := "members", "organizations"
	vars := mux.Vars(r)
	orgId, memberId := vars["id"], vars["mem_id"]

	pOrgId, err := primitive.ObjectIDFromHex(orgId)
	if err != nil {
		utils.GetError(errors.New("invalid id"), http.StatusBadRequest, w)
		return
	}

	orgDoc, _ := utils.GetMongoDbDoc(org_collection, bson.M{"_id": pOrgId})
	if orgDoc == nil {
		fmt.Printf("org with id %s doesn't exist!", orgId)
		utils.GetError(errors.New("operation failed"), http.StatusBadRequest, w)
		return
	}

	deleteUpdate := bson.M{"deleted": true, "deleted_at": time.Now()}
	res, err := utils.UpdateOneMongoDbDoc(member_collection, memberId, deleteUpdate)

	if err != nil {
		utils.GetError(fmt.Errorf("an error occured: %s", err), http.StatusInternalServerError, w)
	}

	if res.MatchedCount != 1 {
		utils.GetError(fmt.Errorf("member with id %s not found", memberId), http.StatusNotFound, w)
		return
	}

	if res.ModifiedCount != 1 {
		utils.GetError(errors.New("an error occured, cannot delete user"), http.StatusInternalServerError, w)
		return
	}
	utils.GetSuccess("successfully deactivated member", nil, w)
}

// Update a member profile
func UpdateProfile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	org_collection, member_collection := "organizations", "members"

	id := mux.Vars(r)["id"]
	memId := mux.Vars(r)["mem_id"]

	// Check if organization id is valid
	orgId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		utils.GetError(errors.New("invalid id"), http.StatusBadRequest, w)
		return
	}

	// Check if organization id is exists in the database
	orgDoc, _ := utils.GetMongoDbDoc(org_collection, bson.M{"_id": orgId})
	if orgDoc == nil {
		fmt.Printf("organization with ID: %s does not exist ", id)
		utils.GetError(errors.New("operation failed"), http.StatusBadRequest, w)
		return
	}

	// Check if member id is valid
	pMemId, err := primitive.ObjectIDFromHex(memId)
	if err != nil {
		utils.GetError(errors.New("invalid id"), http.StatusBadRequest, w)
		return
	}

	memberDoc, _ := utils.GetMongoDbDoc(member_collection, bson.M{"_id": pMemId, "org_id": id})
	if memberDoc == nil {
		fmt.Printf("member with id %s doesn't exist!", memId)
		utils.GetError(errors.New("member with id doesn't exist"), http.StatusBadRequest, w)
		return
	}

	// Get data from request
	var memberProfile Profile
	err = utils.ParseJsonFromRequest(r, &memberProfile)
	if err != nil {
		utils.GetError(err, http.StatusUnprocessableEntity, w)
		return
	}

	if len(memberProfile.Socials) > 3 {
		utils.GetError(errors.New("socials must be 3 or less"), http.StatusBadRequest, w)
		return
	}

	// convert struct to map
	mProfile, err := utils.StructToMap(memberProfile)
	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, w)
		return
	}

	// Fetch and update the MemberDoc from collection
	update, err := utils.UpdateOneMongoDbDoc(member_collection, memId, mProfile)
	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, w)
		return
	}

	if update.ModifiedCount == 0 {
		utils.GetError(errors.New("operation failed"), http.StatusInternalServerError, w)
		return
	}

	utils.GetSuccess("Member Profile updated succesfully", nil, w)
}

// Toggle a member's presence
func TogglePresence(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	org_collection, member_collection := "organizations", "members"
	id := mux.Vars(r)["id"]
	memId := mux.Vars(r)["mem_id"]

	// Check if organization id is valid
	orgId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		utils.GetError(errors.New("invalid id"), http.StatusBadRequest, w)
		return
	}

	// Check if organization id exists in the database
	orgDoc, _ := utils.GetMongoDbDoc(org_collection, bson.M{"_id": orgId})
	if orgDoc == nil {
		fmt.Printf("organization with ID: %s does not exist ", id)
		utils.GetError(errors.New("operation failed"), http.StatusBadRequest, w)
		return
	}

	// Check if member id is valid
	pMemId, err := primitive.ObjectIDFromHex(memId)
	if err != nil {
		utils.GetError(errors.New("invalid id"), http.StatusBadRequest, w)
		return
	}

	memberDoc, _ := utils.GetMongoDbDoc(member_collection, bson.M{"_id": pMemId, "org_id": id})
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
	update, err := utils.UpdateOneMongoDbDoc(member_collection, memId, org_filter)
	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, w)
		return
	}

	if update.ModifiedCount == 0 {
		utils.GetError(errors.New("operation failed"), http.StatusInternalServerError, w)
		return
	}

	utils.GetSuccess("Member presence toggled", nil, w)
}

func UpdateMemberSettings(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	org_collection, member_collection := "organizations", "members"

	vars := mux.Vars(r)
	id, memberId := vars["id"], vars["mem_id"]

	// Check if organization id is valid and it exists in the database
	orgId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		utils.GetError(errors.New("invalid id"), http.StatusBadRequest, w)
		return
	}

	orgDoc, _ := utils.GetMongoDbDoc(org_collection, bson.M{"_id": orgId})
	if orgDoc == nil {
		fmt.Printf("organization with ID: %s does not exist ", id)
		utils.GetError(errors.New("operation failed"), http.StatusBadRequest, w)
		return
	}

	// Check if member id is valid and it exists in the database
	pMemId, err := primitive.ObjectIDFromHex(memberId)
	if err != nil {
		utils.GetError(errors.New("invalid id"), http.StatusBadRequest, w)
		return
	}

	memberDoc, _ := utils.GetMongoDbDoc(member_collection, bson.M{"_id": pMemId, "org_id": id})
	if memberDoc == nil {
		fmt.Printf("member with id %s doesn't exist!", memberId)
		utils.GetError(errors.New("member with id doesn't exist"), http.StatusBadRequest, w)
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
	update, err := utils.UpdateOneMongoDbDoc(member_collection, memberId, memberSettings)
	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, w)
		return
	}

	if update.ModifiedCount == 0 {
		utils.GetError(errors.New("operation failed"), http.StatusInternalServerError, w)
		return
	}

	utils.GetSuccess("Member settings updated successfully", nil, w)
}

// Activate single member in an organization
func ReactivateMember(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	member_collection, org_collection := "members", "organizations"
	vars := mux.Vars(r)
	orgId, memberId := vars["id"], vars["mem_id"]

	pOrgId, err := primitive.ObjectIDFromHex(orgId)
	if err != nil {
		utils.GetError(errors.New("invalid id"), http.StatusBadRequest, w)
		return
	}

	orgDoc, _ := utils.GetMongoDbDoc(org_collection, bson.M{"_id": pOrgId})
	if orgDoc == nil {
		fmt.Printf("org with id %s doesn't exist!", orgId)
		utils.GetError(errors.New("operation failed"), http.StatusBadRequest, w)
		return
	}

	ActivatedMember := bson.M{"deleted": false, "deleted_at": time.Time{}}
	res, err := utils.UpdateOneMongoDbDoc(member_collection, memberId, ActivatedMember)

	if err != nil {
		utils.GetError(fmt.Errorf("an error occured: %s", err), http.StatusInternalServerError, w)
	}

	if res.MatchedCount != 1 {
		utils.GetError(fmt.Errorf("member with id %s not found", memberId), http.StatusNotFound, w)
		return
	}

	if res.ModifiedCount != 1 {
		utils.GetError(errors.New("an error occured, cannot activate user"), http.StatusInternalServerError, w)
		return
	}
	utils.GetSuccess("successfully reactivated member", nil, w)
}

// Add invited guest to an organization
func InviteGuest(w http.ResponseWriter, r *http.Request) {
	// 0. Extract and validate UUID
	guestUUID := mux.Vars(r)["uuid"]
	validUUID, err := ValidateUUID(guestUUID)
	if err != nil {
		utils.GetError(err, http.StatusBadRequest, w)
		return
	}

	fmt.Println(validUUID)
	// 1. Query organization  invites collection
	orgInvitesCollection := "organizations_invites"
	res, err := utils.GetMongoDbDoc(orgInvitesCollection, bson.M{"uuid": validUUID})
	if err != nil {
		utils.GetError(err, http.StatusBadRequest, w)
		return
	}
	fmt.Println(res)

}

func ValidateUUID(s string) (uuid.UUID, error) {
	if len(s) != 36 {
		return uuid.Nil, errors.New("invalid uuid format")
	}

	b, err := uuid.FromString(s)
	if err != nil {
		return uuid.Nil, err
	}

	return b, nil
}

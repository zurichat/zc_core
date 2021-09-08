package organizations

import (
	"errors"
	"fmt"
	"net/http"

	"zuri.chat/zccore/user"

	"github.com/gorilla/mux"
	"github.com/mitchellh/mapstructure"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"zuri.chat/zccore/utils"
)

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

	orgMembers, err := utils.GetMongoDbDocs(member_collection, bson.M{"org_id": orgId})
	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, w)
		return
	}

	utils.GetSuccess("Members retrieved successfully", orgMembers, w)
}

func CreateMember(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	org_collection, user_collection, member_collection := "organizations", "users", "members"

	orgId, err := primitive.ObjectIDFromHex(mux.Vars(r)["id"])
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

	userId, err := primitive.ObjectIDFromHex(requestData["user_id"])
	if err != nil {
		utils.GetError(errors.New("invalid id"), http.StatusBadRequest, w)
		return
	}

	userDoc, _ := utils.GetMongoDbDoc(user_collection, bson.M{"_id": userId})
	if userDoc == nil {
		fmt.Printf("user with id %s doesn't exist!", userId.String())
		utils.GetError(errors.New("operation failed"), http.StatusBadRequest, w)
		return
	}

	// convert user to struct
	var user user.User
	mapstructure.Decode(userDoc, &user)

	// get organization
	orgDoc, _ := utils.GetMongoDbDoc(org_collection, bson.M{"_id": orgId})
	if orgDoc == nil {
		fmt.Printf("organization with id %s doesn't exist!", orgId.String())
		utils.GetError(errors.New("operation failed"), http.StatusBadRequest, w)
		return
	}

	// convert org to struct
	var org Organization
	mapstructure.Decode(orgDoc, &org)

	newMember := Member{
		Email: user.Email,
		OrgId: orgId,
	}

	// conv to struct
	memStruc, err := utils.StructToMap(newMember)
	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, w)
		return
	}

	// check that member isn't already in the organization
	memDoc, _ := utils.GetMongoDbDoc(member_collection, bson.M{"org_id": orgId, "email": newMember.Email})
	if memDoc != nil {
		fmt.Printf("organization %s has member with email %s!", orgId.String(), newMember.Email)
		utils.GetError(errors.New("operation failed"), http.StatusBadRequest, w)
		return
	}

	// add new member to member collection
	createdMember, err := utils.CreateMongoDbDoc(member_collection, memStruc)
	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, w)
		return
	}

	// Todo: update user workspace with the org ID
	utils.GetSuccess("Member created successfully", createdMember, w)
}

// endpoint to update a member profile picture
func UpdateProfilePicture(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-Type", "application/json")
	org_collection, member_collection := "organizations", "members"

	orgId := mux.Vars(r)["id"]
	member_Id := mux.Vars(r)["mem_id"]

	requestData := make(map[string]string)
	if err := utils.ParseJsonFromRequest(r, &requestData); err != nil {
		utils.GetError(err, http.StatusUnprocessableEntity, w)
		return
	}
	image_url := requestData["image_url"]

	pMemId, err := primitive.ObjectIDFromHex(member_Id)
	if err != nil {
		utils.GetError(errors.New("invalid id"), http.StatusBadRequest, w)
		return
	}

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

	memberDoc, _ := utils.GetMongoDbDoc(member_collection, bson.M{"_id": pMemId, "org_id": pOrgId})
	if memberDoc == nil {
		fmt.Printf("member with id %s doesn't exist!", member_Id)
		utils.GetError(errors.New("operation failed"), http.StatusBadRequest, w)
		return

	}

	result, err := utils.UpdateOneMongoDbDoc(member_collection, member_Id, bson.M{"image_url": image_url})
	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, w)
		return
	}

	utils.GetSuccess("Profile picture updated", result, w)

}

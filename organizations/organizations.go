package organizations

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"zuri.chat/zccore/utils"
)

func GetOrganization(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	collection := "organizations"

	orgId := mux.Vars(r)["id"]
	objId, _ := primitive.ObjectIDFromHex(orgId)

	save, err := utils.GetMongoDbDocs(collection, bson.M{"_id": objId})

	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, w)
		return
	}
	utils.GetSuccess("organization retrieved successfully", save, w)
}

func Create(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var newOrg Organization
	collection, user_collection := "organizations", "users"

	// Try to decode the request body into the struct. If there is an error,
	// respond to the client with the error message and a 400 status code.
	err := json.NewDecoder(r.Body).Decode(&newOrg)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// validate that email is not empty and it meets the format
	if !utils.IsValidEmail(newOrg.Email) {
		utils.GetError(fmt.Errorf("invalid email format : %s", newOrg.Email), http.StatusInternalServerError, w)
		return
	}

	// set default name if name is empty
	if newOrg.Name == "" {
		newOrg.Name = "ZuriWorkspace"
	}

	// confirm if user_id exists
	user_filter := make(map[string]interface{})
	user_filter["user_id"] = newOrg.CreatorID
	user, _ := utils.GetMongoDbDoc(user_collection, user_filter)
	if user == nil {
		utils.GetError(errors.New("invalid user id"), http.StatusBadRequest, w)
		return
	}

	// save organization
	newOrg.ImageURL = newOrg.Name + ".zuri.chat"
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

	save, err := utils.DeleteOneMongoDoc(collection, orgId)
	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, w)
		return
	}

	utils.GetSuccess("organization deleted successfully", save, w)
}

func AddMember(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var newMember Member

	orgId := mux.Vars(r)["id"]
	collection, user_collection := "organizations", "users"

	err := json.NewDecoder(r.Body).Decode(&newMember)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// validate that email is not empty and it meets the format
	if !utils.IsValidEmail(newMember.Email) {
		utils.GetError(fmt.Errorf("invalid email format : %s", newMember.Email), http.StatusInternalServerError, w)
		return
	}

	// confirm if user_id exists
	user_filter := make(map[string]interface{})
	user_filter["email"] = newMember.Email
	user, _ := utils.GetMongoDbDoc(user_collection, user_filter)
	if user == nil {
		utils.GetError(errors.New("invalid user id"), http.StatusBadRequest, w)
		return
	}

	// convert to map object
	response, err := utils.StructToMap(newMember, "bson")
	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, w)
		return
	}
	org, err := utils.GetMongoDbDoc(collection, bson.M{"id": orgId})
	fmt.Println(org)
	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, w)
		return
	}

	// save organization
	newMember.JoinedAt = time.Now()
	save, err := utils.UpdateOneMongoDbDoc(collection, orgId, response)
	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, w)
		return
	}
	utils.GetSuccess("Member created", save, w)


}

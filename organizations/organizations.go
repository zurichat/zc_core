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
	objId, err := primitive.ObjectIDFromHex(newOrg.CreatorID)

	if err != nil {
		utils.GetError(errors.New("invalid id"), http.StatusBadRequest, w)
		return
	}

	user, _ := utils.GetMongoDbDoc(user_collection, bson.M{"_id": objId})
	if user == nil {
		fmt.Printf("users with id %s exists!", newOrg.CreatorID)
		utils.GetError(errors.New("operation failed"), http.StatusBadRequest, w)
		return
	}

	newOrg.URL = newOrg.Name + ".zuri.chat"
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
	url := requestData["url"]

	collection := "organizations"
	org_filter := make(map[string]interface{})
	org_filter["url"] = url
	update, err := utils.UpdateOneMongoDbDoc(collection, orgId, org_filter)
	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, w)
		return
	}

	utils.GetSuccess("organization url updated successfully", update, w)

}

func UpdateName(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	orgId := mux.Vars(r)["id"]
	requestData := make(map[string]string)
	if err := utils.ParseJsonFromRequest(r, &requestData); err != nil {
		utils.GetError(err, http.StatusUnprocessableEntity, w)
		return
	}
	organization_name := requestData["organization_name"]

	collection := "organizations"
	org_filter := make(map[string]interface{})
	org_filter["name"] = organization_name
	change, err := utils.UpdateOneMongoDbDoc(collection, orgId, org_filter)
	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, w)
		return
	}

	utils.GetSuccess("organization name successfully changed", change, w)

}

func UpdateLogo(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")
	orgId := mux.Vars(r)["id"]
	requestData := make(map[string]string)
	if err := utils.ParseJsonFromRequest(r, &requestData); err != nil {
		utils.GetError(err, http.StatusUnprocessableEntity, w)
		return
	}
	organization_logo := requestData["organization_logo"]

	collection := "organizations"
	org_filter := make(map[string]interface{})
	org_filter["image_url"] = organization_logo
	change, err := utils.UpdateOneMongoDbDoc(collection, orgId, org_filter)
	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, w)
		return
	}

	utils.GetSuccess("organization logo successfully changed", change, w)

}

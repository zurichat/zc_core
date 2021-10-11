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

var ErrNoAuthToken = errors.New("no authorization header provided")

func (oh *OrganizationHandler) AddOrganizationPlugin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var orgPlugin OrgPluginBody

	OrgID := mux.Vars(r)["id"]

	err := json.NewDecoder(r.Body).Decode(&orgPlugin)
	if err != nil {
		utils.GetError(err, http.StatusBadRequest, w)
		return
	}

	// confirm if plugin_id exists
	pluginID, err := primitive.ObjectIDFromHex(orgPlugin.PluginID)

	if err != nil {
		utils.GetError(errors.New("invalid plugin id"), http.StatusBadRequest, w)
		return
	}

	plugin, _ := utils.GetMongoDBDoc(PluginCollection, bson.M{"_id": pluginID})

	if plugin == nil {
		utils.GetError(errors.New("operation failed"), http.StatusBadRequest, w)
		return
	}

	// confirm if user_id exists
	creatorID, err := primitive.ObjectIDFromHex(orgPlugin.UserID)

	if err != nil {
		utils.GetError(errors.New("invalid user id"), http.StatusBadRequest, w)
		return
	}

	user, _ := utils.GetMongoDBDoc(MemberCollectionName, bson.M{"_id": creatorID, "org_id": OrgID})
	if user == nil {
		utils.GetError(errors.New("member doesn't exist in the organization"), http.StatusBadRequest, w)
		return
	}

	var member Member
	if err = utils.ConvertStructure(user, &member); err != nil {
		utils.GetError(err, http.StatusInternalServerError, w)
		return
	}

	if member.Role != OwnerRole && member.Role != AdminRole {
		utils.GetError(errors.New("access denied"), http.StatusForbidden, w)
		return
	}

	orgCollectionName := GetOrgPluginCollectionName(OrgID)

	p, _ := utils.GetMongoDBDoc(orgCollectionName, bson.M{"plugin_id": orgPlugin.PluginID})

	if p != nil {
		utils.GetError(errors.New("plugin has already been added"), http.StatusBadRequest, w)
		return
	}

	userName := user["first_name"].(string) + " " + user["last_name"].(string)

	installedPlugin := InstalledPlugin{
		PluginID:    orgPlugin.PluginID,
		Plugin:      plugin,
		AddedBy:     userName,
		ApprovedBy:  userName,
		InstalledAt: time.Now(),
	}

	var pluginMap map[string]interface{}

	pluginJSON, _ := json.Marshal(installedPlugin)

	err = json.Unmarshal(pluginJSON, &pluginMap)
	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, w)
		return
	}

	// save organization
	save, err := utils.CreateMongoDBDoc(orgCollectionName, pluginMap)
	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, w)
		return
	}

	utils.GetSuccess("plugin saved successfully", save, w)
}

func (oh *OrganizationHandler) GetOrganizationPlugins(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	collection := "organizations"

	orgID := mux.Vars(r)["id"]
	objID, err := primitive.ObjectIDFromHex(orgID)

	if err != nil {
		utils.GetError(errors.New("invalid id"), http.StatusBadRequest, w)
		return
	}

	save, _ := utils.GetMongoDBDoc(collection, bson.M{"_id": objID})

	if save == nil {
		utils.GetError(fmt.Errorf("organization %s not found", orgID), http.StatusNotFound, w)
		return
	}

	var org Organization

	// convert bson to struct
	bsonBytes, _ := bson.Marshal(save)
	err = bson.Unmarshal(bsonBytes, &org)

	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, w)
		return
	}

	utils.GetSuccess("plugins Retrieved successfully", org.OrgPlugins(), w)
}

func (oh *OrganizationHandler) GetOrganizationPlugin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	orgID := mux.Vars(r)["id"]
	pluginID := mux.Vars(r)["plugin_id"]

	objID, err := primitive.ObjectIDFromHex(orgID)

	if err != nil {
		utils.GetError(errors.New("invalid id"), http.StatusBadRequest, w)
		return
	}

	save, _ := utils.GetMongoDBDoc(OrganizationCollectionName, bson.M{"_id": objID})

	if save == nil {
		utils.GetError(fmt.Errorf("organization %s not found", orgID), http.StatusNotFound, w)
		return
	}

	orgCollectionName := GetOrgPluginCollectionName(orgID)

	doc, err := utils.GetMongoDBDoc(orgCollectionName, bson.M{"plugin_id": pluginID})

	if err != nil {
		// plugin not found in organization.
		utils.GetError(err, http.StatusNotFound, w)
		return
	}

	utils.GetSuccess("plugins returned successfully", doc, w)
}

func (oh *OrganizationHandler) RemoveOrganizationPlugin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var orgPlugin OrgPluginBody

	orgID := mux.Vars(r)["id"]
	pluginID := mux.Vars(r)["plugin_id"]

	err := json.NewDecoder(r.Body).Decode(&orgPlugin)
	if err != nil {
		utils.GetError(err, http.StatusBadRequest, w)
		return
	}

	// confirm if user_id exists
	creatorID, err := primitive.ObjectIDFromHex(orgPlugin.UserID)

	if err != nil {
		utils.GetError(errors.New("invalid user id"), http.StatusBadRequest, w)
		return
	}

	user, _ := utils.GetMongoDBDoc(MemberCollectionName, bson.M{"_id": creatorID, "org_id": orgID})
	if user == nil {
		utils.GetError(errors.New("member doesn't exist in the organization"), http.StatusBadRequest, w)
		return
	}

	var member Member
	if err = utils.ConvertStructure(user, &member); err != nil {
		utils.GetError(err, http.StatusInternalServerError, w)
		return
	}

	if member.Role != OwnerRole && member.Role != AdminRole {
		utils.GetError(errors.New("access denied"), http.StatusForbidden, w)
		return
	}

	orgCollectionName := GetOrgPluginCollectionName(orgID)

	orgPluginDoc, _ := utils.GetMongoDBDoc(orgCollectionName, bson.M{"plugin_id": pluginID})
	if orgPluginDoc == nil {
		// plugin not found in organization.
		utils.GetError(errors.New("plugin does not exist"), http.StatusBadRequest, w)
		return
	}

	orgPluginObjID := orgPluginDoc["_id"]
	orgPluginID := orgPluginObjID.(primitive.ObjectID).Hex()

	doc, err := utils.DeleteOneMongoDBDoc(orgCollectionName, orgPluginID)

	if err != nil {
		// plugin not found in organization.
		utils.GetError(err, http.StatusNotFound, w)
		return
	}

	if doc.DeletedCount == 0 {
		utils.GetError(errors.New("plugin failed to uninstall"), http.StatusBadRequest, w)
		return
	}

	utils.GetSuccess("plugin removed successfully", doc, w)
}

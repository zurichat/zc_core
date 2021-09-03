package organizations

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"zuri.chat/zccore/utils"
)

func AddOrganizationPlugin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	org_collection, plugin_collection, user_collection := "organizations", "plugins", "users"
	var orgPlugin OrgPluginBody

	OrgId := mux.Vars(r)["id"]

	err := json.NewDecoder(r.Body).Decode(&orgPlugin)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// confirm if organization_id exists
	orgId, err:= primitive.ObjectIDFromHex(OrgId)

	if err != nil {
		utils.GetError(errors.New("invalid organization id"), http.StatusBadRequest, w)
		return
	}

	org, _ := utils.GetMongoDbDoc(org_collection, bson.M{"_id": orgId})

	if org == nil {
		utils.GetError(errors.New("operation failed"), http.StatusBadRequest, w)
		return
	}

	// confirm if plugin_id exists
	pluginId, err:= primitive.ObjectIDFromHex(orgPlugin.PluginId)

	if err != nil {
		utils.GetError(errors.New("invalid plugin id"), http.StatusBadRequest, w)
		return
	}

	plugin, _ := utils.GetMongoDbDoc(plugin_collection, bson.M{"_id": pluginId})

	if plugin == nil {
		utils.GetError(errors.New("operation failed"), http.StatusBadRequest, w)
		return
	}

	// confirm if user_id exists
	creatorId, err:= primitive.ObjectIDFromHex(orgPlugin.UserId)

	if err != nil {
		utils.GetError(errors.New("invalid user id"), http.StatusBadRequest, w)
		return
	}

	user, _ := utils.GetMongoDbDoc(user_collection, bson.M{"_id": creatorId})
	if user == nil {
		utils.GetError(errors.New("operation failed"), http.StatusBadRequest, w)
		return
	}

	orgCollectionName := GetOrgPluginCollectionName(org["name"].(string))
	userName := user["first_name"].(string) + " " + user["last_name"].(string)

	installedPlugin := InstalledPlugin {
		PluginID: orgPlugin.PluginId,
		Plugin: plugin,
		AddedBy: userName,
		ApprovedBy: userName,
		InstalledAt: time.Now(),
	}

	var pluginMap map[string]interface{}
	pluginJson, _ := json.Marshal(installedPlugin)
	json.Unmarshal(pluginJson, &pluginMap)

	// save organization
	save, err := utils.CreateMongoDbDoc(orgCollectionName, pluginMap)
	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, w)
		return
	}
	utils.GetSuccess("plugin saved successfully", save, w)
}

func GetOrganizationPlugins(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	collection := "organizations"

	orgId := mux.Vars(r)["id"]
	objId, _ := primitive.ObjectIDFromHex(orgId)

	doc, err := utils.GetMongoDbDoc(collection, bson.M{"_id": objId})
	if err != nil {
		// org not found.
		utils.GetError(err, http.StatusNotFound, w)
		return
	}

	orgCollectionName := GetOrgPluginCollectionName(doc["name"].(string))

	docs, err := utils.GetMongoDbDocs(orgCollectionName, nil)
  
	if err != nil {
		// org plugins not found.
		utils.GetError(err, http.StatusNotFound, w)
		return
	}

	utils.GetSuccess("Plugins returned successfully", docs, w)
}

func GetOrganizationPlugin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	collection := "organizations"

	orgId := mux.Vars(r)["id"]
	pluginId := mux.Vars(r)["plugin_id"]

	OrgId, _ := primitive.ObjectIDFromHex(orgId)
	
	OrgDoc, err := utils.GetMongoDbDoc(collection, bson.M{"_id": OrgId})
	if err != nil {
		// org not found.
		utils.GetError(err, http.StatusNotFound, w)
		return
	}

	orgCollectionName := GetOrgPluginCollectionName(OrgDoc["name"].(string))

	doc, err := utils.GetMongoDbDoc(orgCollectionName, bson.M{"plugin_id": pluginId})

	if err != nil {
		// plugin not found in organiztion.
		utils.GetError(err, http.StatusNotFound, w)
		return
	}

	utils.GetSuccess("Plugins returned successfully", doc, w)

}

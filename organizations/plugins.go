package organizations

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	// "zuri.chat/zccore/auth"

	"zuri.chat/zccore/utils"
)

var NoAuthToken = errors.New("No Authorization header provided.")

func AddOrganizationPlugin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	plugin_collection, user_collection := "plugins", "users"
	var orgPlugin OrgPluginBody

	OrgId := mux.Vars(r)["id"]

	err := json.NewDecoder(r.Body).Decode(&orgPlugin)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// confirm if plugin_id exists
	pluginId, err := primitive.ObjectIDFromHex(orgPlugin.PluginId)

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
	creatorId, err := primitive.ObjectIDFromHex(orgPlugin.UserId)

	if err != nil {
		utils.GetError(errors.New("invalid user id"), http.StatusBadRequest, w)
		return
	}

	user, _ := utils.GetMongoDbDoc(user_collection, bson.M{"_id": creatorId})
	if user == nil {
		utils.GetError(errors.New("user does not exist"), http.StatusBadRequest, w)
		return
	}

	orgCollectionName := GetOrgPluginCollectionName(OrgId)

	p, _ := utils.GetMongoDbDoc(orgCollectionName, bson.M{"plugin_id": orgPlugin.PluginId})

	if p != nil {
		utils.GetError(errors.New("plugin has already been added"), http.StatusBadRequest, w)
		return
	}

	userName := user["first_name"].(string) + " " + user["last_name"].(string)

	installedPlugin := InstalledPlugin {
		PluginID:    orgPlugin.PluginId,
		Plugin:      plugin,
		AddedBy:     userName,
		ApprovedBy:  userName,
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
	// loggedInUser := r.Context().Value("user").(auth.AuthUser)

	orgId := mux.Vars(r)["id"]

	orgCollectionName := GetOrgPluginCollectionName(orgId)
	// member_collection, user_collection := "members", "users"

	docs, err := utils.GetMongoDbDocs(orgCollectionName, nil)
	if err != nil {
		// org plugins not found.
		utils.GetError(err, http.StatusNotFound, w)
		return
	}

	if len(docs) == 0 {
		utils.GetError(errors.New("plugin has not been added"), http.StatusBadRequest, w)
		return
	}

	// userDoc, _ := utils.GetMongoDbDoc(user_collection, bson.M{"_id": loggedInUser.ID.Hex()})
	// if userDoc == nil {
	// 	utils.GetError(errors.New("Invalid User"), http.StatusBadRequest, w)
	// 	return
	// }

	// convert user to struct
	// var user user.User
	// mapstructure.Decode(userDoc, &user)

	// memDoc, _ := utils.GetMongoDbDoc(member_collection, bson.M{"org_id": orgId, "email": user.Email})
	// if memDoc == nil {
	// 	utils.GetError(errors.New("You're not authorized to access this resources"), http.StatusUnauthorized, w)
	// 	return
	// }

	utils.GetSuccess("Plugins Retrived successfully", docs, w)
}

func GetOrganizationPlugin(w http.ResponseWriter, r *http.Request) {
	// loggedInUser := r.Context().Value("user").(auth.AuthUser)

	orgId := mux.Vars(r)["id"]
	pluginId := mux.Vars(r)["plugin_id"]

	orgCollectionName := GetOrgPluginCollectionName(orgId)
	// member_collection, user_collection := "members", "users"

	doc, err := utils.GetMongoDbDoc(orgCollectionName, bson.M{"plugin_id": pluginId})

	if err != nil {
		// plugin not found in organiztion.
		utils.GetError(err, http.StatusNotFound, w)
		return
	}

	utils.GetSuccess("Plugins returned successfully", doc, w)
}

func OrganizationPlugins(orgId string) ([]map[string]interface{}, error) {
	orgCollectionName := GetOrgPluginCollectionName(orgId)

	orgPlugins, err := utils.GetMongoDbDocs(orgCollectionName, nil)
	if err != nil {
		// org plugins not found.
		return nil, err
	}

	var pluginsMap []map[string]interface{}
	pluginJson, _ := json.Marshal(orgPlugins)
	json.Unmarshal(pluginJson, &pluginsMap)

	return pluginsMap, nil
}
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

var NoAuthToken = errors.New("No Authorization header provided.")

func (oh *OrganizationHandler) AddOrganizationPlugin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	plugin_collection := "plugins"
	var orgPlugin OrgPluginBody

	OrgId := mux.Vars(r)["id"]

	err := json.NewDecoder(r.Body).Decode(&orgPlugin)
	if err != nil {
		utils.GetError(err, http.StatusBadRequest, w)
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

	user, _ := utils.GetMongoDbDoc(MemberCollectionName, bson.M{"_id": creatorId, "org_id": OrgId})
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

	orgCollectionName := GetOrgPluginCollectionName(OrgId)

	p, _ := utils.GetMongoDbDoc(orgCollectionName, bson.M{"plugin_id": orgPlugin.PluginId})

	if p != nil {
		utils.GetError(errors.New("plugin has already been added"), http.StatusBadRequest, w)
		return
	}

	userName := user["first_name"].(string) + " " + user["last_name"].(string)

	installedPlugin := InstalledPlugin{
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

func (oh *OrganizationHandler) GetOrganizationPlugins(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	// loggedInUser := r.Context().Value("user").(auth.AuthUser)

	collection := "organizations"

	orgId := mux.Vars(r)["id"]
	objId, err := primitive.ObjectIDFromHex(orgId)

	if err != nil {
		utils.GetError(errors.New("invalid id"), http.StatusBadRequest, w)
		return
	}

	save, _ := utils.GetMongoDbDoc(collection, bson.M{"_id": objId})

	if save == nil {
		utils.GetError(fmt.Errorf("organization %s not found", orgId), http.StatusNotFound, w)
		return
	}

	var org Organization
	// convert bson to struct
	bsonBytes, _ := bson.Marshal(save)
	bson.Unmarshal(bsonBytes, &org)

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

	utils.GetSuccess("Plugins Retrived successfully", org.OrgPlugins(), w)
}

func (oh *OrganizationHandler) GetOrganizationPlugin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	// loggedInUser := r.Context().Value("user").(auth.AuthUser)

	orgId := mux.Vars(r)["id"]
	pluginId := mux.Vars(r)["plugin_id"]

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

	orgCollectionName := GetOrgPluginCollectionName(orgId)

	doc, err := utils.GetMongoDbDoc(orgCollectionName, bson.M{"plugin_id": pluginId})

	if err != nil {
		// plugin not found in organiztion.
		utils.GetError(err, http.StatusNotFound, w)
		return
	}

	utils.GetSuccess("Plugins returned successfully", doc, w)
}

func (oh *OrganizationHandler) RemoveOrganizationPlugin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	var orgPlugin OrgPluginBody

	orgId := mux.Vars(r)["id"]
	pluginId := mux.Vars(r)["plugin_id"]

	err := json.NewDecoder(r.Body).Decode(&orgPlugin)
	if err != nil {
		utils.GetError(err, http.StatusBadRequest, w)
		return
	}

	// confirm if user_id exists
	creatorId, err := primitive.ObjectIDFromHex(orgPlugin.UserId)

	if err != nil {
		utils.GetError(errors.New("invalid user id"), http.StatusBadRequest, w)
		return
	}

	user, _ := utils.GetMongoDbDoc(MemberCollectionName, bson.M{"_id": creatorId, "org_id": orgId})
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

	orgCollectionName := GetOrgPluginCollectionName(orgId)

	orgPluginDoc, _ := utils.GetMongoDbDoc(orgCollectionName, bson.M{"plugin_id": pluginId})
	if orgPluginDoc == nil {
		// plugin not found in organiztion.
		utils.GetError(errors.New("plugin does not exist"), http.StatusBadRequest, w)
		return
	}
	
	orgPluginObjId := orgPluginDoc["_id"]
	orgPluginId := orgPluginObjId.(primitive.ObjectID).Hex()
	
	doc, err := utils.DeleteOneMongoDoc(orgCollectionName, orgPluginId)

	if err != nil {
		// plugin not found in organiztion.
		utils.GetError(err, http.StatusNotFound, w)
		return
	}

	if doc.DeletedCount == 0 {
		utils.GetError(errors.New("plugin failed to uninstall"), http.StatusBadRequest, w)
		return
	}

	utils.GetSuccess("plugin removed successfully", doc, w)
}
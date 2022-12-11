package organizations

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"zuri.chat/zccore/logger"
	"zuri.chat/zccore/utils"
)

var ErrNoAuthToken = errors.New("no authorization header provided")

// Install plugin into an organization.
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

	plugin, _ := utils.GetMongoDBDoc(PluginCollectionName, bson.M{"_id": pluginID})

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
	if err = utils.BsonToStruct(user, &member); err != nil {
		utils.GetError(err, http.StatusInternalServerError, w)
		return
	}

	if member.Role != OwnerRole && member.Role != AdminRole {
		utils.GetError(errors.New("access denied"), http.StatusForbidden, w)
		return
	}

	var p bson.M
	if strings.Contains(OrgID, "-org") {
		p, _ = utils.GetMongoDBDoc(OrganizationCollectionName, bson.M{"_id": OrgID},
			options.FindOne().SetProjection(bson.D{{Key: PluginCollectionName, Value: 1}, {Key: "_id", Value: 0}}))

	} else {
		pOrgID, err := primitive.ObjectIDFromHex(OrgID)
		if err != nil {
			utils.GetError(errors.New("invalid organization id"), http.StatusBadRequest, w)
			return
		}
		p, _ = utils.GetMongoDBDoc(OrganizationCollectionName, bson.M{"_id": pOrgID},
			options.FindOne().SetProjection(bson.D{{Key: PluginCollectionName, Value: 1}, {Key: "_id", Value: 0}}))

	}

	plugins := make(map[string]interface{})

	if err = utils.ConvertStructure(p[PluginCollectionName], &plugins); err != nil {
		utils.GetError(errors.New("operation failed"), http.StatusBadRequest, w)
		return
	}

	if _, ok := plugins[orgPlugin.PluginID]; ok {
		utils.GetError(errors.New("plugin has already been added"), http.StatusBadRequest, w)
		return
	}

	userName := member.UserName

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

	plugins[orgPlugin.PluginID] = pluginMap

	wg := sync.WaitGroup{}
	num := 1

	wg.Add(num)

	var save *mongo.UpdateResult

	go func() {
		defer wg.Done()

		addPlugin := bson.M{"plugins": plugins}

		save, err = utils.UpdateOneMongoDBDoc(OrganizationCollectionName, OrgID, addPlugin)
	}()

	wg.Wait()

	if err != nil || save.ModifiedCount != 1 {
		utils.GetError(err, http.StatusInternalServerError, w)
		return
	}

	var increaseCount *mongo.UpdateResult

	wg.Add(num)

	go func() {
		defer wg.Done()

		increaseCount, err = utils.IncrementOneMongoDBDocField(PluginCollectionName, orgPlugin.PluginID, "install_count")
	}()

	wg.Wait()

	if err != nil || increaseCount.ModifiedCount != 1 {
		utils.GetError(err, http.StatusInternalServerError, w)
		return
	}

	data := map[string]interface{}{
		"plugin_id": orgPlugin.PluginID,
	}

	utils.GetSuccess("plugin saved successfully", data, w)
}

// Get an organization plugins.
func (oh *OrganizationHandler) GetOrganizationPlugins(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var save bson.M
	orgID := mux.Vars(r)["id"]

	if strings.Contains(orgID, "-org") {
		save, _ = utils.GetMongoDBDoc(OrganizationCollectionName, bson.M{"_id": orgID})
	} else {
		objID, err := primitive.ObjectIDFromHex(orgID)
		if err != nil {
			utils.GetError(errors.New("invalid id"), http.StatusBadRequest, w)
			return
		}
		save, _ = utils.GetMongoDBDoc(OrganizationCollectionName, bson.M{"_id": objID})
	}

	if save == nil {
		utils.GetError(fmt.Errorf("organization %s not found", orgID), http.StatusNotFound, w)
		return
	}

	var org Organization

	// convert bson to struct
	bsonBytes, _ := bson.Marshal(save)
	err := bson.Unmarshal(bsonBytes, &org)

	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, w)
		return
	}

	utils.GetSuccess("plugins retrieved successfully", org.OrgPlugins(), w)
}

// Get an organization plugin.
func (oh *OrganizationHandler) GetOrganizationPlugin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var save bson.M
	orgID := mux.Vars(r)["id"]
	pluginID := mux.Vars(r)["plugin_id"]

	if strings.Contains(orgID, "-org") {
		save, _ = utils.GetMongoDBDoc(OrganizationCollectionName, bson.M{"_id": orgID})
	} else {
		objID, err := primitive.ObjectIDFromHex(orgID)

		if err != nil {
			utils.GetError(errors.New("invalid id"), http.StatusBadRequest, w)
			return
		}
		save, _ = utils.GetMongoDBDoc(OrganizationCollectionName, bson.M{"_id": objID})
	}

	if save == nil {
		utils.GetError(fmt.Errorf("organization %s not found", orgID), http.StatusNotFound, w)
		return
	}

	var org Organization

	// convert bson to struct
	bsonBytes, _ := bson.Marshal(save)
	err := bson.Unmarshal(bsonBytes, &org)

	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, w)
		return
	}

	doc := map[string]interface{}{}

	if _, ok := org.Plugins[pluginID]; !ok {
		logger.Error("plugin does not exist")
		utils.GetError(errors.New("plugin does not exist"), http.StatusNotFound, w)

		return
	}

	plugin := org.Plugins[pluginID]
	doc[pluginID] = plugin

	utils.GetSuccess("plugin returned successfully", doc, w)
}

// Remove an organization plugin.
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
	if err = utils.BsonToStruct(user, &member); err != nil {
		utils.GetError(err, http.StatusInternalServerError, w)
		return
	}

	if member.Role != OwnerRole && member.Role != AdminRole {
		utils.GetError(errors.New("access denied"), http.StatusForbidden, w)
		return
	}

	// confirm if user_id exists
	objID, err := primitive.ObjectIDFromHex(orgID)

	if err != nil {
		utils.GetError(errors.New("invalid user id"), http.StatusBadRequest, w)
		return
	}

	save, _ := utils.GetMongoDBDoc(OrganizationCollectionName, bson.M{"_id": objID})

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

	if len(org.Plugins) == 0 {
		utils.GetSuccess("organization has no plugin", nil, w)
	}

	if _, ok := org.Plugins[pluginID]; !ok {
		// plugin not found in organization.
		logger.Error("plugin does not exist")
		utils.GetError(errors.New("plugin does not exist"), http.StatusNotFound, w)

		return
	}

	delete(org.Plugins, pluginID)

	plugins := org.Plugins

	updatedPlugins := make(map[string]interface{})
	updatedPlugins["plugins"] = plugins

	update, err := utils.UpdateOneMongoDBDoc(OrganizationCollectionName, orgID, updatedPlugins)

	if err != nil || update.ModifiedCount != 1 {
		logger.Error("plugin failed to uninstall")
		utils.GetError(errors.New("plugin failed to uninstall"), http.StatusBadRequest, w)

		return
	}

	utils.GetSuccess("plugin removed successfully", nil, w)
}

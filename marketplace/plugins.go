package marketplace

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"zuri.chat/zccore/models"
	"zuri.chat/zccore/utils"
)

type M map[string]interface{}

func GetAllApprovedPlugins(w http.ResponseWriter, r *http.Request) {
	docs, err := utils.GetMongoDbDocs(models.PluginCollectionName, M{"approved": true})
	if err != nil {
		utils.GetError(err, http.StatusNotFound, w)
		return
	}

	// we will not expose all plugin info at marketplace, only name, description and iconurl
	mm := make([]M, len(docs))
	for i, doc := range docs {
		m := M{
			"id":          doc["_id"],
			"name":        doc["name"],
			"description": doc["description"],
			"icon_url":    doc["icon_url"],
		}
		mm[i] = m
	}
	utils.GetSuccess("success", mm, w)
}

func GetOneApprovedPlugin(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	objId, _ := primitive.ObjectIDFromHex(id)
	doc, err := utils.GetMongoDbDoc(models.PluginCollectionName, M{"_id": objId, "approved": true})
	if err != nil {
		utils.GetError(err, http.StatusNotFound, w)
		return
	}

	// we will not expose all plugin info at marketplace, only name, description and iconurl
	m := M{
		"id":          doc["_id"],
		"name":        doc["name"],
		"description": doc["description"],
		"icon_url":    doc["icon_url"],
		"install_url": doc["install_url"],
	}

	utils.GetSuccess("success", m, w)
}

/*
request = `{
	"plugin_id": "xxx",
	"organization_id": "xxx",
	"user_id": "xxx" //user installing plugin.
}`
*/
// It only installs plugins to an organization, organization has to load the plugins.
func InstallPluginToOrg(w http.ResponseWriter, r *http.Request) {
	requestData := make(map[string]string)
	if err := utils.ParseJsonFromRequest(r, &requestData); err != nil {
		utils.GetError(err, http.StatusUnprocessableEntity, w)
		return
	}
	orgId := requestData["organization_id"]
	pluginId := requestData["plugin_id"]
	userId := requestData["user_id"]
	//TODO check if these records exists

	// get plugin struct
	pluginObjID, _ := primitive.ObjectIDFromHex(pluginId)
	doc, err := utils.GetMongoDbDoc(models.PluginCollectionName, bson.M{"_id": pluginObjID})
	if err != nil {
		// plugin does not exist
		utils.GetError(err, http.StatusNotFound, w)
		return
	}
	// add plugin and org to installed_plugins coll
	_, err = utils.CreateMongoDbDoc(models.InstalledPluginsCollectionName, bson.M{
		"plugin_id":       pluginId,
		"added_by":        userId,
		"installed_at":    time.Now(),
		"organization_id": orgId,
		"plugin":          doc,
	})
	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, w)
		return
	}
	utils.GetSuccess("plugin successfully installed", doc, w)
}

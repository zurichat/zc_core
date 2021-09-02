package marketplace

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"zuri.chat/zccore/plugin"
	"zuri.chat/zccore/utils"
)

const (
	InstalledPluginsCollectionName = "installed_plugins"
)

type M map[string]interface{}

func GetAllPlugins(w http.ResponseWriter, r *http.Request) {
	ps, err := plugin.FindPlugins(bson.M{"approved": true})
	if err != nil {
		switch err {
		case mongo.ErrNoDocuments:
			w.WriteHeader(http.StatusNotFound)
			utils.GetSuccess("No plugins found", nil, w)
		default:
			utils.GetError(err, http.StatusNotFound, w)
		}
		return
	}
	utils.GetSuccess("success", ps, w)
}

func GetPlugin(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	p, err := plugin.FindPluginByID(id)
	if err != nil || !p.Approved {
		utils.GetError(err, http.StatusNotFound, w)
		return
	}

	utils.GetSuccess("success", p, w)
}

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

	p, err := plugin.FindPluginByID(pluginId)
	if err != nil {
		utils.GetError(err, http.StatusNotFound, w)
		return
	}
	// TODO: this has to be handled by organization guys
	// add plugin and org to installed_plugins coll
	_, err = utils.CreateMongoDbDoc(InstalledPluginsCollectionName, bson.M{
		"plugin_id":       pluginId,
		"added_by":        userId,
		"installed_at":    time.Now(),
		"organization_id": orgId,
		"plugin":          p,
	})
	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, w)
		return
	}
	utils.GetSuccess("plugin successfully installed", p, w)
}

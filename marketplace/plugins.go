package marketplace

import (
	"errors"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"zuri.chat/zccore/plugin"
	"zuri.chat/zccore/utils"
)

type M map[string]interface{}

func GetAllPlugins(w http.ResponseWriter, r *http.Request) {
	ps, err := plugin.FindPlugins(r.Context(), bson.M{"approved": true, "deleted":false})
	if err != nil {
		switch err {
		case mongo.ErrNoDocuments:
			utils.GetError(errors.New("no plugin available"), http.StatusNotFound, w)
		default:
			utils.GetError(err, http.StatusNotFound, w)
		}
		return
	}
	utils.GetSuccess("success", ps, w)
}

func GetPlugin(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	p, err := plugin.FindPluginByID(r.Context(), id)
	if err != nil {
		utils.GetError(err, http.StatusNotFound, w)
		return
	}
	if !p.Approved {
		utils.GetError(errors.New("plugin is not approved"), http.StatusForbidden, w)
		return
	}

	if p.Deleted {
		utils.GetError(errors.New("plugin no longer exists"), http.StatusForbidden, w)
		return
	}

	utils.GetSuccess("success", p, w)
}


// an endpoint to remove plugins from marketplace
func RemovePlugin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	
	pluginID := mux.Vars(r)["id"]

	pluginExists, err := plugin.FindPluginByID(r.Context(), pluginID)
	if err != nil {
		utils.GetError(err, http.StatusNotFound, w)
		return
	}

	if pluginExists == nil {
		utils.GetError(errors.New("plugin does not exist"), http.StatusBadRequest, w)
		return
	}

	update := M{"deleted": true, "deleted_at": time.Now().String()}

	updateRes, err := utils.UpdateOneMongoDbDoc(plugin.PluginCollectionName, pluginID, update)
	if err != nil {
		utils.GetError(errors.New("plugin removal failed"), http.StatusBadRequest, w)
		return
	}

	utils.GetSuccess("plugin successfully removed", updateRes, w)
}
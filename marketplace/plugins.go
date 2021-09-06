package marketplace

import (
	"errors"
	"net/http"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"zuri.chat/zccore/plugin"
	"zuri.chat/zccore/utils"
)

type M map[string]interface{}

func GetAllPlugins(w http.ResponseWriter, r *http.Request) {
	ps, err := plugin.FindPlugins(r.Context(), bson.M{"approved": true})
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

	utils.GetSuccess("success", p, w)
}

package marketplace

import (
	"net/http"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"zuri.chat/zccore/plugin"
	"zuri.chat/zccore/utils"
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

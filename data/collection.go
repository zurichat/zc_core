package data

import (
	"errors"
	"net/http"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"zuri.chat/zccore/utils"
)

func ListCollections(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pluginId, orgId := vars["plugin_id"], vars["org_id"]
	docs, err := make([]bson.M, 0), error(nil)
	if orgId != "" {
		docs, err = getPluginCollectionsForOrganization(pluginId, orgId)
	} else {
		docs, err = getPluginCollections(pluginId)
	}
	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, w)
		return
	}
	if len(docs) == 0 {
		utils.GetError(errors.New("no record found"), http.StatusNotFound, w)
		return
	}
	utils.GetSuccess("collections retrieved successfully", docs, w)
}

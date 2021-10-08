package data

import (
	"errors"
	"net/http"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"zuri.chat/zccore/utils"
)

// ListCollections returns a list of collections a plugin has created
func ListCollections(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pluginID, orgID := vars["plugin_id"], vars["org_id"]
	docs, err := make([]bson.M, 0), error(nil)
	if orgID != "" {
		docs, err = getPluginCollectionsForOrganization(pluginID, orgID)
	} else {
		docs, err = getPluginCollections(pluginID)
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

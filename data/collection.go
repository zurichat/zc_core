package data

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"zuri.chat/zccore/utils"
)

// ListCollections returns a list of collections a plugin has created.
func ListCollections(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pluginID, orgID := vars["plugin_id"], vars["org_id"]

	var err error

	var docs []bson.M

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

// CollectionDetail returns details about a collection
func CollectionDetail(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	pluginID, orgID, collName := vars["plugin_id"], vars["org_id"], vars["coll_name"]

	prefixedCollName := getPrefixedCollectionName(pluginID, orgID, collName)
	coll := utils.GetCollection(prefixedCollName)
	count, err := coll.CountDocuments(r.Context(), bson.M{})

	if err != nil {
		utils.GetError(fmt.Errorf("unable to get collection details: %v", err), http.StatusInternalServerError, w)
		return
	}

	utils.GetSuccess("success", utils.M{
		"count": count,
	}, w)
}

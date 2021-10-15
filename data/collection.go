package data

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"zuri.chat/zccore/utils"
)

// CollectionDetail returns details about a collection
func CollectionDetail(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	pluginID, orgID, collName := vars["plugin_id"], vars["org_id"], vars["coll_name"]

	if orgID == "__none__" {
		orgID = ""
	}

	actualCollName := mongoCollectionName(pluginID, collName)

	coll := utils.GetCollection(actualCollName)
	count, err := coll.CountDocuments(r.Context(), bson.M{"organization_id": orgID})

	if err != nil {
		utils.GetError(fmt.Errorf("unable to get collection details: %v", err), http.StatusInternalServerError, w)
		return
	}

	utils.GetSuccess("success", utils.M{
		"count": count,
	}, w)
}

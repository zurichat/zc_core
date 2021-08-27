package data

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"zuri.chat/zccore/utils"
)

type M map[string]interface{}

func ReadData(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	collName, pluginID := vars["coll_name"], vars["plugin_id"]

	filter := M{"collection_name": collName}
	pluginData, err := utils.GetMongoDbDoc(_COLLECTION_NAME, filter)

	if err != nil {
		// error retrieving plugin data
		utils.GetError(fmt.Errorf("error retreiving data: %v", err), http.StatusInternalServerError, w)
		return
	}

	ownerID, ok := pluginData["owner_plugin_id"].(primitive.ObjectID)
	if !ok {
		// invalid plugin id
		utils.GetError(fmt.Errorf("invalid plugin id"), http.StatusUnprocessableEntity, w)
		return
	}

	if ownerID.Hex() != pluginID {
		// not allowed to access collection that is not yours.
		utils.GetError(fmt.Errorf("not allowed to access this data"), http.StatusForbidden, w)
		return
	}
	// TODO: check for query params to filter results

	data, err := utils.GetMongoDbDocs(collName, M{})

	if err != nil {
		// error retrieving results
		utils.GetError(err, http.StatusInternalServerError, w)
		return
	}

	utils.GetSuccess("success", data, w)
}
